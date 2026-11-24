package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) objectCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	return l.objectCallArguments(node, name, node.AsCallExpression().Arguments.Nodes)
}

func (l *lowering) objectCallArguments(node *ast.Node, name string, written []*ast.Node) (ir.Expression, bool, error) {
	// Class whole readers before record dispatch or a shape refusal can consume the call.
	reader := ir.ObjectCall{Method: name, Checked: true}
	class := viewReadOf(l.result, reader)
	for _, argument := range written {
		l.recordRegExpServedWhole(node, l.checker.GetTypeAtLocation(argument), reader, class == viewReadsNoField)
	}
	if value, handled, err := l.recordObjectCallArguments(node, name, written); handled {
		return value, true, err
	}
	refused := func(reason string) (ir.Expression, bool, error) {
		return nil, true, &Refused{Where: l.program.Where(node), What: "Object." + name, Fix: reason}
	}
	switch name {
	case "defineProperty", "defineProperties", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors":
		return refused("property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores")
	case "getPrototypeOf", "setPrototypeOf", "create":
		return refused("prototypes expose or replace fields outside the declared shape; use a declared object or class with composition")
	case "fromEntries":
		return refused("tsc returns an index-signature object with unproven keys; Adamic fixes object shapes and refuses index signatures; use Map")
	case "groupBy":
		return nil, true, l.notYet(node, "Object.groupBy's partial record with dynamically present keys (use Map and an explicitly typed grouping loop)")
	}
	count := 1
	if name == "is" || name == "hasOwn" {
		count = 2
	}
	if name == "assign" {
		if len(written) < 1 {
			return nil, true, l.notYet(node, "Object.assign without a target")
		}
	} else if len(written) != count {
		return nil, true, l.notYet(node, "Object."+name+" with these arguments")
	}
	if name == "freeze" {
		proven := l.checker.GetTypeAtLocation(written[0])
		if proven.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike|checker.TypeFlagsUndefined) != 0 {
			value, err := l.expression(written[0])
			return value, true, err
		}
	}
	call := ir.ObjectCall{Method: name, Returns: ir.Boolean}
	switch name {
	case "is":
		for _, argument := range written {
			value, err := l.expression(argument)
			if err != nil {
				return nil, true, err
			}
			call.Arguments = append(call.Arguments, fit(value, ir.Union))
		}
	case "isFrozen":
		value, err := l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		call.Arguments = []ir.Expression{fit(value, ir.Union)}
	case "keys", "values", "entries", "freeze", "hasOwn", "assign":
		if name == "keys" && !l.includesUndefined(l.checker.GetTypeAtLocation(written[0])) {
			if _, _, known := l.nodeArrayLayoutOf(l.checker.GetTypeAtLocation(written[0])); known {
				value, err := l.expression(written[0])
				if err != nil {
					return nil, true, err
				}
				call.Arguments = []ir.Expression{value}
				call.Returns = ir.Array
				return call, true, nil
			}
		}
		// Assignment and freezing retain their existing exact-shape requirement.
		// A plain const's literal initializer proves the complete shape, including field presence.
		shape := ast.SkipParentheses(written[0])
		if shape.Kind == ast.KindAsExpression {
			assertion := shape.AsAsExpression()
			if assertion.Type.Kind == ast.KindTypeReference && assertion.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
				shape = ast.SkipParentheses(assertion.Expression)
			}
		}
		if name == "keys" || name == "values" || name == "entries" {
			// As in objectKeys: Node enumerates none of an Error's own properties.
			if !l.exactObject(shape, 0) && l.errorMayHide(written[0]) {
				return nil, true, l.notYet(node, "Object."+name+" of a value that may hold an Error, whose own properties Node doesn't enumerate")
			}
			if err := l.enumerationDescriptors(node, written[0], false, false); err != nil {
				return nil, true, err
			}
		}
		if name != "keys" && name != "values" && name != "entries" && !l.exactObject(shape, 0) && !(name == "hasOwn" && isClassInstance(l.checker.GetTypeAtLocation(written[0]))) {
			return nil, true, l.notYet(written[0], "Object."+name+" on a shape not proven by a plain literal or its const binding")
		}
		value, err := l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		if name == "keys" && value.Type() == ir.Array {
			if _, _, known := l.nodeArrayLayoutOf(l.checker.GetTypeAtLocation(written[0])); known {
				call.Arguments = []ir.Expression{value}
				call.Returns = ir.Array
				return call, true, nil
			}
		}
		if value.Type() != ir.Object || l.includesUndefined(l.checker.GetTypeAtLocation(written[0])) || checker.IsTupleType(l.checker.GetTypeAtLocation(written[0])) {
			return nil, true, l.notYet(written[0], "Object."+name+" on other than a present plain object")
		}
		call.Arguments = []ir.Expression{value}
		switch name {
		case "hasOwn":
			key := ast.SkipParentheses(written[1])
			if key.Kind != ast.KindStringLiteral || !l.hasProperty(written[0], key.Text()) || len(key.Text()) > 0 && key.Text()[0] == '#' {
				return refused("hasOwn requires a string literal naming a declared public field or method; use Map for arbitrary keys")
			}
			keyValue, err := l.expression(key)
			if err != nil {
				return nil, true, err
			}
			call.Arguments = append(call.Arguments, keyValue)
		case "freeze":
			call.Returns = ir.Object
		case "keys":
			if l.decodedResultValue(written[0]) {
				return ir.ObjectKeys{Object: value}, true, nil
			}
			call.Returns = ir.Array
		case "values", "entries":
			result := l.checker.GetTypeAtLocation(node)
			// The interface overload may return any. Its destination supplies the
			// contract we must prove or check, never permission to erase it.
			if contextual := l.slotType(node); contextual != nil && l.isArrayType(contextual) {
				result = contextual
			}
			arguments := l.checker.GetTypeArguments(result)
			if len(arguments) != 1 {
				return refused("tsc's result must be an array with a proven element type")
			}
			element := arguments[0]
			if name == "entries" {
				// The destination's element can be anything the pairs are written into, unknown or a
				// union of tuples, and only a reference has type arguments: the checker dereferences nil
				// asking anything else for them (#egxccg8). It isn't read through a Weak: entries makes
				// strong pairs, and tsc lets them into an element of Weak<[string, V]>.
				var pair []*checker.Type
				if element.Flags()&checker.TypeFlagsObject != 0 && element.ObjectFlags()&checker.ObjectFlagsReference != 0 {
					pair = l.checker.GetTypeArguments(element)
				}
				if len(pair) != 2 {
					return refused("tsc's entries must have a proven value type")
				}
				element = pair[1]
			}
			call.Element, _ = l.representation(element)
			if call.Element != ir.Number && call.Element != ir.String && call.Element != ir.Boolean {
				return refused("tsc's result must have one homogeneous number, string or boolean value type; any and widened field views are unsound")
			}
			if declaration := l.enumObject(written[0]); declaration != nil {
				fields, err := l.enumFields(declaration)
				if err != nil {
					return nil, true, err
				}
				for _, field := range fields {
					if field.Value.Type() != call.Element {
						return refused("numeric enum reverse mappings add string values; use Object.keys or read members individually")
					}
				}
			}
			for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(written[0])) {
				of, known := l.representation(l.checker.GetTypeOfSymbol(field))
				if !known || of != call.Element || field.Flags&ast.SymbolFlagsOptional != 0 {
					return refused("every present field must have tsc's result element representation; optional or heterogeneous fields cannot be read soundly")
				}
			}
			call.Checked = !l.enumerationProven(written[0], element, 0)
			call.ElementName = l.checker.TypeToString(element)
			call.Allowed = l.viewLiterals(element)
			if l.openNumericEnumType(element) {
				call.Allowed = nil
			}
			if call.Checked && !interfaceScalar(element) {
				return nil, true, l.notYet(node, "checked Object enumeration of a non-primitive value contract")
			}
			if call.Checked {
				if err := l.enumerationDescriptors(node, written[0], false, true); err != nil {
					return nil, true, err
				}
			}
			call.Returns = ir.Array
		case "assign":
			if l.checker.GetTypeAtLocation(node).Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
				return refused("tsc's result is any or unknown, not a proven object type; use one to three typed sources")
			}
			target := l.checker.GetTypeAtLocation(written[0])
			for _, argument := range written[1:] {
				if !l.exactObject(argument, 0) {
					return refused("each source must have its complete shape proven by a literal or const binding; a widened source can hide overwriting fields")
				}
				source := l.checker.GetTypeAtLocation(argument)
				for _, field := range l.checker.GetPropertiesOfType(source) {
					into := l.checker.GetPropertyOfType(target, field.Name)
					if into == nil {
						return nil, true, l.notYet(argument, "Object.assign adding a field to its target's fixed shape")
					}
					fromType, toType := l.checker.GetTypeOfSymbol(field), l.checker.GetTypeOfSymbol(into)
					of, known := l.representation(fromType)
					if !known || (of != ir.Number && of != ir.Boolean && of != ir.String) || !l.enumAssignable(fromType, toType) || !l.enumAssignable(toType, fromType) || !l.checker.IsTypeAssignableTo(fromType, toType) || !l.checker.IsTypeAssignableTo(toType, fromType) {
						return refused("source and target field types must agree in both directions with tsc's intersection result; widening, conflicting fields and reference cycles are refused")
					}
				}
				lowered, err := l.expression(argument)
				if err != nil {
					return nil, true, err
				}
				l.wholeObjectCopies = append(l.wholeObjectCopies, wholeObjectCopy{node: argument, source: lowered, copied: l.concrete(source)})
				call.Arguments = append(call.Arguments, lowered)
			}
			// Object.assign returns its target, so its result is admitted only as the target's own type: tsc's T & U
			// must claim nothing the target doesn't already hold (Adamic's ruling on #ny0c3gb, #k8aak04).
			if !l.checker.IsTypeAssignableTo(target, l.checker.GetTypeAtLocation(node)) {
				return refused("Object.assign returns its target, and tsc's result type claims more than the target's own type holds")
			}
			call.Returns = ir.Object
		}
	default:
		return nil, true, l.notYet(node, "Object."+name)
	}
	if call.Checked {
		l.graphDynamicSite = node
	}
	return call, true, nil
}

