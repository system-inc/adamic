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
// a string written as an identifier or a string literal, which runs nothing, so v can't be rebound
// between the target and the receiver (#m26ebzy). For a class instance, Object(v) returns v and
// the getter receives that very instance as this. The runtime stops on primitive targets before
// dispatch. Every other Reflect.get is NotYet here, as is Reflect or Object(x) as a value elsewhere.
func (l *lowering) reflectGet(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "get" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Reflect") {
		return nil, false, nil
	}
	const shape = "Reflect.get other than Reflect.get(Object(value), key, value) with one identifier as value"
	if call.QuestionDotToken != nil || callee.AsPropertyAccessExpression().QuestionDotToken != nil || len(call.Arguments.Nodes) != 3 || hasSpread(node) {
		return nil, true, l.notYet(node, shape)
	}
	arguments := call.Arguments.Nodes
	wrapped := ast.SkipParentheses(arguments[0])
	if wrapped.Kind != ast.KindCallExpression {
		return nil, true, l.notYet(node, shape)
	}
	wrapper := wrapped.AsCallExpression()
	if wrapper.QuestionDotToken != nil || !l.isLibraryGlobal(wrapper.Expression, "Object") || len(wrapper.Arguments.Nodes) != 1 || hasSpread(wrapped) {
		return nil, true, l.notYet(node, shape)
	}
	value := ast.SkipParentheses(wrapper.Arguments.Nodes[0])
	receiver := ast.SkipParentheses(arguments[2])
	if !ast.IsIdentifier(value) || !ast.IsIdentifier(receiver) || l.symbol(value) == nil || l.symbol(value) != l.symbol(receiver) {
		return nil, true, l.notYet(node, shape)
	}
	// JavaScript evaluates the key between the target and the receiver, so a key that runs code could
	// rebind value between them, and the getter would run on the receiver, not the target the read uses
	// (#m26ebzy). One symbol is one instance only where the key runs nothing: an identifier or a string
	// literal (Adamic's ruling (a1) on #bbssdcy).
	if key := ast.SkipParentheses(arguments[1]); !ast.IsIdentifier(key) && key.Kind != ast.KindStringLiteral {
		return nil, true, l.notYet(arguments[1], "Reflect.get by a key other than an identifier or a string literal (a key that runs code could rebind the value between the target and the receiver)")
	}
	// Object(null) and Object(undefined) are a new empty object, which no backend's read sees: the
	// node is only ever made for a value the checker proved is neither.
	proven := l.checker.GetTypeAtLocation(value)
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 || l.includesNull(proven) || l.includesUndefined(proven) {
		return nil, true, l.notYet(value, "Reflect.get of a value that may be null or undefined (Object(value) is then a new object)")
	}
	keyType := l.checker.GetTypeAtLocation(arguments[1])
	if held, known := l.representation(keyType); !known || held != ir.String || l.includesNull(keyType) || l.includesUndefined(keyType) {
		return nil, true, l.notYet(arguments[1], "Reflect.get by a key that isn't a string (symbol keys and key coercion)")
	}
	object, err := l.expression(value)
	if err != nil {
		return nil, true, err
	}
	if object.Type() != ir.Union {
		return nil, true, l.notYet(value, "Reflect.get of a value not held as unknown (a read by its runtime shape needs the union's tag)")
	}
	key, err := l.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String {
		return nil, true, l.notYet(arguments[1], "Reflect.get by a key that isn't a string (symbol keys and key coercion)")
	}
	if err := l.checkDynamicKeyedGetterSyntax(node); err != nil {
		return nil, true, err
	}
	if l.reflectKeyedSite == nil {
		l.reflectKeyedSite = node
	}
	l.graphDynamicSite = node
	return ir.DynamicKeyedProperty{Object: object, Key: key, Reflect: true}, true, nil
}

// dynamicKeyedGetterReturns are the representations a class getter's result can have where a
// run-time key read calls it: the runtime boxes each into the union it gives (union.c, keyed_getter).
var dynamicKeyedGetterReturns = []ir.Type{
	ir.Number, ir.Boolean, ir.String, ir.Object, ir.Array, ir.Map, ir.MaybeNumber, ir.Closure, ir.MaybeBoolean, ir.Union, ir.Record,
}

// Readonly views retain the syntax judgment for getters. Reflect.get has a possible getter call
// in the IR, with a throw edge and an escaping receiver. Keep the separate refusal for implicit
// object-to-primitive conversion.
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
func (l *lowering) checkDynamicKeyedGetterSyntax(node *ast.Node) error {
	for _, declaration := range l.classes {
		for _, member := range declaration.Members() {
			if member.Kind != ast.KindGetAccessor || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
				continue
			}
			if reason := l.dynamicKeyedGetterReason(node, member, -1); reason != "" {
				return l.notYet(node, dynamicKeyedRead(node)+" "+reason)
			}
		}
	}
	return nil
}

func (l *lowering) dynamicKeyedGetterReason(site, declaration *ast.Node, getter int) string {
	reason := l.dynamicKeyedGetterBody(declaration)
	// The early check has no IR. Defer this reason until the getter's
	// lowered body and the bodies of its possible callees are available.
	if site.Kind == ast.KindCallExpression && reason == dynamicKeyedGetterSyntax &&
		(getter < 0 || l.keyedGetterReadsAndThrows(getter)) {
		return ""
	}
	return reason
}

