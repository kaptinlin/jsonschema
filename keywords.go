package jsonschema

// Keyword represents a schema keyword that can be applied to any schema
type Keyword func(*Schema)

// ===============================
// String keywords
// ===============================

// MinLen sets the minLength keyword
func MinLen(min int) Keyword {
	return func(s *Schema) {
		f := float64(min)
		s.MinLength = &f
	}
}

// MaxLen sets the maxLength keyword
func MaxLen(max int) Keyword {
	return func(s *Schema) {
		f := float64(max)
		s.MaxLength = &f
	}
}

// Pattern sets the pattern keyword
func Pattern(pattern string) Keyword {
	return func(s *Schema) {
		s.Pattern = &pattern
	}
}

// Format sets the format keyword
func Format(format string) Keyword {
	return func(s *Schema) {
		s.Format = &format
	}
}

// ===============================
// Number keywords
// ===============================

// Min sets the minimum keyword
func Min(min float64) Keyword {
	return func(s *Schema) {
		s.Minimum = NewRat(min)
	}
}

// Max sets the maximum keyword
func Max(max float64) Keyword {
	return func(s *Schema) {
		s.Maximum = NewRat(max)
	}
}

// ExclusiveMin sets the exclusiveMinimum keyword
func ExclusiveMin(min float64) Keyword {
	return func(s *Schema) {
		s.ExclusiveMinimum = NewRat(min)
	}
}

// ExclusiveMax sets the exclusiveMaximum keyword
func ExclusiveMax(max float64) Keyword {
	return func(s *Schema) {
		s.ExclusiveMaximum = NewRat(max)
	}
}

// MultipleOf sets the multipleOf keyword
func MultipleOf(multiple float64) Keyword {
	return func(s *Schema) {
		s.MultipleOf = NewRat(multiple)
	}
}

// ===============================
// Array keywords
// ===============================

// Items sets the items keyword
func Items(itemSchema *Schema) Keyword {
	return func(s *Schema) {
		s.Items = itemSchema
	}
}

// MinItems sets the minItems keyword
func MinItems(min int) Keyword {
	return func(s *Schema) {
		f := float64(min)
		s.MinItems = &f
	}
}

// MaxItems sets the maxItems keyword
func MaxItems(max int) Keyword {
	return func(s *Schema) {
		f := float64(max)
		s.MaxItems = &f
	}
}

// UniqueItems sets the uniqueItems keyword
func UniqueItems(unique bool) Keyword {
	return func(s *Schema) {
		s.UniqueItems = &unique
	}
}

// Contains sets the contains keyword
func Contains(schema *Schema) Keyword {
	return func(s *Schema) {
		s.Contains = schema
	}
}

// MinContains sets the minContains keyword
func MinContains(min int) Keyword {
	return func(s *Schema) {
		f := float64(min)
		s.MinContains = &f
	}
}

// MaxContains sets the maxContains keyword
func MaxContains(max int) Keyword {
	return func(s *Schema) {
		f := float64(max)
		s.MaxContains = &f
	}
}

// PrefixItems sets the prefixItems keyword
func PrefixItems(schemas ...*Schema) Keyword {
	return func(s *Schema) {
		s.PrefixItems = schemas
	}
}

// UnevaluatedItems sets the unevaluatedItems keyword
func UnevaluatedItems(schema *Schema) Keyword {
	return func(s *Schema) {
		s.UnevaluatedItems = schema
	}
}

// ===============================
// Object keywords
// ===============================

// Required sets the required keyword
func Required(fields ...string) Keyword {
	return func(s *Schema) {
		s.Required = fields
	}
}

// AdditionalProps sets the additionalProperties keyword
func AdditionalProps(allowed bool) Keyword {
	return func(s *Schema) {
		s.AdditionalProperties = &Schema{Boolean: &allowed}
	}
}

