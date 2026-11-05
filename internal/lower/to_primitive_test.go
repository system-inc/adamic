package lower

import (
	"context"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// toPrimitiveHelper is the program's to_string_hint, which every String of an unknown calls.
func toPrimitiveHelper(t *testing.T, program *ir.Program) ir.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == "to_string_hint" {
			return function
		}
	}
	t.Fatal("no to_string_hint in a program that converts an unknown to a string")
	return ir.Function{}
}

// The closed world (Adamic's ruling on #09dvwny): every non-static class the program makes has an arm,
// so a class's object reaching an unknown is never a gap, whichever classes reach one. A base, a
// subclass, a generic class's two instantiations, a class with a static side, one that never reaches
// an unknown, and an object literal with a getter, which is a class of its own. Mutant: a class left
// out of instanceToString's arms, which this names.
func TestToPrimitiveCoversEveryClass(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class Base {
	toString(): string {
		return 'base';
	}
}
class Derived extends Base {
	readonly size = 1;
}
class Box<T> {
	constructor(readonly value: T) {}
}
class Counter {
	static count = 0;
	increment(): number {
		return Counter.count + 1;
	}
}
class Unseen {
	readonly flag = true;
}
const accessor = { get area(): number { return 2; } };
function show(value: unknown): void {
	console.log(String(value));
}
show(new Base());
show(new Derived());
show(new Box<number>(1));
show(new Box<string>('a'));
console.log(String(new Counter().increment()));
console.log(String(new Unseen().flag));
console.log(String(accessor.area));
`)
	if err != nil {
		t.Fatal(err)
	}
	covered := map[int]bool{}
	walk(toPrimitiveHelper(t, program).Body, func(node any) bool {
		if test, ok := node.(ir.InstanceOf); ok && test.Exact {
			covered[test.Class] = true
		}
		return true
	})
	instances, statics := 0, 0
	for index, class := range program.Classes {
		if class.Static {
			statics++
			continue
		}
		instances++
		if !covered[index+1] {
			t.Errorf("class %d (%s) has no arm in to_string_hint", index+1, class.Name)
		}
	}
	// The program's own shape, so the test can't pass on a program with fewer classes than it names.
	if instances < 7 || statics == 0 {
		t.Fatalf("want at least seven classes an object can have and a static side, got %d and %d", instances, statics)
	}
}

// A toString or valueOf field a base class declares is an own field of every subclass's object, which
// Node reads before any method (Lower's should on #09dvwny), so a subclass's arm stops as its base's
// does, never calling an inherited method or answering "[object Object]". It holds because a subclass's
// Fields begin with its base's (class.go's instantiate). Every arm here stops, Valued's and Below's too,
// where Node answers "[object Object]" and "method": a field of either name stops the class, an
// over-stop that fails closed. Mutant: classToString reading only the class's
// own fields, from OwnStart, so Derived's arm answers "[object Object]" and Below's calls its method.
func TestToPrimitiveInheritedFieldStops(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class Base {
	toString = (): string => 'field';
}
class Derived extends Base {}
class Valued {
	valueOf = (): number => 2;
}
class Below extends Valued {
	toString(): string {
		return 'method';
	}
}
function show(value: unknown): void {
	console.log(String(value));
}
show(new Base());
show(new Derived());
show(new Below());
`)
	if err != nil {
		t.Fatal(err)
	}
	want := "converting an object whose class has a toString or valueOf field to a string is not yet lowered"
	// The class arms sit inside the instance kind's branch, so they're found at any depth.
	arms := map[string]string{}
	walk(toPrimitiveHelper(t, program).Body, func(node any) bool {
		branch, ok := node.(ir.If)
		if !ok {
			return true
		}
		test, ok := branch.Condition.(ir.InstanceOf)
		if !ok || !test.Exact {
			return true
		}
		name := program.Classes[test.Class-1].Name
		arms[name] = ""
		if len(branch.Then) == 1 {
			if panicked, ok := branch.Then[0].(ir.Panic); ok {
				if text, ok := panicked.Message.(ir.StringConstant); ok {
					arms[name] = program.Strings[text.Index]
				}
			}
		}
		return true
	})
	for _, name := range []string{"Base", "Derived", "Valued", "Below"} {
		got, found := arms[name]
		if !found {
			t.Errorf("no arm for %s", name)
			continue
		}
		if got != want {
			t.Errorf("%s's arm stops with %q, want %q", name, got, want)
		}
	}
}

// The arms no program reaches yet: unknownView refuses a function, a Map, a Set and a tuple where one
// flows into an unknown, so each is held by the helper's shape, a loud stop naming the kind as not yet,
// rather than by a program beside Node. The static class arm is reachable, and is held beside Node too
// (internal/oracle's TestToPrimitiveStopsAlikeOnBothBackends); the array arm joins
// (TestToPrimitiveArrayJoinLetsGoInFinally).
func TestToPrimitiveUnreachableArmsStop(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "function show(value: unknown): void {\n\tconsole.log(String(value));\n}\nshow(1);\n")
	if err != nil {
		t.Fatal(err)
	}
	arms := map[int][]ir.Statement{}
	body := toPrimitiveHelper(t, program).Body
	for _, statement := range body {
		branch, ok := statement.(ir.If)
		if !ok {
			continue
		}
		if test, ok := branch.Condition.(ir.Binary); ok && test.Operator == ir.Equal {
			if number, ok := test.Right.(ir.NumberConstant); ok {
				arms[int(number.Value)] = branch.Then
			}
		}
	}
	stop := func(statements []ir.Statement) string {
		if len(statements) != 1 {
			return ""
		}
		panicked, ok := statements[0].(ir.Panic)
		if !ok {
			return ""
		}
		text, ok := panicked.Message.(ir.StringConstant)
		if !ok {
			return ""
		}
		return program.Strings[text.Index]
	}
	for kind, what := range map[int]string{
		ir.UnknownFunction:    "a function",
		ir.UnknownCollection:  "a Map or a Set",
		ir.UnknownTypedArray:  "a typed array",
		ir.UnknownStaticClass: "a class object",
	} {
		if got, want := stop(arms[kind]), "converting "+what+" to a string is not yet lowered"; got != want {
			t.Errorf("kind %d's arm stops with %q, want %q", kind, got, want)
		}
	}
	if len(body) < 2 || !strings.Contains(stop(body[len(body)-2:len(body)-1]), "a value of this kind") {
		t.Errorf("want the helper to end in a stop for every other kind")
	}
}

