package ir

// NodeFSFile is a host operation. Arguments are evaluated in order and borrowed
// during the call; watchFile, fs.watch and FSWatcher.on additionally retain their
// listener until cancellation, close or loop shutdown (RetainsListener). Results are owned.
// fs.watch's result is an FSWatcher host object the runtime also holds while it's open.
// Environment assignment and delete mutate live process state, returning the assigned string
// and true respectively. Every operation except raw output, environment operations and platform
// reads, exists and the Stats/Date observations may throw. host_require_missing, a require of a module
// that isn't installed, always throws: Node's MODULE_NOT_FOUND for its specifier.
// host_inspector_unavailable also always throws, on every target, including WASI.
type NodeFSFile struct {
	Operation string
	Arguments []Expression
	Of        Type
}

func (call NodeFSFile) Type() Type { return call.Of }

// RetainsListener reports whether the host keeps a function value it's handed past the call and
// calls it later from the loop: the scheduler must run, and what the listener reaches escapes.
func (call NodeFSFile) RetainsListener() bool {
	return call.Operation == "host_watch_file" || call.Operation == "host_watch" || call.Operation == "host_watch_on"
}

func (call NodeFSFile) MayThrow() bool {
	switch call.Operation {
	case "host_stdout_columns", "host_stdout_is_tty", "host_stdout_write", "host_env", "host_platform", "host_env_set", "host_env_delete", "exists", "is_file", "is_directory", "is_symbolic_link", "date_new", "date_now", "date_time", "time_now":
		return false
	}
	return true
}
