package split

import (
	"context"
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

	root := parseHeadingTree(processed, s.MaxDepth)
	if len(root.children) == 0 {
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
	return convertToChunks(docs, fileName, processed, &s.StrategyBase), nil
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
	if prefixLen >= s.ChunkSize {
		truncated := prefixText
		if s.ChunkSize > 0 {
			r := []rune(prefixText)
			if len(r) > s.ChunkSize {
				truncated = string(r[:s.ChunkSize])
			}
		}
		return []*schema.Document{newDocument(applyTrimSpaceIfNeeded(truncated, &s.StrategyBase), docTitle, sec.depth)}
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
		docs = append(docs, newDocument(chunkText, docTitle, sec.depth))
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

			parts := splitTextForStructure(unitText, available)
			for _, p := range parts {
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

	protected, blocks := protectCodeBlocks(trimmed)
	parts := recursiveSplit(protected, []string{"\n\n", "\n", "。", "！", "!", "？", "?", "；", ";", "，", ",", " "}, chunkSize)

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
