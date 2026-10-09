# Hidden boundaries deferred evidence

Evidence for #wvhe1jz, roadmap step 30. No compiler or test source changes.

Observed results: recut {'pass': 5, 'missing': 9}; main {'pass': 5, 'missing': 9}.

Tested recut `64a5179f933c3478f6eecbf91f85f32de1112f89`, fetched through `cloud/land-stack-hidden-boundaries-64a5179f`. The short SHA fetch returned "couldn't find remote ref". The fallback `36e7fac4` was not used.

Control: `e77a4ae41f473c149aee910c51b73637686a804a` (`origin/main` as fetched before testing).

The exact 14-name list came from `fast.json` on `gate-logs/64a5179f933c/20261009T150114Z/fast-phases`, fetched as `origin/hb-gate-record`. No dry-run planner was used. The record identifies its actual gated SHA as `3657477018d569110dfcd87f2339dabcf0c8654a`; this unit tests the requested recut SHA. Selected record fields are preserved in [gate-record.json](gate-record.json).

Each command ran alone, sequentially, on the recut then main:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_GATE_UNCACHED=1
go test ./<package> -run '^<Name>$' -count=1 -timeout 90s -json > <lane>-<Name>.log 2>&1
```

Seconds below are command wall seconds, including build/link startup. A pass requires a JSON pass verdict for the exact test. Missing means no matching test ran, even though the command exited 0. Every command also had a 180-second outer process-group limit for build startup; [runs.json](runs.json) records arguments, exit codes, timings and revisions.

| Test | Package | Branch result | Seconds | Main result | Seconds |
|---|---|---|---:|---|---:|
| TestCaseMappingMatchesNode | internal/native | [pass](branch-TestCaseMappingMatchesNode.log) | 28.349 | [pass](main-TestCaseMappingMatchesNode.log) | 26.766 |
| TestMapHashProbeCatchesMutants | internal/native | [pass](branch-TestMapHashProbeCatchesMutants.log) | 22.360 | [pass](main-TestMapHashProbeCatchesMutants.log) | 21.691 |
| TestNormalizeMatchesNode | internal/native | [missing](branch-TestNormalizeMatchesNode.log) | 2.075 | [missing](main-TestNormalizeMatchesNode.log) | 1.921 |
| TestNormalizeRandomMatchesNode | internal/native | [pass](branch-TestNormalizeRandomMatchesNode.log) | 12.567 | [pass](main-TestNormalizeRandomMatchesNode.log) | 12.767 |
| TestRecordMutants | internal/native | [missing](branch-TestRecordMutants.log) | 1.977 | [missing](main-TestRecordMutants.log) | 2.074 |
| TestRecordReadMutants | internal/native | [pass](branch-TestRecordReadMutants.log) | 22.257 | [pass](main-TestRecordReadMutants.log) | 3.734 |
| TestRegExpBytecodeRandomNode | internal/native | [missing](branch-TestRegExpBytecodeRandomNode.log) | 1.872 | [missing](main-TestRegExpBytecodeRandomNode.log) | 2.078 |
| TestRegExpBytecodeTest262 | internal/native | [pass](branch-TestRegExpBytecodeTest262.log) | 15.921 | [pass](main-TestRegExpBytecodeTest262.log) | 17.059 |
| TestSplitTSGoAgrees | internal/native | [missing](branch-TestSplitTSGoAgrees.log) | 1.977 | [missing](main-TestSplitTSGoAgrees.log) | 1.971 |
| TestWASI | internal/native | [missing](branch-TestWASI.log) | 1.923 | [missing](main-TestWASI.log) | 1.924 |
| TestCompilerAndStage1Agree | stage1/cohere/lint | [missing](branch-TestCompilerAndStage1Agree.log) | 7.437 | [missing](main-TestCompilerAndStage1Agree.log) | 7.201 |
| TestJsxLintTrees | stage1/cohere/lint | [missing](branch-TestJsxLintTrees.log) | 2.071 | [missing](main-TestJsxLintTrees.log) | 2.071 |
| TestNodeTableIsLinkOnly | stage1/cohere/lint | [missing](branch-TestNodeTableIsLinkOnly.log) | 2.121 | [missing](main-TestNodeTableIsLinkOnly.log) | 2.124 |
| TestShardsAgree | stage1/cohere/lint | [missing](branch-TestShardsAgree.log) | 2.226 | [missing](main-TestShardsAgree.log) | 2.021 |

Missing names are absent as exact top-level tests in both tested trees. Several native names now use Unit or Points/Contexts suffixes; lint uses product/leaf tests. Those successors were outside this exact-name brief and were not substituted. This evidence cannot turn absent tests into green coverage.

Mutants exercised by the requested tests: `TestMapHashProbeCatchesMutants` checks old hash against the integer probe bound, zero normalization against "zero hashes differ", and NaN normalization against "NaN hashes differ". `TestRecordReadMutants` checks prototype-membership-restored and missing-read-silent against the exact diagnostic/exit contract, and own-read-checked-as-missing against the own-hit fixture. Their individual outcomes and catcher messages are in the logs. No new checks or mutants were added; absent `TestRecordMutants` ran none of its successor mutants.

Setup succeeded. Its complete output and timing lines are in [setup.log](setup.log); `nproc=5`, cgroup quota 4 CPUs. No new fixture or test leaf was added, so counts were not refreshed. The whole gate and whole packages were not run. No _test.go file was edited, so the call-target-reader guard was not required.

Lane check output is in [lane-checks.log](lane-checks.log). The required default invocation checks inherited recut changes against the merge base with current main as well as this evidence commit; any inherited findings are reported without fixing source in this unit.
