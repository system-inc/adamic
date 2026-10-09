package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// buildNoUndefInitMessage renders the finding, which names the binding.
//
// Built per report rather than held as a package constant, because the name is interpolated and
// `rule.Message` carries no rendering layer. That means the id assertion in a fixture cannot see
// anything this format string does, so the test asserts the rendered Description directly.
func buildNoUndefInitMessage(name string) rule.Message {
	return rule.Message{
		Id: "unnecessaryUndefinedInit",
		Description: fmt.Sprintf(
			"`%s` is initialized to `undefined`, which is the value it would already have. The "+
				"initializer adds a token and no behaviour, and it costs the reader a moment "+
				"deciding whether the author meant something by it. Drop it.", name),
	}
}

// NoUndefInit flags a variable declaration whose initializer is the literal identifier `undefined`.
//
//	valid:   var a;
//	valid:   const foo = undefined
//	valid:   var undefined = 5; var foo = undefined;
//	valid:   class C { field = undefined; }
//	invalid: var a = undefined;
//	invalid: let a = 1, b = undefined, c = 5;
//
// # Three separate decisions, and the corpus separates them
//
// **`const` and its relatives are exempt.** Upstream's `CONSTANT_BINDINGS` is `const`, `using` and
// `await using`, and the reason is that for those three the initializer is not optional: `const foo;`
// is a syntax error, so `const foo = undefined` is the only way to write it and there is nothing to
// drop. `ast.IsVarConstLike` is that same set read off the parser rather than off a keyword string,
// and it already covers `using` and `await using` because both carry the const flag.
//
// **A shadowed `undefined` suppresses the rule.** `var undefined = 5; var foo = undefined;` is clean,
// because `foo` then holds `5` and removing the initializer would change what the program does.
// Upstream asks its scope analysis whether any binding named `undefined` has a definition; we ask
// the checker whether the identifier resolves to a symbol declared in a source file. Measured on the
// installed rule: a parameter named `undefined`, a block-scoped `let undefined`, and a file-level
// `var undefined` all suppress it, and our checker answers the same three ways.
//
// **A class field is not a variable declaration.** `class C { field = undefined; }` is clean
// upstream because the listener is `VariableDeclarator` and a property declaration is not one. That
// falls out of anchoring on the same node here rather than needing an arm.
//
// # The global `undefined` has NO declarations, which is the trap
//
// The shipped shadow test in this package, `resolvesToAGlobal` in `no_new_native_nonconstructor.go`,
// asks whether the symbol's first declaration lives in a declaration file. That is right for
// `Symbol` and `BigInt` and WRONG for this name: measured with a probe over five inputs, the real
// global `undefined` resolves to a symbol with **zero declarations**, so the shipped predicate
// answers false for it and a rule built on it would be silent on every reporting case. The question
// here is therefore the complement, asked directly: does this identifier resolve to a symbol that
// something in SOURCE declares. Zero declarations means the global, which reports.
//
// # The fixer, and its four declines
//
// Upstream returns null from the fixer in four cases, and each is a decision rather than an
// omission. `var` is excluded because `var a = undefined` and `var a` differ under hoisting once the
// declaration is reached a second time in a loop, so the edit is not meaning-preserving. A
// destructuring target is excluded because `var [a] = undefined` throws and `var [a]` does not, so
// the "repair" would delete a crash. And a comment anywhere between the binding name and the end of
// the declarator is excluded, in both directions, because the removal range would swallow it.
//
// The corpus asserts all of this: nine of the twenty-one reporting cases carry `output: null`, six
// of them purely about comment placement. `let a = undefined/* comment */;` fixes to
// `let a/* comment */;` while `let a/**/ = undefined;` does not fix at all, and the difference is
// only which side of the equals sign the comment sits on.
var NoUndefInit = rule.Rule{
	Name:             "no-undef-init",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				declaration := node.AsVariableDeclaration()
				if declaration == nil || declaration.Initializer == nil {
					return
				}

				// The initializer must be the bare identifier. Upstream reads `node.init.name`,
				// which is undefined for every other expression kind, so `a = void 0` and
				// `a = window.undefined` are both clean and neither needs an arm.
				initializer := declaration.Initializer
				if initializer.Kind != ast.KindIdentifier || initializer.Text() != "undefined" {
					return
				}

				// `const`, `using` and `await using` all require an initializer, so there is nothing
				// to drop. Combined flags rather than the node's own, since the keyword lives on the
				// enclosing declaration list.
				if ast.IsVarConstLike(node) {
					return
				}

				if undefinedIsShadowed(ctx, initializer) {
					return
				}

				name := declaration.Name()
				if name == nil {
					return
				}

				message := buildNoUndefInitMessage(ctx.SourceFile.Text()[rule.TokenRange(ctx.SourceFile, name).Pos():name.End()])
				if fix, offered := noUndefInitFix(ctx, node, name); offered {
					ctx.ReportNodeWithFixes(node, message, fix)
					return
				}
				ctx.ReportNode(node, message)
			},
		}
	},
}

