package jsonschema

import "errors"

// === Network and IO Related Errors ===
var (
	// ErrNoLoaderRegistered is returned when no loader is registered for the specified scheme.
	ErrNoLoaderRegistered = errors.New("no loader registered for scheme")

	// ErrDataRead is returned when data cannot be read from the specified URL.
	ErrDataRead = errors.New("data read failed") // Former: ErrFailedToReadData

	// ErrNetworkFetch is returned when there is an error fetching from the URL.
	ErrNetworkFetch = errors.New("network fetch failed") // Former: ErrFailedToFetch

	// ErrInvalidStatusCode is returned when an invalid HTTP status code is returned.
	ErrInvalidStatusCode = errors.New("invalid http status code") // Former: ErrInvalidHTTPStatusCode

	// ErrFileWrite is returned when file writing fails.
	ErrFileWrite = errors.New("file write failed") // Former: ErrFailedToWriteFile

	// ErrFileCreation is returned when file creation fails.
	ErrFileCreation = errors.New("file creation failed") // Former: ErrFailedToCreateFile

	// ErrDirectoryCreation is returned when directory creation fails.
	ErrDirectoryCreation = errors.New("directory creation failed") // Former: ErrFailedToCreateDirectory

	// ErrContentWrite is returned when writing content to file fails.
	ErrContentWrite = errors.New("content write failed") // Former: ErrFailedToWriteContent

	// ErrInvalidFilenamePath is returned when filename path is invalid.
	ErrInvalidFilenamePath = errors.New("invalid filename path")
)

// === Serialization Related Errors ===
var (
	// ErrJSONUnmarshal is returned when there is an error unmarshalling JSON.
	ErrJSONUnmarshal = errors.New("json unmarshal failed") // Former: ErrJSONUnmarshalError

	// ErrXMLUnmarshal is returned when there is an error unmarshalling XML.
	ErrXMLUnmarshal = errors.New("xml unmarshal failed") // Former: ErrXMLUnmarshalError

	// ErrYAMLUnmarshal is returned when there is an error unmarshalling YAML.
	ErrYAMLUnmarshal = errors.New("yaml unmarshal failed") // Former: ErrYAMLUnmarshalError

	// ErrJSONDecode is returned when JSON decoding fails.
	ErrJSONDecode = errors.New("json decode failed") // Former: ErrFailedToDecodeJSON

	// ErrSourceEncode is returned when source encoding fails.
	ErrSourceEncode = errors.New("source encode failed") // Former: ErrFailedToEncodeSource

	// ErrIntermediateJSONDecode is returned when intermediate JSON decoding fails.
	ErrIntermediateJSONDecode = errors.New("intermediate json decode failed") // Former: ErrFailedToDecodeIntermediateJSON

	// ErrDataEncode is returned when data encoding fails.
	ErrDataEncode = errors.New("data encode failed") // Former: ErrFailedToEncodeData

	// ErrNestedValueEncode is returned when encoding nested values fails.
	ErrNestedValueEncode = errors.New("nested value encode failed") // Former: ErrFailedToEncodeNestedValue
)

