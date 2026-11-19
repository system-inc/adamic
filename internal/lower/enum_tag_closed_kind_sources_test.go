package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// closedKindSourcesNarrowing is an exact kind predicate over Body | Ex whose type, Ident, drops Ex where
// its body's comparison keeps it, since Ident's kind is in K(Ex): tsc's isPropertyAccessExpression over
// Expression | JsxTagNameExpression, small. Narrowing Ex by Ident's kind is Ident's only when every source
// of that kind in K(Ex) relates to Ident (closedKindSourcesFit, Adamic's ruling on #x2cjvy9 cause A).
const closedKindSourcesNarrowing = "function isIdent(n: Node): n is Ident { return n.kind === K.Ident; }\n" +
	"function show(x: Body | Ex): string { return isIdent(x) ? x.name : 'other'; }\n" +
	"console.log(show(body));\nconsole.log(show(ident));\n"

// Every source of Ident's kind in K(Ex) is Ident or a value built at a type that relates to it, so isIdent's
// narrowing over Body | Ex lowers and agrees with Node, on the JavaScript backend and sanitized natively.
// Beside each honest twin, its lie: a source that doesn't relate to Ident fails the call closed and keeps the
// open-enum refusal (Compiler's fourth condition). The lies: an anonymous literal typed Ex at Ident's kind
// with no name (Ex's layout), a class implementing Ex at Ident's kind with no name, and a literal with no
// name cast `as Ex`. Their twins carry a name and lower.
// Mutant: closedKindSourcesFit's relation dropped (every source fits) lowers every lie, and the anonymous
// literal's reads name off a value that has none.
func TestClosedKindSourcesNarrowByThePredicate(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		{"honest", "", "other\nx\n"},
		{"anonymous literal", "const raw = { kind: K.Ident as const, pos: 9, name: 'y', _expressionBrand: undefined };\nconst anon: Ex = raw;\nconsole.log(show(anon));\n", "other\nx\ny\n"},
		{"anonymous literal, no name", "const raw = { kind: K.Ident as const, pos: 9, _expressionBrand: undefined };\nconst anon: Ex = raw;\nconsole.log(show(anon));\n", ""},
		{"class", "class OwnIdent { readonly kind = K.Ident as const; readonly pos = 0; readonly name = 'c'; readonly _expressionBrand = undefined; }\nconst made: Ex = new OwnIdent();\nconsole.log(show(made));\n", "other\nx\nc\n"},
		{"class, no name", "class OwnIdent { readonly kind = K.Ident as const; readonly pos = 0; readonly _expressionBrand = undefined; }\nconst made: Ex = new OwnIdent();\nconsole.log(show(made));\n", ""},
		{"literal cast as Ex", "const cast = { kind: K.Ident as const, pos: 1, name: 'z', _expressionBrand: undefined } as Ex;\nconsole.log(show(cast));\n", "other\nx\nz\n"},
		{"literal cast as Ex, no name", "const cast = { kind: K.Ident as const, pos: 1, _expressionBrand: undefined } as Ex;\nconsole.log(show(cast));\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := closedKindTypes + closedKindSourcesNarrowing + probe.source
			if probe.stdout == "" {
				if _, err := lowerSourceAs(t, source, ".ts"); err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
					t.Fatalf("want %q, got %v", kindPredicateRefusal, err)
				}
				return
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// A source only a declaration file declares, imported by type alone, is a source too (Compiler's third
// condition): lib.d.ts declares WideIdent, an Ex at Ident's kind. Without a name it doesn't relate to Ident,
// and the narrowing keeps the refusal though nothing builds a WideIdent; with one it relates, and the program
// lowers and prints what Node prints, natively under the sanitizers. A type-only import of a .ts file needs
// no row: that file is one of the program's own, which K(S) reads already.
// Mutant: the declared sources read from the program's own files only (closedKindDeclaredTypes), never its
// declaration files, lowers the first row.
func TestClosedKindSourcesReadTypeOnlyImports(t *testing.T) {
	t.Parallel()
	const exports = "export type { K, Ex, Ident };\nimport type { WideIdent } from './lib';\ntype Wide = WideIdent;\n"
	for _, probe := range []struct {
		name, lib string
		lowers    bool
	}{
		{"no name", "import type { K, Ex } from './main';\nexport interface WideIdent extends Ex { readonly kind: K.Ident; readonly other: number }\n", false},
		{"named", "import type { K, Ex } from './main';\nexport interface WideIdent extends Ex { readonly kind: K.Ident; readonly name: string; readonly other: number }\n", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			main := writeAgreementSource(t, "main.ts", closedKindTypes+exports+closedKindSourcesNarrowing)
			if err := os.WriteFile(filepath.Join(filepath.Dir(main), "lib.d.ts"), []byte(probe.lib), 0o644); err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{main})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			program, err := Lower(context.Background(), loaded)
			if !probe.lowers {
				if err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
					t.Fatalf("want %q, got %v", kindPredicateRefusal, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "other\nx\n" {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "other\nx\n")
			}
		})
	}
}

// A value fact never proves a representation fact (Adamic's condition 5, Lower's Nov 17 rule): a LongIdent,
// an Ident with a field ahead of name in its literal, narrowed to Ident and read natively under ASan and
// UBSan, reads name off its own storage, beside its twin built at Ident itself.
func TestClosedKindSourcesReadEachSourceStorage(t *testing.T) {
	t.Parallel()
	source := closedKindTypes + "interface LongIdent extends Ident { readonly extra: number }\n" +
		"const long: LongIdent = { kind: K.Ident, pos: 2, extra: 7, name: 'w', _expressionBrand: undefined };\n" +
		closedKindSourcesNarrowing + "console.log(show(long));\nconsole.log(String(long.extra));\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "other\nx\nw\n7\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "other\nx\nw\n7\n")
	}
}
