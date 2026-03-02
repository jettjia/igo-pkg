package reranker

import (
	"context"
	"fmt"

	"github.com/jettjia/go-pkg/aipkg/models/rerank"
	"github.com/jettjia/go-pkg/aipkg/schema"
)

// ModelReranker 适配器，将 models/rerank.Reranker 适配到 reranker.Reranker 接口
type ModelReranker struct {
	modelReranker rerank.Reranker
}

// NewModelReranker 创建一个新的 ModelReranker 实例
func NewModelReranker(modelReranker rerank.Reranker) *ModelReranker {
	return &ModelReranker{
		modelReranker: modelReranker,
	}
}

// Rerank 实现 reranker.Reranker 接口
func (r *ModelReranker) Rerank(ctx context.Context, req *schema.RerankRequest) (*schema.RerankResponse, error) {
	if req == nil || req.Data == nil || len(req.Data) == 0 {
		return nil, fmt.Errorf("invalid request: no data provided")
	}

	// 提取所有文档内容
	documents := make([]string, 0)
	docMap := make(map[string]*schema.RerankData)
	docIndex := 0

	// 遍历二维数组
	for _, resultList := range req.Data {
		for _, item := range resultList {
			if item != nil && item.Document != nil {
				documents = append(documents, item.Document.Content)
				// 保存原始分数和文档
				docMap[fmt.Sprintf("%d", docIndex)] = &schema.RerankData{
					Document: item.Document,
					Score:    item.Score,
				}
				docIndex++
			}
		}
	}

	// 调用 models/rerank.Reranker 接口
	rankResults, err := r.modelReranker.Rerank(ctx, req.Query, documents)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &schema.RerankResponse{
		SortedData: make([]*schema.RerankData, 0, len(rankResults)),
		TokenUsage: new(int64), // 初始化TokenUsage
	}

	// 填充排序后的结果 - 直接使用模型返回顺序
	response.SortedData = make([]*schema.RerankData, len(rankResults))
	for i, result := range rankResults {
		// 使用rankResults中的Index字段匹配原始文档
		key := fmt.Sprintf("%d", result.Index)
		if data, exists := docMap[key]; exists {
			// 直接使用模型返回的相关性分数
			data.Score = result.RelevanceScore
			response.SortedData[i] = data
		}
	}

	// 应用TopN参数（不改变模型返回顺序）
	if req.TopN != nil && *req.TopN > 0 && *req.TopN < int64(len(response.SortedData)) {
		response.SortedData = response.SortedData[:*req.TopN]
	}

	// 尝试从rankResults中获取TokenUsage（假设它有这个字段）
	// *response.TokenUsage = ...

	return response, nil
}