// === Schema Compilation and Parsing Related Errors ===
var (
	// ErrSchemaCompilation is returned when a schema compilation fails.
	ErrSchemaCompilation = errors.New("schema compilation failed") // Former: ErrFailedToCompileSchema

	// ErrReferenceResolution is returned when a reference cannot be resolved.
	ErrReferenceResolution = errors.New("reference resolution failed") // Former: ErrFailedToResolveReference

	// ErrGlobalReferenceResolution is returned when a global reference cannot be resolved.
	ErrGlobalReferenceResolution = errors.New("global reference resolution failed") // Former: ErrFailedToResolveGlobalReference

	// ErrDefinitionResolution is returned when definitions in $defs cannot be resolved.
	ErrDefinitionResolution = errors.New("definition resolution failed") // Former: ErrFailedToResolveDefinitions

	// ErrItemResolution is returned when items in an array schema cannot be resolved.
	ErrItemResolution = errors.New("item resolution failed") // Former: ErrFailedToResolveItems

	// ErrJSONPointerSegmentDecode is returned when a segment cannot be decoded.
	ErrJSONPointerSegmentDecode = errors.New("json pointer segment decode failed") // Former: ErrFailedToDecodeSegmentWithJSONPointer

	// ErrJSONPointerSegmentNotFound is returned when a segment is not found in the schema context.
	ErrJSONPointerSegmentNotFound = errors.New("json pointer segment not found") // Former: ErrSegmentNotFoundForJSONPointer

	// ErrInvalidSchemaType is returned when the JSON schema type is invalid.
	ErrInvalidSchemaType = errors.New("invalid schema type") // Former: ErrInvalidJSONSchemaType

	// ErrSchemaIsNil is returned when schema is nil.
	ErrSchemaIsNil = errors.New("schema is nil")

	// ErrSchemaInternalsIsNil is returned when schema internals is nil.
	ErrSchemaInternalsIsNil = errors.New("schema internals is nil")
)

// === Data Validation Related Errors ===
var (
	// ErrValueValidationFailed is returned when value validation fails.
	ErrValueValidationFailed = errors.New("value validation failed")

	// ErrInvalidRuleFormat is returned when rule format is invalid.
	ErrInvalidRuleFormat = errors.New("invalid rule format")

	// ErrRuleRequiresParameter is returned when a rule requires a parameter.
	ErrRuleRequiresParameter = errors.New("rule requires parameter")

	// ErrEmptyRuleName is returned when rule name is empty.
	ErrEmptyRuleName = errors.New("empty rule name")

	// ErrValidatorAlreadyExists is returned when a validator with the same name already exists.
	ErrValidatorAlreadyExists = errors.New("validator already exists")
)

// === Type Conversion Related Errors ===
var (
	// Basic type conversion
	// ErrTypeConversion is returned when type conversion fails.
	ErrTypeConversion = errors.New("type conversion failed")

	// ErrTimeConversion is returned when time conversion fails.
	ErrTimeConversion = errors.New("time conversion failed") // Former: ErrTimeTypeConversion

	// ErrTimeParsing is returned when time parsing fails.
	ErrTimeParsing = errors.New("time parsing failed") // Former: ErrTimeParseFailure

	// ErrRatConversion is returned when rat conversion fails.
	ErrRatConversion = errors.New("rat conversion failed") // Former: ErrFailedToConvertToRat

	// Collection type conversion
	// ErrSliceConversion is returned when slice conversion fails.
	ErrSliceConversion = errors.New("slice conversion failed") // Former: ErrCannotConvertSlice

	// ErrSliceElementConversion is returned when slice element conversion fails.
	ErrSliceElementConversion = errors.New("slice element conversion failed") // Former: ErrCannotConvertSliceElement

	// ErrFirstSliceConversion is returned when first slice conversion fails.
	ErrFirstSliceConversion = errors.New("first slice conversion failed") // Former: ErrCannotConvertFirstSlice

	// ErrSecondSliceConversion is returned when second slice conversion fails.
	ErrSecondSliceConversion = errors.New("second slice conversion failed") // Former: ErrCannotConvertSecondSlice

	// ErrNilConversion is returned when nil conversion fails.
	ErrNilConversion = errors.New("nil conversion failed") // Former: ErrCannotConvertNil

	// ErrNilPointerConversion is returned when nil pointer conversion fails.
	ErrNilPointerConversion = errors.New("nil pointer conversion failed") // Former: ErrCannotConvertNilPointer

	// ErrValueParsing is returned when value parsing fails.
	ErrValueParsing = errors.New("value parsing failed") // Former: ErrCannotParseValue

	// ErrValueAssignment is returned when value assignment fails.
	ErrValueAssignment = errors.New("value assignment failed") // Former: ErrCannotAssignValue

	// Specific type support
	// ErrUnsupportedType is returned when an unsupported type is encountered.
	ErrUnsupportedType = errors.New("unsupported type")

	// ErrUnsupportedRatType is returned when the type is unsupported for conversion to *big.Rat.
	ErrUnsupportedRatType = errors.New("unsupported rat type") // Former: ErrUnsupportedTypeForRat

	// ErrUnsupportedInputType is returned when an unsupported input type is encountered.
	ErrUnsupportedInputType = errors.New("unsupported input type")

	// ErrUnsupportedGenerationType is returned when an unsupported type is encountered during code generation.
	ErrUnsupportedGenerationType = errors.New("unsupported generation type") // Former: ErrUnsupportedTypeForGeneration

	// ErrUnsupportedConversion is returned when conversion is not supported.
	ErrUnsupportedConversion = errors.New("unsupported conversion")

	// ErrUnrepresentableType is returned when a type cannot be represented.
	ErrUnrepresentableType = errors.New("unrepresentable type")

	// ErrInvalidTransformType is returned when type is invalid for transform.
	ErrInvalidTransformType = errors.New("invalid transform type") // Former: ErrInvalidTypeForTransform
)

