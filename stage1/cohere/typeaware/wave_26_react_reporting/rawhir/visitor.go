// Visitor plumbing: the generic walks over an instruction's places, a terminal's successors, and a
// function's blocks.
//
// # Why these are functions and not an interface
//
// The obvious shape for a visitor in Go is an interface with one method per node type, which a pass
// embeds a default implementation of. That was rejected. A pass over this IR almost always cares
// about two or three of the 43 instruction values and wants to treat the rest uniformly - "every
// place this instruction reads" rather than "the second operand of a binary expression". An
// interface makes the uniform case the expensive one, since a pass must implement or inherit 43
// methods to express it.
//
// So the primitive is `EachPlace`, which yields every place an instruction touches without the
// caller knowing which variant it is, and a pass that needs a specific variant type-switches for
// it directly. That is the same division upstream's `visitors.rs` lands on after 1,147 lines.
//
// # Exhaustiveness
//
// Go cannot check a type switch for completeness, so `EachPlace` and `EachSuccessor` are the single
// place where all 43 instructions and all 20 terminals are enumerated. `TestEachPlaceCoversEvery
// InstructionValue` reflects over the package's variants and fails when one is missing a case, so a
// variant added without updating these goes red rather than silently contributing no places.
package high_level_intermediate_representation

// PlaceRole is why an instruction touches a place.
//
// A pass computing effects needs to distinguish the value being stored from the object being stored
// into, and a pass computing liveness needs to know which places are defined here rather than used.
// Both questions are answerable from the variant, but every pass asking them would re-derive the
// same mapping, so it is stated once here.
type PlaceRole uint8

const (
	// PlaceRoleUse is a place whose value is read.
	PlaceRoleUse PlaceRole = iota
	// PlaceRoleDefine is a place this instruction assigns to, including the instruction's LValue.
	PlaceRoleDefine
	// PlaceRoleReceiver is the object of a member access, method call, or property store.
	//
	// Separate from Use because it is the place a mutation attaches to: `o.f()` may mutate `o`, and
	// an effect pass wants the receiver distinguishable without re-matching the variant.
	PlaceRoleReceiver
)

// EachPlace calls visit for every place the instruction value touches, with the role it plays.
//
// The instruction's own LValue is NOT included; it is on the Instruction rather than the value, and
// `EachInstructionPlace` is the version that includes it.
//
// Places inside a nested function are not visited. A FunctionExpression yields its captures, which
// are places in THIS function, and the nested body is reached through Function.Functions.
func EachPlace(value InstructionValue, visit func(place Place, role PlaceRole)) {
	switch v := value.(type) {
	case *LoadLocal:
		visit(v.Place, PlaceRoleUse)
	case *LoadContext:
		visit(v.Place, PlaceRoleUse)
	case *DeclareLocal:
		visit(v.LValue, PlaceRoleDefine)
	case *DeclareContext:
		visit(v.LValue, PlaceRoleDefine)
	case *StoreLocal:
		visit(v.LValue, PlaceRoleDefine)
		visit(v.Value, PlaceRoleUse)
	case *StoreContext:
		visit(v.LValue, PlaceRoleDefine)
		visit(v.Value, PlaceRoleUse)
	case *Destructure:
		eachPatternPlace(v.LValue, visit)
		visit(v.Value, PlaceRoleUse)
	case *LoadGlobal:
		// No places: the name is resolved statically.
	case *StoreGlobal:
		visit(v.Value, PlaceRoleUse)
	case *PropertyLoad:
		visit(v.Object, PlaceRoleReceiver)
	case *PropertyStore:
		visit(v.Object, PlaceRoleReceiver)
		visit(v.Value, PlaceRoleUse)
	case *PropertyDelete:
		visit(v.Object, PlaceRoleReceiver)
	case *ComputedLoad:
		visit(v.Object, PlaceRoleReceiver)
		visit(v.Property, PlaceRoleUse)
	case *ComputedStore:
		visit(v.Object, PlaceRoleReceiver)
		visit(v.Property, PlaceRoleUse)
		visit(v.Value, PlaceRoleUse)
	case *ComputedDelete:
		visit(v.Object, PlaceRoleReceiver)
		visit(v.Property, PlaceRoleUse)
	case *CallExpression:
		visit(v.Callee, PlaceRoleUse)
		eachArgumentPlace(v.Args, visit)
	case *MethodCall:
		visit(v.Receiver, PlaceRoleReceiver)
		visit(v.Property, PlaceRoleUse)
		eachArgumentPlace(v.Args, visit)
	case *NewExpression:
		visit(v.Callee, PlaceRoleUse)
		eachArgumentPlace(v.Args, visit)
	case *BinaryExpression:
		visit(v.Left, PlaceRoleUse)
		visit(v.Right, PlaceRoleUse)
	case *UnaryExpression:
		visit(v.Value, PlaceRoleUse)
	case *PrefixUpdate:
		visit(v.LValue, PlaceRoleDefine)
		visit(v.Value, PlaceRoleUse)
	case *PostfixUpdate:
		visit(v.LValue, PlaceRoleDefine)
		visit(v.Value, PlaceRoleUse)
	case *Primitive:
		// No places.
	case *RegExpLiteral:
		// No places.
	case *TemplateLiteral:
		for _, subexpr := range v.Subexprs {
			visit(subexpr, PlaceRoleUse)
		}
	case *TaggedTemplateExpression:
		visit(v.Tag, PlaceRoleUse)
		for _, subexpr := range v.Subexprs {
			visit(subexpr, PlaceRoleUse)
		}
	case *TypeCastExpression:
		visit(v.Value, PlaceRoleUse)
	case *MetaProperty:
		// No places.
	case *ObjectExpression:
		for _, property := range v.Properties {
			if property.ComputedKey != nil {
				visit(*property.ComputedKey, PlaceRoleUse)
			}
			visit(property.Value, PlaceRoleUse)
		}
	case *ObjectMethod:
		// The body is a nested function, reached through Function.Functions.
	case *ArrayExpression:
		for _, element := range v.Elements {
			if element.Hole {
				continue
			}
			visit(element.Place, PlaceRoleUse)
		}
	case *FunctionExpression:
		for _, capture := range v.Captures {
			visit(capture, PlaceRoleUse)
		}
	case *Await:
		visit(v.Value, PlaceRoleUse)
	case *GetIterator:
		visit(v.Value, PlaceRoleUse)
	case *IteratorNext:
		visit(v.Iterator, PlaceRoleReceiver)
		visit(v.Collection, PlaceRoleUse)
	case *NextPropertyOf:
		visit(v.Value, PlaceRoleUse)
	case *JsxExpression:
		if v.Tag.Place != nil {
			visit(*v.Tag.Place, PlaceRoleUse)
		}
		for _, prop := range v.Props {
			visit(prop.Value, PlaceRoleUse)
		}
		for _, child := range v.Children {
			visit(child, PlaceRoleUse)
		}
	case *JsxFragment:
		for _, child := range v.Children {
			visit(child, PlaceRoleUse)
		}
	case *JsxText:
		// No places.
	case *StartMemoize:
		for _, dep := range v.Deps {
			// A global-rooted dependency carries a NAME rather than a place, so it holds a zero
			// `Place` that names no identifier. Visiting it would hand every operand consumer an
			// identifier id of zero, which single-assignment renaming would then treat as a real
			// binding.
			if dep.Root.IsGlobal {
				continue
			}
			visit(dep.Root.Place, PlaceRoleUse)
		}
	case *FinishMemoize:
		visit(v.Value, PlaceRoleUse)
	case *Debugger:
		// No places.
	case *UnsupportedNode:
		// No places: the node is opaque.
	}
}

