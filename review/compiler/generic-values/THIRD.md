Built the ruled higher-rank refusal unit #td6c9vy toward step 16, with named slot paths and free binders in .a and .ts.
Commit: this commit, following pushed alias unit e2dca338.
Checks: full internal/lower passes 53.886s; focused leaves pass 0.367s; generic-value Node/native/JavaScript oracles pass 2.691s; TestCallTargetReaders passes 17.837s; counts refresh passes 87.426s; explicit go vet passes.
Mutants: disabling the higher-rank collector fails four diagnostic assertions (0.253s); a declared-default wrong result is rejected by source Node in both backends (1.16s).
Uncovered: explicit instantiation expressions and nested generic callable escapes remain NotYet; census and admission-delta proof are the following unit.

Higher-rank slots are refused before callable lowering, including a polymorphic field, parameter, and inferred array element. Diagnostics name every signature binder and its slot path, with concrete-slot and direct const-alias fixes. docs/0.1.md records the ruled reason: no monomorphized instance can be selected, and an erased entry would trust an unproved type. The planted Node control prints 4 x, while both file extensions receive the ruled refusal. An unused source binder with no declared default is also refused rather than silently inferred as unknown. A declared type-parameter default can select the missing binder; its value is checked against Node in both backends and a wrong-result mutant.

The graph scan has a conservative finite traversal bound (depth 32 or 1,024 visited object types), with an explicit Refused diagnostic and concrete-slot fix when exceeded. It never silently accepts an incomplete scan. An initial unbounded scan exhausted memory; a NotYet bound then displaced the existing growing-generic-class Refused diagnostic. The final bounded Refused preserves that policy and the full lowering suite passes. Intrinsic const aliases retain their existing opaque-token specialization rule and are still checked at escape.

Commands (all output logged; outer timeout 120s except counts 210s):
- go test ./internal/lower -timeout 90s
- go test ./internal/lower -run '^(TestHigherRank|TestGenericValueUnusedBinder|TestGenericValueDeclaredDefault)' -timeout 90s -v
- go test ./internal/oracle -run '^TestGenericValue' -timeout 90s -v
- go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 210s -args -update-counts
- go test -overlay=/tmp/generic-values-rank-mutant.json ./internal/lower -run '^TestHigherRank' -timeout 90s -v (expected test failure, no compile failure)
- go vet ./internal/lower ./internal/oracle

New/touched leaves: higher-rank .a/.ts/parameter/container and unused-binder checks each below 0.12s; declared-default Node/native/JavaScript control 0.36s; higher-rank Node control 0.65s; declared-default wrong-result mutant 1.16s. No fixture count row is added for the refused program; the successful count refresh changes no counts.
