package ir

// NodeFSFile is a host operation. Arguments are evaluated in order and borrowed
// during the call; watchFile additionally retains its listener until cancellation
// or loop shutdown. Results are owned.
// Environment assignment and delete mutate live process state, returning the assigned string
// and true respectively. Every operation except raw output, environment operations and platform
// reads, exists and the Stats/Date observations may throw. host_require_missing, a require of a module
// that isn't installed, always throws: Node's MODULE_NOT_FOUND for its specifier.
type NodeFSFile struct {
	Operation string
	Arguments []Expression
	Of        Type
}

func (call NodeFSFile) Type() Type { return call.Of }
func (call NodeFSFile) MayThrow() bool {
	switch call.Operation {
	case "host_stdout_columns", "host_stdout_is_tty", "host_stdout_write", "host_env", "host_platform", "host_env_set", "host_env_delete", "exists", "is_file", "is_directory", "is_symbolic_link", "date_new", "date_now", "date_time", "time_now":
		return false
	}
	return true
}
