# Axolotl checker wall after waves 1 through 7

Measured 2026-11-02 UTC. Main is pinned at `246167051e0d35e7910e32b7b8fa443f0543545a`. The report branch has that commit as its sole parent.

## Headline

| Measure | Main | Main plus six accepted heads |
| --- | ---: | ---: |
| Sorted compiler roots | 79 | 79 |
| Files past checker, own file | 76/79 | 79/79 |
| Files past checker, whole program including imports | 2/79 | 79/79 |
| Ordinary successful lowering roots | 2/79 | not measured: census panic |
| Whole-program diagnostics | 9 | 0 |
| Distinct diagnostics across own-root observations | 10 | 0 |
| New diagnostics across own-root observations | 0 | 0 |

## Measurement and matching

The CLI passes all 79 sorted `.ts` and `.a` roots recursively under `src/compiler` to `adamic types`, with unmodified stage-0 checker options and prelude. On main the ordinary census loads each file as its own root and the same 79-root aggregate; its full diagnostic multiset agrees with the CLI, including elaboration chains. On the composed tree, the ordinary census panics before producing its first record. The unchanged CLI therefore checks each of the 79 singleton roots to complete the checker-only census; each stdout is retained compressed. The successful 79-root CLI remains the aggregate observation. Own-file attribution follows the meter: primary diagnostic location, deduplicating full strings across roots. JSON inputs fail the extension gate and are excluded from the denominator.

Row IDs come from `typescript/axolotl-classify` at `ad5d21513f76663bef354d66fc0d995b4610c15f`, matched by file, code and source-line text with surrounding whitespace ignored. Duplicate source keys use exact full-message equality, then occurrence order only when original and final sites form a complete bijection. Old coordinates are not match keys. Unmatched rows are marked new.

**Owner-data limitation:** the exact pinned classification JSON has no `owner` field, and its Markdown diagnostic table has no Owner column. Therefore every classification owner is recorded as unavailable. The separate Head owner column uses the ownership supplied in this task, not an invented classification value. Head claims are read from each accepted README. 94 is a dependency helper and claims no diagnostic row directly.

## Main tip remaining diagnostics

9 whole-program diagnostics.

