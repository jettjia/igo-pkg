package murmur3

import "testing"

// go test -v -run=Test_GenerateShortUrl .
func Test_GenerateShortUrl(t *testing.T) {
	url := "https://www.example.com/some-long-url-that-needs-to-be-shortened"
	surl := GenerateShortUrl(url)
	t.Log(surl)
}
