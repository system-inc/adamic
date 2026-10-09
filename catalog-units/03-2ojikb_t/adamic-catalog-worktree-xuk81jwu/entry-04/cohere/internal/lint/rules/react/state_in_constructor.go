package react

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// StateInConstructorMode is which of the two initialization styles the configuration wants.
type StateInConstructorMode string

const (
	// StateInConstructorAlways wants state assigned in the constructor, so a class property
	// named `state` is what reports. This is upstream's default and the unconfigured answer.
	StateInConstructorAlways StateInConstructorMode = "Always"

	// StateInConstructorNever wants state declared as a class property, so a `this.state`
	// assignment inside the constructor is what reports.
	StateInConstructorNever StateInConstructorMode = "Never"
)

// StateInConstructorOptions configures the rule.
//
// Upstream's option is a bare string, `"always"` or `"never"`, and the config takes it in exactly
// that form. It used to take `{"mode": "Always"}`, a shape no ESLint version accepts, which refused
// upstream's own spelling; that form is now refused instead (#d21war2).
type StateInConstructorOptions struct {
	// Mode is which style the rule enforces. Absent means Always, which is upstream's default.
	Mode StateInConstructorMode
}

// DefaultStateInConstructorOptions is the unconfigured answer.
//
// Upstream reads `context.options[0] || 'always'`, so an absent option and an explicitly configured
// `always` are the same rule. Both of those spellings appear in the corpus over byte-identical
// source, which is how the corpus states that they agree.
func DefaultStateInConstructorOptions() StateInConstructorOptions {
	return StateInConstructorOptions{Mode: StateInConstructorAlways}
}

// DecodeStateInConstructorOptions reads this rule's configuration from the config layer.
//
// The option is a bare enum string rather than an object, so the generic helper, which unmarshals
// into a struct, cannot read it. Upstream's schema is `[{ enum: ["always", "never"] }]`, so anything
// else is refused, naming the value: the two modes disagree about every input this rule can see, and
// a value read as neither would silently select a third behaviour of reporting nothing at all.
//
// A rule configured as bare `"error"` is handed no options at all, and the empty case has to return
// the default rather than the zero value for the same reason: an empty Mode matches neither arm.
func DecodeStateInConstructorOptions(raw []byte) (any, error) {
	options := DefaultStateInConstructorOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var mode string
	if err := json.Unmarshal(raw, &mode); err != nil {
		return options, fmt.Errorf(`expected a mode string, "always" or "never", got %s`, raw)
	}
	switch mode {
	case "always":
		return options, nil
	case "never":
		options.Mode = StateInConstructorNever
		return options, nil
	}
	return options, fmt.Errorf(`mode %q is not "always" or "never"`, mode)
}

var messageStateInConstructorInConstructor = rule.Message{
	Id: "stateInitConstructor",
	Description: "This component declares `state` as a class property while the configuration " +
		"asks for state to be set up in the constructor. Both styles work; what costs the " +
		"reader is a codebase that uses each in about half its components, because then " +
		"finding where a component's initial state comes from means checking two places " +
		"instead of one. Move the initial value into `this.state = ...` in the constructor.",
}

var messageStateInConstructorClassProperty = rule.Message{
	Id: "stateInitClassProp",
	Description: "This component assigns `this.state` in its constructor while the " +
		"configuration asks for state to be declared as a class property. Both styles work; " +
		"what costs the reader is a codebase that uses each in about half its components, " +
		"because then finding where a component's initial state comes from means checking " +
		"two places instead of one. Declare `state = ...` as a class property instead.",
}

