package ir

// ProcessCall observes or changes the prelude's process, never a same-named user object.
// Exit and setExitCode can throw while validating a numeric code. A valid exit ends immediately.
type ProcessCall struct {
	Operation string
	Arguments []Expression
	Of        Type
}

func (call ProcessCall) Type() Type { return call.Of }
