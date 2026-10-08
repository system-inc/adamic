package lower

import (
	"fmt"
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

var errorFamilies = []string{"Object", "Error", "TypeError", "RangeError", "SyntaxError", "ReferenceError", "EvalError", "URIError", "AggregateError"}

func (l *lowering) errorConstructor(node *ast.Node) (int, bool) {
	for kind, name := range errorFamilies {
		if kind > 0 && l.isLibraryGlobal(ast.SkipParentheses(node), name) {
			return kind, true
		}
	}
	return 0, false
}
func (l *lowering) errorObjectType(node *ast.Node) bool {
	return l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)), errorFamilies[1:]...)
}

// An Error interface is structural. Until the compiler proves the actual
// object origin, refuse a program containing a compatible non-Error shape;
// never reinterpret that plain object's fields as Error internal slots.
func (l *lowering) errorSurfaceHazard(node *ast.Node) bool {
	view := l.checker.GetTypeAtLocation(node)
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	hazard := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if hazard {
			return true
		}
		if n.Kind == ast.KindObjectLiteralExpression || n.Kind == ast.KindNewExpression {
			shape := l.checker.GetTypeAtLocation(n)
			if !l.isLibraryType(shape, errorFamilies[1:]...) && l.checker.IsTypeAssignableTo(shape, view) {
				hazard = true
				return true
			}
		}
		n.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return hazard
}
func (l *lowering) errorOwnProof(node *ast.Node) bool {
	return l.errorNativeTarget(node) || l.errorBuiltinPrototype(node)
}
func (l *lowering) errorHostImports() bool {
	for _, file := range l.program.Files() {
		found := false
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if ast.IsPartOfTypeNode(n) {
				return false
			}
			if l.isLibraryGlobal(n, "process") || l.isLibraryGlobal(n, "Buffer") {
				found = true
				return true
			}
			return n.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if found {
			return true
		}
		for _, statement := range file.Statements.Nodes {
			if statement.Kind == ast.KindImportDeclaration && runtimeModuleStatement(statement) {
				text := statement.ModuleSpecifier().Text()
				if text == "adamic" {
					continue
				}
				if text == "fs" || text == "path" || text == "os" || text == "crypto" || text == "process" || text == "buffer" || len(text) >= 5 && text[:5] == "node:" {
					return true
				}
			}
		}
	}
	return false
}
func (l *lowering) errorBuiltinPrototype(node *ast.Node) bool {
	for depth := 0; depth < 16; depth++ {
		node = ast.SkipParentheses(node)
		if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" {
			_, ok := l.errorConstructor(node.AsPropertyAccessExpression().Expression)
			return ok
		}
		if !ast.IsIdentifier(node) {
			return false
		}
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
		if node == nil {
			return false
		}
	}
	return false
}
func (l *lowering) errorObjectSyntax(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if l.errorObjectType(node) {
		return true
	}
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" {
		_, ok := l.errorConstructor(node.AsPropertyAccessExpression().Expression)
		return ok
	}
	if node.Kind == ast.KindCallExpression {
		c := node.AsCallExpression()
		callee := ast.SkipParentheses(c.Expression)
		return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "getPrototypeOf" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") && len(c.Arguments.Nodes) == 1 && l.errorObjectSyntax(c.Arguments.Nodes[0])
	}
	return false
}

