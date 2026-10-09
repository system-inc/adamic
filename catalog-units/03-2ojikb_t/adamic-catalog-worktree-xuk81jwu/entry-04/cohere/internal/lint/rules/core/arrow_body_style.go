package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ArrowBodyStyleMode selects which shape of arrow body the rule wants.
type ArrowBodyStyleMode string

const (
	// ArrowBodyStyleAsNeeded reports a braced body that only returns a value. The default.
	ArrowBodyStyleAsNeeded ArrowBodyStyleMode = "AsNeeded"

	// ArrowBodyStyleAlways reports every concise body, wanting braces everywhere.
	ArrowBodyStyleAlways ArrowBodyStyleMode = "Always"

	// ArrowBodyStyleNever reports every braced body, wanting braces nowhere.
	//
	// Unlike the other two this reports a body it cannot repair: a block doing real work has no
	// concise equivalent, so the finding stands and the fixer declines.
	ArrowBodyStyleNever ArrowBodyStyleMode = "Never"
)

// ArrowBodyStyleOptions configures the rule.
//
// Upstream's option surface is a two-element positional list whose second element is only legal
// beside `as-needed`: `["error", "as-needed", {"requireReturnForObjectLiteral": true}]`.
// `DecodeArrowBodyStyleOptions` reads that list, and the mode here is the string-literal union our
// conventions want rather than upstream's kebab spelling. The decision each one selects is
// identical; only the Go spelling moved.
type ArrowBodyStyleOptions struct {
	// Mode is which body shape the rule wants. Absent means AsNeeded, which is upstream's default.
	Mode ArrowBodyStyleMode

	// RequireReturnForObjectLiteral turns the AsNeeded judgment around for an object literal.
	//
	// Only meaningful under AsNeeded, matching upstream's schema, where it is the second element of
	// the list whose first element is `as-needed`. With it on, `() => ({})` is reported and wants
	// braces, and a braced body returning an object literal stops being reported.
	RequireReturnForObjectLiteral bool
}

// DecodeArrowBodyStyleOptions reads upstream's option list off the config.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so an unrecognized mode fails loudly. This rule's
// three modes disagree about every input, so a typo would silently select a fourth behaviour of
// reporting nothing at all. A second element beside `always` or `never` is refused rather than
// ignored, which is upstream's schema: its `anyOf` gives those two a one-element list.
func DecodeArrowBodyStyleOptions(list []byte) (any, error) {
	options := ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return options, err
	}

	var mode string
	if err := json.Unmarshal(elements[0], &mode); err != nil {
		return options, fmt.Errorf("arrow-body-style element 1 takes a string, got %s", elements[0])
	}
	switch mode {
	case "as-needed":
		options.Mode = ArrowBodyStyleAsNeeded
	case "always":
		options.Mode = ArrowBodyStyleAlways
	case "never":
		options.Mode = ArrowBodyStyleNever
	default:
		return options, &arrowBodyStyleModeError{mode: mode}
	}
	if len(elements) < 2 {
		return options, nil
	}
	if options.Mode != ArrowBodyStyleAsNeeded {
		return options, fmt.Errorf(
			"arrow-body-style: %q takes no second element, so %s would never be read; only "+
				"\"as-needed\" reads requireReturnForObjectLiteral", mode, elements[1])
	}

	var second struct {
		RequireReturnForObjectLiteral bool `json:"requireReturnForObjectLiteral"`
	}
	if err := rule.UnmarshalOptions(elements[1], &second); err != nil {
		return options, fmt.Errorf("arrow-body-style element 2: %w", err)
	}
	options.RequireReturnForObjectLiteral = second.RequireReturnForObjectLiteral
	return options, nil
}

// arrowBodyStyleModeError names an unrecognized mode and the three that are recognized.
type arrowBodyStyleModeError struct{ mode string }

func (e *arrowBodyStyleModeError) Error() string {
	return "arrow-body-style: unknown mode " + e.mode +
		", wanted one of as-needed, always, never"
}

