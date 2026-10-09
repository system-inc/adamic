package main

import (
 "fmt"
 "github.com/system-inc/adamic/internal/ir"
)

func main() {
 throwing := ir.Program{Functions: []ir.Function{{MayThrow: true}}, Locals: []ir.Local{{}}}
 fmt.Printf("unknown all-throwing effect: %t\n", throwing.ClosureMayThrow(ir.CallClosure{Closure: ir.Read{Local: 0}}))
 slots := ir.Program{Functions: []ir.Function{{Parameters: []int{0,1}}}}
 slots.PrepareArgumentSlots()
 fmt.Printf("two-parameter fixed slots: %d\n", slots.FixedArgumentSlots)
 optional := ir.Program{Locals: []ir.Local{{Type: ir.Number},{Type:ir.MaybeNumber}}, Functions: []ir.Function{{Closure:true,Parameters:[]int{0}},{Closure:true,Parameters:[]int{1}}}, FunctionTypeTargets: map[int][]int{1:{0,1}}}
 fmt.Printf("mixed optional fixed layout: %+v\n", optional.ClosureArgumentLayout(ir.CallClosure{FunctionType:1}).Fixed)
}
