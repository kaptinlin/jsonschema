// Package main implements the schemagen code generation tool.
// This tool generates Schema methods for Go structs with jsonschema tags,
// enabling easy JSON Schema generation from Go struct definitions.
//
// Usage:
//
//	schemagen [flags] [packages...]
//
// Flags:
//
//	-suffix string     Output file suffix (default: "_schema.go")
//	-package string    Specify package name (default: auto-detect)
//	-verbose          Verbose output
//	-dry-run          Preview generated code without writing files
//	-force            Force regeneration of all files
package main

import (
	"flag"
	"fmt"
	"log"
)

// Command line flags
var (
	outputSuffix = flag.String("suffix", "_schema.go", "Output file suffix (e.g., '_jsonschema.go', '_validators.go')")
	packageName  = flag.String("package", "", "Specify package name (default: auto-detect)")
	verbose      = flag.Bool("verbose", false, "Verbose output")
	dryRun       = flag.Bool("dry-run", false, "Preview generated code without writing files")
	force        = flag.Bool("force", false, "Force regeneration of all files")
	help         = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Get target packages from command line arguments
	packages := flag.Args()
	if len(packages) == 0 {
		// Default to current directory
		packages = []string{"."}
	}

	if *verbose {
		log.Printf("🚀 Starting schemagen code generation")
		log.Printf("📦 Target packages: %v", packages)
		log.Printf("📝 Output suffix: %s", *outputSuffix)
		if *packageName != "" {
			log.Printf("📋 Package name override: %s", *packageName)
		}
		if *force {
			log.Printf("🔥 Force regeneration enabled")
		}
		if *dryRun {
			log.Printf("🔍 Dry run mode enabled")
		}
	}

	// Create generator configuration
	config := &GeneratorConfig{
		OutputSuffix: *outputSuffix,
		PackageName:  *packageName,
		Verbose:      *verbose,
		DryRun:       *dryRun,
		Force:        *force,
	}

	// Create code generator
	generator, err := NewCodeGenerator(config)
	if err != nil {
		log.Fatalf("❌ Failed to create code generator: %v", err)
	}

	// Process each package
	var hasErrors bool
	for _, pkg := range packages {
		if *verbose {
			log.Printf("📦 Processing package: %s", pkg)
		}

		err := generator.ProcessPackage(pkg)
		if err != nil {
			log.Printf("❌ Error processing package %s: %v", pkg, err)
			hasErrors = true
			continue
		}

		if *verbose {
			log.Printf("✅ Successfully processed package: %s", pkg)
		}
	}

	if hasErrors {
		log.Fatalf("❌ Code generation completed with errors")
	}

	if *verbose {
		log.Printf("🎉 Code generation completed successfully")
	}
}

// showHelp displays the help message
func showHelp() {
	fmt.Println(`schemagen - JSON Schema Code Generation Tool

Generates Schema methods for Go structs with jsonschema tags,
enabling easy JSON Schema generation from Go struct definitions.

USAGE:
    schemagen [flags] [packages...]

FLAGS:`)
	flag.PrintDefaults()
	fmt.Println(`
EXAMPLES:
    # Generate for current package
    schemagen

    # Generate for specific packages
    schemagen ./models ./api

    # Dry run to preview generated code
    schemagen -dry-run -verbose

    # Use custom output suffix
    schemagen -suffix="_jsonschema.go"

DIRECTIVES:
    Add //go:generate schemagen to your Go files to enable automatic
    code generation when running 'go generate'.

    Example:
        //go:generate schemagen
        type User struct {
            Name string ` + "`jsonschema:\"required,minLength=2\"`" + `
        }

OUTPUT:
    Generated files follow the pattern: <original>_schema.go (default)
    or with custom suffix: <original><suffix>
    Each generated file contains Schema() methods for structs with
    jsonschema tags, providing comprehensive JSON Schema validation.`)
}
