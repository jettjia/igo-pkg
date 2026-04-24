package split

import (
	"context"
	"math"
	"strings"
	"unicode"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

// SemanticStrategy 语义分块策略
//
// - 无标题文本：按句子拆分，基于 TF-IDF 计算余弦相似度
// - 相似度 >= Threshold 的相邻句子优先合并为同一块
// - 单句超过 chunk_size 时：优先按语义短语拆分（逗号/分号/顿号/空格），再兜底硬切
type SemanticStrategy struct {
	StrategyBase
	Threshold float64
	Mode      SemanticMode

	BufferSize           int
	BreakpointPercentile int
	MinChunkSize         int

	InitialThreshold   float64
	AppendingThreshold float64
	MergingThreshold   float64
	MaxChunkSize       int
	MergingRange       int
	MergingSeparator   string
}

type SemanticMode string

const (
	SemanticModeGreedy           SemanticMode = "greedy"
	SemanticModeWindowBreakpoint SemanticMode = "window_breakpoint"
	SemanticModeDoubleMerging    SemanticMode = "double_merging"
)

func NewSemanticStrategy() *SemanticStrategy {
	return &SemanticStrategy{
		StrategyBase: StrategyBase{
			ChunkSize:           500,
			OverlapRatio:        0.1,
			RemoveURLAndEmail:   false,
			NormalizeWhitespace: false,
			TrimSpace:           false,
		},
		Threshold: 0.5,
		Mode:      SemanticModeGreedy,

		BufferSize:           1,
		BreakpointPercentile: 95,
		MinChunkSize:         0,

		InitialThreshold:   0.2,
		AppendingThreshold: 0.3,
		MergingThreshold:   0.3,
		MaxChunkSize:       0,
		MergingRange:       1,
		MergingSeparator:   "",
	}
}

func (s *SemanticStrategy) GetType() StrategyType {
	return StrategyTypeSemantic
}

func (s *SemanticStrategy) Validate() error {
	if err := s.validateBase(); err != nil {
		return err
	}
	if s.Threshold == 0 {
		s.Threshold = 0.5
	}
	if s.Threshold < 0 || s.Threshold > 1 {
		s.Threshold = 0.5
	}
	if s.Mode == "" {
		s.Mode = SemanticModeGreedy
	}
	switch s.Mode {
	case SemanticModeGreedy, SemanticModeWindowBreakpoint, SemanticModeDoubleMerging:
	default:
		s.Mode = SemanticModeGreedy
	}

	if s.BufferSize < 0 {
		s.BufferSize = 0
	}
	if s.BreakpointPercentile == 0 {
		s.BreakpointPercentile = 95
	}
	if s.BreakpointPercentile < 50 {
		s.BreakpointPercentile = 50
	}
	if s.BreakpointPercentile > 99 {
		s.BreakpointPercentile = 99
	}
	if s.MinChunkSize < 0 {
		s.MinChunkSize = 0
	}

	if s.InitialThreshold <= 0 {
		s.InitialThreshold = 0.2
	}
	if s.AppendingThreshold <= 0 {
		s.AppendingThreshold = 0.3
	}
	if s.MergingThreshold <= 0 {
		s.MergingThreshold = 0.3
	}
	if s.InitialThreshold < 0 || s.InitialThreshold > 1 {
		s.InitialThreshold = 0.2
	}
	if s.AppendingThreshold < 0 || s.AppendingThreshold > 1 {
		s.AppendingThreshold = 0.3
	}
	if s.MergingThreshold < 0 || s.MergingThreshold > 1 {
		s.MergingThreshold = 0.3
	}
	if s.MaxChunkSize <= 0 {
		s.MaxChunkSize = s.ChunkSize
	}
	if s.MaxChunkSize <= 0 {
		s.MaxChunkSize = 500
	}
	if s.MergingRange <= 0 {
		s.MergingRange = 1
	}
	if s.MergingRange > 2 {
		s.MergingRange = 2
	}
	return nil
}

func (s *SemanticStrategy) Split(ctx context.Context, text string, fileName string) ([]*Chunk, error) {
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

	sentences := splitIntoSentencesPreserveNewlinesWithCodeBlocks(processed)
	if len(sentences) == 0 {
		return nil, nil
	}

	var chunks []string
	var err error
	switch s.Mode {
	case SemanticModeWindowBreakpoint:
		chunks, err = semanticSplitWindowBreakpoint(ctx, sentences, s.ChunkSize, s.BufferSize, s.BreakpointPercentile, s.MinChunkSize, s.OverlapRatio)
	case SemanticModeDoubleMerging:
		chunks, err = semanticSplitDoubleMerging(
			ctx,
			sentences,
			s.ChunkSize,
			s.MaxChunkSize,
			s.MinChunkSize,
			s.InitialThreshold,
			s.AppendingThreshold,
			s.MergingThreshold,
			s.MergingRange,
			s.MergingSeparator,
		)
	default:
		chunks, err = semanticSplitWithSentenceOverlap(ctx, sentences, s.ChunkSize, s.Threshold, s.OverlapRatio, s.TrimSpace)
	}
	if err != nil {
		return nil, err
	}

	docs := make([]*schema.Document, 0, len(chunks))
	for _, c := range chunks {
		c = applyTrimSpaceIfNeeded(c, &s.StrategyBase)
		if c == "" {
			continue
		}
		docs = append(docs, newDocument(c, "", 0))
	}
	return convertToChunks(docs, fileName, processed, &s.StrategyBase), nil
}

func semanticSplitDoubleMerging(
	ctx context.Context,
	sentences []string,
	chunkSize int,
	maxChunkSize int,
	minChunkSize int,
	initialThreshold float64,
	appendingThreshold float64,
	mergingThreshold float64,
	mergingRange int,
	mergingSeparator string,
) ([]string, error) {
	if len(sentences) == 0 {
		return nil, nil
	}
	if maxChunkSize <= 0 {
		maxChunkSize = chunkSize
	}
	if maxChunkSize <= 0 {
		maxChunkSize = 500
	}
	if minChunkSize < 0 {
		minChunkSize = 0
	}
	if mergingRange <= 0 {
		mergingRange = 1
	}
	if mergingRange > 2 {
		mergingRange = 2
	}

	cache := newTfidfCache(len(sentences) * 2)

	precomputedVectors := precomputeTfidfVectors(sentences)
	for text, vec := range precomputedVectors {
		cache.setPrecomputed(text, vec)
	}

	getVector := func(text string) sparseVector {
		if vec, ok := cache.get(text); ok {
			return vec
		}
		vecs := buildTfidfVectors([]string{text})
		if len(vecs) > 0 {
			cache.put(text, vecs[0])
			return vecs[0]
		}
		return sparseVector{}
	}

	similarity := func(a, b string) float64 {
		vecA := getVector(a)
		vecB := getVector(b)
		if len(vecA) == 0 || len(vecB) == 0 {
			return 0
		}
		return cosineSimilarity(vecA, vecB)
	}

	sepLen := runeLen(mergingSeparator)

	var initialChunks []string
	chunk := sentences[0]
	newChunk := true
	chunkSentences := []string{sentences[0]}
	lastSentences := sentences[0]

	flushChunk := func() {
		if strings.TrimSpace(chunk) == "" {
			chunk = ""
			chunkSentences = nil
			newChunk = true
			lastSentences = ""
			return
		}
		initialChunks = append(initialChunks, chunk)
		chunk = ""
		chunkSentences = nil
		newChunk = true
		lastSentences = ""
	}

	for _, sentence := range sentences[1:] {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if strings.TrimSpace(sentence) == "" {
			continue
		}

		if runeLen(sentence) > maxChunkSize {
			flushChunk()
			if isCodeBlockSentence(sentence) {
				initialChunks = append(initialChunks, sentence)
			} else {
				initialChunks = append(initialChunks, splitLongSentence(sentence, maxChunkSize)...)
			}
			continue
		}

		if chunk == "" {
			chunk = sentence
			chunkSentences = []string{sentence}
			newChunk = true
			lastSentences = sentence
			continue
		}

		if newChunk {
			if similarity(chunk, sentence) < initialThreshold && runeLen(chunk) >= minChunkSize && runeLen(chunk)+sepLen+runeLen(sentence) <= maxChunkSize {
				initialChunks = append(initialChunks, chunk)
				chunk = sentence
				chunkSentences = []string{sentence}
				newChunk = true
				lastSentences = sentence
				continue
			}

			chunkSentences = []string{chunk}
			if runeLen(chunk)+sepLen+runeLen(sentence) <= maxChunkSize {
				chunkSentences = append(chunkSentences, sentence)
				chunk = strings.Join(chunkSentences, mergingSeparator)
				newChunk = false
			} else {
				initialChunks = append(initialChunks, chunk)
				chunk = sentence
				chunkSentences = []string{sentence}
				newChunk = true
				lastSentences = sentence
				continue
			}

			if len(chunkSentences) >= 2 {
				lastSentences = strings.Join(chunkSentences[len(chunkSentences)-2:], mergingSeparator)
			} else {
				lastSentences = chunkSentences[0]
			}
			continue
		}

		if (similarity(lastSentences, sentence) > appendingThreshold || runeLen(chunk) < minChunkSize) && runeLen(chunk)+sepLen+runeLen(sentence) <= maxChunkSize {
			chunkSentences = append(chunkSentences, sentence)
			if len(chunkSentences) >= 2 {
				lastSentences = strings.Join(chunkSentences[len(chunkSentences)-2:], mergingSeparator)
			} else {
				lastSentences = chunkSentences[0]
			}
			if mergingSeparator == "" {
				chunk += sentence
			} else {
				chunk += mergingSeparator + sentence
			}
			continue
		}

		initialChunks = append(initialChunks, chunk)
		chunk = sentence
		chunkSentences = []string{sentence}
		newChunk = true
		lastSentences = sentence
	}
	if strings.TrimSpace(chunk) != "" {
		initialChunks = append(initialChunks, chunk)
	}

	merged := make([]string, 0, len(initialChunks))
	skip := 0
	current := initialChunks[0]
	for i := 1; i < len(initialChunks); i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if skip > 0 {
			skip--
			continue
		}

		next := initialChunks[i]
		if minChunkSize > 0 && runeLen(current) < minChunkSize {
			cand := current
			if mergingSeparator == "" {
				cand += next
			} else {
				cand += mergingSeparator + next
			}
			if runeLen(cand) <= maxChunkSize {
				current = cand
				continue
			}
		}
		if runeLen(current) >= maxChunkSize {
			merged = append(merged, current)
			current = next
			continue
		}

		canMerge := func(a, b string) bool {
			if similarity(a, b) <= mergingThreshold {
				return false
			}
			return runeLen(a)+sepLen+runeLen(b) <= maxChunkSize
		}

		if canMerge(current, next) {
			if mergingSeparator == "" {
				current += next
			} else {
				current += mergingSeparator + next
			}
			continue
		}

		if i <= len(initialChunks)-2 && canMerge(current, initialChunks[i+1]) {
			cand := current
			if mergingSeparator == "" {
				cand += next + initialChunks[i+1]
			} else {
				cand += mergingSeparator + next + mergingSeparator + initialChunks[i+1]
			}
			if runeLen(cand) <= maxChunkSize {
				current = cand
				skip = 1
				continue
			}
		}

		if mergingRange == 2 && i < len(initialChunks)-2 && canMerge(current, initialChunks[i+2]) {
			cand := current
			if mergingSeparator == "" {
				cand += next + initialChunks[i+1] + initialChunks[i+2]
			} else {
				cand += mergingSeparator + next + mergingSeparator + initialChunks[i+1] + mergingSeparator + initialChunks[i+2]
			}
			if runeLen(cand) <= maxChunkSize {
				current = cand
				skip = 2
				continue
			}
		}

		merged = append(merged, current)
		current = next
	}
	merged = append(merged, current)

	final := make([]string, 0, len(merged))
	for _, m := range merged {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if chunkSize > 0 && runeLen(m) > chunkSize {
			final = append(final, splitLongSentence(m, chunkSize)...)
			continue
		}
		final = append(final, m)
	}
	return final, nil
}

