package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// __filename, __dirname and require are what Node gives a CommonJS module. An Adamic program is an
// ES module, where Node has none of them, so they're never bound. Each is NotYet under a reason of
// its own, so the meter names who owns it (#redxes2); none is registered.
const (
	nodeFilenameMember = "node:module.__filename"
	nodeDirnameMember  = "node:module.__dirname"
	nodeRequireMember  = "node:module.require"
)

// nodeCommonJSModuleValues is the one reason for __filename and __dirname, so the meter counts them together.
const nodeCommonJSModuleValues = "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the path in, such as one built from process.cwd())"

// nodeCommonJSLoading is what each named require's reason says after naming it.
const nodeCommonJSLoading = ": CommonJS loading with no ES-module answer yet"

// nodeCommonJSUseRefusal names __filename and __dirname, and the requires the meter tracks: of
// crypto, of source-map-support, and of a computed module path. It refuses the nodes the generic
// NotYet would, at the same positions, so a site counts as it did: a call's node and its callee's
// carry one reason at one place. Any other require gives found false and keeps node:module.require.
func (l *lowering) nodeCommonJSUseRefusal(node *ast.Node) (bool, error) {
	switch l.nodeLibraryMember(node) {
	case nodeFilenameMember, nodeDirnameMember:
		return true, l.notYet(node, nodeCommonJSModuleValues)
	case nodeRequireMember:
	default:
		return false, nil
	}
	call := node
	if node.Kind != ast.KindCallExpression {
		call = nodeGlobalsCall(node)
	}
	if call == nil || len(call.AsCallExpression().Arguments.Nodes) == 0 {
		return false, nil
	}
	argument := ast.SkipParentheses(call.AsCallExpression().Arguments.Nodes[0])
	if argument.Kind != ast.KindStringLiteral && argument.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return true, l.notYet(node, nodeRequireMember+" of a computed module path"+nodeCommonJSLoading+" (fix: import each module it can load with an import declaration)")
	}
	switch argument.Text() {
	case "crypto", "source-map-support":
		return true, l.notYet(node, nodeRequireMember+`("`+argument.Text()+`")`+nodeCommonJSLoading+" (fix: import it with an import declaration)")
	}
	return false, nil
}
