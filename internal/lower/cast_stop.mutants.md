# Mutants for the cast causes (#b5w3ycg, Nov 22)

## M1: one sentence again

`castStop` returns main's refusal, "a cast the runtime can't check" with castRepair's fix, for every cause in .ts as it does in .a.

Expected: `TestCastStopNamesItsCause` goes red on every .ts probe (each want names its cause), and `TestUncheckableCastsStayRefused` stays green, since every one of its probes is still a refusal naming adamic/no-unchecked-cast.

## M2: a ruled check refused

`castStop` answers `castUntaggedUnion` with a refusal in the NotYet's words, so the kind moves and the words don't.

Expected: `TestCastStopNamesItsCause/untagged_union` goes red on the kind (want refused=false), the words alone wouldn't catch it.

## M3: an array's interface read as a dictionary

`castExtendsArray` answers false, so an interface whose base-type chain reaches Array or ReadonlyArray (tsc's NodeArray, MutableNodeArray, MissingList, SortedArray) falls through to castIndexSignature (#2v3xtgm, Compiler's ruling).

Expected: `TestCastStopNamesItsCause/missing_list` and `TestCastStopNamesItsCause/node_or_node_array` go red on the kind (a NotYet dictionary view where the array's refusal is wanted), and `TestCastStopNamesItsCause/map-like_dictionary` stays green, since a real dictionary never had an array base.
