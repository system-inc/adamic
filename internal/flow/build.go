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
	// The pending shared-SSA switch owns general ordering. Extend its terminal adapter here.
	order := EvaluationOrder(1)
	for _, block := range builder.function.Blocks {
		order += EvaluationOrder(len(block.Instructions))
		switch terminal := block.Terminal.(type) {
		case *Suspend:
			terminal.Order = order
		case *Yield:
			terminal.Order = order
		}
		order++
	}
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
func (builder *builder) tracked(local int) bool {
	declared := builder.program.Locals[local]
	return !declared.Global && !declared.Captured && !declared.ExpressionAssigned
}

// place is a tracked local's place: its one identifier, minted the first time it's met.
func (builder *builder) place(local int) Place {
	identifier, ok := builder.identifiers[local]
	if !ok {
		identifier = builder.function.NewIdentifier(builder.program.Locals[local].Name, DeclarationId(local+1)).Id
		builder.identifiers[local] = identifier
	}
	return Place{Identifier: identifier}
}

// emit adds an instruction to the current block: part of the statement at at (see Instruction.Part).
// One that can throw ends its block, which goes on to the next or to wherever a throw goes from here.
func (builder *builder) emit(at *ir.Statement, part int, expression ir.Expression, uses []Place, defines []Place) {
	instruction := &Instruction{At: at, Part: part, Expression: expression, Uses: uses, Defines: defines}
	builder.function.AddInstruction(builder.current, instruction)
	if CanThrow(builder.program, instruction) {
		next := builder.function.NewBlock()
		builder.current.Terminal = &MayThrow{Next: next.Id, Handler: builder.throwTarget()}
		builder.current = next
	}
}

// throwTarget is where a throw goes from here.
func (builder *builder) throwTarget() BlockId {
	if len(builder.throwTo) > 0 {
		return builder.throwTo[len(builder.throwTo)-1]
	}
	if builder.thrown == nil {
		builder.thrown = builder.function.NewBlock()
		builder.thrown.Terminal = &Throw{}
	}
	return builder.thrown.Id
}

// route is where a jump to target goes first when it leaves every try from the one at index on: the
// innermost finally among them, each of which is told to go on to the next one out, and the
// outermost to target.
func (builder *builder) route(target BlockId, index int) BlockId {
	next := target
	for ; index < len(builder.attempts); index++ {
		if attempt := builder.attempts[index]; attempt.finally != nil {
			attempt.exits = append(attempt.exits, next)
			next = attempt.finally.Id
		}
	}
	return next
}

// leaving is the index of the first open try a jump to the jump at depth leaves.
func (builder *builder) leaving(depth int) int {
	for index, attempt := range builder.attempts {
		if attempt.jumpDepth > depth {
			return index
		}
	}
	return len(builder.attempts)
}

// terminate ends the current block and starts a new one with no way in yet: what follows a return,
// a break or a panic is unreachable until something jumps to it, and Finalize drops it if nothing does.
func (builder *builder) terminate(terminal Terminal) {
	builder.current.Terminal = terminal
	builder.current = builder.function.NewBlock()
}

// enter ends the current block with a jump to block, and makes block current.
func (builder *builder) enter(block *BasicBlock) {
	builder.current.Terminal = &Goto{Block: block.Id}
	builder.current = block
}

// statements builds each statement where it stands in its slice, so an instruction can name it by
// address: the same address the JavaScript backend's trace names it by.
func (builder *builder) statements(statements []ir.Statement) {
	for index := range statements {
		builder.statement(&statements[index])
	}
}

