package lower

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// handNumbers fills a Map and a Set of numbers whose values() and keys() feed the probes below.
const handNumbers = "const sizes = new Map<string, number>();\nconst odd = new Set<number>();\nfor (let index = 0; index < 3; index++) { sizes.set(`size ${String(index)}`, index + 7); odd.add(index * 2 + 1); }\n"

// handFirst is a function whose IterableIterator<number> parameter is stepped by hand, its value read
// where done has narrowed the result.
const handFirst = "function first(numbers: IterableIterator<number>): number { const status = numbers.next(); if (status.done) { return -1; } return status.value; }\n"

// A hand-driven next() through an IterableIterator<T> parameter lowers when every origin is a
// collection's iterator (#qk8rztp S2c-2a, Adamic's (a) and (b)): checker.ts:21636's loop, a status
// kept across a later next(), and the `| undefined` form. internal/oracle's hand_next.a holds them
// to Node.
func TestHandDrivenNextLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"one next, value read where done narrows it", handNumbers + handFirst + "console.log(String(first(sizes.values())) + String(first(odd.keys())));"},
		{"checker.ts:21636's loop", handNumbers + "function total(numbers: IterableIterator<number>): number { let sum = 0; for (let status = numbers.next(); !status.done; status = numbers.next()) { sum += status.value; } return sum; }\nconsole.log(String(total(sizes.values())));"},
		{"a status kept across a later next", handNumbers + "function two(numbers: IterableIterator<number>): number { const a = numbers.next(); const b = numbers.next(); return (a.done ? 0 : a.value) + (b.done ? 0 : b.value); }\nconsole.log(String(two(odd.values())));"},
		{"the | undefined form, tested first", handNumbers + "function first(numbers: IterableIterator<number> | undefined): number { if (numbers === undefined) { return -1; } const status = numbers.next(); return status.done ? -2 : status.value; }\nconsole.log(String(first(sizes.values())));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
		})
	}
}

// Every other road stops (Adamic's ruling on #qk8rztp S2c-2a). Not yet: next(x), an origin the proof
// can't name (the parameter of an arrow kept in a local, whose callers the graph doesn't see), and an
// optional call of next or of return(), as any call through ?. is. Still refused as main refuses it:
// a detached next, by the unbound-method rule. The value read unnarrowed, or in the
// done arm, is any, which stops at the any rules (Adamic's void pin). Mutants
// (internal/oracle/testdata/iterable_views/mutants.md): the origin check dropped admits the arrow
// row; the next(x) stop dropped admits its row.
func TestHandDrivenNextStopsOnEveryOtherRoad(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"next(x)", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next(1); return status.done ? -1 : status.value; }\nconsole.log(String(first(sizes.values())));", "next(x) through a view"},
		{"an arrow's parameter, its callers unseen", handNumbers + "const first = (numbers: IterableIterator<number>): number => { const status = numbers.next(); return status.done ? -1 : status.value; };\nconsole.log(String(first(sizes.values())));", "a view over an iterable of unproven origin"},
		{"an optional call of next", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next?.(); return status === undefined || status.done ? -1 : status.value; }\nconsole.log(String(first(sizes.values())));", "a call through ?. (an optional call)"},
		{"return()", handNumbers + "function first(numbers: IterableIterator<number>): number { numbers.return?.(); return 0; }\nconsole.log(String(first(sizes.values())));", "a call through ?. (an optional call)"},
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
	// A for-of beside a hand-driven next over the same parameter, in .a and in .ts (Compiler's must).
	for _, extension := range []string{".a", ".ts"} {
		mixed := handNumbers + "function both(numbers: IterableIterator<number>): number { let sum = 0; const first = numbers.next(); if (!first.done) { sum += first.value; } for (const value of numbers) { sum += value; } return sum; }\nconsole.log(String(both(sizes.values())));"
		t.Run("a for-of beside a next in "+extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, mixed, extension)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "a for-of beside a hand-driven next() over the same parameter") {
				t.Fatalf("got %v, want a NotYet saying a for-of beside a hand-driven next()", err)
			}
		})
	}
	for _, probe := range []struct{ name, source, reason string }{
		{"a detached next", handNumbers + "function first(numbers: IterableIterator<number>): number { const step = numbers.next; const status = step(); return status.done ? -1 : status.value; }\nconsole.log(String(first(sizes.values())));", "a method read as a value"},
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
	// The value unnarrowed, or in the done arm, is any. Which any rule stops it, refused or not yet,
	// is the any rules' to say; that it stops, and names any, is the pin.
	for _, probe := range []struct{ name, source string }{
		{"value unnarrowed", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next(); return status.value; }\nconsole.log(String(first(sizes.values())));"},
		{"value in the done arm", handNumbers + "function first(numbers: IterableIterator<number>): number { const status = numbers.next(); if (status.done) { return status.value; } return 0; }\nconsole.log(String(first(sizes.values())));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), "any") {
				t.Fatalf("got %v, want it stopped as a value of type any", err)
			}
		})
	}
}

