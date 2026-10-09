package native

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) integerCounter(local int) bool {
	if _, suspended := e.asyncSlots[local]; suspended {
		return false
	}
	if mode, overridden := e.counterModes[local]; overridden {
		return mode
	}
	return e.program.Locals[local].Counter
}

// counterBlock emits mutually exclusive integer and ordinary double versions
// of the same source block. Mode overrides are local to this emitter and scope;
// neither the IR nor any other backend's view of its locals is changed.
func (e *emitter) counterBlock(block ir.Block) bool {
	if len(block.Body) != 2 {
		return false
	}
	declaration, ok := block.Body[0].(ir.Declare)
	if !ok {
		return false
	}
	plan := e.program.Locals[declaration.Local].CounterGuard
	if plan == nil {
		return false
	}
	// Suppress nested specializations inside a double fallback. Independent
	// nested guards otherwise duplicate code exponentially with nesting depth.
	if e.doubleCounterPath {
		e.line("{")
		e.nested(block.Body, nil)
		e.line("}")
		return true
	}
	for _, dependency := range plan.Dependencies {
		if !e.integerCounter(dependency) {
			e.line("{")
			e.nested(block.Body, nil)
			e.line("}")
			return true
		}
	}
	if e.counterModes == nil {
		e.counterModes = map[int]bool{}
	}
	previous, present := e.counterModes[declaration.Local]
	defer func() {
		if present {
			e.counterModes[declaration.Local] = previous
		} else {
			delete(e.counterModes, declaration.Local)
		}
	}()
	if plan.Bound == nil {
		e.counterModes[declaration.Local] = true
		e.line("{")
		e.nested(block.Body, nil)
		e.line("}")
		return true
	}
	bound, step := e.value(plan.Bound), e.value(plan.Step)
	exact := "9007199254740992.0"
	checks := []string{
		fmt.Sprintf("(%s >= -%s && %s <= %s && %s == trunc(%s))", bound, exact, bound, exact, bound, bound),
		fmt.Sprintf("(%s >= 1.0 && %s <= %s && %s == trunc(%s))", step, step, exact, step, step),
	}
	if plan.Square {
		checks = append(checks, fmt.Sprintf("(%s <= %s)", step, strconv.FormatFloat(plan.MaxStep, 'f', 1, 64)))
	} else {
		size := step
		if !plan.Inclusive {
			size = "(" + step + " - 1.0)"
		}
		if plan.Ascending {
			checks = append(checks, fmt.Sprintf("(%s <= %s - %s)", bound, exact, size))
		} else {
			checks = append(checks, fmt.Sprintf("(%s >= -%s + %s)", bound, exact, size))
		}
	}
	e.line("if (%s) { // integer counter entry guard", strings.Join(checks, " && "))
	e.counterModes[declaration.Local] = true
	e.nested(block.Body, nil)
	e.line("} else {")
	e.counterModes[declaration.Local] = false
	outerDoublePath := e.doubleCounterPath
	e.doubleCounterPath = true
	e.nested(block.Body, nil)
	e.doubleCounterPath = outerDoublePath
	e.line("}")
	return true
}
