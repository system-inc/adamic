Built: rebased six existing rule candidates and nine delivered helpers onto actual current main plus current lint integration; configured the owned project-import raw witness; no new claims.
Commits: pre-evidence rule 5ebf94179394647f9facd0a0554bee91a541435a; helper 6af3119ec987b1f735e94830841afc563e7b47d4; final evidence commits are pushed only to their owned branch names.
Checks: all 17 required checks PASS, zero skips/failures; six-rule aggregate PASS 665.52s / 2,221 cases / 79,633,235 identical bytes; configured witness PASS 53.16s / 523 bytes; helpers PASS 246.313s; owned packages and vets PASS; setup 69.928s, nproc 5.
Mutants: all 54 owned controls freshly caught, 53 by actual-Go semantic comparison and one explicit-refusal comparison; required checks also pass 67 named external-library mutation proofs and three JSON printer controls.
Uncovered: shared default rule gate remains red on parser recovery and outside-import raw witness filename; eight partial adapters and exact bigint counter remain blocked; no full repository gate, fresh throughput, or new helper claim.

Both branches now contain origin/main 4e0bfda50a19c705a1aac0d9932e08483806d61c AND origin/area/stage1-lint bb2ece564842c4b2f909b9f75c27e74c2efa4f29. Initial rebases onto lint integration were clean, but an ancestry check found that integration had not yet taken main's runtime/developer-tool merge. The owned commits were then rebased onto actual current main and lint integration was merged into each owned branch, without conflicts. No unowned source was manually edited. DEDUP_LEDGER.md is unchanged: retain the six winning/uncontested rules and keep the three losing copies removed. Rule descriptors retain named kinds and supplied-node visitors. No main/area push or PR is made.

The lint integration adds supported .options.json witness sidecars. The project-import directory now owns testdata/positive.options.json with libraryDirectory="/" and allowed=[], a real upstream configuration that guards an absolute temporary filename. Its original positive.case.json retains the library-specific logical path/options. The new owned witness-validation_test.go.txt is enabled only in the detached scratch checkout as wave12_project_witness_test.go. It calls shared ownedWitnessRows without modification, requires exactly one actual-Go finding, and compares source Node, emitted JavaScript and ASan/UBSan native: 523 identical bytes. No oracle filename or upstream rule body is rewritten to force the result.

The outside-import rule has no equivalent directory option. Shared ownedWitnesses writes every source to a flat temporary filename, discarding its /libraries/nexus/ logical path. TestOwnedWitnesses still fails with 'nexus/boundary-no-nexus-outside-import witness reports no findings'. The required shared fix is logical-filename support for raw witnesses; source or options alone cannot supply it. The following project witness is now separately certified rather than merely inferred.

TestRulesAgree still fails at case-389/Thing.ts, unsupported primary Unknown at 87. TestSharedNexusParserGaps preserves source, filename and options from the actual upstream originals in gaps/cases.json and its .a probes: Go reports findings, while source Node, emitted JavaScript and sanitized native all exit 70 for internal recovery (Unknown 87), project recovery (Unknown 84), and outside top-level await (expected semicolon 21). These refusals are not recovered-finding parity. The combined rule command exits 1 in 978.258s solely on TestRulesAgree and TestOwnedWitnesses; the owned aggregate, options/shapes and configured witness all pass. The landing cap remains unsatisfied, so no new helper is claimed and no helper inventory exhaustion is asserted.

The owned supported aggregate covers 356 pinned compiler/stage1 sources across six rules, 2,136 source/rule rows plus 85 supported fixture/witness/path/options cases. Complete finding/fix output matches actual Go, source Node, emitted JavaScript and sanitized native: 79,633,235 bytes. The three parser-blocked originals are independently tested as refusals. The required compiler/stage1 lint check covers 448 files and matches 20,823,038 bytes across actual Go, source Node, emitted JavaScript and sanitized native. All required inputs are provided: TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8; postcss 8.5.16, postcss-scss 4.0.9, prettier 3.9.6, graphql 17.0.2, postcss-media-query-parser 0.2.3, postcss-selector-parser 2.2.3 and postcss-values-parser 2.0.1. Package locks remain in evidence/required-libraries-package*.json. ADAMIC_JSON_NODE_ONLY is unset and ADAMIC_GITIGNORE_LARGEST=1. No check, guard or assertion is skipped, relaxed or removed.

