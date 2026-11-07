package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Mutable locations are invariant in 0.1 (docs/0.1.md, adamic/invariant-mutable): tsc relates an
// array's elements, a property and a map's values covariantly, so a Dog[] is accepted as an Animal[],
// and the wider name can then write a cat where the dogs are read. Stage 0 doesn't run cohere's lint,
// so it refuses the same views itself, before lowering: every place a value goes into a typed slot
// (an initializer, an assignment, an argument, a return, a literal's property or element), the value's
// type is walked against the slot's, all the way down, as the Weak view check walks them.
//
// At each mutable slot the wider side can write, what it writes must be something the narrower side
// can read: the slot's type in the target must be assignable to its type in the source. tsc proved
// the other direction when it accepted the site, so together the slot is invariant. A readonly slot
// only reads, and stays covariant.
//
// In .a, an index signature is a slot too, under every name a declared property doesn't take: a
// declared property seen through a wider one can be written what it can't hold, and a dictionary
// seen through a declared property can be read as what it doesn't hold (records_declared_keys.go).
//
// Two shapes reach a mutable array where tsc's relation can't be walked part by part, and each is
// judged as cohere's rule judges it (cohere 838c6516):
//   - A type parameter target (const pack: Pack = narrow, Narrow extends Pack extends Animal[]): what
//     Pack holds is no type to compare against until it's instantiated, so a type parameter whose
//     constraint can be written is one mutable slot, and only the parameter itself (or it narrowed,
//     as T & {}) fills it.
//   - An intersection with an array (Dog[] & { tag } into Animal[] & { tag }): its array, tuple and
//     container members are paired member by member, as well as its properties.

// widening is a view refused: the parts of the source and the target at the slot where the target can
// write what the source can't read.
type widening struct {
	source, target *checker.Type

	// readonlyField names a readonly field seen as a writable one, when that's the slot.
	readonlyField string
	enum          bool

	// parameter is a function's parameter of type source seen as taking target, which it can't.
	parameter bool

	// declaredKey names a declared property an index signature reaches in .a (records_declared_keys.go):
	// the target's index signature can write target under it where source is declared, or, when
	// dictionary is set, the source's index signature holds source where the target reads it as target.
	declaredKey string
	dictionary  bool

	// viewed is the value seen, when it's one branch of a conditional whose own type tsc reduced to
	// the wider side (flag ? options : dict is a Dict), so a refusal can name it.
	viewed *checker.Type
}

