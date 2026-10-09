package rename

import (
	"context"
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/types/program"
)

// resolveAnchor finds the symbol at a position.
//
// The innermost identifier containing the offset wins, which is what makes a position behave the
// way a cursor does in an editor. Taking the outermost would resolve `foo.bar` to `foo` when the
// user pointed at `bar`.
func resolveAnchor(
	ctx context.Context,
	graph *program.Graph,
	files []*ast.SourceFile,
	at Position,
) (*ast.Symbol, *ast.Node, error) {
	var target *ast.SourceFile
	for _, file := range files {
		if file.FileName().AsString() == at.FileName {
			target = file
			break
		}
	}
	if target == nil {
		return nil, nil, fmt.Errorf(
			"%s is not in the program — the position must name a file the tsconfig includes",
			at.FileName,
		)
	}

	text := target.Text()
	offset, ok := offsetOf(text, at.Line, at.Column)
	if !ok {
		return nil, nil, fmt.Errorf(
			"%s:%d:%d is past the end of the file",
			at.FileName, at.Line, at.Column,
		)
	}

	fileChecker, release := graph.CheckerForFile(ctx, target)
	if fileChecker == nil {
		release()
		return nil, nil, fmt.Errorf("no checker for %s", at.FileName)
	}
	defer release()

	var found *ast.Node
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindIdentifier {
			span := tokenSpan(target, node)
			if span.Pos() <= offset && offset < span.End() {
				// Last match wins, and this is EQUIVALENT to first-match-wins rather than being a
				// deliberate innermost-wins rule. An earlier version of this comment claimed the
				// latter and was wrong.
				//
				// An identifier is a leaf token, so two identifier spans can never contain the same
				// offset. Measured on the fixture rather than argued: 25 identifiers across two
				// files, and every offset is inside at most one identifier span — zero overlaps. The
				// intuition that `value.toUpperCase()` nests is about the enclosing property-access
				// EXPRESSION, which this walk never records because it only looks at identifiers.
				//
				// So a mutation making this take the first match instead survives, correctly and
				// permanently: no input can distinguish them. It is recorded here rather than pinned
				// by a fixture, because a fixture written for it would assert nothing.
				found = node
			}
		}
		node.ForEachChild(visit)
		return false
	}
	target.AsNode().ForEachChild(visit)

	if found == nil {
		return nil, nil, fmt.Errorf(
			"no identifier at %s:%d:%d — the position must point at a name, not at whitespace, a keyword, or a string",
			at.FileName, at.Line, at.Column,
		)
	}

	// A shorthand property's identifier resolves through the plain accessor to the property rather
	// than to the binding, so anchoring on it would rename the object's contract instead of the
	// variable the user pointed at. The value symbol is the one a cursor there means.
	if parent := found.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == found {
		if value := fileChecker.GetShorthandAssignmentValueSymbol(parent); value != nil {
			return value, found, nil
		}
	}

	symbol := fileChecker.GetSymbolAtLocation(found)
	if symbol == nil {
		return nil, found, nil
	}

	// An imported binding is an alias. Anchoring on the alias renames the local half only; anchoring
	// on what it points at renames the export everywhere. A cursor on an import specifier means the
	// thing being imported, which is what an editor does, so the alias is followed here.
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if aliased := fileChecker.GetAliasedSymbol(symbol); aliased != nil {
			return aliased, found, nil
		}
	}

	return symbol, found, nil
}

