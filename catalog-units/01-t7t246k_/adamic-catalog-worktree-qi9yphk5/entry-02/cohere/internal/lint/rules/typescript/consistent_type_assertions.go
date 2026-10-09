package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConsistentTypeAssertionsStyle is which assertion spelling the rule enforces.
type ConsistentTypeAssertionsStyle string

const (
	// ConsistentTypeAssertionsAs requires `x as T`, which is the default.
	ConsistentTypeAssertionsAs ConsistentTypeAssertionsStyle = "as"
	// ConsistentTypeAssertionsAngleBracket requires `<T>x`.
	ConsistentTypeAssertionsAngleBracket ConsistentTypeAssertionsStyle = "angle-bracket"
	// ConsistentTypeAssertionsNever forbids both.
	ConsistentTypeAssertionsNever ConsistentTypeAssertionsStyle = "never"
)

// ConsistentTypeAssertionsLiteralSetting is how a literal used as an initializer is treated.
type ConsistentTypeAssertionsLiteralSetting string

const (
	// ConsistentTypeAssertionsAllow permits an assertion on a literal, the default.
	ConsistentTypeAssertionsAllow ConsistentTypeAssertionsLiteralSetting = "allow"
	// ConsistentTypeAssertionsAllowAsParameter permits one only where the literal is an argument.
	ConsistentTypeAssertionsAllowAsParameter ConsistentTypeAssertionsLiteralSetting = "allow-as-parameter"
	// ConsistentTypeAssertionsNeverLiteral forbids one outright.
	ConsistentTypeAssertionsNeverLiteral ConsistentTypeAssertionsLiteralSetting = "never"
)

// ConsistentTypeAssertionsOptions is the rule's option surface.
type ConsistentTypeAssertionsOptions struct {
	AssertionStyle             ConsistentTypeAssertionsStyle
	ObjectLiteralTypeAssertion ConsistentTypeAssertionsLiteralSetting
	ArrayLiteralTypeAssertion  ConsistentTypeAssertionsLiteralSetting
}

// DefaultConsistentTypeAssertionsSettings is upstream's `defaultOptions`.
//
// Note that NONE of the three defaults is a zero value, which makes the decoder load bearing rather
// than defensive: a decoder handing back a zero struct would give an empty assertion style that
// matches no branch, and the rule would report nothing on every input.
func DefaultConsistentTypeAssertionsSettings() ConsistentTypeAssertionsOptions {
	return ConsistentTypeAssertionsOptions{
		AssertionStyle:             ConsistentTypeAssertionsAs,
		ObjectLiteralTypeAssertion: ConsistentTypeAssertionsAllow,
		ArrayLiteralTypeAssertion:  ConsistentTypeAssertionsAllow,
	}
}

// consistentTypeAssertionsRawOptions is the wire shape, with pointers so an absent key stays
// distinguishable from an explicit one.
type consistentTypeAssertionsRawOptions struct {
	AssertionStyle             *string `json:"assertionStyle"`
	ObjectLiteralTypeAssertion *string `json:"objectLiteralTypeAssertions"`
	ArrayLiteralTypeAssertion  *string `json:"arrayLiteralTypeAssertions"`
}

// DecodeConsistentTypeAssertionsOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what arrives
// is the bare object rather than upstream's one-element array.
//
// Upstream's schema is a discriminated union: `{assertionStyle: "never"}` admits no other key, while
// the other two styles carry the two literal settings. The union is not reproduced as a type here,
// because the rule's own body already ignores both literal settings when the style is never, and a
// second representation of that fact is a second place for it to drift.
func DecodeConsistentTypeAssertionsOptions(raw []byte) (any, error) {
	options := DefaultConsistentTypeAssertionsSettings()
	if len(raw) == 0 {
		return options, nil
	}

	decoded, err := rule.DecodeOptionsInto[consistentTypeAssertionsRawOptions]()(raw)
	if err != nil {
		return options, err
	}
	wire, _ := decoded.(consistentTypeAssertionsRawOptions)

	if wire.AssertionStyle != nil {
		switch ConsistentTypeAssertionsStyle(*wire.AssertionStyle) {
		case ConsistentTypeAssertionsAs,
			ConsistentTypeAssertionsAngleBracket,
			ConsistentTypeAssertionsNever:
			options.AssertionStyle = ConsistentTypeAssertionsStyle(*wire.AssertionStyle)
		}
	}
	options.ObjectLiteralTypeAssertion = consistentTypeAssertionsLiteralSettingOf(
		wire.ObjectLiteralTypeAssertion, options.ObjectLiteralTypeAssertion)
	options.ArrayLiteralTypeAssertion = consistentTypeAssertionsLiteralSettingOf(
		wire.ArrayLiteralTypeAssertion, options.ArrayLiteralTypeAssertion)
	return options, nil
}

// consistentTypeAssertionsLiteralSettingOf reads one literal setting, keeping the default when the
// key is absent or carries a value outside upstream's enum.
//
// A value outside the enum is a configuration error ESLint refuses before the rule runs, so there is
// nothing to reproduce; falling back to the default rather than to a zero value keeps a mistyped
// configuration from silently turning the check off.
func consistentTypeAssertionsLiteralSettingOf(
	wire *string,
	fallback ConsistentTypeAssertionsLiteralSetting,
) ConsistentTypeAssertionsLiteralSetting {
	if wire == nil {
		return fallback
	}
	switch ConsistentTypeAssertionsLiteralSetting(*wire) {
	case ConsistentTypeAssertionsAllow,
		ConsistentTypeAssertionsAllowAsParameter,
		ConsistentTypeAssertionsNeverLiteral:
		return ConsistentTypeAssertionsLiteralSetting(*wire)
	}
	return fallback
}

// ConsistentTypeAssertions enforces one spelling of a type assertion, and optionally bans asserting
// a literal into place.
//
//	valid (default):   const x = y as T;
//	valid:             const x = <const>[1, 2];          `as const` is exempt from every style
//	invalid (default): const x = <T>y;                   angle-bracket where `as` is required
//	invalid (never):   const x = y as T;                 no assertion is allowed at all
//	invalid:           const x = {} as T;                under objectLiteralTypeAssertions: never
//
// Two separate judgments live in one rule. The first is spelling: pick `as` or the angle bracket and
// use it everywhere, because mixing them makes assertions harder to grep and the angle bracket
// collides with JSX. The second is stronger and is off by default: asserting a literal into a type
// is a claim the compiler cannot check, where an annotation on the binding gets checked, so
// `const x: T = {}` is safer than `const x = {} as T`.
//
// # `as const` is exempt everywhere, and the two exemptions are different
//
// The style check exempts it only under `never`, so `<const>x` still reports under `as`. The literal
// checks exempt it through a type test that also lets a QUALIFIED name through, so `Foo.const` is not
// treated as the const assertion it resembles. Both are upstream's and both are reproduced.
//
// # The fixer reasons about PRECEDENCE, which is where a naive text swap breaks
//
// Only one direction is fixable: angle bracket to `as`. The repair is not a text swap, because `as`
// binds more loosely than the angle-bracket form, so both the asserted expression and the whole
// assertion may need wrapping:
//
//	<string>(foo + bar)   ->   (foo + bar) as string      the operand keeps its parentheses
//	<string>foo + bar     ->   (foo as string) + bar      the assertion gains them
//
// Upstream computes this with two precedence comparisons, one against the expression and one against
// the parent, and skips the second when the assertion is already parenthesized. All three parts are
// reproduced, and twenty-one of upstream's cases record the exact text the result must be.
//
// # The literal checks offer SUGGESTIONS rather than a fix, and that distinction is upstream's
//
// Turning `{} as T` into `const x: T = {}` moves the type to another node, and turning it into
// `{} satisfies T` changes what is checked. Neither is a spelling change, so neither is applied
// unattended. The annotation suggestion is offered only when the assertion initializes a declarator
// that has no annotation already, because otherwise there would be two.
//
// # Cost
//
// Two rare anchors and no checker: every decision here is syntactic. The rule declares no type
// checker for that reason, which is worth stating because upstream reaches for parser services in
// its fixer, purely to map a node across trees rather than to ask a type question.
var ConsistentTypeAssertions = rule.Rule{
	Name: "@typescript-eslint/consistent-type-assertions",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[ConsistentTypeAssertionsOptions](options)
		if !isSettings {
			settings = DefaultConsistentTypeAssertionsSettings()
		}
		state := &consistentTypeAssertionsState{ctx: ctx, settings: settings}

		return rule.Listeners{
			ast.KindAsExpression: func(node *ast.Node) {
				assertion := node.AsAsExpression()
				if settings.AssertionStyle != ConsistentTypeAssertionsAs {
					state.reportIncorrectStyle(node, assertion.Expression, assertion.Type)
					return
				}
				state.checkLiteralAssertion(node, assertion.Expression, assertion.Type)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				assertion := node.AsTypeAssertion()
				if settings.AssertionStyle != ConsistentTypeAssertionsAngleBracket {
					state.reportIncorrectStyle(node, assertion.Expression, assertion.Type)
					return
				}
				state.checkLiteralAssertion(node, assertion.Expression, assertion.Type)
			},
		}
	},
}

