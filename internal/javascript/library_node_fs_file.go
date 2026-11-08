package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (javascriptEmitter *emitter) nodeFileSystemFile(call ir.NodeFSFile) string {
	commandArguments := make([]string, len(call.Arguments))
	for index, argument := range call.Arguments {
		commandArguments[index] = javascriptEmitter.value(argument)
	}
	switch call.Operation {
	case "host_watch_file":
		return "adamicNodeFSFile.watchFile(" +
			commandArguments[0] +
			", {persistent:" +
			commandArguments[1] +
			", interval:" +
			commandArguments[2] +
			"}, " + "adamicWatchFileListener(" +
			commandArguments[3] +
			"))"
	case "host_unwatch_file":
		if len(commandArguments) == 2 {
			commandArguments[1] = "adamicWatchFileListener(" + commandArguments[1] + ")"
		}
		return "adamicNodeFSFile.unwatchFile(" + strings.Join(commandArguments, ", ") + ")"
	case "host_stdout_columns":
		return "adamicNodeProcess.stdout.columns"
	case "host_stdout_is_tty":
		return "adamicNodeProcess.stdout.isTTY"
	case "host_stdout_write":
		return "adamicNodeProcess.stdout.write(" + commandArguments[0] + ")"
	case "host_realpath":
		return "adamicNodeFSFile.realpathSync.native(" + commandArguments[0] + ")"
	case "host_env_set":
		return "adamicHostEnvironment(" + commandArguments[0] + ", 'set', " + commandArguments[1] + ")"
	case "host_env_delete":
		return "adamicHostEnvironment(" + commandArguments[0] + ", 'delete')"
	case "host_env":
		return "adamicHostEnvironment(" + commandArguments[0] + ")"
	case "host_join":
		return "adamicNodePath.join(" + strings.Join(commandArguments, ", ") + ")"
	case "host_tmpdir":
		return "adamicNodeOS.tmpdir()"
	case "host_cwd":
		return "adamicNodeProcess.cwd()"
	case "host_platform":
		return "adamicNodeProcess.platform"
	case "host_chdir":
		return "adamicNodeProcess.chdir(" + commandArguments[0] + ")"
	case "host_inspector_unavailable":
		return "(() => { class Unavailable extends Error { constructor() { super('Inspector is not available'); this.code = 'ERR_INSPECTOR_NOT_AVAILABLE'; } toString() { return this.name + ' [' + this.code + ']: ' + this.message; } } throw new Unavailable(); })()"
	case "host_require_missing":
		// Node's error for a module that isn't installed, with the first line of its message: native's
		// answer, which the JavaScript backend gives too, since the generated module could find one that is.
		return "(() => { throw Object.assign(new Error(\"Cannot find module '\" + " +
			commandArguments[0] +
			" + \"'\"), {code: 'MODULE_NOT_FOUND'}); })()"
	case "date_new":
		return "new Date(" + commandArguments[0] + ")"
	case "date_now":
		return "new Date()"
	case "time_now":
		return "Date.now()"
	case "date_time":
		return "(" + commandArguments[0] + ").getTime()"
	case "is_file":
		return "(" + commandArguments[0] + ").isFile()"
	case "is_directory":
		return "(" + commandArguments[0] + ").isDirectory()"
	case "is_symbolic_link":
		return "(" + commandArguments[0] + ").isSymbolicLink()"
	case "read_buffer", "read_buffer_fd":
		return "adamicNodeFSFile.readFileSync(" + commandArguments[0] + ", {flag:" + commandArguments[1] + "})"
	case "read_file", "read_fd":
		return "adamicNodeFSFile.readFileSync(" +
			commandArguments[0] +
			", {encoding:'utf8', flag:" +
			commandArguments[1] +
			"})"
	case "open":
		return "adamicNodeFSFile.openSync(" +
			commandArguments[0] +
			", " +
			commandArguments[1] +
			", " +
			commandArguments[2] +
			")"
	case "read_sync":
		return "adamicNodeFSFile.readSync(" +
			commandArguments[0] +
			", " +
			commandArguments[1] +
			", " +
			commandArguments[2] +
			", " +
			commandArguments[3] +
			", " +
			commandArguments[4] +
			")"
	case "write":
		return "adamicNodeFSFile.writeSync(" +
			commandArguments[0] +
			", " +
			commandArguments[1] +
			", " +
			commandArguments[2] +
			", 'utf8')"
	case "close":
		return "adamicNodeFSFile.closeSync(" + commandArguments[0] + ")"
	case "write_file", "write_fd", "write_buffer", "write_buffer_fd":
		return "adamicNodeFSFile.writeFileSync(" +
			commandArguments[0] +
			", " +
			commandArguments[1] +
			", {flag:" +
			commandArguments[2] +
			", mode:" +
			commandArguments[3] +
			", flush:" +
			commandArguments[4] +
			", encoding:'utf8'})"
	case "exists":
		return "adamicNodeFSFile.existsSync(" + commandArguments[0] + ")"
	case "stat":
		return "adamicNodeFSFile.statSync(" + commandArguments[0] + ", {throwIfNoEntry:" + commandArguments[1] + "})"
	case "mkdir":
		return "adamicNodeFSFile.mkdirSync(" +
			commandArguments[0] +
			", {recursive:" +
			commandArguments[1] +
			", mode:" +
			commandArguments[2] +
			"})"
	case "mkdtemp":
		return "adamicNodeFSFile.mkdtempSync(" + commandArguments[0] + ")"
	case "rm":
		return "adamicNodeFSFile.rmSync(" +
			commandArguments[0] +
			", {recursive:" +
			commandArguments[1] +
			", force:" +
			commandArguments[2] +
			"})"
	case "unlink":
		return "adamicNodeFSFile.unlinkSync(" + commandArguments[0] + ")"
	case "utimes", "utimes_dates", "utimes_atime_date", "utimes_mtime_date":
		return "adamicNodeFSFile.utimesSync(" +
			commandArguments[0] +
			", " +
			commandArguments[1] +
			", " +
			commandArguments[2] +
			")"
	}
	panic("javascript: unknown fs file operation " + call.Operation)
}

// Node retains real functions; compiled function values use the closure ABI.
// One adapter per closure preserves unwatchFile's listener identity.
const nodeWatchFileRuntime = `const adamicWatchFileListeners = new WeakMap();
const adamicWatchFileListener = callback => {
 if (!adamicWatchFileListeners.has(callback)) {
  adamicWatchFileListeners.set(callback, (curr, prev) => adamicCall(callback, [curr, prev]));
 }
 return adamicWatchFileListeners.get(callback);
};
`
