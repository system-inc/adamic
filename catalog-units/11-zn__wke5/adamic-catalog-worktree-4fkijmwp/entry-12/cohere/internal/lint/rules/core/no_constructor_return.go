package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoConstructorReturn flags a `return` carrying a value out of a class constructor.
//
//	valid:   class C { constructor() { return } }
//	valid:   class C { constructor(a) { if (!a) { return } else { a() } } }
//	valid:   class C { method() { return '' } }
//	valid:   class C { get value() { return '' } }
//	valid:   class C { constructor() { function fn() { return true } } }
//	valid:   class C { constructor() { this.fn = () => { return true } } }
//	valid:   ({ constructor() { return 1 } })
//	valid:   class C { static constructor() { return 1 } }
//	invalid: class C { constructor() { return '' } }
//	invalid: class C { constructor(a) { if (!a) { return '' } else { a() } } }
//
// A constructor's return value is almost always dead. `new C()` evaluates to the newly constructed
// object no matter what the constructor hands back, unless the returned value happens to be an
// object, in which case the language quietly substitutes it for the instance and every field the
// constructor just assigned goes with it. Both outcomes are worth a line: the common one is a value
// nothing can ever read, and the rare one is a class whose `new` silently produces something other
// than an instance of itself, which no signature and nothing at runtime will say out loud.
//
// A bare `return` is legal and common, because it is control flow rather than a value. That is the
// discrimination the whole rule turns on, and it is why the check is on the argument rather than on
// the statement, exactly as in the sibling `no-setter-return`.
//
// # Why this is syntactic
//
// Upstream tracks code paths, pushing every code path root onto a stack and reading the innermost
// one at each `return`. That machinery exists because ESLint has no cheap way to ask a node for its
// enclosing function; the question it answers is purely syntactic. The stack top is the innermost
// function-like ancestor and nothing else, so an ancestry walk that stops at the first function-like
// node computes the same value without the bookkeeping. Nothing here resolves a name, so nothing
// here needs the checker.
//
// The walk stops at the first function-like ancestor because a nested function is its own return
// target. `class C { constructor() { function fn() { return true } } }` returns from `fn`, whose
// value is used, and three of upstream's clean cases are exactly that shape, one each for a function
// declaration, a function expression and an arrow. Reading only the immediate parent would be wrong
// the other way: a `return` inside an `if` inside a `try` inside a constructor is still the
// constructor's, which is why this walks rather than reading one parent.
//
// # Where our parser and upstream's disagree, and which way each was resolved
//
// Upstream keys on the ESTree shape `MethodDefinition` with `kind === "constructor"`. Our parser
// produces a single `KindConstructor` node for the same source, so `ast.IsConstructorDeclaration` is
// the direct translation. It is not a complete one, and the differences were measured by driving the
// installed ESLint 10.8.1 rule through the Linter API rather than reasoned about.
//
// `class C { static constructor() { return 1 } }` is the one that needs a guard. A `static`
// constructor is not a constructor at all, it is a static method that happens to be named
// `constructor`, and upstream is silent on it under both the default parser and
// `@typescript-eslint/parser`. Our parser hands it back as `KindConstructor` carrying a static
// modifier, so without the modifier test below this rule reports a false positive on it. Nothing in
// upstream's corpus covers this, so no imported fixture can see it; the fixture pinning it is one
// this port added from the measurement.
//
// `class C { ['constructor']() { return 1 } }` and `class C { *constructor() { return 1 } }` need no
// guard, because our parser already declines to call either one a constructor and hands back a
// `KindMethodDeclaration`. Upstream is silent on both. That agreement is free rather than earned,
// and it is pinned by fixtures so a later parser change cannot take it away quietly.
//
// `class C { "constructor"() { return 1 } }` reports under both, and again our parser agrees for
// free: a string-literal key spelled `constructor` produces `KindConstructor`.
//
// `class C { private constructor() { return 1 } }` reports. It is TypeScript that upstream's default
// parser cannot read at all, but `@typescript-eslint/parser` parses it and the rule reports, which
// is the authority that matters for a TypeScript file here.
//
// `class C { static { return 1 } }` is silent, and needs no arm. A class static block is not
// function-like in our parser, so the walk reaches the source file and answers nil. Upstream's
// default parser rejects the source outright and `@typescript-eslint/parser` parses it and reports
// nothing, so all three agree by three different routes.
//
// # The one place this reports where upstream renders no verdict
//
// `class C { async constructor() { return 1 } }` is a syntax error: both ESLint's default parser and
// `@typescript-eslint/parser` refuse the source, so upstream never reaches the rule and has no
// opinion to reproduce. Our parser recovers from it and hands back a `KindConstructor` with an async
// modifier and a real body. This rule reports it, deliberately. There is no upstream behaviour to
// diverge from, the node is a constructor returning a value in every sense this rule cares about,
// and the illegal modifier is already the compiler's finding rather than something silence here
// would help with. The alternative, guarding on the async modifier, would be inventing a judgment
// upstream never made in order to imitate a parser failure, which is not the same thing as fidelity.
var NoConstructorReturn = rule.Rule{
	Name: "no-constructor-return",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				statement := node.AsReturnStatement()
				// A bare `return` is control flow rather than a value, and every constructor in
				// upstream's clean list that returns at all returns bare.
				if statement == nil || statement.Expression == nil {
					return
				}

				if enclosingConstructor(node) == nil {
					return
				}

				ctx.ReportNode(node, rule.Message{
					Id: "noConstructorReturn",
					Description: "This returns a value from a constructor, where it is either " +
						"discarded or worse. `new C()` evaluates to the new instance and throws " +
						"the returned value away, unless that value is an object, in which case " +
						"the language substitutes it for the instance and every field this " +
						"constructor assigned is lost. Neither outcome is one a reader expects " +
						"from a `new`, so return nothing and assign to `this` instead.",
				})
			},
		}
	},
}

