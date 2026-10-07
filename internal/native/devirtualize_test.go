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

func TestDevirtualizedCalls(t *testing.T) {
	t.Parallel()
	source, err := load.Load([]string{"../oracle/testdata/devirtualize.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), source)
	if err != nil {
		t.Fatal(err)
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
				t.Errorf("single target is not direct: %s", code)
			}
		case e.exactReceiverClass(call.Arguments[0]) != 0:
			exact++
			if strings.Contains(code, "adamic_virtual") {
				t.Errorf("exact receiver still dispatches: %s", code)
			}
		default:
			multiple++
			if !strings.Contains(code, "adamic_virtual") {
				t.Errorf("unbounded receiver lost dispatch: %s", code)
			}
		}
	})
	if single < 4 || multiple == 0 || exact == 0 {
		t.Fatalf("missing call cases: single=%d multiple=%d exact=%d", single, multiple, exact)
	}
	code := C(program)
	found := false
	for local, declared := range program.Locals {
		if declared.Name != "single" {
			continue
		}
		function, known := e.exactReceiverMethod(ir.Read{Local: local, Of: ir.Object}, "run")
		if !known {
			t.Fatal("interface receiver lost its exact allocation")
		}
		// A direct adapter invocation in the generated body, beyond its declaration.
		if !strings.Contains(code, " = "+stableName("adamic_method", (&emitter{program: program}).sourceReadable(function), (&emitter{program: program}).functionKey(function, map[int]bool{}))+"(") {
			t.Fatal("interface call still looks up its method")
		}
		found = true
	}
	if !found {
		t.Fatal("missing interface receiver")
	}
}

func TestExactReceiverRejectsAssignments(t *testing.T) {
	t.Parallel()
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
		t.Fatalf("assigned receiver taken as exact: %d", got)
	}
}