// The five findings. Upstream splits the braced case four ways by what the block holds, because the
// remedy differs: an empty block wants `undefined` written out, an object literal wants parentheses,
// and a block doing real work cannot be shortened at all.
var (
	messageArrowBodyStyleUnexpectedOtherBlock = rule.Message{
		Id: "unexpectedOtherBlock",
		Description: "This arrow function wraps its body in braces while the configuration asks " +
			"for none. The block does more than return a value, so there is no concise form of " +
			"it and nothing here can be rewritten automatically. Change the body, or allow " +
			"braces.",
	}
	messageArrowBodyStyleUnexpectedEmptyBlock = rule.Message{
		Id: "unexpectedEmptyBlock",
		Description: "This arrow function has an empty braced body, which returns undefined. " +
			"Write that value after the arrow instead, so the arrow says what it produces " +
			"rather than leaving the reader to infer it from an empty block.",
	}
	messageArrowBodyStyleUnexpectedObjectBlock = rule.Message{
		Id: "unexpectedObjectBlock",
		Description: "This arrow function uses a braced body only to return an object literal. " +
			"Parenthesize the object and put it straight after the arrow: the parentheses are " +
			"what stop the leading brace being read as a block.",
	}
	messageArrowBodyStyleUnexpectedSingleBlock = rule.Message{
		Id: "unexpectedSingleBlock",
		Description: "This arrow function uses a braced body only to return a value. Put the " +
			"value straight after the arrow, so the braces and the return keyword stop standing " +
			"between the reader and what the function produces.",
	}
	messageArrowBodyStyleExpectedBlock = rule.Message{
		Id: "expectedBlock",
		Description: "This arrow function has a concise body while the configuration asks for " +
			"braces. Wrap the body in a block with an explicit return, so every arrow in the " +
			"codebase is read the same way.",
	}
)

// ArrowBodyStyle enforces one shape of arrow function body across the codebase.
//
//	valid:   var foo = () => 0;
//	valid:   var foo = () => { bar(); };
//	invalid: var foo = () => { return 0; };
//	invalid: var foo = () => { return { bar: 1 }; };
//	invalid (Always): var foo = () => 0;
//	invalid (Never):  var foo = () => { bar(); };
//
// # What the three modes decide
//
// AsNeeded, the default, reports a braced body whose whole content is one return, because that
// block is pure ceremony. Always reports every concise body. Never reports every braced body,
// including ones it cannot repair.
//
// # Fidelity to upstream's stack, obtained a different way
//
// Upstream tracks whether an `in` operator appeared, by pushing a frame per arrow function and
// setting a flag on every frame up the stack when it meets one. That exists because ESLint hands
// the rule one node at a time and it cannot look down. We have the tree, so the same question is
// asked directly of the arrow's subtree.
//
// The two are equivalent, and the equivalence is not obvious in one direction worth recording. The
// flag propagates through EVERY nested construct, not only nested arrows: an `in` inside a
// `function` expression, a class method, or a bracketed subscript sets the outer arrow's flag,
// because those push no frame of their own. So `for (var f = () => { return a[b in c] };;)` gets
// parentheses it does not need. That is upstream's behaviour, measured against the installed rule
// across six nesting shapes, and it is reproduced rather than tightened: a port that only looked at
// the top-level operator would write different source text than upstream on four of its own cases.
//
// # Why the `in` matters at all
//
// `for (var f = () => a in c;;)` does not parse the way it reads: inside a for statement's
// initializer, `in` is claimed by the `for...in` grammar, so the fixer wraps the returned value in
// parentheses whenever an `in` is anywhere inside an arrow being shortened inside a for initializer.
// Nothing else in this rule is about parsing rather than style.
var ArrowBodyStyle = rule.Rule{
	Name: "arrow-body-style",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := ArrowBodyStyleOptions{Mode: ArrowBodyStyleAsNeeded}
		if decoded, configured := rule.OptionsAs[ArrowBodyStyleOptions](options); configured {
			settings = decoded
			if settings.Mode == "" {
				// A rule configured as a bare severity is handed nil, and `options.(T)` on nil
				// yields the zero value, whose empty mode matches no arm. Upstream's default is
				// AsNeeded, so an empty mode becomes that rather than silently disabling the rule.
				settings.Mode = ArrowBodyStyleAsNeeded
			}
		}
		return rule.Listeners{
			ast.KindArrowFunction: func(node *ast.Node) {
				checkArrowBodyStyle(ctx, node, settings)
			},
		}
	},
}

