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

// nodeEnvironmentAccess recognizes the library declaration, never a receiver's spelling.
func (l *lowering) nodeEnvironmentAccess(node *ast.Node) (*ast.Node, bool) {
	node = ast.SkipParentheses(node)
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = node.AsElementAccessExpression().Expression
	default:
		return nil, false
	}
	return receiver, l.nodeLibraryMember(receiver) == "node:process.Process.env"
}

func (l *lowering) nodeEnvironmentReadRefusal(node *ast.Node) error {
	const member = "node:process.Process.env"
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression {
		return l.notYet(node, member+" deleted")
	}
	if ast.IsAssignmentTarget(outer) {
		return l.notYet(node, member+" written")
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return l.notYet(node, member+" used as a value (optional chain)")
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		if ir.IsNodeEnvironmentPrototypeKey(node.Name().Text()) {
			return l.notYet(node, member+" literal prototype key "+node.Name().Text()+" (fix: use process.env['NAME'] for a real variable, or a plain object for the method)")
		}
	} else {
		key := node.AsElementAccessExpression().ArgumentExpression
		if of, known := l.representation(l.checker.GetTypeAtLocation(key)); !known || of != ir.String {
			return l.notYet(node, member+" read with a non-string key")
		}
	}
	return nil
}

func (l *lowering) nodeEnvironmentRead(node *ast.Node) (ir.Expression, bool, error) {
	if _, found := l.nodeEnvironmentAccess(node); !found {
		return nil, false, nil
	}
	if err := l.nodeEnvironmentReadRefusal(node); err != nil {
		return nil, true, err
	}
	var name ir.Expression
	if node.Kind == ast.KindPropertyAccessExpression {
		name = ir.StringConstant{Index: l.constant(node.Name().Text())}
	} else {
		var err error
		name, err = l.expression(node.AsElementAccessExpression().ArgumentExpression)
		if err != nil {
			return nil, true, err
		}
		if name.Type() != ir.String {
			return nil, true, l.notYet(node, "node:process.Process.env read with a non-string key")
		}
	}
	return ir.NodeFSFile{Operation: "host_env", Arguments: []ir.Expression{name}, Of: ir.String}, true, nil
}

// The refusal visitor sees both process.env and its env name. Exempt them only
// as the receiver of an admitted outer read; the host object itself stays opaque.
func (l *lowering) nodeEnvironmentUseRefusal(node *ast.Node) (bool, error) {
	// Property-name tokens otherwise resolve to Node's generic Dict index signature.
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		if _, found := l.nodeEnvironmentAccess(node.Parent); found {
			return true, l.nodeEnvironmentReadRefusal(node.Parent)
		}
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindInKeyword && l.nodeLibraryMember(binary.Right) == "node:process.Process.env" {
			return true, l.notYet(binary.Right, "node:process.Process.env used as a value")
		}
	}
	if node.Kind == ast.KindDeleteExpression {
		target := ast.SkipParentheses(node.AsDeleteExpression().Expression)
		if _, found := l.nodeEnvironmentAccess(target); found || l.nodeLibraryMember(target) == "node:process.Process.env" {
			return true, l.notYet(node, "node:process.Process.env deleted")
		}
	}
	if _, found := l.nodeEnvironmentAccess(node); found {
		return true, l.nodeEnvironmentReadRefusal(node)
	}
	if l.nodeLibraryMember(node) != "node:process.Process.env" {
		return false, nil
	}
	access := node
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		access = node.Parent
	}
	outer := access
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil {
		if receiver, found := l.nodeEnvironmentAccess(outer.Parent); found && ast.SkipParentheses(receiver) == access {
			return true, l.nodeEnvironmentReadRefusal(outer.Parent)
		}
		if outer.Parent.Kind == ast.KindDeleteExpression {
			return true, l.notYet(access, "node:process.Process.env deleted")
		}
	}
	if ast.IsAssignmentTarget(outer) {
		return true, l.notYet(access, "node:process.Process.env written")
	}
	return true, l.notYet(access, "node:process.Process.env used as a value")
}
