package cloud

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type moduleTargets struct {
	Packages []string `json:"packages"`
	Calls    []string `json:"calls"`
	Imports  []string `json:"imports"`
}

// All Go calls are reviewed, including root commands, so a newly introduced
// cohere cwd or a computed target cannot escape the explicit inventory.
func stage1GoInputs(root string) (calls, imports []string, err error) {
	seen := map[string]bool{}
	err = filepath.WalkDir(filepath.Join(root, "stage1"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, item := range file.Imports {
			name, _ := strconv.Unquote(item.Path.Value)
			if strings.Contains(name, ".") && !strings.HasPrefix(name, "github.com/system-inc/adamic/") {
				seen[name] = true
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if assignment, ok := node.(*ast.AssignStmt); ok {
				for _, left := range assignment.Lhs {
					if field, ok := left.(*ast.SelectorExpr); ok && field.Sel.Name == "Dir" {
						var output bytes.Buffer
						_ = format.Node(&output, fset, assignment)
						relative, _ := filepath.Rel(root, path)
						calls = append(calls, filepath.ToSlash(relative)+": "+output.String())
					}
				}
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			for index, argument := range call.Args {
				value, ok := argument.(*ast.BasicLit)
				if !ok || value.Kind != token.STRING {
					continue
				}
				text, _ := strconv.Unquote(value.Value)
				if text != "go" || index+1 >= len(call.Args) {
					continue
				}
				operation, ok := call.Args[index+1].(*ast.BasicLit)
				if !ok || operation.Kind != token.STRING {
					continue
				}
				op, _ := strconv.Unquote(operation.Value)
				if op != "build" && op != "test" && op != "run" {
					continue
				}
				var output bytes.Buffer
				_ = format.Node(&output, fset, call)
				relative, _ := filepath.Rel(root, path)
				calls = append(calls, filepath.ToSlash(relative)+": "+output.String())
			}
			return true
		})
		return nil
	})
	for name := range seen {
		imports = append(imports, name)
	}
	sort.Strings(calls)
	sort.Strings(imports)
	return
}

func checkModuleTargets(expected moduleTargets, calls, imports []string) error {
	if !reflect.DeepEqual(expected.Calls, calls) || !reflect.DeepEqual(expected.Imports, imports) {
		return fmt.Errorf("stage1 Go invocation/import coverage changed: review every new cohere target, update cohere-module-targets.json packages, then run go test ./cloud -run TestCohereModuleTargets -args -update-module-targets")
	}
	for _, name := range imports {
		if strings.HasPrefix(name, "github.com/system-inc/cohere/") && !contains(expected.Packages, name) {
			return fmt.Errorf("cohere overlay import outside closure: %s", name)
		}
	}
	return nil
}
func contains(values []string, name string) bool {
	for _, value := range values {
		if value == name {
			return true
		}
	}
	return false
}
