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
func (lowering *lowering) elementsHeldOtherwise(node *ast.Node, slot *checker.Type) string {
	node = ast.SkipParentheses(node)
	var built ir.Type
	switch {
	case node.Kind == ast.KindArrayLiteralExpression:
		if len(node.AsArrayLiteralExpression().Elements.Nodes) == 0 || lowering.tupleType(node) != nil {
			// [] stores nothing to be read otherwise, and a tuple is stored as its own type's slots
			// (tupleLiteral), which the reads of it follow (#fsb6n82).
			return ""
		}
		// A literal's elements are stored as elementType builds them: the slot's, where that's an array
		// whose elements are boxed, or the literal's own. Each element is a value at its own slot, so
		// only the literal's own elements are judged here.
		element, elementTypeError := lowering.elementType(node)
		if elementTypeError != nil {
			// The literal lowered (expression judges a value only once it has), so elementType can't
			// restate how its elements were stored, and nothing shows they're read as they're held.
			return "a " +
				lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(node)) +
				" where " +
				lowering.checker.TypeToString(slot) +
				" goes, its elements stored in a way Adamic can't name here"
		}
		built = element
	case node.Kind == ast.KindObjectLiteralExpression:
		// An object literal's fields are values at their own slots, each judged there.
		return ""
	}
	if !lowering.slotSet(node) && patternInitializer(node) {
		return ""
	}
	own := lowering.checker.GetTypeAtLocation(node)
	key := elementFitKey{own: own, slot: slot, built: built, shallow: node.Kind == ast.KindArrayLiteralExpression}
	found, answered := lowering.elementFits[key]
	if !answered {
		found = lowering.elementsSeenAs(own, slot, built, key.shallow, 0, map[[2]*checker.Type]bool{})
		if lowering.elementFits == nil {
			lowering.elementFits = map[elementFitKey]*elementMismatch{}
		}
		lowering.elementFits[key] = found
	}
	if found == nil {
		return ""
	}
	where := "a " + lowering.checker.TypeToString(own) + " where " + lowering.checker.TypeToString(slot) + " goes"
	switch {
	case found.stored == nil:
		return where + ", its containers nested deeper than their elements are judged"
	case found.unnamed != "":
		return where +
			", with no representation Adamic can name for the " +
			found.unnamed +
			" elements (" +
			lowering.checker.TypeToString(found.stored) +
			" stored, read as " +
			lowering.checker.TypeToString(found.read) +
			")"
	}
	return where +
		": its " +
		lowering.checker.TypeToString(found.stored) +
		" elements are stored as " +
		heldName(found.held) +
		" and read there as " +
		lowering.checker.TypeToString(found.read) +
		", held as " +
		heldName(found.as) +
		" (build it as the slot's type where it's made, or copy its elements into one)"
}

