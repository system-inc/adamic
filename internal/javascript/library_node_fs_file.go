package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (javascriptEmitter *emitter) nodeFileSystemFile(call ir.NodeFSFile) string {
	args := make([]string, len(call.Arguments))
	for index, arg := range call.Arguments {
		args[index] = javascriptEmitter.value(arg)
	}
	switch call.Operation {
	case "host_watch_file":
		return "adamicNodeFSFile.watchFile(" +
			args[0] +
			", {persistent:" +
			args[1] +
			", interval:" +
			args[2] +
			"}, " + "adamicWatchFileListener(" +
			args[3] +
			"))"
	case "host_unwatch_file":
		if len(args) == 2 {
			args[1] = "adamicWatchFileListener(" + args[1] + ")"
		}
		return "adamicNodeFSFile.unwatchFile(" + strings.Join(args, ", ") + ")"
	case "host_stdout_columns":
		return "adamicNodeProcess.stdout.columns"
	case "host_stdout_is_tty":
		return "adamicNodeProcess.stdout.isTTY"
	case "host_stdout_write":
		return "adamicNodeProcess.stdout.write(" + args[0] + ")"
	case "host_realpath":
		return "adamicNodeFSFile.realpathSync.native(" + args[0] + ")"
	case "host_env_set":
		return "adamicHostEnvironment(" + args[0] + ", 'set', " + args[1] + ")"
	case "host_env_delete":
		return "adamicHostEnvironment(" + args[0] + ", 'delete')"
	case "host_env":
		return "adamicHostEnvironment(" + args[0] + ")"
	case "host_join":
		return "adamicNodePath.join(" + strings.Join(args, ", ") + ")"
	case "host_tmpdir":
		return "adamicNodeOS.tmpdir()"
	case "host_cwd":
		return "adamicNodeProcess.cwd()"
	case "host_platform":
		return "adamicNodeProcess.platform"
	case "host_chdir":
		return "adamicNodeProcess.chdir(" + args[0] + ")"
	case "host_inspector_unavailable":
		return "(() => { class Unavailable extends Error { constructor() { super('Inspector is not available'); this.code = 'ERR_INSPECTOR_NOT_AVAILABLE'; } toString() { return this.name + ' [' + this.code + ']: ' + this.message; } } throw new Unavailable(); })()"
	case "host_require_missing":
		// Node's error for a module that isn't installed, with the first line of its message: native's
		// answer, which the JavaScript backend gives too, since the generated module could find one that is.
		return "(() => { throw Object.assign(new Error(\"Cannot find module '\" + " +
			args[0] +
			" + \"'\"), {code: 'MODULE_NOT_FOUND'}); })()"
	case "date_new":
		return "new Date(" + args[0] + ")"
	case "date_now":
		return "new Date()"
	case "time_now":
		return "Date.now()"
	case "date_time":
		return "(" + args[0] + ").getTime()"
	case "is_file":
		return "(" + args[0] + ").isFile()"
	case "is_directory":
		return "(" + args[0] + ").isDirectory()"
	case "is_symbolic_link":
		return "(" + args[0] + ").isSymbolicLink()"
	case "read_buffer", "read_buffer_fd":
		return "adamicNodeFSFile.readFileSync(" + args[0] + ", {flag:" + args[1] + "})"
	case "read_file", "read_fd":
		return "adamicNodeFSFile.readFileSync(" + args[0] + ", {encoding:'utf8', flag:" + args[1] + "})"
	case "open":
		return "adamicNodeFSFile.openSync(" + args[0] + ", " + args[1] + ", " + args[2] + ")"
	case "read_sync":
		return "adamicNodeFSFile.readSync(" +
			args[0] +
			", " +
			args[1] +
			", " +
			args[2] +
			", " +
			args[3] +
			", " +
			args[4] +
			")"
	case "write":
		return "adamicNodeFSFile.writeSync(" + args[0] + ", " + args[1] + ", " + args[2] + ", 'utf8')"
	case "close":
		return "adamicNodeFSFile.closeSync(" + args[0] + ")"
	case "write_file", "write_fd", "write_buffer", "write_buffer_fd":
		return "adamicNodeFSFile.writeFileSync(" +
			args[0] +
			", " +
			args[1] +
			", {flag:" +
			args[2] +
			", mode:" +
			args[3] +
			", flush:" +
			args[4] +
			", encoding:'utf8'})"
	case "exists":
		return "adamicNodeFSFile.existsSync(" + args[0] + ")"
	case "stat":
		return "adamicNodeFSFile.statSync(" + args[0] + ", {throwIfNoEntry:" + args[1] + "})"
	case "mkdir":
		return "adamicNodeFSFile.mkdirSync(" + args[0] + ", {recursive:" + args[1] + ", mode:" + args[2] + "})"
	case "mkdtemp":
		return "adamicNodeFSFile.mkdtempSync(" + args[0] + ")"
	case "rm":
		return "adamicNodeFSFile.rmSync(" + args[0] + ", {recursive:" + args[1] + ", force:" + args[2] + "})"
	case "unlink":
		return "adamicNodeFSFile.unlinkSync(" + args[0] + ")"
	case "utimes", "utimes_dates", "utimes_atime_date", "utimes_mtime_date":
		return "adamicNodeFSFile.utimesSync(" + args[0] + ", " + args[1] + ", " + args[2] + ")"
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
