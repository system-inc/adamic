package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// typeConstructorProgram is tsc's Type made through objectAllocator, small (checker.ts's `new Type(checker,
// flags)` through utilities.ts's getTypeConstructor): constructor, an allocator whose getter hands it out `as
// any`, extra declarations, a module cache filled by the || at the new, and createType, then tail. Tp declares
// its fields in another order than any constructor here writes them, so Object.keys reads the writes' order or
// fails.
func typeConstructorProgram(constructor, extra, tail string) string {
	return "interface Tp { id: number; checker?: number; flags: number }\n" +
		"interface TypeAllocator { getTypeConstructor(): new (flags: number, checker: number) => Tp }\n" +
		constructor +
		"const objectAllocator: TypeAllocator = {\n    getTypeConstructor: () => TypeObject as any,\n};\n" +
		extra +
		"let TypeConstructor: (new (flags: number, checker: number) => Tp) | undefined;\n" +
		"function createType(flags: number): Tp {\n" +
		"    return new (TypeConstructor || (TypeConstructor = objectAllocator.getTypeConstructor()))(flags, 7);\n" +
		"}\n" +
		tail
}

// An optional field a function constructor writes on every path is laid present, as slice 1 lays every field,
// in first-write order, and agrees with Node on the JavaScript backend and natively: Object.keys reads the
// writes' order, never Tp's declaration order, JSON.stringify writes the same, and in sees the field (#w8nanzw
// (B)). The Absent half, a field one path skips, waits on #837zjkt's slice 2, whose eligibility admits the if;
// no program lowers to it until then (TestWindowLayout pins the layout itself).
func TestFunctionConstructorOptionalFieldAgrees(t *testing.T) {
	t.Parallel()
	constructor := "function TypeObject(this: Tp, flags: number, checker: number): void {\n" +
		"    this.flags = flags;\n" +
		"    this.checker = checker;\n" +
		"    this.id = 0;\n" +
		"}\n"
	tail := "const made = createType(3);\n" +
		"console.log(Object.keys(made).join(','));\n" +
		"console.log(String(JSON.stringify(made)));\n" +
		"console.log(String('checker' in made) + ' ' + String('id' in made));\n" +
		"console.log(String(made.checker));\n"
	const want = "flags,checker,id\n{\"flags\":3,\"checker\":7,\"id\":0}\ntrue true\n7\n"
	program := lowersAndAgreesWithNodeAs(t, typeConstructorProgram(constructor, "", tail), ".ts")
	code, out, errs := brandedNative(t, program)
	if code != 0 || out != want {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and Node's %q", code, out, errs, want)
	}
}

// Adamic's three attacks on the factory slot (#w8nanzw (B)), each a stop or agreeing with Node on both backends:
// a constructor handing its object to a callback before it writes checker, one writing checker in a loop that
// runs zero times, and two factories sharing one constructor where only one writes checker, after the new has
// returned its object. Slice 1's eligibility stops the first two at the getter's cast before a window is asked;
// the third's write is after the window, so no slot is laid for it.
func TestFunctionConstructorWindowAttacks(t *testing.T) {
	t.Parallel()
	report := "const made = createType(0);\n" +
		"console.log(Object.keys(made).join(','));\n" +
		"console.log(String(JSON.stringify(made)) + ' ' + String('checker' in made));\n"
	for _, attack := range []struct {
		name, source, want string
	}{
		{"a callback handed the object before the write", typeConstructorProgram("function TypeObject(this: Tp, flags: number, checker: number): void {\n    this.flags = flags;\n    this.id = 0;\n    visit(this);\n    this.checker = checker;\n}\n", "function visit(seen: Tp): void {\n    console.log(String('checker' in seen));\n}\n", report), "false\nflags,id,checker\n{\"flags\":0,\"id\":0,\"checker\":7} true\n"},
		{"a write in a loop that runs zero times", typeConstructorProgram("function TypeObject(this: Tp, flags: number, checker: number): void {\n    this.flags = flags;\n    this.id = 0;\n    for (let index = 0; index < flags; index++) {\n        this.checker = checker;\n    }\n}\n", "", report), "flags,id\n{\"flags\":0,\"id\":0} false\n"},
		{"two factories, one writing after the new", typeConstructorProgram("function TypeObject(this: Tp, flags: number, checker: number): void {\n    this.flags = flags;\n    this.id = checker;\n}\n", "", "function createChecked(flags: number): Tp {\n    const node = createType(flags);\n    node.checker = 5;\n    return node;\n}\nconst plain = createType(1);\nconst checked = createChecked(2);\nconsole.log(Object.keys(plain).join(',') + ' ' + Object.keys(checked).join(','));\nconsole.log(String('checker' in plain) + ' ' + String(checked.checker));\n"), "flags,id flags,id,checker\nfalse 5\n"},
	} {
		t.Run(attack.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, attack.source, ".ts"); err != nil {
				var notYet *NotYet
				var refused *Refused
				if !errors.As(err, &notYet) && !errors.As(err, &refused) {
					t.Fatalf("want a NotYet or a Refused, or agreement with Node; got %v", err)
				}
				t.Logf("stops: %v", err)
				return
			}
			program := lowersAndAgreesWithNodeAs(t, attack.source, ".ts")
			code, out, errs := brandedNative(t, program)
			if code != 0 || out != attack.want {
				t.Fatalf("lowers, and natively: exit %d stdout %q stderr %q; want exit 0 and Node's %q", code, out, errs, attack.want)
			}
		})
	}
}

