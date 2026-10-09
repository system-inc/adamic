package mdast

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/markdown/micromark"
)

// handle is upstream's Handle: what runs on entering or exiting a token type.
type handle func(context *compileContext, token *micromark.Token)

// config is upstream's Config, after the fork's mdast extensions are configured in.
type config struct {
	canContainEols []string
	enter          map[string]handle
	exit           map[string]handle
}

// compileData is upstream's CompileData: the flags handlers share across events.
type compileData struct {
	expectingFirstListItemValue  bool
	flowCodeInside               bool
	setextHeadingSlurpLineEnding bool
	atHardBreak                  bool
	inReference                  bool
	referenceType                string
	characterReferenceType       string
	inTable                      bool
	mathFlowInside               bool
}

type tokenStackEntry struct {
	token        *micromark.Token
	errorHandler func(context *compileContext, left *micromark.Token, right *micromark.Token)
}

// compileContext is upstream's CompileContext, with the event's sliceSerialize bound per handler call.
type compileContext struct {
	stack      []*Node
	tokenStack []tokenStackEntry
	config     *config
	data       compileData

	tokenizer *micromark.TokenizeContext
	offsets   []int

	// nodes is where the tree's nodes come from, nil to allocate each (#93dpede).
	nodes *arena.Arena[Node]
}

func (context *compileContext) sliceSerialize(token *micromark.Token) string {
	return context.tokenizer.SliceSerialize(token, false)
}

func (context *compileContext) top() *Node {
	return context.stack[len(context.stack)-1]
}

// point converts a micromark point, whose offset counts UTF-16 units, to a tree point in bytes.
func (context *compileContext) point(point micromark.Point) Point {
	return Point{Line: point.Line, Column: point.Column, Offset: context.offsets[point.Offset]}
}

// FromMarkdown parses markdown the way the fork's parseMarkdown does with the given source, and returns
// the tree. source is the text micromark reads (front matter already blanked by the caller); original is
// the text positions refer to, which has the same UTF-16 length.
func FromMarkdown(source string, original string, nodes *arena.Arena[Node]) (root *Node, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("markdown: %v", recovered)
		}
	}()
	memory := parseMemories.Get().(*parseMemory)
	defer func() {
		memory.parse.Reset()
		parseMemories.Put(memory)
	}()
	memory.units = micromark.AppendSourceUnits(memory.units[:0], source)
	units := memory.units
	memory.offsets = byteOffsets(original, memory.offsets[:0])
	if len(memory.offsets) != len(units)+1 {
		return nil, fmt.Errorf("markdown: the parsed text and the original differ in length (%d and %d UTF-16 units)", len(units), len(memory.offsets)-1)
	}
	events := micromark.Parse(units, micromark.MarkdownConstructs(), memory.parse)
	return compile(events, compileConfig(), memory.offsets, nodes), nil
}

// parseMemory is what one parse uses only while it runs, kept for the next (#93dpede): micromark's tokens
// and tokenizers, the text as UTF-16 units, and the map from UTF-16 index to byte offset. The tree compile
// builds keeps none of them: every point copies its offset out, and every value is copied out of the units
// as a string.
type parseMemory struct {
	parse   *micromark.Memory
	units   []uint16
	offsets []int
}

// parseMemories hold the memory of parses that have finished. A pool, because files are formatted on
// several goroutines at once and each Get is that caller's alone.
var parseMemories = sync.Pool{New: func() any {
	return &parseMemory{parse: micromark.NewMemory()}
}}

// byteOffsets maps every UTF-16 index of text, and its end, to a byte offset, appended to offsets. An index
// inside a surrogate pair maps to the start of its character.
func byteOffsets(text string, offsets []int) []int {
	offsets = slices.Grow(offsets, len(text)+1)
	for index, character := range text {
		offsets = append(offsets, index)
		if character >= 0x10000 && character != utf8.RuneError {
			offsets = append(offsets, index)
		}
	}
	return append(offsets, len(text))
}