// AdditionalPropsSchema sets the additionalProperties keyword with a schema
func AdditionalPropsSchema(schema *Schema) Keyword {
	return func(s *Schema) {
		s.AdditionalProperties = schema
	}
}

// MinProps sets the minProperties keyword
func MinProps(min int) Keyword {
	return func(s *Schema) {
		f := float64(min)
		s.MinProperties = &f
	}
}

// MaxProps sets the maxProperties keyword
func MaxProps(max int) Keyword {
	return func(s *Schema) {
		f := float64(max)
		s.MaxProperties = &f
	}
}

// PatternProps sets the patternProperties keyword
func PatternProps(patterns map[string]*Schema) Keyword {
	return func(s *Schema) {
		schemaMap := SchemaMap(patterns)
		s.PatternProperties = &schemaMap
	}
}

// PropertyNames sets the propertyNames keyword
func PropertyNames(schema *Schema) Keyword {
	return func(s *Schema) {
		s.PropertyNames = schema
	}
}

// UnevaluatedProps sets the unevaluatedProperties keyword
func UnevaluatedProps(schema *Schema) Keyword {
	return func(s *Schema) {
		s.UnevaluatedProperties = schema
	}
}

// DependentRequired sets the dependentRequired keyword
func DependentRequired(dependencies map[string][]string) Keyword {
	return func(s *Schema) {
		s.DependentRequired = dependencies
	}
}

// DependentSchemas sets the dependentSchemas keyword
func DependentSchemas(dependencies map[string]*Schema) Keyword {
	return func(s *Schema) {
		s.DependentSchemas = dependencies
	}
}

// ===============================
// Annotation keywords
// ===============================

// Title sets the title keyword
func Title(title string) Keyword {
	return func(s *Schema) {
		s.Title = &title
	}
}

// Description sets the description keyword
func Description(desc string) Keyword {
	return func(s *Schema) {
		s.Description = &desc
	}
}

// Default sets the default keyword
func Default(value interface{}) Keyword {
	return func(s *Schema) {
		s.Default = value
	}
}

// Examples sets the examples keyword
func Examples(examples ...interface{}) Keyword {
	return func(s *Schema) {
		s.Examples = examples
	}
}

// Deprecated sets the deprecated keyword
func Deprecated(deprecated bool) Keyword {
	return func(s *Schema) {
		s.Deprecated = &deprecated
	}
}

// ReadOnly sets the readOnly keyword
func ReadOnly(readOnly bool) Keyword {
	return func(s *Schema) {
		s.ReadOnly = &readOnly
	}
}

// WriteOnly sets the writeOnly keyword
func WriteOnly(writeOnly bool) Keyword {
	return func(s *Schema) {
		s.WriteOnly = &writeOnly
	}
}

// ===============================
// Content keywords
// ===============================

// ContentEncoding sets the contentEncoding keyword
func ContentEncoding(encoding string) Keyword {
	return func(s *Schema) {
		s.ContentEncoding = &encoding
	}
}

// ContentMediaType sets the contentMediaType keyword
func ContentMediaType(mediaType string) Keyword {
	return func(s *Schema) {
		s.ContentMediaType = &mediaType
	}
}

// ContentSchema sets the contentSchema keyword
func ContentSchema(schema *Schema) Keyword {
	return func(s *Schema) {
		s.ContentSchema = schema
	}
}

// ===============================
// Core identifier keywords
// ===============================

// ID sets the $id keyword
func ID(id string) Keyword {
	return func(s *Schema) {
		s.ID = id
	}
}

// SchemaURI sets the $schema keyword
func SchemaURI(schemaURI string) Keyword {
	return func(s *Schema) {
		s.Schema = schemaURI
	}
}

// Anchor sets the $anchor keyword
func Anchor(anchor string) Keyword {
	return func(s *Schema) {
		s.Anchor = anchor
	}
}

// DynamicAnchor sets the $dynamicAnchor keyword
func DynamicAnchor(anchor string) Keyword {
	return func(s *Schema) {
		s.DynamicAnchor = anchor
	}
}

