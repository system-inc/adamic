package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var (
	messageJsxKeyMissingIterKey = rule.Message{
		Id: "missingIterKey",
		Description: "This element is produced by an iterator and carries no `key`. React reconciles a " +
			"list by position when it has no key, so inserting or reordering an item makes every " +
			"element after it adopt the previous one's state: focus jumps, inputs keep the wrong " +
			"text, and animations restart on rows that did not change. Give each element a `key` " +
			"that identifies the item rather than its position.",
	}
	messageJsxKeyMissingIterKeyUsePragma = rule.Message{
		Id: "missingIterKeyUsePrag",
		Description: "This element is produced by an iterator and is written as a shorthand fragment, " +
			"which has no syntax for a `key`. React reconciles the list by position instead, so " +
			"inserting or reordering an item makes the elements after it adopt the wrong state. " +
			"Write `React.Fragment` in full, which does accept a `key`.",
	}
	messageJsxKeyMissingArrayKey = rule.Message{
		Id: "missingArrayKey",
		Description: "This element sits in an array literal and carries no `key`. React reconciles " +
			"the array by position when it has no key, so changing the order or the length makes " +
			"the surviving elements adopt their neighbours' state. Give each element a `key` that " +
			"identifies it rather than its position.",
	}
	messageJsxKeyMissingArrayKeyUsePragma = rule.Message{
		Id: "missingArrayKeyUsePrag",
		Description: "This element sits in an array literal and is written as a shorthand fragment, " +
			"which has no syntax for a `key`. React reconciles the array by position instead. " +
			"Write `React.Fragment` in full, which does accept a `key`.",
	}
	messageJsxKeyBeforeSpread = rule.Message{
		Id: "keyBeforeSpread",
		Description: "The `key` prop is written after a `{...spread}`. Under the new JSX transform " +
			"`key` is read out of the props object rather than passed separately, so a spread that " +
			"also carries a `key` silently wins over the one written here and the element is " +
			"reconciled by the wrong identity. Move `key` before every spread.",
	}
	messageJsxKeyNonUniqueKeys = rule.Message{
		Id: "nonUniqueKeys",
		Description: "Two elements in this collection carry the same `key`. React uses the key as " +
			"the element's identity, so a repeated key makes it reconcile two different items as " +
			"one: state, focus and refs follow the wrong element and one of them is dropped. Give " +
			"each element a key that is unique within its collection.",
	}
)

// JsxKeyOptions is the decoded option object.
//
// Three fields, matching `meta.schema`'s three properties exactly, and all three default to false,
// which is Go's zero value. So unlike this batch's sibling rule there is no default inversion here,
// and the decoder is the generic one.
type JsxKeyOptions struct {
	// CheckFragmentShorthand turns on the two `UsePrag` messages.
	//
	// A shorthand fragment cannot take a key at all, so the rule is silent about it unless asked.
	// False by default and upstream keeps it false.
	CheckFragmentShorthand bool `json:"checkFragmentShorthand"`

	// CheckKeyMustBeforeSpread reports a `key` written after a `{...spread}`.
	//
	// False by default. Note that this is checked on elements that DO have a key, which is the
	// opposite population from the rest of the rule.
	CheckKeyMustBeforeSpread bool `json:"checkKeyMustBeforeSpread"`

	// WarnOnDuplicates reports two elements in one collection sharing a key.
	//
	// False by default. It is the only judgment here that needs to see a whole collection at once
	// rather than one element.
	WarnOnDuplicates bool `json:"warnOnDuplicates"`
}

