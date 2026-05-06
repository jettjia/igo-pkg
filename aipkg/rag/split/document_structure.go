package split

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

// DocumentStructureStrategy 文档标题结构分块策略
//
// 目标：
// - 识别中文标题（H1/H2/H3）
// - 标题必须与其下属内容处于同一块，禁止标题与内容跨块切割
// - 优先将低层级标题内容归属于上层标题块，超长再拆分
type DocumentStructureStrategy struct {
	StrategyBase
	MaxDepth          int
	SemanticThreshold float64
	SkipEmptyHeadings bool
}

func NewDocumentStructureStrategy() *DocumentStructureStrategy {
	return &DocumentStructureStrategy{
		StrategyBase: StrategyBase{
			ChunkSize:           500,
			OverlapRatio:        0.1,
			RemoveURLAndEmail:   false,
			NormalizeWhitespace: false,
			TrimSpace:           false,
		},
		MaxDepth:          3,
		SemanticThreshold: 0.5,
		SkipEmptyHeadings: false,
	}
}

func (s *DocumentStructureStrategy) GetType() StrategyType {
	return StrategyTypeDocumentStructure
}

func (s *DocumentStructureStrategy) Validate() error {
	if err := s.validateBase(); err != nil {
		return err
	}
	if s.MaxDepth <= 0 {
		s.MaxDepth = 3
	}
	if s.MaxDepth > 6 {
		s.MaxDepth = 6
	}
	if s.SemanticThreshold == 0 {
		s.SemanticThreshold = 0.5
	}
	if s.SemanticThreshold < 0 || s.SemanticThreshold > 1 {
		s.SemanticThreshold = 0.5
	}
	return nil
}

func (s *DocumentStructureStrategy) Split(ctx context.Context, text string, fileName string) ([]*Chunk, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	processed := preProcessText(text, &s.StrategyBase)
	processed = applyTrimSpaceIfNeeded(processed, &s.StrategyBase)
	if processed == "" {
		return nil, nil
	}

	// 提取页码标记，用于后续页码推断
	markers := extractPageMarkers(processed)

	root := parseHeadingTree(processed, s.MaxDepth)
	fmt.Fprintf(os.Stderr, "[DEBUG Split] root.children=%d processedLen=%d\n", len(root.children), runeLen(processed))
	if len(root.children) > 0 {
		for i, child := range root.children {
			fmt.Fprintf(os.Stderr, "[DEBUG Split] child[%d] title=%q depth=%d contentLines=%d numChildren=%d\n", i, child.title, child.depth, len(child.content), len(child.children))
		}
	}
	if len(root.children) == 0 {
		// 检测是否为表格内容（以 JSON 行为主）
		lines := strings.Split(processed, "\n")
		if isMostlyTableRows(lines, 0.5) {
			return s.splitTableRows(ctx, processed, fileName, markers)
		}

		semantic := NewSemanticStrategy()
		semantic.StrategyBase = s.StrategyBase
		semantic.Threshold = s.SemanticThreshold
		return semantic.Split(ctx, processed, fileName)
	}

	var docs []*schema.Document
	for _, top := range root.children {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		docs = append(docs, s.splitSection(ctx, top, nil)...)
	}
	chunks := convertToChunks(docs, fileName, processed, &s.StrategyBase, markers)
	chunks = applyOverlapToChunks(chunks, s.ChunkSize, s.OverlapRatio)
	return chunks, nil
}

// isJSONLine 检测是否为单行 JSON 对象
func isJSONLine(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}'
}

// isMostlyTableRows 检测内容是否以表格 JSON 行为主
func isMostlyTableRows(lines []string, threshold float64) bool {
	if len(lines) == 0 {
		return false
	}
	jsonCount := 0
	for _, line := range lines {
		if isJSONLine(line) {
			jsonCount++
		}
	}
	return float64(jsonCount)/float64(len(lines)) > threshold
}

