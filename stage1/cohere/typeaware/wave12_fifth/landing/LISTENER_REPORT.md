Built: numeric listener declarations in all twelve completed owned .a rule modules; shared numeric node callbacks remain blocked.
Commits: maintenance on codex/typeaware-wave-12, based on fetched main e8ba3d5d; no additional claims.
Checks: external pinned TypeScript SyntaxKind metadata comparison and full owned native oracle/sanitizer/mutant reruns; final results below.
Mutants: twelve numeric +1 metadata mutations are rejected; rule, raw-question and released-handle results remain distinct from metadata checks.
Uncovered: no kind-indexed dispatch, supplied-node callback migration or speed improvement; three React source ports remain blocked.

The declarations use TypeScript's numeric SyntaxKind from the pinned corpus src/compiler/types.ts, not Go ast.Kind. The enums differ: TypeScript includes ShebangTrivia, so Identifier is 80, CallExpression 214, NewExpression 215 and SourceFile 307. Initial Go-numbered declarations were corrected before commit. check_listeners.py independently reads the external TypeScript enum, validates all twelve exact listener lists, and rejects a numeric +1 mutation for every list.

| Rule | Numeric listener kinds |
| --- | --- |
| no-redeclare | 264, 265, 266, 267, 268, 263, 261, 209 |
| no-test-on-global-regex | 214 |
| no-write-only-collection | 261 |
| no-process-exit-after-output | 214 |
| no-uncleared-race-timeout | 214 |
| require-blocking-standard-streams | 307 |
| no-new-func | 214, 215 |
| no-new-native-nonconstructor | 215 |
| no-new-wrappers | 215 |
| prefer-regex-literals | 80 |
| prefer-rest-params | 80 |
| exhaustive-deps | 214 |

Regex literals listens to Identifier because its current production-equivalent analysis starts from global RegExp/global-object reference seeds and follows aliases to constructions. Blocking streams needs the SourceFile root for its whole-program entry/call ordering. Declaration grouping and alias analysis will require driver lifecycle/index integration; these exports are metadata, not an assertion that the current whole-file run methods are event callbacks.

Exact speed-rule blocker: current shared stage1/typescript/parser/nodes.ts defines only readonly kind: string. It has no numeric kind field. Bindings/Rules/Syntax and the existing driver accept table indexes and fetch ParseNode objects internally; no supplied-node numeric callback contract is available. Current implementations therefore still read string kinds and refetch nodes. This change does not claim compliance with those two execution requirements or a performance gain. The instruction to keep changes in owned rule files prevents replacing that shared API here. No parser, shared registration generator, shared harness or protected compiler file was edited. Once numeric ParseNode and the shared supplied-node driver arrive, migrate these modules and their owned helpers, preserving complete byte comparisons.

The other blocker remains separate: set-state-in-effect, set-state-in-render and static-components need native source-to-HIR lowering/SSA/capture/memo preparation; static-components additionally needs JSX parsing. Wave 21 has prepared-HIR validator cores but still explicitly lacks native source integration. No further claims were made.

Test output is redirected directly to /tmp/wave-12-listeners-{first,other,metadata}.log. The four native suites use the same frozen 77 compiler and 287 repository roots and their independent pinned production Go oracles. Artifacts are in /workspace/wave-12/listeners/{first,next,third,fourth}. Exact rerun results and compressed logs are appended after completion. The previous whole-process native versus Go timings remain in REPORT.md; declarations alone do not justify new performance claims. Full repository gate, inherited 26-rule regression and emitted-JavaScript lint execution remain outside this gate.

## Constructor refusal and owned workaround

The first fourth-batch attempt passed normal and sanitized controls, then its regex mutant build exited 1 with an escaping-constructor refusal inside Syntax.unwrap (a method). Repeating the preserved mutant build reproduced the refusal. This is a build failure, not a qualifying mutant catch. An initial manual probe used the wrong binary name; the corrected probe used the preserved adamic binary. A later probe omitted the toolchain environment and reached a missing-clang failure; the sourced rerun succeeded with empty build output. The reproduced refusal and successful sourced-build logs are preserved separately; the two manual setup mistakes are recorded here and were not retained as separate raw logs.

