package chunker

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jettjia/go-pkg/aipkg/schema"
)

// NewRecursiveParagraphStrategy 创建新的递归段落分块策略实例
func NewRecursiveParagraphStrategy() *RecursiveParagraphStrategy {
	return &RecursiveParagraphStrategy{
		StrategyBase: StrategyBase{
			ChunkSize:       500,   // 默认块大小
			Overlap:         0.1,   // 默认重叠比例
			Concurrency:     0,     // 默认不使用并发
			TrimSpace:       false, // 默认不清理空格
			TrimEllipsis:    false, // 默认不清理多余点
			TrimURLAndEmail: false, // 默认不移除URL和邮箱
		},
		Separator: "\n\n", // 默认段落分隔符
		MaxDepth:  5,      // 默认不分层
		SaveTitle: true,   // 默认保留标题
	}
}

// WithSeparator 设置段落分隔符
func (s *RecursiveParagraphStrategy) WithSeparator(separator string) *RecursiveParagraphStrategy {
	s.Separator = separator
	return s
}

// WithMaxDepth 设置最大分层深度
func (s *RecursiveParagraphStrategy) WithMaxDepth(depth int) *RecursiveParagraphStrategy {
	s.MaxDepth = depth
	return s
}

// WithSaveTitle 设置是否保留标题
func (s *RecursiveParagraphStrategy) WithSaveTitle(save bool) *RecursiveParagraphStrategy {
	s.SaveTitle = save
	return s
}

// GetType 获取策略类型
func (s *RecursiveParagraphStrategy) GetType() StrategyType {
	return StrategyTypeRecursiveParagraph
}

// Chunk 执行递归段落分块操作
func (s *RecursiveParagraphStrategy) Chunk(ctx context.Context, text string) ([]*schema.Document, error) {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 直接使用原始文本，不进行预处理
	processedText := text

	// 如果设置了分层，使用分层分块
	if s.MaxDepth > 0 {
		chunks, err := s.chunkWithHierarchy(processedText)
		if err != nil {
			return nil, err
		}
		// 如果分层处理得到了有效结果，直接返回
		if len(chunks) > 0 {
			return chunks, nil
		}
	}

	// 否则使用普通段落分块
	return s.chunkFlat(processedText), nil
}

// Validate 验证策略参数有效性
func (s *RecursiveParagraphStrategy) Validate() error {
	// 验证块大小
	if s.ChunkSize <= 0 {
		s.ChunkSize = 500 // 设置默认值
	}

	// 验证重叠比例
	if s.Overlap < 0 || s.Overlap > 1 {
		s.Overlap = 0.1 // 设置默认值
	}

	// 验证并发数
	if s.Concurrency < 0 {
		s.Concurrency = 0 // 设置默认值
	}

	// 验证最大深度
	if s.MaxDepth < 0 {
		s.MaxDepth = 0 // 设置默认值
	}

	// 验证分隔符
	if s.Separator == "" {
		s.Separator = "\n\n" // 设置默认分隔符
	}

	return nil
}

// chunkFlat 普通段落分块实现
func (s *RecursiveParagraphStrategy) chunkFlat(text string) []*schema.Document {
	// 统一处理换行符
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\\n", "\n")

	// 按段落分割（空行分隔）
	lines := strings.Split(text, "\n")
	var paragraphs []string
	var currentParagraph []string
	var emptyLineCount int

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			emptyLineCount++
			// 遇到空行，且当前段落不为空，则结束当前段落
			if len(currentParagraph) > 0 && emptyLineCount >= 2 {
				paragraph := strings.Join(currentParagraph, "\n")
				paragraphs = append(paragraphs, paragraph)
				currentParagraph = []string{}
				emptyLineCount = 0 // 重置空行计数
			}
		} else {
			emptyLineCount = 0 // 重置空行计数
			// 将非空行添加到当前段落
			currentParagraph = append(currentParagraph, line)
		}
	}

	// 添加最后一个段落（如果有）
	if len(currentParagraph) > 0 {
		paragraph := strings.Join(currentParagraph, "\n")
		paragraphs = append(paragraphs, paragraph)
	}

	// 检查段落数量和总长度，决定是否使用句子分割
	totalSize := len([]rune(text))
	useSentenceSplit := len(paragraphs) <= 5 || totalSize > int(float64(s.ChunkSize))

	var chunks []*schema.Document

	if useSentenceSplit {
		// 使用splitLargeSection进行进一步分割
		chunks = s.splitLargeSection(text, "", 0)
	} else {
		// 直接使用段落分割
		chunks = make([]*schema.Document, 0, len(paragraphs))
		for _, para := range paragraphs {
			paraTrim := strings.TrimSpace(para)
			if paraTrim == "" {
				continue
			}

			// 检查段落大小，如果过大则使用splitLargeSection进一步分割
			contentSize := len([]rune(paraTrim))
			if contentSize > int(float64(s.ChunkSize)) {
				subChunks := s.splitLargeSection(paraTrim, "", 0)
				chunks = append(chunks, subChunks...)
			} else {
				doc := createDocumentWithPageExtractionParag(paraTrim, "", 0)
				chunks = append(chunks, doc)
			}
		}
	}

	return chunks
}