// Calls ending at null are admitted; invoking getPrototypeOf on that null
// needs compiler-owned catchable TypeError lowering.
func (l *lowering) errorPrototypeDepth(node *ast.Node) int { return l.errorPrototypeDepthAt(node, 0) }
func (l *lowering) errorPrototypeDepthAt(node *ast.Node, depth int) int {
	if depth >= 16 {
		return 0
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "prototype" {
		if kind, ok := l.errorConstructor(node.AsPropertyAccessExpression().Expression); ok {
			if kind == 1 {
				return 2
			}
			return 3
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if kind, ok := l.errorConstructor(callee); ok {
			if kind == 1 {
				return 3
			}
			return 4
		}
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "getPrototypeOf" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") && len(nodesOf(call.Arguments)) == 1 {
			return l.errorPrototypeDepthAt(nodesOf(call.Arguments)[0], depth+1) - 1
		}
	}
	if node.Kind == ast.KindNewExpression {
		if kind, ok := l.errorConstructor(node.AsNewExpression().Expression); ok {
			if kind == 1 {
				return 3
			}
			return 4
		}
	}
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.AsVariableDeclaration().Initializer != nil {
				return l.errorPrototypeDepthAt(declaration.AsVariableDeclaration().Initializer, depth+1)
			}
		}
	}
	// An opaque structural Error can also be its prototype. Do not assume an
	// instance's extra ancestor and accidentally call getPrototypeOf(null).
	if l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)), errorFamilies[2:]...) {
		return 3
	}
	return 2
}
func (l *lowering) errorFrame(node *ast.Node) ir.Expression {
	name := "main"
	if l.function != nil {
		name = fmt.Sprintf("adamic_function_%d_%s", l.functionIndex, regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(l.function.Name, ""))
	}
	return ir.StringConstant{Index: l.constant(fmt.Sprintf("\nAdamic frame: %s (%v)", name, l.program.Where(node)))}
}
func (l *lowering) errorConstruct(node, callee *ast.Node, arguments []*ast.Node) (ir.Expression, error) {
	kind, _ := l.errorConstructor(callee)
	messageIndex := 0
	result := ir.MakeError{Family: errorFamilies[kind], MessageAbsent: true, Message: ir.StringConstant{Index: l.constant("")}, Frames: l.errorFrame(node), Limit: ir.Read{Local: l.errorLimitLocal(), Of: ir.Number}}
	if kind == 8 {
		if len(arguments) == 0 {
			return nil, l.notYet(node, "AggregateError requires a supported array of errors")
		}
		value, err := l.expression(arguments[0])
		if err != nil {
			return nil, err
		}
		if value.Type() != ir.Array {
			return nil, l.notYet(node, "AggregateError iterable protocols other than a represented array")
		}
		element, err := l.elementType(arguments[0])
		if err != nil {
			return nil, err
		}
		if element != ir.Number && element != ir.Boolean && !element.IsReference() {
			return nil, l.notYet(node, "AggregateError elements without a boxed representation")
		}
		result.Errors, result.ErrorElement = value, element
		messageIndex = 1
	}
	if len(arguments) > messageIndex+2 {
		return nil, l.notYet(node, "Error constructor extra arguments require evaluation-order lowering")
	}
	if len(arguments) > messageIndex {
		argument := arguments[messageIndex]
		if l.checker.GetTypeAtLocation(argument).Flags()&checker.TypeFlagsUndefined == 0 {
			if l.includesUndefined(l.checker.GetTypeAtLocation(argument)) {
				return nil, l.notYet(argument, "Error message may be undefined: compiler conditional message-presence lowering required")
			}
			value, err := l.expression(argument)
			if err != nil {
				return nil, err
			}
			switch value.Type() {
			case ir.String:
			case ir.Number:
				value = ir.NumberToString{Value: value}
			case ir.Boolean:
				value = ir.BooleanToString{Value: value}
			default:
				if ast.SkipParentheses(argument).Kind == ast.KindNullKeyword {
					value = ir.StringConstant{Index: l.constant("null")}
				} else {
					return nil, l.notYet(node, "Error message object coercion requires ToPrimitive lowering")
				}
			}
			result.Message, result.MessageAbsent = value, false
		} else if !ast.IsIdentifier(ast.SkipParentheses(argument)) || ast.SkipParentheses(argument).Text() != "undefined" {
			return nil, l.notYet(argument, "Error message exact undefined expression must be evaluated by compiler")
		}
	}
	if len(arguments) > messageIndex+1 {
		options := ast.SkipParentheses(arguments[messageIndex+1])
		if options.Kind != ast.KindObjectLiteralExpression {
			return nil, l.notYet(options, "Error options require a plain literal with a supported cause value")
		}
		for _, property := range options.AsObjectLiteralExpression().Properties.Nodes {
			if (property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment) || property.Name().Text() != "cause" {
				return nil, l.notYet(property, "Error options spreads, getters and other fields require HasProperty/Get lowering")
			}
			argument := property.Name()
			if property.Kind == ast.KindPropertyAssignment {
				argument = property.AsPropertyAssignment().Initializer
			}
			var value ir.Expression
			var err error
			if property.Kind == ast.KindShorthandPropertyAssignment {
				value, err = l.shorthand(property)
			} else {
				value, err = l.value(argument)
			}
			if err != nil {
				return nil, err
			}
			result.Cause = fit(value, ir.Union)
		}
	}
	return result, nil
}

