package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
)

// The shape predicates for consistency-require-constant-casing.
//
// Each one answers a single question about what a constant holds, and every one of them exists
// because it was missing once and the rule contradicted itself on a real line. The comments say
// which line, because that is the part that ports.

// dataStructureConstructors are built-in constructors whose instances hold data rather than act.
//
// `new Map()` and `new Set()` produce a data structure that happens to be read through methods, so
// an exported one is exported data and keeps PascalCase. Everything else built with `new` is a class
// someone here wrote, which is the case that actually acts.
var dataStructureConstructors = map[string]bool{
	"Map": true, "Set": true, "WeakMap": true, "WeakSet": true, "Date": true, "RegExp": true,
	"Array": true, "Int8Array": true, "Uint8Array": true, "Uint8ClampedArray": true,
	"Int16Array": true, "Uint16Array": true, "Int32Array": true, "Uint32Array": true,
	"Float32Array": true, "Float64Array": true, "BigInt64Array": true, "BigUint64Array": true,
	"ArrayBuffer": true, "SharedArrayBuffer": true, "DataView": true, "URL": true,
	"URLSearchParams": true, "Headers": true, "TextEncoder": true, "TextDecoder": true,
	// A boxed string holds one value and acts on nothing; it is here for the classes that extend it
	// (#w26f1b0).
	"String": true,
}

// unwrapAssertions peels the type-level wrappers that do not change what a value is.
//
// `as const`, `satisfies`, a non-null assertion and a parenthesis all wrap an expression without
// altering it, so a predicate reading the outermost node would miss what is underneath. `export
// const gql = gqlImplementation as unknown as UnifiedGqlType` is a function reached through a double
// cast, and the identifier underneath is what says so.
func unwrapAssertions(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		default:
			return node
		}
	}
	return nil
}

// calleeName reads the name a call is made through, whether it is bare or reached through an object.
// `createContext(...)` and `React.createContext(...)` both answer "createContext".
func calleeName(callee *ast.Node) string {
	callee = unwrapAssertions(callee)
	if callee == nil {
		return ""
	}
	if callee.Kind == ast.KindIdentifier {
		return callee.Text()
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		if name := callee.AsPropertyAccessExpression().Name(); name != nil {
			return name.Text()
		}
	}
	return ""
}

// isFunctionNode reports the two shapes a function literal takes.
func isFunctionNode(node *ast.Node) bool {
	node = unwrapAssertions(node)
	return node != nil && (node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression)
}

// returnsAFunction reports whether a function literal hands back a function.
//
// Both bodies are read: a concise arrow body is the returned expression itself, and a block body
// needs its return statements found.
func returnsAFunction(node *ast.Node) bool {
	body := react.FunctionBody(node)
	if body == nil {
		return false
	}
	// Concise arrow body: `() => function Foo() {}`.
	if body.Kind != ast.KindBlock {
		return isFunctionNode(body)
	}
	return blockReturnsMatching(body, isFunctionNode)
}

// blockReturnsMatching reports whether any return statement directly in a block returns something
// the predicate accepts. Only the block's own statements are read, since a return nested inside
// another function belongs to that function rather than this one.
func blockReturnsMatching(body *ast.Node, matches func(*ast.Node) bool) bool {
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}
	for _, statement := range body.AsBlock().Statements.Nodes {
		if statement.Kind != ast.KindReturnStatement {
			continue
		}
		if matches(statement.AsReturnStatement().Expression) {
			return true
		}
	}
	return false
}

