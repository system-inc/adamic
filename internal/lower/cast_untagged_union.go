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
// or a union of them (union target (a), #b5w3ycg), each a member or narrower than one (castAssignable to it),
// so the cast is a downcast; a sideways target keeps its stop. A union target's view reads the checker's common
// properties of its members, each through its own field contract, and a field the members hold in different
// representations stops at that contract (supportsUntaggedRead, view_unions_untagged.go), which is (b)'s, never
// a read this arm admits (Lower, on #b5w3ycg). A union with a field every member holds as a literal is tagged,
// whether or not the literals tell the members apart: unionMemberProof proves its cast or refuses it, and its
// refusals (duplicate enum aliases, mixed tag representations, a payload refinement) are rulings this arm
// keeps. A target union tagged that way is the tag road's too, so it keeps its stop here.
//
// The arm never turns a NotYet into a Refused. A cast the unchecked-widening rule would refuse, as
// untaggedWidening judges it, keeps the untagged stop: a target narrowing a field to an enum literal anywhere but
// its tag (FlowCall, whose node is a CallExpression of kind SyntaxKind.CallExpression), and a readonly field the
// target makes writable (FlowLabel's antecedent). A field the target narrows to a string literal passes that rule
// and takes the arm: its read through the view is checked against the literal (viewedField's ViewAllowed), as any
// view's is. A target whose tag every member holds open takes the arm too (tsc's `node as CallExpression` of an
// Expression | Statement, each declaring kind as the whole SyntaxKind): untaggedUnionCast tests that tag at the
// cast, so the write judgment walks past it, and judges every other slot as widened does (#qxt10dd; Diagnostics'
// must on the census, J-2f10jv9).
func (l *lowering) untaggedUnionArm(node *ast.Node, source, target *checker.Type, members, targets []*checker.Type) bool {
	if len(members) < 2 || len(targets) == 0 || l.unionHoldsLiterals(members) || len(targets) > 1 && l.unionHoldsLiterals(targets) {
		return false
	}
	for _, member := range members {
		if !l.plainViewObject(member) {
			return false
		}
	}
	for _, wanted := range targets {
		if !l.plainViewObject(wanted) {
			return false
		}
		downcast := false
		for _, member := range members {
			downcast = downcast || l.castAssignable(wanted, member)
		}
		if !downcast {
			return false
		}
	}
	// Judged last, since a tag field's judgment builds the target's view contract (viewCastEnumSlots), which
	// only a cast the arm would take otherwise should pay for.
	return l.untaggedWidening(node, source, target, members) == nil
}

// untaggedWidening is the arm's write judgment: the first slot at which the union, seen as the target, could be
// written something it can't hold. A target with a tag field (untaggedTagField) is judged by widenedPast, past
// that one field alone, and only where viewCastEnumSlots finds it a member-literal slot the view checks at each
// read and reached through readonly fields only. The tag test at the cast checks it before the view is had, and a
// write through the view can't reach it, so seeing the members' open kind as the target's literal stores nothing
// the target's readers don't hold. Every other slot, an enum slot one field down included, is judged as widened
// judges it, so FlowCall's node.kind keeps its stop. A slot is kept by its holder and its field together, the
// target's own tag and never a field of that name on some other holder, so an entry that can't be told to be the
// tag's stays out and the judgment fails closed to widened's. A target with no tag field is widened's exactly.
func (l *lowering) untaggedWidening(node *ast.Node, source, target *checker.Type, members []*checker.Type) *widening {
	field := l.untaggedTagField(members, target)
	if field == "" {
		return l.widened(source, target, map[[2]*checker.Type]bool{})
	}
	var slots map[viewCastSlot]bool
	for slot := range l.viewCastEnumSlots(node, source, target) {
		if slot.holder != target || slot.field != field {
			continue
		}
		if slots == nil {
			slots = map[viewCastSlot]bool{}
		}
		slots[slot] = true
	}
	return l.widenedPast(source, target, slots, map[[2]*checker.Type]bool{})
}

// untaggedTagField is the field an untagged union's cast tests at the cast (#qxt10dd): the first property of a
// one-object target that the target holds as a single literal (fieldLiteral), declared readonly and required in
// the target and in every member, or "" where there's none. Readonly on every side means no write through the
// view or through any member's name can make the tag lie after the test, and required means every value the
// test reads holds the field. Readonly is held twice: viewCastEnumSlots walks past a slot reached through readonly
// fields only, so a writable kind would never be walked past here either, and no pin can catch this test's
// removal alone. A union target, or a field any member lacks, writes, or leaves optional, names no
// tag, and the cast keeps the view-only lowering it had.
func (l *lowering) untaggedTagField(members []*checker.Type, target *checker.Type) string {
	if target.Flags()&checker.TypeFlagsObject == 0 {
		return ""
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if l.fieldLiteral(target, property.Name) == nil || !l.checker.IsReadonlySymbol(property) || property.Flags&ast.SymbolFlagsOptional != 0 {
			continue
		}
		held := true
		for _, member := range members {
			inside := l.checker.GetPropertyOfType(member, property.Name)
			held = held && inside != nil && l.checker.IsReadonlySymbol(inside) && inside.Flags&ast.SymbolFlagsOptional == 0
		}
		if held {
			return property.Name
		}
	}
	return ""
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
// the value's own runtime shape, as overlapViewCast does: a value of another member reads through it what it
// holds, and stops at its first read of a field it lacks, or holds as another representation, with the view's
// own located message. The value is the operand itself, so its identity is Node's. Its view is marked untagged,
// so in a graph program viewPairGuard pairs every source arm with the target's, related or not (Memory's item 4
// on #axqgdck). Every view this cast's lowering made is marked, not one picked by its place in the list: each is
// this cast's, and marking one more untagged only makes the guard pair more, so no reorder inside view() can
// leave this cast's view unmarked (Lower's should).
//
// A target with a tag field (untaggedTagField) has it tested at the cast, as interfaceCast tests one literal,
// with row 9's message (#qxt10dd): the arm's write judgment walked past that field because this test checks it,
// so a value of another kind stops here, before any read, even one whose other fields the target's view would
// read as it holds them. A tag the target narrows to a string literal is tested the same way, beside the read
// check its view already makes. The CheckedCast is then the view's entry, so its views are marked cast too, as
// interfaceCast's are (checkViewWidenedReads). A literal no constant can say keeps the cast stopped. A target
// with no tag field is lowered as it was, a view with no test at the cast, and not marked cast: no CheckedCast
// wraps the value, so the view's entry is the operand, as a structural cast's is.
func (l *lowering) untaggedUnionCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	field := l.untaggedTagField(castMembers(source), target)
	var allowed ir.Expression
	var fieldType ir.Type
	if field != "" {
		var constant bool
		if allowed, fieldType, constant = l.literalConstant(l.fieldLiteral(target, field)); !constant {
			return nil, l.notYet(node, "a checked view of "+l.checker.TypeToString(target)+" over the union "+l.checker.TypeToString(source)+", whose tag "+field+" no constant can test")
		}
	}
	views := len(l.viewSources)
	checked, err := l.overlapViewCast(node, value, source, target)
	for index := views; index < len(l.viewSources); index++ {
		l.viewSources[index].untagged = true
		if field != "" {
			l.viewSources[index].cast = true
		}
	}
	if err != nil || field == "" {
		return checked, err
	}
	return ir.CheckedCast{Value: checked, Field: field, FieldType: fieldType, Allowed: []ir.Expression{allowed}, CheckedFields: true, Message: l.castFailure(source, target)}, nil
}
