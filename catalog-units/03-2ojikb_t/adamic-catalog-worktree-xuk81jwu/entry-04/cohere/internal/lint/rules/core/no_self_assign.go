package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageSelfAssignment = rule.Message{
	Id: "selfAssignment",
	Description: "This assigns a value to itself, so it does nothing. The statement is not merely " +
		"redundant: it is almost always a typo for an assignment that was meant to happen, so the " +
		"value the author intended to write is still missing and nothing reports that it is.",
}

// NoSelfAssignOptions configures whether member expressions are compared.
//
// Upstream's schema is a single object with one boolean, `props`, defaulting to true. Our config
// layer unwraps the severity tuple before dispatch, so the decoder receives that object directly.
type NoSelfAssignOptions struct {
	// Props compares member expressions as well as names, so `a.b = a.b` and `a[b] = a[b]` report.
	// Absent means true, which is upstream's default and the reason the decoder is hand-written.
	Props bool `json:"props"`
}

// DefaultNoSelfAssignOptions is the unconfigured answer.
func DefaultNoSelfAssignOptions() NoSelfAssignOptions {
	return NoSelfAssignOptions{Props: true}
}

// DecodeNoSelfAssignOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` for two reasons, both about the default being
// true. A bare `"error"` arrives as empty input and must resolve to the default rather than error,
// and `{}` must keep `props` on, which decoding into a zero struct would silently turn off.
func DecodeNoSelfAssignOptions(raw []byte) (any, error) {
	options := DefaultNoSelfAssignOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// NoSelfAssign flags an assignment whose two sides are the same reference.
//
//	valid:   a = b
//	valid:   a.b = a.c
//	valid:   [a, b] = [b, a]
//	valid:   a.b = a.b           (with props: false)
//	invalid: a = a
//	invalid: a.b = a.b
//	invalid: a.b = a?.b
//	invalid: a.b = a["b"]
//	invalid: a[b] = a[b]
//	invalid: [a, b] = [a, b]
//	invalid: ({ a, b } = { a, b })
//
// Ported from `no-self-assign` in ESLint 10.8.1, `lib/rules/no-self-assign.js`, and held to its
// test rows by the corpus replay in `registry`.
//
// Destructuring is most of the work and the reason this is not a two-line rule. `[a, b] = [b, a]`
// is a swap and must stay silent, so the comparison is positional and elementwise rather than
// set-based, and a spread on either side truncates it because everything after a spread sits at an
// index nobody can name.
//
// The `props` option gates the member-expression half, and it is on by default. That half needs a
// real reference comparison rather than token equality, because ESLint treats `a.b`, `a?.b`, and
// `a["b"]` as the same reference and their tokens differ. Shipping the syntactic half alone would
// agree with the gate on every fixture anyone would think to write and silently miss the shape the
// option exists for, which is the mistake use-isnan shipped with.
//
// Only the four assignment operators that write the whole value are considered: `=`, `&&=`, `||=`,
// and `??=`. `a += a` doubles a, which is a different statement and not this rule's business.
var NoSelfAssign = rule.Rule{
	Name: "no-self-assign",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		props := DefaultNoSelfAssignOptions().Props
		if resolved, isNoSelfAssignOptions := rule.OptionsAs[NoSelfAssignOptions](options); isNoSelfAssignOptions {
			props = resolved.Props
		}

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil {
					return
				}
				switch binary.OperatorToken.Kind {
				case ast.KindEqualsToken,
					ast.KindAmpersandAmpersandEqualsToken,
					ast.KindBarBarEqualsToken,
					ast.KindQuestionQuestionEqualsToken:
				default:
					return
				}

				// Our parser spells a destructuring default `[a = a] = []` as an assignment, where
				// ESTree has an AssignmentPattern that upstream's AssignmentExpression listener never
				// visits. It is a fallback rather than a self-assignment: a takes the element when
				// there is one.
				if ast.IsAssignmentTarget(node) {
					return
				}

				reportSelfAssignments(ctx, binary.Left, binary.Right, props)
			},
		}
	},
}

// reportSelfAssignments walks a matched pair of assignment target and value.
//
// Recursive because a destructuring assignment is a tree of targets against a tree of values, and a
// self-assignment can sit at any depth: `[a, [b]] = [a, [b]]` is two of them.
func reportSelfAssignments(ctx rule.Context, left *ast.Node, right *ast.Node, props bool) {
	if left == nil || right == nil {
		return
	}
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	if left == nil || right == nil {
		return
	}

	switch {
	case left.Kind == ast.KindIdentifier && right.Kind == ast.KindIdentifier:
		if left.Text() == right.Text() {
			ctx.ReportNode(right, messageSelfAssignment)
		}

	case left.Kind == ast.KindArrayLiteralExpression && right.Kind == ast.KindArrayLiteralExpression:
		reportArraySelfAssignments(ctx, left, right, props)

	case left.Kind == ast.KindObjectLiteralExpression && right.Kind == ast.KindObjectLiteralExpression:
		reportObjectSelfAssignments(ctx, left, right, props)

	case left.Kind == ast.KindSpreadElement && right.Kind == ast.KindSpreadElement:
		reportSelfAssignments(ctx, left.AsSpreadElement().Expression, right.AsSpreadElement().Expression, props)

	case props && ast.IsAccessExpression(left) && ast.IsAccessExpression(right):
		if isSameReference(left, right) {
			ctx.ReportNode(right, messageSelfAssignment)
		}
	}
}

