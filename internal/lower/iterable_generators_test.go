package lower

import (
	"errors"
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// iterableCounted is a generator whose call the probes below hand into an Iterable<number> parameter.
// Its TReturn is undefined, not void, so a road the one-read rule closes reaches lowering: a road of a
// Generator<T, void> outside a stepped site is refused earlier by the void rules.
const iterableCounted = "function* counted(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n"

// iterableWords is a generator of built strings, for an Iterable<string> parameter, held as a union.
const iterableWords = "function* words(limit: number): Generator<string, undefined, undefined> { for (let index = 0; index < limit; index++) { yield `w${String(index)}`; } }\n"

// iterableTotal is a function whose one use of its Iterable<number> parameter is one for-of over it.
const iterableTotal = "function total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\n"

// iterableFirst is tsc's firstOrUndefinedIterator (core.ts:1088), verbatim: its Iterable<T> | undefined
// parameter's uses are one undefined test and one for-of (#qk8rztp S2c-3).
const iterableFirst = "function firstOrUndefinedIterator<T>(iter: Iterable<T> | undefined): T | undefined { if (iter !== undefined) { for (const value of iter) { return value; } } return undefined; }\n"

// iterableShown prints what iterableFirst gives back, undefined among it.
const iterableShown = "function shown(value: number | undefined): string { return value === undefined ? 'undefined' : String(value); }\n"

// A generator's direct call passed straight into a one-body function's Iterable<T> parameter, whose
// only use is one for-of, lowers (#qk8rztp S2c-1, Adamic's ruling): a module function, through
// IterableIterator<T> too, a generic instance, and the shapes Language named that must stay
// admitted, a type-only typeof and a nested block's own name. The oracle holds what these run to
// (testdata/generators); this holds that they lower at all.
func TestGeneratorIntoAnIterableParameterLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"a module function", iterableCounted + iterableTotal + "console.log(String(total(counted(3))));"},
		// steppedSite's arm: a void TReturn is never read through the parameter's one for-of.
		{"a void TReturn", "function* counted(limit: number): Generator<number, void, unknown> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n" + iterableTotal + "console.log(String(total(counted(3))));"},
		{"an IterableIterator parameter", iterableCounted + "function total(items: IterableIterator<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
		{"a generic instance", iterableCounted + "function weigh<T>(items: Iterable<T>, by: (item: T) => number): number { let sum = 0; for (const item of items) { sum += by(item); } return sum; }\nconsole.log(String(weigh(counted(3), (value: number): number => value + 1)));"},
		{"one function at two call sites", iterableCounted + iterableTotal + "console.log(String(total(counted(2)) + total(counted(4))));"},
		// The union arm: an Iterable<string> slot fed a generator, and fed a Set's keys() too.
		{"an Iterable<string> parameter", iterableWords + "function joined(items: Iterable<string>): string { let all = ''; for (const word of items) { all += word; } return all; }\nconsole.log(joined(words(2)));"},
		{"an Iterable<string> parameter beside a Set's keys()", iterableWords + "function joined(items: Iterable<string>): string { let all = ''; for (const word of items) { all += word; } return all; }\nconst seen = new Set<string>([`s${String(1)}`]);\nconsole.log(joined(words(2)) + joined(seen.keys()));"},
		// The dispatch's two arms: one slot fed a generator at one call and a Map's values() at another.
		{"a generator beside a collection's iterator", iterableCounted + iterableTotal + "const sizes = new Map<string, number>([[`a${String(1)}`, 1]]);\nconsole.log(String(total(counted(2)) + total(sizes.values())));"},
		{"an outer loop over the for-of", iterableCounted + "function firsts(items: Iterable<number>): number { let sum = 0; for (let pass = 0; pass < 2; pass++) { for (const value of items) { sum += value; break; } } return sum; }\nconsole.log(String(firsts(counted(3))));"},
		{"a type-only typeof", iterableCounted + "function total(items: Iterable<number>): number { const none: typeof items | undefined = undefined; let sum = none === undefined ? 0 : 1; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
		{"a nested block's own name", iterableCounted + "function total(items: Iterable<number>): number { let sum = 0; { const items = 4; sum += items; } for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
		// S2c-3: a nullable parameter, held as an object, its undefined tests not counted as reads.
		{"tsc's firstOrUndefinedIterator", iterableCounted + iterableFirst + iterableShown + "console.log(shown(firstOrUndefinedIterator(counted(3))));"},
		{"firstOrUndefinedIterator beside a Map's values()", iterableCounted + iterableFirst + iterableShown + "const sizes = new Map<string, number>([[`a${String(1)}`, 1]]);\nconsole.log(shown(firstOrUndefinedIterator(counted(3))) + shown(firstOrUndefinedIterator(sizes.values())));"},
		{"a nullable parameter tested === undefined, returning early", iterableCounted + "function first(items: Iterable<number> | undefined): number { if (items === undefined) { return -1; } for (const value of items) { return value; } return -1; }\nconsole.log(String(first(counted(3))));"},
		{"an IterableIterator<T> | undefined parameter", iterableCounted + "function first(items: IterableIterator<number> | undefined): number { if (items !== undefined) { for (const value of items) { return value; } } return -1; }\nconsole.log(String(first(counted(3))));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
		})
	}
}