// exactObject proves there are no hidden fields and no synthetic absent slots. An explicit type
// annotation, alias, spread or call would need a separate proof and is deliberately not guessed.
func (l *lowering) exactObject(node *ast.Node, depth int) bool {
	if depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	if l.enumObject(node) != nil {
		return true
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
			if field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment {
				return false
			}
			if field.Name().Kind != ast.KindIdentifier && field.Name().Kind != ast.KindStringLiteral {
				return false
			}
			if strings.ContainsRune(field.Name().Text(), 0) || field.Name().Text() == "__proto__" || len(field.Name().Text()) > 0 && field.Name().Text()[0] == '#' {
				return false
			}
		}
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Type != nil || variable.Initializer == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	return l.exactObject(variable.Initializer, depth+1)
}

// objectCanFreeze is conservative across aliases and calls: any Object.freeze in the program makes
// every object write one that can throw Node's TypeError (ir.Program's ObjectsMayFreeze).
func (l *lowering) objectCanFreeze() bool {
	found := false
	walk(l.result.Main, func(node any) bool {
		if call, ok := node.(ir.ObjectCall); ok && call.Method == "freeze" {
			found = true
		}
		return !found
	})
	for _, function := range l.result.Functions {
		walk(function.Body, func(node any) bool {
			if call, ok := node.(ir.ObjectCall); ok && call.Method == "freeze" {
				found = true
			}
			return !found
		})
	}
	return found
}

