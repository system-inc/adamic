package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Ambient declarations are recognized by their module and declaration, never by a user's spelling.
func (l *lowering) nodeProcessPath(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) == 0 {
			return ""
		}
		declaration := symbol.Declarations[0]
		if !load.IsPrelude(ast.GetSourceFileOfNode(declaration)) {
			return ""
		}
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindModuleDeclaration {
				switch parent.Name().Text() {
				case "node:process":
					return "process." + symbol.Name
				case "node:os":
					return "os." + symbol.Name
				case "node:perf_hooks":
					return symbol.Name
				}
			}
		}
		if symbol.Name == "performance" {
			return "performance"
		}
		return ""
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken == nil {
			if path := l.nodeProcessPath(access.Expression); path != "" {
				return path + "." + node.Name().Text()
			}
		}
	}
	return ""
}

func (l *lowering) nodeProcessValue(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind == ast.KindPrefixUnaryExpression {
		outer := node.AsPrefixUnaryExpression()
		inner := ast.SkipParentheses(outer.Operand)
		if outer.Operator == ast.KindExclamationToken && inner.Kind == ast.KindPrefixUnaryExpression {
			innerPrefix := inner.AsPrefixUnaryExpression()
			if innerPrefix.Operator == ast.KindExclamationToken && l.processPath(innerPrefix.Operand) == "process.nextTick" {
				return ir.ProcessCall{Operation: "nextTickFeature", Of: ir.Boolean}, true, nil
			}
		}
	}

	if node.Kind == ast.KindTypeOfExpression {
		operand := node.AsTypeOfExpression().Expression
		path := l.processPath(operand)
		if path == "" {
			path = l.nodeProcessPath(operand)
		}
		switch path {
		case "process.nextTick", "performance.now", "performance.mark", "performance.measure", "performance.clearMarks", "performance.clearMeasures", "process.cwd", "process.memoryUsage":
			return ir.StringConstant{Index: l.constant("function")}, true, nil
		}
	}
	path := l.processPath(node)
	if path == "" {
		path = l.nodeProcessPath(node)
	}
	call := ir.ProcessCall{}
	if node.Kind == ast.KindCallExpression {
		written := node.AsCallExpression()
		path = l.processPath(written.Expression)
		if path == "" {
			path = l.nodeProcessPath(written.Expression)
		}
		count, minimum := 0, 0
		switch path {
		case "process.cwd", "os.platform":
			call.Operation, call.Of = "cwd", ir.String
			if path == "os.platform" {
				call.Operation = "platform"
			}
		case "process.memoryUsage":
			call.Operation, call.Of = "memoryUsage", ir.Object
		case "process.stdout.write":
			if node.Parent == nil || node.Parent.Kind != ast.KindExpressionStatement {
				return nil, true, l.notYet(node, "observing the backpressure return of process.stdout.write")
			}
			call.Operation, call.Of, count, minimum = "stdoutWrite", ir.Object, 1, 1
		case "performance.now":
			call.Operation, call.Of = "now", ir.Number
		case "performance.mark":
			call.Operation, call.Of, count, minimum = "mark", ir.Object, 1, 1
		case "performance.measure":
			call.Operation, call.Of, count, minimum = "measure", ir.Object, 3, 1
		case "performance.clearMarks":
			call.Operation, call.Of, count = "clearMarks", ir.Object, 1
		case "performance.clearMeasures":
			call.Operation, call.Of, count = "clearMeasures", ir.Object, 1
		default:
			return nil, false, nil
		}
		if len(written.Arguments.Nodes) < minimum || len(written.Arguments.Nodes) > count {
			return nil, true, l.notYet(node, path+" with arguments outside the census contract")
		}
		for _, argument := range written.Arguments.Nodes {
			value, err := l.expression(argument)
			if err != nil {
				return nil, true, err
			}
			if value.Type() != ir.String && value.Type() != ir.Object {
				return nil, true, l.notYet(argument, path+" with a non-string argument")
			}
			if _, absent := value.(ir.Undefined); absent {
				value = ir.Undefined{Of: ir.String}
			}
			call.Arguments = append(call.Arguments, value)
		}
		for len(call.Arguments) < count {
			call.Arguments = append(call.Arguments, ir.Undefined{Of: ir.String})
		}
		return call, true, nil
	}
	switch path {
	case "process.platform":
		call.Operation, call.Of = "platform", ir.String
	case "process.pid":
		call.Operation, call.Of = "pid", ir.Number
	case "process.argv":
		call.Operation, call.Of = "argv", ir.Array
	case "process.execArgv":
		call.Operation, call.Of = "execArgv", ir.Array
	case "process.stdout._handle":
		call.Operation, call.Of = "handle", ir.Object
	case "process.stdout.columns":
		call.Operation, call.Of = "columns", ir.MaybeNumber
	case "os.EOL":
		call.Operation, call.Of = "eol", ir.String
	case "performance.timeOrigin":
		call.Operation, call.Of = "timeOrigin", ir.Number
	default:
		return nil, false, nil
	}
	if to, known := l.representation(l.checker.GetTypeAtLocation(node)); known {
		return fit(call, to), true, nil
	}
	return call, true, nil
}

func (l *lowering) nodeProcessMethodObservation(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || outer.Parent.Kind != ast.KindTypeOfExpression {
		return false
	}
	path := l.nodeProcessPath(node)
	switch path {
	case "performance.now", "performance.mark", "performance.measure", "performance.clearMarks", "performance.clearMeasures":
		return true
	}
	return false
}
