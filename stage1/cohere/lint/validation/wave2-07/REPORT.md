# Lint batch wave2-07 pre-gate

Tested tree: `251e5625e15ff69a96ebf19e356241f71c025663`. Base: `f179cfb4bbacdc54911fcf3b5ec365b7c24e6914`.

## Merge scope

The first two merges are the authorized shared exceptions: nonprogressing-fix-time `6ca65dbec`, then shards-program-line `eaddd5f73`. Rule/helper merges follow in the requested order, with two parents each. No rebase or squash. The explicitly authorized capturedTypedCases count changed from 373 to 415 for @typescript-eslint/prefer-reduce-type-parameter, after checking upstream RunTypedFiles tests. No strict-alone cases were added. The owned Next document oracle adapter was formatted.

| Candidate | Pinned commit | Result |
| --- | --- | --- |
| yoda | 0eb2a004 | merged |
| lint-rules/wave1-14 | 10b2a42c0 | merged |
| react-default-props-match-prop-types | 168f4f53e | merged |
| lint-rules/wave1-02-ready | 179e22d22 | merged |
| lint-landing/no-irregular-whitespace-wave2-02 | 19795e9e | merged |
| react-jsx-no-target-blank | 24f173097 | merged |
| codex/lint-land-wave-23 | 25bccfc59 | merged |
| lint-rules/wave1-03-ready | 2920f4dbf | merged |
| no-explicit-any | 35a930bbf7 | merged |
| structure-storage-no-direct-local-storage | 3f5515631 | merged |
| structure-react-hook-require-result-naming | 406cf23be | merged |
| no-useless-call | 5efaf862f | merged |
| react-jsx-no-script-url | 65acae9f0 | merged |
| lint-rules/wave15-landing | 7edc01d8e | merged |
| react-no-unknown-property | 811208db0 | merged |
| react-no-unused-state | 9b9f4b62c | merged |
| structure-next-require-api-parameter-name | a0c743509 | merged |
| react-hooks-void-use-memo | b2e5f9a65 | merged |
| one-var | b5bca97f8 | merged |
| react-forbid-dom-props | b5c8169df | merged |
| constructor-super | bdb23f59 | merged |
| lint-rules/wave1-13-unparked | d308a2b46 | merged |
| codex/lint-wave1-05-clean | e4b228c4d | merged |
| react-jsx-no-duplicate-props | e98e9024e | merged |
| react-no-set-state | f123fc6b2 | merged |
| prefer-spread | f1490cf7c | merged |
| no-this-before-super | fac54961e | merged |
| react-no-access-state-in-setstate | fb0387acb | merged |
| react-no-danger | 32f795310 | merged |
| lint-rules/nextjs-land | 7cc48f1ba | left out |
| lint-helpers/rules-typescript | 2f3a36d8 | left out |
| lint-helpers/ecmascript-classmembers | 32193c9b | left out |
| lint-helpers/ecmascript-classmembers | ecfd7ffc | merged |
| lint-helpers/rules-react | 673b91d8d | already contained |
| lint-helpers/ecmascript-nextjs | 71ec7caf | merged |
| lint-helpers/rules-next | 9325a8efd | merged |
| lint-helpers/ecmascript-comments | 967d8feaf | merged |
| lint-helpers/jsx | a5c4c2e2b | left out |
| lint-helpers/ecmascript-regexpattern | bca49a5a | already contained |
| lint-helpers/ecmascript-regexpattern | c55c6335 | already contained |

### Exclusions

