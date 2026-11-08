package lower

import (
	"maps"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (lowering *lowering) builtinArguments(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindIdentifier || node.Text() != "arguments" {
		return false
	}
	symbol := lowering.checker.GetSymbolAtLocation(node)
	return symbol != nil && len(symbol.Declarations) == 0
}

func (lowering *lowering) isArgumentsLength(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := node.AsPropertyAccessExpression()
	return access.Name().Text() == "length" &&
		access.QuestionDotToken == nil &&
		lowering.builtinArguments(access.Expression)
}

// argumentsObjectRefusal checks a use of JavaScript's arguments object.
func (lowering *lowering) argumentsObjectRefusal(node *ast.Node) error {
	if node.Kind == ast.KindIdentifier && node.Text() == "arguments" {
		// JavaScript's arguments object, not a variable the program named arguments.
		if symbol := lowering.checker.GetSymbolAtLocation(node); symbol != nil && len(symbol.Declarations) == 0 {
			if argumentsRefusalError := lowering.argumentsRefusal(node); argumentsRefusalError != nil {
				return argumentsRefusalError
			}
		}
	}
	return nil
}

func init() {
	refusalRules.register(2600, "argumentsObjectRefusal", (*lowering).argumentsObjectRefusal)
}

func (lowering *lowering) argumentsRefusal(node *ast.Node) error {
	for owner := node.Parent; owner != nil; owner = owner.Parent {
		if ast.IsFunctionLike(owner) {
			if owner.Kind == ast.KindArrowFunction {
				return &Refused{
					Where: lowering.program.Where(node),
					What:  "arguments inside an arrow function",
					Fix:   "read arguments.length in the enclosing non-arrow function and capture that number",
				}
			}
			break
		}
	}
	value := node
	for value.Parent != nil && value.Parent.Kind == ast.KindParenthesizedExpression {
		value = value.Parent
	}
	if value.Parent != nil && lowering.isArgumentsLength(value.Parent) {
		length := value.Parent
		for length.Parent != nil && length.Parent.Kind == ast.KindParenthesizedExpression {
			length = length.Parent
		}
		if !ast.IsAssignmentTarget(length) {
			return nil
		}
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "writing arguments.length",
			Fix:   "keep the count read-only; write an explicit local number instead",
		}
	}
	return &Refused{
		Where: lowering.program.Where(node),
		What:  "arguments other than a read of arguments.length",
		Fix:   "name the parameters, or take a rest parameter; only arguments.length may be read",
	}
}

func (lowering *lowering) readArgumentsCount() ir.Expression {
	if lowering.function == nil {
		panic("lower: arguments.length outside a function")
	}
	lowering.function.ReadsArguments = true
	if lowering.function.ArgumentsCount == 0 {
		local := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "argument_count", Type: ir.Number, Function: lowering.functionIndex},
		)
		lowering.function.ArgumentsCount = local + 1
	}
	return ir.Read{Local: lowering.function.ArgumentsCount - 1, Of: ir.Number}
}

