Built collapse.CompareBreakpoints in one .a helper, reusing exact Go-width Integer arithmetic.
Claim 9d12e729 was pushed before code; implementation commit contains this report.
Go consumer tests pass; 22,342 pairs yield 63,689 identical bytes on source Node, emitted JavaScript and sanitized native.
Mutant uses the wrong first operand for ascending subtraction; successful, empty-stderr executions differ from Go on all three runtimes.
Uncovered: Windows/32-bit Go hosts, invalid UTF-8 strings, standalone bucket implementation and full consumer findings parity.

Six dependency entries removed from the same better-tailwindcss consumers as the other two helpers; zero final blockers alone. All 145 captured fixtures from those consumers actually run through Go; two comparator calls are observed. Controls pair the first helper's 2,589 inputs against numeric, signed, unit, equal-value and function inputs, supplementing with arbitrary deterministic pairs and Unicode scalar order controls. compare_consumers.json gives exact counts and corpus hash.

Behavior follows the real Go function: equal strings return zero before consulting buckets; different buckets compare alphabetically independent of direction; missing numeric prefixes fall back to raw strings, also independent of direction; numeric prefixes subtract in ascending/descending order with Go signed 64-bit wrap. Unicode scalar comparison matches UTF-8 byte lexicographic comparison for valid Unicode strings, avoiding JavaScript UTF-16 order for astral scalars. The result retains Integer's exact high/low words and decimal serialization.

Already-owned breakpointBucket is an explicit callback dependency. Go supplies left/right bucket facts through the original private helper for the isolated comparison; no guessed replacement or duplicate bucket implementation is delivered. This establishes the comparator conditional on that dependency, not whole-rule readiness. Capture overlays only wrap/rename the unchanged comparator body and redirect test fixture package paths to the installed pinned Tailwind 4.3.3.

Reproduce with setup environment and npm installation from LEADING_REPORT.md: go test ./stage1/cohere/lint/helpers -run '^TestWave05' -count=1 -v > /tmp/w05-helper-tests.log 2>&1, or python3 stage1/cohere/lint/helpers/wave05/compare_validate.py > /tmp/w05-helper-compare.log 2>&1. Logs under evidence/compare-* and compare_evidence.json record byte sizes, hashes and empty stderr. compare_cases.json.gz retains exact inputs with Go-supplied bucket facts.