// reportArraySelfAssignments compares array destructuring positionally.
//
// Positional rather than by membership, which is what keeps `[a, b] = [b, a]` silent: a swap
// assigns each name a different value and is the reason anyone writes this shape at all.
func reportArraySelfAssignments(ctx rule.Context, left *ast.Node, right *ast.Node, props bool) {
	leftElements := left.AsArrayLiteralExpression().Elements.Nodes
	rightElements := right.AsArrayLiteralExpression().Elements.Nodes

	end := min(len(leftElements), len(rightElements))
	for index := range end {
		leftElement := ast.SkipParentheses(leftElements[index])
		rightElement := ast.SkipParentheses(rightElements[index])

		// A rest target absorbs everything remaining, so it only matches a spread that is also
		// last. `[...a] = [...a, 1]` assigns a a longer array than it had.
		if leftElement != nil && leftElement.Kind == ast.KindSpreadElement && index < len(rightElements)-1 {
			return
		}

		reportSelfAssignments(ctx, leftElement, rightElement, props)

		// After a spread on the value side, every later index is unknown, so no further position
		// can be claimed to match.
		if rightElement != nil && rightElement.Kind == ast.KindSpreadElement {
			return
		}
	}
}

// reportObjectSelfAssignments compares object destructuring by property name.
//
// By name rather than positionally, because object properties are unordered and `{ a, b } = { b, a }`
// really is a self-assignment of both.
func reportObjectSelfAssignments(ctx rule.Context, left *ast.Node, right *ast.Node, props bool) {
	leftProperties := left.AsObjectLiteralExpression().Properties.Nodes
	rightProperties := right.AsObjectLiteralExpression().Properties.Nodes
	if len(rightProperties) == 0 {
		return
	}

	// Everything up to and including the last spread can be overwritten by it, so only properties
	// after it are safe to claim. `({ a } = { ...b, a })` is a self-assignment of a; `({ a } = { a, ...b })`
	// is not, since b may carry its own a.
	start := 0
	for index := len(rightProperties) - 1; index >= 0; index-- {
		if rightProperties[index].Kind == ast.KindSpreadAssignment {
			start = index + 1
			break
		}
	}

	for _, leftProperty := range leftProperties {
		leftName, leftOk := propertyName(leftProperty)
		if !leftOk {
			continue
		}
		for _, rightProperty := range rightProperties[start:] {
			// A method or accessor is a definition rather than a value read back, so it can never
			// be the same reference as the thing it is assigned to. Two things exclude them and
			// either one alone is sufficient: propertyName answers false for any kind other than a
			// property or shorthand assignment, and propertyValue answers nil for the same set.
			//
			// That redundancy is worth naming, because it is invisible to the sweep. Mutating
			// either guard alone leaves the fixtures green, since the other still stops the case;
			// mutating both together kills 3. So a single-mutation sweep reports two unmeasured
			// lines where the truth is one measured behavior held by two lines. The distinction
			// matters: "no fixture covers this" would have sent me writing a fixture that cannot
			// exist, and "this line does nothing" would have sent me deleting a guard that does.
			rightName, rightOk := propertyName(rightProperty)
			if !rightOk || leftName != rightName {
				continue
			}
			reportSelfAssignments(ctx, propertyValue(leftProperty), propertyValue(rightProperty), props)
		}
	}
}

// propertyName reads a property's static name, reporting false when it has none.
//
// Upstream's `getStaticPropertyName` on a Property. A computed key is only static when it is a
// literal: `{ [k]: v }` names a property nobody can know at lint time, and treating it as matching
// would report an assignment that may not be one. A template without substitutions counts, so a
// key written as a bracketed template of `a` matches `{ a: b }`.
func propertyName(member *ast.Node) (string, bool) {
	var name *ast.Node
	switch member.Kind {
	case ast.KindPropertyAssignment:
		name = member.AsPropertyAssignment().Name()
	case ast.KindShorthandPropertyAssignment:
		name = member.AsShorthandPropertyAssignment().Name()
	default:
		return "", false
	}
	if name == nil {
		return "", false
	}

	switch name.Kind {
	case ast.KindIdentifier:
		return name.Text(), true
	case ast.KindComputedPropertyName:
		return noSelfAssignStaticStringValue(ast.SkipParentheses(name.AsComputedPropertyName().Expression))
	}
	return noSelfAssignStaticStringValue(name)
}

