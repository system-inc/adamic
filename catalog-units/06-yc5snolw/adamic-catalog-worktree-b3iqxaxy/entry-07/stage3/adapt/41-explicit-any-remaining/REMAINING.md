# Unfinished any contracts

124 explicit tokens remain. The latent census still observes 36 direct any blockers. This list is an incomplete adaptation handoff, not a claim that these contracts are impossible. The groups reuse adaptation 40’s historical analysis; current locations and source snippets are independently parsed in evidence/remaining-sites.json. Each remaining owner still needs caller, type and behavior proof.

## Remaining groups

| Tokens | Historical class | Proposed contract |
| ---: | --- | --- |
| 38 | json | Unknown plus checked narrowing at genuinely untyped JSON/API boundaries; propagate validated recursive JSON/config value types through the internal pipeline. |
| 16 | timer | Generic parameters for the host timer handle and argument tuple, carried together through registration, storage, callback and cancellation declarations. |
| 11 | constructors | Decline until constructor return and staged initialization contracts are modeled at their owners. |
| 10 | copy | Decline until enumerable copy results and writable key relationships have an owner contract. |
| 6 | ast-projection | Real type: the precise node union or manufactured node shape at the owning declaration. |
| 5 | builder-factory | Decline until the generic program type is tied to the selected factory and persisted builder result at the owner. |
| 5 | display-projection | Real type: declare the actual optional display fields or recursive option display value at the object owner, preserving existing control flow. |
| 4 | sorted-proof | Decline until sortedness proof and its branded result are expressible at the sorting/empty-array owner. |
| 3 | dependent-dispatch | Generic parameter associating the node kind and concrete node type with the callback input/output at registration and dispatch. |
| 3 | generic-default | Generic parameter: retain the argument tuple and checked input/output relationship explicitly, removing permissive defaults after auditing explicit generic callers. |
| 3 | host-extension | Unknown plus checked narrowing only for genuinely untyped host extensions at process.browser, process.recordreplay and stdout._handle. |
| 3 | instrumentation | Generic parameter for before/after instrumentation data, with void for an ignored after-hook result. |
| 2 | data-comparison | Real type: recursive compiler data domain including object/array, scalar, null/undefined and ignored function members. |
| 2 | debug-mutation | Decline: the restore assignment stores a cache entry object where the declared slot is callable; a truthful callable replacement exposes a runtime contract mismatch. |
| 2 | host-standard | Real type: existing Array, String and Error APIs, with optional host availability declared at their owners where older runtimes omit the API. |
| 2 | program-feature | Real type: an optional getSemanticDiagnosticsOfNextAffectedFile capability on the program owner, preserving the existing presence test. |
| 2 | watch-callback | Generic parameter carrying the watcher callback argument tuple through the cache and logging wrapper. |
| 1 | comparer | Decline until the string default comparer is restricted to a string element branch or a generic caller supplies its comparator. |
| 1 | config-host | Real type: the actual ParseConfigFileHost capability subset at the host declaration, proving required optional System methods before construction. |
| 1 | distribution | Generic parameter: use the identity condition U extends U to preserve distribution without a top-type escape. |
| 1 | generator-input | Generic parameter carrying the delegated iterator next-input type through flatMapIterator. |
| 1 | module-key | Real type: the actual ModuleResolutionKind string-key domain at the declaration producing moduleKindName. |
| 1 | node-family | Real type: the variable-like declaration union including JSDoc property-like tags required by its callers. |
| 1 | resolution-write | Real type: declare the mutable resolvedModule result field at its actual owner. |

## Current file and line ledger

### json

Config parsing and option validation, package.json, sourcemaps and JSON.parse. Boundary sites are 9, 10, 19, 29, 30, 98, 119 and 180: config/read APIs, public config converters, parsed package data, sourcemap guard and JSON.parse result. Other sites carry these values, construct JSON-shaped values, or validate them; they are not independent untyped host inputs. Recovery can produce undefined. parseOwnConfigOfJson and raw config merging read/write fields without proving the input object and each field shape. Existing truthiness and hasProperty checks do not establish those contracts. Whole-class unknown substitution would require runtime checks or expose consumer errors; declaring every value JSON would also falsely restrict public API inputs before validation.

