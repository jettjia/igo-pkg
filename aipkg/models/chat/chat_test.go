package chat_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jettjia/go-pkg/aipkg/models/chat"
	"github.com/jettjia/go-pkg/aipkg/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestChat_Remote_Chat(t *testing.T) {
	// Mock OpenAI API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-3.5-turbo",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you today?"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 20,
				"total_tokens": 30
			}
		}`))
	}))
	defer ts.Close()

	// Configure Chat
	config := &chat.ChatConfig{
		Source:    types.ModelSourceRemote,
		BaseURL:   ts.URL + "/v1", // OpenAI client adds /chat/completions, usually expects /v1 base
		ModelName: "gpt-3.5-turbo",
		APIKey:    "test-key",
	}

	// Create Chat instance
	c, err := chat.NewChat(config)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	// Test Chat
	ctx := context.Background()
	messages := []chat.Message{
		{Role: "user", Content: "Hello"},
	}
	opts := &chat.ChatOptions{
		Temperature: 0.7,
	}

	resp, err := c.Chat(ctx, messages, opts)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	expectedContent := "Hello! How can I help you today?"
	assert.Equal(t, expectedContent, resp.Content)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

func TestChat_Remote_ChatStream(t *testing.T) {
	// Mock OpenAI API server for streaming
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		assert.True(t, ok, "Expected http.ResponseWriter to be an http.Flusher")

		// Send chunks
		chunks := []string{
			`{"id":"1","choices":[{"delta":{"role":"assistant","content":""}}]}`,
			`{"id":"1","choices":[{"delta":{"content":"Hello"}}]}`,
			`{"id":"1","choices":[{"delta":{"content":" World"}}]}`,
			`[DONE]`,
		}

		for _, chunk := range chunks {
			if chunk == "[DONE]" {
				fmt.Fprintf(w, "data: %s\n\n", chunk)
			} else {
				fmt.Fprintf(w, "data: %s\n\n", chunk)
			}
			flusher.Flush()
		}
	}))
	defer ts.Close()

	config := &chat.ChatConfig{
		Source:    types.ModelSourceRemote,
		BaseURL:   ts.URL + "/v1",
		ModelName: "gpt-3.5-turbo",
		APIKey:    "test-key",
	}

	c, err := chat.NewChat(config)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	ctx := context.Background()
	messages := []chat.Message{
		{Role: "user", Content: "Hello"},
	}

	streamChan, err := c.ChatStream(ctx, messages, nil)
	assert.NoError(t, err)
	assert.NotNil(t, streamChan)

	var fullContent string
	for resp := range streamChan {
		if !resp.Done {
			fullContent += resp.Content
		}
	}

	expectedContent := "Hello World"
	assert.Equal(t, expectedContent, fullContent)
}
