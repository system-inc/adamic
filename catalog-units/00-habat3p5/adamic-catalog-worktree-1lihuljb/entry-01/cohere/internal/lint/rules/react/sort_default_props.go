package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// SortDefaultPropsOptions configures whether the comparison folds case.
type SortDefaultPropsOptions struct {
	// IgnoreCase lowercases both keys before comparing. Absent means false.
	IgnoreCase bool
}

// DefaultSortDefaultPropsOptions is the unconfigured answer: a case sensitive comparison.
//
// Upstream reads `configuration.ignoreCase || false`, so absent and explicit false are the same
// rule, which is why the plain bool is safe here where `self-closing-comp` needed a pointer.
func DefaultSortDefaultPropsOptions() SortDefaultPropsOptions {
	return SortDefaultPropsOptions{IgnoreCase: false}
}

// DecodeSortDefaultPropsOptions reads this rule's configuration from the config layer.
//
// Our config layer unwraps the severity tuple before dispatch, so this receives upstream's option
// OBJECT rather than upstream's one-element array. Empty input is a rule configured as a bare
// `"error"` and resolves to the default.
func DecodeSortDefaultPropsOptions(raw []byte) (any, error) {
	options := DefaultSortDefaultPropsOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var wire struct {
		IgnoreCase bool `json:"ignoreCase"`
	}
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	options.IgnoreCase = wire.IgnoreCase
	return options, nil
}

var messageDefaultPropsNotSorted = rule.Message{
	Id: "propsNotSorted",
	Description: "Default prop declarations should be sorted alphabetically. A sorted list is " +
		"one a reader can scan for a name instead of reading end to end, and it keeps two people " +
		"adding a default in the same week from colliding in the same spot.",
}

