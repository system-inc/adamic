Closed-regex has an exclusive catch in the completed matrix, but package uniqueness remains unknown.
Throughput and boundary rows are not defended after three production attempts each.
All tests retained unchanged; standalone production diffs and bounded evidence are saved.

Starting origin/main: 955e3eb92b9cd04aca420974d1006048b4615d51. Scope came from go test -list: 514 current top-level tests; all three requested names remain present in print_test.go. The audit REPORT.md, results.json and plan.json were read from the fully fetched audit ref before mutation. This session's rows.json is authoritative for its observations. Other defenders' branch evidence is preserved separately and supports no session claim here.

## Code and oracle

The native RegExp test implementation and CSS production Input, parser diagnostics, formatter and document printer were code under test. Go cohere, Prettier fork/npm 3.9.6 and Node were unchanged external oracles. The closed-regex row additionally requires the self-written output 2\nOk\n. No test was deleted, rewritten or weakened.

## Coverage and differences

See coverage-method.md, per-row logs/profiles, raw compressed V8 traces, coverage-exclusive-lines.json and coverage-compiler-exclusive-blocks.json. Go cannot instrument TypeScript or embedded C. Supplementary lower-package coverage therefore describes compiler preparation, not production port coverage. The boundary row is Node-only and snapshots use prebuilt products, so both compiler profiles are zero. Instrumented throughput cooked at 90 seconds without a final Go profile; its partial V8 traces include Node count executions. Uninstrumented throughput then passed in 79.855 seconds.

V8 actual execution removes comments while source-map generation retains them. Source-map output was aligned with identical comment-free generated lines before mapping to original TypeScript; unmatched fragments support no claim. Initial incorrect mappings were corrected. Boundary has 144 mapped exclusive line leads versus the six boolean cases, including BOM/front-matter/diagnostic paths. Throughput has count-summary lines 76-78 exclusive versus snapshots. Closed-regex has no exclusive TypeScript port lines here: its lead is native stateful global RegExp.test versus formatter non-global tests and global exec.

## Production attempts

R1 offsets the fast stateful RegExp.test starting position by one. Closed-regex native output becomes 0\nOk\n; six neighboring tests and the freshly built full snapshot comparison pass. It is a positive defense candidate, but no package-wide unique kill is claimed: 505 current top-level tests are unknown under R1.

B1 changes BOM recognition; B2 off-by-ones the unknown-word diagnostic end; B3 changes nonempty front-matter refusal from YAML to TOML. Each fails boundary and snapshots while its old optional-boolean subsumer passes. Thus the audit's named subsumer does not cover these attempts, but another row does. This is not defended after three aimed attempts; it does not establish universal redundancy.

C1 disables document width caching. The isolated probe returns 100 in both versions, with stringWidth called once clean and twenty times mutated. All completed agreement checks and throughput pass; throughput completes in 76.863 seconds. No speed threshold is asserted, so more work is accepted. D2 changes the exact-width fits bound, failing throughput and snapshots together. D3 substitutes same-length colons for declaration semicolons: full-output checks fail while throughput passes in 81.836 seconds with its original count/length checksum. These three production attempts do not defend throughput. D3 is a concrete equal-length wrong-output witness, not an equivalent mutant.

Two initially selected count-driver edits were scope mistakes. They are quarantined under excluded-driver-evidence, with diff text marked excluded. H1 changes measurement accumulation; H2 changes count-option dispatch. H2 preparation was interrupted deliberately. They are harness edits and contribute no production kill, unique kill, or verdict. Replacement D2/D3 mutate actual formatter logic. No oracle was edited.

## Matrix and builds

Each production mutant has its own ADAMIC_BUILD_CACHE_DIR. All seven standalone unified diffs apply to the starting origin/main and produced successful native printer, counted and profiled builds. See production-validations.json and each artifacts log. Source changes were restored after every run and git diff was clean before publication. matrix.json enumerates every current test as observed pass/fail or unknown. Timeout-only events never count as mutant kills. Per-mutant pass lists and exact failing lines are there. All seven small regression groups completed without panic aborts.

