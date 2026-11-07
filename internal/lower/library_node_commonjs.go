package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// __filename, __dirname and require are what Node gives a CommonJS module. An Adamic program is an
// ES module unless its build declares one bundled image of CommonJS modules, and an ES module has
// none of them. Each is NotYet under a reason of its own, so the meter names who owns it (#redxes2);
// none is registered. Two requires tsc makes inside a try lower as the throw its catch sees instead
// (Adamic's ruling, Nov 3): of source-map-support, and of a computed module path. __filename and
// __dirname read in a .ts file of a build that declares one bundled image lower as the running
// image's path and its directory (nodeCommonJSModuleValuesBound).
const (
	nodeFilenameMember = "node:module.__filename"
	nodeDirnameMember  = "node:module.__dirname"
	nodeRequireMember  = "node:module.require"
)

// nodeCommonJSModuleValues is the one reason for __filename and __dirname, so the meter counts them together.
const nodeCommonJSModuleValues = "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)"

// nodeCommonJSLoading is what each named require's reason says after naming it.
const nodeCommonJSLoading = ": CommonJS loading with no ES-module answer yet"

// nodeSourceMapSupport is the one module a literal require lowers for. tsc 6.0.3 ships no
// dependencies, so in its installed layout Node answers its require with MODULE_NOT_FOUND, and
// that's the answer native gives (installed-layout in internal/oracle's ruled divergences).
const nodeSourceMapSupport = "source-map-support"

// nodeCommonJSModuleValueShape is what a reason for __filename or __dirname used other than by a read
// of its own name says after naming it, in a file where they're bound.
const nodeCommonJSModuleValueShape = ": CommonJS module values are bound only for a read of the name itself (fix: read __filename or __dirname by name and keep the path in a variable of the program's own)"

// nodeJavaScriptModuleOpening and nodeJavaScriptModuleClosing are the message a computed require
// throws with, around the path it was given, in Diagnostics' words (#redxes2): which load failed, the
// limit, and the one fix. Stock Node would load the module; native can't run JavaScript
// (no-javascript-plugins).
const (
	nodeJavaScriptModuleOpening = "Cannot load JavaScript module '"
	nodeJavaScriptModuleClosing = "': a native Adamic program can't run JavaScript; run it on Node to load modules at run time"
)

// nodeCommonJSUseRefusal names __filename and __dirname, and the requires the meter tracks: of
// crypto, and of a computed module path called with a spread. It refuses the nodes the generic
// NotYet would, at the same positions, so a site counts as it did: a call's node and its callee's
// carry one reason at one place. A require that lowers (nodeCommonJSRequireThrows), and a read of
// __filename or __dirname where they're bound, is found with no refusal. Any other require gives
// found false and keeps node:module.require.
func (l *lowering) nodeCommonJSUseRefusal(node *ast.Node) (bool, error) {
	switch l.nodeLibraryMember(node) {
	case nodeFilenameMember, nodeDirnameMember:
		if !l.nodeCommonJSModuleValuesBound(node) {
			return true, l.notYet(node, nodeCommonJSModuleValues)
		}
		return true, l.nodeCommonJSModuleValueRefusal(node)
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
	if nodeCommonJSRequireThrows(call) || nodeInspectorRequireThrows(call) {
		return true, nil
	}
	argument := nodeCommonJSRequireArgument(call)
	if argument.Kind != ast.KindStringLiteral && argument.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return true, l.notYet(node, nodeRequireMember+" of a computed module path called with a spread or a second argument"+nodeCommonJSLoading+" (fix: pass the module path as its one argument)")
	}
	if argument.Text() == "crypto" {
		return true, l.notYet(node, nodeRequireMember+`("`+argument.Text()+`")`+nodeCommonJSLoading+" (fix: import it with an import declaration)")
	}
	return false, nil
}

// nodeCommonJSModuleValuesBound says whether __filename and __dirname are bound where node is:
// only in a build that declares one bundled image (ModuleFormatCommonJSBundle), a build input
// lowered.Options keys, and only in a .ts file, since a .a file is an ES module whatever the build
// declares (Adamic's ruling, Nov 6, #redxes2). Bound, every module of the program gets the image's
// two values, its real path and its directory, as every module of tsc's _tsc.js gets the bundle's
// (Adamic's ruling on #0ta3dk9), never a path of its own. Everywhere else they keep
// nodeCommonJSModuleValues' NotYet. It's the one place the format and the file are judged, so the
// refusal and the lowering can't disagree.
func (l *lowering) nodeCommonJSModuleValuesBound(node *ast.Node) bool {
	return l.moduleFormat == ModuleFormatCommonJSBundle && strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts")
}

