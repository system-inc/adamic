package lower

import (
	"errors"
	"strings"
	"testing"
)

// The parameter renames its local so the fix must name the source member.
// Mutants and their observed failures are recorded in method-destructuring.mutants.md.
func TestMethodDestructuringFix(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, member, where string }{
		{"variable", "interface H { f(): number }\nclass C { x = 1; f(): number { return this.x; } }\nconst h: H = new C();\nconst { f } = h;\n", "f", ":4:9:"},
		{"parameter", "interface H { readValue(): number }\nclass C { x = 1; readValue(): number { return this.x; } }\nfunction read({ readValue: detached }: H): number { return detached(); }\nconsole.log(String(read(new C())));\n", "readValue", ":3:17:"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + probe.where + " Adamic 0.1 refuses a method in object destructuring; call " + probe.member + " on its receiver or wrap that call in an arrow, since destructuring would lose this; if nothing that fills " + probe.member + " reads this, declare it as a function-typed property (" + probe.member + ": (...) => T) instead of a method"
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want the complete refusal ending %q", err, want)
				}
			})
		}
	}
}

// The named fix is sound for a function that reads no this, or a receiver call.
func TestMethodDestructuringFixesAgree(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"property", "type P = { f: () => number };\nconst p: P = { f: () => 1 };\nconst { f } = p;\nconsole.log(String(f()));\n"},
		{"receiver", "interface H { f(): number }\nclass C { x = 1; f(): number { return this.x; } }\nconst h: H = new C();\nconsole.log(String(h.f()));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, probe.source, extension)
			})
		}
	}
}

// A function-typed view cannot make a receiver-dependent method safe to
// detach. Keep this premise even if its present NotYet becomes a Refused.
func TestThisReadingMethodInFunctionPropertyDoesNotLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"declaration", "class C { x = 1; f(): number { return this.x; } }\nconst p: { f: () => number } = new C();\nconst { f } = p;\nconsole.log(String(f()));\n"},
		{"parameter", "class C { x = 1; f(): number { return this.x; } }\nfunction g({ f }: { f: () => number }): number { return f(); }\nconsole.log(String(g(new C())));\n"},
		{"read-across-call", "class C { x = 1; f(): number { return this.x; } }\nfunction h(p: { f: () => number }): number { const k = p.f; return k(); }\nconsole.log(String(h(new C())));\n"},
		{"literal", "const lit = { x: 1, f(): number { return this.x; } };\nconst p: { f: () => number } = lit;\nconst { f } = p;\nconsole.log(String(f()));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var refused *Refused
				var later *NotYet
				if !errors.As(err, &refused) && !errors.As(err, &later) {
					t.Fatalf("this-reading method behind a function-typed property must stop, got %v", err)
				}
			})
		}
	}
}
