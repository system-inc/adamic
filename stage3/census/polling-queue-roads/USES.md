# Static queue use inventory

Counts are syntactic sites, not runtime event counts. Coordinates in `field-uses.json` are relative to `evidence/owner.original.ts`, the exact pinned after-text. Declarations/type annotations are excluded.

| Category inside owner | Reads | Writes | Other |
|---|---:|---:|---|
| pollingInterval | 3 | 1 | initialization; pollQueue interval and chunk lookup; reschedule |
| pollIndex | 2 | 2 | initialization; poll result assignment/input index; empty assertion |
| pollScheduled | 2 | 3 | initialization; empty reset; scheduling checks; timer result assignment |
| array index via pollQueue alias | 0 | 3 | two interval-queue removal sites; one changed-list-only site |
| array length | 1 | 0 | nonempty check |
| push | 0 | 1 | addToPollingIntervalQueue |
| splice / iteration syntax | 0 | 0 | none inside owner |
| passing as array | — | — | one to local pollQueue, then one to external pollWatchedFileQueue |
| identity comparison | 2 | 0 | equality/inequality with changedFilesInLastPoll |

The three index-write sites are in onWatchFileStat: changed file, changed-list-to-low, low/medium promotion. Only the first and third can operate on a PollingIntervalQueue; the middle branch operates exclusively on changedFilesInLastPoll. Thus there are **two reachable interval-queue indexed writes**, plus one shared-callback write that serves only the plain changed list. Both identity checks are reachable on an interval queue.

`pollWatchedFileQueue`, outside the owner, has **2 indexed reads, 4 indexed writes, 2 length reads, 1 length write**, and one numeric for-loop, implementing hole skipping, closed-file removal, compaction and wraparound/truncation. Neither push, splice nor the iterator protocol is used there. These helper sites also serve other watcher implementations.

All queue value flow:

- Three factory calls create low, medium and high queues, stored in three local consts.
- Factory creates one array, initializes three fields, and returns it to its local caller.
- `pollingIntervalQueue` returns one of those three consts (three return sites), only to local callers.
- `pollPollingIntervalQueue` receives a queue, passes its array to local `pollQueue`, then reads/writes fields and length.
- `pollLowPollingIntervalQueue` receives a queue, passes it to `pollPollingIntervalQueue`, then reads pollScheduled. Its other pollQueue call passes the separate plain changed-file list.
- `pollQueue` forwards the queue to external `pollWatchedFileQueue` together with a callback closing over it. The helper does not return/store the queue; it invokes the callback synchronously and returns a numeric index.
- `addToPollingIntervalQueue` obtains the selected queue for push. The scheduling guard obtains it for the pollScheduled read.
- `scheduleNextPoll` obtains a queue for the pollScheduled write and again for `const t2`. The firing arrow closes over t2 and the selected local callback. `setTimeout.call` passes the host, arrow, delay and string label to the external timer, not a direct queue argument. The timer host retains the arrow, which retains the queue.
- The outer function returns `watchFile`, retaining the local queues transitively. No queue is directly returned from the outer function or assigned to an external container. **Yes, its array value goes to a function outside the owner (`pollWatchedFileQueue`); yes, queues escape transitively through closures.**

Road (a) rewrites the three direct array sites to `.files`; callback writes/comparisons and helper operations already operate on the array alias and need no further spelling change. The object container no longer goes to the external polling helper; its files array does.

Search coverage: `checkout-search.txt` records the initial rg search; `tracked-recursive-search.txt` records `git grep --recurse-submodules` over every tracked file at the pinned HEAD (715 lines, including the cohere TypeScript compiler fixture). `adapted-search.txt` separately searches the complete adapted src and tests trees, including src/harness and src/testRunner. Name matches in historical measurement JSON, adaptation pins and copied fixtures are evidence/source copies, not additional consumers of this local type. The three field spellings are also used for unrelated polling configuration or watcher implementations; those matches do not imply a PollingIntervalQueue escape.

The final adapted-tree search has 563 matching lines: all are in src/compiler/sys.ts or tests/baselines/reference logs. There are no matching direct accesses in src/harness, src/testRunner, or executable tests. The baseline log matches are timer labels such as pollPollingIntervalQueue, rather than consumers of the queue value.

## Array and identity sites inside the owner

Owner-relative lines in evidence/owner.original.ts:

| Line | Exact operation | Applies to interval queue? |
|---:|---|---|
| 47 | `pollQueue(queue, ...)` array argument | yes |
| 49 | `queue.length` read | yes |
| 63 | `pollPollingIntervalQueue(_timeoutType, queue)` value argument | yes; carries the whole container |
| 74 | `pollWatchedFileQueue(host, queue, ...)` array argument | yes; also plain changed list |
| 84 | `queue !== changedFilesInLastPoll` | yes |
| 85 | `queue[pollIndex] = undefined` | yes |
| 92 | `queue === changedFilesInLastPoll` | yes, false for interval queue |
| 95 | `queue[pollIndex] = undefined` | no; changed-list-only branch |
| 100 | `queue[pollIndex] = undefined` | yes, low/medium promotion |
| 118 | `pollingIntervalQueue(pollingInterval).push(file)` | yes |

The other `.push` sites (watchedFiles at line 25 and changedFilesInLastPoll at line 123) and changedFilesInLastPoll.length at lines 60/66 operate on separate plain arrays, not PollingIntervalQueue values. They remain untouched in all roads.

Bare-name searches supplement dotted-access searches so bracket access, destructuring and declarations are covered too: `tracked-field-words.txt.gz` has 1,248 matching lines across the tracked recursive checkout; `adapted-field-words.txt.gz` has 10,988 across adapted src/tests. Exact commands and uncompressed sizes are in field-word-searches.json.

There are 17 non-sys source matches in the bare-name search. All concern the scalar pollingInterval or a single-file watcher record, not a PollingIntervalQueue:

- compiler/tsbuildPublic.ts, watchPublic.ts and watchUtilities.ts, plus server/types.ts: watcher method parameters/signatures and interval forwarding.
- harness/watchUtils.ts: TestFileWatcher is `{ cb, pollingInterval }`.
- harness/harnessLanguageService.ts: ServerHostFileWatcher is `{ cb, pollingInterval }`, stored through watchUtils.pollingWatch.
- testRunner/unittests/helpers/virtualFileSystemWithWatch.ts: TestFileWatcher is `{ cb, pollingInterval, event? }`; watchFileWorker stores that single-watcher record.
- testRunner/unittests/helpers/tsserver.ts: a test watcher record's undefined pollingInterval initializer.

None of these records extends Array or has pollIndex/pollScheduled. Bare field-name matches in reference baselines include serialized watcher/configuration records and are not evidence that the local queue container escapes there.