- `src/compiler/commandLineParser.ts:2267:110`: `export function readConfigFile(fileName: string, readFile: (path: string) => string | undefined): { config?: any; error?: Diagnostic | undefined; } {`
- `src/compiler/commandLineParser.ts:2277:91`: `export function parseConfigFileTextToJson(fileName: string, jsonText: string): { config?: any; error?: Diagnostic | undefined; } {`
- `src/compiler/commandLineParser.ts:2430:16`: `value: any,`
- `src/compiler/commandLineParser.ts:2441:4`: `): any {`
- `src/compiler/commandLineParser.ts:2467:84`: `export function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): any {`
- `src/compiler/commandLineParser.ts:2484:4`: `): any {`
- `src/compiler/commandLineParser.ts:2494:8`: `): any {`
- `src/compiler/commandLineParser.ts:2495:23`: `const result: any = returnValue ? {} : undefined;`
- `src/compiler/commandLineParser.ts:2538:110`: `function convertPropertyValueToJson(valueExpression: Expression, option: CommandLineOption | undefined): any {`
- `src/compiler/commandLineParser.ts:2606:79`: `function isCompilerOptionsValue(option: CommandLineOption | undefined, value: any): value is CompilerOptionsValue {`
- `src/compiler/commandLineParser.ts:3013:50`: `export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine {`
- `src/compiler/commandLineParser.ts:3038:31`: `function isNullOrUndefined(x: any): x is null | undefined { // eslint-disable-line no-restricted-syntax`
- `src/compiler/commandLineParser.ts:3060:11`: `json: any,`
- `src/compiler/commandLineParser.ts:3300:45`: `function startsWithConfigDirTemplate(value: any): value is string {`
- `src/compiler/commandLineParser.ts:3354:48`: `export function canJsonReportNoInputFiles(raw: any): boolean {`
- `src/compiler/commandLineParser.ts:3377:10`: `raw: any;`
- `src/compiler/commandLineParser.ts:3406:11`: `json: any,`
- `src/compiler/commandLineParser.ts:3497:11`: `json: any,`
- `src/compiler/commandLineParser.ts:3599:16`: `value: any,`
- `src/compiler/commandLineParser.ts:3725:57`: `function convertCompileOnSaveOptionFromJson(jsonOption: any, basePath: string, errors: Diagnostic[]): boolean {`
- `src/compiler/commandLineParser.ts:3733:61`: `export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; } {`
- `src/compiler/commandLineParser.ts:3739:61`: `export function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: TypeAcquisition; errors: Diagnostic[]; } {`
- `src/compiler/commandLineParser.ts:3752:60`: `function convertCompilerOptionsFromJsonWorker(jsonOptions: any, basePath: string, errors: Diagnostic[], configFileName?: string): CompilerOptions {`
- `src/compiler/commandLineParser.ts:3765:60`: `function convertTypeAcquisitionFromJsonWorker(jsonOptions: any, basePath: string, errors: Diagnostic[], configFileName?: string): TypeAcquisition {`
- `src/compiler/commandLineParser.ts:3771:57`: `function convertWatchOptionsFromJsonWorker(jsonOptions: any, basePath: string, errors: Diagnostic[]): WatchOptions | undefined {`
- `src/compiler/commandLineParser.ts:3775:94`: `function convertOptionsFromJson(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: any, basePath: string, defaultOptions: undefined, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]): WatchOptions | undefined;`
- `src/compiler/commandLineParser.ts:3776:94`: `function convertOptionsFromJson(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: any, basePath: string, defaultOptions: CompilerOptions | TypeAcquisition, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]): CompilerOptions | TypeAcquisition;`
- `src/compiler/commandLineParser.ts:3777:94`: `function convertOptionsFromJson(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: any, basePath: string, defaultOptions: CompilerOptions | TypeAcquisition | WatchOptions | undefined, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]) {`
- `src/compiler/commandLineParser.ts:3803:12`: `value: any,`
- `src/compiler/commandLineParser.ts:3835:90`: `function normalizeNonListOptionValue(option: CommandLineOption, basePath: string, value: any): CompilerOptionsValue {`
- `src/compiler/commandLineParser.ts:3880:22`: `values: readonly any[],`
- `src/compiler/commandLineParser.ts:3886:4`: `): any[] {`
- `src/compiler/commandLineParser.ts:4276:48`: `function getOptionValueWithEmptyStrings(value: any, option: CommandLineOption): {} | undefined {`
- `src/compiler/moduleSpecifiers.ts:1256:54`: `const packageJsonContent: Record<string, any> | undefined = cachedPackageJson?.contents.packageJsonContent || tryParseJson(host.readFile!(packageJsonPath)!);`
- `src/compiler/sourcemap.ts:404:28`: `function isStringOrNull(x: any) {`
- `src/compiler/sourcemap.ts:408:28`: `function isRawSourceMap(x: any): x is RawSourceMap {`
- `src/compiler/types.ts:7716:11`: `raw?: any;`
- `src/compiler/utilities.ts:7809:45`: `export function tryParseJson(text: string): any {`
### timer

