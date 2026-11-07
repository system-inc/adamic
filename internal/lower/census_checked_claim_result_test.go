package lower

import (
	"errors"
	"strings"
	"testing"
)

// The census serves an overload whose result is a function claiming `p is R` from an implementation
// whose result relates with the claim erased to boolean, in .ts, where every call through the promised
// signature checks the claim (censusCheckedClaimResult, Adamic's ruling on #1w5bve2).

// censusClaimNodes is a node family in tsc's shape, with find's callback a checked claim. It holds no
// parent chain, so it isn't a graph program, whose checked-view adapters are refused (cycle-capable).
const censusClaimNodes = `interface N { readonly kind: number; }
interface A extends N { readonly kind: 1; readonly a: string; }
interface B extends N { readonly kind: 2; readonly b: string; }
interface C extends N { readonly kind: 3; readonly c: string; }
function isA(n: N): n is A { return n.kind === 1; }
export function find<T extends N>(nodes: readonly N[], callback: (element: N) => element is T): T | undefined {
  for (const node of nodes) {
    if (callback(node)) { return node; }
  }
  return undefined;
}
`

// censusClaimOr is or with an implementation whose binders rename the overload's by position, holding
// its predicates in a rest of plain callbacks, as Optimizer's by-position or on #1w5bve2 does.
const censusClaimOr = `export function or<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2;
export function or<P, R1 extends P, R2 extends P>(...fs: ((p: P) => boolean)[]): (p: P) => boolean {
    return (p: P): boolean => {
        for (const f of fs) {
            if (f(p)) {
                return true;
            }
        }
        return false;
    };
}
`

const censusClaimHonestB = "function isB(n: N): n is B { return n.kind === 2; }\n"

const censusClaimCalls = `const root: C = { kind: 3, c: "root" };
const middle: B = { kind: 2, b: "middle" };
const leaf: N = { kind: 9 };
const two = find([leaf, middle, root], or(isA, isB));
console.log(two === undefined ? "none" : String(two.kind));
const none = find([leaf, root], or(isA, isB));
console.log(none === undefined ? "none" : String(none.kind));
`

// censusClaimNextStop is where or by position stops once the census serves its checked claim: any or
// is a graph program, since its closure holds an array of closures, and a checked view in one is
// refused. It's the wall after this rule, not this rule's, so these pins name it, and each turns to
// agreement with Node when it lifts. The end-to-end lie through tsc's own or is the done-when of
// whichever change makes or lower (Adamic's ruling on #1w5bve2).
const censusClaimNextStop = "main.ts:12:1: Adamic 0.1 refuses a checked-view adapter in a graph program; keep checked views and dynamic shapes outside graph adoption (adamic/cycle-capable)"

// or by position, honest and lying, gets past the census's result check to that wall, and Node's
// answers are pinned for the day it lifts: "2" and "none" honest, and the lie read as undefined.
func TestCensusCheckedClaimResultNextStop(t *testing.T) {
	t.Parallel()
	lie := censusClaimNodes + "function isB(n: N): n is B { return n.kind === 2 || n.kind === 3; }\n" + censusClaimOr + `const root: C = { kind: 3, c: "root" };
const leaf: N = { kind: 9 };
const found = find([leaf, root], or(isA, isB));
if (found !== undefined && found.kind !== 1) { console.log(found.b); }
`
	for _, probe := range []struct{ name, source, node string }{
		{"honest", censusClaimNodes + censusClaimHonestB + censusClaimOr + censusClaimCalls, "2\nnone\n"},
		{"lie", lie, "undefined\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if node := runAgreementNode(t, writeAgreementSource(t, "main.ts", probe.source)); node.code != 0 || string(node.stdout) != probe.node {
				t.Fatalf("Node: exit %d, stdout %q; want 0 and %q", node.code, node.stdout, probe.node)
			}
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), censusClaimNextStop) {
				t.Fatalf("got %v, want a refusal ending %q", err, censusClaimNextStop)
			}
		})
	}
}

// Each of these keeps the census's result refusal: .a, where no claim is checked, and a promised
// intersection, which is more than the one call signature the erasure reads. The census judges the
// declaration, so neither twin calls through its result. An asserts result isn't here: the census's
// ordinary relation serves it before this rule is asked (Language's a01 and a02 on #1w5bve2).
func TestCensusCheckedClaimResultStaysRefused(t *testing.T) {
	t.Parallel()
	tagged := `export function tag<P, R extends P>(f1: (p1: P) => p1 is R): ((p: P) => p is R) & { tag: string };
export function tag<P, R extends P>(...fs: ((p: P) => boolean)[]): ((p: P) => boolean) & { tag: string } {
    return Object.assign((p: P): boolean => {
        for (const f of fs) {
            if (f(p)) {
                return true;
            }
        }
        return false;
    }, { tag: "t" });
}
const tagB = tag(isB);
console.log(tagB === undefined ? "none" : "made");
`
	for _, probe := range []struct{ name, source, extension string }{
		// or's declarations alone. In .a the overload's predicate parameters are refused first
		// (adamic/no-type-predicate), so the census's .a exclusion is defense behind that.
		{"a", censusClaimOr + "console.log(\"made\");\n", ".a"},
		{"intersection", censusClaimNodes + censusClaimHonestB + tagged, ".ts"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			t.Logf("%s stops at: %v", probe.name, err)
			var refused *Refused
			want := "cannot be served by implementation result"
			if probe.extension == ".a" {
				want = "(adamic/no-type-predicate)"
			}
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), want) {
				t.Fatalf("want the census's result refusal, got %v", err)
			}
		})
	}
}
