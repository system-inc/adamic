Built: rebased both owned branches onto lint integration b46914832, containing current fetched main c7991b900; kept ledger winners and unchanged .a implementations; no new helper claimed.
Commits: pre-evidence rule a91758980f3884b6111ddcf9ade6b96b3beafa05 and helper 3a988db7e57eb8f2acd00414528cd7fdce4c86dd; final evidence commits pushed to their own branch names only.
Checks: six-rule aggregate PASS 699.303s, 2,221 cases / 79,633,235 identical bytes; required shared 446-file corpus matches 20,821,400 bytes; nine helpers PASS 313.196s; owned packages and vets pass; setup retry PASS 527s, nproc 5.
Mutants: all 54 owned controls freshly caught: 7 full-rule semantic, 23 listener/path/partial-decision, 23 helper actual-Go semantic and 1 helper refusal comparison; compilation/sanitizer failures are not mutant credit.
Uncovered: full shared gate remains red on parser recovery and raw witness path; eight partial finding adapters and exact bigint counter remain blocked; sixteen other mandatory checks retain prior evidence rather than a fresh full-gate claim.

Both branches now contain b46914832d70e00847d82d5d221ab7bb24040c53 and origin/main c7991b900362796aefd111474e65eb5398e91953. Rebases completed cleanly. Read the integration dedup ledger first; its verdicts are unchanged. The three losing TypeScript copies remain removed. The six owned registered winners retain ast.Kind names, node:true and supplied-node visitors. No compiler, shared finding/context/main, generator, shared oracle or comparison source is edited. New legacy registry rules, RuleContext.has and the options guard are accepted from integration unchanged. Neither main nor area is pushed. Source behavior is unchanged; this refresh contributes evidence only.

The owned Boolean outcome and project-import adapters already decode field 5 into ConsistencyNoBooleanOutcomeOptions and BoundaryNoProjectImportOptions. The other four owned rules have no upstream options type. No owned rule triggers the new dropped-options guard in either complete replay. The five Boolean JSON decoder controls (unknown fields, case-folded keys, null items and both duplicate-key orders), upstream project-import options and additional option/shape comparison remain exercised. No guard is weakened or bypassed.

The six-rule aggregate uses the actual landed harness and actual pinned Go cohere. Its 356 compiler/stage1 source files contribute 2,136 source/rule rows; 85 supported upstream fixture, witness, path and JSON option rows yield 2,221 total cases. Source Node, emitted JavaScript and ASan/UBSan native agree byte for byte with Go on 79,633,235 bytes including findings and fixes. The aggregate's explicit supported-domain limits remain those documented in the rule sibling's AREA-LANDING.md; parser recovery cases are independently checked as refusals, not green parity. The default required corpus expands to 446 files and agrees on 20,821,400 bytes. Additional owned Boolean/ORM option/shape comparison agrees on 35,048 bytes.

The guarded shared command exits 1 in 433.144s: TestRulesAgree FAIL 151.89s, TestCompilerAndStage1Agree PASS 213.70s, TestOwnedWitnesses FAIL 16.88s, TestWave12OptionsAndShapes PASS 50.60s. Parser recovery now fails at upstream case-389/Thing.ts (renumbered from case-359 by added registry captures), unsupported primary Unknown at byte 87. The exact source, logical path and options remain in the rule sibling's gaps/cases.json and internal-recovery.a; the source ends with literal backslash-n. Three owned parser probes still observe actual Go findings and explicit exit 70 on source Node, emitted JS and sanitized native: internal recovery Unknown 87, project recovery Unknown 84, outside top-level await expected semicolon 21. TestOwnedWitnesses observes no Go findings for nexus/boundary-no-nexus-outside-import at the raw repository witness path, outside libraries/nexus. The following project witness appears to need libraryDirectory options, but that is inspection/inference because execution stops at the outside-import witness. These failures remain recorded; no test is skipped, relaxed or deleted. The landing cap is not satisfied, so no helper claim or exhaustion claim is made.

