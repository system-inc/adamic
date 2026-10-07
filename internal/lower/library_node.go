package lower

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// Register only members with real lowering and runtime support. Host units add
// their names during init; the default is a named NotYet, never erased code.
var implementedNodeMembers = map[string]bool{"node:console.Console.log": true, "node:console.Console.error": true, "node:console.console": true}

func RegisterNodeLibraryMembers(names ...string) {
	for _, name := range names {
		implementedNodeMembers[name] = true
	}
}

func (l *lowering) nodeLibraryMember(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		node = ast.SkipParentheses(node.AsCallExpression().Expression)
	}
	if node.Kind == ast.KindNewExpression {
		node = ast.SkipParentheses(node.AsNewExpression().Expression)
	}
	symbol := l.symbol(node)
	if symbol == nil {
		return ""
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if !load.IsNodeLibrary(file) {
			continue
		}
		if declaration.Kind == ast.KindModuleDeclaration {
			return ""
		}
		module := ""
		owners := []string{}
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if (parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindInterfaceDeclaration) && parent.Name() != nil {
				owners = append([]string{parent.Name().Text()}, owners...)
			}
			if parent.Kind == ast.KindModuleDeclaration && parent.Name() != nil && parent.Name().Kind == ast.KindIdentifier && parent.Name().Text() == "realpathSync" {
				owners = append([]string{parent.Name().Text()}, owners...)
			}
			if parent.Kind == ast.KindModuleDeclaration && parent.Name() != nil && parent.Name().Kind == ast.KindStringLiteral {
				module = parent.Name().Text()
				break
			}
		}
		if module == "" {
			module = strings.TrimSuffix(path.Base(string(file.FileName())), ".d.ts")
		}
		if !strings.HasPrefix(module, "node:") {
			module = "node:" + module
		}
		return strings.Join(append([]string{module}, append(owners, symbol.Name)...), ".")
	}
	return ""
}

func (l *lowering) nodeLibraryRefusal(node *ast.Node) error {
	if l.nodeRequireGlobal(node, "require") {
		return nil
	}
	if _, call := l.nodeRequireCall(node); call {
		return nil
	}
	if !ast.IsExpressionNode(node) || ast.IsPartOfTypeNode(node) {
		return nil
	}
	name := l.nodeLibraryMember(node)
	if ast.IsIdentifier(node) && node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		node = node.Parent
	}
	if strings.HasPrefix(name, "node:globals.Dict.") {
		if node.Kind == ast.KindElementAccessExpression && l.processPath(node.AsElementAccessExpression().Expression) == "process.env" {
			return nil
		}
		if node.Kind == ast.KindPropertyAccessExpression && l.processPath(node.AsPropertyAccessExpression().Expression) == "process.env" {
			return nil
		}
	}
	if (name == "node:stream.Writable.write" || name == "node:net.Socket.write") && l.processPath(ast.SkipParentheses(node)) != "process.stdout.write" {
		if node.Kind != ast.KindCallExpression || l.processPath(node.AsCallExpression().Expression) != "process.stdout.write" {
			return l.notYet(node, name)
		}
	}
	if name != "" && !implementedNodeMembers[name] {
		return l.notYet(node, name)
	}
	return nil
}
