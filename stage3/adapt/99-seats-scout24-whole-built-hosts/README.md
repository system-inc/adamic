# Whole default watch and build hosts

Adamic ruled this adaptation on Nov 24 2026 (#bbssdcy, the scout pass's proposal 20, casts cause 2; road two for the watch and build hosts, 18:03:08Z, and the copies from `createWatchHost` as construction, 19:54:35Z): each host built whole at construction, JavaScript-changing and value-identical; Adamic accepts the head after its review and census.

The three units are `performBuild`, `createWatchOfConfigFile`, and
`createWatchOfFilesAndCompilerOptions`. Their four CLI sites at
`executeCommandLine.ts:855:27`, `:886:23`, `:1022:31`, and `:1044:31`
use the default builder type and supply no factory. They call four
internal nongeneric twins that flatten the intervening generic builder
layers. Each final host is one literal. The intermediate layers are
represented by the ordered initializers in that literal's constructor.

The generic chain continues to carry the factory cast at
`watch.ts:862:41` and the five downcasts at `watch.ts:873:20`,
`:934:18`, `:965:18`, `tsbuildPublic.ts:297:18`, and `:326:18`.
All other callers keep that chain. The CLI's post-construction writes
stay at their original sites.

The program prelude preserves the order of eager calls and System reads.
The solution prelude evaluates the first binding of `now` at its original
point and the second at the end of the solution initializers. The
literal uses the second value at the original key position. The watch
callback reads the receiver's current `onWatchStatusChange` upon
invocation. The host is stored in a cell declared before the literal.

The two exempt construction copies at `watch.ts:874` and
`tsbuildPublic.ts:328` take a `WatchHost` made at the same point as in
stock. The six explicit reads follow in this order.

| Order | Own member | Presence in createWatchHost literal |
| --- | --- | --- |
| 1 | `onWatchStatusChange` | Always |
| 2 | `watchFile` | Always |
| 3 | `watchDirectory` | Always |
| 4 | `setTimeout` | Always |
| 5 | `clearTimeout` | Always |
| 6 | `preferNonRecursiveWatch` | Always, including undefined |

The six occupy the copy's positions after the program keys, with the
solution-only initializers preceding them in the solution watch variant.
`enumeration.cjs` finds two exempt construction copies across 713 roots
in the `-after` source scanned by `test.cjs`. The raw grep finds 10,756 matches in 406 files. Reads that can be
located in the builder bodies are deferred closures. The row at
`watch.ts:884` is the live receiver read. The built host does not escape
through a callable before return, except as the destination of the
exempt construction copy, which only sets keys. Resolved CLI reach and unresolved dispatch are counted separately.

| Host | Members | Reads | Resolved CLI reach | CLI reach unclassified | During construction |
| --- | ---: | ---: | ---: | ---: | ---: |
| ProgramHost | 28 | 118 | 99 | 16 | 0 |
| WatchCompilerHost | 37 | 196 | 156 | 34 | 0 |
| WatchCompilerHostOfFilesAndCompilerOptions | 41 | 204 | 164 | 34 | 0 |
| WatchCompilerHostOfConfigFile | 44 | 213 | 173 | 34 | 0 |
| SolutionBuilderHostBase | 38 | 131 | 97 | 31 | 0 |
| SolutionBuilderHost | 39 | 134 | 97 | 34 | 0 |
| SolutionBuilderWithWatchHost | 44 | 183 | 130 | 49 | 0 |

The adapter pins exact syntax nodes for the seven stock builders, the
three watch input interfaces, both texts of the `createWatchHost` implementation and
both overloads, `OwnWatchHost`, `copyProperties`, and
the four CLI calls and import specifiers. It compares preserved nodes
before and after planning and after readback. A second call verifies
the existing output without writing it.

Run `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-scout24-whole-built-hosts/test.cjs APPLIED_TREE`.
The test compares the adapter's before and after refs, uses the stock
constructor graph and new literal constructors in a VM, and checks
ordered System access, own presence, property and function values,
memoization, and post-return callback replacement. It checks the
`WatchHost` object itself for inherited enumerable keys and reruns the
typed enumeration scan against the post-adapter output. The scan
resolves library declarations, import aliases, and const aliases with
direct initializers. Dynamic keys, reassigned aliases, callbacks
carrying enumeration functions, and host types erased before a call
remain outside the scan. The pinned grep had only the two typed
construction copies. The drift rows cover a namespace element access,
an alias replacing a pinned import, and a Reflect.ownKeys call. Additional
rows check a const bracket key, an escaped overload alias, and
Object.getOwnPropertyDescriptors on a typed host, each on first apply and replay.
The adapter's reference walk resolves bracket keys by their literal types.
On both trees, the new rows check shorthand destructuring, a typed Pick,
a structural view, a generic member read, a direct destructured call, an
any-typed alias, an unresolved generic bracket key, and a 33rd use of
the ts namespace. The presence row checks Object.getOwnPropertyDescriptor,
Reflect.has, and Reflect.getOwnPropertyDescriptor.
On first apply and replay, the module rows check self-import, nested
namespace assignment, a resolved dynamic import, shorthand capture,
twin assignment, import-equals rebinding, an unresolved dynamic import,
and export surface drift in the pinned module and one level below it,
including an `export * as` line appended one level below.

Road (c) adds a required-reporter internal overload that returns all six
own members while the two stock copies resolve to the general overload.
The checker counts calls by their resolved overload signature and pins
the two stock resolutions and the three twin resolutions. The walk counts
or refuses visible value references to `createWatchHost` by symbol and
resolves shorthand bindings through their pattern. A value whose type
holds a twin or `createWatchHost` is allowed as the object of an access
or at one of tsc's 32 uses of `ts`, pinned by file, enclosing function,
and exact text. A source file's module namespace value passes as an
access object, at one of those 32 sites, or at a matching additional
pin whose recorded export surface still holds. Each additional pin
admits only its named refusal. The walk refuses import-equals
rebinding of such a namespace at top level or within a namespace.
It refuses dynamic imports of source modules except the pinned call in
`importTests` and refuses unresolved dynamic imports. A property name
in an assignment pattern resolves through the assigned value's type,
as a binding pattern's name does.
The implementation declares `OwnWatchHost` with a type-only annotation
under Adamic's amended requirement, "keeps every stock JavaScript byte and every public declaration".
Each watch twin takes the Nov 22 shape with a plain `const system = inputSystem` copy before its system bindings and reads.

> The internal overload's promise is a type contract, not a different run-time road. `createWatchHost`'s body sets all six members on every call, whatever the reporter. So the 32 sites, `runner.ts`'s `(globalThis as any).ts = ts` included, can't make the function behave differently. They could only hand the narrower type to a caller the walk didn't count, and tsc's own sites don't.

- `src/compiler/debug.ts:420:34`, `formatSyntaxKind`, `ts`
- `src/compiler/debug.ts:424:34`, `formatSnippetKind`, `ts`
- `src/compiler/debug.ts:428:34`, `formatScriptKind`, `ts`
- `src/compiler/debug.ts:432:35`, `formatNodeFlags`, `ts`
- `src/compiler/debug.ts:436:35`, `formatNodeCheckFlags`, `ts`
- `src/compiler/debug.ts:440:35`, `formatModifierFlags`, `ts`
- `src/compiler/debug.ts:444:35`, `formatTransformFlags`, `ts`
- `src/compiler/debug.ts:448:35`, `formatEmitFlags`, `ts`
- `src/compiler/debug.ts:452:35`, `formatSymbolFlags`, `ts`
- `src/compiler/debug.ts:456:35`, `formatTypeFlags`, `ts`
- `src/compiler/debug.ts:460:35`, `formatSignatureFlags`, `ts`
- `src/compiler/debug.ts:464:35`, `formatObjectFlags`, `ts`
- `src/compiler/debug.ts:468:35`, `formatFlowFlags`, `ts`
- `src/compiler/debug.ts:472:36`, `formatRelationComparisonResult`, `ts`
- `src/compiler/debug.ts:476:34`, `formatCheckMode`, `ts`
- `src/compiler/debug.ts:480:34`, `formatSignatureCheckMode`, `ts`
- `src/compiler/debug.ts:484:35`, `formatTypeFacts`, `ts`
- `src/compiler/emitter.ts:5791:141`, `makeName`, `ts`
- `src/server/project.ts:1799:265`, `Project.updateGraphWorker`, `ts`
- `src/server/project.ts:2182:68`, `Project.enableProxy`, `ts`
- `src/harness/fourslashImpl.ts:4721:11`, `runCode`, `ts`
- `src/testRunner/unittests/config/matchFiles.ts:175:158`, `baselineMatches`, `ts`
- `src/testRunner/unittests/config/tsconfigParsing.ts:407:129`, `baselineWildcards`, `ts`
- `src/testRunner/unittests/evaluation/esDecorators.ts:1909:57`, top level, `ts`
- `src/testRunner/unittests/evaluation/esDecoratorsMetadata.ts:17:57`, top level, `ts`
- `src/testRunner/unittests/moduleResolution.ts:1058:45`, top level, `ts`
- `src/testRunner/unittests/helpers/baseline.ts:65:52`, `baselineProgram`, `ts`
- `src/testRunner/unittests/reuseProgramStructure.ts:46:68`, `baselineDiagnostics`, `ts`
- `src/testRunner/unittests/reuseProgramStructure.ts:57:64`, `baselineProgram`, `ts`
- `src/testRunner/unittests/skipJSDocParsing.ts:19:69`, `diffSourceFiles`, `ts`
- `src/testRunner/runner.ts:264:5`, `startTestEnvironment`, `(globalThis as any).ts = ts`
- `src/testRunner/runner.ts:264:30`, `startTestEnvironment`, `ts`

The four additional pins record the following sites and surfaces.

- `session-command-names`, `src/server/session.ts:323:35`, top level, `protocol`. Admits module for `server/protocol.ts`, hashing 1 module(s).
- `text-changes-delete-node`, `src/testRunner/unittests/services/textChanges.ts:131:28`, top level, `ts.textChanges`. Admits module for `services/_namespaces/ts.textChanges.ts`, hashing 2 module(s).
- `tsserver-command-names`, `src/testRunner/unittests/tsserver/session.ts:215:83`, top level, `ts.server.protocol`. Admits module for `server/_namespaces/ts.server.protocol.ts`, hashing 2 module(s).
- `runner-import-tests`, `src/testRunner/runner.ts:258:12`, `importTests`, `import("./tests.js")`. Admits import for `testRunner/tests.ts`, hashing 234 module(s).

A surface hashes export declarations and assignments, exported statements,
imports and import-equals declarations, and top-level declarations named
by local export lists. It follows star and namespace re-exports one level.

> None of the four can make `createWatchHost` behave differently or hand a twin or the internal overload to anyone, because their modules' export surfaces, as hashed, hold no twin and no `createWatchHost` at any depth. Seats read them against 6.0.3 (050880ce).

The scope excludes plain `require()` calls and values of TypeScript
`namespace` declarations. Import-equals rebinding of a source module
namespace still refuses.

The `createProgram` wrapper passes all six arguments unchanged and in order, and all three reads of `createProgram` on these hosts only call it, with zero identity comparisons or reads of `length`, `name`, or `toString`, so the members and values agree for everything tsc can observe.

Ordinary prototypes are a premise because a setter on a destination prototype requires a prototype installed by the program, which Adamic refuses, so no admitted program can tell the `copyProperties` assignment from the literal's own keys.

`Function.name` of the two `afterProgramCreate` arrows and the `onUnRecoverableConfigFileDiagnostic` arrow is `""` in stock and the property name in the twins, and `src` has zero reads of those members' names.

The twins make the default watch reporter one call earlier than `createWatchHost` would, with nothing evaluated between, as the creation log in `test.cjs` shows on both sides.

Rulings 12, 14 and 39 govern preserving the casts in existing generic
paths and calling the default factory directly from the twins. Row 290
requires construction with all keys present. Rows 406 and 408 govern
public shape and the stripped declaration check.

## Pinned resolved call inventory

| Owner | Position | Resolution | Factory argument | Disposition |
| --- | --- | --- | --- | --- |
| createWatchCompilerHost | src/compiler/watch.ts:873:20 | `<T>(system: System, createProgram: CreateProgram<T> \| undefined): ProgramHost<T>` (T) | `createProgram: CreateProgram<T> \| undefined` | Preserved. |
| createWatchCompilerHostOfConfigFile | src/compiler/watch.ts:934:18 | `<T>(system: System \| undefined, createProgram: CreateProgram<T> \| undefined, reportDiagnostic: DiagnosticReporter, reportWatchStatus?: WatchStatusReporter): WatchCompilerHost<T>` (T) | `createProgram: CreateProgram<T> \| undefined` | Preserved. |
| createWatchCompilerHostOfFilesAndCompilerOptions | src/compiler/watch.ts:965:18 | `<T>(system: System \| undefined, createProgram: CreateProgram<T> \| undefined, reportDiagnostic: DiagnosticReporter, reportWatchStatus?: WatchStatusReporter): WatchCompilerHost<T>` (T) | `createProgram: CreateProgram<T> \| undefined` | Preserved. |
| createWatchCompilerHost | src/compiler/watchPublic.ts:371:16 | `<T>({ rootFiles, options, watchOptions, projectReferences, system, createProgram, reportDiagnostic, reportWatchStatus, }: import("./watch.js").CreateWatchCompilerHostOfFilesAndCompilerOptionsInput<T>): WatchCompilerHostOfFilesAndCompilerOptions<T>` (T) | `createProgram` | Preserved. |
| createWatchCompilerHost | src/compiler/watchPublic.ts:383:16 | `<T>({ configFileName, optionsToExtend, watchOptionsToExtend, extraFileExtensions, system, createProgram, reportDiagnostic, reportWatchStatus, }: import("./watch.js").CreateWatchCompilerHostOfConfigFileInput<T>): WatchCompilerHostOfConfigFile<T>` (T) | `createProgram` | Preserved. |
| createSolutionBuilderHostBase | src/compiler/tsbuildPublic.ts:297:18 | `<T>(system: System, createProgram: CreateProgram<T> \| undefined): ProgramHost<T>` (T) | `createProgram: CreateProgram<T> \| undefined` | Preserved. |
| createSolutionBuilderHost | src/compiler/tsbuildPublic.ts:314:18 | `<T>(system: System, createProgram: CreateProgram<T> \| undefined, reportDiagnostic?: DiagnosticReporter, reportSolutionBuilderStatus?: DiagnosticReporter): SolutionBuilderHostBase<T>` (T) | `createProgram: CreateProgram<T> \| undefined` | Preserved. |
| createSolutionBuilderWithWatchHost | src/compiler/tsbuildPublic.ts:326:18 | `<T>(system: System, createProgram: CreateProgram<T> \| undefined, reportDiagnostic?: DiagnosticReporter, reportSolutionBuilderStatus?: DiagnosticReporter): SolutionBuilderHostBase<T>` (T) | `createProgram: CreateProgram<T> \| undefined` | Preserved. |
| executeCommandLineWorker | src/compiler/executeCommandLine.ts:676:20 | `(system: System, cb: ExecuteCommandLineCallbacks, reportDiagnostic: DiagnosticReporter, configParseResult: ParsedCommandLine, optionsToExtend: CompilerOptions, watchOptionsToExtend: WatchOptions \| undefined, extendedConfigCache: Map<string, ExtendedConfigCacheEntry>): WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>` (nongeneric) | `no factory parameter` | Preserved. |
| executeCommandLineWorker | src/compiler/executeCommandLine.ts:687:13 | `(sys: System, cb: ExecuteCommandLineCallbacks, reportDiagnostic: DiagnosticReporter, config: ParsedCommandLine): number \| undefined` (nongeneric) | `sys: System has no declared factory property` | Preserved. |
| executeCommandLineWorker | src/compiler/executeCommandLine.ts:716:20 | `(system: System, cb: ExecuteCommandLineCallbacks, reportDiagnostic: DiagnosticReporter, rootFiles: string[], options: CompilerOptions, watchOptions: WatchOptions \| undefined): import("./watchPublic.js").WatchOfFilesAndCompilerOptions<EmitAndSemanticDiagnosticsBuilderProgram>` (nongeneric) | `no factory parameter` | Preserved. |
| executeCommandLineWorker | src/compiler/executeCommandLine.ts:726:13 | `(sys: System, cb: ExecuteCommandLineCallbacks, reportDiagnostic: DiagnosticReporter, config: ParsedCommandLine): number \| undefined` (nongeneric) | `sys: System has no declared factory property` | Preserved. |
| executeCommandLine | src/compiler/executeCommandLine.ts:765:17 | `(sys: System, cb: ExecuteCommandLineCallbacks, buildOptions: BuildOptions, watchOptions: WatchOptions \| undefined, projects: string[], errors: Diagnostic[]): number \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| undefined` (nongeneric) | `no factory parameter` | Preserved. |
| executeCommandLine | src/compiler/executeCommandLine.ts:775:20 | `(sys: System, cb: ExecuteCommandLineCallbacks, buildOptions: BuildOptions, watchOptions: WatchOptions \| undefined, projects: string[], errors: Diagnostic[]): number \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| undefined` (nongeneric) | `no factory parameter` | Preserved. |
| performBuild | src/compiler/executeCommandLine.ts:855:27 | `<EmitAndSemanticDiagnosticsBuilderProgram>(system?: System, createProgram?: CreateProgram<EmitAndSemanticDiagnosticsBuilderProgram> \| undefined, reportDiagnostic?: DiagnosticReporter, reportSolutionBuilderStatus?: DiagnosticReporter, reportWatchStatus?: import("./watchPublic.js").WatchStatusReporter): import("./tsbuildPublic.js").SolutionBuilderWithWatchHost<EmitAndSemanticDiagnosticsBuilderProgram>` (EmitAndSemanticDiagnosticsBuilderProgram) | `undefined: undefined` | Redirected to the default twin. |
| performBuild | src/compiler/executeCommandLine.ts:886:23 | `<EmitAndSemanticDiagnosticsBuilderProgram>(system?: System, createProgram?: CreateProgram<EmitAndSemanticDiagnosticsBuilderProgram> \| undefined, reportDiagnostic?: DiagnosticReporter, reportSolutionBuilderStatus?: DiagnosticReporter, reportErrorSummary?: ReportEmitErrorSummary): import("./tsbuildPublic.js").SolutionBuilderHost<EmitAndSemanticDiagnosticsBuilderProgram>` (EmitAndSemanticDiagnosticsBuilderProgram) | `undefined: undefined` | Redirected to the default twin. |
| createWatchOfConfigFile | src/compiler/executeCommandLine.ts:1022:31 | `<EmitAndSemanticDiagnosticsBuilderProgram>({ configFileName, optionsToExtend, watchOptionsToExtend, extraFileExtensions, system, createProgram, reportDiagnostic, reportWatchStatus, }: import("./watch.js").CreateWatchCompilerHostOfConfigFileInput<EmitAndSemanticDiagnosticsBuilderProgram>): import("./watchPublic.js").WatchCompilerHostOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>` (EmitAndSemanticDiagnosticsBuilderProgram) | `no factory property in object literal` | Redirected to the default twin. |
| createWatchOfFilesAndCompilerOptions | src/compiler/executeCommandLine.ts:1044:31 | `<EmitAndSemanticDiagnosticsBuilderProgram>({ rootFiles, options, watchOptions, projectReferences, system, createProgram, reportDiagnostic, reportWatchStatus, }: import("./watch.js").CreateWatchCompilerHostOfFilesAndCompilerOptionsInput<EmitAndSemanticDiagnosticsBuilderProgram>): import("./watchPublic.js").WatchCompilerHostOfFilesAndCompilerOptions<EmitAndSemanticDiagnosticsBuilderProgram>` (EmitAndSemanticDiagnosticsBuilderProgram) | `no factory property in object literal` | Redirected to the default twin. |
| createSolutionBuilderHostForBaseline | src/testRunner/unittests/helpers/solutionBuilder.ts:27:18 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>(system?: ts.System, createProgram?: ts.CreateProgram<ts.EmitAndSemanticDiagnosticsBuilderProgram> \| undefined, reportDiagnostic?: ts.DiagnosticReporter, reportSolutionBuilderStatus?: ts.DiagnosticReporter, reportErrorSummary?: ts.ReportEmitErrorSummary): ts.SolutionBuilderHost<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `undefined: undefined` | Preserved. |
| createSolutionBuilderWithWatchHostForBaseline | src/testRunner/unittests/helpers/tscWatch.ts:84:18 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>(system?: ts.System, createProgram?: ts.CreateProgram<ts.EmitAndSemanticDiagnosticsBuilderProgram> \| undefined, reportDiagnostic?: ts.DiagnosticReporter, reportSolutionBuilderStatus?: ts.DiagnosticReporter, reportWatchStatus?: ts.WatchStatusReporter): ts.SolutionBuilderWithWatchHost<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `undefined: undefined` | Preserved. |
| createWatchCompilerHostOfConfigFileForBaseline | src/testRunner/unittests/helpers/tscWatch.ts:97:18 | `<T>({ configFileName, optionsToExtend, watchOptionsToExtend, extraFileExtensions, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfConfigFileInput<T>): ts.WatchCompilerHostOfConfigFile<T>` (T) | `{ ...input, reportDiagnostic: ts.createDiagnosticReporter(input.system, /*pretty*/ true), reportWatchStatus: ts.createWatchStatusReporter(input.system, /*pretty*/ true), }.createProgram: ts.CreateProgram<T> \| undefined` | Preserved. |
| createWatchCompilerHostOfFilesAndCompilerOptionsForBaseline | src/testRunner/unittests/helpers/tscWatch.ts:113:18 | `<T>({ rootFiles, options, watchOptions, projectReferences, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfFilesAndCompilerOptionsInput<T>): ts.WatchCompilerHostOfFilesAndCompilerOptions<T>` (T) | `{ ...input, reportDiagnostic: ts.createDiagnosticReporter(input.system, /*pretty*/ true), reportWatchStatus: ts.createWatchStatusReporter(input.system, /*pretty*/ true), }.createProgram: ts.CreateProgram<T> \| undefined` | Preserved. |
| verifyProgramWithoutConfigFile | src/testRunner/unittests/reuseProgramStructure.ts:846:51 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>({ rootFiles, options, watchOptions, projectReferences, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfFilesAndCompilerOptionsInput<ts.EmitAndSemanticDiagnosticsBuilderProgram>): ts.WatchCompilerHostOfFilesAndCompilerOptions<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `no factory property in object literal` | Preserved. |
| verifyProgramWithConfigFile | src/testRunner/unittests/reuseProgramStructure.ts:856:51 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>({ configFileName, optionsToExtend, watchOptionsToExtend, extraFileExtensions, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfConfigFileInput<ts.EmitAndSemanticDiagnosticsBuilderProgram>): ts.WatchCompilerHostOfConfigFile<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `no factory property in object literal` | Preserved. |
| <top level or anonymous> | src/testRunner/unittests/reuseProgramStructure.ts:1010:51 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>({ rootFiles, options, watchOptions, projectReferences, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfFilesAndCompilerOptionsInput<ts.EmitAndSemanticDiagnosticsBuilderProgram>): ts.WatchCompilerHostOfFilesAndCompilerOptions<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `no factory property in object literal` | Preserved. |
| <top level or anonymous> | src/testRunner/unittests/reuseProgramStructure.ts:1048:51 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>({ rootFiles, options, watchOptions, projectReferences, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfFilesAndCompilerOptionsInput<ts.EmitAndSemanticDiagnosticsBuilderProgram>): ts.WatchCompilerHostOfFilesAndCompilerOptions<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `no factory property in object literal` | Preserved. |
| <top level or anonymous> | src/testRunner/unittests/reuseProgramStructure.ts:1076:51 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>({ rootFiles, options, watchOptions, projectReferences, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfFilesAndCompilerOptionsInput<ts.EmitAndSemanticDiagnosticsBuilderProgram>): ts.WatchCompilerHostOfFilesAndCompilerOptions<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `no factory property in object literal` | Preserved. |
| <top level or anonymous> | src/testRunner/unittests/tsbuild/graphOrdering.ts:23:16 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>(system?: ts.System, createProgram?: ts.CreateProgram<ts.EmitAndSemanticDiagnosticsBuilderProgram> \| undefined, reportDiagnostic?: ts.DiagnosticReporter, reportSolutionBuilderStatus?: ts.DiagnosticReporter, reportErrorSummary?: ts.ReportEmitErrorSummary): ts.SolutionBuilderHost<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `factory argument absent` | Preserved. |
| <top level or anonymous> | src/testRunner/unittests/tsbuild/publicApi.ts:43:31 | `<ts.EmitAndSemanticDiagnosticsBuilderProgram>(system?: ts.System, createProgram?: ts.CreateProgram<ts.EmitAndSemanticDiagnosticsBuilderProgram> \| undefined, reportDiagnostic?: ts.DiagnosticReporter, reportSolutionBuilderStatus?: ts.DiagnosticReporter, reportErrorSummary?: ts.ReportEmitErrorSummary): ts.SolutionBuilderHost<ts.EmitAndSemanticDiagnosticsBuilderProgram>` (ts.EmitAndSemanticDiagnosticsBuilderProgram) | `undefined: undefined` | Preserved. |
| <top level or anonymous> | src/testRunner/unittests/tsbuild/sample.ts:225:30 | `<ts.BuilderProgram>(system?: ts.System, createProgram?: ts.CreateProgram<ts.BuilderProgram> \| undefined, reportDiagnostic?: ts.DiagnosticReporter, reportSolutionBuilderStatus?: ts.DiagnosticReporter, reportErrorSummary?: ts.ReportEmitErrorSummary): ts.SolutionBuilderHost<ts.BuilderProgram>` (ts.BuilderProgram) | `ts.createAbstractBuilder: { (newProgram: ts.Program, host: ts.BuilderProgramHost, oldProgram?: ts.BuilderProgram, configFileParsingDiagnostics?: readonly ts.Diagnostic[]): ts.BuilderProgram; (rootNames: readonly string[] \| undefined, options: ts.CompilerOptions \| undefined, host?: ts.CompilerHost, oldProgram?: ts.BuilderProgram, configFileParsingDiagnostics?: readonly ts.Diagnostic[], projectReferences?: readonly ts.ProjectReference[]): ts.BuilderProgram; }` | Preserved. |
| verifyCancellation | src/testRunner/unittests/tsc/cancellationToken.ts:93:13 | `(): void` (nongeneric) | `argument absent` | Preserved. |
| baselineBuild | src/testRunner/unittests/tsc/cancellationToken.ts:162:17 | `(): void` (nongeneric) | `argument absent` | Preserved. |
| verify | src/testRunner/unittests/tscWatch/watchApi.ts:472:26 | `<ts.SemanticDiagnosticsBuilderProgram>({ configFileName, optionsToExtend, watchOptionsToExtend, extraFileExtensions, system, createProgram, reportDiagnostic, reportWatchStatus, }: ts.CreateWatchCompilerHostOfConfigFileInput<ts.SemanticDiagnosticsBuilderProgram>): ts.WatchCompilerHostOfConfigFile<ts.SemanticDiagnosticsBuilderProgram>` (ts.SemanticDiagnosticsBuilderProgram) | `createProgram: ts.createSemanticDiagnosticsBuilderProgram` | Preserved. |

### Round 10 createWatchHost resolutions

| Owner | Position | Resolution | Reporter argument | Disposition |
| --- | --- | --- | --- | --- |
| createWatchCompilerHost | src/compiler/watch.ts:890:28 | `(system?: System, reportWatchStatus?: WatchStatusReporter \| undefined): WatchHost` | optional reporter | General overload, stock call preserved. |
| createDefaultWatchCompilerHostOfConfigFile | src/compiler/watch.ts:1016:21 | `(system: System, reportWatchStatus: WatchStatusReporter): OwnWatchHost` | explicit or default reporter | Required overload. |
| createDefaultWatchCompilerHostOfFilesAndCompilerOptions | src/compiler/watch.ts:1104:21 | `(system: System, reportWatchStatus: WatchStatusReporter): OwnWatchHost` | explicit or default reporter | Required overload. |
| createSolutionBuilderWithWatchHost | src/compiler/tsbuildPublic.ts:332:23 | `(system?: System, reportWatchStatus?: WatchStatusReporter \| undefined): WatchHost` | optional reporter | General overload, stock call preserved. |
| createDefaultSolutionBuilderWithWatchHost | src/compiler/tsbuildPublic.ts:406:21 | `(system: System, reportWatchStatus: WatchStatusReporter): OwnWatchHost` | explicit or default reporter | Required overload. |
