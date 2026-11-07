package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) labeled(statement ir.Labeled) {
	nativeEmitter.temporaries++
	target := &loop{
		sourceName: statement.Name,
		breakLabel: fmt.Sprintf("adamic_break_%d", nativeEmitter.temporaries),
		depth:      len(nativeEmitter.scopes),
	}
	nativeEmitter.loops = append(nativeEmitter.loops, target)
	nativeEmitter.line("{")
	nativeEmitter.nested(statement.Body, nil)
	nativeEmitter.line("}")
	nativeEmitter.loops = nativeEmitter.loops[:len(nativeEmitter.loops)-1]
	if target.broken {
		nativeEmitter.line("%s:;", target.breakLabel)
	}
}

func (nativeEmitter *emitter) linkLabels(names []string, iteration *loop) {
	for _, name := range names {
		for index := len(nativeEmitter.loops) - 1; index >= 0; index-- {
			if nativeEmitter.loops[index].sourceName == name {
				nativeEmitter.loops[index].iteration = iteration
				break
			}
		}
	}
}

func (nativeEmitter *emitter) labeledJump(name string, continued bool) {
	for index := len(nativeEmitter.loops) - 1; index >= 0; index-- {
		target := nativeEmitter.loops[index]
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
		nativeEmitter.finallies(nativeEmitter.innerHandlers(depth))
		nativeEmitter.releaseScopes(depth)
		nativeEmitter.line("goto %s;", label)
		return
	}
	panic("native: unknown label " + name)
}
