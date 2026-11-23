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
// and stays refused when the members' types differ. The two admitted rows go past the cast to the read after
// it, f(a).m(), which stops at the callable-field read through a checked view (rulings.md:27), its own road;
// the readonly row went past before too, since a union's property is readonly when any member's is. Measured
// (J-p01898d). Mutant: castUnionUnwritten taken out, so the methods row is refused again as "writable field m".
func TestCastUnionMethodsAreNoSlot(t *testing.T) {
	t.Parallel()
	const tagged = "type A = { readonly kind: 'a'; m(): number };\n"
	for _, row := range []struct {
		name, source string
		// stop is what the program's first stop must say.
		stop string
	}{
		{"methods in every member", tagged + "type B = { readonly kind: 'b'; m(): string };\n" +
			"function f(x: A | B): A { return x as A; }\n" +
			"const a: A = { kind: 'a', m(): number { return 1; } };\n" +
			"console.log(String(f(a).m()));\n", "5:20: stage 0 can't lower a callable field m read through a checked view"},
		{"a method beside a readonly field", tagged + "type B = { readonly kind: 'b'; readonly m: () => string };\n" +
			"function f(x: A | B): A { return x as A; }\n" +
			"const a: A = { kind: 'a', m(): number { return 1; } };\n" +
			"console.log(String(f(a).m()));\n", "5:20: stage 0 can't lower a callable field m read through a checked view"},
		{"a method beside a writable field", tagged + "type B = { readonly kind: 'b'; m: () => string };\n" +
			"function f(x: A | B): A { return x as A; }\n" +
			"const a: A = { kind: 'a', m(): number { return 1; } };\n" +
			"console.log(String(f(a).m()));\n", "3:34: Adamic 0.1 refuses a cast the runtime can't check: a cast over A | B, whose writable field m"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			t.Logf("%s: %v", row.name, err)
			if err == nil || !strings.Contains(err.Error(), row.stop) {
				t.Fatalf("got %v, want the first stop %q", err, row.stop)
			}
		})
	}
}

// A write to the method's name through such a cast stops on its own road (Compiler's probe on #b5w3ycg, J-39z6vc8):
// the write target reads the method as a value, which unbound-method refuses, written through the cast or through
// its saved value alike. So the name castUnionUnwritten lets past is no slot a write through the cast reaches. The
// cast itself, with no write, lowers and agrees with Node (native and JavaScript in the probe, JavaScript here).
func TestCastUnionMethodWriteStops(t *testing.T) {
	t.Parallel()
	const header = "type A = { readonly kind: 'a'; m(): number };\n" +
		"type B = { readonly kind: 'b'; m(): string };\n" +
		"function pick(n: number): A | B { return n > 0 ? { kind: 'a', m(): number { return 1; } } : { kind: 'b', m(): string { return 's'; } }; }\n" +
		"const u = pick(1);\n"
	for _, row := range []struct{ name, source, stop string }{
		{"through the cast", header + "(u as A).m = (): number => 2;\nconsole.log(String((u as A).m()));\n", "5:1: Adamic 0.1 refuses a method read as a value"},
		{"through the cast's value", header + "const a = u as A;\na.m = (): number => 2;\nconsole.log(String(a.m()));\n", "6:1: Adamic 0.1 refuses a method read as a value"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), row.stop) {
				t.Fatalf("got %v, want the first stop %q", err, row.stop)
			}
		})
	}
	t.Run("the cast alone", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, header+"const a = u as A;\nconsole.log(String(a.kind));\n", ".ts")
	})
}
