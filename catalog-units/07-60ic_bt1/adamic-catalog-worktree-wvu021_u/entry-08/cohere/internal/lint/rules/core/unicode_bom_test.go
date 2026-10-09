package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// unicodeBomFile is where the fixtures pretend to live.
const unicodeBomFile = "/repository/source/UnicodeBom.ts"

// decodedUnicodeBomOptions routes a fixture's options through the rule's own exported decoder
// rather than building the options struct directly.
//
// This is the whole reason the decoder is testable at all. The rule's default is `never`, so a
// struct built by hand in a fixture would carry a non-nil Require and every case would pass while
// the live config's bare `"error"` still handed the rule a nil the rule read as silence. An empty
// string here means exactly that bare configuration.
func decodedUnicodeBomOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	options, err := DecodeUnicodeBomOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return options
}

// The corpus is ESLint's own, at `tests/lib/rules/unicode-bom.js`, copied rather than rewritten:
// 3 clean cases and 4 reporting ones, which is the whole upstream file. Each reporting case carries
// its own `output`, so the rewrite asserted below is upstream's data rather than a guess.
//
// Every case string was built from the code point rather than typed, so nothing could cook the mark
// into a different character on the way in, and each verdict was reproduced by driving the installed
// eslint at 10.8.1 before being written here.
//
// The option spelling is the BARE string upstream writes second in its `["error", "always"]` tuple.
// cohere's config layer strips the tuple, so what a decoder receives is `"always"` rather than
// `["always"]`, and a fixture copying ESLint's array spelling would fail on every row.
func TestUnicodeBomFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
		wantId     string
		wantFixed  string
	}{
		{"an unmarked file under always", "var a = 123;", "\"always\"", "expected", "\ufeffvar a = 123;"},
		{"an unmarked file opening with a comment under always", " // here's a comment \nvar a = 123;", "\"always\"", "expected", "\ufeff // here's a comment \nvar a = 123;"},
		// Upstream writes this case with no options at all, which is the default-`never` path and
		// the one the live config produces. It is the case a zero-valued options struct gets wrong.
		{"a marked file with no options at all", "\ufeff var a = 123;", "", "unexpected", " var a = 123;"},
		{"a marked file under never", "\ufeff var a = 123;", "\"never\"", "unexpected", " var a = 123;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, UnicodeBom, unicodeBomFile, testCase.sourceText,
				decodedUnicodeBomOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

func TestUnicodeBomStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
	}{
		{"a marked file under always", "\ufeff var a = 123;", "\"always\""},
		{"an unmarked file under never", "var a = 123;", "\"never\""},
		// The mark only counts at position zero. Anywhere else it is an ordinary zero-width
		// no-break space, which is why the rule tests a prefix rather than searching.
		{"a mark at the end under never", "var a = 123; \ufeff", "\"never\""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, UnicodeBom, unicodeBomFile, testCase.sourceText,
				decodedUnicodeBomOptions(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestUnicodeBomReportsAtPositionZero pins where the finding points.
//
// Upstream overrides the location explicitly, to line 1 column 0, for both arms. Nothing in a
// message-id fixture can see this, and getting it wrong is easy in a specific way: the source file
// NODE's own range begins AFTER the mark, measured at position 3 on a marked file, so reporting the
// node would point at the first real token in exactly the case where the mark is what is being
// complained about.
func TestUnicodeBomReportsAtPositionZero(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		options    string
	}{
		{"the absent mark under always", "var a = 123;", "\"always\""},
		{"the present mark under never", "\ufeff var a = 123;", "\"never\""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, UnicodeBom, unicodeBomFile, testCase.sourceText,
				decodedUnicodeBomOptions(t, testCase.options))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0].Range
			if reported.Pos() != 0 || reported.End() != 0 {
				t.Fatalf("expected the empty range at position zero, got [%d,%d)", reported.Pos(), reported.End())
			}
		})
	}
}

