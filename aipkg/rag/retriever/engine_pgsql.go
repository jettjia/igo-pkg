package retriever

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/jettjia/go-pkg/aipkg/pkg/rrf"
	"github.com/jettjia/go-pkg/aipkg/schema"
	"github.com/jettjia/go-pkg/pkg/util"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// pgRepository implements PostgreSQL-based retrieval operations
type pgRepository struct {
	db *gorm.DB // Database connection
}

// NewPostgresRetrieveEngineRepository creates a new PostgreSQL retriever repository
func NewPostgresRetrieveEngineRepository(db *gorm.DB) RetrieveEngineRepository {
	slog.Info("[Postgres] Initializing PostgreSQL retriever engine repository")
	return &pgRepository{db: db}
}

// EngineType returns the retriever engine type (PostgreSQL)
func (r *pgRepository) EngineType() schema.RetrieverEngineType {
	return schema.PostgresRetrieverEngineType
}

// Support returns the retrieval types supported by this repository (Keywords, Vector, and Hybrid)
func (r *pgRepository) Support() []schema.RetrieverType {
	return []schema.RetrieverType{schema.KeywordsRetrieverType, schema.VectorRetrieverType, schema.HybridRetrieverType}
}

func (e *pgRepository) Save(ctx context.Context, doc *schema.Document, params map[string]any) error {
	if doc.ID == "" {
		doc.ID = util.Ulid()
	}
	embeddingDB := schema.ToDBVectorEmbeddingPgSQL(doc, params)
	if embeddingDB == nil {
		return fmt.Errorf("embeddingDB is nil")
	}

	if err := e.db.Create(embeddingDB).Error; err != nil {
		return fmt.Errorf("create embeddingDB failed: %w", err)
	}

	return nil
}

func (e *pgRepository) BatchSave(ctx context.Context, docs []*schema.Document, params map[string]any) error {
	if len(docs) == 0 {
		return fmt.Errorf("docs is empty")
	}

	embeddingDBs := make([]*schema.PgVector, 0, len(docs))
	for _, doc := range docs {
		embeddingDB := schema.ToDBVectorEmbeddingPgSQL(doc, params)
		if embeddingDB == nil {
			return fmt.Errorf("embeddingDB is nil")
		}
		embeddingDBs = append(embeddingDBs, embeddingDB)
	}

	if err := e.db.Create(embeddingDBs).Error; err != nil {
		return fmt.Errorf("create embeddingDBs failed: %w", err)
	}

	return nil
}

func (e *pgRepository) DeleteByChunkIDList(ctx context.Context, chunkIDs []string, params map[string]any) error {
	if len(chunkIDs) == 0 {
		return fmt.Errorf("chunkIDs is empty")
	}

	if err := e.db.Delete(&schema.PgVector{}, "chunk_id IN ?", chunkIDs).Error; err != nil {
		return fmt.Errorf("delete embeddingDBs failed: %w", err)
	}

	return nil
}

func (e *pgRepository) DeleteByKnowledgeID(ctx context.Context, knowledgeID string, params map[string]any) error {
	if knowledgeID == "" {
		return fmt.Errorf("knowledgeID is empty")
	}

	if err := e.db.Delete(&schema.PgVector{}, "knowledge_id = ?", knowledgeID).Error; err != nil {
		return fmt.Errorf("delete embeddingDBs failed: %w", err)
	}

	return nil
}

func (e *pgRepository) DeleteByID(ctx context.Context, docID string, params map[string]any) error {
	if docID == "" {
		return fmt.Errorf("docID is empty")
	}

	if err := e.db.Delete(&schema.PgVector{}, "ulid = ?", docID).Error; err != nil {
		return fmt.Errorf("delete embeddingDB failed: %w", err)
	}

	return nil
}

// SaveFullContent saves the full content
func (e *pgRepository) SaveFullContent(ctx context.Context, doc *schema.Document) error {
	if doc.ID == "" {
		doc.ID = util.Ulid()
	}
	embeddingDB := schema.ToDBVectorEmbeddingPgSQLDoc(doc)
	if embeddingDB == nil {
		return fmt.Errorf("embeddingDB is nil")
	}

	if err := e.db.Create(embeddingDB).Error; err != nil {
		return fmt.Errorf("create embeddingDB failed: %w", err)
	}

	return nil
}

// DeleteFullContentByID deletes the full content by id
func (e *pgRepository) DeleteFullContentByID(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is empty")
	}

	if err := e.db.Delete(&schema.PgVectorDoc{}, "ulid = ?", id).Error; err != nil {
		return fmt.Errorf("delete embeddingDB failed: %w", err)
	}

	return nil
}