// undefinedIsShadowed reports whether this `undefined` names a binding written in source.
//
// A thin naming of `identifierIsShadowed` for the one caller that reads better with the name in it.
// The shared form exists because `no-alert` asks the identical question of three different names,
// and two spellings of one decision is how the four-implementation census in
// `internal/utilities/ecmascript/reference` started.
func undefinedIsShadowed(ctx rule.Context, identifier *ast.Node) bool {
	return identifierIsShadowed(ctx, identifier)
}

// identifierIsShadowed reports whether an identifier names a binding written in source rather than
// the global of that name.
//
// The test is "does anything in a non-declaration file declare this", and a name nothing declares is
// the global. That is the COMPLEMENT of `resolvesToAGlobal` in `no_new_native_nonconstructor.go`
// rather than its negation, and the difference is measured rather than stylistic: probed over five
// inputs, the real global `undefined` resolves to a symbol with ZERO declarations, so asking
// "is the first declaration in a declaration file" answers false for it and any rule built on that
// predicate is silent on every case it should report. `window` and `globalThis` do have declaration
// file declarations, so both predicates agree for those two and only this one is right for all
// three.
//
// A declaration living in a `.d.ts` is NOT a shadow. ESLint's scope analysis is per-file and cannot
// see another file's ambient declaration at all, so upstream reports whatever a neighbouring
// declaration file says; ours can see it, and fidelity is to the decision rather than to the
// mechanism. Reachable rather than theoretical: with a real `.d.ts` in the program the symbol
// carries one declaration, in a declaration file, and dropping this test makes the rule read it as a
// shadow and go silent.
//
// A symbol we cannot resolve answers false, which reports. That is the conservative direction for
// these callers: an unresolvable `alert` or `undefined` is overwhelmingly the global in a file the
// checker could not fully type, and declining there would silence the rule on every such file.
func identifierIsShadowed(ctx rule.Context, identifier *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if file := ast.GetSourceFileOfNode(declaration); file != nil && !file.IsDeclarationFile {
			return true
		}
	}
	return false
}

// noUndefInitFix builds the removal upstream would build, or declines the way upstream declines.
//
// The range runs from the end of the binding name to the end of the declarator, which takes the
// equals sign, the initializer and any whitespace between them in one edit. Upstream computes the
// same two offsets from `node.id.range[1]` and `node.range[1]`.
//
// Three declines, all of them upstream's, all of them asserted by an `output: null` in the corpus:
// a `var` declaration, a destructuring target, and a comment sitting inside the range about to be
// removed.
func noUndefInitFix(ctx rule.Context, node *ast.Node, name *ast.Node) (rule.Fix, bool) {
	// `var a = undefined` and `var a` are not the same program. A `var` re-reached in a loop keeps
	// whatever it last held when there is no initializer, and is reset to `undefined` when there is
	// one, so the edit changes behaviour rather than spelling.
	if !ast.IsVarLet(node) {
		return rule.Fix{}, false
	}

	// `var [a] = undefined` throws at runtime and `var [a]` is a syntax error. Neither direction is
	// a repair, so upstream declines and so does this.
	if name.Kind != ast.KindIdentifier {
		return rule.Fix{}, false
	}

	// The removal starts after everything the declaration keeps, not after the name. Upstream reads
	// `node.id.range[1]`, and in ESTree a TypeScript annotation is part of the id node, so that one
	// offset already sits past it. Here `Type` and `ExclamationToken` are siblings of the name, and
	// starting at `name.End()` swallows them: `let x: Thing | undefined = undefined` fixed to
	// `let x`, silently widening the declaration to `any`. It compiles, so nothing fails, and the
	// only trace is the diff. Measured on eight files in libraries/structure.
	removalStart := name.End()
	if declaration := node.AsVariableDeclaration(); declaration != nil {
		if declaration.Type != nil {
			removalStart = declaration.Type.End()
		} else if declaration.ExclamationToken != nil {
			removalStart = declaration.ExclamationToken.End()
		}
	}

	removal := core.NewTextRange(removalStart, node.End())
	for _, comment := range comments.ForFile(ctx) {
		// Upstream asks `commentsExistBetween(node.id, lastToken)`, which is this half-open range.
		// A comment after the initializer but before the terminator sits outside it and is kept,
		// which is why `let a = undefined/* comment */;` fixes and `let a /**/ = undefined;` does
		// not.
		if comment.Range.Pos() >= removal.Pos() && comment.Range.End() <= removal.End() {
			return rule.Fix{}, false
		}
	}
	return rule.RemoveRange(removal), true
}
