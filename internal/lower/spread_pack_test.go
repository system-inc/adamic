package lower

import (
	"errors"
	"strings"
	"testing"
)

// A call with a spread packs each argument as it's held and its callee reads each slot as its
// parameter's type, so an argument held unlike its slot stops (#p9zm9q5, spreadPackStop). A
// reference read as a union is the same pointer and lowers.
func TestSpreadPackStops(t *testing.T) {
	t.Parallel()
	const union = "union of differently held members"
	header := "type A = number | string;\nconst nums: number[] = [1, 2];\nconst strs: string[] = [\"a\"];\n"
	// The closure road judges a spread against the value's signature, which stands in for the target's own
	// slots: sound only because a function value whose parameters are held otherwise than its type is
	// refused where it's made (#2t3h708), so this row has to stop before it runs, never at run time.
	_, err := lowerSource(t, header+"const f: (...rest: number[]) => number = (...rest: A[]): number => rest.length;\nconsole.log(String(f(...nums)));\n")
	if err == nil {
		t.Fatal("a rest held as a union behind a number rest type lowered; the spread would be judged against the wrong slot")
	}
	var wide *NotYet
	if errors.As(err, &wide) && strings.HasPrefix(wide.What, "a call with a spread") {
		t.Fatalf("got %v; want the function-value relation's stop, not the spread's", err)
	}
	for _, row := range []struct{ name, source, what string }{
		{"number spread into a union rest", "function show(...rest: A[]): number { return rest.length; }\nconsole.log(String(show(...nums)));\n", "a call with a spread of elements held as a number where the callee reads a " + union},
		{"number beside a spread into a union parameter", "function show(first: A, ...rest: number[]): number { return rest.length; }\nconsole.log(String(show(7, ...nums)));\n", "a call with a spread passing a number where the callee reads a " + union},
		{"number spread that can land in a union parameter", "function show(first: number, second?: A, ...rest: number[]): number { return rest.length; }\nconsole.log(String(show(1, ...nums)));\n", "a call with a spread of elements held as a number where the callee reads a " + union},
		{"number spread through a function value into a union rest", "const show = (...rest: A[]): number => rest.length;\nconsole.log(String(show(...nums)));\n", "a call with a spread of elements held as a number where the callee reads a " + union},
		// lead's type admits show, so its layout collects show's union rest from the pack its number spread is in
		// (#wzr4atz, checkClosureSpreadLayouts), though lead's own signature reads the spread alike.
		{"number spread through a function value whose layout holds another target's union rest", "const show = (...rest: A[]): void => { for (const value of rest) console.log(typeof value + \":\" + String(value)); };\nshow(\"a\", \"b\");\nconst lead = (first: A, ...rest: number[]): void => { console.log(typeof first + \":\" + String(first) + \":\" + String(rest.length)); };\nlead(7, ...nums);\n", "a call with a spread through a function value whose layout reads a " + union + " where the call packs a number"},
		{"string spread through a function value into a union rest", "const show = (...rest: A[]): number => rest.length;\nconsole.log(String(show(...strs)));\n", ""},
		{"number beside a spread where another target reads number | undefined", "const loose = (first?: number, ...rest: number[]): number => rest.length;\nconst strict = (first: number, ...rest: number[]): number => rest.length + first;\nconsole.log(String(loose(1, ...nums)) + String(strict(2, ...nums)));\n", ""},
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
