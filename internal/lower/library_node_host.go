package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// This host slice recognizes library declarations, including import aliases.
// It borrows only strings and does not expose a mutable process object.
func (l *lowering) nodeHost(node *ast.Node) (ir.Expression, bool, error) {
	if value, found, err := l.nodeRealpathCall(node); found {
		return value, true, err
	}
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
	RegisterNodeLibraryMembers("node:path.join", "node:os.tmpdir", "node:process.Process.cwd", "node:process.Process.chdir", "node:process.cwd", "node:process.chdir", "node:process.process", "node:globals.process", "node:process.Process.platform")
}

// nodePlatformRead lowers process.platform, recognized by its library member, never its spelling,
// to a host read (#t90fcv5): the binary answers the platform it runs on, as Node does, and the answer
// is never folded into a constant at compile time. Only a plain read is admitted.
func (l *lowering) nodePlatformRead(node *ast.Node) (ir.Expression, bool, error) {
	if l.nodeLibraryMember(node) != "node:process.Process.platform" || (node.Kind != ast.KindPropertyAccessExpression && node.Kind != ast.KindElementAccessExpression) {
		return nil, false, nil
	}
	if err := l.nodePlatformReadRefusal(node); err != nil {
		return nil, true, err
	}
	return ir.NodeFSFile{Operation: "host_platform", Of: ir.String}, true, nil
}

// nodePlatformReadRefusal names every use of process.platform but a plain read: a write, a delete,
// an optional or element read, and the member anywhere else, such as a destructured binding.
func (l *lowering) nodePlatformReadRefusal(node *ast.Node) error {
	const member = "node:process.Process.platform"
	if node.Kind != ast.KindPropertyAccessExpression {
		return l.notYet(node, member+" used as a value")
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	switch {
	case outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression:
		return l.notYet(node, member+" deleted")
	case ast.IsAssignmentTarget(outer):
		return l.notYet(node, member+" written")
	case node.Flags&ast.NodeFlagsOptionalChain != 0:
		return l.notYet(node, member+" read through an optional chain")
	}
	return nil
}

// nodePlatformUseRefusal is the refusal visitor's view of process.platform: its property name is
// part of the admitted read, the read itself is judged by nodePlatformReadRefusal, and the member
// anywhere else is NotYet.
func (l *lowering) nodePlatformUseRefusal(node *ast.Node) (bool, error) {
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node && l.nodeLibraryMember(node.Parent) == "node:process.Process.platform" {
		return true, l.nodePlatformReadRefusal(node.Parent)
	}
	if l.nodeLibraryMember(node) != "node:process.Process.platform" {
		return false, nil
	}
	return true, l.nodePlatformReadRefusal(node)
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

const nodeRealpathWalkNotYet = "node:fs.realpathSync's JS walk isn't bound yet; use realpathSync.native"

// The namespace's native declaration is identified by the Node library symbol;
// the receiver disambiguates it from any other namespace's native member.
func (l *lowering) nodeRealpathNative(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && l.nodeLibraryMember(node) == "node:fs.native" && l.nodeLibraryMember(node.AsPropertyAccessExpression().Expression) == "node:fs.realpathSync"
}

func (l *lowering) nodeRealpathUseRefusal(node *ast.Node) (bool, error) {
	if node.Kind == ast.KindCallExpression {
		node = ast.SkipParentheses(node.AsCallExpression().Expression)
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		node = node.Parent
	}
	if l.nodeRealpathNative(node) {
		if node.Flags&ast.NodeFlagsOptionalChain != 0 || ast.IsAssignmentTarget(node) {
			return true, l.notYet(node, "node:fs.realpathSync.native outside a read or call")
		}
		return true, l.nodeRealpathReferenceRefusal(node)
	}
	if l.nodeLibraryMember(node) == "node:fs.realpathSync" {
		if node.Parent != nil && l.nodeRealpathNative(node.Parent) {
			return true, nil
		}
		return true, l.notYet(node, nodeRealpathWalkNotYet)
	}
	return false, nil
}

func (l *lowering) nodeRealpathCall(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindCallExpression || !l.nodeRealpathNative(node.AsCallExpression().Expression) {
		return nil, false, nil
	}
	call := node.AsCallExpression()
	if len(call.Arguments.Nodes) != 1 {
		return nil, true, l.notYet(node, "node:fs.realpathSync.native with encoding, options, or callback arguments; need exactly one string path")
	}
	argument := call.Arguments.Nodes[0]
	if argument.Kind == ast.KindSpreadElement {
		return nil, true, l.notYet(argument, "node:fs.realpathSync.native with spread arguments")
	}
	if l.checker.GetTypeAtLocation(argument).Flags()&checker.TypeFlagsStringLike == 0 {
		return nil, true, l.notYet(argument, "node:fs.realpathSync.native with a non-string path (Buffer and URL paths aren't bound yet)")
	}
	value, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	if value.Type() != ir.String {
		return nil, true, l.notYet(argument, "node:fs.realpathSync.native with a non-string path (Buffer and URL paths aren't bound yet)")
	}
	return ir.NodeFSFile{Operation: "host_realpath", Arguments: []ir.Expression{value}, Of: ir.String}, true, nil
}

// Use the ordinary global forwarder initialization: one owned closure per program,
// with a real throwing IR body rather than an opaque intrinsic alias token.
func (l *lowering) nodeRealpathValue(node *ast.Node) (ir.Expression, bool, error) {
	if !l.nodeRealpathNative(node) {
		return nil, false, nil
	}
	if expected := l.checker.GetContextualType(node, checker.ContextFlagsNone); expected != nil {
		signatures := l.checker.GetSignaturesOfType(expected, checker.SignatureKindCall)
		valid := len(signatures) == 1
		if valid {
			signature := signatures[0]
			parameters := signature.Parameters()
			valid = signature.ThisParameter() == nil && len(parameters) == 1 && signature.MinArgumentCount() == 1 && signature.Flags()&checker.SignatureFlagsHasRestParameter == 0 && len(signature.TypeParameters()) == 0
			if valid {
				valid = l.concrete(l.checker.GetTypeOfSymbol(parameters[0])).Flags() == checker.TypeFlagsString && l.concrete(l.checker.GetReturnTypeOfSignature(signature)).Flags() == checker.TypeFlagsString
			}
		}
		if !valid {
			return nil, true, l.notYet(node, "node:fs.realpathSync.native value needs exactly (path: string) => string, not "+l.checker.TypeToString(expected))
		}
	}
	for index, function := range l.result.Functions {
		if function.Source.Role == "host_realpath" {
			return ir.Read{Local: l.forwarders[index], Of: ir.Closure}, true, nil
		}
	}
	index, parameter := len(l.result.Functions), len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "path", Type: ir.String, Function: index})
	function := ir.Function{Name: "host_realpath_value", SourceLength: 1, Closure: true, Parameters: []int{parameter}, Returns: ir.String, Body: []ir.Statement{ir.Return{Value: ir.NodeFSFile{Operation: "host_realpath", Arguments: []ir.Expression{ir.Read{Local: parameter, Of: ir.String}}, Of: ir.String}}}}
	function.Source.Role = "host_realpath"
	l.result.Functions = append(l.result.Functions, function)
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "host_realpath_value", Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	l.forwarders[index] = held
	return ir.Read{Local: held, Of: ir.Closure}, true, nil
}