// checkArrowBodyStyle judges one arrow function's body.
func checkArrowBodyStyle(ctx rule.Context, node *ast.Node, settings ArrowBodyStyleOptions) {
	body := node.AsArrowFunction().Body
	if body == nil {
		return
	}

	if body.Kind == ast.KindBlock {
		checkArrowBodyStyleBracedBody(ctx, node, body, settings)
		return
	}

	// A concise body. Wanted braced under Always, and under AsNeeded when it is an object literal
	// and the option asks for the return to be written out.
	wantsBraces := settings.Mode == ArrowBodyStyleAlways ||
		(settings.Mode == ArrowBodyStyleAsNeeded && settings.RequireReturnForObjectLiteral &&
			arrowBodyStyleUnwrapParentheses(body).Kind == ast.KindObjectLiteralExpression)
	if !wantsBraces {
		return
	}
	reportArrowBodyStyleWantingBraces(ctx, node, body)
}

// checkArrowBodyStyleBracedBody judges a `{ ... }` body.
func checkArrowBodyStyleBracedBody(
	ctx rule.Context,
	node *ast.Node,
	body *ast.Node,
	settings ArrowBodyStyleOptions,
) {
	statements := body.AsBlock().Statements
	var blockBody []*ast.Node
	if statements != nil {
		blockBody = statements.Nodes
	}

	// Under AsNeeded and Always, a block holding anything other than exactly one statement is left
	// alone: there is nothing to shorten it to. Never reports it anyway, which is the whole
	// difference between Never and the other two here.
	if len(blockBody) != 1 && settings.Mode != ArrowBodyStyleNever {
		return
	}

	// Under AsNeeded, the object-literal option EXEMPTS a block that returns one, which is the
	// mirror of the concise arm above: with the option on, that block is the shape the rule wants.
	if settings.Mode == ArrowBodyStyleAsNeeded && settings.RequireReturnForObjectLiteral &&
		len(blockBody) == 1 && blockBody[0].Kind == ast.KindReturnStatement {
		if returned := blockBody[0].AsReturnStatement().Expression; returned != nil &&
			arrowBodyStyleUnwrapParentheses(returned).Kind == ast.KindObjectLiteralExpression {
			return
		}
	}

	reportable := settings.Mode == ArrowBodyStyleNever ||
		(settings.Mode == ArrowBodyStyleAsNeeded && len(blockBody) == 1 &&
			blockBody[0].Kind == ast.KindReturnStatement)
	if !reportable {
		return
	}

	reportArrowBodyStyleWantingNoBraces(ctx, node, body, blockBody)
}

// reportArrowBodyStyleWantingNoBraces reports a braced body, with the repair when one is possible.
//
// The message is chosen by what the block holds, because the remedy differs and upstream says so in
// four different sentences. The FIX is offered only for the one shape that has a concise
// equivalent, which is a single return carrying a value; every other shape reports with no repair,
// reproducing upstream's `output: null`.
func reportArrowBodyStyleWantingNoBraces(
	ctx rule.Context,
	node *ast.Node,
	body *ast.Node,
	blockBody []*ast.Node,
) {
	message := messageArrowBodyStyleUnexpectedOtherBlock
	switch {
	case len(blockBody) == 0:
		message = messageArrowBodyStyleUnexpectedEmptyBlock
	case len(blockBody) > 1 || blockBody[0].Kind != ast.KindReturnStatement:
		message = messageArrowBodyStyleUnexpectedOtherBlock
	case blockBody[0].AsReturnStatement().Expression == nil:
		message = messageArrowBodyStyleUnexpectedSingleBlock
	case arrowBodyStyleStartsWithBrace(ctx, blockBody[0].AsReturnStatement().Expression):
		message = messageArrowBodyStyleUnexpectedObjectBlock
	default:
		message = messageArrowBodyStyleUnexpectedSingleBlock
	}

	fixes := arrowBodyStyleRemoveBracesFixes(ctx, node, body, blockBody)
	if len(fixes) == 0 {
		// Upstream's fixer returns an empty edit list rather than declining, which the engine
		// treats as no fix. Reproduced as reporting with no repair, which is what its corpus
		// records as `output: null`.
		ctx.ReportRange(rule.TokenRange(ctx.SourceFile, body), message)
		return
	}
	ctx.ReportRangeWithFixes(rule.TokenRange(ctx.SourceFile, body), message, fixes...)
}

