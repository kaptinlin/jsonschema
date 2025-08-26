//go:generate jschemagen
package testdata

import "time"

// User demonstrates basic validation rules and common field types
type User struct {
	ID       string    `json:"id" jsonschema:"required,format=uuid"`
	Name     string    `json:"name" jsonschema:"required,minLength=2,maxLength=50"`
	Email    string    `json:"email" jsonschema:"required,format=email"`
	Age      int       `json:"age" jsonschema:"required,minimum=18,maximum=120"`
	Status   string    `json:"status" jsonschema:"enum=active inactive suspended"`
	IsAdmin  bool      `json:"is_admin" jsonschema:"default=false"`
	Bio      *string   `json:"bio" jsonschema:"maxLength=500"`
	CreateAt time.Time `json:"created_at" jsonschema:"required"`
}

// Address demonstrates struct references and pattern validation
type Address struct {
	Street  string `json:"street" jsonschema:"required,minLength=1,maxLength=200"`
	City    string `json:"city" jsonschema:"required,minLength=2,maxLength=100"`
	State   string `json:"state" jsonschema:"required,pattern=^[A-Z]{2}$"`
	ZipCode string `json:"zip_code" jsonschema:"required,pattern=^[0-9]{5}(-[0-9]{4})?$"`
	Country string `json:"country" jsonschema:"required,enum=US CA UK FR DE JP"`
}

// Product demonstrates arrays, complex validation, and enum types
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

// Category demonstrates circular references and $refs/$defs generation
type Category struct {
	ID          string      `json:"id" jsonschema:"required,format=uuid"`
	Name        string      `json:"name" jsonschema:"required,minLength=1,maxLength=100"`
	Description *string     `json:"description" jsonschema:"maxLength=500"`
	Parent      *Category   `json:"parent" jsonschema:""`
	Children    []*Category `json:"children" jsonschema:"items=Category,minItems=0,maxItems=100"`
	Products    []*Product  `json:"products" jsonschema:"items=Product,minItems=0"`
	Level       int         `json:"level" jsonschema:"minimum=0,maximum=10"`
	IsActive    bool        `json:"is_active" jsonschema:"default=true"`
}

// ReferenceDemo demonstrates manual reference usage for advanced JSON Schema features
type ReferenceDemo struct {
	// Manual $ref usage - reference to another schema definition
	UserRef interface{} `json:"user_ref" jsonschema:"ref=#/$defs/User"`

	// Manual $anchor usage - creates an anchor point for referencing
	MainField string `json:"main_field" jsonschema:"anchor=main,required"`

	// Manual $defs usage - includes schema definitions (typically at root level)
	Definitions interface{} `json:"definitions" jsonschema:"defs=User,Address"`

	// Manual $dynamicRef usage - dynamic reference for recursive schemas
	DynamicRef interface{} `json:"dynamic_ref" jsonschema:"dynamicRef=#meta"`

	// Mixed usage - field with both reference and additional validation
	ValidatedRef interface{} `json:"validated_ref" jsonschema:"ref=#/$defs/Product,description=Product reference with validation"`
}
