package lower

import (
	"strings"
	"testing"
)

// tsc's sortAndDeduplicate falls back to a default comparer through `as unknown as Comparer<T>`
// (core.ts:810, the wall in front of 73 roots), a cast that's true only by the overloads: the one
// overload that lets a caller leave the comparers out takes strings. Per-overload proof erases the pair
// with no check (overload_default_cast.go; Adamic's ruling on #b5w3ycg, #25t5461). tsc's shape, with its
// overload written `<T extends string>(array: readonly string[])`, lowers and agrees with Node, natively
// too, and so does the same shape with that overload's binder left off, which infers T from array. The
// twins stay refused: a second overload that also lets the comparer be left out, under a generic T, and
// a default that's a number comparer under the string overload. In .a the pair keeps its refusal.
// Mutants: overloadAdmitsAbsence answering true for every overload, under which the tsc shape is refused
// (its comparer-taking overload is judged, at a generic T); and castAssignable skipped in
// overloadDefaultPair, under which both twins lower.
func TestOverloadDefaultCastIsProvenPerOverload(t *testing.T) {
	t.Parallel()
	const want = "1,2,3\n"
	for _, source := range []struct{ name, text string }{{"tsc", "type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate<T extends string>(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n"}, {"plain", "type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n"}} {
		t.Run(source.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, source.text, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "a,b,c\n"+want {
				t.Fatalf("native: exit %d, stdout %q, stderr %q", code, stdout, stderr)
			}
		})
	}
	for _, twin := range []struct{ name, text, where string }{{"generic", "type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n", ":18:50"}, {"number", "type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareNumbers as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareNumbers as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n", ":18:50"}} {
		t.Run(twin.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, twin.text, ".ts")
			if err == nil || !strings.Contains(err.Error(), twin.where+": Adamic 0.1 refuses a cast the runtime can't check") {
				t.Fatalf("want the cast refused at %s, got %v", twin.where, err)
			}
		})
	}
	t.Run("a", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSourceAs(t, "type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction sortAndDeduplicate(array: readonly string[]): string[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: Comparer<T>): T[];\nfunction sortAndDeduplicate<T>(array: readonly T[], comparer?: Comparer<T>, equalityComparer?: Comparer<T>): T[] {\n\treturn dedupeSorted(sortWith(array, comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>), equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>);\n}\nconsole.log(sortAndDeduplicate(['b', 'a', 'c', 'a']).join(','));\nconsole.log(sortAndDeduplicate([3, 1, 2, 1], compareNumbers).join(','));\n", ".a")
		if err == nil {
			t.Fatal("want the pair refused in .a, got it lowered")
		}
		t.Logf("a: %v", err)
	})
	t.Run("lone", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSourceAs(t, "type Comparison = -1 | 0 | 1;\ntype Comparer<T> = (a: T, b: T) => Comparison;\nfunction compareStringsCaseSensitive(a: string, b: string): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction compareNumbers(a: number, b: number): Comparison { return a < b ? -1 : a > b ? 1 : 0; }\nfunction sortWith<T>(array: readonly T[], comparer: Comparer<T>): T[] { return [...array].sort(comparer); }\nfunction dedupeSorted<T>(sorted: readonly T[], comparer: Comparer<T>): T[] {\n\tconst out: T[] = [];\n\tfor (const item of sorted) {\n\t\tif (!out.some(seen => comparer(seen, item) === 0)) {\n\t\t\tout.push(item);\n\t\t}\n\t}\n\treturn out;\n}\nfunction pick<T>(comparer?: Comparer<T>): unknown {\n\treturn comparer ?? compareStringsCaseSensitive as unknown;\n}\nfunction pick2(comparer?: Comparer<string>): Comparer<string> {\n\treturn comparer ?? compareStringsCaseSensitive as unknown as Comparer<string>;\n}\nconsole.log(String(pick() === compareStringsCaseSensitive));\nconsole.log(String(pick2()('a', 'b')));\n", ".ts")
		t.Logf("lone: %v", err)
	})
}