func semanticSplitWindowBreakpoint(
	ctx context.Context,
	sentences []string,
	chunkSize int,
	bufferSize int,
	percentile int,
	minChunkSize int,
	overlapRatio float64,
) ([]string, error) {
	if len(sentences) == 0 {
		return nil, nil
	}

	windowTexts := make([]string, len(sentences))
	for i := 0; i < len(sentences); i++ {
		start := i - bufferSize
		if start < 0 {
			start = 0
		}
		end := i + bufferSize + 1
		if end > len(sentences) {
			end = len(sentences)
		}
		windowTexts[i] = strings.Join(sentences[start:end], "")
	}

	vectors := buildTfidfVectors(windowTexts)
	if len(vectors) <= 1 {
		return []string{strings.Join(sentences, "")}, nil
	}

	distances := make([]float64, 0, len(vectors)-1)
	for i := 0; i < len(vectors)-1; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		sim := cosineSimilarity(vectors[i], vectors[i+1])
		d := 1 - sim
		if d < 0 {
			d = 0
		}
		if d > 1 {
			d = 1
		}
		distances = append(distances, d)
	}

	threshold := percentileValue(distances, percentile)
	breakpoints := make([]int, 0, len(distances))
	for i, d := range distances {
		if d > threshold {
			breakpoints = append(breakpoints, i)
		}
	}

	groups := make([][]string, 0, len(breakpoints)+1)
	start := 0
	for _, bp := range breakpoints {
		end := bp + 1
		if end <= start {
			continue
		}
		groups = append(groups, sentences[start:end])
		start = end
	}
	if start < len(sentences) {
		groups = append(groups, sentences[start:])
	}

	groups = mergeSmallSemanticGroups(groups, minChunkSize, chunkSize)
	groups = splitGroupsByMaxSize(groups, chunkSize)
	groups = applySentenceOverlapToGroups(groups, overlapTokens(chunkSize, overlapRatio), chunkSize)

	out := make([]string, 0, len(groups))
	for _, g := range groups {
		if len(g) == 0 {
			continue
		}
		out = append(out, strings.Join(g, ""))
	}
	return out, nil
}

