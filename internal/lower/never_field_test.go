package lower

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Ruling 2 (Adamic on #bbssdcy, Nov 24): a field every declaring member declares never takes the plain
// member's storage, and every read of it is a :76 read (rulings.md:76). These pin it at the lowering;
// internal/oracle/never_field_test.go holds the programs that run to Node.

// neverFieldArrow is tsc's ArrowFunction in small: a Base whose name is optional, and an Arrow that
// declares it never, made by a cast from a Base, as tsc's factory makes its nodes.
const neverFieldArrow = `interface Base {
	readonly name?: string;
	readonly kind: number;
}
interface Arrow extends Base {
	readonly name: never;
}
function makeArrow(made: Base): Arrow {
	return made as Arrow;
}
`

// neverFieldTrapMessages is the message of every never_read trap in a lowered program.
func neverFieldTrapMessages(program *ir.Program) []string {
	var messages []string
	for _, function := range program.Functions {
		if function.Name != "never_read" || len(function.Body) != 1 {
			continue
		}
		if panics, ok := function.Body[0].(ir.Panic); ok {
			if message, ok := panics.Message.(ir.StringConstant); ok {
				messages = append(messages, program.Strings[message.Index])
			}
		}
	}
	return messages
}

// neverFieldTrapped is whether some never_read trap's message is the one a read or a write of the field at
// where makes: "unreachable <what> of <source> at <file>:<line>:<column>, a field declared never".
func neverFieldTrapped(program *ir.Program, what, source, where string) bool {
	for _, message := range neverFieldTrapMessages(program) {
		if strings.HasPrefix(message, "unreachable "+what+" of "+source+" at ") && strings.HasSuffix(message, where+", a field declared never") {
			return true
		}
	}
	return false
}

// A read of a field every declaration types never lowers to the trap, whatever holds the field: a plain
// Arrow, and the intersection tsc's sites hold (Expression & ArrowFunction), which before ruling 2 stopped
// at the intersection itself. A program that writes a getter of the name keeps the read stopped as not
// yet, since the trap would leave the getter's code out.
func TestNeverFieldReadIsATrap(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, read string }{
		{"plain", neverFieldArrow + "const arrow = makeArrow({ kind: 1 });\nconsole.log(String(arrow.name));\n", "arrow.name"},
		{"intersection", "interface Expression {\n\treadonly kind: number;\n}\n" + neverFieldArrow + "function nameOf(node: Expression & Arrow): string {\n\treturn String(node.name);\n}\nconsole.log(nameOf(makeArrow({ kind: 1 })));\n", "node.name"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if where := neverReadWhere(t, probe.source, probe.read); !neverFieldTrapped(program, "read", probe.read, where) {
				t.Fatalf("want a trap naming the read of %s at %s as a field declared never, got traps %q", probe.read, where, neverFieldTrapMessages(program))
			}
		})
	}
	t.Run("getter", func(t *testing.T) {
		t.Parallel()
		source := neverFieldArrow + "class Named {\n\tget name(): string { return \"named\"; }\n}\nconsole.log(new Named().name);\nconst arrow = makeArrow({ kind: 1 });\nconsole.log(String(arrow.name));\n"
		_, err := lowerSourceAs(t, source, ".ts")
		var notYet *NotYet
		if want := neverReadWhere(t, source, "arrow.name"); !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, want) || notYet.What != "a read of field name, declared never, that a getter may serve" {
			t.Fatalf("got %v, want a not yet at %s for a read a getter may serve", err, want)
		}
	})
}

// A field reduced to never by conflict keeps today's stop with today's words (#mhggcdw's case): its
// members declare string and number, a never and a string, or, in the chained row, a member inherits the
// conflicting pair from its own base, so not every member that declares name declares never. Mutant:
// tmp-audit/wrong-conflict-held.patch holds every never field, and each row lowers.
func TestNeverFieldConflictKeepsItsStop(t *testing.T) {
	t.Parallel()
	const words = "an intersection whose field name has no representation: never (no value has it)"
	for _, probe := range []struct{ name, source string }{
		{"conflict", "function show(named: { readonly name: string } & { readonly name: number }): string {\n\treturn \"shown\";\n}\nconsole.log(\"conflict\");\n"},
		{"mixed", neverFieldArrow + "function show(named: Arrow & { readonly name: string }): string {\n\treturn \"shown\";\n}\nconsole.log(\"mixed\");\n"},
		{"chained", "type Pair = { readonly name: string } & { readonly name: number };\ninterface Chained extends Pair {\n\treadonly kind: number;\n}\nfunction show(named: Chained & { readonly flags: number }): string {\n\treturn \"shown\";\n}\nconsole.log(\"chained\");\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.What, words) {
				t.Fatalf("got %v, want a not yet ending %q", err, words)
			}
		})
	}
}

