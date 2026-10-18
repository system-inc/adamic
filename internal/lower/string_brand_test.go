package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestStringBrandRepresentation(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`type Target = string & { readonly __brand: void };`,
		`type Target = (string & { __brand: void }) | undefined;`,
		`enum Name { Call="__call", Missing="__missing" } type Target = (string & { __brand: void }) | (void & { __brand: void }) | Name;`,
		`enum Name { Call="__call" } type Target = (string & { __brand: void }) | (void & { __brand: void }) | Name | undefined;`,
	} {
		l, target, release := mixedUnionLowering(t, source)
		held, known := l.representation(target)
		release()
		if !known || held != ir.String {
			t.Fatalf("want plain string: %s, got %v %t", source, held, known)
		}
	}
}

func TestStringBrandRepresentationRejectsLayout(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`type Target = string & { actual: number };`,
		`type Target = string & { length: void };`,
		`type Target = string & { "0": void };`,
		`type Target = string & { (): void };`,
		`type Target = string & { new(): object };`,
		`type Target = string & { [key: string]: void };`,
		`type Target = (string & { __brand: void }) | number;`,
		`type Target = (string & { __brand: void }) | null;`,
		`type Target = void & { __brand: void };`,
	} {
		l, target, release := mixedUnionLowering(t, source)
		known, _ := l.stringBrandRepresentation(target)
		release()
		if known {
			t.Fatalf("string proof admitted layout or distinct member: %s", source)
		}
	}
}

func TestStringBrandCastKeepsLiterals(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `type Brand = "only" & { readonly marker: void }; function refine(value:string):Brand { return value as Brand; } console.log(typeof refine);`)
	if err == nil {
		t.Fatal("unproven string literal refinement admitted")
	}
}

func TestStringBrandCastChecksMissing(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `type Brand = string & { readonly marker: void }; function cast(value:Brand | undefined):string { return value as string; } console.log(typeof cast);`)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if _, ok := node.(ir.Defined); ok {
				found = true
			}
			return true
		})
	}
	if !found {
		t.Fatal("missing checked undefined removal")
	}
}
