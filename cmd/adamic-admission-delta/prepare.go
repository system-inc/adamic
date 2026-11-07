package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Tools come from cloud/setup.sh's environment. Checkout-local declarations
// must instead be installed from this checkout's own package lock.
func prepareCheckout(tree string) error {
	api := filepath.Join(tree, "stage3", "api")
	if _, statError := os.Stat(filepath.Join(api, "package-lock.json")); os.IsNotExist(statError) {
		return nil
	} else if statError != nil {
		return statError
	}
	// npm is the gates' own, found as a build's tools are (ADAMIC_TOOLS, then PATH), since a house box's default PATH
	// has none (Pupa, #sa5ptyj); execute runs it under toolchain.Env, whose PATH also finds the node it needs.
	npm, pathError := toolchain.Path("npm")
	if pathError != nil {
		return fmt.Errorf("provision stage3/api: %w", pathError)
	}
	install := execute(api, 90*time.Second, npm, "ci", "--ignore-scripts")
	if install.Exit != 0 || install.Error != "" {
		return fmt.Errorf("provision stage3/api: %s %s", install.Error, install.Stderr)
	}
	if verifyNodeTypesError := verifyNodeTypes(api); verifyNodeTypesError != nil {
		return verifyNodeTypesError
	}
	return correctPreparedNodeTypes(api)
}
func verifyNodeTypes(api string) error {
	path := filepath.Join(api, "node_modules", "@types", "node", "package.json")
	data, readError := os.ReadFile(path)
	if readError != nil {
		return readError
	}
	var pin struct {
		Name    string
		Version string
	}
	if decodeError := json.Unmarshal(data, &pin); decodeError != nil {
		return decodeError
	}
	if pin.Name != "@types/node" || pin.Version != "25.3.3" {
		return fmt.Errorf("want @types/node 25.3.3, found %s %s", pin.Name, pin.Version)
	}
	_, statError := os.Stat(filepath.Join(filepath.Dir(path), "index.d.ts"))
	return statError
}

// npm ci lays the untouched publication first. The stock checker then sees the
// same host corrections as Adamic's bundled checker, never a second row list.
func correctPreparedNodeTypes(api string) error {
	files, nodeHostAxiomFilesError := load.NodeHostAxiomFiles()
	if nodeHostAxiomFilesError != nil {
		return nodeHostAxiomFilesError
	}
	for _, file := range files {
		name := filepath.Join(api, "node_modules", "@types", "node", filepath.FromSlash(file))
		published, readError := os.ReadFile(name)
		if readError != nil {
			return readError
		}
		corrected, correctNodeHostAxiomsError := load.CorrectNodeHostAxioms(file, string(published))
		if correctNodeHostAxiomsError != nil {
			return correctNodeHostAxiomsError
		}
		if writeError := os.WriteFile(name, []byte(corrected), 0600); writeError != nil {
			return writeError
		}
	}
	return nil
}
