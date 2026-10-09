Defended in the complete reached-main replay: D1 fails only TestCSSThroughput.
All 388 parser shards and 16 former-subsumer composition shards pass, plus related builders and witnesses.
Keep for repeat-count consistency; speed is logged, not asserted.

# CSS throughput defense

CODE UNDER TEST: the TypeScript raw CSS port, including main.ts count/repeat driver and Input.position. Native and source Node are two executions of that implementation. ORACLE: independent pinned PostCSS 8.5.16/postcss-scss 4.0.9 aggregate counts. Go cohere enumerates the cases; the composition family independently compares its complete trees. No test, harness or oracle was edited.

## Starting point and disk

Fresh origin/main 955e3eb92b9cd04aca420974d1006048b4615d51, on test-defend/stage1-cohere-css-composition_shards. Audit branch fetched with full destination refspec. Read REPORT.md, CODE-AND-ORACLE.md and rows.json, including the audit's bounded verdicts and weaknesses; copies are retained. The audit's 95 requested functions were a slice of its 514 listed package functions. Current go test -list also has 514 top-level functions, with no added or vanished names. scope.json records them all.

First command was df -h /tmp /workspace. Named previous-unit scratch/cache directories were removed only under /tmp. Available /tmp space improved from 1.8G to 8.6G; its entire capacity is 8.8G, so 15G cannot be attained. /workspace remained 14G free. Repository and tools were not removed. DISK.md records exact before/after figures and removal scope. No disk error occurred in baseline or replay.

Warm toolchain worked, so setup was skipped. Ran npm ci in stage3/api and installed pinned PostCSS, postcss-scss and Prettier in /tmp/defend-css/library. Fixture tree is system-inc/prettier commit cb4b33fba24a8428d00e54be85fc886288a374ea with the documented sparse paths. environment.json records all opt-ins and tool versions; nproc=5.

## Clean runs and coverage

Whole-package baseline with ADAMIC_CSS_BENCH=1 and external library/fixtures enabled timed out at 90.017 binary seconds, without an assertion failure. Six rows completed; throughput itself passed before timeout. Later checks are unknown from that run. This is over budget, not a red baseline.

Isolated clean throughput and composition-family runs used -coverpkg=./internal/lower,./internal/native,./internal/javascript and -coverprofile, with -json -count=1 -timeout 90s inside timeout 120. They passed in 64.391s and 47.423s binary time respectively; throughput instrumentation included supplemental V8 coverage. Profiles are preserved.

The Go profile has 4426 target-only compiler blocks, but composition reused cached build products. Those blocks do not prove exclusive TypeScript behavior. Supplemental NODE_V8_COVERAGE records real port execution, with complete compressed records, selected port records, and function inventories saved. Throughput executes main.ts; its repeat branch at original line 59 is taken, and the transformed '? 10' interval [1397,1401] has count 11 per source Node invocation. Composition executes compose_main.ts instead; its '? 10' interval [1403,1407] has count zero. It takes the one-round branch. Throughput also observes count-output lines 74..75 rather than complete trees. coverage-diff.json separates compiler cache effects from these runtime differences.

## Aimed attempts

C1, input.ts:63, removes the binary-search midpoint policy by advancing to low+1. This is the requested answer-preserving cost attempt: searches become linear in the line count. Native throughput and all 16 composition shards pass. A separate source probe compared positions at every offset of a 20,000-unit, 10,000-line string, then queried the final position 1000 times. All answers agree; checksum is 10001000 on both. Binary lookup took 121814ns, linear lookup 17935756ns, about 147 times slower in that specific probe. This is not claimed as the throughput corpus's slowdown factor. C1 survived the two-row bounded replay; kills elsewhere are unknown. The cost-row rule explicitly allows removing a fast path. No uniqueness verdict rests on C1.

