Built: #49yq1ah step 17 unions applied as one commit on exactly the requested main base.
Commits: source 2d9d5504a1d85dee817bd7702282e73adc31e378; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is the pushed compiler/loom-r-unions branch tip.
Commands and outputs: all 14 original focused command groups passed; 49 imported plus 11 declared reader packages compiled, other selected checks and all reader vet passed; typeaware request check hit its 60-second limit; counts regenerated once.
Mutants: all 22 recorded source mutants caught by their named assertions, plus the embedded IR field-check removal mutant caught by its baseline/unchecked-output comparison.
Not covered: completion of TestInspectRequestRefusals (60-second timeout, never rerun); exhaustive package tests/full gate, ahra loom check, WASI, general callable/object-kind unions, optional writes/delete, hidden never-array lowering.

Exact-base application: git fetch origin, checkout -b compiler/loom-r-unions dbdf977f5b, and git cherry-pick --no-commit 2d9d5504a1d85dee817bd7702282e73adc31e378. The source SHA was absent after the ordinary fetch; an explicit bounded fetch of that SHA recovered it. Only its own parent-to-commit change was applied. No comparison from main to the source tip was applied. Main was not advanced or rebased to a newer base.

Conflicts and resolutions: none. Git automatically merged internal/lower/expression.go, internal/lower/object.go and internal/oracle/counts.md. Main's library-spread infrastructure remains in object.go beside the new checked union field reads. There are no conflicted hunks. The counts auto-merge was discarded in favor of main's complete table before the single successful regeneration; no counts row was hand-merged.

One extra fixture-layout correction is required by this unit: rename scout_nullable_storage/ to scout_nullable_storage_refused/ and update nullableObjectStorageOutcome's reader. All six .a sources are byte-for-byte unchanged, all six Node/NotYet tests remain, and no flow or freshness test is edited. Conservative assumption, also stated in the commit message: keep unsupported shared-storage cases explicitly refused and use bounded representative checks for each reader package rather than an exhaustive package/full-gate run.

Runtime C lines for clearance: none. Both git diff 2d9d5504a1d85dee817bd7702282e73adc31e378^ 2d9d5504a1d85dee817bd7702282e73adc31e378 -- internal/native/runtime and the delivery diff against the pinned main base are empty. runtime-comparison.json records both empty diffs. The three added native/union.go lines are Go emission, retained from the source change. Main's runtime work is preserved.

Counts regeneration: 169.696 seconds command wall, exit 0, Linux with house clang 20.1.8. Exact argv is counts-update.json and output counts-update.log. All 1106 Go pass records (including shard parents) completed, maximum measured leaf 17.73 seconds. Regeneration measured the full generated counts family once, not just the new fixtures. Independent TestCountsAreRecordedTable then passed. There are exactly 17 changed rows: 16 newly admitted fixture rows and one existing row. The latter, taste_stage3_representations.a, changes retains/releases 11/34 to 13/36. The current-tree counterfactual disables only the new checked string-field route and restores 11/34, and scout_boxed_scalar_fields.a returns to 95/150 from 114/169. That control passed; counts-cause.log preserves the observation. The counterfactual retains main's newer library-spread code.

