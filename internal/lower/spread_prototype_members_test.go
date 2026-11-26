package lower

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// spreadContextHeldPrelude declares what each row's spread copies and where it may be held.
const spreadContextHeldPrelude = "interface L { readonly label: string; }\nconst built: L = { label: \"t\" };\nlet held: L = { label: \"l\" };\n"

// spreadContextHeld answers whether a spread's literal is held as its contextual type, and whether as its own type
// too, by the literal's place alone (Diagnostics' re-read of #vjvcnnj head 2). An assignment's target always holds
// the copy, and where the assignment's value is read the expression is typed as the literal's own type, so the copy
// is held both ways: (true, true). Discarded, as a statement or a for loop's initializer, the target is the one
// holder. A function's return is held as its context only when the return is annotated or the function itself is
// held: under `satisfies` the function keeps its own type, whose return is the literal's own. Diagnostics'
// assign_value probes stop at earlier stage-0 stops today, so these rows hold the rule while those stops hide it.
// Mutants (spread_prototype_members.mutants.md): alsoOwn always false reds the read assignment's row, and
// spreadReturnHeld answering held for any function reds the two satisfies rows.
func TestSpreadContextHeldAnswersByThePlace(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source  string
		held, alsoOwn bool
	}{
		{"an assignment statement", "held = { ...built };\n", true, false},
		{"an assignment whose value is read", "const value = (held = { ...built });\nconsole.log(value.label);\n", true, true},
		{"a for loop's initializer", "for (held = { ...built }; held.label === \"x\"; ) {}\n", true, false},
		{"an arrow under satisfies", "const f = (() => ({ ...built })) satisfies () => L;\n", false, false},
		{"a return under satisfies", "const f = (() => {\n\treturn { ...built };\n}) satisfies () => L;\n", false, false},
		{"an arrow held by an annotated declaration", "const f: () => L = () => ({ ...built });\n", true, false},
		{"a return of an annotated function", "function f(): L {\n\treturn { ...built };\n}\n", true, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			withAccessChecker(t, spreadContextHeldPrelude+probe.source, func(_ *lowering, statements []*ast.Node) {
				var literal *ast.Node
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if node.Kind == ast.KindSpreadAssignment && literal == nil {
						literal = node.Parent
					}
					return node.ForEachChild(visit)
				}
				for _, statement := range statements {
					statement.ForEachChild(visit)
				}
				if literal == nil {
					t.Fatal("no spread in the source")
				}
				if held, alsoOwn := spreadContextHeld(literal); held != probe.held || alsoOwn != probe.alsoOwn {
					t.Fatalf("spreadContextHeld = (%v, %v), want (%v, %v)", held, alsoOwn, probe.held, probe.alsoOwn)
				}
			})
		})
	}
}
