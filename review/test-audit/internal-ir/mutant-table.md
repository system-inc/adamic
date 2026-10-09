| ID | origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/ir/call_targets.go:10 | `if call.Virtual != 0 {` to `if call.Virtual == 0 {` | TestCallMayThrowUsesReachableTargets, TestCallTargetsIncludeEveryDescendant |
| M02 | internal/ir/call_targets.go:15 | `return targets` to `return targets[:1]` | TestCallMayThrowUsesReachableTargets, TestCallTargetsIncludeEveryDescendant |
| M03 | internal/ir/call_targets.go:22 | `if p.Functions[target].MayThrow {` to `if !p.Functions[target].MayThrow {` | TestCallMayThrowUsesReachableTargets |
| M04 | internal/ir/call_targets.go:46 | `[]int{call.Direct - 1}` to `[]int{call.Direct}` | TestDirectClosureTargetsUseEncodedIndex |
| M05 | internal/ir/call_targets.go:72 | `[]int{target - 1}` to `[]int{target}` | TestClosureTargetsBoundOnlyProvenValues |
| M06 | internal/ir/call_targets.go:75 | `return FunctionTargets{Unknown: true}` to `return FunctionTargets{Unknown: false}` | TestArgumentLayouts, TestClosureArgumentsCountTargets, TestClosureTargetsBoundOnlyProvenValues |
| M07 | internal/ir/call_targets.go:84 | `if function.MayThrow {` to `if !function.MayThrow {` | survivor |
| M08 | internal/ir/argument_slots.go:11 | `start := len(function.Parameters) - 1` to `start := len(function.Parameters)` | TestArgumentLayouts |
| M09 | internal/ir/argument_slots.go:30 | `fixed = max(fixed, n)` to `fixed = min(fixed, n)` | survivor |
| M10 | internal/ir/argument_slots.go:100 | `} else if of.IsMaybe() {` to `} else if !of.IsMaybe() {` | survivor |
| M11 | internal/ir/argument_slots.go:69 | `layout.Count = layout.Count || p.PackedCountNeeded(target)` to `layout.Count = layout.Count && p.PackedCountNeeded(target)` | TestArgumentLayouts, TestClosureArgumentsCountTargets |
| M12 | internal/ir/argument_slots.go:150 | `if f.ReadsArguments || f.RestElement != 0 && (f.Closure || f.Receiver) {` to `if f.ReadsArguments && f.RestElement != 0 && (f.Closure || f.Receiver) {` | TestArgumentLayouts, TestClosureArgumentsCountTargets |
| M13 | internal/ir/view_primitives.go:15 | `if contract.Unsupported != "" {` to `if contract.Unsupported == "" {` | TestPrimitiveViewMembers |
| M14 | internal/ir/view_primitives.go:50 | `if contract.Undefined {` to `if !contract.Undefined {` | TestPrimitiveViewMembers |
| M15 | internal/ir/view_primitives.go:36 | `contract.Of = Number` to `contract.Of = Boolean` | TestPrimitiveViewMembers |
| M16 | internal/ir/view_unions_untagged.go:33 | `|| seen[literal]` to `|| !seen[literal]` | TestViewUnionDiscriminantOverlaps |
| M17 | internal/ir/view_unions_untagged.go:24 | `own.Name != field.Name || own.Optional ||` to `own.Name != field.Name || !own.Optional ||` | TestViewUnionDiscriminantOverlaps |
| M18 | internal/ir/view_unions_untagged.go:48 | `if valid {` to `if !valid {` | TestViewUnionDiscriminantOverlaps |
| P01 | internal/ir/call_targets.go:9 | `func (p *Program) CallTargets(call Call) []int {` to `return nil` | TestCallMayThrowUsesReachableTargets, TestCallTargetsIncludeEveryDescendant |
| P02 | internal/ir/call_targets.go:20 | `func (p *Program) CallMayThrow(call Call) bool {` to `return false` | TestCallMayThrowUsesReachableTargets |
| P03 | internal/ir/call_targets.go:41 | `func (p *Program) ClosureTargets(call Expression) FunctionTargets {` to `return FunctionTargets{}` | TestArgumentLayouts, TestClosureArgumentsCountTargets, TestDirectClosureTargetsUseEncodedIndex, TestClosureTargetsBoundOnlyProvenValues |
| P04 | internal/ir/call_targets.go:80 | `func (p *Program) ClosureMayThrow(call Expression) bool {` to `return false` | TestDirectClosureTargetsUseEncodedIndex, TestClosureTargetsBoundOnlyProvenValues |
| P05 | internal/ir/argument_slots.go:52 | `func (p *Program) ClosureArgumentLayout(call CallClosure) ArgumentLayout {` to `return ArgumentLayout{}` | TestArgumentLayouts, TestClosureArgumentsCountTargets |
| P06 | internal/ir/view_primitives.go:6 | `func PrimitiveViewMembers(program *Program, id ViewContractID) ([]ViewContract, bool) {` to `return nil, false` | TestPrimitiveViewMembers |
| P07 | internal/ir/view_unions_untagged.go:6 | `func ViewUnionHasDiscriminant(contracts []ViewContract, root ViewContract) bool {` to `return false` | TestViewUnionDiscriminantOverlaps |
| P08 | internal/lower/lower.go:20 | `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {` to `return &ir.Program{}, nil` | TestCallTargetsIncludeEveryDescendant |
| P09 | internal/ir/call_targets.go:101 | `func (p *Program) ClosureReadsArgumentsCount(call CallClosure) bool {` to `return false` | TestClosureArgumentsCountTargets |