// widened finds the first mutable slot at which a value of type from, seen as type to, could be
// written something it can't hold, or returns nil.
func (lowering *lowering) widened(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	from, to = lowering.withoutUndefined(from), lowering.withoutUndefined(to)
	if from == to || visited[[2]*checker.Type{from, to}] || from.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	visited[[2]*checker.Type{from, to}] = true
	if lowering.openNumericEnumType(to) {
		for _, member := range lowering.definedMembers(from) {
			if !lowering.enumAssignable(member, to) {
				return &widening{source: from, target: to, enum: true}
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsNumberLiteral != 0 && lowering.openNumericEnumType(from) {
		return &widening{source: from, target: to, enum: true}
	}
	if members := lowering.definedMembers(from); len(members) > 1 {
		// A union is any one of its members, each seen as the target.
		for _, member := range members {
			if found := lowering.widened(member, to, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if members := lowering.definedMembers(to); len(members) > 1 {
		// Seen as a union, a value is seen as each member it can be, and narrowing picks any of them
		// to write through. For a declared key, a value entering a union that names its own type stays
		// that member (#6c4swvt G2, records.go), and seeing the union as another member later is a view
		// of its own, judged there.
		if lowering.declaredKeys {
			for _, member := range members {
				if member == from {
					return nil
				}
			}
		}
		for _, member := range members {
			if !lowering.checker.IsTypeAssignableTo(from, member) {
				continue
			}
			if found := lowering.widened(from, member, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if lowering.narrowedFrom(from, to) {
			return nil
		}
		if constraint := lowering.checker.GetBaseConstraintOfType(
			to,
		); constraint != nil && lowering.canWrite(constraint, map[*checker.Type]bool{}) {
			return &widening{source: from, target: to}
		}
		return nil
	}
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		// What it can be is what its constraint allows: seen as anything wider, it's that seen so.
		constraint := lowering.checker.GetBaseConstraintOfType(from)
		if constraint == nil {
			return nil
		}
		return lowering.widened(constraint, to, visited)
	}
	if !lowering.enumAssignable(from, to) {
		return &widening{source: from, target: to, enum: true}
	}
	if !lowering.structured(from) || !lowering.structured(to) {
		// A class instance target is walked as any object is: its fields are slots as a literal's
		// are, so Box<Dog> seen as Box<Animal>, or a plain object seen as a class, is judged by them.
		return nil
	}
	if lowering.declaredKeys {
		if found := lowering.declaredKeyWidened(from, to, nil, false); found != nil {
			return found
		}
	}
	if source, target := lowering.recordElement(from), lowering.recordElement(to); source != nil && target != nil {
		if !lowering.checker.IsTypeAssignableTo(target, source) {
			return &widening{source: source, target: target}
		}
		return lowering.widened(source, target, visited)
	}
	fromSignatures := lowering.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := lowering.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		// A function seen as another is handed the other's arguments, and its results are seen as
		// the other's: each a view of its own.
		if lowering.censusNeverRestSignature(toSignatures[0]) {
			if lowering.censusDiscardedMarkerPredicate(fromSignatures[0], toSignatures[0]) {
				return nil
			}
			source := lowering.checker.GetReturnTypeOfSignature(fromSignatures[0])
			target := lowering.checker.GetReturnTypeOfSignature(toSignatures[0])
			if !lowering.checker.IsTypeAssignableTo(source, target) {
				return &widening{source: source, target: target}
			}
			return lowering.widened(source, target, visited)
		}
		if len(fromSignatures) == 1 &&
			len(toSignatures) == 1 &&
			len(fromSignatures[0].TypeParameters()) > 0 &&
			len(toSignatures[0].TypeParameters()) == 0 {
			fromSignatures = []*checker.Signature{
				instantiateSignatureInContextOf(lowering.checker, fromSignatures[0], toSignatures[0], nil, nil),
			}
		}
		// Each position the viewed signature hands an argument to (anyRelatedPositions), typed as a
		// call there sees it: a rest parameter gives its element at that position, never its whole
		// tuple or array, so a listener `(seen: Error) => void` seen as `(...args: [error: Error]) =>
		// void` is related Error to Error (#6xzs6gj).
		for position := range anyRelatedPositions(fromSignatures[0], toSignatures[0]) {
			takes, given := lowering.callableParameterTypeAt(
				fromSignatures[0],
				position,
			), lowering.callableParameterTypeAt(toSignatures[0], position)
			if !lowering.enumAssignable(given, takes) || !lowering.checker.IsTypeAssignableTo(given, takes) {
				// tsc relates a method's parameters both ways (method bivariance), so a method taking
				// a Dog can be seen as one taking any Animal, and handed a Cat.
				return &widening{source: takes, target: given, parameter: true}
			}
			if found := lowering.widened(given, takes, visited); found != nil {
				return found
			}
		}
		return lowering.widened(
			lowering.checker.GetReturnTypeOfSignature(fromSignatures[0]),
			lowering.checker.GetReturnTypeOfSignature(toSignatures[0]),
			visited,
		)
	}
	fromContainers := lowering.containers(from)
	for _, viewed := range lowering.containers(to) {
		for _, inside := range fromContainers {
			if !lowering.sameContainer(inside, viewed) {
				continue
			}
			if found := lowering.widenedElements(inside, viewed, visited); found != nil {
				return found
			}
			break
		}
	}
	if to.Flags()&checker.TypeFlagsObject != 0 && len(lowering.containers(to)) > 0 {
		// An array, tuple, map or set: its elements are its slots, and its own properties (length,
		// size) are numbers on both sides.
		return nil
	}
	return lowering.widenedProperties(from, to, nil, false, visited)
}

// widenedProperties walks to's properties against from's, but those skip names: each field a slot
// to can write unless it's readonly, and each method a function view of its own. A copy's fields
// (fresh) are slots of their own, which nothing else writes, so only what they hold is walked.
func (lowering *lowering) widenedProperties(
	from *checker.Type,
	to *checker.Type,
	skip map[string]bool,
	fresh bool,
	visited map[[2]*checker.Type]bool,
) *widening {
	for _, viewed := range lowering.checker.GetPropertiesOfType(to) {
		inside := lowering.checker.GetPropertyOfType(from, viewed.Name)
		if inside == nil || skip[viewed.Name] {
			continue
		}
		source, target := lowering.checker.GetTypeOfSymbol(inside), lowering.checker.GetTypeOfSymbol(viewed)
		if viewed.Flags&ast.SymbolFlagsMethod != 0 || inside.Flags&ast.SymbolFlagsMethod != 0 {
			// A method isn't a slot anything writes, but what it returns and takes are views: the
			// program's own methods. The library's (an Iterable's, an array's own) are left to the
			// rules for the library's types, and a generic one has no one type to compare.
			if len(viewed.Declarations) == 0 ||
				load.IsLibrary(ast.GetSourceFileOfNode(viewed.Declarations[0])) ||
				lowering.generic(target) ||
				lowering.generic(source) {
				continue
			}
			if found := lowering.widened(source, target, visited); found != nil {
				return found
			}
			continue
		}
		if !fresh && !lowering.checker.IsReadonlySymbol(viewed) && lowering.checker.IsReadonlySymbol(inside) {
			// tsc ignores readonly when it relates properties, so a readonly field, which may hold
			// something narrower than its type says (it's covariant), can be seen as a writable one,
			// and the wider type written into it. Only the identical type, which this walk never
			// enters, is the same field.
			return &widening{source: source, target: target, readonlyField: viewed.Name}
		}
		if !fresh &&
			!lowering.checker.IsReadonlySymbol(viewed) &&
			(!lowering.enumAssignable(target, source) || !lowering.checker.IsTypeAssignableTo(target, source)) {
			return &widening{source: source, target: target}
		}
		if found := lowering.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// generic reports whether a function type has a type parameter of its own.
func (lowering *lowering) generic(proven *checker.Type) bool {
	for _, signature := range lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindCall) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}

// widenedElements walks an array's, a tuple's, a map's or a set's type arguments in from against
// to's, each a slot to can write unless it's read only.
func (lowering *lowering) widenedElements(
	from *checker.Type,
	to *checker.Type,
	visited map[[2]*checker.Type]bool,
) *widening {
	mutable := !lowering.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") &&
		!(checker.IsTupleType(to) && to.TargetTupleType().IsReadonly())
	return lowering.widenedArguments(from, to, mutable, visited)
}

// widenedArguments is widenedElements with whether to's own slots can be written given: a fresh
// container's can't be written through anything else (freshValue).
func (lowering *lowering) widenedArguments(
	from *checker.Type,
	to *checker.Type,
	mutable bool,
	visited map[[2]*checker.Type]bool,
) *widening {
	fromArguments, toArguments := lowering.checker.GetTypeArguments(from), lowering.checker.GetTypeArguments(to)
	if checker.IsTupleType(from) && !checker.IsTupleType(to) && len(toArguments) == 1 {
		// A tuple seen as an array: every element of it is seen as the array's element.
		widenedTo := make([]*checker.Type, len(fromArguments))
		for index := range widenedTo {
			widenedTo[index] = toArguments[0]
		}
		toArguments = widenedTo
	}
	for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
		source, target := fromArguments[index], toArguments[index]
		if mutable &&
			(!lowering.enumAssignable(target, source) || !lowering.checker.IsTypeAssignableTo(target, source)) {
			return &widening{source: source, target: target}
		}
		if found := lowering.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// containers is a type's arrays, tuples, maps and sets: itself, or its members that are, when it's an
// intersection.
func (lowering *lowering) containers(proven *checker.Type) []*checker.Type {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsIntersection != 0 {
		members = proven.Types()
	}
	var found []*checker.Type
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject != 0 && member.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
			(lowering.isArrayType(member) ||
				checker.IsTupleType(member) ||
				lowering.isLibraryType(member, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
			found = append(found, member)
		}
	}
	return found
}

// sameContainer reports whether two containers are the same kind, whose type arguments pair up: two
// arrays or tuples (readonly or not), two maps, or two sets.
func (lowering *lowering) sameContainer(from *checker.Type, to *checker.Type) bool {
	switch {
	case lowering.isArrayType(to) || checker.IsTupleType(to):
		return lowering.isArrayType(from) || checker.IsTupleType(from)
	case lowering.isLibraryType(to, "Map", "ReadonlyMap"):
		return lowering.isLibraryType(from, "Map", "ReadonlyMap")
	default:
		return lowering.isLibraryType(from, "Set", "ReadonlySet")
	}
}

// canWrite reports whether something can be written through a value of a type, at any depth: an
// array, tuple, map or set that isn't read only, or a property that isn't readonly, or one inside a
// readonly one.
func (lowering *lowering) canWrite(proven *checker.Type, visited map[*checker.Type]bool) bool {
	proven = lowering.withoutUndefined(proven)
	if proven == nil ||
		visited[proven] ||
		!lowering.structured(proven) ||
		len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 {
		return false
	}
	visited[proven] = true
	if lowering.recordElement(proven) != nil {
		return true
	}
	containers := lowering.containers(proven)
	for _, container := range containers {
		if !lowering.isLibraryType(container, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") &&
			!(checker.IsTupleType(container) && container.TargetTupleType().IsReadonly()) {
			return true
		}
		for _, element := range lowering.checker.GetTypeArguments(container) {
			if lowering.canWrite(element, visited) {
				return true
			}
		}
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 && len(containers) > 0 {
		return false
	}
	for _, property := range lowering.checker.GetPropertiesOfType(proven) {
		if property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if !lowering.checker.IsReadonlySymbol(property) ||
			lowering.canWrite(lowering.checker.GetTypeOfSymbol(property), visited) {
			return true
		}
	}
	return false
}

// narrowedFrom reports whether from is the type parameter to itself, or it narrowed: an intersection
// with it among its members (T & {}, what a check against undefined leaves).
func (lowering *lowering) narrowedFrom(from *checker.Type, to *checker.Type) bool {
	if from == to {
		return true
	}
	if from.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range from.Types() {
			if member == to {
				return true
			}
		}
	}
	return false
}

// structured reports whether a type has parts to walk: an object or an intersection.
func (lowering *lowering) structured(proven *checker.Type) bool {
	return proven.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0
}

// withoutUndefined is a type with undefined taken out of it: the one type left, or the union itself
// when several are, whose members widened walks one by one (definedMembers).
func (lowering *lowering) withoutUndefined(proven *checker.Type) *checker.Type {
	if members := lowering.definedMembers(proven); len(members) == 1 {
		return members[0]
	}
	return proven
}

// definedMembers is a union's members but undefined, or the type itself when it isn't a union.
func (lowering *lowering) definedMembers(proven *checker.Type) []*checker.Type {
	if proven == nil {
		return nil
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return []*checker.Type{proven}
	}
	var members []*checker.Type
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined == 0 {
			members = append(members, member)
		}
	}
	return members
}

// isClassInstance is the instance side of a class, generic or not.
func isClassInstance(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if proven.ObjectFlags()&checker.ObjectFlagsClass != 0 {
		return true
	}
	return proven.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		proven.Target() != nil &&
		proven.Target().ObjectFlags()&checker.ObjectFlagsClass != 0
}

// viewSite reports whether an expression is a value going into a typed slot: an initializer (a
// variable's, a parameter's default, a class field's), the right of an assignment, an argument, a
// returned value, a literal's property (shorthand too) or element, or an arrow's expression body.
// Parentheses around it are the same site.
func viewSite(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		// A destructuring declaration without a type keeps nothing of the value: its names take the
		// types of its parts, and the pattern's own type (its fields any) is no view of it.
		declaration := parent.AsVariableDeclaration()
		if ast.IsBindingPattern(declaration.Name()) && declaration.Type == nil {
			return false
		}
		return declaration.Initializer == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression != node
	case ast.KindReturnStatement, ast.KindArrayLiteralExpression:
		return true
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindArrowFunction:
		return parent.AsArrowFunction().Body == node
	}
	return false
}

// refuseWidening refuses a value seen through a type that can write what it can't hold.
func (lowering *lowering) refuseWidening(node *ast.Node) error {
	if lowering.nodeFSFileReadOnlyArgument(node) || lowering.regexCallbackIntrinsicView(node) {
		return nil
	}
	var own, contextual *checker.Type
	var found *widening
	switch {
	case node.Kind == ast.KindAsExpression:
		as := node.AsAsExpression()
		if as.Type.Kind == ast.KindTypeReference &&
			ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) &&
			as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
			return nil
		}
		source, target := lowering.checker.GetTypeAtLocation(as.Expression), lowering.checker.GetTypeAtLocation(node)
		if lowering.checker.IsTypeAssignableTo(source, target) ||
			lowering.checker.IsTypeAssignableTo(lowering.checker.GetWidenedType(source), target) {
			if provenRelationError := lowering.provenRelation(node, as.Expression, target); provenRelationError != nil {
				return provenRelationError
			}
		} else {
			// Preserve the existing checks on downcasts; their lowering belongs to cast.go.
			own, contextual = source, target
			if found = lowering.freshOrWidened(as.Expression, own, contextual); found != nil {
				break
			}
		}
		own, contextual, found = lowering.castSlotWidened(node)
	case node.Kind == ast.KindSatisfiesExpression:
		satisfies := node.AsSatisfiesExpression()
		if provenRelationError := lowering.provenRelation(
			node,
			satisfies.Expression,
			lowering.checker.GetTypeAtLocation(satisfies.Type),
		); provenRelationError != nil {
			return provenRelationError
		}
		own, contextual, found = lowering.castSlotWidened(node)
	case node.Kind == ast.KindShorthandPropertyAssignment:
		// { pets } is { pets: pets }: the variable seen as the literal's property, judged exactly as the
		// named form's initializer is, against the slot its name has (slotType). Reading the property off
		// the literal's slot instead missed it whenever that slot was a union holding undefined, such as a
		// return type `{ p: W } | undefined`: undefined has no p, so the union had none, and a readonly R
		// went out as a writable W unjudged (#20nf0sm). The name's slot is the checker's contextual type
		// for the property, or carriedSlot's, both of which look past undefined. Where the checker names no
		// slot for the name but the literal has one, the property of each member the literal can sit in is
		// judged instead, so a slot the checker can't name never lets the value through unjudged.
		name := node.Name()
		literal := lowering.slotType(node.Parent)
		var slots []*checker.Type
		if slot := lowering.slotType(name); slot != nil {
			slots = append(slots, slot)
		} else if literal != nil {
			for _, member := range lowering.literalSlotMembers(node.Parent, literal) {
				if property := lowering.checker.GetPropertyOfType(member, name.Text()); property != nil {
					slots = append(slots, lowering.checker.GetTypeOfSymbol(property))
				}
			}
		}
		own = lowering.checker.GetTypeAtLocation(name)
		for _, slot := range slots {
			contextual = slot
			if found = lowering.freshOrWidened(name, own, contextual); found != nil {
				break
			}
		}
		if found == nil && literal != nil && len(slots) > 0 {
			contextual, found = lowering.memberLiteralSlotWidened(name, own, literal, name.Text(), contextual)
		}
	case node.Kind == ast.KindSpreadAssignment:
		// { ...kennel } copies kennel's fields, not what they hold: each field not written again after
		// it is kennel's value seen as the literal's field. It's judged against each member of the
		// literal's slot the literal can sit in, since a union's own properties are only those every
		// member has, and `{ p: W } | undefined` has none (#20nf0sm, Lower's neighbor).
		literal := lowering.slotType(node.Parent)
		if literal == nil {
			return nil
		}
		skip := map[string]bool{}
		after := false
		for _, property := range node.Parent.AsObjectLiteralExpression().Properties.Nodes {
			if property == node {
				after = true
			} else if after && property.Name() != nil {
				skip[property.Name().Text()] = true
			}
		}
		own = lowering.checker.GetTypeAtLocation(node.AsSpreadAssignment().Expression)
		for _, member := range lowering.literalSlotMembers(node.Parent, literal) {
			contextual = member
			if lowering.declaredKeys {
				// A copy writes each of the spread's entries under its key, so a dictionary's can land
				// under the literal's declared properties; its own declared properties are only copied.
				found = lowering.declaredKeyWidened(own, member, skip, true)
			}
			if found == nil {
				found = lowering.widenedProperties(own, member, skip, true, map[[2]*checker.Type]bool{})
			}
			if found != nil {
				break
			}
		}
	default:
		if node.Kind == ast.KindParenthesizedExpression || !lowering.isExpression(node) {
			return nil
		}
		if lowering.genericFunction(node) {
			// same<Item> written where a (value: number) => number goes is instantiated to exactly
			// that type, Item = number, so there's no wider view of it to judge.
			return nil
		}
		if pattern := destructuringTarget(node); pattern != nil {
			// [a, b] = tuple and { a, b } = value keep nothing of the pattern: each element or field
			// is read out and stored into its target, so each is the view, seen as its target's type,
			// and the pattern isn't one.
			if pattern.Kind == ast.KindObjectLiteralExpression {
				return lowering.refuseFieldWidening(node, pattern)
			}
			return lowering.refuseElementWidening(node, pattern)
		}
		if viewSite(node) {
			contextual = lowering.slotType(node)
		}
		if contextual == nil && !lowering.conditionalBranchInSlot(node) {
			// With no type written for it, a value can still be taken into a wider one tsc made: the
			// union of a conditional's branches reduced to the wider (flag ? dogs : animals is an
			// Animal[]), a literal's elements likewise, a function's returns likewise.
			contextual = lowering.impliedTarget(node)
		}
		if contextual == nil {
			return nil
		}
		own = lowering.checker.GetTypeAtLocation(node)
		found = lowering.freshOrWidened(node, own, contextual)
		if found == nil &&
			node.Parent != nil &&
			node.Parent.Kind == ast.KindPropertyAssignment &&
			node.Parent.AsPropertyAssignment().Initializer == node {
			if name := node.Parent.Name(); name != nil &&
				(ast.IsIdentifier(name) || ast.IsStringLiteral(name) || ast.IsNumericLiteral(name)) {
				if literal := lowering.slotType(node.Parent.Parent); literal != nil {
					contextual, found = lowering.memberLiteralSlotWidened(node, own, literal, name.Text(), contextual)
				}
			}
		}
	}
	if found == nil {
		return nil
	}
	return lowering.wideningRefusal(node, own, contextual, found)
}

// literalSlotMembers is each member of an object literal's slot the literal is judged against: the slot
// itself when it isn't a union, and otherwise every member that isn't a primitive (undefined and null
// among them) and that the literal's own type is assignable to, since a member it isn't assignable to
// never holds it. That filter only narrows which members are judged, and it's no relation of its own:
// tsc lets a literal into a discriminated union it fits as a whole but fits no one member of, by
// splitting its discriminant (`{ ...h, k }` with k: 'x' | 'y', into `{ k: 'x'; p: W } | { k: 'y'; p: W }`),
// so when no member passes the filter, every non-primitive member is judged, and the list is never empty
// while the slot has an object member (#20nf0sm, Compiler's road on d2a29ebf).
func (lowering *lowering) literalSlotMembers(literal *ast.Node, slot *checker.Type) []*checker.Type {
	if slot.Flags()&checker.TypeFlagsUnion == 0 {
		return []*checker.Type{slot}
	}
	own := lowering.checker.GetTypeAtLocation(literal)
	var members, objects []*checker.Type
	for _, member := range slot.Types() {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		objects = append(objects, member)
		if lowering.checker.IsTypeAssignableTo(own, member) {
			members = append(members, member)
		}
	}
	if len(members) == 0 {
		return objects
	}
	return members
}

// conditionalBranchInSlot is whether node is a branch of a conditional that goes straight into a typed
// slot. The conditional's own type, its branches' union reduced to the wider, is then held by nothing:
// the value goes into the slot, and refuseWidening judges each branch against that slot at the
// conditional (freshOrWidened's conditional arm). Judging a branch against the reduced union as well
// refused tsc's `this._alternatives = spec ? parseRange(spec) : emptyArray` (semver.ts:211, #2vjqseg),
// a never[] seen as Comparator[][] that no reference ever holds at that type, where the slot is
// readonly. A conditional with no slot keeps its implied target: in `const all = flag ? dogs :
// animals` the variable holds the union, and a push through it writes into either branch.
func (lowering *lowering) conditionalBranchInSlot(node *ast.Node) bool {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil || parent.Kind != ast.KindConditionalExpression {
		return false
	}
	conditional := parent.AsConditionalExpression()
	if conditional.WhenTrue != child && conditional.WhenFalse != child {
		return false
	}
	return viewSite(parent) && lowering.slotType(parent) != nil
}

// castSlotWidened judges an as or satisfies at a view site as the value it is: the cast's own type
// against the slot it goes into, as the default branch judges any other value there. The cast's relation
// to its operand is judged apart from it, and passing that says nothing of the slot, so without this
// inner satisfies string[] went into a (string | number)[] and wrote a number past the strings (#b6m00y7).
// The operand under the casts decides freshness, so a literal made right there stays fresh.
func (lowering *lowering) castSlotWidened(node *ast.Node) (*checker.Type, *checker.Type, *widening) {
	if !viewSite(node) {
		return nil, nil, nil
	}
	contextual := lowering.slotType(node)
	if contextual == nil {
		return nil, nil, nil
	}
	operand := ast.SkipParentheses(node)
	for operand.Kind == ast.KindAsExpression || operand.Kind == ast.KindSatisfiesExpression {
		operand = ast.SkipParentheses(operand.Expression())
	}
	own := lowering.checker.GetTypeAtLocation(node)
	return own, contextual, lowering.freshOrWidened(operand, own, contextual)
}

// castSlotRefusal refuses a cast whose own type goes into a slot that can write what it can't hold, the
// judgment castSlotWidened makes, for a checked cast, whose relation to its operand stringWideningRefusal
// leaves to cast.go.
func (lowering *lowering) castSlotRefusal(node *ast.Node) error {
	if lowering.nodeFSFileReadOnlyArgument(node) || lowering.regexCallbackIntrinsicView(node) {
		return nil
	}
	own, contextual, found := lowering.castSlotWidened(node)
	if found == nil {
		return nil
	}
	return lowering.wideningRefusal(node, own, contextual, found)
}

// An object literal written where a union goes is built as one of its members, so a property is
// judged against each member's own slot, not the union's property. For { kind: Kind.Text } |
// { kind: Kind.Count } that property is Kind.Text | Kind.Count, the whole open enum, which takes any
// number; the member it's built as promises one literal. Each member-literal slot the value can be
// assigned to is judged as if it were the only one; the union's own property stays the answer when
// none of them refuses.
func (lowering *lowering) memberLiteralSlotWidened(
	node *ast.Node,
	own, literal *checker.Type,
	name string,
	contextual *checker.Type,
) (*checker.Type, *widening) {
	if literal.Flags()&checker.TypeFlagsUnion == 0 {
		return contextual, nil
	}
	for _, member := range literal.Types() {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		property := lowering.checker.GetPropertyOfType(member, name)
		if property == nil {
			continue
		}
		slot := lowering.checker.GetNonNullableType(lowering.checker.GetTypeOfSymbol(property))
		if slot.Flags()&checker.TypeFlagsNumberLiteral == 0 && !lowering.numericEnum(lowering.enumIdentity(slot)) ||
			lowering.openNumericEnumType(slot) {
			continue
		}
		if !lowering.checker.IsTypeAssignableTo(own, slot) {
			continue
		}
		if found := lowering.freshOrWidened(node, own, slot); found != nil {
			return slot, found
		}
	}
	return contextual, nil
}

func (lowering *lowering) wideningRefusal(node *ast.Node, own, contextual *checker.Type, found *widening) error {
	if found.declaredKey != "" {
		return lowering.declaredKeyRefusal(node, own, contextual, found)
	}
	if found.enum {
		if enumObjectSymbol(found.target) != nil {
			return &Refused{
				Where: lowering.program.Where(node),
				What: "a structural object seen as " +
					lowering.checker.TypeToString(found.target) +
					"; the complete enum shape is unproven",
				Fix: "use the enum's runtime object or a typeof alias, or give the ordinary object an explicit interface",
			}
		}
		if lowering.flagEnum(lowering.enumIdentity(found.target)) &&
			!lowering.numericEnum(lowering.enumIdentity(found.target)) {
			return lowering.flagWriteRefusal(node, found.target)
		}
		if lowering.numericEnum(lowering.enumIdentity(found.target)) ||
			lowering.numericEnum(lowering.enumIdentity(found.source)) {
			return &Refused{
				Where: lowering.program.Where(node),
				What: "an unproven value assigned to a numeric literal or enum member slot " +
					lowering.checker.TypeToString(found.target),
				Fix: "compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)",
			}
		}
		return &Refused{
			Where: lowering.program.Where(node),
			What: "an arbitrary number or a value from another enum assigned to " +
				lowering.checker.TypeToString(found.target) +
				"; its members are a closed union",
			Fix: "use a declared member of this enum, or compare the number with its members and return the matching member (adamic/enum-members)",
		}
	}
	what := "a value of type " +
		lowering.checker.TypeToString(own) +
		" seen as " +
		lowering.checker.TypeToString(contextual) +
		", which can write " +
		lowering.checker.TypeToString(found.target) +
		" where " +
		lowering.checker.TypeToString(found.source) +
		" is read"
	fix := "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
	if found.parameter {
		what = "a function taking " +
			lowering.checker.TypeToString(found.source) +
			" seen as one taking " +
			lowering.checker.TypeToString(found.target) +
			" (tsc relates a method's parameters both ways), so it can be handed what it can't take"
		fix = "write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)"
	}
	if found.readonlyField != "" {
		what = "a value of type " +
			lowering.checker.TypeToString(own) +
			" seen as " +
			lowering.checker.TypeToString(contextual) +
			", whose readonly field " +
			found.readonlyField +
			" becomes writable: a readonly field may hold something narrower than " +
			lowering.checker.TypeToString(found.source) +
			", which a write of " +
			lowering.checker.TypeToString(found.target) +
			" would replace"
		fix = "keep " +
			found.readonlyField +
			" readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
	}
	if found.target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		what = "a value of type " +
			lowering.checker.TypeToString(found.source) +
			" seen as " +
			lowering.checker.TypeToString(found.target) +
			", a type parameter whose constraint " +
			lowering.checker.TypeToString(lowering.checker.GetBaseConstraintOfType(found.target)) +
			" can be written, so it can write what " +
			lowering.checker.TypeToString(found.source) +
			" can't hold"
		fix = "take it as " +
			lowering.checker.TypeToString(found.source) +
			", or constrain " +
			lowering.checker.TypeToString(found.target) +
			" to something readonly, which can't write (adamic/invariant-mutable)"
	}
	return &Refused{Where: lowering.program.Where(node), What: what, Fix: fix}
}

// freshOrWidened is widened for a value at a site, but a fresh one: a container made right there and
// held by nothing else (freshValue) can't be written through another name, so its own slots are
// covariant, and only what's inside it, held elsewhere too, is walked as a view.
func (lowering *lowering) freshOrWidened(node *ast.Node, own *checker.Type, contextual *checker.Type) *widening {
	node = ast.SkipParentheses(node)
	own = lowering.contextualGenericType(node, own)
	// A numeric enum read narrowed to never has a non-returning IR check at this exact site.
	// It cannot write a value into the contextual slot, including a closed string-enum slot.
	if own.Flags()&checker.TypeFlagsNever != 0 && lowering.enumNeverIdentity(node, map[*ast.Node]bool{}) != nil {
		return nil
	}
	targetLiteral := (contextual.Flags()&checker.TypeFlagsNumberLiteral != 0 ||
		lowering.numericEnum(lowering.enumIdentity(contextual))) && !lowering.openNumericEnumType(contextual)
	if targetLiteral {
		if declared := lowering.enumStoredType(node); declared != nil {
			if lowering.openNumericEnumType(declared) &&
				!lowering.enumMemberOrigin(node, lowering.enumIdentity(declared), map[*ast.Node]bool{}) {
				return &widening{source: declared, target: contextual, enum: true}
			}
		}
	}
	if found := lowering.flagMemberWidened(
		node,
		contextual,
	); found != nil && !lowering.numericEnum(lowering.enumIdentity(found.target)) {
		return found
	}
	if target := lowering.flagTarget(contextual); target != nil && lowering.flagDomain(node, target) {
		return nil
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// Made as the type it's written into, held by nothing else: its own parts are sites.
		return nil
	case ast.KindConditionalExpression:
		// Either branch is the value: each judged as it is.
		conditional := node.AsConditionalExpression()
		for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
			if found := lowering.freshOrWidened(
				branch,
				lowering.checker.GetTypeAtLocation(branch),
				contextual,
			); found != nil {
				if found.viewed == nil {
					found.viewed = lowering.checker.GetTypeAtLocation(branch)
				}
				return found
			}
		}
		return nil
	}
	visited := map[[2]*checker.Type]bool{}
	if !lowering.freshValue(node) {
		return lowering.widened(own, contextual, visited)
	}
	own, contextual = lowering.withoutUndefined(own), lowering.withoutUndefined(contextual)
	for _, viewed := range lowering.containers(contextual) {
		for _, inside := range lowering.containers(own) {
			if lowering.sameContainer(inside, viewed) {
				return lowering.widenedArguments(inside, viewed, false, visited)
			}
		}
	}
	return lowering.widened(own, contextual, visited)
}

// freshValue reports whether an expression makes a container nothing else holds: a copy (slice,
// filter, map, concat, Array.from) or a new Map, Set or Array.
func (lowering *lowering) freshValue(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNewExpression:
		callee := node.AsNewExpression().Expression
		return lowering.isLibraryGlobal(callee, "Map") ||
			lowering.isLibraryGlobal(callee, "Set") ||
			lowering.isLibraryGlobal(callee, "Array")
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return false
		}
		access := callee.AsPropertyAccessExpression()
		method := access.Name().Text()
		if method == "from" {
			return lowering.isLibraryGlobal(access.Expression, "Array")
		}
		switch method {
		case "slice", "filter", "map", "concat":
			return lowering.isArrayType(lowering.checker.GetTypeAtLocation(access.Expression)) ||
				lowering.isLibraryType(lowering.checker.GetTypeAtLocation(access.Expression), "ReadonlyArray")
		}
	}
	return false
}

// isExpression reports whether a node is one a value can be: anything but a statement, a declaration,
// a name being declared or a type.
func (lowering *lowering) isExpression(node *ast.Node) bool {
	return ast.IsExpression(node) && !ast.IsPartOfTypeNode(node)
}

// impliedTarget is the type tsc gave what a value with no type written for it goes into, or nil: a
// conditional's (or ??'s, ||'s, &&'s) own type for its branches, an array literal's element type for
// its elements, and a function's inferred result for what it returns.
func (lowering *lowering) impliedTarget(node *ast.Node) *checker.Type {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindConditionalExpression:
		if context := lowering.callableContext(parent); context != nil {
			return context
		}
		if conditional := parent.AsConditionalExpression(); conditional.WhenTrue == child ||
			conditional.WhenFalse == child {
			return lowering.impliedExpressionTarget(parent)
		}
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindQuestionQuestionToken, ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
			if context := lowering.callableContext(parent); context != nil {
				return context
			}
			if contextual := lowering.slotType(
				parent,
			); contextual != nil && lowering.isArrayType(lowering.withoutUndefined(contextual)) {
				return contextual
			}
			return lowering.impliedExpressionTarget(parent)
		}
	case ast.KindArrayLiteralExpression:
		return lowering.impliedExpressionTarget(parent)
	case ast.KindReturnStatement, ast.KindArrowFunction:
		if parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body != child {
			return nil
		}
		function := parent
		for function != nil && !ast.IsFunctionLike(function) {
			function = function.Parent
		}
		if function == nil {
			return nil
		}
		return lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(function))
	}
	return nil
}

