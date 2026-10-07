Built: refreshed both existing owned branches onto integration b84a9d931 containing main c7991b900; no new helper claimed or source behavior changed.
Commits: pre-evidence rebased rule 0139c2f212a2418135f0a1d4ea832c366786b8eb, helper df32a8831ce1a6dd5c7d2f24e86596bfa2223697; each final evidence commit is pushed to its own branch only.
Checks: owned packages/vets pass; 399-file Go/Node/emitted-JavaScript/sanitized-native corpus matches 20,820,793 bytes; earlier d65a8f931 mandatory suite passes all 17 checks with zero skips; setup 63s, nproc 5.
Mutants: all 47 owned controls caught freshly; 23 rule listener/path/partial decisions, 23 helper actual-Go semantic controls, one helper refusal comparison; historical full-rule and external-library controls are identified separately.
Uncovered: full shared rule gate fails parser recovery and raw witness path; eight partial adapters and bigint counter remain blocked, other sixteen mandatory checks not rerun after the last rebase, no full repository gate or new claim.

Both owned branches were rebased cleanly onto lint integration b84a9d9314b65d3d0261ee017e233287b4f071da, containing current fetched main c7991b900362796aefd111474e65eb5398e91953. The new integration changes proven-predicate lowering and the native record runtime; the lint harness, parser and DEDUP_LEDGER.md are unchanged. The ledger remains applied: this unit's losing no-this-alias, no-non-null-asserted-optional-chain and no-unnecessary-type-constraint copies stay removed. Six owned registered rules retain named kinds and supplied-node visitors. No main/area branch, shared compiler, harness, generator, oracle or comparison file is changed or pushed.

With /workspace/adamic-tools/env.sh sourced and GOCACHE=/tmp/wave12-required-go-cache, the final-base commands were:

```sh
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-c799-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-c799-helper-vet.log 2>&1
go test -p 1 ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... -count=1 -v -timeout=20m > /tmp/wave12-c799-owned.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... > /tmp/wave12-c799-owned-vet.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^(TestCompilerAndStage1Agree|TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout=25m > /tmp/wave12-c799-rules.log 2>&1
```

Helper test/vet run in the helper branch. Rule and shared commands run in /workspace/scratch/wave12-area-validation, detached at b84a9d931, with only six owned rule directories copied and existing owned Go validation artifacts enabled. There is no shared compatibility patch. The full helper package passes in 220.488s. Rule package results: listeners/path/parser refusals PASS 94.209s, Tailwind decisions PASS 22.698s, Next decisions PASS 7.561s; both vets are clean. The shared command exits 1 in 298.402s: TestCompilerAndStage1Agree passes in 152.93s (399 files, 20,820,793 identical actual-Go/source-Node/emitted-JavaScript/sanitized-native finding/fix bytes), while TestRulesAgree fails in 130.83s and TestOwnedWitnesses fails in 14.61s. No skips occur in these runs.

The shared-rule failures are observed blockers, not a green landing. TestRulesAgree refuses upstream case-359/Thing.ts: parser slice unsupported primary Unknown at 87. Owned gaps/cases.json contains the precise source, path and options; the source ends with literal backslash-n after the import. TestSharedNexusParserGaps again shows actual Go findings but explicit exit 70 on source Node, emitted JavaScript and sanitized native for internal recovery (Unknown 87), project recovery (Unknown 84), and outside top-level await (expected semicolon 21). TestOwnedWitnesses reports no Go findings for nexus/boundary-no-nexus-outside-import at its raw repository witness path, which is outside libraries/nexus. Inspection suggests the project's following witness also needs libraryDirectory options, but the runner stops at the outside-import witness; that later failure is an inference. No shared check is relaxed, deleted or marked skipped. The landing cap prevents any new helper claim.

Fresh owned controls total 47: fourteen listener declarations (first subscribed kind replaced by valid Unknown), one resolved-path mutation, three Tailwind decisions, five Next decisions, and twenty-four helper controls. Listener/path/decision controls compile and exit cleanly on all three backends before actual-Go comparisons catch them. Helpers have twenty-three semantic controls caught against actual Go and one separate compiling unsupported-separator guard caught by refusal comparison. The exact-counter approximation is one of those twenty-three semantic controls, caught at the 2^53 boundary. The nil-precondition refusals are separate checks, not semantic mutant credit. Existing per-helper reports and AREA-LANDING.md list the full anchors and catchers; the fresh named subtests are listed below. Seven full-rule semantic mutants and the larger 1,939-case aggregate remain historical AREA-LANDING evidence from d65a8f931, not newly rerun evidence for c7991b900. No new throughput benchmark or full repository gate is claimed.

