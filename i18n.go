package jsonschema

import (
	"embed"

	"github.com/kaptinlin/go-i18n"
)

//go:embed locales/*.json
var localesFS embed.FS

func GetI18n() (*i18n.I18n, error) {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
	)

	err := bundle.LoadFS(localesFS, "locales/*.json")

	return bundle, err
}
