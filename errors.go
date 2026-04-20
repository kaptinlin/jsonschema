package jsonschema

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNoLoaderRegistered reports a missing loader for a URL scheme.
	ErrNoLoaderRegistered = errors.New("no loader registered for scheme")

	// ErrDataRead reports a data read failure.
	ErrDataRead = errors.New("data read failed")

	// ErrNetworkFetch reports a network fetch failure.
	ErrNetworkFetch = errors.New("network fetch failed")

	// ErrInvalidStatusCode reports an invalid HTTP status code.
	ErrInvalidStatusCode = errors.New("invalid http status code")

	// ErrFileWrite reports a file write failure.
	ErrFileWrite = errors.New("file write failed")

	// ErrFileCreation reports a file creation failure.
	ErrFileCreation = errors.New("file creation failed")

	// ErrDirectoryCreation reports a directory creation failure.
	ErrDirectoryCreation = errors.New("directory creation failed")

	// ErrContentWrite reports a content write failure.
	ErrContentWrite = errors.New("content write failed")

	// ErrInvalidFilenamePath reports an invalid filename path.
	ErrInvalidFilenamePath = errors.New("invalid filename path")
)

var (
	// ErrJSONUnmarshal reports a JSON unmarshal failure.
	ErrJSONUnmarshal = errors.New("json unmarshal failed")

	// ErrXMLUnmarshal reports an XML unmarshal failure.
	ErrXMLUnmarshal = errors.New("xml unmarshal failed")

	// ErrYAMLUnmarshal reports a YAML unmarshal failure.
	ErrYAMLUnmarshal = errors.New("yaml unmarshal failed")

	// ErrJSONDecode reports a JSON decode failure.
	ErrJSONDecode = errors.New("json decode failed")

	// ErrSourceEncode reports a source encoding failure.
	ErrSourceEncode = errors.New("source encode failed")

	// ErrIntermediateJSONDecode reports an intermediate JSON decode failure.
	ErrIntermediateJSONDecode = errors.New("intermediate json decode failed")

	// ErrDataEncode reports a data encoding failure.
	ErrDataEncode = errors.New("data encode failed")

	// ErrNestedValueEncode reports a nested value encoding failure.
	ErrNestedValueEncode = errors.New("nested value encode failed")
)

var (
	// ErrSchemaCompilation reports a schema compilation failure.
	ErrSchemaCompilation = errors.New("schema compilation failed")

	// ErrReferenceResolution reports a reference resolution failure.
	ErrReferenceResolution = errors.New("reference resolution failed")

	// ErrGlobalReferenceResolution reports a global reference resolution failure.
	ErrGlobalReferenceResolution = errors.New("global reference resolution failed")

	// ErrDefinitionResolution reports a `$defs` resolution failure.
	ErrDefinitionResolution = errors.New("definition resolution failed")

	// ErrItemResolution reports an array item resolution failure.
	ErrItemResolution = errors.New("item resolution failed")

	// ErrJSONPointerSegmentDecode reports a JSON Pointer segment decode failure.
	ErrJSONPointerSegmentDecode = errors.New("json pointer segment decode failed")

	// ErrJSONPointerSegmentNotFound reports a missing JSON Pointer segment.
	ErrJSONPointerSegmentNotFound = errors.New("json pointer segment not found")

	// ErrInvalidSchemaType reports an invalid schema type.
	ErrInvalidSchemaType = errors.New("invalid schema type")

	// ErrSchemaIsNil reports a nil schema.
	ErrSchemaIsNil = errors.New("schema is nil")

	// ErrSchemaInternalsIsNil reports missing schema internals.
	ErrSchemaInternalsIsNil = errors.New("schema internals is nil")

	// ErrRegexValidation reports an invalid regular expression in a schema.
	ErrRegexValidation = errors.New("regex validation failed")
)

// RegexPatternError provides structured context for invalid regular expressions discovered during schema compilation.
type RegexPatternError struct {
	// Keyword identifies the JSON Schema keyword containing the invalid pattern.
	// Examples: "pattern", "patternProperties".
	Keyword string

	// Location is the JSON Pointer path to the keyword instance.
	// Example: "#/properties/email/pattern".
	Location string

	// Pattern is the regex pattern that failed to compile.
	Pattern string

	// Err is the underlying regexp compilation error.
	Err error
}

// Error formats the regex compilation error with keyword, location, and pattern context.
func (e *RegexPatternError) Error() string {
	var sb strings.Builder
	sb.WriteString("regex pattern error")

	var parts []string
	if e.Keyword != "" {
		parts = append(parts, "keyword="+e.Keyword)
	}
	if e.Location != "" {
		parts = append(parts, "location="+e.Location)
	}
	if e.Pattern != "" {
		parts = append(parts, fmt.Sprintf("pattern=%q", e.Pattern))
	}
	if len(parts) > 0 {
		sb.WriteString(" (")
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteByte(')')
	}

	if e.Err != nil {
		sb.WriteString(": ")
		sb.WriteString(e.Err.Error())
	}

	return sb.String()
}

