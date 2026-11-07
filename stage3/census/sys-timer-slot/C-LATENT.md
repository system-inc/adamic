# C full sys.ts site ledger

All locations below are mapped to main; raw source locations and unit identities remain in C.json. Gone means not observed, with checker eligibility caveats in REPORT.md.

## C-a

| State | Main location | Kind | Phase | Exact reason |
| --- | --- | --- | --- | --- |
| kept | src/compiler/core.ts:2332:13 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/core.ts:2365:60 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/debug.ts:227:5 | NotYet | lowering | a checked predicate target whose check would read past the depth bound (HostDirectoryWatcher) |
| kept | src/compiler/debug.ts:252:33 | NotYet | lowering | a value of type never |
| kept | src/compiler/sys.ts:1007:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:1030:54 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1052:140 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1061:84 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1065:90 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1100:40 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1103:30 | NotYet | lowering | reading defaultFallbackPolling |
| kept | src/compiler/sys.ts:1125:17 | NotYet | lowering | reading fileSystemEntryExists |
| kept | src/compiler/sys.ts:1186:48 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1198:13 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read |
| kept | src/compiler/sys.ts:1214:25 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read |
| kept | src/compiler/sys.ts:1253:77 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/sys.ts:1269:89 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1277:17 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:1281:24 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:129:9 | NotYet | lowering | a computed field name |
| kept | src/compiler/sys.ts:1387:31 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:1388:5 | NotYet | lowering | replacing a represented method at runtime |
| kept | src/compiler/sys.ts:1388:5 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:141:10 | NotYet | lowering | a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it |
| kept | src/compiler/sys.ts:144:81 | Refused | refusal_scan | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace |
| kept | src/compiler/sys.ts:162:17 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:162:35 | Refused | refusal_scan | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| kept | src/compiler/sys.ts:1642:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1646:59 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:1647:18 | NotYet | lowering | reading inspector |
| kept | src/compiler/sys.ts:1647:32 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1647:42 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:29 | NotYet | lowering | new a PropertyAccessExpression |
| kept | src/compiler/sys.ts:1651:29 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:33 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:43 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1652:13 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1652:13 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1652:21 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1654:13 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1654:13 | NotYet | lowering | reading session |
| kept | src/compiler/sys.ts:1654:13 | Refused | refusal_scan | a checked view with a callable contract |
| kept | src/compiler/sys.ts:1654:21 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:25 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1690:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1692:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1692:31 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1694:29 | NotYet | lowering | node:fs.StatsBase.isDirectory through an optional call |
| kept | src/compiler/sys.ts:1703:25 | NotYet | lowering | node:fs.writeFileSync: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| kept | src/compiler/sys.ts:1706:21 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1706:21 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1706:23 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1724:41 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| kept | src/compiler/sys.ts:1724:41 | NotYet | lowering | reading __filename |
| kept | src/compiler/sys.ts:178:13 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:1818:24 | NotYet | lowering | reading byteOrderMarkIndicator |
| kept | src/compiler/sys.ts:185:81 | Refused | lowering | a spread after the first field |
| kept | src/compiler/sys.ts:1869:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1870:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1967:12 | Refused | refusal_scan | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| kept | src/compiler/sys.ts:201:79 | NotYet | lowering | a PostfixUnaryExpression |
| kept | src/compiler/sys.ts:207:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:216:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:226:36 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:240:17 | NotYet | lowering | assigning a field of a value |
| kept | src/compiler/sys.ts:263:11 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:288:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:289:23 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:296:61 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:308:64 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:310:19 | Refused | refusal_scan | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| kept | src/compiler/sys.ts:324:13 | Refused | refusal_scan | a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read |
| kept | src/compiler/sys.ts:356:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:368:9 | Refused | lowering | inherited library member push read as an own field |
| kept | src/compiler/sys.ts:436:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:441:29 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:442:33 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:444:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:445:54 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:447:34 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:450:59 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:521:17 | Refused | lowering | a cast the runtime can't check |
| kept | src/compiler/sys.ts:521:17 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:57:1 | NotYet | lowering | a function without a body |
| kept | src/compiler/sys.ts:57:1 | Refused | refusal_scan | an ambient function declaration |
| kept | src/compiler/sys.ts:58:1 | NotYet | lowering | a function without a body |
| kept | src/compiler/sys.ts:58:1 | Refused | refusal_scan | an ambient function declaration |
| kept | src/compiler/sys.ts:603:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:627:33 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:655:29 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:690:27 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:739:9 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:773:13 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:777:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:781:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:786:17 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:794:87 | Refused | refusal_scan | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| kept | src/compiler/sys.ts:813:117 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:838:13 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:881:34 | Refused | refusal_scan | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |
| kept | src/compiler/sys.ts:941:13 | NotYet | lowering | optional record property access |
| gone | src/compiler/sys.ts:1494:121 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1494:35 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1494:35 | NotYet | lowering | reading __filename |
| gone | src/compiler/sys.ts:1494:92 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1519:15 | NotYet | lowering | a function value that captures the variable its own initializer declares |
| gone | src/compiler/sys.ts:1534:13 | Refused | refusal_scan | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| gone | src/compiler/sys.ts:1535:13 | Refused | refusal_scan | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| gone | src/compiler/sys.ts:1578:21 | NotYet | lowering | node:fs.StatsBase.isFile through an optional call |
| gone | src/compiler/sys.ts:1588:42 | NotYet | lowering | a union of differently held members variable a function value captures |
| gone | src/compiler/sys.ts:1590:191 | NotYet | lowering | process.recordreplay as a value (use its supported process operation directly) |
| gone | src/compiler/sys.ts:1590:192 | NotYet | refusal_scan | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| gone | src/compiler/sys.ts:1605:32 | Refused | refusal_scan | a cast the runtime can't check |
| gone | src/compiler/sys.ts:1605:33 | NotYet | refusal_scan | node:process.Process.stdout used as a value |
| gone | src/compiler/sys.ts:1612:22 | Refused | refusal_scan | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| gone | src/compiler/sys.ts:1614:77 | Refused | refusal_scan | a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read |
| gone | src/compiler/sys.ts:1967:12 | NotYet | lowering | a literal method through a view that erases its receiver |
| gone | src/compiler/sys.ts:384:79 | NotYet | refusal_scan | a member typed any ((0)) related to string at the argument site |
| gone | src/compiler/sys.ts:470:9 | NotYet | lowering | a value of type any |
| gone | src/compiler/sys.ts:490:9 | NotYet | lowering | assigning to an Identifier |
| gone | src/compiler/sys.ts:496:37 | NotYet | lowering | reading pollScheduled |
| gone | src/compiler/sys.ts:497:9 | NotYet | lowering | assigning to an Identifier |
| new | src/compiler/sys.ts:490:9 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:496:37 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:497:9 | NotYet | lowering | a union of differently held members variable a function value captures |

