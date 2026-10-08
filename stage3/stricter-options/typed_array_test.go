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

func TestTypedArrayRepresentationBoundaries(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ name, body, message string }{
		{"for-of", "for (const value of values) {console.log(`${value}`);}", "typed array iteration"},
		{"iterator", "for (const value of values.values()) {console.log(`${value}`);}", "typed array methods"},
		{"data-input", "const copy = new Uint8Array([1, 2]);", "typed array length not statically proven"},
		{"fractional-length", "const copy = new Uint8Array(1.5);", "typed array length not statically proven"},
		{"dynamic-length", "const length = parseInt('2'); const copy = new Uint8Array(length);", "typed array length not statically proven"},
		{"method", "values.fill(7);", "typed array methods"},
		{"write", "values[0] = 257;", "typed array writes"},
		{"buffer", "const buffer = values.buffer;", "typed array properties other than length"},
		{"keys", "Object.keys(values);", "Object operations on typed arrays"},
		{"from", "const copy = Array.from(values);", "library iteration over typed arrays"},
		{"spread", "const copy = {...values};", "typed array spread"},
		{"view", "const view: {length:number} = values;", "seen as"},
		{"cast-view", "const view = values as {readonly length:number};", "cast changing the storage representation"},
		{"stringify", "JSON.stringify(values);", "JSON.stringify typed arrays"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			directory := t.TempDir()
			for name, contents := range map[string]string{
				"main.ts":       "const values = new Uint8Array(2); " + fixture.body,
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
				t.Fatalf("want explicit typed-array refusal %q, got %v", fixture.message, err)
			}
		})
	}
}
