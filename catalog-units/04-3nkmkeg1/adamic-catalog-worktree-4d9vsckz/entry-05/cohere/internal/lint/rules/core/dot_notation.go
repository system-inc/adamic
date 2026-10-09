package core

import (
	"github.com/system-inc/cohere/internal/lint/ecmascript/dotnotation"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// DotNotationOptions configures which computed accesses are exempt.
//
// Upstream's schema is a single object with `allowKeywords` (default true) and `allowPattern`
// (default empty). Our config layer unwraps the severity tuple before dispatch, so the decoder
// receives that object directly rather than upstream's one-element array.
//
// `AllowKeywords` is a POINTER because its default is TRUE, and that is the one field in this
// package where the zero value is the wrong answer. A plain `bool` decodes an absent option to
// false, which inverts the rule: every `a.class` in the tree would start reporting `useBrackets`.
// The generic `rule.DecodeOptionsInto` would do exactly that, which is why the decoder below is
// hand-written.
type DotNotationOptions struct {
	// AllowKeywords permits `a.class` and friends. Absent means true, which is upstream's default.
	AllowKeywords *bool `json:"allowKeywords"`

	// AllowPattern is a regular expression source. A computed key matching it is left alone, which
	// is how a project keeps `a['snake_case']` while still converting `a['b']`.
	AllowPattern string `json:"allowPattern"`
}

// DotNotation prefers dot notation over a bracket access with a literal key.
//
//	valid:   a.b;
//	valid:   a[b];
//	valid:   a['b-c'];
//	valid:   a['1'];
//	invalid: a['b'];
//	invalid: a[`b`];
//	invalid: a[null];
//	invalid: a.true;   (with allowKeywords: false)
//
// Ported from `dot-notation` in ESLint, read from the clone at `lib/rules/dot-notation.js`. The
// judgment and the repairs live in `ecmascript/dotnotation`, shared with
// `@typescript-eslint/dot-notation`; this file is the option surface. Two
// options, two messages, and `meta.fixable` is `"code"`, so the repair is part of the port.
//
// The whole 69-case corpus was extracted from upstream's own tester and replayed against the rule
// through the ESLint Linter API before any code was written, including all 34 `output` fixtures.
//
// # The corpus needs espree, and finding that out was the measurement
//
// Five of upstream's reporting cases are legacy octals: `01['prop']`, `08['prop']` and friends.
// Driven through `@typescript-eslint/parser` all five come back as FATAL PARSE ERRORS, which read
// exactly like clean verdicts in a findings count -- the rule reports nothing and nothing says why.
// Driving the same corpus through espree in script mode moves the oracle from 64 of 69 to 69 of 69
// and from 34 findings to 39.
//
// That is a fact about the ORACLE rather than about this rule. What our own parser does with those
// five is measured separately and recorded in the test file, because a legacy octal is a syntax
// error in TypeScript and this rule's verdict on it is therefore ours to establish rather than
// upstream's to hand over.
//
// # The fixer, and the four repairs upstream refuses to make
//
// `output: null` on a corpus case means upstream reports it and deliberately declines to fix it.
// There are four, and each is a repair that would change meaning rather than spelling:
//
//	foo[ /* comment */ 'bar' ]    a comment inside the brackets would be deleted
//	foo[ 'bar' /* comment */ ]    likewise
//	foo. /* comment */ while      likewise, between the dot and the name
//	let.if()                      `let[` at statement position parses as a DESTRUCTURING
//	                              declaration, not a member access, so bracketing it changes
//	                              what the statement is
//
// All four are reproduced. A fixer that repairs a case upstream refuses to touch is a defect no
// message-id fixture can see, and this rule's repairs are applied unattended.
//
// # The space before the dot, which is the repair most likely to corrupt a file
//
// `5['prop']` must become `5 .prop` and not `5.prop`, because the second re-lexes as the numeric
// literal `5.` followed by an identifier and stops parsing. Upstream inserts the space only when the
// object is a plain decimal integer, which is `decimalInteger` in `ecmascript/dotnotation`;
// `5.000_000` and `0b1010_1010` and `01` all take the no-space branch. Ten corpus cases pin the two
// branches against each other.
var DotNotation = rule.Rule{
	Name: "dot-notation",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := dotnotation.Settings{AllowKeywords: true}
		if resolved, isDotNotationOptions := rule.OptionsAs[DotNotationOptions](options); isDotNotationOptions {
			if resolved.AllowKeywords != nil {
				settings.AllowKeywords = *resolved.AllowKeywords
			}
			settings.AllowPattern = dotnotation.CompileAllowPattern(resolved.AllowPattern)
		}
		return dotnotation.Listeners(ctx, settings)
	},
}

// DefaultDotNotationOptions is the unconfigured answer.
//
// `allowKeywords` defaults to TRUE, which is why the field is a pointer: an absent option must not
// decode to the zero value. See the type's doc comment.
func DefaultDotNotationOptions() DotNotationOptions {
	allowKeywords := true
	return DotNotationOptions{AllowKeywords: &allowKeywords}
}

// DecodeDotNotationOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` for two reasons, and the first is a defect the
// generic helper would ship: `allowKeywords` defaults to true, so decoding an absent option into a
// plain bool would invert the rule and start reporting every `a.class` in the tree. The second is
// that a bare `"error"` configuration arrives as empty input, which the generic decoder errors on.
func DecodeDotNotationOptions(raw []byte) (any, error) {
	options := DefaultDotNotationOptions()
	if len(raw) == 0 {
		return options, nil
	}
	// Decode into a fresh value rather than over the default, so an explicit `false` is
	// distinguishable from an absent field by the pointer being non-nil.
	var decoded DotNotationOptions
	if err := rule.UnmarshalOptions(raw, &decoded); err != nil {
		return options, err
	}
	if decoded.AllowKeywords == nil {
		decoded.AllowKeywords = options.AllowKeywords
	}
	return decoded, nil
}
