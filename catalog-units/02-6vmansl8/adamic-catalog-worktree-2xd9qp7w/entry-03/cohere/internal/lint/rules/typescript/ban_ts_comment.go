package typescript

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageBanTsCommentBanned is upstream's `tsDirectiveComment`, the outright refusal of a directive.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer, so upstream's `{{ directive }}`
// becomes concatenation at the report site. The directive name is worth carrying: a file holding a
// `@ts-nocheck` and a `@ts-check` produces two findings under the same id, and the name is the only
// thing in the text that separates them.
func messageBanTsCommentBanned(directive string) rule.Message {
	return rule.Message{
		Id: "tsDirectiveComment",
		Description: "This file silences the TypeScript compiler with `@ts-" + directive +
			"`. A directive comment switches off checking for a line or a whole file, so the error " +
			"it hides stays in the code and stops being reported to anyone. Fix the underlying " +
			"type error instead, or narrow the suppression to `@ts-expect-error` with a " +
			"description saying why it is necessary.",
	}
}

// messageBanTsCommentPreferExpectError is upstream's `tsIgnoreInsteadOfExpectError`.
//
// The only arm that offers a repair, and the only one whose text names a replacement rather than a
// removal, because the two directives differ in a way that matters: `@ts-expect-error` becomes an
// error itself once the line stops failing, and `@ts-ignore` silently rots.
func messageBanTsCommentPreferExpectError() rule.Message {
	return rule.Message{
		Id: "tsIgnoreInsteadOfExpectError",
		Description: "This suppression uses `@ts-ignore`, which keeps silencing the line even " +
			"after the underlying error is gone, so a stale suppression is invisible forever. " +
			"`@ts-expect-error` reports when the line it guards has stopped failing, which is what " +
			"makes the suppression removable.",
	}
}

// messageBanTsCommentReplaceIgnore is upstream's `replaceTsIgnoreWithTsExpectError`, the suggestion.
func messageBanTsCommentReplaceIgnore() rule.Message {
	return rule.Message{
		Id:          "replaceTsIgnoreWithTsExpectError",
		Description: "Replace `@ts-ignore` with `@ts-expect-error`.",
	}
}

// messageBanTsCommentRequiresDescription is upstream's `tsDirectiveCommentRequiresDescription`.
//
// The minimum length is rendered into the text because it is configurable, so a reader seeing the
// finding with no number would have to open the config to learn what would satisfy it.
func messageBanTsCommentRequiresDescription(directive string, minimumLength int) rule.Message {
	return rule.Message{
		Id: "tsDirectiveCommentRequiresDescription",
		Description: "This `@ts-" + directive + "` carries no explanation, so nobody reading it " +
			"later can tell whether the suppression is still needed or what it was hiding. Write " +
			"at least " + strconv.Itoa(minimumLength) + " characters after the directive saying which " +
			"error it suppresses and why the error cannot be fixed.",
	}
}

// messageBanTsCommentDescriptionFormat is upstream's `tsDirectiveCommentDescriptionNotMatchPattern`.
//
// The pattern is rendered because it is the only actionable part: a reader told their description
// is wrong without being told the shape has nothing to act on.
func messageBanTsCommentDescriptionFormat(directive string, pattern string) rule.Message {
	return rule.Message{
		Id: "tsDirectiveCommentDescriptionNotMatchPattern",
		Description: "The explanation after this `@ts-" + directive + "` does not match the shape " +
			"this project requires, `" + pattern + "`. The format exists so a suppression can be " +
			"traced back to the error it hides, which a free-form note cannot be.",
	}
}

