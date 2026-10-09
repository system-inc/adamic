package react

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noArrowFunctionLifecycleMessage builds the finding for one lifecycle method written as an arrow.
//
// One message id with the property name interpolated, so an id assertion cannot see anything the
// format string does and a fixture has to assert the rendered text as well.
func noArrowFunctionLifecycleMessage(propertyName string) rule.Message {
	return rule.Message{
		Id: "lifecycle",
		Description: fmt.Sprintf(
			"%s is a React lifecycle method, and should not be an arrow function or in a class "+
				"field. Use an instance method instead.",
			propertyName,
		),
	}
}

// noArrowFunctionLifecycleInstanceMethods is upstream's instance list, verbatim.
//
// Order does not matter; membership does. Copied from `lib/util/lifecycleMethods.js` rather than
// retyped from the documentation, because the two unsafe-prefixed spellings are easy to drop and a
// missing name is a silent false negative.
var noArrowFunctionLifecycleInstanceMethods = map[string]bool{
	"getDefaultProps":                  true,
	"getInitialState":                  true,
	"getChildContext":                  true,
	"componentWillMount":               true,
	"UNSAFE_componentWillMount":        true,
	"componentDidMount":                true,
	"componentWillReceiveProps":        true,
	"UNSAFE_componentWillReceiveProps": true,
	"shouldComponentUpdate":            true,
	"componentWillUpdate":              true,
	"UNSAFE_componentWillUpdate":       true,
	"getSnapshotBeforeUpdate":          true,
	"componentDidUpdate":               true,
	"componentDidCatch":                true,
	"componentWillUnmount":             true,
	"render":                           true,
}

// noArrowFunctionLifecycleStaticMethods is upstream's static list, verbatim.
//
// One name, and the split is load-bearing rather than tidy. A STATIC class field is checked against
// this list alone, so `static render = () => {}` is CLEAN and `static getDerivedStateFromProps =
// () => {}` reports. An INSTANCE field is checked against the other list alone, so the two verdicts
// invert. Both measured on the installed build; upstream's corpus states the reporting half only.
var noArrowFunctionLifecycleStaticMethods = map[string]bool{
	"getDerivedStateFromProps": true,
}

// noArrowFunctionLifecycleIsEs5ComponentCall reports whether a call is `createReactClass(...)`.
//
// Deliberately NOT `react.IsEs5ComponentCall`, and this is a measured divergence rather than a
// preference. That shelf helper also accepts the name `createClass`, both bare and on the React
// object, and its doc comment presents the wider set as a correctness feature. Upstream resolves
// the factory name through `getCreateClass`, which defaults to `createReactClass` and nothing else.
// Measured on the installed build: `createClass({render: () => <div/>})` and
// `React.createClass({...})` are both CLEAN, and only `createReactClass` reports. Reaching for the
// shelf here would report on two shapes upstream ignores.
//
// The member form is kept because upstream keeps it: `React.createReactClass` would match its
// pragma-and-name test. Nothing in the corpus writes it.
func noArrowFunctionLifecycleIsEs5ComponentCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := node.AsCallExpression().Expression
	for callee != nil && callee.Kind == ast.KindParenthesizedExpression {
		callee = callee.AsParenthesizedExpression().Expression
	}
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "createReactClass"
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindIdentifier ||
			access.Expression.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "createReactClass"
	}
	return false
}

