package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// overlapViewCast lowers a cast from a represented union to one object type that only overlaps it (overlapTarget):
// a checked view of the target over the value's own runtime shape, whatever its tag and its layout, with no test at
// the cast. A value that holds the target's fields reads through it; one that doesn't stops at its first read, with
// the view's own located message. A write through the view is judged member by member, past the member-literal
// slots the view checks at each read (viewCastEnumSlots): where a member could be written something its own slot
// can't hold, the cast is NotYet, naming the member.
//
// No shape reaches that per-member loop today: main's view-site rule (a cast never launders a view its slot would
// refuse) judges the same widening at the cast first and refuses it, stricter than this NotYet. The loop stays as
// defense in depth if that rule ever narrows; TestUnionToInterfaceCastsStayRefused's "a member the view could write
// past" pins the shape as refused by that rule's own words, so a lift goes red there and points here (Adamic, #ha4hbqt).
func (l *lowering) overlapViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Object {
		return nil, l.notYet(node, "a checked view of "+l.checker.TypeToString(target)+" over a union held as other than an object")
	}
	if len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return nil, l.notYet(node, "a dictionary checked view")
	}
	slots := l.viewCastEnumSlots(node, source, target)
	for _, member := range castMembers(source) {
		if widening := l.widenedPast(member, target, slots, map[[2]*checker.Type]bool{}); widening != nil {
			return nil, l.notYet(node, "a checked view of "+l.checker.TypeToString(target)+" that could write "+l.checker.TypeToString(member)+" a field it can't hold")
		}
	}
	return l.view(node, value, source, target)
}

// viewCastRefusal judges a view cast castRefusal recorded (1900), an overlap or a structural view, as
// refuseWidening judges a downcast, the operand's type seen as the target, but past the member-literal slots
// the view checks at each read (viewCastEnumSlots), and then the slot the cast's own type goes into
// (castSlotRefusal). A cast with no such slot, and an upcast, which is no view cast, are refuseWidening's
// exactly, as they always were.
func (l *lowering) viewCastRefusal(node *ast.Node) error {
	if l.stringReadOnlyArgument(node) || l.regexReplacementArgument(node) || l.nodeFSFileReadOnlyArgument(node) || l.regexCallbackIntrinsicView(node) {
		return nil
	}
	as := node.AsAsExpression()
	source, target := l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node)
	if l.checker.IsTypeAssignableTo(source, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target) {
		return l.refuseWidening(node)
	}
	slots := l.viewCastEnumSlots(node, source, target)
	if slots == nil {
		return l.refuseWidening(node)
	}
	if found := l.widenedPast(source, target, slots, map[[2]*checker.Type]bool{}); found != nil {
		return l.wideningRefusal(node, source, target, found)
	}
	return l.castSlotRefusal(node)
}

