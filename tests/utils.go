package tests

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"

	"github.com/kaptinlin/jsonschema"
)

// Helper function to create *float64
func ptrFloat64(v float64) *float64 {
	return &v
}

// Helper function to create *string
func ptrString(v string) *string {
	return &v
}

// startTestServer starts an HTTP server for serving remote schemas.
func startTestServer() *http.Server {
	server := &http.Server{
		Addr:              ":1234",
		Handler:           http.FileServer(http.Dir("../testdata/JSON-Schema-Test-Suite/remotes")),
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
	return server
}

// stopTestServer stops the given HTTP server.
func stopTestServer(server *http.Server) {
	if err := server.Shutdown(context.TODO()); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}

// TestJSONSchemaTestSuiteWithFilePath runs schema validation tests from a JSON file against the ForTestSuiteschema implementation.
func testJSONSchemaTestSuiteWithFilePath(t *testing.T, filePath string, exclusions ...string) {
	t.Helper()

	// Start the server
	server := startTestServer()
	defer stopTestServer(server)

	// Read the JSON file containing the test definitions.
	data, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read test file: %s", err)
	}

	type Test struct {
		Description string      `json:"description"`
		Data        interface{} `json:"data"`
		Valid       bool        `json:"valid"`
	}
	type TestCase struct {
		Description string      `json:"description"`
		SchemaData  interface{} `json:"schema"`
		Tests       []Test      `json:"tests"`
	}
	var testCases []TestCase

	// Unmarshal the JSON into the test cases struct.
	if err := sonic.Unmarshal(data, &testCases); err != nil {
		t.Fatalf("Failed to unmarshal test cases: %v", err)
	}

	exclusionMap := make(map[string]bool)
	for _, exc := range exclusions {
		exclusionMap[exc] = true
	}

	// Run each test case.
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			if exclusionMap[tc.Description] {
				t.Skip("Skipping test case due to exclusion settings.")
			}

			// Convert schema data into JSON bytes and compile it into a Schema object.
			schemaJSON, err := sonic.Marshal(tc.SchemaData)
			if err != nil {
				t.Fatalf("Failed to marshal schema data: %v", err)
			}

			// Initialize the compiler with necessary configurations.
			compiler := jsonschema.NewCompiler()

			// Assert format for optional/format test cases.
			if strings.Contains(filePath, "optional/format") {
				compiler.SetAssertFormat(true)
			}

			schema, err := compiler.Compile(schemaJSON)
			if err != nil {
				t.Fatalf("Failed to compile schema: %v", err)
			}

			for _, test := range tc.Tests {
				if exclusionMap[tc.Description+"/"+test.Description] {
					t.Run(test.Description, func(t *testing.T) {
						t.Skip("Skipping test due to nested exclusion settings.")
					})
					continue
				}
				t.Run(test.Description, func(t *testing.T) {
					// Evaluate the data against the schema.
					result := schema.Validate(test.Data)

					// Check if the test should pass or fail.
					if test.Valid {
						if !result.IsValid() {
							t.Errorf("Expected data to be valid, but got error: %v", result.ToList())
						}
					} else {
						if result.IsValid() {
							t.Error("Expected data to be invalid, but got no error")
						}
					}
				})
			}
		})
	}
}
