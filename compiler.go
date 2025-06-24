package jsonschema

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

// FormatDef defines a custom format validation rule
type FormatDef struct {
	// Type specifies which JSON Schema type this format applies to (optional)
	// Supported values: "string", "number", "integer", "boolean", "array", "object"
	// Empty string means applies to all types
	Type string

	// Validate is the validation function
	Validate func(interface{}) bool
}

// Compiler represents a JSON Schema compiler that manages schema compilation and caching.
type Compiler struct {
	mu             sync.RWMutex                                       // Protects concurrent access to schemas map
	schemas        map[string]*Schema                                 // Cache of compiled schemas.
	allSchemas     []*Schema                                          // All compiled schemas, including those without IDs
	unresolvedRefs map[string][]*Schema                               // Track schemas that have unresolved references by URI
	Decoders       map[string]func(string) ([]byte, error)            // Decoders for various encoding formats.
	MediaTypes     map[string]func([]byte) (interface{}, error)       // Media type handlers for unmarshalling data.
	Loaders        map[string]func(url string) (io.ReadCloser, error) // Functions to load schemas from URLs.
	DefaultBaseURI string                                             // Base URI used to resolve relative references.
	AssertFormat   bool                                               // Flag to enforce format validation.

	// JSON encoder/decoder configuration
	jsonEncoder func(v interface{}) ([]byte, error)
	jsonDecoder func(data []byte, v interface{}) error

	// Default function registry
	defaultFuncs map[string]DefaultFunc // Registry for dynamic default value functions

	// Custom format registry
	customFormats   map[string]*FormatDef // Registry for custom format definitions
	customFormatsRW sync.RWMutex          // Protects concurrent access to custom formats
}

// DefaultFunc represents a function that can generate dynamic default values
type DefaultFunc func(args ...any) (any, error)

// NewCompiler creates a new Compiler instance and initializes it with default settings.
func NewCompiler() *Compiler {
	compiler := &Compiler{
		schemas:        make(map[string]*Schema),
		allSchemas:     make([]*Schema, 0),
		unresolvedRefs: make(map[string][]*Schema),
		Decoders:       make(map[string]func(string) ([]byte, error)),
		MediaTypes:     make(map[string]func([]byte) (interface{}, error)),
		Loaders:        make(map[string]func(url string) (io.ReadCloser, error)),
		DefaultBaseURI: "",
		AssertFormat:   false,
		defaultFuncs:   make(map[string]DefaultFunc),
		customFormats:  make(map[string]*FormatDef),

		// Default to standard library JSON implementation
		jsonEncoder: json.Marshal,
		jsonDecoder: json.Unmarshal,
	}
	compiler.initDefaults()
	return compiler
}

// WithEncoderJSON configures custom JSON encoder implementation
func (c *Compiler) WithEncoderJSON(encoder func(v interface{}) ([]byte, error)) *Compiler {
	c.jsonEncoder = encoder
	return c
}

// WithDecoderJSON configures custom JSON decoder implementation
func (c *Compiler) WithDecoderJSON(decoder func(data []byte, v interface{}) error) *Compiler {
	c.jsonDecoder = decoder
	return c
}

// Compile compiles a JSON schema and caches it. If an URI is provided, it uses that as the key; otherwise, it generates a hash.
func (c *Compiler) Compile(jsonSchema []byte, uris ...string) (*Schema, error) {
	schema, err := newSchema(jsonSchema)
	if err != nil {
		return nil, err
	}

	uri := schema.ID
	if uri == "" && len(uris) > 0 {
		uri = uris[0]
	}

	if uri != "" && isValidURI(uri) {
		schema.uri = uri

		c.mu.RLock()
		existingSchema, exists := c.schemas[uri]
		c.mu.RUnlock()

		if exists {
			return existingSchema, nil
		}
	}

	schema.initializeSchema(c, nil)

	// Track all schemas, whether they have an ID or not
	c.mu.Lock()
	c.allSchemas = append(c.allSchemas, schema)

	if schema.uri != "" && isValidURI(schema.uri) {
		c.schemas[schema.uri] = schema
	}

	// Track unresolved references from this schema
	c.trackUnresolvedReferences(schema)

	// If this schema has a URI, check if any previously compiled schemas were waiting for it
	var schemasToResolve []*Schema
	if schema.uri != "" {
		if waitingSchemas, exists := c.unresolvedRefs[schema.uri]; exists {
			schemasToResolve = make([]*Schema, len(waitingSchemas))
			copy(schemasToResolve, waitingSchemas)
			delete(c.unresolvedRefs, schema.uri) // Clear the waiting list
		}
	}
	c.mu.Unlock()

	// Only re-resolve schemas that were actually waiting for this URI
	for _, waitingSchema := range schemasToResolve {
		waitingSchema.ResolveUnresolvedReferences()
		// Re-track any still unresolved references
		c.mu.Lock()
		c.trackUnresolvedReferences(waitingSchema)
		c.mu.Unlock()
	}

	return schema, nil
}

