package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferLiteralEnumMemberOptions is the rule's option surface.
type PreferLiteralEnumMemberOptions struct {
	// AllowBitwiseExpressions permits a bitwise operator in an initializer, and permits an
	// initializer to name a sibling member of the same enum while inside one.
	//
	// It defaults to FALSE, so the generic decoder would happen to produce the right zero value.
	// The decoder below is hand-rolled anyway, because relying on that coincidence leaves the next
	// person to add a default-true key inheriting a decoder that silently inverts their rule.
	AllowBitwiseExpressions bool
}

// DefaultPreferLiteralEnumMemberSettings is upstream's `defaultOptions`.
func DefaultPreferLiteralEnumMemberSettings() PreferLiteralEnumMemberOptions {
	return PreferLiteralEnumMemberOptions{AllowBitwiseExpressions: false}
}

// preferLiteralEnumMemberRawOptions is the wire shape, with a pointer so an absent key stays
// distinguishable from an explicit false.
type preferLiteralEnumMemberRawOptions struct {
	AllowBitwiseExpressions *bool `json:"allowBitwiseExpressions"`
}

// DecodePreferLiteralEnumMemberOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what arrives
// is the bare object rather than upstream's one-element array.
func DecodePreferLiteralEnumMemberOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[preferLiteralEnumMemberRawOptions]()(raw)
	if err != nil {
		return DefaultPreferLiteralEnumMemberSettings(), err
	}

	wire, _ := decoded.(preferLiteralEnumMemberRawOptions)
	options := DefaultPreferLiteralEnumMemberSettings()
	if wire.AllowBitwiseExpressions != nil {
		options.AllowBitwiseExpressions = *wire.AllowBitwiseExpressions
	}
	return options, nil
}

// PreferLiteralEnumMember flags an enum member initialized to anything but a literal value.
//
//	valid:   enum Valid { A = 42, B = 'text', C = -1, D = null, E = /re/ }
//	valid:   enum Valid { A }                             no initializer at all
//	valid:   enum Flags { A = 1 << 0, B = A | 1 }         only under allowBitwiseExpressions
//	invalid: enum Invalid { A = 2 + 2 }
//	invalid: enum Invalid { A = 'x', B = A }              naming a sibling, outside a bitwise
//	invalid: enum Invalid { A = someVariable }
//
// A member initialized from anything computed makes the enum's values depend on evaluation order and
// on whatever the initializer reads, which is exactly the property an enum exists to remove. A
// literal is auditable at the line, survives a reorder, and means the same thing to the compiler as
// it does to a reader of the diff.
//
// # The finding anchors on the member's NAME, not on the thing being judged
//
// Upstream reports `node: node.id`, so the span is the member key while the offending expression is
// its initializer. That reads as backwards, and it is deliberate rather than accidental: every one of
// upstream's thirty-two recorded findings carries an end column exactly one past its start on a
// single-character key. This port anchors the same way, and the fixtures slice the expected text out
// of upstream's own line and column numbers so the choice is pinned rather than described.
//
// # A parenthesis is a real node here and is invisible upstream
//
// This is the one place our tree and upstream's disagree about the input rather than about the
// judgment. The typescript-eslint parser produces an ESTree, which folds a parenthesis away
// entirely, so upstream's recursion never sees one. Our parser keeps KindParenthesizedExpression as
// a node, so a port that walked the initializer as written would report two of upstream's own
// passing cases:
//
//	enum Foo { A = 1 << 0, B = 1 << 1, C = 1 << 2, D = Foo.A | (Foo.B & ~Foo.C) }
//	enum Foo { A = (1) }
//
// Measured against the installed 8.x build rather than reasoned about: both are silent there, and a
// control input in the same run reported, so the silence is a verdict rather than a rule that never
// ran. The unwrap below is therefore fidelity, not a convenience. It is written as a loop rather than
// a single step because `((2))` nests, and it is deliberately NOT `ast.SkipParentheses`: that helper
// dereferences its argument, and an initializer can be nil.
//
// # What `partOfBitwiseComputation` actually gates, and the arm that surprises
//
// A member may name a sibling member only INSIDE a bitwise computation, which is why `C = A | B`
// passes under the option while `B = A` does not. The flag threads down the recursion, and the two
// unary arms treat it oppositely:
//
//	~A    sets the flag to true, so the sibling is allowed          silent
//	-A    passes the flag THROUGH unchanged, so at top level it is  reports
//	      still false and the sibling is not allowed
//
// Both measured against the installed build under allowBitwiseExpressions. The `-A` case is the one
// that looks like a defect in this port and is not: upstream's `['-', '+']` arm recurses with
// `partOfBitwiseComputation` rather than with `true`, and it is reproduced.
//
// # Key matching is on the COOKED text of the key
//
// `hasEnumMember` accepts a member whose key is an identifier of that name, or a string literal whose
// static value is that name, so a quoted key is reachable by a bare identifier:
//
//	enum Foo { 'A' = 1 << 0, B = A | 1 }    silent under allowBitwiseExpressions
//
// Measured on the installed build. A string literal node's `.Text` here is already the cooked value,
// so the quotes are not part of the comparison and no unquoting step is needed. The same applies to
// a computed access, `Foo['A']` and `Foo[`A`]`, both of which upstream resolves through the same
// helper and both of which are silent.
//
// # A numeric enum key is a divergence, and it is a parser difference rather than a decision
//
// `enum Bar { 1 = 1 }` is a parse error for the typescript-eslint parser, which refuses the file
// outright, so upstream's rule never runs on it and its silence says nothing about what the rule
// believes. Our parser recovers and hands back a real member with a numeric name. Upstream's
// `getStaticStringValue` returns the number as a string for a numeric Literal, so a reference to
// `Bar[1]` would match; this port declines to match a numeric key, because the comparison upstream
// performs on that shape has never run on any input and reproducing a code path nobody has executed
// is a guess rather than fidelity. It can only be reached by source that does not compile. Stated
// here rather than left silent.
//
// # Cost
//
// One listener on KindEnumMember, a rare anchor, and the recursion is bounded by the initializer's
// own depth. The sibling-key scan walks the enclosing enum's member list and is reached only for an
// identifier or member access inside a bitwise computation, which requires the option to be on.
// No checker, no program, no per-file state.
var PreferLiteralEnumMember = rule.Rule{
	Name: "@typescript-eslint/prefer-literal-enum-member",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[PreferLiteralEnumMemberOptions](options)
		if !isSettings {
			settings = DefaultPreferLiteralEnumMemberSettings()
		}

		return rule.Listeners{
			ast.KindEnumMember: func(node *ast.Node) {
				member := node.AsEnumMember()

				// No initializer means the member is just its name, which the rule ignores.
				if member.Initializer == nil {
					return
				}

				declaration := enclosingEnumDeclaration(node)
				if declaration == nil {
					return
				}

				if isAllowedEnumInitializer(declaration, member.Initializer, false, settings.AllowBitwiseExpressions) {
					return
				}

				message := preferLiteralEnumMemberNotLiteralMessage
				if settings.AllowBitwiseExpressions {
					message = preferLiteralEnumMemberNotLiteralOrBitwiseMessage
				}
				ctx.ReportNode(member.Name(), message)
			},
		}
	},
}

