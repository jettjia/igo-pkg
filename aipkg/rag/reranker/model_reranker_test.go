package reranker

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jettjia/go-pkg/aipkg/models/rerank"
	"github.com/jettjia/go-pkg/aipkg/pkg/types"
	"github.com/jettjia/go-pkg/aipkg/schema"
	"github.com/jettjia/go-pkg/pkg/util"
	"github.com/stretchr/testify/assert"
)

// 测试构造函数
// go test -v -run Test_NewModelReranker ./
func Test_NewModelReranker(t *testing.T) {
	config := &rerank.RerankerConfig{
		APIKey:    os.Getenv("OPENAI_API_KEY"),
		BaseURL:   os.Getenv("OPENAI_BASE_URL"),
		ModelName: "Qwen/Qwen3-Reranker-0.6B",
		Source:    types.ModelSourceRemote,
		ModelID:   "Qwen/Qwen3-Reranker-0.6B",
	}

	reranker, err := rerank.NewReranker(config)
	if err != nil {
		t.Fatalf("Failed to create reranker: %v", err)
	}

	modelReranker := NewModelReranker(reranker)
	if modelReranker == nil {
		t.Fatalf("Failed to create model reranker")
	}

	// 创建测试数据（二维数组）
	doc1 := &schema.Document{
		ID:      "1",
		DocID:   "doc1",
		Content: "文档内容1",
		Score:   0.5,
	}
	doc2 := &schema.Document{
		ID:      "2",
		DocID:   "doc2",
		Content: "你好，文档内容2",
		Score:   0.3,
	}

	// 测试Rerank方法
	topN := int64(2)
	req := &schema.RerankRequest{
		Query: "你好",
		Data: [][]*schema.RerankData{
			{{Document: doc1}, {Document: doc2}},
		},
		TopN: &topN,
	}

	resp, err := modelReranker.Rerank(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to rerank: %v", err)
	}
	if resp == nil {
		t.Fatalf("Rerank response is nil")
	}

	// 验证响应
	assert.NotEmpty(t, resp.SortedData)
	assert.Len(t, resp.SortedData, 2)

	// 打印结果
	fmt.Println(util.PrintJson(resp))
}
