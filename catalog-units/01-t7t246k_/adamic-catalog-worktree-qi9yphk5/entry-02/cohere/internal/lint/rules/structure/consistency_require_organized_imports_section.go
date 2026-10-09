package structure

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// directiveComments are the trailing comments the two directives are re-emitted with.
var directiveComments = map[string]string{
	"use client": "// Uses client-only features",
	"use server": "// Uses server-only features",
}

// classifiedImport is one import declaration with the group it belongs in and the text it renders
// as, which is its own source text with any comments it owns stacked above it.
type classifiedImport struct {
	source string
	group  string
	text   string
}

// leadingDirective returns the `'use client'` or `'use server'` directive when it is the file's
// first statement, along with the statement node.
//
// Only the first statement counts. A directive written after an import is not a directive to the
// runtime either, so declining it is fidelity to the language rather than to the original.
func leadingDirective(statements *ast.NodeList) (string, *ast.Node) {
	if statements == nil || len(statements.Nodes) == 0 {
		return "", nil
	}
	first := statements.Nodes[0]
	if first.Kind != ast.KindExpressionStatement {
		return "", nil
	}
	expression := first.AsExpressionStatement().Expression
	if expression == nil || expression.Kind != ast.KindStringLiteral {
		return "", nil
	}
	text := expression.Text()
	if text != "use client" && text != "use server" {
		return "", nil
	}
	return text, first
}

// importSectionStart is the offset the canonical rendering is compared from.
//
// Three answers, in the original's order. With a directive the section starts at the directive. With
// no directive it starts at the earliest `// Dependencies` line comment above the first import, if
// there is one. Otherwise it starts at the first import's own text.
//
// **That last case is where the original's fixer loses a disable comment**, because the replaced
// range in ESTree begins at trivia while this offset does not, so a comment above the first import
// is inside the range and outside the preservation filter. We only compare here, so the same offset
// is harmless: a file whose first import carries a comment simply has that comment outside the
// compared section, which is what the original's own comparison does too.
func importSectionStart(
	sourceFile *ast.SourceFile,
	sourceText string,
	fileComments []comments.Comment,
	firstImport *ast.Node,
	directiveNode *ast.Node,
) int {
	if directiveNode != nil {
		return tokenStart(sourceFile, directiveNode)
	}

	firstImportStart := tokenStart(sourceFile, firstImport)
	earliest := -1
	for _, comment := range fileComments {
		if comment.IsBlock {
			continue
		}
		if comment.Range.End() > firstImportStart {
			continue
		}
		if comment.Range.Pos() < leadingTriviaStart(firstImport) {
			continue
		}
		if !strings.HasPrefix(strings.TrimLeft(commentInnerText(comment), " \t"), "Dependencies") {
			continue
		}
		if earliest < 0 || comment.Range.Pos() < earliest {
			earliest = comment.Range.Pos()
		}
	}
	if earliest >= 0 {
		return earliest
	}
	return firstImportStart
}

// fileDescriptionBlock is the block comment that describes the whole file rather than one import.
//
// Only present when the file opens with a directive, which is the original's condition and not an
// accident: without a directive there is nothing above the block to separate it from a preamble, so
// the original leaves it alone. The last such block wins, matching the original's `.pop()`.
func fileDescriptionBlock(
	sourceFile *ast.SourceFile,
	fileComments []comments.Comment,
	firstImport *ast.Node,
	directiveNode *ast.Node,
) *comments.Comment {
	if directiveNode == nil {
		return nil
	}
	directiveEnd := directiveNode.End()
	firstImportStart := tokenStart(sourceFile, firstImport)

	var found *comments.Comment
	for index := range fileComments {
		comment := fileComments[index]
		if !comment.IsBlock {
			continue
		}
		if comment.Range.Pos() < directiveEnd {
			continue
		}
		if comment.Range.End() >= firstImportStart {
			continue
		}
		found = &fileComments[index]
	}
	return found
}