// BanTsComment flags the TypeScript directive comments that switch off type checking.
//
//	valid:   // just a comment containing @ts-expect-error somewhere
//	valid:   // @ts-expect-error here is why the error is expected
//	valid:   const a = 1;\n// @ts-nocheck           too late to apply, so not a directive
//	valid:   /* @ts-ignore\n * not the last line */
//	invalid: // @ts-ignore
//	invalid: // @ts-expect-error
//	invalid: /// @ts-nocheck
//
// A port of typescript-eslint 8.71.0's `@typescript-eslint/ban-ts-comment`, the rule's defining
// implementation. It was first ported from oxc's reimplementation, which differs from it in the seven
// ways below, each now settled by upstream and pinned by its corpus (ban_ts_comment_corpus_data_test.go:
// all 114 of upstream's rows, plus 40 edge rows, replayed against the installed rule):
//
//   - The length check and the format check are one or the other: a description too short is not
//     also judged against the format. oxc ran both and could report one comment twice.
//   - A description's length is counted in characters as a reader sees them (getStringLength:
//     `Intl.Segmenter` graphemes for anything not ASCII, here `text.GraphemeCount`). oxc counted
//     bytes, so two accented letters passed a minimum of three.
//   - The description is trimmed as JavaScript trims, which keeps U+0085 and strips U+FEFF; Go's
//     `strings.TrimSpace` does the reverse for both.
//   - A finding covers the whole comment, delimiters included, where oxc's covered its interior.
//   - Replacing `@ts-ignore` is a suggestion, not a fix, and it rewrites the first occurrence only:
//     upstream's `comment.value.replace(/@ts-ignore/, ...)` has no `g` flag.
//   - `@ts-check` and `@ts-nocheck` are read only from a line comment, through TypeScript's own
//     pragma pattern, which accepts three slashes. oxc exempted `///` instead.
//   - A `@ts-nocheck` on or after the line the first statement starts on is silent, because
//     TypeScript only honors one before any code.
//
// # A JavaScript file is still declined, as a stand-in for the config
//
// oxc's eighth difference is kept for now, recorded rather than settled. Upstream's rule has no
// file-type gate, so which files it sees is the config's decision, and an ESLint config commonly
// scopes typescript-eslint's rules to TypeScript files. cohere's settings can do the same, through an
// override's `files` globs, so the scope belongs in cohere:adamic rather than here. What stops it
// moving today is readiness: it resolves cohere:adamic once, for `index.ts`, so a file-scoped
// override in the set would not reach it. Measured on TanStack Query: without the gate,
// `// @ts-nocheck` atop scripts/create-github-release.mjs reports, where its own ESLint config does
// not apply the rule to that file at all. #6aa3wrx makes readiness resolve the set per file, moves
// the scope into the set, and removes this gate.
//
// # The three patterns, read without a regular expression engine
//
// Upstream matches each comment against one of three patterns, taken from TypeScript's scanner and
// parser. Each is a run of one character class, then another, then a fixed `@ts-` directive, then
// the rest of the line, and no two adjacent classes share a character, so a left-to-right scan takes
// the same path the backtracking engine does. `\s` is JavaScript's, which is exactly
// `type_checking.IsStrWhiteSpace`, and `.` stops at a line terminator:
//
//	line comment, `//` + value    ^\/\/\/?\s*@ts-(check|nocheck)(.*)$
//	line comment, value           ^\/*\s*@ts-(expect-error|ignore)(.*)
//	block comment, last line      ^\s*(?:\/|\*)*\s*@ts-(expect-error|ignore)(.*)
//
// The first is tried first, and only on a line comment. A block comment's value is split into lines
// and only the last is read, so a directive above the closing line is invisible.
var BanTsComment = rule.Rule{
	Name: "@typescript-eslint/ban-ts-comment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}
		// See "A JavaScript file is still declined" above.
		if !isTypeScriptSourceFile(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		// The zero value is not the default and this is the normal path, not a defensive branch. A
		// rule configured as a bare `"error"` is handed nil options, because `DecodeOptionsInto`
		// errors on empty input and the config layer turns that into nil. A bare type assertion
		// would then yield four zero-valued directive settings, which match no arm, and the rule
		// would register on every file and report nothing while every fixture stayed green.
		parsed, decoded := rule.OptionsAs[BanTsCommentOptions](options)
		if !decoded {
			parsed = DefaultBanTsCommentOptions()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				reportBanTsComments(ctx, parsed)
			},
		}
	},
}

