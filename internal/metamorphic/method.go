package metamorphic

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// methodSites are runs of statements that can live in a method of their own: a class declared at the
// top level (stage 0 lowers no class inside a function) with one method, run, whose body is the run,
// called once where the run was, new Block().run(...). What the run reads from an enclosing function
// it receives as parameters with the same names, so its text is unchanged; that's only sound for what
// nothing assigns (a const, a parameter nothing writes), so a variable something reassigns has to be
// declared inside the run, or the statement reading it stays where it is. The module's own variables
// need no parameter: the class is at the top level, beside them. A run can't leave early (return, or
// a break or continue to a loop outside it), can't use this or super, holds no function or class
// declaration, and can't declare what's read after it.
func methodSites(p *program) []site {
	found := p.references()
	var sites []site
	visit(p.file.AsNode(), func(node *ast.Node) bool {
		if node.Kind != ast.KindSourceFile && node.Kind != ast.KindBlock {
			return true
		}
		statements := node.StatementList()
		if statements == nil || len(statements.Nodes) == 0 {
			return true
		}
		for _, run := range p.runs(node, statements.Nodes, found) {
			if change, ok := p.methodSite(run, found); ok {
				sites = append(sites, change)
			}
		}
		return true
	})
	return sites
}

// runs are the longest runs of a list's statements that can move together.
func (p *program) runs(container *ast.Node, statements []*ast.Node, found *references) [][]*ast.Node {
	barrier := make([]bool, len(statements))
	needs := make([][]int, len(statements))
	index := map[*ast.Node]int{}
	for position, statement := range statements {
		index[statement] = position
	}
	for position, statement := range statements {
		movable, needed := p.movable(statement, container, index, found)
		barrier[position], needs[position] = !movable, needed
	}
	// A statement that declares something read outside its run stays, and so does one that reads a
	// variable something reassigns declared outside its run; the runs are cut again, to a fixed point.
	for changed := true; changed; {
		changed = false
		for start := 0; start < len(statements); {
			if barrier[start] {
				start++
				continue
			}
			end := start
			for end < len(statements) && !barrier[end] {
				end++
			}
			first, last := statements[start], statements[end-1]
			for position := start; position < end; position++ {
				outside := p.declaresReadOutside(statements[position], first.Pos(), last.End(), found)
				for _, needed := range needs[position] {
					outside = outside || needed < start || needed >= end
				}
				if outside {
					barrier[position] = true
					changed = true
				}
			}
			start = end
		}
	}
	var result [][]*ast.Node
	for start := 0; start < len(statements); {
		if barrier[start] {
			start++
			continue
		}
		end := start
		for end < len(statements) && !barrier[end] {
			end++
		}
		run := statements[start:end]
		for _, statement := range run {
			if !declarationOnly(statement) {
				result = append(result, run)
				break
			}
		}
		start = end
	}
	return result
}

func declarationOnly(statement *ast.Node) bool {
	switch statement.Kind {
	case ast.KindVariableStatement, ast.KindFunctionDeclaration, ast.KindEmptyStatement:
		return true
	}
	return false
}

// movable says whether one statement could be part of a run, and which of its list's statements
// declare a variable it reads that something reassigns: the run has to hold them too, since a
// parameter would be a copy.
func (p *program) movable(statement *ast.Node, container *ast.Node, index map[*ast.Node]int, found *references) (bool, []int) {
	switch statement.Kind {
	case ast.KindImportDeclaration, ast.KindImportEqualsDeclaration, ast.KindExportDeclaration, ast.KindExportAssignment,
		ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration, ast.KindFunctionDeclaration:
		// A function declaration inside the method would be one inside a function, which stage 0
		// doesn't lower yet.
		return false, nil
	}
	if ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport|ast.ModifierFlagsAmbient) {
		return false, nil
	}
	ok := true
	var needs []int
	var walk func(node *ast.Node, loops int, switches int, labels []string)
	walk = func(node *ast.Node, loops int, switches int, labels []string) {
		if !ok {
			return
		}
		switch node.Kind {
		case ast.KindReturnStatement, ast.KindThisKeyword, ast.KindSuperKeyword, ast.KindYieldExpression, ast.KindAwaitExpression,
			ast.KindMetaProperty, ast.KindClassDeclaration, ast.KindClassExpression:
			ok = false
			return
		case ast.KindBreakStatement, ast.KindContinueStatement:
			label := node.Label()
			switch {
			case label != nil:
				inside := false
				for _, name := range labels {
					if name == label.Text() {
						inside = true
					}
				}
				ok = ok && inside
			case node.Kind == ast.KindBreakStatement:
				ok = ok && loops+switches > 0
			default:
				ok = ok && loops > 0
			}
			return
		case ast.KindIdentifier:
			if node.Text() == "arguments" {
				ok = false
				return
			}
			symbol := found.symbolOf[node]
			if symbol == nil || !variableLike(symbol) || ast.IsPartOfTypeNode(node) {
				return
			}
			declaration := symbol.ValueDeclaration
			if declaration == nil || !p.inFile(declaration) || within(declaration, statement) || moduleScoped(declaration) {
				return
			}
			if found.constant(symbol) {
				return
			}
			// Something reassigns it: a parameter would be a copy, so the run has to hold its
			// declaration, which has to be one of the list's own statements.
			for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
				if ancestor.Parent == container {
					if position, found := index[ancestor]; found {
						needs = append(needs, position)
						return
					}
				}
			}
			ok = false
			return
		}
		if ast.IsFunctionLike(node) && node.Kind != ast.KindArrowFunction {
			// A function of its own: its returns and its this are its own. Its reads of the enclosing
			// scope still count.
			node.ForEachChild(func(child *ast.Node) bool {
				walkReads(child, func(identifier *ast.Node) {
					walk(identifier, 0, 0, nil)
				})
				return false
			})
			return
		}
		if node.Kind == ast.KindArrowFunction {
			node.ForEachChild(func(child *ast.Node) bool {
				walk(child, 0, 0, nil)
				return false
			})
			return
		}
		nextLoops, nextSwitches, nextLabels := loops, switches, labels
		switch node.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement:
			nextLoops++
		case ast.KindSwitchStatement:
			nextSwitches++
		case ast.KindLabeledStatement:
			nextLabels = append(append([]string{}, labels...), node.Label().Text())
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child, nextLoops, nextSwitches, nextLabels)
			return false
		})
	}
	walk(statement, 0, 0, nil)
	return ok, needs
}

