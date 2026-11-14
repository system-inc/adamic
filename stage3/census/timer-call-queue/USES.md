# Queue composition: every use and reader

The AST inventories are `queue-uses.1.json` and `queue-uses.2.json`; every matching identifier, metadata/array operation and passing site has exact source text and coordinates. The external helper inventories are byte-identical. Whole adapted src/harness/testRunner/tests and recursive tracked-checkout search receipts are in `queue-searches.json` and its compressed outputs. Exact PollingIntervalQueue type-name matches in the executable adapted tree occur only in compiler/sys.ts.

| Metadata, counting factory initialization | Reads | Writes |
|---|---:|---:|
| pollingInterval | 3 | 1 |
| pollIndex | 2 | 2 |
| pollScheduled | 2 | 3 |
| Total | 7 | 6 |

Before composition these are 13 dotted metadata accesses. After composition there are 10 dotted accesses plus 3 object-literal initializations, preserving all 13 operations. The new files field has one literal initialization and three reads. Its three reads are the only direct array-use substitutions: local pollQueue argument, nonempty length test, and selector-result push.

| Array operations | Inside owner | External pollWatchedFileQueue |
|---|---|---|
| Indexed reads | 0 | 2 |
| Indexed writes | 3 syntactic; 2 reachable for interval files | 4 |
| Length reads | 1 | 2 |
| Length writes | 0 | 1 |
| Push | 1 | 0 |
| Identity comparisons with changedFilesInLastPoll | 2 | 0 |
| Splice / iterator protocol | 0 | 0 |
| Numeric loops | 0 | 1 |

The three onWatchFileStat writes serve changed-file removal, changed-list-to-low transfer and low/medium promotion. The middle write is guarded by queue === changedFilesInLastPoll and serves only that separate plain list; only the other two can write undefined into interval files. Both identity comparisons still operate on the array alias, never the new object container.

The only element readers are unchanged:

1. sys.ts:185 reads queue[pollIndex] into watchedFile. The existing `if (!watchedFile) continue` at 186 establishes the defined record before field reads, the stat helper and the callback. Its existing isClosed guards and compaction stay untouched.
2. sys.ts:205 reads queue[pollIndex] only as a truthiness predicate before compaction. It does not consume the element as WatchedFileWithUnchangedPolls.

There is no direct element read inside createDynamicPriorityPollingWatchFile. The full census adds no Refused or NotYet (indeed, no finding of any kind), so no reader needs a new narrowing and no check is added. The files type admits undefined because both pollQueue and the shared helper can write it.

All container/value flow is accounted for:

- Three factory calls bind low, medium and high local consts. One factory return constructs a container plus its fresh files array.
- pollingIntervalQueue has exactly three return sites, one for each const, and four calls: push, scheduling guard, scheduled-handle assignment and captured t2.
- Two queue parameters carry containers: pollPollingIntervalQueue and pollLowPollingIntervalQueue. The latter forwards the whole container once to the former.
- The interval array passes once to local pollQueue as queue.files, then once to external pollWatchedFileQueue as its queue array alias. The helper invokes the callback synchronously and returns a numeric index; it does not retain or return the array.
- There are four call/passing sites in the AST inventory: those two array passes, one separate changedFilesInLastPoll call, and the whole-container forwarding call.
- The timer's zero-argument arrow captures t2 and the selected callback. The timer host receives the arrow, delay and label, not the queue as a direct argument. The returned watchFile closes over the queues transitively. No whole queue container is returned from the outer function or put into an external collection.
- The owner has 26 queue-name identifier nodes after composition: 6 declarations/parameters and 20 value uses (three selected-queue returns, nine pollPollingIntervalQueue uses, two pollLowPollingIntervalQueue uses and six pollQueue-array uses). Before composition there are 31; removing the factory's queue local removes its declaration, three initialization receivers and return read.
- The external helper has ten queue-name nodes: one parameter plus nine value uses (six indexed operations and three length operations).

Unrelated watchedFiles and changedFilesInLastPoll arrays retain their pushes, lengths and declarations. Search matches outside sys.ts concern scalar pollingInterval parameters, single-file watcher records, baseline timer labels or historical/copied source; none gains the local queue container. Recursive checkout matches in adaptation pins and the cohere TypeScript fixture are source/proof copies, not additional consumers of this adapted closure-local value.
