package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func init() { refusalRules.register(2075, "classMergeRefusal", (*lowering).classMergeRefusal) }

func (l *lowering) classMergeRefusal(node *ast.Node) error {
	if node.Kind != ast.KindInterfaceDeclaration && node.Kind != ast.KindModuleDeclaration {
		return nil
	}
	if !ast.IsIdentifier(node.Name()) || !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return nil
	}
	symbol := l.symbol(node.Name())
	if symbol == nil || symbol.Flags&ast.SymbolFlagsClass == 0 {
		return nil
	}

	var kind, member, fix string
	if node.Kind == ast.KindInterfaceDeclaration {
		members := node.AsInterfaceDeclaration().Members.Nodes
		if len(members) == 0 {
			return nil
		}
		kind = "an interface"
		if name := members[0].Name(); name != nil {
			member = sourceExpression(name)
			fix = "declare " + member + " on the class itself"
		} else {
			member = strings.TrimPrefix(members[0].Kind.String(), "Kind")
			fix = "declare the class's members on the class itself"
		}
	} else {
		member = firstNamespaceMergeMember(node, symbol)
		if member == "" {
			return nil
		}
		kind = "a namespace"
		fix = "declare " + member + " as a static member on the class itself or export it separately"
	}
	return &Refused{
		Where: l.program.Where(node),
		What:  fmt.Sprintf("%s declaration merged into class %s, adding member %s", kind, node.Name().Text(), member),
		Fix:   fix,
	}
}

// The merged export table also contains class statics and exports from other
// namespace blocks. Pick the first export declared by this block in source order.
func firstNamespaceMergeMember(namespace *ast.Node, symbol *ast.Symbol) string {
	var first *ast.Node
	var name string
	for _, exported := range symbol.Exports {
		for _, declaration := range exported.Declarations {
			for parent := declaration.Parent; parent != nil; parent = parent.Parent {
				if parent == namespace {
					if first == nil || declaration.Pos() < first.Pos() {
						first, name = declaration, ast.SymbolName(exported)
					}
					break
				}
			}
		}
	}
	return name
}
