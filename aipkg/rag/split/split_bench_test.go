package split

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func loadBenchmarkInput(tb testing.TB) string {
	tb.Helper()

	paths := []string{
		filepath.Join("..", "..", "..", "example", "aipkg", "rag", "split", "test-3.md"),
		filepath.Join("..", "..", "..", "example", "aipkg", "rag", "split", "test-2.md"),
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil && len(b) > 0 {
			return string(b)
		}
	}
	return "第一段。\n\n第二段很长很长很长很长很长很长，包含逗号，继续。\n第三行。"
}

func benchmarkSplit(b *testing.B, s Splitter, fileName string) {
	ctx := context.Background()
	input := loadBenchmarkInput(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Split(ctx, input, fileName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSplit_FixedSize(b *testing.B) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	benchmarkSplit(b, s, "bench.md")
}

func BenchmarkSplit_RecursiveCharacter(b *testing.B) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.15
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	benchmarkSplit(b, s, "bench.md")
}

func BenchmarkSplit_Semantic(b *testing.B) {
	s := NewSemanticStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.Mode = SemanticModeDoubleMerging
	s.MaxChunkSize = 400
	s.MinChunkSize = 120
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 2
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	benchmarkSplit(b, s, "bench.md")
}

func BenchmarkSplit_DocumentStructure(b *testing.B) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.MaxDepth = 3
	s.SemanticThreshold = 0.5
	s.SkipEmptyHeadings = true
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	benchmarkSplit(b, s, "bench.md")
}
