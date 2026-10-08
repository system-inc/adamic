package flow

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// Build makes the graph of one IR function, or of the program's top level when function is -1, and
// finalizes it (reverse postorder, predecessors). It isn't in single assignment form until Construct.
//
// Before Construct, every read and every write of a variable names the same identifier, the
// variable's own: Braun's construction keys on the declaration, and mints a value per definition.
func Build(program *ir.Program, function int) *Function {
	builder := &builder{program: program, identifiers: map[int]IdentifierId{}}
	var statements []ir.Statement
	if function < 0 {
		builder.function = NewFunction("main")
		statements = program.Main
	} else {
		builder.function = NewFunction(program.Functions[function].Name)
		statements = program.Functions[function].Body
	}
	entry := builder.function.NewBlock()
	builder.function.Entry = entry.Id
	builder.current = entry
	if function >= 0 {
		for _, parameter := range program.Functions[function].Parameters {
			if builder.tracked(parameter) {
				builder.function.Params = append(builder.function.Params, builder.place(parameter))
			}
		}
		if count := program.Functions[function].ArgumentsCount; count != 0 && builder.tracked(count-1) {
			builder.function.Params = append(builder.function.Params, builder.place(count-1))
		}

	}
	builder.function.Program = program
	builder.statements(statements)
	// Falling off the end returns, as a void function and the top level do.
	builder.terminate(&Return{})
	Finalize(builder.function)
	return builder.function
}

type builder struct {
	program  *ir.Program
	function *Function

	// current is the block instructions go into now.
	current *BasicBlock

	// identifiers is each tracked local's one pre-construction identifier, by IR local index.
	identifiers map[int]IdentifierId

	// jumps holds, innermost last, where break and continue go from each open loop or switch. A
	// switch has no continue target (InvalidBlock): a continue inside one belongs to a loop outside.
	jumps []jump

	// throwTo holds, innermost last, where a throw goes from here: a catch, a finally, or (when it's
	// empty) the block that throws out of the function, thrown, made when first needed.
	throwTo []BlockId
	thrown  *BasicBlock

	// attempts holds, innermost last, each try whose body or catch is being built: a jump out of one
	// with a finally goes through that finally first.
	attempts []*attempt
}

// attempt is an open try: its finally, if it has one, where that finally may go on to, and how many
// jumps were open when it began (a break or continue to one of those leaves it).
type attempt struct {
	finally   *BasicBlock
	exits     []BlockId
	jumpDepth int
}

type jump struct {
	name                string
	breakTo, continueTo BlockId
}

// tracked reports whether a local is a value the graph follows: one only its own function writes.
func (b *builder) tracked(local int) bool {
	declared := b.program.Locals[local]
	return !declared.Global && !declared.Captured && !declared.ExpressionAssigned
}

// place is a tracked local's place: its one identifier, minted the first time it's met.
func (b *builder) place(local int) Place {
	identifier, ok := b.identifiers[local]
	if !ok {
		identifier = b.function.NewIdentifier(b.program.Locals[local].Name, DeclarationId(local+1)).Id
		b.identifiers[local] = identifier
	}
	return Place{Identifier: identifier}
}

// emit adds an instruction to the current block: part of the statement at at (see Instruction.Part).
// One that can throw ends its block, which goes on to the next or to wherever a throw goes from here.
func (b *builder) emit(at *ir.Statement, part int, expression ir.Expression, uses []Place, defines []Place) {
	instruction := &Instruction{At: at, Part: part, Expression: expression, Uses: uses, Defines: defines}
	b.function.AddInstruction(b.current, instruction)
	if CanThrow(b.program, instruction) {
		next := b.function.NewBlock()
		b.current.Terminal = &MayThrow{Next: next.Id, Handler: b.throwTarget()}
		b.current = next
	}
}

// throwTarget is where a throw goes from here.
func (b *builder) throwTarget() BlockId {
	if len(b.throwTo) > 0 {
		return b.throwTo[len(b.throwTo)-1]
	}
	if b.thrown == nil {
		b.thrown = b.function.NewBlock()
		b.thrown.Terminal = &Throw{}
	}
	return b.thrown.Id
}

// route is where a jump to target goes first when it leaves every try from the one at index on: the
// innermost finally among them, each of which is told to go on to the next one out, and the
// outermost to target.
func (b *builder) route(target BlockId, index int) BlockId {
	next := target
	for ; index < len(b.attempts); index++ {
		if attempt := b.attempts[index]; attempt.finally != nil {
			attempt.exits = append(attempt.exits, next)
			next = attempt.finally.Id
		}
	}
	return next
}

// leaving is the index of the first open try a jump to the jump at depth leaves.
func (b *builder) leaving(depth int) int {
	for index, attempt := range b.attempts {
		if attempt.jumpDepth > depth {
			return index
		}
	}
	return len(b.attempts)
}

