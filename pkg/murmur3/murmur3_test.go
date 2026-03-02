package murmur3

import (
	"fmt"
	"testing"
)

// go test -v -run=Test_murmur128 .
func Test_murmur128(t *testing.T) {
	x, y := Sum128([]byte("chocolate-covered-espresso-beans"))
	fmt.Printf("%x %x\n", x, y)
}