Includes five callback/rest argument tokens and eleven opaque timer handle tokens. System and WatchHost timer contracts are host-polymorphic; forwarding an opaque handle does not justify unknown plus invented runtime narrowing. A blanket Node timer type would exclude other supported hosts.

- `src/compiler/sys.ts:51:48`: `declare function setTimeout(handler: (...args: any[]) => void, timeout: number): any;`
- `src/compiler/sys.ts:51:82`: `declare function setTimeout(handler: (...args: any[]) => void, timeout: number): any;`
- `src/compiler/sys.ts:52:39`: `declare function clearTimeout(handle: any): void;`
- `src/compiler/sys.ts:464:24`: `let pollScheduled: any;`
- `src/compiler/sys.ts:618:36`: `let timerToUpdateChildWatches: any;`
- `src/compiler/sys.ts:1441:37`: `setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;`
- `src/compiler/sys.ts:1441:74`: `setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;`
- `src/compiler/sys.ts:1441:82`: `setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;`
- `src/compiler/sys.ts:1442:30`: `clearTimeout?(timeoutId: any): void;`
- `src/compiler/tsbuildPublic.ts:424:37`: `timerToBuildInvalidatedProject: any;`
- `src/compiler/watchPublic.ts:169:37`: `setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;`
- `src/compiler/watchPublic.ts:169:74`: `setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;`
- `src/compiler/watchPublic.ts:169:82`: `setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;`
- `src/compiler/watchPublic.ts:171:30`: `clearTimeout?(timeoutId: any): void;`
- `src/compiler/watchPublic.ts:443:31`: `let timerToUpdateProgram: any; // timer callback to recompile the program`
- `src/compiler/watchPublic.ts:444:51`: `let timerToInvalidateFailedLookupResolutions: any; // timer callback to invalidate resolutions for changes in failed lookup locations`
### constructors

NodeLinks construction, Symbol.links initialization and nine objectAllocator constructor casts claim richer interfaces than the bare constructor initially supplies. A cast to the final interface would hide the same missing proof.

- `src/compiler/checker.ts:2940:76`: `return nodeLinks[nodeId] || (nodeLinks[nodeId] = new (NodeLinks as any)());`
- `src/compiler/utilities.ts:8486:14`: `(this as any).links = undefined; // used by TransientSymbol`
- `src/compiler/utilities.ts:8553:39`: `getNodeConstructor: () => Node as any,`
- `src/compiler/utilities.ts:8554:41`: `getTokenConstructor: () => Token as any,`
- `src/compiler/utilities.ts:8555:51`: `getIdentifierConstructor: () => Identifier as any,`
- `src/compiler/utilities.ts:8556:52`: `getPrivateIdentifierConstructor: () => Node as any,`
- `src/compiler/utilities.ts:8557:45`: `getSourceFileConstructor: () => Node as any,`
- `src/compiler/utilities.ts:8558:43`: `getSymbolConstructor: () => Symbol as any,`
- `src/compiler/utilities.ts:8559:39`: `getTypeConstructor: () => Type as any,`
- `src/compiler/utilities.ts:8560:49`: `getSignatureConstructor: () => Signature as any,`
- `src/compiler/utilities.ts:8561:61`: `getSourceMapSourceConstructor: () => SourceMapSource as any,`
### copy

