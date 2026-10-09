package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoUnnecessaryTypeConstraint = rule.Message{
	Id: "noUnnecessaryTypeConstraint",
	Description: "A type parameter with no `extends` clause is already constrained to `unknown`, so " +
		"writing `extends any` or `extends unknown` constrains nothing. The clause reads as a " +
		"deliberate bound and communicates one, which makes it worse than noise: a reader looking " +
		"for where `T` is restricted finds this and stops. Remove it.",
}

var suggestionRemoveTheConstraint = rule.Message{
	Id:          "removeTheConstraint",
	Description: "Remove the constraint that does nothing.",
}

// NoUnnecessaryTypeConstraint flags a generic type parameter constrained to `any` or `unknown`.
//
//	valid:   function data<T>() {}
//	valid:   function data<T extends number>() {}
//	valid:   function data<T extends any | number>() {}
//	valid:   type X = any; function data<T extends X>() {}
//	invalid: function data<T extends any>() {}
//	invalid: interface Data<T extends unknown> {}
//	invalid: class Data { member<T extends unknown>() {} }
//
// Ported from `@typescript-eslint/no-unnecessary-type-constraint` by way of oxc's
// `typescript/no_unnecessary_type_constraint`.
//
// # The set of unnecessary constraints is exactly two, and it was measured rather than reasoned
//
// oxc destructures two variants and continues on everything else, so only the bare `any` and
// `unknown` keywords report. The neighbouring candidates a reader expects to behave the same way do
// not, and each was pinned against the release binary rather than argued from the source:
//
//	function data<T extends {}>() {}          silent
//	function data<T extends object>() {}      silent
//	function data<T extends never>() {}       silent
//	function data<T extends void>() {}        silent
//	function data<T extends any | number>() {}  silent, and upstream ships this as a pass case
//	type X = any; function data<T extends X>() {}  silent, also an upstream pass case
//
// The last two are the interesting ones and they say the same thing twice: the decision is made on
// the *syntax* of the constraint, not on the type it denotes. A union containing `any` collapses to
// `any` and an alias for `any` resolves to `any`, so a rule reasoning about types would report both.
// Upstream reasons about the parse tree, needs no checker to do it, and this port does the same.
//
// # Parentheses, measured in the direction that matters
//
// `function data<T extends (any)>() {}` is silent upstream, because a parenthesized type is a
// `TSParenthesizedType` there and the destructure declines it. Our parser agrees: it produces a
// `KindParenthesizedType` wrapping the keyword. So the faithful port is to *withhold* a
// parenthesis skip rather than to add one. This is written down because skipping parentheses reads
// as a free correctness improvement and adding it here would ship a silent divergence upstream's
// own corpus cannot see, since the corpus writes no parenthesized constraint at all.
var NoUnnecessaryTypeConstraint = rule.Rule{
	Name: "@typescript-eslint/no-unnecessary-type-constraint",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindTypeParameter: func(node *ast.Node) {
				typeParameter := node.AsTypeParameterDeclaration()
				if typeParameter.Constraint == nil {
					return
				}

				switch typeParameter.Constraint.Kind {
				case ast.KindAnyKeyword, ast.KindUnknownKeyword:
				default:
					return
				}

				name := typeParameter.Name()

				// The finding points at the parameter's name, which is oxc's primary label. oxc
				// carries the constraint's span as a second label; we have one range per finding,
				// so the primary is the one that survives, and it is the more useful of the two
				// because it names which parameter is at fault when several are declared.
				//
				// Reported through the node helper rather than with `name.Loc`, and that is the
				// difference between pointing at `U` and pointing at ` U`. A node's `Pos()` starts
				// at the end of the previous token, so it swallows the whitespace after the comma
				// in `<T, U extends any>`; `ReportNode*` routes through `TokenRange`, which starts
				// at the token itself. Upstream's snapshot puts this finding at column 18, which is
				// the `U`. The first parameter in a list has no leading trivia to swallow, so a
				// range built from `Pos()` looks right on every single-parameter case and only the
				// second parameter exposes it.
				ctx.ReportNodeWithSuggestions(name, messageNoUnnecessaryTypeConstraint,
					rule.Suggestion{
						Message: suggestionRemoveTheConstraint,
						Fixes: []rule.Fix{rule.ReplaceRange(
							core.NewTextRange(name.End(), typeParameter.Constraint.End()),
							trailingCommaForConstraintRemoval(ctx, node, typeParameter))},
					})
			},
		}
	},
}

