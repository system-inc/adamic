package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestScalarCastPreservesUncheckableRefusals(t *testing.T) {
	for _, source := range []string{
		`function select(value: any): string { return value as string; }`,
		`function select(value: unknown): string { return value as string; }`,
		`declare const brand: unique symbol; type Brand = string & { readonly [brand]: true }; function select(value:string):Brand{return value as Brand;}`,
		`function select(value:string|number):string{return value as unknown as string;}`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "cast") && !strings.Contains(err.Error(), "primitive brand") {
			t.Fatalf("uncheckable scalar assertion must refuse, got %v: %s", err, source)
		}
	}
}

func TestScalarCastWholeNumericEnumStaysOpen(t *testing.T) {
	program, err := lowerSource(t, `enum Flags { A=1, B=2 }; const flags = 8 as Flags; console.log(String(flags));`)
	if err != nil {
		t.Fatal(err)
	}
	// If the source needs a checked helper, it must not acquire an enumerator
	// whitelist. Numeric enum domains intentionally include undeclared numbers.
	for _, function := range program.Functions {
		if function.Name != "checked_scalar_cast" {
			continue
		}
		outer := function.Body[0].(ir.If)
		if len(outer.Then) != 1 {
			t.Fatal("whole numeric enum acquired a finite-member check")
		}
	}
}
