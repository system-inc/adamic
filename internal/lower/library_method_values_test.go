package lower

import (
	"strings"
	"testing"
)

func TestLibraryMethodValues(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const f = Array.prototype.indexOf; console.log(`${f.call([4, 5], 5)}`);",
		"const f = Array.prototype.indexOf; const g = f; console.log(`${g.apply([4, 5], [5])}`);",
		"const f = String.prototype.trim; console.log(f.call(' x '));",
		"const f = String.prototype.substring; console.log(f.apply('abcdef', [4, 1]));",
		"const f = Number.parseInt; console.log(`${f('12', 10)}`);",
		"const f = Object.prototype.toString; console.log(f.call([4]));",
		"const f = Object.prototype.hasOwnProperty; console.log(`${f.call({x: 1}, 'x')}`);",
		"console.log([1, 2].map(String).join(','));",
		"function show(value: string | number): void { console.log(`${value}`); } const f = Object.prototype.toString; show(f.call([1]));",
		"console.log([].map(String).join(','));",
		"const f = String.prototype.trim.bind(' x '); console.log(f());",
		"console.log(['10', '10', '10'].map(Number.parseInt).join(','));",
		"const f = Number.parseInt; console.log(['10', '10'].map(f).join(','));",
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLibraryMethodValueBoundaries(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ source, reason string }{
		{"const f = Array.prototype.indexOf; f(1);", "unbound Array"},
		{"const f = Array.prototype.slice; const wrong: string[] = f.call([1, 2]);", "different element type"},
		{"const f = String.prototype.trim; f.call(1);", "primitive string"},
		{"const f = Array.prototype.indexOf; f.call('x', 1);", "dense array"},
		{"const f = Array.prototype.indexOf; const args: [number] = [1]; f.apply([1], args);", "dense argument literal"},
		{"const f = Array.prototype.indexOf; f.apply([1], [...([1] as const)]);", "holes or spread"},
		{"const f = Array.prototype.indexOf; console.log(`${f === f}`);", "outside a const alias"},
		{"const f = Array.prototype.indexOf; const o = {f};", "outside a const alias"},
		{"const f = Array.prototype.indexOf; f.bind([1]);", "captured receiver"},
		{"console.log([{x: 1}].map(String).join(','));", "primitive adapter"},
		{"const f = String.prototype.trim; ['x'].map(f);", "primitive adapter"},
	} {
		t.Run(test.reason, func(t *testing.T) {
			_, err := lowerSource(t, test.source)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want %s, got %v", test.reason, err)
			}
		})
	}
}

// These are admissibility checks, not diagnostic wording checks. A source mutant must actually
// admit the unsafe program to fail one of them.
func TestLibraryMethodValueSafety(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"mutable alias":         "let f = Number.parseInt; f = (text: string): number => 99; console.log(`${f('12', 10)}`);",
		"opaque field":          "const f = Number.parseInt; const object = {f: f}; console.log(`${object.f('12')}`);",
		"opaque shorthand":      "const f = Number.parseInt; const g = f; const object = {g}; console.log(`${object.g('12')}`);",
		"opaque return":         "function get() { const f = Number.parseInt; return f; } console.log(`${get()('12')}`);",
		"absent array":          "function get(): number[] | undefined { return undefined; } const f = Array.prototype.indexOf; f.call(get(), 1);",
		"array element erasure": "const f = Array.prototype.slice; const wrong: string[] = f.call([1, 2]); console.log(wrong.join());",
		"apply spread":          "const f = String.prototype.substring; console.log(f.apply('abcd', [1, ...([2] as const)]));",
		"call spread":           "const f = Number.parseInt; console.log(`${f.call(undefined, '10', ...([16] as const))}`);",
		"string receiver":       "const f = String.prototype.toString; f.call(1);",
		"number receiver":       "const f = Number.prototype.toString; f.call('10');",
		"bind arity":            "const f = String.prototype.substring; f.bind('abcd');",
		"bind absent receiver":  "function get(): string | undefined { return undefined; } String.prototype.trim.bind(get());",
		"opaque identity":       "const f = Number.parseInt; console.log(`${f === f}`);",
		"unsupported callback":  "const f = String.prototype.trim; console.log([' x '].map(f).join());",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := lowerSource(t, source); err == nil {
				t.Fatal("unsafe method value was admitted")
			}
		})
	}
}

func TestLibraryMethodConsoleScalarBoundary(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const f = String.prototype.trim; console.log(f.call(' x '));",
		"const f = String.prototype.normalize; console.log((f.call('x')));",
		"const f = String.prototype.padEnd; console.log(f.call('x', 3));",
		"const f = Object.prototype.toString; console.log(f.call([1]));",
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLibraryMethodConsoleScalarBoundaryRefusals(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const f = String.prototype.at; console.log(f.call('x', 9));",
		"const f = String.prototype.at; console.log(f.call('x', 9)!);",
		"const f = String.prototype.trim; const narrowed = f as (this: string) => 'trusted'; console.log(narrowed.call(' x '));",
		"function show(value: string | null | undefined): void {} const f = String.prototype.trim; show(f.call(' x '));",
		"function show(value: string | null | undefined): void { console.log(value); } show(undefined);",
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err == nil {
				t.Fatal("unproven alias result or general nullable view was admitted")
			}
		})
	}
}