The delivered normalizeValueFunctionArgument and segment remove two prerequisites each from six Tailwind consumers. Seven CFG helpers provide prerequisites to array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Zero rules lose their final blocker. Exact nextBuildCount remains blocked before backend emission because stage 0 cannot lower a function returning bigint. Eight JSX/Tailwind kernels remain independently tested decision-only ports, without complete extraction, utility evaluation or finding adapters.

The required external inputs were supplied, never bypassed: TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8; npm postcss 8.5.16, postcss-scss 4.0.9, prettier 3.9.6, graphql 17.0.2, postcss-media-query-parser 0.2.3, postcss-selector-parser 2.2.3 and postcss-values-parser 2.0.1. Exact package.json and lockfile are retained. ADAMIC_JSON_NODE_ONLY was unset and ADAMIC_GITIGNORE_LARGEST=1 enabled the 100 MiB boundary.

All 17 formerly input-dependent checks passed with ZERO skips and ZERO failures on the preceding integration d65a8f931 (main 39638d9e2). The final remote check then found the new main/integration compiler changes. The owned packages, all 47 controls, vets and required lint/compiler-corpus comparison were rerun on b84a9d931 as reported above. The other sixteen required checks were not rerun after that final rebase; the complete prior run is explicitly historical. This follows the worker's touched-packages plus filtered-oracle gate; it is not a current-base full gate claim.

The initial isolated validation checkout had cohere symlinked to the helper checkout. Go package discovery consequently missed newly created oracle overlay tests, producing no-tests-to-run/missing-answer setup failures. That invalid attempt was interrupted and retained, not called green. The scratch symlink was replaced by a physical local clone at Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db and its TypeScript checkout at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. The complete corrected 17-check run then passed. Only four named stopped-job build directories were removed to recover disk space; source and logs were retained. This environment repair did not modify production sources.

Setup command bash cloud/setup.sh exited 0: Go 1.27.1, clang 20.1.8, Node 24.19.0 ready at 0-1s, cache warm and total 63s. nproc reports 5; worker quota is four cores with 17.6 GB memory. All tests redirect directly to log files.

Reproduction of the completed 17-check run, in the physical isolated checkout, after sourcing the toolchain:

```sh
unset ADAMIC_JSON_NODE_ONLY
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3
export ADAMIC_CSS_LIBRARY=/workspace/scratch/wave12-required-libraries
export ADAMIC_CSS_PRINTER_LIBRARY=/workspace/scratch/wave12-required-libraries
export ADAMIC_JSON_PRETTIER=/workspace/scratch/wave12-required-libraries
export ADAMIC_GRAPHQL_LIBRARY=/workspace/scratch/wave12-required-libraries
export ADAMIC_MEDIA_QUERY_LIBRARY=/workspace/scratch/wave12-required-libraries
export ADAMIC_SELECTOR_LIBRARY=/workspace/scratch/wave12-required-libraries
export ADAMIC_VALUES_LIBRARY=/workspace/scratch/wave12-required-libraries
export ADAMIC_GITIGNORE_LARGEST=1
export GOCACHE=/tmp/wave12-required-go-cache
go test -p 2 ./stage1/typescript/parser ./stage1/cohere/css ./stage1/cohere/graphql ./stage1/cohere/mediaquery ./stage1/cohere/selector ./stage1/cohere/values ./stage1/cohere/json ./stage1/cohere/gitignore ./stage1/cohere/lint -run '^(TestCompilerExpressionsAgree|TestWholeCompilerAgrees|TestThePortParsesAsGoCohereDoes|TestCSSPrinterAgreesWithGo|TestCSSPrinterBoundaryProofs|TestUpstreamNumericSeparatorGap|TestUpstreamRepositoryCorpusParity|TestExternalComparisonCatchesThreePrinterMutants|TestPortMatchesGoCohere|TestAdditionalJSONBoundaries|TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars|TestThePortAnswersAsGoCohereAndGitDo|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=30m > /tmp/wave12-required-correctness-physical.log 2>&1
```

