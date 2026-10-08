package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) dateCall(call ir.NodeFSFile) string {
	args := make([]string, len(call.Arguments))
	for i, argument := range call.Arguments {
		args[i] = e.value(argument)
	}
	name := strings.TrimPrefix(call.Operation, "date_")
	switch name {
	case "new", "new_iso":
		return "new Date(" + args[0] + ")"
	case "new_now":
		return "new Date()"
	case "copy":
		return "new Date(" + args[0] + ")"
	case "time":
		return "(" + args[0] + ").getTime()"
	case "now_function":
		return "adamicDateNow"
	case "now", "UTC", "parse":
		return "Date." + name + "(" + strings.Join(args, ", ") + ")"
	case "constructor_own":
		return "Date.hasOwnProperty(" + args[0] + ")"
	case "prototype_own":
		return "Date.prototype.hasOwnProperty(" + args[0] + ")"
	}
	return "(" + args[0] + ")." + name + "()"
}

// The JavaScript backend holds closures as code/environment records. Keep one
// descriptor for Date.now so repeated reads retain the intrinsic's identity.
func dateRuntime(code string) string {
	if strings.Contains(code, "adamicDateNow") {
		return "const adamicDateNow = {code: () => Date.now()};\n" + code
	}
	return code
}