| Location | Code | Row id | Classification owner | Head owner | Claiming head | Full message |
| --- | --- | --- | --- | --- | --- | --- |
| src/compiler/commandLineParser.ts:2301:91 | TS18046 | D032 | unavailable | Correctness | 99d | 'e' is of type 'unknown'. |
| src/compiler/commandLineParser.ts:2677:9 | TS2375 | D033 | unavailable | Correctness | 99c | Type '{ showConfig: undefined; configFile: undefined; configFilePath: undefined; help: undefined; init: undefined; listFiles: undefined; listEmittedFiles: undefined; project: undefined; build: undefined; version: undefined; }' is not assignable to type 'CompilerOptions' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.<br>  Types of property 'build' are incompatible.<br>    Type 'undefined' is not assignable to type 'boolean'. |
| src/compiler/commandLineParser.ts:2691:9 | TS2322 | D035 | unavailable | Correctness | 99c | Type '{ prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; path: string; originalPath: undefined; }[] &#124; undefined' is not assignable to type 'readonly ProjectReference[] &#124; undefined'.<br>  Type '{ prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; path: string; originalPath: undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.<br>    Type '{ prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; path: string; originalPath: undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.<br>      Types of property 'originalPath' are incompatible.<br>        Type 'undefined' is not assignable to type 'string'. |
| src/compiler/program.ts:406:25 | TS18046 | D041 | unavailable | CLI | 97e | 'e' is of type 'unknown'. |
| src/compiler/program.ts:441:25 | TS18046 | D042 | unavailable | CLI | 97e | 'e' is of type 'unknown'. |
| src/compiler/sys.ts:1281:43 | TS18046 | D054 | unavailable | Performance | 98-c | 'e' is of type 'unknown'. |
| src/compiler/sys.ts:1598:69 | TS2307 | D055 | unavailable | Performance | 98-c | Cannot find module 'source-map-support' or its corresponding type declarations. |
| src/compiler/sys.ts:1617:13 | TS2322 | D056 | unavailable | Performance | 98-e | Type '(baseDir: string, moduleName: string) =&gt; { module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type '(baseDir: string, moduleName: string) =&gt; ModuleImportResult'.<br>  Type '{ module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type 'ModuleImportResult'.<br>    Type '{ module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type 'ModuleImportResult'.<br>      Type '{ module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type '{ module: undefined; modulePath?: undefined; error: { stack?: string; message?: string; }; }'.<br>        Types of property 'error' are incompatible.<br>          Type 'unknown' is not assignable to type '{ stack?: string; message?: string; }'. |
| src/compiler/sys.ts:1708:56 | TS2345 | D057 | unavailable | Performance | 98-c | Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string &#124; ArrayBufferView&lt;ArrayBufferLike&gt;'.<br>  Type 'undefined' is not assignable to type 'string &#124; ArrayBufferView&lt;ArrayBufferLike&gt;'. |

Additional own-file union variants, absent from the aggregate:

| Location | Code | Row id | Head owner | Claiming head | Full message |
| --- | --- | --- | --- | --- | --- |
| src/compiler/commandLineParser.ts:2691:9 | TS2322 | D034 | Correctness | 99c | Type '{ path: string; originalPath: undefined; prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; }[] &#124; undefined' is not assignable to type 'readonly ProjectReference[] &#124; undefined'.<br>  Type '{ path: string; originalPath: undefined; prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.<br>    Type '{ path: string; originalPath: undefined; prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.<br>      Types of property 'originalPath' are incompatible.<br>        Type 'undefined' is not assignable to type 'string'. |

Matched source lines and resolution evidence are retained in `main-diagnostics-matched.json.gz`.

## Composed remaining diagnostics

0 whole-program diagnostics.

| Location | Code | Row id | Classification owner | Head owner | Claiming head | Full message |
| --- | --- | --- | --- | --- | --- | --- |
| none | | | | | | |

Matched source lines and resolution evidence are retained in `composed-diagnostics-matched.json.gz`.

## Main latent lowering census

Completed in 1106.637 seconds. `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, first-error mode, with a 5400-second timeout and 60-second kill grace. All 79 compiler roots are covered. The no-output assertions succeeded.

Distinct NotYet: **553**. Distinct Refused: **4339**. Skipped dependency sites: 1. Errors: 0. Panics: 0. Recovery boundaries: 0. Recorded finding events: 5880.

Measured on a checker-rejected program. Counts are unique `(kind, where, reason, text)` sites across attempts, using the pinned main meter's `latent_summary`. They do not establish successful lowering or native output.

| Reason | Owner | NotYet | Refused | Sites |
| --- | --- | ---: | ---: | ---: |
| an object refinement using an open numeric enum as a literal tag | #drwm5pj (Lower) | 0 | 1251 | 1251 |
| a cast the runtime can't check | 01a11410 | 0 | 1013 | 1013 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 86 | 86 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 69 | 69 |
| optional property id in Node absent from structural source never, which can hide fields | OWNER BLANK | 0 | 56 | 56 |
| checked predicate membership over an open tag domain | #k1t2wp9 (Optimizer) | 47 | 0 | 47 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 36 | 36 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 32 | 0 | 32 |
| a call through ?. (an optional call) | OWNER BLANK | 28 | 0 | 28 |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | OWNER BLANK | 0 | 26 | 26 |

Owners use the measured main `stage3/meter/owners.json`, its longest-prefix rule and variance-family override. OWNER BLANK means that mapping has no owner.

## Exact composition and adapter application

| Head | Commit | Directory |
| --- | --- | --- |
| 94 | `8a7cd267b434efa0a60d481529a1256cc3c92a19` | `stage3/adapt/94-axolotl-property-read` |
| 98-c | `3fb2903296db936eef4147d19d5d252fa224bb2e` | `stage3/adapt/98-axolotl-c` |
| 98-e | `fcce3b29a4bc2fc1b73f73307a1a3a56c3a0a977` | `stage3/adapt/98-axolotl-e` |
| 97e | `555addc2d789788866de4c5550ccc5c85be6ec67` | `stage3/adapt/97-cli-axolotl-e` |
| 99d | `cc2bfae014073a16ba9c54712f67b372e2c76429` | `stage3/adapt/99-correctness-axolotl-d` |
| 99c | `820d57df8395a56359c8eec0c3e109b8d7d27226` | `stage3/adapt/99-correctness-axolotl-c` |

**Composed ordinary-census limitation:** its first root, `src/compiler/_namespaces/ts.moduleSpecifiers.ts`, triggers `panic: interface conversion: ast.nodeData is *ast.ClassExpression, not *ast.ClassDeclaration` in `internal/lower/namespaces.go:271`. Exit status is 2 and no ordinary census rows are emitted. No compiler source is patched. Composed lowering counts are unavailable; the complete checker-only CLI census supplies its diagnostics and checker file counts. The crash and command are retained raw.

Scratch worktrees were created from the pinned origin/main. Each directory was brought with `git checkout <sha> -- stage3/adapt/<directory>`, with no edits to the heads or their directories. Exact Git tree matches are verified in `composition-verification.json.gz`. 94 is included once. The 97e LEDGER.md entry is appended after main's last entry and numbered 7; no ledger or adaptation changes are included in this report commit. The scratch composition is never pushed.

Both adapted trees are built by unmodified `stage3/apply.sh`, with the meter's pinned TypeScript v6.0.3 input, `050880ce59e30b356b686bd3144efe24f875ebc8`, and lockfile installation via `npm ci --ignore-scripts --no-audit --no-fund`. All adapters must have complete before/after refs and patch-table rows; `adapter-coverage.json.gz` retains the check. No checker, adapter, prelude, or tool source is modified. Both trees use the compiler built from pinned main and its recorded submodule pins.

Setup ran with `GOPROXY=https://proxy.golang.org|direct` and `timeout --kill-after=60s 1800 bash cloud/setup.sh`, exit 0 in 576.935 seconds. Full logs are retained. The cloud runtime skill was used for environment and network readiness.

## Per-file census

### main

| File | Own file | Whole program | Lowering | Own diagnostics |
| --- | --- | --- | --- | ---: |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | pass | fail | blocked | 0 |
| src/compiler/_namespaces/ts.performance.ts | pass | fail | blocked | 0 |
| src/compiler/_namespaces/ts.ts | pass | fail | blocked | 0 |
| src/compiler/binder.ts | pass | fail | blocked | 0 |
| src/compiler/builder.ts | pass | fail | blocked | 0 |
| src/compiler/builderPublic.ts | pass | fail | blocked | 0 |
| src/compiler/builderState.ts | pass | fail | blocked | 0 |
| src/compiler/builderStatePublic.ts | pass | fail | blocked | 0 |
| src/compiler/checker.ts | pass | fail | blocked | 0 |
| src/compiler/commandLineParser.ts | fail | fail | blocked | 4 |
| src/compiler/core.ts | pass | fail | blocked | 0 |
| src/compiler/corePublic.ts | pass | pass | pass | 0 |
| src/compiler/debug.ts | pass | fail | blocked | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | pass | fail | blocked | 0 |
| src/compiler/emitter.ts | pass | fail | blocked | 0 |
| src/compiler/executeCommandLine.ts | pass | fail | blocked | 0 |
| src/compiler/expressionToTypeNode.ts | pass | fail | blocked | 0 |
| src/compiler/factory/baseNodeFactory.ts | pass | fail | blocked | 0 |
| src/compiler/factory/emitHelpers.ts | pass | fail | blocked | 0 |
| src/compiler/factory/emitNode.ts | pass | fail | blocked | 0 |
| src/compiler/factory/nodeChildren.ts | pass | fail | blocked | 0 |
| src/compiler/factory/nodeConverters.ts | pass | fail | blocked | 0 |
| src/compiler/factory/nodeFactory.ts | pass | fail | blocked | 0 |
| src/compiler/factory/nodeTests.ts | pass | fail | blocked | 0 |
| src/compiler/factory/parenthesizerRules.ts | pass | fail | blocked | 0 |
| src/compiler/factory/utilities.ts | pass | fail | blocked | 0 |
| src/compiler/factory/utilitiesPublic.ts | pass | fail | blocked | 0 |
| src/compiler/hostErrors.ts | pass | pass | pass | 0 |
| src/compiler/moduleNameResolver.ts | pass | fail | blocked | 0 |
| src/compiler/moduleSpecifiers.ts | pass | fail | blocked | 0 |
| src/compiler/parser.ts | pass | fail | blocked | 0 |
| src/compiler/path.ts | pass | fail | blocked | 0 |
| src/compiler/performance.ts | pass | fail | blocked | 0 |
| src/compiler/performanceCore.ts | pass | fail | blocked | 0 |
| src/compiler/program.ts | fail | fail | blocked | 2 |
| src/compiler/programDiagnostics.ts | pass | fail | blocked | 0 |
| src/compiler/resolutionCache.ts | pass | fail | blocked | 0 |
| src/compiler/scanner.ts | pass | fail | blocked | 0 |
| src/compiler/semver.ts | pass | fail | blocked | 0 |
| src/compiler/sourcemap.ts | pass | fail | blocked | 0 |
| src/compiler/symbolWalker.ts | pass | fail | blocked | 0 |
| src/compiler/sys.ts | fail | fail | blocked | 4 |
| src/compiler/tracing.ts | pass | fail | blocked | 0 |
| src/compiler/transformer.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/classFields.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/classThis.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/declarations.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/declarations/diagnostics.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/destructuring.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/es2015.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/es2016.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/es2017.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/es2018.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/es2019.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/es2020.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/es2021.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/esDecorators.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/esnext.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/generators.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/jsx.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/legacyDecorators.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/module/module.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/module/system.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/namedEvaluation.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/taggedTemplate.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/ts.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/typeSerializer.ts | pass | fail | blocked | 0 |
| src/compiler/transformers/utilities.ts | pass | fail | blocked | 0 |
| src/compiler/tsbuild.ts | pass | fail | blocked | 0 |
| src/compiler/tsbuildPublic.ts | pass | fail | blocked | 0 |
| src/compiler/types.ts | pass | fail | blocked | 0 |
| src/compiler/utilities.ts | pass | fail | blocked | 0 |
| src/compiler/utilitiesPublic.ts | pass | fail | blocked | 0 |
| src/compiler/visitorPublic.ts | pass | fail | blocked | 0 |
| src/compiler/watch.ts | pass | fail | blocked | 0 |
| src/compiler/watchPublic.ts | pass | fail | blocked | 0 |
| src/compiler/watchUtilities.ts | pass | fail | blocked | 0 |
### composed

| File | Own file | Whole program | Lowering | Own diagnostics |
| --- | --- | --- | --- | ---: |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | pass | pass | not measured | 0 |
| src/compiler/_namespaces/ts.performance.ts | pass | pass | not measured | 0 |
| src/compiler/_namespaces/ts.ts | pass | pass | not measured | 0 |
| src/compiler/binder.ts | pass | pass | not measured | 0 |
| src/compiler/builder.ts | pass | pass | not measured | 0 |
| src/compiler/builderPublic.ts | pass | pass | not measured | 0 |
| src/compiler/builderState.ts | pass | pass | not measured | 0 |
| src/compiler/builderStatePublic.ts | pass | pass | not measured | 0 |
| src/compiler/checker.ts | pass | pass | not measured | 0 |
| src/compiler/commandLineParser.ts | pass | pass | not measured | 0 |
| src/compiler/core.ts | pass | pass | not measured | 0 |
| src/compiler/corePublic.ts | pass | pass | not measured | 0 |
| src/compiler/debug.ts | pass | pass | not measured | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | pass | pass | not measured | 0 |
| src/compiler/emitter.ts | pass | pass | not measured | 0 |
| src/compiler/executeCommandLine.ts | pass | pass | not measured | 0 |
| src/compiler/expressionToTypeNode.ts | pass | pass | not measured | 0 |
| src/compiler/factory/baseNodeFactory.ts | pass | pass | not measured | 0 |
| src/compiler/factory/emitHelpers.ts | pass | pass | not measured | 0 |
| src/compiler/factory/emitNode.ts | pass | pass | not measured | 0 |
| src/compiler/factory/nodeChildren.ts | pass | pass | not measured | 0 |
| src/compiler/factory/nodeConverters.ts | pass | pass | not measured | 0 |
| src/compiler/factory/nodeFactory.ts | pass | pass | not measured | 0 |
| src/compiler/factory/nodeTests.ts | pass | pass | not measured | 0 |
| src/compiler/factory/parenthesizerRules.ts | pass | pass | not measured | 0 |
| src/compiler/factory/utilities.ts | pass | pass | not measured | 0 |
| src/compiler/factory/utilitiesPublic.ts | pass | pass | not measured | 0 |
| src/compiler/hostErrors.ts | pass | pass | not measured | 0 |
| src/compiler/moduleNameResolver.ts | pass | pass | not measured | 0 |
| src/compiler/moduleSpecifiers.ts | pass | pass | not measured | 0 |
| src/compiler/parser.ts | pass | pass | not measured | 0 |
| src/compiler/path.ts | pass | pass | not measured | 0 |
| src/compiler/performance.ts | pass | pass | not measured | 0 |
| src/compiler/performanceCore.ts | pass | pass | not measured | 0 |
| src/compiler/program.ts | pass | pass | not measured | 0 |
| src/compiler/programDiagnostics.ts | pass | pass | not measured | 0 |
| src/compiler/resolutionCache.ts | pass | pass | not measured | 0 |
| src/compiler/scanner.ts | pass | pass | not measured | 0 |
| src/compiler/semver.ts | pass | pass | not measured | 0 |
| src/compiler/sourcemap.ts | pass | pass | not measured | 0 |
| src/compiler/symbolWalker.ts | pass | pass | not measured | 0 |
| src/compiler/sys.ts | pass | pass | not measured | 0 |
| src/compiler/tracing.ts | pass | pass | not measured | 0 |
| src/compiler/transformer.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/classFields.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/classThis.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/declarations.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/declarations/diagnostics.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/destructuring.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/es2015.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/es2016.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/es2017.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/es2018.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/es2019.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/es2020.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/es2021.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/esDecorators.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/esnext.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/generators.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/jsx.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/legacyDecorators.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/module/module.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/module/system.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/namedEvaluation.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/taggedTemplate.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/ts.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/typeSerializer.ts | pass | pass | not measured | 0 |
| src/compiler/transformers/utilities.ts | pass | pass | not measured | 0 |
| src/compiler/tsbuild.ts | pass | pass | not measured | 0 |
| src/compiler/tsbuildPublic.ts | pass | pass | not measured | 0 |
| src/compiler/types.ts | pass | pass | not measured | 0 |
| src/compiler/utilities.ts | pass | pass | not measured | 0 |
| src/compiler/utilitiesPublic.ts | pass | pass | not measured | 0 |
| src/compiler/visitorPublic.ts | pass | pass | not measured | 0 |
| src/compiler/watch.ts | pass | pass | not measured | 0 |
| src/compiler/watchPublic.ts | pass | pass | not measured | 0 |
| src/compiler/watchUtilities.ts | pass | pass | not measured | 0 |

## Raw evidence

Raw CLI, census and latent outputs are gzipped beside this report, together with command arguments, statuses, UTC starts, elapsed times, source hashes, matching input/output, adapter coverage, patch tables, submodule pins, binary hashes, build/setup/apply logs, scripts and provenance. Main CLI/aggregate equality, complete 79-root checker coverage on both trees, row matching, source immutability and exact adapter directory matches are checked. Complete latent coverage is verified when that census finishes; an incomplete run is labeled explicitly. `SHA256SUMS` covers every report artifact.
