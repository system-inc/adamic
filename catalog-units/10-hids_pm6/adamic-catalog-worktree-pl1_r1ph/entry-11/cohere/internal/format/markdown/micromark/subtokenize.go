package micromark

import "slices"

// subtokenize is micromark-util-subtokenize: tokenize the content of every chunk that has a content type,
// and splice the result in. Returns the events and whether no subtokens were found.
//
// Upstream wraps the events in a SpliceBuffer, a gap buffer that makes repeated splices cheap. It has
// array semantics, so a slice is the port.
func subtokenize(events []Event) ([]Event, bool) {
	jumps := map[int]int{}
	index := -1
	more := false

	for index+1 < len(events) {
		index++
		for {
			jump, present := jumps[index]
			if !present {
				break
			}
			index = jump
		}

		event := events[index]

		// Add a hook for the GFM tasklist extension, which needs to know if text is in the first content
		// of a list item.
		if index > 0 &&
			event.Token.Type == TypeChunkFlow &&
			events[index-1].Token.Type == TypeListItemPrefix {
			subevents := event.Token.tokenizer.Events
			otherIndex := 0

			if otherIndex < len(subevents) && subevents[otherIndex].Token.Type == TypeLineEndingBlank {
				otherIndex += 2
			}

			if otherIndex < len(subevents) && subevents[otherIndex].Token.Type == TypeContent {
				for {
					otherIndex++
					if otherIndex >= len(subevents) {
						break
					}
					if subevents[otherIndex].Token.Type == TypeContent {
						break
					}

					if subevents[otherIndex].Token.Type == TypeChunkText {
						subevents[otherIndex].Token.isInFirstContentOfListItem = true
						otherIndex++
					}
				}
			}
		}

		if event.Enter {
			// Enter.
			if event.Token.ContentType != "" {
				var gaps map[int]int
				events, gaps = subcontent(events, index)
				for key, value := range gaps {
					jumps[key] = value
				}
				index = jumps[index]
				more = true
			}
		} else if event.Token.container {
			// Exit.
			otherIndex := index
			lineIndex := -1

			for otherIndex > 0 {
				otherIndex--
				otherEvent := events[otherIndex]

				if otherEvent.Token.Type == TypeLineEnding || otherEvent.Token.Type == TypeLineEndingBlank {
					if otherEvent.Enter {
						if lineIndex > 0 {
							events[lineIndex].Token.Type = TypeLineEndingBlank
						}

						otherEvent.Token.Type = TypeLineEnding
						lineIndex = otherIndex
					}
				} else if otherEvent.Token.Type == TypeLinePrefix || otherEvent.Token.Type == TypeListItemIndent {
					// Move past.
				} else {
					break
				}
			}

			// Upstream tests `lineIndex` for truthiness, so index 0 counts as none.
			if lineIndex > 0 {
				// Fix position.
				event.Token.End = events[lineIndex].Token.Start

				// Switch container exit w/ line endings.
				parameters := append([]Event{event}, slices.Clone(events[lineIndex:index])...)
				events = splice(events, lineIndex, index-lineIndex+1, parameters)
			}
		}
	}

	return events, !more
}

// subcontent tokenizes embedded tokens, splices them in, and returns the gaps to jump over.
func subcontent(events []Event, eventIndex int) ([]Event, map[int]int) {
	token := events[eventIndex].Token
	context := events[eventIndex].Context
	startPosition := eventIndex - 1
	var startPositions []int

	tokenizer := token.tokenizer

	if tokenizer == nil {
		start := token.Start
		tokenizer = context.Parser.create(token.ContentType, &start)

		if token.contentTypeTextTrailing {
			tokenizer.contentTypeTextTrailing = true
		}
	}

	childEvents := tokenizer.Events
	var jumps [][2]int
	gaps := map[int]int{}
	var previous *Token
	current := token
	adjust := 0
	start := 0
	breaks := []int{start}

	// Loop forward through the linked tokens to pass them in order to the subtokenizer.
	for current != nil {
		// Find the position of the event for this token.
		for {
			startPosition++
			if events[startPosition].Token == current {
				break
			}
		}

		startPositions = append(startPositions, startPosition)

		if current.tokenizer == nil {
			stream := context.SliceStream(current)

			if current.Next == nil {
				stream = append(stream, codeChunk(CodeEof))
			}

			if previous != nil {
				tokenizer.DefineSkip(current.Start)
			}

			if current.isInFirstContentOfListItem {
				tokenizer.gfmTasklistFirstContentOfListItem = true
			}

			tokenizer.Write(stream)

			if current.isInFirstContentOfListItem {
				tokenizer.gfmTasklistFirstContentOfListItem = false
			}
		}

		// Unravel the next token.
		previous = current
		current = current.Next
	}

	// The tokenizer may have replaced its events while resolving.
	childEvents = tokenizer.Events

	// Now, loop back through all events (and linked tokens), to figure out which parts belong where.
	current = token

	for index := 0; index < len(childEvents); index++ {
		if
		// Find a void token that includes a break.
		!childEvents[index].Enter &&
			childEvents[index-1].Enter &&
			childEvents[index].Token.Type == childEvents[index-1].Token.Type &&
			childEvents[index].Token.Start.Line != childEvents[index].Token.End.Line {
			start = index + 1
			breaks = append(breaks, start)
			// Help GC.
			current.tokenizer = nil
			current.Previous = nil
			current = current.Next
		}
	}

	// Help GC.
	tokenizer.Events = nil

	// If there’s one more token (which is the cases for lines that end in an EOF), that’s perfect: the
	// last point we found starts it. If there isn’t then make sure any remaining content is added to it.
	if current != nil {
		// Help GC.
		current.tokenizer = nil
		current.Previous = nil
	} else {
		breaks = breaks[:len(breaks)-1]
	}

	// Now splice the events from the subtokenizer into the current events, moving back to front so that
	// splice indices aren’t affected.
	for index := len(breaks) - 1; index >= 0; index-- {
		var slice []Event
		if index+1 < len(breaks) {
			slice = slices.Clone(childEvents[breaks[index]:breaks[index+1]])
		} else {
			slice = slices.Clone(childEvents[breaks[index]:])
		}
		start := startPositions[len(startPositions)-1]
		startPositions = startPositions[:len(startPositions)-1]
		jumps = append(jumps, [2]int{start, start + len(slice) - 1})
		events = splice(events, start, 2, slice)
	}

	slices.Reverse(jumps)

	for _, jump := range jumps {
		gaps[adjust+jump[0]] = adjust + jump[1]
		adjust += jump[1] - jump[0] - 1
	}

	return events, gaps
}