// terminate ends the current block and starts a new one with no way in yet: what follows a return,
// a break or a panic is unreachable until something jumps to it, and Finalize drops it if nothing does.
func (b *builder) terminate(terminal Terminal) {
	b.current.Terminal = terminal
	b.current = b.function.NewBlock()
}

// enter ends the current block with a jump to block, and makes block current.
func (b *builder) enter(block *BasicBlock) {
	b.current.Terminal = &Goto{Block: block.Id}
	b.current = block
}

// statements builds each statement where it stands in its slice, so an instruction can name it by
// address: the same address the JavaScript backend's trace names it by.
func (b *builder) statements(statements []ir.Statement) {
	for index := range statements {
		b.statement(&statements[index])
	}
}

func (b *builder) statement(at *ir.Statement) {
	switch statement := (*at).(type) {
	case ir.WriteLine:
		b.emit(at, 0, statement.Value, b.uses(statement.Value), nil)
	case ir.Evaluate:
		b.emit(at, 0, statement.Value, b.uses(statement.Value), nil)
	case ir.SetIndex:
		b.emit(at, 0, nil, b.uses(statement), nil)
	case ir.SetProperty:
		b.emit(at, 0, nil, b.uses(statement), nil)
	case ir.AllocateEnvironment:
		b.emit(at, 0, nil, nil, nil)
	case ir.Declare:
		b.emit(at, 0, statement.Value, b.uses(statement.Value), b.defines(statement.Local))
	case ir.Assign:
		b.emit(at, 0, statement.Value, b.uses(statement.Value), b.defines(statement.Local))
	case ir.Return:
		b.emit(at, 0, statement.Value, b.uses(statement.Value), nil)
		if len(b.attempts) == 0 {
			b.terminate(&Return{})
			break
		}
		// Through every finally around it, then out.
		out := b.function.NewBlock()
		out.Terminal = &Return{}
		b.terminate(&Goto{Block: b.route(out.Id, 0)})
	case ir.Panic:
		b.emit(at, 0, statement.Message, b.uses(statement.Message), nil)
		b.terminate(&Unreachable{})
	case ir.Labeled:
		after := b.function.NewBlock()
		b.jumps = append(b.jumps, jump{name: statement.Name, breakTo: after.Id, continueTo: InvalidBlock})
		b.statements(statement.Body)
		b.enter(after)
		b.jumps = b.jumps[:len(b.jumps)-1]
	case ir.Block:
		b.statements(statement.Body)
	case ir.If:
		b.emit(at, 0, statement.Condition, b.uses(statement.Condition), nil)
		consequent, alternate, after := b.function.NewBlock(), b.function.NewBlock(), b.function.NewBlock()
		b.current.Terminal = &If{Consequent: consequent.Id, Alternate: alternate.Id}
		b.current = consequent
		b.statements(statement.Then)
		b.enter(after)
		b.current = alternate
		b.statements(statement.Else)
		b.enter(after)
	case ir.Loop:
		b.loop(at, statement)
	case ir.ForOf:
		b.forOf(at, statement)
	case ir.Switch:
		b.switchStatement(at, statement)
	case ir.Break:
		depth := len(b.jumps) - 1 - statement.Depth
		if statement.Label != "" {
			for depth >= 0 && b.jumps[depth].name != statement.Label {
				depth--
			}
		}
		b.terminate(&Goto{Block: b.route(b.jumps[depth].breakTo, b.leaving(depth))})
	case ir.Continue:
		for index := len(b.jumps) - 1; index >= 0; index-- {
			if b.jumps[index].continueTo != InvalidBlock && ((statement.Label == "" && b.jumps[index].name == "") || b.jumps[index].name == statement.Label) {
				b.terminate(&Goto{Block: b.route(b.jumps[index].continueTo, b.leaving(index))})
				return
			}
		}
		panic("flow: a continue outside every loop")
	case ir.Throw:
		b.emit(at, 0, statement.Value, b.uses(statement.Value), nil)
		b.terminate(&Goto{Block: b.throwTarget()})
	case ir.Try:
		b.try(at, statement)
	default:
		// Lowering only produces statements this switch knows. A new one has to be taught here, or
		// every analysis over the graph would miss what it reads and writes.
		panic(fmt.Sprintf("flow: no graph for %T", statement))
	}
}