// consistentTypeAssertionsState carries what every step needs.
type consistentTypeAssertionsState struct {
	ctx      rule.Context
	settings ConsistentTypeAssertionsOptions
}

// textOf reads a node's own source, excluding leading trivia.
func (state *consistentTypeAssertionsState) textOf(node *ast.Node) string {
	nodeRange := rule.TokenRange(state.ctx.SourceFile, node)
	return state.ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
}

// reportIncorrectStyle is upstream's `reportIncorrectAssertionType`.
func (state *consistentTypeAssertionsState) reportIncorrectStyle(
	node *ast.Node,
	expression *ast.Node,
	typeNode *ast.Node,
) {
	// `as const` is not an assertion anybody can rewrite, so `never` leaves it alone. The other two
	// styles still report it, because the spelling is the thing being enforced.
	if state.settings.AssertionStyle == ConsistentTypeAssertionsNever && isConstTypeNode(typeNode) {
		return
	}

	if state.settings.AssertionStyle == ConsistentTypeAssertionsNever {
		state.ctx.ReportNode(node, rule.Message{
			Id:          "never",
			Description: "Do not use any type assertions.",
		})
		return
	}

	cast := state.textOf(typeNode)
	if state.settings.AssertionStyle == ConsistentTypeAssertionsAngleBracket {
		state.ctx.ReportNode(node, rule.Message{
			Id:          "angle-bracket",
			Description: "Use '<" + cast + ">' instead of 'as " + cast + "'.",
		})
		return
	}

	// Only the angle-bracket-to-`as` direction is fixable upstream, and this is it.
	state.ctx.ReportNodeWithFixes(node, rule.Message{
		Id:          "as",
		Description: "Use 'as " + cast + "' instead of '<" + cast + ">'.",
	}, rule.ReplaceRange(rule.TokenRange(state.ctx.SourceFile, node),
		state.buildAsExpressionText(node, expression, cast)))
}

// buildAsExpressionText is upstream's fixer body.
//
// Two wrappings, decided independently. The operand is wrapped when it binds no tighter than `as`,
// and the whole result is wrapped when `as` binds no tighter than the parent, which is skipped when
// the assertion already sits inside parentheses and therefore needs none of its own.
func (state *consistentTypeAssertionsState) buildAsExpressionText(
	node *ast.Node,
	expression *ast.Node,
	cast string,
) string {
	asPrecedence := ast.GetOperatorPrecedence(ast.KindAsExpression, ast.KindUnknown,
		ast.OperatorPrecedenceFlagsNone)

	expressionText := wrapWhenLooser(state.textOf(expression),
		assertedExpressionPrecedence(expression), asPrecedence)
	text := expressionText + " as " + cast

	// An assertion already inside parentheses borrows them, so adding a second pair would be noise.
	// Upstream expresses this as `isParenthesized(node, sourceCode)`; our parser keeps the
	// parenthesis as a real node, so the same question is whether the parent is one.
	if node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		return text
	}
	return wrapWhenLooser(text, asPrecedence, parentPrecedenceOf(node))
}

