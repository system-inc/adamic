package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoDidMountSetState = rule.Message{
	Id: "noDidMountSetState",
	Description: "`setState` is called in `componentDidMount`. The component has already " +
		"committed to the screen by then, so setting state schedules a second render before " +
		"the browser has painted, and the first render's output is thrown away. The user can " +
		"see the intermediate frame as a flash, and any layout read between the two renders " +
		"thrashes. Compute the value in the constructor or in `getDerivedStateFromProps` so " +
		"the first render is already correct.",
}

// NoDidMountSetState flags `this.setState(...)` inside a component's `componentDidMount`.
//
//	valid:   class H extends React.Component { componentDidUpdate() { this.setState({}); } }
//	valid:   createReactClass({ componentDidMount: function() { this.someHandler = this.setState; } })
//	valid:   createReactClass({ componentDidMount: function() { setTimeout(() => this.setState({})); } })
//	valid:   function Hello() { this.setState({ data: 1 }); }          (no enclosing component)
//	invalid: class H extends React.Component { componentDidMount() { this.setState({}); } }
//	invalid: createReactClass({ componentDidMount: function() { this.setState({}); } })
//	invalid: createReactClass({ componentDidMount: function() { setTimeout(() => this.setState({})); } })
//	                                                                   (under "disallow-in-func")
//
// Ported from `react/no-did-mount-set-state`, read against oxc's `no_did_mount_set_state.rs` and
// the shared walk it delegates to at `oxc/crates/oxc_linter/src/utils/react.rs:606`,
// `function_count_before_lifecycle_component`. There is no `eslint-plugin-react` source in this
// tree to read as a second opinion, so oxc is the sole authority and the corpus is the only check
// on it.
//
// # The whole rule is one question, and it is not the one the name asks
//
// The name reads as "is this `setState` inside `componentDidMount`", and answering that question is
// how a port gets this wrong. Upstream asks a *counted* version of it: walking outward from the
// call, how many function scopes are crossed before reaching the property that names
// `componentDidMount`. One means the call is written directly in the lifecycle body, and reports.
// More than one means the call is inside a callback or a nested function declaration, and is
// silent unless the option asks for it.
//
// That count is why the boundary cannot be guessed. Five separate passing cases upstream ships are
// a `setState` genuinely lexically inside `componentDidMount` -- an event callback, a nested
// function declaration, an arrow passed to `this.handleEvent`, a `setTimeout` arrow, and a promise
// `.then` arrow -- and every one of them is clean by default. A port implementing "inside
// componentDidMount" reports all five. Meanwhile the fail corpus includes a `setState` inside an
// `if` block and one inside the consequent of a conditional expression, which are *not* function
// scopes and therefore still count as one. So the boundary is function scopes crossed and nothing
// else: blocks, conditionals, loops and try bodies are all transparent to it.
//
// The counting also has to stop. Once the lifecycle property is found the walk stops counting but
// keeps climbing, because the count must not pick up the `render: function() {}` sibling or the
// arrow a whole component is assigned to. Upstream spells that as an `in_lifecycle` latch and this
// reproduces it.
//
// # Three ways to write the lifecycle method, and our AST splits them differently
//
// oxc matches three node kinds by static key name, which map onto four of ours. The mapping was
// taken from `no_is_mounted.go` in this package, which measured the same correspondence against
// oxlint for its own ancestor set, and then extended by one kind that rule deliberately excludes:
//
//	oxc ObjectProperty      -> KindPropertyAssignment     { componentDidMount: function() {} }
//	                        -> KindMethodDeclaration      { componentDidMount() {} }  (shared)
//	oxc MethodDefinition    -> KindMethodDeclaration      class { componentDidMount() {} }
//	oxc PropertyDefinition  -> KindPropertyDeclaration    class { componentDidMount = () => {} }
//
// `KindPropertyDeclaration` is the one that differs from the neighbouring rule and it is not a
// judgment call here: oxc's `is_lifecycle_component_method` lists `PropertyDefinition` explicitly,
// and upstream ships `componentDidMount = () => {}` as a *failing* case. `no-is-mounted` excludes
// the same kind because oxc's match there omits it. Two rules in one package, opposite answers,
// both correct, and each one only checkable at its own reference.
//
// Accessors are absent from all of this, matching oxc: `get componentDidMount() {}` is neither an
// `ObjectProperty` nor a `MethodDefinition` in its tree, and nothing in the corpus writes one.
//
// # The component gate is load-bearing and is what the walk returns nothing for
//
// The walk climbs past the lifecycle method until it finds a component, and answers nothing if it
// reaches the top of the file without one. Upstream pins this with `function Hello() {
// this.setState({ data: 123 }); }`, which is a plain function containing exactly the call this rule
// is about and is clean. So a `componentDidMount` written on an ordinary object at the top level of
// a file reports nothing, which is the same syntactic-versus-component distinction `no-is-mounted`
// resolves in the opposite direction. This rule really does ask about components.
//
// But it asks a *looser* question than it looks, and this was measured on the release binary rather
// than reasoned about. Nothing checks that the lifecycle property is a member OF the component: the
// walk only requires that some ancestor further out is one. So a plain object literal carrying a
// `componentDidMount`, declared inside the render method of a real component, reports. That is
// upstream being loose rather than upstream being right, and it is reproduced rather than improved
// on, because a differential run against oxlint is what this port has to survive.
//
// The factory-name set is narrower than the shelf helper this rule calls. See
// `isEs5ComponentCallStrict` below: oxc accepts only `createReactClass` and
// `React.createReactClass`, while `internal/utilities/react` also accepts `createClass`. Following the
// shelf would report on clean upstream code.
//
// # Where the finding points
//
// oxc reports `call_expr.callee.span()`, not the call. For `this.setState({ data: 1 })` that is
// thirteen characters ending before the open parenthesis, which the snapshot confirms by underlining
// exactly `this.setState`. The neighbouring `no-is-mounted` reports its whole call including the
// parentheses, so the two rules in this package genuinely differ here and a message-id assertion
// cannot see it. Every fixture below asserts the sliced text.
var NoDidMountSetState = rule.Rule{
	// No namespace prefix. The config writes `react/no-did-mount-set-state` and the parity guard
	// strips the namespace on a `/` boundary, so `react-no-did-mount-set-state` would match no
	// inventory entry, lint no files, and still pass every fixture in this package.
	Name: "react/no-did-mount-set-state",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// No file gate. oxc gated this on `source_type().is_jsx()`, and that gate came along when the
		// rule was ported from oxc, but the authority here is eslint-plugin-react, which does not
		// gate on the file name at all. A React class component in a `.ts` file is ordinary and
		// legal, and the gate made the rule silent on 3,359 of them against 2,118 it could see.
		// Measured before removal: the same class reported in `.tsx` and produced nothing in `.ts`.

		// An unconfigured rule gets the zero value, which is upstream's `allow-in-func` default.
		settings, _ := rule.OptionsAs[NoMethodSetStateOptions](options)

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := node.AsCallExpression().Expression
				if !isThisSetStateCallee(callee) {
					return
				}

				functionsCrossed, insideComponent := functionCountBeforeLifecycleMethod(node, "componentDidMount")
				if !insideComponent {
					return
				}

				// One means written directly in the lifecycle body. Anything higher crossed a
				// function scope on the way out, which is the option's subject.
				if functionsCrossed > 1 && !settings.DisallowInFunc {
					return
				}

				// The callee, not the call. See the doc above: thirteen characters for
				// `this.setState`, which the upstream snapshot underlines and the fixtures slice.
				ctx.ReportNode(callee, messageNoDidMountSetState)
			},
		}
	},
}

