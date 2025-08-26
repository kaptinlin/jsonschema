package jsonschema

import "errors"

var (
	// ErrNoLoaderRegistered is returned when no loader is registered for the specified scheme.
	ErrNoLoaderRegistered = errors.New("no loader registered for scheme")

	// ErrFailedToReadData is returned when data cannot be read from the specified URL.
	ErrFailedToReadData = errors.New("failed to read data from URL")

	// ErrJSONUnmarshalError is returned when there is an error unmarshalling JSON.
	ErrJSONUnmarshalError = errors.New("json unmarshal error")

	// ErrXMLUnmarshalError is returned when there is an error unmarshalling XML.
	ErrXMLUnmarshalError = errors.New("xml unmarshal error")

	// ErrYAMLUnmarshalError is returned when there is an error unmarshalling YAML.
	ErrYAMLUnmarshalError = errors.New("yaml unmarshal error")

	// ErrFailedToFetch is returned when there is an error fetching from the URL.
	ErrFailedToFetch = errors.New("failed to fetch from URL")

	// ErrInvalidHTTPStatusCode is returned when an invalid HTTP status code is returned.
	ErrInvalidHTTPStatusCode = errors.New("invalid HTTP status code returned")

	// ErrIPv6AddressNotEnclosed is returned when an IPv6 address is not enclosed in brackets.
	ErrIPv6AddressNotEnclosed = errors.New("ipv6 address is not enclosed in brackets")

	// ErrInvalidIPv6Address is returned when the IPv6 address is invalid.
	ErrInvalidIPv6Address = errors.New("invalid ipv6 address")

	// ErrUnsupportedTypeForRat is returned when the type is unsupported for conversion to *big.Rat.
	ErrUnsupportedTypeForRat = errors.New("unsupported type for conversion to *big.Rat")

	// ErrFailedToConvertToRat is returned when a string fails to convert to *big.Rat.
	ErrFailedToConvertToRat = errors.New("failed to convert string to *big.Rat")

	// ErrFailedToResolveGlobalReference is returned when a global reference cannot be resolved.
	ErrFailedToResolveGlobalReference = errors.New("failed to resolve global reference")

	// ErrFailedToDecodeSegmentWithJSONPointer is returned when a segment cannot be decoded.
	ErrFailedToDecodeSegmentWithJSONPointer = errors.New("failed to decode segment")

	// ErrSegmentNotFoundForJSONPointer is returned when a segment is not found in the schema context.
	ErrSegmentNotFoundForJSONPointer = errors.New("segment not found in the schema context")

	// ErrFailedToResolveReference is returned when a reference cannot be resolved.
	ErrFailedToResolveReference = errors.New("failed to resolve reference")

	// ErrFailedToResolveDefinitions is returned when definitions in $defs cannot be resolved.
	ErrFailedToResolveDefinitions = errors.New("failed to resolve definitions in $defs")

	// ErrFailedToResolveItems is returned when items in an array schema cannot be resolved.
	ErrFailedToResolveItems = errors.New("failed to resolve items")

	// ErrInvalidJSONSchemaType is returned when the JSON schema type is invalid.
	ErrInvalidJSONSchemaType = errors.New("invalid JSON schema type")

	// ErrNilConstValue is returned when trying to unmarshal into a nil ConstValue.
	ErrNilConstValue = errors.New("cannot unmarshal into nil ConstValue")

	// ErrExpectedStructType is returned when a non-struct type is provided where a struct type is expected.
	ErrExpectedStructType = errors.New("expected struct type")

	// ErrFailedToParseStructTags is returned when struct tags cannot be parsed.
	ErrFailedToParseStructTags = errors.New("failed to parse struct tags")

	// ErrUnsupportedType is returned when an unsupported type is encountered.
	ErrUnsupportedType = errors.New("unsupported type")

	// ErrFailedToCompileSchema is returned when a schema compilation fails.
	ErrFailedToCompileSchema = errors.New("failed to compile schema")

	// ErrFailedToDecodeJSON is returned when JSON decoding fails.
	ErrFailedToDecodeJSON = errors.New("failed to decode JSON")

	// ErrFailedToEncodeSource is returned when source encoding fails.
	ErrFailedToEncodeSource = errors.New("failed to encode source")

	// ErrFailedToDecodeIntermediateJSON is returned when intermediate JSON decoding fails.
	ErrFailedToDecodeIntermediateJSON = errors.New("failed to decode intermediate JSON")

	// ErrFailedToApplyDefaults is returned when applying defaults fails.
	ErrFailedToApplyDefaults = errors.New("failed to apply defaults")

	// ErrFailedToEvaluateDefaultValue is returned when evaluating default values fails.
	ErrFailedToEvaluateDefaultValue = errors.New("failed to evaluate default value")

	// ErrFailedToParseFunctionCall is returned when parsing function calls fails.
	ErrFailedToParseFunctionCall = errors.New("failed to parse function call")

	// ErrFailedToApplyArrayDefaults is returned when applying array defaults fails.
	ErrFailedToApplyArrayDefaults = errors.New("failed to apply defaults for array item")

	// ErrFailedToEncodeData is returned when data encoding fails.
	ErrFailedToEncodeData = errors.New("failed to encode data for fallback")

	// ErrFailedToSetField is returned when setting a field fails.
	ErrFailedToSetField = errors.New("failed to set field")

	// ErrFailedToEncodeNestedValue is returned when encoding nested values fails.
	ErrFailedToEncodeNestedValue = errors.New("failed to encode nested value")

	// ErrUnderlyingError is returned as a generic underlying error.
	ErrUnderlyingError = errors.New("underlying error")

	// ErrConfigCannotBeNil is returned when a configuration is nil.
	ErrConfigCannotBeNil = errors.New("config cannot be nil")

	// ErrUnsupportedTypeForGeneration is returned when an unsupported type is encountered during code generation.
	ErrUnsupportedTypeForGeneration = errors.New("unsupported type for generation")

	// ErrStructNodeNotFound is returned when a struct node cannot be found.
	ErrStructNodeNotFound = errors.New("struct node not found")

	// ErrFailedToCreateAnalyzer is returned when analyzer creation fails.
	ErrFailedToCreateAnalyzer = errors.New("failed to create analyzer")

	// ErrFailedToCreateWriter is returned when writer creation fails.
	ErrFailedToCreateWriter = errors.New("failed to create writer")

	// ErrFailedToAnalyzePackage is returned when package analysis fails.
	ErrFailedToAnalyzePackage = errors.New("failed to analyze package")

	// ErrFailedToGenerateCode is returned when code generation fails.
	ErrFailedToGenerateCode = errors.New("failed to generate code")

	// ErrFailedToGenerateProperty is returned when property generation fails.
	ErrFailedToGenerateProperty = errors.New("failed to generate property")

	// ErrFailedToParseJSONSchemaTag is returned when jsonschema tag parsing fails.
	ErrFailedToParseJSONSchemaTag = errors.New("failed to parse jsonschema tag")

	// ErrFailedToAnalyzeFields is returned when field analysis fails.
	ErrFailedToAnalyzeFields = errors.New("failed to analyze fields")

	// ErrFailedToParseGozodTag is returned when gozod tag parsing fails.
	ErrFailedToParseGozodTag = errors.New("failed to parse gozod tag")

	// ErrInvalidRuleFormat is returned when rule format is invalid.
	ErrInvalidRuleFormat = errors.New("invalid rule format")

	// ErrRuleRequiresParameter is returned when a rule requires a parameter.
	ErrRuleRequiresParameter = errors.New("rule requires parameter")

	// ErrEmptyRuleName is returned when rule name is empty.
	ErrEmptyRuleName = errors.New("empty rule name")

	// ErrFailedToLoadTemplates is returned when template loading fails.
	ErrFailedToLoadTemplates = errors.New("failed to load templates")

	// ErrFailedToCreateOutputDirectory is returned when output directory creation fails.
	ErrFailedToCreateOutputDirectory = errors.New("failed to create output directory")

	// ErrFailedToWriteFile is returned when file writing fails.
	ErrFailedToWriteFile = errors.New("failed to write file")

	// ErrFailedToGenerateFieldSchemas is returned when field schema generation fails.
	ErrFailedToGenerateFieldSchemas = errors.New("failed to generate field schemas")

	// ErrFailedToExecuteTemplate is returned when template execution fails.
	ErrFailedToExecuteTemplate = errors.New("failed to execute template")

	// ErrFailedToParseMainTemplate is returned when main template parsing fails.
	ErrFailedToParseMainTemplate = errors.New("failed to parse main template")

	// ErrFailedToAnalyzeDependencies is returned when dependency analysis fails.
	ErrFailedToAnalyzeDependencies = errors.New("failed to analyze dependencies")

	// ErrValidatorAlreadyExists is returned when a validator with the same name already exists.
	ErrValidatorAlreadyExists = errors.New("validator already exists")

	// ErrUnsupportedInputType is returned when an unsupported input type is encountered.
	ErrUnsupportedInputType = errors.New("unsupported input type")

	// ErrUnrepresentableType is returned when a type cannot be represented.
	ErrUnrepresentableType = errors.New("unrepresentable type")

	// ErrArrayItemNotSchema is returned when an array item is not a schema.
	ErrArrayItemNotSchema = errors.New("array item not schema")

	// ErrCannotConvertSliceElement is returned when slice element conversion fails.
	ErrCannotConvertSliceElement = errors.New("cannot convert slice element")

	// ErrNotSliceArrayOrString is returned when input is not a slice, array, or string.
	ErrNotSliceArrayOrString = errors.New("not slice, array, or string")

	// ErrCannotConvertFirstSlice is returned when first slice conversion fails.
	ErrCannotConvertFirstSlice = errors.New("cannot convert first slice")

	// ErrCannotConvertSecondSlice is returned when second slice conversion fails.
	ErrCannotConvertSecondSlice = errors.New("cannot convert second slice")

	// ErrCannotConvertSlice is returned when slice conversion fails.
	ErrCannotConvertSlice = errors.New("cannot convert slice")

	// ErrCannotConvertNil is returned when nil conversion fails.
	ErrCannotConvertNil = errors.New("cannot convert nil")

	// ErrInvalidTypeForTransform is returned when type is invalid for transform.
	ErrInvalidTypeForTransform = errors.New("invalid type for transform")

	// ErrUnableToConvertValue is returned when value conversion fails.
	ErrUnableToConvertValue = errors.New("unable to convert value")

	// ErrSchemaDoesNotImplementInterface is returned when schema doesn't implement required interface.
	ErrSchemaDoesNotImplementInterface = errors.New("schema does not implement required interface")

	// ErrNilPointerToArray is returned when trying to convert nil pointer to array.
	ErrNilPointerToArray = errors.New("nil pointer to array")

	// ErrExpectedArrayOrSlice is returned when array or slice is expected.
	ErrExpectedArrayOrSlice = errors.New("expected array or slice")

	// ErrNilPointerToRecord is returned when trying to convert nil pointer to record.
	ErrNilPointerToRecord = errors.New("nil pointer to record")

	// ErrNonStringKeyFound is returned when non-string key is found in map.
	ErrNonStringKeyFound = errors.New("non-string key found in map")

	// ErrExpectedMapStringAny is returned when map[string]any is expected.
	ErrExpectedMapStringAny = errors.New("expected map[string]any")

	// ErrValueValidationFailed is returned when value validation fails.
	ErrValueValidationFailed = errors.New("value validation failed")

	// ErrNilPointerToObject is returned when trying to convert nil pointer to object.
	ErrNilPointerToObject = errors.New("nil pointer to object")

	// ErrFieldNotFound is returned when field is not found.
	ErrFieldNotFound = errors.New("field not found")

	// ErrCannotAssignValue is returned when value assignment fails.
	ErrCannotAssignValue = errors.New("cannot assign value")

	// ErrSchemaIsNil is returned when schema is nil.
	ErrSchemaIsNil = errors.New("schema is nil")

	// ErrSchemaInternalsIsNil is returned when schema internals is nil.
	ErrSchemaInternalsIsNil = errors.New("schema internals is nil")

	// ErrCannotConvertNilPointer is returned when nil pointer conversion fails.
	ErrCannotConvertNilPointer = errors.New("cannot convert nil pointer")

	// ErrCannotParseValue is returned when value parsing fails.
	ErrCannotParseValue = errors.New("cannot parse value")

	// ErrValueOverflows is returned when value overflows target type.
	ErrValueOverflows = errors.New("value overflows target type")

	// ErrEmptyInput is returned when input is empty.
	ErrEmptyInput = errors.New("empty input")

	// ErrNegativeValue is returned when negative value cannot be converted.
	ErrNegativeValue = errors.New("negative value cannot be converted")

	// ErrNotWholeNumber is returned when value is not a whole number.
	ErrNotWholeNumber = errors.New("not a whole number")

	// ErrUnsupportedConversion is returned when conversion is not supported.
	ErrUnsupportedConversion = errors.New("unsupported conversion")

	// ErrNilValue is returned when value is nil.
	ErrNilValue = errors.New("nil value")

	// ErrInvalidFilenamePath is returned when filename path is invalid.
	ErrInvalidFilenamePath = errors.New("invalid filename path")

	// ErrFailedToParseTemplate is returned when template parsing fails.
	ErrFailedToParseTemplate = errors.New("failed to parse template")

	// ErrFailedToFormatCode is returned when code formatting fails.
	ErrFailedToFormatCode = errors.New("failed to format generated code")

	// ErrFailedToCreateDirectory is returned when directory creation fails.
	ErrFailedToCreateDirectory = errors.New("failed to create directory")

	// ErrFailedToResolveAbsolutePath is returned when absolute path resolution fails.
	ErrFailedToResolveAbsolutePath = errors.New("failed to resolve absolute path")

	// ErrFailedToGetCurrentDirectory is returned when getting current directory fails.
	ErrFailedToGetCurrentDirectory = errors.New("failed to get current directory")

	// ErrPathOutsideCurrentDirectory is returned when path is outside current directory.
	ErrPathOutsideCurrentDirectory = errors.New("path outside current directory")

	// ErrFailedToCreateFile is returned when file creation fails.
	ErrFailedToCreateFile = errors.New("failed to create file")

	// ErrFailedToWriteContent is returned when writing content to file fails.
	ErrFailedToWriteContent = errors.New("failed to write content to file")
)
