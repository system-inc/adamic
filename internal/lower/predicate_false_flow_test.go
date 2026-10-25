package lower

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A flow node the walk doesn't know stops the read: the false-direction walk fails closed by
// construction, not by listing every step it refuses (#vdt9f1q). FlowFlagsShared alone names no
// kind of step. Mutant: the walk's default case passing turns this red.
func TestPredicateFalseWalkStopsAtAnUnknownFlowNode(t *testing.T) {
	t.Parallel()
	walk := &predicateFalseFlow{steps: map[predicateFalseKey]*predicateFalseStep{}}
	unknown := &ast.FlowNode{Flags: ast.FlowFlagsShared}
	step := walk.search(unknown, nil, nil, "")
	if step == nil || step.unknown != unknown {
		t.Fatalf("an unknown flow node passed the walk: %+v", step)
	}
}

// A loop's nodes finish together with one answer, and each is memoized, so what one node of the
// cycle reaches, every node of it reaches, and a later read stops at any of them. Mutant: each
// member memoized with its own partial answer, not the cycle's, answers back with nothing.
func TestPredicateFalseWalkFinishesALoopTogether(t *testing.T) {
	t.Parallel()
	walk := &predicateFalseFlow{steps: map[predicateFalseKey]*predicateFalseStep{}}
	start := &ast.FlowNode{Flags: ast.FlowFlagsStart}
	unknown := &ast.FlowNode{Flags: ast.FlowFlagsShared}
	loop := &ast.FlowNode{Flags: ast.FlowFlagsLoopLabel}
	body := &ast.FlowNode{Flags: ast.FlowFlagsCall, Antecedent: loop}
	back := &ast.FlowNode{Flags: ast.FlowFlagsArrayMutation, Antecedent: body}
	// The loop is entered from start and from its own body, which reaches the unknown node only
	// through a branch the loop's label comes back to.
	branch := &ast.FlowNode{Flags: ast.FlowFlagsBranchLabel, Antecedents: &ast.FlowList{Flow: back, Next: &ast.FlowList{Flow: unknown}}}
	loop.Antecedents = &ast.FlowList{Flow: start, Next: &ast.FlowList{Flow: branch}}
	if step := walk.search(body, nil, nil, ""); step == nil || step.unknown != unknown {
		t.Fatalf("the loop body missed the unknown node behind its label: %+v", step)
	}
	for _, member := range []*ast.FlowNode{loop, body, back, branch} {
		step, done := walk.steps[predicateFalseKey{flow: member}]
		if !done || step == nil || step.unknown != unknown {
			t.Fatalf("a member of the loop finished without its answer: %+v, %v", step, done)
		}
	}
	if step, done := walk.steps[predicateFalseKey{flow: start}]; !done || step != nil {
		t.Fatalf("start, outside the loop, took the loop's answer: %+v, %v", step, done)
	}
}
