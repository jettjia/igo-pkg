package schema

import (
	"maps"
	"slices"
)

type VectorEmbedding struct {
	Content     string    `json:"content" gorm:"column:content;not null"`     // Text content of the chunk
	DocID       string    `json:"doc_id" gorm:"column:doc_id;not null"`       // ID of the document
	ChunkID     string    `json:"chunk_id" gorm:"column:chunk_id"`            // Unique ID of the text chunk
	KnowledgeID string    `json:"knowledge_id" gorm:"column:knowledge_id"`    // ID of the knowledge item
	Embedding   []float32 `json:"embedding" gorm:"column:embedding;not null"` // Vector embedding of the content
}

// ToDBVectorEmbeddingOpensearch converts IndexInfo to Elasticsearch document format
func ToDBVectorEmbeddingOpensearch(embedding *Document, additionalParams map[string]interface{}) *VectorEmbedding {
	vector := &VectorEmbedding{
		DocID:       embedding.DocID,
		Content:     embedding.Content,
		ChunkID:     embedding.ChunkId,
		KnowledgeID: embedding.KnowledgeID,
	}
	// Add embedding data if available in additionalParams
	if additionalParams != nil && slices.Contains(slices.Collect(maps.Keys(additionalParams)), "embedding") {
		if embeddingMap, ok := additionalParams["embedding"].(map[string][]float32); ok {
			vector.Embedding = embeddingMap[embedding.DocID]
		}
	}
	return vector
}

type VectorDoc struct {
	DocID       string `json:"doc_id" gorm:"column:doc_id;not null"`    // ID of the document
	Content     string `json:"content" gorm:"column:content;not null"`  // Text content of the chunk
	KnowledgeID string `json:"knowledge_id" gorm:"column:knowledge_id"` // ID of the knowledge item
}

// ToDBVectorEmbeddingOpensearchDoc 全文
func ToDBVectorEmbeddingOpensearchDoc(embedding *Document) *VectorDoc {
	vector := &VectorDoc{
		DocID:       embedding.DocID,
		Content:     embedding.Content,
		KnowledgeID: embedding.KnowledgeID,
	}
	return vector
}
