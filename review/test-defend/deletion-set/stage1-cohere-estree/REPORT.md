# ESTree deletion-set replay

Starting commit: `7b9d4272c28f59530ab13daa5c49067e47933b06`. All nine gathered standalone diffs applied unchanged.

The clean baseline with all three candidates skipped passed in 1093.424 wall seconds (1087.326 binary seconds). Completed mutant replays used 5559.058 wall seconds. All commands use ADAMIC_GATE_UNCACHED=1 and a fresh ADAMIC_BUILD_CACHE_DIR, with a 30 minute test timeout.

Mutant runs with ordinary failures were stopped after the first clean non-witness failure and the buffered events were drained. Their matrix statuses list only observed rows; later rows are unknown. Runs with no ordinary catcher completed the entire default package gate. Witness failures and panicking tests are reported separately and never credited as sole catchers.

The two interface candidates jointly protect the lost catches. At least one of that pair must remain; this replay does not establish that both are individually necessary. The unattached decorator control is deletable only in the narrow sense that every gathered mutant it failed has an observed ordinary catcher outside the set. Nothing was deleted or weakened.

Twelve optional rows skipped in the baseline: TestCookedSurrogateLibraryGap, TestCorpusNativeRefusals, TestDecoratedExportLibraries, TestJSXOriginalLibraries, TestOriginalLibraries, TestPinnedNumericGaps, TestRecoveryLibraryGaps, TestRepositoryAgreement, TestScalarOriginalLibraries, TestSyntaxLibraries, TestThroughput, TestTypeMemberLibraryGap. Conclusions cover the default gate, including newly discovered tests, rather than unavailable optional corpora.

Costs and interruptions: historical branches reused mutant names in root and session matrices, so each diff was paired with its adjacent matrix and given a branch-qualified filename. An initial partial M06 replay was abandoned to fix panic recording; it has no verdict. The workspace restarted during gaps D3; that partial log has no verdict, and remaining replays restarted with fresh replay-v3 caches. /tmp has a total capacity below the requested 15 GB free, despite removing prior-unit scratch. No baseline or completed replay failed from disk exhaustion.

See mutant-list.json for the catalog written before replay, matrix.json for commands and all observed statuses, report.json for the decision, and the adjacent raw JSON logs. Production diffs were restored after each run. Only this evidence directory is committed and pushed.

D6 exhausted the 30 minute test timeout in TestAcceptanceGrammar, leaving later rows unasked. Its native subprocess was verified and terminated only after it became an orphan. A fresh-cache rerun skipped TestAcceptanceGrammar and failed cleanly in TestAcceptanceDiagnostics at acceptance_test.go:46: child CPU deadline exceeded (2 second budget). This clean rerun is the retained catch; the aborted run has no catcher credit.