// createDocumentWithPageExtractionParag 创建文档并提取页码信息
func createDocumentWithPageExtractionParag(text string, title string, depth int) *schema.Document {
	// 提取页码
	pageNum := extractPageNumber(text)

	// 移除页码标记
	cleanContent := removePageMarkers(text)

	// 创建文档
	doc := &schema.Document{
		ID:      uuid.New().String(),
		Content: cleanContent,
		Title:   title,
		Depth:   depth,
		Page:    pageNum,
	}

	return doc
}

// chunkWithHierarchy 分层段落分块实现
func (s *RecursiveParagraphStrategy) chunkWithHierarchy(text string) ([]*schema.Document, error) {
	// 统一处理换行符
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\\n", "\n")

	lines := strings.Split(text, "\n")
	headingRegexp := regexp.MustCompile(`^(#+)\s+(.*)$`)
	var chunks []*schema.Document
	var lastTitle string
	var lastDepth int
	var contentBuf []string
	hasHeading := false

	for _, line := range lines {
		matches := headingRegexp.FindStringSubmatch(line)
		if matches != nil {
			hasHeading = true
			// 新标题，先结算上一个 chunk
			if lastTitle != "" || len(contentBuf) > 0 {
				content := strings.Join(contentBuf, "\n")
				content = strings.TrimSpace(content)
				if content != "" {
					// 检查内容是否过大，如果过大则使用splitLargeSection进一步分割
					contentSize := len([]rune(content))
					if contentSize > int(float64(s.ChunkSize)) {
						subChunks := s.splitLargeSection(content, lastTitle, lastDepth)
						chunks = append(chunks, subChunks...)
					} else {
						var docTitle string
						var docDepth int
						if s.SaveTitle {
							docTitle = lastTitle
							docDepth = lastDepth
						}
						doc := createDocumentWithPageExtractionParag(content, docTitle, docDepth)
						chunks = append(chunks, doc)
					}
				}
			}
			// 清洗标题，去除无用的 *、#、空格
			cleanTitle := strings.TrimSpace(matches[2])
			cleanTitle = strings.Trim(cleanTitle, "*# ")
			lastTitle = cleanTitle
			lastDepth = len(matches[1])
			contentBuf = []string{}
		} else {
			contentBuf = append(contentBuf, line)
		}
	}

	// 收尾
	if lastTitle != "" || len(contentBuf) > 0 {
		content := strings.Join(contentBuf, "\n")
		content = strings.TrimSpace(content)
		if content != "" {
			// 检查内容是否过大，如果过大则使用splitLargeSection进一步分割
			contentSize := len([]rune(content))
			if contentSize > int(float64(s.ChunkSize)) {
				subChunks := s.splitLargeSection(content, lastTitle, lastDepth)
				chunks = append(chunks, subChunks...)
			} else {
				var docTitle string
				var docDepth int
				if s.SaveTitle {
					docTitle = lastTitle
					docDepth = lastDepth
				}
				doc := createDocumentWithPageExtractionParag(content, docTitle, docDepth)
				chunks = append(chunks, doc)
			}
		}
	}

	// 如果没有检测到任何 Markdown 标题，则自动按段落和句子分割
	if !hasHeading {
		// 直接用 splitLargeSection 进行分段和句子切分
		return s.splitLargeSection(text, "", 0), nil
	}
	return chunks, nil
}