// isThisSetStateCallee reports whether a callee reads `setState` directly off `this`.
//
// oxc requires `member_expr.object()` to be a `ThisExpression` and `static_property_name()` to
// answer `setState`. `static_property_name` resolves a computed key as well as a dotted one, so
// `this["setState"]()` matches upstream and matches here, while `this[setState]()` names some other
// property through a variable and does not.
//
// The object must be `this` itself, so `that.setState()` and `this.child.setState()` are both
// declined, the second because its object is a property access rather than the `this` keyword.
// Neither appears in the corpus and both are pinned by an invented fixture below.
//
// Note what this does NOT do: it never reads `Text()` on the callee node. A property access
// expression panics outright on `Text()` in our shim, and `this.setState` is exactly that kind, so
// the name is taken off the access's own name node after checking that node's kind.
func isThisSetStateCallee(callee *ast.Node) bool {
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if !isThisExpression(access.Expression) {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "setState"

	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if !isThisExpression(access.Expression) {
			return false
		}
		return staticSetStateName(access.ArgumentExpression)
	}
	return false
}

// isThisExpression reports whether an expression is the `this` keyword.
//
// Parentheses are not skipped, matching oxc: its `matches!(member_expr.object(),
// Expression::ThisExpression(_))` is a shape match on the node as written, and `(this).setState()`
// is a `ParenthesizedExpression` there as it is here. Upstream is narrow rather than right about
// this, and the narrowness is reproduced because a differential run against oxlint is what this
// port has to survive.
func isThisExpression(expression *ast.Node) bool {
	return expression != nil && expression.Kind == ast.KindThisKeyword
}

