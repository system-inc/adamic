Built: original toPath, performance.measure and createBinaryExpression member contracts certified; Set producer shape probes retain the existing JavaScript refusal.
Commits: follows pushed 0515a19d; this report belongs to the next commit on codex/views-callables.
Checks: 17 fixtures pass Node, release/sanitized native, JavaScript and leaks; original spans, restored tests, lane counts and oracle vet pass; whole-table counts updater fails outside this group.
Mutants: seven applicable native/JavaScript callable and deferred payload mutations caught, 17 pair-level checks, all restored.
Uncovered: 45/2818 pairs and 1593/11063 candidate reads certified; 2773 pairs and 9470 reads remain; intrinsic Set producers require further representation and signature support.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 85 and 86 contribute 20 candidate reads each; rank 88 contributes 19. Complete original signatures and member reads are retained, with reduced adjacent carriers. Certification does not establish exact production reachability.

toPath checks its string parameter and Path result. performance.measure preserves all three original parameters, including both optional marks; omitted and supplied marks agree with Node. createBinaryExpression preserves its Expression parameters and BinaryOperator | BinaryOperatorToken operator. Scalar and aggregate operators agree with Node; a narrower producer is refused at the member read. A reached left.value read checks its string payload lazily.

The original-declaration verifier now records exported performance function headers separately from fixture method signatures. It follows ResolutionCacheHost to resolutionCache.ts and performance.measure to performance.ts. Twenty retained original members cover 426 conservative candidate reads; the new group verifier checks all 17 fixtures.

Mutants: native-arity/javascript-arity and native-parameters/javascript-parameters are caught for all three pairs. Native-result/javascript-result are caught for toPath and createBinaryExpression. Payload-reads is caught by the binary left.value pin. Seven runs and 17 pair-level checks; all restored, no compilation failures counted as mutant evidence.

The next Set family is blocked. set-intrinsic/add.a and has.a are reduced shape probes, not certified original rank 89/90 AST witnesses. Source Node prints 1 and true. JavaScript stops at the member read with function with unknown signature. Native compilation rejects passing adamic_map * into an adamic_object * parameter. The failed native probe is preserved in logs/path-measure-binary/set-native-boundary.log. TestCheckedViewCallableLaterRankedIntrinsicSetRefusal pins Node and the JavaScript refusal only. These probes are excluded from certification and native counts. No intrinsic metadata, callback convention or representation guard is relaxed.

Commands used source /workspace/adamic-tools/env.sh, with test output redirected directly to logs:

```sh
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-original 61,67,68,69,70,72,73,74,75,76,77,78,79,81,82,83,84,85,86,88 later-ranked-original-witnesses.json
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 85,86,88
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(resolution-path|performance-measure|binary)$' -count=1 -timeout 5m
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py resolution-path performance-measure binary
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts
go vet ./internal/oracle
```

Focused oracle passed 7.242s, restored tests 46.297s, lane counts 34.306s. Counts add exactly 17 rows and leave existing rows unchanged. Whole-table updater exited one in 35.816s, including outside-group missing @types/node 25.3.3 requirements. Vet passed with an empty log. Durable logs are under logs/path-measure-binary.

This continuation pushed four groups: exit/null/parenthesized (375160dc), type-reference/options/write (ccdf894e), declaration/access/source-files (0515a19d), and this path/measure/binary group. Together they add 13 certified pairs, 267 candidate reads and 72 fixtures; 28 mutant runs cover 72 pair-level checks. Original createIfStatement destructuring and liftToBlock method-value witnesses remain refusals. Static totals remain 4/308 pairs and 34/1503 reads; 304 pairs and 1469 reads remain.

Setup reused: done 417.425s, nproc 5, quota 4 CPUs; Go 1.27.1, clang 20.1.8, Node 24.19.0. No production compiler changes, whole-package tests, full gate or PR.
