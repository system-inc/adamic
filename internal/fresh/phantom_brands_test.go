package fresh_test

import (
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestPhantomMemberEvaluatesOperandWithoutEscaping(t *testing.T) {
	for _, optional := range []bool{false, true} {
		writes := fresh.ProveWrites(regexTreeProgram(ir.PhantomMember{Value: ir.Read{Local: 0, Of: ir.Object}, Name: "brand", Optional: optional}))
		if len(writes) != 1 || !writes[0].Proven {
			t.Fatalf("optional %v: %+v", optional, writes)
		}
	}
}
