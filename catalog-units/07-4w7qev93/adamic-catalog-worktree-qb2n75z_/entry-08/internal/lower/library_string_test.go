package lower

import (
	"strings"
	"testing"
)

// The explicit receiver exemption must not admit detached methods, arbitrary object conversion,
// or a mutable wider view outside the intrinsic that only reads it.
func TestLibraryStringRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"detached", "const trim = String.prototype.trim;\nconsole.log(trim());\n", "method read as a value"},
		{"null receiver", "String.prototype.trim.call(null);\n", "null or undefined"},
		{"undefined receiver", "String.prototype.trim.call(undefined);\n", "null or undefined"},
		{"String internal slot", "String.prototype.valueOf.call(42);\n", "requires a String internal slot"},
		{"object conversion", "console.log(String({ value: 1 }));\n", "ToPrimitive is not lowered"},
		{"collation", "console.log(`${'Z'.localeCompare('a')}`);\n", "locale collation"},
		{"loose inequality", "function different(left: string, right: string): boolean { return left != right; }\nconsole.log(`${different('a', 'b')}`);\n", "refuses !="},
		{"first class constructor", "const convert = String;\nconsole.log(typeof convert);\n", "outside a const alias"},
		{"raw wider view", "const template = { raw: ['a'] };\nfunction change(value: { raw: ArrayLike<string> | readonly string[] }): void { value.raw = { length: 0 }; }\nchange(template);\n", "which can write"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want refusal containing %q, got %v", probe.reason, err)
			}
		})
	}
}
