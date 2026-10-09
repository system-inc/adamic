package react

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

const messageIdNoRedundantShouldComponentUpdate = "noShouldCompUpdate"

// NoRedundantShouldComponentUpdate flags a class that extends PureComponent and also writes
// shouldComponentUpdate.
//
//	valid:   class Foo extends React.Component { shouldComponentUpdate() { return true; } }
//	valid:   class Foo extends React.PureComponent { render() { return null; } }
//	invalid: class Foo extends React.PureComponent { shouldComponentUpdate() { return true; } }
//	invalid: var Foo = class extends PureComponent { shouldComponentUpdate() { return true; } }
//
// PureComponent already implements shouldComponentUpdate as a shallow comparison of props and
// state. Writing your own replaces that comparison entirely, so the class pays for the base class's
// machinery and then discards it, and a reader who sees PureComponent in the heritage will believe
// a shallow compare is happening when it is not. Extend Component instead, or delete the method.
//
// Ported from `eslint-plugin-react`'s `no-redundant-should-component-update.js`. `meta.schema` is
// `[]`, checked against the installed build, so there is no option surface. `meta.fixable` is
// absent, and there is nothing mechanical to write anyway: the two repairs are deleting a method and
// changing a base class, which mean different things and only the author knows which was intended.
//
// # What counts as PureComponent, and it is narrower than the shelf
//
// `react.IsEs6ComponentClass` is NOT used here, deliberately. It accepts `Component` and
// `PureComponent` alike, which is the right question for the eight rules that ask "is this a class
// component" and the wrong one here: `class Foo extends React.Component` with a
// shouldComponentUpdate is upstream's very FIRST valid case, and routing through the shelf would
// report it. The two predicates answer different questions and this one needs its own.
//
// Upstream tests the heritage's SOURCE TEXT against `^(React\.)?PureComponent$`, which is why the
// accepted set is exactly two spellings and why a third package's is not. Measured on the installed
// build:
//
//	class Foo extends React.PureComponent { shouldComponentUpdate() {} }    REPORTS
//	class Foo extends PureComponent { shouldComponentUpdate() {} }          REPORTS
//	class Foo extends (React.PureComponent) { shouldComponentUpdate() {} }  REPORTS
//	class Foo extends React.Component { shouldComponentUpdate() {} }        SILENT
//	class Foo extends Preact.PureComponent { shouldComponentUpdate() {} }   SILENT
//	class Foo extends Foo.PureComponent { shouldComponentUpdate() {} }      SILENT
//	class Foo extends PureComponent {}                                      SILENT
//	class Foo { shouldComponentUpdate() {} }                                SILENT
//
// The parenthesized line is the interesting one and it looks like a contradiction: that regex is
// anchored, so it cannot match a string with parentheses in it. It reports anyway because a
// parenthesis is not a node in the tree the reference implementation walks, so `node.superClass` is
// already the member access and `getText` on it never sees the parentheses. Our parser does give
// them a node, so reproducing that answer takes an explicit skip where upstream needed none. Same
// shape as the callee in `no_namespace.go`, measured the same way.
//
// # What counts as writing the method, which is wider than "a method"
//
// Upstream takes `node.body.body`, every class member with no filtering at all, and asks
// `getPropertyName` for each. So there is no method test, no static test and no accessor test, and
// every one of those shapes reports. Measured:
//
//	shouldComponentUpdate() { return true; }         REPORTS   a method
//	shouldComponentUpdate = () => { return true; }   REPORTS   a class field
//	shouldComponentUpdate;                           REPORTS   a bare field with no initializer
//	static shouldComponentUpdate() {}                REPORTS   static is not excluded
//	get shouldComponentUpdate() { return true; }     REPORTS   nor is an accessor
//	#shouldComponentUpdate() {}                      REPORTS   nor is a private name
//	['shouldComponentUpdate']() {}                   SILENT    computed
//	'shouldComponentUpdate'() {}                     SILENT    a string key
//
// The last two lines are the ones a reader will want to "fix", and they are not a special case:
// `getPropertyName` reads `nameNode.name`, a field only an Identifier and a PrivateIdentifier carry.
// A string-literal key has `.value` and a computed key's node is an expression, so both yield the
// empty string and neither can equal the method name. The private-name line reports for the same
// mechanical reason in the other direction, and it is worth stating that `#shouldComponentUpdate`
// is not the lifecycle hook at all: React never calls it, so upstream is arguably wrong here. It is
// reproduced rather than corrected, because a divergence I invent is worse than one I inherit, and
// the intuitive reading is recorded so the next reader does not helpfully correct it.
//
// # The class name in the message
//
// Upstream reads `node.id.name`, then `node.parent.id.name`, then the empty string. So a named class
// uses its own name even when assigned to a differently named variable, an anonymous class
// expression borrows the name only from a variable declarator, and everything else renders an empty
// name into the message. Measured, and the empty renderings are real output rather than an error:
//
//	class Foo extends PureComponent               -> "Foo"
//	const Foo = class Bar extends PureComponent   -> "Bar"    the class's own id wins
//	var Foo = class extends PureComponent         -> "Foo"    borrowed from the declarator
//	export default class extends PureComponent    -> ""       no id, parent has none either
//	foo(class extends PureComponent {})           -> ""       an argument has no id
//	const o = { Foo: class extends PureComponent }-> ""       a property is not a declarator
//
// Reproduced exactly, empty string included. A port that helpfully fell back to the property name
// would report text upstream never produces, and no fixture asserting a message id could see it.
var NoRedundantShouldComponentUpdate = rule.Rule{
	Name: "react/no-redundant-should-component-update",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		check := func(node *ast.Node) {
			if !extendsPureComponent(node) {
				return
			}
			members := classMembers(node)
			for _, member := range members {
				if identifierMemberName(member) != "shouldComponentUpdate" {
					continue
				}
				ctx.ReportNode(node, rule.Message{
					Id: messageIdNoRedundantShouldComponentUpdate,
					Description: fmt.Sprintf(
						"%s extends PureComponent and also writes shouldComponentUpdate. "+
							"PureComponent's whole contribution IS an implementation of that method, "+
							"a shallow comparison of props and state, and defining your own replaces "+
							"it rather than adding to it. So the class pays for the base class and "+
							"then discards what it bought, and a reader who sees PureComponent in "+
							"the heritage will believe a shallow compare is happening when it is "+
							"not. Extend Component instead, or delete the method.",
						classNameForMessage(node),
					),
				})
				// Upstream's `properties.some(...)` stops at the first match, so a class writing
				// the method twice reports ONCE. Measured on the installed build with a duplicated
				// method: one finding, not two.
				return
			}
		}

		return rule.Listeners{
			ast.KindClassDeclaration: check,
			ast.KindClassExpression:  check,
		}
	},
}

