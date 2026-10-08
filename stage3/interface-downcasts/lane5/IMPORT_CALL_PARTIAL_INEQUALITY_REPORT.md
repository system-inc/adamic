Built: original import declaration function call partial emission and strict inequality callable contracts.
Commits: follows pushed 433e1ec09fdba7c83fe2a3d5153cf28962a23022 on codex/views-callables.
Checks: 27 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 27 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 28 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 95/2818 pairs and 2261/11063 candidate reads certified; 2723 pairs and 8802 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 173,185,187,188 add 37 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 173,185,187,188 certify complete original createImportDeclaration, createFunctionCallCall, createPartiallyEmittedExpression and createStrictInequality declarations and original nodeFactory/factory reads. Structural carriers and producer bodies are reduced to numeric value fields. Optional arguments are exercised absent, explicitly undefined and present; arrays include empty and supplied elements. The function-call helper uses deferred element reads. Each producer payload string read is checked lazily.

Seven runtime runs cover 28 pair-level assertions. Original source verification also retains rank 165 onEmitNode as a function property with its stored previousOnEmitNode read. The compatible reduced implementation prints 4 in source Node, release native and generated JavaScript via oracle/node.mjs. The wrong-arity implementation prints 9 on source Node but native build refuses the reached onEmitNode read with unsupported callable contract. This pair remains uncertified and excluded from counts; a full nested callback signature/producer certificate is still needed. No existing callback guard changed. The initial direct generated-JavaScript run lacked the adamic package mapping; rerunning with the repository oracle loader prints 4. Its earlier loader error is not a callable result. The wrong-arity native binary was not produced, and its build refusal is the observation used.

Rank 162 remains pending a certificate that elaborates Required<Pick<SymbolTracker, "reportInferenceFallback">> while preserving the original optional SymbolTracker declaration; it is not counted through a rewritten required method. Intrinsic array/Map/Set/SymbolTable methods, external writeSync overloads, generic createModifier, predicate Array.isArray and overload createImportClause/createYieldExpression remain uncertified. Module/name unions, createPropertyDeclaration, class-this, dynamic callable elements and the host intersection still need their original carrier/read fixtures. This records pending work, not observed refusals for untested families.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for import-declaration, function-call-call, partial-expression, strict-inequality
javascript-arity: caught for import-declaration, function-call-call, partial-expression, strict-inequality
native-result: caught for import-declaration, function-call-call, partial-expression, strict-inequality
javascript-result: caught for import-declaration, function-call-call, partial-expression, strict-inequality
native-parameters: caught for import-declaration, function-call-call, partial-expression, strict-inequality
javascript-parameters: caught for import-declaration, function-call-call, partial-expression, strict-inequality
payload-reads: caught for import-declaration, function-call-call, partial-expression, strict-inequality
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/import-call-partial-inequality:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 173,185,187,188 > /tmp/lane5-expression-members-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(import-declaration|function-call-call|partial-expression|strict-inequality)$' -count=1 -timeout 5m > /tmp/lane5-expression-members-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py import-declaration function-call-call partial-expression strict-inequality > /tmp/lane5-expression-members-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(import-declaration|function-call-call|partial-expression|strict-inequality)$' -count=1 -timeout 5m > /tmp/lane5-expression-members-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-expression-members-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-expression-members-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	12.857s
restored: ok  	github.com/system-inc/adamic/internal/oracle	12.627s
counts: ok  	github.com/system-inc/adamic/internal/oracle	73.882s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
