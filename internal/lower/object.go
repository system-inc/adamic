package lower

import (
	"fmt"
	"math"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// objectLiteral lowers { name: value, ... }, and the one spread 0.1 allows: { ...source, fields },
// where every field replaces one the source's type already has.
func (l *lowering) objectLiteral(node *ast.Node) (ir.Expression, error) {
	if literal, handled, err := l.accessorLiteral(node); handled {
		return literal, err
	}
	literal := ir.ObjectLiteral{}
	for index, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		switch property.Kind {
		case ast.KindSpreadAssignment:
			if index != 0 {
				return nil, &Refused{Where: l.program.Where(property), What: "a spread after the first field", Fix: "spread once, first: { ...source, field: value } (adamic/single-spread)"}
			}
			spread, err := l.expression(property.AsSpreadAssignment().Expression)
			if err != nil {
				return nil, err
			}
			if spread.Type() != ir.Object {
				return nil, l.notYet(property, "spreading a "+typeName(spread.Type()))
			}
			literal.NoReuse = l.hasPrivateStorage(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) || l.hasAccessorStorage(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression))
			literal.Spread = spread
			literal.SpreadMaybeUndefined = l.includesUndefined(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression))
		case ast.KindMethodDeclaration:
			name, known := l.methodName(property)
			if !known {
				return nil, l.notYet(property, "an unsupported computed method")
			}
			value, err := l.objectMethod(property)
			if err != nil {
				return nil, err
			}
			literal.Fields = append(literal.Fields, ir.Field{Name: name, Value: value})
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			name := property.Name()
			if (ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral) && name.Text() == iteratorSlot {
				return nil, l.notYet(property, "a string field with the reserved iterator slot name")
			}
			fieldName, known := l.methodName(property)
			if !known {
				return nil, l.notYet(name, "a computed field name")
			}
			if property.Kind == ast.KindPropertyAssignment && fieldName == "__proto__" {
				return nil, &Refused{Where: l.program.Where(property), What: "__proto__ in an object literal", Fix: "JavaScript changes the prototype instead of making an own field; Adamic objects have fixed shapes and no prototype mutation"}
			}
			var value ir.Expression
			var err error
			if property.Kind == ast.KindPropertyAssignment {
				value, err = l.expression(property.AsPropertyAssignment().Initializer)
			} else {
				value, err = l.shorthand(property)
			}
			if err != nil {
				return nil, err
			}
			if literal.Spread != nil && !l.hasProperty(node.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression, fieldName) {
				return nil, l.notYet(property, "a spread that adds a field the source doesn't have")
			}
			if declared := l.declaredField(node, fieldName); declared != 0 && !slotless(declared) {
				// Store the value as the member's slot holds it, rather than the initializer's type.
				value = fit(value, declared)
			}
			if slotless(value.Type()) {
				return nil, l.notYet(property, "a field holding "+typeName(value.Type()))
			}
			literal.Fields = append(literal.Fields, ir.Field{Name: fieldName, Value: value})
		default:
			return nil, l.notYet(property, describe(property)+" in an object literal")
		}
	}
	if literal.SpreadMaybeUndefined {
		empty, err := l.emptySpread(node, literal.Fields)
		if err != nil {
			return nil, err
		}
		literal.Empty = empty
	}
	return literal, nil
}

// emptySpread is the object { ...source, fields } makes when source is undefined: JavaScript's is
// only the literal's own fields, and a read of any other field the source's type has is undefined,
// so each of those is there, undefined. A field that can hold undefined only as a pair (a boolean),
// or that keeps a handle (a Weak), says NotYet.
func (l *lowering) emptySpread(node *ast.Node, own []ir.Field) ([]ir.Field, error) {
	spread := node.AsObjectLiteralExpression().Properties.Nodes[0]
	given := map[string]bool{}
	for _, field := range own {
		given[field.Name] = true
	}
	empty := []ir.Field{}
	source := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(spread.AsSpreadAssignment().Expression))
	for _, property := range l.checker.GetPropertiesOfType(source) {
		if given[property.Name] {
			continue
		}
		fieldType, _ := l.representation(l.checker.GetTypeOfSymbol(property))
		switch fieldType {
		case ir.Number, ir.MaybeNumber:
			empty = append(empty, ir.Field{Name: property.Name, Value: ir.MaybeOf{Of: ir.MaybeNumber}})
		case ir.String, ir.Object, ir.Array, ir.Map, ir.Closure:
			empty = append(empty, ir.Field{Name: property.Name, Value: ir.Undefined{}})
		default:
			return nil, l.notYet(spread, "spreading a value that may be undefined, whose field "+property.Name+" can't be left undefined yet")
		}
	}
	return empty, nil
}

// typeOfSymbol is what's left at runtime of a symbol's declared type, or NotYet at node.
func (l *lowering) typeOfSymbol(node *ast.Node, symbol *ast.Symbol) (ir.Type, error) {
	declared := l.checker.GetTypeOfSymbol(symbol)
	if valueType, isKnown := l.representation(declared); isKnown {
		return valueType, nil
	}
	return 0, l.notYet(node, "a value of type "+l.checker.TypeToString(declared))
}

// declaredField is the representation of a field as the type an object literal is written into
// declares it, or 0 when there is no such type to say.
func (l *lowering) declaredField(literal *ast.Node, name string) ir.Type {
	contextual := l.checker.GetContextualType(literal, checker.ContextFlagsNone)
	if contextual == nil {
		return 0
	}
	if contextual.Flags()&checker.TypeFlagsUnion != 0 {
		// The union's combined property can hold unrelated types. The literal belongs to a
		// member, whose fields are what reads after discriminant narrowing expect.
		made := l.checker.GetTypeAtLocation(literal)
		var shared ir.Type
		for _, member := range contextual.Types() {
			if !l.checker.IsTypeAssignableTo(made, member) {
				continue
			}
			field := l.checker.GetPropertyOfType(member, name)
			if field == nil {
				continue
			}
			declared, _ := l.representation(l.checker.GetTypeOfSymbol(field))
			if shared != 0 && shared != declared {
				return 0
			}
			shared = declared
		}
		return shared
	}
	field := l.checker.GetPropertyOfType(contextual, name)
	if field == nil {
		return 0
	}
	declared, _ := l.representation(l.checker.GetTypeOfSymbol(field))
	return declared
}

// hasProperty reports whether a value's type has a field of that name.
func (l *lowering) hasProperty(node *ast.Node, name string) bool {
	// A source that may be undefined has its fields where it's there.
	for _, property := range l.checker.GetPropertiesOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node))) {
		if property.Name == name {
			return true
		}
	}
	return false
}

func (l *lowering) arrayLiteral(node *ast.Node) (ir.Expression, error) {
	if tuple := l.tupleType(node); tuple != nil {
		return l.tupleLiteral(node, tuple)
	}
	element, err := l.elementType(node)
	if err != nil {
		return nil, err
	}
	literal := ir.ArrayLiteral{Element: element}
	items := node.AsArrayLiteralExpression().Elements.Nodes
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil && l.checker.IsArrayType(contextual) {
		declared, known := l.representation(l.checker.GetElementTypeOfArrayType(contextual))
		if !known || declared != element {
			for _, item := range items {
				if item.Kind == ast.KindSpreadElement {
					if plan, err := l.planIteration(item.AsSpreadElement().Expression); err != nil {
						return nil, err
					} else if plan != nil {
						return nil, l.notYet(node, "a custom spread whose element representation differs from its destination")
					}
				}
			}
		}
	}

	if len(items) == 1 && items[0].Kind == ast.KindSpreadElement && !l.checker.IsArrayType(l.checker.GetTypeAtLocation(items[0].AsSpreadElement().Expression)) {
		// [...text] is the text's code points, [...set] its elements, [...map] its entries, and
		// [...map.keys()] and the rest what they give (collections.go).
		spread, given, err := l.iterated(items[0].AsSpreadElement().Expression)
		if err == nil && given != element {
			return nil, l.notYet(node, "a spread whose element representation differs from its destination")
		}
		return spread, err
	}
	spreads := false
	for _, item := range items {
		if item.Kind == ast.KindOmittedExpression {
			return nil, l.notYet(item, describe(item)+" in an array literal")
		}
		spread := item.Kind == ast.KindSpreadElement
		if spread {
			item = item.AsSpreadElement().Expression
		}
		var value ir.Expression
		if spread {
			value, _, err = l.iterated(item)
		} else {
			value, err = l.expression(item)
		}
		if err != nil {
			return nil, err
		}
		if spread {
			if value.Type() != ir.Array {
				return nil, l.notYet(item, "spreading a "+typeName(value.Type())+" among other elements")
			}
			if _, other, err := l.iteratedType(item); err != nil || other != element {
				return nil, l.notYet(item, "spreading an array of other elements")
			}
			spreads = true
		} else {
			// An element of number | undefined is a packed word: a number or undefined is made one.
			value = fit(value, element)
		}
		literal.Elements = append(literal.Elements, value)
		literal.Spread = append(literal.Spread, spread)
	}
	if !spreads {
		literal.Spread = nil
	}
	return literal, nil
}

