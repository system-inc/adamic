Built: original Scanner.getTokenEnd and getTokenFullStart member contracts checked at viewed reads.
Commits: follows pushed 685888b2 on codex/views-callables.
Checks: eight fixtures preserve full original declarations/reads and pass Node, release/sanitized native, JavaScript and leaks; eight counts rows added without changing existing rows.
Mutants: native-arity, javascript-arity, native-result and javascript-result caught for both pairs, four runs and eight pair-level checks, all restored.
Uncovered: 50/2818 pairs and 1685/11063 candidate reads certified; 2768 pairs and 9378 reads remain.

Reporting date October 12. Ranks 95 and 96 contribute 18 reads each. Original source pin 050880ce59e30b356b686bd3144efe24f875ebc8. Fixtures retain scanner.getTokenEnd and scanner.getTokenFullStart with their original zero-argument numeric signatures. Arity variants execute on Node and print 9 but stop at the viewed member read in both backends. String-result producers likewise stop at that member read. No parameter or aggregate-payload mutant applies to these zero-argument scalar members. Exact production reachability remains unmeasured.

Commands used source /workspace/adamic-tools/env.sh, with direct output redirection into /tmp/lane5-scanner-*.log, preserved under logs/scanner-end-full-start:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 95,96 > /tmp/lane5-scanner-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(token-end|token-full-start)$' -count=1 -timeout 5m > /tmp/lane5-scanner-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py token-end token-full-start > /tmp/lane5-scanner-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(token-end|token-full-start)$' -count=1 -timeout 5m > /tmp/lane5-scanner-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-scanner-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-scanner-counts.log 2>&1
```

Oracle passed 3.496s, restored oracle 3.444s, lane counts 42.116s. Whole-table counts failed outside this lane in 36.381s. Source evidence additionally prepared ranks 102, 103, 106 and 107 in next-original.log. Set and Map intrinsic members remain deferred with requirements in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals remain 4/308 pairs and 34/1503 reads. No production changes or full gate; setup reused, done 417.425s, nproc 5, quota 4 CPUs.
