// Package jsonschema is a JSON Schema Draft 2020-12 backend for Go. It compiles
// schemas, validates instances, and applies defaults during unmarshaling.
//
// The main entry point is [Schema.Validate]. It accepts raw JSON bytes, decoded
// maps, or Go structs and dispatches internally:
//
//	schema, err := jsonschema.NewCompiler().Compile(schemaBytes)
//	if err != nil {
//		// handle compilation error
//	}
//	result := schema.Validate(input)
//	if !result.IsValid() {
//		// inspect result.Errors
//	}
//
// [Schema.ValidateJSON], [Schema.ValidateMap], and [Schema.ValidateStruct] are
// type-specific fast paths with the same semantics; reach for them when the
// input type is known and the extra dispatch or conversion shows up in a
// profile.
//
// Compilation and unmarshaling errors are returned to callers. The package
// does not expose Must* entry points or public APIs that panic on invalid
// input.
//
// Scope: this package is a JSON Schema validator. Building higher-level
// products on top of it — schema registries, CLI protocols, form renderers,
// approval workflows — belongs in adapter layers above the library.
//
// Credit to https://github.com/santhosh-tekuri/jsonschema for format validators.
package jsonschema
