# An exported entry function's parameters are unknown: mutant

Applied alone, then reverted. Run
`go test ./internal/lower -run '^(TestExportedEntryFunctionParameterIsUnknown|TestExportedIterableViewParameterIsNotProven)$' -count=1 -v`.
Not run yet: written on a machine that doesn't build Go, so "goes red" below is what the code says, not what a run
printed.

## 1. The marking removed

File: internal/lower/view_flow.go, the end of newAllocationFlowGraph.

Old:

```go
		for _, parameter := range program.Functions[export.Function].Parameters {
			graph.unknownParameters[parameter+1] = true
		}
```

New:

```go
		for _, parameter := range program.Functions[export.Function].Parameters {
			_ = parameter
		}
```

Expected: TestExportedEntryFunctionParameterIsUnknown/exported_from_the_entry_module is known, the inside literal
alone, though outside code can pass the exported function anything. TestExportedIterableViewParameterIsNotProven's
exported row lowers, as on main: the view over largest's parameter is taken as proven by the program's one call.