// staticSetStateName reports whether a computed key names `setState` without evaluation.
//
// This is oxc's `static_property_name()` narrowed to the one name this rule compares against. A
// string literal and a no-substitution template both answer; an interpolated template is a
// different node kind and falls through, and an identifier subscript is a variable holding some
// other name. oxc also answers for a regular-expression literal, which is absent here for the same
// reason `no-string-refs` omits it: no regular expression renders as `setState`, so the branch
// could not change a verdict and no fixture could guard it.
func staticSetStateName(argument *ast.Node) bool {
	if argument == nil {
		return false
	}
	switch argument.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return argument.Text() == "setState"
	}
	return false
}

// functionCountBeforeLifecycleMethod counts function scopes crossed on the way out to a named
// lifecycle method, and reports whether a component encloses the whole thing.
//
// This is oxc's `function_count_before_lifecycle_component` at our AST's spelling
// (`oxc/crates/oxc_linter/src/utils/react.rs:606`). The two-value return replaces its `Option<usize>`:
// the boolean is `Some` versus `None`, which is to say whether a component was ever reached, and the
// count is meaningless when it is false.
//
// The walk has three phases running in one pass, and the latch between the first two is the part a
// reimplementation drops:
//
//	before the lifecycle method   every function-like ancestor increments the count
//	after it                      counting stops, climbing continues
//	at a component                answer
//
// Without the latch the count picks up whatever function the component itself is written inside,
// which for `var Hello = createReactClass({...})` is none but for a component defined inside a
// factory or a module wrapper is one or more, and every direct `setState` in such a file would fall
// out of the report as though it were nested.
//
// The count starts at the call's parent rather than the call, matching upstream's `.skip(1)`. A
// call expression is not function-like so the skip changes nothing here, and it is kept because
// upstream's is load-bearing for the shape rather than for this rule.
//
// # Where our AST loses a count that oxc has, which is the whole reason this function exists
//
// oxc's `MethodDefinition` is a wrapper whose `value` is a separate `Function` node
// (`oxc_ast/src/ast/js.rs:2247`), so a class method contributes TWO ancestors to its walk and the
// inner one is function-like. Our `KindMethodDeclaration` is a single node carrying its own body,
// so the same source contributes one ancestor and it is not in the counted set. Measured rather
// than reasoned, by logging every ancestor kind of `this.setState` in each of the four shapes the
// corpus writes:
//
//	class { componentDidMount() { HERE } }       oxc 1, naive walk here 0
//	class { componentDidMount = () => { HERE } } oxc 1, naive walk here 1   (the arrow is real)
//	createReactClass({ m: function() { HERE } }) oxc 1, naive walk here 1   (the expression is real)
//	the same, one callback deeper                oxc 2, naive walk here 2
//
// So only the class-method row is short, and it is short by exactly one. Left uncorrected the
// numbers are off by one for the entire ES6 half of the rule: a direct `setState` in a class
// method counts 0 and still reports, so the imported failing cases all pass, while an arrow inside
// that method counts 1 and reports too. That is upstream's two clean cases
// "setState inside an arrow passed to a method" and "setState inside a promise then arrow"
// reported as violations, which is what the first run of this port did.
//
// The correction is applied where the latch is set, because that is the only place the missing
// node's identity is known: the lifecycle property has just been recognized and its kind says
// whether it was the function as well as the name. A property assignment and a property
// declaration are not corrected, because in those shapes the function is a real separate node the
// walk has already counted.
func functionCountBeforeLifecycleMethod(node *ast.Node, lifecycleMethodName string) (int, bool) {
	functionCount := 0
	inLifecycle := false

	for current := node.Parent; current != nil; current = current.Parent {
		if !inLifecycle {
			if isLifecycleComponentMethod(current, lifecycleMethodName) {
				inLifecycle = true
				// A class method or accessor is one node here and two in oxc, so its function
				// scope is counted at the latch rather than by the walk. See the doc above.
				switch current.Kind {
				case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
					functionCount++
				}
			} else if isFunctionScope(current) {
				functionCount++
			}
		}

		// Deliberately not an `else`: upstream re-tests this in the same iteration that sets the
		// latch, so a component whose own node is also the lifecycle property would answer here.
		// No such shape exists, but the ordering is upstream's and is kept rather than tidied.
		if inLifecycle && (isEs5ComponentCallStrict(current) || utilsreact.IsEs6ComponentClass(current)) {
			return functionCount, true
		}
	}
	return 0, false
}

