package main

import (
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeDeclarationsArePinned(t *testing.T) {
	t.Parallel()
	api := t.TempDir()
	dir := filepath.Join(api, "node_modules/@types/node")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.d.ts"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(dir, "package.json")
	if err := os.WriteFile(metadata, []byte(`{"name":"@types/node","version":"25.3.3"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyNodeTypes(api); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadata, []byte(`{"name":"@types/node","version":"24.0.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyNodeTypes(api); err == nil || !strings.Contains(err.Error(), "want @types/node") {
		t.Fatal("wrong declarations accepted", err)
	}
}

// stage3/api is provisioned by the gates' own npm, found as a build's tools are, and a box without one is told so
// by name. Mutant: execute runs plain "npm" from the caller's PATH again (the sandbox's refusal never appears).
//
// Not parallel: t.Setenv sets the process-wide ADAMIC_SANDBOX and ADAMIC_TOOLS.
func TestProvisioningFindsNpmAsBuildsFindTools(t *testing.T) {
	tree := t.TempDir()
	api := filepath.Join(tree, "stage3", "api")
	if err := os.MkdirAll(api, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(api, "package-lock.json"), []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_SANDBOX", "1")
	t.Setenv("ADAMIC_TOOLS", t.TempDir())
	err := prepareCheckout(tree)
	if err == nil || !strings.Contains(err.Error(), "provision stage3/api: npm: ADAMIC_SANDBOX=1 runs only tools under ADAMIC_TOOLS") {
		t.Fatalf("provisioning without the gates' npm: %v, want the missing npm named", err)
	}
}

func TestPreparedNodeTypesUseHostAxioms(t *testing.T) {
	t.Parallel()
	api := t.TempDir()
	files, err := load.NodeHostAxiomFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		published, err := os.ReadFile(filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(api, "node_modules", "@types", "node", file)
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, published, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := correctPreparedNodeTypes(api); err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		published, err := os.ReadFile(filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		want, err := load.CorrectNodeHostAxioms(file, string(published))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(api, "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s: prepared checker differs from bundled checker", file)
		}
	}
}
