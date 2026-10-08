Built: original class constructor template update scanner and context callable checks.
Commits: follows pushed 70010b53191f400610929e16698079668ce76043 on codex/views-callables.
Checks: 34 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 34 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 32 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 134/2818 pairs and 2567/11063 candidate reads certified; 2684 pairs and 8496 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 246,247,249,251,256,257 add 42 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 246,247,249,251,256,257 certify complete original class-expression, constructor and tagged-template update declarations/read expressions, scanner setSkipJsDocLeadingAsterisks, getEmitHost and resumeLexicalEnvironment. Structural carriers and helper bodies are reduced. All optional and array parameter slots are retained. Valid cases cover absent/empty/supplied arrays and carriers, both scanner booleans and lazy returned emit-host payloads. TemplateLiteral preserves TemplateExpression | NoSubstitutionTemplateLiteral and original kinds 229 and 15; both alternatives run against Node and both backends. Changing TemplateExpression 229 to 228 fails the original discriminator assertion and is restored, with separate source-mutant evidence.

Seven runtime runs catch 32 pair-level assertions: native/JavaScript arity for six, result for four, parameters for four, and deferred aggregate/returned payload reads for four. No production compiler or callback calling convention changes. Carrier verification command: node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 249.

The earlier optional-method probes are expanded with an original required member: CompilerHost.getCurrentDirectory(): string from types.ts and DirectoryStructureHost.fileExists(path: string): boolean from watchUtilities.ts. Their source Node outputs remain abc/3 and true and both CLI c probes still refuse the cast with adamic/no-unchecked-cast. The original optional method declarations and binding/condition read contexts remain intact. This distinguishes the earlier optional-member-only reductions from these expanded reductions; it does not claim that every field of the original complete hosts was reproduced or that the original hosts are globally blocked. Both probes remain excluded from certification and measured counts. Commands use go run ./cmd/adamic c and node --disable-warning=ExperimentalWarning oracle/node.mjs with each retained probe path; logs are named realpath-required-probe/node and directory-required-probe/node.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for class-expression-update, constructor-update, tagged-template-update, scanner-jsdoc, context-emit-host, lexical-resume
javascript-arity: caught for class-expression-update, constructor-update, tagged-template-update, scanner-jsdoc, context-emit-host, lexical-resume
native-result: caught for class-expression-update, constructor-update, tagged-template-update, context-emit-host
javascript-result: caught for class-expression-update, constructor-update, tagged-template-update, context-emit-host
native-parameters: caught for class-expression-update, constructor-update, tagged-template-update, scanner-jsdoc
javascript-parameters: caught for class-expression-update, constructor-update, tagged-template-update, scanner-jsdoc
payload-reads: caught for class-expression-update, constructor-update, tagged-template-update, context-emit-host
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/class-constructor-template-scanner-context:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 246,247,249,251,256,257 > /tmp/lane5-updates-scanner-context-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(class-expression-update|constructor-update|tagged-template-update|scanner-jsdoc|context-emit-host|lexical-resume)$' -count=1 -timeout 5m > /tmp/lane5-updates-scanner-context-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py class-expression-update constructor-update tagged-template-update scanner-jsdoc context-emit-host lexical-resume > /tmp/lane5-updates-scanner-context-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(class-expression-update|constructor-update|tagged-template-update|scanner-jsdoc|context-emit-host|lexical-resume)$' -count=1 -timeout 5m > /tmp/lane5-updates-scanner-context-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-updates-scanner-context-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-updates-scanner-context-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	16.565s
restored: ok  	github.com/system-inc/adamic/internal/oracle	15.151s
counts: ok  	github.com/system-inc/adamic/internal/oracle	105.236s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
