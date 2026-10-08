package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// These are representation boundaries, rather than witnesses for individual ledger sites.
// Until these operations preserve holes, refusing them keeps dense-only runtime loops sound.
func TestSparseRepresentationBoundaries(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ name, body, message string }{
		{"slice", "values.slice();", "array methods other than fill, push and at"},
		{"join", "values.join();", "array methods other than fill, push and at"},
		{"for-of", "for (const value of values) { console.log(`${value}`); }", "for...of over arrays"},
		{"entries", "for (const [index, value] of values.entries()) { console.log(`${index}:${value}`); }", "array methods other than fill, push and at"},
		{"spread", "const copied = [...values];", "array spread"},
		{"stringify", "JSON.stringify(values);", "JSON.stringify arrays"},
		{"from", "Array.from(values);", "array iteration by a library function"},
		{"set", "const copied = new Set(values);", "collection construction from arrays"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			directory := t.TempDir()
			source := "const values: number[] = new Array<number>(2); " + fixture.body
			for name, contents := range map[string]string{
				"main.ts":       source,
				"tsconfig.json": `{"compilerOptions":{"strict":true,"lib":["es2024"],"module":"esnext","moduleDetection":"force"},"files":["main.ts"]}`,
			} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0644); err != nil {
					t.Fatal(err)
				}
			}
			checked, err := load.Load([]string{filepath.Join(directory, "main.ts")})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), checked)
			if err == nil || !strings.Contains(err.Error(), fixture.message) {
				t.Fatalf("want explicit sparse refusal %q, got %v", fixture.message, err)
			}
		})
	}
}