// splitTableRows 表格专用分词：以 JSON 行为原子单位合并
func (s *DocumentStructureStrategy) splitTableRows(ctx context.Context, text string, fileName string, markers []pageMarkerInfo) ([]*Chunk, error) {
	lines := strings.Split(text, "\n")
	var chunks []string
	var current strings.Builder
	currentSize := 0

	for _, line := range lines {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lineSize := runeLen(line)

		// 单行超长则保留为原子
		if lineSize > s.ChunkSize {
			if current.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(current.String()))
				current.Reset()
				currentSize = 0
			}
			chunks = append(chunks, line)
			continue
		}

		// 加上此行会超限，开始新 chunk
		if currentSize > 0 && currentSize+lineSize > s.ChunkSize {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
			currentSize = 0
		}

		if current.Len() > 0 {
			current.WriteString("\n")
			currentSize++
		}
		current.WriteString(line)
		currentSize += lineSize
	}

	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	// 转换为 schema.Document
	docs := make([]*schema.Document, 0, len(chunks))
	for _, c := range chunks {
		c = applyTrimSpaceIfNeeded(c, &s.StrategyBase)
		if c == "" {
			continue
		}
		docs = append(docs, newDocument(c, "", 0))
	}

	return applyOverlapToChunks(convertToChunks(docs, fileName, text, &s.StrategyBase, markers), s.ChunkSize, s.OverlapRatio), nil
}

type sectionNode struct {
	depth      int
	title      string
	heading    string
	content    []string
	children   []*sectionNode
	hasHeading bool
}

var (
	markdownHeadingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	h1ArabicRegex        = regexp.MustCompile(`^\s*(\d+)\.\s*(.+?)\s*$`)
	h1CnRegex            = regexp.MustCompile(`^\s*([一二三四五六七八九十百千]+)、\s*(.+?)\s*$`)
	h2ArabicRegex        = regexp.MustCompile(`^\s*(\d+\.\d+)\s+(.+?)\s*$`)
	h2CnRegex            = regexp.MustCompile(`^\s*（([一二三四五六七八九十百千]+)）\s*(.+?)\s*$`)
	h3ArabicRegex        = regexp.MustCompile(`^\s*(\d+\.\d+\.\d+)\s+(.+?)\s*$`)
	spaceLineRegex       = regexp.MustCompile(`^\s*$`)
)

func parseHeadingLine(line string, maxDepth int, markdownOnly bool) (depth int, title string, heading string, ok bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return 0, "", "", false
	}

	if m := markdownHeadingRegex.FindStringSubmatch(trimmed); m != nil {
		d := len(m[1])
		if d <= 0 || d > maxDepth {
			return 0, "", "", false
		}
		t := strings.TrimSpace(m[2])
		t = strings.Trim(t, "*# ")
		return d, t, trimmed, true
	}

	if markdownOnly {
		return 0, "", "", false
	}

	if maxDepth >= 3 {
		if m := h3ArabicRegex.FindStringSubmatch(trimmed); m != nil {
			return 3, strings.TrimSpace(m[2]), trimmed, true
		}
	}

	if maxDepth >= 2 {
		if m := h2ArabicRegex.FindStringSubmatch(trimmed); m != nil {
			return 2, strings.TrimSpace(m[2]), trimmed, true
		}
		if m := h2CnRegex.FindStringSubmatch(trimmed); m != nil {
			return 2, strings.TrimSpace(m[2]), trimmed, true
		}
	}

	if maxDepth >= 1 {
		if m := h1ArabicRegex.FindStringSubmatch(trimmed); m != nil {
			return 1, strings.TrimSpace(m[2]), trimmed, true
		}
		if m := h1CnRegex.FindStringSubmatch(trimmed); m != nil {
			return 1, strings.TrimSpace(m[2]), trimmed, true
		}
	}

	return 0, "", "", false
}

