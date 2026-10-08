package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// nullSentinelType accepts only the migrated string reference kind, with null and/or undefined.
func (l *lowering) nullSentinelType(proven *checker.Type) bool {
	proven = l.concrete(proven)
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	reference := false
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			continue
		}
		if member.Flags()&checker.TypeFlagsStringLike == 0 {
			return false
		}
		reference = true
	}
	return reference
}

// Project TypeScript's == null is a nullish observation; Adamic's .a coercion refusal remains.
func (l *lowering) nullishComparison(node *ast.Node) bool {
	return l.program.UsesProjectOptions() && isNullishComparison(node)
}

func isNullishComparison(node *ast.Node) bool {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken.Kind
	if operator != ast.KindEqualsEqualsToken && operator != ast.KindExclamationEqualsToken {
		return false
	}
	return ast.SkipParentheses(binary.Left).Kind == ast.KindNullKeyword || ast.SkipParentheses(binary.Right).Kind == ast.KindNullKeyword
}