// handCounted is a number generator, handed to a hand-driven parameter by its direct call.
const handCounted = "function* counted(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n"

// iteratorNexts is every IteratorNext in a lowered program.
func iteratorNexts(program *ir.Program) []ir.IteratorNext {
	var found []ir.IteratorNext
	visit := func(node any) bool {
		if next, isNext := node.(ir.IteratorNext); isNext {
			found = append(found, next)
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return found
}

// A hand-driven next() through an IterableIterator<T> parameter a generator's direct call feeds
// lowers (#qk8rztp S2c-2b), each next rewritten into IteratorNext: a number generator, the narrowed
// `| undefined` form, a generator and a Map's values() into one parameter at two call sites, whose
// nexts say a collection's iterator feeds them too, and Generator<string | undefined>, a reference
// that may be undefined (Language's pin). internal/oracle's hand_next_generator*.a hold them to Node.
// Mutant (internal/oracle/testdata/iterable_views/mutants.md): the exemption of a next from the hand
// count dropped, and every row stops, read as many times as it has nexts.
func TestHandDrivenNextOverAGeneratorLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		nexts        int
		collection   bool
	}{
		{"a number generator, one next", handCounted + handFirst + "console.log(String(first(counted(3))));", 1, false},
		{"checker.ts:21636's loop", handCounted + "function total(numbers: IterableIterator<number>): number { let sum = 0; for (let status = numbers.next(); !status.done; status = numbers.next()) { sum += status.value; } return sum; }\nconsole.log(String(total(counted(4))));", 2, false},
		{"the | undefined form, tested first", handCounted + "function first(numbers: IterableIterator<number> | undefined): number { if (numbers === undefined) { return -1; } const status = numbers.next(); return status.done ? -2 : status.value; }\nconsole.log(String(first(counted(2))));", 1, false},
		{"a generator and a Map's values() at two call sites", handCounted + handNumbers + handFirst + "console.log(String(first(counted(3))) + String(first(sizes.values())));", 1, true},
		{"Generator<string | undefined>", "function* maybe(limit: number): Generator<string | undefined, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index % 2 === 0 ? `name ${String(index)}` : undefined; } }\nfunction names(values: IterableIterator<string | undefined>): string { let all = ''; for (let status = values.next(); !status.done; status = values.next()) { all += status.value === undefined ? 'undefined;' : `${status.value};`; } return all; }\nconsole.log(names(maybe(3)));", 2, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			nexts := iteratorNexts(program)
			if len(nexts) != probe.nexts {
				t.Fatalf("%d IteratorNexts, want %d, one per next", len(nexts), probe.nexts)
			}
			for _, next := range nexts {
				if next.Collection != probe.collection || !program.Functions[next.Function].Generator || next.Site == "" {
					t.Fatalf("got %+v, want the generator's step, Collection %t, at its site", next, probe.collection)
				}
			}
		})
	}
}

