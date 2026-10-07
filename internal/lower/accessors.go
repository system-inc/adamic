package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func classMethod(node *ast.Node) bool {
	return node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor
}

// The colon cannot occur in an admitted method name, so the two descriptor halves
// have independent virtual slots without colliding with source methods.
func classMethodName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindGetAccessor:
		return "get:" + node.Name().Text()
	case ast.KindSetAccessor:
		return "set:" + node.Name().Text()
	}
	return node.Name().Text()
}

func accessorDeclaration(symbol *ast.Symbol, kind ast.Kind) *ast.Node {
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == kind {
				return declaration
			}
		}
	}
	return nil
}

func accessorSymbol(symbol *ast.Symbol) bool {
	return accessorDeclaration(symbol, ast.KindGetAccessor) != nil || accessorDeclaration(symbol, ast.KindSetAccessor) != nil
}

func accessorReadonly(c *checker.Checker, symbol *ast.Symbol) bool {
	return c.IsReadonlySymbol(symbol) || (accessorDeclaration(symbol, ast.KindGetAccessor) != nil && accessorDeclaration(symbol, ast.KindSetAccessor) == nil)
}

func (l *lowering) accessorRefusal(node *ast.Node) error {
	if node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
		if node.Parent.Kind == ast.KindObjectLiteralExpression {
			return l.notYet(node, "object-literal accessors (adamic/object-accessor)")
		}
		if node.Parent.Kind != ast.KindClassDeclaration {
			return l.notYet(node, "an accessor declaration outside a class (adamic/accessor-declaration)")
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
			return l.notYet(node, "static accessors (adamic/static-accessor)")
		}
		if !ast.IsIdentifier(node.Name()) {
			return l.notYet(node, "computed or private accessor names (adamic/accessor-name)")
		}
		if node.Body() == nil {
			return l.notYet(node, "abstract or ambient accessors (adamic/abstract-accessor)")
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression && accessorSymbol(l.checker.GetSymbolAtLocation(node.Name())) {
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken != nil {
			return l.notYet(node, "optional accessor access (adamic/optional-accessor)")
		}
		receiverType := l.checker.GetTypeAtLocation(access.Expression)
		if ast.SkipParentheses(access.Expression).Kind != ast.KindThisKeyword && receiverType.Flags()&checker.TypeFlagsTypeParameter != 0 {
			return l.notYet(node, "accessor dispatch through a type parameter (adamic/generic-accessor-receiver)")
		}
		if ast.SkipParentheses(access.Expression).Kind != ast.KindThisKeyword && !isClassInstance(l.checker.GetNonNullableType(l.concrete(receiverType))) && len(l.definedMembers(receiverType)) <= 1 {
			return &Refused{Where: l.program.Where(node), What: "an accessor through a structural property view (adamic/accessor-field-view)", Fix: "keep the nominal class type; an interface does not prove the virtual slot's layout"}
		}
	}
	var written *ast.Node
	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		_, compound := compoundAssignments[binary.OperatorToken.Kind]
		if compound || binary.OperatorToken.Kind == ast.KindEqualsToken {
			written = ast.SkipParentheses(binary.Left)
		}
	case ast.KindPostfixUnaryExpression:
		written = ast.SkipParentheses(node.AsPostfixUnaryExpression().Operand)
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
			written = ast.SkipParentheses(prefix.Operand)
		}
	}
	if written != nil && written.Kind == ast.KindPropertyAccessExpression && accessorSymbol(l.checker.GetSymbolAtLocation(written.Name())) {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		if outer.Parent != nil && outer.Parent.Kind != ast.KindExpressionStatement && outer.Parent.Kind != ast.KindForStatement {
			return l.notYet(node, "an accessor assignment or update used as a value (adamic/accessor-update-value)")
		}
	}
	var source *ast.Node
	switch node.Kind {
	case ast.KindSpreadAssignment:
		source = node.AsSpreadAssignment().Expression
	case ast.KindVariableDeclaration:
		if node.Name().Kind == ast.KindObjectBindingPattern {
			source = node.AsVariableDeclaration().Initializer
		}
	case ast.KindParameter:
		if node.Name().Kind == ast.KindObjectBindingPattern {
			source = node
		}
	case ast.KindElementAccessExpression:
		source = node.AsElementAccessExpression().Expression
	}
	if source != nil {
		for _, member := range l.definedMembers(l.checker.GetTypeAtLocation(source)) {
			for _, property := range l.checker.GetPropertiesOfType(member) {
				if accessorSymbol(property) {
					return l.notYet(node, "spread, destructuring or indexed access of an accessor-bearing value (adamic/accessor-property-operation)")
				}
			}
		}
	}
	return nil
}

