package comments

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// Three rules judge comments rather than nodes, and a comment is not a node: it is trivia the
// parser skips, recorded only as a range in the source text. So they cannot register a listener for
// it, and each would otherwise write its own scan. This is that scan, written once.

// Comment is one comment in a file: where it is, what kind it is, and its text.
type Comment struct {
	Range core.TextRange

	// IsBlock distinguishes a slash-star comment from a double-slash one. JSDoc is a block comment
	// whose text opens with an asterisk, so the two rules that care about JSDoc need this to tell
	// the shapes apart.
	IsBlock bool

	// Text is the full source text of the comment, delimiters included. Rules that want the inner
	// content strip the delimiters themselves, because what counts as content differs by rule: the
	// JSDoc rule wants the asterisk-prefixed lines, the line-length rule wants the whole line.
	Text string

	// StartLine, StartColumn, and EndLine are zero-based positions. The run-detecting rule needs
	// them to answer two questions no range can: whether two comments sit on adjacent lines, and
	// whether they start at the same column, which is what separates a stacked passage from a
	// trailing comment that happens to follow another.
	StartLine   int
	StartColumn int
	EndLine     int
}

// All returns every comment in a file, in source order, without duplicates.
//
// The parser skips trivia, so there is no list to read; the comments have to be rescanned out of
// the source text. The scan walks from position zero and from the start of every node, because the
// scanner's iterator yields the run of comments beginning at a position and stops at the first
// token. Positions repeat across a walk (a node and its first child usually start at the same
// place), so the results are deduplicated by range rather than assumed distinct.
// # Comments inside an otherwise-empty pair of delimiters
//
// A node's `Pos()` is where its leading trivia begins and its `End()` is past its closing token, so
// both sit *outside* the delimiters. When the thing between the delimiters is empty there is no
// child node to anchor a position inside them, and a comment sitting there was invisible to this
// scan. Measured: `switch(foo) { case 0: { /* falls through */ } case 1: b(); }` returned zero
// comments, and that input is one of upstream's clean cases for `no-fallthrough`.
//
// The gap was wider than the empty *block* it was first reported as. Nine shapes were measured
// silent, including `class A { /* c */ }`, `interface I { /* c */ }`, `enum E { /* c */ }`,
// `f( /* c */ )` and `function g(/* c */) {}`, none of which involve a block at all. What they share
// is an empty `NodeList`, not an empty block.
//
// The fix is `collectListInteriors`, which anchors on the empty list's own `Pos()`. The parser
// records that position between the delimiters whether or not the list has elements, so it is the
// interior anchor the walk was missing.
//
// It costs. Measured over 300 real files, three runs each: 5.45ms per pass before, 6.44ms after, so
// about 18%, and the whole-tree `--timing` run agrees at 95.9ms against 116ms for `comments.All`
// across 3,407 files. That is the price of a comment scan that can no longer go silent, paid once
// per file and shared by every rule that asks. The work itself is small and the checking is most of
// it: 195,800 nodes own 12,975 lists, of which only 1,547 are empty and actually reach the scanner.
//
// # Why the anchor has to come from the parser
//
// A cheaper-looking fix is to sweep the text for `/` characters the walk did not cover and scan
// from each. It was tried and it is wrong: on `const s = 'http://x'; // real` it invented a comment
// `//x'; // real` out of the `//` inside the string literal. Only the parser knows which slashes are
// code, so every anchor here is one the parser handed us. That is also why `canBeginAt` is allowed
// to be an over-approximation while this is not: the guard only ever *declines* to scan a position
// the walk already trusted, whereas a text sweep would *invent* positions the walk never saw.
func All(sourceFile *ast.SourceFile) []Comment {
	if sourceFile == nil {
		return nil
	}
	text := sourceFile.Text()

	var factory ast.NodeFactory
	seenRanges := make(map[core.TextRange]bool)
	var comments []Comment

	record := func(commentRange ast.CommentRange) {
		if seenRanges[commentRange.TextRange] {
			return
		}
		seenRanges[commentRange.TextRange] = true

		start, end := commentRange.Pos(), commentRange.End()
		if start < 0 || end > len(text) || start >= end {
			return
		}
		startLine, startColumn := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, start)
		endLine, _ := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, end)

		comments = append(comments, Comment{
			Range:       commentRange.TextRange,
			IsBlock:     commentRange.Kind == ast.KindMultiLineCommentTrivia,
			Text:        text[start:end],
			StartLine:   startLine,
			StartColumn: startColumn,
			EndLine:     endLine,
		})
	}

	// Both directions are needed, and the pair is the whole reason this helper has a coverage test.
	// A leading scan sees a comment that begins a run of trivia; a comment sitting after code on the
	// same line is trailing trivia and is invisible to it. That gap costs nothing at parse time and
	// everything at review time, because a rule silently reports nothing for the comments it never
	// received, and the fixtures pass.
	// Adjacent nodes share positions constantly: a node's Pos is often its parent's Pos and its
	// first child's Pos, so the same offset arrives many times. Measured on one real file, 55.9% of
	// the positions surviving the guard had already been scanned, and the scanner yielded 131 comment
	// ranges for 57 distinct comments, a 2.3x duplication that `record` then discarded.
	//
	// Remembering the positions already visited is exact rather than approximate, so unlike the guard
	// it cannot change what is found: scanning the same offset twice returns the same ranges.
	scannedPositions := make(map[int]bool)

	collectAt := func(position int) {
		if position < 0 || position > len(text) {
			return
		}
		// A comment must begin with a slash, so a position whose following whitespace run reaches a
		// non-slash character cannot start one. This is the lexical definition rather than an
		// approximation, which is what makes it safe to skip on.
		//
		// It matters because this runs at every node's Pos and End: 1,553 invocations to find 57
		// comments in one real file, and 96.4% of 463,463 candidate positions across 300 files were
		// measured to reach the scanner for nothing.
		if !canBeginAt(text, position) {
			return
		}
		if scannedPositions[position] {
			return
		}
		scannedPositions[position] = true
		for commentRange := range scanner.GetLeadingCommentRanges(&factory, text, position) {
			record(commentRange)
		}
		for commentRange := range scanner.GetTrailingCommentRanges(&factory, text, position) {
			record(commentRange)
		}
	}

	// Position zero catches the file's opening comments, which precede every node.
	collectAt(0)

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		// A node's Pos() is where its leading trivia begins, not where its first token does, which
		// is exactly the position the comment scanner wants. End() is where a trailing comment
		// would sit, and the two positions find different comments.
		collectAt(node.Pos())
		collectAt(node.End())
		// The interior of an empty delimiter pair, which neither position above can reach. See the
		// section on empty delimiters in this function's doc comment.
		collectListInteriors(node, collectAt)
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())

	// The end-of-file token carries any comment that trails the last statement. Without this a
	// comment after the final line is invisible, which is the failure a rule would never notice
	// because nothing reports a comment it did not see.
	collectAt(sourceFile.EndOfFileToken.Pos())

	sortByPosition(comments)
	return comments
}

