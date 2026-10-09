package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) builtinArguments(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindIdentifier || node.Text() != "arguments" {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	return symbol != nil && len(symbol.Declarations) == 0
}

func (l *lowering) isArgumentsLength(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := node.AsPropertyAccessExpression()
	return access.Name().Text() == "length" && access.QuestionDotToken == nil && l.builtinArguments(access.Expression)
}

func (l *lowering) argumentsRefusal(node *ast.Node) error {
	for owner := node.Parent; owner != nil; owner = owner.Parent {
		if ast.IsFunctionLike(owner) {
			if owner.Kind == ast.KindArrowFunction {
				return &Refused{Where: l.program.Where(node), What: "arguments inside an arrow function", Fix: "read arguments.length in the enclosing non-arrow function and capture that number"}
			}
			break
		}
	}
	value := node
	for value.Parent != nil && value.Parent.Kind == ast.KindParenthesizedExpression {
		value = value.Parent
	}
	if value.Parent != nil && l.isArgumentsLength(value.Parent) {
		length := value.Parent
		for length.Parent != nil && length.Parent.Kind == ast.KindParenthesizedExpression {
			length = length.Parent
		}
		if !ast.IsAssignmentTarget(length) {
			return nil
		}
		return &Refused{Where: l.program.Where(node), What: "writing arguments.length", Fix: "keep the count read-only; write an explicit local number instead"}
	}
	return &Refused{Where: l.program.Where(node), What: "arguments other than a read of arguments.length", Fix: "name the parameters, or take a rest parameter; only arguments.length may be read"}
}

func (l *lowering) readArgumentsCount() ir.Expression {
	if l.function == nil {
		panic("lower: arguments.length outside a function")
	}
	l.function.ReadsArguments = true
	if l.function.ArgumentsCount == 0 {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument_count", Type: ir.Number, Function: l.functionIndex})
		l.function.ArgumentsCount = local + 1
	}
	return ir.Read{Local: l.function.ArgumentsCount - 1, Of: ir.Number}
}

// callArguments keeps spreads at their evaluation point. Backends expand them once,
// before evaluating the next argument, so a later mutation cannot change the count.
func (l *lowering) callArguments(nodes []*ast.Node) ([]ir.Expression, []bool, error) {
	arguments := []ir.Expression{}
	spread := make([]bool, len(nodes))
	anySpread := false
	for index, node := range nodes {
		value := node
		if node.Kind == ast.KindSpreadElement {
			value = node.AsSpreadElement().Expression
			proven := l.concrete(l.checker.GetTypeAtLocation(value))
			if checker.IsTupleType(proven) && ast.SkipParentheses(value).Kind == ast.KindArrayLiteralExpression {
				elements := l.checker.GetTypeArguments(proven)
				element := ir.Number
				for position, provenElement := range elements {
					of, known := l.representation(provenElement)
					if !known || slotless(of) || position > 0 && of != element {
						return nil, nil, l.notYet(node, "a call spreading a tuple with differently represented elements")
					}
					element = of
				}
				values, nested, err := l.callArguments(ast.SkipParentheses(value).AsArrayLiteralExpression().Elements.Nodes)
				if err != nil {
					return nil, nil, err
				}
				arguments = append(arguments, ir.ArrayLiteral{Element: element, Elements: values, Spread: nested})
				spread[index], anySpread = true, true
				continue
			}
			if !l.checker.IsArrayType(proven) {
				return nil, nil, l.notYet(node, "a call spreading something other than an array")
			}
			elements := l.checker.GetTypeArguments(proven)
			if len(elements) != 1 {
				return nil, nil, l.notYet(node, "a call spreading an array without one element type")
			}
			element, known := l.representation(elements[0])
			if !known || slotless(element) {
				return nil, nil, l.notYet(node, "a call spreading elements that cannot be packed")
			}
			spread[index], anySpread = true, true
		}
		lowered, err := l.expression(value)
		if err != nil {
			return nil, nil, err
		}
		arguments = append(arguments, lowered)
	}
	if !anySpread {
		spread = nil
	}
	return arguments, spread, nil
}

// resolveCountTypes uses the same assignability relation and closed-world function
// records as the cycle finder. A parameter or mutable function value is bounded
// by every implementation that can be seen through its function type.
func (l *lowering) resolveCountTypes(modules []*ast.SourceFile) {
	types := map[int]*checker.Type{}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				// Console writes are intrinsics, not callable values. Node's console
				// declaration can merge with the prelude in different source orders;
				// its broad signature must not group unrelated runtime callbacks.
				if _, err := l.consoleStream(callee); err == nil {
					return node.ForEachChild(visit)
				}
				proven := l.concrete(l.checker.GetTypeAtLocation(callee))
				types[int(proven.Id())] = proven
				for _, argument := range node.AsCallExpression().Arguments.Nodes {
					callback := l.concrete(l.checker.GetTypeAtLocation(argument))
					if len(l.checker.GetSignaturesOfType(callback, checker.SignatureKindCall)) > 0 {
						types[int(callback.Id())] = callback
					}
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, proven := range l.localTypes {
		if proven != nil {
			proven = l.concrete(proven)
			types[int(proven.Id())] = proven
		}
	}
	l.resolveStructuralMethodThunks(modules)
	l.result.FunctionTypeTargets = map[int][]int{}
	for id, proven := range types {
		l.result.FunctionTypeTargets[id] = []int{}
		for _, record := range l.closureRecords {
			if l.checker.IsTypeAssignableTo(record.proven, proven) {
				l.result.FunctionTypeTargets[id] = append(l.result.FunctionTypeTargets[id], record.function)
			}
		}
	}
	// Reader facts must remain separate from padding a virtual C signature.
	for changed := true; changed; {
		changed = false
		for index := range l.result.Functions {
			function := &l.result.Functions[index]
			if function.ForwardsArguments != 0 && !function.ReadsArguments && l.result.Functions[function.ForwardsArguments-1].ReadsArguments {
				function.ReadsArguments = true
				changed = true
			}
		}
	}
	l.result.PrepareArgumentSlots()
	// Every virtual implementation must have the same native signature, including
	// an implementation that does not itself observe the count.
	changed := true
	for changed {
		changed = false
		for index := range l.result.Functions {
			function := &l.result.Functions[index]
			if function.ForwardsArguments == 0 || function.ArgumentsCount != 0 || l.result.Functions[function.ForwardsArguments-1].ArgumentsCount == 0 {
				continue
			}
			local := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument_count", Type: ir.Number, Function: index})
			function.ArgumentsCount = local + 1
			changed = true
		}
		for signature, implementations := range l.result.MethodTargets {
			targets := append([]int{signature}, implementations...)
			reads := false
			for _, target := range targets {
				reads = reads || l.result.Functions[target].ArgumentsCount != 0
			}
			if !reads {
				continue
			}
			for _, target := range targets {
				if l.result.Functions[target].ArgumentsCount == 0 {
					local := len(l.result.Locals)
					l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument_count", Type: ir.Number, Function: target})
					l.result.Functions[target].ArgumentsCount = local + 1
					changed = true
				}
			}
		}
	}
}

func (l *lowering) fitCallArguments(function int, arguments []ir.Expression, spread []bool) {
	// Fixed arguments are fitted by the native call boundary, after ownership
	// planning. Fitting them here creates extra owned Weak and Union values.
	declared := l.result.Functions[function]
	position := 0
	expanded := false
	for index := range arguments {
		if len(spread) > index && spread[index] {
			expanded = true
			continue
		}
		if declared.RestElement != 0 && (expanded || position >= len(declared.Parameters)-1) {
			arguments[index] = fit(arguments[index], declared.RestElement)
		}
		position++
	}
}

// Optional method admission must account for the receiver, not just an assignable
// function signature. An own closure facade does not make every class with a
// same-signature method reachable. This does not narrow the count candidate set.
func (l *lowering) resolveStructuralMethodThunks(modules []*ast.SourceFile) {
	l.result.StructuralMethodThunks = map[int]bool{}
	finder := cycleFinder{l: l}
	// A derived receiver can expose an inherited implementation through a wider
	// structural view. Include every instantiation using that implementation.
	actualReceivers := map[int][]*checker.Type{}
	for _, instance := range l.instances {
		for _, function := range instance.methods {
			for _, local := range instance.thisLocals {
				actualReceivers[function] = append(actualReceivers[function], l.localTypes[local])
				actualReceivers[function] = append(actualReceivers[function], l.localAlso[local]...)
			}
		}
	}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee.Kind == ast.KindElementAccessExpression {
					// Computed callees do not retain a receiver-name proof here.
					// Preserve every method thunk rather than guessing a target.
					for index, function := range l.result.Functions {
						if function.Receiver {
							l.result.StructuralMethodThunks[index] = true
						}
					}
				}
				if callee.Kind == ast.KindPropertyAccessExpression {
					access := callee.AsPropertyAccessExpression()
					receiver := l.concrete(l.checker.GetTypeAtLocation(access.Expression))
					method := l.checker.GetSymbolAtLocation(callee)
					nativeMethod := method != nil && len(method.Declarations) > 0 && method.Declarations[0].Kind == ast.KindMethodDeclaration && access.QuestionDotToken == nil
					if !nativeMethod {
						for index, function := range l.result.Functions {
							if !function.Receiver || function.MethodName != access.Name().Text() || len(function.Parameters) == 0 {
								continue
							}
							local := function.Parameters[0]
							actuals := append([]*checker.Type{l.localTypes[local]}, l.localAlso[local]...)
							actuals = append(actuals, actualReceivers[index]...)
							for _, actual := range actuals {
								if actual == nil || finder.template(receiver) || finder.template(actual) || l.checker.IsTypeAssignableTo(actual, receiver) {
									l.result.StructuralMethodThunks[index] = true
								}
							}
						}
					}
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
}
