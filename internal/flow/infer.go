package flow

import (
	"reflect"
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// InferAliasingEffects is Adamic's own effect inference: what each instruction of a graph in single
// assignment form does to the values it reads and writes, from the IR it points back at.
//
// cohere makes effects from a signature table keyed by the callee's name (effects.go says why that
// isn't lifted). Adamic needs none: every runtime operation is its own IR node, so what push, set or
// splice does is known exactly, and only a call (a function, a closure, or a runtime loop calling a
// callback) is assumed to do what cohere assumes of an unknown one: to mutate, conditionally and
// transitively, anything it's handed.
//
// # Escaped values
//
// A call can also mutate what it isn't handed: anything it can reach through a global or a
// captured variable (neither of which this graph tracks). So a value is escaped when something
// outside this function may hold it: a parameter (the caller may have stored it anywhere), a value
// read from a global or a captured variable, a call's result, or a value handed to a call or stored
// into an escaped value. Every call mutates, conditionally and transitively, every escaped value
// that reaches it. That's conservative on purpose: a range too long costs a missed optimization,
// and a range too short is a miscompile waiting for the pass that trusts it.
//
// Strings, numbers and booleans can't be mutated (Adamic's strings are immutable values), so a
// value of those types is created primitive and takes part in nothing.
func InferAliasingEffects(function *Function) *AliasingEffects {
	inference := &inference{function: function, escaped: map[IdentifierId]bool{}, iterated: map[*ir.Statement][]Place{}, parts: map[[2]uint32]IdentifierId{}}
	// Escape grows as effects are made (a value handed to a call escapes), and every call has to
	// mutate every value escaped by then, wherever in the function it escaped. So the effects are
	// made again until the escaped set stops growing; it only grows, and it's bounded by the
	// function's values.
	for {
		before := len(inference.escaped)
		inference.effects = &AliasingEffects{byInstruction: map[InstructionId][]AliasingEffect{}}
		inference.run()
		if len(inference.escaped) == before {
			return inference.effects
		}
	}
}

type inference struct {
	function *Function
	effects  *AliasingEffects
	escaped  map[IdentifierId]bool

	// current is, while an instruction's effects are made, the value each variable holds there.
	current map[DeclarationId]IdentifierId

	// iterated is what each for...of iterates, as its first part found it, for the elements its
	// second part binds.
	iterated map[*ir.Statement][]Place

	// list is the effects of the instruction being made, and instruction its id.
	list        []AliasingEffect
	instruction InstructionId

	// parts are the temporaries minted for values read out of a container (see define), by the
	// instruction and the container, so each run of the inference reuses the same ones.
	parts map[[2]uint32]IdentifierId
}

// shape is what an expression's value is to the values this graph tracks.
type shape struct {
	same    []Place // the very same object as one of these
	part    []Place // read out of one of these: a field, an element
	holds   []Place // a container made here, holding some of these
	fresh   bool    // a container made here
	unknown bool    // from somewhere this function can't see: an untracked variable, a call
}

func (s shape) roots() []Place {
	return append(append(append([]Place{}, s.same...), s.part...), s.holds...)
}

func (s *shape) merge(other shape) {
	s.same = append(s.same, other.same...)
	s.part = append(s.part, other.part...)
	s.holds = append(s.holds, other.holds...)
	s.fresh = s.fresh || other.fresh
	s.unknown = s.unknown || other.unknown
}

// run makes every instruction's effects, walking the blocks in reverse postorder with the value
// each variable holds.
func (n *inference) run() {
	exits := map[BlockId]map[DeclarationId]IdentifierId{}
	for _, parameter := range n.function.Params {
		n.escaped[parameter.Identifier] = true
	}
	for _, block := range n.function.Blocks {
		n.current = map[DeclarationId]IdentifierId{}
		if block.Id == n.function.Entry {
			for _, parameter := range n.function.Params {
				n.current[n.declaration(parameter)] = parameter.Identifier
			}
		}
		// A variable's value on entry is the same from every predecessor already walked, or there'd
		// be a phi; a predecessor not yet walked is a back edge, whose value a phi carries.
		for _, predecessor := range block.Predecessors {
			for variable, value := range exits[predecessor] {
				n.current[variable] = value
			}
		}
		for _, phi := range block.Phis {
			n.current[n.declaration(phi.Place)] = phi.Place.Identifier
			for _, operand := range phi.Operands {
				if n.escaped[operand.Identifier] {
					n.escaped[phi.Place.Identifier] = true
				}
			}
		}
		for _, id := range block.Instructions {
			n.list = nil
			n.instruction = id
			n.run1(n.function.Instructions[id])
			if len(n.list) > 0 {
				n.effects.byInstruction[id] = n.list
			}
		}
		exits[block.Id] = n.current
	}
}

func (n *inference) declaration(place Place) DeclarationId {
	return n.function.Identifiers[place.Identifier].Declaration
}

// run1 makes one instruction's effects: what evaluating it does, then what it defines.
func (n *inference) run1(instruction *Instruction) {
	switch statement := (*instruction.At).(type) {
	case ir.SetProperty:
		object := n.value(statement.Object)
		value := n.value(statement.Value)
		n.store(object, value)
	case ir.SetIndex:
		array := n.value(statement.Array)
		n.value(statement.Index)
		value := n.value(statement.Value)
		n.store(array, value)
	case ir.ForOf:
		if instruction.Part == 0 {
			n.iterated[instruction.At] = n.value(statement.Iterable).roots()
			break
		}
		// Each element is read out of what the loop iterates, as its first part found it (reverse
		// postorder walks that part first): a mutation through the element is one of that too.
		for _, define := range instruction.Defines {
			n.define(define, shape{part: n.iterated[instruction.At]})
		}
	default:
		var defined shape
		if instruction.Expression != nil {
			defined = n.value(instruction.Expression)
		}
		for _, define := range instruction.Defines {
			n.define(define, defined)
		}
	}
	// What an instruction defines, it defines last: a read of a variable in the same instruction is
	// of its old value, and the new one holds from here on.
	for _, define := range instruction.Defines {
		n.current[n.declaration(define)] = define.Identifier
	}
}

// define gives a defined value its effects: created, and related to what its expression's value is.
func (n *inference) define(into Place, value shape) {
	local := n.function.Identifiers[into.Identifier].Declaration - 1
	if !mutable(n.function.Program.Locals[local].Type) {
		n.list = append(n.list, create(into, EffectValuePrimitive))
		return
	}
	if len(value.part) == 1 && len(value.same) == 0 && len(value.holds) == 0 && !value.fresh && !value.unknown {
		n.list = append(n.list, flow(AliasingEffectCreateFrom, value.part[0], into))
	} else {
		n.list = append(n.list, create(into, EffectValueMutable))
		for _, from := range value.same {
			n.list = append(n.list, flow(AliasingEffectAssign, from, into))
		}
		// A value read out of a container, mixed with others (a ?? b, c ? d : e), goes through a
		// temporary created from the container, which is what React's graph has for every read:
		// only CreateFrom carries a mutation back to the container transitively, on to whatever the
		// container holds, and an Alias straight from the container wouldn't.
		for _, from := range value.part {
			key := [2]uint32{uint32(n.instruction), uint32(from.Identifier)}
			part, ok := n.parts[key]
			if !ok {
				part = n.function.NewIdentifier("", 0).Id
				n.parts[key] = part
			}
			n.list = append(n.list, flow(AliasingEffectCreateFrom, from, Place{Identifier: part}))
			n.list = append(n.list, flow(AliasingEffectAssign, Place{Identifier: part}, into))
			if n.escaped[from.Identifier] {
				n.escaped[part] = true
			}
		}
		for _, from := range value.holds {
			n.list = append(n.list, flow(AliasingEffectCapture, from, into))
		}
	}
	if value.unknown {
		n.escaped[into.Identifier] = true
	}
	for _, root := range value.roots() {
		if n.escaped[root.Identifier] {
			n.escaped[into.Identifier] = true
		}
	}
}

// store is a write into a container: the container is mutated, and holds what was written.
func (n *inference) store(container shape, value shape) {
	for _, root := range container.roots() {
		n.list = append(n.list, mutate(AliasingEffectMutate, root))
		for _, from := range value.roots() {
			n.list = append(n.list, flow(AliasingEffectCapture, from, root))
		}
	}
	if container.unknown || anyEscaped(n.escaped, container.roots()) {
		n.escape(value)
		// An escaped container may be reachable from any other escaped value (two parameters may
		// be the same object, or one may hold the other), so writing into it may change them all.
		n.mutateEscaped()
	}
}

// call is what a call does to what it's handed and to every escaped value: mutates them,
// conditionally and transitively. Its result is unknown, and may be any of them.
func (n *inference) call(operands []shape) shape {
	result := shape{unknown: true}
	for _, operand := range operands {
		n.escape(operand)
	}
	n.mutateEscaped()
	for _, operand := range operands {
		result.same = append(result.same, operand.roots()...)
	}
	return result
}

// mutateEscaped mutates, conditionally and transitively, every escaped value that reaches here: what
// a call can reach, or what a write into an escaped value may change.
func (n *inference) mutateEscaped() {
	// In a fixed order, since an effect's position is the alias graph's time.
	variables := make([]DeclarationId, 0, len(n.current))
	for variable := range n.current {
		variables = append(variables, variable)
	}
	slices.Sort(variables)
	for _, variable := range variables {
		value := n.current[variable]
		if n.escaped[value] && mutable(n.typeOf(value)) {
			n.list = append(n.list, mutate(AliasingEffectMutateTransitiveConditionally, Place{Identifier: value}))
		}
	}
}

func (n *inference) escape(value shape) {
	for _, root := range value.roots() {
		n.escaped[root.Identifier] = true
	}
}

func (n *inference) typeOf(value IdentifierId) ir.Type {
	return n.function.Program.Locals[n.function.Identifiers[value].Declaration-1].Type
}

func anyEscaped(escaped map[IdentifierId]bool, places []Place) bool {
	for _, place := range places {
		if escaped[place.Identifier] {
			return true
		}
	}
	return false
}

// mutable reports whether a value of a type can be mutated: an object, an array, a map, a closure
// (whose cells can change), or a union that may be one. Numbers, booleans and strings can't.
func mutable(valueType ir.Type) bool {
	return valueType.IsReference() && valueType != ir.String
}

// value is the shape of an expression's value, and makes the effects of evaluating it.
func (n *inference) value(expression ir.Expression) shape {
	if expression == nil {
		return shape{}
	}
	switch expression := expression.(type) {
	case ir.Read:
		declared := n.function.Program.Locals[expression.Local]
		if !mutable(expression.Of) {
			return shape{}
		}
		if declared.Global || declared.Captured {
			return shape{unknown: true}
		}
		value, ok := n.current[DeclarationId(expression.Local+1)]
		if !ok {
			return shape{}
		}
		return shape{same: []Place{{Identifier: value}}}
	case ir.Conditional:
		n.value(expression.Condition)
		result := n.value(expression.WhenTrue)
		result.merge(n.value(expression.WhenNot))
		return result
	case ir.Coalesce:
		result := n.value(expression.Value)
		result.merge(n.value(expression.Fallback))
		n.value(expression.Panic)
		return result
	case ir.Box:
		return n.value(expression.Value)
	case ir.Narrow:
		return n.value(expression.Value)
	case ir.Unwrap:
		return n.value(expression.Value)
	case ir.CheckedCast:
		result := n.value(expression.Value)
		for _, allowed := range expression.Allowed {
			n.value(allowed)
		}
		return result
	case ir.ObjectLiteral:
		result := shape{fresh: true}
		if expression.Spread != nil {
			// A copy of the spread's object: its fields are the same values the spread holds.
			result.holds = append(result.holds, n.value(expression.Spread).roots()...)
		}
		for _, field := range expression.Fields {
			result.holds = append(result.holds, n.value(field.Value).roots()...)
		}
		return result
	case ir.ArrayLiteral:
		result := shape{fresh: true}
		for _, element := range expression.Elements {
			result.holds = append(result.holds, n.value(element).roots()...)
		}
		return result
	case ir.Property:
		return shape{part: n.value(expression.Object).roots()}
	case ir.ArrayIndex:
		array := n.value(expression.Array)
		n.value(expression.Index)
		return shape{part: array.roots()}
	case ir.ArrayPush:
		array := n.value(expression.Array)
		value := n.value(expression.Value)
		n.store(array, value)
		return shape{}
	case ir.ArrayPop:
		array := n.value(expression.Array)
		n.store(array, shape{})
		return shape{part: array.roots()}
	case ir.ArrayReverse:
		array := n.value(expression.Array)
		n.store(array, shape{})
		return array
	case ir.DateCall:
		return n.dateCall(expression)
	case ir.MakeClosure:
		// Its cells hold captured variables, which this graph doesn't track.
		return shape{fresh: true, unknown: true}
	case ir.Call:
		operands := []shape{}
		for _, argument := range expression.Arguments {
			operands = append(operands, n.value(argument))
		}
		return n.call(operands)
	case ir.CallClosure:
		// Every target, bounded or Unknown, has the same conservative effect here:
		// anything it can reach may change. Evaluate the callee as an operand too.
		return n.call(n.operands(expression))
	}
	// Everything else: the effects of what it evaluates, field by field, and then, as cohere assumes
	// of a call it has no signature for, that it mutates and may return any of them. That covers
	// the methods that call a callback (map, forEach, a Map's or Set's forEach, reduce, sort, Array.from), the ones that write
	// (splice, fill, set, delete), and a node added to the IR after this was written, which is then
	// treated as the most it could be rather than as nothing.
	operands := n.operands(expression)
	if !mutable(expression.Type()) && !writes(expression) && !callsBack(expression) {
		return shape{}
	}
	return n.call(operands)
}

// operands evaluates every expression an IR node holds, in the order its fields are declared.
func (n *inference) operands(expression ir.Expression) []shape {
	var operands []shape
	var walk func(value reflect.Value)
	walk = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			if inner, ok := value.Interface().(ir.Expression); ok {
				operands = append(operands, n.value(inner))
				return
			}
			walk(value.Elem())
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				walk(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				walk(value.Index(index))
			}
		}
	}
	walk(reflect.ValueOf(expression))
	return operands
}

// writes reports whether an IR node writes into a container it's handed.
func writes(expression ir.Expression) bool {
	switch call := expression.(type) {
	case ir.NodeBufferCall:
		return call.Function == "buffer_set" || call.Function == "hash_update" || call.Function == "hash_digest"
	case ir.ArraySplice, ir.ArrayFill, ir.ArraySort, ir.MapSet, ir.MapDelete:
		return true
	}
	return false
}

// callsBack reports whether an IR node calls a function value it's handed.
func callsBack(expression ir.Expression) bool {
	switch expression.(type) {
	case ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.ArraySort, ir.MapForEach:
		return true
	}
	return false
}
