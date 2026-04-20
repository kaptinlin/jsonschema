package jsonschema

// Keyword applies a JSON Schema keyword to a schema built with the constructor API.
type Keyword func(*Schema)

// MinLength sets `minLength`.
func MinLength(minLen int) Keyword {
	return func(s *Schema) {
		f := float64(minLen)
		s.MinLength = &f
	}
}

// MaxLength sets `maxLength`.
func MaxLength(maxLen int) Keyword {
	return func(s *Schema) {
		f := float64(maxLen)
		s.MaxLength = &f
	}
}

// Pattern sets `pattern`.
func Pattern(pattern string) Keyword {
	return func(s *Schema) {
		s.Pattern = &pattern
	}
}

// Format sets `format`.
func Format(format string) Keyword {
	return func(s *Schema) {
		s.Format = &format
	}
}

// Min sets `minimum`.
func Min(minVal float64) Keyword {
	return func(s *Schema) {
		s.Minimum = NewRat(minVal)
	}
}

// Max sets `maximum`.
func Max(maxVal float64) Keyword {
	return func(s *Schema) {
		s.Maximum = NewRat(maxVal)
	}
}

// ExclusiveMin sets `exclusiveMinimum`.
func ExclusiveMin(minVal float64) Keyword {
	return func(s *Schema) {
		s.ExclusiveMinimum = NewRat(minVal)
	}
}

// ExclusiveMax sets `exclusiveMaximum`.
func ExclusiveMax(maxVal float64) Keyword {
	return func(s *Schema) {
		s.ExclusiveMaximum = NewRat(maxVal)
	}
}

// MultipleOf sets `multipleOf`.
func MultipleOf(multiple float64) Keyword {
	return func(s *Schema) {
		s.MultipleOf = NewRat(multiple)
	}
}

// Items sets `items`.
func Items(itemSchema *Schema) Keyword {
	return func(s *Schema) {
		s.Items = itemSchema
	}
}

// MinItems sets `minItems`.
func MinItems(minItems int) Keyword {
	return func(s *Schema) {
		f := float64(minItems)
		s.MinItems = &f
	}
}

// MaxItems sets `maxItems`.
func MaxItems(maxItems int) Keyword {
	return func(s *Schema) {
		f := float64(maxItems)
		s.MaxItems = &f
	}
}

// UniqueItems sets `uniqueItems`.
func UniqueItems(unique bool) Keyword {
	return func(s *Schema) {
		s.UniqueItems = &unique
	}
}

// Contains sets `contains`.
func Contains(schema *Schema) Keyword {
	return func(s *Schema) {
		s.Contains = schema
	}
}

// MinContains sets `minContains`.
func MinContains(minContains int) Keyword {
	return func(s *Schema) {
		f := float64(minContains)
		s.MinContains = &f
	}
}

// MaxContains sets `maxContains`.
func MaxContains(maxContains int) Keyword {
	return func(s *Schema) {
		f := float64(maxContains)
		s.MaxContains = &f
	}
}

// PrefixItems sets `prefixItems`.
func PrefixItems(schemas ...*Schema) Keyword {
	return func(s *Schema) {
		s.PrefixItems = schemas
	}
}

// UnevaluatedItems sets `unevaluatedItems`.
func UnevaluatedItems(schema *Schema) Keyword {
	return func(s *Schema) {
		s.UnevaluatedItems = schema
	}
}

// Required sets `required`.
func Required(fields ...string) Keyword {
	return func(s *Schema) {
		s.Required = fields
	}
}

// AdditionalProps sets `additionalProperties` to a boolean schema.
func AdditionalProps(allowed bool) Keyword {
	return func(s *Schema) {
		s.AdditionalProperties = &Schema{Boolean: &allowed}
	}
}

// AdditionalPropsSchema sets `additionalProperties` to a schema.
func AdditionalPropsSchema(schema *Schema) Keyword {
	return func(s *Schema) {
		s.AdditionalProperties = schema
	}
}

// MinProps sets `minProperties`.
func MinProps(minProps int) Keyword {
	return func(s *Schema) {
		f := float64(minProps)
		s.MinProperties = &f
	}
}

// MaxProps sets `maxProperties`.
func MaxProps(maxProps int) Keyword {
	return func(s *Schema) {
		f := float64(maxProps)
		s.MaxProperties = &f
	}
}