// elementType is the representation of an array's elements, from the checker's type for the node.
func (l *lowering) elementType(node *ast.Node) (ir.Type, error) {
	arrayType := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node))
	if l.isLibraryType(arrayType, "RegExpExecArray", "RegExpMatchArray") {
		return ir.String, nil
	}
	if l.isLibraryType(arrayType, "RegExpIndicesArray") {
		return ir.Object, nil
	}
	if literal := ast.SkipParentheses(node); literal.Kind == ast.KindArrayLiteralExpression {
		// [] is never[] to the checker; what it will hold is the type it's written into, as in
		// const values: number[] = []. So is [node] written into a Weak<Node>[]: its elements are
		// kept weakly.
		if contextual := l.checker.GetContextualType(literal, checker.ContextFlagsNone); contextual != nil && l.checker.IsArrayType(contextual) {
			if declared, _ := l.representation(l.checker.GetElementTypeOfArrayType(contextual)); len(literal.AsArrayLiteralExpression().Elements.Nodes) == 0 || declared == ir.Weak {
				arrayType = contextual
			}
		}
	}
	if target := l.weakTarget(arrayType); target != nil {
		// A Weak<Node[]> narrowed to present is the array.
		arrayType = target
	}
	if !l.checker.IsArrayType(arrayType) {
		return 0, l.notYet(node, "a value of type "+l.checker.TypeToString(arrayType)+" where an array goes")
	}
	element := l.checker.GetElementTypeOfArrayType(arrayType)
	valueType, isKnown := l.kept(element)
	if !isKnown || slotless(valueType) {
		// An element is one adamic_value, and number | undefined needs two words.
		return 0, l.notYet(node, "an array of "+l.checker.TypeToString(element))
	}
	return valueType, nil
}

// property lowers object.name, array.length, and Math's constants.
func (l *lowering) property(node *ast.Node) (ir.Expression, error) {
	if err := l.staticProperty(node); err != nil {
		return nil, err
	}
	if call, handled, err := l.superAccessor(node, nil); handled {
		return call, err
	}
	access := node.AsPropertyAccessExpression()
	name := l.fieldName(node.Name())
	if _, iterator := l.libraryIteratorElement(access.Expression); iterator && name != "next" {
		return nil, l.notYet(node, "a collection iterator property other than next")
	}
	if access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		// The rest of a chain after a ?., which short-circuits with it.
		return nil, l.notYet(node, "an optional chain longer than one step")
	}
	if value, known := l.libraryMathNumberProperty(node); known {
		return value, nil
	}
	if l.isLibraryGlobal(access.Expression, "Math") {
		switch name {
		case "PI":
			return ir.NumberConstant{Value: 3.141592653589793}, nil
		case "E":
			return ir.NumberConstant{Value: 2.718281828459045}, nil
		case "random":
			return nil, refusedRandom(l, node)
		}
		return nil, l.notYet(node, "Math."+name)
	}
	if l.isLibraryGlobal(access.Expression, "Number") {
		if value, isConstant := numberConstants[name]; isConstant && name != "Infinity" {
			return ir.NumberConstant{Value: value}, nil
		}
		return nil, l.notYet(node, "Number."+name)
	}
	// A library declaration proves a prototype member exists, never an own slot. Keep this
	// guard in lowering too, even when the up-front unbound-method pass has already refused it.
	if l.inheritedLibraryMember(node) && !l.regexRuntimeProperty(access.Expression, name) && name != "length" && name != "size" && !(l.isLibraryType(l.checker.GetTypeAtLocation(access.Expression), "Error") && (name == "name" || name == "message")) {
		return nil, l.prototypeRead(node, name)
	}
	if err := l.erasedLiteralMethod(node); err != nil {
		return nil, err
	}
	if read := l.checker.GetSymbolAtLocation(node.Name()); read != nil && len(read.Declarations) > 0 && read.Declarations[0].Kind == ast.KindMethodDeclaration && !isCallee(node) {
		// A method read off its object, not called: JavaScript loses its this (unbound-method,
		// docs/0.1.md). A plain call never comes here, since callOrMethod lowers it; object?.method()
		// reads it as its call's callee, through the object's methods, and keeps its this.
		object := "object"
		if receiver := ast.SkipParentheses(access.Expression); ast.IsIdentifier(receiver) {
			object = receiver.Text()
		} else if receiver.Kind == ast.KindThisKeyword {
			object = "this"
		}
		return nil, &Refused{
			Where: l.program.Where(node),
			What:  "a method read off its object, which loses its this when called (unbound-method)",
			Fix:   fmt.Sprintf("wrap the call in an arrow function, which keeps its object: (value) => %s.%s(value)", object, name),
		}
	}
	if receiver := l.checker.GetTypeAtLocation(access.Expression); name != "length" && (checker.IsTupleType(receiver) || checker.IsTupleType(l.checker.GetNonNullableType(receiver))) {
		// A tuple is held as an object of its elements, "0", "1", ..., read by index; an array's
		// methods aren't fields of it, and reading them as fields would find nothing. Its length is
		// read below (tupleLength).
		return nil, l.notYet(node, "."+name+" on a tuple")
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() == ir.Object && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), "RegExp") {
		switch name {
		case "lastIndex", "source", "flags", "global", "ignoreCase", "multiline", "unicode", "sticky", "hasIndices", "unicodeSets", "dotAll":
		default:
			return nil, l.notYet(node, "a prototype property on a RegExp")
		}
	}
	if object.Type() == ir.Object && l.regexGroups(access.Expression) {
		of, err := l.typeOf(node)
		if err != nil {
			return nil, err
		}
		if of != ir.String && of != ir.Object {
			return nil, l.notYet(node, "a named-group key with an Object-prototype type")
		}
		return ir.RegExpGroup{Object: object, Name: name, Of: of, Optional: access.QuestionDotToken != nil}, nil
	}
	if object.Type() == ir.Object && name == "done" {
		proven := l.checker.GetTypeAtLocation(access.Expression)
		members := []*checker.Type{proven}
		if proven.Flags()&checker.TypeFlagsUnion != 0 {
			members = proven.Types()
		}
		iterator := true
		for _, member := range members {
			iterator = iterator && l.isLibraryType(member, "IteratorYieldResult", "IteratorReturnResult")
		}
		if iterator {
			of, e := l.typeOf(node)
			return ir.RegExpCall{Value: object, Method: "iteratorDone", Returns: of}, e
		}
	}
	object = l.privateStaticReceiver(node.Name(), object, false)
	if access.QuestionDotToken != nil && (object.Type() == ir.Array || object.Type() == ir.String) && name == "length" {
		// words?.length and text?.length: undefined where the array or string is, a number | undefined.
		if object.Type() == ir.Array {
			return ir.Length{Array: object, Optional: true}, nil
		}
		return ir.StringLength{Value: object, Optional: true}, nil
	}
	if object.Type() == ir.Array && name != "length" && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray") {
		of, e := l.typeOf(node)
		if e != nil {
			return nil, e
		}
		switch name {
		case "index", "input", "groups", "indices":
		default:
			return nil, l.notYet(node, "a prototype property on a RegExp result")
		}
		stored := of
		if name == "index" {
			if l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), "RegExpMatchArray") {
				stored = ir.MaybeNumber
			} else {
				stored = ir.Number
			}
		}
		result := ir.Expression(ir.RegExpProperty{Array: object, Name: name, Of: stored, Optional: access.QuestionDotToken != nil})
		if stored == ir.MaybeNumber && of == ir.Number {
			result = ir.Unwrap{Value: result}
		}
		return result, nil
	}
	if access.QuestionDotToken != nil && object.Type() != ir.Object {
		return nil, l.notYet(node, "optional chaining to ."+name+" on a "+typeName(object.Type()))
	}
	switch {
	case object.Type() == ir.Array && name == "length":
		return ir.Length{Array: object}, nil
	case object.Type() == ir.String && name == "length":
		return ir.StringLength{Value: object}, nil
	case object.Type() == ir.Map && name == "size":
		return ir.MapSize{Map: object}, nil
	case object.Type() == ir.Object && name == "length" && checker.IsTupleType(l.checker.GetTypeAtLocation(access.Expression)):
		// A tuple is an object of fields "0", "1" and on, with no length to read; a fixed one's length
		// is its type's, read here when nothing about reading the tuple itself could be seen.
		return l.tupleLength(node, access.Expression)
	case object.Type() == ir.Object:
		if field := l.checker.GetSymbolAtLocation(node.Name()); field != nil {
			if declared, _ := l.representation(l.checker.GetTypeOfSymbol(field)); declared == ir.Weak {
				// The field keeps a handle, whatever the checker narrowed the read to.
				return l.readObjectField(node, ir.Property{Object: object, Name: name, Of: ir.Weak, Optional: access.QuestionDotToken != nil}), nil
			}
		}
		if field := l.checker.GetSymbolAtLocation(node.Name()); field != nil && (accessorSymbol(field) || (l.accessorNames[node.Name().Text()] && !isClassInstance(l.checker.GetTypeAtLocation(access.Expression)) && ast.SkipParentheses(access.Expression).Kind != ast.KindThisKeyword)) {
			declared := l.checker.GetTypeOfSymbol(field)
			observed := l.checker.GetTypeAtLocation(node)
			if !l.classAssignable(declared, observed) {
				return nil, &Refused{Where: l.program.Where(node), What: "a narrowed accessor reread, which can return a different value", Fix: "read the getter into a local once, then narrow and use that local"}
			}
		}
		of, err := l.typeOf(node)
		if err != nil && l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
			// Narrowed to undefined (just assigned it): read as the field is declared.
			if field := l.checker.GetSymbolAtLocation(node.Name()); field != nil {
				of, err = l.typeOfSymbol(node, field)
			}
		}
		if err != nil {
			return nil, err
		}
		optional := access.QuestionDotToken != nil
		if field := l.checker.GetSymbolAtLocation(node.Name()); field != nil {
			if stored, known := l.representation(l.checker.GetTypeOfSymbol(field)); known && stored.IsMaybe() && of == stored.Present() {
				// Read the declared representation before trusting the narrowing. A call or an
				// alias write may have restored undefined, just as for a narrowed variable.
				if slotless(stored) {
					return nil, l.notYet(node, "a narrowed boolean | undefined field; copy the field into a local and narrow that local instead")
				}
				read := l.readObjectField(node, ir.Property{Object: object, Name: name, Of: stored, Optional: optional, Class: l.classOf(node)})
				if l.acceptsUndefined(node) {
					return read, nil
				}
				return ir.Unwrap{Value: read}, nil
			}
		}
		if of.IsMaybe() && optional {
			// box?.size is number | undefined because box may be; the field itself is what's stored.
			if field := l.checker.GetSymbolAtLocation(node.Name()); field != nil {
				if stored, isKnown := l.representation(l.checker.GetTypeOfSymbol(field)); isKnown && stored == of.Present() {
					return l.readObjectField(node, ir.Property{Object: object, Name: name, Of: stored, Optional: true, Class: l.classOf(node)}), nil
				}
			}
		}
		if slotless(of) {
			return nil, l.notYet(node, "a field of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node)))
		}
		if of == ir.MaybeNumber {
			// number | undefined, whether the field holds it or ?. makes it: the packed word, or
			// undefined when the object is.
			return l.readObjectField(node, ir.Property{Object: object, Name: name, Of: ir.MaybeNumber, Optional: optional, Class: l.classOf(node)}), nil
		}
		if optional && !of.IsReference() {
			return nil, l.notYet(node, "?. to a "+typeName(of)+", which would be "+typeName(of)+" | undefined")
		}
		return l.defined(node, l.readObjectField(node, ir.Property{Object: object, Name: name, Of: of, Optional: optional, Class: l.classOf(node)})), nil
	}
	return nil, l.notYet(node, "."+name+" on a "+typeName(object.Type()))
}