func (l *lowering) errorObjectsValue(node *ast.Node) (ir.Expression, bool, error) {
	if l.errorObjectType(node) && !l.errorNativeTarget(node) && !l.errorBuiltinPrototype(node) && l.errorSurfaceHazard(node) {
		return nil, true, l.notYet(node, "Error structural view may contain a non-Error object; compiler object-origin proof required")
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		receiver := node
		if node.Kind == ast.KindPropertyAccessExpression {
			receiver = node.AsPropertyAccessExpression().Expression
		}
		if node.Kind == ast.KindElementAccessExpression {
			receiver = node.AsElementAccessExpression().Expression
		}
		if node.Kind == ast.KindCallExpression {
			receiver = ast.SkipParentheses(node.AsCallExpression().Expression)
			if receiver.Kind == ast.KindPropertyAccessExpression {
				receiver = receiver.AsPropertyAccessExpression().Expression
			}
		}
		if l.errorObjectType(receiver) {
			return nil, true, l.notYet(node, "Error optional access requires compiler short-circuit lowering")
		}
	}
	number := func(n int) ir.Expression { return ir.NumberConstant{Value: float64(n)} }
	call := func(method string, of ir.Type, args ...ir.Expression) ir.Expression {
		return ir.ObjectCall{Method: method, Returns: of, Arguments: args}
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		receiver := ast.SkipParentheses(access.Expression)
		if l.errorObjectType(receiver) {
			key := ast.SkipParentheses(access.ArgumentExpression)
			if key.Kind != ast.KindStringLiteral || key.Text() != "stack" {
				return nil, true, l.notYet(node, "Error bracket members require compiler own-property metadata lowering; use a direct property read")
			}
			if !l.errorOwnProof(receiver) {
				return nil, true, l.notYet(node, "Error stack from an unproven or host origin requires compiler Error metadata proof")
			}
			value, err := l.expression(receiver)
			return call("errorReadStack", ir.String, value), true, err
		}
		if receiver.Kind == ast.KindPropertyAccessExpression && receiver.Name().Text() == "errors" && l.errorObjectType(receiver.AsPropertyAccessExpression().Expression) {
			value, err := l.expression(receiver)
			if err != nil {
				return nil, true, err
			}
			index, err := l.expression(access.ArgumentExpression)
			if err != nil {
				return nil, true, err
			}
			if index.Type() != ir.Number {
				return nil, true, l.notYet(node, "AggregateError errors indexing requires a numeric index")
			}
			return ir.ArrayIndex{Array: value, Index: index, Element: ir.Union}, true, nil
		}
	}
	if node.Kind == ast.KindTypeOfExpression {
		if _, ok := l.errorConstructor(node.AsTypeOfExpression().Expression); ok {
			return ir.StringConstant{Index: l.constant("function")}, true, nil
		}
	}
	if node.Kind == ast.KindBinaryExpression {
		b := node.AsBinaryExpression()
		if b.OperatorToken.Kind == ast.KindInstanceOfKeyword {
			kind, ok := l.errorConstructor(b.Right)
			if !ok && l.isLibraryGlobal(b.Right, "Object") && l.errorObjectSyntax(b.Left) {
				kind, ok = 0, true
			}
			if ok {
				value, err := l.expression(b.Left)
				if err != nil {
					return nil, true, err
				}
				return call("errorInstanceOf", ir.Boolean, fit(value, ir.Union), number(kind)), true, err
			}
		}
		left := ast.SkipParentheses(b.Left)
		if left.Kind == ast.KindElementAccessExpression && l.errorObjectType(left.AsElementAccessExpression().Expression) {
			if b.OperatorToken.Kind == ast.KindEqualsToken {
				return nil, true, l.notYet(node, "Error bracket writes require compiler own-property metadata lowering")
			}
		}
		if left.Kind == ast.KindPropertyAccessExpression && l.errorObjectType(left.AsPropertyAccessExpression().Expression) {
			name := left.Name().Text()
			if b.OperatorToken.Kind != ast.KindEqualsToken {
				if _, assignment := compoundAssignments[b.OperatorToken.Kind]; assignment {
					return nil, true, l.notYet(node, "Error property compound updates need compiler evaluation-order lowering")
				}
				return nil, false, nil
			}
			if name == "cause" || name == "errors" {
				return nil, true, l.notYet(node, "Error cause/errors writes need compiler boxed-property lowering")
			}
			if name == "name" || name == "message" {
				if l.errorCaptureFrozenProgram() {
					return nil, true, l.notYet(node, "Error name/message writes in a program that freezes or seals objects require compiler catchable descriptor failures")
				}
				if !l.errorNativeTarget(left.AsPropertyAccessExpression().Expression) && l.errorHostImports() {
					return nil, true, l.notYet(node, "Error prototype mutation with host modules requires library host Error ancestry metadata")
				}
				receiver, err := l.expression(left.AsPropertyAccessExpression().Expression)
				if err != nil {
					return nil, true, err
				}
				value, err := l.expression(b.Right)
				if err != nil {
					return nil, true, err
				}
				if value.Type() != ir.String {
					return nil, true, l.notYet(node, "Error name/message writes require strings")
				}
				member := 0
				if name == "message" {
					member = 1
				}
				return call("errorSetMember", ir.String, receiver, number(member), value), true, nil
			}
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		name := node.Name().Text()
		if kind, ok := l.errorConstructor(access.Expression); ok {
			switch name {
			case "prototype":
				return call("errorPrototype", ir.Object, number(kind)), true, nil
			case "name":
				return ir.StringConstant{Index: l.constant(errorFamilies[kind])}, true, nil
			case "length":
				length := 1
				if kind == 8 {
					length = 2
				}
				return number(length), true, nil
			}
		}
		if name == "prototype" && l.isLibraryGlobal(access.Expression, "Object") {
			parent := node.Parent
			allowed := parent != nil && parent.Kind == ast.KindBinaryExpression && (parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || parent.AsBinaryExpression().OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken)
			allowed = allowed || parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "isPrototypeOf"
			if !allowed {
				return nil, true, l.notYet(node, "Object.prototype outside Error prototype identity/traversal needs compiler generic prototype lowering")
			}
			return call("errorPrototype", ir.Object, number(0)), true, nil
		}
		if l.errorObjectType(access.Expression) {
			value, err := l.expression(access.Expression)
			if err != nil {
				return nil, true, err
			}
			switch name {
			case "name", "message":
				member := 0
				if name == "message" {
					member = 1
				}
				return call("errorMember", ir.String, value, number(member)), true, nil
			case "cause":
				return call("errorCause", ir.Union, value), true, nil
			case "errors":
				return call("errorErrors", ir.Array, value), true, nil
			case "stack":
				if !l.errorOwnProof(access.Expression) {
					return nil, true, l.notYet(node, "Error stack from an unproven or host origin requires compiler Error metadata proof")
				}
				return call("errorReadStack", ir.String, value), true, nil
			}
		}
	}
	if node.Kind != ast.KindCallExpression {
		return nil, false, nil
	}
	c := node.AsCallExpression()
	callee := ast.SkipParentheses(c.Expression)
	arguments := nodesOf(c.Arguments)
	if _, ok := l.errorConstructor(callee); ok {
		value, err := l.errorConstruct(node, callee, arguments)
		return value, true, err
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	name := callee.Name().Text()
	if kind, ok := l.errorConstructor(access.Expression); ok && name == "hasOwnProperty" && len(arguments) == 1 {
		key := ast.SkipParentheses(arguments[0])
		if key.Kind != ast.KindStringLiteral {
			return nil, true, l.notYet(node, "Error constructor own-property probe requires a literal key")
		}
		text := key.Text()
		own := text == "name" || text == "length" || text == "prototype" || kind == 1 && (text == "captureStackTrace" || text == "stackTraceLimit" || text == "prepareStackTrace" || text == "isError")
		return ir.BooleanConstant{Value: own}, true, nil
	}
	prototype := l.errorBuiltinPrototype(access.Expression)
	if ast.SkipParentheses(access.Expression).Kind == ast.KindPropertyAccessExpression {
		receiver := ast.SkipParentheses(access.Expression)
		prototype = prototype || receiver.Name().Text() == "prototype" && l.isLibraryGlobal(receiver.AsPropertyAccessExpression().Expression, "Object")
	}
	if name == "isPrototypeOf" && prototype && len(arguments) == 1 {
		if !l.errorObjectSyntax(arguments[0]) {
			return nil, true, l.notYet(node, "Error prototype traversal needs a represented Error object")
		}
		receiver, err := l.expression(access.Expression)
		if err != nil {
			return nil, true, err
		}
		value, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		return call("errorIsPrototypeOf", ir.Boolean, receiver, fit(value, ir.Union)), true, nil
	}
	if l.isLibraryGlobal(access.Expression, "Object") {
		if name == "getPrototypeOf" && len(arguments) == 1 && l.errorObjectSyntax(arguments[0]) {
			if l.errorPrototypeDepth(arguments[0]) < 1 {
				return nil, true, l.notYet(node, "getPrototypeOf(null) requires compiler catchable TypeError lowering")
			}
			v, err := l.expression(arguments[0])
			return call("errorGetPrototype", ir.Union, fit(v, ir.Union)), true, err
		}
		if (name == "keys" || name == "getOwnPropertyNames") && len(arguments) == 1 && l.errorObjectType(arguments[0]) {
			if !l.errorOwnProof(arguments[0]) {
				return nil, true, l.notYet(node, "Error own reflection from an unproven or host origin requires compiler Error metadata proof")
			}
			v, err := l.expression(arguments[0])
			return call(name, ir.Array, fit(v, ir.Union)), true, err
		}
		if name == "hasOwn" && len(arguments) == 2 && l.errorObjectType(arguments[0]) {
			if !l.errorOwnProof(arguments[0]) {
				return nil, true, l.notYet(node, "Error own reflection from an unproven or host origin requires compiler Error metadata proof")
			}
			v, err := l.expression(arguments[0])
			if err != nil {
				return nil, true, err
			}
			key, err := l.expression(arguments[1])
			if key.Type() != ir.String {
				return nil, true, l.notYet(node, "Error own-property keys require strings")
			}
			return call("hasOwn", ir.Boolean, v, key), true, err
		}
	}
	if l.errorObjectType(access.Expression) {
		v, err := l.expression(access.Expression)
		if err != nil {
			return nil, true, err
		}
		if name == "toString" && len(arguments) == 0 {
			return call("errorToString", ir.String, v), true, nil
		}
		if name == "hasOwnProperty" && len(arguments) == 1 {
			if !l.errorOwnProof(access.Expression) {
				return nil, true, l.notYet(node, "Error own reflection from an unproven or host origin requires compiler Error metadata proof")
			}
			key, err := l.expression(arguments[0])
			if err != nil {
				return nil, true, err
			}
			if key.Type() != ir.String {
				return nil, true, l.notYet(node, "Error own-property keys require strings")
			}
			return call("hasOwn", ir.Boolean, v, key), true, nil
		}
		if name == "propertyIsEnumerable" && len(arguments) == 1 {
			if !l.errorOwnProof(access.Expression) {
				return nil, true, l.notYet(node, "Error own reflection from an unproven or host origin requires compiler Error metadata proof")
			}
			key, err := l.expression(arguments[0])
			if err != nil {
				return nil, true, err
			}
			if key.Type() != ir.String {
				return nil, true, l.notYet(node, "Error enumeration keys require strings")
			}
			return call("errorEnumerable", ir.Boolean, v, key), true, nil
		}
	}
	return nil, false, nil
}