// newConfig is upstream's compiler config with the fork's mdastExtensions configured in, in order:
// gfmFromMarkdown() (autolink literal without its transform, footnote, strikethrough, table, task list
// item), mathFromMarkdown(), the wiki-link fromMarkdown() and liquidFromMarkdown().
// compileConfig is newConfig, made once: every compile reads the same handlers and never changes them, and
// making them was a measurable share of a format's allocation (#93dpede).
var compileConfig = sync.OnceValue(newConfig)

func newConfig() *config {
	config := &config{
		canContainEols: []string{"emphasis", "fragment", "heading", "paragraph", "strong"},
		enter: map[string]handle{
			"autolink":                    opener(newLink, nil),
			"autolinkProtocol":            onEnterData,
			"autolinkEmail":               onEnterData,
			"atxHeading":                  opener(newHeading, nil),
			"blockQuote":                  opener(newBlockQuote, nil),
			"characterEscape":             onEnterData,
			"characterReference":          onEnterData,
			"codeFenced":                  opener(newCodeFlow, nil),
			"codeFencedFenceInfo":         buffer,
			"codeFencedFenceMeta":         buffer,
			"codeIndented":                opener(newCodeFlow, buffer),
			"codeText":                    opener(newCodeText, buffer),
			"codeTextData":                onEnterData,
			"data":                        onEnterData,
			"codeFlowValue":               onEnterData,
			"definition":                  opener(newDefinition, nil),
			"definitionDestinationString": buffer,
			"definitionLabelString":       buffer,
			"definitionTitleString":       buffer,
			"emphasis":                    opener(newParent("emphasis"), nil),
			"hardBreakEscape":             opener(newHardBreak, nil),
			"hardBreakTrailing":           opener(newHardBreak, nil),
			"htmlFlow":                    opener(newHTML, buffer),
			"htmlFlowData":                onEnterData,
			"htmlText":                    opener(newHTML, buffer),
			"htmlTextData":                onEnterData,
			"image":                       opener(newImage, nil),
			"label":                       buffer,
			"link":                        opener(newLink, nil),
			"listItem":                    opener(newListItem, nil),
			"listItemValue":               onEnterListItemValue,
			"listOrdered":                 opener(newList, onEnterListOrdered),
			"listUnordered":               opener(newList, nil),
			"paragraph":                   opener(newParent("paragraph"), nil),
			"reference":                   onEnterReference,
			"referenceString":             buffer,
			"resourceDestinationString":   buffer,
			"resourceTitleString":         buffer,
			"setextHeading":               opener(newHeading, nil),
			"strong":                      opener(newParent("strong"), nil),
			"thematicBreak":               opener(newThematicBreak, nil),
		},
		exit: map[string]handle{
			"atxHeading":                          closer(nil),
			"atxHeadingSequence":                  onExitAtxHeadingSequence,
			"autolink":                            closer(nil),
			"autolinkEmail":                       onExitAutolinkEmail,
			"autolinkProtocol":                    onExitAutolinkProtocol,
			"blockQuote":                          closer(nil),
			"characterEscapeValue":                onExitData,
			"characterReferenceMarkerHexadecimal": onExitCharacterReferenceMarker,
			"characterReferenceMarkerNumeric":     onExitCharacterReferenceMarker,
			"characterReferenceValue":             onExitCharacterReferenceValue,
			"characterReference":                  onExitCharacterReference,
			"codeFenced":                          closer(onExitCodeFenced),
			"codeFencedFence":                     onExitCodeFencedFence,
			"codeFencedFenceInfo":                 onExitCodeFencedFenceInfo,
			"codeFencedFenceMeta":                 onExitCodeFencedFenceMeta,
			"codeFlowValue":                       onExitData,
			"codeIndented":                        closer(onExitCodeIndented),
			"codeText":                            closer(onExitCodeText),
			"codeTextData":                        onExitData,
			"data":                                onExitData,
			"definition":                          closer(nil),
			"definitionDestinationString":         onExitDefinitionDestinationString,
			"definitionLabelString":               onExitDefinitionLabelString,
			"definitionTitleString":               onExitDefinitionTitleString,
			"emphasis":                            closer(nil),
			"hardBreakEscape":                     closer(onExitHardBreak),
			"hardBreakTrailing":                   closer(onExitHardBreak),
			"htmlFlow":                            closer(onExitHTMLFlow),
			"htmlFlowData":                        onExitData,
			"htmlText":                            closer(onExitHTMLText),
			"htmlTextData":                        onExitData,
			"image":                               closer(onExitImage),
			"label":                               onExitLabel,
			"labelText":                           onExitLabelText,
			"lineEnding":                          onExitLineEnding,
			"link":                                closer(onExitLink),
			"listItem":                            closer(nil),
			"listOrdered":                         closer(nil),
			"listUnordered":                       closer(nil),
			"paragraph":                           closer(nil),
			"referenceString":                     onExitReferenceString,
			"resourceDestinationString":           onExitResourceDestinationString,
			"resourceTitleString":                 onExitResourceTitleString,
			"resource":                            onExitResource,
			"setextHeading":                       closer(onExitSetextHeading),
			"setextHeadingLineSequence":           onExitSetextHeadingLineSequence,
			"setextHeadingText":                   onExitSetextHeadingText,
			"strong":                              closer(nil),
			"thematicBreak":                       closer(nil),
		},
	}

	// Each extension's handlers replace any earlier ones for the same type (Object.assign), and its
	// canContainEols append.
	for _, extension := range []struct {
		canContainEols []string
		enter          map[string]handle
		exit           map[string]handle
	}{
		gfmAutolinkLiteralFromMarkdown(),
		gfmFootnoteFromMarkdown(),
		gfmStrikethroughFromMarkdown(),
		gfmTableFromMarkdown(),
		gfmTaskListItemFromMarkdown(),
		mathFromMarkdown(),
		wikiLinkFromMarkdown(),
		liquidFromMarkdown(),
	} {
		config.canContainEols = append(config.canContainEols, extension.canContainEols...)
		for tokenType, handler := range extension.enter {
			config.enter[tokenType] = handler
		}
		for tokenType, handler := range extension.exit {
			config.exit[tokenType] = handler
		}
	}

	return config
}

