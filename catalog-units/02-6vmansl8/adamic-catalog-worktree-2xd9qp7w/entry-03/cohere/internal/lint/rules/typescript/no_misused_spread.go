package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoMisusedSpreadOptions is the rule's option surface, matching upstream's schema.
//
// One key, `allow`, carrying the same heterogeneous array of strings and specifier objects that
// prefer-promise-reject-errors takes, which is why the decoder is hand-written and shares that
// rule's wire type. An empty allow list is the default and is also the correct zero value, so a rule
// handed nil options behaves correctly without any fallback.
type NoMisusedSpreadOptions struct {
	Allow       []type_checking.TypeOrValueSpecifier
	AllowInline []string
}

// NoMisusedSpread flags a spread whose argument is a value that spreading silently mishandles.
//
//	valid:   const a = [...[1, 2]]
//	valid:   const a = { ...{ x: 1 } }
//	valid:   declare const m: Map<string, number>; const a = Object.fromEntries(m)
//	valid:   const a = [...'string']  (upstream reports this; cohere does not, see below)
//	invalid: declare const m: Map<string, number>; const a = { ...m }
//	invalid: declare const p: Promise<{ a: 1 }>; const a = { ...p }
//	invalid: declare const arr: number[]; const a = { ...arr }
//
// Every finding here is a case where the language does something defensible and almost never what
// the author meant. Spreading a Map into an object produces `{}`, because a Map's entries live in
// internal slots rather than in own properties. Spreading an array into an object produces
// `{0: ..., 1: ...}`. Spreading a Promise produces its own properties, which are none. Spreading a
// class instance drops the prototype, so every method disappears.
//
// # The string branch upstream has and this rule does not
//
// Upstream also reports a string spread into an array or a call (`noStringSpread`), because spread
// yields Unicode code points and so splits a family emoji or a combining sequence that a reader
// sees as one glyph, and it asks for `Intl.Segmenter`. Cohere drops that branch on purpose, by
// Kirk's ruling of 2026-10-01. Whether a string should be walked by code point or by grapheme is
// intent the rule cannot see, and spread is the CORRECT code-point iteration (it is what
// `.split("")`, which the upstream message lumps it with, gets wrong). Every string spread ahra
// had wanted code points on purpose, and upstream's only repair, `Array.from(text)`, is the same
// iteration under another name, which launders the finding rather than fixing anything. A finding
// whose fix is a synonym is a false positive. That arm was the rule's only interest in a spread
// element, so the rule now listens to object spreads and JSX spread attributes alone, and the
// divergence is recorded in no_misused_spread.md and as gate-side entries in the differential's
// acknowledged list.
//
// # The object cascade
//
// A spread in an OBJECT, or a JSX attribute, which is the same operation with different syntax, is
// wrong in eight ways, and the order they are asked in is the rule rather than an implementation
// detail. The cascade is: promise, function without properties, Map, array, iterable, class
// instance, class declaration. Each arm returns, so the FIRST match wins and nothing reports twice.
//
// The ordering is load-bearing in both directions. A Map is iterable, so asking `isIterable` first
// would report every Map with the iterable message and lose the `Object.fromEntries` suggestion
// entirely. An array is iterable too, and a class instance may be. Reversing any adjacent pair
// changes which message a real input gets, and the corpus pins several of those pairs directly.
//
// # The iterable arm's string exclusion
//
// `isIterable` excludes strings, and upstream says why at the line: TypeScript already errors on
// spreading a string into an object, so reporting it here would double up on a diagnostic the
// compiler gives. Upstream's array-and-call arm did not share the exclusion, but that arm is gone
// here (see above), so a string is now silent everywhere this rule looks.
//
// # Where a finding points
//
// At the whole spread element, including the three dots, rather than at its argument. That is what
// makes `[...str]` report `...str` and not `str`. Upstream's recorded columns pin it and the
// fixtures assert the span rather than only the id.
//
// # The two suggestions
//
// A Promise gets `addAwait`, which is precedence-sensitive: `...promise` becomes `...await promise`,
// while `...(promise || {})` becomes `...(await (promise || {}))` because `await` binds tighter than
// `||` and inserting the keyword alone would change what is awaited. The shelf already answers the
// precedence question through `IsHigherPrecedenceThanAwait`, so this is a two-branch decision rather
// than a port of upstream's general fixer.
//
// A Map gets `replaceMapSpreadInObject`, which rewrites the spread into `Object.fromEntries(...)`.
// It has two forms and the difference is real: when the spread is the object's ONLY property, the
// whole object literal is replaced, so `{ ...map }` becomes `Object.fromEntries(map)` rather than
// `{ ...Object.fromEntries(map) }`, which would still be wrong. With any sibling property, only the
// spread's argument is wrapped.
//
// Both are SUGGESTIONS rather than fixes, and upstream declares no `fixable` at all. That is right:
// awaiting a value changes when the surrounding code runs, and rewriting a Map spread changes the
// shape of the result. Neither is a repair the engine may apply unattended.
//
// # What upstream's general fixer does that this does not, and why the corpus permits it
//
// Upstream builds the Map suggestion through `getWrappingFixer`, a 276-line general utility that
// re-prints the inner node, parenthesizes it when its precedence is weak, parenthesizes the RESULT
// when the parent's precedence is weak, and inserts a leading semicolon when the previous statement
// lacks one. Only the first of those three can fire here, because the node being replaced is always
// an object literal or a JSX attribute value, neither of which is a weak-precedence parent and
// neither of which can begin a statement that needs a guarding semicolon.
//
// So the narrow version is written directly and the reason is recorded rather than left implicit. It
// reproduces all 21 of upstream's recorded suggestion outputs, including the two that exercise the
// precedence branch: `{ ...(map) }` loses the redundant parentheses because the inner node is read
// through `SkipParentheses`, and `{ ...(map, map) }` keeps them because a comma expression has weak
// precedence and would otherwise change meaning.
var NoMisusedSpread = rule.Rule{
	Name: "@typescript-eslint/no-misused-spread",

	// Every arm asks the checker what the spread argument IS. None of it is answerable from syntax.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		settings, _ := rule.OptionsAs[NoMisusedSpreadOptions](options)

		allowed := func(t *checker.Type) bool {
			return type_checking.TypeMatchesSomeSpecifier(t, settings.Allow, settings.AllowInline, ctx.Program)
		}

		// checkObjectSpread is the cascade. Every arm returns, so the first match wins.
		checkObjectSpread := func(spread *ast.Node, argument *ast.Node) {
			argumentType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
			if argumentType == nil || allowed(argumentType) {
				return
			}

			if noMisusedSpreadIsTypeRecurser(argumentType, func(t *checker.Type) bool {
				return type_checking.IsPromiseLike(ctx.Program, ctx.TypeChecker, t)
			}) {
				ctx.ReportNodeWithSuggestions(spread, buildNoPromiseSpreadInObjectMessage(),
					noMisusedSpreadAwaitSuggestion(ctx, argument))
				return
			}

			if noMisusedSpreadIsFunctionWithoutProperties(ctx.TypeChecker, argumentType) {
				ctx.ReportNode(spread, buildNoFunctionSpreadInObjectMessage())
				return
			}

			if noMisusedSpreadIsMap(ctx, argumentType) {
				ctx.ReportNodeWithSuggestions(spread, buildNoMapSpreadInObjectMessage(),
					noMisusedSpreadMapSuggestions(ctx, spread, argument, argumentType)...)
				return
			}

			if noMisusedSpreadIsArray(ctx.TypeChecker, argumentType) {
				ctx.ReportNode(spread, buildNoArraySpreadInObjectMessage())
				return
			}

			// The string exclusion is upstream's and it is not an optimisation: TypeScript already
			// errors on spreading a string into an object, so reporting here would double a
			// diagnostic the compiler already gives.
			if noMisusedSpreadIsIterable(ctx.TypeChecker, argumentType) && !noMisusedSpreadIsString(argumentType) {
				ctx.ReportNode(spread, buildNoIterableSpreadInObjectMessage())
				return
			}

			// An instance whose copy is complete is not reported, and is not handed on either: an
			// instance's symbol has the class as its value declaration, so the declaration arm below
			// would otherwise report it under the wrong message. See noMisusedSpreadCopiesCompletely.
			if noMisusedSpreadIsTypeRecurser(argumentType, func(part *checker.Type) bool {
				return noMisusedSpreadIsClassInstance(ctx.TypeChecker, part) && !noMisusedSpreadCopiesCompletely(ctx, part)
			}) {
				ctx.ReportNode(spread, buildNoClassInstanceSpreadInObjectMessage())
				return
			}

			if noMisusedSpreadIsTypeRecurser(argumentType, func(part *checker.Type) bool {
				return !noMisusedSpreadIsClassInstance(ctx.TypeChecker, part) && noMisusedSpreadIsClassDeclaration(part)
			}) {
				ctx.ReportNode(spread, buildNoClassDeclarationSpreadInObjectMessage())
			}
		}

		return rule.Listeners{
			// Upstream writes four selectors; two of them (`ArrayExpression > SpreadElement` and
			// `CallExpression > SpreadElement`) fed only the string branch dropped above, so a
			// KindSpreadElement has no listener here. The remaining two are the object spread and
			// the JSX spread attribute. Our parser gives an object spread its own KIND, a
			// KindSpreadAssignment, which is why there is no `if parent is ObjectExpression`.
			ast.KindSpreadAssignment: func(node *ast.Node) {
				argument := node.AsSpreadAssignment().Expression
				if argument == nil {
					return
				}
				checkObjectSpread(node, argument)
			},

			ast.KindJsxSpreadAttribute: func(node *ast.Node) {
				argument := node.AsJsxSpreadAttribute().Expression
				if argument == nil {
					return
				}
				checkObjectSpread(node, argument)
			},
		}
	},
}