// PatternProps sets `patternProperties`.
func PatternProps(patterns map[string]*Schema) Keyword {
	return func(s *Schema) {
		schemaMap := SchemaMap(patterns)
		s.PatternProperties = &schemaMap
	}
}

// PropertyNames sets `propertyNames`.
func PropertyNames(schema *Schema) Keyword {
	return func(s *Schema) {
		s.PropertyNames = schema
	}
}

// UnevaluatedProps sets `unevaluatedProperties`.
func UnevaluatedProps(schema *Schema) Keyword {
	return func(s *Schema) {
		s.UnevaluatedProperties = schema
	}
}

// DependentRequired sets `dependentRequired`.
func DependentRequired(dependencies map[string][]string) Keyword {
	return func(s *Schema) {
		s.DependentRequired = dependencies
	}
}

// DependentSchemas sets `dependentSchemas`.
func DependentSchemas(dependencies map[string]*Schema) Keyword {
	return func(s *Schema) {
		s.DependentSchemas = dependencies
	}
}

// Title sets `title`.
func Title(title string) Keyword {
	return func(s *Schema) {
		s.Title = &title
	}
}

// Description sets `description`.
func Description(desc string) Keyword {
	return func(s *Schema) {
		s.Description = &desc
	}
}

// Default sets `default`.
func Default(value any) Keyword {
	return func(s *Schema) {
		s.Default = value
	}
}

// Examples sets `examples`.
func Examples(examples ...any) Keyword {
	return func(s *Schema) {
		s.Examples = examples
	}
}

// Deprecated sets `deprecated`.
func Deprecated(deprecated bool) Keyword {
	return func(s *Schema) {
		s.Deprecated = &deprecated
	}
}

// ReadOnly sets `readOnly`.
func ReadOnly(readOnly bool) Keyword {
	return func(s *Schema) {
		s.ReadOnly = &readOnly
	}
}

// WriteOnly sets `writeOnly`.
func WriteOnly(writeOnly bool) Keyword {
	return func(s *Schema) {
		s.WriteOnly = &writeOnly
	}
}

// ContentEncoding sets `contentEncoding`.
func ContentEncoding(encoding string) Keyword {
	return func(s *Schema) {
		s.ContentEncoding = &encoding
	}
}

// ContentMediaType sets `contentMediaType`.
func ContentMediaType(mediaType string) Keyword {
	return func(s *Schema) {
		s.ContentMediaType = &mediaType
	}
}

// ContentSchema sets `contentSchema`.
func ContentSchema(schema *Schema) Keyword {
	return func(s *Schema) {
		s.ContentSchema = schema
	}
}

// ID sets `$id`.
func ID(id string) Keyword {
	return func(s *Schema) {
		s.ID = id
	}
}

// SchemaURI sets `$schema`.
func SchemaURI(schemaURI string) Keyword {
	return func(s *Schema) {
		s.Schema = schemaURI
	}
}

// Anchor sets `$anchor`.
func Anchor(anchor string) Keyword {
	return func(s *Schema) {
		s.Anchor = anchor
	}
}

// DynamicAnchor sets `$dynamicAnchor`.
func DynamicAnchor(anchor string) Keyword {
	return func(s *Schema) {
		s.DynamicAnchor = anchor
	}
}

// Defs sets `$defs`.
func Defs(defs map[string]*Schema) Keyword {
	return func(s *Schema) {
		s.Defs = defs
	}
}

