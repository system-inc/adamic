package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// __filename, __dirname and require are what Node gives a CommonJS module. An Adamic program is an
// ES module, where Node has none of them, so they're never bound. Each is NotYet under a reason of
// its own, so the meter names who owns it (#redxes2); none is registered. Two requires tsc makes
// inside a try lower as the throw its catch sees instead (Adamic's ruling, Nov 3): of
// source-map-support, and of a computed module path.
const (
	nodeFilenameMember = "node:module.__filename"
	nodeDirnameMember  = "node:module.__dirname"
	nodeRequireMember  = "node:module.require"
)

// nodeCommonJsModuleValues is the one reason for __filename and __dirname, so the meter counts them together.
const nodeCommonJsModuleValues = "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)"

// nodeCommonJsLoading is what each named require's reason says after naming it.
const nodeCommonJsLoading = ": CommonJS loading with no ES-module answer yet"

// nodeSourceMapSupport is the one module a literal require lowers for. tsc 6.0.3 ships no
// dependencies, so in its installed layout Node answers its require with MODULE_NOT_FOUND, and
// that's the answer native gives (installed-layout in internal/oracle's ruled divergences).
const nodeSourceMapSupport = "source-map-support"

// nodeJavaScriptModuleOpening and nodeJavaScriptModuleClosing are the message a computed require
// throws with, around the path it was given, in Diagnostics' words (#redxes2): which load failed, the
// limit, and the one fix. Stock Node would load the module; native can't run JavaScript
// (no-javascript-plugins).
const (
	nodeJavaScriptModuleOpening = "Cannot load JavaScript module '"
	nodeJavaScriptModuleClosing = "': a native Adamic program can't run JavaScript; run it on Node to load modules at run time"
)

// nodeCommonJsUseRefusal names __filename and __dirname, and the requires the meter tracks: of
// crypto, and of a computed module path called with a spread. It refuses the nodes the generic
// NotYet would, at the same positions, so a site counts as it did: a call's node and its callee's
// carry one reason at one place. A require that lowers (nodeCommonJSRequireThrows) is found with no
// refusal. Any other require gives found false and keeps node:module.require.
func (lowering *lowering) nodeCommonJsUseRefusal(node *ast.Node) (bool, error) {
	switch lowering.nodeLibraryMember(node) {
	case nodeFilenameMember, nodeDirnameMember:
		return true, lowering.notYet(node, nodeCommonJsModuleValues)
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
	if nodeCommonJsRequireThrows(call) || nodeInspectorRequireThrows(call) {
		return true, nil
	}
	argument := nodeCommonJsRequireArgument(call)
	if argument.Kind != ast.KindStringLiteral && argument.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return true, lowering.notYet(node, nodeRequireMember+
			" of a computed module path called with a spread or a second argument"+
			nodeCommonJsLoading+
			" (fix: pass the module path as its one argument)")
	}
	if argument.Text() == "crypto" {
		return true, lowering.notYet(node, nodeRequireMember+
			`("`+
			argument.Text()+
			`")`+
			nodeCommonJsLoading+
			" (fix: import it with an import declaration)")
	}
	return false, nil
}

// nodeCommonJsRequireArgument is a require call's first argument as it's judged: past the
// parentheses, assertions and non-nulls that leave its value as it is, so a literal wrapped in one
// (`require("source-map-support" as string)`) is the literal it wraps, never a computed path.
func nodeCommonJsRequireArgument(call *ast.Node) *ast.Node {
	return ast.SkipOuterExpressions(call.AsCallExpression().Arguments.Nodes[0], ast.OEKAll)
}

// nodeCommonJsRequireThrows says whether a require call lowers as the throw Adamic ruled for it: a
// literal require of source-map-support, or a require of a path computed as the program runs. Every
// other literal require stays refused, so this is never a way to load code.
func nodeCommonJsRequireThrows(call *ast.Node) bool {
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return false
	}
	argument := nodeCommonJsRequireArgument(call)
	switch argument.Kind {
	case ast.KindSpreadElement:
		return false
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return argument.Text() == nodeSourceMapSupport
	}
	return true
}

