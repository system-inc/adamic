Built: updated both owned branches with lint integration b28757f3 while retaining current main 71d7e491; added owned supported-fixture certification with unattached node copies; no new claims.
Commits: pre-evidence rule 578d4f8d5f68f35a3f4b38da96f95897b3fc030e; helper 9002bfcc8b704ddd73888a389429386409750a5b; final commits are pushed to their owned names only.
Checks: fresh six-rule aggregate PASS 521.84s, 2,221 cases / 79648385 identical bytes; 85 supported fixture rows with unattached copies match 34781 Go bytes; helpers PASS 181.767s; required 448-file lint corpus PASS 144.63s / 20,826,493 bytes; setup 55.028s, nproc 5.
Mutants: all 54 owned controls freshly caught, 53 by actual-Go comparison and one by refusal comparison; seven full-rule mutants additionally caught against Go with unattached rows on sanitized native.
Uncovered: three shared test failures represent the same parser recovery and witness-filename blockers; 16 unchanged mandatory checks retain prior passes; eight partial adapters and exact bigint counter remain blocked; no full gate, new throughput, or new claim.

Observed bases are origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and origin/area/stage1-lint b28757f339f3253ed3796a4dd6138d9ee736d887. Main is unchanged since the previous rebase. Both existing branches take the new integration through clean merges and retain both ancestries. DEDUP_LEDGER.md remains unchanged: retain six owned winning/uncontested rules, with named kinds and supplied-node listeners; losing copies stay removed. No main/area push or PR is made.

The incoming integration changes only six lint files: fix budget handling, reparsing under the original path, unconverged output transport, an unattached-node-row test mode and two raw witnesses. Shared compiler/parser/helpers and 16 other mandatory check sources/inputs are unchanged. All 17 mandatory checks previously passed without skips in MAIN4E-LANDING; only the affected TestCompilerAndStage1Agree is freshly rerun here. It covers 448 files and matches 20,826,493 bytes between actual Go, source Node, emitted JavaScript and ASan/UBSan native. No skip, guard relaxation or test deletion is used. Existing external-input pins/locks and 67 library mutation proofs plus three JSON printer controls remain prior evidence, not new runs.

The only new owned code is in rules/nexus-boundary-no-internal-import/landing-validation_test.go.txt, enabled as a Go test only in the detached validation checkout. It reuses the sanitized native build, sends the existing 85 supported selected-rule fixture/path/options rows through --junk-rows, and requires complete output equality with actual Go: 34781 bytes. These rows cover all six candidates. The seven existing full-rule mutations are also run in that mode and caught only by Go finding/fix byte comparison after compiling and exiting cleanly. The shared driver, context, registry, oracle, comparison and TestNodeTableIsLinkOnly are not manually edited; all shared changes are carried exactly from origin integration. No production rule/helper behavior is changed.

Fresh full-rule support covers 356 compiler/stage1 sources, 2,136 source/rule rows plus 85 fixture/witness/path/options rows, 2,221 cases and 79648385 identical bytes. The configured project-import raw witness still produces one upstream finding and 523 identical bytes on source Node, emitted JavaScript and sanitized native. Its temporary path changes output length between runs; this report records the current run, not an inherited byte count. The Boolean/ORM options/shapes comparison matches 35,048 bytes.

Three shared tests remain red: TestRulesAgree fails on case-389/Thing.ts with unsupported primary Unknown at 87; TestNodeTableIsLinkOnly fails on the same case while executing the plain baseline, before comparing its unattached rows; TestOwnedWitnesses fails because nexus/boundary-no-nexus-outside-import's flat temporary filename is outside /libraries/nexus/. The broad node-table guard is not called green; the independent owned supported-fixture mode does pass. The parser reproducer remains in gaps/cases.json and gaps/internal-recovery.a, including literal backslash-n after the import. Owned gap probes again observe Go findings and exit 70 from all three backends for internal recovery (Unknown 87), project recovery (Unknown 84) and outside top-level await (expected semicolon 21). These refusals are not finding parity. The landing cap remains unsatisfied, so no new helper is claimed and no exhaustion claim is made.

