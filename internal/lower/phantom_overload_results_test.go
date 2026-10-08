package lower

import (
	"errors"
	"reflect"
	"testing"
)

const overloadResultFix = "make the implementation result covariant with every overload result"

func TestPhantomOverloadLiteralResultRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function narrowed(value: string): 'a'; function narrowed(value: string): string { return value; } console.log(narrowed('b'));`)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != `overload 1 of narrowed result "a" cannot be served by implementation result string` || refused.Fix != overloadResultFix {
		t.Fatalf("got %v, want named overload result refusal and fix", err)
	}
}

func TestPhantomOverloadAnyBrandRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `type Path = string & { readonly __pathBrand: any };`)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a primitive brand member __pathBrand whose type is not void" || refused.Fix != "make __pathBrand void (or optional and typed undefined) so the brand is phantom" {
		t.Fatalf("got %v, want __pathBrand refusal with void fix", err)
	}
}

func TestPhantomOverloadResultCastsAreErased(t *testing.T) {
	t.Parallel()
	branded := `type Path = string & { readonly __pathBrand: void };
function path(value: Path): Path;
function path(value: string): Path;
function path(value: string): string { return value + '/'; }
console.log(path('built' + 'path'));`
	plain := `type Path = string;
function path(value: string): string { return value + '/'; }
console.log(path('built' + 'path'));`
	got, err := lowerSource(t, branded)
	if err != nil {
		t.Fatal(err)
	}
	want, err := lowerSource(t, plain)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("phantom overload results introduce runtime operations or change representation")
	}
}

func TestPhantomOverloadUnusedResultRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function narrowed(value: 'a'): 'a'; function narrowed(value: string): string; function narrowed(value: string): string { return value; }`)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != `overload 1 of narrowed result "a" cannot be served by implementation result string` || refused.Fix != overloadResultFix {
		t.Fatalf("got %v, want unused literal overload refusal", err)
	}
}

func TestPhantomOverloadParameterProof(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Animal {readonly kind:string} interface Dog extends Animal {readonly bark:()=>string} function bad(value: Dog[]): string; function bad(value: Animal[]): string {value.push({kind:'animal'}); return 'ok';}`)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "overload 1 of bad parameter value cannot be served by implementation parameter value" {
		t.Fatalf("got %v, want unsafe overload parameter refusal", err)
	}
}

func TestPhantomOverloadFunctionValueNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function value(text: string): string; function value(text: string): string {return text;} const read = value;`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "an overloaded function as a value" {
		t.Fatalf("got %v, want overload value not yet", err)
	}
}

func TestPhantomOverloadBrandLiteralConstraintRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `type Brand = 'a' & { marker: void }; function value(text: string): Brand; function value(text: string): string {return text;}`)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "overload 1 of value result Brand cannot be served by implementation result string" || refused.Fix != overloadResultFix {
		t.Fatalf("got %v, want branded literal constraint refusal", err)
	}
}

func TestPhantomOverloadArrayResult(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `type Brand = readonly number[] & { marker: undefined }; function branded(value: readonly number[]): Brand; function branded(value: readonly number[]): readonly number[] {return value;} const a = branded([1]); console.log(String(a.marker));`)
	if err != nil {
		t.Fatal(err)
	}
}