func (builder *builder) statement(at *ir.Statement) {
	switch statement := (*at).(type) {
	case ir.SourceLine:
		// Source metadata emits no instruction and keeps the path.
	case ir.Debugger:
		// With no debugger attached, it reads and writes nothing and keeps the path.
		return
	case ir.WriteLine:
		builder.emit(at, 0, statement.Value, builder.uses(statement.Value), nil)
	case ir.Evaluate:
		if wait, ok := statement.Value.(ir.Await); ok {
			builder.suspend(at, wait, -1)
			return
		}
		if yield, ok := statement.Value.(ir.Yield); ok {
			builder.yield(at, yield)
			return
		}
		builder.emit(at, 0, statement.Value, builder.uses(statement.Value), nil)
	case ir.SetIndex:
		builder.emit(at, 0, nil, builder.uses(statement), nil)
	case ir.SetProperty:
		builder.emit(at, 0, nil, builder.uses(statement), nil)
	case ir.AllocateEnvironment:
		builder.emit(at, 0, nil, nil, nil)
	case ir.Declare:
		if wait, ok := statement.Value.(ir.Await); ok {
			builder.suspend(at, wait, statement.Local)
			return
		}
		builder.emit(at, 0, statement.Value, builder.uses(statement.Value), builder.defines(statement.Local))
	case ir.Assign:
		builder.emit(at, 0, statement.Value, builder.uses(statement.Value), builder.defines(statement.Local))
	case ir.Return:
		builder.emit(at, 0, statement.Value, builder.uses(statement.Value), nil)
		if !slices.ContainsFunc(builder.attempts, func(attempt *attempt) bool { return attempt.finally != nil }) {
			builder.terminate(&Return{})
			break
		}
		// Through every finally around it, then out.
		out := builder.function.NewBlock()
		out.Terminal = &Return{}
		builder.terminate(&Goto{Block: builder.route(out.Id, 0)})
	case ir.Panic:
		builder.emit(at, 0, statement.Message, builder.uses(statement.Message), nil)
		builder.terminate(&Unreachable{})
	case ir.Labeled:
		after := builder.function.NewBlock()
		builder.jumps = append(builder.jumps, jump{name: statement.Name, breakTo: after.Id, continueTo: InvalidBlock})
		builder.statements(statement.Body)
		builder.enter(after)
		builder.jumps = builder.jumps[:len(builder.jumps)-1]
	case ir.Block:
		builder.statements(statement.Body)
	case ir.If:
		builder.emit(at, 0, statement.Condition, builder.uses(statement.Condition), nil)
		consequent, alternate, after := builder.function.NewBlock(), builder.function.NewBlock(), builder.function.NewBlock()
		builder.current.Terminal = &If{Consequent: consequent.Id, Alternate: alternate.Id}
		builder.current = consequent
		builder.statements(statement.Then)
		builder.enter(after)
		builder.current = alternate
		builder.statements(statement.Else)
		builder.enter(after)
	case ir.Loop:
		builder.loop(at, statement)
	case ir.ForOf:
		builder.forOf(at, statement)
	case ir.Switch:
		builder.switchStatement(at, statement)
	case ir.Break:
		depth := len(builder.jumps) - 1
		remaining := statement.Depth
		for depth >= 0 {
			target := builder.jumps[depth]
			if statement.Label != "" {
				if target.name == statement.Label {
					break
				}
			} else if target.name == "" {
				if remaining == 0 {
					break
				}
				remaining--
			}
			depth--
		}
		if depth < 0 {
			panic("flow: unknown break target")
		}
		builder.terminate(&Goto{Block: builder.route(builder.jumps[depth].breakTo, builder.leaving(depth))})
	case ir.Continue:
		for index := len(builder.jumps) - 1; index >= 0; index-- {
			if builder.jumps[index].continueTo != InvalidBlock &&
				((statement.Label == "" && builder.jumps[index].name == "") ||
					builder.jumps[index].name == statement.Label) {
				builder.terminate(&Goto{Block: builder.route(builder.jumps[index].continueTo, builder.leaving(index))})
				return
			}
		}
		panic("flow: a continue outside every loop")
	case ir.Throw:
		builder.emit(at, 0, statement.Value, builder.uses(statement.Value), nil)
		builder.terminate(&Goto{Block: builder.throwTarget()})
	case ir.Try:
		builder.try(at, statement)
	default:
		// Lowering only produces statements this switch knows. A new one has to be taught here, or
		// every analysis over the graph would miss what it reads and writes.
		panic(fmt.Sprintf("flow: no graph for %T", statement))
	}
}

