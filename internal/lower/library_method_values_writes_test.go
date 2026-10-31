package lower

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// A delayed Array method writes what call or apply hands it, which lib's any[] Array.prototype never
// checks, so each value it writes into the receiver must be the receiver's element type, as the checker
// judges it (#gw3yrsn, Language's k3 to k10 on main 31fdd16d). k10 is an object pushed into a string[],
// which sanitized native read as a string and died on a null load under UBSan where Node throws. unshift
// has no adapter and stays NotYet. Mutant: delayedArrayWritesFit answering nil, under which push, fill,
// splice and concat lower again, and k10 dies natively under UBSan.
func TestDelayedArrayWritesMustFitTheElements(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"k10 push an object", `const push = Array.prototype.push;
const array = ["a", "b"];
const other = { x: 1 };
push.call(array, other);
const last = array[2] ?? "";
console.log(last.toUpperCase());`, "a delayed push writing a value that isn't the receiver's element type"},
		{"k6 push a number", `const push = Array.prototype.push;
const array = ["a", "b"];
push.call(array, 5);
console.log(String((array[2] ?? "").length));`, "a delayed push writing a value that isn't the receiver's element type"},
		{"k8 push through apply", `const push = Array.prototype.push;
const array = ["a", "b"];
push.apply(array, [5]);
console.log(String((array[2] ?? "").length));`, "a delayed push writing a value that isn't the receiver's element type"},
		{"k5 fill", `const fill = Array.prototype.fill;
const array = ["a", "b"];
fill.call(array, 5);
console.log(String((array[0] ?? "").length));`, "a delayed fill writing a value that isn't the receiver's element type"},
		{"k4 splice", `const splice = Array.prototype.splice;
const array = ["a", "b"];
splice.apply(array, [0, 0, 5]);
console.log(String((array[0] ?? "").length));`, "a delayed splice writing a value that isn't the receiver's element type"},
		{"k3 concat an array", `const concat = Array.prototype.concat;
const array = ["a", "b"];
const combined: string[] = concat.apply(array, [[5]]);
console.log(combined.join("|") + " " + String((combined[2] ?? "").length));`, "a delayed concat of an array whose elements aren't the receiver's element type"},
		{"k7 concat a value", `const concat = Array.prototype.concat;
const array = ["a", "b"];
const combined: string[] = concat.call(array, 5);
console.log(String((combined[2] ?? "").length));`, "a delayed concat of an array whose elements aren't the receiver's element type"},
		{"k9 unshift", `const unshift = Array.prototype.unshift;
const array = ["a", "b"];
unshift.call(array, 5);
console.log(String((array[0] ?? "").length));`, "delayed Array.prototype.unshift (no adapter proves"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, probe.source)
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}

// A delayed write gets the argument site's own check against the receiver's element type, not tsc's
// relation alone (Language's must on c8f64485, live on main): a string[] pushed, filled or spliced into a
// (string | number)[][] could have a number written into it through the wider view, which native read past
// the string array's end (heap-buffer-overflow under ASan), and a void function pushed where a function
// returning unknown is read is void widening. Mutant: refuseInSlot skipped for delayed writes, under which
// g1 lowers and dies natively.
func TestDelayedArrayWritesGetTheSlotCheck(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"g1 push", `const push = Array.prototype.push;
const inner: string[] = ["a"];
const outer: (string | number)[][] = [];
push.call(outer, inner);
const first = outer[0];
if (first !== undefined) first.push(5);
const last = inner[1] ?? "";
console.log(last.toUpperCase());`, "seen as (string | number)[]"},
		{"g6 fill", `const fill = Array.prototype.fill;
const inner: string[] = ["a"];
const outer: (string | number)[][] = [[]];
fill.call(outer, inner);
const first = outer[0];
if (first !== undefined) first.push(5);
console.log((inner[1] ?? "").toUpperCase());`, "seen as (string | number)[]"},
		{"g7 splice", `const splice = Array.prototype.splice;
const inner: string[] = ["a"];
const outer: (string | number)[][] = [];
splice.apply(outer, [0, 0, inner]);
const first = outer[0];
if (first !== undefined) first.push(5);
console.log((inner[1] ?? "").toUpperCase());`, "seen as (string | number)[]"},
		{"g2 void function", `const push = Array.prototype.push;
function make(): void {}
const fs: (() => unknown)[] = [];
push.call(fs, make);
console.log(String(fs.length));`, "a void-returning function at the argument site seen as () => unknown"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want Refused with %q, got %v", probe.want, err)
			}
		})
	}
}

