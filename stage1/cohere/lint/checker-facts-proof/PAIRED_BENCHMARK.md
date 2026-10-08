# Final paired lint package benchmark

Area: ad7bd06632f119abc7680719ad3a7d3b71100f58. Facts: frozen candidate b3ab52dd5, containing the merge of ad7bd066 and the narrow-fact implementation. The 69 implementation and fixture files match lint-checker/facts; final reporting files are outside the executable workload. Both corpus checkouts are clean and pinned.

Same box, back to back, Go 1.27.1, nproc=5, CPU quota=4, GOMAXPROCS=4, GOFLAGS=-buildvcs=false. Command: go test -json -count=1 -timeout=90m ./stage1/cohere/lint. ADAMIC_TYPESCRIPT_SOURCE is clean TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. ADAMIC_LINT_BENCH=1. Each run has a fresh directory shared by ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS. GOPROXY=https://proxy.golang.org|direct. Both Go packages were prewarmed; source-keyed native caches and source-byte canaries stay enabled. Native checks use -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all.

| Tree | Pass | Fail | Skip | Wall seconds | Mean load (1/5/15 min) |
| --- | ---: | ---: | ---: | ---: | --- |
| Area | 151 | 0 | 1 | 1329.393 | 4.60 / 4.27 / 4.38 |
| Facts | 167 | 0 | 1 | 1822.973 | 5.10 / 4.99 / 4.76 |

Facts minus area: 493.580s. The approximately 2700s seat budget is met on this box.

The sole skip on both trees is TestCheckerBridgeRefusalPending: awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer. Library commits remain absent.

## Every top-level test own time

Own time sums run-to-pause and cont-to-terminal intervals, excluding time paused. Parallel times overlap and must not be added to obtain package wall time. A dash means the test does not exist in that tree.

| Top-level test | Area own seconds | Facts own seconds |
| --- | ---: | ---: |
| TestCapturedCaughtErrorOptions | — | 0.000 |
| TestCapturedOracleRecovery | 0.191 | 0.248 |
| TestCheckerBridgeRefusalPending | 0.000 | 0.000 |
| TestCheckerCacheSourceByte | 61.950 | 78.782 |
| TestCheckerHashes | 0.697 | 1.074 |
| TestCheckerLibraryMemberReadContract | — | 127.153 |
| TestCheckerNoProgramCoverage | 2.043 | 2.381 |
| TestCheckerProgramReadsMatchCohere | — | 0.071 |
| TestCheckerProgramReadsMutant | — | 0.134 |
| TestCheckerReplayEntryControl | 6.306 | 6.778 |
| TestChildCPUHangGuard | 1.145 | 1.193 |
| TestChildCPUWaitGuard | 1.647 | 1.635 |
| TestChildWallBackstop | 0.206 | 0.204 |
| TestCommandDiagnosticsDropOnlyModuleDownloads | 0.000 | 0.000 |
| TestCommentFoldMutant | 32.506 | 44.833 |
| TestCompilerAndStage1Agree | 311.336 | 282.343 |
| TestCompilerCorpusSourcesExcludeOnlyNamedFolders | 0.091 | 0.088 |
| TestCompilerGuardBackend | 0.000 | 0.000 |
| TestCompleteSuggestionSerialization | 51.119 | 57.641 |
| TestCountGuardMutant | 21.689 | 34.205 |
| TestDecodedOptionsAndMutant | 13.366 | 37.593 |
| TestDecorationOptionMutant | 25.100 | 35.496 |
| TestDotARename | 15.958 | 19.544 |
| TestEmittedJavaScriptMismatch | 42.726 | 51.710 |
| TestExecuteFailsOnStderrOtherThanModuleDownloads | 0.521 | 0.630 |
| TestFactoryHooks | 36.235 | 67.278 |
| TestJsxInventoryDiscovery | 0.000 | 0.000 |
| TestJsxLintReleaseAndThroughput | 117.246 | 156.485 |
| TestJsxLintTrees | 214.932 | 162.335 |
| TestLegacyMutants | 25.298 | 38.378 |
| TestMutants | 550.902 | 1163.955 |
| TestNestedConstructorGap | 0.268 | 0.238 |
| TestNestedOutsideModuleCopy | 0.990 | 4.437 |
| TestNodeTableIsLinkOnly | 4.721 | 7.250 |
| TestNonprogressingFix | 29.512 | 28.504 |
| TestNonprogressingFixPanicMutant | 14.821 | 18.060 |
| TestNonprogressingFixPlanOrder | 20.617 | 21.862 |
| TestOptionAndComparatorGaps | 0.714 | 0.908 |
| TestOwnedWitnesses | 18.604 | 50.655 |
| TestPositionIndexMutant | 30.111 | 45.202 |
| TestProfileArtifacts | 49.672 | 59.133 |
| TestProfileCompilation | 77.555 | 149.108 |
| TestProfileSnapshotsAgree | 370.929 | 375.265 |
| TestRecoveryClassificationPreservesModes | 0.000 | 0.000 |
| TestRegistrationMutant | 12.847 | 18.574 |
| TestRulesAgree | 449.362 | 1147.254 |
| TestShardsAgree | 256.845 | 324.864 |
| TestSuggestionAlongsideAutomaticFix | 48.925 | 34.661 |
| TestThroughput | 119.768 | 131.338 |
| TestTypedProjectInputsAndMutants | — | 0.852 |
| TestUnmarkedMalformedOracleInputStillFails | 0.014 | 0.017 |
| TestWitnessScriptKind | 41.645 | 28.780 |

