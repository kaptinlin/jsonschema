# Multiple Input Types Example

这个示例展示了 JSON Schema 验证器如何处理不同类型的输入数据，以及如何使用 `Unmarshal` 功能应用默认值。

## 🎯 功能演示

### 1. 输入类型验证 (Input Type Validation)
展示验证器支持的各种输入类型：
- **JSON Bytes** (`[]byte`) - 最常用的 API 输入格式
- **Go Struct** - 直接验证 Go 结构体
- **Map Data** - 验证 `map[string]interface{}`
- **JSON String** - 通过转换为 `[]byte` 处理 JSON 字符串

### 2. 带默认值的反序列化 (Unmarshal with Defaults)
演示如何使用 `Unmarshal` 方法：
- 验证数据有效性
- 应用 schema 中定义的默认值
- 反序列化到 Go 结构体

### 3. 最佳实践 (Best Practices)
提供使用建议：
- JSON 字符串处理方式
- 错误处理模式
- 默认值的正确使用
- 验证失败示例

## 🚀 运行示例

```bash
go run main.go
```

## 📝 主要改进

相比之前的版本，这个优化版本：

1. **结构更清晰** - 分为3个独立的演示部分
2. **代码更简洁** - 减少重复代码，使用辅助函数
3. **输出更友好** - 使用 emoji 和清晰的格式
4. **专注核心功能** - 移除冗余示例，突出重要特性
5. **易于理解** - 每个部分都有明确的目的和说明

## 🔍 代码亮点

- **模块化设计** - 每个功能独立函数实现
- **表格驱动测试** - 使用结构化数据演示多个例子
- **一致性API** - 遵循 `json.Unmarshal` 的输入模式
- **错误处理** - 展示正确的错误检查模式 