func parseHeadingTree(text string, maxDepth int) *sectionNode {
	root := &sectionNode{depth: 0}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	inCodeBlock := false
	hasMarkdownHeading := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		if markdownHeadingRegex.MatchString(trimmed) {
			hasMarkdownHeading = true
			break
		}
	}

	stack := []*sectionNode{root}
	inCodeBlock = false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			stack[len(stack)-1].content = append(stack[len(stack)-1].content, line)
			continue
		}
		if inCodeBlock {
			stack[len(stack)-1].content = append(stack[len(stack)-1].content, line)
			continue
		}

		if d, title, heading, ok := parseHeadingLine(line, maxDepth, hasMarkdownHeading); ok {
			for len(stack) > 0 && stack[len(stack)-1].depth >= d {
				stack = stack[:len(stack)-1]
			}
			parent := stack[len(stack)-1]
			node := &sectionNode{
				depth:      d,
				title:      title,
				heading:    heading,
				hasHeading: true,
			}
			parent.children = append(parent.children, node)
			stack = append(stack, node)
			continue
		}
		stack[len(stack)-1].content = append(stack[len(stack)-1].content, line)
	}

	return root
}

type sectionUnit struct {
	kind string
	text string
	sec  *sectionNode
}

func truncateString(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}

func (s *DocumentStructureStrategy) splitSection(ctx context.Context, sec *sectionNode, titlePath []string) []*schema.Document {
	if sec == nil || !sec.hasHeading {
		return nil
	}

	nextPath := append(append([]string{}, titlePath...), sec.title)
	docTitle := strings.Join(nextPath, " / ")

	directContent := strings.TrimSpace(strings.Join(sec.content, "\n"))
	if s.SkipEmptyHeadings && strings.TrimSpace(directContent) == "" && len(sec.children) > 0 {
		var out []*schema.Document
		for _, child := range sec.children {
			if ctx.Err() != nil {
				return nil
			}
			out = append(out, s.splitSection(ctx, child, nextPath)...)
		}
		return out
	}

	prefixText := strings.TrimSpace(sec.heading) + "\n\n"
	prefixLen := runeLen(prefixText)
	fmt.Fprintf(os.Stderr, "[DEBUG splitSection START] title=%q heading=%q numChildren=%d directContentLen=%d\n", docTitle, sec.heading, len(sec.children), runeLen(directContent))
	if runeLen(directContent) > 0 {
		fmt.Fprintf(os.Stderr, "[DEBUG splitSection DIRECT] first200=%q\n", truncateString(directContent, 200))
	}
	if prefixLen >= s.ChunkSize {
		truncated := prefixText
		if s.ChunkSize > 0 {
			r := []rune(prefixText)
			if len(r) > s.ChunkSize {
				truncated = string(r[:s.ChunkSize])
			}
		}
		return []*schema.Document{newDocumentWithHeading(applyTrimSpaceIfNeeded(truncated, &s.StrategyBase), docTitle, sec.depth, nextPath)}
	}

	available := s.ChunkSize - prefixLen

	var units []sectionUnit

	direct := directContent
	direct = strings.TrimLeftFunc(direct, func(r rune) bool { return r == '\n' || r == '\r' || r == '\t' || r == ' ' })
	if direct != "" && !spaceLineRegex.MatchString(direct) {
		units = append(units, sectionUnit{kind: "direct", text: direct})
	}

	for _, child := range sec.children {
		childText := strings.TrimSpace(renderSection(child))
		if childText == "" {
			continue
		}
		units = append(units, sectionUnit{kind: "child", text: childText, sec: child})
	}

	var docs []*schema.Document

	addChunkBody := func(body string) {
		body = applyTrimSpaceIfNeeded(body, &s.StrategyBase)
		if body == "" {
			return
		}
		chunkText := prefixText + body
		chunkText = applyTrimSpaceIfNeeded(chunkText, &s.StrategyBase)
		if s.ChunkSize > 0 && runeLen(chunkText) > s.ChunkSize {
			if isJSONLine(body) {
				docs = append(docs, newDocumentWithHeading(chunkText, docTitle, sec.depth, nextPath))
				return
			}
			if hasJSONLines(body) {
				bodyLines := strings.Split(body, "\n")
				var kept []string
				keptSize := prefixLen
				for _, bl := range bodyLines {
					bl = strings.TrimSpace(bl)
					if bl == "" {
						continue
					}
					lineSize := runeLen(bl) + 1
					if keptSize+lineSize > s.ChunkSize && len(kept) > 0 {
						break
					}
					kept = append(kept, bl)
					keptSize += lineSize
				}
				if len(kept) > 0 {
					chunkText = prefixText + strings.Join(kept, "\n")
					chunkText = applyTrimSpaceIfNeeded(chunkText, &s.StrategyBase)
				}
				remaining := bodyLines[len(kept):]
				for len(remaining) > 0 {
					remainingText := strings.Join(remaining, "\n")
					remainingText = applyTrimSpaceIfNeeded(remainingText, &s.StrategyBase)
					if remainingText == "" {
						break
					}
					if runeLen(remainingText) <= s.ChunkSize {
						docs = append(docs, newDocumentWithHeading(remainingText, docTitle, sec.depth, nextPath))
						break
					}
					var subKept []string
					subSize := 0
					for _, rl := range remaining {
						rl = strings.TrimSpace(rl)
						if rl == "" {
							continue
						}
						rlSize := runeLen(rl) + 1
						if subSize+rlSize > s.ChunkSize && len(subKept) > 0 {
							break
						}
						subKept = append(subKept, rl)
						subSize += rlSize
					}
					if len(subKept) > 0 {
						subText := strings.Join(subKept, "\n")
						docs = append(docs, newDocumentWithHeading(subText, docTitle, sec.depth, nextPath))
						remaining = remaining[len(subKept):]
					} else {
						break
					}
				}
			} else {
				runes := []rune(chunkText)
				chunkText = string(runes[:s.ChunkSize])
			}
		}
		docs = append(docs, newDocumentWithHeading(chunkText, docTitle, sec.depth, nextPath))
	}

	var curUnits []string
	curSize := 0

	flushUnits := func() {
		if len(curUnits) == 0 {
			return
		}
		body := strings.Join(curUnits, "\n\n")
		addChunkBody(body)
		curUnits = nil
		curSize = 0
	}

	for _, u := range units {
		if ctx.Err() != nil {
			return nil
		}

		unitText := strings.TrimSpace(u.text)
		if unitText == "" {
			continue
		}

		if runeLen(unitText) > available {
			flushUnits()

			if u.kind == "child" && u.sec != nil {
				childDocs := s.splitSection(ctx, u.sec, nextPath)
				docs = append(docs, childDocs...)
				continue
			}

			hasJSON := hasJSONLines(unitText)
			isSingleJSON := isJSONLine(strings.TrimSpace(unitText))
			fmt.Fprintf(os.Stderr, "[DEBUG splitSection] title=%q kind=%s unitLen=%d available=%d hasJSONLines=%v isSingleJSON=%v first100=%q\n", docTitle, u.kind, runeLen(unitText), available, hasJSON, isSingleJSON, truncateString(unitText, 100))
			parts := splitTextForStructure(unitText, available)
			fmt.Fprintf(os.Stderr, "[DEBUG splitSection] parts count=%d\n", len(parts))
			for pi, p := range parts {
				isJL := isJSONLine(strings.TrimSpace(p))
				fmt.Fprintf(os.Stderr, "[DEBUG splitSection] part[%d] len=%d isJSONLine=%v first80=%q\n", pi, runeLen(p), isJL, truncateString(p, 80))
				addChunkBody(p)
			}
			continue
		}

		extraSep := 0
		if len(curUnits) > 0 {
			extraSep = runeLen("\n\n")
		}

		if curSize+extraSep+runeLen(unitText) > available && curSize > 0 {
			flushUnits()
			extraSep = 0
		}

		curUnits = append(curUnits, unitText)
		curSize += extraSep + runeLen(unitText)
	}
	flushUnits()

	return docs
}