// viewCastEnumSlots is the member-literal slots a view cast's target names that its view checks at each read
// (#snypn8c c6, Adamic's ruling under rulings.md:18): present, and holding one of the slot's members, by the
// field's own contract, strictViewContract's scalar with the slot's literals allowed, which every read through a
// receiver of the holder's type is checked by. A check at the cast would be a snapshot a later write to the
// source could falsify, so the read is where it's made. widenedPast walks past exactly these, and judges every
// other slot as widened always has (Lower's seam: widened gets no mode, and no arm learns about views).
//
// The walk pairs the source's parts with the target's wherever widenedPast threads its past, and a slot is in
// only when every visit to it is one widened refuses for the enum alone: it's reached through readonly fields
// and readonly containers only, it's a finite set of one numeric enum's members (viewContractLiterals, so not a
// whole enum, a one-member enum or an enum object), and the source holds that enum's open value there. A visit
// through a writable field or a mutable container's element would let a write through the view store a value
// the read never checks, and a number, another enum's value or a non-member slot is refused as it always was;
// any one of those keeps the slot out wherever else it's met, since widenedPast names a slot by its holder type
// alone. The target's contract is built here, as view builds it, and a slot whose holder no contract reached
// from the target's checks as a literal field stays out, so a slot the view can't check fails closed.
func (l *lowering) viewCastEnumSlots(node *ast.Node, source, target *checker.Type) map[viewCastSlot]bool {
	walk := viewCastSlotWalk{in: map[viewCastSlot]bool{}, out: map[viewCastSlot]bool{}, visited: map[viewCastVisit]bool{}}
	l.walkViewCastSlots(&walk, source, target, true)
	if len(walk.in) == 0 {
		return nil
	}
	root, err := l.viewContract(node, target)
	if err != nil {
		return nil
	}
	reached := map[ir.ViewContractID]bool{}
	var reach func(ir.ViewContractID)
	reach = func(id ir.ViewContractID) {
		if id == 0 || reached[id] {
			return
		}
		reached[id] = true
		contract := l.result.ViewContracts[id-1]
		for _, field := range contract.Fields {
			reach(field.Contract)
		}
		for _, member := range contract.Members {
			reach(member)
		}
		reach(contract.Element)
	}
	reach(root)
	var slots map[viewCastSlot]bool
	for slot := range walk.in {
		if walk.out[slot] || !l.viewCastSlotChecked(slot, reached) {
			continue
		}
		if slots == nil {
			slots = map[viewCastSlot]bool{}
		}
		slots[slot] = true
	}
	return slots
}

// viewCastSlotWalk is what walkViewCastSlots has found: the slots each visit so far admits (in) and those some
// visit refuses (out), and the pairs it has walked.
type viewCastSlotWalk struct {
	in, out map[viewCastSlot]bool
	visited map[viewCastVisit]bool
}

// viewCastVisit is one pair walkViewCastSlots walks, with whether it's reached through readonly parts only.
type viewCastVisit struct {
	from, to *checker.Type
	readonly bool
}

// walkViewCastSlots pairs from's parts with to's as widenedPast does where it threads its past: a union's
// members on either side, a type parameter's constraint, a container's elements and each field's value. It may
// walk more than widenedPast visits, never less, so every visit widenedPast walks past is one this classified.
// readonly is whether every field and container on the way is one a write through the view can't reach.
func (l *lowering) walkViewCastSlots(walk *viewCastSlotWalk, from, to *checker.Type, readonly bool) {
	if l.sameUnionView(from, to) {
		return
	}
	from, to = l.withoutUndefined(from), l.withoutUndefined(to)
	visit := viewCastVisit{from: from, to: to, readonly: readonly}
	if from == to || walk.visited[visit] || from.Flags()&checker.TypeFlagsNever != 0 {
		return
	}
	walk.visited[visit] = true
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(from); constraint != nil && constraint != from {
			l.walkViewCastSlots(walk, constraint, to, readonly)
		}
		return
	}
	if members := l.definedMembers(from); len(members) > 1 {
		for _, member := range members {
			l.walkViewCastSlots(walk, member, to, readonly)
		}
		return
	}
	if members := l.definedMembers(to); len(members) > 1 {
		// A member the value is assignable to is walked as widenedPast walks it, and one it isn't is the
		// member tsc's decomposition reaches, whose fields widenedPast judges one by one (openEnumTagDecomposed).
		for _, member := range members {
			if member.Flags()&checker.TypeFlagsPrimitive != 0 {
				continue
			}
			if l.checker.IsTypeAssignableTo(from, member) {
				l.walkViewCastSlots(walk, from, member, readonly)
			} else if l.structured(from) {
				l.walkViewCastFields(walk, from, member, readonly, false)
			}
		}
		return
	}
	if to.Flags()&checker.TypeFlagsTypeParameter != 0 || !l.structured(from) || !l.structured(to) {
		return
	}
	// A record's values and a function's parameters and results are judged by widened, not widenedPast, so
	// nothing under them is walked past.
	if l.recordElement(from) != nil && l.recordElement(to) != nil {
		return
	}
	if len(l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)) > 0 && len(l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)) > 0 {
		return
	}
	fromContainers := l.containers(from)
	for _, viewed := range l.containers(to) {
		for _, inside := range fromContainers {
			if !l.sameContainer(inside, viewed) {
				continue
			}
			through := readonly && !l.mutableContainer(viewed)
			fromArguments, toArguments := l.checker.GetTypeArguments(inside), l.checker.GetTypeArguments(viewed)
			for index, element := range fromArguments {
				// A tuple seen as an array: every element of it is seen as the array's element (widenedArguments).
				if checker.IsTupleType(inside) && !checker.IsTupleType(viewed) && len(toArguments) == 1 {
					l.walkViewCastSlots(walk, element, toArguments[0], through)
				} else if index < len(toArguments) {
					l.walkViewCastSlots(walk, element, toArguments[index], through)
				}
			}
			break
		}
	}
	if to.Flags()&checker.TypeFlagsObject != 0 && len(l.containers(to)) > 0 {
		return
	}
	l.walkViewCastFields(walk, from, to, readonly, true)
}

