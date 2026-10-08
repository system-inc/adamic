package ir

// Debugger requests a debugging action if a debugger is attached. Without one,
// it has no effects; native builds emit nothing and JavaScript keeps the statement.
type Debugger struct{}

func (Debugger) statement() {}
