Built: original initializer helper referenced import and catch update callable reads.
Commits: follows pushed 7d406f6c1f0f2249b81f5fe4f1fdb6124ef9be43 on codex/views-callables.
Checks: 21 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 21 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 21 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 144/2818 pairs and 2627/11063 candidate reads certified; 2674 pairs and 8436 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 274,275,300 add 18 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 274,275,300 retain complete original EmitHelperFactory.createRunInitializersHelper, EmitResolver.getReferencedImportDeclaration and NodeFactory.updateCatchClause declarations and read expressions. emitHelpers() and context.factory are retained as read contexts, with reduced producer/helper bodies. Numeric value carriers are reduced; full AST execution is not claimed. Missing, explicit undefined and supplied initializer arguments are exercised; absent/present resolver results are exercised; absent/present catch variable declarations are exercised. Parameter/result reads are lazy.

The first reduced producer attempt used string/number additions and mismatched Node expectations. It was corrected before certification to direct payload returns and constant incompatible-parameter results. No failed compilation is counted as a mutant. The first original-source verifier lookup used types.ts for EmitHelperFactory; it now reads its actual factory/emitHelpers.ts declaration. Set and arrow stay deferred under their previously recorded requirements. All earlier callback convention refusals remain. No integrator Union exception decision received.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for run-initializers, referenced-import, catch-update
javascript-arity: caught for run-initializers, referenced-import, catch-update
native-result: caught for run-initializers, referenced-import, catch-update
javascript-result: caught for run-initializers, referenced-import, catch-update
native-parameters: caught for run-initializers, referenced-import, catch-update
javascript-parameters: caught for run-initializers, referenced-import, catch-update
payload-reads: caught for run-initializers, referenced-import, catch-update
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/initializer-import-catch:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 274,275,300 > /tmp/lane5-initializer-import-catch-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(run-initializers|referenced-import|catch-update)$' -count=1 -timeout 5m > /tmp/lane5-initializer-import-catch-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py run-initializers referenced-import catch-update > /tmp/lane5-initializer-import-catch-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(run-initializers|referenced-import|catch-update)$' -count=1 -timeout 5m > /tmp/lane5-initializer-import-catch-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-initializer-import-catch-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-initializer-import-catch-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	10.113s
restored: ok  	github.com/system-inc/adamic/internal/oracle	9.508s
counts: ok  	github.com/system-inc/adamic/internal/oracle	108.894s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
