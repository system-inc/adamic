package lower

import (
	"os"
	"strings"
	"testing"
)

func TestLibraryArrayHeterogeneousCrashIsRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/library_array_refused/library_array_heterogeneous.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	if err == nil || !strings.Contains(err.Error(), "heterogeneous arrays need tagged elements") {
		t.Fatalf("want a slot representation refusal before C emission, got %v", err)
	}
}

func TestLibraryArrayGenericRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"detached", "const find = Array.prototype.indexOf; console.log(`${find(1)}`);", "method read as a value"},
		{"shadowed Array", "const Array = {prototype: {indexOf(value: number): number {return value;}}}; const find = Array.prototype.indexOf;", "method read as a value"},
		{"widened shape", "const hidden = {0: 1, 2: 9, length: 3}; const view: {0: number; length: number} = hidden; console.log(`${Array.prototype.indexOf.call(view, 9)}`);", "shape not proven"},
		{"reassigned shape", "let view = {0: 1, length: 3}; const bigger = {0: 1, 2: 9, length: 3}; view = bigger; console.log(`${Array.prototype.indexOf.call(view, 9)}`);", "shape not proven"},
		{"object coercion", "console.log(`${Array.prototype.indexOf.call({0: 1, length: {valueOf: (): number => 1}}, 1)}`);", "object ToPrimitive"},
		{"empty object field", "const value = {}; const receiver = {0: value, length: 1}; console.log(`${Array.prototype.indexOf.call(receiver, value)}`);", "field that can hide primitive slots"},
		{"optional field value", "function test(value: number | undefined): number {const receiver = {0: value, length: 1}; return Array.prototype.indexOf.call(receiver, undefined);} console.log(`${test(undefined)}`);", "field with nullable or undefined union values"},
		{"optional search value", "function test(value: string | undefined): number {return Array.prototype.indexOf.call({0: undefined, length: 1}, value);} console.log(`${test(undefined)}`);", "union search value"},
		{"hidden object field", "const array = [1]; const view: {length: number} = array; console.log(`${Array.prototype.indexOf.call({0: view, length: 1}, array)}`);", "field with an unproven runtime representation"},
		{"hidden search view", "const array = [1]; const view: {} = array; console.log(`${Array.prototype.indexOf.call({0: array, length: 1}, view)}`);", "object search view"},
		{"null array search", "const values: (string | undefined)[] = [undefined, \"a\"]; console.log(`${Array.prototype.indexOf.call(values, null)}`);", "distinct null and undefined reference slots"},
		{"null receiver", "console.log(`${Array.prototype.indexOf.call(null, 1)}`);", "catchable TypeError"},
		{"hidden array", "const value: {} = [1]; console.log(`${Array.isArray(value)}`);", "object view that can hide an array"},
		{"optional array", "function test(a: number[] | undefined): boolean {return Array.isArray(a);} console.log(`${test(undefined)}`);", "mixed or optional representation"},
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
