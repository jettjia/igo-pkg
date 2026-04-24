package schema

// Document 文档信息
type Document struct {
	ID          string         `json:"id"`           // opensearch id 或者 pgsql的表主键id
	DocID       string         `json:"doc_id"`       // 文档ID
	Content     string         `json:"content"`      // 文档内容
	Score       float64        `json:"score"`        // 相似度分数（可选）
	ChunkId     string         `json:"chunk_id"`     // 分块ID（可选）
	KnowledgeID string         `json:"knowledge_id"` // 知识库ID（可选）
	Title       string         `json:"title"`        // 层级标题（可选）
	HeadingPath []string       `json:"heading_path"` // 标题层级路径，如 ["第一章", "1.1节"]
	Embedding   []float32      `json:"embedding"`    // 嵌入向量（可选）
	Depth       int            `json:"depth"`        // 层级（可选）
	MetaData    map[string]any `json:"meta_data"`    // 元数据（可选）
	Page        int            `json:"page"`         // 页码（可选）
}