// collectListInteriors offers the scan the position just inside a node's delimiters.
//
// A node's `Pos()` sits before its opening token and its `End()` sits after its closing one, so a
// comment written between the two, with nothing else between them, is reachable from neither. The
// parser already knows the position that would answer: every delimited child sequence is a
// `NodeList`, and a `NodeList` carries its own `Pos()` recorded *inside* the delimiters whether or
// not it holds any elements. `ForEachChild` flattens a list into its elements and discards the
// list, which is why the walk never saw it; this asks the nodes that own one directly.
//
// Only the empty case actually needs this. A non-empty list has a first child whose `Pos()` the
// walk already visits, and that position and the list's `Pos()` are the same. Scanning it anyway
// would be correct and wasteful, so the check below is what keeps this off the hot path: on the
// tree the cost is one map lookup per node that owns a list, and a scan only where a list is empty.
//
// The kinds enumerated here are the ones measured to hide a comment. Each was reproduced before it
// was added, rather than being written from the shape of the AST: a kind listed on a hunch would be
// untested code, and a kind that belongs here and is missing shows up as a comment nothing reports,
// which is the failure this whole file exists to prevent. Anything not listed keeps the behavior it
// had before, so a missing kind is the old gap rather than a new defect.
func collectListInteriors(node *ast.Node, collectAt func(int)) {
	offer := func(list *ast.NodeList) {
		if list == nil {
			return
		}
		// A NON-EMPTY list names one position nothing else does: its `End()`, which sits past a
		// trailing comma that the last element's own `End()` stops short of. So
		// `['a', // one\n 'b', // two\n]` loses the SECOND comment, because nothing probes between
		// the comma and the bracket.
		//
		// Measured, with a control: the same array written without a trailing comma on the last
		// element finds both comments, and with one finds only the first. On Kirk's tree this was 8
		// findings across 4 files for `no-inline-comments` before the fix, and the shape is common
		// in hand-maintained lists.
		if len(list.Nodes) > 0 {
			collectAt(list.End())
			return
		}
		// `Pos()` and `End()` are interchangeable here and a mutation swapping them survives the
		// sweep, correctly: the early return above means only empty lists reach this line, and an
		// empty list spans zero characters, so the two are the same number. Measured rather than
		// argued, across 2,407 empty lists in 600 real files, zero of which differed. `Pos()` is
		// written because it is the position being asked for, not because it is the one that works.
		collectAt(list.Pos())
	}

	// `FunctionLikeData` and `ClassLikeData` answer across every kind that has the shape, so a
	// method, an arrow, a constructor and a class expression are all covered without naming each.
	//
	// Gating these behind `IsFunctionLikeKind` and `IsClassLike` was tried and reverted, because it
	// was slower rather than faster: 6.57ms against 6.44ms per pass over 300 real files, three runs
	// each. The dispatch these return is cheap enough that the extra kind test costs more than it
	// saves, so the straightforward spelling is also the fast one. Recorded because the optimization
	// looks obviously right and is not.
	if functionLike := node.FunctionLikeData(); functionLike != nil {
		offer(functionLike.Parameters)
	}
	if classLike := node.ClassLikeData(); classLike != nil {
		offer(classLike.Members)
	}

	switch node.Kind {
	case ast.KindBlock:
		offer(node.AsBlock().Statements)
	case ast.KindCaseBlock:
		offer(node.AsCaseBlock().Clauses)
	case ast.KindCaseClause, ast.KindDefaultClause:
		offer(node.AsCaseOrDefaultClause().Statements)
	case ast.KindModuleBlock:
		offer(node.AsModuleBlock().Statements)
	case ast.KindInterfaceDeclaration:
		offer(node.AsInterfaceDeclaration().Members)
	case ast.KindEnumDeclaration:
		offer(node.AsEnumDeclaration().Members)
	case ast.KindTypeLiteral:
		offer(node.AsTypeLiteralNode().Members)
	case ast.KindObjectLiteralExpression:
		offer(node.AsObjectLiteralExpression().Properties)
	case ast.KindArrayLiteralExpression:
		offer(node.AsArrayLiteralExpression().Elements)
	case ast.KindCallExpression:
		offer(node.AsCallExpression().Arguments)
	case ast.KindNewExpression:
		offer(node.AsNewExpression().Arguments)
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		offer(node.AsBindingPattern().Elements)
	case ast.KindNamedImports:
		offer(node.AsNamedImports().Elements)
	case ast.KindNamedExports:
		offer(node.AsNamedExports().Elements)

	case ast.KindJsxExpression:
		// `{/* comment */}` in JSX, whose interior no other position reaches.
		//
		// The shape is the same as every entry above -- an empty delimiter pair holding nothing but
		// a comment -- but it is not a NodeList, so `offer` cannot express it. A JsxExpression with
		// a nil Expression spans exactly `{` and `}`, so its interior begins one byte past `Pos()`.
		//
		// Measured before adding: the scanner returns ZERO comments for `<div>{/*c*/}</div>` in a
		// .tsx file, while the same comment written anywhere else in the same file is found. The
		// container's `Pos()` sits ON the brace, and `GetTrailingCommentRanges` at `Pos()+1` returns
		// the comment, so this is a position nothing probed rather than trivia the scanner cannot
		// classify. Controlled against an ordinary trailing comment through the same call, which
		// returns 1.
		//
		// This is load-bearing for `no-inline-comments`, whose JSX exception is entirely about
		// comments in this position: 31 of its 49 corpus cases are JSX, and without this the rule
		// is silent on all of them -- the exception passing vacuously, which reads exactly like the
		// exception working.
		if expression := node.AsJsxExpression(); expression != nil && expression.Expression == nil {
			collectAt(node.Pos() + 1)
		}
	}
}

