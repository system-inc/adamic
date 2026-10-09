package typescript

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/ecmascript/tokens"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The repair for `consistent-type-imports`, ported from upstream's fixer and delivered as ONE edit.
//
// # Why one edit
//
// Upstream's fixer yields several primitive edits per report: an insertion before the declaration,
// ` type` after the `import` keyword, a removal per moved specifier. Our engine flattens a report's
// fixes into independent proposals and refuses two insertions at one point as a mutual overlap,
// measured at zero applied and 1024 rejections for upstream's ordinary split. So this file computes
// the same primitive edits upstream does, applies them to a copy of the declaration's own text, and
// proposes the result as a single replacement of the declaration. Every primitive edit upstream makes
// lands inside its own declaration but one, which is the next section.
//
// # The one divergence: no merging into another declaration
//
// When the file already has `import type { X } from 'm'`, upstream inserts the moved names into that
// declaration instead of writing a new one, which is a second statement and so a second edit. Here
// the moved names always go into a new declaration written beside the reported one. The result is two
// declarations from the same module where upstream writes one; it type-checks and means the same, and
// keeping the edit inside the reported declaration is what keeps it one edit. Inline style reads the
// reported declaration as the value import it adds `type` to, where upstream reads the first value
// import from the module, which differs only when that first import binds no named specifiers.
//
// # What the edit can never do
//
// The edit only moves the names the report already found type-only, and that decision is the rule's,
// including the decorator-metadata exemption: a class that `emitDecoratorMetadata` serializes is a
// value use and is never in the set this fixer receives. A declaration whose primitive edits overlap,
// or that carries import attributes (a type-only import cannot), gets no fix and still reports.

// consistentTypeImportsSpecifiers is upstream's `classifySpecifier`, plus the count of every specifier
// in the declaration.
type consistentTypeImportsSpecifiers struct {
	defaultSpecifier   *ast.Node
	namespaceSpecifier *ast.Node
	namedSpecifiers    []*ast.Node
	total              int
}

func classifyConsistentTypeImportsSpecifiers(statement *ast.Node) consistentTypeImportsSpecifiers {
	bindings := imports.BindingsOf(statement)
	classified := consistentTypeImportsSpecifiers{
		defaultSpecifier:   bindings.Default,
		namespaceSpecifier: bindings.Namespace,
	}
	for _, named := range bindings.Named {
		if named.Kind == ast.KindImportSpecifier {
			classified.namedSpecifiers = append(classified.namedSpecifiers, named)
		}
	}
	if classified.defaultSpecifier != nil {
		classified.total++
	}
	if classified.namespaceSpecifier != nil {
		classified.total++
	}
	classified.total += len(classified.namedSpecifiers)
	return classified
}

// consistentTypeImportsFixer holds what every primitive edit is computed against.
type consistentTypeImportsFixer struct {
	sourceFile *ast.SourceFile
	text       string
	statement  *ast.Node
	start      int
	end        int
	tokens     tokens.List
	edits      []consistentTypeImportsEdit
}

// consistentTypeImportsEdit is one of upstream's primitive fixes: replace [start, end) with text.
type consistentTypeImportsEdit struct {
	start int
	end   int
	text  string
}

func newConsistentTypeImportsFixer(sourceFile *ast.SourceFile, statement *ast.Node) *consistentTypeImportsFixer {
	statementRange := rule.TokenRange(sourceFile, statement)
	fixer := &consistentTypeImportsFixer{
		sourceFile: sourceFile,
		text:       sourceFile.Text(),
		statement:  statement,
		start:      statementRange.Pos(),
		end:        statementRange.End(),
		tokens:     tokens.Of(sourceFile, statement),
	}
	return fixer
}

// nodeRange is a node's span without its leading trivia, which is what an ESTree range is.
func (fixer *consistentTypeImportsFixer) nodeRange(node *ast.Node) (int, int) {
	nodeRange := rule.TokenRange(fixer.sourceFile, node)
	return nodeRange.Pos(), nodeRange.End()
}