## Cost and deadline interpretation

The added typed cases retain their real fixture projects. One checker program is opened per manifest run, and the Go oracle shares its program across that manifest’s rows. Distinct captured projects cannot share a program without changing inputs. The timed per-row events include Go program plus lint, native program plus lint plus recording, Node replay and emitted-JavaScript replay; recording cannot be separated from native lint using these events alone. The table identifies the full rule-corpus and control costs. No private checker, source hash, transcript field, cache guard, corpus input or timeout was removed or relaxed.

The old deadline failure was TestCompilerAndStage1Agree: /tmp/adamic-gate/adamic-lint-native-538308627/scanner --manifest /tmp/adamic-gate/TestCompilerAndStage1Agree4088875629/002/manifest.txt. On the same frozen 881-file manifest, area completed in 598.691236s and facts in 566.115077s with all input hashes unchanged. This is evidence of a healthy child near the old 600s wall guard, not a facts-induced slowdown. The ad7 merge supplies the progress/sharding and CPU=600s, wall=3600s guards with their canaries; this unit does not raise a bound.

## Stopped attempts

Earlier pairs were interrupted by runtime transitions or stopped at first failures. The mock witness gained three calls while one project-control expectation remained at one; correcting the fixed expected count to three retained all project-input mutants. The next attempt stopped because the scratch repository corpus was dirty; committing the identical frozen candidate and initializing pinned real submodules preserved the cleanliness guard. A later complete rule-corpus run passed, but two original wrapper mutant anchors still named the replaced declaration decoder. Both anchors were retargeted to the equivalent narrow declaration-file decision; all 104 owned anchors were checked and the affected pilots passed. These stopped facts halves are not passing package results or accepted timings. Their logs remain available. Old completed/stopped build directories were cleaned to preserve storage; logs, sources, transcripts and cache guards were retained.

## Recorded typed-runtime cost

These are summed command durations from TestRulesAgree, not package wall time. Area has 263 timed typed rows; facts has 884. Source-Node replay includes its source-module startup/loading. Each Go sample combines program creation and lint; each native sample combines program creation, lint and recording. No finer separation is claimed.

| Recorded phase | Area seconds | Facts seconds | Increase seconds |
| --- | ---: | ---: | ---: |
| Go program and lint | 51.355 | 64.649 | 13.294 |
| native program, lint and recording | 64.776 | 99.455 | 34.679 |
| Node replay | 254.429 | 787.894 | 533.464 |
| emitted JavaScript replay | 51.823 | 152.714 | 100.891 |

Captured combinations increase from 5205 to 5856 (651 added). Capture now also preserves syntax-only cases without a program. Native checker builds in TestRulesAgree hit the source-keyed cache on both trees (area 71.614ms; facts 42.430ms). All 104 rule mutants bypass the persistent mutant-binary cache and are caught. The additional library-read control costs 127.153s of own time; read-declaration equality and its mutant cost 0.071s and 0.134s; captured-option validation costs less than 1ms, and project-input controls cost 0.852s. These overlap other tests. The package is 493.580s above area and 877.027s below the approximate 2700s budget.
