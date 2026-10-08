# Area and merge timing attempt

Area: ad7bd06632f119abc7680719ad3a7d3b71100f58. Merge: 798d43097, before this unit’s narrow facts. Identical flags: GOMAXPROCS=4, GOFLAGS=-buildvcs=false; go test -json -count=1 -timeout=90m ./stage1/cohere/lint. All required inputs supplied, fresh profile directories, clean TypeScript 6.0.3 at 050880ce.

Area complete: 151 pass, 0 fail, 1 named pending skip; outer wall 3137.205s; nproc 5. Mean load: 4.99, 4.84, 4.55.

Merge INTERRUPTED during TestProfileArtifacts by the runtime transition. No package result or valid complete comparison. Both logs retained under /tmp/checker-facts-ad7-{area,merged}-r2.jsonl. This table is diagnostic evidence, not acceptance of the paired timing condition.

Own time includes run to pause plus cont to terminal event; excludes paused intervals. Interrupted tests show only completed intervals.

| Top-level test | Area own seconds / result | Interrupted merge own seconds / result |
| --- | --- | --- |
| TestCapturedOracleRecovery | 0.241 / pass | not run |
| TestCheckerBridgeRefusalPending | 0.000 / skip | 0.000 / running |
| TestCheckerCacheSourceByte | 128.379 / pass | 0.000 / running |
| TestCheckerHashes | 11.522 / pass | 0.000 / running |
| TestCheckerNoProgramCoverage | 3.298 / pass | 0.000 / running |
| TestCheckerReplayEntryControl | 12.485 / pass | 0.000 / running |
| TestChildCPUHangGuard | 2.357 / pass | 1.273 / pass |
| TestChildCPUWaitGuard | 1.664 / pass | 1.643 / pass |
| TestChildWallBackstop | 0.199 / pass | 0.203 / pass |
| TestCommandDiagnosticsDropOnlyModuleDownloads | 0.000 / pass | 0.000 / running |
| TestCommentFoldMutant | 55.889 / pass | not run |
| TestCompilerAndStage1Agree | 790.434 / pass | 0.000 / running |
| TestCompilerCorpusSourcesExcludeOnlyNamedFolders | 0.627 / pass | 0.000 / running |
| TestCompilerGuardBackend | 0.000 / pass | 0.000 / pass |
| TestCompleteSuggestionSerialization | 240.431 / pass | 0.000 / running |
| TestCountGuardMutant | 69.078 / pass | 0.000 / running |
| TestDecodedOptionsAndMutant | 44.585 / pass | not run |
| TestDecorationOptionMutant | 50.652 / pass | 0.000 / running |
| TestDotARename | 37.950 / pass | 0.000 / running |
| TestEmittedJavaScriptMismatch | 156.513 / pass | 0.000 / running |
| TestExecuteFailsOnStderrOtherThanModuleDownloads | 1.152 / pass | 0.000 / running |
| TestFactoryHooks | 440.799 / pass | not run |
| TestJsxInventoryDiscovery | 0.004 / pass | 0.001 / pass |
| TestJsxLintReleaseAndThroughput | 221.914 / pass | 793.938 / pass |
| TestJsxLintTrees | 617.603 / pass | 0.000 / running |
| TestLegacyMutants | 316.489 / pass | 0.000 / running |
| TestMutants | 1359.714 / pass | 0.000 / running |
| TestNestedConstructorGap | 0.440 / pass | 0.000 / running |
| TestNestedOutsideModuleCopy | 1.553 / pass | not run |
| TestNodeTableIsLinkOnly | 8.279 / pass | 0.000 / running |
| TestNonprogressingFix | 45.262 / pass | 288.495 / pass |
| TestNonprogressingFixPanicMutant | 29.040 / pass | 190.453 / pass |
| TestNonprogressingFixPlanOrder | 25.232 / pass | 173.242 / pass |
| TestOptionAndComparatorGaps | 1.240 / pass | 0.000 / running |
| TestOwnedWitnesses | 40.444 / pass | not run |
| TestPositionIndexMutant | 80.764 / pass | not run |
| TestProfileArtifacts | 78.926 / pass | 0.000 / running |
| TestProfileCompilation | 182.388 / pass | not run |
| TestProfileSnapshotsAgree | 707.190 / pass | not run |
| TestRecoveryClassificationPreservesModes | 0.000 / pass | not run |
| TestRegistrationMutant | 43.161 / pass | not run |
| TestRulesAgree | 1172.172 / pass | 0.000 / running |
| TestShardsAgree | 511.891 / pass | 0.000 / running |
| TestSuggestionAlongsideAutomaticFix | 566.526 / pass | 0.000 / running |
| TestThroughput | 168.965 / pass | 259.784 / pass |
| TestUnmarkedMalformedOracleInputStillFails | 0.020 / pass | not run |
| TestWitnessScriptKind | 115.315 / pass | 0.000 / running |

Named area skip: awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer.
