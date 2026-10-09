// Package react answers the questions a rule asks about React code that are not questions about
// JSX: is this call constructing an element, is this a component, which component encloses this.
//
// It is a sibling of `internal/utilities/jsx/` rather than part of it, because none of these touch a
// JSX node. `createElement` is an ordinary call expression and a class component is an ordinary
// class; a rule can ask both about a file with no JSX in it at all.
//
// Built on measured demand rather than on one caller's word. In oxc's own react rules,
// `is_create_element_call` has nine callers and the component predicates have twelve, and oxc
// factored both into `utils/react.rs` for the same reason this package exists. Two independent
// research passes on `#rulesreact` reached the component predicate from the rule sources without
// any brief mentioning it, which is what separates a real shared surface from a guess about one.
package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// reactPragma is the object name the namespaced spellings are checked against.
//
// Upstream calls this the pragma because it is configurable there. It is fixed here until a rule
// needs it otherwise, and that is worth stating: a rule that needs a different pragma should make
// this an option rather than reach past it.
const reactPragma = "React"

// IsCreateElementCall reports whether a call constructs a React element.
//
// Three things about this are deliberate and each would be easy to get wrong in the safe-looking
// direction.
//
// **The object is not required to be React.** A bare `createElement(...)` counts, and so does
// `Preact.createElement(...)`. Requiring the React namespace would silence the rule on every
// codebase that imports the function directly, which is the common modern spelling.
//
// **`document.createElement` is rejected by name.** It is the one call that shares the property
// name and constructs a DOM node rather than a React element, and upstream special-cases it
// exactly this way.
//
// **A computed member counts.** `React['createElement'](...)` is the same call written differently,
// and a check that read only static members would miss it while looking complete.
//
// **Not every upstream rule asks this question the same way, so check before reaching for this.**
// oxc has one shared `is_create_element_call` and also rules that test the callee themselves, and
// the two disagree in three directions at once. `no_danger_with_children.rs:76` destructures a
// `StaticMemberExpression` and returns early otherwise, so it never sees a bare `createElement(...)`,
// does not reject `document.`, and does not accept a computed member.
//
// A port of that rule calling this helper starts reporting where upstream is silent, and its own
// corpus cannot catch the difference because none of its fixtures writes a bare call. Verified at
// both sources rather than inferred, after `@system_cohere_format` found it while researching the
// react ports.
//
// This function matches `no-children-prop`, which is the shape oxc factored out. A rule that reads
// its callee inline upstream should read it inline here too, or this needs a parameter rather than
// a second copy. Nobody has read the remaining seven callers, so that decision is deliberately open
// rather than guessed.
func IsCreateElementCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "createElement"

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if isIdentifierNamed(access.Expression, "document") {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "createElement"

	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if isIdentifierNamed(access.Expression, "document") {
			return false
		}
		argument := access.ArgumentExpression
		return argument != nil && argument.Kind == ast.KindStringLiteral &&
			argument.Text() == "createElement"
	}
	return false
}

// IsEs5ComponentCall reports whether a call creates a component the old way.
//
// Both spellings count: `React.createClass(...)` and a bare `createClass(...)`, the second being
// what a file importing the function directly writes.
func IsEs5ComponentCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return isCreateClassName(callee.Text())

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if !isIdentifierNamed(access.Expression, reactPragma) {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && isCreateClassName(name.Text())
	}
	return false
}

// isCreateClassName accepts both names the factory has been called.
//
// `createReactClass` is the standalone package and `createClass` is the older method on the React
// object. Rules meet both, and treating one as the only spelling silences the rule on whichever
// half of the ecosystem writes the other.
func isCreateClassName(name string) bool {
	return name == "createReactClass" || name == "createClass"
}

