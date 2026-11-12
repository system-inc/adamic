package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// overloadCallbackDeepNodes is overloadCallbackNodes with an N that reaches a chain of 40 more interfaces,
// each through an optional field, so overloadCallbackHoldsAny reads about 80 types going down from N with
// no any among them, as it reads 1,896 from tsc's own Node (#68x1t2y). No value below sets a link.
func overloadCallbackDeepNodes(source string) string {
	var chain strings.Builder
	chain.WriteString("interface N { readonly kind: number; readonly link?: L1 }\n")
	for level := 1; level < 40; level++ {
		chain.WriteString("interface L" + strconv.Itoa(level) + " { readonly next?: L" + strconv.Itoa(level+1) + " }\n")
	}
	chain.WriteString("interface L40 { readonly end: number }\n")
	return strings.Replace(source, "interface N { readonly kind: number }\n", chain.String(), 1)
}

// tsc's assertNode and adapted assertEachNode over a Node whose types run deeper than 32 positions, with no
// any in them. The walk that asks whether the overload's callback parameter holds any once failed closed
// past 32 positions, so both bodies kept the census's refusal on tsc's types though each lowers over a
// shallow Node: destructuring.ts's four assertEachNode units and its assertNode unit (Correctness's census
// on #68x1t2y). Both lower here and agree with Node. Mutant: overloadCallbackAnyCap set to 32, and both
// keep the refusal.
func TestOverloadCallbackDeepNodeTypes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"assertNode", overloadCallbackDeepNodes(overloadCallbackAssertNode(overloadCallbackParameters, `    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
`) + overloadCallbackCall)},
		{"assertEachNode", overloadCallbackDeepNodes(overloadCallbackAssertEachNode("readonly ", "", overloadCallbackEachFor("const n of nodes", "")) + overloadCallbackEachCall("readonly "))},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
		})
	}
}

// A generic that holds a deeper instance of itself makes a new type at every step of the walk, so only
// overloadCallbackAnyCap ends it, and it fails closed: N reaching Grow<number> keeps the census's refusal of
// overload 1's test, and the walk finishes. Mutant: a walk past the cap answering no, and it lowers.
func TestOverloadCallbackGrowingGenericRefuses(t *testing.T) {
	t.Parallel()
	const refusal = "overload 1 of assertNode parameter test cannot be served by implementation parameter test"
	source := strings.Replace(overloadCallbackAssertNode(overloadCallbackParameters, `    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
`)+overloadCallbackCall, "interface N { readonly kind: number }\n", "interface Grow<T> { readonly value?: T; readonly next?: Grow<Grow<T>> }\ninterface N { readonly kind: number; readonly grow?: Grow<number> }\n", 1)
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, refusal) {
		t.Fatalf("want %q, got %v", refusal, err)
	}
}

// overloadCallbackHoldsAny's memo records no for every type a clean walk read, and yes only for the type a
// walk that found any started from. Parent's walk stops at its own any after it has seen Child and before
// it reads Child, so Child is asked fresh and still answers yes. Mutant: a walk that found any recording
// no for every other type it saw, and Child answers no.
func TestOverloadCallbackHoldsAnyMemo(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := `interface Leaf { readonly g: any }
interface Child { readonly c: Leaf }
interface Parent { readonly a: Child; readonly k: any }
interface Plain { readonly n: number }
interface Clean { readonly a: Plain; readonly self?: Clean }
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	declared := map[string]*checker.Type{}
	for _, statement := range file.Statements.Nodes {
		if statement.Kind == ast.KindInterfaceDeclaration {
			declared[statement.Name().Text()] = checked.GetDeclaredTypeOfSymbol(checked.GetSymbolAtLocation(statement.Name()))
		}
	}
	for _, query := range []struct {
		name string
		want bool
	}{
		{"Parent", true},
		{"Child", true},
		{"Leaf", true},
		{"Clean", false},
		{"Plain", false},
		{"Child", true},
	} {
		if got := l.overloadCallbackHoldsAny(declared[query.name]); got != query.want {
			t.Fatalf("%s holds any: got %v, want %v", query.name, got, query.want)
		}
	}
	if held, known := l.anyHolders[declared["Plain"]]; !known || held {
		t.Fatalf("Clean's walk should record Plain as holding no any, got held %v, known %v", held, known)
	}
}
