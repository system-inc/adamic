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
	return l.callArgumentsFor(nodes, nil, -1)
}

func (l *lowering) callArgumentsFor(nodes []*ast.Node, call *ast.CallExpression, function int) ([]ir.Expression, []bool, error) {
	arguments := []ir.Expression{}
	boxedRest := -1
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
				// Family 23, rulings.md:225 and :342 as amended Nov 24 at 19:29:49Z
				// on #bbssdcy. l.isArrayType above admits only Array or ReadonlyArray,
				// excluding subclasses. Built-in Symbol.iterator writes are refused.
				// The accepted checker layout keeps phantom brands out of this road.
				if !known || element != ir.Union || l.unbranded(proven) != proven ||
					!l.boxedRestSpread(call, function, index, elements[0]) {
					return nil, nil, l.notYet(node, "a call spreading elements that cannot be packed")
				}
				boxedRest = index
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
	if boxedRest >= 0 {
		rest := len(l.result.Functions[function].Parameters) - 1
		for index, argument := range arguments {
			if index == boxedRest {
				continue
			}
			// The spread pack reads fixed parameters in their declared slot
			// representations, so fit explicit fixed arguments before packing.
			if index < rest {
				local := l.result.Functions[function].Parameters[index]
				held := l.result.Locals[local].Type
				arguments[index] = fit(argument, held)
				if arguments[index].Type() != held {
					return nil, nil, l.notYet(nodes[boxedRest], "a call spreading elements that cannot be packed")
				}
			}
		}
	}
	return arguments, spread, nil
}

