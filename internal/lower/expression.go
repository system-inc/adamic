package lower

import (
	"errors"
	"reflect"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// typeOf is what's left at runtime of the type the checker proved for a node: a number, a boolean or
// a string. A union counts when every member is the same one ('Fizz' | 'Buzz' is a string).
func (lowering *lowering) typeOf(node *ast.Node) (ir.Type, error) {
	if ast.IsDeclarationName(node) && lowering.placeholderOrigin(node) != "" {
		if declared, _ := lowering.representation(
			lowering.checker.GetTypeOfSymbol(lowering.symbol(node)),
		); declared == ir.Weak {
			return 0, lowering.notYet(
				node,
				"a placeholder slot holding Weak; its target-read semantics need a separate checked boundary",
			)
		}
		if stored, known := lowering.representation(
			lowering.checker.GetNonNullableType(lowering.checker.GetTypeOfSymbol(lowering.symbol(node))),
		); known {
			if stored == ir.Weak {
				return 0, lowering.notYet(
					node,
					"a placeholder slot holding Weak; its target-read semantics need a separate checked boundary",
				)
			}
			return placeholderStorage(stored), nil
		}
	}

	if lowering.enumNeverIdentity(node, map[*ast.Node]bool{}) != nil {
		if symbol := lowering.flagValueSymbol(ast.SkipParentheses(node)); symbol != nil {
			if stored, known := lowering.representation(lowering.checker.GetTypeOfSymbol(symbol)); known {
				return stored, nil
			}
		}
		return ir.Number, nil
	}
	if ast.IsIdentifier(node) && lowering.exactPlainObject(node) {
		return ir.Object, nil
	}
	if node.Kind == ast.KindPropertyAccessExpression &&
		lowering.checker.GetTypeAtLocation(node).Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
		// A field narrowed to a nullish value still occupies its declared slot, which may
		// hold a packed optional scalar or a tagged reference. Read that representation.
		if field := lowering.checker.GetSymbolAtLocation(node.Name()); field != nil {
			return lowering.typeOfSymbol(node, field)
		}
	}
	if lowering.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 &&
		(node.Kind == ast.KindVoidExpression || voidInitializer(node)) {
		return ir.Object, nil
	}
	if valueType, isKnown := lowering.representation(lowering.checker.GetTypeAtLocation(node)); isKnown {
		return valueType, nil
	}
	return 0, lowering.notYet(
		node,
		"a value of type "+lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(node)),
	)
}

func (lowering *lowering) representation(proven *checker.Type) (ir.Type, bool) {
	proven = lowering.concrete(proven)
	if _, _, known := lowering.nodeArrayLayoutOf(proven); known {
		return ir.Array, true
	}
	if kind := lowering.typedArrayKind(proven); kind != 0 {
		return kind, true
	}
	if lowering.isLibraryType(proven, "Function") {
		// Opaque callable values keep closure storage; the refusal pass forbids invocation.
		return ir.Closure, true
	}
	flags := proven.Flags()
	if flags&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0 {
		return ir.Union, true
	}
	if flags&checker.TypeFlagsIntersection != 0 {
		for _, part := range proven.Types() {
			if part.Flags()&checker.TypeFlagsNonPrimitive != 0 {
				return ir.Union, true
			}
		}
	}
	if lowering.nodeBufferType(proven, "Buffer") {
		return ir.Array, true
	}
	if flags&checker.TypeFlagsObject != 0 &&
		(lowering.enumerationRecordStorageType(proven) || lowering.enumerationReadonlyType(proven)) {
		return ir.Object, true
	}
	if lowering.finitePartialRecordElement(proven) != nil {
		return ir.Record, true
	}
	if flags&checker.TypeFlagsObject != 0 &&
		enumObjectSymbol(proven) == nil &&
		len(lowering.checker.GetIndexInfosOfType(proven)) > 0 &&
		!lowering.isArrayType(proven) &&
		!checker.IsTupleType(proven) &&
		!lowering.isLibraryType(proven, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray") {
		if lowering.recordElement(proven) != nil {
			return ir.Record, true
		}
		regex := false
		for _, info := range lowering.checker.GetIndexInfosOfType(proven) {
			if declaration := info.Declaration(); declaration != nil {
				file := ast.GetSourceFileOfNode(declaration)
				if load.IsLibrary(file) && strings.Contains(string(file.AsSourceFile().FileName()), ".regexp.") {
					regex = true
				}
			}
		}
		if !regex {
			return 0, false
		}
	}
	if flags&checker.TypeFlagsTypeParameter != 0 {
		// Inside a generic class, a type parameter is what this instantiation made it.
		substituted, isKnown := lowering.substitution[proven]
		return substituted, isKnown
	}
	if flags&checker.TypeFlagsIntersection != 0 {
		// Target & WeakBrand is what a Weak<Target> reads as where it's present: the target.
		if target := lowering.weakTarget(proven); target != nil {
			return lowering.representation(target)
		}
		if primitive := lowering.phantomBase(proven); primitive != nil {
			if lowering.phantomUndefined(proven) {
				return ir.Object, true
			}
			return lowering.representation(primitive)
		}
		return lowering.objectIntersection(proven)
	}
	switch {
	case flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0:
		return ir.Object, true
	case flags&checker.TypeFlagsNumberLike != 0:
		return ir.Number, true
	case flags&checker.TypeFlagsStringLike != 0:
		return ir.String, true
	case flags&checker.TypeFlagsBooleanLike != 0:
		return ir.Boolean, true
	case flags&checker.TypeFlagsObject != 0 &&
		(lowering.isArrayType(proven) ||
			lowering.isLibraryType(proven, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray")):
		return ir.Array, true
	case flags&checker.TypeFlagsObject != 0 &&
		lowering.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
		// A Set is held as a Map whose values aren't used (set.go).
		return ir.Map, true
	case flags&checker.TypeFlagsObject != 0 && lowering.isLibraryType(proven, "Promise"):
		return ir.Promise, true
	case flags&checker.TypeFlagsObject != 0 &&
		!isClassInstance(proven) &&
		(lowering.checker.IsTypeAssignableTo(lowering.checker.GetNumberType(), proven) ||
			lowering.checker.IsTypeAssignableTo(lowering.checker.GetStringType(), proven) ||
			lowering.checker.IsTypeAssignableTo(lowering.checker.GetBooleanType(), proven)):
		// Structural types such as {} admit primitives. Their slots must preserve
		// the runtime brand with the same boxes used for scalar/reference unions.
		return ir.Union, true
	case flags&checker.TypeFlagsObject != 0 &&
		len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) == 0:
		return ir.Object, true
	case flags&checker.TypeFlagsObject != 0:
		// An object with call signatures is a function, held as a closure.
		return ir.Closure, true
	case flags&checker.TypeFlagsUnion != 0:
		if lowering.includesNull(proven) && !dynamicObjectType(proven) {
			// Preserve the area's nullable reference and RegExp protocols. Only
			// differently stored members need the boxed union representation.
			boxed := lowering.includesUndefined(proven)
			for _, member := range proven.Types() {
				if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
					continue
				}
				of, known := lowering.representation(member)
				if !known {
					return 0, false
				}
				// A dictionary is boxed too: a bare record pointer has no null of its own.
				boxed = boxed || !of.IsReference() || of == ir.String || of == ir.Union || of == ir.Record
			}
			if boxed {
				return ir.Union, true
			}
		}
		var shared ir.Type
		mixed, weak := false, false
		for _, member := range proven.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 ||
				lowering.phantomUndefined(member) {
				// undefined joins a union of references as a null pointer; it's checked below that
				// the rest are references.
				continue
			}
			if lowering.weakTarget(member) != nil {
				// Weak<Target> is (Target & WeakBrand) | undefined: kept weakly.
				weak = true
				continue
			}
			memberType, isKnown := lowering.representation(member)
			if !isKnown {
				return 0, false
			}
			if shared != 0 && memberType != shared {
				mixed = true
			}
			shared = memberType
		}
		if weak {
			return ir.Weak, shared == 0
		}
		if mixed {
			// Members held differently (string | number) are one Union, which holds undefined too.
			return ir.Union, true
		}
		if shared != 0 && !shared.IsReference() && lowering.includesUndefined(proven) {
			// number | undefined and boolean | undefined are each a present-and-value pair.
			if maybe := ir.Maybe(shared); maybe.IsMaybe() {
				return maybe, true
			}
			return 0, false
		}
		return shared, shared != 0
	}
	return 0, false
}

// isLibraryType reports whether a type is one of the library's, by name: Map, not a program's own
// interface that happens to be called Map. A phantom extension is judged as its base (unbranded.go), so
// tsc's SortedReadonlyArray<T> is a ReadonlyArray wherever lowering asks.
func (lowering *lowering) isLibraryType(proven *checker.Type, names ...string) bool {
	symbol := lowering.unbranded(proven).Symbol()
	if symbol == nil ||
		len(symbol.Declarations) == 0 ||
		!load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return false
	}
	for _, name := range names {
		if symbol.Name == name {
			return true
		}
	}
	return false
}

func (lowering *lowering) includesUndefined(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return proven.Flags()&checker.TypeFlagsUndefined != 0 || lowering.phantomUndefined(proven)
	}
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 || lowering.phantomUndefined(member) {
			return true
		}
	}
	return false
}

func (lowering *lowering) includesNull(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return proven.Flags()&checker.TypeFlagsNull != 0
	}
	for _, member := range proven.Types() {
		if lowering.includesNull(member) {
			return true
		}
	}
	return false
}

