package rrf

import (
	"context"
	"fmt"
	"testing"

	"github.com/jettjia/igo-pkg/aipkg/pkg/ptr"
	"github.com/jettjia/igo-pkg/aipkg/schema"
	"github.com/stretchr/testify/assert"
)

// TestRRF_Rerank 测试RRF重排功能
// go test -v -run TestRRF_Rerank ./
func TestRRF_Rerank(t *testing.T) {
	// 创建RRF实例
	rrf := NewRRFReranker(60)
	ctx := context.Background()

	// 测试场景1: 基本重排功能
	doc1 := &schema.Document{ID: "doc1", Content: "Document 1 content"}
	doc2 := &schema.Document{ID: "doc2", Content: "Document 2 content"}
	doc3 := &schema.Document{ID: "doc3", Content: "Document 3 content"}

	req1 := &schema.RerankRequest{
		Query: "test query",
		Data: [][]*schema.RerankData{
			{
				{
					Document: doc1,
					Score:    1.0,
				},
				{
					Document: doc2,
					Score:    0.8,
				},
			},
			{
				{
					Document: doc2,
					Score:    1.0,
				},
				{
					Document: doc3,
					Score:    0.9,
				},
			},
		},
		TopN: ptr.Of(int64(3)),
	}

	resp1, err := rrf.Rerank(ctx, req1)
	assert.NoError(t, err)
	assert.NotNil(t, resp1)
	assert.Len(t, resp1.SortedData, 3)

	// 验证重排顺序是否符合预期 (修正后doc2应该排第一)
	assert.Equal(t, "doc2", resp1.SortedData[0].Document.ID)
	assert.Equal(t, "doc1", resp1.SortedData[1].Document.ID)
	assert.Equal(t, "doc3", resp1.SortedData[2].Document.ID)

	// 测试场景2: 空请求
	req2 := &schema.RerankRequest{}
	resp2, err := rrf.Rerank(ctx, req2)
	assert.NoError(t, err) // 空请求现在应该返回nil错误
	assert.NotNil(t, resp2)
	assert.Empty(t, resp2.SortedData)

	// 测试场景3: 空数据
	req3 := &schema.RerankRequest{
		Data: [][]*schema.RerankData{},
		TopN: ptr.Of(int64(3)),
	}
	resp3, err := rrf.Rerank(ctx, req3)
	assert.NoError(t, err)
	assert.NotNil(t, resp3)
	assert.Empty(t, resp3.SortedData)

	// 测试场景4: TopN参数
	req4 := &schema.RerankRequest{
		Query: "test query",
		Data:  req1.Data,
		TopN:  ptr.Of(int64(2)),
	}
	resp4, err := rrf.Rerank(ctx, req4)
	assert.NoError(t, err)
	assert.NotNil(t, resp4)
	assert.Len(t, resp4.SortedData, 2)
	assert.Equal(t, "doc2", resp4.SortedData[0].Document.ID)
	assert.Equal(t, "doc1", resp4.SortedData[1].Document.ID)

	// 测试场景5: 不同的k值
	rrfWithK := NewRRFReranker(5)
	resp5, err := rrfWithK.Rerank(ctx, req1)
	assert.NoError(t, err)
	assert.NotNil(t, resp5)
	assert.Len(t, resp5.SortedData, 3)

	// 输出结果
	for _, data := range resp5.SortedData {
		t.Logf("ID: %s, Score: %f, RRFScore: %f", data.Document.ID, data.Score, data.RRFScore)
	}
}

// TestRRF_Rerank_SingleList 测试单结果列表的情况
// go test -v -run TestRRF_Rerank_SingleList ./
func TestRRF_Rerank_SingleList(t *testing.T) {
	rrf := NewRRFReranker(60)
	ctx := context.Background()

	doc1 := &schema.Document{ID: "doc1", Content: "Document 1 content"}
	doc2 := &schema.Document{ID: "doc2", Content: "Document 2 content"}

	req := &schema.RerankRequest{
		Query: "test query",
		Data: [][]*schema.RerankData{
			{
				{
					Document: doc1,
					Score:    1.0,
				},
				{
					Document: doc2,
					Score:    0.8,
				},
			},
		},
		TopN: ptr.Of(int64(2)),
	}

	resp, err := rrf.Rerank(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.SortedData, 2)
	assert.Equal(t, "doc1", resp.SortedData[0].Document.ID)
	assert.Equal(t, "doc2", resp.SortedData[1].Document.ID)
}

// TestRRF_Rerank_DuplicateDocs 测试重复文档的情况
// go test -v -run TestRRF_Rerank_DuplicateDocs ./
func TestRRF_Rerank_DuplicateDocs(t *testing.T) {
	rrf := NewRRFReranker(60)
	ctx := context.Background()

	doc1 := &schema.Document{ID: "doc1", Content: "Document 1 content"}
	doc2 := &schema.Document{ID: "doc2", Content: "Document 2 content"}
	doc3 := &schema.Document{ID: "doc3", Content: "Document 3 content"}

	req := &schema.RerankRequest{
		Query: "test query",
		Data: [][]*schema.RerankData{
			{
				{
					Document: doc1,
					Score:    1.0,
				},
				{
					Document: doc2,
					Score:    0.8,
				},
			},
			{
				{
					Document: doc1,
					Score:    0.9,
				},
				{
					Document: doc3,
					Score:    0.7,
				},
			},
		},
		TopN: ptr.Of(int64(3)),
	}

	resp, err := rrf.Rerank(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.SortedData, 3)

	// 验证doc1是否排名第一
	assert.Equal(t, "doc1", resp.SortedData[0].Document.ID)
	// 验证其他文档的排名
	assert.Contains(t, []string{"doc2", "doc3"}, resp.SortedData[1].Document.ID)
	assert.Contains(t, []string{"doc2", "doc3"}, resp.SortedData[2].Document.ID)

	fmt.Println(resp)
}
