package jsonschema

import (
	"encoding/binary"
	"hash/maphash"
	"reflect"
	"slices"
)

const (
	hashNull byte = iota
	hashFalse
	hashTrue
	hashNumber
	hashString
	hashArray
	hashObject
	hashUnsupported
)

type jsonVisit struct {
	typeOf   reflect.Type
	pointer  uintptr
	length   int
	capacity int
}

type jsonComparisonState struct {
	a []jsonVisit
	b []jsonVisit
}

// evaluateUniqueItems checks if all elements in the array are unique when the "uniqueItems" property is set to true.
// According to the JSON Schema Draft 2020-12:
//   - If "uniqueItems" is false, the data always validates successfully.
//   - If "uniqueItems" is true, the data validates successfully only if all elements in the array are unique.
//
// This implementation uses hash-based comparison for O(n) average complexity.
// Each item is hashed using maphash, and deep equality is only checked on hash collisions.
//
// Reference: https://json-schema.org/draft/2020-12/json-schema-validation#name-uniqueitems
func evaluateUniqueItems(schema *Schema, data []any) *EvaluationError {
	// If uniqueItems is false or not set, no validation is needed
	if schema.UniqueItems == nil || !*schema.UniqueItems {
		return nil
	}

	// Determine the array length to validate
	maxLength := len(data)

	// If items is false, only validate items defined by prefixItems
	if schema.Items != nil && schema.Items.Boolean != nil && !*schema.Items.Boolean {
		if schema.PrefixItems != nil {
			maxLength = min(len(schema.PrefixItems), len(data))
		} else {
			maxLength = 0
		}
	}

	// If there are 0 or 1 items, they are always unique
	if maxLength <= 1 {
		return nil
	}

	// Use hash-based uniqueness check
	hashes := make(map[uint64][]int, maxLength) // hash -> indices
	seed := maphash.MakeSeed()

	for i := range maxLength {
		item := data[i]
		var h maphash.Hash
		h.SetSeed(seed)
		hashJSONValue(&h, item)
		hashValue := h.Sum64()

		// Check for hash collisions
		if indices := hashes[hashValue]; len(indices) > 0 {
			for _, j := range indices {
				if deepEqualJSON(item, data[j]) {
					return NewEvaluationError("uniqueItems", "unique_items_mismatch",
						"Array items at indices {index1} and {index2} are not unique", map[string]any{
							"index1": j,
							"index2": i,
						})
				}
			}
		}
		hashes[hashValue] = append(hashes[hashValue], i)
	}

	return nil
}

// hashJSONValue writes a deterministic hash of a JSON value to the hash.
func hashJSONValue(h *maphash.Hash, v any) {
	var inline [8]jsonVisit
	hashJSONValueState(h, v, inline[:0])
}

func hashJSONValueState(h *maphash.Hash, v any, active []jsonVisit) {
	if number, ok := numberRat(v); ok {
		if number == nil {
			_ = h.WriteByte(hashUnsupported)
			return
		}
		writeHashString(h, hashNumber, number.RatString())
		return
	}

	switch val := v.(type) {
	case nil:
		_ = h.WriteByte(hashNull)

	case bool:
		if val {
			_ = h.WriteByte(hashTrue)
		} else {
			_ = h.WriteByte(hashFalse)
		}

	case string:
		writeHashString(h, hashString, val)

	case []any:
		var ok bool
		active, ok = enterJSONValue(active, reflect.ValueOf(val))
		if !ok {
			_ = h.WriteByte(hashUnsupported)
			return
		}
		writeHashLength(h, hashArray, len(val))
		for _, item := range val {
			hashJSONValueState(h, item, active)
		}

	case map[string]any:
		var ok bool
		active, ok = enterJSONValue(active, reflect.ValueOf(val))
		if !ok {
			_ = h.WriteByte(hashUnsupported)
			return
		}
		keys := make([]string, 0, len(val))
		for key := range val {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		writeHashLength(h, hashObject, len(keys))
		for _, key := range keys {
			writeHashString(h, hashString, key)
			hashJSONValueState(h, val[key], active)
		}

	default:
		hashJSONValueReflect(h, reflect.ValueOf(v), active)
	}
}

func writeHashLength(h *maphash.Hash, tag byte, length int) {
	_ = h.WriteByte(tag)
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(length)) //nolint:gosec // Length framing only.
	_, _ = h.Write(buf[:])
}

func writeHashString(h *maphash.Hash, tag byte, value string) {
	writeHashLength(h, tag, len(value))
	_, _ = h.WriteString(value)
}

