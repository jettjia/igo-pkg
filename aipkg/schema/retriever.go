package schema

// RetrieverType represents the type of retriever
type RetrieverEngineType string

// RetrieverEngineType constants
const (
	PostgresRetrieverEngineType      RetrieverEngineType = "postgres"
	ElasticsearchRetrieverEngineType RetrieverEngineType = "opensearch"
)

// RetrieverType represents the type of retriever
type RetrieverType string

// RetrieverType constants
const (
	KeywordsRetrieverType  RetrieverType = "keywords"  // Keywords retriever
	VectorRetrieverType    RetrieverType = "vector"    // Vector retriever
	WebSearchRetrieverType RetrieverType = "websearch" // Web search retriever
	HybridRetrieverType    RetrieverType = "hybrid"    // Hybrid retriever
)

// RetrieveParams 检索参数
type RetrieveParams struct {
	// Query text
	Query string
	// Query embedding (used for vector retrieval)
	Embedding []float32
	// Excluded knowledge IDs
	ExcludeKnowledgeIDs []string
	// Excluded chunk IDs
	ExcludeChunkIDs []string
	// Number of results to return
	TopK int64
	// Similarity threshold
	Threshold float64
	// Additional parameters, different retrievers may require different parameters
	AdditionalParams map[string]interface{}
	// Retriever type
	RetrieverType RetrieverType // Retriever type
}

// RetrieveResult 检索结果
// 可扩展更多字段
type RetrieveResult struct {
	Results             []*Document
	RetrieverEngineType RetrieverEngineType
	RetrieverType       RetrieverType
	Error               error
}
