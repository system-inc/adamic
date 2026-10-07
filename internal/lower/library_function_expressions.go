package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// functionExpression uses the same calling convention and counted captures as arrows. Ordinary
// functions have their own dynamic this; until that convention exists, every use of it is refused.
func (l *lowering) functionExpression(node *ast.Node) (ir.Expression, error) {
	if node.Kind == ast.KindFunctionExpression && node.AsFunctionExpression().AsteriskToken != nil || node.Kind == ast.KindFunctionDeclaration && node.AsFunctionDeclaration().AsteriskToken != nil {
		return nil, l.notYet(node, "a generator function expression")
	}
	if len(node.TypeParameters()) != 0 {
		return nil, l.notYet(node, "a generic function expression")
	}
	for _, parameter := range node.Parameters() {
		if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
			return nil, l.notYet(parameter, "a function expression with a this parameter (dynamic receivers are not implemented)")
		}
	}
	var invalid error
	var visit ast.Visitor
	visit = func(inner *ast.Node) bool {
		if invalid != nil {
			return true
		}
		// An arrow inherits this; another ordinary function or method owns a different this.
		if ast.IsFunctionLike(inner) && inner.Kind != ast.KindArrowFunction {
			return false
		}
		if inner.Kind == ast.KindThisKeyword && !ast.IsPartOfTypeNode(inner) {
			invalid = &Refused{Where: l.program.Where(inner), What: "this in a function expression", Fix: "dynamic receivers are not implemented; capture a named object, or use an arrow in a method"}
			return true
		}
		return inner.ForEachChild(visit)
	}
	node.Body().ForEachChild(visit)
	if invalid != nil {
		return nil, invalid
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "function_expression", Closure: true}
	if name := node.Name(); node.Kind == ast.KindFunctionExpression && name != nil {
		function.Name = name.Text()
		if l.locals == nil {
			l.locals = map[*ast.Symbol]int{}
		}
		local := len(l.result.Locals)
		l.locals[l.symbol(name)] = local
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name.Text(), Type: ir.Closure, Function: index})
		l.noteLocal(local, l.checker.GetTypeAtLocation(node), name)
		function.Body = []ir.Statement{ir.Declare{Local: local, Value: ir.ClosureSelf{}}}
	}
	l.result.Functions = append(l.result.Functions, function)
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: index, node: node})
	if l.instance != nil {
		l.instance.templates = append(l.instance.templates, template{closure: index, proven: l.checker.GetTypeAtLocation(node), isClosure: true})
	}
	l.closures = append(l.closures, index)
	outerThis := l.this
	l.this = -1
	err := l.lowerFunction(index, node, -1)
	l.this = outerThis
	l.closures = l.closures[:len(l.closures)-1]
	if err != nil {
		return nil, err
	}
	return ir.MakeClosure{Function: index}, nil
}
