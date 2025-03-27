
## Creating a Superset (Union) of JSON Schemas


Listed below are the rules that we followed to merge two json-schema to make a new merged json schema.

### 1. Basic Schema Properties

#### 1.1 Schema Metadata
- **$id, $schema**: Generate a new ID for the merged schema, but keep the schema version of the newer one.
- **title, description**: Could combine with text like "Superset of Schema A and Schema B" or choose one.
- **format**: If the formats conflict, omit from the merged schema to be more permissive.
- **examples**: Include examples from both schemas.
- **default**: Choose the latest one if exists otherwise the old one.
- **deprecated, readOnly, writeOnly**: Use the less restrictive option (false over true).

#### 1.2 Boolean Schemas
- If any schema is a boolean schema:
  - If either is `true`, result is `true` (for superset)
  - If both are `false`, result is `false`

### 2. Type Constraints

#### 2.1 Basic Type Handling
- **type**: Union the allowed types. For example, if schema1 allows ["string", "number"] and schema2 allows ["string", "boolean"], the merged schema should allow ["string", "number", "boolean"].

#### 2.2 Enum and Const
- **enum**: Union the enum values (include all values from both enums).
- **const**: If schemas specify different constant values, use `enum` with both values in the merged schema.

### 3. Logical Composition

We can ommit the logical check from the new schema.

### 4. Property Validation (when doing deep merging)

#### 4.1 Object Properties
- **properties**: Include all properties from both schemas.
  - For properties that exist in both schemas, recursively create supersets of those property schemas.
- **patternProperties**: Combine all patterns, recursively creating supersets for overlapping patterns.
- **required**: For a superset, only include properties required by BOTH schemas (intersection of required arrays).
- **additionalProperties**: Use the more permissive constraint:
  - If either schema has `additionalProperties: true`, use `true`.
  - If both have schemas, create a superset of those schemas.
- **propertyNames**: Create a superset of the property name validation schemas.
- **minProperties, maxProperties**: Use the less restrictive values (lower min, higher max).

#### 4.2 String Validation
- **minLength, maxLength**: Use the less restrictive values (lower min, higher max).
- **pattern**: This is tricky for a superset - might need to use `anyOf` with both patterns or create a new regex that is the union.

#### 4.3 Numeric Validation
- **minimum, exclusiveMinimum**: Use the lower value of the two schemas.
- **maximum, exclusiveMaximum**: Use the higher value of the two schemas.
- **multipleOf**: This requires finding a common divisor or using `anyOf` to represent the union, if not found then we can remove the `multipleOf` property from the new schema.

#### 4.4 Array Validation
- **items**: Create a superset of item schemas.
- **prefixItems**: Requires special handling based on lengths and content.
- **contains**: Create a superset of contains schemas.
- **minItems, maxItems**: Use the less restrictive values (lower min, higher max).
- **uniqueItems**: Only true if both schemas require it.
- **minContains, maxContains**: Use the less restrictive values.


