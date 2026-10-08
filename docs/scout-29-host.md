# Scout: roadmap step 29 — the Node host used by tsc

Scouting for #96rpfw2, 2026-10-08. Expanded research, design and executable source-shape proof; no production changes. Base: `origin/area/runtime` at `cdfa22555589194e6f3133d986dd110061aafaa9`; branch `runtime/scout-29-host`.

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

An independent TypeScript AST scan of the compiler sources found no additional host-root calls beyond the table and supplements. Independent checks cover all 126 Node-derived access observations after canonicalizing receiver names, all 44 System contract members, measured mode exit/output controls, recorded root totals and every referenced native source path. A native one-string console probe succeeds; the direct node:fs import fails at the checker boundary. The targeted `go test ./internal/oracle -run '^TestInputAgreesWithNode$/internal/oracle/testdata/(read_files|arguments|walk|write_files).a$' -count=1` passes, covering existing `'adamic'` support, not Node host support. The original scout added no fixtures. The expanded work and its targeted verification are recorded below; no whole-package test was run.


## Expanded research: where step 29 bites

The tables above retain every API and exact source site. Here is the file-level distribution, including property probes rather than counting only calls. The host ledger includes overlapping System and host observations; its 283 observations are not 283 unique Node calls. The Node-derived column totals 126. Site counts are static, while the CLI ranking above counts dynamic calls.

| Upstream file | Host ledger observations | Node-derived observations |
| --- | ---: | ---: |
| `src/compiler/commandLineParser.ts` | 1 | 0 |
| `src/compiler/core.ts` | 2 | 1 |
| `src/compiler/debug.ts` | 0 | 1 |
| `src/compiler/executeCommandLine.ts` | 107 | 0 |
| `src/compiler/performance.ts` | 7 | 0 |
| `src/compiler/performanceCore.ts` | 8 | 6 |
| `src/compiler/program.ts` | 19 | 0 |
| `src/compiler/sys.ts` | 79 | 99 |
| `src/compiler/tracing.ts` | 3 | 19 |
| `src/compiler/tsbuildPublic.ts` | 11 | 0 |
| `src/compiler/watch.ts` | 44 | 0 |
| `src/compiler/watchPublic.ts` | 2 | 0 |

The CLI entry `src/tsc/tsc.ts:8,16,20,24` installs output, optionally enables source maps, sets stdout blocking and enters executeCommandLine. It is outside the compiler-ledger denominator. The largest host consumer is executeCommandLine; sys.ts contains the implementation boundary. `sys.ts:59–66` is the pure hash fallback selected when crypto is unavailable, not the usual Node SHA-256 path. `sys.ts:1787` reads raw bytes and decodes BOMs; `sys.ts:1819` writes through an explicitly opened descriptor; `sys.ts:1629,1886,1901,1905` classify entries. `performanceCore.ts` chooses the clock and `performance.ts` supplies marks/measures; tracing.ts additionally needs appendable descriptors and timestamped output. The detailed upstream sites and adapted not-yet roots above remain separate evidence sets.

### Node, specification, and TypeScript-go evidence

Node v24.19.0 is the executable oracle. The linked Node 24 API documentation describes the contract; it does not specify Adamic representations. ECMAScript specifies string and numeric semantics, not files, directories, environment, watchers or process shutdown. TypeScript-go is useful implementation evidence, not a substitute oracle. Its inspected source is pinned to `d92d9bfee114c80be2c375d72edae966176e3a4f` under `cohere/TypeScript/tsc`.