All 54 controls rerun: seven full-rule mutations, fourteen valid-Unknown listener substitutions, one resolved-path mutant, three Tailwind decision mutants, five Next decision mutants and twenty-four helper controls. Helper controls are twenty-three actual-Go comparisons, including allocation aliasing and the number-counter approximation at 2^53, plus one separately identified separator-guard refusal comparison. Nil-precondition assertions are separate, not mutant credit. Full-rule names are orm_bare_decorator_ignored, boolean_outcome_companion_ignored, next_document_import_path_exemption_ignored, last_internal_owner, alias_prefix_boundary, whitelist_widened and role-suffix regex disabled. Counts are per mutation, not multiplied for the additional native mode. Names and exact catches remain in fresh logs and subtests below.

Existing helper prerequisites remain unchanged: normalizeValueFunctionArgument and segment each serve six better-tailwindcss rules (enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes, no-unknown-classes). Seven CFG helpers serve array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Zero consumers lose their final blocker. Exact nextBuildCount cannot lower a function returning bigint. Eight partial JSX/Tailwind kernels lack complete extraction/utility evaluation/finding adapters. Historical findings/s remain in AREA-LANDING.md and REPORT.md; no fresh throughput measurement is claimed.

Setup passes in 55.028s with Go 1.27.1, clang 20.1.8 and Node 24.19.0, nproc 5, four-core quota and 17.6 GB memory. Go build is ready at 54.426s. Tests write directly to logs, using the existing disk GOCACHE and TMPDIR; no storage/cache failures occur. Helpers run in their branch. Rules run in the detached scratch validation checkout at the pre-evidence owned rule integration merge, with owned .go.txt test artifacts enabled, physical pinned cohere and no shared source patches.

Commands after sourcing /workspace/adamic-tools/env.sh and exporting GOCACHE=/workspace/scratch/wave12-required-go-cache, TMPDIR=/workspace/scratch/wave12-b469-gate:

```sh
GOFLAGS=-p=2 bash cloud/setup.sh > /tmp/wave12-b287-setup.log 2>&1
go test -p 1 ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-b287-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-b287-helpers-vet.log 2>&1
go test -p 1 ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... -count=1 -v -timeout=20m > /tmp/wave12-b287-owned.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... > /tmp/wave12-b287-owned-vet.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 go test -p 1 ./stage1/cohere/lint -run '^(TestLandingExistingCandidates|TestRulesAgree|TestOwnedWitnesses|TestWave12OptionsAndShapes|TestWave12ConfiguredProjectWitness|TestCompilerAndStage1Agree|TestNodeTableIsLinkOnly)$' -count=1 -v -timeout=30m > /tmp/wave12-b287-rules.log 2>&1
```

Fresh results:

```
--- FAIL: TestRulesAgree (101.64s)
--- PASS: TestCompilerAndStage1Agree (144.63s)
--- FAIL: TestNodeTableIsLinkOnly (41.02s)
--- FAIL: TestOwnedWitnesses (4.31s)
--- PASS: TestWave12OptionsAndShapes (60.06s)
--- PASS: TestLandingExistingCandidates (521.84s)
--- PASS: TestWave12ConfiguredProjectWitness (51.38s)
FAIL	github.com/system-inc/adamic/stage1/cohere/lint	925.077s
--- PASS: TestSharedNexusParserGaps (55.20s)
--- PASS: TestNamedListenerDeclarations (13.67s)
--- PASS: TestResolvedImportPaths (2.81s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import	71.700s
--- PASS: TestTailwindDecisions (20.76s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/better-tailwindcss-enforce-shorthand-classes	20.773s
--- PASS: TestExtractedDecisions (5.82s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/next-google-font-preconnect	5.831s
--- PASS: TestBigIntNormalizationMatchesGo (3.93s)
--- PASS: TestBigIntNormalizationMutants (9.51s)
--- PASS: TestControlFlowBlockAllocation (5.11s)
--- PASS: TestBreakableStatementMatchesGo (15.44s)
--- PASS: TestCounterExactPrimitiveGap (1.07s)
--- PASS: TestDecoratorsMatchesGo (6.10s)
--- PASS: TestLabelsOfMatchesGo (52.67s)
--- PASS: TestLabelsNilRefused (2.63s)
--- PASS: TestNormalizationMatchesGo (3.20s)
--- PASS: TestNormalizationMutants (7.84s)
--- PASS: TestSegmentMatchesGo (5.77s)
--- PASS: TestSegmentMutants (28.29s)
--- PASS: TestSegmentSeparatorRefusals (3.68s)
--- PASS: TestAlwaysTruthyTestMatchesGo (29.01s)
--- PASS: TestAlwaysTruthyNilRefused (2.45s)
--- PASS: TestTypeParametersMatchesGo (5.05s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/wave12	181.767s
```

