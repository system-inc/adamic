package micromark

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// htmlTextFixtures exercise core html-text.js. Through the fork's extension set core htmlText is never
// what matches: the fork's copy is tried first on `<` and accepts the same inputs, so core only ever runs
// after it has already failed. TestHtmlTextEvents therefore compares against an oracle run without the
// override (overrideHtmlTextSyntax() dropped from the Node options, htmlTextOverrideExtension dropped from
// the Go list), where core is the one that wins. TestHtmlTextOverrideEvents runs these same inputs again
// through the full extension set.
//
// They avoid the constructs still pending, outside a matched tag (inside one, htmlText consumes the
// characters and nothing else is tried): attention (`*`, `_`), label ends and so any `[` or `]` (an
// unmatched `[` needs labelEnd's resolveAll to fall back to data), definitions, lists (no line starts with
// `-`, `+`, `*` or a digit and a dot), math (`$`), liquid (`{`), wiki links, strikethrough (`~`), tables
// (`|`), footnotes, the GFM autolink literals (no `www.`, `http://` or `a@b.c` left over by a failure) and
// htmlFlow: every first line starts with text, and no later line starts with a `<` that a kind 1 to 6
// HTML block could open on (kind 7 cannot interrupt a paragraph). CommonMark examples 618, 619 (a failure
// leaves `_` or `*`) and 627 (`$` around a failure would be math) are left out for that reason.
var htmlTextFixtures = []string{
	// CommonMark 0.31.2 "Raw HTML" examples, the ones that stay inside ported constructs.
	"<a><bab><c2c>\n",
	"<a/><b2/>\n",
	"<a  /><b2\ndata=\"foo\" >\n",
	"<a foo=\"bar\" bam = 'baz <em>\"</em>'\n_boolean zoop:33=zoop:33 />\n",
	"Foo <responsive-image src=\"foo.jpg\" />\n",
	"<a href=\"hi'> <a href=hi'>\n",
	"< a><\nfoo><bar/ >\n<foo bar=baz\nbim!bop />\n",
	"<a href='bar'title=title>\n",
	"</a></foo >\n",
	"</a href=\"foo\">\n",
	"foo <!-- this is a --\ncomment - with hyphens -->\n",
	"foo <!--> foo -->\n\nfoo <!---> foo -->\n",
	"foo <!ELEMENT br EMPTY>\n",
	"foo <![CDATA[>&<]]>\n",
	"foo <a href=\"&ouml;\">\n",
	"foo <a href=\"\\*\">\n",
	"<a href=\"\\\"\">\n",

	// open: what may follow `<`.
	"a <",
	"a <\nb",
	"a <1> b",
	"a < b> c",
	"a <é> b",
	"a <𝔸> b",
	"a <b:c> d",

	// declarationOpen, commentOpenInside, comment, commentClose, commentEnd.
	"a <!> b",
	"a <!",
	"a <!-x> b",
	"a <!-",
	"a <!--",
	"a <!--> b",
	"a <!---> b",
	"a <!----> b",
	"a <!--b--> c",
	"a <!--b-c--> d",
	"a <!--b--c--> d",
	"a <!--b---> c",
	"a <!--b",
	"a <!--b-",
	"a <!--b--",
	"a <!--b\nc--> d",
	"a <!--b\r\nc--> d",
	"a <!--b\rc--> d",
	"a <!--b\n\n--> c",
	"a <!--é𝔸--> b",

	// cdataOpenInside, cdata, cdataClose, cdataEnd.
	"a <![CDATA[b]]> c",
	"a <![CDATA[]]> b",
	"a <![CDATA[b]c]]> d",
	"a <![CDATA[b]]]> c",
	"a <![CDATA[b]]c]]> d",
	"a <![CDATA[b\nc]]> d",
	"a <![CDATA[<b>]]> c",

	// declaration.
	"a <!DOCTYPE html> b",
	"a <!b> c",
	"a <!b",
	"a <!b\nc> d",
	"a <!1> b",

	// instruction, instructionClose.
	"a <?> b",
	"a <??> b",
	"a <?b?> c",
	"a <?b?c?> d",
	"a <?",
	"a <?b?",
	"a <?b\nc?> d",

	// tagCloseStart, tagClose, tagCloseBetween.
	"a </b> c",
	"a </b-1> c",
	"a </b \t> c",
	"a </b\n> c",
	"a </b c> d",
	"a </b/> c",
	"a </1> b",
	"a </> b",
	"a </",
	"a </b",
	"a </b ",

	// tagOpen.
	"a <b> c",
	"a <b-c1> d",
	"a <b/> c",
	"a <b /> c",
	"a <b/ > c",
	"a <b.c> d",
	"a <bé> c",
	"a <b",
	"a <b\n> c",

	// tagOpenBetween, tagOpenAttributeName, tagOpenAttributeNameAfter.
	"a <b c> d",
	"a <b :c> d",
	"a <b c d> e",
	"a <b c-d.e:f9> g",
	"a <b 1> c",
	"a <b c",
	"a <b c ",
	"a <b\nc=d> e",
	"a <b c\n=d> e",

	// tagOpenAttributeValueBefore, the quoted and unquoted values, tagOpenAttributeValueQuotedAfter.
	"a <b c=d> e",
	"a <b c = d> e",
	"a <b c=\"d\"> e",
	"a <b c='d'> e",
	"a <b c=\"d'e\"> f",
	"a <b c=\"\"> d",
	"a <b c=d/> e",
	"a <b c=\"d\"/> e",
	"a <b c=\"d\"e> f",
	"a <b c=> d",
	"a <b c=<> d",
	"a <b c==> d",
	"a <b c=`> d",
	"a <b c=d\"> e",
	"a <b c=d'e> f",
	"a <b c=d<e> f",
	"a <b c=d=e> f",
	"a <b c=d`e> f",
	"a <b c=é> d",
	"a <b c=\"𝔸\"> d",
	"a <b c=",
	"a <b c=d",
	"a <b c=\"d",
	"a <b c=\nd> e",
	"a <b c=d\n> e",
	"a <b c=\"d\ne\"> f",
	"a <b c=\"d\r\ne\"> f",
	"a <b c='d\re'> f",
	"a <b c=\"d\"\ne> f",
	"a <b\n\n> c",

	// lineEndingAfter: the next line's indent, a linePrefix of at most 3 columns here (factorySpace max 4).
	"a <b\n c> d",
	"a <b\n   c> d",
	"a <b\n    c> d",
	"a <b\n     c> d",
	"a <b\n\tc> d",
	"a <b\n \tc> d",
	"a <!--b\n   c--> d",
	"a <!--b\n     c--> d",
	"a <b c=\"d\n  e\"> f",
	"a <b c=\n   \"d\"> e",

	// Around it: other text, escapes, headings, block quotes, setext, several in a row.
	"a <b><ab:c> d",
	"a <ab:c><b> d",
	"a \\<b> c",
	"a `<b>` c",
	"a&amp;<b>&amp;c",
	"é<b>𝔸",
	"# a <b> c",
	"# a <b c=\"d\"> #",
	"> a <b\n> c=\"d\"> e",
	"> a <b c=\"d\n> e\"> f",
	"a <b c=\"d\n  e\">\n===",
	"a <b><c></b></c>",
	"a  \n<b> c",
	"a\\\n<b> c",
}