// StateInConstructor enforces one of the two ways a React class component can set up its initial
// state, and reports the other one.
//
//	Always (the default):
//	  valid:   class F extends React.Component { constructor(p) { super(p); this.state = {}; } }
//	  valid:   class F extends React.Component { baz = {}; }         (some other property)
//	  valid:   class Plain { state = {}; }                            (not a component)
//	  invalid: class F extends React.Component { state = {}; }
//
//	Never:
//	  valid:   class F extends React.Component { state = {}; }
//	  valid:   class F extends React.Component { constructor(p) { super(p); this.baz = {}; } }
//	  invalid: class F extends React.Component { constructor(p) { super(p); this.state = {}; } }
//
// Ported from `react/state-in-constructor` in `eslint-plugin-react`, read from the clone at
// `lib/rules/state-in-constructor.js` together with the three helpers it calls:
// `componentUtil.getParentES6Component`, `componentUtil.isStateMemberExpression`, and
// `astUtil.inConstructor`. Every divergence below was measured by driving the installed build
// (`eslint-plugin-react` 7.37.5) through the ESLint Linter API rather than reasoned from the source.
//
// # There is no file-suffix gate, and three sibling rules in this package have one
//
// Those siblings inherited it from oxc, which gates on `source_type().is_jsx()`. This rule was
// ported from the authority instead, and the authority has no such gate anywhere. Measured by
// running the installed rule over a component with no JSX in it, so every suffix parses and the
// parser cannot be confused with the rule, under `.tsx`, `.jsx`, `.ts` and `.js`: all four report.
// A `.ts` file holding JSX is a parse error rather than a rule decision, which is why the probe had
// to drop the JSX to separate the two questions.
//
// # What the Always arm keys on, and the two places our parser disagrees with ESLint's
//
// Upstream tests `!node.static && node.key.name === 'state'`, and `key.name` is a field that only
// some key shapes carry. That makes the predicate depend on the parser's key representation rather
// than on anything the rule decided, and the two parsers represent two key shapes differently:
//
//	state = {}         reports upstream    Identifier, name `state`
//	static state = {}  SILENT upstream     excluded by !node.static
//	["state"] = {}     SILENT upstream     computed, so key is a Literal with no `.name`
//	"state" = {}       SILENT upstream     also a Literal with no `.name`
//	#state = {}        REPORTS upstream    PrivateIdentifier, whose `.name` is the bare `state`
//	state              reports upstream    no initializer; the rule never looks at the value
//	declare state: any reports upstream    a declaration is still a property definition
//
// The last two rows of that list are the interesting ones and they run opposite ways in our tree,
// which is why the comparison here is written against the key's KIND rather than only its text:
//
//   - A private name's `Text()` here is `#state`, with the hash, where the ESLint AST's `.name` is
//     the bare `state`. So a text comparison alone would go SILENT on `#state` and lose a finding
//     upstream produces. Upstream reporting there is a defect rather than a decision, since `#state`
//     is not React's state at all, but it is reproduced: fidelity is to what the rule decides, and
//     a port that silently improves on it is a divergence nothing records.
//
//   - A string-literal key's `Text()` here is the bare `state`, where ESLint's Literal carries the
//     value on `.value` and has no `.name`. So a text comparison alone would REPORT on `"state" =
//     {}` where upstream is silent, inventing a finding. This is the direction that costs a false
//     positive, and it is the reason the key kind is checked at all.
//
// Neither shape is anywhere in the corpus, so both would have shipped green.
//
// # What the Never arm keys on
//
// Upstream tests `isStateMemberExpression(node.left) && inConstructor(...) && getParentES6Component(...)`.
// `isStateMemberExpression` requires the left side to be a member expression whose object is a
// `ThisExpression` and whose property is named `state`, so the target must be exactly `this.state`
// and nothing deeper. Measured:
//
//	this.state = {}          reports    the direct target
//	this.state.x = 1         SILENT     the target is `this.state.x`, whose object is not `this`
//	this["state"] = {}       SILENT     an element access, not a member expression with a name
//	this.state += 1          reports    every assignment operator, not only `=`
//	(this.state) = {}        reports    see the parentheses note below
//	(this).state = {}        reports    see the parentheses note below
//
// `inConstructor` walks the SCOPE chain upward rather than the syntactic ancestor chain, which
// sounds equivalent and is not: a write inside a callback declared in the constructor is still
// "in the constructor" to this rule, and upstream reports it. Measured:
//
//	constructor(p) { super(p); go(() => { this.state = {} }) }   REPORTS
//
// Reproduced here by walking ancestors and stopping at the first class, so a nested function does
// not end the walk but a nested class does. The nested-class stop is what makes a plain class
// written inside a component's constructor keep its own writes, and it is measured: a
// `class G { constructor() { this.state = {} } }` inside a real component's constructor is silent
// upstream, because `getParentES6Component` finds the inner class first and that one is not a
// component.
//
// # Parentheses, which have to be handled explicitly here and do not upstream
//
// ESLint's AST does not materialize a parenthesized expression as a node in these positions, so
// upstream never has to think about them and its corpus writes none. Ours does materialize them, so
// a port that ignores them goes silent on two shapes upstream reports. Both measured on the
// installed rule:
//
//	(this.state) = {}     REPORTS   parens around the whole target
//	(this).state = {}     REPORTS   parens around the receiver only
//
// So the target is unwrapped, and so is the receiver inside it. That is the opposite of the
// advice `no-direct-mutation-state` in this package records, where paren-skipping inside the chain
// would have INVENTED findings; the difference is that the rule being ported there was oxc's, whose
// tree keeps parens where ESLint's discards them. The authority decides, and the authority here is
// ESLint's.
//
// The span in both parenthesized cases is the whole assignment expression including the parens,
// which is what upstream reports and what the fixtures below assert by slicing the source.
//
// # Where the finding points
//
// Upstream passes the node itself to `report` in both arms, so the Always arm underlines the whole
// property declaration and the Never arm underlines the whole assignment expression, operator and
// right-hand side included. Measured by slicing the reported span out of the source:
//
//	state = { bar: 0 }                  the Always span, initializer included
//	state                               the Always span when there is no initializer
//	this.state = { bar: 0 }             the Never span, right-hand side included
//
// A message-id assertion cannot see any of that, so every fixture below asserts the sliced text.
var StateInConstructor = rule.Rule{
	// No namespace prefix beyond the plugin's own. The config writes `react/state-in-constructor`
	// and the parity guard strips the namespace on a `/` boundary.
	Name: "react/state-in-constructor",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[StateInConstructorOptions](options)
		if !ok {
			// A rule configured as bare `"error"` reaches here with nil options, which type-asserts
			// to the zero value whose Mode matches neither arm. Falling back to the default is what
			// keeps the rule from being registered and inert, which is the failure the brief
			// describes as a non-zero registration count with zero findings.
			settings = DefaultStateInConstructorOptions()
		}
		if settings.Mode == "" {
			settings.Mode = StateInConstructorAlways
		}

		if settings.Mode == StateInConstructorAlways {
			return rule.Listeners{
				ast.KindPropertyDeclaration: func(node *ast.Node) {
					if !isStateInConstructorStateProperty(node) {
						return
					}
					if stateInConstructorIsComponentClass(enclosingClassOf(node)) {
						ctx.ReportNode(node, messageStateInConstructorInConstructor)
					}
				},
			}
		}

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				assignment := node.AsBinaryExpression()
				// Upstream anchors on `AssignmentExpression`, which in the ESLint AST is a distinct
				// node covering every assignment operator. Ours is a binary expression, so the
				// operator set is the discriminator. `this.state += 1` reports upstream, which is
				// what makes the whole operator set the right question rather than `=` alone.
				if !ast.IsAssignmentOperator(assignment.OperatorToken.Kind) {
					return
				}
				if !isStateInConstructorThisStateTarget(assignment.Left) {
					return
				}
				if !stateInConstructorIsInsideConstructor(node) {
					return
				}
				if stateInConstructorIsComponentClass(enclosingClassOf(node)) {
					ctx.ReportNode(node, messageStateInConstructorClassProperty)
				}
			},
		}
	},
}

