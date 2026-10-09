package micromark

// pendingConstruct stands in for a construct not yet ported: it never matches, so input parses as if
// the construct did not exist. Each one lives in its own pending_*.go file, deleted with the port.
func pendingConstruct(name string) *Construct {
	return &Construct{Name: name, Tokenize: func(self *Self, effects *Effects, ok State, nok State) State {
		return nok
	}}
}
