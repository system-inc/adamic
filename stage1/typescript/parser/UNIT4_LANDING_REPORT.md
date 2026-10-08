The incomplete compiler gate compares every selected input on typescript-go,
Node and sanitized native. `incomplete_test.go` prepares the healthy builds,
then starts its workers before the parent joins the parallel queue. It schedules
22,497 cases from the pinned 77-file compiler corpus on
`min(runtime.NumCPU(), 8)` workers. This box starts five workers. Each manifest
contains at most 64 cases and 8 MiB of input; one unusually large input is kept
whole. Every manifest child has a two-minute hang guard.

`incomplete_shards_test.go` requires ordered case headers, compares every case's
body without stopping at the first differing case, and names missing, extra, duplicate and changed cases. The full gate also
checks the planned count, the generated count, a completion ledger and the final
compared count. It rejects `ADAMIC_RECOVERY_START`, so the full gate cannot
silently become a suffix of the corpus. A failure retains inputs and all three
answers under `ADAMIC_RECOVERY_ARTIFACTS`, along with a runnable manifest.

`shared_test.go` shares the healthy oracle and sanitized/release port builds only
within one test process. Every scratch mutant directory builds independently.
Clang keeps the standard flags: sanitized builds use `-O1` with address and
undefined-behavior sanitizers; release builds use `-O2`. Main clang children have
four-minute hang guards. The merged lint entry requires its checker archive;
`checker_build_shared_test.go` links the real sanitized archive and uses the
landed split builder with one clang worker. A private compiler proxy guards
each clang call without changing the parent environment. Successful shard answers use the test's temporary
directory and are removed after comparison.

`testdata/batch_node.mjs` formats each console call with Node's formatter and
writes bounded chunks, including a final flush on normal completion or exit.
It is used by the incomplete gate only. `nodes.ts` reuses fixed field text and
common depth text; greater depths retain ordinary number formatting. Paired
77-file checks require identical output. The measurements are in
`validation/unit4/node-buffer-timing.json`, `writer-timing.json` and
`depth-timing.json`. The slower Map-based offset-cache experiment was discarded.

The one-byte and dropped-case experiments changed `main.ts` in scratch copies.
Both real test commands exited 1 and named
`testdata/recovery/corePublic.ts-cut-37.ts.txt`. Their source changes and exact
changed/dropped counts are recorded in `validation/unit4/mutants.json`; the
logs are `mutant-byte.log` and `mutant-drop.log`. The temporary probe tests and
mutant sources were never committed.

`validation/unit4/before-times.csv` records all Unit 3 top-level durations.
`after-times.csv` sorts all final top-level durations by active time, calculated
as run-to-pause plus continue-to-pass. `summarize.py` regenerates that table and
pass/fail/skip counts from `parser-final.jsonl`. These times overlap and do not
sum to the package wall time; parent tests include their parallel children's
completion. Corpus workers can run while their parent is paused; the requested
active-time formula excludes that queue interval, while package wall time and
child timers continue to include it. Unit 3 did not record package-boundary load readings; `before-box.json`
marks them unavailable rather than substituting late benchmark readings.

`testdata/lint_cases/wave13_top_level_await_new.ts` and its expected-difference
sidecar remain unchanged. The collection contains 58 cases: 57 agree and this
one retains its documented difference.

Run the full parser gate with the pinned, clean TypeScript checkout in
`ADAMIC_TYPESCRIPT_SOURCE`, `ADAMIC_PARSER_BENCH=1`, and
`go test ./stage1/typescript/parser -count=1 -json -v -failfast=false -timeout=30m`.
The opt-in benchmark flag avoids the two performance-test skips. Also run
`go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=10m`
and `go test ./stage1/typescript/parser -run '^TestLintCases$' -count=1 -v -timeout=10m`
with the same TypeScript source. Final measurements and counts are recorded in
`validation/unit4/report.json` and `after-box.json`.
