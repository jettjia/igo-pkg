package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jettjia/igo-pkg/aipkg/rag/split"
)

func main() {
	ctx := context.Background()

	fileName := "test-3.md"
	text, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	input := string(text)

	runFixedSize(ctx, input, fileName)
	runRecursiveCharacter(ctx, input, fileName)
	runSemantic(ctx, input, fileName)
	runDocumentStructure(ctx, input, fileName)
}

type resultFile struct {
	Type                 string         `json:"type"`
	FileName             string         `json:"file_name"`
	ChunkSize            int            `json:"chunk_size"`
	OverlapRatio         float64        `json:"overlap_ratio"`
	RemoveURLAndEmail    bool           `json:"remove_url_and_email"`
	NormalizeWhitespace  bool           `json:"normalize_whitespace"`
	TrimSpace            bool           `json:"trim_space"`
	ChunkOverlap         int            `json:"chunk_overlap,omitempty"`
	SemanticThreshold    float64        `json:"semantic_threshold,omitempty"`
	SemanticMode         string         `json:"semantic_mode,omitempty"`
	SemanticBufferSize   int            `json:"semantic_buffer_size,omitempty"`
	SemanticPercentile   int            `json:"semantic_breakpoint_percentile,omitempty"`
	SemanticMinChunk     int            `json:"semantic_min_chunk_size,omitempty"`
	SemanticInitial      float64        `json:"semantic_initial_threshold,omitempty"`
	SemanticAppending    float64        `json:"semantic_appending_threshold,omitempty"`
	SemanticMerging      float64        `json:"semantic_merging_threshold,omitempty"`
	SemanticMaxChunk     int            `json:"semantic_max_chunk_size,omitempty"`
	SemanticMergingRange int            `json:"semantic_merging_range,omitempty"`
	DocumentMaxDepth     int            `json:"document_max_depth,omitempty"`
	SkipEmptyHeadings    bool           `json:"skip_empty_headings,omitempty"`
	Chunks               []*split.Chunk `json:"chunks"`
	TotalChunks          int            `json:"total_chunks"`
	TotalChars           int            `json:"total_chars"`
}

func writeResult(filename string, r resultFile) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filename, b, 0o644); err != nil {
		panic(err)
	}
}

func getTotalChars(chunks []*split.Chunk) int {
	total := 0
	for _, c := range chunks {
		if c != nil {
			total += len([]rune(c.SliceContent.Text))
		}
	}
	return total
}

func runFixedSize(ctx context.Context, input string, fileName string) {
	s := split.NewFixedSizeStrategy()
	s.ChunkSize = 400
	s.ChunkOverlap = 0
	s.OverlapRatio = 0.1
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	docs, err := s.Split(ctx, input, fileName)
	if err != nil {
		panic(err)
	}

	writeResult("fixed_size.json", resultFile{
		Type:                string(s.GetType()),
		FileName:            fileName,
		ChunkSize:           s.ChunkSize,
		OverlapRatio:        s.OverlapRatio,
		RemoveURLAndEmail:   s.RemoveURLAndEmail,
		NormalizeWhitespace: s.NormalizeWhitespace,
		TrimSpace:           s.TrimSpace,
		ChunkOverlap:        s.ChunkOverlap,
		Chunks:              docs,
		TotalChunks:         len(docs),
		TotalChars:          getTotalChars(docs),
	})
	fmt.Printf("FixedSizeStrategy chunks=%d -> fixed_size.json\n", len(docs))
}

func runRecursiveCharacter(ctx context.Context, input string, fileName string) {
	s := split.NewRecursiveCharacterStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.15
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	docs, err := s.Split(ctx, input, fileName)
	if err != nil {
		panic(err)
	}

	writeResult("recursive_character.json", resultFile{
		Type:                string(s.GetType()),
		FileName:            fileName,
		ChunkSize:           s.ChunkSize,
		OverlapRatio:        s.OverlapRatio,
		RemoveURLAndEmail:   s.RemoveURLAndEmail,
		NormalizeWhitespace: s.NormalizeWhitespace,
		TrimSpace:           s.TrimSpace,
		Chunks:              docs,
		TotalChunks:         len(docs),
		TotalChars:          getTotalChars(docs),
	})
	fmt.Printf("RecursiveCharacterStrategy chunks=%d -> recursive_character.json\n", len(docs))
}

func runSemantic(ctx context.Context, input string, fileName string) {
	s := split.NewSemanticStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.Mode = split.SemanticModeDoubleMerging
	s.MaxChunkSize = 400
	s.MinChunkSize = 120
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 2
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	docs, err := s.Split(ctx, input, fileName)
	if err != nil {
		panic(err)
	}

	writeResult("semantic.json", resultFile{
		Type:                 string(s.GetType()),
		FileName:             fileName,
		ChunkSize:            s.ChunkSize,
		OverlapRatio:         s.OverlapRatio,
		RemoveURLAndEmail:    s.RemoveURLAndEmail,
		NormalizeWhitespace:  s.NormalizeWhitespace,
		TrimSpace:            s.TrimSpace,
		SemanticMode:         string(s.Mode),
		SemanticBufferSize:   s.BufferSize,
		SemanticPercentile:   s.BreakpointPercentile,
		SemanticMinChunk:     s.MinChunkSize,
		SemanticInitial:      s.InitialThreshold,
		SemanticAppending:    s.AppendingThreshold,
		SemanticMerging:      s.MergingThreshold,
		SemanticMaxChunk:     s.MaxChunkSize,
		SemanticMergingRange: s.MergingRange,
		Chunks:               docs,
		TotalChunks:          len(docs),
		TotalChars:           getTotalChars(docs),
	})
	fmt.Printf("SemanticStrategy chunks=%d -> semantic.json\n", len(docs))
}

func runDocumentStructure(ctx context.Context, input string, fileName string) {
	s := split.NewDocumentStructureStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.MaxDepth = 3
	s.SemanticThreshold = 0.5
	s.SkipEmptyHeadings = true
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	docs, err := s.Split(ctx, input, fileName)
	if err != nil {
		panic(err)
	}

	writeResult("document_structure.json", resultFile{
		Type:                string(s.GetType()),
		FileName:            fileName,
		ChunkSize:           s.ChunkSize,
		OverlapRatio:        s.OverlapRatio,
		RemoveURLAndEmail:   s.RemoveURLAndEmail,
		NormalizeWhitespace: s.NormalizeWhitespace,
		TrimSpace:           s.TrimSpace,
		SemanticThreshold:   s.SemanticThreshold,
		DocumentMaxDepth:    s.MaxDepth,
		SkipEmptyHeadings:   s.SkipEmptyHeadings,
		Chunks:              docs,
		TotalChunks:         len(docs),
		TotalChars:          getTotalChars(docs),
	})
	fmt.Printf("DocumentStructureStrategy chunks=%d -> document_structure.json\n", len(docs))
}
