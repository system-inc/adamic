package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func main() {
	program := &ir.Program{Source: "checked-read-survivor", Locals: []ir.Local{{Name: "pending", Type: ir.Number, Global: true, Uninitialized: true, Function: -1}}}
	program.Main = []ir.Statement{ir.WriteLine{Stream: ir.Stdout, Value: ir.Read{Local: 0, Of: ir.Number, Checked: true}}}
	fmt.Print(javascript.JavaScript(program))
}
