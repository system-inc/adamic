package lower

import (
	"errors"
	"strings"
	"testing"
)

// The census's first witness renames an implementation's binders by position. When that is
// refused, a second reads the whole signature (censusInferWitness), and it may only find a
// better witness: every constraint, parameter and result check still has to hold for it, and a
// refused second witness returns the first refusal in its own words (#yq1xjde).

// censusWitnessSwap names its binders in the other order from the implementation, so position
// renames X to A where X stands for B. Its first overload is proven only by the second witness,
// and its second by the first, so a first attempt that left state behind would show there.
const censusWitnessSwap = `function swap<A, B>(b: B, a: A): B;
function swap<X, Y>(x: X, y: Y): X;
function swap<X, Y>(x: X, y: Y): X { return x; }
const first: string = swap<number, string>('kept', 4);
const second: number = swap<number, string>(7, 'dropped');
console.log(first + ' ' + String(second));
`

// censusWitnessOr is tsc's core.ts or, two of its overloads over its rest implementation, where T
// is the arguments' tuple and P the predicate's subject: the second witness reads T as [P].
const censusWitnessOr = `interface N { readonly kind: number; }
interface A extends N { readonly kind: 1; }
interface B extends N { readonly kind: 2; }
function isA(n: N): n is A { return n.kind === 1; }
function isB(n: N): n is B { return n.kind === 2; }
export function or<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2;
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U;
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U {
    return (...args) => {
        let lastResult: U;
        for (const f of fs) {
            lastResult = f(...args);
            if (lastResult) {
                return lastResult;
            }
        }
        return lastResult!;
    };
}
const either = or(isA, isB);
const n: N = { kind: 2 };
console.log(either(n) ? 'yes' : 'no');
`

func TestCensusOverloadSecondWitness(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run("swap"+extension, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, censusWitnessSwap, extension); err != nil {
				t.Fatalf("want the second witness to prove swap's first overload, got %v", err)
			}
		})
		t.Run("or"+extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, censusWitnessOr, extension)
			if err != nil && strings.Contains(err.Error(), "cannot satisfy implementation parameter") {
				t.Fatalf("want or past the census's binder check, got %v", err)
			}
			t.Logf("or%s stops at: %v", extension, err)
		})
	}
}

// Each of these is refused by both witnesses, and must stay refused in the first one's words.
func TestCensusOverloadSecondWitnessOnlyProves(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stop string }{
		// The read witness T = [P], U = Q satisfies T's constraint and serves f1, but the
		// implementation's number can't serve the overload's 1.
		{"result", `function wrap<P, Q>(f1: (p1: P) => Q): (p: P) => 1;
function wrap<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => number { return (...args) => fs.length; }
const wrapped = wrap((value: number) => value);
console.log(String(wrapped(1)));
`, "main.ts:1:15: Adamic 0.1 refuses overload 1 of wrap type parameter P cannot satisfy implementation parameter T; make implementation constraints accept every type admitted by the overload"},
		// V has no evidence in either witness, so it takes bottom, and the result still has to be
		// served: a second attempt that accepted a binder with no witness, skipping the checks after
		// it, would lower this (Adamic's condition: no witness found is never a proof).
		{"no_evidence", `function wrap<P, Q>(f1: (p1: P) => Q): (p: P) => 1;
function wrap<T extends unknown[], U, V>(...fs: ((...args: T) => U)[]): (...args: T) => number { return (...args) => fs.length; }
const wrapped = wrap((value: number) => value);
console.log(String(wrapped(1)));
`, ""},
		// Language's first: the body writes the packed tuple through T's wider view (args.unshift(42)),
		// so isA would be handed a number. Not only a spread, so invariance stands (Adamic's ruling).
		{"written", `interface N { readonly kind: number; readonly name: string; }
interface A extends N { readonly kind: 1; }
function isA(n: N): n is A { return n.name.length > 0 && n.kind === 1; }
export function first<P, R1 extends P>(f1: (p1: P) => p1 is R1): (p: P) => p is R1;
export function first<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U;
export function first<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U {
  return (...args: T): U => {
    args.unshift(42);
    const f = fs[0];
    if (f === undefined) { throw new Error("none"); }
    return f(...args);
  };
}
const test = first(isA);
console.log(test({ kind: 1, name: "a" }) ? "yes" : "no");
`, "main.ts:4:23: Adamic 0.1 refuses overload 1 of first type parameter P cannot satisfy implementation parameter T; make implementation constraints accept every type admitted by the overload"},
		// Compiler's twin: T types a second parameter beside the rest, a value no call packed, so
		// invariance stands though the body never touches the tuple.
		{"twice", `export function pair<P, Q>(f1: (p1: P, p2: P) => Q): (p: P) => Q;
export function pair<T extends unknown[], U>(...fs: ((a: T, ...args: T) => U)[]): (...args: T) => U {
  return (...args: T): U => { throw new Error("none"); };
}
const added = pair((x: number, y: number) => x + y);
console.log(added === undefined ? "none" : "made");
`, "main.ts:1:22: Adamic 0.1 refuses overload 1 of pair type parameter P cannot satisfy implementation parameter T; make implementation constraints accept every type admitted by the overload"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			t.Logf("%s stops at: %v", probe.name, err)
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "refuses overload 1 of ") || !strings.HasSuffix(err.Error(), probe.stop) {
				t.Fatalf("want a refusal ending %q, got %v", probe.stop, err)
			}
		})
	}
}
