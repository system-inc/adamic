Built: original class and function creation and function expression update callable reads.
Commits: follows pushed ae16a6605f0f0f2f16e8d2fc14b75ca7f912346e on codex/views-callables.
Checks: 21 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 21 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 21 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 156/2818 pairs and 2699/11063 candidate reads certified; 2662 pairs and 8364 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 288,290,304 add 18 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 288,290,304 retain complete original createClassExpression, createFunctionDeclaration and updateFunctionExpression signatures and witnessed factory/context.factory member reads. Adjacent AST carriers and producer bodies are reduced. Class/function creation exercises string, Identifier and undefined names; absent, empty and supplied modifier/type-parameter/member/parameter arrays are exercised. Function-expression update preserves its required Block body and exercises missing optional arguments with that body supplied. Array producer payload reads remain lazy and pinned by string-read refusals. No recursive original AST or full compiler execution is claimed.

These three families were previously pending in this batch and are now certified. Original get-accessor, import-equals, export/method/property updates still need their complete original tagged aliases and signatures. Rank 294's original destructuring method read remains pending calling-context proof. Static class, debug intersection, nominal, higher-order, generic/predicate, optional-host binding/condition and intrinsic families remain pending as the other group reports detail. Set/arrow remain deferred; existing callback conventions are unchanged. No integrator Union exception decision was received.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for class-expression-create, function-declaration-create, function-expression-update
javascript-arity: caught for class-expression-create, function-declaration-create, function-expression-update
native-result: caught for class-expression-create, function-declaration-create, function-expression-update
javascript-result: caught for class-expression-create, function-declaration-create, function-expression-update
native-parameters: caught for class-expression-create, function-declaration-create, function-expression-update
javascript-parameters: caught for class-expression-create, function-declaration-create, function-expression-update
payload-reads: caught for class-expression-create, function-declaration-create, function-expression-update
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/class-function-create-update:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 288,290,304 > /tmp/lane5-class-function-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(class-expression-create|function-declaration-create|function-expression-update)$' -count=1 -timeout 5m > /tmp/lane5-class-function-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py class-expression-create function-declaration-create function-expression-update > /tmp/lane5-class-function-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(class-expression-create|function-declaration-create|function-expression-update)$' -count=1 -timeout 5m > /tmp/lane5-class-function-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-class-function-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-class-function-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	12.005s
restored: ok  	github.com/system-inc/adamic/internal/oracle	11.137s
counts: ok  	github.com/system-inc/adamic/internal/oracle	116.980s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
