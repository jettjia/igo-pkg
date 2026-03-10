package split

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixedSizeStrategy_SplitWithOverlap(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 10
	s.ChunkOverlap = 2
	s.OverlapRatio = 0.1

	docs, err := s.Split(context.Background(), "abcdefghij1234567890")
	require.NoError(t, err)
	require.Len(t, docs, 3)
	require.Equal(t, "abcdefghij", docs[0].Content)
	require.True(t, strings.HasPrefix(docs[1].Content, "ij"))
	require.True(t, strings.HasPrefix(docs[2].Content, "90") || strings.HasSuffix(docs[2].Content, "90"))
}

func TestFixedSizeStrategy_PreprocessOptions(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 500
	s.ChunkOverlap = 0
	s.RemoveURLAndEmail = true
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	text := "联系 test@example.com\t\t访问 https://a.com  \n\n\n\n  结束"
	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.NotContains(t, docs[0].Content, "example.com")
	require.NotContains(t, docs[0].Content, "http")
	require.NotContains(t, docs[0].Content, "\t")
	require.NotContains(t, docs[0].Content, "\n\n\n")
}

func TestRecursiveCharacterStrategy_SeparatorsPriority(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 40
	s.OverlapRatio = 0.1
	s.TrimSpace = true

	text := "第一段。\n\n第二段很长很长很长很长很长很长，包含逗号，继续。\n第三行。"
	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	for _, d := range docs {
		require.LessOrEqual(t, runeLen(d.Content), s.ChunkSize)
	}
}

func TestSemanticStrategy_MergeSimilarSentences(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 80
	s.OverlapRatio = 0.1
	s.Threshold = 0.5
	s.TrimSpace = true

	text := "苹果手机电池不耐用。苹果手机电池老化很快。今天天气很好。"
	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(docs), 2)

	require.Contains(t, docs[0].Content, "苹果手机电池不耐用。")
	require.Contains(t, docs[0].Content, "苹果手机电池老化很快。")
}

func TestSemanticStrategy_CosineSimilarityBaseline(t *testing.T) {
	sentences := []string{"苹果手机电池不耐用。", "苹果手机电池老化很快。", "今天天气很好。"}
	vecs := buildTfidfVectors(sentences)
	require.Len(t, vecs, 3)
	sim := cosineSimilarity(vecs[0], vecs[1])
	require.GreaterOrEqual(t, sim, 0.5)
}

func TestSemanticStrategy_PreserveMarkdownNewlinesAndCodeBlocks(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.Threshold = 0.5
	s.NormalizeWhitespace = true
	s.TrimSpace = true
	s.Mode = SemanticModeGreedy

	text := strings.Join([]string{
		"# 顶部标题",
		"",
		"## 二、核心特性",
		"",
		"### 2.1 自动化部署与扩展",
		"",
		"正文一段。",
		"",
		"```yaml",
		"apiVersion: apps/v1",
		"kind: Deployment",
		"```",
		"",
		"代码块后续说明。",
	}, "\n")

	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	joined := ""
	for _, d := range docs {
		joined += d.Content + "\n\n"
	}
	require.Contains(t, joined, "# 顶部标题\n")
	require.NotContains(t, joined, "# 顶部标题##")
	require.Contains(t, joined, "```yaml\napiVersion: apps/v1")
	require.NotContains(t, joined, "```yamlapiVersion")
}

func TestSemanticStrategy_WindowBreakpoint_MoreCompact(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = SemanticModeWindowBreakpoint
	s.BufferSize = 1
	s.BreakpointPercentile = 95
	s.MinChunkSize = 120
	s.ChunkSize = 400
	s.OverlapRatio = 0.1
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	text := strings.Join([]string{
		"# 顶部标题",
		"",
		"## 二、核心特性",
		"",
		"### 2.1 自动化部署与扩展",
		"",
		"正文一段。",
		"",
		"```yaml",
		"apiVersion: apps/v1",
		"kind: Deployment",
		"```",
		"",
		"代码块后续说明。",
		"",
		"### 2.2 自我修复",
		"说明一段。",
	}, "\n")

	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	require.LessOrEqual(t, len(docs), 6)

	joined := ""
	for _, d := range docs {
		joined += d.Content + "\n\n"
	}
	require.Contains(t, joined, "# 顶部标题\n")
	require.Contains(t, joined, "```yaml\napiVersion: apps/v1")
	require.NotContains(t, joined, "```yamlapiVersion")
}