// extendsPureComponent reports whether a class extends one of the two accepted PureComponent
// spellings.
//
// Deliberately not `react.IsEs6ComponentClass`, which also accepts `Component` and would report
// upstream's first valid case. See the rule's doc comment.
func extendsPureComponent(node *ast.Node) bool {
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
		// An `implements` clause is a different kind of node under our parser than an `extends` one,
		// but both live in HeritageClauses, so the token is what separates them. Upstream reads
		// `node.superClass`, which is the extends target and nothing else.
		if clause.Kind != ast.KindHeritageClause {
			continue
		}
		if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
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
			if isPureComponentBase(typeNode.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

// isPureComponentBase reports whether an extends target is `PureComponent` or `React.PureComponent`.
//
// The parenthesis skip reproduces upstream rather than improving on it: `node.superClass` in the
// reference implementation is already the member access, because its tree has no parenthesis node,
// so `class Foo extends (React.PureComponent)` reports there. Measured on the installed build.
//
// The nil test comes BEFORE the skip, because `ast.SkipParentheses` dereferences its argument and
// the walk recovers per FILE rather than per rule.
func isPureComponentBase(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	expression = ast.SkipParentheses(expression)
	if expression == nil {
		return false
	}

	switch expression.Kind {
	case ast.KindIdentifier:
		return expression.Text() == "PureComponent"

	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		// The object must be exactly `React`. Upstream builds its regex from the configured pragma
		// and compares the whole rendered text, so `Preact.PureComponent` and `Foo.PureComponent`
		// both decline on that anchor. Both measured silent.
		object := access.Expression
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "PureComponent"
	}
	return false
}

// classMembers returns every member of a class, with no filtering.
//
// Upstream's `getComponentProperties` returns `node.body.body` for both class forms, which is every
// member: methods, fields, accessors, static and instance alike. The absence of a filter here is
// load-bearing rather than an omission, and it is what makes static, accessor and field spellings
// all report. See the rule's doc comment for the measurements.
func classMembers(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindClassDeclaration:
		if node.AsClassDeclaration().Members == nil {
			return nil
		}
		return node.AsClassDeclaration().Members.Nodes
	case ast.KindClassExpression:
		if node.AsClassExpression().Members == nil {
			return nil
		}
		return node.AsClassExpression().Members.Nodes
	}
	return nil
}

// identifierMemberName reads a class member's name the way upstream's `getPropertyName` does.
//
// That function returns `nameNode.name`, which exists on an Identifier and on a PrivateIdentifier
// and on nothing else, so a string-literal key and a computed key both yield the empty string and
// can never match. Returning the empty string for them here is the same decision, reached the same
// way, and both are measured silent upstream.
//
// A private name is accepted because upstream accepts it: `#shouldComponentUpdate` reports there,
// measured, even though React never calls a private method and the finding is arguably wrong.
// Reproduced rather than corrected.
func identifierMemberName(member *ast.Node) string {
	if member == nil {
		return ""
	}
	name := member.Name()
	if name == nil {
		return ""
	}
	// A computed key's name node is a ComputedPropertyName and a string key's is a StringLiteral.
	// Both decline here, which reproduces upstream and is ALSO crash protection: `Text()` has no
	// case for a ComputedPropertyName and reaches its unhandled-case panic, measured directly. The
	// walk recovers per FILE, so a single computed member in a PureComponent would cost every rule
	// in this package every finding in that file. No ExpectFindings fixture can see a panic, which
	// is why TestNoRedundantShouldComponentUpdateDoesNotPanicOnComputedMembers exists separately.
	if name.Kind != ast.KindIdentifier && name.Kind != ast.KindPrivateIdentifier {
		return ""
	}
	// A PrivateIdentifier's Text() already carries the leading hash, measured: the node for
	// `#shouldComponentUpdate` answers "#shouldComponentUpdate" rather than the bare name. Upstream
	// compares `nameNode.name`, which for its PrivateIdentifier is the name WITHOUT the hash, so a
	// direct comparison here would be silent on a case upstream reports. The hash is trimmed rather
	// than the comparison being widened, so the returned value means the same thing for both kinds.
	if name.Kind == ast.KindPrivateIdentifier {
		return strings.TrimPrefix(name.Text(), "#")
	}
	return name.Text()
}

// classNameForMessage reproduces upstream's `getNodeName`: the class's own id, else the id of its
// parent, else the empty string.
//
// The empty string is a real rendering rather than a failure. Upstream prints a message beginning
// with a space for an anonymous class in an argument or an object property, and four such shapes are
// measured in the rule's doc comment. A port that fell back to a property name would produce text
// upstream never writes, and no message-id fixture could see it.
func classNameForMessage(node *ast.Node) string {
	if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
		return name.Text()
	}
	// Upstream reads `node.parent.id`, which only a VariableDeclarator carries. An object property,
	// a call argument and an assignment all have no `id` and render the empty name.
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindVariableDeclaration {
		if declared := parent.Name(); declared != nil && declared.Kind == ast.KindIdentifier {
			return declared.Text()
		}
	}
	return ""
}
