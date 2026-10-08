package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func dynamicObjectType(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsIntersection|checker.TypeFlagsUnion) != 0 {
		for _, part := range proven.Types() {
			if dynamicObjectType(part) {
				return true
			}
		}
	}
	return false
}

func (l *lowering) inProperty(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	key := ast.SkipParentheses(binary.Left)
	if key.Kind != ast.KindStringLiteral {
		return nil, l.notYet(key, "in with a key other than a string literal (dynamic key coercion)")
	}
	if strings.ContainsRune(key.Text(), 0) {
		return nil, l.notYet(key, "in with a NUL key (native shape names)")
	}
	if checker.IsTupleType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(binary.Right))) {
		return nil, l.notYet(binary.Right, "in on a tuple (array presence metadata)")
	}
	if key.Text() == "stack" || key.Text() == "cause" || key.Text() == "errno" || key.Text() == "syscall" || key.Text() == "path" {
		return nil, l.notYet(key, "in on an Error internal or optional property (host descriptor metadata)")
	}
	if reason := l.presenceHazard(key.Text()); reason != "" {
		return nil, l.notYet(node, "in when "+reason)
	}
	if l.classes[l.symbol(ast.SkipParentheses(binary.Right))] != nil {
		return nil, l.notYet(binary.Right, "in on a class constructor (static property presence descriptors)")
	}
	object, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	if object.Type() != ir.Object && object.Type() != ir.Array && object.Type() != ir.Union {
		return nil, l.notYet(binary.Right, "in on this runtime representation")
	}
	return ir.HasProperty{Object: fit(object, ir.Union), Name: key.Text()}, nil
}

func (l *lowering) dynamicProperty(node *ast.Node, object ir.Expression, name string) (ir.Expression, error) {
	if strings.ContainsRune(name, 0) {
		return nil, l.notYet(node, "a dynamic property name containing NUL (native shape names)")
	}
	if l.dynamicReadHazard(name) {
		return nil, l.notYet(node, "a dynamic getter or method read, or a nullable field (runtime dispatch, slot tags and narrowed rereads)")
	}
	// Prototype function values and dynamic array elements need callable and element metadata.
	switch name {
	case "constructor", "__proto__", "toString", "valueOf", "hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "toLocaleString", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__", "map", "filter", "push", "pop", "slice", "join", "entries", "values", "keys":
		return nil, l.notYet(node, "a dynamic prototype property value (intrinsic identity and ToPrimitive)")
	}
	value := ir.Expression(ir.DynamicProperty{Object: object, Name: name})
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	if of != ir.Union {
		value = ir.Narrow{Value: value, To: of}
	}
	return value, nil
}

// A dynamic read can be converted with String only when every program shape's field of that
// name is scalar. This is a conservative closed-program proof, not a promise from unknown.
func (l *lowering) dynamicScalarProperty(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 && symbol.Declarations[0].Kind == ast.KindVariableDeclaration {
			declaration := symbol.Declarations[0]
			if declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.AsVariableDeclaration().Initializer != nil {
				node = ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
			}
		}
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := node.Name().Text()
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	known := true
	var visit ast.Visitor
	visit = func(candidate *ast.Node) bool {
		if l.isJSONParse(candidate) {
			known = false
		}
		if candidate.Kind == ast.KindObjectLiteralExpression || candidate.Kind == ast.KindNewExpression || candidate.Kind == ast.KindArrayLiteralExpression {
			if field := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(candidate), name); field != nil {
				proven := l.checker.GetTypeOfSymbol(field)
				if !l.writable(proven) {
					known = false
				}
			}
		}
		candidate.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return known
}

// Unknown views must preserve runtime tags. Opaque host objects and tuples have additional
// JavaScript properties their native layouts do not yet describe.
func (l *lowering) unknownView(node *ast.Node, own, contextual *checker.Type) error {
	if !dynamicObjectType(contextual) {
		return nil
	}
	// Library intrinsics lower their arguments with their own schema or operation. Their
	// declaration's unknown parameter does not create an unknown runtime view.
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindCallExpression {
		callee := outer.Parent.AsCallExpression().Expression
		if l.librarySymbol(l.memberSymbol(callee)) {
			return nil
		}
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "stringify" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") {
			return nil
		}
	}
	present := l.checker.GetNonNullableType(own)
	if held, known := l.representation(present); known && held == ir.Closure {
		return l.notYet(node, "a function viewed as unknown or object (dynamic function descriptors)")
	}
	for _, name := range []string{"JSON", "Math", "Reflect"} {
		if l.isLibraryGlobal(ast.SkipParentheses(node), name) {
			return l.notYet(node, "an intrinsic identity viewed as unknown or object (dynamic object descriptors)")
		}
	}
	for _, field := range l.checker.GetPropertiesOfType(present) {
		if l.includesNull(l.checker.GetTypeOfSymbol(field)) {
			return l.notYet(node, "an object with a nullable field viewed as unknown or object (null and undefined slot tags)")
		}
	}
	if checker.IsTupleType(present) {
		return l.notYet(node, "a tuple viewed as unknown or object (array presence metadata)")
	}
	if l.isLibraryType(present, "Map", "ReadonlyMap", "Set", "ReadonlySet", "RegExp", "Date", "Stats", "Hash", "Buffer") {
		return l.notYet(node, "an opaque host or collection value viewed as unknown or object (dynamic property metadata)")
	}
	return nil
}

// A key that could denote a getter or a prototype method cannot yet be read dynamically.
func (l *lowering) dynamicReadHazard(name string) bool {
	if l.accessorNames[name] {
		return true
	}
	for _, declaration := range l.classes {
		for _, member := range declaration.Members() {
			if member.Name() != nil && member.Name().Text() == name && member.Kind != ast.KindPropertyDeclaration {
				return true
			}
		}
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	nullable := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindNewExpression {
			if field := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(node), name); field != nil && l.includesNull(l.checker.GetTypeOfSymbol(field)) {
				nullable = true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return nullable
}

// Native absent-spread slots and ambient class fields are storage, not proof of JS presence.
func (l *lowering) presenceHazard(name string) string {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return "runtime shapes are not known"
	}
	reason := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				if property.Kind == ast.KindSpreadAssignment && l.includesUndefined(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {
					reason = "a possibly absent spread has no property presence descriptors"
				}
			}
		}
		if node.Kind == ast.KindClassDeclaration {
			for _, member := range node.Members() {
				if member.Kind == ast.KindMethodDeclaration && member.Name().Text() == name {
					for _, parameter := range member.Parameters() {
						if held, known := l.representation(l.checker.GetTypeAtLocation(parameter.Name())); known && slotless(held) {
							reason = "a method's presence is not in its native dispatch table"
						}
					}
					for _, signature := range l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(member), checker.SignatureKindCall) {
						if held, known := l.representation(l.checker.GetReturnTypeOfSignature(signature)); known && slotless(held) {
							reason = "a method's presence is not in its native dispatch table"
						}
					}
				}
				if member.Name() != nil && member.Name().Text() == name && ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient) {
					reason = "an ambient class field has no JavaScript own property"
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return reason
}
