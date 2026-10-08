Built: original array bundle export import spread and type factory callable reads.
Commits: follows pushed e7297f1115a4035d454329914c3435a63d821622 on codex/views-callables.
Checks: 45 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 45 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 49 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 141/2818 pairs and 2609/11063 candidate reads certified; 2677 pairs and 8454 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 286,287,289,293,295,297,299 add 42 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 286,287,289,293,295,297,299 retain complete original NodeFactory declarations and witnessed reads. createBundle uses context.factory; createNamedImports uses nodeFactory. Other reads use factory. Adjacent AST carriers contain a reduced numeric value payload. No recursive original AST claim is made. Empty and populated source/import/type arrays are exercised; type members also exercise undefined. Result and producer parameter payloads are checked only when read.

The prior ranked receivers remain pending: 258 nominal Version; 259,260,261,267,269,270,279,280,281,282,283 collection/Date intrinsics and signatures; 262 AnyFunction callback, 263 generic asserted type; 264 namespace IdentifierNameMap with Identifier/PrivateIdentifier union; 265 tracing Phase/Args; 266 original stored CompilerHost method; 268 debug extension receiver; 271,272,273 optional host method binding/condition reads; 276 original Declaration/AnyImportSyntax union; 277 binding/assignment union; 278 Info alias and source declaration location; 284 optional host method; 285 inherited host method. These are pending inspection or proof, not newly observed refusals. Initializer helper 274, resolver 275 and updateCatchClause 300 original spans are recorded for the next group and not counted here. Factory ranks 288,290,291,292,294,296,298 remain next; no family is counted using a changed declaration.

Set needs the previously recorded collection-to-view conversion and intrinsic callable signature including its self result. Arrow needs the original ConciseBody union/read contract. Existing callback convention refusals remain intact. No integrator Union exception decision was received.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for array-type, export-default, not-emitted, spread-element, bundle, named-imports, type-literal
javascript-arity: caught for array-type, export-default, not-emitted, spread-element, bundle, named-imports, type-literal
native-result: caught for array-type, export-default, not-emitted, spread-element, bundle, named-imports, type-literal
javascript-result: caught for array-type, export-default, not-emitted, spread-element, bundle, named-imports, type-literal
native-parameters: caught for array-type, export-default, not-emitted, spread-element, bundle, named-imports, type-literal
javascript-parameters: caught for array-type, export-default, not-emitted, spread-element, bundle, named-imports, type-literal
payload-reads: caught for array-type, export-default, not-emitted, spread-element, bundle, named-imports, type-literal
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/array-bundle-export-import-spread-types:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 286,287,289,293,295,297,299 > /tmp/lane5-night-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(array-type|export-default|not-emitted|spread-element|bundle|named-imports|type-literal)$' -count=1 -timeout 5m > /tmp/lane5-night-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py array-type export-default not-emitted spread-element bundle named-imports type-literal > /tmp/lane5-night-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(array-type|export-default|not-emitted|spread-element|bundle|named-imports|type-literal)$' -count=1 -timeout 5m > /tmp/lane5-night-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-night-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-night-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	19.589s
restored: ok  	github.com/system-inc/adamic/internal/oracle	20.618s
counts: ok  	github.com/system-inc/adamic/internal/oracle	105.084s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