// loop lays out every loop: the test before the body (for, while) or after it (do...while), the
// update after the body, and continue going to the update.
func (builder *builder) loop(at *ir.Statement, statement ir.Loop) {
	test, body, update, after := builder.function.NewBlock(), builder.function.NewBlock(), builder.function.NewBlock(), builder.function.NewBlock()
	if statement.CheckAfter {
		builder.enter(body)
	} else {
		builder.enter(test)
	}
	builder.linkLabels(statement.Labels, update.Id)
	builder.jumps = append(builder.jumps, jump{breakTo: after.Id, continueTo: update.Id})
	builder.current = body
	builder.statements(statement.Body)
	builder.enter(update)
	builder.statements(statement.Update)
	builder.enter(test)
	builder.jumps = builder.jumps[:len(builder.jumps)-1]
	builder.statements(statement.Test)
	if statement.Condition == nil {
		builder.current.Terminal = &Goto{Block: body.Id}
	} else {
		builder.emit(at, 0, statement.Condition, builder.uses(statement.Condition), nil)
		builder.current.Terminal = &If{Consequent: body.Id, Alternate: after.Id}
	}
	builder.current = after
}

// forOf evaluates the iterable once, then at the head of each pass either binds the next element
// and runs the body, or leaves. continue goes back to the head.
func (builder *builder) forOf(at *ir.Statement, statement ir.ForOf) {
	builder.emit(at, 0, statement.Iterable, builder.uses(statement.Iterable), nil)
	head, body, after := builder.function.NewBlock(), builder.function.NewBlock(), builder.function.NewBlock()
	builder.enter(head)
	head.Terminal = &If{Consequent: body.Id, Alternate: after.Id}
	builder.current = body
	var defines []Place
	if statement.Pattern != nil {
		for _, binding := range statement.Pattern {
			defines = append(defines, builder.defines(binding.Local)...)
		}
	} else {
		defines = builder.defines(statement.Local)
	}
	builder.emit(at, 1, nil, nil, defines)
	builder.linkLabels(statement.Labels, head.Id)
	builder.jumps = append(builder.jumps, jump{breakTo: after.Id, continueTo: head.Id})
	builder.statements(statement.Body)
	builder.jumps = builder.jumps[:len(builder.jumps)-1]
	builder.enter(head)
	builder.current = after
}

// switchStatement evaluates the value, then each case's tests in order until one matches, and runs
// that case's body, or the default's when none does. 0.1 has no fallthrough: every body ends by
// leaving the switch.
func (builder *builder) switchStatement(at *ir.Statement, statement ir.Switch) {
	builder.emit(at, 0, statement.Value, builder.uses(statement.Value), nil)
	after := builder.function.NewBlock()
	builder.jumps = append(builder.jumps, jump{breakTo: after.Id})
	bodies := make([]*BasicBlock, len(statement.Cases))
	part := 0
	for index, switchCase := range statement.Cases {
		bodies[index] = builder.function.NewBlock()
		for _, testCase := range switchCase.Tests {
			next := builder.function.NewBlock()
			part++
			builder.emit(at, part, testCase, builder.uses(testCase), nil)
			builder.current.Terminal = &If{Consequent: bodies[index].Id, Alternate: next.Id}
			builder.current = next
		}
	}
	// No test matched: the default, from the block the last test left current.
	builder.statements(statement.Default)
	builder.enter(after)
	for index, switchCase := range statement.Cases {
		builder.current = bodies[index]
		builder.statements(switchCase.Body)
		builder.enter(after)
	}
	builder.jumps = builder.jumps[:len(builder.jumps)-1]
	builder.current = after
}

// defines is the place a write to local defines, or none for a local the graph doesn't track.
func (builder *builder) defines(local int) []Place {
	if !builder.tracked(local) {
		return nil
	}
	return []Place{builder.place(local)}
}

