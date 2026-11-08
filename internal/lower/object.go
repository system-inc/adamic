package lower

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// objectLiteral lowers { name: value, ... }, and the one spread 0.1 allows: { ...source, fields },
// where every field replaces one the source's type already has.
func (lowering *lowering) objectLiteral(node *ast.Node) (ir.Expression, error) {
	if element := lowering.recordLiteralElement(node); element != nil {
		return lowering.recordLiteral(node, element)
	}
	if literal, handled, accessorLiteralError := lowering.accessorLiteral(node); handled {
		return literal, accessorLiteralError
	}
	literal := ir.ObjectLiteral{SpreadReadiness: sourceExpression(node), Record: lowering.entriesRecordLiteral(node)}
	for index, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		switch property.Kind {
		case ast.KindSpreadAssignment:
			if index != 0 {
				return nil, &Refused{
					Where: lowering.program.Where(property),
					What:  "a spread after the first field",
					Fix:   "spread once, first: { ...source, field: value } (adamic/single-spread)",
				}
			}
			spread, expressionError := lowering.expression(property.AsSpreadAssignment().Expression)
			if expressionError != nil {
				return nil, expressionError
			}
			if spread.Type() != ir.Object {
				modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
				if moduleOrderError != nil {
					return nil, moduleOrderError
				}
				if librarySpreadRefusalError := lowering.librarySpreadRefusal(librarySpread{
					node: property,
					view: lowering.concrete(
						lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression),
					),
				}, lowering.librarySpreadCandidates(modules)); librarySpreadRefusalError != nil {
					return nil, librarySpreadRefusalError
				}
				return nil, lowering.notYet(property, "spreading a "+typeName(spread.Type()))
			}
			literal.NoReuse = lowering.hasPrivateStorage(
				lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression),
			) || lowering.hasAccessorStorage(
				lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression),
			)
			literal.Spread = spread
			lowering.librarySpreads = append(lowering.librarySpreads, librarySpread{
				node:   property,
				source: spread,
				view:   lowering.concrete(lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)),
			})
			lowering.wholeObjectCopies = append(lowering.wholeObjectCopies, wholeObjectCopy{
				node:   property,
				source: spread,
				copied: lowering.concrete(lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)),
			})
			literal.SpreadMaybeUndefined = lowering.includesUndefined(
				lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression),
			)
		case ast.KindMethodDeclaration:
			name, known := lowering.methodName(property)
			if !known {
				return nil, lowering.notYet(property, "an unsupported computed method")
			}
			value, objectMethodError := lowering.objectMethod(property)
			if objectMethodError != nil {
				return nil, objectMethodError
			}
			literal.Fields = append(literal.Fields, ir.Field{Name: name, Value: value})
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			name := property.Name()
			if (ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral) && name.Text() == iteratorSlot {
				return nil, lowering.notYet(property, "a string field with the reserved iterator slot name")
			}
			fieldName, known := lowering.methodName(property)
			if !known {
				return nil, lowering.notYet(name, "a computed field name")
			}
			if property.Kind == ast.KindPropertyAssignment && fieldName == "__proto__" {
				return nil, &Refused{
					Where: lowering.program.Where(property),
					What:  "__proto__ in an object literal",
					Fix:   "JavaScript changes the prototype instead of making an own field; Adamic objects have fixed shapes and no prototype mutation",
				}
			}
			if property.Kind == ast.KindPropertyAssignment &&
				lowering.uninitializedInitializer(property.AsPropertyAssignment().Initializer) {
				declared := lowering.declaredField(node, fieldName)
				initializer := ast.SkipParentheses(property.AsPropertyAssignment().Initializer)
				if declared == 0 && initializer.Kind == ast.KindAsExpression {
					// An annotated placeholder reserves the annotation's storage even
					// when the object literal has no contextual type.
					declared, _ = lowering.representation(lowering.checker.GetTypeAtLocation(initializer))
				}
				declared = placeholderStorage(declared)
				if declared == 0 {
					return nil, lowering.notYet(
						property,
						"an uninitialized object field without a supported declared slot type",
					)
				}
				if literal.Spread != nil && !lowering.hasProperty(
					node.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression,
					fieldName,
				) {
					return nil, lowering.notYet(property, "a spread that adds a field the source doesn't have")
				}
				value := ir.Expression(zeroValue(declared))
				if declared.IsReference() {
					value = ir.Undefined{Of: declared}
				}
				value = lowering.placeholderInitialValue(property.AsPropertyAssignment().Initializer, declared)
				literal.Fields = append(
					literal.Fields,
					ir.Field{Name: fieldName, Value: value, Uninitialized: true, Unset: true},
				)
				continue
			}
			var value ir.Expression
			var err error
			if property.Kind == ast.KindPropertyAssignment {
				value, err = lowering.expression(property.AsPropertyAssignment().Initializer)
			} else {
				value, err = lowering.shorthand(property)
			}
			if err != nil {
				return nil, err
			}
			if literal.Spread != nil && !lowering.hasProperty(
				node.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression,
				fieldName,
			) {
				return nil, lowering.notYet(property, "a spread that adds a field the source doesn't have")
			}
			if lowering.placeholderOrigin(property.Name()) != "" {
				value = fit(value, placeholderStorage(value.Type()))
			}
			if declared := lowering.declaredField(node, fieldName); declared != 0 && !censusFieldSlotless(declared) {
				// Store the value as the member's slot holds it, rather than the initializer's type.
				value = fit(value, declared)
			}
			if literal.Spread != nil {
				source := node.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression
				if sourceField := lowering.checker.GetPropertyOfType(
					lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(source)),
					fieldName,
				); sourceField != nil {
					if held, typeOfSymbolError := lowering.typeOfSymbol(
						source,
						sourceField,
					); typeOfSymbolError == nil && held == ir.Union {
						// Spread keeps its source shape. Overrides must retain that slot's
						// counted representation rather than write an unboxed scalar into it.
						value = fit(value, held)
					}
				}
			}
			if censusFieldSlotless(value.Type()) {
				return nil, lowering.notYet(property, "a field holding "+typeName(value.Type()))
			}
			literal.Fields = append(literal.Fields, ir.Field{Name: fieldName, Value: value})
		default:
			return nil, lowering.notYet(property, describe(property)+" in an object literal")
		}
	}
	if literal.Record {
		if literal.Spread != nil {
			return nil, lowering.notYet(node, "a spread into record storage")
		}
		for index := range literal.Fields {
			field := &literal.Fields[index]
			if field.Value.Type() != ir.Number &&
				field.Value.Type() != ir.Boolean &&
				field.Value.Type() != ir.String &&
				field.Value.Type() != ir.Union {
				return nil, lowering.notYet(node, "record allocation outside scalar storage")
			}
			field.Value = fit(field.Value, ir.Union)
		}
	}
	if literal.SpreadMaybeUndefined {
		empty, emptySpreadError := lowering.emptySpread(node, literal.Fields)
		if emptySpreadError != nil {
			return nil, emptySpreadError
		}
		literal.Empty = empty
	}
	return lowering.reserveFactoryFields(node, literal)
}

// emptySpread is the object { ...source, fields } makes when source is undefined: JavaScript's is
// only the literal's own fields, and a read of any other field the source's type has is undefined,
// so each of those is there, undefined. A field that can hold undefined only as a pair (a boolean),
// or that keeps a handle (a Weak), says NotYet.
func (lowering *lowering) emptySpread(node *ast.Node, own []ir.Field) ([]ir.Field, error) {
	spread := node.AsObjectLiteralExpression().Properties.Nodes[0]
	given := map[string]bool{}
	for _, field := range own {
		given[field.Name] = true
	}
	empty := []ir.Field{}
	source := lowering.checker.GetNonNullableType(
		lowering.checker.GetTypeAtLocation(spread.AsSpreadAssignment().Expression),
	)
	for _, property := range lowering.checker.GetPropertiesOfType(source) {
		if given[property.Name] {
			continue
		}
		fieldType, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(property))
		switch fieldType {
		case ir.Number, ir.MaybeNumber:
			empty = append(empty, ir.Field{Name: property.Name, Value: ir.MaybeOf{Of: ir.MaybeNumber}})
		case ir.String, ir.Object, ir.Array, ir.Map, ir.Closure:
			empty = append(empty, ir.Field{Name: property.Name, Value: ir.Undefined{}})
		default:
			return nil, lowering.notYet(
				spread,
				"spreading a value that may be undefined, whose field "+property.Name+" can't be left undefined yet",
			)
		}
	}
	return empty, nil
}

// typeOfSymbol is what's left at runtime of a symbol's declared type, or NotYet at node.
func (lowering *lowering) typeOfSymbol(node *ast.Node, symbol *ast.Symbol) (ir.Type, error) {
	declared := lowering.checker.GetTypeOfSymbol(symbol)
	if lowering.placeholderSymbolOrigin(symbol) != "" ||
		lowering.result.PlaceholderViews[lowering.placeholderKey(symbol)] != "" {
		if stored, _ := lowering.representation(declared); stored == ir.Weak {
			return 0, lowering.notYet(
				node,
				"a placeholder slot holding Weak; its target-read semantics need a separate checked boundary",
			)
		}
		if held, known := lowering.representation(lowering.checker.GetNonNullableType(declared)); known {
			if held == ir.Weak {
				return 0, lowering.notYet(
					node,
					"a placeholder slot holding Weak; its target-read semantics need a separate checked boundary",
				)
			}
			return placeholderStorage(held), nil
		}
	}
	if valueType, isKnown := lowering.representation(declared); isKnown {
		if lowering.placeholderSymbolOrigin(symbol) != "" ||
			lowering.result.PlaceholderViews[lowering.placeholderKey(symbol)] != "" {
			valueType = placeholderStorage(valueType)
		}
		return valueType, nil
	}
	return 0, lowering.notYet(node, "a value of type "+lowering.checker.TypeToString(declared))
}

// declaredField is the representation of a field as the type an object literal is written into
// declares it, or 0 when there is no such type to say.
func (lowering *lowering) declaredField(literal *ast.Node, name string) ir.Type {
	contextual := lowering.slotType(literal)
	if contextual == nil {
		return 0
	}
	if contextual.Flags()&checker.TypeFlagsUnion != 0 {
		// The union's combined property can hold unrelated types. The literal belongs to a
		// member, whose fields are what reads after discriminant narrowing expect.
		made := lowering.checker.GetTypeAtLocation(literal)
		var shared ir.Type
		for _, member := range contextual.Types() {
			if !lowering.checker.IsTypeAssignableTo(made, member) {
				continue
			}
			field := lowering.checker.GetPropertyOfType(member, name)
			if field == nil {
				continue
			}
			declared, _ := lowering.typeOfSymbol(literal, field)
			if shared != 0 && shared != declared {
				return 0
			}
			shared = declared
		}
		return shared
	}
	field := lowering.checker.GetPropertyOfType(contextual, name)
	if field == nil {
		return 0
	}
	declared, _ := lowering.typeOfSymbol(literal, field)
	return declared
}

// hasProperty reports whether a value's type has a field of that name.
func (lowering *lowering) hasProperty(node *ast.Node, name string) bool {
	// A source that may be undefined has its fields where it's there.
	for _, property := range lowering.checker.GetPropertiesOfType(
		lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(node)),
	) {
		if property.Name == name {
			return true
		}
	}
	return false
}

