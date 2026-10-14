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

// processValue handles only complete supported operations. Intermediate objects cannot escape
// into ordinary object storage, whose slots could not model this live external state.
func (l *lowering) processValue(node *ast.Node) (ir.Expression, bool, error) {
	if value, known, err := l.nodeProcessValue(node); known {
		return value, known, err
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