// expression lowers a value. What's kept weakly (a Weak<Target> variable, field, element or map value)
// is read here as its target, so no value of a Weak type goes further; keeping one is fit's WeakOf.
func (lowering *lowering) expression(node *ast.Node) (ir.Expression, error) {
	if nodeHostCallbackArgumentError := lowering.nodeHostCallbackArgument(node); nodeHostCallbackArgumentError != nil {
		return nil, nodeHostCallbackArgumentError
	}
	if weakArrayCallbackError := lowering.weakArrayCallback(node); weakArrayCallbackError != nil {
		return nil, weakArrayCallbackError
	}
	if libraryIteratorUnsupportedUseError := lowering.libraryIteratorUnsupportedUse(
		node,
	); libraryIteratorUnsupportedUseError != nil {
		return nil, libraryIteratorUnsupportedUseError
	}
	if nodeBufferUnsupportedUseError := lowering.nodeBufferUnsupportedUse(node); nodeBufferUnsupportedUseError != nil {
		return nil, nodeBufferUnsupportedUseError
	}
	if regexUnsupportedUseError := lowering.regexUnsupportedUse(node); regexUnsupportedUseError != nil {
		return nil, regexUnsupportedUseError
	}
	value, valueError := lowering.value(node)
	if valueError == nil {
		value = lowering.graphAllocation(value, node)
	}
	if valueError == nil {
		if value.Type() == ir.Object &&
			lowering.nullableObjectUnion(lowering.checker.GetTypeAtLocation(node)) &&
			!nullableObjectLookupTypeOf(node, value) {
			return nil, lowering.notYet(node, "a nullable object lookup without separate null/undefined tags")
		}
		if contextual := lowering.slotType(node); contextual != nil {
			if valueSlotTagsError := lowering.valueSlotTags(node, contextual); valueSlotTagsError != nil {
				return nil, valueSlotTagsError
			}
		}
	}
	if valueError == nil &&
		value.Type() != ir.Weak &&
		lowering.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
		if _, constant := value.(ir.Undefined); !constant {
			if contextual := lowering.slotType(node); contextual != nil {
				if to, known := lowering.representation(
					contextual,
				); known && to != value.Type() && to != ir.Union && (to.IsMaybe() || to.IsReference()) {
					// Keep the read or call's effects when converting its only possible value.
					value = ir.Narrow{Value: ir.Box{Value: value}, To: to}
				}
			}
		}
	}
	if literal := ast.SkipParentheses(node).Kind; valueError == nil &&
		value.Type().IsReference() &&
		literal != ast.KindArrayLiteralExpression &&
		literal != ast.KindObjectLiteralExpression {
		if contextual := lowering.slotType(node); contextual != nil {
			if valueSlotKeepsError := lowering.valueSlotKeeps(node, contextual); valueSlotKeepsError != nil {
				return nil, valueSlotKeepsError
			}
		}
	}
	// A generic body names its parameter's constraint, not what the call infers, so an argument is
	// judged against the parameter as declared too, a literal made right there included (#8k12t4s).
	if valueError == nil && value.Type().IsReference() {
		if declared := lowering.nodeHostDeclaredArgument(node); declared != nil {
			if own := lowering.checker.GetTypeAtLocation(
				node,
			); !lowering.nodeBufferType(lowering.present(own), "Buffer") &&
				!lowering.nodeBufferType(lowering.present(own), "Hash") {
				if reached := lowering.nodeHostViewAt(own, declared); reached != "" {
					return nil, lowering.nodeHostViewRefusal(node, own, declared, reached)
				}
			}
		}
	}
	// Anything the check above doesn't see (a tuple inside a literal, or a union around it) is still
	// not yet: wherever a tuple flows into an array slot, at any depth.
	if skipped := ast.SkipParentheses(node); valueError == nil && skipped.Kind != ast.KindSpreadElement {
		if contextual := lowering.slotType(
			skipped,
		); contextual != nil &&
			lowering.tupleWhereArrayGoes(lowering.checker.GetTypeAtLocation(skipped), contextual, 0) {
			return nil, lowering.notYet(
				skipped,
				"a tuple where an array goes (as "+lowering.checker.TypeToString(contextual)+")",
			)
		}
		// So is a value held other than as a map where a map is held, literals included: a literal is
		// made as an object, never as the slot's map (container_fit.go). And so is a real Map, Set or
		// array where a structural object type is held, unless the slot reads it only by members it's
		// served by (shape_fit.go).
		if contextual := lowering.slotType(skipped); contextual != nil {
			if stop := lowering.containerWhereOtherGoes(
				lowering.checker.GetTypeAtLocation(skipped),
				contextual,
				lowering.libraryCallArgument(skipped),
			); stop != "" {
				return nil, lowering.notYet(skipped, stop)
			}
			// And so is a container whose elements are stored otherwise than the slot reads them
			// (element_fit.go), literals included, since a union slot doesn't make one as its own.
			if stop := lowering.elementsHeldOtherwise(skipped, contextual); stop != "" {
				return nil, lowering.notYet(skipped, stop)
			}
		}
	}
	if valueError != nil || value.Type() != ir.Weak {
		if valueError == nil {
			value = lowering.nullableObjectBoundary(node, value)
		}
		return value, valueError
	}
	read := lowering.checker.GetTypeAtLocation(node)
	to, present := ir.Object, false
	if target, isKnown := lowering.representation(
		read,
	); isKnown && target != ir.Weak && read.Flags()&checker.TypeFlagsUndefined == 0 {
		// The checker narrowed it to present (Target & WeakBrand).
		to, present = target, true
	} else if read.Flags()&checker.TypeFlagsUnion != 0 {
		// A union of branded targets stays represented as Weak after narrowing.
		// Absence, not that representation, decides whether its read needs liveness.
		present = !lowering.includesUndefined(read)
		for _, member := range read.Types() {
			if target := lowering.weakTarget(member); target != nil {
				if to, isKnown = lowering.representation(target); !isKnown {
					return nil, lowering.notYet(node, "a Weak of "+lowering.checker.TypeToString(target))
				}
			}
		}
	}
	if !to.IsReference() {
		return nil, lowering.notYet(node, "a Weak of "+lowering.checker.TypeToString(read))
	}
	return ir.WeakTarget{Value: value, To: to, Present: present}, nil
}

// sameKeeping reports whether a value of type from, seen as type to, keeps everything inside it the
// same way through both: every field to has, every element, map key and value, and every parameter
// and result of a function, Weak in both or in neither, all the way down.
func (lowering *lowering) sameKeeping(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	from, to = lowering.present(from), lowering.present(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	visited[[2]*checker.Type{from, to}] = true
	if fromKind, toKind := lowering.typedArrayKind(from), lowering.typedArrayKind(to); fromKind != 0 || toKind != 0 {
		return fromKind != 0 && fromKind == toKind
	}
	if !lowering.nodeBufferView(from, to) {
		return false
	}
	if a, b := lowering.recordElement(from), lowering.recordElement(to); a != nil && b != nil {
		return lowering.sameKeeping(a, b, visited)
	}
	same := func(inside, viewed *checker.Type) bool {
		fromKept, _ := lowering.kept(inside)
		toKept, _ := lowering.kept(viewed)
		return (fromKept == ir.Weak) == (toKept == ir.Weak) && lowering.sameKeeping(inside, viewed, visited)
	}
	fromSignatures := lowering.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := lowering.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	switch {
	case len(fromSignatures) > 0 && len(toSignatures) > 0:
		// A closure's slots have no dynamic type descriptor. A view changing a boxed union
		// to an unboxed scalar (or back) needs an adapter. References share the union pointer ABI.
		sameCallable := func(inside, viewed *checker.Type) bool {
			fromKept, _ := lowering.kept(inside)
			toKept, _ := lowering.kept(viewed)
			return !(fromKept == ir.Union && !toKept.IsReference() || toKept == ir.Union && !fromKept.IsReference()) &&
				same(inside, viewed)
		}
		// Intrinsic callback adapters call the source signature, not the library's wider view.
		if declaration := toSignatures[0].Declaration(); declaration != nil &&
			load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			sameCallable = same
		}
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		if lowering.censusNeverRestSignature(toSignatures[0]) {
			toParameters = nil
		}
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			if !sameCallable(
				lowering.checker.GetTypeOfSymbol(fromParameters[index]),
				lowering.checker.GetTypeOfSymbol(toParameters[index]),
			) {
				return false
			}
		}
		return sameCallable(
			lowering.checker.GetReturnTypeOfSignature(fromSignatures[0]),
			lowering.checker.GetReturnTypeOfSignature(toSignatures[0]),
		)
	case from.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		to.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		(lowering.isArrayType(from) ||
			checker.IsTupleType(from) ||
			lowering.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet")):
		fromArguments, toArguments := lowering.typeArguments(from), lowering.typeArguments(to)
		for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
			if !same(fromArguments[index], toArguments[index]) {
				return false
			}
		}
	default:
		for _, viewed := range lowering.checker.GetPropertiesOfType(to) {
			if viewed.Flags&ast.SymbolFlagsMethod != 0 {
				if inside := lowering.checker.GetPropertyOfType(from, viewed.Name); inside != nil &&
					!lowering.sameWeakMethodSlots(
						lowering.checker.GetTypeOfSymbol(inside),
						lowering.checker.GetTypeOfSymbol(viewed),
					) {
					return false
				}
				continue
			}
			if inside := lowering.checker.GetPropertyOfType(
				from,
				viewed.Name,
			); inside != nil &&
				!same(lowering.checker.GetTypeOfSymbol(inside), lowering.checker.GetTypeOfSymbol(viewed)) {
				return false
			}
		}
	}
	return true
}

// tupleSeenAsArray finds, in a value of type from seen as type to, the first tuple seen as an array:
// at the top, or as an element, field, map key or value, parameter or result, all the way down as
// sameKeeping walks. It returns the tuple's type and the array's, or nils.
func (lowering *lowering) tupleSeenAsArray(
	from *checker.Type,
	to *checker.Type,
	visited map[[2]*checker.Type]bool,
) (*checker.Type, *checker.Type) {
	from, to = lowering.present(from), lowering.present(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return nil, nil
	}
	visited[[2]*checker.Type{from, to}] = true
	if checker.IsTupleType(from) && !checker.IsTupleType(to) {
		return from, to
	}
	fromSignatures := lowering.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := lowering.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	switch {
	case len(fromSignatures) > 0 && len(toSignatures) > 0:
		// A function is handed the other's arguments, and its results are seen as the other's.
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		if lowering.censusNeverRestSignature(toSignatures[0]) {
			toParameters = nil
		}
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			if tuple, array := lowering.tupleSeenAsArray(
				lowering.checker.GetTypeOfSymbol(toParameters[index]),
				lowering.checker.GetTypeOfSymbol(fromParameters[index]),
				visited,
			); tuple != nil {
				return tuple, array
			}
		}
		return lowering.tupleSeenAsArray(
			lowering.checker.GetReturnTypeOfSignature(fromSignatures[0]),
			lowering.checker.GetReturnTypeOfSignature(toSignatures[0]),
			visited,
		)
	case from.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		to.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		(lowering.isArrayType(from) ||
			checker.IsTupleType(from) ||
			lowering.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet")):
		fromArguments, toArguments := lowering.typeArguments(from), lowering.typeArguments(to)
		if checker.IsTupleType(from) && lowering.isArrayType(to) {
			return from, to
		}
		for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
			if tuple, array := lowering.tupleSeenAsArray(
				fromArguments[index],
				toArguments[index],
				visited,
			); tuple != nil {
				return tuple, array
			}
		}
	default:
		for _, viewed := range lowering.checker.GetPropertiesOfType(to) {
			if viewed.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}
			if inside := lowering.checker.GetPropertyOfType(from, viewed.Name); inside != nil {
				if tuple, array := lowering.tupleSeenAsArray(
					lowering.checker.GetTypeOfSymbol(inside),
					lowering.checker.GetTypeOfSymbol(viewed),
					visited,
				); tuple != nil {
					return tuple, array
				}
			}
		}
	}
	return nil, nil
}

