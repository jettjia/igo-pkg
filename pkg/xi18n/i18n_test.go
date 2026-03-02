package xi18n

import (
	"context"
	"fmt"
	"testing"
)

func Test_NewI18n_1(t *testing.T) {
	fmt.Println(NewI18n().T(context.TODO(), "BadRequestErr"))
}

func Test_NewI18n_setLang(t *testing.T) {
	fmt.Println(NewI18n(WithLang("zh-CN")).T(context.TODO(), "BadRequestErr"))
}

func Test_NewI18n_panic(t *testing.T) {
	fmt.Println(NewI18n(WithLang("zh-CN")).T(context.TODO(), "InternalServerErr"))
}
