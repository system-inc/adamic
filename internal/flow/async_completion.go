package flow

import "github.com/system-inc/adamic/internal/ir"

// Async completions are executable ordinary IR, rather than the conservative Choose
// edges used to analyze synchronous finally. Each finally has its own selector and
// payload cells, so an await or a nested try cannot overwrite an outer completion.
type asyncCompletion struct {
	n               *asyncNormalizer
	next            int
	source, emitted []int
	loops           map[int]bool
	names           map[int]string
	attempts        []*asyncFinally
	locals          []ir.Statement
}
type asyncFinally struct {
	boundary, selector int
	outside            map[int]bool
	exits              []asyncExit
}
type asyncExit struct {
	statement ir.Statement
	target    int
}

func (completion *asyncCompletion) temporary(of ir.Type) int {
	local := completion.n.temporary(of)
	completion.locals = append(completion.locals, ir.Declare{Local: local})
	return local
}

func (completion *asyncCompletion) id() int { completion.next++; return completion.next }
func (completion *asyncCompletion) read(local int) ir.Read {
	return ir.Read{Local: local, Of: completion.n.program.Locals[local].Type}
}

// Named wrappers participate in lexical lookup, but never in an unlabeled break's depth.
func (completion *asyncCompletion) target(label string, depth int, continued bool) int {
	for i := len(completion.source) - 1; i >= 0; i-- {
		id := completion.source[i]
		if label != "" {
			if completion.names[id] == label {
				return id
			}
			continue
		}
		if completion.names[id] != "" || (continued && !completion.loops[id]) {
			continue
		}
		if depth == 0 {
			return id
		}
		depth--
	}
	panic("async completion target outside function")
}
func (completion *asyncCompletion) depth(target int) int {
	depth := 0
	for i := len(completion.emitted) - 1; i >= 0; i-- {
		id := completion.emitted[i]
		if completion.names[id] != "" {
			continue
		}
		if id == target {
			return depth
		}
		depth++
	}
	panic("async completion target outside function")
}
func (completion *asyncCompletion) abrupt(statement ir.Statement, target int) []ir.Statement {
	if _, ok := statement.(ir.Throw); ok {
		return []ir.Statement{statement}
	}
	for i := len(completion.attempts) - 1; i >= 0; i-- {
		attempt := completion.attempts[i]
		if target != 0 && !attempt.outside[target] {
			continue
		}
		var before []ir.Statement
		if ret, ok := statement.(ir.Return); ok && ret.Value != nil {
			if ret.Value.Type() == 0 {
				before = append(before, ir.Evaluate{Value: ret.Value})
				statement = ir.Return{}
			} else {
				local := completion.temporary(ret.Value.Type())
				completion.n.program.Locals[local].Name = "pending return"
				before = append(before, ir.Assign{Local: local, Value: ret.Value})
				statement = ir.Return{Value: completion.read(local)}
			}
		}
		attempt.exits = append(attempt.exits, asyncExit{statement, target})
		before = append(
			before,
			ir.Assign{Local: attempt.selector, Value: ir.NumberConstant{Value: float64(len(attempt.exits))}},
			ir.Break{Depth: completion.depth(attempt.boundary)},
		)
		return before
	}
	if _, ok := statement.(ir.Break); ok {
		if name := completion.names[target]; name != "" {
			statement = ir.Break{Label: name}
		} else {
			statement = ir.Break{Depth: completion.depth(target)}
		}
	}
	return []ir.Statement{statement}
}
func (completion *asyncCompletion) statements(body []ir.Statement) []ir.Statement {
	var out []ir.Statement
	for _, statement := range body {
		switch value := statement.(type) {
		case ir.Return:
			out = append(out, completion.abrupt(value, 0)...)
			continue
		case ir.Break:
			out = append(out, completion.abrupt(value, completion.target(value.Label, value.Depth, false))...)
			continue
		case ir.Continue:
			target := completion.target(value.Label, 0, true)
			out = append(out, completion.abrupt(value, target)...)
			continue
		case ir.If:
			value.Then = completion.statements(value.Then)
			value.Else = completion.statements(value.Else)
			statement = value
		case ir.Block:
			value.Body = completion.statements(value.Body)
			statement = value
		case ir.Loop, ir.ForOf, ir.Switch, ir.Labeled:
			id := completion.id()
			completion.source = append(completion.source, id)
			completion.emitted = append(completion.emitted, id)
			switch value := value.(type) {
			case ir.Labeled:
				completion.names[id] = value.Name
				value.Body = completion.statements(value.Body)
				statement = value
			case ir.Loop:
				completion.loops[id] = true
				value.Body = completion.statements(value.Body)
				value.Test = completion.statements(value.Test)
				value.Update = completion.statements(value.Update)
				statement = value
			case ir.ForOf:
				completion.loops[id] = true
				value.Body = completion.statements(value.Body)
				statement = value
			case ir.Switch:
				for i := range value.Cases {
					value.Cases[i].Body = completion.statements(value.Cases[i].Body)
				}
				value.Default = completion.statements(value.Default)
				statement = value
			}
			completion.source = completion.source[:len(completion.source)-1]
			completion.emitted = completion.emitted[:len(completion.emitted)-1]
		case ir.Try:
			if !value.HasFinally {
				value.Body = completion.statements(value.Body)
				value.Catch = completion.statements(value.Catch)
				statement = value
				break
			}
			selector := completion.temporary(ir.Number)
			caught := completion.n.temporary(ir.Object)
			payload := completion.temporary(ir.Object)
			attempt := &asyncFinally{boundary: completion.id(), selector: selector, outside: map[int]bool{}}
			for _, id := range completion.source {
				attempt.outside[id] = true
			}
			completion.emitted = append(completion.emitted, attempt.boundary)
			completion.attempts = append(completion.attempts, attempt)
			inner := value
			inner.HasFinally = false
			inner.Finally = nil
			inner.Body = completion.statements(inner.Body)
			inner.Catch = completion.statements(inner.Catch)
			completion.attempts = completion.attempts[:len(completion.attempts)-1]
			// The wrapper catches both explicit throws and rejection/throw edges from
			// the original body and catch. Its payload is separate from frame->error.
			attempt.exits = append(attempt.exits, asyncExit{statement: ir.Throw{Value: completion.read(payload)}})
			guardedBody := []ir.Statement{inner}
			if !inner.HasCatch {
				guardedBody = inner.Body
			}
			guarded := ir.Try{Body: guardedBody, HasCatch: true, CatchLocal: caught, Catch: []ir.Statement{
				ir.Assign{Local: payload, Value: completion.read(caught)},
				ir.Assign{Local: selector, Value: ir.NumberConstant{Value: float64(len(attempt.exits))}},
			}}
			out = append(
				out,
				ir.Assign{Local: selector, Value: ir.NumberConstant{Value: 0}},
				ir.Switch{Value: ir.BooleanConstant{Value: true}, Default: []ir.Statement{guarded}},
			)
			completion.emitted = completion.emitted[:len(completion.emitted)-1]
			out = append(out, completion.statements(value.Finally)...)
			// Dispatch uses no synthetic breakable, so original break depths remain
			// valid. An abrupt finally never reaches this pending dispatch.
			for i, exit := range attempt.exits {
				out = append(out, ir.If{Condition: ir.Binary{
					Operator: ir.Equal,
					Left:     completion.read(selector),
					Right:    ir.NumberConstant{Value: float64(i + 1)},
				}, Then: completion.abrupt(exit.statement, exit.target)})
			}
			continue
		}
		out = append(out, statement)
	}
	return out
}
