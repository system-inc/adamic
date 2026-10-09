package typescript

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoDeprecated is typescript-eslint's `no-deprecated`: using something marked `@deprecated` is a
// finding at the use site, not at the declaration.
//
//	valid:   /** @deprecated */ function a() {}            the declaration itself
//	valid:   import { a } from './a';                       the import, not a use
//	invalid: a();                                           the use
//	invalid: export { /** @deprecated */ b }; ... b();       the tag on the ALIAS, not the original
//
// # The tag can sit on a symbol nobody would think to ask
//
// The last shape above is the whole reason this rule is hard, and it is why the shim carries two
// extra accessors. `export { /** @deprecated */ NormalClass }` puts the tag on the export specifier
// while the class declaration underneath is unmarked. Resolving the import to its target jumps
// straight past the specifier and answers "not deprecated", which is wrong; walking one hop at a
// time sees it. Measured on exactly that input before this file was written:
//
//	resolveAlias                -> KindClassDeclaration    deprecated=false
//	getImmediateAliasedSymbol   -> KindExportSpecifier     deprecated=true, reason="Reason"
//
// So `deprecationInAliasChain` below walks hop by hop rather than resolving, and the third row of
// that measurement is the one that matters most: `normalFunction`, exported through the same block
// with no tag, answers false at every hop. A walk that reported unconditionally would pass both of
// the first two rows and fail only that one.
//
// # Where this diverges from upstream, and why it has to
//
// Upstream reads deprecation through `symbol.getJsDocTags(checker)`, a LANGUAGE SERVICE API.
// typescript-go has no such method: it exposes deprecation at the DECLARATION level instead, via
// `ast.GetJSDocDeprecatedTag`. The two are not the same question. `getJsDocTags` merges the tags of
// every declaration a symbol has; the declaration-level call answers about one declaration.
//
// For this rule the difference is visible in exactly one place, and upstream documents it as a
// deliberate choice rather than an accident: a function with overloads has one declaration per
// overload, and asking the merged symbol whether it is deprecated says yes when ANY overload is.
// Upstream works around its own API there by passing `false` for `checkDeprecationsOfAliasedSymbol`
// on function and method declarations and relying on the resolved SIGNATURE instead. Reading
// declarations one at a time gives that behaviour directly rather than as a workaround, so
// `deprecationOnSymbol` unions across declarations while `deprecationOnSignatureDeclaration` asks
// about the single declaration a call actually resolved to.
//
// # The reason text
//
// `deprecatedWithReason` is a distinct message id and 26 of upstream's cases assert its text, so the
// reason is part of the contract rather than decoration. It comes off the tag's comment list, which
// is a NodeList rather than a string because `{@link stat}` parses into its own node. Concatenating
// the pieces' text reproduces `displayPartsToString`.
var NoDeprecated = rule.Rule{
	Name: "@typescript-eslint/no-deprecated",

	// The rule cannot answer anything without resolving symbols across files.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Every listener reads the checker unconditionally, so decline the file once here rather
		// than test for nil at the top of four listeners and once per node inside them.
		//
		// Unreachable through registration, since NeedsTypeChecker is declared right above. It is
		// here for the harness path, where a Context is built by hand and the checker is nil: the
		// registry's crash corpus takes exactly that path, and without this the rule panicked on
		// the first identifier of every shape it was handed. A test in this package pins both the
		// declaration and this decline, so removing either fails loudly rather than going quiet.
		if ctx.TypeChecker == nil {
			return nil
		}

		settings, _ := rule.OptionsAs[NoDeprecatedOptions](options)

		// report emits with the right message id for whether a reason was given.
		//
		// Upstream branches on `reason ? withReason : plain`, so an EMPTY reason, a bare
		// `/** @deprecated */`, takes the plain branch. The empty string and the absent tag are
		// different states and only the presence flag separates them, which is why every helper
		// below returns (string, bool) rather than a string whose emptiness means absent.
		report := func(node *ast.Node, name string, reason string) {
			if reason != "" {
				ctx.ReportNode(node, rule.Message{
					Id:          "deprecatedWithReason",
					Description: "`" + name + "` is deprecated. " + reason,
				})
				return
			}
			ctx.ReportNode(node, rule.Message{
				Id:          "deprecated",
				Description: "`" + name + "` is deprecated.",
			})
		}

		checkIdentifier := func(node *ast.Node) {
			if isDeprecationDeclarationSite(node) || isInsideImportForDeprecation(node) {
				return
			}
			reason, found := deprecationReasonFor(ctx, node)
			if !found {
				return
			}
			if deprecationIsAllowed(ctx, node, settings) {
				return
			}
			report(node, reportedDeprecationName(node), reason)
		}

		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil {
					return
				}
				switch parent.Kind {
				case ast.KindExportDeclaration, ast.KindNamespaceExport:
					return
				case ast.KindJsxClosingElement:
					// `<A></A>` names the component twice and is one use. Upstream's JSXIdentifier
					// listener skips a closing element for the same reason; without this the rule
					// reports the same deprecation twice on the same element.
					return
				}
				// A computed member's property is handled by the element-access listener, which
				// reads the property NAME off the type rather than resolving the identifier.
				if parent.Kind == ast.KindElementAccessExpression &&
					parent.AsElementAccessExpression().ArgumentExpression == node {
					return
				}
				if parent.Kind == ast.KindExportSpecifier {
					// Only the exported side, and only when the alias itself carries no tag: a tag
					// on the alias is the export's own statement about itself, not a use.
					if parent.AsExportSpecifier().Name() != node {
						return
					}
					if _, tagged := deprecationOnSymbol(ctx, ctx.TypeChecker.GetSymbolAtLocation(node)); tagged {
						return
					}
				}
				checkIdentifier(node)
			},
			ast.KindPrivateIdentifier: checkIdentifier,
			ast.KindSuperKeyword:      checkIdentifier,
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkDeprecatedElementAccess(ctx, node, settings, report)
			},
		}
	},
}

