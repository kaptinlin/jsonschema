package jsonschema

import (
	"fmt"
	"testing"
)

// Compare mutable builder validation with compiling the same document once.
// The unused definitions expose readiness-check cost independently of evaluation.
func BenchmarkConstructorLifecycle(b *testing.B) {
	for _, definitions := range []int{0, 1, 100, 1000} {
		b.Run(fmt.Sprintf("defs=%d", definitions), func(b *testing.B) {
			build := func() *Schema {
				value := Integer()
				defs := make(map[string]*Schema, definitions)
				if definitions > 0 {
					defs["value"] = value
					value = Ref("#/$defs/value")
					for i := 1; i < definitions; i++ {
						defs[fmt.Sprintf("unused%d", i)] = String()
					}
				}
				return Object(Defs(defs), Prop("value", value), Required("value"))
			}
			source := build()
			data, err := source.MarshalJSON()
			if err != nil {
				b.Fatal(err)
			}
			compiled, err := NewCompiler().Compile(data)
			if err != nil {
				b.Fatal(err)
			}
			instance := map[string]any{"value": 3}
			for name, schema := range map[string]*Schema{"direct": source, "compiled": compiled} {
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						if !schema.ValidateMap(instance).IsValid() {
							b.Fatal("valid instance rejected")
						}
					}
				})
			}
			b.Run("build", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = build()
				}
			})
			b.Run("compile", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := NewCompiler().Compile(data); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
