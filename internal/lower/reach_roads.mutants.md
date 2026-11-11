# ReachingAllocations' roads: mutants

Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/lower -run 'TestReach|TestFunctionReferenceFields' -count=1 -v`: the named tests go red and
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
			case false:
```

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
			case function.OptionalParameters[param]:
```

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

File: internal/lower/view_flow.go, newAllocationFlowGraph.

Old:

```go
		if !running[i] {
```

New:

```go
		if false {
```

Expected: TestReachUncalledWriterMakesNoWrite, TestReachSystemIsItsLiteral and
TestReachSystemThroughANamespaceBarrelIsItsLiteral go Unknown, "local or result has no tracked producer": the uncalled
writer's parameter has no caller.

## 3a. A call not counted as naming its function

File: internal/lower/view_flow.go, functionReferenceFields. Delete the `"Call.Function": 0,` line.

Expected: TestReachRunningWriterKeepsItsWrite/called reaches only the first literal, its called writer taken for one
that never runs and its write dropped, and TestReachSystemSetOnceIsBothLiterals loses the other literal the same way.
TestFunctionReferenceFieldsAreEveryFunctionField goes red too, "IR field Call.Function may hold a function".

## 3b. An export not counted as an entry

File: internal/lower/view_flow.go, functionReferenceFields. Delete the `"Export.Function": 0,` line.

Expected: TestReachRunningWriterKeepsItsWrite/exported_from_the_entry_module is known, the first literal alone, though
outside code can call the exported writer with anything. The field guard goes red as in 3a.

## 3c. A function value not counted as naming its function

File: internal/lower/view_flow.go, functionReferenceFields. Delete the `"MakeClosure.Function": 0,` line.

Expected: TestReachRunningWriterKeepsItsWrite/made_a_function_value is known, the first literal alone: the arrow, held
by a let, so no Local.ConstantClosure names it, is taken for a function that never runs. The field guard goes red as
in 3a.

Every other entry of functionReferenceFields, deleted, has the field guard as its witness: each one's field name
matches the guard's pattern. Program.AsyncEntry has no other: a program with top-level await moves Main into the
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
