package chunker

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	spaceRegex      = regexp.MustCompile(`\s+`)
	urlRegex        = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex      = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	ellipsisRegex   = regexp.MustCompile(`\.{3,}|……+`)
	emptySpaceRegex = regexp.MustCompile(`\s{2,}`)                  // 匹配2个或更多连续的空白字符
	pageMarkerRegex = regexp.MustCompile(`-{2,}\s*第(\d+)页\s*-{2,}`) // 匹配 "--- 第1页 ---" 格式，并捕获页码数字
)

// calculateOverlap 计算重叠大小
func calculateOverlap(chunkSize int, overlapRatio float64) int {
	// 确保重叠比例在合理范围内
	if overlapRatio < 0 {
		overlapRatio = 0
	} else if overlapRatio > 1 {
		overlapRatio = 1
	}
	return int(float64(chunkSize) * overlapRatio)
}

// removeURLAndEmail 移除URL和邮箱
func removeURLAndEmail(text string) string {
	text = urlRegex.ReplaceAllString(text, "")
	text = emailRegex.ReplaceAllString(text, "")
	return text
}

// removeEllipsis 移除连续的点（...）
func removeEllipsis(text string) string {
	return ellipsisRegex.ReplaceAllString(text, "")
}

// preProcessContent 预处理文本内容
func preProcessContent(text string, s *RecursiveParagraphStrategy) string {
	// 进行内容预处理
	processedText := text

	// 移除URL和邮箱
	if s.TrimURLAndEmail {
		processedText = removeURLAndEmail(processedText)
	}

	// 移除多余的点
	if s.TrimEllipsis {
		processedText = removeEllipsis(processedText)
	}

	// 规范化空白字符
	if s.TrimSpace {
		processedText = normalizeSpace(processedText)
	}

	// 规范化转义空白字符
	processedText = normalizeEscapedWhitespace(processedText)

	return processedText
}

// normalizeEscapedWhitespace 将文本中以文字形式存在的转义空白序列替换为单个空格
func normalizeEscapedWhitespace(text string) string {
	// 先将常见的文字转义序列替换为空格
	text = strings.ReplaceAll(text, "\\n", " ")
	text = strings.ReplaceAll(text, "\\r", " ")
	text = strings.ReplaceAll(text, "\\t", " ")
	// 再将连续空白折叠为单个空格
	text = spaceRegex.ReplaceAllString(text, " ")
	return text
}

// extractPageNumber 从文本中提取页码
func extractPageNumber(text string) int {
	// 用正则查找文本中所有的页码标记
	allMatches := pageMarkerRegex.FindAllStringSubmatch(text, -1)
	if len(allMatches) > 0 {
		// 取最后一个匹配的页码（最新的页码）
		lastMatch := allMatches[len(allMatches)-1]
		if len(lastMatch) > 1 {
			pageNum := 0
			fmt.Sscanf(lastMatch[1], "%d", &pageNum)
			return pageNum
		}
	}
	return 0 // 没有找到页码
}

// removePageMarkers 从文本中移除页码标记
func removePageMarkers(text string) string {
	// 移除所有 "--- 第N页 ---" 格式的标记
	text = pageMarkerRegex.ReplaceAllString(text, "")
	// 清理多余的空白
	text = strings.TrimSpace(text)
	return text
}

// normalizeSpace 规范化空白字符
func normalizeSpace(text string) string {
	return emptySpaceRegex.ReplaceAllString(text, " ")
}