func TestHtmlTextEvents(t *testing.T) {
	t.Parallel()
	compareEventsWithoutHtmlTextOverride(t, htmlTextFixtures)
}

// compareEventsWithoutHtmlTextOverride is compareEvents with the fork's html-text override left out of
// both sides, so core htmlText is the construct that runs on `<` after autolink.
func compareEventsWithoutHtmlTextOverride(t *testing.T, inputs []string) {
	t.Helper()
	script := strings.Replace(oracleScript, "    overrideHtmlTextSyntax(),\n", "", 1)
	if script == oracleScript {
		t.Fatal("the oracle script no longer lists overrideHtmlTextSyntax() where this test removes it")
	}
	expected := htmlTextUpstreamEvents(t, script, inputs)
	extensions := []*Extension{gfmExtension(), mathExtension(), wikiLinkExtension(), liquidExtension()}
	for index, input := range inputs {
		actual := htmlTextPortEvents(input, extensions)
		if difference := firstEventDifference(expected[index], actual); difference != "" {
			t.Errorf("input %q: %s", truncateInput(input), difference)
		}
	}
}

// htmlTextUpstreamEvents is oracle_test.go's upstreamEvents with the script as a parameter.
func htmlTextUpstreamEvents(t *testing.T, script string, inputs []string) [][]string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed; the event oracle needs it")
	}
	root := forkRoot(t)

	path := filepath.Join(t.TempDir(), "oracle.mjs")
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(script, "FORK", root)), 0o644); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", path)
	command.Stdin = strings.NewReader(string(encoded))
	output, err := command.Output()
	if err != nil {
		if exitError, isExit := err.(*exec.ExitError); isExit {
			t.Fatalf("the event oracle failed: %v\n%s", err, exitError.Stderr)
		}
		t.Fatal(err)
	}
	var outputs [][]string
	if err := json.Unmarshal(output, &outputs); err != nil {
		t.Fatal(err)
	}
	return outputs
}

// htmlTextPortEvents is oracle_test.go's portEvents with the extensions as a parameter.
func htmlTextPortEvents(input string, extensions []*Extension) (described []string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			described = []string{fmt.Sprintf("panic: %v", recovered)}
		}
	}()
	for _, event := range Parse(SourceUnits(input), Combine(extensions), nil) {
		kind := "exit"
		if event.Enter {
			kind = "enter"
		}
		start, end := event.Token.Start, event.Token.End
		described = append(described, fmt.Sprintf("%s %s %d:%d:%d-%d:%d:%d", kind, event.Token.Type,
			start.Line, start.Column, start.Offset, end.Line, end.Column, end.Offset))
	}
	return described
}
