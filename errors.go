package jsonschema

import "errors"

// ErrNoLoaderRegistered is returned when no loader is registered for the specified scheme.
var ErrNoLoaderRegistered = errors.New("no loader registered for scheme")

// ErrFailedToReadData is returned when data cannot be read from the specified URL.
var ErrFailedToReadData = errors.New("failed to read data from URL")

// ErrJSONUnmarshalError is returned when there is an error unmarshalling JSON.
var ErrJSONUnmarshalError = errors.New("json unmarshal error")

// ErrXMLUnmarshalError is returned when there is an error unmarshalling XML.
var ErrXMLUnmarshalError = errors.New("xml unmarshal error")

// ErrYAMLUnmarshalError is returned when there is an error unmarshalling YAML.
var ErrYAMLUnmarshalError = errors.New("yaml unmarshal error")

// ErrFailedToFetch is returned when there is an error fetching from the URL.
var ErrFailedToFetch = errors.New("failed to fetch from URL")

// ErrInvalidHTTPStatusCode is returned when an invalid HTTP status code is returned.
var ErrInvalidHTTPStatusCode = errors.New("invalid HTTP status code returned")

// ErrIPv6AddressNotEnclosed is returned when an IPv6 address is not enclosed in brackets.
var ErrIPv6AddressNotEnclosed = errors.New("ipv6 address is not enclosed in brackets")

// ErrInvalidIPv6Address is returned when the IPv6 address is invalid.
var ErrInvalidIPv6Address = errors.New("invalid ipv6 address")

// ErrUnsupportedTypeForRat is returned when the type is unsupported for conversion to *big.Rat.
var ErrUnsupportedTypeForRat = errors.New("unsupported type for conversion to *big.Rat")

// ErrFailedToConvertToRat is returned when a string fails to convert to *big.Rat.
var ErrFailedToConvertToRat = errors.New("failed to convert string to *big.Rat")

// ErrFailedToResolveGlobalReference is returned when a global reference cannot be resolved.
var ErrFailedToResolveGlobalReference = errors.New("failed to resolve global reference")

// ErrFailedToDecodeSegmentWithJSONPointer is returned when a segment cannot be decoded.
var ErrFailedToDecodeSegmentWithJSONPointer = errors.New("failed to decode segment")

// ErrSegmentNotFoundForJSONPointer is returned when a segment is not found in the schema context.
var ErrSegmentNotFoundForJSONPointer = errors.New("segment not found in the schema context")

// ErrFailedToResolveReference is returned when a reference cannot be resolved.
var ErrFailedToResolveReference = errors.New("failed to resolve reference")

// ErrFailedToResolveDefinitions is returned when definitions in $defs cannot be resolved.
var ErrFailedToResolveDefinitions = errors.New("failed to resolve definitions in $defs")

// ErrFailedToResolveItems is returned when items in an array schema cannot be resolved.
var ErrFailedToResolveItems = errors.New("failed to resolve items")

// ErrInvalidJSONSchemaType is returned when the JSON schema type is invalid.
var ErrInvalidJSONSchemaType = errors.New("invalid JSON schema type")
