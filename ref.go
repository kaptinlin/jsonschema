package jsonschema

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/kaptinlin/jsonpointer"
)

// resolveRef resolves a reference to another schema, either locally or globally, supporting both $ref and $dynamicRef.
func (s *Schema) resolveRef(ref string) (*Schema, error) {
	if ref == "#" {
		return s.rootSchema(), nil
	}

	if anchor, ok := strings.CutPrefix(ref, "#"); ok {
		return s.resolveAnchor(anchor)
	}

	// Resolve the full URL if ref is a relative URL
	if !isAbsoluteURI(ref) && s.baseURI != "" {
		ref = resolveRelativeURI(s.baseURI, ref)
	}

	// Handle full URL references
	return s.resolveRefWithFullURL(ref)
}

func (s *Schema) resolveAnchor(anchorName string) (*Schema, error) {
	if strings.HasPrefix(anchorName, "/") {
		schema, err := s.resolveJSONPointer(anchorName)
		if schema == nil && s.parent != nil {
			return s.parent.resolveAnchor(anchorName)
		}
		return schema, err
	}

	if schema, ok := s.anchors[anchorName]; ok {
		return schema, nil
	}
	if schema, ok := s.dynamicAnchors[anchorName]; ok {
		return schema, nil
	}
	if s.parent != nil {
		return s.parent.resolveAnchor(anchorName)
	}
	return nil, nil
}

// resolveRefWithFullURL resolves a full URL reference to another schema.
func (s *Schema) resolveRefWithFullURL(ref string) (*Schema, error) {
	root := s.rootSchema()
	if resolved, err := root.getSchema(ref); err == nil {
		return resolved, nil
	}

	resolved, err := s.Compiler().Schema(ref)
	if err != nil {
		return nil, ErrGlobalReferenceResolution
	}
	return resolved, nil
}

// resolveJSONPointer resolves a JSON Pointer within the schema based on JSON Schema structure.
func (s *Schema) resolveJSONPointer(pointer string) (*Schema, error) {
	if pointer == "/" {
		return s, nil
	}

	segments := jsonpointer.Parse(pointer)
	currentSchema := s

	for i := 0; i < len(segments); i++ {
		segment, err := decodeJSONPointerSegment(segments[i])
		if err != nil {
			return nil, err
		}

		nextSchema, err := currentSchema.schemaForPointerSegment(segment, segments, &i)
		if err != nil {
			return nil, err
		}
		currentSchema = nextSchema
	}

	return currentSchema, nil
}

func decodeJSONPointerSegment(segment string) (string, error) {
	decodedSegment, err := url.PathUnescape(segment)
	if err != nil {
		return "", ErrJSONPointerSegmentDecode
	}
	return decodedSegment, nil
}

func (s *Schema) schemaForPointerSegment(segment string, segments []string, index *int) (*Schema, error) {
	switch segment {
	case "properties":
		if s.Properties == nil {
			return nil, ErrJSONPointerSegmentNotFound
		}
		return schemaMapPointerTarget(map[string]*Schema(*s.Properties), segments, index)
	case "patternProperties":
		if s.PatternProperties == nil {
			return nil, ErrJSONPointerSegmentNotFound
		}
		return schemaMapPointerTarget(map[string]*Schema(*s.PatternProperties), segments, index)
	case "$defs", "definitions":
		return schemaMapPointerTarget(s.Defs, segments, index)
	case "dependentSchemas":
		return schemaMapPointerTarget(s.DependentSchemas, segments, index)
	case "prefixItems":
		return schemaSlicePointerTarget(s.PrefixItems, segments, index)
	case "allOf":
		return schemaSlicePointerTarget(s.AllOf, segments, index)
	case "anyOf":
		return schemaSlicePointerTarget(s.AnyOf, segments, index)
	case "oneOf":
		return schemaSlicePointerTarget(s.OneOf, segments, index)
	case "not":
		return schemaPointerTarget(s.Not)
	case "if":
		return schemaPointerTarget(s.If)
	case "then":
		return schemaPointerTarget(s.Then)
	case "else":
		return schemaPointerTarget(s.Else)
	case "items":
		return schemaPointerTarget(s.Items)
	case "contains":
		return schemaPointerTarget(s.Contains)
	case "additionalProperties":
		return schemaPointerTarget(s.AdditionalProperties)
	case "propertyNames":
		return schemaPointerTarget(s.PropertyNames)
	case "unevaluatedItems":
		return schemaPointerTarget(s.UnevaluatedItems)
	case "unevaluatedProperties":
		return schemaPointerTarget(s.UnevaluatedProperties)
	case "contentSchema":
		return schemaPointerTarget(s.ContentSchema)
	}
	return nil, ErrJSONPointerSegmentNotFound
}