// compile turns events into a tree, upstream's compile.
func compile(events []micromark.Event, config *config, offsets []int, nodes *arena.Arena[Node]) *Node {
	tree := nodes.New(Node{NodeType: "root", IsParent: true, Children: []*Node{}})
	context := &compileContext{stack: []*Node{tree}, config: config, offsets: offsets, nodes: nodes}
	var listStack []int

	for index := 0; index < len(events); index++ {
		// We preprocess lists to add `listItem` tokens, and to infer whether items the list itself are
		// spread out.
		if events[index].Token.Type == micromark.TypeListOrdered || events[index].Token.Type == micromark.TypeListUnordered {
			if events[index].Enter {
				listStack = append(listStack, index)
			} else {
				tail := listStack[len(listStack)-1]
				listStack = listStack[:len(listStack)-1]
				index = prepareList(&events, tail, index)
			}
		}
	}

	for _, event := range events {
		handlers := config.exit
		if event.Enter {
			handlers = config.enter
		}

		if handler, present := handlers[event.Token.Type]; present {
			context.tokenizer = event.Context
			handler(context, event.Token)
		}
	}

	// Handle tokens still being open.
	if len(context.tokenStack) > 0 {
		tail := context.tokenStack[len(context.tokenStack)-1]
		handler := tail.errorHandler
		if handler == nil {
			handler = defaultOnError
		}
		handler(context, nil, tail.token)
	}

	// Figure out `root` position.
	if len(events) > 0 {
		tree.Position = &Position{
			Start: context.point(events[0].Token.Start),
			End:   context.point(events[len(events)-2].Token.End),
		}
	} else {
		tree.Position = &Position{Start: Point{Line: 1, Column: 1}, End: Point{Line: 1, Column: 1}}
	}

	return tree
}