// arrowBodyStyleRemoveBracesFixes builds the edits that turn `{ return v; }` into `v`.
//
// Returns nothing when the shape has no concise equivalent, or when removing the braces would
// change what the next line means. That second case is automatic semicolon insertion: a following
// line beginning with one of `([/` or a backtick, plus and minus would continue the expression
// once the block's closing brace stops separating them, so the repair is declined and the finding
// stands alone.
func arrowBodyStyleRemoveBracesFixes(
	ctx rule.Context,
	node *ast.Node,
	body *ast.Node,
	blockBody []*ast.Node,
) []rule.Fix {
	if len(blockBody) != 1 || blockBody[0].Kind != ast.KindReturnStatement {
		return nil
	}
	returned := blockBody[0].AsReturnStatement().Expression
	if returned == nil {
		return nil
	}
	if arrowBodyStyleHasInsertionHazardAfter(ctx, body) {
		return nil
	}

	text := ctx.SourceFile.Text()
	openingBrace := core.NewTextRange(body.Pos(), body.Pos())
	if index := strings.Index(text[body.Pos():body.End()], "{"); index >= 0 {
		openingBrace = core.NewTextRange(body.Pos()+index, body.Pos()+index+1)
	}
	closingBrace := core.NewTextRange(body.End()-1, body.End())

	valueStart := rule.TokenRange(ctx.SourceFile, returned).Pos()
	// Upstream anchors on the return statement's LAST TOKEN, which is the semicolon when one is
	// written. So `{ return a ; }` shortens to `a ` with the space kept and the semicolon dropped,
	// rather than to `a`. Ending the span at the expression instead swallows that whitespace, which
	// is a one-character difference on three of upstream's own cases and invisible to every
	// message-id fixture.
	valueEnd := returned.End()
	semicolon := arrowBodyStyleTrailingSemicolonOf(ctx, blockBody[0])

	// Parenthesize when the concise form would otherwise be misread. A leading brace would start a
	// block, a comma expression would be read as the arrow's parameter list continuing, and an `in`
	// inside a for statement's initializer belongs to the `for...in` grammar.
	unwrapped := arrowBodyStyleUnwrapParentheses(returned)
	needsParentheses := arrowBodyStyleStartsWithBrace(ctx, returned) ||
		(unwrapped.Kind == ast.KindBinaryExpression &&
			unwrapped.AsBinaryExpression().OperatorToken != nil &&
			unwrapped.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken) ||
		(arrowBodyStyleSubtreeHasInOperator(node) && arrowBodyStyleInsideForInitializer(node))
	if arrowBodyStyleIsParenthesized(returned) {
		needsParentheses = false
	}
	openParenthesis, closeParenthesis := "", ""
	if needsParentheses {
		openParenthesis, closeParenthesis = "(", ")"
	}

	// The trailing window starts at the statement's LAST token, which is the semicolon when one is
	// written, not at the end of the value. A comment between the value and its semicolon therefore
	// does NOT take the preserving branch: `{ return 5 /* c */; }` shortens through the ordinary
	// path and the comment survives because it sits inside the value-to-semicolon span that path
	// keeps. Measured against the installed rule, which the imported corpus could not show because
	// its only comment case has comments on both sides at once.
	trailingFrom := valueEnd
	if semicolon >= 0 {
		trailingFrom = semicolon + 1
	}

	// # Every parenthesis rides inside a removal, and none is its own insertion
	//
	// The engine treats two edits sharing a boundary as an overlap and drops one of them, and a
	// zero-width insertion AT the edge of a removal is exactly that shape. Emitting `(` as its own
	// edit at the value's start, beside a removal ending there, silently lost the parenthesis in
	// the real fix phase while `ExpectFixedSource` applied both and went green -- so `() => {
	// return { a: 1 }; }` was rewritten to `() => { a: 1 }`, which parses as a labelled block and
	// means something else entirely. Found by running the fix engine over a seeded tree; nothing in
	// the 87-case corpus could see it.
	//
	// So each parenthesis is carried as the replacement TEXT of the removal it sits against, which
	// makes the whole repair a set of disjoint replacements the engine cannot split up.
	var fixes []rule.Fix
	// A comment anywhere in the ceremony has to survive, so only the tokens themselves are removed
	// and the whitespace around them is left in place. Without a comment the whitespace goes too,
	// which is what makes the common case read as one clean line.
	if arrowBodyStyleCommentsBetween(ctx, openingBrace.End(), valueStart) ||
		arrowBodyStyleCommentsBetween(ctx, trailingFrom, closingBrace.Pos()) {
		returnKeyword := rule.TokenRange(ctx.SourceFile, blockBody[0])
		returnRange := core.NewTextRange(returnKeyword.Pos(), returnKeyword.Pos()+len("return"))
		fixes = append(fixes,
			rule.RemoveRange(openingBrace),
			// The opening parenthesis replaces the `return` keyword rather than being inserted
			// beside it, for the disjointness reason above.
			rule.ReplaceRange(returnRange, openParenthesis))
		if semicolon >= 0 {
			// A concise body is an expression, so the statement's semicolon has no place in it,
			// and the closing parenthesis takes the byte it vacates.
			fixes = append(fixes,
				rule.ReplaceRange(core.NewTextRange(semicolon, semicolon+1), closeParenthesis),
				rule.RemoveRange(closingBrace))
		} else {
			fixes = append(fixes, rule.ReplaceRange(closingBrace, closeParenthesis))
		}
	} else {
		// The head removal carries the opening parenthesis, and the tail removal carries the
		// closing one. The tail is anchored on the statement's last token rather than on the
		// expression, so the whitespace between a value and its semicolon survives exactly as
		// upstream leaves it: `{ return a ; }` shortens to `a ` and not to `a`.
		fixes = append(fixes,
			rule.ReplaceRange(core.NewTextRange(openingBrace.Pos(), valueStart), openParenthesis))
		if semicolon >= 0 {
			fixes = append(fixes,
				rule.ReplaceRange(core.NewTextRange(semicolon, closingBrace.End()),
					closeParenthesis))
		} else {
			fixes = append(fixes,
				rule.ReplaceRange(core.NewTextRange(valueEnd, closingBrace.End()),
					closeParenthesis))
		}
	}

	return fixes
}