// present is a type without undefined, and a Weak's narrowing without its brand: the one object,
// array, map or function type it is, or nil when it's none or several.
func (lowering *lowering) present(proven *checker.Type) *checker.Type {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		var only *checker.Type
		for _, member := range proven.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				continue
			}
			if only != nil {
				return nil
			}
			only = member
		}
		if only == nil {
			return nil
		}
		proven = only
	}
	if target := lowering.weakTarget(proven); target != nil {
		proven = target
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	return proven
}

// kept is how an element or a map's value of a type is kept: as its representation, except that one
// narrowed from a Weak (Node & WeakBrand, as filter(x => x !== undefined) or every leaves an array of
// Weak<Node>) is still a Weak, a handle, since that's what the array holds.
func (lowering *lowering) kept(proven *checker.Type) (ir.Type, bool) {
	if lowering.weakTarget(proven) != nil {
		return ir.Weak, true
	}
	return lowering.representation(proven)
}

// typeArguments is a generic type's arguments, Map<K, V>'s K and V, through a Weak's narrowing
// (Map<K, V> & WeakBrand), or nil for a type that has none.
func (lowering *lowering) typeArguments(proven *checker.Type) []*checker.Type {
	if target := lowering.weakTarget(proven); target != nil {
		proven = target
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return nil
	}
	return lowering.checker.GetTypeArguments(proven)
}

// weakTarget is the Target of Target & WeakBrand, the present half of a Weak<Target>, or nil for
// any other type.
func (lowering *lowering) weakTarget(proven *checker.Type) *checker.Type {
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return nil
	}
	var target *checker.Type
	branded := false
	for _, member := range proven.Types() {
		if symbol := member.Symbol(); symbol != nil &&
			symbol.Name == "WeakBrand" &&
			len(symbol.Declarations) > 0 &&
			load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
			branded = true
		} else if target == nil {
			target = member
		} else {
			return nil
		}
	}
	if !branded {
		return nil
	}
	return target
}

// value lowers a value, as expression does, but leaves a Weak as it's kept.
func (lowering *lowering) value(node *ast.Node) (ir.Expression, error) {
	if identity := lowering.enumNeverIdentity(node, map[*ast.Node]bool{}); identity != nil {
		value, enumNeverValueError := lowering.enumNeverValue(node)
		if enumNeverValueError != nil {
			return nil, enumNeverValueError
		}
		return lowering.enumNeverCheck(node, value, identity), nil
	}
	value, enumNeverValueError := lowering.enumNeverValue(node)
	if enumNeverValueError != nil {
		return nil, enumNeverValueError
	}
	return lowering.programAllocation(node, value)
}

func (lowering *lowering) enumNeverValue(node *ast.Node) (ir.Expression, error) {
	if value, known, processValueError := lowering.processValue(node); known {
		return value, processValueError
	}
	if value, known, nodeFsDirectoryValueError := lowering.nodeFSDirectoryValue(node); known {
		return value, nodeFsDirectoryValueError
	}
	node = ast.SkipParentheses(node)
	if value, handled, functionObservationError := lowering.functionObservation(node); handled {
		return value, functionObservationError
	}
	if value, handled, mergedComputedReadError := lowering.mergedComputedRead(node); handled {
		return value, mergedComputedReadError
	}
	if member, handled, phantomMemberError := lowering.phantomMember(node); handled {
		return member, phantomMemberError
	}
	if value, handled, typedArrayExpressionError := lowering.typedArrayExpression(node); handled {
		return value, typedArrayExpressionError
	}
	if value, known, nodeRealpathValueError := lowering.nodeRealpathValue(node); known {
		return value, nodeRealpathValueError
	}
	if value, known, namespaceExpressionError := lowering.namespaceExpression(node); known {
		return value, namespaceExpressionError
	}
	if value, known, libraryMethodValueError := lowering.libraryMethodValue(node); known {
		return value, libraryMethodValueError
	}
	if lowering.isArgumentsLength(node) {
		return lowering.readArgumentsCount(), nil
	}
	if value, known, enumExpressionError := lowering.enumExpression(node); known {
		return value, enumExpressionError
	}
	if value, handled, recordExpressionError := lowering.recordExpression(node); handled {
		return value, recordExpressionError
	}
	if observed, known := lowering.libraryArrayObservation(node); known {
		return observed, nil
	}
	switch node.Kind {
	case ast.KindAwaitExpression:
		return lowering.awaitExpression(node)
	case ast.KindNullKeyword:
		return ir.Null{}, nil
	case ast.KindRegularExpressionLiteral:
		return lowering.regexConstant(node)
	case ast.KindNumericLiteral:
		return lowering.numericLiteral(node)
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return ir.StringConstant{Index: lowering.constant(node.Text())}, nil
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return ir.BooleanConstant{Value: node.Kind == ast.KindTrueKeyword}, nil
	case ast.KindIdentifier:
		if row, _ := lowering.hostBindingOf(node); row != nil {
			// hostBindingRefusal admits only a typeof or a read its guard proves dead, and neither reaches here.
			return nil, errors.New("lower: compiler bug: " +
				lowering.program.Where(node) +
				": a read of the host binding " +
				row.Name +
				" reached lowering")
		}
		if value, known, libraryGlobalValueError := lowering.libraryGlobalValue(node); known {
			return value, libraryGlobalValueError
		}
		if value, handled := lowering.staticClassRead(node); handled {
			return value, nil
		}
		if value, handled, nestedReferenceError := lowering.nestedReference(node); handled {
			return value, nestedReferenceError
		}
		local, isLocal := lowering.local(node)
		if !isLocal && node.Text() == "undefined" {
			return ir.Undefined{}, nil
		}
		if !isLocal && (lowering.isLibraryGlobal(node, "NaN") || lowering.isLibraryGlobal(node, "Infinity")) {
			return ir.NumberConstant{Value: numberConstants[node.Text()]}, nil
		}
		if function, isFunction := lowering.functions[lowering.symbol(node)]; !isLocal && isFunction {
			return lowering.functionValue(node, function)
		}
		if declaration, isGeneric := lowering.generics[lowering.symbol(node)]; !isLocal && isGeneric {
			if generatorDeclaration(declaration) {
				// As functionValue stops a non-generic generator, before any instance is asked for.
				return nil, lowering.generatorNotYet(node, "used as a value")
			}
			return lowering.genericFunctionValue(node, declaration)
		}
		if !isLocal && lowering.isLibraryGlobal(node, "String") {
			return nil, lowering.notYet(
				node,
				"reading String as a first-class constructor (its any-typed call signature, construction and static members need an intrinsic value representation)",
			)
		}
		if !isLocal {
			return nil, lowering.notYet(node, "reading "+node.Text())
		}
		return lowering.localRead(node, local)
	case ast.KindVoidExpression:
		operand, discardedValueError := lowering.discardedValue(node.AsVoidExpression().Expression)
		if discardedValueError != nil {
			return nil, discardedValueError
		}
		if operand.Type() == ir.Promise {
			return nil, &Refused{
				Where: lowering.program.Where(node),
				What:  "an unawaited async task through the void operator",
				Fix:   "await the call or keep its Promise",
			}
		}
		return ir.Void{Value: operand}, nil
	case ast.KindPrefixUnaryExpression:
		return lowering.prefix(node)
	case ast.KindTypeOfExpression:
		if value, handled, hostBindingTypeOfError := lowering.hostBindingTypeOf(node); handled {
			return value, hostBindingTypeOfError
		}
		if lowering.isLibraryGlobal(node.AsTypeOfExpression().Expression, "Number") {
			return ir.StringConstant{Index: lowering.constant("function")}, nil
		}
		if value, intrinsic := lowering.stringTypeOf(node); intrinsic {
			return value, nil
		}
		if value, handled, methodPresenceTypeOfError := lowering.methodPresenceTypeOf(node); handled {
			return value, methodPresenceTypeOfError
		}
		operand, expressionError := lowering.expression(node.AsTypeOfExpression().Expression)
		if expressionError != nil {
			return nil, expressionError
		}
		written := node.AsTypeOfExpression().Expression
		null := lowering.typeOfNull(written)
		if operand.Type() == ir.Union {
			null = false
		}
		if null && lowering.includesUndefined(lowering.concrete(lowering.checker.GetTypeAtLocation(written))) {
			switch operand.(type) {
			case ir.ArrayIndex, ir.MapGet, ir.ArrayPop:
				// The lookup still has a presence slot, so typeof can distinguish null from undefined.
			default:
				return nil, lowering.notYet(
					node,
					"typeof a value holding both null and undefined without a presence slot",
				)
			}
		}
		return ir.TypeOf{Value: operand, Null: null}, nil
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if _, compound := compoundAssignments[binary.OperatorToken.Kind]; compound ||
			binary.OperatorToken.Kind == ast.KindEqualsToken {
			return lowering.assignmentValue(node)
		}
		if binary.OperatorToken.Kind == ast.KindCommaToken {
			return lowering.comma(node)
		}
		if logicalAssignment(binary.OperatorToken.Kind) {
			return lowering.logicalAssignment(node)
		}
		if binary.OperatorToken.Kind == ast.KindInKeyword {
			return lowering.inProperty(node)
		}
		if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken {
			return lowering.coalesce(node)
		}
		// All arrays are truthy, including []; only undefined takes this fallback.
		if binary.OperatorToken.Kind == ast.KindBarBarToken &&
			lowering.arrayOrUndefined(lowering.checker.GetTypeAtLocation(binary.Left)) {
			return lowering.coalesce(node)
		}
		if binary.OperatorToken.Kind == ast.KindInstanceOfKeyword {
			if lowered, isCaught := lowering.caughtInstanceOfError(node); isCaught {
				return lowered, nil
			}
			return lowering.classInstanceOf(node)
		}
		if value, handled, hostBindingAndError := lowering.hostBindingAnd(node); handled {
			return value, hostBindingAndError
		}
		if value, handled, methodPresenceAndError := lowering.methodPresenceAnd(node); handled {
			return value, methodPresenceAndError
		}
		// A flow-narrowed never expression still reads its declared boolean slots.
		// Only a live result needs a placeholder for a non-returning arm.
		if (binary.OperatorToken.Kind == ast.KindBarBarToken ||
			binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken) &&
			lowering.isNever(binary.Right) &&
			!lowering.isNever(node) {
			return lowering.logicalNever(node)
		}
		left, expressionError := lowering.expression(binary.Left)
		if expressionError != nil {
			return nil, expressionError
		}
		right, expressionError := lowering.expression(binary.Right)
		if expressionError != nil {
			return nil, expressionError
		}
		if binary.OperatorToken.Kind == ast.KindPlusToken {
			left, right = lowering.spelled(binary.Left, left), lowering.spelled(binary.Right, right)
		}
		return lowering.combine(node, binary.OperatorToken.Kind, left, right)
	case ast.KindTaggedTemplateExpression:
		return lowering.stringRawTemplate(node)
	case ast.KindTemplateExpression:
		return lowering.template(node)
	case ast.KindConditionalExpression:
		return lowering.conditional(node)
	case ast.KindObjectLiteralExpression:
		return lowering.objectLiteral(node)
	case ast.KindArrayLiteralExpression:
		return lowering.arrayLiteral(node)
	case ast.KindPropertyAccessExpression:
		return lowering.property(node)
	case ast.KindElementAccessExpression:
		return lowering.elementAccess(node)
	case ast.KindNewExpression:
		if lowering.isLibraryGlobal(node.AsNewExpression().Expression, "Promise") {
			return nil, lowering.notYet(node, "Promise executors and pending-I/O cancellation")
		}
		if genericFunctionIdentityNewError := lowering.genericFunctionIdentityNew(
			node,
		); genericFunctionIdentityNewError != nil {
			return nil, genericFunctionIdentityNewError
		}
		return lowering.newExpression(node)
	case ast.KindThisKeyword:
		if lowering.this < 0 {
			return nil, lowering.notYet(node, "this outside a method")
		}
		if useOfThisError := lowering.useOfThis(node); useOfThisError != nil {
			return nil, useOfThisError
		}
		lowering.touch(lowering.this)
		return ir.Read{Local: lowering.this, Of: ir.Object}, nil
	case ast.KindArrowFunction:
		return lowering.closure(node)
	case ast.KindNonNullExpression:
		return lowering.nonNull(node)
	case ast.KindAsExpression:
		return lowering.cast(node)
	case ast.KindSatisfiesExpression:
		satisfies := node.AsSatisfiesExpression()
		if provenRelationError := lowering.provenRelation(
			node,
			satisfies.Expression,
			lowering.checker.GetTypeAtLocation(satisfies.Type),
		); provenRelationError != nil {
			return nil, provenRelationError
		}
		return lowering.expression(satisfies.Expression)
	case ast.KindFunctionExpression:
		return lowering.functionExpression(node)
	case ast.KindYieldExpression:
		// A yield's result is what next(x) sends in, which isn't built (#qk8rztp S5).
		return nil, lowering.generatorNotYet(node, "using a yield's result as a value")
	case ast.KindCallExpression:
		if value, handled, promiseValueError := lowering.promiseValue(node); handled {
			return value, promiseValueError
		}
		if genericFunctionIdentityCallError := lowering.genericFunctionIdentityCall(
			node,
		); genericFunctionIdentityCallError != nil {
			return nil, genericFunctionIdentityCallError
		}
		if value, known, optionalIntrinsicError := lowering.optionalIntrinsic(node); known {
			return value, optionalIntrinsicError
		}
		if optionalCallError := lowering.optionalCall(node); optionalCallError != nil {
			return nil, optionalCallError
		}
		if lowered, isBuiltin, builtinError := lowering.builtin(node); isBuiltin {
			if builtinError == nil && lowered.Type() == 0 {
				return ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: lowered}}, Result: ir.Undefined{}}, nil
			}
			return lowered, builtinError
		}
		call, callOrMethodError := lowering.callOrMethod(node)
		if callOrMethodError != nil {
			return nil, callOrMethodError
		}
		if call.Type() == 0 {
			return ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: call}}, Result: ir.Undefined{}}, nil
		}
		return call, nil
	}
	return nil, lowering.notYet(node, describe(node))
}

