package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jettjia/igo-pkg/aipkg/models/chat"
)

func getLLM() chat.Chat {
	chatModel, err := chat.NewChat(&chat.ChatConfig{
		ModelID:   "1",
		APIKey:    os.Getenv("OPENAI_API_KEY"),
		BaseURL:   os.Getenv("OPENAI_BASE_URL"),
		ModelName: "deepseek-ai/DeepSeek-V3",
		Source:    "remote",
	})
	if err != nil {
		panic(err)
	}

	return chatModel
}

func TestChat() {
	llm := getLLM()
	resp, err := llm.Chat(context.Background(), []chat.Message{
		{
			Role:    "user",
			Content: "你好",
		},
	}, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func main() {
	TestChat()
}