// readObjectField keeps an optional own field distinct from optional chaining of its receiver.
// Absence is a read result, never a synthetic own field: hasOwnProperty and object spread still see
// the shape that was actually made. A narrowed number checks the declared optional representation.
func (l *lowering) readObjectField(node *ast.Node, property ir.Property) ir.Expression {
	field := l.checker.GetSymbolAtLocation(node.Name())
	if field == nil || field.Flags&ast.SymbolFlagsOptional == 0 {
		return property
	}
	property.Absent = true
	if declared, _ := l.representation(l.checker.GetTypeOfSymbol(field)); declared == ir.MaybeNumber && property.Of == ir.Number {
		property.Of = ir.MaybeNumber
		return fit(property, ir.Number)
	}
	return property
}

// hasOwnProperty lowers object.hasOwnProperty(key) when the method is the library's, not a field the
// object declares. The checker puts Object's prototype methods on every object's type. The shape
// holds only the object's own fields, and reading hasOwnProperty as one of them panics: the field
// the checker proved is not there.
func (l *lowering) hasOwnProperty(node *ast.Node, receiver *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	symbol := l.checker.GetSymbolAtLocation(callee)
	if symbol == nil || len(symbol.Declarations) == 0 || !load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return nil, false, nil
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, "hasOwnProperty with other than one argument")
	}
	if keyType, _ := l.representation(l.checker.GetTypeAtLocation(arguments[0])); keyType != ir.String {
		return nil, true, l.notYet(node, "hasOwnProperty with a key that isn't a string")
	}
	// JavaScript's order: the object, then the key.
	object, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	return ir.HasOwn{Object: object, Key: key}, true, nil
}

// isLibraryGlobal reports whether a node names one of the library's globals, rather than something
// of the program's that happens to share its name.
func (l *lowering) isLibraryGlobal(node *ast.Node, name string) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) || node.Text() != name {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	return symbol != nil && len(symbol.Declarations) > 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// mathFunctions are the Math functions 0.1 has (docs/0.1.md), with how many arguments each takes;
// -1 is any number.
var mathFunctions = map[string]int{
	"abs": 1, "ceil": 1, "floor": 1, "round": 1, "trunc": 1, "sign": 1, "sqrt": 1,
	"pow": 2, "max": -1, "min": -1,
	// V8's fdlibm, ported bit for bit (runtime/ieee754.c), and its hypot (runtime/hypot.c).
	"sin": 1, "cos": 1, "tan": 1, "asin": 1, "acos": 1, "atan": 1, "atan2": 2,
	"sinh": 1, "cosh": 1, "tanh": 1, "asinh": 1, "acosh": 1, "atanh": 1,
	"exp": 1, "expm1": 1, "log": 1, "log1p": 1, "log2": 1, "log10": 1, "cbrt": 1, "hypot": -1,
}

// refusedRandom refuses Math.random for good in 0.1: a program's output would no longer be a function
// of its source, and the oracle compares it with Node's byte for byte (docs/0.1.md).
func refusedRandom(l *lowering, node *ast.Node) error {
	return &Refused{Where: l.program.Where(node), What: "Math.random", Fix: "0.1 programs are deterministic, so the oracle can hold them to Node; compute the values you need, with a generator of your own seeded by a constant"}
}

// builtin lowers a call to Math or a number's toFixed. isBuiltin is false for any other call.
func (l *lowering) builtin(node *ast.Node) (ir.Expression, bool, error) {
	if value, handled, err := l.userMethodCall(node); handled {
		return value, true, err
	}
	if value, known, err := l.libraryMathNumberCall(node); known {
		return value, true, err
	}
	if lowered, isString, err := l.libraryString(node); isString {
		return lowered, true, err
	}
	if value, matched, err := l.regexBuiltin(node); matched {
		return value, true, err
	}
	if value, handled, err := l.objectKeys(node); handled {
		return value, true, err
	}
	if lowered, isInput, err := l.input(node); isInput {
		return lowered, true, err
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	// The global parseInt and parseFloat are Number's, the same functions.
	if l.isLibraryGlobal(callee, "parseInt") || l.isLibraryGlobal(callee, "parseFloat") {
		return l.numberCall(node, callee.Text())
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	if l.isLibraryGlobal(receiver, "Object") {
		return l.objectCall(node, name)
	}
	if lowered, handled, err := l.objectPrototypeCall(node, receiver, name); handled {
		return lowered, true, err
	}
	if l.isLibraryGlobal(receiver, "JSON") {
		return l.jsonCall(node, name)
	}
	if l.isLibraryGlobal(receiver, "Number") {
		return l.numberCall(node, name)
	}
	if l.isLibraryGlobal(receiver, "Map") && name == "groupBy" {
		return l.libraryMapGroupBy(node)
	}
	if l.isLibraryGlobal(receiver, "Array") && name == "from" {
		return l.arrayFrom(node)
	}
	if l.isLibraryGlobal(receiver, "String") && (name == "fromCharCode" || name == "fromCodePoint") {
		return l.stringFromCodes(node, name == "fromCodePoint")
	}
	if receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); receiverType == ir.Number && name == "toString" && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		value, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		return ir.NumberToString{Value: value}, true, nil
	}
	if receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); receiverType == ir.Number && (name == "toExponential" || name == "toPrecision" || name == "toString") {
		return l.numberFormat(node, receiver, name)
	}
	isMath := l.isLibraryGlobal(receiver, "Math")
	receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver))
	isToFixed := name == "toFixed" && receiverType == ir.Number
	if receiverType == ir.Array && name == "pop" && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		element, err := l.elementType(receiver)
		if err != nil {
			return nil, true, err
		}
		array, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		return ir.ArrayPop{Array: array, Element: element}, true, nil
	}
	if receiverType == ir.Array && libraryArrayMethods[name] {
		return l.libraryArrayMethod(node, receiver, name)
	}
	if _, isVisit := visits[name]; receiverType == ir.Array && (isVisit || arrayMethods[name]) {
		return l.arrayMethod(node, receiver, name)
	}
	if receiverType == ir.Map && (name == "keys" || name == "values" || name == "entries") {
		return l.libraryCollectionIterator(node, receiver, name)
	}
	if receiverType == ir.Map && l.isSet(receiver) {
		return l.setMethod(node, receiver, name)
	}
	if receiverType == ir.Map && (name == "clear" || name == "forEach") {
		return l.clearOrVisit(node, receiver, name, false)
	}
	if receiverType == ir.Map && (name == "get" || name == "set" || name == "has" || name == "delete") {
		return l.mapMethod(node, receiver, name)
	}
	if receiverType == ir.String && (name == "trim" || name == "charCodeAt") {
		return l.stringMethod(node, receiver, name)
	}
	if _, isKnown := stringMethods[name]; isKnown && receiverType == ir.String {
		return l.stringCall(node, receiver, name)
	}
	if receiverType == ir.Object && checker.IsTupleType(l.checker.GetTypeAtLocation(receiver)) {
		// A tuple is held as an object: called as an object, an array method would be read as a field.
		return nil, true, l.notYet(node, name+" on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made)")
	}
	if !isMath && !isToFixed {
		return nil, false, nil
	}
	// JavaScript's order: the receiver, then the arguments.
	var value ir.Expression
	if isToFixed {
		var err error
		if value, err = l.expression(receiver); err != nil {
			return nil, true, err
		}
	}
	if isMath && (name == "max" || name == "min" || name == "hypot") && hasSpread(node) {
		spread, err := l.spreadNumbers(node)
		if err != nil {
			return nil, true, err
		}
		return ir.MathCall{Function: name, Spread: spread}, true, nil
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, "a spread argument to "+name)
		}
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if lowered.Type() != ir.Number {
			return nil, true, l.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	if isMath {
		count, isKnown := mathFunctions[name]
		if name == "random" {
			return nil, true, refusedRandom(l, node)
		}
		if !isKnown {
			return nil, true, l.notYet(node, "Math."+name)
		}
		if count >= 0 && len(arguments) != count {
			return nil, true, l.notYet(node, "Math."+name+" with other than its arguments")
		}
		return ir.MathCall{Function: name, Arguments: arguments}, true, nil
	}
	if len(arguments) > 1 {
		return nil, true, l.notYet(node, "toFixed with more than one argument")
	}
	digits := ir.Expression(ir.NumberConstant{Value: 0})
	if len(arguments) == 1 {
		digits = arguments[0]
	}
	return ir.ToFixed{Value: value, Digits: digits}, true, nil
}