// Accessor calls use the same typed call and target summaries as ordinary virtual
// methods. No analysis has a special field-load interpretation of this operation.
func (l *lowering) accessorCall(target *ast.Node, kind ast.Kind, object ir.Expression, value ir.Expression) (ir.Call, error) {
	symbol := l.checker.GetSymbolAtLocation(target.Name())
	declaration := accessorDeclaration(symbol, kind)
	if declaration == nil {
		reason := "a setter-only property read (adamic/setter-only-read)"
		if kind == ast.KindSetAccessor {
			reason = "a getter-only property write (adamic/getter-only-write)"
		}
		return ir.Call{}, &Refused{Where: l.program.Where(target), What: reason, Fix: "declare the missing accessor, or use an explicit method"}
	}
	receiver := target.AsPropertyAccessExpression().Expression
	if len(l.definedMembers(l.checker.GetTypeAtLocation(receiver))) > 1 {
		return ir.Call{}, l.notYet(target, "accessor dispatch through a union (adamic/accessor-union)")
	}
	lowered := l.instance
	if (ast.SkipParentheses(receiver).Kind != ast.KindThisKeyword && ast.SkipParentheses(receiver).Kind != ast.KindSuperKeyword) || lowered == nil {
		var err error
		lowered, err = l.instantiate(declaration.Parent, l.checker.GetTypeAtLocation(receiver), target)
		if err != nil {
			return ir.Call{}, err
		}
	}
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		if l.instance == nil || l.instance.base == nil {
			return ir.Call{}, l.notYet(target, "super outside a derived class")
		}
		lowered = l.instance.base
	}
	name := classMethodName(declaration)
	function, exists := lowered.methods[name]
	if !exists {
		return ir.Call{}, l.notYet(target, "an accessor without a virtual implementation")
	}
	arguments := []ir.Expression{object}
	if kind == ast.KindSetAccessor {
		parameter := l.result.Functions[function].Parameters[1]
		value = fit(value, l.result.Locals[parameter].Type)
		if value.Type() != l.result.Locals[parameter].Type {
			return ir.Call{}, l.notYet(target, "a setter argument with a different native representation")
		}
		arguments = append(arguments, value)
	}
	virtual := lowered.slots[name] + 1
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		virtual = 0
	}
	return ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns, Virtual: virtual}, nil
}

func (l *lowering) getAccessor(target *ast.Node) (ir.Expression, error) {
	object, err := l.accessorReceiver(target)
	if err != nil {
		return nil, err
	}
	call, err := l.accessorCall(target, ast.KindGetAccessor, object, nil)
	if err != nil {
		return nil, err
	}
	of, err := l.typeOf(target)
	if err != nil {
		return nil, err
	}
	value := ir.Expression(call)
	if call.Returns.IsMaybe() && of == call.Returns.Present() {
		value = ir.Unwrap{Value: call}
	}
	if value.Type() != of {
		return nil, l.notYet(target, "a narrowed accessor result with a different native representation (adamic/accessor-narrowing)")
	}
	return l.defined(target, value), nil
}

func (l *lowering) setAccessor(target, valueNode *ast.Node) ([]ir.Statement, error) {
	object, err := l.accessorReceiver(target)
	if err != nil {
		return nil, err
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	call, err := l.accessorCall(target, ast.KindSetAccessor, object, value)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Evaluate{Value: call}}, nil
}

func (l *lowering) accessorLocal(name string, value ir.Expression) (ir.Statement, ir.Expression) {
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: value.Type(), Function: l.functionIndex})
	return ir.Declare{Local: local, Value: value}, ir.Read{Local: local, Of: value.Type()}
}

