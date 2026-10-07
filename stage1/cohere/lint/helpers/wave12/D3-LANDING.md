Built: rebased both existing owned branches onto integration d3a37422 containing main b6b1538b; unchanged owned .a implementations, unchanged dedup verdicts; no new helper claimed.
Commits: rebased pre-evidence rule 523957012c2b4d2c4c0d277e4c8347b80ac8825e and helper 06c3cef925a4dac37c2b7dd7b1f638c12ca67f7e; each final evidence commit is pushed only to its own branch name.
Checks: all 17 mandatory input-dependent checks PASS with ZERO skips/failures; owned six-rule aggregate PASS 672.62s, 2,221 cases / 79,633,235 identical bytes; nine helpers PASS 356.254s; owned packages/vets pass; setup 294s, nproc 5.
Mutants: all 54 owned controls freshly caught, 53 against actual Go and one explicit-refusal comparison; the mandatory suite also passes 67 named external-library mutation-proof subtests and its three JSON printer controls.
Uncovered: broader shared rule gate remains red on parser recovery and raw witness path; eight partial finding adapters and exact bigint counter remain blocked; no complete repository gate, fresh throughput or new helper claim.

The observed bases are origin/main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898. Both owned branches contain both bases. The integration delta brings native typeof/null/lookup-presence fixes; it changes no lint harness, parser, helper or DEDUP_LEDGER.md source. Rebases are clean and the previous ledger resolution remains: keep the six owned winning/uncontested rules, with named kinds and supplied-node visitors; the three losing TypeScript copies stay removed. No rule/helper behavior or shared compiler/harness/generator/oracle/comparison source is edited. No main/area push or PR is made.

Setup passes with the existing disk-backed task cache. Go 1.27.1 ready 0s, clang 20.1.8 ready 1s, Node 24.19.0 ready 1s, submodules ready 1s, cache warm and total 294s. nproc is 5, quota four cores, memory 17.6 GB. Before starting these jobs, the inactive task Go cache was moved from /tmp/wave12-required-go-cache to /workspace/scratch/wave12-required-go-cache, recovering more than 8 GB of tmpfs capacity. It was moved only when no task was using it. Test temporary files use /workspace/scratch/wave12-b469-gate (mode 1777). There are no cache-race or storage failures in this run; the earlier incidents and their invalid logs remain historical B469-LANDING evidence.

All mandatory checks have their exact external inputs supplied: TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, plus postcss 8.5.16, postcss-scss 4.0.9, prettier 3.9.6, graphql 17.0.2, postcss-media-query-parser 0.2.3, postcss-selector-parser 2.2.3 and postcss-values-parser 2.0.1 in /workspace/scratch/wave12-required-libraries. Exact package.json and lockfile are already retained under evidence/required-libraries-package*.json. ADAMIC_JSON_NODE_ONLY is unset, and ADAMIC_GITIGNORE_LARGEST=1 enables the 100 MiB boundary. No mandatory input is absent, no skip is accepted, and no test or assertion is relaxed/deleted. All 17 current-base checks finish green. This supersedes the previous reports' sixteen-check fresh-validation limitation.

The required compiler/stage1 lint corpus has 446 files and matches 20,821,400 bytes between actual Go, source Node, emitted JavaScript and sanitized native. The owned supported aggregate covers 356 pinned compiler/stage1 files across six selected rules, 2,136 source/rule combinations plus 85 supported upstream/witness/path/option cases, 2,221 cases total and 79,633,235 matching finding/fix bytes. The additional Boolean/ORM options-and-shapes comparison matches 35,048 bytes. Supported domains and exclusions remain explicit in the rule sibling's AREA-LANDING.md: the three malformed/top-level-await originals are independently exercised as refusals, and complete Next JSX fixture parity and the eight partial adapters are not claimed.

Owned package results: listeners/path/parser probes PASS 118.925s, Tailwind decisions PASS 40.999s, Next decisions PASS 24.663s; both rule and helper vets are clean. All nine delivered helper contracts pass in 356.254s against actual Go on source Node, emitted JS and ASan/UBSan native. The combined rule command exits 1 in 928.224s: TestRulesAgree FAIL 159.70s, TestOwnedWitnesses FAIL 21.19s, TestWave12OptionsAndShapes PASS 74.59s, TestLandingExistingCandidates PASS 672.62s. Thus the passing owned replay is not mislabeled as a passing full shared gate.