// reportArrowBodyStyleWantingBraces reports a concise body and offers the block form.
//
// The repair has two shapes because a returned object literal arrives already wrapped in the
// parentheses the concise form required. Those parentheses become redundant inside a block, so they
// are consumed rather than left behind: the opening one becomes `{return ` and the closing one is
// deleted wherever it sits, which is not always the last token -- `() => ({}).foo()` closes it in
// the middle.
func reportArrowBodyStyleWantingBraces(ctx rule.Context, node *ast.Node, body *ast.Node) {
	message := messageArrowBodyStyleExpectedBlock
	text := ctx.SourceFile.Text()

	bodyStart := rule.TokenRange(ctx.SourceFile, body).Pos()
	lastToken := node.End()

	// A parenthesized object literal, whose parentheses the concise form REQUIRED and the block
	// form does not, so they are consumed rather than left behind.
	//
	// The object need not be the whole body: `() => ({foo: 1}).foo()` opens with the same forced
	// parentheses and closes them in the MIDDLE of the expression, so the closing one is found by
	// matching the opening one rather than by taking the last token. Upstream walks up from the
	// object to the nearest parenthesized ancestor and takes the token after it; this matches
	// forward from the opening paren, which is the same span reached without a parent walk.
	if object := arrowBodyStyleForcedParenthesizedObject(ctx, body); object != nil {
		openingParen := core.NewTextRange(bodyStart, bodyStart+1)
		objectStart := rule.TokenRange(ctx.SourceFile, object).Pos()
		closingParen := arrowBodyStyleMatchingCloseParen(text, bodyStart)

		var fixes []rule.Fix
		if arrowBodyStyleSameLine(text, openingParen.Pos(), objectStart) {
			fixes = append(fixes, rule.ReplaceRange(openingParen, "{return "))
		} else {
			// Splitting `{` from `return` across the newline would let automatic semicolon
			// insertion end the statement before the value, so the keyword moves down to the brace.
			fixes = append(fixes,
				rule.ReplaceRange(openingParen, "{"),
				rule.ReplaceRange(core.NewTextRange(objectStart, objectStart), "return "))
		}

		// The closing parenthesis is deleted and a closing brace is appended, and in the common
		// shape `() => ({x: 1})` those two edits MEET: the parenthesis is the last token, so its
		// removal ends exactly where the brace would be inserted. The engine reads a shared
		// boundary as an overlap and drops one of the pair, which silently threw away the entire
		// repair -- measured by running the fix phase over a seeded tree and watching nothing
		// change. Merged into one replacement when they touch, kept apart when they do not, which
		// is the `() => ({x: 1}).x` shape where the parenthesis closes mid-expression.
		switch {
		case closingParen < 0:
			fixes = append(fixes,
				rule.ReplaceRange(core.NewTextRange(lastToken, lastToken), "}"))
		case closingParen+1 == lastToken:
			fixes = append(fixes,
				rule.ReplaceRange(core.NewTextRange(closingParen, lastToken), "}"))
		default:
			fixes = append(fixes,
				rule.RemoveRange(core.NewTextRange(closingParen, closingParen+1)),
				rule.ReplaceRange(core.NewTextRange(lastToken, lastToken), "}"))
		}
		ctx.ReportRangeWithFixes(rule.TokenRange(ctx.SourceFile, body), message, fixes...)
		return
	}

	ctx.ReportRangeWithFixes(rule.TokenRange(ctx.SourceFile, body), message,
		rule.ReplaceRange(core.NewTextRange(bodyStart, bodyStart), "{return "),
		rule.ReplaceRange(core.NewTextRange(lastToken, lastToken), "}"))
}