// reportedDeprecationName is upstream's `getReportedNodeName`.
func reportedDeprecationName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindSuperKeyword:
		return "super"
	case ast.KindPrivateIdentifier:
		return "#" + strings.TrimPrefix(node.Text(), "#")
	}
	return node.Text()
}

// deprecationTagOn finds the `@deprecated` tag governing a declaration.
//
// The walk up is not optional and not defensive. JSDoc attaches to the STATEMENT, so
// `/** @deprecated */ export const bare = 1;` hangs its tag on the VariableStatement while the
// symbol's declaration is the VariableDeclaration underneath. Asking the declaration directly
// answers no. Measured on exactly that input: direct lookup false, walk-up true, while an
// undeprecated sibling stays false under both.
//
// The flag test first is the compiler's own fast path: the parser sets
// NodeFlagsPossiblyContainsDeprecatedTag when a comment's text contains `@deprecated`, so a tree
// with no such comment anywhere costs one flag test per declaration rather than a JSDoc walk.
func deprecationTagOn(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	if ast.GetCombinedNodeFlags(declaration)&ast.NodeFlagsPossiblyContainsDeprecatedTag == 0 {
		return nil
	}
	for node := declaration; node != nil; node = node.Parent {
		if node.Flags&ast.NodeFlagsPossiblyContainsDeprecatedTag != 0 {
			return ast.GetJSDocDeprecatedTag(node)
		}
	}
	return nil
}

// deprecationReasonOnDeclaration answers about ONE declaration, returning the tag's text and whether
// a tag was there at all.
//
// The two returns are not redundant: `/** @deprecated */` with no text is a real finding with an
// empty reason, and it takes the `deprecated` message rather than `deprecatedWithReason`. Collapsing
// them into a single string would make an untexted tag indistinguishable from no tag.
func deprecationReasonOnDeclaration(declaration *ast.Node) (string, bool) {
	tag := deprecationTagOn(declaration)
	if tag == nil {
		return "", false
	}
	comment := tag.AsJSDocDeprecatedTag().Comment
	if comment == nil {
		return "", true
	}
	// `displayPartsToString`. The comment is a list rather than a string because an inline
	// `{@link stat}` is its own node, and 26 of upstream's cases assert the reassembled text.
	var builder strings.Builder
	for _, piece := range comment.Nodes {
		builder.WriteString(piece.Text())
	}
	return strings.TrimSpace(builder.String()), true
}