// DecodeJsxKeyOptions turns the configured JSON into the struct the rule reads.
//
// Exported so fixtures can drive the same path the config drives. Every default is false, so this
// is a thin wrapper over `rule.UnmarshalOptions` plus a nil path: `rule.DecodeOptionsInto` ERRORS on empty
// input, and a rule configured as a bare `"error"` is handed exactly that, so the generic decoder
// would refuse a configuration upstream accepts.
func DecodeJsxKeyOptions(raw []byte) (any, error) {
	var options JsxKeyOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// JsxKey flags an element in a collection or an iterator that carries no `key`.
//
//	valid:   [<App key={0} />, <App key={1} />];
//	valid:   [1, 2, 3].map(x => <App key={x} />);
//	valid:   <App />;
//	valid:   [1, 2, 3].map(someFn);
//	valid:   React.Children.toArray([1, 2, 3].map(x => <App />));
//	invalid: [<App />];
//	invalid: [1, 2, 3].map(x => <App />);
//	invalid: Array.from([1, 2, 3], x => <App />);
//
// Ported from `react/jsx-key` in `eslint-plugin-react`, which is the implementation that defined
// this rule. Its 37 clean and 30 reporting cases were extracted from the clone and byte-verified
// through a second independent extraction.
//
// # THE CLONE AND THE INSTALLED BUILD DISAGREE, and this port follows the CLONE
//
// This is the one rule in its batch where the two authorities differ, and the difference is a whole
// branch rather than a detail. The clone's `checkFunctionsBlockStatement` inspects the returned
// expression and descends into a conditional or a logical:
//
//	if (argument.type === 'ConditionalExpression') { ...consequent, ...alternate }
//	else if (argument.type === 'LogicalExpression' && isJSX(argument.right)) { ...right }
//	else { checkIteratorElement(argument) }
//
// The installed build, 7.37.5, passes `returnStatement.argument` straight to
// `checkIteratorElement`, which handles only a JSX element or fragment and silently ignores
// anything else. So SIX of upstream's own reporting cases are silent on the installed build:
//
//	[1,2,3].map((item) => { return c ? <div/> : <span/>; })     clone 2, installed 0
//	[1,2,3].map(function(item) { return c ? <div/> : <span/>; }) clone 2, installed 0
//	Array.from([1,2,3], (item) => { return c ? <div/> : <span/>; }) clone 2, installed 0
//	the Fragment-wrapped variant of the first                    clone 2, installed 0
//	[1,2,3].map(x => { return x && <App />; });                  clone 1, installed 0
//	[1,2,3].map(x => { return x || y || <App />; });             clone 1, installed 0
//
// Measured on the installed build, and confirmed by diffing the two rule files: the clone carries
// unreleased commits on top of the 7.37.5 tag while still reporting that version string.
//
// **The clone is followed** because it is the implementation that defines the rule and it is the
// one whose corpus states these verdicts. The installed build is a snapshot of it that is behind.
// Every one of the six is a fixture, marked, so a later reader comparing our output against the
// installed ESLint sees exactly which six inputs are expected to differ and why, rather than
// finding six unexplained extra findings.
//
// Note what does NOT differ: the ARROW-BODY forms, `map(x => x && <App/>)` and
// `map(x => x ? <A/> : <B/>)`, report on both, because `checkArrowFunctionWithJSX` has always had
// the conditional and logical arms. Only the BLOCK-bodied path was missing them. That pairing is
// what makes the divergence easy to misread as a bug in the port.
//
// # Four listeners, and what each one is actually anchored on
//
// Upstream registers five selectors and two of them are one thing:
//
//	ArrayExpression, JSXElement > JSXElement    the collection arm, both anchors in one handler
//	JSXFragment                                 the shorthand-fragment-in-an-array arm
//	CallExpression[...map]                      the iterator arm
//	CallExpression[...from]                     the iterator arm again, second argument
//	Children.toArray(...)                       a suppression scope, not a judgment
//
// The collection arm's two anchors do NOT report the same things. Its `missingArrayKey` finding is
// guarded by `node.type === 'ArrayExpression'`, so a keyless element nested directly inside another
// element is SILENT: `<div><App /></div>` reports nothing, measured. What the JSX anchor is for is
// the other two findings, duplicates and key-before-spread, which apply to an element's children as
// much as to an array. A port that reported `missingArrayKey` from both anchors would fire on every
// nested element in the tree and pass the entire imported corpus, which writes no such case.
//
// # The two arms disagree about the CASE of the attribute name, and it is not a typo
//
// The iterator arm asks `hasProp(attributes, 'key')`, whose default options are
// `{spreadStrict: true, ignoreCase: true}`, so it uppercases both sides before comparing. The
// collection arm filters `x.name && x.name.name === 'key'` inline, which is exact. Measured on the
// installed build, and both directions are real:
//
//	[1,2,3].map(x => <App KEY={x} />);   SILENT   the iterator arm accepts it
//	[<App KEY={1} />];                   REPORTS  the collection arm does not
//
// Nothing in the corpus writes a capitalized key. Unifying the two would be the obvious cleanup and
// would change a verdict in one direction or the other, so both are spelled out here with the
// matcher named at each site.
//
// A spread also answers differently at the two sites. `hasProp` with `spreadStrict: true` returns
// FALSE for a spread, so the iterator arm treats a spread as not supplying a key and reports:
// `[1,2,3].map(x => <App {...p} />)` reports. The collection arm's filter finds no matching name in
// a spread either, so it reports too. They agree here by two different routes.
//
// # Children.toArray, which is a scope rather than a check
//
// Upstream sets a flag when it enters a `React.Children.toArray(...)` call and clears it on exit,
// and every other listener returns early while it is set. That is a containment question, so it is
// answered here by walking up the parent chain from the node being judged rather than by keeping a
// flag across a traversal whose order this rule does not control.
//
// The two accepted shapes are `<pragma>.Children.toArray(...)` and a bare `Children.toArray(...)`.
// Measured: `Foo.Children.toArray(...)` does NOT suppress and a bare `toArray(...)` does not
// either, so the `Children` segment is required in both spellings. The pragma is read from the
// file's `@jsx` comment, the same mechanism `no-deprecated` in this package documents at length.
var JsxKey = rule.Rule{
	// No namespace prefix. The config writes `react/jsx-key` and the parity guard strips the
	// namespace on a `/` boundary, so `react-jsx-key` would match no inventory entry and lint no
	// files while passing every fixture in this package.
	Name: "react/jsx-key",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[JsxKeyOptions](options)
		pragma := reactPragmaFor(ctx)

		// Upstream keeps a WeakSet so a key attribute reported once as a duplicate is not reported
		// again. THERE IS NO EQUIVALENT HERE, because in this anchoring nothing can be reported
		// twice, and carrying a guard that cannot fire would read as a discrimination this rule
		// makes.
		//
		// The reason is the anchoring rather than an argument about the rule. Upstream's selector
		// is `ArrayExpression, JSXElement > JSXElement`, so its JSX arm fires once per CHILD and
		// hands the handler that child's parent's children, visiting one collection as many times
		// as it has elements. This rule anchors the JSX arm on the PARENT instead, so every
		// collection is visited exactly once.
		//
		// After that, the two anchors are disjoint by construction: an element is reached by the
		// array arm when its parent is an array literal and by the JSX arm when its parent is a
		// JSX element, and a node has one parent. Measured rather than argued, by walking the
		// parse of four shapes including an array literal nested inside JSX children, which is the
		// case that looks like it should reach both: the parent kinds come back
		// `ArrayLiteralExpression` and `JsxElement` and never both for one node, because the array
		// sits inside a `JsxExpression` wrapper rather than being a direct child.
		//
		// The verdict expires if a third anchor is added, or if the JSX arm is moved back onto the
		// child. Either change needs this paragraph re-argued rather than merely re-read.

		reportCollection := func(node *ast.Node, elements []*ast.Node, isArrayLiteral bool) {
			if isWithinChildrenToArray(node, pragma) {
				return
			}
			jsxElements := jsxElementsAmong(elements)
			if len(jsxElements) == 0 {
				return
			}

			// Keyed by the key attribute's SOURCE TEXT, which is what upstream compares:
			// `getText(context, attr.value)` is the value's own slice, so `key={x}` and `key={ x }`
			// are different keys and `key="a"` and `key={'a'}` are too. Reproduced rather than
			// evaluated, because evaluating would report on code whose real keys are unknown.
			byKeyText := map[string][]*ast.Node{}
			// Insertion order, because a map is unordered in Go and upstream emits duplicates in
			// the order the keys were first seen. Without this the finding ORDER varies run to run
			// while the count stays fixed, which is the shape a fixture asserting ids cannot see.
			keyOrder := []string{}

			for _, element := range jsxElements {
				attributes := attributesOfJsxElement(element)
				keyAttributes := exactlyNamedAttributes(attributes, "key")

				if len(keyAttributes) == 0 {
					// Guarded by the anchor: a keyless element nested in another element is
					// silent, and only an array literal reports here.
					if isArrayLiteral {
						ctx.ReportNode(element, messageJsxKeyMissingArrayKey)
					}
					continue
				}

				for _, keyAttribute := range keyAttributes {
					text := keyAttributeValueText(ctx, keyAttribute)
					if _, seen := byKeyText[text]; !seen {
						keyOrder = append(keyOrder, text)
					}
					byKeyText[text] = append(byKeyText[text], keyAttribute)

					// Reported once per KEY ATTRIBUTE rather than once per element, which is
					// upstream's loop shape, and it reports on the CONTAINER rather than on the
					// element. Both measured: the span for an array literal covers the whole
					// `[...]` and for a nested element covers the whole parent element.
					if settings.CheckKeyMustBeforeSpread && isKeyAfterSpread(attributes) {
						ctx.ReportNode(node, messageJsxKeyBeforeSpread)
					}
				}
			}

			if !settings.WarnOnDuplicates {
				return
			}
			for _, text := range keyOrder {
				group := byKeyText[text]
				if len(group) < 2 {
					continue
				}
				for _, keyAttribute := range group {
					// The KEY ATTRIBUTE, not the element. Measured: the span covers `key="a"`.
					ctx.ReportNode(keyAttribute, messageJsxKeyNonUniqueKeys)
				}
			}
		}

		// checkIteratorElement is upstream's function of the same name, and it is the ONLY place
		// the two iterator messages are produced.
		checkIteratorElement := func(node *ast.Node) {
			// Skipped once, here, so the element reaches every check below and its finding sits on
			// the element rather than on the paren, where ESLint's is. Skipping only in the kind test
			// let `map(x => (<A key={x} />))` pass as an element and then read the paren's attributes,
			// which are none: 103 keyed elements reported as missing a key across ahra and www
			// (#ge8899e).
			node = skipParenthesesSafely(node)
			if node == nil {
				return
			}
			if isJsxElementNode(node) {
				attributes := attributesOfJsxElement(node)
				// Case-INSENSITIVE here, unlike the collection arm. See the rule's doc comment.
				if !hasKeyAttributeIgnoringCase(attributes) {
					ctx.ReportNode(node, messageJsxKeyMissingIterKey)
					return
				}
				if settings.CheckKeyMustBeforeSpread && isKeyAfterSpread(attributes) {
					// The ELEMENT here, where the collection arm reports the container. Measured.
					ctx.ReportNode(node, messageJsxKeyBeforeSpread)
				}
				return
			}
			if settings.CheckFragmentShorthand && node.Kind == ast.KindJsxFragment {
				ctx.ReportNode(node, messageJsxKeyMissingIterKeyUsePragma)
			}
		}

		checkIteratorCallback := func(callback *ast.Node) {
			// Parentheses are skipped at every position this rule reads a node, and it is fidelity
			// rather than a widening. espree produces no parenthesized-expression node at all, so
			// `Array.from([1,2,3], (x => <App />))` reaches upstream's check as a bare arrow. Our
			// parser keeps the node, so without this the rule goes silent on two of upstream's own
			// reporting cases, `invalid-14` and `invalid-15`, which are the only two in the corpus
			// that write a paren. Measured on the installed build: every parenthesized position
			// this rule touches reports there, including doubled parens and a parenthesized callee.
			callback = skipParenthesesSafely(callback)
			if !isFunctionLikeExpressionForJsxKey(callback) {
				return
			}
			checkArrowFunctionWithJsx(callback, checkIteratorElement)
			checkFunctionsBlockStatement(callback, checkIteratorElement)
		}

		return rule.Listeners{
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				elements := node.AsArrayLiteralExpression().Elements
				if elements == nil {
					return
				}
				reportCollection(node, elements.Nodes, true)
			},

			// The JSX anchor of the collection arm. Upstream's selector is `JSXElement >
			// JSXElement`, a child element, and it hands the handler `node.parent.children`, so
			// the collection is the PARENT's children and the parent is what gets reported for a
			// key-before-spread. Anchoring on the parent directly reaches the same collection once
			// instead of once per child, which is what stops the duplicate findings from
			// multiplying by the number of children.
			ast.KindJsxElement: func(node *ast.Node) {
				children := node.AsJsxElement().Children
				if children == nil {
					return
				}
				reportCollection(node, children.Nodes, false)
			},

			// The shorthand fragment sitting DIRECTLY in an array literal. Upstream checks
			// `node.parent.type === 'ArrayExpression'`, so a fragment inside a JSX element's
			// children is silent even under the option. Measured: `<div><></></div>` reports
			// nothing with `checkFragmentShorthand` on.
			ast.KindJsxFragment: func(node *ast.Node) {
				if !settings.CheckFragmentShorthand || isWithinChildrenToArray(node, pragma) {
					return
				}
				if node.Parent == nil || node.Parent.Kind != ast.KindArrayLiteralExpression {
					return
				}
				ctx.ReportNode(node, messageJsxKeyMissingArrayKeyUsePragma)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if isWithinChildrenToArray(node, pragma) {
					return
				}
				call := node.AsCallExpression()
				callee := skipParenthesesSafely(call.Expression)
				if callee == nil {
					return
				}
				name, named := calleePropertyName(callee)
				if !named {
					return
				}
				arguments := []*ast.Node{}
				if call.Arguments != nil {
					arguments = call.Arguments.Nodes
				}

				switch name {
				case "map":
					// `node.arguments.length > 0 && node.arguments[0]`.
					if len(arguments) == 0 {
						return
					}
					checkIteratorCallback(arguments[0])
				case "from":
					// `node.arguments.length > 1 && node.arguments[1]`. Note upstream does NOT
					// check that the callee's object is `Array`, so `whatever.from(a, x => <App/>)`
					// is checked too. Reproduced.
					if len(arguments) < 2 {
						return
					}
					checkIteratorCallback(arguments[1])
				}
			},
		}
	},
}

