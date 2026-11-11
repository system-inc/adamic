package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// S2c-4-1 (#qk8rztp): a generator's direct call stored in a const of a function and read once, by a
// for-of over the local or as the one-read Iterable<T> argument of a one-body callee, lowers, the read
// a move of the frame out of the local (generators_stored.go). The oracle holds what the fixtures run
// to (internal/oracle/testdata/generators/stored_*.a); this holds that they and the inline shapes
// lower, and that every other use stops in Lower's words.

// storedGeneratorFixtures are the oracle's S2c-4-1 programs: tsc's generateJsxChildren shape (pin 1),
// a for-of over a stored generator (pin 2), a throw, a return and a break between the declaration and
// the read (pin 3), the throw's twin with a finally in the body (pin 4), a read on a branch that may
// not run (pin 5), and a later argument's throw after the read (pin 6).
var storedGeneratorFixtures = []string{"stored_jsx_children.a", "stored_for_of.a", "stored_exits.a", "stored_exits_finally.a", "stored_untaken.a", "stored_later_throw.a"}

// storedGeneratorNumbers is a module generator of numbers whose TReturn is undefined, so a road this
// head stops reaches lowering's stop rather than the void rules', and a module callee whose one use of
// its parameter is one for-of over it.
const storedGeneratorNumbers = "function* numbers(count: number): Generator<number, undefined, undefined> { for (let index = 0; index < count; index++) { yield index; } }\n" +
	"function total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\n"

// storedGeneratorHost wraps body in a function beside storedGeneratorNumbers.
func storedGeneratorHost(body string) string {
	return storedGeneratorNumbers + "function host(flag: boolean): void {\n" + body + "\n}\nhost(String(1).length > 0);\n"
}

// storedGeneratorWords are the words every S2c-4-1 stop says (Lower's, on #qk8rztp S2c-4).
const storedGeneratorWords = "a generator stored in a local, used other than once by a consumer that steps it (#qk8rztp)"

// Every fixture lowers, and each lowered program moves a stored generator out of its local into a
// consumer, with checkGeneratorRoads holding. The inline rows are the shapes the fixtures don't carry:
// a read in parentheses, a const declared in a block and read in a deeper one, declared and read in
// one loop body, a nested generator into a nested callee, one written in place, and the parameter
// forms: `Iterable<T> | undefined`, and IterableIterator<T> stepped by hand (S2c-2b's view).
func TestStoredGeneratorsLower(t *testing.T) {
	t.Parallel()
	for _, name := range storedGeneratorFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loaded, err := load.Load([]string{"../oracle/testdata/generators/" + name})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			program, err := Lower(context.Background(), loaded)
			if err != nil {
				t.Fatal(err)
			}
			if moved := storedGeneratorMoves(program); len(moved) == 0 {
				t.Error("no stored generator moved into a consumer")
			}
			if err := checkGeneratorRoads(program); err != nil {
				t.Errorf("the lowered program's roads: %v", err)
			}
		})
	}
	for _, probe := range []struct{ name, source string }{
		{"handed to a module callee", storedGeneratorHost("const stored = numbers(3);\nconsole.log(String(total(stored)));")},
		{"a for-of over it", storedGeneratorHost("const stored = numbers(3);\nfor (const value of stored) { console.log(String(value)); }")},
		{"a read in parentheses", storedGeneratorHost("const stored = numbers(2);\nfor (const value of (stored)) { console.log(String(value)); }")},
		{"declared in a block, read in a deeper one", storedGeneratorHost("if (flag) {\nconst stored = numbers(2);\nif (flag) {\nconsole.log(String(total(stored)));\n}\n}")},
		{"declared and read in one loop body", storedGeneratorHost("for (let pass = 1; pass < 3; pass++) {\nconst stored = numbers(pass);\nconsole.log(String(total(stored)));\n}")},
		{"a nested generator into a nested callee", "function host(): void {\nfunction* inner(): Generator<number, undefined, undefined> { yield 1; yield 2; }\nfunction sum(items: Iterable<number>): number { let all = 0; for (const value of items) { all += value; } return all; }\nconst stored = inner();\nconsole.log(String(sum(stored)));\n}\nhost();\n"},
		{"written in place", storedGeneratorHost("const stored = (function* (): Generator<number, undefined, undefined> { yield 4; })();\nconsole.log(String(total(stored)));")},
		{"an Iterable<T> | undefined parameter", storedGeneratorNumbers + "function first(items: Iterable<number> | undefined): number { if (items !== undefined) { for (const value of items) { return value; } } return -1; }\nfunction host(): void {\nconst stored = numbers(2);\nconsole.log(String(first(stored)));\n}\nhost();\n"},
		{"an IterableIterator<T> parameter stepped by hand", handFirst + storedGeneratorNumbers + "function host(): void {\nconst stored = numbers(2);\nconsole.log(String(first(stored)));\n}\nhost();\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if moved := storedGeneratorMoves(program); len(moved) != 1 {
				t.Errorf("%d stored generators moved into a consumer, want 1", len(moved))
			}
		})
	}
}

