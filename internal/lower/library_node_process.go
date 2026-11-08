package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

func init() {
	RegisterNodeLibraryMembers(
		"node:os.tmpdir",
		"node:os.platform",
		"node:os.EOL",
		"node:process.process",
		"node:globals.process",
	)
	for _, name := range []string{
		"cwd",
		"chdir",
		"argv",
		"execArgv",
		"env",
		"platform",
		"pid",
		"stdout",
		"stderr",
		"exitCode",
		"exit",
		"memoryUsage",
		"nextTick",
	} {
		RegisterNodeLibraryMembers("node:process.Process."+name, "node:process."+name)
	}
	RegisterNodeLibraryMembers(
		"node:process.MemoryUsage.heapUsed",
		"node:stream.Writable.write",
		"node:net.Socket.write",
	)
	RegisterNodeLibraryMembers("node:perf_hooks.performance", "node:performance.performance")
	for _, name := range []string{"timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures"} {
		RegisterNodeLibraryMembers("node:perf_hooks.Performance." + name)
	}
	for _, name := range []string{"name", "entryType", "startTime", "duration"} {
		RegisterNodeLibraryMembers("node:perf_hooks.PerformanceEntry." + name)
	}
	RegisterNodeLibraryMembers(
		"node:perf_hooks.PerformanceMark.entryType",
		"node:perf_hooks.PerformanceMeasure.entryType",
	)
}

// Ambient declarations are recognized by their module and declaration, never by a user's spelling.
func (lowering *lowering) nodeProcessPath(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := lowering.symbol(node)
		if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
			symbol = lowering.checker.GetShorthandAssignmentValueSymbol(node.Parent)
		}
		if symbol == nil || len(symbol.Declarations) == 0 {
			return ""
		}
		for _, declaration := range symbol.Declarations {
			if load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) &&
				declaration.Kind == ast.KindModuleDeclaration {
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
		if !load.IsPrelude(ast.GetSourceFileOfNode(declaration)) &&
			!load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) {
			return ""
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
			if path := lowering.nodeProcessPath(access.Expression); path != "" {
				if path == "performanceModule" && node.Name().Text() == "performance" {
					return "performance"
				}
				return path + "." + node.Name().Text()
			}
		}
	}
	return ""
}

// Mutations share main's declaration recognition, refusal order and key conversion.
func (lowering *lowering) nodeProcessEnvironmentMutation(node *ast.Node) (ir.Expression, bool, error) {
	node = ast.SkipParentheses(node)
	var target, assigned *ast.Node
	operation := "host_env_delete"
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		target, assigned, operation = node.AsBinaryExpression().Left, node.AsBinaryExpression().Right, "host_env_set"
	} else if node.Kind == ast.KindDeleteExpression {
		target = node.AsDeleteExpression().Expression
	} else {
		return nil, false, nil
	}
	target = ast.SkipParentheses(target)
	if _, known := lowering.nodeEnvironmentAccess(target); !known {
		return nil, false, nil
	}
	if nodeEnvironmentReadRefusalError := lowering.nodeEnvironmentReadRefusal(
		target,
	); nodeEnvironmentReadRefusalError != nil {
		return nil, true, nodeEnvironmentReadRefusalError
	}
	key, nodeEnvironmentKeyError := lowering.nodeEnvironmentKey(target)
	if nodeEnvironmentKeyError != nil {
		return nil, true, nodeEnvironmentKeyError
	}
	call := ir.NodeFSFile{Operation: operation, Arguments: []ir.Expression{key}, Of: ir.Boolean}
	if assigned != nil {
		value, expressionError := lowering.expression(assigned)
		if expressionError != nil {
			return nil, true, expressionError
		}
		call.Arguments = append(call.Arguments, value)
		call.Of = ir.String
	}
	return call, true, nil
}

// Error.stackTraceLimit is Node's writable frame count on the global Error, 10 by default. Adamic
// refuses .stack, so the count is all a program can observe of it. As with process.platform the
// member isn't registered: a position nodeStackTraceLimitUseRefusal misses falls to the generic NotYet.
const nodeStackTraceLimitMember = "node:globals.ErrorConstructor.stackTraceLimit"