// prepareList adds listItem tokens and infers spread. Returns the (moved) index of the list's exit.
func prepareList(eventsPointer *[]micromark.Event, start int, length int) int {
	events := *eventsPointer
	index := start - 1
	containerBalance := -1
	listSpread := false
	var listItem *micromark.Token
	lineIndex := 0           // Upstream's undefined is 0 here: every test of it is for truthiness.
	firstBlankLineIndex := 0 // Same.
	atMarker := false

	for index+1 <= length {
		index++
		event := events[index]

		switch event.Token.Type {
		case micromark.TypeListUnordered, micromark.TypeListOrdered, micromark.TypeBlockQuote:
			if event.Enter {
				containerBalance++
			} else {
				containerBalance--
			}

			atMarker = false

		case micromark.TypeLineEndingBlank:
			if event.Enter {
				if listItem != nil && !atMarker && containerBalance == 0 && firstBlankLineIndex == 0 {
					firstBlankLineIndex = index
				}

				atMarker = false
			}

		case micromark.TypeLinePrefix, micromark.TypeListItemValue, micromark.TypeListItemMarker,
			micromark.TypeListItemPrefix, micromark.TypeListItemPrefixWhitespace:
			// Empty.

		default:
			atMarker = false
		}

		if (containerBalance == 0 && event.Enter && event.Token.Type == micromark.TypeListItemPrefix) ||
			(containerBalance == -1 && !event.Enter &&
				(event.Token.Type == micromark.TypeListUnordered || event.Token.Type == micromark.TypeListOrdered)) {
			if listItem != nil {
				tailIndex := index
				lineIndex = 0

				for tailIndex > 0 {
					tailIndex--
					tailEvent := events[tailIndex]

					if tailEvent.Token.Type == micromark.TypeLineEnding || tailEvent.Token.Type == micromark.TypeLineEndingBlank {
						if !tailEvent.Enter {
							continue
						}

						if lineIndex != 0 {
							events[lineIndex].Token.Type = micromark.TypeLineEndingBlank
							listSpread = true
						}

						tailEvent.Token.Type = micromark.TypeLineEnding
						lineIndex = tailIndex
					} else if tailEvent.Token.Type == micromark.TypeLinePrefix ||
						tailEvent.Token.Type == micromark.TypeBlockQuotePrefix ||
						tailEvent.Token.Type == micromark.TypeBlockQuotePrefixWhitespace ||
						tailEvent.Token.Type == micromark.TypeBlockQuoteMarker ||
						tailEvent.Token.Type == micromark.TypeListItemIndent {
						// Empty
					} else {
						break
					}
				}

				if firstBlankLineIndex != 0 && (lineIndex == 0 || firstBlankLineIndex < lineIndex) {
					listItem.Spread = true
				}

				// Fix position.
				if lineIndex != 0 {
					listItem.End = events[lineIndex].Token.Start
				} else {
					listItem.End = event.Token.End
				}

				at := index
				if lineIndex != 0 {
					at = lineIndex
				}
				events = slices.Insert(events, at, micromark.Event{Enter: false, Token: listItem, Context: event.Context})
				index++
				length++
			}

			// Create a new list item.
			if event.Token.Type == micromark.TypeListItemPrefix {
				item := &micromark.Token{Type: "listItem", Spread: false, Start: event.Token.Start}
				listItem = item
				events = slices.Insert(events, index, micromark.Event{Enter: true, Token: item, Context: event.Context})
				index++
				length++
				firstBlankLineIndex = 0
				atMarker = true
			}
		}
	}

	events[start].Token.Spread = listSpread
	*eventsPointer = events
	return length
}