The observed remaining shared blockers are unchanged: upstream case-389/Thing.ts refuses with unsupported primary Unknown at 87, and the outside-import witness reports no Go findings at its raw repository path. The parser reproducer in the rule sibling's gaps/cases.json and gaps/internal-recovery.a preserves source, logical filename and options, including the literal backslash-n after the import. TestSharedNexusParserGaps again observes actual Go findings but exit 70 from source Node, emitted JS and sanitized native for internal recovery (Unknown 87), project recovery (Unknown 84), and outside top-level await (expected semicolon 21). Those refusals are not byte-equal recovered findings. The outside-import rule requires a libraries/nexus path; its shared raw witness has a different path. Inspection suggests the following project witness also needs libraryDirectory options, but the runner stops first at the outside-import witness, so that later point remains an inference. Owned Boolean/project adapters still decode field 5 into upstream Go options types; no owned adapter trips the options guard. The guard is unchanged. The landing cap is not satisfied, so no new claim or helper-exhaustion claim is made.

Every owned control is freshly rerun: seven full-rule semantic mutations; fourteen first-kind-to-valid-Unknown listener declarations; one resolved-path mutation; three Tailwind decisions; five Next decisions; twenty-four helper controls. Of the helper controls, twenty-three are actual-Go semantic comparisons (including allocation aliasing and the number-counter approximation at 2^53), and one is the separate unsupported-separator guard refusal comparison. Nil-precondition refusals are separate assertions, not mutant credit. All counted mutations compile and run before their stated comparisons catch them; compilation/sanitizer failure never counts as a semantic catch. Full names and catcher logs follow.

The existing normalizeValueFunctionArgument and segment each remove a prerequisite from better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Seven CFG helpers supply prerequisites for array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Zero consumers lose their final blocker. Exact nextBuildCount still fails before backend emission because stage 0 cannot lower a function returning bigint. Eight JSX/Tailwind kernels retain independently checked decision logic without complete extraction, utility evaluation and finding adapters. No new throughput benchmark is claimed; historical per-rule findings/s remain in AREA-LANDING.md.

Commands, after sourcing /workspace/adamic-tools/env.sh, redirect test output directly to files. Helpers run on their branch. Rule and required commands run in /workspace/scratch/wave12-area-validation, detached at d3a37422, with only the six owned rule directories copied in and existing owned Go validation artifacts enabled. No shared compatibility patch is present. The validation checkout uses a physical Go cohere clone at 715ba94f3608a6500086b1076ce5cb7e51b836db and its TypeScript checkout at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a.

```sh
GOCACHE=/workspace/scratch/wave12-required-go-cache GOFLAGS=-p=2 bash cloud/setup.sh > /tmp/wave12-d3-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/scratch/wave12-b469-gate
export GOCACHE=/workspace/scratch/wave12-required-go-cache
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-d3-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-d3-helper-vet.log 2>&1
go test -p 1 ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... -count=1 -v -timeout=20m > /tmp/wave12-d3-owned.log 2>&1
go vet ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import/... > /tmp/wave12-d3-owned-vet.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/scratch/wave12-landing-rules/stage1 go test ./stage1/cohere/lint -run '^(TestLandingExistingCandidates|TestRulesAgree|TestOwnedWitnesses|TestWave12OptionsAndShapes)$' -count=1 -v -timeout=30m > /tmp/wave12-d3-rules.log 2>&1
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
go test -p 2 ./stage1/typescript/parser ./stage1/cohere/css ./stage1/cohere/graphql ./stage1/cohere/mediaquery ./stage1/cohere/selector ./stage1/cohere/values ./stage1/cohere/json ./stage1/cohere/gitignore ./stage1/cohere/lint -run '^(TestCompilerExpressionsAgree|TestWholeCompilerAgrees|TestThePortParsesAsGoCohereDoes|TestCSSPrinterAgreesWithGo|TestCSSPrinterBoundaryProofs|TestUpstreamNumericSeparatorGap|TestUpstreamRepositoryCorpusParity|TestExternalComparisonCatchesThreePrinterMutants|TestPortMatchesGoCohere|TestAdditionalJSONBoundaries|TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars|TestThePortAnswersAsGoCohereAndGitDo|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=30m > /tmp/wave12-d3-required.log 2>&1
```

Completed mandatory checks and package results (repeated test names belong to their respective packages in log order):