// objectIntersection represents tsc's intersection of plain object views. Every constituent must
// be a plain shape, as a checked view's intersection members are: not an array, not callable or
// constructable, and of no family of its own (a class instance, an index signature's dictionary, a
// collection, a tuple). Every resulting field must have a storage a record would give it: an
// object, array or boxed union field is held as it is in any object, and a Weak is not
// (recordStorageType). An optional boolean, a tagged byte the dictionary's rule refuses, is held as
// any plain object holds it where every member that has the field holds it that way too
// (intersectionMembersHoldMaybeBoolean), and stops otherwise. A field held as a closure waits on its
// own ruling. A field that holds the intersection itself is the object being proven.
func (l *lowering) objectIntersection(proven *checker.Type) (ir.Type, bool, string) {
	for _, part := range proven.Types() {
		kind := ""
		switch {
		case part.Flags()&checker.TypeFlagsTypeParameter != 0:
			kind = "a type parameter"
		case part.Flags()&checker.TypeFlagsObject == 0:
			kind = "a primitive"
		case l.isArrayType(part):
			kind = "an array"
		case len(l.checker.GetSignaturesOfType(part, checker.SignatureKindCall)) != 0 ||
			len(l.checker.GetSignaturesOfType(part, checker.SignatureKindConstruct)) != 0:
			kind = "a function"
		default:
			switch l.unsupportedViewFamily(part) {
			case "nominal class":
				kind = "a class instance"
			case "dictionary":
				kind = "a dictionary"
			case "collection":
				kind = "a collection"
			case "tuple":
				kind = "a tuple"
			case "":
			default:
				// Every other family still refuses; it only lacks its own words.
				kind = "a " + l.unsupportedViewFamily(part) + " view"
			}
		}
		if kind != "" {
			return 0, false, "an intersection whose member " + l.checker.TypeToString(part) + " isn't a plain object shape (" + kind + ")"
		}
	}
	// A field that holds an intersection already being proven ends there, assumed held. The assumption is
	// noted at its depth, so no answer that leaned on it is kept past the walk that made it.
	if depth, proving := l.objectIntersections[proven]; proving {
		l.objectIntersectionLow = min(l.objectIntersectionLow, depth)
		return ir.Object, true, ""
	}
	key := objectIntersectionKey{proven: proven, mapper: l.typeMapper}
	if verdict, settled := l.objectIntersectionVerdicts[key]; settled {
		return verdict.held, verdict.known, verdict.reason
	}
	depth := len(l.objectIntersections)
	// A chain that makes a new intersection at each level (a generic instantiated deeper through its fields)
	// meets no intersection twice, so it's bounded by depth instead, failing closed (#vp9f4yd).
	if depth >= objectIntersectionDepthBound {
		return 0, false, "an intersection nested past the depth bound of " + strconv.Itoa(objectIntersectionDepthBound)
	}
	if l.objectIntersections == nil {
		l.objectIntersections = map[*checker.Type]int{}
	}
	l.objectIntersections[proven] = depth
	outerLow := l.objectIntersectionLow
	l.objectIntersectionLow = depth + 1
	held, known, reason := func() (ir.Type, bool, string) {
		defer delete(l.objectIntersections, proven)
		return l.objectIntersectionFields(proven)
	}()
	// Only a walk that leaned on nothing above it is settled when it succeeds: one that met only itself again
	// is its own proof. A failure stands whatever it assumed, since an assumption only ever lets a field pass.
	if !known || l.objectIntersectionLow >= depth {
		if l.objectIntersectionVerdicts == nil {
			l.objectIntersectionVerdicts = map[objectIntersectionKey]objectIntersectionVerdict{}
		}
		l.objectIntersectionVerdicts[key] = objectIntersectionVerdict{held: held, known: known, reason: reason}
	}
	l.objectIntersectionLow = min(outerLow, l.objectIntersectionLow)
	return held, known, reason
}

