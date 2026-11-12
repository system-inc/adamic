package lower

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

const overloadClassInitializerShape = `interface Node { readonly kind: number }
interface Tagged extends Node { readonly kind: 1 }
function tag(n: Node, sure: true): Tagged;
function tag(n: Node, sure: boolean): Node;
function tag(n: Node, sure: boolean): Node { return n; }
const n: Node = {kind: KIND};
CLASS
const c = new C();
console.log('after');
`

// Mutant run and restored: bypass both noteOverloadCall and overloadResultCheck for the field call.
// Without the initializer walk the lying program prints after; with it, lowering panics naming
// the call lowered without its result check, before either backend can build the program.
func TestOverloadClassInitializers(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, declaration string }{
		{"instance", "class C { value = tag(n, true); }"},
		{"generic instance stays on notes", "class C<T> { value = tag(n, true); }"},
		{"static block", "class C { static { const value = tag(n, true); console.log(String(value.kind)); } }"},
		{"static field", "class C { static value = tag(n, true); }"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			source := strings.Replace(overloadClassInitializerShape, "CLASS", row.declaration, 1)
			honest := lowersAndAgreesWithNodeAs(t, strings.Replace(source, "KIND", "1", 1), ".ts")
			if honest.OverloadResultChecks.Checked != 1 {
				t.Fatalf("want one checked call, got %+v", honest.OverloadResultChecks)
			}
			lie := strings.Replace(source, "KIND", "2", 1)
			overloadCheckStops(t, lie, "", "overload 1 of tag result at ", ": expected Tagged")
		})
	}
}

func TestOverloadUnconstructedClassInitializers(t *testing.T) {
	t.Parallel()
	for _, exported := range []bool{false, true} {
		name := "unused"
		prefix := ""
		if exported {
			name = "exported without importers"
			prefix = "export "
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source := strings.Replace(overloadClassInitializerShape, "CLASS", prefix+"class C { value = tag(n, true); }", 1)
			source = strings.Replace(source, "const c = new C();\n", "", 1)
			program := lowersAndAgreesWithNodeAs(t, strings.Replace(source, "KIND", "2", 1), ".ts")
			if counts := program.OverloadResultChecks; counts.Checked != 0 || counts.Proven != 0 || counts.Erased != 0 {
				t.Fatalf("unlowered field acquired a call: %+v", counts)
			}
		})
	}
}

// Static execution does not imply that the instance field was lowered.
func TestOverloadUnusedInstanceBesideStaticBlock(t *testing.T) {
	t.Parallel()
	source := strings.Replace(overloadClassInitializerShape, "CLASS", "class C { value = tag(n, true); static { console.log('static'); } }", 1)
	source = strings.Replace(source, "const c = new C();\n", "", 1)
	program := lowersAndAgreesWithNodeAs(t, strings.Replace(source, "KIND", "2", 1), ".ts")
	if program.OverloadResultChecks.Checked != 0 {
		t.Fatalf("unused instance call was lowered: %+v", program.OverloadResultChecks)
	}
}

// The audit must find a missing call independently of the notes. Removing its
// initializer walk makes this valid AST with an unsettled narrow call pass.
func TestOverloadClassInitializerAuditCatchesMissingCall(t *testing.T) {
	t.Parallel()
	source := strings.ReplaceAll(overloadClassInitializerShape, "KIND", "2")
	source = strings.Replace(source, "CLASS", "class C { value = tag(n, true); }", 1)
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	var class, root *ast.Node
	var find ast.Visitor
	find = func(n *ast.Node) bool {
		if n.Kind == ast.KindClassDeclaration {
			class = n
			root = n.Members()[0].AsPropertyDeclaration().Initializer
			return true
		}
		return n.ForEachChild(find)
	}
	program.Files()[0].AsNode().ForEachChild(find)
	if root == nil {
		t.Fatal("missing initializer")
	}
	overload := checked.GetResolvedSignature(root).Declaration()
	l := lowering{program: program, checker: checked, narrowOverloads: map[*ast.Node]bool{overload: true}, classInitializers: map[*ast.Node]map[*ast.Node]bool{class: {root: true}}}
	want := fmt.Sprintf("lower: compiler bug: a call to an overload accepted narrow at %s lowered without its result check", program.Where(root))
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("want %q, got %v", want, got)
		}
	}()
	l.checkOverloadResults(nil)
}