// checkArrowFunctionWithJsx is upstream's function of the same name.
//
// It handles an arrow with an EXPRESSION body, and its three arms are the ones the installed build
// already had. Note upstream's own guard shape: the first `if` requires an arrow, and the following
// `if`/`else if` do NOT, so a function EXPRESSION with a conditional body would reach them. That is
// unreachable in JavaScript, since a function expression's body is always a block, and it is left
// structurally faithful here rather than collapsed, because collapsing it would be a claim about
// the parser rather than about the rule.
func checkArrowFunctionWithJsx(node *ast.Node, checkIteratorElement func(*ast.Node)) {
	if node == nil {
		return
	}
	body := functionLikeBodyOf(node)
	if body == nil {
		return
	}

	// Upstream's first arm is `isArrFn && isJSX(node.body)`, and the arrow test is REDUNDANT here.
	// Measured by walking the parse of four callback shapes: a function expression's body is always
	// a `KindBlock`, which none of the three arms below matches, so only an arrow can reach any of
	// them. A mutant dropping the arrow test survives the whole fixture set for that reason.
	//
	// Dropped rather than kept, because a guard that cannot fire reads as a discrimination this
	// rule makes. The verdict expires if `functionLikeBodyOf` learns a third node kind, or if the
	// callback filter widens past function expressions and arrows.
	//
	// A block body was previously declined here with an early return, which was also inert for the
	// same reason and additionally diverged from upstream, which calls this function on a
	// block-bodied arrow and simply matches nothing. Removed on the same measurement.
	if isJsxElementOrFragment(body) {
		checkIteratorElement(body)
	}

	if body.Kind == ast.KindConditionalExpression {
		conditional := body.AsConditionalExpression()
		if isJsxElementOrFragment(conditional.WhenTrue) {
			checkIteratorElement(conditional.WhenTrue)
		}
		if isJsxElementOrFragment(conditional.WhenFalse) {
			checkIteratorElement(conditional.WhenFalse)
		}
		return
	}
	if isLogicalExpression(body) {
		right := body.AsBinaryExpression().Right
		if isJsxElementOrFragment(right) {
			checkIteratorElement(right)
		}
	}
}