func TestSemanticStrategy_DoubleMerging_MoreCompact(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = SemanticModeDoubleMerging
	s.ChunkSize = 400
	s.MaxChunkSize = 400
	s.MinChunkSize = 120
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MergingSeparator = ""
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	text := strings.Join([]string{
		"# 顶部标题",
		"",
		"## 二、核心特性",
		"",
		"### 2.1 自动化部署与扩展",
		"",
		"正文一段。",
		"补充一句。",
		"",
		"```yaml",
		"apiVersion: apps/v1",
		"kind: Deployment",
		"```",
		"",
		"代码块后续说明。",
		"",
		"### 2.2 自我修复",
		"说明一段。",
		"再补充一句。",
	}, "\n")

	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	require.LessOrEqual(t, len(docs), 12)
	for _, d := range docs {
		require.LessOrEqual(t, runeLen(d.Content), s.ChunkSize)
	}

	joined := ""
	for _, d := range docs {
		joined += d.Content + "\n\n"
	}
	require.Contains(t, joined, "# 顶部标题\n")
	require.Contains(t, joined, "```yaml\napiVersion: apps/v1")
	require.NotContains(t, joined, "```yamlapiVersion")
}

func TestDocumentStructureStrategy_HeadingMustStayWithContent(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 120
	s.OverlapRatio = 0.1
	s.TrimSpace = true

	text := strings.Join([]string{
		"一、总则",
		"这是总则的第一段。",
		"",
		"（一）范围",
		"范围内容A。",
		"范围内容B。",
		"",
		"1. 第二章",
		"第二章内容。",
		"1.1 子节",
		"子节内容第一句。子节内容第二句。",
		"1.1.1 小节",
		"小节内容很长很长很长很长很长很长很长很长很长很长很长。",
	}, "\n")

	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	findChunk := func(substr string) string {
		for _, d := range docs {
			if strings.Contains(d.Content, substr) {
				return d.Content
			}
		}
		return ""
	}

	ch1 := findChunk("一、总则")
	require.NotEmpty(t, ch1)
	require.Contains(t, ch1, "这是总则的第一段。")

	ch2 := findChunk("（一）范围")
	require.NotEmpty(t, ch2)
	require.Contains(t, ch2, "范围内容A。")

	ch3 := findChunk("1. 第二章")
	require.NotEmpty(t, ch3)
	require.Contains(t, ch3, "第二章内容。")

	ch4 := findChunk("1.1 子节")
	require.NotEmpty(t, ch4)
	require.Contains(t, ch4, "子节内容第一句。")

	ch5 := findChunk("1.1.1 小节")
	require.NotEmpty(t, ch5)
	require.Contains(t, ch5, "小节内容很长")
}

func TestDocumentStructureStrategy_MarkdownHeadingsAvoidListItems(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 2000
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	text := strings.Join([]string{
		"# 总标题",
		"",
		"## 一、章节",
		"",
		"### 1.1 小节",
		"",
		"这里是正文。",
		"",
		"**调度逻辑**：",
		"",
		"1. 筛选：排除不满足条件的节点。",
		"2. 打分：对符合条件的节点打分。",
		"3. 选择：选中得分最高的节点。",
	}, "\n")

	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	for _, d := range docs {
		require.NotEqual(t, "筛选：排除不满足条件的节点。", d.Title)
		require.NotContains(t, d.Title, "筛选：")
		require.NotContains(t, d.Title, "打分：")
		require.NotContains(t, d.Title, "选择：")
	}

	found := false
	for _, d := range docs {
		if d.Title == "总标题" || d.Title == "一、章节" || d.Title == "1.1 小节" {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestDocumentStructureStrategy_PreserveCodeBlockNewlines(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	text := strings.Join([]string{
		"# 顶部标题",
		"",
		"## 二、核心特性",
		"",
		"### 2.1 自动化部署与扩展",
		"",
		"正文一段。",
		"",
		"```yaml",
		"apiVersion: apps/v1",
		"kind: Deployment",
		"```",
		"",
		"代码块后续说明。",
	}, "\n")

	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	joined := ""
	for _, d := range docs {
		joined += d.Content + "\n\n"
	}
	require.Contains(t, joined, "```yaml\napiVersion: apps/v1")
	require.NotContains(t, joined, "```yamlapiVersion")
}

func TestDocumentStructureStrategy_SkipEmptyHeadings(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 2000
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3
	s.SkipEmptyHeadings = true

	text := strings.Join([]string{
		"# 根",
		"",
		"## 父标题",
		"",
		"### 子标题",
		"子标题内容一。",
		"子标题内容二。",
	}, "\n")

	docs, err := s.Split(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	for _, d := range docs {
		require.NotEqual(t, "根 / 父标题", d.Title)
	}

	foundChild := false
	for _, d := range docs {
		if d.Title == "根 / 父标题 / 子标题" {
			foundChild = true
			require.Contains(t, d.Content, "### 子标题")
			require.Contains(t, d.Content, "子标题内容一。")
		}
	}
	require.True(t, foundChild)
}
