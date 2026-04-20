package jsonschema

// defaultCompiler is the compiler used by the constructor API.
var defaultCompiler = NewCompiler()

// SetDefaultCompiler sets the compiler used by the constructor API.
func SetDefaultCompiler(c *Compiler) {
	defaultCompiler = c
}

// DefaultCompiler returns the compiler used by the constructor API.
func DefaultCompiler() *Compiler {
	return defaultCompiler
}

// Property pairs a property name with its schema.
type Property struct {
	Name   string
	Schema *Schema
}

// Prop returns a named property for Object.
func Prop(name string, schema *Schema) Property {
	return Property{Name: name, Schema: schema}
}

// Object returns an object schema built from properties and keywords.
func Object(items ...any) *Schema {
	schema := &Schema{Type: SchemaType{"object"}}

	var properties []Property
	var keywords []Keyword

	for _, item := range items {
		switch v := item.(type) {
		case Property:
			properties = append(properties, v)
		case Keyword:
			keywords = append(keywords, v)
		}
	}

	if len(properties) > 0 {
		props := make(SchemaMap)
		for _, prop := range properties {
			props[prop.Name] = prop.Schema
		}
		schema.Properties = &props
	}

	for _, keyword := range keywords {
		keyword(schema)
	}

	schema.initializeSchema(nil, nil)
	return schema
}

func newTypedSchema(typeName string, keywords []Keyword) *Schema {
	schema := &Schema{}
	if typeName != "" {
		schema.Type = SchemaType{typeName}
	}
	for _, keyword := range keywords {
		keyword(schema)
	}
	schema.initializeSchema(nil, nil)
	return schema
}

// String returns a string schema with the given keywords.
func String(keywords ...Keyword) *Schema { return newTypedSchema("string", keywords) }

// Integer returns an integer schema with the given keywords.
func Integer(keywords ...Keyword) *Schema { return newTypedSchema("integer", keywords) }

// Number returns a number schema with the given keywords.
func Number(keywords ...Keyword) *Schema { return newTypedSchema("number", keywords) }

// Boolean returns a boolean schema with the given keywords.
func Boolean(keywords ...Keyword) *Schema { return newTypedSchema("boolean", keywords) }

// Null returns a null schema with the given keywords.
func Null(keywords ...Keyword) *Schema { return newTypedSchema("null", keywords) }

// Array returns an array schema with the given keywords.
func Array(keywords ...Keyword) *Schema { return newTypedSchema("array", keywords) }

// Any returns a schema with no type restriction.
func Any(keywords ...Keyword) *Schema { return newTypedSchema("", keywords) }

// Const returns a schema with `const` set to value.
func Const(value any) *Schema {
	schema := &Schema{
		Const: &ConstValue{Value: value, IsSet: true},
	}
	schema.initializeSchema(nil, nil)
	return schema
}

// Enum returns a schema with `enum` set to values.
func Enum(values ...any) *Schema {
	schema := &Schema{Enum: values}
	schema.initializeSchema(nil, nil)
	return schema
}

// OneOf returns a schema with `oneOf` set to schemas.
func OneOf(schemas ...*Schema) *Schema {
	schema := &Schema{OneOf: schemas}
	schema.initializeSchema(nil, nil)
	return schema
}

// AnyOf returns a schema with `anyOf` set to schemas.
func AnyOf(schemas ...*Schema) *Schema {
	schema := &Schema{AnyOf: schemas}
	schema.initializeSchema(nil, nil)
	return schema
}

// AllOf returns a schema with `allOf` set to schemas.
func AllOf(schemas ...*Schema) *Schema {
	schema := &Schema{AllOf: schemas}
	schema.initializeSchema(nil, nil)
	return schema
}

// Not returns a schema with `not` set to schema.
func Not(schema *Schema) *Schema {
	result := &Schema{Not: schema}
	result.initializeSchema(nil, nil)
	return result
}

// If begins an `if`/`then`/`else` schema.
func If(condition *Schema) *ConditionalSchema {
	return &ConditionalSchema{condition: condition}
}

// ConditionalSchema builds `if`/`then`/`else` schemas.
type ConditionalSchema struct {
	condition *Schema
	then      *Schema
	otherwise *Schema
}

// Then sets the `then` branch.
func (cs *ConditionalSchema) Then(then *Schema) *ConditionalSchema {
	cs.then = then
	return cs
}

// Else sets the `else` branch and returns the completed schema.
func (cs *ConditionalSchema) Else(otherwise *Schema) *Schema {
	cs.otherwise = otherwise
	return cs.ToSchema()
}

// ToSchema returns the conditional schema as a regular schema.
func (cs *ConditionalSchema) ToSchema() *Schema {
	schema := &Schema{
		If:   cs.condition,
		Then: cs.then,
		Else: cs.otherwise,
	}
	schema.initializeSchema(nil, nil)
	return schema
}

// Ref returns a schema with `$ref` set to ref.
func Ref(ref string) *Schema {
	schema := &Schema{Ref: ref}
	schema.initializeSchema(nil, nil)
	return schema
}