// fit makes a value fit where a value of type to goes: a number or a boolean, or undefined, where
// number | undefined or boolean | undefined goes, since that is two words and they are one. It also
// unwraps a maybe value the checker narrowed to its present type. Anything else is left as it is.
func fit(value ir.Expression, to ir.Type) ir.Expression {
	if discarded, ok := value.(ir.Void); ok && (to.IsMaybe() || to.IsReference()) {
		discarded.Of = to
		return discarded
	}
	if to == ir.Weak && value != nil && value.Type() != ir.Weak {
		return ir.WeakOf{Value: value}
	}
	if to == ir.Union && value != nil && value.Type() != ir.Union {
		return ir.Box{Value: value}
	}
	if _, isNull := value.(ir.Null); isNull && to.IsReference() && to != ir.Union {
		return ir.Null{Of: to}
	}
	if _, isUndefined := value.(ir.Undefined); isUndefined && to.IsReference() && to != ir.Union {
		// undefined going where a string, an array or a function may be missing is that reference,
		// missing: typed as it, so the C holding it is.
		return ir.Undefined{Of: to}
	}
	if !to.IsMaybe() || value == nil {
		if value != nil && value.Type().IsMaybe() && value.Type().Present() == to {
			return ir.Unwrap{Value: value}
		}
		return value
	}
	if value.Type() == to.Present() {
		return ir.MaybeOf{Value: value, Of: to}
	}
	if _, isUndefined := value.(ir.Undefined); isUndefined {
		return ir.MaybeOf{Of: to}
	}
	return value
}

// numericLiteral is the value the checker read from the literal, so 0x1F, 1_000 and 1e3 all mean
// what JavaScript says they mean without a second parser here.
func (lowering *lowering) numericLiteral(node *ast.Node) (ir.Expression, error) {
	literal := lowering.checker.GetTypeAtLocation(node)
	if literal.Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return nil, errors.New("lower: " +
			lowering.program.Where(node) +
			": the checker gave a numeric literal a type that isn't a number literal")
	}
	value := reflect.ValueOf(literal.AsLiteralType().Value())
	if value.Kind() != reflect.Float64 {
		return nil, errors.New("lower: " + lowering.program.Where(node) + ": a numeric literal's value isn't a float64")
	}
	return ir.NumberConstant{Value: value.Float()}, nil
}

func (lowering *lowering) prefix(node *ast.Node) (ir.Expression, error) {
	prefix := node.AsPrefixUnaryExpression()
	if prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
		operand := ast.SkipParentheses(prefix.Operand)
		if !ast.IsIdentifier(operand) {
			return nil, lowering.notYet(node, "a prefix update value whose target is not a plain name")
		}
		local, known := lowering.local(operand)
		if !known || lowering.result.Locals[local].Type != ir.Number {
			return nil, lowering.notYet(node, "a prefix update value whose name is not stored as a number")
		}
		// An immediately called helper makes the write and resulting read one expression.
		// Capturing a local carries its existing cell, so the caller observes the write.
		index := len(lowering.result.Functions)
		lowering.result.Functions = append(
			lowering.result.Functions,
			ir.Function{Name: "prefix_update", Closure: true, Returns: ir.Number},
		)
		outerIndex := lowering.functionIndex
		lowering.functionIndex = index
		lowering.closures = append(lowering.closures, index)
		body, err := lowering.increment(node)
		var value ir.Expression
		if err == nil {
			value, err = lowering.expression(operand)
		}
		lowering.closures = lowering.closures[:len(lowering.closures)-1]
		lowering.functionIndex = outerIndex
		if err != nil {
			return nil, err
		}
		if value.Type() != ir.Number {
			return nil, lowering.notYet(node, "a prefix update of a non-number")
		}
		lowering.result.Functions[index].Body = append(body, ir.Return{Value: value})
		return ir.CallClosure{Closure: ir.MakeClosure{Function: index}, Returns: ir.Number}, nil
	}
	if prefix.Operator == ast.KindPlusToken ||
		prefix.Operator == ast.KindMinusToken ||
		prefix.Operator == ast.KindTildeToken {
		operand, libraryNumberError := lowering.libraryNumber(prefix.Operand)
		if libraryNumberError != nil {
			return nil, libraryNumberError
		}
		operator := ir.Plus
		if prefix.Operator == ast.KindMinusToken {
			operator = ir.Negate
		}
		if prefix.Operator == ast.KindTildeToken {
			operator = ir.BitNot
		}
		return ir.Unary{Operator: operator, Operand: operand}, nil
	}
	if prefix.Operator == ast.KindExclamationToken {
		if negated, handled, methodPresenceNotError := lowering.methodPresenceNot(prefix.Operand); handled {
			return negated, methodPresenceNotError
		}
	}
	operand, expressionError := lowering.expression(prefix.Operand)
	if expressionError != nil {
		return nil, expressionError
	}
	if prefix.Operator == ast.KindExclamationToken {
		return ir.Unary{Operator: ir.Not, Operand: censusCondition(operand)}, nil
	}
	return nil, lowering.notYet(node, describe(node)+" on a "+typeName(operand.Type()))
}

var arithmetic = map[ast.Kind]ir.Operator{
	ast.KindPlusToken:             ir.Add,
	ast.KindMinusToken:            ir.Subtract,
	ast.KindAsteriskToken:         ir.Multiply,
	ast.KindSlashToken:            ir.Divide,
	ast.KindPercentToken:          ir.Remainder,
	ast.KindAsteriskAsteriskToken: ir.Power,
}

