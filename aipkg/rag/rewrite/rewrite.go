package rewrite

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/jettjia/go-pkg/aipkg/models/chat"
	config "github.com/jettjia/go-pkg/pkg/conf"
)

// Rewriter 负责关键词提取和问题重写
type Rewriter struct {
	chatClient chat.Chat
	config     *config.Config
}

// NewRewriter 创建一个新的重写器实例
func NewRewriter(chatClient chat.Chat, config *config.Config) (*Rewriter, error) {
	return &Rewriter{
		chatClient: chatClient,
		config:     config,
	}, nil
}

// ExtractKeywords 从用户问题中提取关键词
func (r *Rewriter) ExtractKeywords(ctx context.Context, query string) (string, error) {
	// 准备提示词
	messages := []chat.Message{
		{
			Role:    "system",
			Content: r.config.Ai.KeywordsExtractionPrompt,
		},
		{
			Role:    "user",
			Content: query,
		},
	}

	// 调用聊天模型提取关键词
	response, err := r.chatClient.Chat(ctx, messages, &chat.ChatOptions{})
	if err != nil {
		return "", fmt.Errorf("调用聊天模型提取关键词失败: %w", err)
	}

	return response.Content, nil
}

// MessageItem 表示历史对话中的一条消息对
type MessageItem struct {
	Query  string
	Answer string
}

// TemplateData 用于模板渲染的数据结构
type TemplateData struct {
	Conversation []MessageItem
	Query        string
}

// RewriteQuery 改写用户问题，进行指代消解和省略补全
func (r *Rewriter) RewriteQuery(ctx context.Context, history []chat.Message, query string) (string, error) {
	// 将历史消息转换为模板所需的结构
	messageItems := []MessageItem{}
	var currentUser, currentAssistant string

	// 处理历史消息，将user和assistant消息配对
	for _, msg := range history {
		switch msg.Role {
		case "user":
			// 如果有未配对的assistant消息，先添加到列表
			if currentAssistant != "" {
				messageItems = append(messageItems, MessageItem{
					Query:  currentUser,
					Answer: currentAssistant,
				})
				currentAssistant = ""
			}
			currentUser = msg.Content
		case "assistant":
			currentAssistant = msg.Content
		}
	}

	// 添加最后一对（如果有）
	if currentUser != "" {
		messageItems = append(messageItems, MessageItem{
			Query:  currentUser,
			Answer: currentAssistant,
		})
	}

	// 准备模板数据
	data := TemplateData{
		Conversation: messageItems,
		Query:        query,
	}

	// 使用模板引擎渲染用户消息
	tmpl, err := template.New("prompt").Parse(r.config.Ai.RewritePromptUser)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	var userMessageBuffer strings.Builder
	if err = tmpl.Execute(&userMessageBuffer, data); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}

	userMessage := userMessageBuffer.String()

	// 准备提示词
	messages := []chat.Message{
		{
			Role:    "system",
			Content: r.config.Ai.RewritePromptSystem,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	// 调用聊天模型改写问题
	response, err := r.chatClient.Chat(ctx, messages, &chat.ChatOptions{})
	if err != nil {
		return "", fmt.Errorf("调用聊天模型改写问题失败: %w", err)
	}

	return response.Content, nil
}
