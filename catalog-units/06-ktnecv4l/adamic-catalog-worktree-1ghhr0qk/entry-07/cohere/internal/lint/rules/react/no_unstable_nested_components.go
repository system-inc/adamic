package react

import (
	"regexp"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnstableNestedComponentsOptions configures the rule.
//
// Upstream's schema is a single object with three keys. Our config layer unwraps the severity tuple
// before dispatch, so the decoder receives that object directly rather than upstream's one-element
// array.
type NoUnstableNestedComponentsOptions struct {
	// AllowAsProps permits a component defined inside a prop. Absent means false.
	AllowAsProps bool `json:"allowAsProps"`

	// PropNamePattern is the glob a prop name is matched against to be treated as a render prop.
	// Absent means `render*`, which is upstream's default.
	PropNamePattern string `json:"propNamePattern"`

	// CustomValidators is in upstream's schema and is read by its prop-types machinery rather than
	// by this rule. It is accepted so a configuration carrying it decodes instead of erroring, and
	// it is deliberately unused: no branch in upstream's `validate` consults it.
	CustomValidators []string `json:"customValidators"`
}

// DefaultNoUnstableNestedComponentsOptions is the unconfigured answer.
//
// `propNamePattern` defaults to `render*` through upstream's
// `(context.options[0] || {}).propNamePattern || 'render*'`, so an absent key and an explicit empty
// string both land on `render*` there. The empty string is falsy in JavaScript, which is a
// distinction Go does not make for free and which the decoder below reproduces deliberately.
// The pattern's default is defended at THREE sites: this literal, the decoder's fold of an empty
// string, and Run's fallback for nil options. A mutation emptying any one of them SURVIVES the
// sweep, because the other two still produce `render*`. Mutating all three together fails 12 lines,
// which is what establishes the fixtures can see the default at all.
//
// That redundancy is deliberate rather than accidental: the three sites cover three different
// arrival paths, a configured value, an explicitly empty configured value, and no configuration at
// all, and the brief records that a default-inversion reaching the rule is invisible to fixtures
// built from a struct. Recorded here because a single-site survivor reads as a fixture gap and is
// not one.
func DefaultNoUnstableNestedComponentsOptions() NoUnstableNestedComponentsOptions {
	return NoUnstableNestedComponentsOptions{
		AllowAsProps:    false,
		PropNamePattern: "render*",
	}
}

// DecodeNoUnstableNestedComponentsOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` because the pattern's default is a non-zero
// value. The generic helper would leave `PropNamePattern` empty on a bare `"error"` configuration,
// and an empty pattern matches only the empty string, so every render prop would stop being
// recognised and the rule would report where upstream is silent. That is the default-inversion trap
// the brief names, arriving through a string rather than through a bool.
func DecodeNoUnstableNestedComponentsOptions(raw []byte) (any, error) {
	options := DefaultNoUnstableNestedComponentsOptions()
	if len(raw) == 0 {
		return options, nil
	}
	// Decoded into a fresh value so an absent key is distinguishable from an explicit empty string,
	// then folded onto the defaults the way upstream's `||` folds a falsy value.
	var wire NoUnstableNestedComponentsOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	options.AllowAsProps = wire.AllowAsProps
	options.CustomValidators = wire.CustomValidators
	if wire.PropNamePattern != "" {
		options.PropNamePattern = wire.PropNamePattern
	}
	return options, nil
}

// componentAsPropsInfo is appended to the message when the component sits in a prop and the option
// that would allow it is off. Copied byte for byte from upstream's constant of the same name.
const componentAsPropsInfo = " If you want to allow component creation in props, set allowAsProps option to true."

// NoUnstableNestedComponents keeps a component from being defined inside another component's render.
//
//	valid:   a component defined at module scope and used inside another
//	valid:   a component created by `useCallback` inside a parent, which upstream calls a
//	         known false negative rather than an exemption
//	valid:   `<Component renderFooter={() => <div />} />`, a render prop by name
//	valid:   `items.map((item) => <li />)`, which is a callback rather than a component
//	invalid: `function Parent() { function Nested() { return <div />; } ... }`
//	invalid: `function Parent() { const Nested = () => <div />; ... }`
//	invalid: `<Component footer={() => <div />} />` unless `allowAsProps` is set
//
// React reconciles by component type, so a component whose identity is recreated on every parent
// render is a different type every time. React unmounts the old subtree, discards its state and its
// DOM nodes, and mounts a new one. The finding is therefore about WHERE a function is defined
// relative to a component boundary rather than about what it contains.
//
// # This rule is a filter over a detection pass that already exists in this package
//
// Upstream is `Components.detect(...)`, which merges its own detection visitors ahead of the rule's
// and hands the rule a populated component registry. Every finding comes from asking whether a node
// is a component and whether an ancestor is one; the rule's own body is a list of exceptions.
//
// `no-multi-comp` already ports that detection pass as `collectDetectedComponents` and
// `statelessComponentFor`, against the same authority and the same version. This rule reaches for
// them rather than deriving component-hood a second time, for the reason the brief gives about
// extension rules: two implementations of one question are two chances to disagree about it.
//
// # The scope walk is an AST walk here, measured rather than assumed
//
// Upstream's `getParentStatelessComponent` walks `scope.upper` and calls `getStatelessComponent` on
// each `scope.block`. We have no scope analysis, and the natural substitute is the chain of
// enclosing function-like AST ancestors. Those are different objects and the substitution had to be
// measured rather than argued.
//
// Measured on 2026-08-27 by instrumenting the installed 7.37.5 build and capturing both chains at
// every call the corpus produces: 484 calls, and after collapsing the consecutive duplicates ESLint
// emits for a function's own parameter scope, the two sequences agree on all 484. Controlled by
// removing the function kinds from the AST side, which took the disagreement count to 432, so the
// comparison can distinguish the two rather than passing vacuously.
//
// # Parentheses
//
// Our parser keeps `KindParenthesizedExpression`, which estree folds away, so an ancestry walk sees
// a node upstream never sees. Every walk here goes through `semanticParentOf`, which `no-multi-comp`
// already defines to skip them. Measured against the installed build: `<C footer={(() => <div />)} />`
// reports upstream, and without the skip the port was silent on it.
//
// # No fix
//
// Upstream ships none, and there is none available: moving the component out means choosing where
// it goes and rewriting every value it closed over into a prop. That is a design decision rather
// than a repair.
var NoUnstableNestedComponents = rule.Rule{
	Name: "react/no-unstable-nested-components",

	// Reached through this package's `isPragmaCreateElementCall`, which `returnsJsx` consults to
	// decide whether a bare `createElement(...)` names the React function. Upstream answers the
	// same question with scope analysis over the file's imports.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoUnstableNestedComponentsOptions](options)
		if !ok {
			// A rule configured as a bare `"error"` is handed nil options, which is not this type.
			// Without this the zero value would apply and the empty pattern would match nothing.
			settings = DefaultNoUnstableNestedComponentsOptions()
		}
		if settings.PropNamePattern == "" {
			settings.PropNamePattern = "render*"
		}
		propNameMatches := compilePropNamePattern(settings.PropNamePattern)

		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				// The detection pass runs once over the whole file before any judgment, which is
				// what upstream's merged visitors achieve by ordering. Doing it here rather than
				// per node also means an ancestor's component-hood is known when a descendant is
				// judged, which a pre-order walk cannot otherwise guarantee for the shapes where
				// the ancestor is a wrapping call.
				detected := collectDetectedComponents(ctx, sourceFile)
				isComponent := make(map[*ast.Node]bool, len(detected))
				for _, component := range detected {
					isComponent[component.node] = true
				}

				state := &unstableNestedState{
					ctx:             ctx,
					isComponent:     isComponent,
					allowAsProps:    settings.AllowAsProps,
					propNameMatches: propNameMatches,
				}

				var visit func(*ast.Node)
				visit = func(node *ast.Node) {
					switch node.Kind {
					case ast.KindFunctionDeclaration, ast.KindArrowFunction,
						ast.KindFunctionExpression, ast.KindClassDeclaration,
						ast.KindCallExpression:
						state.validate(node)
					}
					node.ForEachChild(func(child *ast.Node) bool {
						visit(child)
						return false
					})
				}
				visit(sourceFile)
			},
		}
	},
}

