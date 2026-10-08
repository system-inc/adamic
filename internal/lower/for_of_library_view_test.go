package lower

import (
	"errors"
	"testing"
)

func TestForOfLibraryViewOriginChecks(t *testing.T) {
	t.Parallel()
	const set = `const values = new Set<number>(); values.add(1); `
	const consumer = `function consume(source: Iterable<number>): void { for (const value of source) { } } `
	const custom = `const custom = { [Symbol.iterator]() { return { next() { return { done: true, value: 1 }; }, return() { return { done: true, value: 0 }; }, throw() { return { done: true, value: 0 }; } }; } }; `
	for _, probe := range []struct{ name, source string }{
		{"no callers", set + consumer},
		{"array caller", set + consumer + `consume([1]);`},
		{"mixed callers", set + consumer + `consume(values.values()); consume([1]);`},
		{"stored iterator", set + consumer + `const iterator = values.values(); consume(iterator);`},
		{"factory wrapper", set + consumer + `function make() { return values.values(); } consume(make());`},
		{"escaped function", set + consumer + `const alias = consume; consume(values.values());`},
		{"exported function", set + `export ` + consumer + `consume(values.values());`},
		{"exported alias", set + consumer + `export { consume as exported }; consume(values.values());`},
		{"parameter reassigned", set + `function consume(source: Iterable<number>): void { source = values.values(); for (const value of source) { } } consume(values.values());`},
		{"parameter aliased", set + `function consume(source: Iterable<number>): void { const alias = source; for (const value of source) { } } consume(values.values());`},
		{"default parameter", set + `function consume(source: Iterable<number> = values.values()): void { for (const value of source) { } } consume(values.values());`},
		{"generic consumer", set + `function consume<T>(source: Iterable<T>): void { for (const value of source) { } } consume(values.values());`},
		{"generic unused type", set + `function consume<T>(source: Iterable<number>): void { for (const value of source) { } } consume<string>(values.values());`},
		{"maybe element", `const values = new Set<number | undefined>(); function consume(source: Iterable<number | undefined>): void { for (const value of source) { } } consume(values.values());`},
		{"object binding", `const values = new Set<{ readonly value: number }>(); function consume(source: Iterable<{ readonly value: number }>): void { for (const { value } of source) { } } consume(values.values());`},
		{"custom iterator", consumer + custom + `consume(custom);`},
		{"custom reassignment", set + custom + `function consume(source: Iterable<number>): void { source = custom; for (const value of source) { } } consume(values.values());`},
		{"spread shifts parameter", set + custom + `function consume(prefix: number, source: Iterable<number>, ignored?: Iterable<number>): void { for (const value of source) { } } consume(...[1, custom] as const, values.values());`},
		{"weak storage", `import type { Weak } from 'adamic'; interface Row { readonly name: string; } const values = new Set<Weak<Row>>(); function consume(source: Iterable<Weak<Row>>): void { for (const value of source) { } } consume(values.values());`},
		{"present weak storage", `import type { Weak } from 'adamic'; interface Row { readonly name: string; } type Present = Exclude<Weak<Row>, undefined>; const values = new Set<Present>(); function consume(source: Iterable<Present>): void { for (const value of source) { } } consume(values.values());`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != "for...of over an object" {
				t.Fatalf("want the unproved iterator-view stop, got %v", err)
			}
		})
	}
}
