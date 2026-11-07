package lower

import (
	"strings"
	"testing"
)

// tsc's sortAndDeduplicate falls back to a default comparer through `as unknown as Comparer<T>`
// (core.ts:810, the wall in front of 73 roots), a cast that's true only by the overloads: the one
// overload that lets a caller leave the comparers out takes strings. Per-overload proof erases the pair
// with no check (overload_default_cast.go; Adamic's ruling on #b5w3ycg, #25t5461). tsc's shape (core.ts:805-
// 810: an overload `(array: readonly string[])`, one taking a comparer, and the implementation's
// `equalityComparer ?? comparer ?? x as unknown as Comparer<T>`, with Comparer standing in for its
// EqualityComparer) lowers and agrees with Node, natively too; the string overload's T is inferred from
// array. The
// twins stay refused: a second overload that also lets the comparer be left out, under a generic T, and
// a default that's a number comparer under the string overload. In .a the pair keeps its refusal.
// The proof holds only while every call is resolved against the overloads (Compiler's road on 986ad82d).
// Assigning the implementation where its comparer is optional fails tsc's strictFunctionTypes today
// (Language's probes), but the proof shouldn't rest on that rule: any reference that isn't a callee keeps
// the pair refused, a plain alias and a .call among them, and so does any this parameter in the set,
// which shifts the positions the proof compares. A pair in a function with no overloads keeps its refusal (pick2; a
// lone hop to unknown, pick's, is an upcast main admits). Mutants: overloadAdmitsAbsence
// answering true for every overload, under which the tsc shape is refused (its comparer-taking overload
// is judged, at a generic T); castAssignable skipped in overloadDefaultPair, under which both twins
// lower; the overload count dropped, under which pick2's pair lowers; onlyCalled answering true, under
// which the function value and the .call lower; and the this-parameter check dropped, under which the
// this-parameter twin lowers.
func TestOverloadDefaultCastIsProvenPerOverload(test *testing.T) {
	test.Parallel()
	const want = "1,2,3\n"
	for _, source := range []struct{ name, text string }{{
		"tsc",
		"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n",
	}} {
		test.Run(source.name, func(test *testing.T) {
			test.Parallel()
			program := lowersAndAgreesWithNodeAs(test, source.text, ".ts")
			if code, stdout, stderr := brandedNative(test, program); code != 0 || stdout != "a,b,c\n"+want {
				test.Fatalf("native: exit %d, stdout %q, stderr %q", code, stdout, stderr)
			}
		})
	}
	for _, twin := range []struct{ name, text, where string }{
		{
			"generic",
			"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n",
			":18:50",
		},
		{
			"function_value",
			"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\nconst alias = sortAndDeduplicate;\nconsole.log(alias(['b', 'a']).join(','));\n",
			":18:50",
		},
		{
			"call",
			"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\nfunction compareAny(a: unknown, b: unknown): Comparison { return a === b ? 0 : 1; }\nconsole.log(sortAndDeduplicate.call(null, ['b', 'a'], compareAny).join(','));\n",
			":18:50",
		},
		{
			"this_parameter",
			"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(this: void, array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(this: void, array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(this: void, array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n",
			":18:50",
		},
		{
			"number",
			"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareNumbers as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareNumbers as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n",
			":18:50",
		},
	} {
		test.Run(twin.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceAsError := lowerSourceAs(test, twin.text, ".ts")
			if lowerSourceAsError == nil ||
				!strings.Contains(
					lowerSourceAsError.Error(),
					twin.where+": Adamic 0.1 refuses a cast the runtime can't check",
				) {
				test.Fatalf("want the cast refused at %s, got %v", twin.where, lowerSourceAsError)
			}
		})
	}
	test.Run("a", func(test *testing.T) {
		test.Parallel()
		_, lowerSourceAsError := lowerSourceAs(
			test,
			"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n",
			".a",
		)
		if lowerSourceAsError == nil {
			test.Fatal("want the pair refused in .a, got it lowered")
		}
		test.Logf("a: %v", lowerSourceAsError)
	})
	test.Run("no_overloads", func(test *testing.T) {
		test.Parallel()
		_, lowerSourceAsError := lowerSourceAs(
			test,
			"type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction pick<T>(comparer?: Comparer<T>): unknown {\n\treturn comparer ?? compareStringsCaseSensitive as unknown;\n}\nfunction pick2(comparer?: Comparer<string>): Comparer<string> {\n\treturn comparer ?? compareStringsCaseSensitive as unknown as Comparer<string>;\n}\nconsole.log(String(pick() === compareStringsCaseSensitive));\nconsole.log(String(pick2()('a', 'b')));\n",
			".ts",
		)
		if lowerSourceAsError == nil ||
			!strings.Contains(lowerSourceAsError.Error(), ":19:21: Adamic 0.1 refuses a cast the runtime can't check") {
			test.Fatalf("want pick2's pair refused at :19:21, got %v", lowerSourceAsError)
		}
	})
}