// noArrowFunctionLifecyclePropertyName returns the plain name a property is written with.
//
// Upstream's `getPropertyName` returns nothing useful for a computed key, and the rule's membership
// test then misses, so `['render'] = () => {}` is CLEAN. Measured. The second return separates "no
// readable name" from a name that happens to be empty.
//
// A string-literal key is a readable name upstream, since its `key.value` is the string. Both
// spellings of a quoted key are therefore treated the same as the bare identifier.
//
// The kind switch is what excludes a computed key, and it does so by omission rather than by an
// explicit test. An earlier revision carried a separate `isComputed` flag beside it; a mutant
// removing that flag survived every fixture, and the reason is that the two guards are the same
// guard: a computed key parses as `KindComputedPropertyName`, which is in neither arm. The flag was
// removed rather than kept, because no input can distinguish the two versions.
//
// The omission is also crash protection, which is the half a survival verdict cannot see.
// `Text()` panics on a `ComputedPropertyName` rather than returning anything, and the walk recovers
// per FILE, so calling it would cost every rule in the package every finding in that file. Probed
// directly: a rule reading `Name().Text()` on `['render'] = () => <div />` panics with
// `Unhandled case in Node.Text`. Widening this switch is therefore not a false-positive risk, it is
// an outage.
func noArrowFunctionLifecyclePropertyName(key *ast.Node) (string, bool) {
	if key == nil {
		return "", false
	}
	switch key.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return key.Text(), true
	}
	return "", false
}

// noArrowFunctionLifecycleParameterNames renders an arrow's parameters for the repair.
//
// Upstream maps each parameter to `p.name`, which is the identifier's name and is UNDEFINED for
// every destructuring or defaulted parameter, so its own fixer writes the literal text `undefined`
// into the source for those. That is a defect rather than a decision, and it is why this port
// declines the fix on any parameter that is not a plain identifier. See the rule doc comment.
func noArrowFunctionLifecycleParameterNames(parameters *ast.NodeList) (string, bool) {
	if parameters == nil {
		return "", true
	}
	names := make([]string, 0, len(parameters.Nodes))
	for _, parameter := range parameters.Nodes {
		declaration := parameter.AsParameterDeclaration()
		name := declaration.Name()
		// A parameter that is not a plain identifier, or that carries a type, a default, a
		// question mark or a rest marker, cannot be rendered by a name-only join without losing
		// what it says. Upstream loses it silently; this port declines the repair instead.
		if name == nil || name.Kind != ast.KindIdentifier ||
			declaration.Initializer != nil || declaration.Type != nil ||
			declaration.DotDotDotToken != nil || declaration.QuestionToken != nil {
			return "", false
		}
		names = append(names, name.Text())
	}
	return strings.Join(names, ", "), true
}