// unstableNestedState carries what every judgment needs, so the helpers below read as upstream's
// free functions rather than as a chain of parameters.
type unstableNestedState struct {
	ctx         rule.Context
	isComponent map[*ast.Node]bool

	allowAsProps    bool
	propNameMatches func(string) bool
}

// validate reproduces upstream's function of the same name.
//
// The order of the exceptions is upstream's order and it is load bearing: several of them overlap,
// and `isDeclaredInsideProps` is computed BEFORE the component test because the message depends on
// it even when component-hood is established by a different route.
func (s *unstableNestedState) validate(node *ast.Node) {
	if node == nil || node.Parent == nil {
		return
	}

	isDeclaredInsideProps := s.isComponentInProp(node)

	if !s.isComponent[node] &&
		!s.isFunctionComponentInsideClassComponent(node) &&
		!isDeclaredInsideProps {
		return
	}

	switch {
	// The `allowAsProps` option, and the render-prop exemption that applies regardless of it.
	case isDeclaredInsideProps && (s.allowAsProps || s.isComponentInRenderProp(node)):
		return

	// A component created inside `items.map(...)` is a callback rather than a nested component.
	// Upstream tests the node AND its parent, which is what catches both the arrow itself and a
	// function expression sitting under one.
	case isMapCallExpression(node), isMapCallExpression(semanticParentOf(node)):
		return

	// A component returned from a hook is the hook's product, not the render's. This also covers
	// the `() => null` cleanup an effect returns.
	case s.isReturnStatementOfHook(node):
		return

	// An object whose key is a render prop holds a renderer rather than a nested component.
	case s.isDirectValueOfRenderProperty(node):
		return

	// A class component nested in a render method is reported through the class itself, so the
	// render method must not report it a second time.
	case s.isInsideRenderMethod(node):
		return

	// A detected "component" that returns no JSX is a handler that happened to be capitalized.
	case s.isStatelessComponentReturningNull(node):
		return
	}

	parentComponentNode := s.closestComponentAncestor(node)
	if parentComponentNode == nil {
		return
	}

	parentName := resolveNestedComponentName(parentComponentNode)

	// A lowercase parent is not a component React will render, so nothing nested inside it is
	// remounted by this mechanism. Upstream names `function createTestComponent()` as the case.
	if parentName != "" && !startsUppercase(parentName) {
		return
	}

	// No deduplication guard here, and that is measured rather than assumed. An earlier draft
	// carried a map keyed on the reported node, on the theory that a wrapping `memo(...)` anchor
	// could be reached from two different nodes. It cannot: this rule's walk visits each node once
	// and `validate` keys on the node it was handed, not on the anchor. Probed over three wrapper
	// shapes, including a doubly-nested `memo(forwardRef(...))`, and no node was validated more
	// than once in any of them. A mutation removing the guard survived the whole suite, which is
	// what sent me to check.
	//
	// Upstream cannot report twice either, for the same structural reason: its five listeners are
	// keyed on disjoint node types.
	//
	// The enumerated callers are this file's own `KindSourceFile` walk and nothing else. If a
	// second entry point is ever added, this verdict is void.
	message := unstableNestedMessage(parentName)
	if isDeclaredInsideProps && !s.allowAsProps {
		message += componentAsPropsInfo
	}

	s.ctx.ReportNode(node, rule.Message{
		Id:          "unstableNestedComponent",
		Description: message,
	})
}

