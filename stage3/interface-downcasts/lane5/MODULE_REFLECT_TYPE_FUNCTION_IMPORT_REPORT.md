Built: original module reflect type function import and directory callable checks.
Commits: follows pushed f241a55934c5eac1fd9b8791c566bcac70740fc0 on codex/views-callables.
Checks: 50 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 50 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 50 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 118/2818 pairs and 2454/11063 candidate reads certified; 2700 pairs and 8609 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 210,211,212,213,215,216,218,221 add 64 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 210,211,212,213,215,216,218,221 certify complete original module block, reflect-get, type operator, type parameter, function/import update and Program/System directory getter declarations/read expressions. Structural carriers and helpers are reduced. Both context.factory update reads and factory creation reads retain original spelling. The eight-parameter updateFunctionDeclaration signature is complete. Valid cases exercise absent, empty and supplied arrays, optional reflection receiver, optional type-parameter bounds and default, string/Identifier name arguments and supplied/absent function carriers.

createTypeOperatorNode retains the original union of SyntaxKind.KeyOfKeyword, UniqueKeyword and ReadonlyKeyword. Independent original checker values 143,158,148 are validated in every type-operator fixture; all three run against Node and both backends. The carrier verifier now checks requested enum members. Changing fixture KeyOfKeyword 143 to 142 fails the original enum-value assertion and is restored, logged separately. This certifies represented signatures and reached reads, not every narrower literal-subtype producer or recursive original AST carrier.

Seven runtime runs catch fifty pair-level assertions: native/JavaScript arity and result for all eight, parameters for six, and reached statement.value, target.value, type.value, modifier.value, typeParameter.value and moduleSpecifier.value reads. No production compiler or existing callback calling-convention guards change. Enum verification command: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 212.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for module-block, reflect-get, type-operator, type-parameter, function-update, import-update, program-directory, system-directory
javascript-arity: caught for module-block, reflect-get, type-operator, type-parameter, function-update, import-update, program-directory, system-directory
native-result: caught for module-block, reflect-get, type-operator, type-parameter, function-update, import-update, program-directory, system-directory
javascript-result: caught for module-block, reflect-get, type-operator, type-parameter, function-update, import-update, program-directory, system-directory
native-parameters: caught for module-block, reflect-get, type-operator, type-parameter, function-update, import-update
javascript-parameters: caught for module-block, reflect-get, type-operator, type-parameter, function-update, import-update
payload-reads: caught for module-block, reflect-get, type-operator, type-parameter, function-update, import-update
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/module-reflect-type-function-import:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 210,211,212,213,215,216,218,221 > /tmp/lane5-module-function-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(module-block|reflect-get|type-operator|type-parameter|function-update|import-update|program-directory|system-directory)$' -count=1 -timeout 5m > /tmp/lane5-module-function-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py module-block reflect-get type-operator type-parameter function-update import-update program-directory system-directory > /tmp/lane5-module-function-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(module-block|reflect-get|type-operator|type-parameter|function-update|import-update|program-directory|system-directory)$' -count=1 -timeout 5m > /tmp/lane5-module-function-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-module-function-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-module-function-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	24.298s
restored: ok  	github.com/system-inc/adamic/internal/oracle	23.181s
counts: ok  	github.com/system-inc/adamic/internal/oracle	90.331s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