// === Struct and Reflection Related Errors ===
var (
	// ErrExpectedStructType is returned when a non-struct type is provided where a struct type is expected.
	ErrExpectedStructType = errors.New("expected struct type")

	// ErrStructTagParsing is returned when struct tags cannot be parsed.
	ErrStructTagParsing = errors.New("struct tag parsing failed") // Former: ErrFailedToParseStructTags

	// ErrFieldNotFound is returned when field is not found.
	ErrFieldNotFound = errors.New("field not found")

	// ErrFieldAssignment is returned when field assignment fails.
	ErrFieldAssignment = errors.New("field assignment failed") // Former: ErrFailedToSetField

	// ErrFieldAnalysis is returned when field analysis fails.
	ErrFieldAnalysis = errors.New("field analysis failed") // Former: ErrFailedToAnalyzeFields

	// ErrNilDestination is returned when destination cannot be nil.
	ErrNilDestination = errors.New("destination cannot be nil")

	// ErrNotPointer is returned when destination must be a pointer.
	ErrNotPointer = errors.New("destination must be a pointer")

	// ErrNilPointer is returned when destination pointer cannot be nil.
	ErrNilPointer = errors.New("destination pointer cannot be nil")
)

// === Default Value Handling Related Errors ===
var (
	// ErrDefaultApplication is returned when applying defaults fails.
	ErrDefaultApplication = errors.New("default application failed") // Former: ErrFailedToApplyDefaults

	// ErrDefaultEvaluation is returned when evaluating default values fails.
	ErrDefaultEvaluation = errors.New("default evaluation failed") // Former: ErrFailedToEvaluateDefaultValue

	// ErrArrayDefaultApplication is returned when applying array defaults fails.
	ErrArrayDefaultApplication = errors.New("array default application failed") // Former: ErrFailedToApplyArrayDefaults

	// ErrFunctionCallParsing is returned when parsing function calls fails.
	ErrFunctionCallParsing = errors.New("function call parsing failed") // Former: ErrFailedToParseFunctionCall
)

