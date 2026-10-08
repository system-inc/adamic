package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// The symbol slot has a reserved spelling only in IR. Source string fields may not use it.
const iteratorSlot = "__adamic_symbol_iterator"

func (l *lowering) symbolIterator(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "iterator" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Symbol")
}

func (l *lowering) methodName(node *ast.Node) (string, bool) {
	name := node.Name()
	if name == nil {
		return "", false
	}
	if ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral {
		return name.Text(), name.Text() != iteratorSlot
	}
	if name.Kind == ast.KindComputedPropertyName && l.symbolIterator(name.AsComputedPropertyName().Expression) {
		return iteratorSlot, true
	}
	return "", false
}

func (l *lowering) iteratorMember(proven *checker.Type) *ast.Symbol {
	for _, member := range l.checker.GetPropertiesOfType(l.concrete(proven)) {
		for _, root := range l.checker.GetRootSymbols(member) {
			for _, declaration := range root.Declarations {
				if name, known := l.methodName(declaration); known && name == iteratorSlot && !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
					return member
				}
			}
		}
	}
	return nil
}

// Literal methods receive this as an argument, rather than capturing their owner. Their ordinary
// lexical captures still use the existing cycle analysis and reference-counted closure environment.
func (l *lowering) objectMethod(node *ast.Node) (ir.Expression, error) {
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "object_method", Closure: true, Receiver: true})
	this := l.iterationLocal("this", ir.Object, index)
	l.noteLocal(this, l.checker.GetTypeAtLocation(node.Parent), node)
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.checker.GetTypeAtLocation(node), function: index, node: node})
	l.closures = append(l.closures, index)
	err := l.lowerFunction(index, node, this)
	l.closures = l.closures[:len(l.closures)-1]
	if err != nil {
		return nil, err
	}
	return ir.MakeClosure{Function: index}, nil
}

func (l *lowering) iterationLocal(name string, of ir.Type, function int) int {
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: function})
	return local
}

// memberReceiver rejects erased method origins: a class method, a literal method and an arrow
// field have different calling conventions, which a bare structural interface does not preserve.
func (l *lowering) memberReceiver(where *ast.Node, member *ast.Symbol) (bool, error) {
	if member == nil || member.Flags&ast.SymbolFlagsOptional != 0 {
		return false, l.notYet(where, "an optional iterator method (runtime method presence is not represented)")
	}
	convention := -1
	for _, root := range l.checker.GetRootSymbols(member) {
		for _, declaration := range root.Declarations {
			current := 0
			switch declaration.Kind {
			case ast.KindMethodDeclaration:
				current = 1
			case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindPropertyDeclaration:
				if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAmbient) {
					return false, l.notYet(where, "a declared iterator method with no runtime field")
				}
			default:
				return false, l.notYet(where, "an iterator method whose concrete origin is erased by a structural signature")
			}
			if convention >= 0 && convention != current {
				return false, l.notYet(where, "iterator methods with mixed receiver conventions")
			}
			convention = current
		}
	}
	if convention < 0 {
		return false, l.notYet(where, "an iterator method without a concrete declaration")
	}
	return convention == 1, nil
}

func (l *lowering) memberFunction(where *ast.Node, proven *checker.Type, value ir.Expression, member *ast.Symbol) (ir.Expression, bool, int, error) {
	receiver, err := l.memberReceiver(where, member)
	if err != nil {
		return nil, false, -1, err
	}
	for _, root := range l.checker.GetRootSymbols(member) {
		for _, declaration := range root.Declarations {
			if declaration.Kind == ast.KindMethodDeclaration && declaration.Parent.Kind == ast.KindClassDeclaration {
				if !isClassInstance(l.concrete(proven)) {
					return nil, false, -1, l.notYet(where, "a class iterator method through a structural object view")
				}
				class := l.classes[l.symbol(declaration.Parent.Name())]
				if class == nil {
					return nil, false, -1, l.notYet(where, "an iterator class not declared in this program")
				}
				instance, err := l.instantiate(class, proven, where)
				if err != nil {
					return nil, false, -1, err
				}
				key := l.methodKey(declaration, instance.class)
				call := ir.Call{Function: instance.methods[key], Virtual: instance.slots[key] + 1}
				return call, true, call.Function, nil
			}
		}
	}
	name := member.Name
	for _, root := range l.checker.GetRootSymbols(member) {
		for _, declaration := range root.Declarations {
			if key, known := l.methodName(declaration); known {
				name = key
			}
		}
	}
	return ir.Property{Object: value, Name: name, Of: ir.Closure}, receiver, -1, nil
}

