package micromark

// lineEnding is micromark-core-commonmark/lib/line-ending.js.
var lineEnding = &Construct{Name: "lineEnding", Tokenize: tokenizeLineEnding}

func tokenizeLineEnding(self *Self, effects *Effects, ok State, nok State) State {
	return func(code Code) State {
		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		return factorySpace(effects, ok, TypeLinePrefix, 0)
	}
}
