package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// library-member-provenance requires only ReadsDefaultLibrary. It exposes raw
// declaration kind/name/owner and embedded-library membership, never source text.
func (p *Program) libraryMemberProvenance(source *ast.SourceFile, node *ast.Node, c *checker.Checker, out *fields, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 4 || node != source.AsNode() {
		return "", fmt.Errorf("library-member-provenance requires a SourceFile selector")
	}
	start, e1 := strconv.ParseUint(parts[1], 10, 64)
	end, e2 := strconv.ParseUint(parts[2], 10, 64)
	if e1 != nil || e2 != nil || strconv.FormatUint(start, 10) != parts[1] || strconv.FormatUint(end, 10) != parts[2] {
		return "", fmt.Errorf("invalid library provenance selector")
	}
	_, selected, err := p.exact(source.FileName().AsString(), start, end, parts[3])
	if err != nil {
		return "", err
	}
	symbol := c.GetSymbolAtLocation(selected)
	out.yes(symbol != nil)
	if symbol != nil {
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil {
				return "", fmt.Errorf("library declaration has no source")
			}
			out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.PathKey()))
			out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
			name := ""
			if n := declaration.Name(); n != nil && ast.IsPropertyNameLiteral(n) {
				name = n.Text()
			}
			out.text(name)
			parentKind, parentName := "", ""
			if parent := declaration.Parent; parent != nil {
				parentKind = strings.TrimPrefix(parent.Kind.String(), "Kind")
				if n := parent.Name(); n != nil && ast.IsPropertyNameLiteral(n) {
					parentName = n.Text()
				}
			}
			out.text(parentKind)
			out.text(parentName)
		}
	}
	return out.String(), nil
}