// classifyDeclarations turns each import into its group and its rendered text.
func classifyDeclarations(
	sourceFile *ast.SourceFile,
	sourceText string,
	fileComments []comments.Comment,
	importDeclarations []*ast.Node,
	sectionStart int,
	descriptionBlock *comments.Comment,
) []classifiedImport {
	classified := make([]classifiedImport, 0, len(importDeclarations))

	for _, declaration := range importDeclarations {
		source := moduleSpecifierText(declaration)
		isTypeOnly := isTypeOnlyImport(declaration)
		localNames := localBindingNames(declaration)

		preserved := preservedCommentsFor(sourceFile, sourceText, fileComments, declaration, sectionStart, descriptionBlock)

		importText := sourceText[tokenStart(sourceFile, declaration):declaration.End()]
		if trailing, hasTrailing := trailingCommentFor(sourceText, fileComments, declaration); hasTrailing {
			importText += " " + trailing
		}
		fullText := importText
		if len(preserved) > 0 {
			fullText = strings.Join(preserved, "\n") + "\n" + importText
		}

		classified = append(classified, classifiedImport{
			source: source,
			group:  classifyImport(source, isTypeOnly, localNames),
			text:   fullText,
		})
	}
	return classified
}

// trailingCommentFor returns a comment written on the same line as an import, after it.
//
// # Why this is not left to preservedCommentsFor
//
// A trailing comment is, syntactically, the leading trivia of whatever comes next. So a section
// that only ever collects leading comments moves `import zebra from 'zebra'; // note about zebra`
// to sit above whichever import sorts next, and the note ends up describing a module it was never
// about. Measured against the original on exactly that input: the comment came back above the
// React import.
//
// That is one of three defects the original's fixer carries, and it is the reason this rule
// declined to ship a fixer at all. Attaching the comment to the import it trails removes it, which
// is what lets the fixer exist.
//
// A comment counts as trailing when it starts after the import ends and no newline separates them.
// The newline test is the whole predicate: a comment on the next line is a leading comment of the
// next statement and belongs to it, which is the case this must not steal.
func trailingCommentFor(sourceText string, fileComments []comments.Comment,
	declaration *ast.Node) (string, bool) {
	for _, comment := range fileComments {
		if comment.Range.Pos() < declaration.End() {
			continue
		}
		between := sourceText[declaration.End():comment.Range.Pos()]
		if strings.ContainsAny(between, "\n\r") {
			continue
		}
		if text.TrimWhitespace(between) != "" {
			continue
		}
		return comment.Text, true
	}
	return "", false
}

// preservedCommentsFor collects the comments that render above one import.
//
// Three kinds are dropped rather than preserved, because the rendering regenerates them: a group
// header, a directive comment, and the file description. Everything else a reader wrote above an
// import travels with it.
func preservedCommentsFor(
	sourceFile *ast.SourceFile,
	sourceText string,
	fileComments []comments.Comment,
	declaration *ast.Node,
	sectionStart int,
	descriptionBlock *comments.Comment,
) []string {
	preserved := []string{}
	triviaStart := leadingTriviaStart(declaration)
	declarationStart := tokenStart(sourceFile, declaration)

	for _, comment := range fileComments {
		if comment.Range.Pos() < triviaStart || comment.Range.End() > declarationStart {
			continue
		}
		// A comment on the same line as whatever precedes it trails that statement rather than
		// leading this one, and `trailingCommentFor` has already attached it there. Without this
		// the comment renders twice: once beside the import it describes and once above the import
		// that happens to sort next.
		if trailsPrecedingLine(sourceText, comment.Range.Pos()) {
			continue
		}
		// A comment above the section start is preamble: a license header, a file banner. It stays
		// where it is rather than being pulled into a group.
		if comment.Range.Pos() < sectionStart {
			continue
		}
		inner := strings.TrimLeft(commentInnerText(comment), " \t")
		if strings.HasPrefix(inner, "Dependencies -") {
			continue
		}
		if strings.Contains(inner, "client-only features") || strings.Contains(inner, "server-only features") {
			continue
		}
		if descriptionBlock != nil && comment.Range.Pos() == descriptionBlock.Range.Pos() {
			continue
		}
		preserved = append(preserved, comment.Text)
	}
	return preserved
}

