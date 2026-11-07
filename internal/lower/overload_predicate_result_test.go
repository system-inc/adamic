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

const predicateOverloadFind = predicateOverloadShapes +
	`function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined {
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
func TestPredicateOverloadInPlaceCallback(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNodeAs(test, predicateOverloadFind+predicateOverloadCall, ".ts")
}

// A function declaration passed by name checks or proves its claim in the value it becomes.
func TestPredicateOverloadNamedCallback(test *testing.T) {
	test.Parallel()
	source := predicateOverloadFind + `function isCircle(x: Shape): x is Circle { return x.kind === "circle"; }
const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
const circle = find(shapes, isCircle);
console.log(circle === undefined ? "none" : ` + "`radius ${circle.radius}`" + `);
`
	lowersAndAgreesWithNodeAs(test, source, ".ts")
}

// A callback that grows the array still answers for the local find returns, so the result agrees.
func TestPredicateOverloadCallbackPushes(test *testing.T) {
	test.Parallel()
	source := predicateOverloadFind +
		`const grown: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
const circle = find(grown, (x): x is Circle => {
    grown.push({ kind: "square", side: 1 });
    return x.kind === "circle";
});
console.log(circle === undefined ? "none" : ` +
		"`radius ${circle.radius}`" +
		`);
console.log(String(grown.length));
`
	lowersAndAgreesWithNodeAs(test, source, ".ts")
}

// A return of array[i] isn't the value the callback checked, so the overload stays refused, naming it.
func TestPredicateOverloadReturnsAnotherRead(test *testing.T) {
	test.Parallel()
	source := strings.Replace(predicateOverloadFind, "            return value;", "            return array[i];", 1) +
		predicateOverloadCall
	_, err := lowerSourceAs(test, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) ||
		!strings.Contains(refused.What, "(return at ") ||
		!strings.Contains(refused.What, predicateOverloadReturnFailed) ||
		!strings.Contains(refused.What, "isn't the value its predicate checked)") {
		test.Fatalf("want the overload refused at the return of array[i], got %v", err)
	}
}

// A let written between the call and the return, here by the call's own later argument, isn't the
// value the callback said yes to, so the overload stays refused.
func TestPredicateOverloadLocalWrittenAfterTheCall(test *testing.T) {
	test.Parallel()
	source := strings.Replace(
		predicateOverloadFind,
		"        const value = array[i]!;\n        if (predicate(value, i)) {",
		"        let value = array[i]!;\n        if (predicate(value, (value = array[0]!, i))) {",
		1,
	) + predicateOverloadCall
	if !strings.Contains(source, "let value") {
		test.Fatal("the replacement didn't apply")
	}
	_, err := lowerSourceAs(test, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) ||
		!strings.Contains(refused.What, predicateOverloadReturnFailed) ||
		!strings.Contains(refused.What, "isn't the value its predicate checked)") {
		test.Fatalf("want the overload refused at the return of the written let, got %v", err)
	}
}

// An overload promising U | undefined over a boolean callback has no claim to narrow by: refused as before.
func TestPredicateOverloadWithoutPredicate(test *testing.T) {
	test.Parallel()
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
	_, err := lowerSourceAs(test, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) ||
		!strings.Contains(refused.What, "cannot be served by implementation result") ||
		strings.Contains(refused.What, "isn't the value its predicate checked") {
		test.Fatalf("want the result check's refusal unchanged, got %v", err)
	}
}

// A callback from another function's parameter could be an ambient predicate nothing checks, so the
// overload's narrower result would be its lie: predicateArguments refuses it, as it refuses any
// predicate argument whose body doesn't prove both branches.
func TestPredicateOverloadCallbackFromParameter(test *testing.T) {
	test.Parallel()
	source := predicateOverloadFind +
		`function firstCircle(shapes: Shape[], test: (x: Shape) => x is Circle): Circle | undefined {
    return find(shapes, test);
}
const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
const circle = firstCircle(shapes, (x): x is Circle => x.kind === "circle");
console.log(circle === undefined ? "none" : ` +
		"`radius ${circle.radius}`" +
		`);
`
	_, err := lowerSourceAs(test, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) ||
		!strings.HasPrefix(refused.What, "an unproven predicate argument for parameter predicate") {
		test.Fatalf("want predicateArguments' refusal of the parameter passed as find's callback, got %v", err)
	}
}

// In .a a predicate value is refused, so the in-place program .ts admits still doesn't lower there.
func TestPredicateOverloadStaysOutOfA(test *testing.T) {
	test.Parallel()
	if _, err := lowerSourceAs(test, predicateOverloadFind+predicateOverloadCall, ".a"); err == nil {
		test.Fatal("want the .a program not lowered, got it lowered")
	}
}

// Language's three lies on 82c1dc7c (#2qvx0r2), each returning a square typed as a circle on Node: a
// field holding a predicate reassigned to one that always says yes is an unproven argument; a callback
// tested beside another condition isn't the claim's yes; and a named predicate whose body lies stops at
// its own checked claim, loudly, at run time (rulings.md:132, :135).
func TestPredicateOverloadLanguageLies(test *testing.T) {
	test.Parallel()
	const shapes = `const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
function isCircle(x: Shape): x is Circle { return x.kind === "circle"; }
function isSquareish(x: Shape): x is Circle { return x.kind === "square"; }
`
	const show = "console.log(circle === undefined ? \"none\" : `radius ${String(circle.radius)}`);\n"
	test.Run("field reassigned", func(test *testing.T) {
		test.Parallel()
		source := predicateOverloadFind +
			shapes +
			"const checker = { test: isCircle }; checker.test = (x: Shape): x is Circle => true; const circle = find(shapes, checker.test);\n" +
			show
		_, err := lowerSourceAs(test, source, ".ts")
		var refused *Refused
		if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "an unproven predicate argument") {
			test.Fatalf("want an unproven predicate argument, got %v", err)
		}
	})
	test.Run("or beside the call", func(test *testing.T) {
		test.Parallel()
		source := strings.Replace(
			predicateOverloadFind,
			"if (predicate(value, i)) {",
			"if (predicate(value, i) || i === 0) {",
			1,
		) + shapes + "const circle = find(shapes, isCircle);\n" + show
		if !strings.Contains(source, "|| i === 0") {
			test.Fatal("the replacement didn't apply")
		}
		_, err := lowerSourceAs(test, source, ".ts")
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "isn't the value its predicate checked") {
			test.Fatalf("want the overload refused at its return, got %v", err)
		}
	})
	test.Run("lying named predicate", func(test *testing.T) {
		test.Parallel()
		source := predicateOverloadFind + shapes + "const circle = find(shapes, isSquareish);\n" + show
		program, err := lowerSourceAs(test, source, ".ts")
		if err != nil {
			test.Fatalf("want it lowered with the predicate's own check, got %v", err)
		}
		if code, stdout, stderr := brandedNative(
			test,
			program,
		); code != 70 || !strings.Contains(stderr, "isSquareish") || strings.Contains(stdout, "radius") {
			test.Fatalf(
				"native: exit %d, stdout %q, stderr %q; want exit 70 at isSquareish's check",
				code,
				stdout,
				stderr,
			)
		}
	})
}
