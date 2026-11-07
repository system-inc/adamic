package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nodeFSFile(call ir.NodeFSFile) string {
	args := make([]string, len(call.Arguments))
	for i, arg := range call.Arguments {
		args[i] = e.value(arg)
	}
	switch call.Operation {
	case "set_timeout":
		return "adamicSetTimeout(" + args[0] + ", " + args[1] + ", [" + strings.Join(args[2:], ", ") + "])"
	case "clear_timeout":
		return "adamicClearTimeout(" + args[0] + ")"
	case "host_watch_file":
		return "adamicNodeFSFile.watchFile(" + args[0] + ", {persistent:" + args[1] + ", interval:" + args[2] + "}, " + "adamicWatchFileListener(" + args[3] + "))"
	case "host_watch":
		return "adamicFSWatch(" + strings.Join(args, ", ") + ")"
	case "host_watch_on":
		return "adamicFSWatchOn(" + args[0] + ", " + args[1] + ")"
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
		return "(() => { throw Object.assign(new Error(\"Cannot find module '\" + " + args[0] + " + \"'\"), {code: 'MODULE_NOT_FOUND'}); })()"
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
		return "adamicNodeFSFile.readSync(" + args[0] + ", " + args[1] + ", " + args[2] + ", " + args[3] + ", " + args[4] + ")"
	case "write":
		return "adamicNodeFSFile.writeSync(" + args[0] + ", " + args[1] + ", " + args[2] + ", 'utf8')"
	case "close":
		return "adamicNodeFSFile.closeSync(" + args[0] + ")"
	case "write_file", "write_fd", "write_buffer", "write_buffer_fd":
		return "adamicNodeFSFile.writeFileSync(" + args[0] + ", " + args[1] + ", {flag:" + args[2] + ", mode:" + args[3] + ", flush:" + args[4] + ", encoding:'utf8'})"
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

// setTimeout called from source is Node's own, its Timeout Node's, and its callback a function in
// the closure calling convention, handed the extra arguments Node hands it. A delay Node would warn
// for and make 1 ms stops here as native's timer provider does (timers_impl.h), in the same words,
// a run-time NotYet (#995spev).
// clearTimeout is Node's own after native's check (#995spev): an object that isn't a Timeout, known
// by the prototype a probe timer has, but holds a truthy own _onTimeout stops before Node would
// unenroll it.
const nodeTimersRuntime = `const adamicSetTimeout = (callback, delay, values) => {
 if (!(delay >= 0 && delay <= 2147483647)) panic("a timer delay outside Node's range (NaN, negative or past 2^31 - 1 ms) isn't built yet");
 return globalThis.setTimeout((...handed) => { adamicCall(callback, handed); }, delay, ...values);
};
const adamicTimeoutPrototype = (() => { const probe = globalThis.setTimeout(() => {}, 1); globalThis.clearTimeout(probe); return Object.getPrototypeOf(probe); })();
const adamicClearTimeout = (value) => {
 if (typeof value === 'object' && value !== null && Object.getPrototypeOf(value) !== adamicTimeoutPrototype && Object.hasOwn(value, '_onTimeout') && value._onTimeout) panic("clearTimeout of an object that isn't a timer but holds _onTimeout isn't built yet");
 globalThis.clearTimeout(value);
};
`

// fs.watch's FSWatcher stand-in: Node's watcher behind an object whose close and on are own
// closures in the closure calling convention, as native's shape holds them, given a prototype that
// isn't Object's so a run-time key read stops on it as native's host shape does. One adapter per
// listener keeps its identity, as watchFile's does, and hands it Node's own arguments. recursive,
// and on through a view with another event or after close, stop by name, as native's do; the
// direct on("error"), which lowering recognized, is Node's own after close too.
const nodeWatchRuntime = `const adamicFSWatchListeners = new WeakMap();
const adamicFSWatchListener = callback => {
 if (!adamicFSWatchListeners.has(callback)) {
  adamicFSWatchListeners.set(callback, (...values) => adamicCall(callback, values));
 }
 return adamicFSWatchListeners.get(callback);
};
const adamicFSWatchOns = new WeakMap();
class AdamicFSWatcher {
 constructor(watcher) {
  let closed = false;
  const on = callback => { watcher.on('error', adamicFSWatchListener(callback)); return this; };
  this.close = {code: () => { closed = true; watcher.close(); }};
  this.on = {code: (closure, values) => {
   if (values[0] !== 'error') panic("NotYet: FSWatcher.on with an event other than 'error'");
   if (closed) panic('NotYet: FSWatcher.on through a structural view after close');
   return on(values[1]);
  }};
  adamicFSWatchOns.set(this, on);
 }
}
const adamicFSWatch = (path, persistent, recursive, callback) => {
 if (recursive) panic('NotYet: fs.watch with recursive: true');
 return new AdamicFSWatcher(adamicNodeFSFile.watch(path, {persistent}, adamicFSWatchListener(callback)));
};
const adamicFSWatchOn = (watcher, callback) => adamicFSWatchOns.get(watcher)(callback);
`
