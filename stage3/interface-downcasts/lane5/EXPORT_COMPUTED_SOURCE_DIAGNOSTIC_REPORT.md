Built: export declaration, computed property name, source-file update and diagnostic member contracts checked through views.
Commits: follows pushed 868325297f09a4a1e3b9ee19c4138b6582d7752e on codex/views-callables.
Checks: 25 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 25 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 26 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 54/2818 pairs and 1753/11063 candidate reads certified; 2764 pairs and 9310 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 98,99,100,101 add 68 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

createExportDeclaration retains all five original parameters and tests omitted optional values and supplied modifier, expression and attributes carriers. Its scalar/array parameter alternatives are checked at the member read; an exportClause.value payload is checked lazily. updateComputedPropertyName preserves context.factory.updateComputedPropertyName and checks the reached node.value producer read. updateSourceFile retains all seven parameters, including optional arrays and both Boolean flags, and checks a supplied statement element when statement.value is reached. addDiagnostic preserves the original diagnostic creation call and checks diag.value when read by the producer.

Source verification follows FileWatcher to sys.ts for the upcoming group. Prepared evidence includes subsequent ranks through 118 and records inherited writer declarations. The failed first FileWatcher lookup and corrected source extraction are retained in the logs. Map/SymbolTable and other intrinsic families remain outside these object-member certificates; the existing intrinsic-member refusal evidence in ORIGINAL_GROUP_REPORT.md remains applicable, with no fabricated plain-object producer counted for an intrinsic pair.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for export-declaration, computed-name, source-file-update, context-diagnostic
javascript-arity: caught for export-declaration, computed-name, source-file-update, context-diagnostic
native-result: caught for export-declaration, computed-name, source-file-update
javascript-result: caught for export-declaration, computed-name, source-file-update
native-parameters: caught for export-declaration, computed-name, source-file-update, context-diagnostic
javascript-parameters: caught for export-declaration, computed-name, source-file-update, context-diagnostic
payload-reads: caught for export-declaration, computed-name, source-file-update, context-diagnostic
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/export-computed-source-diagnostic:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 98,99,100,101 > /tmp/lane5-export-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(export-declaration|computed-name|source-file-update|context-diagnostic)$' -count=1 -timeout 5m > /tmp/lane5-export-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py export-declaration computed-name source-file-update context-diagnostic > /tmp/lane5-export-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(export-declaration|computed-name|source-file-update|context-diagnostic)$' -count=1 -timeout 5m > /tmp/lane5-export-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-export-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-export-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	12.775s
restored: ok  	github.com/system-inc/adamic/internal/oracle	12.522s
counts: ok  	github.com/system-inc/adamic/internal/oracle	43.863s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
