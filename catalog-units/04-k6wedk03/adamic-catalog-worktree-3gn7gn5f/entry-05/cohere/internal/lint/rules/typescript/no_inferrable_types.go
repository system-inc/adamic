package typescript

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageNoInferrableTypesId is the id. The rendered message names the type, which moves.
const messageNoInferrableTypesId = "noInferrableType"

// noInferrableTypesMessage renders the finding.
//
// Upstream's template names the same type twice -- "Type {{type}} trivially inferred from a {{type}}
// literal" -- so the two are always the same word and one value fills both.
func noInferrableTypesMessage(typeName string) rule.Message {
	return rule.Message{
		Id: messageNoInferrableTypesId,
		Description: fmt.Sprintf("Type `%s` is trivially inferred from a `%s` literal, so the "+
			"annotation restates what the initializer already says. Remove it and the declaration "+
			"means exactly the same thing with one fewer place to keep in sync.", typeName, typeName),
	}
}

// NoInferrableTypesOptions is the rule's whole configuration.
//
// Both flags default to false, which is also the zero value, so unlike most options on these rules
// an absent configuration and an explicit `{}` genuinely mean the same thing here.
type NoInferrableTypesOptions struct {
	// IgnoreParameters skips a function parameter with a default value.
	IgnoreParameters bool `json:"ignoreParameters"`

	// IgnoreProperties skips a class property with an initializer.
	IgnoreProperties bool `json:"ignoreProperties"`
}

// DecodeNoInferrableTypesOptions reads this rule's configuration.
//
// Empty input is legal and means both flags off, which is upstream's `defaultOptions`. The generic
// decoder errors on empty input, which is the only reason this is hand written: the defaults here
// really are the zero value.
func DecodeNoInferrableTypesOptions(raw []byte) (any, error) {
	var options NoInferrableTypesOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, fmt.Errorf("no-inferrable-types takes an options object: %w", err)
	}
	return options, nil
}

// NoInferrableTypes reports a type annotation that restates what its initializer already proves.
//
//	valid:   const a = 5;                          no annotation to remove
//	valid:   const a: number = someCall();         the initializer proves nothing
//	valid:   class C { readonly a: number = 5; }   readonly is exempt, see below
//	invalid: const a: number = 5;
//	invalid: function f(a: string = 'x') {}
//	invalid: const a: RegExp = /a/;
//
// # The judgment is a table of annotation kind against initializer shape
//
// Each arm below is upstream's, in upstream's order, and the pairs are narrower than "the types
// match". `const a: number = 5` reports; `const a: number = getFive()` does not, because the rule
// never asks the checker anything -- it is a syntactic rule about what a reader can see, which is
// why the audit correctly records it as needing no type information.
//
// Three arms carry a unary prefix the others do not. A number may be written `-5` or `+5`, a bigint
// only `-1n` (a `+` on a bigint is a TypeScript error, which is why upstream's comment says so), and
// `undefined` may be written `void 0`. Reading the operand through the prefix is what makes those
// report; a port testing the literal directly is silent on all three.
//
// # `readonly` is exempt for a compiler reason, not a stylistic one
//
// `class C { readonly a: number = 5 }` is CLEAN. Upstream cites Microsoft/TypeScript#14416: without
// the annotation a readonly property's type narrows to the literal `5` rather than to `number`, so
// removing it changes what the class means and can break every assignment elsewhere. `optional` is
// exempt for the same family of reason. This is the `output: null` situation one level up -- a case
// where the repair would change behaviour, expressed by declining to report at all.
//
// # The fixer removes a token the annotation does not contain
//
// Removing `: number` is the whole repair for an ordinary declaration. Two shapes need one byte
// more, and both are in the corpus:
//
//	function f(a?: number = 5) {}      the `?` has to go too
//	class C { a!: number = 5; }        the `!` has to go too
//
// Neither token is inside the type annotation, so a fixer that removes only the annotation leaves
// `function f(a? = 5)`, which does not parse. Upstream reaches them with `getTokenBefore`; here they
// are their own nodes on the parameter and the property, measured, so each is removed by its own
// range rather than by scanning.
var NoInferrableTypes = rule.Rule{
	Name: "@typescript-eslint/no-inferrable-types",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, configured := rule.OptionsAs[NoInferrableTypesOptions](options)
		if !configured {
			settings = NoInferrableTypesOptions{}
		}

		listeners := rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration == nil {
					return
				}
				reportNoInferrableType(ctx, node, declaration.Type, declaration.Initializer, nil)
			},
		}

		if !settings.IgnoreParameters {
			parameterVisitor := func(node *ast.Node) {
				for _, parameter := range node.Parameters() {
					checkNoInferrableTypesParameter(ctx, parameter)
				}
			}
			// Upstream listens on the three function forms that can carry a defaulted parameter.
			// A method, constructor or accessor reaches the same code through its own kind here,
			// and upstream reaches them because ESTree nests a FunctionExpression inside each --
			// so listening on those kinds too is fidelity to the verdict rather than an addition.
			for _, kind := range []ast.Kind{
				ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
				ast.KindMethodDeclaration, ast.KindConstructor,
				ast.KindGetAccessor, ast.KindSetAccessor,
			} {
				listeners[kind] = parameterVisitor
			}
		}

		if !settings.IgnoreProperties {
			listeners[ast.KindPropertyDeclaration] = func(node *ast.Node) {
				checkNoInferrableTypesProperty(ctx, node)
			}
		}

		return listeners
	},
}

