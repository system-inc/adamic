Built: original tagged parameter and property declaration update contracts.
Commits: follows pushed 428f10a1b8abaabc4772839ea7e2fd2346bebb3f on codex/views-callables.
Checks: 14 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 14 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 14 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 91/2818 pairs and 2224/11063 candidate reads certified; 2727 pairs and 8839 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 159,160 add 22 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 159 and 160 certify the complete original updateParameterDeclaration and updatePropertyDeclaration signatures and factory member reads. BindingName/BindingPattern and PropertyName aliases retain every original alternative and original numeric kind discriminator. Token kinds 26,58,54 are independently resolved from the original checker; changing ExclamationToken kind 54 to 55 is rejected by the new token assertion and restored. Structural payloads, enum-adjacent carriers and helpers are reduced. Valid fixtures cover all name alternatives, string names, missing optional carriers, arrays, both property token variants and supplied parameter tokens. Settings, source compiler reachability and literal-subtype producer admission are outside this group's scope.

Seven runtime runs catch fourteen pair-level assertions: native/JavaScript arity, result and parameters, plus deferred node.value reads for both pairs. No compiler guard or callback convention changes. Original token verification command: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 159,160; its independent fixture-token mutant log is retained separately.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for parameter-update, property-update
javascript-arity: caught for parameter-update, property-update
native-result: caught for parameter-update, property-update
javascript-result: caught for parameter-update, property-update
native-parameters: caught for parameter-update, property-update
javascript-parameters: caught for parameter-update, property-update
payload-reads: caught for parameter-update, property-update
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/tagged-parameter-property-updates:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 159,160 > /tmp/lane5-declaration-updates-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(parameter-update|property-update)$' -count=1 -timeout 5m > /tmp/lane5-declaration-updates-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py parameter-update property-update > /tmp/lane5-declaration-updates-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(parameter-update|property-update)$' -count=1 -timeout 5m > /tmp/lane5-declaration-updates-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-declaration-updates-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-declaration-updates-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	7.124s
restored: ok  	github.com/system-inc/adamic/internal/oracle	8.156s
counts: ok  	github.com/system-inc/adamic/internal/oracle	69.826s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
