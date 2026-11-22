# Mutants for a view cast's member-literal slots (#snypn8c c6, Adamic's ruling, Lower's seam)

Each is one exact, unique line replaced, in `internal/lower/cast_overlap_view.go` or `internal/lower/invariance.go`,
run on a house box, then put back.

## M1: the skip set dropped from widened's call

The view cast's relation is walked with no slots, so it's refused at the cast as on main.

In `viewCastRefusal`, old:
	if found := l.widenedPast(source, target, slots, map[[2]*checker.Type]bool{}); found != nil {
new:
	if found := l.widenedPast(source, target, nil, map[[2]*checker.Type]bool{}); found != nil {

Must red: TestViewCastMemberSlotsLower, all four subtests. Also expected red: the oracle's
TestViewCastEnumSlotsHonestAgreeWithNode and TestViewCastEnumSlotsLiesStopAtTheRead, each at `lowered`.

## M2: every field walked past, not just the recorded ones

Any view cast with a slot recorded skips every field visit, Lower's ask: some refused row must turn.

In `widenedProperties`, old:
		if inside == nil || skip[viewed.Name] || past[viewCastSlot{to, viewed.Name}] {
new:
		if inside == nil || skip[viewed.Name] || past != nil {

Must red: TestViewCastSlotsStayRefused's a_whole_enum_slot (the one-member enum's slot, whose contract allows any
number, lowers). Also expected red: a_writable_parent and a_whole_object_slot. a_number_in_a_member_slot stays
green, since no slot is recorded there and the walk is refuseWidening's.

## M3: the readonly-only test dropped

A slot reached through a writable field is recorded as if the field were readonly.

In `walkViewCastFields`, old:
		through := path.readonly && l.checker.IsReadonlySymbol(viewed)
new:
		through := path.readonly

Must red: TestViewCastSlotsStayRefused's a_writable_parent. The overlap's Wider member holds a writable parent, and
its Doc's kind is then walked past, so the cast lowers.

## M4: the unbuilt bit set past every field, whatever its contract

A slot behind any field is walked past as if every field's contract were unbuilt, so a holder the view has no
contract for is walked past too.

In `walkViewCastFields`, old:
			if field.contract != 0 && strings.HasPrefix(l.result.ViewContracts[field.contract-1].Unsupported, unbuiltFieldFamily) {
new:
			if true || strings.HasPrefix(l.result.ViewContracts[field.contract-1].Unsupported, unbuiltFieldFamily) {

Must red: TestViewCastSlotsUnderAnotherIdentityStayRefused, and TestViewCastSlotsStayRefused's
a_holder_with_no_contract.

## M5: the unbuilt bit never set

A slot behind an unbuilt field is judged only by the checked arm, as c6 alone judged it.

In `walkViewCastFields`, old:
				field.behind = true
new:
				field.behind = path.behind

Must red: TestViewCastSlotsBehindAnUnbuiltFieldLower, the cast refused at Holder's kind.

## Measured

Not yet run.
