Built: updated both owned branches to lint integration e667e3e1 while retaining main 71d7e491; relocated the outside-import witness to its real nested nexus path; the witness filename blocker is closed; no new claims.
Commits: pre-evidence rule a755f797a71a5a61347f509d3709809813377f95; helper 0183c8e61f68ba0463a9402550bf81925f59c00f; final evidence commits are pushed only to owned branch names.
Checks: six-rule aggregate PASS 595.48s / 2,221 cases / 79657407 identical bytes; shared TestOwnedWitnesses PASS 57.60s / 126,838 bytes; nested outside-import positive/negative/all comparison matches 1126 bytes; helpers PASS 192.360s; required lint corpus PASS 166.47s / 20,828,284 bytes; setup 76.238s, nproc 5.
Mutants: all 54 owned controls freshly caught, 53 by actual-Go semantic comparison and one by refusal comparison; seven full-rule mutants also caught on sanitized native with unattached node rows.
Uncovered: two shared tests remain red on the same parser recovery fixture; 16 unchanged mandatory checks retain prior passes; eight partial adapters and exact bigint counter remain blocked; no full repository gate, fresh throughput or new claim.

Observed bases: origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b, origin/area/stage1-lint e667e3e1dbdfd1b9125c3256961bfbc8ec31946b. Main is unchanged since the previous owned rebase. Both branches take the new integration through clean merges, retain both ancestries, and are pushed only to their own names. DEDUP_LEDGER.md remains unchanged; six winning/uncontested rules retain named-kind supplied-node registration and losing copies stay removed. No main/area push or PR is made.

Incoming shared changes preserve captured-case directory paths and recursively discover nested raw witnesses. They also support several automatic edits per finding. No shared driver/context/finding/oracle/registry/comparison source is manually edited. Compiler/parser and other 16 mandatory test sources/inputs are unchanged. TestCompilerAndStage1Agree is rerun fresh: 448 files and 20,828,284 actual-Go/source-Node/emitted-JavaScript/ASan-UBSan-native bytes match. All 17 required checks previously passed without skips in MAIN4E-LANDING; the other 16 are retained prior evidence, not new runs here. No test, assertion or options guard is skipped, relaxed or deleted. External-input pins and 67 library mutation proofs plus three JSON printer controls remain recorded in that prior report.

The owned outside-import witness moves from testdata/positive.ts.txt to testdata/libraries/nexus/Thing.ts.txt, with its original case metadata adjacent. The shared harness preserves that relative path; the rule now sees /libraries/nexus/ and actual Go reports one finding. A new owned witness-validation_test.go.txt, enabled only in the scratch checkout as wave12_outside_witness_test.go, calls shared ownedWitnessRows without modification. It verifies the retained directory, requires Go count 1 inside nexus and 0 for identical source under app/Thing.ts, and compares selected plus all-rule rows on all three backends: 1126 bytes. The shared TestOwnedWitnesses now passes for the whole discovered registry, 126,838 identical bytes; the prior outside-import filename failure is resolved. The project options witness independently passes again, one finding and 524 identical bytes.

The six-rule supported aggregate covers 356 compiler/stage1 sources, 2,136 source/rule rows plus 85 supported fixture/witness/path/options rows, 2,221 cases and 79657407 identical bytes. The 85 supported fixture rows with unattached node copies match 35241 actual-Go bytes on sanitized native. Seven existing full-rule mutants are additionally caught in that mode. Fixture counts do not multiply by execution mode. Boolean/ORM options-and-shapes match 35,048 bytes.

Remaining shared failures: TestRulesAgree refuses with unsupported primary Unknown at 87 in case-389/repo/source/other/Thing.ts; TestNodeTableIsLinkOnly refuses on the same source in its plain baseline before comparing unattached rows. Preserving the original directory is now observed, rather than inferred. The literal backslash-n after the import remains in the minimal owned repro gaps/cases.json and gaps/internal-recovery.a. Independent gap probes again observe actual-Go findings and exit 70 on all three port backends for internal recovery (Unknown 87), project recovery (Unknown 84) and outside top-level await (expected semicolon 21). Those refusals are not recovered-finding parity. The broad node-table guard is not green, although the supported owned-fixture mode is. The landing cap remains unsatisfied, so no new helper is claimed and no helper exhaustion is asserted.

