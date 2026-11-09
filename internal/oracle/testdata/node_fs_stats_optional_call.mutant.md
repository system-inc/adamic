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
line. The ordinary full-Lower pins run in the lower unit; the native, JavaScript
and WASI fixture controls run in the oracle unit. The catcher builds and runs only
the mutated lower binary. The shared receiver nullish guard is retained in this mutant.

## Dirent guarded-path mark witness

`TestNodeFSDirentOptionalReceiverRequiresGuard` calls `nodeFSDirectoryCall`
directly and requires `node:fs.Dirent.isFile through an optional call`.
Its catch line is `unguarded nodeFSDirectoryCall must refuse node:fs.Dirent.isFile through an optional call`.
`TestNodeFSDirentOptionalReceiverMarkMutantIsCaught` uses `buildcache.GoTestMutated`
to remove only the Dirent mark check. Both mark mutants retain the shared
nullish guard. Each catcher requires only its direct-call pin to fail; the full-Lower
pins and the complete entry fixture run in their own units. Optional function calls
and continuations remain refused.

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

Earlier observation on Linux with the restored WASI SDK environment: both mark-mutant
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

## Lower-unit isolation

At `a2a96e312582ca1a354d5016743d7d5ca7a43083`, both direct-call mutants are
caught and both full-Lower controls pass, but the oracle control fails with:

```
input_test.go:126: node --version: exec: "node": executable file not found in $PATH:
```

This was observed with a fresh HOME and Go cache, PATH containing only Go's bin
and `/usr/bin:/bin`, and `GOFLAGS=-mod=mod`. This checkout needs `GOWORK=off`
for that module mode; module downloads were populated before the stripped run.
The private Go builds also need `CC=gcc` on this box: toolchain.Env otherwise
selects clang, which is absent from that PATH. No Node, WASI or pinned-source
settings were supplied. The oracle cache reports zero native, node and probe
hits and misses because its identity check fails before a fixture runs.

The lower catchers now build only their mutated lower binary and require the
direct pin to exit 1 with its catch line. They neither build nor run an oracle
binary, and add no skip or environment switch. The existing ordinary full-Lower
pins and oracle fixture rows retain their independent checks.
