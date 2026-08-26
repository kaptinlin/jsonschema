package jsonschema

import (
	stdjson "encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"encoding/json/jsontext"
)

// Rat wraps a big.Rat to enable custom JSON marshaling and unmarshaling.
type Rat struct {
	*big.Rat
}

// UnmarshalJSON decodes a JSON number into r. Quoted numeric strings are rejected.
func (r *Rat) UnmarshalJSON(data []byte) error {
	var number stdjson.Number
	if err := unmarshalJSON(data, &number); err != nil {
		return err
	}

	converted, ok := numberRat(number)
	if !ok {
		return ErrUnsupportedRatType
	}
	if converted == nil {
		return ErrRatConversion
	}
	r.Rat = converted.Rat
	return nil
}

// MarshalJSON encodes r as an exact JSON number. It returns an error when r has
// no finite decimal representation.
func (r *Rat) MarshalJSON() ([]byte, error) {
	formatted, err := formatRatJSON(r)
	if err != nil {
		return nil, err
	}
	return []byte(formatted), nil
}

// NewRat creates a new Rat instance from a given value.
func NewRat(value any) *Rat {
	if text, ok := value.(string); ok {
		rat, _ := parseRat(text)
		return rat
	}
	rat, _ := numberRat(value)
	return rat
}

func numberRat(value any) (*Rat, bool) {
	if number, ok := value.(stdjson.Number); ok {
		if _, ok := jsonNumberToken(number); !ok {
			return nil, true
		}
		rat, _ := parseRat(number.String())
		return rat, true
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, false
	}

	var text string
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		text = strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		text = strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		text = strconv.FormatFloat(rv.Float(), 'g', -1, rv.Type().Bits())
	default:
		return nil, false
	}
	rat, _ := parseRat(text)
	return rat, true
}

func jsonNumberToken(number stdjson.Number) (jsontext.Value, bool) {
	raw := jsontext.Value(number.String())
	if len(raw) == 0 || (raw[0] != '-' && (raw[0] < '0' || raw[0] > '9')) ||
		raw[len(raw)-1] < '0' || raw[len(raw)-1] > '9' {
		return nil, false
	}
	return raw, raw.Kind() == '0' && raw.IsValid()
}

func parseRat(text string) (*Rat, bool) {
	number := new(big.Rat)
	if _, ok := number.SetString(text); !ok {
		return nil, false
	}
	return &Rat{number}, true
}

func formatRatJSON(r *Rat) (string, error) {
	if r == nil || r.Rat == nil {
		return "null", nil
	}
	if r.IsInt() {
		return r.Num().String(), nil
	}

	denominator := new(big.Int).Set(r.Denom())
	twos := denominator.TrailingZeroBits()
	denominator.Rsh(denominator, twos)

	five := big.NewInt(5)
	quotient := new(big.Int)
	remainder := new(big.Int)
	var fives uint
	for {
		quotient.QuoRem(denominator, five, remainder)
		if remainder.Sign() != 0 {
			break
		}
		denominator.Set(quotient)
		fives++
	}
	if !denominator.IsInt64() || denominator.Int64() != 1 {
		return "", fmt.Errorf("%w: %s has no finite decimal representation", ErrRatConversion, r.RatString())
	}

	precision := max(twos, fives)
	if precision > uint(^uint(0)>>1) {
		return "", fmt.Errorf("%w: decimal precision exceeds int range", ErrRatConversion)
	}
	return r.FloatString(int(precision)), nil
}

// FormatRat formats r without losing its mathematical value. Non-terminating
// decimals use fraction notation.
func FormatRat(r *Rat) string {
	formatted, err := formatRatJSON(r)
	if err == nil {
		return formatted
	}
	return r.RatString()
}
