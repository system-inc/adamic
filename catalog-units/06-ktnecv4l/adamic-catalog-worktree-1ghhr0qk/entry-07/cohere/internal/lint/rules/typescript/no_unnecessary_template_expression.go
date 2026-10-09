package typescript

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnnecessaryTemplateExpression flags a template interpolation that adds nothing.
//
//	valid:   `${a}${b}`                      two interpolations, neither removable alone
//	valid:   `${1}`                          a number is not already a string
//	valid:   tag`${'literal'}`               a tagged template is the tag's business
//	invalid: `${'literal'}`                  a string literal interpolated into a template
//	invalid: `head${'mid'}tail`
//	invalid: `${s}` where s is a string      the whole template is just `s`
//
// Interpolating something that is already a string produces the same string with more syntax, and
// the reader has to check whether the interpolation was doing conversion work before concluding it
// was not.
//
// # The refusals are the rule, and being cleverer than upstream here changes behaviour
//
// A number, a union, `any`, an enum member, a type parameter: all declined, because stringifying
// them is real work whose result differs from their source text. `${1}` is not `1`; it is the
// string "1", and the template is what performs that conversion. The rule only removes an
// interpolation when the text it holds means the same thing spliced in raw.
//
// Two more refusals that read as omissions and are not:
//
//	a COMMENT between the quasis     declined, because the fix would delete it
//	a whitespace literal before a
//	  newline in the next quasi      declined, because `${'    '}\n` is how a trailing space is
//	                                 written deliberately; the template is protecting it from a
//	                                 trim-trailing-whitespace editor
//
// # Two anchors, because a template literal TYPE is the same shape
//
// A template value and a template literal TYPE are separate nodes with nearly the same
// type side declines a type parameter and an enum member where the value side does not, because
// those stringify differently in a type position.
//
// # The fixer is the dangerous half, and it is all about accidental `${`
//
// Removing an interpolation splices its text into the surrounding quasis, and two adjacent pieces
// that were harmless apart can form `${` together. Upstream walks the interpolations in REVERSE
// carrying one bit of state, `nextCharacterIsOpeningCurlyBrace`, so each removal knows whether what
// follows it now begins with `{`. Where a `$` would meet that `{`, the `$` is escaped.
//
// Three sites need it, and they are three because the `$` can come from three places:
//
//	the literal's own text        `${'...$'}{...`   escape the trailing `$` in the spliced value
//	a nested template's last
//	  quasi                       `${`... $`}${'{'} escape inside the nested template
//	the preceding quasi           `... $${'{...'}`  escape the `$` already in the source
//
// The escaping is "an even number of preceding backslashes", not "any backslash": `\$` is already
// escaped and must be left alone, while `\\$` ends in a real `$` and must be escaped again. That
// distinction is upstream's regex and it is reproduced rather than simplified.
//
// # Where the offsets come from, and why upstream's do not transfer
//
// Upstream's ESTree gives parallel `quasis` and `expressions` arrays and computes every range by
// arithmetic on them. Our parser gives a HEAD plus a list of SPANS, where each span carries its
// expression and the literal that FOLLOWS it. Probed:
//
//	`head${'mid'}tail`   head=[head] span0={expr:'mid', literal:tail, kind:TemplateTail}
//	`${1}${2}`           head=[]     span0={expr:1, literal:"", kind:TemplateMiddle}
//	                                 span1={expr:2, literal:"", kind:TemplateTail}
//
// So upstream's `prevQuasi` for span i is the head when i is zero and span i-1's literal otherwise,
// and its `nextQuasi` is span i's own literal. Every range below is derived from that mapping
// rather than from upstream's numbers.
//
// The raw text matters and is recoverable: a head spans from its backtick through `${`, a middle
// from `}` through `${`, and a tail from `}` through the closing backtick, so the body is the
// interior of those delimiters. Cooked text would silently differ on every escape sequence, since
// `a\nb` is four raw characters and three cooked ones. Measured both ways.
//
// # Cost
//
// Two rare anchors, and the checker is consulted only for the single-interpolation case.
var NoUnnecessaryTemplateExpression = rule.Rule{
	Name: "@typescript-eslint/no-unnecessary-template-expression",

	// Only the single-interpolation case asks a type question, but it is the one that decides
	// whether `${s}` collapses, so the checker is required.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// NeedsTypeChecker governs registration; the harness builds a Context by hand and can hand
		// this rule a nil checker. Declining the file once here is cheaper than a guard per node.
		if ctx.TypeChecker == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindTemplateExpression: func(node *ast.Node) {
				// A tagged template's interpolations are arguments to the tag function, which can
				// do anything with them, so removing one is not a rewrite of the same meaning.
				if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
					return
				}
				reportUnnecessaryTemplateParts(ctx, node, false)
			},
			ast.KindTemplateLiteralType: func(node *ast.Node) {
				reportUnnecessaryTemplateParts(ctx, node, true)
			},
		}
	},
}