// trackUnresolvedReferences tracks which schemas have unresolved references to which URIs
// This method should be called with mutex locked
func (c *Compiler) trackUnresolvedReferences(schema *Schema) {
	unresolvedURIs := schema.getUnresolvedReferenceURIs()
	for _, uri := range unresolvedURIs {
		if c.unresolvedRefs[uri] == nil {
			c.unresolvedRefs[uri] = make([]*Schema, 0)
		}
		// Check if schema is already in the list to avoid duplicates
		found := false
		for _, existing := range c.unresolvedRefs[uri] {
			if existing == schema {
				found = true
				break
			}
		}
		if !found {
			c.unresolvedRefs[uri] = append(c.unresolvedRefs[uri], schema)
		}
	}
}

// resolveSchemaURL attempts to fetch and compile a schema from a URL.
func (c *Compiler) resolveSchemaURL(url string) (*Schema, error) {
	id, anchor := splitRef(url)

	c.mu.RLock()
	schema, exists := c.schemas[id]
	c.mu.RUnlock()

	if exists {
		return schema, nil // Return cached schema if available
	}

	loader, ok := c.Loaders[getURLScheme(url)]
	if !ok {
		return nil, ErrNoLoaderRegistered
	}

	body, err := loader(url)
	if err != nil {
		return nil, err
	}
	defer body.Close() //nolint:errcheck

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, ErrFailedToReadData
	}

	compiledSchema, err := c.Compile(data, id)

	if err != nil {
		return nil, err
	}

	if anchor != "" {
		return compiledSchema.resolveAnchor(anchor)
	}

	return compiledSchema, nil
}

// SetSchema associates a specific schema with a URI.
func (c *Compiler) SetSchema(uri string, schema *Schema) *Compiler {
	c.mu.Lock()
	c.schemas[uri] = schema
	c.mu.Unlock()
	return c
}

// GetSchema retrieves a schema by reference. If the schema is not found in the cache and the ref is a URL, it tries to resolve it.
func (c *Compiler) GetSchema(ref string) (*Schema, error) {
	baseURI, anchor := splitRef(ref)

	c.mu.RLock()
	schema, exists := c.schemas[baseURI]
	c.mu.RUnlock()

	if exists {
		if baseURI == ref {
			return schema, nil
		}
		return schema.resolveAnchor(anchor)
	}

	return c.resolveSchemaURL(ref)
}

// SetDefaultBaseURI sets the default base URL for resolving relative references.
func (c *Compiler) SetDefaultBaseURI(baseURI string) *Compiler {
	c.DefaultBaseURI = baseURI
	return c
}

// SetAssertFormat enables or disables format assertion.
func (c *Compiler) SetAssertFormat(assert bool) *Compiler {
	c.AssertFormat = assert
	return c
}

// RegisterDecoder adds a new decoder function for a specific encoding.
func (c *Compiler) RegisterDecoder(encodingName string, decoderFunc func(string) ([]byte, error)) *Compiler {
	c.Decoders[encodingName] = decoderFunc
	return c
}

// RegisterMediaType adds a new unmarshal function for a specific media type.
func (c *Compiler) RegisterMediaType(mediaTypeName string, unmarshalFunc func([]byte) (interface{}, error)) *Compiler {
	c.MediaTypes[mediaTypeName] = unmarshalFunc
	return c
}

// RegisterLoader adds a new loader function for a specific URI scheme.
func (c *Compiler) RegisterLoader(scheme string, loaderFunc func(url string) (io.ReadCloser, error)) *Compiler {
	c.Loaders[scheme] = loaderFunc
	return c
}

