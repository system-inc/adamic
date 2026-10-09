# Defense of class repairs and definite-assignment readiness

Starting origin/main: 7b9d4272c28f59530ab13daa5c49067e47933b06. Audit base: 8171b3173bdbfce1f7982d3c4f731279307ece37.
Two requested rows defended; private repair not defended after three attempts.
No test, oracle or harness was changed. Production sources are restored.

## Code under test and oracle

Code under test: internal/lower Lower and its class storage, static/instance method dispatch, Object.keys classification, and readiness propagation. The class-repair oracles now compare source execution in Node against generated JavaScript execution in Node (stdout, exit, relevant stderr). These assertions have changed since the audit's absence-of-error descriptions. DefiniteAssignmentUsesReadiness uses handwritten counts of nonempty ir.Read/ir.Property readiness tags, plus Node agreement for its initialized controls. The count oracle does not identify which read carries a tag or check its text. None of the three rows is an executor twin or a cost row.

## Baseline and scope

npm ci --prefix stage3/api completed before baseline. Warm /workspace/adamic-tools/env.sh worked; setup skipped. nproc=5. Clean whole-package baseline passed in 38.695 test-binary seconds. All three assigned names and subsumers remain present. Current discovery has 276 tests, audit 239. Added and vanished names are in inventory.json. Every mutant's initial matrix used the full current package.

Baseline skips: TestOriginalCycleLedger needs pristine pinned upstream TypeScript and generated diagnostics; TestOptionalWideningCensus needs an external project's config and output path. Their bodies inventory import-cycle/type-relation rules rather than call Lower or the mutated functions. The interface_Node/readonly_Node[] subcase of TestMixedUnionContractGraph remains explicitly skipped awaiting support. No skipped case is asserted to have passed. D1 and D3 uniqueness is established among the completed available rows of the current package; no repository-wide claim is made.

## Coverage and semantic leads

For each requested row and its named subsumer, ran: timeout 120 go test -count=1 -timeout 90s -run '^NAME$' -coverpkg=./internal/lower -coverprofile=PATH ./internal/lower/ > PATH-coverage.log 2>&1. All six profiles passed. These instrument Go lowering only, not executed JavaScript or C. Exact exclusive coverage blocks are saved beside profiles.

KeysRepair has 262 covered blocks absent from PrivateStorage. Its semantic distinction is an explicit fresh string-key copy of a structurally viewed iterable, alongside a nominal iterable class. D3 disables the fresh-copy exemption; only KeysRepair fails, on the symbol-key-view refusal. PrivateRepair has 367 blocks absent from SoundNeighbors, including staticInstance and callOrMethod. It delegates a static method's instance parameter to an instance method reading private storage. D2 damages receiver instantiation but virtual dispatch still selects the right method in this input; the entire package passes. D4 removes the static receiver parameter and fails PrivateRepair plus PrivateAndPublicStaticsAgreeWithNode. D5 misindexes static dispatch and panics in both rows when run individually. DefiniteAssignmentUsesReadiness has 151 blocks absent from ReadinessElisionRequiresDominatingAssignment. Its last subcase writes a field, calls a function that rebinds the box, then reads the field. D1 suppresses call invalidation; only this row loses its required guard. The current named subsumer exercises eager TypeScript non-null assertions rather than this mutable field readiness history.

## Mutant observations

D1: internal/lower/readiness.go:376: Change call detection result from true to false; calls no longer invalidate field readiness facts. Failed: TestDefiniteAssignmentUsesReadiness. Wall: 45.909s. Standalone diff: D1.diff; go vet log: D1-vet.log.

D2: internal/lower/class.go:364: Change the external-receiver instantiation condition from OR to AND. Failed: none. Wall: 46.621s. Standalone diff: D2.diff; go vet log: D2-vet.log.

D3: internal/lower/class_features.go:83: Change fresh-object classification to false, disabling the explicit-copy exemption. Failed: TestClassWrongOutputKeysRepair. Wall: 46.571s. Standalone diff: D3.diff; go vet log: D3-vet.log.

D4: internal/lower/class_static.go:202: Change the static-method receiver parameter from thisLocal(function) to -1. Failed: TestClassWrongOutputPrivateRepair, TestPrivateAndPublicStaticsAgreeWithNode. Wall: 46.375s. Standalone diff: D4.diff; go vet log: D4-vet.log.

D5: internal/lower/class_static.go:177: Off-by-one the static dispatch table function index: function to function - 1. Failed: TestClassWrongOutputPrivateRepair, TestPrivateAndPublicStaticsAgreeWithNode. Wall: 20.087s. Standalone diff: D5.diff; go vet log: D5-vet.log.

All five standalone diffs passed go vet ./internal/lower/ and were generated from the starting source. D1, D2, D3 and D4 completed whole-package matrices. D5 aborted on a Go index-out-of-range panic, so its unfinished rows remain unknown in matrix.json. Reran all three requested rows and PrivateAndPublicStaticsAgreeWithNode alone under D5. Both class-repair neighbors and the readiness row were directly observed; the private row and the additional static-method row panic. No timeout occurred. Full passed lists for D1 and D3 are in matrix.json.

## Issues, limits, and owner finding

The requested 15 GB free-space threshold cannot be met on /tmp, whose total capacity is 8.8 GB. Initial /workspace free space was 4.0 GB. Deleted the earlier completed unit's scratch/cache directory under /tmp, never repository or tools. Rechecked disk and observed ample room for this unit's small artifacts; no disk failure occurred. The two mounts are distinct, so deleting /tmp cannot increase workspace capacity.

The audit's rows.json is a list of names; verdict objects live in results.json and REPORT.md. Assuming rows.json held objects caused two harmless parsing retries. Audit assertions are stale relative to this main: current class repairs execute Node agreement, and readiness gained a call-rebinding subcase. Default shell cwd is /workspace, so one read command needed its explicit repo workdir. A read-only network check failed through the restricted proxy; the authorized fetch succeeded with escalation. Existing defense branch contains a different pair of rows, so this session is nested and preserved alongside that prior evidence without force-pushing.

PrivateRepair is not defended by these three attempts: one survives and two are caught by another current row. Its name promises a working instance-delegation repair and its assertions check that generated JavaScript agrees with source Node on the printed secret. No name/assertion mismatch found. It does not exercise native execution, inheritance, borrowed receivers, or additional secrets; no deletion recommendation follows from this bounded attempt set. D2's survival is not claimed to expose unguarded output behavior. No oracle weakening, native-runtime mutation, additional package replay, or repository-wide uniqueness was attempted.
