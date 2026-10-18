package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (l *lowering) viewArrayContract(node *ast.Node, target *checker.Type, buildElement func(*checker.Type) error) (bool, error) {
	if checker.IsTupleType(target) {
		return true, l.notYet(node, "a checked tuple view with per-position optional and rest contracts")
	}
	if !l.checker.IsArrayType(target) {
		return false, nil
	}
	element := l.checker.GetElementTypeOfArrayType(target)
	if element == nil {
		return false, nil
	}
	if buildElement == nil {
		return true, l.notYet(node, "an array view without a recursive element contract builder")
	}
	return true, buildElement(element)
}
