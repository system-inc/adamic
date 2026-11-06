# Mutants for element-checked views of readonly arrays (V3, #6wfvfhm)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No element contract | view_unions_dispatch.go, `unionAggregateContract`: the `contract.Unsupported = ""` under `contract.ArrayReadonly && elementContractCheckable(...)` removed, so every readonly array arm keeps "array element contract" | internal/oracle TestCheckedViewV2ArrayArmBoundary, the twin stops NotYet instead of agreeing with Node |
| Unchecked union reads | view_arrays.go, `viewArrayElementCheck`: return nil before the contract is built, so no element read is checked | internal/oracle TestMapUnionValueContextualArray and its TestCountsAreRecorded row read 38 retains and 53 releases against the recorded 42 and 57, the price of the element-type key; and TestViewArrays and TestViewArrayAllowlist go red on every checked road |
