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

	docs, err := s.Split(context.Background(), "abcdefghij1234567890", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 3)
	require.Equal(t, "abcdefghij", docs[0].SliceContent.Text)
	require.True(t, strings.HasPrefix(docs[1].SliceContent.Text, "ij"))
	require.True(t, strings.HasPrefix(docs[2].SliceContent.Text, "90") || strings.HasSuffix(docs[2].SliceContent.Text, "90"))
}

func TestFixedSizeStrategy_PreprocessOptions(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 500
	s.ChunkOverlap = 0
	s.RemoveURLAndEmail = true
	s.NormalizeWhitespace = true
	s.TrimSpace = true

	text := "联系 test@example.com\t\t访问 https://a.com  \n\n\n\n  结束"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.NotContains(t, docs[0].SliceContent.Text, "example.com")
	require.NotContains(t, docs[0].SliceContent.Text, "http")
	require.NotContains(t, docs[0].SliceContent.Text, "\t")
	require.NotContains(t, docs[0].SliceContent.Text, "\n\n\n")
}

func TestFixedSizeStrategy_RemoveImageURL(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 500
	s.ChunkOverlap = 0
	s.RemoveImageURL = true
	s.TrimSpace = true

	text := "这是一个文档 ![](abc.jpg) 包含图片 ![alt](def.png) 结束"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 1)
	// 图片链接应该被移除
	require.NotContains(t, docs[0].SliceContent.Text, "![")
	require.NotContains(t, docs[0].SliceContent.Text, "abc.jpg")
	require.NotContains(t, docs[0].SliceContent.Text, "def.png")
	// 普通文本应该保留
	require.Contains(t, docs[0].SliceContent.Text, "这是一个文档")
	require.Contains(t, docs[0].SliceContent.Text, "包含图片")
	require.Contains(t, docs[0].SliceContent.Text, "结束")
}

func TestRecursiveCharacterStrategy_RemoveImageURL(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 500
	s.OverlapRatio = 0.1
	s.RemoveImageURL = true
	s.TrimSpace = true

	text := "第一段内容 ![](img1.jpg)。\n\n第二段 ![alt2](img2.jpg) 继续。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	joined := ""
	for _, d := range docs {
		joined += d.SliceContent.Text + " "
	}
	// 所有图片链接应该被移除
	require.NotContains(t, joined, "![")
	require.NotContains(t, joined, "img1.jpg")
	require.NotContains(t, joined, "img2.jpg")
	// 普通文本应该保留
	require.Contains(t, joined, "第一段内容")
	require.Contains(t, joined, "第二段")
}

func TestSemanticStrategy_RemoveImageURL(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.RemoveImageURL = true
	s.TrimSpace = true

	text := "文本开始 ![](start.jpg) 中间内容 ![alt](mid.png) 文本结束"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	joined := ""
	for _, d := range docs {
		joined += d.SliceContent.Text + " "
	}
	// 图片链接应该被移除
	require.NotContains(t, joined, "![")
	require.NotContains(t, joined, "start.jpg")
	require.NotContains(t, joined, "mid.png")
	// 文本应该保留
	require.Contains(t, joined, "文本开始")
	require.Contains(t, joined, "文本结束")
}

func TestRecursiveCharacterStrategy_SeparatorsPriority(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 40
	s.OverlapRatio = 0.1
	s.TrimSpace = true

	text := "第一段。\n\n第二段很长很长很长很长很长很长，包含逗号，继续。\n第三行。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	for _, d := range docs {
		require.LessOrEqual(t, runeLen(d.SliceContent.Text), s.ChunkSize)
	}
}

func TestSemanticStrategy_MergeSimilarSentences(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 80
	s.OverlapRatio = 0.1
	s.Threshold = 0.5
	s.TrimSpace = true

	text := "苹果手机电池不耐用。苹果手机电池老化很快。今天天气很好。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(docs), 2)

	require.Contains(t, docs[0].SliceContent.Text, "苹果手机电池不耐用。")
	require.Contains(t, docs[0].SliceContent.Text, "苹果手机电池老化很快。")
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

	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	joined := ""
	for _, d := range docs {
		joined += d.SliceContent.Text + "\n\n"
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

	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	require.LessOrEqual(t, len(docs), 6)

	joined := ""
	for _, d := range docs {
		joined += d.SliceContent.Text + "\n\n"
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

	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	require.LessOrEqual(t, len(docs), 12)
	for _, d := range docs {
		require.LessOrEqual(t, runeLen(d.SliceContent.Text), s.ChunkSize)
	}

	joined := ""
	for _, d := range docs {
		joined += d.SliceContent.Text + "\n\n"
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

	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	findChunk := func(substr string) string {
		for _, d := range docs {
			if strings.Contains(d.SliceContent.Text, substr) {
				return d.SliceContent.Text
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

	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	// Verify chunks contain expected headings at the start
	found := false
	for _, d := range docs {
		// Title is now merged into Text at the beginning, before first \n\n
		if strings.HasPrefix(d.SliceContent.Text, "总标题") || strings.HasPrefix(d.SliceContent.Text, "一、章节") || strings.HasPrefix(d.SliceContent.Text, "1.1 小节") {
			found = true
			// Verify body content is also present
			require.Contains(t, d.SliceContent.Text, "筛选：")
			require.Contains(t, d.SliceContent.Text, "打分：")
			require.Contains(t, d.SliceContent.Text, "选择：")
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

	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	joined := ""
	for _, d := range docs {
		joined += d.SliceContent.Text + "\n\n"
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

	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	foundChild := false
	for _, d := range docs {
		if strings.HasPrefix(d.SliceContent.Text, "根 / 父标题 / 子标题") {
			foundChild = true
			require.Contains(t, d.SliceContent.Text, "### 子标题")
			require.Contains(t, d.SliceContent.Text, "子标题内容一。")
		}
	}
	require.True(t, foundChild)
}
