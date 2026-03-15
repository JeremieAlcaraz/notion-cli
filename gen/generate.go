//go:build ignore

// generate.go is run via `just generate` (or `go run ./gen/generate.go`).
// It reads spec/notion-openapi.json, applies the templates, and writes
// the generated Go source files into cmd/generated/.
package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/4ier/notion-cli/gen"
)

const (
	specPath   = "spec/notion-openapi.json"
	outputDir  = "cmd/generated"
	modulePath = "github.com/4ier/notion-cli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse spec
	spec, err := gen.ParseSpec(specPath)
	if err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}

	byTag := spec.OperationsByTag()
	fmt.Printf("Loaded spec: %s — %d tags, %d operations\n",
		spec.Info.Title, len(byTag), len(spec.Operations()))

	// Ensure output directory exists and is clean
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}
	// Remove previously generated files
	entries, _ := os.ReadDir(outputDir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".go") && e.Name() != "helpers.go" {
			os.Remove(filepath.Join(outputDir, e.Name()))
		}
	}

	// Generate one file per tag
	for tag, ops := range byTag {
		slug := strings.ToLower(strings.ReplaceAll(tag, " ", "_"))
		outPath := filepath.Join(outputDir, slug+".go")

		raw, err := gen.RenderCommandFileBytes(tag, ops)
		if err != nil {
			return fmt.Errorf("render %s: %w", tag, err)
		}

		formatted, err := format.Source(raw)
		if err != nil {
			// Write unformatted for debugging, then fail
			_ = os.WriteFile(outPath+".broken", raw, 0644)
			return fmt.Errorf("gofmt %s: %w\n(unformatted written to %s.broken)", outPath, err, outPath)
		}

		if err := os.WriteFile(outPath, formatted, 0644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		fmt.Printf("  wrote %s (%d ops)\n", outPath, len(ops))
	}

	// Generate register.go (AddTo function)
	registerPath := filepath.Join(outputDir, "register.go")
	raw, err := gen.RenderRootFileBytes(byTag)
	if err != nil {
		return fmt.Errorf("render register: %w", err)
	}
	formatted, err := format.Source(raw)
	if err != nil {
		_ = os.WriteFile(registerPath+".broken", raw, 0644)
		return fmt.Errorf("gofmt register.go: %w\n(unformatted written to %s.broken)", err, registerPath)
	}
	if err := os.WriteFile(registerPath, formatted, 0644); err != nil {
		return fmt.Errorf("write register.go: %w", err)
	}
	fmt.Printf("  wrote %s\n", registerPath)

	fmt.Println("Done.")
	return nil
}
