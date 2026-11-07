package flow

import (
	"maps"
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
	inference := &inference{
		function: function,
		escaped:  map[IdentifierId]bool{},
		iterated: map[*ir.Statement][]Place{},
		parts:    map[[2]uint32]IdentifierId{},
	}
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

func (shape shape) roots() []Place {
	return append(append(append([]Place{}, shape.same...), shape.part...), shape.holds...)
}

func (shape *shape) merge(other shape) {
	shape.same = append(shape.same, other.same...)
	shape.part = append(shape.part, other.part...)
	shape.holds = append(shape.holds, other.holds...)
	shape.fresh = shape.fresh || other.fresh
	shape.unknown = shape.unknown || other.unknown
}

// run makes every instruction's effects, walking the blocks in reverse postorder with the value
// each variable holds.
func (inference *inference) run() {
	exits := map[BlockId]map[DeclarationId]IdentifierId{}
	// thrown is what a block ending in MayThrow hands its handler: the values from before its last
	// instruction, whose definitions a throw never made.
	thrown := map[BlockId]map[DeclarationId]IdentifierId{}
	for _, parameter := range inference.function.Params {
		inference.escaped[parameter.Identifier] = true
	}
	for _, block := range inference.function.Blocks {
		inference.current = map[DeclarationId]IdentifierId{}
		if block.Id == inference.function.Entry {
			for _, parameter := range inference.function.Params {
				inference.current[inference.declaration(parameter)] = parameter.Identifier
			}
		}
		// A variable's value on entry is the same from every predecessor already walked, or there'd
		// be a phi; a predecessor not yet walked is a back edge, whose value a phi carries.
		for _, predecessor := range block.Predecessors {
			from := exits[predecessor]
			if walked, ok := inference.function.Block(predecessor); ok && throwsTo(walked, block.Id) {
				from = thrown[predecessor]
			}
			for variable, value := range from {
				inference.current[variable] = value
			}
		}
		for _, phi := range block.Phis {
			inference.current[inference.declaration(phi.Place)] = phi.Place.Identifier
			for _, operand := range phi.Operands {
				if inference.escaped[operand.Identifier] {
					inference.escaped[phi.Place.Identifier] = true
				}
			}
		}
		_, endsInThrow := block.Terminal.(*MayThrow)
		for position, id := range block.Instructions {
			// run1 changes current only by the instruction's definitions, last, so what it holds now
			// is what a throw out of this instruction leaves.
			if endsInThrow && position == len(block.Instructions)-1 {
				thrown[block.Id] = maps.Clone(inference.current)
			}
			inference.list = nil
			inference.instruction = id
			inference.run1(inference.function.Instructions[id])
			if Suspends(block.Terminal) && position == len(block.Instructions)-1 {
				// Other queued tasks may mutate anything reachable outside this activation.
				inference.mutateEscaped()
			}
			if len(inference.list) > 0 {
				inference.effects.byInstruction[id] = inference.list
			}
		}
		exits[block.Id] = inference.current
	}
}

func (inference *inference) declaration(place Place) DeclarationId {
	return inference.function.Identifiers[place.Identifier].Declaration
}

// run1 makes one instruction's effects: what evaluating it does, then what it defines.
func (inference *inference) run1(instruction *Instruction) {
	switch statement := (*instruction.At).(type) {
	case ir.SetProperty:
		object := inference.value(statement.Object)
		value := inference.value(statement.Value)
		inference.store(object, value)
	case ir.SetIndex:
		array := inference.value(statement.Array)
		inference.value(statement.Index)
		value := inference.value(statement.Value)
		inference.store(array, value)
	case ir.ForOf:
		if instruction.Part == 0 {
			inference.iterated[instruction.At] = inference.value(statement.Iterable).roots()
			break
		}
		// Each element is read out of what the loop iterates, as its first part found it (reverse
		// postorder walks that part first): a mutation through the element is one of that too.
		for _, define := range instruction.Defines {
			inference.define(define, shape{part: inference.iterated[instruction.At]})
		}
	default:
		var defined shape
		if declaration, ok := (*instruction.At).(ir.Declare); ok && instruction.Part == 1 {
			if _, awaited := declaration.Value.(ir.Await); awaited {
				defined.unknown = true
			}
		}
		if instruction.Expression != nil {
			defined = inference.value(instruction.Expression)
		}
		for _, define := range instruction.Defines {
			inference.define(define, defined)
		}
		// A declaration or assignment that defines nothing writes a local the graph doesn't track (a
		// global, a captured variable, one assigned inside an expression), and any call may reach the
		// value through it, so the value escapes.
		switch (*instruction.At).(type) {
		case ir.Declare, ir.Assign:
			if len(instruction.Defines) == 0 {
				inference.escape(defined)
			}
		}
	}
	// What an instruction defines, it defines last: a read of a variable in the same instruction is
	// of its old value, and the new one holds from here on.
	for _, define := range instruction.Defines {
		inference.current[inference.declaration(define)] = define.Identifier
	}
}

// define gives a defined value its effects: created, and related to what its expression's value is.
func (inference *inference) define(into Place, value shape) {
	local := inference.function.Identifiers[into.Identifier].Declaration - 1
	if !mutable(inference.function.Program.Locals[local].Type) {
		inference.list = append(inference.list, create(into, EffectValuePrimitive))
		return
	}
	if len(value.part) == 1 && len(value.same) == 0 && len(value.holds) == 0 && !value.fresh && !value.unknown {
		inference.list = append(inference.list, flow(AliasingEffectCreateFrom, value.part[0], into))
	} else {
		inference.list = append(inference.list, create(into, EffectValueMutable))
		for _, from := range value.same {
			inference.list = append(inference.list, flow(AliasingEffectAssign, from, into))
		}
		// A value read out of a container, mixed with others (a ?? b, c ? d : e), goes through a
		// temporary created from the container, which is what React's graph has for every read:
		// only CreateFrom carries a mutation back to the container transitively, on to whatever the
		// container holds, and an Alias straight from the container wouldn't.
		for _, from := range value.part {
			key := [2]uint32{uint32(inference.instruction), uint32(from.Identifier)}
			part, ok := inference.parts[key]
			if !ok {
				part = inference.function.NewIdentifier("", 0).Id
				inference.parts[key] = part
			}
			inference.list = append(inference.list, flow(AliasingEffectCreateFrom, from, Place{Identifier: part}))
			inference.list = append(inference.list, flow(AliasingEffectAssign, Place{Identifier: part}, into))
			if inference.escaped[from.Identifier] {
				inference.escaped[part] = true
			}
		}
		for _, from := range value.holds {
			inference.list = append(inference.list, flow(AliasingEffectCapture, from, into))
		}
	}
	if value.unknown {
		inference.escaped[into.Identifier] = true
	}
	for _, root := range value.roots() {
		if inference.escaped[root.Identifier] {
			inference.escaped[into.Identifier] = true
		}
	}
}

// store is a write into a container: the container is mutated, and holds what was written.
func (inference *inference) store(container shape, value shape) {
	for _, root := range container.roots() {
		inference.list = append(inference.list, mutate(AliasingEffectMutate, root))
		for _, from := range value.roots() {
			inference.list = append(inference.list, flow(AliasingEffectCapture, from, root))
		}
	}
	if container.unknown || anyEscaped(inference.escaped, container.roots()) {
		inference.escape(value)
		// An escaped container may be reachable from any other escaped value (two parameters may
		// be the same object, or one may hold the other), so writing into it may change them all.
		inference.mutateEscaped()
	}
}

// call is what a call does to what it's handed and to every escaped value: mutates them,
// conditionally and transitively. Its result is unknown, and may be any of them.
func (inference *inference) call(operands []shape) shape {
	result := shape{unknown: true}
	for _, operand := range operands {
		inference.escape(operand)
	}
	inference.mutateEscaped()
	for _, operand := range operands {
		result.same = append(result.same, operand.roots()...)
	}
	return result
}

// mutateEscaped mutates, conditionally and transitively, every escaped value that reaches here: what
// a call can reach, or what a write into an escaped value may change.
func (inference *inference) mutateEscaped() {
	// In a fixed order, since an effect's position is the alias graph's time.
	variables := make([]DeclarationId, 0, len(inference.current))
	for variable := range inference.current {
		variables = append(variables, variable)
	}
	slices.Sort(variables)
	for _, variable := range variables {
		value := inference.current[variable]
		if inference.escaped[value] && mutable(inference.typeOf(value)) {
			inference.list = append(
				inference.list,
				mutate(AliasingEffectMutateTransitiveConditionally, Place{Identifier: value}),
			)
		}
	}
}

func (inference *inference) escape(value shape) {
	for _, root := range value.roots() {
		inference.escaped[root.Identifier] = true
	}
}

func (inference *inference) typeOf(value IdentifierId) ir.Type {
	return inference.function.Program.Locals[inference.function.Identifiers[value].Declaration-1].Type
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
func (inference *inference) value(expression ir.Expression) shape {
	if expression == nil {
		return shape{}
	}
	switch expression := expression.(type) {
	case ir.TypedArrayNew:
		inference.value(expression.Source)
		return shape{fresh: true}
	case ir.TypedArraySubarray:
		// Views share storage. Conservatively alias the parent for every mutation.
		array := inference.value(expression.Array)
		for _, argument := range expression.Arguments {
			inference.value(argument)
		}
		return array
	case ir.TypedArraySort:
		array := inference.value(expression.Array)
		inference.store(array, shape{})
		return array
	case ir.TypedArrayFill:
		array := inference.value(expression.Array)
		for _, argument := range expression.Arguments {
			inference.value(argument)
		}
		inference.store(array, shape{})
		return array
	case ir.TypedArraySet:
		array := inference.value(expression.Array)
		for _, argument := range expression.Arguments {
			inference.value(argument)
		}
		inference.store(array, shape{})
		return shape{}
	case ir.Read:
		declared := inference.function.Program.Locals[expression.Local]
		if !mutable(expression.Of) {
			return shape{}
		}
		// A local the graph doesn't track holds whatever was last written into it, which may be any
		// escaped value.
		if declared.Global || declared.Captured || declared.ExpressionAssigned {
			return shape{unknown: true}
		}
		value, ok := inference.current[DeclarationId(expression.Local+1)]
		if !ok {
			return shape{}
		}
		return shape{same: []Place{{Identifier: value}}}
	case ir.Void:
		inference.value(expression.Value)
		return shape{}
	case ir.Comma:
		inference.value(expression.Left)
		return inference.value(expression.Right)
	case ir.Logical:
		result := inference.value(expression.Left)
		result.merge(inference.value(expression.Right))
		return result
	case ir.Conditional:
		inference.value(expression.Condition)
		result := inference.value(expression.WhenTrue)
		result.merge(inference.value(expression.WhenNot))
		return result
	case ir.Coalesce:
		result := inference.value(expression.Value)
		result.merge(inference.value(expression.Fallback))
		inference.value(expression.Panic)
		return result
	case ir.Box:
		return inference.value(expression.Value)
	case ir.Narrow:
		return inference.value(expression.Value)
	case ir.PlaceholderUse:
		// A checked use forwards the value; it cannot create a reference or a write.
		return inference.value(expression.Value)
	case ir.Unwrap:
		return inference.value(expression.Value)
	case ir.CheckedCast:
		result := inference.value(expression.Value)
		for _, allowed := range expression.Allowed {
			inference.value(allowed)
		}
		return result
	case ir.RecordCoalesce:
		holder := inference.value(expression.Record)
		inference.value(expression.Key)
		result := inference.value(expression.Value)
		inference.store(holder, result)
		result.part = append(result.part, holder.roots()...)
		return result
	case ir.RecordCall:
		operands := inference.operands(expression)
		switch expression.Method {
		case "set":
			inference.store(operands[0], operands[2])
			return operands[2]
		case "defineHidden":
			inference.store(operands[0], operands[2])
			return shape{}
		case "delete":
			inference.store(operands[0], shape{})
			return shape{}
		case "get":
			return shape{part: operands[0].roots()}
		case "values", "entries":
			return shape{fresh: true, holds: operands[0].roots()}
		case "keys":
			return shape{fresh: true}
		case "has", "hasOwn":
			return shape{}
		}
		// A method this doesn't know yet is the most it could be, as a node is below.
		return inference.call(operands)
	case ir.RecordLiteral:
		result := shape{fresh: true}
		if expression.Spread != nil {
			result.holds = append(result.holds, inference.value(expression.Spread).roots()...)
		}
		for _, entry := range expression.Entries {
			inference.value(entry.Key)
			result.holds = append(result.holds, inference.value(entry.Value).roots()...)
		}
		return result
	case ir.ObjectLiteral:
		result := shape{fresh: true}
		if expression.Spread != nil {
			// A copy of the spread's object: its fields are the same values the spread holds.
			result.holds = append(result.holds, inference.value(expression.Spread).roots()...)
		}
		for _, field := range expression.Fields {
			result.holds = append(result.holds, inference.value(field.Value).roots()...)
		}
		return result
	case ir.ArrayLiteral:
		result := shape{fresh: true}
		for _, element := range expression.Elements {
			result.holds = append(result.holds, inference.value(element).roots()...)
		}
		return result
	case ir.Property:
		return shape{part: inference.value(expression.Object).roots()}
	case ir.ArrayIndex:
		array := inference.value(expression.Array)
		inference.value(expression.Index)
		return shape{part: array.roots()}
	case ir.ArrayPush:
		array := inference.value(expression.Array)
		value := inference.value(expression.Value)
		inference.store(array, value)
		return shape{}
	case ir.ArrayPop:
		array := inference.value(expression.Array)
		inference.store(array, shape{})
		return shape{part: array.roots()}
	case ir.ArrayReverse:
		array := inference.value(expression.Array)
		inference.store(array, shape{})
		return array
	case ir.MakeClosure:
		// Its cells hold captured variables, which this graph doesn't track.
		return shape{fresh: true, unknown: true}
	case ir.Call:
		operands := []shape{}
		for _, argument := range expression.Arguments {
			operands = append(operands, inference.value(argument))
		}
		return inference.call(operands)
	case ir.CallClosure:
		// Every target, bounded or Unknown, has the same conservative effect here:
		// anything it can reach may change. Evaluate the callee as an operand too.
		return inference.call(inference.operands(expression))
	}
	// Everything else: the effects of what it evaluates, field by field, and then, as cohere assumes
	// of a call it has no signature for, that it mutates and may return any of them. That covers
	// the methods that call a callback (map, forEach, a Map's or Set's forEach, reduce, sort, Array.from), the ones
	// that write (splice, fill, set, delete, clear), and every node whose value can be mutated. Only a node
	// that does none of these gives nothing, and writes_test.go fails on a node added to the IR
	// until it's named as one of them, so a new write can't be taken for nothing by its type.
	operands := inference.operands(expression)
	if !mutable(expression.Type()) && !writes(expression) && !callsBack(expression) {
		return shape{}
	}
	return inference.call(operands)
}

// statement makes the effects of a statement an expression holds (an Effects' body, a logical
// assignment's write) as run1 makes a statement's, a write into a field or an element being a
// store, and is the shapes of the expressions it holds, as operands is.
func (inference *inference) statement(statement ir.Statement) []shape {
	switch statement := statement.(type) {
	case ir.SetProperty:
		object := inference.value(statement.Object)
		value := inference.value(statement.Value)
		inference.store(object, value)
		return []shape{object, value}
	case ir.SetIndex:
		array := inference.value(statement.Array)
		index := inference.value(statement.Index)
		value := inference.value(statement.Value)
		inference.store(array, value)
		return []shape{array, index, value}
	}
	return inference.operands(statement)
}

// operands evaluates every expression an IR node holds, in the order its fields are declared. A
// statement it holds is made by statement, so a write among an Effects' statements, or a logical
// assignment's, is a store.
func (inference *inference) operands(node any) []shape {
	var operands []shape
	var walk func(value reflect.Value)
	walk = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			if inner, ok := value.Interface().(ir.Expression); ok {
				operands = append(operands, inference.value(inner))
				return
			}
			if inner, ok := value.Interface().(ir.Statement); ok {
				operands = append(operands, inference.statement(inner)...)
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
	walk(reflect.ValueOf(node))
	return operands
}

// writes reports whether an IR node writes into a supplied container or live process state. It's asked before the
// node's type is, since a write can give something that can't be mutated (a clear is void, a
// readSync is a count). writes_test.go fails on an IR expression it doesn't find classified: here,
// in callsBack, in a case of value's own, as holding statements, or as writing nothing.
func writes(expression ir.Expression) bool {
	switch call := expression.(type) {
	case ir.NodeBufferCall:
		return call.Function == "buffer_set" || call.Function == "hash_update" || call.Function == "hash_digest"
	case ir.NodeFSFile:
		// readSync writes its buffer; environment mutations write live process state.
		return call.Operation == "read_sync" || call.Operation == "host_env_set" || call.Operation == "host_env_delete"
	case ir.ObjectCall:
		return call.Method == "assign" || call.Method == "freeze"
	case ir.RegExpCall:
		// A pattern's lastIndex: exec, test, match and replace move or reset it, search sets it and
		// puts it back, and an iterator's next moves its own. split and matchAll work on a copy of
		// the pattern, and iteratorDone reads a result's done.
		return call.Method != "split" && call.Method != "matchAll" && call.Method != "iteratorDone"
	case ir.ArraySplice, ir.ArrayFill, ir.ArraySort, ir.MapSet, ir.MapDelete, ir.MapClear, ir.SetAdd:
		return true
	case ir.GeneratorNext:
		// A step moves the generator it's handed on, and its body may write what the frame holds.
		return true
	}
	return false
}

// callsBack reports whether an IR node calls a function value it's handed.
func callsBack(expression ir.Expression) bool {
	if call, ok := expression.(ir.NodeFSFile); ok && call.Operation == "host_watch_file" {
		return true
	}
	if call, ok := expression.(ir.RegExpCall); ok {
		return call.Replacement != nil
	}
	switch expression.(type) {
	// A run-time key read calls whichever class getter its key names.
	case ir.DynamicKeyedProperty:
		return true
	// Fail closed for encoding: lowering refuses getter-bearing operands;
	// schema field reads must never introduce an untracked callback.
	case ir.JSONEncode, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.ArraySort, ir.MapForEach, ir.ParallelMap:
		return true
	// A step runs the generator's body, which may call anything.
	case ir.GeneratorNext:
		return true
	}
	return false
}
