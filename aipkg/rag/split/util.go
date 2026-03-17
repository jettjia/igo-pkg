package split

import (
	"encoding/json"
	"math"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"github.com/jettjia/igo-pkg/aipkg/schema"
)

var (
	urlRegex   = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	tableRegex = regexp.MustCompile(`(?s)<table>.*?</table>`)

	spaceOnlyRegex = regexp.MustCompile(` {2,}`)
	newline3Regex  = regexp.MustCompile(`\n{3,}`)
)

func runeLen(s string) int {
	return len([]rune(s))
}

func overlapTokens(chunkSize int, overlapRatio float64) int {
	if chunkSize <= 0 {
		return 0
	}
	if overlapRatio <= 0 {
		return 0
	}
	return int(math.Round(float64(chunkSize) * overlapRatio))
}

func preProcessText(text string, base *StrategyBase) string {
	processed := text

	processed = tableRegex.ReplaceAllStringFunc(processed, func(tableHTML string) string {
		jsonBytes, err := tableToJSON(tableHTML)
		if err != nil {
			return tableHTML // 转换失败则返回原样
		}
		return string(jsonBytes)
	})

	if base.RemoveURLAndEmail {
		processed = urlRegex.ReplaceAllString(processed, "")
		processed = emailRegex.ReplaceAllString(processed, "")
	}

	processed = strings.ReplaceAll(processed, "\r\n", "\n")

	if base.NormalizeWhitespace {
		processed = strings.ReplaceAll(processed, "\t", " ")
		processed = spaceOnlyRegex.ReplaceAllString(processed, " ")
		processed = newline3Regex.ReplaceAllString(processed, "\n\n")
	}

	return processed
}

func getNodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var f func(node *html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return b.String()
}

func tableToJSON(tableHTML string) ([]byte, error) {
	doc, err := html.Parse(strings.NewReader(tableHTML))
	if err != nil {
		return nil, err
	}

	var headers []string
	var dataRows [][]string
	headerProcessed := false

	var findRows func(*html.Node)
	findRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var cells []string
			isHeaderRow := !headerProcessed

			// In the header row, look for <th> or <td>
			if isHeaderRow {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && (c.Data == "th" || c.Data == "td") {
						headers = append(headers, strings.TrimSpace(getNodeText(c)))
					}
				}
				headerProcessed = true
			} else { // For data rows, look for <td>
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "td" {
						cells = append(cells, strings.TrimSpace(getNodeText(c)))
					}
				}
				if len(cells) > 0 {
					dataRows = append(dataRows, cells)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findRows(c)
		}
	}

	findRows(doc)

	var result []map[string]string
	for _, row := range dataRows {
		item := make(map[string]string)
		for i, cell := range row {
			if i < len(headers) {
				item[headers[i]] = cell
			}
		}
		result = append(result, item)
	}

	return json.Marshal(result)
}

func applyTrimSpaceIfNeeded(text string, base *StrategyBase) string {
	if base.TrimSpace {
		return strings.TrimSpace(text)
	}
	return text
}

func newDocument(content string, title string, depth int) *schema.Document {
	doc := &schema.Document{
		Content: content,
	}
	if title != "" {
		doc.Title = title
	}
	if depth > 0 {
		doc.Depth = depth
	}
	return doc
}

func applyOverlapToStrings(chunks []string, chunkSize int, overlapRatio float64) []string {
	if len(chunks) <= 1 {
		return chunks
	}
	overlap := overlapTokens(chunkSize, overlapRatio)
	if overlap <= 0 {
		return chunks
	}

	out := make([]string, 0, len(chunks))
	out = append(out, chunks[0])
	for i := 1; i < len(chunks); i++ {
		prev := []rune(out[i-1])
		next := chunks[i]

		prefix := ""
		if len(prev) > 0 {
			start := len(prev) - overlap
			if start < 0 {
				start = 0
			}
			prefix = string(prev[start:])
		}

		merged := prefix + next
		if runeLen(merged) > chunkSize && chunkSize > 0 {
			mergedRunes := []rune(merged)
			merged = string(mergedRunes[len(mergedRunes)-chunkSize:])
		}
		out = append(out, merged)
	}
	return out
}
