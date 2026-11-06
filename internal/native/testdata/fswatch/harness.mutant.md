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
| The error path retains its watcher and the failed shared source | error | TestFSWatchErrorClosesTheWatcher | its own 3 s deadline, the listener called every turn |

The edits, in node_fs_watch_impl.h, omit fs_watcher_close(watcher) in fs_watcher_fail
and the failed shared descriptor's retirement in fs_watcher_fail_shared. The error bit then comes back every turn, as sources.h says it does until the owner
closes, and the loop spins. TestFSWatchMutantErrorKeepsSource builds it through timerMutant's
private runtime copy and RuntimeLibraryForSource, so it compiles; a compile failure never counts.

Shared-source pins (#9pg753z, failure policy pending Runtime's ruling):
TestFSWatchSharedHangupRetiresAffectedGeneration replaces the shared fd with a hung-up pipe.
TestFSWatchSharedReadFailureRetiresAffectedGeneration replaces it with an open directory: poll
reports readability and read fails with EISDIR. No production hook or fabricated errno is used.
Both pin errors in opening order, an error listener closing a sibling, and a newly opened watch
surviving the old generation's failure. They run in release and under ASan/UBSan with balanced counts.
TestFSWatchSharedInodeKeepsFirstBasename holds two hard-link paths to aliases.cjs on Node: both
self events name the first path, and the entry owns that basename until its last handle closes.
