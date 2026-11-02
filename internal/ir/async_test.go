package ir

import "testing"

func TestHasPromisesInExpressions(t *testing.T) {
	t.Parallel()
	observation := WriteLine{Value: TypeOf{Value: PromiseValue{Value: NumberConstant{Value: 3}}}}
	for _, program := range []*Program{
		{Main: []Statement{observation}},
		{Functions: []Function{{Body: []Statement{If{Condition: BooleanConstant{Value: true}, Then: []Statement{observation}}}}}},
	} {
		if !program.HasPromises() {
			t.Fatal("expression-only Promise must enable the runtime")
		}
	}
	if (&Program{Main: []Statement{WriteLine{Value: StringConstant{Index: 0}}}}).HasPromises() {
		t.Fatal("ordinary string expression must not enable the Promise runtime")
	}
}

func TestWatchFileRunsTheLoopWithoutPromises(t *testing.T) {
	t.Parallel()
	watch := NodeFSFile{Operation: "host_watch_file", Of: Number}
	for _, program := range []*Program{{Main: []Statement{Evaluate{Value: watch}}}, {Functions: []Function{{Body: []Statement{Evaluate{Value: watch}}}}}} {
		if !program.HasPromises() {
			t.Fatal("a retained watcher must enter the existing loop")
		}
	}
	if (&Program{Main: []Statement{Evaluate{Value: NodeFSFile{Operation: "host_unwatch_file", Of: Number}}}}).HasPromises() {
		t.Fatal("cancellation alone must not create loop work")
	}
}
