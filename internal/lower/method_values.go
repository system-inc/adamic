package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Arrows share the method's receiver; ordinary nested functions have their own.
func methodReadsThis(method *ast.Node) bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindThisKeyword {
			found = true
			return true
		}
		if ast.IsFunctionLike(node) && node.Kind != ast.KindArrowFunction {
			return false
		}
		node.ForEachChild(visit)
		return false
	}
	if method.Body() != nil {
		method.ForEachChild(visit)
	}
	return found
}

func (l *lowering) thisFreeMethod(access *ast.Node, symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	found := false
	for _, root := range l.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if declaration.Kind == ast.KindMethodSignature {
				continue
			}
			if declaration.Kind != ast.KindMethodDeclaration || declaration.Body() == nil || methodReadsThis(declaration) {
				return false
			}
			found = true
		}
	}
	// Cover overrides and concrete implementations of structural method signatures.
	readsThis := false
	for _, module := range l.program.Files() {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindMethodDeclaration && l.methodNamed(node, symbol.Name) && l.methodMayReach(access, node) {
				found = true
				readsThis = readsThis || methodReadsThis(node)
			}
			node.ForEachChild(visit)
			return false
		}
		module.AsNode().ForEachChild(visit)
	}
	return found && !readsThis
}

func methodBindRead(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name().Text() == "bind" && called(node.Parent)
}

func (l *lowering) methodValueAllowed(node *ast.Node, symbol *ast.Symbol) bool {
	if l.libraryMember(node) {
		return false
	}
	return methodBindRead(node) || strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") || l.thisFreeMethod(node, symbol)
}

func (l *lowering) userMethodValue(node *ast.Node) (ir.Expression, bool, error) {
	var bound *ast.Node
	access := node
	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "bind" {
			return nil, false, nil
		}
		access = ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
		if access.Kind != ast.KindPropertyAccessExpression {
			return nil, false, nil
		}
		symbol := l.checker.GetSymbolAtLocation(access)
		if symbol == nil || symbol.Flags&ast.SymbolFlagsMethod == 0 || l.libraryMember(access) {
			return nil, false, nil
		}
		if len(node.AsCallExpression().Arguments.Nodes) != 1 {
			return nil, true, l.notYet(node, "a method bind with partial arguments; bind only its receiver")
		}
		bound = node.AsCallExpression().Arguments.Nodes[0]
	} else if access.Kind != ast.KindPropertyAccessExpression || called(access) {
		return nil, false, nil
	}
	symbol := l.checker.GetSymbolAtLocation(access)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsMethod == 0 || l.libraryMember(access) {
		return nil, false, nil
	}
	if !l.methodValueAllowed(access, symbol) {
		return nil, false, nil
	}
	if access.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, true, l.notYet(access, "an optional method extraction")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(access), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(access, "an overloaded method value")
	}
	for _, parameter := range signatures[0].Parameters() {
		if of, known := l.censusCallableParameter(parameter); !known || censusCallableSlotless(of) {
			return nil, true, l.notYet(access, "a method value parameter without a callable slot")
		}
	}
	result := l.checker.GetReturnTypeOfSignature(signatures[0])
	if result.Flags()&checker.TypeFlagsVoid == 0 {
		if of, known := l.representation(result); !known || censusCallableSlotless(of) {
			return nil, true, l.notYet(access, "a method value result without a callable slot")
		}
	}
	receiver := access.AsPropertyAccessExpression().Expression
	for _, declaration := range symbol.Declarations {
		for _, parameter := range declaration.Parameters() {
			if parameter.AsParameterDeclaration().DotDotDotToken != nil {
				return nil, true, l.notYet(access, "a method value with rest parameters")
			}
		}
		if declaration.Kind == ast.KindMethodDeclaration && declaration.Parent.Kind == ast.KindClassDeclaration {
			if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsStatic) {
				if _, err := l.staticInstance(l.staticClass(l.checker.GetTypeAtLocation(receiver), declaration.Parent)); err != nil {
					return nil, true, err
				}
				continue
			}
			if _, err := l.instantiate(declaration.Parent, l.checker.GetTypeAtLocation(receiver), access); err != nil {
				return nil, true, err
			}
		}
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if object.Type() != ir.Object {
		return nil, true, l.notYet(access, "a method value without an object receiver")
	}
	value := ir.Property{Object: object, Name: l.fieldName(access.Name()), Of: ir.Closure, Extracted: true}
	if bound != nil {
		receiverType := l.checker.GetTypeAtLocation(receiver)
		boundType := l.checker.GetTypeAtLocation(bound)
		if !l.checker.IsTypeAssignableTo(boundType, receiverType) || (isClassInstance(receiverType) && !isClassInstance(boundType)) {
			return nil, true, l.notYet(bound, "a bound receiver outside its method's represented object type")
		}
		if l.includesUndefined(boundType) {
			return nil, true, l.notYet(bound, "an absent bound receiver; narrow the object before binding")
		}
		value.Bound, err = l.expression(bound)
		if err != nil {
			return nil, true, err
		}
		if value.Bound.Type() != ir.Object {
			return nil, true, l.notYet(bound, "a bound method without an object receiver")
		}
	}
	return value, true, nil
}