// enclosingConstructor returns the constructor this node returns from, or nil if it returns from
// anything else.
//
// The first function-like ancestor wins, because that is what a `return` binds to. Reaching a
// constructor first means the value is dead or is silently replacing the instance; reaching any
// other function first means the `return` belongs to that function and is somebody else's business,
// even when a constructor encloses it further up. This is what upstream's code path stack computes,
// without the stack.
//
// A `return` at the top level of a file reaches neither and answers nil, which is upstream's
// behaviour on the one clean case that returns outside any function at all.
//
// # Two mutants survive here and both are equivalent, with the measurements recorded
//
// Starting the walk at `node` rather than at `node.Parent` survives every fixture. It is
// equivalent: the extra first step tests the ReturnStatement itself, and a node has exactly one
// kind, so a `KindReturnStatement` can never also satisfy `IsConstructorDeclaration` or
// `IsFunctionLike`. There is no input that distinguishes them. `node.Parent` is kept because it
// says what the walk means, which is "ask the ancestors".
//
// Narrowing `ast.IsFunctionLike` to `ast.IsFunctionLikeDeclaration` in the quit arm also survives
// every fixture, and that one was resolved by probe rather than by argument. The two predicates
// differ only on the signature kinds, `KindMethodSignature`, `KindCallSignature`,
// `KindConstructSignature`, `KindIndexSignature`, `KindFunctionType` and `KindConstructorType`, and
// a signature has no body, so no statement can be nested inside one. A throwaway package walked
// every ancestor of every ReturnStatement across nine shapes written to put a signature in the way,
// including `type G = { new (): void }` and `interface J { (): void }` inside a constructor body,
// and no return ever reached a signature ancestor before reaching a function-like declaration. The
// broader predicate is kept because the shipped sibling `no-setter-return` uses it for the same
// walk, and matching it is worth more than a narrowing that nothing can observe.
func enclosingConstructor(node *ast.Node) *ast.Node {
	return ast.FindAncestorOrQuit(node.Parent, func(ancestor *ast.Node) ast.FindAncestorResult {
		if ast.IsConstructorDeclaration(ancestor) {
			// A `static` constructor is a static method named `constructor` rather than a
			// constructor, and upstream is silent on it. Our parser calls it `KindConstructor`
			// anyway, so this is the arm that keeps the port from reporting a false positive
			// nothing in the imported corpus could catch. Measured against ESLint 10.8.1 under
			// both its default parser and `@typescript-eslint/parser`: silent under both.
			//
			// `ast.IsStatic` is the wrong question here, because it also answers true for a class
			// static block, which is a different node this walk never reaches anyway. The modifier
			// test is the narrow one and it says what is actually being asked.
			if ast.HasSyntacticModifier(ancestor, ast.ModifierFlagsStatic) {
				return ast.FindAncestorQuit
			}
			return ast.FindAncestorTrue
		}
		// Every other function-like construct is its own return target. This arm is what keeps a
		// nested function, arrow, method, getter or setter inside a constructor from reporting.
		if ast.IsFunctionLike(ancestor) {
			return ast.FindAncestorQuit
		}
		return ast.FindAncestorFalse
	})
}
