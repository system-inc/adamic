# Non-null refusal defense

Base: e77a4ae4 (full commit in base.txt). Audit base: cb04a5ac599ede196c934d353e1aa9dec9748867. Audit branch and all report components were fetched with a full refspec and retained as prior-* files. The three assigned names still exist. Current discovery has 198 top-level tests; five were added, none vanished. All five added rows were included in the expanded matrices.

## Code under test and oracle

Adamic's production non-null refusal decision in internal/lower/refusals.go and checkedAssertionSource in non_null.go, entered through Lower; production diagnostic positions in internal/load.Program.Where. Related callers uninitializedInitializer and lazyAssertionInitializer use the same source-policy helper. No TypeScript checker, Node oracle, test, fixture or harness was modified.

All three assigned rows use a self oracle: hand-written Refused kind, What and Fix text. TestNonNullAdamicRefusal additionally checks exact file:line:column and complete diagnostic suffix. The checked-constant and possible-assertion rows do not check source positions. No outside authority supplies their expected refusal policy or text.

## Coverage and input differences

Each assigned row was run alone with -coverpkg=./internal/lower,./internal/load and its own coverage profile. All three passed. Complete profiles, named function percentages, reached-functions.json, and coverage-differences.json are saved. Against their named subsumers, Checked has 106 exclusive covered blocks, Possible 49, and NonNull 74. These blocks include input-dependent preparation and traversal and are leads, not proofs of unique behavior.

The actual defenses are semantic on shared lines:

- Checked supplies a const numeric binding whose assertion can be proven present. D2 incorrectly exempts number-literal operand types from Adamic's refusal policy. It admits this program while the other refusal cases remain refused.
- Possible supplies nullable Map.get call results, both interpolated directly and followed by toString. D3 incorrectly exempts call-result assertions. It admits these programs while identifier, literal-nullish and indexed assertions remain refused.
- NonNull pins the diagnostic position in a nullable parameter return and an undefined initializer. D1 advances the token-position argument by one only for non-null syntax in .a files. What, Fix and error kind remain correct, so rows checking only those fields pass; this row detects both wrong columns.

Mutant guards use semantic node kinds, operand types and the language's source extension, never a test identity or fixture path. D1 is an off-by-one bound change; D2/D3 flip the source-policy decision for general semantic operand classes. The explicit node-kind guard preserves other callers of checkedAssertionSource.

## Baselines and matrix scope

npm ci ran in stage3/api. Warm env.sh worked; setup was skipped. nproc=5, Go 1.27.1, Node 24.19.0.

The clean whole-package baseline exceeded the binary's 90-second budget at 90.207 s. No assertion failure was observed before timeout. It was stopped and narrowed. The 11-row non-null slice passed in 2.550 s. The five added rows passed separately in 53.220 s, including the full review agreement corpus.

The expanded matrix regex was:

^Test(CheckedNonNull.*|ImpossibleNonNull.*|PossibleNonNull.*|MigratedNonNull.*|NonNull.*|ParserNonNull.*|StatementsSmallRulings|FractionalPowersReachRuntime|ReviewPrograms.*)$

That selects 16 top-level rows. matrix.json lists every observed passing and failing row. Each run used ADAMIC_GATE_UNCACHED=1, ADAMIC_DEFENSE_MUTANT=ID, its own ADAMIC_BUILD_CACHE_DIR, timeout 120 outside and -timeout 90s inside. No other package was tested. All three expanded matrices completed without a panic. Each has exactly one failing row and 15 passing rows. These are bounded defenses, not proven whole-package or repo-wide uniqueness. Results outside the 16 rows are unknown. The generic NativeAgreesWithNode fixture row and unrelated Lower callers were not exhaustively replayed. Central replay remains necessary.

Five new rows, all included: TestFractionalPowersReachRuntime, TestReviewProgramsAgreeWithNode, TestReviewProgramsNoLooseFiles, TestReviewProgramsRefuse, TestReviewProgramsSelfTest. The review non-null source inventory is saved; its non-null refusal case remains refused under D2/D3 because the operand is undefined, not a numeric literal type or call.

## Validation and replay

D1.diff, D2.diff and D3.diff are standalone unified diffs against the starting commit, without a selector. Every diff passes git apply --check against restored base files and go vet using overlays that restore BOTH affected files to base plus exactly that one mutation. Commands, exit status and duration are in standalone-checks.json. switch.diff preserves the session's selector source. Apply one standalone diff, omit ADAMIC_DEFENSE_MUTANT, and run the same expanded matrix with a private build cache to replay it. Tests, oracle and fixtures are unchanged. Production switches were restored before the final clean three-row baseline.

## Brief problems, mistakes and costs

1. The requested 15 GB free cannot be achieved on an 8.8 GB /tmp filesystem. It began with 5.3 GB free; named earlier-unit scratch products and caches were removed, raising it to 5.9 GB. /workspace remained 9.9 GB free. The repository and tools were untouched. Unknown bootstrap/corpus directories were retained. No ENOSPC failure occurred.
2. Go coverage measures actual production load/lower code here, but the whole package cannot fit the stated budget. A bounded unique catcher must not be described as proven package-wide uniqueness.
3. The audit's source-extension flip was too broad to distinguish the three negative rows. Production inputs differ semantically despite sharing the refusal line. Position precision is a separate contract from error kind and suggested fix.
4. I initially assumed checkedAssertionSource only received non-null nodes. Its readiness callers also pass ordinary initializers. The initial D2/D3 selectors panicked on an unrelated NumericLiteral in StatementsSmallRulings, aborting the binary before the parallel target rows ran. Those logs are retained as D2-aborted.log and D3-aborted.log; they establish no target kill or uniqueness. The corrected node-kind-guarded variants were rerun, as were all expanded matrices. The mistake was in my scratch mutation, not the baseline code.
5. Two write attempts used the read-only default sandbox and failed before writing; they were rerun with the required filesystem permission. No test output was piped.
6. I read the target test bodies while locating the audit files, before reading the audit report components. All report components were then read before any mutation was planted. This was a deviation from the requested reading order.
7. There are no cost rows or executor twins among the three assigned rows. Their names' refusal promises are checked by their assertions. All three were defended in the bounded matrix, so no unsupported name/assertion promise was found for an undefended row.

## Timing and remaining scope

Setup skipped. npm wall duration was not separately recorded. Whole baseline 90.207 s, clean 11-row baseline 2.550 s, added-row baseline 53.220 s. The first expanded matrix took 75.226 s and the second 74.057 s; timings.json records the third and all other binary durations. Standalone vet checks took approximately 0.363, 0.846 and 0.995 s. Session elapsed approximately 20 minutes, including cleanup, reads, coverage, mutation repairs and publication. Native rebuilds occur inside those binary times and were not separately instrumented. No exhaustive package uniqueness, other-package run, oracle weakening, test rewrite or deletion was performed.
