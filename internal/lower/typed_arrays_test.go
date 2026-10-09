package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestTypedArraysLower(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"Uint8Array", "Int32Array", "Float64Array"} {
		t.Run(kind, func(t *testing.T) {
			p, err := lowerSource(t, `const source: number[] = [300, -1, NaN];
const values = new `+kind+`(source);
values[0] = 300;
values.fill(-1, 1);
values.set(values.subarray(0, 1), 2);
for (const value of values) { console.log(value.toString()); }
console.log((values[values.length] ?? -999).toString());
`)
			if err != nil {
				t.Fatal(err)
			}
			if !p.Locals[1].Type.IsTypedArray() {
				t.Fatalf("typed array represented as %v", p.Locals[1].Type)
			}
			write, ok := p.Main[2].(ir.SetIndex)
			if !ok || write.Element != ir.Number {
				t.Fatalf("index write: %#v", p.Main[2])
			}
		})
	}
}

func TestTypedArrayGaps(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, message string }{
		{`const b = new ArrayBuffer(4);`, "ArrayBuffer"},
		{`const b = new SharedArrayBuffer(4);`, "sharing typed arrays across parallel tasks"},
		{`function size(b: SharedArrayBuffer): number { return b.byteLength; }`, "sharing typed arrays across parallel tasks"},
		{`const v = new DataView(new ArrayBuffer(4));`, "DataView"},
		{`const b = new ArrayBuffer(4, {maxByteLength: 8});`, "resizable ArrayBuffer"},
		{`const v = new Int8Array(4);`, "typed array element type Int8Array"},
		{`const v = new Uint8Array(4); const b = new Uint8Array(v.buffer);`, "another view's buffer"},
		{`const v = new Uint8Array(4); v.set(new Int32Array(2));`, "different element type"},
		{`const v = new Uint8Array(4); console.log(v.slice(1).length.toString());`, "typed array method slice"},
	} {
		_, err := lowerSource(t, probe.source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(err.Error(), probe.message) {
			t.Errorf("%s: want %q, got %v", probe.source, probe.message, err)
		}
	}
}

func TestTypedArrayViewsCannotChangeRepresentation(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const narrow = new Uint8Array(2); const shape: {readonly length: number} = narrow;",
		"const narrow = new Uint8Array(2); const shape = narrow as {readonly length: number};",
		"const source = {data: new Uint8Array(2)}; const wide: {readonly data: {readonly length: number}} = source;",
	} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) {
			t.Errorf("representation change accepted: %s: %v", source, err)
		}
	}
}
