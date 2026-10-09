package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// arrayCallbackTargetMethods is the set of array methods whose callback must return a value, plus
// `forEach`, whose callback must not.
//
// Upstream spells this as one regular expression,
// `/^(?:every|filter|find(?:Last)?(?:Index)?|flatMap|forEach|map|reduce(?:Right)?|some|sort|toSorted)$/u`,
// which expands to exactly these thirteen. Written out rather than matched, because the expansion is
// the thing a reader needs to check and a regular expression hides it: `findLastIndex` and
// `reduceRight` are easy to lose in the nesting, and neither has a test that would notice.
var arrayCallbackTargetMethods = map[string]bool{
	"every":         true,
	"filter":        true,
	"find":          true,
	"findIndex":     true,
	"findLast":      true,
	"findLastIndex": true,
	"flatMap":       true,
	"forEach":       true,
	"map":           true,
	"reduce":        true,
	"reduceRight":   true,
	"some":          true,
	"sort":          true,
	"toSorted":      true,
}

// arrayMethodNameFor reports which array method a function is the callback of, if any.
//
// This is upstream's `getArrayMethodName`, and it is the half of the rule that decides WHETHER to
// judge at all. It walks UP from the function, because the callback can be nested inside expressions
// that choose between callbacks without changing which method receives one:
//
//	foo.every(cb || function(){})         a logical expression picks one
//	foo.every(a ? f : function(){})       a conditional picks one
//	foo?.every(function(){})              an optional chain
//	foo.every(function(){ return function(){}; }())   an immediately invoked wrapper RETURNS one
//
// All four are in the corpus and all four are reproduced. The fourth is the one worth reading twice:
// a `return` is followed out of the wrapper only when that wrapper is itself being called, which is
// what makes the returned function the callback rather than an ordinary return value.
//
// # The async and generator asymmetry is deliberate upstream
//
// A generator is never a checked callback, whatever the method. An async function is checked only
// for `Array.fromAsync`, and upstream expresses that by putting the `!node.async` guard around the
// `Array.from` and target-method arms while leaving the `fromAsync` arm outside it. Measured against
// the installed rule: `foo.map(async function(){})` is clean, `foo.every(function*(){})` is clean,
// and `Array.fromAsync(x, async function(){})` reports.
func arrayMethodNameFor(node *ast.Node) (string, bool) {
	if isGeneratorFunctionLike(node) {
		return "", false
	}
	isAsync := node.ModifierFlags()&ast.ModifierFlagsAsync != 0

	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return "", false
		}

		switch parent.Kind {
		case ast.KindBinaryExpression:
			// Upstream matches `LogicalExpression`, which is only the three short-circuit
			// operators. A binary expression here that is not one of them is not a choice between
			// callbacks, so it is not walked through.
			operator := parent.AsBinaryExpression().OperatorToken
			if operator == nil {
				return "", false
			}
			switch operator.Kind {
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				current = parent
				continue
			}
			return "", false

		case ast.KindConditionalExpression, ast.KindParenthesizedExpression:
			current = parent
			continue

		case ast.KindReturnStatement:
			// The immediately invoked wrapper case. The enclosing function is followed only when it
			// is itself the callee of a call, because that is what makes what it returns the value
			// the array method receives.
			enclosing := enclosingFunctionOf(parent)
			if enclosing == nil {
				return "", false
			}
			// Upstream writes `currentNode = func.parent`, which in ESLint's tree IS the call,
			// because that tree has no parenthesized-expression node at all. Ours does, so the
			// wrapper's parent is the paren and the call is one step further up. Climbing to the
			// call itself is the same position in both trees.
			//
			// Measured on `foo.forEach((function () { return (bar) => bar; })())`, which is in the
			// corpus: the wrapper's parent chain is paren, call, call, and taking `.Parent` landed on
			// the paren, so the walk saw a parenthesized expression rather than the call and gave up.
			call := callConsuming(enclosing)
			if call == nil {
				return "", false
			}
			current = call
			continue

		case ast.KindCallExpression:
			return arrayMethodNameForCall(parent, current, isAsync)

		default:
			return "", false
		}
	}
	return "", false
}

// arrayMethodNameForCall answers whether one call expression passes `argument` as a checked callback.
//
// The three arms are upstream's, in upstream's order, and the order matters because `Array.from` and
// a target method can both match a callee spelled `Array.from` only under different argument
// positions.
func arrayMethodNameForCall(call *ast.Node, argument *ast.Node, isAsync bool) (string, bool) {
	callExpression := call.AsCallExpression()
	if callExpression == nil {
		return "", false
	}
	// The callee is handed over as written. Both consumers below unwrap parentheses and chains
	// themselves, so unwrapping here as well is redundant: a mutant removing it survived the whole
	// corpus, and the two call sites at `targetMethodNameOf` and `isArrayStaticMethod` are the
	// enumeration that verdict rests on. Adding a third consumer that does not unwrap voids it.
	callee := callExpression.Expression
	arguments := callExpression.Arguments

	if !isAsync {
		// `Array.from(iterable, mapFn)` takes its callback SECOND, which is the whole reason this
		// arm is separate from the target-method one below.
		if isArrayStaticMethod(callee, "from") &&
			arguments != nil && len(arguments.Nodes) >= 2 && arguments.Nodes[1] == argument {
			return "from", true
		}
		// Every prototype method in the set takes its callback FIRST.
		if name, isTarget := targetMethodNameOf(callee); isTarget &&
			arguments != nil && len(arguments.Nodes) >= 1 && arguments.Nodes[0] == argument {
			return name, true
		}
	}

	// Outside the async guard on purpose: `Array.fromAsync` is the one method whose callback may be
	// async, and upstream places this arm after the return above so a non-async callback to it is
	// still reached.
	if isArrayStaticMethod(callee, "fromAsync") &&
		arguments != nil && len(arguments.Nodes) >= 2 && arguments.Nodes[1] == argument {
		return "fromAsync", true
	}

	return "", false
}

