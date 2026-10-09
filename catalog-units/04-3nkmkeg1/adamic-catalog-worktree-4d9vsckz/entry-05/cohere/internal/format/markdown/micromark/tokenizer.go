package micromark

import (
	"slices"
	"unicode/utf16"
)

// createTokenizer is upstream's createTokenizer, micromark/lib/create-tokenizer.js.
//
// from sets the point before the first character; later lines are indented with defineSkip.
func createTokenizer(parser *ParseContext, initialize *InitialConstruct, from *Point) *TokenizeContext {
	context := parser.memory.tokenizer()
	context.Previous = CodeEof
	context.ContainerState = &context.initialContainerState
	context.Parser = parser
	context.consumed = true
	context.initialize = initialize
	context.point = Point{bufferIndex: -1, index: 0, Line: 1, Column: 1, Offset: 0}
	if from != nil {
		if from.Line != 0 {
			context.point.Line = from.Line
		}
		if from.Column != 0 {
			context.point.Column = from.Column
		}
		context.point.Offset = from.Offset
	}

	context.initialSelf = Self{TokenizeContext: context}
	context.state = initialize.Tokenize(&context.initialSelf, context.effects)

	if initialize.ResolveAll != nil {
		context.resolveAllConstructs = append(context.resolveAllConstructs, resolvable{initialize: initialize})
	}

	return context
}

// newTokenizeContext is a tokenizer with only its effects and its column map, which a Memory keeps
// across parses (#vbjv3d6).
func newTokenizeContext() *TokenizeContext {
	context := &TokenizeContext{columnStart: map[int]int{}}
	context.effects = &Effects{
		Attempt:   context.constructFactory(context.onSuccessfulConstruct, false),
		Check:     context.constructFactory(context.onSuccessfulCheck, false),
		Consume:   context.consume,
		Enter:     context.enter,
		Exit:      context.exit,
		Interrupt: context.constructFactory(context.onSuccessfulCheck, true),
		tokenizer: context,
	}
	return context
}

// release zeroes the tokenizer for its next parse, keeping its effects, its column map, cleared, and its
// stack's array, which no other slice shares (restore copies into it).
func (context *TokenizeContext) release() {
	clear(context.columnStart)
	clear(context.stack[:cap(context.stack)])
	*context = TokenizeContext{columnStart: context.columnStart, effects: context.effects, stack: context.stack[:0]}
}

// Write feeds chunks to the tokenizer. Upstream's context.write: events come back once the last chunk is
// the end of input, and an empty list before that.
func (context *TokenizeContext) Write(slice []Chunk) []Event {
	context.chunks = pushChunks(context.chunks, slice)

	context.main()

	// Exit if we’re not done, resolve might change stuff.
	if last := context.chunks[len(context.chunks)-1]; last.IsText || last.Code != CodeEof {
		return nil
	}

	context.addResult(resolvable{initialize: context.initialize}, 0)

	// Otherwise, resolve, and exit.
	context.Events = resolveAllEntries(context.resolveAllConstructs, context.Events, context)

	return context.Events
}

// SliceSerialize is the source text of a token.
func (context *TokenizeContext) SliceSerialize(token *Token, expandTabs bool) string {
	return serializeChunks(context.SliceStream(token), expandTabs)
}

// SliceStream is the chunks of a token.
func (context *TokenizeContext) SliceStream(token *Token) []Chunk {
	return sliceChunks(context.chunks, token)
}

// Now is the current point.
func (context *TokenizeContext) Now() Point {
	return context.point
}

// DefineSkip records that lines starting at value's line start at value's column.
func (context *TokenizeContext) DefineSkip(value Point) {
	context.columnStart[value.Line] = value.Column
	context.accountForPotentialSkip()
}

