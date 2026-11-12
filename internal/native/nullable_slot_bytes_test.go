package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Absent storage has no evaluated value, even if a producer attaches a nullish
// contract. Its byte describes the layout without inspecting that missing value.
func TestAbsentNullableFieldMetadata(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		field    ir.Field
		byteText string
	}{
		{"null", ir.Field{Value: ir.ObjectLiteral{}, Nullish: ir.NullRepresentation}, "4"},
		{"undefined", ir.Field{Value: ir.ObjectLiteral{}, Nullish: ir.UndefinedRepresentation}, "4"},
		{"uninitialized_union", ir.Field{Value: ir.Undefined{Of: ir.Union}, Nullish: ir.NullRepresentation, Uninitialized: true}, "10"},
	} {
		t.Run(test.name, func(t *testing.T) {
			field := test.field
			field.Name, field.Absent = "target", true
			program := &ir.Program{Source: "absent nullable field", Main: []ir.Statement{
				ir.Evaluate{Value: ir.ObjectLiteral{Fields: []ir.Field{field}}},
			}}
			code := C(program)
			if !strings.Contains(code, ")[0] = "+test.byteText+";") {
				t.Fatalf("absent storage must keep its plain representation byte %s", test.byteText)
			}
			if strings.Contains(code, "adamic_object_new(") {
				t.Fatal("the absent field's value was evaluated")
			}
			binary := filepath.Join(t.TempDir(), "absent")
			if err := Build(code, binary, Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			if output, err := exec.CommandContext(t.Context(), binary).CombinedOutput(); err != nil || len(output) != 0 {
				t.Fatalf("reserved field: %v, output %q", err, output)
			}
		})
	}
}
