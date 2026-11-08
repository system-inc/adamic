# graphFlows' exhaustive follow: mutants (#y7wexf5)

Each mutant is one exact edit on this pass's commit, applied alone, then reverted. "Witness" is a test in
internal/oracle/graph_regions_test.go running a fixture in internal/oracle/testdata/graph_regions/ through
graphOracleWitness: Node agreement, the sanitized build, the JavaScript backend, the leak check, then the counted
build's balance. "Guard" is internal/lower/graph_flow_classified_test.go.

## 1. ArraySort's arm made silent: toSorted's copy goes unplanned

File: internal/lower/graph_flow_classes.go

Old:

```go
toSorted's copy is the ArraySlice inside", sources: graphFlowOperands(func(sort ir.ArraySort) []ir.Expression { return []ir.Expression{sort.Array} })},
```

New:

```go
toSorted's copy is the ArraySlice inside"},
```

Expected: TestGraphOracleGraphRegionsFlowToSorted leaks (the copy is counted, held by `a` and holding `b`, a cycle
through a counted array, as at main: Memory's 3 of 10 frees). The guard reds too: a wrapper that names no sources,
and TestGraphFlowSteps' "toSorted's sort enters its copy".

## 2. ArrayReverse's arm made silent: toReversed's copy goes unplanned

File: internal/lower/graph_flow_classes.go

Old:

```go
toReversed's copy is the ArraySlice inside", sources: graphFlowOperands(func(reverse ir.ArrayReverse) []ir.Expression { return []ir.Expression{reverse.Array} })},
```

New:

```go
toReversed's copy is the ArraySlice inside"},
```

Expected: TestGraphOracleGraphRegionsFlowToReversed leaks (Memory's 2 of 9 at main). The guard reds as in 1.

## 3. Coalesce's arm made silent: groupBy's buckets go unplanned

File: internal/lower/graph_flow_classes.go

Old:

```go
groupBy's bucket ?? [] plans the literal, never an existing bucket", sources: graphFlowOperands(func(coalesce ir.Coalesce) []ir.Expression {
		return []ir.Expression{coalesce.Value, coalesce.Fallback}
	})},
```

New:

```go
groupBy's bucket ?? [] plans the literal, never an existing bucket"},
```

Expected: TestGraphOracleGraphRegionsFlowGroupBy leaks (the bucket literal is counted, held by the planned map and
holding `b`; Memory's 5 of 15 at main). The guard reds as in 1, and on TestGraphFlowSteps' groupBy row.

## 4. The Coalesce join adopted: groupBy's bucket ?? [] planned as the join, not its arm

File: internal/native/emit_branches.go, the end of coalesce.

Old:

```go
	e.indent--
	e.line("}")
	if coalesce.Of.IsReference() {
		e.owned = append(e.owned, result)
	}
	return result
}
```

New:

```go
	e.indent--
	e.line("}")
	if coalesce.Of == ir.Array && len(e.program.GraphTypes) != 0 {
		e.line("%s = adamic_graph_adopt(%s, sizeof *%s);", result, result, result)
	}
	if coalesce.Of.IsReference() {
		e.owned = append(e.owned, result)
	}
	return result
}
```

Expected: TestGraphOracleGraphRegionsFlowGroupBy panics with "compiler bug: graph adoption needs a new allocation"
(runtime graph_regions.c:104) at the first element: the join is the literal its own plan already adopted, or an
existing bucket, a member, and adamic_graph_adopt takes neither. The mutant calls adamic_graph_adopt, the primitive
Memory named, on purpose: the emitter's own adoptGraph calls adamic_graph_adopt_owned, which returns a value that's
already a member unchanged (graph_regions.c:292), so adopting the join through it would pass silently whenever the
join holds a member. That's a runtime leniency, not a plan the guard can see; it's why the rule is that a wrapper is
never adopted, held by the class, not by the runtime.

## 5. RecordCall marked can't hold: Object.values' array goes unplanned

File: internal/lower/graph_flow_classes.go

Old:

```go
	reflect.TypeFor[ir.RecordCall](): {class: graphFlowPlanned, why:
```

New:

```go
	reflect.TypeFor[ir.RecordCall](): {class: graphFlowCannotHold, why:
```

Expected: TestGraphOracleGraphRegionsFlowRecordValues leaks (the values array is counted, held by `a` and holding
`b`; Memory's 6 of 13 at main). The guard reds too: RecordCall carries GraphTypes and isn't planned, and a can't-hold
class names sources.

IteratorNext marked can't hold (its class is stops) has no witness that lowers to leak through: Memory's row 12
(a held IteratorResult<Node>) stops earlier, refused as an inherited library member read as an own field. Until that
stop lifts, the mutant shows only as TestGraphFlowSteps' "an iterator's result stops" going red.

## 6. The default arm made silent

File: internal/lower/graph_flow_classes.go, graphFlowStepOf.

Old:

```go
		// The default arm: a type no rule names is a stop, never a silent return.
		return graphFlowStep{stop: true}
```

New:

```go
		// The default arm: a type no rule names is a stop, never a silent return.
		return graphFlowStep{}
```

Expected: TestGraphFlowFollowsEveryExpression reds: "an unclassified expression's step is {...}, want a stop".

## 7. A rule dropped: an IR type left unclassified

File: internal/lower/graph_flow_classes.go. Delete the ArraySort line (the one mutant 1 edits).

Expected: TestGraphFlowFollowsEveryExpression reds: "ir.ArraySort isn't classified for graph flow", and
TestGraphOracleGraphRegionsFlowToSorted stops at lowering, "a graph value through ir.ArraySort the plan doesn't
follow yet", at sortedByName, whose result is demanded, rather than leaking.

## 8. The values or entries array retained before its adoption

File: internal/native/emit_records.go, the end of the values and entries case.

Old:

```go
		e.adoptGraph(result, "sizeof *"+result, graph)
```

New:

```go
		e.line("adamic_retain(%s);", result)
		e.adoptGraph(result, "sizeof *"+result, graph)
```

Expected: TestRecordGraphArraysAdoptedUnretained (internal/native/record_graph_emission_test.go) reds, "the record
array ... is retained before it's adopted", for the values array and the entries array both. Run, the planned
witnesses would panic too: adamic_graph_adopt refuses an allocation whose count isn't 1 (graph_regions.c:104).

## 9. Entries pairs left counted

File: internal/native/emit_records.go, inside the entries walk.

Old:

```go
				e.adoptGraph(pair, fmt.Sprintf("adamic_object_allocation_size(%s)", pair), graph)
```

New: the line deleted.

Expected: TestRecordGraphArraysAdoptedUnretained reds, "the entries array ...'s pairs aren't adopted as they're
made", and TestGraphOracleGraphRegionsFlowRecordEntries leaks: each counted pair sits between the member array that
holds it and the member it holds, a cycle no count frees.
