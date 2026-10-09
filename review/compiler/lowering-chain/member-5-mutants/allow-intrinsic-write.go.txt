package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Intrinsic fast paths require the whole program to leave Symbol.iterator intact.
func (l *lowering) intrinsicIteratorWrite(node *ast.Node) error {
	return nil
	if node.Kind != ast.KindElementAccessExpression || !ast.IsAssignmentTarget(node) {
		return nil
	}
	access := node.AsElementAccessExpression()
	if !l.iteratorKey(access.ArgumentExpression, map[*ast.Symbol]bool{}) {
		return nil
	}
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	builtin := l.checker.IsArrayType(receiver) || checker.IsTupleType(receiver) || l.readonlyArrayView(receiver) != nil || l.typedArrayKind(receiver) != 0 || receiver.Flags()&checker.TypeFlagsStringLike != 0 || l.isLibraryType(receiver, "String", "Map", "ReadonlyMap", "Set", "ReadonlySet")
	if !builtin {
		target := ast.SkipParentheses(access.Expression)
		if target.Kind == ast.KindPropertyAccessExpression && target.Name().Text() == "prototype" {
			for _, name := range []string{"Array", "String", "Map", "Set", "Int8Array", "Uint8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array", "Int32Array", "Uint32Array", "Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array"} {
				builtin = builtin || l.isLibraryGlobal(target.AsPropertyAccessExpression().Expression, name)
			}
		}
	}
	if !builtin {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: "writing Symbol.iterator on a built-in or its prototype (adamic/intrinsic-iterator)", Fix: "use a separate object with its own Symbol.iterator method; built-in iteration cannot be overridden"}
}

func (l *lowering) iteratorKey(node *ast.Node, visited map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	if l.symbolIterator(node) {
		return true
	}
	proven := l.checker.GetTypeAtLocation(node)
	if proven.Flags()&checker.TypeFlagsUniqueESSymbol != 0 {
		if symbol := proven.Symbol(); symbol != nil && symbol.Name == "iterator" {
			for _, declaration := range symbol.Declarations {
				if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
					return true
				}
			}
		}
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || visited[symbol] {
		return false
	}
	visited[symbol] = true
	if len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := symbol.Declarations[0]
	initializer := declaration.AsVariableDeclaration().Initializer
	return initializer != nil && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && l.iteratorKey(initializer, visited)
}
