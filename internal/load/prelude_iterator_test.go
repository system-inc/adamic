package load

import "testing"

// The prelude's Map and Set iterators return a done-required result, and its done arm is
// IteratorReturnResult<undefined>, so inferring T from Iterable<T> takes the element alone.
// tsc's builder.ts:1246 is the first program it reached: arrayFrom over a Set<Path> (#onediag).
// Mutant: the done arm written as { readonly done: true; readonly value: undefined }, which
// gives every element type below a | undefined, and r 'Path | undefined' with it.
func TestPreludeIteratorsInferTheirElement(test *testing.T) {
	test.Parallel()
	paths := writeProgram(test, [2]string{"main.ts", `type Path = string & { __pathBrand: any };
function arrayFrom<T, U>(iterator: Iterable<T>, map: (t: T) => U): U[] {
	const result: U[] = [];
	for (const value of iterator) {
		result.push(map(value));
	}
	return result;
}
const rootFileNames = new Set<Path>();
const root = arrayFrom(rootFileNames, r => r.length);
const sizes = new Map<string, number>();
const sizeKeys = Array.from(sizes.keys());
const keyCopy = new Set(sizes.keys());
const valueCopy = new Set(sizes.values());
const step = rootFileNames.values().next();
if (step.done) {
	const finished = step.value;
} else {
	const element = step.value;
}
`})
	byName := declarationsByName(test, paths)
	for name, want := range map[string]string{
		"parameter r":        "Path",
		"variable root":      "number[]",
		"variable sizeKeys":  "string[]",
		"variable keyCopy":   "Set<string>",
		"variable valueCopy": "Set<number>",
		"variable finished":  "undefined",
		"variable element":   "Path",
	} {
		if got := byName[name].Type; got != want {
			test.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}
