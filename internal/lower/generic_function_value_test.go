package lower

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestGenericFunctionValuesHaveSeparateInstances(test *testing.T) {
	test.Parallel()
	source, err := os.ReadFile("../oracle/testdata/generic_function_value.a")
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowerSource(test, string(source))
	if err != nil {
		test.Fatal(err)
	}
	instances, forwarders := map[ir.Type]int{}, map[ir.Type]int{}
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "identity_") {
			if function.Closure {
				forwarders[function.Returns]++
			} else {
				instances[function.Returns]++
			}
		}
	}
	for _, representation := range []ir.Type{ir.String, ir.Number, ir.Object, ir.Array} {
		if instances[representation] != 1 || forwarders[representation] != 1 {
			test.Fatalf(
				"representation %v: instances %d, forwarders %d; want one of each",
				representation,
				instances[representation],
				forwarders[representation],
			)
		}
	}
}

// Distinct reference shapes share a pointer representation, but must keep their
// checker identities. A mutant deleting the type identity must fail this count.
func TestGenericFunctionValueReferenceInstances(test *testing.T) {
	test.Parallel()
	program, err := lowerSource(test, `
function identity<T>(x: T): T { return x; }
const first: (x: { readonly left: string }) => { readonly left: string } = identity;
const second: (x: { readonly right: number }) => { readonly right: number } = identity;
first({ left: 'a' });
second({ right: 3 });
`)
	if err != nil {
		test.Fatal(err)
	}
	instances := 0
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "identity_") && !function.Closure {
			instances++
		}
	}
	if instances != 2 {
		test.Fatalf("got %d instances for two reference types, want 2", instances)
	}
}

func TestGenericFunctionValueIdentityComparisonIsNotYet(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `
function identity<T>(x: T): T { return x; }
const first: (x: string) => string = identity;
const second: (x: string) => string = identity;
console.log(first === second ? 'same' : 'different');
`)
	if err == nil || !strings.Contains(err.Error(), "function identity comparison") {
		test.Fatalf("want identity comparison NotYet, got %v", err)
	}
}

func TestGenericFunctionValueDifferentReferenceRepresentations(test *testing.T) {
	test.Parallel()
	program, err := lowerSource(test, `
function identity<T>(x: T): T { return x; }
const first: (x: { readonly left: string }) => { readonly left: string } = identity;
const second: (x: readonly number[]) => readonly number[] = identity;
first({ left: 'a' });
second([3]);
`)
	if err != nil {
		test.Fatal(err)
	}
	instances := 0
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "identity_") && !function.Closure {
			instances++
		}
	}
	if instances != 2 {
		test.Fatalf("got %d instances for two reference representations, want 2", instances)
	}
}

func TestGenericFunctionValueEarlierIdentityComparisonIsNotYet(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `
function compare(a: (x: string) => string, b: (x: string) => string): boolean { return a === b; }
function identity<T>(x: T): T { return x; }
const first: (x: string) => string = identity;
const second: (x: string) => string = identity;
console.log(compare(first, second) ? 'same' : 'different');
`)
	if err == nil || !strings.Contains(err.Error(), "function identity comparison") {
		test.Fatalf("want earlier identity comparison NotYet, got %v", err)
	}
}

func TestGenericFunctionValueIdentityCallsAreNotYet(test *testing.T) {
	test.Parallel()
	for _, expression := range []string{
		"Object.is(first, second)",
		"[first].includes(second)",
		"[first].indexOf(second)",
		"new Set([first, second])",
		"new Map([[first, 1], [second, 2]])",
	} {
		test.Run(expression, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, `
function identity<T>(x: T): T { return x; }
const first: (x: string) => string = identity;
const second: (x: string) => string = identity;
`+"const observed = "+expression+`;`)
			if err == nil || !strings.Contains(err.Error(), "function identity observation") {
				test.Fatalf("want identity observation NotYet, got %v", err)
			}
		})
	}
}

func TestGenericFunctionValueUnionIdentityIsNotYet(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `
function compare(a: ((x: string) => string) | string, b: ((x: string) => string) | string): boolean { return a === b; }
function identity<T>(x: T): T { return x; }
const first: (x: string) => string = identity;
console.log(compare(first, first) ? 'same' : 'different');
`)
	if err == nil || !strings.Contains(err.Error(), "function identity comparison") {
		test.Fatalf("want union identity comparison NotYet, got %v", err)
	}
}

func TestGenericFunctionValueNamespaceDirectCallsKeepIdentity(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `namespace Values {
 export function identity<T>(value: T): T { return value; }
 export function read(value: number): number { return value; }
 }
 const copy = Values.read;
 console.log(String(Values.identity(2)) + ' ' + String(copy === Values.read));`)
	if err != nil {
		test.Fatal(err)
	}
}
