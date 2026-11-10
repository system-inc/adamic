# Restore the scalar-only Union gate

In destructureFrom, restore
`slotless(stored) && !(stored == ir.Union && l.writable(fieldType))`.
TestDestructureUnionAgrees/semver stops with
"a destructured name held otherwise than its field" instead of agreeing with Node.
Run TestDestructureUnionAgrees with ADAMIC_GATE_UNCACHED=1.
