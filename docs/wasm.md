# Adamic's WASI runtime

The runtime builds for `wasm32-wasi` (WASI Preview 1), using clang and a WASI
sysroot. Every runtime difference uses `ADAMIC_TARGET_WASI`, set with
`-DADAMIC_TARGET_WASI=1`. The compiler's target and driver are a separate unit;
this branch changes neither `native.go` nor the driver. Executable programs
still enter through `main`.

`wasm32-unknown-unknown` alone is insufficient: the runtime needs allocation,
libm, byte/string functions, output and process exit. WASI libc supplies those,
and translates the existing POSIX-shaped file operations to WASI imports.
A libc-free build would need replacements for those services, not just another
clang target flag.

## Portability inventory

| Area | WASI behavior |
| --- | --- |
| `adamic.c` signals | SIGPIPE handling and SIGTERM/SIGINT/SIGHUP flushing are compiled out. WASI has no process signals; there is no public runtime signal API to reach. Ordinary output, flush, explicit panics and exit 70 remain. |
| Atomics | The only runtime atomics here are `atomic_signal_fence` in the output buffer. They compile unchanged; these are compiler fences, not shared-memory synchronization. |
| Threads, pool, `_Thread_local`, `parallel.c` | None exists in the base checkout (`ef3d907`). No threading API was stubbed and no sequential `parallelMap` fallback was implemented. The compiler unit must emit a sequential loop, like the Node witness, when that operation arrives. |
| `stack.c` | No `getrlimit`. The constructor uses wasm-ld's `__stack_low` plus a 16 KiB panic margin. Generated `ADAMIC_CHECK_STACK` checks linear-memory frames. The tests reserve a 128 KiB stack and independently protect its bottom 32 bytes with a canary. |
| Engine call stack | Functions without linear-memory frames can exhaust the engine's separate call stack first. Two oracle fixtures demonstrate this limitation below. A balanced call-depth guard in compiler output, or a tested host strategy preserving buffered output and panic semantics, is still needed. |
| `input.c` | `open`, `read`, `write`, `close` and `fstat` go through WASI libc. Empty read/write paths are normalized to ENOENT; `fstat` identifies a directory before `fd_read`, whose EBADF otherwise differs from Node's EISDIR. UTF-8 decoding is unchanged. |
| `directory.c` | `opendir`, `readdir`, `lstat` and `stat` go through WASI libc. Directory listing and file status work inside preopened directories. No host filesystem is granted implicitly. |
| `sort.c` | Ordinary TimSort is unchanged. WASI SDK's `setjmp`/`longjmp` requires experimental wasm exception handling. That path is compiled out; a comparator that throws panics with `wasm32: throwing sort comparators are not supported`, exit 70. Ordinary emitted throw/catch cleanup paths and region cleanup still work. |
| `weak.c` | Pointer hashing widens a wasm32 pointer to `uint64_t` before shifting by 33. Weak lifetime, reuse and ownership behavior are unchanged. |
| Sanitizer-only paths | `heap.c` detects AddressSanitizer and includes its poisoning interface only in native sanitized builds. The normal WASI build does not enter these paths. Native ASan/UBSan flags are not supported by this WASI toolchain and are not used for wasm. Counted wasm builds supply the independent lifetime probes. |
| `tsgo.c` | The external native Go-checker bridge remains disabled unless `ADAMIC_TSGO` is explicitly selected. A native Go archive cannot be linked into this WASI module. |

All 48 runtime translation units compile with `-std=c11 -Wall -Wextra -Werror
-pedantic`, without the generated-C unused-variable/function exemptions. The
five changed runtime files produce byte-identical native objects to the base
with native clang `-O2`; the native package gate and full native oracle pass.

## Running the opt-in test

