package lower

import "testing"

// A probe for #k8ed1f2, never for landing: a literal's method with an explicit this parameter can push the
// literal into a { tag: string }[] while the view writes undefined into its tag. It logs whether main lowers it.
func TestProbeLiteralThisParameter(t *testing.T) {
	source := "const stash: { tag: string }[] = [];\n" +
		"const seen: { tag: string | undefined; keep(): void } = { tag: \"a\", keep(this: { tag: string }) { stash.push(this); } };\n" +
		"seen.keep();\n" +
		"seen.tag = undefined;\n" +
		"console.log(String(stash[0]?.tag));\n"
	for _, extension := range []string{".a", ".ts"} {
		_, err := lowerSourceAs(t, source, extension)
		t.Logf("PROBE this-parameter %s: %v", extension, err)
	}
}
