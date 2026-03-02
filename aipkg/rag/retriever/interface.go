package retriever

import (
	"context"
	"fmt"

	"github.com/jettjia/go-pkg/aipkg/schema"
)

// RetrieverEngineType 表示检索引擎类型
// 支持多种存储后端
const (
	EngineTypePostgres   = "postgres"
	EngineTypeOpensearch = "opensearch"
)

// RetrieverType 表示检索类型
const (
	RetrieverTypeKeyword = "keyword" // 关键词检索
	RetrieverTypeVector  = "vector"  // 向量检索
	RetrieverTypeHybrid  = "hybrid"  // 混合检索
)

// RetrieveEngine defines the retrieve engine interface
type RetrieveEngine interface {
	// EngineType gets the retrieve engine type
	EngineType() schema.RetrieverEngineType

	// Retrieve executes the retrieve
	Retrieve(ctx context.Context, params schema.RetrieveParams) ([]*schema.RetrieveResult, error)

	// Support gets the supported retrieve types
	Support() []schema.RetrieverType
}

// RetrieverEngine 检索引擎接口
type RetrieveEngineRepository interface {
	// -------------------- //
	// doc chunk
	// Save saves the index info
	Save(ctx context.Context, doc *schema.Document, params map[string]any) error
	// BatchSave saves the index info list
	BatchSave(ctx context.Context, docs []*schema.Document, params map[string]any) error
	// DeleteByChunkIDList deletes the index info by chunk id list
	DeleteByChunkIDList(ctx context.Context, chunkIDs []string, params map[string]any) error
	// DeleteByKnowledgeID deletes the index info by knowledge id
	DeleteByKnowledgeID(ctx context.Context, knowledgeID string, params map[string]any) error
	// DeleteByID deletes the index info by id
	DeleteByID(ctx context.Context, id string, params map[string]any) error

	// -------------------- //
	// doc full content
	// SaveFullContent saves the full content
	SaveFullContent(ctx context.Context, doc *schema.Document) error
	// DeleteFullContentByID deletes the full content by id
	DeleteFullContentByID(ctx context.Context, id string) error
	// DeleteFullContentByKnowledgeID deletes the full content by knowledge id
	DeleteFullContentByKnowledgeID(ctx context.Context, knowledgeID string) error

	RetrieveEngine
}

// EngineRegistry 检索引擎注册表
// 支持动态注册/获取不同后端实现
type EngineRegistry struct {
	engines map[schema.RetrieverEngineType]RetrieveEngineRepository
}

func NewEngineRegistry() *EngineRegistry {
	return &EngineRegistry{engines: make(map[schema.RetrieverEngineType]RetrieveEngineRepository)}
}

func (r *EngineRegistry) Register(engine RetrieveEngineRepository) {
	r.engines[engine.EngineType()] = engine
}

func (r *EngineRegistry) Get(engineType schema.RetrieverEngineType) (RetrieveEngineRepository, error) {
	e, ok := r.engines[engineType]
	if !ok {
		return nil, fmt.Errorf("engine %s not found", engineType)
	}
	return e, nil
}
