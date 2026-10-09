Built contextual module generic comparer/default values toward step 16, with source identity across specializations.
Commit: this commit, on compiler/generic-values above 0d0b927a and dependency 50654a40.
Checks: focused oracles pass 5.185s; release/source fixture selection passes 18.589s; full lower passes 89.713s; reader guard passes 49.513s; counts refresh passes 139.631s.
Mutants: false comparer returns on all three reductions and distinct per-adapter identities are caught by Node stdout in native and JavaScript; valid execution, sanitizers and leaks checked.
Uncovered: nested generic values, explicit instantiation expressions, unknown function descriptors, const aliases (next unit), higher-rank slot diagnostics (next unit).

The supplied ruling is marked approved in docs/generic-function-values.md. No task-thread reader is available in this environment, so the quoted ruling is the authority used.

The original fixtures reduce core.ts:220, core.ts:814 and utilities.ts:10629 without copying upstream implementations. An outer generic default composes its context with the caller mapper. The checker instantiates the source signature; explicit read-back evidence requires every binder to occur in the contextual contract. Existing universal body relations still prove the instantiated signature against the rigid outer binder.

Module declaration identity is separate from call ABI. Repeated reads reuse adapters; different specializations compare equal through represented callable unions, Map/Set keys and array lookup. Unknown function descriptors remain an explicit existing NotYet. Ordinary closure creation identity is unchanged. The identity mutant changes each adapter's SourceIdentity, so both equality and collection output disagree with Node.

New top-level oracle leaf durations in the focused run are in first-oracles.log.gz; every leaf is below 60 seconds. Full lower initially failed because setup did not expose pinned Node types at stage3/api/node_modules. A cache symlink also failed declaration identity recognition after realpath resolution. npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund installed the pinned declarations physically; the full suite then passed.

Commands (all output redirected to logs, commands bounded):
- GOPROXY=https://proxy.golang.org|direct bash cloud/setup.sh; source /workspace/adamic-tools/env.sh; nproc = 5 (four-CPU quota). Timing lines are in setup-turn2.log.gz.
- go test ./internal/oracle -run '^(TestGenericValue|TestStep16GenericOutcomes)' -timeout 90s -v
- go test ./internal/oracle -run '^TestNativeAgreesWithNode/(stage3|internal)/(fixtures|oracle)/(generic-values|generics|testdata)/(0[1-4]|map_foreach_closures|set_foreach_closures|closure_convention_host24)' -timeout 90s -v
- go test ./internal/lower -timeout 150s
- go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 210s -args -update-counts

The admission-delta tool is built from detached origin/compiler/admission-delta (543925aa), referencing the identical cohere pin; it is not merged. Final admission delta and census remeasurement follow the remaining units. Conservative scope: unsupported layouts retain explicit diagnostics rather than selecting an erased entry.
