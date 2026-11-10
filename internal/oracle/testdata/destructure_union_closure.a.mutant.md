# Restore the scalar-only Union gate

In internal/lower/collections.go, destructureFrom, replace
`slotless(stored) && stored != ir.Union` with
`slotless(stored) && !(stored == ir.Union && l.writable(fieldType))`.

The required field and the optional field with its arrow default hold a closure
member, which writable refuses. TestDestructureUnionAgrees/closure/native,
/sanitized, /javascript and /wasi stop with
"a destructured name held otherwise than its field" instead of agreeing with Node.

Run with ADAMIC_GATE_UNCACHED=1 and ADAMIC_ORACLE_WASI=1:

```
go test -count=1 -timeout 4h ./internal/oracle -run '^TestDestructureUnionAgrees$/closure$'
```

Use a source overlay restoring the gate to run this mutant; the admitted source
and its registered fixture stay unchanged on disk.