// nodeStackTraceLimitRefusal names every use of Error.stackTraceLimit but a plain read, or a plain =
// write of a number, through the global Error identifier.
func (lowering *lowering) nodeStackTraceLimitRefusal(node *ast.Node) error {
	if node.Kind != ast.KindPropertyAccessExpression {
		return lowering.notYet(
			node,
			nodeStackTraceLimitMember+" outside a property access (fix: read or write Error.stackTraceLimit)",
		)
	}
	if !lowering.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Error") {
		return lowering.notYet(node, nodeStackTraceLimitMember+
			" through a receiver other than the global Error (fix: read or write Error.stackTraceLimit)")
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	switch {
	case node.Flags&ast.NodeFlagsOptionalChain != 0:
		return lowering.notYet(
			node,
			nodeStackTraceLimitMember+" read through an optional chain (fix: read Error.stackTraceLimit directly)",
		)
	case outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression:
		return lowering.notYet(node, nodeStackTraceLimitMember+" deleted (fix: assign it a number)")
	case ast.IsAssignmentTarget(outer):
		parent := outer.Parent
		if parent == nil ||
			parent.Kind != ast.KindBinaryExpression ||
			parent.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken ||
			parent.AsBinaryExpression().Left != outer {
			return lowering.notYet(
				node,
				nodeStackTraceLimitMember+" written other than by a plain = (fix: assign the computed number with =)",
			)
		}
		if of, known := lowering.representation(
			lowering.checker.GetTypeAtLocation(parent.AsBinaryExpression().Right),
		); !known || of != ir.Number {
			return lowering.notYet(
				parent.AsBinaryExpression().Right,
				nodeStackTraceLimitMember+" written with a value not typed number",
			)
		}
	}
	return nil
}

// nodeStackTraceLimitUseRefusal is the refusal visitor's view of Error.stackTraceLimit: its property
// name is part of the access, the access is judged by nodeStackTraceLimitRefusal, and the member
// anywhere else is NotYet.
func (lowering *lowering) nodeStackTraceLimitUseRefusal(node *ast.Node) (bool, error) {
	if node.Parent != nil &&
		node.Parent.Kind == ast.KindPropertyAccessExpression &&
		node.Parent.Name() == node &&
		lowering.nodeLibraryMember(node.Parent) == nodeStackTraceLimitMember {
		return true, lowering.nodeStackTraceLimitRefusal(node.Parent)
	}
	if lowering.nodeLibraryMember(node) != nodeStackTraceLimitMember {
		return false, nil
	}
	return true, lowering.nodeStackTraceLimitRefusal(node)
}

// nodeStackTraceLimit lowers a read of Error.stackTraceLimit to the runtime's stored count, and a
// plain = write of a number to a store that yields the number, as the assignment does in JavaScript.
func (lowering *lowering) nodeStackTraceLimit(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		target := ast.SkipParentheses(node.AsBinaryExpression().Left)
		if target.Kind != ast.KindPropertyAccessExpression ||
			lowering.nodeLibraryMember(target) != nodeStackTraceLimitMember {
			return nil, false, nil
		}
		if nodeStackTraceLimitRefusalError := lowering.nodeStackTraceLimitRefusal(
			target,
		); nodeStackTraceLimitRefusalError != nil {
			return nil, true, nodeStackTraceLimitRefusalError
		}
		value, expressionError := lowering.expression(node.AsBinaryExpression().Right)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if value.Type() != ir.Number {
			return nil, true, lowering.notYet(
				node.AsBinaryExpression().Right,
				nodeStackTraceLimitMember+" written with a value not typed number",
			)
		}
		return ir.ProcessCall{
			Operation: "setStackTraceLimit",
			Arguments: []ir.Expression{value},
			Of:        ir.Number,
		}, true, nil
	}
	if (node.Kind != ast.KindPropertyAccessExpression && node.Kind != ast.KindElementAccessExpression) ||
		lowering.nodeLibraryMember(node) != nodeStackTraceLimitMember {
		return nil, false, nil
	}
	if nodeStackTraceLimitRefusalError := lowering.nodeStackTraceLimitRefusal(
		node,
	); nodeStackTraceLimitRefusalError != nil {
		return nil, true, nodeStackTraceLimitRefusalError
	}
	return ir.ProcessCall{Operation: "stackTraceLimit", Of: ir.Number}, true, nil
}