// wrapWhenLooser is upstream's `getWrappedCode`.
//
// It parenthesizes when the inner form does NOT bind more tightly, which is the direction that
// matters: equal precedence still needs the parentheses, because associativity decides the rest and
// the fixer has no way to know it is safe.
func wrapWhenLooser(text string, inner ast.OperatorPrecedence, outer ast.OperatorPrecedence) string {
	if inner > outer {
		return text
	}
	return "(" + text + ")"
}

// assertedExpressionPrecedence is upstream's `getOperatorPrecedenceForNode` on the asserted operand.
//
// `ast.GetExpressionPrecedence` answers it for every shape but one. `Generic<string>` is a
// TSInstantiationExpression in upstream's tree and a KindExpressionWithTypeArguments in ours, and the
// two tables disagree about it: ours calls it a member-access-like 20, upstream treats it as low
// enough that `as` must wrap it. Measured against the installed build, which fixes
// `new (<Foo>Generic<string>)()` to `new ((Generic<string>) as Foo)()` with the operand
// parenthesized. Two of upstream's own cases record that output, which is what caught it.
func assertedExpressionPrecedence(expression *ast.Node) ast.OperatorPrecedence {
	if expression.Kind == ast.KindExpressionWithTypeArguments {
		return ast.OperatorPrecedenceLowest
	}
	return ast.GetExpressionPrecedence(expression)
}

// parentPrecedenceOf is upstream's parent-side `getOperatorPrecedence` call.
//
// A binary parent contributes its operator, and a `new` parent contributes whether it was written
// with arguments, which changes how tightly it binds. Both are upstream's and both are passed
// through rather than defaulted, because defaulting either produces a wrong parenthesization on a
// shape the corpus records.
func parentPrecedenceOf(node *ast.Node) ast.OperatorPrecedence {
	parent := node.Parent
	if parent == nil {
		return ast.OperatorPrecedenceLowest
	}

	// An arrow function's EXPRESSION body needs the assertion wrapped, and the two compilers'
	// tables disagree about why. Upstream's kind-based helper puts ArrowFunction in the primary
	// group alongside the literals, so `as` never binds more tightly and the wrap always happens;
	// ours answers the Yield precedence of 3, so `as` at 11 wins and no wrap is emitted. Answering
	// Primary here reproduces upstream's table rather than our compiler's.
	//
	// Measured rather than reasoned: upstream fixes `() => <Foo>bar` to `() => (bar as Foo)` and
	// `() => <Foo>{ bar: 5 }` to `() => ({ bar: 5 } as Foo)`, while a BLOCK body and a return
	// statement are both left unwrapped. So the wrap belongs to the expression-body position rather
	// than to arrow functions generally, and only that position is special-cased here.
	if parent.Kind == ast.KindArrowFunction &&
		parent.AsArrowFunction().Body == node {
		return ast.OperatorPrecedencePrimary
	}

	operator := ast.KindUnknown
	if parent.Kind == ast.KindBinaryExpression {
		if token := parent.AsBinaryExpression().OperatorToken; token != nil {
			operator = token.Kind
		}
	}

	flags := ast.OperatorPrecedenceFlagsNone
	if parent.Kind == ast.KindNewExpression {
		arguments := parent.AsNewExpression().Arguments
		if arguments == nil || len(arguments.Nodes) == 0 {
			flags = ast.OperatorPrecedenceFlagsNewWithoutArguments
		}
	}
	return ast.GetOperatorPrecedence(parent.Kind, operator, flags)
}