// unstableNestedMessage reproduces upstream's `generateErrorMessageWithParentName`.
//
// The text is copied byte for byte, including the curly quotes around the parent name and the
// typographic apostrophe in "subtree’s", because the corpus asserts the whole string. The spacing
// around the name is upstream's own: a known parent contributes ` “Name” ` with spaces on both
// sides, and an unknown one contributes a single space, so the two branches differ by more than the
// name itself.
func unstableNestedMessage(parentName string) string {
	name := " "
	if parentName != "" {
		name = " “" + parentName + "” "
	}
	return "Do not define components during render. React will see a new component type on every " +
		"render and destroy the entire subtree’s DOM nodes and state " +
		"(https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move " +
		"this component definition out of the parent component" + name + "and pass data as props."
}

// closestComponentAncestor is upstream's `getClosestMatchingParent(node, context, components.get)`.
//
// Strictly an ANCESTOR search: upstream starts at `node.parent` and never considers the node
// itself, which matters because the node being judged is usually a component too and would
// otherwise be its own parent.
//
// Upstream also stops at `Program`, so a component at file scope has no parent component and is
// never reported. That is the whole reason this rule is quiet on ordinary code.
// The explicit file stop is equivalent to letting the loop run out, and a mutation removing it
// survives: `SourceFile.Parent` is nil, so the walk terminates there either way, and nothing above
// a file can be in the component map. Measured in a throwaway probe on 2026-08-27, with a control
// that asserted the opposite and failed: a parsed source file's `Parent` is nil. The probe is
// deleted; the measurement is here, which is where the next reader will look.
//
// It is kept because it states upstream's `node.parent.type === 'Program'` stop at the line, which
// is a readability claim rather than a behavioural one.
func (s *unstableNestedState) closestComponentAncestor(node *ast.Node) *ast.Node {
	for current := semanticParentOf(node); current != nil; current = semanticParentOf(current) {
		if current.Kind == ast.KindSourceFile {
			return nil
		}
		if s.isComponent[current] {
			return current
		}
	}
	return nil
}

// closestMatchingAncestor is upstream's `getClosestMatchingParent` with an arbitrary matcher.
//
// Same two properties as above: starts at the parent, stops below the file.
func closestMatchingAncestor(node *ast.Node, matches func(*ast.Node) bool) *ast.Node {
	for current := semanticParentOf(node); current != nil; current = semanticParentOf(current) {
		if current.Kind == ast.KindSourceFile {
			return nil
		}
		if matches(current) {
			return current
		}
	}
	return nil
}

