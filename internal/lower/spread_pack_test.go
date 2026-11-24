package lower

import (
	"errors"
	"testing"
)

// A call with a spread packs each argument as it's held and its callee reads each slot as its
// parameter's type, so an argument held unlike its slot stops (#p9zm9q5, spreadPackStop). A
// reference read as a union is the same pointer and lowers.
func TestSpreadPackStops(t *testing.T) {
	t.Parallel()
	const union = "union of differently held members"
	header := "type A = number | string;\nconst nums: number[] = [1, 2];\nconst strs: string[] = [\"a\"];\n"
	for _, row := range []struct{ name, source, what string }{
		{"number spread into a union rest", "function show(...rest: A[]): number { return rest.length; }\nconsole.log(String(show(...nums)));\n", "a call with a spread of elements held as a number where the callee reads a " + union},
		{"number beside a spread into a union parameter", "function show(first: A, ...rest: number[]): number { return rest.length; }\nconsole.log(String(show(7, ...nums)));\n", "a call with a spread passing a number where the callee reads a " + union},
		{"number spread that can land in a union parameter", "function show(first: number, second?: A, ...rest: number[]): number { return rest.length; }\nconsole.log(String(show(1, ...nums)));\n", "a call with a spread of elements held as a number where the callee reads a " + union},
		{"number spread through a function value into a union rest", "const show = (...rest: A[]): number => rest.length;\nconsole.log(String(show(...nums)));\n", "a call with a spread of elements held as a number where the callee reads a " + union},
		{"string spread through a function value into a union rest", "const show = (...rest: A[]): number => rest.length;\nconsole.log(String(show(...strs)));\n", ""},
		{"string spread into a union rest", "function show(...rest: A[]): number { return rest.length; }\nconsole.log(String(show(...strs)));\n", ""},
		{"number spread into a number rest", "function show(first: A, ...rest: number[]): number { return rest.length; }\nconsole.log(String(show(\"x\", ...nums)));\n", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, header+row.source)
			if row.what == "" {
				if err != nil {
					t.Fatalf("got %v; want it lowered", err)
				}
				return
			}
			var gap *NotYet
			if !errors.As(err, &gap) || gap.What != row.what {
				t.Fatalf("got %v; want NotYet %q", err, row.what)
			}
		})
	}
}
