Built the item 127 boxed-union every refusal, item 128 findLast guard, and item 131 visit comment correction.
Delivery: compiler/fx5-every-refusal. Implementation e49ae19ee17b2830d731030bfc345ebe4fd67f0d; merged main a7448d73cd17f16362b6cbc5c5c111080da64e43 in e02662335700f4b3e938aca364d1fa7037389b35.
Checks: full internal/lower 69.406s; TestCallTargetReaders 2.386s; counts refresh 78.306s, unchanged.
Mutants: every compiles with wrong native stdout; findLast compiles to C that clang rejects. Both refusal tests fail when reverted.
Limits: this refuses unsupported narrowing; it does not implement stored-layout reads or fix the separately observed all-number union-literal storage issue.

The temporary every restriction applies only when the stored element representation is ir.Union and the predicate target has a different known representation. The diagnostic carries the exact program path, line and column of the call, the requested wording, and an explicit fix. Named guards, inferred arrows, readonly receivers, early returns followed by reduce, property receivers, forwarding to a number-array function, and Cat | number receivers each have an independent refusal test.

The pinned checker library has a receiver predicate only on every among these visits. Negating some does not narrow its receiver; find and findLast narrow the returned element. Therefore some-negated stays accepted, find retains its existing result-layout guard, and findLast now uses that guard. filter's existing boxed-union conversion refusal remains unchanged. forEach, findIndex and findLastIndex have no narrowing result or receiver predicate. The findLast fixture is refused like find, rather than compiled. Its source on Node prints found 4 / double 8 (verified by the mutant's JavaScript agreement before clang).

The optional-number and optional-string controls call lowersAndAgreesWithNode and also verify native agreement. Additional native controls cover negated some, findLastIndex, a boolean callback without a receiver predicate, and a predicate that retains boxed-union storage.

The primary every witness and boolean control construct genuinely mixed union storage, then pop the string. This isolates narrowing from a separate initialization issue: an all-number literal annotated as a union array can crash natively even without a receiver-narrowing predicate. initial-literal-layout.a preserves that input; lower-final.log.gz records native empty stdout versus Node true. That separate issue is not fixed here.

Exact final commands, each with output redirected to its named log:
- source /workspace/adamic-tools/env.sh
- timeout 420 go test ./internal/lower -count=1 -v -timeout 6m > review/compiler/fx5-every-refusal/lower-green.log 2>&1
- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s > review/compiler/fx5-every-refusal/call-target-readers-final.log 2>&1
- timeout 780 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 12m -args -update-counts > review/compiler/fx5-every-refusal/counts-final.log 2>&1
- timeout 380 python3 review/compiler/fx5-every-refusal/run-mutants.py > review/compiler/fx5-every-refusal/mutants.log 2>&1

The every reverting diff removes only the new guard. TestArrayNarrowingEveryRefused detects acceptance. The independent native witness still agrees with source Node through the JavaScript backend, then native prints total 9.2730926099636e-310 with exit 0 versus Node total 3 with exit 0. This kill is a behavior comparison, not a build warning. The number reflects the observed run, not a deterministic expected value.

The findLast reverting diff removes findLast from the find name test. TestArrayNarrowingFindLastRefused detects acceptance. Its witness agrees with Node through JavaScript, then clang rejects assignment to adamic_maybe_number from incompatible adamic_heap *. That build error is the specifically requested item 128 reproduction. The runner restores source and removes its temporary Go witness in finally blocks.

Setup: GOPROXY was exported as https://proxy.golang.org|direct before bash cloud/setup.sh. The initial timeout 240 run stopped during cache warming: Go 0.067s, Node 0.081s, clang 0.465s, Markdown dependencies 4.236s, submodules 41.354s. The timeout 300 retry succeeded: Go 0.022s, Node 0.024s, submodules 0.102s, Markdown dependencies 0.126s, clang 0.195s, Go build 35.212s, cache warm 35.323s, done 35.456s. nproc is 5, with cgroup cpu.max 400000 100000 (four CPU quota).

The first full lower run exposed missing pinned Node types. timeout 120 npm ci --prefix stage3/api installed @types/node 25.3.3 and TypeScript 6.0.3 from the existing lockfile. The first counts run hit its 90s package timeout; the longer bounded rerun passed. Earlier failures are preserved in compressed logs. No other full package or full repository gate was run.

Counts regeneration rewrote the existing table with identical bytes. These are lower testdata fixtures, so they add no registered oracle counts rows.

New test leaf times from the final full package run:
- TestArrayNarrowingBooleanCallbackAgrees: 0.73s
- TestArrayNarrowingCatRefused: 0.08s
- TestArrayNarrowingEarlyReturnRefused: 0.17s
- TestArrayNarrowingEveryRefused: 0.15s
- TestArrayNarrowingFindLastIndexAgrees: 0.61s
- TestArrayNarrowingFindLastRefused: 0.09s
- TestArrayNarrowingFindRefused: 0.08s
- TestArrayNarrowingForwardRefused: 0.07s
- TestArrayNarrowingInferredRefused: 0.14s
- TestArrayNarrowingOptionalNumberAgrees: 0.74s
- TestArrayNarrowingOptionalStringAgrees: 0.75s
- TestArrayNarrowingPreservedUnionAgrees: 0.71s
- TestArrayNarrowingPropertyRefused: 0.18s
- TestArrayNarrowingReadonlyRefused: 0.14s
- TestArrayNarrowingSomeNegatedAgrees: 0.72s

After merging current main, the full lower package, call-target reader guard, counts regeneration, and both mutants were rerun and passed their intended checks. No production source changed in the main merge. Counts remain unchanged relative to the merged main table.

Committed-tree lane output: lane checks 2.0 s: gofmt and tools on 2 Go files, t.Parallel on 1 test packages; vet 1 packages

Lane command: timeout 180 bash -c 'git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'. The initial multi-ref fetch did not create origin/cloud/merge-tree in this single-branch checkout; explicit destination refspecs for both lane refs made the prescribed command work.
