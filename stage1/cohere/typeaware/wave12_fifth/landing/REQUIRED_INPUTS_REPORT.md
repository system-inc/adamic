Built: provisioned exact external inputs and ran the formerly input-dependent stage-1 correctness comparisons without skips; no rule/shared-source changes or new claims.
Commits: tested existing wave-12 head 6e481d31971e2d3efa02423fa13919a39952538a; fetched main 39638d9e278d38bb5aeae887f46d55a70e47aaad and area d65a8f931c98655936ae04c6899f38f14862b73e were unchanged and already ancestors.
Checks: all nine selected packages pass, 15 parent checks and 93 total checks; Go command exit 0, artifact checker exit 0, zero skips/failures. Shared compiler/stage1 lint compares 399 files.
Mutants: existing scanner/parser/CSS/GraphQL/selector/values/mediaquery/JSON byte mutations caught; a real omitted-library input produces a skip which the no-skip checker rejects with exit 1.
Uncovered: this is focused required-input coverage, not the full repository gate; three active source ports still lack shared checker integration and three HIR claims remain parked. Shared wave reservation discovery remains unmet.

Only codex/typeaware-wave-12 is pushed. git rebase origin/area/stage1-lint reports up to date. All 160 previously tested runtime/rule source hashes remain unchanged; the preceding twelve-rule byte/sanitizer/lifetime results and native-versus-Go timings still apply without a code rebase. No protected compiler file, shared harness, check, input requirement, finding model, registration generator or submodule pin is edited. The only new executable file is an owned Python artifact checker under wave12_fifth/landing; it does not alter the shared gate. No more rules were reserved.

Pinned inputs were installed in /workspace/wave-12/required-inputs with npm install --prefix /workspace/wave-12/required-inputs --ignore-scripts --no-audit --no-fund postcss@8.5.16 postcss-scss@4.0.9 postcss-selector-parser@2.2.3 postcss-values-parser@2.0.1 postcss-media-query-parser@0.2.3 graphql@17.0.2 prettier@3.9.6. PASS: thirteen dependency packages installed in three seconds. Scripts are disabled; exact package/lock manifests and installation log are preserved. The TypeScript checkout is the previously verified pinned 050880ce59e30b356b686bd3144efe24f875ebc8 at /workspace/wave-12/typescript. This corpus differs from the frozen subset used in rule timing; neither frozen rule manifest was expanded.

Commands source /workspace/adamic-tools/env.sh. Toolchain is unchanged: Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc 5. The preceding runtime unit's setup passed in 182 seconds; setup was not repeated here because refs, tools and source did not change. Output goes directly to log files.

The positive Go command sets ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/typescript, and ADAMIC_CSS_LIBRARY, ADAMIC_CSS_PRINTER_LIBRARY, ADAMIC_GRAPHQL_LIBRARY, ADAMIC_SELECTOR_LIBRARY, ADAMIC_VALUES_LIBRARY, ADAMIC_MEDIA_QUERY_LIBRARY and ADAMIC_JSON_PRETTIER all to /workspace/wave-12/required-inputs. It runs go test -json -count=1 -timeout 30m -p 1 -parallel 2 on ./stage1/typescript/scanner, ./stage1/typescript/parser, ./stage1/cohere/css, ./stage1/cohere/graphql, ./stage1/cohere/selector, ./stage1/cohere/values, ./stage1/cohere/mediaquery, ./stage1/cohere/json and ./stage1/cohere/lint with this filter:

^Test(ScannerAgreesWithTypescriptGo|CompilerExpressionsAgree|WholeCompilerAgrees|CompilerAndStage1Agree|ThePortParsesAsGoCohereDoes|CSSPrinterAgreesWithGo|CSSPrinterBoundaryProofs|TheLibraryDoesNotReturnOnUnconsumedNamespaceBars|UpstreamNumericSeparatorGap|UpstreamRepositoryCorpusParity|ExternalComparisonCatchesThreePrinterMutants)$

All selected correctness parents and their children run, including native/source Node/emitted JavaScript/sanitizer comparisons where the existing test provides them. No benchmark/profile skip is selected. The fifteen parents include the external-input subcomparisons; no assertion is made that the announcement's seventeen checks are seventeen top-level Go functions. All input-dependent correctness sites found on this fetched branch are covered, with exact test events retained.

Observed package PASS times:

| Package | Seconds |
| --- | ---: |
| scanner | 36.176 |
| parser | 53.100 |
| css | 324.674 |
| graphql | 39.101 |
| selector | 31.290 |
| values | 40.338 |
| mediaquery | 14.100 |
| json | 76.268 |
| lint | 124.136 |

Scanner, compiler expressions and whole-compiler comparisons pass against real pinned TypeScript inputs. PostCSS parsing passes. CSS printing in both default and narrow modes agrees on 4,463 formats and 19,013 shared refusals; the same twenty already-recorded discrepancy occurrences are observed (BOM ten, CR normalization two, whitespace fast path two, YAML embedding six). Those are existing exact assertions, not relaxed tolerances. GraphQL passes 3,426 texts across Go, native, source Node, emitted JavaScript and graphql-js. Selector parsing and its four upstream nontermination witnesses run with the library present. Values/media-query parsing passes. JSON auditing observes exactly the nine already-recorded upstream differences in 1,752 texts, and its three printer mutants are caught. The shared lint compiler/stage1 comparison loads 399 files and passes. Detailed bytes, observations and all mutant labels are in comparisons.txt, summary.json and the complete event log.

check_required_inputs.py rejects any skip or failure event, requires all nine package results and all fifteen named correctness parents, and reports 93 passed checks with zero skips. Its independent negative proof runs env -u ADAMIC_VALUES_LIBRARY go test -json -count=1 -timeout 30m ./stage1/cohere/values -run '^TestThePortParsesAsGoCohereDoes$/^as_postcss-values-parser$'. This Go command exits 0 but emits a real skip for the absent input. The owned checker exits 1 with required correctness check skip for that exact child. This negative run is not a gate pass, does not relax any check and is kept separately. It demonstrates why Go's successful process exit alone is insufficient under the new no-skip requirement.

No positive run failed or skipped. No fix or unrelated-scope reproducer was needed. The shared lint-wave-check.py still cannot discover this legacy type-aware reservation; this focused compiler comparison does not imply a green receipt for that separate gate. The three active JSX source ports remain incomplete for the measured shared checker context/runtime/profile gap, while their eighteen prepared findings and 6,202 quoting cases retain the preceding four-backend evidence. HIR/SSA/capture-dependent claims remain parked. Further source work is blocked outside the owned rule directories, so no new claim is taken.

The full repository gate and performance/profile tests were not run. No new timing claim is made; latest source-rule native timings remain 6.76-9.82 times Go on compiler roots and 2.14-4.17 times on repository roots, as recorded in RUNTIME_REFRESH_REPORT.md. Complete positive and negative logs, lockfiles, command exits, mutation observations and input-checker hash are in evidence/required-inputs. Report/checker publication does not change tested rule or runtime code.