// impliedExpressionTarget asks once per parent expression, even for nil tuple
// element targets. The checker and source tree stay fixed for this lowering;
// branches share the parent's type and array elements share its element type.
func (lowering *lowering) impliedExpressionTarget(parent *ast.Node) *checker.Type {
	if target, ok := lowering.slots.implied[parent]; ok {
		return target
	}
	query := lowering.impliedTargetQuery
	if query == nil {
		query = lowering.checker.GetTypeAtLocation
	}
	target := query(parent)
	if parent.Kind == ast.KindArrayLiteralExpression {
		literal := target
		target = nil
		if lowering.isArrayType(literal) {
			if arguments := lowering.checker.GetTypeArguments(literal); len(arguments) == 1 {
				target = arguments[0]
			}
		}
	}
	if lowering.slots.implied == nil {
		lowering.slots.implied = make(map[*ast.Node]*checker.Type)
	}
	lowering.slots.implied[parent] = target
	return target
}

// slotTargets are the slots a value is judged against that the checker's contextual type doesn't give.
type slotTargets struct {
	// explicit is the slot a caller set for a value the checker gives none to, which slotType returns
	// before the contextual type: a delayed Array write's element, whose parameter lib types any.
	explicit map[*ast.Node]*checker.Type
	// implied memoizes impliedExpressionTarget's answer per parent, within this checker.
	implied map[*ast.Node]*checker.Type
}