| Counts fixture | Before | After | Cause |
|---|---|---|---|
| internal/oracle/testdata/scout_boxed_scalar_fields.a | absent | 62/62/114/169/6/0 | new admitted fixture |
| internal/oracle/testdata/scout_boxed_scalar_stale_field.a | absent | 2/1/3/3/2/0 | new admitted fixture |
| internal/oracle/testdata/scout_class_boolean_fields.a | absent | 22/22/9/31/4/0 | new admitted fixture |
| internal/oracle/testdata/scout_hir_optional_boolean.a | absent | 1/1/0/1/1/0 | new admitted fixture |
| internal/oracle/testdata/scout_nullable_object_null_field.a | absent | 2/2/11/12/2/0 | new admitted fixture |
| internal/oracle/testdata/scout_nullable_object_stale_field.a | absent | 2/1/6/5/2/0 | new admitted fixture |
| internal/oracle/testdata/scout_nullable_object_stale_local.a | absent | 1/1/6/5/1/0 | new admitted fixture |
| internal/oracle/testdata/scout_nullable_objects.a | absent | 31/31/139/162/5/0 | new admitted fixture |
| internal/oracle/testdata/scout_nullable_string_stale_field.a | absent | 1/0/4/3/1/0 | new admitted fixture |
| internal/oracle/testdata/scout_nullable_strings.a | absent | 34/34/130/156/5/0 | new admitted fixture |
| internal/oracle/testdata/scout_selector_optional_boolean.a | absent | 2/2/1/4/2/0 | new admitted fixture |
| internal/oracle/testdata/scout_string_number_fields_locals.a | absent | 157/157/137/295/5/0 | new admitted fixture |
| internal/oracle/testdata/scout_string_number_stale_field.a | absent | 2/0/3/3/2/0 | new admitted fixture |
| internal/oracle/testdata/scout_tsc_nullish_content.a | absent | 0/0/6/6/0/0 | new admitted fixture |
| internal/oracle/testdata/scout_tsc_source_content.a | absent | 0/0/2/2/0/0 | new admitted fixture |
| internal/oracle/testdata/scout_values_optional_boolean.a | absent | 2/2/2/5/2/0 | new admitted fixture |
| internal/oracle/testdata/taste_stage3_representations.a | 24/24/11/34/10/0 | 24/24/13/36/10/0 | new checked string-field route; counterfactual verified |

Focused validation uses the exact source-report selections. run-focused.py records all argv, exits, command wall seconds and Go leaf events in focused-results.json. The fixtures run source Node, native release/ASan/UBSan/leak checks and backend Node as their harness specifies; checked stale-flow fixtures retain the inserted-panic assertions. Coverage includes nullable strings and objects, scalar field tags, optional boolean class fields, all six storage refusals, RegExp adapter, changed stage1 gap leaves, flow/freshness remainders and native ownership consumer. Guards include SortedMap IR storage, exact runtime-feature witnesses, registered native agreement rows, parallel tests and TestCallTargetReaders (-timeout 90s). No new IR map, package, runtime feature set or build product was introduced by this recut; no new compiler dependency declaration is needed. Product/test sources add no fixed build-directory or machine-path literals. No cohere implementation was copied.

| Focused command group | Exit | Command seconds | Completed Go test records |
|---|---:|---:|---:|
| lower-focus | 0 | 14.037 | 11 |
| oracle-focus | 0 | 28.964 | 7 |
| fixtures | 0 | 10.843 | 145 |
| regex-adapter | 0 | 42.078 | 131 |
| stage1-gaps | 0 | 50.745 | 4 |
| flow-consumers | 0 | 74.073 | 32 |
| fresh-consumers | 0 | 10.731 | 45 |
| native-consumers | 0 | 44.306 | 1 |
| guards-ir | 0 | 11.842 | 1 |
| guards-runtime | 0 | 14.835 | 1 |
| guards-registration | 0 | 4.959 | 1 |
| guards-parallel | 0 | 4.855 | 1 |
| readers | 0 | 19.088 | 1 |
| vet | 0 | 3.047 | 0 |

All 49 touched or direct-reader packages were discovered from go list production, test and external-test imports. consumers.json and selected-consumers.json name them and their checks. go test -c compiled every package; three are compile-only, and the remaining 46 have selected test coverage (five core packages reuse the focused checks instead of rerunning them). A further 11 compiler readers declared in cloud/fast-gate/compiler-dependencies.json, including CLI/source readers without direct compiler imports, were compiled and checked with the bounded selections in declared-readers.json. All selected checks passed except the typeaware request timeout. These are representative exact test selections, not whole-package runs. Each additional consumer test has a 60-second Go limit and a 65-second outer kill; Go compilation has its own 180-second bound. Every completed test leaf in all logs, including setup failures and mutants, is below 60 seconds; the one timed-out check is explicitly red. The maximum completed leaf is 48.81 seconds. TestInspectRequestRefusals hit its 60-second alarm (60.050 seconds process wall) and was not rerun. Test-binary compilation and command wall time are distinguished from test-leaf setup; some commands exceed 60 seconds because they compile Go before tests start. No over-budget test was rerun.

