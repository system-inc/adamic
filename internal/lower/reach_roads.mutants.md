# ReachingAllocations' roads: mutants

Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/lower -run 'TestReach|TestFunctionReferenceFieldsAreEveryFunctionField|TestFunctionFieldTablesAgree' -count=1 -v`: the named tests go red and
every other test in reach_roads_test.go stays green. None of these has been run yet: they were written on a machine
that doesn't build Go, so each "goes red" below is what the code says, not what a run printed.

## 1a. An omitted optional or defaulted argument left unknown

File: internal/lower/view_flow.go, newAllocationFlowGraph's pass.

Old:

```go
			case function.OptionalParameters[param] && !slices.Contains(spread, true):
```

New:

```go
			case false && !slices.Contains(spread, true):
```

`case false:` alone does the same; this spelling keeps the line's shape.

Expected: TestReachOmittedArgumentTakesTheDefault and TestReachOmittedOptionalArgumentIsUndefined go Unknown,
"parameter may receive untracked or omitted arguments", and so does
TestReachOmittedArgumentStaysUnknownUnlessOptional/optional.

## 1b. A default's fallback not followed

File: internal/lower/view_flow.go, follow's Coalesce arm.

Old:

```go
			follow(value.Fallback)
```

New:

```go
			_ = value.Fallback
```

Expected: TestReachOmittedArgumentTakesTheDefault reaches only the passed literal, a known answer missing the default
the omitted call takes: the exact set catches it.

## 1c. An omitted argument taken as undefined under a spread

File: internal/lower/view_flow.go, newAllocationFlowGraph's pass.

Old:

```go
			case function.OptionalParameters[param] && !slices.Contains(spread, true):
```

New:

```go
			case function.OptionalParameters[param] && !slices.Contains(spread[:0], true):
```

`spread[:0]` is empty, so the test is never true, which drops the spread guard.

Expected: TestReachOmittedArgumentStaysUnknownUnlessOptional/optional_under_a_spread is known, the literal alone,
where the spread may put anything at that position.

## 1d. A required parameter's omitted argument taken as undefined

File: internal/lower/view_flow.go, newAllocationFlowGraph's pass.

Old:

```go
			case function.OptionalParameters[param] && !slices.Contains(spread, true):
```

New:

```go
			case !slices.Contains(spread, true):
```

Expected: TestReachOmittedArgumentStaysUnknownUnlessOptional/required is known, the literal alone.

## 2. undefined followed as an untracked source again

File: internal/lower/view_flow.go, follow.

Old:

```go
	case ir.Undefined, ir.Null:
```

New:

```go
	case ir.Null:
