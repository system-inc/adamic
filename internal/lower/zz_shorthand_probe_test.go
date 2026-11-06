package lower

import (
	"os"
	"testing"
)

func TestZZShorthandProbe(t *testing.T) {
	os.Setenv("ADAMIC_SHORTHAND_PROBE", "1")
	tail := "const r: R = { t: 2 };\nconst x = make(r);\nif (x) { x.p.t = 1; }\nconsole.log(String(r.t));\n"
	head := "interface R { readonly t: number }\ninterface W { t: number }\n"
	for _, variant := range []struct{ name, make string }{
		{"A_narrowed_shorthand", "function make(p: R): { p: W } | undefined { if (p.t > 0) { return { p }; } return undefined; }\n"},
		{"B_unconditional_shorthand", "function make(p: R): { p: W } | undefined { return { p }; }\n"},
		{"C_narrowed_named", "function make(p: R): { p: W } | undefined { if (p.t > 0) { return { p: p }; } return undefined; }\n"},
		{"D_narrowed_shorthand_no_union", "function make(p: R): { p: W } { if (p.t > 0) { return { p }; } return { p: { t: 0 } }; }\n"},
		{"E_unconditional_shorthand_no_union", "function make(p: R): { p: W } { return { p }; }\n"},
	} {
		os.Stderr.WriteString("=== " + variant.name + "\n")
		_, err := lowerSourceAs(t, head+variant.make+tail, ".ts")
		os.Stderr.WriteString("result: " + func() string { if err == nil { return "lowered" }; return err.Error() }() + "\n")
	}
}
