package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's core.ts find (#2qvx0r2), a predicate overload over a boolean one. The read is array[i]!,
// stage 3's adaptation, since noUncheckedIndexedAccess types array[i] T | undefined. The body's
// first return is line 11 of every program below.
const predicateOverloadShapes = `interface Circle { readonly kind: "circle"; readonly radius: number }
interface Square { readonly kind: "square"; readonly side: number }
type Shape = Circle | Square;
function find<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => element is U, startIndex?: number): U | undefined;
function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined;
`

const predicateOverloadFind = predicateOverloadShapes + `function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined {
    if (array === undefined) return undefined;
    for (let i = startIndex ?? 0; i < array.length; i++) {
        const value = array[i]!;
        if (predicate(value, i)) {
            return value;
        }
    }
    return undefined;
}
`

const predicateOverloadCall = `const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
const circle = find(shapes, (x): x is Circle => x.kind === "circle");
console.log(circle === undefined ? "none" : ` + "`radius ${circle.radius}`" + `);
`

const predicateOverloadReturnFailed = "main.ts:11:"

// A predicate written in place checks its own claim, so find's narrowed result is the circle.
func TestPredicateOverloadInPlaceCallback(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, predicateOverloadFind+predicateOverloadCall, ".ts")
}

// A function declaration passed by name checks or proves its claim in the value it becomes.
func TestPredicateOverloadNamedCallback(t *testing.T) {
	t.Parallel()
	source := predicateOverloadFind + `function isCircle(x: Shape): x is Circle { return x.kind === "circle"; }
const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
const circle = find(shapes, isCircle);
console.log(circle === undefined ? "none" : ` + "`radius ${circle.radius}`" + `);
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// A callback that grows the array still answers for the local find returns, so the result agrees.
func TestPredicateOverloadCallbackPushes(t *testing.T) {
	t.Parallel()
	source := predicateOverloadFind + `const grown: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
const circle = find(grown, (x): x is Circle => {
    grown.push({ kind: "square", side: 1 });
    return x.kind === "circle";
});
console.log(circle === undefined ? "none" : ` + "`radius ${circle.radius}`" + `);
console.log(String(grown.length));
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// A return of array[i] isn't the value the callback checked, so the overload stays refused, naming it.
func TestPredicateOverloadReturnsAnotherRead(t *testing.T) {
	t.Parallel()
	source := strings.Replace(predicateOverloadFind, "            return value;", "            return array[i];", 1) + predicateOverloadCall
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "(return at ") || !strings.Contains(refused.What, predicateOverloadReturnFailed) || !strings.Contains(refused.What, "isn't the value its predicate checked)") {
		t.Fatalf("want the overload refused at the return of array[i], got %v", err)
	}
}

// A let written between the call and the return, here by the call's own later argument, isn't the
// value the callback said yes to, so the overload stays refused.
func TestPredicateOverloadLocalWrittenAfterTheCall(t *testing.T) {
	t.Parallel()
	source := strings.Replace(predicateOverloadFind, "        const value = array[i]!;\n        if (predicate(value, i)) {", "        let value = array[i]!;\n        if (predicate(value, (value = array[0]!, i))) {", 1) + predicateOverloadCall
	if !strings.Contains(source, "let value") {
		t.Fatal("the replacement didn't apply")
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, predicateOverloadReturnFailed) || !strings.Contains(refused.What, "isn't the value its predicate checked)") {
		t.Fatalf("want the overload refused at the return of the written let, got %v", err)
	}
}

// An overload promising U | undefined over a boolean callback has no claim to narrow by: refused as before.
func TestPredicateOverloadWithoutPredicate(t *testing.T) {
	t.Parallel()
	source := `function pick<T, U extends T>(array: readonly T[], predicate: (element: T) => boolean): U | undefined;
function pick<T>(array: readonly T[], predicate: (element: T) => boolean): T | undefined {
    for (const value of array) {
        if (predicate(value)) {
            return value;
        }
    }
    return undefined;
}
console.log("never");
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "cannot be served by implementation result") || strings.Contains(refused.What, "isn't the value its predicate checked") {
		t.Fatalf("want the result check's refusal unchanged, got %v", err)
	}
}

// A callback from another function's parameter could be an ambient predicate nothing checks, so the
// overload's narrower result would be its lie: predicateArguments refuses it, as it refuses any
// predicate argument whose body doesn't prove both branches.
func TestPredicateOverloadCallbackFromParameter(t *testing.T) {
	t.Parallel()
	source := predicateOverloadFind + `function firstCircle(shapes: Shape[], test: (x: Shape) => x is Circle): Circle | undefined {
    return find(shapes, test);
}
const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
const circle = firstCircle(shapes, (x): x is Circle => x.kind === "circle");
console.log(circle === undefined ? "none" : ` + "`radius ${circle.radius}`" + `);
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "an unproven predicate argument for parameter predicate") {
		t.Fatalf("want predicateArguments' refusal of the parameter passed as find's callback, got %v", err)
	}
}

