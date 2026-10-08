package lower

import (
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
}

// widened finds the first mutable slot at which a value of type from, seen as type to, could be
// written something it can't hold, or returns nil.
func (l *lowering) widened(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	from, to = l.withoutUndefined(from), l.withoutUndefined(to)
	if from == to || visited[[2]*checker.Type{from, to}] || from.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	visited[[2]*checker.Type{from, to}] = true
	if l.openNumericEnumType(to) {
		for _, member := range l.definedMembers(from) {
			if !l.enumAssignable(member, to) {
				return &widening{source: from, target: to, enum: true}
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsNumberLiteral != 0 && l.openNumericEnumType(from) {
		return &widening{source: from, target: to, enum: true}
	}
	if members := l.definedMembers(from); len(members) > 1 {
		// A union is any one of its members, each seen as the target.
		for _, member := range members {
			if found := l.widened(member, to, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if members := l.definedMembers(to); len(members) > 1 {
		// Seen as a union, a value is seen as each member it can be, and narrowing picks any of them
		// to write through.
		for _, member := range members {
			if !l.checker.IsTypeAssignableTo(from, member) {
				continue
			}
			if found := l.widened(from, member, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if l.narrowedFrom(from, to) {
			return nil
		}
		if constraint := l.checker.GetBaseConstraintOfType(to); constraint != nil && l.canWrite(constraint, map[*checker.Type]bool{}) {
			return &widening{source: from, target: to}
		}
		return nil
	}
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		// What it can be is what its constraint allows: seen as anything wider, it's that seen so.
		constraint := l.checker.GetBaseConstraintOfType(from)
		if constraint == nil {
			return nil
		}
		return l.widened(constraint, to, visited)
	}
	if !l.enumAssignable(from, to) {
		return &widening{source: from, target: to, enum: true}
	}
	if !l.structured(from) || !l.structured(to) {
		// A class instance target is walked as any object is: its fields are slots as a literal's
		// are, so Box<Dog> seen as Box<Animal>, or a plain object seen as a class, is judged by them.
		return nil
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		// A function seen as another is handed the other's arguments, and its results are seen as
		// the other's: each a view of its own.
		if l.censusNeverRestSignature(toSignatures[0]) {
			if l.censusDiscardedMarkerPredicate(fromSignatures[0], toSignatures[0]) {
				return nil
			}
			source := l.checker.GetReturnTypeOfSignature(fromSignatures[0])
			target := l.checker.GetReturnTypeOfSignature(toSignatures[0])
			if !l.checker.IsTypeAssignableTo(source, target) {
				return &widening{source: source, target: target}
			}
			return l.widened(source, target, visited)
		}
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			takes, given := l.censusCallableParameterType(fromParameters[index]), l.censusCallableParameterType(toParameters[index])
			if !l.enumAssignable(given, takes) || !l.checker.IsTypeAssignableTo(given, takes) {
				// tsc relates a method's parameters both ways (method bivariance), so a method taking
				// a Dog can be seen as one taking any Animal, and handed a Cat.
				return &widening{source: takes, target: given, parameter: true}
			}
			if found := l.widened(given, takes, visited); found != nil {
				return found
			}
		}
		return l.widened(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]), visited)
	}
	fromContainers := l.containers(from)
	for _, viewed := range l.containers(to) {
		for _, inside := range fromContainers {
			if !l.sameContainer(inside, viewed) {
				continue
			}
			if found := l.widenedElements(inside, viewed, visited); found != nil {
				return found
			}
			break
		}
	}
	if to.Flags()&checker.TypeFlagsObject != 0 && len(l.containers(to)) > 0 {
		// An array, tuple, map or set: its elements are its slots, and its own properties (length,
		// size) are numbers on both sides.
		return nil
	}
	return l.widenedProperties(from, to, nil, false, visited)
}

// widenedProperties walks to's properties against from's, but those skip names: each field a slot
// to can write unless it's readonly, and each method a function view of its own. A copy's fields
// (fresh) are slots of their own, which nothing else writes, so only what they hold is walked.
func (l *lowering) widenedProperties(from *checker.Type, to *checker.Type, skip map[string]bool, fresh bool, visited map[[2]*checker.Type]bool) *widening {
	for _, viewed := range l.checker.GetPropertiesOfType(to) {
		inside := l.checker.GetPropertyOfType(from, viewed.Name)
		if inside == nil || skip[viewed.Name] {
			continue
		}
		source, target := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)
		if viewed.Flags&ast.SymbolFlagsMethod != 0 || inside.Flags&ast.SymbolFlagsMethod != 0 {
			// A method isn't a slot anything writes, but what it returns and takes are views: the
			// program's own methods. The library's (an Iterable's, an array's own) are left to the
			// rules for the library's types, and a generic one has no one type to compare.
			if len(viewed.Declarations) == 0 || load.IsLibrary(ast.GetSourceFileOfNode(viewed.Declarations[0])) || l.generic(target) || l.generic(source) {
				continue
			}
			if found := l.widened(source, target, visited); found != nil {
				return found
			}
			continue
		}
		if !fresh && !l.checker.IsReadonlySymbol(viewed) && l.checker.IsReadonlySymbol(inside) {
			// tsc ignores readonly when it relates properties, so a readonly field, which may hold
			// something narrower than its type says (it's covariant), can be seen as a writable one,
			// and the wider type written into it. Only the identical type, which this walk never
			// enters, is the same field.
			return &widening{source: source, target: target, readonlyField: viewed.Name}
		}
		if !fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source)) {
			return &widening{source: source, target: target}
		}
		if found := l.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// generic reports whether a function type has a type parameter of its own.
func (l *lowering) generic(proven *checker.Type) bool {
	for _, signature := range l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}

// widenedElements walks an array's, a tuple's, a map's or a set's type arguments in from against
// to's, each a slot to can write unless it's read only.
func (l *lowering) widenedElements(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	mutable := !l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(to) && to.TargetTupleType().IsReadonly())
	return l.widenedArguments(from, to, mutable, visited)
}

// widenedArguments is widenedElements with whether to's own slots can be written given: a fresh
// container's can't be written through anything else (freshValue).
func (l *lowering) widenedArguments(from *checker.Type, to *checker.Type, mutable bool, visited map[[2]*checker.Type]bool) *widening {
	fromArguments, toArguments := l.checker.GetTypeArguments(from), l.checker.GetTypeArguments(to)
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
		if mutable && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source)) {
			return &widening{source: source, target: target}
		}
		if found := l.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// containers is a type's arrays, tuples, maps and sets: itself, or its members that are, when it's an
// intersection.
func (l *lowering) containers(proven *checker.Type) []*checker.Type {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsIntersection != 0 {
		members = proven.Types()
	}
	var found []*checker.Type
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject != 0 && member.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
			(l.checker.IsArrayType(member) || checker.IsTupleType(member) || l.isLibraryType(member, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
			found = append(found, member)
		}
	}
	return found
}

// sameContainer reports whether two containers are the same kind, whose type arguments pair up: two
// arrays or tuples (readonly or not), two maps, or two sets.
func (l *lowering) sameContainer(from *checker.Type, to *checker.Type) bool {
	switch {
	case l.checker.IsArrayType(to) || checker.IsTupleType(to):
		return l.checker.IsArrayType(from) || checker.IsTupleType(from)
	case l.isLibraryType(to, "Map", "ReadonlyMap"):
		return l.isLibraryType(from, "Map", "ReadonlyMap")
	default:
		return l.isLibraryType(from, "Set", "ReadonlySet")
	}
}

// canWrite reports whether something can be written through a value of a type, at any depth: an
// array, tuple, map or set that isn't read only, or a property that isn't readonly, or one inside a
// readonly one.
func (l *lowering) canWrite(proven *checker.Type, visited map[*checker.Type]bool) bool {
	proven = l.withoutUndefined(proven)
	if proven == nil || visited[proven] || !l.structured(proven) || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 {
		return false
	}
	visited[proven] = true
	containers := l.containers(proven)
	for _, container := range containers {
		if !l.isLibraryType(container, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(container) && container.TargetTupleType().IsReadonly()) {
			return true
		}
		for _, element := range l.checker.GetTypeArguments(container) {
			if l.canWrite(element, visited) {
				return true
			}
		}
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 && len(containers) > 0 {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if !l.checker.IsReadonlySymbol(property) || l.canWrite(l.checker.GetTypeOfSymbol(property), visited) {
			return true
		}
	}
	return false
}

// narrowedFrom reports whether from is the type parameter to itself, or it narrowed: an intersection
// with it among its members (T & {}, what a check against undefined leaves).
func (l *lowering) narrowedFrom(from *checker.Type, to *checker.Type) bool {
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
func (l *lowering) structured(proven *checker.Type) bool {
	return proven.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0
}

// withoutUndefined is a type with undefined taken out of it: the one type left, or the union itself
// when several are, whose members widened walks one by one (definedMembers).
func (l *lowering) withoutUndefined(proven *checker.Type) *checker.Type {
	if members := l.definedMembers(proven); len(members) == 1 {
		return members[0]
	}
	return proven
}

// definedMembers is a union's members but undefined, or the type itself when it isn't a union.
func (l *lowering) definedMembers(proven *checker.Type) []*checker.Type {
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
	return proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil && proven.Target().ObjectFlags()&checker.ObjectFlagsClass != 0
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
func (l *lowering) refuseWidening(node *ast.Node) error {
	if l.nodeFSFileReadOnlyArgument(node) || l.regexCallbackIntrinsicView(node) {
		return nil
	}
	var own, contextual *checker.Type
	var found *widening
	switch {
	case node.Kind == ast.KindAsExpression:
		as := node.AsAsExpression()
		if as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
			return nil
		}
		source, target := l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node)
		if l.checker.IsTypeAssignableTo(source, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target) {
			return l.provenRelation(node, as.Expression, target)
		}
		// Preserve the existing checks on downcasts; their lowering belongs to cast.go.
		own, contextual = source, target
		found = l.freshOrWidened(as.Expression, own, contextual)
	case node.Kind == ast.KindSatisfiesExpression:
		satisfies := node.AsSatisfiesExpression()
		return l.provenRelation(node, satisfies.Expression, l.checker.GetTypeAtLocation(satisfies.Type))
	case node.Kind == ast.KindShorthandPropertyAssignment:
		// { pets } is { pets: pets }: the variable seen as the literal's property.
		literal := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
		if literal == nil {
			return nil
		}
		property := l.checker.GetPropertyOfType(literal, node.Name().Text())
		if property == nil {
			return nil
		}
		own, contextual = l.checker.GetTypeAtLocation(node.Name()), l.checker.GetTypeOfSymbol(property)
		found = l.freshOrWidened(node.Name(), own, contextual)
	case node.Kind == ast.KindSpreadAssignment:
		// { ...kennel } copies kennel's fields, not what they hold: each field not written again after
		// it is kennel's value seen as the literal's field.
		literal := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
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
		own, contextual = l.checker.GetTypeAtLocation(node.AsSpreadAssignment().Expression), literal
		found = l.widenedProperties(own, literal, skip, true, map[[2]*checker.Type]bool{})
	default:
		if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
			return nil
		}
		if l.genericFunction(node) {
			// same<Item> written where a (value: number) => number goes is instantiated to exactly
			// that type, Item = number, so there's no wider view of it to judge.
			return nil
		}
		if pattern := destructuringTarget(node); pattern != nil {
			// [a, b] = tuple keeps nothing of the pattern: each element is read out and stored into
			// its name, so each element is the view, seen as its name's type, and the pattern isn't one.
			return l.refuseElementWidening(node, pattern)
		}
		if viewSite(node) {
			contextual = l.checker.GetContextualType(node, checker.ContextFlagsNone)
		}
		if contextual == nil {
			// With no type written for it, a value can still be taken into a wider one tsc made: the
			// union of a conditional's branches reduced to the wider (flag ? dogs : animals is an
			// Animal[]), a literal's elements likewise, a function's returns likewise.
			contextual = l.impliedTarget(node)
		}
		if contextual == nil {
			return nil
		}
		own = l.checker.GetTypeAtLocation(node)
		found = l.freshOrWidened(node, own, contextual)
	}
	if found == nil {
		return nil
	}
	return l.wideningRefusal(node, own, contextual, found)
}

func (l *lowering) wideningRefusal(node *ast.Node, own, contextual *checker.Type, found *widening) error {
	if found.enum {
		if enumObjectSymbol(found.target) != nil {
			return &Refused{Where: l.program.Where(node), What: "a structural object seen as " + l.checker.TypeToString(found.target) + "; the complete enum shape is unproven", Fix: "use the enum's runtime object or a typeof alias, or give the ordinary object an explicit interface"}
		}
		if l.flagEnum(l.enumIdentity(found.target)) && !l.numericEnum(l.enumIdentity(found.target)) {
			return l.flagWriteRefusal(node, found.target)
		}
		if l.numericEnum(l.enumIdentity(found.target)) || l.numericEnum(l.enumIdentity(found.source)) {
			return &Refused{Where: l.program.Where(node), What: "an unproven value assigned to a numeric literal or enum member slot " + l.checker.TypeToString(found.target), Fix: "compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"}
		}
		return &Refused{Where: l.program.Where(node), What: "an arbitrary number or a value from another enum assigned to " + l.checker.TypeToString(found.target) + "; its members are a closed union", Fix: "use a declared member of this enum, or compare the number with its members and return the matching member (adamic/enum-members)"}
	}
	what := "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read"
	fix := "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
	if found.parameter {
		what = "a function taking " + l.checker.TypeToString(found.source) + " seen as one taking " + l.checker.TypeToString(found.target) + " (tsc relates a method's parameters both ways), so it can be handed what it can't take"
		fix = "write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)"
	}
	if found.readonlyField != "" {
		what = "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", whose readonly field " + found.readonlyField + " becomes writable: a readonly field may hold something narrower than " + l.checker.TypeToString(found.source) + ", which a write of " + l.checker.TypeToString(found.target) + " would replace"
		fix = "keep " + found.readonlyField + " readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
	}
	if found.target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		what = "a value of type " + l.checker.TypeToString(found.source) + " seen as " + l.checker.TypeToString(found.target) + ", a type parameter whose constraint " + l.checker.TypeToString(l.checker.GetBaseConstraintOfType(found.target)) + " can be written, so it can write what " + l.checker.TypeToString(found.source) + " can't hold"
		fix = "take it as " + l.checker.TypeToString(found.source) + ", or constrain " + l.checker.TypeToString(found.target) + " to something readonly, which can't write (adamic/invariant-mutable)"
	}
	return &Refused{Where: l.program.Where(node), What: what, Fix: fix}
}