// numberFormat lowers value.toExponential(digits), value.toPrecision(digits) and
// value.toString(radix), the argument perhaps left out: the receiver first, then the argument, as
// JavaScript evaluates them.
func (l *lowering) numberFormat(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) > 1 {
		return nil, true, l.notYet(node, name+" with more than one argument")
	}
	format := ir.NumberFormat{Method: name, Value: value}
	if len(arguments) == 1 {
		argument, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		if argument.Type() != ir.Number {
			return nil, true, l.notYet(arguments[0], "a "+typeName(argument.Type())+" argument to "+name)
		}
		format.Argument = argument
	}
	return format, true, nil
}

// numberFunctions are the functions of Number that 0.1 has (docs/0.1.md), with what each takes; the
// last of parseInt's, the radix, may be left out.
var numberFunctions = map[string][]ir.Type{
	"parseInt": {ir.String, ir.Number}, "parseFloat": {ir.String},
	"isNaN": {ir.Number}, "isFinite": {ir.Number}, "isInteger": {ir.Number}, "isSafeInteger": {ir.Number},
}

// numberCall lowers Number.parseInt, parseFloat, isNaN, isFinite, isInteger and isSafeInteger.
func (l *lowering) numberCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	takes, isKnown := numberFunctions[name]
	if !isKnown {
		return nil, true, l.notYet(node, "Number."+name)
	}
	written := node.AsCallExpression().Arguments.Nodes
	optional := 0
	if name == "parseInt" {
		optional = 1
	}
	if len(written) > len(takes) || len(written) < len(takes)-optional {
		return nil, true, l.notYet(node, name+" with these arguments")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if lowered.Type() != takes[index] {
			// Number.isNaN('x') is false in JavaScript, and a string there is a mistake 0.1 won't guess at.
			return nil, true, l.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	return ir.NumberCall{Function: name, Arguments: arguments}, true, nil
}

// numberConstants are Number's constants, and the globals NaN and Infinity.
var numberConstants = map[string]float64{
	"MAX_SAFE_INTEGER": 9007199254740991, "MIN_SAFE_INTEGER": -9007199254740991,
	"EPSILON": 2.220446049250313e-16, "MAX_VALUE": math.MaxFloat64, "MIN_VALUE": math.SmallestNonzeroFloat64,
	"POSITIVE_INFINITY": math.Inf(1), "NEGATIVE_INFINITY": math.Inf(-1), "NaN": math.NaN(), "Infinity": math.Inf(1),
}

// forOf lowers for (const element of array).
func (l *lowering) forOf(node *ast.Node) ([]ir.Statement, error) {
	if statements, known, err := l.libraryArrayForOf(node); known {
		return statements, err
	}
	statement := node.AsForInOrOfStatement()
	if statement.AwaitModifier != nil {
		return nil, l.notYet(node, "for await")
	}
	initializer := statement.Initializer
	if initializer.Kind != ast.KindVariableDeclarationList || initializer.Flags&ast.NodeFlagsBlockScoped == 0 {
		return nil, l.notYet(initializer, "a for...of that doesn't declare its variable with const or let")
	}
	declarations := initializer.AsVariableDeclarationList().Declarations.Nodes
	if len(declarations) != 1 {
		return nil, l.notYet(initializer, "a for...of declaring more than one variable")
	}
	name := declarations[0].Name()
	if !ast.IsIdentifier(name) && name.Kind != ast.KindArrayBindingPattern {
		return nil, l.notYet(initializer, "a for...of destructuring an object")
	}
	if plan, err := l.planIteration(statement.Expression); err != nil {
		return nil, err
	} else if plan != nil {
		return l.forOfUser(node, plan, name)
	}
	// map.entries(), map.keys() and map.values() are the map itself, iterated for that part.
	iterated, mapPart := statement.Expression, ""
	if call := ast.SkipParentheses(iterated); call.Kind == ast.KindCallExpression {
		if err := l.optionalCall(call); err != nil {
			return nil, err
		}
	}
	if call := ast.SkipParentheses(iterated); call.Kind == ast.KindCallExpression && len(call.AsCallExpression().Arguments.Nodes) == 0 {
		if callee := ast.SkipParentheses(call.AsCallExpression().Expression); callee.Kind == ast.KindPropertyAccessExpression {
			receiver := callee.AsPropertyAccessExpression().Expression
			if receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); receiverType == ir.Map {
				switch part := callee.Name().Text(); part {
				case "entries", "keys", "values":
					iterated, mapPart = receiver, part
				}
			}
		}
	}
	iterable, err := l.expression(iterated)
	if err != nil {
		return nil, err
	}
	if element, iterator := l.libraryIteratorElement(iterated); iterator {
		return l.libraryForOfIterator(node, iterable, element, name)
	}
	var element ir.Type
	switch iterable.Type() {
	case ir.Array:
		if element, err = l.elementType(statement.Expression); err != nil {
			return nil, err
		}
	case ir.String:
		// A string's elements are its code points, each a string.
		element = ir.String
	case ir.Map:
		if ast.IsIdentifier(name) && (mapPart == "entries" || (mapPart == "" && !l.isSet(iterated))) {
			var key, value ir.Type
			if l.isSet(iterated) {
				key, err = l.setElement(iterated)
				value = key
			} else {
				key, value, err = l.mapTypes(iterated)
			}
			if err != nil {
				return nil, err
			}
			iterator := ir.CollectionIterator{Collection: iterable, Part: "entries", Key: key, Value: value, Set: l.isSet(iterated)}
			return l.libraryForOfIterator(node, iterator, ir.Object, name)
		}
		return l.forOfMap(node, iterable, iterated, mapPart, name)
	case ir.Object:
		if !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(iterated)), "RegExpStringIterator") {
			return nil, l.notYet(iterated, "for...of over an object")
		}
		element = ir.Array
	default:
		return nil, l.notYet(statement.Expression, "for...of over a "+typeName(iterable.Type()))
	}
	lowered := ir.ForOf{Iterable: iterable, Element: element, RegexIterator: iterable.Type() == ir.Object}
	if ast.IsIdentifier(name) {
		if lowered.Local, err = l.declareLocal(name); err != nil {
			return nil, err
		}
		if element == ir.Weak {
			// An array of Weak<Node> narrowed to present (by filter) still holds handles, so the
			// variable holds one too, and each read of it is the target.
			l.result.Locals[lowered.Local].Type = ir.Weak
		}
	} else {
		// for (const [a, b] of pairs): each name reads a field of the tuple, "0", "1", ...
		if element != ir.Object {
			return nil, l.notYet(name, "destructuring a "+typeName(element))
		}
		for index, binding := range name.AsBindingPattern().Elements.Nodes {
			if skipped(binding) {
				continue
			}
			bound := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || bound.Initializer != nil || bound.DotDotDotToken != nil {
				return nil, l.notYet(binding, "a destructured name that isn't plain")
			}
			local, err := l.declareLocal(binding.Name())
			if err != nil {
				return nil, err
			}
			if of := l.result.Locals[local].Type; slotless(of) {
				return nil, l.notYet(binding, "a tuple element of type "+typeName(of))
			}
			lowered.Pattern = append(lowered.Pattern, ir.Binding{Local: local, Field: strconv.Itoa(index)})
		}
	}
	if lowered.Body, err = l.statement(statement.Statement); err != nil {
		return nil, err
	}
	return []ir.Statement{lowered}, nil
}

// forOfMap lowers for...of over a map: for (const [key, value] of map), and over map.keys() or
// map.values() into one name. An entry as one name would be a tuple made each step, which stage 0
// doesn't do yet.
func (l *lowering) forOfMap(node *ast.Node, iterable ir.Expression, iterated *ast.Node, part string, name *ast.Node) ([]ir.Statement, error) {
	if l.isSet(iterated) {
		return l.forOfSet(node, iterable, iterated, part, name)
	}
	key, value, err := l.mapTypes(iterated)
	if err != nil {
		return nil, err
	}
	if part == "" {
		part = "entries"
	}
	lowered := ir.ForOf{Iterable: iterable, MapPart: part, Key: key, Value: value}
	switch {
	case part == "entries" && name.Kind == ast.KindArrayBindingPattern:
		elements := name.AsBindingPattern().Elements.Nodes
		if len(elements) > 2 {
			return nil, l.notYet(name, "destructuring more than a key and a value")
		}
		for index, binding := range elements {
			if skipped(binding) {
				continue
			}
			bound := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || bound.Initializer != nil || bound.DotDotDotToken != nil {
				return nil, l.notYet(binding, "a destructured name that isn't plain")
			}
			local, err := l.declareLocal(binding.Name())
			if err != nil {
				return nil, err
			}
			lowered.Pattern = append(lowered.Pattern, ir.Binding{Local: local, Field: strconv.Itoa(index)})
		}
	case part != "entries" && ast.IsIdentifier(name):
		if lowered.Local, err = l.declareLocal(name); err != nil {
			return nil, err
		}
	default:
		return nil, l.notYet(name, "for...of over a map's "+part+" into this name (write [key, value], or iterate keys() or values())")
	}
	if lowered.Body, err = l.statement(node.AsForInOrOfStatement().Statement); err != nil {
		return nil, err
	}
	return []ir.Statement{lowered}, nil
}