// trailsPrecedingLine reports whether a comment begins on a line that already holds code.
//
// Walking backwards to the line start and asking whether anything non-blank precedes it is the
// whole test, and it is the difference between a comment that describes the statement it sits
// beside and one that introduces the statement below.
func trailsPrecedingLine(sourceText string, commentStart int) bool {
	for cursor := commentStart - 1; cursor >= 0; cursor-- {
		character := sourceText[cursor]
		if character == '\n' || character == '\r' {
			return false
		}
		if character != ' ' && character != '\t' {
			return true
		}
	}
	return false
}

// interleavedStatements collects the non-import statements written between the first and last
// import, and returns the offset the section ends at.
//
// The original appends these below the organized imports, which is one of the three reasons this
// port declines to ship the fixer: a statement that ran before an import would run after it. They
// are still collected, because the comparison has to reproduce the original's verdict exactly, and
// a file with interleaved code is one the original reports on.
func interleavedStatements(
	sourceFile *ast.SourceFile,
	sourceText string,
	fileComments []comments.Comment,
	statements *ast.NodeList,
	importDeclarations []*ast.Node,
	sectionStart int,
) ([]string, int) {
	firstImport := importDeclarations[0]
	lastImport := importDeclarations[len(importDeclarations)-1]

	firstIndex, lastIndex := -1, -1
	for index, statement := range statements.Nodes {
		if statement == firstImport {
			firstIndex = index
		}
		if statement == lastImport {
			lastIndex = index
		}
	}
	sectionEnd := lastImport.End()
	if firstIndex < 0 || lastIndex < 0 {
		return nil, extendPastNewline(sourceText, sectionEnd)
	}

	interleaved := []string{}
	for index := firstIndex; index <= lastIndex; index++ {
		statement := statements.Nodes[index]
		if statement.Kind == ast.KindImportDeclaration {
			continue
		}
		commentLines := []string{}
		triviaStart := leadingTriviaStart(statement)
		statementStart := tokenStart(sourceFile, statement)
		for _, comment := range fileComments {
			if comment.Range.Pos() < triviaStart || comment.Range.End() > statementStart {
				continue
			}
			if comment.Range.Pos() < sectionStart {
				continue
			}
			commentLines = append(commentLines, comment.Text)
		}
		statementText := sourceText[statementStart:statement.End()]
		if len(commentLines) > 0 {
			statementText = strings.Join(commentLines, "\n") + "\n" + statementText
		}
		interleaved = append(interleaved, statementText)

		if statement.End() > sectionEnd {
			sectionEnd = statement.End()
		}
	}

	return interleaved, extendPastNewline(sourceText, sectionEnd)
}

// extendPastNewline swallows one trailing newline after the section, matching the original.
//
// A comment trailing the last statement is swallowed first, because the rendering now carries it
// with the import it trails. Leaving it outside the span would put it in both places at once: the
// rendered section would hold a copy and the source would keep the original, so the fix would
// duplicate the comment. The verdict is unaffected, since the comparison uses the same span on both
// sides.
func extendPastNewline(sourceText string, offset int) int {
	offset = skipTrailingComment(sourceText, offset)
	if offset < len(sourceText) && sourceText[offset] == '\n' {
		return offset + 1
	}
	return offset
}

// skipTrailingComment advances past a comment written on the same line as what precedes it.
//
// Only a line comment: a block comment can carry a newline of its own, and swallowing one would
// pull an unrelated following line into the section. The same-line test is what makes this a
// trailing comment rather than the leading comment of whatever comes next.
func skipTrailingComment(sourceText string, offset int) int {
	cursor := offset
	for cursor < len(sourceText) && (sourceText[cursor] == ' ' || sourceText[cursor] == '\t') {
		cursor++
	}
	if cursor+1 >= len(sourceText) || sourceText[cursor] != '/' || sourceText[cursor+1] != '/' {
		return offset
	}
	for cursor < len(sourceText) && sourceText[cursor] != '\n' && sourceText[cursor] != '\r' {
		cursor++
	}
	return cursor
}