// windowLayoutWindow is the construction window of a factory that allocates o, runs between, and returns it.
// Its parameters are debugging, a boolean, and visit, a callback the window can't see into.
func windowLayoutWindow(between ...ir.Statement) fresh.Window {
	body := []ir.Statement{ir.Declare{Local: 0, Value: ir.ObjectLiteral{}}}
	body = append(body, between...)
	body = append(body, ir.Return{Value: ir.Read{Local: 0, Of: ir.Object}})
	program := &ir.Program{
		Locals:    []ir.Local{{Name: "o", Type: ir.Object, Function: 0}, {Name: "debugging", Type: ir.Boolean, Function: 0}, {Name: "visit", Type: ir.Closure, Function: 0}},
		Functions: []ir.Function{{Name: "factory", Parameters: []int{1, 2}, Returns: ir.Object, Body: body}},
	}
	return fresh.ConstructionWindow(program, 0, &program.Functions[0].Body[0])
}

// The layout a function constructor's window lays (windowLayout, #w8nanzw (B)), pinned on the window itself
// until #837zjkt's slice 2 lowers a constructor that writes on one path: every path writing every field in the
// layout's order keeps the fields present as built; every path writing them in another order, or one path
// skipping checker, reserves every field Absent, so each is an own key at its first write, in the order the
// writes run; a write after the object is handed to a call is no write in the window, and names the field.
// Mutant (b): the window's fields laid present instead of Absent turns the other-order and one-path rows red.
func TestWindowLayout(t *testing.T) {
	t.Parallel()
	write := func(name string) ir.Statement {
		return ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: name, Value: ir.NumberConstant{Value: 1}, Site: 1}
	}
	debugging := ir.Read{Local: 1, Of: ir.Boolean}
	for _, row := range []struct {
		name      string
		between   []ir.Statement
		absent    bool
		unwritten string
	}{
		{"every path, in the layout's order", []ir.Statement{write("flags"), write("checker")}, false, ""},
		{"every path, in another order", []ir.Statement{write("checker"), write("flags")}, true, ""},
		{"one path", []ir.Statement{write("flags"), ir.If{Condition: debugging, Then: []ir.Statement{write("checker")}}}, true, ""},
		{"after a callback is handed the object", []ir.Statement{write("flags"), ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 2, Of: ir.Closure}, Arguments: []ir.Expression{ir.Read{Local: 0, Of: ir.Object}}}}, write("checker")}, false, "checker"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			fields := []ir.Field{{Name: "flags", Value: ir.NumberConstant{}}, {Name: "checker", Value: ir.NumberConstant{}}}
			laid, unwritten := windowLayout(fields, windowLayoutWindow(row.between...))
			if unwritten != row.unwritten {
				t.Fatalf("unwritten %q, want %q", unwritten, row.unwritten)
			}
			if row.unwritten != "" {
				return
			}
			if len(laid) != 2 || laid[0].Name != "flags" || laid[1].Name != "checker" {
				t.Fatalf("laid %+v, want flags then checker, the layout's order", laid)
			}
			for _, field := range laid {
				if field.Absent != row.absent || field.Uninitialized != row.absent {
					t.Errorf("%s laid Absent %v Uninitialized %v, want both %v", field.Name, field.Absent, field.Uninitialized, row.absent)
				}
			}
			for _, field := range fields {
				if field.Absent || field.Uninitialized {
					t.Errorf("%s was changed where it was built; the layout is a new slice", field.Name)
				}
			}
		})
	}
}

