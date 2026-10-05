package flow

import (
	"fmt"
	"reflect"

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
	}
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
}

type jump struct {
	breakTo, continueTo BlockId
}

// tracked reports whether a local is a value the graph follows: one only its own function writes.
func (b *builder) tracked(local int) bool {
	declared := b.program.Locals[local]
	return !declared.Global && !declared.Captured
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

// emit adds an instruction to the current block.
func (b *builder) emit(statement ir.Statement, expression ir.Expression, uses []Place, defines []Place) {
	b.function.AddInstruction(b.current, &Instruction{Statement: statement, Expression: expression, Uses: uses, Defines: defines})
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

func (b *builder) statements(statements []ir.Statement) {
	for _, statement := range statements {
		b.statement(statement)
	}
}

func (b *builder) statement(statement ir.Statement) {
	switch statement := statement.(type) {
	case ir.WriteLine:
		b.emit(statement, statement.Value, b.uses(statement.Value), nil)
	case ir.Evaluate:
		b.emit(statement, statement.Value, b.uses(statement.Value), nil)
	case ir.SetIndex:
		b.emit(statement, nil, b.uses(statement), nil)
	case ir.SetProperty:
		b.emit(statement, nil, b.uses(statement), nil)
	case ir.Declare:
		b.emit(statement, statement.Value, b.uses(statement.Value), b.defines(statement.Local))
	case ir.Assign:
		b.emit(statement, statement.Value, b.uses(statement.Value), b.defines(statement.Local))
	case ir.Return:
		b.emit(statement, statement.Value, b.uses(statement.Value), nil)
		b.terminate(&Return{})
	case ir.Panic:
		b.emit(statement, statement.Message, b.uses(statement.Message), nil)
		b.terminate(&Unreachable{})
	case ir.Block:
		b.statements(statement.Body)
	case ir.If:
		b.emit(statement, statement.Condition, b.uses(statement.Condition), nil)
		consequent, alternate, after := b.function.NewBlock(), b.function.NewBlock(), b.function.NewBlock()
		b.current.Terminal = &If{Consequent: consequent.Id, Alternate: alternate.Id}
		b.current = consequent
		b.statements(statement.Then)
		b.enter(after)
		b.current = alternate
		b.statements(statement.Else)
		b.enter(after)
	case ir.Loop:
		b.loop(statement)
	case ir.ForOf:
		b.forOf(statement)
	case ir.Switch:
		b.switchStatement(statement)
	case ir.Break:
		b.terminate(&Goto{Block: b.jumps[len(b.jumps)-1].breakTo})
	case ir.Continue:
		for index := len(b.jumps) - 1; index >= 0; index-- {
			if b.jumps[index].continueTo != InvalidBlock {
				b.terminate(&Goto{Block: b.jumps[index].continueTo})
				return
			}
		}
		panic("flow: a continue outside every loop")
	default:
		// Lowering only produces statements this switch knows. A new one has to be taught here, or
		// every analysis over the graph would miss what it reads and writes.
		panic(fmt.Sprintf("flow: no graph for %T", statement))
	}
}

// loop lays out every loop: the test before the body (for, while) or after it (do...while), the
// update after the body, and continue going to the update.
func (b *builder) loop(statement ir.Loop) {
	test, body, update, after := b.function.NewBlock(), b.function.NewBlock(), b.function.NewBlock(), b.function.NewBlock()
	if statement.CheckAfter {
		b.enter(body)
	} else {
		b.enter(test)
	}
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
		b.emit(statement, statement.Condition, b.uses(statement.Condition), nil)
		b.current.Terminal = &If{Consequent: body.Id, Alternate: after.Id}
	}
	b.current = after
}

// forOf evaluates the iterable once, then at the head of each pass either binds the next element
// and runs the body, or leaves. continue goes back to the head.
func (b *builder) forOf(statement ir.ForOf) {
	b.emit(statement, statement.Iterable, b.uses(statement.Iterable), nil)
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
	b.emit(statement, nil, nil, defines)
	b.jumps = append(b.jumps, jump{breakTo: after.Id, continueTo: head.Id})
	b.statements(statement.Body)
	b.jumps = b.jumps[:len(b.jumps)-1]
	b.enter(head)
	b.current = after
}

// switchStatement evaluates the value, then each case's tests in order until one matches, and runs
// that case's body, or the default's when none does. 0.1 has no fallthrough: every body ends by
// leaving the switch.
func (b *builder) switchStatement(statement ir.Switch) {
	b.emit(statement, statement.Value, b.uses(statement.Value), nil)
	after := b.function.NewBlock()
	b.jumps = append(b.jumps, jump{breakTo: after.Id})
	bodies := make([]*BasicBlock, len(statement.Cases))
	for index, switchCase := range statement.Cases {
		bodies[index] = b.function.NewBlock()
		for _, test := range switchCase.Tests {
			next := b.function.NewBlock()
			b.emit(statement, test, b.uses(test), nil)
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
