# Mutants of andFalsyDropped and falsyArms (#snypn8c, c1)

Each mutant replaces exactly one line, checked to match once, runs
`go test -run 'TestAndFalsyDropped|TestAndLeft' ./internal/lower/`, and is put back from a copy of the head file,
checked byte for byte. Each is red beside the same pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `andFalsyDropped` (logical_and_left.go): `if parent == nil {` made `if parent == nil \|\| true {`, so the helper answers no | TestAndFalsyDroppedLowers/boolean_or, double_and_or (freshOrWidened's && case), number_and_assign, if_condition, element_flags, not_and (refuseWidening's site walk) |
| M1a | refuseWidening's site walk (invariance.go): `if andFalsyDropped(and) {` made `if false {` | TestAndFalsyDroppedLowers/number_and_assign, if_condition, element_flags, not_and; TestAndFalsyDroppedWalks |
| M1b | freshOrWidened's && case (invariance.go): `if !andFalsyDropped(node) {` made `if true {` | TestAndFalsyDroppedLowers/boolean_or, double_and_or; TestAndFalsyDroppedWalks |
| M2 | `andFalsyDropped`: an \|\| counted as dropping from either side, `&& ... Left == child` removed | TestAndFalsyDroppedStopsAreRefused/or_right |
| M3 | `andFalsyDropped`: `??` counted beside \|\| | TestAndFalsyDroppedStopsAreRefused/nullish |
| M4 | `falsyArms`: `if l.truthyLiteral(arm) && ...` made `if false && ...`, so truthy literals are kept again | TestAndLeftFalsyArms (flag, one, yes, word, bigOne, member); TestAndFalsyDroppedLowers/truthy_literal |
| M5 | `TestAndFalsyDroppedWalks`: the list of walks short by freshOrWidened | TestAndFalsyDroppedWalks |
| M6 | `openEnumMember`: `impure \|\| ...` made `impure && ...`, so a whole open numeric enum's nonzero members are dropped as truthy | TestAndLeftFalsyArms (ternary) |

M1a and M1b each take the helper out of one walk, so they name which walk holds which pin: a boolean left's false
reaches a whole Ternary slot only through freshOrWidened's && case, judged at the && under the ||, and a number or
flag left's 0 reaches a member slot (Ternary.False, ElementFlags.Optional) through the site walk, which judges the
left against the &&'s own type.

No row for a string literal left: tsc refuses `"a" && checked(n)` (TS2872, this kind of expression is always truthy),
probed by `adamic types` in J-w6kz3ab. `true && checked(n)` lowers, but native's emitter panics on it ("no conversion
from 2 to 1", internal/native/taste.go's logicalConverted), as it does on `true && next(n)` into a plain number with no
enum anywhere, which c1 doesn't change, so falsyArms' own pin (yes) holds `true` and truthy_literal holds 1.

Measured by Loom's proof job J-w6kz3ab at 401b17a1: the pins green at head, every mutant red as listed but for
element_flags and not_and, which came later; element_flags then had a None member, and its && typed two members,
lowered on main too. Both are measured with the rest by the part a job at the commit carrying this file.