Completed package/test results, in log order (repeated parser test names belong to the preceding respective package):

```
--- PASS: TestCompilerExpressionsAgree (40.71s)
--- PASS: TestWholeCompilerAgrees (58.50s)
ok  	github.com/system-inc/adamic/stage1/typescript/parser	99.257s
--- PASS: TestThePortParsesAsGoCohereDoes (87.15s)
--- PASS: TestCSSPrinterAgreesWithGo (465.16s)
--- PASS: TestCSSPrinterBoundaryProofs (3.73s)
ok  	github.com/system-inc/adamic/stage1/cohere/css	556.058s
--- PASS: TestThePortParsesAsGoCohereDoes (8.20s)
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	51.393s
--- PASS: TestThePortParsesAsGoCohereDoes (2.33s)
ok  	github.com/system-inc/adamic/stage1/cohere/mediaquery	12.779s
--- PASS: TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars (0.00s)
--- PASS: TestThePortParsesAsGoCohereDoes (24.33s)
ok  	github.com/system-inc/adamic/stage1/cohere/selector	51.437s
--- PASS: TestThePortParsesAsGoCohereDoes (2.47s)
ok  	github.com/system-inc/adamic/stage1/cohere/values	38.178s
--- PASS: TestUpstreamNumericSeparatorGap (13.89s)
--- PASS: TestExternalComparisonCatchesThreePrinterMutants (14.00s)
--- PASS: TestAdditionalJSONBoundaries (40.65s)
--- PASS: TestUpstreamRepositoryCorpusParity (139.26s)
--- PASS: TestPortMatchesGoCohere (1000.38s)
ok  	github.com/system-inc/adamic/stage1/cohere/json	1000.455s
--- PASS: TestThePortAnswersAsGoCohereAndGitDo (8.20s)
ok  	github.com/system-inc/adamic/stage1/cohere/gitignore	98.466s
--- PASS: TestCompilerAndStage1Agree (196.35s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint	196.387s
```

Fresh named owned subtests (allocation alias, counter approximation and separator guard are logged in their parent tests):

```
TestNamedListenerDeclarations/base-correctness-require-orm-column-declare (1.89s)
TestNamedListenerDeclarations/nexus-consistency-no-boolean-outcome (0.84s)
TestNamedListenerDeclarations/next-google-font-preconnect (0.82s)
TestNamedListenerDeclarations/next-inline-script-id (0.88s)
TestNamedListenerDeclarations/next-next-script-for-ga (0.66s)
TestNamedListenerDeclarations/next-no-before-interactive-script-outside-document (0.94s)
TestNamedListenerDeclarations/next-no-css-tags (0.56s)
TestNamedListenerDeclarations/next-no-document-import-in-page (0.66s)
TestNamedListenerDeclarations/better-tailwindcss-enforce-shorthand-classes (0.58s)
TestNamedListenerDeclarations/better-tailwindcss-no-concatenated-classes (0.58s)
TestNamedListenerDeclarations/better-tailwindcss-no-conflicting-classes (0.59s)
TestNamedListenerDeclarations/nexus-boundary-no-internal-import (0.68s)
TestNamedListenerDeclarations/nexus-boundary-no-nexus-outside-import (0.72s)
TestNamedListenerDeclarations/nexus-boundary-no-project-import (1.51s)
TestTailwindDecisions/better-tailwindcss-enforce-shorthand-classes (5.00s)
TestTailwindDecisions/better-tailwindcss-no-concatenated-classes (5.52s)
TestTailwindDecisions/better-tailwindcss-no-conflicting-classes (4.73s)
TestExtractedDecisions/next-google-font-preconnect (0.54s)
TestExtractedDecisions/next-no-css-tags (0.49s)
TestExtractedDecisions/next-inline-script-id (0.69s)
TestExtractedDecisions/next-next-script-for-ga (0.72s)
TestExtractedDecisions/next-no-before-interactive-script-outside-document (1.27s)
TestBigIntNormalizationMutants/hex-digit-value (3.76s)
TestBigIntNormalizationMutants/negative-zero (2.88s)
TestBigIntNormalizationMutants/invalid-fallback (2.95s)
TestBreakableStatementMatchesGo/missing-switch (4.41s)
TestBreakableStatementMatchesGo/nil-loop (3.32s)
TestDecoratorsMatchesGo/drop-first-decorator (1.41s)
TestDecoratorsMatchesGo/duplicate-expression (1.32s)
TestLabelsOfMatchesGo/omit-outer-labels (14.23s)
TestLabelsOfMatchesGo/ignore-child-identity (11.52s)
TestLabelsOfMatchesGo/wrong-kind (10.25s)
TestNormalizationMutants/namespace-suffix (2.24s)
TestNormalizationMutants/newline-crossing (2.04s)
TestNormalizationMutants/nested-key-join (2.09s)
TestSegmentMutants/drop-final-empty (5.85s)
TestSegmentMutants/pop-unmatched-closer (7.16s)
TestSegmentMutants/escape-disabled (4.72s)
TestAlwaysTruthyTestMatchesGo/zero-bigint (4.09s)
TestAlwaysTruthyTestMatchesGo/skip-parentheses (5.23s)
TestAlwaysTruthyTestMatchesGo/empty-string (9.84s)
TestTypeParametersMatchesGo/drop-constraint (1.03s)
TestTypeParametersMatchesGo/swap-constraint-default (0.91s)
```