// hashJSONValueReflect handles hashing for types that need reflection.
func hashJSONValueReflect(h *maphash.Hash, rv reflect.Value, active []jsonVisit) {
	var ok bool
	rv, ok = indirectJSONValue(rv)
	if !ok {
		_ = h.WriteByte(hashUnsupported)
		return
	}
	if !rv.IsValid() {
		_ = h.WriteByte(hashNull)
		return
	}

	if rv.CanInterface() {
		if number, ok := numberRat(rv.Interface()); ok {
			if number == nil {
				_ = h.WriteByte(hashUnsupported)
				return
			}
			writeHashString(h, hashNumber, number.RatString())
			return
		}
	}

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			_ = h.WriteByte(hashTrue)
		} else {
			_ = h.WriteByte(hashFalse)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return // Numeric kinds are normalized before the switch.

	case reflect.String:
		writeHashString(h, hashString, rv.String())

	case reflect.Slice, reflect.Array:
		var ok bool
		active, ok = enterJSONValue(active, rv)
		if !ok {
			_ = h.WriteByte(hashUnsupported)
			return
		}
		writeHashLength(h, hashArray, rv.Len())
		for i := range rv.Len() {
			hashJSONValueReflect(h, rv.Index(i), active)
		}

	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			_ = h.WriteByte(hashUnsupported)
			return
		}
		active, ok = enterJSONValue(active, rv)
		if !ok {
			_ = h.WriteByte(hashUnsupported)
			return
		}
		keys := make([]string, 0, rv.Len())
		for _, key := range rv.MapKeys() {
			keys = append(keys, key.String())
		}
		slices.Sort(keys)
		writeHashLength(h, hashObject, len(keys))
		for _, key := range keys {
			writeHashString(h, hashString, key)
			hashJSONValueReflect(h, mapValueByString(rv, key), active)
		}

	case reflect.Invalid, reflect.Interface, reflect.Pointer, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Struct, reflect.UnsafePointer:
		_ = h.WriteByte(hashUnsupported)
	}
}

// deepEqualJSON performs deep equality comparison for JSON values.
func deepEqualJSON(a, b any) bool {
	var aInline, bInline [8]jsonVisit
	state := jsonComparisonState{a: aInline[:0], b: bInline[:0]}
	return deepEqualJSONReflect(reflect.ValueOf(a), reflect.ValueOf(b), state)
}

// deepEqualJSONReflect performs reflection-based deep equality.
func deepEqualJSONReflect(a, b reflect.Value, state jsonComparisonState) bool {
	var aOK, bOK bool
	a, aOK = indirectJSONValue(a)
	b, bOK = indirectJSONValue(b)
	if !aOK || !bOK {
		return false
	}
	if !a.IsValid() || !b.IsValid() {
		return !a.IsValid() && !b.IsValid()
	}

	if a.CanInterface() && b.CanInterface() {
		ra, aIsNumber := numberRat(a.Interface())
		rb, bIsNumber := numberRat(b.Interface())
		if aIsNumber || bIsNumber {
			return aIsNumber && bIsNumber && ra != nil && rb != nil && ra.Cmp(rb.Rat) == 0
		}
	}

	switch {
	case a.Kind() == reflect.Bool && b.Kind() == reflect.Bool:
		return a.Bool() == b.Bool()

	case a.Kind() == reflect.String && b.Kind() == reflect.String:
		return a.String() == b.String()

	case isArrayKind(a.Kind()) && isArrayKind(b.Kind()):
		if a.Len() != b.Len() {
			return false
		}
		if !state.enter(a, b) {
			return false
		}
		for i := range a.Len() {
			if !deepEqualJSONReflect(a.Index(i), b.Index(i), state) {
				return false
			}
		}
		return true

	case a.Kind() == reflect.Map && b.Kind() == reflect.Map:
		if a.Type().Key().Kind() != reflect.String || b.Type().Key().Kind() != reflect.String || a.Len() != b.Len() {
			return false
		}
		if !state.enter(a, b) {
			return false
		}
		for _, key := range a.MapKeys() {
			bValue := mapValueByString(b, key.String())
			if !bValue.IsValid() || !deepEqualJSONReflect(a.MapIndex(key), bValue, state) {
				return false
			}
		}
		return true
	}

	return false
}

func indirectJSONValue(value reflect.Value) (reflect.Value, bool) {
	var inline [8]jsonVisit
	pointers := inline[:0]
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return reflect.Value{}, true
		}
		if value.Kind() == reflect.Pointer {
			var ok bool
			pointers, ok = enterJSONValue(pointers, value)
			if !ok {
				return reflect.Value{}, false
			}
		}
		value = value.Elem()
	}
	return value, true
}

func (state *jsonComparisonState) enter(a, b reflect.Value) bool {
	var aOK bool
	state.a, aOK = enterJSONValue(state.a, a)
	if !aOK {
		return false
	}
	var bOK bool
	state.b, bOK = enterJSONValue(state.b, b)
	return bOK
}

func referenceJSONVisit(value reflect.Value) (jsonVisit, bool) {
	if value.Kind() != reflect.Map && value.Kind() != reflect.Pointer && value.Kind() != reflect.Slice {
		return jsonVisit{}, false
	}
	return makeJSONVisit(value), true
}

func makeJSONVisit(value reflect.Value) jsonVisit {
	visit := jsonVisit{typeOf: value.Type(), pointer: value.Pointer()}
	if value.Kind() == reflect.Slice {
		visit.length = value.Len()
		visit.capacity = value.Cap()
	}
	return visit
}

func enterJSONValue(active []jsonVisit, value reflect.Value) ([]jsonVisit, bool) {
	visit, tracked := referenceJSONVisit(value)
	if !tracked {
		return active, true
	}
	if slices.Contains(active, visit) {
		return active, false
	}
	return append(active, visit), true
}

func isArrayKind(kind reflect.Kind) bool {
	return kind == reflect.Array || kind == reflect.Slice
}

func mapValueByString(value reflect.Value, key string) reflect.Value {
	mapKey := reflect.New(value.Type().Key()).Elem()
	mapKey.SetString(key)
	return value.MapIndex(mapKey)
}
