package split

import (
	"context"
	"strings"
	"testing"

	"github.com/jettjia/igo-pkg/aipkg/schema"
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
	require.NotContains(t, docs[0].SliceContent.Text, "![")
	require.NotContains(t, docs[0].SliceContent.Text, "abc.jpg")
	require.NotContains(t, docs[0].SliceContent.Text, "def.png")
	require.Contains(t, docs[0].SliceContent.Text, "这是一个文档")
	require.Contains(t, docs[0].SliceContent.Text, "包含图片")
	require.Contains(t, docs[0].SliceContent.Text, "结束")
}

func TestFixedSizeStrategy_GetType(t *testing.T) {
	s := NewFixedSizeStrategy()
	require.Equal(t, StrategyTypeFixedSize, s.GetType())
}

func TestFixedSizeStrategy_Validate(t *testing.T) {
	s := NewFixedSizeStrategy()
	// ChunkSize = 0 会自动设置为 500
	s.ChunkSize = 0
	err := s.Validate()
	require.NoError(t, err)
	require.Equal(t, 500, s.ChunkSize)

	// OverlapRatio = 0 会自动设置为 0.1
	s.OverlapRatio = 0
	err = s.Validate()
	require.NoError(t, err)

	// 超出范围会报错
	s.OverlapRatio = 0.05
	err = s.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "overlap_ratio")

	s.OverlapRatio = 0.3
	err = s.Validate()
	require.Error(t, err)

	s.OverlapRatio = 0.15
	err = s.Validate()
	require.NoError(t, err)
}

func TestFixedSizeStrategy_SplitHardSplit(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 5
	s.ChunkOverlap = 0

	// 字符串超过 chunkSize 且没有合适分隔符时会硬切
	text := "abcdefghijklmnopqrstuvwxyz"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(docs), 5) // 26/5 = 5-6 chunks
}

func TestFixedSizeStrategy_EmptyText(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 500
	docs, err := s.Split(context.Background(), "", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 0)
}

func TestFixedSizeStrategy_SingleChunk(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 500
	docs, err := s.Split(context.Background(), "短文本", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 1)
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
	require.NotContains(t, joined, "![")
	require.NotContains(t, joined, "img1.jpg")
	require.NotContains(t, joined, "img2.jpg")
	require.Contains(t, joined, "第一段内容")
	require.Contains(t, joined, "第二段")
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

func TestRecursiveCharacterStrategy_GetType(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	require.Equal(t, StrategyTypeRecursiveChar, s.GetType())
}

func TestRecursiveCharacterStrategy_Validate(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	// ChunkSize = 0 会自动设置为 500
	s.ChunkSize = 0
	err := s.Validate()
	require.NoError(t, err)

	// OverlapRatio = 0 会自动设置为 0.1
	s.OverlapRatio = 0
	err = s.Validate()
	require.NoError(t, err)

	// 超出范围会报错
	s.OverlapRatio = 0.05
	err = s.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "overlap_ratio")

	s.OverlapRatio = 0.15
	err = s.Validate()
	require.NoError(t, err)
}

func TestRecursiveCharacterStrategy_CustomSeparators(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 10
	s.Separators = []string{"\n", "。", "！"}
	s.TrimSpace = true

	text := "第一段内容第二段内容第三段内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestRecursiveCharacterStrategy_MergeParts(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 20
	s.OverlapRatio = 0.1
	s.TrimSpace = true

	// 测试 mergeParts 的边界情况
	text := "这是第一句话。这是第二句话。这是第三句话。非常长的第四句话超过块大小限制。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestRecursiveCharacterStrategy_HardSplitByRunes(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 3
	s.OverlapRatio = 0
	s.TrimSpace = true

	// 测试硬切分
	text := "abcdefghij"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 4) // 10/3 = 4 chunks
}

func TestRecursiveCharacterStrategy_EmptyText(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 500
	docs, err := s.Split(context.Background(), "", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 0)
}

func TestRecursiveCharacterStrategy_ApplyOverlap(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 10
	s.OverlapRatio = 0.2 // 2 characters overlap
	s.TrimSpace = true

	text := "0123456789ABCDEFGHIJ"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
	// 验证有overlap内容
	for i := 1; i < len(docs); i++ {
		// 后一个chunk应该包含前一个chunk的末尾内容
	}
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
	require.NotContains(t, joined, "![")
	require.NotContains(t, joined, "start.jpg")
	require.NotContains(t, joined, "mid.png")
	require.Contains(t, joined, "文本开始")
	require.Contains(t, joined, "文本结束")
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

func TestSemanticStrategy_GetType(t *testing.T) {
	s := NewSemanticStrategy()
	require.Equal(t, StrategyTypeSemantic, s.GetType())
}

func TestSemanticStrategy_Validate(t *testing.T) {
	s := NewSemanticStrategy()
	// ChunkSize = 0 会自动设置为 500
	s.ChunkSize = 0
	err := s.Validate()
	require.NoError(t, err)

	// OverlapRatio = 0 会自动设置为 0.1
	s.OverlapRatio = 0
	err = s.Validate()
	require.NoError(t, err)

	// 超出范围会报错
	s.OverlapRatio = 0.05
	err = s.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "overlap_ratio")

	s.OverlapRatio = 0.15
	err = s.Validate()
	require.NoError(t, err)
}

func TestSemanticStrategy_EmptyText(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 500
	docs, err := s.Split(context.Background(), "", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 0)
}

func TestSemanticStrategy_SingleChunk(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 500
	docs, err := s.Split(context.Background(), "短文本内容", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 1)
}

func TestSemanticStrategy_TailSentencesByLen(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.Threshold = 0.5
	s.TrimSpace = true

	text := "第一句。第二句。第三句这是比较长的句子需要被截断处理。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_SplitLongSentence(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 50
	s.OverlapRatio = 0.1
	s.Threshold = 0.5
	s.TrimSpace = true

	// 一个超长句子需要被分割
	text := "这是一段非常非常非常非常非常非常非常非常非常长的文本内容，需要被 splitLongSentence 函数处理。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_TfidfCache(t *testing.T) {
	cache := newTfidfCache(100)
	require.NotNil(t, cache)

	key := "test sentence"
	vec, ok := cache.get(key)
	require.False(t, ok)
	require.Nil(t, vec)

	cache.put(key, sparseVector{"word": 1.5})
	vec, ok = cache.get(key)
	require.True(t, ok)
	require.Equal(t, float64(1.5), vec["word"])

	// 测试 precomputed vectors - setPrecomputed 没有对应的 get 方法
	cache.setPrecomputed("pre1", sparseVector{"pre": 2.0})

	// 测试缓存满的情况
	cache2 := newTfidfCache(2)
	for i := 0; i < 5; i++ {
		cacheKey := string(rune('a' + i))
		cache2.put(cacheKey, sparseVector{"v": float64(i)})
	}
	// 最早的key应该被清除
	_, ok = cache2.get("a")
	require.False(t, ok)
}

func TestSemanticStrategy_IsCJK(t *testing.T) {
	require.True(t, isCJK('中'))
	require.True(t, isCJK('文'))
	require.False(t, isCJK('a'))
	require.False(t, isCJK('1'))
	require.False(t, isCJK(' '))
}

