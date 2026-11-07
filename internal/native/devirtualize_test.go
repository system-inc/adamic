package native

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestDevirtualizedCalls(test *testing.T) {
	test.Parallel()
	source, loadError := load.Load([]string{"../oracle/testdata/devirtualize.a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), source)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	e := &emitter{program: program}
	single, multiple, exact := 0, 0, 0
	walkExpressions(program, func(expression ir.Expression) {
		call, ok := expression.(ir.Call)
		if !ok || call.Virtual <= 0 {
			return
		}
		arguments := make([]string, len(call.Arguments))
		for index := range arguments {
			arguments[index] = fmt.Sprintf("argument_%d", index)
		}
		code := e.callCode(call, arguments)
		targets := program.CallTargets(call)
		switch {
		case len(targets) == 1:
			single++
			if !strings.HasPrefix(code, e.functionName(targets[0])+"(") {
				test.Errorf("single target is not direct: %s", code)
			}
		case e.exactReceiverClass(call.Arguments[0]) != 0:
			exact++
			if strings.Contains(code, "adamic_virtual") {
				test.Errorf("exact receiver still dispatches: %s", code)
			}
		default:
			multiple++
			if !strings.Contains(code, "adamic_virtual") {
				test.Errorf("unbounded receiver lost dispatch: %s", code)
			}
		}
	})
	if single < 4 || multiple == 0 || exact == 0 {
		test.Fatalf("missing call cases: single=%d multiple=%d exact=%d", single, multiple, exact)
	}
	code := C(program)
	found := false
	for local, declared := range program.Locals {
		if declared.Name != "single" {
			continue
		}
		function, known := e.exactReceiverMethod(ir.Read{Local: local, Of: ir.Object}, "run")
		if !known {
			test.Fatal("interface receiver lost its exact allocation")
		}
		// A direct adapter invocation in the generated body, beyond its declaration.
		if !strings.Contains(
			code,
			" = "+stableName("adamic_method", e.sourceReadable(function), e.functionKey(function, map[int]bool{}))+"(",
		) {
			test.Fatal("interface call still looks up its method")
		}
		found = true
	}
	if !found {
		test.Fatal("missing interface receiver")
	}
}

func TestExactReceiverRejectsAssignments(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Locals:    []ir.Local{{Type: ir.Object}},
		Functions: []ir.Function{{}},
		Classes:   []ir.Class{{Constructor: 0}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.Call{Function: 0, Returns: ir.Object}},
			ir.Try{Body: []ir.Statement{ir.Assign{Local: 0, Value: ir.Undefined{Of: ir.Object}}}},
		},
	}
	e := &emitter{program: program}
	if got := e.exactReceiverClass(ir.Read{Local: 0, Of: ir.Object}); got != 0 {
		test.Fatalf("assigned receiver taken as exact: %d", got)
	}
}