// === Code Generation Related Errors ===
var (
	// ErrNilConfig is returned when a configuration is nil.
	ErrNilConfig = errors.New("config cannot be nil") // Former: ErrConfigCannotBeNil

	// ErrAnalyzerCreation is returned when analyzer creation fails.
	ErrAnalyzerCreation = errors.New("analyzer creation failed") // Former: ErrFailedToCreateAnalyzer

	// ErrWriterCreation is returned when writer creation fails.
	ErrWriterCreation = errors.New("writer creation failed") // Former: ErrFailedToCreateWriter

	// ErrPackageAnalysis is returned when package analysis fails.
	ErrPackageAnalysis = errors.New("package analysis failed") // Former: ErrFailedToAnalyzePackage

	// ErrCodeGeneration is returned when code generation fails.
	ErrCodeGeneration = errors.New("code generation failed") // Former: ErrFailedToGenerateCode

	// ErrPropertyGeneration is returned when property generation fails.
	ErrPropertyGeneration = errors.New("property generation failed") // Former: ErrFailedToGenerateProperty

	// ErrJSONSchemaTagParsing is returned when jsonschema tag parsing fails.
	ErrJSONSchemaTagParsing = errors.New("json schema tag parsing failed") // Former: ErrFailedToParseJSONSchemaTag

	// ErrGozodTagParsing is returned when gozod tag parsing fails.
	ErrGozodTagParsing = errors.New("gozod tag parsing failed") // Former: ErrFailedToParseGozodTag

	// ErrTemplateLoading is returned when template loading fails.
	ErrTemplateLoading = errors.New("template loading failed") // Former: ErrFailedToLoadTemplates

	// ErrOutputDirectoryCreation is returned when output directory creation fails.
	ErrOutputDirectoryCreation = errors.New("output directory creation failed") // Former: ErrFailedToCreateOutputDirectory

	// ErrFieldSchemaGeneration is returned when field schema generation fails.
	ErrFieldSchemaGeneration = errors.New("field schema generation failed") // Former: ErrFailedToGenerateFieldSchemas

	// ErrTemplateExecution is returned when template execution fails.
	ErrTemplateExecution = errors.New("template execution failed") // Former: ErrFailedToExecuteTemplate

	// ErrMainTemplateParsing is returned when main template parsing fails.
	ErrMainTemplateParsing = errors.New("main template parsing failed") // Former: ErrFailedToParseMainTemplate

	// ErrDependencyAnalysis is returned when dependency analysis fails.
	ErrDependencyAnalysis = errors.New("dependency analysis failed") // Former: ErrFailedToAnalyzeDependencies

	// ErrTemplateParsing is returned when template parsing fails.
	ErrTemplateParsing = errors.New("template parsing failed") // Former: ErrFailedToParseTemplate

	// ErrCodeFormatting is returned when code formatting fails.
	ErrCodeFormatting = errors.New("code formatting failed") // Former: ErrFailedToFormatCode

	// ErrStructNodeNotFound is returned when a struct node cannot be found.
	ErrStructNodeNotFound = errors.New("struct node not found")
)

// === Array and Object Related Errors ===
var (
	// ErrArrayItemNotSchema is returned when an array item is not a schema.
	ErrArrayItemNotSchema = errors.New("array item not schema")

	// ErrExpectedArrayOrSlice is returned when array or slice is expected.
	ErrExpectedArrayOrSlice = errors.New("expected array or slice")

	// ErrNilPointerToArray is returned when trying to convert nil pointer to array.
	ErrNilPointerToArray = errors.New("nil pointer to array")

	// ErrNilPointerToRecord is returned when trying to convert nil pointer to record.
	ErrNilPointerToRecord = errors.New("nil pointer to record")

	// ErrNilPointerToObject is returned when trying to convert nil pointer to object.
	ErrNilPointerToObject = errors.New("nil pointer to object")

	// ErrExpectedMapStringAny is returned when map[string]any is expected.
	ErrExpectedMapStringAny = errors.New("expected map[string]any")

	// ErrNonStringKeyFound is returned when non-string key is found in map.
	ErrNonStringKeyFound = errors.New("non-string key found in map")
)

