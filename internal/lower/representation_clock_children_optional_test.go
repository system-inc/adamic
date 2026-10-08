package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestClockChildrenOptionalRepresentation(t *testing.T) {
	for _, probe := range []struct {
		name, constraint string
		admitted         bool
	}{
		{"readonly array", "readonly ChildNode[]", true},
		{"mutable array", "ChildNode[]", true},
		{"array ancestry", "NodeArray<ChildNode>", true},
		{"tuple", "readonly [ChildNode, ChildNode]", false},
		{"structural length", "{ readonly length: number }", false},
		{"unknown constraint", "unknown", false},
		{"union constraint", "ChildNode[] | { readonly length: number }", false},
		{"callable constraint", "() => void", false},
		{"unconstrained", "", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main.a")
			constraint := ""
			if probe.constraint != "" {
				constraint = " extends " + probe.constraint
			}
			source := `interface ChildNode { readonly text: string }
interface NodeArray<T> extends ReadonlyArray<T> { readonly pos: number }
function probe<Children` + constraint + `>(children: Children | undefined): void {}`
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{checker: checked, program: program}
			parameter := file.Statements.Nodes[2].AsFunctionDeclaration().Parameters.Nodes[0]
			proven := checked.GetTypeAtLocation(parameter.Name())
			got, known := l.representation(proven)
			if known != probe.admitted || known && got != ir.Array {
				t.Fatalf("optional representation = %v, %t; want Array admitted %t", got, known, probe.admitted)
			}
			unresolved := checked.GetTypeAtLocation(file.Statements.Nodes[2].AsFunctionDeclaration().TypeParameters.Nodes[0].Name())
			got, known = l.representation(unresolved)
			if known != probe.admitted || known && got != ir.Array {
				t.Fatalf("unresolved representation = %v, %t; want Array admitted %t", got, known, probe.admitted)
			}
			l.substitution = map[*checker.Type]ir.Type{unresolved: ir.Object}
			if got, known := l.representation(unresolved); !known || got != ir.Object {
				t.Fatalf("substitution lost: %v %t", got, known)
			}
		})
	}
}

func TestClockChildrenOptionalHeaderABI(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/representation_clock_children_optional.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "size_") && len(function.Parameters) == 1 {
			if got := program.Locals[function.Parameters[0]].Type; got != ir.Array {
				t.Fatalf("size header parameter = %v, want Array", got)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("no concrete size array header")
	}
}

func TestClockChildrenOptionalKeepsArrayBoundaries(t *testing.T) {
	for _, probe := range []struct{ name, source, reason string }{
		{"tuple", `interface ChildNode { readonly text: string }
function size<Children extends readonly ChildNode[]>(children: Children | undefined): number { return children === undefined ? -1 : children.length; }
const pair: readonly [ChildNode, ChildNode] = [{ text: 'a' }, { text: 'b' }];
const count = size<readonly ChildNode[]>(pair);`, "tuple"},
		{"element keeping", `import type { Weak } from 'adamic';
interface ChildNode { readonly text: string }
function size<Children extends readonly Weak<ChildNode>[]>(children: Children | undefined): number { return children === undefined ? -1 : children.length; }
const children: readonly ChildNode[] = [{ text: 'a' }];
const count = size<readonly Weak<ChildNode>[]>(children);`, "weakly"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %s refusal, got %v", probe.reason, err)
			}
		})
	}
}
