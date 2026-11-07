package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
)

func (javascriptEmitter *emitter) nodeHostCall(call ir.NodeHostCall) string {
	if call.Module == "node:path" {
		return "adamicNodePath." + call.Member + "(" + javascriptEmitter.values(call.Arguments) + ")"
	}
	switch call.Member {
	case "native":
		return "adamicNodeFS.realpathSync.native(" + javascriptEmitter.values(call.Arguments) + ")"
	case "symlinkSync", "readdirSync", "realpathSync":
		return "adamicNodeFS." + call.Member + "(" + javascriptEmitter.values(call.Arguments) + ")"
	default:
		return "(" + javascriptEmitter.value(call.Arguments[0]) + ")." + call.Member + "()"
	}
}
