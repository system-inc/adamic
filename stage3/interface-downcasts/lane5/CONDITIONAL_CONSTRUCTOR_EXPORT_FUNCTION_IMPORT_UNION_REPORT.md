Built: original conditional constructor export function import omitted and union-type callable checks.
Commits: follows pushed 343df8fbbc249cf294c3766858bb4e9ed4ee78b5 on codex/views-callables.
Checks: 45 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 45 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 46 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 128/2818 pairs and 2525/11063 candidate reads certified; 2690 pairs and 8538 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 237,238,239,240,241,242,243 add 49 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 237 through 243 certify the complete original conditional type, constructor declaration, export assignment, function type, import specifier, omitted expression and union-type factory member declarations/read expressions. Adjacent structural carriers and helper bodies are reduced. ModuleExportName retains Identifier | StringLiteral and their original kinds 80 and 11; both alternatives run against Node and both backends. createExportAssignment preserves boolean | undefined and tests undefined, false and true. Its producer requiring boolean is refused by the original parameter admission check. Optional/empty/supplied arrays and aggregate descendant reads remain lazy.

Seven runtime runs catch 46 pair-level assertions: native/JavaScript arity and result for all seven, parameters for six, and string payload reads for six. All sources restored. Original alias/tag verification command: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 241. No compiler or calling-convention guards change; narrower literal-subtype producers and recursive original AST payloads are not certified by these reductions.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for conditional-type, constructor-declaration, export-assignment, function-type, import-specifier, omitted-expression, union-type
javascript-arity: caught for conditional-type, constructor-declaration, export-assignment, function-type, import-specifier, omitted-expression, union-type
native-result: caught for conditional-type, constructor-declaration, export-assignment, function-type, import-specifier, omitted-expression, union-type
javascript-result: caught for conditional-type, constructor-declaration, export-assignment, function-type, import-specifier, omitted-expression, union-type
native-parameters: caught for conditional-type, constructor-declaration, export-assignment, function-type, import-specifier, union-type
javascript-parameters: caught for conditional-type, constructor-declaration, export-assignment, function-type, import-specifier, union-type
payload-reads: caught for conditional-type, constructor-declaration, export-assignment, function-type, import-specifier, union-type
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/conditional-constructor-export-function-import-union:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 237,238,239,240,241,242,243 > /tmp/lane5-types-construct-export-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(conditional-type|constructor-declaration|export-assignment|function-type|import-specifier|omitted-expression|union-type)$' -count=1 -timeout 5m > /tmp/lane5-types-construct-export-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py conditional-type constructor-declaration export-assignment function-type import-specifier omitted-expression union-type > /tmp/lane5-types-construct-export-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(conditional-type|constructor-declaration|export-assignment|function-type|import-specifier|omitted-expression|union-type)$' -count=1 -timeout 5m > /tmp/lane5-types-construct-export-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-types-construct-export-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-types-construct-export-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	20.557s
restored: ok  	github.com/system-inc/adamic/internal/oracle	20.283s
counts: ok  	github.com/system-inc/adamic/internal/oracle	94.932s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