func invokeMember(function ir.Expression, receiver bool, direct int, value ir.Expression, arguments []ir.Expression, returns ir.Type) ir.Expression {
	if receiver {
		arguments = append([]ir.Expression{value}, arguments...)
	}
	if direct >= 0 {
		call := ir.Call{Function: direct, Arguments: arguments, Returns: returns}
		if method, ok := function.(ir.Call); ok {
			call.Virtual = method.Virtual
		}
		return call
	}
	return ir.CallClosure{Closure: function, Arguments: arguments, Returns: returns}
}

func (l *lowering) memberResult(where *ast.Node, member *ast.Symbol) (*checker.Type, error) {
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(member), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].Parameters()) != 0 {
		return nil, l.notYet(where, "an iterator protocol method with arguments or overloads")
	}
	result := l.concrete(l.checker.GetReturnTypeOfSignature(signatures[0]))
	if of, known := l.representation(result); !known || of != ir.Object || l.includesUndefined(result) {
		return nil, l.notYet(where, "an iterator protocol method that does not return a represented object")
	}
	return result, nil
}

type iterationPlan struct {
	source, iterator, step *checker.Type
	entry, next, close     *ast.Symbol
	element                ir.Type
}

func (l *lowering) planIteration(where *ast.Node) (*iterationPlan, error) {
	source := l.concrete(l.checker.GetTypeAtLocation(where))
	entry := l.iteratorMember(source)
	if entry == nil {
		return nil, nil
	}
	if of, known := l.representation(source); !known || of != ir.Object || l.includesUndefined(source) {
		return nil, l.notYet(where, "a custom iterable without a present object representation")
	}
	if _, err := l.memberReceiver(where, entry); err != nil {
		return nil, err
	}
	iterator, err := l.memberResult(where, entry)
	if err != nil {
		return nil, err
	}
	receiverFactory := l.iterationFactoryReturnsReceiver(entry)
	if receiverFactory && isClassInstance(source) {
		// Returning this preserves the source's subtype, including a newly introduced return.
		iterator = source
	}
	next := l.checker.GetPropertyOfType(iterator, "next")
	if _, err := l.memberReceiver(where, next); err != nil {
		return nil, err
	}
	step, err := l.memberResult(where, next)
	if err != nil {
		return nil, err
	}
	done := l.checker.GetPropertyOfType(step, "done")
	if done == nil || done.Flags&ast.SymbolFlagsOptional != 0 {
		return nil, l.notYet(where, "an iterator result without a required boolean done field")
	}
	if of, known := l.representation(l.checker.GetTypeOfSymbol(done)); !known || of != ir.Boolean {
		return nil, l.notYet(where, "an iterator result whose done is not a boolean")
	}
	value := l.checker.GetPropertyOfType(step, "value")
	if value == nil || value.Flags&ast.SymbolFlagsOptional != 0 {
		return nil, l.notYet(where, "an iterator result without a represented value field")
	}
	element, known := l.representation(l.checker.GetTypeOfSymbol(value))
	if !known || slotless(element) || element == ir.Weak {
		return nil, l.notYet(where, "an iterator result whose value has no single-slot representation")
	}
	close := l.checker.GetPropertyOfType(iterator, "return")
	if close != nil {
		if _, err := l.memberReceiver(where, close); err != nil {
			return nil, err
		}
		if _, err := l.memberResult(where, close); err != nil {
			return nil, err
		}
	}
	// A view which omits return must not hide a runtime close method. Method signatures themselves
	// are refused above, but a concrete literal can also be viewed as another concrete literal type.
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return nil, err
	}
	sourceKnown := l.iterationOrigin(where, []*ast.Symbol{entry}, nil, 0)
	iteratorKnown := sourceKnown && l.iterationFactoryKnown(entry, next, close)
	if receiverFactory {
		presence := close != nil
		members := []*ast.Symbol{next}
		if close != nil {
			members = append(members, close)
		}
		iteratorKnown = sourceKnown && l.iterationOrigin(where, members, &presence, 0)
	}
	if isClassInstance(source) && len(l.checker.GetTypeArguments(source)) > 0 && !l.knownIterationClass(where, source, 0) {
		return nil, l.notYet(where, "a generic iterable view without proven native type arguments")
	}
	if isClassInstance(iterator) && len(l.checker.GetTypeArguments(iterator)) > 0 && (!sourceKnown || !l.genericIteratorFactoryKnown(entry, source, iterator)) {
		return nil, l.notYet(where, "a generic iterator view without proven native type arguments")
	}

	var hazard error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if hazard != nil {
			return true
		}
		if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindNewExpression {
			shape := l.checker.GetTypeAtLocation(node)
			if l.iterationShapeFits(shape, step) {
				for _, key := range []string{"done", "value"} {
					if field := l.checker.GetPropertyOfType(shape, key); field != nil {
						for _, declaration := range field.Declarations {
							if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAmbient) {
								hazard = l.notYet(where, "an iterator result with declared fields absent at runtime")
								return true
							}
						}
					}
				}
			}
			if !sourceKnown && l.iterationShapeFits(shape, source) {
				actual := l.iteratorMember(shape)
				if actual == nil || !l.iterationConventionFits(actual, entry, shape, source) {
					hazard = l.notYet(where, "an iterable view that erases its method receiver convention")
					return true
				}
			}
			if !iteratorKnown && l.iterationShapeFits(shape, iterator) {
				actual := l.checker.GetPropertyOfType(shape, "next")
				if !l.iterationConventionFits(actual, next, shape, iterator) {
					hazard = l.notYet(where, "an iterator view that erases its method receiver convention")
					return true
				}
				actualClose := l.checker.GetPropertyOfType(shape, "return")
				if close != nil && !l.iterationConventionFits(actualClose, close, shape, iterator) {
					hazard = l.notYet(where, "an iterator view that erases its return receiver convention")
					return true
				}
			}
			if node.Kind == ast.KindObjectLiteralExpression {
				for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
					if property.Kind == ast.KindSpreadAssignment && (l.iterationShapeFits(shape, iterator) || l.iterationShapeFits(shape, source)) {
						hazard = l.notYet(where, "an iterator built by object spread (own method presence is not proved)")
						return true
					}
				}
			}
			if !iteratorKnown && l.iterationShapeFits(shape, iterator) && (l.checker.GetPropertyOfType(shape, "return") != nil) != (close != nil) {
				hazard = l.notYet(where, "an iterator view that can hide a return method; retain the concrete subclass type, or return a separate iterator object")
				return true
			}
		}
		if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			target := ast.SkipParentheses(node.AsBinaryExpression().Left)
			if target.Kind == ast.KindPropertyAccessExpression {
				access := target.AsPropertyAccessExpression()
				key := access.Name().Text()
				targetType := l.checker.GetTypeAtLocation(access.Expression)
				if (key == "next" || key == "return") && (l.checker.IsTypeAssignableTo(iterator, targetType) || l.checker.IsTypeAssignableTo(targetType, iterator)) {
					hazard = l.notYet(where, "replacing an iterator protocol method at runtime")
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if hazard != nil {
		return nil, hazard
	}
	return &iterationPlan{source: source, iterator: iterator, step: step, entry: entry, next: next, close: close, element: element}, nil
}

type iterationState struct {
	plan     *iterationPlan
	iterator ir.Expression
	next     ir.Expression
	receiver bool
	direct   int
	active   int
}

func (l *lowering) startIteration(where *ast.Node, plan *iterationPlan, source ir.Expression) ([]ir.Statement, *iterationState, error) {
	function, receiver, direct, err := l.memberFunction(where, plan.source, source, plan.entry)
	if err != nil {
		return nil, nil, err
	}
	iterator := l.iterationLocal("iterator", ir.Object, l.functionIndex)
	active := l.iterationLocal("iterator_close_needed", ir.Boolean, l.functionIndex)
	read := ir.Read{Local: iterator, Of: ir.Object}
	statements := []ir.Statement{ir.Declare{Local: iterator, Value: invokeMember(function, receiver, direct, source, nil, ir.Object)}, ir.Declare{Local: active, Value: ir.BooleanConstant{Value: false}}}
	next, receiver, direct, err := l.memberFunction(where, plan.iterator, read, plan.next)
	if err != nil {
		return nil, nil, err
	}
	if direct < 0 {
		cached := l.iterationLocal("iterator_next", ir.Closure, l.functionIndex)
		statements = append(statements, ir.Declare{Local: cached, Value: next})
		next = ir.Read{Local: cached, Of: ir.Closure}
	}
	return statements, &iterationState{plan: plan, iterator: read, next: next, receiver: receiver, direct: direct, active: active}, nil
}

func (l *lowering) iterationStep(state *iterationState) ([]ir.Statement, ir.Expression) {
	step := l.iterationLocal("iterator_result", ir.Object, l.functionIndex)
	read := ir.Read{Local: step, Of: ir.Object}
	statements := []ir.Statement{
		ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}},
		ir.Declare{Local: step, Value: invokeMember(state.next, state.receiver, state.direct, state.iterator, nil, ir.Object)},
		ir.If{Condition: ir.Property{Object: read, Name: "done", Of: ir.Boolean}, Then: []ir.Statement{ir.Break{}}},
	}
	value := ir.Property{Object: read, Name: "value", Of: state.plan.element}
	return statements, value
}

