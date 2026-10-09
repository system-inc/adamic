package micromark

// defaultConstructs is micromark/lib/constructs.js: CommonMark's constructs by the code they start at.
func defaultConstructs() *Extension {
	return &Extension{
		Document: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeAsterisk:    {list},
			CodePlusSign:    {list},
			CodeDash:        {list},
			CodeDigit0:      {list},
			CodeDigit1:      {list},
			CodeDigit2:      {list},
			CodeDigit3:      {list},
			CodeDigit4:      {list},
			CodeDigit5:      {list},
			CodeDigit6:      {list},
			CodeDigit7:      {list},
			CodeDigit8:      {list},
			CodeDigit9:      {list},
			CodeGreaterThan: {blockQuote},
		}},
		ContentInitial: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeLeftSquareBracket: {definition},
		}},
		FlowInitial: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeHorizontalTab: {codeIndented},
			CodeVirtualSpace:  {codeIndented},
			CodeSpace:         {codeIndented},
		}},
		Flow: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeNumberSign:  {headingAtx},
			CodeAsterisk:    {thematicBreak},
			CodeDash:        {setextUnderline, thematicBreak},
			CodeLessThan:    {htmlFlow},
			CodeEqualsTo:    {setextUnderline},
			CodeUnderscore:  {thematicBreak},
			CodeGraveAccent: {codeFenced},
			CodeTilde:       {codeFenced},
		}},
		String: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeAmpersand: {characterReference},
			CodeBackslash: {characterEscape},
		}},
		Text: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeCarriageReturn:         {lineEnding},
			CodeLineFeed:               {lineEnding},
			CodeCarriageReturnLineFeed: {lineEnding},
			CodeExclamationMark:        {labelStartImage},
			CodeAmpersand:              {characterReference},
			CodeAsterisk:               {attention},
			CodeLessThan:               {autolink, htmlText},
			CodeLeftSquareBracket:      {labelStartLink},
			CodeBackslash:              {hardBreakEscape, characterEscape},
			CodeRightSquareBracket:     {labelEnd},
			CodeUnderscore:             {attention},
			CodeGraveAccent:            {codeText},
		}},
		InsideSpan:       []*Construct{attention, resolveTextConstruct},
		AttentionMarkers: []Code{CodeAsterisk, CodeUnderscore},
		Disable:          []string{},
	}
}

// MarkdownExtensions are the syntax extensions the fork's parseMarkdown passes to micromark, in its order:
// gfm(), math(), the wiki-link syntax, the fork's liquid syntax and its html-text override
// (src/language-markdown/parse/parse-markdown.js).
func MarkdownExtensions() []*Extension {
	return []*Extension{
		gfmExtension(),
		mathExtension(),
		wikiLinkExtension(),
		liquidExtension(),
		htmlTextOverrideExtension(),
	}
}

// gfmExtension is micromark-extension-gfm's gfm(): five extensions combined into one.
func gfmExtension() *Extension {
	return combineExtensions([]*Extension{
		gfmAutolinkLiteralExtension(),
		gfmFootnoteExtension(),
		gfmStrikethroughExtension(),
		gfmTableExtension(),
		gfmTaskListItemExtension(),
	})
}