// checkNoInferrableTypesParameter handles a defaulted function parameter.
//
// Upstream unwraps a `TSParameterProperty` to its inner parameter and then requires an
// `AssignmentPattern`, which is its shape for a parameter with a default. Here a parameter property
// is an ordinary `KindParameter` carrying modifiers and the default is its `Initializer`, so the
// FIELD reads collapse into reading the parameter directly.
//
// # The report span does not collapse with them, and that difference was measured rather than
// reasoned
//
// Unwrapping to `param.parameter` also moves where upstream REPORTS: the finding starts at the
// parameter's name, after `public readonly`, not at the modifiers. A first draft reported the whole
// parameter node, which is what this tree's flat shape hands you, and a differential over real files
// caught it -- 202 of 204 findings matched position exactly and the two that did not were both
// parameter properties, differing only in column:
//
//	public readonly output: string = ''      upstream column 29, the naive port's column 13
//	public active: boolean = true            upstream column 20, the naive port's column 13
//
// Neither the 99-case corpus nor any message-id fixture can see this. Upstream's corpus has
// parameter-property cases, but `RuleTester` compares a column only when a case declares one, and
// none of them do. The span is anchored on the name for that reason.
func checkNoInferrableTypesParameter(ctx rule.Context, parameter *ast.Node) {
	if parameter == nil || parameter.Kind != ast.KindParameter {
		return
	}
	declaration := parameter.AsParameterDeclaration()
	if declaration == nil || declaration.Initializer == nil {
		// No default value, so nothing proves the annotation redundant.
		return
	}

	// A parameter carrying an accessibility or `readonly` modifier is upstream's
	// TSParameterProperty, and upstream reports its INNER parameter. The equivalent here is to
	// anchor on the name, since the modifiers are the only thing between the two.
	anchor := parameter
	if parameter.Modifiers() != nil && len(parameter.Modifiers().Nodes) > 0 {
		if name := declaration.Name(); name != nil {
			anchor = name
		}
	}

	// The `?` on an optional parameter is removed alongside the annotation, because
	// `function f(a? = 5)` does not parse.
	reportNoInferrableType(ctx, anchor, declaration.Type, declaration.Initializer,
		declaration.QuestionToken)
}

// checkNoInferrableTypesProperty handles a class property with an initializer.
func checkNoInferrableTypesProperty(ctx rule.Context, node *ast.Node) {
	declaration := node.AsPropertyDeclaration()
	if declaration == nil || declaration.Initializer == nil {
		return
	}

	// `readonly` is exempt because removing the annotation narrows the property to its literal
	// type, which changes the class rather than tidying it. `optional` is exempt beside it.
	// Upstream cites Microsoft/TypeScript#14416.
	if noInferrableTypesHasModifier(node, ast.KindReadonlyKeyword) {
		return
	}
	if declaration.PostfixToken != nil && declaration.PostfixToken.Kind == ast.KindQuestionToken {
		return
	}

	// A definite-assignment `!` is removed alongside the annotation, for the same parse reason the
	// parameter's `?` is.
	var definite *ast.Node
	if declaration.PostfixToken != nil && declaration.PostfixToken.Kind == ast.KindExclamationToken {
		definite = declaration.PostfixToken
	}
	reportNoInferrableType(ctx, node, declaration.Type, declaration.Initializer, definite)
}