// NoArrowFunctionLifecycle flags a React lifecycle method written as an arrow function.
//
//	valid:   class H extends React.Component { render() { return <div />; } }
//	valid:   class H extends React.Component { onClick = () => {}; }      not a lifecycle name
//	valid:   class H extends React.Component { static render = () => {}; } static uses the other list
//	valid:   class H extends Foo { render = () => <div />; }              not a component
//	invalid: class H extends React.Component { render = () => <div />; }
//	invalid: var H = createReactClass({ render: () => <div /> });
//
// Ported from `react/no-arrow-function-lifecycle` in `eslint-plugin-react`. No options, one
// message, and a fixer that rewrites the arrow into a method.
//
// All 73 corpus cases were replayed against the installed build, version 7.37.5, through the ESLint
// Linter API before any code was written, and it agreed on 72. The one disagreement is a harness
// artifact rather than a rule difference: upstream's `valid 36` writes a TypeScript type annotation
// and the default parser cannot parse it, so the run reported a parse error rather than a verdict.
// Our parser reads it, and it is carried as a fixture at the clean verdict its corpus states.
//
// # Component detection is narrower than the shelf helper
//
// The ES5 factory is `createReactClass` and nothing else, which the rule-local predicate above
// reproduces. `react.IsEs5ComponentCall` also accepts `createClass`, and using it would report on
// two shapes the installed build calls clean. Measured both ways.
//
// The ES6 side uses `react.IsEs6ComponentClass`, which agrees with upstream on every shape the
// corpus and the probes write. Its documented parenthesis skip is not a divergence here:
// `class H extends (React.Component)` REPORTS on the installed build, so skipping is what
// reproduces upstream rather than what improves on it. Measured.
//
// # The static and instance lists are separate, and the split inverts two verdicts
//
// A static class field is checked against the static list alone and an instance field against the
// instance list alone, so:
//
//	static getDerivedStateFromProps = () => {}   REPORTS
//	getDerivedStateFromProps = () => {}          CLEAN, not on the instance list
//	static render = () => {}                     CLEAN, not on the static list
//	render = () => {}                            REPORTS
//
// All four measured. Upstream's corpus writes only the reporting halves, so a port that used one
// combined list would pass every imported case and report two shapes upstream ignores.
//
// An ES5 object property is never static in this sense, so it always uses the instance list. That
// is why `createReactClass({getDerivedStateFromProps: () => {}})` is CLEAN while the class field of
// the same name reports. Measured.
//
// # This port declines the fix on parameters upstream renders wrongly
//
// Upstream's fixer builds the parameter list by mapping each parameter to `p.name`, which is
// undefined for anything that is not a plain identifier, so a destructured or defaulted parameter
// is written into the repaired source as the literal text `undefined`. That is a defect that
// silently deletes what the parameter said, and applying it unattended would corrupt the file.
//
// This port reports those cases and declines to repair them, which is the subset that can be shown
// correct. Reporting without a fix is recoverable; a wrong fix applied unattended is not. The
// decline also covers a typed parameter, which upstream's corpus cannot contain and which our tree
// is full of: rendering `properties: Properties` as `properties` would drop the annotation exactly
// the way the brief's `no-undef-init` incident dropped one.
//
// # The concise-body repair, and what it has to preserve
//
// An arrow with an expression body becomes a block returning that expression, and the comments on
// either side of the expression have to survive into the block. Upstream widens the replaced range
// to cover them, and adds one character when the body is an object literal to account for the
// wrapping parenthesis. Both behaviours are reproduced, and both are asserted by upstream's own
// output vectors.
var NoArrowFunctionLifecycle = rule.Rule{
	Name: "react/no-arrow-function-lifecycle",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		sourceText := ctx.SourceFile.Text()

		textOf := func(node *ast.Node) string {
			span := rule.TokenRange(ctx.SourceFile, node)
			if span.Pos() < 0 || span.End() > len(sourceText) || span.Pos() > span.End() {
				return ""
			}
			return sourceText[span.Pos():span.End()]
		}

		// check reports one property whose value is an arrow and whose name is a lifecycle method.
		//
		// `isClassField` decides both which list applies and what the repaired head looks like: a
		// class field becomes `name(params) ` and an object property becomes `name: function(params) `.
		check := func(property *ast.Node, key *ast.Node, value *ast.Node, isStatic bool, isClassField bool) {
			if value == nil || value.Kind != ast.KindArrowFunction {
				return
			}
			propertyName, readable := noArrowFunctionLifecyclePropertyName(key)
			if !readable {
				return
			}

			// The two lists are consulted separately. A static class field never falls back to the
			// instance list and an instance field never reaches the static one.
			lifecycle := noArrowFunctionLifecycleInstanceMethods
			if isStatic && isClassField {
				lifecycle = noArrowFunctionLifecycleStaticMethods
			}
			if !lifecycle[propertyName] {
				return
			}

			arrow := value.AsArrowFunction()
			body := arrow.Body
			if body == nil {
				ctx.ReportNode(property, noArrowFunctionLifecycleMessage(propertyName))
				return
			}

			parameters, renderable := noArrowFunctionLifecycleParameterNames(arrow.Parameters)
			if !renderable {
				// Reported without a repair. See the doc comment: upstream would write the literal
				// text `undefined` here, and a wrong fix applied unattended is worse than none.
				ctx.ReportNode(property, noArrowFunctionLifecycleMessage(propertyName))
				return
			}

			// A return annotation is lost by the same rewrite for the same reason, and it sits
			// outside the parameter list so the check above cannot see it. The head is rebuilt as
			// `(params) `, which replaces everything between the key and the body, so
			// `componentDidMount = (): void => {}` would be repaired to `componentDidMount() {}`
			// and the `: void` would be gone. Upstream never meets this because its corpus is
			// JavaScript. Declining matches what this rule already does for a typed parameter:
			// report the arrow, leave the repair to a human who can see the type.
			if arrow.Type != nil {
				ctx.ReportNode(property, noArrowFunctionLifecycleMessage(propertyName))
				return
			}

			head := fmt.Sprintf("(%s) ", parameters)
			if !isClassField {
				head = fmt.Sprintf(": function(%s) ", parameters)
			}

			// The head runs from the end of the key to the start of the body, which swallows the
			// parameter list, the arrow and any type annotation between them. Comments before the
			// body belong to the body's replacement rather than to the head, so the head stops at
			// the first of them.
			//
			// "Before the body" means after the `=>`, which is upstream's `getCommentsBefore(body)`:
			// the comments between the body and the token preceding it. The scan once started at the
			// key instead, so `render = /* c */ () => {...}` stopped the head at the comment and left
			// the old parameter list and arrow standing behind it, writing
			// `render() /* c */ () => {...}`, which does not parse.
			//
			// A comment in the head itself, before the `=>`, is declined rather than dropped.
			// Upstream's head range swallows it, so its output parses and the comment is gone; a fix
			// runs unattended, and deleting what someone wrote is not a spelling change.
			bodyStart := rule.TokenRange(ctx.SourceFile, body).Pos()
			if arrow.EqualsGreaterThanToken == nil {
				ctx.ReportNode(property, noArrowFunctionLifecycleMessage(propertyName))
				return
			}
			arrowEnd := rule.TokenRange(ctx.SourceFile, arrow.EqualsGreaterThanToken).End()
			if len(commentsBetween(sourceText, key.End(), arrowEnd)) > 0 {
				ctx.ReportNode(property, noArrowFunctionLifecycleMessage(propertyName))
				return
			}
			leading := commentsBetween(sourceText, arrowEnd, bodyStart)
			headEnd := bodyStart
			if len(leading) > 0 {
				headEnd = leading[0].start
			}
			headRange := core.NewTextRange(key.End(), headEnd)

			if body.Kind == ast.KindBlock {
				ctx.ReportNodeWithFixes(property, noArrowFunctionLifecycleMessage(propertyName),
					rule.ReplaceRange(headRange, head))
				return
			}

			// A concise body becomes a block returning the expression, carrying whatever comments
			// sat on either side of it into the block so nothing is lost.
			// Upstream's body node is the expression itself, because its parser folds the
			// wrapping parentheses away. Ours keeps them, so an object body arrives as a
			// parenthesized expression and the emitted text has to be the object inside it.
			// Upstream then extends its own range by one character to swallow the closing paren,
			// which is the same adjustment expressed from the other side. Corpus case `invalid 35`
			// asserts the result: the repair writes the object with no parentheses around it.
			bodyText := body
			bodyEnd := body.End()
			if body.Kind == ast.KindParenthesizedExpression {
				bodyText = stylePropObjectUnwrapParentheses(body)
				if bodyText == nil {
					bodyText = body
				}
			}
			// The comment scan runs to the end of the LINE holding the body rather than to the end
			// of the property. A trailing comment written after a concise body sits outside the
			// property's own span, since the property ends at the expression, so bounding the scan
			// by the property would miss it. Corpus case `invalid 34` asserts the comment survives.
			trailing := commentsBetween(sourceText, bodyEnd, endOfLineAt(sourceText, bodyEnd))
			replacedEnd := bodyEnd
			if len(trailing) > 0 {
				replacedEnd = trailing[len(trailing)-1].end
			}
			// A trailing semicolon inside the property is consumed by the replacement, because the
			// block form supplies its own.
			for replacedEnd < len(sourceText) && sourceText[replacedEnd] == ';' {
				replacedEnd++
				break
			}

			leadingText := ""
			for _, comment := range leading {
				leadingText += sourceText[comment.start:comment.end]
			}
			trailingText := ""
			for _, comment := range trailing {
				trailingText += sourceText[comment.start:comment.end]
			}

			ctx.ReportNodeWithFixes(property, noArrowFunctionLifecycleMessage(propertyName),
				rule.ReplaceRange(headRange, head),
				rule.ReplaceRange(
					core.NewTextRange(leadingStart(leading, bodyStart), replacedEnd),
					fmt.Sprintf("{ return %s%s%s; }", leadingText, textOf(bodyText), trailingText),
				),
			)
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if !noArrowFunctionLifecycleIsEs5ComponentCall(node) {
					return
				}
				arguments := node.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return
				}
				object := arguments.Nodes[0]
				if object.Kind != ast.KindObjectLiteralExpression {
					return
				}
				for _, property := range object.AsObjectLiteralExpression().Properties.Nodes {
					if property.Kind != ast.KindPropertyAssignment {
						continue
					}
					assignment := property.AsPropertyAssignment()
					name := assignment.Name()
					check(property, name, assignment.Initializer, false, false)
				}
			},

			ast.KindClassDeclaration: func(node *ast.Node) {
				noArrowFunctionLifecycleCheckClass(node, check)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				noArrowFunctionLifecycleCheckClass(node, check)
			},
		}
	},
}

