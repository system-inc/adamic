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
const accessor = { get size(): number { return 2; } };
function show(value: unknown): void {
	console.log(String(value));
}
show(new Base());
show(new Derived());
show(new Box<number>(1));
show(new Box<string>('a'));
console.log(String(new Counter().increment()));
console.log(String(new Unseen().flag));
console.log(String(accessor.size));
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

// The arms no program reaches yet: unknownView refuses a function, a Map, a Set and a tuple where one
// flows into an unknown, so each is held by the helper's shape, a loud stop naming the kind as not yet,
// rather than by a program beside Node. The static class and array arms are reachable, and are held
// beside Node too (internal/oracle's TestToPrimitiveStopsAlikeOnBothBackends).
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
		ir.UnknownArray:       "an array or a tuple",
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
