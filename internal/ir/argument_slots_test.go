package ir

import "testing"

func TestArgumentLayouts(test *testing.T) {
	test.Parallel()
	program := Program{
		Locals: []Local{{Type: Number}, {Type: MaybeNumber}, {Type: Array}},
		Functions: []Function{
			{Closure: true, Parameters: []int{0}},
			{Closure: true, Parameters: []int{0, 1}},
			{Closure: true, Parameters: []int{2}, RestElement: Number},
			{Closure: true, ReadsArguments: true},
		},
		FunctionTypeTargets: map[int][]int{1: {0, 1}, 2: {0, 2}, 3: {0, 3}, 4: {}},
	}
	program.PrepareArgumentSlots()
	optional := program.ClosureArgumentLayout(CallClosure{FunctionType: 1})
	if optional.Count || len(optional.Fixed) != 2 || optional.Fixed[1] != MaybeNumber {
		test.Fatalf("optional layout: %+v", optional)
	}
	rest := program.ClosureArgumentLayout(CallClosure{FunctionType: 2})
	if !rest.Count || len(rest.Rest) != 1 || rest.Rest[0].Start != 0 {
		test.Fatalf("mixed fixed/rest layout: %+v", rest)
	}
	reader := program.ClosureArgumentLayout(CallClosure{FunctionType: 3})
	if !reader.Count || program.RestArgumentSlots[rest.Rest[0]] < program.FixedArgumentSlots {
		test.Fatalf("reader layout: %+v", reader)
	}
	if program.ClosureArgumentLayout(CallClosure{FunctionType: 4}).Count {
		test.Fatal("a proven empty type has no reader")
	}
	unknown := program.ClosureArgumentLayout(CallClosure{Closure: Read{Local: 0}})
	if !unknown.Count {
		test.Fatal("unknown type must include the program's reader")
	}
	bounded := program.ClosureArgumentLayout(CallClosure{Closure: MakeClosure{Function: 0}})
	if bounded.Count {
		test.Fatal("a bounded unrelated function must not inherit a count")
	}
}
