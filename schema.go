package jsonschema

import (
	"regexp"

	"github.com/goccy/go-json"
)

// Schema represents a JSON Schema as per the 2020-12 draft, containing all
// necessary metadata and validation properties defined by the specification.
type Schema struct {
	compiledPatterns      map[string]*regexp.Regexp // Cached compiled regular expressions for pattern properties.
	compiler              *Compiler                 // Reference to the associated Compiler instance.
	parent                *Schema                   // Parent schema for hierarchical resolution.
	uri                   string                    // Internal schema identifier resolved during compilation.
	baseURI               string                    // Base URI for resolving relative references within the schema.
	anchors               map[string]*Schema        // Anchors for quick lookup of internal schema references.
	dynamicAnchors        map[string]*Schema        // Dynamic anchors for more flexible schema references.
	schemas               map[string]*Schema        // Cache of compiled schemas.
	compiledStringPattern *regexp.Regexp            // Cached compiled regular expressions for string patterns.

	ID     string  `json:"$id,omitempty"`     // Public identifier for the schema.
	Schema string  `json:"$schema,omitempty"` // URI indicating the specification the schema conforms to.
	Format *string `json:"format,omitempty"`  // Format hint for string data, e.g., "email" or "date-time".

	// Schema reference keywords, see https://json-schema.org/draft/2020-12/json-schema-core#ref
	Ref                string             `json:"$ref,omitempty"`           // Reference to another schema.
	DynamicRef         string             `json:"$dynamicRef,omitempty"`    // Reference to another schema that can be dynamically resolved.
	Anchor             string             `json:"$anchor,omitempty"`        // Anchor for resolving relative JSON Pointers.
	DynamicAnchor      string             `json:"$dynamicAnchor,omitempty"` // Anchor for dynamic resolution
	Defs               map[string]*Schema `json:"$defs,omitempty"`          // An object containing schema definitions.
	ResolvedRef        *Schema            `json:"-"`                        // Resolved schema for $ref
	ResolvedDynamicRef *Schema            `json:"-"`                        // Resolved schema for $dynamicRef

	// Boolean JSON Schemas, see https://json-schema.org/draft/2020-12/json-schema-core#name-boolean-json-schemas
	Boolean *bool `json:"-"` // Boolean schema, used for quick validation.

	// Applying subschemas with logical keywords, see https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subsch
	AllOf []*Schema `json:"allOf,omitempty"` // Array of schemas for validating the instance against all of them.
	AnyOf []*Schema `json:"anyOf,omitempty"` // Array of schemas for validating the instance against any of them.
	OneOf []*Schema `json:"oneOf,omitempty"` // Array of schemas for validating the instance against exactly one of them.
	Not   *Schema   `json:"not,omitempty"`   // Schema for validating the instance against the negation of it.

	// Applying subschemas conditionally, see https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subsche
	If               *Schema            `json:"if,omitempty"`               // Schema to be evaluated as a condition
	Then             *Schema            `json:"then,omitempty"`             // Schema to be evaluated if 'if' is successful
	Else             *Schema            `json:"else,omitempty"`             // Schema to be evaluated if 'if' is not successful
	DependentSchemas map[string]*Schema `json:"dependentSchemas,omitempty"` // Dependent schemas based on property presence

	// Applying subschemas to array keywords, see https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subschem
	PrefixItems []*Schema `json:"prefixItems,omitempty"` // Array of schemas for validating the array items' prefix.
	Items       *Schema   `json:"items,omitempty"`       // Schema for items in an array.
	Contains    *Schema   `json:"contains,omitempty"`    // Schema for validating items in the array.

	// Applying subschemas to objects keywords, see https://json-schema.org/draft/2020-12/json-schema-core#name-keywords-for-applying-subschemas
	Properties           *SchemaMap `json:"properties,omitempty"`           // Definitions of properties for object types.
	PatternProperties    *SchemaMap `json:"patternProperties,omitempty"`    // Definitions of properties for object types matched by specific patterns.
	AdditionalProperties *Schema    `json:"additionalProperties,omitempty"` // Can be a boolean or a schema, controls additional properties handling.
	PropertyNames        *Schema    `json:"propertyNames,omitempty"`        // Can be a boolean or a schema, controls property names validation.

	// Any validation keywords, see https://json-schema.org/draft/2020-12/json-schema-validation#section-6.1
	Type  SchemaType    `json:"type,omitempty"`  // Can be a single type or an array of types.
	Enum  []interface{} `json:"enum,omitempty"`  // Enumerated values for the property.
	Const *ConstValue   `json:"const,omitempty"` // Constant value for the property.

	// Numeric validation keywords, see https://json-schema.org/draft/2020-12/json-schema-validation#section-6.2
	MultipleOf       *Rat `json:"multipleOf,omitempty"`       // Number must be a multiple of this value, strictly greater than 0.
	Maximum          *Rat `json:"maximum,omitempty"`          // Maximum value of the number.
	ExclusiveMaximum *Rat `json:"exclusiveMaximum,omitempty"` // Number must be less than this value.
	Minimum          *Rat `json:"minimum,omitempty"`          // Minimum value of the number.
	ExclusiveMinimum *Rat `json:"exclusiveMinimum,omitempty"` // Number must be greater than this value.

	// String validation keywords, see https://json-schema.org/draft/2020-12/json-schema-validation#section-6.3
	MaxLength *float64 `json:"maxLength,omitempty"` // Maximum length of a string.
	MinLength *float64 `json:"minLength,omitempty"` // Minimum length of a string.
	Pattern   *string  `json:"pattern,omitempty"`   // Regular expression pattern to match the string against.

	// Array validation keywords, see https://json-schema.org/draft/2020-12/json-schema-validation#section-6.4
	MaxItems    *float64 `json:"maxItems,omitempty"`    // Maximum number of items in an array.
	MinItems    *float64 `json:"minItems,omitempty"`    // Minimum number of items in an array.
	UniqueItems *bool    `json:"uniqueItems,omitempty"` // Whether the items in the array must be unique.
	MaxContains *float64 `json:"maxContains,omitempty"` // Maximum number of items in the array that can match the contains schema.
	MinContains *float64 `json:"minContains,omitempty"` // Minimum number of items in the array that must match the contains schema.

	// https://json-schema.org/draft/2020-12/json-schema-core#name-unevaluateditems
	UnevaluatedItems *Schema `json:"unevaluatedItems,omitempty"` // Schema for unevaluated items in an array.

	// Object validation keywords, see https://json-schema.org/draft/2020-12/json-schema-validation#section-6.5
	MaxProperties     *float64            `json:"maxProperties,omitempty"`     // Maximum number of properties in an object.
	MinProperties     *float64            `json:"minProperties,omitempty"`     // Minimum number of properties in an object.
	Required          []string            `json:"required,omitempty"`          // List of required property names for object types.
	DependentRequired map[string][]string `json:"dependentRequired,omitempty"` // Properties required when another property is present.

	// https://json-schema.org/draft/2020-12/json-schema-core#name-unevaluatedproperties
	UnevaluatedProperties *Schema `json:"unevaluatedProperties,omitempty"` // Schema for unevaluated properties in an object.

	// Content validation keywords, see https://json-schema.org/draft/2020-12/json-schema-validation#name-a-vocabulary-for-the-conten
	ContentEncoding  *string `json:"contentEncoding,omitempty"`  // Encoding format of the content.
	ContentMediaType *string `json:"contentMediaType,omitempty"` // Media type of the content.
	ContentSchema    *Schema `json:"contentSchema,omitempty"`    // Schema for validating the content.

	// Meta-data for schema and instance description, see https://json-schema.org/draft/2020-12/json-schema-validation#name-a-vocabulary-for-basic-meta
	Title       *string       `json:"title,omitempty"`       // A short summary of the schema.
	Description *string       `json:"description,omitempty"` // A detailed description of the purpose of the schema.
	Default     interface{}   `json:"default,omitempty"`     // Default value of the instance.
	Deprecated  *bool         `json:"deprecated,omitempty"`  // Indicates that the schema is deprecated.
	ReadOnly    *bool         `json:"readOnly,omitempty"`    // Indicates that the property is read-only.
	WriteOnly   *bool         `json:"writeOnly,omitempty"`   // Indicates that the property is write-only.
	Examples    []interface{} `json:"examples,omitempty"`    // Examples of the instance data that validates against this schema.
}

