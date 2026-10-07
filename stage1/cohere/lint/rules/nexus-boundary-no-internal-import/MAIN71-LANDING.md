Built: updated both existing owned branches onto current main 71d7e491 plus lint integration bb2ece56; unchanged six rule candidates and nine delivered helpers; no new claims.
Commits: pre-evidence rule dbcd01418b8280b01a9b6a14e22420bbd40fb678; helper 8b282471b01563571ed4f10459d098d90897ac40; final report commits are pushed to their own branch names only.
Checks: fresh six-rule aggregate PASS 483.54s, 2,221 cases and 79,633,235 identical bytes; helpers PASS 164.781s; owned packages/vets, configured witness and required 448-file lint check pass; setup 91.011s, nproc 5.
Mutants: all 54 owned controls freshly caught, 53 by actual-Go semantic comparison and one by the separately identified refusal comparison.
Uncovered: shared default rule gate remains red on parser recovery and outside-import raw witness filename; eight partial adapters and exact bigint counter remain blocked; 16 unchanged mandatory checks retain prior results, no full repository gate or new throughput run.

Observed bases: origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and origin/area/stage1-lint bb2ece564842c4b2f909b9f75c27e74c2efa4f29. Main advanced only in stage3. A git diff --quiet against the previously validated main/branch confirms no change to internal, cmd, cloud, stage1, cohere, CLAUDE.md or lint-registration docs. Both owned rebases onto current main and merges retaining lint integration ancestry are clean. Both contain both fetched bases. DEDUP_LEDGER.md and its prior resolution remain unchanged; losing copies stay removed. No shared compiler, parser, harness, oracle or registry source is manually edited. No new implementation or claim is written. No main/area push or PR is made.

Fresh current-main checks compare the six supported rule candidates and all nine helpers against actual Go, source Node, emitted JavaScript and ASan/UBSan native. The six-rule aggregate covers 356 compiler/stage1 files, 2,136 source/rule rows plus 85 supported fixture/witness/path/options rows, 2,221 total cases and 79,633,235 identical finding/fix bytes. The required compiler/stage1 lint corpus covers 448 files and matches 20,823,038 bytes. The configured project-import raw witness uses its supported positive.options.json directory guard "/", yields exactly one upstream finding, and matches 523 bytes on all three backends. The Boolean/ORM options-and-shapes comparison also passes.

All 17 mandatory input-dependent checks passed without skips/failures in the prior MAIN4E-LANDING run (rules 1093b0395, helpers 4bbc5f019). Sixteen are not rerun here: their compiler/runtime/parser/test sources and exact external inputs are unchanged, and the incoming main changes are confined to stage3. TestCompilerAndStage1Agree is rerun fresh. No mandatory input is absent, no skip is accepted, and no test/guard/assertion is relaxed or removed. Exact inputs remain TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 and the pinned npm lock in evidence/required-libraries-package*.json. The 67 external-library mutation proofs and three JSON printer controls in MAIN4E-LANDING remain prior evidence, not fresh runs in this report.

The combined fresh rule command exits 1 on exactly two tests: TestRulesAgree fails at case-389/Thing.ts with 'parser slice unsupported primary Unknown at 87'; TestOwnedWitnesses fails with 'nexus/boundary-no-nexus-outside-import witness reports no findings'. Shared ownedWitnesses discards the logical /libraries/nexus/ filename for a flat temporary filename, and that rule has no directory option. The project options-sidecar witness is independently certified again; filename transport remains shared work. The independent owned gaps/cases.json probes again observe actual-Go findings and exit 70 on all three port backends for internal recovery (Unknown 87), project recovery (Unknown 84) and outside top-level await (expected semicolon 21). These refusals are not byte-equal recovered findings. The landing cap remains unsatisfied, so no new helper is claimed, and helper exhaustion is not asserted.

