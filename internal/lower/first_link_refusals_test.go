package lower

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"testing"
)

// firstLinkPrecheckGap holds firstLinkRefuse's pre-checks to refuse's, by name and in order: the l.<check>(module)
// calls refuse makes before its contracts walk are firstLinkRefuse's list, which ends with firstLinkContracts, its
// copy of that walk. A pre-check refuse gains and firstLinkRefuse doesn't would go unjudged in a first-link
// lowering with nothing saying so. It's empty when they agree.
func firstLinkPrecheckGap(refusals, firstLink string) string {
	function := func(source, name string) *ast.FuncDecl {
		file, err := parser.ParseFile(token.NewFileSet(), name+".go", source, 0)
		if err != nil {
			return nil
		}
		for _, declaration := range file.Decls {
			if found, isFunction := declaration.(*ast.FuncDecl); isFunction && found.Name.Name == name && found.Body != nil {
				return found
			}
		}
		return nil
	}
	refuse := function(refusals, "refuse")
	firstLinkRefuse := function(firstLink, "firstLinkRefuse")
	if refuse == nil || firstLinkRefuse == nil {
		return "refuse or firstLinkRefuse not found"
	}
	// refuse's pre-checks are its top-level `if err := l.<check>(module); err != nil` statements, before the
	// contracts walk is assigned.
	var want []string
	for _, statement := range refuse.Body.List {
		if assignment, isAssignment := statement.(*ast.AssignStmt); isAssignment && len(assignment.Lhs) == 1 {
			if name, isName := assignment.Lhs[0].(*ast.Ident); isName && name.Name == "contracts" {
				break
			}
		}
		conditional, isConditional := statement.(*ast.IfStmt)
		if !isConditional || conditional.Init == nil {
			continue
		}
		initial, isAssignment := conditional.Init.(*ast.AssignStmt)
		if !isAssignment || len(initial.Rhs) != 1 {
			continue
		}
		call, isCall := initial.Rhs[0].(*ast.CallExpr)
		if !isCall {
			continue
		}
		if method, isMethod := call.Fun.(*ast.SelectorExpr); isMethod {
			if receiver, isReceiver := method.X.(*ast.Ident); isReceiver && receiver.Name == "l" {
				want = append(want, method.Sel.Name)
			}
		}
	}
	if len(want) == 0 {
		return "refuse makes no pre-check calls"
	}
	want = append(want, "firstLinkContracts")
	// firstLinkRefuse's list is the one slice of functions its pre-check loop ranges over.
	var got []string
	lists := 0
	ast.Inspect(firstLinkRefuse.Body, func(node ast.Node) bool {
		loop, isLoop := node.(*ast.RangeStmt)
		if !isLoop {
			return true
		}
		literal, isLiteral := loop.X.(*ast.CompositeLit)
		if !isLiteral {
			return true
		}
		if _, isArray := literal.Type.(*ast.ArrayType); !isArray {
			return true
		}
		lists++
		for _, element := range literal.Elts {
			method, isMethod := element.(*ast.SelectorExpr)
			if !isMethod {
				got = append(got, fmt.Sprintf("%T", element))
				continue
			}
			got = append(got, method.Sel.Name)
		}
		return true
	})
	if lists != 1 {
		return fmt.Sprintf("firstLinkRefuse ranges over %d literal lists of pre-checks, want 1", lists)
	}
	if !slices.Equal(got, want) {
		return fmt.Sprintf("firstLinkRefuse's pre-checks are %v, want refuse's %v", got, want)
	}
	return ""
}

// readRefusalSources reads refusals.go and first_link_refusals.go.
func readRefusalSources(t *testing.T) (string, string) {
	t.Helper()
	refusals, err := os.ReadFile("refusals.go")
	if err != nil {
		t.Fatal(err)
	}
	firstLink, err := os.ReadFile("first_link_refusals.go")
	if err != nil {
		t.Fatal(err)
	}
	return string(refusals), string(firstLink)
}

// A first-link lowering runs every pre-check refuse runs, by name and in refuse's order.
func TestFirstLinkRefuseRunsRefusesPrechecks(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/refusals.go
	// Reads: internal/lower/first_link_refusals.go
	refusals, firstLink := readRefusalSources(t)
	if gap := firstLinkPrecheckGap(refusals, firstLink); gap != "" {
		t.Fatal(gap)
	}
}

// Mutant: firstLinkRefuse's list without one of refuse's pre-checks, and the guard goes red.
func TestFirstLinkRefuseDroppedPrecheckMutant(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/refusals.go
	// Reads: internal/lower/first_link_refusals.go
	refusals, firstLink := readRefusalSources(t)
	mutant := strings.Replace(firstLink, "l.mergedFieldRulings, ", "", 1)
	if mutant == firstLink {
		t.Fatal("the mutant found no l.mergedFieldRulings in firstLinkRefuse's list")
	}
	if gap := firstLinkPrecheckGap(refusals, mutant); !strings.Contains(gap, "mergedFieldRulings") {
		t.Fatalf("dropping a pre-check: got %q, want the guard to name mergedFieldRulings", gap)
	}
}