// noMisusedSpreadIsTypeRecurser answers upstream's helper of the same name: does any constituent of
// a union or an intersection satisfy the predicate.
//
// Every type question in this rule goes through it, so a `Map<string, number> | number[]` is judged
// by whichever arm its first matching constituent belongs to rather than being declined for not
// being uniformly one thing.
func noMisusedSpreadIsTypeRecurser(t *checker.Type, predicate func(*checker.Type) bool) bool {
	if type_checking.IsUnionType(t) || type_checking.IsIntersectionType(t) {
		for _, part := range t.Types() {
			if noMisusedSpreadIsTypeRecurser(part, predicate) {
				return true
			}
		}
		return false
	}
	return predicate(t)
}

func noMisusedSpreadIsString(t *checker.Type) bool {
	return noMisusedSpreadIsTypeRecurser(t, func(part *checker.Type) bool {
		return type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLike)
	})
}

func noMisusedSpreadIsArray(typeChecker *checker.Checker, t *checker.Type) bool {
	return noMisusedSpreadIsTypeRecurser(t, func(part *checker.Type) bool {
		return checker.Checker_isArrayType(typeChecker, part) || checker.IsTupleType(part)
	})
}

// noMisusedSpreadIsFunctionWithoutProperties is the arm that spares a callable carrying data.
//
// The property count is what makes it narrow: a plain function spread into an object contributes
// nothing, which is almost certainly a forgotten call, while a function with properties attached is
// a namespace object and spreading it is meaningful. Upstream reports only the former.
func noMisusedSpreadIsFunctionWithoutProperties(typeChecker *checker.Checker, t *checker.Type) bool {
	return noMisusedSpreadIsTypeRecurser(t, func(part *checker.Type) bool {
		return len(type_checking.GetCallSignatures(typeChecker, part)) > 0 &&
			len(checker.Checker_getPropertiesOfType(typeChecker, part)) == 0
	})
}

