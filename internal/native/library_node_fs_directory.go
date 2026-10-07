package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nodeHostCall(call ir.NodeHostCall) string {
	arguments := []string{}
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	var code string
	if call.Module == "node:path" {
		if call.Member == "relative" {
			code = "adamic_node_path_relative(" + arguments[0] + ", " + arguments[1] + ")"
		} else if call.Member == "dirname" {
			code = "adamic_node_path_dirname(" + arguments[0] + ")"
		} else {
			array := "NULL"
			if len(arguments) > 0 {
				array = "(adamic_string *const[]){" + strings.Join(arguments, ", ") + "}"
			}
			code = fmt.Sprintf("adamic_node_path_%s(%d, %s)", call.Member, len(arguments), array)
		}
	} else {
		switch call.Member {
		case "readdirSync":
			options := "NULL"
			if len(arguments) > 1 {
				options = arguments[1]
			}
			code = fmt.Sprintf("adamic_node_fs_readdir(%s, %s)", arguments[0], options)
		case "realpathSync", "native":
			code = fmt.Sprintf("adamic_node_fs_realpath(%s, %t)", arguments[0], call.Member == "native")
		default:
			code = fmt.Sprintf("adamic_node_fs_dirent_is(%s, %s)", arguments[0], cString(call.Member))
		}
	}
	var result string
	if call.Returns.IsReference() {
		result = e.own(call.Returns, code)
	} else {
		result = e.snapshot(call.Returns, code)
	}
	if call.Throws {
		e.checkThrown()
	}
	return result
}