// uses is every tracked local an IR node reads, found by walking every field it holds.
//
// The walk is by reflection rather than a switch over the IR's expressions, which grow with every
// stream's work: a switch that missed a new expression would silently drop its reads, and a missing
// use is a wrong answer from every analysis built on this, while the reflective walk can't miss one.
// It never meets a statement inside an expression, because the IR has none (only statements hold
// statements), so the only reads it can find are ir.Read.
func (builder *builder) uses(node any) []Place {
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
				if local := int(value.FieldByName("Local").Int()); builder.tracked(local) {
					places = append(places, builder.place(local))
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
func (builder *builder) try(at *ir.Statement, statement ir.Try) {
	outerThrow := builder.throwTarget()
	after := builder.function.NewBlock()
	attempt := &attempt{jumpDepth: len(builder.jumps)}
	if statement.HasFinally {
		attempt.finally = builder.function.NewBlock()
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
		catch = builder.function.NewBlock()
		bodyThrow = catch.Id
	case attempt.finally != nil:
		bodyThrow = attempt.finally.Id
	}
	builder.attempts = append(builder.attempts, attempt)
	builder.throwTo = append(builder.throwTo, bodyThrow)
	builder.statements(statement.Body)
	builder.throwTo = builder.throwTo[:len(builder.throwTo)-1]
	builder.enter(ended)
	if catch != nil {
		builder.current = catch
		catchThrow := outerThrow
		if attempt.finally != nil {
			catchThrow = attempt.finally.Id
		}
		builder.throwTo = append(builder.throwTo, catchThrow)
		var defines []Place
		if statement.CatchLocal >= 0 {
			defines = builder.defines(statement.CatchLocal)
		}
		builder.emit(at, 1, nil, nil, defines)
		builder.statements(statement.Catch)
		builder.throwTo = builder.throwTo[:len(builder.throwTo)-1]
		builder.enter(ended)
	}
	builder.attempts = builder.attempts[:len(builder.attempts)-1]
	if attempt.finally != nil {
		builder.current = attempt.finally
		builder.statements(statement.Finally)
		exits := []BlockId{}
		for _, exit := range attempt.exits {
			if !slices.Contains(exits, exit) {
				exits = append(exits, exit)
			}
		}
		builder.current.Terminal = &Choose{Blocks: exits}
	}
	builder.current = after
}

// CanThrow reports whether an instruction can throw: it calls a function a throw can leave, directly,
// through a function value when one in the program can throw, or as a sort's comparator, or it reads
// or writes a binding in its temporal dead zone (ir.Program's DeadZoneMayThrow). A throw
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
	// An assignment's own dead-zone test runs after its value, as JavaScript's store throws once the
	// right side has run (#bxmdk73), and the instruction's Expression is only that value. The throw
	// edge sees the state before the instruction's define, as every throw edge does.
	if assign, isAssign := (*instruction.At).(ir.Assign); isAssign && program.DeadZoneMayThrow(assign) {
		return true
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
			if literal, ok := value.Interface().(ir.ObjectLiteral); ok && program.ObjectSpreadMayThrow(literal) {
				throws = true
			}
			if call, ok := value.Interface().(ir.RegExpCall); ok &&
				call.Replacement != nil &&
				program.ClosuresMayThrow {
				throws = true
			}
			if _, ok := value.Interface().(ir.ParallelMap); ok && program.ClosuresMayThrow {
				throws = true
			}
			if call, ok := value.Interface().(ir.NodeFSFile); ok && call.MayThrow() {
				throws = true
			}
			if call, ok := value.Interface().(ir.ToFixed); ok && call.MayThrow() {
				throws = true
			}
			if call, ok := value.Interface().(ir.NumberFormat); ok && call.MayThrow() {
				throws = true
			}
			if call, ok := value.Interface().(ir.ArrayFill); ok && call.MayThrow() {
				throws = true
			}
			if call, ok := value.Interface().(ir.ArrayFrom); ok && call.MayThrow() {
				throws = true
			}
			if call, ok := value.Interface().(ir.StringCall); ok && call.MayThrow(program.Strings) {
				throws = true
			}
			if call, ok := value.Interface().(ir.StringFromCodes); ok && call.MayThrow() {
				throws = true
			}
			if program.RecordWriteMayThrow(value.Interface()) || program.FrozenWriteMayThrow(value.Interface()) {
				throws = true
			}
			if call, ok := value.Interface().(ir.RegExpCall); ok && program.GlobalFlagMayThrow(call) {
				throws = true
			}
			// A read in its temporal dead zone, or a logical assignment's write, throws JavaScript's
			// ReferenceError (#bxmdk73).
			if program.DeadZoneMayThrow(value.Interface()) {
				throws = true
			}
			switch value.Type() {
			case reflect.TypeOf(ir.Defined{}):
				if value.Interface().(ir.Defined).Throws() {
					throws = true
				}
			case reflect.TypeOf(ir.PhantomMember{}):
				throws = throws || !value.Interface().(ir.PhantomMember).Optional
			case reflect.TypeOf(ir.ProcessCall{}):
				operation := value.Interface().(ir.ProcessCall).Operation
				throws = throws ||
					operation == "exit" ||
					operation == "setExitCode" ||
					operation == "cwd" ||
					operation == "chdir" ||
					operation == "measure"
			case reflect.TypeOf(ir.NodeHostCall{}):
				throws = throws || value.Interface().(ir.NodeHostCall).Throws
			case callType:
				if program.CallMayThrow(value.Interface().(ir.Call)) {
					throws = true
				}
			case generatorNextType:
				// A step runs the generator's body, which a throw can leave.
				if program.GeneratorNextMayThrow(value.Interface().(ir.GeneratorNext)) {
					throws = true
				}
			case callClosureType, arrayMapType, arrayVisitType, arrayReduceType, arrayFromType, mapForEachType:
				// A call through a function value, written out or made by the runtime's loop.
				if program.ClosureMayThrow(value.Interface().(ir.Expression)) {
					throws = true
				}
			case arraySortType:
				sort := value.Interface().(ir.ArraySort)
				if (sort.Callback != nil && program.ClosuresMayThrow) ||
					(sort.Callback == nil && program.Functions[sort.Comparator].MayThrow) {
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

	generatorNextType = reflect.TypeOf(ir.GeneratorNext{})
)

func (builder *builder) linkLabels(names []string, target BlockId) {
	for _, name := range names {
		for index := len(builder.jumps) - 1; index >= 0; index-- {
			if builder.jumps[index].name == name {
				builder.jumps[index].continueTo = target
				break
			}
		}
	}
}

func (builder *builder) suspend(at *ir.Statement, wait ir.Await, local int) {
	builder.emit(at, 0, wait.Value, builder.uses(wait.Value), nil)
	fulfilled := builder.function.NewBlock()
	builder.current.Terminal = &Suspend{
		Fulfilled: fulfilled.Id,
		Rejected:  builder.throwTarget(),
		Local:     local,
		Of:        wait.Of,
	}
	builder.current = fulfilled
	if local >= 0 {
		builder.function.AddInstruction(builder.current, &Instruction{At: at, Part: 1, Defines: builder.defines(local)})
	}
}

// yield evaluates the yielded value as its block's last instruction, then leaves for the consumer;
// the next step goes on in a new block. The emitter reads the value from that last instruction, so
// one that could throw would end its block first and leave the yield with nothing to hand over:
// normalization makes the value a read of its own snapshot, which can't.
func (builder *builder) yield(at *ir.Statement, yield ir.Yield) {
	builder.emit(at, 0, yield.Value, builder.uses(yield.Value), nil)
	if len(builder.current.Instructions) == 0 {
		panic("flow: a yield whose value can throw was not normalized")
	}
	next := builder.function.NewBlock()
	builder.current.Terminal = &Yield{Next: next.Id}
	builder.current = next
}