// loop lays out every loop: the test before the body (for, while) or after it (do...while), the
// update after the body, and continue going to the update.
func (b *builder) loop(at *ir.Statement, statement ir.Loop) {
	test, body, update, after := b.function.NewBlock(), b.function.NewBlock(), b.function.NewBlock(), b.function.NewBlock()
	if statement.CheckAfter {
		b.enter(body)
	} else {
		b.enter(test)
	}
	b.linkLabels(statement.Labels, update.Id)
	b.jumps = append(b.jumps, jump{breakTo: after.Id, continueTo: update.Id})
	b.current = body
	b.statements(statement.Body)
	b.enter(update)
	b.statements(statement.Update)
	b.enter(test)
	b.jumps = b.jumps[:len(b.jumps)-1]
	if statement.Condition == nil {
		b.current.Terminal = &Goto{Block: body.Id}
	} else {
		b.emit(at, 0, statement.Condition, b.uses(statement.Condition), nil)
		b.current.Terminal = &If{Consequent: body.Id, Alternate: after.Id}
	}
	b.current = after
}

// forOf evaluates the iterable once, then at the head of each pass either binds the next element
// and runs the body, or leaves. continue goes back to the head.
func (b *builder) forOf(at *ir.Statement, statement ir.ForOf) {
	b.emit(at, 0, statement.Iterable, b.uses(statement.Iterable), nil)
	head, body, after := b.function.NewBlock(), b.function.NewBlock(), b.function.NewBlock()
	b.enter(head)
	head.Terminal = &If{Consequent: body.Id, Alternate: after.Id}
	b.current = body
	var defines []Place
	if statement.Pattern != nil {
		for _, binding := range statement.Pattern {
			defines = append(defines, b.defines(binding.Local)...)
		}
	} else {
		defines = b.defines(statement.Local)
	}
	b.emit(at, 1, nil, nil, defines)
	b.linkLabels(statement.Labels, head.Id)
	b.jumps = append(b.jumps, jump{breakTo: after.Id, continueTo: head.Id})
	b.statements(statement.Body)
	b.jumps = b.jumps[:len(b.jumps)-1]
	b.enter(head)
	b.current = after
}

// switchStatement evaluates the value, then each case's tests in order until one matches, and runs
// that case's body, or the default's when none does. 0.1 has no fallthrough: every body ends by
// leaving the switch.
func (b *builder) switchStatement(at *ir.Statement, statement ir.Switch) {
	b.emit(at, 0, statement.Value, b.uses(statement.Value), nil)
	after := b.function.NewBlock()
	b.jumps = append(b.jumps, jump{breakTo: after.Id})
	bodies := make([]*BasicBlock, len(statement.Cases))
	part := 0
	for index, switchCase := range statement.Cases {
		bodies[index] = b.function.NewBlock()
		for _, test := range switchCase.Tests {
			next := b.function.NewBlock()
			part++
			b.emit(at, part, test, b.uses(test), nil)
			b.current.Terminal = &If{Consequent: bodies[index].Id, Alternate: next.Id}
			b.current = next
		}
	}
	// No test matched: the default, from the block the last test left current.
	b.statements(statement.Default)
	b.enter(after)
	for index, switchCase := range statement.Cases {
		b.current = bodies[index]
		b.statements(switchCase.Body)
		b.enter(after)
	}
	b.jumps = b.jumps[:len(b.jumps)-1]
	b.current = after
}

// defines is the place a write to local defines, or none for a local the graph doesn't track.
func (b *builder) defines(local int) []Place {
	if !b.tracked(local) {
		return nil
	}
	return []Place{b.place(local)}
}

// uses is every tracked local an IR node reads, found by walking every field it holds.
//
// The walk is by reflection rather than a switch over the IR's expressions, which grow with every
// stream's work: a switch that missed a new expression would silently drop its reads, and a missing
// use is a wrong answer from every analysis built on this, while the reflective walk can't miss one.
// It never meets a statement inside an expression, because the IR has none (only statements hold
// statements), so the only reads it can find are ir.Read.
func (b *builder) uses(node any) []Place {
	var places []Place
	var walk func(value reflect.Value)
	walk = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				walk(value.Elem())
			}
		case reflect.Struct:
			if value.Type() == readType {
				if local := int(value.FieldByName("Local").Int()); b.tracked(local) {
					places = append(places, b.place(local))
				}
				return
			}
			for index := 0; index < value.NumField(); index++ {
				walk(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				walk(value.Index(index))
			}
		}
	}
	if node != nil {
		walk(reflect.ValueOf(node))
	}
	return places
}

var readType = reflect.TypeOf(ir.Read{})