// slotType is the slot a value at node goes into: the one a caller set for it (slots.explicit), or the
// one tsc's contextual typing carries down from it to a part of the value (carriedSlot), or else the
// checker's contextual type, exactly. Every refusal rule reads its slot here and never from the checker
// (TestRefusalRulesReadTheirSlotThroughSlotType), so a caller that judges a value against a slot the
// checker doesn't give sets it and runs the whole registry over the value, and no rule is skipped
// because a second list beside the registry forgot it (#b6m00y7).
func (lowering *lowering) slotType(node *ast.Node) *checker.Type {
	if len(lowering.slots.explicit) > 0 {
		if slot, carried, _ := lowering.carriedSlot(node); carried {
			return slot
		}
	}
	return lowering.checker.GetContextualType(node, checker.ContextFlagsNone)
}

// keepSlot sets slot as node's own for the rest of the lowering. refuseInSlot only judges, but lowering
// has checks of its own at a value that read slotType (expression's weak or strong keep, a host object's
// view, a tuple where an array goes), so a value written into a slot the checker doesn't give is lowered
// with that slot kept, as what it's written into, the way a direct write's value is.
func (lowering *lowering) keepSlot(node *ast.Node, slot *checker.Type) {
	if lowering.slots.explicit == nil {
		lowering.slots.explicit = map[*ast.Node]*checker.Type{}
	}
	lowering.slots.explicit[node] = slot
}

