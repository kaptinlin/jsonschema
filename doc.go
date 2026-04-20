// Package jsonschema implements a high-performance JSON Schema Draft 2020-12
// validator for Go, providing direct struct validation, smart unmarshaling
// with defaults, and a separated validation workflow.
//
// Compilation and unmarshaling errors are returned to callers; the package does
// not expose Must* entry points or public APIs that panic on invalid input.
//
// Credit to https://github.com/santhosh-tekuri/jsonschema for format validators.
package jsonschema