// freshOrWidened is widened for a value at a site, but a fresh one: a container made right there and
// held by nothing else (freshValue) can't be written through another name, so its own slots are
// covariant, and only what's inside it, held elsewhere too, is walked as a view.
func (l *lowering) freshOrWidened(node *ast.Node, own *checker.Type, contextual *checker.Type) *widening {
	node = ast.SkipParentheses(node)
	// A numeric enum read narrowed to never has a non-returning IR check at this exact site.
	// It cannot write a value into the contextual slot, including a closed string-enum slot.
	if own.Flags()&checker.TypeFlagsNever != 0 && l.enumNeverIdentity(node, map[*ast.Node]bool{}) != nil {
		return nil
	}
	targetLiteral := (contextual.Flags()&checker.TypeFlagsNumberLiteral != 0 || l.numericEnum(l.enumIdentity(contextual))) && !l.openNumericEnumType(contextual)
	if targetLiteral {
		if declared := l.enumStoredType(node); declared != nil {
			if l.openNumericEnumType(declared) && !l.enumMemberOrigin(node, l.enumIdentity(declared), map[*ast.Node]bool{}) {
				return &widening{source: declared, target: contextual, enum: true}
			}
		}
	}
	if found := l.flagMemberWidened(node, contextual); found != nil && !l.numericEnum(l.enumIdentity(found.target)) {
		return found
	}
	if target := l.flagTarget(contextual); target != nil && l.flagDomain(node, target) {
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
			if found := l.freshOrWidened(branch, l.checker.GetTypeAtLocation(branch), contextual); found != nil {
				return found
			}
		}
		return nil
	}
	visited := map[[2]*checker.Type]bool{}
	if !l.freshValue(node) {
		return l.widened(own, contextual, visited)
	}
	own, contextual = l.withoutUndefined(own), l.withoutUndefined(contextual)
	for _, viewed := range l.containers(contextual) {
		for _, inside := range l.containers(own) {
			if l.sameContainer(inside, viewed) {
				return l.widenedArguments(inside, viewed, false, visited)
			}
		}
	}
	return l.widened(own, contextual, visited)
}