func (lowering *lowering) arrayLiteral(node *ast.Node) (ir.Expression, error) {
	if tuple := lowering.tupleType(node); tuple != nil {
		return lowering.tupleLiteral(node, tuple)
	}
	element, elementTypeError := lowering.elementType(node)
	if elementTypeError != nil {
		return nil, elementTypeError
	}
	literal := ir.ArrayLiteral{Element: element}
	items := node.AsArrayLiteralExpression().Elements.Nodes
	if contextual := lowering.slotType(node); contextual != nil && lowering.isArrayType(contextual) {
		declared, known := lowering.representation(lowering.elementTypeOfArray(contextual))
		if !known || declared != element {
			for _, item := range items {
				if item.Kind == ast.KindSpreadElement {
					if function, isGenerator, generatorCalleeError := lowering.generatorCallee(
						item.AsSpreadElement().Expression,
					); generatorCalleeError != nil {
						return nil, generatorCalleeError
					} else if isGenerator {
						if lowering.result.Functions[function].Yields != declared {
							return nil, lowering.generatorNotYet(
								item,
								"spread into an array held differently from what it yields",
							)
						}
					} else if plan, planIterationError := lowering.planIteration(
						item.AsSpreadElement().Expression,
					); planIterationError != nil {
						return nil, planIterationError
					} else if plan != nil {
						return nil, lowering.notYet(
							node,
							"a custom spread whose element representation differs from its destination",
						)
					}
				}
			}
		}
	}

	if len(items) == 1 &&
		items[0].Kind == ast.KindSpreadElement &&
		!lowering.isArrayType(lowering.checker.GetTypeAtLocation(items[0].AsSpreadElement().Expression)) {
		// [...text] is the text's code points, [...set] its elements, [...map] its entries, and
		// [...map.keys()] and the rest what they give (collections.go).
		spread, given, iteratedError := lowering.iterated(items[0].AsSpreadElement().Expression)
		if iteratedError == nil && given != element {
			return nil, lowering.notYet(node, "a spread whose element representation differs from its destination")
		}
		return spread, iteratedError
	}
	spreads := false
	for _, item := range items {
		if item.Kind == ast.KindOmittedExpression {
			return nil, lowering.notYet(item, describe(item)+" in an array literal")
		}
		spread := item.Kind == ast.KindSpreadElement
		if spread {
			item = item.AsSpreadElement().Expression
		}
		var value ir.Expression
		var err error
		if spread {
			value, _, err = lowering.iterated(item)
		} else {
			value, err = lowering.expression(item)
		}
		if err != nil {
			return nil, err
		}
		if spread {
			if value.Type() != ir.Array {
				return nil, lowering.notYet(item, "spreading a "+typeName(value.Type())+" among other elements")
			}
			if _, other, iteratedTypeError := lowering.iteratedType(
				item,
			); iteratedTypeError != nil || other != element {
				return nil, lowering.notYet(item, "spreading an array of other elements")
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
func (lowering *lowering) elementType(node *ast.Node) (ir.Type, error) {
	arrayType := lowering.checker.GetNonNullableType(lowering.concrete(lowering.checker.GetTypeAtLocation(node)))
	if lowering.nodeBufferType(arrayType, "Buffer") {
		return ir.Number, nil
	}
	if lowering.isLibraryType(arrayType, "RegExpExecArray", "RegExpMatchArray") {
		return ir.String, nil
	}
	if lowering.isLibraryType(arrayType, "RegExpIndicesArray") {
		return ir.Object, nil
	}
	if literal := ast.SkipParentheses(node); literal.Kind == ast.KindArrayLiteralExpression {
		// A fresh [] has no slots yet. Use the same contextual element path as
		// generic returns; nested literals also have a best-common destination.
		if len(literal.AsArrayLiteralExpression().Elements.Nodes) == 0 {
			contexts := []*checker.Type{lowering.slotType(literal)}
			child := literal
			for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
				child = child.Parent
			}
			if parent := child.Parent; parent != nil && parent.Kind == ast.KindArrayLiteralExpression {
				// The checker can leave the inner literal as never[] even when
				// the enclosing array has an explicit union-of-arrays destination.
				contexts = append(contexts, lowering.literalArrayElement(lowering.slotType(parent)))
			}
			contexts = append(contexts, lowering.impliedTarget(literal))
			for _, contextual := range contexts {
				if element := lowering.literalArrayElement(contextual); element != nil {
					if held, known := lowering.kept(element); known && boxedSlot(held) {
						return held, nil
					}
				}
			}
		} else if contextual := lowering.slotType(literal); contextual != nil && lowering.isArrayType(contextual) {
			// Layout must follow the destination, even when the literal infers narrower elements.
			if declared, known := lowering.representation(
				lowering.elementTypeOfArray(contextual),
			); known && declared == ir.Union {
				if inferred, known := lowering.kept(
					lowering.elementTypeOfArray(arrayType),
				); known && inferred != ir.Union {
					return 0, &Refused{
						Where: lowering.program.Where(literal),
						What:  "union-typed array built from a narrower literal",
						Fix:   "declare the literal's element type explicitly, or build the array with a mixed literal",
					}
				}
			}
			// Keep every proven nonempty contextual element representation, including weak handles.
			if declared, _ := lowering.representation(
				lowering.elementTypeOfArray(contextual),
			); declared != 0 && boxedSlot(declared) {
				arrayType = contextual
			}
		}
	}
	if target := lowering.weakTarget(arrayType); target != nil {
		// A Weak<Node[]> narrowed to present is the array.
		arrayType = target
	}
	if _, element, known := lowering.nodeArrayLayoutOf(arrayType); known {
		held, _ := lowering.representation(element)
		return held, nil
	}
	if !lowering.isArrayType(arrayType) {
		return 0, lowering.notYet(
			node,
			"a value of type "+lowering.checker.TypeToString(arrayType)+" where an array goes",
		)
	}
	element := lowering.elementTypeOfArray(arrayType)
	if literal := ast.SkipParentheses(node); element.Flags()&checker.TypeFlagsNever != 0 &&
		literal.Kind == ast.KindArrayLiteralExpression &&
		len(literal.AsArrayLiteralExpression().Elements.Nodes) == 0 {
		return ir.Number, nil
	}
	if element.Flags()&checker.TypeFlagsUnknown != 0 {
		return 0, lowering.notYet(
			node,
			"an array of unknown with erased element storage (retain its declared element type before reading elements)",
		)
	}
	valueType, isKnown := lowering.kept(element)
	if !isKnown || (slotless(valueType) && valueType != ir.Union) {
		// An element is one adamic_value, and number | undefined needs two words.
		return 0, lowering.notYet(node, "an array of "+lowering.checker.TypeToString(element))
	}
	return valueType, nil
}

// readElementType admits unions only at element reads. Other elementType callers include
// writes and mutating methods, which must still require one array type.
func (lowering *lowering) readElementType(node *ast.Node) (ir.Type, error) {
	array := lowering.checker.GetNonNullableType(lowering.concrete(lowering.checker.GetTypeAtLocation(node)))
	if array.Flags()&checker.TypeFlagsUnion == 0 {
		return lowering.elementType(node)
	}
	elements := make([]*checker.Type, 0, len(array.Types()))
	for _, arm := range array.Types() {
		if !lowering.isArrayType(arm) {
			return 0, lowering.notYet(
				node,
				"an array union read with non-array arm "+lowering.checker.TypeToString(arm),
			)
		}
		elements = append(elements, lowering.elementTypeOfArray(arm))
	}
	result, known := lowering.kept(lowering.checker.GetUnionType(elements))
	if !known || !boxedSlot(result) {
		return 0, lowering.notYet(node, "an array union read with unsupported result element representation")
	}
	for _, element := range elements {
		held, known := lowering.kept(element)
		// Fitting a reference to Union only casts its pointer. Scalars need boxes,
		// and Weak holds a handle rather than the target that Union would observe.
		if element.Flags()&checker.TypeFlagsUnknown != 0 ||
			!known ||
			!boxedSlot(held) ||
			!(held == result || result == ir.Union && held.IsReference() && held != ir.Weak) {
			return 0, lowering.notYet(node, "an array union read whose arm element "+
				lowering.checker.TypeToString(element)+
				" has a different slot representation from the union element")
		}
	}
	return result, nil
}

// property lowers object.name, array.length, and Math's constants.
func (lowering *lowering) property(node *ast.Node) (ir.Expression, error) {
	if lowering.methodPresenceTest(node) {
		// Admitted as an observation (method_presence.go), so its consumer lowers it; one that doesn't
		// isn't built, and reading the method here would detach it.
		return nil, lowering.notYet(
			node,
			"a method read only to test for it, where that test doesn't lower through methodPresence",
		)
	}
	if value, found, nodeStdoutIsTtyReadError := lowering.nodeStdoutIsTtyRead(node); found {
		return value, nodeStdoutIsTtyReadError
	}
	if value, found, nodeRealpathValueError := lowering.nodeRealpathValue(node); found {
		return value, nodeRealpathValueError
	}
	if value, found, nodeEnvironmentReadError := lowering.nodeEnvironmentRead(node); found {
		return value, nodeEnvironmentReadError
	}
	if value, found, nodePlatformReadError := lowering.nodePlatformRead(node); found {
		return value, nodePlatformReadError
	}
	if eepPropertyPresenceError := lowering.eepPropertyPresence(node); eepPropertyPresenceError != nil {
		return nil, eepPropertyPresenceError
	}
	if staticPropertyError := lowering.staticProperty(node); staticPropertyError != nil {
		return nil, staticPropertyError
	}
	if call, handled, superAccessorError := lowering.superAccessor(node, nil); handled {
		return call, superAccessorError
	}
	access := node.AsPropertyAccessExpression()
	if lowering.entriesStringIndexed(access.Expression) &&
		lowering.checker.GetPropertyOfType(
			lowering.checker.GetTypeAtLocation(access.Expression),
			node.Name().Text(),
		) == nil &&
		lowering.entriesProgramHasRecords() {
		return nil, lowering.notYet(node, "an indexed record field read without a named declared property")
	}
	name := lowering.fieldName(node.Name())
	if _, iterator := lowering.libraryIteratorElement(access.Expression); iterator && name != "next" {
		return nil, lowering.notYet(node, "a collection iterator property other than next")
	}
	if access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		// The rest of a chain after a ?., which short-circuits with it.
		return nil, lowering.notYet(node, "an optional chain longer than one step")
	}
	if value, known := lowering.libraryMathNumberProperty(node); known {
		return value, nil
	}
	if lowering.isLibraryGlobal(access.Expression, "Math") {
		switch name {
		case "PI":
			return ir.NumberConstant{Value: 3.141592653589793}, nil
		case "E":
			return ir.NumberConstant{Value: 2.718281828459045}, nil
		case "random":
			return nil, refusedRandom(lowering, node)
		}
		return nil, lowering.notYet(node, "Math."+name)
	}
	if lowering.isLibraryGlobal(access.Expression, "Number") {
		if value, isConstant := numberConstants[name]; isConstant && name != "Infinity" {
			return ir.NumberConstant{Value: value}, nil
		}
		return nil, lowering.notYet(node, "Number."+name)
	}
	// A library declaration proves a prototype member exists, never an own slot. Keep this
	// guard in lowering too, even when the up-front unbound-method pass has already refused it. The
	// one callee it lets by is a hand-driven `x.next()` callClosure marked (handDrivenNext): a
	// direct call keeps its receiver, and checkIterableViews proves the receiver is a runtime-built
	// iterator, whose next is its own closure (Adamic's (a) on #qk8rztp S2c-2).
	if lowering.inheritedLibraryMember(node) &&
		node != lowering.handNext &&
		!lowering.iteratorResultField(access.Expression, lowering.memberSymbol(node)) &&
		!lowering.regexRuntimeProperty(access.Expression, name) &&
		name != "length" &&
		name != "size" &&
		!(lowering.isLibraryType(lowering.checker.GetTypeAtLocation(access.Expression), errorClasses...) &&
			(name == "name" || name == "message")) {
		return nil, lowering.prototypeRead(node, name)
	}
	if erasedLiteralMethodError := lowering.erasedLiteralMethod(node); erasedLiteralMethodError != nil {
		return nil, erasedLiteralMethodError
	}
	if read := lowering.checker.GetSymbolAtLocation(node.Name()); read != nil &&
		len(read.Declarations) > 0 &&
		read.Declarations[0].Kind == ast.KindMethodDeclaration &&
		!isCallee(node) {
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
			Where: lowering.program.Where(node),
			What:  "a method read off its object, which loses its this when called (unbound-method)",
			Fix: fmt.Sprintf(
				"wrap the call in an arrow function, which keeps its object: (value) => %s.%s(value)",
				object,
				name,
			),
		}
	}
	if receiver := lowering.checker.GetTypeAtLocation(access.Expression); name != "length" &&
		(checker.IsTupleType(receiver) || checker.IsTupleType(lowering.checker.GetNonNullableType(receiver))) {
		// A tuple is held as an object of its elements, "0", "1", ..., read by index; an array's
		// methods aren't fields of it, and reading them as fields would find nothing. Its length is
		// read below (tupleLength).
		return nil, lowering.notYet(node, "."+name+" on a tuple")
	}
	object, expressionError := lowering.expression(access.Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	if object.Type() == ir.Object || object.Type() == ir.Array {
		if value, handled, factoryFieldReadError := lowering.factoryFieldRead(node, object, name); handled {
			return value, factoryFieldReadError
		}
	}
	if object.Type() == ir.Union &&
		access.QuestionDotToken != nil &&
		lowering.nullableObjectUnion(lowering.checker.GetTypeAtLocation(access.Expression)) {
		object = lowering.optionalNullableObject(object)
	}
	if object.Type() == ir.Union {
		return lowering.dynamicProperty(node, object, name)
	}
	if field := lowering.checker.GetSymbolAtLocation(
		node.Name(),
	); field != nil && lowering.checker.GetTypeOfSymbol(field).Flags()&checker.TypeFlagsUnknown != 0 {
		return lowering.dynamicProperty(node, fit(object, ir.Union), name)
	}
	if object.Type() == ir.Object &&
		lowering.isLibraryType(
			lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(access.Expression)),
			"RegExp",
		) {
		switch name {
		case "lastIndex", "source", "flags", "global", "ignoreCase", "multiline", "unicode", "sticky", "hasIndices", "unicodeSets", "dotAll":
		default:
			return nil, lowering.notYet(node, "a prototype property on a RegExp")
		}
	}
	if object.Type() == ir.Object && lowering.regexGroups(access.Expression) {
		of, typeOfError := lowering.typeOf(node)
		if typeOfError != nil {
			return nil, typeOfError
		}
		if of != ir.String && of != ir.Object {
			return nil, lowering.notYet(node, "a named-group key with an Object-prototype type")
		}
		return ir.RegExpGroup{Object: object, Name: name, Of: of, Optional: access.QuestionDotToken != nil}, nil
	}
	if object.Type() == ir.Object && name == "done" {
		proven := lowering.checker.GetTypeAtLocation(access.Expression)
		members := []*checker.Type{proven}
		if proven.Flags()&checker.TypeFlagsUnion != 0 {
			members = proven.Types()
		}
		iterator := true
		for _, member := range members {
			iterator = iterator && lowering.isLibraryType(member, "IteratorYieldResult", "IteratorReturnResult")
		}
		if iterator {
			of, typeOfError := lowering.typeOf(node)
			return ir.RegExpCall{Value: object, Method: "iteratorDone", Returns: of}, typeOfError
		}
	}
	object = lowering.privateStaticReceiver(node.Name(), object, false)
	if access.QuestionDotToken != nil && (object.Type() == ir.Array || object.Type() == ir.String) && name == "length" {
		// words?.length and text?.length: undefined where the array or string is, a number | undefined.
		if object.Type() == ir.Array {
			return ir.Length{Array: object, Optional: true}, nil
		}
		return ir.StringLength{Value: object, Optional: true}, nil
	}
	if object.Type() == ir.Array && name != "length" {
		if value, handled, nodeArrayPropertyError := lowering.nodeArrayProperty(node, object, name); handled {
			return value, nodeArrayPropertyError
		}
	}
	if object.Type() == ir.Array && name != "length" && lowering.isLibraryType(
		lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(access.Expression)),
		"RegExpExecArray",
		"RegExpMatchArray",
		"RegExpIndicesArray",
	) {
		of, typeOfError := lowering.typeOf(node)
		if typeOfError != nil {
			return nil, typeOfError
		}
		switch name {
		case "index", "input", "groups", "indices":
		default:
			return nil, lowering.notYet(node, "a prototype property on a RegExp result")
		}
		stored := of
		if name == "index" {
			if lowering.isLibraryType(
				lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(access.Expression)),
				"RegExpMatchArray",
			) {
				stored = ir.MaybeNumber
			} else {
				stored = ir.Number
			}
		}
		result := ir.Expression(
			ir.RegExpProperty{Array: object, Name: name, Of: stored, Optional: access.QuestionDotToken != nil},
		)
		if stored == ir.MaybeNumber && of == ir.Number {
			result = ir.Unwrap{Value: result}
		}
		return result, nil
	}
	if access.QuestionDotToken != nil && object.Type() != ir.Object {
		return nil, lowering.notYet(node, "optional chaining to ."+name+" on a "+typeName(object.Type()))
	}
	switch {
	case object.Type() == ir.Array && name == "length":
		return ir.Length{Array: object}, nil
	case object.Type() == ir.String && name == "length":
		return ir.StringLength{Value: object}, nil
	case object.Type() == ir.Map && name == "size":
		return ir.MapSize{Map: object}, nil
	case object.Type() == ir.Object &&
		name == "length" &&
		checker.IsTupleType(lowering.checker.GetTypeAtLocation(access.Expression)):
		// A tuple is an object of fields "0", "1" and on, with no length to read; a fixed one's length
		// is its type's, read here when nothing about reading the tuple itself could be seen.
		return lowering.tupleLength(node, access.Expression)
	case object.Type() == ir.Object:
		if origin := lowering.placeholderReadOrigin(node); origin != "" {
			stored, typeOfSymbolError := lowering.typeOfSymbol(node, lowering.symbol(node))
			if typeOfSymbolError != nil {
				return nil, typeOfSymbolError
			}
			property := ir.Property{
				Object:   object,
				Name:     name,
				Of:       stored,
				Optional: access.QuestionDotToken != nil,
				Class:    lowering.classOf(node),
				Unset:    true,
			}
			property.Absent = lowering.symbol(node).Flags&ast.SymbolFlagsOptional != 0
			if lowering.placeholderOrigin(node) == "" {
				present, _ := lowering.representation(
					lowering.checker.GetNonNullableType(lowering.checker.GetTypeOfSymbol(lowering.symbol(node))),
				)
				property.UnsetType = present
				if present != ir.Number &&
					present != ir.Boolean &&
					present != ir.String &&
					present != ir.Object &&
					present != ir.Array &&
					present != ir.Map {
					return nil, lowering.notYet(
						node,
						"a placeholder field view without a physical-kind check for this representation",
					)
				}
				property.View = sourceExpression(node)
				property.ViewType = lowering.checker.TypeToString(
					lowering.checker.GetTypeOfSymbol(lowering.symbol(node)),
				)
				property.ViewAllowed = lowering.viewLiterals(lowering.checker.GetTypeOfSymbol(lowering.symbol(node)))
				lowering.result.PlaceholderChecks = append(lowering.result.PlaceholderChecks, ir.PlaceholderCheck{
					Origin: origin,
					Use:    "field representation",
					Path:   sourceExpression(node),
					Where:  lowering.program.Where(node),
					Status: "checked",
				})
			}
			return lowering.placeholderRead(node, property, origin), nil
		}
		if field := lowering.checker.GetSymbolAtLocation(node.Name()); field != nil {
			if declared, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(field)); declared == ir.Weak {
				// The field keeps a handle, whatever the checker narrowed the read to.
				return lowering.readObjectField(
					node,
					ir.Property{Object: object, Name: name, Of: ir.Weak, Optional: access.QuestionDotToken != nil},
				), nil
			}
		}
		if field := lowering.checker.GetSymbolAtLocation(
			node.Name(),
		); field != nil && (accessorSymbol(field) || (lowering.accessorNames[node.Name().Text()] &&
			!isClassInstance(lowering.checker.GetTypeAtLocation(access.Expression)) &&
			ast.SkipParentheses(access.Expression).Kind != ast.KindThisKeyword)) {
			declared := lowering.checker.GetTypeOfSymbol(field)
			observed := lowering.checker.GetTypeAtLocation(node)
			if !lowering.classAssignable(declared, observed) {
				return nil, &Refused{
					Where: lowering.program.Where(node),
					What:  "a narrowed accessor reread, which can return a different value",
					Fix:   "read the getter into a local once, then narrow and use that local",
				}
			}
		}
		of, err := lowering.typeOf(node)
		if err != nil && lowering.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
			// Narrowed to undefined (just assigned it): read as the field is declared.
			if field := lowering.checker.GetSymbolAtLocation(node.Name()); field != nil {
				of, err = lowering.typeOfSymbol(node, field)
			}
		}
		if err != nil {
			return nil, err
		}
		optional := access.QuestionDotToken != nil
		if field := lowering.checker.GetSymbolAtLocation(node.Name()); field != nil {
			if lowering.nullableObjectUnion(lowering.checker.GetTypeOfSymbol(field)) &&
				(of == ir.Object || of == ir.Union) {
				read := lowering.readObjectField(
					node,
					ir.Property{
						Object:   object,
						Name:     name,
						Of:       ir.Union,
						Optional: optional,
						Class:    lowering.classOf(node),
					},
				)
				return lowering.checkedNullableObject(node, read), nil
			}
			if declared := lowering.concrete(
				lowering.checker.GetTypeOfSymbol(field),
			); nullableStringUnion(declared) && of == ir.String {
				read := lowering.readObjectField(
					node,
					ir.Property{
						Object:   object,
						Name:     name,
						Of:       ir.Union,
						Optional: optional,
						Class:    lowering.classOf(node),
					},
				)
				return lowering.checkedNullableString(node, read), nil
			}
			if stored, known := lowering.representation(lowering.checker.GetTypeOfSymbol(field)); known &&
				stored == ir.Union &&
				(of.Present() == ir.Number || of.Present() == ir.Boolean || of == ir.String) {
				read := lowering.readObjectField(
					node,
					ir.Property{
						Object:   object,
						Name:     name,
						Of:       ir.Union,
						Optional: optional,
						Class:    lowering.classOf(node),
					},
				)
				return lowering.checkedBoxedScalarField(node, read, of), nil
			}
			if stored, known := lowering.representation(
				lowering.checker.GetTypeOfSymbol(field),
			); known && stored.IsMaybe() && of == stored.Present() {
				// Read the declared representation before trusting the narrowing. A call or an
				// alias write may have restored undefined, just as for a narrowed variable.
				if censusFieldSlotless(stored) {
					return nil, lowering.notYet(
						node,
						"a narrowed boolean | undefined field; copy the field into a local and narrow that local instead",
					)
				}
				read := lowering.readObjectField(
					node,
					ir.Property{
						Object:   object,
						Name:     name,
						Of:       stored,
						Optional: optional,
						Class:    lowering.classOf(node),
					},
				)
				if lowering.acceptsUndefined(node) {
					return read, nil
				}
				return ir.Unwrap{Value: read}, nil
			}
		}
		if (of == ir.Union || of.IsMaybe()) && lowering.entriesProgramHasRecords() {
			receiver := lowering.checker.GetTypeAtLocation(access.Expression)
			if lowering.recordMayReach(receiver) {
				// A .ts read tests its receiver's storage as it runs (weakFieldRead). A .a read stays
				// stopped: .a is to refuse a record seen as a weak type at the relation, so a read that
				// reaches here in .a is a hole in that rule. A name Object.prototype holds is no record's
				// own key to read, so it stays stopped too.
				if !strings.HasSuffix(lowering.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") ||
					recordPrototypeName(name) {
					return nil, lowering.notYet(node, "a union or optional field read in a program with record storage")
				}
				return lowering.weakFieldRead(node, object, name, of, optional)
			}
			// The id the read is checked under, routed as the read is, or a routed read would look
			// unchecked here and fail open (Lower's must on 148e6df3).
			lowering.recordGuardedReads = append(
				lowering.recordGuardedReads,
				recordGuardedRead{
					node:           node,
					receiverTypeId: lowering.viewFieldReceiverTypeID(node, receiver),
					name:           name,
				},
			)
		}
		return lowering.fixedFieldRead(node, object, name, of, optional)
	}
	return nil, lowering.notYet(node, "."+name+" on a "+typeName(object.Type()))
}

