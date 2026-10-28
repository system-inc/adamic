package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// refuseVoidRead refuses a read of a value typed exactly void (H2, ruled Oct 18, #x2gt39p). tsc lets
// a function that returns a value stand as one that returns void, so a void value can hold anything:
// `const make: () => void = () => 42; make() === undefined` is false on Node. Adamic releases what a
// void call returns (#f418fbz), so a read of it would see undefined where Node sees 42. A void value is
// unobservable instead: reading a binding of one, comparing it, returning it from a function whose
// return type isn't exactly void, passing it, and binding it to a wider type are refused, in .a and .ts
// alike. Relating a value-returning function to a void function type stays legal, and so do
// #se1zcgr's discards: `return f()` and a concise arrow body inside a function whose own return type is
// exactly void, or for an async function one that resolves to exactly void.
func (l *lowering) refuseVoidRead(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || ast.IsDeclarationName(node) || ast.IsAssignmentTarget(node) {
		return nil
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.AsPropertyAccessExpression().Name() == node {
		return nil // The access is the read, never the name it reads.
	}
	if l.checker.GetTypeAtLocation(node).Flags() != checker.TypeFlagsVoid || l.brandRead(node) {
		return nil
	}
	read := l.voidReadRole(node)
	if read == "" {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: "a read of a void value, as " + read, Fix: "call it as a statement, or give the function a return type (adamic/no-void-read)"}
}

func init() { refusalRules.register(3250, "refuseVoidRead", (*lowering).refuseVoidRead) }

// voidReadRole is what the place a void value stands in does with it, or "" where it's discarded. A
// conditional's branch, and the right of &&, || and ??, stand where the whole expression stands.
func (l *lowering) voidReadRole(node *ast.Node) string {
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil {
		return "a value"
	}
	switch parent.Kind {
	case ast.KindExpressionStatement, ast.KindVoidExpression:
		return ""
	case ast.KindForStatement:
		statement := parent.AsForStatement()
		if statement.Initializer == site || statement.Incrementor == site {
			return ""
		}
		return "a loop condition"
	case ast.KindAwaitExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
		// Each hands on the value it wraps: it stands where the whole expression stands (#chmg7wr).
		return l.voidReadRole(parent)
	case ast.KindConditionalExpression:
		if parent.AsConditionalExpression().Condition == site {
			return "a condition"
		}
		return l.voidReadRole(parent)
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken:
			if binary.Left == site {
				return ""
			}
			return l.voidReadRole(parent)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			if binary.Right == site {
				return l.voidReadRole(parent)
			}
			return "the left operand of " + scanner.TokenToString(binary.OperatorToken.Kind)
		case ast.KindEqualsToken:
			return "an assigned value"
		}
		return "an operand of " + scanner.TokenToString(binary.OperatorToken.Kind)
	case ast.KindReturnStatement:
		if l.returnsExactlyVoid(ast.GetContainingFunction(parent)) {
			return ""
		}
		return "a returned value"
	case ast.KindArrowFunction:
		if parent.Body() == site {
			if l.returnsExactlyVoid(parent) {
				return ""
			}
			return "a returned value"
		}
	case ast.KindVariableDeclaration, ast.KindPropertyDeclaration, ast.KindParameter, ast.KindBindingElement:
		// A binding that stays void keeps nothing anyone can see: a read of its name is refused there.
		// One typed wider would hand the value on.
		if name := parent.Name(); name != nil && name.Kind == ast.KindIdentifier && l.checker.GetTypeAtLocation(name).Flags() == checker.TypeFlagsVoid {
			return ""
		}
		return "a binding's initializer"
	case ast.KindCallExpression, ast.KindNewExpression:
		if parent.Kind != ast.KindCallExpression || parent.AsCallExpression().Expression != site {
			return "an argument"
		}
	case ast.KindTemplateSpan:
		return "a template literal part"
	case ast.KindTypeOfExpression:
		return "typeof's operand"
	}
	return "a value"
}

// returnsExactlyVoid is whether function's own return type is void, not void | X and not unknown, or for
// an async function whether what it resolves to is: only there is a returned void value nobody can read.
func (l *lowering) returnsExactlyVoid(function *ast.Node) bool {
	if function == nil || !ast.IsFunctionLike(function) {
		return false
	}
	signature := l.checker.GetSignatureFromDeclaration(function)
	if signature == nil {
		return false
	}
	returns := l.checker.GetReturnTypeOfSignature(signature)
	if ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) {
		returns = checker.Checker_getAwaitedType(l.checker, returns)
	}
	return returns != nil && returns.Flags() == checker.TypeFlagsVoid
}

// brandRead is whether node reads a property off a value that is a primitive in every member of its
// type, as tsc's __String brand `string & { __escapedIdentifier: void }` is. A primitive carries no own
// property, so the read is undefined on Node as it is here: the void it names was never a value.
func (l *lowering) brandRead(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	// Not GetNonNullableType: it drops void as nullish, which would strip `void & { … }` to its object.
	receiver := l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression)
	members := []*checker.Type{receiver}
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		members = receiver.Types()
	}
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 && member.Flags()&checker.TypeFlagsIntersection == 0 {
			continue // An optional chain's missing receiver reads nothing.
		}
		if member.Flags()&checker.TypeFlagsIntersection == 0 {
			return false
		}
		primitive := false
		for _, part := range member.Types() {
			primitive = primitive || part.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|checker.TypeFlagsESSymbolLike|checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0
		}
		if !primitive {
			return false
		}
	}
	return true
}
