package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// This unit consumes the existing exact-shape proof; it does not broaden that proof.
func (l *lowering) objectDescriptorCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	if name == "keys" || name == "getOwnPropertyNames" {
		args := node.AsCallExpression().Arguments.Nodes
		if len(args) == 1 && !hasSpread(node) && (l.descriptorResult(args[0], 0) || l.descriptorMapResult(args[0], 0)) {
			value, err := l.expression(args[0])
			if err != nil {
				return nil, true, err
			}
			return ir.ObjectCall{Method: name, Arguments: []ir.Expression{fit(value, ir.Union)}, Returns: ir.Array}, true, nil
		}
		return nil, false, nil
	}
	switch name {
	case "getOwnPropertyDescriptor", "getOwnPropertyDescriptors", "defineProperty", "defineProperties":
	default:
		return nil, false, nil
	}
	refuse := func(reason string) (ir.Expression, bool, error) {
		return nil, true, &Refused{Where: l.program.Where(node), What: "Object." + name + " property descriptors", Fix: reason}
	}
	if l.descriptorMutation() {
		return refuse("writing descriptor results requires tagged descriptor stores; returned snapshots are read-only in this subset")
	}
	args := node.AsCallExpression().Arguments.Nodes
	count := 2
	if name == "getOwnPropertyDescriptors" {
		count = 1
	}
	if name == "defineProperty" {
		count = 3
	}
	if len(args) != count || hasSpread(node) {
		return refuse("requires the declared arguments without spreads")
	}
	if !l.exactObject(args[0], 0) || l.enumObject(args[0]) != nil {
		return refuse("the complete plain data-property shape is not proven; structural views, accessors and changing shapes need the compiler's shape proof")
	}
	target, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	if target.Type() != ir.Object || checker.IsTupleType(l.checker.GetTypeAtLocation(args[0])) {
		return refuse("requires a present plain object, not an array, tuple or exotic")
	}
	call := ir.ObjectCall{Method: name, Arguments: []ir.Expression{target}, Returns: ir.Object}
	for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(args[0])) {
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || of == ir.Weak || of == ir.MaybeBoolean || of == ir.Union || field.Flags&ast.SymbolFlagsOptional != 0 {
			return refuse("every actual field must have a proven, present slot representation")
		}
		call.DescriptorFields = append(call.DescriptorFields, ir.ObjectDescriptorField{Name: field.Name, Of: of})
	}
	if name == "getOwnPropertyDescriptors" {
		if l.descriptorCapturedStack() {
			return refuse("captured stack properties require a complete runtime shape proof for descriptor enumeration")
		}
		return call, true, nil
	}
	if name == "getOwnPropertyDescriptor" || name == "defineProperty" {
		key, err := l.expression(args[1])
		if err != nil {
			return nil, true, err
		}
		if key.Type() != ir.String {
			return refuse("requires a string property key; coercion and Symbols need their own representation")
		}
		call.Arguments = append(call.Arguments, key)
	}
	if name == "getOwnPropertyDescriptor" {
		if l.descriptorCapturedStack() {
			key := ast.SkipParentheses(args[1])
			if key.Kind != ast.KindStringLiteral || key.Text() == "stack" {
				return refuse("Node's captured stack is an accessor descriptor, which the plain data-property runtime does not represent")
			}
		}
		return call, true, nil
	}
	descriptor := args[count-1]
	if name == "defineProperty" {
		key := ast.SkipParentheses(args[1])
		if key.Kind != ast.KindStringLiteral {
			return refuse("a definition requires a literal key naming an existing field")
		}
		if reason := l.proveDataDescriptor(args[0], key.Text(), descriptor); reason != "" {
			return refuse(reason)
		}
	} else {
		literal := l.descriptorLiteral(descriptor, 0)
		if literal == nil {
			return refuse("the descriptor map must be a complete plain literal or const binding")
		}
		for _, field := range literal.AsObjectLiteralExpression().Properties.Nodes {
			if field.Kind != ast.KindPropertyAssignment {
				return refuse("the descriptor map cannot contain spreads, accessors or methods")
			}
			if reason := l.proveDataDescriptor(args[0], field.Name().Text(), field.AsPropertyAssignment().Initializer); reason != "" {
				return refuse(reason)
			}
		}
	}
	value, err := l.expression(descriptor)
	if err != nil {
		return nil, true, err
	}
	call.Arguments = append(call.Arguments, value)
	return call, true, nil
}

