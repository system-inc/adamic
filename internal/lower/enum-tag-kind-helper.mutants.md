# Pure enum-kind helpers (#s8gy89r U2, application of #00rd54t)

kindHelperMaxDepth is 8 simultaneously active helpers. A depth exhaustion or
cycle rejects the complete outer predicate. Only one plain direct argument is
read: the node parameter's own field, or a scalar helper's parameter. Saved
aliases are deliberately outside this head's call-argument grammar. U1's guard,
true/false set filtering, per-kind source fit, signature claim and checker
identity checks are unchanged. There is no new lowering or IR field.

Each producer is a top-level TestProduct_KindHelper<Name>Mutant: t.Parallel(),
one buildcache.GoTestMutated build, no subtests and no additional assertions.
TestKindHelperMutantsAreCaught requires the intended direct recognizer witness,
so a later lowering refusal cannot conceal a fail-open body parser. The normal
rows also run in TestKindHelperShapesAgreeWithNode and the whole lower suite.

| Mutant | Catching test | Catch line in enum_tag_kind_helper_test.go |
| --- | --- | --- |
| Depth | TestKindHelperFallbacksRejectWholeBody/depth | 118 |
| Cycle | TestKindHelperFallbacksRejectWholeBody/cycle | 118 |
| NonHelper | TestKindHelperFallbacksRejectWholeBody/(non_helper|other_read) | 118 |
| WrittenParameter | TestKindHelperFallbacksRejectWholeBody/written_parameter | 118 |
| WrongArgument | TestKindHelperFallbacksRejectWholeBody/(wrong_field|arithmetic_argument) | 118 |
| Call | TestKindHelperShapesAgreeWithNode/default_chain | 45 |
| Scalar | TestKindHelperShapesAgreeWithNode/default_chain | 45 |
| Default | TestKindHelperShapesAgreeWithNode/default_chain | 45 |
| Tail | TestKindHelperShapesAgreeWithNode/switch_tail | 45 |
| ReturnBody | TestKindHelperShapesAgreeWithNode/return_chain | 45 |
| Arrow | TestKindHelperShapesAgreeWithNode/arrow | 45 |
| DefaultMask | TestKindHelperDefaultExcludesExplicitFalse | 72 |
| DefaultParameter | TestKindHelperFallbacksRejectWholeBody/default_parameter | 118 |
| Enum | TestKindHelperFallbacksRejectWholeBody/two_enums | 118 |
| Overload | TestKindHelperFallbacksRejectWholeBody/overload | 118 |
| Let | TestKindHelperFallbacksRejectWholeBody/let_helper | 118 |
| AliasArgument | TestKindHelperFallbacksRejectWholeBody/alias_argument | 118 |
| MixedDisjunction | TestKindHelperFallbacksRejectWholeBody/mixed_disjunction | 118 |

Depth disables the limit and admits a finite nine-helper chain. Cycle treats a
back edge as an empty successful contribution (the partial-set memo bug), rather
than rejecting it; merely dropping that check still fails closed at depth 8.
NonHelper incorrectly keeps a partial set when the rest of a helper body reads
something else or calls a non-helper. Both rows must fail. WrongArgument must
fail on both the other-field and arithmetic-argument rows. WrittenParameter
reads only the last return, ignoring the assignment before it. DefaultMask
incorrectly keeps a delegated kind shadowed by an explicit false case. The
remaining mutations remove their named gate or lose their grammar addition.

The default helper's true set excludes every explicit case label, including
false labels. Duplicate labels still use U1's checker identity check. The same
enum identity is shared through the complete chain. Calls, reads, clauses or
statements outside this grammar invalidate the entire outer body.

.a's predicate verifier is unchanged: each admitted .ts outer shape remains
Refused there with `a type predicate whose return is not proven (return
expression is not a trusted check on n)`. A const boolean-return arrow is
admitted. Without its return annotation, TS infers a predicate signature and
TestKindHelperInferredArrowKeepsPredicateBoundary pins the existing .ts NotYet:
`a call through a predicate signature nothing checks or proves`.
TestKindHelperLieStopsAtCheckedRead holds the lying predicate to exit 70 on
JavaScript and sanitized native at `field read failed: n.a is not a string`.
The two isExpression/isStatement fallback rows also keep the family 1 refusal
through full Lower, not just the direct recognizer.