// slotSet is whether node's slot comes from one a caller set (carriedSlot). A rule that reads a site's
// slot some other way than the contextual type, such as an argument's parameter from the call's resolved
// signature, reads slotType instead when this holds: a delayed write's call is lib's, whose parameter is
// any, and the slot the caller set is the one the value is written into.
func (lowering *lowering) slotSet(node *ast.Node) bool {
	if len(lowering.slots.explicit) == 0 {
		return false
	}
	_, carried, _ := lowering.carriedSlot(node)
	return carried
}

// carriedSlot is node's slot when it comes from one a caller set: node's own, or the one tsc's
// contextual typing would carry down to it from an enclosing value that has one, through parentheses, a
// non-null, an as const, a conditional's branches, a logical operand, the right of a comma, and a
// literal's property, shorthand, spread or element. carried is false when no explicit slot reaches node,
// and the checker's contextual type is its slot. readable is false when one reaches it but Adamic can't
// read node's part off it (a computed name, a property the slot doesn't declare, an element of a slot
// that isn't an array or a tuple, or a parent kind this doesn't model, such as an await's operand), and
// refuseInSlot fails closed rather than judging node by the checker's. Only the kinds tsc carries
// nothing through are let go (slotIndependent), so a kind nobody listed fails closed (Compiler's must on
// c06ed629).
func (lowering *lowering) carriedSlot(node *ast.Node) (slot *checker.Type, carried bool, readable bool) {
	if slot, set := lowering.slots.explicit[node]; set {
		return slot, true, true
	}
	parent := node.Parent
	if parent == nil {
		return nil, false, false
	}
	switch parent.Kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		return lowering.carriedSlot(parent)
	case ast.KindAsExpression:
		// x as const keeps the slot x goes into; any other assertion gives x the asserted type.
		if as := parent.AsAsExpression(); as.Type.Kind == ast.KindTypeReference &&
			ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) &&
			as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
			return lowering.carriedSlot(parent)
		}
	case ast.KindConditionalExpression:
		if conditional := parent.AsConditionalExpression(); conditional.WhenTrue == node ||
			conditional.WhenFalse == node {
			return lowering.carriedSlot(parent)
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return lowering.carriedSlot(parent)
		case ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			// Only the right is the value; tsc gives the left no slot.
			if binary.Right == node {
				return lowering.carriedSlot(parent)
			}
			if _, carried, _ := lowering.carriedSlot(parent); carried {
				return nil, true, true
			}
		}
	case ast.KindPropertyAssignment:
		if parent.AsPropertyAssignment().Initializer == node {
			literal, carried, readable := lowering.carriedSlot(parent.Parent)
			if !carried || !readable {
				return nil, carried, readable
			}
			property, readable := lowering.carriedProperty(literal, parent.Name())
			return property, true, readable
		}
	case ast.KindObjectLiteralExpression:
		// A member of the literal: a spread is the literal's whole slot, and a property, a shorthand, a
		// method or an accessor is its property's.
		if node.Kind == ast.KindSpreadAssignment {
			return lowering.carriedSlot(parent)
		}
		if literal, carried, readable := lowering.carriedSlot(parent); carried {
			if !readable {
				return nil, true, false
			}
			property, readable := lowering.carriedProperty(literal, node.Name())
			return property, true, readable
		}
	case ast.KindShorthandPropertyAssignment:
		// { p } is the variable written as the literal's property p.
		if parent.Name() == node {
			literal, carried, readable := lowering.carriedSlot(parent.Parent)
			if !carried || !readable {
				return nil, carried, readable
			}
			property, readable := lowering.carriedProperty(literal, node)
			return property, true, readable
		}
	case ast.KindSpreadAssignment:
		return lowering.carriedSlot(parent.Parent)
	case ast.KindSpreadElement:
		if parent.Parent.Kind == ast.KindArrayLiteralExpression {
			return lowering.carriedSlot(parent.Parent)
		}
	case ast.KindArrayLiteralExpression:
		array, carried, readable := lowering.carriedSlot(parent)
		if !carried || !readable || array == nil {
			return nil, carried, readable
		}
		array = lowering.checker.GetNonNullableType(array)
		elements := parent.AsArrayLiteralExpression().Elements.Nodes
		if checker.IsTupleType(array) {
			index := slices.Index(elements, node)
			arguments := lowering.checker.GetTypeArguments(array)
			if index < 0 || index >= len(arguments) || slices.ContainsFunc(
				elements[:index],
				func(element *ast.Node) bool { return element.Kind == ast.KindSpreadElement },
			) {
				return nil, true, false
			}
			return arguments[index], true, true
		}
		if lowering.isArrayType(array) {
			return lowering.elementTypeOfArray(array), true, true
		}
		return nil, true, false
	}
	if slotIndependent(parent, node) {
		return nil, false, false
	}
	if _, carried, _ := lowering.carriedSlot(parent); carried {
		return nil, true, false
	}
	return nil, false, false
}

