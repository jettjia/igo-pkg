package chunker

import (
	"context"

	"github.com/jettjia/go-pkg/aipkg/schema"
)

// ChunkingStrategy 分块策略接口
type ChunkingStrategy interface {
	// GetType 获取策略类型
	GetType() StrategyType

	// Chunk 执行分块操作
	Chunk(ctx context.Context, text string) ([]*schema.Document, error)

	// Validate 验证策略参数有效性
	Validate() error
}

// StrategyBase 分块策略基类
// 包含所有策略共享的基础参数
type StrategyBase struct {
	ChunkSize       int     // 切片最大长度（token数，按字符数近似）
	Overlap         float64 // 相邻切片的重叠比例（0-1）
	Concurrency     int     // 并发数，0表示不使用并发
	TrimSpace       bool    // 是否清理多余空格
	TrimEllipsis    bool    // 是否清理多余.
	TrimURLAndEmail bool    // 是否移除URL和邮箱
}

// RecursiveCharacterStrategy 递归字符分块策略
type RecursiveCharacterStrategy struct {
	StrategyBase
	Separators []string // 递归分块的分隔符列表
}

// RecursiveParagraphStrategy 递归段落分块策略
type RecursiveParagraphStrategy struct {
	StrategyBase
	Separator string // 文本分割的分隔符
	MaxDepth  int    // 分层切片的最大层级（0为不分层）
	SaveTitle bool   // 是否保留层级标题
}

// SemanticStrategy 语义分块策略
type SemanticStrategy struct {
	StrategyBase
	// 语义分块特有的参数可以在这里添加
	Threshold float64 // 语义相似度阈值
}

// Chunker 分块器接口
type Chunker interface {
	// Chunk 使用指定策略执行分块
	Chunk(ctx context.Context, text string, strategy ChunkingStrategy) ([]*schema.Document, error)
}
