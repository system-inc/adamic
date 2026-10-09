package micromark

// documentInitial is upstream's document, micromark/lib/initialize/document.js: containers (block
// quotes, list items) and the flow inside them.
var documentInitial *InitialConstruct

func init() { documentInitial = &InitialConstruct{Tokenize: initializeDocument} }

// containerConstruct tries a new container after optional indentation.
var containerConstruct = &Construct{Tokenize: tokenizeContainer}

// stackItem is a container construct and its state.
type stackItem struct {
	construct *Construct
	state     *ContainerState
}

func initializeDocument(self *Self, effects *Effects) State {
	var stack []stackItem
	continued := 0
	var childFlow *TokenizeContext
	var childToken *Token
	var lineStartOffset int

	var start, documentContinue, checkNewContainers, thereIsANewContainer, thereIsNoNewContainer,
		documentContinued, containerContinue, flowStart, flowContinue State
	var writeToChild func(token *Token, endOfFile bool)
	var exitContainers func(size int)
	var closeFlow func()

	start = func(code Code) State {
		// First we iterate through the open blocks, starting with the root document, and descending
		// through last children down to the last open block. Each block imposes a condition that the line
		// must satisfy if the block is to remain open. In this phase we may match all or just some of the
		// open blocks. But we cannot close unmatched blocks yet, because we may have a lazy continuation
		// line.
		if continued < len(stack) {
			item := stack[continued]
			self.ContainerState = item.state
			return effects.Attempt(item.construct.Continuation, documentContinue, checkNewContainers)(code)
		}

		// Done.
		return checkNewContainers(code)
	}

	documentContinue = func(code Code) State {
		continued++

		// Note: this field is called `_closeFlow` but it also closes containers.
		if self.ContainerState.closeFlow {
			self.ContainerState.closeFlow = false

			if childFlow != nil {
				closeFlow()
			}

			// Note: this algorithm for moving events around is similar to the algorithm when dealing with
			// lazy lines in `writeToChild`.
			indexBeforeExits := len(self.Events)
			indexBeforeFlow := indexBeforeExits
			var point Point

			// Find the flow chunk.
			for indexBeforeFlow > 0 {
				indexBeforeFlow--
				if !self.Events[indexBeforeFlow].Enter && self.Events[indexBeforeFlow].Token.Type == TypeChunkFlow {
					point = self.Events[indexBeforeFlow].Token.End
					break
				}
			}

			exitContainers(continued)

			// Fix positions.
			index := indexBeforeExits

			for index < len(self.Events) {
				self.Events[index].Token.End = point
				index++
			}

			// Inject the exits earlier (they’re still also at the end).
			exits := append([]Event(nil), self.Events[indexBeforeExits:]...)
			self.Events = splice(self.Events, indexBeforeFlow+1, 0, exits)

			// Discard the duplicate exits.
			self.Events = self.Events[:index]

			return checkNewContainers(code)
		}

		return start(code)
	}

	checkNewContainers = func(code Code) State {
		// Next, after consuming the continuation markers for existing blocks, we look for new block starts
		// (e.g. `>` for a block quote). If we encounter a new block start, we close any blocks unmatched
		// in step 1 before creating the new block as a child of the last matched block.
		if continued == len(stack) {
			// No need to `check` whether there’s a container, of `exitContainers` would be moot. We can
			// instead immediately `attempt` to parse one.
			if childFlow == nil {
				return documentContinued(code)
			}

			// If we have concrete content, such as block HTML or fenced code, we can’t have containers
			// “pierce” into them, so we can immediately start.
			if childFlow.CurrentConstruct != nil && childFlow.CurrentConstruct.Concrete {
				return flowStart(code)
			}

			// If we do have flow, it could still be a blank line, but we’d be interrupting it w/ a new
			// container if there’s a current construct.
			self.SetInterrupt(childFlow.CurrentConstruct != nil && !childFlow.gfmTableDynamicInterruptHack)
		}

		// Check if there is a new container.
		self.ContainerState = &ContainerState{}
		return effects.Check(containerConstruct, thereIsANewContainer, thereIsNoNewContainer)(code)
	}

	thereIsANewContainer = func(code Code) State {
		if childFlow != nil {
			closeFlow()
		}
		exitContainers(continued)
		return documentContinued(code)
	}

	thereIsNoNewContainer = func(code Code) State {
		self.Parser.Lazy[self.Now().Line] = continued != len(stack)
		lineStartOffset = self.Now().Offset
		return flowStart(code)
	}

	documentContinued = func(code Code) State {
		// Try new containers.
		self.ContainerState = &ContainerState{}
		return effects.Attempt(containerConstruct, containerContinue, flowStart)(code)
	}

	containerContinue = func(code Code) State {
		continued++
		stack = append(stack, stackItem{construct: self.CurrentConstruct, state: self.ContainerState})
		// Try another.
		return documentContinued(code)
	}

	flowStart = func(code Code) State {
		if code == CodeEof {
			if childFlow != nil {
				closeFlow()
			}
			exitContainers(0)
			effects.Consume(code)
			return nil
		}

		if childFlow == nil {
			now := self.Now()
			childFlow = self.Parser.create(ContentTypeFlow, &now)
		}
		effects.Enter(TypeChunkFlow, &Token{
			tokenizer:   childFlow,
			ContentType: ContentTypeFlow,
			Previous:    childToken,
		})

		return flowContinue(code)
	}

	flowContinue = func(code Code) State {
		if code == CodeEof {
			writeToChild(effects.Exit(TypeChunkFlow), true)
			exitContainers(0)
			effects.Consume(code)
			return nil
		}

		if markdownLineEnding(code) {
			effects.Consume(code)
			writeToChild(effects.Exit(TypeChunkFlow), false)
			// Get ready for the next line.
			continued = 0
			self.SetInterrupt(false)
			return start
		}

		effects.Consume(code)
		return flowContinue
	}

	writeToChild = func(token *Token, endOfFile bool) {
		stream := self.SliceStream(token)
		if endOfFile {
			stream = append(stream, codeChunk(CodeEof))
		}
		token.Previous = childToken
		if childToken != nil {
			childToken.Next = token
		}
		childToken = token
		childFlow.DefineSkip(token.Start)
		childFlow.Write(stream)

		// Alright, so we just added a lazy line. If the lazy line started a new flow block, we exit the
		// current containers between the two flow blocks. See upstream for the three examples.
		if self.Parser.Lazy[token.Start.Line] {
			index := len(childFlow.Events)

			for index > 0 {
				index--
				if
				// The token starts before the line ending…
				childFlow.Events[index].Token.Start.Offset < lineStartOffset &&
					// …and either is not ended yet…
					(!childFlow.Events[index].Token.ended() ||
						// …or ends after it.
						childFlow.Events[index].Token.End.Offset > lineStartOffset) {
					// Exit: there’s still something open, which means it’s a lazy line part of something.
					return
				}
			}

			// Note: this algorithm for moving events around is similar to the algorithm when closing flow
			// in `documentContinue`.
			indexBeforeExits := len(self.Events)
			indexBeforeFlow := indexBeforeExits
			seen := false
			var point Point

			// Find the previous chunk (the one before the lazy line).
			for indexBeforeFlow > 0 {
				indexBeforeFlow--
				if !self.Events[indexBeforeFlow].Enter && self.Events[indexBeforeFlow].Token.Type == TypeChunkFlow {
					if seen {
						point = self.Events[indexBeforeFlow].Token.End
						break
					}

					seen = true
				}
			}

			exitContainers(continued)

			// Fix positions.
			index = indexBeforeExits

			for index < len(self.Events) {
				self.Events[index].Token.End = point
				index++
			}

			// Inject the exits earlier (they’re still also at the end).
			exits := append([]Event(nil), self.Events[indexBeforeExits:]...)
			self.Events = splice(self.Events, indexBeforeFlow+1, 0, exits)

			// Discard the duplicate exits.
			self.Events = self.Events[:index]
		}
	}

	exitContainers = func(size int) {
		index := len(stack)

		// Exit open containers.
		for index > size {
			index--
			entry := stack[index]
			self.ContainerState = entry.state
			entry.construct.Exit(self, effects)
		}

		stack = stack[:size]
	}

	closeFlow = func() {
		childFlow.Write([]Chunk{codeChunk(CodeEof)})
		childToken = nil
		childFlow = nil
		self.ContainerState.closeFlow = false
	}

	return start
}

func tokenizeContainer(self *Self, effects *Effects, ok State, nok State) State {
	limit := tabSize
	for _, name := range self.Parser.Constructs.Disable {
		if name == "codeIndented" {
			limit = 0
		}
	}
	return factorySpace(effects, effects.Attempt(self.Parser.Constructs.Document, ok, nok), TypeLinePrefix, limit)
}
