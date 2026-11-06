Mutant evidence for the fs.watch fixtures and the harness files they touch (#j71fn59).

Each fixture does its file system work before the loop runs, so Node and both backends read the
same inotify queue. Shared-source pins additionally cover same-wd fan-out, a microtask closing
another handle, reference transitions and delete/recreate without assuming unrelated-wd order. library_node_watch_test.go holds each to Node on native under ASan and UBSan with an
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

TestFSWatchMutant* build each against a private copy of the runtime and require the catch named.
The shared source's data is a descriptor generation, never a watcher pointer. The close mutant
omits retiring the now-empty source, leaving the loop live until its own deadline.

The WASI table (wasi_refusals_test.go) records each fixture's first refused host entry exactly,
fs.watch's own or an earlier fs.mkdtempSync. The fs adapter probes in wasi_host_adaptation_test.go
now expand node_fs_watch_impl.h beside node_watch_file_impl.h before their symbol renaming, since
the include defines public fs entry points too; no comparator changes. reflect_get's
stop_host_watcher.a is the FSWatcher host shape's pin in host_shapes_test.go's table: its close
reads as Node's prototype method, so a run-time key read stops on both backends where Node prints
function.

Shared-descriptor regressions (#9pg753z):

| Mutation | Fixture | Named red pin | Catch |
| --- | --- | --- | --- |
| Restore an owned private inotify instance and kernel watch per handle | descriptor_cap.a | TestFSWatchDescriptorCapMatchesNode | EMFILE before 100, under the same RLIMIT_NOFILE=64 as Node |
| inotify_rm_watch on any handle close | same_path_close.a | TestFSWatchSamePathSurvivesCloseMatchesNode | surviving handle never receives its change; 5 s deadline |

The resource mutant preserves shared dispatch but opens and closes a private instance with each
handle, isolating descriptor consumption from event delivery. It is not a leaked unused fd. The
limited child execs Node/native after setrlimit; no prlimit tool, root-only sysctl or skip is needed.
TestFSWatchMutantPrivateInstancesHitDescriptorCap asserts the real EMFILE message and a Node
mismatch. TestFSWatchMutantRemovingSharedWatchStopsSurvivor asserts its own deadline.

Products first, then: go test ./internal/oracle -count=1 -run 'TestFSWatch' -v
same_path_fanout.a also requires both handles to receive an event; same_path_micro_close.a requires
the first listener's microtask to close the sibling before its turn. same_path_unref.a pins demotion
after the last persistent close. delete_recreate.a pins a fresh generation opened from a callback.