// opener is upstream's opener: create a node and enter it, then run `and`.
func opener(create func(token *micromark.Token) Node, and handle) handle {
	return func(context *compileContext, token *micromark.Token) {
		context.enter(context.nodes.New(create(token)), token, nil)
		if and != nil {
			and(context, token)
		}
	}
}

// buffer pushes a fragment to collect text into.
func buffer(context *compileContext, _ *micromark.Token) {
	context.stack = append(context.stack, context.nodes.New(Node{NodeType: "fragment", IsParent: true, Children: []*Node{}}))
}

func (context *compileContext) enter(node *Node, token *micromark.Token, errorHandler func(*compileContext, *micromark.Token, *micromark.Token)) {
	parent := context.top()
	parent.Children = append(parent.Children, node)
	context.stack = append(context.stack, node)
	context.tokenStack = append(context.tokenStack, tokenStackEntry{token: token, errorHandler: errorHandler})
	node.Position = &Position{Start: context.point(token.Start)}
}

// closer is upstream's closer: run `and`, then exit.
func closer(and handle) handle {
	return func(context *compileContext, token *micromark.Token) {
		if and != nil {
			and(context, token)
		}
		context.exit(token, nil)
	}
}

func (context *compileContext) exit(token *micromark.Token, onExitError func(*compileContext, *micromark.Token, *micromark.Token)) {
	node := context.stack[len(context.stack)-1]
	context.stack = context.stack[:len(context.stack)-1]

	if len(context.tokenStack) == 0 {
		panic(fmt.Sprintf("Cannot close `%s`: it’s not open", token.Type))
	}
	open := context.tokenStack[len(context.tokenStack)-1]
	context.tokenStack = context.tokenStack[:len(context.tokenStack)-1]

	if open.token.Type != token.Type {
		if onExitError != nil {
			onExitError(context, token, open.token)
		} else {
			handler := open.errorHandler
			if handler == nil {
				handler = defaultOnError
			}
			handler(context, token, open.token)
		}
	}

	node.Position.End = context.point(token.End)
}

// resume pops the top node and returns its text.
func (context *compileContext) resume() string {
	node := context.stack[len(context.stack)-1]
	context.stack = context.stack[:len(context.stack)-1]
	return ToString(node)
}

func defaultOnError(_ *compileContext, left *micromark.Token, right *micromark.Token) {
	if left != nil {
		panic(fmt.Sprintf("Cannot close `%s`: a different token (`%s`) is open", left.Type, right.Type))
	}
	panic(fmt.Sprintf("Cannot close document, a token (`%s`) is still open", right.Type))
}

//
// Handlers.
//

func onEnterListOrdered(context *compileContext, _ *micromark.Token) {
	context.data.expectingFirstListItemValue = true
}

func onEnterListItemValue(context *compileContext, token *micromark.Token) {
	if context.data.expectingFirstListItemValue {
		ancestor := context.stack[len(context.stack)-2]
		start := parseIntPrefix(context.sliceSerialize(token))
		ancestor.Start = &start
		context.data.expectingFirstListItemValue = false
	}
}

// parseIntPrefix is Number.parseInt in base 10 for the digits micromark allows in a list item value.
func parseIntPrefix(text string) int {
	value, _ := strconv.Atoi(text)
	return value
}

func onExitCodeFencedFenceInfo(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Lang = &data
}

func onExitCodeFencedFenceMeta(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Meta = &data
}

func onExitCodeFencedFence(context *compileContext, token *micromark.Token) {
	// Exit if this is the closing fence.
	if context.data.flowCodeInside {
		return
	}
	buffer(context, token)
	context.data.flowCodeInside = true
}

func onExitCodeFenced(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Value = trimOneLineEnding(data, true, true)
	context.data.flowCodeInside = false
}

func onExitCodeIndented(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Value = trimOneLineEnding(data, false, true)
}

