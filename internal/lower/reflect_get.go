package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// reflectGet lowers Reflect.get(Object(v), key, v), a read of an unknown by a key known only at run
// time, as 94's propertyOf writes it (Adamic's ruling on #axochk; the shape accepted on #zhgxzjg). It
// matches that one shape: the same identifier v in both places, held as a union and proven by the
// checker neither null nor undefined there (propertyOf's guard narrows it to {}), and a key that is
// a string. So the node never sees a nullish v, Object(v) is v, and the receiver is moot: what's read
// is ir.DynamicKeyedProperty{v, key}. Every other Reflect.get is NotYet here, as is Reflect or
// Object(x) as a value anywhere else (namespaceExpression, libraryGlobalValue), exactly as before.
func (lowering *lowering) reflectGet(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression ||
		callee.Name().Text() != "get" ||
		!lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Reflect") {
		return nil, false, nil
	}
	const shape = "Reflect.get other than Reflect.get(Object(value), key, value) with one identifier as value"
	if call.QuestionDotToken != nil ||
		callee.AsPropertyAccessExpression().QuestionDotToken != nil ||
		len(call.Arguments.Nodes) != 3 ||
		hasSpread(node) {
		return nil, true, lowering.notYet(node, shape)
	}
	arguments := call.Arguments.Nodes
	wrapped := ast.SkipParentheses(arguments[0])
	if wrapped.Kind != ast.KindCallExpression {
		return nil, true, lowering.notYet(node, shape)
	}
	wrapper := wrapped.AsCallExpression()
	if wrapper.QuestionDotToken != nil ||
		!lowering.isLibraryGlobal(wrapper.Expression, "Object") ||
		len(wrapper.Arguments.Nodes) != 1 ||
		hasSpread(wrapped) {
		return nil, true, lowering.notYet(node, shape)
	}
	value := ast.SkipParentheses(wrapper.Arguments.Nodes[0])
	receiver := ast.SkipParentheses(arguments[2])
	if !ast.IsIdentifier(value) ||
		!ast.IsIdentifier(receiver) ||
		lowering.symbol(value) == nil ||
		lowering.symbol(value) != lowering.symbol(receiver) {
		return nil, true, lowering.notYet(node, shape)
	}
	// Object(null) and Object(undefined) are a new empty object, which no backend's read sees: the
	// node is only ever made for a value the checker proved is neither.
	proven := lowering.checker.GetTypeAtLocation(value)
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 ||
		lowering.includesNull(proven) ||
		lowering.includesUndefined(proven) {
		return nil, true, lowering.notYet(
			value,
			"Reflect.get of a value that may be null or undefined (Object(value) is then a new object)",
		)
	}
	keyType := lowering.checker.GetTypeAtLocation(arguments[1])
	if held, known := lowering.representation(
		keyType,
	); !known || held != ir.String || lowering.includesNull(keyType) || lowering.includesUndefined(keyType) {
		return nil, true, lowering.notYet(
			arguments[1],
			"Reflect.get by a key that isn't a string (symbol keys and key coercion)",
		)
	}
	object, expressionError := lowering.expression(value)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if object.Type() != ir.Union {
		return nil, true, lowering.notYet(
			value,
			"Reflect.get of a value not held as unknown (a read by its runtime shape needs the union's tag)",
		)
	}
	key, expressionError := lowering.expression(arguments[1])
	if expressionError != nil {
		return nil, true, expressionError
	}
	if key.Type() != ir.String {
		return nil, true, lowering.notYet(
			arguments[1],
			"Reflect.get by a key that isn't a string (symbol keys and key coercion)",
		)
	}
	if checkDynamicKeyedGetterSyntaxError := lowering.checkDynamicKeyedGetterSyntax(
		node,
	); checkDynamicKeyedGetterSyntaxError != nil {
		return nil, true, checkDynamicKeyedGetterSyntaxError
	}
	if lowering.dynamicKeyedSite == nil {
		lowering.dynamicKeyedSite = node
	}
	lowering.graphDynamicSite = node
	return ir.DynamicKeyedProperty{Object: object, Key: key}, true, nil
}

// dynamicKeyedGetterReturns are the representations a class getter's result can have where a
// run-time key read calls it: the runtime boxes each into the union it gives (union.c, keyed_getter).
var dynamicKeyedGetterReturns = []ir.Type{
	ir.Number,
	ir.Boolean,
	ir.String,
	ir.Object,
	ir.Array,
	ir.Map,
	ir.MaybeNumber,
	ir.Closure,
	ir.MaybeBoolean,
	ir.Union,
	ir.Record,
}