// fixedFieldRead is a union, maybe or plain field read of object, a fixed object, from its slot. A
// weak field read (weakFieldRead) reads a fixed object's field through it too.
func (lowering *lowering) fixedFieldRead(
	node *ast.Node,
	object ir.Expression,
	name string,
	of ir.Type,
	optional bool,
) (ir.Expression, error) {
	if of.IsMaybe() && optional {
		// box?.size is number | undefined because box may be; the field itself is what's stored.
		if field := lowering.checker.GetSymbolAtLocation(node.Name()); field != nil {
			if stored, isKnown := lowering.representation(
				lowering.checker.GetTypeOfSymbol(field),
			); isKnown && stored == of.Present() {
				return lowering.readObjectField(
					node,
					ir.Property{Object: object, Name: name, Of: stored, Optional: true, Class: lowering.classOf(node)},
				), nil
			}
		}
	}
	if symbol := lowering.checker.GetSymbolAtLocation(node.Name()); symbol != nil {
		id := lowering.result.ViewContractTypes[int(lowering.checker.GetTypeOfSymbol(symbol).Id())]
		if id > 0 && strings.Contains(lowering.result.ViewContracts[id-1].Unsupported, "views-v3: array element kind") {
			return nil, lowering.notYet(
				node,
				"checked view union arm awaits "+lowering.result.ViewContracts[id-1].Unsupported,
			)
		}
	}
	if censusFieldSlotless(of) && !lowering.unionReadCertificate(node, of) {
		return nil, lowering.notYet(
			node,
			"a field of type "+lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(node)),
		)
	}
	if of.IsMaybe() {
		// number | undefined, whether the field holds it or ?. makes it: the packed word, or
		// undefined when the object is.
		return lowering.readObjectField(
			node,
			ir.Property{Object: object, Name: name, Of: of, Optional: optional, Class: lowering.classOf(node)},
		), nil
	}
	if optional && !of.IsReference() {
		return nil, lowering.notYet(node, "?. to a "+typeName(of)+", which would be "+typeName(of)+" | undefined")
	}
	return lowering.defined(node, lowering.readObjectField(
		node,
		ir.Property{Object: object, Name: name, Of: of, Optional: optional, Class: lowering.classOf(node)},
	)), nil
}

// readObjectField keeps an optional own field distinct from optional chaining of its receiver.
// Absence is a read result, never a synthetic own field: hasOwnProperty and object spread still see
// the shape that was actually made. A narrowed number checks the declared optional representation.
func (lowering *lowering) readObjectField(node *ast.Node, property ir.Property) ir.Expression {
	property.Readiness = sourceExpression(node)
	property = lowering.viewedField(node, property)
	field := lowering.checker.GetSymbolAtLocation(node.Name())
	if field != nil {
		for _, declaration := range field.Declarations {
			if declaration.Kind == ast.KindPropertyDeclaration {
				initializer := declaration.AsPropertyDeclaration().Initializer
				if lowering.lazyAssertionInitializer(initializer) && !lowering.uninitializedInitializer(initializer) {
					property.Readiness = sourceExpression(initializer)
				}
			}
		}
	}
	if field == nil || field.Flags&ast.SymbolFlagsOptional == 0 {
		return property
	}
	property.Absent = true
	if declared, _ := lowering.representation(
		lowering.checker.GetTypeOfSymbol(field),
	); declared.IsMaybe() && property.Of == declared.Present() {
		property.Of = declared
		return fit(property, declared.Present())
	}
	return property
}

// viewedField gives a field read what a checked view of its receiver checks it by. Readiness drops
// all of it again unless lowering ends with the field checked.
func (lowering *lowering) viewedField(node *ast.Node, property ir.Property) ir.Property {
	property.View = sourceExpression(node)
	property.ViewWhere = lowering.program.Where(node)
	if node.Kind == ast.KindPropertyAccessExpression {
		receiver := lowering.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression)
		property.ViewReceiverTypeID = lowering.viewFieldReceiverTypeID(node, receiver)
	}
	if symbol := lowering.checker.GetSymbolAtLocation(node.Name()); symbol != nil {
		declared := lowering.checker.GetTypeOfSymbol(symbol)
		if narrowed := lowering.checker.GetTypeAtLocation(node); narrowed.Id() != declared.Id() {
			// A field read the checker narrowed, as a predicate narrows its argument (view_widened.go).
			lowering.narrowedFields = append(lowering.narrowedFields, narrowed)
		}
		property.ViewTypeID = lowering.viewTypeId(declared)
		property.ViewContract = lowering.result.ViewContractTypes[property.ViewTypeID]
		property.ViewType = lowering.checker.TypeToString(declared)
		property.ViewAllowed = lowering.viewLiterals(declared)
	}
	return property
}

// hasOwnProperty lowers object.hasOwnProperty(key) when the method is the library's, not a field the
// object declares. The checker puts Object's prototype methods on every object's type. The shape
// holds only the object's own fields, and reading hasOwnProperty as one of them panics: the field
// the checker proved is not there.
func (lowering *lowering) hasOwnProperty(node *ast.Node, receiver *ast.Node) (ir.Expression, bool, error) {
	return lowering.hasOwnPropertyArguments(node, receiver, node.AsCallExpression().Arguments.Nodes)
}