clone, extend, copyProperties and SourceFile metadata copying enumerate string keys and populate initially empty objects. T or T1 & T2 does not prove that copying enumerable own strings preserves prototypes, symbol keys or all members of arbitrary input types.

- `src/compiler/core.ts:1479:19`: `const result: any = {};`
- `src/compiler/core.ts:1482:37`: `result[id] = (object as any)[id];`
- `src/compiler/core.ts:1496:35`: `const result: T1 & T2 = {} as any;`
- `src/compiler/core.ts:1499:24`: `(result as any)[id] = (second as any)[id];`
- `src/compiler/core.ts:1499:46`: `(result as any)[id] = (second as any)[id];`
- `src/compiler/core.ts:1505:24`: `(result as any)[id] = (first as any)[id];`
- `src/compiler/core.ts:1505:45`: `(result as any)[id] = (first as any)[id];`
- `src/compiler/core.ts:1516:23`: `(first as any)[id] = second[id];`
- `src/compiler/factory/nodeFactory.ts:6145:22`: `(node as any)[p] = (source as any)[p];`
- `src/compiler/factory/nodeFactory.ts:6145:43`: `(node as any)[p] = (source as any)[p];`
### ast-projection

Fresh parentName.isTypeOf mutation; parent access on JSDoc parsing elements; optional declaration.type access in symbolWalker. Existing union narrowing and actual factory return shapes must be reconciled at owners.

- `src/compiler/checker.ts:6834:40`: `(parentName as any).isTypeOf = true; // mutably update, node is freshly manufactured anyhow`
- `src/compiler/parser.ts:10207:25`: `if ((element as any).parent) {`
- `src/compiler/parser.ts:10208:40`: `const parent = (element as any).parent as Node;`
- `src/compiler/symbolWalker.ts:207:27`: `if ((d as any).type && (d as any).type.kind === SyntaxKind.TypeQuery) {`
- `src/compiler/symbolWalker.ts:207:46`: `if ((d as any).type && (d as any).type.kind === SyntaxKind.TypeQuery) {`
- `src/compiler/symbolWalker.ts:208:41`: `const query = (d as any).type as TypeQueryNode;`
### builder-factory

Default emit-and-semantic factory and readBuilderProgram results are cast to arbitrary caller T. T can denote a different builder family; a replacement assertion T would perpetuate the mismatch.

- `src/compiler/tsbuildPublic.ts:1355:64`: `return readBuilderProgram(parsed.options, compilerHost) as any as T;`
- `src/compiler/watch.ts:862:91`: `createProgram: createProgram || createEmitAndSemanticDiagnosticsBuilderProgram as any as CreateProgram<T>,`
- `src/compiler/watchPublic.ts:150:88`: `createProgram = createProgram || createEmitAndSemanticDiagnosticsBuilderProgram as any as CreateProgram<T>;`
- `src/compiler/watchPublic.ts:151:61`: `const oldProgram = readBuilderProgram(options, host) as any as T;`
- `src/compiler/watchPublic.ts:555:75`: `builderProgram = readBuilderProgram(compilerOptions, compilerHost) as any as T;`
### display-projection

__debugFlags, function.name, nested compiler option values, Type.objectFlags/intrinsicName and SourceFile.path. The transformer cast follows a SourceFile kind test. Optional Type fields differ by subtype; internal values are not host input warranting unknown.

- `src/compiler/checker.ts:10953:77`: ``return Debug.fail(`Unhandled class member kind! ${(p as any).__debugFlags || p.flags}`);``
- `src/compiler/moduleNameResolver.ts:954:69`: ``str += `${key}: ${compilerOptionValueToString((value as any)[key])}`;``
- `src/compiler/tracing.ts:232:42`: `const objectFlags = (type as any).objectFlags;`
- `src/compiler/tracing.ts:318:41`: `intrinsicName: (type as any).intrinsicName,`
- `src/compiler/transformer.ts:338:116`: `tracing?.push(tracing.Phase.Emit, "transformNodes", node.kind === SyntaxKind.SourceFile ? { path: (node as any as SourceFile).path } : { kind: node.kind, pos: node.pos, end: node.end });`
### sorted-proof

