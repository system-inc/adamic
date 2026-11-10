# Paired never-call census, task 8n7eet3

Parent `bdbc73864f82dbc0c51434decf0ffe96d4782767` against head `392cb60b7a967a4515f4f0041592359b57746955`, measured November 10, 2026. Report only; no compiler change is included. The remote optimizer branch was checked before setup and still pointed at the requested head.

The seven occurrences at debug.ts:227 disappear in the semver full run: six have the Version-fields target and one has the Comparator-array target. They deduplicate to two diagnostic findings. The standalone debug full run has no changes because those generic declarations have no concrete instantiation witness in that selected file. The full Version run retains its static-initialization refusal at semver.ts:47:41, along with its other existing stops.

Status: full runs complete; the corrected complete entry census is in progress. No entry-program result is claimed yet.

## Pins, scope and guards

One adapted tree was built from the parent with `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, `ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-6.0.3`, and `bash stage3/apply.sh --build-product /workspace/never-call-adapted`. Stock source is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both compiler commits use cohere `5360753e19e4417675a90e92a1823bea66b13918` and checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go is 1.27.1, Node is v24.19.0. Each side was built from its own SHA with its own scratch overlay. Compiler source stayed unchanged.

Full runs pass the adapted `src/compiler` directory, set `LATENT_FULL=1`, and select exactly one adapted file with `LATENT_ONLY_FILE`. The entry runs load only `src/tsc/tsc.ts` as the original root and use the scratch `Program.LatentReplayEntryReach()` hook, corrected to preserve the original entry as `Files()[0]`, to expand the observation ledger to its 81 resolved source files. The original root remains unchanged. Both entry driver copies are retained as `.go.txt` evidence.

All reported runs set `LATENT_ASSERT_NO_OUTPUT=1`. The ordinary loader is disabled, ordinary lowering must return nil IR and the measurement-only error, and no backend runs. Unit and prepass timings are separate streams. Every child exit is recorded, and the wrapper stops a run after 30 minutes without a change in its output evidence.

The requested reference `stage3/census/core-sites-af42d3dc/REPORT.md` is absent from the parent commit. The specified full-mode flags were used directly. The actual headers report `checker_rejected: false`, with no checker diagnostics. The meter retains its fixed label "measured on a checker-rejected program"; that label does not override the observed checker status. These are lowering censuses, not successful native compilations.

The first entry invocation loaded the program but recorded only the root file. Those two preliminary runs are retained under `tsc-first-*` as excluded diagnostic evidence. The corrected census is exclusively `tsc-entry-ordered-first-*`; its ledger must match the independently enumerated 81-file `closure.json`.

## Times and peak memory

| Run | Side | Elapsed seconds | Peak RSS KiB | Child exit | Silence cutoff |
| --- | --- | --- | --- | --- | --- |
| debug-full | parent | 36.050 | 881,820 | 0 | false |
| debug-full | head | 33.044 | 881,588 | 0 | false |
| semver-full | parent | 46.117 | 967,080 | 0 | false |
| semver-full | head | 57.455 | 970,400 | 0 | false |

The four full runs were sequential. The final two entry runs overlapped to stay within the session window; their wall times include that concurrency and are not an isolated performance comparison. Full-run elapsed time is wrapper wall time, with up to one second of completion-poll overhead; full-run peak RSS uses Linux getrusage(RUSAGE_CHILDREN).ru_maxrss. The entry launchers were interrupted during an environment reconnection while the compilers survived. An exit observer recovered entry wall time from kernel process start ticks and the boot clock, actual wait status from the kernel, and lifetime peak RSS from /proc/PID/status VmHWM at PTRACE_EVENT_EXIT. It forwards ordinary signals unchanged and does not enable EXITKILL. Entry times include this late exit observation and concurrency, so they are not an isolated performance benchmark. A controlled 64 MiB observer probe returned the expected exit 7 and reported 71,552 KiB peak RSS, exactly matching getrusage for that child. All RSS values describe each compiler process, not the container. Setup and builds are excluded.

## Cleared, new and changed counts

| Comparison | View | Cleared | Changed reason at site | New findings | New sites | Changed unit first stop | Regressions |
| --- | --- | --- | --- | --- | --- | --- | --- |
| debug-full | raw_findings | 0 | 0 | 0 | 0 | 0 | 0 |
| debug-full | distinct_reads | 0 | 0 | 0 | 0 | 0 | 0 |
| semver-full | raw_findings | 4 | 0 | 0 | 0 | 6 | 0 |
| semver-full | distinct_reads | 3 | 0 | 0 | 0 | 6 | 0 |

These are the pair tool's exact counts, including measurement `Boundary` records. "Cleared" means an old finding has no remaining finding at its identity. "Changed reason at site" means an old finding disappeared but another remains there. "New findings" includes exposed reasons in already-blocked units; it is not a regression count. Raw identity preserves kind, source identity, reason and text. Distinct identity uses the outermost contiguous access chain and exact reason.

For semver full mode, the four cleared raw findings are two NotYet reasons at debug.ts:227:5 and two Boundary records at debug.ts:235:9. They become three distinct entries: two diagnostic reasons and one boundary reason. The seven original NotYet occurrences and seven accompanying boundary occurrences are removed. No new full-mode finding appears in either selected-file comparison.

## debug.ts: assertion helpers and callers

Standalone debug full mode records 81 declaration units and 22 finding occurrences on each side, with 19 raw and 19 distinct entries including boundaries. No finding occurs at debug.ts:227 or in lines 227 through 250 in that selected-file run. The six relevant declaration units are deferred on both sides, with `instantiated: false`; absence of findings there is not a claim that their bodies lowered.

| Declaration | Unit | Parent status | Head status | Parent / head reason |
| --- | --- | --- | --- | --- |
| assertIsDefined | src/compiler/debug.ts:227:5 | deferred_generic_instantiation | deferred_generic_instantiation | No finding; generic instantiation deferred on both sides |
| checkDefined | src/compiler/debug.ts:234:5 | deferred_generic_instantiation | deferred_generic_instantiation | No finding; generic instantiation deferred on both sides |
| assertEachIsDefined | src/compiler/debug.ts:239:5 | deferred_generic_instantiation | deferred_generic_instantiation | No finding; generic instantiation deferred on both sides |
| assertEachIsDefined | src/compiler/debug.ts:240:5 | deferred_generic_instantiation | deferred_generic_instantiation | No finding; generic instantiation deferred on both sides |
| assertEachIsDefined | src/compiler/debug.ts:241:5 | deferred_generic_instantiation | deferred_generic_instantiation | No finding; generic instantiation deferred on both sides |
| checkEachDefined | src/compiler/debug.ts:247:5 | deferred_generic_instantiation | deferred_generic_instantiation | No finding; generic instantiation deferred on both sides |

| Direct call in debug.ts | Caller | Call text | Parent / head finding at this local call site |
| --- | --- | --- | --- |
| debug.ts:235:9 | checkDefined | `assertIsDefined(value, message, stackCrawlMark \|\| checkDefined)` | None in standalone debug full mode |
| debug.ts:243:13 | assertEachIsDefined | `assertIsDefined(v, message, stackCrawlMark \|\| assertEachIsDefined)` | None in standalone debug full mode |
| debug.ts:248:9 | checkEachDefined | `assertEachIsDefined(value, message, stackCrawlMark \|\| checkEachDefined)` | None in standalone debug full mode |

The semver-selected run supplies concrete caller witnesses and records every occurrence below at debug.ts:227:5. "Absent" means that exact site has no finding in that caller attempt on the head. All rows have phase `lowering`.

| Attempting caller unit | Site | Parent reason | Head reason |
| --- | --- | --- | --- |
| Version `src/compiler/semver.ts:46:1` | src/compiler/debug.ts:227:5 | NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent |
| VersionRange `src/compiler/semver.ts:207:1` | src/compiler/debug.ts:227:5 | NotYet: "checked predicate overload target Comparator[][]" | Absent |
| VersionRange `src/compiler/semver.ts:207:1` | src/compiler/debug.ts:227:5 | NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent |
| parsePartial `src/compiler/semver.ts:297:1` | src/compiler/debug.ts:227:5 | NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent |
| parseHyphen `src/compiler/semver.ts:314:1` | src/compiler/debug.ts:227:5 | NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent |
| parseComparator `src/compiler/semver.ts:336:1` | src/compiler/debug.ts:227:5 | NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent |
| testComparator `src/compiler/semver.ts:425:1` | src/compiler/debug.ts:227:5 | NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent |

The associated checkDefined call at debug.ts:235:9 loses seven boundary occurrences with reason "failed statement; state rolled back; continued at next statement or declaration". Neither full stream has a finding inside assertEachIsDefined (239 through 245) or at its call in checkEachDefined (248). Every full-mode finding, including unchanged callers and dependency sites, is listed below.

## semver.ts: Version and Version.zero

Version spans lines 46 through 139; Version.zero is declared at line 47. Full mode continues after failed statements, so its first recorded failure and its later static-field failure are separate observations. The Version unit's first recorded failure changes from the checked Version-fields predicate at debug.ts:227:5 to debug.ts:252:33, "a value of type never". The static initializer still reports semver.ts:47:41, "a constructor object escaping before static fields are initialized", on both sides.

| Every finding site in Version | Kind | Parent reason | Head reason | Occurrences parent / head |
| --- | --- | --- | --- | --- |
| src/compiler/semver.ts:118:17 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 6 / 6 |
| src/compiler/semver.ts:123:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 6 / 6 |
| src/compiler/semver.ts:124:13 | NotYet | "a destructured name that isn't plain" | "a destructured name that isn't plain" | 6 / 6 |
| src/compiler/semver.ts:130:28 | NotYet | "reading major" | "reading major" | 6 / 6 |
| src/compiler/semver.ts:130:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 6 / 6 |
| src/compiler/semver.ts:46:1 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 |
| src/compiler/semver.ts:47:41 | Refused | "a constructor object escaping before static fields are initialized" | "a constructor object escaping before static fields are initialized" | 1 / 1 |

The dependency sites reached while attempting Version itself are also exhaustive here:

| Dependency site | Kind | Parent reason | Head reason |
| --- | --- | --- | --- |
| src/compiler/debug.ts:227:5 | NotYet | "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent |
| src/compiler/debug.ts:235:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | Absent |
| src/compiler/debug.ts:252:33 | NotYet | "a value of type never" | "a value of type never" |


## Other full-mode movements

There are no other changed finding sites in either full-mode stream beyond debug.ts:227:5 and its boundary site debug.ts:235:9. Six semver unit first stops change because that earlier finding disappears; all remain blocked.

| Unit | Parent first recorded stop | Head first recorded stop |
| --- | --- | --- |
| Version `src/compiler/semver.ts:46:1` | `src/compiler/debug.ts:227:5` NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | `src/compiler/debug.ts:252:33` NotYet: "a value of type never" |
| VersionRange `src/compiler/semver.ts:207:1` | `src/compiler/debug.ts:227:5` NotYet: "checked predicate overload target Comparator[][]" | `src/compiler/debug.ts:252:33` NotYet: "a value of type never" |
| parsePartial `src/compiler/semver.ts:297:1` | `src/compiler/debug.ts:227:5` NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | `src/compiler/debug.ts:252:33` NotYet: "a value of type never" |
| parseHyphen `src/compiler/semver.ts:314:1` | `src/compiler/debug.ts:227:5` NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | `src/compiler/debug.ts:252:33` NotYet: "a value of type never" |
| parseComparator `src/compiler/semver.ts:336:1` | `src/compiler/debug.ts:227:5` NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | `src/compiler/debug.ts:252:33` NotYet: "a value of type never" |
| testComparator `src/compiler/semver.ts:425:1` | `src/compiler/debug.ts:227:5` NotYet: "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | `src/compiler/debug.ts:252:33` NotYet: "a value of type never" |


## Complete selected-file full-mode finding ledgers

These tables include every recorded finding site and exact reason, including boundaries and dependencies. Counts preserve all attempting-unit occurrences; the paired JSON provides the raw and access-chain identities. "Absent" is absence at that site, not successful lowering of its caller.

### debug-full

| Site | Kind | Parent reason | Head reason | Occurrences parent / head | Every attempting unit |
| --- | --- | --- | --- | --- | --- |
| src/compiler/core.ts:1049:5 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/debug.ts:389:5 |
| src/compiler/core.ts:1049:82 | NotYet | "a comparator that isn't a function" | "a comparator that isn't a function" | 1 / 1 | src/compiler/debug.ts:389:5 |
| src/compiler/debug.ts:252:33 | NotYet | "a value of type never" | "a value of type never" | 1 / 1 | src/compiler/debug.ts:252:5 |
| src/compiler/debug.ts:252:5 | Boundary | "function signature or prologue failed; continued at next declaration" | "function signature or prologue failed; continued at next declaration" | 1 / 1 | src/compiler/debug.ts:252:5 |
| src/compiler/debug.ts:257:86 | Refused | "overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | "overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | 2 / 2 | src/compiler/debug.ts:261:5<br>src/compiler/debug.ts:262:5 |
| src/compiler/debug.ts:261:5 | Boundary | "function signature or prologue failed; continued at next declaration" | "function signature or prologue failed; continued at next declaration" | 1 / 1 | src/compiler/debug.ts:261:5 |
| src/compiler/debug.ts:262:5 | Boundary | "function signature or prologue failed; continued at next declaration" | "function signature or prologue failed; continued at next declaration" | 1 / 1 | src/compiler/debug.ts:262:5 |
| src/compiler/debug.ts:273:82 | Refused | "overload 1 of assertNode parameter test cannot be served by implementation parameter test" | "overload 1 of assertNode parameter test cannot be served by implementation parameter test" | 2 / 2 | src/compiler/debug.ts:274:5<br>src/compiler/debug.ts:275:5 |
| src/compiler/debug.ts:274:5 | Boundary | "function signature or prologue failed; continued at next declaration" | "function signature or prologue failed; continued at next declaration" | 1 / 1 | src/compiler/debug.ts:274:5 |
| src/compiler/debug.ts:275:5 | Boundary | "function signature or prologue failed; continued at next declaration" | "function signature or prologue failed; continued at next declaration" | 1 / 1 | src/compiler/debug.ts:275:5 |
| src/compiler/debug.ts:299:78 | Refused | "overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | "overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | 2 / 2 | src/compiler/debug.ts:301:5<br>src/compiler/debug.ts:302:5 |
| src/compiler/debug.ts:301:5 | Boundary | "function signature or prologue failed; continued at next declaration" | "function signature or prologue failed; continued at next declaration" | 1 / 1 | src/compiler/debug.ts:301:5 |
| src/compiler/debug.ts:302:5 | Boundary | "function signature or prologue failed; continued at next declaration" | "function signature or prologue failed; continued at next declaration" | 1 / 1 | src/compiler/debug.ts:302:5 |
| src/compiler/debug.ts:350:135 | NotYet | "a template interpolating an object, an array, a map, a function or undefined" | "a template interpolating an object, an array, a map, a function or undefined" | 1 / 1 | src/compiler/debug.ts:349:5 |
| src/compiler/debug.ts:350:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/debug.ts:349:5 |
| src/compiler/debug.ts:400:13 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/debug.ts:389:5 |
| src/compiler/debug.ts:400:27 | NotYet | "a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)" | "a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)" | 1 / 1 | src/compiler/debug.ts:389:5 |
| src/compiler/debug.ts:401:13 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/debug.ts:389:5 |
| src/compiler/debug.ts:401:24 | NotYet | "reading value" | "reading value" | 1 / 1 | src/compiler/debug.ts:389:5 |


### semver-full

| Site | Kind | Parent reason | Head reason | Occurrences parent / head | Every attempting unit |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts:227:5 | NotYet | "checked predicate overload target Comparator[][]" | Absent | 1 / 0 | src/compiler/semver.ts:207:1 |
| src/compiler/debug.ts:227:5 | NotYet | "checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; }" | Absent | 6 / 0 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/debug.ts:235:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | Absent | 7 / 0 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/debug.ts:252:33 | NotYet | "a value of type never" | "a value of type never" | 6 / 6 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/semver.ts:118:17 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 6 / 6 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/semver.ts:123:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 6 / 6 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/semver.ts:124:13 | NotYet | "a destructured name that isn't plain" | "a destructured name that isn't plain" | 6 / 6 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/semver.ts:130:28 | NotYet | "reading major" | "reading major" | 6 / 6 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/semver.ts:130:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 6 / 6 | src/compiler/semver.ts:207:1<br>src/compiler/semver.ts:297:1<br>src/compiler/semver.ts:314:1<br>src/compiler/semver.ts:336:1<br>src/compiler/semver.ts:425:1<br>src/compiler/semver.ts:46:1 |
| src/compiler/semver.ts:342:9 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/semver.ts:336:1 |
| src/compiler/semver.ts:380:13 | NotYet | "a case whose type differs from the switch's" | "a case whose type differs from the switch's" | 1 / 1 | src/compiler/semver.ts:336:1 |
| src/compiler/semver.ts:439:13 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/semver.ts:425:1 |
| src/compiler/semver.ts:452:37 | NotYet | "a template interpolating an object, an array, a map, a function or undefined" | "a template interpolating an object, an array, a map, a function or undefined" | 1 / 1 | src/compiler/semver.ts:451:1 |
| src/compiler/semver.ts:452:5 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/semver.ts:451:1 |
| src/compiler/semver.ts:46:1 | Boundary | "failed statement; state rolled back; continued at next statement or declaration" | "failed statement; state rolled back; continued at next statement or declaration" | 1 / 1 | src/compiler/semver.ts:46:1 |
| src/compiler/semver.ts:47:41 | Refused | "a constructor object escaping before static fields are initialized" | "a constructor object escaping before static fields are initialized" | 1 / 1 | src/compiler/semver.ts:46:1 |


## Coverage and checker-relation audit

Final coverage audit is pending the complete entry streams.

## Excluded entry diagnostics and measurement correction

The initial `tsc-first-*` invocation recorded only the root file and is excluded. The next `tsc-entry-first-*` expansion observed 81 files but put corePublic.ts first. Production `source_identity.go` seeds import-declaration routes from `program.Files()[0]`; this caused 2,406 parent measurement panics, including Version, with reason "lower: source module has no import-declaration route: /workspace/never-call-adapted/src/compiler/semver.ts" for Version. The matching head diagnostic was stopped with SIGTERM once the cause was identified. Neither stream is used in the compiler comparison.

The final `tsc-entry-ordered-first-*` overlay preserves the single original entry both in LatentRoots and as Files()[0], appending each remaining resolved source once. It does not replace the root with 81 independent roots, suppress panics or change compiler proof logic. Both no-output guards remain active. The driver asserts entry-first order before measurement, and the audit checks the first file, complete closure, unit count and unchanged source hashes. The hook and driver are retained verbatim as `.go.txt` evidence.

One ordered-head launch was attempted before its executable existed. No compiler child started; wrapper exit was 1, with no compiler exit code or memory measurement. It is a launch diagnostic, not a census panic. The valid head measurement started after both builds completed. The original ordered-head checkpoint retains this launch failure; the head-complete checkpoint contains the valid completed run.

| Diagnostic run | Side | Elapsed seconds | Peak RSS KiB | Compiler child exit | Use |
| --- | --- | --- | --- | --- | --- |
| tsc-first | parent | 20.065 | 801960 | 0 | Excluded |
| tsc-first | head | 23.039 | 860524 | 0 | Excluded |
| tsc-entry-first | parent | 1667.692 | 2781216 | 0 | Excluded |
| tsc-entry-first | head | 100.146 | 986148 | -15 | Excluded |


## Evidence and checkpoints

`provenance.json`, setup/apply/build logs, binary hashes, overlay hashes, `.go.txt` entry drivers, `closure.json`, `inventory.json`, run wrappers, unit timing streams and exact pairing outputs are retained beside this report. The `*-paired.json` files are the complete comparisons. Streams over 20,000,000 bytes are gzip-compressed with a fixed gzip timestamp. `SHA256SUMS` records the SHA-256 of every retained artifact; `stream-hashes.json` additionally records compressed and decompressed stream identities.

Checkpoints were pushed to separate `typescript/never-call-census-checkpoint-*` refs after apply and after each run. This keeps the final `typescript/never-call-census` branch as one commit directly on the parent without force pushing or rewriting history. The two `tsc-first-*` checkpoints contain the excluded root-only diagnostic observations; the `tsc-entry-ordered-first-*` checkpoints contain the corrected complete entry census.

The preliminary root-only runs took 20.065s (801,960 KiB peak RSS, exit 0) / 23.039s (860,524 KiB peak RSS, exit 0) for parent / head. They are not used in any census delta.
