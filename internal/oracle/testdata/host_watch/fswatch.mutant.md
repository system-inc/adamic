Mutant evidence for the fs.watch fixtures and the harness files they touch (#j71fn59).

Each fixture does its file system work before the loop runs, so Node and both backends read the
same inotify queue, and each uses one watcher, since Node doesn't define the order across watchers
in one turn. library_node_watch_test.go holds each to Node on native under ASan and UBSan with an
explicit environment, and on the JavaScript backend. The same fixtures run in the ordinary native,
JavaScript, gob and counted families. recursive.a, view_other_event.a and view_after_close.a aren't
among them: Node goes on in each, watching recursively or keeping the listener, and
TestFSWatchRecursiveStopsByName, TestFSWatchViewOtherEventStopsByName and
TestFSWatchViewAfterCloseStopsByName hold native, sanitized and in release, and the JavaScript backend
to one named stop instead. The two views are tsc's FsWatchWorkerWatcher, the road on takes at
sys.ts:1270, where lowering's literal "error" can't look. two_events.a writes both files before
the loop runs, so their two IN_CREATE events are one read, and its mutant shows they are: with one
checkpoint for the read, native prints both names before either microtask.

| Mutation in node_fs_watch_impl.h | Fixture | Named pin | Caught by |
| --- | --- | --- | --- |
| 'change' and 'rename' swapped | create.a | TestFSWatchCreateMatchesNode | stdout against Node |
| A self event's basename dropped | write.a | TestFSWatchWriteMatchesNode | stdout against Node |
| close leaves the source registered | self_close.a | TestFSWatchSelfCloseExitsMatchesNode | its own 5 s deadline |
| The per-event job drain dropped, one checkpoint for the whole read | two_events.a | TestFSWatchTwoEventsMatchesNode | stdout against Node: a, b, micro, micro |
| on's run-time event name check dropped | view_other_event.a | TestFSWatchViewOtherEventStopsByName | the run doesn't stop by name |

The pending close (#7s8hgd5, Runtime's ruling on #j71fn59) is loop state, so its mutants edit
the loop's own files rather than node_fs_watch_impl.h. close_pending.a is Workers' program in
Adamic's form: its directory goes with fs.rmSync before the loop rather than in
process.on('exit'), which doesn't lower, and the output is Node's either way, since the file's
watch hears nothing of its directory. unwatch_file_pending.a pins watchFile's site, the poller's
retirement in node_watch_file_impl.h, since that poller lives on the timer provider. The no-hang
pin is close_twice.a, a lone persistent watcher closed with nothing else armed. One turn only is
the readiness source's pin, TestSourcePendingCloseKeepsOneTurn
(internal/native/testdata/sources/readiness.mutant.md): a listener's own write is read in the
same dispatch, as libuv reads until `EAGAIN`, so no program here keeps a file changing across
that turn deterministically.

| Mutation | Fixture | Named pin | Caught by |
| --- | --- | --- | --- |
| sources_impl.h: the pending close counted as done at once | close_pending.a | TestFSWatchClosePendingMatchesNode | stdout against Node: the three event lines vanish |
| async_host_impl.h: the pending close's wait takes the default timeout | close_twice.a | TestFSWatchCloseTwiceMatchesNode | its own 5 s deadline |
| node_watch_file_impl.h: the poller's retirement pends no close | unwatch_file_pending.a | TestFSWatchUnwatchFilePendingMatchesNode | stdout against Node: the three event lines vanish |

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
