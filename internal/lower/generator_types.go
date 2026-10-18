package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func isGenerator(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

func (l *lowering) generatorType(proven *checker.Type) bool {
	return l.isLibraryType(l.concrete(proven), "Generator", "IterableIterator")
}
func (l *lowering) generatorTypes(node *ast.Node) (*ir.GeneratorTypes, error) {
	result := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(node)))
	arguments := l.checker.GetTypeArguments(result)
	if len(arguments) != 3 {
		return nil, l.notYet(node, "generator suspension result needs separate yield, return and next types")
	}
	held := [3]ir.Type{}
	for i, argument := range arguments {
		if argument.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0 {
			held[i] = ir.Object
			continue
		}
		var known bool
		held[i], known = l.representation(argument)
		if !known {
			return nil, l.notYet(node, "generator suspension type "+l.checker.TypeToString(argument)+" has no represented storage")
		}
	}
	return &ir.GeneratorTypes{Yield: held[0], Return: held[1], Next: held[2], NextAllowsUndefined: l.includesUndefined(arguments[2]) || arguments[2].Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0}, nil
}
func (l *lowering) generatorYield(node *ast.Node) (ir.Expression, error) {
	if l.function == nil || l.function.Generator == nil {
		return nil, l.notYet(node, "yield outside a lowered generator")
	}
	yield := node.AsYieldExpression()
	value := ir.Expression(ir.Undefined{Of: ir.Object})
	var err error
	if yield.Expression != nil {
		value, err = l.expression(yield.Expression)
		if err != nil {
			return nil, err
		}
	}
	result := ir.GeneratorYield{Where: l.program.Where(node), Value: value, Next: l.function.Generator.Next, Delegate: yield.AsteriskToken != nil}
	if result.Delegate {
		resultType := l.checker.GetTypeAtLocation(node)
		result.Result = ir.Object
		if resultType.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
			result.Result, err = l.typeOf(node)
		}
		if err != nil {
			return nil, err
		}
		proven := l.concrete(l.checker.GetTypeAtLocation(yield.Expression))
		if l.checker.IsArrayType(proven) {
			result.Kind = "array"
			result.Element, _ = l.representation(l.checker.GetTypeArguments(proven)[0])
		} else if l.isSet(yield.Expression) {
			result.Kind = "set"
			result.Element, err = l.setElement(yield.Expression)
		} else if !l.generatorType(proven) {
			return nil, l.notYet(node, "generator suspension delegate protocol is not known")
		}
		if err != nil {
			return nil, err
		}
		if result.Kind == "" {
			if !l.generatorOrigin(yield.Expression, 0) {
				return nil, l.notYet(yield.Expression, "generator suspension delegate has no proved generator factory origin")
			}
			if err := l.generatorProtocolStable(yield.Expression); err != nil {
				return nil, err
			}
			result.Kind = "generator"
		}
	}
	return result, nil
}
func (l *lowering) generatorCall(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind == ast.KindElementAccessExpression {
		access := callee.AsElementAccessExpression()
		if !l.symbolIterator(access.ArgumentExpression) || !l.generatorType(l.checker.GetTypeAtLocation(access.Expression)) {
			return nil, false, nil
		}
		if !l.generatorOrigin(access.Expression, 0) {
			return nil, true, l.notYet(access.Expression, "generator suspension receiver has no proved generator factory origin")
		}
		if len(call.Arguments.Nodes) != 0 {
			return nil, true, l.notYet(node, "generator iterator call with extra arguments")
		}
		if err := l.generatorProtocolStable(access.Expression); err != nil {
			return nil, true, err
		}
		receiver, err := l.expression(access.Expression)
		if err != nil {
			return nil, true, err
		}
		b := l.libraryArrayBuilder([]ir.Expression{receiver})
		object := b.read(b.parameters[0])
		return b.finish("generator_iterator", ir.CallClosure{Closure: ir.Property{Object: object, Name: iteratorSlot, Of: ir.Closure}, Arguments: []ir.Expression{object}, Returns: ir.Object}), true, nil
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	if !l.generatorType(l.checker.GetTypeAtLocation(access.Expression)) {
		return nil, false, nil
	}
	if !l.generatorOrigin(access.Expression, 0) {
		return nil, true, l.notYet(access.Expression, "generator suspension receiver has no proved generator factory origin")
	}
	method := access.Name().Text()
	if method != "next" && method != "return" && method != "throw" {
		return nil, true, l.notYet(node, "generator protocol method "+method)
	}
	if len(call.Arguments.Nodes) > 1 {
		return nil, true, l.notYet(node, "generator protocol call with extra arguments")
	}
	if err := l.generatorProtocolStable(access.Expression); err != nil {
		return nil, true, err
	}
	receiver, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	input := ir.Expression(ir.Undefined{Of: ir.Union})
	if len(call.Arguments.Nodes) == 1 {
		input, err = l.expression(call.Arguments.Nodes[0])
		if err != nil {
			return nil, true, err
		}
		if method == "throw" && !l.generatorErrorOrigin(call.Arguments.Nodes[0], 0) {
			return nil, true, l.notYet(node, "generator throw argument needs the represented Error completion protocol")
		}
		input = fit(input, ir.Union)
	}
	b := l.libraryArrayBuilder([]ir.Expression{receiver, input})
	object := b.read(b.parameters[0])
	sent := b.read(b.parameters[1])
	invoked := ir.CallClosure{Closure: ir.Property{Object: object, Name: method, Of: ir.Closure}, Arguments: []ir.Expression{object, sent}, Returns: ir.Object}
	return b.finish("generator_"+method, invoked), true, nil
}
func (l *lowering) iteratorResultType(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.iteratorResultType(member) {
				return false
			}
		}
		return true
	}
	return l.isLibraryType(proven, "IteratorYieldResult", "IteratorReturnResult")
}
func (l *lowering) generatorResultRead(node *ast.Node) (ir.Expression, bool, error) {
	access := node.AsPropertyAccessExpression()
	if access.Name().Text() != "value" || !l.iteratorResultType(l.checker.GetTypeAtLocation(access.Expression)) {
		return nil, false, nil
	}
	if l.dynamicReadHazard("value") {
		return nil, true, l.notYet(node, "generator result value needs a known data-field protocol")
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	held, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	value := ir.Expression(ir.DynamicProperty{Object: fit(object, ir.Union), Name: "value"})
	if held != ir.Union {
		value = ir.Narrow{Value: value, To: held}
	}
	return value, true, nil
}
func (l *lowering) generatorLibraryMember(member *ast.Symbol) bool {
	for _, root := range l.checker.GetRootSymbols(member) {
		for _, declaration := range root.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) && declaration.Parent != nil && declaration.Parent.Name() != nil && declaration.Parent.Name().Text() == "Generator" {
				return true
			}
		}
	}
	return false
}

// Error is structural in TypeScript. The native exception completion must keep
// a real Error's identity rather than trusting a compatible record annotation.
func (l *lowering) generatorErrorOrigin(node *ast.Node, depth int) bool {
	if node == nil || depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNewExpression {
		return l.isLibraryGlobal(node.AsNewExpression().Expression, "Error")
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if l.caught[symbol] {
		return true
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	return declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && l.generatorErrorOrigin(declaration.AsVariableDeclaration().Initializer, depth+1)
}