func percentileValue(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}
	if percentile <= 0 {
		percentile = 1
	}
	if percentile >= 100 {
		percentile = 99
	}
	sorted := append([]float64{}, values...)
	for i := 1; i < len(sorted); i++ {
		j := i
		for j > 0 && sorted[j-1] > sorted[j] {
			sorted[j-1], sorted[j] = sorted[j], sorted[j-1]
			j--
		}
	}

	if len(sorted) == 1 {
		return sorted[0]
	}
	pos := int(math.Ceil((float64(percentile) / 100) * float64(len(sorted)-1)))
	if pos < 0 {
		pos = 0
	}
	if pos >= len(sorted) {
		pos = len(sorted) - 1
	}
	return sorted[pos]
}

func mergeSmallSemanticGroups(groups [][]string, minChunkSize int, chunkSize int) [][]string {
	if len(groups) == 0 {
		return nil
	}
	if minChunkSize <= 0 {
		return groups
	}

	merged := make([][]string, 0, len(groups))
	i := 0
	for i < len(groups) {
		g := groups[i]
		if len(g) == 1 && isMarkdownHeadingLine(g[0]) && i+1 < len(groups) {
			next := groups[i+1]
			combined := append(append([]string{}, g...), next...)
			if sentenceGroupLen(combined) <= chunkSize {
				merged = append(merged, combined)
				i += 2
				continue
			}
		}

		if sentenceGroupLen(g) >= minChunkSize || i == len(groups)-1 {
			merged = append(merged, g)
			i++
			continue
		}

		if i+1 < len(groups) {
			next := groups[i+1]
			combined := append(append([]string{}, g...), next...)
			if sentenceGroupLen(combined) <= chunkSize {
				merged = append(merged, combined)
				i += 2
				continue
			}
		}

		if len(merged) > 0 {
			prev := merged[len(merged)-1]
			combined := append(append([]string{}, prev...), g...)
			if sentenceGroupLen(combined) <= chunkSize {
				merged[len(merged)-1] = combined
				i++
				continue
			}
		}

		merged = append(merged, g)
		i++
	}
	return merged
}