// templatePart is one interpolation together with the quasis on either side of it.
//
// This is upstream's `InterpolationInfo`, rebuilt from our head-and-spans shape. `previousLiteral`
// is nil for the first part, where the head plays that role.
type templatePart struct {
	interpolation *ast.Node
	previousRaw   string
	nextRaw       string

	// The three boundaries the fixer needs, all in source offsets.
	//
	//	openStart   the `$` of this part's `${`
	//	closeEnd    one past the `}` that ends this part
	//	previousEnd one past the previous quasi's raw text, which is where its `${` begins
	openStart   int
	closeEnd    int
	previousEnd int
}

// templatePartsOf rebuilds upstream's interpolation list from our template node.
//
// Handles both a template expression and a template literal type: the two node kinds carry the same
// head-plus-spans shape under different accessors, and the rest of the rule is identical.
func templatePartsOf(ctx rule.Context, node *ast.Node, isType bool) (head *ast.Node, parts []templatePart) {
	var spans []*ast.Node
	if isType {
		literalType := node.AsTemplateLiteralTypeNode()
		head = literalType.Head
		if literalType.TemplateSpans != nil {
			spans = literalType.TemplateSpans.Nodes
		}
	} else {
		expression := node.AsTemplateExpression()
		head = expression.Head
		if expression.TemplateSpans != nil {
			spans = expression.TemplateSpans.Nodes
		}
	}
	if head == nil || len(spans) == 0 {
		return head, nil
	}

	previous := head
	for _, span := range spans {
		var interpolation, literal *ast.Node
		if isType {
			typeSpan := span.AsTemplateLiteralTypeSpan()
			interpolation, literal = typeSpan.Type, typeSpan.Literal
		} else {
			valueSpan := span.AsTemplateSpan()
			interpolation, literal = valueSpan.Expression, valueSpan.Literal
		}
		if interpolation == nil || literal == nil {
			return head, nil
		}

		previousEnd := quasiRawRange(ctx, previous).End()
		parts = append(parts, templatePart{
			interpolation: interpolation,
			previousRaw:   quasiRawText(ctx, previous),
			nextRaw:       quasiRawText(ctx, literal),
			// The `${` sits immediately after the previous quasi's raw text, and the `}` sits
			// immediately before this literal's raw text.
			openStart:   previousEnd,
			closeEnd:    quasiRawRange(ctx, literal).Pos(),
			previousEnd: previousEnd,
		})
		previous = literal
	}
	return head, parts
}

// quasiRawRange answers the span of a quasi's RAW body, excluding its delimiters.
//
// A head runs from its backtick through `${`, a middle from `}` through `${`, and a tail from `}`
// through the closing backtick. So the body always starts one character in, and ends two characters
// back for a head or middle and one for a tail or a no-substitution literal.
//
// `TokenRange` rather than `Pos()`, because `Pos()` includes leading trivia and would put the start
// before the backtick. Measured: the head of a template preceded by a space reported its raw body
// as "`x" until the trim was added.
func quasiRawRange(ctx rule.Context, quasi *ast.Node) core.TextRange {
	start := rule.TokenRange(ctx.SourceFile, quasi).Pos()
	trailer := 1
	switch quasi.Kind {
	case ast.KindTemplateHead, ast.KindTemplateMiddle:
		trailer = 2
	}
	end := quasi.End() - trailer
	if end < start+1 {
		end = start + 1
	}
	return core.NewTextRange(start+1, end)
}

