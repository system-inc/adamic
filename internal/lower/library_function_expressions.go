package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// functionExpression lowers ordinary function expressions and switch declarations with the same
// calling convention and counted captures as arrows. Ordinary
// functions have their own dynamic this; until that convention exists, every use of it is refused.
func (lowering *lowering) functionExpression(node *ast.Node) (ir.Expression, error) {
	if node.Body() == nil {
		return nil, lowering.notYet(node, "a function without a body")
	}
	if generatorDeclaration(node) {
		// A generator written in place lowers only through inPlaceGenerator, which its consumers ask,
		// so one reached here is used some other way (#qk8rztp S3b).
		switch {
		case generatorInPlace(node):
			return nil, lowering.inPlaceNotYet(node)
		case node.Name() != nil && calledInPlace(node) != nil:
			// Its name binds it inside its body, where a call through it would make a second frame
			// through a carrier no rule covers yet (Lower, on #qk8rztp S3b).
			return nil, lowering.notYet(node, "a generator expression with a name of its own (#qk8rztp)")
		}
		return nil, lowering.generatorNotYet(node, generatorForm(node))
	}
	return lowering.closureExpression(node)
}

// closureExpression lowers a function expression, or a switch declaration, as the closure it is:
// its body as a function of its own, with the cells it captures, and the MakeClosure that makes it.
// A generator written in place is lowered here too, once, by lowerInPlaceGenerator.
func (lowering *lowering) closureExpression(node *ast.Node) (ir.Expression, error) {
	if len(node.TypeParameters()) != 0 {
		return nil, lowering.notYet(node, "a generic function expression")
	}
	for _, parameter := range node.Parameters() {
		if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
			return nil, lowering.notYet(
				parameter,
				"a function expression with a this parameter (dynamic receivers are not implemented)",
			)
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
			invalid = &Refused{
				Where: lowering.program.Where(inner),
				What:  "this in a function expression",
				Fix:   "dynamic receivers are not implemented; capture a named object, or use an arrow in a method",
			}
			return true
		}
		return inner.ForEachChild(visit)
	}
	node.Body().ForEachChild(visit)
	if invalid != nil {
		return nil, invalid
	}
	index := len(lowering.result.Functions)
	function := ir.Function{Name: "function_expression", Closure: true}
	if node.Name() != nil {
		function.Name = node.Name().Text()
	}
	if name := node.Name(); node.Kind == ast.KindFunctionExpression && name != nil {
		function.Name = name.Text()
		if lowering.locals == nil {
			lowering.locals = map[*ast.Symbol]int{}
		}
		local := len(lowering.result.Locals)
		lowering.locals[lowering.symbol(name)] = local
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: name.Text(), Type: ir.Closure, Function: index},
		)
		lowering.noteLocal(local, lowering.checker.GetTypeAtLocation(node), name)
		function.Body = []ir.Statement{ir.Declare{Local: local, Value: ir.ClosureSelf{}}}
	}
	lowering.result.Functions = append(lowering.result.Functions, function)
	lowering.closureRecords = append(
		lowering.closureRecords,
		closureRecord{proven: lowering.concrete(lowering.checker.GetTypeAtLocation(node)), function: index, node: node},
	)
	if lowering.instance != nil {
		lowering.instance.templates = append(
			lowering.instance.templates,
			template{closure: index, proven: lowering.checker.GetTypeAtLocation(node), isClosure: true},
		)
	}
	lowering.closures = append(lowering.closures, index)
	outerThis := lowering.this
	lowering.this = -1
	lowerFunctionError := lowering.lowerFunction(index, node, -1)
	lowering.this = outerThis
	lowering.closures = lowering.closures[:len(lowering.closures)-1]
	if lowerFunctionError != nil {
		return nil, lowerFunctionError
	}
	return ir.MakeClosure{Function: index}, nil
}
