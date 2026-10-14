package native

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Re-emission must observe edited IR, while checked fields alone must not enable
// reflection's different shape keys and registration declarations.
func TestObjectMetadataCacheIsPerEmission(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "metadata", Main: []ir.Statement{
		ir.Evaluate{Value: ir.ObjectLiteral{Fields: []ir.Field{
			{Name: "value", Value: ir.NumberConstant{Value: 1}},
		}}},
	}}
	plain := C(program)
	program.CheckedFields = map[string]bool{"value": true}
	checked := C(program)
	if checked == plain || !strings.Contains(checked, "adamic_object_field_types(") {
		t.Fatal("new emission reused metadata from before the field contract was added")
	}
	if strings.Contains(checked, "adamic_register_shape_types(") {
		t.Fatal("field-type query was used as the dynamic-property cache key")
	}
	program.CheckedFields = nil
	if again := C(program); again != plain {
		t.Fatal("new emission reused metadata after the field contract was removed")
	}
}

// Checked writes need slot representation tags even without a checked read.
func TestObjectMetadataCheckedWritesIsPerEmission(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "metadata-writes", Main: []ir.Statement{
		ir.Evaluate{Value: ir.ObjectLiteral{Fields: []ir.Field{
			{Name: "value", Value: ir.NumberConstant{Value: 1}},
		}}},
	}}
	plain := C(program)
	program.CheckedWrites = map[string]bool{"value": true}
	checked := C(program)
	if !strings.Contains(checked, "adamic_object_field_types(") {
		t.Fatal("checked write lost its slot representation metadata")
	}
	if strings.Contains(checked, "adamic_register_shape_types(") {
		t.Fatal("checked write enabled dynamic-property metadata")
	}
	program.CheckedWrites = nil
	if again := C(program); again != plain {
		t.Fatal("new emission reused checked-write metadata after its removal")
	}
}
