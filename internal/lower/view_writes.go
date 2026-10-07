package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Slot certificates describe the declared logical type, independently of the
// current payload. Unsupported contracts have no certificate and fail closed.
func (l *lowering) slotContract(node *ast.Node, target *checker.Type) ir.ViewContractID {
	if target.Flags()&checker.TypeFlagsUndefined != 0 {
		id, err := l.viewContract(node, target)
		if err != nil {
			return 0
		}
		return id
	}
	if !interfaceScalar(target) && target.Flags()&checker.TypeFlagsObject == 0 {
		return 0
	}
	of, known := l.representation(target)
	if !known || (of != ir.Number && of != ir.Boolean && of != ir.String && of != ir.MaybeNumber && of != ir.MaybeBoolean && of != ir.Object) {
		return 0
	}
	if of == ir.Object && len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return 0
	}
	saved := len(l.result.ViewContracts)
	types := map[int]ir.ViewContractID{}
	for key, id := range l.result.ViewContractTypes {
		types[key] = id
	}
	id, err := l.viewContract(node, target)
	if err != nil {
		l.result.ViewContracts = l.result.ViewContracts[:saved]
		l.result.ViewContractTypes = types
		return 0
	}
	return id
}

func (l *lowering) optionalViewWriteField(name string) {
	if l.result == nil {
		return
	}
	if l.result.OptionalViewFields == nil {
		l.result.OptionalViewFields = map[string]bool{}
	}
	l.result.OptionalViewFields[name] = true
}

func freshOptionalReceiver(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		return freshOptionalReceiver(node.AsAsExpression().Expression)
	}
	return node.Kind == ast.KindObjectLiteralExpression
}

func (l *lowering) neverOptionalReceiver(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		return l.neverOptionalReceiver(node.AsAsExpression().Expression)
	}
	return l.isNever(node)
}
