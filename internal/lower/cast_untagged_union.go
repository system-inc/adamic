package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// untaggedUnionArm is whether a cast from a union no literal tag tells apart takes the checked view of its
// target (#axqgdck, rulings.md:18: an untagged downcast makes a checked view with no tag test). Every member
// and the target are plain objects held as ir.Object: no class instance, callable, array, tuple, dictionary
// or phantom extension, so nothing reaches the view that a field read can't check. The target is one object,
// a member or narrower than one (castAssignable to it), so the cast is a downcast; a sideways target and a
// union target keep their stops. A union one tag does tell apart is unionMemberProof's to prove or refuse
// (a payload refinement stays refused), never this arm's.
func (l *lowering) untaggedUnionArm(members, targets []*checker.Type) bool {
	if len(members) < 2 || len(targets) != 1 || !l.plainViewObject(targets[0]) || l.unionTagged(members) {
		return false
	}
	downcast := false
	for _, member := range members {
		if !l.plainViewObject(member) {
			return false
		}
		downcast = downcast || l.castAssignable(targets[0], member)
	}
	return downcast
}

// plainViewObject is an object type a checked view reads field by field: an object held as ir.Object that's
// no class instance, callable, array, tuple, dictionary or phantom extension.
func (l *lowering) plainViewObject(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(proven) || l.arrayLikeType(proven) || checker.IsTupleType(proven) {
		return false
	}
	if of, known := l.representation(proven); !known || of != ir.Object {
		return false
	}
	if base, _ := l.phantomExtension(proven); base != nil {
		return false
	}
	return len(l.checker.GetIndexInfosOfType(proven)) == 0 && len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) == 0
}

// unionTagged is whether some field holds a distinct literal in every member, the tag unionMemberProof
// proves a member cast by. A one-member enum's member tells no member apart (tagHoldsAnyNumber).
func (l *lowering) unionTagged(members []*checker.Type) bool {
	for _, property := range l.checker.GetPropertiesOfType(members[0]) {
		literals := []*checker.Type{}
		valid := true
		for _, member := range members {
			literal := l.fieldLiteral(member, property.Name)
			if literal == nil || l.tagHoldsAnyNumber(literal) {
				valid = false
				break
			}
			for _, previous := range literals {
				valid = valid && tagLiteralsDistinct(previous, literal)
			}
			literals = append(literals, literal)
		}
		if valid {
			return true
		}
	}
	return false
}

// untaggedUnionCast lowers an untagged union's cast (untaggedUnionArm) as a checked view of the target over
// the value's own runtime shape, with no test at the cast, as overlapViewCast does: a value of another member
// reads through it what it holds, and stops at its first read of a field it lacks, or holds as another
// representation, with the view's own located message. The value is the operand itself, so its identity is
// Node's. Its view is marked untagged, so in a graph program viewPairGuard pairs every source arm with the
// target's, related or not (Memory's item 4 on #axqgdck). It isn't marked cast: no CheckedCast wraps the value,
// so the view's entry is the operand, as a structural cast's is (checkViewWidenedReads).
func (l *lowering) untaggedUnionCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	views := len(l.viewSources)
	checked, err := l.overlapViewCast(node, value, source, target)
	if len(l.viewSources) > views {
		l.viewSources[views].untagged = true
	}
	return checked, err
}
