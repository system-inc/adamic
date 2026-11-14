# Polling queue roads: measurement only

**Use counts:** 13 field accesses (7 reads, 6 writes); inside the owner, 1 length read, 1 push, 2 reachable interval-queue indexed writes, 2 identity comparisons, and 2 array-passing sites (local pollQueue, external pollWatchedFileQueue). A third indexed write serves only the plain changed-file list. No splice or iterator-protocol use. The external helper adds 2 indexed reads, 4 indexed writes, 2 length reads, 1 length write and one numeric loop. [Every use and value flow](USES.md), [field sites](evidence/field-uses.json).

The queue is local in type/name but its array goes to the external polling helper. The returned watcher and timer callback retain queues through closures. No queue is directly returned from the outer function or stored in an external container.

| Scratch road | Stock tsc 6.0.3, 79 roots | Full sys.ts census delta | Emitted sys.ts JavaScript diff | Node witness |
|---|---|---|---|---|
| (a) Object with files array | 0 diagnostics; 0 new | 272:23 cast gone; 15 gone, 3 new (includes Boundary) | +4/−8 lines; 1270 diff bytes | Same 236 events, 6 change callbacks |
| (b) Object.assign array + fields | 0 diagnostics; 0 new | 272:23 cast gone; 1 gone, 1 new (includes Boundary) | +1/−4 lines; 469 diff bytes | Same 236 events, 6 change callbacks |
| (c) Local Array subclass | 0 diagnostics; 0 new | 272:23 cast gone; 1 gone, 0 new (includes Boundary) | +10/−5 lines; 1061 diff bytes | Same 236 events, 6 change callbacks |

The original stock check has zero diagnostics. Baseline census: 132 unique sys.ts findings (24 Refused, 46 NotYet, 62 Boundary), including the exact Refused “a cast the runtime can’t check” at 272:23. Removing this refusal does not establish queue lowering or whole-program acceptance; existing queue-related barriers also remain. Each road’s complete unique-site delta is listed below; raw attempts and the final relation report remain in census.*.jsonl. Five additional dependency findings in core.ts/debug.ts are retained in those streams and are unchanged across all roads; they are excluded from sys.ts totals. [Method, exact options, setup and reproduction](METHOD.md).

Road (a) is refused at the files-array argument because the helper may write undefined into an array declared to contain only WatchedFileWithUnchangedPolls; it also exposes the existing Function.call refusal. Road (b) is refused because the Object.assign intersection infers pollScheduled as boolean while the annotated interface permits unknown writes. Road (c) adds no findings, but five queue-value/return-type NotYet sites and the inherited push refusal remain. Its factory stops at the return-type/prologue boundary, and the ledger has no separate constructor unit. **No road is established as accepted by Adamic.** The matching Node witnesses are supplementary.

## Original-tree lane oracle coverage

createPollingIntervalQueue ran **9 times**, with **3,234 queue property operations** (9 / 3,234 in the test bundle processes). The counter retained the test parent and all eight workers. Counts include the three field initializations per queue; see the [per-property totals and oracle result](evidence/oracle-summary.json), [raw per-process counters](evidence/oracle-counts/) and [full oracle report](evidence/oracle/report.json).

Oracle counts: `{"failing": 1, "passing": 106366, "pending": 0}`; runners `all`, workers `8`, test filter `None`. Oracle status is `fail`; the sole failure is the expected `api/typescript.d.ts` public-API acknowledgment. Counts, phase exits, failure title and baseline diff match stage3/lane/expected.json for Linux x64 / Node v24.19.0; this checks the oracle outcome, not the outer lane declaration guard.

## What this measurement cannot see

No adaptation or production compiler change is proposed or committed. These are scratch candidates against the pinned owner after-text, not amendments to owners.json. The full single-file census looks beyond sys.ts:160, but neither builds nor executes an Adamic program. It is evidence about individual latent lowering attempts, not a successful 79-root compilation. JavaScript comparisons are stock per-module ES2024/ESNext emissions. The full oracle runs only on the original adapted tree; no full-suite equivalence is claimed for a road. The Node witness accelerates thresholds and covers a deterministic schedule; it does not prove all possible reentrant callbacks, host timer handles or time sequences. The subclass changes prototype identity; the witness exercises the actual watcher operations, not arbitrary Array consumers.