// reportNoInferrableType is upstream's `reportInferrableType`.
//
// `extraToken` is the `?` or `!` that has to be removed alongside the annotation, or nil.
func reportNoInferrableType(
	ctx rule.Context,
	node *ast.Node,
	annotation *ast.Node,
	initializer *ast.Node,
	extraToken *ast.Node,
) {
	if annotation == nil || initializer == nil {
		return
	}
	typeName, inferrable := noInferrableTypesAnnotationName(annotation, initializer)
	if !inferrable {
		return
	}

	fixes := []rule.Fix{}
	if extraToken != nil {
		fixes = append(fixes, rule.RemoveRange(rule.TokenRange(ctx.SourceFile, extraToken)))
	}
	// The annotation's own span does not include the `:` that introduces it, so the removal runs
	// from just past the preceding token. Leaving the colon behind writes `const a: = 5`.
	fixes = append(fixes, rule.RemoveRange(noInferrableTypesAnnotationSpan(ctx, annotation)))

	ctx.ReportNodeWithFixes(node, noInferrableTypesMessage(typeName), fixes...)
}

// noInferrableTypesAnnotationSpan is the span to delete for an annotation, colon included.
//
// The colon is found by scanning back from the type node over whitespace, which is where it must be:
// nothing else can sit between a name and its type annotation. Comments are deliberately not skipped
// past, matching upstream, which removes the annotation node and its preceding token rather than a
// computed range.
func noInferrableTypesAnnotationSpan(ctx rule.Context, annotation *ast.Node) core.TextRange {
	annotationRange := rule.TokenRange(ctx.SourceFile, annotation)
	if ctx.SourceFile == nil {
		return annotationRange
	}
	text := ctx.SourceFile.Text()
	offset := annotationRange.Pos() - 1
	for offset >= 0 && noInferrableTypesIsSpaceByte(text[offset]) {
		offset--
	}
	if offset < 0 || text[offset] != ':' {
		return annotationRange
	}
	return core.NewTextRange(offset, annotationRange.End())
}

// noInferrableTypesIsSpaceByte answers whether a byte is ASCII whitespace.
func noInferrableTypesIsSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}

// noInferrableTypesAnnotationName answers upstream's `isInferrable`, and the type name to report.
//
// The name is returned rather than derived by the caller because upstream reads it from the same
// switch that decides inferrability -- its `keywordMap`, plus the literal string "RegExp" for the
// one type-reference arm it accepts.
func noInferrableTypesAnnotationName(
	annotation *ast.Node,
	initializer *ast.Node,
) (string, bool) {
	switch annotation.Kind {
	case ast.KindBigIntKeyword:
		// A bigint may be written `-1n` but not `+1n`, which is a TypeScript error, so only the
		// minus is unwrapped. Upstream says so at the line.
		unwrapped := noInferrableTypesSkipUnary(initializer, ast.KindMinusToken)
		return "bigint", noInferrableTypesIsCallTo(unwrapped, "BigInt") ||
			noInferrableTypesIsAnyLiteral(unwrapped)

	case ast.KindBooleanKeyword:
		return "boolean", noInferrableTypesIsUnary(initializer, ast.KindExclamationToken) ||
			noInferrableTypesIsCallTo(initializer, "Boolean") ||
			noInferrableTypesIsBooleanLiteral(initializer)

	case ast.KindNumberKeyword:
		unwrapped := noInferrableTypesSkipUnary(initializer, ast.KindPlusToken, ast.KindMinusToken)
		return "number", noInferrableTypesIsIdentifierNamed(unwrapped, "Infinity", "NaN") ||
			noInferrableTypesIsCallTo(unwrapped, "Number") ||
			unwrapped.Kind == ast.KindNumericLiteral

	case ast.KindLiteralType:
		// `null` in TYPE position is a literal type wrapping the keyword here, not a bare keyword
		// node -- measured, and the reason a kind check against KindNullKeyword alone is silent on
		// `const a: null = null`, which upstream reports and repairs.
		//
		// Only the null literal type is upstream's; every other literal type (`const a: 5 = 5`) has
		// no arm there and stays clean, which is why the inner kind is checked rather than the
		// wrapper accepted outright.
		literal := annotation.AsLiteralTypeNode()
		if literal == nil || literal.Literal == nil || literal.Literal.Kind != ast.KindNullKeyword {
			return "", false
		}
		// Upstream tests `init.value == null` on a Literal, which in ESTree is true for the null
		// literal only -- `undefined` is an Identifier there, not a Literal.
		return "null", initializer.Kind == ast.KindNullKeyword

	case ast.KindStringKeyword:
		return "string", noInferrableTypesIsCallTo(initializer, "String") ||
			initializer.Kind == ast.KindStringLiteral ||
			noInferrableTypesIsTemplateLiteral(initializer)

	case ast.KindSymbolKeyword:
		return "symbol", noInferrableTypesIsCallTo(initializer, "Symbol")

	case ast.KindTypeReference:
		// The only reference upstream accepts is RegExp, and its own comment marks the arm as the
		// place more would go.
		reference := annotation.AsTypeReferenceNode()
		if reference == nil || reference.TypeName == nil ||
			reference.TypeName.Kind != ast.KindIdentifier ||
			reference.TypeName.Text() != "RegExp" {
			return "", false
		}
		return "RegExp", initializer.Kind == ast.KindRegularExpressionLiteral ||
			noInferrableTypesIsCallTo(initializer, "RegExp") ||
			noInferrableTypesIsNewOf(initializer, "RegExp")

	case ast.KindUndefinedKeyword:
		return "undefined", noInferrableTypesIsUnary(initializer, ast.KindVoidKeyword) ||
			noInferrableTypesIsIdentifierNamed(initializer, "undefined")
	}
	return "", false
}