// Every road the one-read rule closes stops as not yet built, each in its own words (Adamic's ruling
// on #qk8rztp S2c-1, conditions 2 and 3, Compiler's roads and Language's shapes): a second read of the
// parameter in the lowered IR (a typeof or an identity test, a read after a labeled break out of the
// loop, a default that names it; a hand-driven next is refused earlier, below), a
// closure that captures it, never called, passing it on, assigning it, reading arguments, and a callee
// the call can't name as one body (an async function, a function value, a method), or a generator
// that isn't the argument's own call (one stored first). And what S2c-1 doesn't build: two generator
// functions into one parameter, a string beside a generator into an Iterable<string> parameter, an
// origin no view admits, and yields held otherwise than the view reads them.
// Spreading or destructuring the parameter stops before the proof, in its own words, so those rows
// hold only that it doesn't compile. Mutants (internal/oracle/testdata/generators/mutants.md): the
// one-read test made false admits the second-read rows; the call site taking any generator's call as
// an Iterable<T> argument turns the stored row into a compiler bug.
func TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad(t *testing.T) {
	t.Parallel()
	const readTwice = "read 2 times"
	const halved = "function* halved(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index / 2; } }\n"
	const other = "used other than by a for-of"
	function := func(body string) string {
		return iterableCounted + "function total(items: IterableIterator<number>): number { let sum = 0; " + body + " return sum; }\nconsole.log(String(total(counted(3))));"
	}
	for _, probe := range []struct{ name, source, reason string }{
		{"typeof it", function("if (typeof items === 'object') { sum += 1; } for (const value of items) { sum += value; }"), readTwice},
		{"a read after a labeled break", function("outer: for (let pass = 0; pass < 2; pass++) { for (const value of items) { sum += value; break outer; } } if (typeof items === 'object') { sum += 1; }"), readTwice},
		{"it === other", iterableCounted + "function total(items: Iterable<number>, other: Iterable<number>): number { let sum = items === other ? 1 : 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2), counted(3))));", readTwice},
		{"a default naming it", iterableCounted + "function total(items: Iterable<number>, base: number = typeof items === 'object' ? 1 : 0): number { let sum = base; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2))));", readTwice},
		{"a closure capturing it, never called", function("const keep = (): Iterable<number> => items; for (const value of items) { sum += value; }"), "which a closure captures"},
		{"passing it on", iterableCounted + iterableTotal + "function outer(items: Iterable<number>): number { return total(items); }\nconsole.log(String(outer(counted(3))));", "whose one use isn't a for-of over it"},
		{"reassigning it", iterableCounted + "function total(items: Iterable<number>, other: Iterable<number>): number { let sum = 0; items = other; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2), counted(3))));", "which the function writes"},
		{"arguments", function("sum += arguments.length; for (const value of items) { sum += value; }"), "reads arguments"},
		{"an async callee", iterableCounted + "async function total(items: Iterable<number>): Promise<number> { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(await total(counted(2))));", other},
		{"a call through a function value", iterableCounted + iterableTotal + "const add = total;\nconsole.log(String(add(counted(2))));", other},
		{"a method", iterableCounted + "class Summer { total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; } }\nconsole.log(String(new Summer().total(counted(2))));", other},
		{"a stored generator passed in", iterableCounted + iterableTotal + "const made = counted(2);\nconsole.log(String(total(made)));", other},
		// Language's identity and reassignment probes with a generator as the other value. Written
		// as another generator's call in the callee, it stops where that call is lowered, a generator
		// used other than by a consumer, before any count; given as a second generator parameter, the
		// one-read count is what stops it (the "it === other" and "reassigning it" rows above).
		{"it === g2()", iterableCounted + halved + "function total(items: Iterable<number>): number { let sum = items === halved(2) ? 1 : 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));", other},
		{"it = g2() before the loop", iterableCounted + halved + "function total(items: Iterable<number>): number { let sum = 0; items = halved(2); for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));", other},
		{"two generator functions", iterableCounted + iterableTotal + halved + "console.log(String(total(counted(2)) + total(halved(2))));", "generators of more than one function"},
		{"a string beside a generator into Iterable<string>", iterableWords + "function joined(items: Iterable<string>): string { let all = ''; for (const word of items) { all += word; } return all; }\nconsole.log(joined(words(2)) + joined(`text ${String(1)}`));", "unproven origin"},
		{"yields held otherwise", iterableCounted + "function show(items: Iterable<number | boolean>): void { for (const value of items) { console.log(String(value)); } }\nshow(counted(2));", "held differently"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				t.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
	// A hand-driven next through the parameter, before the loop or inside its body, is refused before
	// the proof: a library iterator's next read off an object is the unbound-method rule's (an
	// inherited library member isn't an own field). That stop fails closed; were it ever lifted, the
	// second read is counted, and the one-read rule stops the program as not yet built.
	for _, probe := range []struct{ name, source string }{
		{"next through the parameter", function("const first = items.next(); if (!first.done) { sum += first.value; } for (const value of items) { sum += value; }")},
		{"next inside the loop", function("for (const value of items) { items.next(); sum += value; }")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "(unbound-method)") {
				t.Fatalf("got %v, want the unbound-method refusal", err)
			}
		})
	}
	// A for-of head and a destructuring assignment that rebind the parameter stop before the proof
	// today, a for-of whose head isn't a const or let declaration and a pattern assigned to an existing
	// name being not yet built, so these rows hold only that neither compiles; were either to lower,
	// its ForOf.Local or Binding.Local write is counted (TestLocalUseCountSeesEveryWrite).
	const sizes = "const sizes = new Map<string, number>([[`a${String(1)}`, 1]]);\n"
	for _, probe := range []struct{ name, source string }{
		{"spreading it", function("const all = [...items]; sum += all.length; for (const value of items) { sum += value; }")},
		{"destructuring it", function("const [first] = items; sum += first ?? 0; for (const value of items) { sum += value; }")},
		{"a for-of head rebinding it", sizes + function("for (items of [sizes.values()]) { sum += 1; } for (const value of items) { sum += value; }")},
		{"a destructuring assignment rebinding it", sizes + function("[items] = [sizes.values()]; for (const value of items) { sum += value; }")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			var refused *Refused
			if !errors.As(err, &gap) && !errors.As(err, &refused) {
				t.Fatalf("got %v, want it stopped", err)
			}
		})
	}
}

