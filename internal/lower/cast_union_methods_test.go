package lower

import (
	"strings"
	"testing"
)

// A method every member of a union declares is no slot a write through the union can reach, so a tag-proven
// cast over that union isn't refused for it (castUnionUnwritten, #b5w3ycg: Diagnostics' count found about 13
// of tsc's casts refused this way, getSourceFile over BuilderProgram | CompilerHost among them). The checker's
// property for a union is synthetic and carries no method flag, so castUnionWrites read the method as a
// writable field whose type no member could hold. Each row is judged by its member symbols: methods in every
// member, a method beside a readonly field, and a method beside a writable function field, which stays a slot
// and stays refused when the members' types differ. Mutant: castUnionUnwritten answering false, so the two
// method rows are refused again as "writable field m".
func TestCastUnionMethodsAreNoSlot(t *testing.T) {
	t.Parallel()
	const tagged = "type A = { readonly kind: 'a'; m(): number };\n"
	for _, row := range []struct {
		name, source string
		// refused is the words a refusal must hold, or "" when the cast must not stop as a union write.
		refused string
	}{
		{"methods in every member", tagged + "type B = { readonly kind: 'b'; m(): string };\n" +
			"function f(x: A | B): A { return x as A; }\n" +
			"const a: A = { kind: 'a', m(): number { return 1; } };\n" +
			"console.log(String(f(a).m()));\n", ""},
		{"a method beside a readonly field", tagged + "type B = { readonly kind: 'b'; readonly m: () => string };\n" +
			"function f(x: A | B): A { return x as A; }\n" +
			"const a: A = { kind: 'a', m(): number { return 1; } };\n" +
			"console.log(String(f(a).m()));\n", ""},
		{"a method beside a writable field", tagged + "type B = { readonly kind: 'b'; m: () => string };\n" +
			"function f(x: A | B): A { return x as A; }\n" +
			"const a: A = { kind: 'a', m(): number { return 1; } };\n" +
			"console.log(String(f(a).m()));\n", "writable field m"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			t.Logf("%s: %v", row.name, err)
			switch {
			case row.refused != "" && (err == nil || !strings.Contains(err.Error(), row.refused)):
				t.Fatalf("got %v, want a refusal naming %q", err, row.refused)
			case row.refused == "" && err != nil && strings.Contains(err.Error(), "writable field m"):
				t.Fatalf("got %v, want the cast not refused as a union write of a method", err)
			}
		})
	}
}
