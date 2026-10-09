package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

func TestForOfHeaderBlockKinds(t *testing.T) {
	t.Parallel()
	function := lowerTypedFunctions(t, "loop.ts", `function useRows(rows) {
for (const row of rows) { for (const cell of row) { consume(cell); } }
}`)[0]
	Construct(function)
	loops := 0
	for _, block := range function.Blocks {
		loop, found := block.Terminal.(*ForOf)
		if !found {
			continue
		}
		loops++
		for _, blockId := range []static_single_assignment.BlockId{loop.Init, loop.Test} {
			header, found := function.Block(blockId)
			if !found || header.Kind != BlockKindLoop {
				t.Errorf("header block %d must have loop kind", blockId)
			}
		}
		continuation, found := function.Block(loop.Fallthrough)
		if !found || continuation.Kind != BlockKindBlock {
			t.Error("the loop continuation must remain an ordinary statement block")
		}
	}
	if loops != 2 {
		t.Fatalf("found %d loops, want two", loops)
	}
}

func TestForOfHeaderScopeAlignment(t *testing.T) {
	t.Parallel()
	for _, location := range []string{"initializer", "test", "before", "after"} {
		t.Run(location, func(t *testing.T) {
			t.Parallel()
			function := lowerTypedFunctions(t, "loop.ts", `function useRows(rows) {
before();
for (const row of rows) { row.read(); row.finish(); }
after();
}`)[0]
			Construct(function)
			var loop *ForOf
			for _, block := range function.Blocks {
				if terminal, found := block.Terminal.(*ForOf); found {
					loop = terminal
				}
			}
			if loop == nil {
				t.Fatal("no loop was lowered")
			}
			var target *Instruction
			for _, instruction := range function.Instructions {
				switch value := instruction.Value.(type) {
				case *GetIterator:
					if location == "initializer" {
						target = instruction
					}
				case *IteratorNext:
					if location == "test" {
						target = instruction
					}
				case *CallExpression:
					if calleeName(function, instruction, value.Callee) == location {
						target = instruction
					}
				}
			}
			if target == nil {
				t.Fatal("the fixture did not produce the target instruction")
			}
			original := mutation_aliasing.MutableRange{Start: target.Order, End: target.Order + 1}
			expected := original
			if location == "initializer" || location == "test" {
				body, _ := function.Block(loop.Loop)
				continuation, _ := function.Block(loop.Fallthrough)
				original.End = startingIdOf(function, body) + 1
				expected = mutation_aliasing.MutableRange{Start: TerminalOrder(loop), End: startingIdOf(function, continuation)}
			}
			identifier := target.LValue.Identifier
			scopes := &ReactiveScopes{
				byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{identifier: 1},
				ranges:       map[ScopeId]mutation_aliasing.MutableRange{1: original},
				members:      map[ScopeId][]static_single_assignment.IdentifierId{1: {identifier}},
				order:        []ScopeId{1},
			}
			aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
			if actual := aligned.RangeOf(1); actual != expected {
				t.Fatalf("range %v aligned to %v, want %v", original, actual, expected)
			}
			if violations := blockNestingViolations(scopeNestingItems(alignedRangeMap(aligned)), programBlockSubtrees(function)); violations != 0 {
				t.Errorf("aligned scope leaves %d nesting violations", violations)
			}
			BuildReactiveScopeTerminals(function, scopes, MergedScopeIdentity{Aligned: aligned, Merged: merged})
			instructions := 0
			for _, block := range function.Blocks {
				instructions += len(block.Instructions)
			}
			tree, result := BuildReactiveFunction(function)
			if tree == nil || result.Instructions != instructions || result.DoubleEmitted != 0 || result.UnmatchedGotos != 0 {
				t.Errorf("conversion must preserve all %d instructions without duplicate blocks or unmatched gotos: %+v", instructions, result)
			}
		})
	}
}

func TestForOfScopedHeadersPreserveInstructions(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"listener dispatch", `function notify() { for (const listener of listeners) listener(); }`},
		{"conditional dispatch", `function dispatch(channel, envelope) {
for (const subscriber of subscribers) {
if (!subscriber.channels.includes(channel)) continue;
subscriber.handlers.current?.onEnvelope(envelope);
}
}`},
		{"nested dispatch", `function dispatch(channels) {
for (const subscriber of subscribers) {
for (const channel of subscriber.channels) {
if (!channels.includes(channel)) continue;
subscriber.handlers.current?.onResync?.(channel);
}
}
}`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function := lowerTypedFunctions(t, "dispatch.ts", testCase.source)[0]
			Construct(function)
			scopes := AssignReactiveScopes(function)
			aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
			BuildReactiveScopeTerminals(function, scopes, MergedScopeIdentity{Aligned: aligned, Merged: merged})
			instructions := 0
			for _, block := range function.Blocks {
				instructions += len(block.Instructions)
			}
			tree, result := BuildReactiveFunction(function)
			if instructions == 0 || tree == nil {
				t.Fatal("the fixture must produce a nonempty graph and reactive tree")
			}
			if result.Instructions != instructions || result.DoubleEmitted != 0 || result.UnmatchedGotos != 0 {
				t.Errorf("conversion must preserve all %d instructions without duplicate blocks or unmatched gotos: %+v", instructions, result)
			}
		})
	}
}
