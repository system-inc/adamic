package fuzz

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/system-inc/adamic/internal/load"
)

// sourceTree is a program written by a person, as the shrinker sees it through TypeScript's own
// parser: its lists are statements, class and interface members, declarations, arguments, array
// elements and object properties; its compounds are the statements that hold statements; its
// expressions are those the checker gives a type something simpler shares. A candidate is the source
// with one span of a parsed node taken out or replaced, cut at the node's own boundaries, never at a
// line or by a pattern, and it is parsed again before anything runs it.
type sourceTree struct {
	// path is the file's name, which the parser and checker read the language from (.a or .ts).
	path string
	text string
	// parsed is the parse of text, made when first asked for.
	parsed *parsedSource
	// sites are the expressions and what each may become, found by the checker for an earlier text
	// and moved along with every replacement since, or nil until first asked for.
	sites *[]site
}

// parsedSource is what the parser found in a text: every list and compound, outermost first.
type parsedSource struct {
	// clean is whether the text parsed without an error. A candidate that doesn't is never run.
	clean     bool
	lists     []itemList
	compounds []compoundStatement
}

// itemList is a list's items by their spans: start is where the item's leading trivia begins, so a
// comment written on an item goes with it.
type itemList struct {
	starts, ends []int
	// commas is whether a comma separates the items, and trailingComma where a comma after the last
	// one ends, or -1.
	commas        bool
	trailingComma int
	// apart is whether the items are apart, with code between them that stays: the comments.
	apart bool
}

// compoundStatement is a statement that holds statements: the statement's span, without its leading
// trivia, and the spans each of its parts' statements fill.
type compoundStatement struct {
	start, end int
	parts      []span
}

type span struct{ start, end int }

// site is one expression's span and the texts or spans it may be replaced by: a literal of its type,
// or one of its own children that has the same type.
type site struct {
	at       span
	literals []string
	children []span
}

func newSourceTree(path string, text string) *sourceTree {
	return &sourceTree{path: path, text: text}
}

func (t *sourceTree) Source() string { return t.text }

// with is the text with span replaced by replacement, as a new tree whose sites move with the text.
func (t *sourceTree) with(at span, replacement string) *sourceTree {
	text := t.text[:at.start] + replacement + t.text[at.end:]
	result := &sourceTree{path: t.path, text: text}
	if t.sites != nil {
		var moved []site
		shift := len(replacement) - (at.end - at.start)
		for _, existing := range *t.sites {
			if relocated, kept := moveSpan(existing.at, at, shift); kept {
				entry := site{at: relocated, literals: existing.literals}
				for _, child := range existing.children {
					if relocatedChild, keptChild := moveSpan(child, at, shift); keptChild {
						entry.children = append(entry.children, relocatedChild)
					}
				}
				moved = append(moved, entry)
			}
		}
		result.sites = &moved
	}
	return result
}

// withChild replaces an expression with one of its own children, whose sites move with it into the
// parent's place.
func (t *sourceTree) withChild(parent span, child span) *sourceTree {
	result := t.with(parent, t.text[child.start:child.end])
	if t.sites == nil {
		return result
	}
	// with dropped the child's own sites, which were inside the replaced span: put them back, moved to
	// where the child now starts.
	shift := parent.start - child.start
	var inside []site
	for _, existing := range *t.sites {
		if existing.at.start >= child.start && existing.at.end <= child.end && existing.at != parent {
			entry := site{at: span{existing.at.start + shift, existing.at.end + shift}, literals: existing.literals}
			for _, grandchild := range existing.children {
				entry.children = append(entry.children, span{grandchild.start + shift, grandchild.end + shift})
			}
			inside = append(inside, entry)
		}
	}
	merged := append([]site{}, *result.sites...)
	merged = append(merged, inside...)
	sortSites(merged)
	result.sites = &merged
	return result
}

// moveSpan moves a span past an edit of at that changed the text's length by shift. A span inside
// the edit is gone; one around it grows or shrinks with it.
func moveSpan(existing span, at span, shift int) (span, bool) {
	switch {
	case existing.end <= at.start:
		return existing, true
	case existing.start >= at.end:
		return span{existing.start + shift, existing.end + shift}, true
	case existing.start <= at.start && existing.end >= at.end && existing != at:
		return span{existing.start, existing.end + shift}, true
	}
	return span{}, false
}

func sortSites(sites []site) {
	// Parents before children: by start, and the longer first where two start together.
	for index := 1; index < len(sites); index++ {
		for back := index; back > 0; back-- {
			left, right := sites[back-1].at, sites[back].at
			if left.start < right.start || (left.start == right.start && left.end >= right.end) {
				break
			}
			sites[back-1], sites[back] = sites[back], sites[back-1]
		}
	}
}