```

Expected: TestReachUndefinedAddsNoSite, TestReachSystemIsItsLiteral and TestReachSystemThroughANamespaceBarrelIsItsLiteral
go Unknown, "untracked source ir.Undefined".
Road 1's TS pins go red too, since an omitted argument arrives as undefined.

## 2n. A value that reaches no allocation taken as known

File: internal/lower/shape_flow.go, reachingAllocations.

Old:

```go
	if len(sites) == 0 {
```

New:

```go
	if false {
```

Expected: TestReachOnlyUndefinedStaysUnknown and TestReachNonNullAssertionOfOnlyUndefinedStaysUnknown are known with no
site. This is the guard road 2 leans on: undefined adds nothing, so only this line keeps "only undefined" from reading
as a proof.

## 3. A function that never runs collected anyway

File: internal/lower/view_flow.go, collectAllocationFlowGraph.

Old:

```go
		if !graph.running[i] {
```

New:

```go
		if false && !graph.running[i] {
```

The set is the graph's field, which checkFlowLiveness reads at the end of lowering, so the plain `if false {` compiles
too; this spelling keeps the line's shape.

Expected: TestReachUncalledWriterMakesNoWrite, TestReachSystemIsItsLiteral and
TestReachSystemThroughANamespaceBarrelIsItsLiteral go Unknown, "local or result has no tracked producer": the uncalled
writer's parameter has no caller.

## 3a. A call not counted as naming its function

File: internal/lower/view_flow.go, functionReferenceFields. Delete this line:

```go
	"Call.Function":          0,
```

Expected: TestReachRunningWriterKeepsItsWrite/called reaches only the first literal, its called writer taken for one
that never runs and its write dropped, and TestReachSystemSetOnceIsBothLiterals loses the other literal the same way.
TestFunctionReferenceFieldsAreEveryFunctionField goes red too, "IR field Call.Function may hold a function", and so
does TestFunctionFieldTablesAgree, since localFields says Call.Function is a function's index.

## 3b. An export not counted as an entry

File: internal/lower/view_flow.go, functionReferenceFields. Delete this line:

```go
	"Export.Function":    0,
```

Expected: TestReachRunningWriterKeepsItsWrite/exported_from_the_entry_module is known, the first literal alone, though
outside code can call the exported writer with anything. The field guard goes red too, "IR field Export.Function holds
an integer, in Export, which only this table covers", and so does TestFunctionFieldTablesAgree, which lists
Export.Function as outside localFields and so counted.

## 3c. A function value not counted as naming its function

File: internal/lower/view_flow.go, functionReferenceFields. Delete this line:

```go
	"MakeClosure.Function":           0,
```

Expected: TestReachRunningWriterKeepsItsWrite/made_a_function_value is known, the first literal alone: the arrow, held
by a let, so no Local.ConstantClosure names it, is taken for a function that never runs. The field guard and
TestFunctionFieldTablesAgree go red as in 3a.

Every other entry of functionReferenceFields, deleted, has the field guard as its witness: a field statements and
expressions hold matches the guard's name pattern, and every integer field of the other structs a program holds is
asked about whatever its name. Program.AsyncEntry has no other: a program with top-level await moves Main into the
function it names, so dropping it would drop the whole module's writes, and no pin here lowers an async module.

## 4. A non-null assertion's panic followed as a fallback

File: internal/lower/view_flow.go, follow's Coalesce arm.

Old:

```go
		if value.Panic == nil {
```

New:

```go
		if true {
```

Expected: TestReachNonNullAssertionHandsOnItsValue, TestReachSystemIsItsLiteral and
TestReachSystemThroughANamespaceBarrelIsItsLiteral go Unknown, "missing expression": the assertion's Fallback is nil.

## 5. Either operand of || or && not followed

File: internal/lower/view_flow.go, follow's Logical arm.

Old:

```go
		follow(value.Right)
```

New:

```go
		_ = value.Right
```

Expected: TestReachLogicalOrIsEitherOperand reaches only the first literal, a known answer missing the fallback the
right side gives where the left is undefined: the exact set catches it.

## 7. Road 3's late-reference check removed

File: internal/lower/view_flow.go, checkFlowLiveness.

Old:

```go
		if dead[function] && running[function] {
```

New:

```go
		if false && dead[function] && running[function] {
```

Expected: TestReachLateReferenceIsACompilerBug gets no error after Main calls the function the graph took to never
run. Deleting the call to checkFlowLiveness at the end of lower.go's lowerWithElementKey instead has no witness: no
program reaches a late reference today, which is what the check holds, so only a run that hit one would show it.

## 8. A one-based field read as an index

File: internal/lower/view_flow.go, functionReferenceFields.

Old:

```go
	"Local.NestedFunction":           1,
```

New:

```go
	"Local.NestedFunction":           0,
```

Expected: TestReachWriterNamedOnlyByAOneBasedField reaches only current's own literal: the local's NestedFunction, the
writer's index plus one, names the function after it, so the writer is taken for one that never runs and its write is
dropped. The writer is function 1, not 0, because every other local's NestedFunction is zero, which this mutant reads
as function 0; a writer at 0 would stay running and the pin would stay green.

## 9. An index read as one-based

File: internal/lower/view_flow.go, functionReferenceFields.

Old:

```go
	"Method.Function":                0,
```

New:

```go
	"Method.Function":                1,
```

Expected: TestReachWriterNamedOnlyByAMethod reaches only current's own literal, the method's writer taken for one that
never runs. TestFunctionFieldTablesAgree goes red too: localFields says Method.Function is a function's index, offset
zero.

## 10. An integer field outside localFields' reach left unclassified

File: internal/ir/ir.go, Class.

Old:

```go
	StaticParent int   // one-based hidden parent slot, zero for a root
```

New:

```go
	StaticParent int   // one-based hidden parent slot, zero for a root
	StaticInit   int
```

StaticInit is named outside the old name pattern, so a guard that only asked about function-looking names would stay
green, and no statement or expression holds a Class, so localFields doesn't ask either.

Expected: TestFunctionReferenceFieldsAreEveryFunctionField goes red, "IR field Class.StaticInit holds an integer, in
Class, which only this table covers". No other test in reach_roads_test.go reads it: runningFunctions reads only
the fields functionReferenceFields names.

## 11. A lowering pass builds its flow graph around the wrappers

File: internal/lower/host_member_write.go, in checkHostMemberWrites.

Old:

```go
	graph := l.flowGraph(l.result)
```

New:

```go
	graph := newAllocationFlowGraph(l.result)
```

Expected: TestEveryLoweringGraphNotesItsDead names host_member_write.go's call. The graph's dead functions go
unnoted, so a later pass naming one of them would leave checkFlowLiveness silent.

## 12. A narrowed read not followed

File: internal/lower/view_flow.go, follow's Defined arm.

Old:

```go
	case ir.Defined:
		// Road 6: a read the checker narrowed undefined (or null) out of is checked, and hands on Value
		// or stops, as graph_flow_classes.go's wrapper reads it.
		follow(value.Value)
```

New:

```go
	case ir.Defined:
		unknown("untracked source ir.Defined")
```

Expected: TestReachNarrowedReadHandsOnItsValue/.ts and /.a go Unknown, "untracked source ir.Defined", and
TestTsSysPresenceLowers/.a and TestTsSysPresenceAgreesWithNode/main.a refuse execute.a:19:6's presence test, as
368e1506 did before this road.

## 13. An early, valueless cell taken as a source

File: internal/lower/view_flow.go, collectAllocationFlowGraph's Declare arm.

Old:

```go
				if value.Uninitialized && value.Value == nil {
```

New:

```go
				if false && value.Uninitialized && value.Value == nil {
```

Expected: TestReachEarlyCellIsWhatItsDeclarationAssigns/uncaptured and /captured go Unknown, "missing expression",
and TestTsSysPresenceLowers/.ts and TestTsSysPresenceAgreesWithNode/main.ts refuse execute.ts:19:6's presence test,
as 368e1506 did before this road.