func (lowering *lowering) hasOwnPropertyArguments(
	node *ast.Node,
	receiver *ast.Node,
	arguments []*ast.Node,
) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	symbol := lowering.checker.GetSymbolAtLocation(callee)
	if symbol == nil ||
		len(symbol.Declarations) == 0 ||
		!load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return nil, false, nil
	}
	if len(arguments) != 1 {
		return nil, true, lowering.notYet(node, "hasOwnProperty with other than one argument")
	}
	if keyType, _ := lowering.representation(lowering.checker.GetTypeAtLocation(arguments[0])); keyType != ir.String {
		return nil, true, lowering.notYet(node, "hasOwnProperty with a key that isn't a string")
	}
	// JavaScript's order: the object, then the key.
	object, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	key, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	return ir.HasOwn{Object: object, Key: key}, true, nil
}

// isLibraryGlobal reports whether a node names one of the library's globals, rather than something
// of the program's that happens to share its name.
func (lowering *lowering) isLibraryGlobal(node *ast.Node, name string) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) || node.Text() != name {
		return false
	}
	symbol := lowering.checker.GetSymbolAtLocation(node)
	return symbol != nil &&
		len(symbol.Declarations) > 0 &&
		load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
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

// refusedRandom stops Math.random until it's built: from the OS's entropy, with V8's range, under the
// oracle discipline the current time follows (a fixture prints facts about it, never the value). Ruled
// November 1, 2026: determinism is how we test, not what the language is, so this is not yet, not a
// refusal (docs/0.1.md).
func refusedRandom(lowering *lowering, node *ast.Node) error {
	return lowering.notYet(node, "Math.random")
}

