package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// nodeHostMember follows @types/node declarations, including aliases and the
// node:* modules' re-exports of fs/path. The shared fs_file loader owns their source.
func (l *lowering) nodeHostMember(node *ast.Node) (string, string) {
	node = ast.SkipParentheses(node)
	symbol := l.symbol(node)
	if symbol == nil {
		return "", ""
	}
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			continue
		}
		filename := strings.ReplaceAll(source.FileName().AsString(), "\\", "/")
		if !strings.HasSuffix(filename, "/node/fs.d.ts") && !strings.HasSuffix(filename, "/node/path.d.ts") {
			continue
		}
		owner := ""
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind != ast.KindModuleDeclaration && parent.Kind != ast.KindClassDeclaration && parent.Kind != ast.KindInterfaceDeclaration {
				continue
			}
			if parent.Name() == nil {
				continue
			}
			name := parent.Name().Text()
			if parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindInterfaceDeclaration {
				owner = name
			}
			if parent.Kind != ast.KindModuleDeclaration {
				continue
			}
			if name == "realpathSync" {
				owner = name
			}
			module := ""
			if name == "fs" || name == "node:fs" {
				module = "node:fs"
			}
			if name == "path" || name == "node:path" {
				module = "node:path"
			}
			if module == "" {
				continue
			}
			if owner == "StatsBase" {
				return "", ""
			}
			member := symbol.Name
			if module == "node:fs" {
				switch member {
				case "native":
					if owner != "realpathSync" {
						member = owner + "." + member
					}
				case "isFile", "isDirectory", "isSymbolicLink", "isBlockDevice", "isCharacterDevice", "isFIFO", "isSocket", "name":
					if owner != "Dirent" {
						member = owner + "." + member
					}
				}
			}
			return module, member
		}
	}
	return "", ""
}
func (l *lowering) nodeFSDirectoryCall(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	module, member := l.nodeHostMember(call.Expression)
	if module == "" {
		return nil, false, nil
	}
	lowered := ir.NodeHostCall{Module: module, Member: member}
	if module == "node:path" {
		switch member {
		case "resolve", "dirname", "join", "relative":
			lowered.Returns = ir.String
			lowered.Throws = nodePathThrows(member)
		default:
			return nil, false, nil
		}
	} else {
		switch member {
		case "readdirSync":
			lowered.Returns = ir.Array
			lowered.Throws = true
		case "realpathSync", "native":
			lowered.Returns = ir.String
			lowered.Throws = true
		case "isFile", "isDirectory", "isSymbolicLink", "isBlockDevice", "isCharacterDevice", "isFIFO", "isSocket":
			if call.Expression.Kind != ast.KindPropertyAccessExpression {
				return nil, true, l.notYet(node, "a detached Dirent method")
			}
			receiver, err := l.expression(call.Expression.AsPropertyAccessExpression().Expression)
			if err != nil {
				return nil, true, err
			}
			lowered.Arguments = append(lowered.Arguments, receiver)
			lowered.Returns = ir.Boolean
		default:
			return nil, false, nil
		}
	}
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, "a spread into a Node host call")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		lowered.Arguments = append(lowered.Arguments, value)
	}
	if module == "node:path" {
		if (member == "dirname" && len(lowered.Arguments) != 1) || (member == "relative" && len(lowered.Arguments) != 2) {
			return nil, true, l.notYet(node, module+"."+member+" with these arguments")
		}
		for _, value := range lowered.Arguments {
			if value.Type() != ir.String {
				return nil, true, l.notYet(node, module+"."+member+" with a non-string path")
			}
		}
	} else if member == "readdirSync" || member == "realpathSync" || member == "native" {
		if len(lowered.Arguments) == 0 || lowered.Arguments[0].Type() != ir.String {
			return nil, true, l.notYet(node, module+"."+member+" with a non-string path")
		}
		if member != "readdirSync" && len(lowered.Arguments) != 1 {
			return nil, true, l.notYet(node, module+"."+member+" with options")
		}
		if member == "readdirSync" && len(call.Arguments.Nodes) > 1 {
			options := ast.SkipParentheses(call.Arguments.Nodes[1])
			if options.Kind != ast.KindObjectLiteralExpression {
				return nil, true, l.notYet(node, module+"."+member+" with non-literal options")
			}
			for _, property := range options.AsObjectLiteralExpression().Properties.Nodes {
				if property.Kind != ast.KindPropertyAssignment {
					return nil, true, l.notYet(node, module+"."+member+" with these options")
				}
				name := property.Name().Text()
				value := ast.SkipParentheses(property.AsPropertyAssignment().Initializer)
				if name == "withFileTypes" && (value.Kind == ast.KindTrueKeyword || value.Kind == ast.KindFalseKeyword) {
					continue
				}
				if name == "encoding" && value.Kind == ast.KindStringLiteral && value.Text() == "utf8" {
					continue
				}
				return nil, true, l.notYet(node, module+"."+member+" option "+name)
			}
		}
	}
	return lowered, true, nil
}

// Realpath function values are used by sys.ts's native fallback selection.
func (l *lowering) nodeFSDirectoryValue(node *ast.Node) (ir.Expression, bool, error) {
	module, member := l.nodeHostMember(node)
	if module == "" {
		return nil, false, nil
	}
	if module == "node:fs" && member == "name" {
		return nil, false, nil
	}
	if module != "node:fs" || (member != "realpathSync" && member != "native") {
		return nil, true, l.notYet(node, module+"."+member+" as a value")
	}
	index := len(l.result.Functions)
	parameter := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "path", Type: ir.String, Function: index})
	value := ir.NodeHostCall{Module: module, Member: member, Arguments: []ir.Expression{ir.Read{Local: parameter, Of: ir.String}}, Returns: ir.String, Throws: true}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "node_realpath", Parameters: []int{parameter}, Returns: ir.String, Closure: true, MayThrow: true, Body: []ir.Statement{ir.Return{Value: value}}})
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: index, node: node})
	return ir.MakeClosure{Function: index}, true, nil
}