All 54 owned controls rerun: seven full-rule semantic mutants, fourteen valid-Unknown listener substitutions, one resolved-path mutant, three Tailwind decisions, five Next decisions and twenty-four helper controls. Of helper controls, twenty-three are actual-Go comparisons including allocation aliasing and the number-counter approximation at 2^53; one is the unsupported-separator guard refusal comparison. Nil-precondition assertions are separate and not counted as mutants. Every counted semantic mutation compiles and runs before the byte comparison catches it on source Node, emitted JavaScript and sanitized native. Full-rule names: orm_bare_decorator_ignored, boolean_outcome_companion_ignored, next_document_import_path_exemption_ignored, last_internal_owner, alias_prefix_boundary, whitelist_widened and role-suffix regex disabled. Remaining names and their catcher observations follow and remain in fresh logs.

Existing helper fanout is unchanged: normalizeValueFunctionArgument and segment each supply a prerequisite for better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Seven CFG helpers supply prerequisites for array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Zero consumers lose their final blocker. Exact nextBuildCount cannot be lowered as a function returning bigint. Eight partial JSX/Tailwind kernels retain independent decision proofs but lack complete extraction/utility evaluation/finding adapters. Historical findings/s remain in AREA-LANDING.md and REPORT.md; no fresh throughput is claimed.

Setup passes in 91.011s: Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5, CPU quota four cores, memory 17.6 GB. Go build is ready at 90.571s; shared default defers test binary warming. All test output is redirected to logs. Rule checks run in the detached scratch checkout at the pre-evidence rule merge above, with existing owned Go .txt validation artifacts enabled and no shared patches. Helper checks run in their owned branch. No storage/cache failure occurs.

Commands after sourcing /workspace/adamic-tools/env.sh and exporting GOCACHE=/workspace/scratch/wave12-required-go-cache, TMPDIR=/workspace/scratch/wave12-b469-gate:

```sh
GOFLAGS=-p=2 bash cloud/setup.sh > /tmp/wave12-71-setup.log 2>&1
go test -p 1 ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-71-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-71-helpers-vet.log 2>&1
go test -p 1 ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... -count=1 -v -timeout=20m > /tmp/wave12-71-owned.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... > /tmp/wave12-71-owned-vet.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 go test -p 1 ./stage1/cohere/lint -run '^(TestLandingExistingCandidates|TestRulesAgree|TestOwnedWitnesses|TestWave12OptionsAndShapes|TestWave12ConfiguredProjectWitness|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=30m > /tmp/wave12-71-rules.log 2>&1
```

Fresh results:

```
--- FAIL: TestRulesAgree (98.58s)
--- PASS: TestCompilerAndStage1Agree (140.48s)
--- FAIL: TestOwnedWitnesses (9.32s)
--- PASS: TestWave12OptionsAndShapes (43.04s)
--- PASS: TestLandingExistingCandidates (483.54s)
--- PASS: TestWave12ConfiguredProjectWitness (48.77s)
FAIL	github.com/system-inc/adamic/stage1/cohere/lint	823.783s
--- PASS: TestSharedNexusParserGaps (51.76s)
--- PASS: TestNamedListenerDeclarations (13.10s)
--- PASS: TestResolvedImportPaths (3.27s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import	68.146s
--- PASS: TestTailwindDecisions (17.58s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/better-tailwindcss-enforce-shorthand-classes	17.595s
--- PASS: TestExtractedDecisions (5.56s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/next-google-font-preconnect	5.574s
--- PASS: TestBigIntNormalizationMatchesGo (5.96s)
--- PASS: TestBigIntNormalizationMutants (8.32s)
--- PASS: TestControlFlowBlockAllocation (4.33s)
--- PASS: TestBreakableStatementMatchesGo (11.10s)
--- PASS: TestCounterExactPrimitiveGap (0.84s)
--- PASS: TestDecoratorsMatchesGo (7.05s)
--- PASS: TestLabelsOfMatchesGo (44.73s)
--- PASS: TestLabelsNilRefused (1.90s)
--- PASS: TestNormalizationMatchesGo (2.40s)
--- PASS: TestNormalizationMutants (9.31s)
--- PASS: TestSegmentMatchesGo (6.96s)
--- PASS: TestSegmentMutants (19.09s)
--- PASS: TestSegmentSeparatorRefusals (4.12s)
--- PASS: TestAlwaysTruthyTestMatchesGo (29.47s)
--- PASS: TestAlwaysTruthyNilRefused (2.67s)
--- PASS: TestTypeParametersMatchesGo (6.51s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/wave12	164.781s
```

