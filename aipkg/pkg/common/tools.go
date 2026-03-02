package common

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ParseLLMJsonResponse parses a JSON response from LLM, handling cases where JSON is wrapped in code blocks.
// This is useful when LLMs return responses like:
// ```json
// {"key": "value"}
// ```
// or regular JSON responses directly.
func ParseLLMJsonResponse(content string, target interface{}) error {
	// First, try to parse directly as JSON
	err := json.Unmarshal([]byte(content), target)
	if err == nil {
		return nil
	}

	// If direct parsing fails, try to extract JSON from code blocks with improved regex
	// This regex can handle multiple code blocks and different formatting styles
	re := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			// Extract the JSON content within the code block
			jsonContent := strings.TrimSpace(match[1])
			if err := json.Unmarshal([]byte(jsonContent), target); err == nil {
				return nil
			}
		}
	}

	// If no valid JSON found in code blocks, try to find JSON-like structures
	// This handles cases where the response might contain multiple JSON objects
	jsonLikeRe := regexp.MustCompile(`\{[^{}]*\}`)
	jsonLikeMatches := jsonLikeRe.FindAllString(content, -1)
	for _, jsonLike := range jsonLikeMatches {
		if err := json.Unmarshal([]byte(jsonLike), target); err == nil {
			return nil
		}
	}

	// If all attempts fail, return the original error with more context
	return err
}

func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	// Handle edge cases
	if len(slice) == 0 {
		return [][]T{}
	}

	if chunkSize <= 0 {
		panic("chunkSize must be greater than 0")
	}

	// Calculate how many sub-slices are needed
	chunks := make([][]T, 0, (len(slice)+chunkSize-1)/chunkSize)

	// Split the slice
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// MapSlice applies a function to each element of a slice and returns a new slice with the results
func MapSlice[A any, B any](in []A, f func(A) B) []B {
	out := make([]B, 0, len(in))
	for _, item := range in {
		out = append(out, f(item))
	}
	return out
}
