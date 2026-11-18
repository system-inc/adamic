# Cycle finder nesting (#ykd89v9)

One edit, run on a house box or a sealed one, then put back.

## The enter forced ok

In `use` (cycles.go), keep the call and its deferred leave, and force the answer:

| Old text | New text |
|---|---|
| `	if !ok {` (the first one in `use`) | `	if ok = true; !ok {` |

Run:

```
go test ./internal/lower -run 'TestCyclesNesting' -count=1 -v -timeout 60s
```

The red is Go's test timeout, `panic: test timed out after 1m0s`, with every row of
`TestCyclesNestingStopsOnAGenericThatDeepensItself` still running, as on main before the bound.
`TestCyclesNestingControlLowers` never grows a type and passes before the timeout.
