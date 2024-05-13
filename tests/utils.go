package tests

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/goccy/go-json"

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

// // Helper function to create *uint64
// func ptrUint64(v uint64) *uint64 {
// 	return &v
// }

// TestJSONSchemaTestSuiteWithFilePath runs schema validation tests from a JSON file against the ForTestSuiteschema implementation.
func testJSONSchemaTestSuiteWithFilePath(t *testing.T, filePath string, exclusions ...string) {
	t.Helper()

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
	if err := json.Unmarshal(data, &testCases); err != nil {
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
			schemaJSON, err := json.Marshal(tc.SchemaData)
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
							t.Errorf("Expected data to be valid, but got error: %v", result)
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

// Declare a global instance of sync.Once to ensure the server starts only once.
var (
	once   sync.Once
	server *http.Server
)

// startTestServer starts an HTTP server for serving remote schemas and ensures it only starts once.
func startTestServer() *http.Server {
	once.Do(func() {
		server = &http.Server{
			Addr:              ":1234",
			Handler:           http.FileServer(http.Dir("../testdata/JSON-Schema-Test-Suite/remotes")),
			ReadHeaderTimeout: 1,
		}

		// Start the server in a new goroutine
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Failed to start server: %v", err)
			}
		}()

		log.Println("Serving JSON Schema Test Suite remotes on http://localhost:1234")
	})

	return server
}
