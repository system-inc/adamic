package lower

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// tsc's core.ts filter (:261 to :297 at the stage 3 pin, J-1ncez5n), its eight overloads over one generic
// implementation, verbatim but for each declaration's export and @internal marker, and its two element
// reads array[i]!, stage 3's adaptation (stage3/adapt/32-indexed-reads-program), since
// noUncheckedIndexedAccess types array[i] T | undefined. Overloads 2 and 6 promise a mutable T[] over
// the implementation's readonly T[] | undefined, which certificate D serves (#4z7y6pq); overloads 3 and 7,
// a predicate's readonly U[], it doesn't.
const overloadCertificateFilter = `function filter<T, U extends T>(array: T[], f: (x: T) => x is U): U[];
function filter<T>(array: T[], f: (x: T) => boolean): T[];
function filter<T, U extends T>(array: readonly T[], f: (x: T) => x is U): readonly U[];
function filter<T>(array: readonly T[], f: (x: T) => boolean): readonly T[];
function filter<T, U extends T>(array: T[] | undefined, f: (x: T) => x is U): U[] | undefined;
function filter<T>(array: T[] | undefined, f: (x: T) => boolean): T[] | undefined;
function filter<T, U extends T>(array: readonly T[] | undefined, f: (x: T) => x is U): readonly U[] | undefined;
function filter<T>(array: readonly T[] | undefined, f: (x: T) => boolean): readonly T[] | undefined;
function filter<T>(array: readonly T[] | undefined, f: (x: T) => boolean): readonly T[] | undefined {
    if (array !== undefined) {
        const len = array.length;
        let i = 0;
        while (i < len && f(array[i]!)) i++;
        if (i < len) {
            const result = array.slice(0, i);
            i++;
            while (i < len) {
                const item = array[i]!;
                if (f(item)) {
                    result.push(item);
                }
                i++;
            }
            return result;
        }
    }
    return array;
}
`

// overloadCertificateFilterAgrees lowers source in .ts and wants it to agree with Node on the JavaScript
// backend and natively, want being what Node prints. The native run is the one that matters, since only
// there is the argument handed back a reference the caller already holds.
func overloadCertificateFilterAgrees(t *testing.T, source, want string) {
	t.Helper()
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != want {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want %q, what Node prints", code, stdout, stderr, want)
	}
}

// Overload 2 with a predicate that keeps every element: filter hands back its argument, so the push into
// the result shows in the argument, as Node shows it. Before certificate D the call stopped at overload
// 2's census refusal, its result T[] over readonly T[] | undefined.
func TestOverloadCertificateFilterKeptArgumentComesBack(t *testing.T) {
	t.Parallel()
	source := overloadCertificateFilter + `const items: number[] = [3, 1, 2];
const kept = filter(items, x => x > 0);
kept.push(4);
console.log(String(items.length));
console.log(kept.join(","));
`
	overloadCertificateFilterAgrees(t, source, "4\n3,1,2,4\n")
}

// Overload 2 with a predicate that drops one element: the result is the fresh slice the body pushed into,
// so the push into it leaves the argument's length as it was.
func TestOverloadCertificateFilterDroppedResultIsFresh(t *testing.T) {
	t.Parallel()
	source := overloadCertificateFilter + `const items: number[] = [3, 1, 2];
const kept = filter(items, x => x !== 1);
kept.push(4);
console.log(String(items.length));
console.log(kept.join(","));
`
	overloadCertificateFilterAgrees(t, source, "3\n3,2,4\n")
}

// Overload 6, its argument typed number[] | undefined so the checker resolves the call past overload 2:
// undefined comes back as undefined, and an array as a fresh one the caller can push into.
func TestOverloadCertificateFilterMaybeArray(t *testing.T) {
	t.Parallel()
	source := overloadCertificateFilter + `function pick(present: boolean): number[] | undefined {
    return present ? [3, 1, 2] : undefined;
}
const none = filter(pick(false), x => x !== 1);
console.log(none === undefined ? "none" : none.join(","));
const some = filter(pick(true), x => x !== 1);
console.log(some === undefined ? "none" : some.join(","));
if (some !== undefined) {
    some.push(5);
    console.log(String(some.length));
}
`
	overloadCertificateFilterAgrees(t, source, "none\n3,2\n3\n")
}

