package native

// These adapters are emitted with the library binding. Runtime's primitive C
// sources remain the single owners of status, clock and OS observations.
func (e *emitter) hostHelper(name string) {
	if e.hostHelpers == nil {
		e.hostHelpers = map[string]bool{}
		e.declarations = append(e.declarations, "#include <errno.h>\n#include <stdlib.h>\n#include <string.h>")
	}
	if e.hostHelpers[name] {
		return
	}
	e.hostHelpers[name] = true
	switch name {
	case "error":
		e.declarations = append(e.declarations, hostErrorAdapter)
	case "exitCode":
		e.declarations = append(e.declarations, `static adamic_maybe_number adamic_library_exit_code(void) {
 int32_t code = 0;
 bool present = adamic_host_exit_code(&code);
 return (adamic_maybe_number){.present = present, .number = (double)code};
}`)
	case "setExitCode":
		e.declarations = append(e.declarations, `static void adamic_library_set_exit_code(adamic_maybe_number code) {
 if (code.present && (!isfinite(code.number) || trunc(code.number) != code.number)) {
  static adamic_string prefix = ADAMIC_STRING("The value of \"code\" is out of range. It must be an integer. Received ");
  static adamic_string name = ADAMIC_STRING("RangeError");
  static adamic_string error_code = ADAMIC_STRING("ERR_OUT_OF_RANGE");
  adamic_string *number = adamic_string_from_number(code.number);
  adamic_string *message = adamic_string_concat(2, (adamic_string *const[]){&prefix, number});
  adamic_node_error(&name, message, &error_code);
  adamic_release(number); adamic_release(message);
  return;
 }
 adamic_host_set_exit_code(code.present, code.present ? (int32_t)adamic_bitwise_or(code.number, 0) : 0);
}`)
	case "cwd":
		e.hostHelper("error")
		e.declarations = append(e.declarations, `static adamic_string *adamic_library_cwd_value;
static bool adamic_library_cwd_registered;
static void adamic_library_cwd_finish(void) { adamic_release(adamic_library_cwd_value); }
static adamic_string *adamic_library_cwd(void) {
 if (adamic_library_cwd_value != NULL) { return adamic_retain(adamic_library_cwd_value); }
 adamic_string *result = adamic_host_cwd();
 if (result == NULL) { adamic_library_host_error(errno, "uv_cwd"); }
 if (result != NULL) {
  adamic_library_cwd_value = adamic_retain(result);
  if (!adamic_library_cwd_registered) { atexit(adamic_library_cwd_finish); adamic_library_cwd_registered = true; }
 }
 return result;
}`)
	case "chdir":
		e.hostHelper("cwd")
		e.declarations = append(e.declarations, `static void adamic_library_chdir(const adamic_string *path) {
 adamic_node_chdir(path);
 if (adamic_thrown == NULL) { adamic_release(adamic_library_cwd_value); adamic_library_cwd_value = NULL; }
}`)
	case "execPath":
		e.hostHelper("error")
		e.declarations = append(e.declarations, "#define ADAMIC_HOST_EXECUTABLE_LIBRARIES 1")
		e.declarations = append(e.declarations, `static adamic_string *adamic_library_exec_path(void) {
 adamic_string *result = adamic_host_executing_file_path();
 if (result == NULL) { adamic_library_host_error(errno, "execPath"); }
 return result;
}`)
	case "dirname":
		e.hostHelper("execPath")
		e.declarations = append(e.declarations, `static adamic_string *adamic_library_dirname(void) {
 adamic_string *path = adamic_library_exec_path();
 if (path == NULL) { return NULL; }
 adamic_string *result = adamic_node_path_dirname(path);
 adamic_release(path);
 return result;
}`)
	case "stdout":
		e.declarations = append(e.declarations, `static adamic_object *adamic_library_stdout_object;
static void adamic_library_stdout_finish(void) { adamic_release(adamic_library_stdout_object); }
static adamic_object *adamic_library_stdout(void) {
 if (adamic_library_stdout_object == NULL) {
  static const adamic_shape shape = {0, NULL, NULL, NULL};
  adamic_library_stdout_object = adamic_object_new(&shape);
  atexit(adamic_library_stdout_finish);
 }
 return adamic_retain(adamic_library_stdout_object);
}`)
	case "stdoutWrite":
		e.hostHelper("error")
		e.declarations = append(e.declarations, `static void adamic_library_stdout_write(const adamic_string *text) {
 if (!adamic_write_raw(adamic_stdout, text)) { adamic_library_host_error(errno == 0 ? EPIPE : errno, "write"); }
}`)
	case "argv", "execArgv":
		e.hostHelper("error")
		primitive := "adamic_host_argv"
		if name == "execArgv" {
			primitive = "adamic_host_exec_argv"
		}
		e.declarations = append(e.declarations, "static adamic_array *adamic_library_"+name+"_value;\n"+
			"static void adamic_library_"+name+"_finish(void) { adamic_release(adamic_library_"+name+"_value); }\n"+
			"static adamic_array *adamic_library_"+name+"(void) {\n"+
			" if (adamic_library_"+name+"_value == NULL) {\n"+
			"  adamic_library_"+name+"_value = "+primitive+"();\n"+
			"  if (adamic_library_"+name+"_value == NULL) { adamic_library_host_error(errno, \"execPath\"); return NULL; }\n"+
			"  atexit(adamic_library_"+name+"_finish);\n }\n"+
			" return adamic_retain(adamic_library_"+name+"_value);\n}")
	case "memoryUsage":
		e.hostHelper("error")
		e.declarations = append(e.declarations, `static adamic_object *adamic_library_memory_usage(void) {
 adamic_host_memory memory;
 if (!adamic_host_memory_usage(&memory)) { adamic_library_host_error(errno, "memoryUsage"); return NULL; }
 static const char *const names[] = {"rss", "heapTotal", "heapUsed", "external", "arrayBuffers"};
 static const bool references[] = {false, false, false, false, false};
 static const adamic_shape shape = {5, names, references, NULL};
 adamic_object *result = adamic_object_new(&shape);
 result->slots[0].number = memory.rss;
 result->slots[1].number = memory.heapTotal;
 result->slots[2].number = memory.heapUsed;
 result->slots[3].number = memory.external;
 result->slots[4].number = memory.arrayBuffers;
 return result;
}`)
	case "performance":
		e.declarations = append(e.declarations, `static adamic_value adamic_library_now_method(adamic_closure *self, adamic_value *args) {
 (void)self; (void)args;
 return (adamic_value){.number = adamic_host_performance_now()};
}
static adamic_object *adamic_library_performance(void) {
 static adamic_closure now = {{0, adamic_kind_closure, 0}, adamic_library_now_method, 0};
 adamic_object *result = adamic_node_performance();
 result->slots[0].number = adamic_host_time_origin();
 result->slots[1].reference = &now;
 return result;
}`)
	default:
		panic("unknown startup helper: " + name)
	}
}

