package lower

import (
	"strings"
	"testing"
)

// A method seen as a writable field is the field's slot (#9dmh4em): the view may write only what the
// method's type holds, so (): never seen as a writable () => void is refused where the alias is made, in .a
// and .ts, for an object literal's method and a class's. Node runs the replacement there, and a proof that
// trusted the method's body (#8n7eet3's neverCall) read the wrong answer. A readonly field and a method seen
// as a method write nothing, and the same type writes only what the method holds, so each lowers.
// Mutant: method-field-widening.mutants.md.
func TestMethodSeenAsWritableFieldIsRefused(t *testing.T) {
	t.Parallel()
	const literal = "const H = {\n\tfail(): never {\n\t\tthrow new Error('no');\n\t},\n};\n"
	const class = "class F {\n\tfail(): never {\n\t\tthrow new Error('no');\n\t}\n}\nconst H = new F();\n"
	const alias = "const G: { fail: () => void } = H;\nG.fail = (): void => {};\nconsole.log('replaced');\n"
	const what = "a value of type "
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source, where, want string }{
			{"an object literal's method", literal + alias, "6:33", "whose method fail becomes a writable field: a write of () => void would replace a method of type () => never"},
			{"a class's method", class + alias, "7:33", "whose method fail becomes a writable field: a write of () => void would replace a method of type () => never"},
			{"a readonly field", literal + "const G: { readonly fail: () => void } = H;\nconsole.log(typeof G.fail);\n", "", ""},
			{"a method seen as a method", literal + "const G: { fail(): void } = H;\nconsole.log(typeof G.fail);\n", "", ""},
			{"the same type", literal + "const G: { fail: () => never } = H;\nconsole.log(typeof G.fail);\n", "", ""},
		} {
			t.Run(extension+"/"+row.name, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, row.source, extension)
				if row.want == "" {
					if err != nil && strings.Contains(err.Error(), "becomes a writable field") {
						t.Fatalf("refused as a method field: %v", err)
					}
					return
				}
				if err == nil || !strings.Contains(err.Error(), "/main"+extension+":"+row.where+": ") || !strings.Contains(err.Error(), what) || !strings.Contains(err.Error(), row.want) {
					t.Errorf("got %v, want main%s:%s: ...%s", err, extension, row.where, row.want)
				}
			})
		}
	}
}