Cold setup recovery, with original failed logs retained:

- ESTree initially stopped because its Go oracle owner was unprepared. TestProduct_EstreeGoOracle in stage1/cohere/estree/oracles passed, then TestAcceptanceGrammar passed (48.81 seconds leaf elapsed). estree-retry.json records owner/consumer command times.
- The root overlay filled (32 GB; /tmp is a separate 8.8 GB filesystem). A comments check, two compile-profile checks and four package builds failed from ENOSPC; the first consumer driver then stopped while opening a new log. Prior-unit scratch trees and one immutable cached export product were relocated to /tmp, preserving their contents and the cache path via a symlink. The detached worktree metadata was repaired. space-recovery.json and cache-relocation.json record the exact moves. No source, test or runtime code was changed for storage recovery.
- Only failed or unrecorded packages resumed. Comments then exposed its separate required Go oracle owner. TestProduct_CommentsOracle passed (19.72 seconds leaf), then TestJsxParserGapIsExplicit passed. All original attempts remain in consumer-results.jsonl and *.first.log; the comments owner recovery is in comments-retry.json. Every successful final reader check is accounted for in final-results.json.
- Typeaware TestInspectRequestRefusals timed out at its 60-second alarm while preparing checker archives and native probes (60.050 seconds process wall). The log records checker preparation 13.505s, wrong-kind archive 11.93s, unknown-question archive 11.38s, plus several 3.7s clang steps. No union assertion failure preceded the timeout. This does not establish behavior on the base or prove the timing is independent of this change. It was never rerun or weakened. The separate TestTypeAwareUnitDeadline passed; this is additional bounded package coverage, not completion of the timed-out request check.
- stage3/lane TestApplyPython initially failed in 0.54s because its ordinary-apply test requires the pinned TypeScript gate checkout. Providing ADAMIC_TYPESCRIPT_SOURCE for that command with the existing verified v6.0.3 checkout (050880ce59e30b356b686bd3144efe24f875ebc8) resolved it. The retry passed without changing product/test source; lane-pinned-source-retry.json records its command seconds.

Vet passed for all 49 touched/direct-reader packages (68.598 seconds, exit 0), and separately for all 11 additional declared readers; changed-package vet also passed as the original focused group. gofmt and source whitespace checks are checked before committing. The catalog 08 patch applies cleanly with git apply --check; the full catalog runner was not run.

Every recorded source mutant and its actual catcher follows. All mutations were rebuilt as Go overlays over the current files, preserving main additions; no stale whole-file overlay from the old source was used. Patches and mutant sources are .patch/.go.txt. The first preparation stopped before executing a mutant because a one-line seam was ambiguous; three lines of unchanged context resolved it. No compiler warning, build failure or Go panic is counted as a semantic kill.

