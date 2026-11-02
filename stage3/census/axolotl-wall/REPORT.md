# Axolotl checker wall on the composed wave-7 tree

Measured 2026-11-02 UTC. **9 whole-program diagnostics remain; 10 distinct diagnostics across all meter records; 0 diagnostics marked new.**

| Measure | Result |
| --- | ---: |
| Sorted compiler roots | 79 |
| Files past checker, own file | 76/79 |
| Files past checker, whole program including imports | 2/79 |
| Ordinary successful lowering roots | 2/79 |
| Aggregate whole-program diagnostics | 9 |
| Global diagnostics in meter union | 0 |
| External diagnostics in meter union | 0 |
| Latent distinct NotYet sites | 644 |
| Latent distinct Refused sites | 4355 |
| Latent skipped dependency sites | 1 |
| Latent errors / panics | 0 / 0 |

## Pinned composition

| Input | Commit |
| --- | --- |
| Wave 7 | `24b1bab92a0214c344a293fefff2cdfc62eaeae4` |
| origin/main and report branch base | `80c2d3282e44b527555d59690bb413cc883d2bf2` |
| Scratch-only merge, compiler and adapters measured | `b52d5eb2320388b246a8bdc8b0124961ba413b43` |
| Classification | `ad5d21513f76663bef354d66fc0d995b4610c15f` |
| 99a command reference | `74dda102dc060de34dd7c8e7b4fd63f3539ef270` |
| Upstream TypeScript v6.0.3 | `050880ce59e30b356b686bd3144efe24f875ebc8` |

Ran `git merge origin/main` in a detached worktree at the exact wave-7 commit. Only three generated lane files conflicted: `stage3/lane/shard-plan.json`, `stage3/lane/shards.json`, and `stage3/lane/shards_generated_test.go`. They retain wave-7 versions. They are not inputs to apply, the checker, or either census. Main's sanctions generator and compiler changes merged. No adaptation, measurement tool, checker option, or prelude was edited. The scratch merge is not pushed.

All **59 adapters applied successfully** through unmodified `stage3/apply.sh`, using the meter's pinned upstream checkout and `npm ci --ignore-scripts --no-audit --no-fund`. This was a keyed product cache miss and local full build. Completion is checked against every adapter's before/after Git refs and the final patch table. Source hashes were frozen before measurement and rechecked after it.

Setup ran exactly as requested, with `GOPROXY=https://proxy.golang.org|direct` and the 1800-second timeout. The scratch worktree uses the recorded cohere and TypeScript submodule pins, retained in `submodule-pins.txt.gz`. Local-cache initialization failed on missing promisor objects; the meter's retry from recorded origins succeeded.

## Remaining diagnostics

The CLI passes the same 79 sorted `.ts`/`.a` roots described by the 99a README to `adamic types`. Its complete diagnostic multiset equals the ordinary census aggregate, including messages and elaboration chains. The table below is the distinct union of ordinary census diagnostics, preserving variants seen only when a file is its own root. `Whole` identifies diagnostics also present in the 79-root CLI run. Coordinates refer to the final adapted tree.

Row matching uses file, checker code, and source line text, ignoring surrounding whitespace. It does not use old line numbers. Duplicate keys are resolved by exact full-message equality, then by the relative occurrence order of identical source lines only when the original and final occurrences form a complete bijection. This distinguishes D034/D035 message variants and the two unchanged catch sites D041/D042. No match relies on old coordinates alone; unmatched diagnostics are marked **new**. Full classification cause and rationale are retained in `diagnostics-matched.json.gz`.

