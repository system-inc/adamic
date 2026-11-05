package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An array, a map or a set stores its elements as they're held where it's made, and every read and write
// through a slot works them as the slot's element type holds them, with nothing between to convert. So a
// container made with its elements stored one way can't go where they're held another (#b8fwyfx): [0, 1]
// stores numbers, and written into "string" | (number | string)[], by an object literal's field, a
// declaration, an assignment, an argument or a return, it was pushed a boxed string and read back as
// boxes, which printed undefined and segfaulted natively where Node printed 0, 1, c. A copy
// (kinds.slice()) and a new Map did the same in a plain (number | string)[] or Map<string, number | string>
// slot, read only too. invariant-mutable lets a fresh container's own slots widen (freshOrWidened), since
// nothing else reads it at its own type; that says nothing of how its elements are laid out, which is
// judged here.
//
// The storage is decided where the value is made, and it's what every slot that holds the value reads,
// as #fsb6n82 found for a literal's tuple. A union slot is any member the value is assignable to, and
// narrowing can pick any of them, so every such container member must read the elements as they're
// stored, not one. An array literal is judged at the representation lowering built it with. Any other
// value is judged at its own type, the one it was stored as when it was made, a named one included:
// invariant-mutable lets a named number[] go where readonly (number | string)[] is held, since that slot
// can't write, but its reads still worked the numbers as boxes and segfaulted natively, and a
// Map<string, number> read through ReadonlyMap<string, number | string> printed undefined (#3vsp1ds).

// elementFitDepth bounds elementsSeenAs. A container reached no sooner is judged, so it fails closed.
const elementFitDepth = 16

// elementFitKey is what elementsSeenAs's answer at the top depends on, the memo's key: the value's
// type, the slot's, how a literal stored its elements, and whether it's a literal, whose elements are
// judged at their own sites. The walk reads nothing else, so one answer serves every site of a pair.
type elementFitKey struct {
	own, slot *checker.Type
	built     ir.Type
	shallow   bool
}

// elementMismatch is where a container's elements are stored otherwise than a slot reads them: the two
// element types, or which side Adamic can't name a representation for, or neither when the walk went
// deeper than elementFitDepth.
type elementMismatch struct {
	stored, read *checker.Type
	held, as     ir.Type
	unnamed      string
}

// elementsHeldOtherwise is the stop for the value made at node going into slot when a container in it
// stores its elements otherwise than a container slot reads them, or "". It's the one judgment, made where
// expression hands every lowered value to its slot (TestElementsHeldOtherwiseIsJudgedOnce).
func (l *lowering) elementsHeldOtherwise(node *ast.Node, slot *checker.Type) string {
	node = ast.SkipParentheses(node)
	var built ir.Type
	switch {
	case node.Kind == ast.KindArrayLiteralExpression:
		if len(node.AsArrayLiteralExpression().Elements.Nodes) == 0 || l.tupleType(node) != nil {
			// [] stores nothing to be read otherwise, and a tuple is stored as its own type's slots
			// (tupleLiteral), which the reads of it follow (#fsb6n82).
			return ""
		}
		// A literal's elements are stored as elementType builds them: the slot's, where that's an array
		// whose elements are boxed, or the literal's own. Each element is a value at its own slot, so
		// only the literal's own elements are judged here.
		element, err := l.elementType(node)
		if err != nil {
			// The literal lowered (expression judges a value only once it has), so elementType can't
			// restate how its elements were stored, and nothing shows they're read as they're held.
			return "a " + l.checker.TypeToString(l.checker.GetTypeAtLocation(node)) + " where " + l.checker.TypeToString(slot) + " goes, its elements stored in a way Adamic can't name here"
		}
		built = element
	case node.Kind == ast.KindObjectLiteralExpression:
		// An object literal's fields are values at their own slots, each judged there.
		return ""
	}
	if !l.slotSet(node) && patternInitializer(node) {
		return ""
	}
	own := l.checker.GetTypeAtLocation(node)
	key := elementFitKey{own: own, slot: slot, built: built, shallow: node.Kind == ast.KindArrayLiteralExpression}
	found, answered := l.elementFits[key]
	if !answered {
		found = l.elementsSeenAs(own, slot, built, key.shallow, 0, map[[2]*checker.Type]bool{})
		if l.elementFits == nil {
			l.elementFits = map[elementFitKey]*elementMismatch{}
		}
		l.elementFits[key] = found
	}
	if found == nil {
		return ""
	}
	where := "a " + l.checker.TypeToString(own) + " where " + l.checker.TypeToString(slot) + " goes"
	switch {
	case found.stored == nil:
		return where + ", its containers nested deeper than their elements are judged"
	case found.unnamed != "":
		return where + ", with no representation Adamic can name for the " + found.unnamed + " elements (" + l.checker.TypeToString(found.stored) + " stored, read as " + l.checker.TypeToString(found.read) + ")"
	}
	return where + ": its " + l.checker.TypeToString(found.stored) + " elements are stored as " + heldName(found.held) + " and read there as " + l.checker.TypeToString(found.read) + ", held as " + heldName(found.as) + " (build it as the slot's type where it's made, or copy its elements into one)"
}

