package ir

// ProcessCall observes or changes the prelude's process, never a same-named user object.
// cwd, chdir and measure can throw through the ordinary exception cleanup paths.
type ProcessCall struct {
	Operation string
	Arguments []Expression
	Of        Type
}

func (call ProcessCall) Type() Type { return call.Of }
