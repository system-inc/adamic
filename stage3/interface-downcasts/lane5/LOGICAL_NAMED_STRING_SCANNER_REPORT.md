Built: logical-and, named-export and string scanner callable member reads checked through views; arrow-body union read retained as a refusal.
Commits: follows pushed c0e852823efbc48e34f877cd654362c9ed8dae28 on codex/views-callables.
Checks: 21 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 21 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 22 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 62/2818 pairs and 1875/11063 candidate reads certified; 2756 pairs and 9188 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 115,116,117,118 add 60 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

createLogicalAnd retains its original Expression parameters and BinaryExpression result. Its left.value producer read is checked lazily. createNamedExports retains its readonly ExportSpecifier array parameter and NamedExports result. Empty and supplied arrays agree with Node, and a string-typed producer element receives a check at the reached element.value read. getTokenText and getTokenValue retain the original zero-argument string signatures and scanner reads. Numeric-result producers are refused at each callable member read; both scanner arity controls are also pinned.

Intrinsic Map/SymbolTable, RegExp and array method families encountered between these members remain uncertified. Prior Map intrinsic-method refusal evidence is in ORIGINAL_GROUP_REPORT.md; extending the Set representation/signature boundary to other intrinsic receivers is conservative inference, not a certificate of their original reads. Callable rest signatures and object-union callback conventions are not changed. No new integrator decision on the Union callable exception was received during this batch.

The next aggregate member, rank 120 createArrowFunction, remains uncertified. arrow-function/good.a retains its complete original six-parameter declaration and factory.createArrowFunction read. Its reduced ConciseBody carrier keeps Expression | Block. Source Node prints 3; lowering refuses the reached body.value read with unsupported untagged object union contract. This witness needs a proved union representation and descendant-read contract. TestCheckedViewCallableLaterRankedOriginalReadRefusals pins this outcome alongside ranks 102 and 106; it is excluded from counts and certification. No narrower alias or synthetic tag is used to claim the original union contract.

The oracle and restored commands also include |^TestCheckedViewCallableLaterRankedOriginalReadRefusals$ in the selection. The final complete later-ranked command uses ^TestCheckedViewCallableLaterRanked and passes in 85.820s; its direct log is preserved. Complete fixture verification includes all 40 source members and confirms 190 retained declarations/read spans. The next ranked supported-member candidates after the arrow boundary are createComma (121), updateVariableDeclaration (122) and readEmitHelpers (124), each with 14 candidate reads.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for logical-and, named-exports, token-text, token-value
javascript-arity: caught for logical-and, named-exports, token-text, token-value
native-result: caught for logical-and, named-exports, token-text, token-value
javascript-result: caught for logical-and, named-exports, token-text, token-value
native-parameters: caught for logical-and, named-exports
javascript-parameters: caught for logical-and, named-exports
payload-reads: caught for logical-and, named-exports
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/logical-named-string-scanner:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 115,116,117,118 > /tmp/lane5-logical-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(logical-and|named-exports|token-text|token-value)$' -count=1 -timeout 5m > /tmp/lane5-logical-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py logical-and named-exports token-text token-value > /tmp/lane5-logical-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(logical-and|named-exports|token-text|token-value)$' -count=1 -timeout 5m > /tmp/lane5-logical-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-logical-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-logical-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	9.553s
restored: ok  	github.com/system-inc/adamic/internal/oracle	10.369s
counts: ok  	github.com/system-inc/adamic/internal/oracle	47.012s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