// switchStatement lowers switch, whose cases 0.1 requires to be constants.
func (l *lowering) switchStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsSwitchStatement()
	if err := l.enumSwitch(node); err != nil {
		return nil, err
	}
	value, err := l.expression(statement.Expression)
	if err != nil {
		return nil, err
	}
	lowered := ir.Switch{Value: value}
	groups := []switchGroup{}
	prefix := []ir.Statement{}
	identity := l.enumIdentity(l.checker.GetTypeAtLocation(statement.Expression))
	checkedDefault := l.numericEnum(identity) && l.enumSwitchCovered(node) && !l.enumDefaultUnreachable(node)
	var neverCheck ir.Statement
	if checkedDefault {
		// The unmatched edge is never to the checker, including an implicit default. Hold the
		// scrutinee once so the check names the original value without repeating its effects.
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "enum_switch_value", Type: value.Type(), Function: l.functionIndex})
		prefix = append(prefix, ir.Declare{Local: local, Value: value})
		value = ir.Read{Local: local, Of: value.Type()}
		lowered.Value = value
		neverCheck = ir.Evaluate{Value: l.enumNeverCheck(statement.Expression, value, identity)}
		lowered.Default = []ir.Statement{neverCheck}
	}
	tests := []ir.Expression{}
	defaultPending := false
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		for _, inner := range clause.AsCaseOrDefaultClause().Statements.Nodes {
			if inner.Kind == ast.KindVariableStatement {
				// A declaration directly in a case is scoped to the whole switch in JavaScript, where
				// another case can see it (and hit its dead zone).
				return nil, l.notYet(inner, "a declaration directly in a case (wrap the case in a block)")
			}
		}
		isDefault := clause.Kind == ast.KindDefaultClause
		if isDefault && len(tests) == 0 && l.enumDefaultUnreachable(node) && (len(groups) == 0 || switchBodyLeaves(groups[len(groups)-1].body)) {
			continue
		}
		if !isDefault {
			test, err := l.expression(clause.AsCaseOrDefaultClause().Expression)
			if err != nil {
				return nil, err
			}
			switch test.(type) {
			case ir.NumberConstant, ir.StringConstant, ir.BooleanConstant:
			default:
				if l.enumMember(clause.AsCaseOrDefaultClause().Expression) == nil {
					return nil, l.notYet(clause, "a case that isn't a constant")
				}
			}
			if test.Type() != value.Type() {
				return nil, l.notYet(clause, "a case whose type differs from the switch's")
			}
			tests = append(tests, test)
		}
		body, err := l.statements(clause.AsCaseOrDefaultClause().Statements.Nodes)
		if err != nil {
			return nil, err
		}
		defaultPending = defaultPending || isDefault
		if len(body) == 0 {
			// Empty labels enter the next body, including labels on either side of default.
			continue
		}
		groups = append(groups, switchGroup{tests: tests, body: body, isDefault: defaultPending})
		if defaultPending {
			lowered.Default = body
		}
		if len(tests) > 0 {
			// Default is a fallback position, but its grouped tests still compete in source order.
			lowered.Cases = append(lowered.Cases, ir.Case{Tests: tests, Body: body})
		}
		tests = []ir.Expression{}
		defaultPending = false
	}
	if len(tests) > 0 {
		// A trailing empty label matches and leaves the switch without running default.
		lowered.Cases = append(lowered.Cases, ir.Case{Tests: tests})
	}
	if len(tests) > 0 || defaultPending {
		groups = append(groups, switchGroup{tests: tests, isDefault: defaultPending})
	}
	if checkedDefault {
		return append(prefix, l.fallthroughSwitch(value, groups, neverCheck)...), nil
	}
	for index, group := range groups {
		// A default with tests must also have one body, rather than sharing statement
		// addresses between two branches (flow instrumentation identifies those addresses).
		if (group.isDefault && len(group.tests) > 0) || (index+1 < len(groups) && len(group.body) > 0 && !switchBodyLeaves(group.body)) {
			return append(prefix, l.fallthroughSwitch(value, groups, nil)...), nil
		}
	}
	return append(prefix, lowered), nil
}

// arrayMethod lowers array.push(value) and array.join(separator).
func (l *lowering) arrayMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	if created := ast.SkipParentheses(receiver); name == "fill" && created.Kind == ast.KindNewExpression && l.isLibraryGlobal(created.AsNewExpression().Expression, "Array") {
		return l.newArrayFilled(node, created)
	}
	element, err := l.elementType(receiver)
	if err != nil {
		return nil, true, err
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if name == "sort" {
		return l.arraySort(node, array, element)
	}
	if name == "map" {
		arguments := node.AsCallExpression().Arguments.Nodes
		if len(arguments) != 1 {
			return nil, true, l.notYet(node, "map with other than one callback")
		}
		callback, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		if callback.Type() != ir.Closure {
			return nil, true, l.notYet(arguments[0], "map with a callback that isn't a function")
		}
		result, err := l.elementType(node)
		if err != nil {
			return nil, true, err
		}
		return ir.ArrayMap{Array: array, Callback: callback, Element: element, Result: result}, true, nil
	}
	if _, isVisit := visits[name]; isVisit {
		return l.arrayVisit(node, array, element, name)
	}
	if name == "reduce" {
		return l.arrayReduce(node, array, element)
	}
	if name == "concat" && element == ir.Array {
		// TypeScript says grid.concat(row) appends row as one element; JavaScript spreads any array it's
		// given, so the types and the program would disagree.
		return nil, true, l.notYet(node, "concat on an array of arrays")
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, lowered)
	}
	if name == "slice" {
		for _, argument := range arguments {
			if argument.Type() != ir.Number {
				return nil, true, l.notYet(node, "slice with an index that isn't a number")
			}
		}
		if len(arguments) > 2 {
			return nil, true, l.notYet(node, "slice with more than two arguments")
		}
		return ir.ArraySlice{Array: array, Arguments: arguments}, true, nil
	}
	if name == "push" {
		if len(arguments) != 1 {
			return nil, true, l.notYet(node, "push with other than one value")
		}
		return ir.ArrayPush{Array: array, Value: fit(arguments[0], element), Element: element, Site: l.writeSite(receiver)}, true, nil
	}
	switch name {
	case "includes", "indexOf":
		if len(arguments) != 1 {
			return nil, true, l.notYet(node, name+" with a starting index")
		}
		if arguments[0] = fit(arguments[0], element); arguments[0].Type() != element {
			return nil, true, l.notYet(node, name+" with a value of another type than the elements")
		}
		return ir.ArraySearch{Array: array, Value: arguments[0], Element: element, Includes: name == "includes"}, true, nil
	case "at":
		if len(arguments) != 1 || arguments[0].Type() != ir.Number {
			return nil, true, l.notYet(node, "at with other than one number")
		}
		return ir.ArrayIndex{Array: array, Index: arguments[0], Element: element, Relative: true}, true, nil
	case "reverse":
		return ir.ArrayReverse{Array: array}, true, nil
	case "fill":
		if len(arguments) > 0 {
			arguments[0] = fit(arguments[0], element)
		}
		if len(arguments) == 0 || len(arguments) > 3 || arguments[0].Type() != element {
			return nil, true, l.notYet(node, "fill with other than a value of the elements' type")
		}
		fill := ir.ArrayFill{Array: array, Value: arguments[0], Element: element, Site: l.writeSite(receiver)}
		for index, bound := range arguments[1:] {
			if bound.Type() != ir.Number {
				return nil, true, l.notYet(node, "fill with a bound that isn't a number")
			}
			if index == 0 {
				fill.Start = bound
			} else {
				fill.End = bound
			}
		}
		return fill, true, nil
	case "splice":
		if len(arguments) == 0 || arguments[0].Type() != ir.Number || (len(arguments) > 1 && arguments[1].Type() != ir.Number) {
			return nil, true, l.notYet(node, "splice without a start and a count that are numbers")
		}
		splice := ir.ArraySplice{Array: array, Start: arguments[0], Element: element, Site: l.writeSite(receiver)}
		if len(arguments) > 1 {
			splice.Count = arguments[1]
			for _, item := range arguments[2:] {
				if item = fit(item, element); item.Type() != element {
					return nil, true, l.notYet(node, "splice inserting a value of another type than the elements")
				}
				splice.Items = append(splice.Items, item)
			}
		}
		return splice, true, nil
	case "concat":
		for index, argument := range arguments {
			if argument.Type() != ir.Array {
				return nil, true, l.notYet(node, "concat with a value that isn't an array (JavaScript appends it)")
			}
			if other, err := l.elementType(node.AsCallExpression().Arguments.Nodes[index]); err != nil || other != element {
				return nil, true, l.notYet(node, "concat of arrays of different elements")
			}
		}
		return ir.ArrayConcat{Array: array, Others: arguments}, true, nil
	}
	if element != ir.Number && element != ir.Boolean && element != ir.String && element != ir.MaybeNumber {
		// JavaScript writes an object as "[object Object]", a function as its source, and an array as
		// its own join, flattened; 0.1 has no use for any of that.
		return nil, true, l.notYet(node, "join on an array of objects, arrays, maps or functions")
	}
	separator := ir.Expression(ir.StringConstant{Index: l.constant(",")})
	if len(arguments) == 1 {
		arguments[0] = l.orDefault(node.AsCallExpression().Arguments.Nodes[0], arguments[0], ",")
		if arguments[0].Type() != ir.String {
			return nil, true, l.notYet(node, "join with a separator that isn't a string")
		}
		separator = arguments[0]
	} else if len(arguments) > 1 {
		return nil, true, l.notYet(node, "join with more than one argument")
	}
	return ir.ArrayJoin{Array: array, Separator: separator, Element: element}, true, nil
}

// arrayMethods are the array methods arrayMethod lowers, beside the visits.
var arrayMethods = map[string]bool{
	"push": true, "join": true, "slice": true, "sort": true, "map": true, "reduce": true,
	"includes": true, "indexOf": true, "at": true, "reverse": true, "concat": true, "splice": true,
	"fill": true,
}