```
--- PASS: TestCompilerExpressionsAgree (68.17s)
--- PASS: TestWholeCompilerAgrees (91.33s)
ok  	github.com/system-inc/adamic/stage1/typescript/parser	159.559s
--- PASS: TestThePortParsesAsGoCohereDoes (120.20s)
--- PASS: TestCSSPrinterAgreesWithGo (534.07s)
--- PASS: TestCSSPrinterBoundaryProofs (4.00s)
ok  	github.com/system-inc/adamic/stage1/cohere/css	658.328s
--- PASS: TestThePortParsesAsGoCohereDoes (2.84s)
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	59.290s
--- PASS: TestThePortParsesAsGoCohereDoes (2.49s)
ok  	github.com/system-inc/adamic/stage1/cohere/mediaquery	14.895s
--- PASS: TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars (0.00s)
--- PASS: TestThePortParsesAsGoCohereDoes (22.77s)
ok  	github.com/system-inc/adamic/stage1/cohere/selector	49.090s
--- PASS: TestThePortParsesAsGoCohereDoes (1.82s)
ok  	github.com/system-inc/adamic/stage1/cohere/values	50.022s
--- PASS: TestUpstreamNumericSeparatorGap (4.30s)
--- PASS: TestExternalComparisonCatchesThreePrinterMutants (5.03s)
--- PASS: TestAdditionalJSONBoundaries (29.68s)
--- PASS: TestUpstreamRepositoryCorpusParity (134.93s)
--- PASS: TestPortMatchesGoCohere (1018.30s)
ok  	github.com/system-inc/adamic/stage1/cohere/json	1018.370s
--- PASS: TestThePortAnswersAsGoCohereAndGitDo (8.91s)
ok  	github.com/system-inc/adamic/stage1/cohere/gitignore	111.057s
--- PASS: TestCompilerAndStage1Agree (166.66s)
ok  	github.com/system-inc/adamic/stage1/cohere/lint	166.713s
```

Fresh full-rule semantic mutants, each caught only by Go finding/fix byte comparison on source Node, emitted JavaScript and sanitized native:

- alias_prefix_boundary
- boolean_outcome_companion_ignored
- last_internal_owner
- next_document_import_path_exemption_ignored
- orm_bare_decorator_ignored
- role-suffix regex disabled
- whitelist_widened

Fresh named owned subtests (allocation aliasing, counter approximation and separator guard observations are in parent logs):

```
TestNamedListenerDeclarations/base-correctness-require-orm-column-declare (1.50s)
TestNamedListenerDeclarations/nexus-consistency-no-boolean-outcome (1.44s)
TestNamedListenerDeclarations/next-google-font-preconnect (1.18s)
TestNamedListenerDeclarations/next-inline-script-id (1.10s)
TestNamedListenerDeclarations/next-next-script-for-ga (1.49s)
TestNamedListenerDeclarations/next-no-before-interactive-script-outside-document (1.53s)
TestNamedListenerDeclarations/next-no-css-tags (1.08s)
TestNamedListenerDeclarations/next-no-document-import-in-page (1.42s)
TestNamedListenerDeclarations/better-tailwindcss-enforce-shorthand-classes (1.49s)
TestNamedListenerDeclarations/better-tailwindcss-no-concatenated-classes (0.93s)
TestNamedListenerDeclarations/better-tailwindcss-no-conflicting-classes (0.85s)
TestNamedListenerDeclarations/nexus-boundary-no-internal-import (1.43s)
TestNamedListenerDeclarations/nexus-boundary-no-nexus-outside-import (1.14s)
TestNamedListenerDeclarations/nexus-boundary-no-project-import (1.33s)
TestTailwindDecisions/better-tailwindcss-enforce-shorthand-classes (9.09s)
TestTailwindDecisions/better-tailwindcss-no-concatenated-classes (6.20s)
TestTailwindDecisions/better-tailwindcss-no-conflicting-classes (13.92s)
TestExtractedDecisions/next-google-font-preconnect (1.26s)
TestExtractedDecisions/next-no-css-tags (1.77s)
TestExtractedDecisions/next-inline-script-id (1.35s)
TestExtractedDecisions/next-next-script-for-ga (0.75s)
TestExtractedDecisions/next-no-before-interactive-script-outside-document (0.95s)
TestBigIntNormalizationMutants/hex-digit-value (8.35s)
TestBigIntNormalizationMutants/negative-zero (4.98s)
TestBigIntNormalizationMutants/invalid-fallback (3.97s)
TestBreakableStatementMatchesGo/missing-switch (9.07s)
TestBreakableStatementMatchesGo/nil-loop (8.49s)
TestDecoratorsMatchesGo/drop-first-decorator (3.96s)
TestDecoratorsMatchesGo/duplicate-expression (5.35s)
TestLabelsOfMatchesGo/omit-outer-labels (19.02s)
TestLabelsOfMatchesGo/ignore-child-identity (25.62s)
TestLabelsOfMatchesGo/wrong-kind (32.74s)
TestNormalizationMutants/namespace-suffix (11.71s)
TestNormalizationMutants/newline-crossing (4.56s)
TestNormalizationMutants/nested-key-join (5.35s)
TestSegmentMutants/drop-final-empty (12.85s)
TestSegmentMutants/pop-unmatched-closer (7.63s)
TestSegmentMutants/escape-disabled (9.99s)
TestAlwaysTruthyTestMatchesGo/zero-bigint (11.20s)
TestAlwaysTruthyTestMatchesGo/skip-parentheses (7.38s)
TestAlwaysTruthyTestMatchesGo/empty-string (9.53s)
TestTypeParametersMatchesGo/drop-constraint (3.44s)
TestTypeParametersMatchesGo/swap-constraint-default (3.28s)
TestLandingExistingCandidates/base-correctness-require-orm-column-declare (44.65s)
TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome (84.12s)
TestLandingExistingCandidates/next-no-document-import-in-page (38.80s)
TestLandingExistingCandidates/nexus-boundary-no-internal-import (36.63s)
TestLandingExistingCandidates/nexus-boundary-no-nexus-outside-import (32.42s)
TestLandingExistingCandidates/nexus-boundary-no-project-import (36.76s)
TestLandingExistingCandidates/nexus-consistency-no-boolean-outcome#01 (36.06s)
```

