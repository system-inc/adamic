package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"strings"
)

// nativeHostCall reports whether native emits a host call: node:path and node:fs, synchronously. A
// workers: module is JavaScript's alone, and a Promise form settles only under JavaScript's await.
func nativeHostCall(call ir.NodeHostCall) bool {
	return (call.Module == "node:path" || call.Module == "node:fs") && call.Returns != ir.Promise
}

// HostCallRefusal names the first host operation in a program that native doesn't emit, or is nil. A
// native or Wasm build of a program holding a JavaScript-only host operation (workers:KvNamespace.get,
// reached when a corpus or a build lowers a Worker's program as an ordinary one) is refused by name,
// before emission, as an ordinary error. WASI and WASIExports return it, and internal/lower's
// cycles_equality_test.go asks it before native.C. nodeHostCall's panic remains as the backstop for a
// caller that emits without asking.
func HostCallRefusal(program *ir.Program) error {
	var refusal error
	var walk func(reflect.Value)
	walk = func(value reflect.Value) {
		if refusal != nil || !value.IsValid() {
			return
		}
		// A type is matched only on a concrete value, never on an interface holding one, which the walk
		// enters next, so each call is met once.
		if value.Kind() != reflect.Interface && value.CanInterface() {
			if call, isCall := value.Interface().(ir.NodeHostCall); isCall && !nativeHostCall(call) {
				refusal = fmt.Errorf("native: refused host operation %s.%s: native emits only node:path and node:fs, synchronously; fix: build the program with adamic worker, and keep the function that reaches it out of --wasm", call.Module, call.Member)
				return
			}
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				walk(value.Elem())
			}
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				walk(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				walk(value.Index(index))
			}
		}
	}
	walk(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		walk(reflect.ValueOf(function.Body))
	}
	return refusal
}

func (e *emitter) nodeHostCall(call ir.NodeHostCall) string {
	// HostCallRefusal refuses these by name before emission; one reaching here came by a caller that
	// didn't ask, a compiler bug said out loud, before any branch below could read it as node:fs.
	if !nativeHostCall(call) {
		panic(fmt.Sprintf("native: compiler bug: no native emission for host operation %s.%s", call.Module, call.Member))
	}
	arguments := []string{}
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	var code string
	if call.Module == "node:path" {
		if call.Member == "basename" {
			suffix := "NULL"
			if len(arguments) > 1 {
				suffix = arguments[1]
			}
			code = "adamic_node_path_basename(" + arguments[0] + ", " + suffix + ")"
		} else if call.Member == "relative" {
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
		case "symlinkSync":
			code = fmt.Sprintf("adamic_node_fs_symlink(%s, %s)", arguments[0], arguments[1])
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
