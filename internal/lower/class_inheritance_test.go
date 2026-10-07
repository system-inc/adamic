package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestInheritanceRefusesUnsoundOverrides(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, rule, fix string }{
		{"parameter bivariance", `interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
class A { accept(value: Animal): string { return value.name; } }
class B extends A { override accept(value: Dog): string { return value.bark(); } }
console.log(new B().accept({ name: 'rex', bark: () => 'woof' }));`, "contravariant-override", "accept the base method's parameter type"},
		{"mutable field narrowing", `interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
class A { pet: Animal = { name: 'cat' }; }
class B extends A { override pet: Dog = { name: 'dog', bark: () => 'woof' }; }
console.log(new B().pet.bark());`, "invariant-mutable", "keep the base field's type"},
		{"mutable parameter contents", `interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
class A { accept(values: Dog[]): void {} }
class B extends A { override accept(values: Animal[]): void { values.push({ name: 'cat' }); } }
const value = new B();`, "contravariant-override", "accept the base method's parameter type"},
		{"readonly mutable contents", `interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
class A { readonly pets: Animal[] = []; }
class B extends A { override readonly pets: Dog[] = []; }
const value = new B();`, "invariant-mutable", "make the contents readonly"},
		{"inherited private cycle", `class Link { #next: Link | undefined = undefined; setNext(value: Link): void { this.#next = value; } }
class Child extends Link { readonly label = 'child'; }
const child = new Child(); child.setNext(child);`, "cycle-capable", "Weak"},
		{"inherited cycle", `class Link { next: Link | undefined = undefined; }
class Child extends Link { readonly label = 'child'; }
const child = new Child();
child.next = child;`, "cycle-capable", "Weak"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.rule) || !strings.Contains(err.Error(), probe.fix) {
				t.Fatalf("want refusal naming %s with fix %q, got %v", probe.rule, probe.fix, err)
			}
		})
	}
}

func TestInheritanceKeepsCheckerConstructorRules(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, message string }{
		{"abstract construction", `abstract class A { abstract value(): number; }
new A();`, "abstract"},
		{"this before super", `class A {}
class B extends A { value: number; constructor() { this.value = 1; super(); } }
new B();`, "super"},
		{"abstract method missing", `abstract class A { abstract value(): number; }
class B extends A {}
new B();`, "abstract"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(probe.source), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := load.Load([]string{path})
			var checked *load.CheckError
			if !errors.As(err, &checked) {
				t.Fatalf("want checker refusal, got %v", err)
			}
			if !strings.Contains(err.Error(), probe.message) {
				t.Fatalf("want %s, got %v", probe.message, err)
			}
		})
	}
}

func TestInheritanceAllowsSoundOverrides(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
class A { readonly pet: Animal = { name: 'cat' }; accept(value: Dog): Animal { return value; } }
class B extends A { override readonly pet: Dog = { name: 'dog', bark: () => 'woof' }; override accept(value: Animal): Dog { return this.pet; } }
const base: A = new B();
console.log(base.accept({ name: 'dog', bark: () => 'woof' }).name);`)
	if err != nil {
		t.Fatal(err)
	}
}

// Class identities retain their nominal parent even when a subclass adds no fields.
func TestInheritanceHasClassIdentity(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class A {} class B extends A {} const value: A = new B(); console.log("ok");`)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Classes) != 2 || program.Classes[1].Base != 1 {
		t.Fatalf("missing ancestry: %+v", program.Classes)
	}
}

func TestInheritanceCycleFinderIncludesInheritedFields(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(`class Base { parent: Base | undefined = undefined; }
class Child extends Base { readonly label = 'child'; }`), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	child := file.Statements.Nodes[1]
	finder := cycleFinder{l: &lowering{checker: checked}}
	fields := finder.fields(checked.GetTypeAtLocation(child.Name()))
	for _, field := range fields {
		if field.Name == "parent" && field.Declarations[0].Parent.Kind == ast.KindClassDeclaration && field.Declarations[0].Parent.Name().Text() == "Base" {
			return
		}
	}
	t.Fatal("cycle traversal lost Base.parent when visiting Child")
}

func TestInheritanceRejectsUnsupportedConstructorShapes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"replacement object", `class A {} class B extends A { constructor() { super(); return this; } } const value = new B();`, "replacement value"},
		{"union dispatch", `class A { value(): number { return 1; } } class B { first(): number { return 0; } value(): number { return 2; } } function choose(flag: boolean): A | B { return flag ? new A() : new B(); } console.log('' + choose(true).value());`, "union of class types"},
		{"computed base", `class A {} class B extends (() => A)() {} const value = new B();`, "computed class base"},
		{"declare field", `class A { value: number = 1; } class B extends A { declare value: number; } const value = new B();`, "declare or abstract"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want explicit unsupported construct %q, got %v", probe.want, err)
			}
		})
	}
}