// boxedRestSpread proves the fixed prefix ends before the spread and that the
// resolved call and its actual lowered target agree on the rest array layout.
func (l *lowering) boxedRestSpread(call *ast.CallExpression, function, position int, source *checker.Type) bool {
	if call == nil || source.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	target := l.result.Functions[function]
	if target.Receiver || target.RestElement != ir.Union {
		return false
	}
	// An explicit slot for the spread is separate from the rest slot read here.
	// Keep the stop when a caller set either spread node's slot.
	spread := call.Arguments.Nodes[position]
	if l.slotSet(spread) || l.slotSet(spread.AsSpreadElement().Expression) {
		return false
	}
	signature := l.checker.GetResolvedSignature(call.AsNode())
	if signature == nil || !signature.HasRestParameter() ||
		signature.Declaration() == nil || signature.Declaration().Body() == nil {
		return false
	}
	rest := len(signature.Parameters()) - 1
	if rest < 0 || position < rest || rest != len(target.Parameters)-1 {
		return false
	}
	for index, argument := range call.Arguments.Nodes {
		// Other spreads can shift the prefix or contribute differently packed
		// slots. Keep the existing stop for a mixed pack.
		if index != position && argument.Kind == ast.KindSpreadElement {
			return false
		}
	}
	array := l.concrete(l.checker.GetTypeOfSymbol(signature.Parameters()[rest]))
	// The rest array keeps the same phantom brand exclusion as the operand.
	if !l.isArrayType(array) || l.unbranded(array) != array {
		return false
	}
	destination := l.elementTypeOfArray(array)
	held, known := l.representation(destination)
	if !known || held != ir.Union || destination.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	// Exact member identity preserves all of Adamic's writable, optional
	// and nominal guarantees. No structural upcast is needed for this copy.
	for _, arm := range source.Types() {
		if !slices.Contains(destination.Types(), arm) {
			return false
		}
	}
	return true
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

// spreadPackFit makes each argument of a call with a spread what the slots its callee reads hold
// (#p9zm9q5). Such a call packs each argument as it's held (packCallArguments), and the callee reads
// each slot as its own parameter's type and copies a rest slot for slot (copyRest), so nothing
// converts on the way: a number or a boolean packed where a union is read is its bits read as a heap
// pointer. So each argument is converted here, before it's packed, into arguments in place. A fixed
// argument before the first spread meets its own parameter; a spread's elements, and every argument
// after a spread, meet each slot they can land in. An argument held as each slot it can land in reads
// (packedAlike) stays as it is. Otherwise every slot it can land in must be one type, the one it's
// made: a fixed argument fitted to it, and a spread's elements converted element by element
// (spreadConverted). Slots that differ stop, even where they're alike to each other, since a union's
// box landing in a string slot would be read as a string. The first skip arguments, a method's
// receiver, have no node and aren't judged. slots are the callee's parameters before its rest, and
// rest its rest's element, or 0 for none: a function's from its locals (functionSlots), a function
// value's from its one signature (signatureSlots). Stopping here fails closed.
func (l *lowering) spreadPackFit(nodes []*ast.Node, arguments []ir.Expression, spread []bool, skip int, slots []ir.Type, rest ir.Type) error {
	if len(spread) == 0 {
		return nil
	}
	from := func(position int) []ir.Type {
		landing := slices.Clone(slots[min(position, len(slots)):])
		if rest != 0 {
			landing = append(landing, rest)
		}
		return landing
	}
	position := 0
	expanded := false
	for index, argument := range arguments {
		var landing []ir.Type
		held := argument.Type()
		doing := "passing a " + typeName(held)
		spreads := len(spread) > index && spread[index]
		switch {
		case spreads:
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
		if !slices.ContainsFunc(landing, func(slot ir.Type) bool { return !packedAlike(slot, held) }) {
			continue
		}
		to := landing[0]
		for _, slot := range landing {
			if slot != to {
				return l.notYet(nodes[index-skip], "a call with a spread "+doing+" where the callee's slots read as a "+typeName(to)+" and a "+typeName(slot))
			}
		}
		stop := "a call with a spread " + doing + " where the callee reads a " + typeName(to)
		if !spreads {
			arguments[index] = fit(argument, to)
			if arguments[index].Type() != to {
				return l.notYet(nodes[index-skip], stop)
			}
			continue
		}
		converted, err := l.spreadConverted(nodes[index-skip], nodes[index-skip], argument, held, to, stop)
		if err != nil {
			return err
		}
		// spreadPacks and checkClosureSpreadLayouts read a spread's elements through spreadElement, which
		// takes an ir.ArrayLiteral's Element and anything else's from the checker's type of the spread's
		// source, the elements before they were converted. Wrapped in a literal of to that spreads the
		// converted array, the spread reads as to, as it's packed.
		arguments[index] = ir.ArrayLiteral{Element: to, Elements: []ir.Expression{converted}, Spread: []bool{true}}
	}
	return nil
}

// functionSlots is a lowered function's parameter slots before its rest, and its rest's element.
func (l *lowering) functionSlots(function int) ([]ir.Type, ir.Type) {
	declared := l.result.Functions[function]
	slots := []ir.Type{}
	for index, parameter := range declared.Parameters {
		if declared.RestElement != 0 && index == len(declared.Parameters)-1 {
			break
		}
		slots = append(slots, l.result.Locals[parameter].Type)
	}
	return slots, declared.RestElement
}

// signatureSlots is what a function value's one signature takes, as the closure road fits its fixed
// arguments: each parameter's representation, a defaulted one's maybe, and a rest's element. A slot it
// can't represent stops, since nothing could be judged against it.
func (l *lowering) signatureSlots(node *ast.Node, signature *checker.Signature) ([]ir.Type, ir.Type, error) {
	slots := []ir.Type{}
	for _, parameter := range signature.Parameters() {
		declaration := (*ast.Node)(nil)
		if len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter {
			declaration = parameter.Declarations[0]
		}
		proven := l.checker.GetTypeOfSymbol(parameter)
		if declaration != nil && declaration.AsParameterDeclaration().DotDotDotToken != nil {
			elements := []*checker.Type{}
			if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
				elements = l.checker.GetTypeArguments(proven)
			}
			element, known := ir.Type(0), false
			if len(elements) == 1 {
				element, known = l.representation(elements[0])
			}
			if !known {
				return nil, 0, l.notYet(node, "a call with a spread through a function value whose rest element has no representation")
			}
			return slots, element, nil
		}
		takes, known := l.representation(proven)
		if !known {
			return nil, 0, l.notYet(node, "a call with a spread through a function value whose parameter has no representation")
		}
		if declaration != nil && declaration.AsParameterDeclaration().Initializer != nil {
			takes = ir.Maybe(takes)
		}
		slots = append(slots, takes)
	}
	return slots, 0, nil
}

// spreadElement is how a spread argument's elements are held: a tuple literal's own element, or the
// element of the array type callArguments admitted. Its first type argument is every element's type
// only because callArguments admits nothing else: a spread that isn't an array stops there, a tuple
// that isn't a literal among them, and so does an array without exactly one element type. An iterable
// admitted later must be read here too, or a tuple's first element would be taken for all of them
// (Emit's should on #whe5bdk).
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
// reference. A weak reference is made by adamic_weak_of, so it's alike only to itself. A number read as
// number | undefined is alike too: both are the slot's number member, and adamic_maybe_number_unpack reads
// every double but ADAMIC_UNDEFINED_BITS, a NaN no arithmetic makes, as present. A boolean isn't: number |
// undefined's twin, boolean | undefined, is its own member.
func packedAlike(slot, held ir.Type) bool {
	if slot == held || slot == ir.MaybeNumber && held == ir.Number {
		return true
	}
	return slot.IsReference() && held.IsReference() && slot != ir.Weak && held != ir.Weak
}

// closureSpreadSite is a call with a spread through a function value, as callClosure made it: its
// argument nodes, and the call whose layout the emitter packs for (ir.ClosureArgumentLayout), its
// Packs what each argument packs.
type closureSpreadSite struct {
	nodes []*ast.Node
	call  ir.CallClosure
}

// spreadPacks is how each argument of a call with a spread is packed: a fixed one as it's held, a
// spread's elements as spreadElement says, and an empty array literal's spread as 0, packing nothing.
func (l *lowering) spreadPacks(nodes []*ast.Node, arguments []ir.Expression, spread []bool) []ir.Type {
	packs := []ir.Type{}
	for index, argument := range arguments {
		if len(spread) <= index || !spread[index] {
			packs = append(packs, argument.Type())
			continue
		}
		if literal, isLiteral := argument.(ir.ArrayLiteral); isLiteral && len(literal.Elements) == 0 {
			packs = append(packs, 0)
			continue
		}
		packs = append(packs, l.spreadElement(nodes[index], argument))
	}
	return packs
}

// checkClosureSpreadLayouts judges each call with a spread through a function value against every target
// its layout packs for (#wzr4atz). spreadPackFit fitted it to the function value's one signature,
// but the emitter packs the arguments once and fills every slot of ir.ClosureArgumentLayout from that
// pack. The layout's targets are every closure record its function type admits (FunctionTypeTargets).
// Each rest tail is copied from the pack segment by segment, each slot made what the tail holds
// (copySegments, by the call's Packs), so a number lands boxed in one target's rest of a union and as
// itself in another's rest of numbers. A fixed slot is read as the layout's type with nothing converted
// (packedParameter), so each argument is judged by packedAlike against each fixed slot it can land in: a
// fixed argument before the first spread, at position p, against the fixed slot p; a spread's elements,
// and every argument after a spread, against every fixed slot from p on, since where they land is known
// only at run time. A rest tail holds a weak reference only as itself (restSlot converts none to or from
// one), so a tail it lands in is judged as a fixed slot is where either is weak. Targets that aren't
// bounded are every function, which may read anything, so such a call stops wherever it packs something.
// FunctionTypeTargets is final once findCycles has run resolveCountTypes, the one pass that fills it, so
// this runs after it. Stopping here fails closed.
func (l *lowering) checkClosureSpreadLayouts() error {
	for _, site := range l.closureSpreadSites {
		if len(site.nodes) > 0 && l.firstLinkSkips(site.nodes[0]) {
			continue
		}
		layout, unknown := l.result.ClosureArgumentSlots(site.call)
		position := 0
		expanded := false
		for index, held := range site.call.Packs {
			var landing, tails []ir.Type
			switch {
			case len(site.call.Spread) > index && site.call.Spread[index]:
				expanded = true
				if held == 0 {
					continue
				}
				landing = layout.Fixed[min(position, len(layout.Fixed)):]
			case expanded:
				landing = layout.Fixed[min(position, len(layout.Fixed)):]
			default:
				if position < len(layout.Fixed) {
					landing = layout.Fixed[position : position+1]
				}
				for _, rest := range layout.Rest {
					if rest.Start <= position {
						tails = append(tails, rest.Element)
					}
				}
				position++
			}
			if expanded {
				for _, rest := range layout.Rest {
					tails = append(tails, rest.Element)
				}
			}
			if unknown {
				return l.notYet(site.nodes[index], "a call with a spread through a function value whose targets aren't bounded, so its layout may read anything where the call packs a "+typeName(held))
			}
			for _, tail := range tails {
				if (tail == ir.Weak || held == ir.Weak) && tail != held {
					landing = append(slices.Clone(landing), tail)
				}
			}
			for _, slot := range landing {
				if !packedAlike(slot, held) {
					return l.notYet(site.nodes[index], "a call with a spread through a function value whose layout reads a "+typeName(slot)+" where the call packs a "+typeName(held))
				}
			}
		}
	}
	return nil
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
