package lower

import (
	"strings"
	"testing"
)

func TestLibrarySequenceIteratorRefusals(t *testing.T) {
	for _, source := range []string{
		`console.log(Object.keys([1].values()).join(','));`,
		"console.log(`${Object.hasOwn('abc'[Symbol.iterator](), 'next')}`);",
		`const copy = { ...[1].entries() };`,
		`const copy = Object.assign({}, 'abc'[Symbol.iterator]());`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			if err == nil || !strings.Contains(err.Error(), "next is inherited") {
				t.Fatalf("want private iterator slot refusal, got %v", err)
			}
		})
	}
}

func TestLibrarySequenceIteratorCycles(t *testing.T) {
	for _, factory := range []string{"items.values()", "items[Symbol.iterator]()"} {
		t.Run(factory, func(t *testing.T) {
			_, err := lowerSource(t, `interface Item { iterator: ArrayIterator<Item> | undefined; }
const items: Item[] = [];
const item: Item = { iterator: undefined };
items.push(item);
item.iterator = `+factory+`;`)
			if err == nil || !strings.Contains(err.Error(), "cycle") {
				t.Fatalf("want hidden iterator source cycle refused, got %v", err)
			}
		})
	}
}

func TestLibraryIteratorCustomProtocolDispatchRefused(t *testing.T) {
	_, err := lowerSource(t, `const other = ['other'].values();
const custom: ArrayIterator<string> = {
 next: (): IteratorResult<string, undefined> => ({ done: false, value: 'custom' }),
 [Symbol.iterator]: (): ArrayIterator<string> => other,
};
for (const value of custom) { console.log(value); break; }`)
	if err == nil || !strings.Contains(err.Error(), "custom iterator protocol dispatch") {
		t.Fatalf("want compiler protocol dispatch boundary, got %v", err)
	}
}

func TestLibraryIteratorCompilerSlotsRefused(t *testing.T) {
	for _, probe := range []struct{ source, reason string }{
		{`const iterator = [true].values();`, "compiler slot representation support"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %s refusal, got %v", probe.reason, err)
			}
		})
	}
}

func TestLibraryIteratorBuiltinOverridesRefused(t *testing.T) {
	for _, source := range []string{
		`const values = [1]; values[Symbol.iterator] = () => [9].values();`,
		`const values = new Map<number, number>(); values[Symbol.iterator] = () => new Map<number, number>().entries();`,
		`const values = new Set<number>(); values[Symbol.iterator] = () => [9].values();`,
		`Array.prototype[Symbol.iterator] = () => [9].values();`,
		`String.prototype[Symbol.iterator] = () => 'other'[Symbol.iterator]();`,
		`Map.prototype[Symbol.iterator] = () => new Map<number, number>().entries();`,
		`Set.prototype[Symbol.iterator] = () => [9].values();`,
		`const values = new Uint8Array(1); values[Symbol.iterator] = () => [9].values();`,
		`Uint8Array.prototype[Symbol.iterator] = () => [9].values();`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			if err == nil || !(strings.Contains(err.Error(), "inherited library member") || strings.Contains(err.Error(), "assigning an element of a value") || strings.Contains(err.Error(), "new an Identifier")) {
				t.Fatalf("want compiler built-in write refusal, got %v", err)
			}
		})
	}
}

func TestLibraryIteratorCompletedUndefinedValue(t *testing.T) {
	if _, err := lowerSource(t, "const step = [1].values().next(); if (step.done) { console.log(`${step.value === undefined}`); }"); err != nil {
		t.Fatal(err)
	}
}