// isFunctionValued reports whether the initializer holds a function rather than a value.
//
// The whole convention rests on casing signalling reach, which only reads cleanly on values. A
// function is named for what it does, and it is called: `formatNumber()` says it formats a number,
// `FormatNumber()` reads like a constructor. React hooks are stricter still, since the `use` prefix
// is how React's own linting recognizes a hook and capitalizing one breaks that detection.
//
// This reads the initializer's syntax rather than guessing from the name, so a value that merely
// starts with `get` is still judged on its casing. A function declared with the `function` keyword
// never reaches this rule at all, since it is not a variable declaration.
func isFunctionValued(initializer *ast.Node) bool {
	expression := unwrapAssertions(initializer)
	if expression == nil {
		return false
	}
	if isFunctionNode(expression) {
		return true
	}
	if expression.Kind != ast.KindCallExpression {
		return false
	}

	// A memoized function: the value is whatever the callback hands back, so look through the memo.
	call := expression.AsCallExpression()
	name := calleeName(call.Expression)
	if name != "useMemo" && name != "useCallback" && name != "memo" {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	firstArgument := call.Arguments.Nodes[0]

	// useCallback and memo take the function directly.
	if name != "useMemo" {
		return isFunctionNode(firstArgument)
	}
	// useMemo takes a factory, so the function is what that factory returns.
	return returnsAFunction(firstArgument)
}

// isReactComponentFactory reports whether the initializer produces a React component.
//
// A component is a function, so the function clause would demand camelCase and be wrong: JSX needs
// the capital, and memo is how a component is wrapped for identity stability. The tell is the
// wrapped function's own name, since `memo(function MenuSearch(...))` names its component in
// PascalCase, which is the convention asserting what it is.
func isReactComponentFactory(initializer *ast.Node) bool {
	expression := unwrapAssertions(initializer)
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return false
	}

	call := expression.AsCallExpression()
	name := calleeName(call.Expression)
	if name != "memo" && name != "forwardRef" && name != "useMemo" {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	firstArgument := call.Arguments.Nodes[0]

	// `memo(function MenuSearch(...))` — the wrapped function names the component.
	if firstArgument.Kind == ast.KindFunctionExpression {
		if named := firstArgument.AsFunctionExpression().Name(); named != nil && startsUppercase(named.Text()) {
			return true
		}
	}
	// `useMemo(function () { return function FormNotice(...) {} }, [])` — read one level deeper.
	if name == "useMemo" && firstArgument.Kind == ast.KindFunctionExpression {
		return blockReturnsMatching(firstArgument.AsFunctionExpression().Body, func(returned *ast.Node) bool {
			if returned == nil || returned.Kind != ast.KindFunctionExpression {
				return false
			}
			named := returned.AsFunctionExpression().Name()
			return named != nil && startsUppercase(named.Text())
		})
	}
	return false
}

// componentTypeNames are the annotations that declare a value to be a React component.
var componentTypeNames = map[string]bool{
	"FC": true, "FunctionComponent": true, "ComponentType": true, "VoidFunctionComponent": true,
}

// isReactComponentAnnotation reports whether the declaration is annotated as a React component.
//
// JSX decides between a component and an intrinsic element by the tag's first letter, so a lowercase
// name stops being a reference and becomes a literal `<shareblock>` tag. The factory check catches
// memo and forwardRef, but a component can just as well say what it is in its type: `export const
// ShareBlock: React.FC<Properties> = ...`. Recasing that one produced `<shareBlock />`, which
// TypeScript then reported as an unknown HTML element.
//
// Matched on the annotation's own spelling rather than through the checker, since every one of these
// names is written by hand.
func isReactComponentAnnotation(declaration *ast.VariableDeclaration) bool {
	typeName := annotationTypeName(declaration)
	return typeName != "" && componentTypeNames[typeName]
}

// annotationTypeName reads the rightmost name of a bare type reference annotation, or "" when the
// declaration has no such annotation. `: Client` answers "Client" and `: React.FC` answers "FC".
func annotationTypeName(declaration *ast.VariableDeclaration) string {
	if declaration == nil || declaration.Type == nil || declaration.Type.Kind != ast.KindTypeReference {
		return ""
	}
	reference := declaration.Type.AsTypeReferenceNode()
	if reference == nil || reference.TypeName == nil {
		return ""
	}
	switch reference.TypeName.Kind {
	case ast.KindIdentifier:
		return reference.TypeName.Text()
	case ast.KindQualifiedName:
		if right := reference.TypeName.AsQualifiedName().Right; right != nil {
			return right.Text()
		}
	}
	return ""
}

// isReactContextBinding reports whether the initializer creates a React context.
//
// A context object is component-adjacent: it is read as `ThemeContext.Provider` in element position,
// and React's own convention capitalizes it. This checks the call shape rather than a `*Context`
// name suffix, so a value merely named that way is still judged on its casing.
func isReactContextBinding(initializer *ast.Node) bool {
	expression := unwrapAssertions(initializer)
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return false
	}
	return calleeName(expression.AsCallExpression().Expression) == "createContext"
}

// isDynamicImportBinding reports the `const NodeFileSystem = await import('node:fs')` shape.
//
// That is an import wearing a const, and import-require-node-namespace mandates exactly this
// PascalCase namespace spelling. Without this check the two rules contradict each other on the same
// line, which is a defect rather than a preference.
//
// The wrappers are peeled because a default export is reached as `(await import('better-sqlite3'))
// .default`, which puts a member access outside the await. Missing that told `Database` to be
// lowercase while the next line reads `new Database(...)`.
func isDynamicImportBinding(initializer *ast.Node) bool {
	expression := initializer
	for expression != nil {
		switch expression.Kind {
		case ast.KindPropertyAccessExpression:
			expression = expression.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			expression = expression.AsElementAccessExpression().Expression
		case ast.KindAwaitExpression:
			expression = expression.AsAwaitExpression().Expression
		case ast.KindNonNullExpression:
			expression = expression.AsNonNullExpression().Expression
		case ast.KindParenthesizedExpression:
			expression = expression.AsParenthesizedExpression().Expression
		case ast.KindCallExpression:
			// `import('x')` parses as a call whose callee is the import keyword.
			call := expression.AsCallExpression()
			return call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword
		default:
			return false
		}
	}
	return false
}

