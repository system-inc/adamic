# A for-of's and a catch's writes to their bindings: mutants

Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/lower -run '^(TestForOfReassignedBindingIsUnknown|TestForOfReassignedPatternBindingIsUnknown|TestForOfConstBindingStillDispatches|TestCatchBindingAssignedIsUnknown|TestLocalWritesAreLocalFieldsWrites|TestForOfReassignedSpreadSourceIsNotPlain|TestForOfReassignedBindingIsNoPlainSpreadSource)$' -count=1 -v`.
Not run yet: written on a machine that doesn't build Go, so "goes red" below is what the code says, not what a run
printed.

collect takes every local-writing field by view_flow.go's flowLocalWrites table, so each mutant drops one row. Each also
turns TestLocalWritesAreLocalFieldsWrites red, since the table no longer has every write row localFields lists; the
pins below are what show the hole it would reopen. A for-of over a pattern writes its locals through Binding.Local,
and one without through ForOf.Local, so one row can't turn both r1 and r2 red: 1a is r1's and 1b is r2's.

## 1a. The for-of's binding dropped

File: internal/lower/view_flow.go, flowLocalWrites. Delete:

```go
	"ForOf.Local":    "a for-of binding takes each element",
```

Expected: TestForOfReassignedBindingIsUnknown is known, the reassigned literal alone, and Main reads weight as a plain
field, as on main. The oracle's forof_writes r1 fixtures go red on the JavaScript backend (undefined then 1) and native
(the missing-field panic). TestForOfConstBindingStillDispatches stays green: a binding nothing assigns has no
producer, Unknown either way.

## 1b. The for-of pattern's bindings dropped

File: internal/lower/view_flow.go, flowLocalWrites. Delete:

```go
	"Binding.Local":  "a for-of binding takes each element",
```

Expected: TestForOfReassignedPatternBindingIsUnknown is known, the reassigned literal alone, and the r2 fixtures go red
(NaN on the JavaScript backend).

## 2. The catch's write dropped

File: internal/lower/view_flow.go, flowLocalWrites. Delete:

```go
	"Try.CatchLocal": "a catch binding takes what was thrown",
```

Expected: TestCatchBindingAssignedIsUnknown is known, the literal its catch assigns alone.

## 3. library_spread's reader back on the parameter map alone

File: internal/lower/library_spread.go, plainSpreadSources.

Old:

```go
		if _, unknownSource := graph.unknownAt(node); active[node] || unknownSource || len(graph.sources[node]) == 0 {
```

New:

```go
		if _, unknownSource := graph.unknownAt(node); active[node] || unknownSource && graph.unknownParameters[node] || len(graph.sources[node]) == 0 {
```

That is main's reader, `graph.unknownParameters[node]`, kept in a spelling that compiles: dropping `unknownSource`
outright would leave it declared and not used.

Expected: TestForOfReassignedSpreadSourceIsNotPlain lowers with no refusal, and
TestForOfReassignedBindingIsNoPlainSpreadSource, the same asked of plainSpreadSources in IR, takes the binding as
plain: the for-of binding's sources are the literal alone, proven plain, as on main.
