# graphFlows' collect classifies every local writer: mutants

Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/lower -run '^(TestGraphLocalWritesAreLocalFieldsWrites|TestGraphUnknownWriters|TestGraphFlowPatternOverLiteralStops|TestGraphFlowCatchBindingStops)$' -count=1 -v`.
Run as Loom job J-drf7kyj on 09ef97d0 (#hny64km), each mutant applied alone and restored: mutant 1 turned
TestGraphFlowCatchBindingStops and TestGraphUnknownWriters red, mutant 2 turned TestGraphLocalWritesAreLocalFieldsWrites
red, and every pin passed restored.

collect takes every local-writing field by graph_flow.go's graphLocalWrites table, and graphUnknownWriters records
each local a node writes with a value collect doesn't follow, which a graph-typed demand of the local stops at.

## 1. The catch's unknown writer dropped

File: internal/lower/graph_flow.go, graphUnknownWriters. Delete:

```go
	case ir.Try:
		if value.HasCatch && value.CatchLocal >= 0 {
			unknown[value.CatchLocal+1] = "a graph value a catch binding takes, which the plan doesn't follow"
		}
```

Expected: TestGraphUnknownWriters goes red on its row "a catch's binding", which records nothing. If p2's program
reaches graphFlows, TestGraphFlowCatchBindingStops goes red too, since it lowers: the caught value stored into
head.next is planned nowhere. If it stops before graphFlows, as the probe will say, that test stays green and the IR
row is the mutant's witness.

## 2. The pattern's row dropped from the table

File: internal/lower/graph_flow.go, graphLocalWrites. Delete:

```go
	"Binding.Local":  "a pattern over an array literal is unknown, its elements' parts never seeded; over anything else, an element was planned where it was stored",
```

Expected: TestGraphLocalWritesAreLocalFieldsWrites goes red: localFields says Binding.Local writes a local, and
graphLocalWrites doesn't say how graphFlows' collect takes it.
