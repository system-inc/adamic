// Package micromark is micromark 4, the markdown tokenizer Prettier's markdown parser runs, ported to Go.
//
// The fork parses `.md` with mdast-util-from-markdown over micromark, with the GFM, math, wiki-link and
// liquid extensions and an html-text override (src/language-markdown/parse/parse-markdown.js). The printer
// slices the original text by the offsets this tokenizer records, so a CommonMark library that builds a
// similar tree is not a substitute: the tree has to be this tree, position for position.
//
// # A port, not a reimplementation
//
// Each file follows one upstream file, state for state, and names its source. The state machines are
// closures in JavaScript and closures here. Where JavaScript semantics leak into behavior, the Go copies
// the semantics rather than the intent:
//
//   - Text is UTF-16 code units, the way JavaScript strings are. Offsets and columns count code units, the
//     printer slices by them, and the character classes test single code units, so an astral symbol
//     beside an asterisk is two non-punctuation units here exactly as it is upstream.
//   - The end of input is JavaScript's null. Upstream compares it with < and >, where null is 0, and a
//     real NUL never reaches a tokenizer because preprocessing replaces it, so CodeEof is 0.
//   - The interrupt factory hands constructs an object whose prototype is the tokenizer context. Self
//     reproduces that binding: reads are live, and a write lands on the view when there is one.
package micromark

// Code is a character code: a UTF-16 code unit, or one of the negative virtual codes, or CodeEof.
type Code int

// State is one state of a state machine. A nil State is upstream's undefined: the machine is done.
type State func(code Code) State

// Point is a place in the source. Line and column are 1-based, offset is 0-based, all in UTF-16 code
// units. index and bufferIndex locate the point in the chunk list, as upstream's _index and _bufferIndex.
type Point struct {
	Line        int
	Column      int
	Offset      int
	index       int
	bufferIndex int
}

// Chunk is a string of code units, or a single code. Upstream's Chunk is `string | number`.
type Chunk struct {
	Text   []uint16
	Code   Code
	IsText bool
}

func textChunk(text []uint16) Chunk { return Chunk{Text: text, IsText: true} }
func codeChunk(code Code) Chunk     { return Chunk{Code: code} }

// Token is a span of the source with a type, plus the private fields constructs hang on it.
//
// One struct carries every field any construct uses, because upstream's tokens are open objects and the
// fields are read across files. The ones prefixed with an underscore upstream are unexported here.
type Token struct {
	Type  string
	Start Point
	End   Point

	// ContentType marks a chunk whose contents another tokenizer parses: flow, content, string or text.
	ContentType string
	Previous    *Token
	Next        *Token

	tokenizer                    *TokenizeContext
	container                    bool
	isInFirstContentOfListItem   bool
	contentTypeTextTrailing      bool
	open                         bool
	close                        bool
	inactive                     bool
	balanced                     bool
	gfmAutolinkLiteralWalkedInto bool

	// Spread is upstream's _spread, read and written by mdast-util-from-markdown's prepareList.
	Spread bool
	// Align is upstream's _align on a GFM table token: "left", "right", "center" or "none" per column.
	Align []string
}

// ended reports whether the token has been exited. Upstream tests `token.end` for undefined.
func (token *Token) ended() bool { return token.End.Line != 0 }

// Event is an enter or an exit of a token, in the context that produced it.
type Event struct {
	Enter   bool
	Token   *Token
	Context *TokenizeContext
}

// Resolver rewrites events. It is a pointer so that resolveAll can call each one once, the way upstream
// compares functions by identity: labelStartLink, labelStartImage and labelEnd share one resolveAll.
type Resolver struct {
	Resolve func(events []Event, context *TokenizeContext) []Event
}

// Construct is a piece of syntax.
type Construct struct {
	Name         string
	Tokenize     func(self *Self, effects *Effects, ok State, nok State) State
	Continuation *Construct
	Exit         func(self *Self, effects *Effects)
	Previous     func(self *Self, code Code) bool
	Resolve      *Resolver
	ResolveTo    *Resolver
	ResolveAll   *Resolver
	Partial      bool
	Concrete     bool
	// Add is "after" for a construct that goes after the existing ones when extensions combine.
	Add string
}

