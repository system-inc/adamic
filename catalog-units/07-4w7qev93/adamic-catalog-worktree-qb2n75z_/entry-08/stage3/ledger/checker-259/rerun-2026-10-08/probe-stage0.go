// Tooling for the checker ledger. This loads witnesses without invoking lowering.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/system-inc/adamic/internal/load"
)

func main() {
	names := os.Args[1:]
	whole := len(names) > 0 && names[0] == "--whole"
	if whole {
		names = names[1:]
	}
	for _, name := range names {
		paths := []string{name}
		if whole {
			paths = names
			name = "<whole>"
		}
		var err error
		if os.Getenv("LEDGER_STRICT") == "1" {
			tree := os.Getenv("LEDGER_TREE")
			configPath := filepath.Join(tree, "src/compiler/tsconfig.json")
			data, readErr := os.ReadFile(configPath)
			if readErr != nil {
				panic(readErr)
			}
			if strings.Count(string(data), `"types": ["node"]`) != 1 {
				panic("unexpected compiler config shape")
			}
			config := strings.Replace(string(data), `"types": ["node"]`, `"types": ["node"], "noUncheckedIndexedAccess": true, "exactOptionalPropertyTypes": true, "useUnknownInCatchVariables": true, "strictBindCallApply": true`, 1)
			if os.Getenv("LEDGER_PROFILE") == "census" {
				config = `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":true,"exactOptionalPropertyTypes":true,"erasableSyntaxOnly":false,"verbatimModuleSyntax":true,"allowImportingTsExtensions":true,"noEmit":true,"module":"ESNext","moduleDetection":"force","moduleResolution":"Bundler","target":"ES2024","lib":["es2024"],"types":["node"]},"include":["**/*"]}`
			}
			_, err = load.LoadOverlay(paths, map[string]string{configPath: config})
		} else {
			_, err = load.Load(paths)
		}
		result := struct {
			File        string   `json:"file"`
			Diagnostics []string `json:"diagnostics"`
			Error       string   `json:"error,omitempty"`
		}{File: name, Diagnostics: []string{}}
		if err != nil {
			if check, ok := err.(*load.CheckError); ok {
				result.Diagnostics = check.Diagnostics
			} else {
				result.Error = err.Error()
			}
		}
		data, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		if whole {
			break
		}
	}
}