// quasiRawText answers a quasi's raw source text.
//
// Raw rather than cooked, because every escaping decision below counts backslashes and a cooked
// value has already consumed them. `a\nb` is four raw characters and three cooked ones.
func quasiRawText(ctx rule.Context, quasi *ast.Node) string {
	textRange := quasiRawRange(ctx, quasi)
	text := ctx.SourceFile.Text()
	if textRange.Pos() < 0 || textRange.End() > len(text) || textRange.Pos() > textRange.End() {
		return ""
	}
	return text[textRange.Pos():textRange.End()]
}

// reportUnnecessaryTemplateParts is the shared body behind both anchors.
func reportUnnecessaryTemplateParts(ctx rule.Context, node *ast.Node, isType bool) {
	head, parts := templatePartsOf(ctx, node, isType)
	if head == nil || len(parts) == 0 {
		return
	}

	// The single-interpolation case collapses the WHOLE template to its interpolation, which is a
	// different repair from removing one part, so it is decided first and returns.
	// Upstream returns ONLY when this reports. When the type test fails it falls through to the
	// general filter below, which is how `${1}` reports at all: a number is not string-like, so the
	// whole-template collapse declines, and then the ordinary literal rule catches it. Returning
	// unconditionally here loses eighteen of upstream's reporting cases, all of them a lone
	// non-string literal, and every one of those was silent in the first draft.
	if len(parts) == 1 && parts[0].previousRaw == "" && parts[0].nextRaw == "" &&
		!commentsBetweenQuasis(ctx, parts[0]) {
		if singleInterpolationCollapses(ctx, parts[0].interpolation, isType) {
			reportSingleInterpolation(ctx, node, parts[0])
			return
		}
	}

	unnecessary := parts[:0:0]
	for _, part := range parts {
		if isUnnecessaryInterpolation(ctx, part, isType) {
			unnecessary = append(unnecessary, part)
		}
	}
	if len(unnecessary) == 0 {
		return
	}
	reportTemplateParts(ctx, unnecessary)
}

// singleInterpolationCollapses answers whether `${x}` alone can become x.
//
// The value side wants the constrained type to be string-like. The type side additionally declines
// a type parameter and an enum member, because both stringify differently in a type position than
// their source text suggests.
func singleInterpolationCollapses(ctx rule.Context, interpolation *ast.Node, isType bool) bool {
	interpolationType := ctx.TypeChecker.GetTypeAtLocation(interpolation)
	if interpolationType == nil {
		return false
	}
	constraintType, isTypeParameter := type_checking.GetConstraintInfo(ctx.TypeChecker, interpolationType)
	if constraintType == nil || !isStringLikeType(constraintType) {
		return false
	}
	if isType && (isTypeParameter || isEnumMemberType(constraintType)) {
		return false
	}
	return true
}

