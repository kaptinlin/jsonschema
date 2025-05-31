# Internationalization Example

Demonstrates multilingual error messages using Chinese (zh-Hans) and English locales.

## What it does

- Shows validation errors in both Chinese and English
- Demonstrates the separation of validation and unmarshaling
- Shows production pattern with localized error handling

## Run

```bash
go run main.go
```

## Expected output

- Invalid data with validation errors in both languages
- Unmarshal still works despite validation failure
- Production pattern with successful validation and processing 
