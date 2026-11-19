# Enum tag kind sets (#s8gy89r U1, an application of #00rd54t)

The flow walk reads the complete predicate body as a finite set, never a partial
answer. Its hazard, checker-identity, accessor and predicateProves checks remain.
The .a predicate verifier is unchanged (rulings.md:141); .ts keeps checked views.

`TestKindSetBodyMutantsAreCaught` uses buildcache.GoTestMutated. Eight independent,
parallel top-level product tests build one binary apiece; their children run the
named direct grammar control, so another lowering stop cannot mask a mutation.
Unmutated controls also run in the focused suite and whole internal/lower.

| Mutation | Catching test | Catch line in enum_tag_kind_set_test.go |
| --- | --- | --- |
| Disjunction: omit the OR comparisons | TestKindSetBodyGrammar/disjunction | 90, `kind-set grammar disjunction` |
| Switch: omit switch recognition | TestKindSetBodyGrammar/switch | 90, `kind-set grammar switch` |
| SwitchTail: ignore false after the switch | TestKindSetBodyGrammar/switch_tail | 90, `kind-set grammar switch_tail` |
| Alias: forget the const's symbol | TestKindSetBodyGrammar/alias | 90, `kind-set grammar alias` |
| Guard: omit the parameter && comparisons form | TestKindSetBodyGrammar/guard | 90, `kind-set grammar guard` |
| DoubleGuard: omit !!p normalization | TestKindSetBodyGrammar/double_guard | 90, `kind-set grammar double_guard` |
| IfGuard: omit if (p) { ... } return false | TestKindSetBodyGrammar/if_guard_alias | 90, `kind-set grammar if_guard_alias` |
| FalseSide: drop a tag when it meets the set, instead of when wholly inside it | TestKindSetFalseSideKeepsPartialTag | 213, `kind-set partial tag was dropped from the false side` |

The partial tag is K.B | K.C, and the predicate set K.A, K.B: C must stay on
the false side. R-T80 currently stops the full program with membership over an
open tag domain; the direct flow proof pins this step independently of that stop.
`TestKindSetPredicateLieStopsAtCheckedRead` also holds the checked-view road to
exit 70 on JavaScript and sanitized native: a program catch cannot hide the read.
