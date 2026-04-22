package split

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

var (
	urlRegex      = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex    = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	imageURLRegex = regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	tableRegex    = regexp.MustCompile(`(?s)<table>.*?</table>`)
	pageRegex     = regexp.MustCompile(`(?i)<!--\s*Page:\s*(\d+)\s*-->`)

	// 用于识别预处理阶段生成的简短占位符
	tablePlaceholderRegex = regexp.MustCompile(`HTML_TABLE_PLACEHOLDER_(\d+)`)

	spaceOnlyRegex = regexp.MustCompile(` {2,}`)
	newline3Regex  = regexp.MustCompile(`\n{3,}`)
)

func runeLen(s string) int {
	return len([]rune(s))
}

func overlapTokens(chunkSize int, overlapRatio float64) int {
	if chunkSize <= 0 {
		return 0
	}
	if overlapRatio <= 0 {
		return 0
	}
	return int(math.Round(float64(chunkSize) * overlapRatio))
}

func preProcessText(text string, base *StrategyBase) string {
	processed := text

	// 提前将表格内容替换为简短占位符，防止在 Split 过程中被切断
	base.tableCache = nil
	processed = tableRegex.ReplaceAllStringFunc(processed, func(tableHTML string) string {
		placeholder := fmt.Sprintf("HTML_TABLE_PLACEHOLDER_%d", len(base.tableCache))
		base.tableCache = append(base.tableCache, tableHTML)
		return "\n\n" + placeholder + "\n\n"
	})

	if base.RemoveURLAndEmail {
		processed = urlRegex.ReplaceAllString(processed, "")
		processed = emailRegex.ReplaceAllString(processed, "")
	}

	if base.RemoveImageURL {
		processed = imageURLRegex.ReplaceAllString(processed, "")
	}

	processed = strings.ReplaceAll(processed, "\r\n", "\n")

	if base.NormalizeWhitespace {
		processed = strings.ReplaceAll(processed, "\t", " ")
		processed = spaceOnlyRegex.ReplaceAllString(processed, " ")
		processed = newline3Regex.ReplaceAllString(processed, "\n\n")
	}

	return processed
}

func getNodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var f func(node *html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return b.String()
}

func decodeHexEscapes(s string) string {
	// Handle \uXXXX unicode escapes that appear in the raw text
	if !strings.Contains(s, "\\u") {
		return s
	}
	// Regex to find \uXXXX patterns
	re := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		hex := match[2:]
		if v, err := strconv.ParseUint(hex, 16, 32); err == nil {
			return string(rune(v))
		}
		return match
	})
}

