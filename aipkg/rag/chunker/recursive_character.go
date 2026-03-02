package chunker

import (
	"context"
	"strings"

	"github.com/jettjia/go-pkg/aipkg/schema"
)

// NewRecursiveCharacterStrategy 创建新的递归字符分块策略实例
func NewRecursiveCharacterStrategy() *RecursiveCharacterStrategy {
	return &RecursiveCharacterStrategy{
		StrategyBase: StrategyBase{
			ChunkSize:       500,   // 默认块大小
			Overlap:         0.1,   // 默认重叠比例
			Concurrency:     0,     // 默认不使用并发
			TrimSpace:       false, // 默认不清理空格
			TrimEllipsis:    false, // 默认不清理多余点
			TrimURLAndEmail: false, // 默认不移除URL和邮箱
		},
		Separators: []string{"\n\n", "\n", " ", ""}, // 默认分隔符优先级
	}
}

// WithSeparators 设置自定义分隔符列表
func (s *RecursiveCharacterStrategy) WithSeparators(separators []string) *RecursiveCharacterStrategy {
	s.Separators = separators
	return s
}

// WithChunkSize 设置块大小
func (s *RecursiveCharacterStrategy) WithChunkSize(size int) *RecursiveCharacterStrategy {
	s.ChunkSize = size
	return s
}

// WithOverlap 设置重叠比例
func (s *RecursiveCharacterStrategy) WithOverlap(overlap float64) *RecursiveCharacterStrategy {
	s.Overlap = overlap
	return s
}

// WithConcurrency 设置并发数
func (s *RecursiveCharacterStrategy) WithConcurrency(concurrency int) *RecursiveCharacterStrategy {
	s.Concurrency = concurrency
	return s
}

// WithTrimSpace 设置是否清理空格
func (s *RecursiveCharacterStrategy) WithTrimSpace(trim bool) *RecursiveCharacterStrategy {
	s.TrimSpace = trim
	return s
}

// WithTrimEllipsis 设置是否清理多余点
func (s *RecursiveCharacterStrategy) WithTrimEllipsis(trim bool) *RecursiveCharacterStrategy {
	s.TrimEllipsis = trim
	return s
}

// WithTrimURLAndEmail 设置是否移除URL和邮箱
func (s *RecursiveCharacterStrategy) WithTrimURLAndEmail(trim bool) *RecursiveCharacterStrategy {
	s.TrimURLAndEmail = trim
	return s
}

// GetType 获取策略类型
func (s *RecursiveCharacterStrategy) GetType() StrategyType {
	return StrategyTypeRecursiveCharacter
}

// Chunk 执行分块操作
func (s *RecursiveCharacterStrategy) Chunk(ctx context.Context, text string) ([]*schema.Document, error) {
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

	// 获取分隔符列表
	separators := s.Separators
	if len(separators) == 0 {
		// 默认分隔符优先级：段落、行、词、字符
		separators = []string{"\n\n", "\n", " ", ""}
	}

	// 执行递归分块
	chunksText := recursiveSplitWithSeparators(processedText, separators, s.ChunkSize, s.Overlap)

	// 转换为Document切片
	docs := make([]*schema.Document, 0, len(chunksText))
	for _, ct := range chunksText {
		ctTrim := ct
		if s.TrimSpace {
			ctTrim = strings.TrimSpace(ctTrim)
		}
		if ctTrim == "" {
			continue
		}

		// 创建文档并提取页码
		doc := createDocumentWithPageExtraction(ctTrim, "", 0)
		docs = append(docs, doc)
	}

	// 添加页码传递机制，确保没有页码的切片继承最近的页码
	var lastPageNum int = 0
	for i, doc := range docs {
		pageNum := doc.Page
		if pageNum > 0 {
			lastPageNum = pageNum
		} else if lastPageNum > 0 {
			// 如果当前切片没有页码但前面有有效的页码，则继承
			docs[i].Page = lastPageNum
		}
	}

	return docs, nil
}

// Validate 验证策略参数有效性
func (s *RecursiveCharacterStrategy) Validate() error {
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

	// 验证分隔符列表
	if len(s.Separators) == 0 {
		s.Separators = []string{"\n\n", "\n", " ", ""}
	}

	return nil
}

