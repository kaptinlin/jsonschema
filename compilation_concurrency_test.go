package jsonschema_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/jsonschema"
)

func TestConcurrentColdDependencies(t *testing.T) {
	for _, mode := range []string{"compile", "batch", "lookup", "cycle"} {
		for _, different := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/different=%t", mode, different), func(t *testing.T) {
				c := jsonschema.NewCompiler()
				entered := make(chan struct{}, 2)
				release := make(chan struct{})
				var calls atomic.Int32
				c.RegisterLoader("urn", func(uri string) (io.ReadCloser, error) {
					if uri == "urn:cold:other" {
						return io.NopCloser(strings.NewReader(`{"$ref":"urn:cold:shared"}`)), nil
					}
					if uri != "urn:cold:shared" {
						return nil, fmt.Errorf("unexpected resource %s", uri)
					}
					n := calls.Add(1)
					if n <= 2 {
						entered <- struct{}{}
						<-release
					}
					typ := "integer"
					if different && n == 2 {
						typ = "string"
					}
					doc := fmt.Sprintf(`{"type":%q}`, typ)
					if mode == "cycle" {
						doc = fmt.Sprintf(`{"type":%q,"$defs":{"other":{"$ref":"urn:cold:other"}}}`, typ)
					}
					return io.NopCloser(strings.NewReader(doc)), nil
				})
				type outcome struct {
					schema *jsonschema.Schema
					err    error
				}
				done := make(chan outcome, 2)
				for i := range 2 {
					go func() {
						var schema *jsonschema.Schema
						var err error
						switch {
						case mode == "lookup" || mode == "cycle":
							schema, err = c.Schema("urn:cold:shared")
						case mode == "batch" && i == 1:
							var batch map[string]*jsonschema.Schema
							batch, err = c.CompileBatch(map[string][]byte{"urn:cold:root:1": []byte(`{"$ref":"urn:cold:shared"}`)})
							schema = batch["urn:cold:root:1"]
						default:
							schema, err = c.Compile([]byte(`{"$ref":"urn:cold:shared"}`), fmt.Sprintf("urn:cold:root:%d", i))
						}
						done <- outcome{schema, err}
					}()
				}
				for range 2 {
					select {
					case <-entered:
					case <-time.After(5 * time.Second):
						close(release)
						t.Fatal("concurrent loaders did not start")
					}
				}
				close(release)
				var targets []*jsonschema.Schema
				for range 2 {
					select {
					case result := <-done:
						require.NoError(t, result.err)
						target := result.schema
						if mode == "compile" || mode == "batch" {
							target = target.ResolvedRef
						}
						targets = append(targets, target)
					case <-time.After(5 * time.Second):
						t.Fatal("concurrent dependency resolution did not finish")
					}
				}
				stored, err := c.Schema("urn:cold:shared")
				require.NoError(t, err)
				for _, target := range targets {
					require.Same(t, stored, target)
				}
				require.NotEqual(t, stored.Validate(3).IsValid(), stored.Validate("text").IsValid())
				if mode == "cycle" {
					other, err := c.Schema("urn:cold:other")
					require.NoError(t, err)
					require.Same(t, stored, other.ResolvedRef)
				}
			})
		}
	}
}

func TestColdDependencyFailureCanRetry(t *testing.T) {
	c := jsonschema.NewCompiler()
	cause := errors.New("temporarily unavailable")
	var calls int
	c.RegisterLoader("urn", func(string) (io.ReadCloser, error) {
		calls++
		if calls == 1 {
			return nil, cause
		}
		return io.NopCloser(strings.NewReader(`{"type":"integer"}`)), nil
	})
	_, err := c.Compile([]byte(`{"$ref":"urn:retry:shared"}`), "urn:retry:root")
	require.ErrorIs(t, err, cause)
	schema, err := c.Compile([]byte(`{"$ref":"urn:retry:shared"}`), "urn:retry:root")
	require.NoError(t, err)
	require.True(t, schema.Validate(1).IsValid())
	require.False(t, schema.Validate("text").IsValid())
	require.Equal(t, 2, calls)
}

func TestConcurrentPublicationRebuildsPrivateGraph(t *testing.T) {
	c := jsonschema.NewCompiler()
	entered := make(chan struct{}, 4)
	release := [2]chan struct{}{make(chan struct{}), make(chan struct{})}
	var loads atomic.Int32
	c.RegisterLoader("urn", func(uri string) (io.ReadCloser, error) {
		if uri == "urn:publish:shared" {
			typ := "integer"
			if loads.Add(1) == 2 {
				typ = "string"
			}
			return io.NopCloser(strings.NewReader(fmt.Sprintf(`{"type":%q}`, typ))), nil
		}
		index := 0
		if uri == "urn:publish:gate:1" {
			index = 1
		}
		entered <- struct{}{}
		<-release[index]
		return io.NopCloser(strings.NewReader(`true`)), nil
	})
	type outcome struct {
		schema *jsonschema.Schema
		err    error
	}
	done := [2]chan outcome{make(chan outcome, 1), make(chan outcome, 1)}
	for i := range 2 {
		go func() {
			schema, err := c.Compile([]byte(fmt.Sprintf(`{"allOf":[{"$ref":"urn:publish:shared"},{"$ref":"urn:publish:gate:%d"}]}`, i)), fmt.Sprintf("urn:publish:root:%d", i))
			done[i] <- outcome{schema, err}
		}()
	}
	for range 2 {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			close(release[0])
			close(release[1])
			t.Fatal("private graphs did not reach publication gates")
		}
	}
	close(release[0])
	var first outcome
	select {
	case first = <-done[0]:
	case <-time.After(5 * time.Second):
		close(release[1])
		t.Fatal("first publication did not finish")
	}
	close(release[1])
	require.NoError(t, first.err)
	select {
	case second := <-done[1]:
		require.NoError(t, second.err)
		stored, err := c.Schema("urn:publish:shared")
		require.NoError(t, err)
		require.Same(t, stored, first.schema.AllOf[0].ResolvedRef)
		require.Same(t, stored, second.schema.AllOf[0].ResolvedRef)
	case <-time.After(5 * time.Second):
		t.Fatal("private graph was not rebuilt")
	}
	require.EqualValues(t, 2, loads.Load(), "retry should reuse the published dependency")
}