func (fixer *consistentTypeImportsFixer) nodeText(node *ast.Node) string {
	start, end := fixer.nodeRange(node)
	return fixer.text[start:end]
}

// nextTokenOrComment is upstream's `getTokenAfter(..., { includeComments: true })`, as a position.
func (fixer *consistentTypeImportsFixer) nextTokenOrComment(position int) int {
	return scanner.SkipTriviaEx(fixer.text, position, &scanner.SkipTriviaOptions{StopAtComments: true})
}

func (fixer *consistentTypeImportsFixer) insertBefore(position int, text string) {
	fixer.edits = append(fixer.edits, consistentTypeImportsEdit{start: position, end: position, text: text})
}

func (fixer *consistentTypeImportsFixer) removeRange(start int, end int) {
	fixer.edits = append(fixer.edits, consistentTypeImportsEdit{start: start, end: end})
}

func (fixer *consistentTypeImportsFixer) replaceRange(start int, end int, text string) {
	fixer.edits = append(fixer.edits, consistentTypeImportsEdit{start: start, end: end, text: text})
}

// importToken is the declaration's `import` keyword, which is its first token.
func (fixer *consistentTypeImportsFixer) importToken() (tokens.Token, bool) {
	if len(fixer.tokens) == 0 || fixer.tokens[0].Kind != ast.KindImportKeyword {
		return tokens.Token{}, false
	}
	return fixer.tokens[0], true
}

func (fixer *consistentTypeImportsFixer) sourceText() string {
	return fixer.nodeText(fixer.statement.AsImportDeclaration().ModuleSpecifier)
}

// compose applies the primitive edits to the declaration's text and returns the one replacement.
//
// Ordered by position with insertion order kept on a tie, which is how ESLint merges one report's
// fixes. Two edits that overlap rather than touch refuse the whole repair, because no order of them
// is the one upstream meant.
func (fixer *consistentTypeImportsFixer) compose() (rule.Fix, bool) {
	if len(fixer.edits) == 0 {
		return rule.Fix{}, false
	}
	edits := append([]consistentTypeImportsEdit{}, fixer.edits...)
	sort.SliceStable(edits, func(first int, second int) bool {
		if edits[first].start != edits[second].start {
			return edits[first].start < edits[second].start
		}
		return edits[first].end < edits[second].end
	})
	var replacement strings.Builder
	cursor := fixer.start
	for _, edit := range edits {
		if edit.start < cursor || edit.end > fixer.end || edit.start < fixer.start {
			return rule.Fix{}, false
		}
		replacement.WriteString(fixer.text[cursor:edit.start])
		replacement.WriteString(edit.text)
		cursor = edit.end
	}
	replacement.WriteString(fixer.text[cursor:fixer.end])
	return rule.ReplaceRange(core.NewTextRange(fixer.start, fixer.end), replacement.String()), true
}

// consistentTypeImportsTypeImportFix is upstream's `fixToTypeImportDeclaration`, for a declaration
// some or all of whose names are used only as types.
func consistentTypeImportsTypeImportFix(
	sourceFile *ast.SourceFile,
	statement *ast.Node,
	typeSpecifiers map[*ast.Node]bool,
	fixStyle ConsistentTypeImportsFixStyle,
) (rule.Fix, bool) {
	if statement.AsImportDeclaration().Attributes != nil {
		return rule.Fix{}, false
	}
	fixer := newConsistentTypeImportsFixer(sourceFile, statement)
	if _, found := fixer.importToken(); !found {
		return rule.Fix{}, false
	}
	specifiers := classifyConsistentTypeImportsSpecifiers(statement)
	if !fixer.toTypeImportDeclaration(specifiers, typeSpecifiers, fixStyle) {
		return rule.Fix{}, false
	}
	return fixer.compose()
}

