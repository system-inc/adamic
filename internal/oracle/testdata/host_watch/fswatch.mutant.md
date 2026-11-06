Mutant evidence for the fs.watch fixtures and the harness files they touch (#j71fn59).

Each fixture does its file system work before the loop runs, so Node and both backends read the
same inotify queue, and each uses one watcher, since Node doesn't define the order across watchers
in one turn. library_node_watch_test.go holds each to Node on native under ASan and UBSan with an
explicit environment, and on the JavaScript backend. The same fixtures run in the ordinary native,
JavaScript, gob and counted families. recursive.a isn't one of them: Node watches recursively there,
and TestFSWatchRecursiveStopsByName holds both backends to the same stop instead.

| Mutation in node_fs_watch_impl.h | Fixture | Named pin | Caught by |
| --- | --- | --- | --- |
| 'change' and 'rename' swapped | create.a | TestFSWatchCreateMatchesNode | stdout against Node |
| A self event's basename dropped | write.a | TestFSWatchWriteMatchesNode | stdout against Node |
| close leaves the source registered | self_close.a | TestFSWatchSelfCloseExitsMatchesNode | its own 5 s deadline |

TestFSWatchMutant* build each against a private copy of the runtime and require the catch named.
The source's data is the watcher's identity, never the record, so the third mutant spins on a stale
source rather than reading freed memory.

The WASI table (wasi_refusals_test.go) records each fixture's first refused host entry exactly,
fs.watch's own or an earlier fs.mkdtempSync. The fs adapter probes in wasi_host_adaptation_test.go
now expand node_fs_watch_impl.h beside node_watch_file_impl.h before their symbol renaming, since
the include defines public fs entry points too; no comparator changes. reflect_get's
stop_host_watcher.a is the FSWatcher host shape's pin in host_shapes_test.go's table: its close
reads as Node's prototype method, so a run-time key read stops on both backends where Node prints
function.