// newArrayFilled lowers new Array(length).fill(value). new Array(length) alone is an array of holes,
// which 0.1 can't hold, so it lowers only filled, and filled whole.
func (l *lowering) newArrayFilled(node *ast.Node, created *ast.Node) (ir.Expression, bool, error) {
	element, err := l.elementType(node)
	if err != nil {
		return nil, true, err
	}
	constructed := created.AsNewExpression().Arguments
	filled := node.AsCallExpression().Arguments.Nodes
	if constructed == nil || len(constructed.Nodes) != 1 || len(filled) != 1 {
		return nil, true, l.notYet(node, "new Array filled with other than one length and one value (a part left unfilled is a hole)")
	}
	length, err := l.expression(constructed.Nodes[0])
	if err != nil {
		return nil, true, err
	}
	value, err := l.expression(filled[0])
	if err != nil {
		return nil, true, err
	}
	if value = fit(value, element); length.Type() != ir.Number || value.Type() != element {
		return nil, true, l.notYet(node, "new Array(length).fill(value) with a length that isn't a number or a value of another type")
	}
	return ir.ArrayFill{Length: length, Value: value, Element: element}, true, nil
}

// visits are the array methods that call a function per element and look at what it returns.
var visits = map[string]struct{}{"forEach": {}, "filter": {}, "some": {}, "every": {}, "find": {}, "findIndex": {}}

// arrayVisit lowers forEach, filter, some, every, find and findIndex. All but forEach decide by what
// the callback returns, which 0.1 requires to be a boolean: JavaScript would take any value's
// truthiness there, and 0.1 has none.
func (l *lowering) arrayVisit(node *ast.Node, array ir.Expression, element ir.Type, name string) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, name+" with other than one callback")
	}
	callback, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(arguments[0], name+" with a callback that isn't a function")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(arguments[0]), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(arguments[0], name+" with an overloaded callback")
	}
	var returns ir.Type
	if result := l.checker.GetReturnTypeOfSignature(signatures[0]); result.Flags()&checker.TypeFlagsVoid == 0 {
		var isKnown bool
		if returns, isKnown = l.representation(result); !isKnown {
			return nil, true, l.notYet(arguments[0], name+" with a callback returning "+l.checker.TypeToString(result))
		}
	}
	if name != "forEach" && returns != ir.Boolean {
		return nil, true, &Refused{Where: l.program.Where(arguments[0]), What: "a " + name + " callback that doesn't return a boolean", Fix: "return a comparison, like word.length > 0: 0.1 has no truthiness"}
	}
	return ir.ArrayVisit{Method: name, Array: array, Callback: callback, Element: element, Returns: returns}, true, nil
}

// arrayReduce lowers array.reduce(callback, initial). 0.1 requires the initial value (docs/0.1.md):
// without one, an empty array throws and a one-element array returns it without a call.
func (l *lowering) arrayReduce(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) == 1 {
		return nil, true, &Refused{Where: l.program.Where(node), What: "reduce without an initial value", Fix: "pass one, like reduce((sum, value) => sum + value, 0): without it, an empty array throws"}
	}
	if len(arguments) != 2 {
		return nil, true, l.notYet(node, "reduce with other than a callback and an initial value")
	}
	callback, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(arguments[0], "reduce with a callback that isn't a function")
	}
	initial, err := l.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	initial = fit(initial, result)
	if result != initial.Type() || slotless(result) {
		return nil, true, l.notYet(node, "reduce to a "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node)))
	}
	return ir.ArrayReduce{Array: array, Callback: callback, Initial: initial, Element: element, Result: result}, true, nil
}

// mapTypes is a Map's key and value representations. 0.1's maps have string or number keys.
func (l *lowering) mapTypes(node *ast.Node) (ir.Type, ir.Type, error) {
	arguments := l.typeArguments(l.checker.GetTypeAtLocation(node))
	if len(arguments) != 2 {
		return 0, 0, l.notYet(node, "a Map whose key and value types aren't known")
	}
	key, keyKnown := l.representation(arguments[0])
	value, valueKnown := l.kept(arguments[1])
	if !keyKnown || !keyable(key) {
		return 0, 0, l.notYet(node, "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions")
	}
	// number | undefined is held in a value's one slot packed (native/slots.go).
	if !valueKnown || slotless(value) {
		return 0, 0, l.notYet(node, "a Map of "+l.checker.TypeToString(arguments[1]))
	}
	return key, value, nil
}

// newExpression lowers new Map(), and new Map([[key, value], ...]) with its pairs written out, which
// is what the array of pairs means.
func (l *lowering) newExpression(node *ast.Node) (ir.Expression, error) {
	if l.isLibraryGlobal(node.AsNewExpression().Expression, "RegExp") {
		return l.regexConstant(node)
	}
	created := node.AsNewExpression()
	if declaration, isClass := l.classes[l.symbol(ast.SkipParentheses(created.Expression))]; isClass {
		return l.construct(node, declaration)
	}
	if l.isLibraryGlobal(created.Expression, "Set") {
		return l.newSet(node)
	}
	if l.isLibraryGlobal(created.Expression, "Error") {
		return l.newError(node)
	}
	if !l.isLibraryGlobal(created.Expression, "Map") {
		return nil, l.notYet(node, "new "+describe(created.Expression))
	}
	key, value, err := l.mapTypes(node)
	if err != nil {
		return nil, err
	}
	lowered := ir.MapNew{Key: key, Value: value}
	if created.Arguments == nil || len(created.Arguments.Nodes) == 0 {
		return lowered, nil
	}
	if len(created.Arguments.Nodes) != 1 {
		return nil, l.notYet(node, "new Map with more than one argument")
	}
	pairs := ast.SkipParentheses(created.Arguments.Nodes[0])
	if l.libraryEmptyCollectionArgument(pairs) {
		return lowered, nil
	}
	if !writtenOut(pairs) {
		// Pairs from anywhere else: an array of them, another Map, its entries() (collections.go).
		return l.newMapFrom(node, pairs, key, value)
	}
	for _, pair := range pairs.AsArrayLiteralExpression().Elements.Nodes {
		pair = ast.SkipParentheses(pair)
		if pair.Kind != ast.KindArrayLiteralExpression || len(pair.AsArrayLiteralExpression().Elements.Nodes) != 2 {
			return nil, l.notYet(pair, "a Map entry that isn't [key, value] written out")
		}
		elements := pair.AsArrayLiteralExpression().Elements.Nodes
		entryKey, err := l.expression(elements[0])
		if err != nil {
			return nil, err
		}
		entryValue, err := l.expression(elements[1])
		if err != nil {
			return nil, err
		}
		// A number, or undefined, where number | undefined goes is made that pair.
		entryKey = fit(entryKey, key)
		entryValue = fit(entryValue, value)
		if entryKey.Type() != key || entryValue.Type() != value {
			return nil, l.notYet(pair, "a Map entry whose key or value is of another type than the Map's")
		}
		lowered.Entries = append(lowered.Entries, [2]ir.Expression{entryKey, entryValue})
	}
	return lowered, nil
}

// writtenOut reports whether a new Map's argument is its pairs written out: an array literal of
// [key, value] literals, each set in order as written.
func writtenOut(pairs *ast.Node) bool {
	if pairs.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	for _, pair := range pairs.AsArrayLiteralExpression().Elements.Nodes {
		pair = ast.SkipParentheses(pair)
		if pair.Kind != ast.KindArrayLiteralExpression || len(pair.AsArrayLiteralExpression().Elements.Nodes) != 2 {
			return false
		}
		for _, element := range pair.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
				return false
			}
		}
	}
	return true
}

// mapMethod lowers map.get, set, has and delete.
func (l *lowering) mapMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	key, value, err := l.mapTypes(receiver)
	if err != nil {
		return nil, true, err
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, lowered)
	}
	if len(arguments) > 0 {
		arguments[0] = fit(arguments[0], key)
	}
	want := 1
	if name == "set" {
		want = 2
	}
	if name == "set" && len(arguments) == 2 {
		arguments[1] = fit(arguments[1], value)
	}
	if len(arguments) != want || arguments[0].Type() != key || (name == "set" && arguments[1].Type() != value) {
		return nil, true, l.notYet(node, "map."+name+" with arguments of other types")
	}
	switch name {
	case "get":
		return ir.MapGet{Map: object, Key: arguments[0], KeyType: key, ValueType: value}, true, nil
	case "set":
		return ir.MapSet{Map: object, Key: arguments[0], Value: arguments[1], KeyType: key, ValueType: value, Site: l.writeSite(receiver)}, true, nil
	case "has":
		return ir.MapHas{Map: object, Key: arguments[0], KeyType: key}, true, nil
	}
	return ir.MapDelete{Map: object, Key: arguments[0], KeyType: key}, true, nil
}

// stringMethod lowers string.trim() and string.charCodeAt(index).
func (l *lowering) stringMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if name == "trim" {
		if len(arguments) != 0 {
			return nil, true, l.notYet(node, "trim with arguments")
		}
		return ir.Trim{Value: value}, true, nil
	}
	index := ir.Expression(ir.NumberConstant{Value: 0})
	if len(arguments) > 1 {
		return nil, true, l.notYet(node, "charCodeAt with more than one argument")
	}
	if len(arguments) == 1 {
		if index, err = l.expression(arguments[0]); err != nil {
			return nil, true, err
		}
		if index.Type() != ir.Number {
			return nil, true, l.notYet(node, "charCodeAt with an index that isn't a number")
		}
	}
	return ir.CharCodeAt{Value: value, Index: index}, true, nil
}

// shorthand lowers the value of { value }. The name there is the field's, and asked for its symbol
// the checker gives the field; the variable it reads is a separate question.
func (l *lowering) shorthand(property *ast.Node) (ir.Expression, error) {
	symbol := l.checker.GetShorthandAssignmentValueSymbol(property)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	if symbol != nil {
		symbol = l.checker.GetExportSymbolOfSymbol(symbol)
	}
	local, isLocal := l.locals[symbol]
	if !isLocal {
		return nil, l.notYet(property, "reading "+property.Name().Text())
	}
	l.touch(local)
	of := l.result.Locals[local].Type
	if slotless(of) {
		return nil, l.notYet(property, "a field from a "+typeName(of)+" variable")
	}
	read := ir.Read{Local: local, Of: of, Checked: l.checked(local)}
	if identity := l.enumNeverIdentity(property.Name(), map[*ast.Node]bool{}); identity != nil {
		return l.enumNeverCheck(property.Name(), read, identity), nil
	}
	return read, nil
}

