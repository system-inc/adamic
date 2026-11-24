# Mutants of enumComparedMember (#snypn8c, c3)

Each mutant replaces exactly one line, checked to match once, runs `go test -run TestEnumCompared ./internal/lower/`,
and is put back from a copy of the head file. Each but M6 is red beside the same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `freshOrWidened` (invariance.go): the `&& !l.enumComparedMember(...)` call removed | TestEnumComparedMemberLowers/case_chain, or_and, and_chain, conditional |
| M2 | `enumComparedMember`: where no comparison proves a set, the checker's narrowed type at the read trusted instead (`members = l.definedMembers(l.checker.GetTypeAtLocation(node))`) | TestEnumComparedMemberStopsAreRefused/excluded, arrow, default, fallthrough |
| M3 | `enumComparedMember`: the `const` test dropped, any variable declaration taken | TestEnumComparedMemberStopsAreRefused/let |
| M4 | `comparedMembers`: the function and class boundary test made `if false {` | TestEnumComparedMemberStopsAreRefused/arrow |
| M5 | `caseMembers`: the clause-with-statements test made `if false {` | TestEnumComparedMemberStopsAreRefused/fallthrough |
| M6 | `provedMembers`: `ast.KindExclamationEqualsEqualsToken` added beside `===` and `==` | none: equivalent, it survives |

M2 is the thing the rule exists to refuse: over an open enum the narrowed type can't tell an exclusion by `!==`,
which leaves every other number in, from a comparison that names the member. let stays green under it, since the
const test refuses a let before any proof is asked for.

M6, `!==` counted as proving the member it names, is equivalent: any member a `!==` names is one tsc's narrowing has
already excluded from the value written, so a slot that takes it and that value is the whole open enum, which no
member proof is asked for, and the member never fits a slot that is asked. Under && two such members intersect to
nothing, which proves nothing.

No stop row for two members of one value: in `enum K { A = 1, B = 1, C = 2 }`, `t === K.A && t === K.B` then
`take(t)` with `take(x: K.C)` is refused by tsc (TS2345: Argument of type 'K.A' is not assignable to parameter of
type 'K.C'), probed in J-gcs3ecp. The checker gives members of one value one type, so the read is narrowed to it and
the write never type-checks; an empty intersection is refused anyway.

Measured by Loom's proof job J-gcs3ecp at 8746bc3a (code 65ad075b): the pins green at head, M1 to M5 each red as
listed and M6 surviving, each file put back and checked byte for byte. M1 and M3 to M5 were red the same way in
J-f8nmpvp at b55201bb.