// resolveNestedComponentName is upstream's `resolveComponentName`.
//
// Upstream reads `node.id.name` first, which covers a function declaration and a named function
// expression, and falls back to `node.parent.id.name` for an arrow, which covers
// `const Parent = () => ...`. A class declaration's name arrives through the first branch here
// because our parser puts it in the same `Name()` position.
func resolveNestedComponentName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindClassDeclaration,
		ast.KindClassExpression:
		if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}

	// Upstream's fallback is guarded on `ArrowFunctionExpression` specifically, so a function
	// expression assigned to a const does NOT take its name from the binding. Measured against the
	// installed build on 2026-08-27: `const Parent = function () { const N = () => <div/>; return
	// <N/>; }` reports with the no-name wording, while the same file written with an arrow parent
	// names `Parent`. Reproducing the narrower guard rather than the more useful one, because the
	// message text is asserted.
	if node.Kind == ast.KindArrowFunction {
		if parent := semanticParentOf(node); parent != nil &&
			parent.Kind == ast.KindVariableDeclaration {
			if name := parent.AsVariableDeclaration().Name(); name != nil &&
				name.Kind == ast.KindIdentifier {
				return name.Text()
			}
		}
	}
	return ""
}

// startsUppercase is upstream's `parentName[0] === parentName[0].toLowerCase()`, inverted.
//
// Deliberately NOT `hasComponentCapitalization`, which is upstream's `isFirstLetterCapitalized` and
// strips leading underscores before testing. This site does neither: it indexes character zero and
// compares it against its own lowercase, so `_Foo` is a LOWERCASE parent here and a capitalized
// name there. Measured against the installed build on 2026-08-27: a nested component inside
// `function _Parent()` is silent, and inside `function Parent()` reports, while `no-multi-comp`
// counts `_Foo` as a component. Two predicates in one plugin that look interchangeable and are not.
func startsUppercase(name string) bool {
	if name == "" {
		return false
	}
	first := []rune(name)[0]
	return string(first) != strings.ToLower(string(first))
}

// isMapCallExpression is upstream's `isMapCall`.
//
// Upstream reads `node.callee.property.name === 'map'` with no check that the node is a call at
// all, relying on the undefined chain to answer false. It does not resolve the receiver, so
// `anything.map(...)` counts and a bare `map(...)` does not.
func isMapCallExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := skipParenthesesOptional(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "map"
}

