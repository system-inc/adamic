Fixture paths now use .ts.txt/.tsx.txt; see `FIXTURE_EXTENSIONS_LANDING_REPORT.md`.

The top-level await/new difference was a parser lookahead bug, not a language
gap. `parser.ts:456` now shares the same-line identifier, keyword and literal
lookahead used for yield. `parser.ts:1792` uses it for await, matching
`cohere/TypeScript/tsc/internal/parser/parser.go:5195` and its helper at line 4085.
The old tree split await and the constructor into separate statements, making
lint report no-new on the constructor. The new tree nests the constructor under
AwaitExpression and matches typescript-go's tree and diagnostics byte for byte.

The original input is `testdata/lint_cases/wave13_top_level_await_new.ts.txt`. Its
expected-difference sidecar is removed. The two new inputs are
`testdata/lint_cases/top_level_await_new_script.ts.txt` and
`testdata/lint_cases/top_level_await_new_module.ts.txt`. They cover constructor calls,
constructors without parentheses, argument and initializer expressions,
parentheses, binary and conditional expressions, ordinary and async function
bodies, and script/module files. Numeric, bigint, string and keyword operands
exercise the shared lookahead. A plain new expression remains a positive no-new
control. All three paths are in `stage1/cohere/lint/lint_test.go:266` for both
no-new and all-rule comparisons in the ordinary lint gate.

`validation/await-new/run_gates.py` records the four requested gates, with
ADAMIC_TYPESCRIPT_SOURCE at clean pinned checkout 050880ce, parser benchmarks
enabled, and fresh profiling artifacts in /tmp/adamic-await-new-profile.
TestProfileArtifacts builds the release and profiled snapshots that
TestProfileSnapshotsAgree compares; neither profile test skips.

All gates passed on this exact change, with zero failures and zero skips:
parser package 942.807 seconds (178 passes, all 22,497 incomplete comparisons);
TestRulesAgree 138.083 seconds (one pass); standalone TestLintCases 22.174
seconds (61 passes including its 60 files); TestProfileSnapshotsAgree 178.200
seconds (one pass). Fresh TestProfileArtifacts also passed in 70.910 seconds;
the combined profile package took 249.214 seconds. All 60 lint parser files
now match; no expected-difference sidecars remain. Counts and full command
wall times are in `validation/await-new/report.json`.

A temporary `TestAwaitNewRegressionMutantProbe` restored the old await
condition in a scratch parser copy and used the real lint-case comparison.
Both Node and sanitized native rejected its output on each of the three
inputs: six caught checks. The proof took 12.440 package seconds. Its source
and mutated parser were not committed; the mutation and input paths are in
`validation/await-new/mutant.json`, and its results are in `mutant.log`.
Vet passed for the parser and lint packages. This unit is pushed once after
all fixtures, gates and mutant checks passed locally.
