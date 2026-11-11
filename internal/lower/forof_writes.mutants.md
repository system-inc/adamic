# A for-of's and a catch's writes to their bindings: mutants

Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/lower -run '^(TestForOfReassignedBindingIsUnknown|TestForOfReassignedPatternBindingIsUnknown|TestForOfConstBindingStillDispatches|TestCatchBindingAssignedIsUnknown)$' -count=1 -v`.
Not run yet: written on a machine that doesn't build Go, so "goes red" below is what the code says, not what a run
printed.

## 1. The for-of's writes dropped

File: internal/lower/view_flow.go, newAllocationFlowGraph's collect. Delete the ForOf case whole:

```go
			case ir.ForOf:
				// Each step writes the element into Local, or destructures it into Pattern's locals,
				// Local unused then (a for-in is one too, over its object's keys).
				if len(value.Pattern) == 0 {
					graph.unknownWrites[value.Local+1] = "a for-of binding takes each element"
				}
				for _, binding := range value.Pattern {
					graph.unknownWrites[binding.Local+1] = "a for-of binding takes each element"
				}
```

Expected: TestForOfReassignedBindingIsUnknown and TestForOfReassignedPatternBindingIsUnknown are known, the
reassigned literal alone, and Main reads weight as a plain field, as on main. TestForOfConstBindingStillDispatches
stays green: a binding nothing assigns has no producer, Unknown either way. The oracle's forof_writes fixtures go red
on the JavaScript backend (undefined then 1, NaN for r2) and native (the missing-field panic).

## 2. The catch's write dropped

File: internal/lower/view_flow.go, newAllocationFlowGraph's collect. Delete the Try case whole:

```go
			case ir.Try:
				if value.CatchLocal >= 0 {
					graph.unknownWrites[value.CatchLocal+1] = "a catch binding takes what was thrown"
				}
```

Expected: TestCatchBindingAssignedIsUnknown is known, the literal its catch assigns alone.