// reportBanTsComments is upstream's `Program` listener.
func reportBanTsComments(ctx rule.Context, options BanTsCommentOptions) {
	// Upstream compares the line the first statement's first token is on with the comment's own,
	// so a `@ts-nocheck` sharing the first statement's line is as late as one below it.
	firstStatementLine := -1
	if statements := ctx.SourceFile.Statements.Nodes; len(statements) > 0 {
		start := scanner.GetTokenPosOfNode(statements[0], ctx.SourceFile, false)
		firstStatementLine, _ = scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, start)
	}

	for _, comment := range comments.ForFile(ctx) {
		value, readable := banTsCommentValue(comment)
		if !readable {
			continue
		}
		directive, description, found := findBanTsCommentDirective(value, comment.IsBlock)
		if !found {
			continue
		}

		if directive == banTsCommentNoCheck && firstStatementLine >= 0 {
			commentLine, _ := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, comment.Range.Pos())
			if firstStatementLine <= commentLine {
				continue
			}
		}

		setting := options.settingFor(directive)
		switch setting.Kind {
		case BanTsCommentBoolean:
			if !setting.Banned {
				continue
			}
			if directive == banTsCommentIgnore {
				replaced := strings.Replace(value, "@ts-ignore", "@ts-expect-error", 1)
				repaired := "//" + replaced
				if comment.IsBlock {
					repaired = "/*" + replaced + "*/"
				}
				ctx.ReportRangeWithSuggestions(comment.Range, messageBanTsCommentPreferExpectError(),
					rule.Suggestion{
						Message: messageBanTsCommentReplaceIgnore(),
						Fixes:   []rule.Fix{{Range: comment.Range, Text: repaired}},
					})
				continue
			}
			ctx.ReportRange(comment.Range, messageBanTsCommentBanned(directive))

		case BanTsCommentRequireDescription, BanTsCommentDescriptionFormat:
			// An object setting is only a description setting when its format is set and non-empty:
			// upstream tests `option.descriptionFormat` for truth, so `{}`, a null format and an
			// empty one check nothing at all.
			if setting.Kind == BanTsCommentDescriptionFormat && setting.DescriptionFormat == nil {
				continue
			}
			if banTsCommentLength(strings.TrimFunc(description, type_checking.IsStrWhiteSpace)) <
				options.MinimumDescriptionLength {
				ctx.ReportRange(comment.Range, messageBanTsCommentRequiresDescription(
					directive, options.MinimumDescriptionLength))
				continue
			}
			// The format is matched against the description as written, untrimmed, which is what
			// makes `@ts-ignore    : TS1234` fail a pattern beginning `^:`. A match that overruns
			// esregexp's time bound counts as a match, so a pathological pattern is silent rather
			// than reporting every description.
			if setting.DescriptionFormat != nil && !setting.DescriptionFormat.TestOrTimeout(description) {
				ctx.ReportRange(comment.Range, messageBanTsCommentDescriptionFormat(
					directive, setting.DescriptionFormat.Source()))
			}
		}
	}
}

// banTsCommentLength is upstream's `getStringLength`: the byte length of ASCII, the grapheme count of
// anything else.
func banTsCommentLength(value string) int {
	for index := 0; index < len(value); index++ {
		if value[index] >= utf8.RuneSelf {
			return text.GraphemeCount(value)
		}
	}
	return len(value)
}

// banTsCommentValue is ESLint's `comment.value`: the text between a comment's delimiters.
//
// The boolean is false only for a block comment too short to have an interior, which error recovery
// produces for an unterminated `/*`; ESLint never sees one, because the file does not parse.
func banTsCommentValue(comment comments.Comment) (string, bool) {
	if !comment.IsBlock {
		return comment.Text[2:], true
	}
	if len(comment.Text) < 4 {
		return "", false
	}
	return comment.Text[2 : len(comment.Text)-2], true
}

// The four directives, spelled once.
const (
	banTsCommentExpectError = "expect-error"
	banTsCommentIgnore      = "ignore"
	banTsCommentNoCheck     = "nocheck"
	banTsCommentCheck       = "check"
)