// IsEs6ComponentClass reports whether a class extends a React component base.
//
// A class expression counts as well as a declaration, because `const Thing = class extends
// React.Component {}` is the same thing written where a declaration will not fit.
//
// The extends clause is what decides, and its absence is the discriminating case rather than an
// edge one: `class Hello { }` with no heritage is 2 of the 8 passing fixtures upstream ships for
// `no-direct-mutation-state`, which is to say the gate exists mostly to stay silent on plain
// classes.
func IsEs6ComponentClass(node *ast.Node) bool {
	if node == nil {
		return false
	}

	var heritage *ast.NodeList
	switch node.Kind {
	case ast.KindClassDeclaration:
		heritage = node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		heritage = node.AsClassExpression().HeritageClauses
	default:
		return false
	}
	if heritage == nil {
		return false
	}

	for _, clause := range heritage.Nodes {
		if clause.Kind != ast.KindHeritageClause {
			continue
		}
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typeNode := range types.Nodes {
			if typeNode.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			if isComponentBase(typeNode.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

// isComponentBase reports whether an extends target names a React component base class.
//
// `React.Component`, `React.PureComponent`, and the bare `Component` and `PureComponent` a file
// importing them directly writes. The namespaced form checks the object is React specifically,
// because `Foo.Component` is somebody else's class.
func isComponentBase(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return isComponentBaseName(expression.Text())

	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		if !isIdentifierNamed(access.Expression, reactPragma) {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && isComponentBaseName(name.Text())
	}
	return false
}

// isComponentBaseName accepts the two base classes a component may extend.
func isComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

// EnclosingComponent returns the nearest ancestor that is a component by either definition, or nil.
//
// Both definitions are searched in one walk rather than in two passes, because a rule asking "am I
// inside a component" does not care which era wrote it, and two walks would let the answer depend
// on which ran first.
//
// The node itself is considered, so a rule that has already matched a class can ask this without
// stepping to the parent first.
//
// **Confirm the rule you are porting actually asks about component membership before reaching for
// this, because several react rules test syntactic containment instead and look identical from
// here.** `no_is_mounted.rs:69` walks ancestors for an `ObjectProperty` or a `MethodDefinition` and
// stops there: it never asks what the enclosing object was constructed by, and never asks what the
// enclosing class extends. So `class Whatever { m() { this.isMounted(); } }` with no heritage
// reports upstream, and `IsEs6ComponentClass` answers false for exactly that shape. A port reaching
// for this narrows the rule and drops findings.
//
// The corpus will not catch it. All three of that rule's fail cases are written inside
// `createReactClass`, so a port narrowed to real components passes every one of them. Verified at
// the source after `@system_cohere_format` found it, and it is the second time a react helper has
// been correct for the rule that motivated it and wrong for a neighbour that looks the same.
//
// The distinction to hold: *is this inside a component* and *is this inside a method-like thing*
// are different questions, and upstream asks both under names that do not distinguish them.
func EnclosingComponent(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if IsEs6ComponentClass(current) || IsEs5ComponentCall(current) {
			return current
		}
	}
	return nil
}

// isIdentifierNamed reports whether an expression is exactly this identifier.
//
// Parentheses are skipped because `(React).createElement(...)` is the same call, and a check on the
// raw node would decline it while looking correct.
func isIdentifierNamed(expression *ast.Node, name string) bool {
	expression = ast.SkipParentheses(expression)
	return expression != nil && expression.Kind == ast.KindIdentifier && expression.Text() == name
}

// IsNamespacedMember reports whether a node is `React.<name>` where name satisfies the predicate.
//
// Lifted after a census found four copies of this shape in two packages, differing only in whether
// they skipped parentheses on the receiver. Parentheses are skipped here, for the same reason
// `isIdentifierNamed` skips them: `(React).useEffect(...)` is the same call, and a check on the raw
// node declines it while looking correct.
//
// The receiver must be React specifically. `somethingElse.useThing(...)` is not a React hook, and a
// copy that accepted any namespace exempted aliases the gate still reports, which is measured in
// `consistency-no-property-alias` and is why that rule reaches for the shared predicate rather than
// writing a fifth one.
//
// The pragma is fixed rather than configurable, matching the rest of this package. A rule needing a
// different one should make it an option rather than reach past this.
func IsNamespacedMember(node *ast.Node, matches func(name string) bool) bool {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := node.AsPropertyAccessExpression()
	if !isIdentifierNamed(access.Expression, reactPragma) {
		return false
	}
	name := access.Name()
	return name != nil && name.Kind == ast.KindIdentifier && matches(name.Text())
}