// nodeCommonJsRequireRoot is the require call a value's evaluation starts with, when it's one that
// lowers: node itself, or the receiver its member accesses and calls are made on, through
// parentheses, assertions and a non-null. The require throws before anything after it in the chain
// is evaluated, so the whole chain lowers as its throw. Nil for any other node.
func (lowering *lowering) nodeCommonJsRequireRoot(node *ast.Node) *ast.Node {
	original := node
	for {
		switch node.Kind {
		case ast.KindCallExpression:
			if lowering.nodeLibraryMember(node) == nodeRequireMember {
				if !nodeCommonJsRequireThrows(node) && !(node == original && nodeInspectorRequireThrows(node)) {
					return nil
				}
				return node
			}
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		default:
			return nil
		}
		node = node.Expression()
	}
}

// nodeCommonJsRequireValue lowers a value whose evaluation starts with a require that lowers. A
// require of source-map-support is the host's MODULE_NOT_FOUND, with Node's name, code and the
// first line of its message (host_require_missing). A computed require evaluates its argument, then
// throws an Error naming the path. Either way the value is never had: as a statement it's
// undefined, as gc's call is, and anywhere else a placeholder of the value's type that the throw
// keeps from being read.
func (lowering *lowering) nodeCommonJsRequireValue(node *ast.Node) (ir.Expression, bool, error) {
	call := lowering.nodeCommonJsRequireRoot(node)
	if call == nil {
		return nil, false, nil
	}
	effects := ir.Effects{Result: ir.Undefined{}}
	if outer := nodeGlobalsOuter(node); outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {
		of, typeOfError := lowering.typeOf(node)
		if typeOfError != nil {
			return nil, true, typeOfError
		}
		effects.Result = placeholderZero(of)
	}
	if nodeInspectorRequireThrows(call) {
		effects.Body = append(
			effects.Body,
			ir.Evaluate{Value: ir.NodeFSFile{Operation: "host_inspector_unavailable", Of: ir.Number}},
		)
		return effects, true, nil
	}
	if argument := nodeCommonJsRequireArgument(
		call,
	); argument.Kind == ast.KindStringLiteral || argument.Kind == ast.KindNoSubstitutionTemplateLiteral {
		missing := ir.NodeFSFile{
			Operation: "host_require_missing",
			Arguments: []ir.Expression{ir.StringConstant{Index: lowering.constant(argument.Text())}},
			Of:        ir.Number,
		}
		effects.Body = append(effects.Body, ir.Evaluate{Value: missing})
		return effects, true, nil
	}
	// The path is evaluated as written, wrappers and all: a non-null assertion may carry a check.
	argument := call.AsCallExpression().Arguments.Nodes[0]
	path, expressionError := lowering.expression(argument)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if path.Type() != ir.String {
		return nil, true, lowering.notYet(argument, nodeRequireMember+
			" of a computed module path that isn't a string"+
			nodeCommonJsLoading+
			" (fix: pass the module path as a string)")
	}
	index := len(lowering.result.Functions)
	parameter := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "modulePath", Type: ir.String, Function: index},
	)
	failure := ir.MakeError{Message: ir.Concat{Parts: []ir.Expression{
		ir.StringConstant{Index: lowering.constant(nodeJavaScriptModuleOpening)},
		ir.Read{Local: parameter, Of: ir.String},
		ir.StringConstant{Index: lowering.constant(nodeJavaScriptModuleClosing)},
	}}}
	lowering.result.Functions = append(lowering.result.Functions, ir.Function{
		Name:       "require_javascript",
		Parameters: []int{parameter},
		Body:       []ir.Statement{ir.Throw{Value: failure}},
	})
	effects.Body = append(effects.Body, ir.Evaluate{Value: ir.Call{Function: index, Arguments: []ir.Expression{path}}})
	return effects, true, nil
}

// Inspector has no runtime value on the configured no-inspector host. Keep this
// literal-only rule separate from the form-scoped Modelled require audit. Member
// chains are not admitted: their independent inspector bindings belong to head 2.
func nodeInspectorRequireThrows(call *ast.Node) bool {
	if len(call.AsCallExpression().Arguments.Nodes) != 1 {
		return false
	}
	argument := nodeCommonJsRequireArgument(call)
	if argument.Kind != ast.KindStringLiteral && argument.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return false
	}
	switch argument.Text() {
	case "inspector", "node:inspector", "inspector/promises":
		return true
	}
	return false
}
