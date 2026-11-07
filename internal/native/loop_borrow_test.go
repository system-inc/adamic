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

func loopProgram(test *testing.T, path string) *ir.Program {
	test.Helper()
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	return program
}

func TestLoopBorrowPlan(test *testing.T) {
	test.Parallel()
	program := loopProgram(test, "../oracle/testdata/borrow_loop.a")
	plan, lending := planElementBorrows(program)
	want := map[string]bool{
		"popBody":             false,
		"spliceBody":          false,
		"storeBody":           false,
		"callBody":            false,
		"closureBody":         false,
		"reassignBody":        false,
		"pushBody":            true,
		"throwBody":           true,
		"caughtBody":          true,
		"bindingReuse":        true,
		"assignedBinding":     false,
		"capturedBinding":     false,
		"virtualBody":         false,
		"freshBody":           false,
		"closureReassignBody": false,
		"globalReassignBody":  false,
		"throwHeldBody":       false,
		"parameterBody":       true,
	}
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
						test.Errorf("%s: borrowed %t, want %t", function.Name, plan[&list[index]], expected)
					}
					if expected && (!program.Locals[loop.Local].Borrowed || !lending[loop.Iterable.(ir.Read).Local]) {
						test.Errorf("%s: missing binding or lending fact", function.Name)
					}
				}
				walkStatement(statement, func(ir.Expression) {}, statements)
			}
		}
		statements(function.Body)
		if found != 1 {
			test.Errorf("%s: found %d loops", function.Name, found)
		}
		delete(want, function.Name)
	}
	if len(want) != 0 {
		test.Fatalf("missing functions: %v", want)
	}
}

func TestNbodyBorrowedLoopC(test *testing.T) {
	test.Parallel()
	program := loopProgram(test, "../../bench/nbody.ts")
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
					if !program.Locals[loop.Local].Borrowed ||
						strings.Contains(generated, "adamic_release("+name+")") ||
						strings.Contains(generated, name+" = adamic_retain(") {
						test.Errorf("function %d %s still counts binding %s", index, function.Name, name)
					}
				}
				walkStatement(statement, func(ir.Expression) {}, statements)
			}
		}
		statements(function.Body)
	}
}

func TestGlobalArgumentLending(test *testing.T) {
	test.Parallel()
	program := loopProgram(test, "../oracle/testdata/borrow_global_call.a")
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
				test.Errorf(
					"%s virtual %d argument %d lent %t, want %t",
					program.Functions[call.Function].Name,
					call.Virtual,
					index,
					lent,
					want,
				)
			}
			checked++
		}
	})
	if checked < 4 {
		test.Fatalf("checked only %d global arguments", checked)
	}
}

// Inspect the loop's actual iterator temporary, so retaining the binding or releasing another
// array cannot hide a missing hold. The throw cases check both borrowed and owned cleanup.
func TestLoopArrayHoldC(test *testing.T) {
	test.Parallel()
	for _, path := range []string{"../oracle/testdata/borrow_loop.a", "../../bench/nbody.ts"} {
		program := loopProgram(test, path)
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
				test.Fatalf("%s: missing function definition", function.Name)
			}
			body := strings.SplitN(generated[at[1]:], "\n}\n", 2)[0]
			iterator := regexp.MustCompile(`< (adamic_temporary_[A-Za-z0-9_]+)->length`).FindStringSubmatch(body)
			if iterator == nil {
				test.Fatalf("%s: missing array iterator", function.Name)
			}
			counted := strings.Contains(body, "adamic_release("+iterator[1]+");")
			if counted != expected {
				test.Errorf("%s: iterator owns count %t, want %t", function.Name, counted, expected)
			}
			if !expected {
				initialization := regexp.MustCompile(iterator[1] + ` = ([^;]+);`).FindStringSubmatch(body)
				if initialization == nil || strings.Contains(initialization[1], "adamic_retain(") {
					test.Errorf("%s: redundant iterator retain", function.Name)
				}
			}
			checked++
		}
		minimum := 11
		if strings.HasSuffix(path, "nbody.ts") {
			minimum = 2
		}
		if checked != minimum {
			test.Fatalf("%s: checked %d functions, want %d", path, checked, minimum)
		}
	}
}