func (lowering *lowering) nodeProcessValue(node *ast.Node) (ir.Expression, bool, error) {

	if value, known, nodeProcessEnvironmentMutationError := lowering.nodeProcessEnvironmentMutation(node); known {
		return value, known, nodeProcessEnvironmentMutationError
	}
	if value, known, nodeStackTraceLimitError := lowering.nodeStackTraceLimit(node); known {
		return value, known, nodeStackTraceLimitError
	}
	if value, known, nodeGlobalsValueError := lowering.nodeGlobalsValue(node); known {
		return value, known, nodeGlobalsValueError
	}
	if value, known, nodeCommonJsRequireValueError := lowering.nodeCommonJsRequireValue(node); known {
		return value, known, nodeCommonJsRequireValueError
	}
	if node.Kind == ast.KindPrefixUnaryExpression {
		outer := node.AsPrefixUnaryExpression()
		inner := ast.SkipParentheses(outer.Operand)
		if outer.Operator == ast.KindExclamationToken && inner.Kind == ast.KindPrefixUnaryExpression {
			innerPrefix := inner.AsPrefixUnaryExpression()
			if innerPrefix.Operator == ast.KindExclamationToken &&
				lowering.processPath(innerPrefix.Operand) == "process.nextTick" {
				return ir.ProcessCall{Operation: "nextTickFeature", Of: ir.Boolean}, true, nil
			}
		}
	}

	if node.Kind == ast.KindTypeOfExpression {
		operand := node.AsTypeOfExpression().Expression
		if lowering.nodeProcessHostBinding(operand) {
			return ir.StringConstant{Index: lowering.constant("object")}, true, nil
		}
		path := lowering.processPath(operand)
		if path == "" {
			path = lowering.nodeProcessPath(operand)
		}
		switch path {
		case "process.nextTick", "performance.now", "performance.mark", "performance.measure", "performance.clearMarks", "performance.clearMeasures", "process.cwd", "process.memoryUsage":
			return ir.StringConstant{Index: lowering.constant("function")}, true, nil
		}
	}
	path := lowering.processPath(node)
	if path == "" {
		path = lowering.nodeProcessPath(node)
	}
	call := ir.ProcessCall{}
	if node.Kind == ast.KindCallExpression {
		written := node.AsCallExpression()
		path = lowering.processPath(written.Expression)
		if path == "" {
			path = lowering.nodeProcessPath(written.Expression)
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
				return nil, true, lowering.notYet(node, "observing the backpressure return of process.stdout.write")
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
				return nil, true, lowering.notYet(node, path)
			}
			return nil, false, nil
		}
		if len(written.Arguments.Nodes) < minimum || len(written.Arguments.Nodes) > count {
			return nil, true, lowering.notYet(node, path+" with arguments outside the census contract")
		}
		for _, argument := range written.Arguments.Nodes {
			value, expressionError := lowering.expression(argument)
			if expressionError != nil {
				return nil, true, expressionError
			}
			if _, absent := value.(ir.Undefined); !absent && value.Type() != ir.String {
				return nil, true, lowering.notYet(argument, path+" with a non-string argument")
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
		return nil, true, lowering.notYet(node, "node:os namespace as a first-class value")
	case "performanceModule":
		parent := node.Parent
		if parent == nil ||
			parent.Kind != ast.KindVariableDeclaration ||
			parent.Name().Kind != ast.KindObjectBindingPattern {
			return nil, true, lowering.notYet(
				node,
				"node:perf_hooks namespace as a value outside performance destructuring",
			)
		}
		for _, binding := range parent.Name().AsBindingPattern().Elements.Nodes {
			name := binding.Name().Text()
			if binding.AsBindingElement().PropertyName != nil {
				name = binding.AsBindingElement().PropertyName.Text()
			}
			if name != "performance" {
				return nil, true, lowering.notYet(node, "node:perf_hooks."+name)
			}
		}
		lowering.result.ClosuresMayThrow = true
		return ir.ObjectLiteral{
			Fields: []ir.Field{{Name: "performance", Value: ir.ProcessCall{Operation: "performance", Of: ir.Object}}},
		}, true, nil
	case "performance":
		lowering.result.ClosuresMayThrow = true
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
	case "os.EOL":
		call.Operation, call.Of = "eol", ir.String
	case "performance.timeOrigin":
		call.Operation, call.Of = "timeOrigin", ir.Number
	default:
		if strings.HasPrefix(path, "os.") || strings.HasPrefix(path, "performance.") {
			return nil, true, lowering.notYet(node, path)
		}
		return nil, false, nil
	}
	if to, known := lowering.representation(lowering.checker.GetTypeAtLocation(node)); known {
		return fit(call, to), true, nil
	}
	return call, true, nil
}

func (lowering *lowering) nodeProcessMethodObservation(node *ast.Node) bool {
	if lowering.nodeCaptureStackTraceRead(node) {
		return true
	}
	path := lowering.processPath(node)
	if path == "process.nextTick" && node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
		inner := node.Parent.AsPrefixUnaryExpression()
		outer := node.Parent.Parent
		if inner.Operator == ast.KindExclamationToken &&
			outer != nil &&
			outer.Kind == ast.KindPrefixUnaryExpression &&
			outer.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
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
			symbol := lowering.symbol(receiver)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if declaration.Kind == ast.KindVariableDeclaration &&
						declaration.AsVariableDeclaration().Initializer != nil &&
						lowering.processPath(
							declaration.AsVariableDeclaration().Initializer,
						) == "process.stdout._handle" {
						return true
					}
				}
			}
		}
	}
	path = lowering.nodeProcessPath(node)
	if path == "" {
		path = lowering.processPath(node)
	}
	if name := lowering.nodeLibraryMember(node); strings.HasPrefix(name, "node:perf_hooks.Performance.") {
		path = "performance." + node.Name().Text()
	}
	switch path {
	case "process.nextTick", "process.cwd", "process.memoryUsage", "performance.now", "performance.mark", "performance.measure", "performance.clearMarks", "performance.clearMeasures":
		return true
	}
	return false
}

