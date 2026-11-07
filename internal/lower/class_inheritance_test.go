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

func TestInheritanceRefusesUnsoundOverrides(test *testing.T) {
	test.Parallel()
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
		{
			"inherited private cycle",
			`class Link { #next: Link | undefined = undefined; setNext(value: Link): void { this.#next = value; } }
class Child extends Link { readonly label = 'child'; }
const child = new Child(); child.setNext(child);`,
			"cycle-capable",
			"Weak",
		},
		{"inherited cycle", `class Link { next: Link | undefined = undefined; }
class Child extends Link { readonly label = 'child'; }
const child = new Child();
child.next = child;`, "cycle-capable", "Weak"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			if probe.rule == "cycle-capable" {
				graphAccepted(test, probe.source)
				return
			}
			_, lowerSourceError := lowerSource(test, probe.source)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) ||
				!strings.Contains(lowerSourceError.Error(), probe.rule) ||
				!strings.Contains(lowerSourceError.Error(), probe.fix) {
				test.Fatalf("want refusal naming %s with fix %q, got %v", probe.rule, probe.fix, lowerSourceError)
			}
		})
	}
}

func TestInheritanceKeepsCheckerConstructorRules(test *testing.T) {
	test.Parallel()
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
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "main.a")
			if writeError := os.WriteFile(path, []byte(probe.source), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			_, loadError := load.Load([]string{path})
			var checked *load.CheckError
			if !errors.As(loadError, &checked) {
				test.Fatalf("want checker refusal, got %v", loadError)
			}
			if !strings.Contains(loadError.Error(), probe.message) {
				test.Fatalf("want %s, got %v", probe.message, loadError)
			}
		})
	}
}

func TestInheritanceAllowsSoundOverrides(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, `interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
class A { readonly pet: Animal = { name: 'cat' }; accept(value: Dog): Animal { return value; } }
class B extends A { override readonly pet: Dog = { name: 'dog', bark: () => 'woof' }; override accept(value: Animal): Dog { return this.pet; } }
const base: A = new B();
console.log(base.accept({ name: 'dog', bark: () => 'woof' }).name);`)
}

// Class identities retain their nominal parent even when a subclass adds no fields.
func TestInheritanceHasClassIdentity(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		`class A {} class B extends A {} const value: A = new B(); console.log(value instanceof A ? 'base' : 'missing base'); console.log(value instanceof B ? 'child' : 'missing child');`,
	)
}

// Node behavior cannot observe the cycle finder's internal inherited-field traversal.
func TestInheritanceCycleFinderIncludesInheritedFields(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(`class Base { parent: Base | undefined = undefined; }
class Child extends Base { readonly label = 'child'; }`), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	child := file.Statements.Nodes[1]
	finder := cycleFinder{l: &lowering{checker: checked}}
	fields := finder.fields(checked.GetTypeAtLocation(child.Name()))
	for _, field := range fields {
		if field.Name == "parent" &&
			field.Declarations[0].Parent.Kind == ast.KindClassDeclaration &&
			field.Declarations[0].Parent.Name().Text() == "Base" {
			return
		}
	}
	test.Fatal("cycle traversal lost Base.parent when visiting Child")
}

func TestInheritanceRejectsUnsupportedConstructorShapes(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{
			"replacement object",
			`class A {} class B extends A { constructor() { super(); return this; } } const value = new B();`,
			"replacement value",
		},
		{
			"union dispatch",
			`class A { value(): number { return 1; } } class B { first(): number { return 0; } value(): number { return 2; } } function choose(flag: boolean): A | B { return flag ? new A() : new B(); } console.log('' + choose(true).value());`,
			"union of class types",
		},
		{"computed base", `class A {} class B extends (() => A)() {} const value = new B();`, "computed class base"},
		{
			"declare field",
			`class A { value: number = 1; } class B extends A { declare value: number; } const value = new B();`,
			"declare or abstract",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var gap *NotYet
			if !errors.As(lowerSourceError, &gap) || !strings.Contains(lowerSourceError.Error(), probe.want) {
				test.Fatalf("want explicit unsupported construct %q, got %v", probe.want, lowerSourceError)
			}
		})
	}
}

