package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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

// nodeCommonJSModuleValues is the one reason for __filename and __dirname, so the meter counts them together.
const nodeCommonJSModuleValues = "node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started)"

// nodeCommonJSLoading is what each named require's reason says after naming it.
const nodeCommonJSLoading = ": CommonJS loading with no ES-module answer yet"

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

// nodeCommonJSUseRefusal names __filename and __dirname, and the requires the meter tracks: of
// crypto, and of a computed module path called with a spread. It refuses the nodes the generic
// NotYet would, at the same positions, so a site counts as it did: a call's node and its callee's
// carry one reason at one place. A require that lowers (nodeCommonJSRequireThrows) is found with no
// refusal. Any other require gives found false and keeps node:module.require.
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
	if nodeCommonJSRequireThrows(call) {
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
	for {
		switch node.Kind {
		case ast.KindCallExpression:
			if l.nodeLibraryMember(node) == nodeRequireMember {
				if !nodeCommonJSRequireThrows(node) {
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
// keeps from being read. A value typed with require's own any has no type to hold, so its placeholder
// is held as the slot it's written into holds it (sys.ts:1619's `module: require(modulePath)`), and
// with no such slot it's NotYet as before.
func (l *lowering) nodeCommonJSRequireValue(node *ast.Node) (ir.Expression, bool, error) {
	call := l.nodeCommonJSRequireRoot(node)
	if call == nil {
		return nil, false, nil
	}
	effects := ir.Effects{Result: ir.Undefined{}}
	if outer := nodeGlobalsOuter(node); outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {
		var of ir.Type
		if slot := l.slotType(node); slot != nil && l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsAny != 0 {
			if held, known := l.representation(slot); known && held != ir.Weak {
				of = held
			}
		}
		if of == 0 {
			typed, err := l.typeOf(node)
			if err != nil {
				return nil, true, err
			}
			of = typed
		}
		effects.Result = placeholderZero(of)
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
