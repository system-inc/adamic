package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// These are checker rules, not runtime freezing. A mutable object held in a readonly field remains mutable.
func TestClassFeaturesReadonlyChecker(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`class Box { readonly value = 'initial'; } const box = new Box(); box.value = 'changed';`,
		`class Box { readonly value = 'initial'; change(): void { this.value = 'changed'; } }`,
		`class Base { readonly value: string; constructor() { this.value = 'base'; } } class Child extends Base { constructor() { super(); this.value = 'child'; } }`,
		`class Box { static readonly value = 'initial'; static { this.value = 'changed'; } }`,
		`class Box { readonly #value = 'initial'; change(): void { this.#value = 'changed'; } }`,
	} {
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := load.Load([]string{path})
		if err == nil || !strings.Contains(err.Error(), "read-only") {
			t.Fatalf("want tsc readonly diagnostic, got %v", err)
		}
	}
	_, err := lowerSource(t, `class Box { readonly value: string; readonly item = { text: 'initial' }; constructor() { this.value = 'set'; this.value = 'again'; } } const box = new Box(); box.item.text = 'changed'; console.log(box.item.text);`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClassFeaturesPrivateChecker(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(`class Box { #value = 'private'; #method(): string { return this.#value; } read(): string { return this.#method(); } } const box = new Box(); console.log(box.#value); console.log(box.#method());`), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := load.Load([]string{path})
	if err == nil || !strings.Contains(err.Error(), "private identifier") {
		t.Fatalf("want tsc private scope diagnostic, got %v", err)
	}
}

func TestClassFeaturesPrivateStorage(t *testing.T) {
	t.Parallel()
	checked, err := load.Load([]string{"../oracle/testdata/class_features_private.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	hidden := 0
	for _, class := range program.Classes {
		for _, field := range class.Fields {
			if field.Private {
				hidden++
			}
		}
	}
	if hidden < 2 {
		t.Fatalf("private storage lost its visibility metadata: %d", hidden)
	}
}

func TestClassFeaturesAccessorRefusals(t *testing.T) {
	for _, source := range []string{
		`class Base { get value(): string { return 'base'; } set value(value: string) {} } class Child extends Base { get value(): string { return 'child'; } set value(value: 'one') {} } const child = new Child(); console.log(child.value);`,
		`class Base { get value(): string { return 'base'; } set value(value: string) {} } class Child extends Base { get value(): string { return 'child'; } } const child = new Child(); console.log(child.value);`,
		`class Base { get value(): 'one' | 'two' { return 'one'; } set value(value: 'one' | 'two') {} } class Child extends Base { get value(): 'one' { return 'one'; } set value(value: 'one') {} } const child = new Child(); console.log(child.value);`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "accessor override") {
			t.Fatalf("want unsafe accessor override refused, got %v", err)
		}
	}
	_, err := lowerSource(t, `const source = { get text(): string { throw new Error('fail'); } }; const copied = { ...source }; console.log(copied.text);`)
	if err == nil || !strings.Contains(err.Error(), "getter may throw") {
		t.Fatalf("want throwing spread diagnosed, got %v", err)
	}
}

func TestClassFeaturesStaticDeclarationsExecute(t *testing.T) {
	for _, source := range []string{
		`class Box { static { console.log('static side effect'); } } console.log('done');`,
		`class Box { static value = 'initialized'; static { console.log(this.value); } } console.log('done');`,
	} {
		program, err := lowerSource(t, source)
		if err != nil {
			t.Fatal(err)
		}
		if len(program.Classes) == 0 || !program.Classes[0].Static || len(program.Main) < 2 {
			t.Fatal("static initialization was dropped")
		}
	}
}

func TestClassFeaturesStaticSoundness(t *testing.T) {
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
		program, err := lowerSource(t, source)
		if strings.Contains(source, "= Box;") || strings.Contains(source, "Base.child = Child;") || strings.Contains(source, "this.#value=this;") {
			if err != nil || len(program.GraphTypes) == 0 {
				t.Errorf("want static graph ownership, got %v", err)
			}
			continue
		}
		if err == nil {
			t.Errorf("unsafe static program accepted: %s", source)
		}
	}
}

func TestClassFeaturesAccessorCaptureCycle(t *testing.T) {
	program, err := lowerSource(t, `function make(): { readonly value: string } {
        let source: { readonly value: string } | undefined;
        const literal = { get value(): string { return source === undefined ? 'empty' : source.value; } };
        source = literal;
        return literal;
    } console.log(make().value);`)
	if err != nil || len(program.GraphTypes) == 0 {
		t.Fatalf("want graph ownership, got %v", err)
	}
}

func TestClassFeaturesNarrowedAccessor(t *testing.T) {
	_, err := lowerSource(t, `let calls = 0; const source: { readonly value: string | undefined } = { get value(): string | undefined { calls++; return calls === 1 ? 'first' : undefined; } }; if (source.value !== undefined) console.log(source.value);`)
	if err == nil || !strings.Contains(err.Error(), "narrowed accessor reread") {
		t.Fatalf("want changing getter reread refused, got %v", err)
	}
}

func TestClassFeaturesStaticParentCycle(t *testing.T) {
	program, err := lowerSource(t, `class Base { static child: typeof Child | undefined = undefined; } class Child extends Base { readonly tag='child'; constructor(required: string) { super(); console.log(required); } } Base.child=Child;`)
	if err != nil || len(program.GraphTypes) == 0 {
		t.Fatalf("want graph ownership, got %v", err)
	}
}

func TestClassFeaturesStaticInterfaceCycle(t *testing.T) {
	program, err := lowerSource(t, `type Constructable={new(required:string):Child}; class Base { static child: Constructable | undefined = undefined; } class Child extends Base { readonly tag='child'; constructor(required: string) { super(); console.log(required); } } Base.child=Child;`)
	if err != nil || len(program.GraphTypes) == 0 {
		t.Fatalf("want graph ownership, got %v", err)
	}
}
