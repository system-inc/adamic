package flow

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"testing"
)

func TestAsyncCompletionLeavesSynchronousFlowUnchanged(test *testing.T) {
	test.Parallel()
	body := []ir.Statement{ir.Try{
		Body:       []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 7}}},
		HasFinally: true,
		Finally:    []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 0}}},
	}}
	program := &ir.Program{
		Strings:   []string{"cleanup"},
		Functions: []ir.Function{{Name: "sync", Returns: ir.Number, Body: body}},
	}
	before := Build(program, 0)
	if normalizeAsyncError := NormalizeAsync(program); normalizeAsyncError != nil {
		test.Fatal(normalizeAsyncError)
	}
	if !reflect.DeepEqual(body, program.Functions[0].Body) {
		test.Fatal("synchronous IR changed")
	}
	after := Build(program, 0)
	if !reflect.DeepEqual(before, after) {
		test.Fatal("synchronous flow changed")
	}
}