Empty arrays are vacuously sorted; deduplication preserves its sorted input. as any as currently bridges the nominal SortedArray/SortedReadonlyArray marker. Merely swapping the bridge assertion for the target type would not constitute the owner proof requested.

- `src/compiler/core.ts:736:50`: `if (array.length === 0) return emptyArray as any as SortedReadonlyArray<T>;`
- `src/compiler/core.ts:759:28`: `return deduplicated as any as SortedReadonlyArray<T>;`
- `src/compiler/core.ts:764:18`: `return [] as any as SortedArray<T>; // TODO: GH#19873`
- `src/compiler/program.ts:3284:130`: `return rootNames.length ? sortAndDeduplicateDiagnostics(getTypeChecker().getGlobalDiagnostics().slice()) : emptyArray as any as SortedReadonlyArray<Diagnostic>;`
### dependent-dispatch

ParenthesizerRule, forEachChildTable and visitEachChildTable erase heterogeneous function types. A broad Node callback loses contravariant input relationships; the stored callback and its invocation must share a type parameter.

- `src/compiler/emitter.ts:1274:53`: `var currentParenthesizerRule: ParenthesizerRule<any> | undefined;`
- `src/compiler/parser.ts:1253:78`: `const fn = (forEachChildTable as Record<SyntaxKind, ForEachChildFunction<any>>)[node.kind];`
- `src/compiler/visitorPublic.ts:603:82`: `const fn = (visitEachChildTable as Record<SyntaxKind, VisitEachChildFunction<any> | undefined>)[node.kind];`
### generic-default

maybeBind already carries A; cast and tryCast default TIn to any. Replacing the default with an unrelated universal type can break predicate variance or weaken the TOut extends TIn proof.

- `src/compiler/core.ts:1522:40`: `export function maybeBind<T, A extends any[], R>(obj: T, fn: ((this: T, ...args: A) => R) | undefined): ((...args: A) => R) | undefined {`
- `src/compiler/core.ts:1778:49`: `export function tryCast<TOut extends TIn, TIn = any>(value: TIn | undefined, test: (value: TIn) => value is TOut): TOut | undefined {`
- `src/compiler/core.ts:1783:46`: `export function cast<TOut extends TIn, TIn = any>(value: TIn | undefined, test: (value: TIn) => value is TOut): TOut {`
### host-extension

These three sites inspect external runtime extensions: browser marker, replay marker and private stream handle with optional setBlocking. Current truthiness or optional chaining does not prove an arbitrary handle method callable. Changing to checked unknown cannot be presumed byte-identical.

- `src/compiler/core.ts:2593:25`: `&& !(process as any).browser`
- `src/compiler/sys.ts:1595:203`: `debugMode: !!process.env.NODE_INSPECTOR_IPC || !!process.env.VSCODE_INSPECTOR_OPTIONS || some(process.execArgv, arg => /^--(?:inspect|debug)(?:-brk)?(?:=\d+)?$/i.test(arg)) || !!(process as any).recordreplay,`
- `src/compiler/sys.ts:1610:51`: `const handle = (process.stdout as any)?._handle as { setBlocking?: (value: boolean) => void; };`
### instrumentation

ResolutionCache host hooks return a before payload and pass it unchanged into the matching after hook; after return value is ignored. This is an opaque correlated payload, not inspector or perfLogger code.

- `src/compiler/resolutionCache.ts:209:8`: `): any;`
- `src/compiler/resolutionCache.ts:215:15`: `data: any,`
- `src/compiler/resolutionCache.ts:216:8`: `): any;`
### data-comparison

compareDataObjects recursively compares compiler option data and skips functions. It is not restricted to JSON, and unknown is not justified by an untyped external boundary here. Object indexing and recursive null behavior need an owner contract.

- `src/compiler/utilities.ts:8147:41`: `export function compareDataObjects(dst: any, src: any): boolean {`
- `src/compiler/utilities.ts:8147:51`: `export function compareDataObjects(dst: any, src: any): boolean {`
### debug-mutation

