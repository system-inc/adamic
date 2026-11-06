package lower

import (
	"sort"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// spreadTree is a plain-typed global and a function whose spread of its second parameter may meet a
// getter, which each probe below hands it. Tree's fields are readonly, as a getter without a setter
// is, so a class instance with one is assignable to it without the readonly-seen-as-writable stop.
const spreadTree = "type Tree = { readonly value: number; readonly label: string };\nlet tree: Tree = { value: 1, label: `tree${String(1)}` };\nfunction describe(node: Tree, extra: Tree): string { const copy = { ...extra }; return `${String(node.value + copy.value)} ${node.label}`; }\n"

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
// records none, so reuse and purity judge its spreads as before. Mutant
// (internal/oracle/testdata/iterable_views/mutants.md): spreadGetters' call in finishAccessors made
// nil, and the three getter rows find none.
func TestSpreadRecordsTheGettersItMayRun(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		getter       bool
	}{
		{"an object literal's getter", spreadTree + "const source: Tree = { get value(): number { return tree.value + 10; }, label: `source${String(2)}` };\nconsole.log(describe(tree, source));", true},
		{"a class's getter", spreadTree + "class Sized { label = `sized${String(3)}`; get value(): number { return tree.value + 20; } }\nconsole.log(describe(tree, new Sized()));", true},
		{"a getter inherited from a base", spreadTree + "class Base { get value(): number { return tree.value + 30; } }\nclass Derived extends Base { label = `derived${String(4)}`; }\nconsole.log(describe(tree, new Derived()));", true},
		{"a getter of a class no Tree can be", spreadTree + "class Counter { count = 0; get next(): number { this.count++; return this.count; } }\nconst counter = new Counter();\nconsole.log(String(counter.next));\nconsole.log(describe(tree, { value: 10, label: `source${String(2)}` }));", false},
		{"no getter anywhere", spreadTree + "const source: Tree = { value: 10, label: `source${String(2)}` };\nconsole.log(describe(tree, source));", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
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