// deprecationOnSymbol unions across a symbol's declarations, which is what upstream's
// `getJsDocDeprecation` does by merging tags.
//
// The first tagged declaration wins its reason. For a symbol with one declaration, the common case,
// that is simply that declaration's answer.
func deprecationOnSymbol(ctx rule.Context, symbol *ast.Symbol) (string, bool) {
	if symbol == nil {
		return "", false
	}
	for _, declaration := range symbol.Declarations {
		if reason, found := deprecationReasonOnDeclaration(declaration); found {
			return reason, true
		}
	}
	return "", false
}

// deprecationInAliasChain walks the alias chain one hop at a time.
//
// Resolving the alias in one jump is the tempting version and it is wrong: the tag frequently sits
// on an intermediate export specifier that a resolve skips straight past. See the measurement in
// this rule's header.
//
// `checkTarget` mirrors upstream's `checkDeprecationsOfAliasedSymbol`. It is false for calls to
// functions and methods, where the ultimate target's merged tags would report a deprecated OVERLOAD
// for a call that resolved to an undeprecated one; those calls consult the resolved signature's own
// declaration instead.
//
// The hop count is bounded by the chain the checker built, and each hop must produce a symbol the
// checker itself returned, so a cycle cannot be constructed here. The guard on the alias flag is
// required rather than defensive: both accessors PANIC on a non-alias, in every build, which is
// asserted in internal/program/shim_alias_test.go.
func deprecationInAliasChain(ctx rule.Context, symbol *ast.Symbol, checkTarget bool) (string, bool) {
	if symbol == nil {
		return "", false
	}
	if symbol.Flags&ast.SymbolFlagsAlias == 0 {
		if checkTarget {
			return deprecationOnSymbol(ctx, symbol)
		}
		return "", false
	}

	target := checker.Checker_resolveAlias(ctx.TypeChecker, symbol)

	for symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if reason, found := deprecationOnSymbol(ctx, symbol); found {
			return reason, true
		}
		if len(symbol.Declarations) == 0 {
			break
		}
		next := checker.Checker_getImmediateAliasedSymbol(ctx.TypeChecker, symbol)
		if next == nil {
			break
		}
		symbol = next
		if checkTarget && symbol == target {
			return deprecationOnSymbol(ctx, symbol)
		}
	}
	return "", false
}

// isDeprecationDeclarationSite is upstream's `isDeclaration`: the name in a declaration is the thing
// being declared, not a use of it, so `/** @deprecated */ function a() {}` does not report on `a`.
//
// Upstream switches on the estree parent type. Our parser names the same positions differently and,
// in two places, does not distinguish them at all, so this asks the question positionally: is the
// node the NAME of its parent, rather than something the parent refers to.
func isDeprecationDeclarationSite(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression,
		ast.KindVariableDeclaration, ast.KindEnumMember,
		ast.KindMethodDeclaration, ast.KindPropertyDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindEnumDeclaration, ast.KindInterfaceDeclaration,
		ast.KindMethodSignature, ast.KindPropertySignature,
		ast.KindModuleDeclaration, ast.KindTypeAliasDeclaration,
		ast.KindTypeParameter, ast.KindImportEqualsDeclaration,
		ast.KindParameter:
		return parent.Name() == node

	case ast.KindBindingElement:
		// `const { b } = a` reads `a.b`, so the shorthand binding is a USE of the source property
		// even though it also declares `b`. Upstream reaches the same answer by a different route:
		// its parser models this as a Property whose key and value are the same node, and it
		// declares only the value side, leaving the key side to report.
		//
		// A RENAMING form splits the two roles across two nodes, `const { b: renamed } = a`, and
		// then the roles are the obvious ones: the property name is the use, the new name is the
		// declaration.
		element := parent.AsBindingElement()
		// A default value is neither the property read nor the binding declared: in
		// `const [c = a] = []` the `a` is an ordinary use that happens to sit inside a pattern.
		if element.Initializer == node {
			return false
		}
		if element.PropertyName != nil {
			return element.Name() == node
		}
		// An array pattern has no property to read, so its binding is only a declaration:
		// `const [a] = xs` names `a` and reads no member of anything.
		if parent.Parent != nil && parent.Parent.Kind == ast.KindArrayBindingPattern {
			return true
		}
		return false

	case ast.KindPropertyAssignment:
		// `const baz = { foo: bar }` -- `foo` names a key this literal declares, `bar` is a use.
		// In an object PATTERN the roles invert, but our parser routes patterns through
		// BindingElement rather than PropertyAssignment, so the inversion is handled above.
		return parent.AsPropertyAssignment().Name() == node

	case ast.KindShorthandPropertyAssignment:
		// `{ foo }` in a literal declares the key and uses the binding through one identifier.
		// Upstream reaches the value symbol separately; the identifier itself is the key's
		// declaration, so returning true here would silence the use. It does not: the use is
		// resolved through the shorthand value symbol in deprecationReasonFor.
		return false
	}
	return false
}

