# Exact JSON Numbers

## Overview

The package preserves the mathematical value of JSON numbers across schema
compilation, validation, defaults-aware unmarshaling, embedded JSON content, and
schema serialization. This specification owns dynamic number representation,
numeric normalization, JSON equality and hashing, and custom codec ownership.

Backward compatibility with a dynamic `float64` representation is not a goal.
This contract does not change the separation between validation and
`Schema.Unmarshal`, nor does it make `format` assertions active by default.

## Terminology

| Term | Definition | Not | Owner / Evidence |
|------|------------|-----|------------------|
| **Untyped JSON number** | A number decoded into `any` without a concrete numeric destination. | A caller-provided Go numeric value. | Package JSON codec in `compiler.go`. |
| **Exact number** | An untyped JSON number represented by its valid JSON token as `encoding/json.Number`. | A binary floating-point approximation. | `compiler.go`, `rat.go`. |
| **JSON equality** | Mathematical equality for numbers and structural equality for JSON arrays and string-keyed objects. | Go type identity or `reflect.DeepEqual`. | `unique_items.go`. |
| **Custom codec** | Encoder or decoder explicitly installed on a `Compiler`. | Package-owned default behavior. | `Compiler.WithEncoderJSON`, `Compiler.WithDecoderJSON`. |

## Usage / API Usage

### Exact dynamic values

```go
import stdjson "encoding/json"

var constant jsonschema.ConstValue
if err := stdjson.Unmarshal([]byte(`18446744073709551615`), &constant); err != nil {
	return err
}

number := constant.Value.(stdjson.Number)
fmt.Println(number.String()) // 18446744073709551615
```

### Numeric callbacks

```go
zero := jsonschema.NewRat(0)
hundred := jsonschema.NewRat(100)

compiler.RegisterFormat("percentage", func(value any) bool {
	number := jsonschema.NewRat(value)
	return number != nil &&
		number.Cmp(zero.Rat) >= 0 &&
		number.Cmp(hundred.Rat) <= 0
}, "number")
```

`NewRat` accepts native and named integer and floating-point types,
`encoding/json.Number`, and its documented constructor-only string forms.
Validation strings remain JSON strings and are not reclassified as numbers.

## Contracts

### Package-owned JSON codec

- Every package-owned decode into `any` produces `encoding/json.Number` for a
  JSON number, recursively through arrays and objects.
- Package-owned encoding writes a valid `encoding/json.Number` as a JSON number
  token and rejects malformed values.
- `Rat` JSON decoding accepts number tokens only. Quoted integers, decimals,
  and fractions are not numeric keyword values.
- `Rat` JSON encoding emits an exact finite decimal or returns an error when
  the rational has no finite decimal representation; it never emits a rounded
  approximation or changes the JSON type to string.
- Schema `const`, `enum`, `default`, `examples`, and preserved extension values
  use the same representation and round-trip without precision loss.
- Instance validation, `Schema.Unmarshal`, and the built-in `application/json`
  media handler use the compiler's configured instance codec. Schema documents
  always use the package codec so schema semantics are stable.

### Numeric semantics

- Native and named `int*`, `uint*`, `float32`, and `float64` values plus
  `encoding/json.Number` share one numeric-to-rational interpretation.
- Type classification and the numeric keywords use that interpretation.
- Integral floating-point values are classified mathematically, without an
  overflowing integer conversion.
- Invalid `encoding/json.Number` values are recognized as malformed numbers;
  they are never reinterpreted as strings or fraction syntax.
- `NewRat` string and fraction parsing is an explicit in-memory constructor
  capability. It does not widen JSON numeric keyword syntax.
- `FormatRat` is lossless display text: finite values use exact decimal text
  and non-terminating values use exact fraction notation.

### Equality and hashing

- Numeric equality is mathematical across supported Go representations.
- Arrays compare by length and ordered elements.
- Maps whose key kind is `string`, including named string types, compare by key
  text and recursively by value.
- Equality is total for caller input: nil and unsupported or cyclic values do
  not panic. Cyclic values are unsupported; shared acyclic values retain JSON
  tree semantics.
- Hash framing distinguishes JSON kinds and container boundaries. If two
  supported values compare equal, they produce the same hash.

### Unmarshaling

