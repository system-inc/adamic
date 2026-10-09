package lower

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) generatorBody(index int, node *ast.Node, this int, defaults []defaulted, patterns []patterned) error {
	original := l.result.Functions[index]
	types := original.Generator
	types.Lowering = true
	l.result.Functions[index].Returns = types.Return
	if err := l.lowerBody(index, node, this, defaults, patterns); err != nil {
		return err
	}
	types.Lowering = false
	function := l.result.Functions[index]
	prefix, body := function.Body[:types.Prologue], function.Body[types.Prologue:]
	frame := l.iterationLocal("generator_frame", ir.Object, index)
	l.result.Locals[frame].Captured = true
	l.noteLocal(frame, l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(node))), node)
	types.Frame = frame
	resume := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "generator_resume", Closure: true, Returns: ir.Object})
	g := &generatorMachine{l: l, node: node, factory: index, resume: resume, frame: frame, slots: map[int]bool{}, types: types}
	for local, declared := range l.result.Locals {
		if local != frame && declared.Function == index {
			g.slots[local] = true
		}
	}
	// A closure retaining this frame would make a strong frame -> closure -> frame
	// cycle. Do not let moving locals into slots erase the existing capture proof.
	for target, nested := range l.result.Functions {
		if target == index {
			continue
		}
		for _, captured := range nested.Environment {
			if g.slots[captured] {
				return l.notYet(node, "generator suspension nested capture "+l.result.Locals[captured].Name+" needs a weak frame reference")
			}
		}
	}
	caught := l.iterationLocal("generator_error", ir.Object, resume)
	root := generatorContext{handler: -1}
	complete := g.finish(ir.Undefined{Of: ir.Union}, root)
	entry := g.statements(body, complete, root)
	if g.failure != nil {
		return g.failure
	}
	// Frame fields retain every incoming and defaulted parameter at the call. Body
	// locals are initialized by resume states; their placeholder is unobservable.
	initialized := map[int]bool{}
	for _, local := range function.Parameters {
		initialized[local] = true
	}
	for _, statement := range prefix {
		if declaration, ok := statement.(ir.Declare); ok {
			initialized[declaration.Local] = true
		}
	}
	fields := []ir.Field{{Name: "pc", Value: ir.NumberConstant{Value: float64(entry)}}, {Name: "status", Value: ir.NumberConstant{}}, {Name: "mode", Value: ir.NumberConstant{}}, {Name: "input", Value: ir.Undefined{Of: ir.Union}}, {Name: "returned", Value: ir.Undefined{Of: ir.Union}}, {Name: "error", Value: ir.Undefined{Of: ir.Object}}}
	locals := []int{}
	for local := range g.slots {
		locals = append(locals, local)
	}
	sort.Ints(locals)
	types.FrameLocals = append(append([]int{}, locals...), function.Environment...)
	for _, local := range locals {
		declared := l.result.Locals[local]
		value := ir.Expression(ir.Undefined{Of: declared.Type})
		if declared.Type == ir.Number {
			value = ir.NumberConstant{}
		} else if declared.Type == ir.Boolean {
			value = ir.BooleanConstant{}
		} else if declared.Type.IsMaybe() {
			value = ir.MaybeOf{Of: declared.Type}
		}
		if initialized[local] {
			value = ir.Read{Local: local, Of: declared.Type}
		}
		fields = append(fields, ir.Field{Name: generatorSlot(local), Value: value})
	}
	prefix = append(prefix, ir.Declare{Local: frame, Value: ir.GeneratorFrame{Value: ir.ObjectLiteral{Fields: fields}}})
	environment := append([]int{frame}, function.Environment...)
	for _, local := range environment {
		l.result.Locals[local].Captured = true
	}
	mode, input := l.iterationLocal("generator_mode", ir.Number, resume), l.iterationLocal("generator_input", ir.Union, resume)
	method := ir.Function{Name: "generator_resume", Closure: true, Returns: ir.Object, Parameters: []int{mode, input}, Environment: environment}
	typeError := ir.MakeError{Message: ir.StringConstant{Index: l.constant("Generator is already running")}, Name: ir.StringConstant{Index: l.constant("TypeError")}}
	method.Body = []ir.Statement{ir.If{Condition: g.equals("status", 1), Then: []ir.Statement{ir.Throw{Value: typeError}}}}
	supplied := ir.Read{Local: input, Of: ir.Union}
	completed := []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: ir.Read{Local: mode, Of: ir.Number}, Right: ir.NumberConstant{Value: 2}}, Then: []ir.Statement{ir.Throw{Value: ir.Narrow{Value: supplied, To: ir.Object}}}}, ir.Return{Value: g.packet(ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: ir.Read{Local: mode, Of: ir.Number}, Right: ir.NumberConstant{Value: 1}}, WhenTrue: supplied, WhenNot: ir.Undefined{Of: ir.Union}, Of: ir.Union}, true)}}
	method.Body = append(method.Body, ir.If{Condition: g.equals("status", 3), Then: completed})
	// suspended-start discards next's argument. Abrupt resume never starts the body.
	early := g.finish(supplied, root)
	earlyState := g.states[early].body
	method.Body = append(method.Body, ir.If{Condition: g.equals("status", 0), Then: []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.NotEqual, Left: ir.Read{Local: mode, Of: ir.Number}, Right: ir.NumberConstant{}}, Then: append([]ir.Statement{g.store("status", ir.NumberConstant{Value: 3}), ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: ir.Read{Local: mode, Of: ir.Number}, Right: ir.NumberConstant{Value: 2}}, Then: []ir.Statement{ir.Throw{Value: ir.Narrow{Value: supplied, To: ir.Object}}}}}, earlyState...)}, ir.Assign{Local: input, Value: ir.Undefined{Of: ir.Union}}}})
	method.Body = append(method.Body, g.store("status", ir.NumberConstant{Value: 1}), g.store("mode", ir.Read{Local: mode, Of: ir.Number}), g.store("input", supplied))
	dispatch := ir.Switch{Value: g.field("pc", ir.Number)}
	for pc, state := range g.states {
		handler := []ir.Statement{g.store("error", ir.Read{Local: caught, Of: ir.Object})}
		if state.handler >= 0 {
			handler = append(handler, g.jump(state.handler)...)
		} else {
			handler = append(handler, g.store("status", ir.NumberConstant{Value: 3}), ir.Throw{Value: ir.Read{Local: caught, Of: ir.Object}})
		}
		dispatch.Cases = append(dispatch.Cases, ir.Case{Tests: []ir.Expression{ir.NumberConstant{Value: float64(pc)}}, Body: []ir.Statement{ir.Try{Body: state.body, HasCatch: true, CatchLocal: caught, Catch: handler}}})
	}
	method.Body = append(method.Body, ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: []ir.Statement{dispatch}})
	l.result.Functions[resume] = method
	resumed := l.iterationLocal("generator_resume_function", ir.Closure, index)
	l.result.Locals[resumed].Captured = true
	prefix = append(prefix, ir.Declare{Local: resumed, Value: ir.MakeClosure{Function: resume}})
	objectFields := []ir.Field{}
	for i, name := range []string{"next", "return", "throw"} {
		target := len(l.result.Functions)
		receiver := l.iterationLocal("generator_this", ir.Object, target)
		argument := l.iterationLocal("generator_argument", ir.Union, target)
		wrapper := ir.Function{Name: "generator_" + name, Closure: true, Receiver: true, Returns: ir.Object, Parameters: []int{receiver, argument}, OptionalParameters: map[int]bool{argument: true}, Environment: []int{resumed}}
		wrapper.Body = []ir.Statement{ir.Return{Value: ir.CallClosure{Closure: ir.Read{Local: resumed, Of: ir.Closure}, Arguments: []ir.Expression{ir.NumberConstant{Value: float64(i)}, ir.Read{Local: argument, Of: ir.Union}}, Returns: ir.Object}}}
		l.result.Functions = append(l.result.Functions, wrapper)
		objectFields = append(objectFields, ir.Field{Name: name, Value: ir.MakeClosure{Function: target}})
	}
	target := len(l.result.Functions)
	receiver := l.iterationLocal("generator_iterator_this", ir.Object, target)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "generator_iterator", Closure: true, Receiver: true, Parameters: []int{receiver}, Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: ir.Read{Local: receiver, Of: ir.Object}}}})
	objectFields = append(objectFields, ir.Field{Name: iteratorSlot, Value: ir.MakeClosure{Function: target}})
	function.Returns, function.Body, function.Generator = ir.Object, append(prefix, ir.Return{Value: ir.ObjectLiteral{Fields: objectFields}}), types
	l.result.Functions[index] = function
	return g.failure
}
