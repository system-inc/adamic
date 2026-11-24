# Mutants for the cast causes (#b5w3ycg, Nov 22)

## M1: one sentence again

`castStop` returns main's refusal, "a cast the runtime can't check" with castRepair's fix, for every cause in .ts as it does in .a.

Expected: `TestCastStopNamesItsCause` goes red on every .ts probe (each want names its cause), and `TestUncheckableCastsStayRefused` stays green, since every one of its probes is still a refusal naming adamic/no-unchecked-cast.

## M2: a ruled check refused

`castStop` answers `castUntaggedUnion` with a refusal in the NotYet's words, so the kind moves and the words don't.

Expected: `TestCastStopNamesItsCause/untagged_union` goes red on the kind (want refused=false), the words alone wouldn't catch it.

## M3: every array cast unnamed again (#53amvnn)

`arrayCastCause` answers castUnnamed for every array cast, as main did before #53amvnn.

Expected: `TestCastStopNamesItsCause` goes red on array_elements, mutable_array, tuple_from_array, tuple_from_tuples and array_arm (each names its cause), and stays green on the rest.

## M4: a field refinement refused (#bbssdcy, item 19)

`castStop` answers `castFieldRefinement` with a refusal in the NotYet's words, so the kind moves and the words don't.

Expected: `TestCastStopNamesItsCause/field_refinement` goes red on the kind (want refused=false), and stays green on the rest; the words alone wouldn't catch it.

## M5: a field refinement asked first (#bbssdcy, item 19)

`objectCastCause` asks the field refinement's case above `castWidenedTarget`'s, so a target that reads one field narrower and widens another reads as a refinement.

Expected: `TestCastStopNamesItsCause/widened_target` goes red on the kind (want refused=true, got the refinement's NotYet), and so does field_refinement_widening_another_field; the rest stay green.