// isInsideImportForDeprecation is upstream's `isInsideImport`: importing a deprecated thing is not
// using it, so only the later reference reports.
//
// The walk stops at the same node kinds upstream stops at. Those stops are what keeps it from
// running to the file root on every identifier in the program: an ordinary expression is inside a
// statement inside a function or the Program, all of which answer false immediately.
func isInsideImportForDeprecation(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindImportDeclaration, ast.KindImportEqualsDeclaration:
			return true
		case ast.KindArrowFunction, ast.KindExportDeclaration, ast.KindExportAssignment,
			ast.KindBlock, ast.KindClassDeclaration, ast.KindInterfaceDeclaration,
			ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindSourceFile, ast.KindUnionType, ast.KindVariableDeclaration:
			return false
		}
	}
	return false
}

// callLikeParentFor is upstream's `getCallLikeNode`: climbs out of a member access so `a.b.c()`
// asks about the CALL when the identifier is `c`, and about nothing when it is `a`.
func callLikeParentFor(node *ast.Node) *ast.Node {
	callee := node
	for callee.Parent != nil &&
		callee.Parent.Kind == ast.KindPropertyAccessExpression &&
		callee.Parent.AsPropertyAccessExpression().Name() == callee {
		callee = callee.Parent
	}
	parent := callee.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == callee {
			return parent
		}
	case ast.KindNewExpression:
		if parent.AsNewExpression().Expression == callee {
			return parent
		}
	case ast.KindTaggedTemplateExpression:
		if parent.AsTaggedTemplateExpression().Tag == callee {
			return parent
		}
	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		// Both opening forms, through the shared accessor: `<A />` parses as a self-closing element
		// and never produces a JsxOpeningElement, so reading only the first form would make the
		// rule silent on the shape most JSX is actually written in.
		if tagName, _ := jsx.ElementParts(parent); tagName == callee {
			return parent
		}
	}
	return nil
}

// deprecationOnSignatureDeclaration asks about the ONE declaration a call resolved to.
//
// This is where the declaration-level API is a better fit than upstream's merged-symbol one. A
// function with overloads has one declaration per overload; asking the symbol says yes when any
// overload is deprecated, which reports a call that resolved to an undeprecated overload. Upstream
// works around that by suppressing the aliased-symbol check for functions and methods and consulting
// the signature. Asking the resolved signature's own declaration is the same answer, directly.
func deprecationOnSignatureDeclaration(ctx rule.Context, callLike *ast.Node) (string, bool) {
	signature := ctx.TypeChecker.GetResolvedSignature(callLike)
	if signature == nil {
		return "", false
	}
	return deprecationReasonOnDeclaration(signature.Declaration())
}

