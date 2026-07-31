package jsonschema

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/go-json-experiment/json"
)

// Rat wraps a big.Rat to enable custom JSON marshaling and unmarshaling.
type Rat struct {
	*big.Rat
}

// jsonNumber preserves the source representation of numbers decoded into any.
type jsonNumber string

func (n *jsonNumber) UnmarshalJSON(data []byte) error {
	*n = jsonNumber(data)
	return nil
}

func (n jsonNumber) MarshalJSON() ([]byte, error) {
	return []byte(n), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for Rat.
func (r *Rat) UnmarshalJSON(data []byte) error {
	var tmp any
	if err := unmarshalJSONExact(data, &tmp); err != nil {
		return err
	}

	converted, err := convertToBigRat(tmp)
	if err != nil {
		return err
	}

	r.Rat = converted
	return nil
}

// MarshalJSON implements the json.Marshaler interface for Rat.
func (r *Rat) MarshalJSON() ([]byte, error) {
	formattedValue := FormatRat(r)
	if strings.Contains(formattedValue, "/") {
		// Output as a JSON string if it still contains a fraction
		return json.Marshal(formattedValue)
	}
	// Output as a JSON number
	return []byte(formattedValue), nil
}

// convertToBigRat converts various types to big.Rat.
func convertToBigRat(data any) (*big.Rat, error) {
	var str string
	switch v := data.(type) {
	case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		str = fmt.Sprint(v)
	case jsonNumber:
		str = string(v)
	case string:
		str = v
	default:
		return nil, ErrUnsupportedRatType
	}

	numRat := new(big.Rat)
	if _, ok := numRat.SetString(str); !ok {
		return nil, ErrRatConversion
	}
	return numRat, nil
}

// NewRat creates a new Rat instance from a given value.
func NewRat(value any) *Rat {
	converted, err := convertToBigRat(value)
	if err != nil {
		return nil
	}
	return &Rat{converted}
}

func numberRat(value any) (*Rat, bool) {
	if number, ok := value.(jsonNumber); ok {
		rat := NewRat(number)
		return rat, rat != nil
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, false
	}

	var rat *Rat
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rat = NewRat(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rat = NewRat(rv.Uint())
	case reflect.Float32, reflect.Float64:
		rat = NewRat(rv.Float())
	default:
		return nil, false
	}
	return rat, rat != nil
}

// FormatRat formats a Rat as a string.
func FormatRat(r *Rat) string {
	if r == nil {
		return "null"
	}

	// Check if the Rat is an integer
	if r.IsInt() {
		return r.Num().String() // Output as a plain integer string
	}

	// Format as a decimal maintaining precision
	dec := r.FloatString(10) // You might adjust precision as needed

	// Trim unnecessary trailing zeros and decimal point if no fractional part
	trimmedDec := strings.TrimRight(dec, "0")
	trimmedDec = strings.TrimRight(trimmedDec, ".")

	if trimmedDec == "" {
		return "0" // correct trimming edge case of "0.0000"
	}

	return trimmedDec
}