- lint-rules/nextjs-land `7cc48f1ba`: rule directory already on pinned base: stage1/cohere/lint/rules/next-no-sync-scripts/
- lint-helpers/rules-typescript `2f3a36d8`: outside owned scope: stage1/cohere/lint/helpers/sourcename/treated_as.a
- lint-helpers/ecmascript-classmembers `32193c9b`: outside owned scope: stage1/cohere/lint/helpers/property/name_tagged.a; outside owned scope: stage1/cohere/lint/helpers/property/tagged_test.go; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/REPORT.md; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/calls.json; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/capture-core.jsonl; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/capture-property.jsonl; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/capture-react.jsonl; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/capture-typescript.jsonl; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/capture.py; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/coverage.json; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/main.a; outside owned scope: stage1/cohere/lint/helpers/property/testdata/tagged/selected.log
- lint-helpers/rules-react `673b91d8d`: pinned commit is already an ancestor of the assembled batch
- lint-helpers/jsx `a5c4c2e2b`: rule directory already on pinned base: stage1/cohere/lint/rules/next-no-sync-scripts/
- lint-helpers/ecmascript-regexpattern `bca49a5a`: pinned commit is already an ancestor of the assembled batch
- lint-helpers/ecmascript-regexpattern `c55c6335`: pinned commit is already an ancestor of the assembled batch

## Gate inputs and commands

