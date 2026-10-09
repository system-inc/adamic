V3 soundly admits the narrowed array; the stale refusal test now checks passing and misfit reads.
The admission changed in b6093b70; the initial V3 base was 14690251.
Node and all backends pass the valid case; the map misfit exits 70 in both backends, including sanitized native.
Removing the array guard and restoring the old refusal each fail their owning test.
This unit checks array narrowing from an array/map union; other object representations retain their refusals.

The old test feeds `internal/oracle/testdata/reland_refused/narrowed_union_object_tag.a`.
Its `shared` slot has type `number[] | Map<string, number>` and is narrowed to an array at the length read.
The original test on V3 reproduces `want refusal naming the path and fix, got <nil>`.
Commit b6093b70 changes `localRead` in `internal/lower/locals.go` to bypass the typeof-only refusal for arrays.
The same function emits ArrayIsArray before Narrow. JavaScript emits Array.isArray;
native checks the retained union object's kind against adamic_kind_array before reading array storage.
Thus the old refusal is obsolete for this shape. No production change or weaker check is needed.

The passing test runs the exact original source. Node, generated JavaScript, release native and
ASan/UBSan native exit 0 with stdout `1\n` and empty stderr.
The misfit test adds `change()` between `shared = [1]` and the length read;
that function stores an empty Map in the same slot. Node exits 0 with `undefined\n`.
All Adamic runs exit 70 with empty stdout and:
`adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back`.

The remove-array-guard Go overlay omits the panic guard for narrowed arrays without touching storage.
The misfit test fails: both native modes exit 0 with `0\n`, while JavaScript exits 0 with `undefined\n`.
This is an executable failure of the exact guard, rather than a build or sanitizer failure.
The restore-old-refusal overlay restores the pre-b6093b70 condition; the passing test fails with the old
path, object-tag refusal and separately-typed-variables fix at line 7 column 16.
The related TestNarrowedUnionMemberCheck and both new leaves pass without overlays.

Cold measurements run each top-level leaf separately with ADAMIC_GATE_UNCACHED=1, GOMAXPROCS=4,
-count=1 and an empty XDG_CACHE_HOME. The Go dependency cache prepared by setup is retained;
wall time includes go test startup, runtime libraries, lowering, both native builds and JavaScript/Node.
See timings.json and each leaf's JSON log. Both leaves are below 60 seconds.

Setup completed in 44.682 s; go ready 0.028 s, Node ready 0.030 s, submodules 0.067 s,
markdown dependencies 0.089 s, clang 0.187 s, go build 44.392 s.
nproc is 5; cpu.max is 400000/100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node 24.19.0.
The first probe ran out of disk space; removing 2,180,745,976 bytes of old regenerable Go cache allowed the rerun.
No fixture registry entries were added, so counts.md is unchanged.
No whole package tests or full gate ran. Lane checks and the final merged-base verification are recorded separately.
