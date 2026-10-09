package release

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestDispatcherPackageShipsTheSettingsSchemas builds the dispatcher package from this module and
// requires every settings schema in it, byte for byte the committed one, with schema/ in the files npm
// publishes. A schema left out of `files` would be staged here and dropped by `npm publish`.
func TestDispatcherPackageShipsTheSettingsSchemas(t *testing.T) {
	t.Parallel()

	module, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	staged, err := buildDispatcherPackage(Options{ModuleDirectory: module, OutputDirectory: t.TempDir(), Version: "1.0.0"}, placeholderChecksums)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range SettingsSchemaFileNames {
		committed, err := os.ReadFile(filepath.Join(module, SchemaDirectoryName, name))
		if err != nil {
			t.Fatal(err)
		}
		shipped, err := os.ReadFile(filepath.Join(staged.Directory, SchemaDirectoryName, name))
		if err != nil {
			t.Fatalf("the dispatcher package does not ship %s: %v", name, err)
		}
		if !bytes.Equal(shipped, committed) {
			t.Fatalf("the dispatcher ships a %s that differs from the committed one", name)
		}
	}

	manifestBytes, err := os.ReadFile(filepath.Join(staged.Directory, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(manifest.Files, SchemaDirectoryName+"/") {
		t.Fatalf("the dispatcher's files are %v, so npm publish would leave the schemas out", manifest.Files)
	}
}

// TestDispatcherPackageRefusesASchemaThatValidatesNothing holds each way the copy could ship a $schema
// target that checks nothing: the file absent, empty, or not JSON.
func TestDispatcherPackageRefusesASchemaThatValidatesNothing(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name     string
		contents *string
		says     string
	}{
		{"absent", nil, "reading the settings schema"},
		{"an empty object", new("{}"), "is not a JSON schema"},
		{"not JSON", new("{\"type\":"), "is not a JSON schema"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			module := t.TempDir()
			for _, name := range SettingsSchemaFileNames {
				writeFile(t, filepath.Join(module, SchemaDirectoryName, name), `{"type":"object"}`)
			}
			broken := filepath.Join(module, SchemaDirectoryName, SettingsSchemaFileNames[len(SettingsSchemaFileNames)-1])
			if testCase.contents == nil {
				if err := os.Remove(broken); err != nil {
					t.Fatal(err)
				}
			} else {
				writeFile(t, broken, *testCase.contents)
			}

			_, err := buildDispatcherPackage(Options{ModuleDirectory: module, OutputDirectory: t.TempDir(), Version: "1.0.0"}, placeholderChecksums)
			if err == nil {
				t.Fatal("the dispatcher was staged with a schema that validates nothing")
			}
			if !strings.Contains(err.Error(), testCase.says) {
				t.Fatalf("refused, but without saying %q: %v", testCase.says, err)
			}
		})
	}
}
