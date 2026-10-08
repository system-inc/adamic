# Step 29 filesystem unit

Branch `codex/host29-fs`, based on library `71f91687` with host surface
`91215715` merged. Scout contract: `2cdc201b`. Node v24.19.0, Linux.
There is no AGENTS.md in this checkout or its parents. Repository instructions
were read from CLAUDE.md; the exact scout sources are unchanged.

Public filesystem descriptors are numbers backed by an owned-handle table.
Open acquires a handle; close removes it. Repeated close and stale reads/writes
throw catchable EBADF Errors. Inherited descriptors 0, 1 and 2 participate in
the same table. Private filesystem operations keep their temporary descriptors
private. Numeric validation remains Node's validation before ownership lookup.

Filesystem system Errors now retain errno, syscall and path alongside code,
name and message. Descriptor errors have no path field; plain and argument
Errors have no system fields. Host Error class metadata visits the additional
owned fields during destruction, and dynamic numeric reads have scalar type
metadata. WASI errno values are translated to the Linux Node host's numbers.
The now-backed errno/syscall/path presence probes are admitted; stack and cause
retain their existing refusals.

The existing Stats/Dirent dispatch, readdirSync withFileTypes, realpathSync and
.native, statSync throwIfNoEntry, recursive mkdir races, unlink and Date/utimes
implementations remain held by their filesystem oracle fixtures. Date values
remain epoch milliseconds; numeric utimes arguments remain seconds.

## Exact scout fixture results

| Fixture | Native and JavaScript | WASI |
| --- | --- | --- |
| 05 writeFile | Refused: unchecked catch value cast to NodeJS.ErrnoException, line 37 | Same compiler refusal |
| 06 fileExists | NotYet: property-valued switch case, line 34 | Same compiler refusal |
| 07 directoryExists | NotYet: property-valued switch case, line 34 | Same compiler refusal |
| 08 getDirectories | NotYet: destructured callback parameter, line 144 | Same compiler refusal |
| 09 realpath | Agrees with Node, native sanitizers/leaks clean | Compile-time fs.mkdtempSync target refusal |
| 10 getModifiedTime | NotYet: optional getTime call, line 32 | Same compiler refusal |
| 11 setModifiedTime | Agrees with Node, native sanitizers/leaks clean | Compile-time fs.mkdtempSync target refusal |
| 12 deleteFile | Agrees with Node, native sanitizers/leaks clean | Compile-time fs.mkdtempSync target refusal |
| 13 createDirectory | Checker TS18046: catch value e is unknown, line 55 | Same checker refusal |
| 24 useCaseSensitiveFileNames | NotYet: property-valued switch case, line 37 | Same compiler refusal |
| 25 readDirectory | Checker: unchecked indexed generic/string reads and assertion-call annotation (TS2345, TS2532, TS2322, TS2775) | Same checker refusal |

The eleven-fixture acceptance runner exits 1 because blocked fixtures are never
passes. It does not rewrite sources, loosen checker flags, supply substitute
System implementations or approximate absent compiler support.

## Verification and mutants

`stage3/host/check_fs_unit.py --compiler /tmp/host29-fs-adamic-final --logs
/tmp/host29-fs-acceptance-final --mutants` records all three backends and all
original Node observations. All eleven Node observations match the scout.
Ten source mutants differ only in stdout: remove BOM (05), swap classification
(06/07), reverse directories (08), change fallback resolution (09), omit mtime
(10), use epoch times (11), omit unlink (12), invert case sensitivity (24), and
omit extension filtering (25). The scout's skip-create mutant (13) differs in
Node exit/stderr too; it is not a clean native mutant proof for the blocked
source.

All 37 filesystem runtime mutants pass their clean-execution requirements and
are caught only by source Node stdout comparisons, including the three added
mutants clearing errno, replacing syscall, and replacing path. The syscall
mutant operates only on system-error shapes, leaving argument Errors intact.
The filesystem directory mutants and shape/union/symlink/options regressions
also pass the focused gate.

The new `node_fs_file_owned_errors.a` fixture agrees on native, JavaScript and
WASI. It covers close/stale operations, reopening, inherited stdout, errno,
syscall, path, missing paths, mkdir/unlink/utimes/readdir/realpath errors, and
absence of system fields on plain and NUL-path argument Errors. Native ASan,
UBSan and leak checks pass.

WASI also agrees on the seven existing directory fixtures (symlink, entries,
union, realpath, System, stat options and permissions), and file write, close
and unlink fixtures. The existing writeFile permission-mode, mkdir permission-
mode and subsecond/out-of-range utimes probes retain explicit WASI refusals;
these skips are not passes. No exact scout fixture passes all three backends.

Focused tests passed in internal/lower, internal/native and internal/oracle;
internal/javascript has no package-local tests, and its emitted programs are
compared in the oracle. Vet covered only those four packages. Linux
TestCountsAreRecorded refreshed counts.md with the new fixture and all affected
existing rows. The new row is 164 allocations, 164 frees, 119 retains,
236 releases, 10 peak live and 0 values in regions. No whole-repository test
gate was run.

Run output stays outside the checkout: /tmp/host29-fs-focused-final.log,
/tmp/host29-fs-tests-final.log, /tmp/host29-fs-wasi.log,
/tmp/host29-fs-wasi-file.log, /tmp/host29-fs-wasi-fields-final.log,
/tmp/host29-fs-counts-final.log, /tmp/host29-fs-vet.log and
/tmp/host29-fs-acceptance-final/.
