package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalWideningIsRefused(t *testing.T) {
	t.Parallel()
	const hidden = `const wide: { name: string; extra: number } = { name: 'a', extra: 1 };
const view: { name: string } = wide;
`
	for _, probe := range []struct{ name, source string }{
		{"assignment", hidden + `const narrow: { name: string; extra?: string } = view; console.log(narrow.extra ?? 'none');`},
		{"explicit type argument", hidden + `function read<T extends { readonly name: string; readonly extra?: string }>(value: T): string { return value.extra ?? 'none'; } console.log(read<{ readonly name: string; readonly extra?: string }>(view));`},
		{"argument", hidden + `function read(value: { name: string; extra?: string }): string { return value.extra ?? 'none'; } console.log(read(view));`},
		{"array element", hidden + `const values: { name: string }[] = [view]; const narrow: { name: string; extra?: string }[] = values; console.log(narrow.length.toFixed(0));`},
		{"source union member", `function read(value: { readonly name: string } | { readonly name: string; readonly extra: string }): string { const narrow: { readonly name: string; readonly extra?: string } = value; return narrow.extra ?? 'none'; } console.log(read({ name: 'a' }));`},
		{"target union member", hidden + `const narrow: { name: string; extra?: string } | { name: string; extra?: boolean } = view; console.log(narrow.name);`},
		{"conditional aliases", `function widen(flag: boolean): { name: string; extra?: string } { const first = { name: 'a' }; const second = { name: 'b' }; return flag ? first : second; } console.log(widen(true).name);`},
		{"return", `function widen(value: { name: string }): { name: string; extra?: string } { return value; } console.log(widen({ name: 'a' }).name);`},
		{"field", hidden + `class Box { readonly value: { name: string; extra?: string } = view; } console.log(new Box().value.name);`},
		{"readonly", hidden + `const narrow: { readonly name: string; readonly extra?: string } = view; console.log(narrow.name);`},
		{"nested alias", hidden + `const box = { value: view }; const narrow: { readonly value: { name: string; extra?: string } } = box; console.log(narrow.value.name);`},
		{"shorthand", hidden + `const narrow: { readonly view: { name: string; extra?: string } } = { view }; console.log(narrow.view.name);`},
		{"spread field", hidden + `const box = { view }; const narrow: { readonly view: { name: string; extra?: string } } = { ...box }; console.log(narrow.view.name);`},
		{"destructuring", hidden + `let narrow: { name: string; extra?: string } = { name: 'b' }; const pair: [{ name: string }] = [view]; [narrow] = pair; console.log(narrow.name);`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want optional widening refused, got %v", err)
			}
			for _, want := range []string{"adamic/no-optional-widening", "extra", "declare", "source type", "copy", "new object"} {
				if !strings.Contains(refused.Error(), want) {
					t.Errorf("refusal %q does not contain %q", refused.Error(), want)
				}
			}
		})
	}
}

func TestOptionalWideningSoundNeighbors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const source: { name: string; extra?: string } = { name: 'a' }; const view: { name: string; extra?: string } = source; console.log(view.extra ?? 'none');`,
		`const view: { name: string; extra?: string } = { name: 'a' }; console.log(view.extra ?? 'none');`,
		`const source = { name: 'a' }; const view: { readonly name: string; readonly extra?: string } = source; console.log(view.extra ?? 'none');`,
		`const source: { readonly name: string; readonly extra?: string } = { name: 'a', extra: 'yes' }; const view: { readonly name: string; readonly extra?: string } = source; console.log(view.extra ?? 'none');`,
		`function fresh(flag: boolean): { name: string; extra?: string } { return flag ? { name: 'a' } : { name: 'b' }; } console.log(fresh(true).name);`,
		`function keep(value: { name: string }): { name: string } | { name: string; extra?: string } { return value; } console.log(keep({ name: 'a' }).name);`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("sound neighbor: %v", err)
		}
	}
}
