# Scout: roadmap step 29 — the Node host used by tsc

Scouting for #96rpfw2, 2026-10-08. No implementation, adaptation or production changes. Base: `origin/area/runtime` at `cdfa22555589194e6f3133d986dd110061aafaa9`; branch `runtime/scout-29-host`.

## Result and ownership

Existing native entry points are declared in [prelude.d.ts](../internal/load/prelude.d.ts), lowered in [input.go](../internal/lower/input.go), emitted in [emit_expressions.go](../internal/native/emit_expressions.go), and implemented in [input.c](../internal/native/runtime/input.c) and [directory.c](../internal/native/runtime/directory.c).

The shipping native surface is the `'adamic'` file/argument API and one-string console output. **There are no shipping Node `fs`, `path`, `process`, `os`, Buffer, performance or inspector bindings in this base.** Existing file/status/directory/argv primitives are useful partial implementations, not working imports of Node modules. A scratch native-source probe importing `readFileSync` from `node:fs` stops at checker error TS2591. `internal/lower/optional_node_host.go` still returns false for the future Node-host argument-consumption hook.

First unblock file reads plus Buffer decoding, then file classification/directory enumeration, startup cwd/identity/argv and exit/output behavior. Emit needs writable-fd/directory operations. Diagnostics makes monotonic timing significant. Watch, tracing and CPU profiling remain separate optional surfaces. Pure tsc path functions in `src/compiler/path.ts` are not Node's `path` module and do not establish native `path.resolve` support.

Status definitions: **supported** means the called form lowers/emits natively today; **partly** means named existing native building blocks cover part of the operation but the Node/tsc contract is absent; **missing** means that contract has no native public implementation. No partly row asserts that the original Node call builds. Error shape, encodings, return values and resource ownership matter as well as the syscall.

Primary owners are proposed implementation boundaries: **library** owns Node declarations/bindings, Buffer/encoding/hash/path behavior and adaptation to existing primitives; **platforms** owns absent OS filesystem/cwd/environment/terminal/watch/profiler capabilities and platform-specific behavior; **runtime** owns callback scheduling/timers, shutdown/resource ordering, timing state and memory metrics. A library-owned partly row can still require platform work; these are not assertions of current team assignments. Language-lowering roots in the census are prerequisites, not automatically host-operation bugs.

## Sources and measurement

- Upstream TypeScript 6.0.3, source commit `050880ce59e30b356b686bd3144efe24f875ebc8`: the vendored `cohere/TypeScript/tsc/testdata/fixtures/compiler` sources, plus the pinned `src/tsc/tsc.ts` fetched into scratch. That CLI installs a logging host, optionally enables debug info/source-map support, calls `sys.setBlocking`, then `executeCommandLine(sys, noop, sys.args)`.
- The source census has **196 System access sites**, 33 distinct accessed members out of 44 declared, and **126 Node-derived accesses**. These overlap: do not add them. `stage3/census/data/host_sites.json`, `node_derived_sites.json`, `system_contract.json` and `host_summary.json` retain sites and receiver/type identity. Buffer indexing, bare globals, Date and watcher disposal were checked directly in source in addition to those ledgers.
- Named-fetch table branch `origin/codex/stage3-notyet-table` at `e8c283b5ed32477805357b652a170b85a04b2469`; its [morning after roots](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/after/roots.csv) use compiler scratch merge `0e5661e4243af95ae3247d066de14ccb2b582a94` and 81 adapted file records. Root sites below use **adapted** locations; API inventory sites use **upstream** locations. Neither census is a native tsc success claim.
- Fresh actual CLI observations use npm `typescript@6.0.3`, `lib/tsc.js` → `lib/_tsc.js`, Node **v24.19.0**, with a scratch preload observer. The downloaded tarball SHA-256 is 33cd0ee1beaa8c9e9d15a9da836c62ddea4c34a42d7c2d349dbc80d94165d22a. The npm launcher also calls optional `node:module.enableCompileCache`; that is included separately, as it is not in upstream `src/tsc/tsc.ts`.
- Four isolated projects have one `input.ts` and `tsconfig.json`, `{target:"ES2020", strict:true}`, `include:["input.ts"]`. Good input: `const count: number = 3; const names: string[] = ["tsc"];`. **No emit** adds `noEmit:true`; **Emit** sets `outDir:"out"`, initially absent; **Diagnostics** adds `--extendedDiagnostics` to no-emit; **Error** uses `const count: number = "wrong";` with no-emit. Each command is `node --require observe.cjs .../lib/tsc.js --project .../tsconfig.json [--extendedDiagnostics]`. Observer methods retain receivers, arguments, results, errors and function properties.
- Ranking counts are **direct calls from CLI/compiler bundle frames**, excluding Node loader/internal calls. Direct means the first non-observer stack frame is `tsc.js` or `_tsc.js`. Required module-loader requests are counted separately, since Node's require implementation introduces frames. No source capability/property read is falsely counted as a function call. Each project runs in a fresh process; calls that fail (such as emit's first open before mkdir) count as attempts. Instrumentation is for call counts, not performance measurement.
- Observed/uninstrumented stdout, stderr and exit match byte for byte for no-emit, emit and the error diagnostic; exit codes are 0, 0 and 2. Diagnostics exits 0 and produces its expected statistics; memory/timing statistics vary under observation and are not byte-held. No watch, CPU-profile, build/incremental, tracing, Windows or macOS path was dynamically measured. A zero below means not observed in these workloads, not unreachable globally.
- Prior `stage3/census/data/host_runtime.json` measured two cold **createProgram noEmit** workloads, not CLI executions: each did 53 reads and four stats. The ranking below uses the fresh CLI measurements instead.

