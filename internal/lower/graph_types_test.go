package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestGraphLiteralMethodAllocation(t *testing.T) {
	program, err := lowerSource(t, `function make():void {let holder:{read():number}|undefined;const value={read():number{return holder===undefined?0:1;}};holder=value;console.log('made');}make();`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("graph types %v", program.GraphTypes)
	found := false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if literal, ok := node.(ir.ObjectLiteral); ok {
				found = true
				if !program.IsGraph(literal.GraphTypes) {
					t.Errorf("cyclic method holder allocation is counted: types %v", literal.GraphTypes)
				}
			}
			return true
		})
	}
	if !found {
		t.Fatal("missing method holder allocation")
	}
}