// sortByPosition puts comments in source order.
//
// The walk reaches nodes in tree order, which is close to source order but not equal to it, and a
// rule reporting findings out of order is a rule whose fixture assertions cannot be written down.
func sortByPosition(comments []Comment) {
	for outerIndex := 1; outerIndex < len(comments); outerIndex++ {
		current := comments[outerIndex]
		innerIndex := outerIndex - 1
		for innerIndex >= 0 && comments[innerIndex].Range.Pos() > current.Range.Pos() {
			comments[innerIndex+1] = comments[innerIndex]
			innerIndex--
		}
		comments[innerIndex+1] = current
	}
}

// IsJsDoc reports whether a block comment is JSDoc: slash-star-star rather than slash-star.
func (c Comment) IsJsDoc() bool {
	return c.IsBlock && strings.HasPrefix(c.Text, "/**")
}

// ContentLines returns a block comment's inner lines with the delimiters and the leading asterisk
// of each line removed, dropping lines that hold nothing else.
//
// This is what "the comment says one thing" means for the JSDoc rules: the asterisks are furniture,
// so a comment that is three lines of furniture around one line of prose is a one-line comment.
func (c Comment) ContentLines() []string {
	inner := c.Text
	inner = strings.TrimPrefix(inner, "/**")
	inner = strings.TrimPrefix(inner, "/*")
	inner = strings.TrimSuffix(inner, "*/")

	var lines []string
	for _, line := range strings.Split(inner, "\n") {
		line = text.TrimWhitespace(line)
		line = strings.TrimPrefix(line, "*")
		line = text.TrimWhitespace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// canBeginAt reports whether a comment could begin at or just after a position.
//
// A comment starts with `/`, so the question is whether the run of whitespace beginning here reaches
// one. Everything else declines without touching the scanner.
//
// The direction of error is deliberate and matches every other gate in this package: this
// **over**-approximates. A `/` that opens a regex literal or a division is admitted here and then
// declined by the scanner, which costs one scan. A comment this hid would be invisible to every
// comment rule, and the fixtures would still pass, so under-approximating is the failure that cannot
// be allowed. `comments_guard_test.go` asserts the comment set is identical with and without this,
// by range rather than by count.
func canBeginAt(text string, position int) bool {
	// A shebang is trivia too, and it is the one non-whitespace thing a comment can sit behind.
	// `GetLeadingCommentRanges` scans past it, so a guard that stopped at the `#` would hide every
	// comment in an executable script. Found by the corpus, not by the hand-written cases: a
	// `#!/usr/bin/env -S pnpm tsx` line cost this file its module comment and its first import
	// comment before the guard learned about it.
	if position == 0 && strings.HasPrefix(text, "#!") {
		return true
	}

	for index := position; index < len(text); index++ {
		switch text[index] {
		case '/':
			return true
		case ' ', '\t', '\n', '\r', '\v', '\f':
			continue
		default:
			return false
		}
	}
	return false
}

// AllWithoutGuard is the scan with canBeginAt removed, kept only so the differential
// test has something to compare against.
//
// It is a duplicate of All by design: a differential test that shared the implementation it
// is checking would prove nothing. The two must be edited together, and the test fails loudly if
// they drift.
func AllWithoutGuard(sourceFile *ast.SourceFile) []Comment {
	if sourceFile == nil {
		return nil
	}
	text := sourceFile.Text()

	var factory ast.NodeFactory
	seenRanges := make(map[core.TextRange]bool)
	var comments []Comment

	record := func(commentRange ast.CommentRange) {
		if seenRanges[commentRange.TextRange] {
			return
		}
		seenRanges[commentRange.TextRange] = true

		start, end := commentRange.Pos(), commentRange.End()
		if start < 0 || end > len(text) || start >= end {
			return
		}
		startLine, startColumn := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, start)
		endLine, _ := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, end)

		comments = append(comments, Comment{
			Range:       commentRange.TextRange,
			IsBlock:     commentRange.Kind == ast.KindMultiLineCommentTrivia,
			Text:        text[start:end],
			StartLine:   startLine,
			StartColumn: startColumn,
			EndLine:     endLine,
		})
	}

	collectAt := func(position int) {
		if position < 0 || position > len(text) {
			return
		}
		for commentRange := range scanner.GetLeadingCommentRanges(&factory, text, position) {
			record(commentRange)
		}
		for commentRange := range scanner.GetTrailingCommentRanges(&factory, text, position) {
			record(commentRange)
		}
	}

	collectAt(0)

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		collectAt(node.Pos())
		collectAt(node.End())
		collectListInteriors(node, collectAt)
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())

	sortByPosition(comments)
	return comments
}

