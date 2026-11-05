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
console.log(grown.length);
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

// A callback from another function's parameter could be an ambient predicate nothing checks: NotYet at the argument.
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
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "a predicate overload's callback from a parameter whose claim nothing checks" {
		t.Fatalf("want NotYet at the parameter passed as find's callback, got %v", err)
	}
}

// In .a a predicate value is refused, so the in-place program .ts admits still doesn't lower there.
func TestPredicateOverloadStaysOutOfA(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, predicateOverloadFind+predicateOverloadCall, ".a"); err == nil {
		t.Fatal("want the .a program not lowered, got it lowered")
	}
}