// storedGeneratorMoves is every stored generator's local program moves into a consumer, by how many
// times: a for-of's own local declared from its read, or IterableOf over it.
func storedGeneratorMoves(program *ir.Program) map[int]int {
	moved := map[int]int{}
	inspect := func(node any) bool {
		var value ir.Expression
		switch node := node.(type) {
		case ir.Declare:
			value = node.Value
		case ir.IterableOf:
			value = node.Value
		}
		if read, isMove := storedGeneratorMove(value); isMove {
			moved[read.Local]++
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return moved
}

// Every use of a stored generator but one read by a consumer, once per run of its declaration, stops in
// Lower's words (Adamic's conditions 1 and 2, Compiler's must, Language's attack). Each row's words are
// a part of its stop, so a row that stops for another reason goes red too. What Node prints for each
// is beside it, where Node runs it. Mutants (internal/oracle/testdata/generators/mutants.md, S2c-4-1):
// two reads admitted, and the two reads row lowers; the loop containment dropped, and Compiler's two
// rows lower.
func TestStoredGeneratorRoadsAreNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		// Node: 0, 1 and 2 on the first pass, and nothing on the second, which finds it done.
		{"Compiler's two passes over it", storedGeneratorHost("const stored = numbers(3);\nfor (let pass = 0; pass < 2; pass++) { for (const value of stored) { console.log(String(value)); } }")},
		// Node: 3, then 0, the second call handed a generator already done.
		{"an argument in a loop it isn't declared in", storedGeneratorHost("const stored = numbers(3);\nfor (let pass = 0; pass < 2; pass++) { console.log(String(total(stored))); }")},
		// Node: 0, 1 and 2, and nothing from the second for-of.
		{"two reads", storedGeneratorHost("const stored = numbers(3);\nfor (const value of stored) { console.log(String(value)); }\nfor (const again of stored) { console.log(String(again)); }")},
		// Node: object, then 0 and 1.
		{"typeof", storedGeneratorHost("const stored = numbers(2);\nconsole.log(typeof stored);\nfor (const value of stored) { console.log(String(value)); }")},
		// Node: false.
		{"===", storedGeneratorHost("const stored = numbers(2);\nconst other = numbers(2);\nconsole.log(String(stored === other));")},
		// Node: 0 and 1.
		{"a let", storedGeneratorHost("let stored = numbers(2);\nfor (const value of stored) { console.log(String(value)); }")},
		// Node: 0 and 1.
		{"a type written for it", storedGeneratorHost("const stored: Iterable<number> = numbers(2);\nfor (const value of stored) { console.log(String(value)); }")},
		// Node: 1.
		{"a capture, called", storedGeneratorHost("const stored = numbers(2);\nconst run = (): number => total(stored);\nconsole.log(String(run()));")},
		// Node: kept.
		{"a capture, never called", storedGeneratorHost("const stored = numbers(2);\nconst keep = (): number => total(stored);\nconsole.log('kept');")},
		// Node: 1, the hoisted function reading it after the declaration ran.
		{"a function beside it", storedGeneratorHost("function later(): number { return total(stored); }\nconst stored = numbers(2);\nconsole.log(String(later()));")},
		// Node: 2.
		{"a field", storedGeneratorHost("const stored = numbers(2);\nconst holder = { steps: stored, size: 2 };\nconsole.log(String(holder.size));")},
		// Node: prints nothing; the generator made is returned and dropped.
		{"a return", storedGeneratorNumbers + "function make(): Generator<number, undefined, undefined> {\nconst stored = numbers(2);\nreturn stored;\n}\nmake();\n"},
		// Node: 0.
		{"destructuring", storedGeneratorHost("const stored = numbers(2);\nconst [first] = stored;\nconsole.log(String(first));")},
		// Node: 0,1.
		{"a spread", storedGeneratorHost("const stored = numbers(2);\nconsole.log([...stored].join(','));")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, storedGeneratorWords) {
				t.Fatalf("got %v, want a NotYet saying %q", err, storedGeneratorWords)
			}
		})
	}
	// Outside a function, or in one that suspends, a stored generator isn't this head's: its call stops
	// where it's lowered, as before S2c-4-1. Node: 0 and 1, each.
	const before = "whose generator is used other than by a for-of over its call or a call as a statement"
	for _, probe := range []struct{ name, source string }{
		{"at a module's top level", storedGeneratorNumbers + "const stored = numbers(2);\nfor (const value of stored) { console.log(String(value)); }\n"},
		{"in an async function", storedGeneratorNumbers + "async function host(): Promise<void> {\nconst stored = numbers(2);\nfor (const value of stored) { console.log(String(value)); }\n}\nawait host();\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, before) {
				t.Fatalf("got %v, want a NotYet saying %q", err, before)
			}
		})
	}
}

// checkGeneratorRoads' backstop under storedGeneratorRead: a stored generator's local is read only by
// its moves into a consumer. The control steps the local of stored_for_of.a's plain() directly, a read
// that is no move, put into the lowered program by hand, which the backstop must see. Mutant
// (internal/oracle/testdata/generators/mutants.md, S2c-4-1): its moves check taken out, and the
// control passes.
func TestStoredGeneratorLocalIsReadOnlyByItsMoves(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{"../oracle/testdata/generators/stored_for_of.a"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkGeneratorRoads(program); err != nil {
		t.Fatalf("the lowered program: %v", err)
	}
	stored := -1
	for local := range storedGeneratorMoves(program) {
		stored = local
	}
	if stored < 0 {
		t.Fatal("no stored generator moved into a consumer")
	}
	program.Main = append(program.Main, ir.Evaluate{Value: ir.GeneratorDone{Generator: ir.Read{Local: stored, Of: ir.Generator}}})
	if err := checkGeneratorRoads(program); err == nil || !strings.Contains(err.Error(), "moves into a consumer") {
		t.Fatalf("a stored generator's local stepped directly: got %v, want the backstop's compiler bug", err)
	}
}
