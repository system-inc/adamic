package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

func init() {
	RegisterNodeLibraryMembers("node:os.tmpdir", "node:os.platform", "node:os.EOL", "node:process.process", "node:globals.process")
	for _, name := range []string{"cwd", "chdir", "argv", "execArgv", "env", "platform", "pid", "stdout", "stderr", "exitCode", "exit", "memoryUsage", "nextTick"} {
		RegisterNodeLibraryMembers("node:process.Process."+name, "node:process."+name)
	}
	RegisterNodeLibraryMembers("node:process.MemoryUsage.heapUsed", "node:tty.WriteStream.columns", "node:tty.WriteStream.isTTY", "node:tty.ReadStream.isTTY", "node:stream.Writable.write", "node:net.Socket.write")
	RegisterNodeLibraryMembers("node:perf_hooks.performance", "node:performance.performance")
	for _, name := range []string{"timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures"} {
		RegisterNodeLibraryMembers("node:perf_hooks.Performance." + name)
	}
	for _, name := range []string{"name", "entryType", "startTime", "duration"} {
		RegisterNodeLibraryMembers("node:perf_hooks.PerformanceEntry." + name)
	}
	RegisterNodeLibraryMembers("node:perf_hooks.PerformanceMark.entryType", "node:perf_hooks.PerformanceMeasure.entryType")
}

// Ambient declarations are recognized by their module and declaration, never by a user's spelling.
func (l *lowering) nodeProcessPath(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
			symbol = l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
		}
		if symbol == nil || len(symbol.Declarations) == 0 {
			return ""
		}
		for _, declaration := range symbol.Declarations {
			if load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) && declaration.Kind == ast.KindModuleDeclaration && declaration.Name().Text() == "node:perf_hooks" {
				return "performanceModule"
			}
		}
		declaration := symbol.Declarations[0]
		if !load.IsPrelude(ast.GetSourceFileOfNode(declaration)) && !load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) {
			return ""
		}
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindModuleDeclaration {
				switch parent.Name().Text() {
				case "node:process", "process":
					return "process." + symbol.Name
				case "node:os", "os":
					return "os." + symbol.Name
				case "node:perf_hooks", "perf_hooks":
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
				if path == "performanceModule" && node.Name().Text() == "performance" {
					return "performance"
				}
				return path + "." + node.Name().Text()
			}
		}
	}
	return ""
}

func (l *lowering) nodeProcessEnvironmentKey(node *ast.Node) (*ast.Node, string, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if l.processPath(access.Expression) == "process.env" {
			return access.ArgumentExpression, "", true
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression && l.processPath(node.AsPropertyAccessExpression().Expression) == "process.env" {
		return nil, node.Name().Text(), true
	}
	return nil, "", false
}

func (l *lowering) nodeProcessEnvironmentMutation(node *ast.Node) (ir.Expression, bool, error) {
	var target, assigned *ast.Node
	operation := ""
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		target, assigned, operation = node.AsBinaryExpression().Left, node.AsBinaryExpression().Right, "envSet"
	} else if node.Kind == ast.KindDeleteExpression {
		target, operation = node.AsDeleteExpression().Expression, "envDelete"
	} else {
		return nil, false, nil
	}
	keyNode, keyText, known := l.nodeProcessEnvironmentKey(target)
	if !known {
		return nil, false, nil
	}
	key := ir.Expression(ir.StringConstant{Index: l.constant(keyText)})
	if keyNode != nil {
		var err error
		key, err = l.expression(keyNode)
		if err != nil {
			return nil, true, err
		}
		if key.Type() != ir.String {
			return nil, true, l.notYet(keyNode, "process.env mutation with a non-string key")
		}
	}
	call := ir.ProcessCall{Operation: operation, Arguments: []ir.Expression{key}, Of: ir.Boolean}
	if assigned != nil {
		value, err := l.expression(assigned)
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.String {
			return nil, true, l.notYet(assigned, "process.env assignment of a non-string value")
		}
		call.Arguments = append(call.Arguments, value)
		call.Of = ir.String
	}
	return call, true, nil
}

