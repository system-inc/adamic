package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Verify the whole switch before using the numeric dispatch. A union's storage
// alone does not distinguish a nullable number from a number mixed with strings.
func (l *lowering) nullishNumericSwitch(statement *ast.SwitchStatement, storage ir.Type) bool {
	if storage != ir.MaybeNumber && storage != ir.Union {
		return false
	}
	proven := l.checker.GetTypeAtLocation(statement.Expression)
	numeric, nullish := false, false
	for _, member := range flattenedMembers(proven) {
		flags := member.Flags()
		switch {
		case flags&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0:
			nullish = true
		case flags&checker.TypeFlagsNumberLike != 0:
			if flags&checker.TypeFlagsEnumLike != 0 && !l.numericEnum(l.enumIdentity(member)) {
				return false
			}
			numeric = true
		default:
			return false
		}
	}
	if !numeric || !nullish {
		return false
	}
	hasLabel := false
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		if clause.Kind == ast.KindDefaultClause {
			continue
		}
		if !l.nullishNumericCase(clause.AsCaseOrDefaultClause().Expression) {
			return false
		}
		hasLabel = true
	}
	return hasLabel
}

func (l *lowering) nullishNumericCase(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNumericLiteral || (ast.IsIdentifier(node) && l.valueSymbol(node) == l.checker.GetUndefinedSymbol()) {
		return true
	}
	if node.Kind == ast.KindPrefixUnaryExpression {
		prefix := node.AsPrefixUnaryExpression()
		return (prefix.Operator == ast.KindPlusToken || prefix.Operator == ast.KindMinusToken) &&
			prefix.Operand.Kind == ast.KindNumericLiteral
	}
	if ast.IsOptionalChain(node) {
		return false
	}
	if node.Kind == ast.KindElementAccessExpression {
		index := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if index.Kind != ast.KindStringLiteral && index.Kind != ast.KindNoSubstitutionTemplateLiteral {
			return false
		}
	}
	member := l.enumMember(node)
	return member != nil && l.numericEnum(l.symbol(member.Parent.Name()))
}

func nullishNumericLabel(label ir.Expression) bool {
	_, missing := label.(ir.Undefined)
	return missing || label.Type() == ir.Number
}

func nullishNumericValue(value ir.Expression) ir.Expression {
	if value.Type() == ir.MaybeNumber {
		return ir.Unwrap{Value: value}
	}
	return ir.Narrow{Value: value, To: ir.Number}
}

func nullishNumericTest(value, nullish, label ir.Expression) ir.Expression {
	if _, missing := label.(ir.Undefined); missing {
		return ir.IsUndefined{Value: value}
	}
	// A normal enum label still reads its object in source order, including
	// on the nullish road. Only the numeric comparison is skipped there.
	return ir.Conditional{
		Condition: nullish,
		WhenTrue:  ir.Comma{Left: label, Right: ir.BooleanConstant{Value: false}},
		WhenNot:   ir.Binary{Operator: ir.Equal, Left: nullishNumericValue(value), Right: label},
	}
}
