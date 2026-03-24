package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/jettjia/igo-pkg/aipkg/rag/contextcompact"
)

type mockModel struct{}

func (m *mockModel) Generate(ctx context.Context, msgs []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: "[Mock summary: This is a simulated summary of the conversation for testing purposes.]",
	}, nil
}

func (m *mockModel) Stream(ctx context.Context, msgs []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, _ := m.Generate(ctx, msgs, opts...)
	sr := schema.StreamReaderFromArray([]*schema.Message{msg})
	return sr, nil
}

func main() {
	ctx := context.Background()

	mw, err := contextcompact.NewAgentMiddleware(ctx, &contextcompact.MiddlewareConfig{
		TokenThreshold:     5000,
		KeepRecentMessages: 2,
		TranscriptDir:      ".transcripts",
		Model:              &mockModel{},
		SystemPrompt:       "You are a conversation summarizer. Summarize concisely.",
	})
	if err != nil {
		fmt.Printf("Failed to create middleware: %v\n", err)
		return
	}

	fmt.Println("=== Test 1: Small conversation (no compression) ===")
	testSmallConversation(ctx, mw)

	fmt.Println("\n=== Test 2: Large conversation (auto_compact) ===")
	testAutoCompact(ctx, mw)

	fmt.Println("\n=== Test 3: Stream messages (full flow) ===")
	testStreamMessages(ctx, mw)

	fmt.Println("\n=== Test 4: Very large conversation ===")
	testVeryLargeConversation(ctx, mw)
}

func testSmallConversation(ctx context.Context, mw adk.AgentMiddleware) {
	messages := []*schema.Message{
		{Role: schema.System, Content: "You are a helpful assistant."},
		{Role: schema.User, Content: "Hello, how are you?"},
		{Role: schema.Assistant, Content: "I'm doing well, thank you!"},
	}

	state := &adk.ChatModelAgentState{Messages: messages}
	err := mw.BeforeChatModel(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("Before: %d messages\n", len(messages))
	fmt.Printf("After: %d messages\n", len(state.Messages))
}

func testAutoCompact(ctx context.Context, mw adk.AgentMiddleware) {
	messages := []*schema.Message{
		{Role: schema.System, Content: "You are a helpful coding assistant."},
	}

	longCode := generateLongCode()
	for i := 0; i < 20; i++ {
		messages = append(messages,
			&schema.Message{Role: schema.User, Content: fmt.Sprintf("Question %d: %s", i+1, longCode)},
			&schema.Message{Role: schema.Assistant, Content: fmt.Sprintf("Answer %d: This is a detailed response about the code. %s", i+1, longCode)},
		)
	}

	state := &adk.ChatModelAgentState{Messages: messages}
	fmt.Printf("Before: %d messages, estimated tokens: %d\n", len(messages), estimateTokens(messages))

	err := mw.BeforeChatModel(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	fmt.Printf("After: %d messages\n", len(state.Messages))
	for i, msg := range state.Messages {
		if msg.Role == schema.Assistant && msg.Name == "summary" {
			fmt.Printf("  Summary[%d]: %s\n", i, truncate(msg.Content, 100))
		}
	}
}

func testStreamMessages(ctx context.Context, mw adk.AgentMiddleware) {
	messages := []*schema.Message{
		{Role: schema.System, Content: "You are a helpful assistant."},
	}

	longContent := strings.Repeat("This is a long message. ", 100)
	for i := 0; i < 15; i++ {
		messages = append(messages,
			&schema.Message{Role: schema.User, Content: fmt.Sprintf("User message %d: %s", i+1, longContent)},
			&schema.Message{Role: schema.Assistant, Content: fmt.Sprintf("Assistant response %d: %s", i+1, longContent)},
		)
	}

	state := &adk.ChatModelAgentState{Messages: messages}
	initialTokens := estimateTokens(messages)
	fmt.Printf("Initial: %d messages, ~%d tokens\n", len(messages), initialTokens)

	for iter := 0; iter < 3; iter++ {
		err := mw.BeforeChatModel(ctx, state)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		fmt.Printf("  Iteration %d: %d messages, ~%d tokens\n", iter+1, len(state.Messages), estimateTokens(state.Messages))
	}
}

func testVeryLargeConversation(ctx context.Context, mw adk.AgentMiddleware) {
	messages := []*schema.Message{
		{Role: schema.System, Content: "You are a helpful AI assistant."},
	}

	longContent := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 50)
	for i := 0; i < 50; i++ {
		messages = append(messages,
			&schema.Message{Role: schema.User, Content: fmt.Sprintf("User question %d: %s", i+1, longContent)},
			&schema.Message{Role: schema.Assistant, Content: fmt.Sprintf("Assistant answer %d: %s", i+1, longContent)},
		)
	}

	state := &adk.ChatModelAgentState{Messages: messages}
	initialTokens := estimateTokens(messages)
	fmt.Printf("Initial: %d messages, ~%d tokens (threshold: 5000)\n", len(messages), initialTokens)

	err := mw.BeforeChatModel(ctx, state)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	finalTokens := estimateTokens(state.Messages)
	fmt.Printf("After compression: %d messages, ~%d tokens\n", len(state.Messages), finalTokens)
	fmt.Printf("Compression ratio: %.1fx\n", float64(initialTokens)/float64(finalTokens))
}

func estimateTokens(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		if msg == nil {
			continue
		}
		total += len(msg.Content) / 4
		if len(msg.ToolCalls) > 0 {
			total += len(msg.ToolCalls) * 10
		}
		total += 4
	}
	return total
}

func generateLongCode() string {
	var sb strings.Builder
	sb.WriteString("func processData(input string) {\n")
	sb.WriteString("    // This is a long function with lots of comments\n")
	for i := 0; i < 10; i++ {
		sb.WriteString(fmt.Sprintf("    data := process%d(input)\n", i))
		sb.WriteString(fmt.Sprintf("    result := transform%d(data)\n", i))
		sb.WriteString(fmt.Sprintf("    output := validate%d(result)\n", i))
	}
	sb.WriteString("}\n")
	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func init() {
	log.SetFlags(0)
	os.MkdirAll(".transcripts", 0755)
}