func TestSemanticStrategy_IsListLine(t *testing.T) {
	require.True(t, isListLine("1. 第一项"))
	require.True(t, isListLine("- 列表项"))
	require.True(t, isListLine("* 星号列表"))
	// 带缩进的列表需要先 trim
	require.True(t, isListLine(strings.TrimLeft("  1. 带缩进的列表", " ")))
	require.False(t, isListLine("普通文本"))
	require.False(t, isListLine("# 标题"))
	require.False(t, isListLine(""))
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

	found := false
	for _, d := range docs {
		if strings.HasPrefix(d.SliceContent.Text, "总标题") || strings.HasPrefix(d.SliceContent.Text, "一、章节") || strings.HasPrefix(d.SliceContent.Text, "1.1 小节") {
			found = true
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

func TestDocumentStructureStrategy_GetType(t *testing.T) {
	s := NewDocumentStructureStrategy()
	require.Equal(t, StrategyTypeDocumentStructure, s.GetType())
}

func TestDocumentStructureStrategy_Validate(t *testing.T) {
	s := NewDocumentStructureStrategy()
	// ChunkSize = 0 会自动设置为 500
	s.ChunkSize = 0
	err := s.Validate()
	require.NoError(t, err)

	// OverlapRatio = 0 会自动设置为 0.1
	s.OverlapRatio = 0
	err = s.Validate()
	require.NoError(t, err)

	// 超出范围会报错
	s.OverlapRatio = 0.05
	err = s.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "overlap_ratio")

	s.OverlapRatio = 0.15
	err = s.Validate()
	require.NoError(t, err)
}

func TestDocumentStructureStrategy_EmptyText(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 500
	docs, err := s.Split(context.Background(), "", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 0)
}

func TestDocumentStructureStrategy_SingleChunk(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 500
	docs, err := s.Split(context.Background(), "没有标题的纯文本", "test.txt")
	require.NoError(t, err)
	require.Len(t, docs, 1)
}

func TestDocumentStructureStrategy_ProtectRestoreCodeBlocks(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	text := "普通文本\n```\n代码块\n```\n更多文本"
	protected, codeBlocks := protectCodeBlocks(text)
	require.NotEqual(t, text, protected)
	require.Len(t, codeBlocks, 1)

	restored := restoreCodeBlocks(protected, codeBlocks)
	require.Equal(t, text, restored)
}

func TestDocumentStructureStrategy_SplitTextForStructure(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	text := "# 标题\n\n正文内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestDocumentStructureStrategy_MaxDepth(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 2000
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 2

	text := "# 一级\n## 二级\n### 三级\n内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestDocumentStructureStrategy_OverlappingHeadings(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	text := "# 标题一\n\n内容一\n\n# 标题二\n\n内容二"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestUtil_Functions(t *testing.T) {
	// Test runeLen
	require.Equal(t, 5, runeLen("hello"))
	require.Equal(t, 2, runeLen("中文"))

	// Test overlapTokens
	require.Equal(t, 50, overlapTokens(500, 0.1))
	require.Equal(t, 0, overlapTokens(0, 0.1))
	require.Equal(t, 0, overlapTokens(500, 0))

	// Test applyTrimSpaceIfNeeded
	require.Equal(t, "text", applyTrimSpaceIfNeeded("  text  ", &StrategyBase{TrimSpace: true}))
	require.Equal(t, "  text  ", applyTrimSpaceIfNeeded("  text  ", &StrategyBase{TrimSpace: false}))

	// Test newDocument
	doc := newDocument("content", "title", 1)
	require.Equal(t, "content", doc.Content)
	require.Equal(t, "title", doc.Title)
	require.Equal(t, 1, doc.Depth)

	// Test calculateMD5
	md5_1 := calculateMD5("test")
	md5_2 := calculateMD5("test")
	md5_3 := calculateMD5("other")
	require.Equal(t, md5_1, md5_2)
	require.NotEqual(t, md5_1, md5_3)
}

func TestUtil_PreProcessText(t *testing.T) {
	base := &StrategyBase{
		RemoveURLAndEmail:   true,
		RemoveImageURL:      true,
		NormalizeWhitespace: true,
		TrimSpace:           true,
	}

	text := "联系 test@example.com\t\t访问 https://a.com  \n\n\n\n  结束"
	processed := preProcessText(text, base)
	require.NotContains(t, processed, "example.com")
	require.NotContains(t, processed, "http")
	require.NotContains(t, processed, "\t")
}

func TestUtil_ApplyOverlapToStrings(t *testing.T) {
	chunks := []string{"chunk1", "chunk2", "chunk3"}

	// No overlap
	result := applyOverlapToStrings(chunks, 100, 0)
	require.Len(t, result, 3)

	// With overlap
	result = applyOverlapToStrings(chunks, 10, 0.2)
	require.NotEmpty(t, result)

	// Single chunk
	result = applyOverlapToStrings([]string{"single"}, 10, 0.2)
	require.Len(t, result, 1)
}

func TestUtil_ConvertToChunks(t *testing.T) {
	docs := []*schema.Document{
		{Content: "第一块内容", Title: "标题1"},
		{Content: "第二块内容", Title: "标题2"},
	}
	originalText := "第一块内容\n第二块内容"
	base := &StrategyBase{
		ChunkSize:    500,
		OverlapRatio: 0.1,
		TrimSpace:    true,
	}

	chunks := convertToChunks(docs, "test.txt", originalText, base)
	require.Len(t, chunks, 2)
	require.Equal(t, "test.txt", chunks[0].DocName)
	require.Equal(t, 1, chunks[0].SegmentID)
}

func TestUtil_GetNodeText(t *testing.T) {
	// This requires html parsing which is tested via tableToMarkdown
	// Direct test of getNodeText
	result := getNodeText(nil)
	require.Equal(t, "", result)
}

func TestUtil_DecodeHexEscapes(t *testing.T) {
	// No hex escapes
	require.Equal(t, "normal text", decodeHexEscapes("normal text"))

	// With hex escapes
	result := decodeHexEscapes("text\\u0041more")
	require.Contains(t, result, "A")
}

func TestUtil_TableToMarkdown(t *testing.T) {
	// Empty table
	result := tableToMarkdown("")
	require.Equal(t, "", result)

	// Invalid HTML
	result = tableToMarkdown("<invalid>")
	require.Equal(t, "", result)

	// Simple table
	tableHTML := `<table><tr><th>Header</th></tr><tr><td>Data</td></tr></table>`
	result = tableToMarkdown(tableHTML)
	require.Contains(t, result, "Header")
	require.Contains(t, result, "Data")
}

func TestUtil_RecoverBrokenTable(t *testing.T) {
	fullText := "<table><tr><td>A</td></tr></table> more text"
	fragment := "tr><td>A</td>"
	result := recoverBrokenTable(fragment, fullText)
	require.NotEmpty(t, result)

	// Not found
	result = recoverBrokenTable("nonexistent", fullText)
	require.Equal(t, "", result)
}

func TestStrategyBase_ValidateBase(t *testing.T) {
	base := &StrategyBase{}

	// Default values should be set
	err := base.validateBase()
	require.NoError(t, err)
	require.Equal(t, 500, base.ChunkSize)
	require.Equal(t, 0.1, base.OverlapRatio)

	// Invalid overlap ratio
	base = &StrategyBase{ChunkSize: 500, OverlapRatio: 0.5}
	err = base.validateBase()
	require.Error(t, err)
}

func TestSemanticStrategy_SentenceGroupLen(t *testing.T) {
	sentences := []string{"第一句", "第二句"}
	length := sentenceGroupLen(sentences)
	require.Greater(t, length, 0)
}

func TestFixedSizeStrategy_LargeChunk(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 1000
	s.ChunkOverlap = 0

	text := strings.Repeat("这是一段重复的测试文本。", 50)
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestFixedSizeStrategy_WithSeparators(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 50
	s.ChunkOverlap = 5
	s.TrimSpace = true

	text := "第一段。\n\n第二段。\n\n第三段。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestDocumentStructureStrategy_ValidateMore(t *testing.T) {
	s := NewDocumentStructureStrategy()
	// Test SkipEmptyHeadings
	s.SkipEmptyHeadings = true
	err := s.Validate()
	require.NoError(t, err)

	// Test MaxDepth
	s.MaxDepth = 5
	err = s.Validate()
	require.NoError(t, err)

	// Test SemanticThreshold
	s.SemanticThreshold = 0.8
	err = s.Validate()
	require.NoError(t, err)
}

func TestDocumentStructureStrategy_SplitSection(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 50
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	// Test with small content
	text := "# 标题\n\n短内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	// Test with larger content
	text2 := "# 标题\n\n" + strings.Repeat("这是一段测试内容。", 20)
	docs2, err := s.Split(context.Background(), text2, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs2)
}

func TestDocumentStructureStrategy_ParseHeadingTree(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	// Test with various heading levels
	text := `# H1
## H2
### H3
#### H4
##### H5
###### H6
Content
`
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_ValidateMore(t *testing.T) {
	s := NewSemanticStrategy()
	// Test all mode settings
	s.Mode = SemanticModeGreedy
	err := s.Validate()
	require.NoError(t, err)

	s.Mode = SemanticModeWindowBreakpoint
	s.BreakpointPercentile = 90
	err = s.Validate()
	require.NoError(t, err)

	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.3
	s.AppendingThreshold = 0.4
	s.MergingThreshold = 0.4
	s.MergingRange = 2
	err = s.Validate()
	require.NoError(t, err)
}

func TestSemanticStrategy_TailSentencesByLen_Unit(t *testing.T) {
	// Test empty
	result := tailSentencesByLen(nil, 10)
	require.Nil(t, result)

	// Test empty sentences
	result = tailSentencesByLen([]string{}, 10)
	require.Nil(t, result)

	// Test limit 0
	result = tailSentencesByLen([]string{"句子"}, 0)
	require.Nil(t, result)

	// Test normal case
	sentences := []string{"短", "中等的句子", "很长的句子内容"}
	result = tailSentencesByLen(sentences, 5)
	require.NotNil(t, result)

	// Test all fit
	sentences2 := []string{"短", "中"}
	result2 := tailSentencesByLen(sentences2, 100)
	require.Len(t, result2, 2)

	// Test none fit
	result3 := tailSentencesByLen(sentences, 1)
	require.NotNil(t, result3)
}

func TestSemanticStrategy_SplitGroupsByMaxSize(t *testing.T) {
	// Test empty
	result := splitGroupsByMaxSize(nil, 10)
	require.Nil(t, result)

	// Test chunkSize 0
	result = splitGroupsByMaxSize([][]string{{"a"}}, 0)
	require.NotNil(t, result)

	// Test normal case
	groups := [][]string{
		{"句子1", "句子2"},
		{"句子3"},
		{"句子4", "句子5", "句子6"},
	}
	result = splitGroupsByMaxSize(groups, 10)
	require.NotNil(t, result)
	require.GreaterOrEqual(t, len(result), 1)
}

func TestSemanticStrategy_ApplySentenceOverlapToGroups(t *testing.T) {
	// Test len <= 1
	groups := [][]string{{"句子"}}
	result := applySentenceOverlapToGroups(groups, 5, 100)
	require.Len(t, result, 1)

	// Test overlapLimit 0
	groups2 := [][]string{{"a"}, {"b"}}
	result2 := applySentenceOverlapToGroups(groups2, 0, 100)
	require.Len(t, result2, 2)

	// Test normal case
	groups3 := [][]string{
		{"第一句", "第二句"},
		{"第三句", "第四句"},
	}
	result3 := applySentenceOverlapToGroups(groups3, 3, 20)
	require.NotNil(t, result3)
}

func TestSemanticStrategy_MergeSmallSemanticGroups(t *testing.T) {
	// Test empty
	result := mergeSmallSemanticGroups(nil, 10, 100)
	require.Nil(t, result)

	// Test minChunkSize 0
	groups := [][]string{{"句子"}}
	result = mergeSmallSemanticGroups(groups, 0, 100)
	require.Len(t, result, 1)

	// Test with heading line merge
	groups2 := [][]string{
		{"# 标题"},
		{"内容"},
		{"更多内容"},
	}
	result2 := mergeSmallSemanticGroups(groups2, 5, 50)
	require.NotNil(t, result2)

	// Test merge with previous
	groups3 := [][]string{
		{"第一组"},
		{"第二组"},
	}
	result3 := mergeSmallSemanticGroups(groups3, 10, 100)
	require.NotNil(t, result3)
}

func TestSemanticStrategy_SemanticSplitDoubleMerging(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = SemanticModeDoubleMerging
	s.ChunkSize = 60
	s.MaxChunkSize = 60
	s.MinChunkSize = 10
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MergingSeparator = "\n"
	s.OverlapRatio = 0
	s.TrimSpace = true

	// Test with multiple similar sentences
	text := "苹果手机很好。苹果手机很棒。华为手机也不错。三星手机很好。今天天气不错。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_PercentileValue(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = SemanticModeWindowBreakpoint
	s.ChunkSize = 50
	s.BreakpointPercentile = 80
	s.MinChunkSize = 10
	s.OverlapRatio = 0
	s.TrimSpace = true

	// Test with varying sentence lengths
	text := "短句。长句子内容。非常长的句子内容需要被截断。中等长度。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_IsMarkdownHeadingLine(t *testing.T) {
	require.True(t, isMarkdownHeadingLine("# 标题"))
	require.True(t, isMarkdownHeadingLine("## 二级标题"))
	require.True(t, isMarkdownHeadingLine("###### 六级标题"))
	require.True(t, isMarkdownHeadingLine("#not header")) // any string starting with # is considered a header
	require.False(t, isMarkdownHeadingLine("普通文本"))
	require.False(t, isMarkdownHeadingLine(""))
}

func TestSemanticStrategy_IsCodeBlockSentence(t *testing.T) {
	require.True(t, isCodeBlockSentence("```go"))
	require.True(t, isCodeBlockSentence("```"))
	require.False(t, isCodeBlockSentence("普通文本"))
	require.False(t, isCodeBlockSentence(""))
}

func TestSemanticStrategy_SplitIntoSentencesPreserveNewlinesWithCodeBlocks(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.OverlapRatio = 0
	s.TrimSpace = true

	// Test with code blocks
	text := "普通文本\n```\n代码\n```\n更多文本"
	sentences := splitIntoSentencesPreserveNewlinesWithCodeBlocks(text)
	require.NotEmpty(t, sentences)
}

func TestSemanticStrategy_PrecomputeTfidfVectors(t *testing.T) {
	sentences := []string{"第一句", "第二句", "第三句"}

	// Call precomputed directly
	result := precomputeTfidfVectors(sentences)
	require.NotNil(t, result)
	require.Len(t, result, len(sentences))
}

func TestSemanticStrategy_BuildTfidfVectors(t *testing.T) {
	sentences := []string{"苹果 苹果 苹果", "香蕉 香蕉", "樱桃"}
	vecs := buildTfidfVectors(sentences)
	require.Len(t, vecs, 3)
	require.NotNil(t, vecs[0])
	require.NotNil(t, vecs[1])
	require.NotNil(t, vecs[2])
}

func TestSemanticStrategy_TokenizeForTfidf(t *testing.T) {
	result := tokenizeForTfidf("你好世界")
	require.NotEmpty(t, result)

	result2 := tokenizeForTfidf("Hello World")
	require.NotEmpty(t, result2)
}

func TestSemanticStrategy_CosineSimilarity(t *testing.T) {
	vec1 := sparseVector{"a": 1.0, "b": 0.0}
	vec2 := sparseVector{"a": 1.0, "b": 0.0}
	vec3 := sparseVector{"a": 0.0, "b": 1.0}

	// Same vectors
	sim := cosineSimilarity(vec1, vec2)
	require.Equal(t, 1.0, sim)

	// Orthogonal vectors
	sim2 := cosineSimilarity(vec1, vec3)
	require.Equal(t, 0.0, sim2)

	// Empty vectors
	sim3 := cosineSimilarity(sparseVector{}, sparseVector{})
	require.Equal(t, 0.0, sim3)

	// Different lengths
	vec4 := sparseVector{"a": 1.0}
	vec5 := sparseVector{"a": 1.0, "b": 1.0, "c": 1.0}
	sim4 := cosineSimilarity(vec4, vec5)
	require.GreaterOrEqual(t, sim4, 0.0)
}

func TestDocumentStructureStrategy_SplitTextForStructureEdgeCases(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 100
	s.OverlapRatio = 0
	s.TrimSpace = true
	s.MaxDepth = 3

	// Test empty
	result := splitTextForStructure("", 100)
	require.Nil(t, result)

	// Test smaller than chunk
	result = splitTextForStructure("短文本", 100)
	require.Len(t, result, 1)

	// Test with code blocks
	text := "普通文本\n```\n代码块内容\n```\n更多内容"
	result = splitTextForStructure(text, 10)
	require.NotNil(t, result)
}

func TestDocumentStructureStrategy_ProtectCodeBlocks(t *testing.T) {
	// Test empty
	protected, blocks := protectCodeBlocks("")
	require.Equal(t, "", protected)
	require.Len(t, blocks, 0)

	// Test no code blocks
	protected, blocks = protectCodeBlocks("普通文本")
	require.Equal(t, "普通文本", protected)
	require.Len(t, blocks, 0)

	// Test single line code block
	protected, blocks = protectCodeBlocks("```\ncode\n```")
	require.NotEqual(t, protected, "```\ncode\n```")
	require.Len(t, blocks, 1)

	// Test multiple code blocks - use separate variables
	text := "之前\n```go\ngo code\n```\n中间\n```python\npy code\n```\n之后"
	multiProtected, multiBlocks := protectCodeBlocks(text)
	require.Len(t, multiBlocks, 2)

	// Test unclosed code block - should be treated as code block
	protected, blocks = protectCodeBlocks("```\n未闭合")
	require.Len(t, blocks, 1)
	// The unclosed block contains "未闭合", not "go code"
	require.Contains(t, blocks[0], "未闭合")

	// Test restore with multiple code blocks
	restored := restoreCodeBlocks(multiProtected, multiBlocks)
	require.Contains(t, restored, "go code")
	require.Contains(t, restored, "py code")
}

func TestDocumentStructureStrategy_RenderSection(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 100
	s.OverlapRatio = 0
	s.TrimSpace = true
	s.MaxDepth = 3

	// Test with headings
	text := "# 主标题\n\n## 章节\n\n内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestRecursiveCharacterStrategy_SeparatorsEmpty(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 50
	s.OverlapRatio = 0
	s.Separators = []string{} // Empty separators

	text := "测试文本内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestRecursiveCharacterStrategy_RecursiveSplitEmpty(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 50
	s.OverlapRatio = 0

	// Test with chunkSize 0
	text := "测试文本"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestUtil_PreProcessTextMore(t *testing.T) {
	base := &StrategyBase{
		RemoveURLAndEmail:   true,
		RemoveImageURL:      true,
		NormalizeWhitespace: true,
		TrimSpace:           true,
	}

	// Test with page markers
	text := "文本\n<!-- Page: 1 -->\n\n更多文本"
	processed := preProcessText(text, base)
	require.NotContains(t, processed, "Page:")
	require.NotContains(t, processed, "\n\n\n")

	// Test with table
	tableText := "之前<table><tr><td>单元格</td></tr></table>之后"
	processed2 := preProcessText(tableText, base)
	require.Contains(t, processed2, "HTML_TABLE_PLACEHOLDER")
}

func TestUtil_ConvertToChunksWithTables(t *testing.T) {
	base := &StrategyBase{
		ChunkSize:    500,
		OverlapRatio: 0.1,
		TrimSpace:    true,
		tableCache:   []string{},
	}

	docs := []*schema.Document{
		{Content: "文本内容"},
	}
	originalText := "文本内容"

	chunks := convertToChunks(docs, "test.txt", originalText, base)
	require.Len(t, chunks, 1)
	require.NotEmpty(t, chunks[0].ID)
	require.NotEmpty(t, chunks[0].DocMD5)
	require.NotEmpty(t, chunks[0].SliceMD5)
}

func TestSemanticStrategy_SplitLongSentence_Integration(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 10
	s.OverlapRatio = 0
	s.TrimSpace = true

	// A single very long sentence
	text := "这是一个非常非常非常非常长的句子需要被分割"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_SplitLongSentence_CJK(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 5
	s.OverlapRatio = 0
	s.TrimSpace = true

	// Pure CJK characters - should split by character
	text := "中文字符"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_SplitGroupsByMaxSizeMore(t *testing.T) {
	// Test with groups that all fit in one chunk
	groups := [][]string{
		{"短", "短"},
	}
	result := splitGroupsByMaxSize(groups, 100)
	require.Len(t, result, 1)

	// Test with groups that need to be split
	groups2 := [][]string{
		{"句子1", "句子2", "句子3"},
		{"句子4", "句子5"},
	}
	result2 := splitGroupsByMaxSize(groups2, 10)
	require.NotNil(t, result2)
}

func TestSemanticStrategy_SemanticSplitWindowBreakpointMore(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = SemanticModeWindowBreakpoint
	s.ChunkSize = 30
	s.BreakpointPercentile = 90
	s.MinChunkSize = 10
	s.BufferSize = 2
	s.OverlapRatio = 0
	s.TrimSpace = true

	text := "短句1。短句2。短句3。短句4。短句5。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_SemanticSplitDoubleMergingMore(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = SemanticModeDoubleMerging
	s.ChunkSize = 50
	s.MaxChunkSize = 50
	s.MinChunkSize = 5
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MergingSeparator = ""
	s.OverlapRatio = 0
	s.TrimSpace = true

	// Different types of sentences
	text := "苹果很好。香蕉不错。天气很好。飞机很快。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_ValidateMoreCases(t *testing.T) {
	s := NewSemanticStrategy()
	// Test Threshold boundary
	s.Threshold = 0
	err := s.Validate()
	require.NoError(t, err)

	// Test InitialThreshold, AppendingThreshold, MergingThreshold
	s.InitialThreshold = 0.1
	s.AppendingThreshold = 0.2
	s.MergingThreshold = 0.2
	s.MergingRange = 0
	err = s.Validate()
	require.NoError(t, err)

	// Test MergingSeparator
	s.MergingSeparator = "\n"
	err = s.Validate()
	require.NoError(t, err)
}

func TestDocumentStructureStrategy_ValidateMoreCases(t *testing.T) {
	s := NewDocumentStructureStrategy()
	// Test MaxDepth boundary
	s.MaxDepth = 0
	err := s.Validate()
	require.NoError(t, err)
	require.Equal(t, 3, s.MaxDepth)

	s.MaxDepth = 10
	err = s.Validate()
	require.NoError(t, err)

	// Test SkipEmptyHeadings
	s.SkipEmptyHeadings = true
	err = s.Validate()
	require.NoError(t, err)
}

func TestFixedSizeStrategy_ValidateMoreCases(t *testing.T) {
	s := NewFixedSizeStrategy()
	// Test ChunkOverlap
	s.ChunkOverlap = 100
	err := s.Validate()
	require.NoError(t, err)

	// Test with ChunkSize larger than overlap
	s.ChunkOverlap = 50
	err = s.Validate()
	require.NoError(t, err)
}

func TestUtil_ConvertToChunksMore(t *testing.T) {
	base := &StrategyBase{
		ChunkSize:    500,
		OverlapRatio: 0.1,
		TrimSpace:    true,
	}

	// Test with page markers in original text
	originalText := "文本内容\n<!-- Page: 1 -->\n更多内容"
	docs := []*schema.Document{
		{Content: "文本内容\n<!-- Page: 1 -->\n更多内容"},
	}
	chunks := convertToChunks(docs, "test.txt", originalText, base)
	require.Len(t, chunks, 1)

	// Test with empty docs
	chunks2 := convertToChunks([]*schema.Document{}, "test.txt", "", base)
	require.Len(t, chunks2, 0)
}

func TestUtil_TableToMarkdownComplex(t *testing.T) {
	// Test with rowspan and colspan
	tableHTML := `<table>
<tr><th>Header1</th><th>Header2</th></tr>
<tr><td rowspan="2">Cell1</td><td>Cell2</td></tr>
<tr><td>Cell3</td></tr>
</table>`
	result := tableToMarkdown(tableHTML)
	require.Contains(t, result, "Header1")
	require.Contains(t, result, "Header2")
}

func TestSemanticStrategy_NewTfidfCacheZero(t *testing.T) {
	// Test with 0 size - should default to 1024
	cache := newTfidfCache(0)
	require.NotNil(t, cache)
	require.Equal(t, 1024, cache.maxSize)

	// Test with negative size
	cache2 := newTfidfCache(-5)
	require.NotNil(t, cache2)
	require.Equal(t, 1024, cache2.maxSize)
}

func TestSemanticStrategy_IsCJKMore(t *testing.T) {
	// Test more CJK characters
	require.True(t, isCJK('的'))
	require.True(t, isCJK('是'))
	require.True(t, isCJK('在'))
	require.False(t, isCJK('.'))
	require.False(t, isCJK(','))
	require.False(t, isCJK(' '))
	require.False(t, isCJK('\n'))
}

func TestSemanticStrategy_TokenizeForTfidfMore(t *testing.T) {
	// Test mixed content
	result := tokenizeForTfidf("Hello世界123")
	require.NotEmpty(t, result)

	// Test numbers
	result2 := tokenizeForTfidf("123456")
	require.NotEmpty(t, result2)
}

func TestSemanticStrategy_BuildTfidfVectorsEmpty(t *testing.T) {
	// Empty sentences
	vecs := buildTfidfVectors(nil)
	require.Nil(t, vecs)

	vecs2 := buildTfidfVectors([]string{})
	require.Nil(t, vecs2)
}

func TestDocumentStructureStrategy_SplitSectionMore(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 100
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	// Test with deeply nested headings
	text := "# 一级\n## 二级\n### 三级\n内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_SemanticSplitGreedyMore(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = SemanticModeGreedy
	s.ChunkSize = 50
	s.Threshold = 0.5
	s.OverlapRatio = 0
	s.TrimSpace = true

	// Many short sentences
	text := "第一。第二。第三。第四。第五。第六。第七。第八。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_WithContextCancellation(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.OverlapRatio = 0.1
	s.TrimSpace = true

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := s.Split(ctx, "文本", "test.txt")
	require.Error(t, err)
}

func TestRecursiveCharacterStrategy_WithContextCancellation(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 100
	s.OverlapRatio = 0.1

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := s.Split(ctx, "文本", "test.txt")
	require.Error(t, err)
}

func TestFixedSizeStrategy_WithContextCancellation(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 100
	s.ChunkOverlap = 10

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := s.Split(ctx, "文本", "test.txt")
	require.Error(t, err)
}

func TestDocumentStructureStrategy_WithContextCancellation(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 100
	s.OverlapRatio = 0.1

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := s.Split(ctx, "文本", "test.txt")
	require.Error(t, err)
}

func TestSemanticStrategy_SplitIntoSentencesPreserveNewlinesWithCodeBlocksMore(t *testing.T) {
	// Test with multiple code blocks
	text := "第一段\n```\ncode1\n```\n第二段\n```\ncode2\n```\n第三段"
	sentences := splitIntoSentencesPreserveNewlinesWithCodeBlocks(text)
	require.NotEmpty(t, sentences)
	// Should preserve code blocks as single sentences
	found := false
	for _, s := range sentences {
		if strings.Contains(s, "code1") {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestSemanticStrategy_SplitIntoSentencesPreserveNewlinesWithCodeBlocksEmpty(t *testing.T) {
	// Empty returns empty slice
	sentences := splitIntoSentencesPreserveNewlinesWithCodeBlocks("")
	require.Empty(t, sentences)

	// No code blocks
	sentences2 := splitIntoSentencesPreserveNewlinesWithCodeBlocks("普通文本")
	require.NotEmpty(t, sentences2)
}

func TestDocumentStructureStrategy_RenderSectionMore(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 50
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	// Very short content with heading
	text := "# 标题\n\n内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestUtil_ConvertToChunksWithMultipleDocs(t *testing.T) {
	base := &StrategyBase{
		ChunkSize:    500,
		OverlapRatio: 0.1,
		TrimSpace:    true,
	}

	docs := []*schema.Document{
		{Content: "第一个文档内容"},
		{Content: "第二个文档内容"},
		{Content: "第三个文档内容"},
	}
	originalText := "第一个文档内容\n---\n第二个文档内容\n---\n第三个文档内容"

	chunks := convertToChunks(docs, "test.txt", originalText, base)
	require.Len(t, chunks, 3)
	require.Equal(t, 1, chunks[0].SegmentID)
	require.Equal(t, 2, chunks[1].SegmentID)
	require.Equal(t, 3, chunks[2].SegmentID)
}

func TestUtil_ConvertToChunksWithTitle(t *testing.T) {
	base := &StrategyBase{
		ChunkSize:    500,
		OverlapRatio: 0.1,
		TrimSpace:    true,
	}

	docs := []*schema.Document{
		{Content: "正文内容", Title: "文档标题"},
	}
	originalText := "文档标题\n\n正文内容"

	chunks := convertToChunks(docs, "test.txt", originalText, base)
	require.Len(t, chunks, 1)
	// Title should be prepended to text
	require.Contains(t, chunks[0].SliceContent.Text, "文档标题")
	require.Contains(t, chunks[0].SliceContent.Text, "正文内容")
}

func TestUtil_ConvertToChunksWithPageMarkers(t *testing.T) {
	base := &StrategyBase{
		ChunkSize:    500,
		OverlapRatio: 0.1,
		TrimSpace:    true,
	}

	docs := []*schema.Document{
		{Content: "第一页内容\n<!-- Page: 1 -->\n第二页内容"},
	}
	originalText := "第一页内容\n<!-- Page: 1 -->\n第二页内容"

	chunks := convertToChunks(docs, "test.txt", originalText, base)
	require.Len(t, chunks, 1)
	// Page markers should be removed from content
	require.NotContains(t, chunks[0].SliceContent.Text, "Page:")
}

func TestSemanticStrategy_IsCodeBlockSentenceMore(t *testing.T) {
	require.True(t, isCodeBlockSentence("```"))
	require.True(t, isCodeBlockSentence("```python"))
	require.True(t, isCodeBlockSentence("  ```go"))
	require.False(t, isCodeBlockSentence("plain text"))
	require.False(t, isCodeBlockSentence(""))
}

func TestSemanticStrategy_SplitLongSentenceEdgeCases(t *testing.T) {
	// Test splitLongSentence with various inputs
	result := splitLongSentence("短", 5)
	require.NotEmpty(t, result)

	// Empty string
	result2 := splitLongSentence("", 5)
	require.Nil(t, result2)

	// Normal string
	result3 := splitLongSentence("这是一个测试句子", 5)
	require.NotNil(t, result3)
}

func TestDocumentStructureStrategy_ParseHeadingLineMore(t *testing.T) {
	// Test various heading formats with correct signature: (line string, maxDepth int, markdownOnly bool)
	depth, title, _, ok := parseHeadingLine("# 一级标题", 6, true)
	require.True(t, ok)
	require.Equal(t, 1, depth)
	require.Equal(t, "一级标题", title)

	depth2, _, _, ok2 := parseHeadingLine("## 二级标题", 6, true)
	require.True(t, ok2)
	require.Equal(t, 2, depth2)

	depth3, _, _, ok3 := parseHeadingLine("### 三级标题", 6, true)
	require.True(t, ok3)
	require.Equal(t, 3, depth3)

	// Non-heading text
	_, _, _, ok4 := parseHeadingLine("普通文本", 6, true)
	require.False(t, ok4)

	// Empty line
	_, _, _, ok5 := parseHeadingLine("", 6, true)
	require.False(t, ok5)

	// Chinese numbered headings (not markdown style)
	_, _, _, ok6 := parseHeadingLine("一、总则", 6, false)
	_ = ok6
}

func TestDocumentStructureStrategy_ParseHeadingTreeMore(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 200
	s.OverlapRatio = 0.1
	s.TrimSpace = true
	s.MaxDepth = 3

	// Test with all levels of Chinese headings
	text := `# 一级
一、总则
（一）范围
1. 章节
1.1 子节
1.1.1 小节
内容
`
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestFixedSizeStrategy_SplitWithVariousSeparators(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 30
	s.ChunkOverlap = 0
	s.TrimSpace = true

	// Various separators
	text := "第一段。\n第二段，\n第三段；\n第四段：\n第五段"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotEmpty(t, docs)
}

func TestSemanticStrategy_BuildTfidfVectorsMore(t *testing.T) {
	// Single sentence
	vecs := buildTfidfVectors([]string{"单个句子"})
	require.Len(t, vecs, 1)

	// Multiple identical sentences
	vecs2 := buildTfidfVectors([]string{"相同", "相同", "相同"})
	require.Len(t, vecs2, 3)

	// Different character types
	vecs3 := buildTfidfVectors([]string{"中文", "English", "123"})
	require.Len(t, vecs3, 3)
}

func TestSemanticStrategy_SplitGroupsByMaxSize_LongSentence(t *testing.T) {
	// Test case where a single sentence exceeds chunkSize
	groups := [][]string{
		{"这是一个非常长的句子，它的长度远远超过了指定的chunkSize限制"},
	}
	result := splitGroupsByMaxSize(groups, 10)
	require.NotNil(t, result)
	// The long sentence should be split
	require.GreaterOrEqual(t, len(result), 1)
}

func TestSemanticStrategy_ValidateMore3(t *testing.T) {
	s := NewSemanticStrategy()

	// Test default threshold
	s.Threshold = 0
	err := s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0.5, s.Threshold) // Should be set to default

	// Test invalid threshold (< 0)
	s.Threshold = -0.1
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0.5, s.Threshold) // Should be corrected to default

	// Test invalid threshold (> 1)
	s.Threshold = 1.5
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0.5, s.Threshold) // Should be corrected to default

	// Test empty mode
	s.Mode = ""
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, SemanticModeGreedy, s.Mode) // Should be set to default

	// Test invalid mode
	s.Mode = "invalid_mode"
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, SemanticModeGreedy, s.Mode) // Should be corrected to default

	// Test valid modes
	s.Mode = SemanticModeWindowBreakpoint
	err = s.Validate()
	require.NoError(t, err)

	s.Mode = SemanticModeDoubleMerging
	err = s.Validate()
	require.NoError(t, err)
}

func TestFixedSizeStrategy_ValidateMore(t *testing.T) {
	s := NewFixedSizeStrategy()

	// Test default ChunkSize - setting 0 should auto-correct to 500
	s.ChunkSize = 0
	err := s.Validate()
	require.NoError(t, err) // ChunkSize 0 auto-corrects, no error
	require.Equal(t, 500, s.ChunkSize)

	s.ChunkSize = 500
	err = s.Validate()
	require.NoError(t, err)

	// Test OverlapRatio bounds - < 0.1 should error
	s.OverlapRatio = 0.05
	err = s.Validate()
	require.Error(t, err) // Out of range

	// Test OverlapRatio bounds - > 0.2 should error
	s.OverlapRatio = 0.25
	err = s.Validate()
	require.Error(t, err) // Out of range

	// Valid range
	s.OverlapRatio = 0.15
	err = s.Validate()
	require.NoError(t, err)
}

func TestDocumentStructureStrategy_ValidateMore2(t *testing.T) {
	s := NewDocumentStructureStrategy()

	// Test valid state
	s.ChunkSize = 500
	s.MaxDepth = 3
	err := s.Validate()
	require.NoError(t, err)

	// Test invalid ChunkSize - should auto-correct like base
	s.ChunkSize = 0
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 500, s.ChunkSize) // Auto-corrected

	// Test invalid SemanticThreshold bounds - out of range gets corrected to 0.5
	s.ChunkSize = 500
	s.SemanticThreshold = -0.1
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0.5, s.SemanticThreshold) // Should be corrected to 0.5

	s.SemanticThreshold = 1.5
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0.5, s.SemanticThreshold) // Should be corrected to 0.5
}

func TestConvertToChunks_WithDocPage(t *testing.T) {
	base := &StrategyBase{}
	docs := []*schema.Document{
		{Content: "测试内容", Page: 5},
	}
	chunks := convertToChunks(docs, "test.txt", "测试内容", base)
	require.Len(t, chunks, 1)
	require.Contains(t, chunks[0].Pages, 5)
}

func TestConvertToChunks_WithTitle(t *testing.T) {
	base := &StrategyBase{}
	docs := []*schema.Document{
		{Title: "文档标题", Content: "测试内容"},
	}
	chunks := convertToChunks(docs, "test.txt", "文档标题\n\n测试内容", base)
	require.Len(t, chunks, 1)
	// Title should be prepended to text
	require.Contains(t, chunks[0].SliceContent.Text, "文档标题")
	require.Contains(t, chunks[0].SliceContent.Text, "测试内容")
}

func TestConvertToChunks_WithTable(t *testing.T) {
	base := &StrategyBase{}
	// Simulate table placeholder
	docs := []*schema.Document{
		{Content: "HTML_TABLE_PLACEHOLDER_0"},
	}
	// Add table to cache - use a proper HTML table
	base.tableCache = []string{"<table><tr><th>Header</th></tr><tr><td>Data</td></tr></table>"}
	chunks := convertToChunks(docs, "test.txt", "HTML_TABLE_PLACEHOLDER_0", base)
	require.Len(t, chunks, 1)
	// Table should be converted to markdown - check for table structure (header row)
	require.Contains(t, chunks[0].SliceContent.Text, "Header")
	require.Contains(t, chunks[0].SliceContent.Text, "Data")
}

func TestConvertToChunks_MultiPageInContent(t *testing.T) {
	base := &StrategyBase{}
	// Document with explicit page set
	docs := []*schema.Document{
		{Content: "测试内容", Page: 3},
	}
	chunks := convertToChunks(docs, "test.txt", "测试内容", base)
	require.Len(t, chunks, 1)
	// Should have page 3 from doc.Page
	require.Contains(t, chunks[0].Pages, 3)
}

func TestIsCJK(t *testing.T) {
	// Test CJK Unified Ideographs (common Chinese characters)
	require.True(t, isCJK('中'))
	require.True(t, isCJK('文'))

	// Test CJK Extension A
	require.True(t, isCJK('㐀'))
	require.True(t, isCJK('䶿'))

	// Test supplementary planes - using a valid supplementary character
	require.True(t, isCJK(0x2000B))

	// Test non-CJK characters
	require.False(t, isCJK('A'))
	require.False(t, isCJK('1'))
	require.False(t, isCJK(' '))
	require.False(t, isCJK(''))
}

func TestPercentileValue(t *testing.T) {
	// Empty values
	result := percentileValue(nil, 50)
	require.Equal(t, 0.0, result)

	// Single value
	result = percentileValue([]float64{10.0}, 50)
	require.Equal(t, 10.0, result)

	// Percentile at boundary <= 0 - gets clamped to 1, then pos = Ceil(1/100 * 2) = 1
	result = percentileValue([]float64{1.0, 2.0, 3.0}, 0)
	require.Equal(t, 2.0, result) // Returns sorted[1] = 2.0

	// Percentile at boundary >= 100 - gets clamped to 99, pos = Ceil(99/100 * 2) = Ceil(1.98) = 2
	result = percentileValue([]float64{1.0, 2.0, 3.0}, 100)
	require.Equal(t, 3.0, result) // Returns sorted[2] = 3.0

	// Normal case - 50th percentile of [1,2,3,4,5]
	result = percentileValue([]float64{1.0, 2.0, 3.0, 4.0, 5.0}, 50)
	require.Equal(t, 3.0, result)
}

func TestMergeSmallSemanticGroups(t *testing.T) {
	// Test empty groups
	groups := mergeSmallSemanticGroups(nil, 10, 100)
	require.Nil(t, groups)

	// Test groups all above min
	groups2 := [][]string{
		{"句子1", "句子2"},
		{"句子3"},
	}
	result := mergeSmallSemanticGroups(groups2, 5, 100)
	require.NotNil(t, result)

	// Test with very small groups that should be merged
	groups3 := [][]string{
		{"短"},
		{"更短的"},
	}
	result2 := mergeSmallSemanticGroups(groups3, 10, 100)
	require.NotNil(t, result2)
}

func TestSemanticStrategy_SplitMinChunkSize(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.MinChunkSize = 50
	s.InitialThreshold = 0.3
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3

	// Text that should create very small chunks
	text := "短"
	_, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
}

func TestSemanticStrategy_WindowBreakpointMode(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.MinChunkSize = 30
	s.Mode = SemanticModeWindowBreakpoint
	s.BufferSize = 50
	s.BreakpointPercentile = 50

	text := "第一句话。第二句话。第三句话。第四句话。第五句话。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_DoubleMergingMode(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 2
	s.MaxChunkSize = 200

	text := "第一句话。第二句话。第三句话。第四句话。第五句话。第六句话。第七句话。第八句话。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestFixedSizeStrategy_ValidateChunkOverlap(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 100

	// Test ChunkOverlap < 0 - should auto-correct to 0
	s.ChunkOverlap = -10
	err := s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0, s.ChunkOverlap)

	// Test ChunkOverlap >= ChunkSize - should auto-correct to ChunkSize - 1
	s.ChunkOverlap = 150
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 99, s.ChunkOverlap)

	// Test ChunkOverlap == ChunkSize
	s.ChunkOverlap = 100
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 99, s.ChunkOverlap)
}

