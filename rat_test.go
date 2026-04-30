package jsonschema

import (
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
		{name: "fraction string", jsonValue: `"1/3"`, wantFormat: "0.3333333333", wantJSON: `0.3333333333`},
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

func TestRatRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jsonValue string
		wantErr   error
	}{
		{name: "invalid numeric string", jsonValue: `"nope"`, wantErr: ErrRatConversion},
		{name: "unsupported object", jsonValue: `{}`, wantErr: ErrUnsupportedRatType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r Rat
			err := r.UnmarshalJSON([]byte(tt.jsonValue))
			require.ErrorIs(t, err, tt.wantErr)
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
