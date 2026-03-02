package php2go

import (
	"fmt"
	"testing"
)

func TestIfExpression(t *testing.T) {
	r := IfExpression(true, "是", "否")
	fmt.Println(r)
}

func TestIfExpression2(t *testing.T) {
	a := 2
	r := IfExpression(a%2 == 0, "偶数", "奇数")
	fmt.Println(r)
}