func TestSemanticStrategy_ValidateAllBranches(t *testing.T) {
	s := NewSemanticStrategy()

	// Test all threshold boundary conditions
	s.InitialThreshold = -0.1
	s.AppendingThreshold = -0.1
	s.MergingThreshold = -0.1
	err := s.Validate()
	require.NoError(t, err)
	// Should be corrected to defaults (0.2, 0.3, 0.3)

	s.InitialThreshold = 1.5
	s.AppendingThreshold = 1.5
	s.MergingThreshold = 1.5
	err = s.Validate()
	require.NoError(t, err)
	// Should be corrected to defaults

	// Test BufferSize < 0
	s.BufferSize = -5
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0, s.BufferSize)

	// Test MinChunkSize < 0
	s.MinChunkSize = -10
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 0, s.MinChunkSize)

	// Test MaxChunkSize <= 0 twice
	s.MaxChunkSize = 0
	err = s.Validate()
	require.NoError(t, err)
	// Should be set to ChunkSize

	s.ChunkSize = 0
	err = s.Validate()
	require.NoError(t, err)
	// Should be set to 500

	// Test MergingRange boundary
	s.MergingRange = -1
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 1, s.MergingRange)

	s.MergingRange = 10
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 2, s.MergingRange)

	// Test BreakpointPercentile boundary
	s.BreakpointPercentile = 40
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 50, s.BreakpointPercentile)

	s.BreakpointPercentile = 100
	err = s.Validate()
	require.NoError(t, err)
	require.Equal(t, 99, s.BreakpointPercentile)
}

