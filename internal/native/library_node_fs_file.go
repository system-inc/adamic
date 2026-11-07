package native

import (
	"fmt"
	"slices"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) nodeFSFile(call ir.NodeFSFile) string {
	if call.Operation == "set_timeout" || call.Operation == "clear_timeout" {
		return e.nodeTimer(call)
	}
	args := make([]string, len(call.Arguments))
	for i, argument := range call.Arguments {
		args[i] = e.value(argument)
	}
	code := fmt.Sprintf("adamic_fs_file_%s(%s)", call.Operation, strings.Join(args, ", "))
	if call.Operation == "host_watch_file" {
		code = "adamic_fs_file_watch(" + strings.Join(args, ", ") + ")"
	}
	if call.Operation == "host_unwatch_file" {
		callback := "NULL"
		if len(args) == 2 {
			callback = args[1]
		}
		code = "adamic_fs_file_unwatch(" + args[0] + ", " + callback + ")"
	}
	if call.Operation == "host_join" {
		array := "NULL"
		if len(args) != 0 {
			array = "(adamic_string *const[]){" + strings.Join(args, ", ") + "}"
		}
		code = fmt.Sprintf("adamic_fs_file_host_join(%d, %s)", len(args), array)
	}
	if call.Operation == "host_stdout_columns" {
		const unavailable = `#ifdef ADAMIC_TARGET_WASI
#error "wasi preview1 reports no terminal size"
#endif`
		if !slices.Contains(e.declarations, unavailable) {
			e.declarations = append(e.declarations, unavailable)
		}
		code = "adamic_host_stdout_columns()"
	}
	if call.Operation == "host_stdout_is_tty" {
		code = "adamic_host_stdout_is_tty()"
	}
	if call.Operation == "host_stdout_write" {
		code = "(adamic_write_raw(adamic_stdout, " + args[0] + "), 0.0)"
	}
	if call.Operation == "host_env_set" {
		e.line("adamic_host_environment_set(%s, %s);", args[0], args[1])
		code = "adamic_retain(" + args[1] + ")"
	}
	if call.Operation == "host_env_delete" {
		code = "adamic_host_environment_delete(" + args[0] + ")"
	}
	if call.Operation == "host_env" {
		code = e.nodeEnvironmentRead(args[0])
	}
	if call.Operation == "host_realpath" {
		code = e.snapshot(ir.String, fmt.Sprintf("adamic_host_realpath(%s)", args[0]))
		e.line("if (%s == NULL) { adamic_fs_file_realpath_error(%s); }", code, args[0])
	}
	result := ""
	if call.Of.IsReference() {
		result = e.own(call.Of, code)
	} else {
		result = e.snapshot(call.Of, code)
	}
	if call.MayThrow() {
		e.checkThrown()
	}
	return result
}

// nodeTimer is setTimeout and clearTimeout called from source, on the timer provider (timers.h).
// setTimeout's extra arguments are handed over as the slots a call through its callback would
// pack, each with whether it's a reference, which the provider retains and copies until the
// timeout fires or is cleared, and its thunk calls the callback with them. clearTimeout hands the
// runtime a reference as the heap value it is, which the runtime tells a Timeout by its shape. A union
// is a reference (unknown and NodeJS.Timeout | boolean are), a Timeout in it the object itself. Every
// other representation, a number, a boolean or a maybe of one, can't hold a Timeout, and goes as NULL.
func (e *emitter) nodeTimer(call ir.NodeFSFile) string {
	if include := "#include \"timers.h\""; !slices.Contains(e.declarations, include) {
		e.declarations = append(e.declarations, include)
	}
	if call.Operation == "clear_timeout" {
		value := e.value(call.Arguments[0])
		handle := "NULL"
		if call.Arguments[0].Type().IsReference() {
			handle = "(const adamic_heap *)" + value
		} else {
			e.line("(void)%s;", value)
		}
		return e.snapshot(ir.Number, "adamic_node_clear_timeout("+handle+")")
	}
	callback := e.value(call.Arguments[0])
	delay := e.value(call.Arguments[1])
	slots, references := []string{}, []string{}
	for _, argument := range call.Arguments[2:] {
		slots = append(slots, fmt.Sprintf("{.%s = %s}", member(argument.Type()), slotted(argument.Type(), e.value(argument))))
		references = append(references, fmt.Sprint(argument.Type().IsReference()))
	}
	arguments, mask := "NULL", "NULL"
	if len(slots) != 0 {
		arguments = "(const adamic_value[]){" + strings.Join(slots, ", ") + "}"
		mask = "(const bool[]){" + strings.Join(references, ", ") + "}"
	}
	return e.own(ir.Object, fmt.Sprintf("adamic_node_set_timeout(%s, %s, %s, %d, %s, %s)", callback, e.timerInvoke(call), delay, len(slots), arguments, mask))
}

// timerInvoke is the thunk the provider calls a timeout's callback through: the arguments it holds
// packed by the program's layout for the callback's function type, as a call through the function
// value packs them, padding the optional slots past them, and the result let go as a void call's
// is. A throw leaves adamic_thrown set for the provider's uncaught path, with nothing to let go.
func (e *emitter) timerInvoke(call ir.NodeFSFile) string {
	through := ir.CallClosure{Closure: call.Arguments[0], FunctionType: call.CallbackType}
	count := len(call.Arguments) - 2
	slots := []string{}
	for index := range count {
		slots = append(slots, fmt.Sprintf("arguments[%d]", index))
	}
	adapter := emitter{program: e.program, reuse: e.reuse, indent: 1, scopes: [][]string{{}}, closureDiscards: e.closureDiscards}
	// The count is the call site's own, known here, and a callback that takes none reads no argument.
	adapter.line("(void)argc;")
	adapter.line("(void)arguments;")
	packed := adapter.closureSlots(through, slots, "", fmt.Sprint(count))
	invoked := adapter.packedClosureCall(through, "callback", packed, fmt.Sprint(count))
	if discard := e.program.ClosureDiscard(through); adapter.discarding(discard) {
		adapter.line("adamic_value result = %s;", invoked)
		adapter.line("if (adamic_thrown == NULL) {")
		adapter.line("\t%s", adapter.discard(discard, "callback", "result"))
		adapter.line("}")
	} else {
		adapter.line("(void)%s;", invoked)
	}
	adapter.releaseScopes(0)
	name, exists := e.sharedDeclaration("adamic_timer_invoke", adapter.out.String())
	if !exists {
		e.declarations = append(e.declarations, "static void "+name+"(adamic_closure *callback, size_t argc, const adamic_value *arguments) {\n"+adapter.out.String()+"}")
	}
	return name
}
