# B: complete finding deltas

Deduplicated by (kind, where, reason, text). All paths are relative to the adapted root. Raw JSONL retains every attempt and phase.

## Exact gone

| Kind | Location | Reason |
|---|---|---|
| Boundary | src/compiler/core.ts:2330:9 | failed statement; state rolled back; continued at next statement or declaration |
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
| NotYet | src/compiler/core.ts:2330:13 | function identity comparison in a program with specialized generic function values |
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
| Boundary | src/compiler/sys.ts:1025:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1047:21 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1051:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1056:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1060:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1095:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1096:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1117:13 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1167:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1181:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1192:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1208:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1248:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1264:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1272:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1276:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1383:5 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1514:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1537:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1537:65 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1540:21 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1541:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1544:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1544:56 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1546:29 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1813:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1864:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1865:17 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1962:1 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:589:1 | function signature or prologue failed; continued at next declaration |
| Boundary | src/compiler/sys.ts:645:25 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:763:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:767:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:776:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:781:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:808:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:828:5 | function signature or prologue failed; continued at next declaration |
| Boundary | src/compiler/sys.ts:936:5 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:992:1 | function signature or prologue failed; continued at next declaration |
| NotYet | src/compiler/sys.ts:1025:54 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1047:140 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1056:84 | reading getModifiedTime |
| NotYet | src/compiler/sys.ts:1060:90 | reading getModifiedTime |
| NotYet | src/compiler/sys.ts:1095:40 | optional record property access |
| NotYet | src/compiler/sys.ts:1098:30 | reading defaultFallbackPolling |
| NotYet | src/compiler/sys.ts:1120:17 | reading fileSystemEntryExists |
| NotYet | src/compiler/sys.ts:1181:48 | optional record property access |
| NotYet | src/compiler/sys.ts:1248:77 | function identity comparison in a program with specialized generic function values |
| NotYet | src/compiler/sys.ts:1264:89 | a boolean \| undefined variable a function value captures |
| NotYet | src/compiler/sys.ts:1272:17 | reading presentWatcher |
| NotYet | src/compiler/sys.ts:1276:24 | reading presentWatcher |
| NotYet | src/compiler/sys.ts:1497:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot |
| NotYet | src/compiler/sys.ts:1514:15 | a function value that captures the variable its own initializer declares |
| NotYet | src/compiler/sys.ts:1585:191 | process.recordreplay as a value (use its supported process operation directly) |
| NotYet | src/compiler/sys.ts:1585:192 | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| NotYet | src/compiler/sys.ts:1595:57 | node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot |
| NotYet | src/compiler/sys.ts:1813:24 | reading byteOrderMarkIndicator |
| NotYet | src/compiler/sys.ts:1962:12 | a literal method through a view that erases its receiver |
| NotYet | src/compiler/sys.ts:645:29 | optional record property access |
| NotYet | src/compiler/sys.ts:763:13 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:767:9 | assigning to an Identifier |
| NotYet | src/compiler/sys.ts:776:9 | assigning to an Identifier |
| NotYet | src/compiler/sys.ts:781:17 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:808:117 | reading timerToUpdateChildWatches |
| NotYet | src/compiler/sys.ts:936:13 | optional record property access |
| Refused | src/compiler/sys.ts:1002:5 | a method in object destructuring |
| Refused | src/compiler/sys.ts:1382:31 | a method read as a value (writeFile would lose its object, and this with it) |
| Refused | src/compiler/sys.ts:1388:49 | inherited library member call read as an own field |
| Refused | src/compiler/sys.ts:1529:13 | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| Refused | src/compiler/sys.ts:1530:13 | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| Refused | src/compiler/sys.ts:1607:22 | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| Refused | src/compiler/sys.ts:1609:77 | a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean |
| Refused | src/compiler/sys.ts:1864:17 | sort without a comparator |
| Refused | src/compiler/sys.ts:1865:17 | sort without a comparator |
| Refused | src/compiler/sys.ts:1962:12 | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| Refused | src/compiler/sys.ts:593:5 | a method in object destructuring |
| Refused | src/compiler/sys.ts:617:33 | a cast the runtime can't check |
| Refused | src/compiler/sys.ts:789:87 | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| Refused | src/compiler/sys.ts:833:13 | a value of type any: any has no proven type |
| Refused | src/compiler/sys.ts:876:34 | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |
| SkippedDependency | src/compiler/sys.ts:498:1 | a dependency function whose body has checker diagnostics (measurement skipped) |

## Source-aligned gone

| Kind | Location | Reason |
|---|---|---|
| Boundary | src/compiler/core.ts:2330:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:512:9 | failed statement; state rolled back; continued at next statement or declaration |
| NotYet | src/compiler/core.ts:2330:13 | function identity comparison in a program with specialized generic function values |
| Refused | src/compiler/sys.ts:unmapped-base-514:17 | a cast the runtime can't check |

## Source-aligned new

| Kind | Location | Reason |
|---|---|---|
| Boundary | src/compiler/sys.ts:1192:9 | failed statement; state rolled back; continued at next statement or declaration |
| Boundary | src/compiler/sys.ts:1208:9 | failed statement; state rolled back; continued at next statement or declaration |
| SkippedDependency | src/compiler/sys.ts:498:1 | a dependency function whose body has checker diagnostics (measurement skipped) |