func applySentenceOverlapToGroups(groups [][]string, overlapLimit int, chunkSize int) [][]string {
	if len(groups) <= 1 || overlapLimit <= 0 {
		return groups
	}
	out := make([][]string, 0, len(groups))
	out = append(out, groups[0])

	for i := 1; i < len(groups); i++ {
		prev := out[i-1]
		cur := groups[i]

		tail := tailSentencesByLen(prev, overlapLimit)
		merged := append(append([]string{}, tail...), cur...)
		if sentenceGroupLen(merged) > chunkSize && len(tail) > 0 {
			merged = cur
		}
		out = append(out, merged)
	}
	return out
}

func tailSentencesByLen(sentences []string, limit int) []string {
	if limit <= 0 || len(sentences) == 0 {
		return nil
	}
	sum := 0
	var tail []string
	for i := len(sentences) - 1; i >= 0; i-- {
		s := sentences[i]
		l := runeLen(s)
		if sum+l > limit && sum > 0 {
			break
		}
		tail = append([]string{s}, tail...)
		sum += l
		if sum >= limit {
			break
		}
	}
	return tail
}

func sentenceGroupLen(sentences []string) int {
	n := 0
	for _, s := range sentences {
		n += runeLen(s)
	}
	return n
}

func splitGroupsByMaxSize(groups [][]string, chunkSize int) [][]string {
	if len(groups) == 0 || chunkSize <= 0 {
		return groups
	}
	out := make([][]string, 0, len(groups))
	for _, g := range groups {
		if len(g) == 0 {
			continue
		}
		if sentenceGroupLen(g) <= chunkSize {
			out = append(out, g)
			continue
		}

		var cur []string
		curLen := 0
		flush := func() {
			if len(cur) == 0 {
				return
			}
			out = append(out, cur)
			cur = nil
			curLen = 0
		}

		for _, s := range g {
			l := runeLen(s)
			if curLen+l > chunkSize && curLen > 0 {
				flush()
			}
			if l > chunkSize {
				flush()
				parts := splitLongSentence(s, chunkSize)
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					out = append(out, []string{p})
				}
				continue
			}
			cur = append(cur, s)
			curLen += l
		}
		flush()
	}
	return out
}