// freshValue reports whether an expression makes a container nothing else holds: a copy (slice,
// filter, map, concat, Array.from) or a new Map, Set or Array.
func (l *lowering) freshValue(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNewExpression:
		callee := node.AsNewExpression().Expression
		return l.isLibraryGlobal(callee, "Map") || l.isLibraryGlobal(callee, "Set") || l.isLibraryGlobal(callee, "Array")
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return false
		}
		access := callee.AsPropertyAccessExpression()
		method := access.Name().Text()
		if method == "from" {
			return l.isLibraryGlobal(access.Expression, "Array")
		}
		switch method {
		case "slice", "filter", "map", "concat":
			return l.checker.IsArrayType(l.checker.GetTypeAtLocation(access.Expression)) || l.isLibraryType(l.checker.GetTypeAtLocation(access.Expression), "ReadonlyArray")
		}
	}
	return false
}

// isExpression reports whether a node is one a value can be: anything but a statement, a declaration,
// a name being declared or a type.
func (l *lowering) isExpression(node *ast.Node) bool {
	return ast.IsExpression(node) && !ast.IsPartOfTypeNode(node)
}

// impliedTarget is the type tsc gave what a value with no type written for it goes into, or nil: a
// conditional's (or ??'s, ||'s, &&'s) own type for its branches, an array literal's element type for
// its elements, and a function's inferred result for what it returns.
func (l *lowering) impliedTarget(node *ast.Node) *checker.Type {
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
		if conditional := parent.AsConditionalExpression(); conditional.WhenTrue == child || conditional.WhenFalse == child {
			return l.checker.GetTypeAtLocation(parent)
		}
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindQuestionQuestionToken, ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
			if contextual := l.checker.GetContextualType(parent, checker.ContextFlagsNone); contextual != nil && l.checker.IsArrayType(l.withoutUndefined(contextual)) {
				return contextual
			}
			return l.checker.GetTypeAtLocation(parent)
		}
	case ast.KindArrayLiteralExpression:
		literal := l.checker.GetTypeAtLocation(parent)
		if l.checker.IsArrayType(literal) {
			if arguments := l.checker.GetTypeArguments(literal); len(arguments) == 1 {
				return arguments[0]
			}
		}
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
		return l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(function))
	}
	return nil
}

