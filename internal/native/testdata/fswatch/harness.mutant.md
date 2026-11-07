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
| The error path calls its listeners and leaves the watcher open | error | TestFSWatchErrorClosesTheWatcher | its own 3 s deadline, the listener called every turn |
| watch_error_shape without syscall | add-error | TestFSWatchAddErrorCarriesNodeKeys | the keys line without syscall, then the read of syscall by name stopping the run |

The edit, in node_fs_watch_impl.h's fs_watcher_fail, is fs_watcher_close(watcher) replaced by a
comment. The error bit then comes back every turn, as sources.h says it does until the owner
closes, and the loop spins. TestFSWatchMutantErrorKeepsSource builds it through timerMutant's
private runtime copy and RuntimeLibraryForSource, so it compiles; a compile failure never counts.

The second, in node_fs_watch_impl.h's adamic_fs_file_host_watch, drops syscall's name, reference and
kind from watch_error_shape's tables and its count from 7 to 6, so it compiles and node_system_error
fills the six slots left. TestFSWatchMutantAddErrorDropsSyscall builds it the same way and holds its
exact output.