// findBanTsCommentDirective is upstream's `findDirectiveInComment`, returning the directive and the
// description after it. The three patterns it reads are set out on the rule.
func findBanTsCommentDirective(value string, isBlock bool) (string, string, bool) {
	if !isBlock {
		// `^\/\/\/?\s*@ts-(check|nocheck)(.*)$` over `//` + value: the value may open with one more
		// slash, then whitespace. A line comment holds no line terminator, so `.*$` is the rest.
		rest := strings.TrimPrefix(value, "/")
		rest = strings.TrimLeftFunc(rest, type_checking.IsStrWhiteSpace)
		if directive, description, found := banTsCommentDirectiveAt(rest, banTsCommentCheck, banTsCommentNoCheck); found {
			return directive, description, true
		}
		// `^\/*\s*@ts-(expect-error|ignore)(.*)` over the value: any number of slashes, then
		// whitespace.
		rest = strings.TrimLeft(value, "/")
		rest = strings.TrimLeftFunc(rest, type_checking.IsStrWhiteSpace)
		return banTsCommentDirectiveAt(rest, banTsCommentExpectError, banTsCommentIgnore)
	}

	// `^\s*(?:\/|\*)*\s*@ts-(expect-error|ignore)(.*)` over the last line of the value, split where
	// ESLint's LINEBREAK_MATCHER splits: CRLF, CR, LF, U+2028 and U+2029.
	line := value
	if last := strings.LastIndexAny(value, "\r\n  "); last >= 0 {
		_, size := utf8.DecodeRuneInString(value[last:])
		line = value[last+size:]
	}
	rest := strings.TrimLeftFunc(line, type_checking.IsStrWhiteSpace)
	rest = strings.TrimLeft(rest, "/*")
	rest = strings.TrimLeftFunc(rest, type_checking.IsStrWhiteSpace)
	return banTsCommentDirectiveAt(rest, banTsCommentExpectError, banTsCommentIgnore)
}

// banTsCommentDirectiveAt reads `@ts-(first|second)(.*)` at the start of text, trying the
// alternatives in order as the pattern does, and returns the description up to the first line
// terminator, where JavaScript's `.` stops.
func banTsCommentDirectiveAt(text string, first string, second string) (string, string, bool) {
	rest, found := strings.CutPrefix(text, "@ts-")
	if !found {
		return "", "", false
	}
	for _, directive := range []string{first, second} {
		description, matched := strings.CutPrefix(rest, directive)
		if !matched {
			continue
		}
		if end := strings.IndexAny(description, "\r\n  "); end >= 0 {
			description = description[:end]
		}
		return directive, description, true
	}
	return "", "", false
}

// BanTsCommentSettingKind names which of the three shapes a directive's configuration took.
//
// A Go struct rather than an interface because the three carry different payloads and the rule
// switches on all three, which is exactly what upstream's enum does.
type BanTsCommentSettingKind string

const (
	// BanTsCommentBoolean is `true` or `false`: ban the directive outright, or allow it.
	BanTsCommentBoolean BanTsCommentSettingKind = "Boolean"

	// BanTsCommentRequireDescription is `"allow-with-description"`: allow it when it carries an
	// explanation of at least the minimum length.
	BanTsCommentRequireDescription BanTsCommentSettingKind = "RequireDescription"

	// BanTsCommentDescriptionFormat is `{ "descriptionFormat": "<regex>" }`: allow it when the
	// explanation both reaches the minimum length and matches the pattern.
	BanTsCommentDescriptionFormat BanTsCommentSettingKind = "DescriptionFormat"
)

// BanTsCommentSetting is how one directive is configured.
//
// `RequireDescription` is not `DescriptionFormat` with a nil pattern: the first checks the length,
// and the second with no pattern checks nothing, as upstream's truth test on `descriptionFormat`
// has it.
type BanTsCommentSetting struct {
	Kind BanTsCommentSettingKind

	// Banned is read only when Kind is BanTsCommentBoolean.
	Banned bool

	// DescriptionFormat is read only when Kind is BanTsCommentDescriptionFormat, and it is nil there
	// for `{}`, a null format and an empty one, all of which upstream treats as no setting at all.
	//
	// Compiled as JavaScript compiles it, because upstream builds it with `new RegExp(descriptionFormat)`:
	// a pattern with a lookahead is one ESLint accepts and Go's regexp cannot compile.
	DescriptionFormat *esregexp.RegExp
}

