package lower

import (
	"errors"
	"strings"
	"testing"
)

// A function declaration or method with a this parameter is not yet (#j8mt4pr): called directly, its
// arguments would shift by one, and the JavaScript backend printed NaN where Node prints 2. Each shape
// stops at its this parameter, in .a and .ts: Optimizer's p2, a predicate shaped the same way (which
// panicked in R-T80's check), class instance and static methods and an object literal's method (CLI's
// read of signature()), and an overloaded declaration. Mutant: signature()'s this guard (functions.go) dropped lowers p2 again, and the JavaScript backend prints NaN.
func TestThisParameterIsNotYet(t *testing.T) {
	t.Parallel()
	const reason = "a function or method with a this parameter"
	for _, probe := range []struct{ name, source string }{
		{"p2", "function inc(this: void, n: number): number { return n + 1; }\nconsole.log(String(inc(1)));\n"},
		{"predicate", "interface N { readonly kind: number; } interface A extends N { readonly kind: 1; }\nfunction isA(this: void, n: N): n is A { return n.kind === 1; }\nconst n: N = { kind: 1 };\nconsole.log(String(isA(n)));\n"},
		{"class method", "class Counter { base = 1; add(this: Counter, n: number): number { return this.base + n; } }\nconsole.log(String(new Counter().add(1)));\n"},
		{"static method", "class Counter { static add(this: void, n: number): number { return n + 1; } }\nconsole.log(String(Counter.add(1)));\n"},
		{"object literal method", "interface Box { base: number; add(this: Box, n: number): number; }\nconst box: Box = { base: 1, add(this: Box, n: number): number { return this.base + n; } };\nconsole.log(String(box.add(1)));\n"},
		{"overloaded", "function pick(this: void, n: number): number;\nfunction pick(this: void, n: number): number { return n; }\nconsole.log(String(pick(3)));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), reason) {
					t.Fatalf("got %v, want NotYet %q", err, reason)
				}
			})
		}
	}
}
