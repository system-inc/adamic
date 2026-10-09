package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

func main() {
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Object, Function: 0}, {Type: ir.Object, Function: 0}},
		Functions: []ir.Function{{Name: "borrowedHolder", Closure: true, Parameters: []int{0}, Body: []ir.Statement{
			ir.Declare{Local: 1, Value: ir.ObjectLiteral{}},
			ir.Evaluate{Value: ir.NodeFSFile{Operation: "is_file", Of: ir.Boolean, Arguments: []ir.Expression{ir.Read{Local: 1, Of: ir.Object}}}},
			ir.SetProperty{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "saved", Value: ir.Read{Local: 0, Of: ir.Object}, Site: 1},
			ir.Return{},
		}}},
	}
	for _, write := range fresh.ProveWrites(program) {
		fmt.Printf("site=%d proven=%v kind=%d why=%s\n", write.Site, write.Proven, write.Kind, write.Why)
	}
}
