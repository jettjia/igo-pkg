package contextcompact

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MiddlewareConfig 配置上下文压缩中间件
type MiddlewareConfig struct {
	TokenThreshold     int    // token 超过此值时触发自动压缩
	KeepRecentMessages int    // 保留最近的 tool result 数量
	TranscriptDir      string // 保存 transcript 的目录
	Model              model.BaseChatModel
	SystemPrompt       string // 总结对话用的 system prompt
}

func NewAgentMiddleware(ctx context.Context, cfg *MiddlewareConfig) (adk.AgentMiddleware, error) {
	if cfg == nil {
		return adk.AgentMiddleware{}, fmt.Errorf("config is nil")
	}
	if cfg.Model == nil {
		return adk.AgentMiddleware{}, fmt.Errorf("model is nil")
	}

	threshold := 50000 // 与 Python 实现保持一致
	if cfg.TokenThreshold > 0 {
		threshold = cfg.TokenThreshold
	}

	keepRecent := 3 // 与 Python 实现 KEEP_RECENT = 3 保持一致
	if cfg.KeepRecentMessages > 0 {
		keepRecent = cfg.KeepRecentMessages
	}

	transcriptDir := ".transcripts"
	if cfg.TranscriptDir != "" {
		transcriptDir = cfg.TranscriptDir
	}

	systemPrompt := cfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a conversation summarizer. Summarize the older messages concisely while preserving key information, decisions, and current state."
	}

	mw := &middleware{
		tokenThreshold:     threshold,
		keepRecentMessages: keepRecent,
		transcriptDir:      transcriptDir,
		model:              cfg.Model,
		systemPrompt:       systemPrompt,
	}

	return adk.AgentMiddleware{BeforeChatModel: mw.BeforeModel}, nil
}

type middleware struct {
	tokenThreshold     int
	keepRecentMessages int
	transcriptDir      string
	model              model.BaseChatModel
	systemPrompt       string
}

// BeforeModel 实现三层压缩策略
// Layer 1: micro_compact - 每次调用前替换旧的 tool result 为占位符
// Layer 2: auto_compact - token 超过阈值时，保存 transcript 并总结
func (m *middleware) BeforeModel(ctx context.Context, state *adk.ChatModelAgentState) error {
	if state == nil || len(state.Messages) == 0 {
		return nil
	}

	messages := state.Messages

	// Layer 1: micro_compact - 替换旧的 tool result 为占位符
	m.microCompact(messages)

	tokens := m.countTokens(messages)
	log.Printf("[compact] before model check: msg_count=%d tokens=%d threshold=%d", len(messages), tokens, m.tokenThreshold)

	// 如果 token 未超过阈值，直接返回
	if tokens <= m.tokenThreshold {
		return nil
	}

	// Layer 2: auto_compact - token 超过阈值，保存 transcript 并总结
	log.Printf("[compact] triggering compression: msg_count=%d tokens=%d", len(messages), tokens)

	// 保存完整的 transcript 到磁盘
	transcriptPath, err := m.saveTranscript(messages)
	if err != nil {
		log.Printf("[compact] save transcript error: %v", err)
		// 继续尝试压缩，不因为保存失败而中断
	}

	summary, err := m.summarize(ctx, messages)
	if err != nil {
		return err
	}

	newMessages := m.buildCompactedState(messages, summary, transcriptPath)
	state.Messages = newMessages

	log.Printf("[compact] compression done: new_msg_count=%d", len(newMessages))

	return nil
}

// microCompact Layer 1: 替换旧的 tool result 为占位符
// 与 Python 实现一致：将超过 keepRecent 数量的 tool result 替换为 "[Previous: used {tool_name}]"
func (m *middleware) microCompact(messages []*schema.Message) {
	if len(messages) == 0 {
		return
	}

	// 收集所有 tool result 和它们对应的 tool_call 信息
	type toolResultInfo struct {
		msgIndex   int
		msg        *schema.Message
		toolName   string
		toolUseID  string
	}

	// 建立 tool_use_id -> tool_name 的映射
	toolNameMap := make(map[string]string)
	for _, msg := range messages {
		if msg == nil || msg.Role != schema.Assistant {
			continue
		}
		for _, tc := range msg.ToolCalls {
			if tc.ID != "" && tc.Function.Name != "" {
				toolNameMap[tc.ID] = tc.Function.Name
			}
		}
	}

	// 检查是否有超过 keepRecent 个 tool 消息
	toolMsgCount := 0
	for _, msg := range messages {
		if msg != nil && msg.Role == schema.Tool {
			toolMsgCount++
		}
	}

	// 如果 tool 消息数量不超过 keepRecent，不需要压缩
	if toolMsgCount <= m.keepRecentMessages {
		return
	}

	// 找到需要压缩的 tool 消息（保留最近的 keepRecent 个）
	var toCompact []*schema.Message
	var toKeep []*schema.Message

	count := 0
	for _, msg := range messages {
		if msg != nil && msg.Role == schema.Tool {
			if count < toolMsgCount-m.keepRecentMessages {
				toCompact = append(toCompact, msg)
				count++
			} else {
				toKeep = append(toKeep, msg)
			}
		} else {
			toKeep = append(toKeep, msg)
		}
	}

	// 替换旧的 tool result 内容
	for _, msg := range toCompact {
		if msg != nil && len(msg.Content) > 100 {
			// 尝试从 tool name map 获取 tool 名称
			toolName := "unknown"
			if msg.Name != "" {
				toolName = msg.Name
			}
			msg.Content = fmt.Sprintf("[Previous: used %s]", toolName)
		}
	}

	log.Printf("[compact] micro_compact: compacted %d tool messages, kept %d", len(toCompact), len(toKeep))
}

