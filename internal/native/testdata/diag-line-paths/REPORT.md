# Source path investigation, Stage 1, #gd4csnt

Recommendation: for automatic local source lookup, use debug-only
`-fdebug-compilation-dir=<entry directory>`, subject to Performance accepting
checkout-specific debug unit keys. It is explicit, leaves C bytes alone, and both
entry and sibling-import `os.Stat` checks pass. Do not omit the directory from the
key: the objects differ. If equal debug keys across checkouts remains mandatory,
use debugger-side source mappings instead and accept a configuration step. None
of the tested compiler-only options gives automatic local lookup and equal keys.
No production flags, maps, emitter, lowering or key code changed in this stage.

## Base and current failure

Fetched main `85c9d7837`, which did not contain the accepted directives feature.
Per the brief, this branch is based on `compiler/diag-line-directives-3`,
`6b66274697d01cf7f8c8104d39bfbb713818acab`.

`TestDebugSourcePaths` invokes the real `go run ./cmd/adamic build ... --debug`,
with a 120-second context and a package-test timeout. It reads each CU's
`DW_AT_comp_dir` and line table. Go's DWARF reader already applies file-table
directory indexes; remaining relative names are joined to that CU's comp_dir.
It requires both resolved paths to exist and to equal the real entry and import.
The fixture imports `../lib/value.a` from `entry/fixture.a`.

The unskipped test fails as intended, for source lookup rather than compilation:

```text
comp_dir="/adamic-build"
row="/adamic-build/fixture.a" -> /adamic-build/fixture.a: ENOENT
row="../lib/value.a" -> /lib/value.a: ENOENT
FAIL: debugger paths do not reach entry and imported source
```

The recorded directory is synthetic `/adamic-build`, rather than the deleted
random temporary directory. It is still unusable for opening either source.
The checked-in test has `t.Skip` explicitly naming #gd4csnt, as requested.
[Current failure](current-failure.txt) holds the full unskipped output.

## Where flags and keys come from

- `native.go`, `linkOne`: compiles `main.c` in its temporary directory; maps that
  directory to `/adamic-build` and the runtime cache to `/adamic-runtime`.
- `units.go`, `compileUnitObject`: runs in a per-unit temporary directory.
  `unitkey.go`, `unitCompileArguments`, maps it to `/adamic-units`.
- `library.go`, `prefixMaps`: uses `-ffile-prefix-map` for given and resolved
  directory spellings. This affects debug information and `__FILE__`.
