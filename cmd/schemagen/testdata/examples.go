//go:generate schemagen
package testdata

import "time"

// Basic validation examples (KISS principle - simple, focused structs)

// User demonstrates basic field validation rules
type User struct {
	ID        string    `json:"id" jsonschema:"required,format=uuid"`
	Name      string    `json:"name" jsonschema:"required,minLength=2,maxLength=50"`
	Email     string    `json:"email" jsonschema:"required,format=email"`
	Age       int       `json:"age" jsonschema:"required,minimum=18,maximum=120"`
	Status    string    `json:"status" jsonschema:"enum=active inactive suspended"`
	IsAdmin   bool      `json:"is_admin" jsonschema:"default=false"`
	Bio       *string   `json:"bio" jsonschema:"maxLength=500"`
	CreatedAt time.Time `json:"created_at" jsonschema:"required"`
}

// Address demonstrates pattern validation
type Address struct {
	Street  string `json:"street" jsonschema:"required,minLength=1,maxLength=200"`
	City    string `json:"city" jsonschema:"required,minLength=2,maxLength=100"`
	State   string `json:"state" jsonschema:"required,pattern=^[A-Z]{2}$"`
	ZipCode string `json:"zip_code" jsonschema:"required,pattern=^[0-9]{5}(-[0-9]{4})?$"`
	Country string `json:"country" jsonschema:"required,enum=US CA UK FR DE JP"`
}

// Product demonstrates array and numeric validation
type Product struct {
	SKU         string   `json:"sku" jsonschema:"required,pattern=^[A-Z0-9-]+$"`
	Name        string   `json:"name" jsonschema:"required,minLength=1,maxLength=200"`
	Description *string  `json:"description" jsonschema:"maxLength=1000"`
	Price       float64  `json:"price" jsonschema:"required,minimum=0.01,maximum=999999.99"`
	Currency    string   `json:"currency" jsonschema:"required,enum=USD EUR GBP JPY"`
	Tags        []string `json:"tags" jsonschema:"items=string,minItems=0,maxItems=10,uniqueItems"`
	Active      bool     `json:"active" jsonschema:"default=true"`
	CategoryID  *string  `json:"category_id" jsonschema:"format=uuid"`
}

// Advanced examples

// Category demonstrates circular references and nested structures
type Category struct {
	ID          string      `json:"id" jsonschema:"required,format=uuid"`
	Name        string      `json:"name" jsonschema:"required,minLength=1,maxLength=100"`
	Description *string     `json:"description" jsonschema:"maxLength=500"`
	Parent      *Category   `json:"parent"`
	Children    []*Category `json:"children" jsonschema:"items=Category,minItems=0,maxItems=100"`
	Products    []*Product  `json:"products" jsonschema:"items=Product,minItems=0"`
	Level       int         `json:"level" jsonschema:"minimum=0,maximum=10"`
	IsActive    bool        `json:"is_active" jsonschema:"default=true"`
}

// RefDemo demonstrates advanced JSON Schema features: $ref, $anchor, $defs, $dynamicRef
type RefDemo struct {
	// $ref - reference to another schema definition
	UserRef any `json:"user_ref" jsonschema:"ref=#/$defs/User"`

	// $anchor - creates an anchor point for referencing
	MainField string `json:"main_field" jsonschema:"anchor=main,required"`

	// $defs - includes schema definitions
	Definitions any `json:"definitions" jsonschema:"defs=User,Address"`

	// $dynamicRef - dynamic reference for recursive schemas
	DynamicRef any `json:"dynamic_ref" jsonschema:"dynamicRef=#meta"`

	// Combined: reference with additional validation
	ValidatedRef any `json:"validated_ref" jsonschema:"ref=#/$defs/Product,description=Product reference with validation"`
}