func noMisusedSpreadIsIterable(typeChecker *checker.Checker, t *checker.Type) bool {
	return noMisusedSpreadIsTypeRecurser(t, func(part *checker.Type) bool {
		return type_checking.GetWellKnownSymbolPropertyOfType(part, "iterator", typeChecker) != nil
	})
}

func noMisusedSpreadIsMap(ctx rule.Context, t *checker.Type) bool {
	return noMisusedSpreadIsTypeRecurser(t, func(part *checker.Type) bool {
		return type_checking.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, part, "Map", "ReadonlyMap", "WeakMap")
	})
}

// noMisusedSpreadIsClassInstance separates an instance from the class itself.
//
// The two tests are opposite ends of the same question and both are needed. A type with its OWN
// construct signature is the class, not an instance, so it is declined here and picked up by the
// class-declaration arm below. A type whose SYMBOL resolves, at one of its declarations, to
// something with a construct signature is an instance of that class.
func noMisusedSpreadIsClassInstance(typeChecker *checker.Checker, t *checker.Type) bool {
	return noMisusedSpreadIsTypeRecurser(t, func(part *checker.Type) bool {
		if len(type_checking.GetConstructSignatures(typeChecker, part)) > 0 {
			return false
		}

		symbol := checker.Type_symbol(part)
		if symbol == nil {
			return false
		}

		// Every declaration is asked, matching upstream's `some` over `getDeclarations()`, and the
		// loop is INERT here rather than load-bearing. Restricting it to the first declaration
		// survives the whole suite, and that verdict is correct.
		//
		// The reason is that `GetTypeOfSymbolAtLocation` resolves the MERGED symbol's type at
		// whichever declaration it is handed, so every entry answers the same question. Probed
		// directly over three merged shapes, printing the construct-signature count per declaration:
		//
		//	interface A {} then class A {}                [0] interface: 1   [1] class: 1
		//	class A {} then interface A {}                [0] class: 1       [1] interface: 1
		//	interface B, interface B, then class B {}     [0] 1  [1] 1  [2] 1
		//
		// So the first declaration already carries the constructor in every ordering, and no input
		// can distinguish the loop from an index. The standing advice never to index [0] still
		// holds and is why the loop is written rather than an index: the equivalence rests on this
		// particular accessor resolving through the merged symbol, and a future change to ask a
		// per-declaration question instead would make the loop live again with no test to notice.
		for _, declaration := range symbol.Declarations {
			declaredType := typeChecker.GetTypeOfSymbolAtLocation(symbol, declaration)
			if declaredType == nil {
				continue
			}
			if len(type_checking.GetConstructSignatures(typeChecker, declaredType)) > 0 {
				return true
			}
		}
		return false
	})
}

