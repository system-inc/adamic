package lower

import (
	"errors"
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
			t.Parallel()
			lowersAndAgreesWithNode(t, source)
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
			t.Parallel()
			_, err := lowerSource(t, test.source)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want %s, got %v", test.reason, err)
			}
		})
	}
}

// Pin the intended refusal as well as admission: an unrelated failure is not evidence
// that the method value kept its receiver and callable ABI.
func TestLibraryMethodValueSafety(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		source, reason string
		refused        bool
	}{
		"mutable alias":         {"let f = Number.parseInt; f = (text: string): number => 99; console.log(`${f('12', 10)}`);", "a method read as a value (parseInt", true},
		"opaque field":          {"const f = Number.parseInt; const object = {f: f}; console.log(`${object.f('12')}`);", "a library method value outside a const alias, typed call/apply", false},
		"opaque shorthand":      {"const f = Number.parseInt; const g = f; const object = {g}; console.log(`${object.g('12')}`);", "an object field erases its receiver and callable ABI", false},
		"opaque return":         {"function get() { const f = Number.parseInt; return f; } console.log(`${get()('12')}`);", "a library method value outside a const alias, typed call/apply", false},
		"absent array":          {"function get(): number[] | undefined { return undefined; } const f = Array.prototype.indexOf; f.call(get(), 1);", "an Array method receiver that may be undefined", false},
		"array element erasure": {"const f = Array.prototype.slice; const wrong: string[] = f.call([1, 2]); console.log(wrong.join());", "erased or different element type", false},
		"apply spread":          {"const f = String.prototype.substring; console.log(f.apply('abcd', [1, ...([2] as const)]));", "apply with holes or spread", false},
		"call spread":           {"const f = Number.parseInt; console.log(`${f.call(undefined, '10', ...([16] as const))}`);", "spread into a delayed library call", false},
		"string receiver":       {"const f = String.prototype.toString; f.call(1);", "primitive string representation is not proven", false},
		"number receiver":       {"const f = Number.prototype.toString; f.call('10');", "a Number method receiver without a proven number internal slot", false},
		"bind arity":            {"const f = String.prototype.substring; f.bind('abcd');", "bind of a method with optional, required or partial arguments", false},
		"bind absent receiver":  {"function get(): string | undefined { return undefined; } String.prototype.trim.bind(get());", "bind without a proven present primitive string receiver", false},
		"opaque identity":       {"const f = Number.parseInt; console.log(`${f === f}`);", "a library method value outside a const alias, typed call/apply", false},
		"unsupported callback":  {"const f = String.prototype.trim; console.log([' x '].map(f).join());", "a library map callback without a proven primitive adapter", false},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, test.source)
			var missing *NotYet
			var refused *Refused
			correctKind := errors.As(err, &missing)
			if test.refused {
				correctKind = errors.As(err, &refused)
			}
			if !correctKind || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want method-value refusal %q (Refused=%t), got %v", test.reason, test.refused, err)
			}
		})
	}
}
