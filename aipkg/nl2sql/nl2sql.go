package nl2sql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jettjia/igo-pkg/aipkg/models/chat"
	"github.com/jettjia/igo-pkg/aipkg/pkg/logger"
	"github.com/jettjia/igo-pkg/aipkg/pkg/types"
	"github.com/jettjia/igo-pkg/aipkg/schema"
	config "github.com/jettjia/igo-pkg/pkg/conf"
)

type NL2SQL interface {
	NL2SQL(ctx context.Context, messages []*types.Message, tables []*schema.TableSchema) (sql string, err error)
}

// nl2sqlImpl 实现 NL2SQL 接口
type nl2sqlImpl struct {
	chatClient chat.Chat
	config     *config.Config
}

// NewNL2SQL 创建 NL2SQL 实例
func NewNL2SQL(chatClient chat.Chat, config *config.Config) (NL2SQL, error) {
	return &nl2sqlImpl{
		chatClient: chatClient,
		config:     config,
	}, nil
}

// NL2SQL 根据消息和表结构生成 SQ
func (n *nl2sqlImpl) NL2SQL(ctx context.Context, messages []*types.Message, tables []*schema.TableSchema) (sql string, err error) {
	// 1. 构建表结构描述
	tableSchemaDesc, err := buildTableSchemaDesc(tables)
	if err != nil {
		return "", fmt.Errorf("构建表结构描述失败: %w", err)
	}

	// 2. 构建历史对话上下文
	historyContext := buildHistoryContext(messages)

	// 3. 构建用户提示词
	userPrompt := fmt.Sprintf(n.config.Ai.NL2SQLPromptUser, tableSchemaDesc, historyContext)

	// 4. 准备提示词消息
	promptMessages := []chat.Message{
		{
			Role:    "system",
			Content: n.config.Ai.NL2SQLPromptSystem,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	// 5. 调用聊天模型生成 SQL
	response, err := n.chatClient.Chat(ctx, promptMessages, &chat.ChatOptions{Temperature: 0.0})
	if err != nil {
		logger.Errorf(ctx, "调用聊天模型生成 SQL 失败: %v", err)
		return "", fmt.Errorf("调用聊天模型生成 SQL 失败: %w", err)
	}

	// 6. 解析模型响应
	var result struct {
		SQL     string `json:"sql"`
		ErrCode int    `json:"err_code"`
		ErrMsg  string `json:"err_msg"`
	}

	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		// 如果解析失败，尝试直接使用响应内容作为 SQL
		logger.Warnf(ctx, "解析模型响应失败，尝试直接使用响应内容: %v, 响应: %s", err, response.Content)
		return response.Content, nil
	}

	// 7. 处理错误码
	if result.ErrCode != 0 {
		return "", fmt.Errorf("生成 SQL 失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	}

	return result.SQL, nil
}

// buildTableSchemaDesc 构建表结构描述字符串
func buildTableSchemaDesc(tables []*schema.TableSchema) (string, error) {
	if len(tables) == 0 {
		return "", fmt.Errorf("表结构为空")
	}

	var builder strings.Builder
	for _, table := range tables {
		builder.WriteString(fmt.Sprintf("表名: %s\n", table.Name))
		builder.WriteString("字段:\n")
		for _, column := range table.Columns {
			builder.WriteString(fmt.Sprintf("- %s: %s\n", column.Name, column.Type))
		}
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// buildHistoryContext 构建历史对话上下文
func buildHistoryContext(messages []*types.Message) string {
	if len(messages) == 0 {
		return "无"
	}

	var builder strings.Builder
	for i, msg := range messages {
		builder.WriteString(fmt.Sprintf("对话%d:\n", i+1))
		builder.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
		builder.WriteString("\n")
	}

	return builder.String()
}