// builtin lowers a call to Math or a number's toFixed. isBuiltin is false for any other call.
func (lowering *lowering) builtin(node *ast.Node) (ir.Expression, bool, error) {
	if value, known, processValueError := lowering.processValue(node); known {
		return value, true, processValueError
	}
	if value, known, reflectGetError := lowering.reflectGet(node); known {
		return value, true, reflectGetError
	}
	if value, known, nodeFileSystemDirectoryCallError := lowering.nodeFileSystemDirectoryCall(node); known {
		return value, true, nodeFileSystemDirectoryCallError
	}
	if value, known, arrayIsArrayError := lowering.arrayIsArray(node); known {
		return value, known, arrayIsArrayError
	}
	if lowering.isOptionalJoin(node) {
		value, optionalJoinError := lowering.optionalJoin(node)
		return value, true, optionalJoinError
	}
	if value, handled, libraryNodeBufferError := lowering.libraryNodeBuffer(node); handled {
		return value, true, libraryNodeBufferError
	}
	if value, handled, detachedOwnCallError := lowering.detachedOwnCall(node); handled {
		return value, true, detachedOwnCallError
	}
	if value, handled, userMethodCallError := lowering.userMethodCall(node); handled {
		return value, true, userMethodCallError
	}
	if value, known, libraryEmptyMapError := lowering.libraryEmptyMap(node); known {
		return value, true, libraryEmptyMapError
	}
	if value, known, libraryMethodCallError := lowering.libraryMethodCall(node); known {
		return value, true, libraryMethodCallError
	}
	if value, known, libraryMathNumberCallError := lowering.libraryMathNumberCall(node); known {
		return value, true, libraryMathNumberCallError
	}
	if lowered, isString, libraryStringError := lowering.libraryString(node); isString {
		return lowered, true, libraryStringError
	}
	if value, matched, regexBuiltinError := lowering.regexBuiltin(node); matched {
		return value, true, regexBuiltinError
	}
	if value, handled, objectKeysError := lowering.objectKeys(node); handled {
		return value, true, objectKeysError
	}
	if lowered, isInput, inputError := lowering.input(node); isInput {
		return lowered, true, inputError
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if value, known, dateNowError := lowering.dateNow(node, callee); known {
		return value, true, dateNowError
	}
	// The global parseInt and parseFloat are Number's, the same functions.
	if lowering.isLibraryGlobal(callee, "parseInt") || lowering.isLibraryGlobal(callee, "parseFloat") {
		return lowering.numberCall(node, callee.Text())
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	if lowering.isLibraryGlobal(receiver, "Object") {
		return lowering.objectCall(node, name)
	}
	if lowered, handled, objectPrototypeCallError := lowering.objectPrototypeCall(node, receiver, name); handled {
		return lowered, true, objectPrototypeCallError
	}
	if lowering.isLibraryGlobal(receiver, "JSON") {
		return lowering.jsonCall(node, name)
	}
	if lowering.isLibraryGlobal(receiver, "Number") {
		return lowering.numberCall(node, name)
	}
	if lowering.isLibraryGlobal(receiver, "Map") && name == "groupBy" {
		return lowering.libraryMapGroupBy(node)
	}
	if lowering.isLibraryGlobal(receiver, "Array") && name == "from" {
		return lowering.arrayFrom(node)
	}
	if lowering.isLibraryGlobal(receiver, "String") && (name == "fromCharCode" || name == "fromCodePoint") {
		return lowering.stringFromCodes(node, name == "fromCodePoint")
	}
	if receiverType, _ := lowering.representation(
		lowering.checker.GetTypeAtLocation(receiver),
	); receiverType == ir.Number && name == "toString" && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		value, expressionError := lowering.expression(receiver)
		if expressionError != nil {
			return nil, true, expressionError
		}
		return ir.NumberToString{Value: value}, true, nil
	}
	if receiverType, _ := lowering.representation(
		lowering.checker.GetTypeAtLocation(receiver),
	); receiverType == ir.Number && (name == "toExponential" || name == "toPrecision" || name == "toString") {
		return lowering.numberFormat(node, receiver, name)
	}
	isMath := lowering.isLibraryGlobal(receiver, "Math")
	receiverType, _ := lowering.representation(lowering.checker.GetTypeAtLocation(receiver))
	isToFixed := name == "toFixed" && receiverType == ir.Number
	if receiverType == ir.Array && name == "pop" && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		element, elementTypeError := lowering.elementType(receiver)
		if elementTypeError != nil {
			return nil, true, elementTypeError
		}
		array, expressionError := lowering.expression(receiver)
		if expressionError != nil {
			return nil, true, expressionError
		}
		return ir.ArrayPop{Array: array, Element: element}, true, nil
	}
	if receiverType == ir.Array && libraryArrayMethods[name] {
		return lowering.libraryArrayMethod(node, receiver, name)
	}
	if _, isVisit := visits[name]; receiverType == ir.Array && (isVisit || arrayMethods[name]) {
		return lowering.arrayMethod(node, receiver, name)
	}
	if receiverType == ir.Map && (name == "keys" || name == "values" || name == "entries") {
		return lowering.libraryCollectionIterator(node, receiver, name)
	}
	if receiverType == ir.Map && lowering.isSet(receiver) {
		return lowering.setMethod(node, receiver, name)
	}
	if receiverType == ir.Map && (name == "clear" || name == "forEach") {
		return lowering.clearOrVisit(node, receiver, name, false)
	}
	if receiverType == ir.Map && (name == "get" || name == "set" || name == "has" || name == "delete") {
		return lowering.mapMethod(node, receiver, name)
	}
	if receiverType == ir.String && (name == "trim" || name == "charCodeAt") {
		return lowering.stringMethod(node, receiver, name)
	}
	if _, isKnown := stringMethods[name]; isKnown && receiverType == ir.String {
		return lowering.stringCall(node, receiver, name)
	}
	if receiverType == ir.Object && checker.IsTupleType(lowering.checker.GetTypeAtLocation(receiver)) {
		// A tuple is held as an object: called as an object, an array method would be read as a field.
		return nil, true, lowering.notYet(node, name+
			" on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made)")
	}
	if !isMath && !isToFixed {
		return nil, false, nil
	}
	// JavaScript's order: the receiver, then the arguments.
	var value ir.Expression
	if isToFixed {
		var expressionError error
		if value, expressionError = lowering.expression(receiver); expressionError != nil {
			return nil, true, expressionError
		}
	}
	if isMath && (name == "max" || name == "min" || name == "hypot") && hasSpread(node) {
		spread, spreadNumbersError := lowering.spreadNumbers(node)
		if spreadNumbersError != nil {
			return nil, true, spreadNumbersError
		}
		return ir.MathCall{Function: name, Spread: spread}, true, nil
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, lowering.notYet(argument, "a spread argument to "+name)
		}
		lowered, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if lowered.Type() != ir.Number {
			return nil, true, lowering.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	if isMath {
		count, isKnown := mathFunctions[name]
		if name == "random" {
			return nil, true, refusedRandom(lowering, node)
		}
		if !isKnown {
			return nil, true, lowering.notYet(node, "Math."+name)
		}
		if count >= 0 && len(arguments) != count {
			return nil, true, lowering.notYet(node, "Math."+name+" with other than its arguments")
		}
		return ir.MathCall{Function: name, Arguments: arguments}, true, nil
	}
	if len(arguments) > 1 {
		return nil, true, lowering.notYet(node, "toFixed with more than one argument")
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
func (lowering *lowering) numberFormat(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	value, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) > 1 {
		return nil, true, lowering.notYet(node, name+" with more than one argument")
	}
	format := ir.NumberFormat{Method: name, Value: value}
	if len(arguments) == 1 {
		argument, expressionError := lowering.expression(arguments[0])
		if expressionError != nil {
			return nil, true, expressionError
		}
		if argument.Type() != ir.Number {
			return nil, true, lowering.notYet(arguments[0], "a "+typeName(argument.Type())+" argument to "+name)
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
func (lowering *lowering) numberCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	return lowering.numberCallArguments(node, name, node.AsCallExpression().Arguments.Nodes)
}

func (lowering *lowering) numberCallArguments(
	node *ast.Node,
	name string,
	written []*ast.Node,
) (ir.Expression, bool, error) {
	takes, isKnown := numberFunctions[name]
	if !isKnown {
		return nil, true, lowering.notYet(node, "Number."+name)
	}
	optional := 0
	if name == "parseInt" {
		optional = 1
	}
	if len(written) > len(takes) || len(written) < len(takes)-optional {
		return nil, true, lowering.notYet(node, name+" with these arguments")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		lowered, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if lowered.Type() != takes[index] {
			// Number.isNaN('x') is false in JavaScript, and a string there is a mistake 0.1 won't guess at.
			return nil, true, lowering.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
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
func (lowering *lowering) forOf(node *ast.Node) ([]ir.Statement, error) {
	if statements, known, libraryArrayForOfError := lowering.libraryArrayForOf(node); known {
		return statements, libraryArrayForOfError
	}
	statement := node.AsForInOrOfStatement()
	if statement.AwaitModifier != nil {
		return nil, lowering.notYet(node, "for await")
	}
	// A call to a program's generator is stepped as the generator it is, whatever iterator type it
	// is declared to return (#qk8rztp, Compiler's risk 2).
	if function, isGenerator, generatorCalleeError := lowering.generatorCallee(
		statement.Expression,
	); generatorCalleeError != nil {
		return nil, generatorCalleeError
	} else if isGenerator {
		initializer := statement.Initializer
		if initializer.Kind != ast.KindVariableDeclarationList ||
			initializer.Flags&ast.NodeFlagsBlockScoped == 0 ||
			len(initializer.AsVariableDeclarationList().Declarations.Nodes) != 1 {
			return nil, lowering.generatorNotYet(
				initializer,
				"consumed by a for-of that doesn't declare one variable with const or let",
			)
		}
		return lowering.forOfGenerator(
			node,
			function,
			initializer.AsVariableDeclarationList().Declarations.Nodes[0].Name(),
		)
	}
	initializer := statement.Initializer
	if initializer.Kind != ast.KindVariableDeclarationList || initializer.Flags&ast.NodeFlagsBlockScoped == 0 {
		return nil, lowering.notYet(initializer, "a for...of that doesn't declare its variable with const or let")
	}
	declarations := initializer.AsVariableDeclarationList().Declarations.Nodes
	if len(declarations) != 1 {
		return nil, lowering.notYet(initializer, "a for...of declaring more than one variable")
	}
	name := declarations[0].Name()
	if !ast.IsIdentifier(name) && name.Kind != ast.KindArrayBindingPattern {
		return nil, lowering.notYet(initializer, "a for...of destructuring an object")
	}
	if plan, planIterationError := lowering.planIteration(statement.Expression); planIterationError != nil {
		return nil, planIterationError
	} else if plan != nil {
		return lowering.forOfUser(node, plan, name)
	}
	// map.entries(), map.keys() and map.values() are the map itself, iterated for that part.
	iterated, mapPart := statement.Expression, ""
	if call := ast.SkipParentheses(iterated); call.Kind == ast.KindCallExpression {
		if optionalCallError := lowering.optionalCall(call); optionalCallError != nil {
			return nil, optionalCallError
		}
	}
	if call := ast.SkipParentheses(
		iterated,
	); call.Kind == ast.KindCallExpression && len(call.AsCallExpression().Arguments.Nodes) == 0 {
		if callee := ast.SkipParentheses(
			call.AsCallExpression().Expression,
		); callee.Kind == ast.KindPropertyAccessExpression {
			receiver := callee.AsPropertyAccessExpression().Expression
			if receiverType, _ := lowering.representation(
				lowering.checker.GetTypeAtLocation(receiver),
			); receiverType == ir.Map {
				switch part := callee.Name().Text(); part {
				case "entries", "keys", "values":
					iterated, mapPart = receiver, part
				}
			}
		}
	}
	iterable, expressionError := lowering.expression(iterated)
	if expressionError != nil {
		return nil, expressionError
	}
	if element, iterator := lowering.libraryIteratorElement(iterated); iterator {
		return lowering.libraryForOfIterator(node, iterable, element, name)
	}
	var element ir.Type
	switch iterable.Type() {
	case ir.Array:
		var elementTypeError error
		if element, elementTypeError = lowering.elementType(statement.Expression); elementTypeError != nil {
			return nil, elementTypeError
		}
	case ir.String:
		// A string's elements are its code points, each a string.
		element = ir.String
	case ir.Map:
		if ast.IsIdentifier(name) && (mapPart == "entries" || (mapPart == "" && !lowering.isSet(iterated))) {
			var key, value ir.Type
			var err error
			if lowering.isSet(iterated) {
				key, err = lowering.setElement(iterated)
				value = key
			} else {
				key, value, err = lowering.mapTypes(iterated)
			}
			if err != nil {
				return nil, err
			}
			iterator := ir.CollectionIterator{
				Collection: iterable,
				Part:       "entries",
				Key:        key,
				Value:      value,
				Set:        lowering.isSet(iterated),
			}
			return lowering.libraryForOfIterator(node, iterator, ir.Object, name)
		}
		return lowering.forOfMap(node, iterable, iterated, mapPart, name)
	case ir.Object:
		if held, isView := lowering.iterableViewElement(iterated); isView {
			// One kind is listed for a view today, a collection's library iterator, so there is no
			// dispatch and no default arm (Lower's condition 5 on #qk8rztp comes with a second kind):
			// the view runs that iterator's own loop, and checkIterableViews proves, once every call
			// is lowered, that nothing else reaches it.
			lowering.iterableViews = append(
				lowering.iterableViews,
				iterableView{where: iterated, iterated: iterable, held: held},
			)
			view := len(lowering.iterableViews) - 1
			statements, libraryForOfIteratorError := lowering.libraryForOfIterator(node, iterable, held.element, name)
			return lowering.viewLoop(view, statements, libraryForOfIteratorError)
		}
		if !lowering.isLibraryType(
			lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(iterated)),
			"RegExpStringIterator",
		) {
			return nil, lowering.notYet(iterated, "for...of over an object")
		}
		element = ir.Array
	case ir.Union:
		held, isView := lowering.iterableViewElement(iterated)
		if !isView {
			return nil, lowering.notYet(statement.Expression, "for...of over a "+typeName(iterable.Type()))
		}
		// Iterable<string> is held as a union, since a string fits it too. The same proof admits it
		// (Lower's ruling on #qk8rztp), and the loop reads the slot as the iterator through the checked
		// unbox, whose tag test stops the program if the proof is ever wrong. A string origin is a second
		// kind, unproven until a dispatch exists.
		lowering.iterableViews = append(
			lowering.iterableViews,
			iterableView{where: iterated, iterated: iterable, held: held},
		)
		view := len(lowering.iterableViews) - 1
		unboxed := ir.Narrow{
			Value:   iterable,
			To:      ir.Object,
			Checked: true,
			Message: "a view over an iterable held something other than a collection's iterator at " +
				lowering.program.Where(iterated),
		}
		statements, libraryForOfIteratorError := lowering.libraryForOfIterator(node, unboxed, held.element, name)
		return lowering.viewLoop(view, statements, libraryForOfIteratorError)
	default:
		return nil, lowering.notYet(statement.Expression, "for...of over a "+typeName(iterable.Type()))
	}
	lowered := ir.ForOf{Iterable: iterable, Element: element, RegexIterator: iterable.Type() == ir.Object}
	if ast.IsIdentifier(name) {
		var declareLocalError error
		if lowered.Local, declareLocalError = lowering.declareLocal(name); declareLocalError != nil {
			return nil, declareLocalError
		}
		if element == ir.Weak {
			// An array of Weak<Node> narrowed to present (by filter) still holds handles, so the
			// variable holds one too, and each read of it is the target.
			lowering.result.Locals[lowered.Local].Type = ir.Weak
		}
	} else {
		// for (const [a, b] of pairs): each name reads a field of the tuple, "0", "1", ...
		if element != ir.Object {
			return nil, lowering.notYet(name, "destructuring a "+typeName(element))
		}
		for index, binding := range name.AsBindingPattern().Elements.Nodes {
			if skipped(binding) {
				continue
			}
			bound := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || bound.Initializer != nil || bound.DotDotDotToken != nil {
				return nil, lowering.notYet(binding, "a destructured name that isn't plain")
			}
			local, declareLocalError := lowering.declareLocal(binding.Name())
			if declareLocalError != nil {
				return nil, declareLocalError
			}
			if of := lowering.result.Locals[local].Type; slotless(of) &&
				!(of == ir.Union && lowering.writable(lowering.checker.GetTypeAtLocation(binding.Name()))) {
				return nil, lowering.notYet(binding, "a tuple element of type "+typeName(of))
			}
			lowered.Pattern = append(lowered.Pattern, ir.Binding{Local: local, Field: strconv.Itoa(index)})
		}
	}
	var statementError error
	if lowered.Body, statementError = lowering.statement(statement.Statement); statementError != nil {
		return nil, statementError
	}
	return []ir.Statement{lowered}, nil
}

// forOfMap lowers for...of over a map: for (const [key, value] of map), and over map.keys() or
// map.values() into one name. An entry as one name would be a tuple made each step, which stage 0
// doesn't do yet.
func (lowering *lowering) forOfMap(
	node *ast.Node,
	iterable ir.Expression,
	iterated *ast.Node,
	part string,
	name *ast.Node,
) ([]ir.Statement, error) {
	if lowering.isSet(iterated) {
		return lowering.forOfSet(node, iterable, iterated, part, name)
	}
	key, value, mapTypesError := lowering.mapTypes(iterated)
	if mapTypesError != nil {
		return nil, mapTypesError
	}
	if part == "" {
		part = "entries"
	}
	lowered := ir.ForOf{Iterable: iterable, MapPart: part, Key: key, Value: value}
	switch {
	case part == "entries" && name.Kind == ast.KindArrayBindingPattern:
		elements := name.AsBindingPattern().Elements.Nodes
		if len(elements) > 2 {
			return nil, lowering.notYet(name, "destructuring more than a key and a value")
		}
		for index, binding := range elements {
			if skipped(binding) {
				continue
			}
			bound := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || bound.Initializer != nil || bound.DotDotDotToken != nil {
				return nil, lowering.notYet(binding, "a destructured name that isn't plain")
			}
			local, declareLocalError := lowering.declareLocal(binding.Name())
			if declareLocalError != nil {
				return nil, declareLocalError
			}
			lowered.Pattern = append(lowered.Pattern, ir.Binding{Local: local, Field: strconv.Itoa(index)})
		}
	case part != "entries" && ast.IsIdentifier(name):
		var declareLocalError error
		if lowered.Local, declareLocalError = lowering.declareLocal(name); declareLocalError != nil {
			return nil, declareLocalError
		}
	default:
		return nil, lowering.notYet(
			name,
			"for...of over a map's "+part+" into this name (write [key, value], or iterate keys() or values())",
		)
	}
	var statementError error
	if lowered.Body, statementError = lowering.statement(node.AsForInOrOfStatement().Statement); statementError != nil {
		return nil, statementError
	}
	return []ir.Statement{lowered}, nil
}

// switchStatement preserves ordered case expression evaluation.
func (lowering *lowering) switchStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsSwitchStatement()
	if enumSwitchError := lowering.enumSwitch(node); enumSwitchError != nil {
		return nil, enumSwitchError
	}
	value, expressionError := lowering.expression(statement.Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	storage, switchBindingsError := lowering.switchBindings(statement.CaseBlock)
	if switchBindingsError != nil {
		return nil, switchBindingsError
	}
	original := value
	held := -1
	if len(storage) > 0 {
		// The discriminant runs outside the case block's lexical environment.
		held = len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "switch_value", Type: value.Type(), Function: lowering.functionIndex},
		)
		value = ir.Read{Local: held, Of: value.Type()}
	}
	lowered := ir.Switch{Value: value}
	groups := []switchGroup{}
	prefix := []ir.Statement{}
	identity := lowering.enumIdentity(lowering.checker.GetTypeAtLocation(statement.Expression))
	checkedDefault := lowering.numericEnum(identity) &&
		lowering.enumSwitchCovered(node) &&
		!lowering.enumDefaultUnreachable(node)
	var neverCheck ir.Statement
	if checkedDefault {
		// The unmatched edge is never to the checker, including an implicit default. Hold the
		// scrutinee once so the check names the original value without repeating its effects.
		local := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "enum_switch_value", Type: value.Type(), Function: lowering.functionIndex},
		)
		prefix = append(prefix, ir.Declare{Local: local, Value: value})
		value = ir.Read{Local: local, Of: value.Type()}
		lowered.Value = value
		neverCheck = ir.Evaluate{Value: lowering.enumNeverCheck(statement.Expression, value, identity)}
		lowered.Default = []ir.Statement{neverCheck}
	}
	tests := []ir.Expression{}
	defaultPending := false
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		isDefault := clause.Kind == ast.KindDefaultClause
		if isDefault &&
			len(tests) == 0 &&
			lowering.enumDefaultUnreachable(node) &&
			(len(groups) == 0 || switchBodyLeaves(groups[len(groups)-1].body)) {
			continue
		}
		if !isDefault {
			test, caseLabelError := lowering.caseLabel(clause.AsCaseOrDefaultClause().Expression)
			if caseLabelError != nil {
				return nil, caseLabelError
			}

			if test.Type() != value.Type() {
				return nil, lowering.notYet(clause, "a case whose type differs from the switch's")
			}
			tests = append(tests, test)
		}
		body, statementsError := lowering.statements(clause.AsCaseOrDefaultClause().Statements.Nodes)
		if statementsError != nil {
			return nil, statementsError
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
	result := []ir.Statement{lowered}
	if checkedDefault {
		result = lowering.fallthroughSwitch(value, groups, neverCheck)
	} else {
		for index, group := range groups {
			if (group.isDefault && len(group.tests) > 0) ||
				(index+1 < len(groups) && len(group.body) > 0 && !switchBodyLeaves(group.body)) {
				result = lowering.fallthroughSwitch(value, groups, nil)
				break
			}
		}
	}
	result = append(prefix, result...)
	if held >= 0 {
		body := []ir.Statement{ir.Declare{Local: held, Value: original}}
		body = append(body, storage...)
		body = append(body, result...)
		result = []ir.Statement{ir.Block{Body: body}}
	}
	return result, nil
}

// arrayMethod lowers array.push(value) and array.join(separator).
func (lowering *lowering) arrayMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	return lowering.arrayMethodArguments(node, receiver, name, node.AsCallExpression().Arguments.Nodes)
}

func (lowering *lowering) arrayMethodArguments(
	node *ast.Node,
	receiver *ast.Node,
	name string,
	written []*ast.Node,
) (ir.Expression, bool, error) {
	if created := ast.SkipParentheses(receiver); name == "fill" &&
		created.Kind == ast.KindNewExpression &&
		lowering.isLibraryGlobal(created.AsNewExpression().Expression, "Array") {
		return lowering.newArrayFilled(node, created)
	}
	element, elementTypeError := lowering.elementType(receiver)
	if elementTypeError != nil {
		return nil, true, elementTypeError
	}
	array, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if name == "sort" {
		return lowering.arraySort(node, array, element)
	}
	if name == "map" {
		arguments := written
		if len(arguments) != 1 {
			return nil, true, lowering.notYet(node, "map with other than one callback")
		}
		callback, known, err := lowering.libraryMapCallback(arguments[0], receiver, element)
		if !known {
			callback, err = lowering.expression(arguments[0])
		}
		if err != nil {
			return nil, true, err
		}
		if callback.Type() != ir.Closure {
			return nil, true, lowering.notYet(arguments[0], "map with a callback that isn't a function")
		}
		result, elementTypeError := lowering.elementType(node)
		if elementTypeError != nil {
			return nil, true, elementTypeError
		}
		callbackType := int(lowering.concrete(lowering.checker.GetTypeAtLocation(arguments[0])).Id())
		if !known {
			var adaptArrayCallbackError error
			callback, callbackType, adaptArrayCallbackError = lowering.adaptArrayCallback(
				arguments[0],
				callback,
				[]ir.Type{element, ir.Number, ir.Array},
			)
			if adaptArrayCallbackError != nil {
				return nil, true, adaptArrayCallbackError
			}
		}
		return ir.ArrayMap{
			Array:        array,
			Callback:     callback,
			Element:      element,
			Result:       result,
			CallbackType: callbackType,
		}, true, nil
	}
	if _, isVisit := visits[name]; isVisit {
		return lowering.arrayVisit(node, array, element, name)
	}
	if name == "reduce" {
		return lowering.arrayReduce(node, array, element)
	}
	if name == "concat" && element == ir.Array {
		// TypeScript says grid.concat(row) appends row as one element; JavaScript spreads any array it's
		// given, so the types and the program would disagree.
		return nil, true, lowering.notYet(node, "concat on an array of arrays")
	}
	arguments := []ir.Expression{}
	for _, argument := range written {
		lowered, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
		}
		arguments = append(arguments, lowered)
	}
	if name == "slice" {
		// An absent bound is the start's 0 or the end's length, which +Infinity clamps to (array.c's
		// adamic_array_slice), as string slice reads one (library_string.go). A bound held as maybe a number is
		// evaluated once, in its place, and only a missing one takes its default.
		for index, argument := range arguments {
			fallback := ir.NumberConstant{Value: 0}
			if index == 1 {
				fallback.Value = math.Inf(1)
			}
			if _, absent := argument.(ir.Undefined); absent {
				arguments[index] = fallback
			} else if argument.Type() == ir.MaybeNumber {
				arguments[index] = ir.Coalesce{Value: argument, Fallback: fallback, Of: ir.Number}
			}
		}
		for _, argument := range arguments {
			if argument.Type() != ir.Number {
				return nil, true, lowering.notYet(node, "slice with an index that isn't a number")
			}
		}
		if len(arguments) > 2 {
			return nil, true, lowering.notYet(node, "slice with more than two arguments")
		}
		return ir.ArraySlice{Array: array, Arguments: arguments, Element: element}, true, nil
	}
	if name == "push" {
		if len(arguments) != 1 {
			return nil, true, lowering.notYet(node, "push with other than one value")
		}
		return ir.ArrayPush{
			Array:   array,
			Value:   fit(arguments[0], element),
			Element: element,
			Site:    lowering.writeSite(receiver),
		}, true, nil
	}
	switch name {
	case "includes", "indexOf":
		if element == ir.Union {
			return nil, true, lowering.notYet(node, "searching boxed union array elements")
		}
		if len(arguments) != 1 {
			return nil, true, lowering.notYet(node, name+" with a starting index")
		}
		if arguments[0] = fit(arguments[0], element); arguments[0].Type() != element {
			return nil, true, lowering.notYet(node, name+" with a value of another type than the elements")
		}
		return ir.ArraySearch{
			Array:    array,
			Value:    arguments[0],
			Element:  element,
			Includes: name == "includes",
		}, true, nil
	case "at":
		if len(arguments) != 1 || arguments[0].Type() != ir.Number {
			return nil, true, lowering.notYet(node, "at with other than one number")
		}
		return ir.ArrayIndex{Array: array, Index: arguments[0], Element: element, Relative: true}, true, nil
	case "reverse":
		return ir.ArrayReverse{Array: array}, true, nil
	case "fill":
		if len(arguments) > 0 {
			arguments[0] = fit(arguments[0], element)
		}
		if len(arguments) == 0 || len(arguments) > 3 || arguments[0].Type() != element {
			return nil, true, lowering.notYet(node, "fill with other than a value of the elements' type")
		}
		fill := ir.ArrayFill{Array: array, Value: arguments[0], Element: element, Site: lowering.writeSite(receiver)}
		for index, bound := range arguments[1:] {
			if bound.Type() != ir.Number {
				return nil, true, lowering.notYet(node, "fill with a bound that isn't a number")
			}
			if index == 0 {
				fill.Start = bound
			} else {
				fill.End = bound
			}
		}
		return fill, true, nil
	case "splice":
		if len(arguments) == 0 ||
			arguments[0].Type() != ir.Number ||
			(len(arguments) > 1 && arguments[1].Type() != ir.Number) {
			return nil, true, lowering.notYet(node, "splice without a start and a count that are numbers")
		}
		splice := ir.ArraySplice{
			Array:   array,
			Start:   arguments[0],
			Element: element,
			Site:    lowering.writeSite(receiver),
		}
		if len(arguments) > 1 {
			splice.Count = arguments[1]
			for _, item := range arguments[2:] {
				if item = fit(item, element); item.Type() != element {
					return nil, true, lowering.notYet(
						node,
						"splice inserting a value of another type than the elements",
					)
				}
				splice.Items = append(splice.Items, item)
			}
		}
		return splice, true, nil
	case "concat":
		for index, argument := range arguments {
			if argument.Type() != ir.Array {
				return nil, true, lowering.notYet(
					node,
					"concat with a value that isn't an array (JavaScript appends it)",
				)
			}
			if other, elementTypeError := lowering.elementType(
				written[index],
			); elementTypeError != nil || other != element {
				return nil, true, lowering.notYet(node, "concat of arrays of different elements")
			}
		}
		return ir.ArrayConcat{Array: array, Others: arguments, Element: element}, true, nil
	}
	if element != ir.Number && element != ir.Boolean && element != ir.String && element != ir.MaybeNumber {
		// JavaScript writes an object as "[object Object]", a function as its source, and an array as
		// its own join, flattened; 0.1 has no use for any of that.
		return nil, true, lowering.notYet(node, "join on an array of objects, arrays, maps or functions")
	}
	separator := ir.Expression(ir.StringConstant{Index: lowering.constant(",")})
	if len(arguments) == 1 {
		arguments[0] = lowering.orDefault(written[0], arguments[0], ",")
		if arguments[0].Type() != ir.String {
			return nil, true, lowering.notYet(node, "join with a separator that isn't a string")
		}
		separator = arguments[0]
	} else if len(arguments) > 1 {
		return nil, true, lowering.notYet(node, "join with more than one argument")
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
func (lowering *lowering) newArrayFilled(node *ast.Node, created *ast.Node) (ir.Expression, bool, error) {
	element, elementTypeError := lowering.elementType(node)
	if elementTypeError != nil {
		return nil, true, elementTypeError
	}
	constructed := created.AsNewExpression().Arguments
	filled := node.AsCallExpression().Arguments.Nodes
	if constructed == nil || len(constructed.Nodes) != 1 || len(filled) != 1 {
		return nil, true, lowering.notYet(
			node,
			"new Array filled with other than one length and one value (a part left unfilled is a hole)",
		)
	}
	length, expressionError := lowering.expression(constructed.Nodes[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	value, expressionError := lowering.expression(filled[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	if value = fit(value, element); length.Type() != ir.Number || value.Type() != element {
		return nil, true, lowering.notYet(
			node,
			"new Array(length).fill(value) with a length that isn't a number or a value of another type",
		)
	}
	return ir.ArrayFill{Length: length, Value: value, Element: element}, true, nil
}

// visits are the array methods that call a function per element and look at what it returns.
var visits = map[string]struct{}{"forEach": {}, "filter": {}, "some": {}, "every": {}, "find": {}, "findIndex": {}}

// arrayVisit lowers forEach, filter, some, every, find, findIndex, findLast and findLastIndex.
// All but forEach decide by what the callback returns, which 0.1 requires to be a boolean:
// JavaScript would take any value's truthiness there, and 0.1 has none.
func (lowering *lowering) arrayVisit(
	node *ast.Node,
	array ir.Expression,
	element ir.Type,
	name string,
) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return nil, true, lowering.notYet(node, name+" with other than one callback")
	}
	callback, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	if callback.Type() != ir.Closure {
		return nil, true, lowering.notYet(arguments[0], name+" with a callback that isn't a function")
	}
	signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(arguments[0]),
		checker.SignatureKindCall,
	)
	if len(signatures) != 1 {
		return nil, true, lowering.notYet(arguments[0], name+" with an overloaded callback")
	}
	claim := predicateOfSignature(lowering.checker, signatures[0])
	if claim != nil && claim.Type() != nil {
		// A predicate changes the receiver's flow type, never its stored elements.
		// Optional numbers and references keep their storage; boxed unions cannot become S[].
		if name == "every" && element == ir.Union {
			target, known := lowering.kept(lowering.concrete(claim.Type()))
			if known && target != element {
				return nil, true, &Refused{
					Where: lowering.program.Where(node),
					What:  "array element narrowing across a union layout is not yet sound",
					Fix:   "read the elements through the union type, or copy them into a new S[]",
				}
			}
		}
		if name == "filter" && element == ir.Union {
			target, known := lowering.kept(lowering.concrete(claim.Type()))
			if known && target != element {
				return nil, true, lowering.notYet(
					node,
					"filter predicate element representation conversion from boxed union to "+
						typeName(target)+
						"; use a loop with an explicit narrowed copy",
				)
			}
		}
		// An undefined result already fits packed optional-number storage.
		// Heap-backed and boxed sources cannot be reused as that result slot.
		if (name == "find" || name == "findLast") &&
			!(element == ir.MaybeNumber && claim.Type().Flags()&checker.TypeFlagsUndefined != 0) {
			result, typeOfError := lowering.typeOf(node)
			if typeOfError != nil || result != ir.Maybe(element) {
				return nil, true, lowering.notYet(node, name+
					" predicate result representation conversion from "+
					typeName(ir.Maybe(element))+
					" to "+
					lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(node))+
					"; use a loop with an explicit narrowed result")
			}
		}
	}
	var returns ir.Type
	if result := lowering.checker.GetReturnTypeOfSignature(signatures[0]); result.Flags()&checker.TypeFlagsVoid == 0 {
		var isKnown bool
		if returns, isKnown = lowering.representation(result); !isKnown {
			return nil, true, lowering.notYet(
				arguments[0],
				name+" with a callback returning "+lowering.checker.TypeToString(result),
			)
		}
	}
	if name != "forEach" && returns != ir.Boolean {
		return nil, true, &Refused{
			Where: lowering.program.Where(arguments[0]),
			What:  "a " + name + " callback that doesn't return a boolean",
			Fix:   "return a comparison, like word.length > 0: 0.1 has no truthiness",
		}
	}
	contract, _, arraySearchContractError := lowering.arraySearchContract(
		node,
		signatures[0],
		ir.ArrayVisit{Method: name, Array: array, Callback: callback, Element: element, Returns: returns},
	)
	if arraySearchContractError != nil {
		return nil, true, arraySearchContractError
	}
	visit := contract.(ir.ArrayVisit)
	first := element
	if visit.SearchFirst != 0 {
		first = visit.SearchFirst
	}
	callback, callbackType, adaptArrayCallbackError := lowering.adaptArrayCallback(
		arguments[0],
		callback,
		[]ir.Type{first, ir.Number, ir.Array},
	)
	if adaptArrayCallbackError != nil {
		return nil, true, adaptArrayCallbackError
	}
	visit.Callback, visit.CallbackType = callback, callbackType
	return visit, true, nil
}

// arrayReduce lowers array.reduce(callback, initial). 0.1 requires the initial value (docs/0.1.md):
// without one, an empty array throws and a one-element array returns it without a call.
func (lowering *lowering) arrayReduce(
	node *ast.Node,
	array ir.Expression,
	element ir.Type,
) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) == 1 {
		return nil, true, &Refused{
			Where: lowering.program.Where(node),
			What:  "reduce without an initial value",
			Fix:   "pass one, like reduce((sum, value) => sum + value, 0): without it, an empty array throws",
		}
	}
	if len(arguments) != 2 {
		return nil, true, lowering.notYet(node, "reduce with other than a callback and an initial value")
	}
	callback, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	if callback.Type() != ir.Closure {
		return nil, true, lowering.notYet(arguments[0], "reduce with a callback that isn't a function")
	}
	initial, expressionError := lowering.expression(arguments[1])
	if expressionError != nil {
		return nil, true, expressionError
	}
	result, typeOfError := lowering.typeOf(node)
	if typeOfError != nil {
		return nil, true, typeOfError
	}
	initial = fit(initial, result)
	if result != initial.Type() || slotless(result) {
		return nil, true, lowering.notYet(
			node,
			"reduce to a "+lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(node)),
		)
	}
	callback, callbackType, adaptArrayCallbackError := lowering.adaptArrayCallback(
		arguments[0],
		callback,
		[]ir.Type{result, element, ir.Number, ir.Array},
	)
	if adaptArrayCallbackError != nil {
		return nil, true, adaptArrayCallbackError
	}
	return ir.ArrayReduce{
		Array:        array,
		Callback:     callback,
		Initial:      initial,
		Element:      element,
		Result:       result,
		CallbackType: callbackType,
	}, true, nil
}

