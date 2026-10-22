package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

// Isolate the cast's metadata demand: checked calls and reads otherwise mask it.
func TestViewCheckedCastMetadata(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Main: []ir.Statement{ir.Evaluate{Value: ir.CheckedCast{CheckedFields: true, Value: ir.ObjectLiteral{}}}}}
	e := &emitter{program: program}
	if !e.fieldTypesNeeded() {
		t.Fatal("CheckedCast omitted field representation metadata")
	}
}
