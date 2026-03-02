package docpipeline

import (
	"context"

	"github.com/jettjia/go-pkg/aipkg/schema"
)

// DocumentProcessor 定义文档处理流水线接口
type DocumentProcessor interface {
	// Process 处理文本，返回带有嵌入向量的文档
	Process(ctx context.Context, text string, options ProcessOptions) ([]*schema.Document, error)

	// BatchProcess 批量处理文本
	BatchProcess(ctx context.Context, texts []string, options ProcessOptions) ([][]*schema.Document, error)

	// 只切片，不向量化
	SliceOnly(ctx context.Context, texts string, options ProcessOptions) ([]*schema.Document, error)
}

// ProcessOptions 处理选项
type ProcessOptions struct {
	ChunkStrategy     any  // chunk策略配置
	EmbedderConfig    any  // 嵌入模型配置
	EnableConcurrency bool // 是否启用并发处理
	MaxConcurrency    int  // 最大并发数
}
