package lower

import (
	"errors"
	"strings"
	"testing"
)

// Mutant run and restored: dropping the registration admits the length probe;
// native exits 70 with "adamic: panic: dynamic property read on an unsupported runtime value".
func TestPrimitiveObjectViewNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want, fixed string }{
		{"length", "const r: { length?: number } = \"abc\";\nconsole.log(String(r.length));\n", "main.a:1:32: stage 0 can't lower a string seen as the object type { length?: number; } isn't lowered yet; read it as string yet", "const r: string = \"abc\";\nconsole.log(String(r.length));\n"},
		{"number", "const r: { toFixed?: (d: number) => string } = 1;\nconsole.log(String(r.toFixed));\n", "main.a:1:48: stage 0 can't lower a number seen as the object type { toFixed?: (d: number) => string; } isn't lowered yet; read it as number yet", "const r: number = 1;\nconsole.log(r.toFixed(0));\n"},
		{"parameter", "function f(x: { length?: number }): number { return x.length ?? 0; }\nconsole.log(String(f(\"abc\")));\n", "main.a:2:22: stage 0 can't lower a string seen as the object type { length?: number; } isn't lowered yet; read it as string yet", "function f(x: string): number { return x.length; }\nconsole.log(String(f(\"abc\")));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Fatalf("got %v, want NotYet ending %q", err, probe.want)
			}
			lowersAndAgreesWithNode(t, probe.fixed)
		})
	}
}

func TestPrimitiveObjectViewExemptions(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"string brand", "type Name = string & { readonly _name?: undefined };\nconst name = 'abc' as Name;\nconsole.log(name.toUpperCase());\n"},
		{"never brand", "type Name = string & { readonly _name?: never };\nconst name = 'abc' as Name;\nconsole.log(name.toUpperCase());\nconsole.log(String(name._name));\n"},
		{"number brand", "type Count = number & { marker?: undefined };\nconsole.log(String((3 as Count)['marker']));\n"},
		{"for of", "for (const c of 'ab') console.log(c);\n"},
		{"primitive union", "const r: string | { length?: number } = 'abc';\nconsole.log(String(r.length));\n"},
		{"empty object", "const e: {} = 'abc';\nconsole.log('ok');\n"},
	} {
		t.Run(probe.name, func(t *testing.T) { t.Parallel(); lowersAndAgreesWithNode(t, probe.source) })
	}
}

// Array.from without a callback already stops in from.go, independently of this
// rule. Its library parameter remains exempt and preserves that existing gap.
func TestPrimitiveObjectViewArrayFromExistingGap(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "console.log(Array.from(\"abc\").join(\",\"));\n")
	var gap *NotYet
	want := "main.a:1:13: stage 0 can't lower Array.from with other than { length } and a callback yet"
	if !errors.As(err, &gap) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want NotYet ending %q", err, want)
	}
}
