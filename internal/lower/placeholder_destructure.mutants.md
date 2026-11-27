# Mutants for a destructured declaration's placeholder slot (#s59xxn3)

One mutant, two edits, run on a house box through Loom beside the unmutated tree in the same job, then put back. It
runs `go test ./internal/oracle/ -run 'TestPlaceholderDestructuredBindingIsChecked|TestOptionalScalarRead'`.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The slot road dropped: walkPlaceholderSlots (placeholder.go) takes no object pattern, and placeholderSlotBinding (placeholder_destructure.go) answers no | `} else if target.Kind == ast.KindObjectBindingPattern && ` and `if binding.Kind != ast.KindBindingElement \|\| binding.Parent == nil` | `} else if false && target.Kind == ast.KindObjectBindingPattern && ` and `if true \|\| binding.Kind != ast.KindBindingElement \|\| binding.Parent == nil` | TestPlaceholderDestructuredBindingIsChecked/access_kind.ts, let_binding.ts and const_binding.ts, and TestOptionalScalarRead/u1_switch.ts: each stops at the destructured binding again, in #jjf9ppe's words |
