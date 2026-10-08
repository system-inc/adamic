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
		{"box", "const box = new String('x');", "String internal slots"},
		{"hidden primitive", "function f(object: { readonly marker: number }): string { return String(object); }", "hidden by the object view"},
		{"mixed collection", "function f(value: Map<string, number> | Set<number>): string { return String(value); }", "mixed Map and Set"},
		{"detached", "const trim = String.prototype.trim;\nconsole.log(trim());\n", "method read as a value"},
		{"String internal slot", "String.prototype.valueOf.call(42);\n", "requires a String internal slot"},
		{"object conversion", "console.log(String(() => 1));\n", "ToPrimitive is not lowered"},
		{"collation", "console.log(`${'Z'.localeCompare('a')}`);\n", "locale collation"},
		{"loose inequality", "function different(left: string, right: string): boolean { return left != right; }\nconsole.log(`${different('a', 'b')}`);\n", "refuses !="},
		{"first class constructor", "const convert = String;\nconsole.log(convert(42));\n", "String as a value outside equality or typeof"},
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

func TestLibraryStringRangeErrorsLower(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"try { String.fromCodePoint(-1); } catch {}",
		"try { String.fromCodePoint(0x110000); } catch {}",
		"try { String.fromCodePoint(0.5); } catch {}",
		"try { String.fromCodePoint(NaN); } catch {}",
		"try { String.fromCodePoint(Infinity); } catch {}",
		"function f(): string { return String.fromCodePoint(-1); } try { f(); } catch {}",
		"const f = () => String.fromCodePoint(-1); try { f(); } catch {}",
		"const codes = [65]; try { String.fromCodePoint(...codes); } catch {}",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("want catchable RangeError lowering for %q, got %v", source, err)
		}
	}
	if _, err := lowerSource(t, "try { String.fromCodePoint(0, 0x10ffff); } catch {}"); err != nil {
		t.Fatal(err)
	}
}