// isReturnStatementOfHook is upstream's function of the same name.
//
// Two conditions, and the second is looser than it reads. The node must sit directly under a return
// statement, and the closest enclosing CALL must have a callee whose name matches `use[A-Z0-9]`.
// Upstream reads `callExpression.callee.name`, which is undefined for a member callee, so
// `React.useCallback(() => ...)` does NOT satisfy this and is exempted by a different route.
func (s *unstableNestedState) isReturnStatementOfHook(node *ast.Node) bool {
	parent := semanticParentOf(node)
	if parent == nil || parent.Kind != ast.KindReturnStatement {
		return false
	}

	call := closestMatchingAncestor(node, func(candidate *ast.Node) bool {
		return candidate.Kind == ast.KindCallExpression
	})
	if call == nil {
		return false
	}
	callee := skipParenthesesOptional(call.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	return hookNamePattern.MatchString(callee.Text())
}

// hookNamePattern is upstream's `HOOK_REGEXP`, copied verbatim.
//
// Note it requires a character after the digit or capital, so bare `use` and `useX` differ: `useX`
// matches through `.*` being allowed to be empty, and `use` alone does not match at all.
var hookNamePattern = regexp.MustCompile(`^use[A-Z0-9].*$`)

// isComponentInRenderProp is upstream's function of the same name.
//
// Three routes, checked in upstream's order.
func (s *unstableNestedState) isComponentInRenderProp(node *ast.Node) bool {
	parent := semanticParentOf(node)

	// `{ renderFooter: () => <div /> }`, matched on the object key.
	if parent != nil && parent.Kind == ast.KindPropertyAssignment {
		if key := parent.AsPropertyAssignment().Name(); key != nil &&
			key.Kind == ast.KindIdentifier && s.propNameMatches(key.Text()) {
			return true
		}
	}

	// `<Component>{() => <div />}</Component>`, a render prop passed as children. Upstream tests
	// `node.parent.type === 'JSXExpressionContainer'` with a `JSXElement` grandparent. Ours spells
	// the container `KindJsxExpression`, and the chain is JsxExpression > JsxElement with no
	// intervening node, matching estree exactly for this shape.
	//
	// # This branch is inert, in upstream as well as here, and it is kept anyway
	//
	// A mutation disabling it SURVIVES the sweep. That is not a fixture gap: `validate` returns at
	// its first gate for every node this branch can answer true for, because a bare arrow passed as
	// children is not a detected component and is not in a prop, so the exceptions below it are
	// never consulted.
	//
	// Measured here across four shapes, a bare arrow, an arrow with a block body, an arrow
	// returning a nested component, and a named function expression: every node for which this
	// branch answers true has `isComponent=false` and `isComponentInProp=false`.
	//
	// Measured at the AUTHORITY too, rather than concluding our port had narrowed something.
	// Instrumenting the installed 7.37.5 build over the whole corpus: 185 function-like nodes
	// reach the rule, upstream's own version of this branch answers true for 2 of them, and NEITHER
	// passes `components.get`. So the exception is dead in the original as well.
	//
	// Kept rather than deleted because the brief's standing instruction is fidelity to the
	// decision: upstream states that children-as-render-prop is exempt, and a future change to the
	// detection pass could make that statement load bearing. Deleting it would silently drop a
	// judgment upstream made. The measurement is here so the next reader does not spend the sweep
	// round I spent.
	if parent != nil && parent.Kind == ast.KindJsxExpression {
		if grandparent := semanticParentOf(parent); grandparent != nil &&
			grandparent.Kind == ast.KindJsxElement {
			return true
		}
	}

	// `<Component renderFooter={() => <div />} />`, matched on the attribute name, plus the
	// explicit `children` prop.
	//
	// Ours puts a `KindJsxAttributes` list between the attribute and the element, which estree does
	// not have. That extra level does not matter here because the walk goes upward from the
	// container to the attribute, which is one step in both trees.
	container := closestMatchingAncestor(node, func(candidate *ast.Node) bool {
		return candidate.Kind == ast.KindJsxExpression
	})
	if container == nil {
		return false
	}
	attribute := semanticParentOf(container)
	if attribute == nil || attribute.Kind != ast.KindJsxAttribute {
		return false
	}
	name := attribute.AsJsxAttribute().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	propName := name.Text()
	return s.propNameMatches(propName) || propName == "children"
}

// isDirectValueOfRenderProperty is upstream's function of the same name.
//
// Narrower than `isComponentInRenderProp`'s first route despite looking identical: this one is not
// gated on the node being in a prop at all, so it exempts a bare `const rows = { render: () =>
// <div /> }` written inside a component.
func (s *unstableNestedState) isDirectValueOfRenderProperty(node *ast.Node) bool {
	parent := semanticParentOf(node)
	if parent == nil || parent.Kind != ast.KindPropertyAssignment {
		return false
	}
	key := parent.AsPropertyAssignment().Name()
	if key == nil || key.Kind != ast.KindIdentifier {
		return false
	}
	return s.propNameMatches(key.Text())
}

// isComponentInProp is upstream's function of the same name.
//
// Upstream's first branch is `isPropertyOfObjectExpressionMatcher(node)`, whose body is
// `node && node.parent && node.parent.type === 'Property'`. Despite the name, which reads as though
// it were about the property rather than about its value, that is the node's DIRECT parent, so it
// asks whether the function is the value of an object property.
//
// I first read it as a grandparent test, on the strength of the helper's name, and wrote a doc
// comment asserting the intuitive reading was wrong. The comment was the thing that was wrong.
// Measured against the installed 7.37.5 build on 2026-08-27 by evaluating the condition at every
// validated node for upstream's invalid cases 23, 30 and 32: every reported arrow has
// `parent=Property` with the condition true, and each of those findings carries the as-props
// suffix, which is only appended when this branch answers true. Under the grandparent reading all
// five of those cases went silent.
//
// That is the "never document a limit you did not measure" failure, arriving as a confident
// paragraph rather than as a missing check, and it is recorded here because a future reader
// comparing the helper's name against its body will have the same instinct.
func (s *unstableNestedState) isComponentInProp(node *ast.Node) bool {
	if parent := semanticParentOf(node); parent != nil &&
		parent.Kind == ast.KindPropertyAssignment {
		return returnsJsx(s.ctx, node)
	}

	attribute := closestMatchingAncestor(node, isJsxAttributeOfExpressionContainer)
	if attribute == nil {
		return s.isComponentInsideCreateElementProps(node)
	}
	return returnsJsx(s.ctx, node)
}

// isJsxAttributeOfExpressionContainer is upstream's matcher of the same name.
//
// The value must be an expression container, so `<C foo="bar" />` does not match. Reading the value
// through the typed accessor rather than through `Text()` matters: the brief records that
// `Node.Text()` panics on several shapes, and a JSX attribute value can be a string literal, an
// expression container, or absent entirely for `<C flag />`. A bare attribute has no value node at
// all, so this is a nil check as much as a kind check.
//
// # The container half is a live survivor, recorded rather than papered over
//
// A mutation accepting any non-nil initializer survives. Instrumented over all 77 corpus cases: 37
// JSX attributes reach this predicate and every one of them carries an expression container, so
// nothing imported can see the kind test.
//
// Two rounds of fixtures failed to close it and each taught something. The first put a nested
// component in the parent's BODY beside a string-valued attribute; that component's ancestry is
// Block then FunctionDeclaration then SourceFile, so it never walks near an attribute and the
// predicate is not consulted. The second tried to make the walk climb PAST a string-valued
// attribute toward an outer container attribute, which is what the guard would really bound, and
// both such shapes are silent upstream for an unrelated reason:
//
//	<Outer footer={() => <div />} />                      reports, with the as-props note
//	<Outer footer={<Inner title="x" render={...} />} />   silent
//	<Outer footer={<Inner title="x">{...}</Inner>} />     silent
//
// So there is no measured input where the kind test changes a verdict. This is NOT filed as an
// equivalence verdict: I could not name such an input, which is a different claim from showing none
// exists. The check is kept because it is upstream's own condition and because the nil half of it is
// load bearing regardless.
func isJsxAttributeOfExpressionContainer(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindJsxAttribute {
		return false
	}
	initializer := node.AsJsxAttribute().Initializer
	return initializer != nil && initializer.Kind == ast.KindJsxExpression
}

// isComponentInsideCreateElementProps is upstream's `isComponentInsideCreateElementsProp`.
//
// Upstream gates on `components.get(node)` first, so a function that is not already a detected
// component takes no part in this. Then it requires the closest enclosing object literal to be
// exactly the SECOND argument of the closest enclosing `createElement` call, which is the props
// position. Object identity is the test upstream uses and node identity is the same test here.
func (s *unstableNestedState) isComponentInsideCreateElementProps(node *ast.Node) bool {
	if !s.isComponent[node] {
		return false
	}

	// The kind guard is not an optimization. `createElementCallCounts` forwards to
	// `isPragmaCreateElementCall`, which opens with an unchecked `node.AsCallExpression()`; its
	// existing callers all reach it from a `KindCallExpression` listener, so it has never needed
	// one. An ancestor walk hands it every kind on the way up, and the first `VariableDeclaration`
	// above a nested component panics. Caught by upstream's own passing case 34 on the first run.
	//
	// The cost of getting this wrong is not this rule's verdict: the walk recovers per FILE, so a
	// panic here takes that file away from every rule in the tree.
	call := closestMatchingAncestor(node, func(candidate *ast.Node) bool {
		return candidate.Kind == ast.KindCallExpression && createElementCallCounts(s.ctx, candidate)
	})
	if call == nil {
		return false
	}
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) < 2 {
		return false
	}

	// Argument index 1 is the props position, which is upstream's `arguments[1]`.
	//
	// A mutation changing it to 0 survives. Instrumented over all 77 corpus cases plus probe
	// shapes: this route is reached 92 times once the paths that short-circuit in
	// `isComponentInProp`'s property branch are excluded, exactly ONE of those reaches a
	// `createElement` call carrying two or more arguments, and for that one both indices answer
	// false. Every other reach is a direct property value, which an earlier branch claims first.
	//
	// A distinguishing input needs a DETECTED component inside a props object that is not itself a
	// direct property value; three attempts to build one produced source upstream does not detect
	// as a component at all. Recorded as unresolved rather than as equivalence.
	object := closestMatchingAncestor(node, func(candidate *ast.Node) bool {
		return candidate.Kind == ast.KindObjectLiteralExpression
	})
	return object != nil && skipParenthesesOptional(arguments.Nodes[1]) == object
}

