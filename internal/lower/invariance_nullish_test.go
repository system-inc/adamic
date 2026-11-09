package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// A nullish value serves its own union member, without widening into the rigid T.
// Mutant run and restored: remove widened's nullish-member early return. The undefined
// row refuses at main.a:2:9 with "a value of type undefined seen as T"; null still passes.
func TestNullishMemberIntoWritableTypeParameterUnion(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"undefined", "null"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			source := "function f<T extends { value: number }>(x: T): T | " + value + " {\n\treturn " + value + ";\n}\nconsole.log(String(f({ value: 1 }) === " + value + "));\n"
			lowersAndAgreesWithNode(t, source)
		})
	}
}

func TestUndefinedAbsentFromWritableTypeParameterUnion(t *testing.T) {
	t.Parallel()
	source := "function f<T extends { value: number }>(x: T): T | null {\n\treturn undefined;\n}\nf({ value: 1 });\n"
	_, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
	want := "main.a:2:2: error TS2322: Type 'undefined' is not assignable to type 'T | null'."
	if err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want an error ending %q", err, want)
	}
}

func TestWritableTypeParameterUnionStillRefusesWidening(t *testing.T) {
	t.Parallel()
	source := "function f<T extends { value: number }, Narrow extends T>(x: Narrow): T | undefined {\n\treturn x;\n}\nf({ value: 1 });\n"
	_, err := lowerSource(t, source)
	want := "main.a:2:9: Adamic 0.1 refuses a value of type Narrow seen as T, a type parameter whose constraint { value: number; } can be written, so it can write what Narrow can't hold; take it as Narrow, or constrain T to something readonly, which can't write (adamic/invariant-mutable)"
	var refused *Refused
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want a refusal ending %q", err, want)
	}
}

// The undefined ternary arm serves the optional member before widened reaches the
// open numeric enum arm. E has one member so E.A is its whole open numeric
// enum type, as in utilities.ts:877. Both calls observe the selected arm against Node.
// Mutant run and restored via a Go overlay: remove widened's nullish-member
// early return. This pin refuses at main.a:3:9 with "an unproven value assigned
// to a numeric literal or enum member slot E" (adamic/enum-literal).
func TestUndefinedMemberIntoOpenNumericEnumUnion(t *testing.T) {
	t.Parallel()
	const source = "enum E { A = 1 }\nfunction f(flag: boolean): E | undefined {\n\treturn flag ? E.A : undefined;\n}\nconsole.log(String(f(true)));\nconsole.log(String(f(false)));\n"
	lowersAndAgreesWithNode(t, source)
}

// The checker rejects the literal 7 before lowering; a value typed number is
// allowed into a whole numeric enum. Keep the literal twin's exact checker stop.
func TestArbitraryNumberIntoOptionalOpenNumericEnum(t *testing.T) {
	t.Parallel()
	const source = "enum E { A = 1 }\nfunction f(flag: boolean): E | undefined {\n\treturn flag ? E.A : 7;\n}\nf(false);\n"
	_, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
	want := "main.a:3:22: error TS2322: Type '7' is not assignable to type 'E'."
	if err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want an error ending %q", err, want)
	}
}