nproc 5; cpu.max `400000 100000`. TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8` at `/workspace/scratch/typescript-6.0.3`; porcelain status including ignored files empty before and after. Fresh common profile directory `/tmp/wave2-07-profile-796384`. WASI sysroot `/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot`. Lint and parser benchmark flags and ADAMIC_GATE_UNCACHED are 1. GOPROXY uses pipe fallback.

Each package ran `go test <package> -count=1 -v -json -timeout=3h`. All 95 changed Go files were checked by gofmt -l and go vet in their correct package/module context; no formatting residue or unused imports. See vet-coverage.json for the adapters and detached proof fixtures.

## Packages

| Package | Pass/fail/skip (all tests) | Pass/fail/skip (top level) | Test elapsed | Command wall |
| --- | --- | --- | --- | --- |
| github.com/system-inc/adamic/stage1/cohere/lint/helpers/comments | 14/0/0 | 7/0/0 | 404.707s | 417.669s |
| github.com/system-inc/adamic/stage1/cohere/lint/helpers/next | 1/0/0 | 1/0/0 | 14.272s | 26.222s |
| github.com/system-inc/adamic/stage1/cohere/lint/helpers/property | 1/0/0 | 1/0/0 | 132.301s | 168.738s |
| github.com/system-inc/adamic/stage1/cohere/lint/helpers | 40/0/0 | 14/0/0 | 1089.617s | 1106.011s |
| github.com/system-inc/adamic/stage1/cohere/lint/shards | 5/0/0 | 5/0/0 | 0.035s | 0.661s |
| github.com/system-inc/adamic/stage1/cohere/lint | 222/4/1 | 43/4/1 | 9047.744s | 9060.097s |
| github.com/system-inc/adamic/stage1/typescript/parser | 179/0/0 | 36/0/0 | 2664.244s | 2679.366s |

Skip: `TestCheckerBridgeRefusalPending`, `checker_pending_test.go:51`: awaits `codex/tsgo-errors-as-values`; `tsgoInspect` must return `TSGoError` from the C error buffer. No input was withheld.

`TestTopLevelTestsAreParallel` passed; no branch was refused for a serial top-level test.

## Lint serial phase and top-level test times

Serial phase (package start to last serial test finish): 1834.482072s, ending with TestProfileArtifacts. First parallel CONT: 1834.511072s. Seat baseline whole wall: 3,421.6s at f44d1b88. This batch includes 67 additional rules. This is a different box and a failing run; wall differences do not establish a green-gate speedup.

| Top-level test | Own time | Result |
| --- | --- | --- |
| TestRulesAgree | 1900.79s | fail |
| TestLegacyMutants | 1243.55s | pass |
| TestCheckerCacheSourceByte | 1218.44s | pass |
| TestProfileSnapshotsAgree | 1154.02s | pass |
| TestNonprogressingFix | 1045.25s | pass |
| TestProfileCompilation | 1025.57s | pass |
| TestJsxLintReleaseAndThroughput | 949.00s | fail |
| TestCompilerAndStage1Agree | 857.69s | pass |
| TestShardsAgree | 853.81s | pass |
| TestSuggestionAlongsideAutomaticFix | 786.00s | pass |
| TestCountGuardMutant | 741.07s | pass |
| TestCompleteSuggestionSerialization | 716.72s | pass |
| TestDecorationOptionMutant | 697.23s | pass |
| TestDotARename | 670.82s | pass |
| TestEmittedJavaScriptMismatch | 667.80s | pass |
| TestJsxLintTrees | 656.49s | fail |
| TestFactoryHooks | 650.50s | pass |
| TestCommentFoldMutant | 585.65s | pass |
| TestPositionIndexMutant | 576.80s | pass |
| TestProfileArtifacts | 462.70s | pass |
| TestThroughput | 418.91s | pass |
| TestWitnessScriptKind | 177.34s | pass |
| TestOwnedWitnesses | 127.53s | fail |
| TestNonprogressingFixPanicMutant | 113.69s | pass |
| TestRegistrationMutant | 87.50s | pass |
| TestDecodedOptionsAndMutant | 85.66s | pass |
| TestNodeTableIsLinkOnly | 57.79s | pass |
| TestCheckerReplayEntryControl | 41.50s | pass |
| TestExecuteFailsOnStderrOtherThanModuleDownloads | 14.08s | pass |
| TestTypedReplayKeepsUpstreamCompilerOptions | 7.67s | pass |
| TestOptionAndComparatorGaps | 5.01s | pass |
| TestCheckerNoProgramCoverage | 4.09s | pass |
| TestNestedOutsideModuleCopy | 2.30s | pass |
| TestChildCPUHangGuard | 1.85s | pass |
| TestChildCPUWaitGuard | 1.65s | pass |
| TestCheckerHashes | 1.38s | pass |
| TestCompilerCorpusSourcesExcludeOnlyNamedFolders | 1.05s | pass |
| TestCapturedOracleRecovery | 0.86s | pass |
| TestNestedConstructorGap | 0.40s | pass |
| TestChildWallBackstop | 0.21s | pass |
| TestMutants | 0.19s | pass |
| TestUnmarkedMalformedOracleInputStillFails | 0.06s | pass |
| TestTopLevelTestsAreParallel | 0.03s | pass |
| TestCompilerGuardBackend | 0.00s | pass |
| TestCommandDiagnosticsDropOnlyModuleDownloads | 0.00s | pass |
| TestCheckerBridgeRefusalPending | 0.00s | skip |
| TestRecoveryClassificationPreservesModes | 0.00s | pass |
| TestJsxInventoryDiscovery | 0.00s | pass |

Go reports TestMutants own time as 0.19s; its parallel children span 4,665.819s from parent CONT to terminal PASS. These are different measurements, and the complete JSON events are retained.

## Registered rules and mutant catches

166 registered rules, 99 at base and 67 added. New rules and owner commits are in new-rules.json. Full per-backend catch messages are in mutant-catches.json. A missing catch is reported as missing, never inferred from another rule.


## Failure evidence

Complete failure messages, including test source locations, are preserved in failure-output.json and the compressed package logs. Branch attribution, test file:line, messages and minimized reproducers are in [FAILURES.md](FAILURES.md).

## Completion and mutant coverage

166/166 rule mutants logged catches on Node and emitted JavaScript. Native canaries: ['react/jsx-no-comment-textnodes'].

Missing two-backend catches: none.

Nonzero command: ./stage1/cohere/lint exit 1. Full JSON event log retained.

## Resource and corpus observations

Lint load before: `2.03 5.61 7.84`; after: `4.41 4.85 4.94`. nproc 5, four-core CPU quota. Full resource samples are in logs/resources.jsonl.gz. Parser load before: `2.40 2.20 2.72`; after: `5.55 5.54 5.79`.

Compiler/stage1 comparison passed for 1,031 files: 39,825,583 bytes identical across Go, Node, emitted JavaScript and sanitized native. Native shard times: 557.158, 363.913, 380.481, 359.362, 269.630 seconds.

Parser passed all 22,497 planned incomplete-source cases.