// EachInstructionPlace calls visit for the instruction's LValue and every place its value touches.
func EachInstructionPlace(instruction *Instruction, visit func(place Place, role PlaceRole)) {
	visit(instruction.LValue, PlaceRoleDefine)
	EachPlace(instruction.Value, visit)
}

func eachArgumentPlace(args []Argument, visit func(place Place, role PlaceRole)) {
	for _, arg := range args {
		visit(arg.Place, PlaceRoleUse)
	}
}

func eachPatternPlace(pattern Pattern, visit func(place Place, role PlaceRole)) {
	switch p := pattern.(type) {
	case *PlacePattern:
		visit(p.Place, PlaceRoleDefine)
	case *ObjectPattern:
		for _, property := range p.Properties {
			if property.ComputedKey != nil {
				visit(*property.ComputedKey, PlaceRoleUse)
			}
			if property.Default != nil {
				visit(*property.Default, PlaceRoleUse)
			}
			eachPatternPlace(property.Value, visit)
		}
		if p.Rest != nil {
			visit(*p.Rest, PlaceRoleDefine)
		}
	case *ArrayPattern:
		for _, element := range p.Elements {
			if element.Value == nil {
				continue
			}
			if element.Default != nil {
				visit(*element.Default, PlaceRoleUse)
			}
			eachPatternPlace(element.Value, visit)
		}
		if p.Rest != nil {
			visit(*p.Rest, PlaceRoleDefine)
		}
	}
}

// EachTerminalPlace calls visit for every place a terminal reads.
//
// A terminal never defines a place, so every role here is a use.
func EachTerminalPlace(terminal Terminal, visit func(place Place, role PlaceRole)) {
	switch t := terminal.(type) {
	case *Return:
		visit(t.Value, PlaceRoleUse)
	case *Throw:
		visit(t.Value, PlaceRoleUse)
	case *If:
		visit(t.Test, PlaceRoleUse)
	case *Branch:
		visit(t.Test, PlaceRoleUse)
	case *Switch:
		visit(t.Test, PlaceRoleUse)
		for _, kase := range t.Cases {
			if kase.Test != nil {
				visit(*kase.Test, PlaceRoleUse)
			}
		}
	case *Try:
		if t.HandlerBinding != nil {
			visit(*t.HandlerBinding, PlaceRoleDefine)
		}
	}
}

