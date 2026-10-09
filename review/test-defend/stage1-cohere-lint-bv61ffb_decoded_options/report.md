# Defense of TestClosedComparatorGaps

Base: 92d196011e78208b45b3768dcf2c8c88e7ec132d, fetched origin/main. Audit branch: test-audit/stage1-cohere-lint-bv61ffb_decoded_options; prior REPORT.md, rows.json, inventory-menu.txt and menu.jsonl read. No test or oracle changed.

Code under test: Adamic lowering and its native and JavaScript emitters for the numeric logical-or, optional Map-array index and positioned lastIndexOf fixtures. Oracle: Node runs the unchanged TypeScript; handwritten pins require 2, 1, 1. Both backends must print those pins. These assertions check bytes, not only exit codes.

The prior M4 swapped all simple native ternary arms. It killed the target by output mismatch and decoded-options by a missing-options panic. The new defense targets an exclusive production lowering block instead.

## Coverage and semantic difference

Profiles instrument internal/lower, internal/native, internal/javascript. target.cover has count 1 on library_string.go:235.4,242.1 (7 statements); subsumer.cover has count 0. exclusive-blocks.txt records all 130 target-only covered blocks. Positioned lastIndexOf and optional-index paths are exclusive in this measured comparison. The target's string a(a(, needle (, position 1 must return 1. The port's method-signature fixer uses one-argument lastIndexOf on a pre-sliced string, avoiding the two-argument bound. position-callers.txt and compiler-callers.txt preserve source searches. These do not prove global dynamic reachability.

Warm product-cache coverage initially reported zero compiler statements. It was rejected as evidence of exclusive lines. The genuine family profile was collected with a private cache containing only the changed-lowered, native and Go-oracle products, so the original lowering was cold. That run passed in 16.701 seconds. Target profile passed in 4.300 seconds.

## D1

Standalone D1.diff applies to the base. At internal/lower/library_string.go:239, change ir.Add to ir.Subtract in the prefix end. This is a change of an operator constant from the allowed menu, planned before mutation. go vet ./internal/lower/ completed successfully, D1-vet.log has no diagnostics. The natively built mutant ran under sanitizers. No switch or supplemental edit was used.

D1.log is a completed 12-test matrix with a private ADAMIC_BUILD_CACHE_DIR=/tmp/lint-gap-defense/cache/D1. Only TestClosedComparatorGaps failed: closed_gaps_test.go:50: native prints "-1\n", Node "1\n". All six decoded-options shards, Union and all four decoded-options product rows passed; exact names are in D1-matrix.json and rows.json. The other two target subcases passed. The failing native assertion ends the positioned subcase before its JavaScript-backend assertion, so that backend's behavior under D1 was not independently observed.

Verdict: defended within this bounded matrix. This is evidence to keep the row, not proof of package-wide uniqueness. Tests outside the completed matrix remain unknown. Central replay of D1.diff must settle package-wide uniqueness.

## Scope, limits and brief costs

Current go test -list has 6,760 top-level names. Twenty-two names were added and one vanished relative to the audit list; scope-changes.txt lists every name. All added names were included in expanded-regex.txt, with the available pinned TypeScript corpus at /tmp/u157-typescript (commit 050880ce59e30b356b686bd3144efe24f875ebc8). The expanded clean and mutant runs are bounded by 90 seconds and retain incomplete rows as unknown, not passes or production kills. The completed matrix deliberately narrows away the large checker groups after the expanded baseline cooks.

The whole-package clean baseline cooked after 90 seconds in TestLegacyMutants, with no assertion failure recorded before timeout. An initial bounded clean baseline also cooked during product preparation; preparation was narrowed to the existing TestProduct_DecodedOptionsNative. It passed in 11.282 seconds, then the clean bounded matrix passed in 5.951 seconds. No mutation preceded that green bounded baseline.

The brief's coverage command needs cache control: a successful cached run can have zero compiler coverage. Cold setup for both original and built-in-mutant products can exceed the test budget, even though the row's own work is small. The 120-second outer backstop killed the first instrumented family run before the binary consumed 90 seconds because its instrumentation compilation was cold. The retry then hit the binary budget; separate existing product preparation resolved it. Huge package and checker groups prevent interpreting a bounded unique kill as proven package uniqueness. The prior subsumption rested on one shared production mutant, not on the exclusive positioned-search behavior demonstrated here.

Warm env.sh worked, setup.sh skipped; nproc=5. npm ci in stage3/api reported 808 ms. Completed D1 binary matrix took 32.551 seconds, including product rebuilding; its lowered original product took 12.59 seconds according to the build log. All raw timing and timeout logs are retained. No cost threshold is promised by this row's name. Its assertions do check its three closed fixtures; there is no demonstrated name/assertion mismatch for the defended row.

No other packages were tested. No test was deleted, rewritten or weakened. Production source is restored before evidence commit. No main push or pull request.

Expanded baseline: 90.112 seconds; expanded D1: 90.070 seconds, both timeout rather than an observed semantic catch in a new row. The post-timeout narrowed D1 matrix completed in 7.514 seconds with the same sole failing row. Restored target passed in 0.553 seconds. expanded-matrices.json separates observed terminal events from unknown rows. Total session was about 19 minutes, dominated by cold compilation, product preparation and bounded timeouts.
