package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// clockNullishString admits only string members and the two independently observable absent states.
func (l *lowering) clockNullishString(proven *checker.Type) bool {
	if !l.includesNull(proven) || !l.includesUndefined(proven) {
		return false
	}
	stringMember := false
	for _, member := range proven.Types() {
		flags := member.Flags()
		if flags&checker.TypeFlagsStringLike != 0 {
			stringMember = true
			continue
		}
		if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			return false
		}
	}
	return stringMember
}

// Keep the binary worker's pointer-only null contract unchanged. This carrier compares its tags.
func clockStringComparison(operator ast.Kind, left, right ir.Expression) (ir.Expression, bool) {
	if operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
		return nil, false
	}
	if left.Type() != ir.NullishString && right.Type() != ir.NullishString {
		return nil, false
	}
	value, other := left, right
	if value.Type() != ir.NullishString {
		value, other = other, value
	}
	var test ir.Expression
	switch other.(type) {
	case ir.Null:
		test = ir.IsNull{Value: value}
	case ir.Undefined:
		test = ir.IsUndefined{Value: value}
	default:
		test = ir.Binary{Operator: ir.Equal, Left: fit(left, ir.Union), Right: fit(right, ir.Union)}
	}
	if operator == ast.KindExclamationEqualsEqualsToken {
		test = ir.Unary{Operator: ir.Not, Operand: test}
	}
	return test, true
}

// Fields keep their declared carrier even when the checker narrowed the read after an earlier check.
func (l *lowering) clockStringProperty(node *ast.Node) (ir.Expression, bool, error) {
	field := l.checker.GetSymbolAtLocation(node.Name())
	if field == nil {
		return nil, false, nil
	}
	declared, known := l.representation(l.checker.GetTypeOfSymbol(field))
	if !known || declared != ir.NullishString {
		return nil, false, nil
	}
	access := node.AsPropertyAccessExpression()
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	if object.Type() != ir.Object {
		return nil, false, nil
	}
	read := ir.Expression(ir.Property{Object: object, Name: node.Name().Text(), Of: declared, Optional: access.QuestionDotToken != nil, Absent: field.Flags&ast.SymbolFlagsOptional != 0, Class: l.classOf(node)})
	narrowed, known := l.representation(l.checker.GetTypeAtLocation(node))
	if !known || narrowed == declared || l.clockStringObservation(node) {
		return read, true, nil
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	held := b.read(b.parameters[0])
	matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant("string")}})
	if l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	if l.includesNull(l.checker.GetTypeAtLocation(node)) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsNull{Value: held}}
	}
	message := "union member where the checker narrowed it away: a call since the narrowing put it back"
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	return b.finish("clock_narrowed_string", ir.Narrow{Value: held, To: narrowed}), true, nil
}

// Observations preserve both tags. A destination that excludes null still needs a narrowing check.
func (l *lowering) clockStringObservation(node *ast.Node) bool {
	if comparedWithUndefined(node) {
		return true
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil {
		if parent.Kind == ast.KindTypeOfExpression {
			return true
		}
		if parent.Kind == ast.KindBinaryExpression {
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken && binary.Left == at {
				return true
			}
		}
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	return contextual != nil && l.clockNullishString(l.concrete(contextual))
}
