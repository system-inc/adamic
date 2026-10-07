package lower

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestFinishClassCallsPreservesAccessorTargets(t *testing.T) {
	t.Parallel()
	program := &ir.Program{MethodTargets: map[int][]int{4: {4, 5}}, Classes: []ir.Class{{Methods: []int{1}}}}
	lowering := &lowering{result: program}
	lowering.finishClassCalls()
	lowering.finishClassCalls()
	if got := program.CallTargets(ir.Call{Function: 4, Virtual: -1}); !reflect.DeepEqual(got, []int{4, 5}) {
		t.Fatalf("accessor targets changed after class discovery: %v", got)
	}
	if got := program.CallTargets(ir.Call{Function: 1, Virtual: 1}); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("class targets missing or duplicated: %v", got)
	}
}
