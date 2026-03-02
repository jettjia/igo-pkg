package schema

import (
	"maps"
	"slices"

	"github.com/pgvector/pgvector-go"
)

// PgVector defines the database model for vector embeddings storage
type PgVector struct {
	Ulid        string              `gorm:"column:ulid;primaryKey;type:varchar(128);comment:ulid;" json:"ulid"`
	CreatedAt   int64               `gorm:"column:created_at;autoCreateTime:milli;type:bigint;comment:创建时间;" json:"created_at"`
	UpdatedAt   int64               `gorm:"column:updated_at;autoUpdateTime:milli;type:bigint;comment:修改时间;" json:"updated_at"`
	DeletedAt   int64               `gorm:"column:deleted_at;autoDeletedTime:milli;type:bigint;comment:删除时间;" json:"deleted_at"`
	CreatedBy   string              `gorm:"column:created_by;type:varchar(32);comment:创建者;" json:"created_by"`
	UpdatedBy   string              `gorm:"column:updated_by;type:varchar(32);comment:修改者;" json:"updated_by"`
	DeletedBy   string              `gorm:"column:deleted_by;type:varchar(32);comment:删除者;" json:"deleted_by"`
	DocID       string              `json:"doc_id" gorm:"column:doc_id"`
	ChunkID     string              `json:"chunk_id" gorm:"column:chunk_id"`
	KnowledgeID string              `json:"knowledge_id" gorm:"column:knowledge_id"`
	Content     string              `json:"content" gorm:"column:content;not null"`
	Dimension   int                 `json:"dimension" gorm:"column:dimension;not null"`
	Embedding   pgvector.HalfVector `json:"embedding" gorm:"column:embedding;not null"`
}

// ToDBVectorEmbeddingPgSQL converts IndexInfo to pgVector database model
func ToDBVectorEmbeddingPgSQL(indexInfo *Document, additionalParams map[string]any) *PgVector {
	pgVector := &PgVector{
		Ulid:        indexInfo.ID,
		DocID:       indexInfo.DocID,
		ChunkID:     indexInfo.ChunkId,
		KnowledgeID: indexInfo.KnowledgeID,
		Content:     indexInfo.Content,
	}
	// Add embedding data if available in additionalParams
	if additionalParams != nil && slices.Contains(slices.Collect(maps.Keys(additionalParams)), "embedding") {
		if embeddingMap, ok := additionalParams["embedding"].(map[string][]float32); ok {
			pgVector.Embedding = pgvector.NewHalfVector(embeddingMap[indexInfo.DocID])
			pgVector.Dimension = len(pgVector.Embedding.Slice())
		}
	}
	return pgVector
}

// TableName specifies the database table name for pgVector
func (PgVector) TableName() string {
	return "doc_embedding"
}

// PgVectorDoc 全文内容
type PgVectorDoc struct {
	Ulid        string `gorm:"column:ulid;primaryKey;type:varchar(128);comment:ulid;" json:"ulid"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:milli;type:bigint;comment:创建时间;" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at;autoUpdateTime:milli;type:bigint;comment:修改时间;" json:"updated_at"`
	DeletedAt   int64  `gorm:"column:deleted_at;autoDeletedTime:milli;type:bigint;comment:删除时间;" json:"deleted_at"`
	CreatedBy   string `gorm:"column:created_by;type:varchar(32);comment:创建者;" json:"created_by"`
	UpdatedBy   string `gorm:"column:updated_by;type:varchar(32);comment:修改者;" json:"updated_by"`
	DeletedBy   string `gorm:"column:deleted_by;type:varchar(32);comment:删除者;" json:"deleted_by"`
	ChunkID     string `json:"chunk_id" gorm:"column:chunk_id"`
	DocID       string `json:"doc_id" gorm:"column:doc_id"`
	KnowledgeID string `json:"knowledge_id" gorm:"column:knowledge_id"`
	Content     string `json:"content" gorm:"column:content;not null"`
}

// TableName specifies the database table name for pgVectorDoc
func (PgVectorDoc) TableName() string {
	return "doc_content"
}

// ToDBVectorEmbeddingPgSQLDoc 全文内容保存
func ToDBVectorEmbeddingPgSQLDoc(indexInfo *Document) *PgVectorDoc {
	pgVector := &PgVectorDoc{
		Ulid:        indexInfo.ID,
		DocID:       indexInfo.DocID,
		KnowledgeID: indexInfo.KnowledgeID,
		Content:     indexInfo.Content,
	}

	return pgVector
}

type PgVectorScore struct {
	Ulid        string              `gorm:"column:ulid;primaryKey;type:varchar(128);comment:ulid;" json:"ulid"`
	CreatedAt   int64               `gorm:"column:created_at;autoCreateTime:milli;type:bigint;comment:创建时间;" json:"created_at"`
	UpdatedAt   int64               `gorm:"column:updated_at;autoUpdateTime:milli;type:bigint;comment:修改时间;" json:"updated_at"`
	DeletedAt   int64               `gorm:"column:deleted_at;autoDeletedTime:milli;type:bigint;comment:删除时间;" json:"deleted_at"`
	CreatedBy   string              `gorm:"column:created_by;type:varchar(32);comment:创建者;" json:"created_by"`
	UpdatedBy   string              `gorm:"column:updated_by;type:varchar(32);comment:修改者;" json:"updated_by"`
	DeletedBy   string              `gorm:"column:deleted_by;type:varchar(32);comment:删除者;" json:"deleted_by"`
	DocID       string              `json:"doc_id" gorm:"column:doc_id"`
	ChunkID     string              `json:"chunk_id" gorm:"column:chunk_id"`
	KnowledgeID string              `json:"knowledge_id" gorm:"column:knowledge_id"`
	Content     string              `json:"content" gorm:"column:content;not null"`
	Dimension   int                 `json:"dimension" gorm:"column:dimension;not null"`
	Embedding   pgvector.HalfVector `json:"embedding" gorm:"column:embedding;not null"`
	Score       float64             `json:"score" gorm:"column:score"`
}

func (PgVectorScore) TableName() string {
	return "doc_embedding"
}

// PgVectorDoc 全文内容
type PgVectorDocScore struct {
	Ulid        string  `gorm:"column:ulid;primaryKey;type:varchar(128);comment:ulid;" json:"ulid"`
	CreatedAt   int64   `gorm:"column:created_at;autoCreateTime:milli;type:bigint;comment:创建时间;" json:"created_at"`
	UpdatedAt   int64   `gorm:"column:updated_at;autoUpdateTime:milli;type:bigint;comment:修改时间;" json:"updated_at"`
	DeletedAt   int64   `gorm:"column:deleted_at;autoDeletedTime:milli;type:bigint;comment:删除时间;" json:"deleted_at"`
	CreatedBy   string  `gorm:"column:created_by;type:varchar(32);comment:创建者;" json:"created_by"`
	UpdatedBy   string  `gorm:"column:updated_by;type:varchar(32);comment:修改者;" json:"updated_by"`
	DeletedBy   string  `gorm:"column:deleted_by;type:varchar(32);comment:删除者;" json:"deleted_by"`
	ChunkID     string  `json:"chunk_id" gorm:"column:chunk_id"`
	DocID       string  `json:"doc_id" gorm:"column:doc_id"`
	KnowledgeID string  `json:"knowledge_id" gorm:"column:knowledge_id"`
	Content     string  `json:"content" gorm:"column:content;not null"`
	Score       float64 `json:"score" gorm:"column:score"`
}

// TableName specifies the database table name for pgVectorDoc
func (PgVectorDocScore) TableName() string {
	return "doc_content"
}