// TestUnicodeBomMessages asserts the rendered text of both messages.
//
// Neither message interpolates anything, so this is an equality check on two constants rather than
// a guard against a format string. It is here because the ids alone are two words that a reword
// could quietly detach from what they describe, and because a rule whose two arms are near-mirrors
// is the shape where one message ends up reported by both arms.
func TestUnicodeBomMessages(t *testing.T) {
	t.Parallel()

	if messageUnicodeBomExpected.Id != "expected" {
		t.Fatalf("expected id %q, got %q", "expected", messageUnicodeBomExpected.Id)
	}
	if messageUnicodeBomUnexpected.Id != "unexpected" {
		t.Fatalf("expected id %q, got %q", "unexpected", messageUnicodeBomUnexpected.Id)
	}
	if !strings.HasPrefix(messageUnicodeBomExpected.Description, "This file is configured to start with a byte order mark and does not.") {
		t.Fatalf("the expected-arm description was reworded: %q", messageUnicodeBomExpected.Description)
	}
	if !strings.HasPrefix(messageUnicodeBomUnexpected.Description, "This file starts with a byte order mark.") {
		t.Fatalf("the unexpected-arm description was reworded: %q", messageUnicodeBomUnexpected.Description)
	}
}

// TestDecodeUnicodeBomOptions is the test the previous attempt at this rule was writing when it ran
// out of time, and it is where this rule is hard.
//
// Three separate things are asserted and each is a way the rule inverts silently:
//
// Empty input must yield `never` rather than a zero value. This is the live config's bare `"error"`
// and `rule.DecodeOptionsInto` would error here, which the config layer turns into nil.
//
// A nil handed straight to the rule must also yield `never`, which is the same defect arriving one
// layer later, past the decoder rather than through it.
//
// An unrecognised string must be an error rather than a quiet fallback, because falling back would
// enforce the opposite of what a misspelled config asked for and say nothing.
func TestDecodeUnicodeBomOptions(t *testing.T) {
	t.Parallel()

	t.Run("empty input defaults to never", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeUnicodeBomOptions(nil)
		if err != nil {
			t.Fatalf("empty input should decode, got %v", err)
		}
		options, isOptions := decoded.(UnicodeBomOptions)
		if !isOptions {
			t.Fatalf("expected UnicodeBomOptions, got %T", decoded)
		}
		if options.Require == nil || *options.Require != UnicodeBomNever {
			t.Fatalf("expected never, got %v", options.Require)
		}
	})

	t.Run("always decodes", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeUnicodeBomOptions(json.RawMessage(`"always"`))
		if err != nil {
			t.Fatalf("expected always to decode, got %v", err)
		}
		options := decoded.(UnicodeBomOptions)
		if options.Require == nil || *options.Require != UnicodeBomAlways {
			t.Fatalf("expected always, got %v", options.Require)
		}
	})

	t.Run("an unknown setting is an error naming the value", func(t *testing.T) {
		t.Parallel()
		_, err := DecodeUnicodeBomOptions(json.RawMessage(`"alway"`))
		if err == nil {
			t.Fatalf("expected an error for an unrecognised setting")
		}
		if !strings.Contains(err.Error(), "alway") {
			t.Fatalf("the error should name the bad value, got %q", err.Error())
		}
	})

	t.Run("the wrong wire shape is an error", func(t *testing.T) {
		t.Parallel()
		// Upstream's own spelling is `["always"]`. cohere strips the tuple, so an array arriving
		// here means the config layer's contract changed, and that should fail loudly rather than
		// quietly enforcing the default.
		if _, err := DecodeUnicodeBomOptions(json.RawMessage(`["always"]`)); err == nil {
			t.Fatalf("expected an array to be refused")
		}
	})

	t.Run("nil options reach the rule as never", func(t *testing.T) {
		t.Parallel()
		// Past the decoder rather than through it: this is what a bare `"error"` produces after the
		// config layer has turned the decoder's error into nil.
		result := rule_testing.RunWithOptions(t, UnicodeBom, unicodeBomFile, "\ufeff var a = 123;", nil)
		rule_testing.ExpectFindings(t, result, "unexpected")
	})
}

