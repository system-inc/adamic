Built step 10 follow-up: contract-free writer guards, one receiver hold, metadata documentation and runtime library object proof.
Commits: implementation 7abb956a, based only on cleared 781766ed; evidence commit delivery SHA is supplied by the push report.
Commands and outputs: checked witnesses PASS 31.011s; selected runtime tests PASS 23.862s; reader guard PASS 7.910s; lane check recorded separately.
Mutants: all 36 existing mutants caught; new broad-slot-to-literal mutant fails the required exit-70 witness.
Limits: push loop remains above main; runtime structural fallback covers broad number/boolean fields only; no full gate or whole-package run.

The cleared branch was not modified. This branch starts at 781766edccfd21ed4f2e9f9106dd488d28751e96. Benchmark main was pinned at 5e33a17b186a8a2218d27b69b21e2de5acc5b750.

Runtime files changed: internal/native/runtime/adamic.h, array.c, map.c, object.c, regexp.c. Array guards retain the independent never_elements check; NULL alone must not disable never[]. Header documentation and contract functions now use the surrounding tab indentation. The native ArrayPush emitter reuses an existing statement-owned receiver and retains a borrowed receiver before argument evaluation.

Existing witnesses did not cover a library push of a runtime-created object with NULL shape contracts into a contracted reference array. The new witness uses actual regex matchAll iterator results; adamic_array_append calls adamic_array_push with kind/source_type zero. Original runtime fails with exit 70. The fixed runtime tags iterator-result storage and proves broad scalar structural fields without inspecting their current values. Literal, branded, reference and getter domains still fail closed without allocation proof. The test compares this native runtime operation with Node on testdata/checked_library_push.a, in release and ASAN/UBSAN, with allocation counting and live==0. This is a runtime C probe, not an end-to-end compiler lowering test of regex result arrays. The .a fixture also passes the CLI type checker. Optional done mirrors TypeScript IteratorYieldResult.

New leaves: TestRuntimeCheckedLibraryPush 17.89s; TestRuntimeCheckedLibraryPushRejectsLiteral 0.44s. The latter requires exit 70 even when the first runtime boolean payload is false. The overlay mutant allowing a broad boolean slot to satisfy a literal field exits 0 and prints match,match,done; the test fails as required (literal-mutant-final.log). Mutant source is .c.txt, not compilable Go.

No oracle fixture registration or counted allocation behavior changed; counts.md was not regenerated. The new runtime probe asserts all allocations released.

Benchmark: 10,000,000 number pushes, default release -O2 without sanitizers, one warmup per binary, rotated base/cleared/follow order, best of five. Instrument: Python time.perf_counter_ns wall clock around subprocess.run. Machine: AMD EPYC 9V74 80-Core Processor; Linux x86_64; nproc 5; cgroup quota 400000/100000 (4 CPUs).
Load averages before [2.37841796875, 1.279296875, 1.60693359375], after [2.37841796875, 1.279296875, 1.60693359375]. Main 0.064524670s; cleared before 0.111192750s (+72.33%); follow after 0.067841525s (+5.14%). Requested NULL guards were added after the initial >2% result. Guards alone did not recover the cost (benchmark-after.json); the duplicate receiver retain accounted for most of it. Residual slowdown is measured, not claimed eliminated. All five runs are in benchmark-final.json.

Commands (all test output captured directly to logs):

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
 timeout 89 go test ./internal/oracle -run 'TestCheckedWider|TestCheckedViewsNext|TestCheckedNeverContractMutant|TestCheckedDiagnosticReferenceMutants|TestCheckedFlowContainerContractMutants|TestCheckedEmit' -count=1 -v -timeout=85s
 timeout 89 go test ./internal/native -run 'TestRuntime|TestTypedArrayRuntime|TestNodeBufferRuntime|TestRegExpIteratorResultShape|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks' -count=1 -v -timeout=85s
 timeout 89 go test -overlay review/compiler/checked-writes-follow/literal-mutant-overlay.json ./internal/native -run '^TestRuntimeCheckedLibraryPushRejectsLiteral$' -count=1 -v -timeout=85s
 timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout=85s
 timeout 30 /tmp/checked-writes-follow-head types internal/native/testdata/checked_library_push.a
