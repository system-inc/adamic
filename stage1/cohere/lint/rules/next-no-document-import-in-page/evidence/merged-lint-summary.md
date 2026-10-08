# Merged lint certification

Upstream: cd56db1d (contains f0ccab34). Merge: 49e12ac2.

Full lint package: PASS, 1800.424s. Focused nine-helper package: PASS, 48.876s.

Helpers: 572 unique upstream cases, 1,274 consuming helper calls, 1,832 total calls; Go/Node/emitted JavaScript/ASan-UBSan native agree. Nine helper mutants caught on all three backends.

Rule: @next/next/no-document-import-in-page; 34 unique upstream cases. Its module-equality mutant caught on Node, emitted JavaScript, and sanitized native.

Merged corpus: 4,610 unique upstream source/rule/options combinations. Compiler/stage1: 927 files, 30,523,048 identical bytes on all four backends. Shards: 4,785 rows, 34,069,244 identical bytes at 1, 2, and 5 shards. Discovered JSX trees: 233 sources, 145,464 identical bytes. All 84 registered rule mutants passed.

No shared JSX count edit or count patch is included. The no-head-import-in-document follow-up is reserved by the older origin/codex/stage1-nextjs-lint claim; see next-head-availability.md.

Top-level checks:

--- SKIP: TestJsxLintReleaseAndThroughput (0.00s)
--- PASS: TestJsxInventoryDiscovery (0.00s)
--- SKIP: TestThroughput (0.00s)
--- SKIP: TestProfileArtifacts (0.00s)
--- PASS: TestCapturedOracleRecovery (57.72s)
--- PASS: TestUnmarkedMalformedOracleInputStillFails (0.02s)
--- PASS: TestRecoveryClassificationPreservesModes (0.00s)
--- SKIP: TestProfileSnapshotsAgree (0.00s)
--- PASS: TestNestedOutsideModuleCopy (1.67s)
--- PASS: TestRegistrationMutant (12.75s)
--- PASS: TestDecodedOptionsAndMutant (161.79s)
--- PASS: TestFactoryHooks (165.39s)
--- PASS: TestOwnedWitnesses (153.95s)
--- PASS: TestOptionAndComparatorGaps (0.85s)
--- PASS: TestCheckerCacheSourceByte (181.72s)
--- PASS: TestPositionIndexMutant (37.23s)
--- PASS: TestCommentFoldMutant (34.91s)
--- PASS: TestWitnessScriptKind (39.97s)
--- PASS: TestJsxLintTrees (99.01s)
--- PASS: TestDotARename (14.49s)
--- PASS: TestEmittedJavaScriptMismatch (50.09s)
--- PASS: TestCheckerHashes (1.02s)
--- SKIP: TestCheckerBridgeRefusalPending (0.00s)
--- PASS: TestCheckerReplayEntryControl (5.99s)
--- PASS: TestNestedConstructorGap (0.33s)
--- PASS: TestCommandDiagnosticsDropOnlyModuleDownloads (0.00s)
--- PASS: TestExecuteFailsOnStderrOtherThanModuleDownloads (0.23s)
--- PASS: TestCheckerNoProgramCoverage (2.59s)
--- PASS: TestSuggestionAlongsideAutomaticFix (194.19s)
--- PASS: TestCompleteSuggestionSerialization (171.05s)
--- PASS: TestDecorationOptionMutant (40.02s)
--- PASS: TestNodeTableIsLinkOnly (81.35s)
--- PASS: TestCountGuardMutant (48.14s)
--- PASS: TestRulesAgree (317.94s)
--- PASS: TestProfileCompilation (150.66s)
--- PASS: TestLegacyMutants (168.90s)
--- PASS: TestCompilerAndStage1Agree (736.62s)
--- PASS: TestShardsAgree (325.67s)
--- PASS: TestMutants (0.07s)
