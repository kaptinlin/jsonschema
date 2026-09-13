package jsonschema_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
)

func TestCompileRejectsUnresolvedReferences(t *testing.T) {
	for _, ref := range []string{"urn:lifecycle:missing", "#missing", "#/$defs/missing", "#/%zz", "#/type"} {
		for _, keyword := range []string{"$ref", "$dynamicRef"} {
			t.Run(keyword+"/"+ref, func(t *testing.T) {
				doc := fmt.Appendf(nil, `{"$id":"urn:lifecycle:root","properties":{"value":{%q:%q}}}`, keyword, ref)
				schema, err := jsonschema.NewCompiler().Compile(doc)
				require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
				require.Nil(t, schema)
				require.ErrorContains(t, err, "urn:lifecycle:root")
				require.ErrorContains(t, err, "#/properties/value/"+keyword)
				require.ErrorContains(t, err, ref)
			})
		}
	}
}

func TestCompilePreservesReferenceLoaderFailure(t *testing.T) {
	cause := errors.New("loader unavailable")
	c := jsonschema.NewCompiler().RegisterLoader("urn", func(string) (io.ReadCloser, error) {
		return nil, cause
	})
	schema, err := c.Compile([]byte(`{"$ref":"urn:lifecycle:dependency"}`))
	require.Nil(t, schema)
	require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
	require.ErrorIs(t, err, cause)
}

func TestCompileBatchPublishesOnlyCompleteGraphs(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	c := jsonschema.NewCompiler().RegisterLoader("urn", func(uri string) (io.ReadCloser, error) {
		if uri != "urn:lifecycle:dependency" {
			return nil, errors.New("resource unavailable")
		}
		close(entered)
		<-release
		return io.NopCloser(strings.NewReader(`{"type":"integer"}`)), nil
	})
	done := make(chan error, 1)
	go func() {
		_, err := c.CompileBatch(map[string][]byte{
			"urn:lifecycle:batch": []byte(`{"$ref":"urn:lifecycle:dependency"}`),
		})
		done <- err
	}()
	<-entered
	schema, err := c.Schema("urn:lifecycle:batch")
	// Release before assertions so even the old implementation leaves no worker behind.
	once.Do(func() { close(release) })
	compileErr := <-done
	require.Error(t, err, "a public lookup must not see an unfinished graph")
	require.Nil(t, schema)
	require.NoError(t, compileErr)
	schema, err = c.Schema("urn:lifecycle:batch")
	require.NoError(t, err)
	require.False(t, schema.Validate(true).IsValid())
	require.True(t, schema.Validate(3).IsValid())
}

func TestCompileBatchReferenceFailureDoesNotPublish(t *testing.T) {
	c := jsonschema.NewCompiler()
	old, err := c.Compile([]byte(`{"$id":"urn:lifecycle:old","type":"integer"}`))
	require.NoError(t, err)
	batch, err := c.CompileBatch(map[string][]byte{
		"urn:lifecycle:good": []byte(`{"$ref":"urn:lifecycle:old"}`),
		"urn:lifecycle:bad":  []byte(`{"$ref":"urn:lifecycle:missing"}`),
	})
	require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
	require.Nil(t, batch)
	for _, uri := range []string{"urn:lifecycle:good", "urn:lifecycle:bad"} {
		schema, err := c.Schema(uri)
		require.Error(t, err)
		require.Nil(t, schema)
	}
	current, err := c.Schema("urn:lifecycle:old")
	require.NoError(t, err)
	require.Same(t, old, current)
	require.False(t, old.Validate(true).IsValid())
}

func TestCompileLoaderCycle(t *testing.T) {
	loads := 0
	c := jsonschema.NewCompiler().RegisterLoader("urn", func(uri string) (io.ReadCloser, error) {
		loads++
		if uri != "urn:lifecycle:child" {
			return nil, errors.New("only the child may be loaded")
		}
		return io.NopCloser(strings.NewReader(`{"type":"object","properties":{"parent":{"$ref":"urn:lifecycle:root"}}}`)), nil
	})
	schema, err := c.Compile([]byte(`{"$id":"urn:lifecycle:root","type":"object","properties":{"child":{"$ref":"urn:lifecycle:child"}}}`))
	require.NoError(t, err)
	require.Equal(t, 1, loads, "the in-progress root must be reused without loading it")
	require.True(t, schema.ValidateMap(map[string]any{"child": map[string]any{"parent": map[string]any{}}}).IsValid())
	require.False(t, schema.ValidateMap(map[string]any{"child": map[string]any{"parent": true}}).IsValid())
}

