package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// An object view can hide a field even when its source was a nominal instance.
// Mutant run and restored through a Go overlay: restore optionalWidened's
// non-object source bail. All six refusal pins fail. The five ordinary views,
// including t2, return nil; the spread reaches the library-class spread refusal.
func TestOptionalWideningObjectRefused(t *testing.T) {
	t.Parallel()
	const types = "interface Held { a: string }\ninterface Read { a?: number }\n"
	for _, probe := range []struct {
		name, source, fixed, want string
	}{
		{
			"t2 call",
			types + "function keep(x: object): object { return x; }\nconst held: Held = { a: \"text\" };\nconst r: Read = keep(held);\nconsole.log(String((r.a ?? 0) + 1));\n",
			"interface Held { a?: number }\ninterface Read { a?: number }\nfunction keep(x: Held): Held { return x; }\nconst held: Held = { a: 1 };\nconst r: Read = keep(held);\nconsole.log(String((r.a ?? 0) + 1));\n",
			"main.a:5:17: Adamic 0.1 refuses optional property a in Read absent from structural source object, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
		},
		{
			"object local",
			types + "const held: Held = { a: \"text\" };\nconst o: object = held;\nconst r: Read = o;\nconsole.log(String((r.a ?? 0) + 1));\n",
			"interface Held { a?: number }\ninterface Read { a?: number }\nconst held: Held = { a: 1 };\nconst o: Held = held;\nconst r: Read = o;\nconsole.log(String((r.a ?? 0) + 1));\n",
			"main.a:5:17: Adamic 0.1 refuses optional property a in Read absent from structural source object, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
		},
		{
			"class through object",
			"interface Read { a?: number }\nclass C { a = \"text\" }\nconst o: object = new C();\nconst r: Read = o;\nconsole.log(String((r.a ?? 0) + 1));\n",
			"interface Read { a?: number }\nclass C { a?: number = 1 }\nconst o: C = new C();\nconst r: Read = o;\nconsole.log(String((r.a ?? 0) + 1));\n",
			"main.a:4:17: Adamic 0.1 refuses optional property a in Read absent from structural source object, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
		},
		{
			"object parameter",
			types + "function f(x: object): void {\nconst r: Read = x;\nconsole.log(String((r.a ?? 0) + 1));\n}\nf({ a: \"text\" });\n",
			"interface Held { a?: number }\ninterface Read { a?: number }\nfunction f(x: Held): void {\nconst r: Read = x;\nconsole.log(String((r.a ?? 0) + 1));\n}\nf({ a: 1 });\n",
			"main.a:4:17: Adamic 0.1 refuses optional property a in Read absent from structural source object, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
		},
		{
			"number as string",
			"interface N { a: number }\ninterface S { a?: string }\nconst held: N = { a: 1 };\nconst o: object = held;\nconst r: S = o;\nconsole.log(r.a === undefined ? \"none\" : r.a.toUpperCase());\n",
			"interface N { a?: string }\ninterface S { a?: string }\nconst held: N = { a: \"text\" };\nconst o: N = held;\nconst r: S = o;\nconsole.log(r.a === undefined ? \"none\" : r.a.toUpperCase());\n",
			"main.a:5:14: Adamic 0.1 refuses optional property a in S absent from structural source object, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
		},
		{
			"object spread",
			types + "const held: Held = { a: \"text\" };\nconst o: object = held;\nconst s: Read = { ...o };\nconsole.log(String((s.a ?? 0) + 1));\n",
			"interface Held { a?: number }\ninterface Read { a?: number }\nconst held: Held = { a: 1 };\nconst o: Held = held;\nconst s: Read = { ...o };\nconsole.log(String((s.a ?? 0) + 1));\n",
			"main.a:5:19: Adamic 0.1 refuses optional property a in Read absent from structural source object, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
		},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want a refusal ending %q", err, probe.want)
			}
			lowersAndAgreesWithNode(t, probe.fixed)
		})
	}
}

func TestOptionalWideningObjectExemptions(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"undefined spread", "interface Read { a?: number }\nfunction f(flag: boolean): void { const s: Read = { ...(flag ? { a: 1 } : undefined) }; console.log(String(s.a)); }\nf(true);\nf(false);\n"},
		{"readonly class view", "interface Read { readonly a?: number }\nclass D { a = 1 }\nconst r: Read = new D();\nconsole.log(String((r.a ?? 0) + 1));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNode(t, probe.source)
		})
	}
}

// These original controls are exempt from optional widening, but main does not
// lower them: boolean spreads reach a representation NotYet, and a writable
// optional view of D.a reaches mutable invariance. Check this rule directly so
// those independent stops cannot conceal a newly introduced optional refusal.
func TestOptionalWideningObjectExemptionsBehindOtherStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"boolean spread", "interface Read { a?: number }\nfunction f(flag: boolean): void { const s: Read = { ...(flag && { a: 1 }) }; console.log(String(s.a)); }\nf(true);\nf(false);\n", "main.a:2:53: stage 0 can't lower spreading a union of differently held members yet"},
		{"direct class", "interface Read { a?: number }\nclass D { a = 1 }\nconst r: Read = new D();\nconsole.log(String((r.a ?? 0) + 1));\n", "main.a:3:17: Adamic 0.1 refuses a value of type D seen as Read, which can write number | undefined where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			finder, _ := cycleIndexTypes(t, probe.source)
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if err := finder.l.refuseOptionalWidening(node); err != nil {
					t.Fatal(err)
				}
				return node.ForEachChild(visit)
			}
			finder.l.program.Files()[0].AsNode().ForEachChild(visit)
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.HasSuffix(err.Error(), probe.want) {
				t.Fatalf("got %v, want the existing stop ending %q", err, probe.want)
			}
		})
	}
}
