package micromark

import (
	"sync"
	"unicode/utf16"
)

// Parse tokenizes markdown into events, the way mdast-util-from-markdown drives micromark:
// postprocess(parse(options).document().write(preprocess()(value, encoding, true))).
//
// source is UTF-16 code units. Every point in the result counts them. The tokens and tokenizers come from
// memory, which may be nil; the events are valid until it is Reset.
//
// constructs is the defaults combined with the extensions (Combine), made once rather than per parse.
func Parse(source []uint16, constructs *FullConstructs, memory *Memory) []Event {
	parser := &ParseContext{Constructs: constructs, Lazy: map[int]bool{}, memory: memory}
	return postprocess(parser.create(ContentTypeDocument, nil).Write(preprocess(source)))
}

// SourceUnits converts text to the UTF-16 code units micromark reads, the way a JavaScript string holds it.
func SourceUnits(text string) []uint16 {
	return AppendSourceUnits(nil, text)
}

// AppendSourceUnits is SourceUnits appended to units, so a caller can reuse one buffer for every text.
func AppendSourceUnits(units []uint16, text string) []uint16 {
	for _, character := range text {
		units = utf16.AppendRune(units, character)
	}
	return units
}

// Combine is the constructs of upstream's parse(), micromark/lib/parse.js: the defaults combined with the
// extensions, which a parse only reads.
func Combine(extensions []*Extension) *FullConstructs {
	return combineExtensions(append([]*Extension{defaultConstructs()}, extensions...))
}

// MarkdownConstructs is Combine(MarkdownExtensions()), made once: every markdown parse reads the same
// table, and combining it was a measurable share of a format's allocation (#93dpede).
var MarkdownConstructs = sync.OnceValue(func() *FullConstructs { return Combine(MarkdownExtensions()) })

// create makes a tokenizer for a content type, upstream's parser.document() and friends.
func (parser *ParseContext) create(contentType string, from *Point) *TokenizeContext {
	var initial *InitialConstruct
	switch contentType {
	case ContentTypeDocument:
		initial = documentInitial
	case ContentTypeFlow:
		initial = flowInitial
	case ContentTypeContent:
		initial = contentInitial
	case ContentTypeString:
		initial = stringInitial
	case ContentTypeText:
		initial = textInitial
	default:
		panic("micromark: unknown content type " + contentType)
	}
	return createTokenizer(parser, initial, from)
}

// combineExtensions is micromark-util-combine-extensions.
func combineExtensions(extensions []*Extension) *Extension {
	all := &Extension{
		Document:       &ConstructRecord{ByCode: map[Code][]*Construct{}},
		ContentInitial: &ConstructRecord{ByCode: map[Code][]*Construct{}},
		FlowInitial:    &ConstructRecord{ByCode: map[Code][]*Construct{}},
		Flow:           &ConstructRecord{ByCode: map[Code][]*Construct{}},
		String:         &ConstructRecord{ByCode: map[Code][]*Construct{}},
		Text:           &ConstructRecord{ByCode: map[Code][]*Construct{}},
	}

	for _, extension := range extensions {
		mergeRecord(all.Document, extension.Document)
		mergeRecord(all.ContentInitial, extension.ContentInitial)
		mergeRecord(all.FlowInitial, extension.FlowInitial)
		mergeRecord(all.Flow, extension.Flow)
		mergeRecord(all.String, extension.String)
		mergeRecord(all.Text, extension.Text)
		all.InsideSpan = mergeConstructs(all.InsideSpan, extension.InsideSpan)
		// Codes and names carry no `add`, so upstream's constructs() puts every one before the existing.
		all.AttentionMarkers = append(append([]Code(nil), extension.AttentionMarkers...), all.AttentionMarkers...)
		all.Disable = append(append([]string(nil), extension.Disable...), all.Disable...)
	}

	for _, record := range []*ConstructRecord{all.Document, all.ContentInitial, all.FlowInitial, all.Flow, all.String, all.Text} {
		record.combine()
	}

	return all
}

func mergeRecord(left *ConstructRecord, right *ConstructRecord) {
	if right == nil {
		return
	}
	// Upstream iterates the right record's keys in insertion order. The order of keys never matters,
	// only the order within a key's list, which mergeConstructs keeps.
	for code, list := range right.ByCode {
		left.ByCode[code] = mergeConstructs(left.ByCode[code], list)
	}
	if right.Null != nil {
		left.Null = mergeConstructs(left.Null, right.Null)
	}
}

// mergeConstructs is upstream's constructs(): new constructs go before the existing ones, unless they
// ask to go after.
func mergeConstructs(existing []*Construct, list []*Construct) []*Construct {
	var before []*Construct
	result := append([]*Construct(nil), existing...)
	for _, construct := range list {
		if construct.Add == "after" {
			result = append(result, construct)
		} else {
			before = append(before, construct)
		}
	}
	return append(before, result...)
}

// preprocess is upstream's preprocess(), micromark/lib/preprocess.js, for a single call with end set.
func preprocess(value []uint16) []Chunk {
	var chunks []Chunk
	column := 1
	atCarriageReturn := false
	startPosition := 0

	// To do: `markdown-rs` actually parses BOMs (byte order mark).
	if len(value) > 0 && value[0] == 0xFEFF {
		startPosition++
	}

	buffer := []uint16(nil)
	for startPosition < len(value) {
		endPosition := startPosition
		for endPosition < len(value) {
			unit := value[endPosition]
			if unit == 0 || unit == '\t' || unit == '\n' || unit == '\r' {
				break
			}
			endPosition++
		}

		if endPosition == len(value) {
			buffer = value[startPosition:]
			break
		}
		code := value[endPosition]

		if code == '\n' && startPosition == endPosition && atCarriageReturn {
			chunks = append(chunks, codeChunk(CodeCarriageReturnLineFeed))
			atCarriageReturn = false
		} else {
			if atCarriageReturn {
				chunks = append(chunks, codeChunk(CodeCarriageReturn))
				atCarriageReturn = false
			}

			if startPosition < endPosition {
				chunks = append(chunks, textChunk(value[startPosition:endPosition]))
				column += endPosition - startPosition
			}

			switch code {
			case 0:
				chunks = append(chunks, codeChunk(CodeReplacementCharacter))
				column++
			case '\t':
				next := ((column + tabSize - 1) / tabSize) * tabSize
				chunks = append(chunks, codeChunk(CodeHorizontalTab))
				for column < next {
					chunks = append(chunks, codeChunk(CodeVirtualSpace))
					column++
				}
				column++
			case '\n':
				chunks = append(chunks, codeChunk(CodeLineFeed))
				column = 1
			default:
				atCarriageReturn = true
				column = 1
			}
		}

		startPosition = endPosition + 1
	}

	if atCarriageReturn {
		chunks = append(chunks, codeChunk(CodeCarriageReturn))
	}
	if len(buffer) > 0 {
		chunks = append(chunks, textChunk(buffer))
	}
	chunks = append(chunks, codeChunk(CodeEof))

	return chunks
}

// postprocess is upstream's postprocess(): subtokenize until nothing is left.
func postprocess(events []Event) []Event {
	for {
		var done bool
		events, done = subtokenize(events)
		if done {
			return events
		}
	}
}