// The whole refusal registry judges a delayed write, with the receiver's element type as the value's slot
// (refuseInSlot, #b6m00y7), so each is stopped for the reason its direct twin is, whatever rule that is:
// Language's h7, a fixed object pushed into a Record<string, number>[], whose record write crashed in map.c
// natively when a second list beside the registry forgot record storage; g1's view and g2's void widening;
// and a view inside a literal written in place, whose property's slot is carried down from the element
// type as tsc would carry it. Mutants: slotType ignoring the slot a caller set, under which h7 lowers and
// dies natively; and carriedSlot carrying nothing to a literal's property, under which the literal's view
// lowers.
func TestDelayedArrayWritesAreJudgedAsTheirDirectTwins(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, declarations, delayed, direct string }{
		{"h7 push", "const named = { a: 1 }; const records: Record<string, number>[] = [];", "push.call(records, named);", "records.push(named);"},
		{"h7 fill", "const named = { a: 1 }; const records: Record<string, number>[] = [{}];", "fill.call(records, named);", "records.fill(named);"},
		{"h7 splice", "const named = { a: 1 }; const records: Record<string, number>[] = [];", "splice.apply(records, [0, 0, named]);", "records.splice(0, 0, named);"},
		{"g1 push", "const inner: string[] = []; const outer: (string | number)[][] = [];", "push.call(outer, inner);", "outer.push(inner);"},
		{"g2 push", "function make(): void {} const fs: (() => unknown)[] = [];", "push.call(fs, make);", "fs.push(make);"},
		{"a view in a literal", "class Animal { name = 'a'; } class Dog extends Animal { bark(): string { return 'woof'; } } const dogs: Dog[] = []; const holders: { pets: Animal[] }[] = [];", "push.call(holders, { pets: dogs });", "holders.push({ pets: dogs });"},
		{"a view in an array literal", "class Animal { name = 'a'; } class Dog extends Animal { bark(): string { return 'woof'; } } const dogs: Dog[] = []; const nested: Animal[][][] = [];", "push.call(nested, [dogs]);", "nested.push([dogs]);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			prefix := "const push = Array.prototype.push; const fill = Array.prototype.fill; const splice = Array.prototype.splice;\n" + probe.declarations + "\n"
			_, delayed := lowerTypeScriptAssertionSource(t, prefix+probe.delayed+"\nconsole.log('ran');")
			_, direct := lowerTypeScriptAssertionSource(t, prefix+probe.direct+"\nconsole.log('ran');")
			if direct == nil {
				t.Fatalf("the direct twin %q lowers, so it proves nothing about the delayed write", probe.direct)
			}
			if delayed == nil || stoppedFor(delayed) != stoppedFor(direct) {
				t.Fatalf("the delayed write stopped as %v, want what its direct twin stops as: %v", delayed, direct)
			}
		})
	}
}

// stoppedFor is whether an error is a refusal or a NotYet, and what it says, without where.
func stoppedFor(err error) string {
	var refused *Refused
	if errors.As(err, &refused) {
		return "Refused: " + refused.What
	}
	var pending *NotYet
	if errors.As(err, &pending) {
		return "NotYet: " + pending.What
	}
	return err.Error()
}

