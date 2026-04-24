package split

import (
	"context"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

// FixedSizeStrategy 固定字符大小分块策略
//
// - chunk_size: 单个块最大字符数
// - chunk_overlap: 相邻块重叠字符数（优先使用该值）
// - overlap_ratio: 当 chunk_overlap 未设置时生效，范围 0.1~0.2
type FixedSizeStrategy struct {
	StrategyBase
	ChunkOverlap int
}

func NewFixedSizeStrategy() *FixedSizeStrategy {
	return &FixedSizeStrategy{
		StrategyBase: StrategyBase{
			ChunkSize:           500,
			OverlapRatio:        0.1,
			RemoveURLAndEmail:   false,
			NormalizeWhitespace: false,
			TrimSpace:           false,
		},
		ChunkOverlap: 0,
	}
}

func (s *FixedSizeStrategy) GetType() StrategyType {
	return StrategyTypeFixedSize
}

func (s *FixedSizeStrategy) Validate() error {
	if err := s.validateBase(); err != nil {
		return err
	}
	if s.ChunkOverlap < 0 {
		s.ChunkOverlap = 0
	}
	if s.ChunkOverlap >= s.ChunkSize && s.ChunkSize > 0 {
		s.ChunkOverlap = s.ChunkSize - 1
	}
	return nil
}

func (s *FixedSizeStrategy) Split(ctx context.Context, text string, fileName string) ([]*Chunk, error) {
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

	chunkSize := s.ChunkSize
	overlap := s.ChunkOverlap
	if overlap == 0 {
		overlap = overlapTokens(chunkSize, s.OverlapRatio)
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize && chunkSize > 0 {
		overlap = chunkSize - 1
	}

	runes := []rune(processed)
	step := chunkSize - overlap
	if step <= 0 {
		step = chunkSize
	}

	var docs []*schema.Document
	for start := 0; start < len(runes); start += step {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunkText := string(runes[start:end])
		chunkText = applyTrimSpaceIfNeeded(chunkText, &s.StrategyBase)
		if chunkText != "" {
			docs = append(docs, newDocument(chunkText, "", 0))
		}
		if end == len(runes) {
			break
		}
	}
	return convertToChunks(docs, fileName, processed, &s.StrategyBase), nil
}