var preferLiteralEnumMemberNotLiteralMessage = rule.Message{
	Id:          "notLiteral",
	Description: "Explicit enum value must only be a literal value (string or number).",
}

var preferLiteralEnumMemberNotLiteralOrBitwiseMessage = rule.Message{
	Id: "notLiteralOrBitwiseExpression",
	Description: "Explicit enum value must only be a literal value (string or number) " +
		"or a bitwise expression.",
}

// enclosingEnumDeclaration walks up from a member to the enum that holds it.
//
// Upstream reads `node.parent.parent`, because its enum has a body node between the declaration and
// its members. Ours does not, so the same walk is written as a search for the kind rather than as a
// fixed number of steps: counting parents would encode the shape of a tree we do not have, and would
// break silently rather than loudly if that shape ever changed.
func enclosingEnumDeclaration(member *ast.Node) *ast.Node {
	for parent := member.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindEnumDeclaration {
			return parent
		}
	}
	return nil
}

// isAllowedEnumInitializer is upstream's `isAllowedInitializerExpressionRecursive`.
//
// `partOfBitwiseComputation` says whether the expression sits inside a bitwise operator, which is the
// only context where naming a sibling member is permitted.
func isAllowedEnumInitializer(
	declaration *ast.Node,
	expression *ast.Node,
	partOfBitwiseComputation bool,
	allowBitwiseExpressions bool,
) bool {
	if expression == nil {
		return false
	}

	// The parenthesis unwrap, and the reason it is here rather than at the call site: upstream's
	// parser removes these nodes before its rule runs, so every recursion below expects to be
	// looking at the expression itself. Doing it once at the top of the recursion puts the unwrap
	// on every path rather than only on the outermost one, which is what `((2))` needs.
	for expression.Kind == ast.KindParenthesizedExpression {
		inner := expression.AsParenthesizedExpression().Expression
		if inner == nil {
			return false
		}
		expression = inner
	}

	if partOfBitwiseComputation && isSelfEnumMember(declaration, expression) {
		return true
	}

	switch expression.Kind {
	// Every shape the typescript-eslint parser calls a Literal. Ours splits them across separate
	// kinds, so the single upstream arm becomes this list. `null` and a regular expression are here
	// because upstream reaches them through the same Literal case, verified silent on the installed
	// build; `true` and `false` are here for the same reason and appear in no upstream case.
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNullKeyword,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		// A backtick string with no interpolation. Upstream's TemplateLiteral arm accepts exactly
		// the case where the expression list is empty, and our parser gives that shape its own kind,
		// so the emptiness test is the kind test. A template WITH an interpolation parses as
		// KindTemplateExpression, which falls to the default arm and reports, matching upstream.
		ast.KindNoSubstitutionTemplateLiteral:
		return true

	case ast.KindPrefixUnaryExpression:
		unary := expression.AsPrefixUnaryExpression()

		// `+123` and `-123`. The flag is passed THROUGH rather than set, so a sibling member under
		// a minus sign at the top level is still not allowed. Measured, not inferred.
		if unary.Operator == ast.KindPlusToken || unary.Operator == ast.KindMinusToken {
			return isAllowedEnumInitializer(declaration, unary.Operand,
				partOfBitwiseComputation, allowBitwiseExpressions)
		}

		if allowBitwiseExpressions {
			return unary.Operator == ast.KindTildeToken &&
				isAllowedEnumInitializer(declaration, unary.Operand, true, allowBitwiseExpressions)
		}
		return false

	case ast.KindBinaryExpression:
		if !allowBitwiseExpressions {
			return false
		}
		binary := expression.AsBinaryExpression()
		if !isBitwiseEnumOperator(binary.OperatorToken.Kind) {
			return false
		}
		return isAllowedEnumInitializer(declaration, binary.Left, true, allowBitwiseExpressions) &&
			isAllowedEnumInitializer(declaration, binary.Right, true, allowBitwiseExpressions)

	default:
		return false
	}
}