// bitwise are the binary bitwise operators, on two numbers.
var bitwise = map[ast.Kind]ir.Operator{
	ast.KindAmpersandToken:                         ir.BitAnd,
	ast.KindBarToken:                               ir.BitOr,
	ast.KindCaretToken:                             ir.BitXor,
	ast.KindLessThanLessThanToken:                  ir.ShiftLeft,
	ast.KindGreaterThanGreaterThanToken:            ir.ShiftRight,
	ast.KindGreaterThanGreaterThanGreaterThanToken: ir.ShiftRightUnsigned,
}

var comparisons = map[ast.Kind]ir.Operator{
	ast.KindLessThanToken:          ir.Less,
	ast.KindLessThanEqualsToken:    ir.LessOrEqual,
	ast.KindGreaterThanToken:       ir.Greater,
	ast.KindGreaterThanEqualsToken: ir.GreaterOrEqual,
}

// combine lowers a binary operator on two lowered operands.
func (lowering *lowering) combine(
	node *ast.Node,
	operator ast.Kind,
	left ir.Expression,
	right ir.Expression,
) (ir.Expression, error) {
	if (operator == ast.KindEqualsEqualsToken || operator == ast.KindExclamationEqualsToken) &&
		lowering.placeholderNullishTest(node) {
		value := left
		if _, isNull := left.(ir.Null); isNull {
			value = right
		}
		test := ir.Expression(
			ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: value}, Right: ir.IsUndefined{Value: value}},
		)
		if _, simple := value.(ir.Read); !simple {
			// Hold effectful receivers once before testing both nullish tags.
			builder := lowering.libraryArrayBuilder([]ir.Expression{value})
			held := builder.read(builder.parameters[0])
			test = ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: held}, Right: ir.IsUndefined{Value: held}}
			if operator == ast.KindExclamationEqualsToken {
				test = ir.Unary{Operator: ir.Not, Operand: test}
			}
			return builder.finish("placeholder_nullish_test", test), nil
		}
		if operator == ast.KindExclamationEqualsToken {
			test = ir.Unary{Operator: ir.Not, Operand: test}
		}
		return test, nil
	}
	both := func(want ir.Type) bool { return left.Type() == want && right.Type() == want }
	if operator == ast.KindPlusToken && (left.Type() == ir.String || right.Type() == ir.String) {
		left, right = lowering.concatenated(left), lowering.concatenated(right)
	}
	if operator == ast.KindPlusToken && both(ir.String) {
		return ir.Concat{Parts: []ir.Expression{left, right}}, nil
	}
	if lowered, isArithmetic := arithmetic[operator]; isArithmetic && both(ir.Number) {
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if lowered, isBitwise := bitwise[operator]; isBitwise && both(ir.Number) {
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if lowered, isComparison := comparisons[operator]; isComparison && (both(ir.Number) || both(ir.String)) {
		// Strings compare in UTF-16 code unit order, as JavaScript's do.
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
		_, leftNull := left.(ir.Null)
		_, rightNull := right.(ir.Null)
		if leftNull != rightNull {
			value := left
			if leftNull {
				value = right
			}
			if !value.Type().IsReference() {
				return nil, lowering.notYet(node, "null comparison with a scalar")
			}
			operand := node.AsBinaryExpression().Left
			if leftNull {
				operand = node.AsBinaryExpression().Right
			}
			test := ir.Expression(ir.IsNull{
				Value: value,
				AlwaysFalse: value.Type() != ir.Union &&
					!lowering.includesNull(lowering.checker.GetTypeAtLocation(operand)),
			})
			if operator == ast.KindExclamationEqualsEqualsToken {
				test = ir.Unary{Operator: ir.Not, Operand: test}
			}
			return test, nil
		}
		// x === undefined tests for a missing reference, whatever x's type.
		_, leftUndefined := left.(ir.Undefined)
		_, rightUndefined := right.(ir.Undefined)
		if leftUndefined != rightUndefined {
			value := left
			if leftUndefined {
				value = right
			}
			if pair := ir.Maybe(value.Type()); pair.IsMaybe() {
				// A number or a boolean the checker narrowed to present: still evaluated, and never
				// undefined.
				value = fit(value, pair)
			}
			if !value.Type().IsReference() && !value.Type().IsMaybe() {
				return nil, lowering.notYet(node, "comparing a "+typeName(value.Type())+" with undefined")
			}
			test := ir.Expression(ir.IsUndefined{Value: value})
			operand := node.AsBinaryExpression().Left
			if leftUndefined {
				operand = node.AsBinaryExpression().Right
			}
			if value.Type() != ir.Union && lowering.includesNull(lowering.checker.GetTypeAtLocation(operand)) {
				test = ir.IsNull{Value: value, AlwaysFalse: true}
			}
			if operator == ast.KindExclamationEqualsEqualsToken {
				test = ir.Unary{Operator: ir.Not, Operand: test}
			}
			return test, nil
		}
	}
	if (operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken) &&
		left.Type() != right.Type() {
		// A pair against what it holds, or against another pair: each made a pair, and compared as one.
		if pair := left.Type(); pair.IsMaybe() && right.Type() == pair.Present() {
			right = fit(right, pair)
		} else if pair := right.Type(); pair.IsMaybe() && left.Type() == pair.Present() {
			left = fit(left, pair)
		} else if left.Type() == ir.Union || right.Type() == ir.Union {
			// A union against anything: both as unions, compared by member and value.
			left, right = fit(left, ir.Union), fit(right, ir.Union)
		}
	}
	if (operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken) &&
		left.Type() == right.Type() {
		if (lowering.functionIdentityType(lowering.checker.GetTypeAtLocation(node.AsBinaryExpression().Left)) ||
			lowering.functionIdentityType(
				lowering.checker.GetTypeAtLocation(node.AsBinaryExpression().Right),
			)) && lowering.hasGenericFunctionValues() {
			return nil, lowering.notYet(
				node,
				"function identity comparison in a program with specialized generic function values",
			)
		}
		lowered := ir.Equal
		if operator == ast.KindExclamationEqualsEqualsToken {
			lowered = ir.NotEqual
		}
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}
	if value, known := lowering.censusBooleanLogical(node, operator, left, right); known {
		return value, nil
	}
	if operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken {
		if !both(ir.Boolean) {
			of, typeOfError := lowering.typeOf(node)
			if typeOfError != nil {
				return nil, typeOfError
			}
			if from := left.Type(); from.IsReference() &&
				from != ir.String &&
				from != ir.Union &&
				from != ir.Weak &&
				of != from &&
				of != ir.Union &&
				!of.IsMaybe() {
				// An object, array, Map or record is falsy only when it's missing, and a bare pointer to
				// one is never null (one | null is boxed as a union). So left && x is x where the left
				// is present and undefined where it isn't, which a reference result holds as its null
				// pointer (holder.item && holder.item.name, a string | undefined). A number or boolean
				// result says the checker proved the left present (node && node.kind === 1), so the
				// missing arm runs only if a type was a lie somewhere: it stops there, loud, rather than
				// read as 0 or false. || would keep the left itself, which the result can't hold.
				leftNode := node.AsBinaryExpression().Left
				if operator == ast.KindAmpersandAmpersandToken && of.IsReference() && of != ir.Weak {
					return ir.Conditional{
						Condition: ir.Truthy{Value: left},
						WhenTrue:  fit(right, of),
						WhenNot:   ir.Undefined{Of: of},
						Of:        of,
					}, nil
				}
				if operator == ast.KindAmpersandAmpersandToken &&
					(of == ir.Number || of == ir.Boolean) &&
					!lowering.mayBeUndefined(leftNode) &&
					!lowering.includesNull(lowering.checker.GetTypeAtLocation(leftNode)) {
					var unreachable ir.Expression = ir.NumberConstant{}
					if of == ir.Boolean {
						unreachable = ir.BooleanConstant{}
					}
					stop := ir.Panic{Message: ir.StringConstant{
						Index: lowering.constant("compiler bug: a left of && proven present was missing"),
					}}
					return ir.Conditional{
						Condition: ir.Truthy{Value: left},
						WhenTrue:  fit(right, of),
						WhenNot:   ir.Effects{Body: []ir.Statement{stop}, Result: unreachable},
						Of:        of,
					}, nil
				}
				return nil, lowering.notYet(node, "a logical "+
					typeName(from)+
					" operand converted to "+
					typeName(of)+
					" without a proven result representation")
			}
			if left.Type() == ir.String && of != ir.String && of != ir.Union {
				// A falsy string is "" (or undefined), which only a string or a union result can hold.
				// The checker types string && x as x alone where a brand hides the string's falsy part
				// (tsc's __String), so Node's "" would reach a slot that can't say it.
				return nil, lowering.notYet(
					node,
					"a logical string operand converted to "+typeName(of)+", whose falsy \"\" the result can't hold",
				)
			}
			return ir.Logical{
				Left:       left,
				Right:      fit(right, of),
				Of:         of,
				KeepTruthy: operator == ast.KindBarBarToken,
			}, nil
		}
		lowered := ir.And
		if operator == ast.KindBarBarToken {
			lowered = ir.Or
		}
		return ir.Binary{Operator: lowered, Left: left, Right: right}, nil
	}

	return nil, lowering.notYet(node, describe(node)+" with a "+typeName(left.Type())+" and a "+typeName(right.Type()))
}

// concatenated writes the primitive operands of string addition as JavaScript does.
// Objects retain their representation, so unsupported ToPrimitive remains NotYet.
func (lowering *lowering) concatenated(value ir.Expression) ir.Expression {
	switch value.(type) {
	case ir.Null:
		return ir.StringConstant{Index: lowering.constant("null")}
	case ir.Undefined:
		return ir.StringConstant{Index: lowering.constant("undefined")}
	}
	switch value.Type() {
	case ir.Number:
		return ir.NumberToString{Value: value}
	case ir.Boolean:
		return ir.BooleanToString{Value: value}
	case ir.MaybeNumber, ir.MaybeBoolean:
		return ir.MaybeToString{Value: value}
	}
	return value
}

// spelled is a string as + and a template write it: one that may be missing (a null reference) is
// written "undefined", as JavaScript writes it.
func (lowering *lowering) spelled(node *ast.Node, value ir.Expression) ir.Expression {
	if value.Type() != ir.String ||
		!(lowering.includesUndefined(lowering.checker.GetTypeAtLocation(node)) ||
			lowering.narrowedAway(ast.SkipParentheses(node))) {
		return value
	}
	return ir.Coalesce{Value: value, Fallback: ir.StringConstant{Index: lowering.constant("undefined")}, Of: ir.String}
}

// template lowers a template literal to a Concat, writing each value as String() would.
func (lowering *lowering) template(node *ast.Node) (ir.Expression, error) {
	template := node.AsTemplateExpression()
	parts := []ir.Expression{}
	if head := template.Head.Text(); head != "" {
		parts = append(parts, ir.StringConstant{Index: lowering.constant(head)})
	}
	for _, span := range template.TemplateSpans.Nodes {
		value, expressionError := lowering.expression(span.AsTemplateSpan().Expression)
		if expressionError != nil {
			return nil, expressionError
		}
		switch value.Type() {
		case ir.Number:
			value = ir.NumberToString{Value: value}
		case ir.Boolean:
			value = ir.BooleanToString{Value: value}
		case ir.MaybeNumber, ir.MaybeBoolean:
			value = ir.MaybeToString{Value: value}
		case ir.Union:
			observed := lowering.checker.GetTypeAtLocation(span.AsTemplateSpan().Expression)
			switch {
			case lowering.writable(observed):
				value = ir.UnionToString{Value: value}
			case dynamicObjectType(observed):
				// As String's: only an unknown or an object (stringConversion).
				value = lowering.toStringHint(value)
			default:
				return nil, lowering.notYet(
					span,
					"a template interpolating a union with an object, an array, a map or a function in it",
				)
			}
		case ir.String:
			value = lowering.spelled(span.AsTemplateSpan().Expression, value)
		default:
			flags := lowering.checker.GetTypeAtLocation(span.AsTemplateSpan().Expression).Flags()
			spelling := ""
			if flags == checker.TypeFlagsNull {
				spelling = "null"
			} else if flags&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 &&
				flags & ^(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) == 0 {
				spelling = "undefined"
			}
			if spelling == "" {
				return nil, lowering.notYet(
					span,
					"a template interpolating an object, an array, a map, a function or undefined",
				)
			}
			value = ir.Effects{
				Body:   []ir.Statement{ir.Evaluate{Value: value}},
				Result: ir.StringConstant{Index: lowering.constant(spelling)},
			}
		}
		parts = append(parts, value)
		if literal := span.AsTemplateSpan().Literal.Text(); literal != "" {
			parts = append(parts, ir.StringConstant{Index: lowering.constant(literal)})
		}
	}
	return ir.Concat{Parts: parts}, nil
}

func (lowering *lowering) conditional(node *ast.Node) (ir.Expression, error) {
	if branch := lowering.literalCallableBranch(node); branch != nil {
		return lowering.expression(branch)
	}
	conditional := node.AsConditionalExpression()
	condition, conditionError := lowering.condition(conditional.Condition)
	if conditionError != nil {
		return nil, conditionError
	}
	if symbol := lowering.hostBindingGuard(conditional.Condition); symbol != nil {
		return lowering.hostBindingConditional(node, condition, symbol)
	}
	if lowering.isNever(conditional.WhenTrue) || lowering.isNever(conditional.WhenFalse) {
		return lowering.conditionalNever(node, condition)
	}
	whenTrue, expressionError := lowering.expression(conditional.WhenTrue)
	if expressionError != nil {
		return nil, expressionError
	}
	whenNot, expressionError := lowering.expression(conditional.WhenFalse)
	if expressionError != nil {
		return nil, expressionError
	}
	if whenTrue.Type() != whenNot.Type() && (lowering.acceptsUndefined(node) || lowering.placeholderAllowsUnset(node)) {
		// A stale narrowing in either branch may still hold undefined. Keep that representation
		// when the whole conditional is observed or written into a slot that accepts it.
		if pair := whenTrue.Type(); pair.IsMaybe() && whenNot.Type() == pair.Present() {
			whenNot = fit(whenNot, pair)
		} else if pair := whenNot.Type(); pair.IsMaybe() && whenTrue.Type() == pair.Present() {
			whenTrue = fit(whenTrue, pair)
		}
	}
	if whenTrue.Type() != whenNot.Type() {
		// flag ? 1 : undefined is number | undefined, and flag ? 1 : 'one' a union: each branch made
		// one.
		if of, typeOfError := lowering.typeOf(
			node,
		); typeOfError == nil && (of.IsMaybe() || of == ir.Union || of.IsReference()) {
			whenTrue, whenNot = fit(whenTrue, of), fit(whenNot, of)
		}
	}
	if whenTrue.Type() != whenNot.Type() {
		// flag ? text : undefined is a reference that may be missing, held as text is, with undefined
		// the null one: the checker's type for the whole says which.
		_, trueUndefined := whenTrue.(ir.Undefined)
		_, notUndefined := whenNot.(ir.Undefined)
		present, presentNode := whenTrue, conditional.WhenTrue
		if trueUndefined {
			present, presentNode = whenNot, conditional.WhenFalse
		}
		// flag ? {} : undefined in a record slot builds its literal as the record (recordLiteralElement),
		// where the checker's own type for the whole is {} | undefined: the record is what it holds.
		builtAsRecord := ast.SkipParentheses(presentNode).Kind == ast.KindObjectLiteralExpression &&
			lowering.recordLiteralElement(ast.SkipParentheses(presentNode)) != nil
		if of, typeOfError := lowering.typeOf(node); typeOfError == nil &&
			trueUndefined != notUndefined &&
			present.Type().IsReference() &&
			(of == present.Type() || builtAsRecord) {
			if builtAsRecord {
				of = present.Type()
			}
			return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot, Of: of}, nil
		}
		return nil, lowering.notYet(node, "a conditional whose branches have different types")
	}
	return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot}, nil
}