// TestUnicodeBomRemovesOneMarkPerPass pins the doubled-mark behaviour.
//
// A file beginning with the mark twice reports ONCE and this rule's single fix removes one of them.
// Measured against the installed build: `linter.cohere` returns exactly one message carrying one
// fix, and `cohereAndFix` strips the second only on a later pass of its own re-lint loop. A port
// removing both in one edit would be doing something upstream's rule never does, and no message-id
// fixture could see the difference.
func TestUnicodeBomRemovesOneMarkPerPass(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, UnicodeBom, unicodeBomFile, "\ufeff\ufeffvar a = 1;",
		decodedUnicodeBomOptions(t, "\"never\""))
	rule_testing.ExpectFindings(t, result, "unexpected")
	rule_testing.ExpectFixedSource(t, result, "\ufeffvar a = 1;")
}

// TestUnicodeBomOnAnEmptyFile covers a case upstream's corpus does not write and its rule does
// answer.
//
// Measured against the installed build: an empty file under `always` reports and its fix produces a
// file holding nothing but the mark, and under `never` it is clean. Worth pinning because an empty
// file is the one input where a prefix test and a length test could plausibly disagree.
func TestUnicodeBomOnAnEmptyFile(t *testing.T) {
	t.Parallel()

	t.Run("always", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, UnicodeBom, unicodeBomFile, "",
			decodedUnicodeBomOptions(t, "\"always\""))
		rule_testing.ExpectFindings(t, result, "expected")
		rule_testing.ExpectFixedSource(t, result, "\ufeff")
	})

	t.Run("never", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(t, UnicodeBom, unicodeBomFile, "",
			decodedUnicodeBomOptions(t, "\"never\""))
		rule_testing.ExpectClean(t, result)
	})
}

// TestUnicodeBomCannotSeeAMarkThroughTheRealReadPath is the test that says why this rule is
// registered and not enabled, and it is the only test in this file that is about cohere rather than
// about the rule.
//
// Every other fixture here parses a Go string, which is not the path a real run takes. A real run
// reads each file through `osvfs`, and that read removes a leading byte order mark, so the rule's
// subject is gone before any listener fires. This asserts the stripping directly, with two controls
// so that a "no mark" reading cannot be an artifact of the probe: an unmarked file of the same
// length must come back unchanged, and a mark written in the middle of a file must survive.
//
// If this test starts failing because the mark now survives the read, that is the signal that the
// rule can be enabled: run `EnableRule.ts 'unicode-bom' --layer universal`, drop the entry from
// `deliberatelyNotEnabled`, and update the rule's doc comment. Nothing else about the port changes.
func TestUnicodeBomCannotSeeAMarkThroughTheRealReadPath(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()

	marked := filepath.Join(directory, "Marked.ts")
	unmarked := filepath.Join(directory, "Unmarked.ts")
	middle := filepath.Join(directory, "Middle.ts")

	markBytes := []byte{0xef, 0xbb, 0xbf}
	if err := os.WriteFile(marked, append(append([]byte{}, markBytes...), []byte("export const a = 1;\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unmarked, []byte("export const b = 2;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(middle, append([]byte("export const c = 3;"), append(append([]byte{}, markBytes...), '\n')...), 0o644); err != nil {
		t.Fatal(err)
	}

	read := func(path string) string {
		t.Helper()
		text, wasRead := osvfs.FS().ReadFile(tspath.RootedFilePath(path))
		if !wasRead {
			t.Fatalf("could not read %s", path)
		}
		return text
	}

	if strings.HasPrefix(read(marked), unicodeBomMark) {
		t.Fatalf("the leading mark now survives the real read path, so unicode-bom can be enabled; " +
			"see the doc comment on the rule for what to change")
	}

	// The controls. Without these, the assertion above is satisfied by any broken read at all.
	if got := read(unmarked); got != "export const b = 2;\n" {
		t.Fatalf("the unmarked control came back changed, so the probe above is not measuring what it claims: %q", got)
	}
	if got := read(middle); !strings.Contains(got, unicodeBomMark) {
		t.Fatalf("a mark in the MIDDLE of a file was stripped too, so the removal is not specific to "+
			"position zero and the reasoning recorded on the rule is wrong: %q", got)
	}
}
