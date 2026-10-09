package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// Origins and declaration shapes are raw compiler facts, not lint decisions.
func writeSymbolOrigin(out *fields, p *Program, symbol *ast.Symbol) {
	out.yes(symbol != nil)
	if symbol == nil {
		return
	}
	out.text(symbol.Name)
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			panic("symbol declaration has no source")
		}
		out.text(source.FileName().AsString())
		out.yes(source.IsDeclarationFile)
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.PathKey()))
	}
}
func writePropertyInfo(out *fields, symbol *ast.Symbol) {
	present := symbol != nil && symbol.ValueDeclaration != nil
	out.yes(present)
	if !present {
		return
	}
	declaration := symbol.ValueDeclaration
	out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
	var initializer *ast.Node
	var parameters *ast.NodeList
	switch declaration.Kind {
	case ast.KindPropertyDeclaration:
		initializer = declaration.AsPropertyDeclaration().Initializer
	case ast.KindPropertyAssignment:
		initializer = declaration.AsPropertyAssignment().Initializer
	case ast.KindMethodDeclaration:
		parameters = declaration.AsMethodDeclaration().Parameters
	case ast.KindMethodSignature:
		parameters = declaration.AsMethodSignatureDeclaration().Parameters
	}
	initializerKind := ""
	if initializer != nil {
		initializerKind = strings.TrimPrefix(initializer.Kind.String(), "Kind")
		if initializer.Kind == ast.KindFunctionExpression {
			parameters = initializer.AsFunctionExpression().Parameters
		}
	}
	out.text(initializerKind)
	name, annotation := "", ""
	if parameters != nil && len(parameters.Nodes) > 0 {
		first := parameters.Nodes[0].AsParameterDeclaration()
		if first.Name() != nil && first.Name().Kind == ast.KindIdentifier {
			name = first.Name().Text()
		}
		if first.Type != nil {
			annotation = strings.TrimPrefix(first.Type.Kind.String(), "Kind")
		}
	}
	out.text(name)
	out.text(annotation)
}
