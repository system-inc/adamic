package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
)

// workersHostCall calls a Workers binding the bridge checked, as the binding's own method. get's value
// check rides the call, never the await that settles it, so ir.Await stays generic: a get that settles
// to anything but text or null stops in adamicWorkersTextOrNull (internal/worker/runtime.mjs), named by
// its typeof. Any other workers: operation has no emission here, said out loud.
func (e *emitter) workersHostCall(call ir.NodeHostCall) string {
	if call.Module != "workers:KvNamespace" || call.Member != "get" || len(call.Arguments) != 2 {
		panic("compiler bug: no JavaScript emission for host operation " + call.Module + "." + call.Member)
	}
	e.workersTextOrNullUsed = true
	return "(" + e.value(call.Arguments[0]) + ").get(" + e.value(call.Arguments[1]) + ").then(adamicWorkersTextOrNull)"
}
