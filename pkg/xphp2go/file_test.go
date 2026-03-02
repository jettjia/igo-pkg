package php2go

import (
	"fmt"
	"testing"
)

func Test_Pathinfo(t *testing.T) {
	tPathinfo := Pathinfo("/home/go/php2go.go.go", -1)
	fmt.Println(tPathinfo)
}