// noMisusedSpreadIsClassDeclaration answers whether the type IS a class rather than an instance.
//
// Two ways to be one. An instantiation expression type, which is what `Foo<string>` produces when
// `Foo` is a generic class referenced without `new`, carries an object flag saying so. Otherwise the
// symbol's value declaration is literally a class declaration or a class expression.
func noMisusedSpreadIsClassDeclaration(t *checker.Type) bool {
	return noMisusedSpreadIsTypeRecurser(t, func(part *checker.Type) bool {
		if type_checking.IsObjectType(part) &&
			checker.Type_objectFlags(part)&checker.ObjectFlagsInstantiationExpressionType != 0 {
			return true
		}

		symbol := checker.Type_symbol(part)
		if symbol == nil || symbol.ValueDeclaration == nil {
			return false
		}
		switch symbol.ValueDeclaration.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return true
		}
		return false
	})
}

// noMisusedSpreadCopiesCompletely says whether spreading an instance of this class loses nothing, so
// the class-instance arm has no harm to report. Cohere departs from upstream here, by @system_cohere's
// ruling of 2026-10-03 on api's decorated GraphQL and Serializable data classes (#ynneze5).
//
// The arm's harm is the prototype the copy drops. A class with nothing on its prototype loses
// nothing, and a spread of one is the ordinary way to build a new object from it. So an instance is
// exempt only when every one of these holds for its class and every base class in its chain:
//
//   - Every declaration of the symbol is a class written in the program's own source, not ambient
//     (`declare class`) and not in a package or a declaration file. Only then is the prototype ours to
//     read. An interface merged into the class can declare methods a mixin installs, so it disqualifies.
//   - No instance method, no get or set accessor, and no auto-accessor (`accessor x`), since each lives
//     on the prototype. Static members live on the constructor, which an instance never carried.
//   - No `#private` member, since a spread copies only public own properties.
//   - Every decorator anywhere on the class resolves to a declaration in the program's own source.
//
// The remaining trust is the last clause's: one of our own decorators could install a prototype
// accessor at runtime, which no type shows, and the copy would then silently lack it. Base's and
// api's decorators register metadata and define nothing on the prototype, and a third-party
// decorator, MobX's @observable or Lit's @property, stays outside the exemption entirely.
//
// A union exempts only when every class instance in it is exempt, because the caller's recurser
// reports on the first constituent that is not.
func noMisusedSpreadCopiesCompletely(ctx rule.Context, t *checker.Type) bool {
	symbol := checker.Type_symbol(t)
	if symbol == nil {
		return false
	}
	seen := map[*ast.Symbol]bool{}
	for symbol != nil {
		if seen[symbol] || !noMisusedSpreadClassLosesNothing(ctx, symbol) {
			return false
		}
		seen[symbol] = true

		declared := checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, symbol)
		if declared == nil {
			return false
		}
		bases := checker.Checker_getBaseTypes(ctx.TypeChecker, declared)
		switch len(bases) {
		case 0:
			return true
		case 1:
			// A base that is not a class (a mixin's intersection, an interface) has no class
			// declaration to read, so the next pass refuses it.
			symbol = checker.Type_symbol(bases[0])
		default:
			return false
		}
	}
	return false
}

