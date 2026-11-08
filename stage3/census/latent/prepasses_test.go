package latent

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Independent units do not execute module initializers. These exemptions describe
// existing census semantics, not permission for newly added passes to be skipped.
var latentPrepassExemptions = map[string]string{
	"enumInitialization":      "Validates enum declarations before production registration. Census registration prepares enum objects lazily and reports validation at actual uses; refusing every unrelated independent unit would invent a first stop.",
	"namespaceInitialization": "Checks readiness along the module execution schedule, including calls from initializers. Independent census units do not execute that schedule. Actual namespace and enum accesses retain production runtime readiness checks.",
}

var latentPrepassTemplates = []string{
	"stage3/census/latent/lower.go.txt",
	"stage3/census/latent/speculative.go.txt",
	"stage3/census/latent/units.go.txt",
	"stage3/census/predicates/tools/lower.go.txt",
}

func latentPrepassRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Lower delegates on current main. Follow its local call graph and require one
// reachable moduleOrder-to-refuse window, rather than silently matching any
// unrelated function or pinning the wrapper's current name.
func latentProductionPrepasses(source string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "internal/lower/lower.go", source, 0)
	if err != nil {
		return nil, err
	}
	functions := map[string]*ast.FuncDecl{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
			functions[fn.Name.Name] = fn
		}
	}
	seen := map[string]bool{}
	var windows [][]string
	var visit func(string)
	visit = func(name string) {
		if seen[name] {
			return
		}
		seen[name] = true
		fn := functions[name]
		if fn == nil || fn.Body == nil {
			return
		}
		active, done := false, false
		var passes []string
		for _, statement := range fn.Body.List {
			ast.Inspect(statement, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if id, ok := call.Fun.(*ast.Ident); ok {
					visit(id.Name)
				}
				method, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				receiver, ok := method.X.(*ast.Ident)
				if !ok || receiver.Name != "lowering" {
					return true
				}
				switch method.Sel.Name {
				case "moduleOrder":
					active = true
				case "refuse":
					if active {
						done = true
					}
				default:
					if active && !done {
						passes = append(passes, method.Sel.Name)
					}
				}
				return true
			})
		}
		if active && done {
			windows = append(windows, passes)
		}
	}
	visit("Lower")
	if len(windows) != 1 || len(windows[0]) == 0 {
		return nil, fmt.Errorf("Lower: expected one reachable moduleOrder-to-refuse prepass window, found %d", len(windows))
	}
	return windows[0], nil
}

func latentPrepassGaps(production string, templates map[string]string) []string {
	passes, err := latentProductionPrepasses(production)
	if err != nil {
		return []string{err.Error()}
	}
	known := map[string]bool{}
	for _, method := range passes {
		known[method] = true
	}
	var gaps []string
	for method, reason := range latentPrepassExemptions {
		if !known[method] || strings.TrimSpace(reason) == "" {
			gaps = append(gaps, "stale or unexplained prepass exemption: "+method)
		}
	}
	for path, source := range templates {
		file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
		if err != nil {
			gaps = append(gaps, err.Error())
			continue
		}
		var calls []string
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			method, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !known[method.Sel.Name] {
				return true
			}
			receiver, ok := method.X.(*ast.Ident)
			if !ok || (receiver.Name != "l" && receiver.Name != "base") {
				return true
			}
			calls = append(calls, method.Sel.Name)
			return true
		})
		cursor := 0
		for _, method := range passes {
			if latentPrepassExemptions[method] != "" && (cursor >= len(calls) || calls[cursor] != method) {
				continue
			}
			if cursor >= len(calls) || calls[cursor] != method {
				gaps = append(gaps, fmt.Sprintf("%s: production prepass %s missing or out of order (calls %v)", path, method, calls))
				continue
			}
			cursor++
		}
		if cursor < len(calls) {
			gaps = append(gaps, fmt.Sprintf("%s: extra or out of order prepass calls %v", path, calls[cursor:]))
		}
	}
	return gaps
}

func TestLatentCensusPrepassesTrackProductionSygfnmc(t *testing.T) {
	t.Parallel()
	root := "../../.."
	templates := map[string]string{}
	for _, path := range latentPrepassTemplates {
		templates[path] = latentPrepassRead(t, filepath.Join(root, path))
	}
	for _, gap := range latentPrepassGaps(latentPrepassRead(t, filepath.Join(root, "internal/lower/lower.go")), templates) {
		t.Error(gap)
	}
}
