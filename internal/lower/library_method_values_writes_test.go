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
// returning unknown is read is void widening. Mutant: slotRelation skipped for delayed writes, under which
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
console.log(String(fs.length));`, "a void-returning function written where"},
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

// Until the refusal registry itself judges a delayed write, the value's widened type must be the element
// type itself, so a rule slotRelation doesn't name can't be skipped (Adamic's ruling on ce1280d9).
// Language's h7: a fixed object pushed into a Record<string, number>[], whose record write then crashed in
// map.c natively, and its fill and splice twins. Mutant: the identity check dropped, under which h7 lowers
// and dies natively.
func TestDelayedArrayWritesAreTheElementTypeItself(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, write string }{
		{"h7 push", "push.call(records, named);"},
		{"h7 fill", "fill.call(records, named);"},
		{"h7 splice", "splice.apply(records, [0, 0, named]);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := `const push = Array.prototype.push; const fill = Array.prototype.fill; const splice = Array.prototype.splice;
const named = { a: 1 };
const records: Record<string, number>[] = [];
` + probe.write + `
const first = records[0];
if (first !== undefined) first["b"] = 2;
console.log(String(Object.keys(named).length));`
			_, err := lowerTypeScriptAssertionSource(t, source)
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), "writing a value whose type isn't the receiver's element type itself") {
				t.Fatalf("want the identity NotYet, got %v", err)
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