// === Numeric and Format Related Errors ===
var (
	// ErrValueOverflows is returned when value overflows target type.
	ErrValueOverflows = errors.New("value overflows target type")

	// ErrEmptyInput is returned when input is empty.
	ErrEmptyInput = errors.New("empty input")

	// ErrNegativeValueConversion is returned when negative value cannot be converted.
	ErrNegativeValueConversion = errors.New("negative value conversion failed") // Former: ErrNegativeValue

	// ErrNonWholeNumber is returned when value is not a whole number.
	ErrNonWholeNumber = errors.New("not a whole number") // Former: ErrNotWholeNumber

	// ErrInvalidSliceArrayString is returned when input is not a slice, array, or string.
	ErrInvalidSliceArrayString = errors.New("invalid slice array string") // Former: ErrNotSliceArrayOrString

	// ErrNilValue is returned when value is nil.
	ErrNilValue = errors.New("nil value")

	// ErrNilConstValue is returned when trying to unmarshal into a nil ConstValue.
	ErrNilConstValue = errors.New("cannot unmarshal into nil ConstValue")

	// ErrIPv6AddressFormat is returned when an IPv6 address is not properly formatted.
	ErrIPv6AddressFormat = errors.New("ipv6 address format error") // Former: ErrIPv6AddressNotEnclosed

	// ErrInvalidIPv6 is returned when the IPv6 address is invalid.
	ErrInvalidIPv6 = errors.New("invalid ipv6 address") // Former: ErrInvalidIPv6Address
)

// === Path and Filesystem Related Errors ===
var (
	// ErrAbsolutePathResolution is returned when absolute path resolution fails.
	ErrAbsolutePathResolution = errors.New("absolute path resolution failed") // Former: ErrFailedToResolveAbsolutePath

	// ErrCurrentDirectoryAccess is returned when getting current directory fails.
	ErrCurrentDirectoryAccess = errors.New("current directory access failed") // Former: ErrFailedToGetCurrentDirectory

	// ErrPathOutsideDirectory is returned when path is outside current directory.
	ErrPathOutsideDirectory = errors.New("path outside directory") // Former: ErrPathOutsideCurrentDirectory
)

// === Legacy and Generic Errors (To be gradually deprecated) ===
var (
	// ErrUnderlyingError is returned as a generic underlying error.
	ErrUnderlyingError = errors.New("underlying error") // To be deprecated: insufficient information
)

// === Backward Compatibility Aliases ===
// The following are deprecated aliases defined to maintain backward compatibility,
// to be removed in future versions

