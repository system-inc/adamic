Built: optional Boolean writer and zero-argument watcher callable member reads checked through views.
Commits: follows pushed 69883a0e2b2df1363c939445c437ac83fab4bec7 on codex/views-callables.
Checks: 8 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 8 counts rows added with existing rows unchanged.
Mutants: 4 applicable runs caught and restored, 6 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 58/2818 pairs and 1815/11063 candidate reads certified; 2760 pairs and 9248 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 109,110 add 30 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

writeLine retains its original optional Boolean force parameter and writer.writeLine() read. Omitted, undefined, false and true arguments agree with Node and exercise both output branches. A zero-argument producer is refused at the viewed member read despite source Node accepting the call; a string force producer is refused by parameter-representation metadata. close retains watcher.close() and its original zero-argument void signature. Its arity control likewise stops at the member read. Void results have no separate incompatible-result fixture, and no aggregate payload check applies to these two members.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for writer-line, watcher-close
javascript-arity: caught for writer-line, watcher-close
native-parameters: caught for writer-line
javascript-parameters: caught for writer-line
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/writer-line-watcher-close:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 109,110 > /tmp/lane5-writer-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(writer-line|watcher-close)$' -count=1 -timeout 5m > /tmp/lane5-writer-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py writer-line watcher-close > /tmp/lane5-writer-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(writer-line|watcher-close)$' -count=1 -timeout 5m > /tmp/lane5-writer-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-writer-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-writer-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	3.629s
restored: ok  	github.com/system-inc/adamic/internal/oracle	3.702s
counts: ok  	github.com/system-inc/adamic/internal/oracle	45.907s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