// Unwrap returns the underlying regexp compilation error.
func (e *RegexPatternError) Unwrap() error {
	return e.Err
}

var (
	// ErrValueValidationFailed reports a value validation failure.
	ErrValueValidationFailed = errors.New("value validation failed")

	// ErrInvalidRuleFormat reports an invalid rule format.
	ErrInvalidRuleFormat = errors.New("invalid rule format")

	// ErrRuleRequiresParameter reports a missing rule parameter.
	ErrRuleRequiresParameter = errors.New("rule requires parameter")

	// ErrEmptyRuleName reports an empty rule name.
	ErrEmptyRuleName = errors.New("empty rule name")

	// ErrValidatorAlreadyExists reports a duplicate validator registration.
	ErrValidatorAlreadyExists = errors.New("validator already exists")
)

var (
	// ErrTypeConversion reports a type conversion failure.
	ErrTypeConversion = errors.New("type conversion failed")

	// ErrTimeConversion reports a time conversion failure.
	ErrTimeConversion = errors.New("time conversion failed")

	// ErrTimeParsing reports a time parsing failure.
	ErrTimeParsing = errors.New("time parsing failed")

	// ErrRatConversion reports a rat conversion failure.
	ErrRatConversion = errors.New("rat conversion failed")

	// ErrSliceConversion reports a slice conversion failure.
	ErrSliceConversion = errors.New("slice conversion failed")

	// ErrSliceElementConversion reports a slice element conversion failure.
	ErrSliceElementConversion = errors.New("slice element conversion failed")

	// ErrFirstSliceConversion reports a first slice conversion failure.
	ErrFirstSliceConversion = errors.New("first slice conversion failed")

	// ErrSecondSliceConversion reports a second slice conversion failure.
	ErrSecondSliceConversion = errors.New("second slice conversion failed")

	// ErrNilConversion reports a nil conversion failure.
	ErrNilConversion = errors.New("nil conversion failed")

	// ErrNilPointerConversion reports a nil pointer conversion failure.
	ErrNilPointerConversion = errors.New("nil pointer conversion failed")

	// ErrValueParsing reports a value parsing failure.
	ErrValueParsing = errors.New("value parsing failed")

	// ErrValueAssignment reports a value assignment failure.
	ErrValueAssignment = errors.New("value assignment failed")

	// ErrUnsupportedType reports an unsupported type.
	ErrUnsupportedType = errors.New("unsupported type")

	// ErrUnsupportedRatType reports an unsupported rat conversion type.
	ErrUnsupportedRatType = errors.New("unsupported rat type")

	// ErrUnsupportedInputType reports an unsupported input type.
	ErrUnsupportedInputType = errors.New("unsupported input type")

	// ErrUnsupportedGenerationType reports an unsupported code generation type.
	ErrUnsupportedGenerationType = errors.New("unsupported generation type")

	// ErrUnsupportedConversion reports an unsupported conversion.
	ErrUnsupportedConversion = errors.New("unsupported conversion")

	// ErrUnrepresentableType reports an unrepresentable type.
	ErrUnrepresentableType = errors.New("unrepresentable type")

	// ErrInvalidTransformType reports an invalid transform type.
	ErrInvalidTransformType = errors.New("invalid transform type")
)

var (
	// ErrExpectedStructType reports a non-struct value where a struct was required.
	ErrExpectedStructType = errors.New("expected struct type")

	// ErrStructTagParsing reports a struct tag parsing failure.
	ErrStructTagParsing = errors.New("struct tag parsing failed")

	// ErrFieldNotFound reports a missing field.
	ErrFieldNotFound = errors.New("field not found")

	// ErrFieldAssignment reports a field assignment failure.
	ErrFieldAssignment = errors.New("field assignment failed")

	// ErrFieldAnalysis reports a field analysis failure.
	ErrFieldAnalysis = errors.New("field analysis failed")

	// ErrNilDestination reports a nil destination.
	ErrNilDestination = errors.New("destination cannot be nil")

	// ErrNotPointer reports a destination that is not a pointer.
	ErrNotPointer = errors.New("destination must be a pointer")

	// ErrNilPointer reports a nil destination pointer.
	ErrNilPointer = errors.New("destination pointer cannot be nil")
)

var (
	// ErrDefaultApplication reports a default application failure.
	ErrDefaultApplication = errors.New("default application failed")

	// ErrDefaultEvaluation reports a default evaluation failure.
	ErrDefaultEvaluation = errors.New("default evaluation failed")

	// ErrArrayDefaultApplication reports an array default application failure.
	ErrArrayDefaultApplication = errors.New("array default application failed")

	// ErrFunctionCallParsing reports a function call parsing failure.
	ErrFunctionCallParsing = errors.New("function call parsing failed")
)