// A write into a field every declaration types never can only write a value of type never, here a
// parameter two checks exhaust, so it is a :76 read at the write: the receiver and the value are evaluated,
// the trap names the write, and nothing is stored. The value is a trusted never read, trapped as well, so
// no run reaches the write's own trap with a value, and the pin is the lowering's. Mutant:
// tmp-audit/wrong-write-unchecked.patch stores the value instead, and the store is found.
func TestNeverFieldWriteIsATrapAtTheWrite(t *testing.T) {
	t.Parallel()
	source := "type Shape = \"circle\" | \"square\";\ninterface Slot {\n\treadonly kind: number;\n\tname: never;\n}\nfunction put(slot: Slot, shape: Shape): string {\n\tif (shape === \"circle\") return \"circle\";\n\tif (shape === \"square\") return \"square\";\n\tslot.name = shape;\n\treturn \"written\";\n}\nconsole.log(\"put\");\n"
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("want it to lower, got %v", err)
	}
	if where := neverReadWhere(t, source, "slot.name = shape"); !neverFieldTrapped(program, "write", "slot.name", where) {
		t.Fatalf("want a trap naming the write of slot.name at %s, got traps %q", where, neverFieldTrapMessages(program))
	}
	stored := false
	scan := func(node any) bool {
		if store, ok := node.(ir.SetProperty); ok && store.Name == "name" {
			stored = true
		}
		return true
	}
	walk(program.Main, scan)
	for _, function := range program.Functions {
		walk(function.Body, scan)
	}
	if stored {
		t.Fatal("the write into a field declared never stores its value")
	}
}

// neverFieldStores counts what in a lowered program can store field name into an object that already
// exists: an ir.SetProperty of that name, an ir.LogicalAssignment whose computed key names it (its keyed
// SetProperty has no name of its own), and every Object.assign (an ir.ObjectCall "assign"), which copies
// whatever its sources hold. They are the object writes ir.Program.FrozenWriteMayThrow counts, which
// native lowers to adamic_object_write_field and adamic_object_assign; an ir.ObjectLiteral makes an
// object rather than writing one, and an ir.SetIndex writes an array.
func neverFieldStores(program *ir.Program, name string) int {
	stores := 0
	scan := func(node any) bool {
		switch node := node.(type) {
		case ir.SetProperty:
			if node.Name == name {
				stores++
			}
		case ir.LogicalAssignment:
			if slices.Contains(node.KeyNames, name) {
				stores++
			}
		case ir.ObjectCall:
			if node.Method == "assign" {
				stores++
			}
		}
		return true
	}
	walk(program.Main, scan)
	for _, function := range program.Functions {
		walk(function.Body, scan)
	}
	return stores
}

// neverFieldRoad is a program whose put writes Slot's name, declared never and writable, by write, after
// two checks leave shape never, as a field a cast reaches: tsc accepts each write.
func neverFieldRoad(write string) string {
	return "type Shape = \"circle\" | \"square\";\ninterface Slot {\n\treadonly kind: number;\n\tname: never;\n}\nfunction put(o: Slot, shape: Shape): string {\n\tif (shape === \"circle\") return \"circle\";\n\tif (shape === \"square\") return \"square\";\n\t" + write + "\n\treturn \"written\";\n}\nconsole.log(\"put\");\n"
}