// arrowBodyStyleHasInsertionHazardAfter reports whether the token after a body would join onto it.
//
// Upstream's `hasASIProblem`: once the closing brace stops separating two lines, a next line
// beginning with `(`, `[`, `/`, a backtick, `+` or `-` continues the expression instead of starting
// a statement. The repair is declined rather than made safe, which is upstream's choice and one of
// the two ways its corpus records a reported-but-unfixed case.
//
// # Comments are trivia, and the first byte after the body is often one
//
// Upstream asks `getTokenAfter`, which skips comments. A byte scan that stops at the first
// non-whitespace character sees the `/` of a `/* c */` and declines every repair that has a trailing
// comment -- including upstream's own case that threads nine comments through one arrow, which it
// does repair. Measured across six spellings against the installed rule: a comment never decides
// this, and the token after it always does, so `/* c */ ;` is fixed and `/* c */ + 1` is not.
func arrowBodyStyleHasInsertionHazardAfter(ctx rule.Context, body *ast.Node) bool {
	text := ctx.SourceFile.Text()
	for position := body.End(); position < len(text); position++ {
		character := text[position]
		if character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			continue
		}
		// A comment is trivia rather than the next token, so it is stepped over and the scan
		// continues. Only `/` can begin one, and the character after it separates a comment from
		// the division operator that IS a hazard.
		if character == '/' && position+1 < len(text) {
			if next := text[position+1]; next == '/' || next == '*' {
				position = arrowBodyStyleEndOfComment(text, position)
				continue
			}
		}
		return strings.IndexByte("([/`+-", character) >= 0
	}
	return false
}

// arrowBodyStyleEndOfComment returns the last offset of the comment starting at `start`.
//
// A line comment ends at the newline, which is left for the caller's loop to skip as whitespace. A
// block comment that never closes runs to the end of the file, which error recovery can produce.
func arrowBodyStyleEndOfComment(text string, start int) int {
	if text[start+1] == '/' {
		for position := start + 2; position < len(text); position++ {
			if text[position] == '\n' || text[position] == '\r' {
				return position - 1
			}
		}
		return len(text) - 1
	}
	for position := start + 2; position+1 < len(text); position++ {
		if text[position] == '*' && text[position+1] == '/' {
			return position + 1
		}
	}
	return len(text) - 1
}

