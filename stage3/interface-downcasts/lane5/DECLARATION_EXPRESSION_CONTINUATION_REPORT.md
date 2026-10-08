Built: three groups certify 10 additional callable pairs, 102 candidate reads and 57 fixtures.
Commits: 428f10a1, 433e1ec0 and c9017a85, each pushed separately to codex/views-callables; this commit records final validation.
Checks: original declarations/read spans, aliases/tags, Node, release/sanitized native, JavaScript, leaks, measured lane counts and complete later-ranked harness pass; whole-table counts fails outside lane.
Mutants: 21 runtime runs catch 57 pair-level assertions; one token discriminator verifier mutant caught; all sources restored.
Uncovered: 95/2818 pairs and 2261/11063 candidate reads certified; 2723 pairs and 8802 reads remain; higher-order hooks, mapped receivers, intrinsic and other untested families remain excluded.

Reporting date October 12. Continued from 8a5b672e6a892ea96002c4481b410e28048ca9ef. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Read counts are conservative inventory counts; exact production execution remains unmeasured.

| Group | Commit | New pairs / reads | Fixtures | Runtime runs / assertions |
|---|---|---:|---:|---:|
| Resolution settings, scanner, lexical entry, writer spacing | 428f10a1 | 4 / 43 | 16 | 7 / 15 |
| Tagged parameter and property declaration updates | 433e1ec0 | 2 / 22 | 14 | 7 / 14 |
| Import declaration, function call, partial emission, strict inequality | c9017a85 | 4 / 37 | 27 | 7 / 28 |

Certified ranks: 159,160,163,164,166,170,173,185,187,188. Complete original declarations and member read expressions are retained. Adjacent structural carriers, enum representations and helper/producer bodies are reduced, as each group report specifies. Original BindingName/BindingPattern and PropertyName alternatives and numeric tags are independently checked; valid cases exercise every name alternative and string names. Original token kinds 26,58,54 are resolved with the original checker; changing ExclamationToken 54 to 55 is rejected and restored. Optional parameters and array elements retain lazy descendant-read checks.

Every runtime mutant and affected family is listed in the three group reports and raw logs. Native/JavaScript arity, result and parameter mutations are caught at reached member reads; omitted payload registration is caught at descendant string reads, including sanitizer evidence. No build failure counts as a mutant result. All compiler/runtime files used for temporary mutations are restored; no production compiler or calling convention changes are included.

The original stored onEmitNode declaration/read is retained in two separate uncertified probes. Its compatible reduced implementation prints 4 on source Node, native release and JavaScript through oracle/node.mjs, using the existing implementation proof. The wrong-arity producer prints 9 on source Node and is refused at the reached native member read with unsupported callable contract. These probes remain outside certification and measured counts; a runtime nested-callback contract needs proof. The initial direct JavaScript CLI execution lacked the adamic module mapping; rerunning with the oracle loader prints 4. The failed native build did not produce a wrong-arity binary. Neither command issue is counted as a callable certificate.

Set remains skipped as instructed and needs collection-to-view conversion plus recorded intrinsic callable signatures, including add's self return. Arrow rank 120 remains skipped and needs a proved original ConciseBody union-read contract. Broad ForInitializer, method-value reads and other earlier boundaries remain as recorded. Rank 162 needs a mapped Required<Pick<SymbolTracker, "reportInferenceFallback">> certificate preserving its original optional declaration; it is not substituted with a rewritten required method. Array/Map/SymbolTable intrinsic and external overloaded members, generic/predicate signatures, other original name/module unions, class-this, dynamic callable elements and host intersections remain uncertified. For untested families these are pending requirements, not observed refusals. No Union exception decision from the integrator was received.

Exact per-group commands and outputs are in RESOLUTION_SCAN_LEXICAL_SPACE_REPORT.md, TAGGED_PARAMETER_PROPERTY_UPDATES_REPORT.md and IMPORT_CALL_PARTIAL_INEQUALITY_REPORT.md. Final checks:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-callable-batch-all-restored.log 2>&1
go vet ./internal/oracle > /tmp/lane5-callable-batch-vet.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original "$(cat /tmp/lane5-next-expression-ranks.txt)" > /tmp/lane5-callable-batch-all-fixtures.log 2>&1
node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 122,141,144,145,159,160 > /tmp/lane5-callable-batch-all-carriers.log 2>&1
git diff --check
```

ok  	github.com/system-inc/adamic/internal/oracle	185.217s
Verified 392 fixtures retain complete original declarations and reads.
Verified 42 fixtures retain original aliases and numeric kind discriminators.

Oracle vet and formatting pass without output. All 57 counts rows are appended with the previous table unchanged. Required whole-table updater failures are preserved in each group log; all three lane updaters pass. No whole-package tests or full gate. Static inventory remains 4/308 pairs and 34/1503 reads certified, 304/1469 remaining. Setup reused: done 417.425s, nproc 5, quota 4 CPUs; complete timing lines remain in earlier reports. Only codex/views-callables pushed; no PR opened.

Group full SHAs:
428f10a1b8abaabc4772839ea7e2fd2346bebb3f
433e1ec09fdba7c83fe2a3d5153cf28962a23022
c9017a85b1a24d73bc204a368316f1f9ac8f8596