// patternInitializer is whether node is the initializer of a variable declaration whose name is a
// binding pattern with no type of its own, through parentheses and as const. tsc gives such an
// initializer the pattern's implied type as its contextual type, every element any ([any, any] for
// const [key, value] = entry), but no slot holds the value at it: the declaration takes the
// initializer's own type, and each binding reads its element there. An annotated pattern's type is a
// slot its bindings read through, so it's still judged.
func patternInitializer(node *ast.Node) bool {
	for node.Parent != nil {
		parent := node.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			node = parent
			continue
		case ast.KindAsExpression:
			if as := parent.AsAsExpression(); as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
				node = parent
				continue
			}
		case ast.KindVariableDeclaration:
			declaration := parent.AsVariableDeclaration()
			return declaration.Initializer == node && declaration.Type == nil && ast.IsBindingPattern(declaration.Name())
		}
		return false
	}
	return false
}

// elementsSeenAs finds, in a value of type stored seen as type viewed, a container whose elements (an
// array's, a map's keys and values, a set's) are stored otherwise than a container member of viewed that
// takes it reads them: at the top, or inside an element, a field or a dictionary's value, all the way
// down. built, when it's set, is how the top array stores its elements, and shallow leaves what's inside
// them to their own sites.
func (l *lowering) elementsSeenAs(stored *checker.Type, viewed *checker.Type, built ir.Type, shallow bool, depth int, visited map[[2]*checker.Type]bool) *elementMismatch {
	if stored == nil || viewed == nil || stored == viewed && built == 0 || visited[[2]*checker.Type{stored, viewed}] {
		return nil
	}
	if depth > elementFitDepth {
		if l.mentionsHeld(stored, []ir.Type{ir.Array, ir.Map}, 0, map[*checker.Type]bool{}) {
			return &elementMismatch{}
		}
		return nil
	}
	visited[[2]*checker.Type{stored, viewed}] = true
	for _, storedMember := range l.heldMembers(stored) {
		for _, viewedMember := range l.heldMembers(viewed) {
			if !l.checker.IsTypeAssignableTo(storedMember, viewedMember) {
				continue
			}
			inside, seen := l.containers(storedMember), l.containers(viewedMember)
			if len(inside) == 0 && len(seen) == 0 {
				if found := l.fieldsSeenAs(storedMember, viewedMember, depth, visited); found != nil {
					return found
				}
				continue
			}
			for _, container := range inside {
				for _, slot := range seen {
					if !l.sameContainer(container, slot) || checker.IsTupleType(container) != checker.IsTupleType(slot) {
						// A tuple seen as an array has its own stop (tupleSeenAsArray).
						continue
					}
					storedArguments, viewedArguments := l.typeArguments(container), l.typeArguments(slot)
					for index := 0; index < len(storedArguments) && index < len(viewedArguments); index++ {
						literal := built != 0 && index == 0 && l.checker.IsArrayType(container)
						if storedArguments[index] == viewedArguments[index] && !literal {
							// One type is held one way, whatever it is.
							continue
						}
						if !literal && storedArguments[index].Flags()&(checker.TypeFlagsNever|checker.TypeFlagsAny) != 0 {
							// A container of never holds no element to read otherwise (tsc's emptyArray, a
							// never[] seen as readonly (readonly number[])[]), and an any element is judged
							// where any meets a typed slot (any_relations.go), not here.
							continue
						}
						held, heldKnown := l.kept(storedArguments[index])
						if literal {
							held, heldKnown = built, true
						}
						as, asKnown := l.kept(viewedArguments[index])
						mismatch := &elementMismatch{stored: storedArguments[index], read: viewedArguments[index], held: held, as: as}
						switch {
						case !heldKnown:
							// Fails closed: a representation that can't be named can't be shown to match.
							mismatch.unnamed = "stored"
							return mismatch
						case !asKnown:
							mismatch.unnamed = "slot's"
							return mismatch
						case !elementReadAs(held, as):
							return mismatch
						}
						if shallow {
							continue
						}
						if found := l.elementsSeenAs(storedArguments[index], viewedArguments[index], 0, false, depth+1, visited); found != nil {
							return found
						}
					}
				}
			}
		}
	}
	return nil
}

