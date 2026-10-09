package core

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noWarningCommentsFile is where the fixtures pretend to live.
const noWarningCommentsFile = "/repository/source/NoWarningComments.ts"

// noWarningCommentsCase is one imported corpus row.
type noWarningCommentsCase struct {
	sourceText string
	options    any
	wantIds    []string
	// wantTerms is the matched term each finding names, in order. That is half the rendered
	// message and no message-id assertion can see it: a rule matching the wrong term of a
	// configured set reports the right count with the wrong word.
	wantTerms []string
}

// runNoWarningComments drives one case through the rule's own exported decoder.
//
// Through the decoder rather than by building the struct, because the decoder is where the location
// enum is validated and where a decoration entry is checked for being exactly one non-whitespace
// character. Neither has an upstream counterpart to inherit correctness from.
func runNoWarningComments(t *testing.T, testCase noWarningCommentsCase) rule_testing.Result {
	t.Helper()
	if testCase.options == nil {
		return rule_testing.Run(t, NoWarningComments, noWarningCommentsFile, testCase.sourceText)
	}
	encoded, err := json.Marshal(testCase.options)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeNoWarningCommentsOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, NoWarningComments, noWarningCommentsFile,
		testCase.sourceText, decoded)
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/no-warning-comments.js` was loaded with its RuleTester stubbed so every case came
// out as data with its options attached, then replayed against the INSTALLED rule to record what it
// reports and which term each finding names. All 61 reproduced, and all 61 are importable -- this
// rule reads only comments, so nothing about it depends on a parser feature or a source type the
// harness cannot express.
func noWarningCommentsFiresCases() []noWarningCommentsCase {
	return []noWarningCommentsCase{
		{"// fixme", nil, []string{"unexpectedComment"}, []string{"fixme"}},
		{"// any fixme", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"fixme"}},
		{"// any fixme", NoWarningCommentsOptions{Terms: []string{"fixme"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"fixme"}},
		{"// any FIXME", NoWarningCommentsOptions{Terms: []string{"fixme"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"fixme"}},
		{"// any fIxMe", NoWarningCommentsOptions{Terms: []string{"fixme"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"fixme"}},
		{"/* any fixme */", NoWarningCommentsOptions{Terms: []string{"FIXME"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"FIXME"}},
		{"/* any FIXME */", NoWarningCommentsOptions{Terms: []string{"FIXME"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"FIXME"}},
		{"/* any fIxMe */", NoWarningCommentsOptions{Terms: []string{"FIXME"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"FIXME"}},
		{"// any fixme or todo", NoWarningCommentsOptions{Terms: []string{"fixme", "todo"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment"}, []string{"fixme", "todo"}},
		{"/* any fixme or todo */", NoWarningCommentsOptions{Terms: []string{"fixme", "todo"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment"}, []string{"fixme", "todo"}},
		{"/* any fixme or todo */", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment"}, []string{"todo", "fixme"}},
		{"/* fixme and todo */", nil, []string{"unexpectedComment"}, []string{"fixme"}},
		{"/* fixme and todo */", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment"}, []string{"todo", "fixme"}},
		{"/* any fixme */", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"fixme"}},
		{"/* fixme! */", NoWarningCommentsOptions{Terms: []string{"fixme"}}, []string{"unexpectedComment"}, []string{"fixme"}},
		{"// regex [litera|$]", NoWarningCommentsOptions{Terms: []string{"[litera|$]"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"[litera|$]"}},
		{"/* eslint one-var: 2 */", NoWarningCommentsOptions{Terms: []string{"eslint"}}, []string{"unexpectedComment"}, []string{"eslint"}},
		{"/* eslint one-var: 2 */", NoWarningCommentsOptions{Terms: []string{"one"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"one"}},
		{"/* any block comment with TODO, FIXME or XXX */", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment", "unexpectedComment"}, []string{"todo", "fixme", "xxx"}},
		{"/* any block comment with (TODO, FIXME's or XXX!) */", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment", "unexpectedComment"}, []string{"todo", "fixme", "xxx"}},
		{"/** \n *any block comment \n*with (TODO, FIXME's or XXX!) **/", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment", "unexpectedComment"}, []string{"todo", "fixme", "xxx"}},
		{"// any comment with TODO, FIXME or XXX", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment", "unexpectedComment", "unexpectedComment"}, []string{"todo", "fixme", "xxx"}},
		{"// TODO: something small", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo"}},
		{"// TODO: something really longer than 40 characters", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo"}},
		{"/* TODO: something \n really longer than 40 characters \n and also a new line */", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo"}},
		{"// TODO: small", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo"}},
		{"// https://github.com/eslint/eslint/pull/13522#discussion_r470293411 TODO", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo"}},
		{"// Comment ending with term followed by punctuation TODO!", NoWarningCommentsOptions{Terms: []string{"todo"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo"}},
		{"// Comment ending with term including punctuation TODO!", NoWarningCommentsOptions{Terms: []string{"todo!"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo!"}},
		{"// Comment ending with term including punctuation followed by more TODO!!!", NoWarningCommentsOptions{Terms: []string{"todo!"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo!"}},
		{"// !TODO comment starting with term preceded by punctuation", NoWarningCommentsOptions{Terms: []string{"todo"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"todo"}},
		{"// !TODO comment starting with term including punctuation", NoWarningCommentsOptions{Terms: []string{"!todo"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"!todo"}},
		{"// !!!TODO comment starting with term including punctuation preceded by more", NoWarningCommentsOptions{Terms: []string{"!todo"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"!todo"}},
		{"// FIX!term ending with punctuation followed word character", NoWarningCommentsOptions{Terms: []string{"FIX!"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"FIX!"}},
		{"// Term starting with punctuation preceded word character!FIX", NoWarningCommentsOptions{Terms: []string{"!FIX"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"!FIX"}},
		{"//!XXX comment starting with no spaces (anywhere)", NoWarningCommentsOptions{Terms: []string{"!xxx"}, Location: NoWarningCommentsAnywhere}, []string{"unexpectedComment"}, []string{"!xxx"}},
		{"//!XXX comment starting with no spaces (start)", NoWarningCommentsOptions{Terms: []string{"!xxx"}, Location: NoWarningCommentsStart}, []string{"unexpectedComment"}, []string{"!xxx"}},
		{"/*\nTODO undecorated multi-line block comment (start)\n*/", NoWarningCommentsOptions{Terms: []string{"todo"}, Location: NoWarningCommentsStart}, []string{"unexpectedComment"}, []string{"todo"}},
		{"///// TODO decorated single-line comment with decoration array \n /////", NoWarningCommentsOptions{Terms: []string{"todo"}, Location: NoWarningCommentsStart, Decoration: []string{"*", "/"}}, []string{"unexpectedComment"}, []string{"todo"}},
		{"///*/*/ TODO decorated single-line comment with multiple decoration characters (start) \n /////", NoWarningCommentsOptions{Terms: []string{"todo"}, Location: NoWarningCommentsStart, Decoration: []string{"*", "/"}}, []string{"unexpectedComment"}, []string{"todo"}},
		{"//**TODO term starts with a decoration character", NoWarningCommentsOptions{Terms: []string{"*todo"}, Location: NoWarningCommentsStart, Decoration: []string{"*"}}, []string{"unexpectedComment"}, []string{"*todo"}},
	}
}

func noWarningCommentsSilentCases() []noWarningCommentsCase {
	return []noWarningCommentsCase{
		{"// any comment", NoWarningCommentsOptions{Terms: []string{"fixme"}}, nil, nil},
		{"// any comment", NoWarningCommentsOptions{Terms: []string{"fixme", "todo"}}, nil, nil},
		{"// any comment", nil, nil, nil},
		{"// any comment", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, nil, nil},
		{"// any comment with TODO, FIXME or XXX", NoWarningCommentsOptions{Location: NoWarningCommentsStart}, nil, nil},
		{"// any comment with TODO, FIXME or XXX", nil, nil, nil},
		{"/* any block comment */", NoWarningCommentsOptions{Terms: []string{"fixme"}}, nil, nil},
		{"/* any block comment */", NoWarningCommentsOptions{Terms: []string{"fixme", "todo"}}, nil, nil},
		{"/* any block comment */", nil, nil, nil},
		{"/* any block comment */", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, nil, nil},
		{"/* any block comment with TODO, FIXME or XXX */", NoWarningCommentsOptions{Location: NoWarningCommentsStart}, nil, nil},
		{"/* any block comment with TODO, FIXME or XXX */", nil, nil, nil},
		{"/* any block comment with (TODO, FIXME's or XXX!) */", nil, nil, nil},
		{"// comments containing terms as substrings like TodoMVC", NoWarningCommentsOptions{Terms: []string{"todo"}, Location: NoWarningCommentsAnywhere}, nil, nil},
		{"// special regex characters don't cause a problem", NoWarningCommentsOptions{Terms: []string{"[aeiou]"}, Location: NoWarningCommentsAnywhere}, nil, nil},
		{"/*eslint no-warning-comments: [2, { \"terms\": [\"todo\", \"fixme\", \"any other term\"], \"location\": \"anywhere\" }]*/\n\nvar x = 10;\n", nil, nil, nil},
		{"/*eslint no-warning-comments: [2, { \"terms\": [\"todo\", \"fixme\", \"any other term\"], \"location\": \"anywhere\" }]*/\n\nvar x = 10;\n", NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}, nil, nil},
		{"// foo", NoWarningCommentsOptions{Terms: []string{"foo-bar"}}, nil, nil},
		{"/** multi-line block comment with lines starting with\nTODO\nFIXME or\nXXX\n*/", nil, nil, nil},
		{"//!TODO ", NoWarningCommentsOptions{Decoration: []string{"*"}}, nil, nil},
	}
}

func TestNoWarningCommentsFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range noWarningCommentsFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoWarningComments(t, testCase), testCase.wantIds...)
		})
	}
}

func TestNoWarningCommentsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range noWarningCommentsSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoWarningComments(t, testCase))
		})
	}
}

// Which TERM each finding names, which the message-id assertions above cannot see.
//
// The term is chosen from the configured set by which pattern matched, so a rule that matched on the
// right comment through the wrong term reports the right count with the wrong word in the message.
// The eight multi-finding cases are where this bites: a comment matching two configured terms
// reports twice and the two findings must name different terms, in the order the terms were
// configured.
func TestNoWarningCommentsNamesTheMatchedTerm(t *testing.T) {
	t.Parallel()

	for _, testCase := range noWarningCommentsFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoWarningComments(t, testCase)
			if len(result.Diagnostics) != len(testCase.wantTerms) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantTerms),
					len(result.Diagnostics))
			}
			for index, wantTerm := range testCase.wantTerms {
				got := result.Diagnostics[index].Message.Description
				// The term is rendered inside backticks, so the needle is anchored on both
				// sides. A bare Contains on the term alone would pass on a message that merely
				// quoted a comment containing it.
				needle := "opens with `" + wantTerm + "`"
				if !strings.Contains(got, needle) {
					t.Errorf("finding %d rendered\n  %q\nwhich does not name the term %q",
						index, got, wantTerm)
				}
			}
		})
	}
}

// The quoted comment in the message, truncated the way upstream truncates it.
//
// Upstream rebuilds the quote word by word and stops BEFORE exceeding forty characters, appending an
// ellipsis when it stopped early. Two consequences, and both are in its corpus:
//
//	the quote never splits a word, so the cut lands on whatever boundary comes before the limit
//	a comment whose FIRST word already exceeds the limit quotes as nothing but the ellipsis
//
// The second looks like a defect and is upstream's behaviour, reached by a real case: a comment
// carrying a long URL followed by the term. Reproduced rather than corrected, because the quoted
// text is the part of the message a reader scans and a port that helpfully showed the URL would
// disagree with the tool it replaces on exactly the comments people write.
//
// Every expectation here was read off the installed rule's own rendered message.
func TestNoWarningCommentsTruncatesOnAWordBoundary(t *testing.T) {
	t.Parallel()

	anywhere := NoWarningCommentsOptions{Location: NoWarningCommentsAnywhere}
	cases := []struct {
		sourceText string
		options    any
		wantQuote  string
	}{
		// Short enough to survive whole.
		{"// TODO: fix this", nil, "TODO: fix this"},
		// Exactly the shape that cuts, from upstream's corpus. It needs Anywhere: the term is
		// not first, so the default Start location leaves it clean, which is itself the
		// discrimination the location option exists for.
		{"/* any block comment with TODO, FIXME or XXX */", anywhere,
			"any block comment with TODO, FIXME or..."},
		// Whitespace collapses, so a multi-line comment quotes as one line. It needs the
		// asterisk allowed as decoration, since the leading `*` of the second line otherwise
		// sits between the anchor and the term.
		{"/*\n * TODO fix\n * this\n */", NoWarningCommentsOptions{Decoration: []string{"*"}},
			"* TODO fix * this"},
		// The pathological case: the first word alone exceeds the limit, so nothing survives.
		{"// https://github.com/eslint/eslint/pull/13522#discussion_r470293411 TODO", anywhere,
			"..."},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoWarningComments(t, noWarningCommentsCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) == 0 {
				t.Fatal("wanted at least one finding")
			}
			// EVERY finding on one comment quotes that same comment, so the quote is asserted on
			// all of them rather than on the first. A comment carrying three configured terms
			// reports three times, and a rule that computed the quote per-term rather than
			// per-comment would differ only on the second and third.
			//
			// The quote is rendered inside Go's %q, so it arrives escaped and quoted. Asserting
			// the exact needle rather than a Contains on the bare text, because a weaker predicate
			// than the property it guards is not a guard.
			needle := ": " + strconv.Quote(testCase.wantQuote) + "."
			for index, diagnostic := range result.Diagnostics {
				if !strings.Contains(diagnostic.Message.Description, needle) {
					t.Errorf("finding %d rendered\n  %q\nwhich does not carry the quote %s",
						index, diagnostic.Message.Description, needle)
				}
			}
		})
	}
}

// Word boundaries are ASCII on both sides, which is measured rather than inherited.
//
// Go's `\b` is ASCII-only. JavaScript's is too, even under the `u` flag upstream passes, and that is
// not obvious: the flag changes how the pattern is interpreted in several other ways, and a port
// that assumed `u` made `\b` Unicode-aware would add a boundary where upstream has none.
//
// The six rows below are where the two would disagree if they did. All six were driven through the
// installed rule and through Go's own regexp engine, and both agree on every one: a non-ASCII letter
// is a NON-word character to both, so a term adjacent to one is bounded and matches.
//
// The Cyrillic row is the control that makes the rest meaningful -- it is clean because the Te is a
// different letter from an ASCII T, not because of any boundary rule.
func TestNoWarningCommentsUsesAsciiWordBoundaries(t *testing.T) {
	t.Parallel()

	anywhere := NoWarningCommentsOptions{
		Terms: []string{"todo"}, Location: NoWarningCommentsAnywhere,
	}
	cases := []noWarningCommentsCase{
		// An accented letter is a non-word character, so the boundary holds and the term matches.
		{"// ÜTODO", anywhere, []string{"unexpectedComment"}, []string{"todo"}},
		{"// TODOé", anywhere, []string{"unexpectedComment"}, []string{"todo"}},
		{"// 日TODO", anywhere, []string{"unexpectedComment"}, []string{"todo"}},
		// An ASCII letter is a word character, so there is no boundary and nothing matches.
		{"// xTODO", anywhere, nil, nil},
		{"// TODOs", anywhere, nil, nil},
		{"// TODO_thing", anywhere, nil, nil},
		// The control: a Cyrillic Te is simply a different letter, so this is not the term at all.
		{"// ТODO", anywhere, nil, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoWarningComments(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The comment VALUE is the body without delimiters, and the difference is load-bearing.
//
// Upstream matches against `node.value`, which strips `//` or `/*` and `*/` and NOTHING else. Our
// comment reader hands back the full source text including delimiters, so they are removed in the
// rule, and getting that wrong in either direction moves a whole class of comment across the line:
// leaving the delimiters in would put `/` between the Start anchor and the term and silence every
// block comment, while stripping too much would make `/** TODO */` report.
//
// That last row is the one to read. `/** TODO */` is CLEAN by default because the surviving asterisk
// sits between the anchor and the term, and configuring that asterisk as a decoration character
// makes the same comment report. Both measured.
func TestNoWarningCommentsReadsTheCommentValue(t *testing.T) {
	t.Parallel()

	decorated := NoWarningCommentsOptions{Decoration: []string{"*"}}
	cases := []noWarningCommentsCase{
		{"// TODO: fix this", nil, []string{"unexpectedComment"}, []string{"todo"}},
		{"/* TODO: fix this */", nil, []string{"unexpectedComment"}, []string{"todo"}},
		// A JSDoc comment is clean by default: the asterisk blocks the anchor.
		{"/** TODO: fix this */", nil, nil, nil},
		// And reports once the asterisk is allowed as decoration.
		{"/** TODO: fix this */", decorated, []string{"unexpectedComment"}, []string{"todo"}},
		{"/**** TODO decorated */", decorated, []string{"unexpectedComment"}, []string{"todo"}},
		// Leading whitespace is always skipped under Start, with no decoration configured.
		{"//   TODO indented", nil, []string{"unexpectedComment"}, []string{"todo"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText+" "+fmtOptions(testCase.options), func(t *testing.T) {
			t.Parallel()
			result := runNoWarningComments(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// fmtOptions distinguishes rows whose source text repeats under different options.
func fmtOptions(options any) string {
	if options == nil {
		return "default"
	}
	encoded, err := json.Marshal(options)
	if err != nil {
		return "?"
	}
	return string(encoded)
}

// A directive comment configuring THIS rule is exempt, and both halves of that test matter.
//
// Upstream's guard is `isDirectiveComment(node) && /\bno-warning-comments\b/`, so a comment must be
// a directive AND name this rule. Dropping either half breaks a different case: without the name
// test, an `eslint-disable` for any other rule would be exempt; without the directive test, an
// ordinary comment discussing the rule would be.
func TestNoWarningCommentsExemptsItsOwnDirective(t *testing.T) {
	t.Parallel()

	cases := []noWarningCommentsCase{
		// A directive naming this rule: exempt even though it contains no term at all, which is
		// why the row below it is the one that proves the exemption does something.
		{"/* eslint no-warning-comments: [2, { terms: [\"todo\"] }] */", nil, nil, nil},
		{"/* eslint no-warning-comments: \"error\" */ // TODO", nil,
			[]string{"unexpectedComment"}, []string{"todo"}},
		// A directive for a DIFFERENT rule is not exempt from being read as a comment, though it
		// carries no term either.
		{"/* eslint no-console: \"error\" */", nil, nil, nil},
		// An ordinary comment mentioning the rule name is not a directive, so it is not exempt.
		{"// TODO stop using no-warning-comments here", nil,
			[]string{"unexpectedComment"}, []string{"todo"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoWarningComments(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The decoder, which enforces two constraints upstream states in its schema.
//
// We have no schema layer, so an unrecognized location would silently select a third behaviour of
// matching nothing, and a multi-character decoration entry would widen the character class in a way
// nobody wrote. Neither is reachable from a fixture that builds the options struct directly.
func TestDecodeNoWarningCommentsOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input yields the defaults", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoWarningCommentsOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options := decoded.(NoWarningCommentsOptions)
		if options.Terms != nil || options.Location != "" {
			t.Errorf("nil input produced %+v, wanted the zero value so the rule applies its "+
				"own defaults", options)
		}
	})

	for _, testCase := range []struct {
		raw  string
		want NoWarningCommentsLocation
	}{
		{`{"location":"start"}`, NoWarningCommentsStart},
		{`{"location":"anywhere"}`, NoWarningCommentsAnywhere},
	} {
		t.Run("upstream's "+testCase.raw+" is read", func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoWarningCommentsOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decoded.(NoWarningCommentsOptions).Location != testCase.want {
				t.Errorf("location = %q, wanted %q", decoded.(NoWarningCommentsOptions).Location, testCase.want)
			}
		})
	}

	// "Start" and "Anywhere" are the spellings cohere invented before #d21war2, which no ESLint
	// version accepts.
	for _, raw := range []string{
		`{"location":"Start"}`,
		`{"location":"Anywhere"}`,
		`{"location":"middle"}`,
		`{"location":""}`,
		`{"location":null}`,
	} {
		t.Run(raw+" is refused", func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeNoWarningCommentsOptions([]byte(raw)); err == nil {
				t.Errorf("%s decoded; upstream's schema refuses it", raw)
			}
		})
	}

	t.Run("a multi-character decoration is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeNoWarningCommentsOptions([]byte(`{"decoration":["**"]}`)); err == nil {
			t.Error("a two-character decoration decoded, which would widen the class")
		}
	})

	t.Run("a whitespace decoration is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeNoWarningCommentsOptions([]byte(`{"decoration":[" "]}`)); err == nil {
			t.Error("a whitespace decoration decoded")
		}
	})
}

// With nil options the rule uses upstream's default terms and the Start location.
//
// `options.(T)` on nil yields the zero value, whose nil Terms and empty Location match nothing, so
// the rule would report on no comment at all. Every fixture above reaches the rule through the
// decoder, so none of them can see that.
func TestNoWarningCommentsWithNilOptionsUsesTheDefaults(t *testing.T) {
	t.Parallel()

	for _, source := range []string{"// TODO x", "// FIXME x", "// XXX x"} {
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoWarningComments,
			noWarningCommentsFile, source), "unexpectedComment")
	}
	// The Start location is the default, so a term in the middle is clean.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoWarningComments, noWarningCommentsFile,
		"// something todo"))
	// And a term not in the default set is clean.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoWarningComments, noWarningCommentsFile,
		"// HACK x"))
}