// isStateInConstructorStateProperty reports whether a class property is the non-static `state` that
// upstream's Always arm looks for.
//
// The key KIND is checked as well as its text, because the two parsers disagree about two key
// shapes in opposite directions and the rule's real predicate is "does this key carry a `.name`
// field spelling `state`". See the rule doc for the measurements. An identifier and a private
// identifier are the two kinds whose ESLint counterpart carries `.name`; a string literal and a
// computed name are the two that do not.
func isStateInConstructorStateProperty(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyDeclaration {
		return false
	}
	if ast.HasStaticModifier(node) {
		return false
	}
	name := node.AsPropertyDeclaration().Name()
	if name == nil {
		return false
	}
	switch name.Kind {
	case ast.KindIdentifier:
		return name.Text() == "state"
	case ast.KindPrivateIdentifier:
		// Our text carries the hash and ESLint's `.name` does not, so the comparison is against
		// the hashed spelling. Upstream reporting on `#state` is a defect it does not know it has;
		// it is reproduced rather than improved on, and this line is the whole of the reproduction.
		return name.Text() == "#state"
	}
	return false
}

// isStateInConstructorThisStateTarget reports whether an assignment target is exactly `this.state`.
//
// Parentheses are unwrapped at the target and again at the receiver, because ESLint's AST does not
// materialize them in either position and ours does. Both `(this.state) = {}` and `(this).state =
// {}` report upstream; without the unwrapping this rule would go silent on both, which no imported
// fixture could see because the corpus writes no parenthesized form at all.
func isStateInConstructorThisStateTarget(target *ast.Node) bool {
	target = ast.SkipParentheses(target)
	if target == nil || target.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := target.AsPropertyAccessExpression()
	if ast.SkipParentheses(access.Expression) == nil ||
		ast.SkipParentheses(access.Expression).Kind != ast.KindThisKeyword {
		return false
	}
	name := access.Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "state"
}