D1, main.ts:59, changes the repeat bound from 10 to 9, an admissible off-by-one. Native and source Node agree with each other on the wrong aggregate. Independent PostCSS catches the defect: 54810 of 245960 stylesheets parsed, 129930 nodes versus 49329 of 221364 stylesheets parsed, 116937 nodes. Exact failure is css_test.go:465 in D1-throughput.log. All 16 former-subsumer composition shards pass. Raw-parser agreement/witness shards invoke main.ts with only the cases argument and do not take its repeat branch; all related product constructors and construction witnesses are replayed too. See main-callers.txt for the caller search, results.json for exact commands, and passed-rows.json for completed passes.

D2 (dropping child-count accumulation) was planned but not planted: D1 provided the aimed distinction, so an additional similar count mutant was unnecessary. There are two executed attempts and two standalone diffs.

Each executed mutant has its own ADAMIC_BUILD_CACHE_DIR under /tmp/defend-css/cache. Each compiles with the actual native port tool during replay; C1 completed native runtime checks, while D1 failed only after native build and native/Node count execution. Sanitized parser-product builds also validate D1. Standalone diffs apply to starting main. Production files are restored after replay. No mutant switch, oracle edit or test edit was needed.

## Budget narrowing and limits

The clean 388-member raw-parser family also exceeded 90 seconds. A started broad D1 retry was stopped at 42.723 command seconds once that baseline result was incorporated. Its partial rows are not used to claim a complete matrix. The family was replaced by four batches of at most 100 top-level tests, first clean and then D1. Each uses timeout 120 go test -json -count=1 -timeout 90s. Batch completion, failures and skips are explicitly recorded. This costs additional repeated setup but avoids treating an unfinished family as green.

The D1 verdict is based on the reached-main matrix plus the named former subsumer, permitted by the big-package rule. Other package rows are not claimed as experimentally passed. Those rows use other entry programs; static caller search supplies the narrowing rationale. C1's bound is smaller because it is a cost attempt, not the unique-kill basis. Complete package replay, unrelated printer/profile tests and repo-wide replay were not performed. Darwin-only product skipped on Linux; optional profiler/artifact jobs were not enabled. The complete passed and skipped lists are machine readable.

## Findings and brief costs

The test is worth keeping for count/repeat consistency. Its name also promises throughput, but assertions enforce no elapsed-time threshold or speed ratio: css_test.go:467 only logs duration and stylesheets/s. The outer Go deadline is shared harness protection, not a row-specific performance contract. C1 gives an answer-preserving slow behavior that the test does not reject. Aggregate counts are weaker than a complete tree comparison and use native output as the initial expected answer; independent PostCSS must be enabled for D1 to be caught.

The instruction to reach 15G free is impossible on this 8.8G /tmp mount. Cleaning prior scratch avoided disk pressure anyway. A fresh-main baseline differs from the audit commit; names were checked rather than copied. The audit's requested slice count is not the package count. Go coverage does not observe the TypeScript implementation and is sensitive to product-cache hits, requiring supplemental V8 records. The full package and full raw-parser family exceed 90 seconds; sequential batches are necessary. A broad D1 family retry was briefly started before narrowing, explicitly stopped, and given no verdict weight. Native rebuild, corpus enumeration and test execution are combined in command timings because the harness does not expose a complete separate rebuild clock. Bounded replay is allowed by the brief, but its strict package-unique wording requires stating the reach scope rather than implying all 514 functions ran in one binary.

## Completed result

D1: 416 top-level functions passed, one failed, one Darwin-only product skipped. The 388 parser shards were fully replayed clean and mutated in four batches. 96 other top-level functions were not replayed and are not claimed as experimentally passed. The reached-entry narrowing is documented by main-callers.txt. This is a bounded defense, not repo-wide uniqueness.

Exact failure: css_test.go:465: PostCSS Node checksum "54810 of 245960 stylesheets parsed, 129930 nodes\n" differs from "49329 of 221364 stylesheets parsed, 116937 nodes\n"

Both standalone diffs were git apply --check validated against unchanged starting main. No production source or test diff remains.

Total controller command wall seconds (clean reached runs and mutant runs, including cooked and stopped attempts): 670.624. Clean instrumented coverage binary time 111.814s and whole baseline 90.017s are additional. API npm ci reported 651ms and independent library install 1s. Warm setup skipped. Native rebuild and execution are combined in the recorded per-run wall times.
