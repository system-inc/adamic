package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

func (lowering *lowering) processPath(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		return lowering.processPath(node.AsAsExpression().Expression)
	}
	if ast.IsIdentifier(node) {
		symbol := lowering.symbol(node)
		if symbol != nil && symbol.Name == "process" && len(symbol.Declarations) > 0 && (load.IsPrelude(
			ast.GetSourceFileOfNode(symbol.Declarations[0]),
		) || load.IsNodeLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))) {
			return "process"
		}
		if path := lowering.nodeProcessPath(node); path == "process" || strings.HasPrefix(path, "process.") {
			return path
		}
		return ""
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken == nil {
			if path := lowering.processPath(access.Expression); path != "" {
				return path + "." + node.Name().Text()
			}
		}
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if access.QuestionDotToken == nil && access.ArgumentExpression.Kind == ast.KindStringLiteral {
			if path := lowering.processPath(access.Expression); path != "" {
				return path + "." + access.ArgumentExpression.Text()
			}
		}
	}
	return ""
}

func (lowering *lowering) isProcessExit(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindCallExpression &&
		lowering.processPath(node.AsCallExpression().Expression) == "process.exit"
}

// processValue handles only complete supported operations. Intermediate objects cannot escape
// into ordinary object storage, whose slots could not model this live external state.
func (lowering *lowering) processValue(node *ast.Node) (ir.Expression, bool, error) {
	node = ast.SkipParentheses(node)
	if value, known, nodeStdoutWriteError := lowering.nodeStdoutWrite(node); known {
		return value, known, nodeStdoutWriteError
	}
	if value, known, nodeStdoutIsTtyReadError := lowering.nodeStdoutIsTtyRead(node); known {
		return value, known, nodeStdoutIsTtyReadError
	}
	if value, known, nodePlatformReadError := lowering.nodePlatformRead(node); known {
		return value, known, nodePlatformReadError
	}
	if value, known, nodeProcessValueError := lowering.nodeProcessValue(node); known {
		return value, known, nodeProcessValueError
	}
	if value, known, nodeEnvironmentReadError := lowering.nodeEnvironmentRead(node); known {
		return value, known, nodeEnvironmentReadError
	}
	if lowering.isProcessExit(node) {
		arguments := node.AsCallExpression().Arguments.Nodes
		if len(arguments) > 1 {
			return nil, true, lowering.notYet(node, "process.exit with more than one argument")
		}
		code := ir.Expression(ir.ProcessCall{Operation: "exitCode", Of: ir.MaybeNumber})
		if len(arguments) == 1 {
			value, expressionError := lowering.expression(arguments[0])
			if expressionError != nil {
				return nil, true, expressionError
			}
			code = fit(value, ir.MaybeNumber)
			if code.Type() != ir.MaybeNumber {
				return nil, true, lowering.notYet(arguments[0], "process.exit with a nonnumeric code")
			}
		}
		return ir.ProcessCall{Operation: "exit", Arguments: []ir.Expression{code}, Of: ir.Number}, true, nil
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if lowering.processPath(binary.Left) == "process.exitCode" {
			if binary.OperatorToken.Kind != ast.KindEqualsToken {
				return nil, true, lowering.notYet(node, "a compound process.exitCode assignment")
			}
			value, expressionError := lowering.expression(binary.Right)
			if expressionError != nil {
				return nil, true, expressionError
			}
			code := fit(value, ir.MaybeNumber)
			if code.Type() != ir.MaybeNumber {
				return nil, true, lowering.notYet(binary.Right, "process.exitCode with a nonnumeric code")
			}
			call := ir.ProcessCall{Operation: "setExitCode", Arguments: []ir.Expression{code}, Of: ir.MaybeNumber}
			if value.Type() == ir.Object {
				return ir.Void{Value: call, Of: ir.Object}, true, nil
			}
			return fit(call, value.Type()), true, nil
		}
	}
	path := lowering.processPath(node)
	switch path {
	case "process.exitCode":
		// Writes to this live status are restricted to the numeric ABI above.
		// The wider Node declaration also admits strings and null, which this unit refuses.
		call := ir.Expression(ir.ProcessCall{Operation: "exitCode", Of: ir.MaybeNumber})
		if to, known := lowering.representation(lowering.checker.GetTypeAtLocation(node)); known && to == ir.Number {
			call = fit(call, to)
		}
		return call, true, nil
	case "":
		return nil, false, nil
	default:
		return nil, true, lowering.notYet(node, path+" as a value (use its supported process operation directly)")
	}
}