// newSchema parses JSON schema data and returns a Schema object.
func newSchema(jsonSchema []byte) (*Schema, error) {
	schema := &Schema{}

	// Parse schema
	if err := json.Unmarshal(jsonSchema, schema); err != nil {
		return nil, err
	}

	return schema, nil
}

// initializeSchema sets up the schema structure, resolves URIs, and initializes nested schemas.
// It populates schema properties from the compiler settings and the parent schema context.
func (s *Schema) initializeSchema(compiler *Compiler, parent *Schema) {
	s.compiler = compiler
	s.parent = parent

	parentBaseURI := s.getParentBaseURI()
	if parentBaseURI == "" {
		parentBaseURI = compiler.DefaultBaseURI
	}
	if s.ID != "" {
		if isValidURI(s.ID) {
			s.uri = s.ID
			s.baseURI = getBaseURI(s.ID)
		} else {
			resolvedURL := resolveRelativeURI(parentBaseURI, s.ID)
			s.uri = resolvedURL
			s.baseURI = getBaseURI(resolvedURL)
		}
	} else {
		s.baseURI = parentBaseURI
	}

	if s.baseURI == "" {
		if s.uri != "" && isValidURI(s.uri) {
			s.baseURI = getBaseURI(s.uri)
		}
	}

	if s.Anchor != "" {
		s.setAnchor(s.Anchor)
	}

	if s.DynamicAnchor != "" {
		s.setDynamicAnchor(s.DynamicAnchor)
	}

	if s.uri != "" && isValidURI(s.uri) {
		root := s.getRootSchema()
		root.setSchema(s.uri, s)
	}

	initializeNestedSchemas(s, compiler)
	s.resolveReferences()
}