// arrowBodyStyleCommentsBetween reports whether any comment sits between two offsets.
//
// `comments.ForFile` caches one scan per file, so this is a scan of an already-built list rather
// than a re-scan of the text. Reaching for `GetLeadingCommentRanges` here would have been the third
// time somebody reimplemented that shelf.
func arrowBodyStyleCommentsBetween(ctx rule.Context, from int, to int) bool {
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= from && comment.Range.End() <= to {
			return true
		}
	}
	return false
}

// arrowBodyStyleStartsWithBrace reports whether an expression's first token is `{`.
//
// This is what separates `unexpectedObjectBlock` from `unexpectedSingleBlock`, and separately what
// decides whether the shortened form needs parentheses. Asked of the TEXT rather than of the node
// kind, because upstream asks it of the token: an object literal starts with a brace, and so does
// nothing else that can be returned, but reading the first byte is the same question upstream asks
// and cannot drift from it.
func arrowBodyStyleStartsWithBrace(ctx rule.Context, expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	start := rule.TokenRange(ctx.SourceFile, expression).Pos()
	text := ctx.SourceFile.Text()
	return start < len(text) && text[start] == '{'
}

// arrowBodyStyleIsParenthesized reports whether a node is wrapped in its own parentheses.
func arrowBodyStyleIsParenthesized(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindParenthesizedExpression
}

// arrowBodyStyleUnwrapParentheses returns the expression inside any parentheses around a node.
func arrowBodyStyleUnwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		inner := node.AsParenthesizedExpression().Expression
		if inner == nil {
			return node
		}
		node = inner
	}
	return node
}

// arrowBodyStyleSameLine reports whether two offsets sit on one line.
func arrowBodyStyleSameLine(text string, from int, to int) bool {
	if to > len(text) {
		to = len(text)
	}
	return !strings.ContainsAny(text[from:to], "\n\r")
}

// arrowBodyStyleSubtreeHasInOperator reports whether an `in` appears anywhere under a node.
//
// This replaces upstream's per-arrow frame stack, and the breadth is deliberate rather than lazy.
// Upstream sets the flag on every frame up the stack, and only an ARROW pushes a frame, so an `in`
// inside a nested `function`, a class method, or a bracketed subscript reaches the enclosing arrow
// just the same. Measured across six nesting shapes against the installed rule; narrowing this to
// the top-level operator would write different text than upstream on four of its own cases.
func arrowBodyStyleSubtreeHasInOperator(node *ast.Node) bool {
	found := false
	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		if current.Kind == ast.KindBinaryExpression {
			if operator := current.AsBinaryExpression().OperatorToken; operator != nil &&
				operator.Kind == ast.KindInKeyword {
				found = true
				return
			}
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(node)
	return found
}

// arrowBodyStyleInsideForInitializer reports whether a node sits in a for statement's initializer.
//
// Only the initializer clause matters, which is the whole reason this walks rather than asking
// whether any for statement encloses the node. `for (var f = () => { return a in c };;)` needs the
// parentheses and `for (var f;f = () => { return a in c };)` does not, because the second arrow is
// in the condition clause where `in` is an ordinary operator. Both are upstream cases and they
// differ only here.
func arrowBodyStyleInsideForInitializer(node *ast.Node) bool {
	child := node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindForStatement &&
			parent.AsForStatement().Initializer == child {
			return true
		}
		child = parent
	}
	return false
}

// arrowBodyStyleTrailingSemicolonOf returns the offset of a statement's closing semicolon, or -1.
//
// A concise arrow body is an expression rather than a statement, so the semicolon that ended the
// return has no place in the shortened form and upstream removes it as its own edit.
func arrowBodyStyleTrailingSemicolonOf(ctx rule.Context, statement *ast.Node) int {
	text := ctx.SourceFile.Text()
	for position := statement.End() - 1; position >= statement.Pos(); position-- {
		switch text[position] {
		case ' ', '\t', '\n', '\r':
			continue
		case ';':
			return position
		default:
			return -1
		}
	}
	return -1
}