// isTitleLine 判断一行文本是否为标题行
// 返回标题深度（1-6）和清理后的标题文本
func (s *RecursiveParagraphStrategy) isTitleLine(line string) (int, string) {
	// 检查Markdown标题格式
	trimmed := strings.TrimSpace(line)
	if len(trimmed) >= 2 {
		// 检查是否以#开头
		if trimmed[0] == '#' {
			// 计算#的数量作为深度
			depth := 0
			for i := 0; i < len(trimmed) && i < 6; i++ {
				if trimmed[i] == '#' {
					depth++
				} else {
					break
				}
			}
			// 确保#后面有空格
			if depth > 0 && len(trimmed) > depth && trimmed[depth] == ' ' {
				title := strings.TrimSpace(trimmed[depth:])
				return depth, title
			}
		}
	}

	// 检查是否为其他标题格式（如等号/减号样式）
	// 简单实现，实际项目中可以根据需求扩展

	return 0, ""
}

// splitLargeSection 分割大的section
func (s *RecursiveParagraphStrategy) splitLargeSection(content, title string, depth int) []*schema.Document {
	// 添加过滤连续点的处理
	if s.TrimEllipsis {
		content = removeEllipsis(content)
	}

	var chunks []*schema.Document

	// 按句子分割（支持中英文标点符号）
	var sentences []string
	currentSentence := ""
	inCodeBlock := false
	codeBlockCount := 0
	codeBlockType := ""

	for i, r := range content {
		// 检查是否在代码块内 (```)
		if r == '`' {
			codeBlockCount++
			if codeBlockCount == 3 {
				inCodeBlock = !inCodeBlock
				if inCodeBlock {
					// 记录代码块类型（如果有）
					if i+1 < len(content) {
						nextChar := rune(content[i+1])
						if nextChar != '\n' && nextChar != '\r' {
							// 尝试获取代码块类型
							j := i + 1
							for ; j < len(content) && rune(content[j]) != '\n' && rune(content[j]) != '\r'; j++ {
							}
							codeBlockType = content[i+1 : j]
							_ = strings.TrimSpace(codeBlockType)
						}
					}
				}
				codeBlockCount = 0
			}
		} else {
			codeBlockCount = 0
		}

		// 如果不在代码块内，检查句子结束符
		if !inCodeBlock && (r == '.' || r == '?' || r == '!' || r == '。' || r == '？' || r == '！' || r == ';' || r == '；') {
			currentSentence += string(r)
			if strings.TrimSpace(currentSentence) != "" {
				sentences = append(sentences, currentSentence)
				currentSentence = ""
			}
		} else {
			currentSentence += string(r)
		}
	}

	// 添加最后一个句子（如果有）
	if strings.TrimSpace(currentSentence) != "" {
		sentences = append(sentences, currentSentence)
	}

	var currentSentences []string
	currentSize := 0

	for _, sentence := range sentences {
		sentenceSize := len([]rune(sentence))

		// 检查是否超过分割阈值
		if currentSize+sentenceSize > int(float64(s.ChunkSize)) {
			if len(currentSentences) > 0 {
				chunkContent := strings.Join(currentSentences, "")
				// 添加适当的标点符号
				if len(chunkContent) > 0 && !strings.HasSuffix(chunkContent, ".") && !strings.HasSuffix(chunkContent, "。") && !strings.HasSuffix(chunkContent, "?") && !strings.HasSuffix(chunkContent, "？") && !strings.HasSuffix(chunkContent, "!") && !strings.HasSuffix(chunkContent, "！") && !strings.HasSuffix(chunkContent, ";") && !strings.HasSuffix(chunkContent, "；") {
					chunkContent += "。"
				}
				if s.TrimSpace {
					chunkContent = strings.TrimSpace(chunkContent)
				}
				// 添加切片
				var docTitle string
				var docDepth int
				if s.SaveTitle {
					docTitle = title
					docDepth = depth
				}
				doc := createDocumentWithPageExtractionParag(chunkContent, docTitle, docDepth)
				chunks = append(chunks, doc)
			}
			// 重置当前句子集合，添加当前句子
			currentSentences = []string{sentence}
			currentSize = sentenceSize
		} else {
			currentSentences = append(currentSentences, sentence)
			currentSize += sentenceSize
		}
	}

	// 处理最后一组句子
	if len(currentSentences) > 0 {
		chunkContent := strings.Join(currentSentences, "")
		// 添加适当的标点符号
		if len(chunkContent) > 0 && !strings.HasSuffix(chunkContent, ".") && !strings.HasSuffix(chunkContent, "。") && !strings.HasSuffix(chunkContent, "?") && !strings.HasSuffix(chunkContent, "？") && !strings.HasSuffix(chunkContent, "!") && !strings.HasSuffix(chunkContent, "！") && !strings.HasSuffix(chunkContent, ";") && !strings.HasSuffix(chunkContent, "；") {
			chunkContent += "。"
		}
		if s.TrimSpace {
			chunkContent = strings.TrimSpace(chunkContent)
		}

		// 检查最后一个切片大小，如果超过阈值则进行二次分割
		if len([]rune(chunkContent)) > int(float64(s.ChunkSize)*0.5) {
			// 改进的二次分割逻辑
			var tempSentences []string
			tempSize := 0
			for _, sent := range currentSentences {
				sentSize := len([]rune(sent))
				if tempSize+sentSize > int(float64(s.ChunkSize)*0.5) {
					if len(tempSentences) > 0 {
						tempContent := strings.Join(tempSentences, "")
						if !strings.HasSuffix(tempContent, ".") && !strings.HasSuffix(tempContent, "。") && !strings.HasSuffix(tempContent, "?") && !strings.HasSuffix(tempContent, "？") && !strings.HasSuffix(tempContent, "!") && !strings.HasSuffix(tempContent, "！") && !strings.HasSuffix(tempContent, ";") && !strings.HasSuffix(tempContent, "；") {
							tempContent += "。"
						}
						var docTitle string
						var docDepth int
						if s.SaveTitle {
							docTitle = title
							docDepth = depth
						}
						doc := createDocumentWithPageExtractionParag(tempContent, docTitle, docDepth)
						chunks = append(chunks, doc)
						tempSentences = []string{sent}
						tempSize = sentSize
					} else {
						// 单个句子超过阈值
						if s.TrimSpace {
							sent = strings.TrimSpace(sent)
						}
						var docTitle string
						var docDepth int
						if s.SaveTitle {
							docTitle = title
							docDepth = depth
						}
						doc := createDocumentWithPageExtractionParag(sent, docTitle, docDepth)
						chunks = append(chunks, doc)
						tempSentences = []string{}
						tempSize = 0
					}
				} else {
					tempSentences = append(tempSentences, sent)
					tempSize += sentSize
				}
			}

			// 处理最后一组临时句子
			if len(tempSentences) > 0 {
				tempContent := strings.Join(tempSentences, "")
				if !strings.HasSuffix(tempContent, ".") && !strings.HasSuffix(tempContent, "。") && !strings.HasSuffix(tempContent, "?") && !strings.HasSuffix(tempContent, "？") && !strings.HasSuffix(tempContent, "!") && !strings.HasSuffix(tempContent, "！") && !strings.HasSuffix(tempContent, ";") && !strings.HasSuffix(tempContent, "；") {
					tempContent += "。"
				}
				if s.TrimSpace {
					tempContent = strings.TrimSpace(tempContent)
				}
				var docTitle string
				var docDepth int
				if s.SaveTitle {
					docTitle = title
					docDepth = depth
				}
				doc := createDocumentWithPageExtractionParag(tempContent, docTitle, docDepth)
				chunks = append(chunks, doc)
			}
		} else {
			// 直接添加最后一个切片
			var docTitle string
			var docDepth int
			if s.SaveTitle {
				docTitle = title
				docDepth = depth
			}
			doc := createDocumentWithPageExtractionParag(chunkContent, docTitle, docDepth)
			chunks = append(chunks, doc)
		}
	}

	return chunks
}