func typeName(valueType ir.Type) string {
	if valueType.IsTypedArray() {
		return typedArrayName(valueType)
	}
	switch valueType {
	case ir.Number:
		return "number"
	case ir.Boolean:
		return "boolean"
	case ir.String:
		return "string"
	case ir.MaybeNumber:
		return "number | undefined"
	case ir.MaybeBoolean:
		return "boolean | undefined"
	case ir.Union:
		return "union of differently held members"
	}
	return "value"
}

// writable reports whether every member of a union is one a template writes: a number, a boolean, a
// string or undefined.
func (lowering *lowering) writable(proven *checker.Type) bool {
	// Judge the members this instantiation stores, rather than the generic binder.
	proven = lowering.concrete(proven)
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		memberType, isKnown := lowering.representation(member)
		if !isKnown || (memberType != ir.Number && memberType != ir.Boolean && memberType != ir.String) {
			return false
		}
	}
	return true
}

// slotless reports whether a value of the type can't yet be held in one word: a field, an element, a
// map's value, a cell, or a function value's argument or result. number | undefined is packed into
// one (a reserved NaN is undefined); boolean | undefined has a tagged byte for object fields
// (censusFieldSlotless), but remains unsupported in the other slot contexts, and a
// Union has to be boxed on its way in, which stage 0 does only where a variable, a parameter or a
// result takes one.
func slotless(valueType ir.Type) bool {
	return valueType == ir.MaybeBoolean || valueType == ir.Union
}

// A fitted Union occupies one counted reference. Keep the general slot refusal
// for callers that have not arranged boxing, and optional booleans for Map slots.
func boxedSlot(valueType ir.Type) bool {
	return !slotless(valueType) || valueType == ir.Union
}

// call lowers a call to one of the module's functions, or to a function value.
func (lowering *lowering) call(node *ast.Node) (ir.Expression, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if function, isNested := lowering.nestedGeneratorCallee(callee); isNested {
		// As a module generator's call below: its consumers lower it before this, and one passed
		// straight into a one-body function's Iterable<T> parameter goes in through IterableOf
		// (#qk8rztp S3a).
		if lowering.iterableArgument(node) {
			return lowering.iterableOfGenerator(node, function)
		}
		return nil, lowering.generatorNotYet(
			node,
			"whose generator is used other than by a for-of over its call or a call as a statement",
		)
	}
	if generatorInPlace(callee) {
		// A generator written in place and called (#qk8rztp S3b): its for-of and its call as a
		// statement lower it before this, and the one-read Iterable<T> parameter takes it here, as a
		// named generator's call. Any other use is a value no slot can hold yet, stopped before its
		// body is lowered.
		if !lowering.iterableArgument(node) {
			return nil, lowering.inPlaceNotYet(node)
		}
		function, inPlaceGeneratorError := lowering.inPlaceGenerator(callee)
		if inPlaceGeneratorError != nil {
			return nil, inPlaceGeneratorError
		}
		return lowering.iterableOfGenerator(node, function)
	}
	if direct := lowering.nestedSibling(callee); direct >= 0 {
		arguments := []ir.Expression{}
		for _, argument := range call.Arguments.Nodes {
			value, expressionError := lowering.expression(argument)
			if expressionError != nil {
				return nil, expressionError
			}
			position := len(arguments)
			if position < len(lowering.result.Functions[direct].Parameters) {
				value = fit(value, lowering.result.Locals[lowering.result.Functions[direct].Parameters[position]].Type)
			}
			arguments = append(arguments, value)
		}
		carrier := ir.Expression(ir.ClosureSelf{})
		if lowering.function.NestedParent != lowering.result.Functions[direct].NestedParent {
			// Anonymous captures have their own layout. Materialize a temporary carrier
			// for the target layout; do not capture a sibling's canonical function value.
			carrier = ir.MakeClosure{Function: direct}
		}
		return ir.CallClosure{
			Closure:   carrier,
			Direct:    direct + 1,
			Arguments: arguments,
			Returns:   lowering.result.Functions[direct].Returns,
		}, nil
	}
	qualified := lowering.namespaceMember(callee)
	if declaration, isGeneric := lowering.generics[lowering.symbol(
		callee,
	)]; (ast.IsIdentifier(callee) || qualified) && isGeneric {
		instance, instantiateFunctionError := lowering.instantiateFunction(node, declaration)
		if instantiateFunctionError != nil {
			return nil, instantiateFunctionError
		}
		if lowering.result.Functions[instance].Generator {
			// A generic generator's instance, called anywhere but by a consumer, stops as a
			// non-generic generator's call does below, and goes into an Iterable<T> parameter as
			// one does.
			if lowering.iterableArgument(node) {
				return lowering.iterableOfGenerator(node, instance)
			}
			return nil, lowering.generatorNotYet(
				node,
				"whose generator is used other than by a for-of over its call or a call as a statement",
			)
		}
		if lowering.result.Functions[instance].Closure {
			arguments := []ir.Expression{}
			for position, argument := range call.Arguments.Nodes {
				value, expressionError := lowering.expression(argument)
				if expressionError != nil {
					return nil, expressionError
				}
				if position < len(lowering.result.Functions[instance].Parameters) {
					value = fit(
						value,
						lowering.result.Locals[lowering.result.Functions[instance].Parameters[position]].Type,
					)
				}
				arguments = append(arguments, value)
			}
			value := ir.CallClosure{
				Closure:   ir.MakeClosure{Function: instance},
				Arguments: arguments,
				Returns:   lowering.result.Functions[instance].Returns,
			}
			return lowering.namespaceReadyCall(callee, value), nil
		}
		value, callFunctionError := lowering.callFunction(call, instance)
		if callFunctionError == nil {
			value = lowering.namespaceReadyCall(callee, value)
		}
		return value, callFunctionError
	}
	function, isFunction := lowering.functions[lowering.symbol(callee)]
	if (!ast.IsIdentifier(callee) && !qualified) || !isFunction {
		if calleeType, _ := lowering.representation(
			lowering.checker.GetTypeAtLocation(callee),
		); calleeType == ir.Closure {
			return lowering.callClosure(node)
		}
		return nil, lowering.notYet(node, "a call to "+describe(callee))
	}
	if lowering.result.Functions[function].Generator {
		// The two consumers that keep a generator to themselves lower its call before this, and one
		// passed straight into a one-body function's Iterable<T> parameter goes in through
		// IterableOf (#qk8rztp S2c-1); a generator anywhere else is a value no slot can hold yet.
		if lowering.iterableArgument(node) {
			return lowering.iterableOfGenerator(node, function)
		}
		return nil, lowering.generatorNotYet(
			node,
			"whose generator is used other than by a for-of over its call or a call as a statement",
		)
	}
	value, callFunctionError := lowering.callFunction(call, function)
	if callFunctionError == nil {
		value = lowering.namespaceReadyCall(callee, value)
	}
	return value, callFunctionError
}