// noArrowFunctionLifecycleCheckClass walks a class's property declarations.
//
// Only a property declaration whose initializer is an arrow can report; a real method declaration
// is already what the rule asks for. `react.IsEs6ComponentClass` gates the whole class, and it
// agrees with upstream on every shape measured, including the parenthesized heritage that its doc
// comment flags as a divergence elsewhere.
func noArrowFunctionLifecycleCheckClass(
	node *ast.Node,
	check func(property *ast.Node, key *ast.Node, value *ast.Node, isStatic bool, isClassField bool),
) {
	if !react.IsEs6ComponentClass(node) {
		return
	}
	members := node.ClassLikeData().Members
	if members == nil {
		return
	}
	for _, member := range members.Nodes {
		if member.Kind != ast.KindPropertyDeclaration {
			continue
		}
		declaration := member.AsPropertyDeclaration()
		name := declaration.Name()
		check(member, name, declaration.Initializer, ast.HasStaticModifier(member), true)
	}
}

// noArrowFunctionLifecycleComment is one comment's span in the source.
type noArrowFunctionLifecycleComment struct {
	start int
	end   int
}

// commentsBetween returns the comments sitting strictly between two offsets.
//
// Written as a scan rather than through the shelf's comment reader because the question is about a
// span inside one expression rather than about a node's leading trivia, and the reader is keyed by
// node. The scan skips string and template content so a `//` inside a literal is not mistaken for a
// comment.
func commentsBetween(sourceText string, from int, to int) []noArrowFunctionLifecycleComment {
	if from < 0 || to > len(sourceText) || from >= to {
		return nil
	}
	var found []noArrowFunctionLifecycleComment
	index := from
	for index < to-1 {
		if sourceText[index] == '/' && sourceText[index+1] == '/' {
			start := index
			for index < to && sourceText[index] != '\n' {
				index++
			}
			found = append(found, noArrowFunctionLifecycleComment{start: start, end: index})
			continue
		}
		if sourceText[index] == '/' && sourceText[index+1] == '*' {
			start := index
			index += 2
			for index < to-1 && !(sourceText[index] == '*' && sourceText[index+1] == '/') {
				index++
			}
			if index < to-1 {
				index += 2
			}
			found = append(found, noArrowFunctionLifecycleComment{start: start, end: index})
			continue
		}
		index++
	}
	return found
}

// leadingStart returns where the replaced body span begins.
func leadingStart(leading []noArrowFunctionLifecycleComment, bodyStart int) int {
	if len(leading) > 0 {
		return leading[0].start
	}
	return bodyStart
}

// endOfLineAt returns the offset of the newline ending the line that contains an offset.
//
// A trailing comment written after a concise arrow body sits OUTSIDE the property's own span,
// because the property ends where the expression does. Bounding the comment scan by the property
// therefore misses it, which cost one of upstream's own fix vectors. The line is the right bound:
// upstream reaches the same comment through `getCommentsAfter`, which stops at the next token.
func endOfLineAt(sourceText string, from int) int {
	if from < 0 {
		return 0
	}
	for index := from; index < len(sourceText); index++ {
		if sourceText[index] == '\n' {
			return index
		}
	}
	return len(sourceText)
}