// InitialConstruct starts a content type. Only its tokenize and resolveAll are used.
type InitialConstruct struct {
	Tokenize   func(self *Self, effects *Effects) State
	ResolveAll *Resolver
}

// ConstructRecord maps a code to the constructs that may start there. Null holds upstream's `null` key,
// the constructs tried at every code.
type ConstructRecord struct {
	ByCode map[Code][]*Construct
	Null   []*Construct

	// withNull is each code's list followed by Null, made once when the record is combined, so an attempt
	// at a code builds no list (#vbjv3d6). Nil for a record made by hand.
	withNull map[Code][]*Construct
}

// constructsAt is upstream's handleMapOfConstructs list: the constructs for the code, then those for every
// code, and none at the end of input. Callers only read it.
func (record *ConstructRecord) constructsAt(code Code) []*Construct {
	if code == CodeEof {
		return nil
	}
	if record.withNull != nil {
		if list, found := record.withNull[code]; found {
			return list
		}
		return record.Null
	}
	var list []*Construct
	list = append(list, record.ByCode[code]...)
	return append(list, record.Null...)
}

// combine fills withNull from ByCode and Null.
func (record *ConstructRecord) combine() {
	record.withNull = make(map[Code][]*Construct, len(record.ByCode))
	for code, list := range record.ByCode {
		record.withNull[code] = append(append([]*Construct(nil), list...), record.Null...)
	}
}

// Constructs is what attempt, check and interrupt accept: one construct, a list, or a record.
type Constructs interface{ isConstructs() }

// ConstructList is a list of constructs tried in order.
type ConstructList []*Construct

func (*Construct) isConstructs()       {}
func (ConstructList) isConstructs()    {}
func (*ConstructRecord) isConstructs() {}

// ContainerState is the state a container construct keeps across lines.
//
// Upstream's is an open record; these are every field a construct in our extension set writes.
type ContainerState struct {
	closeFlow         bool
	marker            Code
	size              int
	listType          string
	initialBlankLine  bool
	furtherBlankLines bool
	open              bool
}

// TokenizeContext is a tokenizer: its events, its position, and the parser it belongs to.
type TokenizeContext struct {
	Previous         Code
	ContainerState   *ContainerState
	Events           []Event
	Parser           *ParseContext
	CurrentConstruct *Construct
	Interrupt        bool

	contentTypeTextTrailing           bool
	gfmTasklistFirstContentOfListItem bool
	gfmTableDynamicInterruptHack      bool

	point                Point
	columnStart          map[int]int
	resolveAllConstructs []resolvable
	chunks               []Chunk
	stack                []*Token
	consumed             bool
	state                State
	initialize           *InitialConstruct
	effects              *Effects
	expectedCode         Code

	// initialContainerState and initialSelf are the container state and the `this` the tokenizer starts
	// with, held here so that starting one allocates neither (#vbjv3d6).
	initialContainerState ContainerState
	initialSelf           Self
}

// resolvable is upstream's resolveAllConstructs entry: a construct, or the initial construct, which also
// carries a resolveAll. Compared by identity.
type resolvable struct {
	construct  *Construct
	initialize *InitialConstruct
}

func (entry resolvable) resolveAll() *Resolver {
	if entry.construct != nil {
		return entry.construct.ResolveAll
	}
	return entry.initialize.ResolveAll
}

// Self is what a construct's tokenize receives as `this`: the context, or, under interrupt, an object
// whose prototype is the context with `interrupt: true` set on it.
//
// Reads go through to the live context. interrupt is the one field upstream ever writes through `this`,
// so it is the one field a view can shadow.
type Self struct {
	*TokenizeContext
	isView        bool
	viewInterrupt bool
}