// A view whose every origin is a collection's iterator keeps S2c-2a's CallClosure over next, never
// IteratorNext (Lower's and Compiler's ruling 2 on #qk8rztp S2c-2b), so hand_next.a, which no
// generator feeds, lowers as it did and its emitted code doesn't move. The proof script diffs its
// emitted C and JavaScript against the head before this one; this holds the IR. Mutant: the rewrite
// taken for every proven hand view, and this finds IteratorNexts.
func TestHandDrivenNextNoGeneratorFeedsKeepsItsCall(t *testing.T) {
	t.Parallel()
	// Reads: internal/oracle/testdata/iterable_views/hand_next.a
	source, err := os.ReadFile("../oracle/testdata/iterable_views/hand_next.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	if nexts := iteratorNexts(program); len(nexts) != 0 {
		t.Fatalf("hand_next.a holds %d IteratorNexts, where no generator feeds any of its parameters", len(nexts))
	}
	calls := 0
	for _, function := range program.Functions {
		for _, parameter := range function.Parameters {
			walk(function.Body, func(node any) bool {
				if _, isNext := handNextReceiver(node, parameter); isNext {
					calls++
				}
				return true
			})
		}
	}
	// elaborate, firstLate, firstTwo and keptNames two each, exhausted three, firstOrMissing one.
	if calls != 12 {
		t.Fatalf("hand_next.a holds %d of S2c-2a's CallClosures over next, want 12", calls)
	}
}

