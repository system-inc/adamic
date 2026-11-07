Mutant evidence for fs.watch's native C harness (#j71fn59).

Programs reach fs.watch through internal/oracle/testdata/host_watch, held to Node there. This
harness pins what no program reaches yet. The rename mode renames a file in a watched directory,
which no program can do while fs.renameSync is unbound, and TestFSWatchRenameMatchesNode runs
oracle.a, the same operations on Node, for its truth. The error modes put a pipe whose writer is
gone where the watcher's inotify descriptor was, with dup2, so the next wait records POLLHUP on its
source. Node takes its error path only when libuv fails reading its descriptor, which it never does
on Linux, so there is no Node answer there and fswatch_test.go holds the whole output.

Controls: TestFSWatchRenameMatchesNode and TestFSWatchErrorClosesTheWatcher run in release, then
under ASan and UBSan with leaks checked and value counts balanced. TestFSWatchErrorWithoutListenerIsUncaught
and TestFSWatchCustomHooksRefused pin the stopping modes by their exact text. Builds are
TestProduct_FSWatchHarnessRelease, TestProduct_FSWatchHarnessSanitized and
TestProduct_FSWatchHarnessMutantErrorKeepsSource.

Command after products: go test ./internal/native -count=1 -run 'TestFSWatch' -v

| Mutation | Mode | Named pin | Caught by |
| --- | --- | --- | --- |
| The error path calls its listeners and leaves the watcher open | error | TestFSWatchErrorClosesTheWatcher | its own 3 s deadline, the listener called every turn |
| An extra inotify instance per watcher inserted at the live header's shared descriptor open, held to process exit | cap | TestFSWatchSharedDescriptorCapMatchesNode | exit 17; `watch failure code=EMFILE message=EMFILE: too many open files, watch '<fixture-path>'` |

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
