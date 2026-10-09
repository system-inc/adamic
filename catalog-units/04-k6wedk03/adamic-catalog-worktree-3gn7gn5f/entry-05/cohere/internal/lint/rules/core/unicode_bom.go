package core

import (
	"encoding/json"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnicodeBomExpected = rule.Message{
	Id: "expected",
	Description: "This file is configured to start with a byte order mark and does not. The mark " +
		"is what tells a reader that has no other signal which encoding and endianness the bytes " +
		"are in, so a project that requires it wants every file to carry it rather than most.",
}

var messageUnicodeBomUnexpected = rule.Message{
	Id: "unexpected",
	Description: "This file starts with a byte order mark. In a project that is already UTF-8 " +
		"everywhere the mark carries no information and is invisible in an editor, while tools " +
		"that read the first bytes literally see it: a shell script stops being executable, a " +
		"concatenated bundle gains a stray character mid-file, and a diff shows a change on a " +
		"line nobody edited.",
}

// unicodeBomMark is the byte order mark as it appears in a UTF-8 source file.
//
// Three bytes, `ef bb bf`, rather than the single code point `U+FEFF` an ESLint rule reasons about.
// This is the whole difference between the two ports: ESLint reads its source as a JavaScript
// string with the mark already stripped into a `hasBOM` flag, so its fixer can write the sentinel
// range `[-1, 0)` and let the fix applier interpret it. cohere's source text is the file's bytes,
// measured: `SourceFile.Text()` on a marked file begins `ef bb bf` and the source file NODE's own
// range starts after them, at position 3. So the mark is ordinary text here and the repair is an
// ordinary edit over `[0, 3)`.
const unicodeBomMark = "\ufeff"

// UnicodeBomSetting is which of the two things this rule enforces.
//
// A string rather than a bool because upstream's option is a string enum and the two spellings are
// what a config author writes. Keeping the name means a misconfiguration reads as a misconfiguration
// rather than as a silently-false flag.
type UnicodeBomSetting string

const (
	// UnicodeBomAlways requires every file to begin with the mark.
	UnicodeBomAlways UnicodeBomSetting = "always"

	// UnicodeBomNever forbids it, which is upstream's default.
	UnicodeBomNever UnicodeBomSetting = "never"
)

// UnicodeBomOptions carries upstream's single positional option.
//
// The pointer is the point. Upstream's `defaultOptions: ["never"]` means an ABSENT option forbids
// the mark, so the zero value of a plain `UnicodeBomSetting` field is the empty string, which
// matches neither arm and makes the rule silent on every file it is meant to catch. A rule
// configured as bare `"error"` in the live config is exactly that case, and it is the common one.
// Keeping the wire value nullable is what lets `DefaultUnicodeBomSettings` fill it in rather than
// having an inversion decided by Go's zero value.
type UnicodeBomOptions struct {
	Require *UnicodeBomSetting
}

// DefaultUnicodeBomSettings is what an unconfigured rule enforces: upstream's `never`.
func DefaultUnicodeBomSettings() UnicodeBomOptions {
	never := UnicodeBomNever
	return UnicodeBomOptions{Require: &never}
}

// DecodeUnicodeBomOptions turns the configured value into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` for two reasons that both matter here and only
// the first is the usual one.
//
// The generic helper errors on empty input, the config layer turns that error into nil for a rule
// whose options are not required, and the rule then reads nil as a zero value. For an option
// defaulting to FALSE that is harmless; for this one it is an inversion, because the default is
// `never` and the zero value is the empty string, which reports nothing at all.
//
// The second reason is the wire shape. Upstream's option is a bare STRING at the top of its schema
// array, so a config writes `["error", "always"]` and cohere's config layer strips the tuple before
// dispatch: what arrives here is the JSON `"always"`, not `["always"]` and not `{"require":...}`.
// A decoder built by naming a struct field could not read that at all.
//
// An unrecognised string is an error rather than a silent fallback to the default. A config saying
// `"alway"` is a mistake somebody made, and turning it into `never` would enforce the opposite of
// what they wrote with nothing anywhere saying so.
func DecodeUnicodeBomOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultUnicodeBomSettings(), nil
	}

	var configured string
	if err := json.Unmarshal(raw, &configured); err != nil {
		return DefaultUnicodeBomSettings(), err
	}

	setting := UnicodeBomSetting(configured)
	switch setting {
	case UnicodeBomAlways, UnicodeBomNever:
		return UnicodeBomOptions{Require: &setting}, nil
	}
	return DefaultUnicodeBomSettings(), &unicodeBomUnknownSetting{configured: configured}
}

// unicodeBomUnknownSetting names the bad value rather than only saying the shape was wrong.
type unicodeBomUnknownSetting struct {
	configured string
}

func (e *unicodeBomUnknownSetting) Error() string {
	// Go whitespace: the bad option value echoed in cohere's own error message, which decides nothing.
	return "unicode-bom takes \"always\" or \"never\", got " + strings.TrimSpace(e.configured)
}