// Standard JSON Schema format names from Draft 2020-12.
const (
	// FormatEmail is the `email` format.
	FormatEmail = "email"
	// FormatDateTime is the `date-time` format (RFC 3339).
	FormatDateTime = "date-time"
	// FormatDate is the `date` format (RFC 3339 full-date).
	FormatDate = "date"
	// FormatTime is the `time` format (RFC 3339 full-time).
	FormatTime = "time"
	// FormatURI is the `uri` format (RFC 3986).
	FormatURI = "uri"
	// FormatURIRef is the `uri-reference` format (RFC 3986).
	FormatURIRef = "uri-reference"
	// FormatUUID is the `uuid` format (RFC 4122).
	FormatUUID = "uuid"
	// FormatHostname is the `hostname` format (RFC 1123).
	FormatHostname = "hostname"
	// FormatIPv4 is the `ipv4` format (RFC 2673).
	FormatIPv4 = "ipv4"
	// FormatIPv6 is the `ipv6` format (RFC 4291).
	FormatIPv6 = "ipv6"
	// FormatRegex is the ECMA-262 regular expression format.
	FormatRegex = "regex"
	// FormatIdnEmail is the `idn-email` format (RFC 6531).
	FormatIdnEmail = "idn-email"
	// FormatIdnHostname is the `idn-hostname` format (RFC 5890).
	FormatIdnHostname = "idn-hostname"
	// FormatIRI is the `iri` format (RFC 3987).
	FormatIRI = "iri"
	// FormatIRIRef is the `iri-reference` format (RFC 3987).
	FormatIRIRef = "iri-reference"
	// FormatURITemplate is the `uri-template` format (RFC 6570).
	FormatURITemplate = "uri-template"
	// FormatJSONPointer is the `json-pointer` format (RFC 6901).
	FormatJSONPointer = "json-pointer"
	// FormatRelativeJSONPointer is the `relative-json-pointer` format.
	FormatRelativeJSONPointer = "relative-json-pointer"
	// FormatDuration is the `duration` format (RFC 3339 appendix A / ISO 8601).
	FormatDuration = "duration"
)

// Email returns a string schema with the `email` format.
func Email() *Schema {
	return String(Format(FormatEmail))
}

// DateTime returns a string schema with the `date-time` format.
func DateTime() *Schema {
	return String(Format(FormatDateTime))
}

// Date returns a string schema with the `date` format.
func Date() *Schema {
	return String(Format(FormatDate))
}

// Time returns a string schema with the `time` format.
func Time() *Schema {
	return String(Format(FormatTime))
}

// URI returns a string schema with the `uri` format.
func URI() *Schema {
	return String(Format(FormatURI))
}

// URIRef returns a string schema with the `uri-reference` format.
func URIRef() *Schema {
	return String(Format(FormatURIRef))
}

// UUID returns a string schema with the `uuid` format.
func UUID() *Schema {
	return String(Format(FormatUUID))
}

// Hostname returns a string schema with the `hostname` format.
func Hostname() *Schema {
	return String(Format(FormatHostname))
}

// IPv4 returns a string schema with the `ipv4` format.
func IPv4() *Schema {
	return String(Format(FormatIPv4))
}

// IPv6 returns a string schema with the `ipv6` format.
func IPv6() *Schema {
	return String(Format(FormatIPv6))
}

// IdnEmail returns a string schema with the `idn-email` format.
func IdnEmail() *Schema {
	return String(Format(FormatIdnEmail))
}

// IdnHostname returns a string schema with the `idn-hostname` format.
func IdnHostname() *Schema {
	return String(Format(FormatIdnHostname))
}

// IRI returns a string schema with the `iri` format.
func IRI() *Schema {
	return String(Format(FormatIRI))
}

// IRIRef returns a string schema with the `iri-reference` format.
func IRIRef() *Schema {
	return String(Format(FormatIRIRef))
}

// URITemplate returns a string schema with the `uri-template` format.
func URITemplate() *Schema {
	return String(Format(FormatURITemplate))
}

// JSONPointer returns a string schema with the `json-pointer` format.
func JSONPointer() *Schema {
	return String(Format(FormatJSONPointer))
}

// RelativeJSONPointer returns a string schema with the `relative-json-pointer` format.
func RelativeJSONPointer() *Schema {
	return String(Format(FormatRelativeJSONPointer))
}

// Duration returns a string schema with the `duration` format.
func Duration() *Schema {
	return String(Format(FormatDuration))
}

// Regex returns a string schema with the `regex` format.
func Regex() *Schema {
	return String(Format(FormatRegex))
}

// PositiveInt returns an integer schema with `minimum` set to 1.
func PositiveInt() *Schema {
	return Integer(Min(1))
}

// NonNegativeInt returns an integer schema with `minimum` set to 0.
func NonNegativeInt() *Schema {
	return Integer(Min(0))
}

// NegativeInt returns an integer schema with `maximum` set to -1.
func NegativeInt() *Schema {
	return Integer(Max(-1))
}

// NonPositiveInt returns an integer schema with `maximum` set to 0.
func NonPositiveInt() *Schema {
	return Integer(Max(0))
}
