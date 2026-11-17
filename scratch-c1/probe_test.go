package lower

import "testing"

// Scratch probe for #snypn8c c1, never landed: probe.sh copies it into internal/lower for one run.
// Each row is a minimal program shaped as one c1 census row (taken at 03c140ec, before e6ee9eda's
// falsyArms), and logs whether main lowers it and agrees with Node or what refuses it, so one run
// says which rows main already clears. A row that disagrees with Node fails, since that is a miscompile.
func TestC1Probe(t *testing.T) {
	t.Parallel()
	const tagShapes = "enum K { A, B, C }\ninterface Expr { readonly kind: K.A; readonly type: { readonly kind: K.C } }\ninterface Tag { readonly kind: K.B; readonly expr: Expr | undefined }\nconst expr: Expr = { kind: K.A, type: { kind: K.C } };\nconst tag: Tag = { kind: K.B, expr };\n"
	const ternary = "const enum Ternary { False = 0, Unknown = 1, Maybe = 3, True = -1 }\nfunction compare(a: number, b: number): Ternary {\n\treturn a === b ? Ternary.True : a < b ? Ternary.Maybe : Ternary.False;\n}\n"
	for _, probe := range []struct{ name, source string }{
		// Pin 1's shape: utilitiesPublic.ts's `tag && tag.typeExpression`, the left's object arm a different
		// member of the enum than the result's kind slot.
		{"tag_expr", tagShapes + "function get(tag: Tag | undefined): Expr | undefined {\n\treturn tag && tag.expr;\n}\nconst none = get(undefined);\nconsole.log(none === undefined ? \"none\" : String(none.kind));\nconst some = get(tag);\nconsole.log(some === undefined ? \"none\" : String(some.kind));\n"},
		// utilitiesPublic.ts:1211 itself (getJSDocTypeTag): the && chain is an if condition, no slot at all.
		{"tag_expr_condition", tagShapes + "function typeTag(tag: Tag | undefined): Tag | undefined {\n\tif (tag && tag.expr && tag.expr.type) {\n\t\treturn tag;\n\t}\n\treturn undefined;\n}\nconsole.log(typeTag(undefined) === undefined ? \"none\" : \"tag\");\nconsole.log(typeTag(tag) === undefined ? \"none\" : \"tag\");\n"},
		// The scout's minimal shape: an optional property on the right.
		{"tag_optional_expr", "const enum K { A, B }\ninterface EB { readonly kind: K.B }\ninterface TA { readonly kind: K.A; readonly e?: EB }\nfunction f(t: TA | undefined): EB | undefined {\n\treturn t && t.e;\n}\nconsole.log(f(undefined) === undefined ? \"none\" : \"some\");\nconsole.log(f({ kind: K.A, e: { kind: K.B } }) === undefined ? \"none\" : \"some\");\n"},
		// checker.ts's `node && getSourceFileOfNode(node)` rows: a call on the right.
		{"node_call", tagShapes + "function exprOf(tag: Tag): Expr {\n\treturn tag.expr ?? expr;\n}\nfunction read(tag: Tag | undefined): Expr | undefined {\n\treturn tag && exprOf(tag);\n}\nconsole.log(read(undefined) === undefined ? \"none\" : \"expr\");\nconsole.log(read(tag) === undefined ? \"none\" : \"expr\");\n"},
		// checker.ts:21974: a boolean && left, the && typed false | Ternary, then || another Ternary.
		{"boolean_left", ternary + "function related(strict: boolean, a: number, b: number): Ternary {\n\tconst result = !strict && compare(a, b) || compare(b, a);\n\treturn result;\n}\nconsole.log(String(related(false, 1, 1)));\nconsole.log(String(related(true, 1, 2)));\nconsole.log(String(related(false, 2, 1)));\n"},
		// checker.ts:22007: two boolean lefts chained, into a let later written Ternary.False.
		{"boolean_left_chain", ternary + "const enum Mode { None = 0, Callback = 1, StrictArity = 2 }\nfunction related(mode: Mode, strict: boolean, a: number, b: number): Ternary {\n\tlet result = !(mode & Mode.Callback) && !strict && compare(a, b) || compare(b, a);\n\tif (result && mode & Mode.StrictArity) {\n\t\tresult = Ternary.False;\n\t}\n\treturn result;\n}\nconsole.log(String(related(Mode.None, false, 1, 1)));\nconsole.log(String(related(Mode.Callback, false, 1, 2)));\nconsole.log(String(related(Mode.StrictArity, true, 3, 3)));\n"},
		// checker.ts:22053: a number && left, `mode & Flag`, the && typed 0 | Ternary.
		{"number_left", ternary + "const enum Mode { None = 0, BivariantCallback = 1, StrictCallback = 2 }\nfunction returns(mode: Mode, a: number, b: number): Ternary {\n\tconst result = mode & Mode.BivariantCallback && compare(b, a) || compare(a, b);\n\treturn result;\n}\nconsole.log(String(returns(Mode.None, 1, 1)));\nconsole.log(String(returns(Mode.BivariantCallback, 2, 1)));\nconsole.log(String(returns(Mode.StrictCallback, 1, 2)));\n"},
		// checker.ts:32534: `!!(elementFlags[index]! && ElementFlags.Optional)`, upstream's && where & is meant.
		{"element_flags", "const enum ElementFlags { None = 0, Required = 1, Optional = 2, Rest = 4, Variadic = 8 }\nfunction optional(flags: readonly ElementFlags[], index: number): boolean {\n\treturn !!(flags[index]! && ElementFlags.Optional);\n}\nconsole.log(String(optional([ElementFlags.Required, ElementFlags.Optional], 0)));\nconsole.log(String(optional([ElementFlags.None], 0)));\n"},
		// checker.ts:13535: `declaration.name || declaration`, its right dead because a type guard proved name
		// an object, so tsc types the || as the name alone.
		{"or_dead_right", "const enum Kind { Identifier, Declaration }\ninterface Name { readonly kind: Kind.Identifier; readonly text: string }\ninterface Decl { readonly kind: Kind.Declaration; readonly name?: Name }\ntype NamedDecl = Decl & { readonly name: Name };\nfunction isNamed(declaration: Decl): declaration is NamedDecl {\n\treturn declaration.name !== undefined;\n}\nfunction at(declaration: Decl): Name | Decl {\n\treturn isNamed(declaration) ? declaration.name || declaration : declaration;\n}\nconsole.log(at({ kind: Kind.Declaration, name: { kind: Kind.Identifier, text: \"x\" } }).kind === Kind.Identifier ? \"name\" : \"declaration\");\nconsole.log(at({ kind: Kind.Declaration }).kind === Kind.Identifier ? \"name\" : \"declaration\");\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, probe.source, ".ts"); err != nil {
				t.Logf("c1probe: %s: refused: %v", probe.name, err)
				return
			}
			// Not parallel: t.Run must return the agreement's verdict before the row logs it.
			if t.Run("agree", func(t *testing.T) { lowersAndAgreesWithNodeAs(t, probe.source, ".ts") }) {
				t.Logf("c1probe: %s: lowers", probe.name)
			} else {
				t.Logf("c1probe: %s: lowers, disagrees with Node (see its FAIL)", probe.name)
			}
		})
	}
}
