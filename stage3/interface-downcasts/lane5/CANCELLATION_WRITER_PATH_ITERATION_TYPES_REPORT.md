Built: original cancellation writer path iteration and type-expression callable checks.
Commits: follows pushed 4c2cc83ea713275fb4a52d3923accde1bb0fc1e9 on codex/views-callables.
Checks: 42 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 42 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 44 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 110/2818 pairs and 2390/11063 candidate reads certified; 2708 pairs and 8673 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 199,200,201,202,206,207,208,209 add 64 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 199,200,201,202,206,207,208,209 certify complete original cancellation, writer text, redirect-host path, iteration resolver, module specifier host directory, expression-with-type-arguments, indexed-access-type and logical-not declarations/read expressions. Path is reduced to its string representation; other adjacent carriers and helper bodies are reduced. Original IterationTypesResolver.resolveIterationType remains a function property with Type | undefined result, exercised absent and present. Optional errorNode and type argument arrays exercise absent, empty and supplied values. Aggregate reads remain deferred.

The original declaration verifier resolves HostForUseSourceOfProjectReferenceRedirect in program.ts and IterationTypesResolver in checker.ts. Redirecting the resolver to types.ts fails its original declaration-owner assertion; the source-verifier mutant is restored and logged separately. Seven runtime runs catch 44 pair-level assertions: native/JavaScript arity for eight, result for seven, parameters for five, and reached aggregate string reads for four. All sources restored. No production compiler or callback calling convention change.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for cancellation-check, writer-text, redirect-path, iteration-type, specifier-directory, expression-type-arguments, indexed-access-type, logical-not
javascript-arity: caught for cancellation-check, writer-text, redirect-path, iteration-type, specifier-directory, expression-type-arguments, indexed-access-type, logical-not
native-result: caught for writer-text, redirect-path, iteration-type, specifier-directory, expression-type-arguments, indexed-access-type, logical-not
javascript-result: caught for writer-text, redirect-path, iteration-type, specifier-directory, expression-type-arguments, indexed-access-type, logical-not
native-parameters: caught for redirect-path, iteration-type, expression-type-arguments, indexed-access-type, logical-not
javascript-parameters: caught for redirect-path, iteration-type, expression-type-arguments, indexed-access-type, logical-not
payload-reads: caught for iteration-type, expression-type-arguments, indexed-access-type, logical-not
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/cancellation-writer-path-iteration-types:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 199,200,201,202,206,207,208,209 > /tmp/lane5-cancellation-type-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(cancellation-check|writer-text|redirect-path|iteration-type|specifier-directory|expression-type-arguments|indexed-access-type|logical-not)$' -count=1 -timeout 5m > /tmp/lane5-cancellation-type-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py cancellation-check writer-text redirect-path iteration-type specifier-directory expression-type-arguments indexed-access-type logical-not > /tmp/lane5-cancellation-type-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(cancellation-check|writer-text|redirect-path|iteration-type|specifier-directory|expression-type-arguments|indexed-access-type|logical-not)$' -count=1 -timeout 5m > /tmp/lane5-cancellation-type-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-cancellation-type-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-cancellation-type-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	18.314s
restored: ok  	github.com/system-inc/adamic/internal/oracle	19.638s
counts: ok  	github.com/system-inc/adamic/internal/oracle	81.393s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
