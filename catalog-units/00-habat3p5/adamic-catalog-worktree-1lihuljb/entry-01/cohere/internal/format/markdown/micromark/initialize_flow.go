package micromark

// flowInitial is upstream's flow, micromark/lib/initialize/flow.js.
var flowInitial *InitialConstruct

func init() { flowInitial = &InitialConstruct{Tokenize: initializeFlow} }

func initializeFlow(self *Self, effects *Effects) State {
	var initial, atBlankEnding, afterConstruct State

	// Upstream's two states are hoisted function declarations, so they exist before the attempts that
	// return to them are built. Here they are assigned first and the chain is built once, after, because
	// hooks and factories keep state in their closures and building them per line would reset it.
	atBlankEnding = func(code Code) State {
		if code == CodeEof {
			effects.Consume(code)
			return nil
		}

		effects.Enter(TypeLineEndingBlank, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEndingBlank)
		self.CurrentConstruct = nil
		return initial
	}

	afterConstruct = func(code Code) State {
		if code == CodeEof {
			effects.Consume(code)
			return nil
		}

		effects.Enter(TypeLineEnding, nil)
		effects.Consume(code)
		effects.Exit(TypeLineEnding)
		self.CurrentConstruct = nil
		return initial
	}

	initial = effects.Attempt(
		// Try to parse a blank line.
		blankLine,
		atBlankEnding,
		// Try to parse initial flow (essentially, only code).
		effects.Attempt(
			self.Parser.Constructs.FlowInitial,
			afterConstruct,
			factorySpace(
				effects,
				effects.Attempt(
					self.Parser.Constructs.Flow,
					afterConstruct,
					effects.Attempt(content, afterConstruct, nil),
				),
				TypeLinePrefix,
				0,
			),
		),
	)

	return initial
}