// isStringLikeType is upstream's `isStringLike`: every union constituent must have some
// intersection constituent carrying the string-like flag.
//
// Written here rather than added to the shared shelf, because five agents are editing that package
// right now and this is four lines over a shimmed flag.
func isStringLikeType(subject *checker.Type) bool {
	for unionPart := range type_checking.UnionTypePartsSeq(subject) {
		matched := false
		for _, part := range type_checking.IntersectionTypeParts(unionPart) {
			if type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLike) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// isEnumMemberType answers whether any constituent is an enum member.
//
// An enum member's text in a type position is its qualified name rather than its value, so
// splicing it in raw changes what the type says.
func isEnumMemberType(subject *checker.Type) bool {
	for unionPart := range type_checking.UnionTypePartsSeq(subject) {
		for _, part := range type_checking.IntersectionTypeParts(unionPart) {
			symbol := checker.Type_symbol(part)
			if symbol != nil && symbol.ValueDeclaration != nil &&
				symbol.ValueDeclaration.Kind == ast.KindEnumMember {
				return true
			}
		}
	}
	return false
}

// isUnnecessaryInterpolation decides one part, on syntax alone.
//
// No checker call: past the single-interpolation case upstream judges only literals, nested
// templates, and three specific identifiers, all of which are the same spliced in raw.
func isUnnecessaryInterpolation(ctx rule.Context, part templatePart, isType bool) bool {
	if commentsBetweenQuasis(ctx, part) {
		return false
	}

	interpolation := part.interpolation
	if isType {
		// A literal TYPE wraps its literal, and the null and undefined keywords are their own
		// nodes rather than literals.
		if interpolation.Kind == ast.KindLiteralType {
			interpolation = interpolation.AsLiteralTypeNode().Literal
		} else if interpolation.Kind == ast.KindNullKeyword ||
			interpolation.Kind == ast.KindUndefinedKeyword {
			return true
		}
		if interpolation == nil {
			return false
		}
	} else if isFixableIdentifier(interpolation) {
		// `undefined`, `Infinity` and `NaN` stringify to their own spelling, so they splice in
		// unchanged. No other identifier does, which is why this is a name test rather than a type
		// test: a variable named `x` holding "x" is not the same as the text `x`.
		return true
	}

	switch {
	case isTemplateLiteralKind(interpolation):
		// A nested template with a single whitespace quasi before a newline is the deliberate
		// trailing-space shape, so it survives.
		if startsWithNewLine(part.nextRaw) {
			return !(nestedTemplateIsSingleWhitespaceQuasi(ctx, interpolation))
		}
		return true
	case isPlainLiteral(interpolation):
		if startsWithNewLine(part.nextRaw) {
			return !(interpolation.Kind == ast.KindStringLiteral &&
				isAllWhitespace(interpolation.Text()))
		}
		return true
	}
	return false
}

// isFixableIdentifier answers upstream's three by-name identifiers.
func isFixableIdentifier(node *ast.Node) bool {
	if !ast.IsIdentifier(node) {
		return false
	}
	switch node.Text() {
	case "undefined", "Infinity", "NaN":
		return true
	}
	return false
}

// isPlainLiteral answers whether a node is one of the literal kinds upstream's `isLiteral` accepts.
func isPlainLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// isTemplateLiteralKind answers whether a node is a template, of either shape.
func isTemplateLiteralKind(node *ast.Node) bool {
	return node.Kind == ast.KindTemplateExpression ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

// nestedTemplateIsSingleWhitespaceQuasi answers whether a nested template is exactly one whitespace
// quasi, which is the shape upstream protects before a newline.
func nestedTemplateIsSingleWhitespaceQuasi(ctx rule.Context, node *ast.Node) bool {
	if node.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return false
	}
	return isAllWhitespace(node.Text())
}

// commentsBetweenQuasis answers whether a comment sits inside this part's `${ ... }`.
//
// The fix deletes that whole span, so a comment in it would be deleted with it. Upstream declines
// rather than moving the comment, and so does this.
// The scan starts at the INTERPOLATION's Pos rather than at the `${`. `Pos()` includes leading
// trivia, so it sits just after the `${` where a leading comment actually is, and the scanner reads
// trivia from there. Handed the `$` instead it reads the token text and finds nothing: measured, it
// returned zero comments for twelve of upstream's passing cases, every one of which then reported.
//
// A trailing comment (`${ 'bar' /* after */ }`) is not leading trivia of the interpolation, so it
// needs its own scan from the interpolation's END.
func commentsBetweenQuasis(ctx rule.Context, part templatePart) bool {
	interpolationStart := rule.TokenRange(ctx.SourceFile, part.interpolation).Pos()
	if type_checking.HasCommentsInRange(ctx.SourceFile,
		core.NewTextRange(part.interpolation.Pos(), interpolationStart)) {
		return true
	}
	return commentTextBetween(ctx, part.interpolation.End(), part.closeEnd)
}

// commentTextBetween answers whether a comment delimiter appears in a span of source.
//
// A trailing comment sits between the interpolation's end and the closing `}`, which is not
// anybody's leading trivia, so the scanner-based helper cannot see it. The span is short and
// contains only whitespace and comments, so looking for the delimiters directly is exact here
// rather than a heuristic.
func commentTextBetween(ctx rule.Context, start int, end int) bool {
	text := ctx.SourceFile.Text()
	if start < 0 || end > len(text) || start >= end {
		return false
	}
	between := text[start:end]
	return strings.Contains(between, "//") || strings.Contains(between, "/*")
}

// reportSingleInterpolation reports the whole-template collapse.
//
// The finding points at `${x}` including its delimiters, and the repair replaces the entire
// template with the interpolation's own text.
func reportSingleInterpolation(ctx rule.Context, node *ast.Node, part templatePart) {
	interpolationText := movedInterpolationText(ctx, node, part.interpolation)
	templateRange := rule.TokenRange(ctx.SourceFile, node)

	// Upstream anchors this finding two characters BEFORE the interpolation and one after it,
	// rather than on the `${` and `}` themselves. With no whitespace those coincide; with
	// `${    'a'    }` they do not, and the reported span is `  'a' ` rather than the whole
	// interpolation. Three corpus cases pin the difference and it is invisible without them.
	interpolationStart := rule.TokenRange(ctx.SourceFile, part.interpolation).Pos()
	ctx.Report(rule.Diagnostic{
		Range:      core.NewTextRange(interpolationStart-2, part.interpolation.End()+1),
		Message:    unnecessaryTemplateExpressionMessage(),
		SourceFile: ctx.SourceFile,
		Fixes: []rule.Fix{
			rule.ReplaceRange(core.NewTextRange(templateRange.Pos(), node.End()), interpolationText),
		},
	})
}

// literalValueText answers `String(literal.value)` for the literal kinds this rule splices.
//
// Our parser has already done the numeric work: a NumericLiteral's Text() is the canonical double
// rendering, which is a bijection onto the value upstream stringifies. Everything else is its own
// source text, except a bigint, whose Text() keeps the `n` suffix that String() drops.
func literalValueText(node *ast.Node, raw string) string {
	switch node.Kind {
	case ast.KindNumericLiteral:
		return node.Text()
	case ast.KindBigIntLiteral:
		return strings.TrimSuffix(node.Text(), "n")
	}
	return raw
}

// movedInterpolationText is upstream's `getMovedNodeCode`: the interpolation's own text, wrapped in
// parentheses when moving it out of the template could change how it binds.
//
// The template was an atom in its old position; its contents are not. Collapsing a template whose
// interpolation is a logical expression, where the template was the receiver of a method call,
// reparses with the call bound to the right operand alone. That is a different program which still
// compiles, and one corpus fix vector catches it.
//
// Two questions, both upstream's. A strong-precedence node never needs parentheses because its
// binding cannot change. Otherwise the DESTINATION decides: only a parent that could bind more
// tightly than what is moving in needs them.
func movedInterpolationText(ctx rule.Context, template *ast.Node, interpolation *ast.Node) string {
	code := ctx.SourceFile.Text()[rule.TokenRange(ctx.SourceFile, interpolation).Pos():interpolation.End()]
	if type_checking.IsStrongPrecedenceNode(interpolation) {
		return code
	}
	if !hasWeakPrecedenceParent(template) {
		return code
	}
	return "(" + code + ")"
}

// hasWeakPrecedenceParent is upstream's `isWeakPrecedenceParent`.
//
// The receiver-position tests matter: `x.y` binds its OBJECT tightly but not its property, and a
// call binds its callee tightly but not its arguments, so the same parent answers differently
// depending on which child the template is.
func hasWeakPrecedenceParent(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindBinaryExpression, ast.KindConditionalExpression,
		ast.KindAwaitExpression, ast.KindTypeOfExpression, ast.KindVoidExpression,
		ast.KindDeleteExpression:
		return true
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression == node
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression == node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression == node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression == node
	case ast.KindTaggedTemplateExpression:
		return parent.AsTaggedTemplateExpression().Tag == node
	}
	return false
}

// evenBackslashPrefix matches a position preceded by an even number of backslashes.
//
// Go's regexp has no lookbehind, so upstream's `(?<!(?:[^\\]|^)(?:\\\\)*\\)` cannot be transcribed.
// The count is done directly instead, which is what that assertion means: a `$` or a backtick is
// already escaped exactly when the run of backslashes before it is odd.
func hasEvenBackslashRunBefore(text string, index int) bool {
	count := 0
	for position := index - 1; position >= 0 && text[position] == '\\'; position-- {
		count++
	}
	return count%2 == 0
}

// endsWithUnescapedDollarSign answers upstream's helper of the same name.
func endsWithUnescapedDollarSign(text string) bool {
	if !strings.HasSuffix(text, "$") {
		return false
	}
	return hasEvenBackslashRunBefore(text, len(text)-1)
}

// escapeTemplateSpecials escapes every backtick and `${` not already escaped.
//
// Spliced text stops being a string literal and becomes template body, where a backtick ends the
// template and `${` opens an interpolation. Both have to be neutralised, and only where they are
// not already escaped. An escaped backtick is left alone, while a backtick preceded by an even
// number of backslashes is a real one and gains an escape.
func escapeTemplateSpecials(text string) string {
	var builder strings.Builder
	for index := 0; index < len(text); index++ {
		character := text[index]
		if character == '`' && hasEvenBackslashRunBefore(text, index) {
			builder.WriteString("\\`")
			continue
		}
		if character == '$' && index+1 < len(text) && text[index+1] == '{' &&
			hasEvenBackslashRunBefore(text, index) {
			builder.WriteString("\\${")
			index++
			continue
		}
		builder.WriteByte(character)
	}
	return builder.String()
}

var trailingDollarSign = regexp.MustCompile(`\$$`)

// reportTemplateParts emits one finding per removable interpolation.
//
// Upstream walks the parts in REVERSE, and the direction is load bearing rather than incidental: it
// carries `nextCharacterIsOpeningCurlyBrace` backwards so each removal knows whether the text that
// will follow it begins with `{`. A `$` meeting that `{` forms a new interpolation in the repaired
// source, which is a parse change rather than a formatting one.
func reportTemplateParts(ctx rule.Context, parts []templatePart) {
	text := ctx.SourceFile.Text()
	nextCharacterIsOpeningCurlyBrace := false

	type pendingReport struct {
		findingRange core.TextRange
		fixes        []rule.Fix
	}
	var pending []pendingReport

	for index := len(parts) - 1; index >= 0; index-- {
		part := parts[index]
		var fixes []rule.Fix

		if part.nextRaw != "" {
			nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(part.nextRaw, "{")
		}

		interpolation := part.interpolation
		if interpolation.Kind == ast.KindLiteralType {
			interpolation = interpolation.AsLiteralTypeNode().Literal
		}

		switch {
		case isPlainLiteral(interpolation) || isFixableIdentifier(interpolation):
			literalRange := rule.TokenRange(ctx.SourceFile, interpolation)
			raw := text[literalRange.Pos():interpolation.End()]

			var spliced string
			if interpolation.Kind == ast.KindStringLiteral {
				// Already a string, so the quotes come off and the interior is the value.
				spliced = raw[1 : len(raw)-1]
			} else {
				// Upstream splices `String(literal.value)`, which is the literal's VALUE rather
				// than its source text, and the two differ for every non-decimal number: `0o25`
				// stringifies to "21", `0b1010` to "10", `1_0` to "10", `1e21` to "1e+21". Four
				// corpus fix vectors caught the raw-text version.
				//
				// No numeric parser is needed: probed, our NumericLiteral's Text() already holds
				// exactly that canonical rendering. BigInt keeps its `n` in Text() and loses it in
				// String(), so that one suffix is stripped. A regular expression and the keyword
				// literals have no value distinct from their text.
				spliced = literalValueText(interpolation, raw)
				// Upstream doubles every backslash, which matters for a regular expression whose
				// source carries them.
				spliced = strings.ReplaceAll(spliced, `\`, `\\`)
			}
			spliced = escapeTemplateSpecials(spliced)

			// `...${'...$'}{...` would splice a `$` directly onto a `{`.
			if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(spliced) {
				spliced = trailingDollarSign.ReplaceAllString(spliced, `\$`)
			}
			if len(spliced) != 0 {
				nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(spliced, "{")
			}

			fixes = append(fixes, rule.ReplaceRange(
				core.NewTextRange(literalRange.Pos(), interpolation.End()), spliced))

		case isTemplateLiteralKind(interpolation):
			nestedRange := rule.TokenRange(ctx.SourceFile, interpolation)
			lastQuasiRaw := nestedTemplateLastQuasiRaw(ctx, interpolation)

			// The nested template's own last quasi can end in `$`, which would meet a following
			// `{` once the backticks come off.
			if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(lastQuasiRaw) {
				escapeAt := interpolation.End() - 2
				fixes = append(fixes, rule.ReplaceRange(
					core.NewTextRange(escapeAt, escapeAt), `\`))
			}
			if firstRaw, single := nestedTemplateSingleQuasiRaw(ctx, interpolation); single && firstRaw != "" {
				nextCharacterIsOpeningCurlyBrace = strings.HasPrefix(firstRaw, "{")
			}

			// Strip the nested template's own backticks; its body is already template body.
			fixes = append(fixes,
				rule.RemoveRange(core.NewTextRange(nestedRange.Pos(), nestedRange.Pos()+1)),
				rule.RemoveRange(core.NewTextRange(interpolation.End()-1, interpolation.End())))

		default:
			nextCharacterIsOpeningCurlyBrace = false
		}

		// `... $${'{...'} ...` and the `$` already sitting in the preceding quasi..
		if nextCharacterIsOpeningCurlyBrace && endsWithUnescapedDollarSign(part.previousRaw) {
			// The `$` sits one character before the `${` that opens this part.
			escapeAt := part.previousEnd - 1
			fixes = append(fixes, rule.ReplaceRange(
				core.NewTextRange(escapeAt, escapeAt+1), `\$`))
		}

		interpolationRange := rule.TokenRange(ctx.SourceFile, part.interpolation)
		fixes = append(fixes,
			// Remove `${` and anything between it and the interpolation.
			rule.RemoveRange(core.NewTextRange(part.openStart, interpolationRange.Pos())),
			// Remove anything between the interpolation and its `}`, and the `}` itself.
			rule.RemoveRange(core.NewTextRange(part.interpolation.End(), part.closeEnd)))

		pending = append(pending, pendingReport{
			findingRange: core.NewTextRange(part.openStart, part.closeEnd),
			fixes:        fixes,
		})
	}

	// Emitted in source order, because the reverse walk exists to compute the escaping rather than
	// to decide the reporting order.
	for index := len(pending) - 1; index >= 0; index-- {
		ctx.Report(rule.Diagnostic{
			Range:      pending[index].findingRange,
			Message:    unnecessaryTemplateExpressionMessage(),
			SourceFile: ctx.SourceFile,
			Fixes:      pending[index].fixes,
		})
	}
}

// nestedTemplateLastQuasiRaw answers the raw text of a nested template's final quasi.
func nestedTemplateLastQuasiRaw(ctx rule.Context, node *ast.Node) string {
	if node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return quasiRawText(ctx, node)
	}
	expression := node.AsTemplateExpression()
	if expression.TemplateSpans == nil || len(expression.TemplateSpans.Nodes) == 0 {
		return quasiRawText(ctx, expression.Head)
	}
	spans := expression.TemplateSpans.Nodes
	return quasiRawText(ctx, spans[len(spans)-1].AsTemplateSpan().Literal)
}

// nestedTemplateSingleQuasiRaw answers a nested template's only quasi, when it has exactly one.
func nestedTemplateSingleQuasiRaw(ctx rule.Context, node *ast.Node) (string, bool) {
	if node.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return "", false
	}
	return quasiRawText(ctx, node), true
}

// isAllWhitespace answers upstream's `isWhitespace`, which accepts the empty string.
func isAllWhitespace(literalText string) bool {
	return text.TrimWhitespace(literalText) == ""
}

// startsWithNewLine answers whether a quasi's raw text begins with a line break.
//
// Upstream tests the match index rather than a prefix, which is the same question asked awkwardly.
func startsWithNewLine(raw string) bool {
	return strings.HasPrefix(raw, "\n") || strings.HasPrefix(raw, "\r") ||
		strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, " ")
}

func unnecessaryTemplateExpressionMessage() rule.Message {
	return rule.Message{
		Id:          "noUnnecessaryTemplateExpression",
		Description: "Template literal expression is unnecessary and can be simplified.",
	}
}