// renderSection builds the canonical text the real section is compared against.
func renderSection(
	directive string,
	descriptionBlock *comments.Comment,
	classified []classifiedImport,
	interleaved []string,
) string {
	lines := []string{}

	if directive != "" {
		lines = append(lines, "'"+directive+"'; "+directiveComments[directive], "")
	}
	if descriptionBlock != nil {
		lines = append(lines, descriptionBlock.Text, "")
	}

	grouped := map[string][]classifiedImport{}
	for _, item := range classified {
		grouped[item.group] = append(grouped[item.group], item)
	}

	firstGroup := true
	for _, groupName := range groupOrder {
		groupImports := grouped[groupName]
		if len(groupImports) == 0 {
			continue
		}
		sortGroup(groupName, groupImports)

		if !firstGroup {
			lines = append(lines, "")
		}
		firstGroup = false

		lines = append(lines, "// Dependencies - "+groupName)
		for _, item := range groupImports {
			lines = append(lines, item.text)
		}
	}

	rendered := strings.Join(lines, "\n") + "\n"

	// **Appending the interleaved text cannot change this rule's verdict, and that was scored as a
	// survivor before it was understood rather than after.**
	//
	// Whenever `interleaved` is non-empty the comparison fails either way. `sectionEnd` was extended
	// to cover the last interleaved node, so the current text ends with that node's own source,
	// while the imports block above ends with an import statement. An interleaved node is by
	// construction not an import declaration, so those two tails cannot match, and the verdict is
	// report with the append or without it. Confirmed by applying the mutation and running the whole
	// suite, including all 137 classification rows: nothing moved.
	//
	// It is kept because the text is what a fixer would emit, and because deleting it would make
	// this function stop describing the canonical form it exists to describe. A reader comparing
	// this to the original would otherwise find a rendering that silently omits a section the
	// original renders.
	if len(interleaved) > 0 {
		rendered += "\n" + strings.Join(interleaved, "\n\n") + "\n"
	}
	return rendered
}

// suppressionPrecedesSection reports whether a lint suppression sits above the section boundary.
//
// A suppression comment silences the line directly beneath it, so its meaning is positional. A
// comment above the first import is outside the compared section by design (only a `Dependencies`
// header moves the boundary earlier), which means the fix rewrites everything below it and leaves
// the comment where it was. For an ordinary comment that is correct and is what keeps a license
// header in place. For a suppression it is not: the line it was silencing has moved, and the
// comment now silences a group header instead.
//
// Upstream deletes that comment outright, which is worse, and is the first of the three defects
// that kept this rule from shipping a fixer. Preserving it displaced is safer than deleting it and
// still not right, so these files are reported without a fix and a reader moves the comment onto
// the import it belongs to.
//
// Measured on Kirk's tree: no file carries the shape, so this costs nothing today. It exists
// because a suppression that stops suppressing is the one failure here that hides.
func suppressionPrecedesSection(sourceText string, fileComments []comments.Comment, sectionStart int) bool {
	for _, comment := range fileComments {
		if comment.Range.End() > sectionStart {
			continue
		}
		inner := strings.TrimLeft(commentInnerText(comment), " \t")
		if strings.HasPrefix(inner, "eslint-disable") || strings.HasPrefix(inner, "cohere-disable") ||
			strings.HasPrefix(inner, "verify-disable") {
			return true
		}
	}
	return false
}

// canRenderSafely reports whether the rendered section can be written back over the source.
//
// # Why a rendering that is correct to compare is not always correct to apply
//
// `renderSection` describes the canonical form, and for the verdict that is all it has to do: a
// file whose section differs from the canonical form is reported, and how the difference would be
// repaired never enters the question. A fixer asks a second thing of the same string, which is that
// writing it back preserves what the file meant.
//
// One shape fails that. A statement written between two imports runs before the imports below it,
// and the rendering appends it after every one of them, so applying the text would move code across
// an evaluation boundary. That is upstream's third fixer defect, reproduced here deliberately
// because the verdict has to match, and it is the one thing that cannot simply be corrected: the
// canonical form has no place to put a statement mid-section, since the imports around it are being
// reordered relative to each other.
//
// So the fixer declines these files rather than guessing. The finding still reports and a reader
// still moves the statement, which is the same outcome the rule had before it could fix anything at
// all. Measured on Kirk's tree: zero files carry the shape, so the decline costs nothing today and
// exists for the file somebody writes next.
func canRenderSafely(interleaved []string) bool {
	return len(interleaved) == 0
}