// isStatelessComponentReturningNull is upstream's function of the same name.
//
// It asks `getStatelessComponent(node)` again rather than reading the registry, because the
// registry holds the ANCHOR a component reports on and this question is about the node itself. A
// `memo(() => null)` is in the registry under the call while this asks about the arrow.
func (s *unstableNestedState) isStatelessComponentReturningNull(node *ast.Node) bool {
	component, isStateless := statelessComponentFor(s.ctx, node, nil)
	if !isStateless {
		return false
	}
	return !returnsJsx(s.ctx, component)
}

// isInsideRenderMethod is upstream's function of the same name, and here it is structurally
// unreachable rather than merely narrow.
//
// # What upstream's guard actually protects against
//
// Its test is `node.parent.type === 'MethodDefinition' && node.parent.key.name === 'render'`, which
// is the node's DIRECT parent. In estree a class method is a `MethodDefinition` whose `value` is a
// separate `FunctionExpression`, so exactly one node in a class satisfies this: the render method's
// own function expression. Upstream's comment calls it "prevent reporting nested class components
// twice", and that is what it does: the render method's function expression is itself detected as a
// component sitting inside a class component, so without this guard the render method would report
// alongside whatever is nested in it.
//
// Measured against the installed 7.37.5 build on 2026-08-27, by evaluating upstream's own two
// conditions at every node the rule validates, for a class whose render nests a class component:
//
//	ClassDeclaration@2   parent=Program          isInsideRenderMethod=false
//	FunctionExpression@3 parent=MethodDefinition isInsideRenderMethod=true
//	ClassDeclaration@4   parent=BlockStatement   isInsideRenderMethod=false
//	FunctionExpression@4 parent=MethodDefinition isInsideRenderMethod=true
//
// The nested class, which is the thing that reports, answers false. Only the two method function
// expressions answer true.
//
// # Why there is nothing to reproduce
//
// Our parser has no separate node for a method's value: a `KindMethodDeclaration` carries its own
// parameters and body, so the node estree gates on does not exist here and is never handed to a
// listener. This rule's own walk anchors on five kinds and `KindMethodDeclaration` is not among
// them, matching upstream's five listeners, which name `FunctionExpression` rather than the method.
//
// So the guard has no input in this tree. It is written as a predicate that always answers false
// rather than omitted, because omitting it silently would leave the next reader comparing this file
// against upstream and finding a missing exception with no explanation.
//
// # What I got wrong, recorded because the wrong version passed forty cases
//
// The first version asked "is this node somewhere inside a render method", walking up through the
// enclosing block to the method. That reads as the same question and is not: it exempted every
// component nested in a render method, which is precisely the population this rule exists to
// report. It made four of upstream's failing cases silent while all forty passing cases stayed
// green, and only the corpus caught it.
func (s *unstableNestedState) isInsideRenderMethod(node *ast.Node) bool {
	// Upstream additionally requires the parent component to be a class, but the parent test above
	// can never pass here, so the class test is not consulted. Both halves are named in the doc
	// comment rather than written as dead code.
	//
	// A method's value has no node of its own in this parser, so nothing can satisfy upstream's
	// `node.parent.type === 'MethodDefinition'`. `TestNoUnstableNestedComponentsRenderMethodGuard`
	// pins that: it walks a class whose render nests a component and asserts no validated node has
	// a method declaration as its semantic parent.
	return false
}