// checkFunctionsBlockStatement is upstream's function of the same name, from THE CLONE.
//
// The three arms below are the divergence documented on the rule: the installed 7.37.5 has only the
// `else` arm, so a returned conditional or logical is silent there and reports here.
func checkFunctionsBlockStatement(node *ast.Node, checkIteratorElement func(*ast.Node)) {
	body := functionLikeBodyOf(node)
	if body == nil || body.Kind != ast.KindBlock {
		return
	}

	for _, returnStatement := range jsxKeyReturnStatements(body) {
		argument := skipParenthesesSafely(returnStatement.AsReturnStatement().Expression)
		if argument == nil {
			continue
		}
		if argument.Kind == ast.KindConditionalExpression {
			conditional := argument.AsConditionalExpression()
			if isJsxElementOrFragment(conditional.WhenTrue) {
				checkIteratorElement(conditional.WhenTrue)
			}
			if isJsxElementOrFragment(conditional.WhenFalse) {
				checkIteratorElement(conditional.WhenFalse)
			}
			continue
		}
		if isLogicalExpression(argument) {
			right := argument.AsBinaryExpression().Right
			if isJsxElementOrFragment(right) {
				checkIteratorElement(right)
			}
			// Upstream's `else if` means a logical whose right side is NOT JSX falls through to
			// nothing rather than to the general arm, so `return x && foo()` is silent. That is a
			// clean case in the corpus.
			//
			// This `continue` is INERT and is kept as structure rather than as a discrimination.
			// Measured: removing it survives the whole fixture set, because the value that would
			// then reach `checkIteratorElement` is the whole logical expression, which is not a JSX
			// element, so both paths arrive at the same silence. Two controls establish the arm is
			// live and covered rather than dead: passing the LEFT operand instead of the right is
			// caught by nine lines, and removing the arm entirely is caught by seven.
			//
			// It stays because it is upstream's control flow, and because it stops being inert the
			// moment `checkIteratorElement` learns to look inside an expression, which is exactly
			// the kind of change that would otherwise arrive with nothing to announce it.
			continue
		}
		checkIteratorElement(argument)
	}
}

