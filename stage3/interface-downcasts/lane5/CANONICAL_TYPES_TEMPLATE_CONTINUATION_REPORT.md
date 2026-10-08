Built: three groups certify 16 additional callable pairs, 113 candidate reads and 94 fixtures.
Commits: 343df8fb, 70010b53 and bb275f0f, each pushed separately to codex/views-callables; this commit records final validation.
Checks: original declarations/read spans, callable aliases/tags, Node, release/sanitized native, JavaScript, leaks, measured lane counts and complete later-ranked harness pass; whole-table counts fails outside lane.
Mutants: 21 runtime runs catch 94 pair-level assertions; two original-source verifier mutants caught; all sources restored.
Uncovered: 134/2818 pairs and 2567/11063 candidate reads certified; 2684 pairs and 8496 reads remain; optional-method cast, original union, callback, intrinsic, mapped and other boundaries remain excluded.

Reporting date October 12. Continued from 554cb1122f0d6e164f14a9b50848b7550177c2db. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Candidate reads are inventory counts; exact original compiler execution remains unmeasured.

| Group | Commit | New pairs / reads | Fixtures | Runtime runs / assertions |
|---|---|---:|---:|---:|
| Canonical names, emit options and package cache | 343df8fb | 3 / 22 | 15 | 7 / 16 |
| Conditional/constructor/export/function/import/omitted/union factories | 70010b53 | 7 / 49 | 45 | 7 / 46 |
| Class/constructor/template updates, scanner and context | bb275f0f | 6 / 42 | 34 | 7 / 32 |

Certified ranks: 217,227,234,237,238,239,240,241,242,243,246,247,249,251,256,257. Complete original declarations and member read expressions are retained, including the original GetCanonicalFileName callable alias and getReferencedFiles argument read. The alias is independently checked in its original core.ts source. ModuleExportName and TemplateLiteral retain all original alternatives and numeric tags, and each is exercised. boolean | undefined export flags cover undefined, false and true; a boolean-only producer is refused. Optional/empty/supplied arrays, optional results and reached payload reads remain lazy. Adjacent structural carriers and helper/producer bodies are reduced as each group report specifies.

Every runtime mutation is listed with affected families in the three group reports and raw logs. Native/JavaScript arity, result and parameter mutations are caught at viewed member reads; omitted payload registration is caught at descendant string reads, including sanitizer evidence. No compilation failure counts as a runtime mutant result. Changing GetCanonicalFileName's fixture parameter from string to number fails its original alias assertion. Changing TemplateExpression's fixture kind 229 to 228 fails its original tag assertion. Both are restored. No production compiler/runtime or calling-convention guard changes remain.

Ranks 223 and 225 retain original method contexts in separate uncertified realpath-binding and directory-condition probes: host.realpath?.bind(host) and host.directoryExists && directoryExists. Corrected source Node outputs are abc/3 and true. Both CLI c commands refuse the cast with adamic/no-unchecked-cast. The retained probes additionally include original required CompilerHost.getCurrentDirectory and DirectoryStructureHost.fileExists members; the expanded reductions produce the same outputs and cast refusals. These observed cast boundaries require checked-view admission and proof/preservation of the original binding/condition use; no method was rewritten to a function property. The complete original hosts are not reproduced, so this does not establish a global block for those hosts. They are excluded from certification and counts. The directory probe's initial boolean console argument was corrected to interpolation before the final observation. This is not an observed unbound-method diagnostic; source read context and exact refusal are recorded in CANONICAL_EMIT_PACKAGE_HOSTS_REPORT.md.

Set stays skipped as instructed and needs collection-to-view conversion plus recorded intrinsic callable signatures, including add's self return. Arrow rank 120 stays skipped and needs the original ConciseBody union/read contract. Earlier ForInitializer, higher-order hook, mapped receiver and intrinsic boundaries remain recorded. ModuleResolutionHost's boolean/function/undefined useCaseSensitiveFileNames contract at rank 235 and higher-order createCallBinding at rank 236 remain uncertified. Overloads, generics, original ForInitializer updateForInStatement and nominal SymbolTrackerImpl/Version receiver families also remain pending; no new observed refusal is claimed for untested families. No single overload, plain-object intrinsic substitute or rewritten mapped/nominal member is counted. No integrator decision on the Union callable exception was received. Narrower literal-subtype producer admission, recursive original AST payloads and whole original compiler execution remain outside these represented contract certificates.

Exact focused commands and outcomes are in CANONICAL_EMIT_PACKAGE_HOSTS_REPORT.md, CONDITIONAL_CONSTRUCTOR_EXPORT_FUNCTION_IMPORT_UNION_REPORT.md and CLASS_CONSTRUCTOR_TEMPLATE_SCANNER_CONTEXT_REPORT.md. Final checks:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 10m > /tmp/lane5-canonical-types-batch-all-restored.log 2>&1
go vet ./internal/oracle > /tmp/lane5-canonical-types-batch-vet.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original "$(cat /tmp/lane5-updates-scanner-context-ranks.txt)" > /tmp/lane5-canonical-types-batch-all-fixtures.log 2>&1
node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 122,141,144,145,159,160,176,177,192,194,212,217,241,249 > /tmp/lane5-canonical-types-batch-all-carriers.log 2>&1
git diff --check
```

ok  	github.com/system-inc/adamic/internal/oracle	289.455s
Verified 626 fixtures retain complete original declarations and reads.
Verified 96 fixtures retain requested original aliases, discriminators and enum values.

Oracle vet and formatting pass without output. All 94 counts rows are appended with the previous table unchanged. Each required whole-table updater fails on outside-lane cases recorded in its complete log; all three lane-specific updaters pass. No whole-package tests or full gate. Source evidence covers 117 members and 1497 candidate reads, including the two excluded optional-host probes. Static inventory remains 4/308 pairs and 34/1503 reads certified, 304/1469 remaining. Setup reused: done 417.425s, nproc 5, quota 4 CPUs; complete timings remain in earlier reports. Only codex/views-callables pushed, no PR opened.

Group full SHAs:
343df8fbbc249cf294c3766858bb4e9ed4ee78b5
70010b53191f400610929e16698079668ce76043
bb275f0f5d518373232eb35c253616486f5e60d9