Setup passes on the final combined tree in 69.928s; Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc 5, CPU quota four cores, memory 17.6 GB. Go build is ready at 69.633s and test binary warming is deferred by the updated shared setup default. The initial superseded integration-only setup took 100s. Its in-progress rule run was intentionally terminated when the missing-main ancestry was discovered; its log is retained as incomplete evidence, never counted as a passing final-tree run. Fresh final-tree contracts and all 17 mandatory checks replace those results. Disk-backed GOCACHE and TMPDIR avoid the previous tmpfs storage incidents.

All 54 owned controls are rerun: seven full-rule semantic mutants; fourteen first-listener-kind substitutions to valid Unknown; one resolved-path mutant; three Tailwind decision mutants; five Next decision mutants; twenty-four helper controls. Helper controls comprise twenty-three actual-Go semantic comparisons, including allocation aliasing and the number-counter approximation at 2^53, and one separately identified unsupported-separator refusal comparison. Counted semantic mutants compile and run successfully before comparisons catch them. Nil-precondition refusals are separate assertions, not mutant credit.

Full-rule mutations caught by actual-Go finding/fix bytes on all three backends: orm_bare_decorator_ignored, boolean_outcome_companion_ignored, next_document_import_path_exemption_ignored, last_internal_owner, alias_prefix_boundary, whitelist_widened, and role-suffix regex disabled. Exact other names and catcher observations are retained in the fresh logs and named subtests below. Required library checks retain their own backend coverage; emitted-JavaScript coverage is not claimed for every external-library check.

Existing helper prerequisites: normalizeValueFunctionArgument and segment each serve better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Seven CFG helpers serve array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Zero consumers lose their final blocker. Exact nextBuildCount still refuses before native emission because a function returning bigint cannot be lowered. Eight JSX/Tailwind decision kernels remain partial, without complete extraction, utility evaluation and finding adapters. Historical findings/second remain in AREA-LANDING.md and REPORT.md; no new throughput measurement is claimed on the changed runtime.

Final-tree commands redirect all test output to logs, after sourcing /workspace/adamic-tools/env.sh and exporting GOCACHE=/workspace/scratch/wave12-required-go-cache and TMPDIR=/workspace/scratch/wave12-b469-gate. Helpers run in their branch. Rule and mandatory checks run in the detached scratch checkout at the pre-evidence rule merge above, with owned Go .txt validation artifacts enabled and no shared source patches. Physical cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db and its TypeScript at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a.

```sh
GOFLAGS=-p=2 bash cloud/setup.sh > /tmp/wave12-main4e-setup.log 2>&1
go test -p 1 ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-main4e-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-main4e-helpers-vet.log 2>&1
go test -p 1 ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... -count=1 -v -timeout=20m > /tmp/wave12-main4e-owned.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... > /tmp/wave12-main4e-owned-vet.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 go test -p 1 ./stage1/cohere/lint -run '^(TestLandingExistingCandidates|TestRulesAgree|TestOwnedWitnesses|TestWave12OptionsAndShapes|TestWave12ConfiguredProjectWitness)$' -count=1 -v -timeout=30m > /tmp/wave12-main4e-rules.log 2>&1
```

The required-suite command and its input exports are the same as D3-LANDING.md, with final-tree log /tmp/wave12-main4e-required.log; its nine package filters select all 17 required checks, -p 2 -count=1 -v -timeout=30m.

Required results:

```
--- PASS: TestCompilerExpressionsAgree (45.33s)
--- PASS: TestWholeCompilerAgrees (56.35s)
ok  	github.com/system-inc/adamic/stage1/typescript/parser	101.721s
--- PASS: TestThePortParsesAsGoCohereDoes (65.36s)
--- PASS: TestCSSPrinterAgreesWithGo (581.39s)
--- PASS: TestCSSPrinterBoundaryProofs (4.69s)
ok  	github.com/system-inc/adamic/stage1/cohere/css	651.471s
--- PASS: TestThePortParsesAsGoCohereDoes (2.84s)
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	58.080s
--- PASS: TestThePortParsesAsGoCohereDoes (2.22s)
ok  	github.com/system-inc/adamic/stage1/cohere/mediaquery	19.966s
--- PASS: TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars (0.00s)
--- PASS: TestThePortParsesAsGoCohereDoes (20.59s)
ok  	github.com/system-inc/adamic/stage1/cohere/selector	51.089s
--- PASS: TestThePortParsesAsGoCohereDoes (2.27s)
ok  	github.com/system-inc/adamic/stage1/cohere/values	52.893s
--- PASS: TestUpstreamNumericSeparatorGap (3.98s)
--- PASS: TestExternalComparisonCatchesThreePrinterMutants (4.39s)
--- PASS: TestAdditionalJSONBoundaries (25.79s)
--- PASS: TestUpstreamRepositoryCorpusParity (165.62s)
--- PASS: TestPortMatchesGoCohere (1001.08s)
ok  	github.com/system-inc/adamic/stage1/cohere/json	1001.170s
--- PASS: TestThePortAnswersAsGoCohereAndGitDo (11.25s)
ok  	github.com/system-inc/adamic/stage1/cohere/gitignore	136.259s
--- PASS: TestCompilerAndStage1Agree (147.13s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint	147.242s
```

Owned results:

```
--- PASS: TestSharedNexusParserGaps (58.18s)
--- PASS: TestNamedListenerDeclarations (21.33s)
--- PASS: TestResolvedImportPaths (3.40s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import	82.927s
--- PASS: TestTailwindDecisions (28.72s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/better-tailwindcss-enforce-shorthand-classes	28.735s
--- PASS: TestExtractedDecisions (16.15s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/rules/nexus-boundary-no-internal-import/gaps/next-google-font-preconnect	16.173s
--- PASS: TestBigIntNormalizationMatchesGo (17.69s)
--- PASS: TestBigIntNormalizationMutants (10.50s)
--- PASS: TestControlFlowBlockAllocation (4.51s)
--- PASS: TestBreakableStatementMatchesGo (12.98s)
--- PASS: TestCounterExactPrimitiveGap (0.77s)
--- PASS: TestDecoratorsMatchesGo (5.83s)
--- PASS: TestLabelsOfMatchesGo (66.62s)
--- PASS: TestLabelsNilRefused (2.80s)
--- PASS: TestNormalizationMatchesGo (3.89s)
--- PASS: TestNormalizationMutants (9.72s)
--- PASS: TestSegmentMatchesGo (11.69s)
--- PASS: TestSegmentMutants (23.86s)
--- PASS: TestSegmentSeparatorRefusals (3.87s)
--- PASS: TestAlwaysTruthyTestMatchesGo (49.48s)
--- PASS: TestAlwaysTruthyNilRefused (11.08s)
--- PASS: TestTypeParametersMatchesGo (11.00s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/wave12	246.313s
--- FAIL: TestRulesAgree (170.42s)
--- FAIL: TestOwnedWitnesses (13.44s)
--- PASS: TestWave12OptionsAndShapes (75.67s)
--- PASS: TestLandingExistingCandidates (665.52s)
--- PASS: TestWave12ConfiguredProjectWitness (53.16s)
FAIL	github.com/system-inc/adamic/stage1/cohere/lint	978.258s
```

Fresh named owned subtests (allocation aliasing, counter approximation and separator guard observations are in parent logs):