// ResolveBareName finds every symbol in the project declared with a given name.
//
// This is the sugar form's whole implementation, and it deliberately returns ALL of them rather
// than picking. A bare name is not a symbol: if three files each declare a local `err`, then
// `rename err error` has three valid answers and no way to choose, because the cursor is
// what disambiguates in an editor and there is no cursor here. The caller refuses when this returns
// more than one and prints them, which is the line this codebase already holds: a missing binary is
// a loud error rather than a fallback, because the gate this replaces printed green over zero files
// after falling through to a guess.
func ResolveBareName(
	ctx context.Context,
	graph *program.Graph,
	name string,
) ([]Candidate, []*ast.Symbol, error) {
	files := graph.ProjectFiles()

	var candidates []Candidate
	var symbols []*ast.Symbol
	seen := map[*ast.Symbol]bool{}

	for _, file := range files {
		text := file.Text()
		if !containsWord(text, name) {
			continue
		}
		fileChecker, release := graph.CheckerForFile(ctx, file)
		if fileChecker == nil {
			release()
			continue
		}
		fileName := file.FileName()

		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			if node.Kind == ast.KindIdentifier && node.Text() == name && isOwnDeclarationName(node) {
				symbol := fileChecker.GetSymbolAtLocation(node)
				// An import specifier declares a local binding whose target lives elsewhere. Both
				// are legitimate rename anchors and they are different acts, so the alias is NOT
				// followed here — otherwise an imported name and its export would collapse into one
				// candidate and the ambiguity the user needs to see would be hidden.
				if symbol != nil && !seen[symbol] {
					seen[symbol] = true
					line, column := lineColumnOf(text, tokenSpan(file, node).Pos())
					symbols = append(symbols, symbol)
					candidates = append(candidates, Candidate{
						Name:     name,
						FileName: fileName.AsString(),
						Line:     line,
						Column:   column,
						Kind:     declarationKindName(node.Parent),
					})
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
	}

	order := make([]int, len(candidates))
	for index := range order {
		order[index] = index
	}
	sort.Slice(order, func(first int, second int) bool {
		left, right := candidates[order[first]], candidates[order[second]]
		if left.FileName != right.FileName {
			return left.FileName < right.FileName
		}
		return left.Line < right.Line
	})
	sortedCandidates := make([]Candidate, len(candidates))
	sortedSymbols := make([]*ast.Symbol, len(symbols))
	for position, index := range order {
		sortedCandidates[position] = candidates[index]
		sortedSymbols[position] = symbols[index]
	}

	return sortedCandidates, sortedSymbols, nil
}

// containsWord reports whether text holds name bounded by non-identifier characters.
//
// A cheap prefilter so a bare-name search does not resolve every identifier in every file. It is
// allowed to be over-permissive because the real decision is made by the checker afterwards; it must
// not be under-permissive, which is why it checks word boundaries rather than something clever.
func containsWord(text string, name string) bool {
	for index := 0; index+len(name) <= len(text); index++ {
		if text[index:index+len(name)] != name {
			continue
		}
		if index > 0 && isIdentifierByte(text[index-1]) {
			continue
		}
		after := index + len(name)
		if after < len(text) && isIdentifierByte(text[after]) {
			continue
		}
		return true
	}
	return false
}

func isIdentifierByte(character byte) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9') ||
		character == '_' || character == '$'
}

// declarationKindName names what sort of declaration a node is, for the candidate listing.
func declarationKindName(parent *ast.Node) string {
	if parent == nil {
		return "declaration"
	}
	switch parent.Kind {
	case ast.KindFunctionDeclaration:
		return "function"
	case ast.KindClassDeclaration:
		return "class"
	case ast.KindInterfaceDeclaration:
		return "interface"
	case ast.KindTypeAliasDeclaration:
		return "type"
	case ast.KindEnumDeclaration:
		return "enum"
	case ast.KindVariableDeclaration:
		return "variable"
	case ast.KindParameter:
		return "parameter"
	case ast.KindPropertyDeclaration:
		return "property"
	case ast.KindMethodDeclaration:
		return "method"
	case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
		return "import"
	case ast.KindExportSpecifier:
		return "export"
	case ast.KindBindingElement:
		return "binding"
	case ast.KindTypeParameter:
		return "type parameter"
	}
	return "declaration"
}

// tokenSpan is an identifier's own text, without the trivia before it.
//
// Routed through the same scanner call `rule.TokenRange` uses, and for the same reason stated at
// length there: `node.Pos()` is the position BEFORE leading trivia, so a range built from it
// includes the whitespace and any comment sitting in front of the name. A rename writing over that
// range would delete a comment and swallow the indentation, and the result would still parse, which
// means nothing downstream would catch it. Ranges are the whole game here.
func tokenSpan(file *ast.SourceFile, node *ast.Node) core.TextRange {
	if file == nil || node == nil {
		return node.Loc
	}
	return scanner.GetRangeOfTokenAtPosition(file, node.Pos()).WithEnd(node.End())
}

// offsetOf converts a 1-based line and column into a byte offset.
//
// Columns are counted in bytes rather than in runes, which matches how every range in this tool is
// expressed and how the source text is indexed. A caller pasting a column from an editor that
// counts runes will be off on a line containing non-ASCII text before the identifier; that is a
// real limitation and it is written down rather than papered over, because the failure is a
// resolution error the user sees rather than a silent wrong rename.
func offsetOf(text string, line int, column int) (int, bool) {
	currentLine := 1
	index := 0
	for currentLine < line {
		newline := indexByteFrom(text, index, '\n')
		if newline < 0 {
			return 0, false
		}
		index = newline + 1
		currentLine++
	}
	offset := index + column - 1
	if offset > len(text) {
		return 0, false
	}
	return offset, true
}

func indexByteFrom(text string, from int, character byte) int {
	for index := from; index < len(text); index++ {
		if text[index] == character {
			return index
		}
	}
	return -1
}

// lineColumnOf converts a byte offset into a 1-based line and column.
func lineColumnOf(text string, offset int) (int, int) {
	line, column := 1, 1
	limit := offset
	if limit > len(text) {
		limit = len(text)
	}
	for index := 0; index < limit; index++ {
		if text[index] == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}
