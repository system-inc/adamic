# Stats optional-call missing-guard witness

The control is `TestInputAgreesWithNode/internal/oracle/testdata/node_fs_stats_optional_call.a`.
It agrees byte for byte with Node on native and JavaScript, including the file,
directory, seven undefined results, and `Stats | Dirent` rows. The receiver count
is 24: seven printed calls and one discarded call on each of three paths.

For the mutant, replace the shared `optionalIntrinsic` conditional's nullish
condition with `ir.BooleanConstant{Value: true}`, preserving its saved receiver.
Run:

```
go test ./internal/oracle -run '^TestInputAgreesWithNode$/internal/oracle/testdata/node_fs_stats_optional_call.a$' -count=1 -v
```

Observed red: the first missing receiver is the `isFile` expression at source
line 9, stdout line 15. Node prints `undefined` and exits 0. Native stops after
14 stdout lines, exits 1, and reports `node_fs_file.c:533:86: runtime error:
member access within null pointer`. JavaScript also stops after 14 stdout
lines, exits 70, and reports `TypeError: Cannot read properties of undefined
(reading 'isFile')`. The control passes after restoring the guard.

`TestNodeFSStatsOptionalCallNullishMutant` permanently removes all eighteen guards
in the lowered fixture, retains their receiver declarations, and requires both
native and JavaScript to disagree with Node. No mutant is retained in production.
Counts were generated with `TestCountsAreRecorded` and `-update-counts`.

## Local guarded-path mark witness

`TestNodeFSStatsOptionalReceiverRequiresGuard` calls `nodeFSFile` directly,
without `optionalIntrinsic`, and requires the exact StatsBase optional-call stop.
The catch line is `unguarded nodeFSFile must refuse node:fs.StatsBase.isFile through an optional call`.
`TestNodeFSStatsOptionalReceiverMarkMutantIsCaught` uses `buildcache.GoTestMutated`
to remove only the mark check. It requires that direct test to fail at the catch
line, while the full-Lower control and the native, JavaScript and WASI fixture
controls remain green. The shared receiver nullish guard is retained in this mutant.

## Dirent guarded-path mark witness

`TestNodeFSDirentOptionalReceiverRequiresGuard` calls `nodeFSDirectoryCall`
directly and requires `node:fs.Dirent.isFile through an optional call`.
Its catch line is `unguarded nodeFSDirectoryCall must refuse node:fs.Dirent.isFile through an optional call`.
`TestNodeFSDirentOptionalReceiverMarkMutantIsCaught` uses `buildcache.GoTestMutated`
to remove only the Dirent mark check. Both mark mutants retain the shared
nullish guard; their controls require full Lower and the complete entry fixture
(native, JavaScript and WASI) to remain green. Optional function calls and
continuations remain refused.

The expanded fixture covers optional `Stats | Dirent | undefined` predicates
and all seven optional `Dirent | undefined` predicates, including an undefined
receiver and a discarded call. The permanent IR nullish-guard mutant removes
eighteen guards; the first missing Stats receiver remains source line 9,
stdout line 15.

## Unmodified-main Dirent witness

At main `0406bbd325c4a914bac93ac4f42b133077dc432b`, `first(d: Dirent | undefined) { return d?.isFile(); }`
called first with undefined and then a real directory entry is admitted. Node
prints `undefined\ntrue\n`, stderr empty, exit 0. Sanitized native prints
nothing, reports a null-pointer member access at `node_fs_file.c:533:86`, and
exits 1. JavaScript prints nothing, reports `adamic: panic: TypeError: Cannot
read properties of undefined (reading 'isFile')`, and exits 70. The working
directory contains a single real file. Node's declarations pass optionalCall's
non-library exemption; the Dirent host arm previously called without a guard.

Observed on Linux with the restored WASI SDK environment: both mark-mutant
catchers pass. Their direct-test subprocesses exit 1 with these catch lines:

```
library_node_fs_stats_optional_test.go:44: unguarded nodeFSFile must refuse node:fs.StatsBase.isFile through an optional call: handled=true value=ir.NodeHostCall err=<nil>
library_node_fs_dirent_optional_test.go:39: unguarded nodeFSDirectoryCall must refuse node:fs.Dirent.isFile through an optional call: handled=true value=ir.NodeHostCall err=<nil>
```

For each mutant, the full-Lower control and the same expanded fixture remain
green on native, JavaScript and WASI. The unmutated fixture also agrees with
Node on all three backends. Its counts were regenerated on Linux with:

```
go test ./internal/oracle -run 'TestCountsAreRecorded/internal/oracle/testdata/node_fs_stats_optional_call.a$' -count=1 -args -update-counts
```
