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
		{"length", "const r: { length?: number } = \"abc\";\nconsole.log(String(r.length));\n", "main.a:1:32: stage 0 can't lower a string seen as the object type { length?: number; } (read it as string) yet", "const r: string = \"abc\";\nconsole.log(String(r.length));\n"},
		{"number", "const r: { toFixed?: (d: number) => string } = 1;\nconsole.log(String(r.toFixed));\n", "main.a:1:48: stage 0 can't lower a number seen as the object type { toFixed?: (d: number) => string; } (read it as number) yet", "const r: number = 1;\nconsole.log(r.toFixed(0));\n"},
		{"parameter", "function f(x: { length?: number }): number { return x.length ?? 0; }\nconsole.log(String(f(\"abc\")));\n", "main.a:2:22: stage 0 can't lower a string seen as the object type { length?: number; } (read it as string) yet", "function f(x: string): number { return x.length; }\nconsole.log(String(f(\"abc\")));\n"},
		{"carried falsy", "function f(flag: number): void {\nconst obj: { valueOf?: () => number } = { valueOf: () => 1 };\nconst v: { valueOf?: () => number } = flag && obj;\nconsole.log('ok');\n}\nf(0);\nf(1);\n", "main.a:3:39: stage 0 can't lower a number seen as the object type { valueOf?: () => number; } (read it as number) yet", "function f(flag: number): void { const v: number = flag; console.log(String(v)); }\nf(0);\nf(1);\n"},
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

// Mutant run and restored: omit the operand exclusions. The two || controls
// and truthy && control stop on discarded numeric members; the nullish control stays green.
func TestPrimitiveObjectViewLogicalValues(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"declaration", "interface Declaration { kind: number }\nfunction pick(flag: number, value: Declaration): Declaration { return (flag && value) || value; }\nconsole.log(String(pick(0, { kind: 1 }).kind));\nconsole.log(String(pick(1, { kind: 2 }).kind));\n"},
		{"length", "function f(flag: number): void {\nconst obj: { length?: number } = { length: 2 };\nconst v: { length?: number } = (flag && obj) || obj;\nconsole.log(String(v.length));\n}\nf(0);\nf(1);\n"},
		{"truthy and", "function f(flag: 1): void {\nconst obj: { length?: number } = { length: 2 };\nconst v: { length?: number } = flag && obj;\nconsole.log(String(v.length));\n}\nf(1);\n"},
		{"nullish", "function f(flag: number): void {\nconst obj: { length?: number } = { length: 2 };\nconst v: { length?: number } = ((flag && obj) || undefined) ?? obj;\nconsole.log(String(v.length));\n}\nf(0);\nf(1);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) { t.Parallel(); lowersAndAgreesWithNode(t, probe.source) })
	}
}

func TestPrimitiveObjectViewCarriedString(t *testing.T) {
	t.Parallel()
	const source = "function f(flag: boolean): void {\nconst v: { length?: number } | undefined = flag ? \"abc\" : undefined;\nconsole.log(String(v?.length));\n}\nf(true);\nf(false);\n"
	_, err := lowerSource(t, source)
	var gap *NotYet
	want := "main.a:2:44: stage 0 can't lower a string seen as the object type { length?: number; } (read it as string) yet"
	if !errors.As(err, &gap) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want NotYet ending %q", err, want)
	}
	lowersAndAgreesWithNode(t, "function f(flag: boolean): void {\nconst v: string | undefined = flag ? \"abc\" : undefined;\nconsole.log(String(v?.length));\n}\nf(true);\nf(false);\n")
}

// At 51e6b358 both calls lower and print 3 on Node and native. Binding the
// primitive to the constraint as an object view would falsely stop these calls.
// Mutant run and restored: judge primitive generic bindings against constraints;
// both acceptance rows go red with the primitive object-view NotYet.
func TestPrimitiveObjectViewGenericPrimitiveCalls(t *testing.T) {
	t.Parallel()
	for _, call := range []string{"len(\"abc\")", "len<string>(\"abc\")"} {
		t.Run(call, func(t *testing.T) {
			t.Parallel()
			source := "function len<T extends { length?: number }>(x: T): string { return String(x.length); }\nconsole.log(" + call + ");\n"
			lowersAndAgreesWithNode(t, source)
		})
	}
}

func TestPrimitiveObjectViewGenericNumberMemberStop(t *testing.T) {
	t.Parallel()
	const source = "function fix<T extends { toFixed?: (d: number) => string }>(x: T): string { return String(x.toFixed); }\nfix(1);\n"
	_, err := lowerSource(t, source)
	var gap *NotYet
	want := "main.a:1:91: stage 0 can't lower .toFixed on a number yet"
	if !errors.As(err, &gap) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want NotYet ending %q", err, want)
	}
	lowersAndAgreesWithNode(t, "function fix(x: number): string { return x.toFixed(0); }\nconsole.log(fix(1));\n")
}

// Mutant run and restored: remove the class exemption. P then gets the weaker
// primitive representation NotYet instead of the nominal ancestry refusal.
func TestPrimitiveObjectViewNominalClassStop(t *testing.T) {
	t.Parallel()
	const source = "class P { length?: number };\nconst p: P = \"abc\";\nconsole.log(String(p.length));\n"
	_, err := lowerSource(t, source)
	var refused *Refused
	want := "main.a:2:14: Adamic 0.1 refuses a value without nominal ancestry seen as P; construct that class or a subclass; use an interface for structural values (adamic/nominal-class)"
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want Refused ending %q", err, want)
	}
	lowersAndAgreesWithNode(t, "class P { length?: number = 3 };\nconst p: P = new P();\nconsole.log(String(p.length));\n")
}