// An extracted method can receive undefined. Check at the property access, not on entry,
// so earlier side effects and methods that only observe this keep JavaScript's order.
func (l *lowering) extractedThisRead(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return value, nil
	}
	var method *ast.Node
	for owner := node.Parent; owner != nil; owner = owner.Parent {
		if owner.Kind == ast.KindMethodDeclaration {
			method = owner
			break
		}
		if ast.IsFunctionLike(owner) && owner.Kind != ast.KindArrowFunction {
			break
		}
	}
	if method == nil {
		return value, nil
	}
	methodSymbol := l.checker.GetSymbolAtLocation(method.Name())
	if methodSymbol == nil {
		return nil, l.notYet(node, "a method receiver without a resolved declaration name")
	}
	extracted := false
	for _, module := range l.program.Files() {
		var visit ast.Visitor
		visit = func(candidate *ast.Node) bool {
			if candidate.Kind == ast.KindPropertyAccessExpression && !called(candidate) && !methodBindRead(candidate) && candidate.Name().Text() == methodSymbol.Name && l.methodMayReach(candidate, method) && !l.libraryMember(candidate) {
				extracted = true
			}
			candidate.ForEachChild(visit)
			return false
		}
		module.AsNode().ForEachChild(visit)
	}
	if !extracted {
		return value, nil
	}
	if parent.Kind != ast.KindPropertyAccessExpression {
		return nil, l.notYet(node, "an extracted receiver escaping before its property read; read its property in the method")
	}
	if parent.AsPropertyAccessExpression().QuestionDotToken != nil {
		return nil, l.notYet(node, "an optional property read through an extracted receiver")
	}
	if write := parent.Parent; write != nil && write.Kind == ast.KindBinaryExpression && write.AsBinaryExpression().Left == parent && write.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		return nil, l.notYet(node, "an extracted receiver used for a plain property write; its receiver check must follow the right-hand side")
	}
	message := "Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
	b := l.libraryArrayBuilder([]ir.Expression{value})
	self := b.read(b.parameters[0])
	b.body = append(b.body, ir.If{Condition: ir.IsUndefined{Value: self}, Then: []ir.Statement{ir.Throw{Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant(message)}, Name: ir.StringConstant{Index: l.constant("TypeError")}}}}})
	return b.finish("extracted_this", self), nil
}

func (l *lowering) methodOverridesSymbol(method *ast.Node, symbol *ast.Symbol) bool {
	if method.Parent.Kind != ast.KindClassDeclaration {
		return false
	}
	for _, root := range l.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if declaration.Parent.Kind == ast.KindClassDeclaration && l.classView(l.checker.GetTypeAtLocation(method.Parent.Name()), declaration.Parent) != nil {
				return true
			}
		}
	}
	return false
}

func (l *lowering) methodMayReach(access, method *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(access)
	if symbol == nil {
		return false
	}
	for _, root := range l.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if declaration == method || l.methodOverridesSymbol(method, root) {
				return true
			}
			if declaration.Kind == ast.KindMethodSignature {
				owner := method.Parent
				actual := l.checker.GetTypeAtLocation(owner)
				if owner.Kind == ast.KindClassDeclaration {
					actual = l.checker.GetTypeAtLocation(owner.Name())
				}
				receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.AsPropertyAccessExpression().Expression))
				if l.checker.IsTypeAssignableTo(actual, receiver) {
					return true
				}
			}
		}
	}
	return false
}

func (l *lowering) methodNamed(method *ast.Node, name string) bool {
	if method.Name() == nil {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(method.Name())
	return symbol != nil && symbol.Name == name
}