// slotIndependent is whether tsc gives node no slot from its parent's, so that whatever slot node has is
// its own: a call's callee and arguments (which take its parameters), an access's object and key, an
// operator's operands but the ones carriedSlot carries, a conditional's test, an assertion's operand
// (the asserted type), an untagged template's spans, a computed key, a property's name, a spread
// argument in a call, and any type. It names the kinds that carry nothing, never the ones that carry,
// so a kind left off it fails closed in carriedSlot.
func slotIndependent(parent, node *ast.Node) bool {
	if ast.IsPartOfTypeNode(node) {
		return true
	}
	switch parent.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindBinaryExpression, ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindDeleteExpression,
		ast.KindConditionalExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression,
		ast.KindTemplateExpression, ast.KindTemplateSpan, ast.KindComputedPropertyName, ast.KindSpreadElement,
		ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		return true
	}
	return false
}

// carriedProperty is the slot tsc carries from a literal's slot to its property name: the property's
// declared type, or the slot's string index signature. readable is false for a computed name or one
// the slot doesn't declare, which Adamic can't read off it.
func (lowering *lowering) carriedProperty(literal *checker.Type, name *ast.Node) (*checker.Type, bool) {
	if literal == nil {
		// No slot reaches the literal, so none reaches its property.
		return nil, true
	}
	if name == nil || !ast.IsIdentifier(name) && !ast.IsStringLiteral(name) && !ast.IsNumericLiteral(name) {
		return nil, false
	}
	literal = lowering.checker.GetNonNullableType(literal)
	if property := lowering.checker.GetPropertyOfType(literal, name.Text()); property != nil {
		return lowering.checker.GetTypeOfSymbol(property), true
	}
	if index := lowering.checker.GetIndexTypeOfType(literal, lowering.checker.GetStringType()); index != nil {
		return index, true
	}
	return nil, false
}

