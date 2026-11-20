# c5' mutants (#snypn8c), internal/lower/invariance.go

Each is one exact, unique line replaced, never applied in the tree. Not yet measured: each is to be applied on its
own and its tests run, with the unmutated pins green in the same run.

## M1: the read removed

An indexed access over a type parameter falls through to the arms below as on main, so the open enum arm asks
enumAssignable of T["kind"] and refuses it.

old:
	if from.Flags()&checker.TypeFlagsIndexedAccess != 0 && typeParameterAccess(from) && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool {
new:
	if false && from.Flags()&checker.TypeFlagsIndexedAccess != 0 && typeParameterAccess(from) && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool {

Must red: TestIndexedAccessSourceIntoWholeEnumSlotLowers (pin 1), TestIndexedAccessSourceIsReadAtItsConstraint,
TestIndexedAccessSourceInUnionMemberLowers (pin 2) and TestIndexedAccessSourceInUnionMemberIsReadAtItsConstraint.

## M2: the object's type parameter guard dropped

Any indexed access is read at its base constraint, so SyntaxNode[P] with P extends "kind" is read at E and its
binding to token's K extends E passes.

old:
	if from.Flags()&checker.TypeFlagsIndexedAccess != 0 && typeParameterAccess(from) && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool {
new:
	if from.Flags()&checker.TypeFlagsIndexedAccess != 0 && true && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool {

Must red: TestIndexedAccessSourceOverAnObjectStaysRefused (pin 6).

## M3: the generic constraint guard dropped

T["kind"] with T extends { kind: U } and U extends E is read at its base constraint, which reads on through U to
E, so its binding to token's K extends E passes.

old:
	if own := constraintOfIndexedAccess(l.checker, access); own == nil || couldContainTypeVariables(l.checker, own) {
new:
	if own := constraintOfIndexedAccess(l.checker, access); own == nil {

Must red: TestIndexedAccessSourceWithGenericConstraintStaysRefused (pin 5).