- `Schema.Unmarshal` applies defaults but does not validate.
- Untyped destinations preserve `encoding/json.Number`.
- Concrete scalar, struct, pointer, and typed-map destinations receive values
  according to their declared Go types through the configured codec.
- A JSON number is not silently coerced into a Go string.

### Custom codec ownership

- `WithDecoderJSON` replaces the dynamic types and precision behavior used for
  instance validation, `Schema.Unmarshal`, and the built-in JSON media handler.
- `WithEncoderJSON` replaces intermediate encoding used by
  `Schema.Unmarshal`.
- Callers that install either hook own its compatibility with
  `encoding/json.Number`; the package does not add a hidden fallback path that
  changes the explicit override.

### Foreign serialization boundaries

- `encoding/json.Number` remains the package-owned representation until the
  value crosses into another format.
- A serializer for another format owns the exact target representation. It must
  emit a valid number without precision loss or return an error; compatibility
  with that serializer is not a reason to coerce the upstream value to
  `float64` or string.
- Exact output means preserving the mathematical value and number type at that
  boundary. A later decoder owns the concrete in-memory type it chooses.

## Failure Semantics

- Invalid JSON produces the existing compilation, validation, or unmarshal
  error for the invoked public entry point.
- Malformed `encoding/json.Number` values fail package-owned encoding and are
  unsupported for numeric conversion.
- A `Rat` without a finite decimal representation fails JSON marshaling with
  `ErrRatConversion` rather than changing value or JSON type.
- Unsupported reflection values compare unequal and hash deterministically;
  they never cause a validation panic.
- Destination conversion failures from `Schema.Unmarshal` retain the existing
  wrapped error chain.

## Prior Decisions

- **Decision**: Use `encoding/json.Number` as the only public dynamic number
  representation. **Why**: callers can name it and its token preserves exact
  decimal input. **Rejected**: a private package type, a second public number
  type, or a parallel `float64` compatibility model.
- **Decision**: Normalize numbers through rational semantics. **Why**: JSON
  Schema compares mathematical values across input paths. **Rejected**:
  keyword-specific conversions and conversion through `float64`.
- **Decision**: Keep arbitrary rational strings constructor-only. **Why**: JSON
  Schema numeric keywords require JSON number tokens, and JSON can exactly
  encode only finite decimals. **Rejected**: accepting quoted fractions or
  serializing non-terminating rationals as rounded numbers or strings.
- **Decision**: Treat custom codecs as authoritative. **Why**: an explicit hook
  is a policy boundary. **Rejected**: bypassing the hook on selected validation
  or media paths to recover package defaults.

## Forbidden

- Do not decode package-owned untyped JSON numbers through `float64`.
- Do not create another numeric conversion path for a keyword or entry point.
- Do not accept quoted values for numeric keywords or approximate a `Rat` while
  crossing a JSON boundary.
- Do not use reflection equality or formatted fallback values for JSON equality
  or hashing.
- Do not make `Schema.Unmarshal` perform validation.
- Do not override caller-installed codec semantics with a hidden exact decoder.
- Do not weaken package-owned number representation to accommodate a foreign
  serializer; adapt or reject the value at that serializer's boundary.

## Acceptance Criteria

- Exact integer boundaries, long decimals, exponent forms, nested schema
  values, embedded JSON content, and media decoding retain their value through
  package-owned codecs. Verification: focused codec and validation tests.
- All numeric keywords preserve long finite decimals across
  compile/marshal/compile. Quoted numeric keywords and non-terminating rational
  marshaling fail explicitly. Verification: compiler, schema, and Rat JSON
  tests.
- `Validate`, `ValidateJSON`, `ValidateMap`, and `ValidateStruct` agree for
  equivalent supported numeric inputs. Verification: the validation entry-point
  matrix in `validate_test.go` and integration tests in `tests/`.
- Numeric constraints, `enum`, `const`, and `uniqueItems` agree across native,
  named, and `encoding/json.Number` values. Verification: unit and property
  tests, including the equality/hash invariant fuzz target.
- Scalar, struct, pointer, `map[string]any`, and typed-map destinations preserve
  the declared destination contract. Verification: `unmarshal_test.go` on the
  Go version declared in `go.mod`.
- Custom decoder and encoder hooks remain authoritative. Verification:
  compiler, validation, media handler, and unmarshal override tests.
- Foreign serializers preserve exact numeric value and number type or fail
  explicitly. Verification: integration tests owned by the serializer adapter.
