package retriever

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/jettjia/go-pkg/aipkg/pkg/rrf"
	"github.com/jettjia/go-pkg/aipkg/schema"
)

// elasticsearchRepository implements the RetrieveEngineRepository interface for Elasticsearch v8
type elasticsearchRepository struct {
	client           *elasticsearch.TypedClient // Elasticsearch client instance
	index            string                     // Name of the Elasticsearch index to use
	indexFullContent string                     // Name of the Elasticsearch index to use for full content
}

// NewElasticsearchEngineRepository creates and initializes a new Elasticsearch v8 repository
// It sets up the index and returns a repository instance ready for use
func NewElasticsearchEngineRepository(client *elasticsearch.TypedClient) RetrieveEngineRepository {
	// Get index name from environment variable or use default
	indexName := os.Getenv("ELASTICSEARCH_INDEX")
	if indexName == "" {
		slog.Warn("[Elasticsearch] ELASTICSEARCH_INDEX environment variable not set, using default index name")
		indexName = "doc_embedding"
	}

	indexFullContent := os.Getenv("ELASTICSEARCH_INDEX_FULL_CONTENT")
	if indexFullContent == "" {
		slog.Warn("[Elasticsearch] ELASTICSEARCH_INDEX_FULL_CONTENT environment variable not set, using default index name")
		indexFullContent = "doc_content"
	}

	// Create repository instance and ensure index exists
	res := &elasticsearchRepository{client: client, index: indexName, indexFullContent: indexFullContent}
	if err := res.createIndexIfNotExists(context.Background()); err != nil {
		slog.Error("[Elasticsearch] Failed to create index", "error", err)
	} else {
		slog.Info("[Elasticsearch] Successfully initialized repository")
	}
	return res
}

func (e *elasticsearchRepository) EngineType() schema.RetrieverEngineType {
	return schema.ElasticsearchRetrieverEngineType
}

// Support returns the retrieval types supported by this repository (Keywords, Vector, and Hybrid)
func (e *elasticsearchRepository) Support() []schema.RetrieverType {
	return []schema.RetrieverType{schema.KeywordsRetrieverType, schema.VectorRetrieverType, schema.HybridRetrieverType}
}

func (e *elasticsearchRepository) Save(ctx context.Context, doc *schema.Document, params map[string]any) error {
	embeddingDB := schema.ToDBVectorEmbeddingOpensearch(doc, params)

	if embeddingDB == nil {
		return fmt.Errorf("[Elasticsearch] Save embeddingDB is nil")
	}

	// Index the document
	resp, err := e.client.Index(e.index).Request(embeddingDB).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to save index", "index", e.index, "err", err)
		return err
	}

	slog.Info("[Elasticsearch] Successfully saved index for chunk ID: %s, document ID: %s", doc.ChunkId, resp.Id_)
	return nil
}

// BatchSave 批量保存文档
func (e *elasticsearchRepository) BatchSave(ctx context.Context, docs []*schema.Document, params map[string]any) error {
	if len(docs) == 0 {
		return nil
	}

	for _, doc := range docs {
		if err := e.Save(ctx, doc, params); err != nil {
			slog.Error("[Elasticsearch] Failed to batch save document", "id", doc.DocID, "error", err)
			return err
		}
	}
	return nil
}

// DeleteByChunkIDList 根据chunk ID列表删除文档
func (e *elasticsearchRepository) DeleteByChunkIDList(ctx context.Context, chunkIDs []string, params map[string]any) error {
	if len(chunkIDs) == 0 {
		return nil
	}

	_, err := e.client.DeleteByQuery(e.index).Query(&types.Query{
		Terms: &types.TermsQuery{TermsQuery: map[string]types.TermsQueryField{"chunk_id.keyword": chunkIDs}},
	}).Do(ctx)

	if err != nil {
		slog.Error("[Elasticsearch] Failed to delete documents by chunk ID list", "chunkIDs", chunkIDs, "error", err)
		return err
	}
	return nil
}