func TestInheritanceRefusesFalseNominalViews(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"weak structural", `import type { Weak } from 'adamic'; class A { value(): number { return 1; } } const weak: Weak<A> = { value: () => 2 }; console.log('ok');`},
		{"source union", `class A { value(): number { return 1; } } class B { value(): number { return 2; } } class C { value(): number { return 3; } } function make(flag: boolean): B | C { return flag ? new B() : new C(); } const value: A = make(true); console.log('ok');`},
		{"target union", `class A { value(): number { return 1; } } class B { value(): number { return 2; } } const value: A | B = { value: () => 3 }; console.log('ok');`},
		{"nested array", `class A { value(): number { return 1; } } const values: readonly A[] = [{ value: () => 2 }]; console.log('ok');`},
		{"unrelated override", `class A { value(): number { return 1; } } class B { value(): number { return 2; } } class Root { accept(value: A): void {} } class Child extends Root { override accept(value: B): void {} } const child = new Child();`},
		{"structural object", `class A { value(): number { return 1; } } const a: A = { value: () => 2 }; console.log('' + a.value());`},
		{"unrelated class", `class A { value(): number { return 1; } } class B { value(): number { return 2; } } const a: A = new B(); console.log('' + a.value());`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refusal *Refused
			if !errors.As(err, &refusal) || (!strings.Contains(err.Error(), "nominal-class") && !strings.Contains(err.Error(), "native-nominal-view") && !strings.Contains(err.Error(), "contravariant-override")) {
				t.Fatalf("want nominal class refusal, got %v", err)
			}
		})
	}
}

func TestInheritanceRefusesThisBeforeSuperReturns(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class A {} class B extends A {
 readonly value = 'ready';
 constructor() { const read = () => this.value; console.log(read()); super(); }
} const value = new B();`)
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "this before super returns") || !strings.Contains(err.Error(), "call super(...)") {
		t.Fatalf("want pre-super this refusal with fix, got %v", err)
	}
}

func TestInheritanceKeepsNominalTupleDestructuring(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class Source { readonly name = 'source'; read(): string { return this.name; } }
const pair: readonly [boolean, boolean, Source] = [false, true, new Source()];
const [, ignored, source] = pair;
console.log(ignored ? source.read() : 'none');`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInheritanceGenericMonomorphizations(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class Base {}
class Box<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } read(): T { return this.value; } }
class Pair<T> extends Box<T> { readonly other: T; constructor(value: T, other: T) { super(value); this.other = other; } override read(): T { return this.other; } }
const a = new Pair<number>(1, 2); const b = new Pair<string>('a', 'b');
const c = new Pair<readonly number[]>([1], [2]); const d = new Pair<{ readonly n: number }>({ n: 1 }, { n: 2 });
console.log('done');`)
	if err != nil {
		t.Fatal(err)
	}
	layouts := map[ir.Type]bool{}
	definition := 0
	for _, class := range program.Classes {
		if !strings.HasPrefix(class.Name, "Box_") {
			continue
		}
		if definition == 0 {
			definition = class.Definition
		}
		if definition != class.Definition {
			t.Error("monomorphizations lost their shared erased identity")
		}
		layouts[class.Fields[0].Value.Type()] = true
	}
	for _, expected := range []ir.Type{ir.Number, ir.String, ir.Array, ir.Object} {
		if !layouts[expected] {
			t.Errorf("missing separate native layout %v", expected)
		}
	}
}

func TestInheritanceRefusesGrowingGenericClasses(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class Base {}
class Grow<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } next(): Grow<readonly T[]> { return new Grow<readonly T[]>([this.value]); } }
const value = new Grow<number>(1);`)
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "polymorphic recursion") || !strings.Contains(err.Error(), "recursive type arguments unchanged") {
		t.Fatalf("want bounded monomorphization refusal with fix, got %v", err)
	}
}

func TestInheritanceGenericSoundness(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, rule string }{
		{"generic narrowing", `class A<T> { accept(value: T): void {} }
class B<T> extends A<T> { override accept(value: T & { readonly extra: string }): void {} }
const b = new B<string>();`, "contravariant-override"},
		{"generic inherited cycle", `interface Holder { back: Link<Holder> | undefined; }
class Link<T> { slot: T | undefined = undefined; set(value: T): void { this.slot = value; } }
class Child<T> extends Link<T> {}
const child = new Child<Holder>(); const holder: Holder = { back: child }; child.set(holder);`, "cycle-capable"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(err.Error(), probe.rule) {
				t.Fatalf("want %s refusal, got %v", probe.rule, err)
			}
		})
	}
}

func TestInheritanceConditionalThisRules(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`class A {} class B extends A { readonly value = 'b'; constructor(flag: boolean) { if (flag) { super(); } const read = () => this.value; console.log(read()); } } const b = new B(false);`,
		`class A { read(): string { return 'a'; } } class B extends A { constructor(flag: boolean) { const read = () => super.read(); console.log(read()); if (flag) { super(); } else { super(); } } } const b = new B(false);`,
		`class A {} class B extends A { readonly value = 'b'; constructor(read: () => string = () => this.value) { super(); console.log(read()); } } const b = new B();`,
	} {
		_, err := lowerSource(t, source)
		var refusal *Refused
		if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "this before super returns") || !strings.Contains(err.Error(), "call super") {
			t.Fatalf("want pre-super binding refusal with fix, got %v", err)
		}
	}
}