// A run-time key read calls whichever class getter its key names, from inside the runtime, where no
// pass that follows calls (effects, exceptions, borrowing, purity, regions) sees the call. So where a
// program reads by a run-time key, every class getter it has must be one none of them needs to see,
// judged by its syntax alone: one return statement of literals and this.<field>, a data field, never
// a getter or a method, joined by operators, with no call, no new and no assignment, which can't throw
// either. An object literal's getter and a static one stop at run time instead, so they aren't asked.
const (
	dynamicKeyedGetterSyntax    = "in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)"
	dynamicKeyedGetterPrimitive = "in a program with a class getter that converts an object field to a primitive (its toString or valueOf would run inside the read, unseen)"
)

// dynamicKeyedRead names the run-time key read a getter judgment stops at: Reflect.get's call, or a
// readonly view's index (viewKeyedRead).
func dynamicKeyedRead(site *ast.Node) string {
	if site.Kind == ast.KindCallExpression {
		return "Reflect.get by a run-time key"
	}
	return "a readonly view read by a run-time key"
}

// checkDynamicKeyedGetterSyntax judges every module class's getters at the read itself, by their
// syntax, before any of their bodies need lowering.
func (lowering *lowering) checkDynamicKeyedGetterSyntax(node *ast.Node) error {
	for _, declaration := range lowering.classes {
		for _, member := range declaration.Members() {
			if member.Kind != ast.KindGetAccessor || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
				continue
			}
			if reason := lowering.dynamicKeyedGetterBody(member); reason != "" {
				return lowering.notYet(node, dynamicKeyedRead(node)+" "+reason)
			}
		}
	}
	return nil
}

// checkDynamicKeyedGetters runs once every function is lowered, over every getter the program's
// classes hold, those checkDynamicKeyedGetterSyntax couldn't reach among them. Each getter's result
// must be one the runtime boxes, and null only as a union, since a reference held bare says null and
// undefined alike as NULL.
func (lowering *lowering) checkDynamicKeyedGetters() error {
	if lowering.dynamicKeyedSite == nil {
		return nil
	}
	for _, class := range lowering.result.Classes {
		if class.Literal || class.Static {
			continue
		}
		for _, accessor := range class.Accessors {
			if accessor.Getter < 0 {
				continue
			}
			if !slices.Contains(dynamicKeyedGetterReturns, lowering.result.Functions[accessor.Getter].Returns) {
				return lowering.notYet(lowering.dynamicKeyedSite, dynamicKeyedRead(lowering.dynamicKeyedSite)+
					" in a program with a class getter whose result has no union form yet")
			}
			declaration := lowering.getterDeclarations[accessor.Getter]
			if declaration == nil {
				return lowering.notYet(
					lowering.dynamicKeyedSite,
					dynamicKeyedRead(lowering.dynamicKeyedSite)+" "+dynamicKeyedGetterSyntax,
				)
			}
			if reason := lowering.dynamicKeyedGetterBody(declaration); reason != "" {
				return lowering.notYet(
					lowering.dynamicKeyedSite,
					dynamicKeyedRead(lowering.dynamicKeyedSite)+" "+reason,
				)
			}
			result := lowering.checker.GetTypeAtLocation(declaration.Name())
			if held, known := lowering.representation(
				result,
			); lowering.includesNull(result) && (!known || held != ir.Union) {
				return lowering.notYet(lowering.dynamicKeyedSite, dynamicKeyedRead(lowering.dynamicKeyedSite)+
					" in a program with a class getter that may give null held bare (null and undefined slot tags)")
			}
		}
	}
	return nil
}

// dynamicKeyedGetterBody says why a getter's body isn't one return statement of an expression
// dynamicKeyedGetterExpression admits, or "".
func (lowering *lowering) dynamicKeyedGetterBody(declaration *ast.Node) string {
	body := declaration.Body()
	if body == nil || body.Kind != ast.KindBlock || len(body.AsBlock().Statements.Nodes) != 1 {
		return dynamicKeyedGetterSyntax
	}
	statement := body.AsBlock().Statements.Nodes[0]
	if statement.Kind != ast.KindReturnStatement || statement.AsReturnStatement().Expression == nil {
		return dynamicKeyedGetterSyntax
	}
	return lowering.dynamicKeyedGetterExpression(statement.AsReturnStatement().Expression, true)
}

