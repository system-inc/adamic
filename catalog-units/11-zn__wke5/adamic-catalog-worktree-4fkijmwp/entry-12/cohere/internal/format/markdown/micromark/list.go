package micromark

import "slices"

// list is micromark-core-commonmark/lib/list.js.
//
// It is assigned in init because the continuation attempts list itself, which Go reports as an
// initialization cycle in a composite literal. constructs.go reads it inside defaultConstructs, after init.
var list *Construct

func init() {
	list = &Construct{
		Continuation: &Construct{Tokenize: tokenizeListContinuation},
		Exit:         tokenizeListEnd,
		Name:         "list",
		Tokenize:     tokenizeListStart,
	}
}

// listItemPrefixWhitespaceConstruct is upstream's file-local construct of the same name.
var listItemPrefixWhitespaceConstruct = &Construct{Partial: true, Tokenize: tokenizeListItemPrefixWhitespace}

// listIndentConstruct is upstream's file-local `indentConstruct`.
var listIndentConstruct = &Construct{Partial: true, Tokenize: tokenizeListIndent}

// To do: `markdown-rs` parses list items on their own and later stitches them together.

// listCodeIndentedDisabled is upstream's `self.parser.constructs.disable.null.includes('codeIndented')`.
func listCodeIndentedDisabled(self *Self) bool {
	return slices.Contains(self.Parser.Constructs.Disable, "codeIndented")
}

func tokenizeListStart(self *Self, effects *Effects, ok State, nok State) State {
	initialSize := 0
	if len(self.Events) > 0 {
		tail := self.Events[len(self.Events)-1]
		if tail.Token.Type == TypeLinePrefix {
			initialSize = utf16Length(tail.Context.SliceSerialize(tail.Token, true))
		}
	}
	size := 0

	var start, inside, atMarker, onBlank, otherPrefix, endOfPrefix State

	start = func(code Code) State {
		// Upstream's `containerState.type`, which is a reserved word here.
		kind := self.ContainerState.listType
		if kind == "" {
			if code == CodeAsterisk || code == CodePlusSign || code == CodeDash {
				kind = TypeListUnordered
			} else {
				kind = TypeListOrdered
			}
		}

		var matches bool
		if kind == TypeListUnordered {
			// A zero marker is upstream's unset one: CodeEof is never a marker.
			matches = self.ContainerState.marker == 0 || code == self.ContainerState.marker
		} else {
			matches = asciiDigit(code)
		}

		if matches {
			if self.ContainerState.listType == "" {
				self.ContainerState.listType = kind
				effects.Enter(kind, &Token{container: true})
			}

			if kind == TypeListUnordered {
				effects.Enter(TypeListItemPrefix, nil)
				if code == CodeAsterisk || code == CodeDash {
					return effects.Check(thematicBreak, nok, atMarker)(code)
				}
				return atMarker(code)
			}

			if !self.IsInterrupt() || code == CodeDigit1 {
				effects.Enter(TypeListItemPrefix, nil)
				effects.Enter(TypeListItemValue, nil)
				return inside(code)
			}
		}

		return nok(code)
	}

	inside = func(code Code) State {
		// `asciiDigit(code) && ++size < 10`: the increment only happens on a digit.
		if asciiDigit(code) {
			size++
			if size < listItemValueSizeMax {
				effects.Consume(code)
				return inside
			}
		}

		var isMarker bool
		if self.ContainerState.marker != 0 {
			isMarker = code == self.ContainerState.marker
		} else {
			isMarker = code == CodeRightParenthesis || code == CodeDot
		}

		if (!self.IsInterrupt() || size < 2) && isMarker {
			effects.Exit(TypeListItemValue)
			return atMarker(code)
		}

		return nok(code)
	}

	atMarker = func(code Code) State {
		effects.Enter(TypeListItemMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeListItemMarker)
		if self.ContainerState.marker == 0 {
			self.ContainerState.marker = code
		}
		// Can’t be empty when interrupting.
		blank := onBlank
		if self.IsInterrupt() {
			blank = nok
		}
		return effects.Check(
			blankLine,
			blank,
			effects.Attempt(listItemPrefixWhitespaceConstruct, endOfPrefix, otherPrefix),
		)
	}

	onBlank = func(code Code) State {
		self.ContainerState.initialBlankLine = true
		initialSize++
		return endOfPrefix(code)
	}

	otherPrefix = func(code Code) State {
		if markdownSpace(code) {
			effects.Enter(TypeListItemPrefixWhitespace, nil)
			effects.Consume(code)
			effects.Exit(TypeListItemPrefixWhitespace)
			return endOfPrefix
		}

		return nok(code)
	}

	endOfPrefix = func(code Code) State {
		self.ContainerState.size = initialSize + utf16Length(self.SliceSerialize(effects.Exit(TypeListItemPrefix), true))
		return ok(code)
	}

	return start
}

func tokenizeListContinuation(self *Self, effects *Effects, ok State, nok State) State {
	var onBlank, notBlank, notInCurrentItem State

	self.ContainerState.closeFlow = false

	onBlank = func(code Code) State {
		self.ContainerState.furtherBlankLines = self.ContainerState.furtherBlankLines || self.ContainerState.initialBlankLine

		// We have a blank line.
		// Still, try to consume at most the items size.
		return factorySpace(effects, ok, TypeListItemIndent, self.ContainerState.size+1)(code)
	}

	notBlank = func(code Code) State {
		if self.ContainerState.furtherBlankLines || !markdownSpace(code) {
			self.ContainerState.furtherBlankLines = false
			self.ContainerState.initialBlankLine = false
			return notInCurrentItem(code)
		}

		self.ContainerState.furtherBlankLines = false
		self.ContainerState.initialBlankLine = false
		return effects.Attempt(listIndentConstruct, ok, notInCurrentItem)(code)
	}

	notInCurrentItem = func(code Code) State {
		// While we do continue, we signal that the flow should be closed.
		self.ContainerState.closeFlow = true
		// As we’re closing flow, we’re no longer interrupting.
		self.SetInterrupt(false)
		// Always populated by defaults.

		limit := tabSize
		if listCodeIndentedDisabled(self) {
			limit = 0
		}
		return factorySpace(effects, effects.Attempt(list, ok, nok), TypeLinePrefix, limit)(code)
	}

	return effects.Check(blankLine, onBlank, notBlank)
}

func tokenizeListIndent(self *Self, effects *Effects, ok State, nok State) State {
	var afterPrefix State

	afterPrefix = func(code Code) State {
		if len(self.Events) > 0 {
			tail := self.Events[len(self.Events)-1]
			if tail.Token.Type == TypeListItemIndent &&
				utf16Length(tail.Context.SliceSerialize(tail.Token, true)) == self.ContainerState.size {
				return ok(code)
			}
		}
		return nok(code)
	}

	return factorySpace(effects, afterPrefix, TypeListItemIndent, self.ContainerState.size+1)
}

func tokenizeListEnd(self *Self, effects *Effects) {
	effects.Exit(self.ContainerState.listType)
}

func tokenizeListItemPrefixWhitespace(self *Self, effects *Effects, ok State, nok State) State {
	var afterPrefix State

	afterPrefix = func(code Code) State {
		if !markdownSpace(code) && len(self.Events) > 0 &&
			self.Events[len(self.Events)-1].Token.Type == TypeListItemPrefixWhitespace {
			return ok(code)
		}
		return nok(code)
	}

	// Always populated by defaults.
	limit := tabSize + 1
	if listCodeIndentedDisabled(self) {
		limit = 0
	}
	return factorySpace(effects, afterPrefix, TypeListItemPrefixWhitespace, limit)
}