// nodeCommonJSModuleValueRefusal judges a use of __filename or __dirname where they're bound. Only a
// read of the name itself lowers. A CommonJS module's two are its wrapper's locals, never the global
// object's, so a read through globalThis is undefined on Node and stays NotYet, as does a write,
// which reaches one module or every module the bundle wraps, as the bundle was made. The name of a
// property access is judged as its access.
func (l *lowering) nodeCommonJSModuleValueRefusal(node *ast.Node) error {
	name := l.nodeLibraryMember(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		return l.notYet(node.Parent, name+" through a receiver"+nodeCommonJSModuleValueShape)
	}
	if !ast.IsIdentifier(node) {
		return l.notYet(node, name+" through a receiver"+nodeCommonJSModuleValueShape)
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if ast.IsAssignmentTarget(outer) || outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression {
		return l.notYet(node, name+" written"+nodeCommonJSModuleValueShape)
	}
	return nil
}

// nodeCommonJSModuleValue lowers a read of __filename to the running image's real path and of
// __dirname to its directory, where they're bound. The refusal visitor has judged every other use,
// and it's judged again here, so a use the visitor never saw can't lower.
func (l *lowering) nodeCommonJSModuleValue(node *ast.Node) (ir.Expression, bool, error) {
	if !ast.IsIdentifier(node) {
		return nil, false, nil
	}
	operation := ""
	switch l.nodeLibraryMember(node) {
	case nodeFilenameMember:
		operation = "filename"
	case nodeDirnameMember:
		operation = "dirname"
	default:
		return nil, false, nil
	}
	if !l.nodeCommonJSModuleValuesBound(node) {
		return nil, true, l.notYet(node, nodeCommonJSModuleValues)
	}
	if err := l.nodeCommonJSModuleValueRefusal(node); err != nil {
		return nil, true, err
	}
	return ir.ProcessCall{Operation: operation, Of: ir.String}, true, nil
}

// nodeCommonJSRequireArgument is a require call's first argument as it's judged: past the
// parentheses, assertions and non-nulls that leave its value as it is, so a literal wrapped in one
// (`require("source-map-support" as string)`) is the literal it wraps, never a computed path.
func nodeCommonJSRequireArgument(call *ast.Node) *ast.Node {
	return ast.SkipOuterExpressions(call.AsCallExpression().Arguments.Nodes[0], ast.OEKAll)
}

// nodeCommonJSRequireThrows says whether a require call lowers as the throw Adamic ruled for it: a
// literal require of source-map-support, or a require of a path computed as the program runs. Every
// other literal require stays refused, so this is never a way to load code.
func nodeCommonJSRequireThrows(call *ast.Node) bool {
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return false
	}
	argument := nodeCommonJSRequireArgument(call)
	switch argument.Kind {
	case ast.KindSpreadElement:
		return false
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return argument.Text() == nodeSourceMapSupport
	}
	return true
}

// nodeCommonJSRequireRoot is the require call a value's evaluation starts with, when it's one that
// lowers: node itself, or the receiver its member accesses and calls are made on, through
// parentheses, assertions and a non-null. The require throws before anything after it in the chain
// is evaluated, so the whole chain lowers as its throw. Nil for any other node.
func (l *lowering) nodeCommonJSRequireRoot(node *ast.Node) *ast.Node {
	original := node
	for {
		switch node.Kind {
		case ast.KindCallExpression:
			if l.nodeLibraryMember(node) == nodeRequireMember {
				if !nodeCommonJSRequireThrows(node) && !(node == original && nodeInspectorRequireThrows(node)) {
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

// nodeCommonJSRequireValue lowers a value whose evaluation starts with a require that lowers. A
// require of source-map-support is the host's MODULE_NOT_FOUND, with Node's name, code and the
// first line of its message (host_require_missing). A computed require evaluates its argument, then
// throws an Error naming the path. Either way the value is never had: as a statement it's
// undefined, as gc's call is, and anywhere else a placeholder of the value's type that the throw
// keeps from being read.
func (l *lowering) nodeCommonJSRequireValue(node *ast.Node) (ir.Expression, bool, error) {
	call := l.nodeCommonJSRequireRoot(node)
	if call == nil {
		return nil, false, nil
	}
	effects := ir.Effects{Result: ir.Undefined{}}
	if outer := nodeGlobalsOuter(node); outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {
		of, err := l.typeOf(node)
		if err != nil {
			return nil, true, err
		}
		effects.Result = placeholderZero(of)
	}
	if nodeInspectorRequireThrows(call) {
		effects.Body = append(effects.Body, ir.Evaluate{Value: ir.NodeFSFile{Operation: "host_inspector_unavailable", Of: ir.Number}})
		return effects, true, nil
	}
	if argument := nodeCommonJSRequireArgument(call); argument.Kind == ast.KindStringLiteral || argument.Kind == ast.KindNoSubstitutionTemplateLiteral {
		missing := ir.NodeFSFile{Operation: "host_require_missing", Arguments: []ir.Expression{ir.StringConstant{Index: l.constant(argument.Text())}}, Of: ir.Number}
		effects.Body = append(effects.Body, ir.Evaluate{Value: missing})
		return effects, true, nil
	}
	// The path is evaluated as written, wrappers and all: a non-null assertion may carry a check.
	argument := call.AsCallExpression().Arguments.Nodes[0]
	path, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	if path.Type() != ir.String {
		return nil, true, l.notYet(argument, nodeRequireMember+" of a computed module path that isn't a string"+nodeCommonJSLoading+" (fix: pass the module path as a string)")
	}
	index := len(l.result.Functions)
	parameter := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "modulePath", Type: ir.String, Function: index})
	failure := ir.MakeError{Message: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant(nodeJavaScriptModuleOpening)}, ir.Read{Local: parameter, Of: ir.String}, ir.StringConstant{Index: l.constant(nodeJavaScriptModuleClosing)}}}}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "require_javascript", Parameters: []int{parameter}, Body: []ir.Statement{ir.Throw{Value: failure}}})
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
	argument := nodeCommonJSRequireArgument(call)
	if argument.Kind != ast.KindStringLiteral && argument.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return false
	}
	switch argument.Text() {
	case "inspector", "node:inspector", "inspector/promises":
		return true
	}
	return false
}
