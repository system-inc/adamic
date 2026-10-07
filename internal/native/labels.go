package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) labeled(statement ir.Labeled) {
	e.temporaries++
	target := &loop{sourceName: statement.Name, breakLabel: fmt.Sprintf("adamic_break_%d", e.temporaries), depth: len(e.scopes)}
	e.loops = append(e.loops, target)
	e.line("{")
	e.nested(statement.Body, nil)
	e.line("}")
	e.loops = e.loops[:len(e.loops)-1]
	if target.broken {
		e.line("%s:;", target.breakLabel)
	}
}

func (e *emitter) linkLabels(names []string, iteration *loop) {
	for _, name := range names {
		for index := len(e.loops) - 1; index >= 0; index-- {
			if e.loops[index].sourceName == name {
				e.loops[index].iteration = iteration
				break
			}
		}
	}
}

func (e *emitter) labeledJump(name string, continued bool) {
	for index := len(e.loops) - 1; index >= 0; index-- {
		target := e.loops[index]
		if target.sourceName != name {
			continue
		}
		label, depth := target.breakLabel, target.depth
		if continued {
			if target.iteration == nil {
				panic("native: labeled continue without an iteration")
			}
			target.iteration.continued = true
			label, depth = target.iteration.label, target.iteration.depth
		} else {
			target.broken = true
		}
		e.finallies(e.innerHandlers(depth))
		e.releaseScopes(depth)
		e.line("goto %s;", label)
		return
	}
	panic("native: unknown label " + name)
}
