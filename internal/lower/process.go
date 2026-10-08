package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

func (l *lowering) processPath(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		return l.processPath(node.AsAsExpression().Expression)
	}
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil && symbol.Name == "process" && len(symbol.Declarations) > 0 && (load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) || load.IsNodeLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))) {
			return "process"
		}
		if path := l.nodeProcessPath(node); path == "process" || strings.HasPrefix(path, "process.") {
			return path
		}
		return ""
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken == nil {
			if path := l.processPath(access.Expression); path != "" {
				return path + "." + node.Name().Text()
			}
		}
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if access.QuestionDotToken == nil && access.ArgumentExpression.Kind == ast.KindStringLiteral {
			if path := l.processPath(access.Expression); path != "" {
				return path + "." + access.ArgumentExpression.Text()
			}
		}
	}
	return ""
}

func (l *lowering) isProcessExit(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindCallExpression && l.processPath(node.AsCallExpression().Expression) == "process.exit"
}

// processValue handles only complete supported operations. Intermediate objects cannot escape
// into ordinary object storage, whose slots could not model this live external state.
func (l *lowering) processValue(node *ast.Node) (ir.Expression, bool, error) {
	if value, known, err := l.nodeProcessValue(node); known {
		return value, known, err
	}
	if l.isProcessExit(node) {
		arguments := node.AsCallExpression().Arguments.Nodes
		if len(arguments) > 1 {
			return nil, true, l.notYet(node, "process.exit with more than one argument")
		}
		code := ir.Expression(ir.ProcessCall{Operation: "exitCode", Of: ir.MaybeNumber})
		if len(arguments) == 1 {
			value, err := l.expression(arguments[0])
			if err != nil {
				return nil, true, err
			}
			if value.Type() != ir.Number && value.Type() != ir.MaybeNumber {
				if _, absent := value.(ir.Undefined); !absent {
					return nil, true, l.notYet(arguments[0], "process.exit with a non-numeric code")
				}
			}
			code = fit(value, ir.MaybeNumber)
		}
		return ir.ProcessCall{Operation: "exit", Arguments: []ir.Expression{code}, Of: ir.Number}, true, nil
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if l.processPath(binary.Left) == "process.exitCode" {
			if binary.OperatorToken.Kind != ast.KindEqualsToken {
				return nil, true, l.notYet(node, "a compound process.exitCode assignment")
			}
			value, err := l.expression(binary.Right)
			if err != nil {
				return nil, true, err
			}
			if value.Type() != ir.Number && value.Type() != ir.MaybeNumber {
				if _, absent := value.(ir.Undefined); !absent {
					return nil, true, l.notYet(binary.Right, "process.exitCode assignment with a non-numeric code")
				}
			}
			call := ir.ProcessCall{Operation: "setExitCode", Arguments: []ir.Expression{fit(value, ir.MaybeNumber)}, Of: ir.MaybeNumber}
			return fit(call, value.Type()), true, nil
		}
	}
	var key *ast.Node
	if node.Kind == ast.KindElementAccessExpression && l.processPath(node.AsElementAccessExpression().Expression) == "process.env" {
		key = node.AsElementAccessExpression().ArgumentExpression
	}
	path := l.processPath(node)
	if key != nil || (node.Kind == ast.KindPropertyAccessExpression && l.processPath(node.AsPropertyAccessExpression().Expression) == "process.env") {
		var name ir.Expression
		if key != nil {
			value, err := l.expression(key)
			if err != nil {
				return nil, true, err
			}
			name = l.spelled(key, value)
			if name.Type() == ir.Number {
				name = ir.NumberToString{Value: name}
			}
			if name.Type() != ir.String {
				return nil, true, l.notYet(key, "an environment key other than a string or number")
			}
		} else {
			name = ir.StringConstant{Index: l.constant(node.Name().Text())}
		}
		return ir.ProcessCall{Operation: "env", Arguments: []ir.Expression{name}, Of: ir.String}, true, nil
	}
	var call ir.ProcessCall
	switch path {
	case "process.exitCode":
		call = ir.ProcessCall{Operation: "exitCode", Of: ir.MaybeNumber}
	case "process.stdout.isTTY":
		call = ir.ProcessCall{Operation: "stdoutTTY", Of: ir.MaybeBoolean}
	case "process.stderr.isTTY":
		call = ir.ProcessCall{Operation: "stderrTTY", Of: ir.MaybeBoolean}
	case "":
		return nil, false, nil
	default:
		return nil, true, l.notYet(node, path+" as a value (use its supported process operation directly)")
	}
	// A getter narrowed to undefined still reads the live external state, as every getter does.
	if to, known := l.representation(l.checker.GetTypeAtLocation(node)); known {
		return fit(call, to), true, nil
	}
	return call, true, nil
}
