package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageForbidForeignPropTypes = rule.Message{
	Id: "forbiddenPropType",
	Description: "Reading another component's `propTypes` breaks in production. The React " +
		"production build strips `propTypes` off components to save bytes, so this expression is " +
		"an object in development and `undefined` in the build users actually run, and the " +
		"failure lands at runtime in the one environment nobody was testing. Declare the shape " +
		"where it is used, or export the plain object both components can import.",
}

// ForbidForeignPropTypesOptions is the decoded option object.
//
// One field, matching `meta.schema`'s one property. It defaults to false, which is Go's zero value,
// so the generic decoder is safe here and no hand-rolled wire type is needed.
type ForbidForeignPropTypesOptions struct {
	// AllowInPropTypes exempts a read that is itself building a `propTypes` declaration.
	//
	// False by default. It is the whole option surface, and it is what separates borrowing another
	// component's shape as a value (always wrong) from re-declaring one field of it inside your own
	// `propTypes` (arguably fine, since both disappear together in production).
	AllowInPropTypes bool `json:"allowInPropTypes"`
}

// ForbidForeignPropTypes flags a read of another component's `propTypes`.
//
//	valid:   import { propTypes } from "SomeComponent"
//	valid:   const foo = propTypes
//	valid:   Foo.propTypes = propTypes
//	valid:   Foo["propTypes"] = propTypes
//	valid:   const propTypes = "bar"; Foo[propTypes];
//	invalid: var Foo = createReactClass({ propTypes: Bar.propTypes })
//	invalid: var Foo = createReactClass({ propTypes: Bar["propTypes"] })
//	invalid: var { propTypes } = SomeComponent
//	invalid: var { propTypes: things, ...foo } = SomeComponent
//	invalid: class C extends React.Component { static fooBar = { baz: Qux.propTypes.baz }; }
//
// Ported from `react/forbid-foreign-prop-types` in `eslint-plugin-react`, read from the clone at
// `lib/rules/forbid-foreign-prop-types.js`. One option, one message, no fixer. The whole 20-case
// corpus was replayed against the installed build (7.37.5) through the ESLint Linter API before any
// code was written, and it agreed on all 20. Everything below that the corpus does not state was
// measured the same way, on inputs written for the question.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on `.tsx`/`.jsx`, which is oxc residue rather than upstream
// behaviour. This rule is ported from the authority, which has no such gate, and most of what it
// would catch is in plain `.ts` files: `const { propTypes } = SomeComponent` needs no JSX at all.
//
// # Two upstream selectors, and neither is about React
//
// The rule never asks whether anything is a component. A member access whose property is named
// `propTypes` reports, and an object destructuring pattern with a `propTypes` key reports, whatever
// the receiver is. That breadth is upstream's judgment and reproducing it means not adding a
// component check that would look like an improvement.
//
// # What counts as the property name
//
//	Foo.propTypes            reports, a plain identifier property
//	Foo["propTypes"]         reports, upstream's second arm matches a Literal property
//	Foo[propTypes]           CLEAN, a computed identifier has no literal text to compare
//	Foo[`propTypes`]         CLEAN, a template is not a Literal to upstream's parser
//	Foo?.propTypes           reports, optional chaining is still a member access
//	Foo.Bar.propTypes        reports, the receiver is not examined
//
// All measured. The element-access arm exists because upstream wrote it, and its corpus carries the
// reporting case; the computed-identifier decline is a corpus case too.
//
// # Being the target of an assignment exempts, and the definition of that is narrow
//
// Upstream calls `ast.isAssignmentLHS`, which is `parent.type === 'AssignmentExpression' &&
// parent.left === node`. Two consequences were measured and neither is obvious:
//
//	Foo.propTypes = {}       CLEAN, it is the left of an assignment
//	Foo.propTypes += 1       CLEAN, a compound assignment is still an AssignmentExpression
//	Foo.propTypes++          REPORTS, an update expression is not one
//	delete Foo.propTypes     REPORTS, a unary operator is not one
//
// The `++` and `delete` cases are upstream reporting on code that plainly writes rather than reads,
// and they are reproduced rather than corrected. `ast.IsAssignmentExpression(node, false)` is the
// exact equivalent: it accepts the plain and the compound operators alike, which is what ESTree's
// single `AssignmentExpression` type covers, and rejects every unary form.
//
// # The parenthesis difference costs a false positive here, in the direction that reports
//
// `(Foo.propTypes) = {}` is CLEAN upstream, measured. Its parser folds the parenthesis away, so the
// member access IS the assignment's `left`. Ours keeps a `KindParenthesizedExpression` between
// them, so a naive `node.Parent` test sees a parenthesis rather than an assignment, concludes the
// access is not a target, and reports where upstream is silent. The parent walk unwraps
// parentheses in a loop, because `((Foo.propTypes)) = {}` nests. Written as a loop rather than with
// `ast.SkipParentheses`, which dereferences its argument.
//
// # `allowInPropTypes` searches upward for two enclosing shapes
//
// Upstream walks up from the reported node to `Program`, twice, looking for an
// `AssignmentExpression` whose left-hand property is named `propTypes`, and for a class property
// whose key is named `propTypes`. Either one exempts. Both walks are unbounded, which is what makes
// the exemption reach arbitrarily deep:
//
//	Foo.propTypes = { a: { b: Bar.propTypes.c } }        exempt, nested two levels
//	class C { static propTypes = { a: () => Q.propTypes.c }; }   exempt, through a closure
//	Foo.other = Bar.propTypes                            REPORTS, the left property is not propTypes
//
// All measured under `{allowInPropTypes: true}`. The class-property walk does NOT require `static`,
// measured: `class C { propTypes = { a: Q.propTypes.b }; }` is exempt too. And a computed class key
// declines, because upstream reads `classProperty.key.name`, which a computed key does not have.
//
// The two walks are separate in upstream and are kept separate here, because the assignment walk
// tests the assignment's LEFT while the class walk tests the declaration's own key, and collapsing
// them into one upward scan would need a different test at each stop anyway.
//
// # The destructuring arm reports once per pattern and points at the whole member
//
// Upstream uses `properties.find`, so a pattern naming `propTypes` twice reports ONCE. Measured:
// `const { propTypes, propTypes: q } = Foo` produces a single finding. The finding is anchored on
// the matched property rather than on its key, so `const { propTypes: p } = Foo` spans
// `propTypes: p` and not just the key. Both are reproduced.
//
// A computed or string-literal key in a pattern declines, because upstream reads `property.key.name`
// there too:
//
//	const { propTypes } = Foo            reports
//	const { propTypes: p } = Foo         reports
//	const { ["propTypes"]: p } = Foo     CLEAN
//	const { "propTypes": p } = Foo       CLEAN
//	const [ propTypes ] = Foo            CLEAN, an array pattern is not an ObjectPattern
//	function f({ propTypes }) {}         reports, any object pattern counts
//
// All measured. `allowInPropTypes` is NOT consulted on this arm, upstream, which is visible in the
// source rather than measured: the `ObjectPattern` handler calls neither `isAllowedAssignment` nor
// anything like it. Reproduced.
var ForbidForeignPropTypes = rule.Rule{
	Name: "react/forbid-foreign-prop-types",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[ForbidForeignPropTypesOptions](options)

		// unwrapParentheses walks up past the parentheses our parser keeps and upstream's folds
		// away. A loop rather than a single step, because `((x))` nests, and written here rather
		// than with `ast.SkipParentheses`, which dereferences its argument.
		unwrapParentheses := func(node *ast.Node) *ast.Node {
			for node != nil && node.Kind == ast.KindParenthesizedExpression {
				node = node.Parent
			}
			return node
		}

		// isAssignmentTarget reproduces upstream's `ast.isAssignmentLHS`.
		//
		// `ast.IsAssignmentExpression(parent, false)` accepts the plain and compound operators
		// alike, which is exactly what ESTree's one `AssignmentExpression` type covers, and rejects
		// an update expression and a `delete`, both of which upstream therefore reports.
		isAssignmentTarget := func(node *ast.Node) bool {
			parent := unwrapParentheses(node.Parent)
			if parent == nil || !ast.IsAssignmentExpression(parent, false) {
				return false
			}
			// The node reached the assignment through however many parentheses, so the assignment's
			// own left is compared after unwrapping downward the same way.
			left := parent.AsBinaryExpression().Left
			for left != nil && left.Kind == ast.KindParenthesizedExpression {
				left = left.AsParenthesizedExpression().Expression
			}
			return left == node
		}

		// isAllowedAssignment reproduces upstream's two upward walks, which run only under the
		// option and stop at the file rather than at any smaller boundary.
		isAllowedAssignment := func(node *ast.Node) bool {
			if !settings.AllowInPropTypes {
				return false
			}

			// Walk one: an enclosing assignment whose left-hand property is named `propTypes`.
			// Upstream reads `assignmentExpression.left.property.name`, which is defined only for a
			// plain member access with an identifier property, so `Foo["propTypes"] = ...` does not
			// exempt through this route.
			for parent := node.Parent; parent != nil && parent.Kind != ast.KindSourceFile; parent = parent.Parent {
				if !ast.IsAssignmentExpression(parent, false) {
					continue
				}
				left := parent.AsBinaryExpression().Left
				for left != nil && left.Kind == ast.KindParenthesizedExpression {
					left = left.AsParenthesizedExpression().Expression
				}
				if left == nil || left.Kind != ast.KindPropertyAccessExpression {
					continue
				}
				property := left.AsPropertyAccessExpression().Name()
				if property != nil && property.Kind == ast.KindIdentifier && property.Text() == "propTypes" {
					return true
				}
			}

			// Walk two: an enclosing class property whose own key is named `propTypes`. Upstream
			// reads `classProperty.key.name`, so a computed key declines, and it does not test
			// `static`, so an instance property exempts too. Both measured.
			for parent := node.Parent; parent != nil && parent.Kind != ast.KindSourceFile; parent = parent.Parent {
				if parent.Kind != ast.KindPropertyDeclaration {
					continue
				}
				key := parent.Name()
				if key != nil && key.Kind == ast.KindIdentifier && key.Text() == "propTypes" {
					return true
				}
			}

			return false
		}

		// reportProperty is the shared tail of both member arms. Upstream anchors the finding on the
		// PROPERTY node rather than on the whole member access, so `Foo.propTypes` spans
		// `propTypes` and `Foo["propTypes"]` spans `"propTypes"` with its quotes.
		reportProperty := func(access *ast.Node, property *ast.Node) {
			if isAssignmentTarget(access) || isAllowedAssignment(access) {
				return
			}
			ctx.ReportNode(property, messageForbidForeignPropTypes)
		}

		return rule.Listeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				// Upstream's first arm requires a non-computed identifier property, which is
				// exactly what a property access carries here.
				property := node.AsPropertyAccessExpression().Name()
				if property == nil || property.Kind != ast.KindIdentifier || property.Text() != "propTypes" {
					return
				}
				reportProperty(node, property)
			},

			ast.KindElementAccessExpression: func(node *ast.Node) {
				// Upstream's second arm matches a Literal property whose value is `propTypes`. A
				// computed identifier and a template literal both decline, measured, because
				// neither is a Literal to its parser.
				argument := node.AsElementAccessExpression().ArgumentExpression
				if argument == nil || argument.Kind != ast.KindStringLiteral || argument.Text() != "propTypes" {
					return
				}
				reportProperty(node, argument)
			},

			ast.KindObjectBindingPattern: func(node *ast.Node) {
				elements := node.AsBindingPattern().Elements
				if elements == nil {
					return
				}
				// Upstream uses `properties.find`, so a pattern naming `propTypes` twice reports
				// once. The loop returns on the first match to reproduce that.
				for _, element := range elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					// Upstream reads `property.key.name`. A shorthand `{propTypes}` has its name as
					// the key; a renamed `{propTypes: p}` carries a PropertyName that is the key,
					// and the local name is separate. A computed or string-literal key has no
					// `name` and declines, measured.
					key := bindingElementPropertyKey(element)
					if key == nil || key.Kind != ast.KindIdentifier || key.Text() != "propTypes" {
						continue
					}
					// Anchored on the whole element, not on its key, which is what makes
					// `{propTypes: p}` span both halves. Measured against upstream's own spans.
					ctx.ReportNode(element, messageForbidForeignPropTypes)
					return
				}
			},
		}
	},
}

// bindingElementPropertyKey returns the key a destructuring element reads from its source object.
//
// This is the half of a binding element that names the PROPERTY, which is not always the half that
// names the local variable, and the distinction is the whole reason this function exists:
//
//	const { propTypes } = Foo        PropertyName is nil, Name() is `propTypes`, both roles at once
//	const { propTypes: p } = Foo     PropertyName is `propTypes`, Name() is `p`
//
// Upstream reads `property.key.name` in both cases, so the key is what this returns. Reading
// `Name()` unconditionally would answer `p` for the renamed form and miss it, while reading
// `PropertyName` unconditionally would answer nil for the shorthand and miss that one instead.
func bindingElementPropertyKey(element *ast.Node) *ast.Node {
	binding := element.AsBindingElement()
	if binding.PropertyName != nil {
		return binding.PropertyName
	}
	return binding.Name()
}
