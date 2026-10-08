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
			_, err := lowerSource(t, `interface Item { iterator?: ArrayIterator<Item>; }
const items: Item[] = [];
const item: Item = {};
items.push(item);
item.iterator = `+factory+`;`)
			if err == nil || !strings.Contains(err.Error(), "cycle") {
				t.Fatalf("want hidden iterator source cycle refused, got %v", err)
			}
		})
	}
}

func TestLibraryIteratorCustomOriginRefused(t *testing.T) {
	_, err := lowerSource(t, `const other = ['other'].values();
const custom: ArrayIterator<string> = {
 next: (): IteratorResult<string, undefined> => ({ done: false, value: 'custom' }),
 [Symbol.iterator]: (): ArrayIterator<string> => other,
};
for (const value of custom) { console.log(value); break; }`)
	if err == nil || !strings.Contains(err.Error(), "compiler protocol origin is not proved") {
		t.Fatalf("want custom iterator origin refusal, got %v", err)
	}
}

func TestLibraryIteratorCompilerSlotsRefused(t *testing.T) {
	for _, probe := range []struct{ source, reason string }{
		{`const iterator = [true].values();`, "compiler slot representation support"},
		{"const step = [1].values().next(); if (step.done) { console.log(`${step.value === undefined}`); }", "a value of type undefined"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %s refusal, got %v", probe.reason, err)
			}
		})
	}
}