// BanTsCommentOptions configures which directives report and what an acceptable description is.
//
// typescript-eslint's five keys and its `defaultOptions`. A default that binds and does nothing is
// the failure this rule is most exposed to, so the corpus's default-options rows pin every one.
type BanTsCommentOptions struct {
	// TsExpectError defaults to allow-with-description, the only directive not defaulting to a
	// boolean. This is why a bare `"error"` config reports an undescribed `@ts-expect-error` and
	// stays silent on a described one.
	TsExpectError BanTsCommentSetting

	// TsIgnore defaults to banned, and it is the only directive offering a repair.
	TsIgnore BanTsCommentSetting

	// TsNoCheck defaults to banned.
	TsNoCheck BanTsCommentSetting

	// TsCheck defaults to ALLOWED. `@ts-check` turns checking on rather than off, so banning it by
	// default would be backwards, and this is the one default a reader is likely to guess wrong.
	TsCheck BanTsCommentSetting

	// MinimumDescriptionLength defaults to 3, counted in graphemes of the trimmed description.
	MinimumDescriptionLength int
}

// settingFor is upstream's `options[fullDirective]`, mapping a directive name onto its setting.
//
// The fallback is unreachable, because `findBanTsCommentDirective` returns only the four names the
// switch covers, and a mutant flipping it to a ban survives for that reason. Go requires a return,
// and the allowing setting is the one that cannot report on input this rule never produces.
func (o BanTsCommentOptions) settingFor(directive string) BanTsCommentSetting {
	switch directive {
	case banTsCommentIgnore:
		return o.TsIgnore
	case banTsCommentCheck:
		return o.TsCheck
	case banTsCommentNoCheck:
		return o.TsNoCheck
	case banTsCommentExpectError:
		return o.TsExpectError
	}
	return BanTsCommentSetting{Kind: BanTsCommentBoolean, Banned: false}
}

// DefaultBanTsCommentOptions is upstream's configured-nothing behavior.
//
// Exported because a fixture asserting the default has to be able to name it, and because a caller
// starting from the Go zero value would get four empty-kind settings and a minimum of zero, which
// silences the rule on every input while looking configured.
func DefaultBanTsCommentOptions() BanTsCommentOptions {
	return BanTsCommentOptions{
		TsExpectError:            BanTsCommentSetting{Kind: BanTsCommentRequireDescription},
		TsIgnore:                 BanTsCommentSetting{Kind: BanTsCommentBoolean, Banned: true},
		TsNoCheck:                BanTsCommentSetting{Kind: BanTsCommentBoolean, Banned: true},
		TsCheck:                  BanTsCommentSetting{Kind: BanTsCommentBoolean, Banned: false},
		MinimumDescriptionLength: 3,
	}
}

// banTsCommentRawOptions is the wire shape.
//
// Each directive is `json.RawMessage` rather than a typed field because the same key accepts a
// boolean, a string, and an object, which is upstream's hand-written `Deserialize` and which no
// struct tag can express. `MinimumDescriptionLength` is a pointer so that an absent key and an
// explicit zero stay distinguishable long enough for the default to be applied.
type banTsCommentRawOptions struct {
	TsExpectError            json.RawMessage `json:"ts-expect-error"`
	TsIgnore                 json.RawMessage `json:"ts-ignore"`
	TsNoCheck                json.RawMessage `json:"ts-nocheck"`
	TsCheck                  json.RawMessage `json:"ts-check"`
	MinimumDescriptionLength *int            `json:"minimumDescriptionLength"`
}