// objectIntersectionDepthBound is how many intersections deep objectIntersection proves fields, each inside
// a field of the one before.
const objectIntersectionDepthBound = 64

// objectIntersectionKey is what objectIntersection's answer is a function of: the intersection, and the
// type mapper its fields' type parameters are read through.
type objectIntersectionKey struct {
	proven *checker.Type
	mapper *typeMapper
}

// objectIntersectionVerdict is one settled answer of objectIntersection.
type objectIntersectionVerdict struct {
	held   ir.Type
	known  bool
	reason string
}

// objectIntersectionFields is objectIntersection's walk of proven's fields, each held as a record would hold
// it.
func (l *lowering) objectIntersectionFields(proven *checker.Type) (ir.Type, bool, string) {
	for _, field := range l.checker.GetPropertiesOfType(proven) {
		fieldType := l.checker.GetTypeOfSymbol(field)
		if held, known := l.recordStorageType(fieldType); !known || held == ir.Closure {
			if plain, plainKnown := l.representation(fieldType); held == ir.MaybeBoolean && plainKnown && plain == ir.MaybeBoolean && l.intersectionMembersHoldMaybeBoolean(proven, field.Name) {
				continue
			}
			kind := "of a type with no representation"
			switch held {
			case ir.MaybeBoolean:
				kind = "an optional boolean"
			case ir.Closure:
				kind = "a function"
			case ir.Weak:
				kind = "a weak reference"
			default:
				if _, _, reason := l.representationReason(fieldType); reason != "" {
					return 0, false, "an intersection whose field " + field.Name + " has no representation: " + reason
				}
			}
			return 0, false, "an intersection whose field " + field.Name + " is " + kind + ", which its record storage doesn't hold"
		}
	}
	return ir.Object, true, ""
}

