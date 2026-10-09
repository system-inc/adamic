package micromark

import "testing"

// htmlTextOverrideFixtures exercise the fork's html-text copy through the full extension set, the way
// Prettier parses. They avoid the same pending constructs htmlTextFixtures do (see there), and center on
// the one changed state: the whitespace after a line ending inside the construct, which core makes a
// linePrefix and the copy keeps in htmlTextData, after every state that can see a line ending.
//
// On `<` the copy is tried first, then autolink, then core htmlText. The copy and core accept the same
// inputs, so core never wins here: where the copy fails, core is tried and fails the same way (the "all
// fail" fixtures run that path). Core winning is covered by TestHtmlTextEvents, without the override.
var htmlTextOverrideFixtures = []string{
	// Which construct wins on `<`: autolink, the copy, or none of the three.
	"a <ab:c> d",
	"a <b@c> d",
	"a <b c> d",
	"a <ab:c><b c> d",
	"a <b c><ab:c> d",
	"a <b:c> d",
	"a <b c:d> e",
	"a <ab:c d> e",
	"a <b@c d> e",
	"a <b-c@d> e",
	"a <b/c> d",

	// The changed state, after each return state: tagOpenBetween.
	"a <b\n c> d",
	"a <b\n   c> d",
	"a <b\n    c> d",
	"a <b\n     c> d",
	"a <b\n\tc> d",
	"a <b\n \tc> d",
	"a <b\r\n  c> d",
	"a <b\r  c> d",

	// tagOpenAttributeNameAfter, tagOpenAttributeValueBefore.
	"a <b c\n  =d> e",
	"a <b c\n\t=d> e",
	"a <b c=\n  d> e",
	"a <b c=\n\t\"d\"> e",

	// tagOpenAttributeValueQuoted, the case the fork's header names.
	"a <b c=\"d\n  e\"> f",
	"a <b c='d\n  e'> f",
	"a <b c=\"d\n\te\"> f",
	"a <b c=\"d\n      e\"> f",
	"a <b c=\"d\r\n  e\"> f",
	"a <b c=\"d\r  e\"> f",
	"a <b c=\"d\n  e\n  f\"> g",
	"a <b c=\"\n  \"> d",
	"a <b c=\"d\n  ",

	// The other return states: comment, CDATA, declaration, instruction, tagCloseBetween.
	"a <!--b\n  c--> d",
	"a <!--b\n\t--> c",
	"a <![CDATA[b\n  c]]> d",
	"a <!b\n  c> d",
	"a <?b\n  c?> d",
	"a </b\n  > c",
	"a </b\n\t> c",

	// Inside other blocks: a block quote's continuation, a heading's single line, a setext heading.
	"> a <b c=\"d\n>   e\"> f",
	"> a <b\n>\tc> d",
	"# a <b c=\"d\"> e",
	"a <b c=\"d\n  e\">\n===",
	"a <b c=\"d\n  e\"> f\n\ng <h\n  i> j",
}

func TestHtmlTextOverrideEvents(t *testing.T) {
	t.Parallel()
	compareEvents(t, htmlTextOverrideFixtures)
	compareEvents(t, htmlTextFixtures)
}

// TestHtmlTextOverrideOrder checks the combined text constructs on `<`: the copy, then autolink, then core.
// mergeConstructs puts every construct not marked "after" before the existing ones, so `add: "before"`
// holds the copy ahead of the defaults.
func TestHtmlTextOverrideOrder(t *testing.T) {
	t.Parallel()
	constructs := MarkdownConstructs().Text.ByCode[CodeLessThan]
	want := []*Construct{htmlTextOverride, autolink, htmlText}
	if len(constructs) != len(want) {
		t.Fatalf("%d constructs on `<`, want %d", len(constructs), len(want))
	}
	for index := range want {
		if constructs[index] != want[index] {
			t.Errorf("construct %d on `<` is %q, not the expected one", index, constructs[index].Name)
		}
	}
}