// RegisterDefaultFunc registers a function for dynamic default value generation
func (c *Compiler) RegisterDefaultFunc(name string, fn DefaultFunc) *Compiler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.defaultFuncs == nil {
		c.defaultFuncs = make(map[string]DefaultFunc)
	}
	c.defaultFuncs[name] = fn
	return c
}

// getDefaultFunc retrieves a registered default function by name
func (c *Compiler) getDefaultFunc(name string) (DefaultFunc, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fn, exists := c.defaultFuncs[name]
	return fn, exists
}

// initDefaults initializes default values for decoders, media types, and loaders.
func (c *Compiler) initDefaults() {
	c.Decoders["base64"] = base64.StdEncoding.DecodeString
	c.setupMediaTypes()
	c.setupLoaders()
}

// setupMediaTypes configures default media type handlers.
func (c *Compiler) setupMediaTypes() {
	c.MediaTypes["application/json"] = func(data []byte) (interface{}, error) {
		var temp interface{}
		if err := c.jsonDecoder(data, &temp); err != nil {
			return nil, ErrJSONUnmarshalError
		}
		return temp, nil
	}

	c.MediaTypes["application/xml"] = func(data []byte) (interface{}, error) {
		var temp interface{}
		if err := xml.Unmarshal(data, &temp); err != nil {
			return nil, ErrXMLUnmarshalError
		}
		return temp, nil
	}

	c.MediaTypes["application/yaml"] = func(data []byte) (interface{}, error) {
		var temp interface{}
		if err := yaml.Unmarshal(data, &temp); err != nil {
			return nil, ErrYAMLUnmarshalError
		}
		return temp, nil
	}
}

// setupLoaders configures default loaders for fetching schemas via HTTP/HTTPS.
func (c *Compiler) setupLoaders() {
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a reasonable timeout for network requests.
	}

	defaultHTTPLoader := func(url string) (io.ReadCloser, error) {
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, ErrFailedToFetch
		}

		if resp.StatusCode != http.StatusOK {
			err = resp.Body.Close()
			if err != nil {
				return nil, err
			}
			return nil, ErrInvalidHTTPStatusCode
		}

		return resp.Body, nil
	}

	c.RegisterLoader("http", defaultHTTPLoader)
	c.RegisterLoader("https", defaultHTTPLoader)
}

// CompileBatch compiles multiple schemas efficiently by deferring reference resolution
// until all schemas are compiled. This is the most efficient approach when you have
// many schemas with interdependencies.
func (c *Compiler) CompileBatch(schemas map[string][]byte) (map[string]*Schema, error) {
	compiledSchemas := make(map[string]*Schema)

	// First pass: compile all schemas without resolving references
	for id, schemaBytes := range schemas {
		schema, err := newSchema(schemaBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema %s: %w", id, err)
		}

		if schema.ID == "" {
			schema.ID = id
		}
		schema.uri = schema.ID

		// Initialize schema structure but skip reference resolution
		schema.compiler = c
		// Initialize basic properties without resolving references
		schema.initializeSchemaWithoutReferences(c, nil)

		compiledSchemas[id] = schema

		c.mu.Lock()
		c.allSchemas = append(c.allSchemas, schema)
		if schema.uri != "" && isValidURI(schema.uri) {
			c.schemas[schema.uri] = schema
		}
		c.mu.Unlock()
	}

	// Second pass: resolve all references at once
	for _, schema := range compiledSchemas {
		schema.resolveReferences()
	}

	return compiledSchemas, nil
}

// RegisterFormat registers a custom format.
// The optional typeName parameter specifies which JSON Schema type the format applies to
// (e.g., "string", "number"). If omitted, the format applies to all types.
func (c *Compiler) RegisterFormat(name string, validator func(interface{}) bool, typeName ...string) *Compiler {
	c.customFormatsRW.Lock()
	defer c.customFormatsRW.Unlock()

	var t string
	if len(typeName) > 0 {
		t = typeName[0]
	}

	c.customFormats[name] = &FormatDef{
		Type:     t,
		Validate: validator,
	}
	return c
}

// UnregisterFormat removes a custom format.
func (c *Compiler) UnregisterFormat(name string) *Compiler {
	c.customFormatsRW.Lock()
	defer c.customFormatsRW.Unlock()

	delete(c.customFormats, name)
	return c
}