// destructuringTarget is the array pattern a value is assigned into, [a, b] = value, or nil.
func destructuringTarget(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node && binary.Left.Kind == ast.KindArrayLiteralExpression {
			return binary.Left
		}
	}
	return nil
}

// refuseElementWidening refuses [a, b] = tuple when an element is seen through its name's type as
// a view that can write what the element can't hold: [animals] = [dogs] puts a Dog[] in an
// Animal[]. A value that isn't a tuple is judged whole, as anywhere.
func (l *lowering) refuseElementWidening(node *ast.Node, pattern *ast.Node) error {
	own := l.checker.GetTypeAtLocation(node)
	if !checker.IsTupleType(own) {
		contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
		if contextual == nil {
			return nil
		}
		if found := l.widened(own, contextual, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{Where: l.program.Where(node), What: "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read", Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"}
		}
		return nil
	}
	elements := l.checker.GetTypeArguments(own)
	for index, target := range pattern.AsArrayLiteralExpression().Elements.Nodes {
		if index >= len(elements) || target.Kind == ast.KindOmittedExpression {
			continue
		}
		to := l.checker.GetTypeAtLocation(target)
		if found := l.widened(elements[index], to, map[[2]*checker.Type]bool{}); found != nil {
			return &Refused{Where: l.program.Where(target), What: "a value of type " + l.checker.TypeToString(elements[index]) + " seen as " + l.checker.TypeToString(to) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read", Fix: "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"}
		}
	}
	return nil
}

// genericFunction reports whether an expression is a function with type parameters of its own,
// which tsc instantiates to the type it's written into rather than viewing it through that type.
func (l *lowering) genericFunction(node *ast.Node) bool {
	for _, signature := range l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}