// checkLiteralAssertion is upstream's two literal-assertion checks.
//
// They differ only in which literal kind and which setting they read, so they are one function here
// with those two as parameters. The message ids stay distinct, because upstream reports a different
// one for each and the corpus asserts both.
func (state *consistentTypeAssertionsState) checkLiteralAssertion(
	node *ast.Node,
	expression *ast.Node,
	typeNode *ast.Node,
) {
	if expression == nil || typeNode == nil {
		return
	}

	switch expression.Kind {
	case ast.KindObjectLiteralExpression:
		state.reportLiteralAssertion(node, expression, typeNode,
			state.settings.ObjectLiteralTypeAssertion,
			"unexpectedObjectTypeAssertion", "Always prefer const x: T = { ... }.",
			"replaceObjectTypeAssertionWithAnnotation", "Use const x: {{cast}} = { ... } instead.",
			"replaceObjectTypeAssertionWithSatisfies", "Use const x = { ... } satisfies {{cast}} instead.")
	case ast.KindArrayLiteralExpression:
		state.reportLiteralAssertion(node, expression, typeNode,
			state.settings.ArrayLiteralTypeAssertion,
			"unexpectedArrayTypeAssertion", "Always prefer const x: T[] = [ ... ].",
			"replaceArrayTypeAssertionWithAnnotation", "Use const x: {{cast}} = [ ... ] instead.",
			"replaceArrayTypeAssertionWithSatisfies", "Use const x = [ ... ] satisfies {{cast}} instead.")
	}
}

// reportLiteralAssertion is the shared body of upstream's two literal checks.
func (state *consistentTypeAssertionsState) reportLiteralAssertion(
	node *ast.Node,
	expression *ast.Node,
	typeNode *ast.Node,
	setting ConsistentTypeAssertionsLiteralSetting,
	reportId string, reportText string,
	annotationId string, annotationTemplate string,
	satisfiesId string, satisfiesTemplate string,
) {
	if setting == ConsistentTypeAssertionsAllow {
		return
	}
	if setting == ConsistentTypeAssertionsAllowAsParameter && isAssertionInParameterPosition(node) {
		return
	}
	if !isCheckableAssertionType(typeNode) {
		return
	}

	cast := state.textOf(typeNode)
	state.ctx.ReportNodeWithSuggestions(node,
		rule.Message{Id: reportId, Description: reportText},
		state.buildLiteralSuggestions(node, expression, cast,
			annotationId, annotationTemplate, satisfiesId, satisfiesTemplate)...)
}

// buildLiteralSuggestions is upstream's `getSuggestions`.
//
// The annotation suggestion is offered only when the assertion initializes a declarator that carries
// no annotation of its own, because adding one where a type already sits would produce two. The
// satisfies suggestion is always offered.
func (state *consistentTypeAssertionsState) buildLiteralSuggestions(
	node *ast.Node,
	expression *ast.Node,
	cast string,
	annotationId string, annotationTemplate string,
	satisfiesId string, satisfiesTemplate string,
) []rule.Suggestion {
	// Upstream reads `getTextWithParentheses`, which keeps a parenthesized expression's own
	// parentheses. Our parser makes the parenthesis a node, so its text already carries them and
	// reading the node is the same answer.
	expressionText := state.textOf(expression)
	nodeRange := rule.TokenRange(state.ctx.SourceFile, node)

	var suggestions []rule.Suggestion
	if declarator := declaratorWithoutAnnotation(node); declarator != nil {
		nameRange := rule.TokenRange(state.ctx.SourceFile, declarator.AsVariableDeclaration().Name())
		suggestions = append(suggestions, rule.Suggestion{
			Message: rule.Message{
				Id:          annotationId,
				Description: renderConsistentTypeAssertionsCast(annotationTemplate, cast),
			},
			Fixes: []rule.Fix{
				rule.ReplaceRange(core.NewTextRange(nameRange.End(), nameRange.End()), ": "+cast),
				rule.ReplaceRange(nodeRange, expressionText),
			},
		})
	}

	suggestions = append(suggestions, rule.Suggestion{
		Message: rule.Message{
			Id:          satisfiesId,
			Description: renderConsistentTypeAssertionsCast(satisfiesTemplate, cast),
		},
		Fixes: []rule.Fix{
			rule.ReplaceRange(nodeRange, expressionText+" satisfies "+cast),
		},
	})
	return suggestions
}

