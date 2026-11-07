| Caller (adapted file:line) | Receiver / declared slot | Callback parameter signature | Extra arguments with types | Handle destination / use |
| --- | --- | --- | --- | --- |
| src/compiler/sys.ts:384 | { getModifiedTime: NonNullable<System["getModifiedTime"]>; setTimeout: NonNullable<System["setTimeout"]>; } / src/compiler/sys.ts:1447 | (_timeoutType: string, queue: PollingIntervalQueue): void | pollingInterval === PollingInterval.Low ? "pollLowPollingIntervalQueue" : "pollPollingIntervalQueue": "pollLowPollingIntervalQueue" \| "pollPollingIntervalQueue"; pollingIntervalQueue(pollingInterval): PollingIntervalQueue | Stored in PollingIntervalQueue.pollScheduled: boolean (declaration line 258). Two truthiness reads and two writes of false. No handle member read. unknown cannot be assigned to the current boolean annotation. |
| src/compiler/sys.ts:497 | { getModifiedTime: NonNullable<System["getModifiedTime"]>; setTimeout: NonNullable<System["setTimeout"]>; } / src/compiler/sys.ts:1447 | (): void | "pollQueue": "pollQueue" | Stored in local pollScheduled: any (line 470); tested for truthiness and reset to undefined. No cancellation or member read. |
| src/compiler/sys.ts:774 | identifier / src/compiler/sys.ts:1448 | N/A (cancellation) | N/A | cancels timerToUpdateChildWatches: any |
| src/compiler/sys.ts:777 | identifier / src/compiler/sys.ts:1447 | (): void | "timerToUpdateChildWatches": "timerToUpdateChildWatches" | Stored in timerToUpdateChildWatches: any (line 624); truthiness/negation, clearTimeout, reset to undefined, and string interpolation in sysLog at line 813. No handle member read. |
| src/compiler/watchPublic.ts:856 | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> / src/compiler/watchPublic.ts:170 | N/A (cancellation) | N/A | cancels timerToInvalidateFailedLookupResolutions: any |
| src/compiler/watchPublic.ts:867 | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> / src/compiler/watchPublic.ts:168 | (): void | "timerToInvalidateFailedLookupResolutions": "timerToInvalidateFailedLookupResolutions" | Stored in timerToInvalidateFailedLookupResolutions: any (line 443); truthiness, cancellation through WatchHost, reset to undefined. |
| src/compiler/watchPublic.ts:886 | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> / src/compiler/watchPublic.ts:170 | N/A (cancellation) | N/A | cancels timerToUpdateProgram: any |
| src/compiler/watchPublic.ts:889 | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> / src/compiler/watchPublic.ts:168 | (): void | "timerToUpdateProgram": "timerToUpdateProgram" | Stored in timerToUpdateProgram: any (line 442); truthiness, cancellation through WatchHost, reset to undefined. |
| src/compiler/tsbuildPublic.ts:2065 | SolutionBuilderWithWatchHost<T> / src/compiler/watchPublic.ts:170 | N/A (cancellation) | N/A | cancels state.timerToBuildInvalidatedProject: {} |
| src/compiler/tsbuildPublic.ts:2067 | SolutionBuilderWithWatchHost<T> / src/compiler/watchPublic.ts:168 | <T extends BuilderProgram>(_timeoutType: string, state: SolutionBuilderState<T>, changeDetected: boolean): void | "timerToBuildInvalidatedProject": "timerToBuildInvalidatedProject"; state: SolutionBuilderState<T>; changeDetected: boolean | Stored in SolutionBuilderState.timerToBuildInvalidatedProject: unknown (line 424); truthiness, clearTimeout, reset to undefined, matching TIMER-SLOT.md. |
| src/server/utilities.ts:27 | ServerHost / src/server/types.ts:34 | N/A (cancellation) | N/A | cancels pendingTimeout: any |
| src/server/utilities.ts:30 | ServerHost / src/server/types.ts:33 | (operationId: string, self: ThrottledOperations, cb: () => void): void | operationId: string; this: this; cb: () => void | Stored as a value in ThrottledOperations.pendingTimeouts: Map<string, any> (line 12). Retrieved as any, tested for truthiness, passed to clearTimeout, and removed from the map. No handle member read. |
| src/server/utilities.ts:39 | ServerHost / src/server/types.ts:34 | N/A (cancellation) | N/A | cancels pendingTimeout: any |
| src/server/utilities.ts:63 | ServerHost / src/server/types.ts:33 | (self: GcTimer): void | this: this | Stored in GcTimer.timerId: any (line 53); compared with undefined using !== at line 60 and reset to undefined at line 67. No handle member read. |
| src/server/session.ts:406 | ServerHost / src/server/types.ts:33 | (): void | actionType: string | Passed to setTimerHandle(timerHandle: any) at line 450, then stored in this.timerHandle: any (line 370). Compared with undefined, canceled, reset, and converted to a boolean by !!. No handle member read. |
| src/server/session.ts:452 | ServerHost / src/server/types.ts:34 | N/A (cancellation) | N/A | cancels this.timerHandle: any |
| src/server/typingInstallerAdapter.ts:242 | server.ServerHost / src/server/types.ts:33 | (): void | `${request.projectName}::${request.kind}`: string | Return value discarded. |
| src/testRunner/parallel/host.ts:320 | identifier / node_modules/@types/node/web-globals/timers.d.ts:41 | N/A (cancellation) | N/A | cancels worker.timer: any |
| src/testRunner/parallel/host.ts:326 | identifier / node_modules/@types/node/web-globals/timers.d.ts:27 | (timeout: TaskTimeout): never | data.payload: TaskTimeout | Stored in worker.timer: any, canceled with the global clearTimeout. This is a global timer call, outside System's slot. |
| src/testRunner/parallel/worker.ts:61 | this / src/testRunner/parallel/worker.ts:67 | N/A (cancellation) | N/A | no argument; unrelated method |
| src/testRunner/unittests/sys/symlinkWatching.ts:20 | ts.System / src/compiler/sys.ts:1447 | (): void | none | Return value discarded. op has type () => void; no extra arguments. |
| src/testRunner/unittests/tsserver/reloadProjects.ts:86 | TestServerHost / src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111 | N/A (cancellation) | N/A | cancels timeoutId: number |

