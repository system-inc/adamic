package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// resolved-signature v1 exposes the chosen return type and declaration ancestry.
// It deliberately carries no foreign source/body text: those require askFile.
func (p *Program) resolvedSignature(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "resolved-signature" || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) {
		return "", fmt.Errorf("resolved-signature requires a call")
	}
	signature := c.GetResolvedSignature(node)
	out.yes(signature != nil)
	if signature == nil {
		return out.String(), nil
	}
	graph := &graph{program: p}
	out.number(graph.id(c.GetReturnTypeOfSignature(signature)))
	declaration := signature.Declaration()
	out.yes(declaration != nil)
	if declaration == nil {
		return out.String(), nil
	}
	if err := p.writeSignatureDeclaration(out, declaration); err != nil {
		return "", err
	}
	var parents []*ast.Node
	for parent := declaration.Parent; parent != nil; parent = parent.Parent {
		parents = append(parents, parent)
	}
	out.number(uint64(len(parents)))
	for _, parent := range parents {
		out.text(strings.TrimPrefix(parent.Kind.String(), "Kind"))
		name, kind := "", ""
		if named := parent.Name(); named != nil {
			kind = strings.TrimPrefix(named.Kind.String(), "Kind")
			if ast.IsPropertyNameLiteral(named) || named.Kind == ast.KindPrivateIdentifier {
				name = named.Text()
			}
		}
		out.text(name)
		out.text(kind)
	}
	return out.String(), nil
}

// The signature question needs declaration shape, not JSDoc or parameter text.
// Fields are path, kind, byte span, flags, declaration-file flag, parent kind,
// parent name and byte span. No default-library query or source/body payload.
func (p *Program) writeSignatureDeclaration(out *fields, declaration *ast.Node) error {
	source := ast.GetSourceFileOfNode(declaration)
	if source == nil {
		return fmt.Errorf("signature declaration has no source")
	}
	out.text(source.FileName().AsString())
	out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
	out.number(uint64(declaration.Pos()))
	out.number(uint64(declaration.End()))
	out.number(uint64(declaration.Flags))
	out.yes(source.IsDeclarationFile)
	parentKind, parentName := "", ""
	first, last := uint64(0), uint64(0)
	if parent := declaration.Parent; parent != nil {
		parentKind = strings.TrimPrefix(parent.Kind.String(), "Kind")
		first, last = uint64(parent.Pos()), uint64(parent.End())
		if name := parent.Name(); name != nil && (ast.IsPropertyNameLiteral(name) || name.Kind == ast.KindPrivateIdentifier) {
			parentName = name.Text()
		}
	}
	out.text(parentKind)
	out.text(parentName)
	out.number(first)
	out.number(last)
	return nil
}