func TestInheritanceGenericViewsKeepNominalArguments(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class Root { first(): string { return 'first'; } second(): string { return 'second'; } }
class Other { second(): string { return 'other second'; } first(): string { return 'other first'; } }
class Base {}
class Box<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } read(): T { return this.value; } }
const actual = new Box<Other>(new Other()); const view: Box<Root> = actual; console.log(view.read().first());`)
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "nominal ancestry") {
		t.Fatalf("want nominal generic argument refusal, got %v", err)
	}
}

func TestInheritanceGenericNominalConstraints(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`class Root { read(): string { return 'root'; } } class Other { read(): string { return 'other'; } } class Box<T extends Root> { readonly value: T; constructor(value: T) { this.value = value; } } const box = new Box<Other>(new Other());`,
		`class Root { read(): string { return 'root'; } } class Other { read(): string { return 'other'; } } function read<T extends Root>(value: T): string { return value.read(); } console.log(read<Other>(new Other()));`,
	} {
		_, err := lowerSource(t, source)
		var refusal *Refused
		if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "nominal ancestry") || !strings.Contains(err.Error(), "interface") {
			t.Fatalf("want nominal constraint refusal with structural repair, got %v", err)
		}
	}
}

func TestInheritanceGenericFactoryLayouts(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `class Base {}
class Box<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } read(): T { return this.value; } }
class Projected<T extends { readonly native: number | string }> extends Box<T['native']> {}
function make<T extends { readonly native: number | string }>(value: T['native']): Projected<T> { return new Projected<T>(value); }
const n = make<{ readonly native: number }>(1);
const s = make<{ readonly native: string }>('s');
console.log('done');`)
	if err != nil {
		t.Fatal(err)
	}
	layouts := map[ir.Type]bool{}
	for _, class := range program.Classes {
		if strings.HasPrefix(class.Name, "Box_") {
			layouts[class.Fields[0].Value.Type()] = true
		}
	}
	if !layouts[ir.Number] || !layouts[ir.String] {
		t.Fatalf("generic factories lost their concrete base layouts: %v", layouts)
	}
}

func TestInheritanceNativeSignatureNeighbors(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"identical default", `class Base { scale(value: number, factor: number = 2): number { return value * factor; } }
class Child extends Base { override scale(value: number, factor: number = 2): number { return value * factor + 1; } }
const value: Base = new Child(); console.log(value.scale(3).toString());`},
		{"wider same representation", `class Base { scale(value: 3): number { return value; } }
class Child extends Base { override scale(value: number): number { return value + 1; } }
const value: Base = new Child(); console.log(value.scale(3).toString());`},
		{"identical optional", `class Base { scale(factor?: number): number { return factor ?? 2; } }
class Child extends Base { override scale(factor?: number): number { return (factor ?? 2) + 1; } }
const value: Base = new Child(); console.log(value.scale().toString());`},
		{"void result", `class Base { scale(): void {} }
class Child extends Base { override scale(): void {} }
const value: Base = new Child(); value.scale();`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInheritanceNativeSignatureLimits(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want, fix string }{
		{"optional added", `class Base { scale(factor: number): number { return factor; } }
class Child extends Base { override scale(factor?: number): number { return factor ?? 2; } }
const value: Base = new Child();`, `parameter "factor"`, "parameter form"},
		{"boolean default added", `class Base { scale(flag: boolean): boolean { return flag; } }
class Child extends Base { override scale(flag: boolean = true): boolean { return flag; } }
const value: Base = new Child();`, `parameter "flag"`, "parameter form"},
		{"generic default added", `class Base<T> { scale(factor: T): T { return factor; } }
class Child<T> extends Base<T> { override scale(factor: T = this.fallback): T { return factor; } readonly fallback: T; constructor(fallback: T) { super(); this.fallback = fallback; } }
const value: Base<number> = new Child<number>(2);`, `parameter "factor"`, "parameter form"},
		{"parameter count", `class Base { scale(factor: number): number { return factor; } }
class Child extends Base { override scale(factor: number, extra: number = 2): number { return factor * extra; } }
const value: Base = new Child();`, "parameter count", "parameter count"},
		{"void to number result", `class Base { scale(): void {} }
class Child extends Base { override scale(): number { return 2; } }
const value: Base = new Child();`, "result representation", "result form"},
		{"number to optional result", `class Base { scale(): number | undefined { return 2; } }
class Child extends Base { override scale(): number { return 2; } }
const value: Base = new Child();`, "result representation", "result form"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(err.Error(), probe.want) || !strings.Contains(err.Error(), "keep the base method's "+probe.fix) {
				t.Fatalf("want NotYet naming %s and its repair, got %v", probe.want, err)
			}
		})
	}
}