// jsxKeyReturnStatements is upstream's `getReturnStatements`, which is deliberately shallow.
//
// It descends into an `IfStatement`'s branches recursively and otherwise reads only the DIRECT
// statements of a block. So a return inside a loop, a `try`, or a `switch` is invisible, and so is
// one inside a nested function. Measured on the installed build, all silent:
//
//	map(x => { for (;;) { return <A />; } })       silent
//	map(x => { try { return <A />; } catch(e){} }) silent
//	map(x => { function f() { return <A />; } })   silent for the inner return
//
// The nested-function case is worth naming separately: upstream is silent because the inner
// function's body is not in the outer block's statement list, not because it filters functions out.
// A recursive walk that simply skipped functions would still report the loop and try cases, so
// reproducing this means copying the SHALLOWNESS rather than adding an exclusion.
func jsxKeyReturnStatements(node *ast.Node) []*ast.Node {
	collected := []*ast.Node{}

	// The `IfStatement` arm recurses into both branches, and a branch may be a block, in which case
	// the `Array.isArray(node.body)` arm below reads its direct statements. Written as an explicit
	// recursion over the two shapes upstream handles rather than as a general walk.
	var visit func(current *ast.Node, recursing bool)
	visit = func(current *ast.Node, recursing bool) {
		if current == nil {
			return
		}
		switch current.Kind {
		case ast.KindIfStatement:
			statement := current.AsIfStatement()
			visit(statement.ThenStatement, true)
			visit(statement.ElseStatement, true)
		case ast.KindReturnStatement:
			collected = append(collected, current)
		case ast.KindBlock:
			// Upstream's `Array.isArray(node.body)` arm, which takes only the DIRECT statements and
			// recurses only for a nested `IfStatement`.
			statements := current.AsBlock().Statements
			if statements == nil {
				return
			}
			for _, statement := range statements.Nodes {
				if statement.Kind == ast.KindIfStatement {
					visit(statement, true)
				}
				if statement.Kind == ast.KindReturnStatement {
					collected = append(collected, statement)
				}
			}
		}
		_ = recursing
	}

	visit(node, false)
	return collected
}