// destructuringTarget is the array or object pattern a value is assigned into, [a, b] = value or
// { a, b } = value, or nil.
func destructuringTarget(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken &&
			binary.Right == node &&
			(binary.Left.Kind == ast.KindArrayLiteralExpression ||
				binary.Left.Kind == ast.KindObjectLiteralExpression) {
			return binary.Left
		}
	}
	return nil
}

// refuseWholeWidening judges a value assigned into a pattern as one value seen as the pattern's
// type, the way any other value at a view site is judged, for a value whose parts the pattern
// doesn't read out one by one. A fresh literal, ({ x: options[name] } = { x: 3 }), keeps
// freshOrWidened's exemption: it's made as the pattern's type and held by nothing else.
func (lowering *lowering) refuseWholeWidening(node *ast.Node, own *checker.Type) error {
	contextual := lowering.slotType(node)
	if contextual == nil {
		return nil
	}
	if found := lowering.freshOrWidened(node, own, contextual); found != nil {
		return lowering.wideningRefusal(node, own, contextual, found)
	}
	return nil
}

// refuseElementWidening refuses [a, b] = tuple when an element is seen through its name's type as
// a view that can write what the element can't hold: [animals] = [dogs] puts a Dog[] in an
// Animal[]. A value that isn't a tuple is judged whole, as anywhere.
func (lowering *lowering) refuseElementWidening(node *ast.Node, pattern *ast.Node) error {
	own := lowering.checker.GetTypeAtLocation(node)
	if !checker.IsTupleType(own) {
		return lowering.refuseWholeWidening(node, own)
	}
	elements := lowering.checker.GetTypeArguments(own)
	for index, target := range pattern.AsArrayLiteralExpression().Elements.Nodes {
		if index >= len(elements) || target.Kind == ast.KindOmittedExpression {
			continue
		}
		to := lowering.checker.GetTypeAtLocation(target)
		if found := lowering.widened(elements[index], to, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{
				Where: lowering.program.Where(target),
				What: "a value of type " +
					lowering.checker.TypeToString(elements[index]) +
					" seen as " +
					lowering.checker.TypeToString(to) +
					", which can write " +
					lowering.checker.TypeToString(found.target) +
					" where " +
					lowering.checker.TypeToString(found.source) +
					" is read",
				Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)",
			}
		}
	}
	return nil
}