// Lower's question on ruling 2 (#bbssdcy): the intersection gives a never field no slot, so every road
// that writes a declared-never field reached through a cast traps or stops at compile time, and none
// stores. A plain = is the :76 read at the write (declaredNeverWrite). The logical assignments, ++ and --
// each ask the field's representation first (logicalAssignment's typeOf of the write, updateProperty's of
// the field), never (no value has it). A destructuring target that is a field stops for any field
// (destructuringAssignment, objectDestructuringAssignment), and here the array's tuple element of type
// never stops sooner, at the value. A string-literal key goes to finiteElementWrite, whose finiteKeyFields
// finds no representation for the field. Object.assign stops at a target no literal proves, which a field
// declared never always is, since only a cast or a never value makes one. o.name += x isn't a row: tsc
// refuses it, since never + never is a number, and every arithmetic compound makes a number or a string,
// none of which never holds. Each control writes an ordinary field the same way and lowers, and the same
// count finds its store, so the count can see one; destructuring has none, since it stores no field at all.
func TestNeverFieldEveryWriteRoadTrapsOrStops(t *testing.T) {
	t.Parallel()
	const objectSource = "type Source = { readonly k: \"a\"; readonly v: string } | { readonly k: \"b\"; readonly v: number };\ninterface Slot {\n\treadonly kind: number;\n\tname: never;\n}\nfunction put(o: Slot, src: Source): string {\n\tif (src.k === \"a\") return \"a\";\n\tif (src.k === \"b\") return \"b\";\n\t({ v: o.name } = src);\n\treturn \"written\";\n}\nconsole.log(\"put\");\n"
	const neverWhat = "a value of type never, never (no value has it)"
	for _, road := range []struct{ name, source, at, what string }{
		{"plain", neverFieldRoad("o.name = shape;"), "o.name = shape", ""},
		{"or_assign", neverFieldRoad("o.name ||= shape;"), "o.name ||= shape", neverWhat},
		{"and_assign", neverFieldRoad("o.name &&= shape;"), "o.name &&= shape", neverWhat},
		{"nullish_assign", neverFieldRoad("o.name ??= shape;"), "o.name ??= shape", neverWhat},
		{"increment", neverFieldRoad("o.name++;"), "o.name++", neverWhat},
		{"decrement", neverFieldRoad("o.name--;"), "o.name--", neverWhat},
		{"object_destructuring", objectSource, "o.name } = src", "assigning to a PropertyAccessExpression in a destructuring assignment"},
		{"array_destructuring", neverFieldRoad("[o.name] = [shape];"), "shape];", "a tuple element of type never"},
		{"element", neverFieldRoad("o[\"name\"] = shape;"), "o[\"name\"] = shape", "a computed field with an unsupported representation"},
		{"object_assign", neverFieldRoad("Object.assign(o, { name: shape });"), "o, { name: shape }", "Object.assign on a shape not proven by a plain literal or its const binding"},
	} {
		t.Run(road.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, road.source, ".ts")
			where := neverReadWhere(t, road.source, road.at)
			if road.what != "" {
				var notYet *NotYet
				var refused *Refused
				switch {
				case errors.As(err, &notYet) && notYet.What == road.what && strings.HasSuffix(notYet.Where, where):
				case errors.As(err, &refused) && refused.What == road.what && strings.HasSuffix(refused.Where, where):
				default:
					t.Fatalf("got %v, want a stop at %s saying %q", err, where, road.what)
				}
				return
			}
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if !neverFieldTrapped(program, "write", "o.name", where) {
				t.Fatalf("want a trap naming the write of o.name at %s, got traps %q", where, neverFieldTrapMessages(program))
			}
			if stores := neverFieldStores(program, "name"); stores != 0 {
				t.Fatalf("the write into a field declared never stores %d times", stores)
			}
		})
	}
	const box = "interface Box {\n\tkind: number;\n\tcount: number;\n\tlabel: string;\n}\nfunction touch(o: Box, text: string): void {\n\t"
	for _, control := range []struct{ name, write, field string }{
		{"control_assign", "o.label = text;", "label"},
		{"control_or_assign", "o.label ||= text;", "label"},
		{"control_increment", "o.count++;", "count"},
		{"control_element", "o[\"label\"] = text;", "label"},
		{"control_object_assign", "const held = { kind: 1, label: \"a\" };\n\tObject.assign(held, { label: text });", "label"},
	} {
		t.Run(control.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, box+control.write+"\n}\nconsole.log(\"touch\");\n", ".ts")
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if neverFieldStores(program, control.field) == 0 {
				t.Fatalf("the count finds no store of %s, so it couldn't see one of a never field", control.field)
			}
		})
	}
}

