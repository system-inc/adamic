package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func init() {
	RegisterNodeLibraryMembers("node:fs.symlinkSync", "node:fs.readdirSync", "node:fs.realpathSync", "node:fs.native",
		"node:fs.Dirent.name", "node:fs.Dirent.isFile", "node:fs.Dirent.isDirectory", "node:fs.Dirent.isSymbolicLink",
		"node:path.resolve", "node:path.dirname", "node:path.join", "node:path.relative", "node:path.basename")
}

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
		case "resolve", "dirname", "join", "relative", "basename":
			lowered.Returns = ir.String
			lowered.Throws = nodePathThrows(member)
		default:
			return nil, false, nil
		}
	} else {
		switch member {
		case "symlinkSync":
			lowered.Returns = ir.Number
			lowered.Throws = true
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
	if member == "symlinkSync" {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		discarded := outer.Parent != nil && outer.Parent.Kind == ast.KindExpressionStatement
		returned := outer.Parent != nil && outer.Parent.Kind == ast.KindReturnStatement && l.function != nil && l.function.Returns == 0
		if !discarded && !returned {
			return nil, true, l.notYet(node, "node:fs.symlinkSync used as a value")
		}
		if len(call.Arguments.Nodes) < 2 || len(call.Arguments.Nodes) > 3 {
			return nil, true, l.notYet(node, "node:fs.symlinkSync with these arguments")
		}
		if len(call.Arguments.Nodes) == 3 {
			typeArgument := ast.SkipParentheses(call.Arguments.Nodes[2])
			if typeArgument.Kind != ast.KindNullKeyword && !(ast.IsIdentifier(typeArgument) && typeArgument.Text() == "undefined" && l.checker.GetTypeAtLocation(typeArgument).Flags()&checker.TypeFlagsUndefined != 0) &&
				!(typeArgument.Kind == ast.KindStringLiteral && (typeArgument.Text() == "file" || typeArgument.Text() == "dir" || typeArgument.Text() == "junction")) {
				return nil, true, l.notYet(node, "node:fs.symlinkSync type other than constant file, dir, junction, null or undefined")
			}
		}
	}
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, module+"."+member+" with a spread")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		lowered.Arguments = append(lowered.Arguments, value)
	}
	if member == "symlinkSync" {
		if lowered.Arguments[0].Type() != ir.String || lowered.Arguments[1].Type() != ir.String {
			return nil, true, l.notYet(node, "node:fs.symlinkSync with non-string target or path")
		}
		// POSIX ignores the validated type. Literal arguments have no effects.
		lowered.Arguments = lowered.Arguments[:2]
	}
	if module == "node:path" {
		if member == "basename" {
			if len(lowered.Arguments) < 1 || len(lowered.Arguments) > 2 {
				return nil, true, l.notYet(node, "node:path.basename with these arguments")
			}
			if len(lowered.Arguments) == 2 {
				if _, missing := lowered.Arguments[1].(ir.Undefined); missing {
					lowered.Arguments = lowered.Arguments[:1]
				}
			}
		}
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

// Refuse broader declared overloads before argument invariance tries to view
// native internal-slot values (Buffer or URL) as general PathLike objects.
func (l *lowering) nodeFSDirectorySignature(node *ast.Node) error {
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	module, member := l.nodeHostMember(call.Expression)
	if module == "node:path" && member == "basename" {
		if len(call.Arguments.Nodes) < 1 || len(call.Arguments.Nodes) > 2 {
			return l.notYet(node, "node:path.basename with these arguments")
		}
		for _, argument := range call.Arguments.Nodes {
			if argument.Kind == ast.KindSpreadElement {
				return l.notYet(node, "node:path.basename with a spread")
			}
		}
		return nil
	}
	if module != "node:fs" || member != "symlinkSync" {
		return nil
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || (outer.Parent.Kind != ast.KindExpressionStatement && outer.Parent.Kind != ast.KindReturnStatement) {
		return l.notYet(node, "node:fs.symlinkSync used as a value")
	}
	if len(call.Arguments.Nodes) < 2 || len(call.Arguments.Nodes) > 3 {
		return l.notYet(node, "node:fs.symlinkSync with these arguments")
	}
	for _, argument := range call.Arguments.Nodes[:2] {
		if l.checker.GetTypeAtLocation(argument).Flags()&checker.TypeFlagsStringLike == 0 {
			return l.notYet(node, "node:fs.symlinkSync with non-string target or path")
		}
	}
	return nil
}
