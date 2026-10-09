package core

import (
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noBitwiseOperatorSpellings maps a token kind to the operator text upstream renders in its message.
//
// Upstream's `BITWISE_OPERATORS` is a list of STRINGS, because ESTree carries `node.operator` as
// text. Our AST carries a token kind instead, so the list becomes this map and it does two jobs at
// once: membership is the rule's test, and the value is what the message says.
//
// Thirteen entries, which is upstream's list exactly. The four that a reader tends to expect and
// that are deliberately absent are `&&`, `||`, `??` and their assignment forms: those are logical
// rather than bitwise, and upstream's corpus asserts `a &&= b`, `a ||= b` and `a ??= b` are all
// clean. `!` is likewise absent and asserted clean.
var noBitwiseOperatorSpellings = map[ast.Kind]string{
	ast.KindCaretToken:                                   "^",
	ast.KindBarToken:                                     "|",
	ast.KindAmpersandToken:                               "&",
	ast.KindLessThanLessThanToken:                        "<<",
	ast.KindGreaterThanGreaterThanToken:                  ">>",
	ast.KindGreaterThanGreaterThanGreaterThanToken:       ">>>",
	ast.KindCaretEqualsToken:                             "^=",
	ast.KindBarEqualsToken:                               "|=",
	ast.KindAmpersandEqualsToken:                         "&=",
	ast.KindLessThanLessThanEqualsToken:                  "<<=",
	ast.KindGreaterThanGreaterThanEqualsToken:            ">>=",
	ast.KindGreaterThanGreaterThanGreaterThanEqualsToken: ">>>=",
	ast.KindTildeToken:                                   "~",
}

// NoBitwiseOptions carries upstream's two options.
//
// Both defaults are the zero value, which is the safe direction: an absent key exempts nothing and
// turns no hint on.
type NoBitwiseOptions struct {
	// Allow lists operator spellings that are permitted, e.g. `["~", "|"]`.
	Allow []string `json:"allow"`
	// Int32Hint exempts `x|0`, the idiom for coercing to a 32-bit integer.
	Int32Hint bool `json:"int32Hint"`
}

// DecodeNoBitwiseOptions reads the option object off the config.
//
// # The config layer hands this a different shape than a fixture does
//
// `internal/lint/configuration/configuration.go` keeps `setting.Options = tuple[1]`, a single JSON
// value after the severity, while ESLint's `context.options` is every element after it. This rule's
// `meta.schema` declares ONE element, so the two coincide and the object arrives whole. Written out
// rather than assumed, because a decoder wrong for the config shape passes every fixture: a fixture
// hands the decoder bytes the test built, never bytes the config layer sliced.
//
// An operator outside upstream's enum is REFUSED rather than ignored. Upstream gets that refusal
// from its schema before the rule runs; there is no schema layer here, so it lives here. A rule
// silently dropping an unknown spelling would leave a project believing it had exempted `>>>` when
// it had written `>>>=`, and the two differ.
func DecodeNoBitwiseOptions(raw []byte) (any, error) {
	var options NoBitwiseOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	for _, spelling := range options.Allow {
		if !slices.Contains(noBitwiseAllowedSpellings(), spelling) {
			return options, fmt.Errorf(
				"no-bitwise does not know the operator %q; the operators are %v",
				spelling, noBitwiseAllowedSpellings())
		}
	}
	return options, nil
}

// noBitwiseAllowedSpellings is the operator list the decoder accepts, in upstream's own order.
//
// # Derived from the map rather than written beside it, and that was a defect before it was a design
//
// The first draft wrote the thirteen spellings out here as a literal, which reads as harmless
// duplication and is not. A mutation changing ONE spelling in this list survived the whole fixture
// set: an entry the decoder accepts but the map does not contain is a spelling a project can put in
// `allow` that then exempts nothing, and an entry the map contains but this list rejects is an
// operator nobody can exempt. Both are silent, and no fixture written against either side alone can
// see them, because each side is internally consistent.
//
// So the list is now READ OUT of the map, and the order is imposed separately. The two can no longer
// disagree about membership, which is the property that matters; the order is only for the refusal
// message, and a spelling missing from `noBitwiseSpellingOrder` still appears, appended, rather than
// vanishing from the message.
func noBitwiseAllowedSpellings() []string {
	spellings := make([]string, 0, len(noBitwiseOperatorSpellings))
	for _, spelling := range noBitwiseSpellingOrder {
		if noBitwiseMapContainsSpelling(spelling) {
			spellings = append(spellings, spelling)
		}
	}
	// Anything in the map that the order does not name, so a new entry is reported rather than
	// silently excluded from the message.
	for _, spelling := range noBitwiseOperatorSpellings {
		if !slices.Contains(spellings, spelling) {
			spellings = append(spellings, spelling)
		}
	}
	return spellings
}

// noBitwiseSpellingOrder is upstream's `BITWISE_OPERATORS` order, used only to render the refusal
// message in a familiar sequence. Membership is decided by the map, never by this.
//
// A mutation corrupting an entry here SURVIVES the fixture set, and the survival is genuine
// equivalence rather than a blind spot. Corrupting `"^"` to `"^^"` makes the first loop skip it, and
// the fallback loop then appends `"^"` from the map, so the same thirteen spellings are accepted and
// only their order in the refusal message changes. The corresponding mutation on the MAP is caught
// by six lines, which is the pair that shows where the decision actually lives. Recorded because a
// later reader chasing the survivor deserves to know which of the two it is.
var noBitwiseSpellingOrder = []string{
	"^", "|", "&", "<<", ">>", ">>>", "^=", "|=", "&=", "<<=", ">>=", ">>>=", "~",
}

// noBitwiseMapContainsSpelling reports whether the rule can actually act on a spelling.
func noBitwiseMapContainsSpelling(spelling string) bool {
	for _, known := range noBitwiseOperatorSpellings {
		if known == spelling {
			return true
		}
	}
	return false
}

var messageNoBitwise = rule.Message{
	Id: "unexpected",
	Description: "A bitwise operator here is usually a typo for its logical twin: `&` for `&&`, " +
		"`|` for `||`. When it is deliberate the intent is worth stating, because the reader has " +
		"to decide which of the two you meant every time they pass it.",
}

// NoBitwise flags a bitwise operator.
//
//	valid:   a && b
//	valid:   a ||= b
//	valid:   a|0            with {"int32Hint": true}
//	valid:   ~x             with {"allow": ["~"]}
//	invalid: a & b
//	invalid: a >>>= b
//	invalid: ~a
//
// # Our AST folds three upstream node types into one, and that is a simplification rather than a gap
//
// Upstream listens on `AssignmentExpression`, `BinaryExpression` and `UnaryExpression`, because
// ESTree gives compound assignment its own node type. The TypeScript parser makes `a &= b` an
// ordinary `KindBinaryExpression` whose operator token is `&=`, so the first two listeners collapse
// into one and the operator map covers both halves of upstream's list. `~a` is the only unary form
// and arrives as `KindPrefixUnaryExpression`.
//
// The map is keyed on the token kind rather than on rendered text, so no operator has to be spelled
// twice and a kind our parser produces that upstream has no spelling for simply is not in the map.
//
// # `int32Hint` is narrower than its name suggests, and its edges were measured rather than guessed
//
// It exempts `x | 0` and nothing else: the operator must be `|`, and the RIGHT operand must be a
// numeric literal whose VALUE is zero. Measured against the installed 10.8.1 build with the hint on:
//
//	x|0     exempt          x|0.0   exempt        the value is zero, not the spelling
//	x|0x0   exempt          x|0e0   exempt
//	0|x     REPORTS         only the right operand counts, so the idiom is directional
//	x|1     REPORTS         x|'0'   REPORTS       a string is not a numeric literal
//	x|0n    REPORTS         a BigInt zero is a different literal type
//	x&0     REPORTS         the operator must be `|`
//
// The spelling rows are why this reads the numeric literal's `Text` rather than its source bytes:
// our parser has already canonicalised `0x0`, `0.0` and `0e0` to `"0"`, which is exactly the
// equivalence class upstream's `=== 0` value comparison produces. `0n` is a `KindBigIntLiteral` and
// never reaches the numeric arm, and `'0'` is a string literal, so both fall through and report
// without needing a test of their own.
//
// # The span is the whole expression
//
// Upstream reports `node`, so `a >>> b` reports columns 1 to 8 rather than pointing at the operator.
// Measured against the installed 10.8.1 build across all thirteen operators.
//
// # No fixer
//
// `meta.fixable` is unset. There is no repair: whether `&` should have been `&&` is the question the
// rule is asking, and a tool that answered it would be guessing at intent.
var NoBitwise = rule.Rule{
	Name: "no-bitwise",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as bare "error" is handed nil, so the assertion yields the zero value.
		// Correct here only because both options default to off.
		settings, _ := rule.OptionsAs[NoBitwiseOptions](options)

		report := func(node *ast.Node, operator ast.Kind) {
			spelling, bitwise := noBitwiseOperatorSpellings[operator]
			if !bitwise {
				return
			}
			if slices.Contains(settings.Allow, spelling) {
				return
			}
			if noBitwiseIsInt32Hint(node, operator, settings) {
				return
			}
			ctx.ReportNode(node, rule.Message{
				Id:          messageNoBitwise.Id,
				Description: fmt.Sprintf("Unexpected use of '%s'. %s", spelling, messageNoBitwise.Description),
			})
		}

		return rule.Listeners{
			// Compound assignment is a binary expression here, so this one listener is upstream's
			// AssignmentExpression and BinaryExpression both.
			ast.KindBinaryExpression: func(node *ast.Node) {
				report(node, node.AsBinaryExpression().OperatorToken.Kind)
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				report(node, node.AsPrefixUnaryExpression().Operator)
			},
		}
	},
}

// noBitwiseIsInt32Hint is upstream's `isInt32Hint`: the `x | 0` coercion idiom.
//
// Three conditions and all three are upstream's. The option must be on, the operator must be `|`,
// and the right operand must be a numeric literal whose value is zero. Upstream tests
// `node.right.value === 0` on an ESTree Literal; the closest thing here is the numeric literal's
// `Text`, which our parser has already canonicalised -- `0x0`, `0.0` and `0e0` all arrive as `"0"`,
// which is the same equivalence class upstream's value comparison produces.
//
// Deliberately NOT unwrapping parentheses on the right operand. Upstream's parser folds them, so
// `a|(0)` is exempt there and reports here; that is a divergence and it is pinned by a fixture
// rather than corrected, because this arm is an exemption and widening an exemption silently is the
// direction that loses findings for real.
func noBitwiseIsInt32Hint(node *ast.Node, operator ast.Kind, settings NoBitwiseOptions) bool {
	if !settings.Int32Hint || operator != ast.KindBarToken {
		return false
	}
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	right := node.AsBinaryExpression().Right
	if right == nil || right.Kind != ast.KindNumericLiteral {
		return false
	}
	return right.Text() == "0"
}
