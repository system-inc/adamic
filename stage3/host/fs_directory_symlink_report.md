Built: string symlinkSync with omitted type and constant file/dir/junction/null/undefined type; POSIX targets, catchable Node errors and named unsupported-overload refusals.
Commits: first piece after 6362fdc64c1c934fd552efb31964a4cf8ab6a033; pushed SHA is reported with the commit.
Checks: lowering PASS 37.648s, native PASS 384.136s, flow PASS 336.771s, freshness PASS 110.493s, directory oracle PASS 7.365s, complete Linux counts regeneration PASS 56.884s.
Mutants: wrong target, wrong error code and wrong error message all caught only by Node stdout, with exit 0 and clean ASan/UBSan and leak checks.
Uncovered: area SHA with mkdtempSync/rmSync has not arrived, so no upstream fixture is claimed newly green; stat const options are the next piece, fixture 25 has a language checker blocker, and macOS was not executed.

The branch remains codex/host-fs-directory-land. No rebase, force-push, main push or PR. The directory runtime now defines _DARWIN_C_SOURCE immediately after _POSIX_C_SOURCE; the area macro fixes will be preserved when merged.

## Behavior and evidence

The fixture checks relative file and directory targets, all admitted type constants, broken links, Dirent identity and stat following. It compares EEXIST, ENOENT, EACCES and empty-target errors byte-for-byte, as well as NUL target/path validation and validation order. Symlink errors retain the existing name/message/code shape and report both target and path. Source, JavaScript and native run in independent writable directories under uid 65534 when root; the input oracle compares remaining file effects and runs sanitizers and a separate leak check.

Broader declared Buffer/URL paths, dynamic type arguments and void-value use stay named NotYet. Types outside @types/node's declaration are checker errors. Supported void calls may appear as discarded statements or returns from void functions, following the fs-file lowering convention. An early signature refusal prevents the generic native-slot invariance check from obscuring symlinkSync's name.

All commands source /workspace/adamic-tools/env.sh and write directly to the retained logs in [logs/fs_directory_symlink](logs/fs_directory_symlink/).

```
go test ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh -count=1 -timeout 30m -parallel 3
go test ./internal/lower -count=1 -timeout 30m -parallel 3
go test ./internal/oracle -run 'TestNodeFSDirectorySymlinkMutants|TestNodeFSDirectoryPermissions|TestInputAgreesWithNode/internal/oracle/testdata/node_(fs_directory|path_)' -count=1 -timeout 30m -parallel 3 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -parallel 3 -args -update-counts
```

The first package run's only failure was the new void-value refusal reporting generic void; the early named refusal fixes it, and the complete lowering rerun passes. JavaScript has no package tests; its behavior is held by the oracle. Initial development failures are preserved, including an unsupported optional Stats method exposing an existing native null dereference and a mutant scratch directory sharing its binary name. The fixture now follows tsc's existing guarded stat pattern; a named refusal for the optional Stats method is part of the stat-options piece.

## Fixture 25 language blocker

Fresh `go run ./cmd/adamic types stage3/fixtures/host/25_readDirectory.a` reports the first diagnostic at 1001:23: TS2345, Argument of type 'T | undefined' is not assignable to parameter of type 'T'. The call is `predicate(array[i], i)` in findIndex's bounds-checked loop. This reproduces without any Node declarations:

```
function find<T>(a: readonly T[], p: (x:T)=>boolean): void { for(let i=0;i<a.length;i++) p(a[i]); }
```

`go run ./cmd/adamic types /tmp/fs-directory-gap-checker.a` records that same TS2345. Other diagnostics include TS2532 at 1041 and 1046 on results indexing, plus TS2322, TS2775 and TS7030. Exact diagnostics are retained. No fixture source, declaration contract or checker option was changed to work around them.

Fixture 24 separately compiles through declarations but stops at NotYet on the const-asserted options binding: `const statSyncOptions={throwIfNoEntry:false} as const; fs.statSync(path,statSyncOptions)`. That is a library lowering gap, to be fixed next.