All 54 controls rerun: seven full-rule semantic mutations, fourteen valid-Unknown listener substitutions, one resolved-path mutant, three Tailwind decision mutations, five Next decision mutations, and twenty-four helper controls. Twenty-three helper controls compare actual-Go semantics (including allocation aliasing and the number-counter approximation at 2^53); one is the separate separator-guard refusal comparison. Nil-precondition assertions are not mutant credit. Counted semantic mutations compile and exit cleanly before byte comparison catches them. Full-rule names: orm_bare_decorator_ignored, boolean_outcome_companion_ignored, next_document_import_path_exemption_ignored, last_internal_owner, alias_prefix_boundary, whitelist_widened and role-suffix regex disabled. Other names and exact catcher observations are in fresh logs and subtests below.

Helper prerequisites are unchanged: normalizeValueFunctionArgument and segment each serve six better-tailwindcss consumers (enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes, no-unknown-classes). Seven CFG helpers serve array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Zero consumers lose their final blocker. Exact nextBuildCount cannot lower a function returning bigint. Eight partial JSX/Tailwind kernels lack complete extraction/utility evaluation/finding adapters. Historical findings/s remain in AREA-LANDING.md and REPORT.md, not a new measurement on this integration.

Setup uses export GOPROXY='https://proxy.golang.org|direct' before bash cloud/setup.sh as requested. It succeeds without module/download failures in 76.238s; Go build ready 75.800s; Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5, four-core quota, 17.6 GB memory. Every build/test invocation uses the same fallback. Tests redirect output directly to logs with disk-backed GOCACHE and TMPDIR; no storage/cache failure occurs. Helpers run on their branch. Rule checks run in the detached scratch checkout at the pre-evidence owned integration merge, with owned test artifacts and nested witness changes copied in, pinned physical cohere and no shared source patches.

Commands after sourcing /workspace/adamic-tools/env.sh, exporting GOPROXY='https://proxy.golang.org|direct', GOCACHE=/workspace/scratch/wave12-required-go-cache and TMPDIR=/workspace/scratch/wave12-b469-gate:

```sh
GOFLAGS=-p=2 bash cloud/setup.sh > /tmp/wave12-e667-setup.log 2>&1
go test -p 1 ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-e667-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-e667-helpers-vet.log 2>&1
go test -p 1 ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... -count=1 -v -timeout=20m > /tmp/wave12-e667-owned.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... > /tmp/wave12-e667-owned-vet.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 go test -p 1 ./stage1/cohere/lint -run '^(TestLandingExistingCandidates|TestRulesAgree|TestOwnedWitnesses|TestWave12OptionsAndShapes|TestWave12ConfiguredProjectWitness|TestWave12NestedOutsideWitness|TestCompilerAndStage1Agree|TestNodeTableIsLinkOnly)$' -count=1 -v -timeout=30m > /tmp/wave12-e667-rules.log 2>&1
```

Fresh results:

```
--- FAIL: TestRulesAgree (115.40s)
--- PASS: TestCompilerAndStage1Agree (166.47s)
--- FAIL: TestNodeTableIsLinkOnly (55.61s)
--- PASS: TestOwnedWitnesses (57.60s)
--- PASS: TestWave12OptionsAndShapes (78.10s)
--- PASS: TestLandingExistingCandidates (595.48s)
--- PASS: TestWave12NestedOutsideWitness (55.87s)
--- PASS: TestWave12ConfiguredProjectWitness (49.61s)
FAIL	github.com/system-inc/adamic/stage1/cohere/lint	1174.180s
--- PASS: TestSharedNexusParserGaps (48.55s)
--- PASS: TestNamedListenerDeclarations (13.82s)
--- PASS: TestResolvedImportPaths (2.80s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import	65.185s
--- PASS: TestTailwindDecisions (18.71s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/better-tailwindcss-enforce-shorthand-classes	18.726s
--- PASS: TestExtractedDecisions (8.53s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/next-google-font-preconnect	8.559s
--- PASS: TestBigIntNormalizationMatchesGo (4.81s)
--- PASS: TestBigIntNormalizationMutants (11.29s)
--- PASS: TestControlFlowBlockAllocation (4.99s)
--- PASS: TestBreakableStatementMatchesGo (10.96s)
--- PASS: TestCounterExactPrimitiveGap (0.95s)
--- PASS: TestDecoratorsMatchesGo (6.39s)
--- PASS: TestLabelsOfMatchesGo (49.96s)
--- PASS: TestLabelsNilRefused (2.23s)
--- PASS: TestNormalizationMatchesGo (2.90s)
--- PASS: TestNormalizationMutants (9.43s)
--- PASS: TestSegmentMatchesGo (9.96s)
--- PASS: TestSegmentMutants (25.54s)
--- PASS: TestSegmentSeparatorRefusals (3.76s)
--- PASS: TestAlwaysTruthyTestMatchesGo (35.64s)
--- PASS: TestAlwaysTruthyNilRefused (5.81s)
--- PASS: TestTypeParametersMatchesGo (7.71s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/wave12	192.360s
```

