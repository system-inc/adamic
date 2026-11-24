package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// widenedMembers is a target a literal partly writes, its fields held every way the widened rule admits: a string,
// an array, a number, an object and optional ones. A literal that writes origin, optional in Big, has a type Big
// doesn't extend, so its cast is sideways, as commandLineParser.ts:2291's is (N4 head B on #b5w3ycg).
const widenedMembers = "interface Owner { name: string } interface Item { label: string } interface Big { name: string; items: Item[]; count: number; parent: Owner; origin?: Owner; note?: string; size?: number } "

// widenedProbe loads source as main<extension> and gives a lowering over it with its one cast, which
// constructionCast and constructionJSON are asked of directly: behavior can't say which road a cast took.
func widenedProbe(t *testing.T, source, extension string) (*lowering, *ast.Node) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main"+extension)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
	t.Cleanup(release)
	var cast *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindAsExpression {
			cast = node
		}
		node.ForEachChild(visit)
		return false
	}
	loaded.Files()[0].AsNode().ForEachChild(visit)
	if cast == nil {
		t.Fatal("no cast in the probe")
	}
	return &lowering{program: loaded, checker: checked, result: &ir.Program{}}, cast
}

// constructionCast is the only way castProof sets construction, so its answer is the road's. A sideways literal
// to a mixed target takes the widened road in .ts; .a keeps its sideways refusal. A literal the target already has a
// road for keeps it: one assignable to the target is castProof's upcast, and one whose type the target extends is a
// structural view, both reached only after constructionCast says no. Mutant: constructionWidened's castSideways
// guard dropped; the structural view row goes red. A written field whose type isn't the slot's stays stopped, and
// the scalar-only rule answers as it did.
func TestWidenedConstructionRoads(t *testing.T) {
	t.Parallel()
	const held = "const boss: Owner = { name: 'boss' }; const item: Item = { label: 'l' }; "
	for _, probe := range []struct {
		name, source, extension string
		construction            bool
	}{
		{"sideways literal to a mixed target", widenedMembers + held + "const big = { name: 'n', items: [item], origin: boss } as Big;", ".ts", true},
		{"sideways literal in .a", widenedMembers + held + "const big = { name: 'n', items: [item], origin: boss } as Big;", ".a", false},
		{"literal assignable to the target", widenedMembers + held + "const big = { name: 'n', items: [item], count: 1, parent: boss, origin: boss } as Big;", ".ts", false},
		{"literal the target extends", widenedMembers + held + "const big = { name: 'n', items: [item] } as Big;", ".ts", false},
		{"written field wider than its slot", widenedMembers + "function f(thing: Item | Owner, boss: Owner): Big { return { name: 'n', items: [thing], origin: boss } as Big; }", ".ts", false},
		{"scalar target", "interface S { a: number; b: string } const s = { a: 1 } as S;", ".ts", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, cast := widenedProbe(t, probe.source, probe.extension)
			if got := l.constructionCast(cast); got != probe.construction {
				t.Fatalf("constructionCast says %v, want %v", got, probe.construction)
			}
		})
	}
}

// A widened construction has no JSON descriptor for its object, array and union slots, so constructionJSON says
// it doesn't know, as for anything it can't describe, rather than a schema with an empty kind; the scalar
// construction keeps its schema.
func TestWidenedConstructionHasNoJSONSchema(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		known        bool
	}{
		{"widened", widenedMembers + "const boss: Owner = { name: 'boss' }; const item: Item = { label: 'l' }; const big = { name: 'n', items: [item], origin: boss } as Big;", false},
		{"scalar", "interface S { a: number; b: string } const s = { a: 1 } as S;", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, cast := widenedProbe(t, probe.source, ".ts")
			if !l.constructionCast(cast) {
				t.Fatal("the probe's cast is no construction, so its schema answer proves nothing")
			}
			schema, known := l.constructionJSON(cast.Parent.Name())
			if known != probe.known {
				t.Fatalf("constructionJSON knows %v, want %v", known, probe.known)
			}
			if known {
				for _, field := range schema.Fields {
					if field.Schema.Kind == "" {
						t.Fatalf("slot %d has no kind", field.Slot)
					}
				}
			}
		})
	}
}

// The widened construction lowers, its literal laying each field it doesn't write absent: a required one
// Uninitialized too, so its read stops, and an optional one really absent, so its read is undefined. Mutant: an
// optional unwritten field laid Uninitialized too (unwrittenField); note and size go red here.
func TestWidenedConstructionLaysUnwrittenFields(t *testing.T) {
	t.Parallel()
	source := widenedMembers + "function make(name: string, x: Item, origin: Owner): Big { return { name, items: [x], origin } as Big; }\nconst big = make('n', { label: 'l' }, { name: 'o' });\nconsole.log(big.name + ' ' + big.items[0].label + ' ' + String(big.note) + ' ' + String(big.size === undefined));\n"
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	var laid []ir.Field
	find := func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok {
			for _, field := range literal.Fields {
				if field.Name == "parent" {
					laid = literal.Fields
				}
			}
		}
		return true
	}
	walk(program.Main, find)
	for _, function := range program.Functions {
		walk(function.Body, find)
	}
	if laid == nil {
		t.Fatal("no literal lays parent")
	}
	want := map[string][2]bool{"name": {false, false}, "items": {false, false}, "origin": {false, false}, "count": {true, true}, "parent": {true, true}, "note": {true, false}, "size": {true, false}}
	if len(laid) != len(want) {
		t.Fatalf("the literal has %d fields, want %d", len(laid), len(want))
	}
	for _, field := range laid {
		if expected, named := want[field.Name]; !named || field.Absent != expected[0] || field.Uninitialized != expected[1] {
			t.Errorf("%s: Absent %v Uninitialized %v, want %v", field.Name, field.Absent, field.Uninitialized, expected)
		}
	}
}

// A written field whose type isn't assignable to its slot keeps the sideways stop, in castStop's words.
func TestWidenedConstructionWiderFieldStaysStopped(t *testing.T) {
	t.Parallel()
	source := widenedMembers + "function f(thing: Item | Owner, boss: Owner): Big { return { name: 'n', items: [thing], origin: boss } as Big; }\nconsole.log(f({ label: 'l' }, { name: 'b' }).name);\n"
	_, err := lowerSourceAs(t, source, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "which Big doesn't extend") {
		t.Fatalf("want castSideways' NotYet, got %v", err)
	}
}