// trimOneLineEnding is `.replace(/^(\r?\n|\r)|(\r?\n|\r)$/g, ”)` with either side optional: one line
// ending at the start, and one at the end. Prettier normalized line endings, so only \n occurs.
func trimOneLineEnding(data string, start bool, end bool) string {
	if start {
		data = strings.TrimPrefix(data, "\n")
	}
	if end {
		data = strings.TrimSuffix(data, "\n")
	}
	return data
}

func onExitDefinitionLabelString(context *compileContext, token *micromark.Token) {
	label := context.resume()
	node := context.top()
	node.Label = &label
	node.Identifier = strings.ToLower(micromark.NormalizeIdentifier(context.sliceSerialize(token)))
}

func onExitDefinitionTitleString(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Title = &data
}

func onExitDefinitionDestinationString(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().URL = data
}

func onExitAtxHeadingSequence(context *compileContext, token *micromark.Token) {
	node := context.top()
	if node.Depth == 0 {
		node.Depth = len(context.sliceSerialize(token))
	}
}

func onExitSetextHeadingText(context *compileContext, _ *micromark.Token) {
	context.data.setextHeadingSlurpLineEnding = true
}

func onExitSetextHeadingLineSequence(context *compileContext, token *micromark.Token) {
	node := context.top()
	if strings.HasPrefix(context.sliceSerialize(token), "=") {
		node.Depth = 1
	} else {
		node.Depth = 2
	}
}

func onExitSetextHeading(context *compileContext, _ *micromark.Token) {
	context.data.setextHeadingSlurpLineEnding = false
}

func onEnterData(context *compileContext, token *micromark.Token) {
	node := context.top()
	var tail *Node
	if len(node.Children) > 0 {
		tail = node.Children[len(node.Children)-1]
	}

	if tail == nil || tail.NodeType != "text" {
		// Add a new text node.
		tail = context.nodes.New(Node{NodeType: "text", IsLiteral: true})
		tail.Position = &Position{Start: context.point(token.Start)}
		node.Children = append(node.Children, tail)
	}

	context.stack = append(context.stack, tail)
}

func onExitData(context *compileContext, token *micromark.Token) {
	tail := context.stack[len(context.stack)-1]
	context.stack = context.stack[:len(context.stack)-1]
	tail.Value += context.sliceSerialize(token)
	tail.Position.End = context.point(token.End)
}

func onExitLineEnding(context *compileContext, token *micromark.Token) {
	node := context.top()

	// If we’re at a hard break, include the line ending in there.
	if context.data.atHardBreak {
		tail := node.Children[len(node.Children)-1]
		tail.Position.End = context.point(token.End)
		context.data.atHardBreak = false
		return
	}

	if !context.data.setextHeadingSlurpLineEnding && slices.Contains(context.config.canContainEols, node.NodeType) {
		onEnterData(context, token)
		onExitData(context, token)
	}
}

func onExitHardBreak(context *compileContext, _ *micromark.Token) {
	context.data.atHardBreak = true
}

func onExitHTMLFlow(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Value = data
}

func onExitHTMLText(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Value = data
}

func onExitCodeText(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Value = data
}

func onExitLink(context *compileContext, _ *micromark.Token) {
	node := context.top()

	// Note: there are also `identifier` and `label` fields on this link node! These are used / cleaned
	// here.
	if context.data.inReference {
		referenceType := context.data.referenceType
		if referenceType == "" {
			referenceType = "shortcut"
		}

		node.NodeType += "Reference"
		node.ReferenceType = referenceType
		node.URL = ""
		node.Title = nil
	} else {
		node.Identifier = ""
		node.Label = nil
	}

	context.data.referenceType = ""
}

func onExitImage(context *compileContext, token *micromark.Token) {
	onExitLink(context, token)
}

func onExitLabelText(context *compileContext, token *micromark.Token) {
	text := context.sliceSerialize(token)
	ancestor := context.stack[len(context.stack)-2]
	label := micromark.DecodeString(text)
	ancestor.Label = &label
	ancestor.Identifier = strings.ToLower(micromark.NormalizeIdentifier(text))
}

