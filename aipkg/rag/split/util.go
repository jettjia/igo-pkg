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

	"golang.org/x/net/html"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

var (
	urlRegex              = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex            = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	imageURLRegex         = regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	tableRegex            = regexp.MustCompile(`(?s)<table>.*?</table>`)
	pageRegex             = regexp.MustCompile(`(?i)<!--\s*Page:\s*(\d+)\s*-->`)
	tablePlaceholderRegex = regexp.MustCompile(`HTML_TABLE_PLACEHOLDER_(\d+)`)
	spaceOnlyRegex        = regexp.MustCompile(` {2,}`)
	multiNewlineRegex     = regexp.MustCompile(`\n{2,}`)
	pageMarkerRegex       = regexp.MustCompile(`(?i)\n?\s*<!--\s*Page:\s*\d+\s*-->\s*\n?`)
)

// cellInfo holds HTML table cell data including colspan/rowspan
type cellInfo struct {
	text    string
	rowspan int
	colspan int
}

// TableData holds structured table data for Header Injection approach
type TableData struct {
	Header []string   // Table headers: ["col1", "col2", "col3"]
	Rows   [][]string // Data rows: [["val1", "val2", "val3"], ...]
}

// tableDataRegex matches TABLE_DATA_ followed by an index number
var tableDataRegex = regexp.MustCompile(`TABLE_DATA_(\d+)`)

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

	// Step 1: 将HTML表格存储到缓存，使用短占位符
	// 后续在 convertToChunks 中解析并转换为 Header Injection 格式
	processed = tableRegex.ReplaceAllStringFunc(processed, func(tableHTML string) string {
		// 先解析为 TableData 存储
		tableData := parseHTMLTableToDataFrame(tableHTML)
		if tableData == nil || len(tableData.Rows) == 0 {
			return ""
		}
		idx := len(base.tableCache)
		base.tableCache = append(base.tableCache, tableData)
		result := "\n\n" + fmt.Sprintf("TABLE_DATA_%d", idx) + " "
		return result
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

// truncateMarkdownTableAtRows truncates a markdown table at complete row boundaries
// to ensure the table doesn't exceed maxChars while preserving row integrity.
// maxChars: maximum character length including all rows kept
func truncateMarkdownTableAtRows(tableMarkdown string, maxChars int) string {
	if maxChars <= 0 || runeLen(tableMarkdown) <= maxChars {
		return tableMarkdown
	}

	lines := strings.Split(tableMarkdown, "\n")
	if len(lines) <= 3 { // header + separator + at least one data row
		return tableMarkdown
	}

	var result []string
	resultLen := 0

	for i, line := range lines {
		lineLen := runeLen(line)
		// Add 1 for newline if not first line
		sep := 0
		if i > 0 {
			sep = 1
		}

		if resultLen+sep+lineLen > maxChars {
			break
		}
		result = append(result, line)
		resultLen += sep + lineLen
	}

	// Ensure we keep header and separator (lines[0] and lines[1])
	if len(result) < 2 {
		return strings.Join(lines[:3], "\n") // Keep at least header + separator + one row
	}

	return strings.Join(result, "\n")
}

// parseHTMLTableToDataFrame parses an HTML table into structured TableData
// This is used for the Header Injection approach where each row becomes a self-explanatory chunk
func parseHTMLTableToDataFrame(tableHTML string) *TableData {
	doc, err := html.Parse(strings.NewReader(tableHTML))
	if err != nil {
		return nil
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

	if len(allRows) < 2 { // Need at least header row + 1 data row
		return nil
	}

	// First row is header, rest are data rows
	headerRow := allRows[0]
	dataRows := allRows[1:]

	// Extract header texts
	header := make([]string, 0, len(headerRow))
	for _, cell := range headerRow {
		header = append(header, decodeHexEscapes(cell.text))
	}

	// Build 2D grid with colspan/rowspan handling for data rows
	maxCols := len(header)
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
					grid[rowIdx][colIdx+c] = decodeHexEscapes(cell.text)
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

	// Convert to [][]string format
	rows := make([][]string, 0, len(grid))
	for _, row := range grid {
		// Skip rows that are completely empty
		allEmpty := true
		for _, cell := range row {
			if cell != "" {
				allEmpty = false
				break
			}
		}
		if !allEmpty {
			rows = append(rows, row)
		}
	}

	return &TableData{
		Header: header,
		Rows:   rows,
	}
}

// tableRowToChunkContent constructs a self-explanatory chunk from a table row
// Format: "[heading] | [header1]: [value1] | [header2]: [value2] | ..."
// heading can be empty string if no title is available
func tableRowToChunkContent(heading string, header []string, row []string) string {
	var b strings.Builder

	// Add heading if available
	if heading != "" {
		b.WriteString(heading)
		b.WriteString(" | ")
	}

	// Build "col1: val1 | col2: val2 | ..." format
	for i, val := range row {
		if i > 0 {
			b.WriteString(" | ")
		}
		if i < len(header) {
			b.WriteString(header[i])
			b.WriteString(": ")
		}
		b.WriteString(val)
	}

	return b.String()
}

// tableRowToContent constructs content from a table row without heading
// Format: "[header1]: [value1] | [header2]: [value2] | ..."
// Empty values are skipped
func tableRowToContent(header []string, row []string) string {
	var b strings.Builder
	first := true
	for i, val := range row {
		if strings.TrimSpace(val) == "" {
			continue
		}
		if !first {
			b.WriteString(" | ")
		}
		if i < len(header) {
			b.WriteString(header[i])
			b.WriteString(": ")
		}
		b.WriteString(val)
		first = false
	}
	return b.String()
}

// splitRowByContent splits row content by columns if it exceeds maxChars
// heading is NOT included in the content (it goes to SliceContent.Title separately)
func splitRowByContent(row []string, header []string, maxChars int) []string {
	if maxChars <= 0 {
		return []string{tableRowToContent(header, row)}
	}

	// Check if entire row fits
	fullContent := tableRowToContent(header, row)
	if runeLen(fullContent) <= maxChars {
		return []string{fullContent}
	}

	// Need to split by columns
	var results []string
	var currentVals []string
	currentLen := 0

	for i, val := range row {
		colHeader := ""
		if i < len(header) {
			colHeader = header[i]
		}
		colEntry := colHeader + ": " + val
		colLen := runeLen(colEntry)

		sep := 0
		if len(currentVals) > 0 {
			sep = 3 // " | "
		}

		if currentLen > 0 && currentLen+sep+colLen > maxChars {
			results = append(results, buildContentWithHeader(header, currentVals))
			currentVals = nil
			currentLen = 0
		}

		currentVals = append(currentVals, val)
		currentLen += sep + colLen
	}

	if len(currentVals) > 0 {
		results = append(results, buildContentWithHeader(header, currentVals))
	}

	if len(results) == 0 {
		return []string{fullContent}
	}

	return results
}

// buildContentWithHeader builds content with header but only using selected columns
func buildContentWithHeader(header []string, values []string) string {
	var b strings.Builder
	for i, val := range values {
		if i > 0 {
			b.WriteString(" | ")
		}
		if i < len(header) {
			b.WriteString(header[i])
			b.WriteString(": ")
		}
		b.WriteString(val)
	}
	return b.String()
}

// splitRowByColumns splits a row by columns if it exceeds maxChars
// Each split chunk includes heading + header repeated
// Returns multiple chunks if the row is too long for a single chunk
func splitRowByColumns(heading string, header []string, row []string, maxChars int) []string {
	if maxChars <= 0 {
		return []string{tableRowToChunkContent(heading, header, row)}
	}

	// Check if entire row fits
	fullChunk := tableRowToChunkContent(heading, header, row)
	if runeLen(fullChunk) <= maxChars {
		return []string{fullChunk}
	}

	// Need to split by columns
	// header prefix: "heading | col1: val1 | col2: val2 | ... | "
	// Calculate base overhead per chunk (heading + partial columns)
	var results []string
	var currentCols []string
	currentLen := 0

	for i, val := range row {
		colHeader := ""
		if i < len(header) {
			colHeader = header[i]
		}
		// Format: "colHeader: val"
		colEntry := colHeader + ": " + val
		colLen := runeLen(colEntry)

		// Add separator if not first column in this chunk
		sep := 0
		if len(currentCols) > 0 {
			sep = 3 // " | "
		}

		// Check if adding this column would exceed limit
		if currentLen > 0 && currentLen+sep+colLen > maxChars {
			// Save current chunk and start new one
			results = append(results, buildChunkWithHeader(heading, header, currentCols))
			currentCols = nil
			currentLen = 0
		}

		currentCols = append(currentCols, val)
		currentLen += sep + colLen
	}

	// Don't forget the last chunk
	if len(currentCols) > 0 {
		results = append(results, buildChunkWithHeader(heading, header, currentCols))
	}

	// Ensure we at least return something
	if len(results) == 0 {
		return []string{fullChunk}
	}

	return results
}

// buildChunkWithHeader builds a chunk with heading and header, but only using selected columns
func buildChunkWithHeader(heading string, header []string, values []string) string {
	var b strings.Builder

	if heading != "" {
		b.WriteString(heading)
		b.WriteString(" | ")
	}

	for i, val := range values {
		if i > 0 {
			b.WriteString(" | ")
		}
		// Use header[i] if available
		if i < len(header) {
			b.WriteString(header[i])
			b.WriteString(": ")
		}
		b.WriteString(val)
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

		// Build JSON object with header names as keys
		obj := make(map[string]string)
		for colIdx, header := range headers {
			header = decodeHexEscapes(header)
			cell := decodeHexEscapes(row[colIdx])
			// Clean cell content: remove newlines, excessive spaces
			cell = strings.ReplaceAll(cell, "\n", " ")
			cell = strings.ReplaceAll(cell, "\r", " ")
			cell = collapseSpaces(cell)
			obj[header] = cell
		}

		jsonBytes, err := json.Marshal(obj)
		if err != nil {
			continue
		}
		result.WriteString(string(jsonBytes))
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

// processTableDataChunks handles TABLE_DATA_\d+ short placeholders and creates Header Injection format chunks
func processTableDataChunks(content string, doc *schema.Document, fileName string, docMD5 string, finalPages []int, base *StrategyBase) []*Chunk {
	var chunks []*Chunk

	// Find all placeholders in content
	matches := tableDataRegex.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return nil
	}

	// Extract and process each table placeholder
	for matchIdx, match := range matches {
		if len(match) < 4 {
			continue
		}
		// match[0..1] is the full match, match[2..3] is the first group (table index)
		idxStart := match[2]
		idxEnd := match[3]
		idxStr := content[idxStart:idxEnd]

		// Parse the index
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			continue
		}

		// Look up table data from cache
		if idx < 0 || idx >= len(base.tableCache) {
			continue
		}
		tableData := base.tableCache[idx]
		if tableData == nil {
			continue
		}

		// Build heading from doc.Title
		heading := doc.Title
		if len(doc.HeadingPath) > 0 {
			heading = strings.Join(doc.HeadingPath, " / ")
		}

		// Process each row as a separate chunk
		for rowIdx, row := range tableData.Rows {
			// For Header Injection format, the content is "col1: val1 | col2: val2"
			// without the heading prefix (heading goes to SliceContent.Title)
			chunkText := tableRowToContent(row, tableData.Header)

			// Handle long rows by splitting columns
			if base.ChunkSize > 0 && runeLen(chunkText) > base.ChunkSize {
				parts := splitRowByContent(row, tableData.Header, base.ChunkSize)
				for _, part := range parts {
					chunk := createChunkFromText(part, heading, fileName, docMD5, matchIdx*1000+rowIdx, finalPages, doc)
					chunks = append(chunks, chunk)
				}
			} else {
				chunk := createChunkFromText(chunkText, heading, fileName, docMD5, matchIdx*1000+rowIdx, finalPages, doc)
				chunks = append(chunks, chunk)
			}
		}
	}

	return chunks
}

// createChunkFromText creates a Chunk from text content
func createChunkFromText(text string, title string, fileName string, docMD5 string, index int, pages []int, doc *schema.Document) *Chunk {
	text = strings.TrimSpace(text)
	sliceMD5 := calculateMD5(text)
	id := fmt.Sprintf("%s-%d", sliceMD5[:8], index)

	// Text 需要包含 title（ES 两个字段：title 和 content）
	fullText := text
	if title != "" {
		fullText = title + "\n\n" + text
	}

	return &Chunk{
		DocName:    fileName,
		DocMD5:     docMD5,
		SliceMD5:   sliceMD5,
		ID:         id,
		Pages:      pages,
		SegmentID:  index + 1,
		SuperiorID: doc.DocID,
		HeadingPath: doc.HeadingPath,
		SliceContent: SliceContent{
			Title:   title,
			Text:    fullText,
			Table:   "",
			Picture: "",
		},
	}
}

func convertToChunks(docs []*schema.Document, fileName string, originalText string, base *StrategyBase) []*Chunk {
	docMD5 := calculateMD5(originalText)
	chunks := make([]*Chunk, 0, len(docs))

	for i, doc := range docs {
		sliceMD5 := calculateMD5(doc.Content)
		id := fmt.Sprintf("%s-%d", sliceMD5[:8], i)

		// 提取当前切片内容中的页码标识
		pageMap := make(map[int]bool)

		// 1. 检查文档自带的 Page
		if doc.Page > 0 {
			pageMap[doc.Page] = true
		}

		// 2. 扫描内容中的页码标识（页码标记保留在文本中）
		contentMatches := pageRegex.FindAllStringSubmatch(doc.Content, -1)
		for _, cm := range contentMatches {
			if p, err := strconv.Atoi(cm[1]); err == nil {
				pageMap[p] = true
			}
		}

		// 3. 如果仍没有页码信息，默认第1页
		if len(pageMap) == 0 {
			pageMap[1] = true
		}

		finalPages := make([]int, 0, len(pageMap))
		for p := range pageMap {
			finalPages = append(finalPages, p)
		}
		sort.Ints(finalPages)

		// 内容已经是处理好的文本（可能包含 TABLE_DATA_PLACEHOLDER）
		content := pageRegex.ReplaceAllString(doc.Content, "")
		content = multiNewlineRegex.ReplaceAllString(content, "\n")
		content = applyTrimSpaceIfNeeded(content, &StrategyBase{TrimSpace: true})

		// 检查是否包含表格占位符
		if tableDataRegex.MatchString(content) {
			// 使用 Header Injection 方式处理表格
			chunks = append(chunks, processTableDataChunks(content, doc, fileName, docMD5, finalPages, base)...)
			continue
		}

		// 构建 title - 使用 HeadingPath 或 doc.Title
		title := ""
		if len(doc.HeadingPath) > 0 {
			title = strings.Join(doc.HeadingPath, " / ")
		} else if doc.Title != "" {
			title = doc.Title
		}

		// 清理后的内容
		cleanContent := multiNewlineRegex.ReplaceAllString(content, "\n")

		// Text 需要包含 title（ES 两个字段：title 和 content）
		text := cleanContent
		if title != "" {
			text = title + "\n\n" + cleanContent
		}

		chunk := &Chunk{
			DocName:     fileName,
			DocMD5:      docMD5,
			SliceMD5:    sliceMD5,
			ID:          id,
			Pages:       finalPages,
			SegmentID:   i + 1,
			SuperiorID:  doc.DocID,
			HeadingPath:  doc.HeadingPath,
			SliceContent: SliceContent{
				Title:   title,
				Text:    text,
				Table:   "",
				Picture: "",
			},
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}
