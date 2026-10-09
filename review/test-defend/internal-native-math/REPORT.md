# Defense of random normalization agreement

Starting origin/main: 7d113268b1903e4e47289f226e94ec2fb609e15b. Branch: test-defend/internal-native-math. nproc: 5.

CODE UNDER TEST: Adamic Unicode normalization in internal/native/runtime/normalize.c, especially long-run combining-class reordering. ORACLE: independent Node String.prototype.normalize output, compared as exact WTF-8 hex for all four forms and 100000 lines. Neither oracle, tests nor harness were mutated.

Verdict: defended in the bounded matrix of all identified current direct normalization callers. D1 is an off-by-one bound in the counting-sort copy at origin/main line 225. Only TestNormalizeRandomMatchesNode failed, reporting 115 of 100000 lines differ. Every one of the 18 TestNormalizeMatchesNode family members passed. TestNormalizeLongMeasurements skipped behind its normal measurement opt-in. See matrix.json for the exact observed list. Other package rows remain unknown, so this is not an unconditional package-wide uniqueness claim.

## Coverage and semantic difference

Per-test Go coverage used -coverpkg=./internal/native with random and family selectors. There are zero random-only Go blocks. Go coverage cannot instrument embedded runtime C. Supplemental LLVM instrumentation preserved production semantics, instrumented C with -fprofile-instr-generate -fcoverage-mapping, and preserved linked executables before temporary directory cleanup. coverage-compiler.py contains the wrapper. Both instrumented clean runs passed.

The random row covers 462/481 C lines; the sweep family covers 391/481. c-exclusive-lines.json records 72 random-only executed C lines. Random execution reaches line 225 1792 times, while the family never reaches it. The family exhaustively enumerates Unicode code points and short contexts. Random inputs include scrambled long combining-mark runs and repeated prefixes, reaching the long-run counting-sort branch. This is a production algorithm difference, not a fixture-specific condition.

## Reproduction

Apply D1.diff to the stated base, or recreate the overlay mapping internal/native/runtime/normalize.c to the mutated scratch file. Run the command in rows.json with its separate build cache. The overlay substitutes production C only. Runtime archive keys also hash embedded C/header contents, flags and compiler version, so the changed runtime cannot reuse the original archive.

D1.diff passed git apply --check. D1-release-compile.log records successful compilation with runtime release clang flags: -std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2. Matrix native sanitizer products compiled and executed; the kill was an output disagreement, not a build failure or panic. Production and test source in the checkout remain unchanged.

## Baseline, timings and limits

Warm env.sh worked; setup skipped. npm ci in stage3/api completed in 0.926 seconds. Current scope contains 282 top-level tests. Whole-package clean baseline cooked at 90.143 seconds, without an earlier per-test failure. It was narrowed to current normalization callers identified in normalization-callers.txt. Clean Go coverage runs passed: random 23.998 seconds, family 45.588 seconds. C-instrumented clean runs passed: random 59.502 seconds, family 48.436 seconds. D1 matrix took 47.577 test-binary seconds. These include native builds performed inside tests; separate native build times were not extracted. Logged test runs total about 315 seconds; elapsed work about 14 minutes. No three-run medians were requested for this defense.

## Brief feedback

The requested 15 GB free threshold is impossible on this workspace's 8.8 GB /tmp filesystem. Initial free space was 6.3 GB there and 6.5 GB on /workspace. Earlier identified /tmp/defend-enums scratch/cache was removed; protected scratch inputs required chmod before removal. /workspace is a separate filesystem, so removing /tmp caches does not free it. No full-disk failure occurred.

Go per-test coverage alone misses the code under test here. Actual C coverage required a compiler wrapper and LLVM tools found outside PATH under /workspace/adamic-tools/llvm/bin. Profiling changes cost but not answers, and both profiling baselines passed. A future brief should explicitly allow language-appropriate runtime coverage.

The whole package's 282 tests exceeded the binary budget during heavy regexp/native compilation. The allowed bounded matrix was necessary. The brief permits bounded matrices but defines defended in package-wide terms; this report makes the narrower observed uniqueness explicit. The opt-in long measurement row is listed as skipped, not passed. No unknown row is counted as passing.

The row's name matches its assertions: deterministic random agreement against independently generated Node inputs, exact output comparison and line count. No test was deleted, rewritten or weakened. One successful targeted attempt sufficed; no further mutants were needed. Repo-wide uniqueness and indirect callers outside the bounded selector were not tested.