func TestCompileRejectsResourceConflicts(t *testing.T) {
	for _, batch := range []bool{false, true} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			old, err := c.Compile([]byte(`{"$id":"urn:lifecycle:occupied","type":"integer"}`))
			require.NoError(t, err)
			if batch {
				_, err = c.CompileBatch(map[string][]byte{"urn:lifecycle:occupied": []byte(`{"type":"string"}`)})
			} else {
				_, err = c.Compile([]byte(`{"$id":"urn:lifecycle:occupied","type":"string"}`))
			}
			require.Error(t, err)
			current, err := c.Schema("urn:lifecycle:occupied")
			require.NoError(t, err)
			require.Same(t, old, current)
			require.False(t, current.Validate("text").IsValid())
		})
	}
	t.Run("batch duplicate id", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		_, err := c.CompileBatch(map[string][]byte{
			"urn:lifecycle:one": []byte(`{"$id":"urn:lifecycle:duplicate","type":"integer"}`),
			"urn:lifecycle:two": []byte(`{"$id":"urn:lifecycle:duplicate","type":"string"}`),
		})
		require.Error(t, err)
		_, err = c.Schema("urn:lifecycle:duplicate")
		require.Error(t, err)
	})
	t.Run("nested duplicate id", func(t *testing.T) {
		_, err := jsonschema.NewCompiler().Compile([]byte(`{"$defs":{"a":{"$id":"urn:lifecycle:duplicate"},"b":{"$id":"urn:lifecycle:duplicate"}}}`))
		require.Error(t, err)
	})
	t.Run("resolved relative id", func(t *testing.T) {
		c := jsonschema.NewCompiler().SetDefaultBaseURI("https://example.com/schemas/")
		original, err := c.Compile([]byte(`{"$id":"./value","type":"integer"}`))
		require.NoError(t, err)
		found, err := c.Schema("./value")
		require.NoError(t, err)
		require.Same(t, original, found)
		_, err = c.Compile([]byte(`{"$id":"https://example.com/schemas/value","type":"string"}`))
		require.Error(t, err)
	})
}

func TestCompileRetrievalAliasAndRelativeID(t *testing.T) {
	var loaded []string
	c := jsonschema.NewCompiler().RegisterLoader("https", func(uri string) (io.ReadCloser, error) {
		loaded = append(loaded, uri)
		switch uri {
		case "https://example.com/doc.json":
			return io.NopCloser(strings.NewReader(`{"$id":"schemas/root.json","$defs":{"value":{"$ref":"value.json"}},"$ref":"#/$defs/value"}`)), nil
		case "https://example.com/schemas/value.json":
			return io.NopCloser(strings.NewReader(`{"type":"integer"}`)), nil
		default:
			return nil, fmt.Errorf("unexpected URI %q", uri)
		}
	})
	schema, err := c.Schema("https://example.com/doc.json#/$defs/value")
	require.NoError(t, err)
	require.False(t, schema.Validate(true).IsValid())
	canonical, err := c.Schema("https://example.com/schemas/root.json#/$defs/value")
	require.NoError(t, err)
	require.Same(t, schema, canonical)
	require.Equal(t, []string{"https://example.com/doc.json", "https://example.com/schemas/value.json"}, loaded)
}

func TestConcurrentCompileConflict(t *testing.T) {
	c := jsonschema.NewCompiler()
	start := make(chan struct{})
	done := make(chan error, 2)
	for _, typ := range []string{"integer", "string"} {
		go func() {
			<-start
			_, err := c.Compile(fmt.Appendf(nil, `{"$id":"urn:lifecycle:concurrent","type":%q}`, typ))
			done <- err
		}()
	}
	close(start)
	first, second := <-done, <-done
	require.True(t, (first == nil) != (second == nil), "exactly one definition must be accepted: %v, %v", first, second)
	stored, err := c.Schema("urn:lifecycle:concurrent")
	require.NoError(t, err)
	require.NotEqual(t, stored.Validate(3).IsValid(), stored.Validate("text").IsValid())
}

func TestPublishedGraphConcurrentCompilation(t *testing.T) {
	c := jsonschema.NewCompiler()
	schema, err := c.Compile([]byte(`{"$id":"urn:lifecycle:stable","type":"integer"}`))
	require.NoError(t, err)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			<-start
			for range 100 {
				if !schema.Validate(3).IsValid() || schema.Validate(true).IsValid() {
					t.Error("published graph changed")
				}
			}
		})
	}
	workers.Go(func() {
		<-start
		for i := range 100 {
			_, err := c.CompileBatch(map[string][]byte{
				fmt.Sprintf("urn:lifecycle:new:%d", i): []byte(`{"$ref":"urn:lifecycle:stable"}`),
			})
			if err != nil {
				t.Error(err)
			}
		}
	})
	close(start)
	workers.Wait()
}