const hostErrorAdapter = `static void adamic_library_host_error(int error, const char *operation) {
 const char *code, *reason;
 switch (error) {
 case ENOENT: code = "ENOENT"; reason = "no such file or directory"; break;
 case EACCES: code = "EACCES"; reason = "permission denied"; break;
 case ENOTDIR: code = "ENOTDIR"; reason = "not a directory"; break;
 case ENAMETOOLONG: code = "ENAMETOOLONG"; reason = "name too long"; break;
 case ELOOP: code = "ELOOP"; reason = "too many symbolic links encountered"; break;
 case ENOMEM: code = "ENOMEM"; reason = "not enough memory"; break;
 case EIO: code = "EIO"; reason = "i/o error"; break;
 case ENOSYS: code = "ENOSYS"; reason = "function not implemented"; break;
 case EPIPE: code = "EPIPE"; reason = "broken pipe"; break;
 case EBADF: code = "EBADF"; reason = "bad file descriptor"; break;
 case EINVAL: code = "EINVAL"; reason = "invalid argument"; break;
 default: adamic_panic("NotYet: host system errno", sizeof "NotYet: host system errno" - 1);
 }
 if (error == ENOENT && strcmp(operation, "uv_cwd") == 0) {
  reason = "process.cwd failed with error no such file or directory, the current working directory was likely removed without changing the working directory";
 }
 static const char *const names[] = {"name", "message", "code", "errno", "syscall"};
 static const bool references[] = {true, true, true, false, true};
 static const adamic_shape shape = {5, names, references, NULL};
 static adamic_string name = ADAMIC_STRING("Error"), colon = ADAMIC_STRING(": "), comma = ADAMIC_STRING(", "), space = ADAMIC_STRING(" ");
 adamic_string *prefix = adamic_decode_utf8((const unsigned char *)code, strlen(code));
 adamic_string *detail = adamic_decode_utf8((const unsigned char *)reason, strlen(reason));
 adamic_string *syscall = adamic_decode_utf8((const unsigned char *)operation, strlen(operation));
 adamic_object *result = adamic_object_new(&shape);
 result->slots[0].reference = &name;
 result->slots[1].reference = strcmp(operation, "write") == 0
  ? adamic_string_concat(3, (adamic_string *const[]){syscall, &space, prefix})
  : adamic_string_concat(5, (adamic_string *const[]){prefix, &colon, detail, &comma, syscall});
 result->slots[2].reference = prefix;
 result->slots[3].number = -(double)error;
 result->slots[4].reference = syscall;
 adamic_release(detail);
 adamic_error_tag(result);
 adamic_thrown = result;
}`