// isWithinChildrenToArray answers upstream's traversal flag by walking up instead.
//
// Upstream sets a boolean when its selector matches on enter and clears it on exit, which makes
// "am I inside one of these calls" a property of traversal order. Our walk is pre-order with no
// exit hook, so the same question is answered by looking at the ancestors of the node being judged,
// which is order-independent and reaches the same set of nodes.
//
// The two accepted spellings, and what is NOT accepted, both measured on the installed build:
//
//	React.Children.toArray(...)   suppresses (the pragma's Children)
//	Children.toArray(...)         suppresses (a bare Children)
//	Foo.Children.toArray(...)     does NOT suppress
//	toArray(...)                  does NOT suppress
func isWithinChildrenToArray(node *ast.Node, pragma string) bool {
	for current := node; current != nil; current = current.Parent {
		if current.Kind != ast.KindCallExpression {
			continue
		}
		if isChildrenToArrayCall(current, pragma) {
			return true
		}
	}
	return false
}

// isChildrenToArrayCall matches upstream's two selector alternatives.
//
// The first requires `callee.object.object.name === pragma` and `callee.object.property.name ===
// 'Children'`; the second requires `callee.object.name === 'Children'`. Both require
// `callee.property.name === 'toArray'`.
func isChildrenToArrayCall(node *ast.Node, pragma string) bool {
	callee := node.AsCallExpression().Expression
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "toArray" {
		return false
	}

	object := access.Expression
	if object == nil {
		return false
	}
	// `Children.toArray(...)`
	if object.Kind == ast.KindIdentifier {
		return object.Text() == "Children"
	}
	// `<pragma>.Children.toArray(...)`
	if object.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	inner := object.AsPropertyAccessExpression()
	innerName := inner.Name()
	if innerName == nil || innerName.Kind != ast.KindIdentifier || innerName.Text() != "Children" {
		return false
	}
	return inner.Expression != nil && inner.Expression.Kind == ast.KindIdentifier &&
		inner.Expression.Text() == pragma
}

