# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 2 | 1 |
| New findings | 2 | 1 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before Boundary | 86 | 86 |
| before NotYet | 73 | 73 |
| before Refused | 18 | 18 |
| after Boundary | 86 | 86 |
| after NotYet | 73 | 73 |
| after Refused | 18 | 18 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| a PostfixUnaryExpression | 1 | 1 | 1 | 1 | 0 |
| a boolean \| undefined variable a function value captures | 2 | 2 | 2 | 2 | 0 |
| a call through ?. (an optional call) | 1 | 1 | 1 | 1 | 0 |
| a computed field name | 1 | 1 | 1 | 1 | 0 |
| a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function value that captures the variable its own initializer declares | 1 | 1 | 1 | 1 | 0 |
| a literal method through a view that erases its receiver | 1 | 1 | 1 | 1 | 0 |
| a method in object destructuring | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (writeFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a namespace member without a lowered binding | 5 | 5 | 5 | 5 | 0 |
| a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 1 | 1 | 1 | 1 | 0 |
| a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it | 1 | 1 | 1 | 1 | 0 |
| a spread after the first field | 1 | 1 | 1 | 1 | 0 |
| a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean | 1 | 1 | 1 | 1 | 0 |
| a value of type any: any has no proven type | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 1 | 1 | 1 | 1 | 0 |
| assigning a non-tuple array's length: array writes must preserve its existing indices | 1 | 1 | 1 | 1 | 0 |
| assigning an element of a value | 1 | 1 | 1 | 1 | 0 |
| assigning to a PropertyAccessExpression | 1 | 1 | 1 | 1 | 0 |
| assigning to an Identifier | 3 | 3 | 3 | 3 | 0 |
| failed statement; state rolled back; continued at next statement or declaration | 83 | 83 | 83 | 83 | 0 |
| function identity comparison in a program with specialized generic function values | 1 | 1 | 1 | 1 | 0 |
| function signature or prologue failed; continued at next declaration | 3 | 3 | 3 | 3 | 0 |
| inherited library member call read as an own field | 1 | 1 | 1 | 1 | 0 |
| node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot | 2 | 2 | 2 | 2 | 0 |
| optional record property access | 4 | 4 | 4 | 4 | 0 |
| process.recordreplay as a value (use its supported process operation directly) | 1 | 0 | 1 | 0 | 0 |
| process.recordreplay with an unrepresented type | 0 | 1 | 0 | 1 | 0 |
| reading WatchDirectoryKind | 3 | 3 | 3 | 3 | 0 |
| reading WatchFileKind | 5 | 5 | 5 | 5 | 0 |
| reading byteOrderMarkIndicator | 1 | 1 | 1 | 1 | 0 |
| reading closeFileWatcherOf | 2 | 2 | 2 | 2 | 0 |
| reading currentModifiedTime | 4 | 4 | 4 | 4 | 0 |
| reading defaultFallbackPolling | 1 | 1 | 1 | 1 | 0 |
| reading directorySeparator | 3 | 3 | 3 | 3 | 0 |
| reading directoryWatcher | 6 | 6 | 6 | 6 | 0 |
| reading elapsed | 1 | 1 | 1 | 1 | 0 |
| reading emptyArray | 2 | 2 | 2 | 2 | 0 |
| reading emptyFileSystemEntries | 1 | 1 | 1 | 1 | 0 |
| reading fileSystemEntryExists | 1 | 1 | 1 | 1 | 0 |
| reading getModifiedTime | 2 | 2 | 2 | 2 | 0 |
| reading isString | 2 | 2 | 2 | 2 | 0 |
| reading noop | 1 | 1 | 1 | 1 | 0 |
| reading presentWatcher | 2 | 2 | 2 | 2 | 0 |
| reading timerToUpdateChildWatches | 2 | 2 | 2 | 2 | 0 |
| reading timestamp | 3 | 3 | 3 | 3 | 0 |
| reading watchDirectoryKind | 1 | 1 | 1 | 1 | 0 |
| reading watchFileKind | 1 | 1 | 1 | 1 | 0 |
| sort without a comparator | 2 | 2 | 2 | 2 | 0 |
| writing a possibly absent optional own field | 2 | 2 | 2 | 2 | 0 |

## Raw Findings

Moved by reason:

- 1: failed statement; state rolled back; continued at next statement or declaration -> failed statement; state rolled back; continued at next statement or declaration
- 1: process.recordreplay as a value (use its supported process operation directly) -> process.recordreplay with an unrepresented type

New findings:

- `/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191`: NotYet: process.recordreplay with an unrepresented type
- `/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1517:9`: Boundary: failed statement; state rolled back; continued at next statement or declaration

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:

- 1: process.recordreplay as a value (use its supported process operation directly) -> process.recordreplay with an unrepresented type

New findings:

- `/workspace/scratch/recordreplay/control/src/compiler/sys.ts:1588:191`: NotYet: process.recordreplay with an unrepresented type

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