// noMisusedSpreadClassLosesNothing is one link of the chain: the symbol's own declarations.
func noMisusedSpreadClassLosesNothing(ctx rule.Context, symbol *ast.Symbol) bool {
	if len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !ast.IsClassLike(declaration) || !noMisusedSpreadIsOwnSource(declaration) {
			return false
		}
		for _, member := range declaration.Members() {
			if ast.IsStatic(member) {
				continue
			}
			if ast.IsMethodDeclaration(member) || ast.IsGetAccessorDeclaration(member) ||
				ast.IsSetAccessorDeclaration(member) || ast.IsAutoAccessorPropertyDeclaration(member) ||
				ast.IsPrivateIdentifierClassElementDeclaration(member) {
				return false
			}
		}
		if !noMisusedSpreadDecoratorsAreOwnSource(ctx, declaration) {
			return false
		}
	}
	return true
}

// noMisusedSpreadDecoratorsAreOwnSource asks every decorator the class can carry: its own, each
// member's, and each constructor parameter's. Those are the only places one can sit on a class whose
// members are all fields, and reading them never enters a function body, which the rule's Shapes
// reach promises not to do.
func noMisusedSpreadDecoratorsAreOwnSource(ctx rule.Context, classNode *ast.Node) bool {
	holders := []*ast.Node{classNode}
	for _, member := range classNode.Members() {
		holders = append(holders, member)
		if ast.IsConstructorDeclaration(member) {
			holders = append(holders, member.Parameters()...)
		}
	}
	for _, holder := range holders {
		modifiers := holder.Modifiers()
		if modifiers == nil {
			continue
		}
		for _, modifier := range modifiers.Nodes {
			if ast.IsDecorator(modifier) && !noMisusedSpreadDecoratorIsOwnSource(ctx, modifier.AsDecorator().Expression) {
				return false
			}
		}
	}
	return true
}

// noMisusedSpreadDecoratorIsOwnSource resolves `@Name`, `@Name(...)`, `@namespace.Name` and
// `@namespace.Name(...)` through any import alias to the declarations of the function applied.
func noMisusedSpreadDecoratorIsOwnSource(ctx rule.Context, expression *ast.Node) bool {
	target := ast.SkipParentheses(expression)
	if ast.IsCallExpression(target) {
		target = ast.SkipParentheses(target.AsCallExpression().Expression)
	}
	if ast.IsPropertyAccessExpression(target) {
		target = target.AsPropertyAccessExpression().Name()
	}
	if !ast.IsIdentifier(target) {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(target)
	if symbol == nil {
		return false
	}
	symbol = checker.SkipAlias(symbol, ctx.TypeChecker)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !noMisusedSpreadIsOwnSource(declaration) {
			return false
		}
	}
	return true
}

// noMisusedSpreadIsOwnSource is a declaration whose implementation the program's own source holds:
// not ambient (`declare class`, `declare function`, anything in a declaration file or a `declare`
// block), and not inside a package.
func noMisusedSpreadIsOwnSource(declaration *ast.Node) bool {
	sourceFile := ast.GetSourceFileOfNode(declaration)
	return sourceFile != nil && !sourceFile.IsDeclarationFile && declaration.Flags&ast.NodeFlagsAmbient == 0 &&
		!strings.Contains(sourceFile.FileName().AsString(), "/node_modules/")
}