// arrowBodyStyleForcedParenthesizedObject returns the object literal whose parentheses the concise
// body required, or nil when the body has none.
//
// The parentheses are forced by the grammar: an arrow body beginning with `{` would be read as a
// block, so returning an object literal concisely has to wrap it. Those parentheses are exactly
// what the block form makes redundant.
//
// The object does not have to BE the body. `() => ({foo: 1}).foo()` and `() => ({foo: 1}.foo())`
// both open with the forced parentheses and both are repaired by consuming them, so this looks past
// a member access or a call to the leftmost thing the body starts with rather than testing the
// body's own kind.
func arrowBodyStyleForcedParenthesizedObject(ctx rule.Context, body *ast.Node) *ast.Node {
	if !arrowBodyStyleStartsWithOpenParen(ctx, body) {
		return nil
	}
	// Descend the leftmost spine: whatever the body evaluates to, the token after the arrow's
	// parenthesis is the start of its leftmost operand.
	current := body
	for current != nil {
		if current.Kind == ast.KindParenthesizedExpression {
			// The parentheses may hold the object directly, as in `({foo: 1})`, or hold an
			// expression whose LEFTMOST operand is the object, as in `({foo: 1}.foo())`. Both are
			// the forced-parenthesis shape and both are repaired the same way, so the spine is
			// descended again inside rather than the content being tested once.
			return arrowBodyStyleLeftmostObjectLiteral(
				current.AsParenthesizedExpression().Expression)
		}
		next := arrowBodyStyleLeftmostOperand(current)
		if next == nil || next == current {
			return nil
		}
		current = next
	}
	return nil
}

// arrowBodyStyleLeftmostOperand returns the receiver of a member access or call, or nil.
func arrowBodyStyleLeftmostOperand(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	case ast.KindCallExpression:
		return node.AsCallExpression().Expression
	}
	return nil
}

// arrowBodyStyleStartsWithOpenParen reports whether a node's first token is `(`.
func arrowBodyStyleStartsWithOpenParen(ctx rule.Context, node *ast.Node) bool {
	if node == nil {
		return false
	}
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	text := ctx.SourceFile.Text()
	return start < len(text) && text[start] == '('
}

// arrowBodyStyleMatchingCloseParen returns the offset of the `)` closing the `(` at `from`, or -1.
//
// Scanned rather than read off a node, because the parenthesis being closed may belong to a node
// that is not the body: in `() => ({foo: 1}).foo()` the closing paren sits in the middle of the
// expression and no single node's End() names it. String and template contents are skipped so a
// parenthesis inside a literal cannot be mistaken for the real one.
func arrowBodyStyleMatchingCloseParen(text string, from int) int {
	depth := 0
	for position := from; position < len(text); position++ {
		switch character := text[position]; character {
		case '\'', '"', '`':
			position = arrowBodyStyleSkipStringLiteral(text, position)
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return position
			}
		}
	}
	return -1
}

// arrowBodyStyleSkipStringLiteral returns the offset of a literal's closing quote.
//
// Returns the last index when the literal never closes, which error recovery can produce, so the
// caller's loop terminates rather than running past the end.
func arrowBodyStyleSkipStringLiteral(text string, start int) int {
	quote := text[start]
	for position := start + 1; position < len(text); position++ {
		if text[position] == '\\' {
			position++
			continue
		}
		if text[position] == quote {
			return position
		}
	}
	return len(text) - 1
}

// arrowBodyStyleLeftmostObjectLiteral returns the object literal an expression starts with, or nil.
//
// `{foo: 1}` answers itself, `{foo: 1}.foo()` answers the object, and `foo({})` answers nil because
// it starts with an identifier. This is the same descent `arrowBodyStyleForcedParenthesizedObject`
// makes outside the parentheses, applied inside them.
func arrowBodyStyleLeftmostObjectLiteral(node *ast.Node) *ast.Node {
	current := arrowBodyStyleUnwrapParentheses(node)
	for current != nil {
		if current.Kind == ast.KindObjectLiteralExpression {
			return current
		}
		next := arrowBodyStyleLeftmostOperand(current)
		if next == nil || next == current {
			return nil
		}
		current = arrowBodyStyleUnwrapParentheses(next)
	}
	return nil
}
