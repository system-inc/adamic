package lower

import (
	"maps"
	"slices"

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

// argumentsObjectRefusal checks a use of JavaScript's arguments object.
func (l *lowering) argumentsObjectRefusal(node *ast.Node) error {
	if node.Kind == ast.KindIdentifier && node.Text() == "arguments" {
		// JavaScript's arguments object, not a variable the program named arguments.
		if symbol := l.checker.GetSymbolAtLocation(node); symbol != nil && len(symbol.Declarations) == 0 {
			if err := l.argumentsRefusal(node); err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	refusalRules.register(2600, "argumentsObjectRefusal", (*lowering).argumentsObjectRefusal)
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
			if !l.isArrayType(proven) {
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
				proven := l.concrete(l.checker.GetTypeAtLocation(node.AsCallExpression().Expression))
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
	// Once per pair of identity classes of targets and records, not per pair of types (count_types_classes.go).
	l.countTypeTargets(types)
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
		// By signature, never in the map's order: each local's index is in its C name.
		for _, signature := range slices.Sorted(maps.Keys(l.result.MethodTargets)) {
			targets := append([]int{signature}, l.result.MethodTargets[signature]...)
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

// spreadPackStop judges a call with a spread against the slots its callee reads (#p9zm9q5). Such a
// call packs each argument as it's held (packCallArguments), and the callee reads each slot as its
// own parameter's type and copies a rest slot for slot (copyRest), so nothing converts on the way:
// a number or a boolean packed where a union is read is its bits read as a heap pointer. So an
// argument is admitted only where it's held as the slot is (packedAlike). A fixed argument before
// the first spread meets its own parameter; a spread's elements, and every argument after a spread,
// meet each slot they can land in. The first skip arguments, a method's receiver, have no node and
// aren't judged. Stopping here fails closed; packing each value as its slot is #p9zm9q5's Must for
// Lower.
func (l *lowering) spreadPackStop(nodes []*ast.Node, arguments []ir.Expression, spread []bool, skip int, function int) error {
	if len(spread) == 0 {
		return nil
	}
	declared := l.result.Functions[function]
	slots := []ir.Type{}
	for index, parameter := range declared.Parameters {
		if declared.RestElement != 0 && index == len(declared.Parameters)-1 {
			break
		}
		slots = append(slots, l.result.Locals[parameter].Type)
	}
	from := func(position int) []ir.Type {
		landing := slices.Clone(slots[min(position, len(slots)):])
		if declared.RestElement != 0 {
			landing = append(landing, declared.RestElement)
		}
		return landing
	}
	position := 0
	expanded := false
	for index, argument := range arguments {
		var landing []ir.Type
		held := argument.Type()
		doing := "passing a " + typeName(held)
		switch {
		case len(spread) > index && spread[index]:
			expanded = true
			if literal, isLiteral := argument.(ir.ArrayLiteral); isLiteral && len(literal.Elements) == 0 {
				continue
			}
			held = l.spreadElement(nodes[index-skip], argument)
			doing = "of elements held as a " + typeName(held)
			landing = from(position)
		case expanded:
			landing = from(position)
		case position < len(slots):
			landing = slots[position : position+1]
			position++
		default:
			landing = from(position)
			position++
		}
		if index < skip {
			continue
		}
		for _, slot := range landing {
			if !packedAlike(slot, held) {
				return l.notYet(nodes[index-skip], "a call with a spread "+doing+" where the callee reads a "+typeName(slot))
			}
		}
	}
	return nil
}

// spreadElement is how a spread argument's elements are held: a tuple literal's own element, or the
// element of the array type callArguments admitted.
func (l *lowering) spreadElement(node *ast.Node, argument ir.Expression) ir.Type {
	if literal, isLiteral := argument.(ir.ArrayLiteral); isLiteral {
		return literal.Element
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(node.AsSpreadElement().Expression))
	element, _ := l.representation(l.checker.GetTypeArguments(proven)[0])
	return element
}

// packedAlike says a packed slot holding a value of type held reads back as type slot with nothing
// converted, the rule copyRest's restHoldsAlike copies by: the same type, or a reference read as a
// reference. A weak reference is made by adamic_weak_of, so it's alike only to itself.
func packedAlike(slot, held ir.Type) bool {
	if slot == held {
		return true
	}
	return slot.IsReference() && held.IsReference() && slot != ir.Weak && held != ir.Weak
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
