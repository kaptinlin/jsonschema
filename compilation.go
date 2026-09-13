package jsonschema

import (
	"errors"
	"fmt"
	"strings"
)

// compilation owns every new node until its complete reference graph is ready.
// Published resources are only read; loaders join this same private graph.
type compilation struct {
	compiler  *Compiler
	resources map[string]*Schema
	roots     []*Schema
	loaded    []string
}

// compileGraph shares publication and retry rules across all compiler entry
// points. A retry discards only private nodes and binds against the winning
// published resources; already-published graphs are never patched in place.
func compileGraph[T any](c *Compiler, build func(*compilation) (T, error)) (T, error) {
	for {
		b := newCompilation(c)
		result, err := build(b)
		if err == nil {
			err = b.finish()
		}
		if b.shouldRetry(err) {
			continue
		}
		if err != nil {
			var zero T
			return zero, err
		}
		return result, nil
	}
}

func (b *compilation) shouldRetry(err error) bool {
	if !errors.Is(err, ErrSchemaConflict) {
		return false
	}
	b.compiler.mu.RLock()
	defer b.compiler.mu.RUnlock()
	for _, uri := range b.loaded {
		if b.compiler.schemas[uri] != nil {
			// This retrieval was absent when loading began. Each retry therefore
			// observes publication progress, rather than repeating a fixed conflict.
			return true
		}
	}
	return false
}

func newCompilation(c *Compiler) *compilation {
	return &compilation{compiler: c, resources: make(map[string]*Schema)}
}

func (b *compilation) parse(data []byte, uri string) (*Schema, error) {
	schema, err := parseSchema(data)
	if err != nil {
		return nil, err
	}
	if uri != "" {
		uri = resolveRelativeURI(b.compiler.DefaultBaseURI, uri)
	}
	schema.retrievalURI = uri
	schema.compiler = b.compiler
	base := b.compiler.DefaultBaseURI
	if uri != "" {
		base = uri
	}
	if schema.ID != "" {
		if err := b.register(resolveRelativeURI(base, schema.ID), schema); err != nil {
			return nil, err
		}
	}
	if err := b.register(uri, schema); err != nil {
		return nil, err
	}
	b.roots = append(b.roots, schema)
	return schema, nil
}

func (b *compilation) prepare(schema *Schema) error {
	if err := schema.applyDialectsWithResources(b.compiler, b.resources); err != nil {
		return err
	}
	if schema.ID == "" {
		schema.ID = schema.retrievalURI
	}
	schema.initializeSchemaWithoutReferences(b.compiler, nil)
	if err := b.registerTree(schema); err != nil {
		return err
	}
	return schema.validateRegexSyntax()
}

func (b *compilation) lookup(ref string) (*Schema, error) {
	uri, anchor := splitRef(ref)
	schema := b.resources[uri]
	if schema == nil {
		b.compiler.mu.RLock()
		schema = b.compiler.schemas[uri]
		b.compiler.mu.RUnlock()
	}
	if schema == nil {
		b.loaded = append(b.loaded, uri)
		data, err := b.compiler.loadSchema(uri)
		if err != nil {
			return nil, err
		}
		schema, err = b.parse(data, uri)
		if err != nil {
			return nil, err
		}
		if err := b.prepare(schema); err != nil {
			return nil, err
		}
	}
	if anchor == "" {
		return schema, nil
	}
	return schema.resolveAnchor(anchor)
}

func (b *compilation) finish() error {
	// Resolving one root may load and append more roots. All are checked before
	// any of them become visible through the compiler's registry.
	for i := 0; i < len(b.roots); i++ {
		if err := b.roots[i].bindReferences(b.lookup, "#"); err != nil {
			return err
		}
	}
	b.compiler.mu.Lock()
	defer b.compiler.mu.Unlock()
	for uri := range b.resources {
		if _, exists := b.compiler.schemas[uri]; exists {
			return fmt.Errorf("%w: %s", ErrSchemaConflict, uri)
		}
	}
	for _, schema := range b.roots {
		schema.markCompiled()
	}
	for uri, schema := range b.resources {
		b.compiler.schemas[uri] = schema
	}
	return nil
}

func (s *Schema) markCompiled() {
	s.compiled = true
	s.forEachChild((*Schema).markCompiled)
}

func (b *compilation) register(uri string, schema *Schema) error {
	uri = strings.TrimSuffix(uri, "#")
	if uri == "" {
		return nil
	}
	if existing := b.resources[uri]; existing != nil && existing != schema {
		return fmt.Errorf("%w: %s", ErrSchemaConflict, uri)
	}
	b.compiler.mu.RLock()
	_, exists := b.compiler.schemas[uri]
	b.compiler.mu.RUnlock()
	if exists {
		return fmt.Errorf("%w: %s", ErrSchemaConflict, uri)
	}
	b.resources[uri] = schema
	return nil
}

func (b *compilation) registerTree(schema *Schema) error {
	err := b.register(schema.uri, schema)
	schema.forEachChild(func(child *Schema) {
		if err == nil {
			err = b.registerTree(child)
		}
	})
	return err
}

func (b *compilation) prepareAll() error {
	for _, schema := range b.roots {
		if err := b.prepare(schema); err != nil {
			return fmt.Errorf("compiling schema %s: %w", schema.ID, err)
		}
	}
	return nil
}