func (l *lowering) nodeProcessValue(node *ast.Node) (ir.Expression, bool, error) {
	// Node types call stdout a TTY WriteStream even when a pipe lacks these fields.
	// Preserve the runtime absence in guarded reads instead of unwrapping the declared number.
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken {
		binary := node.AsBinaryExpression()
		operation, of := "", ir.Type(0)
		switch l.processPath(binary.Left) {
		case "process.stdout.columns":
			operation, of = "columns", ir.MaybeNumber
		case "process.stdout.isTTY":
			operation, of = "stdoutTTY", ir.MaybeBoolean
		case "process.stderr.isTTY":
			operation, of = "stderrTTY", ir.MaybeBoolean
		}
		if operation != "" {
			fallback, err := l.expression(binary.Right)
			if err != nil {
				return nil, true, err
			}
			return ir.Coalesce{Value: ir.ProcessCall{Operation: operation, Of: of}, Fallback: fallback, Of: fallback.Type()}, true, nil
		}
	}

	if value, known, err := l.nodeProcessEnvironmentMutation(node); known {
		return value, known, err
	}
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
		case "process.cwd", "os.platform", "os.tmpdir":
			call.Operation, call.Of = "cwd", ir.String
			if path == "os.platform" {
				call.Operation = "platform"
			} else if path == "os.tmpdir" {
				call.Operation = "tmpdir"
			}
		case "process.chdir":
			call.Operation, call.Of, count, minimum = "chdir", ir.Object, 1, 1
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
			if strings.HasPrefix(path, "os.") || strings.HasPrefix(path, "performance.") {
				return nil, true, l.notYet(node, path)
			}
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
			if _, absent := value.(ir.Undefined); !absent && value.Type() != ir.String {
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
	case "performanceModule":
		parent := node.Parent
		if parent == nil || parent.Kind != ast.KindVariableDeclaration || parent.Name().Kind != ast.KindObjectBindingPattern {
			return nil, true, l.notYet(node, "node:perf_hooks namespace as a value outside performance destructuring")
		}
		for _, binding := range parent.Name().AsBindingPattern().Elements.Nodes {
			name := binding.Name().Text()
			if binding.AsBindingElement().PropertyName != nil {
				name = binding.AsBindingElement().PropertyName.Text()
			}
			if name != "performance" {
				return nil, true, l.notYet(node, "node:perf_hooks."+name)
			}
		}
		l.result.ClosuresMayThrow = true
		return ir.ObjectLiteral{Fields: []ir.Field{{Name: "performance", Value: ir.ProcessCall{Operation: "performance", Of: ir.Object}}}}, true, nil
	case "performance":
		l.result.ClosuresMayThrow = true
		call.Operation, call.Of = "performance", ir.Object
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
	case "process.stdout.isTTY", "process.stderr.isTTY":
		if of, known := l.representation(l.checker.GetTypeAtLocation(node)); known && of == ir.Boolean {
			return nil, true, l.notYet(node, path+" without a missing-value guard on the Node TTY declaration")
		}
		return nil, false, nil
	case "process.stdout.columns":
		return nil, true, l.notYet(node, path+" without a missing-value guard on the Node TTY declaration")
	case "os.EOL":
		call.Operation, call.Of = "eol", ir.String
	case "performance.timeOrigin":
		call.Operation, call.Of = "timeOrigin", ir.Number
	default:
		if strings.HasPrefix(path, "os.") || strings.HasPrefix(path, "performance.") {
			return nil, true, l.notYet(node, path)
		}
		return nil, false, nil
	}
	if to, known := l.representation(l.checker.GetTypeAtLocation(node)); known {
		return fit(call, to), true, nil
	}
	return call, true, nil
}

func (l *lowering) nodeProcessMethodObservation(node *ast.Node) bool {
	path := l.processPath(node)
	if path == "process.nextTick" && node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
		inner := node.Parent.AsPrefixUnaryExpression()
		outer := node.Parent.Parent
		if inner.Operator == ast.KindExclamationToken && outer != nil && outer.Kind == ast.KindPrefixUnaryExpression && outer.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			return true
		}
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || outer.Parent.Kind != ast.KindTypeOfExpression {
		return false
	}
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "setBlocking" {
		receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
		if ast.IsIdentifier(receiver) {
			symbol := l.symbol(receiver)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if declaration.Kind == ast.KindVariableDeclaration && declaration.AsVariableDeclaration().Initializer != nil && l.processPath(declaration.AsVariableDeclaration().Initializer) == "process.stdout._handle" {
						return true
					}
				}
			}
		}
	}
	path = l.nodeProcessPath(node)
	if path == "" {
		path = l.processPath(node)
	}
	if name := l.nodeLibraryMember(node); strings.HasPrefix(name, "node:perf_hooks.Performance.") {
		path = "performance." + node.Name().Text()
	}
	switch path {
	case "process.nextTick", "process.cwd", "process.memoryUsage", "performance.now", "performance.mark", "performance.measure", "performance.clearMarks", "performance.clearMeasures":
		return true
	}
	return false
}

// Only the external environment has deletable keys; ordinary object shapes remain fixed.
func (l *lowering) nodeProcessEnvironmentDelete(node *ast.Node) bool {
	if node.Kind != ast.KindDeleteExpression {
		return false
	}
	_, _, known := l.nodeProcessEnvironmentKey(node.AsDeleteExpression().Expression)
	return known
}
