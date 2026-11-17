# Mutants for a slot's enum part (#snypn8c c2, Adamic's ruling), internal/lower/enums.go

Each is one exact, unique line replaced in `openEnumSlot`, run on a house box, then put back.

## M1: the object counted again

The slot's members are counted whole, as `openNumericEnumType` counts them on main, so `Options | Emit` has one
member more than Emit and reads as a literal slot.

old:
	part := enumPart(flattenedMembers(l.checker.GetNonNullableType(slot)))
new:
	part := flattenedMembers(l.checker.GetNonNullableType(slot))

Must red: TestWholeEnumIntoObjectOrEnumSlotLowers (pin 1). Also expected red: TestNumberIntoObjectOrWholeEnumSlotLowers
and TestSlotEnumPartIsItsNumericMembers.

## M2: any enum member present makes the slot open

The slot is open whenever its first numeric member is a numeric enum's, whole or not.

old:
	return !impure && len(whole) == 1 && whole[0].size > 1 && whole[0].size == len(part)
new:
	return !impure && (len(whole) > 0 || len(part) > 0) && l.numericEnum(l.enumIdentity(part[0]))

Must red: TestNumberIntoObjectOrMemberSlotStaysRefused, both subtests, a_member and a_strict_subset (pin 2), and
TestNumberIntoWholeEnumBesideAnotherMemberStaysRefused (pin 4). Also expected red: TestSlotEnumPartIsItsNumericMembers.

## Measured

J-t6mhwaa (sealed, at 8abcc87c on main 910f21af, before M3 and the two-member rule), each mutant applied alone with the unmutated pins green in
the same run:

- M1 red: TestWholeEnumIntoObjectOrEnumSlotLowers, TestNumberIntoObjectOrWholeEnumSlotLowers and
  TestSlotEnumPartIsItsNumericMembers.
- M2 red: TestNumberIntoObjectOrMemberSlotStaysRefused's a_member and a_strict_subset,
  TestNumberIntoWholeEnumBesideAnotherMemberStaysRefused and TestSlotEnumPartIsItsNumericMembers.

