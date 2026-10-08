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

func (c *asyncCompletion) temporary(of ir.Type) int {
	local := c.n.temporary(of)
	c.locals = append(c.locals, ir.Declare{Local: local})
	return local
}

func (c *asyncCompletion) id() int { c.next++; return c.next }
func (c *asyncCompletion) read(local int) ir.Read {
	return ir.Read{Local: local, Of: c.n.program.Locals[local].Type}
}
func (c *asyncCompletion) depth(target int) int {
	for i := len(c.emitted) - 1; i >= 0; i-- {
		if c.emitted[i] == target {
			return len(c.emitted) - 1 - i
		}
	}
	panic("async completion target outside function")
}
func (c *asyncCompletion) abrupt(statement ir.Statement, target int) []ir.Statement {
	if _, ok := statement.(ir.Throw); ok {
		return []ir.Statement{statement}
	}
	for i := len(c.attempts) - 1; i >= 0; i-- {
		attempt := c.attempts[i]
		if target != 0 && !attempt.outside[target] {
			continue
		}
		var before []ir.Statement
		if ret, ok := statement.(ir.Return); ok && ret.Value != nil {
			local := c.temporary(ret.Value.Type())
			c.n.program.Locals[local].Name = "pending return"
			before = append(before, ir.Assign{Local: local, Value: ret.Value})
			statement = ir.Return{Value: c.read(local)}
		}
		attempt.exits = append(attempt.exits, asyncExit{statement, target})
		before = append(before, ir.Assign{Local: attempt.selector, Value: ir.NumberConstant{Value: float64(len(attempt.exits))}}, ir.Break{Depth: c.depth(attempt.boundary)})
		return before
	}
	if _, ok := statement.(ir.Break); ok {
		statement = ir.Break{Depth: c.depth(target)}
	}
	return []ir.Statement{statement}
}
func (c *asyncCompletion) statements(body []ir.Statement) []ir.Statement {
	var out []ir.Statement
	for _, statement := range body {
		switch value := statement.(type) {
		case ir.Return:
			out = append(out, c.abrupt(value, 0)...)
			continue
		case ir.Break:
			out = append(out, c.abrupt(value, c.source[len(c.source)-1-value.Depth])...)
			continue
		case ir.Continue:
			target := 0
			for i := len(c.source) - 1; i >= 0; i-- {
				if c.loops[c.source[i]] {
					target = c.source[i]
					break
				}
			}
			out = append(out, c.abrupt(value, target)...)
			continue
		case ir.If:
			value.Then = c.statements(value.Then)
			value.Else = c.statements(value.Else)
			statement = value
		case ir.Block:
			value.Body = c.statements(value.Body)
			statement = value
		case ir.Loop, ir.ForOf, ir.Switch:
			id := c.id()
			c.source = append(c.source, id)
			c.emitted = append(c.emitted, id)
			switch value := value.(type) {
			case ir.Loop:
				c.loops[id] = true
				value.Body = c.statements(value.Body)
				value.Test = c.statements(value.Test)
				value.Update = c.statements(value.Update)
				statement = value
			case ir.ForOf:
				c.loops[id] = true
				value.Body = c.statements(value.Body)
				statement = value
			case ir.Switch:
				for i := range value.Cases {
					value.Cases[i].Body = c.statements(value.Cases[i].Body)
				}
				value.Default = c.statements(value.Default)
				statement = value
			}
			c.source = c.source[:len(c.source)-1]
			c.emitted = c.emitted[:len(c.emitted)-1]
		case ir.Try:
			if !value.HasFinally {
				value.Body = c.statements(value.Body)
				value.Catch = c.statements(value.Catch)
				statement = value
				break
			}
			selector := c.temporary(ir.Number)
			caught := c.n.temporary(ir.Object)
			payload := c.temporary(ir.Object)
			attempt := &asyncFinally{boundary: c.id(), selector: selector, outside: map[int]bool{}}
			for _, id := range c.source {
				attempt.outside[id] = true
			}
			c.emitted = append(c.emitted, attempt.boundary)
			c.attempts = append(c.attempts, attempt)
			inner := value
			inner.HasFinally = false
			inner.Finally = nil
			inner.Body = c.statements(inner.Body)
			inner.Catch = c.statements(inner.Catch)
			c.attempts = c.attempts[:len(c.attempts)-1]
			// The wrapper catches both explicit throws and rejection/throw edges from
			// the original body and catch. Its payload is separate from frame->error.
			attempt.exits = append(attempt.exits, asyncExit{statement: ir.Throw{Value: c.read(payload)}})
			guardedBody := []ir.Statement{inner}
			if !inner.HasCatch {
				guardedBody = inner.Body
			}
			guarded := ir.Try{Body: guardedBody, HasCatch: true, CatchLocal: caught, Catch: []ir.Statement{ir.Assign{Local: payload, Value: c.read(caught)}, ir.Assign{Local: selector, Value: ir.NumberConstant{Value: float64(len(attempt.exits))}}}}
			out = append(out, ir.Assign{Local: selector, Value: ir.NumberConstant{Value: 0}}, ir.Switch{Value: ir.BooleanConstant{Value: true}, Default: []ir.Statement{guarded}})
			c.emitted = c.emitted[:len(c.emitted)-1]
			out = append(out, c.statements(value.Finally)...)
			// Dispatch uses no synthetic breakable, so original break depths remain
			// valid. An abrupt finally never reaches this pending dispatch.
			for i, exit := range attempt.exits {
				out = append(out, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: c.read(selector), Right: ir.NumberConstant{Value: float64(i + 1)}}, Then: c.abrupt(exit.statement, exit.target)})
			}
			continue
		}
		out = append(out, statement)
	}
	return out
}