func (l *lowering) descriptorLiteral(node *ast.Node, depth int) *ast.Node {
	if depth > 16 {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression && l.exactObject(node, 0) {
		return node
	}
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || declaration.AsVariableDeclaration().Type != nil {
		return nil
	}
	if declaration.AsVariableDeclaration().Initializer == nil {
		return nil
	}
	return l.descriptorLiteral(declaration.AsVariableDeclaration().Initializer, depth+1)
}
func (l *lowering) proveDataDescriptor(target *ast.Node, key string, node *ast.Node) string {
	field := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(target), key)
	if field == nil || strings.HasPrefix(key, "#") || key == "stack" {
		return "adding a property would change the fixed shape; only existing public data fields are proven"
	}
	literal := l.descriptorLiteral(node, 0)
	if literal == nil {
		return "the data descriptor must be a complete plain literal or const binding"
	}
	for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			return "descriptor accessors, methods, spreads and shorthand are not proven"
		}
		value := property.AsPropertyAssignment().Initializer
		switch property.Name().Text() {
		case "get", "set":
			return "accessor descriptors change plain loads and stores and remain refused"
		case "enumerable", "configurable", "writable":
			if ast.SkipParentheses(node).Kind != ast.KindObjectLiteralExpression {
				return "attributes on an aliased descriptor are not proven immutable; write the literal descriptor at the definition"
			}
			if ast.SkipParentheses(value).Kind != ast.KindTrueKeyword {
				return "individual attribute changes need per-property runtime metadata; only omitted attributes or literal true are represented"
			}
		case "value":
			from, to := l.checker.GetTypeAtLocation(value), l.checker.GetTypeOfSymbol(field)
			a, known := l.representation(from)
			b, ok := l.representation(to)
			if !known || !ok || a != b || (b != ir.Number && b != ir.Boolean && b != ir.String) || !l.checker.IsTypeAssignableTo(from, to) || !l.enumAssignable(from, to) {
				return "the descriptor value must preserve the field's proven number, boolean or string type"
			}
		default:
			return "only plain value, writable, enumerable and configurable descriptor fields are supported"
		}
	}
	return ""
}

func (l *lowering) objectDescriptorPrototypeCall(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	return l.objectDescriptorPrototypeCallArguments(node, receiver, name, node.AsCallExpression().Arguments.Nodes)
}

func (l *lowering) objectDescriptorPrototypeCallArguments(node, receiver *ast.Node, name string, args []*ast.Node) (ir.Expression, bool, error) {
	if name != "propertyIsEnumerable" {
		return nil, false, nil
	}
	of, _ := l.representation(l.checker.GetTypeAtLocation(receiver))
	if of != ir.Object {
		return nil, false, nil
	}
	if !l.exactObject(receiver, 0) && !l.descriptorResult(receiver, 0) && !l.descriptorMapResult(receiver, 0) {
		// Class instances and other established receivers retain the existing
		// prototype-hazard proof; this handler only adds exact data shapes.
		return nil, false, nil
	}
	if len(args) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "propertyIsEnumerable with these arguments")
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String {
		return nil, true, l.notYet(node, "propertyIsEnumerable with a non-string key")
	}
	return ir.ObjectCall{Method: name, Arguments: []ir.Expression{object, key}, Returns: ir.Boolean}, true, nil
}

// Descriptor member reads are specialized only for results with proven provenance. The
// library's any-valued .value becomes a tagged unknown, never an unchecked scalar load.
func (l *lowering) objectDescriptorRead(node *ast.Node) (ir.Expression, bool, error) {
	access := node.AsPropertyAccessExpression()
	if origin := l.descriptorMapOrigin(access.Expression, 0); origin != nil {
		args := origin.AsCallExpression().Arguments.Nodes
		if len(args) != 1 || l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(args[0]), node.Name().Text()) == nil {
			return nil, true, &Refused{Where: l.program.Where(node), What: "a descriptor map read", Fix: "only keys in the proven actual shape can be read"}
		}
	}
	if !l.descriptorResult(access.Expression, 0) {
		if l.isLibraryType(l.checker.GetTypeAtLocation(access.Expression), "PropertyDescriptor", "TypedPropertyDescriptor") {
			return nil, true, l.notYet(node, "a descriptor read without proven snapshot provenance")
		}
		return nil, false, nil
	}
	name := node.Name().Text()
	method := "descriptorValue"
	of := ir.Union
	switch name {
	case "writable", "enumerable", "configurable":
		method = "descriptorFlag"
		of = ir.MaybeBoolean
	case "value":
		if held, known := l.representation(l.checker.GetTypeAtLocation(node)); known {
			of = held
		}
	case "get", "set":
		method = "descriptorAbsent"
		of = ir.Closure
	default:
		return nil, false, nil
	}
	value, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	if access.QuestionDotToken == nil {
		value = l.defined(access.Expression, value)
	}
	key := ir.StringConstant{Index: l.constant(name)}
	return ir.ObjectCall{Method: method, Arguments: []ir.Expression{value, key}, Returns: of}, true, nil
}
func (l *lowering) descriptorResult(node *ast.Node, depth int) bool {
	if depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		d := symbol.Declarations[0]
		if d.Kind != ast.KindVariableDeclaration || d.Parent.Flags&ast.NodeFlagsConst == 0 || d.AsVariableDeclaration().Type != nil {
			return false
		}
		n := d.AsVariableDeclaration().Initializer
		return n != nil && l.descriptorResult(n, depth+1)
	}
	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		return callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") && callee.Name().Text() == "getOwnPropertyDescriptor"
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		return l.descriptorMapResult(node.AsPropertyAccessExpression().Expression, depth+1)
	}
	if node.Kind == ast.KindElementAccessExpression {
		return l.descriptorMapResult(node.AsElementAccessExpression().Expression, depth+1)
	}

	return false
}

