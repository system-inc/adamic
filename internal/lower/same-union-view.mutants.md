# Same-union view mutants

Each change ran alone through a Go overlay on the Linux worker. The unmodified
tests passed in both .a and .ts. Each mutant run exited 1.

| Mutant | Edit | Red test and failure |
|---|---|---|
| A. Accept equal arm sets | In `sameUnionView`, filter the nullish difference as before, then compare union member sets instead of checker identity. | `TestInvarianceSameUnionViewStillJudgesDifferentIdentity` fails at `invariance_same_union_test.go:102` in both dialects. The two distinct named unions lower with no refusal. |
| B. Unregister 3350 | Delete the `narrowedArmReadGap` registration. | `TestInvarianceSameUnionViewDependsOnNarrowedArm` fails at `invariance_same_union_test.go:60` in both dialects. The stale arm read lowers with no diagnostic. |
| C. Allow nullish removal | Delete the early return for `sourceNullish&^viewNullish != 0` in `sameUnionView`. | `TestInvarianceSameUnionViewStillJudgesNullishRemoval` fails at `invariance_same_union_test.go:76`. Both undefined and null bypass the arm-by-arm judgment in both dialects. |

| D. Element check below the skip | In `widenedArguments`, drop the mutable element's assignability check, so the element recurses into `widened` unasked. | `TestInvarianceSameUnionViewStillJudgesContainerElement` fails: `(Flow \| undefined)[]` over a `Flow[]` reaches the skip and lowers. |

As a baseline, remove the entire skip from `widened`. The admission pin fails
at `invariance_same_union_test.go:30` in both dialects. The old refusal says
`Flow seen as Flow | undefined, which can write string[] | undefined where string[] is read`.
The dependence fixture stops at its optional view before the stale read. The
distinct-identity and nullish-removal pins still pass.
