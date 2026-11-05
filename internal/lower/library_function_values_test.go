package lower

import (
	"strings"
	"testing"
)

// Date.now read as a value into a slot of exactly `() => number` lowers (#4q1b2er head 2,
// library_function_values.go): a const's, a function's result through a ?: arm, a parameter's and a
// field's. The oracle holds what these run to (testdata/library_function_values in internal/oracle);
// this holds that they lower at all.
func TestLibraryFunctionValuesLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"a const of its type", "const clock: () => number = Date.now;\nconsole.log(String(clock() > 0));\n"},
		{"an arm of ?: in a return", "function pick(native: boolean): () => number {\n\treturn native ? () => 1 : Date.now;\n}\nconsole.log(String(pick(false)() > 0));\n"},
		{"an argument", "function run(clock: () => number): boolean {\n\treturn clock() > 0;\n}\nconsole.log(String(run(Date.now)));\n"},
		{"a field", "const holder: { readonly clock: () => number } = { clock: Date.now };\nconsole.log(String(holder.clock() > 0));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
		})
	}
}

// Every other read keeps today's stop, in its own words, at the read. A slot wider than `() =>
// number` reads the result boxed, where the function value returns an unboxed number; a read with
// no slot, or an === operand, has no type to see it as; another Date member isn't allowlisted, whose
// body would be another call; an Array method's value loses its object, and a Math method's isn't
// this rule's. A function value's name is an inherited library member, refused for every function
// value, and its length is not yet. Mutant (testdata/library_function_values/mutants.md in
// internal/oracle): the allowlist dropped to its global admits Date.parse, and its row goes red.
func TestLibraryFunctionValueReadsElsewhereKeepTheirStop(t *testing.T) {
	t.Parallel()
	const refused = "Adamic 0.1 refuses a method read as a value (now would lose its object, and this with it)"
	for _, probe := range []struct{ name, source, want string }{
		{"a slot returning unknown", "const clock: () => unknown = Date.now;\nconsole.log(typeof clock());\n", "main.a:1:30: " + refused},
		{"a slot returning number | string", "const clock: () => number | string = Date.now;\nconsole.log(typeof clock());\n", "main.a:1:38: " + refused},
		{"no slot", "const clock = Date.now;\nconsole.log(String(clock() > 0));\n", "main.a:1:15: " + refused},
		{"an === operand", "const clock: () => number = Date.now;\nconsole.log(String(clock === Date.now));\n", "main.a:2:30: " + refused},
		{"another Date member", "const parse: (text: string) => number = Date.parse;\nconsole.log(String(parse('2026-01-01')));\n", "main.a:1:41: Adamic 0.1 refuses a method read as a value (parse would lose its object, and this with it)"},
		{"an Array method", "const xs: number[] = [1];\nconst add: (...items: number[]) => number = xs.push;\nconsole.log(String(add(2)));\n", "main.a:2:45: Adamic 0.1 refuses a method read as a value (push would lose its object, and this with it)"},
		{"a Math method handed on", "function run(f: (x: number) => number): number {\n\treturn f(-2);\n}\nconsole.log(String(run(Math.abs)));\n", "main.a:4:24: Adamic 0.1 refuses a method read as a value (abs would lose its object, and this with it)"},
		{"its name", "const clock: () => number = Date.now;\nconsole.log(clock.name);\n", "main.a:2:13: Adamic 0.1 refuses inherited library member name read as an own field"},
		{"its length", "const clock: () => number = Date.now;\nconsole.log(String(clock.length));\n", "main.a:2:20: stage 0 can't lower .length on a value yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.want) {
				t.Errorf("got %v, want a stop containing %q", err, probe.want)
			}
		})
	}
}
