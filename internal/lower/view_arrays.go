package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The shared contract builder owns interning and recursion. This hook builds only
// the declared element contract; it never inspects array contents at a cast/read.
// An error from a nested object/union/callable contract must not be discarded.
func (l *lowering) viewArrayContract(node *ast.Node, target *checker.Type, buildElement func(*checker.Type) error) (bool, error) {
	if checker.IsTupleType(target) {
		return true, l.notYet(node, "a checked tuple view with per-position optional and rest contracts")
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
