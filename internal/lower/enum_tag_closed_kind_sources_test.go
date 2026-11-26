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
// Lower's slice-4 roads (J-1f4mec2) are sources too, each lie beside its twin (Compiler's must on 79afa0cf): a
// value built at an inline Node & {...}, at a function's literal return type, or at a parameter's literal
// type, with no name, keeps the refusal, and its twin with a name lowers. A literal whose kind widens to the
// open enum holds every kind, Ident's among them, and its K-typed kind doesn't relate to Ident's, so both
// twins keep the refusal (the relation mutant lowers both).
// A declared source whose field the predicate's type types otherwise, NumIdent's name a number, keeps the
// refusal: castAssignable refuses it, as it refuses a wider, optional or unknown field, a method's parameter or
// result, an array's element, a nested field and a missing method (J-q288d38, each with and without the
// checker's assignability asked first, so that step was left out).
// A declared union is judged member by member: Ident | NumLit lowers, since each member is its own source, and
// a union hiding an anonymous member at Ident's kind with no name keeps the refusal.
// Mutants: closedKindSourcesFit's relation dropped (every source fits) lowers every lie, and the anonymous
// literal's reads name off a value that has none; a union judged whole refuses the declared union of sources;
// a union judged by its first member alone lowers the anonymous member's union.
func TestClosedKindSourcesNarrowByThePredicate(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		{"honest", "", "other\nx\n"},
		{"anonymous literal", "const raw = { kind: K.Ident as const, pos: 9, name: 'y', _expressionBrand: undefined };\nconst anon: Ex = raw;\nconsole.log(show(anon));\n", "other\nx\ny\n"},
		// anonymous literal, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/anonymous_literal,_no_name".
		{"class", "class OwnIdent { readonly kind = K.Ident as const; readonly pos = 0; readonly name = 'c'; readonly _expressionBrand = undefined; }\nconst made: Ex = new OwnIdent();\nconsole.log(show(made));\n", "other\nx\nc\n"},
		// class, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/class,_no_name".
		{"literal cast as Ex", "const cast = { kind: K.Ident as const, pos: 1, name: 'z', _expressionBrand: undefined } as Ex;\nconsole.log(show(cast));\n", "other\nx\nz\n"},
		// literal cast as Ex, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/literal_cast_as_Ex,_no_name".
		{"declared union of sources", "type IdentOrNum = Ident | NumLit;\n", "other\nx\n"},
		// declared union with an anonymous member, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/declared_union_with_an_anonymous_member,_no_name".
		{"inline intersection", "const inter: Node & { readonly kind: K.Ident; readonly pos: number; readonly _expressionBrand: undefined; readonly name: string } = { kind: K.Ident, pos: 3, _expressionBrand: undefined, name: 'i' };\nconst viaInter: Ex = inter;\nconsole.log(show(viaInter));\n", "other\nx\ni\n"},
		// inline intersection, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/inline_intersection,_no_name".
		{"return type", "function make(): { readonly kind: K.Ident; readonly pos: number; readonly _expressionBrand: undefined; readonly name: string } { return { kind: K.Ident, pos: 4, _expressionBrand: undefined, name: 'r' }; }\nconst made: Ex = make();\nconsole.log(show(made));\n", "other\nx\nr\n"},
		// return type, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/return_type,_no_name".
		{"parameter type", "function take(odd: { readonly kind: K.Ident; readonly pos: number; readonly _expressionBrand: undefined; readonly name: string }): string { const viaParam: Ex = odd; return show(viaParam); }\nconsole.log(take({ kind: K.Ident, pos: 5, _expressionBrand: undefined, name: 'p' }));\n", "other\nx\np\n"},
		// parameter type, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/parameter_type,_no_name".
		// widened kind: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/widened_kind".
		// widened kind, no name: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/widened_kind,_no_name".
		// declared source with another field type: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestClosedKindSourcesNarrowByThePredicate/declared_source_with_another_field_type".
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := closedKindTypes + closedKindSourcesNarrowing + probe.source
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
// The program names WideIdent only as a parameter's type, never in an alias of its own, since an alias in a
// program file is a declared source too and hid the declaration file's (J-8gfsn48).
// Mutant: the declared sources read from the program's own files only (closedKindDeclaredTypes), never its
// declaration files, fits the first row, so it takes no swing site and goes red.
func TestClosedKindSourcesReadTypeOnlyImports(t *testing.T) {
	t.Parallel()
	const exports = "export type { K, Ex, Ident };\nimport type { WideIdent } from './lib';\nfunction wideOther(w: WideIdent): number { return w.other; }\n"
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
			// Without a name the sources don't fit, so the walk can't prove the refinement and the swing reads
			// it through the checked view (enumTagSwingSites); with one they fit, and cause A narrows it.
			if swung := enumTagSwungAt(main); probe.lowers == (len(swung) > 0) {
				t.Fatalf("want the swing at a site %v, got sites %v and %v", !probe.lowers, swung, err)
			}
			if !probe.lowers {
				if err != nil {
					t.Fatalf("want it lowered through the swing, got %v", err)
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

// Language's u02 (#x2cjvy9, J-bdt983a): a nested declared union (YA | XB) | XA whose YA is kind A with a
// number where XA has a string. tsc narrows isA(s) to XA and drops YA, which is tsc's lie. Cause A never judges
// it: with the probe on, u02 lowers with no fit row and no enum-tag row (J-ng3x53g), where a judged YA would have
// failed castAssignable as NumIdent does. The checked predicate admits it, as on main 893acd78, and the read it
// checks stops loudly where Node prints 3number. Pinned by its words, so a change to either road shows here.
func TestClosedKindSourcesNestedUnionLieStopsAtTheRead(t *testing.T) {
	t.Parallel()
	const source = "enum Kind {\n  A,\n  B,\n  C,\n}\ninterface Base {\n  readonly kind: Kind;\n}\ninterface XA extends Base {\n  readonly kind: Kind.A;\n  readonly a: string;\n}\ninterface XB extends Base {\n  readonly kind: Kind.B;\n  readonly b: number;\n}\nfunction isA(node: Base): node is XA {\n  return node.kind === Kind.A;\n}\nfunction guard(run: () => string): string {\n  try {\n    return run();\n  } catch (error) {\n    return \"threw\";\n  }\n}\ninterface YA extends Base {\n  readonly kind: Kind.A;\n  readonly a: number;\n}\ntype Inner = YA | XB;\ntype Outer = Inner | XA;\nfunction show(s: Outer): string {\n  return isA(s) ? guard(() => String(s.a).toUpperCase() + typeof s.a) : \"b\";\n}\nconst ya: YA = { kind: Kind.A, a: 3 };\nconsole.log(show(ya));\n"
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("want it lowered to the checked read, got %v", err)
	}
	if code, stdout, stderr := brandedNative(t, program); code != 70 || !strings.Contains(stderr, "s.a is not a string") {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 70 and %q", code, stdout, stderr, "s.a is not a string")
	}
}