```

Setup timings: node .025s, Go .027s, submodules .057s, markdown .075s, clang .151s, shared cache .815s, Go build 25.259s, total 25.391s. Full setup output is setup.log.

Every existing mutant (PASS means the harness caught it):

- TestCheckedNeverContractMutant: caught (7.09s).
- TestCheckedFlowContainerContractMutants/container-map-misfit: caught (0.66s).
- TestCheckedFlowContainerContractMutants/flow-node-misfit: caught (0.95s).
- TestCheckedFlowContainerContractMutants/container-alias-misfit: caught (0.97s).
- TestCheckedFlowContainerContractMutants/container-map-object-misfit: caught (0.87s).
- TestCheckedFlowContainerContractMutants/container-fill-misfit: caught (0.86s).
- TestCheckedFlowContainerContractMutants/container-boolean-misfit: caught (0.82s).
- TestCheckedFlowContainerContractMutants/container-object-misfit: caught (0.84s).
- TestCheckedFlowContainerContractMutants/container-nested-misfit: caught (0.93s).
- TestCheckedFlowContainerContractMutants/container-splice-misfit: caught (0.97s).
- TestCheckedFlowContainerContractMutants/flow-undefined-misfit: caught (0.97s).
- TestCheckedFlowContainerContractMutants/container-string-misfit: caught (1.03s).
- TestCheckedFlowContainerContractMutants/flow-array-misfit: caught (0.98s).
- TestCheckedFlowContainerContractMutants/container-number-misfit: caught (1.07s).
- TestCheckedDiagnosticReferenceMutants/drop_alias_check: caught (0.99s).
- TestCheckedDiagnosticReferenceMutants/drop_nested_contract: caught (0.94s).
- TestCheckedDiagnosticReferenceMutants/drop_allocation_proof: caught (1.11s).
- TestCheckedViewsNextMutants/optional-boolean-misfit: caught (1.20s).
- TestCheckedViewsNextMutants/overload-callback-misfit: caught (1.16s).
- TestCheckedViewsNextMutants/overload-result-misfit: caught (1.25s).
- TestCheckedViewsNextMutants/generic-site-misfit: caught (1.26s).
- TestCheckedViewsNextMutants/branch-bottom-number-misfit: caught (1.07s).
- TestCheckedViewsNextMutants/branch-bottom-string-misfit: caught (1.16s).
- TestCheckedViewsNextMutants/branch-length-misfit: caught (1.38s).
- TestCheckedViewsNextMutants/branch-start-misfit: caught (1.23s).
- TestCheckedViewsNextMutants/branch-file-misfit: caught (1.12s).
- TestCheckedViewsNextMutants/required-boolean-misfit: caught (1.15s).
- TestCheckedViewsNextMutants/optional-false-misfit: caught (1.34s).
- TestCheckedWiderWriteMutants/drop_check: caught (1.27s).
- TestCheckedWiderWriteMutants/drop_literal_set: caught (1.16s).
- TestCheckedEmitContractsMutants/emit-drop-check: caught (0.55s).
- TestCheckedEmitContractsMutants/emit-callback-misfit: caught (0.58s).
- TestCheckedEmitContractsMutants/emit-auto-misfit: caught (0.50s).
- TestCheckedEmitContractsMutants/emit-node-misfit: caught (0.49s).
- TestCheckedEmitContractsMutants/emit-resolution-misfit: caught (0.48s).
- TestCheckedEmitContractsMutants/emit-comment-misfit: caught (0.44s).

Lane checks PASS: gofmt and tools on 42 Go files, t.Parallel on 5 test packages, a-check on 16 .a files, vet on 5 packages, 7.7s. First invocation lacked the toolchain environment and could not find gofmt; sourcing the setup environment resolved it. Required command: timeout 300 bash -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -', following git fetch -q origin main devtools/fast-gate cloud/merge-tree.