// isBitwiseEnumOperator is upstream's `['&', '^', '<<', '>>', '>>>', '|']`.
//
// Written as a switch over token kinds rather than a set of strings, so a token that renders the same
// way in a different context cannot join the list by accident.
func isBitwiseEnumOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindAmpersandToken,
		ast.KindCaretToken,
		ast.KindLessThanLessThanToken,
		ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken,
		ast.KindBarToken:
		return true
	}
	return false
}

// isSelfEnumMember answers whether an expression names a member of the enum being declared.
//
// Three spellings reach a member, and all three are upstream's: a bare identifier, `Enum.Member`, and
// a computed `Enum['Member']` whose key is a static string.
func isSelfEnumMember(declaration *ast.Node, expression *ast.Node) bool {
	if expression.Kind == ast.KindIdentifier {
		return hasEnumMemberNamed(declaration, expression.AsIdentifier().Text)
	}

	if expression.Kind == ast.KindPropertyAccessExpression {
		access := expression.AsPropertyAccessExpression()
		if !isIdentifierNamed(access.Expression, enumDeclarationName(declaration)) {
			return false
		}
		if access.Name() != nil && access.Name().Kind == ast.KindIdentifier {
			return hasEnumMemberNamed(declaration, access.Name().AsIdentifier().Text)
		}
		return false
	}

	if expression.Kind == ast.KindElementAccessExpression {
		access := expression.AsElementAccessExpression()
		if !isIdentifierNamed(access.Expression, enumDeclarationName(declaration)) {
			return false
		}
		// Upstream reads the key through `getStaticStringValue`, which answers for a string literal
		// and for a template with no interpolation, and returns null for anything else. An empty
		// answer is falsy there and declines here, which is the same verdict by a different route.
		propertyName, isStatic := staticStringOfEnumKey(access.ArgumentExpression)
		if !isStatic || propertyName == "" {
			return false
		}
		return hasEnumMemberNamed(declaration, propertyName)
	}

	return false
}

// hasEnumMemberNamed is upstream's `hasEnumMember`.
//
// A member key is either an identifier or a string literal, and upstream compares the identifier's
// name or the literal's static value. Our parser hands a string literal's `.Text` already cooked, so
// the two branches read the same value and no unquoting happens here.
func hasEnumMemberNamed(declaration *ast.Node, name string) bool {
	if name == "" {
		return false
	}
	for _, member := range declaration.AsEnumDeclaration().Members.Nodes {
		key := member.AsEnumMember().Name()
		if key == nil {
			continue
		}
		switch key.Kind {
		case ast.KindIdentifier:
			if key.AsIdentifier().Text == name {
				return true
			}
		case ast.KindStringLiteral:
			if key.AsStringLiteral().Text == name {
				return true
			}
		case ast.KindNoSubstitutionTemplateLiteral:
			if key.AsNoSubstitutionTemplateLiteral().Text == name {
				return true
			}
		}
	}
	return false
}

// staticStringOfEnumKey is the slice of upstream's `getStaticStringValue` this rule can reach.
//
// It is asked only about the key of a computed access, where upstream's own helper answers for a
// string literal and for a template with no interpolation. The other shapes that helper handles,
// a numeric literal, `null`, a regular expression and a bigint, all render to a string upstream
// would then compare against a member name; none of them can be a legal enum key, so no input
// reaches them, and returning false rather than a rendered number keeps this port from claiming a
// behavior nobody has run.
func staticStringOfEnumKey(key *ast.Node) (string, bool) {
	if key == nil {
		return "", false
	}
	switch key.Kind {
	case ast.KindStringLiteral:
		return key.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return key.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

// isIdentifierNamed is upstream's `isIdentifierWithName`.
func isIdentifierNamed(node *ast.Node, name string) bool {
	if node == nil || name == "" {
		return false
	}
	return node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == name
}

// enumDeclarationName is the enum's own name, which a `Enum.Member` reference has to match.
func enumDeclarationName(declaration *ast.Node) string {
	name := declaration.AsEnumDeclaration().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}
