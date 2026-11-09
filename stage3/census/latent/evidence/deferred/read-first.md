# Production deferred refusal contract

At base `018b7c21ea7499e53113007984a5773a0b24e0d8`, `Lower` reaches the
module loop in `internal/lower/lower.go`. It finishes prepasses, then calls
`lowering.refuse` for each module in module order. Any returned error stops the
program before declaration registration or body lowering.

`refuse` resets `l.refusal.deferred` immediately before its registry visitor.
The visitor follows production traversal and registry ordinal order, retaining
the first ordinary refusal. After the complete walk, only `found == nil` permits
selection of `l.refusal.deferred`. This fallback is per module, not per program
and not per census unit. The earliest deferred candidate is retained. An ordinary
refusal later in the module still wins. A deferred failure in an earlier module
can stop production before any later module is visited.

`refuseNode` deliberately leaves the deferred selection to its caller.
`refuseInSlot` saves the enclosing deferred state, resets it, runs `refuseNode`,
selects the fallback only when no ordinary refusal returned, then restores it.

A search for assignments to `refusal.deferred` finds exactly one refusal rule:
`refuseAnyRelation`, registered at ordinal 3270 in `internal/lower/any_relations.go`.
It records either a member typed any related to a typed member, or the 64-pair
structural-depth limit. Other assignments in `refusals.go` only reset or restore
the state. There is no second deferring refusal today.

The census continuing clone clears its ordinary binding after recording each
finding. Its inherited final `found == nil` fallback could consequently record
a deferred result beside higher-priority refusals. The clone now leaves that
selection to the driver, which checks the scan's findings count. Independent
units use the production node walk with an isolated refusal state and prepared
unit context. Their deferred phase precedes body lowering; an ordinary refusal
anywhere in that unit suppresses the fallback. No production lowering source or
lowering field is added or renamed.
