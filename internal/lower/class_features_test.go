package lower

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestClassFeaturesPrivateChecker(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(
		`class Box { #value = 'private'; #method(): string { return this.#value; } read(): string { return this.#method(); } } const box = new Box(); console.log(box.#value); console.log(box.#method());`,
	), 0644); err != nil {
		test.Fatal(err)
	}
	_, err := load.Load([]string{path})
	if err == nil || !strings.Contains(err.Error(), "private identifier") {
		test.Fatalf("want tsc private scope diagnostic, got %v", err)
	}
}

func TestClassFeaturesPrivateStorage(test *testing.T) {
	test.Parallel()
	source, err := os.ReadFile("../oracle/testdata/class_features_private.a")
	if err != nil {
		test.Fatal(err)
	}
	lowersAndAgreesWithNode(test, string(source))
}

func TestClassFeaturesAccessorRefusals(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`class Base { get value(): string { return 'base'; } set value(value: string) {} } class Child extends Base { get value(): string { return 'child'; } set value(value: 'one') {} } const child = new Child(); console.log(child.value);`,
		`class Base { get value(): string { return 'base'; } set value(value: string) {} } class Child extends Base { get value(): string { return 'child'; } } const child = new Child(); console.log(child.value);`,
		`class Base { get value(): 'one' | 'two' { return 'one'; } set value(value: 'one' | 'two') {} } class Child extends Base { get value(): 'one' { return 'one'; } set value(value: 'one') {} } const child = new Child(); console.log(child.value);`,
	} {
		_, err := lowerSource(test, source)
		if err == nil || !strings.Contains(err.Error(), "accessor override") {
			test.Fatalf("want unsafe accessor override refused, got %v", err)
		}
	}
}

func TestClassFeaturesStaticDeclarationsExecute(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`class Box { static { console.log('static side effect'); } } console.log('done');`,
		`class Box { static value = 'initialized'; static { console.log(this.value); } } console.log('done');`,
	} {
		lowersAndAgreesWithNode(test, source)
	}
}

func TestClassFeaturesStaticSoundness(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`class Box { static first = this.read(); static later = 'ready'; static read(): string { return this.later; } } console.log(Box.first);`,
		`class Box { static first = read(); static later = 'ready'; } function read(): string { return Box.later; } console.log(Box.first);`,
		`class Box { static first = read(); static later = 'ready'; } function read(): string { const alias=Box; return alias.later; } console.log(Box.first);`,
		`const read=():string=>Box.later; class Box { static first=Array.from({length:1},read); static later='ready'; } console.log('done');`,
		`class Base { static method(value: string): string { return value; } } class Child extends Base { static override method(value: 'one'): string { return value; } } console.log(Child.method('one'));`,
		`class Base { static field: string = 'one'; } class Child extends Base { static override field: 'one' = 'one'; } console.log(Child.field);`,
		`class Box { static value: typeof Box | undefined = undefined; } Box.value = Box;`,
		`class Box { static value='ready'; static { read(); } } function read(): void { console.log(Box.value); }`,
		`class Base { static child: typeof Child | undefined = undefined; } class Child extends Base {} Base.child = Child;`,
		`class Box { static #value: typeof Box | undefined = undefined; static set(): void { this.#value=this; } } Box.set();`,
	} {
		if strings.Contains(source, "Box.value = Box;") ||
			strings.Contains(source, "Base.child = Child;") ||
			strings.Contains(source, "this.#value=this;") {
			graphAccepted(test, source)
			continue
		}
		_, err := lowerSource(test, source)
		if err == nil {
			test.Errorf("unsafe static program accepted: %s", source)
		}
	}
}

func TestClassFeaturesAccessorCaptureCycle(test *testing.T) {
	test.Parallel()
	program := graphAccepted(test, `function make(): { readonly value: string } {
        let source: { readonly value: string } | undefined;
        const literal = { get value(): string { return source === undefined ? 'empty' : 'cycle'; } };
        source = literal;
        return literal;
    } console.log(make().value);`)
	if len(program.GraphTypes) == 0 {
		test.Fatal("accessor environment stayed counted")
	}
}

func TestClassFeaturesNarrowedAccessor(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`let calls = 0; const source: { readonly value: string | undefined } = { get value(): string | undefined { calls++; return calls === 1 ? 'first' : undefined; } }; if (source.value !== undefined) console.log(source.value);`,
	)
	if err == nil || !strings.Contains(err.Error(), "narrowed accessor reread") {
		test.Fatalf("want changing getter reread refused, got %v", err)
	}
}

func TestClassFeaturesStaticInterfaceCycle(test *testing.T) {
	test.Parallel()
	program := graphAccepted(
		test,
		`type Constructable={new(required:string):Child}; class Base { static child: Constructable | undefined = undefined; } class Child extends Base { readonly tag='child'; constructor(required: string) { super(); console.log(required); } } Base.child=Child;`,
	)
	if len(program.GraphTypes) == 0 {
		test.Fatal("missing graph ownership")
	}
}

func TestClassFeaturesThrowingAccessorSpread(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		`const source = { get text(): string { throw new Error('fail'); } }; try { const copied = { ...source }; console.log(copied.text); } catch (error) { if (error instanceof Error) console.log(error.message); else throw error; }`,
	)
}
