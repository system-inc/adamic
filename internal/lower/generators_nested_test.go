package lower

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// S3a's nested generator lowers as a nested function does (Lower's term 1 on #qk8rztp S3a): it is
// bound with its siblings and hoisted, so a call written above its declaration lowers, and its call
// makes its frame through a carrier, direct to its function: from the parent, the parent's binding
// of it; from a sibling, the sibling's own closure; from an arrow in the parent, one made for the
// call. Each consumer takes it as a module generator's: a for-of over its call, the call as a
// statement, and a one-read Iterable<T> parameter's argument. Mutants: nestedGeneratorCall's
// ClosureSelf arm made a MakeClosure, and the sibling row goes red; generatorCallee's nested lookup
// dropped, and every consumer stops as a value road instead.
func TestNestedGeneratorsLower(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function consume(items: Iterable<string>): string {
	let joined = '';
	for (const item of items) {
		joined += item;
	}
	return joined;
}
function host(seed: string): string {
	const early = consume(parts(1));
	function* parts(count: number): Generator<string, void, undefined> {
		for (let index = 0; index < count; index++) {
			yield seed + String(index);
		}
	}
	function sibling(): string {
		let joined = '';
		for (const part of parts(2)) {
			joined += part;
		}
		return joined;
	}
	const arrow = (): string => {
		let joined = '';
		for (const part of parts(3)) {
			joined += part;
		}
		return joined;
	};
	parts(4);
	return early + sibling() + arrow();
}
console.log(host('s'));
`)
	if err != nil {
		t.Fatal(err)
	}
	generator := -1
	for index, function := range program.Functions {
		if function.Name == "parts" {
			generator = index
		}
	}
	if generator < 0 {
		t.Fatal("no function parts")
	}
	parts := program.Functions[generator]
	if !parts.Generator || !parts.Closure || parts.NestedParent == 0 || parts.Returns != ir.Generator || parts.Yields != ir.String {
		t.Fatalf("parts: Generator %t, Closure %t, NestedParent %d, Returns %d, Yields %d; want a nested closure making a Generator of strings", parts.Generator, parts.Closure, parts.NestedParent, parts.Returns, parts.Yields)
	}
	carriers := map[string][]string{}
	record := func(owner string) func(node any) bool {
		return func(node any) bool {
			made, isMade := node.(ir.CallClosure)
			if !isMade || made.Returns != ir.Generator {
				return true
			}
			// Its target is read through ClosureTargets, as every reader of a call's targets must
			// (internal/ir's call_targets_guard_test.go); its carrier is the one closure-shaped value
			// among its operands, since every argument here is a number.
			if targets := program.ClosureTargets(made); targets.Unknown || len(targets.Functions) != 1 || targets.Functions[0] != generator {
				t.Errorf("%s: a Generator call to %v, want parts (%d) alone", owner, targets, generator)
			}
			found := []string{}
			walk(made, func(node any) bool {
				switch carrier := node.(type) {
				case ir.Read:
					if carrier.Of == ir.Closure {
						found = append(found, "binding")
					}
				case ir.ClosureSelf:
					found = append(found, "self")
				case ir.MakeClosure:
					if carrier.Function != generator {
						t.Errorf("%s: a carrier made for function %d, want parts", owner, carrier.Function)
					}
					found = append(found, "made")
				}
				return true
			})
			if len(found) != 1 {
				t.Errorf("%s: carriers %v in one call, want one", owner, found)
			}
			carriers[owner] = append(carriers[owner], found...)
			return true
		}
	}
	for _, function := range program.Functions {
		walk(function.Body, record(function.Name))
	}
	arrow := ""
	for owner := range carriers {
		if owner != "host" && owner != "sibling" {
			arrow = owner
		}
	}
	if strings.Join(carriers["host"], ",") != "binding,binding" || strings.Join(carriers["sibling"], ",") != "self" || strings.Join(carriers[arrow], ",") != "made" || len(carriers) != 3 {
		t.Errorf("carriers %v, want the parent's binding twice (the Iterable argument above the declaration, and the dropped call), the sibling's self, and one made in the arrow", carriers)
	}
	if err := checkGeneratorRoads(program); err != nil {
		t.Errorf("the lowered program's roads: %v", err)
	}
	// A sibling's write to a captured local is the sibling's own, which the generator's next step
	// reads; and a plain nested function* infers a void TReturn, which a one-read Iterable<T>
	// parameter never reads, so the void rules let its call through as they do a module generator's
	// (generatorStepped reads loweredGenerator, which says nested ones too). Mutant: nestedGenerator
	// taken out of loweredGenerator; the void row reads Refused, and the sibling row stops at the
	// declaration.
	for name, source := range map[string]string{
		"a sibling writes a captured local":                      "function host(): void {\nlet count = 0;\nfunction bump(): void { count++; }\nfunction* counted(): Generator<number, void, undefined> { bump(); yield count; bump(); yield count; }\nfor (const value of counted()) console.log(String(value));\n}\nhost();\n",
		"a plain nested function* into an Iterable<T> parameter": "function consume(items: Iterable<number>): number { let total = 0; for (const item of items) { total += item; } return total; }\nfunction host(): void {\nfunction* numbers() { yield 1; yield 2; }\nconsole.log(String(consume(numbers())));\n}\nhost();\n",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// What S3a doesn't lower about a nested generator stops as not yet built, in its own words (Lower's
// and Memory's terms on #qk8rztp S3a), and a bare call whose completion value would be read stops
// as a module generator's does. Each row's words are a part of its stop, so a row that stops for
// another reason goes red too. Mutants: the captured-write stop dropped, and the write rows lower;
// the collect stop dropped, and the four collect rows lower into a helper with no carrier, whose
// call can't be emitted; the suspending parent's stop dropped, and the async and generator rows
// stop on NormalizeAsync's and checkGeneratorBody's words instead.
func TestNestedGeneratorRoadsAreNotYet(t *testing.T) {
	t.Parallel()
	host := func(body string) string { return "function host(): void {\n" + body + "\n}\nhost();\n" }
	const numbers = "function* numbers(): Generator<number, undefined, undefined> { yield 1; yield 2; }\n"
	const counted = "let count = 0;\nfunction* counted(): Generator<number, undefined, undefined> { BODY yield count; }\nfor (const value of counted()) console.log(String(value));"
	write := func(statement string) string { return host(strings.ReplaceAll(counted, "BODY", statement)) }
	const written = "where 'count', a captured local, is written inside a nested generator"
	const collected = "declared inside a function and consumed by spread, Array.from, new Map or new Set"
	probes := []struct{ name, source, reason string }{
		{"spread", host(numbers + "console.log([...numbers()].join(','));"), collected},
		{"Array.from", host(numbers + "console.log(Array.from(numbers()).join(','));"), collected},
		{"new Set", host(numbers + "console.log(String(new Set(numbers()).size));"), collected},
		{"new Map", host("function* pairs(): Generator<[string, number], undefined, undefined> { yield ['a', 1]; }\nconsole.log(String(new Map(pairs()).size));"), collected},
		{"an assignment to a captured local", write("count = 1;"), written},
		{"a compound assignment", write("count += 1;"), written},
		{"an increment", write("count++;"), written},
		{"a destructuring target", write("[count] = [2];"), written},
		{"a logical assignment", host("let seen = false;\nfunction* flagged(): Generator<boolean, undefined, undefined> { seen ||= true; yield seen; }\nfor (const value of flagged()) console.log(String(value));"), "where 'seen', a captured local, is written inside a nested generator"},
		{"a write by an arrow made inside the generator", write("const bump = (): void => { count++; }; bump();"), written},
		// A prefix update's value is a helper lowering makes, with no declaration of its own, so the
		// count reaches it by the MakeClosure a function inside makes it with. Mutant: that edge
		// dropped, and this row lowers.
		{"a prefix update's helper", host("let n = 0;\nfunction* counted(): Generator<number, undefined, undefined> { yield ++n; }\nfor (const value of counted()) console.log(String(value));"), "where 'n', a captured local, is written inside a nested generator"},
		// An object literal's getter is a function of its own, declared within the generator, so the
		// count reaches it by containment (Compiler's rule). Mutant: containment reduced to the
		// generator's own body, and this row lowers.
		{"an object literal's getter in the body", host("let n = 0;\nfunction* counted(): Generator<number, undefined, undefined> { const o = { get x(): number { n++; return n; } }; yield o.x; }\nfor (const value of counted()) console.log(String(value));"), "where 'n', a captured local, is written inside a nested generator"},
		// A named function declared in the body stops first as a nested declaration in a generator
		// (checkGeneratorBody's NestedFrame clause). Mutant: that clause dropped; the row then stops at
		// the write count instead, which reaches bump by containment, and reads that stop's words.
		{"a named function in the body writing a captured local", host("let n = 0;\nfunction* counted(): Generator<number, undefined, undefined> { function bump(): void { n++; } bump(); yield n; }\nfor (const value of counted()) console.log(String(value));"), "declaring a nested function"},
		{"a class in the body whose method writes a captured local", host("let n = 0;\nfunction* counted(): Generator<number, undefined, undefined> { class Counter { bump(): number { n++; return n; } } yield new Counter().bump(); }\nfor (const value of counted()) console.log(String(value));"), "a class inside a function"},
		{"a nested generic generator", host("function* inner<U>(seed: U): Generator<U, undefined, undefined> { yield seed; }\nfor (const value of inner(1)) console.log(String(value));"), "declared generic inside a function"},
		{"inside a generic function", "function host<T>(seed: T): T {\nfunction* inner(): Generator<number, undefined, undefined> { yield 1; }\nfor (const value of inner()) console.log(String(value));\nreturn seed;\n}\nconsole.log(host('s'));\n", "declared inside a generic function or class"},
		{"inside an async function", "async function host(): Promise<void> {\nfunction* inner(): Generator<number, undefined, undefined> { yield 1; }\nfor (const value of inner()) console.log(String(value));\n}\nawait host();\n", "declared inside an async function or a generator"},
		{"inside a generator", "function* outer(): Generator<number, undefined, undefined> {\nfunction* inner(): Generator<number, undefined, undefined> { yield 1; }\nyield 2;\n}\nfor (const value of outer()) console.log(String(value));\n", "declared inside an async function or a generator"},
		{"in a block", host("if (String(1).length > 0) {\nfunction* inner(): Generator<number, undefined, undefined> { yield 1; }\nfor (const value of inner()) console.log(String(value));\n}"), "declared inside a function or a block"},
		{"the function as a value", host(numbers + "const make = numbers;\nfor (const value of make()) console.log(String(value));"), "used as a value"},
		{"stored", host(numbers + "const held = numbers();\nconsole.log(String(held.next().value));"), "used other than by a for-of"},
		{"passed as an argument", "function take(source: object): number { return source === undefined ? 0 : 1; }\n" + host(numbers+"console.log(String(take(numbers())));"), "used other than by a for-of"},
		{"returned", "function outer(): Generator<number, undefined, undefined> {\n" + numbers + "return numbers();\n}\nfor (const value of outer()) console.log(String(value));\n", "used other than by a for-of"},
		{"consumed inside a nested generator", host(numbers + "function* again(): Generator<number, undefined, undefined> { for (const value of numbers()) { yield value; } }\nfor (const value of again()) console.log(String(value));"), "inside an async function or a generator"},
		{"destructured", host(numbers + "const [first] = numbers();\nconsole.log(String(first));"), "consumed by destructuring"},
		{"a closure over its own local", host("function* inner(): Generator<number, undefined, undefined> { let own = 1; const bump = (): void => { own++; }; bump(); yield own; }\nfor (const value of inner()) console.log(String(value));"), "a closure captures"},
		{"declaring a nested function", host("function* inner(): Generator<number, undefined, undefined> { function helper(): number { return 1; } yield helper(); }\nfor (const value of inner()) console.log(String(value));"), "declaring a nested function"},
	}
	for _, probe := range probes {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				t.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
	// A plain function* infers a void TReturn, which destructuring reads: adamic/no-void-read refuses
	// it before lowering, for a nested generator's call as for a module one's.
	_, err := lowerSource(t, host("function* plain() { yield 1; }\nconst [first] = plain();\nconsole.log(String(first));"))
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "adamic/no-void-read") {
		t.Errorf("destructuring a plain nested function*: got %v, want adamic/no-void-read's refusal", err)
	}
}

// Compiler's guard on #qk8rztp S3a: a nested generator's frame points out at its parent's cells,
// through its carrier, and that edge closes no cycle only while nothing those cells reach can hold a
// generator. That is a fact of every lowered program, not a sentence, so this reads it off each
// generator and iterable view fixture: no captured or environment local holds a Generator, and no
// parameter a generator is handed into through IterableOf is captured or an environment cell.
// S2c-4, which makes generators values, has to meet it, or first extend cycles.go's reachLinks
// with the edge from a generator to its carrier's cells (docs/generators.md, the frame's rule 4).
// The controls are a captured Generator local and a captured view parameter put into a lowered
// program by hand, each of which the guard must see.
func TestNoCapturedOrEnvironmentLocalCanHoldAGenerator(t *testing.T) {
	t.Parallel()
	var paths []string
	for _, pattern := range []string{"../oracle/testdata/generators/*.a", "../oracle/testdata/iterable_views/*.a"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	if len(paths) < 50 {
		t.Fatalf("found only %d programs: the globs no longer find the fixtures", len(paths))
	}
	var nested *ir.Program
	for _, path := range paths {
		loaded, err := load.Load([]string{path})
		if err != nil {
			t.Fatalf("%s: Load: %v", path, err)
		}
		program, err := Lower(context.Background(), loaded)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		for _, offense := range capturedGenerators(program) {
			t.Errorf("%s: %s", path, offense)
		}
		if filepath.Base(path) == "nested_iterable.a" {
			nested = program
		}
	}
	if nested == nil {
		t.Fatal("no nested_iterable.a among the fixtures")
	}
	held := nested.Locals
	nested.Locals = append(append([]ir.Local{}, held...), ir.Local{Name: "held", Type: ir.Generator, Function: 0, Captured: true, EnvironmentCell: true})
	if len(capturedGenerators(nested)) == 0 {
		t.Error("the control's captured Generator local went unseen: the guard can't fail")
	}
	nested.Locals = append([]ir.Local{}, held...)
	parameters := 0
	for _, function := range nested.Functions {
		walk(function.Body, func(node any) bool {
			if call, isCall := node.(ir.Call); isCall {
				for position, argument := range call.Arguments {
					if _, isIterable := argument.(ir.IterableOf); !isIterable {
						continue
					}
					for _, target := range nested.CallTargets(call) {
						nested.Locals[nested.Functions[target].Parameters[position]].Captured = true
						parameters++
					}
				}
			}
			return true
		})
	}
	if parameters == 0 || len(capturedGenerators(nested)) == 0 {
		t.Errorf("the control's %d captured view parameters went unseen: the guard can't fail", parameters)
	}
}

// capturedGenerators is every local of program that a closure's cell or a frame's environment holds
// and that can hold a generator, said for the guard: one whose representation is a Generator, and a
// parameter a generator's call is handed into through IterableOf, the one door from a generator to a
// slot of another representation (checkGeneratorRoads).
func capturedGenerators(program *ir.Program) []string {
	var offenses []string
	for _, declared := range program.Locals {
		if (declared.Captured || declared.EnvironmentCell) && declared.Type == ir.Generator {
			offenses = append(offenses, "the captured local "+declared.Name+" holds a Generator")
		}
	}
	visit := func(node any) bool {
		call, isCall := node.(ir.Call)
		if !isCall {
			return true
		}
		for position, argument := range call.Arguments {
			if _, isIterable := argument.(ir.IterableOf); !isIterable {
				continue
			}
			for _, target := range program.CallTargets(call) {
				parameter := program.Locals[program.Functions[target].Parameters[position]]
				if parameter.Captured || parameter.EnvironmentCell {
					offenses = append(offenses, "the captured parameter "+parameter.Name+" is handed a generator")
				}
			}
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return offenses
}