Whole-checkout searches include tracked nested submodules, adaptations and historical evidence; the dotted-access/type-name substring search over adapted src/harness/testRunner/tests finds matches only in sys.ts and reference baseline logs. Supplementary bare-name searches also cover declarations, bracket spellings and destructuring; their 17 non-sys source matches are scalar interval parameters or single-file watcher records, not polling queues (classified in USES.md). Static searches cannot prove behavior through arbitrary external code. The oracle count is coverage on the original adapted tree, not a performance comparison of roads.

Base origin/main: `1a1c0dc18e6546b5861ee56352c542194471ed70`. Upstream: `050880ce59e30b356b686bd3144efe24f875ebc8` / `v6.0.3`. Recursive submodules and API npm ci succeeded; no credential failure. All edits were in /tmp scratch. The original owner is restored after measurement.

## Every new, gone or changed sys.ts finding

Findings use the census’s distinct `(kind, where, reason, text)` definition. Unchanged source lines are mapped to baseline lines so relocation alone does not masquerade as a substantive change. A changed finding appears as its old and new entries. Full text and attempt context are preserved in the linked JSON.

### Road (a)

117 unchanged after source-line mapping. [Full delta](evidence/delta.a.json); [JavaScript diff](evidence/javascript.a.diff).

Gone:

- `Refused` `sys.ts:272:23` — a cast the runtime can't check
- `NotYet` `sys.ts:246:11` — a value of type PollingIntervalQueue
- `Boundary` `sys.ts:234:1` — function signature or prologue failed; continued at next declaration
- `NotYet` `sys.ts:271:14` — a function returning PollingIntervalQueue
- `Boundary` `sys.ts:271:5` — function signature or prologue failed; continued at next declaration
- `NotYet` `sys.ts:279:61` — a value of type PollingIntervalQueue
- `Boundary` `sys.ts:279:5` — function signature or prologue failed; continued at next declaration
- `NotYet` `sys.ts:291:64` — a value of type PollingIntervalQueue
- `Boundary` `sys.ts:291:5` — function signature or prologue failed; continued at next declaration
- `NotYet` `sys.ts:339:14` — a function returning PollingIntervalQueue
- `Boundary` `sys.ts:339:5` — function signature or prologue failed; continued at next declaration
- `Refused` `sys.ts:351:9` — inherited library member push read as an own field
- `Boundary` `sys.ts:351:9` — failed statement; state rolled back; continued at next statement or declaration
- `Boundary` `sys.ts:361:9` — failed statement; state rolled back; continued at next statement or declaration
- `Boundary` `sys.ts:367:9` — failed statement; state rolled back; continued at next statement or declaration

New:

- `Refused` `sys.ts:277:37` — a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls | undefined)[], which can write WatchedFileWithUnchangedPolls | undefined where WatchedFileWithUnchangedPolls is read
- `Refused` `sys.ts:371:20` — inherited library member call read as an own field
- `Boundary` `sys.ts:371:13` — failed statement; state rolled back; continued at next statement or declaration

### Road (b)

131 unchanged after source-line mapping. [Full delta](evidence/delta.b.json); [JavaScript diff](evidence/javascript.b.diff).

Gone:

- `Refused` `sys.ts:272:23` — a cast the runtime can't check

New:

- `Refused` `sys.ts:272:45` — a value of type WatchedFileWithUnchangedPolls[] & { pollingInterval: PollingInterval; pollIndex: number; pollScheduled: boolean; } seen as PollingIntervalQueue, which can write unknown where boolean is read

### Road (c)

131 unchanged after source-line mapping. [Full delta](evidence/delta.c.json); [JavaScript diff](evidence/javascript.c.diff).

Gone:

- `Refused` `sys.ts:272:23` — a cast the runtime can't check

New:

None.