// DecodeBanTsCommentOptions maps the wire keys onto the settings the rule reads, applying upstream's
// per-key defaults for anything absent.
//
// A hand-written decoder rather than `rule.DecodeOptionsInto` because four of the five keys are
// polymorphic and none of the five defaults is a Go zero value. A value upstream refuses is refused
// here too, naming the key: an unrecognized string, a type the key does not take, and a pattern
// JavaScript cannot compile. This decoder once kept the default instead, on the reasoning that no
// error reached a user from here; a decoder's error now reaches the user as a config refusal naming
// the rule, so the fallback only hid a typo (ruled on #pd2chkx). The config layer checks the shape
// against typescript-eslint's schema before this runs, so the first two refusals are this
// decoder's own guard for a caller that reaches it directly.
func DecodeBanTsCommentOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[banTsCommentRawOptions]()(raw)
	if err != nil {
		return DefaultBanTsCommentOptions(), err
	}

	wire, _ := decoded.(banTsCommentRawOptions)
	options := DefaultBanTsCommentOptions()

	for _, directive := range []struct {
		key     string
		raw     json.RawMessage
		setting *BanTsCommentSetting
	}{
		{"ts-expect-error", wire.TsExpectError, &options.TsExpectError},
		{"ts-ignore", wire.TsIgnore, &options.TsIgnore},
		{"ts-nocheck", wire.TsNoCheck, &options.TsNoCheck},
		{"ts-check", wire.TsCheck, &options.TsCheck},
	} {
		setting, err := settingOrDefaultBanTsComment(directive.raw, *directive.setting)
		if err != nil {
			return DefaultBanTsCommentOptions(), fmt.Errorf("%s: %w", directive.key, err)
		}
		*directive.setting = setting
	}

	if wire.MinimumDescriptionLength != nil {
		options.MinimumDescriptionLength = *wire.MinimumDescriptionLength
	}

	return options, nil
}

// settingOrDefaultBanTsComment reads one directive's wire value, keeping the default when it is
// absent and refusing a value upstream refuses.
//
// It accepts what typescript-eslint's schema accepts: a boolean, the exact string
// `allow-with-description`, or an object whose only key is `descriptionFormat`, a string. Anything
// else is refused, and so is a descriptionFormat `new RegExp` would throw on, which fails ESLint's
// config load when the rule is created.
func settingOrDefaultBanTsComment(raw json.RawMessage, fallback BanTsCommentSetting) (BanTsCommentSetting, error) {
	if len(raw) == 0 {
		return fallback, nil
	}

	var asBool bool
	if json.Unmarshal(raw, &asBool) == nil {
		return BanTsCommentSetting{Kind: BanTsCommentBoolean, Banned: asBool}, nil
	}

	var asString string
	if json.Unmarshal(raw, &asString) == nil {
		if asString == "allow-with-description" {
			return BanTsCommentSetting{Kind: BanTsCommentRequireDescription}, nil
		}
		return fallback, fmt.Errorf("expected true, false, \"allow-with-description\" or {\"descriptionFormat\": ...}, got %s", raw)
	}

	var asObject struct {
		DescriptionFormat *string `json:"descriptionFormat"`
	}
	// Go whitespace: JSON wire bytes, whose whitespace is space, tab, LF and CR in either set.
	if trimmed := strings.TrimSpace(string(raw)); strings.HasPrefix(trimmed, "{") {
		if err := rule.UnmarshalOptions(raw, &asObject); err != nil {
			return fallback, err
		}
		// An empty pattern is falsy to upstream's `option.descriptionFormat` test, so it sets nothing,
		// as an absent or null one does.
		if asObject.DescriptionFormat == nil || *asObject.DescriptionFormat == "" {
			return BanTsCommentSetting{Kind: BanTsCommentDescriptionFormat}, nil
		}
		compiled, compileError := esregexp.Compile(*asObject.DescriptionFormat, "")
		if compileError != nil {
			return fallback, fmt.Errorf("descriptionFormat %q is not a pattern JavaScript can compile: %w", *asObject.DescriptionFormat, compileError)
		}
		return BanTsCommentSetting{
			Kind:              BanTsCommentDescriptionFormat,
			DescriptionFormat: compiled,
		}, nil
	}

	return fallback, fmt.Errorf("expected true, false, \"allow-with-description\" or {\"descriptionFormat\": ...}, got %s", raw)
}
