# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 2 | 2 |
| Moved by reason | 0 | 0 |
| New findings | 3 | 3 |
| New sites | 3 | 3 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before Boundary | 86 | 86 |
| before NotYet | 72 | 72 |
| before Refused | 18 | 18 |
| after Boundary | 87 | 87 |
| after NotYet | 72 | 72 |
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
| failed statement; state rolled back; continued at next statement or declaration | 83 | 84 | 83 | 84 | 1 |
| function identity comparison in a program with specialized generic function values | 1 | 1 | 1 | 1 | 0 |
| function signature or prologue failed; continued at next declaration | 3 | 3 | 3 | 3 | 0 |
| inherited library member call read as an own field | 1 | 1 | 1 | 1 | 0 |
| node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot | 2 | 2 | 2 | 2 | 0 |
| optional record property access | 4 | 4 | 4 | 4 | 0 |
| process.recordreplay as a value (use its supported process operation directly) | 1 | 0 | 1 | 0 | 1 |
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
| reading modulePath | 0 | 1 | 0 | 1 | 0 |
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


New findings:

- `/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1612:21`: Boundary: failed statement; state rolled back; continued at next statement or declaration
- `/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:46`: NotYet: reading modulePath
- `/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:21`: Boundary: failed statement; state rolled back; continued at next statement or declaration

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

- `/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1612:21`: Boundary: failed statement; state rolled back; continued at next statement or declaration
- `/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:46`: NotYet: reading modulePath
- `/workspace/scratch/recordreplay/candidate2/src/compiler/sys.ts:1613:21`: Boundary: failed statement; state rolled back; continued at next statement or declaration

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