// In .a a predicate value is refused, so the in-place program .ts admits still doesn't lower there.
func TestPredicateOverloadStaysOutOfA(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, predicateOverloadFind+predicateOverloadCall, ".a"); err == nil {
		t.Fatal("want the .a program not lowered, got it lowered")
	}
}

// Language's three lies on 82c1dc7c (#2qvx0r2), each returning a square typed as a circle on Node: a
// field holding a predicate reassigned to one that always says yes is an unproven argument; a callback
// tested beside another condition isn't the claim's yes; and a named predicate whose body lies stops at
// its own checked claim, loudly, at run time (rulings.md:132, :135).
func TestPredicateOverloadLanguageLies(t *testing.T) {
	t.Parallel()
	const shapes = `const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
function isCircle(x: Shape): x is Circle { return x.kind === "circle"; }
function isSquareish(x: Shape): x is Circle { return x.kind === "square"; }
`
	const show = "console.log(circle === undefined ? \"none\" : `radius ${String(circle.radius)}`);\n"
	t.Run("field reassigned", func(t *testing.T) {
		t.Parallel()
		source := predicateOverloadFind + shapes + "const checker = { test: isCircle }; checker.test = (x: Shape): x is Circle => true; const circle = find(shapes, checker.test);\n" + show
		_, err := lowerSourceAs(t, source, ".ts")
		var refused *Refused
		if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "an unproven predicate argument") {
			t.Fatalf("want an unproven predicate argument, got %v", err)
		}
	})
	t.Run("or beside the call", func(t *testing.T) {
		t.Parallel()
		source := strings.Replace(predicateOverloadFind, "if (predicate(value, i)) {", "if (predicate(value, i) || i === 0) {", 1) + shapes + "const circle = find(shapes, isCircle);\n" + show
		if !strings.Contains(source, "|| i === 0") {
			t.Fatal("the replacement didn't apply")
		}
		_, err := lowerSourceAs(t, source, ".ts")
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "isn't the value its predicate checked") {
			t.Fatalf("want the overload refused at its return, got %v", err)
		}
	})
	t.Run("lying named predicate", func(t *testing.T) {
		t.Parallel()
		source := predicateOverloadFind + shapes + "const circle = find(shapes, isSquareish);\n" + show
		program, err := lowerSourceAs(t, source, ".ts")
		if err != nil {
			t.Fatalf("want it lowered with the predicate's own check, got %v", err)
		}
		if code, stdout, stderr := brandedNative(t, program); code != 70 || !strings.Contains(stderr, "isSquareish") || strings.Contains(stdout, "radius") {
			t.Fatalf("native: exit %d, stdout %q, stderr %q; want exit 70 at isSquareish's check", code, stdout, stderr)
		}
	})
}

// tsc's findAncestor (utilitiesPublic.ts:786-799) exactly, its two overloads and its body, over a node
// with a parent chain. The callback's answer is held in a const and tested after "quit" is excluded
// (#448a51k, #2qvx0r2's rule through flow, Adamic's ruling on #fx9waqy, Nov 8).
const findAncestorShape = `interface N { readonly kind: number; readonly parent: N | undefined; }
interface I extends N { readonly kind: 80; readonly text: string; }
function findAncestor<T extends N>(node: N | undefined, callback: (element: N) => element is T): T | undefined;
function findAncestor(node: N | undefined, callback: (element: N) => boolean | "quit"): N | undefined;
function findAncestor(node: N | undefined, callback: (element: N) => boolean | "quit"): N | undefined {
    while (node) {
        const result = callback(node);
        if (result === "quit") {
            return undefined;
        }
        else if (result) {
            return node;
        }
        node = node.parent;
    }
    return undefined;
}
function isI(n: N): n is I { return n.kind === 80; }
const top: I = { kind: 80, text: "top", parent: undefined };
const middle: N = { kind: 5, parent: top };
const leaf: N = { kind: 1, parent: middle };
`

// The predicate overload narrows to I through the walk; the boolean overload's "quit" hands back
// undefined; and a callback that never says yes reaches the root. Each agrees with Node.
func TestFindAncestorOverloadResult(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, findAncestorShape+`const found = findAncestor(leaf, isI);
console.log(found === undefined ? "none" : found.text);
const quit = findAncestor(leaf, (n: N) => n.kind === 5 ? "quit" : false);
console.log(quit === undefined ? "quit" : String(quit.kind));
const never = findAncestor(leaf, (n: N): n is I => n.kind === 99);
console.log(never === undefined ? "root" : never.text);
`, ".ts")
}

