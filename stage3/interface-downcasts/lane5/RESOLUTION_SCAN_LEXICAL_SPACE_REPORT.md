Built: original resolution settings, scanner advancement, lexical entry and writer spacing checks.
Commits: follows pushed 8a5b672e6a892ea96002c4481b410e28048ca9ef on codex/views-callables.
Checks: 16 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 16 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 15 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 89/2818 pairs and 2202/11063 candidate reads certified; 2729 pairs and 8861 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 163,164,166,170 add 43 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 163,164,166,170 certify getCompilationSettings, scan, startLexicalEnvironment and writeSpace. CompilerOptions is reduced to a string value field; SyntaxKind to its numeric representation. Helpers and producers are reduced. Original complete declarations and read expressions are retained, including inherited CoreTransformationContext and SymbolWriter owners. Settings payload reads remain deferred. No production compiler changes or callback convention changes.

Seven runtime mutant runs cover 15 pair-level assertions: native and JavaScript arity (four each), result (two each), parameters (writer each), and payload reads (settings). The whole-table counts log records outside-lane failures; the lane updater passes.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for resolution-settings, scanner-scan, lexical-start, writer-space
javascript-arity: caught for resolution-settings, scanner-scan, lexical-start, writer-space
native-result: caught for resolution-settings, scanner-scan
javascript-result: caught for resolution-settings, scanner-scan
native-parameters: caught for writer-space
javascript-parameters: caught for writer-space
payload-reads: caught for resolution-settings
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/resolution-scan-lexical-space:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 163,164,166,170 > /tmp/lane5-resolution-scan-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(resolution-settings|scanner-scan|lexical-start|writer-space)$' -count=1 -timeout 5m > /tmp/lane5-resolution-scan-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py resolution-settings scanner-scan lexical-start writer-space > /tmp/lane5-resolution-scan-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(resolution-settings|scanner-scan|lexical-start|writer-space)$' -count=1 -timeout 5m > /tmp/lane5-resolution-scan-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-resolution-scan-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-resolution-scan-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	7.304s
restored: ok  	github.com/system-inc/adamic/internal/oracle	6.961s
counts: ok  	github.com/system-inc/adamic/internal/oracle	65.106s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