func semanticSplitWithSentenceOverlap(ctx context.Context, sentences []string, chunkSize int, threshold float64, overlapRatio float64, trimSpace bool) ([]string, error) {
	if len(sentences) == 0 {
		return nil, nil
	}

	vectors := buildTfidfVectors(sentences)
	var chunks []string
	var cur string
	var curSentences []string
	var lastVec sparseVector
	hasLastVec := false
	var carry []string

	flush := func() {
		if strings.TrimSpace(cur) == "" {
			cur = ""
			curSentences = nil
			hasLastVec = false
			lastVec = nil
			carry = nil
			return
		}

		if cur != "" {
			chunks = append(chunks, cur)
		}

		overlapLimit := overlapTokens(chunkSize, overlapRatio)
		if overlapLimit > 0 && len(curSentences) > 0 {
			var tail []string
			sum := 0
			for i := len(curSentences) - 1; i >= 0; i-- {
				s := curSentences[i]
				l := runeLen(s)
				if sum+l > overlapLimit && sum > 0 {
					break
				}
				tail = append([]string{s}, tail...)
				sum += l
				if sum >= overlapLimit {
					break
				}
			}
			carry = tail
		} else {
			carry = nil
		}

		cur = ""
		curSentences = nil
		hasLastVec = false
		lastVec = nil
	}

	for i, sentence := range sentences {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if strings.TrimSpace(sentence) == "" {
			continue
		}

		if isMarkdownHeadingLine(sentence) {
			flush()
		}

		if runeLen(sentence) > chunkSize {
			flush()
			longParts := splitLongSentence(sentence, chunkSize)
			chunks = append(chunks, longParts...)
			continue
		}

		if cur == "" {
			if len(carry) > 0 {
				prefix := strings.Join(carry, "")
				if runeLen(prefix)+runeLen(sentence) <= chunkSize {
					cur = prefix + sentence
					curSentences = append(append([]string{}, carry...), sentence)
					carry = nil
					lastVec = vectors[i]
					hasLastVec = true
					continue
				}
				carry = nil
			}

			cur = sentence
			curSentences = []string{sentence}
			lastVec = vectors[i]
			hasLastVec = true
			continue
		}

		canAppendWithoutSim := false
		if len(curSentences) == 1 && isMarkdownHeadingLine(curSentences[0]) {
			canAppendWithoutSim = true
		}
		if isCodeBlockSentence(sentence) {
			canAppendWithoutSim = true
		}

		sim := 0.0
		if hasLastVec {
			sim = cosineSimilarity(lastVec, vectors[i])
		}
		if (canAppendWithoutSim || sim >= threshold) && runeLen(cur)+runeLen(sentence) <= chunkSize {
			cur += sentence
			curSentences = append(curSentences, sentence)
			lastVec = vectors[i]
			hasLastVec = true
			continue
		}

		flush()
		cur = sentence
		curSentences = []string{sentence}
		lastVec = vectors[i]
		hasLastVec = true
	}
	flush()

	return chunks, nil
}