// DeleteFullContentByKnowledgeID deletes the full content by knowledge id
func (e *pgRepository) DeleteFullContentByKnowledgeID(ctx context.Context, knowledgeID string) error {
	if knowledgeID == "" {
		return fmt.Errorf("knowledgeID is empty")
	}

	if err := e.db.Delete(&schema.PgVectorDoc{}, "knowledge_id = ?", knowledgeID).Error; err != nil {
		return fmt.Errorf("delete embeddingDB failed: %w", err)
	}

	return nil
}

func (e *pgRepository) Retrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
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

// KeywordsRetrieve performs keyword-based search using PostgreSQL full-text search
func (g *pgRepository) KeywordsRetrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
	slog.Debug("[Postgres] KeywordsRetrieve", "query", params.Query, "topK", params.TopK)
	var embeddingDBList []schema.PgVectorDocScore

	// 使用结构体的TableName方法获取表名
	tableName := new(schema.PgVectorDocScore).TableName()

	// 修改查询语句，确保正确使用jiebacfg
	query := fmt.Sprintf(`
        SELECT ulid, content,doc_id, chunk_id, knowledge_id,
               ts_rank(to_tsvector('jiebacfg', content), plainto_tsquery('jiebacfg', ?)) as score
        FROM %s
        WHERE to_tsvector('jiebacfg', content) @@ plainto_tsquery('jiebacfg', ?)
        ORDER BY score DESC
        LIMIT ?
    `, tableName)

	// 添加调试日志
	// fmt.Println("[Postgres] KeywordsRetrieve SQL", query)

	err := g.db.WithContext(ctx).Raw(query, params.Query, params.Query, int(params.TopK)).Scan(&embeddingDBList).Error

	// fmt.Println("[Postgres] KeywordsRetrieve results count", len(embeddingDBList))

	if err == gorm.ErrRecordNotFound {
		slog.Warn("[Postgres] No records found for keywords query: %s, topK: %d", params.Query, params.TopK)
		return nil, nil
	}
	if err != nil {
		slog.Error("[Postgres] Keywords retrieval failed", "err", err.Error())
		return nil, err
	}

	results := make([]*schema.Document, len(embeddingDBList))
	for i := range embeddingDBList {
		results[i] = &schema.Document{
			ID:          embeddingDBList[i].Ulid,
			Content:     embeddingDBList[i].Content,
			DocID:       embeddingDBList[i].DocID,
			ChunkId:     embeddingDBList[i].ChunkID,
			KnowledgeID: embeddingDBList[i].KnowledgeID,
			Score:       embeddingDBList[i].Score,
		}
	}
	return []*schema.RetrieveResult{
		{
			Results:             results,
			RetrieverEngineType: schema.PostgresRetrieverEngineType,
			RetrieverType:       schema.KeywordsRetrieverType,
			Error:               nil,
		},
	}, nil
}

// VectorRetrieve performs vector similarity search using pgvector
func (g *pgRepository) VectorRetrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
	// fmt.Println("[Postgres] VectorRetrieve", params.Query, params.TopK, params.Threshold, len(params.Embedding))

	conds := make([]clause.Expression, 0)

	// 检查Embedding是否为空
	if len(params.Embedding) == 0 {
		return nil, fmt.Errorf("embedding vector cannot be empty")
	}

	dimension := len(params.Embedding)

	// 添加维度匹配条件
	conds = append(conds, clause.Expr{SQL: "dimension = ?", Vars: []interface{}{dimension}})

	// 计算实际阈值 (1 - params.Threshold) 并添加日志
	similarityThreshold := 1 - params.Threshold
	// fmt.Println("[Postgres] VectorRetrieve similarity threshold", similarityThreshold)

	// 添加相似度条件
	conds = append(conds, clause.Expr{
		SQL:  fmt.Sprintf("embedding::halfvec(%d) <=> ?::halfvec < ?", dimension),
		Vars: []interface{}{pgvector.NewHalfVector(params.Embedding), similarityThreshold},
	})

	// 添加排序条件
	conds = append(conds, clause.OrderBy{Expression: clause.Expr{
		SQL:  fmt.Sprintf("embedding::halfvec(%d) <=> ?::halfvec", dimension),
		Vars: []interface{}{pgvector.NewHalfVector(params.Embedding)},
	}})

	var embeddingDBList []schema.PgVectorScore

	// 构建查询并执行
	query := g.db.WithContext(ctx).Clauses(conds...).
		Select(fmt.Sprintf(
			"ulid, content,doc_id, chunk_id, knowledge_id, "+"(1 - (embedding::halfvec(%d) <=> ?::halfvec)) as score",
			dimension,
		), pgvector.NewHalfVector(params.Embedding)).
		Limit(int(params.TopK))

	// fmt.Println("[Postgres] VectorRetrieve query", query.Statement.SQL.String(), "params", query.Statement.Vars)

	// 执行查询
	err := query.Find(&embeddingDBList).Error

	// fmt.Println("[Postgres] VectorRetrieve results count", len(embeddingDBList))

	if err == gorm.ErrRecordNotFound {
		slog.Warn("[Postgres] No vector matches found that meet threshold", "threshold", params.Threshold)
		return nil, nil
	}
	if err != nil {
		slog.Error("[Postgres] Vector retrieval failed", "err", err.Error())
		return nil, err
	}

	// 转换结果
	results := make([]*schema.Document, len(embeddingDBList))
	for i := range embeddingDBList {
		results[i] = &schema.Document{
			ID:          embeddingDBList[i].Ulid,
			Content:     embeddingDBList[i].Content,
			DocID:       embeddingDBList[i].DocID,
			ChunkId:     embeddingDBList[i].ChunkID,
			KnowledgeID: embeddingDBList[i].KnowledgeID,
			Score:       embeddingDBList[i].Score,
		}
	}

	return []*schema.RetrieveResult{
		{
			Results:             results,
			RetrieverEngineType: schema.PostgresRetrieverEngineType,
			RetrieverType:       schema.VectorRetrieverType,
			Error:               nil,
		},
	}, nil
}

