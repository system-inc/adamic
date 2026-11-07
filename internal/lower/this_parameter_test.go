package lower

import (
	"errors"
	"strings"
	"testing"
)

// A function declaration or method with a this parameter followed by another parameter is not yet (#j8mt4pr): called
// directly, its arguments would shift by one, and the JavaScript backend printed NaN where Node prints 2. Each shape
// stops at its this parameter, in .a and .ts: Optimizer's p2, a predicate shaped the same way (which
// panicked in R-T80's check), class instance and static methods and an object literal's method (CLI's
// read of signature()), and an overloaded declaration. Mutant: signature()'s this guard (functions.go) dropped lowers
// p2 again, and the JavaScript backend prints NaN.
func TestThisParameterIsNotYet(test *testing.T) {
	test.Parallel()
	const reason = "a function or method with a this parameter followed by another parameter"
	for _, probe := range []struct{ name, source string }{
		{"p2", "function inc(this: void, n: number): number { return n + 1; }\nconsole.log(String(inc(1)));\n"},
		{
			"predicate",
			"interface N { readonly kind: number; } interface A extends N { readonly kind: 1; }\nfunction isA(this: void, n: N): n is A { return n.kind === 1; }\nconst n: N = { kind: 1 };\nconsole.log(String(isA(n)));\n",
		},
		{
			"class method",
			"class Counter { base = 1; add(this: Counter, n: number): number { return this.base + n; } }\nconsole.log(String(new Counter().add(1)));\n",
		},
		{
			"static method",
			"class Counter { static add(this: void, n: number): number { return n + 1; } }\nconsole.log(String(Counter.add(1)));\n",
		},
		{
			"object literal method",
			"interface Box { base: number; add(this: Box, n: number): number; }\nconst box: Box = { base: 1, add(this: Box, n: number): number { return this.base + n; } };\nconsole.log(String(box.add(1)));\n",
		},
		{
			"overloaded",
			"function pick(this: void, n: number): number;\nfunction pick(this: void, n: number): number { return n; }\nconsole.log(String(pick(3)));\n",
		},
	} {
		for _, extension := range []string{".a", ".ts"} {
			test.Run(probe.name+extension, func(test *testing.T) {
				test.Parallel()
				_, lowerSourceAsError := lowerSourceAs(test, probe.source, extension)
				var notYet *NotYet
				if !errors.As(lowerSourceAsError, &notYet) || !strings.Contains(lowerSourceAsError.Error(), reason) {
					test.Fatalf("got %v, want NotYet %q", lowerSourceAsError, reason)
				}
			})
		}
	}
	// With nothing after this, nothing shifts: these lower, and namespace_method_receiver.a holds the
	// object literal's to Node.
	for _, control := range []struct{ name, source string }{
		{
			"object literal method alone",
			"const box = { flags: 7, value(this: { flags: number }): number { return this.flags; } };\nconsole.log(String(box.value()));\n",
		},
		{
			"class method alone",
			"class Counter { base = 1; get(this: Counter): number { return this.base; } }\nconsole.log(String(new Counter().get()));\n",
		},
	} {
		for _, extension := range []string{".a", ".ts"} {
			test.Run(control.name+extension, func(test *testing.T) {
				test.Parallel()
				if _, lowerSourceAsError := lowerSourceAs(test, control.source, extension); lowerSourceAsError != nil {
					test.Fatalf("got %v, want it to lower", lowerSourceAsError)
				}
			})
		}
	}
}