// The array arm's join stack (Compiler's must on #09dvwny): an array of only scalars stops, named,
// before anything is pushed, JoinEnter is asked before the join, and the one JoinLeave is the finally of
// the try around the loop, so a throw from an element's toString lets the array go. internal/oracle's
// arrays.a holds the same beside Node; this names the shape. Mutants: JoinLeave moved after the try,
// and JoinEnter dropped.
func TestToPrimitiveArrayJoinLetsGoInFinally(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "function show(value: unknown): void {\n\tconsole.log(String(value));\n}\nshow(['a']);\n")
	if err != nil {
		t.Fatal(err)
	}
	var arm []ir.Statement
	for _, statement := range toPrimitiveHelper(t, program).Body {
		branch, ok := statement.(ir.If)
		if !ok {
			continue
		}
		if test, ok := branch.Condition.(ir.Binary); ok && test.Operator == ir.Equal {
			if number, ok := test.Right.(ir.NumberConstant); ok && int(number.Value) == ir.UnknownArray {
				arm = branch.Then
			}
		}
	}
	if len(arm) < 3 {
		t.Fatalf("want the array arm's scalar stop, its JoinEnter and its join, got %d statements", len(arm))
	}
	scalars, ok := arm[0].(ir.If)
	if _, tested := scalars.Condition.(ir.ArrayOfScalars); !ok || !tested || len(scalars.Then) != 1 {
		t.Fatalf("want the arm to stop first on an array of only scalars, got %#v", arm[0])
	}
	named := ""
	if panicked, ok := scalars.Then[0].(ir.Panic); ok {
		if text, ok := panicked.Message.(ir.StringConstant); ok {
			named = program.Strings[text.Index]
		}
	}
	if want := "converting an array of only numbers, booleans, null or undefined to a string is not yet lowered"; named != want {
		t.Errorf("want the scalar stop %q, got %q", want, named)
	}
	entered, leaves, finallyLeaves := false, 0, 0
	walk(arm, func(node any) bool {
		switch node := node.(type) {
		case ir.JoinEnter:
			entered = true
		case ir.JoinLeave:
			leaves++
		case ir.Try:
			if node.HasFinally && len(node.Finally) == 1 {
				if evaluated, ok := node.Finally[0].(ir.Evaluate); ok {
					if _, leaving := evaluated.Value.(ir.JoinLeave); leaving {
						finallyLeaves++
					}
				}
			}
		}
		return true
	})
	if !entered {
		t.Errorf("want the arm to ask JoinEnter before it joins")
	}
	if leaves != 1 || finallyLeaves != 1 {
		t.Errorf("want one JoinLeave, the finally of the join's try, got %d, %d of them a finally", leaves, finallyLeaves)
	}
}

// Unknown[] now has union storage. The mixed scalar/nullish literal agrees with
// Node. The reference cycle still stops: its unknown owner has no ownership
// schema, and admitting the reference push would leak the cycle.
func TestToPrimitiveProbesOfUnknownArrays(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, refusal string }{
		{"function show(v: unknown): string { return String(v) + \"|\" + `${v}`; }\nconst a: unknown[] = [\"a\"];\nconst b: unknown[] = [\"b\", a];\na.push(b);\nconsole.log(show(a) + \" \" + show(b));\n", "a reference write into an unknown array whose ownership is not proven"},
		{"function show(v: unknown): string { return String(v) + \"|\" + `${v}`; }\nconst arr: unknown[] = [\"a\", null, undefined, \"b\"];\nconsole.log(show(arr));\n", ""},
	} {
		if probe.refusal == "" {
			lowersAndAgreesWithNode(t, probe.source)
			continue
		}
		_, err := lowerSource(t, probe.source)
		if err == nil || !strings.Contains(err.Error(), probe.refusal) {
			t.Errorf("want %q, got %v: %q", probe.refusal, err, probe.source)
		}
	}
}

// String(symbol) is "Symbol(d)" and a template of one throws, so the two forms part at a symbol, and a
// bigint would be "1" in both. Neither has a representation, so neither reaches an unknown, which is
// why to_string_hint serves String and a template alike. The day one does, this goes red, and the
// helper needs the two forms apart.
func TestToPrimitiveSymbolAndBigIntReachNoUnknown(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const value: unknown = Symbol('d');\nconsole.log(String(value));\n",
		"const value: unknown = Symbol('d');\nconsole.log(`${value}`);\n",
		"const value: unknown = 1n;\nconsole.log(String(value));\n",
	} {
		checked, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
		if err == nil {
			_, err = Lower(context.Background(), checked)
		}
		if err == nil {
			t.Errorf("a symbol or a bigint reached an unknown, where String and a template part: %q", source)
		}
	}
}
