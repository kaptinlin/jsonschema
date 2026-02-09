# DetailedErrors 多语言支持演示

本示例演示 `DetailedErrors` 方法与多语言系统的完美集成。

## 功能特性

- ✅ **完整多语言支持** - 支持9种语言的错误消息
- ✅ **一致的错误数量** - 所有语言返回相同数量的详细错误
- ✅ **保持路径信息** - 错误路径在所有语言中保持一致
- ✅ **简单易用** - 只需一行代码获取本地化错误

## 支持的语言

1. **English (en)** - 默认语言
2. **简体中文 (zh-Hans)** - 简体中文
3. **繁体中文 (zh-Hant)** - 繁体中文  
4. **日本語 (ja-JP)** - 日语
5. **한국어 (ko-KR)** - 韩语
6. **Français (fr-FR)** - 法语
7. **Deutsch (de-DE)** - 德语
8. **Español (es-ES)** - 西班牙语
9. **Português (pt-BR)** - 葡萄牙语

## 使用方法

### 基本用法
```go
// 默认英语错误 (无需传nil，更简洁)
errors := result.DetailedErrors()

// 本地化错误
i18n, _ := jsonschema.I18n()
localizer := i18n.NewLocalizer("zh-Hans")
localizedErrors := result.DetailedErrors(localizer)
```

### 完整示例
```go
func validateWithMultipleLanguages(schema *jsonschema.Schema, data any) {
    result := schema.Validate(data)
    if !result.IsValid() {
        // 英语
        englishErrors := result.DetailedErrors()

        // 中文
        i18n, _ := jsonschema.I18n()
        zhLocalizer := i18n.NewLocalizer("zh-Hans")
        chineseErrors := result.DetailedErrors(zhLocalizer)
        
        // 显示对比
        fmt.Println("English:", englishErrors)
        fmt.Println("中文:", chineseErrors)
    }
}
```

## 运行示例

```bash
cd examples/multilingual-errors
go run main.go
```

## 预期输出

```
=== DetailedErrors 多语言支持演示 ===

1. English (Default):
   /name/minLength: Value should be at least 3 characters
   /age/minimum: -5 should be at least 0
   /email/format: Value does not match format email

2. 简体中文:
   /name/minLength: 值应至少为 3 个字符
   /age/minimum: -5 应至少为 0
   /email/format: 值不符合格式 email

3. 日本語:
   /name/minLength: 値は少なくとも 3 文字である必要があります
   /age/minimum: -5 は少なくとも 0 である必要があります
   /email/format: 値がフォーマット email と一致しません

=== 错误数量统计 ===
English errors: 3
Chinese errors: 3
Japanese errors: 3
French errors: 3
German errors: 3
```

## 架构优势

1. **JSON Schema 规范合规** - 保持完整的验证语义
2. **统一的错误路径** - 所有语言使用相同的字段路径格式
3. **参数化翻译** - 支持动态参数插值 `{property}`, `{minimum}` 等
4. **向后兼容** - 不影响现有的 `result.Errors` 和 `result.ToList()` 
5. **高性能** - 基于相同的底层数据结构，无额外开销