// walkViewCastFields classifies each field of to that from holds, as widenedProperties visits it with to as the
// holder, and walks on into each one that isn't a member-literal slot when descend says widenedPast does. A
// method is judged by widened, never walked past.
func (l *lowering) walkViewCastFields(walk *viewCastSlotWalk, from, to *checker.Type, readonly, descend bool) {
	for _, viewed := range l.checker.GetPropertiesOfType(to) {
		inside := l.checker.GetPropertyOfType(from, viewed.Name)
		if inside == nil || viewed.Flags&ast.SymbolFlagsMethod != 0 || inside.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		source, slot := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)
		through := readonly && l.checker.IsReadonlySymbol(viewed)
		if l.viewCastMemberSlot(slot) {
			named := viewCastSlot{holder: to, field: viewed.Name}
			switch {
			case through && l.viewCastOpenEnum(source, slot):
				walk.in[named] = true
			case !through || !l.enumAssignable(source, slot) || !l.checker.IsTypeAssignableTo(source, slot):
				// A visit widened would refuse for more than the enum, or one a write through the view reaches.
				walk.out[named] = true
			}
			continue
		}
		if descend {
			l.walkViewCastSlots(walk, source, slot, through)
		}
	}
}

// viewCastMemberSlot is whether a field's type, less null and undefined, is a finite set of one numeric enum's
// members: what a view's contract checks a read against, and never a whole enum, which holds any number.
func (l *lowering) viewCastMemberSlot(slot *checker.Type) bool {
	slot = l.checker.GetNonNullableType(slot)
	return l.numericEnum(l.enumIdentity(slot)) && l.viewContractLiterals(slot) != nil
}

// viewCastOpenEnum is whether source, a field's type in the source, is the open value of the slot's enum: the
// whole enum, or two or more of its members, the values the enum arms of widened refuse in a member slot.
func (l *lowering) viewCastOpenEnum(source, slot *checker.Type) bool {
	slot = l.checker.GetNonNullableType(slot)
	if l.openNumericEnumType(source) {
		return l.enumIdentity(l.checker.GetNonNullableType(source)) == l.enumIdentity(slot)
	}
	return l.openEnumTagDecomposed(source, slot)
}

// viewCastSlotChecked is whether a contract reached from the view's checks the slot as a literal field: its
// holder's contract is an object one with the field, and the field's contract a scalar with literals allowed.
func (l *lowering) viewCastSlotChecked(slot viewCastSlot, reached map[ir.ViewContractID]bool) bool {
	id := l.result.ViewContractTypes[int(slot.holder.Id())]
	if id == 0 || !reached[id] {
		return false
	}
	holder := l.result.ViewContracts[id-1]
	if holder.Unsupported != "" || holder.Kind != ir.ViewObject {
		return false
	}
	for _, field := range holder.Fields {
		if field.Name != slot.field {
			continue
		}
		if field.Contract == 0 {
			return false
		}
		checked := l.result.ViewContracts[field.Contract-1]
		return checked.Unsupported == "" && checked.Kind == ir.ViewScalar && len(checked.Allowed) != 0
	}
	return false
}