Fresh named controls/subtests; allocation aliasing, counter approximation and separator-guard controls are recorded in parent logs:

```
--- PASS: TestLandingExistingCandidates/base-correctness-require-orm-column-declare (42.00s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome (60.04s)
--- PASS: TestLandingExistingCandidates/next-no-document-import-in-page (41.43s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-internal-import (39.74s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-nexus-outside-import (47.93s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-project-import (45.03s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome#01 (53.50s)
--- PASS: TestNamedListenerDeclarations/base-correctness-require-orm-column-declare (0.55s)
--- PASS: TestNamedListenerDeclarations/nexus-consistency-no-boolean-outcome (0.48s)
--- PASS: TestNamedListenerDeclarations/next-google-font-preconnect (0.52s)
--- PASS: TestNamedListenerDeclarations/next-inline-script-id (0.51s)
--- PASS: TestNamedListenerDeclarations/next-next-script-for-ga (0.53s)
--- PASS: TestNamedListenerDeclarations/next-no-before-interactive-script-outside-document (0.51s)
--- PASS: TestNamedListenerDeclarations/next-no-css-tags (0.48s)
--- PASS: TestNamedListenerDeclarations/next-no-document-import-in-page (0.47s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-enforce-shorthand-classes (0.50s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-concatenated-classes (0.58s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-conflicting-classes (0.72s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-internal-import (0.74s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-nexus-outside-import (0.67s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-project-import (0.70s)
--- PASS: TestTailwindDecisions/better-tailwindcss-enforce-shorthand-classes (3.67s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-concatenated-classes (4.55s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-conflicting-classes (4.47s)
--- PASS: TestExtractedDecisions/next-google-font-preconnect (0.61s)
--- PASS: TestExtractedDecisions/next-no-css-tags (0.56s)
--- PASS: TestExtractedDecisions/next-inline-script-id (0.61s)
--- PASS: TestExtractedDecisions/next-next-script-for-ga (0.53s)
--- PASS: TestExtractedDecisions/next-no-before-interactive-script-outside-document (0.62s)
--- PASS: TestBigIntNormalizationMutants/hex-digit-value (2.76s)
--- PASS: TestBigIntNormalizationMutants/negative-zero (4.15s)
--- PASS: TestBigIntNormalizationMutants/invalid-fallback (3.30s)
--- PASS: TestBreakableStatementMatchesGo/missing-switch (3.04s)
--- PASS: TestBreakableStatementMatchesGo/nil-loop (3.05s)
--- PASS: TestDecoratorsMatchesGo/drop-first-decorator (1.59s)
--- PASS: TestDecoratorsMatchesGo/duplicate-expression (1.56s)
--- PASS: TestLabelsOfMatchesGo/omit-outer-labels (9.60s)
--- PASS: TestLabelsOfMatchesGo/ignore-child-identity (12.33s)
--- PASS: TestLabelsOfMatchesGo/wrong-kind (11.77s)
--- PASS: TestNormalizationMutants/namespace-suffix (3.32s)
--- PASS: TestNormalizationMutants/newline-crossing (2.42s)
--- PASS: TestNormalizationMutants/nested-key-join (3.34s)
--- PASS: TestSegmentMutants/drop-final-empty (6.15s)
--- PASS: TestSegmentMutants/pop-unmatched-closer (7.47s)
--- PASS: TestSegmentMutants/escape-disabled (11.32s)
--- PASS: TestAlwaysTruthyTestMatchesGo/zero-bigint (7.85s)
--- PASS: TestAlwaysTruthyTestMatchesGo/skip-parentheses (7.02s)
--- PASS: TestAlwaysTruthyTestMatchesGo/empty-string (9.01s)
--- PASS: TestTypeParametersMatchesGo/drop-constraint (1.52s)
--- PASS: TestTypeParametersMatchesGo/swap-constraint-default (1.95s)
```
