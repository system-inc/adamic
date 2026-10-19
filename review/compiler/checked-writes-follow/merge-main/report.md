Built step 10 merge resolution: checked-write emission with main's per-emitter object metadata cache.
Commits: merged current origin/main 0bf6186d into compiler/checked-writes-follow from cleared 29275d4e; merge commit 3d60a19969f30d01785645d6d71b0c2daef99e32; delivery SHA is in the push response.
Commands and outputs: Node-held oracle PASS 26.235s; runtime/metadata catchers PASS .948s; reader guard PASS 10.945s; counts comparison PASS 38.948s.
Mutants: all 52 prior mutants caught again, plus the new cached checked-write metadata mutant (53 total).
Limits: no whole-package test or full gate; runtime files changed since 29275d4e: none.

Main's conflicting change is c13e106c51648d5618d513b6415a51a605a3f5f1, Cache object metadata queries once per C emitter. The only content conflict was fieldTypesNeeded in internal/native/emit_objects.go. Resolution keeps prepareObjectMetadata and both cached queries, with CheckedWrites included alongside CheckedFields in the cached fieldTypes seed. It retains the checked-write contract-sensitive shape keys, allocation identities, shape contract tables and five-member adamic_shape initializers. Dynamic properties and view reads keep main's single shared expression walk. main-cache-change.patch and objects-before.patch show both sides; conflicted-objects.go.txt preserves the original conflict for review.

The inherited main TestObjectMetadataCacheIsPerEmission remains green. New top-level TestObjectMetadataCheckedWritesIsPerEmission (.00s) proves that writes alone need slot representation tags, that they do not enable reflection registration, and that a fresh emission observes removal of the write summary. Its revert mutant removes only CheckedWrites from the cached fieldTypes seed: the test fails with checked write lost its slot representation metadata. Compilation succeeds, so this is a semantic catcher. No runtime guard or checked-write contract was weakened.

The own fixtures were rerun through actual lowering, against Node, in JavaScript, release native and ASAN/UBSAN native, with leak/count assertions where applicable. These include the six runtime-made reference-field objects, the later internal writer, all five exact diagnostic messages, and the original checked-write/view/flow/emit witnesses. Native oracle observations are uncached for the explicit witnesses; source Node observations may use their validated cache. All original .a refusal paths and fixes remain pinned. 105 original runtime cases and 36 original mutant cases were rerun in oracle.log. Four message-only IR mutants are in that same run. Eleven reference/diagnostic overlay mutants and the prior literal mutant were rerun separately; mutant-results.json records successful catchers and their exact selectors. The overlay sources remain the cleared prior-turn snapshots; no runtime source changed in this merge.

Commands, all from the repository root after source /workspace/adamic-tools/env.sh, with direct log-file output:

```
export GOPROXY='https://proxy.golang.org|direct'
timeout 180 bash cloud/setup.sh
nproc
timeout 45 git fetch origin
git merge --no-commit origin/main
timeout 89 go test ./internal/native -run '^TestObjectMetadata' -count=1 -v -timeout=85s
timeout 89 go test ./internal/oracle -run 'TestCheckedWider|TestCheckedViewsNext|TestCheckedNeverContractMutant|TestCheckedDiagnosticReferenceMutants|TestCheckedFlowContainerContractMutants|TestCheckedEmit|TestCheckedWriteMessage|TestCheckedRuntime' -count=1 -v -timeout=85s
timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout=85s
timeout 89 go test ./internal/native -run 'TestRuntimeChecked|TestObjectMetadata' -count=1 -v -timeout=85s
timeout 1000 python3 review/compiler/checked-writes-follow/merge-main/run-mutants.py
timeout 89 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout=85s
```

run-mutants.py executes each listed go test with timeout 89 and Go timeout 85s, captures each output separately and records exit, real assertion failure, absence of build failure, and duration. It was monitored during execution. All 13 separate overlay runs caught their mutant; maximum wall time 16.748s. The new permanent leaf is .00s. All new oracle message/runtime leaves remain below 60s; actual leaf durations are in oracle.log.

