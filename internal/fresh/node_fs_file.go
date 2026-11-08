package fresh

import "github.com/system-inc/adamic/internal/ir"

// nodeFileSystemFile evaluates arguments in order. A watcher retains its listener;
// conservatively let its reachable values escape. Other host calls borrow
// without retaining or storing references and return independently owned
// Stats/Date objects or primitive values. Stats.mtime is independently owned too.
func (analysis *analysis) nodeFileSystemFile(expression ir.NodeFSFile) value {
	for _, argument := range expression.Arguments {
		held := analysis.value(argument)
		if expression.Operation == "host_watch_file" {
			analysis.state.escape(held)
		}
	}
	if expression.Operation == "stat" {
		mtime := analysis.fresh(anyField, value{})
		atime := analysis.fresh(anyField, value{})
		result := analysis.fresh("mtime", mtime)
		analysis.state.store(newestOf(result), "atime", atime)
		return result
	}
	if mutable(expression.Type()) {
		return analysis.fresh(anyField, value{})
	}
	return value{}
}