func (m *middleware) countTokens(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		if msg == nil {
			continue
		}
		contentTokens := len(msg.Content) / 4
		roleTokens := 4
		toolTokens := 0
		if len(msg.ToolCalls) > 0 {
			toolTokens = len(msg.ToolCalls) * 10
		}
		total += contentTokens + roleTokens + toolTokens
	}
	return total
}

func (m *middleware) summarize(ctx context.Context, messages []*schema.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("Summarize this conversation for continuity. Include:\n")
	sb.WriteString("1) What was accomplished\n")
	sb.WriteString("2) Current state\n")
	sb.WriteString("3) Key decisions made\n")
	sb.WriteString("Be concise but preserve critical details.\n\n")

	for _, msg := range messages {
		if msg == nil || msg.Role == schema.System {
			continue
		}
		fmt.Fprintf(&sb, "[%s]\n", msg.Role)
		if msg.Content != "" {
			sb.WriteString(msg.Content)
			sb.WriteString("\n")
		}
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				if tc.Function.Name != "" {
					fmt.Fprintf(&sb, "tool_call: %s\n", tc.Function.Name)
				}
			}
		}
	}

	msgs := []*schema.Message{
		{Role: schema.System, Content: m.systemPrompt},
		{Role: schema.User, Content: sb.String()},
	}

	resp, err := m.model.Generate(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("summarize failed: %w", err)
	}

	summary := strings.TrimSpace(resp.Content)
	if len(summary) > 2000 {
		summary = summary[:2000] + "..."
	}

	return summary, nil
}

func (m *middleware) buildCompactedState(messages []*schema.Message, summary string, transcriptPath string) []*schema.Message {
	if len(messages) == 0 {
		return messages
	}

	newMessages := make([]*schema.Message, 0, len(messages))

	// 保留 system 消息
	for _, msg := range messages {
		if msg != nil && msg.Role == schema.System {
			newMessages = append(newMessages, msg)
			break
		}
	}

	// 添加总结消息，包含 transcript 路径信息
	summaryContent := summary
	if transcriptPath != "" {
		summaryContent = fmt.Sprintf("[Conversation compressed. Transcript: %s]\n\n%s", transcriptPath, summary)
	}
	newMessages = append(newMessages, &schema.Message{
		Role:    schema.Assistant,
		Content: summaryContent,
		Name:    "summary",
	})

	// 保留最近的 tool calls 和 tool result (与 Python 实现一致)
	recentCount := 0
	for i := len(messages) - 1; recentCount < m.keepRecentMessages*2 && i >= 0; i-- {
		msg := messages[i]
		if msg == nil || msg.Role == schema.System {
			continue
		}
		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			recentCount++
			newMessages = append(newMessages, msg)
		}
		if msg.Role == schema.Tool {
			recentCount++
			newMessages = append(newMessages, msg)
		}
	}

	return newMessages
}

// saveTranscript 保存完整对话到磁盘
func (m *middleware) saveTranscript(messages []*schema.Message) (string, error) {
	if m.transcriptDir == "" {
		return "", nil
	}

	// 创建目录
	if err := os.MkdirAll(m.transcriptDir, 0755); err != nil {
		return "", fmt.Errorf("create transcript dir failed: %w", err)
	}

	// 生成文件名
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("transcript_%d.jsonl", timestamp)
	transcriptPath := filepath.Join(m.transcriptDir, filename)

	// 写入 JSONL 文件
	file, err := os.Create(transcriptPath)
	if err != nil {
		return "", fmt.Errorf("create transcript file failed: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, msg := range messages {
		if msg == nil {
			continue
		}
		if err := encoder.Encode(msg); err != nil {
			return "", fmt.Errorf("encode message failed: %w", err)
		}
	}

	log.Printf("[compact] transcript saved: %s", transcriptPath)
	return transcriptPath, nil
}