| Location | Code | Row id | Whole | Full message |
| --- | --- | --- | --- | --- |
| src/compiler/commandLineParser.ts:2301:91 | TS18046 | D032 | yes | 'e' is of type 'unknown'. |
| src/compiler/commandLineParser.ts:2677:9 | TS2375 | D033 | yes | Type '{ showConfig: undefined; configFile: undefined; configFilePath: undefined; help: undefined; init: undefined; listFiles: undefined; listEmittedFiles: undefined; project: undefined; build: undefined; version: undefined; }' is not assignable to type 'CompilerOptions' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.<br>  Types of property 'build' are incompatible.<br>    Type 'undefined' is not assignable to type 'boolean'. |
| src/compiler/commandLineParser.ts:2691:9 | TS2322 | D034 | no | Type '{ path: string; originalPath: undefined; prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; }[] &#124; undefined' is not assignable to type 'readonly ProjectReference[] &#124; undefined'.<br>  Type '{ path: string; originalPath: undefined; prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.<br>    Type '{ path: string; originalPath: undefined; prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.<br>      Types of property 'originalPath' are incompatible.<br>        Type 'undefined' is not assignable to type 'string'. |
| src/compiler/commandLineParser.ts:2691:9 | TS2322 | D035 | yes | Type '{ prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; path: string; originalPath: undefined; }[] &#124; undefined' is not assignable to type 'readonly ProjectReference[] &#124; undefined'.<br>  Type '{ prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; path: string; originalPath: undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.<br>    Type '{ prepend?: boolean &#124; undefined; circular?: boolean &#124; undefined; path: string; originalPath: undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.<br>      Types of property 'originalPath' are incompatible.<br>        Type 'undefined' is not assignable to type 'string'. |
| src/compiler/program.ts:406:25 | TS18046 | D041 | yes | 'e' is of type 'unknown'. |
| src/compiler/program.ts:441:25 | TS18046 | D042 | yes | 'e' is of type 'unknown'. |
| src/compiler/sys.ts:1281:43 | TS18046 | D054 | yes | 'e' is of type 'unknown'. |
| src/compiler/sys.ts:1598:69 | TS2307 | D055 | yes | Cannot find module 'source-map-support' or its corresponding type declarations. |
| src/compiler/sys.ts:1617:13 | TS2322 | D056 | yes | Type '(baseDir: string, moduleName: string) =&gt; { module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type '(baseDir: string, moduleName: string) =&gt; ModuleImportResult'.<br>  Type '{ module: any; modulePath: string; error: undefined; } &#124; { module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type 'ModuleImportResult'.<br>    Type '{ module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type 'ModuleImportResult'.<br>      Type '{ module: undefined; modulePath: undefined; error: unknown; }' is not assignable to type '{ module: undefined; modulePath?: undefined; error: { stack?: string; message?: string; }; }'.<br>        Types of property 'error' are incompatible.<br>          Type 'unknown' is not assignable to type '{ stack?: string; message?: string; }'. |
| src/compiler/sys.ts:1708:56 | TS2345 | D057 | yes | Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string &#124; ArrayBufferView&lt;ArrayBufferLike&gt;'.<br>  Type 'undefined' is not assignable to type 'string &#124; ArrayBufferView&lt;ArrayBufferLike&gt;'. |

### Matched source text

| Location | Row id | Final source line |
| --- | --- | --- |
| src/compiler/commandLineParser.ts:2301:91 | D032 | <code>        return createCompilerDiagnostic(Diagnostics.Cannot_read_file_0_Colon_1, fileName, e.message);</code> |
| src/compiler/commandLineParser.ts:2677:9 | D033 | <code>        compilerOptions: {</code> |
| src/compiler/commandLineParser.ts:2691:9 | D034 | <code>        references: map(configParseResult.projectReferences, r =&gt; ({ ...r, path: r.originalPath ? r.originalPath : "", originalPath: undefined })),</code> |
| src/compiler/commandLineParser.ts:2691:9 | D035 | <code>        references: map(configParseResult.projectReferences, r =&gt; ({ ...r, path: r.originalPath ? r.originalPath : "", originalPath: undefined })),</code> |
| src/compiler/program.ts:406:25 | D041 | <code>                onError(e.message);</code> |
| src/compiler/program.ts:441:25 | D042 | <code>                onError(e.message);</code> |
| src/compiler/sys.ts:1281:43 | D054 | <code>                hitSystemWatcherLimit &#124;&#124;= e.code === "ENOSPC";</code> |
| src/compiler/sys.ts:1598:69 | D055 | <code>                    (require("source-map-support") as typeof import("source-map-support")).install();</code> |
| src/compiler/sys.ts:1617:13 | D056 | <code>            require: (baseDir, moduleName) =&gt; {</code> |
| src/compiler/sys.ts:1708:56 | D057 | <code>                        _fs.writeFileSync(profilePath, JSON.stringify(cleanupPaths(profile)));</code> |

## Latent lowering census

Completed in 1587.279 seconds, below the 90-minute limit. `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`; mode: `first-error`. All 79 compiler source files are covered. The no-output assertions succeeded.

This is **measured on a checker-rejected program**. Counts are unique `(kind, where, reason, text)` sites across attempts, using the meter's own `latent_summary` and ownership resolver. 5985 recorded finding events; 0 recovery boundary sites. This measurement does not establish successful lowering or native output.

### Top 10 reasons by distinct site count

| Reason | Owner | NotYet | Refused | Sites |
| --- | --- | ---: | ---: | ---: |
| an object refinement using an open numeric enum as a literal tag | #drwm5pj (Lower) | 0 | 1251 | 1251 |
| a cast the runtime can't check | 01a11410 | 0 | 1013 | 1013 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 86 | 86 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where SourceFile is read | adaptation 70, stage 3 | 0 | 69 | 69 |
| optional property id in Node absent from structural source never, which can hide fields | OWNER BLANK | 0 | 56 | 56 |
| checked predicate membership over an open tag domain | #k1t2wp9 (Optimizer) | 47 | 0 | 47 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | OWNER BLANK | 0 | 39 | 39 |
| a non-object representation passed through an opaque object parameter | OWNER BLANK | 32 | 0 | 32 |
| a call through ?. (an optional call) | OWNER BLANK | 28 | 0 | 28 |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | OWNER BLANK | 0 | 26 | 26 |

Owners come from the measured tree's `stage3/meter/owners.json`, using its longest-prefix rule and variance-family override. `OWNER BLANK` means no owner mapping exists.

## Ordinary meter file counts

Computed by unmodified `stage3/meter/report.py` over ordinary census JSONL. Whole-program file counts include imported diagnostics; own-file counts attribute each diagnostic to its primary location. The aggregate is a separate 79-root observation. Non-source files fail the extension gate and are excluded from the source denominator.

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
| src/compiler/diagnosticMessages.generated.json | fail | fail | blocked | 0 |
| src/compiler/diagnosticMessages.json | fail | fail | blocked | 0 |
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
| src/compiler/tsconfig.json | fail | fail | blocked | 0 |
| src/compiler/types.ts | pass | fail | blocked | 0 |
| src/compiler/utilities.ts | pass | fail | blocked | 0 |
| src/compiler/utilitiesPublic.ts | pass | fail | blocked | 0 |
| src/compiler/visitorPublic.ts | pass | fail | blocked | 0 |
| src/compiler/watch.ts | pass | fail | blocked | 0 |
| src/compiler/watchPublic.ts | pass | fail | blocked | 0 |
| src/compiler/watchUtilities.ts | pass | fail | blocked | 0 |

## Adapter application census

# Patch set

Compared with the pinned pristine tree after upstream diagnostic generation.
Rows measure incremental edits; total measures the final tree against pristine.

| Adaptation | Files | Lines added | Lines removed |
|---|---:|---:|---:|
| 00-setup | 0 | 0 | 0 |
| 10-type-imports | 73 | 3720 | 3720 |
| 20-optional-declarations | 25 | 399 | 399 |
| 30-indexed-reads | 11 | 209 | 209 |
| 31-indexed-reads-checker | 1 | 394 | 394 |
| 32-indexed-reads-program | 20 | 132 | 132 |
| 33-indexed-reads-emit | 16 | 119 | 119 |
| 40-explicit-any | 18 | 107 | 107 |
| 41-explicit-any-remaining | 8 | 25 | 22 |
| 42-scanner-any | 1 | 1 | 1 |
| 43-any-returns | 2 | 14 | 9 |
| 44-config-generics | 1 | 16 | 16 |
| 45-regex-captures | 5 | 12 | 10 |
| 46-fix-pragma-empty-argument | 1 | 1 | 1 |
| 47-host-errors | 3 | 16 | 5 |
| 48-memoize | 1 | 2 | 2 |
| 49-honest-optionals | 4 | 147 | 144 |
| 50-temporary-scanner-implicit-returns | 1 | 2 | 0 |
| 51-temporary-scanner-fallthrough | 0 | 0 | 0 |
| 52-json-values | 13 | 268 | 65 |
| 60-temporary-factory-local-symbol | 1 | 1 | 1 |
| 61-temporary-factory-type-expression | 1 | 1 | 1 |
| 62-temporary-tracing-write | 1 | 1 | 1 |
| 63-temporary-tracing-legend | 1 | 1 | 1 |
| 64-temporary-parser-range-read | 1 | 1 | 1 |
| 65-temporary-node-builtins | 2 | 7 | 2 |
| 70-readonly-views | 7 | 22 | 22 |
| 71-writable-views | 20 | 88 | 88 |
| 75-optional-widening | 1 | 16 | 5 |
| 76-truthful-casts | 1 | 1 | 1 |
| 78-variance-marker | 1 | 1 | 1 |
| 79-exit-code | 8 | 9 | 9 |
| 90-void-widening | 7 | 66 | 66 |
| 91-patcher-discard | 2 | 2 | 2 |
| 92-tty-host-results | 1 | 2 | 2 |
| 95-checker-axolotl-a | 1 | 13 | 13 |
| 95-checker-axolotl-b | 1 | 7 | 7 |
| 95-checker-axolotl-c | 2 | 3 | 3 |
| 95-checker-axolotl-d | 1 | 1 | 1 |
| 95-checker-axolotl-e | 1 | 2 | 2 |
| 95-checker-axolotl-f | 1 | 1 | 0 |
| 95-checker-axolotl-g | 1 | 11 | 1 |
| 95-checker-axolotl-h | 1 | 4 | 4 |
| 96-emit-axolotl-a | 1 | 14 | 14 |
| 96-emit-axolotl-b | 1 | 2 | 2 |
| 96-emit-axolotl-c | 1 | 19 | 19 |
| 97-cli-axolotl-a | 1 | 1 | 1 |
| 97-cli-axolotl-b | 3 | 12 | 12 |
| 97-cli-axolotl-c | 1 | 2 | 2 |
| 97-cli-axolotl-d | 4 | 24 | 24 |
| 98-axolotl-a | 2 | 8 | 8 |
| 98-axolotl-b | 1 | 5 | 5 |
| 98-axolotl-d | 5 | 13 | 10 |
| 99-correctness-axolotl-a | 4 | 12 | 12 |
| 99-correctness-axolotl-b | 1 | 1 | 1 |
| 99-correctness-axolotl-e | 1 | 4 | 2 |
| 99-correctness-axolotl-f | 5 | 27 | 8 |
| 99-correctness-axolotl-g | 1 | 4 | 3 |
| 99-correctness-axolotl-h | 1 | 5 | 5 |
| **Total** | 93 | 5984 | 5703 |

## Raw evidence

Every raw measurement output, setup/apply/build log, matching result, and provenance record is gzipped beside this report. `commands.json.gz` records exact measurement arguments, exit statuses, UTC starts, and elapsed seconds. `measurement-driver.py.gz`, `analyze.py.gz`, and `freeze.py.gz` retain the scratch measurement and analysis scripts. `report.json.gz` and `report.md.gz` are the meter's unmodified ordinary reports; `latent-summary.json.gz` is its lowering summary. `classification.json.gz` is the exact classification input. The binaries are identified by SHA-256, not committed.