Owned package times: listeners/path/parser refusals PASS 145.616s, Tailwind decisions PASS 42.236s, Next decisions PASS 34.389s. Both owned vets are clean. Helpers PASS 313.196s, with actual-Go/source-Node/emitted-JS/sanitized-native contract comparisons for all nine delivered helpers. Allocation alias and the numeric counter approximation are semantic controls caught by Go; the counter approximation diverges at 2^53. The separator guard is a separate explicit-refusal comparison. Nil-precondition refusals are checked separately, not counted as semantic mutants. Seven full-rule controls, fourteen subscribed-kind controls, one resolved-path control, three Tailwind and five Next decisions, and twenty-four helper controls all compile/run before their particular comparisons catch them. Full names and catcher observations are retained below and in the logs.

Existing normalizeValueFunctionArgument and segment each remove one dependency from better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Seven CFG helpers supply prerequisites for array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Zero consumers lose their final blocker. Exact nextBuildCount still fails before backend emission because stage 0 cannot lower a bigint-returning function. Eight JSX/Tailwind decision kernels remain independently checked, without complete extraction/utility/finding adapters. No fresh throughput measurement is claimed; previous per-rule findings/s remain historical AREA-LANDING.md measurements.

The earlier complete mandatory-input run on d65a8f931 passed all 17 checks with zero skips; its pins, exact npm lockfile, commands and raw output remain in REQUIRED-INPUTS.md / C799-LANDING.md and evidence/required-correctness-physical.log. The final-base required lint/compiler corpus is freshly rerun with ADAMIC_TYPESCRIPT_SOURCE supplied. The other sixteen correctness checks are not rerun here, and the full repository gate is not claimed. The worker gate uses touched packages plus a filtered oracle. No absent-input bypass or skip flag is set.

Toolchain setup initially ran out of workspace capacity while using the default 26 GB Go build cache. Clearing that regenerable cache during active setup caused missing-cache import failures; that failed attempt is retained, not called green. The first isolated-cache retry and initial helper/shared reruns then exhausted the 8.8 GB tmpfs: clang IO failure and Go link 'no space left on device' stopped helper checks, and the shared corpus stopped with a broken pipe. These are incomplete storage runs, not semantic findings or mutant catches. All oracle reruns use disk-backed /workspace/scratch/wave12-b469-gate after sourcing the tools. Setup's final retry reuses a disk-backed cache with GOFLAGS=-p=1; it exits 0. Go 1.27.1 ready 0s, clang 20.1.8 ready 1s, Node 24.19.0 ready 1s, submodules ready 1s, cache warm/total 527s; nproc 5, four-core quota, 17.6 GB. Sources, tests and prior logs are retained. No production source is changed to work around either resource failure.

Commands, output redirected directly to the saved logs:

```sh
GOCACHE=/workspace/scratch/wave12-b469-go-cache GOFLAGS=-p=1 bash cloud/setup.sh > /tmp/wave12-b469-setup-final.log 2>&1
source /workspace/adamic-tools/env.sh
export GOCACHE=/tmp/wave12-required-go-cache
export TMPDIR=/workspace/scratch/wave12-b469-gate
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-b469-helpers-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-b469-helper-vet.log 2>&1
# Owned rule package/vet were completed before switching from /tmp/adamic-gate:
go test -p 1 ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... -count=1 -v -timeout=20m > /tmp/wave12-b469-owned.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... > /tmp/wave12-b469-owned-vet.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestCompilerAndStage1Agree|TestWave12OptionsAndShapes)$' -count=1 -v -timeout=25m > /tmp/wave12-b469-shared-final.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 go test ./stage1/cohere/lint -run '^TestLandingExistingCandidates$' -count=1 -v -timeout=25m > /tmp/wave12-b469-aggregate.log 2>&1
```

Helper commands run on the helper branch. Rule/shared/aggregate commands run in the physical isolated checkout /workspace/scratch/wave12-area-validation, detached at b46914832, with the six owned directories copied in and existing owned Go validation artifacts enabled. No shared compatibility patch is present. Go cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db, its TypeScript checkout at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and the real TypeScript 6.0.3 source corpus at 050880ce59e30b356b686bd3144efe24f875ebc8. The initial storage-failed logs remain alongside final completed logs.

Fresh full-rule semantic mutant names, each caught only by Go findings/fixes byte comparison on source Node, emitted JavaScript and sanitized native:

- alias_prefix_boundary
- boolean_outcome_companion_ignored
- last_internal_owner
- next_document_import_path_exemption_ignored
- orm_bare_decorator_ignored
- role-suffix regex disabled
- whitelist_widened

Fresh named owned subtests (the allocation alias, counter approximation and separator guard are in parent-test logs):

```
TestNamedListenerDeclarations/base-correctness-require-orm-column-declare (1.64s)
TestNamedListenerDeclarations/nexus-consistency-no-boolean-outcome (1.05s)
TestNamedListenerDeclarations/next-google-font-preconnect (1.76s)
TestNamedListenerDeclarations/next-inline-script-id (1.98s)
TestNamedListenerDeclarations/next-next-script-for-ga (1.64s)
TestNamedListenerDeclarations/next-no-before-interactive-script-outside-document (2.17s)
TestNamedListenerDeclarations/next-no-css-tags (2.63s)
TestNamedListenerDeclarations/next-no-document-import-in-page (2.39s)
TestNamedListenerDeclarations/better-tailwindcss-enforce-shorthand-classes (2.16s)
TestNamedListenerDeclarations/better-tailwindcss-no-concatenated-classes (2.39s)
TestNamedListenerDeclarations/better-tailwindcss-no-conflicting-classes (2.13s)
TestNamedListenerDeclarations/nexus-boundary-no-internal-import (0.72s)
TestNamedListenerDeclarations/nexus-boundary-no-nexus-outside-import (0.62s)
TestNamedListenerDeclarations/nexus-boundary-no-project-import (0.61s)
TestTailwindDecisions/better-tailwindcss-enforce-shorthand-classes (9.51s)
TestTailwindDecisions/better-tailwindcss-no-concatenated-classes (8.03s)
TestTailwindDecisions/better-tailwindcss-no-conflicting-classes (12.14s)
TestExtractedDecisions/next-google-font-preconnect (2.00s)
TestExtractedDecisions/next-no-css-tags (1.40s)
TestExtractedDecisions/next-inline-script-id (1.32s)
TestExtractedDecisions/next-next-script-for-ga (1.59s)
TestExtractedDecisions/next-no-before-interactive-script-outside-document (1.28s)
TestBigIntNormalizationMutants/hex-digit-value (7.74s)
TestBigIntNormalizationMutants/negative-zero (3.94s)
TestBigIntNormalizationMutants/invalid-fallback (8.72s)
TestBreakableStatementMatchesGo/missing-switch (4.93s)
TestBreakableStatementMatchesGo/nil-loop (15.47s)
TestDecoratorsMatchesGo/drop-first-decorator (2.42s)
TestDecoratorsMatchesGo/duplicate-expression (3.06s)
TestLabelsOfMatchesGo/omit-outer-labels (25.89s)
TestLabelsOfMatchesGo/ignore-child-identity (21.65s)
TestLabelsOfMatchesGo/wrong-kind (19.09s)
TestNormalizationMutants/namespace-suffix (3.68s)
TestNormalizationMutants/newline-crossing (4.19s)
TestNormalizationMutants/nested-key-join (2.91s)
TestSegmentMutants/drop-final-empty (10.05s)
TestSegmentMutants/pop-unmatched-closer (10.14s)
TestSegmentMutants/escape-disabled (12.18s)
TestAlwaysTruthyTestMatchesGo/zero-bigint (11.49s)
TestAlwaysTruthyTestMatchesGo/skip-parentheses (9.93s)
TestAlwaysTruthyTestMatchesGo/empty-string (11.45s)
TestTypeParametersMatchesGo/drop-constraint (2.96s)
TestTypeParametersMatchesGo/swap-constraint-default (1.42s)
TestLandingExistingCandidates/base-correctness-require-orm-column-declare (44.62s)
TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome (57.94s)
TestLandingExistingCandidates/next-no-document-import-in-page (45.17s)
TestLandingExistingCandidates/nexus-boundary-no-internal-import (44.84s)
TestLandingExistingCandidates/nexus-boundary-no-nexus-outside-import (35.63s)
TestLandingExistingCandidates/nexus-boundary-no-project-import (38.02s)
TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome#01 (38.49s)
```