// SortDefaultProps reports a defaultProps object whose keys are not in ascending order.
//
//	valid:   static defaultProps = {a: 1, b: 2}
//	valid:   static defaultProps = {b: 1, ...rest, a: 2}   a spread restarts the comparison
//	valid:   static defaultProps = {B: 1, a: 2}            uppercase sorts first by default
//	invalid: static defaultProps = {b: 1, a: 2}
//	invalid: static defaultProps = {B: 1, a: 2}            under `{"ignoreCase": true}`
//
// # Upstream deprecated this rule and shipped its fixer commented out
//
// The rule file carries an `@deprecated` tag, and its `fixable: 'code'` line, its `fix` function
// and the `fix` key in the report call are ALL commented out. So the released rule reports and
// proposes nothing, and its corpus carries zero `output` assertions, confirmed by extraction. This
// port ships no repair for the same reason, which is fidelity rather than a decline: there is no
// upstream fixer to decline.
//
// The deprecation is not in `meta.deprecated`, only in the file's JSDoc, so it does not surface
// through the plugin's metadata. It is recorded here because it is the kind of thing a later reader
// deserves to know before building on this rule.
//
// # The comparison is on RAW SOURCE TEXT, not on the resolved key
//
// Upstream's `getKey` calls `getText` on the key node, so the string being compared includes any
// quotes the source wrote. That produces an ordering nobody would design:
//
//	{a: 1, "b": 2}    REPORTS, because `"b"` begins with a quote which sorts before `a`
//	{"a": 1, b: 2}    silent, same reason running the other way
//	{[x]: 1, a: 2}    REPORTS, the computed key renders as `x` and `a` sorts before it
//
// All three measured against the installed build on 2026-08-27. Reproduced rather than corrected:
// comparing resolved names instead would change the verdict on every quoted key in the tree, and
// the corpus cannot see it because it writes no mixed-quoting object.
//
// # A spread restarts the comparison rather than being skipped
//
// Upstream's reduce returns `decls[idx + 1]` when it meets a spread, so the element AFTER the
// spread becomes the accumulator without being compared to anything. That means a spread splits the
// object into independently sorted runs, and `{b: 1, ...x, a: 2}` is clean. Measured.
//
// # Which objects are reached
//
// A class field or an assignment named `defaultProps` or `getDefaultProps`, whose value is either
// an object literal directly or an identifier naming one. A `getDefaultProps` METHOD returning an
// object is silent, because the rule looks at the value rather than at a return statement, measured
// against the installed build even though the name suggests otherwise.
//
// Upstream does not check that the surrounding class is a React component at all, so
// `const o = {defaultProps: {b: 1, a: 2}}` is silent only because the member arm requires an
// assignment, while any class field named `defaultProps` is judged regardless of what the class
// extends. Reproduced.
var SortDefaultProps = rule.Rule{
	Name:             "react/sort-default-props",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := DefaultSortDefaultPropsOptions()
		if configured, isConfigured := rule.OptionsAs[SortDefaultPropsOptions](options); isConfigured {
			settings = configured
		}

		check := func(value *ast.Node) {
			if value == nil || ctx.TypeChecker == nil {
				return
			}
			switch value.Kind {
			case ast.KindObjectLiteralExpression:
				sortDefaultPropsReportUnsorted(ctx, settings, value)

			case ast.KindIdentifier:
				// `C.defaultProps = someObject`. Upstream resolves the name and reads the
				// initializer of its first definition; the checker answers the same question, and
				// the loop over declarations rather than an index is the standing hazard here.
				symbol := ctx.TypeChecker.GetSymbolAtLocation(value)
				if symbol == nil {
					return
				}
				for _, declaration := range symbol.Declarations {
					if initializer := sortDefaultPropsInitializerOf(declaration); initializer != nil &&
						initializer.Kind == ast.KindObjectLiteralExpression {
						sortDefaultPropsReportUnsorted(ctx, settings, initializer)
						return
					}
				}
			}
		}

		return rule.Listeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if !sortDefaultPropsIsDeclaration(node.Name()) {
					return
				}
				check(node.AsPropertyDeclaration().Initializer)
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				expression := node.AsBinaryExpression()
				if expression.OperatorToken == nil ||
					expression.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}
				target := expression.Left
				if target == nil || target.Kind != ast.KindPropertyAccessExpression {
					return
				}
				if !sortDefaultPropsIsDeclaration(target.AsPropertyAccessExpression().Name()) {
					return
				}
				check(expression.Right)
			},
		}
	},
}

// sortDefaultPropsIsDeclaration reports whether a name node names the defaults object.
//
// Upstream accepts two spellings, `defaultProps` and `getDefaultProps`, in both arms.
func sortDefaultPropsIsDeclaration(name *ast.Node) bool {
	if name == nil {
		return false
	}
	text, readable := ast.TryGetTextOfPropertyName(name)
	if !readable {
		return false
	}
	return text == "defaultProps" || text == "getDefaultProps"
}

// sortDefaultPropsInitializerOf reads the value a declaration binds, for the identifier route.
func sortDefaultPropsInitializerOf(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		return declaration.AsVariableDeclaration().Initializer
	case ast.KindPropertyAssignment:
		return declaration.AsPropertyAssignment().Initializer
	case ast.KindPropertyDeclaration:
		return declaration.AsPropertyDeclaration().Initializer
	}
	return nil
}

