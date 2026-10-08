Built: original System.exit, NodeFactory.createNull and NodeFactory.createParenthesizedExpression member contracts certified; original createIfStatement binding retained as a refusal witness.
Commits: continued from a9fea5756320766a76083248e2c16ea0bca5e7ac; this report belongs to the next commit on codex/views-callables.
Checks: 15 executable fixtures, original declarations/read spans, Node, release/sanitized native, JavaScript, leaks and lane counts pass; whole-table counts updater fails outside this group.
Mutants: native/JavaScript arity, result and parameter checks plus deferred payload registration caught across applicable pairs; seven runs restored.
Uncovered: 35/2818 pairs and 1390/11063 candidate reads certified; 2783 pairs and 9673 reads remain; original destructured read and pending Union convention remain uncertified.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 72, 74 and 75 add 22, 21 and 21 candidate reads. Fixtures retain complete original declarations and member reads. Adjacent carriers are reduced. System.exit is a callable member contract with observable producer arguments; the original process implementation is not executed or claimed certified. Omitted, explicit undefined and supplied numeric arguments match Node. The null result is an aggregate. The parenthesized-expression callback reads its argument lazily, with the narrower string field stopped at expression.value.

Rank 73 is not certified. Its original createIfStatement: factoryCreateIfStatement binding and full declaration are retained in if-statement/good.a. Node prints 7. Lowering refuses unsupported destructuring representation conversion contract. TestCheckedViewCallableLaterRankedBindingRefusal pins that refusal. The initial wider test attempt recorded that refusal for all generated binding variants; the final artifact keeps only the original valid witness. No direct property-read replacement is used as certification evidence, and no calling-convention guard changed. Rank 71's library Map callback signature includes any and remains outside these supported member fixtures.

All seven mutants were restored. Native-arity and javascript-arity were caught for all three pairs by the pinned early member-read refusal versus execution. Native-result and javascript-result were caught for null and parenthesized by the member-read result pin versus later result handling. Native-parameters and javascript-parameters were caught for system-exit and parenthesized by the parameter refusal versus execution. Payload-reads was caught for parenthesized by the expression.value string-read pin versus sanitizer output. All are executable failures, with no build failures counted. This is 15 pair-level checks over seven mutation runs.

Commands use source /workspace/adamic-tools/env.sh and redirect test output directly to the listed logs:

```sh
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-original 61,67,68,69,70,72,73,74,75,76 later-ranked-original-witnesses.json > /tmp/lane5-if-original.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 72,73,74,75 > /tmp/lane5-if-original-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-if-restored.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py system-exit null parenthesized > /tmp/lane5-if-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-if-final.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-if-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-if-counts.log 2>&1
```

Restored oracle passed 17.034s before mutants and 16.961s after them, including prior later-ranked members and both refusal witnesses. Original verifier confirms 16 retained fixture declarations and original reads: 15 certified fixtures plus the binding refusal witness. Lane counts updater passed 28.512s, adding 15 measured rows with no existing row changes. Whole-table updater exited one in 35.798s on the previously reported outside-group lowering/count execution cases. Durable raw logs are under logs/exit-null-parenthesized.

Toolchain setup is reused from the prior report: done 417.425s, nproc 5, cgroup quota 4 CPUs; Go 1.27.1, clang 20.1.8 and Node 24.19.0. Static inventory remains 4/308 pairs and 34/1503 reads. These are candidate inventory counts, not measured production reachability. No compiler/runtime changes, full-package tests, full gate or PR.
