package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// This host slice recognizes library declarations, including import aliases.
// It borrows only strings and does not expose a mutable process object.
func (l *lowering) nodeHost(node *ast.Node) (ir.Expression, bool, error) {
	name := l.nodeLibraryMember(node)
	operation, of := "", ir.String
	switch name {
	case "node:path.join":
		operation = "host_join"
	case "node:os.tmpdir":
		operation = "host_tmpdir"
	case "node:process.Process.cwd", "node:process.cwd":
		operation = "host_cwd"
	case "node:process.Process.chdir", "node:process.chdir":
		operation, of = "host_chdir", ir.Number
	default:
		return nil, false, nil
	}
	call := node.AsCallExpression()
	if operation != "host_join" && len(call.Arguments.Nodes) != map[string]int{"host_tmpdir": 0, "host_cwd": 0, "host_chdir": 1}[operation] {
		return nil, true, l.notYet(node, name+" outside the implemented host signature")
	}
	if operation == "host_chdir" {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		discarded := outer.Parent != nil && outer.Parent.Kind == ast.KindExpressionStatement
		returned := outer.Parent != nil && outer.Parent.Kind == ast.KindReturnStatement && l.function != nil && l.function.Returns == 0
		if !discarded && !returned {
			return nil, true, l.notYet(node, name+" used as a value")
		}
	}
	args := []ir.Expression{}
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, name+" with spread arguments")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.String {
			return nil, true, l.notYet(argument, name+" with non-string arguments")
		}
		args = append(args, value)
	}
	return ir.NodeFSFile{Operation: operation, Arguments: args, Of: of}, true, nil
}

func init() {
	RegisterNodeLibraryMembers("node:path.join", "node:os.tmpdir", "node:process.Process.cwd", "node:process.Process.chdir", "node:process.cwd", "node:process.chdir", "node:process.process", "node:globals.process")
}