// The census's own verdicts on the set: overloads 2 and 6 served, by certificate D; 3 and 7, each a
// predicate's readonly U[] the parameter's readonly T[] doesn't relate to, refused for their result.
func TestOverloadCertificateFilterCensus(t *testing.T) {
	t.Parallel()
	source := overloadCertificateFilter + "const items: number[] = [3, 1, 2];\nconsole.log(filter(items, x => x > 1).join(\",\"));\n"
	for _, ordinal := range []int{2, 6} {
		if verdict := overloadVerdictOf(t, source, ".ts", "filter", ordinal); verdict != nil {
			t.Errorf("want overload %d of filter served, got %v", ordinal, verdict)
		}
	}
	for _, ordinal := range []int{3, 7} {
		var refused *Refused
		verdict := overloadVerdictOf(t, source, ".ts", "filter", ordinal)
		if !errors.As(verdict, &refused) || !strings.Contains(refused.What, fmt.Sprintf("overload %d of filter result", ordinal)) || !strings.Contains(refused.What, "cannot be served by implementation result") {
			t.Errorf("want overload %d of filter refused for its result, got %v", ordinal, verdict)
		}
	}
}

// Overloads 3 and 7, each called directly with a type predicate over a readonly argument, stop at the
// call in their own words, as every call to an unserved overload does since #4czdn4q.
func TestOverloadCertificateFilterPredicateOverloadsStayRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, use string
		ordinal   int
	}{
		{"overload 3", "const frozen: readonly number[] = [1, 2, 1];\nconst ones = filter(frozen, (x): x is 1 => x === 1);\nconsole.log(String(ones.length));\n", 3},
		{"overload 7", "function pick(): readonly number[] | undefined {\n    return [1, 2, 1];\n}\nconst ones = filter(pick(), (x): x is 1 => x === 1);\nconsole.log(String(ones === undefined ? 0 : ones.length));\n", 7},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := overloadCertificateFilter + probe.use
			line := strings.Count(source[:strings.Index(source, "= filter(")], "\n") + 1
			_, err := lowerSourceAs(t, source, ".ts")
			var refused *Refused
			words := fmt.Sprintf("overload %d of filter result", probe.ordinal)
			if !errors.As(err, &refused) || !strings.Contains(refused.What, words) || !strings.Contains(refused.What, "cannot be served by implementation result") || !strings.Contains(refused.Where, fmt.Sprintf("main.ts:%d:", line)) {
				t.Fatalf("want %q refused at the call on line %d, got %v", words, line, err)
			}
		})
	}
}

// Bodies that hand back something other than their argument or a fresh array of its elements, each
// still refused for its result. m1 returns an alias it didn't build: a module's readonly array through a
// const, and the second parameter. m2 writes the parameter before returning it. Each hands the caller a
// readonly array as a mutable one. Mutants (overload_certificate_filter.mutants.md, J-5fnqd1q): m3, the
// freshness test dropped so every const that doesn't hold an element counts as fresh, reds "m1 outer alias"
// and, since filter's own len then counts as fresh, the pins; m3name, a const initialized to a name counted
// fresh, reds "m1 outer alias" alone; m4, overloadFilterParameterRead answering true, reds "m2 written
// parameter" alone.
func TestOverloadCertificateFilterNegativesStayRefused(t *testing.T) {
	t.Parallel()
	const observe = "const given: string[] = [\"given\"];\nbad(given, x => true).push(\"pushed\");\nconsole.log(String(frozen.length));\n"
	for _, probe := range []struct{ name, source string }{
		{"m1 outer alias", `const frozen: readonly string[] = ["frozen"];
function bad(array: string[], f: (x: string) => boolean): string[];
function bad(array: readonly string[] | undefined, f: (x: string) => boolean): readonly string[] | undefined {
    if (array !== undefined) {
        const result = frozen;
        return result;
    }
    return array;
}
` + observe},
		{"m1 second parameter", `const frozen: readonly string[] = ["frozen"];
function bad(array: string[], other: readonly string[]): string[];
function bad(array: readonly string[] | undefined, other: readonly string[]): readonly string[] | undefined {
    if (array !== undefined) {
        return other;
    }
    return array;
}
bad(["given"], frozen).push("pushed");
console.log(String(frozen.length));
`},
		{"m2 written parameter", `const frozen: readonly string[] = ["frozen"];
function bad(array: string[], f: (x: string) => boolean): string[];
function bad(array: readonly string[] | undefined, f: (x: string) => boolean): readonly string[] | undefined {
    array = frozen;
    return array;
}
` + observe},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			overloadVerdictRefused(t, probe.source, ".ts", "bad", 1, "overload 1 of bad result string[] ", "cannot be served by implementation result readonly string[] | undefined")
		})
	}
}
