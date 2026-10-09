package micromark

// The string and text content types, micromark/lib/initialize/text.js. String is the content of
// titles, destinations and the like (escapes and references only); text is everything inline.

// resolveTextConstruct is upstream's `resolver`, the data merger insideSpan runs.
var resolveTextConstruct = &Construct{ResolveAll: &Resolver{Resolve: createTextResolver(nil)}}

var stringInitial, textInitial *InitialConstruct

func init() {
	stringInitial = initializeTextFactory(ContentTypeString)
	textInitial = initializeTextFactory(ContentTypeText)
}

func initializeTextFactory(field string) *InitialConstruct {
	var extra func(events []Event, context *TokenizeContext) []Event
	if field == ContentTypeText {
		extra = resolveAllLineSuffixes
	}

	return &InitialConstruct{
		ResolveAll: &Resolver{Resolve: createTextResolver(extra)},
		Tokenize: func(self *Self, effects *Effects) State {
			constructs := self.Parser.Constructs.String
			if field == ContentTypeText {
				constructs = self.Parser.Constructs.Text
			}

			var start, notText, data, text State

			atBreak := func(code Code) bool {
				if code == CodeEof {
					return true
				}

				for _, item := range constructs.ByCode[code] {
					if item.Previous == nil || item.Previous(self, self.Previous) {
						return true
					}
				}

				return false
			}

			start = func(code Code) State {
				if atBreak(code) {
					return text(code)
				}
				return notText(code)
			}

			notText = func(code Code) State {
				if code == CodeEof {
					effects.Consume(code)
					return nil
				}

				effects.Enter(TypeData, nil)
				effects.Consume(code)
				return data
			}

			data = func(code Code) State {
				if atBreak(code) {
					effects.Exit(TypeData)
					return text(code)
				}

				// Data.
				effects.Consume(code)
				return data
			}

			text = effects.Attempt(constructs, start, notText)

			return start
		},
	}
}

// createTextResolver merges adjacent data events, then runs the extra resolver.
func createTextResolver(extraResolver func(events []Event, context *TokenizeContext) []Event) func(events []Event, context *TokenizeContext) []Event {
	return func(events []Event, context *TokenizeContext) []Event {
		index := -1
		enter := -1

		// A rather boring computation (to merge adjacent `data` events) which improves mm performance by
		// 29%.
		for {
			index++
			if index > len(events) {
				break
			}
			if enter == -1 {
				if index < len(events) && events[index].Token.Type == TypeData {
					enter = index
					index++
				}
			} else if index >= len(events) || events[index].Token.Type != TypeData {
				// Don’t do anything if there is one data token.
				if index != enter+2 {
					events[enter].Token.End = events[index-1].Token.End
					events = splice(events, enter+2, index-enter-2, nil)
					index = enter + 2
				}

				enter = -1
			}
		}

		if extraResolver != nil {
			return extraResolver(events, context)
		}
		return events
	}
}

// resolveAllLineSuffixes splits trailing whitespace off data before a line ending or the end, as a line
// suffix or a hard break. Upstream calls it "a rather ugly set of instructions which again looks at
// chunks in the input stream".
func resolveAllLineSuffixes(events []Event, context *TokenizeContext) []Event {
	eventIndex := 0 // Skip first.

	for {
		eventIndex++
		if eventIndex > len(events) {
			break
		}
		if (eventIndex == len(events) || events[eventIndex].Token.Type == TypeLineEnding) &&
			events[eventIndex-1].Token.Type == TypeData {
			data := events[eventIndex-1].Token
			chunks := context.SliceStream(data)
			index := len(chunks)
			bufferIndex := -1
			size := 0
			tabs := false

			// Upstream's `while (index--)`: the test decrements too, so running out leaves index at -1.
			for {
				if index == 0 {
					index = -1
					break
				}
				index--
				chunk := chunks[index]

				if chunk.IsText {
					bufferIndex = len(chunk.Text)

					for bufferIndex > 0 && chunk.Text[bufferIndex-1] == ' ' {
						size++
						bufferIndex--
					}

					if bufferIndex != 0 {
						break
					}
					bufferIndex = -1
				} else if chunk.Code == CodeHorizontalTab {
					// Number
					tabs = true
					size++
				} else if chunk.Code == CodeVirtualSpace {
					// Empty
				} else {
					// Replacement character, exit.
					index++
					break
				}
			}
			// Allow final trailing whitespace.
			if context.contentTypeTextTrailing && eventIndex == len(events) {
				size = 0
			}

			if size != 0 {
				tokenType := TypeHardBreakTrailing
				if eventIndex == len(events) || tabs || size < hardBreakPrefixSizeMin {
					tokenType = TypeLineSuffix
				}
				startBufferIndex := bufferIndex
				if index == 0 {
					startBufferIndex = data.Start.bufferIndex + bufferIndex
				}
				token := &Token{
					Type: tokenType,
					Start: Point{
						bufferIndex: startBufferIndex,
						index:       data.Start.index + index,
						Line:        data.End.Line,
						Column:      data.End.Column - size,
						Offset:      data.End.Offset - size,
					},
					End: data.End,
				}

				data.End = token.Start

				if data.Start.Offset == data.End.Offset {
					data.Type = token.Type
					data.Start = token.Start
					data.End = token.End
				} else {
					events = splice(events, eventIndex, 0, []Event{
						{Enter: true, Token: token, Context: context},
						{Enter: false, Token: token, Context: context},
					})
					eventIndex += 2
				}
			}

			eventIndex++
		}
	}

	return events
}