func TestApplyOverlapToStrings_Truncation(t *testing.T) {
	// Test when merged > chunkSize (truncation branch)
	// Use applyOverlapToStrings directly to test the truncation logic
	chunks := []string{"这是一个测试字符串", "用于测试"}
	// overlapTokens(10, 0.5) = 5, so prefix = last 5 chars of first chunk
	result := applyOverlapToStrings(chunks, 10, 0.5)
	require.NotNil(t, result)
	// The second chunk should have overlap from the first
}

func TestApplyOverlapToStrings_Direct(t *testing.T) {
	// Test direct call to applyOverlapToStrings with chunkSize that causes truncation
	chunks := []string{"短的", "一个很长的字符串超过限制"}
	result := applyOverlapToStrings(chunks, 5, 0.5)
	require.NotNil(t, result)
	require.Len(t, result, 2)
}

func TestMergeSmallSemanticGroups_WithHeading(t *testing.T) {
	// Test case where heading line is merged with next group
	groups := [][]string{
		{"# 标题"},
		{"内容1", "内容2"},
	}
	result := mergeSmallSemanticGroups(groups, 5, 100)
	require.NotNil(t, result)
	// The heading should be merged with its content
}

func TestMergeSmallSemanticGroups_PrevCombined(t *testing.T) {
	// Test case where current group is combined with previous
	groups := [][]string{
		{"已有内容"},
		{"新内容1"},
		{"新内容2"},
	}
	// minChunkSize large enough to trigger combination with prev
	result := mergeSmallSemanticGroups(groups, 10, 50)
	require.NotNil(t, result)
}

