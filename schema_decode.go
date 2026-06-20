package jsonschema

import (
	"bytes"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func decodeExclusiveBound(keyword string, raw jsontext.Value, target **Rat, legacy *jsontext.Value) error {
	if len(raw) == 0 {
		return nil
	}

	trimmed := bytes.TrimSpace(raw)
	if bytes.Equal(trimmed, []byte("true")) || bytes.Equal(trimmed, []byte("false")) {
		*legacy = append((*legacy)[:0], raw...)
		return nil
	}

	rat := &Rat{}
	if err := json.Unmarshal(raw, rat); err != nil {
		return fmt.Errorf("%s: %w", keyword, err)
	}
	*target = rat
	return nil
}
