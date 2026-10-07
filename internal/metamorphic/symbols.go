package metamorphic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// reference is one identifier that names a symbol.
type reference struct {
	node *ast.Node
	// shorthand is a name in an object literal's { name }, which spells the property and reads the
	// variable at once, so it can't be renamed alone.
	shorthand bool
	// written is a reference the program assigns (=, +=, ++).
	written bool
}

// references is every identifier the program's file spells, by the symbol the checker resolves it
// to.
type references struct {
	bySymbol map[*ast.Symbol][]reference
	symbolOf map[*ast.Node]*ast.Symbol
}

func (p *program) references() *references {
	result := &references{bySymbol: map[*ast.Symbol][]reference{}, symbolOf: map[*ast.Node]*ast.Symbol{}}
	visit(p.file.AsNode(), func(node *ast.Node) bool {
		if node.Kind != ast.KindIdentifier {
			return true
		}
		entry := reference{node: node}
		var symbol *ast.Symbol
		if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
			symbol = p.checker.GetShorthandAssignmentValueSymbol(parent)
			entry.shorthand = true
		} else {
			symbol = p.checker.GetSymbolAtLocation(node)
		}
		if symbol == nil {
			return true
		}
		entry.written = ast.IsAssignmentTarget(node)
		result.bySymbol[symbol] = append(result.bySymbol[symbol], entry)
		result.symbolOf[node] = symbol
		return true
	})
	return result
}

// reads is a symbol's references that aren't its declaration's name, in the order they're written.
func (r *references) reads(symbol *ast.Symbol) []reference {
	var result []reference
	for _, entry := range r.bySymbol[symbol] {
		if ast.IsDeclarationName(entry.node) && !entry.shorthand {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// constant says whether a symbol is never assigned after it's declared: a const, a function
// declaration, or a parameter, let or catch variable nothing writes.
func (r *references) constant(symbol *ast.Symbol) bool {
	declaration := symbol.ValueDeclaration
	if declaration == nil {
		return false
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return true
	case ast.KindVariableDeclaration, ast.KindParameter:
		if declaration.Name() == nil || declaration.Name().Kind != ast.KindIdentifier {
			return false
		}
	default:
		return false
	}
	for _, entry := range r.bySymbol[symbol] {
		if entry.written {
			return false
		}
	}
	return true
}

// topStatement is the statement of the file a node is part of, or nil for the file itself.
func topStatement(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil {
		if node.Parent.Kind == ast.KindSourceFile {
			return node
		}
		node = node.Parent
	}
	return nil
}

// moduleScoped says whether a declaration is in the module's own scope, where a class declared at
// the top level sees it: not inside a function, a block or a loop.
func moduleScoped(declaration *ast.Node) bool {
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		list := declaration.Parent
		return list != nil && list.Parent != nil && list.Parent.Kind == ast.KindVariableStatement && list.Parent.Parent != nil && list.Parent.Parent.Kind == ast.KindSourceFile
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return declaration.Parent != nil && declaration.Parent.Kind == ast.KindSourceFile
	case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportEqualsDeclaration:
		return true
	}
	return false
}

// within says whether node is inside (or is) container.
func within(node *ast.Node, container *ast.Node) bool {
	for ; node != nil; node = node.Parent {
		if node == container {
			return true
		}
	}
	return false
}

// inFile says whether a declaration is in the program's own file, not the library's or the
// prelude's.
func (p *program) inFile(declaration *ast.Node) bool {
	return declaration != nil && ast.GetSourceFileOfNode(declaration) == p.file
}