// sortGroup orders one group's imports by module specifier, with react first inside Frameworks.
//
// A stable sort, matching the original. Two imports of the same module keep their written order,
// which is observable: `import a from 'alpha'` above `import b from 'alpha'` stays that way.
func sortGroup(groupName string, groupImports []classifiedImport) {
	sort.SliceStable(groupImports, func(first, second int) bool {
		if groupName == "Frameworks" {
			firstIsReact := groupImports[first].source == "react"
			secondIsReact := groupImports[second].source == "react"
			if firstIsReact != secondIsReact {
				return firstIsReact
			}
		}
		return groupImports[first].source < groupImports[second].source
	})
}

// moduleSpecifierText reads an import's module specifier.
func moduleSpecifierText(declaration *ast.Node) string {
	importDeclaration := declaration.AsImportDeclaration()
	if importDeclaration == nil || importDeclaration.ModuleSpecifier == nil {
		return ""
	}
	if !ast.IsStringLiteralLike(importDeclaration.ModuleSpecifier) {
		return ""
	}
	return importDeclaration.ModuleSpecifier.Text()
}

// isTypeOnlyImport reports `import type { T } from 'm'`, the statement-level form.
//
// Only the statement form counts, matching the original's `node.importKind === 'type'`. An inline
// `import { type T } from 'm'` has a value-level import kind and classifies by its path, which was
// pinned rather than assumed.
func isTypeOnlyImport(declaration *ast.Node) bool {
	importDeclaration := declaration.AsImportDeclaration()
	if importDeclaration == nil || importDeclaration.ImportClause == nil {
		return false
	}
	clause := importDeclaration.ImportClause.AsImportClause()
	return clause != nil && clause.IsTypeOnly()
}

// localBindingNames reads the local names one import introduces, in every shape.
//
// Through `imports.BindingsOf` rather than by destructuring the clause, which is what the retrofit
// guard in internal/dispatch requires: the three shapes live in three places in this AST and a rule
// reading one of them pins half the forms while looking complete.
//
// The original reads `specifier.local?.name` across a flat ESTree specifier array, so a default, a
// namespace and each named binding all contribute. `Bindings` separates them, and all three are
// read here to reproduce that.
func localBindingNames(declaration *ast.Node) []string {
	bindings := imports.BindingsOf(declaration)
	names := []string{}

	if bindings.Default != nil && bindings.Default.Kind == ast.KindIdentifier {
		names = append(names, bindings.Default.Text())
	}
	if bindings.Namespace != nil {
		if name := bindings.Namespace.Name(); name != nil && name.Kind == ast.KindIdentifier {
			names = append(names, name.Text())
		}
	}
	for _, element := range bindings.Named {
		// The LOCAL name, which is what the original reads. `{ Thing as useThing }` contributes
		// `useThing` and classifies as Hooks; reading the imported name would answer `Thing`.
		if name := element.Name(); name != nil && name.Kind == ast.KindIdentifier {
			names = append(names, name.Text())
		}
	}
	return names
}

// commentInnerText strips a comment's delimiters, giving the text the original's `comment.value`
// holds.
func commentInnerText(comment comments.Comment) string {
	text := comment.Text
	if comment.IsBlock {
		text = strings.TrimPrefix(text, "/*")
		return strings.TrimSuffix(text, "*/")
	}
	return strings.TrimPrefix(text, "//")
}

// tokenStart is a node's own text start, past its leading trivia.
//
// The SourceFile is required rather than optional: `rule.TokenRange` falls back to `node.Loc` when
// handed nil, and `Loc.Pos()` is where the trivia begins, which is the opposite of what this asks
// for. A nil here would silently return the comment's start and every offset downstream would be
// wrong by however long the comment is.
func tokenStart(sourceFile *ast.SourceFile, node *ast.Node) int {
	return int(rule.TokenRange(sourceFile, node).Pos())
}

// leadingTriviaStart is where a node's leading trivia begins, which is where its comments live.
func leadingTriviaStart(node *ast.Node) int {
	return node.Pos()
}
