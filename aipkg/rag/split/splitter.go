package split

import (
	"context"
	"errors"
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
	RemoveImageURL      bool    // 是否移除 Markdown 图片 URL，如 ![](xxx.jpg)
	NormalizeWhitespace bool    // 是否替换连续空格/换行/制表符
	TrimSpace           bool    // 是否对块内容做 TrimSpace

	// 内部缓存，用于处理表格等特殊内容
	tableCache []string
	// 页码缓存，存储原始文本中的页码信息（用于切分后推断每个chunk的页码）
	pageCache []string
}

// SliceContent 切片内容结构
type SliceContent struct {
	Title   string `json:"title"`   // 标题内容（已合并到 Text，保留字段）
	Text    string `json:"text"`   // 切片内容，合并了 title + text + table
	Table   string `json:"table"`   // 表格内容（已合并到 Text，保留字段）
	Picture string `json:"picture"` // 图片保存路径
}

// Chunk 表示一个文档切片
type Chunk struct {
	DocName      string       `json:"doc_name"`    // 传入的文件名称
	DocMD5       string       `json:"doc_md5"`     // 文本的md5值
	SliceMD5     string       `json:"slice_md5"`    // slice的md5值
	ID           string       `json:"id"`           // 切片的id值
	Pages        []int       `json:"pages"`        // 当前slice所在页码
	SegmentID    int         `json:"segment_id"`   // 当前切片的序号,按照人的阅读顺序它是第几个block
	SuperiorID    string      `json:"superior_id"`  // 父亲slice的id值
	HeadingPath  []string     `json:"heading_path"` // 标题层级路径，如 ["第一章", "1.1节"]
	SliceContent SliceContent `json:"slice_content"`
}

// Splitter 文本分块器接口
type Splitter interface {
	GetType() StrategyType
	Split(ctx context.Context, text string, fileName string) ([]*Chunk, error)
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