// IsInterrupt reads `this.interrupt`.
func (self *Self) IsInterrupt() bool {
	if self.isView {
		return self.viewInterrupt
	}
	return self.TokenizeContext.Interrupt
}

// SetInterrupt writes `this.interrupt`, onto the view when there is one, as a property assignment would.
func (self *Self) SetInterrupt(value bool) {
	if self.isView {
		self.viewInterrupt = value
		return
	}
	self.TokenizeContext.Interrupt = value
}

// Effects are the tools a construct tokenizes with.
type Effects struct {
	Attempt   func(constructs Constructs, returnState State, bogusState State) State
	Check     func(constructs Constructs, returnState State, bogusState State) State
	Interrupt func(constructs Constructs, returnState State, bogusState State) State
	Consume   func(code Code)
	// Enter opens a token. A non-nil token is upstream's `fields`: it becomes the token, with its type
	// and start set.
	Enter func(tokenType string, token *Token) *Token
	Exit  func(tokenType string) *Token

	// tokenizer is the tokenizer these effects belong to, through which a factory reaches the parse's
	// Memory (#vbjv3d6).
	tokenizer *TokenizeContext
}

// memory is the Memory of the parse these effects tokenize, nil when it allocates each value.
func (effects *Effects) memory() *Memory { return effects.tokenizer.Parser.memory }

// Extension is a syntax extension: per hook, constructs by code.
type Extension struct {
	Document       *ConstructRecord
	ContentInitial *ConstructRecord
	FlowInitial    *ConstructRecord
	Flow           *ConstructRecord
	String         *ConstructRecord
	Text           *ConstructRecord
	InsideSpan     []*Construct
	// AttentionMarkers are codes, upstream's attentionMarkers.null.
	AttentionMarkers []Code
	// Disable are construct names, upstream's disable.null.
	Disable []string
}

// FullConstructs is every hook after combining the defaults with the extensions. Combining yields an
// extension, because upstream's gfm() is itself combineExtensions over five, then combined again, and the
// order that two-level combination produces is observable.
type FullConstructs = Extension

// ParseContext is the parser: the combined constructs and the tokenizer factories.
type ParseContext struct {
	Constructs *FullConstructs
	Defined    []string
	Lazy       map[int]bool

	// GfmFootnotes is upstream's parser.gfmFootnotes, which micromark-extension-gfm-footnote hangs on the
	// parser: the identifiers of the footnote definitions seen so far.
	GfmFootnotes []string

	// memory is where the parse's tokens, attempts and tokenizers come from, nil to allocate each (#93dpede,
	// #vbjv3d6).
	memory *Memory
}

// Content types, upstream's constants.contentType*.
const (
	ContentTypeContent  = "content"
	ContentTypeDocument = "document"
	ContentTypeFlow     = "flow"
	ContentTypeString   = "string"
	ContentTypeText     = "text"
)

// The constants upstream's micromark-util-symbol names, that constructs read.
const (
	attentionSideAfter                   = 2
	attentionSideBefore                  = 1
	atxHeadingOpeningFenceSizeMax        = 6
	autolinkDomainSizeMax                = 63
	autolinkSchemeSizeMax                = 32
	cdataOpeningString                   = "CDATA["
	characterGroupPunctuation            = 2
	characterGroupWhitespace             = 1
	characterReferenceDecimalSizeMax     = 7
	characterReferenceHexadecimalSizeMax = 6
	characterReferenceNamedSizeMax       = 31
	codeFencedSequenceSizeMin            = 3
	hardBreakPrefixSizeMin               = 2
	htmlBasic                            = 6
	htmlCdata                            = 5
	htmlComment                          = 2
	htmlComplete                         = 7
	htmlDeclaration                      = 4
	htmlInstruction                      = 3
	htmlRawSizeMax                       = 8
	htmlRaw                              = 1
	linkResourceDestinationBalanceMax    = 32
	linkReferenceSizeMax                 = 999
	listItemValueSizeMax                 = 10
	tabSize                              = 4
	thematicBreakMarkerCountMin          = 3
)