The earlier complete mandatory suite also passed the following 67 named library mutation-proof subtests. Catchers and backend coverage are in required-correctness-physical.log; this does not claim every external control ran through emitted JavaScript. JSON TestExternalComparisonCatchesThreePrinterMutants additionally records its three actual-library controls in the parent log.

```
TestCSSPrinterAgreesWithGo/default/catches_declarations_lose_their_semicolons (57.83s)
TestCSSPrinterAgreesWithGo/default/catches_groups_ignore_the_remaining_width (30.51s)
TestCSSPrinterAgreesWithGo/default/catches_rule_bodies_lose_indentation (57.00s)
TestCSSPrinterAgreesWithGo/narrow/catches_declarations_lose_their_semicolons (31.88s)
TestCSSPrinterAgreesWithGo/narrow/catches_groups_ignore_the_remaining_width (30.53s)
TestCSSPrinterAgreesWithGo/narrow/catches_rule_bodies_lose_indentation (31.68s)
TestThePortAnswersAsGoCohereAndGitDo/catches_%q_in_uppercase_hexadecimal (34.55s)
TestThePortAnswersAsGoCohereAndGitDo/catches_**_matching_one_segment_too_few (35.16s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_[[:upper:]]_one_letter_short (31.19s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_[[:xdigit:]]_taking_G (23.41s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_a_path_never_cleaned (24.37s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_a_trailing_tab_trimmed_as_a_space (30.00s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_an_escaped_range_end_read_as_the_backslash (25.59s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower (34.15s)
TestThePortAnswersAsGoCohereAndGitDo/catches_a_linked_exclude_file_refused (41.71s)
TestThePortAnswersAsGoCohereAndGitDo/catches_entering_from_below_the_root_forgetting_the_files_above (37.35s)
TestThePortAnswersAsGoCohereAndGitDo/catches_precedence_read_from_the_root_down (20.91s)
TestThePortParsesAsGoCohereDoes/catches_--_a_word_only_before_more_text (7.83s)
TestThePortParsesAsGoCohereDoes/catches_?_not_in_a_unicode_range (7.98s)
TestThePortParsesAsGoCohereDoes/catches_DEL_written_unescaped (5.20s)
TestThePortParsesAsGoCohereDoes/catches_DEL_written_unescaped (8.98s)
TestThePortParsesAsGoCohereDoes/catches_FRAGMENT_VARIABLE_DEFINITION_no_location (14.27s)
TestThePortParsesAsGoCohereDoes/catches_U+001F_written_unescaped (4.26s)
TestThePortParsesAsGoCohereDoes/catches_U+001F_written_unescaped (9.65s)
TestThePortParsesAsGoCohereDoes/catches_U+2000_not_whitespace (3.37s)
TestThePortParsesAsGoCohereDoes/catches_U+2029_not_whitespace (5.27s)
TestThePortParsesAsGoCohereDoes/catches_U+DFFF_no_trailing_surrogate (8.91s)
TestThePortParsesAsGoCohereDoes/catches_U+FEFF_not_whitespace (3.92s)
TestThePortParsesAsGoCohereDoes/catches_a_CRLF_counted_as_two_lines (14.91s)
TestThePortParsesAsGoCohereDoes/catches_a_G_in_a_hex_escape (10.17s)
TestThePortParsesAsGoCohereDoes/catches_a_[_in_a_name (10.34s)
TestThePortParsesAsGoCohereDoes/catches_a_byte_order_mark_not_ignored (5.54s)
TestThePortParsesAsGoCohereDoes/catches_a_feature_value's_sourceIndex_not_counting_the_colon (3.38s)
TestThePortParsesAsGoCohereDoes/catches_a_leading_zero's_error_at_the_number's_start (12.84s)
TestThePortParsesAsGoCohereDoes/catches_a_lone_surrogate_written_as_itself (10.84s)
TestThePortParsesAsGoCohereDoes/catches_a_merged_word_uses_its_first_token_position (16.98s)
TestThePortParsesAsGoCohereDoes/catches_a_no-break_space_ignored (7.44s)
TestThePortParsesAsGoCohereDoes/catches_a_number's_unit_taken_off_its_end (6.98s)
TestThePortParsesAsGoCohereDoes/catches_a_quote_ignores_escape_parity (18.04s)
TestThePortParsesAsGoCohereDoes/catches_a_space_compared_with_the_comma_token's_kind (7.88s)
TestThePortParsesAsGoCohereDoes/catches_a_spread's_absent_arguments_written (10.39s)
TestThePortParsesAsGoCohereDoes/catches_a_word_ending_at_= (8.54s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum's_digits_stopping_at_8 (8.69s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum's_lowercase_stopping_at_y (9.53s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum's_uppercase_stopping_at_Y (9.95s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum_taking__ (8.99s)
TestThePortParsesAsGoCohereDoes/catches_an_escaped_quote_closing_a_string (4.01s)
TestThePortParsesAsGoCohereDoes/catches_an_escaped_triple_quote_keeping_its_backslash (12.16s)
TestThePortParsesAsGoCohereDoes/catches_an_invalid_escape_read_as_nothing (14.12s)
TestThePortParsesAsGoCohereDoes/catches_an_operator's_sourceIndex_its_index (7.71s)
TestThePortParsesAsGoCohereDoes/catches_any_closing_quote_ending_a_string (3.68s)
TestThePortParsesAsGoCohereDoes/catches_closing_braces_no_longer_include_the_closing_byte (10.18s)
TestThePortParsesAsGoCohereDoes/catches_comments_always_disappear_from_values (17.85s)
TestThePortParsesAsGoCohereDoes/catches_custom_properties_lose_their_block_values (18.46s)
TestThePortParsesAsGoCohereDoes/catches_isColor's_digits_stopping_at_8 (7.35s)
TestThePortParsesAsGoCohereDoes/catches_next_starting_at_0,_not_NaN (9.12s)
TestThePortParsesAsGoCohereDoes/catches_no_directives_an_empty_list (15.57s)
TestThePortParsesAsGoCohereDoes/catches_the_first_line's_indent_common (12.22s)
TestThePortParsesAsGoCohereDoes/catches_the_sign_bit_not_reached (9.31s)
TestThePortParsesAsGoCohereDoes/catches_the_strict_calc_check_ignoring_case (9.49s)
TestThePortParsesAsGoCohereDoes/catches_the_suffix_operator_loses_|= (14.91s)
TestThePortParsesAsGoCohereDoes/catches_the_tokenizer's_url_argument_ignoring_case (5.46s)
TestThePortParsesAsGoCohereDoes/catches_the_unexpected_token_not_the_one_named (14.02s)
TestThePortParsesAsGoCohereDoes/catches_the_url_argument_check_ignoring_case (8.69s)
TestThePortParsesAsGoCohereDoes/catches_the_word_after_an_expression_typed_as_a_media_type (3.86s)
TestThePortParsesAsGoCohereDoes/catches_trim_as_trimStart (4.52s)
TestThePortParsesAsGoCohereDoes/catches_unicodeRange's_digits_stopping_at_8 (9.05s)
```
