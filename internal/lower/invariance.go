package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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
}

// widened finds the first mutable slot at which a value of type from, seen as type to, could be
// written something it can't hold, or returns nil.
func (l *lowering) widened(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	from, to = l.withoutUndefined(from), l.withoutUndefined(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return nil
	}
	visited[[2]*checker.Type{from, to}] = true
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
	if !l.structured(from) || !l.structured(to) || isClassInstance(to) {
		// A class instance is nominal in 0.1, the same class with the same type arguments, which
		// leaves nothing inside it to widen. That's cohere's nominal-class rule, which stage 0 doesn't
		// check yet: Box<Dog> seen as Box<Animal> isn't refused here.
		return nil
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		// A function seen as another is handed the other's arguments, and its results are seen as
		// the other's: each a view of its own.
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			if found := l.widened(l.checker.GetTypeOfSymbol(toParameters[index]), l.checker.GetTypeOfSymbol(fromParameters[index]), visited); found != nil {
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
	for _, viewed := range l.checker.GetPropertiesOfType(to) {
		if viewed.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		inside := l.checker.GetPropertyOfType(from, viewed.Name)
		if inside == nil {
			continue
		}
		source, target := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)
		if !l.checker.IsReadonlySymbol(viewed) && !l.checker.IsTypeAssignableTo(target, source) {
			return &widening{source: source, target: target}
		}
		if found := l.widened(source, target, visited); found != nil {
			return found
		}
	}
	return nil
}

// widenedElements walks an array's, a tuple's, a map's or a set's type arguments in from against
// to's, each a slot to can write unless it's read only.
func (l *lowering) widenedElements(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) *widening {
	mutable := !l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(to) && to.TargetTupleType().IsReadonly())
	fromArguments, toArguments := l.checker.GetTypeArguments(from), l.checker.GetTypeArguments(to)
	for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
		source, target := fromArguments[index], toArguments[index]
		if mutable && !l.checker.IsTypeAssignableTo(target, source) {
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

// withoutUndefined is a type with undefined taken out of it, the one type left, or nil when a union of
// several is left: a view of one of several isn't walked.
func (l *lowering) withoutUndefined(proven *checker.Type) *checker.Type {
	if proven == nil || proven.Flags()&checker.TypeFlagsUnion == 0 {
		return proven
	}
	var only *checker.Type
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		if only != nil {
			return nil
		}
		only = member
	}
	return only
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

// viewSite reports whether an expression is a value going into a typed slot: an initializer, the
// right of an assignment, an argument, a returned value, a literal's property or element, or an
// arrow's expression body. Parentheses around it are the same site.
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
		return parent.AsVariableDeclaration().Initializer == node
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
	case ast.KindArrowFunction:
		return parent.AsArrowFunction().Body == node
	}
	return false
}

// refuseWidening refuses a value seen through a type that can write what it can't hold.
func (l *lowering) refuseWidening(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !viewSite(node) {
		return nil
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// Made as the type it's written into, held by nothing else: its own parts are sites.
		return nil
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual == nil {
		return nil
	}
	own := l.checker.GetTypeAtLocation(node)
	found := l.widened(own, contextual, map[[2]*checker.Type]bool{})
	if found == nil {
		return nil
	}
	what := "a value of type " + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ", which can write " + l.checker.TypeToString(found.target) + " where " + l.checker.TypeToString(found.source) + " is read"
	fix := "make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
	if found.target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		what = "a value of type " + l.checker.TypeToString(found.source) + " seen as " + l.checker.TypeToString(found.target) + ", a type parameter whose constraint " + l.checker.TypeToString(l.checker.GetBaseConstraintOfType(found.target)) + " can be written, so it can write what " + l.checker.TypeToString(found.source) + " can't hold"
		fix = "take it as " + l.checker.TypeToString(found.source) + ", or constrain " + l.checker.TypeToString(found.target) + " to something readonly, which can't write (adamic/invariant-mutable)"
	}
	return &Refused{Where: l.program.Where(node), What: what, Fix: fix}
}