func TestInheritanceRefusesFalseNominalViews(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{
			"weak structural",
			`import type { Weak } from 'adamic'; class A { value(): number { return 1; } } const weak: Weak<A> = { value: () => 2 }; console.log('ok');`,
		},
		{
			"source union",
			`class A { value(): number { return 1; } } class B { value(): number { return 2; } } class C { value(): number { return 3; } } function make(flag: boolean): B | C { return flag ? new B() : new C(); } const value: A = make(true); console.log('ok');`,
		},
		{
			"target union",
			`class A { value(): number { return 1; } } class B { value(): number { return 2; } } const value: A | B = { value: () => 3 }; console.log('ok');`,
		},
		{
			"nested array",
			`class A { value(): number { return 1; } } const values: readonly A[] = [{ value: () => 2 }]; console.log('ok');`,
		},
		{
			"unrelated override",
			`class A { value(): number { return 1; } } class B { value(): number { return 2; } } class Root { accept(value: A): void {} } class Child extends Root { override accept(value: B): void {} } const child = new Child();`,
		},
		{
			"structural object",
			`class A { value(): number { return 1; } } const a: A = { value: () => 2 }; console.log('' + a.value());`,
		},
		{
			"unrelated class",
			`class A { value(): number { return 1; } } class B { value(): number { return 2; } } const a: A = new B(); console.log('' + a.value());`,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var refusal *Refused
			if !errors.As(lowerSourceError, &refusal) ||
				(!strings.Contains(lowerSourceError.Error(), "nominal-class") &&
					!strings.Contains(lowerSourceError.Error(), "contravariant-override")) {
				test.Fatalf("want nominal class refusal, got %v", lowerSourceError)
			}
		})
	}
}

func TestInheritanceRefusesThisBeforeSuperReturns(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(test, `class A {} class B extends A {
 readonly value = 'ready';
 constructor() { const read = () => this.value; console.log(read()); super(); }
} const value = new B();`)
	var refusal *Refused
	if !errors.As(lowerSourceError, &refusal) ||
		!strings.Contains(lowerSourceError.Error(), "this before super returns") ||
		!strings.Contains(lowerSourceError.Error(), "call super(...)") {
		test.Fatalf("want pre-super this refusal with fix, got %v", lowerSourceError)
	}
}

func TestInheritanceKeepsNominalTupleDestructuring(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, `class Source { readonly name = 'source'; read(): string { return this.name; } }
const pair: readonly [boolean, boolean, Source] = [false, true, new Source()];
const [, ignored, source] = pair;
console.log(ignored ? source.read() : 'none');`)
}

func TestInheritanceGenericMonomorphizations(test *testing.T) {
	test.Parallel()
	program := lowersAndAgreesWithNode(test, `class Base {}
class Box<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } read(): T { return this.value; } }
class Pair<T> extends Box<T> { readonly other: T; constructor(value: T, other: T) { super(value); this.other = other; } override read(): T { return this.other; } }
const a = new Pair<number>(1, 2); const b = new Pair<string>('a', 'b');
const c = new Pair<readonly number[]>([1], [2]); const d = new Pair<{ readonly n: number }>({ n: 1 }, { n: 2 });
console.log([a.read().toString(), b.read(), c.read().join(','), d.read().n.toString()].join(':'));`)
	// Node behavior cannot observe the separate native layouts or shared erased definition identity.
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
			test.Error("monomorphizations lost their shared erased identity")
		}
		layouts[class.Fields[0].Value.Type()] = true
	}
	for _, expected := range []ir.Type{ir.Number, ir.String, ir.Array, ir.Object} {
		if !layouts[expected] {
			test.Errorf("missing separate native layout %v", expected)
		}
	}
}

func TestInheritanceRefusesGrowingGenericClasses(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(test, `class Base {}
class Grow<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } next(): Grow<readonly T[]> { return new Grow<readonly T[]>([this.value]); } }
const value = new Grow<number>(1);`)
	var refusal *Refused
	if !errors.As(lowerSourceError, &refusal) ||
		!strings.Contains(lowerSourceError.Error(), "polymorphic recursion") ||
		!strings.Contains(lowerSourceError.Error(), "recursive type arguments unchanged") {
		test.Fatalf("want bounded monomorphization refusal with fix, got %v", lowerSourceError)
	}
}

func TestInheritanceGenericSoundness(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, rule string }{
		{"generic narrowing", `class A<T> { accept(value: T): void {} }
class B<T> extends A<T> { override accept(value: T & { readonly extra: string }): void {} }
const b = new B<string>();`, "contravariant-override"},
		{"generic inherited cycle", `interface Holder { back: Link<Holder> | undefined; }
class Link<T> { slot: T | undefined = undefined; set(value: T): void { this.slot = value; } }
class Child<T> extends Link<T> {}
const child = new Child<Holder>(); const holder: Holder = { back: child }; child.set(holder);`, "cycle-capable"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			if probe.rule == "cycle-capable" {
				graphAccepted(test, probe.source)
				return
			}
			_, lowerSourceError := lowerSource(test, probe.source)
			var refusal *Refused
			if !errors.As(lowerSourceError, &refusal) || !strings.Contains(lowerSourceError.Error(), probe.rule) {
				test.Fatalf("want %s refusal, got %v", probe.rule, lowerSourceError)
			}
		})
	}
}

