# Project census membership

A configured TypeScript check keeps every project root and the project's original
options and lib. Requested roots, including explicit ambient roots, remain part of
the program. Only requested implementation files contribute file diagnostics;
configuration and global diagnostics still report. Standalone Adamic continues to
check its complete import graph.

Assumption: callers may explicitly add a root outside the project's include list,
and configuration/global failures must not disappear in a per-file census.
Project `.a` roots use the same `.a.ts` aliases as explicitly requested roots.

The two-file composite namespace-import fixture checks as temporary `.ts` sources:

| Check | Before | After |
| --- | ---: | ---: |
| main alone | 1 (TS6307) | 0 |
| values alone | 0 | 0 |
| both files | 0 | 0 |

A separate negative fixture puts TS2322 in values: main alone reports no error,
values alone and both files each report exactly that error. Existing standalone
import checks still require imported-file diagnostics. Project configuration errors
(including an invalid lib) continue to report.

Run with the setup environment sourced:

```text
go test ./internal/load ./internal/lower -count=1 -timeout 30m
python3 internal/load/testdata/project-census/run_mutant.py
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -timeout 30m
```

The mutant restores the requested-root list and fails TestCompositeProjectCensus
with TS6307. Its Go overlay leaves the working tree unchanged. Logs are
`/tmp/load-census-before.log`, `/tmp/load-census-after.log`,
`/tmp/load-census-packages-final.log`, `/tmp/load-census-mutant.log`, and
`/tmp/load-census-oracle.log`.

Setup: Node 24.19.0, Go 1.27.1, clang 20.1.8, WASI SDK 27; nproc 5 (quota 4).
Timings: Go 0.021s, Node 0.024s, submodules 0.070s, markdown 0.074s,
clang 0.174s, WASI 0.205s, total 69.748s. npm ci for stage3/api succeeded.