// recursiveSplitWithSeparators 递归按分隔符列表拆分，并在每一层将过大的部分继续用下一级分隔符切分
func recursiveSplitWithSeparators(text string, separators []string, chunkSize int, overlapRatio float64) []string {
	if text == "" {
		return nil
	}
	if len(separators) == 0 {
		return splitByRunes(text, chunkSize, overlapRatio)
	}

	sep := separators[0]
	// 当当前分隔符不存在时，降级到下一级分隔符
	var parts []string
	if sep == "" {
		// 字符级别由下方分支处理；此处保持一致性走下一步
		parts = []string{text}
	} else {
		parts = strings.Split(text, sep)
	}

	if len(parts) == 1 {
		// 当前分隔符未命中，使用更低级分隔符
		return recursiveSplitWithSeparators(text, separators[1:], chunkSize, overlapRatio)
	}

	// 对每个部分递归细分，确保都不超过 chunkSize
	refined := make([]string, 0, len(parts))
	for _, p := range parts {
		pTrim := p
		if strings.TrimSpace(pTrim) == "" {
			continue
		}
		if runeLen(pTrim) > chunkSize {
			// 继续用更低级分隔符切分（包括最终字符级）
			sub := recursiveSplitWithSeparators(pTrim, separators[1:], chunkSize, overlapRatio)
			refined = append(refined, sub...)
		} else {
			refined = append(refined, pTrim)
		}
	}

	// 合并为最终块（带重叠），使用当前分隔符重新连接
	return mergePartsWithOverlap(refined, sep, chunkSize, overlapRatio)
}

// splitByRunes 直接按字符切分并合并，作为分隔符递归的最后兜底
func splitByRunes(text string, chunkSize int, overlapRatio float64) []string {
	if text == "" || chunkSize <= 0 {
		return nil
	}
	// 将每个字符视作一个元素，复用合并逻辑
	runes := []rune(text)
	parts := make([]string, len(runes))
	for i, r := range runes {
		parts[i] = string(r)
	}
	return mergePartsWithOverlap(parts, "", chunkSize, overlapRatio)
}

// mergePartsWithOverlap 将若干部分按 chunkSize 合并为块，并按 overlapRatio 做重叠回退
// 当 sep 非空时，在拼接相邻部分时恢复该分隔符
func mergePartsWithOverlap(parts []string, sep string, chunkSize int, overlapRatio float64) []string {
	if len(parts) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		return []string{strings.Join(parts, sep)}
	}

	overlapTokens := calculateOverlap(chunkSize, overlapRatio)
	var chunks []string

	i := 0
	for i < len(parts) {
		// 组装一个块
		var builder []string
		currentSize := 0
		j := i
		for j < len(parts) {
			part := parts[j]
			addSize := runeLen(part)
			if len(builder) > 0 && sep != "" {
				addSize += runeLen(sep)
			}
			// 如果添加当前part会超过chunkSize，则停止添加
			if currentSize+addSize > chunkSize && currentSize > 0 {
				break
			}
			builder = append(builder, part)
			currentSize += addSize
			j++
		}

		// 处理单个part过大的情况
		if len(builder) == 1 && runeLen(builder[0]) > chunkSize {
			// 对于过大的单个part，直接按字符分割
			chunkRunes := []rune(builder[0])
			if len(chunkRunes) > chunkSize {
				// 只取chunkSize大小的字符
				chunks = append(chunks, string(chunkRunes[:chunkSize]))
				// 推进索引，确保不重复处理
				i++
				continue
			}
		}

		chunk := strings.Join(builder, sep)
		chunks = append(chunks, chunk)

		// 计算重叠并推进索引
		if overlapTokens > 0 && i+len(builder) < len(parts) {
			// 从当前builder的尾部开始，计算需要重叠的parts数量
			backLen := 0
			backParts := 0
			// 反向遍历builder中的parts，累计长度直到达到overlapTokens
			for k := len(builder) - 1; k >= 0 && backLen < overlapTokens; k-- {
				partLen := runeLen(builder[k])
				if k < len(builder)-1 && sep != "" {
					partLen += runeLen(sep)
				}
				backLen += partLen
				backParts++
			}
			// 调整i到适当位置以实现重叠
			i += len(builder) - backParts
		} else {
			// 无重叠情况下直接推进索引
			i += len(builder)
		}
	}

	return chunks
}

// runeLen 获取字符串的rune长度（字符数）
func runeLen(s string) int {
	return len([]rune(s))
}

// createDocumentWithPageExtraction 创建文档并提取页码
func createDocumentWithPageExtraction(text string, source string, defaultPage int) *schema.Document {
	// 提取页码
	pageNum := extractPageNumber(text)
	if pageNum == 0 {
		pageNum = defaultPage
	}

	// 移除页码标记
	content := removePageMarkers(text)

	// 创建文档
	return &schema.Document{
		Content: content,
		Page:    pageNum,
	}
}