// fieldsSeenAs is elementsSeenAs through an object's fields and a dictionary's values, where a container
// held in one is read through the other's. A library type holds only what its type arguments say, and a
// function's slots are sameKeeping's to judge.
func (l *lowering) fieldsSeenAs(stored *checker.Type, viewed *checker.Type, depth int, visited map[[2]*checker.Type]bool) *elementMismatch {
	if !l.structured(stored) || !l.structured(viewed) || len(l.checker.GetSignaturesOfType(viewed, checker.SignatureKindCall)) > 0 {
		return nil
	}
	if l.librarySymbol(viewed.Symbol()) && viewed.ObjectFlags()&checker.ObjectFlagsAnonymous == 0 {
		return nil
	}
	for _, field := range l.checker.GetPropertiesOfType(viewed) {
		if field.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if inside := l.checker.GetPropertyOfType(stored, field.Name); inside != nil {
			if found := l.elementsSeenAs(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(field), 0, false, depth+1, visited); found != nil {
				return found
			}
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(viewed) {
		for _, own := range l.checker.GetIndexInfosOfType(stored) {
			if own.KeyType() == info.KeyType() {
				if found := l.elementsSeenAs(own.ValueType(), info.ValueType(), 0, false, depth+1, visited); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// heldMembers is a type's members that can hold a value, a union's but null and undefined, or the type.
func (l *lowering) heldMembers(proven *checker.Type) []*checker.Type {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return []*checker.Type{proven}
	}
	var members []*checker.Type
	for _, member := range proven.Types() {
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			members = append(members, member)
		}
	}
	return members
}

// elementReadAs reports whether an element stored as held reads correctly as as, with nothing to convert
// it (native's converted): the same, a number read as number | undefined, which packs undefined into a
// reserved NaN and every other number as itself, or a reference read as a union, whose box is the
// reference itself. A number or a boolean read as a union needs a box of its own, and a Weak is a handle.
func elementReadAs(held ir.Type, as ir.Type) bool {
	return held == as || held == ir.Number && as == ir.MaybeNumber || as == ir.Union && held.IsReference() && held != ir.Weak
}

// heldName names how an element is held, for a stop: a scalar by its type, a union as its box.
func heldName(held ir.Type) string {
	switch held {
	case ir.Union:
		return "a boxed union"
	case ir.Number, ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean:
		return typeName(held)
	}
	if held.IsReference() {
		return "a reference"
	}
	return typeName(held)
}
