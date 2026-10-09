package micromark

// mathExtension is micromark-extension-math/lib/syntax.js, `math()`.
//
// The fork calls `math()` with no options (src/language-markdown/parse/parse-markdown.js), so
// `singleDollarTextMath` is upstream's default: undefined, which mathText reads as true.
func mathExtension() *Extension {
	return &Extension{
		Flow: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeDollarSign: {mathFlow},
		}},
		Text: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeDollarSign: {mathText(true)},
		}},
	}
}