// A nullable Iterable<T> parameter's undefined tests are the one use besides its for-of the one-read
// rule doesn't count (#qk8rztp S2c-3, Adamic's ruling: strict tests against the literal undefined
// only), and only an ir.IsUndefined whose value is a bare read of it: a truthiness test, a null test
// beside the undefined one, a typeof test (a string comparison of ir.TypeOf over the read), `??`, an
// undefined test whose operand holds the read inside a conditional, and a second for-of each count as
// a read. A test inside an arrow is a capture. The door stays shut to `| null` and to `?`, which stop
// where the call is lowered, a generator used other than by a consumer, and the rule stops
// `Iterable<string> | undefined`, held as a union. An undefined passed in at another call site stops
// not yet, never refused (Node runs it, as `return undefined`): the view can't name it as an origin.
// Mutants (internal/oracle/testdata/generators/mutants.md): the exemption widened to a null test
// admits the null-test row, and to any ir.IsUndefined the conditional row; the door's undefined test
// widened to null turns the `| null` row into the one-read stop; the union check taken out admits the
// Iterable<string> row.
func TestNullishIterableParameterStopsOnEveryOtherRoad(t *testing.T) {
	t.Parallel()
	const readTwice = "read 2 times"
	const other = "used other than by a for-of"
	first := func(parameter, body string) string {
		return iterableCounted + "function first(" + parameter + "): number { let found = -1; " + body + " return found; }\nconsole.log(String(first(counted(3))));"
	}
	const nullable = "items: Iterable<number> | undefined"
	const loop = "if (items !== undefined) { for (const value of items) { found = value; break; } }"
	for _, probe := range []struct{ name, source, reason string }{
		{"a truthiness test", first(nullable, "if (items) { for (const value of items) { found = value; break; } }"), readTwice},
		{"a typeof test", first(nullable, "if (typeof items === 'undefined') { return found; } for (const value of items) { found = value; break; }"), readTwice},
		{"??", first(nullable, "const kept = items ?? undefined; found += kept === undefined ? 0 : 1; "+loop), readTwice},
		// Lower's condition 3: the test's operand holds the read inside something, so it is no
		// undefined test of the parameter and its read counts. The any-operand mutant's red row.
		{"an undefined test of a conditional holding it", iterableCounted + "function first(items: Iterable<number> | undefined, pick: boolean): number { let found = -1; if ((pick ? items : undefined) !== undefined && items !== undefined) { for (const value of items) { found = value; break; } } return found; }\nconsole.log(String(first(counted(3), 'ab'.length > 1)));", readTwice},
		{"an undefined test inside an arrow, called", first(nullable, "const empty = (): boolean => items === undefined; found = empty() ? 0 : 1; "+loop), "which a closure captures"},
		{"an undefined test inside an arrow, never called", first(nullable, "const empty = (): boolean => items === undefined; "+loop), "which a closure captures"},
		{"a null test beside the undefined test", first(nullable, "if (items !== null && items !== undefined) { for (const value of items) { found = value; break; } }"), readTwice},
		{"an undefined test and a second for-of", first(nullable, "if (items !== undefined) { for (const value of items) { found = value; break; } for (const value of items) { found += value; } }"), readTwice},
		{"| null", first("items: Iterable<number> | null", "if (items !== null) { for (const value of items) { found = value; break; } }"), other},
		{"an optional parameter", first("items?: Iterable<number>", "if (items !== undefined) { for (const value of items) { found = value; break; } }"), other},
		{"Iterable<string> | undefined", iterableWords + "function joined(items: Iterable<string> | undefined): string { let all = ''; if (items !== undefined) { for (const word of items) { all += word; } } return all; }\nconsole.log(joined(words(2)));", "which may be undefined and is held as a union"},
		{"undefined passed at another call site", iterableCounted + iterableFirst + iterableShown + "console.log(shown(firstOrUndefinedIterator(counted(3))) + shown(firstOrUndefinedIterator<number>(undefined)));", "a view over an iterable of unproven origin"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				t.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
	// `items == null` is refused before the proof, by the rule against ==; were that lifted, its read
	// is counted.
	for _, probe := range []struct{ name, source, reason string }{
		{"items == null", first(nullable, "if (items == null) { return found; } for (const value of items) { found = value; break; }"), "refuses =="},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want a refusal saying %q", err, probe.reason)
			}
		})
	}
}

