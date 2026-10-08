Built: original ranks 76 through 79 member contracts certified with 21 fixtures, adding 84 candidate reads.
Commits: follows pushed 375160dc; this report belongs to the next commit on codex/views-callables.
Checks: original declarations/read spans and inheritance paths, Node, release/sanitized native, JavaScript, leaks and lane counts pass; whole-table counts updater fails outside this group.
Mutants: all seven native/JavaScript callable and deferred payload mutations caught across applicable pairs and restored; missing-rank verifier input refused.
Uncovered: 39/2818 pairs and 1474/11063 candidate reads certified; 2779 pairs and 9589 reads remain; original binding and Union conventions remain uncertified.

Reporting date October 12. Original source pin 050880ce59e30b356b686bd3144efe24f875ebc8. Four pairs each carry 21 candidate reads: NodeFactory.createTypeReferenceNode, Program.getCompilerOptions, System.write and TransformationContext.getCompilerOptions. Fixtures retain complete declarations and original member reads; adjacent carriers and helpers are reduced. These certify member contracts, not full original compiler execution or exact production reachability.

Type-reference uses its original string | EntityName parameter and optional readonly TypeNode array. Both scalar and object names, omitted and explicit undefined arrays, supplied arrays and an element value read are exercised. A string-only producer refuses at the callable member read. A compatible array producer declaring a string element field receives the deferred typeNode.value check at that actual read. The initial helper used optional indexed access, which stage 0 does not lower; explicit narrowing of the adjacent helper preserves Node behavior and the original member call.

Program and context options return aggregates. Both valid string value reads and incompatible numeric payload reads are pinned; the returned value is checked at its field read, not admitted recursively at the cast. System.write uses the original string-to-void contract and observable parameter length; void results are discarded. No process or system implementation is claimed certified.

The original source verifier now resolves interface members through their declared base interfaces. Program.getCompilerOptions is witnessed through ScriptReferenceHost; TransformationContext.getCompilerOptions through CoreTransformationContext. The actual declaration owner and full inheritance path are recorded in later-ranked-original-witnesses.json. Requested ranks missing from the inventory or evidence are refused rather than accepting a partial selection. A missing rank control exited one with missing requested original member. Initial failed verifier/helper attempts remain in raw logs and are not certification evidence.

Mutants: native-arity and javascript-arity caught all four pairs by early member-read pins versus execution. Native-result and javascript-result caught type-reference, program-options and context-options by result-representation pins versus later handling. Native-parameters and javascript-parameters caught type-reference and system-write by parameter-representation pins versus execution. Payload-reads caught type-reference and both options pairs by descendant string-read pins versus sanitizer output. All seven mutations restored; 21 pair-level checks in seven runs. No build or clang-warning failures counted.

Exact commands use source /workspace/adamic-tools/env.sh and direct log redirection:

```sh
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-original 61,67,68,69,70,72,73,74,75,76,77,78,79 later-ranked-original-witnesses.json > /tmp/lane5-options-original-final.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 76,77,78,79 >> /tmp/lane5-options-original-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(type-reference|program-options|context-options|system-write)$' -count=1 -timeout 5m > /tmp/lane5-options-oracle-final.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py type-reference program-options context-options system-write > /tmp/lane5-options-mutants.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 76,77,78,79 > /tmp/lane5-options-original-restored.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 76,999999 > /tmp/lane5-options-missing-rank.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-options-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-options-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-options-counts.log 2>&1
```

Original source verification passes for 13 evidence members/287 candidate reads and all 21 new fixtures. Focused oracle passed 9.591s. Restored later-ranked oracle passed 24.136s, retaining previous pairs and both refusal witnesses. Lane counts updater passed 31.220s: exactly 21 measured rows added and no existing row changes. Whole-table updater exited one in 40.613s on previously recorded outside-group lowering/count execution cases. Durable logs are under logs/type-reference-options-write.

Toolchain setup reused: done 417.425s, nproc 5, cgroup quota 4 CPUs; Go 1.27.1, clang 20.1.8 and Node 24.19.0. Static inventory unchanged: 4/308 pairs and 34/1503 reads certified. No compiler/runtime changes, calling-convention relaxation, full-package tests, full gate or PR.
