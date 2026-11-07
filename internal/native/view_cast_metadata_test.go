package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

// Isolate the cast's metadata demand: checked calls and reads otherwise mask it.
func TestViewCheckedCastMetadata(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Main: []ir.Statement{ir.Evaluate{Value: ir.CheckedCast{CheckedFields: true, Value: ir.ObjectLiteral{}}}},
	}
	emitter := &emitter{program: program}
	if !emitter.fieldTypesNeeded() {
		test.Fatal("CheckedCast omitted field representation metadata")
	}
}
