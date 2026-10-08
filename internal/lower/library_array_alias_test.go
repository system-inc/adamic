package lower

import (
	"os"
	"strings"
	"testing"
)

func TestLibraryArrayThirdPassRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"reassigned intrinsic", "let f = Array.prototype.indexOf; f = (_value: number): number => 42; console.log(`${f.name}`);", "method read as a value"},
		{"escaped intrinsic", "const f = Array.isArray; const box = {f}; console.log(`${typeof box.f}`);", "method read as a value"},
		{"intrinsic identity", "const f = Array.isArray; console.log(`${f === Array.isArray}`);", "method read as a value"},
		{"intrinsic prototype write", "const f = Array.prototype.join; f.prototype = {};", "method read as a value"},
		{"intrinsic property compound write", "const f = Array.prototype.join; f.prototype += 1;", "method read as a value"},
		{"hoisted reader", "run(); const f = Array.prototype.includes; function run(): void { console.log(`${typeof f}`); }", "method read as a value"},
		{"exported intrinsic", "export const f = Array.isArray; console.log(`${typeof f}`);", "method read as a value"},
		{"chained alias", "const f = Array.isArray; const g = f; console.log(`${typeof g}`);", "method read as a value"},
		{"shadowed intrinsic", "const Array = {isArray(_value: number): boolean {return false;}}; const f = Array.isArray; console.log(`${typeof f}`);", "method read as a value"},
		{"valid sparse map length", "Array.prototype.map.call({length: 4294967295}, (): number => 1);", "proven immutable oversized length"},
		{"fraction below invalid bound", "Array.prototype.map.call({length: 4294967295.9}, (): number => 1);", "proven immutable oversized length"},
		{"mutable map length", "const receiver = {length: 4294967296}; receiver.length = 0; Array.prototype.map.call(receiver, (): number => 1);", "proven immutable oversized length"},
		{"shorthand map escape", "const receiver = {length: 4294967296}; const holder = {receiver}; holder.receiver.length = 0; Array.prototype.map.call(receiver, (): number => 1);", "proven immutable oversized length"},
		{"escaped map receiver", "const receiver = {length: 4294967296}; const alias = receiver; Array.prototype.map.call(alias, (): number => 1);", "proven immutable oversized length"},
		{"callback factory mutates length", "const receiver = {length: 4294967296}; Array.prototype.map.call(receiver, (() => {receiver.length = 0; return (): number => 1;})());", "proven immutable oversized length"},
		{"reassigned map receiver", "let receiver = {length: 4294967296}; receiver = {length: 0}; Array.prototype.map.call(receiver, (): number => 1);", "proven immutable oversized length"},
		{"nonexact pow proof", "Array.prototype.map.call({length: Math.pow(3, 25)}, (): number => 1);", "proven immutable oversized length"},
		{"optional reference never sort", "const values: (string | undefined)[] = [undefined, 'a']; values.sort((): never => {throw new Error('stop');});", "undefined partitioning"},
		{"optional reference numeric sort", "const values: (string | undefined)[] = [undefined, 'a']; values.sort((): number => 0);", "undefined partitioning"},
		{"optional reference copy sort", "const values: (string | undefined)[] = [undefined, 'a']; values.toSorted((): never => {throw new Error('stop');});", "optional elements"},
		{"function source with erased type", "Array.prototype.shift.call((): void => {});", "array-like or boxed receiver"},
		{"generator function receiver", "Array.prototype.shift.call(function* () {});", "generator function"},
		{"nonzero function length", "Array.prototype.shift.call((_value: number): void => {});", "array-like or boxed receiver"},
		{"escaping function receiver", "const receiver = (): void => {}; Array.prototype.shift.call(receiver);", "array-like or boxed receiver"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want explicit refusal containing %q, got %v", probe.reason, err)
			}
		})
	}
}

func TestLibraryArrayMapLengthEscapeRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/library_array_refused/library_array_map_length_escape.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	if err == nil || !strings.Contains(err.Error(), "proven immutable oversized length") {
		t.Fatalf("want a refusal for the mutable escaped receiver, got %v", err)
	}
}

func TestLibraryArrayFunctionTransformRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/library_array_refused/library_array_shift_transformed.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	if err == nil || !strings.Contains(err.Error(), "array-like or boxed receiver") {
		t.Fatalf("want refusal for unproven transformed function text, got %v", err)
	}
}
