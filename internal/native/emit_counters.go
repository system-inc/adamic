package native

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) integerCounter(local int) bool {
	if _, suspended := nativeEmitter.asyncSlots[local]; suspended {
		return false
	}
	if mode, overridden := nativeEmitter.counterModes[local]; overridden {
		return mode
	}
	return nativeEmitter.program.Locals[local].Counter
}

// counterBlock emits mutually exclusive integer and ordinary double versions
// of the same source block. Mode overrides are local to this emitter and scope;
// neither the IR nor any other backend's view of its locals is changed.
func (nativeEmitter *emitter) counterBlock(block ir.Block) bool {
	if len(block.Body) != 2 {
		return false
	}
	declaration, ok := block.Body[0].(ir.Declare)
	if !ok {
		return false
	}
	plan := nativeEmitter.program.Locals[declaration.Local].CounterGuard
	if plan == nil {
		return false
	}
	// Suppress nested specializations inside a double fallback. Independent
	// nested guards otherwise duplicate code exponentially with nesting depth.
	if nativeEmitter.doubleCounterPath {
		nativeEmitter.line("{")
		nativeEmitter.nested(block.Body, nil)
		nativeEmitter.line("}")
		return true
	}
	for _, dependency := range plan.Dependencies {
		if !nativeEmitter.integerCounter(dependency) {
			nativeEmitter.line("{")
			nativeEmitter.nested(block.Body, nil)
			nativeEmitter.line("}")
			return true
		}
	}
	if nativeEmitter.counterModes == nil {
		nativeEmitter.counterModes = map[int]bool{}
	}
	previous, present := nativeEmitter.counterModes[declaration.Local]
	defer func() {
		if present {
			nativeEmitter.counterModes[declaration.Local] = previous
		} else {
			delete(nativeEmitter.counterModes, declaration.Local)
		}
	}()
	if plan.Bound == nil {
		nativeEmitter.counterModes[declaration.Local] = true
		nativeEmitter.line("{")
		nativeEmitter.nested(block.Body, nil)
		nativeEmitter.line("}")
		return true
	}
	bound, step := nativeEmitter.value(plan.Bound), nativeEmitter.value(plan.Step)
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
	nativeEmitter.line("if (%s) { // integer counter entry guard", strings.Join(checks, " && "))
	nativeEmitter.counterModes[declaration.Local] = true
	nativeEmitter.nested(block.Body, nil)
	nativeEmitter.line("} else {")
	nativeEmitter.counterModes[declaration.Local] = false
	outerDoublePath := nativeEmitter.doubleCounterPath
	nativeEmitter.doubleCounterPath = true
	nativeEmitter.nested(block.Body, nil)
	nativeEmitter.doubleCounterPath = outerDoublePath
	nativeEmitter.line("}")
	return true
}