## Missing/partial APIs ranked by CLI calls

Rank by no-emit direct calls, then emit/diagnostics/error counts, with ties lexical. Every row below still lacks its Node contract. User-timing clock counts are shown separately by mode so diagnostics does not inflate ordinary CLI priority. Associated property reads and prerequisite module imports follow the table.

| Node API | No emit | Emit | Diagnostics | Error | Native status | Owner |
| --- | ---: | ---: | ---: | ---: | --- | --- |
| `Buffer.toString` | 53 | 53 | 53 | 53 | missing | library |
| `fs.readFileSync` | 53 | 53 | 53 | 53 | partly | library |
| `fs.statSync` | 3 | 6 | 3 | 3 | partly | library |
| `Dirent.isFile` | 2 | 2 | 2 | 2 | partly | library |
| `Dirent.isSymbolicLink` | 2 | 2 | 2 | 2 | partly | library |
| `process.cwd` | 2 | 2 | 2 | 2 | missing | platforms |
| `Stats.isDirectory` | 1 | 2 | 1 | 1 | partly | library |
| `performance.now` | 1 | 1 | 423 | 1 | missing | runtime |
| `Stats.isFile` | 1 | 1 | 1 | 1 | partly | library |
| `fs.readdirSync` | 1 | 1 | 1 | 1 | partly | library |
| `fs.realpathSync.native` | 1 | 1 | 1 | 1 | missing | platforms |
| `module.enableCompileCache` | 1 | 1 | 1 | 1 | missing | library |
| `os.platform` | 1 | 1 | 1 | 1 | missing | platforms |
| `process.exit` | 1 | 1 | 1 | 1 | missing | runtime |
| `process.stdout._handle.setBlocking` | 1 | 1 | 1 | 1 | missing | platforms |
| `fs.openSync` | 0 | 2 | 0 | 0 | partly | library |
| `fs.closeSync` | 0 | 1 | 0 | 0 | partly | library |
| `fs.mkdirSync` | 0 | 1 | 0 | 0 | missing | platforms |
| `fs.writeSync` | 0 | 1 | 0 | 0 | partly | library |
| `process.stdout.write` | 0 | 0 | 24 | 1 | partly | library |
| `process.memoryUsage` | 0 | 0 | 1 | 0 | missing | runtime |

The default no-emit path’s two dominant rows are one operation chain: `fs.readFileSync` returns a Buffer which `sys.readFile` decodes with `Buffer.toString`. Do not add them as independent file reads. `fs.statSync` drives `fileExists`, `directoryExists`, size and mtime; the Stats/Dirent methods below it classify the results. Emit opens twice because the first attempt precedes creating the missing output directory, then retries; there is one successful file write. No-emit also makes one request each for `perf_hooks`, `fs`, `path`, `os` and `crypto`. They are loader prerequisites, not zero-call optional APIs merely because their binding methods have few calls.

