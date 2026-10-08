package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func TestProvenRelationsRefuse(t *testing.T) {
	t.Parallel()
	animals := `interface Animal { readonly name: string }
interface Dog extends Animal { readonly bark: () => string }
const dog: Dog = { name: 'rex', bark: () => 'woof' };
const dogs: Dog[] = [dog];
`
	for _, probe := range []struct{ name, source, part string }{
		{"satisfies mutable array", animals + `const checked = dogs satisfies Animal[];`, "can write Animal where Dog is read"},
		{"upcast mutable array", animals + `const checked = dogs as Animal[];`, "can write Animal where Dog is read"},
		{"satisfies mutable field", animals + `const holder: { pet: Dog } = { pet: dog }; const checked = holder satisfies { pet: Animal };`, "can write Animal where Dog is read"},
		{"satisfies nested mutable literal", animals + `const checked = ({ pets: dogs } satisfies { readonly pets: Animal[] });`, "can write Animal where Dog is read"},
		{"satisfies readonly to writable", animals + `const holder: { readonly pet: Dog } = { pet: dog }; const checked = holder satisfies { pet: Dog };`, "readonly field pet becomes writable"},
		{"satisfies function parameter", animals + `interface Handler { run(value: Animal): string } const handler = { run: (value: Dog): string => value.bark() }; const checked = handler satisfies Handler;`, "function taking Dog"},
		{"upcast function parameter", animals + `interface Handler { run(value: Animal): string } const handler = { run: (value: Dog): string => value.bark() }; const checked = handler as Handler;`, "function taking Dog"},
		{"upcast nominal identity", `class A { read(): string { return 'a'; } } class B { read(): string { return 'b'; } } const b = new B(); const checked = b as A;`, "nominal ancestry for A"},
		{"satisfies nominal identity", `class A { read(): string { return 'a'; } } const checked = ({ read: () => 'b' } satisfies A);`, "nominal ancestry for A"},
		{"upcast mutable optional elements", `interface Named { readonly name: string } interface Counted extends Named { readonly count?: number } const values: Counted[] = []; const checked = values as Named[];`, "optional field element.count"},
		{"satisfies mutable optional field", `interface Named { readonly name: string } interface Counted extends Named { readonly count?: number } const holder: { pet: Counted } = { pet: { name: 'a' } }; const checked = holder satisfies { pet: Named };`, "optional field pet.count"},
		{"upcast invariant optional class argument", `interface Named { readonly name: string } interface Counted extends Named { readonly count?: number } class Box<T> { readonly value: T; constructor(value: T) { this.value = value; } } const box = new Box<Counted>({ name: 'a' }); const checked = box as Box<Named>;`, "optional field type argument.count"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.part) {
				t.Fatalf("want refusal naming %q, got %v", probe.part, err)
			}
		})
	}
}

// Erasure is held by comparing emitted artifacts to the operand alone, not by counting IR nodes.
func TestProvenRelationsErase(t *testing.T) {
	t.Parallel()
	prefix := `const dog: { readonly name: string; readonly extra: number } = { name: 'rex', extra: 4 }; const value = `
	suffix := `; console.log(value.name);`
	plain, err := lowerSource(t, prefix+"dog"+suffix)
	if err != nil {
		t.Fatal(err)
	}
	for _, relation := range []string{"dog satisfies { readonly name: string }", "dog as { readonly name: string }"} {
		program, err := lowerSource(t, prefix+relation+suffix)
		if err != nil {
			t.Fatal(err)
		}
		if native.C(program) != native.C(plain) {
			t.Errorf("%s added native work", relation)
		}
		if javascript.JavaScript(program) != javascript.JavaScript(plain) {
			t.Errorf("%s added JavaScript work", relation)
		}
	}
}
