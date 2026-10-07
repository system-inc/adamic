package lower

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestAccessorRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, reason string
		notYet               bool
	}{
		{"update value", `class A { get x(): number { return 1; } set x(value: number) {} } const a = new A(); const value = a.x++;`, "accessor-update-value", true},
		{"prefix update value", `class A { get x(): number { return 1; } set x(value: number) {} } const a = new A(); const value = ++a.x;`, "accessor-update-value", true},
		{"compound update value", `class A { get x(): number { return 1; } set x(value: number) {} } const a = new A(); const value = (a.x += 1);`, "accessor-update-value", true},
		{"short circuit update value", `class A { get x(): number { return 1; } set x(value: number) {} } const a = new A(); const value = false && a.x++ > 0;`, "accessor-update-value", true},
		{"setter value", `class A { get x(): number { return 1; } set x(value: number) {} } const a = new A(); const value = (a.x = 2);`, "accessor-update-value", true},
		{"mutable getter result override", `interface Animal { readonly name: string } interface Dog extends Animal { readonly bark: () => string } class A { get pets(): Animal[] { return []; } } class B extends A { override get pets(): Dog[] { return []; } } const b = new B();`, "accessor-override", false},
		{"setter-only read", `class A { set x(value: number) {} } console.log(new A().x.toString());`, "setter-only-read", false},
		{"writable interface", `interface View { x: number } class A { get x(): number { return 1; } } const view: View = new A(); view.x = 2;`, "accessor-field-view", false},
		{"readonly interface", `interface View { readonly x: number } class A { get x(): number { return 1; } } const view: View = new A(); console.log(view.x.toString());`, "accessor-field-view", false},
		{"nested writable interface", `interface View { x: number } class A { get x(): number { return 1; } } const items: readonly View[] = [new A()];`, "accessor-field-view", false},
		{"inferred mixed array", `class A { get total(): number { return 1; } } const items = [new A(), { total: 2 }]; for (const item of items) { console.log(item.total.toString()); }`, "accessor-field-view", false},
		{"inferred mixed array reversed", `class A { get total(): number { return 1; } } const items = [{ total: 2 }, new A()]; for (const item of items) { console.log(item.total.toString()); }`, "accessor-field-view", false},
		{"function view", `interface View { x: number } class A { get x(): number { return 1; } } function make(): A { return new A(); } const f: () => View = make;`, "accessor-field-view", false},
		{"interface inherits accessor", `class A { get x(): number { return 1; } } interface View extends A {} const value: View = new A(); console.log(value.x.toString());`, "accessor-field-view", false},
		{"unrelated getter interface", `class A { get x(): number { return 1; } } interface View extends A {} class B { other(): number { return 3; } get x(): number { return 2; } } const value: View = new B(); console.log(value.x.toString());`, "accessor-field-view", false},
		{"interface getter parameter", `class A { get x(): number { return 1; } } interface View extends A {} function read(value: View): number { return value.x; }`, "accessor-field-view", false},
		{"erased accessor", `interface Empty {} class A { get x(): number { return 1; } } const erased: Empty = new A();`, "accessor-view-erasure", false},
		{"static", `class A { static get x(): number { return 1; } } console.log(A.x.toString());`, "static-accessor", true},
		{"unused static getter", `class A { static get x(): number { return 1; } }`, "static-accessor", true},
		{"unused static setter", `class A { static set x(value: number) {} }`, "static-accessor", true},
		{"inherited static receiver", `class A { static get x(): number { return this === B ? 20 : 10; } static set x(value: number) { console.log(this === B ? 'derived' : 'base'); } } class B extends A {} console.log(B.x.toString()); B.x = 4;`, "static-accessor", true},
		{"generic function receiver", `class A { get x(): number { return 1; } } function read<T extends A>(value: T): number { return value.x; } console.log(read(new A()).toString());`, "generic-accessor-receiver", true},
		{"interface descriptor", `interface View { get x(): number; }`, "accessor-declaration", true},
		{"setter parameter narrowing", `class A { get x(): number { return 1; } set x(value: number) {} } class B extends A { override get x(): number { return 1; } override set x(value: 1) {} } const b = new B();`, "accessor-override", false},
		{"override representation", `class A { get x(): number | string { return 1; } } class B extends A { override get x(): number { return 1; } } const b = new B();`, "accessor-override-representation", true},
		{"literal", `const value = { get x(): number { return 1; } };`, "object-accessor", true},
		{"optional", `class A { get x(): number { return 1; } } function read(a: A | undefined): number | undefined { return a?.x; }`, "optional-accessor", true},
		{"indexed", `class A { get x(): number { return 1; } } const a = new A(); console.log(a['x'].toString());`, "accessor-property-operation", true},
		{"spread", `class A { get x(): number { return 1; } } const a = new A(); const b = { ...a };`, "accessor-property-operation", true},
		{"destructure", `class A { get x(): number { return 1; } } const { x } = new A();`, "accessor-property-operation", true},
		{"parameter destructure", `class A { get x(): number { return 1; } } function read({ x }: A): number { return x; }`, "accessor-property-operation", true},
		{"partial descriptor", `class A { get x(): number { return 1; } set x(value: number) {} } class B extends A { override get x(): number { return 2; } } const b = new B();`, "accessor-descriptor-override", true},
		{"abstract", `abstract class A { abstract get x(): number; }`, "abstract-accessor", true},
		{"computed", `class A { get ['x'](): number { return 1; } }`, "accessor-name", true},
		{"constructor escape", `class A { x: number; get value(): number { return this.x; } constructor() { console.log(this.value.toString()); this.x = 1; } } const value = new A();`, "this escaping", false},
		{"base constructor escape", `class A { get x(): number { return 1; } constructor() { console.log(this.x.toString()); } } class B extends A { field = 1; } const value = new B();`, "this escaping", false},
		{"super initializer escape", `class A { get x(): number { return 1; } } class B extends A { z = super.x; } const value = new B();`, "field initializer", false},
		{"super constructor escape", `class A { get x(): number { return 1; } } class B extends A { y: number; constructor() { super(); console.log(super.x.toString()); this.y = 2; } } const value = new B();`, "this escaping", false},
		{"initializer escape", `class A { get x(): number { return this.y; } z = this.x; y = 1; } const value = new A();`, "field initializer", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			var refused *Refused
			if err == nil || !strings.Contains(err.Error(), probe.reason) || (probe.notYet && !errors.As(err, &notYet)) || (!probe.notYet && !errors.As(err, &refused)) {
				t.Fatalf("want named reason %s (NotYet=%v), got %v", probe.reason, probe.notYet, err)
			}
		})
	}
}