// Only the external environment has deletable keys; ordinary object shapes remain fixed.
func (lowering *lowering) nodeProcessEnvironmentDelete(node *ast.Node) bool {
	if node.Kind != ast.KindDeleteExpression {
		return false
	}
	_, known := lowering.nodeEnvironmentAccess(node.AsDeleteExpression().Expression)
	return known
}

// A presence descriptor is narrower than a callable method implementation.
// These static host receivers have no evaluation effects; arbitrary typed objects
// still use the compiler's receiver evaluation and runtime method lookup.
func (lowering *lowering) nodeProcessMethodPresence(node *ast.Node) (ir.Expression, bool) {
	if lowering.nodeFileSystemRealpathPresence(node) {
		return ir.BooleanConstant{Value: true}, true
	}
	path := lowering.processPath(node)
	if path == "" {
		path = lowering.nodeProcessPath(node)
	}
	// tsc's probe redeclares the global process with its official NodeJS type.
	// An ambient global of another name, or a user method merely named nextTick,
	// is not evidence of this host's process object.
	if path == "" &&
		node.Kind == ast.KindPropertyAccessExpression &&
		lowering.nodeProcessHostBinding(node.AsPropertyAccessExpression().Expression) {
		switch lowering.nodeLibraryMember(node) {
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
func (lowering *lowering) nodeProcessHostBinding(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if lowering.processPath(node) == "process" {
		return true
	}
	if !ast.IsIdentifier(node) || node.Text() != "process" {
		return false
	}
	binding := lowering.symbol(node)
	if binding == nil {
		return false
	}
	host := lowering.checker.GetTypeAtLocation(node).Symbol()
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
func (lowering *lowering) nodeFileSystemRealpathPresence(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || node.Name().Text() != "native" {
		return false
	}
	function := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	module, member := lowering.nodeHostMember(function)
	if module != "node:fs" || member != "realpathSync" {
		return false
	}
	if ast.IsIdentifier(function) {
		symbol := lowering.symbol(function)
		return symbol != nil &&
			symbol.Name == "realpathSync" &&
			len(symbol.Declarations) > 0 &&
			load.IsNodeLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
	}
	if function.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	receiver := ast.SkipParentheses(function.AsPropertyAccessExpression().Expression)
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := lowering.symbol(receiver)
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

// Const aliases keep the full overloaded Node declaration. Its options overload
// must not reach the string-only native closure ABI through a generic method call.
func (lowering *lowering) nodeProcessSignature(node *ast.Node) error {
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	name := lowering.nodeLibraryMember(node)
	prefix := "node:perf_hooks.Performance."
	if !strings.HasPrefix(name, prefix) {
		return nil
	}
	member := strings.TrimPrefix(name, prefix)
	minimum, maximum := 0, 0
	switch member {
	case "now":
	case "mark":
		minimum, maximum = 1, 1
	case "measure":
		minimum, maximum = 1, 3
	case "clearMarks", "clearMeasures":
		maximum = 1
	default:
		return lowering.notYet(node, "performance."+member)
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) < minimum || len(arguments) > maximum {
		return lowering.notYet(node, "performance."+member+" outside the string host contract")
	}
	var stringArgument func(*checker.Type, bool) bool
	stringArgument = func(proven *checker.Type, optional bool) bool {
		if proven.Flags()&checker.TypeFlagsUnion != 0 {
			for _, part := range proven.Types() {
				if !stringArgument(part, optional) {
					return false
				}
			}
			return true
		}
		return proven.Flags()&checker.TypeFlagsStringLike != 0 ||
			optional && proven.Flags()&checker.TypeFlagsUndefined != 0
	}
	for index, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement ||
			!stringArgument(lowering.checker.GetTypeAtLocation(argument), index >= minimum) {
			return lowering.notYet(argument, "performance."+member+" with an argument outside the string host ABI")
		}
	}
	return nil
}

// Structural interfaces can hide the overloaded declaration. Once a program
// holds performance, unresolved calls and writes to its slot names must retain
// the runtime's string/number ABI; otherwise refuse before C emission.
func (lowering *lowering) nodeProcessStorageAbi(modules []*ast.SourceFile) error {
	host := false
	var failure string
	check := func(statements []ir.Statement) {
		walk(statements, func(value any) bool {
			if call, ok := value.(ir.ProcessCall); ok && call.Operation == "performance" {
				host = true
			}
			switch node := value.(type) {
			case ir.SetProperty:
				switch node.Name {
				case "timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures":
					failure = "writing performance host slots through a structural view"
				}
			case ir.CallClosure:
				property, ok := node.Closure.(ir.Property)
				if !ok {
					break
				}
				minimum, maximum, returns := 0, 0, ir.Number
				switch property.Name {
				case "now":
				case "mark":
					minimum, maximum, returns = 1, 1, ir.Object
				case "measure":
					minimum, maximum, returns = 1, 3, ir.Object
				case "clearMarks", "clearMeasures":
					maximum, returns = 1, 0
				default:
					break
				}
				if property.Name != "now" &&
					property.Name != "mark" &&
					property.Name != "measure" &&
					property.Name != "clearMarks" &&
					property.Name != "clearMeasures" {
					break
				}
				valid := len(node.Arguments) >= minimum && len(node.Arguments) <= maximum
				for index, argument := range node.Arguments {
					_, absent := argument.(ir.Undefined)
					valid = valid && argument.Type() == ir.String && (!absent || index >= minimum)
				}
				for _, spread := range node.Spread {
					valid = valid && !spread
				}
				valid = valid && (node.Returns == returns || returns == ir.Object && node.Returns == 0)
				if !valid {
					failure = "indirect performance." + property.Name + " outside the string host ABI"
				}
			}
			return true
		})
	}
	check(lowering.result.Main)
	for _, function := range lowering.result.Functions {
		check(function.Body)
	}
	if host && failure != "" {
		return lowering.notYet(modules[0].AsNode(), failure)
	}
	return nil
}

// The narrow process adapter owns only the optional stdout handle.
// Keep the raw stdout scanner's refusals for all other stream observations.
func (lowering *lowering) nodeProcessGuardedStdoutUse(node *ast.Node) bool {
	outer := node
	if outer.Parent != nil && outer.Parent.Kind == ast.KindPropertyAccessExpression && outer.Parent.Name() == outer {
		outer = outer.Parent
	}
	for outer.Parent != nil {
		parent := outer.Parent
		if parent.Kind == ast.KindParenthesizedExpression ||
			parent.Kind == ast.KindAsExpression ||
			parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == outer {
			outer = parent
			continue
		}
		break
	}
	return lowering.processPath(outer) == "process.stdout._handle"
}
