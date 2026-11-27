package lower

import (
	"maps"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// spreadAccessorPrelude declares a view whose weight a program class may hold on its prototype.
const spreadAccessorPrelude = "interface Source { readonly label: string; readonly weight: number; }\n"

// spreadAccessorNames is every name a spread in the program may pass over as a program class's accessor, a
// setter-only one included, and never one a class or an object literal holds as a method, or an object
// literal holds as an accessor, since the run-time key read stops on those (#vjvcnnj head 3). A spread of a
// class instance's own type passes over nothing its copy's type claims, since the checker types that copy
// without the class's accessors (probe J-qy3z2fy), and a spread of a type parameter passes over every class
// its constraint admits. Mutant (spread_accessor_reads.mutants.md): setter-only names left out reds the
// setter row here and read_setter_typeof.ts in TestSpreadPrototypeField.
func TestSpreadAccessorNames(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		want         []string
	}{
		{"a getter through a view", "class Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nconst built: Source = new Sized();\nconst copy: Source = { ...built };\n", []string{"weight"}},
		{"a setter-only name through a view", "class Sized { readonly label = \"t\"; set weight(value: number) {} }\nconst built: Source = new Sized();\nconst copy: Source = { ...built };\n", []string{"weight"}},
		{"a class instance's own type", "class Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nconst copy = { ...new Sized() };\n", nil},
		{"no spread", "class Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nconst built: Source = new Sized();\n", nil},
		{"a name another class holds as a method", "class Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nclass Scale { readonly label = \"s\"; weight(): number { return 1; } }\nconst built: Source = new Sized();\nconst copy: Source = { ...built };\n", nil},
		{"a name an object literal holds as an accessor", "class Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nconst literal: Source = { label: \"l\", get weight() { return 3; } };\nconst built: Source = new Sized();\nconst copy: Source = { ...built };\n", nil},
		{"a type parameter's constraint", "class Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nfunction copy<T extends Source>(source: T): T { return { ...source }; }\n", []string{"weight"}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			withAccessChecker(t, spreadAccessorPrelude+probe.source, func(l *lowering, _ []*ast.Node) {
				if got := slices.Sorted(maps.Keys(l.spreadAccessorNames())); !slices.Equal(got, probe.want) {
					t.Fatalf("spreadAccessorNames = %v, want %v", got, probe.want)
				}
			})
		})
	}
}

// A read by name answers whatever the receiver holds, so the read's union is narrowed to the field's type or
// undefined only after a test of its tag, and a slot of another kind stops loudly instead of being read as a
// number (Compiler's hold 1 on #vjvcnnj head 3). No program reaches that today: an `as unknown as` pair, a
// function's bivariant parameter, an array's covariance and a predicate's overload each stop at compile time
// on main (probes J-qy3z2fy, J-dx6zd0c, J-vqxgr9r), so the check is a backstop, and behavior can't observe
// it. This holds its shape: the read is Reflect's run-time key read of weight, and the helper it goes
// through panics with spreadAccessorReadMismatch unless typeof says number or the value is undefined, before
// its checked narrow to number | undefined. Mutant (spread_accessor_reads.mutants.md): the check dropped reds
// this test.
func TestSpreadAccessorReadChecksTheKind(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, spreadAccessorPrelude+"class Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nconst built: Source = new Sized();\nconst copy: Source = { ...built };\nconsole.log(typeof copy.weight);\n", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	helper := slices.IndexFunc(program.Functions, func(function ir.Function) bool { return function.Name == "spread_accessor_read" })
	if helper < 0 {
		t.Fatal("no spread_accessor_read helper: the read wasn't read by name")
	}
	function := program.Functions[helper]
	if !function.CheckedUnionNarrow || function.Returns != ir.MaybeNumber || len(function.Body) != 2 {
		t.Fatalf("the helper is %+v, want a checked narrow to number | undefined after one test", function)
	}
	test, tested := function.Body[0].(ir.If)
	if !tested || len(test.Then) != 1 {
		t.Fatalf("the helper's first statement is %#v, want the tag's test", function.Body[0])
	}
	stop, stops := test.Then[0].(ir.Panic)
	message, constant := stop.Message.(ir.StringConstant)
	if !stops || !constant || program.Strings[message.Index] != spreadAccessorReadMismatch {
		t.Fatalf("the test's arm is %#v, want a panic with %q", test.Then[0], spreadAccessorReadMismatch)
	}
	negated, not := test.Condition.(ir.Unary)
	either, or := negated.Operand.(ir.Binary)
	kind, compared := either.Left.(ir.Binary)
	named, constant := kind.Right.(ir.StringConstant)
	_, typed := kind.Left.(ir.TypeOf)
	_, undefined := either.Right.(ir.IsUndefined)
	if !not || negated.Operator != ir.Not || !or || either.Operator != ir.Or || !compared || !typed || !constant || program.Strings[named.Index] != "number" || !undefined {
		t.Fatalf("the test is %#v, want typeof against number or undefined", test.Condition)
	}
	found := false
	walk(program.Main, func(node any) bool {
		if call, ok := node.(ir.Call); ok && call.Function == helper {
			read, keyed := call.Arguments[0].(ir.DynamicKeyedProperty)
			key, named := read.Key.(ir.StringConstant)
			found = keyed && read.Reflect && named && program.Strings[key.Index] == "weight"
		}
		return true
	})
	if !found {
		t.Fatal("the helper isn't called with Reflect's run-time key read of weight")
	}
}