Selected regression rows: TestClosedPrinterRegexGap, TestCSSPrinterBoundaryProofs, TestOptionalBooleanPrinterMatchesGo, TestSharedSliceAppendAgreesWithNode, TestClosedEmptyArrayUnionGap, TestClosedParserRegexGap, TestClosedOptionalBooleanConditionGap. TestCSSProfileArtifacts prepares fresh products; TestCSSProfileSnapshotsAgree compares both default and narrow outputs. Throughput was run for C1/D2/D3. For construction-only artifact rows, success is build evidence, not semantic uniqueness evidence.

The whole clean package cooked at 90.021 seconds with no preceding assertion failure. A wider clean group of sixteen current printer agreement shards also cooked at 90.030 seconds; its unfinished rows remain unknown. The production matrix was narrowed to the complete observed rows above. Parser, printer and composition families outside that matrix, opt-ins outside this unit, platform-only members and repository-wide replay remain uncovered. All selected rows ran without skips. A missing-library-directory failure was configuration error, corrected before mutation; the boundary rerun passed.

## Owner findings and limits

Throughput's name implies cost measurement; it logs timings but asserts no speed, budget or comparative-performance threshold. Its runtime oracle checks only aggregate counts and lengths and accepts D3's wrong formatted text. Boundary's four complete answers are genuinely checked against Go; no name/assertion mismatch was found. Closed-regex checks constructed RegExp.test and mediaquery.parse.kind, but never inspects recursive tree contents or formatted CSS. Keep it pending wider replay of R1; inability to settle uniqueness is not a deletion recommendation.

The 15 GB disk threshold is impossible on /tmp, whose total capacity is 8.8 GB. Cleanup changed initial 4.3 GB free to 8.8 GB free; workspace had 16 GB. Protected earlier permission-test scratch was removed after tests ended. Final df is saved. No baseline failed because of disk exhaustion. Repository and tools were never removed.

Automatic approval review rejected concurrent scratch cleanup, overlapping baseline runs, and finalization while it believed a completed runner active. I waited or verified process completion and retried only after the risk ended. No approval remains blocked. Go coverage cannot measure the production languages here, and source-map comment differences required extra analysis. The supplied audit commands were truncated, so full saved artifacts supplied context. Selecting and then excluding driver edits was my scope error, costing several minutes. Concurrent defenders published on the same branch, so this evidence is isolated in a session directory and merged without replacing theirs.

Warm setup skipped, nproc 5. npm ci and the 419 ms Prettier installation are logged. Clean artifact construction took 51.507 binary seconds; clean snapshot comparison took 53.866 seconds. Clean focused binary times were 1.041 seconds closed-regex, 1.295 seconds boundary and 29.318 seconds optional booleans. Production native rebuild phases and command walls are recorded in matrix-runs.json; they include preparation and oracle work, not clang alone. Seven valid mutants used 24 sequential commands. Total session approximately 36 minutes, exceeding the approximate 30-minute port budget because of cooked baselines, repeated native product builds, source-map correction and replacement of excluded driver attempts. No main push or PR.

Production command wall total: 1027.332 seconds. Supplemental command timings are separate in costs.json.

| Mutant | Starting file:line | Change | Observed failed rows |
|---|---|---|---|
| R1 | internal/native/runtime/regexp.c:627 | global fast-test start off by one | TestClosedPrinterRegexGap |
| B1 | stage1/cohere/css/input.ts:20 | change BOM constant | TestCSSPrinterBoundaryProofs, TestCSSProfileSnapshotsAgree |
| B2 | stage1/cohere/css/parser.ts:88 | unknown-word diagnostic end off by one | TestCSSPrinterBoundaryProofs, TestCSSProfileSnapshotsAgree |
| B3 | stage1/cohere/css/print.ts:269 | change front-matter refusal language | TestCSSPrinterBoundaryProofs, TestCSSProfileSnapshotsAgree |
| C1 | stage1/cohere/css/print_doc.ts:55 | disable width cache while preserving answers |  |
| D2 | stage1/cohere/css/print_doc.ts:103 | exact-width fit bound off by one | TestCSSProfileSnapshotsAgree, TestCSSPrinterThroughput |
| D3 | stage1/cohere/css/print.ts:143 | same-length declaration semicolon substitution | TestOptionalBooleanPrinterMatchesGo, TestCSSPrinterBoundaryProofs, TestCSSProfileSnapshotsAgree |