func isMarkdownHeadingLine(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "#") {
		return true
	}
	return false
}

func isCodeBlockSentence(s string) bool {
	trimmed := strings.TrimSpace(s)
	return strings.HasPrefix(trimmed, "```")
}

func splitIntoSentencesPreserveNewlinesWithCodeBlocks(text string) []string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	endMarks := map[rune]struct{}{
		'.': {}, '!': {}, '?': {}, '。': {}, '！': {}, '？': {}, ';': {}, '；': {},
	}

	lines := strings.Split(normalized, "\n")
	var out []string

	inCode := false
	var codeBuf []string

	var cur strings.Builder
	flushCur := func(withNewline bool) {
		if strings.TrimSpace(cur.String()) == "" {
			cur.Reset()
			return
		}
		if withNewline {
			out = append(out, cur.String()+"\n")
		} else {
			out = append(out, cur.String())
		}
		cur.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			flushCur(true)
			if !inCode {
				inCode = true
				codeBuf = append(codeBuf, line)
				continue
			}
			codeBuf = append(codeBuf, line)
			inCode = false
			out = append(out, strings.Join(codeBuf, "\n")+"\n")
			codeBuf = nil
			continue
		}

		if inCode {
			codeBuf = append(codeBuf, line)
			continue
		}

		if trimmed == "" {
			flushCur(true)
			continue
		}

		if isMarkdownHeadingLine(trimmed) || isListLine(trimmed) {
			flushCur(true)
			out = append(out, line+"\n")
			continue
		}

		lineRunes := []rune(line)
		for idx, r := range lineRunes {
			cur.WriteRune(r)
			if _, ok := endMarks[r]; ok {
				flushCur(idx == len(lineRunes)-1)
			}
		}
		flushCur(true)
	}

	if inCode && len(codeBuf) > 0 {
		out = append(out, strings.Join(codeBuf, "\n")+"\n")
	}

	return out
}

func isListLine(trimmed string) bool {
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return true
	}
	i := 0
	for i < len(trimmed) {
		c := trimmed[i]
		if c < '0' || c > '9' {
			break
		}
		i++
	}
	if i == 0 {
		return false
	}
	if i < len(trimmed) && (trimmed[i] == '.' || trimmed[i] == ')') {
		return true
	}
	return false
}