// mapTypes is a Map's key and value representations. 0.1's maps have string or number keys.
func (lowering *lowering) mapTypes(node *ast.Node) (ir.Type, ir.Type, error) {
	arguments := lowering.typeArguments(lowering.checker.GetTypeAtLocation(node))
	if len(arguments) != 2 {
		return 0, 0, lowering.notYet(node, "a Map whose key and value types aren't known")
	}
	key, keyKnown := lowering.representation(arguments[0])
	value, valueKnown := lowering.kept(arguments[1])
	// never has no inhabitant to store. Empty maps can still be observed through
	// size or a readonly wider view; an unused object slot supplies storage only.
	if arguments[0].Flags()&checker.TypeFlagsNever != 0 {
		key, keyKnown = ir.Object, true
	}
	if arguments[1].Flags()&checker.TypeFlagsNever != 0 {
		value, valueKnown = ir.Object, true
	}
	if !keyKnown || !keyable(key) {
		return 0, 0, lowering.notYet(
			node,
			"a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions",
		)
	}
	// number | undefined is held in a value's one slot packed (native/slots.go).
	// The existing union tag carries object, array and scalar members alike.
	// Keep unknown/nonprimitive erasure and nullable mixed layouts outside this cut.
	boxed := value == ir.Union &&
		arguments[1].Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) == 0 &&
		!lowering.includesNull(arguments[1])
	if !valueKnown || (slotless(value) && !boxed) {
		return 0, 0, lowering.notYet(node, "a Map of "+lowering.checker.TypeToString(arguments[1]))
	}
	return key, value, nil
}

