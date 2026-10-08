Built: returned emit-helper array, case-sensitive filename and diagnostic newline callable member reads checked through views.
Commits: follows pushed 85d2852044c0efa2452ca104427c591408956703 on codex/views-callables.
Checks: 15 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 15 counts rows added with existing rows unchanged.
Mutants: 5 applicable runs caught and restored, 13 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 65/2818 pairs and 1915/11063 candidate reads certified; 2753 pairs and 9148 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 124,127,128 add 40 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

readEmitHelpers retains its original EmitHelper[] | undefined result and context.readEmitHelpers() read. Present and absent arrays agree with Node; a reached helper.value field is checked lazily. useCaseSensitiveFileNames retains its zero-argument Boolean declaration and host read, with true and false producers. Its arity control keeps a Boolean result and uses a conditional only to print 9, so only arity metadata differs. getNewLine retains FormatDiagnosticsHost's original string result and host.getNewLine() read.

The arrow family at rank 120 is skipped as instructed. Its existing original-signature witness needs a proved ConciseBody union representation and descendant-read contract; its refusal pin remains intact and it is excluded from counts and certification. Set remains deferred with its existing recorded collection-view and callable-signature prerequisites.

Rank 121 preserves the original reduceLeft(expressions, factory.createComma) read context. Node prints 7 and lowering refuses unbound-method. No direct-call replacement is used to certify that pair. Rank 122's first reduced witness lacks BindingName tags and refuses its reached name.value read. This is a reduced-carrier observation, not proof that the original BindingName is unsupported: original Identifier, ObjectBindingPattern and ArrayBindingPattern declarations provide distinct numeric kind tags. A subsequent tagged probe lowers successfully and is being prepared for the next group. The untagged reduction remains a separate refusal control.

Fixture verification included ranks 121,122,124,127,128 and confirmed 17 original declarations/reads: 15 certified fixtures plus two refusal probes. Oracle and restored commands also include |^TestCheckedViewCallableLaterRankedOriginalReadRefusals$ in the selection. Source interface resolution now follows FormatDiagnosticsHost to program.ts and ProgramDiagnostics to programDiagnostics.ts. The failed first lookup and corrected extraction are preserved; no substitute interface declaration was used.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for read-helpers, case-sensitive, diagnostic-newline
javascript-arity: caught for read-helpers, case-sensitive, diagnostic-newline
native-result: caught for read-helpers, case-sensitive, diagnostic-newline
javascript-result: caught for read-helpers, case-sensitive, diagnostic-newline
payload-reads: caught for read-helpers
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/helpers-case-newline:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 124,127,128 > /tmp/lane5-helperhost-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(read-helpers|case-sensitive|diagnostic-newline)$' -count=1 -timeout 5m > /tmp/lane5-helperhost-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py read-helpers case-sensitive diagnostic-newline > /tmp/lane5-helperhost-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(read-helpers|case-sensitive|diagnostic-newline)$' -count=1 -timeout 5m > /tmp/lane5-helperhost-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-helperhost-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-helperhost-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	7.347s
restored: ok  	github.com/system-inc/adamic/internal/oracle	7.182s
counts: ok  	github.com/system-inc/adamic/internal/oracle	52.786s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
