package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Until every root consumer has an adapter, newly admitted nullable writable
// views are limited to indexed reads and writes. Scan the closed program
// conservatively; no alias analysis is used to excuse an unchecked consumer.
func (l *lowering) nullableWritableArrayConsumers(cast *ast.Node) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var found error
	array := func(node *ast.Node) bool {
		return node != nil && l.checker.IsArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)))
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		unsupported := false
		switch node.Kind {
		case ast.KindArrayBindingPattern:
			unsupported = true
		case ast.KindPropertyAccessExpression:
			unsupported = array(node.AsPropertyAccessExpression().Expression)
		case ast.KindForOfStatement:
			unsupported = array(node.AsForInOrOfStatement().Expression)
		case ast.KindForInStatement:
			unsupported = array(node.AsForInOrOfStatement().Expression)
		case ast.KindSpreadElement:
			unsupported = array(node.AsSpreadElement().Expression)
		case ast.KindCallExpression:
			for _, argument := range node.AsCallExpression().Arguments.Nodes {
				unsupported = unsupported || array(argument)
			}
		}
		if unsupported {
			found = l.notYet(cast, "a nullable writable array view consumer requiring a checked receiver")
			return true
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}