// newExpression lowers new Map(), and new Map([[key, value], ...]) with its pairs written out, which
// is what the array of pairs means.
func (lowering *lowering) newExpression(node *ast.Node) (ir.Expression, error) {
	if value, found, nodeFileSystemFileDateError := lowering.nodeFileSystemFileDate(node); found {
		return value, nodeFileSystemFileDateError
	}
	if lowering.isLibraryGlobal(node.AsNewExpression().Expression, "RegExp") {
		return lowering.regexConstant(node)
	}
	created := node.AsNewExpression()
	if declaration, isClass := lowering.classes[lowering.symbol(ast.SkipParentheses(created.Expression))]; isClass {
		return lowering.construct(node, declaration)
	}
	if lowering.isLibraryGlobal(created.Expression, "Set") {
		return lowering.newSet(node)
	}
	if lowering.errorClass(node) != "" {
		return lowering.newError(node)
	}
	if !lowering.isLibraryGlobal(created.Expression, "Map") {
		if value, handled, newClassValueError := lowering.newClassValue(node); handled {
			return value, newClassValueError
		}
		return nil, lowering.notYet(node, "new "+describe(created.Expression))
	}
	key, value, mapTypesError := lowering.mapTypes(node)
	if mapTypesError != nil {
		return nil, mapTypesError
	}
	lowered := ir.MapNew{Key: key, Value: value}
	if created.Arguments == nil || len(created.Arguments.Nodes) == 0 {
		return lowered, nil
	}
	if len(created.Arguments.Nodes) != 1 {
		return nil, lowering.notYet(node, "new Map with more than one argument")
	}
	pairs := ast.SkipParentheses(created.Arguments.Nodes[0])
	if lowering.libraryEmptyCollectionArgument(pairs) {
		return lowered, nil
	}
	if !writtenOut(pairs) {
		// Pairs from anywhere else: an array of them, another Map, its entries() (collections.go).
		return lowering.newMapFrom(node, pairs, key, value)
	}
	for _, pair := range pairs.AsArrayLiteralExpression().Elements.Nodes {
		pair = ast.SkipParentheses(pair)
		if pair.Kind != ast.KindArrayLiteralExpression || len(pair.AsArrayLiteralExpression().Elements.Nodes) != 2 {
			return nil, lowering.notYet(pair, "a Map entry that isn't [key, value] written out")
		}
		elements := pair.AsArrayLiteralExpression().Elements.Nodes
		entryKey, expressionError := lowering.expression(elements[0])
		if expressionError != nil {
			return nil, expressionError
		}
		entryValue, expressionError := lowering.expression(elements[1])
		if expressionError != nil {
			return nil, expressionError
		}
		// A number, or undefined, where number | undefined goes is made that pair.
		entryKey = fit(entryKey, key)
		entryValue = fit(entryValue, value)
		if entryKey.Type() != key || entryValue.Type() != value {
			return nil, lowering.notYet(pair, "a Map entry whose key or value is of another type than the Map's")
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
func (lowering *lowering) mapMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	key, value, mapTypesError := lowering.mapTypes(receiver)
	if mapTypesError != nil {
		return nil, true, mapTypesError
	}
	object, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
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
		return nil, true, lowering.notYet(node, "map."+name+" with arguments of other types")
	}
	switch name {
	case "get":
		return ir.MapGet{Map: object, Key: arguments[0], KeyType: key, ValueType: value}, true, nil
	case "set":
		return ir.MapSet{
			Map:       object,
			Key:       arguments[0],
			Value:     arguments[1],
			KeyType:   key,
			ValueType: value,
			Site:      lowering.writeSite(receiver),
		}, true, nil
	case "has":
		return ir.MapHas{Map: object, Key: arguments[0], KeyType: key}, true, nil
	}
	return ir.MapDelete{Map: object, Key: arguments[0], KeyType: key}, true, nil
}

// stringMethod lowers string.trim() and string.charCodeAt(index).
func (lowering *lowering) stringMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	value, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if name == "trim" {
		if len(arguments) != 0 {
			return nil, true, lowering.notYet(node, "trim with arguments")
		}
		return ir.Trim{Value: value}, true, nil
	}
	index := ir.Expression(ir.NumberConstant{Value: 0})
	if len(arguments) > 1 {
		return nil, true, lowering.notYet(node, "charCodeAt with more than one argument")
	}
	if len(arguments) == 1 {
		if index, expressionError = lowering.expression(arguments[0]); expressionError != nil {
			return nil, true, expressionError
		}
		if index.Type() != ir.Number {
			return nil, true, lowering.notYet(node, "charCodeAt with an index that isn't a number")
		}
	}
	return ir.CharCodeAt{Value: value, Index: index}, true, nil
}

// shorthand lowers the value of { value }. The name there is the field's, and asked for its symbol
// the checker gives the field; the variable it reads is a separate question.
func (lowering *lowering) shorthand(property *ast.Node) (ir.Expression, error) {
	if value, known, nodeProcessValueError := lowering.nodeProcessValue(property.Name()); known {
		return value, nodeProcessValueError
	}
	symbol := lowering.checker.GetShorthandAssignmentValueSymbol(property)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = lowering.checker.GetAliasedSymbol(symbol)
	}
	if symbol != nil {
		symbol = lowering.checker.GetExportSymbolOfSymbol(symbol)
	}
	local, isLocal := lowering.locals[symbol]
	if !isLocal {
		if target, isFunction := lowering.functions[symbol]; isFunction {
			return lowering.functionValue(property, target)
		}
		return nil, lowering.notYet(property, "reading "+property.Name().Text())
	}
	lowering.touch(local)
	of := lowering.result.Locals[local].Type
	if censusFieldSlotless(of) {
		return nil, lowering.notYet(property, "a field from a "+typeName(of)+" variable")
	}
	read := ir.Read{Local: local, Of: of, Checked: lowering.checked(local)}
	if identity := lowering.enumNeverIdentity(property.Name(), map[*ast.Node]bool{}); identity != nil {
		return lowering.enumNeverCheck(property.Name(), read, identity), nil
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
	"split":       {[]ir.Type{ir.String}, 0},
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

func (lowering *lowering) stringCall(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	shape := stringMethods[name]
	value, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	written := node.AsCallExpression().Arguments.Nodes
	if name == "slice" || name == "lastIndexOf" {
		return lowering.libraryStringMethod(node, value, name, written)
	}
	if len(written) > len(shape.arguments) || len(written) < len(shape.arguments)-shape.optional {
		return nil, true, lowering.notYet(node, name+" with these arguments")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		lowered, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if fallback, isOptional := optionalStrings[name][index]; isOptional {
			lowered = lowering.orDefault(argument, lowered, fallback)
		}
		if lowered.Type() != shape.arguments[index] {
			return nil, true, lowering.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	// JavaScript's defaults: codePointAt() is position 0, and padStart's fill is a space. slice's end
	// stays missing, which is different from any number, so the emitter is told how many were given.
	switch {
	case name == "codePointAt" && len(arguments) == 0:
		arguments = append(arguments, ir.NumberConstant{Value: 0})
	case (name == "padStart" || name == "padEnd") && len(arguments) == 1:
		arguments = append(arguments, ir.StringConstant{Index: lowering.constant(" ")})
	case name == "normalize" && len(arguments) == 0:
		arguments = append(arguments, ir.StringConstant{Index: lowering.constant("NFC")})
	}
	return ir.StringCall{Method: name, Value: value, Arguments: arguments}, true, nil
}

// arraySort lowers array.sort(comparator). 0.1 requires the comparator (the default sorts numbers as
// strings), and stage 0 takes one of the module's functions by name.
func (lowering *lowering) arraySort(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) == 0 {
		return nil, true, &Refused{
			Where: lowering.program.Where(node),
			What:  "sort without a comparator",
			Fix:   "pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9]",
		}
	}
	comparator := ast.SkipParentheses(arguments[0])
	function, isFunction := lowering.functions[lowering.symbol(comparator)]
	if !ast.IsIdentifier(comparator) || !isFunction {
		// A function value: an arrow, or a variable holding one.
		callback, expressionError := lowering.expression(comparator)
		if expressionError != nil {
			return nil, true, expressionError
		}
		signatures := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeAtLocation(comparator),
			checker.SignatureKindCall,
		)
		if callback.Type() != ir.Closure || len(signatures) != 1 {
			return nil, true, lowering.notYet(comparator, "a comparator that isn't a function")
		}
		if returns, _ := lowering.representation(
			lowering.checker.GetReturnTypeOfSignature(signatures[0]),
		); returns != ir.Number {
			return nil, true, lowering.notYet(comparator, "a comparator that doesn't return a number")
		}
		callback, callbackType, adaptArrayCallbackError := lowering.adaptArrayCallback(
			comparator,
			callback,
			[]ir.Type{element, element},
		)
		if adaptArrayCallbackError != nil {
			return nil, true, adaptArrayCallbackError
		}
		return ir.ArraySort{Array: array, Callback: callback, Element: element, CallbackType: callbackType}, true, nil
	}
	declared := lowering.result.Functions[function]
	if declared.Returns == ir.Number &&
		len(declared.Parameters) == 2 &&
		(lowering.result.Locals[declared.Parameters[0]].Type != element ||
			lowering.result.Locals[declared.Parameters[1]].Type != element) {
		callback, expressionError := lowering.expression(comparator)
		if expressionError != nil {
			return nil, true, expressionError
		}
		callback, callbackType, adaptArrayCallbackError := lowering.adaptArrayCallback(
			comparator,
			callback,
			[]ir.Type{element, element},
		)
		if adaptArrayCallbackError != nil {
			return nil, true, adaptArrayCallbackError
		}
		return ir.ArraySort{Array: array, Callback: callback, Element: element, CallbackType: callbackType}, true, nil
	}
	if declared.Returns != ir.Number ||
		len(declared.Parameters) != 2 ||
		lowering.result.Locals[declared.Parameters[0]].Type != element ||
		lowering.result.Locals[declared.Parameters[1]].Type != element {
		return nil, true, lowering.notYet(comparator, "a comparator that doesn't take two elements and return a number")
	}
	return ir.ArraySort{Array: array, Comparator: function, Element: element}, true, nil
}

// elementAccess lowers array[index], string[index], and tuple[index] with a constant index: the
// tuple's field of that name.
func (lowering *lowering) elementAccess(node *ast.Node) (ir.Expression, error) {
	if value, found, nodeEnvironmentReadError := lowering.nodeEnvironmentRead(node); found {
		return value, nodeEnvironmentReadError
	}
	if value, found, nodePlatformReadError := lowering.nodePlatformRead(node); found {
		return value, nodePlatformReadError
	}
	access := node.AsElementAccessExpression()
	index := ast.SkipParentheses(access.ArgumentExpression)
	if lowering.inheritedLibraryMember(node) &&
		!lowering.iteratorResultField(access.Expression, lowering.memberSymbol(node)) &&
		!lowering.regexRuntimeProperty(access.Expression, index.Text()) {
		return nil, lowering.prototypeRead(node, index.Text())
	}
	optional := access.QuestionDotToken != nil
	if !optional && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return lowering.optionalIndexContinuation(node)
	}
	object, expressionError := lowering.expression(access.Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	if object.Type() == ir.Object && lowering.regexGroups(access.Expression) {
		if index.Kind != ast.KindStringLiteral {
			return nil, lowering.notYet(node, "a computed named-group key")
		}
		of, typeOfError := lowering.typeOf(node)
		if typeOfError != nil {
			return nil, typeOfError
		}
		if of != ir.String && of != ir.Object {
			return nil, lowering.notYet(node, "a named-group key with an Object-prototype type")
		}
		return ir.RegExpGroup{Object: object, Name: index.Text(), Of: of, Optional: optional}, nil
	}
	if optional && (object.Type() == ir.Array || object.Type() == ir.String || object.Type().IsTypedArray()) {
		return lowering.optionalIndex(node, object)
	}
	if optional && object.Type() != ir.Object {
		return nil, lowering.notYet(node, "?.[] on a "+typeName(object.Type()))
	}
	if object.Type() == ir.String {
		position, expressionError := lowering.expression(access.ArgumentExpression)
		if expressionError != nil {
			return nil, expressionError
		}
		if position.Type() != ir.Number {
			return nil, lowering.notYet(node, "a string index that isn't a number")
		}
		return ir.StringIndex{Value: object, Index: position}, nil
	}
	if object.Type() == ir.Array {
		element, readElementTypeError := lowering.readElementType(access.Expression)
		if readElementTypeError != nil {
			return nil, readElementTypeError
		}
		position, expressionError := lowering.expression(access.ArgumentExpression)
		if expressionError != nil {
			return nil, expressionError
		}
		if position.Type() != ir.Number {
			return nil, lowering.notYet(node, "an array index that isn't a number")
		}
		return lowering.defined(
			node,
			ir.ArrayIndex{Array: object, Index: position, Element: element, Optional: optional},
		), nil
	}
	if object.Type() == ir.Object {
		if value, handled, indexedTupleReadError := lowering.indexedTupleRead(node, object); handled {
			return value, indexedTupleReadError
		}
	}
	if object.Type() == ir.Object &&
		!optional &&
		!checker.IsTupleType(lowering.checker.GetTypeAtLocation(access.Expression)) {
		if value, handled, viewKeyedReadError := lowering.viewKeyedRead(node, object); handled {
			return value, viewKeyedReadError
		}
		return lowering.finiteElementRead(node, object)
	}
	// pairs[0]?.[0]: the tuple may be missing, and the read is undefined then (collections.go).
	if optional {
		return lowering.optionalTupleElement(node, object, index)
	}
	if object.Type() != ir.Object ||
		index.Kind != ast.KindNumericLiteral ||
		!checker.IsTupleType(lowering.checker.GetTypeAtLocation(access.Expression)) {
		return nil, lowering.notYet(node, describe(node))
	}
	of, typeOfError := lowering.typeOf(node)
	stored := of
	if elements := lowering.typeArguments(lowering.checker.GetTypeAtLocation(access.Expression)); typeOfError == nil {
		if position, convertError := strconv.Atoi(
			index.Text(),
		); convertError == nil && position >= 0 && position < len(elements) {
			if declared, known := lowering.representation(
				elements[position],
			); known && (declared == ir.Weak || declared == ir.Union) {
				stored = declared
			}
		}
	}
	if typeOfError != nil {
		return nil, typeOfError
	}
	if !boxedSlot(stored) {
		return nil, lowering.notYet(node, "a tuple element of type "+typeName(stored))
	}
	property := ir.Property{Object: object, Name: index.Text(), Of: stored}
	receiver := lowering.checker.GetTypeAtLocation(access.Expression)
	if position, parseError := strconv.Atoi(index.Text()); parseError == nil {
		elements := lowering.typeArguments(receiver)
		if position >= 0 && position < len(elements) {
			property.ViewTypeID = lowering.viewTypeId(elements[position])
			property.ViewContract = lowering.result.ViewContractTypes[property.ViewTypeID]
			property.ViewReceiverTypeID = lowering.viewFieldReceiverTypeID(node, receiver)
			property.ViewType = lowering.checker.TypeToString(elements[position])
			property.ViewAllowed = lowering.viewLiterals(elements[position])
			property.View = sourceExpression(node)
			property.ViewWhere = lowering.program.Where(node)
		}
	}
	if stored == ir.Union && of != stored {
		return ir.Narrow{Value: property, To: of}, nil
	}
	return property, nil
}

// setIndex lowers array[index] = value, as a statement.
func (lowering *lowering) setIndex(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	if body, handled, entriesRecordWriteError := lowering.entriesRecordWrite(target, valueNode); handled {
		return body, entriesRecordWriteError
	}
	access := target.AsElementAccessExpression()
	array, expressionError := lowering.expression(access.Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	if array.Type() != ir.Array {
		return nil, lowering.notYet(target, "assigning an element of a "+typeName(array.Type()))
	}
	element, elementTypeError := lowering.elementType(access.Expression)
	if elementTypeError != nil {
		return nil, elementTypeError
	}
	index, expressionError := lowering.expression(access.ArgumentExpression)
	if expressionError != nil {
		return nil, expressionError
	}
	if index.Type() != ir.Number {
		return nil, lowering.notYet(target, "an array index that isn't a number")
	}
	value, expressionError := lowering.expression(valueNode)
	if expressionError != nil {
		return nil, expressionError
	}
	if lowering.nodeBufferType(lowering.checker.GetTypeAtLocation(access.Expression), "Buffer") {
		if value.Type() != ir.Number {
			return nil, lowering.notYet(valueNode, "a Buffer write of a non-number")
		}
		return []ir.Statement{ir.Evaluate{Value: ir.NodeBufferCall{
			Function:  "buffer_set",
			Arguments: []ir.Expression{array, index, value},
			Returns:   ir.Number,
		}}}, nil
	}
	return []ir.Statement{ir.SetIndex{
		Array:   array,
		Index:   index,
		Value:   fit(value, element),
		Element: element,
		Site:    lowering.writeSite(access.Expression),
	}}, nil
}

// stringFromCodes lowers String.fromCharCode(...) and String.fromCodePoint(...), each argument a
// number, evaluated in order.
func (lowering *lowering) stringFromCodes(node *ast.Node, codePoints bool) (ir.Expression, bool, error) {
	return lowering.stringFromCodesArguments(node, codePoints, node.AsCallExpression().Arguments.Nodes)
}

func (lowering *lowering) stringFromCodesArguments(
	node *ast.Node,
	codePoints bool,
	written []*ast.Node,
) (ir.Expression, bool, error) {
	lowered := ir.StringFromCodes{CodePoints: codePoints}
	if hasSpread(node) {
		spread, spreadNumbersError := lowering.spreadNumbers(node)
		if spreadNumbersError != nil {
			return nil, true, spreadNumbersError
		}
		lowered.Spread = spread
		return lowered, true, nil
	}
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, lowering.notYet(
				argument,
				"a spread argument to String."+node.AsCallExpression().Expression.Name().Text(),
			)
		}
		value, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if value.Type() != ir.Number {
			return nil, true, lowering.notYet(
				argument,
				"a "+typeName(value.Type())+" argument to String."+node.AsCallExpression().Expression.Name().Text(),
			)
		}
		lowered.Codes = append(lowered.Codes, value)
	}
	return lowered, true, nil
}