// stringMethods are the string methods stringCall lowers: the types of their arguments, and how many
// may be left out, each filled in with JavaScript's default.
var stringMethods = map[string]struct {
	arguments []ir.Type
	optional  int
}{
	"slice":       {[]ir.Type{ir.Number, ir.Number}, 2},
	"codePointAt": {[]ir.Type{ir.Number}, 1},
	"padStart":    {[]ir.Type{ir.Number, ir.String}, 1},
	"padEnd":      {[]ir.Type{ir.Number, ir.String}, 1},
	"repeat":      {[]ir.Type{ir.Number}, 0},
	"indexOf":     {[]ir.Type{ir.String, ir.Number}, 1},
	"includes":    {[]ir.Type{ir.String, ir.Number}, 1},
	"startsWith":  {[]ir.Type{ir.String}, 0},
	"endsWith":    {[]ir.Type{ir.String}, 0},
	"split":       {[]ir.Type{ir.String, ir.Number}, 1},
	"lastIndexOf": {[]ir.Type{ir.String}, 0},
	"trimStart":   {nil, 0},
	"trimEnd":     {nil, 0},
	"toUpperCase": {nil, 0},
	"toLowerCase": {nil, 0},
	"normalize":   {[]ir.Type{ir.String}, 1},
	"at":          {[]ir.Type{ir.Number}, 0},
	// A pattern that's a string, and a replacement that's a string: a regular expression or a
	// function there isn't a string, and stays not yet.
	"replace":    {[]ir.Type{ir.String, ir.String}, 0},
	"replaceAll": {[]ir.Type{ir.String, ir.String}, 0},
}

func (l *lowering) stringCall(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	shape := stringMethods[name]
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) > len(shape.arguments) || len(written) < len(shape.arguments)-shape.optional {
		return nil, true, l.notYet(node, name+" with these arguments")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if fallback, isOptional := optionalStrings[name][index]; isOptional {
			lowered = l.orDefault(argument, lowered, fallback)
		}
		if lowered.Type() != shape.arguments[index] {
			return nil, true, l.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	// JavaScript's defaults: codePointAt() is position 0, and padStart's fill is a space. slice's end
	// stays missing, which is different from any number, so the emitter is told how many were given.
	switch {
	case name == "split" && len(arguments) == 2:
		return ir.ArraySlice{
			Array: ir.StringCall{Method: name, Value: value, Arguments: arguments[:1]},
			Arguments: []ir.Expression{ir.NumberConstant{Value: 0},
				ir.Binary{Operator: ir.ShiftRightUnsigned, Left: arguments[1], Right: ir.NumberConstant{Value: 0}}},
		}, true, nil
	case name == "codePointAt" && len(arguments) == 0:
		arguments = append(arguments, ir.NumberConstant{Value: 0})
	case (name == "padStart" || name == "padEnd") && len(arguments) == 1:
		arguments = append(arguments, ir.StringConstant{Index: l.constant(" ")})
	case name == "normalize" && len(arguments) == 0:
		arguments = append(arguments, ir.StringConstant{Index: l.constant("NFC")})
	}
	return ir.StringCall{Method: name, Value: value, Arguments: arguments}, true, nil
}

// arraySort lowers array.sort(comparator). 0.1 requires the comparator (the default sorts numbers as
// strings), and stage 0 takes one of the module's functions by name.
func (l *lowering) arraySort(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) == 0 {
		return nil, true, &Refused{Where: l.program.Where(node), What: "sort without a comparator", Fix: "pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9]"}
	}
	comparator := ast.SkipParentheses(arguments[0])
	function, isFunction := l.functions[l.symbol(comparator)]
	if !ast.IsIdentifier(comparator) || !isFunction {
		// A function value: an arrow, or a variable holding one.
		callback, err := l.expression(comparator)
		if err != nil {
			return nil, true, err
		}
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(comparator), checker.SignatureKindCall)
		if callback.Type() != ir.Closure || len(signatures) != 1 {
			return nil, true, l.notYet(comparator, "a comparator that isn't a function")
		}
		if returns, _ := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0])); returns != ir.Number {
			return nil, true, l.notYet(comparator, "a comparator that doesn't return a number")
		}
		return ir.ArraySort{Array: array, Callback: callback, Element: element}, true, nil
	}
	declared := l.result.Functions[function]
	if declared.Returns != ir.Number || len(declared.Parameters) != 2 || l.result.Locals[declared.Parameters[0]].Type != element || l.result.Locals[declared.Parameters[1]].Type != element {
		return nil, true, l.notYet(comparator, "a comparator that doesn't take two elements and return a number")
	}
	return ir.ArraySort{Array: array, Comparator: function, Element: element}, true, nil
}

// elementAccess lowers array[index], string[index], and tuple[index] with a constant index: the
// tuple's field of that name.
func (l *lowering) elementAccess(node *ast.Node) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	index := ast.SkipParentheses(access.ArgumentExpression)
	if l.inheritedLibraryMember(node) && !l.regexRuntimeProperty(access.Expression, index.Text()) {
		return nil, l.prototypeRead(node, index.Text())
	}
	optional := access.QuestionDotToken != nil
	if !optional && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		// The rest of a chain after a ?., which short-circuits with it.
		return nil, l.notYet(node, "an optional chain longer than one step")
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() == ir.Object && l.regexGroups(access.Expression) {
		if index.Kind != ast.KindStringLiteral {
			return nil, l.notYet(node, "a computed named-group key")
		}
		of, err := l.typeOf(node)
		if err != nil {
			return nil, err
		}
		if of != ir.String && of != ir.Object {
			return nil, l.notYet(node, "a named-group key with an Object-prototype type")
		}
		return ir.RegExpGroup{Object: object, Name: index.Text(), Of: of, Optional: optional}, nil
	}
	if optional && object.Type() != ir.Object {
		// text?.[0] on a string that may be missing: indexing it as a string would read a null one.
		return nil, l.notYet(node, "?.[] on a "+typeName(object.Type()))
	}
	if object.Type() == ir.String {
		position, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if position.Type() != ir.Number {
			return nil, l.notYet(node, "a string index that isn't a number")
		}
		return ir.StringIndex{Value: object, Index: position}, nil
	}
	if object.Type() == ir.Array {
		element, err := l.elementType(access.Expression)
		if err != nil {
			return nil, err
		}
		position, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if position.Type() != ir.Number {
			return nil, l.notYet(node, "an array index that isn't a number")
		}
		return l.defined(node, ir.ArrayIndex{Array: object, Index: position, Element: element}), nil
	}
	// pairs[0]?.[0]: the tuple may be missing, and the read is undefined then (collections.go).
	if optional {
		return l.optionalTupleElement(node, object, index)
	}
	if object.Type() != ir.Object || index.Kind != ast.KindNumericLiteral || !checker.IsTupleType(l.checker.GetTypeAtLocation(access.Expression)) {
		return nil, l.notYet(node, describe(node))
	}
	of, err := l.typeOf(node)
	if elements := l.typeArguments(l.checker.GetTypeAtLocation(access.Expression)); err == nil {
		// What the tuple keeps there, not what the checker narrowed the read to: a Weak keeps a handle.
		if position, convertErr := strconv.Atoi(index.Text()); convertErr == nil && position < len(elements) {
			if declared, isKnown := l.representation(elements[position]); isKnown && declared == ir.Weak {
				of = declared
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if slotless(of) {
		return nil, l.notYet(node, "a tuple element of type "+typeName(of))
	}
	return ir.Property{Object: object, Name: index.Text(), Of: of}, nil
}

// setIndex lowers array[index] = value, as a statement.
func (l *lowering) setIndex(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	access := target.AsElementAccessExpression()
	array, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if array.Type() != ir.Array {
		return nil, l.notYet(target, "assigning an element of a "+typeName(array.Type()))
	}
	element, err := l.elementType(access.Expression)
	if err != nil {
		return nil, err
	}
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	if index.Type() != ir.Number {
		return nil, l.notYet(target, "an array index that isn't a number")
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.SetIndex{Array: array, Index: index, Value: fit(value, element), Element: element, Site: l.writeSite(access.Expression)}}, nil
}

// stringFromCodes lowers String.fromCharCode(...) and String.fromCodePoint(...), each argument a
// number, evaluated in order.
func (l *lowering) stringFromCodes(node *ast.Node, codePoints bool) (ir.Expression, bool, error) {
	lowered := ir.StringFromCodes{CodePoints: codePoints}
	if hasSpread(node) {
		spread, err := l.spreadNumbers(node)
		if err != nil {
			return nil, true, err
		}
		lowered.Spread = spread
		return lowered, true, nil
	}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, "a spread argument to String."+node.AsCallExpression().Expression.Name().Text())
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.Number {
			return nil, true, l.notYet(argument, "a "+typeName(value.Type())+" argument to String."+node.AsCallExpression().Expression.Name().Text())
		}
		lowered.Codes = append(lowered.Codes, value)
	}
	return lowered, true, nil
}

// tupleType is the tuple type an array literal makes, from the checker or from where it's written
// (return ['a', 1] in a function returning [string, number]), or nil when it makes an array.
func (l *lowering) tupleType(node *ast.Node) *checker.Type {
	// The tuple it's written into first: [5] written into [number, string?] is that tuple, with its
	// second element undefined, though the literal's own type is [number].
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil && checker.IsTupleType(contextual) && l.everyKnown(l.checker.GetTypeArguments(contextual)) {
		// (Under as const the contextual tuple's elements are any, and the literal's own type says
		// how they're held.)
		return contextual
	}
	if made := l.checker.GetTypeAtLocation(node); checker.IsTupleType(made) {
		return made
	}
	return nil
}