func (fixer *consistentTypeImportsFixer) toTypeImportDeclaration(
	specifiers consistentTypeImportsSpecifiers,
	typeSpecifiers map[*ast.Node]bool,
	fixStyle ConsistentTypeImportsFixStyle,
) bool {
	inline := fixStyle == ConsistentTypeImportsFixStyleInline

	// A named-only declaration whose remaining names are already inline `type` would be left as
	// `import { type X } from 'm'` once the others move out, and under `verbatimModuleSyntax` that
	// compiles to `import {} from 'm'`, a side-effect import the original never had. So when every name
	// is a type one way or the other, the whole declaration becomes `import type`, inline keywords
	// stripped. Upstream counts inline names as neither type nor value and writes the split; phi api's
	// codemod wrote the whole-declaration form, and the corpus comparison found the four sites.
	if !inline && specifiers.defaultSpecifier == nil && specifiers.namespaceSpecifier == nil {
		everyNameIsAType := len(specifiers.namedSpecifiers) > 0
		for _, named := range specifiers.namedSpecifiers {
			if !typeSpecifiers[named] && !named.AsImportSpecifier().IsTypeOnly {
				everyNameIsAType = false
			}
		}
		if everyNameIsAType {
			return fixer.insertTypeSpecifierForImportDeclaration(false)
		}
	}

	allNamedAreTypes := true
	someNamedAreTypes := false
	for _, named := range specifiers.namedSpecifiers {
		if typeSpecifiers[named] {
			someNamedAreTypes = true
		} else {
			allNamedAreTypes = false
		}
	}

	if specifiers.namespaceSpecifier != nil && specifiers.defaultSpecifier == nil {
		// import * as types from 'foo'
		return fixer.insertTypeSpecifierForImportDeclaration(false)
	}

	if specifiers.defaultSpecifier != nil {
		if typeSpecifiers[specifiers.defaultSpecifier] && len(specifiers.namedSpecifiers) == 0 &&
			specifiers.namespaceSpecifier == nil {
			// import Type from 'foo'
			return fixer.insertTypeSpecifierForImportDeclaration(true)
		}
		if inline && !typeSpecifiers[specifiers.defaultSpecifier] && len(specifiers.namedSpecifiers) > 0 &&
			specifiers.namespaceSpecifier == nil {
			// import AValue, {BValue, Type1, Type2} from 'foo'
			return fixer.insertTypeKeywordInNamedSpecifierList(specifiers, typeSpecifiers)
		}
	} else if specifiers.namespaceSpecifier == nil {
		if inline && someNamedAreTypes {
			// import {AValue, Type1, Type2} from 'foo'
			return fixer.insertTypeKeywordInNamedSpecifierList(specifiers, typeSpecifiers)
		}
		if allNamedAreTypes {
			// import {Type1, Type2} from 'foo'
			return fixer.insertTypeSpecifierForImportDeclaration(false)
		}
	}

	typeNamedSpecifiers := []*ast.Node{}
	for _, named := range specifiers.namedSpecifiers {
		if typeSpecifiers[named] {
			typeNamedSpecifiers = append(typeNamedSpecifiers, named)
		}
	}

	typeNamedSpecifiersText, removals, ok := fixer.namedSpecifierFixes(typeNamedSpecifiers, specifiers.namedSpecifiers)
	if !ok {
		return false
	}
	if len(typeNamedSpecifiers) > 0 {
		// No existing `import type { ... }` is merged into; see the file comment.
		if inline {
			texts := make([]string, 0, len(typeNamedSpecifiers))
			for _, named := range typeNamedSpecifiers {
				texts = append(texts, "type "+fixer.nodeText(named))
			}
			fixer.insertBefore(fixer.start, "import {"+strings.Join(texts, ", ")+"} from "+fixer.sourceText()+";\n")
		} else {
			fixer.insertBefore(fixer.start, "import type {"+typeNamedSpecifiersText+"} from "+fixer.sourceText()+";\n")
		}
	}

	var namespaceRemoval *consistentTypeImportsEdit
	if specifiers.namespaceSpecifier != nil && typeSpecifiers[specifiers.namespaceSpecifier] {
		// import Def, * as Ns from 'foo'
		namespaceStart, namespaceEnd := fixer.nodeRange(specifiers.namespaceSpecifier)
		comma, found := fixer.tokens.Before(namespaceStart)
		if !found || comma.Kind != ast.KindCommaToken {
			return false
		}
		namespaceRemoval = &consistentTypeImportsEdit{start: comma.Start, end: namespaceEnd}
		fixer.insertBefore(fixer.start, "import type "+fixer.nodeText(specifiers.namespaceSpecifier)+
			" from "+fixer.sourceText()+";\n")
	}

	if specifiers.defaultSpecifier != nil && typeSpecifiers[specifiers.defaultSpecifier] {
		if len(typeSpecifiers) == specifiers.total {
			// import type Type from 'foo'
			importToken, _ := fixer.importToken()
			fixer.insertBefore(importToken.End, " type")
		} else {
			defaultStart, defaultEnd := fixer.nodeRange(specifiers.defaultSpecifier)
			comma, found := fixer.tokens.After(defaultEnd)
			if !found || comma.Kind != ast.KindCommaToken {
				return false
			}
			defaultText := text.TrimWhitespace(fixer.text[defaultStart:comma.Start])
			fixer.insertBefore(fixer.start, "import type "+defaultText+" from "+fixer.sourceText()+";\n")
			fixer.removeRange(defaultStart, fixer.nextTokenOrComment(comma.End))
		}
	}

	fixer.edits = append(fixer.edits, removals...)
	if namespaceRemoval != nil {
		fixer.edits = append(fixer.edits, *namespaceRemoval)
	}
	return true
}

