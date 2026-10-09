package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// This adapter admits optional string data, including undefined payloads.
// Accessors and optional receivers still require their own effect/storage adapter.
func (l *lowering) optionalStringViewType(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	hasString, hasUndefined := false, false
	for _, member := range target.Types() {
		switch {
		case member.Flags()&checker.TypeFlagsUndefined != 0:
			hasUndefined = true
		case member.Flags()&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral) != 0:
			hasString = true
		default:
			return false
		}
	}
	return hasString && hasUndefined
}
func (l *lowering) optionalViewAlias(node *ast.Node) bool {
	field := l.checker.GetSymbolAtLocation(node.Name())
	return field != nil && !accessorSymbol(field) && node.AsPropertyAccessExpression().QuestionDotToken == nil && l.optionalStringViewType(l.checker.GetTypeOfSymbol(field))
}

func (l *lowering) optionalReceiverViewAlias(node *ast.Node) bool {
	field := l.checker.GetSymbolAtLocation(node.Name())
	if field == nil || accessorSymbol(field) || field.Flags&ast.SymbolFlagsOptional != 0 {
		return false
	}
	of, known := l.representation(l.checker.GetNonMissingTypeOfSymbol(field))
	return known && (of == ir.Number || of == ir.String)
}
