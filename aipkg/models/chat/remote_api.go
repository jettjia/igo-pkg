package chat

import (
	"context"
	"fmt"

	"github.com/jettjia/igo-pkg/aipkg/pkg/types"
	"github.com/sashabaranov/go-openai"
)

// RemoteAPIChat 实现了基于的聊天
type RemoteAPIChat struct {
	modelName string
	client    *openai.Client
	modelID   string
}

// NewRemoteAPIChat 调用远程API 聊天实例
func NewRemoteAPIChat(chatConfig *ChatConfig) (*RemoteAPIChat, error) {
	apiKey := chatConfig.APIKey
	config := openai.DefaultConfig(apiKey)
	if baseURL := chatConfig.BaseURL; baseURL != "" {
		config.BaseURL = baseURL
	}
	return &RemoteAPIChat{
		modelName: chatConfig.ModelName,
		client:    openai.NewClientWithConfig(config),
		modelID:   chatConfig.ModelID,
	}, nil
}

// convertMessages 转换消息格式为OpenAI格式
func (c *RemoteAPIChat) convertMessages(messages []Message) []openai.ChatCompletionMessage {
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return openaiMessages
}

// buildChatCompletionRequest 构建聊天请求参数
func (c *RemoteAPIChat) buildChatCompletionRequest(messages []Message,
	opts *ChatOptions, isStream bool,
) openai.ChatCompletionRequest {
	req := openai.ChatCompletionRequest{
		Model:    c.modelName,
		Messages: c.convertMessages(messages),
		Stream:   isStream,
	}

	// 添加可选参数
	if opts != nil {
		if opts.Temperature > 0 {
			req.Temperature = float32(opts.Temperature)
		}
		if opts.TopP > 0 {
			req.TopP = float32(opts.TopP)
		}
		if opts.MaxTokens > 0 {
			req.MaxTokens = opts.MaxTokens
		}
		if opts.MaxCompletionTokens > 0 {
			req.MaxCompletionTokens = opts.MaxCompletionTokens
		}
		if opts.FrequencyPenalty > 0 {
			req.FrequencyPenalty = float32(opts.FrequencyPenalty)
		}
		if opts.PresencePenalty > 0 {
			req.PresencePenalty = float32(opts.PresencePenalty)
		}
		if opts.Thinking != nil {
			req.ChatTemplateKwargs = map[string]any{
				"enable_thinking": *opts.Thinking,
			}
		}
	}

	return req
}

// Chat 进行非流式聊天
func (c *RemoteAPIChat) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*types.ChatResponse, error) {
	// 构建请求参数
	req := c.buildChatCompletionRequest(messages, opts, false)

	// 发送请求
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// 转换响应格式
	return &types.ChatResponse{
		Content: resp.Choices[0].Message.Content,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// ChatStream 进行流式聊天
func (c *RemoteAPIChat) ChatStream(ctx context.Context,
	messages []Message, opts *ChatOptions,
) (<-chan types.StreamResponse, error) {
	// 构建请求参数
	req := c.buildChatCompletionRequest(messages, opts, true)

	// 创建流式响应通道
	streamChan := make(chan types.StreamResponse)

	// 启动流式请求
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		close(streamChan)
		return nil, fmt.Errorf("create chat completion stream: %w", err)
	}

	// 在后台处理流式响应
	go func() {
		defer close(streamChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				streamChan <- types.StreamResponse{
					ResponseType: types.ResponseTypeAnswer,
					Done:         true,
				}
				return
			}
			if len(response.Choices) > 0 {
				streamChan <- types.StreamResponse{
					ResponseType: types.ResponseTypeAnswer,
					Content:      response.Choices[0].Delta.Content,
					Done:         false,
				}
			}
		}
	}()

	return streamChan, nil
}

// GetModelName 获取模型名称
func (c *RemoteAPIChat) GetModelName() string {
	return c.modelName
}

// GetModelID 获取模型ID
func (c *RemoteAPIChat) GetModelID() string {
	return c.modelID
}
