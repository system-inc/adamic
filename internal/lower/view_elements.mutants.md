# Mutants for element-checked views of readonly arrays (V3, #6wfvfhm)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No element contract | view_unions_dispatch.go, `unionAggregateContract`: the `contract.Unsupported = ""` under `contract.ArrayReadonly && elementContractCheckable(...)` removed, so every readonly array arm keeps "array element contract" | internal/oracle TestCheckedViewV2ArrayArmBoundary, the twin stops NotYet instead of agreeing with Node |
| A counted checked read | native/borrow.go, `pureKind`: ir.ViewElement removed from the list, so its array is retained per read | internal/native TestPassThroughsAreNotConsumers; internal/oracle TestMapUnionValueContextualArray and TestCountsAreRecorded, the recorded retains and releases of map_union_value_contextual_array.a, nbody_field_values.a, route_targets_callbacks.a and nested-functions/06_parser_token_state.a rise |
| A one-way key | view_arrays.go, `elementKeyed`: its second relation, `l.checker.IsTypeAssignableTo(element, viewed)`, removed, so a read is checked only when a viewed element type is assignable to its own | internal/lower TestViewElementKeyPredicateFilter, the Identifier read that filter(isIdentifier) hands out reads raw |
| An empty pre-pass | view_arrays.go, `collectViewedElements`: returns before its walk, so no element type is viewed | internal/lower TestViewElementKeyAgreesWithLowering, every fixture that views an array misses; each lie fixture still exits 70, since the misses lower it again with every read checked |
| An empty pre-pass, no fallback | the empty pre-pass, and lower.go, `lowerCheckedWithCycleTypes`: the misses ignored, so the first lowering stands | internal/oracle TestViewArrays, every lie fixture that views an array reads its lie raw and exits 0 |
| An unrecorded miss | view_unions_dispatch.go, `unionAggregateContract`: the miss never appended | internal/lower TestViewElementKeyMissesAnUnseenArray, a contract for an array the pre-pass never saw passes unnoticed |
| An unrelated read rebuilt from its type | view_arrays.go, `checkViewArrays`: the unrelated ViewElement put back with `Element: value.Of.Present()`, not `value.Element` | internal/oracle TestViewArrays/unrelated-maybe, the `(number \| undefined)[]` read takes a plain number's layout and disagrees with Node |