func (l *lowering) descriptorMapResult(node *ast.Node, depth int) bool {
	return l.descriptorMapOrigin(node, depth) != nil
}
func (l *lowering) descriptorMapOrigin(node *ast.Node, depth int) *ast.Node {
	if depth > 16 {
		return nil
	}
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return nil
		}
		d := symbol.Declarations[0]
		if d.Kind != ast.KindVariableDeclaration || d.Parent.Flags&ast.NodeFlagsConst == 0 || d.AsVariableDeclaration().Type != nil {
			return nil
		}
		n := d.AsVariableDeclaration().Initializer
		if n != nil {
			return l.descriptorMapOrigin(n, depth+1)
		}
		return nil
	}
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") && callee.Name().Text() == "getOwnPropertyDescriptors" {
		return node
	}
	return nil
}
func (l *lowering) objectDescriptorElement(node *ast.Node) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	origin := l.descriptorMapOrigin(access.Expression, 0)
	if origin == nil {
		return nil, false, nil
	}
	args := origin.AsCallExpression().Arguments.Nodes
	key := ast.SkipParentheses(access.ArgumentExpression)
	if len(args) != 1 || key.Kind != ast.KindStringLiteral || l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(args[0]), key.Text()) == nil {
		return nil, true, &Refused{Where: l.program.Where(node), What: "a descriptor map read", Fix: "only literal keys in the proven actual shape can be read"}
	}
	value, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	return ir.Property{Object: value, Name: key.Text(), Of: ir.Object}, true, nil
}

// captureStackTrace can add a name absent from the checker's shape. Do not guess an
// enumeration layout while the compiler's allocation-shape proof is being built.
func (l *lowering) descriptorCapturedStack() bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "captureStackTrace" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Error") {
			found = true
		}
		if found {
			return true
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}

// Descriptor snapshots use boxed slots, unlike ordinary scalar object fields.
// Never send a write through the generic plain-field store.
func (l *lowering) descriptorMutation() bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var target *ast.Node
		if node.Kind == ast.KindBinaryExpression {
			b := node.AsBinaryExpression()
			_, compound := compoundAssignments[b.OperatorToken.Kind]
			if b.OperatorToken.Kind == ast.KindEqualsToken || compound {
				target = ast.SkipParentheses(b.Left)
			}
		} else if node.Kind == ast.KindPrefixUnaryExpression {
			u := node.AsPrefixUnaryExpression()
			if u.Operator == ast.KindPlusPlusToken || u.Operator == ast.KindMinusMinusToken {
				target = ast.SkipParentheses(u.Operand)
			}
		} else if node.Kind == ast.KindPostfixUnaryExpression {
			target = ast.SkipParentheses(node.AsPostfixUnaryExpression().Operand)
		}
		var receiver *ast.Node
		if target != nil && target.Kind == ast.KindPropertyAccessExpression {
			receiver = target.AsPropertyAccessExpression().Expression
		}
		if target != nil && target.Kind == ast.KindElementAccessExpression {
			receiver = target.AsElementAccessExpression().Expression
		}
		if receiver != nil && (l.descriptorResult(receiver, 0) || l.descriptorMapResult(receiver, 0) || l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "PropertyDescriptor", "TypedPropertyDescriptor")) {
			found = true
		}
		if found {
			return true
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}