func (t *sourceTree) parse() *parsedSource {
	if t.parsed != nil {
		return t.parsed
	}
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: parseName(t.path), Path: tspath.Path(parseName(t.path))}, t.text, core.ScriptKindTS)
	parsed := &parsedSource{clean: len(file.Diagnostics()) == 0}
	comments := itemList{apart: true, trailingComma: -1}
	commented := map[int]bool{}
	addComments := func(ranges func(func(ast.CommentRange) bool)) {
		for comment := range ranges {
			if !commented[comment.Pos()] {
				commented[comment.Pos()] = true
				comments.starts = append(comments.starts, comment.Pos())
				comments.ends = append(comments.ends, comment.End())
			}
		}
	}
	queue := []*ast.Node{file.AsNode()}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		addComments(scanner.GetLeadingCommentRanges(nil, t.text, node.Pos()))
		addComments(scanner.GetTrailingCommentRanges(nil, t.text, node.End()))
		for _, list := range nodeLists(node) {
			// An empty list is kept too, so emptying a list leaves every index after it where it was.
			if list.list != nil {
				parsed.lists = append(parsed.lists, t.itemList(list.list, list.commas))
			}
		}
		if parts := statementParts(node); parts != nil && isStatementInList(node) {
			compound := compoundStatement{start: scanner.SkipTrivia(t.text, node.Pos()), end: node.End()}
			for _, part := range parts {
				compound.parts = append(compound.parts, t.partSpan(part))
			}
			parsed.compounds = append(parsed.compounds, compound)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			queue = append(queue, child)
			return false
		})
	}
	addComments(scanner.GetLeadingCommentRanges(nil, t.text, file.EndOfFileToken.Pos()))
	sortApart(&comments)
	parsed.lists = append(parsed.lists, comments)
	t.parsed = parsed
	return parsed
}

// parseName is the name the parser is given: absolute, as it requires, and TypeScript, which a .a
// file is to the parser.
func parseName(path string) string {
	name := "/" + filepath.Base(path)
	if strings.HasSuffix(name, ".a") {
		return name + ".ts"
	}
	return name
}

type namedList struct {
	list   *ast.NodeList
	commas bool
}

// nodeLists is the lists a node holds whose items can go one by one.
func nodeLists(node *ast.Node) []namedList {
	switch node.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
		return []namedList{{node.StatementList(), false}}
	case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindInterfaceDeclaration, ast.KindTypeLiteral:
		return []namedList{{node.MemberList(), false}}
	case ast.KindEnumDeclaration:
		return []namedList{{node.MemberList(), true}}
	case ast.KindCallExpression, ast.KindNewExpression:
		return []namedList{{node.ArgumentList(), true}}
	case ast.KindArrayLiteralExpression:
		return []namedList{{node.ElementList(), true}}
	case ast.KindObjectLiteralExpression:
		return []namedList{{node.PropertyList(), true}}
	case ast.KindVariableDeclarationList:
		return []namedList{{node.AsVariableDeclarationList().Declarations, true}}
	case ast.KindCaseBlock:
		return []namedList{{node.AsCaseBlock().Clauses, false}}
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindConstructor:
		return []namedList{{node.ParameterList(), true}}
	}
	return nil
}

// sortApart puts a list of separate spans in the order they're written.
func sortApart(list *itemList) {
	for index := 1; index < len(list.starts); index++ {
		for back := index; back > 0 && list.starts[back-1] > list.starts[back]; back-- {
			list.starts[back-1], list.starts[back] = list.starts[back], list.starts[back-1]
			list.ends[back-1], list.ends[back] = list.ends[back], list.ends[back-1]
		}
	}
}

func (t *sourceTree) itemList(list *ast.NodeList, commas bool) itemList {
	result := itemList{commas: commas, trailingComma: -1}
	for _, item := range list.Nodes {
		result.starts = append(result.starts, item.Pos())
		result.ends = append(result.ends, item.End())
	}
	if commas && len(list.Nodes) > 0 {
		after := scanner.SkipTrivia(t.text, list.Nodes[len(list.Nodes)-1].End())
		if after < len(t.text) && t.text[after] == ',' {
			result.trailingComma = after + 1
		}
	}
	return result
}

// statementParts is the parts of a statement that hold statements: the branches of an if, a loop's
// body, the blocks of a try, a bare block's own statements, a label's statement.
func statementParts(node *ast.Node) []*ast.Node {
	var parts []*ast.Node
	add := func(part *ast.Node) {
		if part != nil {
			parts = append(parts, part)
		}
	}
	switch node.Kind {
	case ast.KindIfStatement:
		statement := node.AsIfStatement()
		add(statement.ThenStatement)
		add(statement.ElseStatement)
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindLabeledStatement:
		add(node.Statement())
	case ast.KindTryStatement:
		statement := node.AsTryStatement()
		add(statement.TryBlock)
		if statement.CatchClause != nil {
			add(statement.CatchClause.AsCatchClause().Block)
		}
		add(statement.FinallyBlock)
	case ast.KindBlock:
		add(node)
	}
	return parts
}

// isStatementInList is whether a node is an item of a statement list, where statements can stand in
// for it.
func isStatementInList(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}