setAssertionLevel assigns cachedFunc, not cachedFunc.assertion, to Debug[key]. shouldAssertFunction installs noop. These are writes, not reflection for display. Fixing the restoration statement changes JavaScript, outside this unit.

- `src/compiler/debug.ts:170:31`: `(Debug as any)[key] = cachedFunc;`
- `src/compiler/debug.ts:190:23`: `(Debug as any)[name] = noop;`
### host-standard

Array.at, Error.captureStackTrace, String.fromCodePoint and Error.stackTraceLimit. Some are standard built-ins, others Node/V8 additions; availability is already tested. They are not parsed JSON.

- `src/compiler/core.ts:1066:119`: `export const elementAt: <T>(array: readonly T[] | undefined, offset: number) => T | undefined = !!(Array.prototype as any).at`
- `src/compiler/core.ts:1067:36`: `? (array, offset) => (array as any)?.at(offset)`
### program-feature

A builder family is probed for a method and conditionally invoked. Calling the method on a union requires declaring the optional capability; claiming all programs have the full semantic builder interface is too strong.

- `src/compiler/tsbuildPublic.ts:992:34`: `((program as any as SemanticDiagnosticsBuilderProgram).getSemanticDiagnosticsOfNextAffectedFile) &&`
- `src/compiler/tsbuildPublic.ts:993:33`: `(program as any as SemanticDiagnosticsBuilderProgram).getSemanticDiagnosticsOfNextAffectedFile(cancellationToken, ignoreSourceFile),`
### watch-callback

File and directory callback families have different arities; sys uses a never middle parameter as a variance workaround, while watchUtilities logs and forwards spread arguments. Cannot prove callback compatibility by changing just the spread array element type.

- `src/compiler/sys.ts:517:30`: `(param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))`
- `src/compiler/sys.ts:517:58`: `(param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))`
### comparer

deduplicateSorted falls back to compareStringsCaseSensitive while T is unconstrained. Casting the string comparator to Comparer<T> has a missing domain proof.

- `src/compiler/core.ts:810:120`: `return deduplicateSorted(toSorted(array, comparer), equalityComparer ?? comparer ?? compareStringsCaseSensitive as any as Comparer<T>);`
### config-host

watch adapts a typed System into a config-reading host. It is not untyped host input, and the any cast may conceal optional versus required capabilities.

- `src/compiler/watch.ts:221:49`: `const host: ParseConfigFileHost = system as any;`
### distribution

UnionToIntersection uses a naked type parameter conditional solely to distribute the union. Must verify never, unions and generic instantiations before any change.

- `src/compiler/types.ts:10331:49`: `export type UnionToIntersection<U> = (U extends any ? (k: U) => void : never) extends ((k: infer I) => void) ? I : never;`
### generator-input

yield* can forward caller next values to the delegated iterator, even though this generator does not name a received value. Declaring it void or undefined without checking delegated iterator contracts would overstate the proof.

- `src/compiler/core.ts:437:136`: `export function* flatMapIterator<T, U>(iter: Iterable<T>, mapfn: (x: T) => readonly U[] | Iterable<U> | undefined): Generator<U, void, any> {`
### module-key

Runtime string indexes an enum and falls back to Node16; derive the valid key union rather than asserting every arbitrary string a valid enum key.

- `src/compiler/program.ts:4379:81`: `const moduleResolutionName = ModuleResolutionKind[moduleKindName as any] ? moduleKindName : "Node16";`
### node-family

widenTypeForVariableLikeDeclaration is an internal checker helper. Derive its declaration union from callers and property reads, rather than treating AST objects as unknown.

- `src/compiler/checker.ts:12480:87`: `function widenTypeForVariableLikeDeclaration(type: Type | undefined, declaration: any, reportErrors?: boolean) {`
### resolution-write

The primary resolution result is updated after lookup; the any cast is on the assignment target. Replacing it with a value cast would not fix its ownership or readonly contract.

- `src/compiler/resolutionCache.ts:563:46`: `(primaryResult.resolvedModule as any) = resolvedModule;`
