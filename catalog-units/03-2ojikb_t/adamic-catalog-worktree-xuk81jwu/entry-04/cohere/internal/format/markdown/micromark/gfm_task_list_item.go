package micromark

// gfmTaskListItem is micromark-extension-gfm-task-list-item/lib/syntax.js.

// The extension's own token types.
const (
	typeTaskListCheck               = "taskListCheck"
	typeTaskListCheckMarker         = "taskListCheckMarker"
	typeTaskListCheckValueChecked   = "taskListCheckValueChecked"
	typeTaskListCheckValueUnchecked = "taskListCheckValueUnchecked"
)

// tasklistCheck is upstream's file-local `tasklistCheck`.
var tasklistCheck = &Construct{Name: "tasklistCheck", Tokenize: tokenizeTasklistCheck}

// tasklistSpaceThenNonSpace is the anonymous `{tokenize: spaceThenNonSpace}` upstream checks with. It is
// not marked partial upstream, and check does not read partial, so neither is this.
var tasklistSpaceThenNonSpace = &Construct{Tokenize: tokenizeSpaceThenNonSpace}

// gfmTaskListItemExtension is upstream's gfmTaskListItem().
func gfmTaskListItemExtension() *Extension {
	return &Extension{
		Text: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeLeftSquareBracket: {tasklistCheck},
		}},
	}
}

func tokenizeTasklistCheck(self *Self, effects *Effects, ok State, nok State) State {
	var open, inside, close, after State

	// At start of task list item check.
	//
	// > | * [x] y.
	//       ^
	open = func(code Code) State {
		// `self.previous !== null`: Previous is CodeEof (upstream's null) until a code is consumed, and a
		// real NUL never reaches a tokenizer.
		if
		// Exit if there’s stuff before.
		self.Previous != CodeEof ||
			// Exit if not in the first content that is the first child of a list item. subtokenize sets
			// this on the text tokenizer while it writes a chunk marked isInFirstContentOfListItem.
			!self.gfmTasklistFirstContentOfListItem {
			return nok(code)
		}
		effects.Enter(typeTaskListCheck, nil)
		effects.Enter(typeTaskListCheckMarker, nil)
		effects.Consume(code)
		effects.Exit(typeTaskListCheckMarker)
		return inside
	}

	// In task list item check.
	//
	// > | * [x] y.
	//        ^
	inside = func(code Code) State {
		// Currently we match how GH works in files.
		// To match how GH works in comments, use `markdownSpace` (`[\t ]`) instead
		// of `markdownLineEndingOrSpace` (`[\t\n\r ]`).
		if markdownLineEndingOrSpace(code) {
			effects.Enter(typeTaskListCheckValueUnchecked, nil)
			effects.Consume(code)
			effects.Exit(typeTaskListCheckValueUnchecked)
			return close
		}
		// `X` (88) or `x` (120).
		if code == 88 || code == 120 {
			effects.Enter(typeTaskListCheckValueChecked, nil)
			effects.Consume(code)
			effects.Exit(typeTaskListCheckValueChecked)
			return close
		}
		return nok(code)
	}

	// At close of task list item check.
	//
	// > | * [x] y.
	//         ^
	close = func(code Code) State {
		if code == CodeRightSquareBracket {
			effects.Enter(typeTaskListCheckMarker, nil)
			effects.Consume(code)
			effects.Exit(typeTaskListCheckMarker)
			effects.Exit(typeTaskListCheck)
			return after
		}
		return nok(code)
	}

	after = func(code Code) State {
		// EOL in paragraph means there must be something else after it.
		if markdownLineEnding(code) {
			return ok(code)
		}

		// Space or tab?
		// Check what comes after.
		if markdownSpace(code) {
			return effects.Check(tasklistSpaceThenNonSpace, ok, nok)(code)
		}

		// EOF, or non-whitespace, both wrong.
		return nok(code)
	}

	return open
}

func tokenizeSpaceThenNonSpace(self *Self, effects *Effects, ok State, nok State) State {
	var after State

	// After whitespace, after task list item check.
	//
	// > | * [x] y.
	//           ^
	after = func(code Code) State {
		// EOF means there was nothing, so bad.
		// EOL means there’s content after it, so good.
		// Impossible to have more spaces.
		// Anything else is good.
		if code == CodeEof {
			return nok(code)
		}
		return ok(code)
	}

	return factorySpace(effects, after, TypeWhitespace, 0)
}