// initializeNestedSchemas initializes all nested or related schemas as defined in the structure.
func initializeNestedSchemas(s *Schema, compiler *Compiler) {
	if s.Defs != nil {
		for _, def := range s.Defs {
			def.initializeSchema(compiler, s)
		}
	}
	// Initialize logical schema groupings
	initializeSchemas(s.AllOf, compiler, s)
	initializeSchemas(s.AnyOf, compiler, s)
	initializeSchemas(s.OneOf, compiler, s)

	// Initialize conditional schemas
	if s.Not != nil {
		s.Not.initializeSchema(compiler, s)
	}
	if s.If != nil {
		s.If.initializeSchema(compiler, s)
	}
	if s.Then != nil {
		s.Then.initializeSchema(compiler, s)
	}
	if s.Else != nil {
		s.Else.initializeSchema(compiler, s)
	}
	if s.DependentSchemas != nil {
		for _, depSchema := range s.DependentSchemas {
			depSchema.initializeSchema(compiler, s)
		}
	}

	// Initialize array and object schemas
	if s.PrefixItems != nil {
		for _, item := range s.PrefixItems {
			item.initializeSchema(compiler, s)
		}
	}
	if s.Items != nil {
		s.Items.initializeSchema(compiler, s)
	}
	if s.Contains != nil {
		s.Contains.initializeSchema(compiler, s)
	}
	if s.AdditionalProperties != nil {
		s.AdditionalProperties.initializeSchema(compiler, s)
	}
	if s.Properties != nil {
		for _, prop := range *s.Properties {
			prop.initializeSchema(compiler, s)
		}
	}
	if s.PatternProperties != nil {
		for _, prop := range *s.PatternProperties {
			prop.initializeSchema(compiler, s)
		}
	}
	if s.UnevaluatedProperties != nil {
		s.UnevaluatedProperties.initializeSchema(compiler, s)
	}
	if s.UnevaluatedItems != nil {
		s.UnevaluatedItems.initializeSchema(compiler, s)
	}
	if s.ContentSchema != nil {
		s.ContentSchema.initializeSchema(compiler, s)
	}
	if s.PropertyNames != nil {
		s.PropertyNames.initializeSchema(compiler, s)
	}
}

// setAnchor creates or updates the anchor mapping for the current schema and propagates it to parent schemas.
func (s *Schema) setAnchor(anchor string) {
	if s.anchors == nil {
		s.anchors = make(map[string]*Schema)
	}
	s.anchors[anchor] = s

	root := s.getRootSchema()
	if root.anchors == nil {
		root.anchors = make(map[string]*Schema)
	}

	if _, ok := root.anchors[anchor]; !ok {
		root.anchors[anchor] = s
	}
}

// setDynamicAnchor sets or updates a dynamic anchor for the current schema and propagates it to parents in the same scope.
func (s *Schema) setDynamicAnchor(anchor string) {
	if s.dynamicAnchors == nil {
		s.dynamicAnchors = make(map[string]*Schema)
	}
	if _, ok := s.dynamicAnchors[anchor]; !ok {
		s.dynamicAnchors[anchor] = s
	}

	scope := s.getScopeSchema()
	if scope.dynamicAnchors == nil {
		scope.dynamicAnchors = make(map[string]*Schema)
	}

	if _, ok := scope.dynamicAnchors[anchor]; !ok {
		scope.dynamicAnchors[anchor] = s
	}
}

// setSchema adds a schema to the internal schema cache, using the provided URI as the key.
func (s *Schema) setSchema(uri string, schema *Schema) *Schema {
	if s.schemas == nil {
		s.schemas = make(map[string]*Schema)
	}

	s.schemas[uri] = schema
	return s
}

func (s *Schema) getSchema(ref string) (*Schema, error) {
	baseURI, anchor := splitRef(ref)

	if schema, exists := s.schemas[baseURI]; exists {
		if baseURI == ref {
			return schema, nil
		}
		return schema.resolveAnchor(anchor)
	}

	return nil, ErrFailedToResolveReference
}

