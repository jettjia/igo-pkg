package rrf

import (
	"context"
	"fmt"
	"sort"

	"github.com/jettjia/go-pkg/aipkg/pkg/ptr"
	"github.com/jettjia/go-pkg/aipkg/rag/reranker"
	"github.com/jettjia/go-pkg/aipkg/schema"
)

func NewRRFReranker(k int64) reranker.Reranker {
	if k == 0 {
		k = 60
	}
	return &rrfReranker{k}
}

type rrfReranker struct {
	k int64
}

// 修改 Rerank 方法
func (r *rrfReranker) Rerank(ctx context.Context, req *schema.RerankRequest) (*schema.RerankResponse, error) {
	if req == nil || req.Data == nil || len(req.Data) == 0 {
		return nil, fmt.Errorf("invalid request: no data provided")
	}
	id2Score := make(map[string]float64)
	id2Data := make(map[string]*schema.RerankData)
	for _, resultList := range req.Data {
		// 为不同检索类型设置权重 (关键词=1.0, 向量=1.0)
		weight := 1.0

		for rank := range resultList {
			result := resultList[rank]
			if result != nil && result.Document != nil {
				// RRF算法中排名从0开始
				score := weight / (float64(rank) + float64(r.k))

				// 累加分数而不是只保留最高分数
				id2Score[result.Document.ID] += score

				// 存储文档数据
				if _, exists := id2Data[result.Document.ID]; !exists {
					// 创建新的 RerankData 实例
					newData := &schema.RerankData{
						Document: result.Document,
						Score:    result.Score,
					}
					id2Data[result.Document.ID] = newData
				}
			}
		}
	}
	var sorted []*schema.RerankData
	for id, data := range id2Data {
		// 设置最终RRF分数
		data.RRFScore = id2Score[id]
		sorted = append(sorted, data)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RRFScore > sorted[j].RRFScore
	})
	topN := int64(len(sorted))
	if req.TopN != nil && ptr.From(req.TopN) != 0 && ptr.From(req.TopN) < topN {
		topN = ptr.From(req.TopN)
	}

	return &schema.RerankResponse{SortedData: sorted[:topN]}, nil
}
