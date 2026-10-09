package rule

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// ExportNameIn reports which export of module expression names, following imports, re-exports and
// `export *` chains through other files, and const aliases of them: "useState" for `useState`, `React.useState`
// or a `use` that a barrel re-exports from "react". It returns "*" for a namespace import of module itself
// and "" when expression is not one of module's exports.
//
// It reads other files, and only their shapes: the module export tables the checker builds, top-level
// import and export specifiers, `export *` statements, and the initializers of top-level consts. All of
// that sits outside every function body, so the shape fingerprint (program.SignatureEntry) covers it, and
// a rule keyed on its imports' shapes may ask this (#9bjjk4a). Two things hold that, rather than care:
//
//   - it hands back a string, never a node, so no caller can descend from its answer into another file;
//   - before it reads any node of a file other than expression's own, it requires that node to sit outside
//     every function body, by the syntax hash's own definition of one, and panics naming the node
//     otherwise. The walk contains the panic to the rule, names it, and does not cache the file, so a later
//     edit that widened these reads is loud on its first run instead of a stale verdict forever.
func ExportNameIn(typeChecker *checker.Checker, expression *ast.Node, module string) string {
	if typeChecker == nil || expression == nil {
		return ""
	}
	resolver := exportOriginResolver{
		checker: typeChecker,
		home:    ast.GetSourceFileOfNode(expression),
		module:  module,
		symbols: map[*ast.Symbol]bool{},
		exports: map[moduleExport]bool{},
	}
	origin := resolver.expression(expression)
	if origin.module != nil && origin.module.Text() == module {
		return origin.name
	}
	return ""
}

type moduleExport struct {
	module *ast.Node
	name   string
}

type exportOriginResolver struct {
	checker *checker.Checker
	// home is the file the expression is in, which a rule's findings key on whole.
	home    *ast.SourceFile
	module  string
	symbols map[*ast.Symbol]bool
	exports map[moduleExport]bool
}

// requireShape panics unless node is in the home file or outside every function body of its own file.
func (r *exportOriginResolver) requireShape(node *ast.Node) {
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == r.home {
		return
	}
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if !ast.IsFunctionLikeDeclaration(ancestor) && !ast.IsClassStaticBlockDeclaration(ancestor) {
			continue
		}
		if body := ancestor.Body(); body != nil && node.Pos() >= body.Pos() && node.End() <= body.End() {
			fileName := ""
			if sourceFile != nil {
				fileName = sourceFile.FileName().AsString()
			}
			panic(fmt.Sprintf("rule.ExportNameIn read a %s at %s:%d, inside a function body in another file, which the shape fingerprint does not cover",
				node.Kind, fileName, node.Pos()))
		}
	}
}

func (r *exportOriginResolver) expression(node *ast.Node) moduleExport {
	if node == nil {
		return moduleExport{}
	}
	r.requireShape(node)
	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		return r.expression(node.Expression())
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		receiver := r.expression(node.Expression())
		var name string
		if node.Kind == ast.KindPropertyAccessExpression && node.Name() != nil {
			name = node.Name().Text()
		} else if node.Kind == ast.KindElementAccessExpression {
			property := node.AsElementAccessExpression().ArgumentExpression
			if property != nil && ast.IsStringLiteralLike(property) {
				name = property.Text()
			}
		}
		if receiver.module != nil && (receiver.name == "*" || receiver.name == "default") && name != "" {
			return r.export(receiver.module, name)
		}
	case ast.KindIdentifier:
		return r.symbol(r.checker.GetSymbolAtLocation(node))
	}
	return moduleExport{}
}

func (r *exportOriginResolver) symbol(symbol *ast.Symbol) moduleExport {
	if symbol == nil || r.symbols[symbol] {
		return moduleExport{}
	}
	r.symbols[symbol] = true
	defer delete(r.symbols, symbol)
	for _, declaration := range symbol.Declarations {
		r.requireShape(declaration)
		switch declaration.Kind {
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindExportSpecifier:
			var container *ast.Node
			for parent := declaration.Parent; parent != nil; parent = parent.Parent {
				if parent.Kind == ast.KindImportDeclaration || parent.Kind == ast.KindExportDeclaration {
					container = parent
					break
				}
			}
			if container == nil {
				continue
			}
			var module *ast.Node
			if container.Kind == ast.KindImportDeclaration {
				module = container.AsImportDeclaration().ModuleSpecifier
			} else {
				module = container.AsExportDeclaration().ModuleSpecifier
			}
			if module == nil || !ast.IsStringLiteralLike(module) {
				continue
			}
			name := "default"
			if declaration.Kind == ast.KindNamespaceImport {
				return moduleExport{module: module, name: "*"}
			}
			if declaration.Kind == ast.KindImportSpecifier || declaration.Kind == ast.KindExportSpecifier {
				if declaration.PropertyName() != nil {
					name = declaration.PropertyName().Text()
				} else if declaration.Name() != nil {
					name = declaration.Name().Text()
				}
			}
			if origin := r.export(module, name); origin.module != nil {
				return origin
			}
		case ast.KindVariableDeclaration:
			if declaration.Parent != nil && declaration.Parent.Kind == ast.KindVariableDeclarationList && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
				if origin := r.expression(declaration.AsVariableDeclaration().Initializer); origin.module != nil {
					return origin
				}
			}
		}
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		return r.symbol(checker.Checker_getImmediateAliasedSymbol(r.checker, symbol))
	}
	return moduleExport{}
}

