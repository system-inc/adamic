Built: 20 further pairs / 43 ranked candidate reads; cumulative share c 287 pairs / 773 reads. Fixtures, tests, own counts and evidence only.
Commits: continues edee6e6cd on codex/views-callables-c; delivery SHA is reported after push.
Commands and outputs: array oracle PASS 19.729s; Map oracle PASS 9.317s; original verifier PASS 287 pairs / 574 fixtures; scoped counts update PASS 4.104s and verification PASS 4.321s.
Mutants: ten Array.push write-certificate omissions and ten Map receiver-certificate omissions caught in native and JavaScript, forty executed mutant comparisons.
Uncovered: no complete original tsc execution or arbitrary generic-instantiation claim; Set, optional-host and bind/condition families remain delegated or excluded; no compiler/runtime changes.

The corrected skip rule revisits recorded uncertified ranks. Certified ranks and observed code-needing ranks are skipped. The original ledger remains unchanged. Pair/read totals are static ranked weights, not dynamic execution or allocation reachability.

Array certificates: 2501, 2504, 2507, 2510, 2513, 2567, 2570, 2573, 2576 and 2579, one read each. Nine preserve the original T[], U[] or V[] receiver and original push declaration in an explicit number specialization. The tenth preserves every TypeSystemPropertyName enum member and value. Positives print 2. A boolean-array producer behind the unknown carrier stops before its push with `element read failed: <array write> expected number, found boolean`. The mutant changes exactly one ArrayPush.DictionaryProduction flag; both backends print 2 and exit 0, proving the pinned guard catches its omission. Generic certificates cover these concrete instantiations, not every possible generic payload.

Map certificates: 119 (14 reads), 230 (7), 866 (2), 872 (2), 875 (2), 878 (2), 1712 (1), 1715 (1), 1730 (1), 1736 (1). Original Map receiver types, get/has/delete declarations and member-read paths remain intact. Adjacent payload interfaces are reduced explicitly, following the lane harness. An empty Map with the correct immutable source schema runs the original missing-key call and prints completed. An empty Map<string, boolean> with the wrong value schema stops with exit 70, naming `(value as Target).items`, the expected Map type and the found Map<string, boolean>. This is a receiver-schema certificate for a known intrinsic, not an arbitrary closure proof. Removing exactly one Property.ViewContract permits the missing-key call and completed output in both backends, so the diagnostic assertion catches the mutant. Empty maps intentionally isolate schema checks from payload dereferences and unrelated memory failures.

All twenty positives match source Node in native release, ASan/UBSan, leak/count checks and JavaScript. All negatives pin the complete diagnostic and exit 70 in release, sanitized native and JavaScript; source Node completes. Every mutant executes successfully rather than failing clang or a sanitizer. Exact per-rank messages are in continuation-candidates.json and map-certified.json. Original declaration/read spans, file hashes, original aliases and enum values are verified from independent TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. No cohere implementation was copied.

Two new code-needing pairs:

- Rank 11, NodeFactory.createNodeArray, factory.createNodeArray, 92 reads: both lowerings refuse its generic callable contract at the member read. Needs checked generic function producer/signature and invocation support. Original declaration retained; surrounding Node/NodeArray carriers are reduced.
- Rank 47, Math.min, Math.min, 36 reads: both lowerings refuse `inherited library member min read as an own field`. Needs an intrinsic Math receiver path/certificate preserving the actual library object, not a replacement record closure.

Both source Node controls print completed and their lowering refusal tests pass. SymbolTable get/set ranks 20/29 compile in both backends but have no runtime negative/mutant certificate yet; they remain pending, not code blockers. Existing code blockers, all 31 Set pairs / 107 reads, optional-host pairs and bind/condition contexts remain listed; their owner branches were not merged.

Counts: exactly forty own fixture rows appended, no old row changed. Each row is the measured allocation/use count for one new positive or negative, not a read-coverage count. The required global TestCountsAreRecorded updater fails on sixty inherited fixture subtests, including missing pinned @types/node and existing lowering/graph-region cases. Its complete log is retained; no shared green gate is claimed. The scoped updater and subsequent verification pass. No whole package test or full gate was run.

Commands (all test output redirected):

```
go test ./internal/oracle -run '^TestCheckedViewCallableShareCArray(Intrinsics|Mutants)/rank-(2501|2504|2507|2510|2513|2567|2570|2573|2576|2579)$' -count=1 -timeout=10m -v
go test ./internal/oracle -run '^TestCheckedViewCallableShareCMap(Mutants)?$' -count=1 -timeout=10m -v
go test ./internal/oracle -run '^TestCheckedViewCallableShareCCodeBoundaries/rank-(11|47)$' -count=1 -v
node stage3/interface-downcasts/lane5/share-c/verify.cjs /tmp/lane5-c-original
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableShareCContinuationCounts$' -count=1 -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableShareCContinuationCounts$' -count=1
```

Setup used GOPROXY=https://proxy.golang.org|direct and succeeded: Node 0.032s, Go 0.036s, submodules 0.069s, markdown dependencies 0.082s, clang 0.163s, build 43.752s, deferred test binaries 43.959s, cache 43.960s, done 43.985s. nproc 5, quota four CPUs. Every shell used /workspace/adamic-tools/env.sh. Gofmt and git diff --check pass. Pushes use only codex/views-callables-c.
