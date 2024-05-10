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
			return nil, ErrFailedToDecodeSegment
		}

		nextSchema, found := findSchemaInSegment(currentSchema, decodedSegment, previousSegment)
		if found {
			currentSchema = nextSchema
			previousSegment = decodedSegment // Update the context for the next iteration
			continue
		}

		if !found && i == len(segments)-1 {
			// If no schema is found and it's the last segment, throw error
			return nil, ErrSegmentNotFound
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

func (s *Schema) resolveReferences() error {
	// Resolve the root reference if this schema itself is a reference
	if s.Ref != "" {
		resolved, err := s.resolveRef(s.Ref) // Resolve against root schema

		if err != nil {
			return ErrFailedToResolveDefinitions
		}
		s.ResolvedRef = resolved
	}

	if s.DynamicRef != "" {
		resolved, err := s.resolveRef(s.DynamicRef) // Resolve dynamic references against root schema

		if err != nil {
			return ErrFailedToResolveReference
		}
		s.ResolvedDynamicRef = resolved
	}

	// Recursively resolve references within definitions
	if s.Defs != nil {
		for _, defSchema := range s.Defs {
			if err := defSchema.resolveReferences(); err != nil {
				return ErrFailedToResolveDefinitions
			}
		}
	}

	// Recursively resolve references in properties
	if s.Properties != nil {
		for _, schema := range *s.Properties {
			if schema != nil {
				if err := schema.resolveReferences(); err != nil {
					return ErrFailedToResolveReference
				}
			}
		}
	}

	// Additional fields that can have subschemas
	resolveSubschemaList(s.AllOf)
	resolveSubschemaList(s.AnyOf)
	resolveSubschemaList(s.OneOf)
	if s.Not != nil {
		if err := s.Not.resolveReferences(); err != nil {
			return err
		}
	}
	if s.Items != nil {
		if err := s.Items.resolveReferences(); err != nil {
			return ErrFailedToResolveItems
		}
	}
	if s.PrefixItems != nil {
		for _, schema := range s.PrefixItems {
			if err := schema.resolveReferences(); err != nil {
				return err
			}
		}
	}

	if s.AdditionalProperties != nil {
		if err := s.AdditionalProperties.resolveReferences(); err != nil {
			return err
		}
	}
	if s.Contains != nil {
		if err := s.Contains.resolveReferences(); err != nil {
			return err
		}
	}
	if s.PatternProperties != nil {
		for _, schema := range *s.PatternProperties {
			if err := schema.resolveReferences(); err != nil {
				return err
			}
		}
	}

	// Resolve any other types of nested schemas
	return nil
}

// Helper function to resolve references in a list of schemas
func resolveSubschemaList(schemas []*Schema) error {
	for _, schema := range schemas {
		if schema != nil {
			if err := schema.resolveReferences(); err != nil {
				return err
			}
		}
	}
	return nil
}
