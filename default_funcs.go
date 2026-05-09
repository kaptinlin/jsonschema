package jsonschema

import (
	"strconv"
	"strings"
	"time"
)

// FunctionCall represents a parsed function call with name and arguments
type FunctionCall struct {
	Name string
	Args []any
}

// parseFunctionCall parses a string to determine if it's a function call.
// It returns nil when the input is not in function-call form.
func parseFunctionCall(input string) *FunctionCall {
	if len(input) < 3 || !strings.HasSuffix(input, ")") {
		return nil
	}

	parenIndex := strings.IndexByte(input, '(')
	if parenIndex <= 0 {
		return nil
	}

	name := strings.TrimSpace(input[:parenIndex])
	rawArgs := strings.TrimSpace(input[parenIndex+1 : len(input)-1])

	var args []any
	if rawArgs != "" {
		args = parseArgs(rawArgs)
	}

	return &FunctionCall{
		Name: name,
		Args: args,
	}
}

func parseArgs(raw string) []any {
	args := make([]any, 0)

	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if i, err := strconv.ParseInt(part, 10, 64); err == nil {
			args = append(args, i)
			continue
		}

		if f, err := strconv.ParseFloat(part, 64); err == nil {
			args = append(args, f)
			continue
		}

		args = append(args, part)
	}

	return args
}

// DefaultNowFunc generates current timestamp in various formats
// This function must be manually registered by developers
func DefaultNowFunc(args ...any) (any, error) {
	format := time.RFC3339

	if len(args) > 0 {
		if f, ok := args[0].(string); ok {
			format = f
		}
	}

	return time.Now().Format(format), nil
}