// IteratorClose preserves an incoming throw even if return throws. Break and return instead let
// that close failure replace their completion. The active flag excludes exhaustion and next errors.
func (l *lowering) closeIteration(where *ast.Node, state *iterationState, body []ir.Statement) ([]ir.Statement, error) {
	if state.plan.close == nil {
		return body, nil
	}
	function, receiver, direct, err := l.memberFunction(where, state.plan.iterator, state.iterator, state.plan.close)
	if err != nil {
		return nil, err
	}
	close := []ir.Statement{ir.Evaluate{Value: invokeMember(function, receiver, direct, state.iterator, nil, ir.Object)}}
	thrown := l.iterationLocal("iterator_pending_throw", ir.Boolean, l.functionIndex)
	caught := l.iterationLocal("iterator_error", ir.Object, l.functionIndex)
	// Distinct branches need distinct statement identities for flow analysis and tracing.
	cleanup := ir.If{Condition: ir.Read{Local: state.active, Of: ir.Boolean}, Then: []ir.Statement{
		ir.If{Condition: ir.Read{Local: thrown, Of: ir.Boolean}, Then: []ir.Statement{ir.Try{Body: close, HasCatch: true, CatchLocal: -1}}, Else: append([]ir.Statement{}, close...)},
	}}
	guarded := ir.Try{Body: body, HasCatch: true, CatchLocal: caught, Catch: []ir.Statement{ir.Assign{Local: thrown, Value: ir.BooleanConstant{Value: true}}, ir.Throw{Value: ir.Read{Local: caught, Of: ir.Object}}}, HasFinally: true, Finally: []ir.Statement{cleanup}}
	l.tries = append(l.tries, tryRecord{node: where, body: body})
	return []ir.Statement{ir.Declare{Local: thrown, Value: ir.BooleanConstant{Value: false}}, guarded}, nil
}

