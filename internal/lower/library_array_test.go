package lower

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestLibraryArrayHeterogeneousSlotsAreTagged(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/library_array_refused/library_array_heterogeneous.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	walk(program.Main, func(node any) bool {
		if array, ok := node.(ir.ArrayLiteral); ok {
			found = true
			if array.Element != ir.Union {
				t.Fatalf("heterogeneous slots must be tagged, got %v", array.Element)
			}
			for _, item := range array.Elements {
				if item.Type() != ir.Union {
					t.Fatalf("untagged element: %v", item.Type())
				}
			}
		}
		return true
	})
	if !found {
		t.Fatal("missing heterogeneous array witness")
	}
}

func TestLibraryArrayGenericRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"detached", "const find = Array.prototype.indexOf; console.log(`${find(1)}`);", "unbound Array method call"},
		{"shadowed Array", "const Array = {prototype: {indexOf(value: number): number {return value;}}}; const find = Array.prototype.indexOf;", "method read as a value"},
		{"widened shape", "const hidden = {0: 1, 2: 9, length: 3}; const view: {0: number; length: number} = hidden; console.log(`${Array.prototype.indexOf.call(view, 9)}`);", "shape not proven"},
		{"reassigned shape", "let view = {0: 1, length: 3}; const bigger = {0: 1, 2: 9, length: 3}; view = bigger; console.log(`${Array.prototype.indexOf.call(view, 9)}`);", "shape not proven"},
		{"object coercion", "console.log(`${Array.prototype.indexOf.call({0: 1, length: {valueOf: (): number => 1}}, 1)}`);", "object ToPrimitive"},
		{"empty object field", "const value = {}; const receiver = {0: value, length: 1}; console.log(`${Array.prototype.indexOf.call(receiver, value)}`);", "field with an unrepresented or mixed value"},
		{"optional field value", "function test(value: number | undefined): number {const receiver = {0: value, length: 1}; return Array.prototype.indexOf.call(receiver, undefined);} console.log(`${test(undefined)}`);", "field with nullable or undefined union values"},
		{"optional search value", "function test(value: string | undefined): number {return Array.prototype.indexOf.call({0: undefined, length: 1}, value);} console.log(`${test(undefined)}`);", "union search value"},
		{"hidden object field", "const array = [1]; const view: {length: number} = array; console.log(`${Array.prototype.indexOf.call({0: view, length: 1}, array)}`);", "field with an unrepresented or mixed value"},
		{"hidden search view", "const array = [1]; const view: {} = array; console.log(`${Array.prototype.indexOf.call({0: array, length: 1}, view)}`);", "generic array search with mixed representations"},
		{"null array search", "const values: (string | undefined)[] = [undefined, \"a\"]; console.log(`${Array.prototype.indexOf.call(values, null)}`);", "distinct null and undefined reference slots"},
		{"null receiver", "console.log(`${Array.prototype.indexOf.call(null, 1)}`);", "catchable TypeError"},
		{"hidden array", "const value: {} = [1]; console.log(`${Array.isArray(value)}`);", "erased object/any/unknown view"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want refusal containing %q, got %v", probe.reason, err)
			}
		})
	}
}

// Compiler's runtime brand test replaces the old optional-array limitation.
func TestLibraryArrayOptionalBrand(t *testing.T) {
	t.Parallel()
	if _, err := lowerSource(t, "function test(a: number[] | undefined): boolean {return Array.isArray(a);} console.log(`${test(undefined)} ${test([1])}`);"); err != nil {
		t.Fatal(err)
	}
}