// What the identity check head 1 held a delayed write to refused, the registry gives back: an object
// literal written into an array of an interface (Language's n3), a literal 5 into a (string | number)[],
// and a literal into an array of a literal union, each agreeing with Node. A part of the value whose slot
// can't be read off the element type (a computed key), and a function written in place, whose parameters
// lib's any typed, stay NotYet.
func TestDelayedArrayWritesTheRegistryAdmits(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"n3 an object literal into an interface array", `interface Point { x: number; }
const push = Array.prototype.push;
const points: Point[] = [{ x: 1 }];
push.call(points, { x: 2 });
console.log(points.map((point) => String(point.x)).join(" "));`},
		{"a literal number into a union array", `const push = Array.prototype.push;
const values: (string | number)[] = ["a", 1];
push.call(values, 5);
console.log(values.map((value) => String(value)).join("|"));`},
		{"a literal into a literal union array", `const push = Array.prototype.push;
const kinds: ("a" | "b")[] = ["b"];
push.call(kinds, "a");
console.log(kinds.join("|"));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
		})
	}
	for _, probe := range []struct{ name, source, want string }{
		{"a computed key", `interface Point { x: number; }
const push = Array.prototype.push;
const points: Point[] = [];
push.call(points, { ["x"]: 2 });
console.log(String(points.length));`, "whose own slot can't be read off the one it's written into"},
		{"a function written in place", `const push = Array.prototype.push;
const handlers: ((value: number) => string)[] = [];
push.call(handlers, (value: number) => String(value));
console.log(String(handlers.length));`, "a function written in place in a value written where the checker gives it no slot"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, probe.source)
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}

// Writes of the receiver's own element type still lower and agree with Node, through call and apply.
func TestDelayedArrayWritesOfTheElementsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"push", `const push = Array.prototype.push; const array = ["a", "b"]; push.call(array, "c"); push.apply(array, ["d"]); console.log(array.join("|"));`},
		{"fill", `const fill = Array.prototype.fill; const array = ["a", "b"]; fill.call(array, "z"); console.log(array.join("|"));`},
		{"splice", `const splice = Array.prototype.splice; const array = ["a", "b"]; splice.apply(array, [1, 0, "c"]); console.log(array.join("|"));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
		})
	}
}

// Every Array method a delayed call reaches an adapter for is in exactly one class, a writer or a reader,
// read from libraryMethodCall's own Array.prototype switch, so a name added there without a class is a
// failure here and not an unchecked write (Compiler's change on c8f64485). Mutant: a name dropped from
// delayedArrayReaders, which this reports, and which the switch's default then makes NotYet.
func TestDelayedArrayMethodsAreClassified(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), "library_method_values.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	ast.Inspect(file, func(node ast.Node) bool {
		family, isCase := node.(*ast.CaseClause)
		if !isCase || len(family.List) != 1 {
			return true
		}
		if literal, isLiteral := family.List[0].(*ast.BasicLit); !isLiteral || literal.Value != `"Array.prototype"` {
			return true
		}
		for _, statement := range family.Body {
			inner, isSwitch := statement.(*ast.SwitchStmt)
			if !isSwitch {
				continue
			}
			for _, clause := range inner.Body.List {
				for _, expression := range clause.(*ast.CaseClause).List {
					if literal, isLiteral := expression.(*ast.BasicLit); isLiteral {
						name, err := strconv.Unquote(literal.Value)
						if err != nil {
							t.Fatal(err)
						}
						names = append(names, name)
					}
				}
			}
		}
		return false
	})
	if len(names) == 0 {
		t.Fatal("found no Array.prototype adapter names in libraryMethodCall")
	}
	for _, name := range names {
		if delayedArrayWriters[name] == delayedArrayReaders[name] {
			t.Errorf("%s is a writer: %t, a reader: %t; want exactly one", name, delayedArrayWriters[name], delayedArrayReaders[name])
		}
	}
}
