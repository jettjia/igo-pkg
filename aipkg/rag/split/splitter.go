package split

import (
	"context"
	"errors"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

// StrategyType 表示分块策略类型
type StrategyType string

const (
	StrategyTypeFixedSize         StrategyType = "fixed_size"
	StrategyTypeDocumentStructure StrategyType = "document_structure"
	StrategyTypeRecursiveChar     StrategyType = "recursive_character"
	StrategyTypeSemantic          StrategyType = "semantic"
)

// StrategyBase 为所有策略共享的配置
type StrategyBase struct {
	ChunkSize           int     // 单个块最大字符数（按 rune 计数）
	OverlapRatio        float64 // 相邻块重叠比例，范围建议 0.1~0.2
	RemoveURLAndEmail   bool    // 是否移除 URL 和邮箱
	NormalizeWhitespace bool    // 是否替换连续空格/换行/制表符
	TrimSpace           bool    // 是否对块内容做 TrimSpace
}

// Splitter 文本分块器接口
type Splitter interface {
	GetType() StrategyType
	Split(ctx context.Context, text string) ([]*schema.Document, error)
	Validate() error
}

func (b *StrategyBase) validateBase() error {
	if b.ChunkSize <= 0 {
		b.ChunkSize = 500
	}
	if b.OverlapRatio == 0 {
		b.OverlapRatio = 0.1
	}
	if b.OverlapRatio < 0.1 || b.OverlapRatio > 0.2 {
		return errors.New("overlap_ratio 仅允许 0.1~0.2")
	}
	return nil
}