// noInferrableTypesSkipUnary returns a unary expression's operand when the operator is one of the
// given ones, and the node itself otherwise.
func noInferrableTypesSkipUnary(node *ast.Node, operators ...ast.Kind) *ast.Node {
	if node.Kind != ast.KindPrefixUnaryExpression {
		return node
	}
	unary := node.AsPrefixUnaryExpression()
	if unary == nil {
		return node
	}
	for _, operator := range operators {
		if unary.Operator == operator {
			return unary.Operand
		}
	}
	return node
}

// noInferrableTypesIsUnary answers whether a node is a prefix unary with the given operator.
//
// `void` is a `KindVoidExpression` here rather than a unary operator, so it is answered separately.
func noInferrableTypesIsUnary(node *ast.Node, operator ast.Kind) bool {
	if operator == ast.KindVoidKeyword {
		return node.Kind == ast.KindVoidExpression
	}
	if node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	unary := node.AsPrefixUnaryExpression()
	return unary != nil && unary.Operator == operator
}

// noInferrableTypesIsCallTo answers whether a node calls a named identifier.
//
// Upstream's `isFunctionCall` first applies `skipChainExpression`, which unwraps ESTree's
// `ChainExpression` wrapper around an optional call. Here optionality is a token on the call itself
// rather than a wrapping node, so there is nothing to unwrap and the guard collapses.
func noInferrableTypesIsCallTo(node *ast.Node, name string) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return false
	}
	return call.Expression.Kind == ast.KindIdentifier && call.Expression.Text() == name
}

// noInferrableTypesIsNewOf answers whether a node is `new Name(...)`.
func noInferrableTypesIsNewOf(node *ast.Node, name string) bool {
	if node.Kind != ast.KindNewExpression {
		return false
	}
	expression := node.AsNewExpression()
	if expression == nil || expression.Expression == nil {
		return false
	}
	return expression.Expression.Kind == ast.KindIdentifier && expression.Expression.Text() == name
}

// noInferrableTypesIsIdentifierNamed answers whether a node is an identifier with one of the names.
func noInferrableTypesIsIdentifierNamed(node *ast.Node, names ...string) bool {
	if node.Kind != ast.KindIdentifier {
		return false
	}
	text := node.Text()
	for _, name := range names {
		if text == name {
			return true
		}
	}
	return false
}

// noInferrableTypesIsBooleanLiteral answers whether a node is `true` or `false`.
func noInferrableTypesIsBooleanLiteral(node *ast.Node) bool {
	return node.Kind == ast.KindTrueKeyword || node.Kind == ast.KindFalseKeyword
}

// noInferrableTypesIsTemplateLiteral answers whether a node is a template literal.
//
// Both forms count: a template with no substitutions is its own kind here.
func noInferrableTypesIsTemplateLiteral(node *ast.Node) bool {
	return node.Kind == ast.KindTemplateExpression ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

// noInferrableTypesIsAnyLiteral answers upstream's bare `init.type === Literal` test in the bigint
// arm.
//
// ESTree collapses every primitive literal into one `Literal` node, so that test accepts a number, a
// string, a boolean, a regex and a bigint alike. Here each is its own kind, so the set is written
// out. It is wider than a reader expects -- `const a: bigint = 'x'` satisfies it -- and that is
// upstream's behaviour rather than a transcription slip, so it is reproduced with this note instead
// of narrowed.
func noInferrableTypesIsAnyLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindStringLiteral,
		ast.KindRegularExpressionLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return true
	}
	return false
}

// noInferrableTypesHasModifier answers whether a declaration carries a modifier keyword.
func noInferrableTypesHasModifier(node *ast.Node, kind ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == kind {
			return true
		}
	}
	return false
}
