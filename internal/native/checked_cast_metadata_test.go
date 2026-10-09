package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckedCastWithoutReadSummary(t *testing.T) {
	t.Parallel()
	// Hand-built IR isolates the cast metadata obligation from receiver summaries
	// and Property.View, which now independently enable metadata in lowered p18.
	value := ir.NumberConstant{Value: 8}
	cast := ir.CheckedCast{CheckedFields: true, Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "value", Value: value}}}, Field: "value", FieldType: ir.Number, Allowed: []ir.Expression{value}, Message: "expected 8"}
	program := &ir.Program{Source: "checked cast without read summary", Main: []ir.Statement{ir.WriteLine{Value: ir.NumberToString{Value: ir.Property{Object: cast, Name: "value", Of: ir.Number}}}}}
	directory := t.TempDir()
	path := filepath.Join(directory, "source.a")
	if err := os.WriteFile(path, []byte("const object = { value: 8 }; console.log(String(object.value));\n"), 0600); err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", path)
	generated := C(program)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "native")
		if err := Build(generated, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: native %q; Node %q", sanitize, got, want)
		}
	}
}