// stateInConstructorIsInsideConstructor reports whether a node has any constructor ancestor.
//
// This reproduces upstream's `inConstructor`, which walks the SCOPE chain looking for a scope whose
// block's parent is a constructor. Two things follow from it being the scope chain, and the second
// one cost a mutant here:
//
//   - A function boundary does not end the walk. A write inside a callback declared in the
//     constructor is still inside the constructor, and upstream reports it. Measured.
//
//   - **A class boundary does not end it either.** A class body creates no barrier scope, so the
//     walk climbs straight through an intervening class into an enclosing constructor. This rule
//     stopped at the first class for a whole revision, on the reasoning that a nested class owns
//     its own members. That reasoning describes `getParentES6Component`, which is a different walk
//     asking a different question, and applying it here was wrong.
//
// The distinguishing input, which nothing in the corpus writes, and all three rows measured on the
// installed build:
//
//	class Outer extends React.Component {
//	  constructor(p) { super(p); class Inner extends React.Component { m() { this.state = {} } } }
//	}                                                            REPORTS
//
//	class Outer extends React.Component {
//	  someMethod() { class Inner extends React.Component { m() { this.state = {} } } }
//	}                                                            SILENT
//
//	class Outer extends React.Component { someMethod() { this.state = {} } }
//	                                                             SILENT
//
// The first two have identical inner structure and differ only in whether the OUTER context is a
// constructor, which is the whole proof that the walk crosses the class. The third shows the rule
// still requires a constructor somewhere, so the walk is not vacuous.
//
// The component gate is a separate question and stays separate: `enclosingClassOf` takes the
// NEAREST class, so a plain inner class is still declined there. That is why the first row above
// needs the inner class to extend a component in order to report at all.
func stateInConstructorIsInsideConstructor(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current.Kind == ast.KindConstructor {
			return true
		}
	}
	return false
}

// enclosingClassOf returns the nearest class ancestor of a node, or nil.
//
// Upstream's `getParentES6Component` walks the scope chain to the nearest `class` scope and asks
// whether that class is a component, so the NEAREST class is the one that decides and an outer
// component cannot rescue an inner plain class. Written here rather than reaching for
// `react.EnclosingComponent`, which searches for the nearest thing that IS a component and would
// therefore skip past an intervening plain class and answer with the outer component. That is a
// different question and it reports on inputs upstream is silent on: a
// `class G { state = {} }` written inside a real component's render method is clean upstream, and
// measured so.
func enclosingClassOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsClassLike(current) {
			return current
		}
	}
	return nil
}

