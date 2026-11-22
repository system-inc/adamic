# Mutants for the cast causes (#b5w3ycg, Nov 22)

## M1: one sentence again

`castStop` returns main's refusal, "a cast the runtime can't check" with castRepair's fix, for every cause in .ts as it does in .a.

Expected: `TestCastStopNamesItsCause` goes red on every .ts probe (each want names its cause), and `TestUncheckableCastsStayRefused` stays green, since every one of its probes is still a refusal naming adamic/no-unchecked-cast.

## M2: a ruled check refused

`castStop` answers `castUntaggedUnion` with a refusal in the NotYet's words, so the kind moves and the words don't.

Expected: `TestCastStopNamesItsCause/untagged_union` goes red on the kind (want refused=false), the words alone wouldn't catch it.