The first cold counts comparison hit its Go hard limit at 85.030s after measuring the ordinary fixture rows; it never wrote counts.md. Repeating the comparison with those freshly measured cache rows passed in 38.948s (1009 native cache hits, seven misses). No values moved relative to the automatically merged table, so no update-counts writer was run. Compared with 29275d4e, the table has only main's three getters_census fixture rows; all previously cleared rows are unchanged. The counts overlays and preparatory shard runner were not used or committed.

Hot path: same 10,000,000 number-push source (../push-loop.a), release -O2, no sanitizers, one warmup each, alternating order, best of five. Instrument Python time.perf_counter_ns wall clock around subprocess.run; both stdout pinned to 10000000 9999999 and stderr empty. Main 0bf6186d: 68.239231 ms; resolved head: 70.508599 ms; +3.325606%. AMD EPYC 9V74 80-Core Processor, Linux 6.18.44 x86_64, nproc 5, four-CPU quota 400000/100000; load before/after [2.806640625, 2.60302734375, 1.16796875]. Every run is recorded in benchmark.json. Base compiler built from a detached worktree on origin/main, sharing the existing cohere submodule by symlink; no cohere code copied. Head compiler built from the resolved merge tree. Build commands were timeout 120 go build -o /tmp/checked-writes-merge-{head,main} ./cmd/adamic, then timeout 60 <compiler> build review/compiler/checked-writes-follow/push-loop.a -o /tmp/checked-writes-merge-push-{head,main}. All compilation logs are in this directory.

Setup timings: Go .029s, Node .029s, markdown .079s, submodules .142s, clang .276s, shared cache 1.340s, Go build 31.905s, total 32.041s; nproc 5. setup.log contains the timing and tool version lines.

Every original mutant (PASS means caught):

- TestCheckedNeverContractMutant: caught (0.88s).
- TestCheckedViewsNextMutants/branch-bottom-number-misfit: caught (1.16s).
- TestCheckedViewsNextMutants/optional-boolean-misfit: caught (1.25s).
- TestCheckedViewsNextMutants/overload-callback-misfit: caught (1.06s).
- TestCheckedViewsNextMutants/generic-site-misfit: caught (0.82s).
- TestCheckedViewsNextMutants/overload-result-misfit: caught (0.89s).
- TestCheckedViewsNextMutants/branch-bottom-string-misfit: caught (0.77s).
- TestCheckedViewsNextMutants/branch-file-misfit: caught (0.84s).
- TestCheckedViewsNextMutants/branch-length-misfit: caught (0.89s).
- TestCheckedViewsNextMutants/required-boolean-misfit: caught (0.83s).
- TestCheckedViewsNextMutants/branch-start-misfit: caught (0.98s).
- TestCheckedViewsNextMutants/optional-false-misfit: caught (0.83s).
- TestCheckedEmitContractsMutants/emit-drop-check: caught (0.82s).
- TestCheckedEmitContractsMutants/emit-callback-misfit: caught (0.84s).
- TestCheckedEmitContractsMutants/emit-auto-misfit: caught (0.98s).
- TestCheckedEmitContractsMutants/emit-resolution-misfit: caught (0.90s).
- TestCheckedEmitContractsMutants/emit-node-misfit: caught (0.78s).
- TestCheckedEmitContractsMutants/emit-comment-misfit: caught (0.85s).
- TestCheckedFlowContainerContractMutants/flow-node-misfit: caught (0.99s).
- TestCheckedFlowContainerContractMutants/container-alias-misfit: caught (0.88s).
- TestCheckedFlowContainerContractMutants/container-object-misfit: caught (1.19s).
- TestCheckedFlowContainerContractMutants/container-fill-misfit: caught (1.02s).
- TestCheckedFlowContainerContractMutants/container-boolean-misfit: caught (0.93s).
- TestCheckedFlowContainerContractMutants/container-map-object-misfit: caught (1.29s).
- TestCheckedFlowContainerContractMutants/container-nested-misfit: caught (1.19s).
- TestCheckedFlowContainerContractMutants/container-string-misfit: caught (1.03s).
- TestCheckedFlowContainerContractMutants/container-splice-misfit: caught (1.22s).
- TestCheckedFlowContainerContractMutants/container-map-misfit: caught (1.16s).
- TestCheckedFlowContainerContractMutants/flow-array-misfit: caught (0.86s).
- TestCheckedFlowContainerContractMutants/container-number-misfit: caught (0.90s).
- TestCheckedFlowContainerContractMutants/flow-undefined-misfit: caught (0.94s).
- TestCheckedDiagnosticReferenceMutants/drop_alias_check: caught (0.86s).
- TestCheckedDiagnosticReferenceMutants/drop_nested_contract: caught (0.95s).
- TestCheckedDiagnosticReferenceMutants/drop_allocation_proof: caught (0.93s).
- TestCheckedWiderWriteMutants/drop_check: caught (0.90s).
- TestCheckedWiderWriteMutants/drop_literal_set: caught (0.80s).

