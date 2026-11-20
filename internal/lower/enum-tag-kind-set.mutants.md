# Enum tag kind sets (#s8gy89r U1, an application of #00rd54t)

The flow walk reads the complete predicate body as a finite set, never a partial
answer. Its hazard, checker-identity, accessor and predicateProves checks remain.
The .a predicate verifier is unchanged (rulings.md:141); .ts keeps checked views.

`TestKindSetBodyMutantsAreCaught` uses buildcache.GoTestMutated. Fifteen independent,
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
| FalseSide: drop a tag when it meets the set, instead of when wholly inside it | TestKindSetFalseSideKeepsPartialTag | 227, `kind-set partial tag was dropped from the false side` |
| FirstKindFit: check only the first selected kind | TestKindSetClosedSourcesFitEveryKind/one_does_not_fit | 279, `kind-set all-kinds fit` |

The partial tag is K.B | K.C, and the predicate set K.A, K.B: C must stay on
the false side. R-T80 currently stops the full program with membership over an
open tag domain; the direct flow proof pins this step independently of that stop.
`TestKindSetPredicateLieStopsAtCheckedRead` also holds the checked-view road to
exit 70 on JavaScript and sanitized native: a program catch cannot hide the read.

The claim still comes from predicateOfSignature(GetResolvedSignature(call)).
For a closed S, replacing it by that claim requires closedKindSourcesFit for
all selected kinds in K(S), queried separately because closedKindFitKey includes
kind. TestKindSetClosedSourcesIgnoreKindsOutsideDomain pins a selected Body kind
outside K(Ex), with no source to fit, alongside the fitting Ident kind.
The mixed-fit pin uses a closed Ex: Ident's source fits, but ExtraNum at Num's
kind lacks NumLit's text and brand. Its honest twin extends NumLit and agrees
with Node on JavaScript and sanitized native. The first-kind-only mutant must
make the bad program lower, not merely reach a different compiler stop.

Five fail-open mutants run TestKindSetPredicateFallbacksKeepFamilyOne:
NotEqual (/not_equal), EnumIdentity (/two_enums), PrimitiveGuard
(/primitive_guard), TrueDefault (/true_default), and DuplicateCase
(/duplicate_case). Each must fail at the direct complete-body recognizer catch
at enum_tag_kind_set_test.go:184, rather than depending on a later refusal.
The duplicate_case_true_first row also rejects the reverse duplicate order.
The primitive guard row isolates the recognizer: a string's truthiness is not a
nullish test even when its length is compared against enum members.


U1 follow-up: k04 in TestKindSetPredicateFallbacksKeepFamilyOne names A = 1
and D = 1 separately. The checker interns them as one literal type, so the
production duplicate-case identity check is unchanged. DuplicateCase drops
that check; both /duplicate_case and /k04 fail at line 184. The catcher requires
both witnesses, not just one failing row. There is no extra duplicate mutant.

TestKindSetUnionClaimFitsEverySource holds the whole A | B claim against every
closed source. The control agrees with Node on JavaScript and sanitized native.
Odd declared, Odd built and an anonymous source carry A's kind but B's layout;
all three are refused at the narrowing with the family 1 words.
UnionClaimFit skips closedKindSourcesFit whenever claimed is a union, treating
its answer as true. Each /odd_.* row then lowers and fails at line 380:
`kind-set union claim <row>: a source fitting neither claimed member lowered`.
The catcher requires all three failures. Its independent producer is
TestProduct_KindSetUnionClaimFitMutant, with no subtests or other assertions.

TestKindSetEqualValueTagsStopAtCheckedRead pins today's reverse-alias behavior
for singleton and set predicates, unchanged from main 6cb495cf and f5008927:
Node prints true then undefined; JavaScript and sanitized native print true,
then stop with exit 70 and `adamic: panic: field read failed: n.d is not
initialized; expected string, found missing`. This is a checked-stop pin, not a
change to the singleton proof or the checked-view policy.
