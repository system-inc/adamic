package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

func init() {
	RegisterNodeLibraryMembers("node:os.tmpdir", "node:os.platform", "node:os.EOL", "node:process.process", "node:globals.process")
	for _, name := range []string{"cwd", "chdir", "argv", "execArgv", "execPath", "env", "platform", "pid", "stdout", "stderr", "exitCode", "exit", "memoryUsage", "nextTick"} {
		RegisterNodeLibraryMembers("node:process.Process."+name, "node:process."+name)
	}
	RegisterNodeLibraryMembers("node:process.MemoryUsage.heapUsed", "node:tty.WriteStream.columns", "node:tty.WriteStream.isTTY", "node:tty.ReadStream.isTTY", "node:stream.Writable.write", "node:net.Socket.write")
	RegisterNodeLibraryMembers("node:process.MemoryUsage.rss", "node:process.MemoryUsage.heapTotal", "node:process.MemoryUsage.external", "node:process.MemoryUsage.arrayBuffers", "node:module.__filename", "node:module.__dirname", "node:globals.__filename", "node:globals.__dirname")
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
			if load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) && declaration.Kind == ast.KindModuleDeclaration {
				switch declaration.Name().Text() {
				case "node:perf_hooks":
					return "performanceModule"
				case "node:os":
					return "os"
				case "node:process":
					return "process"
				}
			}
		}
		declaration := symbol.Declarations[0]
		if module := l.nodeRequireBinding(declaration); module != "" {
			if module == "perf_hooks" {
				return "performanceModule"
			}
			return module
		}
		if !load.IsPrelude(ast.GetSourceFileOfNode(declaration)) && !load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) {
			return ""
		}
		if load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) && (symbol.Name == "__filename" || symbol.Name == "__dirname") {
			return symbol.Name
		}
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindModuleDeclaration {
				switch parent.Name().Text() {
				case "node:process", "process":
					if symbol.Name == "process" {
						return "process"
					}
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
	var target *ast.Node
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		target = node.AsBinaryExpression().Left
	} else if node.Kind == ast.KindDeleteExpression {
		target = node.AsDeleteExpression().Expression
	} else {
		return nil, false, nil
	}
	_, _, known := l.nodeProcessEnvironmentKey(target)
	if !known {
		return nil, false, nil
	}
	return nil, true, &Refused{
		Where: l.program.Where(node),
		What:  "process.env mutation",
		Fix:   "supply environment variables when starting the program; this step supports live reads only",
	}
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
		if l.nodeProcessHostBinding(operand) {
			return ir.StringConstant{Index: l.constant("object")}, true, nil
		}
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
		callee := ast.SkipParentheses(written.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "now" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Date") {
			if len(written.Arguments.Nodes) != 0 {
				return nil, true, l.notYet(node, "Date.now with arguments")
			}
			return ir.ProcessCall{Operation: "dateNow", Of: ir.Number}, true, nil
		}
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
	case "os":
		return nil, true, l.notYet(node, "node:os namespace as a first-class value")
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
	case "process.execPath", "__filename":
		call.Operation, call.Of = "execPath", ir.String
		if path == "__filename" {
			call.Operation = "filename"
		}
	case "__dirname":
		call.Operation, call.Of = "dirname", ir.String
	case "process.stdout":
		call.Operation, call.Of = "stdout", ir.Object
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

// A presence descriptor is narrower than a callable method implementation.
// These static host receivers have no evaluation effects; arbitrary typed objects
// still use the compiler's receiver evaluation and runtime method lookup.
func (l *lowering) nodeProcessMethodPresence(node *ast.Node) (ir.Expression, bool) {
	if l.nodeFSRealpathPresence(node) {
		return ir.BooleanConstant{Value: true}, true
	}
	path := l.processPath(node)
	if path == "" {
		path = l.nodeProcessPath(node)
	}
	// tsc's probe redeclares the global process with its official NodeJS type.
	// An ambient global of another name, or a user method merely named nextTick,
	// is not evidence of this host's process object.
	if path == "" && node.Kind == ast.KindPropertyAccessExpression && l.nodeProcessHostBinding(node.AsPropertyAccessExpression().Expression) {
		switch l.nodeLibraryMember(node) {
		case "node:process.Process.nextTick", "node:process.Process.cwd", "node:process.Process.memoryUsage":
			path = "process." + node.Name().Text()
		}
	}
	switch path {
	case "process.nextTick":
		// This is only the Node-like feature observation, not callback scheduling.
		return ir.ProcessCall{Operation: "nextTickFeature", Of: ir.Boolean}, true
	case "process.cwd", "process.memoryUsage", "performance.now", "performance.mark", "performance.measure", "performance.clearMarks", "performance.clearMeasures":
		return ir.BooleanConstant{Value: true}, true
	}
	return nil, false
}

// An ambient declaration of the official process global names the host binding,
// not a null-initialized user object. A different name or user type proves nothing.
func (l *lowering) nodeProcessHostBinding(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if l.processPath(node) == "process" {
		return true
	}
	if !ast.IsIdentifier(node) || node.Text() != "process" {
		return false
	}
	binding := l.symbol(node)
	if binding == nil {
		return false
	}
	host := l.checker.GetTypeAtLocation(node).Symbol()
	if host == nil || host.Name != "Process" {
		return false
	}
	official := false
	for _, declaration := range host.Declarations {
		official = official || load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration))
	}
	if !official {
		return false
	}
	for _, declaration := range binding.Declarations {
		for at := declaration; at != nil && at.Kind != ast.KindSourceFile; at = at.Parent {
			if ast.HasSyntacticModifier(at, ast.ModifierFlagsAmbient) {
				return true
			}
		}
	}
	return false
}

// sys.ts reads the native variant from its static fs module. Never erase an
// arbitrary receiver expression merely because it has realpathSync's type.
func (l *lowering) nodeFSRealpathPresence(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || node.Name().Text() != "native" {
		return false
	}
	function := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	module, member := l.nodeHostMember(function)
	if module != "node:fs" || member != "realpathSync" {
		return false
	}
	if ast.IsIdentifier(function) {
		symbol := l.symbol(function)
		return symbol != nil && symbol.Name == "realpathSync" && len(symbol.Declarations) > 0 && load.IsNodeLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
	}
	if function.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	receiver := ast.SkipParentheses(function.AsPropertyAccessExpression().Expression)
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := l.symbol(receiver)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindModuleDeclaration && load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) {
			name := declaration.Name().Text()
			if name == "fs" || name == "node:fs" {
				return true
			}
		}
	}
	return false
}