// patternInitializer is whether node is what a destructuring pattern takes apart, through parentheses
// and as const: the initializer of a variable declaration whose name is a binding pattern with no type
// of its own, or the right side of an assignment to an array or object literal pattern. tsc gives such
// a value the pattern's type as its contextual type, every element any ([any, any] for
// const [key, value] = entry) or the targets' declared types ([number | string, number | string] for
// [first, second] = tuple), but no slot holds the value at it: each element is read where the value
// is held and written into its own binding or target, which judges it there (#fsb6n82's swap). An
// annotated declaration's type is a slot its bindings read through, so it's still judged.
func patternInitializer(node *ast.Node) bool {
	for node.Parent != nil {
		parent := node.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			node = parent
			continue
		case ast.KindAsExpression:
			if as := parent.AsAsExpression(); as.Type.Kind == ast.KindTypeReference &&
				ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) &&
				as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
				node = parent
				continue
			}
		case ast.KindVariableDeclaration:
			declaration := parent.AsVariableDeclaration()
			return declaration.Initializer == node &&
				declaration.Type == nil &&
				ast.IsBindingPattern(declaration.Name())
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			target := ast.SkipParentheses(binary.Left)
			return binary.OperatorToken.Kind == ast.KindEqualsToken &&
				binary.Right == node &&
				(target.Kind == ast.KindArrayLiteralExpression || target.Kind == ast.KindObjectLiteralExpression)
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
func (lowering *lowering) elementsSeenAs(
	stored *checker.Type,
	viewed *checker.Type,
	built ir.Type,
	shallow bool,
	depth int,
	visited map[[2]*checker.Type]bool,
) *elementMismatch {
	if stored == nil || viewed == nil || stored == viewed && built == 0 || visited[[2]*checker.Type{stored, viewed}] {
		return nil
	}
	if depth > elementFitDepth {
		if lowering.mentionsHeld(stored, []ir.Type{ir.Array, ir.Map}, 0, map[*checker.Type]bool{}) {
			return &elementMismatch{}
		}
		return nil
	}
	visited[[2]*checker.Type{stored, viewed}] = true
	for _, storedMember := range lowering.heldMembers(stored) {
		for _, viewedMember := range lowering.heldMembers(viewed) {
			if !lowering.checker.IsTypeAssignableTo(storedMember, viewedMember) {
				continue
			}
			inside, seen := lowering.containers(storedMember), lowering.containers(viewedMember)
			if len(inside) == 0 && len(seen) == 0 {
				if found := lowering.fieldsSeenAs(storedMember, viewedMember, depth, visited); found != nil {
					return found
				}
				continue
			}
			for _, container := range inside {
				for _, slot := range seen {
					if !lowering.sameContainer(container, slot) ||
						checker.IsTupleType(container) != checker.IsTupleType(slot) {
						// A tuple seen as an array has its own stop (tupleSeenAsArray).
						continue
					}
					storedArguments, viewedArguments := lowering.typeArguments(container), lowering.typeArguments(slot)
					for index := 0; index < len(storedArguments) && index < len(viewedArguments); index++ {
						literal := built != 0 && index == 0 && lowering.isArrayType(container)
						if storedArguments[index] == viewedArguments[index] && !literal {
							// One type is held one way, whatever it is.
							continue
						}
						if !literal &&
							storedArguments[index].Flags()&(checker.TypeFlagsNever|checker.TypeFlagsAny) != 0 {
							// A container of never holds no element to read otherwise (tsc's emptyArray, a
							// never[] seen as readonly (readonly number[])[]), and an any element is judged
							// where any meets a typed slot (any_relations.go), not here.
							continue
						}
						held, heldKnown := lowering.kept(storedArguments[index])
						if literal {
							held, heldKnown = built, true
						}
						as, asKnown := lowering.kept(viewedArguments[index])
						mismatch := &elementMismatch{
							stored: storedArguments[index],
							read:   viewedArguments[index],
							held:   held,
							as:     as,
						}
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
						if found := lowering.elementsSeenAs(
							storedArguments[index],
							viewedArguments[index],
							0,
							false,
							depth+1,
							visited,
						); found != nil {
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
func (lowering *lowering) fieldsSeenAs(
	stored *checker.Type,
	viewed *checker.Type,
	depth int,
	visited map[[2]*checker.Type]bool,
) *elementMismatch {
	if !lowering.structured(stored) ||
		!lowering.structured(viewed) ||
		len(lowering.checker.GetSignaturesOfType(viewed, checker.SignatureKindCall)) > 0 {
		return nil
	}
	if lowering.librarySymbol(viewed.Symbol()) && viewed.ObjectFlags()&checker.ObjectFlagsAnonymous == 0 {
		return nil
	}
	for _, field := range lowering.checker.GetPropertiesOfType(viewed) {
		if field.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if inside := lowering.checker.GetPropertyOfType(stored, field.Name); inside != nil {
			if found := lowering.elementsSeenAs(
				lowering.checker.GetTypeOfSymbol(inside),
				lowering.checker.GetTypeOfSymbol(field),
				0,
				false,
				depth+1,
				visited,
			); found != nil {
				return found
			}
		}
	}
	for _, info := range lowering.checker.GetIndexInfosOfType(viewed) {
		for _, own := range lowering.checker.GetIndexInfosOfType(stored) {
			if own.KeyType() == info.KeyType() {
				if found := lowering.elementsSeenAs(
					own.ValueType(),
					info.ValueType(),
					0,
					false,
					depth+1,
					visited,
				); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// heldMembers is a type's members that can hold a value, a union's but null and undefined, or the type.
func (lowering *lowering) heldMembers(proven *checker.Type) []*checker.Type {
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
	return held == as ||
		held == ir.Number && as == ir.MaybeNumber ||
		as == ir.Union && held.IsReference() && held != ir.Weak
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
