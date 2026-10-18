package fresh_test

import (
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

func TestIteratorFieldAccessorEscapesReceiver(t *testing.T) {
	t.Parallel()
	protocolAccessorEscapesReceiver(t, ir.IteratorField{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "value", Of: ir.Number})
}

func TestIteratorMethodAccessorEscapesReceiver(t *testing.T) {
	t.Parallel()
	protocolAccessorEscapesReceiver(t, ir.IteratorMethod{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "next"})
}

// An accessor can publish its receiver. A later outside value stored into that
// receiver can reach it again, so confinement must no longer prove the write.
func protocolAccessorEscapesReceiver(t *testing.T, lookup ir.Expression) {
	t.Helper()
	program := &ir.Program{
		Locals: []ir.Local{{Name: "receiver", Type: ir.Object, Function: -1}, {Name: "outside", Type: ir.Object, Global: true, Function: -1}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.ObjectLiteral{}},
			ir.Evaluate{Value: lookup},
			ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "saved", Value: ir.Read{Local: 1, Of: ir.Object}, Site: 1},
		},
	}
	writes := fresh.ProveWrites(program)
	if len(writes) != 1 || writes[0].Kind != fresh.WriteField || writes[0].Proven {
		t.Fatalf("protocol accessor must leave the receiver escaped: %+v", writes)
	}
}
