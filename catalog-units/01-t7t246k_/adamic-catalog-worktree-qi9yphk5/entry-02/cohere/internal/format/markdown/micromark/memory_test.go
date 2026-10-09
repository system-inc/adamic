package micromark

import (
	"slices"
	"testing"
)

// TestMemoryReuseMatchesFreshParses parses every fixture through one Memory, reset between parses, and
// compares each with a parse that allocates its own (#vbjv3d6). A reused attempt or tokenizer that kept
// anything of an earlier document shows as a difference, and the second round reuses slots that every
// kind of document has already been through.
func TestMemoryReuseMatchesFreshParses(t *testing.T) {
	t.Parallel()
	inputs := slices.Concat(eventFixtures, attentionFixtures, blockQuoteFixtures, autolinkFixtures,
		codeFencedFixtures, codeTextFixtures, characterReferenceFixtures, codeIndentedFixtures, gfmTableFixtures,
		gfmFootnoteFixtures, definitionFixtures, definitionFailureFixtures, definitionReferenceFixtures,
		gfmAutolinkLiteralFixtures, htmlTextOverrideFixtures, listFixtures, headingAtxFixtures,
		gfmTaskListItemFixtures, gfmStrikethroughFixtures, labelEndFixtures, wikiLinkFixtures,
		wikiLinkRefusalFixtures, liquidFixtures, htmlTextFixtures, htmlFlowFixtures, mathFixtures,
		setextUnderlineFixtures)
	memory := NewMemory()
	for round := 1; round <= 2; round++ {
		for _, input := range inputs {
			expected := portEvents(input)
			actual := portEventsFrom(input, memory)
			memory.Reset()
			if difference := firstEventDifference(expected, actual); difference != "" {
				t.Fatalf("round %d, input %q: %s", round, truncateInput(input), difference)
			}
		}
	}
}

// TestMemoryResetReleasesTokens proves the comparison above can fail: events read after their Memory was
// reset no longer describe the parse.
func TestMemoryResetReleasesTokens(t *testing.T) {
	t.Parallel()
	memory := NewMemory()
	events := Parse(SourceUnits("# a *b*\n\n- c\n"), MarkdownConstructs(), memory)
	before := events[0].Token.Type
	memory.Reset()
	if after := events[0].Token.Type; after == before {
		t.Fatalf("the first token still reads %q after its memory was reset", after)
	}
}