For reference, Node internally implements the 53 no-emit file reads with 53 calls each to `fs.openSync`, `fs.readSync` and `fs.closeSync`. Those are **indirect Node implementation calls**, excluded from the direct ranking and not extra tsc source call sites. No native public readSync/fd API is exposed; `runtime/input.c` has private file reads. `path.dirname`/`resolve` seen inside the Node loader are also excluded. Buffer construction done internally by fs is not a direct `Buffer.from` call from tsc.

Unobserved optional APIs are not ordered by invented dynamic counts: source watch paths use `fs.watchFile`/`unwatchFile`, `fs.watch`, watcher `close`, timers and Date mtimes; profiling uses inspector sessions and fs/path operations; tracing uses fs writes; source maps use `source-map-support`; build can use crypto hashes, mtime updates and deletion. Their source sites and status are inventoried next. Timer/nextTick dispatch and filesystem watchers require callbacks/resource lifetimes, not just type declarations. `process.nextTick` is feature-tested in `core.ts:2592`, not actually called there.

## Complete source host API inventory

Paths in the Sites column are under `src/compiler/`; implementation shorthand `runtime/`, `native/`, `lower/` and `load/` refers to the corresponding `internal/` directories. Sites list every deduplicated Node-derived ledger observation; direct-source supplements group the specific forms noted above. Contract-returned data fields are listed as fields, not counted calls. No socket/network/HTTP, child_process, worker_threads, Node promises-fs or process.hrtime call is present in this scoped compiler/sys/CLI source inventory.

