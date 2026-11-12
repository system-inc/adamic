package fresh

import "github.com/system-inc/adamic/internal/ir"

// nodeFSFile evaluates arguments in order. A watcher retains its listener, and
// FSWatcher.on its receiver too; conservatively let their reachable values escape.
// Other host calls borrow without retaining or storing references and return
// independently owned Stats/Date objects or primitive values. A Stats' four Dates are
// independently owned too, and a read of its birthtime (ir.StatsFieldOperations) hands out
// the Date the Stats holds, as a read of the field would.
func (a *analysis) nodeFSFile(expression ir.NodeFSFile) value {
	held := []value{}
	for _, argument := range expression.Arguments {
		operand := a.value(argument)
		held = append(held, operand)
		if expression.RetainsListener() {
			a.state.escape(operand)
		}
	}
	if expression.Operation == "stat" {
		mtime := a.fresh(anyField, value{})
		atime := a.fresh(anyField, value{})
		ctime := a.fresh(anyField, value{})
		birthtime := a.fresh(anyField, value{})
		result := a.fresh("mtime", mtime)
		a.state.store(newestOf(result), "atime", atime)
		a.state.store(newestOf(result), "ctime", ctime)
		a.state.store(newestOf(result), "birthtime", birthtime)
		return result
	}
	if expression.Operation == ir.StatsFieldOperations["birthtime"] && len(held) == 1 {
		return a.load(held[0], "birthtime")
	}
	if mutable(expression.Type()) {
		return a.fresh(anyField, value{})
	}
	return value{}
}
