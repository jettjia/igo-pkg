package xerror

import (
	"fmt"
	"testing"
)

// go test -v -run Test_NewError ./
func Test_NewError(t *testing.T) {
	err := NewError(BadRequestErr, "请求的参数中，有特殊字符", nil)

	errData := GetError("zh-CN", "../xi18n/i18n", err)
	fmt.Println(errData.HttpCode)
	fmt.Println(errData.Code)
	fmt.Println(errData.Message)
	fmt.Println(errData.Detail)
}
