package lower

import (
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// constantFieldName accepts only keys with no runtime evaluation: literals,
// literal string concatenations, and const enum members. A singleton type alone
// is insufficient: a call returning that type still has to run before its value.
func (l *lowering) constantFieldName(node *ast.Node) (name string, known bool) {
	defer func() {
		if known && !nativeDataFieldName(name) {
			known = false
		}
	}()
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindStringLiteral || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return node.Text(), true
	}
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindPlusToken {
		binary := node.AsBinaryExpression()
		left, leftKnown := l.constantFieldName(binary.Left)
		right, rightKnown := l.constantFieldName(binary.Right)
		// Only string addition is concatenation; numeric keys must not be joined.
		if leftKnown && rightKnown && l.checker.GetTypeAtLocation(binary.Left).Flags()&checker.TypeFlagsStringLike != 0 && l.checker.GetTypeAtLocation(binary.Right).Flags()&checker.TypeFlagsStringLike != 0 {
			return left + right, true
		}
	}
	if member := l.enumMember(node); member != nil && ast.HasSyntacticModifier(member.Parent, ast.ModifierFlagsConst) {
		value, err := l.enumConstant(member)
		if err != nil {
			return "", false
		}
		switch value := value.(type) {
		case ir.StringConstant:
			return l.result.Strings[value.Index], true
		case ir.NumberConstant:
			// Safe integers have the same decimal spelling in Go and JavaScript.
			if math.Abs(value.Value) <= 9007199254740991 && math.Trunc(value.Value) == value.Value {
				if value.Value == 0 {
					return "0", true
				}
				return strconv.FormatFloat(value.Value, 'f', -1, 64), true
			}
		}
	}
	return "", false
}

// runtimeEnumFieldName resolves a fixed enum key without discarding its read.
// Its caller evaluates key before the field value, so a temporal-dead-zone read
// remains observable even though the enum's constant determines the shape.
func (l *lowering) runtimeEnumFieldName(node *ast.Node) (string, ir.Expression, ir.Expression, bool, error) {
	member := l.enumMember(node)
	if member == nil || ast.HasSyntacticModifier(member.Parent, ast.ModifierFlagsConst) {
		return "", nil, nil, false, nil
	}
	constant, err := l.enumConstant(member)
	if err != nil {
		return "", nil, nil, false, err
	}
	name := ""
	switch constant := constant.(type) {
	case ir.StringConstant:
		name = l.result.Strings[constant.Index]
	case ir.NumberConstant:
		if math.Abs(constant.Value) > 9007199254740991 || math.Trunc(constant.Value) != constant.Value {
			return "", nil, nil, false, nil
		}
		if constant.Value == 0 {
			name = "0"
		} else {
			name = strconv.FormatFloat(constant.Value, 'f', -1, 64)
		}
	default:
		return "", nil, nil, false, nil
	}
	if !nativeDataFieldName(name) {
		return "", nil, nil, false, nil
	}
	key, err := l.expression(node)
	return name, key, constant, true, err
}

// Native shapes use C string names. Reject embedded NUL before a computed
// key can alias a different public data field.
func nativeDataFieldName(name string) bool {
	return !strings.ContainsRune(name, 0)
}