// calleePropertyName reads the property a call is made through, which is what upstream's selectors
// match on.
//
// The selectors require a `MemberExpression` or `OptionalMemberExpression` callee, so a bare
// `map(...)` never matches. Our tree spells the optional form with a question-dot token on the same
// property access node, so both reach here through one kind. Measured:
// `[1,2,3]?.map(x => <App/>)` reports.
func calleePropertyName(callee *ast.Node) (string, bool) {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return "", false
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return "", false
	}
	return name.Text(), true
}

// isFunctionLikeExpressionForJsxKey reproduces `astUtil.isFunctionLikeExpression`.
//
// A function EXPRESSION or an arrow, and nothing else. A function DECLARATION does not qualify,
// which cannot be an argument anyway, and neither does an identifier: `[1,2,3].map(someFn)` is
// clean, measured, and it is a corpus case.
func isFunctionLikeExpressionForJsxKey(node *ast.Node) bool {
	return node != nil &&
		(node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction)
}

// functionLikeBodyOf returns the body of an arrow or a function expression.
func functionLikeBodyOf(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	}
	return nil
}

// skipParenthesesSafely is `ast.SkipParentheses` with a nil check in front of it.
//
// The upstream helper dereferences immediately: `ast.IsOuterExpression` reads `node.Kind` with no
// guard, so handing it a nil panics and takes the whole run down. That is reachable from this rule
// on ordinary source, because `return;` with no argument gives a nil expression and it is one of
// upstream's own clean cases (`[1,2,3].map(function(x) { return; })`). Found by a fixture rather
// than by reading, which is the whole reason the clean cases are imported.
func skipParenthesesSafely(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipParentheses(node)
}

// isJsxElementOrFragment reproduces `util/jsx.js`'s `isJSX`.
//
// An element or a fragment, by kind. Upstream also has a self-closing element, which its parser
// spells as a `JSXElement`; ours is a distinct kind, so both are named here. Omitting the
// self-closing kind would silence every arrow returning `<App />`, which is most of the corpus.
func isJsxElementOrFragment(node *ast.Node) bool {
	node = skipParenthesesSafely(node)
	return node != nil && (node.Kind == ast.KindJsxElement ||
		node.Kind == ast.KindJsxSelfClosingElement ||
		node.Kind == ast.KindJsxFragment)
}

// isJsxElementNode is `node.type === 'JSXElement'`, which excludes a fragment.
//
// `checkIteratorElement` branches on element versus fragment and does different things for each, so
// the two predicates are separate rather than one with a flag.
func isJsxElementNode(node *ast.Node) bool {
	node = skipParenthesesSafely(node)
	return node != nil &&
		(node.Kind == ast.KindJsxElement || node.Kind == ast.KindJsxSelfClosingElement)
}

// isLogicalExpression matches `&&`, `||` and `??`, which is what espree calls a LogicalExpression.
//
// Our parser gives all three the same `BinaryExpression` kind as `+`, so the operator token is what
// separates them and a port testing the kind alone would descend into the right-hand side of every
// arithmetic expression.
func isLogicalExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := node.AsBinaryExpression().OperatorToken
	if operator == nil {
		return false
	}
	switch operator.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

// attributesOfJsxElement returns the attribute list of either opening form.
//
// A `JsxElement` carries its attributes on its OpeningElement, while a `JsxSelfClosingElement`
// carries them directly, so the two shapes are unified here before any attribute question is asked.
func attributesOfJsxElement(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	opening := node
	if node.Kind == ast.KindJsxElement {
		opening = node.AsJsxElement().OpeningElement
	}
	if opening == nil {
		return nil
	}
	_, attributes := jsx.ElementParts(opening)
	return attributesOf(attributes)
}

