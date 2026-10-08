package lower_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/refusalprobe"
)

// TestRefusalHelpersRegistered holds the refusal catalog's helper owners in this package, so an
// area that adds a helper to the refusal pass learns it here, not after a merge. Each helper the
// pass calls must register itself beside its definition, naming an entry the catalog has.
// It is an external test because internal/refusalprobe imports this package.
func TestRefusalHelpersRegistered(t *testing.T) {
	t.Parallel()
	calls, err := refusalprobe.RefusalHelperCalls("refusals.go")
	if err != nil {
		t.Fatal(err)
	}
	registered := lower.RefusalHelpers()
	for _, helper := range calls {
		if _, ok := registered[helper]; !ok {
			t.Errorf("refusal helper %s is called by refusals.go but has no registration; add var _ = registerRefusalHelper(%q, \"<catalog entry>\") beside its definition", helper, helper)
		}
	}
	entries := map[string]bool{}
	for _, entry := range refusalprobe.Catalog() {
		entries[entry.Name] = true
	}
	for helper, entry := range registered {
		if !entries[entry] {
			t.Errorf("refusal helper %s is registered to %q, which internal/refusalprobe's catalog lacks", helper, entry)
		}
		if !slices.Contains(calls, helper) {
			t.Errorf("refusal helper %s is registered but refusals.go no longer calls it", helper)
		}
	}

	// A registration belongs in the file that defines its method, where whoever edits the helper sees it.
	registrations := map[string]string{}
	methods := map[string]string{}
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		syntax, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range syntax.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv != nil {
				methods[function.Name.Name] = name
			}
		}
		ast.Inspect(syntax, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee, ok := call.Fun.(*ast.Ident)
			if !ok || callee.Name != "registerRefusalHelper" || len(call.Args) != 2 {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok {
				t.Errorf("%s: a refusal helper registration must name its helper with a string literal", name)
				return true
			}
			helper, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			registrations[helper] = name
			return true
		})
	}
	for helper := range registered {
		if registrations[helper] == "" {
			t.Errorf("refusal helper %s is registered outside a package-level declaration this test can read", helper)
			continue
		}
		if methods[helper] != registrations[helper] {
			t.Errorf("refusal helper %s is registered in %s but defined in %q; keep the registration beside its definition", helper, registrations[helper], methods[helper])
		}
	}
}
