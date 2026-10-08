package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) nodeHostCall(call ir.NodeHostCall) string {
	if call.Module == "node:path" {
		return "adamicNodePath.posix." + call.Member + "(" + e.values(call.Arguments) + ")"
	}
	switch call.Member {
	case "native":
		return "adamicNodeFS.realpathSync.native(" + e.values(call.Arguments) + ")"
	case "symlinkSync", "readdirSync", "realpathSync":
		return "adamicNodeFS." + call.Member + "(" + e.values(call.Arguments) + ")"
	default:
		return "(" + e.value(call.Arguments[0]) + ")." + call.Member + "()"
	}
}
