# Mutants for element-checked views of readonly arrays (V3, #6wfvfhm)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No element contract | view_unions_dispatch.go, `unionAggregateContract`: the `contract.Unsupported = ""` under `contract.ArrayReadonly && elementContractCheckable(...)` removed, so every readonly array arm keeps "array element contract" | internal/oracle TestCheckedViewV2ArrayArmBoundary, the twin stops NotYet instead of agreeing with Node |
| A counted checked read | native/borrow.go, `pureKind`: ir.ViewElement removed from the list, so its array is retained per read | internal/native TestPassThroughsAreNotConsumers; internal/oracle TestMapUnionValueContextualArray and TestCountsAreRecorded, the recorded retains and releases of map_union_value_contextual_array.a, nbody_field_values.a, route_targets_callbacks.a and nested-functions/06_parser_token_state.a rise |
