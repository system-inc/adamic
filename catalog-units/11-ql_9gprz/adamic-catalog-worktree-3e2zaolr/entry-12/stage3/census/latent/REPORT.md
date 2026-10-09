Built: latent census of untouched main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 and area/stage3 e39a299323cad76aea44ec71ca7b740130323d4c, each using its own apply.sh.
Main: NotYet 2059, Refused 2140; measured on a checker-rejected program.
Area/stage3: NotYet 2075, Refused 2104; measured on a checker-rejected program.
Checks: both guarded builds/runs, independent unique-site/ranking/body-span recount, 12 dedicated evidence mutants, and output-guard mutants for both binaries; logs in data/unmerged4/.
Limits: first lowering failure per unit; checker-diagnosed bodies skipped; no usable IR, native run, adaptation oracle suite or full gate.

# Summary

All counts in this report and JSON are **measured on a checker-rejected program**. These are lowering observations after bypassing only the checker rejection in the measurement loader; they do not establish accepted programs or native correctness. Both top-10 tables combine NotYet and Refused exact reason families; lower kinds, dependency skips, ordinary errors and panics are excluded from the ranking.

## main-unmerged

NotYet **2059**; Refused **2140**. **measured on a checker-rejected program**.

| Kind | Exact reason | Count |
| --- | --- | ---: |
| Refused | a type predicate | 588 |
| NotYet | reading SyntaxKind | 484 |
| Refused | the non-null assertion ! | 484 |
| NotYet | a function without a body | 191 |
| Refused | enum | 162 |
| NotYet | an EnumDeclaration | 154 |
| NotYet | a PrefixUnaryExpression on a value | 85 |
| NotYet | a method call through a structural signature in a program with statics; use typeof the declaring class | 80 |
| Refused | an ExportDeclaration | 77 |
| NotYet | a function returning T \| undefined | 66 |

Checker diagnostics: 897; own-file zero diagnostics: 22/78. Functions attempted: 2637; own-body skips: 118. All lowering outcomes: `{'NotYet': 2059, 'Refused': 2140, 'SkippedDependency': 4, 'error': 0, 'panic': 0}`.

## area-unmerged

NotYet **2075**; Refused **2104**. **measured on a checker-rejected program**.

| Kind | Exact reason | Count |
| --- | --- | ---: |
| Refused | a type predicate | 588 |
| Refused | the non-null assertion ! | 520 |
| NotYet | reading SyntaxKind | 484 |
| NotYet | a function without a body | 191 |
| Refused | enum | 162 |
| NotYet | an EnumDeclaration | 154 |
| NotYet | a PrefixUnaryExpression on a value | 86 |
| NotYet | a method call through a structural signature in a program with statics; use typeof the declaring class | 82 |
| Refused | an ExportDeclaration | 77 |
| NotYet | a function returning T \| undefined | 66 |

Checker diagnostics: 845; own-file zero diagnostics: 25/79. Functions attempted: 2655; own-body skips: 102. All lowering outcomes: `{'NotYet': 2075, 'Refused': 2104, 'SkippedDependency': 4, 'error': 0, 'panic': 0}`.

# Provenance and reproduction

Unchanged checker-body eligibility and independent-unit lowering; each untouched commit uses its own apply.sh and adapted source. No feature merges or compiler source edits.

Both scratch heads equal their fetched origin pins, not merge commits made for this measurement. Compiler files, loader options and all adaptation files match the pinned trees byte for byte. The existing scratch overlay supplies measurement hooks and disables ordinary output APIs. Each config uses the same cohere pin recorded by both original trees; dependency and adaptation hashes, patch sets and command logs are archived. Main has 78 source roots; area/stage3 has 79, including its adapter-created hostErrors.ts helper. The two source manifests differ, so this is a two-tree observation, not an isolated compiler-feature delta.

- main-unmerged: `b8fb957aa839a9e8cb0b54279dd9864fa317bd30` on never-pushed `scratch/latent-unmerged4-main`; adapted input `/tmp/tsc-latent4-main`; elapsed seconds `{'overlay': 0.022, 'build': 17.497, 'run': 185.878}`.
- area-unmerged: `e39a299323cad76aea44ec71ca7b740130323d4c` on never-pushed `scratch/latent-unmerged4-area`; adapted input `/tmp/tsc-latent4-area`; elapsed seconds `{'overlay': 0.023, 'build': 17.761, 'run': 180.496}`.

```sh
source /workspace/adamic-tools/env.sh
bash /tmp/latent4-main/stage3/apply.sh /tmp/tsc-latent4-main > /tmp/latent4-main-apply.log 2>&1
bash /tmp/latent4-area/stage3/apply.sh /tmp/tsc-latent4-area > /tmp/latent4-area-apply.log 2>&1
python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent4-main /tmp/latent4-runs /tmp/latent4-worktrees.json > /tmp/latent4-comparisons.log 2>&1
python3 stage3/census/latent/unmerged4.py /tmp/latent4-runs > /tmp/latent4-summary.log 2>&1
python3 stage3/census/latent/audit_unmerged4.py > /tmp/latent4-audit.log 2>&1
python3 stage3/census/latent/audit_corpus.py /tmp/latent4-runs /tmp/tsc-latent4-main main-unmerged /tmp/latent4-main-unmerged-mutant > /tmp/latent4-main-unmerged-corpus-audit.log 2>&1
GOWORK=/tmp/latent4-runs/main-unmerged.go.work python3 stage3/census/latent/audit_output_guards.py /tmp/latent4-main /tmp/latent4-runs/main-unmerged-overlay /tmp/latent4-main-unmerged-guards > /tmp/latent4-main-unmerged-guards.log 2>&1
python3 stage3/census/latent/audit_corpus.py /tmp/latent4-runs /tmp/tsc-latent4-area area-unmerged /tmp/latent4-area-unmerged-mutant > /tmp/latent4-area-unmerged-corpus-audit.log 2>&1
GOWORK=/tmp/latent4-runs/area-unmerged.go.work python3 stage3/census/latent/audit_output_guards.py /tmp/latent4-area /tmp/latent4-runs/area-unmerged-overlay /tmp/latent4-area-unmerged-guards > /tmp/latent4-area-unmerged-guards.log 2>&1
```

The audit independently deduplicates all raw events, attributes findings by actual location, recounts each reason and file, verifies top-10 order/counts and body skip spans, and checks the unmerged head/dependency/source manifests. Dedicated artifact mutants corrupt totals, reason counts, rank counts/order, pins, source hashes, the actual area source denominator and body skip evidence. The real-corpus extra-NotYet mutant runs on each binary and must change only binder.ts by exactly one unique site. An initial attempt to run the two corpus mutants with a shared output filename was stopped; only the isolated reruns are evidence. No internal/ edits are committed on the delivery branch.

Toolchain setup: Go ready 0s, clang 1s, Node 1s, submodules 1s, cache warm 33s; total 33s, nproc 5, CPU quota 4. Exact timings are in the archived setup log.

