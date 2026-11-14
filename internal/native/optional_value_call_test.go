package native

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// optionalValueCallOutput is what testdata/optional_value_call.a prints once its seven calls are
// callback?.(...), by JavaScript's rules: with a function, the argument runs before the callback's
// body; with undefined, no argument runs, so no argument line prints and ran stays where it was, and
// relay's result is undefined. hold's callee is read before clear('e') sets cb to undefined, so the
// value read is called, and cb is undefined after. local's and parameter's first calls read cb, then
// their argument writes it undefined ("cleared") and the value read runs; their second calls find
// undefined, so neither their argument nor a callback runs.
const optionalValueCallOutput = `argument a
callback a
ran 1
ran 1
argument c
callback c
relay c!
relay undefined
ran 2
argument e
held e
cleared
ran 3
cleared
local 1
cleared
parameter 1
`

// optionalValueCallProgram is testdata/optional_value_call.a lowered, with each call through a
// function value whose callee is a ?? panic(...) made the optional call of that value. It fails unless
// it made exactly three, run's and hold's void ones and relay's string one, each an ordinary closure
// call, hold's on a read of its captured cb, so a lowering that stops making them can't leave the
// tests passing on calls that aren't optional.
func optionalValueCallProgram(t *testing.T) *ir.Program {
	t.Helper()
	checked, err := load.Load([]string{filepath.Join("testdata", "optional_value_call.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	made := []ir.CallClosure{}
	// A schema may point back into itself; each pointer is walked once.
	seen := map[uintptr]bool{}
	var visit func(value reflect.Value)
	visit = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			// An interface's value can't be changed where it is: it's changed in a copy, put back.
			held := reflect.New(value.Elem().Type()).Elem()
			held.Set(value.Elem())
			visit(held)
			if call, isCall := held.Interface().(ir.CallClosure); isCall {
				if coalesce, isCoalesce := call.Closure.(ir.Coalesce); isCoalesce && coalesce.Panic != nil && coalesce.Fallback == nil {
					call.Closure = coalesce.Value
					call.Optional = true
					held.Set(reflect.ValueOf(call))
					made = append(made, call)
				}
			}
			value.Set(held)
		case reflect.Pointer:
			if !value.IsNil() && !seen[value.Pointer()] {
				seen[value.Pointer()] = true
				visit(value.Elem())
			}
		case reflect.Struct:
			for index := range value.NumField() {
				if value.Field(index).CanSet() {
					visit(value.Field(index))
				}
			}
		case reflect.Slice, reflect.Array:
			for index := range value.Len() {
				visit(value.Index(index))
			}
		}
	}
	for index := range program.Functions {
		visit(reflect.ValueOf(&program.Functions[index].Body).Elem())
	}
	if len(made) != 7 {
		t.Fatalf("made %d optional calls, want 7 (run's, relay's, hold's and local's and parameter's two each): lowering no longer lowers (callback ?? panic(...))(...) as a call through the Coalesce", len(made))
	}
	returns := map[ir.Type]bool{}
	captured, assigned, parameter := false, false, false
	parameters := map[int]bool{}
	for _, function := range program.Functions {
		for _, local := range function.Parameters {
			parameters[local] = true
		}
	}
	for _, call := range made {
		if call.Bound || call.Direct > 0 {
			t.Fatalf("an optional call made of a bound or direct call (Bound %v, Direct %d): the emitters refuse it", call.Bound, call.Direct)
		}
		if _, _, isMethod := ir.MethodCallee(call.Closure); isMethod {
			t.Fatal("an optional call made on a method callee, not a function value")
		}
		if call.Closure.Type() != ir.Closure {
			t.Fatalf("an optional call's callee is %v, not a function value", call.Closure.Type())
		}
		returns[call.Returns] = true
		if read, isRead := call.Closure.(ir.Read); isRead {
			held := program.Locals[read.Local]
			captured = captured || held.Captured
			assigned = assigned || (!held.Captured && held.ExpressionAssigned && !parameters[read.Local])
			parameter = parameter || (!held.Captured && parameters[read.Local])
		}
	}
	if !returns[0] || !returns[ir.String] {
		t.Fatalf("optional calls return %v, want void and string", returns)
	}
	if !captured {
		t.Fatal("no optional call's callee is a read of a captured variable: hold's cb no longer tests a cell an argument clears")
	}
	// Compiler's pins on #036ve6w: the borrowed copy is safe on these two roads only because the
	// argument's write marks the local ExpressionAssigned, so its read is counted.
	if !assigned {
		t.Fatal("no optional call's callee is an uncaptured, expression-assigned local: local's cb no longer tests the road its own argument writes")
	}
	if !parameter {
		t.Fatal("no optional call's callee is an uncaptured parameter: parameter's cb no longer tests the road its own argument writes")
	}
	return program
}

// Mutant: evaluate the arguments before the presence test (natively, calledThrough's argument text
// written above optionalCallThrough's if; in JavaScript, adamicOptionalValueCall calling values()
// before its test). The calls with undefined then print "argument b" and "argument d" and count them,
// so the fourth line reads "argument b" where "ran 1" goes, every line after it moves down, "ran 2"
// reads "ran 4" and the last line "ran 5" where "ran 3" goes: both tests go red on the whole-output
// comparison.
//
// A second, native only: the held callee taken without a count (read lends it, as borrow.go lends a
// consumer's operand). hold's clear('e') then frees make's closure before the call, which runs freed
// memory: sanitize=true goes red on the address sanitizer's heap-use-after-free.
func TestOptionalValueCallNative(t *testing.T) {
	t.Parallel()
	code := C(optionalValueCallProgram(t))
	for _, sanitize := range []bool{false, true} {
		t.Run("sanitize="+strconv.FormatBool(sanitize), func(t *testing.T) {
			t.Parallel()
			binary := filepath.Join(t.TempDir(), "optional_value_call")
			if err := Build(code, binary, Options{Sanitize: sanitize}); err != nil {
				t.Fatal(err)
			}
			if got := runWithInput(t, "", binary); got != optionalValueCallOutput {
				t.Fatalf("native printed %q, want %q", got, optionalValueCallOutput)
			}
		})
	}
}

// The same program on the JavaScript backend, run as the oracle runs a generated module, its 'adamic'
// the runtime beside oracle/node.mjs. The mutant above goes red here the same way.
func TestOptionalValueCallJavaScript(t *testing.T) {
	t.Parallel()
	runner, err := filepath.Abs(filepath.Join("..", "..", "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	module := filepath.Join(t.TempDir(), "optional_value_call.mjs")
	if err := os.WriteFile(module, []byte(javascript.JavaScript(optionalValueCallProgram(t))), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", runner, module); got != optionalValueCallOutput {
		t.Fatalf("JavaScript printed %q, want %q", got, optionalValueCallOutput)
	}
}