// trailingCommaForConstraintRemoval answers what to write over the deleted constraint: a comma, or
// nothing.
//
// This exists because deleting the constraint can change how the file parses. In a file where `<`
// is ambiguous between a generic parameter list and a JSX element, `const data = <T extends any>()
// => {};` parses as a generic arrow only because the `extends` proves it is not JSX. Remove the
// constraint and `<T>` reads as an opening JSX tag, so the repair has to leave behind the other
// disambiguator TypeScript accepts, a trailing comma: `const data = <T,>() => {};`.
//
// All five of upstream's conditions are reproduced, and each was pinned by applying the repair with
// the release binary's `--fix-suggestions` rather than read off the source:
//
//	.ts  const data = <T extends any>() => {};              becomes <T>          no comma
//	.tsx const data = <T extends any>() => {};              becomes <T,>         comma
//	.mts const data = <T extends any>() => {};              becomes <T,>         comma
//	.cts const data = <T extends any>() => {};              becomes <T,>         comma
//	.tsx function data<T extends any>() {}                  becomes <T>          not an arrow
//	.tsx const data = <T extends any = unknown>() => {};    becomes <T = unknown> has a default
//	.tsx const data = <T extends any, U extends any>() => {}; becomes <T, U>      two parameters
//	.tsx const data = <T extends any,>() => {};             becomes <T,>         comma already there
//
// The default case is not an oversight upstream is carrying: `<T = unknown>` is already unambiguous
// because `=` cannot appear in a JSX tag, so no comma is needed.
//
// # Finding an existing comma needs no scan, because the parser already excluded the comments
//
// oxc searches the source between the parameter's end and the declaration's end for a `,`, and
// filters out any hit that falls inside a comment. Probing our own nodes first, as the shelf rule
// says to, showed the parser has already done that work: the type parameter list's range ends after
// a real trailing comma and stops before a comment, so the text between the parameter's end and the
// list's end contains a comma only when a real one is there.
//
// Pinned on the input that separates the two spellings, `const data = <T extends any /* , */>() =>
// {};`, where a naive text search finds the comma inside the comment and would suppress the one the
// repair owes. The release binary rewrites it to `const data = <T, /* , */>() => {};`, adding a
// comma, and so does this.
func trailingCommaForConstraintRemoval(
	ctx rule.Context,
	node *ast.Node,
	typeParameter *ast.TypeParameterDeclaration,
) string {
	if node.Parent == nil || node.Parent.Kind != ast.KindArrowFunction {
		return ""
	}
	if typeParameter.DefaultType != nil {
		return ""
	}

	typeParameters := node.Parent.TypeParameterList()
	if typeParameters == nil || len(typeParameters.Nodes) != 1 {
		return ""
	}
	if !fileNeedsGenericArrowDisambiguation(ctx.SourceFile.FileName().AsString()) {
		return ""
	}

	// The text between this parameter and the end of the list. A trailing comma, if one was
	// written, is the only thing that can be in here besides trivia.
	tail := ctx.SourceFile.Text()[node.End():typeParameters.End()]
	if strings.Contains(tail, ",") {
		return ""
	}
	return ","
}

// fileNeedsGenericArrowDisambiguation reports whether `<T>` at the head of an arrow function could
// be read as JSX in this file, so that removing a constraint has to leave a comma behind.
//
// oxc asks two questions here, `source_type().is_jsx()` and an extension test against `mts`, `cts`
// and `tsx`. The first is derived from the extension for a file on disk, so on the inputs a linter
// actually sees the two questions coincide and this reproduces the union of them. A `.ts` file is
// not ambiguous, which is why the same source rewrites differently depending only on its name.
func fileNeedsGenericArrowDisambiguation(fileName string) bool {
	for _, extension := range []string{".tsx", ".mts", ".cts"} {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}
	return false
}