func (l *lowering) forOfUser(node *ast.Node, plan *iterationPlan, name *ast.Node) ([]ir.Statement, error) {
	source, err := l.expression(node.AsForInOrOfStatement().Expression)
	if err != nil {
		return nil, err
	}
	held := l.iterationLocal("iterable", ir.Object, l.functionIndex)
	read := ir.Read{Local: held, Of: ir.Object}
	setup, state, err := l.startIteration(node, plan, read)
	if err != nil {
		return nil, err
	}
	step, value := l.iterationStep(state)
	if ast.IsIdentifier(name) {
		local, err := l.declareLocal(name)
		if err != nil {
			return nil, err
		}
		if l.result.Locals[local].Type != plan.element {
			return nil, l.notYet(name, "an iteration variable with a different value representation")
		}
		step = append(step, ir.Declare{Local: local, Value: value})
	} else {
		valueType := l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(plan.step, "value"))
		heldValue := l.iterationLocal("iteration_value", plan.element, l.functionIndex)
		bound, err := l.destructureFrom(name, valueType, plan.element, heldValue)
		if err != nil {
			return nil, err
		}
		step = append(step, ir.Declare{Local: heldValue, Value: value})
		step = append(step, bound...)
	}
	body, err := l.statement(node.AsForInOrOfStatement().Statement)
	if err != nil {
		return nil, err
	}
	step = append(step, ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: true}})
	step = append(step, body...)
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: step, Update: []ir.Statement{ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}}}
	guarded, err := l.closeIteration(node, state, []ir.Statement{loop})
	if err != nil {
		return nil, err
	}
	statements := append([]ir.Statement{ir.Declare{Local: held, Value: source}}, setup...)
	return []ir.Statement{ir.Block{Body: append(statements, guarded...)}}, nil
}