// UnicodeBom requires or forbids a byte order mark at the start of a file.
//
//	valid (always):   <mark> var a = 123;
//	valid (never):    var a = 123;
//	valid (never):    var a = 123; <mark>   the mark only counts at position zero
//	invalid (always): var a = 123;
//	invalid (never):  <mark> var a = 123;
//
// # The default is `never` and getting that wrong makes the rule silent rather than loud
//
// Upstream declares `defaultOptions: ["never"]`, so a rule configured as a bare severity forbids the
// mark. Measured against the installed build at 10.8.1: `unicode-bom: "error"` on a marked file
// reports `unexpected`, identically to the explicit `["error", "never"]`. That is the case the live
// config produces and the case a zero-valued options struct would get wrong, which is why the
// decoder above is hand-rolled and why a fixture routes nil through it.
//
// # Where the finding points
//
// Upstream reports with an explicit `loc: {line: 1, column: 0}` and no end, which is line 1 column
// 1 in the one-based numbering a reader sees, and it does that for BOTH arms including the one where
// the mark is absent. Reporting the source file NODE would be wrong in a way no message-id fixture
// could see: the node's range begins AFTER the mark, measured at position 3 for a marked `var a = 1;`,
// so a marked file would report at the first real token rather than at the mark being complained
// about. The empty range at position zero is used instead, which is where upstream points and the
// only span that is inside the mark rather than after it.
//
// # The repair, and why one pass removes one mark
//
// Both arms are fixable and upstream declares `fixable: "whitespace"`. Adding the mark is an insert
// at position zero; removing it is a delete of the three bytes there.
//
// A file beginning with the mark TWICE reports once and this rule's single fix removes one of them,
// leaving a file that still begins with a mark. That is not a defect and it matches upstream:
// measured, `linter.cohere` on a doubled mark returns exactly one message carrying one fix, and it
// is `cohereAndFix`'s re-lint loop, not the rule, that strips the second on a later pass. A port
// removing both in one edit would be doing something upstream's rule never does.
//
// # This rule is registered and NOT enabled, because it cannot see its subject here
//
// Everything above is a faithful port and every fixture below it passes, and the rule still reports
// nothing on a real file no matter what that file contains. The reason is one layer down and it is
// measured rather than argued.
//
// Every source file in a real run is read through `osvfs.FS().ReadFile`, and that read STRIPS a
// leading byte order mark. Probed directly: a file whose first bytes on disk are `ef bb bf` arrives
// as text beginning `65 78 70`, three bytes shorter, while an unmarked control of the same length
// comes back unchanged and a mark written in the MIDDLE of a file survives. The removal is specific
// to position zero, which is the only position this rule asks about. `cachedvfs` wrapping the same
// reads answers identically, so it is not the caching layer.
//
// The fixtures do not see this because `rule_testing.Run` parses a Go string and never touches the
// filesystem, so `SourceFile.Text()` there holds whatever the fixture wrote, mark included. That is
// the gap between a green suite and a working rule, and it is why the seeded probe tree matters: a
// two-file tree holding one genuinely marked file was offered to this rule, registered a listener,
// and reported zero.
//
// So the rule stays registered and out of the config, with the reason recorded in
// `deliberatelyNotEnabled`. What would make it work is small and belongs elsewhere: a lint phase
// reading the file's real bytes, or a flag on the source file saying a mark was stripped. Worth
// knowing while that is decided: the FIX phase reads through `os.ReadFile` and KEEPS the mark, so
// on a marked file the two phases disagree by three bytes about where every offset in that file
// lives. That is a hazard for any fixable rule, not only this one.
//
// # Only the first position counts
//
// A mark written at the END of a file is clean under `never`, which is upstream's own third valid case. A mark
// anywhere but position zero is an ordinary zero-width no-break space character, so the test is a
// prefix test rather than a search.
var UnicodeBom = rule.Rule{
	Name: "unicode-bom",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare severity is handed nil, which the type assertion turns into a
		// zero-valued struct whose Require is nil. That is the live config's shape, so the fallback
		// below is the ordinary path rather than a defensive one.
		settings, _ := rule.OptionsAs[UnicodeBomOptions](options)
		if settings.Require == nil {
			settings = DefaultUnicodeBomSettings()
		}
		require := *settings.Require

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				hasMark := strings.HasPrefix(ctx.SourceFile.Text(), unicodeBomMark)

				// The empty range at position zero, which is upstream's `{line: 1, column: 0}` and
				// is inside the mark rather than after it.
				start := core.NewTextRange(0, 0)

				switch {
				case require == UnicodeBomAlways && !hasMark:
					ctx.ReportRangeWithFixes(start, messageUnicodeBomExpected,
						rule.ReplaceRange(start, unicodeBomMark))
				case require == UnicodeBomNever && hasMark:
					markRange := core.NewTextRange(0, len(unicodeBomMark))
					ctx.ReportRangeWithFixes(start, messageUnicodeBomUnexpected,
						rule.RemoveRange(markRange))
				}
			},
		}
	},
}