// isFunctionScope reports whether an ancestor is one of the kinds upstream counts.
//
// Exactly two kinds, matching oxc's `AstKind::Function(_) | AstKind::ArrowFunctionExpression(_)`,
// where `Function` is its single node for a declaration and an expression alike. Our tree spells
// that as three.
//
// This is narrower than `scope.EnclosingFunctionLike`, which is why that shelf helper is not used
// despite answering a question with the same name. It includes methods, accessors and constructors,
// and a method is precisely what `componentDidMount() {}` is: counting it would add one to every
// class-component finding and silence the rule on the entire ES6 half of the corpus. The shelf
// function is correct for "what function am I inside" and this rule is not asking that. Read the
// body rather than the name.
func isFunctionScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	// A method is a function scope too, and leaving it out was a live false positive. oxc splits a
	// `MethodDefinition` into two nodes and counts the inner `Function`; our `KindMethodDeclaration`
	// fuses them, so the latch above counts the lifecycle method's own scope by hand. That handled
	// the method the walk STOPS at and silently skipped every method it merely passes through, so
	// `setState` inside an object method declared in the lifecycle body came out one scope shallow
	// and reported where upstream is silent. Measured against the release binary with a control:
	//
	//	componentDidMount() { const h = { run() { this.setState({}) } }; h.run() }
	//
	// is silent at oxlint and reported here before this arm existed. Found by the porter of
	// `no-will-update-set-state`, whose own sweep exposed it in their copy of this walk.
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	}
	return false
}

// isLifecycleComponentMethod reports whether a node is the property naming the lifecycle method.
//
// The kind set is oxc's `ObjectProperty | MethodDefinition | PropertyDefinition` at our spelling.
// See the rule doc for the mapping and for why `KindPropertyDeclaration` belongs here while the
// neighbouring `no-is-mounted` correctly excludes it.
//
// The name is read by kind rather than through `Text()` on the name node, because a computed key is
// a `ComputedPropertyName` and `Text()` panics on it outright. A computed lifecycle name is
// therefore silent, which matches oxc: its `key.static_name()` resolves a computed *string* key, so
// `{["componentDidMount"]: function() {}}` is a genuine divergence, reproduced as silence and
// pinned by a fixture. Nothing in the corpus writes one and the shape does not occur in real code,
// so the divergence is stated rather than closed.
func isLifecycleComponentMethod(node *ast.Node, lifecycleMethodName string) bool {
	var name *ast.Node
	switch node.Kind {
	case ast.KindPropertyAssignment:
		name = node.AsPropertyAssignment().Name()
	case ast.KindMethodDeclaration:
		name = node.AsMethodDeclaration().Name()
	case ast.KindPropertyDeclaration:
		name = node.AsPropertyDeclaration().Name()
	case ast.KindGetAccessor:
		name = node.AsGetAccessorDeclaration().Name()
	case ast.KindSetAccessor:
		name = node.AsSetAccessorDeclaration().Name()
	default:
		return false
	}
	if name == nil {
		return false
	}
	// `TryGetTextOfPropertyName` checks kind before reading text, so it does not panic on a
	// `ComputedPropertyName` the way `Text()` does, and it resolves the identifier, string-literal
	// and computed-holding-a-literal spellings in one call. That set is oxc's `static_name()`
	// (`oxc_ast/src/ast_impl/js.rs:465`) for every spelling this comparison can match.
	text, resolved := ast.TryGetTextOfPropertyName(name)
	return resolved && text == lifecycleMethodName
}

// isEs5ComponentCallStrict is `IsEs5ComponentCall` narrowed to the factory name oxc accepts.
//
// The shelf helper is deliberately wider than this rule's upstream. `internal/utilities/react` accepts
// `createClass` and `React.createClass` as well as the `createReactClass` spellings, on the
// reasoning that rules meet both halves of the ecosystem. oxc does not: `CREATE_CLASS` at
// `oxc_linter/src/utils/react.rs:554` is the single constant `"createReactClass"`, and both arms of
// its `is_es5_component` key on that one constant, so neither the bare nor the namespaced
// `createClass` spelling is a component to this rule.
//
// Measured rather than inferred. `var D = createClass({ componentDidMount: function() {
// this.setState({data:4}); } });` and its `React.createClass` twin are both clean on the release
// oxlint binary, while `createReactClass` and `React.createReactClass` both report.
//
// Narrowed locally rather than in the shelf, because other rules call the shared helper and may
// depend on the wider set. Whether the shelf should be narrowed once for everyone is a question for
// the census, not for this rule to answer by changing a helper it does not own.
//
// The parenthesized receiver is left to the shelf on purpose: `(createReactClass)({...})` reports
// upstream, because oxc's identifier arm goes through `get_identifier_reference()` which does skip
// parentheses, and our helper skips them too. So the two agree there and only the name set differs.
func isEs5ComponentCallStrict(node *ast.Node) bool {
	if !utilsreact.IsEs5ComponentCall(node) {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "createReactClass"
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "createReactClass"
	}
	return false
}
