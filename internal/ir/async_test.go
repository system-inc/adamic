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
