package lower

import (
	"errors"
	"strings"
	"testing"
)

// roadPForms are the two presence tests road P extends (#p2n17c4, rulings.md:117 applied to the right of
// an &&): a class's required method, which method presence folds (method_presence.go), and an optional
// member of a program's interface, which shape 1 reads off its slot (optional_member_presence.go). Each
// declares the receiver h, with a method m, and a function probe(flag, other) that the rows fill in.
var roadPForms = []struct{ name, declared string }{
	{"a class's method", "class H {\n\tm(): number {\n\t\treturn 1;\n\t}\n}\nconst h = new H();\n"},
	{"an optional member", "interface H {\n\tm?(): number;\n}\nconst h: H = { m() { return 1; } };\n"},
}

// roadPProgram is a form's declaration, then probe with body, then calls of probe with each flag.
func roadPProgram(declared, body string) string {
	return declared + "function probe(flag: boolean, other: boolean): string {\n" + body + "\n}\n" +
		"console.log(probe(true, true));\nconsole.log(probe(false, true));\nconsole.log(probe(true, false));\n"
}

// Road P: the right of an && is a presence read where the && is tested for truthiness alone, recursively
// (truthinessOnly): an if, a while, the test of ?:, the operand of !, and either side of an && that is
// itself tested so. Each row lowers in both forms. The nested rows are the recursion's: their && holding
// the test is itself an operand of an && in a condition. The oracle holds what these run to
// (internal/oracle/road_p_test.go). Mutant (road_p.mutants.md): truthinessOnly's recursion taken back to
// one level, and the nested rows stop.
func TestRoadPRightOfAndLowers(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, body string }{
		{"an if", "\tif (flag && h.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';"},
		{"!", "\treturn String(!(flag && h.m));"},
		{"the test of ?:", "\treturn flag && h.m ? 'has' : 'none';"},
		{"a while", "\twhile (flag && h.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';"},
		{"a chain of &&", "\tif (flag && other && h.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';"},
		{"nested on the right", "\tif (flag && (other && h.m)) {\n\t\treturn 'has';\n\t}\n\treturn 'none';"},
		{"nested on the left", "\tif ((flag && h.m) && other) {\n\t\treturn 'has';\n\t}\n\treturn 'none';"},
	} {
		for _, form := range roadPForms {
			for _, extension := range []string{".a", ".ts"} {
				t.Run(row.name+", "+form.name+extension, func(t *testing.T) {
					t.Parallel()
					if _, err := lowerSourceAs(t, roadPProgram(form.declared, row.body), extension); err != nil {
						t.Fatalf("want it to lower, got %v", err)
					}
				})
			}
		}
	}
}

// Where the &&'s value is used, its right keeps the method value, and stays refused as a method read as a
// value, at the read: an initializer, a return, an argument, typeof of the && (which names the value's
// type), the left of an && whose value is used (its falsy left is that value), and the right of ||. Each
// row is refused in both forms. Mutant (road_p.mutants.md): truthinessOnly answering true for any &&
// operand, and these rows lower.
func TestRoadPRightOfAndKeepingTheValueIsRefused(t *testing.T) {
	t.Parallel()
	const refusal = "Adamic 0.1 refuses a method read as a value (m would lose its object, and this with it)"
	for _, row := range []struct{ name, body string }{
		{"an initializer", "\tconst held = flag && h.m;\n\treturn String(held ? 1 : 0);"},
		{"an argument", "\treturn String(keep(flag && h.m));\n}\nfunction keep(v: false | (() => number) | undefined): boolean {\n\treturn v !== undefined && v !== false;"},
		{"a return", "\treturn String(pick(flag) ? 1 : 0);\n}\nfunction pick(flag: boolean): false | (() => number) | undefined {\n\treturn flag && h.m;"},
		{"typeof the &&", "\treturn typeof (flag && h.m);"},
		{"the left of an && whose value is used", "\tconst held = (flag && h.m) && other;\n\treturn String(held);"},
		{"the right of ||", "\tif (flag || h.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';"},
	} {
		for _, form := range roadPForms {
			t.Run(row.name+", "+form.name, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, roadPProgram(form.declared, row.body), ".ts")
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.Error(), refusal) {
					t.Errorf("got %v, want a refusal containing %q", err, refusal)
				}
			})
		}
	}
}