// checkDynamicKeyedGetters runs once every function is lowered, over every getter the program's
// classes hold, those checkDynamicKeyedGetterSyntax couldn't reach among them. Each getter's result
// must be one the runtime boxes, and null only as a union, since a reference held bare says null and
// undefined alike as NULL.
func (l *lowering) checkDynamicKeyedGetters() error {
	site := l.dynamicKeyedSite
	if site == nil {
		site = l.reflectKeyedSite
	}
	if site == nil {
		return nil
	}
	for _, class := range l.result.Classes {
		if class.Literal || class.Static {
			continue
		}
		for _, accessor := range class.Accessors {
			if accessor.Getter < 0 {
				continue
			}
			if !slices.Contains(dynamicKeyedGetterReturns, l.result.Functions[accessor.Getter].Returns) {
				return l.notYet(site, dynamicKeyedRead(site)+" in a program with a class getter whose result has no union form yet")
			}
			declaration := l.getterDeclarations[accessor.Getter]
			if declaration == nil {
				return l.notYet(site, dynamicKeyedRead(site)+" "+dynamicKeyedGetterSyntax)
			}
			if reason := l.dynamicKeyedGetterReason(site, declaration, accessor.Getter); reason != "" {
				return l.notYet(site, dynamicKeyedRead(site)+" "+reason)
			}
			result := l.checker.GetTypeAtLocation(declaration.Name())
			if held, known := l.representation(result); l.includesNull(result) && (!known || held != ir.Union) {
				return l.notYet(site, dynamicKeyedRead(site)+" in a program with a class getter that may give null held bare (null and undefined slot tags)")
			}
		}
	}
	return nil
}

// dynamicKeyedGetterBody says why a getter's body isn't one return statement of an expression
// dynamicKeyedGetterExpression admits, or "".
func (l *lowering) dynamicKeyedGetterBody(declaration *ast.Node) string {
	body := declaration.Body()
	if body == nil || body.Kind != ast.KindBlock || len(body.AsBlock().Statements.Nodes) != 1 {
		return dynamicKeyedGetterSyntax
	}
	statement := body.AsBlock().Statements.Nodes[0]
	if statement.Kind != ast.KindReturnStatement || statement.AsReturnStatement().Expression == nil {
		return dynamicKeyedGetterSyntax
	}
	return l.dynamicKeyedGetterExpression(statement.AsReturnStatement().Expression, true)
}

// dynamicKeyedGetterConverting are the binary operators a getter may join its fields and literals with
// that convert an operand: arithmetic, ordering, loose equality and bitwise. Each runs ToPrimitive on
// an object, which calls its toString or valueOf, so their operands must be primitives.
var dynamicKeyedGetterConverting = []ast.Kind{
	ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken, ast.KindSlashToken, ast.KindPercentToken, ast.KindAsteriskAsteriskToken,
	ast.KindLessThanToken, ast.KindGreaterThanToken, ast.KindLessThanEqualsToken, ast.KindGreaterThanEqualsToken,
	ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
	ast.KindAmpersandToken, ast.KindBarToken, ast.KindCaretToken, ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken, ast.KindGreaterThanGreaterThanGreaterThanToken,
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
func (l *lowering) dynamicKeyedGetterExpression(node *ast.Node, bare bool) string {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return ""
	case ast.KindParenthesizedExpression:
		return l.dynamicKeyedGetterExpression(node.AsParenthesizedExpression().Expression, bare)
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken != nil || ast.SkipParentheses(access.Expression).Kind != ast.KindThisKeyword {
			return dynamicKeyedGetterSyntax
		}
		field := l.checker.GetSymbolAtLocation(node.Name())
		if field == nil || field.Flags&ast.SymbolFlagsProperty == 0 || accessorSymbol(field) {
			return dynamicKeyedGetterSyntax
		}
		if !bare && !dynamicKeyedPrimitive(l.checker.GetTypeOfSymbol(field)) {
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
		if reason := l.dynamicKeyedGetterExpression(binary.Left, operand); reason != "" {
			return reason
		}
		return l.dynamicKeyedGetterExpression(binary.Right, operand)
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindExclamationToken:
			// Truthiness calls nothing.
			return l.dynamicKeyedGetterExpression(unary.Operand, true)
		case ast.KindMinusToken, ast.KindPlusToken, ast.KindTildeToken:
			return l.dynamicKeyedGetterExpression(unary.Operand, false)
		}
		return dynamicKeyedGetterSyntax
	case ast.KindTypeOfExpression:
		return l.dynamicKeyedGetterExpression(node.AsTypeOfExpression().Expression, true)
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		// The condition is tested for truthiness, which calls nothing; the branches stand where the
		// conditional does.
		if reason := l.dynamicKeyedGetterExpression(conditional.Condition, true); reason != "" {
			return reason
		}
		if reason := l.dynamicKeyedGetterExpression(conditional.WhenTrue, bare); reason != "" {
			return reason
		}
		return l.dynamicKeyedGetterExpression(conditional.WhenFalse, bare)
	case ast.KindTemplateExpression:
		for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
			if reason := l.dynamicKeyedGetterExpression(span.AsTemplateSpan().Expression, false); reason != "" {
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