func renderSection(sec *sectionNode) string {
	if sec == nil || !sec.hasHeading {
		return ""
	}
	var b strings.Builder
	b.WriteString(sec.heading)

	direct := strings.TrimSpace(strings.Join(sec.content, "\n"))
	if direct != "" && !spaceLineRegex.MatchString(direct) {
		b.WriteString("\n")
		b.WriteString(direct)
	}

	for _, child := range sec.children {
		childText := strings.TrimSpace(renderSection(child))
		if childText == "" {
			continue
		}
		b.WriteString("\n\n")
		b.WriteString(childText)
	}

	return strings.TrimSpace(b.String())
}

func splitTextForStructure(text string, chunkSize int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	if chunkSize <= 0 || runeLen(trimmed) <= chunkSize {
		return []string{trimmed}
	}

	if hasJSONLines(trimmed) {
		return splitJSONAware(trimmed, chunkSize)
	}

	protected, blocks := protectCodeBlocks(trimmed)
	parts := splitBySentences(protected, chunkSize)

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = restoreCodeBlocks(p, blocks)
		out = append(out, p)
	}
	return out
}

func hasJSONLines(text string) bool {
	lines := strings.Split(text, "\n")
	jsonCount := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if isJSONLine(line) {
			jsonCount++
		}
	}
	return jsonCount >= 2
}