```
--- PASS: TestNamedListenerDeclarations/base-correctness-require-orm-column-declare (0.64s)
--- PASS: TestNamedListenerDeclarations/nexus-consistency-no-boolean-outcome (0.65s)
--- PASS: TestNamedListenerDeclarations/next-google-font-preconnect (0.44s)
--- PASS: TestNamedListenerDeclarations/next-inline-script-id (0.85s)
--- PASS: TestNamedListenerDeclarations/next-next-script-for-ga (1.36s)
--- PASS: TestNamedListenerDeclarations/next-no-before-interactive-script-outside-document (2.34s)
--- PASS: TestNamedListenerDeclarations/next-no-css-tags (1.78s)
--- PASS: TestNamedListenerDeclarations/next-no-document-import-in-page (2.59s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-enforce-shorthand-classes (1.21s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-concatenated-classes (0.65s)
--- PASS: TestNamedListenerDeclarations/better-tailwindcss-no-conflicting-classes (0.81s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-internal-import (0.89s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-nexus-outside-import (0.69s)
--- PASS: TestNamedListenerDeclarations/nexus-boundary-no-project-import (0.73s)
--- PASS: TestTailwindDecisions/better-tailwindcss-enforce-shorthand-classes (7.62s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-concatenated-classes (4.13s)
--- PASS: TestTailwindDecisions/better-tailwindcss-no-conflicting-classes (9.20s)
--- PASS: TestExtractedDecisions/next-google-font-preconnect (1.38s)
--- PASS: TestExtractedDecisions/next-no-css-tags (1.05s)
--- PASS: TestExtractedDecisions/next-inline-script-id (0.74s)
--- PASS: TestExtractedDecisions/next-next-script-for-ga (0.70s)
--- PASS: TestExtractedDecisions/next-no-before-interactive-script-outside-document (0.80s)
--- PASS: TestBigIntNormalizationMutants/hex-digit-value (3.02s)
--- PASS: TestBigIntNormalizationMutants/negative-zero (3.44s)
--- PASS: TestBigIntNormalizationMutants/invalid-fallback (2.99s)
--- PASS: TestBreakableStatementMatchesGo/missing-switch (4.51s)
--- PASS: TestBreakableStatementMatchesGo/nil-loop (3.26s)
--- PASS: TestDecoratorsMatchesGo/drop-first-decorator (1.50s)
--- PASS: TestDecoratorsMatchesGo/duplicate-expression (1.59s)
--- PASS: TestLabelsOfMatchesGo/omit-outer-labels (16.45s)
--- PASS: TestLabelsOfMatchesGo/ignore-child-identity (14.32s)
--- PASS: TestLabelsOfMatchesGo/wrong-kind (15.48s)
--- PASS: TestNormalizationMutants/namespace-suffix (3.08s)
--- PASS: TestNormalizationMutants/newline-crossing (2.76s)
--- PASS: TestNormalizationMutants/nested-key-join (3.33s)
--- PASS: TestSegmentMutants/drop-final-empty (6.36s)
--- PASS: TestSegmentMutants/pop-unmatched-closer (9.63s)
--- PASS: TestSegmentMutants/escape-disabled (6.09s)
--- PASS: TestAlwaysTruthyTestMatchesGo/zero-bigint (6.13s)
--- PASS: TestAlwaysTruthyTestMatchesGo/skip-parentheses (17.33s)
--- PASS: TestAlwaysTruthyTestMatchesGo/empty-string (15.27s)
--- PASS: TestTypeParametersMatchesGo/drop-constraint (1.52s)
--- PASS: TestTypeParametersMatchesGo/swap-constraint-default (1.39s)
--- PASS: TestLandingExistingCandidates/base-correctness-require-orm-column-declare (53.47s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome (39.66s)
--- PASS: TestLandingExistingCandidates/next-no-document-import-in-page (42.63s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-internal-import (41.78s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-nexus-outside-import (43.20s)
--- PASS: TestLandingExistingCandidates/nexus-boundary-no-project-import (49.56s)
--- PASS: TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome#01 (39.79s)
```

Fresh external-library mutation proofs:

```
--- PASS: TestThePortParsesAsGoCohereDoes/catches_custom_properties_lose_their_block_values (8.99s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_comments_always_disappear_from_values (13.22s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_closing_braces_no_longer_include_the_closing_byte (13.77s)
--- PASS: TestCSSPrinterAgreesWithGo/default/catches_declarations_lose_their_semicolons (70.45s)
--- PASS: TestCSSPrinterAgreesWithGo/default/catches_rule_bodies_lose_indentation (61.25s)
--- PASS: TestCSSPrinterAgreesWithGo/default/catches_groups_ignore_the_remaining_width (63.44s)
--- PASS: TestCSSPrinterAgreesWithGo/narrow/catches_declarations_lose_their_semicolons (43.54s)
--- PASS: TestCSSPrinterAgreesWithGo/narrow/catches_rule_bodies_lose_indentation (31.48s)
--- PASS: TestCSSPrinterAgreesWithGo/narrow/catches_groups_ignore_the_remaining_width (39.05s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_CRLF_counted_as_two_lines (9.49s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_an_invalid_escape_read_as_nothing (12.14s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_lone_surrogate_written_as_itself (18.87s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_unexpected_token_not_the_one_named (19.57s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_leading_zero's_error_at_the_number's_start (13.31s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_an_escaped_triple_quote_keeping_its_backslash (15.65s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_U+DFFF_no_trailing_surrogate (8.01s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_first_line's_indent_common (12.82s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_[_in_a_name (19.31s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_byte_order_mark_not_ignored (10.89s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_sign_bit_not_reached (19.22s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_G_in_a_hex_escape (20.38s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_no-break_space_ignored (19.86s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_FRAGMENT_VARIABLE_DEFINITION_no_location (11.05s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_spread's_absent_arguments_written (14.80s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_no_directives_an_empty_list (7.03s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_feature_value's_sourceIndex_not_counting_the_colon (2.96s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_U+FEFF_not_whitespace (3.76s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_U+2000_not_whitespace (5.28s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_DEL_written_unescaped (5.45s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_trim_as_trimStart (6.60s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_an_escaped_quote_closing_a_string (7.86s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_U+2029_not_whitespace (7.27s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_any_closing_quote_ending_a_string (11.58s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_word_after_an_expression_typed_as_a_media_type (12.57s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_U+001F_written_unescaped (5.68s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_quote_ignores_escape_parity (17.27s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_merged_word_uses_its_first_token_position (21.07s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_suffix_operator_loses_|= (21.53s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_number's_unit_taken_off_its_end (6.78s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_alphaNum's_lowercase_stopping_at_y (10.08s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_url_argument_check_ignoring_case (11.06s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_?_not_in_a_unicode_range (15.61s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_strict_calc_check_ignoring_case (11.23s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_alphaNum_taking__ (10.92s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_word_ending_at_= (12.26s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_alphaNum's_digits_stopping_at_8 (9.73s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_isColor's_digits_stopping_at_8 (8.51s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_--_a_word_only_before_more_text (7.27s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_alphaNum's_uppercase_stopping_at_Y (15.79s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_unicodeRange's_digits_stopping_at_8 (14.89s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_a_space_compared_with_the_comma_token's_kind (12.09s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_an_operator's_sourceIndex_its_index (13.49s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_U+001F_written_unescaped (10.75s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_DEL_written_unescaped (11.54s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_next_starting_at_0,_not_NaN (9.30s)
--- PASS: TestThePortParsesAsGoCohereDoes/catches_the_tokenizer's_url_argument_ignoring_case (11.97s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_a_trailing_tab_trimmed_as_a_space (21.01s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower (34.77s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_an_escaped_range_end_read_as_the_backslash (46.81s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_[[:upper:]]_one_letter_short (47.15s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_[[:xdigit:]]_taking_G (53.64s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_a_path_never_cleaned (54.27s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_a_linked_exclude_file_refused (47.85s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_%q_in_uppercase_hexadecimal (55.48s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_precedence_read_from_the_root_down (33.21s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_entering_from_below_the_root_forgetting_the_files_above (31.82s)
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_**_matching_one_segment_too_few (29.88s)
```