// tupleType is the tuple type an array literal makes, from the checker or from where it's written
// (return ['a', 1] in a function returning [string, number]), or nil when it makes an array.
func (lowering *lowering) tupleType(node *ast.Node) *checker.Type {
	// The tuple it's written into first: [5] written into [number, string?] is that tuple, with its
	// second element undefined, though the literal's own type is [number].
	if contextual := lowering.slotType(
		node,
	); contextual != nil &&
		checker.IsTupleType(contextual) &&
		lowering.everyKnown(lowering.checker.GetTypeArguments(contextual)) {
		// (Under as const the contextual tuple's elements are any, and the literal's own type says
		// how they're held.)
		return contextual
	}
	if made := lowering.checker.GetTypeAtLocation(node); checker.IsTupleType(made) {
		return made
	}
	return nil
}

// tupleLiteral lowers [a, b] where it makes a tuple: an object whose fields are named "0", "1", ...,
// as a tuple's elements are read (elementAccess) and destructured (destructure, and for...of), each
// made what its element's type holds. A literal with as many elements as its tuple has is lowered;
// one leaving out an optional element, or spreading, is not yet.
func (lowering *lowering) tupleLiteral(node *ast.Node, tuple *checker.Type) (ir.Expression, error) {
	items := node.AsArrayLiteralExpression().Elements.Nodes
	elements := lowering.checker.GetTypeArguments(tuple)
	if len(items) > len(elements) {
		return nil, lowering.notYet(node, "a tuple literal with more values than its tuple has elements")
	}
	literal := ir.ObjectLiteral{Tuple: true}
	// An optional element left out is undefined, a field of its own, as tuple[index] reads it.
	missing := []ir.Field{}
	for index := len(items); index < len(elements); index++ {
		of, isKnown := lowering.representation(elements[index])
		if !isKnown ||
			!boxedSlot(of) ||
			!lowering.includesUndefined(elements[index]) ||
			!(of.IsMaybe() || of.IsReference()) {
			return nil, lowering.notYet(
				node,
				"a tuple literal leaving out an element of type "+lowering.checker.TypeToString(elements[index]),
			)
		}
		missing = append(missing, ir.Field{Name: strconv.Itoa(index), Value: fit(ir.Undefined{}, of)})
	}
	for index, item := range items {
		if item.Kind == ast.KindSpreadElement || item.Kind == ast.KindOmittedExpression {
			return nil, lowering.notYet(item, describe(item)+" in a tuple literal")
		}
		of, isKnown := lowering.representation(elements[index])
		if !isKnown || !boxedSlot(of) {
			return nil, lowering.notYet(item, "a tuple element of type "+lowering.checker.TypeToString(elements[index]))
		}
		value, expressionError := lowering.expression(item)
		if expressionError != nil {
			return nil, expressionError
		}
		if value = fit(value, of); value.Type() != of {
			if _, isUndefined := value.(ir.Undefined); !isUndefined || !of.IsReference() {
				return nil, lowering.notYet(item, "a tuple element of another type than its tuple's")
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
func (lowering *lowering) spreadNumbers(call *ast.Node) (ir.Expression, error) {
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
					return nil, lowering.notYet(element, describe(element)+" in an array spread into a call")
				}
				value, expressionError := lowering.expression(element)
				if expressionError != nil {
					return nil, expressionError
				}
				if value.Type() != ir.Number {
					return nil, lowering.notYet(element, "a "+typeName(value.Type())+" argument where numbers go")
				}
				literal.Elements = append(literal.Elements, value)
				literal.Spread = append(literal.Spread, false)
			}
			continue
		}
		value, expressionError := lowering.expression(item)
		if expressionError != nil {
			return nil, expressionError
		}
		if spread {
			if element, elementTypeError := lowering.elementType(
				item,
			); elementTypeError != nil || value.Type() != ir.Array || element != ir.Number {
				return nil, lowering.notYet(argument, "spreading other than an array of numbers into a call")
			}
		} else if value.Type() != ir.Number {
			return nil, lowering.notYet(argument, "a "+typeName(value.Type())+" argument where numbers go")
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
func (lowering *lowering) tupleWhereArrayGoes(value *checker.Type, target *checker.Type, depth int) bool {
	if value == nil || target == nil || value == target || depth > 4 {
		return false
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		// Into a union: fine if any member takes it as it is.
		for _, member := range target.Types() {
			if !lowering.tupleWhereArrayGoes(value, member, depth+1) {
				return false
			}
		}
		return true
	}
	if value.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range value.Types() {
			if lowering.tupleWhereArrayGoes(member, target, depth+1) {
				return true
			}
		}
		return false
	}
	valueTuple, targetTuple := checker.IsTupleType(value), checker.IsTupleType(target)
	switch {
	case valueTuple && !targetTuple && lowering.isArrayType(target):
		return true
	case valueTuple && targetTuple:
		return lowering.pairwiseTupleWhereArrayGoes(
			lowering.checker.GetTypeArguments(value),
			lowering.checker.GetTypeArguments(target),
			depth,
		)
	case lowering.isArrayType(value) && lowering.isArrayType(target):
		return lowering.pairwiseTupleWhereArrayGoes(
			lowering.checker.GetTypeArguments(value),
			lowering.checker.GetTypeArguments(target),
			depth,
		)
	case lowering.isLibraryType(value, "Map", "ReadonlyMap") && lowering.isLibraryType(target, "Map", "ReadonlyMap"):
		return lowering.pairwiseTupleWhereArrayGoes(
			lowering.checker.GetTypeArguments(value),
			lowering.checker.GetTypeArguments(target),
			depth,
		)
	}
	if value.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	valueSignatures := lowering.checker.GetSignaturesOfType(value, checker.SignatureKindCall)
	targetSignatures := lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(valueSignatures) == 1 && len(targetSignatures) == 1 {
		// A function: what it returns goes where the target's result goes, and what the target is
		// called with goes to its parameters.
		if lowering.tupleWhereArrayGoes(
			lowering.checker.GetReturnTypeOfSignature(valueSignatures[0]),
			lowering.checker.GetReturnTypeOfSignature(targetSignatures[0]),
			depth+1,
		) {
			return true
		}
		valueParameters, targetParameters := valueSignatures[0].Parameters(), targetSignatures[0].Parameters()
		if lowering.censusNeverRestSignature(targetSignatures[0]) {
			targetParameters = nil
		}
		for index := range valueParameters {
			if index < len(targetParameters) && lowering.tupleWhereArrayGoes(
				lowering.checker.GetTypeOfSymbol(targetParameters[index]),
				lowering.checker.GetTypeOfSymbol(valueParameters[index]),
				depth+1,
			) {
				return true
			}
		}
		return false
	}
	for _, field := range lowering.checker.GetPropertiesOfType(target) {
		if lowering.tupleWhereArrayGoes(
			lowering.checker.GetTypeOfPropertyOfType(value, field.Name),
			lowering.checker.GetTypeOfSymbol(field),
			depth+1,
		) {
			return true
		}
	}
	return false
}

func (lowering *lowering) pairwiseTupleWhereArrayGoes(values []*checker.Type, targets []*checker.Type, depth int) bool {
	for index := range values {
		if index < len(targets) && lowering.tupleWhereArrayGoes(values[index], targets[index], depth+1) {
			return true
		}
	}
	return false
}

// updateIndex lowers array[index] op= value. The array and index are held before the element is
// read, and that read is held before the right side runs, exactly once each and in JavaScript's
// order. The final SetIndex uses those same held values even when the right side changes their source.
func (lowering *lowering) updateIndex(
	node *ast.Node,
	target *ast.Node,
	operator ast.Kind,
	valueNode *ast.Node,
) ([]ir.Statement, error) {
	if lowering.nodeBufferType(
		lowering.checker.GetTypeAtLocation(target.AsElementAccessExpression().Expression),
		"Buffer",
	) {
		return nil, lowering.notYet(node, "a compound Buffer write outside the census")
	}
	access := target.AsElementAccessExpression()
	array, expressionError := lowering.expression(access.Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	if array.Type() != ir.Array {
		return nil, lowering.notYet(target, "assigning an element of a "+typeName(array.Type()))
	}
	element, elementTypeError := lowering.elementType(access.Expression)
	if elementTypeError != nil {
		return nil, elementTypeError
	}
	if element != ir.Number {
		return nil, lowering.notYet(target, "updating an array element of type "+typeName(element))
	}
	index, expressionError := lowering.expression(access.ArgumentExpression)
	if expressionError != nil {
		return nil, expressionError
	}
	if index.Type() != ir.Number {
		return nil, lowering.notYet(target, "an array index that isn't a number")
	}

	arrayLocal := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "array", Type: ir.Array, Function: lowering.functionIndex},
	)
	indexLocal := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "index", Type: ir.Number, Function: lowering.functionIndex},
	)
	currentLocal := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "element", Type: ir.Number, Function: lowering.functionIndex},
	)
	arrayRead := ir.Read{Local: arrayLocal, Of: ir.Array}
	indexRead := ir.Read{Local: indexLocal, Of: ir.Number}
	current := ir.Unwrap{Value: ir.ArrayIndex{Array: arrayRead, Index: indexRead, Element: ir.Number}}
	right, expressionError := lowering.expression(valueNode)
	if expressionError != nil {
		return nil, expressionError
	}
	updated, combineError := lowering.combine(node, operator, ir.Read{Local: currentLocal, Of: ir.Number}, right)
	if combineError != nil {
		return nil, combineError
	}
	return []ir.Statement{ir.Block{Body: []ir.Statement{
		ir.Declare{Local: arrayLocal, Value: array},
		ir.Declare{Local: indexLocal, Value: index},
		ir.Declare{Local: currentLocal, Value: current},
		ir.SetIndex{
			Array:   arrayRead,
			Index:   indexRead,
			Value:   updated,
			Element: ir.Number,
			Site:    lowering.writeSite(access.Expression),
		},
	}}}, nil
}

// isCallee reports whether a property read is the expression a call calls, parentheses aside.
func isCallee(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.Parent != nil &&
		node.Parent.Kind == ast.KindCallExpression &&
		node.Parent.AsCallExpression().Expression == node
}