Four prior message-only mutants: TestCheckedWriteMessageMutant03, TestCheckedWriteMessageMutant13, TestCheckedWriteMessageMutant15 and TestCheckedWriteMessageMutant16; each caught by exact stderr in JavaScript, release native and sanitized native.

Separate overlays (first twelve are the remaining prior mutants; last is the new merge guard):

- revert-native-reference-fields: caught by ^TestCheckedRuntime (12.101s); revert-native-reference-fields.log.
- revert-js-reference-fields: caught by ^TestCheckedRuntime (10.19s); revert-js-reference-fields.log.
- revert-array-field-schema: caught by ^TestCheckedRuntime(Regex|Directory)Fields$ (9.891s); revert-array-field-schema.log.
- drop-runtime-array-contract: caught by TestRuntimeCheckedReferenceArrayKeepsContract (16.198s); drop-runtime-array-contract.log.
- drop-string-heap-kind: caught by TestRuntimeCheckedReferenceArrayKeepsContract (16.195s); drop-string-heap-kind.log.
- accept-runtime-array-literal: caught by TestRuntimeCheckedReferenceArrayRejectsLiteral (16.748s); accept-runtime-array-literal.log.
- accept-weaker-array-contract: caught by TestRuntimeCheckedReferenceArrayRejectsWeakerContract (13.627s); accept-weaker-array-contract.log.
- restore-number-view-native: caught by ^TestCheckedWriteMessage02$ (9.912s); restore-number-view-native.log.
- restore-number-view-js: caught by ^TestCheckedWriteMessage02$ (11.116s); restore-number-view-js.log.
- drop-js-runtime-array-contract: caught by ^TestCheckedRuntimeReferenceArrayLaterWriter$ (8.607s); drop-js-runtime-array-contract.log.
- drop-js-string-heap-kind: caught by ^TestCheckedRuntimeReferenceArrayLaterWriter$ (8.251s); drop-js-string-heap-kind.log.
- prior-literal: caught by ^TestRuntimeCheckedLibraryPushRejectsLiteral$ (9.122s); prior-literal.log.
- drop-checked-write-metadata: caught by ^TestObjectMetadataCheckedWritesIsPerEmission$ (14.89s); drop-checked-write-metadata.log.

Runtime clearance: git diff 29275d4e -- internal/native/runtime is empty, including main changes brought by this merge. No new runtime change needs clearance.

Lane checks after merge commit 3d60a199: PASS in 9.0s, gofmt and tools on 46 Go files, t.Parallel on five test packages, a-check on 16 .a files, vet on five packages. Command: timeout 30 git fetch -q origin main devtools/fast-gate cloud/merge-tree, then timeout 300 bash -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'. Output is lane-checks.log. Source changes pass git diff origin/main --check for emit_objects.go and object_metadata_test.go; context whitespace in preserved evidence patches from main and the original conflict is intentional.
