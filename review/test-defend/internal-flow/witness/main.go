package main

import (
 "fmt"
 "github.com/system-inc/adamic/internal/flow"
)

func main() {
 f := flow.NewFunction("overwrite-before-read")
 block := f.NewBlock()
 f.Entry = block.Id
 block.Terminal = &flow.Return{}
 x := flow.Place{Identifier: f.NewIdentifier("x", 1).Id}
 f.AddInstruction(block, &flow.Instruction{Defines: []flow.Place{x}})
 f.AddInstruction(block, &flow.Instruction{Defines: []flow.Place{x}})
 f.AddInstruction(block, &flow.Instruction{Uses: []flow.Place{x}})
 fmt.Printf("x live after first write, before overwrite: %v\n", flow.LiveOut(f)[0][1])
}