// try lays out a try: a throw in its body goes to its catch, or to its finally when it has none; one
// in its catch goes to its finally, or on out; one in its finally goes on out. The catch begins by
// taking the error (part 1). The finally is entered from every way out of the try, and its end goes
// on to wherever any of them was going: after the try, a rethrow out, or a jump or return that left
// it. That's more paths than run, which only makes more live, never less.
func (b *builder) try(at *ir.Statement, statement ir.Try) {
	outerThrow := b.throwTarget()
	after := b.function.NewBlock()
	attempt := &attempt{jumpDepth: len(b.jumps)}
	if statement.HasFinally {
		attempt.finally = b.function.NewBlock()
		attempt.exits = []BlockId{after.Id, outerThrow}
	}
	// ended is where the body and the catch go when they finish.
	ended := after
	if attempt.finally != nil {
		ended = attempt.finally
	}
	var catch *BasicBlock
	bodyThrow := outerThrow
	switch {
	case statement.HasCatch:
		catch = b.function.NewBlock()
		bodyThrow = catch.Id
	case attempt.finally != nil:
		bodyThrow = attempt.finally.Id
	}
	b.attempts = append(b.attempts, attempt)
	b.throwTo = append(b.throwTo, bodyThrow)
	b.statements(statement.Body)
	b.throwTo = b.throwTo[:len(b.throwTo)-1]
	b.enter(ended)
	if catch != nil {
		b.current = catch
		catchThrow := outerThrow
		if attempt.finally != nil {
			catchThrow = attempt.finally.Id
		}
		b.throwTo = append(b.throwTo, catchThrow)
		var defines []Place
		if statement.CatchLocal >= 0 {
			defines = b.defines(statement.CatchLocal)
		}
		b.emit(at, 1, nil, nil, defines)
		b.statements(statement.Catch)
		b.throwTo = b.throwTo[:len(b.throwTo)-1]
		b.enter(ended)
	}
	b.attempts = b.attempts[:len(b.attempts)-1]
	if attempt.finally != nil {
		b.current = attempt.finally
		b.statements(statement.Finally)
		exits := []BlockId{}
		for _, exit := range attempt.exits {
			if !slices.Contains(exits, exit) {
				exits = append(exits, exit)
			}
		}
		b.current.Terminal = &Choose{Blocks: exits}
	}
	b.current = after
}

// CanThrow reports whether an instruction can throw: it calls a function a throw can leave, directly,
// through a function value when one in the program can throw, or as a sort's comparator. A throw
// statement isn't one of these: it always throws, and its block goes straight to its handler.
func CanThrow(program *ir.Program, instruction *Instruction) bool {
	var node any = instruction.Expression
	if instruction.Expression == nil {
		switch statement := (*instruction.At).(type) {
		case ir.SetIndex, ir.SetProperty:
			node = statement
		default:
			return false
		}
	}
	if _, isThrow := (*instruction.At).(ir.Throw); isThrow {
		return false
	}
	throws := false
	var walk func(value reflect.Value)
	walk = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				walk(value.Elem())
			}
		case reflect.Struct:
			if call, ok := value.Interface().(ir.RegExpCall); ok && call.Replacement != nil && program.ClosuresMayThrow {
				throws = true
			}
			if call, ok := value.Interface().(ir.NodeFSFile); ok && call.MayThrow() {
				throws = true
			}
			switch value.Type() {
			case reflect.TypeOf(ir.ArrayHoles{}), reflect.TypeOf(ir.ArraySetLength{}):
				throws = true
			case reflect.TypeOf(ir.ProcessCall{}):
				call := value.Interface().(ir.ProcessCall)
				throws = throws || call.Operation == "exit" || call.Operation == "setExitCode"
			case reflect.TypeOf(ir.NodeHostCall{}):
				throws = throws || value.Interface().(ir.NodeHostCall).Throws
			case reflect.TypeOf(ir.PhantomMember{}):
				if !value.Interface().(ir.PhantomMember).Optional {
					throws = true
				}
			case callType:
				if program.CallMayThrow(value.Interface().(ir.Call)) {
					throws = true
				}
			case callClosureType, arrayMapType, arrayVisitType, arrayReduceType, arrayFromType, mapForEachType:
				// A call through a function value, written out or made by the runtime's loop.
				if program.ClosuresMayThrow {
					throws = true
				}
			case arraySortType:
				sort := value.Interface().(ir.ArraySort)
				if (sort.Callback != nil && program.ClosuresMayThrow) || (sort.Callback == nil && program.Functions[sort.Comparator].MayThrow) {
					throws = true
				}
			}
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
	return throws
}

var (
	callType        = reflect.TypeOf(ir.Call{})
	callClosureType = reflect.TypeOf(ir.CallClosure{})
	arrayMapType    = reflect.TypeOf(ir.ArrayMap{})
	arrayVisitType  = reflect.TypeOf(ir.ArrayVisit{})
	arrayReduceType = reflect.TypeOf(ir.ArrayReduce{})
	arrayFromType   = reflect.TypeOf(ir.ArrayFrom{})
	arraySortType   = reflect.TypeOf(ir.ArraySort{})
	mapForEachType  = reflect.TypeOf(ir.MapForEach{})
)

func (b *builder) linkLabels(names []string, target BlockId) {
	for _, name := range names {
		for index := len(b.jumps) - 1; index >= 0; index-- {
			if b.jumps[index].name == name {
				b.jumps[index].continueTo = target
				break
			}
		}
	}
}
