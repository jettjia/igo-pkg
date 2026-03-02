package rerank

import (
	"context"
	"fmt"
	"strings"

	"github.com/jettjia/igo-pkg/aipkg/pkg/types"
)

// Reranker defines the interface for document reranking
type Reranker interface {
	// Rerank reranks documents based on relevance to the query
	Rerank(ctx context.Context, query string, documents []string) ([]RankResult, error)

	// GetModelName returns the model name
	GetModelName() string

	// GetModelID returns the model ID
	GetModelID() string
}

type RankResult struct {
	Index          int          `json:"index"`
	Document       DocumentInfo `json:"document"`
	RelevanceScore float64      `json:"relevance_score"`
}

type DocumentInfo struct {
	Text string `json:"text"`
}

type RerankerConfig struct {
	APIKey    string
	BaseURL   string
	ModelName string
	Source    types.ModelSource
	ModelID   string
}

// NewReranker creates a reranker
func NewReranker(config *RerankerConfig) (Reranker, error) {
	switch strings.ToLower(string(config.Source)) {
	case string(types.ModelSourceRemote):
		return NewOpenAIReranker(config)
	default:
		return nil, fmt.Errorf("unsupported rerank model source: %s", config.Source)
	}
}
