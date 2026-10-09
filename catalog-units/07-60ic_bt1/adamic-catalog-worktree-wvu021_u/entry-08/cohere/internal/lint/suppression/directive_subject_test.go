package suppression

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/ecmascript/directives"
)

// directiveSubjectProbeRule stands in for the real rule, which lives in a package this one must not
// import. Registered from init because the map is written only before any test runs, exactly as in
// production.
const directiveSubjectProbeRule = "probe/reports-on-directives"

func init() {
	directives.RegisterSubject(directiveSubjectProbeRule)
}

// TestADirectiveCannotSilenceTheFindingAboutItself is the ESLint table on coversAFindingAboutADirective,
// row for row. Each finding sits at the first byte of the comment it is about, which is where the
// rule anchors it.
//
// The control is the same index asked the same question for an ordinary rule name: every row that
// says a directive-subject finding escapes also asserts that an ordinary finding at the same offset
// is suppressed, so a Suppresses that had simply stopped working would fail here rather than pass.
func TestADirectiveCannotSilenceTheFindingAboutItself(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		source      string
		line        int
		column      int
		wantSilence bool
	}{
		{"blanket same-line", "x; // eslint-disable-line", 0, 3, false},
		{"same-line naming the rule", "x; // eslint-disable-line probe/ordinary-rule, " + directiveSubjectProbeRule, 0, 3, false},
		{"next-line from the line above", "// eslint-disable-next-line -- why\n// eslint-disable-line", 1, 0, false},
		{"block disable on an earlier line", "/* eslint-disable -- why */\n// eslint-disable-line", 1, 0, true},
		{"block disable earlier on the same line", "/* eslint-disable -- why */ /* eslint-disable-line */", 0, 28, false},
		{"a bare block disable is not covered by itself", "/* eslint-disable */", 0, 0, false},
		{"an enable closing an earlier block is still inside it", "/* eslint-disable */\n/* eslint-enable */", 1, 0, true},
		{"a line after the enable is outside the block", "/* eslint-disable -- why */\n/* eslint-enable */\n// eslint-disable-line", 2, 0, false},
		{"cohere spelling behaves the same", "x; // cohere-disable-line", 0, 3, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			index := Build(testCase.source)
			offset := offsetOfLine(testCase.source, testCase.line) + testCase.column
			if !strings.HasPrefix(testCase.source[offset:], "/") {
				t.Fatalf("offset %d does not sit on a comment in %q", offset, testCase.source)
			}

			got := index.Suppresses(directiveSubjectProbeRule, offset)
			if got != testCase.wantSilence {
				t.Fatalf("directive-subject finding at line %d: suppressed=%v, want %v", testCase.line, got, testCase.wantSilence)
			}

			// The control: an ordinary rule at the same offset is covered by whatever covers that line.
			// Every case here has a blanket directive reaching the line, so it must be silenced.
			if !Build(testCase.source).Suppresses("probe/ordinary-rule", offset) {
				t.Fatalf("control: an ordinary rule at the same offset was not suppressed, so this case proves nothing")
			}
		})
	}
}

// TestRecognizeAgreesWithBuild pins directives.Recognize to the index Build produces, in every
// honored spelling, and refuses the near misses that Build refuses.
func TestRecognizeAgreesWithBuild(t *testing.T) {
	t.Parallel()

	cases := []struct {
		comment  string
		wantWord string
	}{
		{"// eslint-disable-next-line foo", directives.ScopeWordDisableNextLine},
		{"// cohere-disable-next-line foo -- why", directives.ScopeWordDisableNextLine},
		{"/* verify-disable-line */", directives.ScopeWordDisableLine},
		{"/* oxlint-disable */", directives.ScopeWordDisable},
		{"/* eslint-disable */", directives.ScopeWordDisable},
		{"/* cohere-enable */", directives.ScopeWordEnable},
		{"/* eslint-enable foo */", directives.ScopeWordEnable},
		// A file-scope disable and an enable are block-comment forms only, ESLint's grammar.
		{"// oxlint-disable", ""},
		{"// eslint-enable foo", ""},
		// A scope word ends at whitespace or the end of the comment, as in ESLint.
		{"// eslint-disable-next-line, no-console", ""},
		{"// eslint-disable-next-lineno-console", ""},
		{"// eslint-disable-lines no-console", ""},
		{"// eslint-disable-line\tno-console", directives.ScopeWordDisableLine},
		{"/*\n * cohere-disable foo\n */", directives.ScopeWordDisable},
		{"// cohere-disabled", ""},
		{"// cohere-enabled", ""},
		{"// mentions cohere-disable in prose", ""},
		{"/** eslint-disable */", ""},
		{"// tslint:disable", ""},
	}
	for _, testCase := range cases {
		word, honored := directives.Recognize(testCase.comment)
		if word != testCase.wantWord || honored != (testCase.wantWord != "") {
			t.Errorf("Recognize(%q) = %q, %v; want %q", testCase.comment, word, honored, testCase.wantWord)
		}

		// Agreement with the index: a recognized disable produces a directive, and a refused comment
		// produces none. Enables are checked through a block they close.
		built := Build(testCase.comment + "\n").Directives()
		isDisable := testCase.wantWord != "" && testCase.wantWord != directives.ScopeWordEnable
		if isDisable != (len(built) == 1) {
			t.Errorf("Build(%q) found %d directives, but Recognize said %q", testCase.comment, len(built), testCase.wantWord)
		}
		if testCase.wantWord == directives.ScopeWordEnable {
			closed := Build("/* eslint-disable */\n" + testCase.comment + "\nx\n").Directives()
			if len(closed) != 1 || closed[0].EndLine != 1 {
				t.Errorf("%q was recognized as an enable but did not close a block", testCase.comment)
			}
		}
	}
}
