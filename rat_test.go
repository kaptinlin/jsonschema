package jsonschema

import (
	stdjson "encoding/json"
	"hash/maphash"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRatJSONRoundTripAndErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		jsonValue  string
		wantFormat string
		wantJSON   string
	}{
		{name: "integer", jsonValue: `12`, wantFormat: "12", wantJSON: `12`},
		{name: "decimal", jsonValue: `1.25`, wantFormat: "1.25", wantJSON: `1.25`},
		{name: "maximum uint64", jsonValue: `18446744073709551615`, wantFormat: "18446744073709551615", wantJSON: `18446744073709551615`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r Rat
			require.NoError(t, r.UnmarshalJSON([]byte(tt.jsonValue)))
			assert.Equal(t, tt.wantFormat, FormatRat(&r))

			data, err := r.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tt.wantJSON, string(data))
		})
	}
}

func TestRatJSONRoundTripPreservesLongFiniteDecimal(t *testing.T) {
	t.Parallel()

	original := NewRat(stdjson.Number("0.123456789012345678901"))
	require.NotNil(t, original)

	data, err := original.MarshalJSON()
	require.NoError(t, err)

	var roundTrip Rat
	require.NoError(t, roundTrip.UnmarshalJSON(data))
	assert.Zero(t, original.Cmp(roundTrip.Rat), "marshaled token %s changed the mathematical value", data)
}

func TestRatRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jsonValue string
	}{
		{name: "quoted integer", jsonValue: `"1"`},
		{name: "quoted decimal", jsonValue: `"1.25"`},
		{name: "quoted fraction", jsonValue: `"1/3"`},
		{name: "invalid numeric string", jsonValue: `"nope"`},
		{name: "null", jsonValue: `null`},
		{name: "unsupported object", jsonValue: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r Rat
			require.Error(t, r.UnmarshalJSON([]byte(tt.jsonValue)))
		})
	}
}

func TestNewRatAndFormatRat(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "null", FormatRat(nil))
	assert.Nil(t, NewRat(struct{}{}))

	r := NewRat("5/2")
	require.NotNil(t, r)
	assert.Equal(t, "2.5", FormatRat(r))
}

func TestFormatRatIsLossless(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "0.123456789012345678901", FormatRat(NewRat(stdjson.Number("0.123456789012345678901"))))
	assert.Equal(t, "1/3", FormatRat(NewRat("1/3")))
}

func TestRatJSONRejectsNonTerminatingDecimal(t *testing.T) {
	t.Parallel()

	data, err := NewRat("1/3").MarshalJSON()
	require.ErrorIs(t, err, ErrRatConversion)
	assert.Nil(t, data)
}

func TestMalformedJSONNumberIsNeverReinterpreted(t *testing.T) {
	fraction := stdjson.Number("1/2")

	for _, value := range []stdjson.Number{fraction, "01", "+1", "bad", " 1", "1 "} {
		assert.Nil(t, NewRat(value))
		assert.False(t, (&Schema{Type: SchemaType{"number"}}).Validate(value).IsValid())
		assert.False(t, (&Schema{Type: SchemaType{"string"}}).Validate(value).IsValid())
		_, err := marshalJSON(value)
		assert.Error(t, err)
	}

	assert.False(t, Enum(fraction).Validate(0.5).IsValid())
	assert.False(t, Const(fraction).Validate(0.5).IsValid())
	assert.True(t, Array(UniqueItems(true)).Validate([]any{fraction, 0.5}).IsValid())
	assert.False(t, deepEqualJSON(fraction, 0.5))

	seed := maphash.MakeSeed()
	assert.NotEqual(t, jsonValueHash(seed, fraction), jsonValueHash(seed, 0.5))
}
