package reranker

import (
	"context"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

type Reranker interface {
	Rerank(ctx context.Context, req *schema.RerankRequest) (*schema.RerankResponse, error)
}
