# DetailedErrors Multilingual Support Demo

This example demonstrates how the `DetailedErrors` method integrates with the optional `i18n` subpackage.

## Features

- ✅ **Full multilingual support** — error messages in 9 languages
- ✅ **Consistent error counts** — every language returns the same number of detailed errors
- ✅ **Stable error paths** — field paths are identical across languages
- ✅ **Opt-in localization** — the root package has zero Intl dependencies; import the `i18n` subpackage only when you need it

## Supported Locales

1. **English (en)** — default
2. **Chinese, Simplified (zh-Hans)**
3. **Chinese, Traditional (zh-Hant)**
4. **Japanese (ja-JP)**
5. **Korean (ko-KR)**
6. **French (fr-FR)**
7. **German (de-DE)**
8. **Spanish (es-ES)**
9. **Portuguese, Brazil (pt-BR)**

## Usage

### Basic Usage

```go
import "github.com/kaptinlin/jsonschema/i18n"

// Default English errors (root package, zero Intl dependencies)
errors := result.DetailedErrors()

// Localized errors: one Translator per locale
zh, _ := i18n.New("zh-Hans")
localizedErrors := result.LocalizedDetailedErrors(zh)
```

### Complete Example

```go
func validateWithMultipleLanguages(schema *jsonschema.Schema, data any) {
    result := schema.Validate(data)
    if !result.IsValid() {
        // English
        englishErrors := result.DetailedErrors()

        // Chinese
        zh, _ := i18n.New("zh-Hans")
        chineseErrors := result.LocalizedDetailedErrors(zh)

        // Compare
        fmt.Println("English:", englishErrors)
        fmt.Println("Chinese:", chineseErrors)
    }
}
```

## Run the Example

```bash
cd examples/multilingual-errors
go run main.go
```

## Expected Output

```
=== DetailedErrors Multilingual Support Demo ===

1. English (Default):
   /name/minLength: Value should be at least 3 characters
   /age/minimum: -5 should be at least 0
   /email/format: Value does not match format email

2. Chinese (Simplified):
   /name/minLength: 值应至少为 3 个字符
   /age/minimum: -5 应至少为 0
   /email/format: 值不符合格式 email

3. Japanese:
   /name/minLength: 値は少なくとも 3 文字である必要があります
   /age/minimum: -5 は少なくとも 0 である必要があります
   /email/format: 値がフォーマット email と一致しません

=== Error Count Statistics ===
English errors: 3
Chinese (Simplified) errors: 3
Japanese errors: 3
French errors: 3
German errors: 3
```

## Architecture Notes

1. **JSON Schema compliant** — full validation semantics are preserved
2. **Uniform error paths** — every language uses the same field path format
3. **Parameterized translations** — dynamic interpolation of `{property}`, `{minimum}`, and friends
4. **Zero-cost by default** — pure validation users never compile or link the Intl chain; localization lives in the `i18n` subpackage
5. **Fallback that never fails** — missing translations fall back to the built-in English message