| Shape and tsc evidence | Node/spec contract and hard cases | TypeScript-go evidence |
| --- | --- | --- |
| Read/decode, `sys.ts:1787`; 53 reads on each measured CLI path | [fs.readFileSync](https://nodejs.org/docs/latest-v24.x/api/fs.html#fsreadfilesyncpath-options) returns Buffer without encoding; [Buffer encodings](https://nodejs.org/docs/latest-v24.x/api/buffer.html#buffers-and-character-encodings) replace invalid UTF-8 with U+FFFD. tsc strips UTF-8 BOM, handles UTF-16LE, swaps UTF-16BE pairs; empty, odd tails, lone surrogates, missing files and directories matter. Directory reads are platform-specific (FreeBSD differs). | `internal/vfs/internal/internal.go:131,156,175` reads bytes, strips BOMs and decodes UTF-16. The UTF-8 path returns an unsafe Go string over bytes, including invalid UTF-8: this cannot establish Node decoding equivalence or Adamic ownership safety. |
| File/directory existence and enumeration, `sys.ts:1629,1886,1901,1905,1909` | [Stats and Dirent](https://nodejs.org/docs/latest-v24.x/api/fs.html#class-fsstats) distinguish regular files/directories/symlinks. tsc follows links as needed, skips broken targets and sorts results. Non-regular files, races, permission errors and symlink loops require tests. | `internal/vfs/internal/internal.go:47,59,64,69` centralizes Stat and enumeration. FileExists uses `!IsDir()`, which includes some non-regular entries excluded by Node isFile; do not copy this difference. |
| Realpath, pathname construction, executable/library identity | [path.resolve](https://nodejs.org/docs/latest-v24.x/api/path.html#pathresolvepaths) uses platform rules and cwd; [realpathSync](https://nodejs.org/docs/latest-v24.x/api/fs.html#fsrealpathsyncpath-options) resolves links or throws. tsc catches failure and returns the input. Windows drives, UNC, case, separators and nonexistent components matter. | `internal/vfs/osvfs/os.go:116` uses platform realpath then absolute paths, returning the input on failure. `internal/tspath/path.go:83,263,622` handles compiler paths separately; `internal/nativepath/realpath_*.go` supplies OS specifics. |
| Write, mkdir, unlink, timestamps, `sys.ts:1466,1487,1819` | [fs synchronous APIs](https://nodejs.org/docs/latest-v24.x/api/fs.html) have observable open flags, error codes, close behavior and partial writes. tsc retries after creating parent directories and optionally prepends BOM. unlink is not recursive deletion. mtime is a Date; utimes uses both access and modification times. | `internal/vfs/osvfs/os.go:140,156,172,186` opens/truncates, ensures directories and changes times. Its Remove uses os.RemoveAll, which is broader than Node unlinkSync; it must not be adopted as the Node contract. |
| cwd, argv, environment, output and exit | [process](https://nodejs.org/docs/latest-v24.x/api/process.html) specifies argv, env, cwd, stdout and exit. Forced exit can truncate pending asynchronous output. tsc memoizes cwd, slices argv, normalizes absent/empty environment to empty string and requests blocking output where available. CLI status 0/1/2 is observable. | `cmd/tsc/sys.go:43,47,51,55,59,64,124` uses cached cwd, writers, TTY/width, LookupEnv and bundled libraries. `cmd/tsc/main.go:15,20,29,31` uses os.Exit, Go argv, signals and execute. Go argv layout differs from Node. |
| Clock, marks/measures, memory, optional GC | [performance.now/timeOrigin](https://nodejs.org/docs/latest-v24.x/api/perf_hooks.html) distinguish monotonic elapsed milliseconds from epoch origin. [process.memoryUsage](https://nodejs.org/docs/latest-v24.x/api/process.html#processmemoryusage) returns V8/OS-specific fields. Wall-clock changes and native allocations cannot be equated to V8 heapUsed. | `cmd/tsc/sys.go:27,31` separates time.Since(start) and time.Now. These are useful clock ingredients, not Node performance mark/measure or V8 memory compatibility. |
| Watch and timers, sys.ts watch factories | [fs.watch/watchFile](https://nodejs.org/docs/latest-v24.x/api/fs.html#fswatchfilename-options-listener) has OS-dependent delivery and missing filename cases; [timers](https://nodejs.org/docs/latest-v24.x/api/timers.html) owns cancellation and handles. Rename/replacement, deleted parents, polling fallbacks, unref and callbacks after cancellation matter. | `internal/fswatch/watcher.go:58,172,185,620` supplies native backends and close. Its WatchFile contract explicitly lacks automatic recovery after parent deletion; TS polling/fallback behavior needs separate work. |
| Hash fallback, `sys.ts:59–66`; crypto branch in sys.ts | [crypto hashes](https://nodejs.org/docs/latest-v24.x/api/crypto.html#cryptocreatehashalgorithm-options) implement SHA-256 and encoded digest. The fallback iterates [UTF-16 code units](https://tc39.es/ecma262/multipage/text-processing.html#sec-string.prototype.charcodeat), [left-shifts with signed 32-bit coercion](https://tc39.es/ecma262/multipage/ecmascript-language-expressions.html#sec-left-shift-operator), then adds as Number and calls [Number.toString](https://tc39.es/ecma262/multipage/numbers-and-dates.html#sec-number.prototype.tostring). It is not an unsigned djb2 variant. | `internal/execute/incremental/snapshot.go:32–36` uses xxh3 128-bit plus hex (optionally text); neither SHA-256 nor this fallback. Copying that algorithm would violate the Node oracle. |
| Trace/profiling, tracing.ts; CLI CPU-profile branches | [inspector.Session](https://nodejs.org/docs/latest-v24.x/api/inspector.html#class-inspectorsession) speaks V8 inspector protocol, including callback errors, connect/disconnect and profiler results. A native profiler cannot advertise that protocol merely because it measures CPU. | The Go CLI/system evidence above does not provide a V8 Session implementation. Watch backends and Go timing are not evidence of inspector compatibility. |

Additional contractual surfaces remain in the complete API inventory: [os](https://nodejs.org/docs/latest-v24.x/api/os.html) supplies platform/EOL/home/CPU queries, and [CommonJS modules](https://nodejs.org/docs/latest-v24.x/api/modules.html) defines require, __filename and __dirname. Their executable layout and platform identity are observable in tsc. Private stdout `_handle.setBlocking` is tsc's feature-probed Node implementation detail, not an ECMAScript requirement or a stable public API promise.

## Design within Adamic's rules

Keep the existing no-GC, retain/drop and region discipline. Node results are the oracle for each advertised call form. The design can fix internal lifetime requirements without choosing unresolved language-visible representations; those decisions are questions below.

Synchronous adapters borrow inputs only for the duration of the call and return owned immutable text or owned bytes. A string decoded from bytes retains its backing storage or copies it before the Buffer owner is released; TypeScript-go's unsafe.String shortcut is not sufficient. Buffer views must keep their common backing allocation alive. Native resources and allocations stay live through the last permitted use, including failures and cleanup; every successful acquisition has exactly one eventual release. A close is not permission to drop a callback environment that is still executing.

Descriptor implementations track acquired state internally and close on every completion/failure path. Hash objects retain their state and inputs until update/digest is done. Stat/Dirent results are snapshots with owned path/name data. Error translation preserves observable Node outcomes rather than mapping every OS error to absence: tsc itself chooses where to catch errors. These are implementation constraints; the public descriptor, error and Buffer types await the questions below.

For watchers and timers, registrations retain callback closures and captured data. Cancellation first stops new delivery, drains or accounts for in-flight calls, then releases retained state. Captured cycles use existing graph-region mechanisms; host registrations cannot rely on a collector or free the closure early. Event-loop, cancellation, callback typing and shutdown policy still need rulings. No fake watcher success, fake SHA digest, fake V8 memory metric or placeholder profiler result should pass capability probes.

Library owns declarations, exact Node call forms, decoding/path/hash adapters and the System bridge. Platforms owns syscalls, errors, file descriptors, timestamps, case/symlink behavior, terminal capability and backend watchers. Runtime owns safe retention, callback delivery, clock state and final shutdown ordering. Cross-platform fixture expectations must be selected from Node on the same OS; Linux green does not establish Windows drive or watch semantics.

After rulings, work should proceed in reviewable contract units: read/decode (fixtures 01–04), classification/enumeration (06–09,24–25), startup identity/environment (14–16,23), output/exit (17–20), then emit/write/timestamps (05,10–13). SHA and its fallback remain separate (21–22). Each unit must satisfy its source-shape fixtures before it is advertised. The dynamic ranking determines priority; the optional watch/trace/profiler inventory remains required scope, with additional event-order fixtures before implementation.

## Questions for system_adamic

1. Should checked Node module imports be supported directly, or should tsc use an explicit `'adamic'` System adapter, and what declaration/ABI boundary is allowed for each?
2. Should file descriptors be observable numbers as in Node or opaque owned handles, and what are the permitted aliasing, repeated-close and stale-handle behaviors?
3. What representation and ownership rules should Buffer, indexed byte access, mutable views and shared backing allocations have?
4. How should Node-style thrown system errors expose code, message, path and syscall, and how does that coexist with existing file APIs that return optional results?
5. What checked representation should nullable references, Date-or-undefined results and generic host callbacks use, including distinct null and undefined?
6. Which process globals are mutable/live, and what language rules govern env property reads/writes and process/stdout identity?
7. What should process.exit do to finally blocks, retained objects, pending output and host registrations while matching Node's observable termination?
8. How should escaping host callbacks be typed and rooted, including variadic/optional arguments, captured cycles and registrations that outlive their creator?
9. What event-loop and threading rules govern callback delivery, timer ordering, cancellation, reentrancy and completion after close?
10. What should Date and clock values expose for epoch, monotonicity, precision, timezone and unavailable clocks?
11. Which native memory metrics, optional GC probes and profiler capabilities may be exposed under Node names without claiming V8 semantics?
12. How should require, CommonJS filename/directory globals, executable identity, source-map hooks and bundled default-library paths map to native executables?
13. Which Buffer encoding overloads/coercions and readonly host properties belong in the checked language surface?
14. How should optional host-property probes such as typeof and private setBlocking be checked without allowing unsupported host calls to appear supported?

## Source-shape fixtures and their mutants

The existing [host fixture suite](../stage3/fixtures/host/check.py) is executable retained evidence, not a proposed future checklist. Each source shape has pinned Node stdout/stderr/exit and exact tsc provenance in [status.json](../stage3/fixtures/host/status.json). The runner checks original Node observations, native build outcomes and one actual source mutation per fixture. Its diagnostic-record mutations also test the census guard, but are not substitutes for the semantic mutants below. All 25 originals and all 25 semantic mutants were rerun on this branch; all existing originals remain native Checker blockers, not native successes.

| Fixture | tsc source sites | Semantic mutant caught by Node |
| --- | --- | --- |
| [01_readFile_utf8.a](../stage3/fixtures/host/01_readFile_utf8.a) | `src/compiler/sys.ts:1787` | `return buffer.toString("utf8", 3); → return buffer.toString("utf8", 0);` |
| [02_readFile_utf16le.a](../stage3/fixtures/host/02_readFile_utf16le.a) | `src/compiler/sys.ts:1787` | `if (len >= 2 && buffer[0] === 0xFF && buffer[1] === 0xFE) → if (false)` |
| [03_readFile_utf16be.a](../stage3/fixtures/host/03_readFile_utf16be.a) | `src/compiler/sys.ts:1787` | `buffer[i] = buffer[i + 1]; → buffer[i] = temp;` |
| [04_readFile_missing.a](../stage3/fixtures/host/04_readFile_missing.a) | `src/compiler/sys.ts:1787` | `return undefined; → return "";` |
| [05_writeFile.a](../stage3/fixtures/host/05_writeFile.a) | `src/compiler/sys.ts:1819`, `src/compiler/sys.ts:1466` | `data = byteOrderMarkIndicator + data; → data = data;` |
| [06_fileExists.a](../stage3/fixtures/host/06_fileExists.a) | `src/compiler/sys.ts:1629`, `src/compiler/sys.ts:1886`, `src/compiler/sys.ts:1901`, `src/compiler/sys.ts:901`, `src/compiler/sys.ts:1487` | `return stat.isFile(); → return stat.isDirectory();` |
| [07_directoryExists.a](../stage3/fixtures/host/07_directoryExists.a) | `src/compiler/sys.ts:1629`, `src/compiler/sys.ts:1886`, `src/compiler/sys.ts:1905`, `src/compiler/sys.ts:901`, `src/compiler/sys.ts:1487` | `return stat.isDirectory(); → return stat.isFile();` |
| [08_getDirectories.a](../stage3/fixtures/host/08_getDirectories.a) | `src/compiler/sys.ts:1629`, `src/compiler/path.ts:526`, `src/compiler/path.ts:146`, `src/compiler/path.ts:151`, `src/compiler/path.ts:165`, `src/compiler/path.ts:250`, `src/compiler/path.ts:41`, `src/compiler/path.ts:140`, `src/compiler/path.ts:807`, `src/compiler/path.ts:579`, `src/compiler/sys.ts:1838`, `src/compiler/sys.ts:1909`, `src/compiler/sys.ts:1487`, `src/compiler/path.ts:28`, `src/compiler/utilities.ts:10312` | `directories.sort(); → directories.reverse();` |
| [09_realpath.a](../stage3/fixtures/host/09_realpath.a) | `src/compiler/sys.ts:1913`, `src/compiler/sys.ts:1917`, `src/compiler/sys.ts:1491` | `return path; → return _path.resolve(path);` |
| [10_getModifiedTime.a](../stage3/fixtures/host/10_getModifiedTime.a) | `src/compiler/sys.ts:1629`, `src/compiler/sys.ts:1926`, `src/compiler/sys.ts:1487` | `return statSync(path)?.mtime; → return undefined;` |
| [11_setModifiedTime.a](../stage3/fixtures/host/11_setModifiedTime.a) | `src/compiler/sys.ts:1930` | `_fs.utimesSync(path, time, time); → _fs.utimesSync(path, new Date(0), new Date(0));` |
| [12_deleteFile.a](../stage3/fixtures/host/12_deleteFile.a) | `src/compiler/sys.ts:1939` | `return _fs.unlinkSync(path); → return;` |
| [13_createDirectory.a](../stage3/fixtures/host/13_createDirectory.a) | `src/compiler/sys.ts:1629`, `src/compiler/sys.ts:1886`, `src/compiler/sys.ts:1905`, `src/compiler/sys.ts:1545`, `src/compiler/sys.ts:901`, `src/compiler/sys.ts:1487` | `if (!nodeSystem.directoryExists(directoryName)) → if (false)` |
| [14_getCurrentDirectory.a](../stage3/fixtures/host/14_getCurrentDirectory.a) | `src/compiler/core.ts:1891`, `src/compiler/sys.ts:1501` | `callback = undefined!; → // callback stays live` |
| [15_getExecutingFilePath.a](../stage3/fixtures/host/15_getExecutingFilePath.a) | `src/compiler/sys.ts:1498`, `src/compiler/sys.ts:1560` | `__filename.endsWith("sys.js") → __filename.endsWith("sys.cjs")` |
| [16_getEnvironmentVariable.a](../stage3/fixtures/host/16_getEnvironmentVariable.a) | `src/compiler/sys.ts:1565` | `return process.env[name] &#124;&#124; ""; → return "";` |
| [17_write.a](../stage3/fixtures/host/17_write.a) | `src/compiler/sys.ts:1527` | `process.stdout.write(s); → process.stdout.write(s + "?");` |
| [18_exit_0.a](../stage3/fixtures/host/18_exit_0.a) | `src/compiler/sys.ts:1587`, `src/compiler/sys.ts:1715` | `nodeSystem.exit(0); → nodeSystem.exit(1);` |
| [19_exit_1.a](../stage3/fixtures/host/19_exit_1.a) | `src/compiler/sys.ts:1587`, `src/compiler/sys.ts:1715` | `nodeSystem.exit(1); → nodeSystem.exit(2);` |
| [20_exit_2.a](../stage3/fixtures/host/20_exit_2.a) | `src/compiler/sys.ts:1587`, `src/compiler/sys.ts:1715` | `nodeSystem.exit(2); → nodeSystem.exit(0);` |
| [21_createHash.a](../stage3/fixtures/host/21_createHash.a) | `src/compiler/sys.ts:1948`, `src/compiler/sys.ts:59`, `src/compiler/sys.ts:1572` | `createHash("sha256") → createHash("sha1")` |
| [22_createHash_fallback.a](../stage3/fixtures/host/22_createHash_fallback.a) | `src/compiler/sys.ts:59`, `src/compiler/sys.ts:1572` | `let acc = 5381; → let acc = 5382;` |
| [23_newLine.a](../stage3/fixtures/host/23_newLine.a) | `src/compiler/sys.ts:1525` | `newLine: _os.EOL → newLine: "\r\n"` |
| [24_useCaseSensitiveFileNames.a](../stage3/fixtures/host/24_useCaseSensitiveFileNames.a) | `src/compiler/sys.ts:1629`, `src/compiler/sys.ts:1886`, `src/compiler/sys.ts:1901`, `src/compiler/sys.ts:1732`, `src/compiler/sys.ts:1722`, `src/compiler/sys.ts:901`, `src/compiler/sys.ts:1487`, `src/compiler/sys.ts:1489` | `return !fileExists(swapCase(__filename)); → return fileExists(swapCase(__filename));` |
| [25_readDirectory.a](../stage3/fixtures/host/25_readDirectory.a) | `src/compiler/types.ts:7841`, `src/compiler/corePublic.ts:43`, `src/compiler/utilities.ts:9781`, `src/compiler/utilities.ts:9787`, `src/compiler/utilities.ts:9612`, `src/compiler/core.ts:1934`, `src/compiler/corePublic.ts:37`, `src/compiler/core.ts:2375`, `src/compiler/corePublic.ts:40`, `src/compiler/corePublic.ts:17`, `src/compiler/path.ts:32`, `src/compiler/path.ts:28`, `src/compiler/path.ts:526`, `src/compiler/path.ts:882`, `src/compiler/path.ts:732`, `src/compiler/path.ts:30`, `src/compiler/path.ts:146`, `src/compiler/path.ts:31`, `src/compiler/path.ts:151`, `src/compiler/path.ts:165`, `src/compiler/path.ts:250`, `src/compiler/path.ts:41`, `src/compiler/path.ts:140`, `src/compiler/path.ts:807`, `src/compiler/path.ts:579`, `src/compiler/path.ts:785`, `src/compiler/path.ts:627`, `src/compiler/path.ts:722`, `src/compiler/core.ts:320`, `src/compiler/core.ts:1750`, `src/compiler/core.ts:965`, `src/compiler/core.ts:985`, `src/compiler/core.ts:926`, `src/compiler/core.ts:17`, `src/compiler/core.ts:399`, `src/compiler/utilities.ts:9607`, `src/compiler/utilities.ts:9609`, `src/compiler/utilities.ts:9776`, `src/compiler/utilities.ts:9618`, `src/compiler/utilities.ts:9634`, `src/compiler/utilities.ts:9644`, `src/compiler/utilities.ts:9650`, `src/compiler/core.ts:618`, `src/compiler/path.ts:538`, `src/compiler/core.ts:1116`, `src/compiler/path.ts:449`, `src/compiler/path.ts:495`, `src/compiler/path.ts:622`, `src/compiler/debug.ts:196`, `src/compiler/debug.ts:213`, `src/compiler/core.ts:1121`, `src/compiler/utilities.ts:9684`, `src/compiler/utilities.ts:9594`, `src/compiler/utilities.ts:9695`, `src/compiler/utilities.ts:9670`, `src/compiler/utilities.ts:9657`, `src/compiler/path.ts:60`, `src/compiler/core.ts:1937`, `src/compiler/core.ts:220`, `src/compiler/core.ts:232`, `src/compiler/utilities.ts:9605`, `src/compiler/core.ts:1951`, `src/compiler/core.ts:2430`, `src/compiler/path.ts:388`, `src/compiler/path.ts:398`, `src/compiler/core.ts:1966`, `src/compiler/path.ts:435`, `src/compiler/path.ts:373`, `src/compiler/path.ts:115`, `src/compiler/path.ts:312`, `src/compiler/utilities.ts:9922`, `src/compiler/core.ts:2032`, `src/compiler/core.ts:1974`, `src/compiler/core.ts:2074`, `src/compiler/core.ts:2079`, `src/compiler/core.ts:145`, `src/compiler/path.ts:967`, `src/compiler/utilities.ts:9892`, `src/compiler/utilities.ts:9802`, `src/compiler/utilities.ts:9817`, `src/compiler/core.ts:1828`, `src/compiler/core.ts:1861`, `src/compiler/core.ts:1835`, `src/compiler/core.ts:1875`, `src/compiler/core.ts:2377`, `src/compiler/core.ts:375`, `src/compiler/core.ts:1038`, `src/compiler/core.ts:2247`, `src/compiler/path.ts:120`, `src/compiler/path.ts:125`, `src/compiler/core.ts:198`, `src/compiler/utilities.ts:9826`, `src/compiler/sys.ts:1629`, `src/compiler/sys.ts:1838`, `src/compiler/sys.ts:1882`, `src/compiler/sys.ts:1913`, `src/compiler/sys.ts:1917`, `src/compiler/sys.ts:1487`, `src/compiler/sys.ts:1491`, `src/compiler/utilities.ts:10312` | `if (extensions && !fileExtensionIsOneOf(name, extensions)) continue; → if (false) continue;` |

### First shape built without a ruling

Added [host_scout_djb2.a](../internal/oracle/testdata/host_scout_djb2.a), with tsc's `sys.ts:59–66` function copied verbatim after CRLF normalization. It tests empty and ASCII text, BMP/supplementary text, NUL, lone high/low surrogates and repeated text that exercises signed-shift coercion. Existing native string/number/bitwise machinery builds the exact source; no production implementation change is needed. This establishes the absent-crypto fallback algorithm only. Node's usual SHA-256 host binding, System adapter, and optional-property probes remain missing.

[The targeted test](../internal/oracle/host_scout_test.go) ties function bytes to vendored upstream, runs the Node oracle and sanitized native original, requires no stderr/nonzero exit and a clean leak check, then changes seed 5381 to 5382. The mutant must compile and run cleanly but disagree with original Node stdout; an unrelated compiler crash cannot count as a caught mutant. The normal oracle fixture is registered for release and native runs as well. This keeps the proof runnable in its own targeted test rather than requiring native tsc.

The other 25 full host fixtures cannot be made native without choosing the Node import/declaration boundary (question 1); Buffer readers additionally require question 3, system-error paths question 4, and exit/watch shapes their shutdown/callback rulings. Native currently reports TS2591 and sometimes additional checker prerequisites recorded in status.json. No binding, language rule or nullable-reference representation was invented to bypass that boundary. Watch, performance and profiling lack additional source-shape fixtures here; their existing site inventories and design questions are retained for follow-up before those APIs can be advertised.

### Expanded verification

Setup used the named branch and the previously recorded base, GOPROXY and cloud environment; Node was v24.19.0. No runtime source was touched, so setup did not request WASI SDK. Machine and CLI observation method are recorded above; there is no performance benchmark in this extension.

- `python3 stage3/fixtures/host/check.py --mutants --logs /tmp/scout-29-host/contracts`: all 25 Node contracts and recorded native Checker outcomes match; all 25 semantic mutants caught, plus 75 diagnostic/code/line-record mutations caught.
- `go test ./internal/oracle -run '^TestHostScoutDjb2ShapeAndMutant$' -count=1 -v`: original agrees, wrong-seed mutant caught, native ASan/UBSan and leak checks clean for both.
- `ADAMIC_ORACLE_RELEASE=1 go test ./internal/oracle -run '^Test(Native|Release)AgreesWithNode$/internal/oracle/testdata/host_scout_djb2.a$' -count=1 -v`: both Node comparisons pass; sanitized native and shipped release lanes pass.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`: passes; counts.md regenerated with the new fixture, committed separately from all other files.

This extension adds the exact native fallback fixture and mutation test, retains the existing host fixture suite, and leaves production unchanged. The host API implementation boundary awaits system_adamic answers rather than a scout selecting language rules.