// renderConsistentTypeAssertionsCast fills upstream's one interpolation slot.
func renderConsistentTypeAssertionsCast(template string, cast string) string {
	rendered := ""
	for index := 0; index < len(template); index++ {
		if index+8 <= len(template) && template[index:index+8] == "{{cast}}" {
			rendered += cast
			index += 7
			continue
		}
		rendered += string(template[index])
	}
	return rendered
}

// declaratorWithoutAnnotation answers upstream's
// `node.parent.type === VariableDeclarator && !node.parent.id.typeAnnotation`.
func declaratorWithoutAnnotation(node *ast.Node) *ast.Node {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	declaration := parent.AsVariableDeclaration()
	if declaration.Initializer != node || declaration.Type != nil {
		return nil
	}
	if declaration.Name() == nil {
		return nil
	}
	return parent
}

// isAssertionInParameterPosition is upstream's `isAsParameter`.
//
// The name is upstream's and is broader than it sounds: a throw argument and a default value are in
// the list too, because all of these are positions where an annotation has nowhere to go.
func isAssertionInParameterPosition(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindNewExpression,
		ast.KindCallExpression,
		ast.KindThrowStatement,
		ast.KindJsxExpression:
		return true
	case ast.KindParameter:
		// Upstream's AssignmentPattern, which our parser spells as a parameter carrying a default.
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindBindingElement:
		return parent.AsBindingElement().Initializer == node
	case ast.KindTemplateSpan:
		// Upstream requires the template be a TAGGED one, because an untagged template coerces its
		// interpolation and an annotation would still be checkable there.
		if parent.Parent == nil || parent.Parent.Parent == nil {
			return false
		}
		return parent.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
	}
	return false
}

// isCheckableAssertionType is upstream's `checkType`.
//
// `any` and `unknown` are exempt because asserting a literal into either says nothing an annotation
// would say better. `as const` is exempt, unless the name is QUALIFIED, which is upstream's carve-out
// for a type that merely ends in the word.
func isCheckableAssertionType(typeNode *ast.Node) bool {
	switch typeNode.Kind {
	case ast.KindAnyKeyword, ast.KindUnknownKeyword:
		return false
	case ast.KindTypeReference:
		if !isConstTypeNode(typeNode) {
			return true
		}
		// Upstream's carve-out for a QUALIFIED name that ends in the word `const`, and it appears to
		// be unreachable. `isConstTypeNode` above already requires the whole type name be a bare
		// identifier spelled `const`, so a qualified name never reaches here; and reaching it by
		// widening that test is not possible either, because `const` is a reserved word that cannot
		// name a namespace member. Measured: `{} as Foo.const` is a parse error, while `{} as N.T`
		// reports through the branch above rather than this one.
		//
		// A mutation replacing this line with a bare false survives every fixture, which is what
		// prompted the check. It is kept because it is upstream's and because the unreachability
		// belongs to the grammar rather than to this rule: a future TypeScript that allowed the name
		// would make this line the only thing standing between the rule and a false positive.
		return typeNode.AsTypeReferenceNode().TypeName.Kind == ast.KindQualifiedName
	}
	return true
}

// isConstTypeNode is upstream's `isConst`.
func isConstTypeNode(typeNode *ast.Node) bool {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return false
	}
	typeName := typeNode.AsTypeReferenceNode().TypeName
	return typeName != nil && typeName.Kind == ast.KindIdentifier &&
		typeName.AsIdentifier().Text == "const"
}
