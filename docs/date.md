# Date UTC and ISO forms

Roadmap step 29 (#96rpfw2), ruling 10: Date stores epoch milliseconds, with Node's TimeClip. Only UTC and ISO forms are admitted. Locale and local-time operations are refused at compile time, with a diagnostic naming ruling 10; pinning TZ does not make them admissible.

The library implements numeric and ISO construction, copying a Date, Date.now (including a function value), getTime, valueOf, numeric Date comparisons, Number(date), Date.UTC, all eight UTC getters, toISOString, toUTCString and Date.parse of proven ISO strings. Date-only ISO forms are UTC. Date-time forms require Z or an explicit numeric offset. Literal types, finite literal unions, constant aliases and toISOString results establish the parse proof; arbitrary strings remain refused. Invalid ISO components return NaN as Node does.

Date.now uses CLOCK_REALTIME on native and WASI and Node's Date.now on JavaScript. It returns whole epoch milliseconds. Clock values depend on the machine and may move backwards; no monotonicity is promised. Tests compare integer type and epoch range, rather than exact clock readings.

Local getters and setters, getTimezoneOffset, toString, toDateString, toTimeString, toLocale*String, multi-argument local construction, Date() and non-ISO parsing are refused under ruling 10. UTC setters, setTime and toJSON are not implemented by this unit. Detached prototype methods, arbitrary receivers, intrinsic overrides and structural erasure of Date's internal slot remain NotYet. Catching toISOString's RangeError remains NotYet because catchable RangeError descriptors are not represented; uncaught invalid-time formatting agrees with the oracle's panic convention. Optional Date calls remain a compiler limitation: `function f(d: Date | undefined) { return d?.getTime(); }`.

## TypeScript compiler inventory

Inventory of TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Paths below are relative to src/compiler. The inventory preceded implementation.

| File and lines | Date use | Result |
| --- | --- | --- |
| performanceCore.ts:102 | Date.now function-value timestamp fallback | Supported |
| sys.ts:107 | new Date(0) missing-file timestamp | Supported |
| sys.ts:435,540,541,1365 | getTime comparisons | Supported |
| sys.ts:1699 | new Date().toISOString() profiling filename | Supported |
| sys.ts:1930 | Date passed to fs.utimesSync | Existing fs host supported with the canonical Date internal slot |
| tsbuildPublic.ts:134 | new Date(-8640000000000000) | Supported |
| tsbuildPublic.ts:205,206 | host.now or new Date() | Date construction supported |
| tsbuildPublic.ts:1655,1710,1720,1733,1749 | Date timestamp comparisons | Numeric Date comparisons supported |
| watch.ts:179 | new Date().toLocaleTimeString() | Refused under ruling 10; watch formatter needs adaptation |
| watch.ts:185 | system.now().toLocaleTimeString("en-US", {timeZone:"UTC"}) | Refused under ruling 10, including the explicit UTC locale call |

Other Date references are type annotations: sys.ts:86,91,110,389,539,891,1421,1422,1449; watchPublic.ts:274; tsbuild.ts:84; tsbuildPublic.ts:225,226,238,364,370,371,419,420,1087,1088,1371,1375,1383,1445,1457,1824. utilities.ts:1956 is a compiler global-name table, not a host Date operation. The two watch call sites conflict with the request to preserve every tsc use: adapting that formatter belongs to the TypeScript patch set, rather than making a locale exception to ruling 10.

## Verification

Linux, Node 24.19.0, test262 commit 2e0a56762801e275a9fdf96dc49d90ba0cddcf63, built-ins/Date, TZ=UTC, adaptation disabled: before 2 pass, 0 disagreement, 199 refused; after 71 pass, 0 disagreement, 130 refused. The remaining 123 TypeScript-checker rejections and 270 harness skips are unchanged. These directory measurements use the native test262 runner.

Eight fixture families cover epoch values, UTC construction, UTC getters, formatting, ISO parsing, clocks, intrinsic own-name metadata and invalid ISO formatting. Each has a clean mutant caught only by comparison with Node, on native, JavaScript and WASI. Format, parse and getter fixtures also pass with TZ=Asia/Kathmandu. Linux allocation counts are recorded in internal/oracle/counts.md. Runtime arithmetic and ISO parsing ports are credited in THIRD_PARTY_NOTICES.md.