// The pin Lower's slice 2 leans on (#w8nanzw (B)): layFunctionConstructorWindows lays presence after every pass
// that reads a literal's fields but none that reads their presence. Slice 1 can't write a constructor whose
// window is settled in another order than its layout, so this lowers the honest one whole, swaps its first two
// writes in F_new's body, as a constructor writing checker before flags would be, and lays the window again:
// every field is reserved Absent and Uninitialized in the literal and its class, the native emitter takes the
// construction path for it (adamic_object_construct), and natively Object.keys and JSON.stringify read the
// writes' order, checker first. The emitted program is the whole lowering's, with only the pass run again.
func TestFunctionConstructorReorderedWindowConstructs(t *testing.T) {
	t.Parallel()
	constructor := "function TypeObject(this: Tp, flags: number, checker: number): void {\n" +
		"    this.flags = flags;\n" +
		"    this.checker = checker;\n" +
		"    this.id = 0;\n" +
		"}\n"
	tail := "const made = createType(3);\n" +
		"console.log(Object.keys(made).join(','));\n" +
		"console.log(String(JSON.stringify(made)));\n" +
		"console.log(String('checker' in made) + ' ' + String('id' in made));\n" +
		"console.log(String(made.checker));\n"
	keyed, err := lowerKeyed(t, writeAgreementSource(t, "main.ts", typeConstructorProgram(constructor, "", tail)))
	if err != nil {
		t.Fatalf("the honest constructor must lower: %v", err)
	}
	built := keyed.lowering
	if built.functionConstructorScan == nil || len(built.functionConstructorScan.ordered) != 1 || built.functionConstructorScan.ordered[0].constructor < 0 {
		t.Fatal("want one built function constructor")
	}
	made := built.functionConstructorScan.ordered[0]
	allocated := func() ir.ObjectLiteral {
		t.Helper()
		for _, statement := range built.result.Functions[made.constructor].Body {
			if declared, isDeclare := statement.(ir.Declare); isDeclare {
				if literal, isLiteral := declared.Value.(ir.ObjectLiteral); isLiteral && literal.Class == made.class {
					return literal
				}
			}
		}
		t.Fatal("F_new allocates no literal of its class")
		return ir.ObjectLiteral{}
	}
	for _, field := range allocated().Fields {
		if field.Absent || field.Uninitialized {
			t.Fatalf("before the swap, %s is laid Absent %v Uninitialized %v; every path writes every field in order, so it stays present", field.Name, field.Absent, field.Uninitialized)
		}
	}
	const construct = "adamic_object_construct(&"
	before := strings.Count(native.C(built.result), construct)
	body := built.result.Functions[made.constructor].Body
	var writes []int
	for index, statement := range body {
		if store, isStore := statement.(ir.SetProperty); isStore && (store.Name == "flags" || store.Name == "checker") {
			writes = append(writes, index)
		}
	}
	if len(writes) != 2 {
		t.Fatalf("want F_new's body to hold the flags and checker writes as two SetProperty statements, found %d in %+v", len(writes), body)
	}
	body[writes[0]], body[writes[1]] = body[writes[1]], body[writes[0]]
	if err := built.layFunctionConstructorWindows(); err != nil {
		t.Fatalf("laying the reordered window: %v", err)
	}
	for _, field := range allocated().Fields {
		if !field.Absent || !field.Uninitialized {
			t.Errorf("the literal's %s is laid Absent %v Uninitialized %v; a window settled in another order reserves every field", field.Name, field.Absent, field.Uninitialized)
		}
	}
	for _, field := range built.result.Classes[made.class-1].Fields {
		if !field.Absent || !field.Uninitialized {
			t.Errorf("the class's %s is laid Absent %v Uninitialized %v, unlike its literal", field.Name, field.Absent, field.Uninitialized)
		}
	}
	if after := strings.Count(native.C(built.result), construct); after <= before {
		t.Errorf("the native C calls %s %d times after the swap and %d before; the reserved literal must take the construction path", construct, after, before)
	}
	const want = "checker,flags,id\n{\"checker\":7,\"flags\":3,\"id\":0}\ntrue true\n7\n"
	code, out, errs := brandedNative(t, built.result)
	if code != 0 || out != want {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and %q, the writes' order", code, out, errs, want)
	}
}