// main is the main loop. `index` and `bufferIndex` in point are modified by consume.
func (context *TokenizeContext) main() {
	for context.point.index < len(context.chunks) {
		chunk := context.chunks[context.point.index]

		// If we’re in a buffer chunk, loop through it.
		if chunk.IsText {
			chunkIndex := context.point.index

			if context.point.bufferIndex < 0 {
				context.point.bufferIndex = 0
			}

			for context.point.index == chunkIndex && context.point.bufferIndex < len(chunk.Text) {
				context.goCode(Code(chunk.Text[context.point.bufferIndex]))
			}
		} else {
			context.goCode(chunk.Code)
		}
	}
}

// goCode deals with one code. Upstream's go.
//
// Upstream's development build asserts that the previous code was consumed. The production build, which
// the oracle runs, does not, and the fork's liquid construct depends on that: its mayClose returns a
// state without consuming, so the same code is fed again to the state it returned. So neither consume
// check is ported; the remaining panics guard invariants whose failure is a porting bug.
func (context *TokenizeContext) goCode(code Code) {
	context.consumed = false
	context.expectedCode = code
	if context.state == nil {
		panic("micromark: expected state")
	}
	context.state = context.state(code)
}

func (context *TokenizeContext) consume(code Code) {
	if code != context.expectedCode {
		panic("micromark: expected given code to equal expected code")
	}

	if markdownLineEnding(code) {
		context.point.Line++
		context.point.Column = 1
		if code == CodeCarriageReturnLineFeed {
			context.point.Offset += 2
		} else {
			context.point.Offset++
		}
		context.accountForPotentialSkip()
	} else if code != CodeVirtualSpace {
		context.point.Column++
		context.point.Offset++
	}

	// Not in a string chunk.
	if context.point.bufferIndex < 0 {
		context.point.index++
	} else {
		context.point.bufferIndex++

		// At end of string chunk.
		if context.point.bufferIndex == len(context.chunks[context.point.index].Text) {
			context.point.bufferIndex = -1
			context.point.index++
		}
	}

	// Expose the previous character.
	context.Previous = code

	// Mark as consumed.
	context.consumed = true
}

func (context *TokenizeContext) enter(tokenType string, token *Token) *Token {
	if token == nil {
		token = context.Parser.memory.tokens().New(Token{})
	}
	token.Type = tokenType
	token.Start = context.Now()

	context.Events = append(context.Events, Event{Enter: true, Token: token, Context: context})

	context.stack = append(context.stack, token)

	return token
}

func (context *TokenizeContext) exit(tokenType string) *Token {
	if len(context.stack) == 0 {
		panic("micromark: cannot close w/o open tokens")
	}
	token := context.stack[len(context.stack)-1]
	context.stack = context.stack[:len(context.stack)-1]
	token.End = context.Now()

	if tokenType != token.Type {
		panic("micromark: expected exit token to match current token, got " + tokenType + " for " + token.Type)
	}

	context.Events = append(context.Events, Event{Enter: false, Token: token, Context: context})

	return token
}

func (context *TokenizeContext) onSuccessfulConstruct(construct *Construct, info *storeInfo) {
	context.addResult(resolvable{construct: construct}, info.from)
}

func (context *TokenizeContext) onSuccessfulCheck(_ *Construct, info *storeInfo) {
	info.restore()
}

// constructFactory is upstream's constructFactory: attempt, check and interrupt differ in what happens on
// success and in whether constructs see `interrupt: true`.
func (context *TokenizeContext) constructFactory(onReturn func(*Construct, *storeInfo), interrupt bool) func(Constructs, State, State) State {
	return func(constructs Constructs, returnState State, bogusState State) State {
		run := context.Parser.memory.attempt()
		run.context, run.onReturn, run.interrupt = context, onReturn, interrupt
		run.returnState, run.bogusState = returnState, bogusState

		switch typed := constructs.(type) {
		case ConstructList:
			return run.handleListOfConstructs(typed)
		case *Construct:
			run.single[0] = typed
			return run.handleListOfConstructs(run.single[:])
		case *ConstructRecord:
			return run.handleMapOfConstructs(typed)
		default:
			panic("micromark: unknown constructs")
		}
	}
}