// initializeSchemas iteratively initializes a list of nested schemas.
func initializeSchemas(schemas []*Schema, compiler *Compiler, parent *Schema) {
	for _, schema := range schemas {
		if schema != nil {
			schema.initializeSchema(compiler, parent)
		}
	}
}

// GetSchemaURI returns the resolved URI for the schema, or an empty string if no URI is defined.
func (s *Schema) GetSchemaURI() string {
	if s.uri != "" {
		return s.uri
	}
	root := s.getRootSchema()
	if root.uri != "" {
		return root.uri
	}

	return ""
}

func (s *Schema) GetSchemaLocation(anchor string) string {
	uri := s.GetSchemaURI()

	return uri + "#" + anchor
}

// getRootSchema returns the highest-level parent schema, serving as the root in the schema tree.
func (s *Schema) getRootSchema() *Schema {
	if s.parent != nil {
		return s.parent.getRootSchema()
	}

	return s
}

func (s *Schema) getScopeSchema() *Schema {
	if s.ID != "" {
		return s
	} else if s.parent != nil {
		return s.parent.getScopeSchema()
	}

	return s
}

// getParentBaseURI returns the base URI from the nearest parent schema that has one defined,
// or an empty string if none of the parents up to the root define a base URI.
func (s *Schema) getParentBaseURI() string {
	for p := s.parent; p != nil; p = p.parent {
		if p.baseURI != "" {
			return p.baseURI
		}
	}
	return ""
}

// MarshalJSON implements json.Marshaler
func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.Boolean != nil {
		return json.Marshal(s.Boolean)
	}

	// Marshal as a normal struct
	type Alias Schema
	return json.Marshal((*Alias)(s))
}

// UnmarshalJSON handles unmarshaling JSON data into the Schema type.
func (s *Schema) UnmarshalJSON(data []byte) error {
	// First try to parse as a boolean
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		s.Boolean = &b
		return nil
	}

	// If not a boolean, parse as a normal struct
	type Alias Schema
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = Schema(alias)

	// Special handling for the const field
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if constData, ok := raw["const"]; ok {
		if s.Const == nil {
			s.Const = &ConstValue{}
		}
		return s.Const.UnmarshalJSON(constData)
	}

	return nil
}

// SchemaMap represents a map of string keys to *Schema values, used primarily for properties and patternProperties.
type SchemaMap map[string]*Schema

// MarshalJSON ensures that SchemaMap serializes properly as a JSON object.
func (sm SchemaMap) MarshalJSON() ([]byte, error) {
	m := make(map[string]*Schema)
	for k, v := range sm {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON ensures that JSON objects are correctly parsed into SchemaMap,
// supporting the detailed structure required for nested schema definitions.
func (sm *SchemaMap) UnmarshalJSON(data []byte) error {
	m := make(map[string]*Schema)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*sm = SchemaMap(m)
	return nil
}

// SchemaType holds a set of SchemaType values, accommodating complex schema definitions that permit multiple types.
type SchemaType []string

// MarshalJSON customizes the JSON serialization of SchemaType.
func (r SchemaType) MarshalJSON() ([]byte, error) {
	if len(r) == 1 {
		return json.Marshal(r[0])
	}
	return json.Marshal([]string(r))
}

// UnmarshalJSON customizes the JSON deserialization into SchemaType.
func (r *SchemaType) UnmarshalJSON(data []byte) error {
	var singleType string
	if err := json.Unmarshal(data, &singleType); err == nil {
		*r = SchemaType{singleType}
		return nil
	}

	var multiType []string
	if err := json.Unmarshal(data, &multiType); err == nil {
		*r = SchemaType(multiType)
		return nil
	}

	return ErrInvalidJSONSchemaType
}

// ConstValue represents a constant value in a JSON Schema.
type ConstValue struct {
	Value interface{}
	IsSet bool
}

// UnmarshalJSON handles unmarshaling a JSON value into the ConstValue type.
func (cv *ConstValue) UnmarshalJSON(data []byte) error {
	// Ensure cv is not nil
	if cv == nil {
		return ErrNilConstValue
	}

	// Set IsSet to true because we are setting a value
	cv.IsSet = true

	// If the input is "null", explicitly set Value to nil
	if string(data) == "null" {
		cv.Value = nil
		return nil
	}

	// Otherwise parse the value normally
	return json.Unmarshal(data, &cv.Value)
}

// MarshalJSON handles marshaling the ConstValue type back to JSON.
func (cv ConstValue) MarshalJSON() ([]byte, error) {
	if !cv.IsSet {
		return []byte("null"), nil
	}
	if cv.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(cv.Value)
}
