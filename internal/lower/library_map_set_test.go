package lower

import (
	"errors"
	"strings"
	"testing"
)

// These programs typecheck, but stage 0 must refuse the protocols or representations it cannot
// implement soundly. The positive cases are held to Node in the slice's oracle fixtures.
func TestLibraryMapSetGapsStayRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"set-like object", `const other: ReadonlySetLike<number> = {
	size: 0,
	has: (value: number): boolean => false,
	keys: () => new Set<number>().keys(),
};
const combined = new Set<number>().union(other);
`, "union with anything but a concrete library Set"},
		{"different representations", `const combined = new Set<number>([1]).union(new Set<string>(['a']));`, "union between Sets whose elements have different representations"},
		{"dynamic thisArg", `function visit(value: number): void {}
new Set<number>([1]).forEach(visit, { label: 'context' });`, "forEach thisArg with a callback other than an arrow"},
		{"groupBy optional widening", `const values: number[] = [1];
const groups = Map.groupBy(values, (value: number | undefined): number => 0);`, "Map.groupBy grouped elements held otherwise than the input's"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q refusal, got %v", probe.reason, err)
			}
		})
	}
}

func TestLibraryMapSetIteratorCopyTypesRefused(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function copy(value: MapIterator<number> | SetIterator<string>): void { const result = { ...value }; }`,
		`function copy(value: MapIterator<number> | undefined): void { const result = { ...value }; }`,
		`const result = { ...(new Set<string>(['a'.repeat(2)]).values()) };`,
		`function copy(value: MapIterator<number> | SetIterator<string>): void { const result = Object.assign({}, value); }`,
	} {
		_, err := lowerSource(t, source)
		var refusal *NotYet
		if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "next is inherited, not an own enumerable field") {
			t.Fatalf("want inherited-next refusal for %s, got %v", source, err)
		}
	}
}

// A generator gives [key, value] pairs, but generator functions are refused, so new Map from one names
// generators rather than calling its pairs something else (tsc's mapIterator and mapDefinedIterator
// sources, #assd1cf). Any other iterable keeps the old reason. Mutant: the generator branch dropped, and
// the first source says it isn't pairs.
func TestNewMapFromAGeneratorNamesGenerators(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"generator", `function collect(source: Generator<readonly [string, number], undefined, unknown>): number {
	return new Map(source).size;
}
console.log('built');
`, "new Map from a generator's pairs"},
		{"iterable", `function collect(source: Iterable<readonly [number, number]>): number {
	return new Map(source).size;
}
console.log(String(collect([[1, 2]])));
`, "new Map from something that isn't [key, value] pairs"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q, got %v", probe.reason, err)
			}
		})
	}
}
