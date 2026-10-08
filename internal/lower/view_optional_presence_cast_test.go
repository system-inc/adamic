package lower

import (
	"strings"
	"testing"
)

func TestOptionalPresenceViewCast(t *testing.T) {
	program, err := lowerSource(t, `interface Message { flag?: boolean } function diag(flag?: boolean) { return { flag }; } const value = diag() as Message; console.log(String(value.flag));`)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.ViewOrigins) == 0 || !program.CheckedFields["flag"] {
		t.Fatal("optional presence cast lost its checked view")
	}
}

func TestOptionalPresenceViewPreservesSlotTypes(t *testing.T) {
	for _, source := range []string{
		`interface Message { flag?: boolean } const raw: { flag: boolean | undefined; other: string } = { flag: true, other: "original" }; const value = raw as Message;`,
		`interface Message { readonly flag?: boolean } function diag(flag?: boolean) { return { flag }; } const value = diag() as Message;`,
		`interface Message { flag?: true } function diag(flag?: boolean) { return { flag }; } const value = diag() as Message;`,
		`interface Message { flag?: boolean; items: (number | string)[] } const raw: { flag: boolean | undefined; items: number[] } = { flag: true, items: [1] }; const value = raw as Message;`,
	} {
		_, err := lowerSource(t, source)
		if err == nil {
			t.Fatalf("unproved source slot change admitted: %s", source)
		}
		if !strings.Contains(err.Error(), "cast") && !strings.Contains(err.Error(), "writable-slot") {
			t.Fatalf("unexpected refusal: %v", err)
		}
	}
}
