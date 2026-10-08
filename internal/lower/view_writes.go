package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Slot certificates describe the declared logical type, independently of the
// current payload. Unsupported contracts have no certificate and fail closed.
func (l *lowering) slotContract(node *ast.Node, target *checker.Type) ir.ViewContractID {
	if isClassInstance(l.checker.GetNonNullableType(target)) {
		return l.mapNominalEntrySlot(node, target)
	}
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
	if !supportedSlotContract(l.result, id, map[ir.ViewContractID]bool{}) {
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

// Lazy descriptors must never become a write certificate.
func supportedSlotContract(program *ir.Program, id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id == 0 {
		return false
	}
	if seen[id] {
		return true
	}
	seen[id] = true
	contract := program.ViewContracts[id-1]
	if contract.Kind == ir.ViewUnknown || contract.Kind == ir.ViewNullable || contract.Kind == ir.ViewMap || contract.Kind == ir.ViewDictionary || contract.Unsupported != "" {
		return false
	}
	for _, field := range contract.Fields {
		if !supportedSlotContract(program, field.Contract, seen) {
			return false
		}
	}
	for _, member := range contract.Members {
		if !supportedSlotContract(program, member, seen) {
			return false
		}
	}
	if contract.Element != 0 && !supportedSlotContract(program, contract.Element, seen) {
		return false
	}
	return true
}

// Only finite class/nullish references use this checked union slot adapter.
func (l *lowering) nominalUnionWriteTarget(target *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(target)
	if symbol == nil {
		return false
	}
	declared := l.concrete(l.checker.GetTypeOfSymbol(symbol))
	of, known := l.representation(declared)
	return known && of == ir.Union && isClassInstance(l.checker.GetNonNullableType(declared)) && l.mapNominalEntrySlot(target, declared) != 0
}