// Defs sets the $defs keyword
func Defs(defs map[string]*Schema) Keyword {
	return func(s *Schema) {
		s.Defs = defs
	}
}

// ===============================
// Format constants
// ===============================

const (
	FormatEmail               = "email"
	FormatDateTime            = "date-time"
	FormatDate                = "date"
	FormatTime                = "time"
	FormatURI                 = "uri"
	FormatURIRef              = "uri-reference"
	FormatUUID                = "uuid"
	FormatHostname            = "hostname"
	FormatIPv4                = "ipv4"
	FormatIPv6                = "ipv6"
	FormatRegex               = "regex"
	FormatIdnEmail            = "idn-email"
	FormatIdnHostname         = "idn-hostname"
	FormatIRI                 = "iri"
	FormatIRIRef              = "iri-reference"
	FormatURITemplate         = "uri-template"
	FormatJSONPointer         = "json-pointer"
	FormatRelativeJSONPointer = "relative-json-pointer"
	FormatDuration            = "duration"
)

// ===============================
// Convenience schema functions
// ===============================

// Email creates an email format string schema
func Email() *Schema {
	return String(Format(FormatEmail))
}

// DateTime creates a date-time format string schema
func DateTime() *Schema {
	return String(Format(FormatDateTime))
}

// Date creates a date format string schema
func Date() *Schema {
	return String(Format(FormatDate))
}

// Time creates a time format string schema
func Time() *Schema {
	return String(Format(FormatTime))
}

// URI creates a URI format string schema
func URI() *Schema {
	return String(Format(FormatURI))
}

// URIRef creates a URI reference format string schema
func URIRef() *Schema {
	return String(Format(FormatURIRef))
}

// UUID creates a UUID format string schema
func UUID() *Schema {
	return String(Format(FormatUUID))
}

// Hostname creates a hostname format string schema
func Hostname() *Schema {
	return String(Format(FormatHostname))
}

// IPv4 creates an IPv4 format string schema
func IPv4() *Schema {
	return String(Format(FormatIPv4))
}

// IPv6 creates an IPv6 format string schema
func IPv6() *Schema {
	return String(Format(FormatIPv6))
}

// IdnEmail creates an internationalized email format string schema
func IdnEmail() *Schema {
	return String(Format(FormatIdnEmail))
}

// IdnHostname creates an internationalized hostname format string schema
func IdnHostname() *Schema {
	return String(Format(FormatIdnHostname))
}

// IRI creates an IRI format string schema
func IRI() *Schema {
	return String(Format(FormatIRI))
}

// IRIRef creates an IRI reference format string schema
func IRIRef() *Schema {
	return String(Format(FormatIRIRef))
}

// URITemplate creates a URI template format string schema
func URITemplate() *Schema {
	return String(Format(FormatURITemplate))
}

// JSONPointer creates a JSON pointer format string schema
func JSONPointer() *Schema {
	return String(Format(FormatJSONPointer))
}

// RelativeJSONPointer creates a relative JSON pointer format string schema
func RelativeJSONPointer() *Schema {
	return String(Format(FormatRelativeJSONPointer))
}

// Duration creates a duration format string schema
func Duration() *Schema {
	return String(Format(FormatDuration))
}

// Regex creates a regex format string schema
func Regex() *Schema {
	return String(Format(FormatRegex))
}

// PositiveInt creates a positive integer schema
func PositiveInt() *Schema {
	return Integer(Min(1))
}

// NonNegativeInt creates a non-negative integer schema
func NonNegativeInt() *Schema {
	return Integer(Min(0))
}

// NegativeInt creates a negative integer schema
func NegativeInt() *Schema {
	return Integer(Max(-1))
}

// NonPositiveInt creates a non-positive integer schema
func NonPositiveInt() *Schema {
	return Integer(Max(0))
}