// DeleteByKnowledgeID 根据knowledge ID删除文档
func (e *elasticsearchRepository) DeleteByKnowledgeID(ctx context.Context, knowledgeID string, params map[string]any) error {
	if knowledgeID == "" {
		return nil
	}

	_, err := e.client.DeleteByQuery(e.index).Query(&types.Query{
		Terms: &types.TermsQuery{TermsQuery: map[string]types.TermsQueryField{"knowledge_id.keyword": []string{knowledgeID}}},
	}).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to delete documents by knowledge ID", "knowledgeID", knowledgeID, "error", err)
		return err
	}
	return nil
}

// DeleteByID 根据文档ID删除文档
func (e *elasticsearchRepository) DeleteByID(ctx context.Context, docID string, params map[string]any) error {
	// 实现删除逻辑
	if docID == "" {
		return nil
	}
	_, err := e.client.DeleteByQuery(e.index).Query(&types.Query{
		Terms: &types.TermsQuery{TermsQuery: map[string]types.TermsQueryField{"doc_id.keyword": []string{docID}}},
	}).Do(ctx)

	if err != nil {
		slog.Error("[Elasticsearch] Failed to delete documents by doc ID", "docID", docID, "error", err)
		return err
	}
	return nil
}

// createIndexIfNotExists checks if the specified index exists and creates it if not
// Returns an error if the operation fails
func (e *elasticsearchRepository) createIndexIfNotExists(ctx context.Context) error {
	slog.Debug("[Elasticsearch] Checking if index exists", "index", e.index)

	// Check if index exists
	exists, err := e.client.Indices.Exists(e.index).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to check if index exists", "index", e.index, "error", err)
		return err
	}

	if exists {
		slog.Debug("[Elasticsearch] Index already exists", "index", e.index)
		return nil
	}

	// Create index if it doesn't exist
	slog.Info("[Elasticsearch] Creating index", "index", e.index)
	_, err = e.client.Indices.Create(e.index).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to create index", "index", e.index, "error", err)
		return err
	}

	slog.Info("[Elasticsearch] Index created successfully", "index", e.index)
	return nil
}

func (e *elasticsearchRepository) SaveFullContent(ctx context.Context, doc *schema.Document) error {
	embeddingDB := schema.ToDBVectorEmbeddingOpensearchDoc(doc)

	if embeddingDB == nil {
		return fmt.Errorf("[Elasticsearch] SaveFullContent embeddingDB is nil")
	}

	// Index the document
	resp, err := e.client.Index(e.indexFullContent).Request(embeddingDB).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to save index", "index", e.indexFullContent, "err", err)
		return err
	}
	slog.Info("[Elasticsearch] Successfully saved index for doc ID: %s, document ID: %s", doc.DocID, resp.Id_)
	return nil
}

// DeleteFullContentByID 删除文档
func (e *elasticsearchRepository) DeleteFullContentByID(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("[Elasticsearch] DeleteFullContentByID id is empty")
	}

	_, err := e.client.DeleteByQuery(e.indexFullContent).Query(&types.Query{
		Terms: &types.TermsQuery{TermsQuery: map[string]types.TermsQueryField{"doc_id.keyword": []string{id}}},
	}).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to delete documents by doc ID", "docID", id, "error", err)
		return err
	}
	return nil
}

func (e *elasticsearchRepository) DeleteFullContentByKnowledgeID(ctx context.Context, knowledgeID string) error {
	if knowledgeID == "" {
		return fmt.Errorf("[Elasticsearch] DeleteFullContentByKnowledgeID knowledgeID is empty")
	}

	_, err := e.client.DeleteByQuery(e.indexFullContent).Query(&types.Query{
		Terms: &types.TermsQuery{TermsQuery: map[string]types.TermsQueryField{"knowledge_id.keyword": []string{knowledgeID}}},
	}).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to delete documents by knowledge ID", "knowledgeID", knowledgeID, "error", err)
		return err
	}
	return nil
}

func (e *elasticsearchRepository) Retrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
	switch params.RetrieverType {
	case schema.KeywordsRetrieverType:
		return e.KeywordsRetrieve(ctx, params)
	case schema.VectorRetrieverType:
		return e.VectorRetrieve(ctx, params)
	case schema.HybridRetrieverType:
		return e.HybridRetrieve(ctx, params)
	default:
		return e.HybridRetrieve(ctx, params)
	}
}

