# TestCompileProfiles defense

Starting origin/main: 157a43552015f41a79331949c2e82b6f8c7caaab. The audit report, rows and code/oracle inventory were read first and are saved as prior-* evidence. One current top-level Test exists, unchanged since the audit. Full clean baseline passed, 39.290 test seconds / 41.341 wall seconds. No skip, red baseline, panic, timeout or narrowed matrix occurred.

## Outcome

Not defended after three production-only source mutations. Every standalone diff applies to origin/main, and every mutated port compiled using the unchanged test's own sanitized and release native builders. Only evidence files remain changed. There are no other package rows, no executor twins, and no subsumer.

D1 profile.a:25 swaps Parser(source,path) to Parser(path,source): the compile gate passes at 39.16 seconds, while the identical source witness changes five findings to zero. D2 profile.a:42 swaps comparator callback parameters left/right: the compile gate passes at 38.04 seconds, while finding offsets reverse from 6,23,34,43,56 to 56,43,34,23,6. D3 profile.a:10 swaps recursive ancestry child/index arguments: the compile gate passes at 37.77 seconds, while the source witness changes success to exit 70, adamic: panic: RangeError: Maximum call stack size exceeded. These are three argument-swap mutations from the permitted menu; none is an insertion, test edit, harness edit or oracle edit. Each full package run uses its own ADAMIC_BUILD_CACHE_DIR=/tmp/underscore-defend/cache/DN. The unchanged source runner is invoked separately only to demonstrate actual changed behavior. No source-witness failure is counted as a Go test kill.

## Coverage

Separate target and rest Go coverpkg profiles cover internal/load, internal/lower, internal/native and internal/javascript. The rest profile uses -run '^$' because the fresh package inventory contains no other Test. coverage-difference.json lists 4,814 compiler blocks reached by this row but not by the empty rest. This proves compiler execution, not execution of any .a rule method. code-and-oracle.md identifies the owned lowered methods and self compilation oracle. No external expected runtime result is asserted in the row.

## Scope and owner findings

The name TestCompileProfiles accurately describes its native compile assertions. It promises no speed threshold or runtime lint agreement. Profiles here means sanitizer/release/backend build variants, not cost measurement; the cost-mutant requirement therefore does not apply. The test log says emitted JavaScript compiled, but the actual assertion only checks os.WriteFile, not JavaScript parsing or execution. That wording overstates the JavaScript check.

These validly compiling runtime mutations do not establish that the compilation gate is redundant or should be deleted. Its protection against compilation regressions remains unmeasured. The brief's requirement for successfully compiling port diffs excludes an intentionally uncompilable port as a meaningful defense. Invalid syntax/type errors or injected clang warning failures were not used. No other packages were run, so no repository-wide uniqueness or runtime guard coverage is claimed. The prior untrue verdict is not a proof that this row cannot fail on a real compilation regression.

## Costs and ambiguities

Warm tools worked; setup skipped; nproc=5. npm ci in stage3/api took 0.492 wall seconds; inventory 6.118. Target compiler coverage took 52.951 wall seconds (37.050 test seconds), empty rest 2.096. Mutant full-package process walls: D1 41.702, D2 40.400, D3 40.123. Combined build/check walls sum to 122.224 seconds; individual native build phase times are not exposed by the unchanged test and are not invented. Separate source witnesses took about 0.15 to 0.17 seconds each.

Two brief limitations cost analysis time: Go coverage cannot instrument the .a code, and semantic mutations cannot assess the stated compilation-only contract. These are disclosed rather than substituting source witness coverage for test coverage. No missing audit files, changed row names, opt-ins, twin ambiguity or tool installation issue was observed. The evidence field's D{1,2,3} is shorthand for three separate runs, not literal shell brace expansion; commands.jsonl records each exact argument vector and results.json records each cache run. All test output goes to saved logs, not pipelines. No tests were deleted, rewritten or weakened; no main push or pull request was made.
