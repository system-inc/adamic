# c5 mutants (#snypn8c), internal/lower/invariance.go

Each is one exact, unique line replaced, never applied in the tree; proof.sh applies each, runs its tests, and restores.

## M1: the constraint read skipped

The constrained type parameter source falls through to the arms below as on main, so the open enum arm asks
enumAssignable of TKind and refuses it. (The read itself sat lower on main; skipping the guard sends TKind to
the open enum arm first, exactly main's road for pin 1.)

old:
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool { return member.Flags()&checker.TypeFlagsTypeParameter != 0 }) {
new:
	if false && from.Flags()&checker.TypeFlagsTypeParameter != 0 && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool { return member.Flags()&checker.TypeFlagsTypeParameter != 0 }) {

Must red: TestTypeParameterSourceIntoWholeEnumSlotLowers (pin 1) and TestTypeParameterSourceIsReadAtItsConstraint.

## M2: a type parameter passes

The read replaced by a pass for every constrained type parameter source (an unconstrained one already passes).

old:
			return l.widened(constraint, to, visited)
new:
			return nil

Must red: TestTypeParameterSourceIntoMemberSlotIsRefusedAsItsConstraint, its subtests "constrained to number"
(pin 2) and "a chain to number" (pin 4).