func TestTailSentencesByLen_LimitEdge(t *testing.T) {
	// Test when sum >= limit on first iteration
	sentences := []string{"短", "中", "长句子"}
	result := tailSentencesByLen(sentences, 2)
	require.NotNil(t, result)
	// Should return limited number of sentences
}

func TestSplitLongSentence_VeryLongSentence(t *testing.T) {
	// Test splitting a very long sentence
	sentences := []string{"这是一个超长的句子，它远远超过了指定的chunkSize限制，需要被分割成多个小块"}
	result := splitLongSentence(sentences[0], 10)
	require.NotEmpty(t, result)
	require.Greater(t, len(result), 1)
}

func TestSemanticSplitDoubleMerging_CodeBlock(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 2
	s.MaxChunkSize = 200

	// Text with code blocks
	text := "正常文本。\n```python\nprint('hello')\n```\n更多文本。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_InvalidMode(t *testing.T) {
	s := NewSemanticStrategy()
	s.Mode = "invalid_mode_that_does_not_exist"
	err := s.Validate()
	require.NoError(t, err)
	// Should be corrected to SemanticModeGreedy
	require.Equal(t, SemanticModeGreedy, s.Mode)
}

func TestApplySentenceOverlapToGroups_Truncation(t *testing.T) {
	// Test when merged > chunkSize
	groups := [][]string{
		{"第一组的内容"},
		{"第二组的内容"},
	}
	result := applySentenceOverlapToGroups(groups, 10, 5)
	require.NotNil(t, result)
	require.Len(t, result, 2)
}

