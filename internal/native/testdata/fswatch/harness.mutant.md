Mutant evidence for fs.watch's native C harness (#j71fn59).

Programs reach fs.watch through internal/oracle/testdata/host_watch, held to Node there. This
harness pins what no program reaches yet. The rename mode renames a file in a watched directory,
which no program can do while fs.renameSync is unbound, and TestFSWatchRenameMatchesNode runs
oracle.a, the same operations on Node, for its truth. The error modes put a pipe whose writer is
gone where the watcher's inotify descriptor was, with dup2, so the next wait records POLLHUP on its
source. Node takes its error path only when libuv fails reading its descriptor, which it never does
on Linux, so there is no Node answer there and fswatch_test.go holds the whole output. The add-error
mode watches a missing relative path and reads the thrown error's keys by name, which no program can
do for errno, syscall, path or filename yet; fswatch_test.go holds Node v24.14.1's keys and values.

Controls: TestFSWatchRenameMatchesNode and TestFSWatchErrorClosesTheWatcher run in release, then
under ASan and UBSan with leaks checked and value counts balanced. TestFSWatchErrorWithoutListenerIsUncaught
and TestFSWatchCustomHooksRefused pin the stopping modes by their exact text.
TestFSWatchAddErrorCarriesNodeKeys runs the add-error mode in release and sanitized, leaks checked
and counts balanced. Builds are TestProduct_FSWatchHarnessRelease, TestProduct_FSWatchHarnessSanitized,
TestProduct_FSWatchHarnessMutantErrorKeepsSource and TestProduct_FSWatchHarnessMutantAddErrorDropsSyscall.

Command after products: go test ./internal/native -count=1 -run 'TestFSWatch' -v

| Mutation | Mode | Named pin | Caught by |
| --- | --- | --- | --- |
| The persistent watcher close does not decrement the live persistent count | unref | TestFSWatchSharedSourceUnrefMatchesNode | its own 3 s deadline; `TestFSWatchSharedSourceUnrefMatchesNode goes red without the persistent-count decrement: deadline exceeded; stdout=""` |
| The error path calls its listeners and leaves the watcher open | error | TestFSWatchErrorClosesTheWatcher | its own 3 s deadline, the listener called every turn |
| An extra inotify instance per watcher inserted at the live header's shared descriptor open, held to process exit | cap | TestFSWatchSharedDescriptorCapMatchesNode | exit 17; `watch failure code=EMFILE message=EMFILE: too many open files, watch '<fixture-path>'` |
| watch_error_shape without syscall | add-error | TestFSWatchAddErrorCarriesNodeKeys | the keys line without syscall, then the read of syscall by name stopping the run |
| The per-event job drain dropped in fs_watcher_dispatch, the loop's checkpoint after the source phase the only one | drain | TestFSWatchDrainsJobsPerEvent | stdout: `pending 96, rename a, rename b, rename end, job, job` where the control reads `pending 96, rename a, job, rename b, job, rename end` |

The drain mode is the per-event checkpoint's witness (#cbj1hpd). It creates a, b and end in the
watched directory before the loop runs, so their three IN_CREATE events are one read, and the listener
queues a job (an awaited settled promise, as the readiness-source harness makes one) on a and on b.
Node takes a microtask checkpoint after each event's listener, so each job prints before the next
event's line. Both orders mean what they say only if the three events are one read, so before the
loop runs the harness prints the bytes waiting on the watcher's descriptor (`FIONREAD`), and both
tests require 96, all three events (Runtime's should). A box that split them would fail on that
line, not on the order. No program can queue that job from a listener: an async listener in a void slot is a
floating task, refused for good (#8j8ecw0), so two_events.a pins that refusal and witnesses no drain.
TestFSWatchMutantDrainDropped builds the mutant through timerMutant's private runtime copy, its
anchor the checkpoint and its comment together so fs_watch_fail's own drain is untouched, and requires
the mutant's order exactly. TestProduct_FSWatchHarnessMutantDrainDropped is its build.

The edit, in node_fs_watch_impl.h's fs_watcher_fail, is fs_watcher_close(watcher) replaced by a
comment. The error bit then comes back every turn, as sources.h says it does until the owner
closes, and the loop spins. TestFSWatchMutantErrorKeepsSource builds it through timerMutant's
private runtime copy and RuntimeLibraryForSource, so it compiles; a compile failure never counts.

TestFSWatchMutantPerWatcherDescriptorHitsCap applies this cap mutation to the live
node_fs_watch_impl.h through the private runtime copy. The insertion anchor must occur
exactly once; drift fails the product before compiling. The fixture creates all files
before opening watchers, then checks EMFILE at watch, the exact message, its explicit
exit 17, and balanced heap counts under ASan and UBSan. The extra kernel descriptors
belong only to the mutant and close at process exit.

TestFSWatchMutantPersistentCountKeepsSource removes only the persistent-count decrement
in fs_watcher_close, with a Count-guarded anchor. In mode unref, the persistent watcher
closes beside an open persistent:false sibling. Node and the control print `unref exits`;
the mutant keeps the shared source referenced and reaches the deadline with empty stdout
and stderr. Its TestProduct_FSWatchSharedMutantPersistentCount builds the sanitized,
counted private runtime; an abort or sanitizer report does not count as the catch.

The add-error mutant, in node_fs_watch_impl.h's adamic_fs_file_host_watch, drops syscall's name, reference and
kind from watch_error_shape's tables and its count from 7 to 6, so it compiles and node_system_error
fills the six slots left. TestFSWatchMutantAddErrorDropsSyscall builds it the same way and holds its
exact output.
