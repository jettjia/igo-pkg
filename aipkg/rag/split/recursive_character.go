package split

import (
	"context"
	"strings"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

// RecursiveCharacterStrategy 递归字符分块策略
//
// 分隔符优先级：
// 空行(\n\n) > 换行(\n) > 句号(。) > 感叹号(!/！) > 问号(？/?) > 分号(；/;) > 逗号(，/,) > 空格
// 无合适分隔符时，按字符数硬切。
type RecursiveCharacterStrategy struct {
	StrategyBase
	Separators []string
}

func NewRecursiveCharacterStrategy() *RecursiveCharacterStrategy {
	return &RecursiveCharacterStrategy{
		StrategyBase: StrategyBase{
			ChunkSize:           500,
			OverlapRatio:        0.1,
			RemoveURLAndEmail:   false,
			NormalizeWhitespace: false,
			TrimSpace:           false,
		},
		Separators: []string{"\n\n", "\n", "。", "！", "!", "？", "?", "；", ";", "，", ",", " "},
	}
}

func (s *RecursiveCharacterStrategy) GetType() StrategyType {
	return StrategyTypeRecursiveChar
}

func (s *RecursiveCharacterStrategy) Validate() error {
	if err := s.validateBase(); err != nil {
		return err
	}
	if len(s.Separators) == 0 {
		s.Separators = []string{"\n\n", "\n", "。", "！", "!", "？", "?", "；", ";", "，", ",", " "}
	}
	return nil
}

func (s *RecursiveCharacterStrategy) Split(ctx context.Context, text string, fileName string) ([]*Chunk, error) {
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

	chunks := recursiveSplit(processed, s.Separators, s.ChunkSize)
	chunks = applyOverlapToStrings(chunks, s.ChunkSize, s.OverlapRatio)

	docs := make([]*schema.Document, 0, len(chunks))
	for _, c := range chunks {
		c = applyTrimSpaceIfNeeded(c, &s.StrategyBase)
		if c == "" {
			continue
		}
		docs = append(docs, newDocument(c, "", 0))
	}
	// 传入 processed 而不是原始 text，这样 convertToChunks 可以正确提取页码等信息
	return convertToChunks(docs, fileName, processed, &s.StrategyBase), nil
}

func recursiveSplit(text string, separators []string, chunkSize int) []string {
	if text == "" {
		return nil
	}
	if chunkSize <= 0 {
		return []string{text}
	}
	if runeLen(text) <= chunkSize {
		return []string{text}
	}
	if len(separators) == 0 {
		return hardSplitByRunes(text, chunkSize)
	}

	sep := separators[0]
	var parts []string
	if sep == "" {
		parts = []string{text}
	} else {
		parts = strings.Split(text, sep)
	}

	if len(parts) == 1 {
		return recursiveSplit(text, separators[1:], chunkSize)
	}

	refined := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if runeLen(p) > chunkSize {
			refined = append(refined, recursiveSplit(p, separators[1:], chunkSize)...)
		} else {
			refined = append(refined, p)
		}
	}

	merged := mergeParts(refined, sep, chunkSize)
	if len(merged) == 0 {
		return recursiveSplit(text, separators[1:], chunkSize)
	}
	return merged
}

func mergeParts(parts []string, sep string, chunkSize int) []string {
	if len(parts) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		return []string{strings.Join(parts, sep)}
	}

	var chunks []string
	i := 0
	for i < len(parts) {
		var builder []string
		currentSize := 0
		j := i
		for j < len(parts) {
			p := parts[j]
			addSize := runeLen(p)
			if len(builder) > 0 && sep != "" {
				addSize += runeLen(sep)
			}
			if currentSize+addSize > chunkSize && currentSize > 0 {
				break
			}
			builder = append(builder, p)
			currentSize += addSize
			j++
		}

		if len(builder) == 1 && runeLen(builder[0]) > chunkSize {
			chunks = append(chunks, hardSplitByRunes(builder[0], chunkSize)...)
			i++
			continue
		}

		chunk := strings.Join(builder, sep)
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		i += len(builder)
	}

	return chunks
}

func hardSplitByRunes(text string, chunkSize int) []string {
	if text == "" || chunkSize <= 0 {
		return nil
	}
	runes := []rune(text)
	if len(runes) <= chunkSize {
		return []string{text}
	}
	var out []string
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}