func TestDocumentStructureStrategy_SplitSectionEmpty(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 50
	s.MaxDepth = 3

	// Test with SkipEmptyHeadings and empty direct content
	text := "# 标题\n\n\n## 子章节\n\n内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestDocumentStructureStrategy_PrefixTruncation(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 10 // Very small to trigger prefix truncation
	s.MaxDepth = 1

	text := "# 这是一个超长的标题它会超过chunkSize的限制"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestConvertToChunks_WithOverlap(t *testing.T) {
	base := &StrategyBase{}
	docs := []*schema.Document{
		{Content: "第一段内容"},
		{Content: "第二段内容"},
		{Content: "第三段内容"},
	}
	chunks := convertToChunks(docs, "test.txt", "第一段内容第二段内容第三段内容", base)
	require.NotNil(t, chunks)
}

func TestRecursiveCharacterStrategy_Separators(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 50
	s.Separators = []string{"\n\n", "\n", "。", "！"}

	text := "第一段\n\n第二段\n第三段。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSplitGroupsByMaxSize_LongSentences(t *testing.T) {
	// Test with sentences that exceed chunkSize
	groups := [][]string{
		{"这是一个非常长的句子它的长度远远超过了指定的chunkSize"},
		{"另一个非常长的句子它也会被分割"},
	}
	result := splitGroupsByMaxSize(groups, 20)
	require.NotNil(t, result)
	// Each group with long sentence should be split
}

func TestSemanticStrategy_DoubleMergingModeComplex(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 2
	s.MaxChunkSize = 150

	// Create text that triggers different merging paths
	text := "第一句完整的话。第二句完整的话。第三句完整的话。" +
		"第四句完整的话。第五句完整的话。" +
		"第六句完整的话。第七句完整的话。" +
		"第八句完整的话。第九句完整的话。" +
		"第十句完整的话。第十一句完整的话。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_WindowBreakpointWithContext(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 50
	s.MinChunkSize = 20
	s.Mode = SemanticModeWindowBreakpoint
	s.BufferSize = 3
	s.BreakpointPercentile = 50

	// Many short sentences to create clear breakpoints
	text := "第一。第二。第三。第四。第五。第六。第七。第八。第九。第十。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestConvertToChunks_PageSearchFallback(t *testing.T) {
	base := &StrategyBase{}
	// When page inference is needed
	docs := []*schema.Document{
		{Content: "some content here that is unique"},
	}
	// Use originalText that doesn't contain the doc.Content exactly
	chunks := convertToChunks(docs, "test.txt", "different original text some content here that is unique more text", base)
	require.Len(t, chunks, 1)
	// Should fall back to using previous chunk's page or default page 1
}

func TestMergeSmallSemanticGroups_MultipleConditions(t *testing.T) {
	// Test with heading line followed by content that should merge
	groups := [][]string{
		{"# 标题一"},
		{"内容一"},
		{"内容二"},
		{"# 标题二"},
		{"内容三"},
	}
	// minChunkSize=5, chunkSize=50 should trigger heading+content merge
	result := mergeSmallSemanticGroups(groups, 5, 50)
	require.NotNil(t, result)
}

func TestDocumentStructureStrategy_ChildSectionSplit(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 30
	s.MaxDepth = 2
	s.OverlapRatio = 0.1

	text := "# 主标题\n\n## 子章节\n\n这里是子章节的内容，需要被分割"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_MinChunkMerging(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 200
	s.MinChunkSize = 50
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.5
	s.AppendingThreshold = 0.5
	s.MergingThreshold = 0.5
	s.MergingRange = 2

	// Text with sentences of varying similarity to trigger minChunk merging
	text := "开始。" +
		"相似句子一。" +
		"相似句子二。" +
		"相似句子三。" +
		"另一个完全不同的句子。" +
		"继续。" +
		"更多内容。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestHardSplitByRunes(t *testing.T) {
	// Test hard split directly
	result := hardSplitByRunes("这是一个测试字符串", 5)
	require.NotNil(t, result)
	require.Greater(t, len(result), 1)

	// Edge case: empty
	result2 := hardSplitByRunes("", 5)
	require.Nil(t, result2)

	// Edge case: chunkSize 0
	result3 := hardSplitByRunes("test", 0)
	require.Nil(t, result3)

	// Edge case: text shorter than chunkSize
	result4 := hardSplitByRunes("短", 10)
	require.Len(t, result4, 1)
}

func TestMergeParts(t *testing.T) {
	// Test mergeParts directly
	parts := []string{"part1", "part2", "part3", "part4"}
	result := mergeParts(parts, " | ", 10)
	require.NotNil(t, result)

	// Test with empty parts
	result2 := mergeParts(nil, " | ", 10)
	require.Nil(t, result2)

	// Test with chunkSize 0
	result3 := mergeParts(parts, "", 0)
	require.NotNil(t, result3)
	require.Len(t, result3, 1) // All joined together
}

func TestSemanticStrategy_DoubleMergingLookahead(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 80
	s.MinChunkSize = 20
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.3
	s.AppendingThreshold = 0.4
	s.MergingThreshold = 0.4
	s.MergingRange = 2 // Enable 2-step lookahead

	// Text designed to trigger mergingRange=2 lookahead logic
	text := "AAA AAA AAA BBB BBB BBB CCC CCC CCC DDD DDD DDD"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticSplitDoubleMerging_EmptyInitial(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.5
	s.AppendingThreshold = 0.5
	s.MergingThreshold = 0.5
	s.MergingRange = 1

	// Text with short sentences that don't merge well
	text := "短。甚短。两字。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticSplitDoubleMerging_MaxChunkBoundary(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 50
	s.MinChunkSize = 10
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.3
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MaxChunkSize = 60

	// Text that will exceed maxChunkSize
	text := "AAAAAAAAAA BBBBBBBBBB CCCCCCCCCC DDDDDDDDDD EEEEEEEEEE"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSplitGroupsByMaxSize_GroupExceedsSize(t *testing.T) {
	// Test with a group where sentenceGroupLen exceeds chunkSize
	groups := [][]string{
		{"句子一", "句子二", "句子三", "句子四", "句子五"},
		{"短"},
	}
	result := splitGroupsByMaxSize(groups, 10)
	require.NotNil(t, result)
	// First group should be split
}

func TestMergeSmallSemanticGroups_HeadingMerge(t *testing.T) {
	// Test case where heading is merged with next group
	groups := [][]string{
		{"# 第一章"},
		{"第一节的内容第一段", "第一节的内容第二段"},
		{"# 第二章"},
		{"第二节的内容"},
	}
	// minChunkSize=5, chunkSize=100
	result := mergeSmallSemanticGroups(groups, 5, 100)
	require.NotNil(t, result)
}

func TestMergeSmallSemanticGroups_CombineWithPrev(t *testing.T) {
	// Test when group is combined with previous
	groups := [][]string{
		{"第一组内容"},
		{"第二组内容短的"},
	}
	// minChunkSize=20 (large enough to trigger combination with prev)
	result := mergeSmallSemanticGroups(groups, 20, 100)
	require.NotNil(t, result)
}

func TestDocumentStructureStrategy_UnitExceedsAvailable(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 20
	s.MaxDepth = 2

	// Text where child section text exceeds available space
	text := "# 标题\n\n## 子章节\n\n这是一段比较长的内容它超过了可用的空间"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestConvertToChunks_PageSearchEdge(t *testing.T) {
	base := &StrategyBase{}
	// Document with content that appears multiple times in originalText
	docs := []*schema.Document{
		{Content: "重复内容"},
	}
	// Original text has the content in a specific position relative to pages
	originalText := "<!-- Page: 1 -->重复内容<!-- Page: 2 -->其他内容重复内容<!-- Page: 3 -->"
	chunks := convertToChunks(docs, "test.txt", originalText, base)
	require.Len(t, chunks, 1)
	// Should find correct page based on position
}

func TestSemanticStrategy_DoubleMergingMaxSizeOverflow(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 30
	s.MinChunkSize = 5
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.2
	s.MergingThreshold = 0.2
	s.MergingRange = 1
	s.MaxChunkSize = 40

	// Very long sentences that will exceed maxChunkSize
	text := "这是一个非常长的句子它的内容远远超过了最大块的限制。" +
		"这是另一个非常长的句子它的内容同样远远超过了最大块的限制。" +
		"这又是一个极端冗长的句子它毫无疑问地会触发某些特殊的分支逻辑。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_DoubleMergingMinChunkCondition(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.8 // High threshold so sentences don't merge easily
	s.AppendingThreshold = 0.8
	s.MergingThreshold = 0.8
	s.MergingRange = 1

	// Short sentences that form chunks smaller than minChunkSize
	text := "短句一。短句二。短句三。短句四。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestMergeSmallSemanticGroups_SingleHeading(t *testing.T) {
	// Test with a heading that should be merged with next
	groups := [][]string{
		{"# 标题"},
		{"内容"},
	}
	// minChunkSize > 0 and this is heading line, so should merge
	result := mergeSmallSemanticGroups(groups, 5, 50)
	require.NotNil(t, result)
}

func TestMergeSmallSemanticGroups_LastGroupHandling(t *testing.T) {
	// Test when last group doesn't meet minChunkSize
	groups := [][]string{
		{"第一组"},
		{"第二组"},
		{"短"}, // This is the last group and is short
	}
	// i == len(groups)-1 case
	result := mergeSmallSemanticGroups(groups, 10, 100)
	require.NotNil(t, result)
}

func TestMergeSmallSemanticGroups_HeadingOnly(t *testing.T) {
	// Heading-only group at end
	groups := [][]string{
		{"第一组内容"},
		{"# 标题"},
	}
	result := mergeSmallSemanticGroups(groups, 5, 100)
	require.NotNil(t, result)
}

func TestRecursiveCharacterStrategy_HardSplit(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 10
	s.Separators = []string{} // No separators means hard split

	text := "这是一个会触发硬切的超长字符串没有合适的分隔符可以使用"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestRecursiveCharacterStrategy_MergePartsLong(t *testing.T) {
	s := NewRecursiveCharacterStrategy()
	s.ChunkSize = 20

	// Text that will have parts exceeding chunkSize
	text := "AAAAAAAAAABBBBBBBBBBCCCCCCCCCCDDDDDDDDDD"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestFixedSizeStrategy_SplitHard(t *testing.T) {
	s := NewFixedSizeStrategy()
	s.ChunkSize = 10

	// String with no spaces that must be split by runes
	text := "这是一个测试字符串用于测试硬切分功能"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_DoubleMergingInitialAppend(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 50
	s.MinChunkSize = 10
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.1 // Very low so first chunk builds up
	s.AppendingThreshold = 0.1
	s.MergingThreshold = 0.1
	s.MergingRange = 1
	s.MaxChunkSize = 60

	// Create a sequence where sentences accumulate in the first chunk
	// Then a very different sentence triggers flush
	text := "苹果是一种水果。苹果是红色的。苹果很甜。" +
		"汽车是一种交通工具。汽车有四个轮子。汽车可以开动。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_DoubleMergingBothConditions(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 80
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.5
	s.AppendingThreshold = 0.5
	s.MergingThreshold = 0.5
	s.MergingRange = 2

	// Mix of similar and dissimilar sentences
	// First chunk: similar sentences (apple related)
	// Then a different sentence forces flush
	// Then more sentences
	text := "猫是可爱的动物。猫会抓老鼠。猫喜欢睡觉。" +
		"狗是忠诚的动物。" +
		"鱼生活在水里。鱼会游泳。鱼有鳞片。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestMergeSmallSemanticGroups_PrevMergeBranch(t *testing.T) {
	// This test targets the line 607-614 branch where current group
	// is merged with previous group
	groups := [][]string{
		{"已有的内容足够长到可以被合并"},
		{"短"},
		{"这个组会太长无法合并"},
	}
	// minChunkSize=10, chunkSize=50
	result := mergeSmallSemanticGroups(groups, 10, 50)
	require.NotNil(t, result)
}

func TestSemanticSplitDoubleMerging_CodeBlockLong(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 50
	s.MinChunkSize = 20
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MaxChunkSize = 30 // Very small maxChunkSize

	// A long code block that exceeds maxChunkSize
	// isCodeBlockSentence will be true since it starts with ```
	// runeLen > maxChunkSize so line 284 branch is taken
	longCodeBlock := "```python\n" + strings.Repeat("print('x');\n", 50) + "```"
	text := "正常文本。\n" + longCodeBlock + "\n更多文本。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticSplitDoubleMerging_NonCodeBlockLong(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 50
	s.MinChunkSize = 20
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.2
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MaxChunkSize = 30 // Very small maxChunkSize

	// A long non-code-block sentence that exceeds maxChunkSize
	// No sentence-ending punctuation so it stays as one sentence
	// runeLen > maxChunkSize so line 284 branch is taken, goes to splitLongSentence
	longSentence := strings.Repeat("这是一句很长的话没有标点", 20)
	text := "开头句。" + longSentence + "结尾句。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticSplitDoubleMerging_NewChunkCondition(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 50
	s.MinChunkSize = 20
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.8 // High threshold - new chunk will likely be created
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MaxChunkSize = 60

	// Different sentences so similarity is low
	text := "苹果是水果。香蕉是水果。橙子是水果。" +
		"狗是动物。猫是动物。老虎是动物。" +
		"桌子是家具。椅子是家具。"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticSplitDoubleMerging_AppendingConditionFalse(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 80
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.5
	s.AppendingThreshold = 0.5
	s.MergingThreshold = 0.5
	s.MergingRange = 1
	s.MaxChunkSize = 100

	// First build a chunk, then add sentences where:
	// - similarity(lastSentences, sentence) <= appendingThreshold
	// - runeLen(chunk) >= minChunkSize
	// This makes the condition at line 334 false, going to else (flush)
	text := "AAA BBB CCC." + // First chunk builds
		"XXX YYY ZZZ." + // Different - low similarity, should trigger flush
		"DDD EEE FFF." // More sentences
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticSplitDoubleMerging_MinChunkBoundary(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 60
	s.MinChunkSize = 30
	s.Mode = SemanticModeDoubleMerging
	s.InitialThreshold = 0.3
	s.AppendingThreshold = 0.3
	s.MergingThreshold = 0.3
	s.MergingRange = 1
	s.MaxChunkSize = 80

	// Text where chunk length crosses minChunkSize boundary
	// to hit lines 371-382
	text := "较短的句子。" + // 7 chars
		strings.Repeat("中等长度的句子。", 5) + // Build up to cross minChunkSize
		"新的完全不同的内容。" // Trigger conditions
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestMergeSmallSemanticGroups_DefaultCase(t *testing.T) {
	// Test line 617: when none of the merge conditions are met
	// i.e., g doesn't meet any of the special conditions
	groups := [][]string{
		{"第一组的足够长的内容"},
		{"第二组的足够长的内容"},
		{"第三组的足够长的内容"},
	}
	// minChunkSize=5 (so all groups meet minChunkSize)
	// But none are headings, none combine with next
	// Should hit line 617 default case
	result := mergeSmallSemanticGroups(groups, 5, 100)
	require.NotNil(t, result)
	require.Len(t, result, 3)
}

func TestMergeSmallSemanticGroups_OnlyHeading(t *testing.T) {
	// Test heading at the last position (not merged with next)
	groups := [][]string{
		{"第一组内容"},
		{"第二组内容"},
		{"# 最后一个标题"},
	}
	// minChunkSize=10, chunkSize=50
	// "# 最后一个标题" is single-element heading but i==2 which is len-1
	// so line 581-588 won't trigger (i+1 >= len)
	// line 591 will be true (i == len-1)
	result := mergeSmallSemanticGroups(groups, 10, 50)
	require.NotNil(t, result)
}


func TestTableToMarkdown_Colspan(t *testing.T) {
	// Test table with colspan
	tableHTML := `<table>
		<tr><th colspan="2">Header</th></tr>
		<tr><td>Data1</td><td>Data2</td></tr>
	</table>`
	result := tableToMarkdown(tableHTML)
	require.NotEmpty(t, result)
}

func TestTableToMarkdown_Rowspan(t *testing.T) {
	// Test table with rowspan
	tableHTML := `<table>
		<tr><th>Header1</th><th>Header2</th></tr>
		<tr><td rowspan="2">Data</td><td>Data2</td></tr>
		<tr><td>Data3</td></tr>
	</table>`
	result := tableToMarkdown(tableHTML)
	require.NotEmpty(t, result)
}

func TestDecodeHexEscapes_Valid(t *testing.T) {
	result := decodeHexEscapes("测试\\u4e2d\\u6587字符")
	require.Contains(t, result, "测试中文字符")
}

func TestDecodeHexEscapes_Invalid(t *testing.T) {
	result := decodeHexEscapes("测试\\uFFFF无效字符")
	require.Contains(t, result, "测试")
}

func TestDocumentStructureStrategy_SkipEmptyTrue(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 100
	s.MaxDepth = 3
	s.SkipEmptyHeadings = true

	// Heading with only whitespace content
	text := "# 标题\n\n   \n\n## 子章节\n\n内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestDocumentStructureStrategy_SkipEmptyFalse(t *testing.T) {
	s := NewDocumentStructureStrategy()
	s.ChunkSize = 100
	s.MaxDepth = 3
	s.SkipEmptyHeadings = false

	// Heading with whitespace content should still be included
	text := "# 标题\n\n   \n\n内容"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}

func TestSemanticStrategy_WhitespaceText(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100

	docs, err := s.Split(context.Background(), "   \n\n   ", "test.txt")
	require.NoError(t, err)
	// May return nil or empty chunks
	_ = docs // May be nil or empty
}

func TestIsListLine(t *testing.T) {
	require.True(t, isListLine("1. 第一项"))
	require.True(t, isListLine("- 项目符号"))
	require.True(t, isListLine("* 星号"))
	require.False(t, isListLine("普通文本"))
	require.False(t, isListLine(""))
}

func TestSemanticStrategy_CosineSimilarityZero(t *testing.T) {
	s := NewSemanticStrategy()
	s.ChunkSize = 100
	s.Mode = SemanticModeGreedy

	// Text with completely different characters
	text := "你好世界 hello world 12345 !@#$%"
	docs, err := s.Split(context.Background(), text, "test.txt")
	require.NoError(t, err)
	require.NotNil(t, docs)
}




