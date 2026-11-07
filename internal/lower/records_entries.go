package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Main's growing literals with structural views use boxed Object storage. Keep
// that representation for the proven literal origins it already supports; a
// homogeneous dictionary cannot acquire those fixed-object parameter views.
func (lowering *lowering) enumerationRecordStorageType(subjectType *checker.Type) bool {
	infos := lowering.checker.GetIndexInfosOfType(subjectType)
	properties := lowering.checker.GetPropertiesOfType(subjectType)
	if len(infos) != 1 ||
		infos[0].IsReadonly() ||
		infos[0].KeyType().Flags() != checker.TypeFlagsString ||
		infos[0].ValueType().Flags()&checker.TypeFlagsUnion == 0 ||
		len(properties) == 0 {
		return false
	}
	for _, property := range properties {
		if property.Flags&ast.SymbolFlagsOptional != 0 || lowering.checker.GetTypeOfSymbol(
			property,
		).Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike) == 0 {
			return false
		}
	}
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return false
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found {
			return true
		}
		if node.Kind == ast.KindObjectLiteralExpression &&
			lowering.slotType(node) == subjectType &&
			lowering.enumerationInModule(ast.GetSourceFileOfNode(node).AsSourceFile()) &&
			lowering.entriesRecordLiteral(node) {
			found = true
			return true
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}

// Readonly index views admitted by main's enumeration proof keep fixed-object
// storage. This does not admit mutable dictionary aliases or readonly writes.
func (lowering *lowering) enumerationReadonlyType(subjectType *checker.Type) bool {
	infos := lowering.checker.GetIndexInfosOfType(subjectType)
	if len(infos) != 1 || !infos[0].IsReadonly() || infos[0].KeyType().Flags() != checker.TypeFlagsString {
		return false
	}
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return false
	}
	for _, module := range modules {
		if lowering.enumerationInModule(module) {
			return true
		}
	}
	return false
}

// readonlyObjectView reports whether a type is a readonly string-index view and nothing else,
// Readonly<Record<string, V>>, held in the viewed object's own fixed storage because the program's
// enumeration proof keeps it there (enumerationReadonlyType). Such a view is the object it views:
// tsc's getEnumMembers reads an enum's table through one (#52zja36).
func (lowering *lowering) readonlyObjectView(subjectType *checker.Type) bool {
	return subjectType.Flags()&checker.TypeFlagsObject != 0 &&
		len(lowering.checker.GetPropertiesOfType(subjectType)) == 0 &&
		len(lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindCall)) == 0 &&
		len(lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindConstruct)) == 0 &&
		lowering.enumerationReadonlyType(subjectType)
}
