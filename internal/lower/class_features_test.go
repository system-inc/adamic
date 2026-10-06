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

func TestClassFeaturesStaticDeclarationsAreDiagnosed(t *testing.T) {
	for _, source := range []string{
		`class Box { static { console.log('static side effect'); } } console.log('done');`,
		`class Box { static value = console.log('static initializer'); } console.log('done');`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "static class") {
			t.Fatalf("static initialization was dropped: %v", err)
		}
	}
}

func TestClassFeaturesAccessorCaptureCycle(t *testing.T) {
	_, err := lowerSource(t, `function make(): { readonly value: string } {
        let source: { readonly value: string } | undefined;
        const literal = { get value(): string { return source === undefined ? 'empty' : source.value; } };
        source = literal;
        return literal;
    } console.log(make().value);`)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("want captured-cell cycle refused, got %v", err)
	}
}

func TestClassFeaturesNarrowedAccessor(t *testing.T) {
	_, err := lowerSource(t, `let calls = 0; const source: { readonly value: string | undefined } = { get value(): string | undefined { calls++; return calls === 1 ? 'first' : undefined; } }; if (source.value !== undefined) console.log(source.value);`)
	if err == nil || !strings.Contains(err.Error(), "narrowed accessor reread") {
		t.Fatalf("want changing getter reread refused, got %v", err)
	}
}