# main-unmerged full per-file counts

| File | Checker | NotYet | Refused |
| --- | ---: | ---: | ---: |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 1 |
| src/compiler/_namespaces/ts.performance.ts | 0 | 0 | 1 |
| src/compiler/_namespaces/ts.ts | 0 | 0 | 75 |
| src/compiler/binder.ts | 14 | 8 | 2 |
| src/compiler/builder.ts | 22 | 32 | 36 |
| src/compiler/builderPublic.ts | 0 | 8 | 0 |
| src/compiler/builderState.ts | 4 | 3 | 20 |
| src/compiler/builderStatePublic.ts | 0 | 0 | 0 |
| src/compiler/checker.ts | 255 | 34 | 32 |
| src/compiler/commandLineParser.ts | 15 | 73 | 72 |
| src/compiler/core.ts | 10 | 228 | 153 |
| src/compiler/corePublic.ts | 1 | 1 | 2 |
| src/compiler/debug.ts | 6 | 2 | 99 |
| src/compiler/diagnosticInformationMap.generated.ts | 1 | 2 | 0 |
| src/compiler/emitter.ts | 14 | 20 | 15 |
| src/compiler/executeCommandLine.ts | 5 | 22 | 26 |
| src/compiler/expressionToTypeNode.ts | 5 | 2 | 1 |
| src/compiler/factory/baseNodeFactory.ts | 0 | 1 | 0 |
| src/compiler/factory/emitHelpers.ts | 1 | 6 | 10 |
| src/compiler/factory/emitNode.ts | 2 | 27 | 8 |
| src/compiler/factory/nodeChildren.ts | 0 | 5 | 1 |
| src/compiler/factory/nodeConverters.ts | 0 | 1 | 11 |
| src/compiler/factory/nodeFactory.ts | 5 | 16 | 5 |
| src/compiler/factory/nodeTests.ts | 0 | 227 | 227 |
| src/compiler/factory/parenthesizerRules.ts | 1 | 1 | 2 |
| src/compiler/factory/utilities.ts | 17 | 67 | 118 |
| src/compiler/factory/utilitiesPublic.ts | 0 | 3 | 3 |
| src/compiler/moduleNameResolver.ts | 33 | 70 | 48 |
| src/compiler/moduleSpecifiers.ts | 10 | 17 | 28 |
| src/compiler/parser.ts | 38 | 46 | 121 |
| src/compiler/path.ts | 1 | 39 | 12 |
| src/compiler/performance.ts | 0 | 8 | 1 |
| src/compiler/performanceCore.ts | 2 | 1 | 2 |
| src/compiler/program.ts | 22 | 24 | 45 |
| src/compiler/programDiagnostics.ts | 0 | 0 | 29 |
| src/compiler/resolutionCache.ts | 13 | 12 | 8 |
| src/compiler/scanner.ts | 19 | 47 | 17 |
| src/compiler/semver.ts | 0 | 8 | 10 |
| src/compiler/sourcemap.ts | 3 | 11 | 27 |
| src/compiler/symbolWalker.ts | 0 | 1 | 7 |
| src/compiler/sys.ts | 61 | 20 | 33 |
| src/compiler/tracing.ts | 8 | 1 | 11 |
| src/compiler/transformer.ts | 8 | 5 | 4 |
| src/compiler/transformers/classFields.ts | 16 | 7 | 5 |
| src/compiler/transformers/classThis.ts | 0 | 5 | 4 |
| src/compiler/transformers/declarations.ts | 4 | 9 | 3 |
| src/compiler/transformers/declarations/diagnostics.ts | 1 | 2 | 12 |
| src/compiler/transformers/destructuring.ts | 1 | 13 | 27 |
| src/compiler/transformers/es2015.ts | 12 | 6 | 6 |
| src/compiler/transformers/es2016.ts | 0 | 0 | 10 |
| src/compiler/transformers/es2017.ts | 10 | 3 | 2 |
| src/compiler/transformers/es2018.ts | 6 | 2 | 2 |
| src/compiler/transformers/es2019.ts | 0 | 1 | 0 |
| src/compiler/transformers/es2020.ts | 0 | 0 | 6 |
| src/compiler/transformers/es2021.ts | 0 | 0 | 1 |
| src/compiler/transformers/esDecorators.ts | 6 | 0 | 0 |
| src/compiler/transformers/esnext.ts | 7 | 5 | 2 |
| src/compiler/transformers/generators.ts | 32 | 6 | 6 |
| src/compiler/transformers/jsx.ts | 5 | 0 | 0 |
| src/compiler/transformers/legacyDecorators.ts | 1 | 0 | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | 6 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 |
| src/compiler/transformers/module/module.ts | 5 | 0 | 0 |
| src/compiler/transformers/module/system.ts | 6 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 0 | 17 | 5 |
| src/compiler/transformers/taggedTemplate.ts | 1 | 2 | 4 |
| src/compiler/transformers/ts.ts | 8 | 2 | 2 |
| src/compiler/transformers/typeSerializer.ts | 0 | 0 | 2 |
| src/compiler/transformers/utilities.ts | 4 | 30 | 24 |
| src/compiler/tsbuild.ts | 1 | 3 | 2 |
| src/compiler/tsbuildPublic.ts | 19 | 52 | 55 |
| src/compiler/types.ts | 73 | 76 | 81 |
| src/compiler/utilities.ts | 46 | 517 | 343 |
| src/compiler/utilitiesPublic.ts | 7 | 139 | 144 |
| src/compiler/visitorPublic.ts | 1 | 33 | 19 |
| src/compiler/watch.ts | 4 | 17 | 41 |
| src/compiler/watchPublic.ts | 13 | 6 | 2 |
| src/compiler/watchUtilities.ts | 6 | 7 | 7 |

# main-unmerged all exact reasons