// LeadingRunFor returns the contiguous run of comments immediately above a node, in source order.
//
// A rule asking "does the comment above this say X" wants the run rather than the nearest comment,
// because a multi-line explanation is several comment nodes and only its first line makes the claim.
// Requiring every line to carry the prefix would forbid explaining anything in a second sentence.
//
// Contiguity is decided by reading the gap between two comments rather than by comparing line
// numbers, which needs no line map: two comments are adjacent exactly when the text between them is
// whitespace containing one newline. Two newlines means a blank line, which ends the run, and
// anything else means code between them.
//
// A block comment holds its whole text in one node and so ends any run.
//
// The scan is the shared per-file one, so a rule reaching for this pays for it once across the whole
// run rather than once per node it asks about. That is not a micro-optimization: three rules calling
// a well-factored scan directly each rescanned the file, measured at 807ms, 619ms and 601ms while
// visiting exactly one node per file.
func LeadingRunFor(ctx rule.Context, node *ast.Node) []Comment {
	if node == nil || ctx.SourceFile == nil {
		return nil
	}

	all := ForFile(ctx)
	if len(all) == 0 {
		return nil
	}

	// The comments above the node are those ending at or before where its own text begins. The
	// token position is used rather than Pos(), since Pos() sits before the leading trivia and would
	// place the node before its own comments.
	nodeStart := rule.TokenRange(ctx.SourceFile, node).Pos()

	last := -1
	for index, comment := range all {
		if comment.Range.End() <= nodeStart {
			last = index
			continue
		}
		break
	}
	if last < 0 {
		return nil
	}

	// Only a comment the node actually follows counts. A comment separated from it by code belongs
	// to that code, and one separated by a blank line is a passage of its own.
	text := ctx.SourceFile.Text()
	if !isAdjacentGap(text, all[last].Range.End(), nodeStart) {
		return nil
	}

	first := last
	for first > 0 {
		previous, current := all[first-1], all[first]
		if previous.IsBlock || current.IsBlock {
			break
		}
		if !isAdjacentGap(text, previous.Range.End(), current.Range.Pos()) {
			break
		}
		first--
	}
	return all[first : last+1]
}

// isAdjacentGap reports whether two source positions are separated by nothing but one line break.
//
// The gap between a comment and whatever follows it is whitespace containing exactly one newline
// when the two are on consecutive lines with nothing between them. Zero newlines means they share a
// line, which is a trailing comment rather than a leading one; two means a blank line, which ends a
// passage.
func isAdjacentGap(text string, from int, to int) bool {
	if from > to || to > len(text) {
		return false
	}
	gap := text[from:to]
	if strings.TrimSpace(gap) != "" {
		return false
	}
	return strings.Count(gap, "\n") == 1
}
