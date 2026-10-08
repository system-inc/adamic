package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	tschecker "github.com/microsoft/TypeScript/tsc/shim/checker"
)

// typeDeclarationOrigins exposes declaration locations, library membership and enclosing ambient
// module names. It makes no allowlist or lint decision.
func (p *Program) typeDeclarationOrigins(question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("type-declaration-origins requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	subject := p.typesByID[id-1]
	symbol := subject.Symbol()
	if symbol == nil {
		if alias := tschecker.Type_alias(subject); alias != nil {
			symbol = alias.Symbol()
		}
	}
	out := &fields{}
	out.number(1)
	out.text("type-declaration-origins")
	out.text(p.Compiler.GetCurrentDirectory().AsString())
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			return "", fmt.Errorf("declaration without source")
		}
		out.text(source.FileName().AsString())
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.PathKey()))
		module := ""
		for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
			if ancestor.Kind == ast.KindModuleDeclaration {
				if name := ancestor.Name(); name != nil && ast.IsStringLiteral(name) {
					module = name.Text()
				}
				break
			}
		}
		out.text(module)
	}
	return out.String(), nil
}