| Mutant | Catcher | Command seconds | Failure observed |
|---|---|---:|---|
| runtime-typeof | TestNativeAgreesWithNode_091/internal/oracle/testdata/scout_tsc_nullish_content.a | 14.739 | stdout differs |
| null-box | TestNativeAgreesWithNode_035/internal/oracle/testdata/scout_nullable_strings.a | 14.989 | stdout differs |
| nullable-admission | TestNullableStringAdmission | 19.660 | /tmp/adamic-gate/TestNullableStringAdmission4079398263/001/main.a:1:15: stage 0 can't lower a value of type string \| null yet |
| lower-typeof | TestNativeAgreesWithNode_091/internal/oracle/testdata/scout_tsc_nullish_content.a | 28.810 | Lower: /workspace/adamic/internal/oracle/testdata/scout_tsc_nullish_content.a:3:17: stage 0 can't lower typeof a value holding both null and undefined without a presence slot yet |
| equality-observation | TestNativeAgreesWithNode_035/internal/oracle/testdata/scout_nullable_strings.a | 41.110 | exit codes differ |
| class-zero | TestNativeAgreesWithNode_042/internal/oracle/testdata/scout_class_boolean_fields.a | 45.138 | stdout differs |
| class-optional | TestScoutClassBooleanKeepsUnsafeViewsRefused | 32.794 | neighbor must remain stopped, got <nil> |
| boxed-check | TestNativeAgreesWithNode_127/internal/oracle/testdata/scout_boxed_scalar_stale_field.a | 64.819 | stdout differs |
| boxed-observation | TestNativeAgreesWithNode_088/internal/oracle/testdata/scout_boxed_scalar_fields.a | 59.443 | exit codes differ |
| object-boundary | TestNativeAgreesWithNode_053/internal/oracle/testdata/scout_nullable_objects.a | 46.580 | stdout differs |
| object-tag | TestNativeAgreesWithNode_119/internal/oracle/testdata/scout_nullable_object_stale_field.a | 42.159 | want the inserted check to fire natively (exit 70) where the source on Node runs on: native 0, Node 0 |
| object-optional | TestNativeAgreesWithNode_053/internal/oracle/testdata/scout_nullable_objects.a | 66.320 | stdout differs |
| object-null-field | TestNativeAgreesWithNode_103/internal/oracle/testdata/scout_nullable_object_null_field.a | 63.849 | stdout differs |
| string-check | TestNativeAgreesWithNode_073/internal/oracle/testdata/scout_string_number_stale_field.a | 14.941 | want the inserted check to fire natively (exit 70) where the source on Node runs on: native 0, Node 0 |
| string-tag | TestNativeAgreesWithNode_029/internal/oracle/testdata/scout_string_number_fields_locals.a | 35.536 | exit codes differ |
| local-nullish | TestNativeAgreesWithNode_029/internal/oracle/testdata/scout_string_number_fields_locals.a | 33.747 | stdout differs |
| regex-adapter | TestNativeAgreesWithNode_122/internal/oracle/testdata/regexp_replace/callback.a | 29.564 | Lower: /workspace/adamic/internal/oracle/testdata/regexp_replace/callback.a:34:31: stage 0 can't lower sharing nullable object storage with different null/undefined tags; copy values into tagged storage yet |
| shared-view | TestNullableObjectStorageArray | 26.742 | want explicit storage view stop, got <nil> |
| lookup | TestNullableObjectStorageLookup | 34.634 | want explicit storage view stop, got <nil> |
| local-equality | TestNativeAgreesWithNode_029/internal/oracle/testdata/scout_string_number_fields_locals.a | 34.411 | exit codes differ |
| object-local-runtime | TestNativeAgreesWithNode_112/internal/oracle/testdata/scout_nullable_object_stale_local.a | 34.150 | exit codes differ |
| object-local | TestNativeAgreesWithNode_053/internal/oracle/testdata/scout_nullable_objects.a | 32.890 | the release build: exit codes differ |

The 23rd mutant is TestNullableStringFieldCheckMutant's IR field-check removal. The checked baseline panics with exit 70, while the unchecked JavaScript mutant prints undefined and exits 0. This independent behavior comparison passed in oracle-focus.log. Source-mutant catcher errors and exact leaf seconds are also in mutant-catchers.json and test-seconds.csv/json. Every completed focused, consumer, owner, count, counterfactual and mutant leaf's seconds are preserved there, including original failed setup attempts.

Setup: export GOPROXY=https://proxy.golang.org|direct; timeout 600 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Completed without failure in 85.831 seconds. nproc=5; cgroup cpu.max=400000 100000, four CPU shares. GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1 were used for validation. Timing lines:

```text
setup: node ready (0.081s)
setup: go ready (0.101s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.016s
setup: markdown dependencies ready (0.244s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.455s)
setup: submodules ready (1.939s)
setup: shared cache ready (2.948s)
setup: go build ready (85.716s)
setup: test binaries deferred (use --warm-tests) (85.799s)
setup: build cache warm (85.800s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.21 0.06 0.02 1/137 683 load-after=4.88 1.31 0.44 3/170 2186
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (85.831s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.nHwwYP
```

No PR, ahra CLI check or Loom submission was made. Only compiler/loom-r-unions is pushed, without force.