Moving only the three fourth-batch listener declarations from file preambles to file ends preserves existing declaration/method positions. The same preserved regex mutant then compiled successfully. No method implementation, compiler guard, shared file or test harness was changed. The complete fourth-batch gate is rerun after that workaround; only its successful rule mutants count below. The observation is that declaration placement changes acceptance. A source-position interaction with constructor checking is an inference, not a diagnosed or fixed compiler bug.

The first batch was rerun after changing initial Go-numbered metadata to TypeScript-numbered metadata. Its definitive artifact directory is final-first; the final fourth batch uses final-fourth. Other two batches use next and third. Initial logs are retained as historical attempts rather than definitive results.

## Definitive gate results

PASS: final first batch 115.822s, next 169.370s, third 114.941s, final fourth 135.535s. All twelve rules ran against their independent production Go oracle. Control findings/fixes/suggestions match in normal and ASan/UBSan/LSan builds with empty native stderr: first 187 controls/89 findings/65,212 bytes, next 142/138/86,901, third 94/56/25,933, fourth 703/596/327,920. Total 1,126 controls and 879 findings. Each batch also matched the same frozen 77 compiler and 287 repository roots normally and under sanitizers. First compiler/repository: 5/1 findings and 8,014/18,903 bytes. Other batches: zero findings and 5,010/18,485 bytes.

Commands used source /workspace/adamic-tools/env.sh; ADAMIC_TYPESCRIPT_SOURCE points to /workspace/wave-12/corpus, and the corresponding ADAMIC_WAVE12[_NEXT|_THIRD|_FOURTH]_COMPILER_MANIFEST and _REPOSITORY_MANIFEST point to the existing frozen manifests. Root: go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave12AgreementAndMutants$'. Next/third/fourth: go test -p 1 -v -count=1 -timeout 30m with those three owned package paths; the final fourth repeat ran only its own package. Logs are compressed in evidence/. Latest setup remains the previous continuation's 128s (Go/clang/Node/submodules 0s, cache 128s); no new setup was needed in this unchanged environment. nproc was checked again: 5. No compiler, Go bridge implementation, harness or registration file changed, so the previous direct checker/Node/vet results remain separate rather than claimed as rerun here.

All twelve rule mutants exit 0 with empty stderr and are caught solely by byte comparison. First differing bytes: redeclare 60, global regex 5,592, write-only collection 9,378; process 3,945, timeout 4,610, blocking streams 719; Function 1,412, nonconstructor 62, wrapper 578; regex literals 528, rest params 5,652, exhaustive deps 7,044. Five raw-question mutants likewise exit 0/empty stderr and fail bytes: ancestry 69, call body 3,945, import targets 13,294, constructor declaration-file 62, identifier declaration-file 69. Four registry-retention mutants remove the expected released-handle panic 70 and are caught by the lifetime assertions. The JSX-presence reversal is caught by valid-control success, not the byte-only rule-mutant criterion. Twelve numeric +1 metadata mutations are caught by the independent external-enum audit, not native findings. First-batch raw-question mutants and the Node one-byte mutant were not repeated in this maintenance gate.

## Quiet whole-process timing

Three alternating rounds per suite/corpus, after all builds and tests finished. The unchanged benchmark_wave_12.py driver compares complete findings/fixes/suggestions bytes in every round. Native and Go timings include loading, analysis, serialization and process overhead. Native remains slower; ignored metadata exports provide no measured driver optimization.

| Batch / corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| first / compiler | 4.160s | 0.601s | 6.93x |
| first / repository | 0.505s | 0.120s | 4.22x |
| next / compiler | 2.597s | 0.372s | 6.98x |
| next / repository | 0.396s | 0.112s | 3.53x |
| third / compiler | 2.640s | 0.343s | 7.69x |
| third / repository | 0.403s | 0.103s | 3.91x |
| fourth / compiler | 3.243s | 0.549s | 5.91x |
| fourth / repository | 0.455s | 0.110s | 4.15x |

Per-round measurements are preserved in evidence/listeners-*-measurements.json. Benchmark output is compressed separately. Full byte streams remain available under /workspace/wave-12/listeners/bench. Main remains the freshly fetched e8ba3d5d base. This maintenance is pushed only to codex/typeaware-wave-12; no new claims or integration pushes.