| Node API or host field | Upstream source sites | Native status | Owner | Existing implementation / remaining contract |
| --- | --- | --- | --- | --- |
| `Buffer.from` | `sys.ts:1614:36`<br>`sys.ts:1615:36` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Buffer.length / indexed byte reads and writes` | `sys.ts:1795`<br>`sys.ts:1796`<br>`sys.ts:1802`<br>`sys.ts:1803` | partly | library | Native Uint8Array byte storage/indexing in `runtime/typed_array.c`; no Node Buffer constructor/subclass identity. |
| `Buffer.toString` | `sys.ts:1614:36`<br>`sys.ts:1615:36`<br>`sys.ts:1805:24`<br>`sys.ts:1809:24`<br>`sys.ts:1813:24`<br>`sys.ts:1816:20` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Date constructor` | `sys.ts:107`<br>`sys.ts:1699`<br>`tsbuildPublic.ts:134`<br>`tsbuildPublic.ts:206`<br>`watch.ts:179` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `Date numeric conversion / valueOf` | `sys.ts:1749`<br>`sys.ts:1750`<br>`sys.ts:1761` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Date.getTime` | `sys.ts:1365`<br>`sys.ts:435`<br>`sys.ts:540`<br>`sys.ts:541` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Date.now` | `performanceCore.ts:102` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `Date.toISOString` | `sys.ts:1699` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Date.toLocaleTimeString` | `watch.ts:179`<br>`watch.ts:185` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `Dirent.isDirectory` | `sys.ts:1869` | partly | library | `fileStatus().type === "directory"`; no readdir Dirent method binding. |
| `Dirent.isFile` | `sys.ts:1866` | partly | library | `fileStatus().type === "file"`; no readdir Dirent method binding. |
| `Dirent.isSymbolicLink` | `sys.ts:1854:55` | partly | library | `fileStatus().symbolicLink` probes paths separately; no readdir Dirent classification. |
| `Dirent.name` | `sys.ts:1845:73` | partly | library | `readDirectory().names`; no Dirent object. |
| `FSWatcher.close / tsc FileWatcher.close` | `sys.ts:1779`<br>`watchUtilities.ts:419` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `Hash.digest` | `sys.ts:1951:20` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Hash.update` | `sys.ts:1950:13` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Stats.isDirectory` | `sys.ts:1698:29`<br>`sys.ts:1895:28` | partly | library | `fileStatus().type === "directory"`; no Stats object method binding. |
| `Stats.isFile` | `sys.ts:1582:21`<br>`sys.ts:1893:28` | partly | library | `fileStatus().type === "file"`; no Stats object method binding. |
| `Stats.mtime` | `sys.ts:1749:46`<br>`sys.ts:1750:22`<br>`sys.ts:1761:27`<br>`sys.ts:1761:43`<br>`sys.ts:1768:47`<br>`sys.ts:1927:20` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `Stats.size` | `sys.ts:1583:28` | partly | library | `fileStatus().size`; no Node Stats field binding. |
| `__dirname` | `sys.ts:1498` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `__filename` | `sys.ts:1498` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `clearTimeout` | `sys.ts:1506`<br>`sys.ts:1604`<br>`sys.ts:767` | missing | runtime | No native declaration/binding or equivalent public operation in this base. |
| `console.log` | `debug.ts:865:16` | supported | runtime | One string argument: `load/prelude.d.ts:4`, `lower/prelude.go`, `native/emit_statements.go:62`, `runtime/adamic.c`; native scratch probe prints `host-scout`. |
| `crypto.createHash` | `sys.ts:1949:26` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `fs.closeSync` | `sys.ts:1833:21`<br>`tracing.ts:113:9`<br>`tracing.ts:345:9` | partly | library | `runtime/input.c` closes its private fds; no exposed Node fd API. |
| `fs.existsSync` | `tracing.ts:78:14` | partly | library | `fileStatus` provides existence/error building blocks; missing exact no-throw Node binding. |
| `fs.mkdirSync` | `sys.ts:1550:25`<br>`sys.ts:1702:29`<br>`tracing.ts:79:13` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.openSync` | `sys.ts:1828:22`<br>`tracing.ts:221:25`<br>`tracing.ts:94:19` | partly | library | `runtime/input.c` privately opens reads/writes; no exposed Node fd/options/error object API. |
| `fs.readFileSync` | `sys.ts:1790:26` | partly | library | `readTextFile` → `runtime/input.c:296`: UTF-8 text only, result-valued errors; missing raw Buffer, UTF-16/BOM dispatch and Node errors. |
| `fs.readdirSync` | `sys.ts:1840:33` | partly | library | `readDirectory` → `runtime/directory.c:67` gives sorted decoded names; no `{withFileTypes:true}` Dirent objects. |
| `fs.realpathSync` | `sys.ts:1491:116`<br>`sys.ts:1491:142`<br>`sys.ts:1491:30`<br>`sys.ts:1914:40`<br>`sys.ts:1914:72` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.realpathSync.native` | `sys.ts:1491:116`<br>`sys.ts:1491:30`<br>`sys.ts:1914:40` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.statSync` | `sys.ts:1632:24` | partly | library | `fileStatus` → `runtime/directory.c:138` follows links and supplies type/size; no Stats object, mtime/Date or `throwIfNoEntry` contract. |
| `fs.unlinkSync` | `sys.ts:1941:24` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.unwatchFile` | `sys.ts:1743:30` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.utimesSync` | `sys.ts:1932:17` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.watch` | `sys.ts:1779:20` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.watchFile` | `sys.ts:1740:13` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `fs.writeFileSync` | `sys.ts:1707:25`<br>`tracing.ts:356:9` | partly | library | `writeTextFile` → `runtime/input.c:370`: UTF-8 whole-file/truncate path, result-valued errors; no Node Buffer/fd/options contract. |
| `fs.writeSync` | `sys.ts:1829:17`<br>`tracing.ts:112:9`<br>`tracing.ts:191:9`<br>`tracing.ts:192:21`<br>`tracing.ts:193:19`<br>`tracing.ts:194:9`<br>`tracing.ts:226:9`<br>`tracing.ts:337:13`<br>`tracing.ts:339:17`<br>`tracing.ts:343:9`<br>`tracing.ts:99:9` | partly | library | `writeTextFile` writes a complete string; no caller-owned fd, offset/position or partial-write API. |
| `fsRealpath alias of fs.realpathSync(.native)` | `sys.ts:1919` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `global.gc` | `sys.ts:1575:21`<br>`sys.ts:1576:21` | missing | runtime | No collector entry point; RC/regions have different lifetime semantics. tsc feature-tests this optional function. |
| `inspector.CallFrame.url` | `sys.ts:1679:21`<br>`sys.ts:1680:50`<br>`sys.ts:1682:25`<br>`sys.ts:1685:25` | missing | library | No V8 inspector/session/profile protocol; native profiling is a separate platform capability, not a silent V8-compatible stub. |
| `inspector.Profile.nodes` | `sys.ts:1678:32` | missing | library | No V8 inspector/session/profile protocol; native profiling is a separate platform capability, not a silent V8-compatible stub. |
| `inspector.ProfileNode.callFrame` | `sys.ts:1679:21`<br>`sys.ts:1680:50`<br>`sys.ts:1682:25`<br>`sys.ts:1685:25` | missing | library | No V8 inspector/session/profile protocol; native profiling is a separate platform capability, not a silent V8-compatible stub. |
| `inspector.Session` | `sys.ts:1651:32`<br>`sys.ts:1655:33` | missing | platforms | No V8 inspector/session/profile protocol; native profiling is a separate platform capability, not a silent V8-compatible stub. |
| `inspector.Session.connect` | `sys.ts:1656:13` | missing | platforms | No V8 inspector/session/profile protocol; native profiling is a separate platform capability, not a silent V8-compatible stub. |
| `inspector.Session.disconnect` | `sys.ts:1710:21` | missing | platforms | No V8 inspector/session/profile protocol; native profiling is a separate platform capability, not a silent V8-compatible stub. |
| `inspector.Session.post` | `sys.ts:1658:13`<br>`sys.ts:1659:17`<br>`sys.ts:1696:17` | missing | platforms | No V8 inspector/session/profile protocol; native profiling is a separate platform capability, not a silent V8-compatible stub. |
| `os.EOL` | `sys.ts:1525` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `os.platform` | `sys.ts:1489` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `path.dirname` | `sys.ts:1498:78`<br>`sys.ts:1675:52`<br>`sys.ts:1702:43` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `path.join` | `sys.ts:1498:67`<br>`sys.ts:1699:43` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `path.resolve` | `sys.ts:1541:34` | missing | library | No native declaration/binding or equivalent public operation in this base. |
| `performance.clearMarks` | `performanceCore.ts:80:16` | missing | runtime | No native clock/user-timing binding; requires runtime clock and mark/measure state. `timeOrigin` is wall-clock epoch metadata, `now` is monotonic. |
| `performance.clearMeasures` | `performanceCore.ts:81:16` | missing | runtime | No native clock/user-timing binding; requires runtime clock and mark/measure state. `timeOrigin` is wall-clock epoch metadata, `now` is monotonic. |
| `performance.mark` | `performanceCore.ts:78:16` | missing | runtime | No native clock/user-timing binding; requires runtime clock and mark/measure state. `timeOrigin` is wall-clock epoch metadata, `now` is monotonic. |
| `performance.measure` | `performanceCore.ts:79:16` | missing | runtime | No native clock/user-timing binding; requires runtime clock and mark/measure state. `timeOrigin` is wall-clock epoch metadata, `now` is monotonic. |
| `performance.now` | `performanceCore.ts:72:62` | missing | runtime | No native clock/user-timing binding; requires runtime clock and mark/measure state. `timeOrigin` is wall-clock epoch metadata, `now` is monotonic. |
| `performance.timeOrigin` | `performanceCore.ts:72:16` | missing | runtime | No native clock/user-timing binding; requires runtime clock and mark/measure state. `timeOrigin` is wall-clock epoch metadata, `now` is monotonic. |
| `process.argv` | `sys.ts:1524:19` | partly | library | `programArguments` → `runtime/input.c:174` supports argv after executable, fresh array; missing full Node argv/executable+script prefix. |
| `process.browser` | `core.ts:2593` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.cwd` | `sys.ts:1501:51`<br>`sys.ts:1883:96` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.env` | `sys.ts:1516:27`<br>`sys.ts:1517:38`<br>`sys.ts:1518:32`<br>`sys.ts:1566:24`<br>`sys.ts:1594:26`<br>`sys.ts:1594:62` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.env[name / TSC_WATCHFILE / TSC_NONPOLLING_WATCHER / TSC_WATCHDIRECTORY / NODE_INSPECTOR_IPC / VSCODE_INSPECTOR_OPTIONS / NODE_ENV]` | `sys.ts:1516`<br>`sys.ts:1517`<br>`sys.ts:1518`<br>`sys.ts:1566`<br>`sys.ts:1594` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.execArgv` | `sys.ts:1592:112`<br>`sys.ts:1592:68`<br>`sys.ts:1594:107` | missing | runtime | No native declaration/binding or equivalent public operation in this base. |
| `process.exit` | `sys.ts:1588:42` | missing | runtime | No native declaration/binding or equivalent public operation in this base. |
| `process.memoryUsage` | `sys.ts:1578:24` | missing | runtime | No public memory query; Adamic allocator counters are not V8 heapUsed and require an explicit native metric contract. |
| `process.memoryUsage().heapUsed` | `sys.ts:1578:24` | missing | runtime | No public memory query; Adamic allocator counters are not V8 heapUsed and require an explicit native metric contract. |
| `process.nextTick` | `core.ts:2592:14` | missing | runtime | No native declaration/binding or equivalent public operation in this base. |
| `process.pid` | `sys.ts:1699:120`<br>`tracing.ts:82:50`<br>`tracing.ts:83:39` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.platform` | `sys.ts:1484:25`<br>`sys.ts:1485:32`<br>`sys.ts:1491:56`<br>`sys.ts:1500:44` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.recordreplay` | `sys.ts:1594` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.stdout` | `sys.ts:1528:17`<br>`sys.ts:1531:24`<br>`sys.ts:1534:24`<br>`sys.ts:1606:17`<br>`sys.ts:1609:33` | missing | runtime | No native declaration/binding or equivalent public operation in this base. |
| `process.stdout._handle.setBlocking` | `sys.ts:1609`<br>`sys.ts:1611` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.stdout.columns` | `sys.ts:1531:24` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.stdout.isTTY` | `sys.ts:1534:24` | missing | platforms | No native declaration/binding or equivalent public operation in this base. |
| `process.stdout.write` | `sys.ts:1528:17`<br>`sys.ts:1606:17` | partly | library | `console.log/error` → `runtime/adamic.c` supports buffered string output with newline; no newline-free stream write/return/backpressure contract. |
| `require(crypto)` | `sys.ts:1476:23` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `require(fs)` | `sys.ts:1470:42`<br>`tracing.ts:63:22` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `require(inspector)` | `sys.ts:1650:59` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `require(modulePath)` | `sys.ts:1619` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `require(os)` | `sys.ts:1472:21` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `require(path)` | `sys.ts:1471:46` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `require(perf_hooks)` | `performanceCore.ts:35:37` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `require(source-map-support)` | `sys.ts:1597:22` | missing | library | No Node module loader/declarations; optional loading may be absent, but required fs/path/os startup imports need an explicit binding. |
| `setTimeout` | `sys.ts:1505`<br>`sys.ts:1603`<br>`sys.ts:770` | missing | runtime | No native declaration/binding or equivalent public operation in this base. |
| `source-map-support.install` | `sys.ts:1597` | missing | library | No native declaration/binding or equivalent public operation in this base. |

Additional Dirent classification detail: `sys.ts:1866–1870` calls `stat.isFile()`/`isDirectory()` on either Stats (symlink/fallback) or Dirent (ordinary entry). The ledger assigns those sites to its widened/any local, so they are not fully represented as Node-derived Dirent methods; both methods are partly covered by `fileStatus` classification and need library bindings. `module.enableCompileCache` is missing (library owner), optional npm `lib/tsc.js:3–5` only. `require` and CommonJS module exports are missing Node host/module-loader contracts; generated launch-wrapper `module.exports = require("./_tsc.js")` is not an upstream compiler API call. `__filename`/`__dirname` require executable/library-layout policy, not the source compiler’s current path embedded into every native binary.

Buffer decoding must preserve tsc’s specific forms: raw byte read; UTF-16BE byte-pair swap and `toString("utf16le",2)`; UTF-16LE BOM skip; UTF-8 BOM skip via `toString("utf8",3)`; default UTF-8 replacement decoding; base64 encode/decode. Existing `readTextFile` uses Node-compatible malformed-UTF-8 replacement but retains the BOM and cannot return raw Buffer/UTF-16 content. A string-only adapter silently changes these paths.

Watchers must preserve polling/native recursive differences, timestamps, event-kind transitions, callback scheduling, cancellation and disposal. `process.platform`/`os.platform` are not a universal Linux constant: sys chooses Windows realpath handling and Darwin/Windows recursive watch behavior, and probes filesystem case sensitivity. stdout includes newline-free writes, terminal columns/isTTY and the private optional setBlocking hook. Error swallowing/rethrow depends on Node error `.code` (notably EEXIST), while current Adamic I/O returns stable own-message result objects.

## System/CLI bridge, all 44 declared members

These are tsc’s own System methods/fields, not additional Node APIs. Static call-site counts are from the source census and are distinct from dynamic CLI counts; references/callback forwarding are not counted as calls. They retain the 11 declared-but-unaccessed members so optional capabilities are not silently lost.

| System member | Direct static call sites | Native status | Owner | Node backing / gap |
| --- | ---: | --- | --- | --- |
| `args` | 0 | partly | library | process.argv.slice(2); programArguments exists. |
| `newLine` | 0 | missing | platforms | os.EOL; no host EOL declaration. |
| `useCaseSensitiveFileNames` | 0 | missing | platforms | os/process platform and filesystem-case probe. |
| `write` | 21 | partly | library | process.stdout.write; existing console adds newline. |
| `writeOutputIsTTY` | 1 | missing | platforms | stdout.isTTY |
| `getWidthOfTerminal` | 2 | missing | platforms | stdout.columns |
| `readFile` | 4 | partly | library | readFileSync + Buffer BOM/encoding/error behavior. |
| `getFileSize` | 0 | partly | library | statSync().size; fileStatus size exists. |
| `writeFile` | 3 | partly | library | openSync/writeSync/closeSync + BOM; writeTextFile exists. |
| `watchFile` | 0 | missing | platforms | watchFile/unwatchFile, callbacks and polling Date. |
| `watchDirectory` | 0 | missing | platforms | fs.watch + watcher.close and recursive platform rules. |
| `preferNonRecursiveWatch` | 0 | missing | platforms | platform capability decision. |
| `resolvePath` | 0 | missing | library | path.resolve and cwd semantics. |
| `fileExists` | 6 | partly | library | statSync/isFile; fileStatus exists. |
| `directoryExists` | 5 | partly | library | statSync/isDirectory; fileStatus exists. |
| `createDirectory` | 3 | missing | platforms | mkdirSync and EEXIST policy. |
| `getExecutingFilePath` | 2 | missing | platforms | __filename/__dirname and sys.js fake lib directory policy. |
| `getCurrentDirectory` | 8 | missing | platforms | memoized process.cwd. |
| `getDirectories` | 2 | partly | library | readdirSync with Dirent, symlinks/stat; readDirectory/fileStatus exist. |
| `readDirectory` | 2 | partly | library | matchFiles/globs/realpath/case/depth atop directory primitives; not just readdir. |
| `getModifiedTime` | 1 | missing | platforms | Stats.mtime as Date. |
| `setModifiedTime` | 1 | missing | platforms | utimesSync sets atime and mtime. |
| `deleteFile` | 1 | missing | platforms | unlinkSync swallowing missing/error cases. |
| `createHash` | 0 | missing | library | crypto SHA256 or tsc djb2 fallback; SHA256 host path absent. |
| `createSHA256Hash` | 0 | missing | library | Must be cryptographic SHA256; hash-table hash is not a substitute. |
| `getMemoryUsage` | 1 | missing | runtime | optional global.gc + memoryUsage().heapUsed; native metric contract needed. |
| `exit` | 22 | missing | runtime | requested status after profiler continuation and output flush; normal native return/panic is not this API. |
| `enableCPUProfiler` | 2 | missing | platforms | inspector Session connect/post. |
| `disableCPUProfiler` | 0 | missing | platforms | post stop, serialize profile, disconnect, continuation. |
| `cpuProfilingEnabled` | 1 | missing | library | activeSession plus process.execArgv capability flags. |
| `realpath` | 1 | missing | platforms | realpathSync(.native), failure identity, Windows long-path dispatch. |
| `getEnvironmentVariable` | 12 | missing | platforms | process.env[name] or empty string. |
| `tryEnableSourceMapsForHost` | 0 | missing | library | optional source-map-support.install. |
| `getAccessibleFileSystemEntries` | 0 | partly | library | Dirent/symlink classification; readDirectory/fileStatus building blocks. |
| `debugMode` | 0 | missing | library | env/execArgv/recordreplay probes. |
| `setTimeout` | 0 | missing | runtime | schedule callback and preserve args/handle. |
| `clearTimeout` | 0 | missing | runtime | cancel timeout handle. |
| `clearScreen` | 1 | partly | library | stdout ANSI writes; native output exists but not newline-free Node stream API. |
| `setBlocking` | 0 | missing | platforms | optional stdout._handle.setBlocking. |
| `base64decode` | 0 | missing | library | Buffer.from(base64).toString(utf8). |
| `base64encode` | 0 | missing | library | Buffer.from(utf8).toString(base64). |
| `require` | 0 | missing | library | resolveJSModule plus runtime Node require/module result. |
| `now` | 1 | missing | platforms | test hook Date factory, not direct production getNodeSystem implementation. |
| `storeSignatureInfo` | 0 | missing | library | test flag; no getNodeSystem implementation/default. Not an OS syscall. |

The CLI entry’s logging callback delegates to System.write. It tests optional source maps under NODE_ENV=development and optional blocking output before calling executeCommandLine. It does not itself add an fs/path API. Watch/build hosts passed into compiler methods can supply their own capabilities; source-level optionality is not evidence of a native host implementation.

## Not-yet-table cross-check

The morning after table has the following own-diagnostic-file NotYet roots. These are **lowering roots on a checker-rejected adapted corpus**, not API dynamic counts and not exhaustive native host refusal counts:

| File | Root sites |
| --- | ---: |
| sys.ts | 96 |
| performanceCore.ts | 1 |
| performance.ts | 10 |
| tracing.ts | 35 |
| executeCommandLine.ts | 58 |
| src/tsc/tsc.ts | 4 |

Relevant exact roots (adapted positions):

| Root site | Exact reason | Scouting implication |
| --- | --- | --- |
| `src/compiler/performanceCore.ts:64:38` | a value of type any | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:108:46` | new an Identifier | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1469:5` | a function inside a function (a closure) | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1630:18` | a function returning any | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1829:22` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1830:17` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1834:21` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1888:19` | a value of type any | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1894:28` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1896:28` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1915:40` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1933:17` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1940:18` | a function returning any | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1950:19` | a value of type any | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1951:13` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/compiler/sys.ts:1952:20` | a call to a PropertyAccessExpression | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/tsc/tsc.ts:16:5` | reading ts | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/tsc/tsc.ts:20:5` | reading ts | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/tsc/tsc.ts:24:1` | reading ts | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |
| `src/tsc/tsc.ts:6:1` | reading ts | Language/module/type lowering stops before or at host dispatch; do not credit a native API from absence of a more specific refusal. |

For example, the syscall-looking PropertyAccessExpression roots coincide with sys write/open/close, stat predicates, realpath, utimes and hash operations; they do not establish that adding a declaration alone implements the backend. `getNodeSystem` is blocked by closure lowering, Date construction is blocked, the stat/hash helpers use any, and CLI reads of `ts` fail. Existing adaptations expose these roots without implementing Node OS behavior. The original host fixture ledger (`stage3/fixtures/host/README.md`) records all 25 fixtures checker-blocked and explicitly claims no native correctness; it is a useful future contract suite, not current support.

## Verification and reproducibility

Read AGENTS.md if present (none found), fetched both branches by exact names, ran `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`, sourced `/workspace/adamic-tools/env.sh`, and verified Node v24.19.0. Machine: Linux x86_64, AMD EPYC 7763, five visible CPUs/four-CPU quota. Setup and scouting logs are in scratch; no benchmark durations or implementation estimates are part of this scout.

Scratch evidence: `/tmp/scout-29-host/observe.cjs`, `source-audit.cjs`/`.json`, each mode’s `.json`/`.stdout`/`.stderr`, the pinned entry `tsc.ts`, package tarball and extracted CLI, and `node-probe.stderr`. These are not production files. The observer wraps fs/path/process/Buffer/performance/crypto and Stats/Dirent functions with Reflect.apply, preserves function properties, guards its own stack collection, records calls originating in the CLI/compiler bundle, and writes JSON with the original fs writer at process exit. Property probes, constructors and loader prerequisites are inventoried from source separately; unwrapped APIs are not assigned measured call counts. The use of stack attribution and only four Linux CLI projects limits what can be inferred about other workloads.

An independent TypeScript AST scan of the compiler sources found no additional host-root calls beyond the table and supplements. Independent checks cover all 126 Node-derived access observations after canonicalizing receiver names, all 44 System contract members, measured mode exit/output controls, recorded root totals and every referenced native source path. A native one-string console probe succeeds; the direct node:fs import fails at the checker boundary. The targeted `go test ./internal/oracle -run '^TestInputAgreesWithNode$/internal/oracle/testdata/(read_files|arguments|walk|write_files).a$' -count=1` passes, covering existing `'adamic'` support, not Node host support. No fixture is added or changed, so counts regeneration and new mutants are not applicable. No whole-package tests are required for this documentation-only scout.
