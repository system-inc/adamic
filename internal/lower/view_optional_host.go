package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An immutable unknown slot can carry an absent optional host method. The
// producer signature is still certified at every reached callable read; this
// relation supplies no callable certificate and never invents a nominal field.
func (l *lowering) optionalHostViewRelation(source, target *checker.Type) bool {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(source) || isClassInstance(target) || len(l.checker.GetIndexInfosOfType(source)) != 0 || len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if l.checker.GetPropertyOfType(source, property.Name) == nil {
			return false
		}
	}
	relaxed := false
	for _, property := range l.checker.GetPropertiesOfType(source) {
		wanted := l.checker.GetPropertyOfType(target, property.Name)
		if wanted == nil {
			return false
		}
		actualType, wantedType := l.checker.GetTypeOfSymbol(property), l.checker.GetTypeOfSymbol(wanted)
		if l.checker.IsReadonlySymbol(property) && actualType.Flags()&checker.TypeFlagsUnknown != 0 && wanted.Flags&ast.SymbolFlagsOptional != 0 && wanted.Flags&ast.SymbolFlagsMethod != 0 && l.runtimeViewCallableShape(wantedType) {
			relaxed = true
			continue
		}
		if !l.checker.IsTypeAssignableTo(wantedType, actualType) || wanted.Flags&ast.SymbolFlagsOptional != 0 && property.Flags&ast.SymbolFlagsOptional == 0 {
			return false
		}
	}
	return relaxed
}

// Keep the context adapters inside this unit's immutable-unknown-slot views.
// Follow local aliases, but do not exempt ordinary methods from their receiver
// discipline merely because their declarations happen to be optional.
func (l *lowering) optionalHostViewReceiver(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
		return false
	}
	seen[symbol] = true
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return false
	}
	initializer = ast.SkipParentheses(initializer)
	if initializer.Kind == ast.KindAsExpression {
		return l.optionalHostViewRelation(l.concrete(l.checker.GetTypeAtLocation(initializer.AsAsExpression().Expression)), l.concrete(l.checker.GetTypeAtLocation(initializer)))
	}
	return l.optionalHostViewReceiver(initializer, seen)
}

// Observing an optional method as the left operand does not detach it: &&
// returns undefined or evaluates the right operand. Binding below is restricted
// to certified receiver-free closures, which the ordinary callable read enforces.
func (l *lowering) optionalHostMethodObservation(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression || !l.optionalHostViewReceiver(node.AsPropertyAccessExpression().Expression, map[*ast.Symbol]bool{}) {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsOptional|ast.SymbolFlagsMethod) != ast.SymbolFlagsOptional|ast.SymbolFlagsMethod || !l.runtimeViewCallableShape(l.checker.GetTypeOfSymbol(symbol)) {
		return false
	}
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.Left != node || binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken || !ast.IsIdentifier(binary.Right) {
			return false
		}
		right := l.symbol(binary.Right)
		return right != nil && len(right.Declarations) == 1 && right.Declarations[0].Kind == ast.KindFunctionDeclaration
	}
	if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node && parent.Parent != nil {
		return l.optionalHostBindOperand(parent.Parent) == node
	}
	return false
}

func (l *lowering) optionalHostBindOperand(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "bind" || callee.AsPropertyAccessExpression().QuestionDotToken == nil || len(call.Arguments.Nodes) != 1 {
		return nil
	}
	operand := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if operand.Kind != ast.KindPropertyAccessExpression || !l.optionalHostViewReceiver(operand.AsPropertyAccessExpression().Expression, map[*ast.Symbol]bool{}) {
		return nil
	}
	receiver := ast.SkipParentheses(operand.AsPropertyAccessExpression().Expression)
	argument := ast.SkipParentheses(call.Arguments.Nodes[0])
	if !ast.IsIdentifier(receiver) || !ast.IsIdentifier(argument) || l.symbol(receiver) != l.symbol(argument) {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(operand)
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsOptional|ast.SymbolFlagsMethod) != ast.SymbolFlagsOptional|ast.SymbolFlagsMethod {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(symbol)), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].Parameters()) != 1 || !l.runtimeViewCallableShape(l.checker.GetTypeOfSymbol(symbol)) {
		return nil
	}
	parameter, known := l.representation(l.checker.GetTypeOfSymbol(signatures[0].Parameters()[0]))
	result, resultKnown := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0]))
	if !known || !resultKnown || !viewCallableScalarRepresentation(parameter) || !viewCallableScalarRepresentation(result) {
		return nil
	}
	return operand
}

// Bind creates a fresh wrapper, preserving observable function identity. The
// checked read admits only recorded receiver-free closure producers; opaque and
// literal-method producers retain their existing unknown-signature refusal.
func (l *lowering) optionalHostBindCall(node *ast.Node) (ir.Expression, bool, error) {
	operand := l.optionalHostBindOperand(node)
	if operand == nil {
		return nil, false, nil
	}
	value, err := l.expression(operand)
	if err != nil {
		return nil, true, err
	}
	signature := l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(operand)), checker.SignatureKindCall)[0]
	parameter, _ := l.representation(l.checker.GetTypeOfSymbol(signature.Parameters()[0]))
	result, _ := l.representation(l.checker.GetReturnTypeOfSignature(signature))
	outer := len(l.result.Functions)
	captured := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "bound_source", Type: ir.Closure, Function: outer, Captured: true})
	l.result.Functions = append(l.result.Functions, ir.Function{})
	wrapper := len(l.result.Functions)
	argument := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "bound_argument", Type: parameter, Function: wrapper})
	read := ir.Read{Local: captured, Of: ir.Closure}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "optional_host_bound", Closure: true, Parameters: []int{argument}, Environment: []int{captured}, Returns: result, Body: []ir.Statement{ir.Return{Value: ir.CallClosure{Closure: read, Arguments: []ir.Expression{ir.Read{Local: argument, Of: parameter}}, Returns: result}}}})
	l.result.Functions[outer] = ir.Function{Name: "optional_host_bind", Parameters: []int{captured}, Returns: ir.Closure, Body: []ir.Statement{ir.If{Condition: ir.IsUndefined{Value: read}, Then: []ir.Statement{ir.Return{Value: ir.Undefined{Of: ir.Closure}}}}, ir.Return{Value: ir.MakeClosure{Function: wrapper}}}}
	return ir.Call{Function: outer, Arguments: []ir.Expression{value}, Returns: ir.Closure}, true, nil
}

func (l *lowering) optionalHostCondition(node *ast.Node) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken || !l.optionalHostMethodObservation(binary.Left) || !ast.IsIdentifier(binary.Right) {
		return nil, false, nil
	}
	symbol := l.symbol(binary.Right)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
		return nil, false, nil
	}
	left, err := l.expression(binary.Left)
	if err != nil {
		return nil, true, err
	}
	right, err := l.expression(binary.Right)
	if err != nil {
		return nil, true, err
	}
	if left.Type() != ir.Closure || right.Type() != ir.Closure {
		return nil, true, l.notYet(node, "an optional host method condition with non-callable storage")
	}
	// The right operand is a hoisted function declaration: reading its identity
	// eagerly has no effects or initialization failure. Calls remain conditional.
	b := l.libraryArrayBuilder([]ir.Expression{left, right})
	return b.finish("optional_host_condition", ir.Conditional{Condition: ir.IsUndefined{Value: b.read(b.parameters[0])}, WhenTrue: ir.Undefined{Of: ir.Closure}, WhenNot: b.read(b.parameters[1])}), true, nil
}
