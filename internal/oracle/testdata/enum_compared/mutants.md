# Mutants of enumComparedMember (#snypn8c, c3)

Each mutant replaces exactly one line, checked to match once, runs `go test -run TestEnumCompared ./internal/lower/`,
and is put back from a copy of the head file. Each is red beside the same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `freshOrWidened` (invariance.go): the `&& !l.enumComparedMember(...)` call removed | TestEnumComparedMemberLowers/case_chain, or_and, and_chain, conditional |
| M2 | `provedMembers`: `ast.KindExclamationEqualsEqualsToken` added beside `===` and `==` | TestEnumComparedMemberStopsAreRefused/excluded |
| M3 | `enumComparedMember`: the `const` test dropped, any variable declaration taken | TestEnumComparedMemberStopsAreRefused/let |
| M4 | `comparedMembers`: the function and class boundary test made `if false {` | TestEnumComparedMemberStopsAreRefused/arrow |
| M5 | `caseMembers`: the clause-with-statements test made `if false {` | TestEnumComparedMemberStopsAreRefused/fallthrough |

M2 reds excluded by way of the && arm: each `!==` read as proving its member, `t !== K.C && t !== K.D` proves the
intersection of {K.C} and {K.D}, which is empty, a branch that never runs, and fits any slot. A lone `!==` can't
be pinned this way: a slot taking every member it leaves and the one it names is the whole open enum, which no
member proof is asked for.

Job: see the commit that lands this file.
