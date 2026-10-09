package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Main's growing literals with structural views use boxed Object storage. Keep
// that representation for the proven literal origins it already supports; a
// homogeneous dictionary cannot acquire those fixed-object parameter views.
func (l *lowering) enumerationRecordStorageType(t *checker.Type) bool {
	infos := l.checker.GetIndexInfosOfType(t)
	properties := l.checker.GetPropertiesOfType(t)
	if len(infos) != 1 || infos[0].IsReadonly() || infos[0].KeyType().Flags() != checker.TypeFlagsString || infos[0].ValueType().Flags()&checker.TypeFlagsUnion == 0 || len(properties) == 0 {
		return false
	}
	for _, property := range properties {
		if property.Flags&ast.SymbolFlagsOptional != 0 || l.checker.GetTypeOfSymbol(property).Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike) == 0 {
			return false
		}
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found {
			return true
		}
		if node.Kind == ast.KindObjectLiteralExpression && l.checker.GetContextualType(node, checker.ContextFlagsNone) == t && l.enumerationInModule(ast.GetSourceFileOfNode(node).AsSourceFile()) && l.entriesRecordLiteral(node) {
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
func (l *lowering) enumerationReadonlyType(t *checker.Type) bool {
	infos := l.checker.GetIndexInfosOfType(t)
	if len(infos) != 1 || !infos[0].IsReadonly() || infos[0].KeyType().Flags() != checker.TypeFlagsString {
		return false
	}
	if _, supported := l.recordInfo(t); supported {
		return false
	}
	value := l.concrete(infos[0].ValueType())
	if value.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	for _, member := range value.Types() {
		if member.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike) == 0 {
			return false
		}
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	for _, module := range modules {
		if l.enumerationInModule(module) {
			return true
		}
	}
	return false
}
