package main

import (
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeDeclarationsArePinned(test *testing.T) {
	test.Parallel()
	api := test.TempDir()
	dir := filepath.Join(api, "node_modules/@types/node")
	if createDirectoryError := os.MkdirAll(dir, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if writeError := os.WriteFile(filepath.Join(dir, "index.d.ts"), nil, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	metadata := filepath.Join(dir, "package.json")
	if writeError := os.WriteFile(
		metadata,
		[]byte(`{"name":"@types/node","version":"25.3.3"}`),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if verifyNodeTypesError := verifyNodeTypes(api); verifyNodeTypesError != nil {
		test.Fatal(verifyNodeTypesError)
	}
	if writeError := os.WriteFile(
		metadata,
		[]byte(`{"name":"@types/node","version":"24.0.0"}`),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if verifyNodeTypesError := verifyNodeTypes(
		api,
	); verifyNodeTypesError == nil || !strings.Contains(verifyNodeTypesError.Error(), "want @types/node") {
		test.Fatal("wrong declarations accepted", verifyNodeTypesError)
	}
}

// stage3/api is provisioned by the gates' own npm, found as a build's tools are, and a box without one is told so
// by name. Mutant: execute runs plain "npm" from the caller's PATH again (the sandbox's refusal never appears).
//
// Not parallel: test.Setenv sets the process-wide ADAMIC_SANDBOX and ADAMIC_TOOLS.
func TestProvisioningFindsNpmAsBuildsFindTools(test *testing.T) {
	tree := test.TempDir()
	api := filepath.Join(tree, "stage3", "api")
	if createDirectoryError := os.MkdirAll(api, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if writeError := os.WriteFile(filepath.Join(api, "package-lock.json"), []byte("{}\n"), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	test.Setenv("ADAMIC_SANDBOX", "1")
	test.Setenv("ADAMIC_TOOLS", test.TempDir())
	prepareCheckoutError := prepareCheckout(tree)
	if prepareCheckoutError == nil || !strings.Contains(
		prepareCheckoutError.Error(),
		"provision stage3/api: npm: ADAMIC_SANDBOX=1 runs only tools under ADAMIC_TOOLS",
	) {
		test.Fatalf("provisioning without the gates' npm: %v, want the missing npm named", prepareCheckoutError)
	}
}

func TestPreparedNodeTypesUseHostAxioms(test *testing.T) {
	test.Parallel()
	api := test.TempDir()
	files, nodeHostAxiomFilesError := load.NodeHostAxiomFiles()
	if nodeHostAxiomFilesError != nil {
		test.Fatal(nodeHostAxiomFilesError)
	}
	for _, file := range files {
		published, readError := os.ReadFile(
			filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file),
		)
		if readError != nil {
			test.Fatal(readError)
		}
		name := filepath.Join(api, "node_modules", "@types", "node", file)
		if createDirectoryError := os.MkdirAll(filepath.Dir(name), 0700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(name, published, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if correctPreparedNodeTypesError := correctPreparedNodeTypes(api); correctPreparedNodeTypesError != nil {
		test.Fatal(correctPreparedNodeTypesError)
	}
	for _, file := range files {
		published, readError := os.ReadFile(
			filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file),
		)
		if readError != nil {
			test.Fatal(readError)
		}
		want, correctNodeHostAxiomsError := load.CorrectNodeHostAxioms(file, string(published))
		if correctNodeHostAxiomsError != nil {
			test.Fatal(correctNodeHostAxiomsError)
		}
		got, readError := os.ReadFile(filepath.Join(api, "node_modules", "@types", "node", file))
		if readError != nil {
			test.Fatal(readError)
		}
		if string(got) != want {
			test.Fatalf("%s: prepared checker differs from bundled checker", file)
		}
	}
}
