package lower

import (
	"errors"
	"strings"
	"testing"
)

// A rest parameter whose type isn't an array reference is NotYet at the parameter. Asking the checker
// for its type arguments first, as the order since 7e635882 did, dereferenced nil in tsc's
// getTypeArguments, and tsc's createNodeFactory panicked the stage 3 meter's latent census once slice 5
// let its nested declarations reach createConstructorTypeNode's Parameters<typeof a | typeof b> (#egxccg8).
func TestRestParametersThatArentArraysAreNotYet(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, where, what string }{
		{
			"union of arrays",
			"function either(...args: string[] | number[]): number {\n    return 0;\n}\nconsole.log(\"one\");\n",
			":1:17",
			"a rest parameter that isn't an array",
		},
		{
			"union of tuples",
			"function f(...args: [string] | [string, number]): void {}\n",
			":1:12",
			"a rest parameter that isn't an array",
		},
		{
			"nested union of tuples",
			"function outer(): void {\n    function f(...args: [string] | [string, number]): void {}\n}\n",
			":2:16",
			"a rest parameter that isn't an array",
		},
		{
			"parameters of a union",
			"function g1(a: string): void {}\nfunction g2(a: string, b: number): void {}\nfunction f(...args: Parameters<typeof g1 | typeof g2>): void {}\n",
			":3:12",
			"a rest parameter that isn't an array",
		},
		{
			"call through a function type",
			"function call(f: (...args: [string] | [string, number]) => void): void {\n    f(\"x\");\n}\n",
			":2:5",
			"a call through a function value whose rest parameter isn't an array",
		},
	} {
		for _, extension := range []string{".ts", ".a"} {
			test.Run(probe.name+extension, func(test *testing.T) {
				test.Parallel()
				defer func() {
					if failure := recover(); failure != nil {
						test.Fatalf("panicked: %v", failure)
					}
				}()
				_, err := lowerSourceAs(test, probe.source, extension)
				var notYet *NotYet
				if !errors.As(err, &notYet) ||
					notYet.What != probe.what ||
					!strings.HasSuffix(notYet.Where, probe.where) {
					test.Fatalf("want NotYet %q at %s, got %v", probe.what, probe.where, err)
				}
			})
		}
	}
}