## C-b

| State | Main location | Kind | Phase | Exact reason |
| --- | --- | --- | --- | --- |
| kept | src/compiler/core.ts:2332:13 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/core.ts:2365:60 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/debug.ts:227:5 | NotYet | lowering | a checked predicate target whose check would read past the depth bound (HostDirectoryWatcher) |
| kept | src/compiler/debug.ts:252:33 | NotYet | lowering | a value of type never |
| kept | src/compiler/sys.ts:1007:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:1030:54 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1052:140 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1061:84 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1065:90 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1100:40 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1103:30 | NotYet | lowering | reading defaultFallbackPolling |
| kept | src/compiler/sys.ts:1125:17 | NotYet | lowering | reading fileSystemEntryExists |
| kept | src/compiler/sys.ts:1186:48 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1198:13 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read |
| kept | src/compiler/sys.ts:1214:25 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read |
| kept | src/compiler/sys.ts:1253:77 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/sys.ts:1269:89 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1277:17 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:1281:24 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:129:9 | NotYet | lowering | a computed field name |
| kept | src/compiler/sys.ts:1387:31 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:1388:5 | NotYet | lowering | replacing a represented method at runtime |
| kept | src/compiler/sys.ts:1388:5 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:141:10 | NotYet | lowering | a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it |
| kept | src/compiler/sys.ts:144:81 | Refused | refusal_scan | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace |
| kept | src/compiler/sys.ts:162:17 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:162:35 | Refused | refusal_scan | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| kept | src/compiler/sys.ts:1642:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1646:59 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:1647:18 | NotYet | lowering | reading inspector |
| kept | src/compiler/sys.ts:1647:32 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1647:42 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:29 | NotYet | lowering | new a PropertyAccessExpression |
| kept | src/compiler/sys.ts:1651:29 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:33 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:43 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1652:13 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1652:13 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1652:21 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1654:13 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1654:13 | NotYet | lowering | reading session |
| kept | src/compiler/sys.ts:1654:13 | Refused | refusal_scan | a checked view with a callable contract |
| kept | src/compiler/sys.ts:1654:21 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:25 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1690:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1692:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1692:31 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1694:29 | NotYet | lowering | node:fs.StatsBase.isDirectory through an optional call |
| kept | src/compiler/sys.ts:1703:25 | NotYet | lowering | node:fs.writeFileSync: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| kept | src/compiler/sys.ts:1706:21 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1706:21 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1706:23 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1724:41 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| kept | src/compiler/sys.ts:1724:41 | NotYet | lowering | reading __filename |
| kept | src/compiler/sys.ts:178:13 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:1818:24 | NotYet | lowering | reading byteOrderMarkIndicator |
| kept | src/compiler/sys.ts:185:81 | Refused | lowering | a spread after the first field |
| kept | src/compiler/sys.ts:1869:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1870:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1967:12 | Refused | refusal_scan | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| kept | src/compiler/sys.ts:201:79 | NotYet | lowering | a PostfixUnaryExpression |
| kept | src/compiler/sys.ts:207:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:216:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:226:36 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:240:17 | NotYet | lowering | assigning a field of a value |
| kept | src/compiler/sys.ts:263:11 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:288:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:289:23 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:296:61 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:308:64 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:310:19 | Refused | refusal_scan | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| kept | src/compiler/sys.ts:324:13 | Refused | refusal_scan | a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read |
| kept | src/compiler/sys.ts:356:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:368:9 | Refused | lowering | inherited library member push read as an own field |
| kept | src/compiler/sys.ts:436:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:441:29 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:442:33 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:444:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:445:54 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:447:34 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:450:59 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:521:17 | Refused | lowering | a cast the runtime can't check |
| kept | src/compiler/sys.ts:521:17 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:603:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:627:33 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:655:29 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:690:27 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:739:9 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:773:13 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:777:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:781:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:786:17 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:794:87 | Refused | refusal_scan | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| kept | src/compiler/sys.ts:813:117 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:838:13 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:881:34 | Refused | refusal_scan | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |
| kept | src/compiler/sys.ts:941:13 | NotYet | lowering | optional record property access |
| gone | src/compiler/sys.ts:1494:121 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1494:35 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1494:35 | NotYet | lowering | reading __filename |
| gone | src/compiler/sys.ts:1494:92 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1519:15 | NotYet | lowering | a function value that captures the variable its own initializer declares |
| gone | src/compiler/sys.ts:1534:13 | Refused | refusal_scan | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| gone | src/compiler/sys.ts:1535:13 | Refused | refusal_scan | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| gone | src/compiler/sys.ts:1578:21 | NotYet | lowering | node:fs.StatsBase.isFile through an optional call |
| gone | src/compiler/sys.ts:1588:42 | NotYet | lowering | a union of differently held members variable a function value captures |
| gone | src/compiler/sys.ts:1590:191 | NotYet | lowering | process.recordreplay as a value (use its supported process operation directly) |
| gone | src/compiler/sys.ts:1590:192 | NotYet | refusal_scan | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| gone | src/compiler/sys.ts:1605:32 | Refused | refusal_scan | a cast the runtime can't check |
| gone | src/compiler/sys.ts:1605:33 | NotYet | refusal_scan | node:process.Process.stdout used as a value |
| gone | src/compiler/sys.ts:1612:22 | Refused | refusal_scan | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| gone | src/compiler/sys.ts:1614:77 | Refused | refusal_scan | a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read |
| gone | src/compiler/sys.ts:1967:12 | NotYet | lowering | a literal method through a view that erases its receiver |
| gone | src/compiler/sys.ts:384:79 | NotYet | refusal_scan | a member typed any ((0)) related to string at the argument site |
| gone | src/compiler/sys.ts:470:9 | NotYet | lowering | a value of type any |
| gone | src/compiler/sys.ts:490:9 | NotYet | lowering | assigning to an Identifier |
| gone | src/compiler/sys.ts:496:37 | NotYet | lowering | reading pollScheduled |
| gone | src/compiler/sys.ts:497:9 | NotYet | lowering | assigning to an Identifier |
| gone | src/compiler/sys.ts:57:1 | NotYet | lowering | a function without a body |
| gone | src/compiler/sys.ts:57:1 | Refused | refusal_scan | an ambient function declaration |
| gone | src/compiler/sys.ts:58:1 | NotYet | lowering | a function without a body |
| gone | src/compiler/sys.ts:58:1 | Refused | refusal_scan | an ambient function declaration |
| new | src/compiler/sys.ts:490:9 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:496:37 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:497:9 | NotYet | lowering | a union of differently held members variable a function value captures |

