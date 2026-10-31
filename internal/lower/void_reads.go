package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// A void call's result is discarded natively and on the JavaScript backend, but Node keeps what the
// function returned: a () => void slot may hold () => 42. Both backends agree on undefined, so the
// oracle can't see it. An expression whose type is exactly void, not a union holding void, is
// refused wherever its value is used, and admitted only where its value is let go (#x2gt39p).
func (l *lowering) refuseVoidRead(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression {
		return nil
	}
	read := l.checker.GetTypeAtLocation
	if node.Kind == ast.KindShorthandPropertyAssignment {
		// { value } reads value, though its name is no expression node.
		symbol := l.checker.GetShorthandAssignmentValueSymbol(node)
		if symbol == nil {
			return nil
		}
		read = func(*ast.Node) *checker.Type { return l.checker.GetTypeOfSymbol(symbol) }
	} else if !ast.IsExpressionNode(node) || ast.IsPartOfTypeNode(node) {
		return nil
	} else if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.AsPropertyAccessExpression().Name() == node {
		// A field's name is read as part of its access, which is judged on its own.
		return nil
	}
	if source := read(node); source == nil || source.Flags() != checker.TypeFlagsVoid || l.voidDiscarded(node) {
		return nil
	}
	file := ast.GetSourceFileOfNode(node)
	text := file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()]
	if line, _, multiline := strings.Cut(text, "\n"); multiline {
		text = strings.TrimSpace(line) + " ..."
	}
	return &Refused{Where: l.program.Where(node), What: "a read of a void value (" + text + ")",
		Fix: "call it as a statement, or give the function a return type"}
}

// voidDiscarded reports whether a value at this position is let go, never read: an expression
// statement, void's operand, a comma's left, a for statement's initializer or incrementor, an
// assignment's target (written, not read), a return
// from a function whose own return type is exactly void, or an arrow's expression body when the
// arrow's is. A comma's right, a conditional's branch and the right of &&, || or ?? are the value
// of the expression around them, so they are let go only where it is.
func (l *lowering) voidDiscarded(node *ast.Node) bool {
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindExpressionStatement, ast.KindVoidExpression:
		return true
	case ast.KindForStatement:
		statement := parent.AsForStatement()
		return statement.Initializer == site || statement.Incrementor == site
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken:
			// The target is written, not read; the right is read into it.
			return binary.Left == site
		case ast.KindCommaToken:
			return binary.Left == site || l.voidDiscarded(parent)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return binary.Right == site && l.voidDiscarded(parent)
		}
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition != site && l.voidDiscarded(parent)
	case ast.KindReturnStatement:
		function := parent
		for function != nil && !ast.IsFunctionLike(function) {
			function = function.Parent
		}
		return function != nil && l.voidFunction(function)
	case ast.KindArrowFunction:
		return parent.AsArrowFunction().Body == site && l.voidFunction(parent)
	}
	return false
}

// voidFunction reports whether a function's own return type is exactly void. An async function's is
// a Promise, whose resolution a wider Promise view can still read.
func (l *lowering) voidFunction(function *ast.Node) bool {
	if ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) {
		return false
	}
	return l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(function)).Flags() == checker.TypeFlagsVoid
}