func schemaMapPointerTarget(schemas map[string]*Schema, segments []string, index *int) (*Schema, error) {
	if len(schemas) == 0 || *index+1 >= len(segments) {
		return nil, ErrJSONPointerSegmentNotFound
	}

	*index += 1
	key, err := decodeJSONPointerSegment(segments[*index])
	if err != nil {
		return nil, err
	}
	schema, ok := schemas[key]
	if !ok || schema == nil {
		return nil, ErrJSONPointerSegmentNotFound
	}
	return schema, nil
}

func schemaSlicePointerTarget(schemas []*Schema, segments []string, index *int) (*Schema, error) {
	if *index+1 >= len(segments) {
		return nil, ErrJSONPointerSegmentNotFound
	}

	*index += 1
	segment, err := decodeJSONPointerSegment(segments[*index])
	if err != nil {
		return nil, err
	}
	itemIndex, err := strconv.Atoi(segment)
	if err != nil || itemIndex < 0 || itemIndex >= len(schemas) || schemas[itemIndex] == nil {
		return nil, ErrJSONPointerSegmentNotFound
	}
	return schemas[itemIndex], nil
}

func schemaPointerTarget(schema *Schema) (*Schema, error) {
	if schema == nil {
		return nil, ErrJSONPointerSegmentNotFound
	}
	return schema, nil
}

// ResolveUnresolvedReferences tries to resolve any previously unresolved references.
// This is called after new schemas are added to the compiler.
func (s *Schema) ResolveUnresolvedReferences() {
	// Try to resolve unresolved $ref
	if s.Ref != "" && s.ResolvedRef == nil {
		if resolved, err := s.resolveRef(s.Ref); err == nil {
			s.ResolvedRef = resolved
		}
	}

	// Try to resolve unresolved $dynamicRef
	if s.DynamicRef != "" && s.ResolvedDynamicRef == nil {
		if resolved, err := s.resolveRef(s.DynamicRef); err == nil {
			s.ResolvedDynamicRef = resolved
		}
	}

	s.walkNestedSchemas((*Schema).ResolveUnresolvedReferences)
}

func (s *Schema) resolveReferences() {
	if s.Ref != "" {
		if resolved, err := s.resolveRef(s.Ref); err == nil {
			s.ResolvedRef = resolved
		}
	}

	if s.DynamicRef != "" {
		if resolved, err := s.resolveRef(s.DynamicRef); err == nil {
			s.ResolvedDynamicRef = resolved
		}
	}

	s.walkNestedSchemas((*Schema).resolveReferences)
}

// walkNestedSchemas applies fn recursively to all nested subschemas.
func (s *Schema) walkNestedSchemas(fn func(*Schema)) {
	if s.Defs != nil {
		for _, defSchema := range s.Defs {
			fn(defSchema)
		}
	}

	if s.Properties != nil {
		for _, schema := range *s.Properties {
			if schema != nil {
				fn(schema)
			}
		}
	}

	for _, schemas := range [][]*Schema{s.AllOf, s.AnyOf, s.OneOf} {
		for _, schema := range schemas {
			if schema != nil {
				fn(schema)
			}
		}
	}

	if s.Not != nil {
		fn(s.Not)
	}
	if s.Items != nil {
		fn(s.Items)
	}
	for _, schema := range s.PrefixItems {
		fn(schema)
	}
	if s.AdditionalProperties != nil {
		fn(s.AdditionalProperties)
	}
	if s.Contains != nil {
		fn(s.Contains)
	}
	if s.PatternProperties != nil {
		for _, schema := range *s.PatternProperties {
			fn(schema)
		}
	}
}

// UnresolvedReferenceURIs returns a list of URIs that this schema references but are not yet resolved.
func (s *Schema) UnresolvedReferenceURIs() []string {
	var unresolvedURIs []string

	var collect func(*Schema)
	collect = func(schema *Schema) {
		if schema.Ref != "" && schema.ResolvedRef == nil {
			unresolvedURIs = append(unresolvedURIs, schema.Ref)
		}
		if schema.DynamicRef != "" && schema.ResolvedDynamicRef == nil {
			unresolvedURIs = append(unresolvedURIs, schema.DynamicRef)
		}
		schema.walkNestedSchemas(collect)
	}
	collect(s)

	return unresolvedURIs
}