// This observes the actual tree supplied to all downstream analyses. The oracle
// separately decides the behavior, so a call-shaped wrong implementation cannot pass both.
func TestAccessorReadsAreVirtualCallsWithThrowEffects(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class A { get x(): number { return 1; } }
class B extends A { override get x(): number { throw new Error('getter'); } }
function read(value: A): number { return value.x; }
try { console.log(read(new B()).toString()); } catch { console.log('caught'); }`)
	if err != nil {
		t.Fatal(err)
	}
	var read *ir.Function
	for index := range program.Functions {
		if program.Functions[index].Name == "read" {
			read = &program.Functions[index]
		}
	}
	if read == nil || !read.MayThrow {
		t.Fatal("getter reader must propagate a virtual throw")
	}
	returned, ok := read.Body[0].(ir.Return)
	if !ok {
		t.Fatalf("reader body: %#v", read.Body)
	}
	call, ok := returned.Value.(ir.Call)
	if !ok || call.Virtual == 0 || len(program.CallTargets(call)) != 2 || !program.CallMayThrow(call) {
		t.Fatalf("getter must be a virtual call to both targets with MayThrow: %#v", returned.Value)
	}
}

func TestAccessorUpdateSnapshotsReceiverBeforeGetterAndOperand(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class A { get x(): number { return 1; } set x(value: number) {} }
function update(value: A): void { value.x += 2; } update(new A());`)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name != "update" {
			continue
		}
		block, ok := function.Body[0].(ir.Block)
		if !ok || len(block.Body) != 3 {
			t.Fatalf("update block: %#v", function.Body)
		}
		held, ok := block.Body[0].(ir.Declare)
		if !ok {
			t.Fatalf("receiver not held: %#v", block.Body[0])
		}
		current, ok := block.Body[1].(ir.Declare)
		if !ok {
			t.Fatalf("getter not held: %#v", block.Body[1])
		}
		getter, ok := current.Value.(ir.Call)
		if !ok || getter.Virtual == 0 {
			t.Fatalf("getter not called: %#v", current.Value)
		}
		evaluated, ok := block.Body[2].(ir.Evaluate)
		if !ok {
			t.Fatalf("setter not evaluated: %#v", block.Body[2])
		}
		setter, ok := evaluated.Value.(ir.Call)
		receiver := ir.Read{Local: held.Local, Of: ir.Object}
		if !ok || setter.Virtual == 0 || !reflect.DeepEqual(getter.Arguments[0], receiver) || !reflect.DeepEqual(setter.Arguments[0], receiver) {
			t.Fatalf("getter and setter must share the held receiver: %#v %#v", getter, setter)
		}
		return
	}
	t.Fatal("update function missing")
}

func TestAccessorOverrideSeparatesReadAndWriteTypes(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class A { get x(): number { return 1; } set x(value: number) {} }
class B extends A { override get x(): 1 { return 1; } override set x(value: number) {} }
const value: A = new B(); value.x = 2; console.log(value.x.toString());`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenericAccessorOverridesUseSubstitutedTypes(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class Base<T> { stored: T; constructor(value: T) { this.stored = value; } get x(): T { return this.stored; } set x(value: T) { this.stored = value; } }
class Derived<A, B> extends Base<B> { override get x(): B { return this.stored; } override set x(value: B) { this.stored = value; } }
const value: Base<string> = new Derived<number, string>('yes'); console.log(value.x); value.x = 'next';`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenericAccessorOverrideRejectsMutableCovariance(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Animal { readonly name: string } interface Dog extends Animal { readonly bark: () => string }
class Base<T> { get x(): T[] { return []; } }
class Derived<T extends Animal> extends Base<Animal> { override get x(): T[] { return []; } }
const value = new Derived<Dog>();`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "accessor-override") {
		t.Fatalf("want accessor override refusal, got %v", err)
	}
}

func TestSuperAccessorsAreDirectCallsWithCurrentReceiver(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class A { stored = 1; get x(): number { return this.stored; } set x(value: number) { this.stored = value; } }
class B extends A { override get x(): number { return super.x; } override set x(value: number) { super.x = value; } } const value = new B();`)
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, function := range program.Functions {
		if function.Name != "B_get:x" && function.Name != "B_set:x" {
			continue
		}
		var call ir.Call
		if function.Name == "B_get:x" {
			call = function.Body[0].(ir.Return).Value.(ir.Call)
		} else {
			call = function.Body[0].(ir.Evaluate).Value.(ir.Call)
		}
		receiver := call.Arguments[0].(ir.Read)
		if call.Virtual != 0 || receiver.Local != function.Parameters[0] || !strings.HasPrefix(program.Functions[call.Function].Name, "A_") {
			t.Fatalf("super must call the base with current this: %#v", call)
		}
		seen++
	}
	if seen != 2 {
		t.Fatalf("want both descriptor halves, got %d", seen)
	}
}
