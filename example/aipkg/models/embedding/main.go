package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jettjia/go-pkg/aipkg/models/embedding"
	"github.com/jettjia/go-pkg/aipkg/pkg/ollama"
	"github.com/jettjia/go-pkg/aipkg/pkg/types"
	"github.com/panjf2000/ants/v2"
)

func main() {
	config := embedding.Config{
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		BaseURL:    os.Getenv("OPENAI_BASE_URL"),
		ModelName:  "BAAI/bge-large-zh-v1.5",
		Source:     types.ModelSourceRemote,
		ModelID:    "BAAI/bge-large-zh-v1.5",
		Dimensions: 1024,
	}

	// 正确创建依赖项
	pool, err := ants.NewPool(10) // 创建一个大小为10的goroutine池
	if err != nil {
		panic("Failed to create goroutine pool: %v" + err.Error())
	}
	defer pool.Release()

	// 创建EmbedderPooler实例
	pooler := embedding.NewBatchEmbedder(pool)

	// 创建OllamaService实例
	ollamaService, err := ollama.GetOllamaService()
	if err != nil {
		panic("Failed to create Ollama service: %v" + err.Error())
	}

	// 传递依赖项给NewEmbedder
	embedder, err := embedding.NewEmbedder(config, pooler, ollamaService)
	if err != nil {
		panic("Failed to create embedder: %v" + err.Error())
	}
	if embedder == nil {
		panic("Embedder is nil but no error was returned")
	}

	// 暂时跳过实际嵌入测试，先确保嵌入器初始化成功
	println("Embedder initialized successfully")

	// 测试BatchEmbed方法
	embeddings, err := embedder.BatchEmbed(context.Background(), []string{"hello", "world"})
	if err != nil {
		panic("Failed to get embeddings: %v" + err.Error())
	}
	if len(embeddings) != 2 {
		panic("Expected 2 embeddings, got %d" + fmt.Sprintf("%d", len(embeddings)))
	}

	// 检查嵌入向量维度
	for i, embedding := range embeddings {
		if len(embedding) != config.Dimensions {
			panic("Embedding %d has wrong dimensions: expected %d, got %d" + fmt.Sprintf("%d", i) + fmt.Sprintf("%d", config.Dimensions) + fmt.Sprintf("%d", len(embedding)))
		}
	}
}
