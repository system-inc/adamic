package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// nodeHostMember checks the declaration's ambient module, including imported aliases.
func (l *lowering) nodeHostMember(node *ast.Node) (string, string) {
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return "", ""
	}
	declaration := symbol.Declarations[0]
	if !load.IsPrelude(ast.GetSourceFileOfNode(declaration)) {
		return "", ""
	}
	for parent := declaration.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindModuleDeclaration && parent.Name() != nil {
			name := parent.Name().Text()
			if name == "node:fs" || name == "node:path" {
				return name, symbol.Name
			}
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
		case "resolve", "dirname", "join":
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
		case "isFile", "isDirectory", "isSymbolicLink":
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
	return lowered, true, nil
}

// Realpath function values are used by sys.ts's native fallback selection.
func (l *lowering) nodeFSDirectoryValue(node *ast.Node) (ir.Expression, bool, error) {
	module, member := l.nodeHostMember(node)
	if module != "node:fs" || (member != "realpathSync" && member != "native") {
		return nil, false, nil
	}
	index := len(l.result.Functions)
	parameter := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "path", Type: ir.String, Function: index})
	value := ir.NodeHostCall{Module: module, Member: member, Arguments: []ir.Expression{ir.Read{Local: parameter, Of: ir.String}}, Returns: ir.String, Throws: true}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "node_realpath", Parameters: []int{parameter}, Returns: ir.String, Closure: true, MayThrow: true, Body: []ir.Statement{ir.Return{Value: value}}})
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: index, node: node})
	return ir.MakeClosure{Function: index}, true, nil
}