// isPascalCaseReBinding reports a constant that re-binds a PascalCase constructor or namespace from
// somewhere else, as in `const NativeMap = globalThis.Map` or `const Buffer = NodeBuffer.Buffer`.
//
// The name is mirroring an upstream spelling rather than choosing one, and the value is a
// constructor, so it is `new NativeMap()` at the call site where the capital is conventional. This
// only fires when the source property is itself PascalCase, so `const total = counts.Sum` is
// unaffected.
func isPascalCaseReBinding(initializer *ast.Node) bool {
	return anyBranchReadsPascalCaseProperty(unwrapAssertions(initializer))
}

// anyBranchReadsPascalCaseProperty reports whether any branch of the expression reads a PascalCase
// property.
//
// A constructor is rarely reached by a single clean member access. It is picked out of a fallback
// chain, `polyfill ?? (typeof window !== 'undefined' ? window.ResizeObserver : undefined)`, where the
// constructor sits two levels down inside a conditional inside a nullish coalesce. Reading only the
// outermost node misses it entirely, and the constant then gets told to be camelCase while the next
// line does `new ResizeObserverConstructor(handler)`.
//
// One arm is enough, since a fallback chain is picking between spellings of the same thing.
func anyBranchReadsPascalCaseProperty(node *ast.Node) bool {
	node = unwrapAssertions(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindBinaryExpression:
		// `a ?? b` and `a || b` — either side may hold the constructor.
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		if binary.OperatorToken.Kind != ast.KindQuestionQuestionToken &&
			binary.OperatorToken.Kind != ast.KindBarBarToken {
			return false
		}
		return anyBranchReadsPascalCaseProperty(binary.Left) || anyBranchReadsPascalCaseProperty(binary.Right)

	case ast.KindConditionalExpression:
		// The condition is a guard, so only the arms matter.
		conditional := node.AsConditionalExpression()
		return anyBranchReadsPascalCaseProperty(conditional.WhenTrue) ||
			anyBranchReadsPascalCaseProperty(conditional.WhenFalse)

	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		return name != nil && startsUppercase(name.Text())
	}
	return false
}

// isCasingConstEnumShape reports the `{ Member: 'Member' } as const` shape.
//
// consistency-require-type-suffix requires these to end in `Kind`, which forces PascalCase, so
// flagging a local one here would put two rules in direct contradiction. The value is also a type as
// much as a value, and we spell string-literal unions in PascalCase, so the capital carries meaning.
//
// Deliberately not the isConstEnumShape that rule already owns, and the difference is not an
// oversight in either. That one matches string members whose value equals the key, because it is
// deciding whether a name needs the Kind suffix. This one additionally accepts numeric members and
// requires PascalCase keys, because it is deciding whether a capital is load bearing, and
// `{ Positive: 1 } as const` is exactly as much a discriminated value set as the string form.
// Widening the shared one to cover both would quietly change which names require the Kind suffix, so
// the two questions keep two answers.
func isCasingConstEnumShape(initializer *ast.Node) bool {
	if initializer == nil || initializer.Kind != ast.KindAsExpression {
		return false
	}

	assertion := initializer.AsAsExpression()
	if assertion.Type == nil || assertion.Type.Kind != ast.KindTypeReference {
		return false
	}
	reference := assertion.Type.AsTypeReferenceNode()
	if reference.TypeName == nil || reference.TypeName.Kind != ast.KindIdentifier {
		return false
	}
	if reference.TypeName.Text() != "const" {
		return false
	}
	if assertion.Expression == nil || assertion.Expression.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	object := assertion.Expression.AsObjectLiteralExpression()
	if object.Properties == nil || len(object.Properties.Nodes) == 0 {
		return false
	}

	for _, property := range object.Properties.Nodes {
		if !isCasingConstEnumMember(property) {
			return false
		}
	}
	return true
}

// isConstEnumMember reports whether one property has the shape an enum member takes.
//
// PascalCase member names are what make this a discriminated value set rather than a lookup table. A
// table's keys are labels (`small: '...'`), an enum's are members (`Positive: 1`). The value is a
// literal it is keyed by, either the matching string or a number.
func isCasingConstEnumMember(property *ast.Node) bool {
	if property == nil || property.Kind != ast.KindPropertyAssignment {
		return false
	}

	assignment := property.AsPropertyAssignment()
	key := assignment.Name()
	if key == nil || key.Kind != ast.KindIdentifier || !startsUppercase(key.Text()) {
		return false
	}

	value := assignment.Initializer
	if value == nil {
		return false
	}
	// A unary minus wraps a negative number, which the parser reports as an expression.
	if value.Kind == ast.KindPrefixUnaryExpression {
		unary := value.AsPrefixUnaryExpression()
		return unary.Operator == ast.KindMinusToken &&
			unary.Operand != nil && unary.Operand.Kind == ast.KindNumericLiteral
	}
	if value.Kind == ast.KindNumericLiteral {
		return true
	}
	if value.Kind != ast.KindStringLiteral {
		return false
	}
	return key.Text() == value.Text()
}

// startsUppercase reports whether a name begins with a capital letter, ignoring leading underscores.
func startsUppercase(name string) bool {
	body := strings.TrimLeft(name, "_")
	if body == "" {
		return false
	}
	first := []rune(body)[0]
	return first >= 'A' && first <= 'Z'
}