// declaresAsFunctionOrMethod reports whether a symbol's first declaration is a function or method.
//
// Upstream's `symbolDeclarationKind` test, and it selects between the two strategies above. A
// PROPERTY whose type is function-like carries its tag on the symbol rather than on any signature,
// which is why the test is on the declaration's kind rather than on whether the thing is callable.
func declaresAsFunctionOrMethod(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	switch symbol.Declarations[0].Kind {
	case ast.KindMethodDeclaration, ast.KindFunctionDeclaration, ast.KindMethodSignature:
		return true
	}
	return false
}

// deprecationForCallLike is upstream's `getCallLikeDeprecation`.
func deprecationForCallLike(ctx rule.Context, node *ast.Node, callLike *ast.Node) (string, bool) {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)

	aliased := symbol
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		aliased = checker.Checker_resolveAlias(ctx.TypeChecker, symbol)
	}

	if !declaresAsFunctionOrMethod(aliased) {
		if reason, found := deprecationInAliasChain(ctx, symbol, true); found {
			return reason, true
		}
		if reason, found := deprecationOnSignatureDeclaration(ctx, callLike); found {
			return reason, true
		}
		return deprecationOnSymbol(ctx, aliased)
	}

	// A function or method. The chain is walked WITHOUT consulting the ultimate target, because the
	// target's merged declarations would answer for every overload rather than the one called; the
	// resolved signature answers for the one called.
	if reason, found := deprecationInAliasChain(ctx, symbol, false); found {
		return reason, true
	}
	return deprecationOnSignatureDeclaration(ctx, callLike)
}

// deprecationForJsxAttribute is upstream's `getJSXAttributeDeprecation`: `<A b="" />` reports on `b`
// when the element's prop type marks it deprecated, so the tag is on a property of the ELEMENT's
// attribute type rather than on anything the attribute itself resolves to.
//
// Two routes to that type, because one route does not reach both element kinds. A COMPONENT's
// attribute type is the contextual type of its tag name, which is upstream's route and the only one
// it needs, since estree hands it a tag name that resolves. An INTRINSIC element's `<div>` has no
// contextual type at the name: its attribute type comes from the JSX namespace's IntrinsicElements
// rather than from a signature, so the name route answers nil and the whole arm goes silent on
// every `<div aria-grabbed>`. Measured with the tag declared in the fixture: the name route reported
// on a component prop and nothing on an intrinsic one.
//
// The second route asks the checker for the attribute's own contextual type and takes the SYMBOL
// that type came from, which is the property the tag sits on. Trying that route alone is worse than
// either: measured, it went silent on both, because for a component the attribute's contextual type
// is the property's type rather than the property.
func deprecationForJsxAttribute(ctx rule.Context, opening *ast.Node, propertyName string) (string, bool) {
	tagName, _ := jsx.ElementParts(opening)
	if tagName != nil {
		if contextual := ctx.TypeChecker.GetContextualType(tagName, 0); contextual != nil {
			if reason, found := deprecationOnSymbol(ctx,
				ctx.TypeChecker.GetPropertyOfType(contextual, propertyName)); found {
				return reason, true
			}
		}
	}
	return deprecationOnIntrinsicAttribute(ctx, opening, propertyName)
}

// deprecationOnIntrinsicAttribute reaches an intrinsic element's attribute type through the
// element's own type rather than through its tag name.
func deprecationOnIntrinsicAttribute(ctx rule.Context, opening *ast.Node, propertyName string) (string, bool) {
	_, attributes := jsx.ElementParts(opening)
	if attributes == nil {
		return "", false
	}
	attributesType := ctx.TypeChecker.GetContextualType(attributes, 0)
	if attributesType == nil {
		return "", false
	}
	return deprecationOnSymbol(ctx, ctx.TypeChecker.GetPropertyOfType(attributesType, propertyName))
}