func (r *exportOriginResolver) moduleSymbol(module *ast.Node) *ast.Symbol {
	symbol := r.checker.GetSymbolAtLocation(module)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = checker.Checker_resolveAlias(r.checker, symbol)
	}
	if symbol == nil || symbol.Flags&ast.SymbolFlagsModule == 0 {
		return nil
	}
	return symbol
}

func (r *exportOriginResolver) exportedSymbol(module *ast.Symbol, name string) *ast.Symbol {
	if module != nil {
		for _, symbol := range r.checker.GetExportsOfModule(module) {
			if symbol.Name == name {
				return symbol
			}
		}
	}
	return nil
}

func (r *exportOriginResolver) resolvedSymbol(symbol *ast.Symbol) *ast.Symbol {
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		return checker.Checker_resolveAlias(r.checker, symbol)
	}
	return symbol
}

func (r *exportOriginResolver) export(module *ast.Node, name string) moduleExport {
	key := moduleExport{module: module, name: name}
	if r.exports[key] {
		return moduleExport{}
	}
	if module.Text() == r.module {
		return key
	}
	r.exports[key] = true
	defer delete(r.exports, key)
	moduleSymbol := r.moduleSymbol(module)
	exported := r.exportedSymbol(moduleSymbol, name)
	if exported == nil {
		return moduleExport{}
	}
	if origin := r.symbol(exported); origin.module != nil {
		return origin
	}
	for _, declaration := range moduleSymbol.Declarations {
		if declaration.Kind != ast.KindSourceFile {
			continue
		}
		for _, statement := range declaration.AsSourceFile().Statements.Nodes {
			if statement.Kind != ast.KindExportDeclaration {
				continue
			}
			r.requireShape(statement)
			star := statement.AsExportDeclaration()
			if star.ExportClause != nil || star.ModuleSpecifier == nil {
				continue
			}
			candidate := r.exportedSymbol(r.moduleSymbol(star.ModuleSpecifier), name)
			if candidate != nil && r.resolvedSymbol(candidate) == r.resolvedSymbol(exported) {
				if origin := r.export(star.ModuleSpecifier, name); origin.module != nil {
					return origin
				}
			}
		}
	}
	return moduleExport{}
}

// ImportBindingKind is how a name came to be in a file's scope, as ImportBindingOf reads it.
type ImportBindingKind uint8

const (
	// ImportBindingGlobal has no declaration in the file: a true global, or none at all.
	ImportBindingGlobal ImportBindingKind = iota
	// ImportBindingModuleLocal is declared in the file and not imported.
	ImportBindingModuleLocal
	ImportBindingDefault
	ImportBindingNamespace
	ImportBindingSpecifier
)

// ImportBinding is where a name in a file comes from: for an import, the module and the export it names.
type ImportBinding struct {
	Kind ImportBindingKind
	// Source is the import's module specifier, for an import.
	Source string
	// Imported is the export an import specifier names, its property name when renamed.
	Imported string
}

// ImportBindingOf reports how symbol, referenced at node, is bound in node's own file: imported, and from
// where, declared there, or neither. It reads only the declarations in node's file and hands back strings,
// so a rule keyed on its imports' shapes may ask it, which the type-reach guard trusts (#9bjjk4a).
func ImportBindingOf(node *ast.Node, symbol *ast.Symbol) ImportBinding {
	binding := ImportBinding{}
	for _, declaration := range DeclarationsIn(ast.GetSourceFileOfNode(node), symbol) {
		switch declaration.Kind {
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
			for container := declaration.Parent; container != nil; container = container.Parent {
				if container.Kind != ast.KindImportDeclaration {
					continue
				}
				module := container.AsImportDeclaration().ModuleSpecifier
				if module == nil || !ast.IsStringLiteralLike(module) {
					break
				}
				binding.Source = module.Text()
				switch declaration.Kind {
				case ast.KindImportSpecifier:
					binding.Kind = ImportBindingSpecifier
					binding.Imported = declaration.Name().Text()
					if declaration.PropertyName() != nil {
						binding.Imported = declaration.PropertyName().Text()
					}
				case ast.KindNamespaceImport:
					binding.Kind = ImportBindingNamespace
				default:
					binding.Kind = ImportBindingDefault
				}
				return binding
			}
		default:
			binding.Kind = ImportBindingModuleLocal
		}
	}
	return binding
}