func (l *lowering) userMethodCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	var receiverNode *ast.Node
	var member *ast.Symbol
	switch callee.Kind {
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if !l.symbolIterator(access.ArgumentExpression) {
			return nil, false, nil
		}
		receiverNode = access.Expression
		member = l.iteratorMember(l.checker.GetTypeAtLocation(receiverNode))
		if member == nil {
			return nil, true, l.notYet(node, "an explicit Symbol.iterator call on a built-in value")
		}
	case ast.KindPropertyAccessExpression:
		member = l.checker.GetSymbolAtLocation(callee)
		if member == nil || !literalMethod(member) {
			return nil, false, nil
		}
		value, err := l.callClosure(node)
		return value, true, err
	default:
		return nil, false, nil
	}
	value, err := l.expression(receiverNode)
	if err != nil {
		return nil, true, err
	}
	arguments := []ir.Expression{value}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, lowered)
	}
	result, _ := l.representation(l.checker.GetTypeAtLocation(node))
	if result == 0 || slotless(result) {
		return nil, true, l.notYet(node, "a literal method call with an unrepresented result")
	}
	function := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "invoke_method", Returns: result})
	parameters := []int{}
	for _, argument := range arguments {
		parameters = append(parameters, l.iterationLocal("argument", argument.Type(), function))
	}
	object := ir.Read{Local: parameters[0], Of: ir.Object}
	method, hasReceiver, direct, err := l.memberFunction(node, l.checker.GetTypeAtLocation(receiverNode), object, member)
	if err != nil {
		return nil, true, err
	}
	passed := []ir.Expression{}
	for index, parameter := range parameters[1:] {
		passed = append(passed, ir.Read{Local: parameter, Of: arguments[index+1].Type()})
	}
	l.result.Functions[function] = ir.Function{Name: "invoke_method", Returns: result, Parameters: parameters, Body: []ir.Statement{ir.Return{Value: invokeMember(method, hasReceiver, direct, object, passed, result)}}}
	return ir.Call{Function: function, Arguments: arguments, Returns: result}, true, nil
}
