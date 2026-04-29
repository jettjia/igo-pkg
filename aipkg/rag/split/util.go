package split

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

var (
	urlRegex          = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex        = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	imageURLRegex     = regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	tableRegex        = regexp.MustCompile(`(?s)<table>.*?</table>`)
	pageRegex         = regexp.MustCompile(`(?i)<!--\s*Page:\s*(\d+)\s*-->`)
	spaceOnlyRegex    = regexp.MustCompile(` {2,}`)
	multiNewlineRegex = regexp.MustCompile(`\n{2,}`)
)

// cellInfo holds HTML table cell data including colspan/rowspan
type cellInfo struct {
	text    string
	rowspan int
	colspan int
}

// pageMarkerInfo 存储页码标记的位置信息
type pageMarkerInfo struct {
	Page    int // 页码
	RunePos int // 在文本中的 rune 位置
	BytePos int // 在文本中的 byte 位置
}

// extractPageMarkers 提取文本中所有页码标记及其位置
func extractPageMarkers(text string) []pageMarkerInfo {
	matches := pageRegex.FindAllStringSubmatchIndex(text, -1)
	markers := make([]pageMarkerInfo, 0, len(matches))

	for _, match := range matches {
		if len(match) >= 4 {
			pageStr := text[match[2]:match[3]]
			page, err := strconv.Atoi(pageStr)
			if err != nil {
				continue
			}
			// 计算该匹配在原文中的 byte 位置
			bytePos := match[0]
			// 计算 rune 位置（用于后续比较）
			runePos := len([]rune(text[:bytePos]))
			markers = append(markers, pageMarkerInfo{
				Page:    page,
				RunePos: runePos,
				BytePos: bytePos,
			})
		}
	}

	return markers
}

