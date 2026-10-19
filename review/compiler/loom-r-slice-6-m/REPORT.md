Re-cut slice 6, keeping exceptions, inheritance guards, feature-set linking and lint features.
Branch compiler/loom-r-slice-6-m; one commit on a22f4251067a5f53a00d07fc64b54a2ac72d340b; delivery SHA is in the final response.
Linux clang 20.1.8 counts regeneration passed; native, guards and vet evidence is recorded here.
Native package mutant tests reran; the separate prior 61-attempt campaign was not repeated and remains in review/compiler/loom-r-slice-6/REPORT.md.
Stage1 suites were not rerun, as requested; this re-cut does not resolve their prior memory failures.

## Method and conflict

Fetched origin, based on a22f4251067a5f53a00d07fc64b54a2ac72d340b. Used exactly `git merge-tree --write-tree --merge-base=3016b8ef~1 origin/main 3016b8ef`, materialized its tree and resolved its one conflict.

`internal/native/regexp_test.go`, regexWitnessRuntimeSources: main added `host_runtime.c` for host runtime transitive symbols. Slice 6 added `features.c` for feature-mask link symbols. Keep both entries in the minimal sanitized regex witness archive. No rule was loosened and no test removed.

Auto-merges checked: main's `eepPropertyPresence` call stays before property lowering; main's PlantedWrongExit mutation targets `return adamic_host_exit_status();`; adamic.h retains main's host_runtime.h include; storage.txt retains all host statics and runtime-file registrations alongside slice 6's TLS/immutable entries. Assumption: these newer main behaviors remain authoritative; no main line was intentionally reverted.

## Counts

Regenerated on Linux with clang 20.1.8 using `timeout 180 go test -p 1 ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 150s -json -args -update-counts`: PASS, 90.993 seconds. No hand merging. Main's counts table equals the prior source base's table; regeneration produced a byte-identical table to 3016b8ef. Zero additional moved rows. The same 147 affected rows relative to current main (26 added, 9 removed, 112 changed) are listed exhaustively in counts-rows.json, each before/after and owner. inherit-guards owns inherited_fields_guard.a; exceptions-21-main owns the others. Feature-set-link and lint-features add no count row.

## Runtime clearance comparison

The runtime C/header added and removed lines in this re-cut relative to current main are EXACTLY IDENTICAL to 3016b8ef's added and removed lines relative to its parent. No additional, removed or altered clearance line. runtime-comparison.json records the comparison, including filenames and every changed line; runtime-diff.patch lists the complete current diff with exact added/deleted lines. Main's host_runtime.c/.h and host_runtime.h include remain present; those are main's lines, not new slice 6 runtime edits. The prior runtime conditions remain in force, including separate thread-local payload/pending flag, pending-aware parallel propagation under existing locks, exact uncaught rendering and immutable name literals.

## Setup and checks

`export GOPROXY='https://proxy.golang.org|direct'`; `timeout 180 bash cloud/setup.sh` passed. Sourced /workspace/adamic-tools/env.sh and the retained task-specific cache/WASI environment. Go 1.27.1, Node24.19.0, clang20.1.8. nproc=5; cgroup quota 400000/100000, GOMAXPROCS=4 and test parallelism4.
Setup timing lines: go0.022s, node0.026s, submodules0.069s, markdown0.082s, clang0.181s, shared-cache-off0.183s, build46.261s, binaries deferred46.444s, cache warm46.445s, done46.483s. Full setup.log retained.

Changed-package gofmt -l produced no output; production/test diff whitespace check passed. Vet passed on all 16 changed normal Go packages and the build-tagged scanner build-metrics.go. The first explicit scanner vet used the wrong filename main.go; corrected command passed, and both logs remain.

Detailed commands are in commands.txt, raw test output in *.jsonl; all invocations have a hard bound. No a-check, broad lower/oracle/flow/fresh/JS or stage1 rerun was requested for this re-cut; prior results and limitations remain documented in the source report. No ahra or Loom submission.

## Final validation outcomes and limits

- Native full tree command completed: internal/native 327.399s, with 814 passing test entries, 95 explicit skips and one failure, TestLongUnicodePortabilityGCC (0.80s). internal/native/runtimeproducts passed in 8.685s. No observed native leaf exceeded 60s; maximum leaf 17.89s, TestAsyncHostPromisesAddress. TestSplitTSGoAgreesUnit00, including setup, passed35.62s; its cache_0 leaf14.13s. Every completed leaf and native mutant test outcome is in native-summary.json and raw native.jsonl. Opt-in/environment skips are not claimed run.
- The sole native failure was GCC14's missing sanitizer/allocator_interface.h in main's new host_runtime.c. A CPATH replay failed because main deliberately sanitizes that variable. No sudo is installed and the GCC include directory is root-owned. A task-local GCC14 profile adds only an explicit -isystem directory containing the house LLVM20.1.8 sanitizer headers and reports that profile in --version, distinguishing its cache identity. The same unchanged test passed4.74s, matching Node with ASan, UBSan and leak detection. gcc-profile.sh records the wrapper, gcc-profile-replay.jsonl records the pass. The stock GCC invocation on this box remains unprovisioned; the original full-tree exit1 is retained, not claimed a clean stock-toolchain run. No C/header source was changed for this workaround.
- All structural guards passed: TestEveryTestHasOneDescriptiveName0.33s; TestEveryTestIsParallelOrSaysWhy0.32s; TestEveryMapAProgramHoldsIsSorted0s; TestNativeAgreementRowsNameRegisteredPrograms0.01s; TestRuntimeSetsAreEachUsed0.93s; TestRuntimeStaticsAreListed0.13s. The first anchored reader selector matched no split names; corrected selectors ran all four split reader leaves and their setup, all passed (readers-Native/Lower/Fresh.jsonl and readers-Flow-replay.jsonl).
- Flow's first export setup/reader failed writing a copied archive for lack of space, before checking the reader. Replay moved only the product output volume to the larger task scratch volume and passed1.779s, reader0.05s. Completed export archives were byte-compared and hardlinked to identical Go cache bodies to reclaim copies; no archive bytes were changed or Go cache entry removed. cache-dedup.log retains evidence.
- Counts passed90.993s with no new rows versus the prior cut; formatting, whitespace and vet passed. No stage1 suite was rerun or repaired. Separate source mutant campaign, a-check and broad non-native reader package suites were not repeated for this one-file conflict resolution; their original evidence and limits remain in the source report.