Implementing, forwarding, binding and declaration members follow. The kind distinguishes an implementation from a public declaration or local alias. Fits are stock checker assignability to the candidate method type; optional absence is not a callable implementation.

| Member (file:line) | Kind / owner | Inferred or declared type | Fits A | Fits B | Distinct reference spans (including definition) |
| --- | --- | --- | --- | --- | --- |
| src/compiler/sys.ts:57 | FunctionDeclaration / None | (handler: () => void, timeout: number) => NodeJS.Timeout | True | False | 3 |
| src/compiler/sys.ts:58 | FunctionDeclaration / None | (handle: NodeJS.Timeout \| undefined) => void | True | True | 3 |
| src/compiler/sys.ts:253 | PropertySignature / host | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 3 |
| src/compiler/sys.ts:466 | PropertySignature / host | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 3 |
| src/compiler/sys.ts:590 | PropertySignature / RecursiveDirectoryWatcherHost | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 3 |
| src/compiler/sys.ts:591 | PropertySignature / RecursiveDirectoryWatcherHost | (timeoutId: any) => void | True | True | 3 |
| src/compiler/sys.ts:606 | BindingElement / {     watchDirectory,     useCaseSensitiveFileNames,     getCurrentDirectory,     getAccessibleS | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 4 |
| src/compiler/sys.ts:607 | BindingElement / {     watchDirectory,     useCaseSensitiveFileNames,     getCurrentDirectory,     getAccessibleS | (timeoutId: any) => void | True | True | 4 |
| src/compiler/sys.ts:977 | PropertySignature / CreateSystemWatchFunctions | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 4 |
| src/compiler/sys.ts:978 | PropertySignature / CreateSystemWatchFunctions | (timeoutId: any) => void | True | True | 4 |
| src/compiler/sys.ts:1000 | BindingElement / {     pollingWatchFileWorker,     getModifiedTime,     setTimeout,     clearTimeout,     fsWatc | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 7 |
| src/compiler/sys.ts:1001 | BindingElement / {     pollingWatchFileWorker,     getModifiedTime,     setTimeout,     clearTimeout,     fsWatc | (timeoutId: any) => void | True | True | 5 |
| src/compiler/sys.ts:1061 | ShorthandPropertyAssignment / createDynamicPriorityPollingWatchFile({ getModifiedTime, setTimeout }) | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 6 |
| src/compiler/sys.ts:1065 | ShorthandPropertyAssignment / createFixedChunkSizePollingWatchFile({ getModifiedTime, setTimeout }) | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 6 |
| src/compiler/sys.ts:1129 | ShorthandPropertyAssignment / createDirectoryWatcherSupportingRecursive({                 useCaseSensitiveFileNames,             | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 6 |
| src/compiler/sys.ts:1130 | ShorthandPropertyAssignment / createDirectoryWatcherSupportingRecursive({                 useCaseSensitiveFileNames,             | (timeoutId: any) => void | True | True | 4 |
| src/compiler/sys.ts:1447 | MethodSignature / System | ((callback: (...args: any[]) => void, ms: number, ...args: any[]) => any) \| undefined | True | True | 18 |
| src/compiler/sys.ts:1448 | MethodSignature / System | ((timeoutId: any) => void) \| undefined | True | True | 15 |
| src/compiler/sys.ts:1501 | ShorthandPropertyAssignment / createSystemWatchFunctions({             pollingWatchFileWorker: fsWatchFileWorker,             ge | (handler: () => void, timeout: number) => NodeJS.Timeout | True | False | 6 |
| src/compiler/sys.ts:1502 | ShorthandPropertyAssignment / createSystemWatchFunctions({             pollingWatchFileWorker: fsWatchFileWorker,             ge | (handle: NodeJS.Timeout \| undefined) => void | True | True | 6 |
| src/compiler/sys.ts:1599 | ShorthandPropertyAssignment / nodeSystem | (handler: () => void, timeout: number) => NodeJS.Timeout | True | False | 20 |
| src/compiler/sys.ts:1600 | ShorthandPropertyAssignment / nodeSystem | (handle: NodeJS.Timeout \| undefined) => void | True | True | 17 |
| src/compiler/watch.ts:678 | PropertyAssignment / return {         onWatchStatusChange,         watchFile: system.watchFile?.bind(system) \|\| returnN | any | True | True | 9 |
| src/compiler/watch.ts:679 | PropertyAssignment / return {         onWatchStatusChange,         watchFile: system.watchFile?.bind(system) \|\| returnN | any | True | True | 9 |
| src/compiler/watchPublic.ts:168 | MethodSignature / WatchHost | ((callback: (...args: any[]) => void, ms: number, ...args: any[]) => any) \| undefined | True | True | 9 |
| src/compiler/watchPublic.ts:170 | MethodSignature / WatchHost | ((timeoutId: any) => void) \| undefined | True | True | 9 |
| src/server/types.ts:33 | MethodSignature / ServerHost | (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | True | True | 18 |
| src/server/types.ts:34 | MethodSignature / ServerHost | (timeoutId: any) => void | True | True | 15 |
| src/harness/harnessLanguageService.ts:593 | MethodDeclaration / SessionServerHost | (_callback: (...args: any[]) => void, _ms: number, ..._args: any[]) => any | True | True | 18 |
| src/harness/harnessLanguageService.ts:597 | MethodDeclaration / SessionServerHost | (_timeoutId: any) => void | True | True | 15 |
| src/lib/dom.generated.d.ts:41718 | MethodSignature / WindowOrWorkerGlobalScope | { (id: number \| undefined): void; (id: number \| undefined): void; (id: number \| undefined): void; } | True | True | 3 |
| src/lib/dom.generated.d.ts:41731 | MethodSignature / WindowOrWorkerGlobalScope | { (handler: TimerHandler, timeout?: number, ...arguments: any[]): number; (handler: TimerHandler, timeout?: number, ...arguments: any[]): number; (handler: TimerHandler, timeout?: number, ...arguments: any[]): number; } | True | True | 3 |
| src/lib/dom.generated.d.ts:44094 | FunctionDeclaration / None | { (id: number \| undefined): void; (id: number \| undefined): void; (id: number \| undefined): void; (timeout: NodeJS.Timeout \| string \| number \| undefined): void; } | True | True | 7 |
| src/lib/dom.generated.d.ts:44107 | FunctionDeclaration / None | typeof setTimeout | True | True | 8 |
| src/lib/webworker.generated.d.ts:13824 | MethodSignature / WindowOrWorkerGlobalScope | { (id: number \| undefined): void; (id: number \| undefined): void; (id: number \| undefined): void; } | True | True | 3 |
| src/lib/webworker.generated.d.ts:13837 | MethodSignature / WindowOrWorkerGlobalScope | { (handler: TimerHandler, timeout?: number, ...arguments: any[]): number; (handler: TimerHandler, timeout?: number, ...arguments: any[]): number; (handler: TimerHandler, timeout?: number, ...arguments: any[]): number; } | True | True | 3 |
| src/lib/webworker.generated.d.ts:14944 | FunctionDeclaration / None | { (id: number \| undefined): void; (id: number \| undefined): void; (id: number \| undefined): void; (timeout: NodeJS.Timeout \| string \| number \| undefined): void; } | True | True | 7 |
| src/lib/webworker.generated.d.ts:14957 | FunctionDeclaration / None | typeof setTimeout | True | True | 8 |
| src/testRunner/parallel/worker.ts:67 | MethodDeclaration / class extends (base as typeof Mocha.Runnable) {             override resetTimeout() {              | () => void | True | True | 3 |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468 | PropertyAssignment / createSystemWatchFunctions({             // We dont have polling watch file             // it is e | any | True | True | 4 |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469 | PropertyAssignment / createSystemWatchFunctions({             // We dont have polling watch file             // it is e | any | True | True | 4 |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103 | MethodDeclaration / TestServerHost | (callback: TimeOutCallback, ms: number, ...args: any[]) => number | True | True | 18 |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111 | MethodDeclaration / TestServerHost | (timeoutId: any) => void | True | True | 15 |
| src/testRunner/unittests/tscWatch/watchApi.ts:190 | PropertyWrite / host | undefined | True | True | 9 |
| src/testRunner/unittests/tscWatch/watchApi.ts:191 | PropertyWrite / host | undefined | True | True | 9 |
| src/testRunner/unittests/tsserver/session.ts:42 | MethodDeclaration / mockHost | () => number | True | True | 18 |
| src/testRunner/unittests/tsserver/session.ts:45 | PropertyAssignment / mockHost | (_?: unknown) => void | True | True | 15 |
| src/tsserver/nodeServer.ts:268 | PropertyWrite / sys | typeof setTimeout | True | True | 18 |
| src/tsserver/nodeServer.ts:269 | PropertyWrite / sys | { (id: number \| undefined): void; (id: number \| undefined): void; (id: number \| undefined): void; (timeout: NodeJS.Timeout \| string \| number \| undefined): void; } | True | True | 15 |

Timer-bearing object literals, including spread forwarding, checked by the stock API:

| Object (file:line) | Context | Timer slots fit A | Timer slots fit B | Spread expressions |
| --- | --- | --- | --- | --- |
| src/compiler/sys.ts:1061 | { getModifiedTime: NonNullable<System["getModifiedTime"]>; setTimeout: NonNullable<System["setTimeout"]>; } | True | True | none |
| src/compiler/sys.ts:1065 | { getModifiedTime: NonNullable<System["getModifiedTime"]>; setTimeout: NonNullable<System["setTimeout"]>; } | True | True | none |
| src/compiler/sys.ts:1122 | RecursiveDirectoryWatcherHost | True | True | none |
| src/compiler/sys.ts:1498 | CreateSystemWatchFunctions | True | False | none |
| src/compiler/sys.ts:1519 | System | True | False | none |
| src/compiler/watch.ts:674 | WatchHost | True | True | none |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:462 | import("../../_namespaces/ts.js").CreateSystemWatchFunctions | True | True | none |
| src/testRunner/unittests/tsserver/session.ts:13 | ts.server.ServerHost | True | True | none |

Inherited hosts and timer neighbors. DirectoryStructureHost and the harness fake System are included even though neither declares a timer:

| Host (file:line) | Heritage | Timer slots and their declaration locations | Explicit source edit forced |
| --- | --- | --- | --- |
| src/compiler/sys.ts:583 RecursiveDirectoryWatcherHost | None | setTimeout from src/compiler/sys.ts:590; clearTimeout from src/compiler/sys.ts:591 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/sys.ts:972 CreateSystemWatchFunctions | None | setTimeout from src/compiler/sys.ts:977; clearTimeout from src/compiler/sys.ts:978 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/sys.ts:1402 System | None | setTimeout from src/compiler/sys.ts:1447; clearTimeout from src/compiler/sys.ts:1448 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/watchUtilities.ts:68 DirectoryStructureHost | None | none | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/watchUtilities.ts:89 CachedDirectoryStructureHost | ['extends DirectoryStructureHost'] | none | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/watchPublic.ts:159 WatchHost | None | setTimeout from src/compiler/watchPublic.ts:168; clearTimeout from src/compiler/watchPublic.ts:170 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/watchPublic.ts:276 WatchCompilerHost | ['extends ProgramHost<T>, WatchHost'] | setTimeout from src/compiler/watchPublic.ts:168; clearTimeout from src/compiler/watchPublic.ts:170 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/watchPublic.ts:290 WatchCompilerHostOfFilesAndCompilerOptions | ['extends WatchCompilerHost<T>'] | setTimeout from src/compiler/watchPublic.ts:168; clearTimeout from src/compiler/watchPublic.ts:170 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/watchPublic.ts:306 WatchCompilerHostOfConfigFile | ['extends WatchCompilerHost<T>, ConfigFileDiagnosticsReporter'] | setTimeout from src/compiler/watchPublic.ts:168; clearTimeout from src/compiler/watchPublic.ts:170 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/watchPublic.ts:329 WatchCompilerHostOfConfigFile | ['extends WatchCompilerHost<T>'] | setTimeout from src/compiler/watchPublic.ts:168; clearTimeout from src/compiler/watchPublic.ts:170 | No; failures are private storage and Node timer heads, not other public slots |
| src/compiler/tsbuildPublic.ts:244 SolutionBuilderWithWatchHost | ['extends SolutionBuilderHostBase<T>, WatchHost'] | setTimeout from src/compiler/watchPublic.ts:168; clearTimeout from src/compiler/watchPublic.ts:170 | No; failures are private storage and Node timer heads, not other public slots |
| src/server/types.ts:29 ServerHost | ['extends System'] | setTimeout from src/server/types.ts:33; clearTimeout from src/server/types.ts:34; setImmediate from src/server/types.ts:35; clearImmediate from src/server/types.ts:36 | No; failures are private storage and Node timer heads, not other public slots |
| src/harness/fakesHosts.ts:24 System | ['implements ts.System'] | none | No; failures are private storage and Node timer heads, not other public slots |
| src/harness/harnessLanguageService.ts:511 SessionServerHost | ['implements ts.server.ServerHost'] | setTimeout from src/harness/harnessLanguageService.ts:593; clearTimeout from src/harness/harnessLanguageService.ts:597; setImmediate from src/harness/harnessLanguageService.ts:601; clearImmediate from src/harness/harnessLanguageService.ts:605 | No; failures are private storage and Node timer heads, not other public slots |
| src/testRunner/parallel/host.ts:69 RemoteTest | ['extends Mocha.Test'] | clearTimeout from node_modules/@types/mocha/index.d.ts:1254 | No; failures are private storage and Node timer heads, not other public slots |
| src/testRunner/parallel/worker.ts:59 None | ['extends (base as typeof Mocha.Runnable)'] | clearTimeout from src/testRunner/parallel/worker.ts:67 | No; failures are private storage and Node timer heads, not other public slots |
| src/testRunner/parallel/worker.ts:103 Hook | ['extends mixin(Mocha.Hook, Timeout)'] | clearTimeout from node_modules/@types/mocha/index.d.ts:1254 | No; failures are private storage and Node timer heads, not other public slots |
| src/testRunner/parallel/worker.ts:109 Test | ['extends mixin(Mocha.Test, Timeout, Clone)'] | clearTimeout from node_modules/@types/mocha/index.d.ts:1254 | No; failures are private storage and Node timer heads, not other public slots |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:396 TestServerHost | ['implements server.ServerHost, FormatDiagnosticsHost, ModuleResolutionHost'] | setTimeout from src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103; clearTimeout from src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111; setImmediate from src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1127; clearImmediate from src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1131 | No; failures are private storage and Node timer heads, not other public slots |

Neighbor declarations (all method/property signatures discovered by the stock AST; indexed System slots inherit their change without a source edit):

| Neighbor (file:line) | Owner | Slot | Explicit edit forced |
| --- | --- | --- | --- |
| src/compiler/sys.ts:253 | host | setTimeout: NonNullable<System["setTimeout"]>; | No separate edit; see paired diagnostics |
| src/compiler/sys.ts:466 | host | setTimeout: NonNullable<System["setTimeout"]>; | No separate edit; see paired diagnostics |
| src/compiler/sys.ts:590 | RecursiveDirectoryWatcherHost | setTimeout: NonNullable<System["setTimeout"]>; | No separate edit; see paired diagnostics |
| src/compiler/sys.ts:591 | RecursiveDirectoryWatcherHost | clearTimeout: NonNullable<System["clearTimeout"]>; | No separate edit; see paired diagnostics |
| src/compiler/sys.ts:977 | CreateSystemWatchFunctions | setTimeout: NonNullable<System["setTimeout"]>; | No separate edit; see paired diagnostics |
| src/compiler/sys.ts:978 | CreateSystemWatchFunctions | clearTimeout: NonNullable<System["clearTimeout"]>; | No separate edit; see paired diagnostics |
| src/compiler/sys.ts:1447 | System | setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any; | No separate edit; see paired diagnostics |
| src/compiler/sys.ts:1448 | System | clearTimeout?(timeoutId: any): void; | No separate edit; see paired diagnostics |
| src/compiler/watchPublic.ts:168 | WatchHost | setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any; | No separate edit; see paired diagnostics |
| src/compiler/watchPublic.ts:170 | WatchHost | clearTimeout?(timeoutId: any): void; | No separate edit; see paired diagnostics |
| src/server/types.ts:33 | ServerHost | setTimeout(callback: (...args: any[]) => void, ms: number, ...args: any[]): any; | No separate edit; see paired diagnostics |
| src/server/types.ts:34 | ServerHost | clearTimeout(timeoutId: any): void; | No separate edit; see paired diagnostics |
| src/lib/dom.generated.d.ts:41718 | WindowOrWorkerGlobalScope | clearTimeout(id: number \| undefined): void; | No separate edit; see paired diagnostics |
| src/lib/dom.generated.d.ts:41731 | WindowOrWorkerGlobalScope | setTimeout(handler: TimerHandler, timeout?: number, ...arguments: any[]): number; | No separate edit; see paired diagnostics |
| src/lib/webworker.generated.d.ts:13824 | WindowOrWorkerGlobalScope | clearTimeout(id: number \| undefined): void; | No separate edit; see paired diagnostics |
| src/lib/webworker.generated.d.ts:13837 | WindowOrWorkerGlobalScope | setTimeout(handler: TimerHandler, timeout?: number, ...arguments: any[]): number; | No separate edit; see paired diagnostics |

Counts and method

All counted references come from TypeScript 6.0.3 LanguageService.findReferences over the whole adapted src, including compiler, services, server, harness, testRunner/unittests and installers. Every discovered named timer member and binding is queried, including System itself. Each per-member count deduplicates (file,line,column) across reference groups and includes declarations. Counts overlap between related symbols and must not be summed. The retained references.json includes every raw group, declaration and use context. Text search was used for discovery of tooling and source reading, never as a reference count.

The first table intentionally includes neighboring and global timers too, with their resolved declaration location. This avoids silently dropping calls forwarded through hosts. Raw handle references and callback declaration text are retained in references.json.

Stock diagnostic comparison

| Tree | Adamic-options diagnostics | Upstream solution diagnostics | Solution exit | Compiler/test build exits |
| --- | --- | --- | --- | --- |
| main | 0 | 0 | 0 | {'main-build-compiler': 0, 'main-build-tests': 0} |
| A | 2 | 2 | 2 | {'A-build-compiler': 1, 'A-build-tests': 1} |
| B | 5 | 5 | 2 | {'B-build-compiler': 1, 'B-build-tests': 1} |

Each added diagnostic follows; repeated diagnostics across the two checker modes are intentional.
A adamic: `src/compiler/sys.ts:384` TS2322: Type 'unknown' is not assignable to type 'boolean'.
A adamic: `src/compiler/sys.ts:384` TS2345: Argument of type '(_timeoutType: string, queue: PollingIntervalQueue) => void' is not assignable to parameter of type '(...args: unknown[]) => void'.   Types of parameters '_timeoutType' and 'args' are incompatible.     Type 'unknown' is not assignable to type 'string'.
A solution: `src/compiler/sys.ts:384` TS2322: Type 'unknown' is not assignable to type 'boolean'.
A solution: `src/compiler/sys.ts:384` TS2345: Argument of type '(_timeoutType: string, queue: PollingIntervalQueue) => void' is not assignable to parameter of type '(...args: unknown[]) => void'.   Types of parameters '_timeoutType' and 'args' are incompatible.     Type 'unknown' is not assignable to type 'string'.
B adamic: `src/compiler/sys.ts:384` TS2322: Type 'unknown' is not assignable to type 'boolean'.
B adamic: `src/compiler/sys.ts:497` TS2554: Expected 2 arguments, but got 3.
B adamic: `src/compiler/sys.ts:777` TS2554: Expected 2 arguments, but got 3.
B adamic: `src/compiler/sys.ts:1501` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.
B adamic: `src/compiler/sys.ts:1599` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.
B solution: `src/compiler/sys.ts:384` TS2322: Type 'unknown' is not assignable to type 'boolean'.
B solution: `src/compiler/sys.ts:497` TS2554: Expected 2 arguments, but got 3.
B solution: `src/compiler/sys.ts:777` TS2554: Expected 2 arguments, but got 3.
B solution: `src/compiler/sys.ts:1501` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.
B solution: `src/compiler/sys.ts:1599` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

Artifact comparison

Compiler and test builds use upstream tasks with --lkg, selecting pinned stock TypeScript 6.0.3. API checks use the separately pinned 6.0.3 compiler API. Adamic-options checks retain all 79 compiler roots, the prelude, Node types and the loader library overlays, rather than relaxing strictFunctionTypes. The stock API solution builder force-builds src/tsconfig.json with every original project option and reference. The actual upstream compiler/test tasks also retain normal typechecking. Main is zero in both modes; candidate failures are preserved.

The typed A/B compiler builds each emitted eight unchanged JavaScript files and skipped tsserverlibrary.js after errors. finish-shim.mjs invokes the unchanged upstream lssl writer callback solely to finish the ninth file. The declaration bundler reads the .d.ts files actually emitted by stock tsc despite diagnostics. These artifact-only completions do not turn either failed build green. See declarations.json and A-declarations.diff.txt / B-declarations.diff.txt for exact built, internal and reference-baseline changes. Baseline edits are drafts derived from actual emitted declaration lines; no reference baseline is written in a candidate. Existing unrelated main-to-upstream API differences are excluded from the incremental diff.

A: compiler JavaScript 9 vs 9 files, byte-identical: True. Internal declaration identical: False.

B: compiler JavaScript 9 vs 9 files, byte-identical: True. Internal declaration identical: False.

Latent census

LATENT_FULL=1 and LATENT_ONLY_FILE select the adapted sys.ts, while all compiler roots remain loaded. The standard scratch overlay disables production Lower. LATENT_ASSERT_NO_OUTPUT=1 checks the disabled ordinary Load and Lower paths. The cited report was read from typescript/next-stops-after-wave; this uses pinned main's make_overlay.py and full-mode driver, with only physical Go evidence filenames changed to .go.txt. No speculative continuation is enabled.

main: raw full census retained in main-latent.jsonl.gz.

A: raw full census retained in A-latent.jsonl.gz.

B: raw full census retained in B-latent.jsonl.gz.

System reference counts and the four answers

| Slot | Declaration | Distinct spans including definition | Spans excluding queried declaration |
| --- | --- | --- | --- |
| setTimeout | src/compiler/sys.ts:1447 | 18 | 17 |
| clearTimeout | src/compiler/sys.ts:1448 | 15 | 14 |

The 711 src TypeScript roots produce 49 audited named declarations/bindings/property writes, eight timer-bearing object literals, and 22 timer-named calls (12 setTimeout and 10 clearTimeout). These are different censuses, not 49 implementations. The two undefined writes at watchApi.ts:190-191 remove optional hooks. The Mocha clearTimeout() call and the parallel worker global timer pair are explicitly outside System. All 49 member reference counts are in the first ledger; related reference groups overlap.

1. The handle operations need no NodeJS.Timeout member. They are storage, truthiness or undefined comparison, cancellation, reset, and one diagnostic interpolation. The current pollScheduled: boolean at sys.ts:258 rejects unknown at sys.ts:384:9 in both candidates. The operations themselves admit unknown if that private field is separately reconsidered; that last statement is inference, not an extra drafted change. The numeric virtual-host return and host-polymorphic public contract do not justify a Node-only handle. The solution-builder unknown precedent is unchanged.

2. The direct System polling callback takes (_timeoutType: string, queue: PollingIntervalQueue), with matching string label and queue arguments. The neighbor callback and argument tuples are fully listed above. A cannot accept that callback. B preserves its tuple but infers [] from pollQueue() and onTimerToUpdateChildWatches(), then rejects their unused string labels at sys.ts:497:74 and 777:83.

3. getNodeSystem installs the file-local Node timer head at sys.ts:1501/1502 and 1599/1600. Its setTimeout fits A and fails B, as both the checker probes and builds prove. The separate native replacements at tsserver/nodeServer.ts:268-269 fit both; the harness SessionServerHost methods, numeric virtual TestServerHost methods and mockHost also fit both. The fake System in harness/fakesHosts.ts:24 supplies neither optional timer. Native clearTimeout passes tsc bivariance in both candidate shapes, but A adds the two exact Adamic refusals below. B's checker-rejected getNodeSystem body prevents an equivalent lowering verdict.

4. No other public slot must be edited for either measured candidate. WatchHost at watchPublic.ts:168/170 remains any and forwards bound System hooks at watch.ts:678/679; WatchCompilerHost, WatchCompilerHostOfFilesAndCompilerOptions, WatchCompilerHostOfConfigFile and SolutionBuilderWithWatchHost inherit its slots. ServerHost at server/types.ts:33/34 independently redeclares the timeout pair; its immediate pair at lines 35/36 is independent of System. DirectoryStructureHost at watchUtilities.ts:68 has no timer at all. The internal RecursiveDirectoryWatcherHost and CreateSystemWatchFunctions use NonNullable<System[...]> at the locations above, so their effective types move automatically. SolutionBuilderHost has no timer unless the with-watch interface is used.

The public AST inventory reconciles adaptation 52 exactly: 16 any tokens on eight timer declarations means System timeout pair (4), WatchHost timeout pair (4), ServerHost timeout pair (4), and ServerHost immediate pair (4). Both candidates leave 12 tokens. No neighbor declaration text changes.

The exact moved lines are built/local/typescript.d.ts:8545-8546, built/local/typescript.internal.d.ts:16663-16664, and draft tests/baselines/reference/api/typescript.d.ts:8491-8492. The .diff.txt files contain full hunks. Both internal declarations move. The baseline draft edits only the AST-identified System members, not WatchHost's identical old spelling.

Full latent site comparison

| Tree | Checker diagnostics | Units | Distinct NotYet/Refused sites | Common with main | New | Not observed from main |
| --- | --- | --- | --- | --- | --- | --- |
| main | 0 | 140 | 123 | N/A | 0 | 0 |
| A | 2 | 140 | 125 | 122 | 3 | 1 |
| B | 5 | 140 | 102 | 102 | 0 | 21 |

Deduplication key is (location, kind, exact reason, phase), matching the cited core report. Boundary events are not compiler stops. Raw streams retain every event and unit. Main has zero checker diagnostics despite the preserved legacy measurement label. A has one split_checker_body; B has four. Deferred generic units are not proven clean.

The timer contract produces one observed main stop, not four separate declaration-token stops: `src/compiler/sys.ts:384:79`, NotYet / refusal_scan, `a member typed any ((0)) related to string at the argument site`. The System declarations at 1447 and 1448 have no independent recorded site. The emission line is:

`internal/lower/any_relations.go:43`:
```go
l.refusal.deferred = l.notYet(node, "a member typed any ("+found.path+") related to "+l.checker.TypeToString(found.target)+" at the "+name+" site")
```

That callback stop is absent in both candidate streams. This is not a successful-lowering claim: the registering function at sys.ts:383 is checker-split in both candidates. A additionally records these new sites:

`src/compiler/sys.ts:1502:13`, Refused / refusal_scan: `a function taking Timeout | undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take`.

`src/compiler/sys.ts:1519:36`, NotYet / refusal_scan: `a member typed any (require()module) related to {} at the initializer site`.

`src/compiler/sys.ts:1600:13`, Refused / refusal_scan: `a function taking Timeout | undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take`.

The two Timeout-to-unknown findings at sys.ts:1502 and 1600 are direct new clearTimeout-contract refusals. Stock tsc accepts these method assignments by bivariance; Adamic deliberately refuses the wider input. The message and raising return are:

`internal/lower/invariance.go:1042`:
```go
what = "a function taking " + l.checker.TypeToString(found.source) + " seen as one taking " + l.checker.TypeToString(found.target) + " (tsc relates a method's parameters both ways), so it can be handed what it can't take"
```

`internal/lower/invariance.go:1053`:
```go
return &Refused{Where: l.program.Where(node), What: what, Fix: fix}
```

The new any-related initializer finding at sys.ts:1519:36 is collateral scanner exposure, not a new timer declaration token. A retains 122 main sites. B records no new site, but its getNodeSystem body is checker-split by the two failed generic timer assignments; the lack of A's new cancellation refusals in B does not prove them safe. B also splits the two zero-parameter callback registration bodies and the dynamic polling registration body, so its 21 absent observations cannot be credited as fixes.

The private `pollScheduled: any` finding at sys.ts:470:9 remains under both A and B: `a value of type any`, raised by `internal/lower/expression.go:53`. Retyping the public slot does not retype private storage. All 123/125/102 sites, their recording units, and every added/absent row are in latent-sites.json.

Provenance, rejected attempts and limits

The base is d15620712c459baf72b3d359c1829df8d0808fda and upstream is 050880ce59e30b356b686bd3144efe24f875ebc8. Setup, npm ci --prefix stage3/api, and both full applies exited 0. The dependency-free canonical apply has no node_modules. All 746 source-file hashes match the measured baseline, and only the exact two declarations differ in each candidate. The initial overlapping dependency install therefore caused no source delta. The separate checkpoint branch is typescript/sys-timer-slot-checkpoint-b5w3ycg at 93b966f1fa00f0489478cadb8a57b55e775c7f86. The final requested branch has one evidence commit on the pinned main. No adaptation, product source or real reference baseline is committed.

Two preparation mistakes are excluded from accepted diagnostic counts: a renamed dependency directory initially broke Node package resolution, and a project API wrapper lacked the proper configuration directory for automatic types. Their rejected logs are retained. Final reference collection uses the corrected dependency layout, Adamic-options checks have zero baseline diagnostics, the stock API solution builder uses the real root project/configuration graph and reports 0/2/5, and the independent upstream compiler/test commands report the same candidate errors. The failed private lssl CLI selection is also retained; its existing writer callback was subsequently invoked successfully. Artifact completion does not claim either candidate builds cleanly.

Reproduce after sourcing /workspace/adamic-tools/env.sh and exporting ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source: install stage3/api, run apply with no upstream dependencies present, install upstream dependencies only after apply, run prepare.py, run.py measure/checks/builds/latent-build/latent, complete-artifacts.py, declarations.py, validate-source.py, latent-analysis.py, summarize.py, and finish-report.py. All commands use scratch paths fixed in the retained scripts. The latent build uses the current main lowering without feature merges, with generated physical .go.txt files. No speculative or replay flags are enabled. This is a bounded full-mode observation ledger, not exhaustive successful lowering or a runtime behavior proof. All nine JavaScript hashes match; declaration changes are intentional.

Inference: the counts support B's tuple-preserving generic direction because the real string-and-queue callback rules out A. They support neither candidate exactly as drafted: both reject the boolean handle field, B also rejects two ignored-argument calls and the current Node timer head, and A exposes two native cancellation variance refusals. A revised generic contract and its provider/storage consequences need a separate reviewed measurement before any declaration lands.

## C: prefix arguments, private storage and Node heads

The requested prefix spelling is retained as C-a/C-b. The union spelling and a separately named full-tuple NoInfer alternative are also measured, without changing A or B. This is a post-apply scratch experiment on the same frozen compiler/adapted main as A/B, not an adaptation or Runtime landing.

| Variant | Slot spelling | Ambient heads | Added stock diagnostics, Adamic / upstream solution | Ordinary compiler / tests exits | Nine compiler JS files identical |
| --- | --- | --- | --- | --- | --- |
| C-a | prefix [...A, ...unknown[]] | 41 retained | 2 / 2 | {'build-compiler': 1, 'build-tests': 1} | yes, all 9 |
| C-b | prefix [...A, ...unknown[]] | deleted | 2 / 2 | {'build-compiler': 1, 'build-tests': 1} | yes, all 9 |
| C-noinfer-a | full tuple + NoInfer | 41 retained | 2 / 2 | {'build-compiler': 1, 'build-tests': 1} | yes, all 9 |
| C-noinfer-b | full tuple + NoInfer | deleted | 0 / 0 | {'build-compiler': 0, 'build-tests': 0} | yes, all 9 |

1. Slot and inference

The requested `...args: [...A, ...unknown[]]` spelling infers A as the typed string/queue prefix and as [] for the two callbacks that ignore their labels. The unused labels then fit the unknown suffix. The tried union spelling `...args: A | [...A, ...unknown[]]` also accepts all three callback shapes. Both fail the provider assignment with either Node head. This is a provider compatibility failure, not failed callback inference. C-inference.json records all six spelling/head trials and their resolved call signatures.

The additional spelling is:
```ts
setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;
clearTimeout?(timeoutId: unknown): void;
```
It infers the complete extra-argument tuple from the arguments because NoInfer prevents the callback from fixing A prematurely. TypeScript permits a callback to read fewer tuple elements, including zero. The small probe succeeds with the @types/node head, and the full C-noinfer-b tree has zero diagnostics in both checker modes. This spelling is not silently substituted for the requested C.

Every added diagnostic follows. Both checker modes produce the same file/line/column/code identities.

C-a adamic: `src/compiler/sys.ts:1501:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...A, ...unknown[]]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

C-a adamic: `src/compiler/sys.ts:1599:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...A, ...unknown[]]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

C-a solution: `src/compiler/sys.ts:1501:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...A, ...unknown[]]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

C-a solution: `src/compiler/sys.ts:1599:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...A, ...unknown[]]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

C-b adamic: `src/compiler/sys.ts:1499:13` TS2322: Type 'typeof setTimeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'args' and 'args' are incompatible.     Type '[...A, ...unknown[]]' is not assignable to type 'MakeVoidParameterOptional<A>'.

C-b adamic: `src/compiler/sys.ts:1597:13` TS2322: Type 'typeof setTimeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'args' and 'args' are incompatible.     Type '[...A, ...unknown[]]' is not assignable to type 'MakeVoidParameterOptional<A>'.

C-b solution: `src/compiler/sys.ts:1499:13` TS2322: Type 'typeof setTimeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'args' and 'args' are incompatible.     Type '[...A, ...unknown[]]' is not assignable to type 'MakeVoidParameterOptional<A>'.

C-b solution: `src/compiler/sys.ts:1597:13` TS2322: Type 'typeof setTimeout' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]) => unknown'.   Types of parameters 'args' and 'args' are incompatible.     Type '[...A, ...unknown[]]' is not assignable to type 'MakeVoidParameterOptional<A>'.

C-noinfer-a adamic: `src/compiler/sys.ts:1501:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

C-noinfer-a adamic: `src/compiler/sys.ts:1599:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

C-noinfer-a solution: `src/compiler/sys.ts:1501:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

C-noinfer-a solution: `src/compiler/sys.ts:1599:13` TS2322: Type '(handler: () => void, timeout: number) => Timeout' is not assignable to type '<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A) => unknown'.   Types of parameters 'timeout' and 'ms' are incompatible.     Type '[ms: number, ...args: A]' is not assignable to type '[timeout: number]'.       Type '[ms: number, ...args: unknown[]]' is not assignable to type '[timeout: number]'.         Target allows only 1 element(s) but source may have more.

2. Private storage and pins

Both private annotations become unknown in every C variant: main sys.ts:258 (PollingIntervalQueue.pollScheduled) and main sys.ts:470 (the local pollScheduled). In the head-deleted variants their physical lines are 256 and 468. Every other source file remains byte-identical, across all 746 files. The drafts contain exactly these two type edits, the two public type edits, and, for (b), the two ambient declaration deletions.

The stock AST maps the enclosing functions to createDynamicPriorityPollingWatchFile (251-386) and createFixedChunkSizePollingWatchFile (464-499). Static inspection found no direct annotation or enclosing-owner pin for either polling slot in the active CJS literals and statically required JSON manifests. This is static pin inspection, not an adapter replay. Adaptation 40's sys.ts class-rule pin is the unrelated stat variable; 90 has no sys.ts target, 91 targets factory/utilities, and 92/93/99-wasm guard other sys.ts owners. Adaptation 41 rules 13 and 14 explicitly pin the two ambient declarations: (b) intentionally deletes them as the requested Runtime counterfactual. No adaptation or pin is modified. See C-pins.json for the inventory and its limits.

3. Node providers and the four census locations

The provider probes below are stock TypeScript checks against each actual contextual System-derived method slot. Deleting the two ambient lines shifts source locations after them by two; every census comparison maps those locations back to main. Raw locations are retained in C.json.

| Variant | Main location | Actual source line | Member | Fits tsc method slot | Full sys.ts findings at that location |
| --- | --- | --- | --- | --- | --- |
| C-a | sys.ts:1501 | 1501 | setTimeout | False | not observed; getNodeSystem checker-split |
| C-a | sys.ts:1502 | 1502 | clearTimeout | True | not observed; getNodeSystem checker-split |
| C-a | sys.ts:1599 | 1599 | setTimeout | False | not observed; getNodeSystem checker-split |
| C-a | sys.ts:1600 | 1600 | clearTimeout | True | not observed; getNodeSystem checker-split |
| C-b | sys.ts:1501 | 1499 | setTimeout | False | not observed; getNodeSystem checker-split |
| C-b | sys.ts:1502 | 1500 | clearTimeout | True | not observed; getNodeSystem checker-split |
| C-b | sys.ts:1599 | 1597 | setTimeout | False | not observed; getNodeSystem checker-split |
| C-b | sys.ts:1600 | 1598 | clearTimeout | True | not observed; getNodeSystem checker-split |
| C-noinfer-a | sys.ts:1501 | 1501 | setTimeout | False | not observed; getNodeSystem checker-split |
| C-noinfer-a | sys.ts:1502 | 1502 | clearTimeout | True | not observed; getNodeSystem checker-split |
| C-noinfer-a | sys.ts:1599 | 1599 | setTimeout | False | not observed; getNodeSystem checker-split |
| C-noinfer-a | sys.ts:1600 | 1600 | clearTimeout | True | not observed; getNodeSystem checker-split |
| C-noinfer-b | sys.ts:1501 | 1499 | setTimeout | True | Refused / refusal_scan: a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take; NotYet / lowering: reading setTimeout |
| C-noinfer-b | sys.ts:1502 | 1500 | clearTimeout | True | Refused / refusal_scan: a function taking string \| number \| Timeout \| undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| C-noinfer-b | sys.ts:1599 | 1597 | setTimeout | True | Refused / refusal_scan: a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| C-noinfer-b | sys.ts:1600 | 1598 | clearTimeout | True | Refused / refusal_scan: a function taking string \| number \| Timeout \| undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take |

With 41 retained the source signatures are `(handler: () => void, timeout: number) => NodeJS.Timeout` and `(handle: NodeJS.Timeout | undefined) => void`. Without them the actual declarations resolve to @types/node/web-globals/timers.d.ts:27 and :41: generic setTimeout with MakeVoidParameterOptional<TArgs>, and clearTimeout accepting NodeJS.Timeout | string | number | undefined. C-check.cjs records the declaration paths and complete signatures. Method bivariance admits the clearTimeout assignment; it does not make the requested generic prefix setTimeout assignment fit either provider.

4. Cancellation witness and remaining lowering refusal

Node v24.19.0 clearTimeout returned undefined without throwing for the number 2147483646, the string "not-a-timer", a plain object {}, and undefined. The witness starts no timer and leaves no pending process. This establishes the requested ordinary runtime cases; it does not change @types/node's declared input domain.

The requested prefix C-a/C-b streams do not record cancellation findings at the four queried locations because getNodeSystem is checker-split by the setTimeout errors. That absence is not a clearance. The independent C-clear-probe.ts.txt has zero checker diagnostics and records:

`a function taking string | number | Timeout | undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take`.

The deciding message is emitted at internal/lower/invariance.go:1042 and the Refused is returned at :1053 on the frozen compiler. C-noinfer-b removes the checker obstruction and directly measures the native cancellation view inside the full sys.ts census. Its exact four-location observations are in the table above.

Full sys.ts census against main

| Variant | Checker diagnostics | Units | Distinct sites | Kept | Gone / not observed | New |
| --- | --- | --- | --- | --- | --- | --- |
| main | 0 | 140 | 123 | 123 | 0 | 0 |
| C-a | 2 | 140 | 105 | 102 | 21 | 3 |
| C-b | 2 | 138 | 101 | 98 | 25 | 3 |
| C-noinfer-a | 2 | 140 | 105 | 102 | 21 | 3 |
| C-noinfer-b | 0 | 138 | 123 | 114 | 9 | 9 |

The key is (main-mapped location, kind, exact reason, phase). Boundary events are excluded. Every kept, gone and new site is listed with its reason in [C-LATENT.md](C-LATENT.md), with raw events and recording units retained in C.json and the compressed JSONL streams. Checker-split bodies, deleted ambient units and generic deferrals are retained explicitly, so a missing observation is not credited as successful lowering. In particular, requested C is blocked only in getNodeSystem, while C-noinfer-b is checker-clean. The removed private any annotation no longer produces the old sys.ts:470:9 `a value of type any` finding; this is separate from the unchanged callback/host/native boundaries.

The checker-clean NoInfer variant also adds private-poller NotYet sites at main-mapped 490:9 and 496:37: `a union of differently held members variable a function value captures`, raised at internal/lower/locals.go:208. At 497:25 it adds `a call through a function value whose rest parameter isn't an array`, raised at internal/lower/expression.go:1704. These are lowering observations exposed by the type-only storage/tuple changes, not JavaScript changes. All nine new sites, including the four provider variance refusals, are retained in the full ledger.

JavaScript and declaration diffs

The ordinary failed builds emit eight compiler JavaScript artifacts; the existing upstream lssl writer supplies the ninth, exactly as documented for A/B. C-noinfer-b emits all nine in its ordinary successful build. Every variant matches every one of the nine main hashes byte for byte. Declaration bundles are generated from the .d.ts files emitted by stock 6.0.3. The complete public/internal and reference-baseline draft hunks are in each C-*-declarations.diff.txt. Only System's two timer members move. NoInfer replaces built public 8545-8546, built internal 16663-16664 and reference 8491-8492 in place. For the requested prefix spelling, the real declaration bundler formats the tuple across eight setTimeout lines: old public 8545-8546 becomes 8545-8553, old internal 16663-16664 becomes 16663-16671, and draft reference 8491-8492 becomes 8491-8499. These are actual formatted output hunks, not the one-line source spelling. The two internal private polling annotations and the ambient deletions do not add a declaration delta. Public and internal declarations do move; JavaScript does not.

Reproduction and scope

Use the existing setup and frozen canonical/main trees. Run C-inference.cjs, C-prepare.py, C-run.py C-checks/C-builds/C-latent/C-witness, C-alternative.py, C-run.py C-checks-alternative/C-builds-alternative/C-latent-alternative, C-pins.cjs, C-declaration-spans.cjs, and C-report.py. NODE_PATH names the installed stage3/api/node_modules and the shell sources /workspace/adamic-tools/env.sh. C-run.py uses the existing bounded runner, stock check options and stock API solution graph; each child's real exit is retained. LATENT_FULL=1, LATENT_ONLY_FILE selects sys.ts, and LATENT_ASSERT_NO_OUTPUT=1 remains enabled. The old compiler binary and all pre-C A/B evidence are unchanged.

Inference: the prefix contract repairs callback inference and unknown makes both private storage annotations consistent with the opaque return, but neither requested C variant is assignable to the installed setTimeout provider. The explicitly named full-tuple NoInfer alternative reaches zero stock diagnostics and identical JavaScript with the ambient heads removed. The Node cancellation witness supports an unknown host input for the four tested runtime values, while the narrower published Node declaration still produces an Adamic variance refusal. Stock-green C-noinfer-b also records the generic callback-view Refused at main-mapped 1501 and 1599: a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void. Stock-green C-noinfer-b is therefore not a lowering-ready landing.
