package native

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func loopProgram(t *testing.T, path string) *ir.Program {
	t.Helper()
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestLoopBorrowPlan(t *testing.T) {
	program := loopProgram(t, "../oracle/testdata/borrow_loop.a")
	plan, lending := planElementBorrows(program)
	want := map[string]bool{"popBody": false, "spliceBody": false, "storeBody": false, "callBody": false, "closureBody": false, "reassignBody": false, "pushBody": true, "throwBody": true, "caughtBody": true, "bindingReuse": true, "assignedBinding": false, "capturedBinding": false, "virtualBody": false, "freshBody": false, "closureReassignBody": false, "globalReassignBody": false, "throwHeldBody": false, "parameterBody": true}
	for _, function := range program.Functions {
		expected, check := want[function.Name]
		if !check {
			continue
		}
		found := 0
		var statements func([]ir.Statement)
		statements = func(list []ir.Statement) {
			for index, statement := range list {
				if loop, ok := statement.(ir.ForOf); ok {
					found++
					if plan[&list[index]] != expected {
						t.Errorf("%s: borrowed %t, want %t", function.Name, plan[&list[index]], expected)
					}
					if expected && (!program.Locals[loop.Local].Borrowed || !lending[loop.Iterable.(ir.Read).Local]) {
						t.Errorf("%s: missing binding or lending fact", function.Name)
					}
				}
				walkStatement(statement, func(ir.Expression) {}, statements)
			}
		}
		statements(function.Body)
		if found != 1 {
			t.Errorf("%s: found %d loops", function.Name, found)
		}
		delete(want, function.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing functions: %v", want)
	}
}

func TestNbodyBorrowedLoopC(t *testing.T) {
	program := loopProgram(t, "../../bench/nbody.ts")
	generated := C(program)
	for index, function := range program.Functions {
		if function.Name != "advance" && function.Name != "offsetMomentum" {
			continue
		}
		emitter := emitter{program: program}
		var statements func([]ir.Statement)
		statements = func(list []ir.Statement) {
			for _, statement := range list {
				if loop, ok := statement.(ir.ForOf); ok {
					name := emitter.localName(loop.Local)
					if !program.Locals[loop.Local].Borrowed || strings.Contains(generated, "adamic_release("+name+")") || strings.Contains(generated, name+" = adamic_retain(") {
						t.Errorf("function %d %s still counts binding %s", index, function.Name, name)
					}
				}
				walkStatement(statement, func(ir.Expression) {}, statements)
			}
		}
		statements(function.Body)
	}
}

func TestGlobalArgumentLending(t *testing.T) {
	program := loopProgram(t, "../oracle/testdata/borrow_global_call.a")
	emitter := emitter{program: program, reuse: planReuse(program, nil)}
	checked := 0
	walkExpressions(program, func(expression ir.Expression) {
		call, ok := expression.(ir.Call)
		if !ok {
			return
		}
		for index, argument := range call.Arguments {
			read, ok := argument.(ir.Read)
			if !ok || !program.Locals[read.Local].Global || read.Of != ir.Object {
				continue
			}
			want := program.Functions[call.Function].Name == "read" && call.Virtual == 0
			for _, later := range call.Arguments[index+1:] {
				want = want && pure(later)
			}
			_, lent := emitter.lentArgument(call, index)
			if lent != want {
				t.Errorf("%s virtual %d argument %d lent %t, want %t", program.Functions[call.Function].Name, call.Virtual, index, lent, want)
			}
			checked++
		}
	})
	if checked < 4 {
		t.Fatalf("checked only %d global arguments", checked)
	}
}

// Inspect the loop's actual iterator temporary, so retaining the binding or releasing another
// array cannot hide a missing hold. The throw cases check both borrowed and owned cleanup.
func TestLoopArrayHoldC(t *testing.T) {
	for _, path := range []string{"../oracle/testdata/borrow_loop.a", "../../bench/nbody.ts"} {
		program := loopProgram(t, path)
		generated := C(program)
		want := map[string]bool{
			"reassignBody": true, "closureReassignBody": true, "globalReassignBody": true,
			"freshBody": true, "throwHeldBody": true, "popBody": true, "virtualBody": true,
			"pushBody": false, "throwBody": false, "caughtBody": false, "parameterBody": false,
			"advance": false, "offsetMomentum": false,
		}
		checked := 0
		for index, function := range program.Functions {
			expected, check := want[function.Name]
			if !check {
				continue
			}
			name := (&emitter{program: program}).functionName(index)
			definition := regexp.MustCompile(`(?m)^static [^\n]+ ` + name + `\([^\n]*\) \{\n`)
			at := definition.FindStringIndex(generated)
			if at == nil {
				t.Fatalf("%s: missing function definition", function.Name)
			}
			body := strings.SplitN(generated[at[1]:], "\n}\n", 2)[0]
			iterator := regexp.MustCompile(`< (adamic_temporary_[0-9]+)->length`).FindStringSubmatch(body)
			if iterator == nil {
				t.Fatalf("%s: missing array iterator", function.Name)
			}
			counted := strings.Contains(body, "adamic_release("+iterator[1]+");")
			if counted != expected {
				t.Errorf("%s: iterator owns count %t, want %t", function.Name, counted, expected)
			}
			if !expected {
				initialization := regexp.MustCompile(iterator[1] + ` = ([^;]+);`).FindStringSubmatch(body)
				if initialization == nil || strings.Contains(initialization[1], "adamic_retain(") {
					t.Errorf("%s: redundant iterator retain", function.Name)
				}
			}
			checked++
		}
		minimum := 11
		if strings.HasSuffix(path, "nbody.ts") {
			minimum = 2
		}
		if checked != minimum {
			t.Fatalf("%s: checked %d functions, want %d", path, checked, minimum)
		}
	}
}