// namedSpecifierFixes is upstream's `getFixesNamedSpecifiers`: the text of the named specifiers that
// move, and the removals that take them out of this declaration.
func (fixer *consistentTypeImportsFixer) namedSpecifierFixes(
	subset []*ast.Node,
	all []*ast.Node,
) (string, []consistentTypeImportsEdit, bool) {
	if len(all) == 0 || len(subset) == 0 {
		return "", nil, true
	}
	if len(subset) == len(all) {
		// import DefType, {...} from 'foo'
		firstStart, _ := fixer.nodeRange(subset[0])
		openingBrace, found := fixer.tokens.Before(firstStart)
		if !found || openingBrace.Kind != ast.KindOpenBraceToken {
			return "", nil, false
		}
		comma, found := fixer.tokens.Before(openingBrace.Start)
		if !found || comma.Kind != ast.KindCommaToken {
			return "", nil, false
		}
		closingBrace, found := fixer.tokens.FirstBetween(openingBrace.End, fixer.end, ast.KindCloseBraceToken)
		if !found {
			return "", nil, false
		}
		return fixer.text[openingBrace.End:closingBrace.Start],
			[]consistentTypeImportsEdit{{start: comma.Start, end: closingBrace.End}}, true
	}

	groups := [][]*ast.Node{}
	group := []*ast.Node{}
	inSubset := map[*ast.Node]bool{}
	for _, named := range subset {
		inSubset[named] = true
	}
	for _, named := range all {
		if inSubset[named] {
			group = append(group, named)
		} else if len(group) > 0 {
			groups = append(groups, group)
			group = []*ast.Node{}
		}
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}

	texts := []string{}
	removals := []consistentTypeImportsEdit{}
	for _, group := range groups {
		removeStart, _ := fixer.nodeRange(group[0])
		_, removeEnd := fixer.nodeRange(group[len(group)-1])
		before, found := fixer.tokens.Before(removeStart)
		if !found {
			return "", nil, false
		}
		textStart := before.End
		if before.Kind == ast.KindCommaToken {
			removeStart = before.Start
		} else {
			removeStart = before.End
		}
		after, found := fixer.tokens.After(removeEnd)
		if !found {
			return "", nil, false
		}
		textEnd := after.Start
		isFirst := all[0] == group[0]
		isLast := all[len(all)-1] == group[len(group)-1]
		if (isFirst || isLast) && after.Kind == ast.KindCommaToken {
			removeEnd = after.End
		}
		texts = append(texts, fixer.text[textStart:textEnd])
		removals = append(removals, consistentTypeImportsEdit{start: removeStart, end: removeEnd})
	}
	return strings.Join(texts, ","), removals, true
}

