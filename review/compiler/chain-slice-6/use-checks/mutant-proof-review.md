# Independent mutant verification

All original campaign records, including `mutants/results.json`'s `native-scalar-properties` caught=false entry, are retained. That entry means the runner's hard-coded diagnostic marker `stale narrowing use failed` was absent, not that its test survived: the unmodified test failed at `narrowed_union_test.go:49`, with exit 70 and `adamic: panic: dynamic property read on an unsupported runtime value`. The expected per-use stderr differs. `scalar-mutant-verification.log` independently repeats that same behavioral assertion failure, exit 1, without a compiler/build failure. No test expectation, original result, or diagnostic was rewritten.

`numeric-enum-mutant/test.log` independently proves that checking a whole numeric enum as a finite list rejects the valid flags agreement: the test fails at the stored-flags exit assertion. Its source is the non-compilable overlay `narrowing_uses.go.txt`.

Automatic approval review rejected a proposed result-filter/marker edit because it could conceal failed evidence. The proposed command did not execute. These separate logs preserve both the raw classification problem and the actual assertion that failed.
