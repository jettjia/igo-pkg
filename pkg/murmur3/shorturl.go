package murmur3

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base62Radix = 62
)

// base62Encode 将整数编码为 base62 字符串
func base62Encode(num uint64) string {
	var encoded string
	for num > 0 {
		remainder := num % uint64(base62Radix)
		num /= uint64(base62Radix)
		encoded = string(base62Chars[remainder]) + encoded
	}
	return encoded
}

// GenerateShortUrl 使用 MurmurHash3 的 h1 生成短链 ID
func GenerateShortUrl(url string) string {
	// 使用 MurmurHash3 对 URL 进行哈希
	hasher := New128WithSeed(0)
	hasher.Write([]byte(url))
	h1, _ := hasher.Sum128()

	// 将 h1 编码为 base62 字符串
	encoded := base62Encode(h1)

	return encoded
}