func TestUncompiledReferencesFailBeforeEvaluation(t *testing.T) {
	for name, schema := range map[string]*jsonschema.Schema{
		"constructor":        jsonschema.Ref("urn:lifecycle:missing"),
		"negation":           jsonschema.Not(jsonschema.Ref("urn:lifecycle:missing")),
		"unused alternative": jsonschema.AnyOf(jsonschema.Any(), jsonschema.Ref("urn:lifecycle:missing")),
		"literal dynamic":    {DynamicRef: "urn:lifecycle:missing"},
	} {
		t.Run(name, func(t *testing.T) {
			for _, result := range []*jsonschema.EvaluationResult{
				schema.Validate(map[string]any{}), schema.ValidateJSON([]byte(`{}`)),
				schema.ValidateMap(map[string]any{}), schema.ValidateStruct(struct{}{}),
			} {
				require.False(t, result.IsValid())
				require.Len(t, result.Errors, 1)
				for _, err := range result.Errors {
					require.Equal(t, "unresolved_reference", err.Code)
				}
			}
		})
	}
}

func TestReusedBuilderDoesNotKeepStaleReference(t *testing.T) {
	ref := jsonschema.Ref("#/$defs/value")
	first := jsonschema.Object(
		jsonschema.Defs(map[string]*jsonschema.Schema{"value": jsonschema.Integer()}),
		jsonschema.Prop("value", ref),
	)
	require.True(t, first.ValidateMap(map[string]any{"value": 3}).IsValid())
	second := jsonschema.Object(jsonschema.Prop("value", ref))
	result := second.ValidateMap(map[string]any{"value": 3})
	require.False(t, result.IsValid(), "a failed binding must not keep an earlier target")
	require.Equal(t, "unresolved_reference", result.Errors["$ref"].Code)
}

func TestCompileConstructorDocument(t *testing.T) {
	c := jsonschema.NewCompiler()
	source := jsonschema.Object(jsonschema.Prop("value", jsonschema.Integer()))
	sourceCompiler := source.Compiler()
	data, err := source.MarshalJSON()
	require.NoError(t, err)
	stored, err := c.Compile(data, "urn:lifecycle:registered")
	require.NoError(t, err)
	require.NotSame(t, source, stored)
	require.Same(t, sourceCompiler, source.Compiler())
	(*source.Properties)["value"].Type = jsonschema.SchemaType{"string"}
	require.True(t, stored.ValidateMap(map[string]any{"value": 3}).IsValid())
	require.False(t, stored.ValidateMap(map[string]any{"value": "text"}).IsValid())
	_, err = c.Compile(data, "urn:lifecycle:registered")
	require.ErrorIs(t, err, jsonschema.ErrSchemaConflict)
	_, err = c.Compile([]byte(`{"$ref":"urn:lifecycle:missing"}`), "urn:lifecycle:bad")
	require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
	_, err = c.Schema("urn:lifecycle:bad")
	require.Error(t, err)
}

func TestComposeCompiledResources(t *testing.T) {
	c := jsonschema.NewCompiler().SetDefaultDialect(jsonschema.Draft7).SetAssertFormat(true)
	c.RegisterFormat("review-code", func(value any) bool { return value == "accepted" })
	root, err := c.Compile([]byte(`{
  "$id":"https://example.com/source",
  "$defs": {
   "target":{"type":"string","format":"review-code"},
   "alias":{"$ref":"source#/$defs/target","type":"integer"}
  },
  "allOf":[{"$ref":"#/$defs/alias"}]
 }`))
	require.NoError(t, err)
	child, err := c.Schema("https://example.com/source#/$defs/alias")
	require.NoError(t, err)
	for _, node := range []*jsonschema.Schema{root, child} {
		location := node.SchemaLocation("")
		// The parent's default dialect and policy deliberately differ from the source.
		parent := jsonschema.Object(jsonschema.Prop("value", node)).SetCompiler(jsonschema.NewCompiler())
		require.Same(t, c, node.Compiler())
		require.Equal(t, jsonschema.Draft7, node.Dialect())
		require.Equal(t, location, node.SchemaLocation(""))
		require.True(t, parent.ValidateMap(map[string]any{"value": "accepted"}).IsValid())
		require.False(t, parent.ValidateMap(map[string]any{"value": "rejected"}).IsValid())
		require.False(t, parent.ValidateMap(map[string]any{"value": 3}).IsValid())
		require.True(t, node.Validate("accepted").IsValid())
	}
}

func TestStructTagReferencesMustResolve(t *testing.T) {
	type Missing struct {
		Value any `jsonschema:"ref=#/$defs/MissingTarget"`
	}
	schema, err := jsonschema.FromStruct[Missing]()
	require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
	require.Nil(t, schema)
}

