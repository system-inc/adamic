package lower

import (
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
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
		{"optional numeric constructor", "function make(value: number | undefined): number { const a = new Array<number | undefined>(value); return a.length; } console.log(`${make(3)}`);", "ambiguous numeric length overload"},
		{"mixed numeric constructor", "function make(value: number | string): number { const a = new Array<number | string>(value); return a.length; } console.log(`${make(3)}`);", "ambiguous numeric length overload"},
		{"sparse constructor length", "const values = new Array<number>(3); console.log(`${values.length}`);", "sparse presence"},
		{"constructor indexed extension", "const values = new Array(1, 2); values[2] = 3;", "outside a proven dense prefix"},
		{"constructor reassignment holes", "let values = new Array(1, 2); values = []; values[3] = 4;", "outside a proven dense prefix"},
		{"constructor integer property", "const values = new Array(1, 2); values[4294967295] = 4;", "outside a proven dense prefix"},
		{"generic boxed result", "Array.prototype.reverse.call(true);", "boxed result"},
		{"generic unknown shape", "Array.prototype.pop.call({0: 1, length: 1});", "array-like or boxed receiver"},
		{"default optional sort", "const values: (string | undefined)[] = [undefined, 'a']; values.sort();", "optional element"},
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

func TestLibraryArrayVisitUndefinedProof(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		element, values string
		allows          bool
	}{
		{"number", "1, 2", false}, {"number | undefined", "1, undefined", true},
		{"Leaf", "{value: 1}, {value: 2}", false}, {"Leaf | undefined", "{value: 1}, undefined", true},
		{"string", "'a', 'b'", false}, {"string | undefined", "'a', undefined", true},
	} {
		t.Run(test.element, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, "interface Leaf {readonly value: number;} const values: ("+test.element+")[] = ["+test.values+"]; console.log(`${values.findIndex((): boolean => false)}`);")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			walk(program.Main, func(node any) bool {
				if visit, ok := node.(ir.ArrayVisit); ok {
					found = true
					if visit.AllowsUndefined != test.allows {
						t.Fatalf("undefined proof is %t, want %t", visit.AllowsUndefined, test.allows)
					}
				}
				return true
			})
			if !found {
				t.Fatal("no search visit")
			}
		})
	}
}