Current-base named external-library mutation-proof subtests; exact catcher/backend observations are in d3-required.log, without claiming emitted-JavaScript coverage for every external check:

```
TestCSSPrinterAgreesWithGo/default/catches_declarations_lose_their_semicolons (51.14s)
TestCSSPrinterAgreesWithGo/default/catches_groups_ignore_the_remaining_width (65.67s)
TestCSSPrinterAgreesWithGo/default/catches_rule_bodies_lose_indentation (41.34s)
TestCSSPrinterAgreesWithGo/narrow/catches_declarations_lose_their_semicolons (28.47s)
TestCSSPrinterAgreesWithGo/narrow/catches_groups_ignore_the_remaining_width (39.66s)
TestCSSPrinterAgreesWithGo/narrow/catches_rule_bodies_lose_indentation (27.62s)
TestThePortAnswersAsGoCohereAndGitDo/catches_%q_in_uppercase_hexadecimal (39.86s)
TestThePortAnswersAsGoCohereAndGitDo/catches_**_matching_one_segment_too_few (48.73s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_[[:upper:]]_one_letter_short (45.93s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_[[:xdigit:]]_taking_G (27.71s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_a_path_never_cleaned (36.57s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_a_trailing_tab_trimmed_as_a_space (33.77s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_an_escaped_range_end_read_as_the_backslash (23.94s)
TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower (46.00s)
TestThePortAnswersAsGoCohereAndGitDo/catches_a_linked_exclude_file_refused (29.56s)
TestThePortAnswersAsGoCohereAndGitDo/catches_entering_from_below_the_root_forgetting_the_files_above (28.38s)
TestThePortAnswersAsGoCohereAndGitDo/catches_precedence_read_from_the_root_down (42.07s)
TestThePortParsesAsGoCohereDoes/catches_--_a_word_only_before_more_text (10.51s)
TestThePortParsesAsGoCohereDoes/catches_?_not_in_a_unicode_range (14.98s)
TestThePortParsesAsGoCohereDoes/catches_DEL_written_unescaped (11.58s)
TestThePortParsesAsGoCohereDoes/catches_DEL_written_unescaped (5.08s)
TestThePortParsesAsGoCohereDoes/catches_FRAGMENT_VARIABLE_DEFINITION_no_location (16.04s)
TestThePortParsesAsGoCohereDoes/catches_U+001F_written_unescaped (12.30s)
TestThePortParsesAsGoCohereDoes/catches_U+001F_written_unescaped (5.73s)
TestThePortParsesAsGoCohereDoes/catches_U+2000_not_whitespace (5.16s)
TestThePortParsesAsGoCohereDoes/catches_U+2029_not_whitespace (5.57s)
TestThePortParsesAsGoCohereDoes/catches_U+DFFF_no_trailing_surrogate (16.77s)
TestThePortParsesAsGoCohereDoes/catches_U+FEFF_not_whitespace (4.67s)
TestThePortParsesAsGoCohereDoes/catches_a_CRLF_counted_as_two_lines (15.43s)
TestThePortParsesAsGoCohereDoes/catches_a_G_in_a_hex_escape (10.10s)
TestThePortParsesAsGoCohereDoes/catches_a_[_in_a_name (14.82s)
TestThePortParsesAsGoCohereDoes/catches_a_byte_order_mark_not_ignored (11.35s)
TestThePortParsesAsGoCohereDoes/catches_a_feature_value's_sourceIndex_not_counting_the_colon (4.32s)
TestThePortParsesAsGoCohereDoes/catches_a_leading_zero's_error_at_the_number's_start (16.89s)
TestThePortParsesAsGoCohereDoes/catches_a_lone_surrogate_written_as_itself (14.35s)
TestThePortParsesAsGoCohereDoes/catches_a_merged_word_uses_its_first_token_position (20.67s)
TestThePortParsesAsGoCohereDoes/catches_a_no-break_space_ignored (14.47s)
TestThePortParsesAsGoCohereDoes/catches_a_number's_unit_taken_off_its_end (9.49s)
TestThePortParsesAsGoCohereDoes/catches_a_quote_ignores_escape_parity (16.21s)
TestThePortParsesAsGoCohereDoes/catches_a_space_compared_with_the_comma_token's_kind (8.57s)
TestThePortParsesAsGoCohereDoes/catches_a_spread's_absent_arguments_written (15.20s)
TestThePortParsesAsGoCohereDoes/catches_a_word_ending_at_= (10.13s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum's_digits_stopping_at_8 (10.40s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum's_lowercase_stopping_at_y (12.70s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum's_uppercase_stopping_at_Y (8.66s)
TestThePortParsesAsGoCohereDoes/catches_alphaNum_taking__ (11.63s)
TestThePortParsesAsGoCohereDoes/catches_an_escaped_quote_closing_a_string (3.53s)
TestThePortParsesAsGoCohereDoes/catches_an_escaped_triple_quote_keeping_its_backslash (13.51s)
TestThePortParsesAsGoCohereDoes/catches_an_invalid_escape_read_as_nothing (14.85s)
TestThePortParsesAsGoCohereDoes/catches_an_operator's_sourceIndex_its_index (12.38s)
TestThePortParsesAsGoCohereDoes/catches_any_closing_quote_ending_a_string (3.85s)
TestThePortParsesAsGoCohereDoes/catches_closing_braces_no_longer_include_the_closing_byte (21.86s)
TestThePortParsesAsGoCohereDoes/catches_comments_always_disappear_from_values (23.27s)
TestThePortParsesAsGoCohereDoes/catches_custom_properties_lose_their_block_values (28.95s)
TestThePortParsesAsGoCohereDoes/catches_isColor's_digits_stopping_at_8 (13.79s)
TestThePortParsesAsGoCohereDoes/catches_next_starting_at_0,_not_NaN (12.16s)
TestThePortParsesAsGoCohereDoes/catches_no_directives_an_empty_list (17.94s)
TestThePortParsesAsGoCohereDoes/catches_the_first_line's_indent_common (14.98s)
TestThePortParsesAsGoCohereDoes/catches_the_sign_bit_not_reached (11.20s)
TestThePortParsesAsGoCohereDoes/catches_the_strict_calc_check_ignoring_case (7.23s)
TestThePortParsesAsGoCohereDoes/catches_the_suffix_operator_loses_|= (13.19s)
TestThePortParsesAsGoCohereDoes/catches_the_tokenizer's_url_argument_ignoring_case (8.03s)
TestThePortParsesAsGoCohereDoes/catches_the_unexpected_token_not_the_one_named (11.34s)
TestThePortParsesAsGoCohereDoes/catches_the_url_argument_check_ignoring_case (13.30s)
TestThePortParsesAsGoCohereDoes/catches_the_word_after_an_expression_typed_as_a_media_type (5.01s)
TestThePortParsesAsGoCohereDoes/catches_trim_as_trimStart (6.05s)
TestThePortParsesAsGoCohereDoes/catches_unicodeRange's_digits_stopping_at_8 (13.84s)
```
