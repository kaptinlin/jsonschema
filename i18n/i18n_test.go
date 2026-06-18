package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
	"github.com/kaptinlin/jsonschema/i18n"
)

func TestNewSupportsAllEmbeddedLocales(t *testing.T) {
	t.Parallel()

	locales := []string{"en", "de-DE", "es-ES", "fr-FR", "ja-JP", "ko-KR", "pt-BR", "zh-Hans", "zh-Hant"}
	for _, locale := range locales {
		t.Run(locale, func(t *testing.T) {
			t.Parallel()

			translator, err := i18n.New(locale)
			require.NoError(t, err)

			message, ok := translator.Translate("string_too_short", map[string]any{"min_length": 3})
			assert.True(t, ok, "built-in catalog should translate string_too_short")
			assert.NotEmpty(t, message)
		})
	}
}

func TestNewReusesLoadedBundle(t *testing.T) {
	t.Parallel()

	first, err := i18n.New("en")
	require.NoError(t, err)
	second, err := i18n.New("en")
	require.NoError(t, err)

	params := map[string]any{"min_length": 3}
	firstMessage, firstOK := first.Translate("string_too_short", params)
	secondMessage, secondOK := second.Translate("string_too_short", params)

	assert.True(t, firstOK)
	assert.True(t, secondOK)
	assert.Equal(t, firstMessage, secondMessage)
}

func TestNewRejectsUnknownLocale(t *testing.T) {
	t.Parallel()

	translator, err := i18n.New("xx-XX")
	assert.Nil(t, translator)
	require.ErrorIs(t, err, i18n.ErrUnsupportedLocale)
	assert.Contains(t, err.Error(), "zh-Hans", "error should list the supported locales")
}

func TestTranslateLocalizesEvaluationError(t *testing.T) {
	t.Parallel()

	zh, err := i18n.New("zh-Hans")
	require.NoError(t, err)

	evalErr := jsonschema.NewEvaluationError(
		"minLength",
		"string_too_short",
		"Value should be at least {min_length} characters",
		map[string]any{"min_length": 3},
	)

	assert.Equal(t, "值应至少为 3 个字符", evalErr.Localize(zh))
}

func TestTranslateMissingKeyFallsBackToEnglish(t *testing.T) {
	t.Parallel()

	zh, err := i18n.New("zh-Hans")
	require.NoError(t, err)

	_, ok := zh.Translate("nonexistent_error_code", nil)
	assert.False(t, ok)

	evalErr := jsonschema.NewEvaluationError("format", "invalid_json", "Invalid JSON format")
	assert.Equal(t, "Invalid JSON format", evalErr.Localize(zh))
}

func TestLocalizedValidationWorkflow(t *testing.T) {
	t.Parallel()

	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "minLength": 3},
			"age": {"type": "integer", "minimum": 20}
		},
		"required": ["name", "age"]
	}`

	schema, err := jsonschema.NewCompiler().Compile([]byte(schemaJSON))
	require.NoError(t, err)

	result := schema.Validate(map[string]any{
		"name": "Jo",
		"age":  18,
	})
	require.False(t, result.IsValid())

	englishErrors := result.DetailedErrors()
	require.NotEmpty(t, englishErrors)

	zh, err := i18n.New("zh-Hans")
	require.NoError(t, err)

	chineseErrors := result.LocalizedDetailedErrors(zh)
	assert.Len(t, chineseErrors, len(englishErrors),
		"localized errors should have same count as English")
	assert.Contains(t, chineseErrors, "/name/minLength")
	assert.Equal(t, "值应至少为 3 个字符", chineseErrors["/name/minLength"])

	list := result.ToLocalizedList(zh)
	assert.False(t, list.Valid)

	// Every supported locale renders the same error set; only the language changes.
	for _, locale := range []string{"ja-JP", "fr-FR", "de-DE"} {
		translator, err := i18n.New(locale)
		require.NoError(t, err)
		assert.Len(t, result.LocalizedDetailedErrors(translator), len(englishErrors),
			"locale %s should render the same error count as English", locale)
	}
}
