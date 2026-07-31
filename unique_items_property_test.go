package jsonschema

import (
	"context"
	stdjson "encoding/json"
	"hash/maphash"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestJSONEqualityImpliesEqualHash(t *testing.T) {
	type (
		Key         string
		NamedString string
		NamedInt    int64
	)
	var nilInt *int

	pairs := []struct {
		name string
		a    any
		b    any
	}{
		{name: "null and nil pointer", a: nil, b: nilInt},
		{name: "heterogeneous number", a: NamedInt(1), b: stdjson.Number("1.0")},
		{name: "named string", a: NamedString("value"), b: "value"},
		{name: "array and slice", a: [2]int{1, 2}, b: []any{stdjson.Number("1"), float64(2)}},
		{name: "named map key", a: map[Key]NamedInt{"answer": 42}, b: map[string]any{"answer": stdjson.Number("42.0")}},
	}

	seed := maphash.MakeSeed()
	for _, pair := range pairs {
		t.Run(pair.name, func(t *testing.T) {
			if !deepEqualJSON(pair.a, pair.b) {
				t.Fatal("values should be equal under JSON semantics")
			}
			if aHash, bHash := jsonValueHash(seed, pair.a), jsonValueHash(seed, pair.b); aHash != bHash {
				t.Fatalf("equal values have different hashes: %d != %d", aHash, bHash)
			}
		})
	}
}

func TestJSONEqualityRejectsUnsupportedValues(t *testing.T) {
	type unsupported struct{ Value int }

	values := []any{
		unsupported{Value: 1},
		map[int]string{1: "one"},
		complex(1, 2),
		make(chan int),
		func() {},
	}
	uniqueSchema := Array(UniqueItems(true))

	for _, value := range values {
		if deepEqualJSON(value, value) {
			t.Fatalf("unsupported %T value compared equal", value)
		}
		if !uniqueSchema.Validate([]any{value, value}).IsValid() {
			t.Fatalf("unsupported %T value was treated as a JSON duplicate", value)
		}
	}
}

func TestJSONEqualityRejectsCycles(t *testing.T) {
	if os.Getenv("JSONSCHEMA_CYCLE_HELPER") == "1" {
		cyclicMap := map[string]any{}
		cyclicMap["self"] = cyclicMap
		cyclicSlice := make([]any, 1)
		cyclicSlice[0] = cyclicSlice
		var cyclicInterface any
		cyclicInterface = &cyclicInterface

		seed := maphash.MakeSeed()
		uniqueSchema := Array(UniqueItems(true))
		for _, value := range []any{cyclicMap, cyclicSlice, cyclicInterface} {
			if deepEqualJSON(value, value) {
				t.Fatalf("cyclic %T value compared equal", value)
			}
			_ = jsonValueHash(seed, value)
			if !uniqueSchema.Validate([]any{value, value}).IsValid() {
				t.Fatalf("cyclic %T value was treated as a JSON duplicate", value)
			}
		}

		shared := map[string]any{"value": 1}
		a := []any{shared, shared}
		b := []any{map[string]any{"value": 1}, map[string]any{"value": 1}}
		if !deepEqualJSON(a, b) || jsonValueHash(seed, a) != jsonValueHash(seed, b) {
			t.Fatal("shared acyclic values should retain JSON tree semantics")
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestJSONEqualityRejectsCycles$")
	cmd.Env = append(os.Environ(), "JSONSCHEMA_CYCLE_HELPER=1")
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("cyclic JSON equality did not terminate: %v", ctx.Err())
	}
	if err != nil {
		t.Fatalf("cycle helper failed: %v\n%s", err, output)
	}
}

func FuzzJSONEqualityHashInvariant(f *testing.F) {
	f.Add(`1`, `1.0`)
	f.Add(`{"answer":42,"items":[true,null]}`, `{"items":[true,null],"answer":42.0}`)
	f.Add(`[1,2,3]`, `[1,2,4]`)
	f.Add(`"value"`, `null`)

	f.Fuzz(func(t *testing.T, aJSON, bJSON string) {
		var a, b any
		if unmarshalJSON([]byte(aJSON), &a) != nil || unmarshalJSON([]byte(bJSON), &b) != nil {
			return
		}

		equal := deepEqualJSON(a, b)
		if equal != deepEqualJSON(b, a) {
			t.Fatal("JSON equality is not symmetric")
		}
		if !equal {
			return
		}

		seed := maphash.MakeSeed()
		if jsonValueHash(seed, a) != jsonValueHash(seed, b) {
			t.Fatal("equal JSON values have different hashes")
		}
	})
}

func jsonValueHash(seed maphash.Seed, value any) uint64 {
	var hash maphash.Hash
	hash.SetSeed(seed)
	hashJSONValue(&hash, value)
	return hash.Sum64()
}