func TestConstructorsPreservePublishedChildren(t *testing.T) {
	c := jsonschema.NewCompiler()
	_, err := c.Compile([]byte(`{"$id":"urn:lifecycle:source","$defs":{"value":{"type":"integer"}}}`))
	require.NoError(t, err)
	child, err := c.Schema("urn:lifecycle:source#/$defs/value")
	require.NoError(t, err)
	location := child.SchemaLocation("")
	parent := jsonschema.Object(jsonschema.Prop("value", child))
	require.Equal(t, location, child.SchemaLocation(""), "composition must not reparent a published child")
	require.Same(t, c, child.Compiler())
	require.True(t, parent.ValidateMap(map[string]any{"value": 3}).IsValid())
	require.False(t, parent.ValidateMap(map[string]any{"value": true}).IsValid())
}

func TestConstructorsPreserveGeneratedGraphs(t *testing.T) {
	type Generated struct {
		Value int `json:"value" jsonschema:"minimum=1"`
	}
	schema, err := jsonschema.FromStructWithOptions[Generated](&jsonschema.StructTagOptions{
		CacheEnabled: false,
	})
	require.NoError(t, err)
	compiler := schema.Compiler()
	parent := jsonschema.Object(jsonschema.Prop("child", schema)).SetCompiler(jsonschema.NewCompiler())
	require.Same(t, compiler, schema.Compiler())
	require.False(t, parent.ValidateMap(map[string]any{"child": map[string]any{"value": 0}}).IsValid())
	require.True(t, parent.ValidateMap(map[string]any{"child": map[string]any{"value": 1}}).IsValid())
}

func TestReferencesUseResourceScope(t *testing.T) {
	t.Run("error identifies nested resource", func(t *testing.T) {
		_, err := jsonschema.NewCompiler().Compile([]byte(`{"$id":"urn:lifecycle:outer","$defs":{"inner":{"$id":"urn:lifecycle:inner","properties":{"value":{"$ref":"#missing"}}}}}`))
		require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
		require.ErrorContains(t, err, "urn:lifecycle:inner")
		require.ErrorContains(t, err, "#/properties/value/$ref")
	})
	t.Run("pointer starts at resource root", func(t *testing.T) {
		schema, err := jsonschema.NewCompiler().Compile([]byte(`{
			"$defs":{"value":{"type":"integer"}},
			"properties":{"value":{"$defs":{"value":{"type":"string"}},"$ref":"#/$defs/value"}}
		}`))
		require.NoError(t, err)
		require.True(t, schema.ValidateMap(map[string]any{"value": 3}).IsValid())
		require.False(t, schema.ValidateMap(map[string]any{"value": "text"}).IsValid())
	})
	t.Run("anchor cannot escape nested resource", func(t *testing.T) {
		_, err := jsonschema.NewCompiler().Compile([]byte(`{
			"$id":"urn:lifecycle:outer","$anchor":"outer",
			"$defs":{"inner":{"$id":"urn:lifecycle:inner","$ref":"#outer"}}
		}`))
		require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
	})
	t.Run("slash is not the empty pointer", func(t *testing.T) {
		_, err := jsonschema.NewCompiler().Compile([]byte(`{"$ref":"#/"}`))
		require.ErrorIs(t, err, jsonschema.ErrReferenceResolution)
	})
}

func TestEmptyReferenceUsesResourceRoot(t *testing.T) {
	for _, keyword := range []string{"$ref", "$dynamicRef"} {
		t.Run(keyword, func(t *testing.T) {
			data := fmt.Appendf(nil, `{"type":"object","properties":{"next":{%q:""}}}`, keyword)
			for range 2 {
				schema, err := jsonschema.NewCompiler().Compile(data)
				require.NoError(t, err)
				require.True(t, schema.ValidateMap(map[string]any{"next": map[string]any{}}).IsValid())
				require.False(t, schema.ValidateMap(map[string]any{"next": true}).IsValid())
				data, err = json.Marshal(schema)
				require.NoError(t, err)
			}
		})
	}
	t.Run("constructor", func(t *testing.T) {
		schema := jsonschema.Object(jsonschema.Prop("next", jsonschema.Ref("")))
		require.False(t, schema.ValidateMap(map[string]any{"next": true}).IsValid())
	})
}

func TestReferenceArrayIndicesAreCanonical(t *testing.T) {
	for _, index := range []string{"01", "+1", "-0"} {
		_, err := jsonschema.NewCompiler().Compile(fmt.Appendf(nil, `{"prefixItems":[true,true],"$ref":"#/prefixItems/%s"}`, index))
		require.ErrorIs(t, err, jsonschema.ErrReferenceResolution, index)
	}
}