func (l *lowering) updateAccessor(node, target *ast.Node, operator ast.Kind, valueNode *ast.Node) ([]ir.Statement, error) {
	receiver, err := l.accessorReceiver(target)
	if err != nil {
		return nil, err
	}
	// Even a variable must be held: the getter or the right operand can reassign it.
	held, object := l.accessorLocal("accessor_receiver", receiver)
	getter, err := l.accessorCall(target, ast.KindGetAccessor, object, nil)
	if err != nil {
		return nil, err
	}
	read, current := l.accessorLocal("accessor_value", getter)
	value := ir.Expression(ir.NumberConstant{Value: 1})
	if valueNode != nil {
		value, err = l.expression(valueNode)
		if err != nil {
			return nil, err
		}
	}
	if operator == ast.KindPlusToken && valueNode != nil {
		current, value = l.spelled(target, current), l.spelled(valueNode, value)
	}
	updated, err := l.combine(node, operator, current, value)
	if err != nil {
		return nil, err
	}
	setter, err := l.accessorCall(target, ast.KindSetAccessor, object, updated)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Block{Body: []ir.Statement{held, read, ir.Evaluate{Value: setter}}}}, nil
}

func (l *lowering) checkAccessorOverride(member *ast.Node, inherited *ast.Symbol, classType, baseType *checker.Type, checkABI bool) error {
	refuse := func(reason string) error {
		return &Refused{Where: l.program.Where(member), What: reason, Fix: "keep the inherited accessor descriptor and its read and write types"}
	}
	if !accessorSymbol(inherited) || (member.Kind != ast.KindGetAccessor && member.Kind != ast.KindSetAccessor) {
		return refuse("an accessor replaced by a field or method (adamic/accessor-member-kind)")
	}
	// JavaScript replaces the whole descriptor, rather than inheriting a missing half.
	for _, kind := range []ast.Kind{ast.KindGetAccessor, ast.KindSetAccessor} {
		old := accessorDeclaration(inherited, kind)
		var next *ast.Node
		for _, own := range member.Parent.Members() {
			if own.Kind == kind && own.Name().Text() == member.Name().Text() {
				next = own
			}
		}
		if (old == nil) != (next == nil) {
			return l.notYet(member, "an override changing the getter/setter descriptor halves (adamic/accessor-descriptor-override)")
		}
		if old == nil {
			continue
		}
		var source, target *checker.Type
		if kind == ast.KindGetAccessor {
			source = l.accessorType(next, classType)
			target = l.accessorType(old, baseType)
		} else {
			source = l.accessorType(old, baseType)
			target = l.accessorType(next, classType)
		}
		if !l.classAssignable(source, target) || l.widened(source, target, map[[2]*checker.Type]bool{}) != nil {
			return refuse("an unsound accessor override (adamic/accessor-override)")
		}
		a, knownA := l.representation(source)
		b, knownB := l.representation(target)
		if checkABI && (!knownA || !knownB || a != b) {
			return l.notYet(member, "an accessor override changing native representation (adamic/accessor-override-representation)")
		}
	}
	return nil
}

// A declaration's signature still names its own type parameters, including when
// the inherited property symbol comes from an instantiated base. Substitute via
// that declaration's nominal view, not the descendant's argument positions.
func (l *lowering) accessorType(declaration *ast.Node, receiver *checker.Type) *checker.Type {
	signature := l.checker.GetSignatureFromDeclaration(declaration)
	var proven *checker.Type
	if declaration.Kind == ast.KindGetAccessor {
		proven = l.checker.GetReturnTypeOfSignature(signature)
	} else {
		proven = l.checker.GetTypeOfSymbol(signature.Parameters()[0])
	}
	view := l.classView(receiver, declaration.Parent)
	if view == nil {
		return proven
	}
	mapper := l.typeMapperOf(declaration.Parent, view)
	if mapper == nil {
		return proven
	}
	return instantiateType(l.checker, proven, mapper)
}

// super selects a descriptor on the base but does not change its receiver.
func (l *lowering) accessorReceiver(target *ast.Node) (ir.Expression, error) {
	receiver := target.AsPropertyAccessExpression().Expression
	if ast.SkipParentheses(receiver).Kind != ast.KindSuperKeyword {
		return l.expression(receiver)
	}
	if l.instance == nil || l.instance.base == nil || l.this < 0 {
		return nil, l.notYet(target, "super outside a derived class")
	}
	if err := l.useOfThis(receiver); err != nil {
		return nil, err
	}
	return ir.Read{Local: l.this, Of: ir.Object}, nil
}
