package chunker

import (
	"context"
	"strings"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

// NewSemanticStrategy 创建新的语义分块策略实例
func NewSemanticStrategy() *SemanticStrategy {
	return &SemanticStrategy{
		StrategyBase: StrategyBase{
			ChunkSize:       500,   // 默认块大小
			Overlap:         0.1,   // 默认重叠比例
			Concurrency:     0,     // 默认不使用并发
			TrimSpace:       false, // 默认不清理空格
			TrimEllipsis:    false, // 默认不清理多余点
			TrimURLAndEmail: false, // 默认不移除URL和邮箱
		},
		Threshold: 0.3, // 默认语义相似度阈值
	}
}

// WithThreshold 设置语义相似度阈值
func (s *SemanticStrategy) WithThreshold(threshold float64) *SemanticStrategy {
	s.Threshold = threshold
	return s
}

// GetType 获取策略类型
func (s *SemanticStrategy) GetType() StrategyType {
	return StrategyTypeSemantic
}

// Chunk 执行语义分块操作
func (s *SemanticStrategy) Chunk(ctx context.Context, text string) ([]*schema.Document, error) {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 进行内容预处理
	processedText := text
	if s.TrimURLAndEmail {
		processedText = removeURLAndEmail(processedText)
	}
	if s.TrimEllipsis {
		processedText = removeEllipsis(processedText)
	}

	// 规范化转义空白字符
	processedText = normalizeEscapedWhitespace(processedText)

	// 语义分块的简化实现：首先按段落分割，然后对过长的段落进行二次切分
	paragraphs := strings.Split(processedText, "\n\n")
	docs := make([]*schema.Document, 0, len(paragraphs))

	for _, para := range paragraphs {
		paraTrim := para
		if s.TrimSpace {
			paraTrim = strings.TrimSpace(paraTrim)
		}
		if paraTrim == "" {
			continue
		}

		// 检查段落长度，如果过长则进行二次切分
		if runeLen(paraTrim) > s.ChunkSize {
			// 使用递归字符分块作为语义分块的兜底方案
			subChunks := splitBySizeWithOverlap(paraTrim, s.ChunkSize, s.Overlap)
			for _, subChunk := range subChunks {
				if subChunk != "" {
					doc := createDocumentWithPageExtraction(subChunk, "", 0)
					docs = append(docs, doc)
				}
			}
		} else {
			// 直接作为一个块
			doc := createDocumentWithPageExtraction(paraTrim, "", 0)
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

// Validate 验证策略参数有效性
func (s *SemanticStrategy) Validate() error {
	// 验证块大小
	if s.ChunkSize <= 0 {
		s.ChunkSize = 500
	}

	// 验证重叠比例
	if s.Overlap < 0 || s.Overlap > 1 {
		s.Overlap = 0.1
	}

	// 验证并发数
	if s.Concurrency < 0 {
		s.Concurrency = 0
	}

	// 验证阈值
	if s.Threshold < 0 || s.Threshold > 1 {
		s.Threshold = 0.3
	}

	return nil
}

// splitBySizeWithOverlap 按大小分割文本并添加重叠
func splitBySizeWithOverlap(text string, chunkSize int, overlapRatio float64) []string {
	if text == "" || chunkSize <= 0 {
		return nil
	}

	textRunes := []rune(text)
	textLen := len(textRunes)
	overlap := int(float64(chunkSize) * overlapRatio)
	if overlap < 0 {
		overlap = 0
	}

	var chunks []string
	for i := 0; i < textLen; i += chunkSize - overlap {
		end := i + chunkSize
		if end > textLen {
			end = textLen
		}
		chunks = append(chunks, string(textRunes[i:end]))

		// 避免最后一个块太短
		if i+chunkSize-overlap >= textLen-overlap {
			break
		}
	}

	return chunks
}
