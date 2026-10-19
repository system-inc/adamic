package ir

// NodeFSFile is a synchronous host operation. Arguments are evaluated in order,
// borrowed during the call, and never retained by the host. Results are owned.
// Environment assignment and delete mutate live process state, returning the assigned string
// and true respectively. Every operation except environment operations, exists and the Stats/Date observations may throw.
type NodeFSFile struct {
	Operation string
	Arguments []Expression
	Of        Type
}

func (call NodeFSFile) Type() Type { return call.Of }
func (call NodeFSFile) MayThrow() bool {
	switch call.Operation {
	case "host_env", "host_env_set", "host_env_delete", "exists", "is_file", "is_directory", "is_symbolic_link", "date_new", "date_time":
		return false
	}
	return true
}
