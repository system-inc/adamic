package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoSetState = rule.Message{
	Id:          "noSetState",
	Description: "Do not use setState",
}

// NoSetState reports every `this.setState(...)` call written inside a React component.
//
//	valid:   this.setState({}) inside a plain function that is not a component
//	valid:   this.setState({}) inside a class with no React heritage
//	valid:   this.someHandler = this.setState, a reference rather than a call
//	invalid: this.setState({}) inside a class extending React.Component
//	invalid: this.setState({}) inside a createReactClass method
//
// The rule exists for codebases keeping component state somewhere else, a store or a hook, where a
// direct setState is the thing that escapes that discipline. Upstream marks it `recommended: false`,
// so it is a house choice rather than a correctness rule.
//
// # Component membership is the gate, and it is a real question here
//
// `class Hello { m(){ this.setState({}); } }` with no heritage is silent, measured against the
// installed build on 2026-08-27. That is worth stating because the neighbouring `no-is-mounted`
// looks identical and asks something different: it walks ancestors for a method-like node and never
// asks what the enclosing class extends, so it reports on exactly that shape. The shelf's
// `EnclosingComponent` doc comment warns about this in the opposite direction. This rule genuinely
// asks membership, so the walk below asks it.
//
// # Which factory spellings count
//
// Only `createReactClass`. Measured: a bare `createClass({...})` and a namespaced
// `React.createClass({...})` are both silent, so this reuses `isCreateReactClassCall` from
// `no-multi-comp` rather than the shelf's `IsEs5ComponentCall`, which accepts all three. Upstream's
// corpus writes `createReactClass` alone and cannot see the difference.
//
// # Every usage reports, not the component
//
// Upstream collects usages per component and reports each one, so a component calling setState
// twice produces two findings. Measured, and reproduced.
//
// # No fix
//
// Removing a setState call means deciding where the state goes instead, which is a refactor.
// Upstream ships no fixer either.
var NoSetState = rule.Rule{
	Name: "react/no-set-state",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := setStateCalleeOf(node)
				if callee == nil {
					return
				}
				if enclosingComponentOf(node) == nil {
					return
				}
				// Upstream reports on the callee, `this.setState`, rather than on the whole call,
				// so the finding stops at the member access and excludes the argument list.
				ctx.ReportNode(callee, messageNoSetState)
			},
		}
	},
}

// setStateCalleeOf returns the `this.setState` callee of a call, or nil.
//
// Upstream tests three things in one condition: the callee is a member expression, its object is a
// `this` expression, and its property is NAMED setState. That last one is why a computed
// `this["setState"]({})` is silent, since upstream reads `callee.property.name` and a computed
// member has no name there. Measured silent against the installed build on 2026-08-27, which is a
// shape the corpus does not write.
//
// Parentheses are unwrapped because our parser keeps them where upstream's folds them away, so
// `(this).setState({})` reaches upstream's member branch as a plain `this`.
func setStateCalleeOf(node *ast.Node) *ast.Node {
	callee := skipParenthesesOptional(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	access := callee.AsPropertyAccessExpression()

	object := skipParenthesesOptional(access.Expression)
	if object == nil || object.Kind != ast.KindThisKeyword {
		return nil
	}

	name := access.Name()
	if name == nil {
		return nil
	}

	// Upstream reads `callee.property.name`, which for a private name node is the bare identifier
	// with no hash, so `this.#setState({})` REPORTS upstream. Measured against the installed build
	// on 2026-08-27, beside a plain control that also reports.
	//
	// Our parser spells the same node as `KindPrivateIdentifier` whose `Text()` carries the hash,
	// so a naive comparison declines it and a naive port silently narrows the rule. The hash is
	// trimmed rather than compared against a hashed literal, so the two spellings meet at
	// upstream's own value. This is the doubled-hash hazard arriving from the other direction: the
	// text already carries the sigil and the question is what upstream compared, not what reads
	// naturally.
	//
	// An element access, `this["setState"]({})`, never reaches here at all: it is a different node
	// kind and the guard above declines it, which reproduces upstream declining it for its own
	// reason, that a computed member has no `property.name`. Measured silent both places.
	var propertyName string
	switch name.Kind {
	case ast.KindIdentifier:
		propertyName = name.Text()
	case ast.KindPrivateIdentifier:
		propertyName = strings.TrimPrefix(name.Text(), "#")
	default:
		return nil
	}
	if propertyName != "setState" {
		return nil
	}
	return callee
}

// enclosingComponentOf returns the nearest ancestor that is a React component, or nil.
//
// Both eras are searched in one walk: a class extending a React base, and an object literal handed
// to `createReactClass`. Deliberately not the shelf's `EnclosingComponent`, which reaches
// `IsEs5ComponentCall` and therefore accepts two factory spellings upstream rejects here, and
// `IsEs6ComponentClass`, which declines a parenthesized heritage receiver. Both narrowings are
// measured in `no-multi-comp` beside the predicates this reuses.
func enclosingComponentOf(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if isComponentClass(current) {
			return current
		}
		// The factory's object literal is the component, and the call is its parent, so the walk
		// tests the call and reports membership from there.
		if current.Kind == ast.KindCallExpression && isCreateReactClassCall(current) {
			return current
		}
	}
	return nil
}
