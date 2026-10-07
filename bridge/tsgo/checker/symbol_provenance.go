package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Raw symbol identity, declaration locations and enclosing ambient modules.
// Alias following is requested explicitly; no rule predicate runs here.
func (p *Program) symbolProvenance(out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) error {
	if question != "symbol-provenance" && question != "symbol-provenance\nalias" {
		return fmt.Errorf("invalid symbol-provenance question")
	}
	if node.Kind != ast.KindIdentifier && node.Kind != ast.KindPrivateIdentifier {
		return fmt.Errorf("symbol-provenance requires an identifier")
	}
	symbol := c.GetSymbolAtLocation(node)
	if question == "symbol-provenance\nalias" && symbol != nil {
		symbol = checker.SkipAlias(symbol, c)
	}
	out.number(p.symbolID(symbol))
	if symbol == nil {
		return nil
	}
	out.text(strings.ToValidUTF8(symbol.Name, "�"))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return fmt.Errorf("symbol declaration has no source")
		}
		out.text(file.FileName())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.yes(file.IsDeclarationFile)
		module := ""
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindModuleDeclaration && parent.Name() != nil && parent.Name().Kind == ast.KindStringLiteral {
				module = parent.Name().Text()
				break
			}
		}
		out.text(module)
	}
	value := symbol.ValueDeclaration
	out.yes(value != nil)
	if value != nil {
		file := ast.GetSourceFileOfNode(value)
		if file == nil {
			return fmt.Errorf("symbol value declaration has no source")
		}
		out.text(file.FileName())
		out.text(strings.TrimPrefix(value.Kind.String(), "Kind"))
		out.number(uint64(value.Pos()))
		out.number(uint64(value.End()))
		out.yes(ast.IsVarConst(value))
	}
	out.number(uint64(len(source.Imports())))
	for _, specifier := range source.Imports() {
		out.text(specifier.Text())
		resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(source, specifier)
		valid := resolved != nil && resolved.IsResolved()
		out.yes(valid)
		if valid {
			out.text(resolved.ResolvedFileName)
			out.yes(resolved.IsExternalLibraryImport)
			out.text(resolved.PackageId.Name)
		}
	}
	return nil
}

func (p *Program) finishSymbolProvenance(out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if err := p.symbolProvenance(out, c, source, node, question); err != nil {
		return "", err
	}
	return out.String(), nil
}
