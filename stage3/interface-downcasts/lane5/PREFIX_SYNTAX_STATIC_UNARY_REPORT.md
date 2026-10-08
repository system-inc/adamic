Built: prefix operand, syntax kind formatter, static block and prefix unary callable member reads checked through views.
Commits: follows pushed de5826ab673a63bcc4e110b5fa3b29b79f6cf4be on codex/views-callables.
Checks: 25 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 25 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 27 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 81/2818 pairs and 2113/11063 candidate reads certified; 2737 pairs and 8950 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 148,151,154,155 add 46 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

parenthesizeOperandOfPrefixUnary retains parenthesizerRules().parenthesizeOperandOfPrefixUnary and its original aggregate parameter/result declaration. operand.value is checked lazily. formatSyntaxKind records the original exported Debug function header separately from the field method signature, and covers original enum values 80/Identifier, 0/Unknown and undefined. Its SyntaxKind carrier is reduced to number and incompatible scalar producers stop at the member read.

createClassStaticBlockDeclaration checks its Block parameter and aggregate result; a reached body.value string producer is checked before output. createPrefixUnaryExpression retains PrefixUnaryOperator and Expression, with the scalar enum carrier reduced to number. PlusToken 40 and MinusToken 41 are confirmed against the original checker and both execute against Node; operand.value is checked lazily.

Rank 146 remains a refusal witness outside certification and native counts. for-update/good.a preserves the original updateForStatement declaration, context.factory.updateForStatement read and ForInitializer = VariableDeclarationList | Expression. The VariableDeclarationList kind is the original 262; Expression's kind remains broad. Node prints 15; lowering refuses the reached initializer.value read with unsupported untagged object union contract. This needs a proved representation and descendant-read contract for that original broad union, rather than a synthetic discriminator or a narrower alias. The original-read refusal test pins Node and lowering, while work continues on following supported families.

Fixture verification includes rank 146 as well as the four certified pairs. Oracle and restored commands also include |^TestCheckedViewCallableLaterRankedOriginalReadRefusals$ in the selection. Source enum values and the original Debug function header are preserved in the logs. Arrow rank 120 remains skipped as instructed, with its own recorded union-read prerequisite and refusal pin intact.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for prefix-operand, syntax-kind, static-block, prefix-unary
javascript-arity: caught for prefix-operand, syntax-kind, static-block, prefix-unary
native-result: caught for prefix-operand, syntax-kind, static-block, prefix-unary
javascript-result: caught for prefix-operand, syntax-kind, static-block, prefix-unary
native-parameters: caught for prefix-operand, syntax-kind, static-block, prefix-unary
javascript-parameters: caught for prefix-operand, syntax-kind, static-block, prefix-unary
payload-reads: caught for prefix-operand, static-block, prefix-unary
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/prefix-syntax-static-unary:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 148,151,154,155 > /tmp/lane5-unary-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(prefix-operand|syntax-kind|static-block|prefix-unary)$' -count=1 -timeout 5m > /tmp/lane5-unary-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py prefix-operand syntax-kind static-block prefix-unary > /tmp/lane5-unary-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(prefix-operand|syntax-kind|static-block|prefix-unary)$' -count=1 -timeout 5m > /tmp/lane5-unary-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-unary-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-unary-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	13.717s
restored: ok  	github.com/system-inc/adamic/internal/oracle	12.447s
counts: ok  	github.com/system-inc/adamic/internal/oracle	62.382s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
