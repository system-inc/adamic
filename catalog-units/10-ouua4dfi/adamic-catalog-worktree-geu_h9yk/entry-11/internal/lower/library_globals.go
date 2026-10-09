package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// These globals have stable identity. Their overloaded call signatures, prototypes and static
// properties are not ordinary closure or object fields, so values currently admit only equality
// and typeof. Calls still use the existing built-in dispatch, which does not read a global value.
func (l *lowering) libraryGlobalValue(node *ast.Node) (ir.Expression, bool, error) {
	for _, name := range []string{"Object", "Array", "String", "Number", "JSON", "Map", "Set"} {
		if !l.isLibraryGlobal(node, name) {
			continue
		}
		parent := node.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		if parent != nil && parent.Kind == ast.KindTypeOfExpression {
			return ir.LibraryGlobal{Name: name}, true, nil
		}
		if parent != nil && parent.Kind == ast.KindBinaryExpression {
			op := parent.AsBinaryExpression().OperatorToken.Kind
			if op == ast.KindEqualsEqualsEqualsToken || op == ast.KindExclamationEqualsEqualsToken {
				return ir.LibraryGlobal{Name: name}, true, nil
			}
		}
		return nil, true, l.notYet(node, name+" as a value outside equality or typeof (overloaded calls and static properties need their own representation)")
	}
	return nil, false, nil
}