var (
	// === Network and IO Related Error Aliases ===
	// Deprecated: Use ErrDataRead instead
	ErrFailedToReadData = ErrDataRead

	// Deprecated: Use ErrNetworkFetch instead
	ErrFailedToFetch = ErrNetworkFetch

	// Deprecated: Use ErrInvalidStatusCode instead
	ErrInvalidHTTPStatusCode = ErrInvalidStatusCode

	// Deprecated: Use ErrFileWrite instead
	ErrFailedToWriteFile = ErrFileWrite

	// Deprecated: Use ErrFileCreation instead
	ErrFailedToCreateFile = ErrFileCreation

	// Deprecated: Use ErrDirectoryCreation instead
	ErrFailedToCreateDirectory = ErrDirectoryCreation

	// Deprecated: Use ErrContentWrite instead
	ErrFailedToWriteContent = ErrContentWrite

	// === Serialization Related Error Aliases ===
	// Deprecated: Use ErrJSONUnmarshal instead
	ErrJSONUnmarshalError = ErrJSONUnmarshal

	// Deprecated: Use ErrXMLUnmarshal instead
	ErrXMLUnmarshalError = ErrXMLUnmarshal

	// Deprecated: Use ErrYAMLUnmarshal instead
	ErrYAMLUnmarshalError = ErrYAMLUnmarshal

	// Deprecated: Use ErrJSONDecode instead
	ErrFailedToDecodeJSON = ErrJSONDecode

	// Deprecated: Use ErrSourceEncode instead
	ErrFailedToEncodeSource = ErrSourceEncode

	// Deprecated: Use ErrIntermediateJSONDecode instead
	ErrFailedToDecodeIntermediateJSON = ErrIntermediateJSONDecode

	// Deprecated: Use ErrDataEncode instead
	ErrFailedToEncodeData = ErrDataEncode

	// Deprecated: Use ErrNestedValueEncode instead
	ErrFailedToEncodeNestedValue = ErrNestedValueEncode

	// === Schema Compilation and Parsing Related Error Aliases ===
	// Deprecated: Use ErrSchemaCompilation instead
	ErrFailedToCompileSchema = ErrSchemaCompilation

	// Deprecated: Use ErrReferenceResolution instead
	ErrFailedToResolveReference = ErrReferenceResolution

	// Deprecated: Use ErrGlobalReferenceResolution instead
	ErrFailedToResolveGlobalReference = ErrGlobalReferenceResolution

	// Deprecated: Use ErrDefinitionResolution instead
	ErrFailedToResolveDefinitions = ErrDefinitionResolution

	// Deprecated: Use ErrItemResolution instead
	ErrFailedToResolveItems = ErrItemResolution

	// Deprecated: Use ErrJSONPointerSegmentDecode instead
	ErrFailedToDecodeSegmentWithJSONPointer = ErrJSONPointerSegmentDecode

	// Deprecated: Use ErrJSONPointerSegmentNotFound instead
	ErrSegmentNotFoundForJSONPointer = ErrJSONPointerSegmentNotFound

	// Deprecated: Use ErrInvalidSchemaType instead
	ErrInvalidJSONSchemaType = ErrInvalidSchemaType

	// === Type Conversion Related Error Aliases ===
	// Deprecated: Use ErrTimeConversion instead
	ErrTimeTypeConversion = ErrTimeConversion

	// Deprecated: Use ErrTimeParsing instead
	ErrTimeParseFailure = ErrTimeParsing

	// Deprecated: Use ErrRatConversion instead
	ErrFailedToConvertToRat = ErrRatConversion

	// Deprecated: Use ErrSliceConversion instead
	ErrCannotConvertSlice = ErrSliceConversion

	// Deprecated: Use ErrSliceElementConversion instead
	ErrCannotConvertSliceElement = ErrSliceElementConversion

	// Deprecated: Use ErrFirstSliceConversion instead
	ErrCannotConvertFirstSlice = ErrFirstSliceConversion

	// Deprecated: Use ErrSecondSliceConversion instead
	ErrCannotConvertSecondSlice = ErrSecondSliceConversion

	// Deprecated: Use ErrNilConversion instead
	ErrCannotConvertNil = ErrNilConversion

	// Deprecated: Use ErrNilPointerConversion instead
	ErrCannotConvertNilPointer = ErrNilPointerConversion

	// Deprecated: Use ErrValueParsing instead
	ErrCannotParseValue = ErrValueParsing

	// Deprecated: Use ErrValueAssignment instead
	ErrCannotAssignValue = ErrValueAssignment

	// Deprecated: Use ErrUnsupportedRatType instead
	ErrUnsupportedTypeForRat = ErrUnsupportedRatType

	// Deprecated: Use ErrUnsupportedGenerationType instead
	ErrUnsupportedTypeForGeneration = ErrUnsupportedGenerationType

	// Deprecated: Use ErrInvalidTransformType instead
	ErrInvalidTypeForTransform = ErrInvalidTransformType

	// === Struct and Reflection Related Error Aliases ===
	// Deprecated: Use ErrStructTagParsing instead
	ErrFailedToParseStructTags = ErrStructTagParsing

	// Deprecated: Use ErrFieldAssignment instead
	ErrFailedToSetField = ErrFieldAssignment

	// Deprecated: Use ErrFieldAnalysis instead
	ErrFailedToAnalyzeFields = ErrFieldAnalysis

	// === Default Value Handling Related Error Aliases ===
	// Deprecated: Use ErrDefaultApplication instead
	ErrFailedToApplyDefaults = ErrDefaultApplication

	// Deprecated: Use ErrDefaultEvaluation instead
	ErrFailedToEvaluateDefaultValue = ErrDefaultEvaluation

	// Deprecated: Use ErrArrayDefaultApplication instead
	ErrFailedToApplyArrayDefaults = ErrArrayDefaultApplication

	// Deprecated: Use ErrFunctionCallParsing instead
	ErrFailedToParseFunctionCall = ErrFunctionCallParsing

	// === Code Generation Related Error Aliases ===
	// Deprecated: Use ErrNilConfig instead
	ErrConfigCannotBeNil = ErrNilConfig

	// Deprecated: Use ErrAnalyzerCreation instead
	ErrFailedToCreateAnalyzer = ErrAnalyzerCreation

	// Deprecated: Use ErrWriterCreation instead
	ErrFailedToCreateWriter = ErrWriterCreation

	// Deprecated: Use ErrPackageAnalysis instead
	ErrFailedToAnalyzePackage = ErrPackageAnalysis

	// Deprecated: Use ErrCodeGeneration instead
	ErrFailedToGenerateCode = ErrCodeGeneration

	// Deprecated: Use ErrPropertyGeneration instead
	ErrFailedToGenerateProperty = ErrPropertyGeneration

	// Deprecated: Use ErrJSONSchemaTagParsing instead
	ErrFailedToParseJSONSchemaTag = ErrJSONSchemaTagParsing

	// Deprecated: Use ErrGozodTagParsing instead
	ErrFailedToParseGozodTag = ErrGozodTagParsing

	// Deprecated: Use ErrTemplateLoading instead
	ErrFailedToLoadTemplates = ErrTemplateLoading

	// Deprecated: Use ErrOutputDirectoryCreation instead
	ErrFailedToCreateOutputDirectory = ErrOutputDirectoryCreation

	// Deprecated: Use ErrFieldSchemaGeneration instead
	ErrFailedToGenerateFieldSchemas = ErrFieldSchemaGeneration

	// Deprecated: Use ErrTemplateExecution instead
	ErrFailedToExecuteTemplate = ErrTemplateExecution

	// Deprecated: Use ErrMainTemplateParsing instead
	ErrFailedToParseMainTemplate = ErrMainTemplateParsing

	// Deprecated: Use ErrDependencyAnalysis instead
	ErrFailedToAnalyzeDependencies = ErrDependencyAnalysis

	// Deprecated: Use ErrTemplateParsing instead
	ErrFailedToParseTemplate = ErrTemplateParsing

	// Deprecated: Use ErrCodeFormatting instead
	ErrFailedToFormatCode = ErrCodeFormatting

	// === Numeric and Format Related Error Aliases ===
	// Deprecated: Use ErrNegativeValueConversion instead
	ErrNegativeValue = ErrNegativeValueConversion

	// Deprecated: Use ErrNonWholeNumber instead
	ErrNotWholeNumber = ErrNonWholeNumber

	// Deprecated: Use ErrInvalidSliceArrayString instead
	ErrNotSliceArrayOrString = ErrInvalidSliceArrayString

	// Deprecated: Use ErrIPv6AddressFormat instead
	ErrIPv6AddressNotEnclosed = ErrIPv6AddressFormat

	// Deprecated: Use ErrInvalidIPv6 instead
	ErrInvalidIPv6Address = ErrInvalidIPv6

	// === Path and Filesystem Related Error Aliases ===
	// Deprecated: Use ErrAbsolutePathResolution instead
	ErrFailedToResolveAbsolutePath = ErrAbsolutePathResolution

	// Deprecated: Use ErrCurrentDirectoryAccess instead
	ErrFailedToGetCurrentDirectory = ErrCurrentDirectoryAccess

	// Deprecated: Use ErrPathOutsideDirectory instead
	ErrPathOutsideCurrentDirectory = ErrPathOutsideDirectory

	// === Aliases from unmarshal.go ===
	// Note: ErrTimeParseFailure and ErrTimeTypeConversion aliases are already defined above
)
