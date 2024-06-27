package jsonschema

import (
	"fmt"
	"math/big"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/goccy/go-json"
)

// replace substitutes placeholders in a template string with actual parameter values.
func replace(template string, params map[string]interface{}) string {
	for key, value := range params {
		placeholder := "{" + key + "}"
		template = strings.ReplaceAll(template, placeholder, fmt.Sprint(value))
	}

	return template
}

// mergeIntMaps merges two integer maps. The values in the second map overwrite the first where keys overlap.
func mergeIntMaps(map1, map2 map[int]bool) map[int]bool {
	for key, value := range map2 {
		map1[key] = value
	}
	return map1
}

// mergeStringMaps merges two string maps. The values in the second map overwrite the first where keys overlap.
func mergeStringMaps(map1, map2 map[string]bool) map[string]bool {
	for key, value := range map2 {
		map1[key] = value
	}
	return map1
}

// getDataType identifies the JSON schema type for a given Go value.
func getDataType(v interface{}) string {
	switch v := v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case json.Number:
		// Try as an integer first
		if _, ok := new(big.Int).SetString(string(v), 10); ok {
			return "integer" // json.Number without a decimal part, can be considered an integer
		}
		// Fallback to big float to check if it is an integer
		if bigFloat, ok := new(big.Float).SetString(string(v)); ok {
			if _, acc := bigFloat.Int(nil); acc == big.Exact {
				return "integer"
			}
			return "number"
		}
	case float32, float64:
		// Convert to big.Float to check if it can be considered an integer
		bigFloat := new(big.Float).SetFloat64(reflect.ValueOf(v).Float())
		if _, acc := bigFloat.Int(nil); acc == big.Exact {
			return "integer" // Treated as integer if no fractional part
		}
		return "number"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case []bool, []json.Number, []float32, []float64, []int, []int8, []int16, []int32, []int64, []uint, []uint8, []uint16, []uint32, []uint64, []string:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
	return "unknown"
}

// getURLScheme extracts the scheme component of a URL string.
func getURLScheme(urlStr string) string {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsedUrl.Scheme
}

// isValidURI verifies if the provided string is a valid URI.
func isValidURI(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

// resolveRelativeURI resolves a relative URI against a base URI.
func resolveRelativeURI(baseURI, relativeURL string) string {
	if isAbsoluteURI(relativeURL) {
		return relativeURL
	}
	base, err := url.Parse(baseURI)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return relativeURL // Return the original if there's a base URL parsing error
	}
	rel, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL // Return the original if there's a relative URL parsing error
	}
	return base.ResolveReference(rel).String()
}

// isAbsoluteURI checks if the given URL is absolute.
func isAbsoluteURI(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// getBaseURI extracts the base URL from an $id URI, falling back if not valid.
func getBaseURI(id string) string {
	if id == "" {
		return ""
	}
	u, err := url.Parse(id)
	if err != nil {
		return ""
	}
	if strings.HasSuffix(u.Path, "/") {
		return u.String()
	}
	u.Path = path.Dir(u.Path)
	if u.Path == "." {
		u.Path = "/"
	}
	if u.Path != "/" && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	if u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.String()
}

// splitRef separates a URI into its base URI and anchor parts.
func splitRef(ref string) (baseURI string, anchor string) {
	parts := strings.SplitN(ref, "#", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return ref, ""
}

// isJSONPointer checks if a string is a JSON Pointer.
func isJSONPointer(s string) bool {
	return strings.HasPrefix(s, "/")
}
