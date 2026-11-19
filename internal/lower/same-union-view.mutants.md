# Same-union view mutants

Each change ran alone, its pin run by name, on Loom job J-tfpzjxs (kind mutant, adamic
c69096ac, sealed): every same-union pin passed unmodified in both .a and .ts first, then each
mutant's pin went red at the line below and the edit was undone before the next. A to C were
first run by the Codex build through a Go overlay; J-tfpzjxs is the record.

| Mutant | Edit | Red test and failure |
|---|---|---|
| A. Accept equal arm sets | In `sameUnionView`, filter the nullish difference as before, then compare union member sets instead of checker identity. | `TestInvarianceSameUnionViewStillJudgesDifferentIdentity` fails at `invariance_same_union_test.go:102` in both dialects. The two distinct named unions lower with no refusal. |
| B. Unregister 3350 | Delete the `narrowedArmReadGap` registration. | `TestInvarianceSameUnionViewDependsOnNarrowedArm` fails at `invariance_same_union_test.go:60` in both dialects. The stale arm read lowers with no diagnostic. |
| C. Allow nullish removal | Delete the early return for `sourceNullish&^viewNullish != 0` in `sameUnionView`. | `TestInvarianceSameUnionViewStillJudgesNullishRemoval` fails at `invariance_same_union_test.go:76`. Both undefined and null bypass the arm-by-arm judgment in both dialects. |
| D. Element check below the skip | In `widenedArguments`, drop the mutable element's assignability check, so the element recurses into `widened` unasked. | `TestInvarianceSameUnionViewStillJudgesContainerElement` fails at `invariance_same_union_test.go:125` in both dialects: `(Flow \| undefined)[]` over a `Flow[]` reaches the skip and lowers. |

As a baseline, remove the entire skip from `widened`. The admission pin fails
at `invariance_same_union_test.go:30` in both dialects. The old refusal says
`Flow seen as Flow | undefined, which can write string[] | undefined where string[] is read`.
The dependence fixture stops at its optional view before the stale read. The
distinct-identity and nullish-removal pins still pass.