| Kind and reason | Count |
| --- | ---: |
| Refused: a type predicate | 588 |
| NotYet: reading SyntaxKind | 484 |
| Refused: the non-null assertion ! | 484 |
| NotYet: a function without a body | 191 |
| Refused: enum | 162 |
| NotYet: an EnumDeclaration | 154 |
| NotYet: a PrefixUnaryExpression on a value | 85 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 80 |
| Refused: an ExportDeclaration | 77 |
| NotYet: a function returning T \| undefined | 66 |
| NotYet: a NonNullExpression | 60 |
| Refused: a value as a condition | 56 |
| NotYet: a function returning T | 50 |
| NotYet: reading CharacterCodes | 37 |
| NotYet: a BinaryExpression with a value and a value | 32 |
| NotYet: a value of type T | 30 |
| NotYet: a value of type any | 30 |
| NotYet: a field of type boolean \| undefined | 25 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 24 |
| Refused: a cast the runtime can't check | 23 |
| NotYet: a parameter that isn't a plain name | 22 |
| NotYet: reading ModifierFlags | 22 |
| NotYet: reading Extension | 20 |
| NotYet: reading NodeFlags | 20 |
| NotYet: a value of type Path | 19 |
| NotYet: a BinaryExpression with a value and a boolean | 18 |
| NotYet: a value of type ResolvedConfigFilePath | 18 |
| NotYet: a value of type __String | 18 |
| NotYet: a function returning U \| undefined | 16 |
| Refused: a value of type EvaluatorResult<number> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where number is read | 16 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 15 |
| Refused: \|\|= | 15 |
| Refused: a string as a condition | 14 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 |
| NotYet: a generic function as a value | 13 |
| NotYet: reading Comparison | 12 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 |
| NotYet: a value of type T \| undefined | 11 |
| NotYet: reading ModuleKind | 11 |
| Refused: a boolean \| undefined as a condition | 11 |
| Refused: a namespace | 11 |
| NotYet: a function returning __String | 10 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 10 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 |
| NotYet: a ModuleDeclaration | 9 |
| NotYet: a call through ?. (an optional call) | 9 |
| NotYet: a value of type unknown | 9 |
| NotYet: an array of T | 9 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 8 |
| NotYet: a PrefixUnaryExpression on a number | 8 |
| NotYet: a PrefixUnaryExpression on a string | 8 |
| NotYet: new an Identifier | 8 |
| Refused: an index signature | 8 |
| NotYet: a BinaryExpression as a statement | 7 |
| NotYet: an array of never | 7 |
| NotYet: for...of over an object | 7 |
| NotYet: reading EmitFlags | 7 |
| NotYet: reading NodeResolutionFeatures | 7 |
| NotYet: this outside a method | 7 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 7 |
| NotYet: a BinaryExpression with a boolean and a value | 6 |
| NotYet: a BinaryExpression with a number and a number | 6 |
| NotYet: a BinaryExpression with a number \| undefined and a number | 6 |
| NotYet: a PrefixUnaryExpression on a number \| undefined | 6 |
| NotYet: a field of type true \| undefined | 6 |
| NotYet: a function returning any | 6 |
| NotYet: a value of type ResolvedConfigFileName | 6 |
| NotYet: reading AssignmentDeclarationKind | 6 |
| NotYet: reading ModuleResolutionKind | 6 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 6 |
| Refused: a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 |
| Refused: a value of type SearchResult<undefined> seen as SearchResult<Resolved>, which can write Resolved \| undefined where undefined is read | 6 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 6 |
| NotYet: a BinaryExpression with a string and a string | 5 |
| NotYet: a BinaryExpression with a value and a number | 5 |
| NotYet: a PrefixUnaryExpression on a boolean \| undefined | 5 |
| NotYet: a function returning Path | 5 |
| NotYet: a value of type NonNullable<T> | 5 |
| NotYet: a value of type object | 5 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 5 |
| NotYet: reading Extensions | 5 |
| NotYet: reading ScriptTarget | 5 |
| NotYet: reading SymbolFlags | 5 |
| NotYet: reading TransformFlags | 5 |
| Refused: a definite assignment assertion ! | 5 |
| Refused: a generator function | 5 |
| Refused: a label | 5 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 5 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 5 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 5 |
| Refused: a value of type BindingElement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 |
| Refused: the comma operator | 5 |
| Refused: yield (generators) | 5 |
| NotYet: a BinaryExpression with a string and a boolean | 4 |
| NotYet: a field of type string \| DiagnosticMessageChain | 4 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 |
| NotYet: a function returning __String \| undefined | 4 |
| NotYet: a function with an optional or rest parameter, as a value | 4 |
| NotYet: a value of type T \| Program | 4 |
| NotYet: reading ScriptKind | 4 |
| NotYet: regex replacement other than a string | 4 |
| Refused: a method in object destructuring | 4 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 4 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node can be written, so it can write what Node can't hold | 4 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 4 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 4 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 |
| NotYet: a computed field name | 3 |
| NotYet: a function returning CompilerOptionsValue | 3 |
| NotYet: a function returning ResolvedConfigFileName | 3 |
| NotYet: a function returning T \| T[] \| undefined | 3 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 |
| NotYet: a value of type K | 3 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 3 |
| NotYet: a value of type string \| null \| undefined | 3 |
| NotYet: an ElementAccessExpression | 3 |
| NotYet: an array of U | 3 |
| NotYet: reading AccessKind | 3 |
| NotYet: reading OuterExpressionKinds | 3 |
| NotYet: reading TypeFlags | 3 |
| NotYet: reading UsingKind | 3 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 |
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 3 |
| Refused: a method read as a value (now would lose its object, and this with it) | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 |
| Refused: a union of differently held members as a condition | 3 |
| Refused: a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type DiagnosticMessageChain seen as T, a type parameter whose constraint DiagnosticMessageChain \| ReusableDiagnosticMessageChain can be written, so it can write what DiagnosticMessageChain can't hold | 3 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 3 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 |
| Refused: a value of type Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type PostfixUnaryExpression \| PrefixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 3 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 |
| Refused: the void operator | 3 |
| NotYet: a BinaryExpression with a boolean and a string | 2 |
| NotYet: a BinaryExpression with a value and a string | 2 |
| NotYet: a PrefixUnaryExpression on a union of differently held members | 2 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 2 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 2 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 |
| NotYet: a field of type true \| Node \| undefined | 2 |
| NotYet: a function inside a function (a closure) | 2 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 |
| NotYet: a function returning Path \| undefined | 2 |
| NotYet: a function returning U | 2 |
| NotYet: a function returning V | 2 |
| NotYet: a function returning object | 2 |
| NotYet: a function value returning boolean \| undefined | 2 |
| NotYet: a function value returning union of differently held members | 2 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 |
| NotYet: a value of type CompilerOptionsValue | 2 |
| NotYet: a value of type T \| T[] | 2 |
| NotYet: a value of type TInArray | 2 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 |
| NotYet: new a ParenthesizedExpression | 2 |
| NotYet: reading BuilderFileEmit | 2 |
| NotYet: reading BuilderProgramKind | 2 |
| NotYet: reading DiagnosticCategory | 2 |
| NotYet: reading FileWatcherEventKind | 2 |
| NotYet: reading JsxEmit | 2 |
| NotYet: reading ModuleInstanceState | 2 |
| NotYet: reading NodeFactoryFlags | 2 |
| NotYet: reading SignatureFlags | 2 |
| NotYet: reading TypePredicateKind | 2 |
| NotYet: reading getOptionsNameMap | 2 |
| Refused: a function taking CallExpression \| (IncludeTypeSpaceImports extends false ? never : ImportTypeNode \| JSDocImportTag) seen as one taking CallExpression \| ImportTypeNode \| JSDocImportTag (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 |
| Refused: a number \| undefined as a condition | 2 |
| Refused: a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Declaration seen as T, a type parameter whose constraint Declaration can be written, so it can write what Declaration can't hold | 2 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 |
| Refused: a value of type ElementAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string is read | 2 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type JsxTagNameExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type LeftHandSideExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MethodDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NodeArray<ClassElement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NodeArray<T> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ParameterDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type PropertyAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type PropertyName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Token<SyntaxKind> seen as T, a type parameter whose constraint Node can be written, so it can write what Token<SyntaxKind> can't hold | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as CompilerOptionsValue, which can write string \| number where string is read | 2 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what any can't hold | 2 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 2 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 |
| Refused: a value of type { readonly args: readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...]; readonly kind: PragmaKindFlags.... seen as PragmaDefinition<string, string, string, string>, whose readonly field args becomes writable: a readonly field may hold something narrower than readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...], which a write of readonly [PragmaArgumentSpecification<string>] \| readonly [PragmaArgumentSpecification<string>, PragmaArgumentSpecification<string>] \| ... \| ... \| undefined would replace | 2 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: false; }; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: false; } is read | 2 |
| Refused: in | 2 |
| Refused: var | 2 |
| NotYet: .length on a value | 1 |
| NotYet: ?.[] on a value | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 |
| NotYet: RegExp with a nonconstant pattern | 1 |
| NotYet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 |
| NotYet: a BinaryExpression with a boolean and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a string and a number | 1 |
| NotYet: a ClassExpression | 1 |
| NotYet: a Map of T | 1 |
| NotYet: a PostfixUnaryExpression | 1 |
| NotYet: a SatisfiesExpression | 1 |
| NotYet: a YieldExpression as a statement | 1 |
| NotYet: a boolean \| undefined variable a function value captures | 1 |
| NotYet: a class instantiated with TOuterState | 1 |
| NotYet: a destructured parameter beside a parameter with a default | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 |
| NotYet: a field of type AnyBuildOrder \| undefined | 1 |
| NotYet: a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | 1 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 |
| NotYet: a field of type string \| false \| undefined | 1 |
| NotYet: a field of type string \| number \| undefined | 1 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 |
| NotYet: a function returning CanonicalKey | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 |
| NotYet: a function returning HasJSDoc \| undefined | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 |
| NotYet: a function returning ModeAwareCacheKey | 1 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 |
| NotYet: a function returning PathPathComponents | 1 |
| NotYet: a function returning ResolvedConfigFilePath | 1 |
| NotYet: a function returning T \| T[] | 1 |
| NotYet: a function returning T \| readonly T[] | 1 |
| NotYet: a function returning T1 & T2 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 |
| NotYet: a function returning TOut | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning TOut \| undefined | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 |
| NotYet: a function returning object \| undefined | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning string \| object | 1 |
| NotYet: a function returning undefined | 1 |
| NotYet: a function returning unknown | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a number \| undefined argument to substring | 1 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a value of type AccessExpression \| RequireOrImportCall | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type HasJSDoc | 1 |
| NotYet: a value of type HasJSDoc \| undefined | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 |
| NotYet: a value of type NamedEvaluation | 1 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type RequireOrImportCall | 1 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type SourceFile | 1 |
| NotYet: a value of type T \| readonly T[] | 1 |
| NotYet: a value of type T1 | 1 |
| NotYet: a value of type TData | 1 |
| NotYet: a value of type TEntry | 1 |
| NotYet: a value of type T["kind"] | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 |
| NotYet: a value of type U | 1 |
| NotYet: a value of type U \| readonly U[] \| undefined | 1 |
| NotYet: a value of type V | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 |
| NotYet: a value of type WrappedExpression<T> | 1 |
| NotYet: a value of type __String & string | 1 |
| NotYet: a value of type false \| RegExpExecArray \| null | 1 |
| NotYet: a value of type object \| undefined | 1 |
| NotYet: a value of type undefined | 1 |
| NotYet: a void call used as a value | 1 |
| NotYet: an array of T \| U | 1 |
| NotYet: an array of V | 1 |
| NotYet: an array of unknown | 1 |
| NotYet: assigning a field of a value | 1 |
| NotYet: lastIndexOf with these arguments | 1 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 |
| NotYet: optional chaining to .size on a value | 1 |
| NotYet: reading Error | 1 |
| NotYet: reading FileIncludeKind | 1 |
| NotYet: reading FlattenLevel | 1 |
| NotYet: reading ImportsNotUsedAsValues | 1 |
| NotYet: reading Instruction | 1 |
| NotYet: reading InternalNodeBuilderFlags | 1 |
| NotYet: reading IntrinsicTypeKind | 1 |
| NotYet: reading IterationTypeKind | 1 |
| NotYet: reading JSDocParsingMode | 1 |
| NotYet: reading ListFormat | 1 |
| NotYet: reading ModuleDetectionKind | 1 |
| NotYet: reading ModuleSpecifierEnding | 1 |
| NotYet: reading NewLineKind | 1 |
| NotYet: reading NodeBuilderFlags | 1 |
| NotYet: reading PragmaKindFlags | 1 |
| NotYet: reading RegularExpressionFlags | 1 |
| NotYet: reading SnippetKind | 1 |
| NotYet: reading StatisticType | 1 |
| NotYet: reading TypeFacts | 1 |
| NotYet: reading UpToDateStatusType | 1 |
| NotYet: reading WatchFileKind | 1 |
| NotYet: reading addAggregateStatistic | 1 |
| NotYet: reading addOutput | 1 |
| NotYet: reading captureMapping | 1 |
| NotYet: reading convertToFunctionBlock | 1 |
| NotYet: reading createBaseSourceFileNode | 1 |
| NotYet: reading createIntlCollatorStringComparer | 1 |
| NotYet: reading createPollingIntervalQueue | 1 |
| NotYet: reading enter | 1 |
| NotYet: reading getAccessorNameVisibilityError | 1 |
| NotYet: reading getPackageJsonInfo | 1 |
| NotYet: reading getSymbolWalker | 1 |
| NotYet: reading optionDependsOnRecursive | 1 |
| NotYet: reading parseStrings | 1 |
| NotYet: reading transformSourceFile | 1 |
| NotYet: reading transformSourceFileOrBundle | 1 |
| NotYet: spreading an array of other elements | 1 |
| NotYet: storing any in a field | 1 |
| NotYet: storing string \| number in a field | 1 |
| NotYet: storing true \| Node \| undefined in a field | 1 |
| Refused: a function taking (node: Node) => T \| undefined seen as one taking (node: Node) => T \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking (value: never, key: never, map: ReadonlyMap<never, never>) => void seen as one taking <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey][] \| PragmaPseudoMap[TKey], key: TKey, map: ReadonlyPragmaMap) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking OrdinalParentheizerRuleSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking ParenthesizerRule<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking ParenthesizerRuleOrSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLink would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (deleteFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (log would lose its object, and this with it) | 1 |
| Refused: a method read as a value (realpath would lose its object, and this with it) | 1 |
| Refused: a method read as a value (remove would lose its object, and this with it) | 1 |
| Refused: a method read as a value (repeat would lose its object, and this with it) | 1 |
| Refused: a method read as a value (replace would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportTruncationError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 |
| Refused: a number as a condition | 1 |
| Refused: a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ArrayBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BigIntLiteral \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BigIntLiteral \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BuilderProgram seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what BuilderProgram can't hold | 1 |
| Refused: a value of type BuilderProgram \| Program seen as BuilderProgram, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type CallExpression seen as RequireOrImportCall, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & Identifier would replace | 1 |
| Refused: a value of type CallExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfListType, which can write "list" \| "listOrElement" where "boolean" is read | 1 |
| Refused: a value of type CommandLineOption \| undefined seen as TsConfigOnlyOption, which can write "object" where "boolean" is read | 1 |
| Refused: a value of type CommandLineOptionOfBooleanType \| CommandLineOptionOfCustomType \| CommandLineOptionOfNumberType \| CommandLineOptionOfStringType \| TsConfigOnlyOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 1 |
| Refused: a value of type CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 |
| Refused: a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type EvaluatorResult<string \| undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string \| undefined is read | 1 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where string is read | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where undefined is read | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where undefined is read | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExpressionWithTypeArguments seen as ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; }, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & (Identifier \| PropertyAccessEntityNameExpression) would replace | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FlowNode seen as FlowLabel, which can write undefined where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as T, a type parameter whose constraint Node can be written, so it can write what Identifier can't hold | 1 |
| Refused: a value of type ImportTypeNode \| undefined seen as ValidImportTypeNode \| undefined, whose readonly field argument becomes writable: a readonly field may hold something narrower than TypeNode, which a write of LiteralTypeNode & { literal: StringLiteral; } would replace | 1 |
| Refused: a value of type JSDocNullableType seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JsxOpeningFragment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type LeftHandSideExpression \| UnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type LiteralLikeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as SyntaxList, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.SyntaxList would replace | 1 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what Node can't hold | 1 |
| Refused: a value of type Node seen as TemplateLiteralTypeSpan, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.TemplateLiteralType would replace | 1 |
| Refused: a value of type Node \| NodeArray<Node> seen as NodeArray<Node>, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ObjectBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalChain seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type ParenthesizedExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 |
| Refused: a value of type PostfixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PrivateIdentifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type QualifiedName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 |
| Refused: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 |
| Refused: a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, which can write { resolved: Resolved; isExternalLibraryImport: true; } \| undefined where Resolved \| undefined is read | 1 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 |
| Refused: a value of type StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TemplateLiteralLikeNode seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 |
| Refused: a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclarationList seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VisitEachChildTable seen as Record<SyntaxKind, VisitEachChildFunction<any> \| undefined>, which can write VisitEachChildFunction<any> \| undefined where VisitEachChildFunction<QualifiedName> is read | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 |
| Refused: a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 |
| Refused: a value of type readonly T[] \| undefined seen as unknown[], which can write unknown where T is read | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 |
| Refused: a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; compilerOptions: CompilerOptions; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: ... \| undefined; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write DiagnosticReporter where (_?: unknown) => void is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 |
| Refused: a value of type { kind: "ambient" \| "node_modules" \| "paths" \| "redirect" \| "relative" \| undefined; moduleSpecifiers: readonly string[]; computedWithoutCache: false; } \| undefined seen as ModuleSpecifierResult \| undefined, which can write boolean where false is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 |
| Refused: a value of type { matchableStringSet: Set<string> \| undefined; patterns: Pattern[] \| undefined; } seen as ParsedPatterns, which can write ReadonlySet<string> \| undefined where Set<string> \| undefined is read | 1 |
| Refused: a value of type { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined seen as Resolved \| undefined, which can write string \| true \| undefined where string \| true is read | 1 |
| Refused: a value of type { resolved: Resolved; isExternalLibraryImport: true; } \| undefined seen as { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined, which can write boolean where true is read | 1 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: true; } \| undefined is read | 1 |
| Refused: an import cycle | 1 |
| Refused: arguments | 1 |
| Refused: debugger | 1 |
| Refused: delete | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 |
| Refused: inherited library member prototype read as an own field | 1 |
| Refused: inherited library member replace read as an own field | 1 |
| Refused: instantiating a generic function makes a value of type T \| undefined written where T is read | 1 |

# area-unmerged full per-file counts

| File | Checker | NotYet | Refused |
| --- | ---: | ---: | ---: |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 1 |
| src/compiler/_namespaces/ts.performance.ts | 0 | 0 | 1 |
| src/compiler/_namespaces/ts.ts | 0 | 0 | 75 |
| src/compiler/binder.ts | 14 | 8 | 2 |
| src/compiler/builder.ts | 22 | 32 | 36 |
| src/compiler/builderPublic.ts | 0 | 8 | 0 |
| src/compiler/builderState.ts | 2 | 3 | 20 |
| src/compiler/builderStatePublic.ts | 0 | 0 | 0 |
| src/compiler/checker.ts | 255 | 34 | 32 |
| src/compiler/commandLineParser.ts | 15 | 73 | 72 |
| src/compiler/core.ts | 5 | 230 | 153 |
| src/compiler/corePublic.ts | 1 | 1 | 2 |
| src/compiler/debug.ts | 5 | 2 | 98 |
| src/compiler/diagnosticInformationMap.generated.ts | 0 | 2 | 0 |
| src/compiler/emitter.ts | 11 | 21 | 16 |
| src/compiler/executeCommandLine.ts | 2 | 24 | 29 |
| src/compiler/expressionToTypeNode.ts | 5 | 2 | 1 |
| src/compiler/factory/baseNodeFactory.ts | 0 | 1 | 0 |
| src/compiler/factory/emitHelpers.ts | 1 | 6 | 10 |
| src/compiler/factory/emitNode.ts | 2 | 27 | 8 |
| src/compiler/factory/nodeChildren.ts | 0 | 5 | 1 |
| src/compiler/factory/nodeConverters.ts | 0 | 1 | 0 |
| src/compiler/factory/nodeFactory.ts | 5 | 16 | 5 |
| src/compiler/factory/nodeTests.ts | 0 | 227 | 227 |
| src/compiler/factory/parenthesizerRules.ts | 1 | 1 | 2 |
| src/compiler/factory/utilities.ts | 17 | 67 | 92 |
| src/compiler/factory/utilitiesPublic.ts | 0 | 3 | 3 |
| src/compiler/hostErrors.ts | 0 | 2 | 2 |
| src/compiler/moduleNameResolver.ts | 33 | 70 | 36 |
| src/compiler/moduleSpecifiers.ts | 7 | 20 | 34 |
| src/compiler/parser.ts | 36 | 47 | 114 |
| src/compiler/path.ts | 1 | 39 | 12 |
| src/compiler/performance.ts | 0 | 8 | 1 |
| src/compiler/performanceCore.ts | 2 | 1 | 2 |
| src/compiler/program.ts | 22 | 24 | 45 |
| src/compiler/programDiagnostics.ts | 0 | 0 | 29 |
| src/compiler/resolutionCache.ts | 13 | 12 | 8 |
| src/compiler/scanner.ts | 17 | 47 | 21 |
| src/compiler/semver.ts | 0 | 8 | 10 |
| src/compiler/sourcemap.ts | 2 | 11 | 27 |
| src/compiler/symbolWalker.ts | 0 | 1 | 7 |
| src/compiler/sys.ts | 60 | 20 | 40 |
| src/compiler/tracing.ts | 6 | 1 | 11 |
| src/compiler/transformer.ts | 8 | 5 | 4 |
| src/compiler/transformers/classFields.ts | 16 | 7 | 5 |
| src/compiler/transformers/classThis.ts | 0 | 5 | 3 |
| src/compiler/transformers/declarations.ts | 3 | 9 | 3 |
| src/compiler/transformers/declarations/diagnostics.ts | 1 | 2 | 12 |
| src/compiler/transformers/destructuring.ts | 1 | 13 | 26 |
| src/compiler/transformers/es2015.ts | 12 | 6 | 6 |
| src/compiler/transformers/es2016.ts | 0 | 0 | 0 |
| src/compiler/transformers/es2017.ts | 10 | 3 | 2 |
| src/compiler/transformers/es2018.ts | 2 | 2 | 23 |
| src/compiler/transformers/es2019.ts | 0 | 1 | 0 |
| src/compiler/transformers/es2020.ts | 0 | 0 | 3 |
| src/compiler/transformers/es2021.ts | 0 | 0 | 1 |
| src/compiler/transformers/esDecorators.ts | 6 | 0 | 0 |
| src/compiler/transformers/esnext.ts | 7 | 5 | 2 |
| src/compiler/transformers/generators.ts | 32 | 6 | 6 |
| src/compiler/transformers/jsx.ts | 5 | 0 | 0 |
| src/compiler/transformers/legacyDecorators.ts | 1 | 0 | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | 6 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 |
| src/compiler/transformers/module/module.ts | 1 | 0 | 0 |
| src/compiler/transformers/module/system.ts | 2 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 0 | 17 | 4 |
| src/compiler/transformers/taggedTemplate.ts | 1 | 2 | 3 |
| src/compiler/transformers/ts.ts | 5 | 2 | 2 |
| src/compiler/transformers/typeSerializer.ts | 0 | 0 | 1 |
| src/compiler/transformers/utilities.ts | 4 | 30 | 23 |
| src/compiler/tsbuild.ts | 1 | 3 | 2 |
| src/compiler/tsbuildPublic.ts | 19 | 52 | 55 |
| src/compiler/types.ts | 73 | 76 | 81 |
| src/compiler/utilities.ts | 41 | 520 | 316 |
| src/compiler/utilitiesPublic.ts | 7 | 139 | 144 |
| src/compiler/visitorPublic.ts | 0 | 34 | 18 |
| src/compiler/watch.ts | 4 | 17 | 41 |
| src/compiler/watchPublic.ts | 13 | 6 | 2 |
| src/compiler/watchUtilities.ts | 2 | 8 | 31 |

# area-unmerged all exact reasons

| Kind and reason | Count |
| --- | ---: |
| Refused: a type predicate | 588 |
| Refused: the non-null assertion ! | 520 |
| NotYet: reading SyntaxKind | 484 |
| NotYet: a function without a body | 191 |
| Refused: enum | 162 |
| NotYet: an EnumDeclaration | 154 |
| NotYet: a PrefixUnaryExpression on a value | 86 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 82 |
| Refused: an ExportDeclaration | 77 |
| NotYet: a function returning T \| undefined | 66 |
| NotYet: a NonNullExpression | 63 |
| Refused: a value as a condition | 56 |
| NotYet: a function returning T | 50 |
| NotYet: reading CharacterCodes | 37 |
| NotYet: a BinaryExpression with a value and a value | 32 |
| NotYet: a value of type T | 30 |
| NotYet: a value of type any | 30 |
| NotYet: a field of type boolean \| undefined | 25 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 25 |
| Refused: a cast the runtime can't check | 23 |
| NotYet: a parameter that isn't a plain name | 22 |
| NotYet: reading ModifierFlags | 22 |
| NotYet: reading Extension | 21 |
| NotYet: reading NodeFlags | 20 |
| NotYet: a value of type Path | 19 |
| NotYet: a BinaryExpression with a value and a boolean | 18 |
| NotYet: a value of type ResolvedConfigFilePath | 18 |
| NotYet: a value of type __String | 18 |
| Refused: \|\|= | 18 |
| NotYet: a function returning U \| undefined | 16 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 15 |
| Refused: a string as a condition | 14 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 |
| NotYet: a generic function as a value | 13 |
| NotYet: a value of type unknown | 13 |
| NotYet: reading Comparison | 12 |
| Refused: a boolean \| undefined as a condition | 12 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 |
| NotYet: a value of type T \| undefined | 11 |
| NotYet: reading ModuleKind | 11 |
| Refused: a namespace | 11 |
| Refused: an index signature | 11 |
| NotYet: a call through ?. (an optional call) | 10 |
| NotYet: a function returning __String | 10 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 10 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 10 |
| NotYet: a ModuleDeclaration | 9 |
| NotYet: a PrefixUnaryExpression on a string | 9 |
| NotYet: an array of T | 9 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 8 |
| NotYet: a PrefixUnaryExpression on a number | 8 |
| NotYet: new an Identifier | 8 |
| NotYet: a BinaryExpression as a statement | 7 |
| NotYet: an array of never | 7 |
| NotYet: for...of over an object | 7 |
| NotYet: reading EmitFlags | 7 |
| NotYet: reading NodeResolutionFeatures | 7 |
| NotYet: this outside a method | 7 |
| NotYet: a BinaryExpression with a boolean and a value | 6 |
| NotYet: a BinaryExpression with a number and a number | 6 |
| NotYet: a BinaryExpression with a number \| undefined and a number | 6 |
| NotYet: a PrefixUnaryExpression on a number \| undefined | 6 |
| NotYet: a field of type true \| undefined | 6 |
| NotYet: a function returning any | 6 |
| NotYet: a value of type ResolvedConfigFileName | 6 |
| NotYet: reading AssignmentDeclarationKind | 6 |
| NotYet: reading ModuleResolutionKind | 6 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 6 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 6 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 6 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 6 |
| NotYet: a BinaryExpression with a string and a string | 5 |
| NotYet: a BinaryExpression with a value and a number | 5 |
| NotYet: a PrefixUnaryExpression on a boolean \| undefined | 5 |
| NotYet: a function returning Path | 5 |
| NotYet: a value of type NonNullable<T> | 5 |
| NotYet: a value of type object | 5 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 5 |
| NotYet: reading Extensions | 5 |
| NotYet: reading ScriptTarget | 5 |
| NotYet: reading SymbolFlags | 5 |
| NotYet: reading TransformFlags | 5 |
| Refused: a definite assignment assertion ! | 5 |
| Refused: a generator function | 5 |
| Refused: a label | 5 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 5 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 5 |
| Refused: the comma operator | 5 |
| Refused: yield (generators) | 5 |
| NotYet: a BinaryExpression with a string and a boolean | 4 |
| NotYet: a field of type string \| DiagnosticMessageChain | 4 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 |
| NotYet: a function returning __String \| undefined | 4 |
| NotYet: a function with an optional or rest parameter, as a value | 4 |
| NotYet: a value of type T \| Program | 4 |
| NotYet: reading ScriptKind | 4 |
| NotYet: regex replacement other than a string | 4 |
| Refused: a method in object destructuring | 4 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 4 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 |
| Refused: a method read as a value (readDirectory would lose its object, and this with it) | 4 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node can be written, so it can write what Node can't hold | 4 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 4 |
| Refused: in | 4 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 4 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 |
| NotYet: a computed field name | 3 |
| NotYet: a function returning CompilerOptionsValue | 3 |
| NotYet: a function returning ResolvedConfigFileName | 3 |
| NotYet: a function returning T \| T[] \| undefined | 3 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 |
| NotYet: a value of type K | 3 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 3 |
| NotYet: a value of type string \| null \| undefined | 3 |
| NotYet: an ElementAccessExpression | 3 |
| NotYet: an array of U | 3 |
| NotYet: reading AccessKind | 3 |
| NotYet: reading OuterExpressionKinds | 3 |
| NotYet: reading TypeFlags | 3 |
| NotYet: reading UsingKind | 3 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 |
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 3 |
| Refused: a method read as a value (now would lose its object, and this with it) | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 |
| Refused: a union of differently held members as a condition | 3 |
| Refused: a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type DiagnosticMessageChain seen as T, a type parameter whose constraint DiagnosticMessageChain \| ReusableDiagnosticMessageChain can be written, so it can write what DiagnosticMessageChain can't hold | 3 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 3 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 3 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 |
| Refused: the void operator | 3 |
| NotYet: a BinaryExpression with a boolean and a string | 2 |
| NotYet: a BinaryExpression with a value and a string | 2 |
| NotYet: a PrefixUnaryExpression on a union of differently held members | 2 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 2 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 2 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 |
| NotYet: a field of type true \| Node \| undefined | 2 |
| NotYet: a function inside a function (a closure) | 2 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 |
| NotYet: a function returning Path \| undefined | 2 |
| NotYet: a function returning U | 2 |
| NotYet: a function returning V | 2 |
| NotYet: a function returning object | 2 |
| NotYet: a function value returning boolean \| undefined | 2 |
| NotYet: a function value returning union of differently held members | 2 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 |
| NotYet: a value of type CompilerOptionsValue | 2 |
| NotYet: a value of type T \| T[] | 2 |
| NotYet: a value of type TInArray | 2 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 |
| NotYet: new a ParenthesizedExpression | 2 |
| NotYet: reading BuilderFileEmit | 2 |
| NotYet: reading BuilderProgramKind | 2 |
| NotYet: reading DiagnosticCategory | 2 |
| NotYet: reading FileWatcherEventKind | 2 |
| NotYet: reading JsxEmit | 2 |
| NotYet: reading ModuleInstanceState | 2 |
| NotYet: reading NodeFactoryFlags | 2 |
| NotYet: reading SignatureFlags | 2 |
| NotYet: reading TypePredicateKind | 2 |
| NotYet: reading getOptionsNameMap | 2 |
| Refused: a function taking CallExpression \| (IncludeTypeSpaceImports extends false ? never : ImportTypeNode \| JSDocImportTag) seen as one taking CallExpression \| ImportTypeNode \| JSDocImportTag (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking PollingInterval seen as one taking PollingInterval \| WatchDirectoryFlags (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 |
| Refused: a number \| undefined as a condition | 2 |
| Refused: a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Block seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Declaration seen as T, a type parameter whose constraint Declaration can be written, so it can write what Declaration can't hold | 2 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 |
| Refused: a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 2 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 |
| Refused: a value of type Statement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Token<SyntaxKind> seen as T, a type parameter whose constraint Node can be written, so it can write what Token<SyntaxKind> can't hold | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as CompilerOptionsValue, which can write string \| number where string is read | 2 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what any can't hold | 2 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 2 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 |
| Refused: a value of type { readonly args: readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...]; readonly kind: PragmaKindFlags.... seen as PragmaDefinition<string, string, string, string>, whose readonly field args becomes writable: a readonly field may hold something narrower than readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...], which a write of readonly [PragmaArgumentSpecification<string>] \| readonly [PragmaArgumentSpecification<string>, PragmaArgumentSpecification<string>] \| ... \| ... \| undefined would replace | 2 |
| Refused: var | 2 |
| NotYet: .length on a value | 1 |
| NotYet: ?.[] on a value | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 |
| NotYet: RegExp with a nonconstant pattern | 1 |
| NotYet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 |
| NotYet: a BinaryExpression with a boolean and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a string and a number | 1 |
| NotYet: a ClassExpression | 1 |
| NotYet: a Map of T | 1 |
| NotYet: a PostfixUnaryExpression | 1 |
| NotYet: a SatisfiesExpression | 1 |
| NotYet: a YieldExpression as a statement | 1 |
| NotYet: a boolean \| undefined variable a function value captures | 1 |
| NotYet: a class instantiated with TOuterState | 1 |
| NotYet: a destructured parameter beside a parameter with a default | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 |
| NotYet: a field of type AnyBuildOrder \| undefined | 1 |
| NotYet: a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | 1 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 |
| NotYet: a field of type string \| false \| undefined | 1 |
| NotYet: a field of type string \| number \| undefined | 1 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 |
| NotYet: a function returning CanonicalKey | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 |
| NotYet: a function returning HasJSDoc \| undefined | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 |
| NotYet: a function returning ModeAwareCacheKey | 1 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 |
| NotYet: a function returning PathPathComponents | 1 |
| NotYet: a function returning ResolvedConfigFilePath | 1 |
| NotYet: a function returning T \| T[] | 1 |
| NotYet: a function returning T \| readonly T[] | 1 |
| NotYet: a function returning T1 & T2 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 |
| NotYet: a function returning TOut | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning TOut \| undefined | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 |
| NotYet: a function returning object \| undefined | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning string \| object | 1 |
| NotYet: a function returning undefined | 1 |
| NotYet: a function returning unknown | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a number \| undefined argument to substring | 1 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a value of type AccessExpression \| RequireOrImportCall | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type HasJSDoc | 1 |
| NotYet: a value of type HasJSDoc \| undefined | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 |
| NotYet: a value of type NamedEvaluation | 1 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type RequireOrImportCall | 1 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type SourceFile | 1 |
| NotYet: a value of type T \| readonly T[] | 1 |
| NotYet: a value of type T1 | 1 |
| NotYet: a value of type TData | 1 |
| NotYet: a value of type TEntry | 1 |
| NotYet: a value of type T["kind"] | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 |
| NotYet: a value of type U | 1 |
| NotYet: a value of type U \| readonly U[] \| undefined | 1 |
| NotYet: a value of type V | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 |
| NotYet: a value of type WrappedExpression<T> | 1 |
| NotYet: a value of type __String & string | 1 |
| NotYet: a value of type false \| RegExpExecArray \| null | 1 |
| NotYet: a value of type object \| undefined | 1 |
| NotYet: a value of type undefined | 1 |
| NotYet: a void call used as a value | 1 |
| NotYet: an array of T \| U | 1 |
| NotYet: an array of V | 1 |
| NotYet: an array of unknown | 1 |
| NotYet: assigning a field of a value | 1 |
| NotYet: lastIndexOf with these arguments | 1 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 |
| NotYet: optional chaining to .size on a value | 1 |
| NotYet: reading Error | 1 |
| NotYet: reading FileIncludeKind | 1 |
| NotYet: reading FlattenLevel | 1 |
| NotYet: reading ImportsNotUsedAsValues | 1 |
| NotYet: reading Instruction | 1 |
| NotYet: reading InternalNodeBuilderFlags | 1 |
| NotYet: reading IntrinsicTypeKind | 1 |
| NotYet: reading IterationTypeKind | 1 |
| NotYet: reading JSDocParsingMode | 1 |
| NotYet: reading ListFormat | 1 |
| NotYet: reading ModuleDetectionKind | 1 |
| NotYet: reading ModuleSpecifierEnding | 1 |
| NotYet: reading NewLineKind | 1 |
| NotYet: reading NodeBuilderFlags | 1 |
| NotYet: reading PragmaKindFlags | 1 |
| NotYet: reading RegularExpressionFlags | 1 |
| NotYet: reading SnippetKind | 1 |
| NotYet: reading StatisticType | 1 |
| NotYet: reading TypeFacts | 1 |
| NotYet: reading UpToDateStatusType | 1 |
| NotYet: reading WatchFileKind | 1 |
| NotYet: reading WatchLogLevel | 1 |
| NotYet: reading addAggregateStatistic | 1 |
| NotYet: reading addOutput | 1 |
| NotYet: reading captureMapping | 1 |
| NotYet: reading convertToFunctionBlock | 1 |
| NotYet: reading createBaseSourceFileNode | 1 |
| NotYet: reading createIntlCollatorStringComparer | 1 |
| NotYet: reading createPollingIntervalQueue | 1 |
| NotYet: reading enter | 1 |
| NotYet: reading getAccessorNameVisibilityError | 1 |
| NotYet: reading getPackageJsonInfo | 1 |
| NotYet: reading getSymbolWalker | 1 |
| NotYet: reading getUnusedExpectations | 1 |
| NotYet: reading getValueCandidate | 1 |
| NotYet: reading optionDependsOnRecursive | 1 |
| NotYet: reading parseStrings | 1 |
| NotYet: reading transformSourceFile | 1 |
| NotYet: reading transformSourceFileOrBundle | 1 |
| NotYet: spreading an array of other elements | 1 |
| NotYet: storing any in a field | 1 |
| NotYet: storing string \| number in a field | 1 |
| NotYet: storing true \| Node \| undefined in a field | 1 |
| Refused: a function taking (node: Node) => T \| undefined seen as one taking (node: Node) => T \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking (value: never, key: never, map: ReadonlyMap<never, never>) => void seen as one taking <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey][] \| PragmaPseudoMap[TKey], key: TKey, map: ReadonlyPragmaMap) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking OrdinalParentheizerRuleSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking ParenthesizerRule<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking ParenthesizerRuleOrSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string seen as one taking [fileName: string] \| [fileName: string, eventKind: FileWatcherEventKind, modifiedTime?: Date \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLink would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (deleteFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (log would lose its object, and this with it) | 1 |
| Refused: a method read as a value (realpath would lose its object, and this with it) | 1 |
| Refused: a method read as a value (remove would lose its object, and this with it) | 1 |
| Refused: a method read as a value (repeat would lose its object, and this with it) | 1 |
| Refused: a method read as a value (replace would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportTruncationError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 |
| Refused: a number as a condition | 1 |
| Refused: a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type BuilderProgram seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what BuilderProgram can't hold | 1 |
| Refused: a value of type BuilderProgram \| Program seen as BuilderProgram, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type CallExpression seen as RequireOrImportCall, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & Identifier would replace | 1 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfListType, which can write "list" \| "listOrElement" where "boolean" is read | 1 |
| Refused: a value of type CommandLineOption \| undefined seen as TsConfigOnlyOption, which can write "object" where "boolean" is read | 1 |
| Refused: a value of type CommandLineOptionOfBooleanType \| CommandLineOptionOfCustomType \| CommandLineOptionOfNumberType \| CommandLineOptionOfStringType \| TsConfigOnlyOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 1 |
| Refused: a value of type CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 |
| Refused: a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type ExpressionWithTypeArguments seen as ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; }, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & (Identifier \| PropertyAccessEntityNameExpression) would replace | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FlowNode seen as FlowLabel, which can write undefined where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as T, a type parameter whose constraint Node can be written, so it can write what Identifier can't hold | 1 |
| Refused: a value of type ImportTypeNode \| undefined seen as ValidImportTypeNode \| undefined, whose readonly field argument becomes writable: a readonly field may hold something narrower than TypeNode, which a write of LiteralTypeNode & { literal: StringLiteral; } would replace | 1 |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as SyntaxList, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.SyntaxList would replace | 1 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what Node can't hold | 1 |
| Refused: a value of type Node seen as TemplateLiteralTypeSpan, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.TemplateLiteralType would replace | 1 |
| Refused: a value of type Node \| NodeArray<Node> seen as NodeArray<Node>, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 |
| Refused: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 |
| Refused: a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved \| undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } \| undefined would replace | 1 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 |
| Refused: a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration[] seen as VariableStatement[], which can write VariableStatement where VariableDeclaration is read | 1 |
| Refused: a value of type VisitEachChildTable seen as Record<SyntaxKind, VisitEachChildFunction<any> \| undefined>, which can write VisitEachChildFunction<any> \| undefined where VisitEachChildFunction<QualifiedName> is read | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 |
| Refused: a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 |
| Refused: a value of type readonly T[] \| undefined seen as unknown[], which can write unknown where T is read | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 |
| Refused: a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 |
| Refused: an import cycle | 1 |
| Refused: arguments | 1 |
| Refused: debugger | 1 |
| Refused: delete | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 |
| Refused: inherited library member prototype read as an own field | 1 |
| Refused: inherited library member replace read as an own field | 1 |
| Refused: instantiating a generic function makes a value of type T \| undefined written where T is read | 1 |
