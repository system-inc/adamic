# no-cond-assign

The port calls the pinned Go cohere rule through its typed string-options adapter. The default and unrecognized modes report direct conditional-test assignments, with the Go asymmetry that a ternary test has its parentheses stripped. The `always` mode reports assignments anywhere in a test, stopping at arrow functions and blocks and excluding initializers, updaters and bodies by containment. Findings cover only the assignment operator, including operators preceded by comments. There are no fixes or suggestions.

Go's `TestNoCondAssignFiresUnderExceptParens/parenthesized_ternary_test` reports the parenthesized assignment in `(((3496.29)).bkufyydt = 2e308) ? foo : bar;`. This port retains that behavior. It also follows Go cohere's single finding per assignment under `always`, rather than the duplicate diagnostics described for the original upstream implementation.

The five witnesses are in `testdata/`: `default.ts.txt`, `always.ts.txt`, `operators.ts.txt`, `unicode.ts.txt` and `unknown-mode.ts.txt`. They include all sixteen assignment operators, all five conditional test positions, omitted for clauses, comment and Unicode spans, nested assignments, function boundaries, and assignments in non-test positions. Options sidecars exercise `always` and an unrecognized mode. Every witness produces a Go finding.

Run from the repository root with the setup environment sourced:

```sh
export GOPROXY='https://proxy.golang.org|direct'
go run ./cmd/lint-registry
python3 stage1/cohere/lint/rules/no-cond-assign/testdata/selected.py
ADAMIC_TYPESCRIPT_SOURCE=/path/to/clean/pinned/TypeScript python3 stage1/cohere/lint/rules/no-cond-assign/testdata/run_full.py
```

`selected.py` uses a temporary Go overlay, adding a parallel test that compares all 73 captured unique upstream cases, five witnesses in selected and all modes, and inherited generated cases. It also makes this rule the sanitized native mutant canary. The shared harness remains unchanged. `testdata/evidence/selected.log` records successful Go, source Node, emitted JavaScript and sanitized native parity, plus a mutant that compiles and runs successfully, misses a ternary finding on Node and emitted JavaScript, and produces identical mutated output natively.

`run_full.py` verifies the clean TypeScript pin and WASI sysroot, enables throughput, creates one fresh directory for both profile variables, runs the whole lint package once, and records test events, counts, wall time, nproc and load. `testdata/evidence/full-summary.json`, `full.log` and `full.jsonl.gz` hold the result. Counts include named subtests and exclude the package result. The runner does not disable the baseline pending bridge-refusal test's explicit skip.

Setup timings and the search of 2,096 remote refs are recorded in `testdata/evidence/setup-initial.log`, `setup-sdk.log` and `remote-check.log`. The rule directory is the complete change; generated registration and upstream cohere remain untouched.

The whole package exited 0 with 152 passing named tests/subtests, zero failures and one skip, `TestCheckerBridgeRefusalPending` (the baseline waits for `TSGoError`). Wall time was 1095.84 seconds, nproc was 5, and one-minute load was 1.19 before and 4.74 after. All requested external input sets were enabled. No rule helper or language gap blocked this port.