// isFunctionComponentInsideClassComponent is upstream's function of the same name.
//
// Upstream's own comment says why it exists: its detection pass fails to find a function component
// declared inside a class component, so this recovers those. All four conjuncts are required, and
// the third is what confines it to classes.
//
// # A mutation making this always false SURVIVES, and the reason is a parser difference
//
// Measured against the installed 7.37.5 build on 2026-08-27 by evaluating upstream's four conjuncts
// beside `components.get` at every validated node across the whole corpus: 196 nodes, the recovery
// fires for 8 of them, and it is DECISIVE, meaning `components.get` was false, for exactly 3.
//
// Those 3 are in upstream's invalid cases 10, 11 and 29, and in every one the decisive node is the
// class component's RENDER METHOD FUNCTION EXPRESSION rather than the nested component. Upstream
// recovers that node into the rule, and then `isInsideRenderMethod` immediately drops it again, so
// the pair is a no-op that exists because estree gives a method's value a node of its own.
//
// Our parser gives it none, so neither half has any input here, which is the same reason
// `isInsideRenderMethod` above is a constant. All three of those cases report correctly without the
// recovery contributing anything.
//
// Kept rather than deleted, on the same reasoning as the children-render-prop branch: it states
// upstream's judgment, and a change to `collectDetectedComponents` could make it load bearing.
// Three probe shapes were built to find an input where it decides something here, a function
// component in a render method, one in an ordinary method, and an arrow in a render method, and our
// detection pass answers true for all three before this is consulted.
func (s *unstableNestedState) isFunctionComponentInsideClassComponent(node *ast.Node) bool {
	parentComponentNode := s.parentComponentOf(node)
	if parentComponentNode == nil {
		return false
	}
	if parentComponentNode.Kind != ast.KindClassDeclaration &&
		parentComponentNode.Kind != ast.KindClassExpression {
		return false
	}

	parentStateless := s.parentStatelessComponentOf(node)
	if parentStateless == nil {
		return false
	}
	if _, isStateless := statelessComponentFor(s.ctx, parentStateless, nil); !isStateless {
		return false
	}
	return returnsJsx(s.ctx, node)
}

