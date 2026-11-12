package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A checked predicate's false side is checked only where a read observes it (#39wtqdj, Adamic's ruling of
// Nov 11, with Language's condition): a read of the narrowed reference whose flow chain reaches the call's
// false condition with no assignment between, whatever later narrowings stand between them, decided by the
// flow graph and never by comparing types. Each row is the first some call's directions, tsc's some, whose
// false side lies for an empty array. The true side is checked in every row. Mutants
// (predicate_false_read.mutants.md): the condition-only arm restored checks the four unread rows; types
// deciding alone drops "capture of a let assigned later"; a merge counted as a false read checks "read
// after the sides merge"; holders read check "replaced before the read", whose xs = ["a"] the checker
// matches against xs; the name after a dot read checks "a property named as the reference"; labels read
// check "a label named as the reference".
func TestCheckedPredicateFalseReadDirections(t *testing.T) {
	t.Parallel()
	prefix := `function some<T>(array: readonly T[] | undefined): array is readonly T[];
function some<T>(array: readonly T[] | undefined, predicate: (value: T) => boolean): boolean;
function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean { if (array === undefined) return false; if (predicate === undefined) return array.length > 0; for (const value of array) { if (predicate(value)) return true; } return false; }
`
	for _, probe := range []struct {
		name, body string
		checked    bool
	}{
		{"empty branches", `function use(xs: readonly string[]): void { if (some(xs)) {} else {} }`, false},
		{"semver's toString", `class Version { readonly build: readonly string[] = []; text(): string { let result = "1.2.3"; if (some(this.build)) result += "+" + this.build.join("."); return result; } }`, false},
		{"read after the sides merge", `function use(xs: readonly string[] | undefined): void { if (some(xs)) console.log("some"); console.log(typeof xs); }`, false},
		{"replaced before the read", `function use(xs: readonly string[] | undefined): void { if (!some(xs)) { xs = ["a"]; console.log(String(xs.length)); } }`, false},
		{"a property named as the reference", `function use(xs: readonly string[] | undefined, holder: { readonly xs: number }): void { if (!some(xs)) console.log(String(holder.xs)); }`, false},
		{"a label named as the reference", `function use(items: readonly string[] | undefined): void { if (!some(items)) { items: for (;;) { break items; } } }`, false},
		{"false branch", `function use(xs: readonly string[] | undefined): void { if (!some(xs)) console.log(typeof xs); }`, true},
		{"else", `function use(xs: readonly string[] | undefined): void { if (some(xs)) {} else console.log(typeof xs); }`, true},
		{"after an early exit", `function use(xs: readonly string[] | undefined): void { if (some(xs)) { return; } console.log(typeof xs); }`, true},
		{"capture", `function use(xs: readonly string[] | undefined): void { if (!some(xs)) { const read = (): string => typeof xs; console.log(read()); } }`, true},
		{"renarrowed", `function use(xs: readonly string[] | undefined): void { if (!some(xs)) { if (xs !== undefined) { const kept: readonly string[] = xs; console.log(String(kept.length)); } } }`, true},
		// The checker doesn't narrow a let inside a closure when the let is assigned after the closure is made,
		// so the captured read's type is the declared type with or without the false side: only the flow graph
		// carries the false answer to it.
		{"capture of a let assigned later", `function use(input: readonly string[] | undefined): void { let xs = input; if (!some(xs)) { const read = (): string => typeof xs; console.log(read()); } xs = undefined; }`, true},
		{"getter", `class Holder { get items(): readonly string[] | undefined { return undefined; } } function use(holder: Holder): void { if (!some(holder.items)) console.log(typeof holder.items); }`, true},
		{"spread", `function use(xs: readonly string[]): void { if (!some(xs)) console.log(String([...xs].length)); }`, true},
		{"the same predicate again", `function use(xs: readonly string[] | undefined): void { if (!some(xs)) console.log(String(some(xs))); }`, true},
		{"conditional", `function use(xs: readonly string[] | undefined): string { return some(xs) ? "some" : typeof xs; }`, true},
		{"|| operand", `function use(xs: readonly string[] | undefined): boolean { return some(xs) || typeof xs === "object"; }`, true},
		{"&& operand", `function use(xs: readonly string[] | undefined): boolean { return !some(xs) && typeof xs === "object"; }`, true},
		{"const alias", `function use(xs: readonly string[] | undefined): void { const present = some(xs); if (!present) console.log(typeof xs); }`, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "uses.ts")
			if err := os.WriteFile(path, []byte(prefix+probe.body), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			var call *ast.CallExpression
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if call == nil && node.Kind == ast.KindCallExpression && node.Expression().Kind == ast.KindIdentifier && node.Expression().Text() == "some" {
					call = node.AsCallExpression()
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			if call == nil {
				t.Fatal("missing probe call")
			}
			claim := predicateOfSignature(l.checker, l.checker.GetResolvedSignature(call.AsNode()))
			if claim == nil {
				t.Fatal("some's call claims no predicate")
			}
			directions := l.checkedPredicateDirections(call, claim)
			if directions&predicateTrue == 0 {
				t.Fatalf("directions %d: the true side isn't checked", directions)
			}
			if got := directions&predicateFalse != 0; got != probe.checked {
				t.Fatalf("false side checked %v, want %v", got, probe.checked)
			}
		})
	}
}
