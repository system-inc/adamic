# u051: internal/native number audit

Starting commit: 1f34d0d300301faebc94d397adee1a090923d2c7.
All four names exist in the test list and remain in the historical files. They are four independent rows, not shards or witnesses. No test, harness or oracle was mutated.

Code under test: Adamic's C formatter, shortest decimal digit implementation and parsers, plus production internal/buildcache. The artifact row calls that production code through a test-only wrapper; the wrapper was left unchanged. Node actually runs for all three numerical rows. The artifact oracle is self-authored directory equality and build count.

## Results

results.json is the deliverable JSON. scope.json lists all six rows in the bounded matrix. matrix.json normalizes kills. Sacred and unique_kills mean unique within that bounded set, not the entire package. Every omitted package or repo kill is unknown. The general formatting row is subsumed by the math row on exactly three production mutants, not a deletion recommendation. Its median is 1.764 s; math is 0.798 s. The power row has a single shared catcher, math, but math is 86.9% slower than its 0.427 s median. The verdict menu omits this case. I asked for a label preference asynchronously; absent a reply, overlapping is the provisional retain label, explicitly not a claim that no common catcher exists. No bounded-scope limitation was used to justify cannot-judge.

All numerical comparisons are actual native runtime answers against actual Node. The artifact row's expected count is hand-written. Every row failed its own entry's empty answer: E01 formatter; E02/E03 parseInt/parseFloat; E04 production Get. No vacuous row was observed. No panic aborted the bounded matrix.

## Plan and replay

plan.json was frozen before any mutation outcome, SHA256 49cdd0ef214fbc2527f06c2a077ab779f6bc433a7165b6a6593f45324c561f01. Twelve production mutants use only the fixed menu; four extra columns are entry-specific empty-answer probes. M05 changes the first of two identically written dtoa conditions, in initial_scaled_start_values_positive_exponent. Its occurrence was disambiguated before freezing the plan.

Every standalone unified diff under diffs/ applies to the starting commit and passed go vet for the mutated package. C diffs also passed clang syntax/compile checks with the runtime's warning policy. validations.json gives commands, times and exits. E04 initially left unreachable statements that go vet rejected. Its replay diff now replaces the Get body with the empty return; its switched implementation already had exactly that runtime effect. The immutable original plan and diff-repairs.json retain this repair. No new mutant was designed from kill results.

selector.diff contains the switched scratch source and helper files. It was reverted before commit. C selectors read the process environment dynamically, so one cached runtime archive exercises every selector. Runtime cache keys hash source/header bytes; Build links those archives anew. Each matrix process also sets ADAMIC_NATIVE_SPLIT=u051-<id>. This value is not the literal split-enable value 1 at this commit; cache freshness rests on source hashing and dynamic selection, not on that label. No compiler lowering was mutated and no port product cache was involved.

Run source /workspace/adamic-tools/env.sh, then python3 review/test-audit/internal-native-number/matrix.py after applying selector.diff. run.py records actual timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ argv and all stdout/stderr in log files, never pipelines. For central replay, apply a single diffs/MNN.diff on the starting commit without selector.diff and run every reached package centrally.

## Reachability

functions-covered.txt and coverage.out show 22 Go production functions reached by the clean four-row baseline. code-under-test.txt lists them and the numerical roots. c-reachable-functions.txt and callgraph-number/dtoa/parse.log provide conservative static C reachability with 73 names; conditional dtoa modes are over-approximated. They are not runtime function coverage. External trim/release transitive bodies were not fully expanded. runtime-callers.txt and cache-callers.txt record other test callers. Many cache callers belong to large fixture families and were omitted to keep the matrix bounded. No tests in other packages were run.

## Survivors

M04 lowers the direct integer fast-path bound from 2^53 to 2^52. Both formatting sweeps and math passed it. The four direct witness values now take different branches but have byte-identical answers before/after: 4503599627370496, 4503599627370497, 9007199254740991, 9007199254740992. No changed-output witness found; equivalent candidate, not unguarded. See M04-before.log and M04-after.log, probes/format/main.go.

M11 replaces product names by a constant in Key. The matrix passes, but Key(alpha)!=Key(beta) before and Key(alpha)==Key(beta) after. Exact hashes are in M11-before.log and M11-after.log; probe is probes/key/main.go. Product-name separation is changed behavior unguarded by this bounded matrix. The artifact row uses only the label fixture, so it cannot observe that distinction.

## Baselines, time and restoration

Warm env.sh worked; setup skipped. npm ci in stage3/api succeeded in 0.669 s. nproc=5. Whole-package clean baseline timed out after 90.125 binary seconds, 91.992 process wall seconds, with no individual failed-test events. The explicit big-package timeout exception was used. Clean four-row baseline, six-row switched control, all twelve target timing runs and the restored original four-row control passed.

Each row's reported seconds is the median of three separate count=1 package ok durations. Comparator rows also received three timing runs. timing.json records command sums. Native runtime archive building occurs inside test-binary time, so the cold switched control took 21.676 process wall seconds and 11.118 binary seconds; the 10.558 s difference is startup/Go-build overhead, not an isolated compilation measurement. Probe builds were not separately timed. C source checks and Go vet total approximately 10.08 s, including the rejected E04 attempt and repair.

## Brief friction and limits

1. The verdict menu has no case for a single shared catcher that exceeds the 20% slowdown limit. This affects the power-of-two row. Provisional overlapping means keep it; its common catcher is disclosed, rather than falsely claiming multiple incomparable catchers.
2. A timed-out baseline is a red go command, while the brief explicitly authorizes narrowing after 90 s. I used that exception and required green slice/control baselines before mutants. The full package baseline never completed.
3. Numerical code is C embedded by Go. Go coverage does not instrument the numeric functions, and go vet alone cannot validate C. I added static C callgraphs and standalone clang checks; exact runtime C function coverage remains uncovered.
4. Distinct numerical entries share a Go Build wrapper. Empty-answer probes target the numerical functions that decide the answers, not Build's error return. The parse row has two applicable probes, and it failed both.
5. The artifact wrapper lives in _test.go, but the actual cache implementation is production internal/buildcache. Mutating the wrapper would test the harness. All cache mutations are in production.
6. The suggested ADAMIC_NATIVE_SPLIT labels are not a general product cache key in this native Build implementation; only value 1 enables splitting. Dynamic C selection and source/header hashing prevent stale mutated behavior here. The actual compiler/port case described by the brief does not arise in this unit.
7. Return-at-entry left unreachable Go code in the initial E04 standalone diff. Removing the old function body preserved the intended empty answer and passed go vet. The failed check remains logged.
8. A reduced optimization bound can preserve every observed answer. M04 is an equivalent candidate; green tests are not enough to call it unguarded behavior.
9. Warm tools still required npm ci. The numerical Node scripts load only built-in node:fs; no additional node_modules directory was found for these tests.
10. Cold archive creation is inside the test binary, whereas Go compilation is outside. I report binary medians and process wall totals separately; an exact isolated original compilation time was not measured.
11. Not every cache caller was run. The chosen matrix includes the artifact row and a Key caller, while larger record, normalization, decoder, WASI and TSGo families remain outside it. Uniqueness is expressly bounded.

Production and harness files are unchanged from the starting commit. Only this evidence directory is committed. No main push and no pull request.
