package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw union constituent declarations and their AST ancestry. The optional
// awaited operation is a checker operation, not an outcome classification.
// Child lists contain direct type-syntax children, excluding declarations and
// statements; a SourceFile contributes its identity, not its whole body.
func (p *Program) inspectTypeDeclarationAncestry(c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 3 || parts[0] != "type-declaration-ancestry" || (parts[2] != "0" && parts[2] != "1") {
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	subject := p.typesByID[id-1]
	if parts[2] == "1" {
		subject = checker.Checker_getAwaitedType(c, subject)
	}
	out := &fields{}
	out.number(1)
	out.text(parts[0])
	constituents := []*checker.Type{}
	if subject != nil {
		constituents = append(constituents, subject)
		if subject.Flags()&checker.TypeFlagsUnion != 0 {
			constituents = subject.Types()
		}
	}
	out.number(uint64(len(constituents)))
	for _, part := range constituents {
		symbol := part.Symbol()
		if symbol == nil {
			out.number(0)
			continue
		}
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			source := ast.GetSourceFileOfNode(declaration)
			if source == nil {
				return "", fmt.Errorf("type declaration has no source")
			}
			out.text(string(source.FileName()))
			ancestors := []*ast.Node{}
			for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
				ancestors = append(ancestors, ancestor)
			}
			out.number(uint64(len(ancestors)))
			for _, ancestor := range ancestors {
				out.text(strings.TrimPrefix(ancestor.Kind.String(), "Kind"))
				name := ""
				if n := ancestor.Name(); n != nil {
					name = n.Text()
				}
				out.text(name)
				out.number(uint64(ancestor.Pos()))
				out.number(uint64(ancestor.End()))
				children := []*ast.Node{}
				ancestor.ForEachChild(func(child *ast.Node) bool {
					if ast.IsTypeNode(child) {
						children = append(children, child)
					}
					return false
				})
				out.number(uint64(len(children)))
				for _, child := range children {
					out.text(strings.TrimPrefix(child.Kind.String(), "Kind"))
					out.number(uint64(child.Pos()))
					out.number(uint64(child.End()))
				}
			}
		}
	}
	return out.String(), nil
}