// patternField is one field an object pattern reads out of a value: the value's field, the target
// it's stored into with that target's declared type, and the default stored there when the field
// is undefined, or nil.
type patternField struct {
	source       *checker.Type
	target       *ast.Node
	targetType   *checker.Type
	defaultValue *ast.Node
}

// refuseFieldWidening refuses { a, b: c } = value when a field is seen through its target's type
// as a view that can write what the field can't hold: ({ pets } = kennel()) puts kennel's Dog[] in
// an Animal[]. A default is judged against its target too. A value the pattern can't read field by
// field is judged whole, as anywhere.
func (lowering *lowering) refuseFieldWidening(node *ast.Node, pattern *ast.Node) error {
	own := lowering.checker.GetTypeAtLocation(node)
	fields, readable := lowering.patternFields(own, pattern)
	if !readable {
		return lowering.refuseWholeWidening(node, own)
	}
	for _, field := range fields {
		if found := lowering.widened(field.source, field.targetType, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{
				Where: lowering.program.Where(field.target),
				What: "a value of type " +
					lowering.checker.TypeToString(field.source) +
					" seen as " +
					lowering.checker.TypeToString(field.targetType) +
					", which can write " +
					lowering.checker.TypeToString(found.target) +
					" where " +
					lowering.checker.TypeToString(found.source) +
					" is read",
				Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)",
			}
		}
		if field.defaultValue == nil {
			continue
		}
		fallback := lowering.checker.GetTypeAtLocation(field.defaultValue)
		if found := lowering.widened(fallback, field.targetType, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{
				Where: lowering.program.Where(field.defaultValue),
				What: "a value of type " +
					lowering.checker.TypeToString(fallback) +
					" seen as " +
					lowering.checker.TypeToString(field.targetType) +
					", which can write " +
					lowering.checker.TypeToString(found.target) +
					" where " +
					lowering.checker.TypeToString(found.source) +
					" is read",
				Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)",
			}
		}
	}
	return nil
}

// patternFields lists the fields an object pattern reads out of a value of type own, each by a
// name own declares. readable is false where the pattern can't be read field by field: own is a
// union or an intersection, an element is a rest or has a computed key, a target is a nested
// pattern, an element access or carries its own default ({ a: b = 1 }), or own doesn't declare
// the name.
func (lowering *lowering) patternFields(own *checker.Type, pattern *ast.Node) (fields []patternField, readable bool) {
	if own.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		return nil, false
	}
	for _, property := range pattern.AsObjectLiteralExpression().Properties.Nodes {
		var field patternField
		switch property.Kind {
		case ast.KindShorthandPropertyAssignment:
			// { a } and { a = 1 }: the field goes into the variable a names.
			symbol := lowering.checker.GetShorthandAssignmentValueSymbol(property)
			if symbol == nil {
				return nil, false
			}
			field.target = property.Name()
			field.targetType = lowering.checker.GetTypeOfSymbol(symbol)
			field.defaultValue = property.AsShorthandPropertyAssignment().ObjectAssignmentInitializer
		case ast.KindPropertyAssignment:
			// { a: b } and { a: this.b }: an assignment target's type is its declared type.
			initializer := property.AsPropertyAssignment().Initializer
			if !ast.IsIdentifier(initializer) && !ast.IsPropertyAccessExpression(initializer) {
				return nil, false
			}
			field.target = initializer
			field.targetType = lowering.checker.GetTypeAtLocation(initializer)
		default:
			return nil, false
		}
		name := property.Name()
		if name == nil || !ast.IsIdentifier(name) && !ast.IsStringLiteral(name) && !ast.IsNumericLiteral(name) {
			return nil, false
		}
		symbol := lowering.checker.GetPropertyOfType(own, name.Text())
		if symbol == nil || field.targetType == nil {
			return nil, false
		}
		field.source = lowering.checker.GetTypeOfSymbol(symbol)
		fields = append(fields, field)
	}
	return fields, true
}

// genericFunction reports whether an expression is a function with type parameters of its own,
// which tsc instantiates to the type it's written into rather than viewing it through that type.
func (lowering *lowering) genericFunction(node *ast.Node) bool {
	for _, signature := range lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(node),
		checker.SignatureKindCall,
	) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}
