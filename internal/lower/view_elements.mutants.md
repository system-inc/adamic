# Mutants for element-checked views of readonly arrays (V3, #6wfvfhm)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| No element contract | view_unions_dispatch.go, `unionAggregateContract`: the `contract.Unsupported = ""` under `contract.ArrayReadonly && elementContractCheckable(...)` removed, so every readonly array arm keeps "array element contract" | internal/oracle TestCheckedViewV2ArrayArmBoundary, the twin stops NotYet instead of agreeing with Node |
