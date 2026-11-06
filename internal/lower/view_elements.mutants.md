# Mutants for element-checked views of readonly arrays (V3, #6wfvfhm)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No element contract | view_unions_dispatch.go, `unionAggregateContract`: the `contract.Unsupported = ""` under `contract.ArrayReadonly && elementContractCheckable(...)` removed, so every readonly array arm keeps "array element contract" | internal/oracle TestCheckedViewV2ArrayArmBoundary, the twin stops NotYet instead of agreeing with Node |
| A counted checked read | native/borrow.go, `pureKind`: ir.ViewElement removed from the list, so its array is retained per read | internal/native TestPassThroughsAreNotConsumers. Since the key, no fixture with a recorded counts row views an array, so no counts test sees it: the C of a viewed read carries the cost unpinned by counts until one does |
| A one-way key | view_arrays.go, `elementKeyed`: its second relation, `l.checker.IsTypeAssignableTo(element, viewed)`, removed, so a read is checked only when a viewed element type is assignable to its own | internal/lower TestViewElementKeyPredicateFilter, the Identifier read that filter(isIdentifier) hands out reads raw |
| An empty pre-pass | view_arrays.go, `collectViewedElements`: returns before its walk, so no element type is viewed | internal/lower TestViewElementKeyAgreesWithLowering, every fixture that views an array misses; each lie fixture still exits 70, since the misses lower it again with every read checked |
| An empty pre-pass, no fallback | the empty pre-pass, and lower.go, `lowerCheckedWithCycleTypes`: the misses ignored, so the first lowering stands | internal/oracle TestViewArrays, every lie fixture that views an array reads its lie raw and exits 0 |
| An unrecorded miss | view_unions_dispatch.go, `unionAggregateContract`: the miss never appended | internal/lower TestViewElementKeyMissesAnUnseenArray, a contract for an array the pre-pass never saw passes unnoticed |
| An unrelated read rebuilt from its type | view_arrays.go, `checkViewArrays`: the unrelated ViewElement put back with `Element: value.Of.Present()`, not `value.Element` | internal/oracle TestViewArrays/unrelated-maybe, the `(number \| undefined)[]` read takes a plain number's layout and disagrees with Node |

## Fixtures that wait on the unknown[] head (#6wfvfhm, Lower's ruling 3)

V3 keeps main's unknown[] refusals, so these internal/oracle/testdata/views_v3 fixtures are out of TestViewArrays' list until that head lands. Each one's point is the unknown[] road itself, which no typed array reaches.

- tagged-null-read.ts: a null in tagged storage read back through a `readonly unknown[]` parameter, which main stops as an array of unknown with erased element storage.
- unknown-wide.ts: a viewed `readonly Node[]` that holds a number, read through a `readonly unknown[]` parameter, the same erased storage stop.
- unknown-null.ts: a viewed `readonly Node[]` that holds a legacy null, read through a `readonly unknown[]` parameter, the same stop before its legacy nullable-pointer one.
- unknown-scalar.ts: a plain `number[]` read through a `readonly unknown[]` parameter, no view at all, the same stop.
- guard-ref-push.ts: pins unknownArrayMutation's reference push into an unknown[], whose literal main refuses as narrower first.
- guard-index-write.ts: pins unknownArrayMutation's index write into an unknown[], whose literal main refuses as narrower first.
- per-read.ts, its first half: it was a lie pushed through the unknown[] alias between two checked reads, ArrayPush's UnknownElements. A push through a typed alias of a viewed array stops as a read through push, so per-read.ts now holds the lie from the start, at the second read only.