func splitJSONAware(text string, chunkSize int) []string {
	lines := strings.Split(text, "\n")
	type textBlock struct {
		isJSON bool
		text   string
	}
	var blocks []textBlock

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if isJSONLine(line) {
			blocks = append(blocks, textBlock{isJSON: true, text: line})
		} else {
			blocks = append(blocks, textBlock{isJSON: false, text: line})
		}
	}

	var result []string
	var current strings.Builder
	currentSize := 0

	flushCurrent := func() {
		if current.Len() > 0 {
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()
			currentSize = 0
		}
	}

	for _, block := range blocks {
		blockSize := runeLen(block.text)

		if block.isJSON {
			if currentSize > 0 && currentSize+1+blockSize > chunkSize {
				flushCurrent()
			}

			if blockSize > chunkSize {
				flushCurrent()
				result = append(result, block.text)
				continue
			}

			if currentSize > 0 {
				current.WriteString("\n")
				currentSize++
			}
			current.WriteString(block.text)
			currentSize += blockSize
		} else {
			if currentSize > 0 && currentSize+1+blockSize > chunkSize {
				flushCurrent()
			}

			if blockSize > chunkSize {
				flushCurrent()
				parts := splitBySentences(block.text, chunkSize)
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						result = append(result, p)
					}
				}
				continue
			}

			if currentSize > 0 {
				current.WriteString("\n")
				currentSize++
			}
			current.WriteString(block.text)
			currentSize += blockSize
		}
	}
	flushCurrent()

	return result
}

// splitBySentences 按句子边界切分，保持语义完整性
// 分隔符优先级：空行 > 句号 > 感叹号 > 问号 > 换行
func splitBySentences(text string, chunkSize int) []string {
	// 按句子分隔符切分，优先空行和完整句子
	separators := []string{"\n\n", "。", "！", "？", "\n", "；", ";", "，", ","}
	return recursiveSplit(text, separators, chunkSize)
}

func protectCodeBlocks(text string) (string, []string) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var out []string
	var blocks []string

	inCode := false
	var buf []string

	flushBlock := func() {
		if len(buf) == 0 {
			return
		}
		blocks = append(blocks, strings.Join(buf, "\n"))
		placeholder := "««CODE_BLOCK_" + strconv.Itoa(len(blocks)-1) + "»»"
		out = append(out, placeholder)
		buf = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if !inCode {
				inCode = true
				buf = append(buf, line)
				continue
			}
			buf = append(buf, line)
			inCode = false
			flushBlock()
			continue
		}

		if inCode {
			buf = append(buf, line)
			continue
		}
		out = append(out, line)
	}

	if inCode {
		flushBlock()
	}

	return strings.Join(out, "\n"), blocks
}

func restoreCodeBlocks(text string, blocks []string) string {
	restored := text
	for i, b := range blocks {
		placeholder := "««CODE_BLOCK_" + strconv.Itoa(i) + "»»"
		restored = strings.ReplaceAll(restored, placeholder, b)
	}
	return restored
}