// VectorRetrieve performs vector similarity search using cosine similarity
// Returns a slice of RetrieveResult containing matching documents
func (e *elasticsearchRepository) VectorRetrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
	filter := e.getBaseConds(params)

	// Build script scoring query with cosine similarity
	queryVectorJSON, err := json.Marshal(params.Embedding)
	if err != nil {
		slog.Error("[Elasticsearch] Failed to marshal query vector", "err", err)
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	scoreSource := "cosineSimilarity(params.query_vector, 'embedding')"
	minScore := float32(params.Threshold)
	scriptScore := &types.ScriptScoreQuery{
		Query: types.Query{Bool: &types.BoolQuery{Filter: filter}},
		Script: types.Script{
			Source: &scoreSource,
			Params: map[string]json.RawMessage{
				"query_vector": json.RawMessage(queryVectorJSON),
			},
		},
		MinScore: &minScore,
	}

	// Execute search with minimum score threshold
	size := int(params.TopK)
	response, err := e.client.Search().Index(e.index).Request(&search.Request{
		Query: &types.Query{ScriptScore: scriptScore},
		Size:  &size,
	}).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Vector search failed", "err", err)
		return nil, err
	}

	// Process search results
	var results []*schema.Document
	for _, hit := range response.Hits.Hits {
		var embedding *schema.Document
		if err := json.Unmarshal(hit.Source_, &embedding); err != nil {
			slog.Error("[Elasticsearch] Failed to unmarshal search result", "err", err)
			return nil, err
		}
		embedding.Score = float64(*hit.Score_)
		results = append(results, embedding)
	}

	if len(results) == 0 {
		slog.Warn("[Elasticsearch] No vector matches found that meet threshold", "threshold", params.Threshold)
	} else {
		slog.Info("[Elasticsearch] Vector retrieval found %d results", "topK", params.TopK, "len(results)", len(results))
		slog.Debug("[Elasticsearch] Top result score: %.4f", "score", results[0].Score)
	}

	return []*schema.RetrieveResult{
		{
			Results:             results,
			RetrieverEngineType: schema.ElasticsearchRetrieverEngineType,
			RetrieverType:       schema.VectorRetrieverType,
			Error:               nil,
		},
	}, nil
}

// KeywordsRetrieve performs keyword-based search in document content
// Returns a slice of RetrieveResult containing matching documents
func (e *elasticsearchRepository) KeywordsRetrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
	filter := e.getBaseConds(params)
	// Build must conditions for content matching
	must := []types.Query{
		{Match: map[string]types.MatchQuery{"content": {Query: params.Query}}},
	}

	size := int(params.TopK)
	response, err := e.client.Search().Index(e.index).Request(&search.Request{
		Query: &types.Query{Bool: &types.BoolQuery{Filter: filter, Must: must}},
		Size:  &size,
	}).Do(ctx)
	if err != nil {
		slog.Error("[Elasticsearch] Keywords search failed", "err", err)
		return nil, err
	}

	// Process search results
	var results []*schema.Document
	for _, hit := range response.Hits.Hits {
		var doc schema.Document
		if err := json.Unmarshal(hit.Source_, &doc); err != nil {
			slog.Error("[Elasticsearch] Failed to unmarshal search result", "err", err)
			return nil, err
		}
		doc.Score = float64(*hit.Score_)
		results = append(results, &doc)
	}

	if len(results) == 0 {
		slog.Warn("[Elasticsearch] No keyword matches found for query", "query", params.Query)
	} else {
		slog.Info("[Elasticsearch] Keywords retrieval found %d results", "topK", params.TopK, "len(results)", len(results))
		slog.Debug("[Elasticsearch] Top result score: %.4f", "score", results[0].Score)
	}

	return []*schema.RetrieveResult{
		{
			Results:             results,
			RetrieverEngineType: schema.ElasticsearchRetrieverEngineType,
			RetrieverType:       schema.KeywordsRetrieverType,
			Error:               nil,
		},
	}, nil
}