// stateInConstructorIsComponentClass reports whether a class extends a React component base,
// including through parentheses around the base expression.
//
// This wraps `react.IsEs6ComponentClass` rather than replacing it, because the shelf helper is
// correct about everything except one paren position and four rules ported against oxc depend on
// its current answer there.
//
// # The divergence, measured rather than reasoned
//
// The shelf skips parentheses around the RECEIVER inside a qualified name, so `extends
// (React).Component` answers true. It does not skip parentheses around the whole base expression,
// so `extends (React.Component)` answers false, because the heritage type's expression is a
// `KindParenthesizedExpression` and neither of the helper's two arms names that kind.
//
// Upstream reports on every parenthesized spelling. Measured by driving the installed build over
// the same component under six heritage clauses: `React.Component`, `(React.Component)`,
// `((React.Component))`, `(React).Component`, `Component` and `(Component)` all report. So the
// missing skip costs findings here rather than inventing them, and the fix is to unwrap before
// asking.
//
// This is worth stating carefully because the shelf's own doc comment records the OPPOSITE
// conclusion: it flags its paren-skip as a divergence on the grounds that `extends
// (React.Component)` is silent upstream. That measurement was taken against oxc, which is the
// authority for the rules that motivated the helper and is not the authority here. Both readings
// are right about their own upstream, and the two upstreams disagree.
//
// A class expression is handled by the shelf and needs nothing extra, because the parentheses in
// question are inside the heritage clause rather than around the class.
func stateInConstructorIsComponentClass(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if utilsreact.IsEs6ComponentClass(node) {
		return true
	}

	// The shelf said no. That is the final answer unless the reason is a parenthesized base, so
	// unwrap the base expressions and ask the shelf again about an equivalent unparenthesized
	// shape rather than reimplementing the base-name comparison here.
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
		// There is deliberately no `extends`-token test here, and the absence is measured rather
		// than an oversight.
		//
		// Only an `extends` clause names a base class, so testing the token reads as obviously
		// correct, and both this file and the shelf's `IsEs6ComponentClass` were written with one.
		// It is inert: this parser gives an `implements` clause `KindTypeReference` and an
		// `extends` clause `KindExpressionWithTypeArguments`, so the kind test three lines below
		// has already declined every input the token test could ever see. Probed over six heritage
		// shapes including `extends Base implements Component` and `extends (React.Component)
		// implements Component`, where both clauses are present on one class: every `implements`
		// clause came back `KindTypeReference` and no `extends` clause did.
		//
		// A mutation forcing the token test always-true survived the whole fixture set twice, once
		// before three `implements` fixtures were added for it and once after, which is what sent
		// the question to the parser instead of to another fixture. The fixtures stay because the
		// silence they assert is real and worth pinning; the guard goes because nothing can reach
		// it. If this parser ever gives an `implements` clause an expression-with-type-arguments,
		// this verdict expires and the token test has to come back.
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typeNode := range types.Nodes {
			if typeNode.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			base := typeNode.AsExpressionWithTypeArguments().Expression
			if base == nil || base.Kind != ast.KindParenthesizedExpression {
				continue
			}
			if stateInConstructorIsComponentBase(ast.SkipParentheses(base)) {
				return true
			}
		}
	}
	return false
}

// stateInConstructorIsComponentBase reports whether an unwrapped base expression names a React
// component base class.
//
// Reached only from the parenthesized path above, where the shelf has already declined and the
// parentheses are the reason. The shelf's own `isComponentBase` is unexported, so the two accepted
// spellings are restated here: a bare `Component` or `PureComponent`, and the same two qualified by
// `React`. The receiver is unwrapped too, so `((React)).Component` inside parentheses is accepted
// the way upstream accepts it.
func stateInConstructorIsComponentBase(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return stateInConstructorIsComponentBaseName(expression.Text())

	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		receiver := ast.SkipParentheses(access.Expression)
		if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier &&
			stateInConstructorIsComponentBaseName(name.Text())
	}
	return false
}

// stateInConstructorIsComponentBaseName accepts the two base classes a component may extend.
func stateInConstructorIsComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}
