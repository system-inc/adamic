package lower

import (
	"slices"
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
	// "key" in value narrows a union of object types to the members that declare the key, and any
	// object carrying that key passes, so the narrowed member's fields would be read on trust. The
	// sound use is presence on an unknown or single object type, where reads stay dynamic.
	if right := l.checker.GetNonNullableType(
		l.checker.GetTypeAtLocation(binary.Right),
	); right.Flags()&checker.TypeFlagsUnion != 0 {
		objects := 0
		for _, member := range right.Types() {
			if member.Flags()&checker.TypeFlagsObject != 0 {
				objects++
			}
		}
		if objects > 1 {
			return nil, &Refused{
				Where: l.program.Where(node),
				What:  "in that narrows a union of object types (any object with the key passes)",
				Fix:   "use a discriminant, or a Map",
			}
		}
	}
	if checker.IsTupleType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(binary.Right))) {
		return nil, l.notYet(binary.Right, "in on a tuple (array presence metadata)")
	}
	if key.Text() == "stack" ||
		key.Text() == "cause" ||
		key.Text() == "errno" ||
		key.Text() == "syscall" ||
		key.Text() == "path" {
		return nil, l.notYet(key, "in on an Error internal or optional property (host descriptor metadata)")
	}
	if reason := l.presenceHazard(key.Text()); reason != "" {
		return nil, l.notYet(node, "in when "+reason)
	}
	if l.classes[l.symbol(ast.SkipParentheses(binary.Right))] != nil {
		return nil, l.notYet(binary.Right, "in on a class constructor (static property presence descriptors)")
	}
	object, expressionError := l.expression(binary.Right)
	if expressionError != nil {
		return nil, expressionError
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
		return nil, l.notYet(
			node,
			"a dynamic getter or method read, or a nullable field (runtime dispatch, slot tags and narrowed rereads)",
		)
	}
	// Prototype function values and dynamic array elements need callable and element metadata.
	switch name {
	case "constructor", "__proto__", "toString", "valueOf", "hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "toLocaleString", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__", "map", "filter", "push", "pop", "slice", "join", "entries", "values", "keys":
		return nil, l.notYet(node, "a dynamic prototype property value (intrinsic identity and ToPrimitive)")
	}
	value := ir.Expression(ir.DynamicProperty{Object: object, Name: name})
	of, typeOfError := l.typeOf(node)
	if typeOfError != nil {
		return nil, typeOfError
	}
	if of != ir.Union {
		value = ir.Narrow{Value: value, To: of}
	}
	l.graphDynamicSite = node
	return value, nil
}

// A dynamic read can be converted with String only when every program shape's field of that
// name is scalar. This is a conservative closed-program proof, not a promise from unknown.
func (l *lowering) dynamicScalarProperty(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil &&
			len(symbol.Declarations) == 1 &&
			symbol.Declarations[0].Kind == ast.KindVariableDeclaration {
			declaration := symbol.Declarations[0]
			if declaration.Parent.Flags&ast.NodeFlagsConst != 0 &&
				declaration.AsVariableDeclaration().Initializer != nil {
				node = ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
			}
		}
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := node.Name().Text()
	modules, moduleOrderError := l.moduleOrder(l.program.Files()[0])
	if moduleOrderError != nil {
		return false
	}
	known := true
	var visit ast.Visitor
	visit = func(candidate *ast.Node) bool {
		if candidate.Kind == ast.KindObjectLiteralExpression ||
			candidate.Kind == ast.KindNewExpression ||
			candidate.Kind == ast.KindArrayLiteralExpression {
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
		if callee.Kind == ast.KindPropertyAccessExpression &&
			callee.Name().Text() == "stringify" &&
			l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") {
			return nil
		}
	}
	if l.includesNull(own) && own.Flags()&checker.TypeFlagsNull == 0 {
		if held, known := l.representation(own); known && held != ir.Union {
			return l.notYet(node, "a nullable reference viewed as unknown or object (null and undefined slot tags)")
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
			return l.notYet(
				node,
				"an object with a nullable field viewed as unknown or object (null and undefined slot tags)",
			)
		}
	}
	if l.reaches(present, l.nullableReferenceArray, map[*checker.Type]bool{}) {
		return l.notYet(
			node,
			"an array with a nullable element viewed as unknown or object (null and undefined element tags)",
		)
	}
	if checker.IsTupleType(present) {
		return l.notYet(node, "a tuple viewed as unknown or object (array presence metadata)")
	}
	// A tuple an unknown holds is read by a run-time key as an array (#g187f1e), which natively it
	// isn't: where an element is optional or rest, its shape can't say its length on Node.
	if l.reaches(present, variableTuple, map[*checker.Type]bool{}) {
		return l.notYet(node, "a tuple with optional or rest elements viewed as unknown or object (tuple length)")
	}
	if l.isLibraryType(
		present,
		"Map",
		"ReadonlyMap",
		"Set",
		"ReadonlySet",
		"RegExp",
		"Date",
		"Stats",
		"Hash",
		"Buffer",
	) {
		return l.notYet(
			node,
			"an opaque host or collection value viewed as unknown or object (dynamic property metadata)",
		)
	}
	return nil
}

