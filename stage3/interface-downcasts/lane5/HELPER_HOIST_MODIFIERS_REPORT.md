Built: getEmitHelperFactory, hoistVariableDeclaration and createModifiersFromModifierFlags checked through views, including deferred object and array element reads.
Commits: follows 619b65937ce3104353b0d9da570dc85e8545f524 on codex/views-callables.
Checks: 17 fixtures retain complete original declarations/member reads; Node, release/sanitized native, JavaScript and leaks pass; 17 counts rows added, existing rows unchanged.
Mutants: seven runs, 17 pair-level checks caught and restored by member-read or lazy descendant-read assertions.
Uncovered: 48/2818 pairs and 1649/11063 candidate reads certified; 2770 pairs and 9414 reads remain.

Reporting date October 12. Ranks 91, 92 and 94 contribute 19, 19 and 18 reads. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Adjacent carriers/helpers are reduced; production reachability remains unmeasured. The returned helper object and modifier array are checked at their reached value reads. The modifier result also covers undefined. The hoisted identifier payload is checked when the producer reads node.value.

Set ranks 89/90 are skipped as instructed. They need a supported collection-to-view representation conversion and recorded callable contracts for intrinsic add/has members, including the self result of add. Prior Set probes observe a native adamic_map*/adamic_object* conversion rejection and a JavaScript unknown-signature refusal. Map rank 93 is conservatively deferred as another intrinsic collection member; applying the Set boundary to it is an inference, not certification of the original Map read. Existing callback calling conventions remain unchanged.

Mutants: native-arity/javascript-arity caught for all three pairs; native-result/javascript-result for helper-factory and modifier-flags; native-parameters/javascript-parameters for hoist-variable and modifier-flags; payload-reads for all three pairs. Each mutant ran the relevant wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Payload mutants were caught at the lazy string reads with sanitizer evidence; no build failure counted. Every mutation was restored.

Commands, with direct output logs under logs/helper-hoist-modifiers:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-original 61,67,68,69,70,72,73,74,75,76,77,78,79,81,82,83,84,85,86,88,91,92,94,95,96 later-ranked-original-witnesses.json > /tmp/lane5-helper-original.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 91,92,94 > /tmp/lane5-helper-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(helper-factory|hoist-variable|modifier-flags)$' -count=1 -timeout 5m > /tmp/lane5-helper-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py helper-factory hoist-variable modifier-flags > /tmp/lane5-helper-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(helper-factory|hoist-variable|modifier-flags)$' -count=1 -timeout 5m > /tmp/lane5-helper-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-helper-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-helper-counts.log 2>&1
```

Initial 16-fixture oracle passed 7.919s. After adding the undefined result fixture, the restored 17-fixture oracle passed 8.296s. Lane counts passed 39.895s. Whole-table updater failed outside the group in 40.398s. Subsequent source extraction includes ranks 98-101 for the next group, recorded in /tmp/lane5-next-original.log. No production compiler changes or full gate. Setup reused: done 417.425s, nproc 5, quota 4 CPUs. Static certification stays 4/308 pairs and 34/1503 reads.