Fresh named subtests; allocation aliasing, counter approximation and separator-guard controls are in parent logs:

```
--- PASS: TestLandingExistingCandidates/base-correctness-require-orm-column-declare (40.10s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome (33.99s)
--- PASS: TestLandingExistingCandidates/next-no-document-import-in-page (39.39s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-internal-import (38.73s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-nexus-outside-import (39.65s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-project-import (40.55s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome#01 (39.42s)
--- PASS: TestNamedListenerDeclarations/base-correctness-require-orm-column-declare (0.46s)
--- PASS: TestNamedListenerDeclarations/nexus-consistency-no-boolean-outcome (0.49s)
--- PASS: TestNamedListenerDeclarations/next-google-font-preconnect (0.48s)
--- PASS: TestNamedListenerDeclarations/next-inline-script-id (0.61s)
--- PASS: TestNamedListenerDeclarations/next-next-script-for-ga (0.65s)
--- PASS: TestNamedListenerDeclarations/next-no-before-interactive-script-outside-document (0.46s)
--- PASS: TestNamedListenerDeclarations/next-no-css-tags (0.44s)
--- PASS: TestNamedListenerDeclarations/next-no-document-import-in-page (0.47s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-enforce-shorthand-classes (0.56s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-concatenated-classes (0.49s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-conflicting-classes (0.63s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-internal-import (0.59s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-nexus-outside-import (0.72s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-project-import (0.52s)
--- PASS: TestTailwindDecisions/better-tailwindcss-enforce-shorthand-classes (3.13s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-concatenated-classes (3.94s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-conflicting-classes (4.92s)
--- PASS: TestExtractedDecisions/next-google-font-preconnect (0.55s)
--- PASS: TestExtractedDecisions/next-no-css-tags (0.49s)
--- PASS: TestExtractedDecisions/next-inline-script-id (0.52s)
--- PASS: TestExtractedDecisions/next-next-script-for-ga (0.50s)
--- PASS: TestExtractedDecisions/next-no-before-interactive-script-outside-document (0.62s)
--- PASS: TestBigIntNormalizationMutants/hex-digit-value (2.44s)
--- PASS: TestBigIntNormalizationMutants/negative-zero (2.44s)
--- PASS: TestBigIntNormalizationMutants/invalid-fallback (2.50s)
--- PASS: TestBreakableStatementMatchesGo/missing-switch (2.95s)
--- PASS: TestBreakableStatementMatchesGo/nil-loop (3.71s)
--- PASS: TestDecoratorsMatchesGo/drop-first-decorator (2.07s)
--- PASS: TestDecoratorsMatchesGo/duplicate-expression (1.69s)
--- PASS: TestLabelsOfMatchesGo/omit-outer-labels (9.93s)
--- PASS: TestLabelsOfMatchesGo/ignore-child-identity (10.09s)
--- PASS: TestLabelsOfMatchesGo/wrong-kind (11.76s)
--- PASS: TestNormalizationMutants/namespace-suffix (2.64s)
--- PASS: TestNormalizationMutants/newline-crossing (2.67s)
--- PASS: TestNormalizationMutants/nested-key-join (3.62s)
--- PASS: TestSegmentMutants/drop-final-empty (5.07s)
--- PASS: TestSegmentMutants/pop-unmatched-closer (8.40s)
--- PASS: TestSegmentMutants/escape-disabled (4.98s)
--- PASS: TestAlwaysTruthyTestMatchesGo/zero-bigint (5.91s)
--- PASS: TestAlwaysTruthyTestMatchesGo/skip-parentheses (7.69s)
--- PASS: TestAlwaysTruthyTestMatchesGo/empty-string (8.60s)
--- PASS: TestTypeParametersMatchesGo/drop-constraint (2.21s)
--- PASS: TestTypeParametersMatchesGo/swap-constraint-default (2.00s)
```