// callArguments keeps spreads at their evaluation point. Backends expand them once,
// before evaluating the next argument, so a later mutation cannot change the count.
func (lowering *lowering) callArguments(nodes []*ast.Node) ([]ir.Expression, []bool, error) {
	arguments := []ir.Expression{}
	spread := make([]bool, len(nodes))
	anySpread := false
	for index, node := range nodes {
		value := node
		if node.Kind == ast.KindSpreadElement {
			value = node.AsSpreadElement().Expression
			proven := lowering.concrete(lowering.checker.GetTypeAtLocation(value))
			if checker.IsTupleType(proven) && ast.SkipParentheses(value).Kind == ast.KindArrayLiteralExpression {
				elements := lowering.checker.GetTypeArguments(proven)
				element := ir.Number
				for position, provenElement := range elements {
					of, known := lowering.representation(provenElement)
					if !known || slotless(of) || position > 0 && of != element {
						return nil, nil, lowering.notYet(
							node,
							"a call spreading a tuple with differently represented elements",
						)
					}
					element = of
				}
				values, nested, callArgumentsError := lowering.callArguments(
					ast.SkipParentheses(value).AsArrayLiteralExpression().Elements.Nodes,
				)
				if callArgumentsError != nil {
					return nil, nil, callArgumentsError
				}
				arguments = append(arguments, ir.ArrayLiteral{Element: element, Elements: values, Spread: nested})
				spread[index], anySpread = true, true
				continue
			}
			if !lowering.isArrayType(proven) {
				return nil, nil, lowering.notYet(node, "a call spreading something other than an array")
			}
			elements := lowering.checker.GetTypeArguments(proven)
			if len(elements) != 1 {
				return nil, nil, lowering.notYet(node, "a call spreading an array without one element type")
			}
			element, known := lowering.representation(elements[0])
			if !known || slotless(element) {
				return nil, nil, lowering.notYet(node, "a call spreading elements that cannot be packed")
			}
			spread[index], anySpread = true, true
		}
		lowered, expressionError := lowering.expression(value)
		if expressionError != nil {
			return nil, nil, expressionError
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
func (lowering *lowering) resolveCountTypes(modules []*ast.SourceFile) {
	types := map[int]*checker.Type{}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression {
				proven := lowering.concrete(lowering.checker.GetTypeAtLocation(node.AsCallExpression().Expression))
				types[int(proven.Id())] = proven
				for _, argument := range node.AsCallExpression().Arguments.Nodes {
					callback := lowering.concrete(lowering.checker.GetTypeAtLocation(argument))
					if len(lowering.checker.GetSignaturesOfType(callback, checker.SignatureKindCall)) > 0 {
						types[int(callback.Id())] = callback
					}
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, proven := range lowering.localTypes {
		if proven != nil {
			proven = lowering.concrete(proven)
			types[int(proven.Id())] = proven
		}
	}
	lowering.resolveStructuralMethodThunks(modules)
	lowering.result.FunctionTypeTargets = map[int][]int{}
	for id, proven := range types {
		lowering.cycleTypeId(proven) // Retain transient count-target types for differential tests.
		lowering.result.FunctionTypeTargets[id] = []int{}
		for _, record := range lowering.closureRecords {
			if lowering.checker.IsTypeAssignableTo(record.proven, proven) {
				lowering.result.FunctionTypeTargets[id] = append(
					lowering.result.FunctionTypeTargets[id],
					record.function,
				)
			}
		}
	}
	// Reader facts must remain separate from padding a virtual C signature.
	for changed := true; changed; {
		changed = false
		for index := range lowering.result.Functions {
			function := &lowering.result.Functions[index]
			if function.ForwardsArguments != 0 &&
				!function.ReadsArguments &&
				lowering.result.Functions[function.ForwardsArguments-1].ReadsArguments {
				function.ReadsArguments = true
				changed = true
			}
		}
	}
	lowering.result.PrepareArgumentSlots()
	// Every virtual implementation must have the same native signature, including
	// an implementation that does not itself observe the count.
	changed := true
	for changed {
		changed = false
		for index := range lowering.result.Functions {
			function := &lowering.result.Functions[index]
			if function.ForwardsArguments == 0 ||
				function.ArgumentsCount != 0 ||
				lowering.result.Functions[function.ForwardsArguments-1].ArgumentsCount == 0 {
				continue
			}
			local := len(lowering.result.Locals)
			lowering.result.Locals = append(
				lowering.result.Locals,
				ir.Local{Name: "argument_count", Type: ir.Number, Function: index},
			)
			function.ArgumentsCount = local + 1
			changed = true
		}
		// By signature, never in the map's order: each local's index is in its C name.
		for _, signature := range slices.Sorted(maps.Keys(lowering.result.MethodTargets)) {
			targets := append([]int{signature}, lowering.result.MethodTargets[signature]...)
			reads := false
			for _, target := range targets {
				reads = reads || lowering.result.Functions[target].ArgumentsCount != 0
			}
			if !reads {
				continue
			}
			for _, target := range targets {
				if lowering.result.Functions[target].ArgumentsCount == 0 {
					local := len(lowering.result.Locals)
					lowering.result.Locals = append(
						lowering.result.Locals,
						ir.Local{Name: "argument_count", Type: ir.Number, Function: target},
					)
					lowering.result.Functions[target].ArgumentsCount = local + 1
					changed = true
				}
			}
		}
	}
}

func (lowering *lowering) fitCallArguments(function int, arguments []ir.Expression, spread []bool) {
	// Fixed arguments are fitted by the native call boundary, after ownership
	// planning. Fitting them here creates extra owned Weak and Union values.
	declared := lowering.result.Functions[function]
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
func (lowering *lowering) resolveStructuralMethodThunks(modules []*ast.SourceFile) {
	lowering.result.StructuralMethodThunks = map[int]bool{}
	finder := cycleFinder{lowering: lowering}
	// A derived receiver can expose an inherited implementation through a wider
	// structural view. Include every instantiation using that implementation.
	actualReceivers := map[int][]*checker.Type{}
	for _, instance := range lowering.instances {
		for _, function := range instance.methods {
			for _, local := range instance.thisLocals {
				actualReceivers[function] = append(actualReceivers[function], lowering.localTypes[local])
				actualReceivers[function] = append(actualReceivers[function], lowering.localAlso[local]...)
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
					for index, function := range lowering.result.Functions {
						if function.Receiver {
							lowering.result.StructuralMethodThunks[index] = true
						}
					}
				}
				if callee.Kind == ast.KindPropertyAccessExpression {
					access := callee.AsPropertyAccessExpression()
					receiver := lowering.concrete(lowering.checker.GetTypeAtLocation(access.Expression))
					method := lowering.checker.GetSymbolAtLocation(callee)
					nativeMethod := method != nil &&
						len(method.Declarations) > 0 &&
						method.Declarations[0].Kind == ast.KindMethodDeclaration &&
						access.QuestionDotToken == nil
					if !nativeMethod {
						for index, function := range lowering.result.Functions {
							if !function.Receiver ||
								function.MethodName != access.Name().Text() ||
								len(function.Parameters) == 0 {
								continue
							}
							local := function.Parameters[0]
							actuals := append([]*checker.Type{lowering.localTypes[local]}, lowering.localAlso[local]...)
							actuals = append(actuals, actualReceivers[index]...)
							for _, actual := range actuals {
								if actual == nil ||
									finder.template(receiver) ||
									finder.template(actual) ||
									lowering.checker.IsTypeAssignableTo(actual, receiver) {
									lowering.result.StructuralMethodThunks[index] = true
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