// attempt is one call of an attempt, check or interrupt. Upstream keeps its state in a closure's
// variables; here it is one struct, and its states are bound once per struct (newAttempt), so trying a
// construct allocates no closure of its own (#93dpede). A Memory reuses the struct, bound states and
// all, in later parses (#vbjv3d6).
type attempt struct {
	context                 *TokenizeContext
	onReturn                func(*Construct, *storeInfo)
	interrupt               bool
	returnState, bogusState State

	listOfConstructs []*Construct
	constructIndex   int
	currentConstruct *Construct
	info             storeInfo

	// pending is the construct start begins: the one handleConstruct named, which start reads before any
	// other construct of this attempt can be named.
	pending *Construct
	// single holds a lone construct as a list, without allocating one.
	single [1]*Construct
	// record is the record startRecord reads, the one handleMapOfConstructs was given.
	record *ConstructRecord
	// self is what the current construct tokenizes with. A construct's states are done with it once
	// they return ok or nok, before the next construct of this attempt starts and sets it again.
	self Self

	ok, nok, start, startMap State
}

// newAttempt is an attempt with its states bound and every other field zero.
func newAttempt() *attempt {
	run := &attempt{}
	run.ok, run.nok, run.start, run.startMap = run.succeed, run.fail, run.startPending, run.startRecord
	return run
}

// release zeroes the attempt for its next parse, keeping its bound states.
func (run *attempt) release() {
	*run = attempt{ok: run.ok, nok: run.nok, start: run.start, startMap: run.startMap}
}

func (run *attempt) handleListOfConstructs(list []*Construct) State {
	run.listOfConstructs = list
	run.constructIndex = 0

	if len(list) == 0 {
		if run.bogusState == nil {
			panic("micromark: expected `bogusState` to be given")
		}
		return run.bogusState
	}

	return run.handleConstruct(list[run.constructIndex])
}

func (run *attempt) handleMapOfConstructs(record *ConstructRecord) State {
	run.record = record
	return run.startMap
}

// startRecord is handleMapOfConstructs' start: the constructs for the code, then those for every code.
func (run *attempt) startRecord(code Code) State {
	return run.handleListOfConstructs(run.record.constructsAt(code))(code)
}

func (run *attempt) handleConstruct(construct *Construct) State {
	run.pending = construct
	return run.start
}

func (run *attempt) startPending(code Code) State {
	construct, context := run.pending, run.context
	run.info = context.store()
	run.currentConstruct = construct

	if !construct.Partial {
		context.CurrentConstruct = construct
	}

	if construct.Name != "" && slices.Contains(context.Parser.Constructs.Disable, construct.Name) {
		return run.nok(code)
	}

	run.self = Self{TokenizeContext: context, isView: run.interrupt, viewInterrupt: run.interrupt}
	return construct.Tokenize(&run.self, context.effects, run.ok, run.nok)(code)
}

func (run *attempt) succeed(code Code) State {
	if code != run.context.expectedCode {
		panic("micromark: expected code")
	}
	run.context.consumed = true
	run.onReturn(run.currentConstruct, &run.info)
	return run.returnState
}

func (run *attempt) fail(code Code) State {
	if code != run.context.expectedCode {
		panic("micromark: expected code")
	}
	run.context.consumed = true
	run.info.restore()

	run.constructIndex++
	if run.constructIndex < len(run.listOfConstructs) {
		return run.handleConstruct(run.listOfConstructs[run.constructIndex])
	}

	return run.bogusState
}