func TestInheritanceConditionalThisRules(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`class A {} class B extends A { readonly value = 'b'; constructor(flag: boolean) { if (flag) { super(); } const read = () => this.value; console.log(read()); } } const b = new B(false);`,
		`class A { read(): string { return 'a'; } } class B extends A { constructor(flag: boolean) { const read = () => super.read(); console.log(read()); if (flag) { super(); } else { super(); } } } const b = new B(false);`,
		`class A {} class B extends A { readonly value = 'b'; constructor(read: () => string = () => this.value) { super(); console.log(read()); } } const b = new B();`,
	} {
		_, lowerSourceError := lowerSource(test, source)
		var refusal *Refused
		if !errors.As(lowerSourceError, &refusal) ||
			!strings.Contains(lowerSourceError.Error(), "this before super returns") ||
			!strings.Contains(lowerSourceError.Error(), "call super") {
			test.Fatalf("want pre-super binding refusal with fix, got %v", lowerSourceError)
		}
	}
}

func TestInheritanceGenericViewsKeepNominalArguments(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		`class Root { first(): string { return 'first'; } second(): string { return 'second'; } }
class Other { second(): string { return 'other second'; } first(): string { return 'other first'; } }
class Base {}
class Box<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } read(): T { return this.value; } }
const actual = new Box<Other>(new Other()); const view: Box<Root> = actual; console.log(view.read().first());`,
	)
	var refusal *Refused
	if !errors.As(lowerSourceError, &refusal) || !strings.Contains(lowerSourceError.Error(), "nominal ancestry") {
		test.Fatalf("want nominal generic argument refusal, got %v", lowerSourceError)
	}
}

func TestInheritanceGenericNominalConstraints(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`class Root { read(): string { return 'root'; } } class Other { read(): string { return 'other'; } } class Box<T extends Root> { readonly value: T; constructor(value: T) { this.value = value; } } const box = new Box<Other>(new Other());`,
		`class Root { read(): string { return 'root'; } } class Other { read(): string { return 'other'; } } function read<T extends Root>(value: T): string { return value.read(); } console.log(read<Other>(new Other()));`,
	} {
		_, lowerSourceError := lowerSource(test, source)
		var refusal *Refused
		if !errors.As(lowerSourceError, &refusal) ||
			!strings.Contains(lowerSourceError.Error(), "nominal ancestry") ||
			!strings.Contains(lowerSourceError.Error(), "interface") {
			test.Fatalf("want nominal constraint refusal with structural repair, got %v", lowerSourceError)
		}
	}
}

func TestInheritanceNativeSignatureNeighbors(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{
			"identical default",
			`class Base { scale(value: number, factor: number = 2): number { return value * factor; } }
class Child extends Base { override scale(value: number, factor: number = 2): number { return value * factor + 1; } }
const value: Base = new Child(); console.log(value.scale(3).toString());`,
		},
		{"wider same representation", `class Base { scale(value: 3): number { return value; } }
class Child extends Base { override scale(value: number): number { return value + 1; } }
const value: Base = new Child(); console.log(value.scale(3).toString());`},
		{"identical optional", `class Base { scale(factor?: number): number { return factor ?? 2; } }
class Child extends Base { override scale(factor?: number): number { return (factor ?? 2) + 1; } }
const value: Base = new Child(); console.log(value.scale().toString());`},
		{"void result", `class Base { scale(): void {} }
class Child extends Base { override scale(): void { console.log('child'); } }
const value: Base = new Child(); value.scale();`},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			lowersAndAgreesWithNode(test, probe.source)
		})
	}
}

func TestInheritanceNativeSignatureLimits(test *testing.T) {
	test.Parallel()
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
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var gap *NotYet
			if !errors.As(lowerSourceError, &gap) ||
				!strings.Contains(lowerSourceError.Error(), probe.want) ||
				!strings.Contains(lowerSourceError.Error(), "keep the base method's "+probe.fix) {
				test.Fatalf("want NotYet naming %s and its repair, got %v", probe.want, lowerSourceError)
			}
		})
	}
}
