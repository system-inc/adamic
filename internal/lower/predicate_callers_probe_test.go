package lower

import "testing"

// Scratch probe, never committed: predicateCallbackCallers (predicates.go) through a shorthand name.
func TestPredicateCallersProbe(t *testing.T) {
	head := "function apply(v: string | number, isS: (x: string | number) => x is string): string {\n    return isS(v) ? `len ${v.length}` : \"not\";\n}\n"
	for _, row := range []struct{ name, body string }{
		{"lie_shorthand", "const box = { apply };\nconsole.log(box.apply(1, (x: string | number): x is string => true));\n"},
		{"lie_plain", "const box = { apply: apply };\nconsole.log(box.apply(1, (x: string | number): x is string => true));\n"},
		{"view_shorthand", "const box: { apply(v: string | number, isS: (x: string | number) => boolean): string } = { apply };\nconsole.log(box.apply(1, (x: string | number) => true));\n"},
		{"view_plain", "const box: { apply(v: string | number, isS: (x: string | number) => boolean): string } = { apply: apply };\nconsole.log(box.apply(1, (x: string | number) => true));\n"},
		{"held_shorthand", "const box = { apply };\nconsole.log(box.apply(1, (x: string | number): x is string => typeof x === \"string\"));\n"},
		{"direct_control", "console.log(apply(1, (x: string | number): x is string => typeof x === \"string\"));\n"},
		{"callback_escape_shorthand", "function run(v: string | number, isS: (x: string | number) => x is string): string {\n    const h = { isS };\n    return h.isS(v) ? `len ${v.length}` : \"not\";\n}\nconsole.log(run(1, (x: string | number): x is string => typeof x === \"string\"));\n"},
	} {
		program, err := lowerSourceAs(t, head+row.body, ".ts")
		if err != nil {
			t.Logf("ROW %s stops: %v", row.name, err)
			continue
		}
		code, stdout, stderr := brandedNative(t, program)
		t.Logf("ROW %s LOWERS: native exit %d stdout %q stderr %.200q", row.name, code, stdout, stderr)
	}
}