func (context *TokenizeContext) addResult(entry resolvable, from int) {
	construct := entry.construct
	resolveAll := entry.resolveAll()
	if resolveAll != nil && !slices.Contains(context.resolveAllConstructs, entry) {
		context.resolveAllConstructs = append(context.resolveAllConstructs, entry)
	}

	if construct != nil && construct.Resolve != nil {
		resolved := construct.Resolve.Resolve(context.Parser.memory.copyEvents(context.Events[from:]), context)
		context.Events = spliceEvents(context.Events, from, len(context.Events)-from, resolved)
	}

	if construct != nil && construct.ResolveTo != nil {
		context.Events = construct.ResolveTo.Resolve(context.Events, context)
	}
}

// storeInfo is upstream's Info: where events were when an attempt started, and how to go back. Upstream's
// restore is a closure over the saved state; here the state is the struct's own fields and restore a method,
// so taking one allocates only the copy of the stack it must keep.
type storeInfo struct {
	from int

	context          *TokenizeContext
	point            Point
	previous         Code
	currentConstruct *Construct
	stack            []*Token
}

func (context *TokenizeContext) store() storeInfo {
	return storeInfo{
		from:             len(context.Events),
		context:          context,
		point:            context.Now(),
		previous:         context.Previous,
		currentConstruct: context.CurrentConstruct,
		stack:            context.Parser.memory.copyStack(context.stack),
	}
}

// restore puts the tokenizer back where store found it.
func (info *storeInfo) restore() {
	context := info.context
	context.point = info.point
	context.Previous = info.previous
	context.CurrentConstruct = info.currentConstruct
	context.Events = context.Events[:info.from]
	// Copied into the stack's own array rather than taking the saved copy's: that copy has no room past
	// its length, so the next enter would reallocate it, after every attempt that fails (#vbjv3d6).
	context.stack = append(context.stack[:0], info.stack...)
	context.accountForPotentialSkip()
}

// accountForPotentialSkip moves the current point a bit forward in the line when it’s on a column skip.
func (context *TokenizeContext) accountForPotentialSkip() {
	if column, defined := context.columnStart[context.point.Line]; defined && context.point.Column < 2 {
		context.point.Column = column
		context.point.Offset += column - 1
	}
}

// sliceChunks gets the chunks from a slice of chunks in the range of a token.
func sliceChunks(chunks []Chunk, token *Token) []Chunk {
	startIndex := token.Start.index
	startBufferIndex := token.Start.bufferIndex
	endIndex := token.End.index
	endBufferIndex := token.End.bufferIndex
	var view []Chunk

	if startIndex == endIndex {
		view = []Chunk{textChunk(chunks[startIndex].Text[startBufferIndex:endBufferIndex])}
	} else {
		view = slices.Clone(chunks[startIndex:endIndex])

		if startBufferIndex > -1 {
			head := view[0]
			if head.IsText {
				view[0] = textChunk(head.Text[startBufferIndex:])
			} else {
				view = view[1:]
			}
		}

		if endBufferIndex > 0 {
			view = append(view, textChunk(chunks[endIndex].Text[:endBufferIndex]))
		}
	}

	return view
}

// serializeChunks gets the string value of a slice of chunks.
func serializeChunks(chunks []Chunk, expandTabs bool) string {
	var result []uint16
	atTab := false

	for _, chunk := range chunks {
		if chunk.IsText {
			result = append(result, chunk.Text...)
		} else {
			switch chunk.Code {
			case CodeCarriageReturn:
				result = append(result, '\r')
			case CodeLineFeed:
				result = append(result, '\n')
			case CodeCarriageReturnLineFeed:
				result = append(result, '\r', '\n')
			case CodeHorizontalTab:
				if expandTabs {
					result = append(result, ' ')
				} else {
					result = append(result, '\t')
				}
			case CodeVirtualSpace:
				if !expandTabs && atTab {
					continue
				}
				result = append(result, ' ')
			default:
				// Currently only replacement character.
				result = append(result, uint16(chunk.Code))
			}
		}

		atTab = !chunk.IsText && chunk.Code == CodeHorizontalTab
	}

	return string(utf16.Decode(result))
}
