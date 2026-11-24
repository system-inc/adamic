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
// union target keep their stops. A union with a field every member holds as a literal is tagged, whether or
// not the literals tell the members apart: unionMemberProof proves its cast or refuses it, and its refusals
// (duplicate enum aliases, mixed tag representations, a payload refinement) are rulings this arm keeps.
//
// The arm never turns a NotYet into a Refused. A cast the unchecked-widening rule would refuse (widened, as
// stringWideningRefusal judges a cast) keeps the untagged stop: a target narrowing a field to an enum literal at
// any depth (FlowCall, whose node is a CallExpression of kind SyntaxKind.CallExpression), and a readonly field the
// target makes writable (FlowLabel's antecedent). A field the target narrows to a string literal passes that rule
// and takes the arm: its read through the view is checked against the literal (viewedField's ViewAllowed), as any
// view's is. Those wait for their own road, a tag test at
// the cast with a write judgment fitted to a downcast (#qxt10dd; Diagnostics' must on the census, J-2f10jv9).
func (l *lowering) untaggedUnionArm(source *checker.Type, members, targets []*checker.Type) bool {
	if len(members) < 2 || len(targets) != 1 || !l.plainViewObject(targets[0]) || l.unionHoldsLiterals(members) || l.widened(source, targets[0], map[[2]*checker.Type]bool{}) != nil {
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

// unionHoldsLiterals is whether some field holds a literal in every member: a tag, distinct or not, which
// unionMemberProof judges. A one-member enum's member counts, so its union keeps the stop that's ruled for it
// (tagHoldsAnyNumber).
func (l *lowering) unionHoldsLiterals(members []*checker.Type) bool {
	for _, property := range l.checker.GetPropertiesOfType(members[0]) {
		held := true
		for _, member := range members {
			held = held && l.fieldLiteral(member, property.Name) != nil
		}
		if held {
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
// so the view's entry is the operand, as a structural cast's is (checkViewWidenedReads). Every view this cast's
// lowering made is marked, not one picked by its place in the list: each is this cast's, and marking one more
// untagged only makes the guard pair more, so no reorder inside view() can leave this cast's view unmarked
// (Lower's should).
func (l *lowering) untaggedUnionCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	views := len(l.viewSources)
	checked, err := l.overlapViewCast(node, value, source, target)
	for index := views; index < len(l.viewSources); index++ {
		l.viewSources[index].untagged = true
	}
	return checked, err
}
