package lower

import (
	"errors"
	"strings"
	"testing"
)

// A function declaration with a this parameter is not yet (#j8mt4pr): called directly, its arguments
// would shift by one, and the JavaScript backend printed NaN where Node prints 2. Each shape stops at
// the declaration's this parameter, in .a and .ts, and so does a predicate shaped the same way, which
// panicked in R-T80's check. Mutant: thisParameterDeclarationGap returning nil lowers p2 again.
func TestThisParameterDeclarationIsNotYet(t *testing.T) {
	t.Parallel()
	const reason = "a function declaration with a this parameter"
	for _, probe := range []struct{ name, source string }{
		{"p2", "function inc(this: void, n: number): number { return n + 1; }\nconsole.log(String(inc(1)));\n"},
		{"predicate", "interface N { readonly kind: number; } interface A extends N { readonly kind: 1; }\nfunction isA(this: void, n: N): n is A { return n.kind === 1; }\nconst n: N = { kind: 1 };\nconsole.log(String(isA(n)));\n"},
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
