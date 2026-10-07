package load

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
)

const NodeTypesVersion = "25.3.3"
const nodeTypesRelative = "stage3/api/node_modules/@types/node"

// Resolve exactly the API seat's installed copy. Never synthesize declarations
// or fall back to a different @types/node in a caller's node_modules.
func nodeTypesIndex(directory string) (string, error) {
	for {
		root := filepath.Join(directory, filepath.FromSlash(nodeTypesRelative))
		data, err := os.ReadFile(filepath.Join(root, "package.json"))
		if err == nil {
			var pin struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			}
			if err = json.Unmarshal(data, &pin); err != nil {
				return "", fmt.Errorf("load: invalid Node type package: %w", err)
			}
			if pin.Name != "@types/node" || pin.Version != NodeTypesVersion {
				return "", fmt.Errorf("load: want @types/node %s at %s; found %s %s", NodeTypesVersion, root, pin.Name, pin.Version)
			}
			index := filepath.Join(root, "index.d.ts")
			if _, err = os.Stat(index); err != nil {
				return "", fmt.Errorf("load: Node declarations: %w", err)
			}
			index, err = filepath.EvalSymlinks(index)
			if err != nil {
				return "", fmt.Errorf("load: Node declaration path: %w", err)
			}
			return filepath.ToSlash(index), nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("load: Node type package: %w", err)
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return "", fmt.Errorf("load: node:* imports require @types/node %s installed in %s", NodeTypesVersion, nodeTypesRelative)
}

func usesNodeModules(program *compiler.Program) bool {
	for _, file := range program.GetSourceFiles() {
		for _, statement := range file.Statements.Nodes {
			if statement.Kind != ast.KindImportDeclaration && statement.Kind != ast.KindExportDeclaration {
				continue
			}
			specifier := statement.ModuleSpecifier()
			if specifier != nil && strings.HasPrefix(specifier.Text(), "node:") {
				return true
			}
		}
	}
	return false
}

// Unit lowerers use declaration identity, not the spelling of an imported name.
func IsNodeLibrary(file *ast.SourceFile) bool {
	return file != nil && strings.Contains(file.FileName().AsString(), "/"+nodeTypesRelative+"/") && strings.HasSuffix(file.FileName().AsString(), ".d.ts")
}

// Preserve the existing console lowerer's declaration identity while letting
// the pinned Node global Console interface merge. The package itself is untouched.
func nodePrelude() string {
	source := strings.Replace(prelude, "declare const console: {\n\tlog(message: string): void;\n\terror(message: string): void;\n};", "declare var console: Console;", 1)
	start := strings.Index(source, "declare const process: {")
	if start >= 0 {
		end := strings.Index(source[start:], "\n};")
		if end >= 0 {
			source = source[:start] + "declare var process: NodeJS.Process;" + source[start+end+3:]
		}
	}
	return source
}
