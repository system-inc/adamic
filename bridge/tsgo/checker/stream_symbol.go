package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Raw symbol declarations and their ancestor shapes. No lint decisions or edits.
func (p *Program) streamSymbol(out *fields, c *checker.Checker, node *ast.Node, source *ast.SourceFile, question string) (string, error) {
	if question == "stream-file" {
		return p.streamFile(out, node, source)
	}
	if question == "stream-program" {
		return p.streamProgram(out, node)
	}
	if question == "stream-signature" {
		return p.streamSignature(out, c, node)
	}
	if question != "stream-symbol" {
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	for _, alias := range []bool{false, true} {
		symbol := c.GetSymbolAtLocation(node)
		if alias && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = c.GetAliasedSymbol(symbol)
		}
		out.yes(symbol != nil)
		if symbol == nil {
			continue
		}
		out.number(p.symbolID(symbol))
		out.number(uint64(symbol.Flags))
		out.text(strings.ToValidUTF8(symbol.Name, "�"))
		out.number(uint64(len(symbol.Declarations)))
		for _, decl := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(decl)
			if file == nil {
				return "", fmt.Errorf("stream declaration has no source")
			}
			out.text(file.FileName().AsString())
			out.yes(file.IsDeclarationFile)
			out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.PathKey()))
			out.yes(ast.IsExternalModule(file))
			var ancestors []*ast.Node
			for at := decl; at != nil; at = at.Parent {
				ancestors = append(ancestors, at)
			}
			out.number(uint64(len(ancestors)))
			for _, at := range ancestors {
				out.text(strings.TrimPrefix(at.Kind.String(), "Kind"))
				out.number(uint64(at.Pos()))
				out.number(uint64(at.End()))
				out.number(uint64(at.Flags))
				name := ""
				if named := at.Name(); named != nil {
					switch named.Kind {
					case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
						name = named.Text()
					}
				}
				out.text(name)
				out.yes(at.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(at))
			}
		}
	}
	return out.String(), nil
}