// TestLoopCallCoverage is the call boundary of borrowed loop bindings and iterator holds.
// owns is true when the iterator keeps its own count: the body can take an element, or the
// array is not a stable local or parameter. False means the element proof borrowed the binding
// and the iterator takes no count. scanLent is false even though its callee replaces a global;
// the caller's retain, not the iterator, is what keeps that array alive, and the lend check
// below is what requires the retain.
func TestLoopCallCoverage(test *testing.T) {
	test.Parallel()
	program := loopProgram(test, "../oracle/testdata/borrow_loop_calls.a")
	plan, lending := planElementBorrows(program)
	owns := map[string]bool{
		"readLocal": false, "readParameter": false, "pushLocal": false,
		"continueAfterRead": false, "throwAfterRead": false,
		"reportQuiet": false, "scanCopy": false, "scanLent": false,
		"fieldLocal": false,
		"popLocal":   true, "popParameter": true, "popAlias": true,
		"popInBranch": true, "popInTry": true,
		"spreadLocal": true, "spreadParameter": true, "mapLocal": true,
		"replaceLocal": true, "earlyReturn": true, "breakAfterPop": true,
		"throwAfterPop": true, "sameStatement": true, "stashThenPop": true,
		"fieldDirect": true, "concreteReader": true, "virtualEater": true,
	}
	generated := C(program)
	reuse := planReuse(program, lending)
	caller := emitter{program: program, reuse: reuse}
	lentQuiet, lentScan := 0, 0
	walkExpressions(program, func(expression ir.Expression) {
		call, ok := expression.(ir.Call)
		if !ok || len(call.Arguments) != 1 {
			return
		}
		read, ok := call.Arguments[0].(ir.Read)
		if !ok || !program.Locals[read.Local].Global {
			return
		}
		_, lent := caller.lentArgument(call, 0)
		switch program.Functions[call.Function].Name {
		case "reportQuiet":
			lentQuiet++
			if !lent {
				test.Errorf("reportQuiet should borrow the global; the body never writes it")
			}
		case "scanLent":
			lentScan++
			if lent {
				test.Errorf("scanLent lent the global its body replaces")
			}
		}
	})
	if lentQuiet != 1 || lentScan != 1 {
		test.Fatalf("global calls: quiet %d, scan %d", lentQuiet, lentScan)
	}
	checked := 0
	for index, function := range program.Functions {
		expected, check := owns[function.Name]
		if !check {
			continue
		}
		found := 0
		var statements func([]ir.Statement)
		statements = func(list []ir.Statement) {
			for at := range list {
				loop, ok := list[at].(ir.ForOf)
				if !ok {
					walkStatement(list[at], func(ir.Expression) {}, statements)
					continue
				}
				found++
				borrowed := plan[&list[at]]
				if borrowed == expected {
					test.Errorf("%s: borrowed %t, want %t", function.Name, borrowed, !expected)
				}
				if !expected && (!program.Locals[loop.Local].Borrowed || !lending[loop.Iterable.(ir.Read).Local]) {
					test.Errorf("%s: missing binding or lending fact", function.Name)
				}
			}
		}
		statements(function.Body)
		if found != 1 {
			test.Errorf("%s: found %d loops", function.Name, found)
		}
		name := (&emitter{program: program}).functionName(index)
		definition := regexp.MustCompile(`(?m)^static [^\n]+ ` + name + `\([^\n]*\) \{\n`)
		at := definition.FindStringIndex(generated)
		if at == nil {
			test.Fatalf("%s: missing function definition", function.Name)
		}
		body := strings.SplitN(generated[at[1]:], "\n}\n", 2)[0]
		iterator := regexp.MustCompile(`< (adamic_temporary_[A-Za-z0-9_]+)->length`).FindStringSubmatch(body)
		if iterator == nil {
			test.Fatalf("%s: missing array iterator", function.Name)
		}
		counted := strings.Contains(body, "adamic_release("+iterator[1]+");")
		if counted != expected {
			test.Errorf("%s: iterator owns count %t, want %t", function.Name, counted, expected)
		}
		if !expected {
			initialization := regexp.MustCompile(iterator[1] + ` = ([^;]+);`).FindStringSubmatch(body)
			if initialization == nil || strings.Contains(initialization[1], "adamic_retain(") {
				test.Errorf("%s: redundant iterator retain", function.Name)
			}
		}
		checked++
		delete(owns, function.Name)
	}
	if len(owns) != 0 || checked != 26 {
		test.Fatalf("missing functions: %v (checked %d)", owns, checked)
	}
}
