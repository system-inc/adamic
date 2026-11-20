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
		through := readonly && l.checker.IsReadonlySymbol(viewed)
new:
		through := readonly

Must red: TestViewCastSlotsStayRefused's a_writable_parent. The overlap's Wider member holds a writable parent, and
its Doc's kind is then walked past, so the cast lowers.

## Measured

Not yet run.