func onExitLabel(context *compileContext, _ *micromark.Token) {
	fragment := context.top()
	value := context.resume()
	node := context.top()

	// Assume a reference.
	context.data.inReference = true

	if node.NodeType == "link" {
		node.Children = fragment.Children
	} else {
		node.Alt = &value
	}
}

func onExitResourceDestinationString(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().URL = data
}

func onExitResourceTitleString(context *compileContext, _ *micromark.Token) {
	data := context.resume()
	context.top().Title = &data
}

func onExitResource(context *compileContext, _ *micromark.Token) {
	context.data.inReference = false
}

func onEnterReference(context *compileContext, _ *micromark.Token) {
	context.data.referenceType = "collapsed"
}

func onExitReferenceString(context *compileContext, token *micromark.Token) {
	label := context.resume()
	node := context.top()
	node.Label = &label
	node.Identifier = strings.ToLower(micromark.NormalizeIdentifier(context.sliceSerialize(token)))
	context.data.referenceType = "full"
}

func onExitCharacterReferenceMarker(context *compileContext, token *micromark.Token) {
	context.data.characterReferenceType = token.Type
}

func onExitCharacterReferenceValue(context *compileContext, token *micromark.Token) {
	data := context.sliceSerialize(token)
	referenceType := context.data.characterReferenceType
	var value string

	if referenceType != "" {
		base := 16
		if referenceType == micromark.TypeCharacterReferenceMarkerNumeric {
			base = 10
		}
		value = micromark.DecodeNumericCharacterReference(data, base)
		context.data.characterReferenceType = ""
	} else {
		value, _ = micromark.DecodeNamedCharacterReference(data)
	}

	context.top().Value += value
}

func onExitCharacterReference(context *compileContext, token *micromark.Token) {
	tail := context.stack[len(context.stack)-1]
	context.stack = context.stack[:len(context.stack)-1]
	tail.Position.End = context.point(token.End)
}

func onExitAutolinkProtocol(context *compileContext, token *micromark.Token) {
	onExitData(context, token)
	context.top().URL = context.sliceSerialize(token)
}

func onExitAutolinkEmail(context *compileContext, token *micromark.Token) {
	onExitData(context, token)
	context.top().URL = "mailto:" + context.sliceSerialize(token)
}

//
// Creaters.
//

func newParent(nodeType string) func(*micromark.Token) Node {
	return func(*micromark.Token) Node {
		return Node{NodeType: nodeType, IsParent: true, Children: []*Node{}}
	}
}

func newBlockQuote(token *micromark.Token) Node { return newParent("blockquote")(token) }

func newCodeFlow(*micromark.Token) Node { return Node{NodeType: "code", IsLiteral: true} }

func newCodeText(*micromark.Token) Node { return Node{NodeType: "inlineCode", IsLiteral: true} }

func newDefinition(*micromark.Token) Node { return Node{NodeType: "definition"} }

func newHeading(*micromark.Token) Node {
	return Node{NodeType: "heading", IsParent: true, Children: []*Node{}}
}

func newHardBreak(*micromark.Token) Node { return Node{NodeType: "break"} }

func newHTML(*micromark.Token) Node { return Node{NodeType: "html", IsLiteral: true} }

func newImage(*micromark.Token) Node { return Node{NodeType: "image"} }

func newLink(*micromark.Token) Node {
	return Node{NodeType: "link", IsParent: true, Children: []*Node{}}
}

func newList(token *micromark.Token) Node {
	return Node{NodeType: "list", Ordered: token.Type == "listOrdered", Spread: token.Spread, IsParent: true, Children: []*Node{}}
}

func newListItem(token *micromark.Token) Node {
	return Node{NodeType: "listItem", Spread: token.Spread, IsParent: true, Children: []*Node{}}
}

func newThematicBreak(*micromark.Token) Node { return Node{NodeType: "thematicBreak"} }