// deprecationReasonFor is upstream's `getDeprecationReason`: the four ways a use can be deprecated,
// tried in upstream's order.
func deprecationReasonFor(ctx rule.Context, node *ast.Node) (string, bool) {
	if callLike := callLikeParentFor(node); callLike != nil {
		return deprecationForCallLike(ctx, node, callLike)
	}

	parent := node.Parent
	if parent != nil && node.Kind != ast.KindSuperKeyword {
		switch parent.Kind {
		case ast.KindJsxAttribute:
			if parent.Name() == node && parent.Parent != nil {
				return deprecationForJsxAttribute(ctx, parent.Parent.Parent, node.Text())
			}

		case ast.KindBindingElement:
			// `const { b } = a` is a read of `a.b`. The binding's own symbol is the new local, whose
			// declaration carries no tag, so the tag has to be looked up as a property of the type
			// being destructured. The pattern's parent is the declaration that holds the initializer.
			if reason, found := deprecationOnDestructuredProperty(ctx, parent, node); found {
				return reason, true
			}

		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			// `({ deprecatedKey: 1 })` reports on the key when the CONTEXTUAL object type marks it
			// deprecated, and `({ deprecatedBinding })` reports on the shorthand's value binding.
			// Both are asked because the shorthand is one identifier playing both roles.
			owner := parent.Parent
			if owner != nil {
				objectType := ctx.TypeChecker.GetTypeAtLocation(owner)
				if objectType != nil {
					if reason, found := deprecationOnSymbol(ctx,
						ctx.TypeChecker.GetPropertyOfType(objectType, node.Text())); found {
						return reason, true
					}
				}
			}
			propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(node)
			if reason, found := deprecationInAliasChain(ctx, propertySymbol, true); found {
				return reason, true
			}
			if reason, found := deprecationOnSymbol(ctx, propertySymbol); found {
				return reason, true
			}
			// `const bar = { test }` writes one identifier that both names the new key and reads the
			// outer binding, and the symbol at that location is the KEY. The tag is on the binding,
			// so reaching it needs the checker's own shorthand accessor; without this the case is
			// silent while the equivalent longhand `{ test: test }` reports.
			if parent.Kind == ast.KindShorthandPropertyAssignment {
				var valueDeclaration *ast.Node
				if propertySymbol != nil {
					valueDeclaration = propertySymbol.ValueDeclaration
				}
				if valueDeclaration != nil {
					return deprecationOnSymbol(ctx,
						ctx.TypeChecker.GetShorthandAssignmentValueSymbol(valueDeclaration))
				}
			}
			return "", false
		}
	}

	return deprecationInAliasChain(ctx, ctx.TypeChecker.GetSymbolAtLocation(node), true)
}

// checkDeprecatedElementAccess is upstream's `checkMemberExpression`, the computed-access arm:
// `a['deprecatedKey']` names the property through a literal rather than an identifier, so the
// property has to be looked up on the object's type by that literal's value.
//
// A non-literal index, `a[k]`, is silent in both: nothing here can know what `k` is.
func checkDeprecatedElementAccess(
	ctx rule.Context,
	node *ast.Node,
	settings NoDeprecatedOptions,
	report func(*ast.Node, string, string),
) {
	access := node.AsElementAccessExpression()
	argument := access.ArgumentExpression
	if argument == nil {
		return
	}

	// Upstream asks whether the argument's TYPE is a literal type, not whether its syntax is a
	// literal. The difference is the whole arm: `const key = 'b'; a[key]` has a non-literal argument
	// whose type is the literal `'b'`, and reading syntax alone is silent on it. Measured: seven of
	// upstream's invalid cases index through such a variable, and one indexes through a `as const`.
	//
	// The negative control is `const complex = Symbol() as any; a[complex]`, whose type is not a
	// literal and which must stay silent; asking the type keeps it silent while a rule that simply
	// resolved the argument would not.
	propertyName, named := literalPropertyName(ctx, argument)
	if !named {
		return
	}

	objectType := ctx.TypeChecker.GetTypeAtLocation(access.Expression)
	if objectType == nil {
		return
	}
	reason, found := deprecationOnSymbol(ctx, ctx.TypeChecker.GetPropertyOfType(objectType, propertyName))
	if !found {
		return
	}
	if deprecationTypeIsAllowed(ctx, objectType, settings) {
		return
	}
	report(argument, propertyName, reason)
}

