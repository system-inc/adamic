package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Every generator form #qk8rztp's S1c doesn't lower stops as not yet built, never compiles. The
// value roads first (Compiler, 04:42): no slot but the consuming loop's own may hold a generator, so
// a generator stored, passed, returned, put in an array, spread or read through a view each stops.
// Then the forms the body or the declaration can take that its frame can't hold yet. Each row's
// reason is a word of its stop, so a row that stops for some other reason goes red too. Mutant: the
// generator arm of expression.go's call taken out; five value-road rows go red, two of them on
// checkGeneratorRoads' compiler bug, and none of the five compiles.
func TestGeneratorFormsThisSliceDoesntLowerAreNotYet(t *testing.T) {
	t.Parallel()
	// TReturn is undefined, not void: a value road of a Generator<T, void> is refused earlier by the
	// void rules (adamic/no-void-read, no-void-widening), which read its last result's value as void.
	const numbers = "function* numbers(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index; } }\n"
	probes := []struct{ name, source, reason string }{
		{"stored in a variable", numbers + "const held = numbers(2); console.log(String(held.next().value));", "used other than by a for-of"},
		{"passed as an argument", numbers + "function take(source: object): number { return source === undefined ? 0 : 1; } console.log(String(take(numbers(2))));", "used other than by a for-of"},
		{"returned", numbers + "function make(): Generator<number, undefined, undefined> { return numbers(2); } for (const value of make()) { console.log(String(value)); }", "used other than by a for-of"},
		{"put in an array", numbers + "const all = [numbers(1), numbers(2)]; console.log(String(all.length));", "used other than by a for-of"},
		{"Array.from with a mapper", numbers + "console.log(Array.from(numbers(3), (value: number): number => value * 2).join(','));", "Array.from with a mapper"},
		{"new Set", numbers + "console.log(String(new Set(numbers(3)).size));", "consumed by new Set"},
		{"destructured", numbers + "const [first] = numbers(3); console.log(String(first));", "consumed by destructuring"},
		{"spread inside a generator", numbers + "function* outer(): Generator<number, undefined, undefined> { const all = [...numbers(2)]; yield all.length; } for (const value of outer()) { console.log(String(value)); }", "inside an async function or a generator"},
		{"read through a view", numbers + "const view: Iterable<number> = numbers(2); for (const value of view) { console.log(String(value)); }", "used other than by a for-of"},
		{"the function as a value", numbers + "const make = numbers; for (const value of make(2)) { console.log(String(value)); }", "used as a value"},
		{"a generator expression", "const make = function* (): Generator<number, void, undefined> { yield 1; }; for (const value of make()) { console.log(String(value)); }", "a generator function"},
		{"a generator method", "class Counter { *each(): Generator<number, void, undefined> { yield 1; } } for (const value of new Counter().each()) { console.log(String(value)); }", "a generator function"},
		{"a nested generator", "function outer(): void { function* inner(): Generator<number, void, undefined> { yield 1; } for (const value of inner()) { console.log(String(value)); } } outer();", "a generator function"},
		{"a generic generator", "function* each<T>(items: T[]): Generator<T, void, undefined> { for (const item of items) { yield item; } } for (const value of each([1, 2])) { console.log(String(value)); }", "a generator function"},
		{"yield*", numbers + "function* again(): Generator<number, undefined, undefined> { yield* numbers(2); } for (const value of again()) { console.log(String(value)); }", "delegating with yield*"},
		{"a yield inside try", "function* guarded(): Generator<number, void, undefined> { try { yield 1; } catch (error) { console.log('caught'); } } for (const value of guarded()) { console.log(String(value)); }", "yielding inside try"},
		{"a yield's result", "function* sent(): Generator<number, void, undefined> { const got = yield 1; console.log(String(got)); } for (const value of sent()) { console.log(String(value)); }", "a yield's result"},
		{"a closure over a body local", "function* counted(): Generator<number, void, undefined> { let count = 0; const bump = (): void => { count++; }; bump(); yield count; } for (const value of counted()) { console.log(String(value)); }", "a closure captures"},
		{"a closure over a parameter", "function* scaled(by: number): Generator<number, void, undefined> { const times = (value: number): number => value * by; yield times(2); } for (const value of scaled(3)) { console.log(String(value)); }", "a closure captures"},
		{"returning a value", "function* finished(): Generator<number, string, undefined> { yield 1; return 'end'; } for (const value of finished()) { console.log(String(value)); }", "returning a value"},
		{"a defaulted parameter", "function* from(start = 1): Generator<number, void, undefined> { yield start; } for (const value of from()) { console.log(String(value)); }", "plain required name"},
		{"a declared ArrayIterator", "function* each(): ArrayIterator<number> { yield 1; } for (const value of each()) { console.log(String(value)); }", "declared to return"},
		{"a union yield", "function* mixed(): Generator<number | string, void, undefined> { yield 1; yield 'a'; } for (const value of mixed()) { console.log(String(value)); }", "one present slot"},
		{"consumed inside a generator", numbers + "function* outer(): Generator<number, void, undefined> { for (const value of numbers(2)) { yield value; } } for (const value of outer()) { console.log(String(value)); }", "inside an async function or a generator"},
		{"consumed inside an async function", numbers + "async function run(): Promise<void> { for (const value of numbers(2)) { console.log(String(value)); } } await run();", "inside an async function or a generator"},
		// Lower's fixture (04:44): the call keeps its targets, so a try around a loop over a
		// generator whose body reaches a failing library call still stops, as any try around one does.
		{"a try reaching a failing library call", "function* padded(width: number): Generator<string, void, undefined> { yield 'x'.repeat(width); } try { for (const line of padded(3)) { console.log(line); } } catch (error) { console.log('caught'); }", "a try around"},
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
}

// What S1c lowers lowers: the generator is a function whose call makes a Generator, its body yields
// as statements, and its consumers are a for-of stepping the call and a call dropped as a statement.
// The step of a generator whose body can throw is one a throw can leave (ResumeMayThrow), and the
// generator's own call is not one (MayThrow stays false, since the call runs none of the body).
func TestGeneratorDeclarationsLowerOntoTheirFrame(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function* checked(values: number[]): Generator<number, void, undefined> {
	for (const value of values) {
		if (value < 0) {
			throw new Error("negative");
		}
		yield value;
	}
}
function* quiet(): Generator<string, void, undefined> {
	yield "a";
}
for (const value of checked([1, 2])) {
	console.log(String(value));
}
quiet();`)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]ir.Function{}
	for _, function := range program.Functions {
		if function.Generator {
			found[function.Name] = function
		}
	}
	checked, quiet := found["checked"], found["quiet"]
	if checked.Returns != ir.Generator || checked.Yields != ir.Number || !checked.ResumeMayThrow || checked.MayThrow {
		t.Errorf("checked: Returns %d, Yields %d, ResumeMayThrow %t, MayThrow %t; want a Generator of numbers whose steps can throw and whose call can't", checked.Returns, checked.Yields, checked.ResumeMayThrow, checked.MayThrow)
	}
	if quiet.Returns != ir.Generator || quiet.Yields != ir.String || quiet.ResumeMayThrow {
		t.Errorf("quiet: Returns %d, Yields %d, ResumeMayThrow %t; want a Generator of strings that can't throw", quiet.Returns, quiet.Yields, quiet.ResumeMayThrow)
	}
	if err := checkGeneratorRoads(program); err != nil {
		t.Errorf("the lowered program's roads: %v", err)
	}
}

// checkGeneratorRoads is the backstop under the lowering's own stops: a generator value anywhere
// but the consuming loop's local and a dropped call is a compiler bug. Its mutant is the program a
// missed stop would make, a generator call stored in an ordinary local.
func TestGeneratorRoadsBackstopFailsOnAStoredGenerator(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "function* numbers(): Generator<number, void, undefined> { yield 1; } numbers();")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkGeneratorRoads(program); err != nil {
		t.Fatalf("the lowered program: %v", err)
	}
	var call ir.Call
	for _, statement := range program.Main {
		if evaluate, ok := statement.(ir.Evaluate); ok {
			if made, ok := evaluate.Value.(ir.Call); ok && made.Returns == ir.Generator {
				call = made
			}
		}
	}
	if call.Returns != ir.Generator {
		t.Fatal("no dropped generator call in main")
	}
	local := len(program.Locals)
	program.Locals = append(program.Locals, ir.Local{Name: "held", Type: ir.Object, Function: -1})
	program.Main = append(program.Main, ir.Declare{Local: local, Value: call})
	if err := checkGeneratorRoads(program); err == nil {
		t.Fatal("a generator call stored in an ordinary local passed the backstop")
	}
}

// A plain function* infers Generator<T, void, unknown>, whose void is only its last value. A spread and
// Array.from step it to its end and never read that value, so the void rules let them through
// (generatorStepped), and a destructuring, which this slice doesn't lower, is still refused by
// adamic/no-void-read as before. Mutant: generatorStepped's call taken out of refuseVoidContainerRead;
// both spread rows read Refused (as a spread). Taken out of refuseVoidWidening; the Array.from row
// reads Refused (seen as any).
func TestGeneratorVoidReturnIsNeverReadBySpreadOrArrayFrom(t *testing.T) {
	t.Parallel()
	const plain = "function* numbers(limit: number) { for (let index = 0; index < limit; index++) { yield index; } }\n"
	for _, source := range []string{
		plain + "console.log([...numbers(3)].join(','));",
		plain + "console.log([0, ...numbers(2)].join(','));",
		plain + "console.log(Array.from(numbers(3)).join(','));",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("%s: %v", source, err)
		}
	}
	_, err := lowerSource(t, plain+"const [first] = numbers(3); console.log(String(first));")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "adamic/no-void-read") {
		t.Errorf("destructuring a plain function*: got %v, want adamic/no-void-read's refusal", err)
	}
	// A yielded type that holds void is read by every step, so the rules still judge it.
	_, err = lowerSource(t, "function* voids(): Generator<void[], void, undefined> { yield []; }\nconsole.log(String([...voids()].length));")
	if !errors.As(err, &refused) {
		t.Errorf("spreading a generator of void[]: got %v, want a void rule's refusal", err)
	}
	// Adamic's conditions (#qk8rztp, 20:22Z). A generator whose yields are void calls is read at
	// every step, so spread and Array.from of it stay refused; the mutant that drops the yielded-type
	// test admits both. Only a direct call is exempt: a stored generator, spread, stays refused; the
	// mutant that widens the exemption to any Generator-typed expression lets it through to lowering.
	const voids = "function f(): void {}\nfunction* g() { yield f(); }\n"
	for _, source := range []string{
		voids + "console.log(String([...g()].length));",
		voids + "console.log(String(Array.from(g()).length));",
		plain + "const stored = numbers(3); console.log([...stored].join(','));",
		// Language's attack: a mapper is a second argument, outside the exemption.
		plain + "console.log(Array.from(numbers(3), (value) => value).join(','));",
	} {
		_, err := lowerSource(t, source)
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "adamic/no-void") {
			t.Errorf("%s: got %v, want a void rule's refusal", source, err)
		}
	}
	// Language's other attack: a generic generator spread at a void instantiation. It is no lowered
	// generator, so the exemption never applies to it, and today rule 2500 stops it before the void
	// rules do. The slice that lowers generics must keep it unadmitted, by judging T at the call.
	generic := "function f(): void {}\nfunction* each<T>(value: T) { yield value; }\nconsole.log(String([...each(f())].length));"
	if _, err := lowerSource(t, generic); err == nil {
		t.Errorf("a generic generator spread at a void instantiation lowered; want it refused or not yet")
	}
}
