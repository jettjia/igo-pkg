package xerror

import (
	"context"

	"github.com/jettjia/go-pkg/pkg/xi18n"
)

var (
	i18nInfo = make(map[int]string)
)

func init() {
	i18nInfo[BadRequestErr] = "BadRequestErr"
	i18nInfo[UnauthorizedErr] = "UnauthorizedErr"
	i18nInfo[ForbiddenErr] = "ForbiddenErr"
	i18nInfo[NotFoundErr] = "NotFoundErr"
	i18nInfo[ConflictErr] = "ConflictErr"
	i18nInfo[InternalServerErr] = "InternalServerErr"
}

func GetI18nValue(lang string, path string, code int) (codeMsg string) {
	if lang == "" {
		lang = "zh-CN"
	}
	iManager := xi18n.NewI18n(xi18n.WithLang(lang), xi18n.WithPath(path))
	codeValue := i18nInfo[code]
	return iManager.T(context.TODO(), codeValue)
}

func LoadTranslation(i18nLoad map[int]string) {
	for key, val := range i18nLoad {
		i18nInfo[key] = val
	}
}
