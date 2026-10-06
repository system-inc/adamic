package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// parallelMap keeps both operands alive through the runtime's join. The result is owned.
func (e *emitter) parallelMap(expression ir.ParallelMap) string {
	items := e.value(expression.Items)
	work := e.value(expression.Work)
	result := e.own(ir.Array, fmt.Sprintf("adamic_parallel_map(%s, %s)", items, work))
	e.closureThrown()
	return result
}
