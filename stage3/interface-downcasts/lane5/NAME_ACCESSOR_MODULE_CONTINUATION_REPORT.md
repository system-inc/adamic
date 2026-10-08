Built: four groups certify 23 additional callable pairs, 193 candidate reads and 138 fixtures.
Commits: 19b8c74a, 4c2cc83e, f241a559 and 77723f55, each pushed separately to codex/views-callables; this commit records final validation.
Checks: original declarations/read spans, aliases/tags/enum values, Node, release/sanitized native, JavaScript, leaks, measured lane counts and complete later-ranked harness pass; whole-table counts fails outside lane.
Mutants: 28 runtime runs catch 141 pair-level assertions; two original-source verifier mutants caught; all sources restored.
Uncovered: 118/2818 pairs and 2454/11063 candidate reads certified; 2700 pairs and 8609 reads remain; recorded union, callback, intrinsic, mapped and other pending families remain excluded.

Reporting date October 12. Continued from 490b24478c9155ac73d2c70d3043a1408a55cc9b. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Candidate reads are inventory counts; exact original compiler execution remains unmeasured.

| Group | Commit | New pairs / reads | Fixtures | Runtime runs / assertions |
|---|---|---:|---:|---:|
| Property declaration and qualified names | 19b8c74a | 2 / 20 | 14 | 7 / 14 |
| Outer expressions, accessor/parenthesized updates, host source files | 4c2cc83e | 5 / 45 | 32 | 7 / 33 |
| Cancellation, text, paths, iteration and type expressions | f241a559 | 8 / 64 | 42 | 7 / 44 |
| Module blocks, reflection, types, function/import updates, directories | 77723f55 | 8 / 64 | 50 | 7 / 50 |

Certified ranks: 176,177,191,192,193,194,196,199,200,201,202,206,207,208,209,210,211,212,213,215,216,218,221. Complete original member declarations and read expressions are retained. Adjacent structural payloads, helper/producer bodies and enum representations are reduced as each group specifies. PropertyName's seven alternatives and EntityName's two alternatives retain original tags and are all exercised. SyntaxKind.KeyOfKeyword, ReadonlyKeyword and UniqueKeyword retain original values 143,148,158 and all three declared operator alternatives run against Node. Original function-property style remains intact for resolveIterationType. Optional arguments, optional results, arrays and reached element/payload reads are tested lazily.

Every runtime mutation is listed with its affected families in the four group reports and raw logs. Native/JavaScript arity, result and parameter mutations are caught at member reads; payload registration mutations are caught at descendant string reads, including sanitizer evidence. No compilation failure counts as a runtime mutant result. Redirecting IterationTypesResolver's original declaration lookup to types.ts fails its source-owner assertion; changing KeyOfKeyword's fixture value from 143 to 142 fails its original enum-value assertion. Both source mutants are restored. No production compiler/runtime or calling-convention guard changes remain.

Set stays skipped as instructed and needs collection-to-view conversion and recorded intrinsic callable signatures, including add's self return. Arrow rank 120 stays skipped and needs a proved original ConciseBody union/read contract. Broad ForInitializer remains unproved; updateForOfStatement rank 214 needs an original carrier/read witness and is excluded, with no new observed refusal claimed for that untested pair. The existing original for-update refusal pin remains intact. The higher-order onEmitNode boundary and mapped Required<Pick<SymbolTracker, "reportInferenceFallback">> receiver remain as described in DECLARATION_EXPRESSION_CONTINUATION_REPORT.md. Intrinsic array/Map/Set/SymbolTable and overloaded/generic/predicate members remain uncertified.

Pending original families include ModuleBody/createModuleDeclaration, class-this onDiagnosticReported, dynamic callable elements, the CompilerHost & ReadBuildProgramHost intersection, generic markNodeReuse, the optional ModuleResolutionCache receiver and GetCanonicalFileName alias. These are pending certificates, not new observations of refusal. No plain-object intrinsic substitute, rewritten required mapped member or single selected overload is counted. No integrator decision on the Union exception was received. Narrower literal-subtype producer admission, recursive original AST payloads and whole original compiler execution are outside these represented contract certificates.

Exact focused commands and outcomes are in PROPERTY_DECLARATION_QUALIFIED_NAME_REPORT.md, OUTER_ACCESSOR_PARENTHESIZED_HOST_REPORT.md, CANCELLATION_WRITER_PATH_ITERATION_TYPES_REPORT.md and MODULE_REFLECT_TYPE_FUNCTION_IMPORT_REPORT.md. Final checks:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 10m > /tmp/lane5-name-module-batch-all-restored.log 2>&1
go vet ./internal/oracle > /tmp/lane5-name-module-batch-vet.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original "$(cat /tmp/lane5-module-function-ranks.txt)" > /tmp/lane5-name-module-batch-all-fixtures.log 2>&1
node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 122,141,144,145,159,160,176,177,192,194,212 > /tmp/lane5-name-module-batch-all-carriers.log 2>&1
git diff --check
```

ok  	github.com/system-inc/adamic/internal/oracle	238.918s
Verified 530 fixtures retain complete original declarations and reads.
Verified 77 fixtures retain requested original aliases, discriminators and enum values.

Oracle vet and formatting pass without output. All 138 counts rows are appended with the previous table unchanged. Each required whole-table updater fails on outside-lane cases recorded in its complete log; all four lane-specific updaters pass. No whole-package tests or full gate. Original source evidence now covers 99 members and 1370 candidate reads. Static inventory remains 4/308 pairs and 34/1503 reads certified, 304/1469 remaining. Setup reused: done 417.425s, nproc 5, quota 4 CPUs; complete timings remain in earlier reports. Only codex/views-callables pushed, no PR opened.

Group full SHAs:
19b8c74ab84edcfadc97d56adab8f96d68ceaa18
4c2cc83ea713275fb4a52d3923accde1bb0fc1e9
f241a55934c5eac1fd9b8791c566bcac70740fc0
77723f55ef2c08850809a28c7ac1a1606ce49371