func splitLongSentence(sentence string, chunkSize int) []string {
	if sentence == "" || chunkSize <= 0 {
		return nil
	}
	if runeLen(sentence) <= chunkSize {
		return []string{sentence}
	}

	separators := []string{"，", ",", "；", ";", "、", " "}
	parts := recursiveSplit(sentence, separators, chunkSize)
	var filtered []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if runeLen(p) > chunkSize {
			filtered = append(filtered, hardSplitByRunes(p, chunkSize)...)
		} else {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

type sparseVector map[string]float64

type tfidfCache struct {
	cache       map[string]sparseVector
	maxSize     int
	order       []string
	precomputed map[string]sparseVector
}

func newTfidfCache(maxSize int) *tfidfCache {
	if maxSize <= 0 {
		maxSize = 1024
	}
	return &tfidfCache{
		cache:       make(map[string]sparseVector, maxSize),
		maxSize:     maxSize,
		order:       make([]string, 0, maxSize),
		precomputed: make(map[string]sparseVector),
	}
}

func (c *tfidfCache) get(text string) (sparseVector, bool) {
	if vec, ok := c.cache[text]; ok {
		return vec, true
	}
	if vec, ok := c.precomputed[text]; ok {
		c.cache[text] = vec
		return vec, true
	}
	return nil, false
}

func (c *tfidfCache) put(text string, vec sparseVector) {
	if len(c.order) >= c.maxSize {
		old := c.order[0]
		delete(c.cache, old)
		c.order = c.order[1:]
	}
	c.cache[text] = vec
	c.order = append(c.order, text)
}

func (c *tfidfCache) setPrecomputed(text string, vec sparseVector) {
	c.precomputed[text] = vec
}

func precomputeTfidfVectors(sentences []string) map[string]sparseVector {
	if len(sentences) == 0 {
		return nil
	}

	type docTokens struct {
		counts map[string]int
		total  int
	}

	docs := make([]docTokens, 0, len(sentences))
	df := make(map[string]int, 1024)

	for _, s := range sentences {
		tokens := tokenizeForTfidf(s)
		counts := make(map[string]int, len(tokens))
		seen := make(map[string]struct{}, len(tokens))
		for _, tok := range tokens {
			if tok == "" {
				continue
			}
			counts[tok]++
			if _, ok := seen[tok]; !ok {
				seen[tok] = struct{}{}
				df[tok]++
			}
		}
		docs = append(docs, docTokens{
			counts: counts,
			total:  len(tokens),
		})
	}

	N := float64(len(docs))
	idf := make(map[string]float64, len(df))
	for term, freq := range df {
		idf[term] = math.Log((N+1)/(float64(freq)+1)) + 1
	}

	result := make(map[string]sparseVector, len(sentences))
	for i, s := range sentences {
		d := docs[i]
		vec := make(sparseVector, len(d.counts))
		if d.total > 0 {
			total := float64(d.total)
			for term, cnt := range d.counts {
				tf := float64(cnt) / total
				vec[term] = tf * idf[term]
			}
		}
		result[s] = vec
	}

	return result
}

func buildTfidfVectors(sentences []string) []sparseVector {
	if len(sentences) == 0 {
		return nil
	}

	type docTokens struct {
		tokens []string
		counts map[string]int
		total  int
	}

	docs := make([]docTokens, 0, len(sentences))
	df := make(map[string]int, 1024)

	for _, s := range sentences {
		tokens := tokenizeForTfidf(s)
		counts := make(map[string]int, len(tokens))
		seen := make(map[string]struct{}, len(tokens))
		for _, tok := range tokens {
			if tok == "" {
				continue
			}
			counts[tok]++
			if _, ok := seen[tok]; !ok {
				seen[tok] = struct{}{}
				df[tok]++
			}
		}
		docs = append(docs, docTokens{
			tokens: tokens,
			counts: counts,
			total:  len(tokens),
		})
	}

	N := float64(len(docs))
	idf := make(map[string]float64, len(df))
	for term, freq := range df {
		idf[term] = math.Log((N+1)/(float64(freq)+1)) + 1
	}

	vectors := make([]sparseVector, 0, len(docs))
	for _, d := range docs {
		vec := make(sparseVector, len(d.counts))
		if d.total == 0 {
			vectors = append(vectors, vec)
			continue
		}
		total := float64(d.total)
		for term, cnt := range d.counts {
			tf := float64(cnt) / total
			vec[term] = tf * idf[term]
		}
		vectors = append(vectors, vec)
	}
	return vectors
}

func tokenizeForTfidf(s string) []string {
	var tokens []string
	var buf []rune
	flush := func() {
		if len(buf) == 0 {
			return
		}
		tokens = append(tokens, strings.ToLower(string(buf)))
		buf = nil
	}

	for _, r := range []rune(s) {
		if isCJK(r) {
			flush()
			tokens = append(tokens, string(r))
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf = append(buf, unicode.ToLower(r))
			continue
		}
		flush()
	}
	flush()
	return tokens
}

func isCJK(r rune) bool {
	switch {
	case r >= 0x4E00 && r <= 0x9FFF:
		return true
	case r >= 0x3400 && r <= 0x4DBF:
		return true
	case r >= 0x20000 && r <= 0x2A6DF:
		return true
	default:
		return false
	}
}

func cosineSimilarity(a, b sparseVector) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	var small, large sparseVector
	if len(a) <= len(b) {
		small, large = a, b
	} else {
		small, large = b, a
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for _, v := range a {
		normA += v * v
	}
	for _, v := range b {
		normB += v * v
	}

	for k, v := range small {
		if w, ok := large[k]; ok {
			dot += v * w
		}
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