// targetMethodNameOf returns the method name a callee accesses, when it is one this rule checks.
//
// Upstream asks `astUtils.isSpecificMemberAccess(node, null, TARGET_METHODS)` with a null object
// pattern, so the RECEIVER is not constrained at all: `foo.every`, `foo.bar().every` and
// `[1,2].every` all match, and so does anything else with a `.every`. That is deliberate upstream,
// since the receiver's type is not knowable from syntax, and it is reproduced rather than narrowed.
//
// Both member forms are accepted because upstream's helper reads a static property name, which
// covers `foo['every']` as well as `foo.every`.
func targetMethodNameOf(callee *ast.Node) (string, bool) {
	callee = skipParenthesesAndChain(callee)
	if callee == nil {
		return "", false
	}
	name, hasName := staticMemberNameOf(callee)
	if !hasName || !arrayCallbackTargetMethods[name] {
		return "", false
	}
	return name, true
}

// isArrayStaticMethod reports whether a callee is `Array.<name>`.
//
// Upstream's `isArrayFromMethod` walks a chain of member accesses so that `[].constructor.from` and
// `Array.from` both match; the object it requires at the head is spelled `Array`. Reproduced at the
// spelling upstream keys on, which is the identifier `Array`.
func isArrayStaticMethod(callee *ast.Node, methodName string) bool {
	callee = skipParenthesesAndChain(callee)
	if callee == nil {
		return false
	}
	name, hasName := staticMemberNameOf(callee)
	if !hasName || name != methodName {
		return false
	}
	object := skipParenthesesAndChain(memberObjectOf(callee))
	if object == nil || object.Kind != ast.KindIdentifier {
		return false
	}
	if methodName == "fromAsync" {
		// `isArrayFromAsyncMethod` keys on the literal identifier `Array` and nothing else.
		return object.Text() == "Array"
	}
	// `isArrayFromMethod` uses `arrayOrTypedArrayPattern`, which is `/Array$/u`: any identifier
	// ENDING in `Array`, so every typed-array constructor matches. `Int32Array.from(x, fn)` is in
	// the corpus and is what makes this a suffix test rather than an equality one.
	return strings.HasSuffix(object.Text(), "Array")
}

// skipParenthesesAndChain unwraps parentheses and non-null assertions around an expression.
//
// Upstream's `skipChainExpression` does the chain half; the parenthesis half is ours, because our
// parser keeps a `KindParenthesizedExpression` node where ESLint's tree does not have one at all.
// `(foo?.filter)(cb)` and `(Array?.from)([], cb)` are both in the corpus and both reach a callee
// wrapped this way.
func skipParenthesesAndChain(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		default:
			return node
		}
	}
	return nil
}

// staticMemberNameOf returns the property name a member access reads, when that name is static.
//
// A computed access whose key is not a string literal has no static name, which is upstream's
// `getStaticPropertyName` returning null.
func staticMemberNameOf(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name == nil {
			return "", false
		}
		return name.Text(), true

	case ast.KindElementAccessExpression:
		argument := skipParenthesesAndChain(node.AsElementAccessExpression().ArgumentExpression)
		if argument == nil {
			return "", false
		}
		switch argument.Kind {
		case ast.KindStringLiteral:
			return argument.Text(), true
		case ast.KindNoSubstitutionTemplateLiteral:
			// Upstream reaches this through `getStaticStringValue`, which reads a template literal
			// with no substitutions as its cooked text. ``foo[`every`]`` is in the corpus and is the
			// only reason this arm exists.
			return argument.Text(), true
		}
		return "", false
	}
	return "", false
}

// memberObjectOf returns the receiver of a member access.
func memberObjectOf(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// isGeneratorFunctionLike reports whether a function carries an asterisk.
//
// An arrow function cannot be a generator, so only the two kinds this rule listens for that can be
// one are asked. Upstream's guard is `node.generator` and it is the very first thing the rule
// checks, before any method matching at all.
func isGeneratorFunctionLike(node *ast.Node) bool {
	if node.Kind == ast.KindFunctionExpression {
		return node.AsFunctionExpression().AsteriskToken != nil
	}
	return false
}

// enclosingFunctionOf returns the function a node runs inside, or nil at the top level.
//
// Upstream's `getUpperFunction`. It stops at the first function-like ancestor, which is what makes a
// `return` attribute to its own function rather than to an outer one.
func enclosingFunctionOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			return current
		case ast.KindSourceFile:
			return nil
		}
	}
	return nil
}

// isCalleeOfCall reports whether a node is the thing being called.
//
// Upstream's `isCallee`. It is what separates an immediately invoked wrapper, whose return value
// becomes the callback, from an ordinary function whose return value goes to its own caller.
func isCalleeOfCall(node *ast.Node) bool {
	return callConsuming(node) != nil
}

// callConsuming returns the call expression that invokes a node, or nil when nothing does.
//
// Parentheses around the callee are skipped on the way up, since `(function(){})()` and
// `function(){}()` are the same call and only the first has a paren node in our tree.
func callConsuming(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node = parent
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return nil
	}
	if parent.AsCallExpression().Expression != node {
		return nil
	}
	return parent
}
