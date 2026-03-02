package xerror

import (
	"fmt"
	"runtime"
)

func RecoverWithStack() {
	fmt.Println(string(stack()))
}

// stack get all stack errors
func stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
