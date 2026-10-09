package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// functionExpression lowers ordinary function expressions and switch declarations with the same
// calling convention and counted captures as arrows. Ordinary
// functions keep their represented dynamic this separate from lexical captures.
func (l *lowering) functionExpression(node *ast.Node) (ir.Expression, error) {
	if node.Body() == nil {
		return nil, l.notYet(node, "a function without a body")
	}
	if node.Kind == ast.KindFunctionExpression && node.AsFunctionExpression().AsteriskToken != nil || node.Kind == ast.KindFunctionDeclaration && node.AsFunctionDeclaration().AsteriskToken != nil {
		return nil, l.notYet(node, "a generator function expression")
	}
	if len(node.TypeParameters()) != 0 {
		return nil, l.notYet(node, "a generic function expression")
	}
	var receiverType *checker.Type
	for _, parameter := range node.Parameters() {
		if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
			receiverType = l.checker.GetTypeAtLocation(parameter.Name())
		}
	}
	var visit ast.Visitor
	visit = func(inner *ast.Node) bool {
		if ast.IsFunctionLike(inner) && inner.Kind != ast.KindArrowFunction {
			return false
		}
		if inner.Kind == ast.KindThisKeyword && !ast.IsPartOfTypeNode(inner) && receiverType == nil {
			receiverType = l.checker.GetTypeAtLocation(inner)
		}
		return inner.ForEachChild(visit)
	}
	node.Body().ForEachChild(visit)
	if receiverType != nil {
		at := node
		for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
			at = at.Parent
		}
		stored := at.Parent != nil && (at.Parent.Kind == ast.KindPropertyAssignment || at.Parent.Kind == ast.KindPropertyDeclaration)
		for owner := at.Parent; owner != nil && !stored; owner = owner.Parent {
			if owner.Kind == ast.KindGetAccessor {
				stored = true
				break
			}
			if ast.IsFunctionLike(owner) {
				break
			}
		}
		if !stored {
			return nil, l.notYet(node, "a function expression with a this parameter outside represented member storage")
		}
		if held, known := l.representation(receiverType); !known || held != ir.Object {
			return nil, l.notYet(node, "a function expression with an unrepresented dynamic receiver")
		}
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "function_expression", Closure: true}
	if node.Name() != nil {
		function.Name = node.Name().Text()
	}
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
	self := -1
	if receiverType != nil {
		self = len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: index})
		l.noteLocal(self, receiverType, node)
	}
	l.this = self
	err := l.lowerFunction(index, node, self)
	l.this = outerThis
	l.closures = l.closures[:len(l.closures)-1]
	if err != nil {
		return nil, err
	}
	return ir.MakeClosure{Function: index}, nil
}
