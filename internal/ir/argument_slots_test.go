package ir

import "testing"

func TestArgumentLayouts(t *testing.T) {
	p := Program{
		Locals: []Local{{Type: Number}, {Type: MaybeNumber}, {Type: Array}},
		Functions: []Function{
			{Closure: true, Parameters: []int{0}},
			{Closure: true, Parameters: []int{0, 1}},
			{Closure: true, Parameters: []int{2}, RestElement: Number},
			{Closure: true, ReadsArguments: true},
		},
		FunctionTypeTargets: map[int][]int{1: {0, 1}, 2: {0, 2}, 3: {0, 3}, 4: {}},
	}
	p.PrepareArgumentSlots()
	optional := p.ClosureArgumentLayout(CallClosure{FunctionType: 1})
	if optional.Count || len(optional.Fixed) != 2 || optional.Fixed[1] != MaybeNumber {
		t.Fatalf("optional layout: %+v", optional)
	}
	rest := p.ClosureArgumentLayout(CallClosure{FunctionType: 2})
	if !rest.Count || len(rest.Rest) != 1 || rest.Rest[0].Start != 0 {
		t.Fatalf("mixed fixed/rest layout: %+v", rest)
	}
	reader := p.ClosureArgumentLayout(CallClosure{FunctionType: 3})
	if !reader.Count || p.RestArgumentSlots[rest.Rest[0]] < p.FixedArgumentSlots {
		t.Fatalf("reader layout: %+v", reader)
	}
	if p.ClosureArgumentLayout(CallClosure{FunctionType: 4}).Count {
		t.Fatal("a proven empty type has no reader")
	}
	unknown := p.ClosureArgumentLayout(CallClosure{Closure: Read{Local: 0}})
	if !unknown.Count {
		t.Fatal("unknown type must include the program's reader")
	}
	bounded := p.ClosureArgumentLayout(CallClosure{Closure: MakeClosure{Function: 0}})
	if bounded.Count {
		t.Fatal("a bounded unrelated function must not inherit a count")
	}
}