// reaches says whether a value of this type is, or reaches through its fields, its index signatures'
// values, its elements or a union's members, an object type found says yes to. The walk follows the
// reach, not only the value's own fields, because an unknown hands out what it holds one read at a
// time: an object's field or a record's value holding such a value passes the nullable-field test
// above.
func (l *lowering) reaches(proven *checker.Type, found func(*checker.Type) bool, seen map[*checker.Type]bool) bool {
	proven = l.concrete(proven)
	if seen[proven] {
		return false
	}
	seen[proven] = true
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.reaches(member, found, seen) {
				return true
			}
		}
		return false
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if found(proven) {
		return true
	}
	if l.isArrayType(proven) {
		return l.reaches(l.elementTypeOfArray(proven), found, seen)
	}
	for _, field := range l.checker.GetPropertiesOfType(proven) {
		if l.reaches(l.checker.GetTypeOfSymbol(field), found, seen) {
			return true
		}
	}
	// A record's values, a string or number index signature's, are handed out by a key read as a
	// field's are (Language's z1).
	for _, info := range l.checker.GetIndexInfosOfType(proven) {
		if l.reaches(info.ValueType(), found, seen) {
			return true
		}
	}
	return false
}

// nullableReferenceArray says whether this is an array whose elements may be null and are held as
// references (a Point | null, a number[] | null, a null). Such an array holds null as a null pointer,
// which a run-time key read (union.c's keyed_array_element) can't tell from undefined, so it would
// read a null element as undefined (#ys4naww, Language's k07, k12 and k02). A string | null or a
// Point | null | undefined is boxed as a Union, whose null is its own, and reads right. A tuple, read
// as an array too (#g187f1e), holds such an element's null as a null pointer in its slot, so its
// elements are held to the same test (union.c's keyed_tuple_slot).
func (l *lowering) nullableReferenceArray(proven *checker.Type) bool {
	if checker.IsTupleType(proven) {
		return slices.ContainsFunc(l.checker.GetTypeArguments(proven), l.nullableReference)
	}
	return l.isArrayType(proven) && l.nullableReference(l.elementTypeOfArray(proven))
}

// nullableReference says whether a value of this type may be null and isn't held as a Union, whose
// null is its own. A representation not known says yes too, so the guard doesn't lean on valueType's
// NotYet for an element it can't place.
func (l *lowering) nullableReference(element *checker.Type) bool {
	if !l.includesNull(element) {
		return false
	}
	held, known := l.representation(element)
	return !known || held != ir.Union
}

// variableTuple says whether this is a tuple with an optional or rest element, whose length on Node
// natively its shape can't say: a left-out optional element is a slot holding undefined (tupleLiteral),
// as one written undefined is, so [1] and [1, undefined] of [number, string?] are alike natively, where
// Node's lengths are 1 and 2. The static read refuses that length too (tupleLength).
func variableTuple(proven *checker.Type) bool {
	return checker.IsTupleType(proven) &&
		checker.TupleType_combinedFlags(proven.TargetTupleType())&checker.ElementFlagsNonRequired != 0
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
	modules, moduleOrderError := l.moduleOrder(l.program.Files()[0])
	if moduleOrderError != nil {
		return true
	}
	nullable := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindNewExpression {
			if field := l.checker.GetPropertyOfType(
				l.checker.GetTypeAtLocation(node),
				name,
			); field != nil && l.includesNull(l.checker.GetTypeOfSymbol(field)) {
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
	modules, moduleOrderError := l.moduleOrder(l.program.Files()[0])
	if moduleOrderError != nil {
		return "runtime shapes are not known"
	}
	reason := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				if property.Kind == ast.KindSpreadAssignment &&
					l.includesUndefined(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {
					reason = "a possibly absent spread has no property presence descriptors"
				}
			}
		}
		if node.Kind == ast.KindClassDeclaration {
			for _, member := range node.Members() {
				if member.Kind == ast.KindMethodDeclaration && member.Name().Text() == name {
					for _, parameter := range member.Parameters() {
						if held, known := l.representation(
							l.checker.GetTypeAtLocation(parameter.Name()),
						); known && slotless(held) {
							reason = "a method's presence is not in its native dispatch table"
						}
					}
					for _, signature := range l.checker.GetSignaturesOfType(
						l.checker.GetTypeAtLocation(member),
						checker.SignatureKindCall,
					) {
						if held, known := l.representation(
							l.checker.GetReturnTypeOfSignature(signature),
						); known && slotless(held) {
							reason = "a method's presence is not in its native dispatch table"
						}
					}
				}
				if member.Name() != nil &&
					member.Name().Text() == name &&
					ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient) {
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
