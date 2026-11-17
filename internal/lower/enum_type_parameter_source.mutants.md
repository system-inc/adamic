# c5 mutants (#snypn8c), internal/lower/invariance.go

Each is one exact, unique line replaced, never applied in the tree. Each was applied on its own and its tests run by J-qv41p21 (sealed, azure, at a6b39edb), with the unmutated pins green in the same run.

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

## Measured, J-qv41p21

- M1 red: TestTypeParameterSourceIntoWholeEnumSlotLowers and TestTypeParameterSourceIsReadAtItsConstraint (2 tests).
- M2 red: TestTypeParameterSourceIntoMemberSlotIsRefusedAsItsConstraint's subtests constrained to number, a chain
  to number, constrained to a union of number and a member, and constrained to an intersection (4 subtests).
- M2 leaves two rows green: the conditional type (`T extends number ? T : never`, whose source is a
  conditional type and not a type parameter, so it never reaches the read) and
  TestTypeParameterSourceIntoWiderArrayStillRefuses. Both stay refused under M2, so a rule other than this
  read refuses them, and these pins hold that against any change to it. This run doesn't show which rule.
