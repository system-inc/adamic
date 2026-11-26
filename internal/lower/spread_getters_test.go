package lower

import (
	"errors"
	"sort"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// spreadTree is a plain-typed global and a function whose spread of its second parameter may meet a
// getter, which each probe below hands it. Tree's fields are readonly, as a getter without a setter
// is, so a class instance with one is assignable to it without the readonly-seen-as-writable stop.
// describe reads the copy's label, never its value: a class's getter isn't in a spread's copy, and a
// read of its name from the copy stops (spreadCopyRead, #7mjgxkc).
const spreadTree = "type Tree = { readonly value: number; readonly label: string };\nlet tree: Tree = { value: 1, label: `tree${String(1)}` };\nfunction describe(node: Tree, extra: Tree): string { const copy = { ...extra }; return `${String(node.value)} ${node.label} ${copy.label}`; }\n"

// narrowTree is spreadTree with describe's spread seen through { readonly label: string }, a view an
// object literal with a getter beside label reaches by width and fails by excess properties (#z6h9s7k).
const narrowTree = "type Tree = { readonly value: number; readonly label: string };\nlet tree: Tree = { value: 1, label: `tree${String(1)}` };\nfunction describe(node: Tree, extra: { readonly label: string }): string { const copy = { ...extra }; return `${String(node.value)} ${node.label} ${copy.label}`; }\n"

// spreadGettersOf is, for every object literal with a spread, the names of the getters lowering says
// it may run, sorted.
func spreadGettersOf(program *ir.Program) [][]string {
	var found [][]string
	visit := func(node any) bool {
		if literal, isLiteral := node.(ir.ObjectLiteral); isLiteral && literal.Spread != nil {
			names := []string{}
			for _, getter := range literal.SpreadGetters {
				names = append(names, program.Functions[getter].Name)
			}
			sort.Strings(names)
			found = append(found, names)
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return found
}

// A spread runs the getters its source's class chain owns (runtime/object.c's
// adamic_object_copy_checked), so lowering records the set it may run on the literal, read off the
// accessor registry (Memory's and Compiler's ruling on #qk8rztp S2c-2b): an object literal's own
// getter, a class's, and one inherited from a base, each a class whose instances are assignable to
// the parameter's plain declared type, which every one of them satisfies, so no test of that type
// could rule them out. A getter of a class no Tree can be isn't taken, and a program with no getter
// records none, so reuse and purity judge its spreads as before. A literal seen through a narrower view
// is related by its regular type, never its fresh one (#z6h9s7k; spread_getters.mutants.md). The class's and the
// inherited rows lower as .ts, where that record is the road: in .a a spread whose type declares as a field a member
// a class reaching it holds on its prototype is refused (spread_prototype_members.go, #vjvcnnj), as Tree's value is,
// so each also lowers as .a and is refused with those words. Mutants: spreadGetters' call in finishAccessors made
// nil (internal/oracle/testdata/iterable_views/mutants.md), and the three getter rows find none; the .a refusal
// answering nil (internal/lower/spread_prototype_members.mutants.md), and the two .a refusals go red.
func TestSpreadRecordsTheGettersItMayRun(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, extension, source string
		getter                  bool
		refusedInA, fixInA      string
	}{
		{"an object literal's getter", ".a", spreadTree + "const source: Tree = { get value(): number { return tree.value + 10; }, label: `source${String(2)}` };\nconsole.log(describe(tree, source));", true, "", ""},
		{"a class's getter", ".ts", spreadTree + "class Sized { label = `sized${String(3)}`; get value(): number { return tree.value + 20; } }\nconsole.log(describe(tree, new Sized()));", true, "a spread of extra, whose copy's own type declares value a field, while class Sized, which may reach it, holds value as a getter on its prototype: a spread copies own properties only, so the copy has no value though its type says it does", "copy the fields you mean explicitly, or declare value a field of Sized"},
		{"a getter inherited from a base", ".ts", spreadTree + "class Base { get value(): number { return tree.value + 30; } }\nclass Derived extends Base { label = `derived${String(4)}`; }\nconsole.log(describe(tree, new Derived()));", true, "a spread of extra, whose copy's own type declares value a field, while class Derived, which may reach it, holds value as a getter on its prototype, inherited from Base: a spread copies own properties only, so the copy has no value though its type says it does", "copy the fields you mean explicitly, or declare value a field of Base"},
		// #z6h9s7k: the literal's own type is fresh, so it reaches a narrower view only by relatedShape.
		{"an object literal's getter through a narrower view", ".a", narrowTree + "const source: Tree = { get value(): number { return tree.value + 10; }, label: `source${String(2)}` };\nconsole.log(describe(tree, source));", true, "", ""},
		{"an unannotated object literal's getter through a narrower view", ".a", narrowTree + "const source = { get value(): number { return tree.value + 10; }, label: `source${String(2)}` };\nconsole.log(describe(tree, source));", true, "", ""},
		{"a getter of a class no Tree can be", ".a", spreadTree + "class Counter { count = 0; get next(): number { this.count++; return this.count; } }\nconst counter = new Counter();\nconsole.log(String(counter.next));\nconsole.log(describe(tree, { value: 10, label: `source${String(2)}` }));", false, "", ""},
		{"no getter anywhere", ".a", spreadTree + "const source: Tree = { value: 10, label: `source${String(2)}` };\nconsole.log(describe(tree, source));", false, "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if probe.refusedInA != "" {
				_, err := lowerSourceAs(t, probe.source, ".a")
				var refused *Refused
				if !errors.As(err, &refused) || refused.What != probe.refusedInA || refused.Fix != probe.fixInA {
					t.Fatalf("as .a, want the spread refused with %q; %q, got %v", probe.refusedInA, probe.fixInA, err)
				}
			}
			program, err := lowerSourceAs(t, probe.source, probe.extension)
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			spreads := spreadGettersOf(program)
			if len(spreads) != 1 {
				t.Fatalf("%d spreads, want describe's one", len(spreads))
			}
			if got := len(spreads[0]) != 0; got != probe.getter {
				t.Fatalf("the spread may run getters %v, want some: %t", spreads[0], probe.getter)
			}
		})
	}
}