// partSpan is the text a part's statements fill: inside a block's braces, or the one statement.
func (t *sourceTree) partSpan(part *ast.Node) span {
	if part.Kind == ast.KindBlock {
		list := part.StatementList()
		return span{list.Pos(), list.End()}
	}
	return span{scanner.SkipTrivia(t.text, part.Pos()), part.End()}
}

func (t *sourceTree) Lists() int { return len(t.parse().lists) }

func (t *sourceTree) Items(list int) int { return len(t.parse().lists[list].starts) }

func (t *sourceTree) Without(list, start, end int) Reducible {
	items := t.parse().lists[list]
	if items.apart {
		// Last first, so the spans before each are still where they were.
		result := t
		for index := end - 1; index >= start; index-- {
			result = result.with(span{items.starts[index], items.ends[index]}, "")
		}
		return result
	}
	count := len(items.starts)
	cut := span{items.starts[start], items.ends[end-1]}
	if items.commas {
		switch {
		case end < count && start == 0:
			// Up to the next item's first token: the items, their commas, and the space after.
			cut.end = scanner.SkipTrivia(t.text, items.starts[end])
		case end < count:
			// Up to the next item's start: the items and the commas after them.
			cut.end = items.starts[end]
		case start > 0:
			// To the end, from the end of the item before: the comma before them goes too.
			cut.start = items.ends[start-1]
			if items.trailingComma >= 0 {
				cut.end = items.trailingComma
			}
		case items.trailingComma >= 0:
			cut.end = items.trailingComma
		}
	}
	return t.with(cut, "")
}

func (t *sourceTree) Compounds() int { return len(t.parse().compounds) }

func (t *sourceTree) Inlined(index int) []Reducible {
	compound := t.parse().compounds[index]
	var candidates []Reducible
	for _, part := range compound.parts {
		candidates = append(candidates, t.with(span{compound.start, compound.end}, t.text[part.start:part.end]))
	}
	return candidates
}

func (t *sourceTree) Expressions() int { return len(t.typedSites()) }

func (t *sourceTree) Simpler(index int) []Reducible {
	expression := t.typedSites()[index]
	var candidates []Reducible
	for _, literal := range expression.literals {
		if t.text[expression.at.start:expression.at.end] != literal {
			candidates = append(candidates, t.with(expression.at, literal))
		}
	}
	for _, child := range expression.children {
		candidates = append(candidates, t.withChild(expression.at, child))
	}
	return candidates
}

// typedSites asks the checker, once, which expressions have something simpler of the same type. The
// sites then move with every replacement, so a whole round of simplifying is checked only once.
func (t *sourceTree) typedSites() []site {
	if t.sites != nil {
		return *t.sites
	}
	found := t.findSites()
	t.sites = &found
	return found
}

func (t *sourceTree) findSites() []site {
	path, err := filepath.Abs(t.path)
	if err != nil {
		return nil
	}
	program, err := load.LoadOverlay([]string{path}, map[string]string{path: t.text})
	if err != nil || len(program.Files()) == 0 {
		// A program this checker refuses has no types to go by; the lists and compounds still shrink it.
		return nil
	}
	file := program.Files()[0]
	if file.Text() != t.text {
		return nil
	}
	typeChecker, release := program.Checker(context.Background(), file)
	defer release()
	var sites []site
	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if replaceable(node) {
			typeOf := typeChecker.TypeToString(typeChecker.GetTypeAtLocation(node))
			entry := site{at: span{scanner.SkipTrivia(t.text, node.Pos()), node.End()}}
			switch typeOf {
			case "number":
				entry.literals = []string{"0"}
			case "string":
				entry.literals = []string{"''"}
			case "boolean":
				entry.literals = []string{"true", "false"}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				if ast.IsExpressionNode(child) && typeChecker.TypeToString(typeChecker.GetTypeAtLocation(child)) == typeOf {
					entry.children = append(entry.children, span{scanner.SkipTrivia(t.text, child.Pos()), child.End()})
				}
				return false
			})
			if len(entry.literals) > 0 || len(entry.children) > 0 {
				sites = append(sites, entry)
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(file.AsNode())
	return sites
}

// replaceable is an expression a value can stand in for: not a literal already, not a name being
// declared or a property's name, not a place written to, and not part of a type.
func replaceable(node *ast.Node) bool {
	if !ast.IsExpressionNode(node) || ast.IsPartOfTypeNode(node) || ast.IsAssignmentTarget(node) || ast.IsDeclarationName(node) {
		return false
	}
	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindThisKeyword, ast.KindSuperKeyword:
		return false
	}
	if parent := node.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			return parent.Name() != node
		case ast.KindCallExpression, ast.KindNewExpression, ast.KindDecorator, ast.KindTaggedTemplateExpression:
			// The callee: no literal or child of the same type calls anything.
			return parent.Expression() != node
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindMethodDeclaration, ast.KindEnumMember:
			return parent.Name() != node
		}
	}
	return true
}