// HybridRetrieve performs hybrid search combining keyword and vector similarity
// Returns a slice of RetrieveResult containing matching documents
// HybridRetrieve performs hybrid search by combining keyword and vector search results
// using RRF (Reciprocal Rank Fusion) algorithm
func (e *elasticsearchRepository) HybridRetrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
	// 获取关键词检索和向量检索结果
	keywordResults, keywordErr := e.KeywordsRetrieve(ctx, params)
	vectorResults, vectorErr := e.VectorRetrieve(ctx, params)

	// 准备错误信息
	var err error
	if keywordErr != nil && vectorErr != nil {
		err = fmt.Errorf("both keyword and vector retrieve failed: %w, %w", keywordErr, vectorErr)
	} else if keywordErr != nil {
		err = keywordErr
	} else if vectorErr != nil {
		err = vectorErr
	}

	// 提取文档数据
	var keywordDocs, vectorDocs []*schema.Document
	if len(keywordResults) > 0 {
		keywordDocs = keywordResults[0].Results
	}
	if len(vectorResults) > 0 {
		vectorDocs = vectorResults[0].Results
	}

	// 准备RRF输入 - 修改为单个数组
	rerankReq := &schema.RerankRequest{
		Data: [][]*schema.RerankData{
			{}, // 合并后的检索结果
		},
		TopN: &params.TopK,
	}

	// 合并关键词和向量检索结果
	allDocs := append(keywordDocs, vectorDocs...)

	// 转换为RerankData
	for _, doc := range allDocs {
		rerankReq.Data[0] = append(rerankReq.Data[0], &schema.RerankData{
			Document: doc,
			Score:    doc.Score,
		})
	}

	// 创建RRF重排器
	reranker := rrf.NewRRFReranker(60) // 使用默认k值60

	// 执行RRF融合
	resp, rrfErr := reranker.Rerank(ctx, rerankReq)
	if rrfErr != nil || resp == nil {
		// 如果RRF失败，返回合并后的结果
		mergedDocs := allDocs // 直接使用已合并的结果
		if len(mergedDocs) > int(params.TopK) {
			mergedDocs = mergedDocs[:params.TopK]
		}
		return []*schema.RetrieveResult{
				{
					Results:             mergedDocs,
					RetrieverEngineType: schema.ElasticsearchRetrieverEngineType,
					RetrieverType:       schema.HybridRetrieverType,
					Error:               fmt.Errorf("rrf fusion failed: %w, %w", rrfErr, err),
				},
			},
			nil
	}

	// 提取融合后的文档
	var hybridDocs []*schema.Document
	for _, item := range resp.SortedData {
		hybridDocs = append(hybridDocs, item.Document)
		if len(hybridDocs) >= int(params.TopK) {
			break
		}
	}

	return []*schema.RetrieveResult{
			{
				Results:             hybridDocs,
				RetrieverEngineType: schema.ElasticsearchRetrieverEngineType,
				RetrieverType:       schema.HybridRetrieverType,
				Error:               err,
			},
		},
		nil
}

// getBaseConds creates the base query conditions for retrieval operations
// Returns a slice of Query objects with must and must_not conditions
func (e *elasticsearchRepository) getBaseConds(params schema.RetrieveParams) []types.Query {
	must := []types.Query{}
	mustNot := make([]types.Query, 0)
	if len(params.ExcludeKnowledgeIDs) > 0 {
		mustNot = append(mustNot, types.Query{Terms: &types.TermsQuery{
			TermsQuery: map[string]types.TermsQueryField{"knowledge_id.keyword": params.ExcludeKnowledgeIDs},
		}})
	}
	if len(params.ExcludeChunkIDs) > 0 {
		mustNot = append(mustNot, types.Query{Terms: &types.TermsQuery{
			TermsQuery: map[string]types.TermsQueryField{"chunk_id.keyword": params.ExcludeChunkIDs},
		}})
	}
	return []types.Query{{Bool: &types.BoolQuery{Must: must, MustNot: mustNot}}}
}
