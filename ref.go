package jsonschema

import (
	"net/url"
	"strconv"
	"strings"
)

// resolveRef resolves a reference to another schema, either locally or globally, supporting both $ref and $dynamicRef.
func (s *Schema) resolveRef(ref string) (*Schema, error) {
	if ref == "#" {
		return s.getRootSchema(), nil
	}

	if strings.HasPrefix(ref, "#") {
		return s.resolveAnchor(ref[1:])
	}

	// Resolve the full URL if ref is a relative URL
	if !isAbsoluteURI(ref) && s.baseURI != "" {
		ref = resolveRelativeURI(s.baseURI, ref)
	}

	// Handle full URL references
	return s.resolveRefWithFullURL(ref)
}

func (s *Schema) resolveAnchor(anchorName string) (*Schema, error) {
	var schema *Schema
	var err error

	if strings.HasPrefix(anchorName, "/") {
		schema, err = s.resolveJSONPointer(anchorName)
	} else {
		if schema, ok := s.anchors[anchorName]; ok {
			return schema, nil
		}

		if schema, ok := s.dynamicAnchors[anchorName]; ok {
			return schema, nil
		}
	}

	if schema == nil && s.parent != nil {
		return s.parent.resolveAnchor(anchorName)
	}

	return schema, err
}

// resolveRefWithFullURL resolves a full URL reference to another schema.
func (s *Schema) resolveRefWithFullURL(ref string) (*Schema, error) {
	root := s.getRootSchema()
	if resolved, err := root.getSchema(ref); err == nil {
		return resolved, nil
	}

	// If not found in the current schema or its parents, look for the reference in the compiler
	if resolved, err := s.compiler.GetSchema(ref); err != nil {
		return nil, ErrFailedToResolveGlobalReference
	} else {
		return resolved, nil
	}
}

// resolveJSONPointer resolves a JSON Pointer within the schema based on JSON Schema structure.
func (s *Schema) resolveJSONPointer(pointer string) (*Schema, error) {
	if pointer == "/" {
		return s, nil
	}

	segments := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	currentSchema := s
	previousSegment := ""

	for i, segment := range segments {
		decodedSegment, err := url.PathUnescape(strings.ReplaceAll(strings.ReplaceAll(segment, "~1", "/"), "~0", "~"))
		if err != nil {
			return nil, ErrFailedToDecodeSegmentWithJSONPointer
		}

		nextSchema, found := findSchemaInSegment(currentSchema, decodedSegment, previousSegment)
		if found {
			currentSchema = nextSchema
			previousSegment = decodedSegment // Update the context for the next iteration
			continue
		}

		if !found && i == len(segments)-1 {
			// If no schema is found and it's the last segment, throw error
			return nil, ErrSegmentNotFoundForJSONPointer
		}

		previousSegment = decodedSegment // Update the context for the next iteration
	}

	return currentSchema, nil
}

// Helper function to find a schema within a given segment
func findSchemaInSegment(currentSchema *Schema, segment string, previousSegment string) (*Schema, bool) {
	switch previousSegment {
	case "properties":
		if currentSchema.Properties != nil {
			if schema, exists := (*currentSchema.Properties)[segment]; exists {
				return schema, true
			}
		}
	case "prefixItems":
		index, err := strconv.Atoi(segment)

		if err == nil && currentSchema.PrefixItems != nil && index < len(currentSchema.PrefixItems) {
			return currentSchema.PrefixItems[index], true
		}
	case "$defs":
		if defSchema, exists := currentSchema.Defs[segment]; exists {
			return defSchema, true
		}
	case "items":
		if currentSchema.Items != nil {
			return currentSchema.Items, true
		}
	}
	return nil, false
}

func (s *Schema) resolveReferences() {
	// Resolve the root reference if this schema itself is a reference
	if s.Ref != "" {
		resolved, _ := s.resolveRef(s.Ref) // Resolve against root schema
		s.ResolvedRef = resolved
	}

	if s.DynamicRef != "" {
		resolved, _ := s.resolveRef(s.DynamicRef) // Resolve dynamic references against root schema
		s.ResolvedDynamicRef = resolved
	}

	// Recursively resolve references within definitions
	if s.Defs != nil {
		for _, defSchema := range s.Defs {
			defSchema.resolveReferences()
		}
	}

	// Recursively resolve references in properties
	if s.Properties != nil {
		for _, schema := range *s.Properties {
			if schema != nil {
				schema.resolveReferences()
			}
		}
	}

	// Additional fields that can have subschemas
	resolveSubschemaList(s.AllOf)
	resolveSubschemaList(s.AnyOf)
	resolveSubschemaList(s.OneOf)
	if s.Not != nil {
		s.Not.resolveReferences()
	}
	if s.Items != nil {
		s.Items.resolveReferences()
	}
	if s.PrefixItems != nil {
		for _, schema := range s.PrefixItems {
			schema.resolveReferences()
		}
	}

	if s.AdditionalProperties != nil {
		s.AdditionalProperties.resolveReferences()
	}
	if s.Contains != nil {
		s.Contains.resolveReferences()
	}
	if s.PatternProperties != nil {
		for _, schema := range *s.PatternProperties {
			schema.resolveReferences()
		}
	}
}

// Helper function to resolve references in a list of schemas
func resolveSubschemaList(schemas []*Schema) {
	for _, schema := range schemas {
		if schema != nil {
			schema.resolveReferences()
		}
	}
}