- `library.go`, `clangIn` / `toolMaps` (#8wgk65j): appends maps from the compiler
  installation and sysroot to `/adamic-compiler` and `/adamic-sysroot`.
  Sanitized flags also strip UBSan file paths separately; prefix maps cannot
  reach those metadata strings. Darwin has separate object debug-map handling.
- `unitkey.go`, `newUnitKeying`: v5 keys include every supplied flag verbatim,
  clang's normalized driver facts, headers and their resolution, environment,
  common files, unit name and raw C bytes. The current key is **not** a hash of
  preprocessed bytes. Its compile recipe normalizes the temporary directory to
  `<directory>`. Adding an entry path to flags changes keys legitimately.
- `unitSourceRefusal` currently refuses function/main units containing later
  directives, so debug units with #line compile uncached. The reported keys are
  **candidate keys from the real key function**, not observed cache hits. The
  scratch log prints that refusal as well as all three keys for both copies.

## Scratch results

These are real generated debug C, clang compilations, DWARF reads and `os.Stat`
checks, run through a Go overlay. The scratch compiler invocation preserves the
existing build-directory map and appends the candidate flag after it. It links
against Adamic's runtime. This isolates each flag without editing production.
The scratch key calculation uses the real split units and real key function;
for the directory-remap candidate only the ephemeral input directory is
normalized to `<directory>`, just as the existing compile recipe does. The entry
directory remains keyed. The tests assert the expected path success/failure.

| Option | Entry/import os.Stat | Same C in two copies | Same candidate keys | Release effect if restricted to Debug |
| --- | --- | --- | --- | --- |
| Current relative #line, current maps | Both fail | Yes | Yes | None |
| `-fdebug-compilation-dir=<entry>` | Both pass | Yes | No | None |
| Absolute #line, source root prefix-mapped to `/adamic-source` | Both fail at virtual paths | No | No | None |
| Absolute #line, source root prefix-mapped back to relative names | Both fail under `/adamic-build` | No | No | None |
| Relative #line, actual build comp_dir prefix-mapped to entry directory | Both pass | Yes | No | None |
| Relative #line, `-fdebug-compilation-dir=.` | Both fail from unrelated cwd | Yes | Yes | None |
| Current output plus debugger source mapping | Raw paths fail; simulated mapped paths pass | Yes | Yes | None |

Absolute paths are part of raw C before any prefix map applies, so mapping them
back cannot preserve the C-byte requirement or the existing unit key. Mapping
the compilation directory to the entry directory works but needs the same local
path knowledge as the explicit flag and is less direct. A map of the *already
mapped* `/adamic-build` name is not a recursive second rewrite; the tested map
matches the actual original build directory.

A `.` comp_dir depends on debugger cwd/search behavior, so it does not provide
a reliable automatic answer. Source mapping can preserve existing artifacts and
keys, but needs a checkout-specific setting. Parent-relative imports require
care: apply the entry mapping before cleaning `..`, or map the normalized sibling
path separately. Our source-map result is a filesystem simulation, not an LLDB
behavior claim, because LLDB could not start here.

[results.txt](results.txt) records the complete flags, both full key arrays,
`os.Stat` outcomes and `llvm-symbolizer` results for every option. No hashes are
pinned in a gate. No changes to release compilation are proposed or implemented.

## Debugger check

The toolchain contains LLDB 20, but both attempts to run it (including on the
real CLI artifact) exit 127 before startup:

```text
lldb: error while loading shared libraries: libpython3.10.so.1.0:
cannot open shared object file: No such file or directory
```

Command attempted:

```sh
/workspace/adamic-tools/llvm/bin/lldb --batch \
  -o 'breakpoint set --file fixture.a --line 3' -o run -o 'source list' \
  /tmp/gd4csnt/real-cli
```

There is no gdb on PATH. Used installed `llvm-symbolizer` on known line-table
addresses of the real CLI artifact instead:

```text
0x315b: /adamic-build/fixture.a:3:61
0x31e6: /adamic-build/../lib/value.a:3:0
```

See [lldb.txt](lldb.txt) and [symbolizer.txt](symbolizer.txt). Scratch binaries
were also symbolized at their own DWARF row addresses, not these CLI addresses.
Symbolization proves names/lines, while the separate `os.Stat` check proves the
lookup defect. Interactive break/run/source-list behavior remains unobserved.

## Reproduction and validation

```sh
source /workspace/adamic-tools/env.sh
# Expected failure, removes only the Skip through a Go overlay:
python3 internal/native/testdata/diag-line-paths/run.py --proof > /tmp/paths-fail.log 2>&1
# Expected pass; leaves compiler untouched, writes binaries under /tmp/gd4csnt:
python3 internal/native/testdata/diag-line-paths/run.py > /tmp/paths-scratch.log 2>&1
go test ./internal/native -run '^TestDebug(SourcePaths|LineTable|TwoDirectories)$' -v -count=1 -timeout 300s > /tmp/paths-gate.log 2>&1
go vet ./internal/native > /tmp/paths-vet.log 2>&1
```

Observed: unskipped proof FAIL (1.382s), seven-option scratch PASS (6.06s),
existing line-table/two-directory proofs PASS and new proof SKIP (2.065s),
vet PASS with no output. Gofmt and diff whitespace checks pass.

Setup completed normally: go ready 0.027s, node ready 0.027s, submodules ready
0.088s, tsc ready 0.203s, clang ready 0.210s, cache warm 62.519s, total 62.552s.
`nproc` is 5; cgroup quota is 4 CPUs. clang is 20.1.8.

This Stage 1 makes no production change and runs no production mutants. The
current bug is the negative witness; the two passing scratch configurations
show that the new path checker can pass once both actual paths are reachable.
Full native/repository suites, Darwin execution, live LLDB interaction and real
split-cache hit behavior were not exercised. Stage 2 needs the owner's ruling
on checkout-specific debug keys versus an explicit debugger configuration.