// intersectionMembersHoldMaybeBoolean reports whether every member of the intersection proven that has
// the field name holds it as an optional boolean by the plain object rule, the one a lone
// `{ valid?: boolean }` is held by (typeOfSymbol), and at least one does (#tq6yzqy). The member's
// property is the checker's, so a field a member inherits through its base chain counts, as tsc's
// NumericLiteral has isUnterminated from LiteralLikeNode. A member held otherwise (a plain boolean, a
// union, a reference, unknown) answers no, and the intersection keeps its stop. An intersection's
// property always comes from a member, so finding none fails closed.
//
// It is sound because a read of the field reads the slot as a tagged byte, false 0, true 1, absent 2
// (slots.go's slotted), and adamic_object_maybe_boolean (object.c) unpacks a plain boolean's byte as
// the same present false or true (maybe.c's adamic_maybe_boolean_unpack), so an object reaching the
// intersection whatever plain type built it reads as Node reads it. The one road that could put an
// absent byte where a plain boolean is read, an object built with a plain boolean written undefined
// through a type that makes the field optional, is refused before any intersection by the widening
// rule (refuseWidening, adamic/invariant-mutable), on a plain alias, a class and an intersection
// alike. A dictionary seen as the intersection, a finite Partial<Record<K, V>> declaring each field
// included, stops at the relation, since a record and an intersection held as a fixed object have
// different storage (recordStorageView, fixedObjectIntersection).
func (l *lowering) intersectionMembersHoldMaybeBoolean(proven *checker.Type, name string) bool {
	held := false
	for _, part := range proven.Types() {
		property := l.checker.GetPropertyOfType(part, name)
		if property == nil {
			continue
		}
		if stored, known := l.representation(l.checker.GetTypeOfSymbol(property)); !known || stored != ir.MaybeBoolean {
			return false
		}
		held = true
	}
	return held
}

// objectAssignCallableSourceRefusal refuses an Object.assign whose source is callable or constructable
// (Adamic's ruling on #ny0c3gb, #k8aak04, from rulings.md:86 and :293). lib types the result T & U, a claim the
// call never proves when U carries a call or construct signature: Object.assign copies own properties, never
// callability, so a record typed as a function would be called where Node throws TypeError. It runs in the
// refusal walk, ahead of every lowering, so it never rests on an intersection value staying not yet. The call is
// matched by the lib declaration it resolves to, never by its spelling, so an alias, Object["assign"] and
// globalThis.Object.assign are judged with Object.assign.
func (l *lowering) objectAssignCallableSourceRefusal(node *ast.Node) error {
	if node.Kind != ast.KindCallExpression || !l.objectAssignCall(node) {
		return nil
	}
	for index, argument := range node.Arguments() {
		if index == 0 {
			continue
		}
		source := argument
		sources := []*checker.Type{l.checker.GetTypeAtLocation(argument)}
		if argument.Kind == ast.KindSpreadElement {
			source = argument.AsSpreadElement().Expression
			sources = l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(source))
			// A spread whose element types can't be read could hold a function: it fails closed.
			if len(sources) == 0 {
				return &Refused{Where: l.program.Where(argument), What: "Object.assign with a spread source whose element types aren't known", Fix: "Object.assign copies own properties, never callability; pass each source as its own argument"}
			}
		}
		for _, proven := range sources {
			if l.callableOrConstructable(proven, map[*checker.Type]bool{}) {
				return &Refused{Where: l.program.Where(argument), What: "Object.assign with the callable source " + l.checker.TypeToString(proven), Fix: "Object.assign copies own properties, never callability, so its result is no function though tsc types it as one; keep the function as its own value"}
			}
		}
	}
	return nil
}

// objectAssignCall reports whether a call resolves to lib's ObjectConstructor.assign.
func (l *lowering) objectAssignCall(call *ast.Node) bool {
	signature := l.checker.GetResolvedSignature(call)
	if signature == nil || signature.Declaration() == nil {
		return false
	}
	declaration := signature.Declaration()
	// A method's name may be computed ([Symbol.iterator]), whose node has no text: only an identifier is asked.
	if declaration.Kind != ast.KindMethodSignature || !ast.IsIdentifier(declaration.Name()) || declaration.Name().Text() != "assign" || !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
		return false
	}
	owner := declaration.Parent
	return owner != nil && owner.Kind == ast.KindInterfaceDeclaration && ast.IsIdentifier(owner.Name()) && owner.Name().Text() == "ObjectConstructor"
}

// callableOrConstructable reports whether any part of a type has a call or construct signature, or is lib's
// Function: a union is callable when any member is, since the source may be that member, and a type
// parameter or another instantiable type is judged by its base constraint, whose union a type parameter's
// own signatures would read only in common (Compiler's read of #k8aak04).
func (l *lowering) callableOrConstructable(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	seen[proven] = true
	if proven.Flags()&checker.TypeFlagsInstantiable != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil && constraint != proven && l.callableOrConstructable(constraint, seen) {
			return true
		}
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, part := range proven.Types() {
			if l.callableOrConstructable(part, seen) {
				return true
			}
		}
		return false
	}
	return l.isLibraryType(proven, "Function", "CallableFunction", "NewableFunction") || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0
}

func init() {
	refusalRules.register(160, "objectAssignCallableSourceRefusal", (*lowering).objectAssignCallableSourceRefusal)
}