// dynamicKeyedGetterConverting are the binary operators a getter may join its fields and literals with
// that convert an operand: arithmetic, ordering, loose equality and bitwise. Each runs ToPrimitive on
// an object, which calls its toString or valueOf, so their operands must be primitives.
var dynamicKeyedGetterConverting = []ast.Kind{
	ast.KindPlusToken,
	ast.KindMinusToken,
	ast.KindAsteriskToken,
	ast.KindSlashToken,
	ast.KindPercentToken,
	ast.KindAsteriskAsteriskToken,
	ast.KindLessThanToken,
	ast.KindGreaterThanToken,
	ast.KindLessThanEqualsToken,
	ast.KindGreaterThanEqualsToken,
	ast.KindEqualsEqualsToken,
	ast.KindExclamationEqualsToken,
	ast.KindAmpersandToken,
	ast.KindBarToken,
	ast.KindCaretToken,
	ast.KindLessThanLessThanToken,
	ast.KindGreaterThanGreaterThanToken,
	ast.KindGreaterThanGreaterThanGreaterThanToken,
}

// dynamicKeyedGetterPassing are the binary operators that never convert an operand: strict equality
// compares without it, and &&, || and ?? test truthiness or presence, which calls nothing, and give an
// operand as it is, so their operands stand where the operator does.
var dynamicKeyedGetterPassing = []ast.Kind{
	ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
	ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
}

// dynamicKeyedGetterExpression says why an expression isn't literals and this.<field> reads joined by
// operators, a template, typeof or ?:, or "". bare is whether the expression's value is given as it
// is, the whole returned expression, a ?: operand there, or typeof's operand: only there may a field
// whose declared type isn't a primitive stand, since anywhere else it's converted.
func (lowering *lowering) dynamicKeyedGetterExpression(node *ast.Node, bare bool) string {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return ""
	case ast.KindParenthesizedExpression:
		return lowering.dynamicKeyedGetterExpression(node.AsParenthesizedExpression().Expression, bare)
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken != nil || ast.SkipParentheses(access.Expression).Kind != ast.KindThisKeyword {
			return dynamicKeyedGetterSyntax
		}
		field := lowering.checker.GetSymbolAtLocation(node.Name())
		if field == nil || field.Flags&ast.SymbolFlagsProperty == 0 || accessorSymbol(field) {
			return dynamicKeyedGetterSyntax
		}
		if !bare && !dynamicKeyedPrimitive(lowering.checker.GetTypeOfSymbol(field)) {
			return dynamicKeyedGetterPrimitive
		}
		return ""
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operand := false
		switch {
		case slices.Contains(dynamicKeyedGetterPassing, binary.OperatorToken.Kind):
			operand = bare
		case !slices.Contains(dynamicKeyedGetterConverting, binary.OperatorToken.Kind):
			return dynamicKeyedGetterSyntax
		}
		if reason := lowering.dynamicKeyedGetterExpression(binary.Left, operand); reason != "" {
			return reason
		}
		return lowering.dynamicKeyedGetterExpression(binary.Right, operand)
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindExclamationToken:
			// Truthiness calls nothing.
			return lowering.dynamicKeyedGetterExpression(unary.Operand, true)
		case ast.KindMinusToken, ast.KindPlusToken, ast.KindTildeToken:
			return lowering.dynamicKeyedGetterExpression(unary.Operand, false)
		}
		return dynamicKeyedGetterSyntax
	case ast.KindTypeOfExpression:
		return lowering.dynamicKeyedGetterExpression(node.AsTypeOfExpression().Expression, true)
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		// The condition is tested for truthiness, which calls nothing; the branches stand where the
		// conditional does.
		if reason := lowering.dynamicKeyedGetterExpression(conditional.Condition, true); reason != "" {
			return reason
		}
		if reason := lowering.dynamicKeyedGetterExpression(conditional.WhenTrue, bare); reason != "" {
			return reason
		}
		return lowering.dynamicKeyedGetterExpression(conditional.WhenFalse, bare)
	case ast.KindTemplateExpression:
		for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
			if reason := lowering.dynamicKeyedGetterExpression(span.AsTemplateSpan().Expression, false); reason != "" {
				return reason
			}
		}
		return ""
	}
	return dynamicKeyedGetterSyntax
}

// dynamicKeyedPrimitive reports whether a field's declared type is a primitive, or a union of
// primitives with undefined or null: what ToPrimitive gives back as it is, calling nothing.
func dynamicKeyedPrimitive(declared *checker.Type) bool {
	members := []*checker.Type{declared}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		members = declared.Types()
	}
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) == 0 {
			return false
		}
	}
	return true
}
