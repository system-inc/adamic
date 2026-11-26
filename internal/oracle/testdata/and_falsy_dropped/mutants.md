# Mutants of andFalsyDropped and falsyArms (#snypn8c, c1)

Each mutant replaces exactly one line, checked to match once, runs
`go test -run 'TestAndFalsyDropped|TestAndLeft' ./internal/lower/`, and is put back from a copy of the head file.
Each is red beside the same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `andFalsyDropped` (logical_and_left.go): `if parent == nil {` made `if parent == nil \|\| true {`, so the helper answers no | measuring |
| M1a | refuseWidening's site walk (invariance.go): `if andFalsyDropped(and) {` made `if false {` | measuring |
| M1b | freshOrWidened's && case (invariance.go): `if !andFalsyDropped(node) {` made `if true {` | measuring |
| M2 | `andFalsyDropped`: an \|\| counted as dropping from either side, `&& ... Left == child` removed | measuring |
| M3 | `andFalsyDropped`: `??` counted beside \|\| | measuring |
| M4 | `falsyArms`: `if l.truthyLiteral(arm) && ...` made `if false && ...`, so truthy literals are kept again | measuring |
| M5 | `TestAndFalsyDroppedWalks`: the list of walks short by freshOrWidened | measuring |
| M6 | `openEnumMember`: made `return false`, so a whole open numeric enum's nonzero members are dropped as truthy | measuring |

M1a and M1b each take the helper out of one walk, so they name which walk holds which pin.
