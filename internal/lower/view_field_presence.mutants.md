# view_field_presence.go's mutants

Each is planted alone on the head and run against the named test, which goes red. The Loom jobs that ran them are on #ya7938b.

| Mutant | Planted | Red test | Why only it |
|---|---|---|---|
| M1 | `viewFieldPresenceOperand` without its `truthinessOnly` test, admitting a read whose value is used | `TestViewFieldPresenceKeepsTheViewsRead/typeof`, and `/and_value` (internal/oracle) | typeof of the read then lowers through `methodPresenceTypeOf`, and a value-reading && through `methodPresenceAnd`'s structural arm, each agreeing with Node or stopping in other words, where the view's read stops with exit 70 |
| M2 | `presenceAnsweredOwn` answering true | `TestViewFieldPresenceKeepsItsCompileStops/e01`, and `/e04` and `/e05` (internal/oracle) | Object.prototype's `toString` on a Date is then read from Date's own slots, and a function's name and an array's length pass too, so each lowers or stops in other words than main's |
| M3 | native's Date keeping its time under `_fsFileTime` again (node_fs_file.c, native/fields.go) | `TestViewFieldPresenceAnswersNode` (internal/oracle) | the fixture's last line asks a Date for `_fsFileTime`, which native then finds and Node doesn't |
| M4 | `SlotDefined` emitted natively as `adamic_object_method_truthy` | `TestViewFieldPresenceAnswersNode` (internal/oracle) | the empty string and 0 are present and falsy, so `=== undefined` answers true for them natively |
