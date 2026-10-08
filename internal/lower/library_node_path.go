package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func init() {
	for _, member := range []string{"resolve", "dirname", "join", "normalize", "relative", "basename", "extname", "isAbsolute", "sep"} {
		RegisterNodeLibraryMembers("node:path."+member, "node:path.PlatformPath."+member)
	}
}

// Only resolve and relative consult cwd; lexical calls on proven strings cannot throw.
func nodePathThrows(member string) bool { return member == "resolve" || member == "relative" }

// Inspect unresolved alias declarations as well: @types/node exports posix and
// win32 as aliases of the same namespace, so resolving a symbol loses the platform.
func (l *lowering) nodePathRefusal(node *ast.Node) error {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsTypeNode(parent) {
			return nil
		}
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) && symbol.Name == "win32" {
			return l.notYet(node, "node:path.win32: Windows path semantics are not implemented; use node:path.posix")
		}
		for parent := declaration; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindImportDeclaration {
				module := parent.ModuleSpecifier()
				if module != nil && (module.Text() == "node:path/win32" || module.Text() == "path/win32") {
					return l.notYet(node, "node:path.win32: Windows path semantics are not implemented; use node:path.posix")
				}
			}
			if parent.Kind == ast.KindImportSpecifier {
				spec := parent.AsImportSpecifier()
				name := spec.Name().Text()
				if spec.PropertyName != nil {
					name = spec.PropertyName.Text()
				}
				if name == "win32" {
					for owner := parent.Parent; owner != nil; owner = owner.Parent {
						if owner.Kind == ast.KindImportDeclaration {
							module := owner.ModuleSpecifier()
							if module != nil && (module.Text() == "node:path" || module.Text() == "path") {
								return l.notYet(node, "node:path.win32: Windows path semantics are not implemented; use node:path.posix")
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (l *lowering) nodePathCall(node *ast.Node, member string) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	result := ir.NodeHostCall{Module: "node:path", Member: member, Returns: ir.String, Throws: nodePathThrows(member)}
	switch member {
	case "resolve", "join", "dirname", "normalize", "relative", "basename", "extname":
	case "isAbsolute":
		result.Returns = ir.Boolean
	default:
		return nil, false, nil
	}
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, "node:path."+member+" with a spread")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		result.Arguments = append(result.Arguments, value)
	}
	count := len(result.Arguments)
	if (member == "basename" && (count < 1 || count > 2)) ||
		(member == "relative" && count != 2) ||
		((member == "dirname" || member == "normalize" || member == "extname" || member == "isAbsolute") && count != 1) {
		return nil, true, l.notYet(node, "node:path."+member+" with these arguments")
	}
	if member == "basename" && count == 2 {
		if _, missing := result.Arguments[1].(ir.Undefined); missing {
			result.Arguments = result.Arguments[:1]
		}
	}
	for _, value := range result.Arguments {
		if value.Type() != ir.String {
			return nil, true, l.notYet(node, "node:path."+member+" with a non-string path")
		}
	}
	return result, true, nil
}
