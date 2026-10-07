package ir

import "testing"

func TestClosureArgumentsCountTargets(t *testing.T) {
	p := Program{Functions: []Function{{}, {ArgumentsCount: 1, ReadsArguments: true}}, FunctionTypeTargets: map[int][]int{10: {0}, 11: {0, 1}}}
	for _, probe := range []struct {
		kind  int
		reads bool
	}{{10, false}, {11, true}, {12, true}} {
		if got := p.ClosureReadsArgumentsCount(CallClosure{FunctionType: probe.kind}); got != probe.reads {
			t.Fatalf("type %d: %v, want %v", probe.kind, got, probe.reads)
		}
	}
}