// EachSuccessor calls visit for every block control can reach directly from this terminal.
//
// Fallthrough is NOT yielded. It is not an edge: control reaches it through the construct's arms,
// not from the terminal. `EachSuccessorAndFallthrough` is the version for a caller that wants the
// structural link as well, such as a printer or a pass rebuilding syntax.
//
// The distinction is load-bearing. A dataflow analysis given the fallthrough as an edge propagates
// values along a path that does not exist, and the error is silent because the fallthrough is
// genuinely reachable by other means.
func EachSuccessor(terminal Terminal, visit func(block BlockId)) {
	visitReal := func(block BlockId) {
		if HasBlock(block) {
			visit(block)
		}
	}
	switch t := terminal.(type) {
	case *Return, *Throw, *Unreachable, *Unsupported:
		// Control leaves the function.
	case *Goto:
		visitReal(t.Block)
	case *If:
		visitReal(t.Consequent)
		visitReal(t.Alternate)
	case *Branch:
		visitReal(t.Consequent)
		visitReal(t.Alternate)
	case *Switch:
		for _, kase := range t.Cases {
			visitReal(kase.Block)
		}
		// A switch with no default can fall past every case, so the fallthrough IS an edge here.
		if !hasDefaultCase(t.Cases) {
			visitReal(t.Fallthrough)
		}
	case *While:
		visitReal(t.Test)
	case *DoWhile:
		visitReal(t.Loop)
	case *For:
		visitReal(t.Init)
	case *ForOf:
		visitReal(t.Init)
	case *ForIn:
		visitReal(t.Init)
	case *Logical:
		visitReal(t.Test)
	case *Ternary:
		visitReal(t.Test)
	case *Optional:
		visitReal(t.Test)
	case *Sequence:
		visitReal(t.Block)
	case *Label:
		visitReal(t.Block)
	case *Try:
		visitReal(t.Block)
		visitReal(t.Handler)
	case *MaybeThrow:
		visitReal(t.Continuation)
		visitReal(t.Handler)
	case *Scope:
		// The scope body is a real edge. The fallthrough is not, exactly as for every other
		// structured terminal: control reaches it by leaving the body, not from here.
		visitReal(t.Block)
	}
}

func hasDefaultCase(cases []SwitchCase) bool {
	for _, kase := range cases {
		if kase.Test == nil {
			return true
		}
	}
	return false
}

// Fallthrough returns the block where the construct this terminal begins completes, if it has one.
func Fallthrough(terminal Terminal) (BlockId, bool) {
	var block BlockId
	switch t := terminal.(type) {
	case *If:
		block = t.Fallthrough
	case *Branch:
		block = t.Fallthrough
	case *Switch:
		block = t.Fallthrough
	case *While:
		block = t.Fallthrough
	case *DoWhile:
		block = t.Fallthrough
	case *For:
		block = t.Fallthrough
	case *ForOf:
		block = t.Fallthrough
	case *ForIn:
		block = t.Fallthrough
	case *Logical:
		block = t.Fallthrough
	case *Ternary:
		block = t.Fallthrough
	case *Optional:
		block = t.Fallthrough
	case *Sequence:
		block = t.Fallthrough
	case *Label:
		block = t.Fallthrough
	case *Try:
		block = t.Fallthrough
	case *Scope:
		block = t.Fallthrough
	default:
		return InvalidBlock, false
	}
	return block, HasBlock(block)
}

// EachSuccessorAndFallthrough calls visit for every successor and then the fallthrough.
//
// For a printer, a structural walk, or a reachability check that wants to reach every block a
// construct owns. Not for dataflow; see EachSuccessor.
func EachSuccessorAndFallthrough(terminal Terminal, visit func(block BlockId)) {
	seen := false
	fallthroughBlock, hasFallthrough := Fallthrough(terminal)
	EachSuccessor(terminal, func(block BlockId) {
		if hasFallthrough && block == fallthroughBlock {
			seen = true
		}
		visit(block)
	})
	if hasFallthrough && !seen {
		visit(fallthroughBlock)
	}
}

// TerminalOrder returns the terminal's position in evaluation order.
func TerminalOrder(terminal Terminal) EvaluationOrder {
	switch t := terminal.(type) {
	case *Return:
		return t.Order
	case *Throw:
		return t.Order
	case *Unreachable:
		return t.Order
	case *Unsupported:
		return t.Order
	case *Goto:
		return t.Order
	case *If:
		return t.Order
	case *Branch:
		return t.Order
	case *Switch:
		return t.Order
	case *While:
		return t.Order
	case *DoWhile:
		return t.Order
	case *For:
		return t.Order
	case *ForOf:
		return t.Order
	case *ForIn:
		return t.Order
	case *Logical:
		return t.Order
	case *Ternary:
		return t.Order
	case *Optional:
		return t.Order
	case *Sequence:
		return t.Order
	case *Label:
		return t.Order
	case *Try:
		return t.Order
	case *MaybeThrow:
		return t.Order
	case *Scope:
		return t.Order
	}
	return 0
}