// insertTypeKeywordInNamedSpecifierList is upstream's inline repair: `type ` before each moved name,
// in place, so comments and layout around it survive.
func (fixer *consistentTypeImportsFixer) insertTypeKeywordInNamedSpecifierList(
	specifiers consistentTypeImportsSpecifiers,
	typeSpecifiers map[*ast.Node]bool,
) bool {
	for _, named := range specifiers.namedSpecifiers {
		if !typeSpecifiers[named] {
			continue
		}
		start, end := fixer.nodeRange(named)
		fixer.replaceRange(start, end, "type "+fixer.text[start:end])
	}
	return true
}

// insertTypeSpecifierForImportDeclaration is upstream's `fixInsertTypeSpecifierForImportDeclaration`:
// ` type` after `import`, the empty-braces split for a default import, and every inline `type` removed
// so the result is never `import type { type T }`.
func (fixer *consistentTypeImportsFixer) insertTypeSpecifierForImportDeclaration(isDefaultImport bool) bool {
	importToken, _ := fixer.importToken()
	fixer.insertBefore(importToken.End, " type")

	moduleStart, _ := fixer.nodeRange(fixer.statement.AsImportDeclaration().ModuleSpecifier)
	if isDefaultImport {
		if openingBrace, found := fixer.tokens.FirstBetween(importToken.End, moduleStart, ast.KindOpenBraceToken); found {
			// import Foo, {} from 'foo'
			comma, found := fixer.tokens.Before(openingBrace.Start)
			if !found || comma.Kind != ast.KindCommaToken {
				return false
			}
			closingBrace, found := fixer.tokens.FirstBetween(openingBrace.End, moduleStart, ast.KindCloseBraceToken)
			if !found {
				return false
			}
			fixer.removeRange(comma.Start, closingBrace.End)
			specifiers := classifyConsistentTypeImportsSpecifiers(fixer.statement)
			if specifiers.total > 1 {
				fixer.insertBefore(fixer.end, "\nimport type"+fixer.text[comma.End:closingBrace.End]+
					" from "+fixer.sourceText()+";")
			}
		}
	}

	for _, named := range classifyConsistentTypeImportsSpecifiers(fixer.statement).namedSpecifiers {
		if named.AsImportSpecifier().IsTypeOnly {
			fixer.removeTypeKeywordOfSpecifier(named)
		}
	}
	return true
}

// removeTypeKeywordOfSpecifier is upstream's `fixRemoveTypeSpecifierFromImportSpecifier`.
func (fixer *consistentTypeImportsFixer) removeTypeKeywordOfSpecifier(specifier *ast.Node) {
	start, end := fixer.nodeRange(specifier)
	typeToken, found := fixer.tokens.FirstBetween(start, end, ast.KindTypeKeyword)
	if !found {
		return
	}
	fixer.removeRange(typeToken.Start, fixer.nextTokenOrComment(typeToken.End))
}

// consistentTypeImportsRemoveTypeFix is upstream's repair under `prefer: no-type-imports`: drop the
// `type` keyword from a declaration, or from one inline specifier.
func consistentTypeImportsRemoveTypeFix(sourceFile *ast.SourceFile, statement *ast.Node, specifier *ast.Node) (rule.Fix, bool) {
	fixer := newConsistentTypeImportsFixer(sourceFile, statement)
	if _, found := fixer.importToken(); !found {
		return rule.Fix{}, false
	}
	if specifier != nil {
		fixer.removeTypeKeywordOfSpecifier(specifier)
		return fixer.compose()
	}
	// import type Foo from 'foo'
	//        ^^^^ remove
	if len(fixer.tokens) < 2 || fixer.tokens[1].Kind != ast.KindTypeKeyword {
		return rule.Fix{}, false
	}
	typeToken := fixer.tokens[1]
	fixer.removeRange(typeToken.Start, fixer.nextTokenOrComment(typeToken.End))
	return fixer.compose()
}