// What a generator's hand-driven next() doesn't build stops not yet, each in its words (#qk8rztp
// S2c-2b): a yield of `boolean | undefined` or `number | undefined`, which generatorSignature stops
// as no one present slot before handNextHolds is asked (Adamic's and Language's pins); a parameter
// read beside its nexts, passed on to forward (Lower's row); and two generator functions into one
// parameter, whose steps would name two. A narrowed field, `holder.it`, is no hand next at all: its
// receiver isn't a parameter's name, so it stops before the count (Lower's row; its matcher mutant
// is held by TestHandNextReceiverIsTheParametersOwnRead). Mutants
// (internal/oracle/testdata/iterable_views/mutants.md): the exemption widened to any call admits the
// forward row.
func TestHandDrivenNextOverAGeneratorStops(t *testing.T) {
	t.Parallel()
	const halved = "function* halved(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index / 2; } }\n"
	for _, probe := range []struct{ name, source, reason string }{
		{"Generator<boolean | undefined>", "function* flags(limit: number): Generator<boolean | undefined, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index === 1 ? undefined : index > 1; } }\nfunction first(values: IterableIterator<boolean | undefined>): string { const status = values.next(); return status.done ? 'done' : String(status.value); }\nconsole.log(first(flags(3)));", "yielding boolean | undefined (each yielded value must be one present slot)"},
		{"Generator<number | undefined>", "function* maybe(limit: number): Generator<number | undefined, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index === 1 ? undefined : index; } }\nfunction first(values: IterableIterator<number | undefined>): string { const status = values.next(); return status.done ? 'done' : String(status.value); }\nconsole.log(first(maybe(3)));", "yielding number | undefined (each yielded value must be one present slot)"},
		{"forward(iterator) beside a next", handCounted + "function forward(numbers: IterableIterator<number>): number { return typeof numbers === 'object' ? 1 : 0; }\nfunction first(numbers: IterableIterator<number>): number { const status = numbers.next(); return forward(numbers) + (status.done ? -1 : status.value); }\nconsole.log(String(first(counted(3))));", "passed into the parameter numbers, read 1 times where its no-argument nexts and undefined tests are its only uses"},
		{"two generator functions", handCounted + halved + handFirst + "console.log(String(first(counted(3)) + first(halved(3))));", "generators of more than one function"},
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
	t.Run("holder.it narrowed, then next()", func(t *testing.T) {
		t.Parallel()
		source := handCounted + "interface Holder { readonly it: IterableIterator<number> | undefined }\nfunction first(holder: Holder): number { if (holder.it === undefined) { return -1; } const status = holder.it.next(); return status.done ? -2 : status.value; }\nconsole.log(String(first({ it: counted(3) })));"
		if _, err := lowerSource(t, source); err == nil {
			t.Fatal("holder.it's next() lowered, where its receiver is no parameter's own read")
		}
	})
}

// handNextReceiver takes a next() over the parameter's own read, bare or under exactly one Defined,
// and nothing else (Lower's ruling 1 on #qk8rztp S2c-2b). The holder.it shape, Defined over a field
// of the parameter, can't come from source, since a next whose receiver isn't a parameter's name is
// no hand next (handDrivenNext), so the matcher is held here on built IR. Mutant: the Defined arm
// takes any operand, and the holder.it row is taken as a next of the parameter, which the count would
// then not count.
func TestHandNextReceiverIsTheParametersOwnRead(t *testing.T) {
	t.Parallel()
	const parameter = 3
	next := func(receiver ir.Expression) ir.CallClosure {
		return ir.CallClosure{Closure: ir.Property{Object: receiver, Name: "next", Method: true}, Returns: ir.Object}
	}
	read := ir.Read{Local: parameter, Of: ir.Object}
	for _, probe := range []struct {
		name string
		node ir.CallClosure
		want bool
	}{
		{"a bare read", next(read), true},
		{"the narrowing's Defined over the read", next(ir.Defined{Value: read}), true},
		{"holder.it narrowed", next(ir.Defined{Value: ir.Property{Object: read, Name: "it"}}), false},
		{"a Defined over a Defined", next(ir.Defined{Value: ir.Defined{Value: read}}), false},
		{"another local", next(ir.Read{Local: parameter + 1, Of: ir.Object}), false},
		{"next(x)", ir.CallClosure{Closure: ir.Property{Object: read, Name: "next", Method: true}, Arguments: []ir.Expression{ir.NumberConstant{Value: 1}}, Returns: ir.Object}, false},
		{"next?.()", ir.CallClosure{Closure: ir.Property{Object: read, Name: "next", Method: true, Optional: true}, Returns: ir.Object}, false},
		{"return()", ir.CallClosure{Closure: ir.Property{Object: read, Name: "return", Method: true}, Returns: ir.Object}, false},
	} {
		if _, got := handNextReceiver(probe.node, parameter); got != probe.want {
			t.Errorf("%s: handNextReceiver says %t, want %t", probe.name, got, probe.want)
		}
	}
}

// A hand-driven next()'s result holds a generator's yield only as a number, a boolean or a reference
// (Adamic's and Language's ruling 4 on #qk8rztp S2c-2b), an allowlist held here over every
// representation the IR names. Mutant: MaybeBoolean admitted, and its row goes red.
func TestHandNextHoldsOnlyNumbersBooleansAndReferences(t *testing.T) {
	t.Parallel()
	for name, of := range ir.TypeTags() {
		want := of == ir.Number || of == ir.Boolean || of == ir.String || of == ir.Object || of == ir.Array || of == ir.Map ||
			of == ir.Record || of == ir.Closure || of == ir.Union || of == ir.Weak || of == ir.Promise || of == ir.Generator ||
			of.IsTypedArray()
		if got := handNextHolds(of); got != want {
			t.Errorf("%s: handNextHolds says %t, want %t", name, got, want)
		}
	}
	for _, of := range []ir.Type{ir.MaybeBoolean, ir.MaybeNumber, ir.NullRepresentation, ir.UndefinedRepresentation, 0} {
		if handNextHolds(of) {
			t.Errorf("handNextHolds admits %d, which no step's result holds", of)
		}
	}
}
