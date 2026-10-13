Built step (2): escaping .ts callable execution through cached checking adapters.
Commit: compiler/views-v4 commit carrying this report, after cache revision 19df538d4.
Tests: oracle 4.786s; focused lower 0.946s, native 7.522s, JavaScript 1.125s.
Mutants: escaped argument and result omission both caught in native, sanitized native and JavaScript against source Node.
Pending: step (3) language identity/no-stacking witnesses, actual later-call-site blame (precision 5), runtime conditions 3/4.

Native reads intern one generated checking closure per underlying/view-type identity. Wrapper code shares immediate-call dispatch: producer arguments are checked and converted from the viewed ABI after all argument expressions evaluate; producer code runs; the result is checked and converted to the view result. Producer metadata remains declaration-derived. Ordinary closures keep their existing ABI. Count-aware programs route closure calls through the convention dispatcher so an adapter carries the actual argument count even if producer target inference only sees the root callable.

JavaScript uses a WeakMap keyed by underlying callable and an inner WeakMap keyed by stable view-type token. Cached adapters are WeakRefs; the cache keeps neither callable nor adapter alive. Adapter-to-root metadata is a WeakMap. The adapter itself closes over its underlying. Both backends already normalize re-viewing at their interning boundary; the complete language identity and bounded re-view integration leaves remain for step (3).

The .ts unavailable-cache boundary is now a passing read-only control. Unused callable signature misfits do not fail on read. Existing presence/readiness/kind field checks remain; this step moves signature relation checking from read to invocation. Unproven .a escapes remain refused with path and fix. .ts unsupported finite value contracts remain refused; recursive target refusal is pinned. Higher-order/generic/function-surface expansion is outside this unit. Actual later source call-site blame is deferred as allowed: diagnostics currently name callable, adapter invocation boundary, creation/read site and both types. The oracle divergence list states this explicitly.

Enabled TestV4EscapeAdapterArgument (1.18s) and Result (1.21s). Both pin exact exit 70/stdout/stderr across backends, Node's argument/producer effect order, and finishing omission mutants that exactly match unchecked source Node and pass leak checks. Additional new controls: UnusedMisfit 0.68s, Returned 0.68s, Boxed 0.77s, Counted 0.56s, Null 0.55s, Void 0.67s, UncheckableRefusal 0.26s. All leaves below 60s. Corrected union/nullish invocation domain ABI metadata after the boxed/count-aware controls exposed its absence; corrected the null fixture to console's string contract.

Commands (all output saved, no full package sweep):
- ADAMIC_GATE_UNCACHED=1 timeout 90 go test ./internal/oracle -run '^Test(V4Escape|V4Direct|ViewAdapterBaselineCounts)' -count=1 -v -timeout 90s
- timeout 90 go test ./internal/lower ./internal/native ./internal/javascript -run '^Test(ViewCallable.*|ViewCallables.*|PrepareViewCallableRead|CallableNamespace.*|RuntimeReleasePaths|ClosureConventionRuntimeFeaturesIgnoreLiterals)$' -count=1 -v -timeout 90s

Three recorded oracle rows remain unchanged; no persistent oracle fixture added, so counts.md has no new row. Native positive and omission-mutant runs finish leak-clean under ASan/UBSan and LeakSanitizer. Runtime revision's seven mutants and byte measurement are in CACHE-REVISION.md. Setup details are in adapter-setup.log. Lane checks run after committing and their output is recorded locally beside this report.