// An optional chain on the parameter, a method read off it by a computed key, stops before the
// proof; the words are expected to be the unbound-method refusal or a not yet at the element access,
// and are pinned from the box's run. Were it ever lowered, its read is counted.
func TestNullishIterableParameterOptionalChainStops(t *testing.T) {
	t.Parallel()
	source := iterableCounted + "function first(items: Iterable<number> | undefined): number { let found = -1; if (items?.[Symbol.iterator] !== undefined && items !== undefined) { for (const value of items) { found = value; break; } } return found; }\nconsole.log(String(first(counted(3))));"
	if _, err := lowerSource(t, source); err == nil {
		t.Fatal("an optional chain on a nullable Iterable<T> parameter a generator feeds lowered, want it stopped")
	}
}

// A nullable parameter's for-of, inside its undefined test, reads it through ir.Defined, the
// narrowing's check (narrowed.go's defined). The dispatch keeps that check in the holder's
// declaration, so an undefined the proof missed panics in Defined's words before any arm, never
// stepped (Lower's condition on #qk8rztp S2c-3). Mutant: dispatchView declaring the bare read, and
// this goes red.
func TestNullishViewHolderKeepsItsDefinedCheck(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, iterableCounted+iterableFirst+iterableShown+"console.log(shown(firstOrUndefinedIterator(counted(3))));")
	if err != nil {
		t.Fatal(err)
	}
	holders := 0
	inspect := func(node any) bool {
		block, isBlock := node.(ir.Block)
		if !isBlock || block.Closes == 0 || len(block.Body) == 0 {
			return true
		}
		declared, isDeclare := block.Body[0].(ir.Declare)
		if !isDeclare || declared.Local != block.Closes-1 {
			return true
		}
		holders++
		defined, isDefined := declared.Value.(ir.Defined)
		if !isDefined {
			t.Errorf("the dispatched view's holder is declared with %T, want the ir.Defined over the parameter's read", declared.Value)
			return true
		}
		if _, isRead := defined.Value.(ir.Read); !isRead {
			t.Errorf("the holder's ir.Defined holds %T, want the parameter's read", defined.Value)
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if holders != 1 {
		t.Fatalf("%d dispatched view holders, want 1", holders)
	}
}

// The view a generator feeds is rewritten into a per-step dispatch on what the object is (Lower's
// (b)): its block closes what it holds at every way out, and each pass tests whether the object is a
// generator, stepping it by GeneratorNext and GeneratorDone, else whether it is a collection's
// iterator, stepping it by the library loop's own step (one read of its next), else panics as a
// compiler bug. And the generator crosses into the parameter through IterableOf, as the call's own
// argument. The mutants that take out an arm are in testdata/generators/mutants.md.
func TestGeneratorViewDispatchesOnTheHeapKind(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, iterableCounted+iterableTotal+"console.log(String(total(counted(3))));")
	if err != nil {
		t.Fatal(err)
	}
	closes, kinds, collections, steps, nexts, iterables := 0, 0, 0, 0, 0, 0
	inspect := func(node any) bool {
		switch node := node.(type) {
		case ir.Block:
			if node.Closes != 0 {
				closes++
			}
		case ir.IsGenerator:
			kinds++
		case ir.IsCollectionIterator:
			collections++
		case ir.GeneratorNext:
			if node.Generator.Type() == ir.Object && program.Functions[node.Function].Generator {
				steps++
			}
		case ir.Call:
			for _, argument := range node.Arguments {
				if _, isIterable := argument.(ir.IterableOf); isIterable {
					iterables++
				}
			}
		case ir.Property:
			if node.Name == "next" {
				nexts++
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if closes != 1 || kinds != 1 || collections != 1 || steps != 1 || nexts != 1 || iterables != 1 {
		t.Fatalf("%d closing blocks, %d generator tests, %d collection tests, %d generator steps of an object, %d reads of a next and %d IterableOf arguments, want one each", closes, kinds, collections, steps, nexts, iterables)
	}
}

// Every IR field that names a local is counted by the one-read rule's walk, as a read, a write or
// another naming (Lower's must on #qk8rztp S2c-1): a for-of head, a pattern's binding and a catch
// write a local with no Assign. Mutant: ForOf.Local classified as no local, and the for-of row
// counts no write.
func TestLocalUseCountSeesEveryWrite(t *testing.T) {
	t.Parallel()
	const local = 7
	for _, probe := range []struct {
		name                 string
		node                 any
		reads, writes, named int
	}{
		{"a read", ir.Read{Local: local, Of: ir.Object}, 1, 0, 0},
		{"an assignment", ir.Assign{Local: local}, 0, 1, 0},
		{"a declaration", ir.Declare{Local: local}, 0, 1, 0},
		{"a for-of head", ir.ForOf{Local: local}, 0, 1, 0},
		{"a for-of pattern's binding", ir.ForOf{Local: 0, Pattern: []ir.Binding{{Local: local, Field: "0"}}}, 0, 1, 0},
		{"a for-of with a pattern leaves Local unused", ir.ForOf{Local: local, Pattern: []ir.Binding{{Local: 3, Field: "0"}}}, 0, 0, 0},
		{"a catch", ir.Try{CatchLocal: local}, 0, 1, 0},
		{"a closing block", ir.Block{Closes: local + 1}, 0, 0, 1},
		{"a per-iteration copy", ir.Loop{PerIteration: []int{local}}, 0, 0, 1},
		{"an environment's cell", ir.AllocateEnvironment{Cells: []int{local}}, 0, 0, 1},
		{"another local", ir.Assign{Local: local + 1}, 0, 0, 0},
	} {
		var uses localUseCount
		uses.node(probe.node, local)
		if uses.reads != probe.reads || uses.writes != probe.writes || uses.named != probe.named {
			t.Errorf("%s: %d reads, %d writes, %d other namings, want %d, %d, %d", probe.name, uses.reads, uses.writes, uses.named, probe.reads, probe.writes, probe.named)
		}
	}
}

// localFields lists every int and []int field of every struct the ir package's statements and
// expressions hold, read from its source, so a field added there that names a local can't be missed
// by the one-read rule's count: it fails here until it's classified. A field listed that the source
// no longer has fails too.
func TestLocalFieldsClassifyEveryIntField(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	structs := map[string]*goast.StructType{}
	roots := map[string]bool{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(file, func(node goast.Node) bool {
			switch node := node.(type) {
			case *goast.TypeSpec:
				if structure, isStruct := node.Type.(*goast.StructType); isStruct {
					structs[node.Name.Name] = structure
				}
			case *goast.FuncDecl:
				// A statement has the unexported statement(), an expression Type() Type.
				if node.Recv != nil && len(node.Recv.List) == 1 && (node.Name.Name == "statement" || node.Name.Name == "Type") {
					receiver := node.Recv.List[0].Type
					if star, isStar := receiver.(*goast.StarExpr); isStar {
						receiver = star.X
					}
					if identifier, isIdentifier := receiver.(*goast.Ident); isIdentifier {
						roots[identifier.Name] = true
					}
				}
			}
			return true
		})
	}
	found := map[string]bool{}
	seen := map[string]bool{}
	var queue []string
	for root := range roots {
		queue = append(queue, root)
	}
	sort.Strings(queue)
	for len(queue) != 0 {
		name := queue[0]
		queue = queue[1:]
		structure, isStruct := structs[name]
		if seen[name] || !isStruct {
			continue
		}
		seen[name] = true
		for _, field := range structure.Fields.List {
			held := field.Type
			if list, isList := held.(*goast.ArrayType); isList && list.Len == nil {
				held = list.Elt
			}
			if identifier, isIdentifier := held.(*goast.Ident); isIdentifier && identifier.Name == "int" {
				for _, fieldName := range field.Names {
					found[name+"."+fieldName.Name] = true
				}
			}
			goast.Inspect(field.Type, func(node goast.Node) bool {
				if identifier, isIdentifier := node.(*goast.Ident); isIdentifier && structs[identifier.Name] != nil && !seen[identifier.Name] {
					queue = append(queue, identifier.Name)
				}
				return true
			})
		}
	}
	for field := range found {
		classified, listed := localFields[field]
		if !listed {
			t.Errorf("ir field %s holds an int and isn't in localFields: say whether it reads, writes or names a local, or why it isn't one (local_uses.go)", field)
		} else if classified.role == localNot && strings.TrimSpace(classified.why) == "" {
			t.Errorf("ir field %s is listed as no local without saying why", field)
		}
	}
	for field := range localFields {
		if !found[field] {
			t.Errorf("localFields lists %s, which no IR statement or expression holds", field)
		}
	}
}

// checkGeneratorRoads counts every Generator-typed value in the lowered program and allows only a
// consumer's own and IterableOf's, a call's own argument; that count is the guard that IterableOf is
// the one door from a generator to an object's slot (Compiler's departure 1 on #qk8rztp S2c-1). A
// generator's call boxed or cast into an Object slot by hand is its compiler bug. Mutant: the count's
// test made false, and both rows go red.
func TestGeneratorRoadsRefuseAGeneratorBoxedOrCastIntoAnObject(t *testing.T) {
	t.Parallel()
	made := ir.Call{Function: 0, Returns: ir.Generator}
	for _, probe := range []struct {
		name  string
		value ir.Expression
	}{
		{"boxed", ir.Box{Value: made}},
		{"cast", ir.CheckedCast{Value: made}},
	} {
		program := &ir.Program{
			Functions: []ir.Function{{Name: "gen", Generator: true, Returns: ir.Generator}},
			Locals:    []ir.Local{{Name: "slot", Type: ir.Object, Function: -1}},
			Main:      []ir.Statement{ir.Declare{Local: 0, Value: probe.value}},
		}
		err := checkGeneratorRoads(program)
		if err == nil || !strings.Contains(err.Error(), "compiler bug") {
			t.Errorf("%s: got %v, want checkGeneratorRoads' compiler bug", probe.name, err)
		}
	}
}

// libraryIteratorLoop steps by libraryIteratorStep, which an iterable view's dispatch shares (Lower's
// condition 4: one step, never a copy). The loop's IR is what it was before the step was its own
// function, statement for statement, so every collection-iterator for-of emits the same C, which is
// a function of the IR alone; this holds the IR to the old shape, written out.
func TestLibraryIteratorLoopIsItsStepAndItsBody(t *testing.T) {
	t.Parallel()
	l := &lowering{result: &ir.Program{Locals: []ir.Local{{Name: "element", Type: ir.Number}}}}
	item := ir.Read{Local: 0, Of: ir.Number}
	iterator := ir.Read{Local: 0, Of: ir.Object}
	body := []ir.Statement{ir.Evaluate{Value: item}}
	got := l.libraryIteratorLoop(0, iterator, item, body)
	held := ir.Read{Local: 1, Of: ir.Object}
	step := ir.Read{Local: 2, Of: ir.Object}
	want := []ir.Statement{
		ir.Declare{Local: held.Local, Value: iterator},
		ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: []ir.Statement{
			ir.Declare{Local: step.Local, Value: ir.CallClosure{Closure: ir.Property{Object: held, Name: "next", Of: ir.Closure}, Returns: ir.Object}},
			ir.If{Condition: ir.Property{Object: step, Name: "done", Of: ir.Boolean}, Then: []ir.Statement{ir.Break{}}},
			ir.Declare{Local: item.Local, Value: ir.Property{Object: step, Name: "value", Of: item.Of}},
			ir.Evaluate{Value: item},
		}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("libraryIteratorLoop's IR changed:\ngot  %#v\nwant %#v", got, want)
	}
	if len(l.result.Locals) != 3 || l.result.Locals[1].Name != "iterator" || l.result.Locals[2].Name != "step" {
		t.Fatalf("libraryIteratorLoop's locals changed: %+v", l.result.Locals)
	}
}

// A view's dispatch reads its held local's next once, in its collection arm, after the collection
// test, and checkGeneratorRoads holds it to that: one next per collection test, so a second next of
// the local is never taken for the dispatch's own (Lower's must on #qk8rztp S2c-1, her option (b)).
// A program's xs.next() on an IterableIterator<T> parameter a generator feeds stops before that
// (the unbound-method row above); this builds the case by hand, a dispatched local with an extra
// next, so checkGeneratorRoads' compiler bug is what's shown. Mutant: the check back to
// reads == stepped + nexts, and this goes red.
func TestGeneratorRoadsRefuseAnotherNextOfADispatchedLocal(t *testing.T) {
	t.Parallel()
	held := ir.Read{Local: 0, Of: ir.Object}
	next := ir.Evaluate{Value: ir.CallClosure{Closure: ir.Property{Object: held, Name: "next", Of: ir.Closure}, Returns: ir.Object}}
	program := &ir.Program{
		Locals: []ir.Local{{Name: "iterator", Type: ir.Object, Function: -1}},
		Main: []ir.Statement{ir.Block{Closes: 1, Body: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.Undefined{Of: ir.Object}},
			ir.Evaluate{Value: ir.IsCollectionIterator{Value: held}},
			next,
			next,
		}}},
	}
	err := checkGeneratorRoads(program)
	if err == nil || !strings.Contains(err.Error(), "compiler bug") {
		t.Fatalf("got %v, want checkGeneratorRoads' compiler bug for a second next of the view's local", err)
	}
}