// sortDefaultPropsReportUnsorted walks one object literal and reports each key out of order.
//
// This is upstream's `reduce`, written as a loop. Three behaviours in it are load-bearing and none
// is obvious from the shape:
//
// The accumulator starts at element ZERO and element zero is also the first thing compared, so the
// first property is compared against itself. That comparison can never fail, so it costs nothing,
// and it is kept rather than skipped because skipping it would change which element the spread
// handling below sees first.
//
// A spread does not advance the accumulator to itself; it advances to the element AFTER it, which
// is then never compared to anything. So a spread splits the object into independently sorted runs.
//
// When a property IS out of order the accumulator does NOT advance past it. So `{c, b, a}` reports
// twice, both against `c`, rather than reporting once and resyncing. Measured.
func sortDefaultPropsReportUnsorted(
	ctx rule.Context,
	settings SortDefaultPropsOptions,
	object *ast.Node,
) {
	literal := object.AsObjectLiteralExpression()
	if literal == nil || literal.Properties == nil {
		return
	}
	properties := literal.Properties.Nodes
	if len(properties) == 0 {
		return
	}

	previous := properties[0]
	for index, current := range properties {
		if current.Kind == ast.KindSpreadAssignment {
			// Upstream returns `decls[idx + 1]`, which is nil past the end and which the next
			// iteration's `getKey` would then read. It never gets the chance, because there is no
			// next iteration when the spread is last.
			//
			// Setting the accumulator to the spread ITSELF is an EQUIVALENT mutation, verified by
			// enumeration over nine spread arrangements rather than argued: a spread has no name
			// node, so its key text is the empty string, and the empty string is less than every
			// other string, so the next key can never report against it and then becomes the
			// accumulator anyway. The sweep correctly scores that rewrite as surviving. Upstream's
			// spelling is kept because it is upstream's, not because the two differ.
			if index+1 < len(properties) {
				previous = properties[index+1]
			}
			continue
		}

		previousKey := sortDefaultPropsKeyText(ctx, previous)
		currentKey := sortDefaultPropsKeyText(ctx, current)
		if settings.IgnoreCase {
			previousKey = strings.ToLower(previousKey)
			currentKey = strings.ToLower(currentKey)
		}

		if currentKey < previousKey {
			ctx.ReportNode(current, messageDefaultPropsNotSorted)
			// The accumulator deliberately does not advance, matching upstream's `return prev`.
			continue
		}
		previous = current
	}
}

// sortDefaultPropsKeyText renders the text upstream compares, which is the RAW SOURCE of the key.
//
// Upstream calls `getText(context, node.key || node.argument)`, so quotes, whitespace inside a
// computed key, and everything else the author wrote is part of the comparison. Reading the source
// span rather than the resolved name is what reproduces that.
//
// `node.argument` is the spread's operand upstream; spreads never reach here because the caller
// skips them, and the fallback is kept for a property whose name node is absent.
func sortDefaultPropsKeyText(ctx rule.Context, property *ast.Node) string {
	name := property.Name()
	if name == nil {
		return ""
	}
	// A computed key is where the two trees disagree about what "the key node" is. ESTree's `key`
	// for `{[x]: 1}` is the INNER expression `x`, and upstream renders that, while our parser wraps
	// it in a ComputedPropertyName whose own text carries the brackets. Rendering the wrapper would
	// compare `[x]`, and the bracket sorts between uppercase and lowercase, so `{[x]: 1, a: 2}`
	// would go silent where upstream reports. Measured against the installed build: that input
	// reports upstream, and this port was silent on it until the unwrap was added.
	if name.Kind == ast.KindComputedPropertyName {
		if inner := name.AsComputedPropertyName().Expression; inner != nil {
			name = inner
		}
	}
	// `Pos()` includes leading trivia, so a raw slice carries the newline, the indentation and any
	// comment written before the key. Whitespace and `/` sort below every letter, so an untrimmed
	// key compares as less than almost anything and a correctly SORTED object reports. Upstream's
	// `getText(node)` returns the node's own text with trivia already excluded, which is what
	// `rule.TokenRange` reproduces; trimming whitespace by hand is NOT enough, because a comment
	// between two keys survives it. Both shapes measured silent upstream and both reported here
	// before this used the token range.
	keyRange := rule.TokenRange(ctx.SourceFile, name)
	text := ctx.SourceFile.Text()
	pos, end := keyRange.Pos(), keyRange.End()
	if pos < 0 || end > len(text) || pos > end {
		return ""
	}
	return text[pos:end]
}