// Language's condition on ruling 2 (#bbssdcy, 16:27Z): a write of name through a base view that declares
// it writable is refused by invariance, as at the pin (invariance.go's wideningRefusal,
// adamic/invariant-mutable). The Arrow is handed to a function taking the Base, which writes a string where
// the Arrow's never is read.
func TestNeverFieldBaseViewWriteIsRefused(t *testing.T) {
	t.Parallel()
	source := "interface Base {\n\tname?: string;\n\tkind: number;\n}\ninterface Arrow extends Base {\n\tname: never;\n}\nfunction rename(base: Base): void {\n\tbase.name = \"written\";\n}\nfunction touch(arrow: Arrow): void {\n\trename(arrow);\n}\nconsole.log(\"touched\");\n"
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if want := neverReadWhere(t, source, "arrow);"); !errors.As(err, &refused) || !strings.HasSuffix(refused.Where, want) || !strings.Contains(err.Error(), "a value of type Arrow seen as Base, which can write ") || !strings.Contains(err.Error(), " where never is read") || !strings.Contains(err.Error(), "(adamic/invariant-mutable)") {
		t.Fatalf("got %v, want the invariance refusal at %s", err, want)
	}
}

// neverFieldLoads counts the loads of field name, an ir.Property, whose receiver reads the local named
// receiver, anywhere in a lowered program.
func neverFieldLoads(program *ir.Program, receiver, name string) int {
	loads := 0
	scan := func(node any) bool {
		if property, ok := node.(ir.Property); ok && property.Name == name && neverFieldReadsLocal(program, property.Object, receiver) {
			loads++
		}
		return true
	}
	walk(program.Main, scan)
	for _, function := range program.Functions {
		walk(function.Body, scan)
	}
	return loads
}

// neverFieldReadsLocal is whether an expression reads the local named name anywhere inside it.
func neverFieldReadsLocal(program *ir.Program, expression ir.Expression, name string) bool {
	found := false
	walk(expression, func(node any) bool {
		if read, ok := node.(ir.Read); ok && program.Locals[read.Local].Name == name {
			found = true
		}
		return !found
	})
	return found
}

// neverFieldTrapsEvaluated is what each never_read trap's comma evaluates before the trap's call, anywhere
// in a lowered program.
func neverFieldTrapsEvaluated(program *ir.Program) []ir.Expression {
	var evaluated []ir.Expression
	scan := func(node any) bool {
		if comma, ok := node.(ir.Comma); ok {
			if call, ok := comma.Right.(ir.Call); ok && program.Functions[call.Function].Name == "never_read" {
				evaluated = append(evaluated, comma.Left)
			}
		}
		return true
	}
	walk(program.Main, scan)
	for _, function := range program.Functions {
		walk(function.Body, scan)
	}
	return evaluated
}

// Adamic's condition on ruling 2 (#bbssdcy, Nov 24 18:03Z): the stop on arrow.name fires without reading
// the slot, so a value in it never reaches a never read. Both readings of the earlier lane's program lower,
// one reading arrow.name before the write through base and one after it, and in each the trap's comma
// evaluates arrow (an ir.Read of the local arrow) and holds no ir.Property at all, and nothing anywhere
// loads name through arrow. The same walk finds the loads it looks for where they are: arrow.kind, and
// base.name, a read of the same name through the writable alias, whose field isn't never. This lowers
// only and builds nothing. Mutant: tmp-audit/wrong-stop-loads-slot.patch loads arrow.name before the trap,
// discarded, and both rows find it.
func TestNeverFieldReadLoadsNoSlot(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"readonly-view-read-before.ts", "readonly-view-read-after.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/never_field/" + name)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSourceAs(t, string(source), ".ts")
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if where := neverReadWhere(t, string(source), "arrow.name"); !neverFieldTrapped(program, "read", "arrow.name", where) {
				t.Fatalf("want a trap naming the read of arrow.name at %s, got traps %q", where, neverFieldTrapMessages(program))
			}
			evaluated := neverFieldTrapsEvaluated(program)
			if len(evaluated) != 1 {
				t.Fatalf("want one never_read trap, got %d", len(evaluated))
			}
			if !neverFieldReadsLocal(program, evaluated[0], "arrow") {
				t.Fatalf("the trap evaluates %#v, which doesn't read arrow", evaluated[0])
			}
			walk(evaluated[0], func(node any) bool {
				if property, ok := node.(ir.Property); ok {
					t.Errorf("the trap loads field %s before it stops", property.Name)
				}
				return true
			})
			if loads := neverFieldLoads(program, "arrow", "name"); loads != 0 {
				t.Errorf("arrow.name is loaded %d times, want none", loads)
			}
			if neverFieldLoads(program, "arrow", "kind") == 0 || neverFieldLoads(program, "base", "name") == 0 {
				t.Errorf("the walk finds no load of arrow.kind or base.name, so it couldn't see one of arrow.name")
			}
		})
	}
}