// callFunction lowers a call's arguments, in order, and the call to function.
func (lowering *lowering) callFunction(call *ast.CallExpression, function int) (ir.Expression, error) {
	arguments, spread, callArgumentsError := lowering.callArguments(call.Arguments.Nodes)
	if callArgumentsError != nil {
		return nil, callArgumentsError
	}
	lowering.fitCallArguments(function, arguments, spread)
	return lowering.censusOverloadResult(call, ir.Call{
		Function:  function,
		Arguments: arguments,
		Spread:    spread,
		Returns:   lowering.result.Functions[function].Returns,
	})
}

// coalesce lowers value ?? fallback, and value ?? panic('why'), evaluating the right side only when
// the left is missing.
func (lowering *lowering) coalesce(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	value, expressionError := lowering.expression(binary.Left)
	if expressionError != nil {
		return nil, expressionError
	}
	// What ?? makes is the checker's: the left side's present members and the right side's, which
	// may be held differently (text ?? count is string | number, a Union).
	of, typeOfError := lowering.typeOf(node)
	if typeOfError != nil && lowering.placeholderReadOrigin(ast.SkipParentheses(binary.Left)) != "" {
		declared, known := lowering.representation(
			lowering.checker.GetNonNullableType(
				lowering.checker.GetTypeOfSymbol(lowering.symbol(ast.SkipParentheses(binary.Left))),
			),
		)
		fallback, fallbackKnown := lowering.representation(lowering.checker.GetTypeAtLocation(binary.Right))
		if known && fallbackKnown {
			of, typeOfError = declared, nil
			if declared != fallback {
				of = ir.Union
			}
		}
	}
	if typeOfError != nil {
		return nil, typeOfError
	}
	if present := value.Type(); !present.IsMaybe() && !present.IsReference() {
		// A value that can't be missing: ?? never runs its right side.
		return fit(value, of), nil
	}
	right := ast.SkipParentheses(binary.Right)
	if right.Kind == ast.KindCallExpression &&
		lowering.isPreludeFunction(right.AsCallExpression().Expression, "panic") &&
		len(right.AsCallExpression().Arguments.Nodes) == 1 {
		message, expressionError := lowering.expression(right.AsCallExpression().Arguments.Nodes[0])
		if expressionError != nil {
			return nil, expressionError
		}
		return ir.Coalesce{Value: value, Panic: message, Of: of}, nil
	}
	if lowering.isNever(binary.Right) {
		fallback, neverValueError := lowering.neverValue(binary.Right, of)
		if neverValueError != nil {
			return nil, neverValueError
		}
		return ir.Coalesce{Value: value, Fallback: fallback, Of: of}, nil
	}
	fallback, expressionError := lowering.expression(binary.Right)
	if expressionError != nil {
		return nil, expressionError
	}
	fallback = fit(fallback, of)
	if _, isUndefined := fallback.(ir.Undefined); fallback.Type() != of && !(isUndefined && of.IsReference()) {
		return nil, lowering.notYet(node, "?? whose sides have different types")
	}
	return ir.Coalesce{Value: value, Fallback: fallback, Of: of}, nil
}

// closure lowers an arrow function to a function of its own and the closure that captures it.
func (lowering *lowering) closure(node *ast.Node) (ir.Expression, error) {
	index := len(lowering.result.Functions)
	lowering.result.Functions = append(lowering.result.Functions, ir.Function{Name: "closure", Closure: true})
	lowering.closureRecords = append(
		lowering.closureRecords,
		closureRecord{proven: lowering.concrete(lowering.checker.GetTypeAtLocation(node)), function: index, node: node},
	)
	if lowering.instance != nil {
		lowering.instance.templates = append(
			lowering.instance.templates,
			template{closure: index, proven: lowering.checker.GetTypeAtLocation(node), isClosure: true},
		)
	}
	lowering.closures = append(lowering.closures, index)
	lowerFunctionError := lowering.lowerFunction(index, node, -1)
	lowering.closures = lowering.closures[:len(lowering.closures)-1]
	if lowerFunctionError != nil {
		return nil, lowerFunctionError
	}
	return ir.MakeClosure{Function: index}, nil
}

// functionValue makes one stable adapter per named function. It binds optional
// and rest parameters and forwards the original count, even though its ordinary
// parameters have been padded or collected before reaching the target.
func (lowering *lowering) functionValue(node *ast.Node, target int) (ir.Expression, error) {
	symbol := lowering.symbol(node)
	if node.Kind == ast.KindShorthandPropertyAssignment {
		symbol = lowering.checker.GetShorthandAssignmentValueSymbol(node)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = lowering.checker.GetAliasedSymbol(symbol)
		}
		symbol = lowering.checker.GetExportSymbolOfSymbol(symbol)
	}
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration &&
				declaration.Body() == nil &&
				lowering.censusImplementation(declaration) != nil {
				return nil, lowering.notYet(node, "an overloaded function as a value")
			}
		}
	}
	if lowering.result.Functions[target].Generator {
		return nil, lowering.generatorNotYet(node, "used as a value")
	}
	if held, isMade := lowering.forwarders[target]; isMade {
		return ir.Read{Local: held, Of: ir.Closure}, nil
	}
	for _, parameter := range symbol.Declarations[0].Parameters() {
		if parameter.Name().Text() == "this" {
			return nil, lowering.notYet(
				parameter,
				"a function with a this parameter used as a value; pass the receiver explicitly or use an arrow",
			)
		}
	}
	callee := lowering.result.Functions[target]
	if censusCallableSlotless(callee.Returns) {
		return nil, lowering.notYet(node, "a function value returning "+typeName(callee.Returns))
	}
	index := len(lowering.result.Functions)
	forwarder := ir.Function{
		SourceLength:      callee.SourceLength,
		Name:              callee.Name + "_value",
		Source:            callee.Source,
		Closure:           true,
		Returns:           callee.Returns,
		RestElement:       callee.RestElement,
		ForwardsArguments: target + 1,
	}
	forwarder.Source.Role = "value_adapter"
	forwarder.Source.Declaration = append(append([]string{}, callee.Source.Declaration...), "value_adapter")
	arguments := []ir.Expression{}
	for _, parameter := range callee.Parameters {
		declared := lowering.result.Locals[parameter]
		if censusCallableSlotless(declared.Type) {
			return nil, lowering.notYet(node, "a function value taking "+typeName(declared.Type))
		}
		local := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: declared.Name, Type: declared.Type, Function: index},
		)
		forwarder.Parameters = append(forwarder.Parameters, local)
		arguments = append(arguments, ir.Read{Local: local, Of: declared.Type})
	}
	call := ir.Call{
		Function:     target,
		Arguments:    arguments,
		Returns:      callee.Returns,
		ForwardCount: true,
		RestPacked:   true,
	}
	if callee.Returns == 0 {
		forwarder.Body = []ir.Statement{ir.Evaluate{Value: call}}
	} else {
		forwarder.Body = []ir.Statement{ir.Return{Value: call}}
	}
	lowering.result.Functions = append(lowering.result.Functions, forwarder)
	// A checked predicate's value checks its own claim at every call through it (#vdt9f1q).
	checked, isChecked, checkedPredicateForwarderError := lowering.checkedPredicateForwarder(node, symbol, index, call)
	if checkedPredicateForwarderError != nil {
		return nil, checkedPredicateForwarderError
	}
	if isChecked {
		lowering.result.Functions[index].Body = checked
	}
	lowering.closureRecords = append(
		lowering.closureRecords,
		closureRecord{proven: lowering.concrete(lowering.checker.GetTypeAtLocation(node)), function: index, node: node},
	)
	// One function value for the function, made before anything runs and held by a global of its
	// own, so reading the function twice gives the same value, === as JavaScript's.
	held := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "value_adapter", Source: forwarder.Source, Type: ir.Closure, Global: true, Function: -1},
	)
	lowering.forwarderValues = append(
		lowering.forwarderValues,
		ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}},
	)
	if lowering.forwarders == nil {
		lowering.forwarders = map[int]int{}
	}
	lowering.forwarders[target] = held
	return ir.Read{Local: held, Of: ir.Closure}, nil
}

