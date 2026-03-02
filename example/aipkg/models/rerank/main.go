package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jettjia/go-pkg/aipkg/models/rerank"
	"github.com/jettjia/go-pkg/aipkg/pkg/types"
)

func main() {
	config := &rerank.RerankerConfig{
		APIKey:    os.Getenv("OPENAI_API_KEY"),
		BaseURL:   os.Getenv("OPENAI_BASE_URL"),
		ModelName: "Qwen/Qwen3-Reranker-0.6B",
		Source:    types.ModelSourceRemote,
		ModelID:   "Qwen/Qwen3-Reranker-0.6B",
	}
	reranker, err := rerank.NewReranker(config)
	if err != nil {
		panic("NewReranker: %v" + err.Error())
	}

	query := "你好"
	documents := []string{"你好", "你好吗", "你好啊"}
	results, err := reranker.Rerank(context.Background(), query, documents)
	if err != nil {
		panic("Rerank: %v" + err.Error())
	}
	for _, result := range results {
		fmt.Printf("Rank: %d, Document: %s, Score: %f\n", result.Index, result.Document, result.RelevanceScore)
	}
}