// exactlyNamedAttributes is the COLLECTION arm's matcher, and it is case-sensitive.
//
// `attrs.filter((x) => x.name && x.name.name === 'key')`, written inline upstream rather than
// through `hasProp`. Measured: `[<App KEY={1} />]` reports, where the iterator arm's equivalent
// accepts the same spelling. See the rule's doc comment.
//
// It returns every match rather than the first, because upstream's loop reports the
// key-before-spread finding once per matching attribute.
func exactlyNamedAttributes(properties []*ast.Node, wanted string) []*ast.Node {
	matches := []*ast.Node{}
	for _, property := range properties {
		name, named := jsx.AttributeName(property)
		if named && jsx.MatchExactly(name, wanted) {
			matches = append(matches, property)
		}
	}
	return matches
}

// hasKeyAttributeIgnoringCase is the ITERATOR arm's matcher, and it is case-insensitive.
//
// `hasProp(attributes, 'key')` with its default `{spreadStrict: true, ignoreCase: true}`. The
// spread strictness is why a spread does NOT satisfy the key here: `hasProp` returns
// `!options.spreadStrict` for a spread, which is false. `jsx.HasAttributeNamed` declines a spread
// for the same reason, by requiring a named attribute, so the two agree.
func hasKeyAttributeIgnoringCase(properties []*ast.Node) bool {
	for _, property := range properties {
		name, named := jsx.AttributeName(property)
		if named && jsx.MatchIgnoringCase(name, "key") {
			return true
		}
	}
	return false
}

// isKeyAfterSpread reproduces upstream's function of the same name.
//
// It scans left to right, remembers whether a spread has been seen, and answers true at the first
// `key` that follows one. A non-attribute that is not a spread is skipped without clearing the
// flag, which matters only for shapes our parser does not produce.
//
// The name comparison here is upstream's `propName(attribute) === 'key'`, which is EXACT and
// therefore agrees with the collection arm rather than with the iterator arm's `hasProp`. So an
// element written `<App {...p} KEY={x} />` is not a key-after-spread even though the iterator arm
// considers it keyed. Both facts come from the same asymmetry and neither is in the corpus.
func isKeyAfterSpread(properties []*ast.Node) bool {
	foundSpread := false
	for _, property := range properties {
		if property.Kind == ast.KindJsxSpreadAttribute {
			foundSpread = true
			continue
		}
		if property.Kind != ast.KindJsxAttribute {
			continue
		}
		name, named := jsx.AttributeName(property)
		if foundSpread && named && jsx.MatchExactly(name, "key") {
			return true
		}
	}
	return false
}

// jsxElementsAmong filters a collection to the JSX ELEMENTS in it, excluding fragments.
//
// Upstream's filter is `x && x.type === 'JSXElement'`, so a fragment sitting in an array is not a
// member of the collection for the duplicate and missing-key questions; the separate `JSXFragment`
// listener is what handles it. The `x &&` guard is for an array HOLE, which our parser spells as an
// omitted-expression node rather than as a null, so it is declined by kind here.
func jsxElementsAmong(nodes []*ast.Node) []*ast.Node {
	elements := []*ast.Node{}
	for _, node := range nodes {
		if isJsxElementNode(node) {
			elements = append(elements, node)
		}
	}
	return elements
}

// keyAttributeValueText returns the key's value as SOURCE TEXT, which is what upstream compares.
//
// `getText(context, attr.value)` is the value node's own slice, so two keys are the same key only
// when they are written the same way. `key="a"` and `key={'a'}` are different, and so are `key={x}`
// and `key={ x }`. Reproduced rather than evaluated: evaluating would mean deciding what an
// arbitrary expression is worth, and a rule that guessed would report on code whose real keys are
// unknown.
//
// A `key` with no value at all yields the empty string, which is what `getText(context, undefined)`
// returns upstream, and two valueless keys therefore collide with each other.
func keyAttributeValueText(ctx rule.Context, attribute *ast.Node) string {
	initializer := attribute.AsJsxAttribute().Initializer
	if initializer == nil {
		return ""
	}
	return sourceSliceOf(ctx, initializer)
}