func (g *pgRepository) HybridRetrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error) {
	// 获取关键词检索和向量检索结果
	keywordResults, keywordErr := g.KeywordsRetrieve(ctx, params)
	vectorResults, vectorErr := g.VectorRetrieve(ctx, params)

	// 准备错误信息
	var err error
	if keywordErr != nil && vectorErr != nil {
		err = fmt.Errorf("both keyword and vector retrieve failed: %w, %w", keywordErr, vectorErr)
		return nil, err
	} else if keywordErr != nil {
		err = keywordErr
		// 即使关键词检索失败，也尝试返回向量检索结果
		if len(vectorResults) > 0 {
			return vectorResults, err
		}
		return nil, err
	} else if vectorErr != nil {
		err = vectorErr
		// 即使向量检索失败，也尝试返回关键词检索结果
		if len(keywordResults) > 0 {
			return keywordResults, err
		}
		return nil, err
	}

	// 提取文档数据
	var keywordDocs, vectorDocs []*schema.Document
	if len(keywordResults) > 0 {
		keywordDocs = keywordResults[0].Results
	}
	if len(vectorResults) > 0 {
		vectorDocs = vectorResults[0].Results
	}

	// 准备RRF输入
	rerankReq := &schema.RerankRequest{
		Data: [][]*schema.RerankData{
			{}, // 合并后的检索结果
		},
		TopN: &params.TopK,
	}

	// 合并关键词检索结果和向量检索结果
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
	resp, err := reranker.Rerank(ctx, rerankReq)
	if err != nil || resp == nil || len(resp.SortedData) == 0 {
		// 如果RRF失败，合并结果并按得分排序
		mergedResults := allDocs
		sort.Slice(mergedResults, func(i, j int) bool {
			return mergedResults[i].Score > mergedResults[j].Score
		})
		// 去重
		deduplicated := make([]*schema.Document, 0)
		seen := make(map[string]bool)
		for _, doc := range mergedResults {
			if !seen[doc.ID] {
				seen[doc.ID] = true
				deduplicated = append(deduplicated, doc)
			}
		}
		// 截断到TopK
		if len(deduplicated) > int(params.TopK) {
			deduplicated = deduplicated[:params.TopK]
		}
		return []*schema.RetrieveResult{
			{
				Results:             deduplicated,
				RetrieverEngineType: schema.PostgresRetrieverEngineType,
				RetrieverType:       schema.HybridRetrieverType,
				Error:               err,
			},
		}, nil
	}

	// 提取融合后的文档并设置RRF分数
	finalResults := make([]*schema.Document, len(resp.SortedData))
	for i, data := range resp.SortedData {
		finalResults[i] = data.Document
		finalResults[i].Score = data.RRFScore
	}

	// 确保结果数量不超过TopK
	if len(finalResults) > int(params.TopK) {
		finalResults = finalResults[:params.TopK]
	}

	fmt.Println("HybridRetrieve final results count", len(finalResults))

	return []*schema.RetrieveResult{
		{
			Results:             finalResults,
			RetrieverEngineType: schema.PostgresRetrieverEngineType,
			RetrieverType:       schema.HybridRetrieverType,
			Error:               nil,
		},
	}, nil
}