// deprecationTypeIsAllowed asks whether a TYPE is on the allowlist.
//
// The early return on an empty allowlist is what keeps the default configuration from paying for
// this at all: with no `allow` key, which is every configuration that does not name one, the rule
// never walks a specifier.
func deprecationTypeIsAllowed(ctx rule.Context, subject *checker.Type, settings NoDeprecatedOptions) bool {
	if subject == nil || (len(settings.Allow) == 0 && len(settings.AllowInline) == 0) {
		return false
	}
	return type_checking.TypeMatchesSomeSpecifier(subject, settings.Allow, settings.AllowInline, ctx.Program)
}

// deprecationIsAllowed asks whether the thing at a node is on the allowlist.
//
// Upstream asks twice, `typeMatchesSomeSpecifier` and then `valueMatchesSomeSpecifier`, because a
// deprecated VALUE and a deprecated TYPE are different questions and the allowlist answers both.
// The value form resolves the node's own symbol to its declaration and matches on where that lives,
// which is what makes `{ from: 'package', name: 'exists', package: 'fs' }` allow the imported
// binding rather than only the type it happens to have.
func deprecationIsAllowed(ctx rule.Context, node *ast.Node, settings NoDeprecatedOptions) bool {
	if len(settings.Allow) == 0 && len(settings.AllowInline) == 0 {
		return false
	}
	subject := ctx.TypeChecker.GetTypeAtLocation(node)
	if deprecationTypeIsAllowed(ctx, subject, settings) {
		return true
	}
	return deprecationValueIsAllowed(ctx, node, settings)
}

// deprecationValueIsAllowed is upstream's `valueMatchesSomeSpecifier`.
//
// Separate from the type question because a deprecated BINDING and a deprecated TYPE are different
// things to allow. `{ from: 'package', name: 'exists', package: 'fs' }` is a statement about where
// the imported `exists` came from, which the type of `exists` alone cannot answer.
func deprecationValueIsAllowed(ctx rule.Context, node *ast.Node, settings NoDeprecatedOptions) bool {
	return type_checking.ValueMatchesSomeSpecifier(
		node, settings.Allow, settings.AllowInline, ctx.Program,
		ctx.TypeChecker.GetTypeAtLocation(node))
}

// literalPropertyName is upstream's `propertyType.isLiteral()` test: the name a computed access
// selects, when the checker can pin it to one.
//
// The question is about the TYPE rather than the syntax. `a['b']`, `const key = 'b'; a[key]` and
// `a[key as const]` all have the literal type `'b'` while only the first has literal syntax, and
// upstream reports on all three. `a[someString]` and `a[Symbol() as any]` do not narrow to a
// literal and stay silent, which is what keeps this from reporting on every computed access.
//
// A numeric literal is stringified, because that is how a property is named: `a[1]` selects `"1"`.
func literalPropertyName(ctx rule.Context, argument *ast.Node) (string, bool) {
	argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
	if argumentType == nil {
		return "", false
	}
	if argumentType.IsStringLiteral() {
		if value, ok := argumentType.AsLiteralType().Value().(string); ok {
			return value, true
		}
		return "", false
	}
	if argumentType.IsNumberLiteral() {
		// The numeric value's own String() is the property name the checker would use, and it is
		// what upstream's `String(propertyType.value)` produces.
		return fmt.Sprint(argumentType.AsLiteralType().Value()), true
	}
	return "", false
}

// deprecationOnDestructuredProperty answers about `const { b } = a`: the tag is on `a`'s property
// `b`, not on the new local the pattern declares.
//
// The type to look the property up on is the type of whatever the pattern destructures, which is
// the initializer of the declaration the pattern belongs to, or the pattern's own type when it is
// nested inside another pattern.
func deprecationOnDestructuredProperty(ctx rule.Context, element *ast.Node, node *ast.Node) (string, bool) {
	pattern := element.Parent
	if pattern == nil {
		return "", false
	}
	sourceType := ctx.TypeChecker.GetTypeAtLocation(pattern)
	if sourceType == nil {
		return "", false
	}
	return deprecationOnSymbol(ctx, ctx.TypeChecker.GetPropertyOfType(sourceType, node.Text()))
}