var (
	// ErrNilConfig reports a nil generator config.
	ErrNilConfig = errors.New("config cannot be nil")

	// ErrAnalyzerCreation reports an analyzer creation failure.
	ErrAnalyzerCreation = errors.New("analyzer creation failed")

	// ErrWriterCreation reports a writer creation failure.
	ErrWriterCreation = errors.New("writer creation failed")

	// ErrPackageAnalysis reports a package analysis failure.
	ErrPackageAnalysis = errors.New("package analysis failed")

	// ErrCodeGeneration reports a code generation failure.
	ErrCodeGeneration = errors.New("code generation failed")

	// ErrPropertyGeneration reports a property generation failure.
	ErrPropertyGeneration = errors.New("property generation failed")

	// ErrJSONSchemaTagParsing reports a jsonschema tag parsing failure.
	ErrJSONSchemaTagParsing = errors.New("json schema tag parsing failed")

	// ErrGozodTagParsing reports a gozod tag parsing failure.
	ErrGozodTagParsing = errors.New("gozod tag parsing failed")

	// ErrTemplateLoading reports a template loading failure.
	ErrTemplateLoading = errors.New("template loading failed")

	// ErrOutputDirectoryCreation reports an output directory creation failure.
	ErrOutputDirectoryCreation = errors.New("output directory creation failed")

	// ErrFieldSchemaGeneration reports a field schema generation failure.
	ErrFieldSchemaGeneration = errors.New("field schema generation failed")

	// ErrTemplateExecution reports a template execution failure.
	ErrTemplateExecution = errors.New("template execution failed")

	// ErrMainTemplateParsing reports a main template parsing failure.
	ErrMainTemplateParsing = errors.New("main template parsing failed")

	// ErrDependencyAnalysis reports a dependency analysis failure.
	ErrDependencyAnalysis = errors.New("dependency analysis failed")

	// ErrTemplateParsing reports a template parsing failure.
	ErrTemplateParsing = errors.New("template parsing failed")

	// ErrCodeFormatting reports a code formatting failure.
	ErrCodeFormatting = errors.New("code formatting failed")

	// ErrStructNodeNotFound reports a missing struct node.
	ErrStructNodeNotFound = errors.New("struct node not found")
)

var (
	// ErrArrayItemNotSchema reports a non-schema array item.
	ErrArrayItemNotSchema = errors.New("array item not schema")

	// ErrExpectedArrayOrSlice reports a value that is not an array or slice.
	ErrExpectedArrayOrSlice = errors.New("expected array or slice")

	// ErrNilPointerToArray reports a nil pointer to an array value.
	ErrNilPointerToArray = errors.New("nil pointer to array")

	// ErrNilPointerToRecord reports a nil pointer to a record value.
	ErrNilPointerToRecord = errors.New("nil pointer to record")

	// ErrNilPointerToObject reports a nil pointer to an object value.
	ErrNilPointerToObject = errors.New("nil pointer to object")

	// ErrExpectedMapStringAny reports a value that is not map[string]any.
	ErrExpectedMapStringAny = errors.New("expected map[string]any")

	// ErrNonStringKeyFound reports a non-string map key.
	ErrNonStringKeyFound = errors.New("non-string key found in map")
)

var (
	// ErrValueOverflows reports an overflow in the target type.
	ErrValueOverflows = errors.New("value overflows target type")

	// ErrEmptyInput reports an empty input.
	ErrEmptyInput = errors.New("empty input")

	// ErrNegativeValueConversion reports a negative value conversion failure.
	ErrNegativeValueConversion = errors.New("negative value conversion failed")

	// ErrNonWholeNumber reports a non-whole number.
	ErrNonWholeNumber = errors.New("not a whole number")

	// ErrInvalidSliceArrayString reports a value that is not a slice, array, or string.
	ErrInvalidSliceArrayString = errors.New("invalid slice array string")

	// ErrNilValue reports a nil value.
	ErrNilValue = errors.New("nil value")

	// ErrNilConstValue reports an unmarshal into a nil ConstValue.
	ErrNilConstValue = errors.New("cannot unmarshal into nil ConstValue")

	// ErrIPv6AddressFormat reports an invalid IPv6 address format.
	ErrIPv6AddressFormat = errors.New("ipv6 address format error")

	// ErrInvalidIPv6 reports an invalid IPv6 address.
	ErrInvalidIPv6 = errors.New("invalid ipv6 address")
)

var (
	// ErrAbsolutePathResolution reports an absolute path resolution failure.
	ErrAbsolutePathResolution = errors.New("absolute path resolution failed")

	// ErrCurrentDirectoryAccess reports a current directory access failure.
	ErrCurrentDirectoryAccess = errors.New("current directory access failed")

	// ErrPathOutsideDirectory reports a path outside the current directory.
	ErrPathOutsideDirectory = errors.New("path outside directory")
)

var (
	// ErrUnderlyingError reports an underlying error without additional context.
	ErrUnderlyingError = errors.New("underlying error")
)