// optionalCall refuses a call in an optional chain, text?.toUpperCase() or run?.(): the receiver has
// to be evaluated once and tested before the call, which stage 0 doesn't lower yet for the library's
// methods or a function value called directly. Lowered as a plain call, it ran the method on
// undefined. The one form lowered is object?.method(...) on a method the program declares (a
// class's, an interface's, or a function value in a field): the call goes through the object's
// methods, which stops at undefined as JavaScript's chain does (ir.Property's Method).
func (lowering *lowering) optionalCall(call *ast.Node) error {
	if call.Flags&ast.NodeFlagsOptionalChain == 0 || lowering.isOptionalJoin(call) {
		return nil
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression &&
		callee.AsPropertyAccessExpression().QuestionDotToken != nil &&
		call.AsCallExpression().QuestionDotToken == nil {
		if method := lowering.checker.GetSymbolAtLocation(callee); method != nil &&
			len(method.Declarations) > 0 &&
			!load.IsLibrary(ast.GetSourceFileOfNode(method.Declarations[0])) {
			return nil
		}
	}
	return lowering.notYet(call, "a call through ?. (an optional call)")
}

// callClosure lowers a call through a function value.
func (lowering *lowering) callClosure(node *ast.Node) (ir.Expression, error) {
	signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(node.AsCallExpression().Expression),
		checker.SignatureKindCall,
	)
	if len(signatures) == 1 && lowering.censusNeverRestSignature(signatures[0]) {
		// never[] admits a zero-argument call in TypeScript. The erased slot does
		// not retain a source signature to prove its required arguments or ABI.
		return nil, lowering.notYet(node, "a call through an erased never-rest callable marker")
	}
	// A hand-driven `x.next()` over an Iterable view parameter: property lets its callee past the
	// library-member guard, and handNextView records the call for checkIterableViews to prove.
	hand := lowering.handDrivenNext(node)
	if hand && len(node.AsCallExpression().Arguments.Nodes) != 0 {
		return nil, lowering.notYet(node, "next(x) through a view, whose value is sent in (#qk8rztp S2c-2)")
	}
	if hand {
		lowering.handNext = node.AsCallExpression().Expression
	}
	closure, expressionError := lowering.expression(node.AsCallExpression().Expression)
	lowering.handNext = nil
	if expressionError != nil {
		return nil, expressionError
	}
	if hand {
		if handNextViewError := lowering.handNextView(node, closure); handNextViewError != nil {
			return nil, handNextViewError
		}
	}
	if property, isProperty := closure.(ir.Property); isProperty {
		// object.name(...) through an interface: the object may be a class's, whose methods aren't
		// fields (ir.Property's Method).
		property.Method = true
		closure = property
	}
	arguments, spread, callArgumentsError := lowering.callArguments(node.AsCallExpression().Arguments.Nodes)
	if callArgumentsError != nil {
		return nil, callArgumentsError
	}
	if property, ok := closure.(ir.Property); ok && property.ViewContract > 0 && len(spread) != 0 {
		contract := lowering.result.ViewContracts[property.ViewContract-1]
		if contract.Of == ir.Closure &&
			contract.Unsupported == "" &&
			(contract.Kind == ir.ViewUnion || contract.Kind == ir.ViewCallable) {
			return nil, lowering.notYet(node, "spread arguments through a checked callable view")
		}
	}
	var returns ir.Type
	if result := lowering.concrete(
		lowering.checker.GetTypeAtLocation(node),
	); result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
		var isKnown bool
		if returns, isKnown = lowering.representation(result); !isKnown {
			return nil, lowering.notYet(node, "a call returning "+lowering.checker.TypeToString(result))
		}
	}
	// Each argument is made what the function value takes: a number or undefined where it takes
	// number | undefined is packed as one.
	if signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(node.AsCallExpression().Expression),
		checker.SignatureKindCall,
	); len(signatures) == 1 {
		parameters := signatures[0].Parameters()
		position := 0
		expanded := false
		for index := range arguments {
			if len(spread) > index && spread[index] {
				expanded = true
				continue
			}
			if len(parameters) == 0 {
				break
			}
			parameter := parameters[min(position, len(parameters)-1)]
			rest := len(parameter.Declarations) > 0 &&
				parameter.Declarations[0].Kind == ast.KindParameter &&
				parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil
			proven := lowering.checker.GetTypeOfSymbol(parameter)
			if rest {
				// Only a reference has type arguments to read. The checker dereferences nil for a rest
				// parameter typed as anything else, a union of tuples or a type parameter (#egxccg8).
				if proven.Flags()&checker.TypeFlagsObject == 0 ||
					proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
					return nil, lowering.notYet(
						node,
						"a call through a function value whose rest parameter isn't an array",
					)
				}
				elements := lowering.checker.GetTypeArguments(proven)
				if len(elements) == 1 {
					proven = elements[0]
				}
			} else if expanded || position >= len(parameters) {
				position++
				continue
			}
			if takes, known := lowering.representation(proven); known {
				// The checker describes a defaulted parameter by its present type,
				// but its incoming slot must preserve undefined until the body runs.
				if !rest &&
					len(parameter.Declarations) > 0 &&
					parameter.Declarations[0].Kind == ast.KindParameter &&
					parameter.Declarations[0].AsParameterDeclaration().Initializer != nil {
					takes = ir.Maybe(takes)
				}
				arguments[index] = fit(arguments[index], takes)
			}
			position++
		}
	}
	for _, argument := range arguments {
		if censusCallableSlotless(argument.Type()) {
			return nil, lowering.notYet(node, "passing "+typeName(argument.Type())+" to a function value")
		}
	}
	if censusCallableSlotless(returns) {
		return nil, lowering.notYet(node, "a function value returning "+typeName(returns))
	}
	// A call through a checked function-type claim checks it (predicates_function_type.go).
	return lowering.checkedFunctionTypeCall(node, closure, ir.CallClosure{
		Closure:   closure,
		Arguments: arguments,
		Spread:    spread,
		FunctionType: int(
			lowering.concrete(lowering.checker.GetTypeAtLocation(node.AsCallExpression().Expression)).Id(),
		),
		Returns: returns,
	})
}

// Optional booleans have a three-state byte in the function-call ABI.
// General union arguments still need a representation adapter.
func argumentSlotless(valueType ir.Type) bool { return valueType == ir.Union }

// valueSlotStops is lowering's own judgment of a value seen as the slot it goes into, beside the refusal
// registry's: valueSlotTags, and valueSlotKeeps for a value held by reference. expression runs the two
// halves where it lowers the value, against the value's own slot. A write that can land under a slot it
// isn't lowered against, a computed-key record write under a declared property, runs both against that
// slot too (namedWriteStop, #b6m00y7), so one list of these checks serves both.
func (lowering *lowering) valueSlotStops(node *ast.Node, slot *checker.Type, reference bool) error {
	if valueSlotTagsError := lowering.valueSlotTags(node, slot); valueSlotTagsError != nil {
		return valueSlotTagsError
	}
	if reference {
		return lowering.valueSlotKeeps(node, slot)
	}
	return nil
}

// valueSlotTags is a value's null and undefined tags against the slot's, and an unknown view.
func (lowering *lowering) valueSlotTags(node *ast.Node, slot *checker.Type) error {
	if lowering.nullableObjectViewHazard(node, slot) {
		return lowering.notYet(
			node,
			"sharing nullable object storage with different null/undefined tags; copy values into tagged storage",
		)
	}
	return lowering.unknownView(node, lowering.checker.GetTypeAtLocation(node), slot)
}

// valueSlotKeeps is a value held by reference against the slot it goes into: a Buffer or Hash seen as
// another type, a host object's view (#8k12t4s), a weak or strong keep, a typed array's buffer and a
// tuple seen as an array. A literal is made as its slot, so it's never judged here.
func (lowering *lowering) valueSlotKeeps(node *ast.Node, slot *checker.Type) error {
	if literal := ast.SkipParentheses(
		node,
	).Kind; literal == ast.KindArrayLiteralExpression || literal == ast.KindObjectLiteralExpression {
		return nil
	}
	// The checker lets { v: Box } be seen as { v: Weak<Box> } and back, an array of Box as one of
	// Weak<Box>, and (x: Weak<Box>) => ... as (x: Box) => ...; but one keeps a handle where the
	// other keeps the target, so the same object, array or function can't be both. A tuple is held
	// as an object, so it can't be seen as an array either, here or anywhere inside. A literal is
	// made as the type it's written into, so it never differs.
	if own := lowering.checker.GetTypeAtLocation(
		node,
	); !lowering.nodeBufferView(lowering.present(own), lowering.present(slot)) &&
		!lowering.nodeBufferReadArgument(node) &&
		!lowering.nodeFSFileBufferArgument(node) {
		return lowering.notYet(node, "Buffer or Hash viewed as another object type (native host internal slots)")
	}
	// Any other host object may be seen through a type only as far as its registration names
	// the members that type reaches (#8k12t4s); Buffer and Hash are refused whole above.
	if own := lowering.checker.GetTypeAtLocation(
		node,
	); !lowering.nodeBufferType(lowering.present(own), "Buffer") &&
		!lowering.nodeBufferType(lowering.present(own), "Hash") {
		if reached := lowering.nodeHostViewAt(own, slot); reached != "" {
			return lowering.nodeHostViewRefusal(node, own, slot, reached)
		}
	}
	if own := lowering.contextualGenericType(
		node,
		lowering.checker.GetTypeAtLocation(node),
	); !lowering.typedArraySetArgument(node) &&
		!lowering.nodeBufferReadArgument(node) &&
		!lowering.nodeFSFileBufferArgument(node) &&
		!lowering.sameKeeping(own, slot, map[[2]*checker.Type]bool{}) {
		if lowering.typedArrayKind(lowering.checker.GetNonNullableType(own)) != 0 {
			return lowering.notYet(
				node,
				"a typed array seen through a structural view that loses its buffer representation",
			)
		}
		return lowering.notYet(node, "a "+
			lowering.checker.TypeToString(own)+
			" seen as a "+
			lowering.checker.TypeToString(slot)+
			" (one keeps something weakly that the other keeps strongly)")
	} else if tuple, array := lowering.tupleSeenAsArray(own, slot, map[[2]*checker.Type]bool{}); tuple != nil {
		return lowering.notYet(node, "a "+
			lowering.checker.TypeToString(tuple)+
			" seen as a "+
			lowering.checker.TypeToString(array)+
			" (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]])")
	}
	return nil
}
