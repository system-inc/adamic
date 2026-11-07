package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNestedEmptyArrayElementKinds(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, source string
		kind         ir.Type
	}{
		{"first", `const cases = [[], [1]];`, ir.Number},
		{"last", `const cases = [[1], []];`, ir.Number},
		{"typed", `const cases: number[][] = [[]];`, ir.Number},
		{"strings", `const cases = [[], ['text']];`, ir.String},
		{"objects", `const cases = [[{ text: 'text' }], []];`, ir.Object},
		{"union", `const cases: (number[] | string[])[] = [[]];`, ir.Number},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			program, lowerSourceError := lowerSource(test, probe.source)
			if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
			outer := program.Main[0].(ir.Declare).Value.(ir.ArrayLiteral)
			found := false
			for _, value := range outer.Elements {
				literal := value.(ir.ArrayLiteral)
				if len(literal.Elements) == 0 {
					found = true
					if literal.Element != probe.kind && !(probe.name == "union" && literal.Element == ir.String) {
						test.Fatalf("empty element kind %v, want %v", literal.Element, probe.kind)
					}
				}
			}
			if !found {
				test.Fatal("empty literal absent")
			}
		})
	}
}

func TestEmptyLiteralGenericReturnUsesSamePath(test *testing.T) {
	test.Parallel()
	program, lowerSourceError := lowerSource(test, `function empty<T>(): T[] { return [] as T[]; }
interface Item { readonly text: string; }
empty<number>(); empty<Item>();`)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	kinds := map[ir.Type]int{}
	for _, function := range program.Functions {
		literal := function.Body[0].(ir.Return).Value.(ir.ArrayLiteral)
		kinds[literal.Element]++
	}
	if kinds[ir.Number] != 1 || kinds[ir.Object] != 1 || len(kinds) != 2 {
		test.Fatalf("generic empty kinds: %v", kinds)
	}
}