Run `bash cloud/setup.sh`, then source the environment file it prints. Install
[WASI SDK 27](https://github.com/WebAssembly/wasi-sdk/releases/tag/wasi-sdk-27)
or supply an equivalent clang, linker, compiler builtins and WASI sysroot.
Node 24 supplies the command runtime and the independent source oracle.

```sh
source /workspace/adamic-tools/env.sh
export PATH=/workspace/adamic-tools/wasi-sdk-27.0-x86_64-linux/bin:$PATH
export WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk-27.0-x86_64-linux/share/wasi-sysroot
ADAMIC_TEST_WASI=1 ADAMIC_WASI_ARTIFACT=/tmp/adamic-request.wasm \
  go test ./internal/native -run '^TestWASI$' -count=1 -v -timeout 15m \
  > /tmp/wasm-test.log 2>&1
```

Paths are examples; use your installation. Without opt-in, Node 24 or a usable
WASI toolchain, the test skips with the missing requirement. Once the minimal
toolchain probe succeeds, runtime compilation failures are failures, never
skips. The test invokes `adamic c <fixture>` from the existing compiler, compiles
all runtime C itself, and links each fixture with:

```sh
clang --target=wasm32-wasi --sysroot="$WASI_SYSROOT" \
  -DADAMIC_TARGET_WASI=1 -std=c11 -Wall -Wextra -Werror -pedantic \
  -O2 -ffp-contract=off -fno-optimize-sibling-calls \
  -Wl,-z,stack-size=131072 -Wl,--export=__stack_low \
  -I runtime-directory main.c runtime-objects/*.o -lm -o main.wasm
node --disable-warning=ExperimentalWarning internal/native/wasm/run.mjs main.wasm
```

Generated C has the same unused-code warning exemptions as native builds.
`run.mjs` preopens only its working directory and checks the reserved-stack
canary after `main` returns or exits through WASI. Oracle command objects are
uncounted; the reactor's objects are rebuilt counted for lifetime measurements.
`ADAMIC_WASI_RUNTIME` selects a scratch runtime directory for isolated mutants.

The test lists 33 fixtures: 30 agree exactly in stdout, stderr and exit status.
They cover all ten 0.1 examples, strings, numbers, bitwise operations, functions,
regions and region exceptions, Weak and reuse, normalization, maps, JSON,
RegExp collections, Object keys, array flattening, file reads and a file/directory
round trip. `read_files.a` runs in its own fixture directory, including malformed
UTF-8, BOMs, NULs, missing files, a directory and an empty path.

The other three are explicitly classified and checked, not counted as equivalent:

| Fixture | Observation |
| --- | --- |
| `closures_throw.a` | Output agrees until the first throwing sort comparator; WASI then gives the explicit unsupported panic and exits 70. Node catches the comparator's error and continues. |
| `stack_forever.a` | Node prints `start`, then the Adamic stack-overflow panic and exit 70. WASI's engine raises RangeError, loses buffered stdout, and exits 1. |
| `stack_tail_call.a` | Node prints `0`, then the same panic and exit 70. WASI has the same engine-stack mismatch as above. |

The opt-in test passes only if the equivalent fixtures agree and these three
known limitations produce their specifically checked observations. Unknown
differences fail. `stack_overflow.a`, which allocates linear stack frames,
agrees with Node and preserves the canary. This is not full wasm oracle coverage.
Permission probes, paths outside preopens, symlink escape behavior, special files,
empty directory/status paths, very large allocations and every library fixture
are not covered here.

## Request ABI and host

The exported application entry is:

```c
adamic_string *adamic_request(const unsigned char *bytes, size_t length);
```

On wasm32 this is `(i32, i32) -> i32`: a borrowed pointer and byte length of UTF-8
request data, returning one owned Adamic string pointer. The function decodes
the request using `adamic_decode_utf8`, calls the handler, releases the decoded
request and returns the response. Invalid UTF-8 uses the existing WHATWG
replacement rules. The host copies the result before releasing it. Adamic
response bytes are WTF-8; the host's TextDecoder replaces lone surrogates.

The supporting exports are `memory`, `malloc`, `free`, `adamic_release`,
`adamic_response_bytes(const adamic_string *) -> const char *` and
`adamic_response_length(const adamic_string *) -> size_t`. All pointer/length
results are wasm32 i32 values. `adamic_live()` and `adamic_regions()` are counted
prototype probes, not required application ABI. Input buffers are host-owned;
responses must be released exactly once. Reacquire memory views after a wasm
call, since allocation can grow memory.

`internal/native/wasm/request.a` is the source witness. Its handler builds and
counts a 63-object tree inside a declaration's statement region and returns
`hello <request>: 63`. The tree dies when that statement ends; the response is
an ordinary owned heap string and survives the region. No current-region global
and no garbage collector is added.

`request-abi.c` is a temporary adapter around today's executable C output,
selecting the emitted handler by name. Only this test adapter renames its
included fixture `main` and links with `-mexec-model=reactor`; ordinary programs
keep `main`. The compiler unit can replace the adapter by emitting the export
and module initialization directly. Node calls `_initialize` exactly once before
serving requests. This fixture has no module initialization to replay beyond
runtime constructors; a general module will need its top-level initialization.

`request-host.mjs` implements a Worker-shaped async `fetch(Request)` returning
`Response`. The synchronous handler runs after the body is read, with no
filesystem preopens. The test verifies a fetch call, warms up with varied ASCII,
Unicode, 1 KiB and NUL-containing requests, then checks 100,000 responses. Between
every request it requires zero live counted values and no linear-memory growth;
it also requires exactly 6,300,000 objects ended through regions. Linear memory
is reusable capacity and cannot shrink; flat capacity alone would not prove
correct release.

The host also imports the very same `.a` source with Node's type stripping and
measures its plain TypeScript handler. Rates include response assertions; wasm
also includes UTF-8 copies, ownership calls and memory/lifetime probes. They are
synchronous handler rates, not HTTP throughput or a Cloudflare performance claim.
The observed numbers and mutants are recorded below.

## Compiler and Worker integration

The compiler target needs the WASI sysroot, `ADAMIC_TARGET_WASI=1`, wasm-compatible
archive tools if it builds an archive, libm, the existing C11 flags, and a linker
stack larger than the 16 KiB panic margin. Preserve sibling-call prevention and
floating-point contraction settings. Initialize constructors before invoking
exports. Keep the existing ownership and statement-region conventions; a response
or other escaping value must never point into an ended request region. Compile
`parallelMap` sequentially when it is introduced, and refuse or clearly panic
for unsupported sort exceptions. The separate engine-stack limit needs a guard
that also works when a function needs no linear-memory frame.

The stripped request artifact imports exactly `fd_close`, `fd_prestat_get`,
`fd_prestat_dir_name`, `fd_seek`, `fd_write` and `proc_exit` from
`wasi_snapshot_preview1`. Its other runtime work uses linear memory and wasm
instructions. A command or a handler using filesystem input will import more.

A real Worker would add an ES-module wasm binding/bundling configuration,
instantiation once per isolate, and a tested WASI Preview 1 import adapter for
allocation startup, output and `proc_exit`. Workers do not supply Node's
`node:wasi`; this host is a local stand-in, not a deployment. Give filesystem
imports explicit behavior with no preopens, route output to host logging, and
translate a panic into a request failure without trying to reuse an instance
whose WASI execution has terminated. Do not share an instance across simultaneous
wasm execution or native threads; runtime counters, caches and pending exceptions
are single-threaded. Fetch bodies may await before the synchronous wasm call.
Measure deployed module size, cold startup, isolate reuse and request throughput
on the actual Worker. No Cloudflare account or deployment was used here.

## Measurements and mutants

Observed on the shared cloud machine with Node 24.19.0 and WASI SDK 27. The
prototype is 272,389 bytes, or 48,940 bytes after `llvm-strip`. Stripping preserves
all exports and the full 100,000-request assertions. The counted prototype's
linear memory remains exactly 393,216 bytes (six 64 KiB pages), with zero live
counted values between requests and 6,300,000 objects ended through regions.

One complete run measured 202,734 wasm requests/s against 904,284 plain Node
requests/s, with 1.86 ms for read/compile and 2.03 ms for instantiate/initialize.
A later run while native tests were also running measured 143,098 against
771,085, with 7.45 and 5.19 ms startup stages. The stripped module measured
138,543 against 429,829 under concurrent load. These are observations on a
shared CPU, not a controlled performance comparison. The wasm boundary and
counted probes are slower for this handler; no speedup is claimed.

Every mutant below compiled and was caught by executing the named check:

| Mutant | Check that caught it |
| --- | --- |
| Console newline replaced by `!` | `01_hello.ts`: exact stdout comparison. |
| Panic prefix replaced | `panic.a`: exact stderr comparison. |
| Panic `_exit(70)` changed to `_exit(0)` | `panic.a`: exit status comparison. |
| Empty-path normalization disabled | `read_files.a`: empty-read message; `wasm/io.a`: empty-write message. |
| Directory-read detection disabled | `read_files.a`: EISDIR message becomes generic failure. |
| A live Weak target always reads absent | `weak_parent.a`: exact stdout comparison. |
| Throwing sort comparator continues sorting | `closures_throw.a`: required explicit unsupported panic is missing, and sort output changes. |
| Linear stack limit set to zero | `stack_overflow.a`: reserved-stack canary is overwritten. Initially stdout/stderr/exit alone missed this mutant; that observation led to the independent canary. |
| Region block `free` omitted | Request loop: linear memory grows after warmup, despite zero logical live counts. |
| `adamic_release` never lets go | Request loop: live allocation assertion fails on dynamically allocated request/response data. |
| Region live count decremented without recording region objects | Request loop: memory and live checks pass, but region count is 0 instead of 6,300,000. |

An initial sort mutant also changed the out-of-memory branch and was stopped by
clang; it was corrected to alter only the comparator's throw path, and only the
runtime failure is counted above. An initial subtest filter selected no fixtures;
those green runs were discarded and all listed witnesses rerun with full paths.

Native verification used the native clang PATH, separately from WASI SDK:

```sh
go test ./internal/native/... -count=1 -timeout 15m > /tmp/wasm-native-final.log 2>&1
go test ./internal/oracle -count=1 -timeout 30m > /tmp/wasm-oracle-full-final.log 2>&1
go vet ./... > /tmp/wasm-vet-all.log 2>&1
gofmt -l cmd internal > /tmp/wasm-format.log
```

The native package run passed in 87.711 s and the complete oracle package in
69.672 s; vet and format checks produced no output. Both new `.a` fixtures were also built with `--sanitize` and compared
to Node: exit 0, identical stdout, empty stderr. The complete repository test
gate was not run; native and oracle packages were the test scope.

Setup printed `go ready`, `clang ready`, `node ready` and `submodules ready` at
0 s, `build cache warm` at 106 s, and `done in 106s on 5 processors` (`nproc` = 5,
cgroup quota four CPUs, 17.6 GB reported memory). Go was 1.27.1 and native clang
20.1.8. The setup environment was `/workspace/adamic-tools/env.sh`.

## Imports

W2 inventory, October 7, 2026, on `wasm/integrate` at
`6f7dce3dc1eace606fe081c8f4ab12034ae11bb4`, WASI SDK 27 and Node 24.19.0.
The inventory uses `WebAssembly.Module.imports` on each actual linked module,
filters `wasi_snapshot_preview1`, and sorts names. No imports are inferred from
C sources. Commands use `native.Build` with `wasm32-wasi`; the compiler request
reactor uses `native.WASI` and `Request: true, Count: true`. The earlier request
prototype comes from `TestWASI/requests` with `ADAMIC_WASI_ARTIFACT`.

These are exact sets, not subsets. The fixture table assigns one set per module.

| Set | Exact `wasi_snapshot_preview1` imports |
| --- | --- |
| A | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |
| B | `args_get, args_sizes_get, fd_close, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |
| R | `fd_close, fd_fdstat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |
| P | `fd_close, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |

`internal/native/wasm/shim.mjs` is 38 physical lines including comments and
blank lines, has no Node imports, and exports `wasiShim({ stdout, stderr })`.
Instantiate with `shim.imports`, call `shim.attach(instance)`, then call `_start`
for a command or `_initialize` once for a reactor before using request exports.
The caller must discard an instance after `AdamicExit`; the shim does not unwind
or reset libc state. Callbacks synchronously consume owned `Uint8Array` copies.
They receive raw bytes, preserving NULs, malformed UTF-8 and iovec boundaries.
Do not decode each callback separately if UTF-8 may span iovecs.

| Import | Choice and reason |
| --- | --- |
| `fd_write` | Decode all wasm32 little-endian pointer/length iovecs, copy their bytes to stdout for fd 1 or stderr for fd 2, write the total byte count, return 0. Other fds return EBADF (8). Reacquire memory views for every call and after callbacks. |
| `proc_exit` | Throw exported `AdamicExit`, with `.name = 'AdamicExit'` and `.code` set to the WASI exit code, so termination cannot silently resume. |
| `fd_prestat_get` | EBADF (8), so libc discovers no preopened directories. |
| `fd_prestat_dir_name`, `fd_close`, `fd_seek`, `fd_fdstat_get` | EBADF (8). No directory, seekable file or terminal descriptor metadata is available. Output still works via `fd_write`. |
| Every other `fd_*` and `path_*`, including `fd_read`, `fd_filestat_get`, `fd_readdir`, `path_open` | EBADF (8), without touching output memory or granting host file access. |
| `args_sizes_get`, `environ_sizes_get` | Write zero count and zero buffer size, return 0: this host has an empty argument vector and environment. |
| `args_get`, `environ_get` | Return 0 without writes: zero entries require zero bytes. |
| `clock_time_get`, `random_get` | Throw `AdamicUnsupportedWASI` naming the import. Fake time and predictable random bytes would not satisfy those services. Neither occurs in the inventoried modules. |
| Other imports | Throw `AdamicUnsupportedWASI` naming the import; no success stub. |

No inventoried file-free fixture requires an import that the shim cannot
satisfy. File fixtures cannot run honestly without file access. The test skips
modules importing any `path_*`, `fd_read`, `fd_readdir` or `fd_filestat_get`.
That conservative artifact rule skips exactly these three fixtures:

| Skipped fixture | Exact imports |
| --- | --- |
| `internal/oracle/testdata/write_stdout_order.a` | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, path_open, proc_exit` |
| `internal/oracle/testdata/write_stderr_order.a` | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, path_open, proc_exit` |
| `internal/oracle/testdata/prompt_then_read.a` | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_read, fd_seek, fd_write, path_open, proc_exit` |

The differential command test uses the same expected witness as
`internal/oracle/wasi_test.go`: Node running source for ordinary fixtures, and
`onJavaScriptBackend` for the 10 fixtures with `checked=true`. The backend
carries Adamic's inserted checks; source Node continues where those checks panic.
The test reports both commands run and actual three-way agreements, plus skips.
The corrected strict run passed with `run=305 agreed=305 skipped=3`: 295
source witnesses and 10 checked backend witnesses, with no mismatches.

`TestWASIShimCheckedWitnessControl` holds `writes_past_end.a` to the checked
backend and separately compares it to source Node. The backend and shim must
agree with exit 70; source must exit 0 and produce `exit codes differ`. Choosing
the source witness for a checked fixture therefore remains a detected error.

`TestWASIShimRequest` compares the compiler reactor under the shim, Node WASI,
and the source handler: empty input, Unicode, echo/reassignment, 1 KiB input,
NULs, dependency/module initialization, and stable counted live values relative
to the module's initialized globals. It passes. The earlier prototype's
100,000-request test also passes under its existing Node WASI host.

| Mutant | Named check that fails |
| --- | --- |
| `fd_write` omits the last iovec | `TestWASIShimContractsAndMutants/iovecs`, assertion `all iovecs, exact bytes` |
| `proc_exit` returns instead of throwing | `TestWASIShimContractsAndMutants/exit`, assertion `exit must terminate` |
| `fd_prestat_get` returns success and a fake directory | `TestWASIShimContractsAndMutants/preopens`, assertion `no preopened directory` |

Every control passed; every isolated mutant process exited 1 with the named
AssertionError. The test also checks empty environment/arguments, named traps,
stdout and stderr independently, byte counts, and memory growth.

Run the command comparison and checked-witness, contract, and request checks with:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIShim' \
  -count=1 -v -timeout 30m > /tmp/wasm-shim-oracle.log 2>&1
```

`ADAMIC_SHIM_INVENTORY_ONLY=1` collects the command import table without running
comparisons. Do not use that option as evidence of execution parity.

| File-free module | Import set |
| --- | --- |
| `internal/native/wasm/request.a` prototype reactor | P |
| `cmd/adamic/testdata/wasi/request.a` compiler reactor | R |
| `internal/oracle/testdata/call_targets_element.a` command | A |
| `internal/oracle/testdata/call_targets_region.a` command | A |
| `internal/oracle/testdata/call_targets_reuse.a` command | A |
| `internal/oracle/testdata/call_targets_closure.a` command | A |
| `internal/oracle/testdata/call_targets_sort.a` command | A |
| `internal/oracle/testdata/library_object_keys.a` command | A |
| `internal/oracle/testdata/library_object_is.a` command | A |
| `internal/oracle/testdata/library_object_has_own.a` command | A |
| `internal/oracle/testdata/library_object_assign.a` command | A |
| `internal/oracle/testdata/library_object_freeze.a` command | A |
| `internal/oracle/testdata/library_object_freeze_write.a` command | A |
| `internal/oracle/testdata/library_object_order.a` command | A |
| `internal/oracle/testdata/library_object_assign_fields.a` command | A |
| `internal/oracle/testdata/library_object_freeze_alias.a` command | A |
| `internal/oracle/testdata/library_object_freeze_assign.a` command | A |
| `internal/oracle/testdata/library_object_same.a` command | A |
| `internal/oracle/testdata/library_object_own.a` command | A |
| `dedication/dedication.a` command | A |
| `cmd/adamic/testdata/wasi/request.a` command | A |
| `internal/oracle/testdata/class_oct6_deep.a` command | A |
| `internal/oracle/testdata/class_oct6_parameters.a` command | A |
| `internal/oracle/testdata/class_oct6_release.a` command | A |
| `internal/oracle/testdata/class_oct6_subclass_holder.a` command | A |
| `internal/oracle/testdata/library_array_join.a` command | A |
| `internal/oracle/testdata/library_array_iterators.a` command | A |
| `internal/oracle/testdata/library_array_metadata.a` command | A |
| `internal/oracle/testdata/library_array_with.a` command | A |
| `internal/oracle/testdata/library_array_flat_map.a` command | A |
| `internal/oracle/testdata/library_array_flat.a` command | A |
| `internal/oracle/testdata/library_array_spliced.a` command | A |
| `internal/oracle/testdata/library_array_copy_within.a` command | A |
| `internal/oracle/testdata/library_array_search.a` command | A |
| `internal/oracle/testdata/library_array_copy.a` command | A |
| `internal/oracle/testdata/library_array_find_last.a` command | A |
| `internal/oracle/testdata/json_stringify_scalars.a` command | A |
| `internal/oracle/testdata/json_stringify_values.a` command | A |
| `internal/oracle/testdata/json_stringify_options.a` command | A |
| `internal/oracle/testdata/json_stringify_escapes.a` command | A |
| `internal/oracle/testdata/json_stringify_numbers.a` command | A |
| `internal/oracle/testdata/json_stringify_undefined.a` command | A |
| `internal/oracle/testdata/json_stringify_indent.a` command | A |
| `internal/oracle/testdata/json_stringify_keys.a` command | A |
| `internal/oracle/testdata/json_stringify_replacer.a` command | A |
| `internal/oracle/testdata/library_function_expressions.a` command | A |
| `internal/oracle/testdata/library_fnexpr_recurse.a` command | A |
| `internal/oracle/testdata/library_fnexpr_store.a` command | A |
| `internal/oracle/testdata/library_fnexpr_loops.a` command | A |
| `internal/oracle/testdata/library_for_in.a` command | A |
| `internal/oracle/testdata/library_for_in_keys.a` command | A |
| `internal/oracle/testdata/library_for_in_live.a` command | A |
| `internal/oracle/testdata/library_globals.a` command | A |
| `internal/oracle/testdata/library_globals_typeof.a` command | A |
| `internal/load/testdata/0.1/compile/01_hello.ts` command | A |
| `internal/load/testdata/0.1/compile/02_fizzbuzz.ts` command | A |
| `internal/load/testdata/0.1/compile/03_shapes.ts` command | A |
| `internal/load/testdata/0.1/compile/04_closures.ts` command | A |
| `internal/load/testdata/0.1/compile/05_wordcount.ts` command | A |
| `internal/load/testdata/0.1/compile/06_stack.ts` command | A |
| `internal/load/testdata/0.1/compile/07_modules/main.ts` command | A |
| `internal/load/testdata/0.1/compile/08_results.ts` command | A |
| `internal/load/testdata/0.1/compile/09_tree.ts` command | A |
| `internal/load/testdata/0.1/compile/10_unicode.ts` command | A |
| `internal/oracle/testdata/strings.a` command | A |
| `internal/oracle/testdata/numbers.a` command | A |
| `internal/oracle/testdata/bitwise_sweep.a` command | A |
| `internal/oracle/testdata/loops.a` command | A |
| `internal/oracle/testdata/booleans.a` command | A |
| `internal/oracle/testdata/shadowing.a` command | A |
| `internal/oracle/testdata/functions.a` command | A |
| `internal/oracle/testdata/effects.a` command | A |
| `internal/oracle/testdata/dead_zone.a` command | A |
| `internal/oracle/testdata/objects.a` command | A |
| `internal/oracle/testdata/modules/main.a` command | A |
| `internal/oracle/testdata/panic.a` command | A |
| `internal/oracle/testdata/maps_and_text.a` command | A |
| `internal/oracle/testdata/lone_surrogates.a` command | A |
| `internal/oracle/testdata/sorting.a` command | A |
| `internal/oracle/testdata/classes.a` command | A |
| `internal/oracle/testdata/closures.a` command | A |
| `internal/oracle/testdata/local_console.a` command | B |
| `internal/oracle/testdata/indexing.a` command | A |
| `internal/oracle/testdata/writes.a` command | A |
| `internal/oracle/testdata/writes_past_end.a` command | A |
| `internal/oracle/testdata/casts.a` command | A |
| `internal/oracle/testdata/cast_fails.a` command | A |
| `internal/oracle/testdata/updates.a` command | A |
| `internal/oracle/testdata/read_order.a` command | A |
| `internal/oracle/testdata/string_index.a` command | A |
| `internal/oracle/testdata/visits.a` command | A |
| `internal/oracle/testdata/searches.a` command | A |
| `internal/oracle/testdata/spreads.a` command | A |
| `internal/oracle/testdata/maybe_numbers.a` command | A |
| `internal/oracle/testdata/defaults.a` command | A |
| `internal/oracle/testdata/search_halves.a` command | A |
| `internal/oracle/testdata/strings_more.a` command | A |
| `internal/oracle/testdata/number_parsing.a` command | A |
| `internal/oracle/testdata/library_math_number_math.a` command | A |
| `internal/oracle/testdata/library_math_number_convert.a` command | A |
| `internal/oracle/testdata/library_math_number_prototype.a` command | A |
| `internal/oracle/testdata/navigation.a` command | A |
| `internal/oracle/testdata/number_formats.a` command | A |
| `internal/oracle/testdata/precision_range.a` command | A |
| `internal/oracle/testdata/radixes.a` command | A |
| `internal/oracle/testdata/radix_range.a` command | A |
| `internal/oracle/testdata/optional_numbers.a` command | A |
| `internal/oracle/testdata/map_iteration.a` command | A |
| `internal/oracle/testdata/sorts.a` command | A |
| `internal/oracle/testdata/sort_releases.a` command | A |
| `internal/oracle/testdata/timsort.a` command | A |
| `internal/oracle/testdata/unused_parameters.a` command | A |
| `internal/oracle/testdata/map_shrinks.a` command | A |
| `internal/oracle/testdata/find_shrinks.a` command | A |
| `internal/oracle/testdata/find_index_shrinks.a` command | A |
| `internal/oracle/testdata/self_assignments.a` command | A |
| `internal/oracle/testdata/method_closures.a` command | A |
| `internal/oracle/testdata/splice_empty.a` command | A |
| `internal/oracle/testdata/narrowed_reads.a` command | A |
| `internal/oracle/testdata/narrowed_writes.a` command | A |
| `internal/oracle/testdata/narrowed_methods.a` command | A |
| `internal/oracle/testdata/narrowed_fields.a` command | A |
| `internal/oracle/testdata/narrowed_numbers.a` command | A |
| `internal/oracle/testdata/narrowed_compared.a` command | A |
| `internal/oracle/testdata/sort_top_level.a` command | A |
| `internal/oracle/testdata/splices.a` command | A |
| `internal/oracle/testdata/fills.a` command | A |
| `internal/oracle/testdata/fill_length.a` command | A |
| `internal/oracle/testdata/array_from.a` command | A |
| `internal/oracle/testdata/array_from_length.a` command | A |
| `internal/oracle/testdata/array_from_undefined.a` command | A |
| `internal/oracle/testdata/weak_parent.a` command | A |
| `internal/oracle/testdata/doubly_linked.a` command | A |
| `internal/oracle/testdata/fresh_parser.a` command | A |
| `internal/oracle/testdata/fresh_writes.a` command | A |
| `internal/oracle/testdata/fresh_calls.a` command | A |
| `internal/oracle/testdata/weak_narrowed.a` command | A |
| `internal/oracle/testdata/exceptions.a` command | A |
| `internal/oracle/testdata/exceptions_uncaught.a` command | A |
| `internal/oracle/testdata/exceptions_empty.a` command | A |
| `internal/oracle/testdata/closures_throw.a` command | A |
| `internal/oracle/testdata/closures_throw_uncaught.a` command | A |
| `internal/oracle/testdata/finally_leaves.a` command | A |
| `internal/oracle/testdata/reuse_foreach_global.a` command | A |
| `internal/oracle/testdata/param_assigned_in_try.a` command | A |
| `internal/oracle/testdata/named_function_values.a` command | A |
| `internal/oracle/testdata/panic_in_try.a` command | A |
| `internal/oracle/testdata/invariance_readonly.a` command | A |
| `internal/oracle/testdata/tuples_kept.a` command | A |
| `internal/oracle/testdata/undefined_keys.a` command | A |
| `internal/oracle/testdata/undefined_strings.a` command | A |
| `internal/oracle/testdata/maybe_booleans.a` command | A |
| `internal/oracle/testdata/maybe_boolean_panic.a` command | A |
| `internal/oracle/testdata/unions.a` command | A |
| `internal/oracle/testdata/maybe_number_slots.a` command | A |
| `internal/oracle/testdata/case_mapping.a` command | A |
| `internal/oracle/testdata/undefined_elements.a` command | A |
| `internal/oracle/testdata/map_zero_keys.a` command | A |
| `internal/oracle/testdata/string_limits.a` command | A |
| `internal/oracle/testdata/string_too_long.a` command | A |
| `internal/oracle/testdata/pad_too_long.a` command | A |
| `internal/oracle/testdata/stack_overflow.a` command | A |
| `internal/oracle/testdata/adversarial_order.a` command | A |
| `internal/oracle/testdata/adversarial_exits.a` command | A |
| `internal/oracle/testdata/adversarial_iteration.a` command | A |
| `internal/oracle/testdata/number_edges.a` command | A |
| `internal/oracle/testdata/long_chain.a` command | A |
| `internal/oracle/testdata/write_after_shrink.a` command | A |
| `internal/oracle/testdata/normalize.a` command | A |
| `internal/oracle/testdata/normalize_form.a` command | A |
| `internal/oracle/testdata/string_positions.a` command | A |
| `internal/oracle/testdata/long_literals.a` command | A |
| `internal/oracle/testdata/class_layouts.a` command | A |
| `internal/oracle/testdata/ascii_scan.a` command | A |
| `internal/oracle/testdata/size_class_churn.a` command | A |
| `internal/oracle/testdata/borrow_reassigned.a` command | A |
| `internal/oracle/testdata/borrow_defined_lent.a` command | A |
| `internal/oracle/testdata/borrow_defined_lent_field.a` command | A |
| `internal/oracle/testdata/writes_in_try.a` command | A |
| `internal/oracle/testdata/class_as_interface.a` command | A |
| `internal/oracle/testdata/optional_class_method.a` command | A |
| `internal/oracle/testdata/set_undefined.a` command | A |
| `internal/oracle/testdata/borrow_map_overwrite.a` command | A |
| `internal/oracle/testdata/spread_snapshot.a` command | A |
| `internal/oracle/testdata/reuse.a` command | A |
| `internal/oracle/testdata/reuse_arrays.a` command | A |
| `internal/oracle/testdata/reuse_forward.a` command | A |
| `internal/oracle/testdata/reuse_global_sibling.a` command | A |
| `internal/oracle/testdata/reuse_weak_during_spread.a` command | A |
| `internal/oracle/testdata/reuse_weak_after_reuse.a` command | A |
| `internal/oracle/testdata/regions.a` command | A |
| `internal/oracle/testdata/regions_throw.a` command | A |
| `internal/oracle/testdata/regions_constructor_capture.a` command | A |
| `internal/oracle/testdata/borrow_element.a` command | A |
| `internal/oracle/testdata/borrow_element_throw.a` command | A |
| `internal/oracle/testdata/borrow_element_virtual_store.a` command | A |
| `internal/oracle/testdata/borrow_element_virtual_move.a` command | A |
| `internal/oracle/testdata/borrow_element_super_move.a` command | A |
| `internal/oracle/testdata/move_throw.a` command | A |
| `internal/oracle/testdata/throw_keeps_old_value.a` command | A |
| `internal/oracle/testdata/throw_keeps_old_value_variants.a` command | A |
| `internal/oracle/testdata/throw_in_writes.a` command | A |
| `internal/oracle/testdata/throw_global_move.a` command | A |
| `internal/oracle/testdata/spread_undefined.a` command | A |
| `internal/oracle/testdata/lent_reads.a` command | A |
| `internal/oracle/testdata/large_output.a` command | A |
| `internal/oracle/testdata/output_then_panic.a` command | A |
| `internal/oracle/testdata/interleaved.a` command | A |
| `internal/oracle/testdata/trig_reduction.a` command | A |
| `internal/oracle/testdata/sets.a` command | A |
| `internal/oracle/testdata/set_maybe_numbers.a` command | A |
| `internal/oracle/testdata/library_map_set.a` command | A |
| `internal/oracle/testdata/library_map_set_keys.a` command | A |
| `internal/oracle/testdata/library_map_set_iterators.a` command | A |
| `internal/oracle/testdata/library_map_set_construct.a` command | A |
| `internal/oracle/testdata/library_map_set_group_by.a` command | A |
| `internal/oracle/testdata/maybe_collections.a` command | A |
| `internal/oracle/testdata/stack_tail_call.a` command | A |
| `internal/oracle/testdata/stack_forever.a` command | A |
| `internal/oracle/testdata/optional_strings.a` command | A |
| `internal/oracle/testdata/concat_too_long.a` command | A |
| `internal/oracle/testdata/replace_all_large.a` command | A |
| `internal/oracle/testdata/collections.a` command | A |
| `internal/oracle/testdata/gaps.a` command | A |
| `internal/oracle/testdata/normalize_long_marks.a` command | A |
| `internal/oracle/testdata/power_of_two_string.a` command | A |
| `internal/oracle/testdata/declared_later.a` command | A |
| `internal/oracle/testdata/return_panic.a` command | A |
| `internal/oracle/testdata/return_panic_fires.a` command | A |
| `internal/oracle/testdata/from_codes.a` command | A |
| `internal/oracle/testdata/from_code_point_fails.a` command | A |
| `internal/oracle/testdata/bitwise.a` command | A |
| `internal/oracle/testdata/tuple_values.a` command | A |
| `internal/oracle/testdata/generic_functions.a` command | A |
| `internal/oracle/testdata/generic_method_return.a` command | A |
| `internal/oracle/testdata/generic_values.a` command | A |
| `internal/oracle/testdata/undefined_references.a` command | A |
| `internal/oracle/testdata/spread_calls.a` command | A |
| `internal/oracle/testdata/utf8_view.a` command | A |
| `internal/oracle/testdata/utf8_view_fails.a` command | A |
| `internal/oracle/testdata/search_from.a` command | A |
| `internal/oracle/testdata/shared_slices.a` command | A |
| `internal/oracle/testdata/string_append.a` command | A |
| `internal/oracle/testdata/shared_slice_append.a` command | A |
| `internal/oracle/testdata/search_from_sweep.a` command | A |
| `internal/oracle/testdata/integer_format.a` command | A |
| `internal/oracle/testdata/reuse_throw.a` command | A |
| `internal/oracle/testdata/reuse_narrowed.a` command | A |
| `internal/oracle/testdata/reuse_lent_global.a` command | A |
| `internal/oracle/testdata/reuse_spread_method.a` command | A |
| `internal/oracle/testdata/reuse_spread_method_alias.a` command | A |
| `internal/oracle/testdata/try_assignments.a` command | A |
| `internal/oracle/testdata/library_string_conversion.a` command | A |
| `internal/oracle/testdata/library_string_prototype.a` command | A |
| `internal/oracle/testdata/library_string_indices.a` command | A |
| `internal/oracle/testdata/library_string_raw.a` command | A |
| `internal/oracle/testdata/library_string_existing.a` command | A |
| `internal/oracle/testdata/library_map_set_visit.a` command | A |
| `internal/oracle/testdata/library_map_set_setops.a` command | A |
| `internal/oracle/testdata/library_map_set_zeros.a` command | A |
| `internal/oracle/testdata/library_map_set_groupby_keys.a` command | A |
| `internal/oracle/testdata/library_map_set_next.a` command | A |
| `internal/oracle/testdata/has_own.a` command | A |
| `internal/oracle/testdata/regexp.a` command | A |
| `internal/oracle/testdata/sweeps/regexp_methods.a` command | A |
| `internal/oracle/testdata/regexp_matchall_nonglobal.a` command | A |
| `internal/oracle/testdata/regexp_replaceall_nonglobal.a` command | A |
| `internal/oracle/testdata/regexp_null_narrowed.a` command | A |
| `internal/oracle/testdata/regexp_replace.a` command | A |
| `internal/oracle/testdata/regexp_split.a` command | A |
| `internal/oracle/testdata/regexp_exec.a` command | A |
| `internal/oracle/testdata/regexp_match.a` command | A |
| `internal/oracle/testdata/regexp_search.a` command | A |
| `internal/oracle/testdata/regexp_unicode.a` command | A |
| `internal/oracle/testdata/regexp_split_pair_pattern.a` command | A |
| `internal/oracle/testdata/class_features_static.a` command | A |
| `internal/oracle/testdata/class_features_static_private.a` command | A |
| `internal/oracle/testdata/class_features_private.a` command | A |
| `internal/oracle/testdata/class_features_accessors.a` command | A |
| `internal/oracle/testdata/class_features_twice.a` command | A |
| `internal/oracle/testdata/class_features_retained.a` command | A |
| `internal/oracle/testdata/class_features_distinct.a` command | A |
| `internal/oracle/testdata/class_inheritance.a` command | A |
| `internal/oracle/testdata/class_inheritance_exceptions.a` command | A |
| `internal/oracle/testdata/class_inheritance_order.a` command | A |
| `internal/oracle/testdata/class_identity.a` command | A |
| `internal/oracle/testdata/class_inheritance_memory.a` command | A |
| `internal/oracle/testdata/class_inheritance_generic.a` command | A |
| `internal/oracle/testdata/class_inheritance_interface.a` command | A |
| `internal/oracle/testdata/class_inheritance_conditional.a` command | A |
| `internal/oracle/testdata/devirtualize.a` command | A |
| `internal/oracle/testdata/user_iterators.a` command | A |
| `internal/oracle/testdata/user_iterators_rest_tdz.a` command | A |
| `internal/oracle/testdata/e4eec87_u02_optional_absent.a` command | A |
| `internal/oracle/testdata/literal_optional_shapes.a` command | A |
| `internal/oracle/testdata/e4eec87_u03_discriminated_undefined.a` command | A |
| `internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_field_narrowed.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_class_narrowed.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_alias_narrowed.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_field_present.a` command | A |
| `internal/oracle/testdata/object_prototype.a` command | A |
| `internal/fresh/testdata/regexp_tree.ts` command | A |
| `internal/oracle/testdata/regexp_cycle_fields.a` command | A |
| `internal/oracle/testdata/regexp_cycle_collections.a` command | A |
| `internal/oracle/testdata/regexp_cycle_closures.a` command | A |
| `internal/oracle/testdata/regexp_cycle_weak.a` command | A |

## Threads

Plain `wasm32-wasi` uses unshared memory and `-mno-atomics`. W7's guarded pool
startup selects one executor regardless of `ADAMIC_THREADS`; no worker is started.
`parallelMap` uses the existing single-thread path, visiting ascending indices,
preserving result positions, and stopping at the first thrown exception. Input,
closure and result graph publication still follows the native ownership ABI.
The request handler remains synchronous within one module instance.

WASI SDK 27 clang lowers the runtime's `_Thread_local` storage and C atomics to
ordinary memory accesses without shared memory. A disassembled probe shows
`i32.load`, `i32.add` and `i32.store` for TLS increment and atomic fetch-add,
with no wasm atomic instruction. No replacement counter implementation is added.
WASI libc supplies single-thread mutex/once operations. Its `pthread_create`
stub returned 6 (ENXIO) and never called the worker in the executed probe; that
stub is a failure path, not a pool implementation.

The WASI panic path skips the native wait for a competing panic thread. Worker
stack initialization leaves the constructor's linear-stack limit intact.
Native preprocessing retains the existing pool, TLS and atomic implementation;
all 50 release runtime objects were compared byte for byte at `-O2`.

The first merged-tree checks exposed two additional integration blockers:
`-pthread` conflicts with `-mno-atomics` in the driver, and `share.c` hashes a
32-bit pointer with a shift by 33. The W7 evidence distinguishes these compile
failures from runtime agreement; a green WASI oracle is required before landing.
See `cloud/reports/wasm-threads/` for the hooks, commands, counts and limitations.

Actual wasm threads would require a wasi-threads-capable host, shared linear
memory, atomics-enabled compilation and linking, worker instantiation and TLS
initialization, per-thread stacks, and tested scheduling/ownership and shutdown.
The native pool cannot acquire those facilities from ordinary WASI Preview 1.
That future target needs its own Node comparisons and race/lifetime proofs.
W7 makes no speed measurement.