// propertyValue reads the value a property assigns, which for shorthand is the name itself.
//
// A shorthand target with a default, the `a = 1` in `({ a = 1 } = { a })`, answers nil. ESTree
// gives that value an AssignmentPattern, which matches nothing on the value side, so upstream never
// reports it, and it is a fallback rather than a self-assignment.
func propertyValue(property *ast.Node) *ast.Node {
	switch property.Kind {
	case ast.KindPropertyAssignment:
		return property.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		shorthand := property.AsShorthandPropertyAssignment()
		if shorthand.ObjectAssignmentInitializer != nil {
			return nil
		}
		return shorthand.Name()
	}
	return nil
}

// noSelfAssignStaticStringValue is upstream's `getStaticStringValue`: the property name a literal
// denotes once JavaScript coerces it to a string.
//
// Wider than the property package's accept sets, because upstream coerces every literal: `a[null]`
// is `a.null`, and the corpus asserts `a['/(?<zero>0)/']` and `a[/(?<zero>0)/]` are one reference.
// The parser normalizes numbers, so `1e1` already reads "10" as `String(1e1)` does.
func noSelfAssignStaticStringValue(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral,
		ast.KindRegularExpressionLiteral:
		return node.Text(), true
	case ast.KindBigIntLiteral:
		return strings.TrimSuffix(node.Text(), "n"), true
	case ast.KindNullKeyword:
		return "null", true
	case ast.KindTrueKeyword:
		return "true", true
	case ast.KindFalseKeyword:
		return "false", true
	}
	return "", false
}

// noSelfAssignStaticAccessedName is upstream's `getStaticPropertyName` on a member expression.
//
// A private name answers nothing, as it does upstream: `#field` is not a string key, so
// `this['#field']` and `this.#field` must never compare equal, though their text agrees.
func noSelfAssignStaticAccessedName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), true
		}
	case ast.KindElementAccessExpression:
		return noSelfAssignStaticStringValue(ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression))
	}
	return "", false
}

// isSameReference reports whether two expressions name the same place in memory.
//
// Upstream's `isSameReference` with `disableStaticComputedKey` false. Not token equality, and the
// difference is the whole reason the props option needs its own comparison. ESLint treats these as
// the same reference:
//
//	a.b  and  a?.b       optional chaining changes when the read happens, not what is read
//	x.y  and  x["y"]     a static computed key is the same property
//	x[y] and  x[y]       an identifier subscript read twice with nothing between
//
// Our parser has no ChainExpression wrapper to unwrap, since `?.` is a token on the access itself,
// so ignoring it is the comparison upstream's unwrapping arrives at. Anything outside the arms below
// is false: `f().x = f().x` calls f twice, and `a[i + 1]` is an expression rather than a reference.
func isSameReference(left *ast.Node, right *ast.Node) bool {
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	if left == nil || right == nil {
		return false
	}

	if ast.IsAccessExpression(left) && ast.IsAccessExpression(right) {
		// When the target's key is static, the value's must be the same static key, which is what
		// lets a dot and a bracket spelling meet before their kinds are compared.
		if leftName, leftStatic := noSelfAssignStaticAccessedName(left); leftStatic {
			rightName, rightStatic := noSelfAssignStaticAccessedName(right)
			return rightStatic && leftName == rightName &&
				isSameReference(accessedObject(left), accessedObject(right))
		}

		// Otherwise both must be written the same way, dotted or bracketed, with the same key
		// expression: `a[b] = a[b]` and `this.#a = this.#a`.
		if left.Kind != right.Kind {
			return false
		}
		if left.Kind == ast.KindPropertyAccessExpression {
			return isSameReference(left.AsPropertyAccessExpression().Expression, right.AsPropertyAccessExpression().Expression) &&
				isSameReference(left.AsPropertyAccessExpression().Name(), right.AsPropertyAccessExpression().Name())
		}
		return isSameReference(left.AsElementAccessExpression().Expression, right.AsElementAccessExpression().Expression) &&
			isSameReference(left.AsElementAccessExpression().ArgumentExpression, right.AsElementAccessExpression().ArgumentExpression)
	}

	if left.Kind != right.Kind {
		return false
	}

	switch left.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return left.Text() == right.Text()
	case ast.KindThisKeyword, ast.KindSuperKeyword,
		ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral:
		// Upstream compares literal values, and `Text()` is the cooked value for a string and the
		// canonical rendering for a number. A template is a TemplateLiteral in ESTree rather than a
		// Literal, so it falls to the default arm there and here.
		return left.Text() == right.Text()
	}

	return false
}

// accessedObject reads the receiver of an access expression.
func accessedObject(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}
