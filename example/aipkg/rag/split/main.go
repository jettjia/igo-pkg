package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

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

	if os.Getenv("SPLIT_PERF_REPORT") == "1" {
		report := runPerfReport(ctx, input, fileName)
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile("split_perf_report.json", b, 0o644); err != nil {
			panic(err)
		}
		fmt.Println(string(b))
		return
	}

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

type perfReport struct {
	Meta     perfMeta        `json:"meta"`
	Items    []perfItem      `json:"items"`
	Summary  perfSummary     `json:"summary"`
	Snapshot runtimeSnapshot `json:"snapshot"`
}

type perfMeta struct {
	FileName     string `json:"file_name"`
	InputBytes   int    `json:"input_bytes"`
	Iterations   int    `json:"iterations"`
	GoVersion    string `json:"go_version"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	CPUCount     int    `json:"cpu_count"`
	GOMAXPROCS   int    `json:"gomaxprocs"`
	BuildVersion string `json:"build_version"`
}

type runtimeSnapshot struct {
	HeapAllocBytes uint64 `json:"heap_alloc_bytes"`
	HeapSysBytes   uint64 `json:"heap_sys_bytes"`
	NumGC          uint32 `json:"num_gc"`
	PauseTotalNs   uint64 `json:"pause_total_ns"`
}

type perfSummary struct {
	TotalDurationMs int64 `json:"total_duration_ms"`
}

type perfItem struct {
	Name    string         `json:"name"`
	Config  map[string]any `json:"config"`
	Metrics perfMetrics    `json:"metrics"`
}

type perfMetrics struct {
	Iterations        int    `json:"iterations"`
	TotalDurationNs   int64  `json:"total_duration_ns"`
	NsPerOp           int64  `json:"ns_per_op"`
	TotalAllocBytes   uint64 `json:"total_alloc_bytes"`
	BytesPerOp        uint64 `json:"bytes_per_op"`
	TotalAllocs       uint64 `json:"total_allocs"`
	AllocsPerOp       uint64 `json:"allocs_per_op"`
	HeapAllocBefore   uint64 `json:"heap_alloc_before"`
	HeapAllocAfter    uint64 `json:"heap_alloc_after"`
	NumGCDelta        uint32 `json:"num_gc_delta"`
	PauseTotalNsDelta uint64 `json:"pause_total_ns_delta"`
	LastChunks        int    `json:"last_chunks"`
	LastChars         int    `json:"last_chars"`
}

func runPerfReport(ctx context.Context, input string, fileName string) perfReport {
	iters := 5
	if v := os.Getenv("SPLIT_PERF_ITERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			iters = n
		}
	}

	targetStrategy := os.Getenv("SPLIT_PERF_STRATEGY")

	var snap runtime.MemStats
	runtime.ReadMemStats(&snap)

	buildInfo := ""
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		buildInfo = bi.String()
	}

	report := perfReport{
		Meta: perfMeta{
			FileName:     fileName,
			InputBytes:   len([]byte(input)),
			Iterations:   iters,
			GoVersion:    runtime.Version(),
			GOOS:         runtime.GOOS,
			GOARCH:       runtime.GOARCH,
			CPUCount:     runtime.NumCPU(),
			GOMAXPROCS:   runtime.GOMAXPROCS(0),
			BuildVersion: buildInfo,
		},
		Snapshot: runtimeSnapshot{
			HeapAllocBytes: snap.HeapAlloc,
			HeapSysBytes:   snap.HeapSys,
			NumGC:          snap.NumGC,
			PauseTotalNs:   snap.PauseTotalNs,
		},
	}

	startAll := time.Now()

	items := make([]perfItem, 0, 4)
	addItem := func(name string, strategy perfStrategy) {
		if targetStrategy == "" || targetStrategy == name {
			items = append(items, measureStrategy(ctx, name, strategy, input, fileName, iters))
		}
	}
	addItem("fixed_size", newFixedSizePerfStrategy())
	addItem("recursive_character", newRecursivePerfStrategy())
	addItem("semantic", newSemanticPerfStrategy())
	addItem("document_structure", newDocumentStructurePerfStrategy())

	report.Items = items
	report.Summary = perfSummary{TotalDurationMs: time.Since(startAll).Milliseconds()}
	return report
}

type perfStrategy struct {
	Splitter split.Splitter
	Config   map[string]any
}

func newFixedSizePerfStrategy() perfStrategy {
	s := split.NewFixedSizeStrategy()
	s.ChunkSize = 400
	s.ChunkOverlap = 0
	s.OverlapRatio = 0.1
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	return perfStrategy{
		Splitter: s,
		Config: map[string]any{
			"chunk_size":           s.ChunkSize,
			"chunk_overlap":        s.ChunkOverlap,
			"overlap_ratio":        s.OverlapRatio,
			"remove_url_and_email": s.RemoveURLAndEmail,
			"normalize_whitespace": s.NormalizeWhitespace,
			"trim_space":           s.TrimSpace,
		},
	}
}

func newRecursivePerfStrategy() perfStrategy {
	s := split.NewRecursiveCharacterStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.15
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	return perfStrategy{
		Splitter: s,
		Config: map[string]any{
			"chunk_size":           s.ChunkSize,
			"overlap_ratio":        s.OverlapRatio,
			"remove_url_and_email": s.RemoveURLAndEmail,
			"normalize_whitespace": s.NormalizeWhitespace,
			"trim_space":           s.TrimSpace,
		},
	}
}

func newSemanticPerfStrategy() perfStrategy {
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
	return perfStrategy{
		Splitter: s,
		Config: map[string]any{
			"chunk_size":                     s.ChunkSize,
			"overlap_ratio":                  s.OverlapRatio,
			"remove_url_and_email":           s.RemoveURLAndEmail,
			"normalize_whitespace":           s.NormalizeWhitespace,
			"trim_space":                     s.TrimSpace,
			"semantic_mode":                  string(s.Mode),
			"semantic_buffer_size":           s.BufferSize,
			"semantic_breakpoint_percentile": s.BreakpointPercentile,
			"semantic_min_chunk_size":        s.MinChunkSize,
			"semantic_initial_threshold":     s.InitialThreshold,
			"semantic_appending_threshold":   s.AppendingThreshold,
			"semantic_merging_threshold":     s.MergingThreshold,
			"semantic_max_chunk_size":        s.MaxChunkSize,
			"semantic_merging_range":         s.MergingRange,
		},
	}
}

func newDocumentStructurePerfStrategy() perfStrategy {
	s := split.NewDocumentStructureStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.MaxDepth = 3
	s.SemanticThreshold = 0.5
	s.SkipEmptyHeadings = true
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	return perfStrategy{
		Splitter: s,
		Config: map[string]any{
			"chunk_size":           s.ChunkSize,
			"overlap_ratio":        s.OverlapRatio,
			"remove_url_and_email": s.RemoveURLAndEmail,
			"normalize_whitespace": s.NormalizeWhitespace,
			"trim_space":           s.TrimSpace,
			"semantic_threshold":   s.SemanticThreshold,
			"document_max_depth":   s.MaxDepth,
			"skip_empty_headings":  s.SkipEmptyHeadings,
		},
	}
}

func measureStrategy(ctx context.Context, name string, ps perfStrategy, input string, fileName string, iters int) perfItem {
	_, _ = ps.Splitter.Split(ctx, input, fileName)

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()
	var last []*split.Chunk
	var err error
	for i := 0; i < iters; i++ {
		last, err = ps.Splitter.Split(ctx, input, fileName)
		if err != nil {
			panic(err)
		}
	}
	elapsed := time.Since(start)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	totalAlloc := after.TotalAlloc - before.TotalAlloc
	totalAllocs := after.Mallocs - before.Mallocs
	ns := elapsed.Nanoseconds()

	var nsPerOp int64
	if iters > 0 {
		nsPerOp = ns / int64(iters)
	}
	var bytesPerOp uint64
	var allocsPerOp uint64
	if iters > 0 {
		bytesPerOp = totalAlloc / uint64(iters)
		allocsPerOp = totalAllocs / uint64(iters)
	}

	item := perfItem{
		Name:   name,
		Config: ps.Config,
		Metrics: perfMetrics{
			Iterations:        iters,
			TotalDurationNs:   ns,
			NsPerOp:           nsPerOp,
			TotalAllocBytes:   totalAlloc,
			BytesPerOp:        bytesPerOp,
			TotalAllocs:       totalAllocs,
			AllocsPerOp:       allocsPerOp,
			HeapAllocBefore:   before.HeapAlloc,
			HeapAllocAfter:    after.HeapAlloc,
			NumGCDelta:        after.NumGC - before.NumGC,
			PauseTotalNsDelta: after.PauseTotalNs - before.PauseTotalNs,
			LastChunks:        len(last),
			LastChars:         getTotalChars(last),
		},
	}
	return item
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