// noMisusedSpreadAwaitSuggestion offers `await` before the spread argument.
//
// Two branches and the precedence question decides between them. A high-precedence argument takes
// the keyword directly, so `...promise` becomes `...await promise`. A low-precedence one has to be
// parenthesized, because `await a || b` parses as `(await a) || b` and would await the wrong thing:
// `...(promise || {})` becomes `...(await (promise || {}))`.
//
// The shelf's `IsHigherPrecedenceThanAwait` answers this, comparing the node's expression precedence
// against the await operator's. Upstream asks the identical question of its own tree.
func noMisusedSpreadAwaitSuggestion(ctx rule.Context, argument *ast.Node) rule.Suggestion {
	// The precedence question is asked of the node INSIDE any parentheses, not of the parenthesized
	// expression, and that is a substrate difference rather than a refinement. ESTree has no
	// parenthesis node, so upstream asks this of the `||` or the conditional directly. Ours wraps
	// them, and a parenthesized expression is high precedence, so asking the outer node takes the
	// keyword-only branch for every input and produces `...await (promise || {})` where upstream
	// produces `...(await (promise || {}))`.
	//
	// Measured on the two corpus cases that carry the low-precedence shape, printing both kinds:
	// the argument is KindParenthesizedExpression answering true, and the inner node is
	// KindBinaryExpression or KindConditionalExpression answering false. A parenthesized IDENTIFIER
	// still answers true from the inside, which is why the skip is the right correction rather than
	// simply inverting the branch.
	inner := argument
	if unwrapped := ast.SkipParentheses(argument); unwrapped != nil {
		inner = unwrapped
	}

	if type_checking.IsHigherPrecedenceThanAwait(inner) {
		return rule.Suggestion{
			Message: buildAddAwaitMessage(),
			Fixes:   []rule.Fix{ctx.InsertBefore(argument, "await ")},
		}
	}
	// The wrap goes around the INNER node rather than around the argument, for the same substrate
	// reason. Upstream's `node` here is the unparenthesized expression, so when the source already
	// spells `(promise || {})` the parentheses it inserts land INSIDE the existing pair and the
	// result reads `(await (promise || {}))`. Wrapping our argument instead, which includes those
	// parentheses, doubles them into `await ((promise || {}))`, which is valid and is not what
	// upstream writes. Both corpus rows assert the exact text, which is what caught it.
	return rule.Suggestion{
		Message: buildAddAwaitMessage(),
		Fixes: []rule.Fix{
			ctx.InsertBefore(inner, "await ("),
			ctx.InsertAfter(inner, ")"),
		},
	}
}

// noMisusedSpreadMapSuggestions offers the `Object.fromEntries` rewrite.
//
// Two shapes. When the spread is the object's ONLY property the whole literal is replaced, because
// `{ ...Object.fromEntries(map) }` would still be an object built by spreading. With any sibling the
// argument alone is wrapped, so the other properties survive.
//
// Returns nothing when the type is a union with a non-Map constituent, matching upstream: the
// rewrite is only sound if every constituent is a Map, and the finding is still reported without it.
func noMisusedSpreadMapSuggestions(ctx rule.Context, spread *ast.Node, argument *ast.Node, argumentType *checker.Type) []rule.Suggestion {
	for part := range type_checking.UnionTypePartsSeq(argumentType) {
		if !noMisusedSpreadIsMap(ctx, part) {
			return nil
		}
	}

	// The argument is read through the parentheses so that `{ ...(map) }` rewrites to
	// `Object.fromEntries(map)` rather than `Object.fromEntries((map))`. Upstream reaches the same
	// text because its tree has no parenthesis node to read.
	inner := argument
	if unwrapped := ast.SkipParentheses(argument); unwrapped != nil {
		inner = unwrapped
	}
	innerRange := rule.TokenRange(ctx.SourceFile, inner)
	innerText := ctx.SourceFile.Text()[innerRange.Pos():innerRange.End()]

	// A weak-precedence inner node keeps its parentheses, because `Object.fromEntries(map, map)`
	// would pass two arguments where the comma expression meant one. Upstream's general fixer makes
	// the same decision through the same predicate.
	if !type_checking.IsStrongPrecedenceNode(inner) {
		innerText = "(" + innerText + ")"
	}

	replacement := "Object.fromEntries(" + innerText + ")"

	// The sole-property case replaces the whole object literal. Only an object literal can be
	// replaced this way: a JSX attribute has no equivalent, so it always takes the argument form,
	// which is what upstream's parent test amounts to here.
	if parent := spread.Parent; parent != nil && ast.IsObjectLiteralExpression(parent) &&
		len(parent.AsObjectLiteralExpression().Properties.Nodes) == 1 {
		return []rule.Suggestion{{
			Message: buildReplaceMapSpreadInObjectMessage(),
			Fixes:   []rule.Fix{rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, parent), replacement)},
		}}
	}

	// The span is the INNER node, so existing parentheses around the argument SURVIVE the rewrite:
	// `{ ...(map), other: 1 }` becomes `{ ...(Object.fromEntries(map)), other: 1 }`. That reads like
	// a defect and it is upstream's behaviour, measured rather than assumed. Upstream replaces the
	// node its own tree hands it, which is the unparenthesized expression, so the parentheses are
	// simply outside the range it touches.
	//
	// This port originally replaced the whole argument INCLUDING the parentheses, which produces the
	// tidier `{ ...Object.fromEntries(map), other: 1 }`. Every imported fixture stayed green over
	// that, because the corpus writes a parenthesized argument only in the sole-property shape, where
	// the whole object literal is replaced and this branch never runs. Found by mutation and settled
	// by driving the installed rule on three parenthesized-with-sibling inputs, all of which keep the
	// parentheses. Tidying is not fidelity, so upstream's answer is reproduced and pinned below.
	return []rule.Suggestion{{
		Message: buildReplaceMapSpreadInObjectMessage(),
		Fixes:   []rule.Fix{rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, inner), replacement)},
	}}
}

func buildNoArraySpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id:          "noArraySpreadInObject",
		Description: "Using the spread operator on an array in an object will result in a list of indices.",
	}
}

func buildNoClassDeclarationSpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id: "noClassDeclarationSpreadInObject",
		Description: "Using the spread operator on class declarations will spread only their static properties, " +
			"and will lose their class prototype.",
	}
}

func buildNoClassInstanceSpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id:          "noClassInstanceSpreadInObject",
		Description: "Using the spread operator on class instances will lose their class prototype.",
	}
}

func buildNoFunctionSpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id: "noFunctionSpreadInObject",
		Description: "Using the spread operator on a function without additional properties can cause " +
			"unexpected behavior. Did you forget to call the function?",
	}
}

func buildNoIterableSpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id:          "noIterableSpreadInObject",
		Description: "Using the spread operator on an Iterable in an object can cause unexpected behavior.",
	}
}

func buildNoMapSpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id: "noMapSpreadInObject",
		Description: "Using the spread operator on a Map in an object will result in an empty object. " +
			"Did you mean to use `Object.fromEntries(map)` instead?",
	}
}

func buildNoPromiseSpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id: "noPromiseSpreadInObject",
		Description: "Using the spread operator on Promise in an object can cause unexpected behavior. " +
			"Did you forget to await the promise?",
	}
}

func buildAddAwaitMessage() rule.Message {
	return rule.Message{
		Id:          "addAwait",
		Description: "Add await operator.",
	}
}

func buildReplaceMapSpreadInObjectMessage() rule.Message {
	return rule.Message{
		Id:          "replaceMapSpreadInObject",
		Description: "Replace map spread in object with `Object.fromEntries()`",
	}
}

// noMisusedSpreadRawOptions is the wire shape. One key, carrying the same heterogeneous allow list
// prefer-promise-reject-errors takes, so the entry type is shared rather than restated: upstream
// builds both schemas from the same `readonlynessOptionsSchema.properties.allow`, and two copies of
// a hand-written `UnmarshalJSON` would be two places for the wire format to drift.
type noMisusedSpreadRawOptions struct {
	Allow []preferPromiseRejectErrorsRawSpecifier `json:"allow"`
}

// DecodeNoMisusedSpreadOptions maps upstream's JSON onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` for the same reason as its sibling: `allow`
// arrives as one array mixing bare strings with specifier objects and leaves as two typed fields,
// which no struct tag expresses.
//
// The empty allow list is both the default and the correct zero value, so unlike restrict-plus-
// operands there is no default to restore on nil: a rule handed nil options allows nothing, which is
// exactly upstream's `defaultOptions`.
func DecodeNoMisusedSpreadOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noMisusedSpreadRawOptions]()(raw)
	if err != nil {
		return NoMisusedSpreadOptions{}, err
	}

	wire, _ := decoded.(noMisusedSpreadRawOptions)

	var options NoMisusedSpreadOptions
	for _, entry := range wire.Allow {
		if entry.inline != "" {
			options.AllowInline = append(options.AllowInline, entry.inline)
			continue
		}

		specifier := type_checking.TypeOrValueSpecifier{Name: entry.Name, Path: entry.Path, Package: entry.Package}
		switch entry.From {
		case "file":
			specifier.From = type_checking.TypeOrValueSpecifierFromFile
		case "lib":
			specifier.From = type_checking.TypeOrValueSpecifierFromLib
		case "package":
			specifier.From = type_checking.TypeOrValueSpecifierFromPackage
		default:
			continue
		}
		options.Allow = append(options.Allow, specifier)
	}

	return options, nil
}
