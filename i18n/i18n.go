// Package i18n provides jsonschema.Translator implementations backed by the
// library's built-in locale catalog. Import it only when you need localized
// validation messages; the root package stays free of any translation
// framework.
package i18n

import (
	"embed"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	goi18n "github.com/kaptinlin/go-i18n"

	"github.com/kaptinlin/jsonschema"
)

// ErrUnsupportedLocale is returned by New for a locale outside the built-in catalog.
var ErrUnsupportedLocale = errors.New("unsupported locale")

//go:embed locales/*.json
var localesFS embed.FS

// locales lists the built-in translation catalogs; "en" first is the default
// fallback locale for the bundle.
var locales = []string{"en", "de-DE", "es-ES", "fr-FR", "ja-JP", "ko-KR", "pt-BR", "zh-Hans", "zh-Hant"}

// loadBundle parses the embedded catalogs once; the bundle is read-only afterwards.
var loadBundle = sync.OnceValues(func() (*goi18n.I18n, error) {
	bundle := goi18n.NewBundle(
		goi18n.WithDefaultLocale("en"),
		goi18n.WithLocales(locales...),
	)

	if err := bundle.LoadFS(localesFS, "locales/*.json"); err != nil {
		return nil, fmt.Errorf("load embedded locales: %w", err)
	}

	return bundle, nil
})

// New returns a Translator bound to the given locale. Each Translator renders
// exactly one locale; an unknown locale is an error here rather than a silent
// fall back to English.
func New(locale string) (jsonschema.Translator, error) {
	if !slices.Contains(locales, locale) {
		return nil, fmt.Errorf("%w: %q (supported: %s)", ErrUnsupportedLocale, locale, strings.Join(locales, ", "))
	}

	bundle, err := loadBundle()
	if err != nil {
		return nil, err
	}

	return &translator{localizer: bundle.NewLocalizer(locale)}, nil
}

// translator adapts a go-i18n Localizer to the jsonschema.Translator interface.
type translator struct {
	localizer *goi18n.Localizer
}

func (t *translator) Translate(code string, params map[string]any) (string, bool) {
	result := t.localizer.Lookup(code, goi18n.Vars(params))
	if result.Source == goi18n.TranslationSourceMissing {
		return "", false
	}
	return result.Text, true
}