## C-noinfer-a

| State | Main location | Kind | Phase | Exact reason |
| --- | --- | --- | --- | --- |
| kept | src/compiler/core.ts:2332:13 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/core.ts:2365:60 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/debug.ts:227:5 | NotYet | lowering | a checked predicate target whose check would read past the depth bound (HostDirectoryWatcher) |
| kept | src/compiler/debug.ts:252:33 | NotYet | lowering | a value of type never |
| kept | src/compiler/sys.ts:1007:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:1030:54 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1052:140 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1061:84 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1065:90 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1100:40 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1103:30 | NotYet | lowering | reading defaultFallbackPolling |
| kept | src/compiler/sys.ts:1125:17 | NotYet | lowering | reading fileSystemEntryExists |
| kept | src/compiler/sys.ts:1186:48 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1198:13 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read |
| kept | src/compiler/sys.ts:1214:25 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read |
| kept | src/compiler/sys.ts:1253:77 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/sys.ts:1269:89 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1277:17 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:1281:24 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:129:9 | NotYet | lowering | a computed field name |
| kept | src/compiler/sys.ts:1387:31 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:1388:5 | NotYet | lowering | replacing a represented method at runtime |
| kept | src/compiler/sys.ts:1388:5 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:141:10 | NotYet | lowering | a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it |
| kept | src/compiler/sys.ts:144:81 | Refused | refusal_scan | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace |
| kept | src/compiler/sys.ts:162:17 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:162:35 | Refused | refusal_scan | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| kept | src/compiler/sys.ts:1642:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1646:59 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:1647:18 | NotYet | lowering | reading inspector |
| kept | src/compiler/sys.ts:1647:32 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1647:42 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:29 | NotYet | lowering | new a PropertyAccessExpression |
| kept | src/compiler/sys.ts:1651:29 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:33 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:43 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1652:13 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1652:13 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1652:21 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1654:13 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1654:13 | NotYet | lowering | reading session |
| kept | src/compiler/sys.ts:1654:13 | Refused | refusal_scan | a checked view with a callable contract |
| kept | src/compiler/sys.ts:1654:21 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:25 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1690:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1692:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1692:31 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1694:29 | NotYet | lowering | node:fs.StatsBase.isDirectory through an optional call |
| kept | src/compiler/sys.ts:1703:25 | NotYet | lowering | node:fs.writeFileSync: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| kept | src/compiler/sys.ts:1706:21 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1706:21 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1706:23 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1724:41 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| kept | src/compiler/sys.ts:1724:41 | NotYet | lowering | reading __filename |
| kept | src/compiler/sys.ts:178:13 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:1818:24 | NotYet | lowering | reading byteOrderMarkIndicator |
| kept | src/compiler/sys.ts:185:81 | Refused | lowering | a spread after the first field |
| kept | src/compiler/sys.ts:1869:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1870:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1967:12 | Refused | refusal_scan | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| kept | src/compiler/sys.ts:201:79 | NotYet | lowering | a PostfixUnaryExpression |
| kept | src/compiler/sys.ts:207:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:216:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:226:36 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:240:17 | NotYet | lowering | assigning a field of a value |
| kept | src/compiler/sys.ts:263:11 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:288:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:289:23 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:296:61 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:308:64 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:310:19 | Refused | refusal_scan | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| kept | src/compiler/sys.ts:324:13 | Refused | refusal_scan | a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read |
| kept | src/compiler/sys.ts:356:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:368:9 | Refused | lowering | inherited library member push read as an own field |
| kept | src/compiler/sys.ts:436:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:441:29 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:442:33 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:444:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:445:54 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:447:34 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:450:59 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:521:17 | Refused | lowering | a cast the runtime can't check |
| kept | src/compiler/sys.ts:521:17 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:57:1 | NotYet | lowering | a function without a body |
| kept | src/compiler/sys.ts:57:1 | Refused | refusal_scan | an ambient function declaration |
| kept | src/compiler/sys.ts:58:1 | NotYet | lowering | a function without a body |
| kept | src/compiler/sys.ts:58:1 | Refused | refusal_scan | an ambient function declaration |
| kept | src/compiler/sys.ts:603:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:627:33 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:655:29 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:690:27 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:739:9 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:773:13 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:777:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:781:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:786:17 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:794:87 | Refused | refusal_scan | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| kept | src/compiler/sys.ts:813:117 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:838:13 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:881:34 | Refused | refusal_scan | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |
| kept | src/compiler/sys.ts:941:13 | NotYet | lowering | optional record property access |
| gone | src/compiler/sys.ts:1494:121 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1494:35 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1494:35 | NotYet | lowering | reading __filename |
| gone | src/compiler/sys.ts:1494:92 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| gone | src/compiler/sys.ts:1519:15 | NotYet | lowering | a function value that captures the variable its own initializer declares |
| gone | src/compiler/sys.ts:1534:13 | Refused | refusal_scan | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| gone | src/compiler/sys.ts:1535:13 | Refused | refusal_scan | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| gone | src/compiler/sys.ts:1578:21 | NotYet | lowering | node:fs.StatsBase.isFile through an optional call |
| gone | src/compiler/sys.ts:1588:42 | NotYet | lowering | a union of differently held members variable a function value captures |
| gone | src/compiler/sys.ts:1590:191 | NotYet | lowering | process.recordreplay as a value (use its supported process operation directly) |
| gone | src/compiler/sys.ts:1590:192 | NotYet | refusal_scan | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| gone | src/compiler/sys.ts:1605:32 | Refused | refusal_scan | a cast the runtime can't check |
| gone | src/compiler/sys.ts:1605:33 | NotYet | refusal_scan | node:process.Process.stdout used as a value |
| gone | src/compiler/sys.ts:1612:22 | Refused | refusal_scan | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| gone | src/compiler/sys.ts:1614:77 | Refused | refusal_scan | a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read |
| gone | src/compiler/sys.ts:1967:12 | NotYet | lowering | a literal method through a view that erases its receiver |
| gone | src/compiler/sys.ts:384:79 | NotYet | refusal_scan | a member typed any ((0)) related to string at the argument site |
| gone | src/compiler/sys.ts:470:9 | NotYet | lowering | a value of type any |
| gone | src/compiler/sys.ts:490:9 | NotYet | lowering | assigning to an Identifier |
| gone | src/compiler/sys.ts:496:37 | NotYet | lowering | reading pollScheduled |
| gone | src/compiler/sys.ts:497:9 | NotYet | lowering | assigning to an Identifier |
| new | src/compiler/sys.ts:490:9 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:496:37 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:497:25 | NotYet | lowering | a call through a function value whose rest parameter isn't an array |