// tupleLiteral lowers [a, b] where it makes a tuple: an object whose fields are named "0", "1", ...,
// as a tuple's elements are read (elementAccess) and destructured (destructure, and for...of), each
// made what its element's type holds. A literal with as many elements as its tuple has is lowered;
// one leaving out an optional element, or spreading, is not yet.
func (l *lowering) tupleLiteral(node *ast.Node, tuple *checker.Type) (ir.Expression, error) {
	items := node.AsArrayLiteralExpression().Elements.Nodes
	elements := l.checker.GetTypeArguments(tuple)
	if len(items) > len(elements) {
		return nil, l.notYet(node, "a tuple literal with more values than its tuple has elements")
	}
	literal := ir.ObjectLiteral{Tuple: true}
	// An optional element left out is undefined, a field of its own, as tuple[index] reads it.
	missing := []ir.Field{}
	for index := len(items); index < len(elements); index++ {
		of, isKnown := l.representation(elements[index])
		if !isKnown || slotless(of) || !l.includesUndefined(elements[index]) || !(of.IsMaybe() || of.IsReference()) {
			return nil, l.notYet(node, "a tuple literal leaving out an element of type "+l.checker.TypeToString(elements[index]))
		}
		missing = append(missing, ir.Field{Name: strconv.Itoa(index), Value: fit(ir.Undefined{}, of)})
	}
	for index, item := range items {
		if item.Kind == ast.KindSpreadElement || item.Kind == ast.KindOmittedExpression {
			return nil, l.notYet(item, describe(item)+" in a tuple literal")
		}
		of, isKnown := l.representation(elements[index])
		if !isKnown || slotless(of) {
			return nil, l.notYet(item, "a tuple element of type "+l.checker.TypeToString(elements[index]))
		}
		value, err := l.expression(item)
		if err != nil {
			return nil, err
		}
		if value = fit(value, of); value.Type() != of {
			if _, isUndefined := value.(ir.Undefined); !isUndefined || !of.IsReference() {
				return nil, l.notYet(item, "a tuple element of another type than its tuple's")
			}
		}
		literal.Fields = append(literal.Fields, ir.Field{Name: strconv.Itoa(index), Value: value})
	}
	literal.Fields = append(literal.Fields, missing...)
	return literal, nil
}

// hasSpread reports whether a call spreads an argument: Math.max(...values).
func hasSpread(call *ast.Node) bool {
	for _, argument := range call.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return true
		}
	}
	return false
}

// spreadNumbers lowers a call's arguments, some of them spread, as the array of numbers they spell,
// evaluated in order, as [a, ...values, b] is.
func (l *lowering) spreadNumbers(call *ast.Node) (ir.Expression, error) {
	literal := ir.ArrayLiteral{Element: ir.Number}
	for _, argument := range call.AsCallExpression().Arguments.Nodes {
		spread := argument.Kind == ast.KindSpreadElement
		item := argument
		if spread {
			item = ast.SkipParentheses(argument.AsSpreadElement().Expression)
		}
		if spread && item.Kind == ast.KindArrayLiteralExpression {
			// ...[3, 4] is its elements, each an argument (the checker may type the literal a tuple).
			for _, element := range item.AsArrayLiteralExpression().Elements.Nodes {
				if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
					return nil, l.notYet(element, describe(element)+" in an array spread into a call")
				}
				value, err := l.expression(element)
				if err != nil {
					return nil, err
				}
				if value.Type() != ir.Number {
					return nil, l.notYet(element, "a "+typeName(value.Type())+" argument where numbers go")
				}
				literal.Elements = append(literal.Elements, value)
				literal.Spread = append(literal.Spread, false)
			}
			continue
		}
		value, err := l.expression(item)
		if err != nil {
			return nil, err
		}
		if spread {
			if element, err := l.elementType(item); err != nil || value.Type() != ir.Array || element != ir.Number {
				return nil, l.notYet(argument, "spreading other than an array of numbers into a call")
			}
		} else if value.Type() != ir.Number {
			return nil, l.notYet(argument, "a "+typeName(value.Type())+" argument where numbers go")
		}
		literal.Elements = append(literal.Elements, value)
		literal.Spread = append(literal.Spread, spread)
	}
	return literal, nil
}

// tupleWhereArrayGoes reports whether a value of type value, going where target is expected, carries a
// tuple to where an array is: directly, as an array's elements, as a Map's keys or values, as a field
// of an object, or as a function's result, or the other way round for a function's parameters (a
// function taking arrays called with tuples). A tuple is held as an object of its elements, so array
// code would read it as something it isn't; until a tuple can be one, it's refused there.
func (l *lowering) tupleWhereArrayGoes(value *checker.Type, target *checker.Type, depth int) bool {
	if value == nil || target == nil || value == target || depth > 4 {
		return false
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		// Into a union: fine if any member takes it as it is.
		for _, member := range target.Types() {
			if !l.tupleWhereArrayGoes(value, member, depth+1) {
				return false
			}
		}
		return true
	}
	if value.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range value.Types() {
			if l.tupleWhereArrayGoes(member, target, depth+1) {
				return true
			}
		}
		return false
	}
	valueTuple, targetTuple := checker.IsTupleType(value), checker.IsTupleType(target)
	switch {
	case valueTuple && !targetTuple && l.checker.IsArrayType(target):
		return true
	case valueTuple && targetTuple:
		return l.pairwiseTupleWhereArrayGoes(l.checker.GetTypeArguments(value), l.checker.GetTypeArguments(target), depth)
	case l.checker.IsArrayType(value) && l.checker.IsArrayType(target):
		return l.pairwiseTupleWhereArrayGoes(l.checker.GetTypeArguments(value), l.checker.GetTypeArguments(target), depth)
	case l.isLibraryType(value, "Map", "ReadonlyMap") && l.isLibraryType(target, "Map", "ReadonlyMap"):
		return l.pairwiseTupleWhereArrayGoes(l.checker.GetTypeArguments(value), l.checker.GetTypeArguments(target), depth)
	}
	if value.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	valueSignatures := l.checker.GetSignaturesOfType(value, checker.SignatureKindCall)
	targetSignatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(valueSignatures) == 1 && len(targetSignatures) == 1 {
		// A function: what it returns goes where the target's result goes, and what the target is
		// called with goes to its parameters.
		if l.tupleWhereArrayGoes(l.checker.GetReturnTypeOfSignature(valueSignatures[0]), l.checker.GetReturnTypeOfSignature(targetSignatures[0]), depth+1) {
			return true
		}
		valueParameters, targetParameters := valueSignatures[0].Parameters(), targetSignatures[0].Parameters()
		for index := range valueParameters {
			if index < len(targetParameters) && l.tupleWhereArrayGoes(l.checker.GetTypeOfSymbol(targetParameters[index]), l.checker.GetTypeOfSymbol(valueParameters[index]), depth+1) {
				return true
			}
		}
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if l.tupleWhereArrayGoes(l.checker.GetTypeOfPropertyOfType(value, field.Name), l.checker.GetTypeOfSymbol(field), depth+1) {
			return true
		}
	}
	return false
}

func (l *lowering) pairwiseTupleWhereArrayGoes(values []*checker.Type, targets []*checker.Type, depth int) bool {
	for index := range values {
		if index < len(targets) && l.tupleWhereArrayGoes(values[index], targets[index], depth+1) {
			return true
		}
	}
	return false
}

// updateIndex lowers array[index] op= value. The array and index are held before the element is
// read, and that read is held before the right side runs, exactly once each and in JavaScript's
// order. The final SetIndex uses those same held values even when the right side changes their source.
func (l *lowering) updateIndex(node *ast.Node, target *ast.Node, operator ast.Kind, valueNode *ast.Node) ([]ir.Statement, error) {
	access := target.AsElementAccessExpression()
	array, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if array.Type() != ir.Array {
		return nil, l.notYet(target, "assigning an element of a "+typeName(array.Type()))
	}
	element, err := l.elementType(access.Expression)
	if err != nil {
		return nil, err
	}
	if element != ir.Number {
		return nil, l.notYet(target, "updating an array element of type "+typeName(element))
	}
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	if index.Type() != ir.Number {
		return nil, l.notYet(target, "an array index that isn't a number")
	}

	arrayLocal := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "array", Type: ir.Array, Function: l.functionIndex})
	indexLocal := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "index", Type: ir.Number, Function: l.functionIndex})
	currentLocal := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "element", Type: ir.Number, Function: l.functionIndex})
	arrayRead := ir.Read{Local: arrayLocal, Of: ir.Array}
	indexRead := ir.Read{Local: indexLocal, Of: ir.Number}
	current := ir.Unwrap{Value: ir.ArrayIndex{Array: arrayRead, Index: indexRead, Element: ir.Number}}
	right, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	updated, err := l.combine(node, operator, ir.Read{Local: currentLocal, Of: ir.Number}, right)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Block{Body: []ir.Statement{
		ir.Declare{Local: arrayLocal, Value: array},
		ir.Declare{Local: indexLocal, Value: index},
		ir.Declare{Local: currentLocal, Value: current},
		ir.SetIndex{Array: arrayRead, Index: indexRead, Value: updated, Element: ir.Number, Site: l.writeSite(access.Expression)},
	}}}, nil
}

// isCallee reports whether a property read is the expression a call calls, parentheses aside.
func isCallee(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.Parent != nil && node.Parent.Kind == ast.KindCallExpression && node.Parent.AsCallExpression().Expression == node
}
