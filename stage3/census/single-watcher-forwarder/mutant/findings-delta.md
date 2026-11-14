# MUTANT: complete finding deltas

Deduplicated by (kind, where, reason, text). All paths are relative to the adapted root. Raw JSONL retains every attempt and phase.

## Exact gone

| Kind | Location | Reason |
|---|---|---|
| Boundary | src/compiler/sys.ts:1028:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1050:21 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1054:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1059:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1063:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1098:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1099:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1120:13 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1170:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1184:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1251:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1267:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1275:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1279:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1386:5 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1517:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1540:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1540:65 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1543:21 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1544:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1547:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1547:56 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1549:29 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1816:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1867:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1868:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1965:1 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:512:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:592:1 | function signature or prologue failed; continued at next declaration |
| Boundary | src/compiler/sys.ts:648:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:766:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:770:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:779:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:784:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:811:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:831:5 | function signature or prologue failed; continued at next declaration |
| Boundary | src/compiler/sys.ts:939:5 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:995:1 | function signature or prologue failed; continued at next declaration |
| NotYet | src/compiler/sys.ts:1028:54 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1050:140 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1059:84 | reading getModifiedTime |
| NotYet | src/compiler/sys.ts:1063:90 | reading getModifiedTime |
| NotYet | src/compiler/sys.ts:1098:40 | optional record property access |
| NotYet | src/compiler/sys.ts:1101:30 | reading defaultFallbackPolling |
| NotYet | src/compiler/sys.ts:1123:17 | reading fileSystemEntryExists |
| NotYet | src/compiler/sys.ts:1184:48 | optional record property access |
| NotYet | src/compiler/sys.ts:1251:77 | function identity comparison in a program with specialized generic function values |
| NotYet | src/compiler/sys.ts:1267:89 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1275:17 | reading presentWatcher |
| NotYet | src/compiler/sys.ts:1279:24 | reading presentWatcher |
| NotYet | src/compiler/sys.ts:1500:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot |
| NotYet | src/compiler/sys.ts:1517:15 | a function value that captures the variable its own initializer declares |
| NotYet | src/compiler/sys.ts:1588:191 | process.recordreplay as a value (use its supported process operation directly) |
| NotYet | src/compiler/sys.ts:1588:192 | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| NotYet | src/compiler/sys.ts:1598:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot |
| NotYet | src/compiler/sys.ts:1816:24 | reading byteOrderMarkIndicator |
| NotYet | src/compiler/sys.ts:1965:12 | a literal method through a view that erases its receiver |
| NotYet | src/compiler/sys.ts:648:29 | optional record property access |
| NotYet | src/compiler/sys.ts:766:13 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:770:9 | assigning to an Identifier |
| NotYet | src/compiler/sys.ts:779:9 | assigning to an Identifier |
| NotYet | src/compiler/sys.ts:784:17 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:811:117 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:939:13 | optional record property access |
| Refused | src/compiler/sys.ts:1005:5 | a method in object destructuring |
| Refused | src/compiler/sys.ts:1196:13 | a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read |
| Refused | src/compiler/sys.ts:1212:25 | a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read |
| Refused | src/compiler/sys.ts:1385:31 | a method read as a value (writeFile would lose its object, and this with it) |
| Refused | src/compiler/sys.ts:1391:49 | inherited library member call read as an own field |
| Refused | src/compiler/sys.ts:1532:13 | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| Refused | src/compiler/sys.ts:1533:13 | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| Refused | src/compiler/sys.ts:1610:22 | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| Refused | src/compiler/sys.ts:1612:77 | a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean |
| Refused | src/compiler/sys.ts:1867:17 | sort without a comparator |
| Refused | src/compiler/sys.ts:1868:17 | sort without a comparator |
| Refused | src/compiler/sys.ts:1965:12 | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| Refused | src/compiler/sys.ts:514:17 | a cast the runtime can't check |
| Refused | src/compiler/sys.ts:596:5 | a method in object destructuring |
| Refused | src/compiler/sys.ts:620:33 | a cast the runtime can't check |
| Refused | src/compiler/sys.ts:792:87 | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| Refused | src/compiler/sys.ts:836:13 | a value of type any: any has no proven type |
| Refused | src/compiler/sys.ts:879:34 | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |

## Exact new

| Kind | Location | Reason |
|---|---|---|
| Boundary | src/compiler/sys.ts:1029:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1051:21 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1055:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1060:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1064:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1099:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1100:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1121:13 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1171:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1185:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1254:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1270:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1278:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1282:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1389:5 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1520:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1543:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1543:65 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1546:21 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1547:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1550:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1550:56 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1552:29 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1819:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1870:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1871:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1968:1 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:513:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:593:1 | function signature or prologue failed; continued at next declaration |
| Boundary | src/compiler/sys.ts:649:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:767:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:771:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:780:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:785:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:812:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:832:5 | function signature or prologue failed; continued at next declaration |
| Boundary | src/compiler/sys.ts:940:5 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:996:1 | function signature or prologue failed; continued at next declaration |
| NotYet | src/compiler/sys.ts:1029:54 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1051:140 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1060:84 | reading getModifiedTime |
| NotYet | src/compiler/sys.ts:1064:90 | reading getModifiedTime |
| NotYet | src/compiler/sys.ts:1099:40 | optional record property access |
| NotYet | src/compiler/sys.ts:1102:30 | reading defaultFallbackPolling |
| NotYet | src/compiler/sys.ts:1124:17 | reading fileSystemEntryExists |
| NotYet | src/compiler/sys.ts:1185:48 | optional record property access |
| NotYet | src/compiler/sys.ts:1254:77 | function identity comparison in a program with specialized generic function values |
| NotYet | src/compiler/sys.ts:1270:89 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1278:17 | reading presentWatcher |
| NotYet | src/compiler/sys.ts:1282:24 | reading presentWatcher |
| NotYet | src/compiler/sys.ts:1503:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot |
| NotYet | src/compiler/sys.ts:1520:15 | a function value that captures the variable its own initializer declares |
| NotYet | src/compiler/sys.ts:1591:191 | process.recordreplay as a value (use its supported process operation directly) |
| NotYet | src/compiler/sys.ts:1591:192 | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| NotYet | src/compiler/sys.ts:1601:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot |
| NotYet | src/compiler/sys.ts:1819:24 | reading byteOrderMarkIndicator |
| NotYet | src/compiler/sys.ts:1968:12 | a literal method through a view that erases its receiver |
| NotYet | src/compiler/sys.ts:649:29 | optional record property access |
| NotYet | src/compiler/sys.ts:767:13 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:771:9 | assigning to an Identifier |
| NotYet | src/compiler/sys.ts:780:9 | assigning to an Identifier |
| NotYet | src/compiler/sys.ts:785:17 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:812:117 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:940:13 | optional record property access |
| Refused | src/compiler/sys.ts:1006:5 | a method in object destructuring |
| Refused | src/compiler/sys.ts:1197:13 | a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read |
| Refused | src/compiler/sys.ts:1214:25 | a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read |
| Refused | src/compiler/sys.ts:1388:31 | a method read as a value (writeFile would lose its object, and this with it) |
| Refused | src/compiler/sys.ts:1394:49 | inherited library member call read as an own field |
| Refused | src/compiler/sys.ts:1535:13 | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| Refused | src/compiler/sys.ts:1536:13 | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| Refused | src/compiler/sys.ts:1613:22 | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| Refused | src/compiler/sys.ts:1615:77 | a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean |
| Refused | src/compiler/sys.ts:1870:17 | sort without a comparator |
| Refused | src/compiler/sys.ts:1871:17 | sort without a comparator |
| Refused | src/compiler/sys.ts:1968:12 | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| Refused | src/compiler/sys.ts:515:17 | a cast the runtime can't check |
| Refused | src/compiler/sys.ts:597:5 | a method in object destructuring |
| Refused | src/compiler/sys.ts:621:33 | a cast the runtime can't check |
| Refused | src/compiler/sys.ts:793:87 | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| Refused | src/compiler/sys.ts:837:13 | a value of type any: any has no proven type |
| Refused | src/compiler/sys.ts:880:34 | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |

## Source-aligned gone

| Kind | Location | Reason |
|---|---|---|

## Source-aligned new

| Kind | Location | Reason |
|---|---|---|