// findPagesForContent 根据内容在原文中的位置，返回该内容对应的页码
// 找到最后一个（最大的）RunePos <= contentRunePos 的页码，即内容起始位置所在的页码
func findPagesForContent(content string, originalText string, markers []pageMarkerInfo) []int {
	if len(markers) == 0 {
		return []int{1}
	}

	pos := strings.Index(originalText, content)
	if pos < 0 {
		return []int{markers[0].Page}
	}

	contentRunePos := len([]rune(originalText[:pos]))

	// 找到最后一个（最大的）RunePos <= contentRunePos 的页码
	var lastPage int
	for _, m := range markers {
		if m.RunePos <= contentRunePos {
			lastPage = m.Page
		}
	}

	if lastPage == 0 {
		return []int{1}
	}
	return []int{lastPage}
}

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

	// Step 1: 将HTML表格直接转换为JSON行格式（每行一个JSON对象）
	// 这样每行都是独立的 chunk，自然避免超长问题
	processed = tableRegex.ReplaceAllStringFunc(processed, func(tableHTML string) string {
		jsonLines := tableToJSONLines(tableHTML)
		if jsonLines == "" {
			return ""
		}
		return "\n\n" + jsonLines + "\n\n"
	})

	// Step 2: 移除URL和邮箱
	if base.RemoveURLAndEmail {
		processed = urlRegex.ReplaceAllString(processed, "")
		processed = emailRegex.ReplaceAllString(processed, "")
	}

	// Step 3: 移除图片URL
	if base.RemoveImageURL {
		processed = imageURLRegex.ReplaceAllString(processed, "")
	}

	// Step 4: 统一换行符
	processed = strings.ReplaceAll(processed, "\r\n", "\n")

	// Step 5: 规范化空白符（可选）
	if base.NormalizeWhitespace {
		processed = strings.ReplaceAll(processed, "\t", " ")
		processed = spaceOnlyRegex.ReplaceAllString(processed, " ")
		processed = multiNewlineRegex.ReplaceAllString(processed, "\n")
	}

	// Step 6: TrimSpace（可选）
	if base.TrimSpace {
		processed = strings.TrimSpace(processed)
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

// tableToJSONLines converts an HTML table to JSON Lines format (one JSON object per row)
// Each row becomes a JSON object with header names as keys
// This solves the problem of long markdown table cells that exceed chunkSize
func tableToJSONLines(tableHTML string) string {
	doc, err := html.Parse(strings.NewReader(tableHTML))
	if err != nil {
		return ""
	}

	// Collect all rows with their cell info
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

	if len(allRows) < 2 {
		return ""
	}

	// First row is header, rest are data rows
	headers := extractCellTexts(allRows[0])
	dataRows := allRows[1:]

	// Build 2D grid with colspan/rowspan handling
	maxCols := len(headers)
	grid := make([][]string, len(dataRows))
	for i := range grid {
		grid[i] = make([]string, maxCols)
	}

	rowspanGrid := make([][]int, len(dataRows))
	for i := range rowspanGrid {
		rowspanGrid[i] = make([]int, maxCols)
	}

	// Fill grid for data rows
	for rowIdx, row := range dataRows {
		colIdx := 0
		for _, cell := range row {
			for colIdx < maxCols && rowspanGrid[rowIdx][colIdx] > 0 {
				colIdx++
			}
			if colIdx >= maxCols {
				break
			}
			for c := 0; c < cell.colspan && colIdx+c < maxCols; c++ {
				if c == 0 {
					grid[rowIdx][colIdx+c] = cell.text
				}
			}
			for r := 1; r < cell.rowspan && rowIdx+r < len(dataRows); r++ {
				for c := 0; c < cell.colspan && colIdx+c < maxCols; c++ {
					rowspanGrid[rowIdx+r][colIdx+c]++
				}
			}
			colIdx += cell.colspan
		}
	}

	// Build JSON Lines output
	var result strings.Builder
	for _, row := range grid {
		// Skip rows that are completely empty
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

		// Build JSON object preserving column order
		var buf strings.Builder
		buf.WriteByte('{')
		first := true
		for colIdx, header := range headers {
			header = decodeHexEscapes(header)
			cell := decodeHexEscapes(row[colIdx])
			// Clean cell content: remove newlines, excessive spaces
			cell = strings.ReplaceAll(cell, "\n", " ")
			cell = strings.ReplaceAll(cell, "\r", " ")
			cell = collapseSpaces(cell)
			if !first {
				buf.WriteByte(',')
			}
			first = false
			keyBytes, _ := json.Marshal(header)
			valBytes, _ := json.Marshal(cell)
			buf.Write(keyBytes)
			buf.WriteByte(':')
			buf.Write(valBytes)
		}
		buf.WriteByte('}')

		result.WriteString(buf.String())
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// extractCellTexts extracts just the text from cells without colspan/rowspan handling
func extractCellTexts(row []cellInfo) []string {
	result := make([]string, 0, len(row))
	for _, cell := range row {
		result = append(result, cell.text)
	}
	return result
}

// collapseSpaces replaces multiple spaces with single space
func collapseSpaces(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' {
			if !prevSpace {
				b.WriteRune(r)
				prevSpace = true
			}
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
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

func newDocumentWithHeading(content string, title string, depth int, headingPath []string) *schema.Document {
	doc := newDocument(content, title, depth)
	doc.HeadingPath = headingPath
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

// isMeaninglessChunk 判断 chunk 内容是否无意义
// 无意义的 chunk 包括：只有页码标记、只有空白字符、只有特殊字符等
func isMeaninglessChunk(text string) bool {
	if text == "" {
		return true
	}

	// 清理后检查是否为空
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return true
	}

	// 检查是否只包含页码标记残留（如 "<!-- Page: XX -->" 或 "-- Page: XX -->"）
	// 这种清理不干净的情况
	pageOnlyRegex := regexp.MustCompile(`^[\s\p{Zs}]*(?:<!--.*?-->|&lt;!--.*?--&gt;| Page:.*)?[\s\p{Zs}]*$`)
	if pageOnlyRegex.MatchString(trimmed) {
		return true
	}

	// 检查是否只包含页码标记的变体（&lt;! 和 &gt; 是 HTML 转义）
	if strings.HasPrefix(trimmed, "&lt;!--") || strings.HasPrefix(trimmed, "<!") {
		if strings.Contains(trimmed, "Page:") && strings.Contains(trimmed, "-->") {
			return true
		}
	}

	// 检查以 "-- Page:" 开头并以 "-->" 结尾的页码标记变体
	if strings.HasPrefix(trimmed, "-- Page:") || strings.HasPrefix(trimmed, "-- page:") {
		if strings.HasSuffix(trimmed, "-->") || strings.HasSuffix(trimmed, "--&gt;") {
			return true
		}
	}

	// 检查是否只有特殊字符和空白
	hasLetter := false
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == ':' {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return true
	}

	return false
}

func convertToChunks(docs []*schema.Document, fileName string, originalText string, base *StrategyBase, markers []pageMarkerInfo) []*Chunk {
	docMD5 := calculateMD5(originalText)
	chunks := make([]*Chunk, 0, len(docs))

	for i, doc := range docs {
		sliceMD5 := calculateMD5(doc.Content)
		id := fmt.Sprintf("%s-%d", sliceMD5[:8], i)

		// 使用位置信息推断页码
		var finalPages []int

		// 优先级：1. doc.Page > 0 直接使用  2. 位置推断  3. 内容搜索  4. 默认第1页
		if doc.Page > 0 {
			// 1. 文档自带的 Page 优先级最高
			finalPages = []int{doc.Page}
		} else {
			// 2. 尝试在 originalText 中找到 chunk 内容的位置，使用位置推断
			pos := strings.Index(originalText, doc.Content)
			if pos >= 0 {
				finalPages = findPagesForContent(doc.Content, originalText, markers)
			} else {
				// 3. 找不到（内容被截断），回退到在内容中搜索页码标记
				pageMap := make(map[int]bool)

				// 扫描内容中的页码标识（页码标记保留在文本中）
				contentMatches := pageRegex.FindAllStringSubmatch(doc.Content, -1)
				for _, cm := range contentMatches {
					if p, err := strconv.Atoi(cm[1]); err == nil {
						pageMap[p] = true
					}
				}

				// 4. 如果仍没有页码信息，默认第1页
				if len(pageMap) == 0 {
					pageMap[1] = true
				}

				finalPages = make([]int, 0, len(pageMap))
				for p := range pageMap {
					finalPages = append(finalPages, p)
				}
				sort.Ints(finalPages)
			}
		}

		// 内容已经是处理好的文本（包含markdown表格），只需清理页码标记
		content := pageRegex.ReplaceAllString(doc.Content, "")
		content = multiNewlineRegex.ReplaceAllString(content, "\n")
		content = applyTrimSpaceIfNeeded(content, &StrategyBase{TrimSpace: true})

		// 跳过空内容和无意义的 chunk（如只有页码标记的情况）
		if content == "" || isMeaninglessChunk(content) {
			continue
		}

		// // 合并 title + text 到 text 字段（表格已在 preProcessText 中直接转为markdown）
		// combinedText := content
		// if doc.Title != "" {
		// 	combinedText = doc.Title + "\n\n" + combinedText
		// }
		// // 合并后再次清理多余的换行符
		// combinedText = multiNewlineRegex.ReplaceAllString(combinedText, "\n")
		// // 确保最终 chunk 不超过 ChunkSize
		// if base.ChunkSize > 0 && runeLen(combinedText) > base.ChunkSize {
		// 	runes := []rune(combinedText)
		// 	combinedText = string(runes[:base.ChunkSize])
		// }

		chunk := &Chunk{
			DocName:     fileName,
			DocMD5:      docMD5,
			SliceMD5:    sliceMD5,
			ID:          id,
			Pages:       finalPages,
			SegmentID:   i + 1,
			SuperiorID:  doc.DocID,
			HeadingPath: doc.HeadingPath,
			SliceContent: SliceContent{
				Title:   doc.Title,
				Text:    content,
				Table:   "", // 内容已合并到 Text
				Picture: "",
			},
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}