// parentComponentOf is upstream's `getParentComponent`, which tries the class forms and then falls
// back to the stateless walk.
//
// # This walk INCLUDES the node itself, matching upstream
//
// Upstream's three components all start from `getScope(context, node)`, and ESLint's scope for a
// class or a function-like node is the scope that node CREATES, not the one it sits in. So
// `getParentES6Component(theClass)` returns that class, and `getParentStatelessComponent(theFn)`
// returns that function.
//
// Measured against the installed 7.37.5 build on 2026-08-27: over the whole corpus, 268 function-
// like nodes reached `getParentStatelessComponent` and `getScope(node).block === node` for all 268,
// with zero exceptions. Instrumenting the real helpers on upstream's invalid case 14 gives
// `ClassDeclaration@2: es6=ClassDeclaration@2` and
// `FunctionDeclaration@4: stateless=FunctionDeclaration@4`, both naming themselves.
//
// An exclusive walk reads as the obviously correct one, because "parent" is in the name and because
// `closestComponentAncestor` below genuinely IS exclusive. They are not the same walk:
// `getClosestMatchingParent` starts at `node.parent` explicitly in its own source, while these
// three start at a scope. Two walks in one rule, opposite on this point, and only one of them says
// so in its name.
//
// # What is NOT established: whether the difference is observable here
//
// An earlier revision of this comment claimed the exclusive version cost nine of upstream's failing
// cases. That was wrong, and it is recorded rather than deleted because it is the exact shape the
// brief warns about, a confident divergence written from attribution rather than from measurement.
// Both under-report groups were fixed by `isInsideRenderMethod` and `isComponentInProp`; running
// the whole corpus with BOTH walks made exclusive again passes all 77.
//
// A mutation flipping this walk back to exclusive therefore SURVIVES the sweep, and no fixture was
// added for it, because none could be justified: three candidate shapes were built to separate the
// two, a function component directly in a render method, one behind a plain helper function, and an
// arrow inside an arrow, and all three produce one finding under both versions and match upstream.
//
// This is deliberately not filed as an equivalence verdict. The only caller requires the result to
// be a CLASS, and a node being validated is never the class enclosing it, so inclusivity cannot
// change that half; the stateless half then asks `statelessComponentFor` about whatever it found,
// and for the inclusive case that is the node itself, which is already a component by construction.
// Both routes reach the same verdict, which is the "same answer by a different path" shape the
// brief names, and I could not prove no input separates them.
//
// The inclusive spelling is kept because it is what upstream does and the measurement above is
// solid, not because the difference was shown to matter.
func (s *unstableNestedState) parentComponentOf(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = semanticParentOf(current) {
		if isComponentClass(current) {
			return current
		}
		if isCreateReactClassCall(current) {
			return current
		}
	}
	return s.parentStatelessComponentOf(node)
}

// parentStatelessComponentOf is upstream's `getParentStatelessComponent`, written as an AST walk.
//
// Upstream walks `scope.upper` calling `getStatelessComponent(scope.block)`. The equivalence of
// that chain to this one is measured rather than assumed; the measurement and its control are
// recorded in this file's header comment.
//
// The node itself is INCLUDED, matching upstream. See `parentComponentOf` for the measurement:
// ESLint's scope for a function-like node is the scope it creates, so the first thing upstream's
// loop asks about is the node itself.
func (s *unstableNestedState) parentStatelessComponentOf(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = semanticParentOf(current) {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			if component, isStateless := statelessComponentFor(s.ctx, current, nil); isStateless {
				return component
			}
		}
	}
	return nil
}

// compilePropNamePattern turns upstream's minimatch pattern into a matcher.
//
// Upstream calls `minimatch(text, pattern)` on a prop name. minimatch is a PATH matcher and its
// extra machinery is all about paths: `/` separators, a leading dot being hidden, brace expansion,
// extended globs. A prop name is a JavaScript identifier, so none of that alphabet can appear.
//
// Measured on 2026-08-27 against minimatch 3.1.5, the version the installed plugin resolves:
// 16,854 comparisons over six patterns, including `render*`, `*Renderer` and `*`, against every
// string up to four characters over the identifier alphabet plus the corpus's own names. Zero
// disagreements. Controlled by disabling the star expansion, which took the disagreement count to
// 2,815, so the comparison can tell the two apart.
//
// The measurement is confined to identifier-shaped names, and that is exactly the input this rule
// feeds it: an object key read as `KindIdentifier` and a JSX attribute name read as
// `KindIdentifier`. A computed key or a namespaced attribute never reaches here, because both are
// declined by kind at the call site rather than being stringified.
//
// `Run` is called once per file, so compiling here compiled the same pattern once per file: 179K
// objects on a cold ahra run. A matcher depends only on its pattern, so each is compiled once and
// kept in propNamePatternMatchers. Keyed by the pattern rather than held in one variable, because a
// run can lint under more than one configuration.
func compilePropNamePattern(pattern string) func(string) bool {
	if cached, found := propNamePatternMatchers.Load(pattern); found {
		return cached.(func(string) bool)
	}
	matcher, _ := propNamePatternMatchers.LoadOrStore(pattern, buildPropNamePattern(pattern))
	return matcher.(func(string) bool)
}

// propNamePatternMatchers holds compilePropNamePattern's matcher for each pattern seen.
var propNamePatternMatchers sync.Map

// buildPropNamePattern compiles one pattern, for compilePropNamePattern to keep.
func buildPropNamePattern(pattern string) func(string) bool {
	var builder strings.Builder
	builder.WriteString("^")
	for _, character := range pattern {
		switch character {
		case '*':
			builder.WriteString(".*")
		case '?':
			builder.WriteString(".")
		default:
			builder.WriteString(regexp.QuoteMeta(string(character)))
		}
	}
	builder.WriteString("$")

	compiled, err := regexp.Compile(builder.String())
	if err != nil {
		// Unreachable for any pattern this builder produces: every literal is escaped and the two
		// metacharacters expand to valid fragments. Declining rather than panicking, because a rule
		// that takes the run down over a configuration string is worse than one that stops matching
		// render props.
		return func(string) bool { return false }
	}
	return compiled.MatchString
}