Fresh named controls/subtests; allocation aliasing, counter approximation and separator guard controls are recorded in parent logs:

```
--- PASS: TestLandingExistingCandidates/base-correctness-require-orm-column-declare (43.80s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome (37.81s)
--- PASS: TestLandingExistingCandidates/next-no-document-import-in-page (38.02s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-internal-import (44.93s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-nexus-outside-import (40.10s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-project-import (46.23s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome#01 (45.07s)
--- PASS: TestNamedListenerDeclarations/base-correctness-require-orm-column-declare (0.46s)
--- PASS: TestNamedListenerDeclarations/nexus-consistency-no-boolean-outcome (0.50s)
--- PASS: TestNamedListenerDeclarations/next-google-font-preconnect (0.45s)
--- PASS: TestNamedListenerDeclarations/next-inline-script-id (0.45s)
--- PASS: TestNamedListenerDeclarations/next-next-script-for-ga (0.60s)
--- PASS: TestNamedListenerDeclarations/next-no-before-interactive-script-outside-document (0.45s)
--- PASS: TestNamedListenerDeclarations/next-no-css-tags (1.12s)
--- PASS: TestNamedListenerDeclarations/next-no-document-import-in-page (0.64s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-enforce-shorthand-classes (0.43s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-concatenated-classes (0.44s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-conflicting-classes (0.45s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-internal-import (0.47s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-nexus-outside-import (0.47s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-project-import (0.57s)
--- PASS: TestTailwindDecisions/better-tailwindcss-enforce-shorthand-classes (4.13s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-concatenated-classes (4.07s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-conflicting-classes (3.78s)
--- PASS: TestExtractedDecisions/next-google-font-preconnect (0.49s)
--- PASS: TestExtractedDecisions/next-no-css-tags (0.53s)
--- PASS: TestExtractedDecisions/next-inline-script-id (0.59s)
--- PASS: TestExtractedDecisions/next-next-script-for-ga (0.61s)
--- PASS: TestExtractedDecisions/next-no-before-interactive-script-outside-document (0.58s)
--- PASS: TestBigIntNormalizationMutants/hex-digit-value (2.91s)
--- PASS: TestBigIntNormalizationMutants/negative-zero (2.92s)
--- PASS: TestBigIntNormalizationMutants/invalid-fallback (2.59s)
--- PASS: TestBreakableStatementMatchesGo/missing-switch (5.55s)
--- PASS: TestBreakableStatementMatchesGo/nil-loop (3.45s)
--- PASS: TestDecoratorsMatchesGo/drop-first-decorator (1.65s)
--- PASS: TestDecoratorsMatchesGo/duplicate-expression (1.55s)
--- PASS: TestLabelsOfMatchesGo/omit-outer-labels (16.01s)
--- PASS: TestLabelsOfMatchesGo/ignore-child-identity (10.88s)
--- PASS: TestLabelsOfMatchesGo/wrong-kind (13.51s)
--- PASS: TestNormalizationMutants/namespace-suffix (2.36s)
--- PASS: TestNormalizationMutants/newline-crossing (2.74s)
--- PASS: TestNormalizationMutants/nested-key-join (2.36s)
--- PASS: TestSegmentMutants/drop-final-empty (7.20s)
--- PASS: TestSegmentMutants/pop-unmatched-closer (8.92s)
--- PASS: TestSegmentMutants/escape-disabled (11.46s)
--- PASS: TestAlwaysTruthyTestMatchesGo/zero-bigint (5.29s)
--- PASS: TestAlwaysTruthyTestMatchesGo/skip-parentheses (5.51s)
--- PASS: TestAlwaysTruthyTestMatchesGo/empty-string (9.21s)
--- PASS: TestTypeParametersMatchesGo/drop-constraint (1.35s)
--- PASS: TestTypeParametersMatchesGo/swap-constraint-default (1.32s)
```
