package fresh

import "github.com/system-inc/adamic/internal/ir"

// nodeFSFile evaluates arguments in order. A watcher retains its listener, and
// FSWatcher.on its receiver too; conservatively let their reachable values escape.
// Other host calls borrow without retaining or storing references and return
// independently owned Stats/Date objects or primitive values. Stats.mtime is independently owned too.
func (a *analysis) nodeFSFile(expression ir.NodeFSFile) value {
	for _, argument := range expression.Arguments {
		held := a.value(argument)
		if expression.RetainsListener() {
			a.state.escape(held)
		}
	}
	if expression.Operation == "stat" {
		mtime := a.fresh(anyField, value{})
		atime := a.fresh(anyField, value{})
		result := a.fresh("mtime", mtime)
		a.state.store(newestOf(result), "atime", atime)
		return result
	}
	if mutable(expression.Type()) {
		return a.fresh(anyField, value{})
	}
	return value{}
}
