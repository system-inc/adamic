package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestCollectFieldWrites(t *testing.T) {
	t.Parallel()
	static := ir.StringConstant{Index: 0}
	heap := ir.Concat{Parts: []ir.Expression{static, static}}
	program := &ir.Program{Main: []ir.Statement{ir.Declare{Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "label", Value: static}}}}}, Functions: []ir.Function{{Body: []ir.Statement{ir.Try{Finally: []ir.Statement{ir.If{Then: []ir.Statement{ir.SetProperty{Name: "label", Value: heap}}}}}}}}}
	writes := collectFieldWrites(program)["label"]
	if !writes.Complete || len(writes.Expressions) != 2 {
		t.Fatalf("missing nested/aliased write: %+v", writes)
	}
	if _, ok := writes.Expressions[1].(ir.Concat); !ok {
		t.Fatal("heap write lost")
	}
	program.Functions = append(program.Functions, ir.Function{Body: []ir.Statement{ir.Evaluate{Value: ir.ObjectLiteral{Spread: ir.Read{Of: ir.Object}}}}})
	if collectFieldWrites(program)["label"].Complete {
		t.Fatal("unknown spread source accepted")
	}
	if collectFieldWrites(program)["kind"].Complete {
		t.Fatal("runtime field accepted")
	}
}