## C-noinfer-b

| State | Main location | Kind | Phase | Exact reason |
| --- | --- | --- | --- | --- |
| kept | src/compiler/core.ts:2332:13 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/core.ts:2365:60 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/debug.ts:227:5 | NotYet | lowering | a checked predicate target whose check would read past the depth bound (HostDirectoryWatcher) |
| kept | src/compiler/debug.ts:252:33 | NotYet | lowering | a value of type never |
| kept | src/compiler/sys.ts:1007:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:1030:54 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1052:140 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1061:84 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1065:90 | NotYet | lowering | reading getModifiedTime |
| kept | src/compiler/sys.ts:1100:40 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1103:30 | NotYet | lowering | reading defaultFallbackPolling |
| kept | src/compiler/sys.ts:1125:17 | NotYet | lowering | reading fileSystemEntryExists |
| kept | src/compiler/sys.ts:1186:48 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:1198:13 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read |
| kept | src/compiler/sys.ts:1214:25 | Refused | refusal_scan | a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read |
| kept | src/compiler/sys.ts:1253:77 | NotYet | lowering | function identity comparison in a program with specialized generic function values |
| kept | src/compiler/sys.ts:1269:89 | NotYet | lowering | a boolean \| undefined variable a function value captures |
| kept | src/compiler/sys.ts:1277:17 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:1281:24 | NotYet | lowering | reading presentWatcher |
| kept | src/compiler/sys.ts:129:9 | NotYet | lowering | a computed field name |
| kept | src/compiler/sys.ts:1387:31 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:1388:5 | NotYet | lowering | replacing a represented method at runtime |
| kept | src/compiler/sys.ts:1388:5 | Refused | refusal_scan | a method read as a value (writeFile would lose its object, and this with it) |
| kept | src/compiler/sys.ts:141:10 | NotYet | lowering | a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it |
| kept | src/compiler/sys.ts:144:81 | Refused | refusal_scan | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace |
| kept | src/compiler/sys.ts:1494:121 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| kept | src/compiler/sys.ts:1494:35 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| kept | src/compiler/sys.ts:1494:35 | NotYet | lowering | reading __filename |
| kept | src/compiler/sys.ts:1494:92 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| kept | src/compiler/sys.ts:1519:15 | NotYet | lowering | a function value that captures the variable its own initializer declares |
| kept | src/compiler/sys.ts:1534:13 | Refused | refusal_scan | a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| kept | src/compiler/sys.ts:1535:13 | Refused | refusal_scan | a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| kept | src/compiler/sys.ts:1578:21 | NotYet | lowering | node:fs.StatsBase.isFile through an optional call |
| kept | src/compiler/sys.ts:1588:42 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1590:191 | NotYet | lowering | process.recordreplay as a value (use its supported process operation directly) |
| kept | src/compiler/sys.ts:1590:192 | NotYet | refusal_scan | a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports |
| kept | src/compiler/sys.ts:1605:32 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:1605:33 | NotYet | refusal_scan | node:process.Process.stdout used as a value |
| kept | src/compiler/sys.ts:1612:22 | Refused | refusal_scan | a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read |
| kept | src/compiler/sys.ts:1614:77 | Refused | refusal_scan | a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read |
| kept | src/compiler/sys.ts:162:17 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:162:35 | Refused | refusal_scan | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields |
| kept | src/compiler/sys.ts:1642:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1646:59 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:1647:18 | NotYet | lowering | reading inspector |
| kept | src/compiler/sys.ts:1647:32 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1647:42 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:29 | NotYet | lowering | new a PropertyAccessExpression |
| kept | src/compiler/sys.ts:1651:29 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:33 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1651:43 | NotYet | refusal_scan | node:inspector.Session |
| kept | src/compiler/sys.ts:1652:13 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1652:13 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1652:21 | NotYet | refusal_scan | node:inspector.Session.connect |
| kept | src/compiler/sys.ts:1654:13 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1654:13 | NotYet | lowering | reading session |
| kept | src/compiler/sys.ts:1654:13 | Refused | refusal_scan | a checked view with a callable contract |
| kept | src/compiler/sys.ts:1654:21 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1655:25 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1690:17 | NotYet | lowering | a union of differently held members variable a function value captures |
| kept | src/compiler/sys.ts:1692:17 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1692:31 | NotYet | refusal_scan | node:inspector.Session.post |
| kept | src/compiler/sys.ts:1694:29 | NotYet | lowering | node:fs.StatsBase.isDirectory through an optional call |
| kept | src/compiler/sys.ts:1703:25 | NotYet | lowering | node:fs.writeFileSync: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| kept | src/compiler/sys.ts:1706:21 | NotYet | lowering | a method of a class stage 0 doesn't have |
| kept | src/compiler/sys.ts:1706:21 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1706:23 | NotYet | refusal_scan | node:inspector.Session.disconnect |
| kept | src/compiler/sys.ts:1724:41 | NotYet | refusal_scan | node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) |
| kept | src/compiler/sys.ts:1724:41 | NotYet | lowering | reading __filename |
| kept | src/compiler/sys.ts:178:13 | NotYet | lowering | assigning an element of a value |
| kept | src/compiler/sys.ts:1818:24 | NotYet | lowering | reading byteOrderMarkIndicator |
| kept | src/compiler/sys.ts:185:81 | Refused | lowering | a spread after the first field |
| kept | src/compiler/sys.ts:1869:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1870:17 | Refused | lowering | sort without a comparator |
| kept | src/compiler/sys.ts:1967:12 | NotYet | lowering | a literal method through a view that erases its receiver |
| kept | src/compiler/sys.ts:1967:12 | Refused | refusal_scan | a method read as a value (getEnvironmentVariable would lose its object, and this with it) |
| kept | src/compiler/sys.ts:201:79 | NotYet | lowering | a PostfixUnaryExpression |
| kept | src/compiler/sys.ts:207:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:216:32 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:226:36 | Refused | refusal_scan | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold |
| kept | src/compiler/sys.ts:240:17 | NotYet | lowering | assigning a field of a value |
| kept | src/compiler/sys.ts:263:11 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:288:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:289:23 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:296:61 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:308:64 | NotYet | lowering | a value of type PollingIntervalQueue |
| kept | src/compiler/sys.ts:310:19 | Refused | refusal_scan | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read |
| kept | src/compiler/sys.ts:324:13 | Refused | refusal_scan | a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read |
| kept | src/compiler/sys.ts:356:14 | NotYet | lowering | a function returning PollingIntervalQueue |
| kept | src/compiler/sys.ts:368:9 | Refused | lowering | inherited library member push read as an own field |
| kept | src/compiler/sys.ts:436:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:441:29 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:442:33 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:444:25 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:445:54 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:447:34 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:450:59 | NotYet | lowering | reading currentModifiedTime |
| kept | src/compiler/sys.ts:521:17 | Refused | lowering | a cast the runtime can't check |
| kept | src/compiler/sys.ts:521:17 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:603:5 | Refused | lowering | a method in object destructuring |
| kept | src/compiler/sys.ts:627:33 | Refused | refusal_scan | a cast the runtime can't check |
| kept | src/compiler/sys.ts:655:29 | NotYet | lowering | optional record property access |
| kept | src/compiler/sys.ts:690:27 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:739:9 | NotYet | lowering | a call through ?. (an optional call) |
| kept | src/compiler/sys.ts:773:13 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:777:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:781:9 | NotYet | lowering | assigning to an Identifier |
| kept | src/compiler/sys.ts:786:17 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:794:87 | Refused | refusal_scan | a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read |
| kept | src/compiler/sys.ts:813:117 | NotYet | lowering | reading timerToUpdateChildWatches |
| kept | src/compiler/sys.ts:838:13 | NotYet | lowering | a value of type any |
| kept | src/compiler/sys.ts:881:34 | Refused | refusal_scan | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read |
| kept | src/compiler/sys.ts:941:13 | NotYet | lowering | optional record property access |
| gone | src/compiler/sys.ts:384:79 | NotYet | refusal_scan | a member typed any ((0)) related to string at the argument site |
| gone | src/compiler/sys.ts:470:9 | NotYet | lowering | a value of type any |
| gone | src/compiler/sys.ts:490:9 | NotYet | lowering | assigning to an Identifier |
| gone | src/compiler/sys.ts:496:37 | NotYet | lowering | reading pollScheduled |
| gone | src/compiler/sys.ts:497:9 | NotYet | lowering | assigning to an Identifier |
| gone | src/compiler/sys.ts:57:1 | NotYet | lowering | a function without a body |
| gone | src/compiler/sys.ts:57:1 | Refused | refusal_scan | an ambient function declaration |
| gone | src/compiler/sys.ts:58:1 | NotYet | lowering | a function without a body |
| gone | src/compiler/sys.ts:58:1 | Refused | refusal_scan | an ambient function declaration |
| new | src/compiler/sys.ts:1501:13 | NotYet | lowering | reading setTimeout |
| new | src/compiler/sys.ts:1501:13 | Refused | refusal_scan | a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| new | src/compiler/sys.ts:1502:13 | Refused | refusal_scan | a function taking string \| number \| Timeout \| undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| new | src/compiler/sys.ts:1519:36 | NotYet | refusal_scan | a member typed any (require()module) related to {} at the initializer site |
| new | src/compiler/sys.ts:1599:13 | Refused | refusal_scan | a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| new | src/compiler/sys.ts:1600:13 | Refused | refusal_scan | a function taking string \| number \| Timeout \| undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take |
| new | src/compiler/sys.ts:490:9 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:496:37 | NotYet | lowering | a union of differently held members variable a function value captures |
| new | src/compiler/sys.ts:497:25 | NotYet | lowering | a call through a function value whose rest parameter isn't an array |