// Every shape that breaks one condition keeps the result check's refusal at its return. Mutants, each in
// predicateOverloadResultTested: the boolean answer check dropped ("quit" possible goes red), the adjacent
// statement check dropped (code between goes red), the const check dropped (the let goes red), and the
// nested-write check dropped from predicateOverloadBinding (the closure goes red). Adamic's by-name mutant
// has no pin: with the const's statement right before the chain and the return first in its branch, no
// second binding of the same name can stand between the call and the return, so symbols and names agree.
func TestFindAncestorOverloadResultRefusals(t *testing.T) {
	t.Parallel()
	const call = "const found = findAncestor(leaf, isI);\nconsole.log(found === undefined ? \"none\" : found.text);\n"
	for _, probe := range []struct{ name, from, to string }{
		// The truthy answer "quit" isn't the claim's yes, so the test must exclude it first.
		{"quit still possible", "        if (result === \"quit\") {\n            return undefined;\n        }\n        else if (result) {", "        if (result) {"},
		// Code run between the call and the test could write node (Language's road 3, the second half).
		{"code between", "        const result = callback(node);\n", "        const result = callback(node);\n        console.log(\"between\");\n"},
		// The tested value is the call's own result only when it's a const (Language's condition 1). This let
		// is written only after the chain, so only the const check refuses it.
		{"a let", "        const result = callback(node);\n", "        let result = callback(node);\n"},
		// The tested call's argument must be the returned binding (Language's condition 2).
		{"another argument", "        const result = callback(node);\n", "        const other = node.parent ?? node;\n        const result = callback(other);\n"},
		// A copy of node isn't followed, so it fails closed.
		{"a copy returned", "        else if (result) {\n            return node;", "        else if (result) {\n            return copy;"},
		// A closure in the body that writes node could run inside the callback.
		{"a closure writes node", "    while (node) {", "    const reset = () => { node = undefined; };\n    if (node === undefined) {\n        reset();\n    }\n    while (node) {"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := strings.Replace(findAncestorShape, probe.from, probe.to, 1)
			if source == findAncestorShape {
				t.Fatal("the replacement didn't apply")
			}
			if probe.name == "a let" {
				source = strings.Replace(source, "        node = node.parent;\n", "        node = node.parent;\n        result = false;\n", 1)
			}
			if probe.name == "a copy returned" {
				source = strings.Replace(source, "        const result = callback(node);\n", "        const copy = node;\n        const result = callback(node);\n", 1)
			}
			_, err := lowerSourceAs(t, source+call, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "findAncestor result") || !strings.Contains(refused.What, "cannot be served by implementation result") {
				t.Fatalf("want findAncestor's result refused, got %v", err)
			}
		})
	}
}

// The return stands on the callback's claim, checked where it's made: a lying annotated arrow held in a
// const is a checked closure, so it stops at its own check on both backends before findAncestor's caller
// reads the result as an I (Language's premise on #45abpv9).
func TestFindAncestorLyingCallbackStops(t *testing.T) {
	t.Parallel()
	brandedStops(t, findAncestorShape+`const lie = (n: N): n is I => n.kind === 1;
const found = findAncestor(leaf, lie);
console.log(found === undefined ? "none" : found.text);
`, "adamic: panic: ")
}

// tsc's Node has writable fields, so a rigid T extends Node is a type a value can be widened into, and
// censusRelated refuses `undefined` returned into `T | undefined` (widened keeps undefined as itself and
// judges it against T). That was the census's stop at utilitiesPublic.ts:792, the "quit" arm's return
// (Correctness's census of 5a86e3b3), which findAncestorShape's readonly N never reached. The return of
// undefined serves the union by its undefined member (predicateOverloadUndefinedServes). Mutant: that arm
// off, and this refuses at the quit arm's return.
func TestFindAncestorOverWritableNode(t *testing.T) {
	t.Parallel()
	source := strings.Replace(findAncestorShape, "interface N { readonly kind: number; readonly parent: N | undefined; }", "interface N { kind: number; parent: N | undefined; }", 1)
	if source == findAncestorShape {
		t.Fatal("the replacement didn't apply")
	}
	lowersAndAgreesWithNodeAs(t, source+`const found = findAncestor(leaf, isI);
console.log(found === undefined ? "none" : found.text);
const quit = findAncestor(leaf, (n: N) => n.kind === 5 ? "quit" : false);
console.log(quit === undefined ? "quit" : String(quit.kind));
`, ".ts")
}
