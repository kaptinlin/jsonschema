# Internationalization (i18n) Example

This example demonstrates how to get localized validation error messages in different languages.

## Running the Example

```bash
go run examples/i18n/main.go
```

## Localization Setup

```go
// Get i18n instance
i18n, err := jsonschema.GetI18n()
if err != nil {
    log.Fatal(err)
}

// Create localizer for Simplified Chinese
localizer := i18n.NewLocalizer("zh-Hans")

// Get localized error messages
errors := result.ToLocalizeList(localizer)
```

## Example Output

Instead of technical English errors like:
```
"message": "Value must be at least 0"
```

You get localized messages like:
```
"message": "数值必须至少为 0"
```