// variableLike says whether a symbol is a variable, a parameter or a function: what a run reads by
// name from the scopes around it.
func variableLike(symbol *ast.Symbol) bool {
	return symbol.Flags&(ast.SymbolFlagsVariable|ast.SymbolFlagsFunction) != 0
}

// walkReads calls read for every identifier under node.
func walkReads(node *ast.Node, read func(*ast.Node)) {
	visit(node, func(inner *ast.Node) bool {
		if inner.Kind == ast.KindIdentifier {
			read(inner)
		}
		return true
	})
}

// declaresReadOutside says whether a statement declares a name something outside start..end reads.
func (p *program) declaresReadOutside(statement *ast.Node, start int, end int, found *references) bool {
	var names []*ast.Node
	switch statement.Kind {
	case ast.KindVariableStatement:
		for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			visit(declaration.Name(), func(node *ast.Node) bool {
				if node.Kind == ast.KindIdentifier {
					names = append(names, node)
				}
				return true
			})
		}
	case ast.KindFunctionDeclaration:
		if statement.Name() != nil {
			names = append(names, statement.Name())
		}
	default:
		return false
	}
	for _, name := range names {
		symbol := found.symbolOf[name]
		if symbol == nil {
			return true
		}
		for _, entry := range found.bySymbol[symbol] {
			if entry.node.Pos() < start || entry.node.End() > end {
				return true
			}
		}
	}
	return false
}

// methodSite is the edits that move a run into a class's method.
func (p *program) methodSite(run []*ast.Node, found *references) (site, bool) {
	first, last := run[0], run[len(run)-1]
	start, end := p.start(first), last.End()
	// The run's parameters: what it reads that an enclosing function declared, by first read.
	var parameters []string
	var arguments []string
	seen := map[*ast.Symbol]bool{}
	for _, statement := range run {
		failed := false
		walkReads(statement, func(identifier *ast.Node) {
			symbol := found.symbolOf[identifier]
			if failed || symbol == nil || seen[symbol] || !variableLike(symbol) || ast.IsPartOfTypeNode(identifier) {
				return
			}
			declaration := symbol.ValueDeclaration
			if declaration == nil || !p.inFile(declaration) || moduleScoped(declaration) || (declaration.Pos() >= start && declaration.End() <= end) {
				return
			}
			if ast.IsDeclarationName(identifier) && !(identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment) {
				return
			}
			seen[symbol] = true
			written := p.checker.TypeToStringEx(p.checker.GetTypeOfSymbolAtLocation(symbol, identifier), nil, checker.TypeFormatFlagsNoTruncation, nil)
			if written == "" || strings.Contains(written, "...") {
				failed = true
				return
			}
			parameters = append(parameters, identifier.Text()+": "+written)
			arguments = append(arguments, identifier.Text())
		})
		if failed {
			return site{}, false
		}
	}
	name := p.fresh("MetamorphicBlock")
	body := p.text[start:end]
	class := "class " + name + " {\n\trun(" + strings.Join(parameters, ", ") + "): void {\n\t\t" + body + "\n\t}\n}"
	call := "new " + name + "().run(" + strings.Join(arguments, ", ") + ");"
	return site{edits: []edit{{start, end, call}}, header: []declaration{{name: name, text: class}}}, true
}