func tableToMarkdown(tableHTML string) string {
	doc, err := html.Parse(strings.NewReader(tableHTML))
	if err != nil {
		return ""
	}

	// Collect all rows with their cell info (including rowspan/colspan)
	type cellInfo struct {
		text    string
		rowspan int
		colspan int
	}
	var allRows [][]cellInfo

	var processRow func(*html.Node)
	processRow = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var cells []cellInfo
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "th" || c.Data == "td") {
					rowspan := 1
					colspan := 1
					for _, attr := range c.Attr {
						if attr.Key == "rowspan" {
							if v, err := strconv.Atoi(attr.Val); err == nil {
								rowspan = v
							}
						}
						if attr.Key == "colspan" {
							if v, err := strconv.Atoi(attr.Val); err == nil {
								colspan = v
							}
						}
					}
					cells = append(cells, cellInfo{
						text:    strings.TrimSpace(getNodeText(c)),
						rowspan: rowspan,
						colspan: colspan,
					})
				}
			}
			if len(cells) > 0 {
				allRows = append(allRows, cells)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processRow(c)
		}
	}
	processRow(doc)

	if len(allRows) == 0 {
		return ""
	}

	// Determine max columns
	maxCols := 0
	for _, row := range allRows {
		colCount := 0
		for _, cell := range row {
			colCount += cell.colspan
		}
		if colCount > maxCols {
			maxCols = colCount
		}
	}

	if maxCols == 0 {
		return ""
	}

	// Build 2D grid with colspan/rowspan handling
	grid := make([][]string, len(allRows))
	for i := range grid {
		grid[i] = make([]string, maxCols)
	}

	// Use rowspan grid to track which cells are already filled
	rowspanGrid := make([][]int, len(allRows))
	for i := range rowspanGrid {
		rowspanGrid[i] = make([]int, maxCols)
	}

	// Fill the grid row by row, handling colspan and rowspan
	for rowIdx, row := range allRows {
		colIdx := 0
		for _, cell := range row {
			// Skip columns already filled by rowspan from previous rows
			for colIdx < maxCols && rowspanGrid[rowIdx][colIdx] > 0 {
				colIdx++
			}
			if colIdx >= maxCols {
				break
			}
			// For colspan > 1, only fill the first column with text, leave others empty
			for c := 0; c < cell.colspan && colIdx+c < maxCols; c++ {
				if c == 0 {
					grid[rowIdx][colIdx+c] = cell.text
				} else {
					grid[rowIdx][colIdx+c] = "" // Empty placeholder for colspan
				}
			}
			// Mark rowspan for subsequent rows
			for r := 1; r < cell.rowspan && rowIdx+r < len(allRows); r++ {
				for c := 0; c < cell.colspan && colIdx+c < maxCols; c++ {
					rowspanGrid[rowIdx+r][colIdx+c]++
				}
			}
			colIdx += cell.colspan
		}
	}

	// First row is header, rest are data rows
	headers := grid[0]
	dataRows := grid[1:]

	if len(headers) == 0 && len(dataRows) == 0 {
		return ""
	}

	// Build markdown
	var b strings.Builder

	// Write header row
	for i, h := range headers {
		if i > 0 {
			b.WriteString("|")
		}
		b.WriteString(decodeHexEscapes(h))
	}
	b.WriteString("\n")

	// Write separator row
	for i := range headers {
		if i > 0 {
			b.WriteString("|")
		}
		b.WriteString("---")
	}
	b.WriteString("\n")

	// Write data rows
	for _, row := range dataRows {
		// Skip rows that are completely empty (due to rowspan filling)
		allEmpty := true
		for _, cell := range row {
			if cell != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}
		for i, cell := range row {
			if i > 0 {
				b.WriteString("|")
			}
			b.WriteString(decodeHexEscapes(cell))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func applyTrimSpaceIfNeeded(text string, base *StrategyBase) string {
	if base.TrimSpace {
		return strings.TrimSpace(text)
	}
	return text
}

func recoverBrokenTable(fragment string, fullText string) string {
	// 清理 fragment 中可能破坏正则的特殊字符
	// 尝试找到 fragment 前后的 <table> 标签位置
	// 这里使用简单策略：在原文中找包含该片段的最完整的 <table>...</table>
	matches := tableRegex.FindAllString(fullText, -1)
	for _, m := range matches {
		// 检查这个 HTML 表格是否包含我们被切碎的片段
		// 去除空白符后进行模糊匹配
		cleanM := strings.Join(strings.Fields(m), "")
		cleanF := strings.Join(strings.Fields(fragment), "")
		if strings.Contains(cleanM, cleanF) {
			return m
		}
	}
	return ""
}

func newDocument(content string, title string, depth int) *schema.Document {
	doc := &schema.Document{
		Content: content,
	}
	if title != "" {
		doc.Title = title
	}
	if depth > 0 {
		doc.Depth = depth
	}
	return doc
}

func applyOverlapToStrings(chunks []string, chunkSize int, overlapRatio float64) []string {
	if len(chunks) <= 1 {
		return chunks
	}
	overlap := overlapTokens(chunkSize, overlapRatio)
	if overlap <= 0 {
		return chunks
	}

	out := make([]string, 0, len(chunks))
	out = append(out, chunks[0])
	for i := 1; i < len(chunks); i++ {
		prev := []rune(out[i-1])
		next := chunks[i]

		prefix := ""
		if len(prev) > 0 {
			start := len(prev) - overlap
			if start < 0 {
				start = 0
			}
			prefix = string(prev[start:])
		}

		merged := prefix + next
		if runeLen(merged) > chunkSize && chunkSize > 0 {
			mergedRunes := []rune(merged)
			merged = string(mergedRunes[len(mergedRunes)-chunkSize:])
		}
		out = append(out, merged)
	}
	return out
}

func calculateMD5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func convertToChunks(docs []*schema.Document, fileName string, originalText string, base *StrategyBase) []*Chunk {
	docMD5 := calculateMD5(originalText)
	chunks := make([]*Chunk, 0, len(docs))

	// 先获取全文的所有页码位置，以便后续判断每个切片所属页码
	type pageInfo struct {
		pageNo int
		pos    int
	}
	var pages []pageInfo
	matches := pageRegex.FindAllStringSubmatchIndex(originalText, -1)
	for _, m := range matches {
		pNo, _ := strconv.Atoi(originalText[m[2]:m[3]])
		pages = append(pages, pageInfo{pageNo: pNo, pos: m[0]})
	}

	for i, doc := range docs {
		sliceMD5 := calculateMD5(doc.Content)
		id := fmt.Sprintf("%s-%d", sliceMD5[:8], i)

		// 提取当前切片内容中的页码标识
		pageMap := make(map[int]bool)
		// 1. 检查文档自带的 Page (如果已由其他逻辑设置)
		if doc.Page > 0 {
			pageMap[doc.Page] = true
		}

		// 2. 扫描内容中的页码标识
		contentMatches := pageRegex.FindAllStringSubmatch(doc.Content, -1)
		for _, cm := range contentMatches {
			if p, err := strconv.Atoi(cm[1]); err == nil {
				pageMap[p] = true
			}
		}

		// 3. 如果内容中没有页码标识，尝试根据它在原图中的大致位置推断页码
		if len(pageMap) == 0 && len(pages) > 0 {
			// 找到切片内容在原图中的起始位置
			// 为了提高匹配准确度，取切片的前 100 个字符进行搜索（避开 overlap 带来的重复匹配）
			searchText := doc.Content
			if runeLen(searchText) > 100 {
				searchText = string([]rune(searchText)[:100])
			}
			// 在 originalText 中从上一个 chunk 结束的位置开始搜索，或者全局搜索
			startIdx := strings.Index(originalText, searchText)
			if startIdx != -1 {
				currentPage := 1 // 默认第一页
				for _, p := range pages {
					if startIdx >= p.pos {
						currentPage = p.pageNo
					} else {
						break
					}
				}
				pageMap[currentPage] = true
			} else {
				// 如果 Index 没找到，可能是因为 preprocess 处理了文本。
				// 暂时简单处理：沿用上一个 chunk 的页码
				if i > 0 && len(chunks) > 0 {
					for _, p := range chunks[i-1].Pages {
						pageMap[p] = true
					}
				} else {
					pageMap[1] = true
				}
			}
		}

		// 4. 补充逻辑：如果是因为 overlap 导致的切片包含了下一页的开始，也记录下来
		if len(pages) > 0 {
			startIdx := strings.Index(originalText, doc.Content)
			if startIdx != -1 {
				endIdx := startIdx + len(doc.Content)
				for _, p := range pages {
					if p.pos > startIdx && p.pos < endIdx {
						pageMap[p.pageNo] = true
					}
				}
			}
		}

		finalPages := make([]int, 0, len(pageMap))
		for p := range pageMap {
			finalPages = append(finalPages, p)
		}
		// 排序页码
		sort.Ints(finalPages)

		// 解析表格占位符（可能包含多个表格）
		var tableMD string
		content := doc.Content
		if tablePlaceholderRegex.MatchString(content) {
			// 找到所有占位符的索引
			matches := tablePlaceholderRegex.FindAllStringSubmatch(content, -1)
			var tables []string
			for _, match := range matches {
				if len(match) > 1 {
					idx, _ := strconv.Atoi(match[1])
					if idx >= 0 && idx < len(base.tableCache) {
						tableHTML := base.tableCache[idx]
						tables = append(tables, tableToMarkdown(tableHTML))
					}
				}
			}
			// 用换行分隔多个表格
			tableMD = strings.Join(tables, "\n\n")
			// 移除所有占位符，保留剩下的文本
			content = tablePlaceholderRegex.ReplaceAllString(content, "")
		}

		// 移除内容中的页码标识
		content = pageRegex.ReplaceAllString(content, "")
		content = applyTrimSpaceIfNeeded(content, &StrategyBase{TrimSpace: true})

		// 合并 title + text + table 到 text 字段
		combinedText := content
		if doc.Title != "" {
			combinedText = doc.Title + "\n\n" + combinedText
		}
		if tableMD != "" {
			combinedText = combinedText + "\n\n" + tableMD
		}

		chunk := &Chunk{
			DocName:    fileName,
			DocMD5:     docMD5,
			SliceMD5:   sliceMD5,
			ID:         id,
			Pages:      finalPages,
			SegmentID:  i + 1,
			SuperiorID: doc.DocID,
			SliceContent: SliceContent{
				Title:   "", // 内容已合并到 Text
				Text:    combinedText,
				Table:   "", // 内容已合并到 Text
				Picture: "",
			},
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}
