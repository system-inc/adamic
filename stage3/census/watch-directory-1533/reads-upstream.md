# recursive read gate — upstream TypeScript v6.0.3

Upstream commit: 050880ce59e30b356b686bd3144efe24f875ebc8.
All paths below are relative to the upstream root. Parameter declarations, comments, and explicit `false` call arguments are not reads of the incoming value.

| File | Line | Read | Effect for undefined and false |
| --- | ---: | --- | --- |
| src/compiler/sys.ts | 1108 | watchDirectory -> fsWatch argument | pass-through |
| src/compiler/sys.ts | 1126 | watchDirectory -> hostRecursiveDirectoryWatcher argument | pass-through |
| src/compiler/sys.ts | 623 | recursive ? createDirectoryWatcher : watchDirectory | chooses watchDirectory |
| src/compiler/sys.ts | 625 | recursive -> nonRecursiveWatchDirectory argument | pass-through |
| src/compiler/sys.ts | 1130 | Debug.assert(!recursive) | asserts true |
| src/compiler/sys.ts | 1160 | recursive -> fsWatch argument | pass-through |
| src/compiler/sys.ts | 1207 | recursive ? fsWatchesRecursive : fsWatches | chooses fsWatches |
| src/compiler/sys.ts | 1211 | recursive -> fsWatchHandlingExistenceOnHost argument in closure | pass-through; no comparison/storage key |
| src/compiler/sys.ts | 1264 | recursive -> fsWatchWorker or fsWatchWorkerHandlingTimestamp argument | pass-through |
| src/compiler/sys.ts | 1362 | recursive -> fsWatchWorker argument | pass-through |
| src/compiler/sys.ts | 1782 | recursive: !!recursive passed to Node fs.watch | exactly false |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts | 889 | TestServerHost.fsWatchWorker -> watchUtils.fsWatch argument | pass-through |
| src/harness/watchUtils.ts | 182 | recursive ? fsWatchesRecursive : fsWatches | chooses fsWatches |

The only concrete factory callers are getNodeSystem (sys.ts:1502) and TestServerHost (virtualFileSystemWithWatch.ts:462). The production worker normalizes with `!!` before crossing into Node. The test worker selects its map by truthiness and does not retain the incoming value. Missing-entry and polling fallback paths do not read recursive. True recursive mode generates explicit false arguments for child watchers at sys.ts:659; those are not incoming-value reads.

Gate: PASS for the concrete upstream graph; no default, nullish operation, equality test, raw key, or stored value later compared was found.
