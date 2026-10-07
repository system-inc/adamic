package fresh

import "github.com/system-inc/adamic/internal/ir"

// nodeFSFile evaluates borrowed arguments in order. The synchronous host keeps
// none of them, writes no reference into them, and returns independently owned
// Stats/Date objects or primitive values. Stats.mtime is independently owned too.
func (a *analysis) nodeFSFile(expression ir.NodeFSFile) value {
	for _, argument := range expression.Arguments {
		a.value(argument)
	}
	if expression.Operation == "stat" {
		return a.fresh("mtime", a.fresh(anyField, value{}))
	}
	if mutable(expression.Type()) {
		return a.fresh(anyField, value{})
	}
	return value{}
}
