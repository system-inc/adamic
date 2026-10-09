package lower

import "testing"

func TestClassArgumentsUseCheckerIdentity(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, types string
		classes     int
	}{
		{"separate literals", `type A = { readonly n: number }; type B = { readonly n: number };`, 1},
		{"property order", `type A = { readonly n: number; readonly s: string }; type B = { readonly s: string; readonly n: number };`, 1},
		{"nested arrays", `type A = readonly { readonly n: number }[]; type B = readonly { readonly n: number }[];`, 1},
		{"tuples", `type A = readonly [number, string]; type B = readonly [number, string];`, 1},
		{"unions", `type A = { readonly n: number } | string; type B = string | { readonly n: number };`, 1},
		{"functions", `type A = (n: number) => string; type B = (value: number) => string;`, 1},
		{"recursive interfaces", `interface A { readonly n: number; readonly next?: A; } interface B { readonly n: number; readonly next?: B; }`, 1},
		{"different fields", `type A = { readonly n: number }; type B = { readonly n: string };`, 2},
		{"optional field", `type A = { readonly n: number }; type B = { readonly n?: number };`, 2},
		{"same representation", `type A = number; type B = 1;`, 2},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNode(t, probe.types+`
class Box<T> {}
const first = new Box<A>();
const second = new Box<B>();
console.log(first === second ? 'same' : 'distinct'); console.log(first instanceof Box ? 'first Box' : 'missing first'); console.log(second instanceof Box ? 'second Box' : 'missing second');
`)
			// Node behavior cannot observe how many native layouts share a checker identity.
			if len(program.Classes) != probe.classes {
				t.Fatalf("got %d class instantiations, want %d", len(program.Classes), probe.classes)
			}
		})
	}
}
