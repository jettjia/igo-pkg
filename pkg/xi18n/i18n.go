package xi18n

import (
	"github.com/gogf/gf/v2/i18n/gi18n"
)

type I18n struct {
	Lang string `json:"lang"` // lang "zh-CN", "zh-TW", "en"
	Path string `json:"path"` // the path specified by i18n does not read i18n in the current directory by default
}

type I18nOpt func(options *I18n)

func WithLang(lang string) I18nOpt {
	return func(op *I18n) {
		op.Lang = lang
	}
}

func WithPath(path string) I18nOpt {
	return func(op *I18n) {
		op.Path = path
	}
}

var (
	l *gi18n.Manager
)

func NewI18n(options ...I18nOpt) *gi18n.Manager {
	opts := &I18n{}
	for _, option := range options {
		option(opts)
	}

	if opts.Lang == "" {
		opts.Lang = "zh-CN"
	}

	cfg := &I18n{
		Lang: opts.Lang,
		Path: opts.Path,
	}
	l = initI18n(cfg)

	return l
}

func initI18n(cfg *I18n) *gi18n.Manager {
	i18nManager := gi18n.New()
	if cfg.Path != "" {
		if err := i18nManager.SetPath(cfg.Path); err != nil {
			panic(err)
		}
	}
	i18nManager.SetLanguage(cfg.Lang)

	return i18nManager
}
